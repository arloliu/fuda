package docgen

import (
	"fmt"
	"io"
	"strings"

	"github.com/arloliu/fuda/cmd/fuda-doc/internal/colors"
	"github.com/arloliu/fuda/cmd/fuda-doc/internal/docutil"
)

// ASCIIPrinter generates terminal-friendly documentation with adaptive colors.
// Colors automatically adjust for light/dark backgrounds and degrade gracefully
// across TrueColor, ANSI 256, and basic ANSI color profiles via lipgloss
// CompleteAdaptiveColor.
type ASCIIPrinter struct {
	w         io.Writer
	seenTypes map[string]bool
}

// NewASCIIPrinter creates a new ASCIIPrinter.
func NewASCIIPrinter(w io.Writer) *ASCIIPrinter {
	return &ASCIIPrinter{w: w, seenTypes: map[string]bool{}}
}

// Print generates terminal-friendly documentation.
func (a *ASCIIPrinter) Print(structName string, doc string, fields []FieldInfo) {
	a.printHeader(structName)

	if doc != "" {
		a.printf("  %s\n\n", doc)
	}

	a.printUsage(structName)
	a.printYAMLExample(fields)
	a.printFieldReference(fields, 0)
}

// ---------------------------------------------------------------------------
// Header
// ---------------------------------------------------------------------------

func (a *ASCIIPrinter) printHeader(structName string) {
	title := fmt.Sprintf(" %s Configuration ", structName)
	width := len(title) + 4
	bar := strings.Repeat("═", width)

	s := colors.HeaderStyle
	a.printf("\n%s\n", s.Render("╔"+bar+"╗"))
	a.printf("%s\n", s.Render("║  "+title+"  ║"))
	a.printf("%s\n\n", s.Render("╚"+bar+"╝"))
}

// ---------------------------------------------------------------------------
// Usage
// ---------------------------------------------------------------------------

func (a *ASCIIPrinter) printUsage(structName string) {
	a.printSectionTitle("Usage")

	dim := colors.DimStyle
	a.printf("  %s\n", dim.Render("var cfg "+structName))
	a.printf("  %s\n", dim.Render(`if err := fuda.Load(&cfg, "config.yaml"); err != nil {`))
	a.printf("  %s\n", dim.Render("    panic(err)"))
	a.printf("  %s\n\n", dim.Render("}"))

	a.printf("  %s\n", colors.MutedStyle.Render("Resolution order (highest priority last):"))

	num := colors.LabelStyle
	a.printf("  %s default tag   %s YAML/JSON file   %s Env vars   %s ref/refFrom   %s DSN template\n\n",
		num.Render("1."), num.Render("2."), num.Render("3."), num.Render("4."), num.Render("5."))
}

// ---------------------------------------------------------------------------
// YAML Example
// ---------------------------------------------------------------------------

func (a *ASCIIPrinter) printYAMLExample(fields []FieldInfo) {
	a.printSectionTitle("Configuration Example")
	a.printYAMLBlock(fields, 1)
	a.printf("\n")
}

func (a *ASCIIPrinter) printYAMLBlock(fields []FieldInfo, indent int) {
	indentStr := strings.Repeat("  ", indent)
	key := colors.CyanStyle
	val := colors.ValueStyle

	for _, field := range fields {
		if !docutil.IsExported(field.Name) {
			continue
		}

		yamlKey := docutil.YAMLKey(&field)
		if yamlKey == "-" {
			continue
		}

		if len(field.Nested) > 0 {
			a.printf("%s%s:\n", indentStr, key.Render(yamlKey))
			a.printYAMLBlock(field.Nested, indent+1)

			continue
		}

		v := docutil.YAMLDefault(&field)
		a.printf("%s%s: %s\n", indentStr, key.Render(yamlKey), val.Render(v))
	}
}

// ---------------------------------------------------------------------------
// Field Reference
// ---------------------------------------------------------------------------

func (a *ASCIIPrinter) printFieldReference(fields []FieldInfo, depth int) {
	if depth == 0 {
		a.printSectionTitle("Field Reference")
	}

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

	indent := strings.Repeat("  ", depth)

	if len(scalars) > 0 {
		a.printFieldTableASCII(scalars, indent)
		a.printf("\n")
		a.printFieldDetailsASCII(scalars, indent)
	}

	for _, ns := range nested {
		a.printSubsectionTitle(ns.Name, depth)

		if ns.Description != "" {
			a.printf("%s  %s\n\n", indent, colors.MutedStyle.Render(docutil.FirstSentence(ns.Description)))
		}

		if a.seenTypes[ns.NestedType] {
			dedup := "↳ Same structure as " + colors.Bold(colors.Muted).Render(ns.NestedType) + " above"
			a.printf("%s  %s\n\n", indent, colors.MutedStyle.Render(dedup))
			a.printFieldTableASCII(ns.Nested, indent)
			a.printf("\n")

			continue
		}

		if ns.NestedType != "" {
			a.seenTypes[ns.NestedType] = true
		}

		a.printFieldReference(ns.Nested, depth+1)
	}
}

// ---------------------------------------------------------------------------
// ASCII table rendering
// ---------------------------------------------------------------------------

func (a *ASCIIPrinter) printFieldTableASCII(fields []FieldInfo, indent string) {
	type row struct {
		field, typ, def, source string
	}

	rows := make([]row, 0, len(fields))

	for _, f := range fields {
		if !docutil.IsExported(f.Name) {
			continue
		}

		yamlKey := docutil.YAMLKey(&f)

		fieldStr := f.Name
		if yamlKey != "" && yamlKey != "-" {
			fieldStr = f.Name + " (" + yamlKey + ")"
		}

		rows = append(rows, row{
			field:  fieldStr,
			typ:    f.Type,
			def:    plainDefault(f),
			source: plainSource(f),
		})
	}

	if len(rows) == 0 {
		return
	}

	// Measure widths
	wField, wType, wDef, wSrc := 5, 4, 7, 12

	for _, r := range rows {
		wField = maxInt(wField, len(r.field))
		wType = maxInt(wType, len(r.typ))
		wDef = maxInt(wDef, len(r.def))
		wSrc = maxInt(wSrc, len(r.source))
	}

	// Cap widths
	wField = minInt(wField, 32)
	wType = minInt(wType, 18)
	wDef = minInt(wDef, 24)
	wSrc = minInt(wSrc, 34)

	totalW := wField + wType + wDef + wSrc + 13 // separators + padding

	// Header
	divider := indent + "  " + colors.MutedStyle.Render(strings.Repeat("─", totalW))

	hdr := colors.Bold(colors.Text)
	a.printf("%s\n", divider)
	a.printf("%s  %s │ %s │ %s │ %s\n",
		indent,
		hdr.Render(padOrTrunc("Field", wField)),
		hdr.Render(padOrTrunc("Type", wType)),
		hdr.Render(padOrTrunc("Default", wDef)),
		hdr.Render(padOrTrunc("Env / Source", wSrc)))
	a.printf("%s\n", divider)

	// Rows
	fld := colors.FieldStyle
	typ := colors.TypeStyle

	for _, r := range rows {
		a.printf("%s  %s │ %s │ %-*s │ %-*s\n",
			indent,
			fld.Render(padOrTrunc(r.field, wField)),
			typ.Render(padOrTrunc(r.typ, wType)),
			wDef, padOrTrunc(r.def, wDef),
			wSrc, padOrTrunc(r.source, wSrc))
	}

	a.printf("%s\n", divider)
}

func (a *ASCIIPrinter) printFieldDetailsASCII(fields []FieldInfo, indent string) {
	for _, f := range fields {
		if !docutil.IsExported(f.Name) {
			continue
		}

		yamlKey := docutil.YAMLKey(&f)

		// Field name heading
		a.printf("%s  %s\n", indent, colors.FieldStyle.Render("▸ "+f.Name))

		// Properties
		a.printPropRow(indent, "YAML key", yamlKey)
		a.printPropRow(indent, "Type", f.Type)

		if v := f.Tags["default"]; v != "" {
			a.printPropRow(indent, "Default", v)
		}

		if v := f.Tags["env"]; v != "" {
			a.printPropRow(indent, "Env", v)
		}

		if v := f.Tags["ref"]; v != "" {
			a.printPropRow(indent, "Ref", v)
		}

		if v := f.Tags["refFrom"]; v != "" {
			a.printPropRow(indent, "Ref from", v)
		}

		if v := f.Tags["dsn"]; v != "" {
			a.printPropRow(indent, "DSN tmpl", v)
		}

		if v := f.Tags["validate"]; v != "" {
			a.printPropRow(indent, "Validate", v)
		}

		// Description
		if f.Description != "" {
			a.printf("\n")
			a.printDescriptionASCII(f.Description, indent+"    ")
		}

		a.printf("\n")
	}
}

func (a *ASCIIPrinter) printPropRow(indent, label, value string) {
	lbl := colors.LabelStyle
	val := colors.ValueStyle
	a.printf("%s    %s %s\n", indent, lbl.Render(docutil.PadRight(label+":", 13)), val.Render(value))
}

func (a *ASCIIPrinter) printDescriptionASCII(desc, indent string) {
	desc = strings.TrimSpace(desc)
	lines := strings.Split(desc, "\n")

	var paragraph []string

	flushParagraph := func() {
		if len(paragraph) > 0 {
			text := strings.Join(paragraph, " ")
			wrapped := docutil.WordWrap(text, 76-len(indent))

			for _, wl := range wrapped {
				a.printf("%s%s\n", indent, wl)
			}

			paragraph = nil
		}
	}

	dim := colors.DimStyle

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			flushParagraph()
			a.printf("\n")

			continue
		}

		isIndented := strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "\t")
		if isIndented {
			flushParagraph()

			codeLine := strings.TrimPrefix(strings.TrimPrefix(line, "  "), "\t")
			a.printf("%s  %s\n", indent, dim.Render(codeLine))
		} else {
			paragraph = append(paragraph, trimmed)
		}
	}

	flushParagraph()
}

// ---------------------------------------------------------------------------
// Section titles
// ---------------------------------------------------------------------------

func (a *ASCIIPrinter) printSectionTitle(title string) {
	bar := strings.Repeat("─", len(title)+2)
	s := colors.SectionStyle
	a.printf("  %s\n", s.Render("┌"+bar+"┐"))
	a.printf("  %s\n", s.Render("│ "+title+" │"))
	a.printf("  %s\n\n", s.Render("└"+bar+"┘"))
}

func (a *ASCIIPrinter) printSubsectionTitle(title string, depth int) {
	indent := strings.Repeat("  ", depth)
	bar := strings.Repeat("─", len(title)+4)
	a.printf("%s  %s\n\n", indent, colors.SubsectionStyle.Render("── "+title+" "+bar))
}

// ---------------------------------------------------------------------------
// Display helpers
// ---------------------------------------------------------------------------

func plainDefault(f FieldInfo) string {
	v := f.Tags["default"]
	if v == "" {
		return "-"
	}

	return docutil.Truncate(v, 24)
}

func plainSource(f FieldInfo) string {
	var parts []string

	if v := f.Tags["env"]; v != "" {
		parts = append(parts, v)
	}

	if v := f.Tags["ref"]; v != "" {
		parts = append(parts, "ref:"+docutil.Truncate(v, 28))
	}

	if v := f.Tags["refFrom"]; v != "" {
		parts = append(parts, "from:"+v)
	}

	if _, ok := f.Tags["dsn"]; ok {
		parts = append(parts, "dsn ✓")
	}

	if len(parts) == 0 {
		return "-"
	}

	return strings.Join(parts, ", ")
}

// ---------------------------------------------------------------------------
// Utilities
// ---------------------------------------------------------------------------

func (a *ASCIIPrinter) printf(format string, args ...any) {
	_, _ = fmt.Fprintf(a.w, format, args...)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func padOrTrunc(s string, width int) string {
	r := []rune(s)
	if len(r) > width {
		return string(r[:width-1]) + "…"
	}

	return s + strings.Repeat(" ", width-len(r))
}
