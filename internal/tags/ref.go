package tags

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/arloliu/fuda/internal/types"
)

// Resolver interface for resolving external references.
type Resolver interface {
	// Resolve returns the content referenced by the uri.
	Resolve(ctx context.Context, uri string) ([]byte, error)
}

// ProcessRef processes 'ref' and 'refFrom' tags.
//
// The ref tag can contain template expressions using ${...} syntax:
//   - ${.FieldName} - references the value of a field in the same struct
//   - ${.Nested.Field} - references nested struct fields
//   - ${env:KEY} - reads an environment variable
//
// Note: Fields referenced in templates must appear earlier in the struct.
//
// The templateData parameter is pre-computed struct data for template execution.
// Pass nil to have it computed on-demand (for backward compatibility).
//
// Example:
//
//	type Config struct {
//	    SecretDir string `default:"/etc/secrets"`
//	    Account   string `yaml:"account"`
//	    Password  string `ref:"file://${.SecretDir}/${.Account}-password"`
//	}
func ProcessRef(
	ctx context.Context,
	field reflect.StructField,
	value reflect.Value,
	parentVal reflect.Value,
	resolver Resolver,
	envPrefix string,
	templateData any,
) error {
	if resolver == nil {
		return nil
	}

	// Only resolve if value is zero
	if !value.IsZero() {
		return nil
	}

	var uri string

	// Check refFrom first
	if refFrom := field.Tag.Get("refFrom"); refFrom != "" {
		// Find the referenced field in parent
		refField := parentVal.FieldByName(refFrom)
		if !refField.IsValid() {
			return fmt.Errorf("refFrom field '%s' not found", refFrom)
		}

		// refFrom only supports string fields
		if refField.Kind() != reflect.String {
			return fmt.Errorf("refFrom field '%s' must be a string, got %s", refFrom, refField.Kind())
		}

		// Get the string value (safe for both exported and unexported fields)
		uriVal := refField.String()

		// "Peek" logic: if refField is zero, check its default tag
		if uriVal == "" {
			// Find the struct field for refField to get tag
			parentType := parentVal.Type()
			if f, ok := parentType.FieldByName(refFrom); ok {
				defaultTag := f.Tag.Get("default")
				if defaultTag != "" && defaultTag != "-" {
					uriVal = defaultTag
				}
			}
		}

		if uriVal != "" {
			uri = uriVal
		}
	}

	// Fallback to ref tag if uri is still empty
	if uri == "" {
		if refTag := field.Tag.Get("ref"); refTag != "" {
			uri = refTag
		}
	}

	if uri == "" {
		return nil
	}

	// Process template expressions in URI if present
	if strings.Contains(uri, "${") {
		config := TemplateConfig{
			Strict:    false, // ref uses permissive mode by default
			Resolver:  resolver,
			EnvPrefix: envPrefix,
		}

		// Use pre-computed data if available, otherwise compute on-demand
		data := templateData
		if data == nil {
			data = StructToData(parentVal)
		}

		expanded, err := ProcessTemplate(ctx, uri, data, config)
		if err != nil {
			return fmt.Errorf("failed to expand ref template: %w", err)
		}

		uri = expanded
	}

	// Normalize URI (add file:// prefix if needed)
	uri = normalizeURI(uri)

	content, err := resolver.Resolve(ctx, uri)
	if err != nil {
		return fmt.Errorf("failed to resolve ref '%s': %w", uri, err)
	}

	return types.Convert(string(content), value)
}

func normalizeURI(uri string) string {
	if strings.Contains(uri, "://") {
		return uri
	}

	return "file://" + uri
}
