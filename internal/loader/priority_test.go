package loader

import (
	"os"
	"testing"

	"github.com/arloliu/fuda/internal/resolver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvSchemeIntegration(t *testing.T) {
	// Test case 1: Basic env:// support
	t.Run("BasicEnvScheme", func(t *testing.T) {
		type Config struct {
			Secret string `ref:"env://TEST_ENV_SCHEME_BASIC"`
		}

		os.Setenv("TEST_ENV_SCHEME_BASIC", "secret-value")
		defer os.Unsetenv("TEST_ENV_SCHEME_BASIC")

		e := &Engine{
			RefResolver: resolver.New(nil), // Use default composite resolver
		}

		var cfg Config
		err := e.Load(&cfg)
		require.NoError(t, err)
		assert.Equal(t, "secret-value", cfg.Secret)
	})

	// Test case 2: env:// in refFrom
	t.Run("RefFromWithEnvScheme", func(t *testing.T) {
		type Config struct {
			SecretPath string `default:"env://TEST_ENV_SCHEME_FROM"`
			Secret     string `refFrom:"SecretPath"`
		}

		os.Setenv("TEST_ENV_SCHEME_FROM", "ref-from-value")
		defer os.Unsetenv("TEST_ENV_SCHEME_FROM")

		e := &Engine{
			RefResolver: resolver.New(nil),
		}

		var cfg Config
		err := e.Load(&cfg)
		require.NoError(t, err)
		assert.Equal(t, "ref-from-value", cfg.Secret)
	})

	// Test case 3: Priority check (User Question)
	// Scenario: APITokenPath points to non-existent file, but env var is set via env tag.
	// fuda should return the env var value and NOT error about the missing file.
	t.Run("PriorityEnvOverMissingRef", func(t *testing.T) {
		type Config struct {
			APITokenPath string `default:"file:///non/existent/path/token.txt"`
			APIToken     string `refFrom:"APITokenPath" env:"WORKFLOW_API_TOKEN"`
		}

		os.Setenv("WORKFLOW_API_TOKEN", "priority-token")
		defer os.Unsetenv("WORKFLOW_API_TOKEN")

		e := &Engine{
			RefResolver: resolver.New(nil),
		}

		var cfg Config
		err := e.Load(&cfg)
		require.NoError(t, err, "Should not return error even if file ref is missing, because env var is set")
		assert.Equal(t, "priority-token", cfg.APIToken)
	})

	// Test case 4: refFrom vs ref precedence (User Question)
	// Scenario: Field has both ref and refFrom.
	// Expected: refFrom takes precedence if the source field is non-empty.
	// If source field is empty, it falls back to ref.
	t.Run("RefFromOverridesRef", func(t *testing.T) {
		type Config struct {
			// Case A: Override
			OverridePath string `default:"env://OVERRIDE_VAL"`
			SecretA      string `ref:"env://DEFAULT_VAL" refFrom:"OverridePath"`

			// Case B: Fallback (refFrom field is empty)
			EmptyPath string // Empty
			SecretB   string `ref:"env://DEFAULT_VAL" refFrom:"EmptyPath"`
		}

		os.Setenv("DEFAULT_VAL", "default-value")
		os.Setenv("OVERRIDE_VAL", "override-value")
		defer os.Unsetenv("DEFAULT_VAL")
		defer os.Unsetenv("OVERRIDE_VAL")

		e := &Engine{
			RefResolver: resolver.New(nil),
		}

		var cfg Config
		err := e.Load(&cfg)
		require.NoError(t, err)

		// Case A: refFrom path ("env://OVERRIDE_VAL") should be used
		assert.Equal(t, "override-value", cfg.SecretA, "refFrom should override ref")

		// Case B: refFrom path is empty, should fallback to ref ("env://DEFAULT_VAL")
		assert.Equal(t, "default-value", cfg.SecretB, "Should fallback to ref if refFrom path is empty")
	})
}
