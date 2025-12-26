package fuda

import "github.com/arloliu/fuda/internal/types"

// FieldError represents an error that occurred while processing a specific field.
type FieldError = types.FieldError

// LoadError represents an error that occurred during the configuration loading process.
type LoadError = types.LoadError

// ValidationError wraps validation errors from the validator package.
type ValidationError = types.ValidationError
