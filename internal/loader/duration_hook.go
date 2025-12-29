package loader

import (
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// preprocessDurationNodes walks a YAML node tree and converts duration values
// with 'd' suffix to hours format that time.Duration can parse.
// Examples: "2d" → "48h", "1d12h" → "36h"
//
// Note: Integer nanoseconds are NOT converted here because we cannot distinguish
// between an integer meant for time.Duration vs a regular int field at the YAML level.
// For integer duration values, use fuda.Duration type which has custom UnmarshalYAML.
func preprocessDurationNodes(node *yaml.Node) {
	if node == nil {
		return
	}

	switch node.Kind {
	case yaml.DocumentNode, yaml.SequenceNode:
		for _, child := range node.Content {
			preprocessDurationNodes(child)
		}
	case yaml.MappingNode:
		// Process key-value pairs
		for i := 0; i < len(node.Content); i += 2 {
			preprocessDurationNodes(node.Content[i+1]) // value node
		}
	case yaml.ScalarNode:
		// Convert 'd' suffix to hours (e.g., "2d" → "48h")
		if node.Tag == "!!str" && hasDaySuffix(node.Value) {
			if converted, ok := convertDaysToHours(node.Value); ok {
				node.Value = converted
			}
		}
	case yaml.AliasNode:
		// Aliases are resolved by yaml.Decode, no preprocessing needed
	}
}

// hasDaySuffix checks if a string contains a 'd' or 'D' suffix for days.
func hasDaySuffix(s string) bool {
	// Look for patterns like "2d", "1d12h", "0.5d", "-1d"
	s = strings.TrimPrefix(s, "-")
	s = strings.TrimPrefix(s, "+")

	for i, c := range s {
		if (c == 'd' || c == 'D') && i > 0 {
			// Check if preceded by a digit
			prev := s[i-1]
			if prev >= '0' && prev <= '9' {
				return true
			}
		}
	}

	return false
}

// convertDaysToHours converts duration strings with 'd' suffix to standard format.
// Examples: "2d" → "48h", "1d12h" → "36h", "0.5d" → "12h"
func convertDaysToHours(s string) (string, bool) {
	// Find and convert 'd' suffix for days to hours
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
				return "", false
			}

			hours := days * 24
			result.WriteString(strconv.FormatFloat(hours, 'f', -1, 64))
			result.WriteString("h")
		} else {
			result.WriteString(numStr)
			result.WriteString(unit)
		}
	}

	// Verify the result is a valid duration
	converted := result.String()
	if _, err := time.ParseDuration(converted); err != nil {
		return "", false
	}

	return converted, true
}
