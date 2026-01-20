package fuda

import (
	"encoding/json"
	"fmt"

	"github.com/arloliu/fuda/internal/types"
	"gopkg.in/yaml.v3"
)

// ByteSize represents a size in bytes with human-readable JSON/YAML serialization.
// It supports parsing both IEC (binary) and SI (decimal) units.
//
// Example:
//
//	type Config struct {
//	    MaxFileSize fuda.ByteSize `yaml:"max_file_size"`
//	}
//	// YAML: max_file_size: 10MiB
//	// JSON: {"max_file_size": "10MiB"}
type ByteSize int64

// Int64 returns the underlying int64 value (bytes).
func (b ByteSize) Int64() int64 {
	return int64(b)
}

// Int returns the value as int.
func (b ByteSize) Int() int {
	return int(b)
}

// Uint64 returns the value as uint64.
// Returns 0 for negative values.
func (b ByteSize) Uint64() uint64 {
	if b < 0 {
		return 0
	}

	return uint64(b) //nolint:gosec // G115: Overflow guarded by negative check above.
}

// String returns a human-readable representation using IEC units.
func (b ByteSize) String() string {
	bytes := int64(b)
	if bytes < 0 {
		return fmt.Sprintf("%d B", bytes)
	}

	const (
		kiB int64 = 1 << 10
		miB int64 = 1 << 20
		giB int64 = 1 << 30
		tiB int64 = 1 << 40
		piB int64 = 1 << 50
		eiB int64 = 1 << 60
	)

	switch {
	case bytes >= eiB:
		return fmt.Sprintf("%.2f EiB", float64(bytes)/float64(eiB))
	case bytes >= piB:
		return fmt.Sprintf("%.2f PiB", float64(bytes)/float64(piB))
	case bytes >= tiB:
		return fmt.Sprintf("%.2f TiB", float64(bytes)/float64(tiB))
	case bytes >= giB:
		return fmt.Sprintf("%.2f GiB", float64(bytes)/float64(giB))
	case bytes >= miB:
		return fmt.Sprintf("%.2f MiB", float64(bytes)/float64(miB))
	case bytes >= kiB:
		return fmt.Sprintf("%.2f KiB", float64(bytes)/float64(kiB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// MarshalJSON outputs size as quoted string.
func (b ByteSize) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.String())
}

// UnmarshalJSON parses size from string or number (bytes).
func (b *ByteSize) UnmarshalJSON(data []byte) error {
	// Try string first (preferred format)
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		parsed, err := types.ParseBytes(s)
		if err != nil {
			return fmt.Errorf("invalid byte size string %q: %w", s, err)
		}
		*b = ByteSize(parsed)

		return nil
	}

	// Fall back to number (bytes) for backwards compatibility
	var n int64
	if err := json.Unmarshal(data, &n); err != nil {
		return fmt.Errorf("byte size must be string or number, got: %s", string(data))
	}
	*b = ByteSize(n)

	return nil
}

// MarshalYAML outputs size as string.
func (b ByteSize) MarshalYAML() (any, error) {
	return b.String(), nil
}

// UnmarshalYAML parses size from string or number (bytes).
func (b *ByteSize) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.ScalarNode {
		return fmt.Errorf("expected scalar value for byte size, got %v", node.Kind)
	}

	// Try as size string first (preferred format)
	parsed, err := types.ParseBytes(node.Value)
	if err == nil {
		*b = ByteSize(parsed)

		return nil
	}

	// Fall back to number (bytes) for backwards compatibility
	var n int64
	if err := node.Decode(&n); err == nil {
		*b = ByteSize(n)

		return nil
	}

	return fmt.Errorf("invalid byte size value: %s", node.Value)
}
