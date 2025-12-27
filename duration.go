package fuda

import (
	"encoding/json"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// Duration wraps time.Duration with human-readable JSON/YAML serialization.
// Unlike time.Duration which marshals to nanoseconds, Duration marshals to
// a string format (e.g., "1h30m", "5s").
//
// Example:
//
//	type Config struct {
//	    Timeout fuda.Duration `yaml:"timeout"`
//	}
//	// YAML: timeout: 5s
//	// JSON: {"timeout": "5s"}
type Duration time.Duration

// Duration returns the underlying time.Duration value.
func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

// String returns the duration string (e.g., "1h30m5s").
func (d Duration) String() string {
	return time.Duration(d).String()
}

// MarshalJSON outputs duration as quoted string.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON parses duration from string or number (nanoseconds).
func (d *Duration) UnmarshalJSON(data []byte) error {
	// Try string first (preferred format)
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		parsed, err := time.ParseDuration(s)
		if err != nil {
			return fmt.Errorf("invalid duration string %q: %w", s, err)
		}
		*d = Duration(parsed)

		return nil
	}

	// Fall back to number (nanoseconds) for backwards compatibility
	var n int64
	if err := json.Unmarshal(data, &n); err != nil {
		return fmt.Errorf("duration must be string or number, got: %s", string(data))
	}
	*d = Duration(n)

	return nil
}

// MarshalYAML outputs duration as string.
func (d Duration) MarshalYAML() (any, error) {
	return d.String(), nil
}

// UnmarshalYAML parses duration from string or number (nanoseconds).
func (d *Duration) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.ScalarNode {
		return fmt.Errorf("expected scalar value for duration, got %v", node.Kind)
	}

	// Try as duration string first (preferred format)
	parsed, err := time.ParseDuration(node.Value)
	if err == nil {
		*d = Duration(parsed)

		return nil
	}

	// Fall back to number (nanoseconds) for backwards compatibility
	var n int64
	if err := node.Decode(&n); err == nil {
		*d = Duration(n)

		return nil
	}

	return fmt.Errorf("invalid duration value: %s", node.Value)
}
