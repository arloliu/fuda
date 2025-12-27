package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arloliu/fuda"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWithDotEnv_SingleFile verifies loading a single .env file.
func TestWithDotEnv_SingleFile(t *testing.T) {
	// Create temp .env file
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(envPath, []byte("DOTENV_HOST=dotenv-host\nDOTENV_PORT=9999\n"), 0o600)
	require.NoError(t, err)

	type Config struct {
		Host string `env:"DOTENV_HOST"`
		Port int    `env:"DOTENV_PORT" default:"8080"`
	}

	loader, err := fuda.New().
		WithDotEnv(envPath).
		Build()
	require.NoError(t, err)

	var cfg Config
	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "dotenv-host", cfg.Host)
	assert.Equal(t, 9999, cfg.Port)
}

// TestWithDotEnvFiles_MultipleFiles verifies overlay pattern with multiple files.
func TestWithDotEnvFiles_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Base .env file
	basePath := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(basePath, []byte("MULTI_HOST=base-host\nMULTI_PORT=8080\n"), 0o600)
	require.NoError(t, err)

	// Local overlay (adds DB_HOST, does not override existing)
	localPath := filepath.Join(tmpDir, ".env.local")
	err = os.WriteFile(localPath, []byte("MULTI_DB_HOST=localhost\n"), 0o600)
	require.NoError(t, err)

	type Config struct {
		Host   string `env:"MULTI_HOST"`
		Port   int    `env:"MULTI_PORT"`
		DBHost string `env:"MULTI_DB_HOST"`
	}

	loader, err := fuda.New().
		WithDotEnvFiles([]string{basePath, localPath}).
		Build()
	require.NoError(t, err)

	var cfg Config
	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "base-host", cfg.Host)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "localhost", cfg.DBHost)
}

// TestWithDotEnvSearch_DirectorySearch verifies searching for .env in directories.
func TestWithDotEnvSearch_DirectorySearch(t *testing.T) {
	tmpDir := t.TempDir()

	// Create subdirectory with .env
	configDir := filepath.Join(tmpDir, "config")
	require.NoError(t, os.MkdirAll(configDir, 0o755))

	envPath := filepath.Join(configDir, ".env")
	err := os.WriteFile(envPath, []byte("SEARCH_VAR=found-in-search\n"), 0o600)
	require.NoError(t, err)

	type Config struct {
		Var string `env:"SEARCH_VAR"`
	}

	loader, err := fuda.New().
		WithDotEnvSearch(".env", []string{tmpDir, configDir}).
		Build()
	require.NoError(t, err)

	var cfg Config
	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "found-in-search", cfg.Var)
}

// TestDotEnvOverride_OverwritesExistingEnv verifies override mode behavior.
func TestDotEnvOverride_OverwritesExistingEnv(t *testing.T) {
	// Set real env var first
	require.NoError(t, os.Setenv("OVERRIDE_TEST_VAR", "real-env-value"))
	defer os.Unsetenv("OVERRIDE_TEST_VAR")

	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(envPath, []byte("OVERRIDE_TEST_VAR=dotenv-value\n"), 0o600)
	require.NoError(t, err)

	type Config struct {
		Var string `env:"OVERRIDE_TEST_VAR"`
	}

	// Without override - real env takes precedence
	loader, err := fuda.New().
		WithDotEnv(envPath).
		Build()
	require.NoError(t, err)

	var cfg1 Config
	err = loader.Load(&cfg1)
	require.NoError(t, err)
	assert.Equal(t, "real-env-value", cfg1.Var)

	// Reset the env var (godotenv.Load doesn't overwrite)
	require.NoError(t, os.Unsetenv("OVERRIDE_TEST_VAR"))
	require.NoError(t, os.Setenv("OVERRIDE_TEST_VAR", "real-env-value"))

	// With override - dotenv value takes precedence
	loaderOverride, err := fuda.New().
		WithDotEnv(envPath, fuda.DotEnvOverride()).
		Build()
	require.NoError(t, err)

	var cfg2 Config
	err = loaderOverride.Load(&cfg2)
	require.NoError(t, err)
	assert.Equal(t, "dotenv-value", cfg2.Var)
}

// TestDotEnv_RealEnvTakesPrecedence verifies default precedence behavior.
func TestDotEnv_RealEnvTakesPrecedence(t *testing.T) {
	// Set real env var first
	require.NoError(t, os.Setenv("PRECEDENCE_VAR", "real-value"))
	defer os.Unsetenv("PRECEDENCE_VAR")

	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(envPath, []byte("PRECEDENCE_VAR=dotenv-value\n"), 0o600)
	require.NoError(t, err)

	type Config struct {
		Var string `env:"PRECEDENCE_VAR"`
	}

	loader, err := fuda.New().
		WithDotEnv(envPath).
		Build()
	require.NoError(t, err)

	var cfg Config
	err = loader.Load(&cfg)
	require.NoError(t, err)

	// Real env var should take precedence
	assert.Equal(t, "real-value", cfg.Var)
}

// TestDotEnv_MissingFileIgnored verifies graceful handling of missing files.
func TestDotEnv_MissingFileIgnored(t *testing.T) {
	type Config struct {
		Host string `default:"default-host"`
	}

	loader, err := fuda.New().
		WithDotEnv("/nonexistent/path/.env").
		Build()
	require.NoError(t, err)

	var cfg Config
	err = loader.Load(&cfg)
	require.NoError(t, err)

	// Should fall back to default
	assert.Equal(t, "default-host", cfg.Host)
}

// TestDotEnv_IntegrationWithEnvTag verifies full integration with struct env tags.
func TestDotEnv_IntegrationWithEnvTag(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config file
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte("database:\n  host: yaml-host\n"), 0o600)
	require.NoError(t, err)

	// Create .env file
	envPath := filepath.Join(tmpDir, ".env")
	err = os.WriteFile(envPath, []byte("DB_PORT=5432\nDB_USER=admin\n"), 0o600)
	require.NoError(t, err)

	type DatabaseConfig struct {
		Host string `yaml:"host" env:"DB_HOST"`
		Port int    `env:"DB_PORT" default:"3306"`
		User string `env:"DB_USER"`
	}
	type Config struct {
		Database DatabaseConfig `yaml:"database"`
	}

	loader, err := fuda.New().
		FromFile(configPath).
		WithDotEnv(envPath).
		Build()
	require.NoError(t, err)

	var cfg Config
	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "yaml-host", cfg.Database.Host) // From YAML
	assert.Equal(t, 5432, cfg.Database.Port)        // From dotenv
	assert.Equal(t, "admin", cfg.Database.User)     // From dotenv
}

// TestDotEnv_WithEnvPrefix verifies dotenv works with env prefix.
func TestDotEnv_WithEnvPrefix(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(envPath, []byte("APP_HOST=prefixed-dotenv-host\nAPP_PORT=7777\n"), 0o600)
	require.NoError(t, err)

	type Config struct {
		Host string `env:"HOST"`
		Port int    `env:"PORT"`
	}

	loader, err := fuda.New().
		WithDotEnv(envPath).
		WithEnvPrefix("APP_").
		Build()
	require.NoError(t, err)

	var cfg Config
	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "prefixed-dotenv-host", cfg.Host)
	assert.Equal(t, 7777, cfg.Port)
}

// TestWithDotEnvSearch_FirstMatchWins verifies search stops at first match.
func TestWithDotEnvSearch_FirstMatchWins(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .env in first search path
	dir1 := filepath.Join(tmpDir, "dir1")
	require.NoError(t, os.MkdirAll(dir1, 0o755))
	env1 := filepath.Join(dir1, ".env")
	err := os.WriteFile(env1, []byte("SEARCH_VAR1=from-dir1\nSHARED_VAR=dir1\n"), 0o600)
	require.NoError(t, err)

	// Create .env in second search path
	dir2 := filepath.Join(tmpDir, "dir2")
	require.NoError(t, os.MkdirAll(dir2, 0o755))
	env2 := filepath.Join(dir2, ".env")
	err = os.WriteFile(env2, []byte("SEARCH_VAR2=from-dir2\nSHARED_VAR=dir2\n"), 0o600)
	require.NoError(t, err)

	type Config struct {
		Var1   string `env:"SEARCH_VAR1"`
		Var2   string `env:"SEARCH_VAR2"`
		Shared string `env:"SHARED_VAR"`
	}

	loader, err := fuda.New().
		// Search dir1 then dir2
		WithDotEnvSearch(".env", []string{dir1, dir2}).
		Build()
	require.NoError(t, err)

	var cfg Config
	err = loader.Load(&cfg)
	require.NoError(t, err)

	// Only first file should be loaded
	assert.Equal(t, "from-dir1", cfg.Var1)
	assert.Empty(t, cfg.Var2)           // Should NOT be loaded
	assert.Equal(t, "dir1", cfg.Shared) // First one wins
}

// TestDotEnv_EmptyFile verifies handling of empty .env file.
func TestDotEnv_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(envPath, []byte(""), 0o600)
	require.NoError(t, err)

	type Config struct {
		Host string `default:"default-host"`
	}

	loader, err := fuda.New().
		WithDotEnv(envPath).
		Build()
	require.NoError(t, err)

	var cfg Config
	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "default-host", cfg.Host)
}

// TestDotEnv_CommentsAndExports verifies godotenv comment and export support.
func TestDotEnv_CommentsAndExports(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	content := `# This is a comment
COMMENT_VAR1=value1
export COMMENT_VAR2=value2
COMMENT_VAR3=value3 # inline comment
`
	err := os.WriteFile(envPath, []byte(content), 0o600)
	require.NoError(t, err)

	type Config struct {
		Var1 string `env:"COMMENT_VAR1"`
		Var2 string `env:"COMMENT_VAR2"`
		Var3 string `env:"COMMENT_VAR3"`
	}

	loader, err := fuda.New().
		WithDotEnv(envPath).
		Build()
	require.NoError(t, err)

	var cfg Config
	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "value1", cfg.Var1)
	assert.Equal(t, "value2", cfg.Var2)
	assert.Equal(t, "value3", cfg.Var3)
}

// TestDotEnv_QuotedValuesWithSpaces verifies quoted values are handled correctly.
func TestDotEnv_QuotedValuesWithSpaces(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	content := `QUOTED_VAR1="hello world"
QUOTED_VAR2='single quotes'
QUOTED_VAR3=no_quotes
`
	err := os.WriteFile(envPath, []byte(content), 0o600)
	require.NoError(t, err)

	type Config struct {
		Var1 string `env:"QUOTED_VAR1"`
		Var2 string `env:"QUOTED_VAR2"`
		Var3 string `env:"QUOTED_VAR3"`
	}

	loader, err := fuda.New().
		WithDotEnv(envPath).
		Build()
	require.NoError(t, err)

	var cfg Config
	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "hello world", cfg.Var1)
	assert.Equal(t, "single quotes", cfg.Var2)
	assert.Equal(t, "no_quotes", cfg.Var3)
}

// TestWithDotEnvFiles_WithOverrideOption verifies WithDotEnvFiles accepts DotEnvOverride option.
func TestWithDotEnvFiles_WithOverrideOption(t *testing.T) {
	// Set real env var first
	require.NoError(t, os.Setenv("FILES_OVERRIDE_VAR", "real-value"))
	defer os.Unsetenv("FILES_OVERRIDE_VAR")

	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(envPath, []byte("FILES_OVERRIDE_VAR=dotenv-value\n"), 0o600)
	require.NoError(t, err)

	type Config struct {
		Var string `env:"FILES_OVERRIDE_VAR"`
	}

	// With override - dotenv should win
	loader, err := fuda.New().
		WithDotEnvFiles([]string{envPath}, fuda.DotEnvOverride()).
		Build()
	require.NoError(t, err)

	var cfg Config
	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "dotenv-value", cfg.Var)
}

// TestWithDotEnvSearch_WithOverrideOption verifies WithDotEnvSearch accepts DotEnvOverride option.
func TestWithDotEnvSearch_WithOverrideOption(t *testing.T) {
	// Set real env var first
	require.NoError(t, os.Setenv("SEARCH_OVERRIDE_VAR", "real-value"))
	defer os.Unsetenv("SEARCH_OVERRIDE_VAR")

	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(envPath, []byte("SEARCH_OVERRIDE_VAR=dotenv-value\n"), 0o600)
	require.NoError(t, err)

	type Config struct {
		Var string `env:"SEARCH_OVERRIDE_VAR"`
	}

	// With override - dotenv should win
	loader, err := fuda.New().
		WithDotEnvSearch(".env", []string{tmpDir}, fuda.DotEnvOverride()).
		Build()
	require.NoError(t, err)

	var cfg Config
	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "dotenv-value", cfg.Var)
}

// TestDotEnv_VariableExpansion verifies godotenv variable expansion.
func TestDotEnv_VariableExpansion(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	content := `EXPAND_BASE=hello
EXPAND_REF=${EXPAND_BASE}_world
`
	err := os.WriteFile(envPath, []byte(content), 0o600)
	require.NoError(t, err)

	type Config struct {
		Base string `env:"EXPAND_BASE"`
		Ref  string `env:"EXPAND_REF"`
	}

	loader, err := fuda.New().
		WithDotEnv(envPath).
		Build()
	require.NoError(t, err)

	var cfg Config
	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "hello", cfg.Base)
	assert.Equal(t, "hello_world", cfg.Ref)
}

// TestWithDotEnvSearch_NoSearchPaths verifies graceful handling of empty search paths.
func TestWithDotEnvSearch_NoSearchPaths(t *testing.T) {
	type Config struct {
		Host string `default:"fallback"`
	}

	loader, err := fuda.New().
		WithDotEnvSearch(".env", []string{}).
		Build()
	require.NoError(t, err)

	var cfg Config
	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "fallback", cfg.Host)
}

// TestWithDotEnvSearch_EmptyName verifies graceful handling of empty search name.
func TestWithDotEnvSearch_EmptyName(t *testing.T) {
	tmpDir := t.TempDir()

	type Config struct {
		Host string `default:"fallback"`
	}

	loader, err := fuda.New().
		WithDotEnvSearch("", []string{tmpDir}).
		Build()
	require.NoError(t, err)

	var cfg Config
	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "fallback", cfg.Host)
}
