package tests

import (
	"strings"
	"testing"

	"github.com/arloliu/fuda"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToKYAML(t *testing.T) {
	t.Run("preserves field ordering in nested YAML", func(t *testing.T) {
		// Create nested YAML with specific field ordering
		yamlContent := `
server:
  host: localhost
  port: 8080
  timeout: 30s
database:
  driver: postgres
  host: db.example.com
  port: 5432
  name: myapp
  credentials:
    username: admin
    password: secret
logging:
  level: info
  format: json
  output: stdout
`
		loader, err := fuda.New().FromBytes([]byte(yamlContent)).Build()
		require.NoError(t, err)

		kyamlBytes, err := loader.ToKYAML()
		require.NoError(t, err)

		kyamlStr := string(kyamlBytes)

		// Verify nested structure ordering is preserved
		// "server" should appear before "database"
		serverIdx := strings.Index(kyamlStr, "server")
		databaseIdx := strings.Index(kyamlStr, "database")
		loggingIdx := strings.Index(kyamlStr, "logging")

		assert.True(t, serverIdx < databaseIdx, "server should appear before database")
		assert.True(t, databaseIdx < loggingIdx, "database should appear before logging")

		// Within server section: host, port, timeout
		hostIdx := strings.Index(kyamlStr, "host")
		portIdx := strings.Index(kyamlStr, "port")
		timeoutIdx := strings.Index(kyamlStr, "timeout")

		assert.True(t, hostIdx < portIdx, "host should appear before port in server section")
		assert.True(t, portIdx < timeoutIdx, "port should appear before timeout in server section")

		// Within database.credentials: username, password
		usernameIdx := strings.Index(kyamlStr, "username")
		passwordIdx := strings.Index(kyamlStr, "password")

		assert.True(t, usernameIdx < passwordIdx, "username should appear before password")
	})

	t.Run("preserves deeply nested field ordering", func(t *testing.T) {
		yamlContent := `
level1:
  alpha: 1
  beta: 2
  gamma:
    first: a
    second: b
    third:
      inner1: x
      inner2: y
      inner3: z
  delta: 4
`
		loader, err := fuda.New().FromBytes([]byte(yamlContent)).Build()
		require.NoError(t, err)

		kyamlBytes, err := loader.ToKYAML()
		require.NoError(t, err)

		kyamlStr := string(kyamlBytes)

		// Verify top-level ordering within level1
		alphaIdx := strings.Index(kyamlStr, "alpha")
		betaIdx := strings.Index(kyamlStr, "beta")
		gammaIdx := strings.Index(kyamlStr, "gamma")
		deltaIdx := strings.Index(kyamlStr, "delta")

		assert.True(t, alphaIdx < betaIdx, "alpha should appear before beta")
		assert.True(t, betaIdx < gammaIdx, "beta should appear before gamma")
		assert.True(t, gammaIdx < deltaIdx, "gamma should appear before delta")

		// Verify nested ordering within gamma
		firstIdx := strings.Index(kyamlStr, "first")
		secondIdx := strings.Index(kyamlStr, "second")
		thirdIdx := strings.Index(kyamlStr, "third")

		assert.True(t, firstIdx < secondIdx, "first should appear before second")
		assert.True(t, secondIdx < thirdIdx, "second should appear before third")

		// Verify deeply nested ordering within third
		inner1Idx := strings.Index(kyamlStr, "inner1")
		inner2Idx := strings.Index(kyamlStr, "inner2")
		inner3Idx := strings.Index(kyamlStr, "inner3")

		assert.True(t, inner1Idx < inner2Idx, "inner1 should appear before inner2")
		assert.True(t, inner2Idx < inner3Idx, "inner2 should appear before inner3")
	})

	t.Run("returns error for empty source", func(t *testing.T) {
		loader, err := fuda.New().Build()
		require.NoError(t, err)

		_, err = loader.ToKYAML()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no source data")
	})

	t.Run("returns error for invalid YAML", func(t *testing.T) {
		invalidYAML := `
key: value
  invalid indentation
another: [unclosed bracket
`
		loader, err := fuda.New().FromBytes([]byte(invalidYAML)).Build()
		require.NoError(t, err)

		_, err = loader.ToKYAML()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not valid YAML")
	})

	t.Run("handles arrays with preserved ordering", func(t *testing.T) {
		yamlContent := `
items:
  - name: first
    value: 1
  - name: second
    value: 2
  - name: third
    value: 3
config:
  enabled: true
  options:
    - alpha
    - beta
    - gamma
`
		loader, err := fuda.New().FromBytes([]byte(yamlContent)).Build()
		require.NoError(t, err)

		kyamlBytes, err := loader.ToKYAML()
		require.NoError(t, err)

		kyamlStr := string(kyamlBytes)

		// Verify items appear before config
		itemsIdx := strings.Index(kyamlStr, "items")
		configIdx := strings.Index(kyamlStr, "config")
		assert.True(t, itemsIdx < configIdx, "items should appear before config")

		// Verify array element ordering (first, second, third)
		firstIdx := strings.Index(kyamlStr, "first")
		secondIdx := strings.Index(kyamlStr, "second")
		thirdIdx := strings.Index(kyamlStr, "third")

		assert.True(t, firstIdx < secondIdx, "first should appear before second")
		assert.True(t, secondIdx < thirdIdx, "second should appear before third")

		// Verify options array ordering
		alphaIdx := strings.Index(kyamlStr, "alpha")
		betaIdx := strings.Index(kyamlStr, "beta")
		gammaIdx := strings.Index(kyamlStr, "gamma")

		assert.True(t, alphaIdx < betaIdx, "alpha should appear before beta")
		assert.True(t, betaIdx < gammaIdx, "beta should appear before gamma")
	})
}
