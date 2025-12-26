package loader

import (
	"context"
	"fmt"
	"reflect"
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
}

func (e *Engine) Load(target any) error {
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

	// 1. Unmarshal Source
	if len(source) > 0 {
		if err := yaml.Unmarshal(source, target); err != nil {
			if e.SourceName != "" {
				return fmt.Errorf("failed to unmarshal %s: %w", e.SourceName, err)
			}

			return fmt.Errorf("failed to unmarshal source: %w", err)
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
	if err := tags.ProcessEnv(field, fieldVal, e.EnvPrefix); err != nil {
		return &types.FieldError{Path: field.Name, Tag: "env", Err: err}
	}

	// Resolve Refs
	if err := tags.ProcessRef(ctx, field, fieldVal, parentVal, e.RefResolver); err != nil {
		return &types.FieldError{Path: field.Name, Tag: "ref", Err: err}
	}

	// Apply Defaults
	if err := tags.ProcessDefault(field, fieldVal); err != nil {
		return &types.FieldError{Path: field.Name, Tag: "default", Err: err}
	}

	return nil
}
