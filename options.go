package fuda

import "github.com/go-playground/validator/v10"

// config holds configuration for Fuda functions.
type config struct {
	validate  bool
	validator *validator.Validate
}

// Option configures Fuda behavior.
type Option func(*config)

// WithValidation enables or disables validation during SetDefaults.
// By default, SetDefaults does NOT perform validation (pure default setting).
// Pass WithValidation(true) to opt-in to validation after defaults are applied.
//
// Example:
//
//	// Defaults only (no validation)
//	fuda.SetDefaults(&cfg)
//
//	// Defaults + Validation
//	fuda.SetDefaults(&cfg, fuda.WithValidation(true))
func WithValidation(enabled bool) Option {
	return func(c *config) {
		c.validate = enabled
	}
}

// WithValidator sets a custom validator instance for SetDefaults or Validate.
// Use this to apply custom validation rules or tags.
//
// For SetDefaults:
// This option only takes effect if validation is enabled via WithValidation(true).
//
// For Validate:
// This option overrides the default validator.
//
// Example:
//
//	v := validator.New()
//	v.RegisterValidation("custom", customFunc)
//
//	fuda.SetDefaults(&cfg, fuda.WithValidation(true), fuda.WithValidator(v))
func WithValidator(v *validator.Validate) Option {
	return func(c *config) {
		c.validator = v
	}
}
