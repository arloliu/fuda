package tests

import (
	"os"
	"testing"

	"github.com/arloliu/fuda"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NullIntScanner implements sql.Scanner for testing custom type fallback
type NullIntScanner struct {
	Valid bool
	Value int
}

func (n *NullIntScanner) Scan(value any) error {
	if value == nil {
		n.Valid = false
		return nil
	}
	// "Empty" string logic for scanner
	if s, ok := value.(string); ok {
		if s == "" {
			n.Valid = false // Empty string -> null/invalid
			return nil
		}
	}

	// Delegate to sql.NullInt64 logic (simplified)
	// In real world, we'd iterate types. Here we just assume string input from env/ref
	// But let's reuse fuda's converter if possible? No, Scanner takes raw value.
	// fuda passes string from ProcessRef.
	return nil
}

func TestTypeSupport_Fallback(t *testing.T) {
	// Setup
	t.Setenv("TEST_INT_VAL", "42")
	t.Setenv("TEST_BOOL_VAL", "true")
	t.Setenv("TEST_FLOAT_VAL", "3.14")

	t.Run("Basic Types Fallback", func(t *testing.T) {
		type Config struct {
			IntVal   int     `ref:"env://MISSING_INT" default:"100"`
			BoolVal  bool    `ref:"env://MISSING_BOOL" default:"true"`
			FloatVal float64 `ref:"env://MISSING_FLOAT" default:"99.9"`
		}

		var cfg Config
		err := fuda.SetDefaults(&cfg)
		require.NoError(t, err)

		assert.Equal(t, 100, cfg.IntVal)
		assert.Equal(t, true, cfg.BoolVal)
		assert.Equal(t, 99.9, cfg.FloatVal)
	})

	t.Run("Basic Types Resolution", func(t *testing.T) {
		type Config struct {
			IntVal   int     `ref:"env://TEST_INT_VAL"`
			BoolVal  bool    `ref:"env://TEST_BOOL_VAL"`
			FloatVal float64 `ref:"env://TEST_FLOAT_VAL"`
		}

		var cfg Config
		err := fuda.SetDefaults(&cfg)
		require.NoError(t, err)

		assert.Equal(t, 42, cfg.IntVal)
		assert.Equal(t, true, cfg.BoolVal)
		assert.Equal(t, 3.14, cfg.FloatVal)
	})

	t.Run("Int Empty Error", func(t *testing.T) {
		type Config struct {
			IntVal int `ref:"env://EMPTY_INT"`
		}

		os.Setenv("EMPTY_INT", "")
		defer os.Unsetenv("EMPTY_INT")

		var cfg Config
		err := fuda.SetDefaults(&cfg)
		// Should error because "" is not a valid int
		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty string")
	})
}

// CustomScanner implements sql.Scanner
type CustomScanner struct {
	Val string
}

func (cs *CustomScanner) Scan(src any) error {
	if src == nil {
		return nil
	}
	switch s := src.(type) {
	case string:
		cs.Val = "scanned:" + s
	case []byte:
		cs.Val = "scanned:" + string(s)
	}

	return nil
}

func TestCustomType_Fallback(t *testing.T) {
	t.Run("Scanner Fallback", func(t *testing.T) {
		type Config struct {
			Custom CustomScanner `ref:"env://MISSING_CUSTOM" default:"default-val"`
		}

		var cfg Config
		// Note: default tag support for Scanner depends on fuda implementation.
		// If fails, verify basic fallback first.
		err := fuda.SetDefaults(&cfg)
		require.NoError(t, err)

		// If default works for scanner
		assert.Equal(t, "scanned:default-val", cfg.Custom.Val)
	})

	t.Run("Scanner Empty", func(t *testing.T) {
		type Config struct {
			Custom CustomScanner `ref:"env://EMPTY_CUSTOM" default:"default-val"`
		}

		os.Setenv("EMPTY_CUSTOM", "")
		defer os.Unsetenv("EMPTY_CUSTOM")

		var cfg Config
		err := fuda.SetDefaults(&cfg)
		require.NoError(t, err)

		// Empty string passed to Scan
		assert.Equal(t, "scanned:", cfg.Custom.Val)
	})
}
