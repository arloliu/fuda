package fuda

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
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
		parsed, err := parseDuration(s)
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
	parsed, err := parseDuration(node.Value)
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

// parseDuration extends time.ParseDuration to support days with 'd' suffix.
// Examples: "5d" -> 5 days, "1d12h" -> 1 day and 12 hours, "2d30m" -> 2 days and 30 minutes.
func parseDuration(s string) (time.Duration, error) {
	// Find and convert 'd' suffix for days to hours
	// We need to handle cases like "5d", "1d12h", "2d30m5s"
	result := strings.Builder{}
	i := 0
	for i < len(s) {
		// Find the start of a number
		numStart := i
		for i < len(s) && (s[i] == '-' || s[i] == '+' || s[i] == '.' || (s[i] >= '0' && s[i] <= '9')) {
			i++
		}
		if i == numStart {
			// No number found, just copy the character
			if i < len(s) {
				result.WriteByte(s[i])
				i++
			}
			continue
		}
		numStr := s[numStart:i]

		// Find the unit
		unitStart := i
		for i < len(s) && ((s[i] >= 'a' && s[i] <= 'z') || (s[i] >= 'A' && s[i] <= 'Z')) {
			i++
		}
		unit := s[unitStart:i]

		// Convert 'd' or 'D' to hours
		if unit == "d" || unit == "D" {
			// Parse the number and multiply by 24
			days, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid duration: %s", s)
			}
			hours := days * 24
			result.WriteString(strconv.FormatFloat(hours, 'f', -1, 64))
			result.WriteString("h")
		} else {
			result.WriteString(numStr)
			result.WriteString(unit)
		}
	}

	return time.ParseDuration(result.String())
}
