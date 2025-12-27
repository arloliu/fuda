package tags

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"text/template"

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

	// Parse dsnStrict option
	config := DSNConfig{
		Strict: field.Tag.Get("dsnStrict") == "true",
	}

	// Preprocess the tag to convert shorthand syntax to template function calls
	// ${ref:uri} -> ${ref "uri"}
	// ${env:KEY} -> ${env "KEY"}
	processedTag := preprocessDSNTag(tag)

	// Build template with custom functions and ${...} delimiters
	funcMap := template.FuncMap{
		"ref": makeRefFunc(ctx, resolver),
		"env": makeEnvFunc(envPrefix),
	}

	// Configure missing key behavior based on strict mode
	missingKeyOpt := "missingkey=zero" // Default: return zero value
	if config.Strict {
		missingKeyOpt = "missingkey=error" // Strict: return error on missing field
	}

	tmpl, err := template.New("dsn").
		Delims("${", "}").
		Funcs(funcMap).
		Option(missingKeyOpt).
		Parse(processedTag)
	if err != nil {
		return fmt.Errorf("failed to parse dsn template: %w", err)
	}

	// Create data from the parent struct for template execution
	// This allows nested field access like ${.Database.Host}
	data := structToData(parentVal)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		// In strict mode, missing keys return an exec error
		if config.Strict && strings.Contains(err.Error(), "map has no entry for key") {
			return fmt.Errorf("dsn template references undefined field: %w", err)
		}
		// Also handle the case where a struct field is missing (nil pointer field etc)
		if config.Strict && strings.Contains(err.Error(), "nil pointer evaluating") {
			return fmt.Errorf("dsn template references nil pointer field: %w", err)
		}

		return fmt.Errorf("failed to execute dsn template: %w", err)
	}

	result := buf.String()

	return types.Convert(result, value)
}

// preprocessDSNTag converts shorthand syntax to Go template function calls.
//   - ${ref:uri} -> ${ref "uri"}
//   - ${env:KEY} -> ${env "KEY"}
//
// This allows users to write clean, unquoted URIs similar to the ref tag.
func preprocessDSNTag(tag string) string {
	// Match ${ref:...} pattern - everything after "ref:" until the closing }
	tag = preprocessFunc(tag, "ref")
	tag = preprocessFunc(tag, "env")
	return tag
}

// preprocessFunc converts ${func:arg} to ${func "arg"} for a specific function name.
func preprocessFunc(tag, funcName string) string {
	prefix := "${" + funcName + ":"
	var result strings.Builder
	remaining := tag

	for {
		idx := strings.Index(remaining, prefix)
		if idx == -1 {
			result.WriteString(remaining)
			break
		}

		// Write everything before the match
		result.WriteString(remaining[:idx])

		// Find the closing brace
		afterPrefix := remaining[idx+len(prefix):]
		closeIdx := findClosingBrace(afterPrefix)
		if closeIdx == -1 {
			// No closing brace found, write as-is
			result.WriteString(remaining[idx:])
			break
		}

		// Extract the argument and convert to quoted form
		// Escape any double quotes in the argument to prevent breaking the string literal
		arg := afterPrefix[:closeIdx]
		arg = strings.ReplaceAll(arg, "\"", "\\\"")

		result.WriteString("${")
		result.WriteString(funcName)
		result.WriteString(" \"")
		result.WriteString(arg)
		result.WriteString("\"}")

		remaining = afterPrefix[closeIdx+1:]
	}

	return result.String()
}

// findClosingBrace finds the index of the closing } for a DSN template expression.
func findClosingBrace(s string) int {
	depth := 0
	for i, c := range s {
		switch c {
		case '{':
			depth++
		case '}':
			if depth == 0 {
				return i
			}
			depth--
		}
	}

	return -1
}

// makeRefFunc creates a template function that resolves URIs.
// Accepts variadic args to support both quoted and unquoted usage:
//   - ${ref "vault:///secret#pass"} - quoted string
//   - ${ref vault:///secret#pass} - unquoted (parsed as multiple args)
func makeRefFunc(ctx context.Context, resolver Resolver) func(...string) (string, error) {
	return func(parts ...string) (string, error) {
		if resolver == nil {
			return "", errors.New("no resolver configured for ref function in dsn template")
		}

		if len(parts) == 0 {
			return "", errors.New("ref function requires a URI argument")
		}

		// Join parts to reconstruct URI (handles unquoted args split by spaces)
		uri := strings.Join(parts, " ")

		// Normalize the URI (add file:// prefix if needed)
		uri = normalizeURI(uri)

		content, err := resolver.Resolve(ctx, uri)
		if err != nil {
			return "", fmt.Errorf("failed to resolve ref '%s' in dsn template: %w", uri, err)
		}

		return strings.TrimSpace(string(content)), nil
	}
}

// makeEnvFunc creates a template function that reads environment variables.
// Accepts variadic args to support both quoted and unquoted usage:
//   - ${env "MY_VAR"} - quoted string
//   - ${env MY_VAR} - unquoted
func makeEnvFunc(prefix string) func(...string) string {
	return func(parts ...string) string {
		if len(parts) == 0 {
			return ""
		}

		// Join parts to handle any split
		key := strings.Join(parts, " ")

		envKey := key
		if prefix != "" {
			envKey = prefix + key
		}

		return os.Getenv(envKey)
	}
}

// structToData converts a reflect.Value to an interface suitable for template execution.
// This preserves struct types so that nested field access works (e.g., ${.Database.Host}).
func structToData(v reflect.Value) any {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	// Return the struct directly to allow nested field access
	if v.CanInterface() {
		return v.Interface()
	}

	// Fallback to map for unexported structs
	return structToMap(v)
}

// structToMap converts a reflect.Value of a struct to a map[string]any.
// This is a fallback when the struct cannot be used directly.
func structToMap(v reflect.Value) map[string]any {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	result := make(map[string]any)
	t := v.Type()

	for i := range v.NumField() {
		field := t.Field(i)
		fieldVal := v.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get the interface value
		if fieldVal.CanInterface() {
			result[field.Name] = fieldVal.Interface()
		}
	}

	return result
}
