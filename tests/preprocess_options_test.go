package tests

import (
	"testing"
	"time"

	"github.com/arloliu/fuda"
	"github.com/stretchr/testify/require"
)

func TestPreprocessOptions_Size(t *testing.T) {
	type Config struct {
		Size int64 `yaml:"size"`
	}

	yamlContent := "size: 1KB\n"

	t.Run("disabled", func(t *testing.T) {
		loader, err := fuda.New().FromBytes([]byte(yamlContent)).WithSizePreprocess(false).Build()
		require.NoError(t, err)

		var cfg Config
		err = loader.Load(&cfg)
		require.Error(t, err)
	})

	t.Run("enabled", func(t *testing.T) {
		loader, err := fuda.New().FromBytes([]byte(yamlContent)).WithSizePreprocess(true).Build()
		require.NoError(t, err)

		var cfg Config
		err = loader.Load(&cfg)
		require.NoError(t, err)
		require.EqualValues(t, 1000, cfg.Size)
	})
}

func TestPreprocessOptions_Duration(t *testing.T) {
	type Config struct {
		Timeout time.Duration `yaml:"timeout"`
	}

	yamlContent := "timeout: 2d\n"

	t.Run("disabled", func(t *testing.T) {
		loader, err := fuda.New().FromBytes([]byte(yamlContent)).WithDurationPreprocess(false).Build()
		require.NoError(t, err)

		var cfg Config
		err = loader.Load(&cfg)
		require.Error(t, err)
	})

	t.Run("enabled", func(t *testing.T) {
		loader, err := fuda.New().FromBytes([]byte(yamlContent)).WithDurationPreprocess(true).Build()
		require.NoError(t, err)

		var cfg Config
		err = loader.Load(&cfg)
		require.NoError(t, err)
		require.Equal(t, 48*time.Hour, cfg.Timeout)
	})
}
