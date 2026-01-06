package tests

import (
	"testing"

	"github.com/arloliu/fuda"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestByteSlice_RefTag(t *testing.T) {
	// Create in-memory filesystem
	memFs := afero.NewMemMapFs()

	// Create binary file in memory
	binaryContent := []byte("certificate-content-here")
	err := afero.WriteFile(memFs, "/certs/server.pem", binaryContent, 0o644)
	require.NoError(t, err)

	// Create config file
	configContent := []byte(`name: test-config`)
	err = afero.WriteFile(memFs, "/config.yaml", configContent, 0o644)
	require.NoError(t, err)

	type Config struct {
		Name string `yaml:"name"`
		Cert []byte `ref:"file:///certs/server.pem"`
	}

	loader, err := fuda.New().
		WithFilesystem(memFs).
		FromFile("/config.yaml").
		Build()
	require.NoError(t, err)

	var cfg Config
	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "test-config", cfg.Name)
	assert.Equal(t, binaryContent, cfg.Cert)
}

func TestByteSlice_RefFromTag(t *testing.T) {
	memFs := afero.NewMemMapFs()

	// Create key file
	keyContent := []byte("private-key-content")
	err := afero.WriteFile(memFs, "/secrets/api.key", keyContent, 0o644)
	require.NoError(t, err)

	// Create config file with keyPath field
	configContent := []byte(`
keyPath: "file:///secrets/api.key"
`)
	err = afero.WriteFile(memFs, "/config.yaml", configContent, 0o644)
	require.NoError(t, err)

	type Config struct {
		KeyPath string `yaml:"keyPath"`
		Key     []byte `refFrom:"KeyPath"`
	}

	loader, err := fuda.New().
		WithFilesystem(memFs).
		FromFile("/config.yaml").
		Build()
	require.NoError(t, err)

	var cfg Config
	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "file:///secrets/api.key", cfg.KeyPath)
	assert.Equal(t, keyContent, cfg.Key)
}

func TestByteSlice_RefWithTemplate(t *testing.T) {
	memFs := afero.NewMemMapFs()

	// Create key file
	keyContent := []byte("templated-key-content")
	err := afero.WriteFile(memFs, "/secrets/prod/api.key", keyContent, 0o644)
	require.NoError(t, err)

	// Create config file
	configContent := []byte(`
secretDir: "/secrets"
env: "prod"
`)
	err = afero.WriteFile(memFs, "/config.yaml", configContent, 0o644)
	require.NoError(t, err)

	type Config struct {
		SecretDir string `yaml:"secretDir"`
		Env       string `yaml:"env"`
		Key       []byte `ref:"file://${.SecretDir}/${.Env}/api.key"`
	}

	loader, err := fuda.New().
		WithFilesystem(memFs).
		FromFile("/config.yaml").
		Build()
	require.NoError(t, err)

	var cfg Config
	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, keyContent, cfg.Key)
}

func TestByteSlice_DefaultTag(t *testing.T) {
	memFs := afero.NewMemMapFs()

	configContent := []byte(`name: test`)
	err := afero.WriteFile(memFs, "/config.yaml", configContent, 0o644)
	require.NoError(t, err)

	type Config struct {
		Name     string `yaml:"name"`
		Fallback []byte `default:"default-fallback-value"`
	}

	loader, err := fuda.New().
		WithFilesystem(memFs).
		FromFile("/config.yaml").
		Build()
	require.NoError(t, err)

	var cfg Config
	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "test", cfg.Name)
	assert.Equal(t, []byte("default-fallback-value"), cfg.Fallback)
}

func TestByteSlice_EnvTag(t *testing.T) {
	memFs := afero.NewMemMapFs()

	t.Setenv("TEST_API_SECRET", "secret-from-env")

	configContent := []byte(`name: env-test`)
	err := afero.WriteFile(memFs, "/config.yaml", configContent, 0o644)
	require.NoError(t, err)

	type Config struct {
		Name   string `yaml:"name"`
		Secret []byte `env:"TEST_API_SECRET"`
	}

	loader, err := fuda.New().
		WithFilesystem(memFs).
		FromFile("/config.yaml").
		Build()
	require.NoError(t, err)

	var cfg Config
	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "env-test", cfg.Name)
	assert.Equal(t, []byte("secret-from-env"), cfg.Secret)
}

func TestByteSlice_BinaryContent(t *testing.T) {
	memFs := afero.NewMemMapFs()

	// Create binary file with non-UTF8 content
	binaryContent := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0x80, 0x7F, 0x89, 0x50, 0x4E, 0x47}
	err := afero.WriteFile(memFs, "/data/binary.bin", binaryContent, 0o644)
	require.NoError(t, err)

	configContent := []byte(`name: binary-test`)
	err = afero.WriteFile(memFs, "/config.yaml", configContent, 0o644)
	require.NoError(t, err)

	type Config struct {
		Name string `yaml:"name"`
		Data []byte `ref:"file:///data/binary.bin"`
	}

	loader, err := fuda.New().
		WithFilesystem(memFs).
		FromFile("/config.yaml").
		Build()
	require.NoError(t, err)

	var cfg Config
	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, binaryContent, cfg.Data, "Binary content should be preserved exactly")
}
