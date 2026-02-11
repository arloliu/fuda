package tui

import (
	"strings"

	"github.com/arloliu/fuda/cmd/fuda-doc/internal/colors"
	"github.com/arloliu/fuda/cmd/fuda-doc/internal/docgen"
	"github.com/arloliu/fuda/cmd/fuda-doc/internal/docutil"
)

// yamlModel renders a YAML preview for the selected node's subtree.
type yamlModel struct {
	width  int
	height int
	offset int
	lines  []string
}

func newYAMLModel() yamlModel {
	return yamlModel{}
}

func (y *yamlModel) setSize(width, height int) {
	y.width = width
	y.height = height
}

// update rebuilds the YAML preview for the given node.
func (y *yamlModel) update(n *Node) {
	y.offset = 0
	y.lines = nil

	if n == nil {
		return
	}

	if n.IsRoot && n.StructDoc != nil {
		y.renderFields(n.StructDoc.Fields, 0)

		return
	}

	if n.Field != nil {
		if n.HasChildren() {
			key := docutil.YAMLKey(n.Field)
			y.addLine(yamlKey(key) + yamlColon())
			y.renderFields(n.Field.Nested, 1)
		} else {
			y.renderSingleField(n.Field, 0)
		}
	}
}

func (y *yamlModel) renderFields(fields []docgen.FieldInfo, indent int) {
	indentStr := strings.Repeat("  ", indent)

	for _, f := range fields {
		if !docutil.IsExported(f.Name) {
			continue
		}

		key := docutil.YAMLKey(&f)
		if key == "-" {
			continue
		}

		if len(f.Nested) > 0 {
			y.addLine(indentStr + yamlKey(key) + yamlColon())
			y.renderFields(f.Nested, indent+1)

			continue
		}

		val := docutil.YAMLDefault(&f)
		y.addLine(indentStr + yamlKey(key) + yamlColon() + " " + yamlVal(val))
	}
}

func (y *yamlModel) renderSingleField(f *docgen.FieldInfo, indent int) {
	indentStr := strings.Repeat("  ", indent)
	key := docutil.YAMLKey(f)
	val := docutil.YAMLDefault(f)
	y.addLine(indentStr + yamlKey(key) + yamlColon() + " " + yamlVal(val))
}

func (y *yamlModel) addLine(s string) {
	y.lines = append(y.lines, s)
}

func (y *yamlModel) scrollUp() {
	if y.offset > 0 {
		y.offset--
	}
}

func (y *yamlModel) scrollDown() {
	maxOff := max(0, len(y.lines)-y.height)
	if y.offset < maxOff {
		y.offset++
	}
}

// view renders the YAML panel content.
func (y *yamlModel) view() string {
	if len(y.lines) == 0 {
		return detailMuted.Render("  Select a field to preview YAML")
	}

	var sb strings.Builder

	// Reserve first line as spacing below the border title.
	sb.WriteByte('\n')

	visible := y.height - 1 // one line used by the spacer above
	end := min(y.offset+visible, len(y.lines))

	for i := y.offset; i < end; i++ {
		sb.WriteString(y.lines[i])

		if i < end-1 {
			sb.WriteByte('\n')
		}
	}

	// Pad
	rendered := end - y.offset
	if padCount := visible - rendered; padCount > 0 {
		padLine := strings.Repeat(" ", y.width)

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

func yamlKey(key string) string {
	return colors.YAMLKeyStyle.Render(key)
}

func yamlColon() string {
	return colors.YAMLPunctStyle.Render(":")
}

func yamlVal(val string) string {
	return colors.YAMLValueStyle.Render(val)
}
