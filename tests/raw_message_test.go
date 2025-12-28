package tests

import (
	"encoding/json"
	"testing"

	"github.com/arloliu/fuda"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestRawMessage_UnmarshalYAML(t *testing.T) {
	yamlData := `
properties:
  name: test
  count: 42
`
	type Config struct {
		Properties fuda.RawMessage `yaml:"properties"`
	}

	var cfg Config
	err := yaml.Unmarshal([]byte(yamlData), &cfg)
	require.NoError(t, err)
	assert.NotEmpty(t, cfg.Properties)

	// Verify we can unmarshal the raw message
	var props struct {
		Name  string `yaml:"name"`
		Count int    `yaml:"count"`
	}
	err = cfg.Properties.Unmarshal(&props)
	require.NoError(t, err)
	assert.Equal(t, "test", props.Name)
	assert.Equal(t, 42, props.Count)
}

func TestRawMessage_MarshalYAML(t *testing.T) {
	type Config struct {
		Properties fuda.RawMessage `yaml:"properties"`
	}

	// First unmarshal to get a properly formatted RawMessage
	originalYAML := `
properties:
  name: test
  count: 42
`
	var cfg Config
	err := yaml.Unmarshal([]byte(originalYAML), &cfg)
	require.NoError(t, err)

	// Marshal back to YAML
	out, err := yaml.Marshal(&cfg)
	require.NoError(t, err)

	// Unmarshal again and verify round-trip
	var result Config
	err = yaml.Unmarshal(out, &result)
	require.NoError(t, err)

	var props struct {
		Name  string `yaml:"name"`
		Count int    `yaml:"count"`
	}
	err = result.Properties.Unmarshal(&props)
	require.NoError(t, err)
	assert.Equal(t, "test", props.Name)
	assert.Equal(t, 42, props.Count)
}

func TestRawMessage_UnmarshalJSON(t *testing.T) {
	jsonData := `{"properties":{"name":"test","count":42}}`

	type Config struct {
		Properties fuda.RawMessage `json:"properties"`
	}

	var cfg Config
	err := json.Unmarshal([]byte(jsonData), &cfg)
	require.NoError(t, err)
	assert.NotEmpty(t, cfg.Properties)

	// Verify we can unmarshal the raw message
	var props struct {
		Name  string `yaml:"name" json:"name"`
		Count int    `yaml:"count" json:"count"`
	}
	err = cfg.Properties.Unmarshal(&props)
	require.NoError(t, err)
	assert.Equal(t, "test", props.Name)
	assert.Equal(t, 42, props.Count)
}

func TestRawMessage_MarshalJSON(t *testing.T) {
	type Config struct {
		Properties fuda.RawMessage `json:"properties"`
	}

	// Create a RawMessage with JSON content
	cfg := Config{
		Properties: fuda.RawMessage(`{"name":"test","count":42}`),
	}

	// Marshal to JSON
	out, err := json.Marshal(&cfg)
	require.NoError(t, err)
	assert.Equal(t, `{"properties":{"name":"test","count":42}}`, string(out))
}

func TestRawMessage_MarshalJSON_Nil(t *testing.T) {
	type Config struct {
		Properties fuda.RawMessage `json:"properties"`
	}

	cfg := Config{Properties: nil}
	out, err := json.Marshal(&cfg)
	require.NoError(t, err)
	assert.Equal(t, `{"properties":null}`, string(out))
}

func TestRawMessage_PolymorphicConfig(t *testing.T) {
	// Simulate dynamic device config scenario
	yamlData := `
devices:
  - type: car
    properties:
      wheels: 4
      engine: v8
  - type: bicycle
    properties:
      wheels: 2
      gears: 21
`
	type Device struct {
		Type       string          `yaml:"type"`
		Properties fuda.RawMessage `yaml:"properties"`
	}
	type Config struct {
		Devices []Device `yaml:"devices"`
	}

	type CarProperties struct {
		Wheels int    `yaml:"wheels"`
		Engine string `yaml:"engine"`
	}
	type BicycleProperties struct {
		Wheels int `yaml:"wheels"`
		Gears  int `yaml:"gears"`
	}

	var cfg Config
	err := yaml.Unmarshal([]byte(yamlData), &cfg)
	require.NoError(t, err)
	require.Len(t, cfg.Devices, 2)

	// Unmarshal car properties
	require.Equal(t, "car", cfg.Devices[0].Type)
	var carProps CarProperties
	err = cfg.Devices[0].Properties.Unmarshal(&carProps)
	require.NoError(t, err)
	assert.Equal(t, 4, carProps.Wheels)
	assert.Equal(t, "v8", carProps.Engine)

	// Unmarshal bicycle properties
	require.Equal(t, "bicycle", cfg.Devices[1].Type)
	var bicycleProps BicycleProperties
	err = cfg.Devices[1].Properties.Unmarshal(&bicycleProps)
	require.NoError(t, err)
	assert.Equal(t, 2, bicycleProps.Wheels)
	assert.Equal(t, 21, bicycleProps.Gears)
}

func TestRawMessage_WithFudaLoader(t *testing.T) {
	yamlData := `
type: car
properties:
  wheels: 4
  engine: v8
`
	type Config struct {
		Type       string          `yaml:"type"`
		Properties fuda.RawMessage `yaml:"properties"`
	}

	var cfg Config
	err := fuda.LoadBytes([]byte(yamlData), &cfg)
	require.NoError(t, err)
	assert.Equal(t, "car", cfg.Type)
	assert.NotEmpty(t, cfg.Properties)

	var props struct {
		Wheels int    `yaml:"wheels"`
		Engine string `yaml:"engine"`
	}
	err = cfg.Properties.Unmarshal(&props)
	require.NoError(t, err)
	assert.Equal(t, 4, props.Wheels)
	assert.Equal(t, "v8", props.Engine)
}

func TestRawMessage_Unmarshal_EmptyMessage(t *testing.T) {
	var m fuda.RawMessage
	var result map[string]any
	err := m.Unmarshal(&result)
	require.NoError(t, err)
	assert.Nil(t, result)
}
