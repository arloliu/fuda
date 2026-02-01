package tests

import (
	"testing"

	"github.com/arloliu/fuda"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
)

type ValidationConfig struct {
	Name string `default:"default_name" validate:"required"`
	Age  int    `default:"0" validate:"gt=0"`
}

func TestSetDefaults_NoValidationByDefault(t *testing.T) {
	cfg := &ValidationConfig{}

	// Should NOT return error even though Age is 0 (which is not gt=0)
	err := fuda.SetDefaults(cfg)
	require.NoError(t, err, "SetDefaults should not validate by default")
	require.Equal(t, "default_name", cfg.Name)
	require.Equal(t, 0, cfg.Age)
}

func TestSetDefaults_WithValidation(t *testing.T) {
	cfg := &ValidationConfig{}

	// Should return error because Age is 0 which fails gt=0 validation
	err := fuda.SetDefaults(cfg, fuda.WithValidation(true))
	require.Error(t, err, "SetDefaults should validate when option is enabled")
	require.Contains(t, err.Error(), "Age")
}

func TestMustSetDefaults_NoValidationByDefault(t *testing.T) {
	cfg := &ValidationConfig{}

	// Should not panic
	require.NotPanics(t, func() {
		fuda.MustSetDefaults(cfg)
	})
}

func TestMustSetDefaults_WithValidation(t *testing.T) {
	cfg := &ValidationConfig{}

	// Should panic because of validation error
	require.Panics(t, func() {
		fuda.MustSetDefaults(cfg, fuda.WithValidation(true))
	})
}

func TestSetDefaults_WithCustomValidator(t *testing.T) {
	// Create custom validator that fails on "default_name"
	// Default validator only checks `validate` tags
	v := validator.New()
	_ = v.RegisterValidation("not_default", func(fl validator.FieldLevel) bool {
		return fl.Field().String() != "default_name"
	})

	type CustomConfig struct {
		Name string `default:"default_name" validate:"not_default"` //nolint:revive // custom validation
	}

	c := &CustomConfig{}

	// Should fail because Name is "default_name" and we use custom validator
	err := fuda.SetDefaults(c, fuda.WithValidation(true), fuda.WithValidator(v))
	require.Error(t, err)
	require.Contains(t, err.Error(), "not_default")
}

func TestValidate_WithCustomValidator(t *testing.T) {
	// Custom validator
	v := validator.New()
	_ = v.RegisterValidation("is_cool", func(fl validator.FieldLevel) bool {
		return fl.Field().String() == "cool"
	})

	type CoolConfig struct {
		Status string `validate:"is_cool"` //nolint:revive // custom validation
	}

	// Default Validate should fail finding the tag or fail validation?
	// Actually default validator doesn't know "is_cool", it might error or panic depending on config.
	// But let's test positive case where custom validator works.

	c2 := &CoolConfig{Status: "cool"}
	err := fuda.Validate(c2, fuda.WithValidator(v))
	require.NoError(t, err)

	c3 := &CoolConfig{Status: "uncool"}
	err = fuda.Validate(c3, fuda.WithValidator(v))
	require.Error(t, err)
	require.Contains(t, err.Error(), "is_cool")
}
