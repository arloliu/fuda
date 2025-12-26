package types

import (
	"fmt"
	"strings"
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
func (e *ValidationError) Error() string {
	if len(e.Errors) == 0 {
		return "validation failed"
	}
	if len(e.Errors) == 1 {
		return fmt.Sprintf("validation failed: %v", e.Errors[0])
	}

	var sb strings.Builder
	sb.WriteString("validation failed:\n")
	for i, err := range e.Errors {
		sb.WriteString("  - ")
		sb.WriteString(err.Error())
		if i < len(e.Errors)-1 {
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
