package tags

import (
	"context"
	"errors"
	"fmt"
	"os"
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
// Returns (resolved, error) where resolved is true if a value was set (even if empty).
//
// The ref tag can contain template expressions using ${...} syntax:
//
//   - ${.FieldName} - references the value of a field in the same struct
//
//   - ${.Nested.Field} - references nested struct fields
//
//   - ${env:KEY} - reads an environment variable
//
//   - ${env:KEY} - reads an environment variable
//
// Priority:
//  1. refFrom: If present and referenced field is non-empty, its value is used as URI.
//  2. ref: Used as fallback if refFrom is absent or referenced field is empty.
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
) (bool, error) {
	if resolver == nil {
		return false, nil
	}

	// Only resolve if value is zero
	if !value.IsZero() {
		return false, nil
	}

	// Create resolver helper
	resolveURI := newURIResolver(ctx, resolver, envPrefix, templateData, parentVal)

	// Try refFrom first
	if refFrom := field.Tag.Get("refFrom"); refFrom != "" {
		resolved, found, err := processRefFrom(refFrom, parentVal, value, resolveURI)
		if err != nil {
			return false, err
		}
		if found {
			return resolved, nil
		}
		// Not found or empty source - fall through to ref tag
	}

	// Try ref tag as fallback
	if refTag := field.Tag.Get("ref"); refTag != "" {
		content, found, err := resolveURI(refTag)
		if err != nil {
			return false, err
		}
		if found {
			err := types.Convert(string(content), value)

			return err == nil, err
		}
		// Not found - return false to allow default tag to apply
	}

	return false, nil
}

// uriResolverFunc is a function type for resolving URIs.
type uriResolverFunc func(uri string) (content []byte, found bool, err error)

// newURIResolver creates a URI resolver function with template support.
func newURIResolver(
	ctx context.Context,
	resolver Resolver,
	envPrefix string,
	templateData any,
	parentVal reflect.Value,
) uriResolverFunc {
	return func(uri string) (content []byte, found bool, err error) {
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
				return nil, false, fmt.Errorf("failed to expand ref template: %w", err)
			}

			uri = expanded
		}

		// Normalize URI (add file:// prefix if needed)
		uri = normalizeURI(uri)

		content, err = resolver.Resolve(ctx, uri)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil, false, nil // Not found, allow fallback
			}

			return nil, false, fmt.Errorf("failed to resolve ref '%s': %w", uri, err)
		}

		return content, true, nil
	}
}

// processRefFrom handles the refFrom tag logic.
// Returns (resolved, found, error) where:
// - resolved: true if a value was set (even if empty)
// - found: true if refFrom should stop the fallback chain (value was resolved or explicitly empty)
func processRefFrom(
	refFrom string,
	parentVal reflect.Value,
	value reflect.Value,
	resolveURI uriResolverFunc,
) (resolved, found bool, err error) {
	// Find the referenced field in parent
	refField := parentVal.FieldByName(refFrom)
	if !refField.IsValid() {
		return false, false, fmt.Errorf("refFrom field '%s' not found", refFrom)
	}

	// Extract URI value from source field
	uriVal, isExplicitlySet, err := extractRefFromValue(refFrom, refField, parentVal)
	if err != nil {
		return false, false, err
	}

	// Nothing to do if no value and not explicitly set
	if uriVal == "" && !isExplicitlySet {
		return false, false, nil
	}

	// Special case: Explicitly set empty string means "use empty value, stop fallback"
	if uriVal == "" && isExplicitlySet {
		err := types.Convert("", value)

		return err == nil, true, err
	}

	// Resolve the URI
	content, resolvedFromURI, err := resolveURI(uriVal)
	if err != nil {
		return false, false, err
	}
	if resolvedFromURI {
		err := types.Convert(string(content), value)

		return err == nil, true, err
	}

	// URI not found - allow fallback to ref tag
	return false, false, nil
}

// extractRefFromValue extracts the URI value from a refFrom source field.
func extractRefFromValue(
	refFrom string,
	refField reflect.Value,
	parentVal reflect.Value,
) (uriVal string, isExplicitlySet bool, err error) {
	// refFrom supports string or *string fields
	switch {
	case refField.Kind() == reflect.String:
		uriVal = refField.String()
		// Basic strings are not "explicitly set" if empty, maintaining old behavior
	case refField.Kind() == reflect.Pointer && refField.Type().Elem().Kind() == reflect.String:
		if !refField.IsNil() {
			uriVal = refField.Elem().String()
			isExplicitlySet = true
		}
	default:
		return "", false, fmt.Errorf("refFrom field '%s' must be string or *string, got %s", refFrom, refField.Kind())
	}

	// "Peek" logic: if value is missing (empty and not explicit), check its default tag
	if uriVal == "" && !isExplicitlySet {
		parentType := parentVal.Type()
		if f, ok := parentType.FieldByName(refFrom); ok {
			defaultTag := f.Tag.Get("default")
			if defaultTag != "" && defaultTag != "-" {
				uriVal = defaultTag
			}
		}
	}

	return uriVal, isExplicitlySet, nil
}

func normalizeURI(uri string) string {
	if strings.Contains(uri, "://") {
		return uri
	}

	return "file://" + uri
}
