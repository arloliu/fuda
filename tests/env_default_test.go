package tests

import (
	"testing"

	"github.com/arloliu/fuda"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnvOverridesDefault verifies that environment variables correctly override
// default values, even when the env value is a "zero" value like false, 0, or "".
// This was a bug where `default:"true"` would override `env:"VAR"` when VAR=false
// because false is the zero value for bool.
func TestEnvOverridesDefault(t *testing.T) {
	t.Run("env false overrides default true for bool", func(t *testing.T) {
		type Config struct {
			Insecure bool `default:"true" env:"TEST_INSECURE"`
		}

		t.Setenv("TEST_INSECURE", "false")

		var cfg Config
		err := fuda.LoadEnv(&cfg)
		require.NoError(t, err)
		assert.False(t, cfg.Insecure, "env 'false' should override default 'true'")
	})

	t.Run("env 0 overrides default non-zero for int", func(t *testing.T) {
		type Config struct {
			Port int `default:"8080" env:"TEST_PORT"`
		}

		t.Setenv("TEST_PORT", "0")

		var cfg Config
		err := fuda.LoadEnv(&cfg)
		require.NoError(t, err)
		assert.Equal(t, 0, cfg.Port, "env '0' should override default '8080'")
	})

	t.Run("env empty string overrides default non-empty for string", func(t *testing.T) {
		type Config struct {
			Host string `default:"localhost" env:"TEST_HOST"`
		}

		t.Setenv("TEST_HOST", "")

		var cfg Config
		err := fuda.LoadEnv(&cfg)
		require.NoError(t, err)
		assert.Equal(t, "", cfg.Host, "env '' should override default 'localhost'")
	})

	t.Run("default applies when env not set", func(t *testing.T) {
		type Config struct {
			Insecure bool   `default:"true" env:"TEST_UNSET_BOOL"`
			Port     int    `default:"8080" env:"TEST_UNSET_PORT"`
			Host     string `default:"localhost" env:"TEST_UNSET_HOST"`
		}

		var cfg Config
		err := fuda.LoadEnv(&cfg)
		require.NoError(t, err)
		assert.True(t, cfg.Insecure, "default 'true' should apply when env not set")
		assert.Equal(t, 8080, cfg.Port, "default '8080' should apply when env not set")
		assert.Equal(t, "localhost", cfg.Host, "default 'localhost' should apply when env not set")
	})

	t.Run("env overrides yaml value with zero value", func(t *testing.T) {
		type Config struct {
			Enabled bool `yaml:"enabled" env:"TEST_ENABLED"`
		}

		yaml := []byte(`enabled: true`)
		t.Setenv("TEST_ENABLED", "false")

		var cfg Config
		err := fuda.LoadBytes(yaml, &cfg)
		require.NoError(t, err)
		assert.False(t, cfg.Enabled, "env 'false' should override yaml 'true'")
	})

	t.Run("env overrides yaml and default with zero value", func(t *testing.T) {
		type Config struct {
			Debug bool `yaml:"debug" default:"true" env:"TEST_DEBUG"`
		}

		yaml := []byte(`debug: true`)
		t.Setenv("TEST_DEBUG", "false")

		var cfg Config
		err := fuda.LoadBytes(yaml, &cfg)
		require.NoError(t, err)
		assert.False(t, cfg.Debug, "env 'false' should override both yaml and default 'true'")
	})
}
