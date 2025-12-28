package tags

import (
	"context"
	"fmt"
	"reflect"

	"github.com/arloliu/fuda/internal/types"
)

// DSNConfig holds configuration for DSN template processing.
type DSNConfig struct {
	// Strict controls error behavior for empty/undefined values.
	// If true, returns an error when a referenced field is empty or undefined.
	// If false (default), outputs empty string for missing values.
	Strict bool
}

// ProcessDSN processes the 'dsn' tag for a field.
// It uses Go template syntax with ${...} delimiters to compose connection strings
// from other fields and resolved values.
//
// The dsn tag is processed AFTER all other tags (env, ref, default) so that
// referenced fields have their final values.
//
// Template syntax (uses ${...} delimiters):
//   - ${.FieldName} - references the value of a field in the same struct
//   - ${.Nested.Field} - references nested struct fields
//   - ${ref "uri"} - resolves a URI inline using the resolver
//   - ${env "KEY"} - reads an environment variable
//
// The templateData parameter is pre-computed struct data for template execution.
// Pass nil to have it computed on-demand (for backward compatibility).
//
// Tag options:
//   - dsn:"template" - the template string
//   - dsnStrict:"true" - enable strict mode (error on empty values)
//
// Example:
//
//	type Config struct {
//	    DBHost     string `yaml:"db_host" default:"localhost"`
//	    DBUser     string `ref:"vault:///secret/data/db#username"`
//	    DBPassword string `ref:"vault:///secret/data/db#password"`
//	    DatabaseDSN string `dsn:"postgres://${.DBUser}:${.DBPassword}@${.DBHost}:5432/mydb"`
//	}
func ProcessDSN(
	ctx context.Context,
	field reflect.StructField,
	value reflect.Value,
	parentVal reflect.Value,
	resolver Resolver,
	envPrefix string,
	templateData any,
) error {
	tag := field.Tag.Get("dsn")
	if tag == "" {
		return nil
	}

	// Only process if field is zero (don't overwrite existing values)
	if !value.IsZero() {
		return nil
	}

	// Only string fields can have dsn tag
	if value.Kind() != reflect.String {
		return fmt.Errorf("dsn tag can only be used on string fields, got %s", value.Kind())
	}

	// Build template config from DSN options
	config := TemplateConfig{
		Strict:    field.Tag.Get("dsnStrict") == "true",
		Resolver:  resolver,
		EnvPrefix: envPrefix,
	}

	// Use pre-computed data if available, otherwise compute on-demand
	data := templateData
	if data == nil {
		data = StructToData(parentVal)
	}

	// Process template using shared template processor
	result, err := ProcessTemplate(ctx, tag, data, config)
	if err != nil {
		return fmt.Errorf("dsn: %w", err)
	}

	return types.Convert(result, value)
}
