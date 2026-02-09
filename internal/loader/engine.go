package loader

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/arloliu/fuda/internal/tags"
	"github.com/arloliu/fuda/internal/types"
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type RefResolver interface {
	// Resolve returns the content referenced by the uri.
	Resolve(ctx context.Context, uri string) ([]byte, error)
}

// Engine is the internal configuration processing engine.
// It handles YAML unmarshaling, tag processing (env, ref, default), and validation.
type Engine struct {
	Validator      *validator.Validate
	RefResolver    RefResolver
	EnvPrefix      string
	Source         []byte
	SourceName     string // Name of the source (e.g., "config.yaml", "reader", "bytes")
	Timeout        time.Duration
	TemplateConfig *TemplateConfig
	TemplateData   any
	DotenvConfig   *DotenvConfig
	Overrides      map[string]any // Programmatic value overrides (dot-notation supported)
	// EnableSizePreprocess controls size-string preprocessing (default: true).
	EnableSizePreprocess *bool
	// EnableDurationPreprocess controls duration-string preprocessing (default: true).
	EnableDurationPreprocess *bool
}

func (e *Engine) Load(target any) error {
	// Load dotenv files first, before any env tag processing
	if err := e.loadDotenvFiles(); err != nil {
		return fmt.Errorf("failed to load dotenv files: %w", err)
	}

	ctx := context.Background()
	if e.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, e.Timeout)
		defer cancel()
	}

	// Process template if configured
	source := e.Source
	if e.TemplateData != nil && len(source) > 0 {
		processed, err := ProcessTemplate(source, e.TemplateData, e.TemplateConfig)
		if err != nil {
			if e.SourceName != "" {
				return fmt.Errorf("failed to process template in %s: %w", e.SourceName, err)
			}

			return fmt.Errorf("failed to process template: %w", err)
		}

		source = processed
	}

	// 1. Apply overrides and unmarshal Source
	// Handle overrides even if source is empty (allows creating config purely from overrides)
	if len(e.Overrides) > 0 {
		var err error
		source, err = e.applyOverrides(source)
		if err != nil {
			return fmt.Errorf("failed to apply overrides: %w", err)
		}
	}

	if len(source) > 0 {
		// Unmarshal to node tree for duration preprocessing
		var node yaml.Node
		if err := yaml.Unmarshal(source, &node); err != nil {
			if e.SourceName != "" {
				return fmt.Errorf("failed to unmarshal %s: %w", e.SourceName, err)
			}

			return fmt.Errorf("failed to unmarshal source: %w", err)
		}

		// Preprocess nodes
		if resolvePreprocessFlag(e.EnableSizePreprocess) {
			preprocessSizeNodesForType(&node, reflect.TypeOf(target))
		}
		if resolvePreprocessFlag(e.EnableDurationPreprocess) {
			preprocessDurationNodesForType(&node, reflect.TypeOf(target))
		}

		// Decode to target struct
		if err := node.Decode(target); err != nil {
			if e.SourceName != "" {
				return fmt.Errorf("failed to decode %s: %w", e.SourceName, err)
			}

			return fmt.Errorf("failed to decode source: %w", err)
		}
	}

	targetVal := reflect.ValueOf(target)

	// Process recursive tags with cycle detection
	// Pass the original pointer so cycle detection can track it
	visited := make(map[uintptr]bool)
	if err := e.processStructWithVisited(ctx, targetVal, visited); err != nil {
		return err
	}

	// 5. Validate
	if e.Validator != nil {
		if err := e.Validator.Struct(target); err != nil {
			return &types.ValidationError{Errors: []error{err}}
		}
	}

	return nil
}

func (e *Engine) processStructWithVisited(ctx context.Context, v reflect.Value, visited map[uintptr]bool) error {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		// Cycle detection: check if we've already visited this pointer
		ptr := v.Pointer()
		if visited[ptr] {
			return fmt.Errorf("cycle detected: pointer %v already visited", v.Type())
		}
		visited[ptr] = true
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()
	for i := range v.NumField() {
		field := t.Field(i)
		fieldVal := v.Field(i)

		// Skip unexported fields
		if !fieldVal.CanSet() {
			continue
		}

		// Process nested elements
		if err := e.processNestedElementsWithVisited(ctx, fieldVal, visited); err != nil {
			return err
		}

		// Apply tags
		if err := e.applyTags(ctx, field, fieldVal, v); err != nil {
			return err
		}
	}

	// Handle Setter interface (Dynamic Defaults)
	// Call SetDefaults after all fields are processed (Post-Order)
	if v.CanAddr() {
		if setter, ok := v.Addr().Interface().(types.Setter); ok {
			setter.SetDefaults()
		}
	}

	return nil
}

func resolvePreprocessFlag(flag *bool) bool {
	if flag == nil {
		return true
	}

	return *flag
}

// processNestedElementsWithVisited recursively processes nested structs, slices, and maps with cycle detection.
func (e *Engine) processNestedElementsWithVisited(ctx context.Context, fieldVal reflect.Value, visited map[uintptr]bool) error {
	//nolint:exhaustive // Only struct-like types need processing
	switch fieldVal.Kind() {
	case reflect.Struct:
		return e.processStructWithVisited(ctx, fieldVal, visited)
	case reflect.Pointer:
		if fieldVal.Type().Elem().Kind() == reflect.Struct {
			return e.processStructWithVisited(ctx, fieldVal, visited)
		}
	case reflect.Slice:
		return e.processSliceElementsWithVisited(ctx, fieldVal, visited)
	case reflect.Map:
		return e.processMapValuesWithVisited(ctx, fieldVal, visited)
	}

	return nil
}

// processSliceElementsWithVisited recursively processes struct elements in a slice with cycle detection.
func (e *Engine) processSliceElementsWithVisited(ctx context.Context, sliceVal reflect.Value, visited map[uintptr]bool) error {
	for j := range sliceVal.Len() {
		elem := sliceVal.Index(j)
		// Check if element is a struct or pointer to struct
		isStruct := elem.Kind() == reflect.Struct
		isPtrToStruct := elem.Kind() == reflect.Pointer && !elem.IsNil() && elem.Elem().Kind() == reflect.Struct
		if isStruct || isPtrToStruct {
			if err := e.processStructWithVisited(ctx, elem, visited); err != nil {
				return err
			}
		}
	}

	return nil
}

// processMapValuesWithVisited recursively processes struct values in a map with cycle detection.
func (e *Engine) processMapValuesWithVisited(ctx context.Context, mapVal reflect.Value, visited map[uintptr]bool) error {
	iter := mapVal.MapRange()
	for iter.Next() {
		val := iter.Value()
		if val.Kind() == reflect.Struct {
			// Map values are not addressable, so we need to copy, process, and set back
			valCopy := reflect.New(val.Type()).Elem()
			valCopy.Set(val)
			if err := e.processStructWithVisited(ctx, valCopy, visited); err != nil {
				return err
			}
			mapVal.SetMapIndex(iter.Key(), valCopy)
		}
	}

	return nil
}

// applyTags applies env, ref, and default tags to a field.
func (e *Engine) applyTags(ctx context.Context, field reflect.StructField, fieldVal, parentVal reflect.Value) error {
	// Apply Env Overrides
	envApplied, err := tags.ProcessEnv(field, fieldVal, e.EnvPrefix)
	if err != nil {
		return &types.FieldError{Path: field.Name, Tag: "env", Err: err}
	}

	// Lazy template data computation - only computed once if either ref or dsn needs it
	var templateData any
	getTemplateData := func() any {
		if templateData == nil {
			templateData = tags.StructToData(parentVal)
		}
		return templateData
	}

	// Resolve Refs
	refResolved, err := tags.ProcessRef(ctx, field, fieldVal, parentVal, e.RefResolver, e.EnvPrefix, getTemplateData())
	if err != nil {
		return &types.FieldError{Path: field.Name, Tag: "ref", Err: err}
	}

	// Apply Defaults (skip if env was applied or ref resolved a value)
	// This ensures env-set zero values (like "false") aren't overwritten by defaults
	if !envApplied && !refResolved {
		if err := tags.ProcessDefault(field, fieldVal); err != nil {
			return &types.FieldError{Path: field.Name, Tag: "default", Err: err}
		}
	}

	// Process DSN templates (after all other tags, so referenced fields have their values)
	if err := tags.ProcessDSN(ctx, field, fieldVal, parentVal, e.RefResolver, e.EnvPrefix, getTemplateData()); err != nil {
		return &types.FieldError{Path: field.Name, Tag: "dsn", Err: err}
	}

	return nil
}

// applyOverrides applies programmatic overrides to the source YAML.
// Returns the modified source as YAML bytes.
func (e *Engine) applyOverrides(source []byte) ([]byte, error) {
	// Parse source into a map
	var data map[string]any
	if err := yaml.Unmarshal(source, &data); err != nil {
		return nil, fmt.Errorf("failed to parse source as map: %w", err)
	}

	// Initialize data if empty
	if data == nil {
		data = make(map[string]any)
	}

	// Apply each override
	for key, value := range e.Overrides {
		setNestedValue(data, key, value)
	}

	// Re-marshal to YAML
	modified, err := yaml.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal modified config: %w", err)
	}

	return modified, nil
}

// setNestedValue sets a value in a nested map using dot notation.
// For example, "database.host" sets data["database"]["host"].
func setNestedValue(data map[string]any, key string, value any) {
	parts := strings.Split(key, ".")

	// Navigate to the parent map, creating intermediate maps as needed
	current := data
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]

		next, exists := current[part]
		if !exists {
			// Create intermediate map
			nextMap := make(map[string]any)
			current[part] = nextMap
			current = nextMap

			continue
		}

		nextMap, ok := next.(map[string]any)
		if !ok {
			// Existing value is not a map, replace it with a map
			nextMap = make(map[string]any)
			current[part] = nextMap
		}

		current = nextMap
	}

	// Set the final key
	current[parts[len(parts)-1]] = value
}
