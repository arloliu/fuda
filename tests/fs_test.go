package tests

import (
	"testing"

	"github.com/arloliu/fuda"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithFilesystem_MemoryFS(t *testing.T) {
	// Create in-memory filesystem
	memFs := afero.NewMemMapFs()

	// Create config file in memory
	configContent := []byte(`
host: localhost
port: 8080
`)
	err := afero.WriteFile(memFs, "/config.yaml", configContent, 0o644)
	require.NoError(t, err)

	type Config struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	}

	// Load config using memory filesystem
	loader, err := fuda.New().
		WithFilesystem(memFs).
		FromFile("/config.yaml").
		Build()
	require.NoError(t, err)

	var cfg Config
	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 8080, cfg.Port)
}

func TestWithFilesystem_FileRefFromMemoryFS(t *testing.T) {
	// Create in-memory filesystem
	memFs := afero.NewMemMapFs()

	// Create config and secret files in memory
	secretContent := []byte("my-secret-password")
	configContent := []byte(`
database:
  host: db.example.com
  password: ""
`)
	err := afero.WriteFile(memFs, "/secrets/db-password.txt", secretContent, 0o644)
	require.NoError(t, err)
	err = afero.WriteFile(memFs, "/config.yaml", configContent, 0o644)
	require.NoError(t, err)

	type Database struct {
		Host     string `yaml:"host"`
		Password string `yaml:"password" ref:"file:///secrets/db-password.txt"`
	}
	type Config struct {
		Database Database `yaml:"database"`
	}

	// Load config using memory filesystem
	loader, err := fuda.New().
		WithFilesystem(memFs).
		FromFile("/config.yaml").
		Build()
	require.NoError(t, err)

	var cfg Config
	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "db.example.com", cfg.Database.Host)
	assert.Equal(t, "my-secret-password", cfg.Database.Password)
}

func TestSetDefaultFs(t *testing.T) {
	// Store original default
	originalFs := fuda.DefaultFs
	defer func() { fuda.DefaultFs = originalFs }()

	// Create in-memory filesystem
	memFs := afero.NewMemMapFs()
	err := afero.WriteFile(memFs, "/test.yaml", []byte("value: test"), 0o644)
	require.NoError(t, err)

	// Set as global default
	fuda.SetDefaultFs(memFs)

	type Config struct {
		Value string `yaml:"value"`
	}

	// Load without explicit WithFilesystem - should use global default
	loader, err := fuda.New().
		FromFile("/test.yaml").
		Build()
	require.NoError(t, err)

	var cfg Config
	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "test", cfg.Value)
}

func TestResetDefaultFs(t *testing.T) {
	// Store original default
	originalFs := fuda.DefaultFs

	// Set custom filesystem
	fuda.SetDefaultFs(afero.NewMemMapFs())
	assert.NotEqual(t, originalFs, fuda.DefaultFs)

	// Reset to OS filesystem
	fuda.ResetDefaultFs()

	// DefaultFs should be a new OsFs (can't compare directly, but type should match)
	_, isOsFs := fuda.DefaultFs.(*afero.OsFs)
	assert.True(t, isOsFs, "DefaultFs should be OsFs after reset")
}

func TestWithFilesystem_FileNotFound(t *testing.T) {
	memFs := afero.NewMemMapFs()

	// Try to load non-existent file
	_, err := fuda.New().
		WithFilesystem(memFs).
		FromFile("/nonexistent.yaml").
		Build()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file does not exist")
}

func TestWithFilesystem_RefFileNotFound(t *testing.T) {
	memFs := afero.NewMemMapFs()

	configContent := []byte(`value: ""`)
	err := afero.WriteFile(memFs, "/config.yaml", configContent, 0o644)
	require.NoError(t, err)

	type Config struct {
		Value string `yaml:"value" ref:"file:///nonexistent.txt"`
	}

	loader, err := fuda.New().
		WithFilesystem(memFs).
		FromFile("/config.yaml").
		Build()
	require.NoError(t, err)

	var cfg Config
	err = loader.Load(&cfg)
	assert.Error(t, err)
}

func TestWithFilesystem_OverridesGlobalDefault(t *testing.T) {
	// Store original default
	originalFs := fuda.DefaultFs
	defer func() { fuda.DefaultFs = originalFs }()

	// Set global default with one file
	globalFs := afero.NewMemMapFs()
	err := afero.WriteFile(globalFs, "/config.yaml", []byte("value: global"), 0o644)
	require.NoError(t, err)
	fuda.SetDefaultFs(globalFs)

	// Create instance-specific fs with different content
	instanceFs := afero.NewMemMapFs()
	err = afero.WriteFile(instanceFs, "/config.yaml", []byte("value: instance"), 0o644)
	require.NoError(t, err)

	type Config struct {
		Value string `yaml:"value"`
	}

	// Use WithFilesystem - should override global default
	loader, err := fuda.New().
		WithFilesystem(instanceFs).
		FromFile("/config.yaml").
		Build()
	require.NoError(t, err)

	var cfg Config
	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "instance", cfg.Value, "Instance filesystem should override global default")
}
