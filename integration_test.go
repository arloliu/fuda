package fuda_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/arloliu/fuda"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Config struct {
	Host      string        `default:"localhost"`
	Port      int           `default:"8080"`
	Timeout   time.Duration `default:"10s"`
	Secret    string        `ref:"file://TEST_SECRET"`
	APIKey    string        `env:"API_KEY"`
	Remote    string        `refFrom:"RemoteRef"`
	RemoteRef string
	Database  DatabaseConfig `yaml:"database"`
}

type DatabaseConfig struct {
	User         string `default:"admin"`
	Password     string `refFrom:"PasswordFile"`
	PasswordFile string `default:"TEST_DB_PASSWORD"`
}

func TestLoad(t *testing.T) {
	// Setup test files

	err := os.WriteFile("TEST_SECRET", []byte("super_secret"), 0o600)
	require.NoError(t, err)
	defer os.Remove("TEST_SECRET")

	err = os.WriteFile("TEST_DB_PASSWORD", []byte("db_password"), 0o600)
	require.NoError(t, err)
	defer os.Remove("TEST_DB_PASSWORD")

	// Set Env
	require.NoError(t, os.Setenv("APP_API_KEY", "env_key"))
	defer os.Unsetenv("APP_API_KEY")

	// Setup Remote Server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("remote_content"))
	}))
	defer ts.Close()

	// Test Case 1: Defaults, Ref, RefFrom, Env, Remote
	var cfg Config
	cfg.RemoteRef = ts.URL // Set dynamic ref

	loader, err := fuda.New().
		WithEnvPrefix("APP_").
		WithTimeout(2 * time.Second).
		Build()
	require.NoError(t, err)

	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, 10*time.Second, cfg.Timeout)
	assert.Equal(t, "super_secret", cfg.Secret)
	assert.Equal(t, "env_key", cfg.APIKey)
	assert.Equal(t, "admin", cfg.Database.User)
	assert.Equal(t, "db_password", cfg.Database.Password)
	assert.Equal(t, "remote_content", cfg.Remote)

	// Test Case 2: Config Override
	configContent := `
host: "remote"
port: 9090
database:
  user: "root"
`
	configFile := "config.yaml"
	err = os.WriteFile(configFile, []byte(configContent), 0o600)
	require.NoError(t, err)
	defer os.Remove(configFile)

	var cfg2 Config
	err = fuda.LoadFile(configFile, &cfg2)
	require.NoError(t, err)

	assert.Equal(t, "remote", cfg2.Host)
	assert.Equal(t, 9090, cfg2.Port)
	assert.Equal(t, "root", cfg2.Database.User)
}
