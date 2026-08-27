package tests

import (
	"testing"

	"github.com/arloliu/fuda"
	"github.com/stretchr/testify/require"
)

func TestDocumentationFirstConfigurationPrecedence(t *testing.T) {
	type Config struct {
		Host string `yaml:"host" default:"localhost" env:"APP_HOST"`
		Port int    `yaml:"port" default:"8080" env:"APP_PORT"`
	}

	load := func(t *testing.T, source string) Config {
		t.Helper()

		loader, err := fuda.New().FromBytes([]byte(source)).Build()
		require.NoError(t, err)

		var cfg Config
		require.NoError(t, loader.Load(&cfg))
		return cfg
	}

	t.Run("uses the file before the default", func(t *testing.T) {
		cfg := load(t, "host: 0.0.0.0\n")

		require.Equal(t, "0.0.0.0", cfg.Host)
		require.Equal(t, 8080, cfg.Port)
	})

	t.Run("uses the environment before the file", func(t *testing.T) {
		t.Setenv("APP_PORT", "9090")
		cfg := load(t, "port: 3000\n")

		require.Equal(t, 9090, cfg.Port)
	})

	t.Run("uses defaults when no value is supplied", func(t *testing.T) {
		cfg := load(t, "")

		require.Equal(t, "localhost", cfg.Host)
		require.Equal(t, 8080, cfg.Port)
	})
}
