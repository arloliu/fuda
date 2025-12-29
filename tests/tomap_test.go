package tests

import (
	"testing"

	"github.com/arloliu/fuda"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToMap(t *testing.T) {
	t.Run("basic map extraction", func(t *testing.T) {
		yamlContent := `
host: localhost
port: 8080
`
		loader, err := fuda.New().FromBytes([]byte(yamlContent)).Build()
		require.NoError(t, err)

		result, err := loader.ToMap()
		require.NoError(t, err)

		assert.Equal(t, "localhost", result["host"])
		assert.Equal(t, 8080, result["port"])
	})

	t.Run("nested structures", func(t *testing.T) {
		yamlContent := `
database:
  host: db.example.com
  port: 5432
  credentials:
    username: admin
    password: secret
`
		loader, err := fuda.New().FromBytes([]byte(yamlContent)).Build()
		require.NoError(t, err)

		result, err := loader.ToMap()
		require.NoError(t, err)

		db, ok := result["database"].(map[string]any)
		require.True(t, ok, "database should be a map")
		assert.Equal(t, "db.example.com", db["host"])
		assert.Equal(t, 5432, db["port"])

		creds, ok := db["credentials"].(map[string]any)
		require.True(t, ok, "credentials should be a map")
		assert.Equal(t, "admin", creds["username"])
		assert.Equal(t, "secret", creds["password"])
	})

	t.Run("array handling", func(t *testing.T) {
		yamlContent := `
tags:
  - app
  - prod
  - v1
`
		loader, err := fuda.New().FromBytes([]byte(yamlContent)).Build()
		require.NoError(t, err)

		result, err := loader.ToMap()
		require.NoError(t, err)

		tags, ok := result["tags"].([]any)
		require.True(t, ok, "tags should be a slice")
		assert.Len(t, tags, 3)
		assert.Equal(t, "app", tags[0])
		assert.Equal(t, "prod", tags[1])
		assert.Equal(t, "v1", tags[2])
	})

	t.Run("mixed nested and arrays", func(t *testing.T) {
		yamlContent := `
servers:
  - name: web1
    port: 8080
  - name: web2
    port: 8081
config:
  enabled: true
`
		loader, err := fuda.New().FromBytes([]byte(yamlContent)).Build()
		require.NoError(t, err)

		result, err := loader.ToMap()
		require.NoError(t, err)

		servers, ok := result["servers"].([]any)
		require.True(t, ok, "servers should be a slice")
		require.Len(t, servers, 2)

		server1, ok := servers[0].(map[string]any)
		require.True(t, ok, "server element should be a map")
		assert.Equal(t, "web1", server1["name"])
		assert.Equal(t, 8080, server1["port"])

		config, ok := result["config"].(map[string]any)
		require.True(t, ok, "config should be a map")
		assert.Equal(t, true, config["enabled"])
	})

	t.Run("empty source error", func(t *testing.T) {
		loader, err := fuda.New().Build()
		require.NoError(t, err)

		_, err = loader.ToMap()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no source data")
	})

	t.Run("invalid YAML error", func(t *testing.T) {
		invalidYAML := `
key: value
  invalid indentation
another: [unclosed bracket
`
		loader, err := fuda.New().FromBytes([]byte(invalidYAML)).Build()
		require.NoError(t, err)

		_, err = loader.ToMap()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not valid YAML/JSON")
	})

	t.Run("JSON source works", func(t *testing.T) {
		jsonContent := `{"host": "localhost", "port": 8080, "debug": true}`

		loader, err := fuda.New().FromBytes([]byte(jsonContent)).Build()
		require.NoError(t, err)

		result, err := loader.ToMap()
		require.NoError(t, err)

		assert.Equal(t, "localhost", result["host"])
		assert.Equal(t, 8080, result["port"])
		assert.Equal(t, true, result["debug"])
	})

	t.Run("nil and empty values", func(t *testing.T) {
		yamlContent := `
explicit_null: ~
empty_value:
zero_int: 0
empty_string: ""
`
		loader, err := fuda.New().FromBytes([]byte(yamlContent)).Build()
		require.NoError(t, err)

		result, err := loader.ToMap()
		require.NoError(t, err)

		assert.Nil(t, result["explicit_null"])
		assert.Nil(t, result["empty_value"])
		assert.Equal(t, 0, result["zero_int"])
		assert.Equal(t, "", result["empty_string"])
	})

	t.Run("deeply nested structure", func(t *testing.T) {
		yamlContent := `
level1:
  level2:
    level3:
      level4:
        value: deep
`
		loader, err := fuda.New().FromBytes([]byte(yamlContent)).Build()
		require.NoError(t, err)

		result, err := loader.ToMap()
		require.NoError(t, err)

		l1, ok := result["level1"].(map[string]any)
		require.True(t, ok)
		l2, ok := l1["level2"].(map[string]any)
		require.True(t, ok)
		l3, ok := l2["level3"].(map[string]any)
		require.True(t, ok)
		l4, ok := l3["level4"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "deep", l4["value"])
	})

	t.Run("various types", func(t *testing.T) {
		yamlContent := `
string_val: hello
int_val: 42
float_val: 3.14
bool_val: true
`
		loader, err := fuda.New().FromBytes([]byte(yamlContent)).Build()
		require.NoError(t, err)

		result, err := loader.ToMap()
		require.NoError(t, err)

		assert.Equal(t, "hello", result["string_val"])
		assert.Equal(t, 42, result["int_val"])
		assert.Equal(t, 3.14, result["float_val"])
		assert.Equal(t, true, result["bool_val"])
	})
}
