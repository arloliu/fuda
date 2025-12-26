package tags

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/arloliu/fuda/internal/types"
)

type Resolver interface {
	// Resolve returns the content referenced by the uri.
	Resolve(ctx context.Context, uri string) ([]byte, error)
}

// ProcessRef processes 'ref' and 'refFrom' tags.
func ProcessRef(ctx context.Context, field reflect.StructField, value reflect.Value, parentVal reflect.Value, resolver Resolver) error {
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

		// "Peek" logic: if refField is zero, check its default tag
		uriVal := fmt.Sprint(refField.Interface())
		if refField.Kind() == reflect.String {
			uriVal = refField.String()
		}

		if refField.IsZero() {
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
			uri = normalizeURI(uriVal)
		}
	}

	// Fallback to ref tag if uri is still empty
	if uri == "" {
		if refTag := field.Tag.Get("ref"); refTag != "" {
			uri = refTag // ref tag is always a fixed URI
		}
	}

	if uri == "" {
		return nil
	}

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
