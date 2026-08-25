package types

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// FieldError represents an error that occurred while processing a specific field.
type FieldError struct {
	Path    string // e.g., "Database.Port"
	Tag     string // e.g., "env", "default"
	Value   string // the invalid value
	Message string
	Err     error
}

// Error returns the string representation of the FieldError.
func (e *FieldError) Error() string {
	var sb strings.Builder
	sb.WriteString("field '")
	sb.WriteString(e.Path)
	sb.WriteString("'")

	if e.Tag != "" {
		sb.WriteString(" (tag '")
		sb.WriteString(e.Tag)
		sb.WriteString("')")
	}

	if e.Value != "" {
		sb.WriteString(": invalid value '")
		sb.WriteString(e.Value)
		sb.WriteString("'")
	}

	if e.Message != "" {
		sb.WriteString(": ")
		sb.WriteString(e.Message)
	}

	if e.Err != nil {
		sb.WriteString(": ")
		sb.WriteString(e.Err.Error())
	}

	return sb.String()
}

// Unwrap returns the underlying error.
func (e *FieldError) Unwrap() error {
	return e.Err
}

// LoadError represents an error that occurred during the configuration loading process.
type LoadError struct {
	Source string // file path or source name
	Errors []FieldError
}

// Error returns the string representation of the LoadError.
func (e *LoadError) Error() string {
	var sb strings.Builder
	sb.WriteString("failed to load configuration")
	if e.Source != "" {
		sb.WriteString(" from ")
		sb.WriteString(e.Source)
	}
	sb.WriteString(":\n")

	for i, err := range e.Errors {
		sb.WriteString("  ")
		sb.WriteString(err.Error())
		if i < len(e.Errors)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// ValidationError wraps validation errors from the validator package.
type ValidationError struct {
	Errors []error
}

// Error returns the string representation of the ValidationError.
//
// Errors coming from go-playground/validator are rendered as one
// "<path>: <plain-language message>" line per failed field, e.g.
// "discovery.pprof.port: must be at most 65535". Tags without a
// plain-language rendering keep the validator's own message unchanged.
// This rendered form is a documented compatibility surface; consumers
// that need structured access should use errors.As to retrieve the
// underlying validator.ValidationErrors instead of parsing the string.
func (e *ValidationError) Error() string {
	lines := e.renderLines()
	if len(lines) == 0 {
		return "validation failed"
	}
	if len(lines) == 1 {
		return "validation failed: " + lines[0]
	}

	var sb strings.Builder
	sb.WriteString("validation failed:\n")
	for i, line := range lines {
		sb.WriteString("  - ")
		sb.WriteString(line)
		if i < len(lines)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// Unwrap returns the first error in the list.
func (e *ValidationError) Unwrap() error {
	if len(e.Errors) > 0 {
		return e.Errors[0]
	}

	return nil
}

// renderLines flattens the wrapped errors into display lines, expanding
// validator.ValidationErrors into one line per failed field.
func (e *ValidationError) renderLines() []string {
	lines := make([]string, 0, len(e.Errors))
	for _, err := range e.Errors {
		var verrs validator.ValidationErrors
		if errors.As(err, &verrs) {
			for _, fe := range verrs {
				lines = append(lines, renderFieldError(fe))
			}

			continue
		}

		lines = append(lines, err.Error())
	}

	return lines
}

// renderFieldError renders one validator field error as
// "<path>: <plain-language message>". The path is the validator
// namespace with the leading struct name removed, so it matches the
// field names a consumer's RegisterTagNameFunc reports (typically the
// config file keys). Tags without a plain-language rendering fall back
// to the validator's own message, which carries its own path.
func renderFieldError(fe validator.FieldError) string {
	msg, ok := renderValidationTag(fe)
	if !ok {
		return fe.Error()
	}

	path := fe.Namespace()
	if i := strings.Index(path, "."); i >= 0 {
		path = path[i+1:]
	} else {
		path = fe.Field()
	}

	return path + ": " + msg
}

// renderValidationTag translates common validation tags into plain
// statements using the tag parameter, e.g. max=65535 on a numeric field
// becomes "must be at most 65535". It reports false for tags it does not
// know how to render.
func renderValidationTag(fe validator.FieldError) (string, bool) {
	param := fe.Param()
	switch fe.Tag() {
	case "required":
		return "is required", true
	case "min", "gte":
		return renderBound(fe.Kind(), param, "at least")
	case "max", "lte":
		return renderBound(fe.Kind(), param, "at most")
	case "gt":
		return renderExclusiveBound(fe.Kind(), param, "greater than", "more than", "more than")
	case "lt":
		return renderExclusiveBound(fe.Kind(), param, "less than", "less than", "fewer than")
	case "oneof":
		// Quoted values may contain spaces; leave those to the raw message
		// rather than splitting them incorrectly.
		if !strings.Contains(param, "'") {
			return "must be one of: " + strings.Join(strings.Fields(param), ", "), true
		}
	case "hostname_port":
		return "must be a host:port address", true
	}

	return "", false
}

// renderBound phrases an inclusive bound ("at least"/"at most") for the
// field kind: plain value for numbers, characters for strings, items for
// slices, arrays, and maps.
func renderBound(kind reflect.Kind, param, bound string) (string, bool) {
	switch {
	case isNumericKind(kind):
		return "must be " + bound + " " + param, true
	case kind == reflect.String:
		return "must be " + bound + " " + param + " characters", true
	case kind == reflect.Slice || kind == reflect.Array || kind == reflect.Map:
		return "must have " + bound + " " + param + " items", true
	default:
		return "", false
	}
}

// renderExclusiveBound phrases a strict bound (gt/lt) for the field kind:
// plain value for numbers, characters for strings, items for slices,
// arrays, and maps. numericPhrase, sizePhrase, and itemsPhrase carry the
// comparison wording for each kind category, since gt and lt phrase
// collections differently ("more than" vs "fewer than") even though
// both phrase numbers with a single word ("greater than" / "less than").
func renderExclusiveBound(kind reflect.Kind, param, numericPhrase, sizePhrase, itemsPhrase string) (string, bool) {
	switch {
	case isNumericKind(kind):
		return "must be " + numericPhrase + " " + param, true
	case kind == reflect.String:
		return "must be " + sizePhrase + " " + param + " characters", true
	case kind == reflect.Slice || kind == reflect.Array || kind == reflect.Map:
		return "must have " + itemsPhrase + " " + param + " items", true
	default:
		return "", false
	}
}

// isNumericKind reports whether the kind is an integer, unsigned
// integer, or float.
func isNumericKind(kind reflect.Kind) bool {
	//nolint:exhaustive // Only numeric kinds return true.
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}
