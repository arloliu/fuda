package tui

import (
	"fmt"
	"strings"

	"github.com/arloliu/fuda/cmd/fuda-doc/internal/docgen"
	"github.com/arloliu/fuda/cmd/fuda-doc/internal/docutil"
)

// detailModel renders the detail view for the selected node.
type detailModel struct {
	width  int
	height int
	offset int // scroll offset for long content
	lines  []string
}

func newDetailModel() detailModel {
	return detailModel{}
}

func (d *detailModel) setSize(width, height int) {
	d.width = width
	d.height = height
}

// update rebuilds the detail content for the given node.
func (d *detailModel) update(n *Node) {
	d.offset = 0
	d.lines = nil

	if n == nil {
		return
	}

	if n.IsRoot && n.StructDoc != nil {
		d.buildStructDetail(n)

		return
	}

	if n.Field != nil {
		d.buildFieldDetail(n)
	}
}

func (d *detailModel) buildStructDetail(n *Node) {
	doc := n.StructDoc

	d.addLine(detailTitle.Render("📦 " + doc.Name))
	d.addLine("")

	if doc.Doc != "" {
		wrapped := docutil.WordWrap(doc.Doc, d.width-2) //nolint:mnd // padding
		for _, line := range wrapped {
			d.addLine(detailDesc.Render(line))
		}

		d.addLine("")
	}

	d.addLine(detailLabel.Render("Fields:") + " " +
		detailValue.Render(fmt.Sprintf("%d top-level", countExported(doc.Fields))))

	total := countAllExported(doc.Fields)
	d.addLine(detailLabel.Render("Total:") + "  " +
		detailValue.Render(fmt.Sprintf("%d (including nested)", total)))
}

func (d *detailModel) buildFieldDetail(n *Node) {
	f := n.Field

	// Title
	title := "▸ " + f.Name
	if n.HasChildren() {
		title = "▸ " + f.Name + " (struct)"
	}

	d.addLine(detailTitle.Render(title))
	d.addLine("")

	// Properties table
	d.addProp("YAML key", docutil.YAMLKey(f))
	d.addProp("Type", detailType.Render(f.Type))

	if v := f.Tags["default"]; v != "" {
		d.addProp("Default", v)
	}

	if v := f.Tags["env"]; v != "" {
		d.addProp("Env", v)
	}

	if v := f.Tags["ref"]; v != "" {
		d.addProp("Ref", v)
	}

	if v := f.Tags["refFrom"]; v != "" {
		d.addProp("Ref from", v)
	}

	if v := f.Tags["dsn"]; v != "" {
		d.addProp("DSN tmpl", v)
	}

	if v := f.Tags["validate"]; v != "" {
		d.addProp("Validate", v)
	}

	if v := f.Tags["required"]; v != "" {
		d.addProp("Required", v)
	}

	// Description
	if f.Description != "" {
		d.addLine("")
		d.addLine(detailLabel.Render("Description:"))

		wrapped := docutil.WordWrap(f.Description, d.width-4) //nolint:mnd // padding
		for _, line := range wrapped {
			d.addLine("  " + detailDesc.Render(line))
		}
	}

	// Nested type info
	if n.HasChildren() && f.NestedType != "" {
		d.addLine("")
		d.addLine(detailLabel.Render("Nested type:") + " " + treeNested.Render(f.NestedType))
		d.addLine(detailLabel.Render("Sub-fields:") + "  " +
			detailValue.Render(fmt.Sprintf("%d", countExported(f.Nested))))
	}
}

func (d *detailModel) addLine(s string) {
	d.lines = append(d.lines, s)
}

func (d *detailModel) addProp(label, value string) {
	padded := docutil.PadRight(label+":", 13)
	d.addLine("  " + detailLabel.Render(padded) + " " + detailValue.Render(value))
}

func (d *detailModel) scrollUp() {
	if d.offset > 0 {
		d.offset--
	}
}

func (d *detailModel) scrollDown() {
	maxOff := max(0, len(d.lines)-d.height)
	if d.offset < maxOff {
		d.offset++
	}
}

// view renders the detail panel content.
func (d *detailModel) view() string {
	if len(d.lines) == 0 {
		return detailMuted.Render("  Select a field to view details")
	}

	var sb strings.Builder

	// Reserve first line as spacing below the border title.
	sb.WriteByte('\n')

	visible := d.height - 1 // one line used by the spacer above
	end := min(d.offset+visible, len(d.lines))

	for i := d.offset; i < end; i++ {
		sb.WriteString(d.lines[i])

		if i < end-1 {
			sb.WriteByte('\n')
		}
	}

	// Pad
	rendered := end - d.offset
	if padCount := visible - rendered; padCount > 0 {
		padLine := strings.Repeat(" ", d.width)

		for i := rendered; i < visible; i++ {
			if i > 0 {
				sb.WriteByte('\n')
			}

			sb.WriteString(padLine)
		}
	}

	return sb.String()
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func countExported(fields []docgen.FieldInfo) int {
	count := 0

	for _, f := range fields {
		if docutil.IsExported(f.Name) {
			count++
		}
	}

	return count
}

func countAllExported(fields []docgen.FieldInfo) int {
	count := 0

	for _, f := range fields {
		if !docutil.IsExported(f.Name) {
			continue
		}

		count++

		if len(f.Nested) > 0 {
			count += countAllExported(f.Nested)
		}
	}

	return count
}
