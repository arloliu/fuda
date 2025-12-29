package tests

import (
	"os"
	"testing"

	"github.com/arloliu/fuda"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithOverrides(t *testing.T) {
	t.Run("top-level override", func(t *testing.T) {
		type Config struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
		}

		yamlContent := `
host: file-host
port: 8080
`
		loader, err := fuda.New().
			FromBytes([]byte(yamlContent)).
			WithOverrides(map[string]any{
				"host": "override-host",
			}).
			Build()
		require.NoError(t, err)

		var cfg Config
		err = loader.Load(&cfg)
		require.NoError(t, err)

		assert.Equal(t, "override-host", cfg.Host)
		assert.Equal(t, 8080, cfg.Port) // Unchanged
	})

	t.Run("nested key dot notation", func(t *testing.T) {
		type Database struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
		}
		type Config struct {
			Database Database `yaml:"database"`
		}

		yamlContent := `
database:
  host: localhost
  port: 5432
`
		loader, err := fuda.New().
			FromBytes([]byte(yamlContent)).
			WithOverrides(map[string]any{
				"database.port": 6543,
			}).
			Build()
		require.NoError(t, err)

		var cfg Config
		err = loader.Load(&cfg)
		require.NoError(t, err)

		assert.Equal(t, "localhost", cfg.Database.Host) // Unchanged
		assert.Equal(t, 6543, cfg.Database.Port)        // Overridden
	})

	t.Run("add new key", func(t *testing.T) {
		type Config struct {
			Host  string `yaml:"host"`
			Debug bool   `yaml:"debug"`
		}

		yamlContent := `
host: localhost
`
		loader, err := fuda.New().
			FromBytes([]byte(yamlContent)).
			WithOverrides(map[string]any{
				"debug": true,
			}).
			Build()
		require.NoError(t, err)

		var cfg Config
		err = loader.Load(&cfg)
		require.NoError(t, err)

		assert.Equal(t, "localhost", cfg.Host)
		assert.True(t, cfg.Debug)
	})

	t.Run("env beats override", func(t *testing.T) {
		type Config struct {
			Host string `yaml:"host" env:"TEST_OVERRIDE_HOST"`
		}

		os.Setenv("TEST_OVERRIDE_HOST", "env-host")
		defer os.Unsetenv("TEST_OVERRIDE_HOST")

		yamlContent := `
host: file-host
`
		loader, err := fuda.New().
			FromBytes([]byte(yamlContent)).
			WithOverrides(map[string]any{
				"host": "override-host",
			}).
			Build()
		require.NoError(t, err)

		var cfg Config
		err = loader.Load(&cfg)
		require.NoError(t, err)

		assert.Equal(t, "env-host", cfg.Host) // Env takes precedence
	})

	t.Run("override beats file", func(t *testing.T) {
		type Config struct {
			Port int `yaml:"port"`
		}

		yamlContent := `
port: 8080
`
		loader, err := fuda.New().
			FromBytes([]byte(yamlContent)).
			WithOverrides(map[string]any{
				"port": 9090,
			}).
			Build()
		require.NoError(t, err)

		var cfg Config
		err = loader.Load(&cfg)
		require.NoError(t, err)

		assert.Equal(t, 9090, cfg.Port)
	})

	t.Run("deep nested override", func(t *testing.T) {
		type Level4 struct {
			Value string `yaml:"value"`
		}
		type Level3 struct {
			Level4 Level4 `yaml:"level4"`
		}
		type Level2 struct {
			Level3 Level3 `yaml:"level3"`
		}
		type Level1 struct {
			Level2 Level2 `yaml:"level2"`
		}
		type Config struct {
			Level1 Level1 `yaml:"level1"`
		}

		yamlContent := `
level1:
  level2:
    level3:
      level4:
        value: original
`
		loader, err := fuda.New().
			FromBytes([]byte(yamlContent)).
			WithOverrides(map[string]any{
				"level1.level2.level3.level4.value": "overridden",
			}).
			Build()
		require.NoError(t, err)

		var cfg Config
		err = loader.Load(&cfg)
		require.NoError(t, err)

		assert.Equal(t, "overridden", cfg.Level1.Level2.Level3.Level4.Value)
	})

	t.Run("override array", func(t *testing.T) {
		type Config struct {
			Tags []string `yaml:"tags"`
		}

		yamlContent := `
tags:
  - original1
  - original2
`
		loader, err := fuda.New().
			FromBytes([]byte(yamlContent)).
			WithOverrides(map[string]any{
				"tags": []string{"new1", "new2", "new3"},
			}).
			Build()
		require.NoError(t, err)

		var cfg Config
		err = loader.Load(&cfg)
		require.NoError(t, err)

		assert.Equal(t, []string{"new1", "new2", "new3"}, cfg.Tags)
	})

	t.Run("multiple overrides", func(t *testing.T) {
		type Config struct {
			Host  string `yaml:"host"`
			Port  int    `yaml:"port"`
			Debug bool   `yaml:"debug"`
		}

		yamlContent := `
host: original-host
port: 8080
debug: false
`
		loader, err := fuda.New().
			FromBytes([]byte(yamlContent)).
			WithOverrides(map[string]any{
				"host":  "new-host",
				"port":  9090,
				"debug": true,
			}).
			Build()
		require.NoError(t, err)

		var cfg Config
		err = loader.Load(&cfg)
		require.NoError(t, err)

		assert.Equal(t, "new-host", cfg.Host)
		assert.Equal(t, 9090, cfg.Port)
		assert.True(t, cfg.Debug)
	})

	t.Run("empty overrides map", func(t *testing.T) {
		type Config struct {
			Host string `yaml:"host"`
		}

		yamlContent := `
host: original-host
`
		loader, err := fuda.New().
			FromBytes([]byte(yamlContent)).
			WithOverrides(map[string]any{}).
			Build()
		require.NoError(t, err)

		var cfg Config
		err = loader.Load(&cfg)
		require.NoError(t, err)

		assert.Equal(t, "original-host", cfg.Host)
	})

	t.Run("nil overrides map", func(t *testing.T) {
		type Config struct {
			Host string `yaml:"host"`
		}

		yamlContent := `
host: original-host
`
		loader, err := fuda.New().
			FromBytes([]byte(yamlContent)).
			WithOverrides(nil).
			Build()
		require.NoError(t, err)

		var cfg Config
		err = loader.Load(&cfg)
		require.NoError(t, err)

		assert.Equal(t, "original-host", cfg.Host)
	})

	t.Run("override with nil value", func(t *testing.T) {
		type Config struct {
			Host *string `yaml:"host"`
		}

		yamlContent := `
host: original-host
`
		loader, err := fuda.New().
			FromBytes([]byte(yamlContent)).
			WithOverrides(map[string]any{
				"host": nil,
			}).
			Build()
		require.NoError(t, err)

		var cfg Config
		err = loader.Load(&cfg)
		require.NoError(t, err)

		assert.Nil(t, cfg.Host)
	})

	t.Run("create nested structure from nothing", func(t *testing.T) {
		type Database struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
		}
		type Config struct {
			Database Database `yaml:"database"`
		}

		// Empty source
		loader, err := fuda.New().
			FromBytes([]byte("")).
			WithOverrides(map[string]any{
				"database.host": "new-host",
				"database.port": 5432,
			}).
			Build()
		require.NoError(t, err)

		var cfg Config
		err = loader.Load(&cfg)
		require.NoError(t, err)

		assert.Equal(t, "new-host", cfg.Database.Host)
		assert.Equal(t, 5432, cfg.Database.Port)
	})

	t.Run("override works with template", func(t *testing.T) {
		type Config struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
		}

		yamlContent := `
host: {{ .Host }}
port: 8080
`
		type TmplData struct {
			Host string
		}

		loader, err := fuda.New().
			FromBytes([]byte(yamlContent)).
			WithTemplate(TmplData{Host: "template-host"}).
			WithOverrides(map[string]any{
				"port": 9090,
			}).
			Build()
		require.NoError(t, err)

		var cfg Config
		err = loader.Load(&cfg)
		require.NoError(t, err)

		assert.Equal(t, "template-host", cfg.Host) // From template
		assert.Equal(t, 9090, cfg.Port)            // Overridden
	})

	t.Run("override with default fallback", func(t *testing.T) {
		type Config struct {
			Host    string `yaml:"host" default:"default-host"`
			Port    int    `yaml:"port" default:"8080"`
			Timeout int    `yaml:"timeout" default:"30"`
		}

		// No source, only overrides and defaults
		loader, err := fuda.New().
			WithOverrides(map[string]any{
				"host": "override-host",
			}).
			Build()
		require.NoError(t, err)

		var cfg Config
		err = loader.Load(&cfg)
		require.NoError(t, err)

		assert.Equal(t, "override-host", cfg.Host) // Overridden
		assert.Equal(t, 8080, cfg.Port)            // Default (no override)
		assert.Equal(t, 30, cfg.Timeout)           // Default (no override)
	})
}
