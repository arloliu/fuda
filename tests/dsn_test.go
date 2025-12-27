package tests

import (
	"testing"

	"github.com/arloliu/fuda"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test structs for DSN integration tests
type DSNConfig struct {
	Host     string `yaml:"host" default:"localhost"`
	Port     int    `yaml:"port" default:"5432"`
	User     string `yaml:"user" env:"TEST_DSN_DB_USER"`
	Password string `yaml:"password" env:"TEST_DSN_DB_PASS"`
	DBName   string `yaml:"db_name" default:"mydb"`
	DSN      string `dsn:"postgres://${.User}:${.Password}@${.Host}:${.Port}/${.DBName}"`
}

type MultiDSNDatabase struct {
	Host     string `yaml:"host" default:"db.example.com"`
	User     string `yaml:"user" default:"dbuser"`
	Password string `yaml:"password" default:"dbpass"`
}

type MultiDSNRedis struct {
	Host string `yaml:"host" default:"redis.example.com"`
	Port int    `yaml:"port" default:"6379"`
}

type MultiDSNConfig struct {
	Database    MultiDSNDatabase `yaml:"database"`
	Redis       MultiDSNRedis    `yaml:"redis"`
	PostgresDSN string           `dsn:"postgres://${.Database.User}:${.Database.Password}@${.Database.Host}:5432/app"`
	RedisDSN    string           `dsn:"redis://${.Redis.Host}:${.Redis.Port}/0"`
}

type DSNWithStrictConfig struct {
	Host string `yaml:"host" default:"localhost"`
	User string `yaml:"user"` // No default, may be empty
	DSN  string `dsn:"postgres://${.User}@${.Host}:5432/db" dsnStrict:"true"`
}

func TestDSN_Integration_BasicComposition(t *testing.T) {
	yamlContent := `
host: db.example.com
port: 5432
user: myuser
password: mypass
db_name: production
`
	var cfg DSNConfig
	loader, err := fuda.New().
		FromBytes([]byte(yamlContent)).
		Build()
	require.NoError(t, err)

	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "postgres://myuser:mypass@db.example.com:5432/production", cfg.DSN)
}

func TestDSN_Integration_WithDefaults(t *testing.T) {
	// No YAML content, rely on defaults
	var cfg DSNConfig
	cfg.User = "defaultuser"
	cfg.Password = "defaultpass"

	loader, err := fuda.New().Build()
	require.NoError(t, err)

	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "postgres://defaultuser:defaultpass@localhost:5432/mydb", cfg.DSN)
}

func TestDSN_Integration_WithEnvOverrides(t *testing.T) {
	t.Setenv("TEST_DSN_DB_USER", "envuser")
	t.Setenv("TEST_DSN_DB_PASS", "envpass")

	yamlContent := `
host: db.example.com
user: yamluser
password: yamlpass
`
	var cfg DSNConfig
	loader, err := fuda.New().
		FromBytes([]byte(yamlContent)).
		Build()
	require.NoError(t, err)

	err = loader.Load(&cfg)
	require.NoError(t, err)

	// env should override yaml values
	assert.Equal(t, "postgres://envuser:envpass@db.example.com:5432/mydb", cfg.DSN)
}

func TestDSN_Integration_NestedStructFields(t *testing.T) {
	yamlContent := `
database:
  host: postgres.internal
  user: admin
  password: secret123
redis:
  host: redis.internal
  port: 6380
`
	var cfg MultiDSNConfig
	loader, err := fuda.New().
		FromBytes([]byte(yamlContent)).
		Build()
	require.NoError(t, err)

	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "postgres://admin:secret123@postgres.internal:5432/app", cfg.PostgresDSN)
	assert.Equal(t, "redis://redis.internal:6380/0", cfg.RedisDSN)
}

func TestDSN_Integration_NestedWithDefaults(t *testing.T) {
	// Use all defaults
	var cfg MultiDSNConfig
	loader, err := fuda.New().Build()
	require.NoError(t, err)

	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "postgres://dbuser:dbpass@db.example.com:5432/app", cfg.PostgresDSN)
	assert.Equal(t, "redis://redis.example.com:6379/0", cfg.RedisDSN)
}

func TestDSN_Integration_PermissiveMode(t *testing.T) {
	// Empty user/password should still work in permissive mode
	yamlContent := `
host: db.example.com
`
	var cfg DSNConfig
	loader, err := fuda.New().
		FromBytes([]byte(yamlContent)).
		Build()
	require.NoError(t, err)

	err = loader.Load(&cfg)
	require.NoError(t, err)

	// Empty user/password produces empty strings in the DSN
	assert.Equal(t, "postgres://:@db.example.com:5432/mydb", cfg.DSN)
}

func TestDSN_Integration_ExistingValueNotOverwritten(t *testing.T) {
	yamlContent := `
host: db.example.com
user: myuser
password: mypass
`
	cfg := DSNConfig{
		DSN: "custom://already-set", // Pre-set value
	}
	loader, err := fuda.New().
		FromBytes([]byte(yamlContent)).
		Build()
	require.NoError(t, err)

	err = loader.Load(&cfg)
	require.NoError(t, err)

	// DSN should not be overwritten since it was already set
	assert.Equal(t, "custom://already-set", cfg.DSN)
}

func TestDSN_Integration_WithTemplate(t *testing.T) {
	// Test that DSN works with template processing
	type TemplateData struct {
		Environment string
	}

	type ConfigWithTemplate struct {
		Host string `yaml:"host"`
		User string `yaml:"user"`
		Pass string `yaml:"password"`
		DSN  string `dsn:"postgres://${.User}:${.Pass}@${.Host}:5432/app"`
	}

	yamlContent := `
host: "{{ .Environment }}-db.example.com"
user: admin
password: secret
`
	var cfg ConfigWithTemplate
	loader, err := fuda.New().
		FromBytes([]byte(yamlContent)).
		WithTemplate(TemplateData{Environment: "prod"}).
		Build()
	require.NoError(t, err)

	err = loader.Load(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "postgres://admin:secret@prod-db.example.com:5432/app", cfg.DSN)
}
