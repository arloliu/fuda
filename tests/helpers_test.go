package tests

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/arloliu/fuda"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMustSetDefaults_Success verifies MustSetDefaults works for valid structs.
func TestMustSetDefaults_Success(t *testing.T) {
	type Config struct {
		Host string `default:"localhost"`
		Port int    `default:"8080"`
	}

	var cfg Config
	require.NotPanics(t, func() {
		fuda.MustSetDefaults(&cfg)
	})

	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 8080, cfg.Port)
}

// TestMustSetDefaults_Panic verifies MustSetDefaults panics on error.
func TestMustSetDefaults_Panic(t *testing.T) {
	require.Panics(t, func() {
		fuda.MustSetDefaults(nil)
	})
}

// TestMustLoadFile_Success verifies MustLoadFile works for valid files.
func TestMustLoadFile_Success(t *testing.T) {
	content := `host: "test-host"`
	tmpFile := "test_must_load.yaml"
	err := os.WriteFile(tmpFile, []byte(content), 0o600)
	require.NoError(t, err)
	defer os.Remove(tmpFile)

	type Config struct {
		Host string `yaml:"host"`
	}

	var cfg Config
	require.NotPanics(t, func() {
		fuda.MustLoadFile(tmpFile, &cfg)
	})

	assert.Equal(t, "test-host", cfg.Host)
}

// TestMustLoadFile_Panic verifies MustLoadFile panics on missing file.
func TestMustLoadFile_Panic(t *testing.T) {
	type Config struct{}
	var cfg Config

	require.Panics(t, func() {
		fuda.MustLoadFile("nonexistent_file.yaml", &cfg)
	})
}

// TestMustLoadBytes_Success verifies MustLoadBytes works.
func TestMustLoadBytes_Success(t *testing.T) {
	type Config struct {
		Value int `yaml:"value"`
	}

	var cfg Config
	require.NotPanics(t, func() {
		fuda.MustLoadBytes([]byte(`value: 42`), &cfg)
	})

	assert.Equal(t, 42, cfg.Value)
}

// TestMustLoadBytes_Panic verifies MustLoadBytes panics on nil target.
func TestMustLoadBytes_Panic(t *testing.T) {
	require.Panics(t, func() {
		fuda.MustLoadBytes([]byte(`value: 42`), nil)
	})
}

// TestMustLoadReader_Success verifies MustLoadReader works.
func TestMustLoadReader_Success(t *testing.T) {
	type Config struct {
		Name string `yaml:"name"`
	}

	var cfg Config
	require.NotPanics(t, func() {
		fuda.MustLoadReader(strings.NewReader(`name: test`), &cfg)
	})

	assert.Equal(t, "test", cfg.Name)
}

// TestMustLoadReader_Panic verifies MustLoadReader panics on nil target.
func TestMustLoadReader_Panic(t *testing.T) {
	require.Panics(t, func() {
		fuda.MustLoadReader(strings.NewReader(`name: test`), nil)
	})
}

// TestValidate_Valid verifies Validate passes for valid structs.
func TestValidate_Valid(t *testing.T) {
	type Config struct {
		Host string `validate:"required"`
		Port int    `validate:"min=1,max=65535"`
	}

	cfg := Config{Host: "localhost", Port: 8080}
	err := fuda.Validate(&cfg)
	require.NoError(t, err)
}

// TestValidate_Invalid verifies Validate fails for invalid structs.
func TestValidate_Invalid(t *testing.T) {
	type Config struct {
		Host string `validate:"required"`
		Port int    `validate:"min=1,max=65535"`
	}

	cfg := Config{Host: "", Port: 0} // invalid: empty host, port 0
	err := fuda.Validate(&cfg)
	require.Error(t, err)
}

// TestLoadEnv verifies LoadEnv reads environment variables.
func TestLoadEnv(t *testing.T) {
	type Config struct {
		Host string `env:"TEST_HOST"`
		Port int    `env:"TEST_PORT" default:"3000"`
	}

	require.NoError(t, os.Setenv("TEST_HOST", "env-host"))
	require.NoError(t, os.Setenv("TEST_PORT", "9000"))
	defer os.Unsetenv("TEST_HOST")
	defer os.Unsetenv("TEST_PORT")

	var cfg Config
	err := fuda.LoadEnv(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "env-host", cfg.Host)
	assert.Equal(t, 9000, cfg.Port)
}

// TestLoadEnvWithPrefix verifies LoadEnvWithPrefix uses prefix.
func TestLoadEnvWithPrefix(t *testing.T) {
	type Config struct {
		Host    string        `env:"HOST"`
		Timeout time.Duration `env:"TIMEOUT"`
	}

	require.NoError(t, os.Setenv("MYAPP_HOST", "prefixed-host"))
	require.NoError(t, os.Setenv("MYAPP_TIMEOUT", "5s"))
	defer os.Unsetenv("MYAPP_HOST")
	defer os.Unsetenv("MYAPP_TIMEOUT")

	var cfg Config
	err := fuda.LoadEnvWithPrefix("MYAPP_", &cfg)
	require.NoError(t, err)

	assert.Equal(t, "prefixed-host", cfg.Host)
	assert.Equal(t, 5*time.Second, cfg.Timeout)
}
