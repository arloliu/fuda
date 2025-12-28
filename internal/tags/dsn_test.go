package tags_test

import (
	"context"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/arloliu/fuda/internal/tags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockResolver for testing ref function in DSN templates.
type dsnMockResolver struct {
	data map[string][]byte
}

func (m *dsnMockResolver) Resolve(_ context.Context, uri string) ([]byte, error) {
	if val, ok := m.data[uri]; ok {
		return val, nil
	}

	return nil, os.ErrNotExist
}

// Test structs
type DSNTestStruct struct {
	Host     string `default:"localhost"`
	Port     int    `default:"5432"`
	User     string `default:"admin"`
	Password string `default:"secret"`
	DBName   string `default:"mydb"`
	DSN      string `dsn:"postgres://${.User}:${.Password}@${.Host}:${.Port}/${.DBName}"`
}

type DSNNestedStruct struct {
	Database DatabaseConfig
	Cache    CacheConfig
	DBDSN    string `dsn:"postgres://${.Database.User}:${.Database.Password}@${.Database.Host}:5432/db"`
	RedisDSN string `dsn:"redis://${.Cache.Host}:${.Cache.Port}"`
}

type DatabaseConfig struct {
	Host     string
	User     string
	Password string
}

type CacheConfig struct {
	Host string
	Port int
}

type DSNWithRefStruct struct {
	Host string `default:"localhost"`
	DSN  string `dsn:"postgres://user:${ref:file://secret.txt}@${.Host}:5432/db"`
}

type DSNWithEnvStruct struct {
	Host string `default:"localhost"`
	DSN  string `dsn:"postgres://${env:TEST_DSN_USER}:${env:TEST_DSN_PASS}@${.Host}:5432/db"`
}

type DSNStrictStruct struct {
	Host string
	DSN  string `dsn:"postgres://user:pass@${.Host}:5432/db" dsnStrict:"true"`
}

type DSNNonStringStruct struct {
	Value int    `dsn:"invalid"`
	DSN   string `dsn:"test"`
}

func TestProcessDSN_BasicFieldReferences(t *testing.T) {
	s := DSNTestStruct{
		Host:     "db.example.com",
		Port:     5432,
		User:     "myuser",
		Password: "mypass",
		DBName:   "production",
	}
	v := reflect.ValueOf(&s).Elem()
	typ := v.Type()
	ctx := context.Background()

	field, _ := typ.FieldByName("DSN")
	val := v.FieldByName("DSN")

	err := tags.ProcessDSN(ctx, field, val, v, nil, "", nil)
	require.NoError(t, err)

	assert.Equal(t, "postgres://myuser:mypass@db.example.com:5432/production", s.DSN)
}

func TestProcessDSN_NestedFieldReferences(t *testing.T) {
	s := DSNNestedStruct{
		Database: DatabaseConfig{
			Host:     "db.example.com",
			User:     "admin",
			Password: "secret123",
		},
		Cache: CacheConfig{
			Host: "redis.example.com",
			Port: 6379,
		},
	}
	v := reflect.ValueOf(&s).Elem()
	typ := v.Type()
	ctx := context.Background()

	// Test database DSN
	field, _ := typ.FieldByName("DBDSN")
	val := v.FieldByName("DBDSN")
	err := tags.ProcessDSN(ctx, field, val, v, nil, "", nil)
	require.NoError(t, err)
	assert.Equal(t, "postgres://admin:secret123@db.example.com:5432/db", s.DBDSN)

	// Test redis DSN
	field, _ = typ.FieldByName("RedisDSN")
	val = v.FieldByName("RedisDSN")
	err = tags.ProcessDSN(ctx, field, val, v, nil, "", nil)
	require.NoError(t, err)
	assert.Equal(t, "redis://redis.example.com:6379", s.RedisDSN)
}

func TestProcessDSN_WithRefFunction(t *testing.T) {
	s := DSNWithRefStruct{
		Host: "db.example.com",
	}
	v := reflect.ValueOf(&s).Elem()
	typ := v.Type()
	ctx := context.Background()

	resolver := &dsnMockResolver{
		data: map[string][]byte{
			"file://secret.txt": []byte("supersecret"),
		},
	}

	field, _ := typ.FieldByName("DSN")
	val := v.FieldByName("DSN")

	err := tags.ProcessDSN(ctx, field, val, v, resolver, "", nil)
	require.NoError(t, err)

	assert.Equal(t, "postgres://user:supersecret@db.example.com:5432/db", s.DSN)
}

func TestProcessDSN_WithEnvFunction(t *testing.T) {
	t.Setenv("TEST_DSN_USER", "envuser")
	t.Setenv("TEST_DSN_PASS", "envpass")

	s := DSNWithEnvStruct{
		Host: "db.example.com",
	}
	v := reflect.ValueOf(&s).Elem()
	typ := v.Type()
	ctx := context.Background()

	field, _ := typ.FieldByName("DSN")
	val := v.FieldByName("DSN")

	err := tags.ProcessDSN(ctx, field, val, v, nil, "", nil)
	require.NoError(t, err)

	assert.Equal(t, "postgres://envuser:envpass@db.example.com:5432/db", s.DSN)
}

func TestProcessDSN_WithEnvPrefix(t *testing.T) {
	t.Setenv("APP_TEST_DSN_USER", "prefixeduser")
	t.Setenv("APP_TEST_DSN_PASS", "prefixedpass")

	s := DSNWithEnvStruct{
		Host: "db.example.com",
	}
	v := reflect.ValueOf(&s).Elem()
	typ := v.Type()
	ctx := context.Background()

	field, _ := typ.FieldByName("DSN")
	val := v.FieldByName("DSN")

	err := tags.ProcessDSN(ctx, field, val, v, nil, "APP_", nil)
	require.NoError(t, err)

	assert.Equal(t, "postgres://prefixeduser:prefixedpass@db.example.com:5432/db", s.DSN)
}

func TestProcessDSN_PermissiveMode(t *testing.T) {
	// Default (permissive) mode: empty fields result in empty strings
	s := DSNTestStruct{
		Host:     "db.example.com",
		Port:     5432,
		User:     "", // Empty user
		Password: "", // Empty password
		DBName:   "mydb",
	}
	v := reflect.ValueOf(&s).Elem()
	typ := v.Type()
	ctx := context.Background()

	field, _ := typ.FieldByName("DSN")
	val := v.FieldByName("DSN")

	err := tags.ProcessDSN(ctx, field, val, v, nil, "", nil)
	require.NoError(t, err)

	// Empty strings are allowed in permissive mode
	assert.Equal(t, "postgres://:@db.example.com:5432/mydb", s.DSN)
}

func TestProcessDSN_StrictMode_Error(t *testing.T) {
	// Strict mode with missing field should NOT error for zero values
	// (zero values are valid, only "<no value>" from undefined map keys triggers error)
	s := DSNStrictStruct{
		Host: "", // Empty host
	}
	v := reflect.ValueOf(&s).Elem()
	typ := v.Type()
	ctx := context.Background()

	field, _ := typ.FieldByName("DSN")
	val := v.FieldByName("DSN")

	err := tags.ProcessDSN(ctx, field, val, v, nil, "", nil)
	// Empty string is still valid, just produces empty value in output
	require.NoError(t, err)
	assert.Equal(t, "postgres://user:pass@:5432/db", s.DSN)
}

func TestProcessDSN_SkipNonZeroValue(t *testing.T) {
	s := DSNTestStruct{
		Host:     "db.example.com",
		Port:     5432,
		User:     "user",
		Password: "pass",
		DBName:   "db",
		DSN:      "existing://value", // Already set
	}
	v := reflect.ValueOf(&s).Elem()
	typ := v.Type()
	ctx := context.Background()

	field, _ := typ.FieldByName("DSN")
	val := v.FieldByName("DSN")

	err := tags.ProcessDSN(ctx, field, val, v, nil, "", nil)
	require.NoError(t, err)

	// Should not overwrite existing value
	assert.Equal(t, "existing://value", s.DSN)
}

func TestProcessDSN_NoDSNTag(t *testing.T) {
	type NoTagStruct struct {
		Host string
		DSN  string // No dsn tag
	}

	s := NoTagStruct{Host: "localhost"}
	v := reflect.ValueOf(&s).Elem()
	typ := v.Type()
	ctx := context.Background()

	field, _ := typ.FieldByName("DSN")
	val := v.FieldByName("DSN")

	err := tags.ProcessDSN(ctx, field, val, v, nil, "", nil)
	require.NoError(t, err)
	assert.Equal(t, "", s.DSN) // Unchanged
}

func TestProcessDSN_NonStringField_Error(t *testing.T) {
	s := DSNNonStringStruct{}
	v := reflect.ValueOf(&s).Elem()
	typ := v.Type()
	ctx := context.Background()

	field, _ := typ.FieldByName("Value")
	val := v.FieldByName("Value")

	err := tags.ProcessDSN(ctx, field, val, v, nil, "", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "dsn tag can only be used on string fields")
}

func TestProcessDSN_InvalidTemplate(t *testing.T) {
	type InvalidTemplateStruct struct {
		DSN string `dsn:"postgres://${.User"` // Missing closing brace
	}

	s := InvalidTemplateStruct{}
	v := reflect.ValueOf(&s).Elem()
	typ := v.Type()
	ctx := context.Background()

	field, _ := typ.FieldByName("DSN")
	val := v.FieldByName("DSN")

	err := tags.ProcessDSN(ctx, field, val, v, nil, "", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse template")
}

func TestProcessDSN_RefNoResolver(t *testing.T) {
	s := DSNWithRefStruct{
		Host: "localhost",
	}
	v := reflect.ValueOf(&s).Elem()
	typ := v.Type()
	ctx := context.Background()

	field, _ := typ.FieldByName("DSN")
	val := v.FieldByName("DSN")

	err := tags.ProcessDSN(ctx, field, val, v, nil, "", nil) // No resolver
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no resolver configured")
}

func TestProcessDSN_RefResolverError(t *testing.T) {
	s := DSNWithRefStruct{
		Host: "localhost",
	}
	v := reflect.ValueOf(&s).Elem()
	typ := v.Type()
	ctx := context.Background()

	resolver := &dsnMockResolver{
		data: map[string][]byte{}, // Empty, will cause ErrNotExist
	}

	field, _ := typ.FieldByName("DSN")
	val := v.FieldByName("DSN")

	err := tags.ProcessDSN(ctx, field, val, v, resolver, "", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resolve ref")
}

func TestProcessDSN_StrictMode_MissingKey(t *testing.T) {
	// Strict mode should error if a referenced field is missing from the map (unexported field scenario)
	// or if we simply reference a non-existent field.
	type StrictStruct struct {
		DSN string `dsn:"value=${.NonExistentField}" dsnStrict:"true"`
	}

	s := StrictStruct{}
	v := reflect.ValueOf(&s).Elem()
	typ := v.Type()
	ctx := context.Background()

	field, _ := typ.FieldByName("DSN")
	val := v.FieldByName("DSN")

	err := tags.ProcessDSN(ctx, field, val, v, nil, "", nil)
	require.Error(t, err)
	// Template execution error for missing field or map key
	errorMsg := err.Error()
	assert.True(t, strings.Contains(errorMsg, "can't evaluate field") || strings.Contains(errorMsg, "map has no entry for key") || strings.Contains(errorMsg, "dsn template references undefined field"),
		"expected error about missing field/key, got: %s", errorMsg)
}

func TestProcessDSN_QuoteEscaping(t *testing.T) {
	// Test that double quotes in the argument are properly escaped by preprocessor
	// We use env function to test this: ${env:VAR_"QUOTE"} -> ${env "VAR_\"QUOTE"}
	t.Setenv(`VAR_"QUOTE"`, "success")

	type QuoteStruct struct {
		DSN string `dsn:"${env:VAR_\"QUOTE\"}"`
	}

	s := QuoteStruct{}
	v := reflect.ValueOf(&s).Elem()
	typ := v.Type()
	ctx := context.Background()

	field, _ := typ.FieldByName("DSN")
	val := v.FieldByName("DSN")

	err := tags.ProcessDSN(ctx, field, val, v, nil, "", nil)
	require.NoError(t, err)
	assert.Equal(t, "success", s.DSN)
}
