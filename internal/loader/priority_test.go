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

// TestGracefulFallbackChain tests the new "not found" vs "empty" distinction
func TestGracefulFallbackChain(t *testing.T) {
	// Test case 1: Unset env var falls back to default
	t.Run("UnsetEnvFallsBackToDefault", func(t *testing.T) {
		type Config struct {
			Secret string `ref:"env://UNSET_ENV_VAR_12345" default:"fallback-default"`
		}

		// Ensure env var is unset
		os.Unsetenv("UNSET_ENV_VAR_12345")

		e := &Engine{
			RefResolver: resolver.New(nil),
		}

		var cfg Config
		err := e.Load(&cfg)
		require.NoError(t, err, "Unset env var should not error")
		assert.Equal(t, "fallback-default", cfg.Secret, "Should use default when env is unset")
	})

	// Test case 2: Empty env var is used (stops fallback)
	t.Run("EmptyEnvStopsFallback", func(t *testing.T) {
		type Config struct {
			Secret string `ref:"env://EMPTY_ENV_VAR_TEST" default:"fallback-default"`
		}

		os.Setenv("EMPTY_ENV_VAR_TEST", "")
		defer os.Unsetenv("EMPTY_ENV_VAR_TEST")

		e := &Engine{
			RefResolver: resolver.New(nil),
		}

		var cfg Config
		err := e.Load(&cfg)
		require.NoError(t, err)
		assert.Equal(t, "", cfg.Secret, "Empty env var should be used, not fallback to default")
	})

	// Test case 3: Missing file falls back to default
	t.Run("MissingFileFallsBackToDefault", func(t *testing.T) {
		type Config struct {
			Secret string `ref:"file:///nonexistent/path/to/secret.txt" default:"file-fallback"`
		}

		e := &Engine{
			RefResolver: resolver.New(nil),
		}

		var cfg Config
		err := e.Load(&cfg)
		require.NoError(t, err, "Missing file should not error")
		assert.Equal(t, "file-fallback", cfg.Secret, "Should use default when file is missing")
	})

	// Test case 4: refFrom with unset env falls back to ref
	t.Run("RefFromUnsetEnvFallsBackToRef", func(t *testing.T) {
		type Config struct {
			SourcePath string `default:"env://UNSET_SOURCE_VAR"`
			Secret     string `refFrom:"SourcePath" ref:"env://FALLBACK_REF_VAR" default:"final-default"`
		}

		os.Unsetenv("UNSET_SOURCE_VAR")
		os.Setenv("FALLBACK_REF_VAR", "ref-fallback-value")
		defer os.Unsetenv("FALLBACK_REF_VAR")

		e := &Engine{
			RefResolver: resolver.New(nil),
		}

		var cfg Config
		err := e.Load(&cfg)
		require.NoError(t, err)
		assert.Equal(t, "ref-fallback-value", cfg.Secret, "Should fallback to ref when refFrom source resolves to 'not found'")
	})

	// Test case 5: Full chain - refFrom missing, ref missing, use default
	t.Run("FullChainFallbackToDefault", func(t *testing.T) {
		type Config struct {
			SourcePath string `default:"env://MISSING_SOURCE"`
			Secret     string `refFrom:"SourcePath" ref:"env://MISSING_REF" default:"ultimate-fallback"`
		}

		os.Unsetenv("MISSING_SOURCE")
		os.Unsetenv("MISSING_REF")

		e := &Engine{
			RefResolver: resolver.New(nil),
		}

		var cfg Config
		err := e.Load(&cfg)
		require.NoError(t, err)
		assert.Equal(t, "ultimate-fallback", cfg.Secret, "Should fallback to default when both refFrom and ref are missing")
	})

	// Test case 6: No fallback, field stays zero
	t.Run("NoFallbackFieldStaysZero", func(t *testing.T) {
		type Config struct {
			Secret string `ref:"env://MISSING_NO_DEFAULT"`
		}

		os.Unsetenv("MISSING_NO_DEFAULT")

		e := &Engine{
			RefResolver: resolver.New(nil),
		}

		var cfg Config
		err := e.Load(&cfg)
		require.NoError(t, err)
		assert.Equal(t, "", cfg.Secret, "Should leave field as zero when no default and ref is missing")
	})

	// Test case 7: refFrom empty string source falls back to ref
	t.Run("RefFromEmptySourceFallsBackToRef", func(t *testing.T) {
		type Config struct {
			SourcePath string // Empty string (not set)
			Secret     string `refFrom:"SourcePath" ref:"env://FALLBACK_VAL"`
		}

		os.Setenv("FALLBACK_VAL", "fallback-from-ref")
		defer os.Unsetenv("FALLBACK_VAL")

		e := &Engine{
			RefResolver: resolver.New(nil),
		}

		var cfg Config
		err := e.Load(&cfg)
		require.NoError(t, err)
		assert.Equal(t, "fallback-from-ref", cfg.Secret, "Empty refFrom source should fallback to ref")
	})
}
