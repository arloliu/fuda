package tests

import (
	"testing"
	"time"

	"github.com/arloliu/fuda"
	"github.com/stretchr/testify/require"
)

// These tests pin down when the `default` tag applies relative to the
// config file: a default fills a field only when the document did not supply it.
// A key explicitly set to a zero value (0, false, "") is the operator's choice and must survive.
// An explicit null counts as absent (matching yaml.v3 decode, which ignores null),
// so the default applies.

func TestYAMLExplicitZeroSurvivesDefault(t *testing.T) {
	type Config struct {
		Workers int `yaml:"workers" default:"4"`
	}

	var cfg Config
	require.NoError(t, fuda.LoadBytes([]byte("workers: 0\n"), &cfg))
	require.Equal(t, 0, cfg.Workers)
}

func TestYAMLExplicitFalseSurvivesDefaultTrue(t *testing.T) {
	type Config struct {
		Enabled bool `yaml:"enabled" default:"true"`
	}

	var cfg Config
	require.NoError(t, fuda.LoadBytes([]byte("enabled: false\n"), &cfg))
	require.False(t, cfg.Enabled)
}

func TestYAMLExplicitEmptyStringSurvivesDefault(t *testing.T) {
	type Config struct {
		Name string `yaml:"name" default:"hello"`
	}

	var cfg Config
	require.NoError(t, fuda.LoadBytes([]byte("name: \"\"\n"), &cfg))
	require.Equal(t, "", cfg.Name)
}

func TestYAMLExplicitZeroDurationSurvivesDefault(t *testing.T) {
	type Config struct {
		RoundInterval time.Duration `yaml:"roundInterval" default:"30s"`
	}

	var cfg Config
	require.NoError(t, fuda.LoadBytes([]byte("roundInterval: 0s\n"), &cfg))
	require.Equal(t, time.Duration(0), cfg.RoundInterval)
}

func TestYAMLAbsentKeyGetsDefault(t *testing.T) {
	type Config struct {
		Workers int    `yaml:"workers" default:"4"`
		Name    string `yaml:"name" default:"hello"`
	}

	var cfg Config
	require.NoError(t, fuda.LoadBytes([]byte("workers: 2\n"), &cfg))
	require.Equal(t, 2, cfg.Workers)
	require.Equal(t, "hello", cfg.Name)
}

func TestYAMLExplicitNullGetsDefault(t *testing.T) {
	type Config struct {
		Workers int `yaml:"workers" default:"4"`
	}

	var cfg Config
	require.NoError(t, fuda.LoadBytes([]byte("workers: null\n"), &cfg))
	require.Equal(t, 4, cfg.Workers)
}

func TestYAMLEmptyValueGetsDefault(t *testing.T) {
	type Config struct {
		Workers int `yaml:"workers" default:"4"`
	}

	var cfg Config
	require.NoError(t, fuda.LoadBytes([]byte("workers:\n"), &cfg))
	require.Equal(t, 4, cfg.Workers)
}

func TestYAMLNestedExplicitZeroSurvivesDefault(t *testing.T) {
	type Sampling struct {
		RoundInterval time.Duration `yaml:"roundInterval" default:"30s"`
	}
	type Defaults struct {
		Sampling Sampling `yaml:"sampling"`
	}
	type Config struct {
		Defaults Defaults `yaml:"defaults"`
	}

	src := "defaults:\n  sampling:\n    roundInterval: 0s\n"
	var cfg Config
	require.NoError(t, fuda.LoadBytes([]byte(src), &cfg))
	require.Equal(t, time.Duration(0), cfg.Defaults.Sampling.RoundInterval)
}

func TestYAMLSuppliedFieldStillOverriddenByEnv(t *testing.T) {
	type Config struct {
		Workers int `yaml:"workers" default:"4" env:"TEST_YAML_DEFAULT_WORKERS"`
	}

	t.Setenv("TEST_YAML_DEFAULT_WORKERS", "0")

	var cfg Config
	require.NoError(t, fuda.LoadBytes([]byte("workers: 7\n"), &cfg))
	require.Equal(t, 0, cfg.Workers)
}

func TestYAMLSliceElementExplicitZeroSurvivesDefault(t *testing.T) {
	type Server struct {
		Port int `yaml:"port" default:"8080"`
	}
	type Config struct {
		Servers []Server `yaml:"servers"`
	}

	src := "servers:\n  - port: 0\n  - {}\n"
	var cfg Config
	require.NoError(t, fuda.LoadBytes([]byte(src), &cfg))
	require.Len(t, cfg.Servers, 2)
	require.Equal(t, 0, cfg.Servers[0].Port)
	require.Equal(t, 8080, cfg.Servers[1].Port)
}

func TestYAMLMapValueExplicitZeroSurvivesDefault(t *testing.T) {
	type Backend struct {
		Weight int `yaml:"weight" default:"10"`
	}
	type Config struct {
		Backends map[string]Backend `yaml:"backends"`
	}

	src := "backends:\n  a:\n    weight: 0\n  b: {}\n"
	var cfg Config
	require.NoError(t, fuda.LoadBytes([]byte(src), &cfg))
	require.Equal(t, 0, cfg.Backends["a"].Weight)
	require.Equal(t, 10, cfg.Backends["b"].Weight)
}

func TestYAMLMergeKeySuppliesField(t *testing.T) {
	type Section struct {
		Workers int `yaml:"workers" default:"4"`
		Depth   int `yaml:"depth" default:"9"`
	}
	type Config struct {
		Base Section `yaml:"base"`
		A    Section `yaml:"a"`
	}

	src := "base: &base\n  workers: 0\na:\n  <<: *base\n  depth: 1\n"
	var cfg Config
	require.NoError(t, fuda.LoadBytes([]byte(src), &cfg))
	// workers: 0 arrives in section a via the merge key and must survive.
	require.Equal(t, 0, cfg.A.Workers)
	require.Equal(t, 1, cfg.A.Depth)
	require.Equal(t, 0, cfg.Base.Workers)
	require.Equal(t, 9, cfg.Base.Depth)
}

func TestYAMLAliasValueSuppliesField(t *testing.T) {
	type Section struct {
		Workers int `yaml:"workers" default:"4"`
	}
	type Config struct {
		Base Section `yaml:"base"`
		A    Section `yaml:"a"`
	}

	src := "base: &base\n  workers: 0\na: *base\n"
	var cfg Config
	require.NoError(t, fuda.LoadBytes([]byte(src), &cfg))
	require.Equal(t, 0, cfg.A.Workers)
}

func TestYAMLOverridesExplicitZeroSurvivesDefault(t *testing.T) {
	type Config struct {
		Workers int `yaml:"workers" default:"4"`
	}

	loader, err := fuda.New().
		FromBytes([]byte("workers: 7\n")).
		WithOverrides(map[string]any{"workers": 0}).
		Build()
	require.NoError(t, err)

	var cfg Config
	require.NoError(t, loader.Load(&cfg))
	require.Equal(t, 0, cfg.Workers)
}

func TestYAMLEmbeddedStructExplicitZeroSurvivesDefault(t *testing.T) {
	type Base struct {
		Workers int `yaml:"workers" default:"4"`
	}
	type Config struct {
		Base `yaml:",inline"`
		Name string `yaml:"name" default:"hello"`
	}

	var cfg Config
	require.NoError(t, fuda.LoadBytes([]byte("workers: 0\n"), &cfg))
	require.Equal(t, 0, cfg.Workers)
	require.Equal(t, "hello", cfg.Name)
}

func TestYAMLKeyCaseMismatchGetsDefault(t *testing.T) {
	// yaml.v3 matches keys case-sensitively against the lowercased field
	// name, so "Workers" does not decode into the field; the default must
	// still apply, consistent with the decode.
	type Config struct {
		Workers int `yaml:"workers" default:"4"`
	}

	var cfg Config
	require.NoError(t, fuda.LoadBytes([]byte("Workers: 0\n"), &cfg))
	require.Equal(t, 4, cfg.Workers)
}

func TestYAMLExplicitEmptySurvivesMissingRefFallback(t *testing.T) {
	// A field with a ref that fails softly (missing file) and a default:
	// when the document explicitly supplies the empty string, that choice
	// wins over the default fallback.
	type Config struct {
		Value string `yaml:"value" ref:"file:///nonexistent.txt" default:"fallback_value"`
	}

	var cfg Config
	require.NoError(t, fuda.LoadBytes([]byte("value: \"\"\n"), &cfg))
	require.Equal(t, "", cfg.Value)
}
