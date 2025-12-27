package fuda_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/arloliu/fuda"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestDuration_String(t *testing.T) {
	d := fuda.Duration(5 * time.Second)
	assert.Equal(t, "5s", d.String())

	d = fuda.Duration(1*time.Hour + 30*time.Minute)
	assert.Equal(t, "1h30m0s", d.String())
}

func TestDuration_Duration(t *testing.T) {
	d := fuda.Duration(5 * time.Second)
	assert.Equal(t, 5*time.Second, d.Duration())
}

func TestDuration_MarshalJSON(t *testing.T) {
	type Config struct {
		Timeout fuda.Duration `json:"timeout"`
	}

	cfg := Config{Timeout: fuda.Duration(5 * time.Second)}
	out, err := json.Marshal(&cfg)
	require.NoError(t, err)
	assert.Equal(t, `{"timeout":"5s"}`, string(out))
}

func TestDuration_UnmarshalJSON_String(t *testing.T) {
	type Config struct {
		Timeout fuda.Duration `json:"timeout"`
	}

	var cfg Config
	err := json.Unmarshal([]byte(`{"timeout":"1h30m"}`), &cfg)
	require.NoError(t, err)
	assert.Equal(t, 1*time.Hour+30*time.Minute, cfg.Timeout.Duration())
}

func TestDuration_UnmarshalJSON_Number(t *testing.T) {
	// Backwards compatibility: accept nanoseconds as number
	type Config struct {
		Timeout fuda.Duration `json:"timeout"`
	}

	var cfg Config
	err := json.Unmarshal([]byte(`{"timeout":5000000000}`), &cfg)
	require.NoError(t, err)
	assert.Equal(t, 5*time.Second, cfg.Timeout.Duration())
}

func TestDuration_MarshalYAML(t *testing.T) {
	type Config struct {
		Timeout fuda.Duration `yaml:"timeout"`
	}

	cfg := Config{Timeout: fuda.Duration(5 * time.Second)}
	out, err := yaml.Marshal(&cfg)
	require.NoError(t, err)
	assert.Equal(t, "timeout: 5s\n", string(out))
}

func TestDuration_UnmarshalYAML_String(t *testing.T) {
	type Config struct {
		Timeout fuda.Duration `yaml:"timeout"`
	}

	var cfg Config
	err := yaml.Unmarshal([]byte("timeout: 1h30m"), &cfg)
	require.NoError(t, err)
	assert.Equal(t, 1*time.Hour+30*time.Minute, cfg.Timeout.Duration())
}

func TestDuration_UnmarshalYAML_Number(t *testing.T) {
	// Backwards compatibility: accept nanoseconds as number
	type Config struct {
		Timeout fuda.Duration `yaml:"timeout"`
	}

	var cfg Config
	err := yaml.Unmarshal([]byte("timeout: 5000000000"), &cfg)
	require.NoError(t, err)
	assert.Equal(t, 5*time.Second, cfg.Timeout.Duration())
}

func TestDuration_WithFudaLoader(t *testing.T) {
	yamlData := `
timeout: 30s
retryInterval: 5m
`
	type Config struct {
		Timeout       fuda.Duration `yaml:"timeout"`
		RetryInterval fuda.Duration `yaml:"retryInterval"`
	}

	var cfg Config
	err := fuda.LoadBytes([]byte(yamlData), &cfg)
	require.NoError(t, err)
	assert.Equal(t, 30*time.Second, cfg.Timeout.Duration())
	assert.Equal(t, 5*time.Minute, cfg.RetryInterval.Duration())
}

func TestDuration_UnmarshalJSON_InvalidString(t *testing.T) {
	type Config struct {
		Timeout fuda.Duration `json:"timeout"`
	}

	var cfg Config
	err := json.Unmarshal([]byte(`{"timeout":"invalid"}`), &cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid duration")
}

func TestDuration_UnmarshalYAML_InvalidValue(t *testing.T) {
	type Config struct {
		Timeout fuda.Duration `yaml:"timeout"`
	}

	var cfg Config
	err := yaml.Unmarshal([]byte("timeout: [1, 2, 3]"), &cfg)
	require.Error(t, err)
}
