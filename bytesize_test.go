package fuda_test

import (
	"encoding/json"
	"testing"

	"github.com/arloliu/fuda"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestByteSize_Parsing(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"1024", 1024},
		{"1KiB", 1024},
		{"1kib", 1024},
		{"10MiB", 10 * 1024 * 1024},
		{"2GiB", 2 * 1024 * 1024 * 1024},
		{"1KB", 1000},
		{"1MB", 1000 * 1000},
		{"1GB", 1000 * 1000 * 1000},
		{"0.5MiB", 512 * 1024},
		{"100B", 100},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			yamlInput := "size: " + tc.input
			var cfg struct {
				Size fuda.ByteSize `yaml:"size"`
			}
			err := yaml.Unmarshal([]byte(yamlInput), &cfg)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, cfg.Size.Int64())
		})
	}
}

func TestByteSize_Methods(t *testing.T) {
	b := fuda.ByteSize(1024 * 1024) // 1 MiB
	assert.Equal(t, int64(1048576), b.Int64())
	assert.Equal(t, 1048576, b.Int())
	assert.Equal(t, uint64(1048576), b.Uint64())
	assert.Equal(t, "1.00 MiB", b.String())
}

func TestByteSize_JSON(t *testing.T) {
	t.Run("unmarshal string", func(t *testing.T) {
		var cfg struct {
			Size fuda.ByteSize `json:"size"`
		}
		err := json.Unmarshal([]byte(`{"size":"10MiB"}`), &cfg)
		require.NoError(t, err)
		assert.Equal(t, int64(10*1024*1024), cfg.Size.Int64())
	})

	t.Run("unmarshal number", func(t *testing.T) {
		var cfg struct {
			Size fuda.ByteSize `json:"size"`
		}
		err := json.Unmarshal([]byte(`{"size":1024}`), &cfg)
		require.NoError(t, err)
		assert.Equal(t, int64(1024), cfg.Size.Int64())
	})

	t.Run("marshal", func(t *testing.T) {
		cfg := struct {
			Size fuda.ByteSize `json:"size"`
		}{Size: fuda.ByteSize(1024 * 1024)}
		data, err := json.Marshal(cfg)
		require.NoError(t, err)
		assert.Equal(t, `{"size":"1.00 MiB"}`, string(data))
	})
}

func TestByteSize_YAML(t *testing.T) {
	t.Run("unmarshal string", func(t *testing.T) {
		var cfg struct {
			Size fuda.ByteSize `yaml:"size"`
		}
		err := yaml.Unmarshal([]byte("size: 10MiB"), &cfg)
		require.NoError(t, err)
		assert.Equal(t, int64(10*1024*1024), cfg.Size.Int64())
	})

	t.Run("unmarshal number", func(t *testing.T) {
		var cfg struct {
			Size fuda.ByteSize `yaml:"size"`
		}
		err := yaml.Unmarshal([]byte("size: 1024"), &cfg)
		require.NoError(t, err)
		assert.Equal(t, int64(1024), cfg.Size.Int64())
	})

	t.Run("marshal", func(t *testing.T) {
		cfg := struct {
			Size fuda.ByteSize `yaml:"size"`
		}{Size: fuda.ByteSize(1024 * 1024)}
		data, err := yaml.Marshal(cfg)
		require.NoError(t, err)
		assert.Equal(t, "size: 1.00 MiB\n", string(data))
	})
}

func TestByteSize_Errors(t *testing.T) {
	t.Run("fractional bytes", func(t *testing.T) {
		var cfg struct {
			Size fuda.ByteSize `yaml:"size"`
		}
		err := yaml.Unmarshal([]byte("size: 0.1B"), &cfg)
		require.Error(t, err)
	})

	t.Run("unknown unit", func(t *testing.T) {
		var cfg struct {
			Size fuda.ByteSize `yaml:"size"`
		}
		err := yaml.Unmarshal([]byte("size: 10XB"), &cfg)
		require.Error(t, err)
	})
}
