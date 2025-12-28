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
)

// TemplateConfig holds configuration for template processing.
type TemplateConfig struct {
	// Strict controls error behavior for empty/undefined values.
	// If true, returns an error when a referenced field is empty or undefined.
	// If false (default), outputs empty string for missing values.
	Strict bool

	// Resolver for ${ref:uri} function in templates.
	Resolver Resolver

	// EnvPrefix for ${env:KEY} function in templates.
	EnvPrefix string
}

// ProcessTemplate expands ${...} template expressions in a string.
//
// Template syntax (uses ${...} delimiters):
//   - ${.FieldName} - references the value of a field in the struct data
//   - ${.Nested.Field} - references nested struct fields
//   - ${ref:uri} or ${ref "uri"} - resolves a URI inline using the resolver
//   - ${env:KEY} or ${env "KEY"} - reads an environment variable
//
// Note: Fields referenced in templates must appear earlier in the struct
// to have their values available (due to sequential field processing).
func ProcessTemplate(ctx context.Context, templateStr string, data any, config TemplateConfig) (string, error) {
	// Preprocess the template to convert shorthand syntax to template function calls
	// ${ref:uri} -> ${ref "uri"}
	// ${env:KEY} -> ${env "KEY"}
	processedTemplate := preprocessTemplate(templateStr)

	// Build template with custom functions and ${...} delimiters
	funcMap := template.FuncMap{
		"ref": makeRefFunc(ctx, config.Resolver),
		"env": makeEnvFunc(config.EnvPrefix),
	}

	// Configure missing key behavior based on strict mode
	missingKeyOpt := "missingkey=zero" // Default: return zero value
	if config.Strict {
		missingKeyOpt = "missingkey=error" // Strict: return error on missing field
	}

	tmpl, err := template.New("template").
		Delims("${", "}").
		Funcs(funcMap).
		Option(missingKeyOpt).
		Parse(processedTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		// In strict mode, missing keys return an exec error
		if config.Strict && strings.Contains(err.Error(), "map has no entry for key") {
			return "", fmt.Errorf("template references undefined field: %w", err)
		}
		// Also handle the case where a struct field is missing (nil pointer field etc)
		if config.Strict && strings.Contains(err.Error(), "nil pointer evaluating") {
			return "", fmt.Errorf("template references nil pointer field: %w", err)
		}

		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// preprocessTemplate converts shorthand syntax to Go template function calls.
//   - ${ref:uri} -> ${ref "uri"}
//   - ${env:KEY} -> ${env "KEY"}
//
// This allows users to write clean, unquoted URIs similar to the ref tag.
func preprocessTemplate(templateStr string) string {
	result := preprocessFunc(templateStr, "ref")
	result = preprocessFunc(result, "env")
	return result
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

// findClosingBrace finds the index of the closing } for a template expression.
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
			return "", errors.New("no resolver configured for ref function in template")
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
			return "", fmt.Errorf("failed to resolve ref '%s' in template: %w", uri, err)
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

// StructToData converts a reflect.Value to an interface suitable for template execution.
// This preserves struct types so that nested field access works (e.g., ${.Database.Host}).
func StructToData(v reflect.Value) any {
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
