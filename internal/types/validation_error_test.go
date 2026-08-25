package types_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"

	"github.com/arloliu/fuda/internal/types"
)

// newYAMLNameValidator returns a validator that reports field names from
// yaml tags, mirroring how consumers typically register tag names so that
// error paths match their config file keys.
func newYAMLNameValidator() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.Split(field.Tag.Get("yaml"), ",")[0]
		if name == "" || name == "-" {
			return field.Name
		}

		return name
	})

	return v
}

// validate runs v.Struct and wraps the result the same way the loader does.
func wrapValidation(t *testing.T, v *validator.Validate, target any) *types.ValidationError {
	t.Helper()
	err := v.Struct(target)
	require.Error(t, err)

	return &types.ValidationError{Errors: []error{err}}
}

func TestValidationErrorRendersMaxOnNumericField(t *testing.T) {
	type pprofConfig struct {
		Port int `yaml:"port" validate:"max=65535"`
	}
	type discoveryConfig struct {
		Pprof pprofConfig `yaml:"pprof"`
	}
	type config struct {
		Discovery discoveryConfig `yaml:"discovery"`
	}

	verr := wrapValidation(t, newYAMLNameValidator(), &config{
		Discovery: discoveryConfig{Pprof: pprofConfig{Port: 70000}},
	})

	require.Equal(t, "validation failed: discovery.pprof.port: must be at most 65535", verr.Error())
}

func TestValidationErrorRendersMinOnNumericField(t *testing.T) {
	type config struct {
		Workers int `yaml:"workers" validate:"min=1"`
	}

	verr := wrapValidation(t, newYAMLNameValidator(), &config{Workers: 0})

	require.Equal(t, "validation failed: workers: must be at least 1", verr.Error())
}

func TestValidationErrorRendersMinMaxOnStringLength(t *testing.T) {
	type config struct {
		Name string `yaml:"name" validate:"min=3"`
	}

	verr := wrapValidation(t, newYAMLNameValidator(), &config{Name: "ab"})

	require.Equal(t, "validation failed: name: must be at least 3 characters", verr.Error())
}

func TestValidationErrorRendersMinOnSliceLength(t *testing.T) {
	type config struct {
		Hosts []string `yaml:"hosts" validate:"min=2"`
	}

	verr := wrapValidation(t, newYAMLNameValidator(), &config{Hosts: []string{"a"}})

	require.Equal(t, "validation failed: hosts: must have at least 2 items", verr.Error())
}

func TestValidationErrorRendersRequired(t *testing.T) {
	type config struct {
		Name string `yaml:"name" validate:"required"`
	}

	verr := wrapValidation(t, newYAMLNameValidator(), &config{})

	require.Equal(t, "validation failed: name: is required", verr.Error())
}

func TestValidationErrorRendersOneof(t *testing.T) {
	type config struct {
		Level string `yaml:"level" validate:"oneof=debug info warn error"`
	}

	verr := wrapValidation(t, newYAMLNameValidator(), &config{Level: "loud"})

	require.Equal(t, "validation failed: level: must be one of: debug, info, warn, error", verr.Error())
}

func TestValidationErrorRendersComparisonTags(t *testing.T) {
	type config struct {
		Gt  int `yaml:"gt" validate:"gt=0"`
		Gte int `yaml:"gte" validate:"gte=2"`
		Lt  int `yaml:"lt" validate:"lt=10"`
		Lte int `yaml:"lte" validate:"lte=5"`
	}

	verr := wrapValidation(t, newYAMLNameValidator(), &config{Gt: 0, Gte: 1, Lt: 10, Lte: 6})

	msg := verr.Error()
	require.Contains(t, msg, "gt: must be greater than 0")
	require.Contains(t, msg, "gte: must be at least 2")
	require.Contains(t, msg, "lt: must be less than 10")
	require.Contains(t, msg, "lte: must be at most 5")
}

func TestValidationErrorRendersHostnamePort(t *testing.T) {
	type config struct {
		Addr string `yaml:"addr" validate:"hostname_port"`
	}

	verr := wrapValidation(t, newYAMLNameValidator(), &config{Addr: "not a hostport"})

	require.Equal(t, "validation failed: addr: must be a host:port address", verr.Error())
}

func TestValidationErrorUnknownTagFallsBackToRawRendering(t *testing.T) {
	type config struct {
		Email string `yaml:"email" validate:"email"`
	}

	verr := wrapValidation(t, newYAMLNameValidator(), &config{Email: "nope"})

	msg := verr.Error()
	require.True(t, strings.HasPrefix(msg, "validation failed: "), msg)
	// The tag has no plain-language rendering, so the raw go-playground
	// message must survive unchanged.
	require.Contains(t, msg, "failed on the 'email' tag")
}

func TestValidationErrorMultipleFieldsRenderAsList(t *testing.T) {
	type config struct {
		Port int    `yaml:"port" validate:"max=65535"`
		Name string `yaml:"name" validate:"required"`
	}

	verr := wrapValidation(t, newYAMLNameValidator(), &config{Port: 70000})

	msg := verr.Error()
	require.True(t, strings.HasPrefix(msg, "validation failed:\n"), msg)
	require.Contains(t, msg, "  - port: must be at most 65535")
	require.Contains(t, msg, "  - name: is required")
}

func TestValidationErrorWithoutTagNameFuncUsesFieldNames(t *testing.T) {
	// Consumers that do not register a tag name function must still get
	// the plain-language message, with go's field names as the path.
	type config struct {
		Port int `yaml:"port" validate:"max=65535"`
	}

	verr := wrapValidation(t, validator.New(), &config{Port: 70000})

	require.Equal(t, "validation failed: Port: must be at most 65535", verr.Error())
}

func TestValidationErrorStillUnwrapsToValidatorErrors(t *testing.T) {
	type config struct {
		Port int `yaml:"port" validate:"max=65535"`
	}

	verr := wrapValidation(t, newYAMLNameValidator(), &config{Port: 70000})

	var raw validator.ValidationErrors
	require.True(t, errors.As(verr, &raw))
	require.Len(t, raw, 1)
	require.Equal(t, "max", raw[0].Tag())
	require.Equal(t, "65535", raw[0].Param())
}

func TestValidationErrorNonValidatorErrorsRenderUnchanged(t *testing.T) {
	verr := &types.ValidationError{Errors: []error{errors.New("custom failure")}}

	require.Equal(t, "validation failed: custom failure", verr.Error())
}
