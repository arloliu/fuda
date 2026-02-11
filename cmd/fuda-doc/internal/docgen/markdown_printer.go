package docgen

import (
	"fmt"
	"io"
	"strings"

	"github.com/arloliu/fuda/cmd/fuda-doc/internal/docutil"
)

// MarkdownPrinter handles markdown output generation.
type MarkdownPrinter struct {
	w         io.Writer
	seenTypes map[string]bool // tracks struct types already documented in detail
}

// NewMarkdownPrinter creates a new MarkdownPrinter that writes to the given writer.
func NewMarkdownPrinter(w io.Writer) *MarkdownPrinter {
	return &MarkdownPrinter{w: w, seenTypes: map[string]bool{}}
}

// Print generates Markdown documentation for the given fields.
func (p *MarkdownPrinter) Print(structName string, doc string, fields []FieldInfo) {
	// Header
	p.printf("# %s\n\n", structName)
	if doc != "" {
		p.printf("%s\n\n", p.formatDescriptionBlock(doc))
	}

	// Usage
	p.printUsage(structName)

	// YAML Example
	p.printf("## Configuration Example\n\n")
	p.printf("```yaml\n")
	p.printYAMLBlock(fields, 0)
	p.printf("```\n\n")

	// Field Reference
	p.printf("---\n\n")
	p.printf("## Field Reference\n\n")
	p.printSectionFields(fields, 2)
}

// ---------------------------------------------------------------------------
// Usage
// ---------------------------------------------------------------------------

func (p *MarkdownPrinter) printUsage(structName string) {
	p.printf("## Usage\n\n")
	p.printf("```go\n")
	p.printf("var cfg %s\n", structName)
	p.printf("if err := fuda.Load(&cfg, \"config.yaml\"); err != nil {\n")
	p.printf("    panic(err)\n")
	p.printf("}\n")
	p.printf("```\n\n")
	p.printf("Configuration is resolved in the following order (highest priority last):\n\n")
	p.printf("1. `default` tag values\n")
	p.printf("2. YAML / JSON configuration file\n")
	p.printf("3. Environment variables (`env` tag)\n")
	p.printf("4. External sources (`ref`, `refFrom` tags — files, Vault, HTTP)\n")
	p.printf("5. Computed fields (`dsn` tag — Go template)\n\n")
}

// ---------------------------------------------------------------------------
// YAML Example
// ---------------------------------------------------------------------------

func (p *MarkdownPrinter) printYAMLBlock(fields []FieldInfo, indent int) {
	indentStr := strings.Repeat("  ", indent)

	for _, field := range fields {
		if !docutil.IsExported(field.Name) {
			continue
		}

		yamlKey := docutil.YAMLKey(&field)
		if yamlKey == "-" {
			continue
		}

		if len(field.Nested) > 0 {
			p.printf("%s%s:\n", indentStr, yamlKey)
			p.printYAMLBlock(field.Nested, indent+1)

			continue
		}

		val := docutil.YAMLDefault(&field)
		p.printf("%s%s: %s\n", indentStr, yamlKey, val)
	}
}

// ---------------------------------------------------------------------------
// Field Reference — per-section tables + detail blocks
// ---------------------------------------------------------------------------

func (p *MarkdownPrinter) printSectionFields(fields []FieldInfo, headingLevel int) {
	var scalars []FieldInfo
	var nested []FieldInfo

	for _, f := range fields {
		if !docutil.IsExported(f.Name) {
			continue
		}

		if len(f.Nested) > 0 {
			nested = append(nested, f)
		} else {
			scalars = append(scalars, f)
		}
	}

	// Scalar fields — compact table then detailed blocks
	if len(scalars) > 0 {
		p.printFieldTable(scalars)
		p.printf("\n")
		p.printFieldDetails(scalars)
	}

	// Nested structs — each gets its own sub-section
	for _, ns := range nested {
		level := min(headingLevel+1, 4)
		hdr := strings.Repeat("#", level)

		p.printf("%s %s\n\n", hdr, ns.Name)

		if ns.Description != "" {
			p.printf("%s\n\n", docutil.FirstSentence(ns.Description))
		}

		// If this struct type was already fully documented, show a back-reference
		if p.seenTypes[ns.NestedType] {
			p.printf("> Same structure as [%s](#%s) above.\n\n",
				ns.NestedType, strings.ToLower(ns.NestedType))
			p.printFieldTable(ns.Nested)
			p.printf("\n")

			continue
		}

		if ns.NestedType != "" {
			p.seenTypes[ns.NestedType] = true
		}

		p.printSectionFields(ns.Nested, level)
	}
}

func (p *MarkdownPrinter) printFieldTable(fields []FieldInfo) {
	p.printf("| Field | Type | Default | Env / Source |\n")
	p.printf("|:------|:-----|:--------|:------------|\n")

	for _, f := range fields {
		if !docutil.IsExported(f.Name) {
			continue
		}

		yamlKey := docutil.YAMLKey(&f)

		fieldCol := fmt.Sprintf("`%s`", f.Name)
		if yamlKey != "" && yamlKey != "-" {
			fieldCol = fmt.Sprintf("`%s`<br><sub>`%s`</sub>", f.Name, yamlKey)
		}

		p.printf("| %s | `%s` | %s | %s |\n", fieldCol, f.Type, defaultDisplay(f), sourceDisplay(f))
	}
}

func (p *MarkdownPrinter) printFieldDetails(fields []FieldInfo) {
	for _, f := range fields {
		if !docutil.IsExported(f.Name) {
			continue
		}

		yamlKey := docutil.YAMLKey(&f)

		// Heading
		p.printf("#### %s\n\n", f.Name)

		// Properties as a definition table
		p.printf("| Property | Value |\n")
		p.printf("|:---------|:------|\n")
		p.printf("| **YAML key** | `%s` |\n", yamlKey)
		p.printf("| **Type** | `%s` |\n", f.Type)

		if v := f.Tags["default"]; v != "" {
			p.printf("| **Default** | `%s` |\n", v)
		}

		if v := f.Tags["env"]; v != "" {
			p.printf("| **Env** | `%s` |\n", v)
		}

		if v := f.Tags["ref"]; v != "" {
			p.printf("| **Ref** | `%s` |\n", v)
		}

		if v := f.Tags["refFrom"]; v != "" {
			p.printf("| **Ref from** | `%s` |\n", v)
		}

		if v := f.Tags["dsn"]; v != "" {
			p.printf("| **DSN template** | `%s` |\n", v)
		}

		if v := f.Tags["validate"]; v != "" {
			p.printf("| **Validation** | `%s` |\n", v)
		}

		p.printf("\n")

		// Description body
		if f.Description != "" {
			p.printf("%s\n\n", p.formatDescriptionBlock(f.Description))
		}

		p.printf("---\n\n")
	}
}

// ---------------------------------------------------------------------------
// Display helpers
// ---------------------------------------------------------------------------

func defaultDisplay(f FieldInfo) string {
	v := f.Tags["default"]
	if v == "" {
		return "-"
	}

	return "`" + docutil.Truncate(v, 24) + "`"
}

func sourceDisplay(f FieldInfo) string {
	var parts []string

	if v := f.Tags["env"]; v != "" {
		parts = append(parts, "`"+v+"`")
	}

	if v := f.Tags["ref"]; v != "" {
		parts = append(parts, "ref: `"+docutil.Truncate(v, 28)+"`")
	}

	if v := f.Tags["refFrom"]; v != "" {
		parts = append(parts, "refFrom: `"+v+"`")
	}

	if _, ok := f.Tags["dsn"]; ok {
		parts = append(parts, "dsn ✓")
	}

	if len(parts) == 0 {
		return "-"
	}

	return strings.Join(parts, "<br>")
}

// ---------------------------------------------------------------------------
// Description formatting (godoc aware)
// ---------------------------------------------------------------------------

// formatDescriptionBlock formats a multiline godoc comment for markdown output.
//
// Godoc conventions:
//   - Consecutive non-blank lines are joined into a single paragraph.
//   - Blank lines (empty //) create paragraph breaks.
//   - Indented lines are preformatted code blocks.
//   - Lines starting with "Example" become highlighted code blocks.
func (p *MarkdownPrinter) formatDescriptionBlock(desc string) string {
	if desc == "" {
		return ""
	}

	desc = strings.TrimSpace(desc)
	lines := strings.Split(desc, "\n")

	var result []string
	var paragraph []string

	inCodeBlock := false

	flushParagraph := func() {
		if len(paragraph) > 0 {
			result = append(result, strings.Join(paragraph, " "))
			paragraph = nil
		}
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Empty line -> paragraph break
		if trimmed == "" {
			if inCodeBlock {
				result = append(result, "```")
				inCodeBlock = false
			}

			flushParagraph()
			result = append(result, "")

			continue
		}

		// Detect "Example:" header
		if strings.HasPrefix(trimmed, "Example:") || strings.HasPrefix(trimmed, "Example ") {
			flushParagraph()
			result = append(result, "", "**"+trimmed+"**", "```")
			inCodeBlock = true

			continue
		}

		// Indented / preformatted line
		isIndented := strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "\t")
		if isIndented {
			if !inCodeBlock {
				flushParagraph()
				result = append(result, "```")
				inCodeBlock = true
			}

			codeLine := strings.TrimPrefix(strings.TrimPrefix(line, "  "), "\t")
			result = append(result, codeLine)
		} else {
			if inCodeBlock {
				result = append(result, "```")
				inCodeBlock = false
			}

			paragraph = append(paragraph, trimmed)
		}
	}

	flushParagraph()

	if inCodeBlock {
		result = append(result, "```")
	}

	return strings.Join(result, "\n")
}

// ---------------------------------------------------------------------------
// Utilities
// ---------------------------------------------------------------------------

func (p *MarkdownPrinter) printf(format string, args ...any) {
	_, _ = fmt.Fprintf(p.w, format, args...)
}
