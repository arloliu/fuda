// Package docutil provides shared types and helper functions for formatting
// field metadata, YAML output, and text layout across the fuda-doc tool.
package docutil

import (
	"strings"
)

// FieldInfo represents metadata about a struct field.
type FieldInfo struct {
	Name        string
	Type        string
	Description string            // From comments (GoDoc)
	Tags        map[string]string // Parsed tags (default, env, etc.)
	Nested      []FieldInfo       // For nested structs
	NestedType  string            // Type name of the nested struct
}

// YAMLKey returns the YAML key for a field, preferring the yaml tag, then
// json tag, then a camelCase-derived name.
func YAMLKey(f *FieldInfo) string {
	if f == nil || len(f.Name) == 0 {
		return ""
	}

	key := f.Tags["yaml"]
	if key == "" {
		key = f.Tags["json"]
	}

	if key == "" {
		return strings.ToLower(f.Name[:1]) + f.Name[1:]
	}

	if idx := strings.Index(key, ","); idx != -1 {
		key = key[:idx]
	}

	return key
}

// YAMLDefault returns a YAML-friendly default value string for a field,
// choosing appropriate formatting based on the field's type.
func YAMLDefault(f *FieldInfo) string {
	d := f.Tags["default"]

	switch {
	case strings.HasPrefix(f.Type, "map"):
		return FormatMapDefault(d)
	case strings.HasPrefix(f.Type, "[]byte"):
		if d == "" {
			return "null"
		}

		return d
	case strings.HasPrefix(f.Type, "[]"):
		return FormatSliceDefault(d)
	case f.Type == "string":
		if d == "" {
			return `""`
		}

		return `"` + d + `"`
	case f.Type == "bool":
		if d == "" {
			return "false"
		}

		return d
	case f.Type == "time.Duration":
		if d == "" {
			return "0s"
		}

		return d
	case strings.Contains(f.Type, "int") || strings.Contains(f.Type, "float"):
		if d == "" {
			return "0"
		}

		return d
	default:
		if d == "" {
			return "null"
		}

		return d
	}
}

// FormatMapDefault formats a comma-separated "k:v,k:v" default into YAML map
// syntax like "{ k: v, k: v }".
func FormatMapDefault(d string) string {
	if d == "" {
		return "{}"
	}

	var entries []string

	for _, pair := range strings.Split(d, ",") {
		kv := strings.SplitN(pair, ":", 2)
		if len(kv) == 2 {
			entries = append(entries, strings.TrimSpace(kv[0])+": "+strings.TrimSpace(kv[1]))
		}
	}

	if len(entries) > 0 {
		return "{ " + strings.Join(entries, ", ") + " }"
	}

	return d
}

// FormatSliceDefault formats a comma-separated default into YAML slice syntax
// like "[a, b, c]".
func FormatSliceDefault(d string) string {
	if d == "" {
		return "[]"
	}

	items := strings.Split(d, ",")
	trimmed := make([]string, len(items))

	for i, item := range items {
		trimmed[i] = strings.TrimSpace(item)
	}

	return "[" + strings.Join(trimmed, ", ") + "]"
}

// IsExported returns true if a Go identifier starts with an uppercase letter.
func IsExported(name string) bool {
	if len(name) == 0 {
		return false
	}

	return name[0] >= 'A' && name[0] <= 'Z'
}

// WordWrap splits text into lines that fit within the given column width.
// It breaks on word boundaries only.
func WordWrap(text string, width int) []string {
	if width <= 0 {
		width = 76
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var lines []string

	current := words[0]

	for _, word := range words[1:] {
		if len(current)+1+len(word) > width {
			lines = append(lines, current)
			current = word
		} else {
			current += " " + word
		}
	}

	lines = append(lines, current)

	return lines
}

// PadRight pads a string with spaces on the right to reach the given width.
func PadRight(s string, width int) string {
	r := []rune(s)
	if len(r) >= width {
		return s
	}

	return s + strings.Repeat(" ", width-len(r))
}

// FirstLine returns the first line of s (up to the first newline).
// If s contains no newline, it is returned unchanged.
func FirstLine(s string) string {
	if idx := strings.IndexByte(s, '\n'); idx != -1 {
		return s[:idx]
	}

	return s
}

// Truncate shortens a string to maxLen, appending "..." if truncated.
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	return s[:maxLen-3] + "..."
}

// FirstSentence extracts the first sentence (up to the first ". ") or the
// first line, whichever is shorter.
func FirstSentence(s string) string {
	s = strings.TrimSpace(s)
	lines := strings.SplitN(s, "\n", 2)
	first := strings.TrimSpace(lines[0])

	if idx := strings.Index(first, ". "); idx > 0 {
		return first[:idx+1]
	}

	if len(first) > 100 {
		if idx := strings.LastIndex(first[:97], " "); idx > 40 {
			return first[:idx] + "..."
		}

		return first[:97] + "..."
	}

	return first
}
