package tests

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/arloliu/fuda"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Pointer to Nested Struct Types ---

// ConnectionConfig is a deeply nested struct
type ConnectionConfig struct {
	Host    string        `yaml:"host" default:"localhost"`
	Port    int           `yaml:"port" default:"5432"`
	Timeout time.Duration `yaml:"timeout" default:"30s"`
}

// DatabaseConfig contains pointer to nested structs
type DatabaseConfig struct {
	Primary  *ConnectionConfig   `yaml:"primary"`
	Replica  *ConnectionConfig   `yaml:"replica"`
	Replicas []*ConnectionConfig `yaml:"replicas"`
}

// CacheConfig for testing nested pointers
type CacheConfig struct {
	Redis *RedisConfig  `yaml:"redis"`
	TTL   time.Duration `yaml:"ttl" default:"1h"`
}

type RedisConfig struct {
	Host string `yaml:"host" default:"localhost"`
	Port int    `yaml:"port" default:"6379"`
}

// RootConfig is the top-level config with nested pointers
type RootConfig struct {
	Database *DatabaseConfig `yaml:"database"`
	Cache    *CacheConfig    `yaml:"cache"`
}

// --- Tests ---

func TestNestedPointer_NilFieldsGetDefaults(t *testing.T) {
	// When YAML has no value, nil pointer fields should remain nil
	// but if YAML initializes them, defaults should apply to nested fields

	cfg := &RootConfig{}
	loader, err := fuda.New().Build()
	require.NoError(t, err)

	err = loader.Load(cfg)
	require.NoError(t, err)

	// Top-level pointers should remain nil when no YAML source
	assert.Nil(t, cfg.Database, "Database should be nil without source")
	assert.Nil(t, cfg.Cache, "Cache should be nil without source")
}

func TestNestedPointer_EmptyKeyVsEmptyObject(t *testing.T) {
	// This test documents the semantic difference between:
	// - `key:` (empty value) -> pointer remains nil, no defaults
	// - `key: {}` (empty object) -> pointer allocated, defaults applied

	type Inner struct {
		Host string `yaml:"host" default:"localhost"`
		Port int    `yaml:"port" default:"8080"`
	}
	type Config struct {
		Server *Inner `yaml:"server"`
	}

	t.Run("empty key leaves pointer nil", func(t *testing.T) {
		yamlContent := `server:`
		cfg := &Config{}
		loader, err := fuda.New().FromBytes([]byte(yamlContent)).Build()
		require.NoError(t, err)

		err = loader.Load(cfg)
		require.NoError(t, err)

		assert.Nil(t, cfg.Server, "Empty key should leave pointer nil - no defaults applied")
	})

	t.Run("empty object allocates and applies defaults", func(t *testing.T) {
		yamlContent := `server: {}`
		cfg := &Config{}
		loader, err := fuda.New().FromBytes([]byte(yamlContent)).Build()
		require.NoError(t, err)

		err = loader.Load(cfg)
		require.NoError(t, err)

		require.NotNil(t, cfg.Server, "Empty object {} should allocate the struct")
		assert.Equal(t, "localhost", cfg.Server.Host, "Should get default host")
		assert.Equal(t, 8080, cfg.Server.Port, "Should get default port")
	})

	t.Run("partial object applies defaults to missing fields", func(t *testing.T) {
		yamlContent := `server:
  host: "custom.local"`
		cfg := &Config{}
		loader, err := fuda.New().FromBytes([]byte(yamlContent)).Build()
		require.NoError(t, err)

		err = loader.Load(cfg)
		require.NoError(t, err)

		require.NotNil(t, cfg.Server)
		assert.Equal(t, "custom.local", cfg.Server.Host, "Should get YAML value")
		assert.Equal(t, 8080, cfg.Server.Port, "Should get default for missing field")
	})

	t.Run("key omitted entirely leaves pointer nil", func(t *testing.T) {
		yamlContent := `other_key: value`
		cfg := &Config{}
		loader, err := fuda.New().FromBytes([]byte(yamlContent)).Build()
		require.NoError(t, err)

		err = loader.Load(cfg)
		require.NoError(t, err)

		assert.Nil(t, cfg.Server, "Omitted key should leave pointer nil")
	})
}

func TestNestedPointer_YAMLInitializesNestedDefaults(t *testing.T) {
	yamlContent := `
database:
  primary:
    host: "main.db.local"
`
	cfg := &RootConfig{}
	loader, err := fuda.New().FromBytes([]byte(yamlContent)).Build()
	require.NoError(t, err)

	err = loader.Load(cfg)
	require.NoError(t, err)

	// Database and Primary should be initialized by YAML
	require.NotNil(t, cfg.Database)
	require.NotNil(t, cfg.Database.Primary)

	// Host from YAML, Port from default
	assert.Equal(t, "main.db.local", cfg.Database.Primary.Host)
	assert.Equal(t, 5432, cfg.Database.Primary.Port, "Port should get default value")
	assert.Equal(t, 30*time.Second, cfg.Database.Primary.Timeout, "Timeout should get default value")

	// Replica should remain nil
	assert.Nil(t, cfg.Database.Replica)
}

func TestNestedPointer_ThreeLevelsDeep(t *testing.T) {
	// Test 3+ levels of nesting
	type Level3 struct {
		Value string `default:"level3_default"`
	}
	type Level2 struct {
		Inner *Level3 `yaml:"inner"`
	}
	type Level1 struct {
		Middle *Level2 `yaml:"middle"`
	}
	type Root struct {
		Outer *Level1 `yaml:"outer"`
	}

	yamlContent := `
outer:
  middle:
    inner:
      value: "custom"
`
	cfg := &Root{}
	loader, err := fuda.New().FromBytes([]byte(yamlContent)).Build()
	require.NoError(t, err)

	err = loader.Load(cfg)
	require.NoError(t, err)

	require.NotNil(t, cfg.Outer)
	require.NotNil(t, cfg.Outer.Middle)
	require.NotNil(t, cfg.Outer.Middle.Inner)
	assert.Equal(t, "custom", cfg.Outer.Middle.Inner.Value)
}

func TestNestedPointer_PartialYAMLWithDefaults(t *testing.T) {
	// YAML provides partial data, defaults fill the rest
	type Level3 struct {
		A string `default:"default_a"`
		B string `default:"default_b"`
		C string `default:"default_c"`
	}
	type Level2 struct {
		Inner *Level3 `yaml:"inner"`
	}
	type Root struct {
		Outer *Level2 `yaml:"outer"`
	}

	yamlContent := `
outer:
  inner:
    b: "yaml_b"
`
	cfg := &Root{}
	loader, err := fuda.New().FromBytes([]byte(yamlContent)).Build()
	require.NoError(t, err)

	err = loader.Load(cfg)
	require.NoError(t, err)

	require.NotNil(t, cfg.Outer)
	require.NotNil(t, cfg.Outer.Inner)
	assert.Equal(t, "default_a", cfg.Outer.Inner.A, "A should get default")
	assert.Equal(t, "yaml_b", cfg.Outer.Inner.B, "B should come from YAML")
	assert.Equal(t, "default_c", cfg.Outer.Inner.C, "C should get default")
}

func TestNestedPointer_MixedPointerAndNonPointer(t *testing.T) {
	type Inner struct {
		Value string `yaml:"value" default:"inner_default"`
	}
	type Config struct {
		PtrField    *Inner `yaml:"ptr_field"`
		NonPtrField Inner  `yaml:"non_ptr_field"`
	}

	// Test that pointer struct gets initialized and default is applied to non-pointer
	yamlContent := `
ptr_field:
  value: "ptr_value"
`
	cfg := &Config{}
	loader, err := fuda.New().FromBytes([]byte(yamlContent)).Build()
	require.NoError(t, err)

	err = loader.Load(cfg)
	require.NoError(t, err)

	require.NotNil(t, cfg.PtrField)
	assert.Equal(t, "ptr_value", cfg.PtrField.Value)
	// NonPtrField was not in YAML, so should get default
	assert.Equal(t, "inner_default", cfg.NonPtrField.Value, "Non-pointer should get default")
}

func TestNestedPointer_SliceOfPointers(t *testing.T) {
	type Server struct {
		Host string `yaml:"host" default:"localhost"`
		Port int    `yaml:"port" default:"8080"`
	}
	type Config struct {
		Servers []*Server `yaml:"servers"`
	}

	yamlContent := `
servers:
  - host: "server1.local"
  - port: 9090
  - host: "server3.local"
    port: 7070
`
	cfg := &Config{}
	loader, err := fuda.New().FromBytes([]byte(yamlContent)).Build()
	require.NoError(t, err)

	err = loader.Load(cfg)
	require.NoError(t, err)

	require.Len(t, cfg.Servers, 3)

	// First: host from YAML, port from default
	assert.Equal(t, "server1.local", cfg.Servers[0].Host)
	assert.Equal(t, 8080, cfg.Servers[0].Port, "Should get default port")

	// Second: host from default, port from YAML
	assert.Equal(t, "localhost", cfg.Servers[1].Host, "Should get default host")
	assert.Equal(t, 9090, cfg.Servers[1].Port)

	// Third: both from YAML
	assert.Equal(t, "server3.local", cfg.Servers[2].Host)
	assert.Equal(t, 7070, cfg.Servers[2].Port)
}

// --- Complex YAML Fixture Tests ---

func TestComplexYAML_FromFixture(t *testing.T) {
	// Feature settings map value type
	type FeatureSetting struct {
		Threshold  float64 `yaml:"threshold"`
		MaxRetries int     `yaml:"max_retries"`
		Enabled    bool    `yaml:"enabled"`
	}

	type FeaturesConfig struct {
		Enabled  []string                  `yaml:"enabled"`
		Settings map[string]FeatureSetting `yaml:"settings"`
	}

	type ReplicaConfig struct {
		Host string `yaml:"host" default:"localhost"`
		Port int    `yaml:"port" default:"5432"`
	}

	type PrimaryConfig struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	}

	type DBConfig struct {
		Primary  *PrimaryConfig  `yaml:"primary"`
		Replicas []ReplicaConfig `yaml:"replicas"`
	}

	type AppConfig struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
	}

	type ComplexConfig struct {
		App      AppConfig      `yaml:"app"`
		Database DBConfig       `yaml:"database"`
		Features FeaturesConfig `yaml:"features"`
	}

	fixturePath := filepath.Join("fixtures", "complex_config.yaml")
	cfg := &ComplexConfig{}

	err := fuda.LoadFile(fixturePath, cfg)
	require.NoError(t, err)

	// App
	assert.Equal(t, "myapp", cfg.App.Name)
	assert.Equal(t, "1.0.0", cfg.App.Version)

	// Database primary
	require.NotNil(t, cfg.Database.Primary)
	assert.Equal(t, "primary.db.local", cfg.Database.Primary.Host)
	assert.Equal(t, 5432, cfg.Database.Primary.Port)

	// Database replicas
	require.Len(t, cfg.Database.Replicas, 2)
	assert.Equal(t, "replica1.db.local", cfg.Database.Replicas[0].Host)
	assert.Equal(t, 5433, cfg.Database.Replicas[1].Port)

	// Features enabled list
	require.Len(t, cfg.Features.Enabled, 2)
	assert.Contains(t, cfg.Features.Enabled, "feature_a")
	assert.Contains(t, cfg.Features.Enabled, "feature_b")

	// Features settings map
	require.Len(t, cfg.Features.Settings, 2)
	assert.Equal(t, 0.8, cfg.Features.Settings["feature_a"].Threshold)
	assert.True(t, cfg.Features.Settings["feature_a"].Enabled)
	assert.Equal(t, 3, cfg.Features.Settings["feature_b"].MaxRetries)
	assert.False(t, cfg.Features.Settings["feature_b"].Enabled)
}

func TestComplexYAML_EmptyFile(t *testing.T) {
	type Config struct {
		Host string `default:"localhost"`
		Port int    `default:"8080"`
	}

	cfg := &Config{}
	loader, err := fuda.New().FromBytes([]byte("")).Build()
	require.NoError(t, err)

	err = loader.Load(cfg)
	require.NoError(t, err)

	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 8080, cfg.Port)
}

func TestComplexYAML_UnicodeValues(t *testing.T) {
	type Config struct {
		Name    string `yaml:"name"`
		Message string `yaml:"message"`
	}

	yamlContent := `
name: "日本語テスト"
message: "Hello 世界! 🌍"
`
	cfg := &Config{}
	loader, err := fuda.New().FromBytes([]byte(yamlContent)).Build()
	require.NoError(t, err)

	err = loader.Load(cfg)
	require.NoError(t, err)

	assert.Equal(t, "日本語テスト", cfg.Name)
	assert.Equal(t, "Hello 世界! 🌍", cfg.Message)
}

func TestComplexYAML_NestedMaps(t *testing.T) {
	type Config struct {
		Labels map[string]string         `yaml:"labels"`
		Nested map[string]map[string]int `yaml:"nested"`
	}

	yamlContent := `
labels:
  env: "production"
  team: "platform"
nested:
  group1:
    a: 1
    b: 2
  group2:
    c: 3
`
	cfg := &Config{}
	loader, err := fuda.New().FromBytes([]byte(yamlContent)).Build()
	require.NoError(t, err)

	err = loader.Load(cfg)
	require.NoError(t, err)

	assert.Equal(t, "production", cfg.Labels["env"])
	assert.Equal(t, "platform", cfg.Labels["team"])
	assert.Equal(t, 1, cfg.Nested["group1"]["a"])
	assert.Equal(t, 2, cfg.Nested["group1"]["b"])
	assert.Equal(t, 3, cfg.Nested["group2"]["c"])
}

func TestNestedPointer_WithRefFrom(t *testing.T) {
	// Setup password file in tests directory
	origDir, _ := os.Getwd()
	pwdFile := "test_nested_password"
	err := os.WriteFile(pwdFile, []byte("secret123"), 0o600)
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(pwdFile)
		_ = os.Chdir(origDir)
	}()

	type DBConfig struct {
		Host         string `yaml:"host" default:"localhost"`
		Password     string `refFrom:"PasswordFile"`
		PasswordFile string `yaml:"password_file"`
	}

	type Config struct {
		Database *DBConfig `yaml:"database"`
	}

	yamlContent := `
database:
  host: "db.local"
  password_file: "test_nested_password"
`
	cfg := &Config{}
	loader, err := fuda.New().FromBytes([]byte(yamlContent)).Build()
	require.NoError(t, err)

	err = loader.Load(cfg)
	require.NoError(t, err)

	require.NotNil(t, cfg.Database)
	assert.Equal(t, "db.local", cfg.Database.Host)
	assert.Equal(t, "secret123", cfg.Database.Password, "Password should be resolved from file")
}
