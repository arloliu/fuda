package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const treePadTop = 1 // blank line between border title and first node

// treeModel manages the tree panel: cursor, scrolling, rendering.
type treeModel struct {
	roots   []*Node
	flat    []*Node // visible flattened nodes (recomputed on expand/collapse/search)
	cursor  int     // index into flat
	offset  int     // scroll offset (first visible row)
	height  int     // visible rows (excludes top padding)
	width   int     // panel content width
	focused bool
}

func newTreeModel(roots []*Node) treeModel {
	t := treeModel{
		roots: roots,
	}
	t.reindex()

	return t
}

// reindex rebuilds the flat list from the current expand/visible state.
func (t *treeModel) reindex() {
	t.flat = Flatten(t.roots, t.flat)

	if t.cursor >= len(t.flat) {
		t.cursor = max(0, len(t.flat)-1)
	}

	t.clampScroll()
}

// selected returns the currently selected node, or nil.
func (t *treeModel) selected() *Node {
	if t.cursor >= 0 && t.cursor < len(t.flat) {
		return t.flat[t.cursor]
	}

	return nil
}

func (t *treeModel) moveUp() {
	if t.cursor > 0 {
		t.cursor--
		t.clampScroll()
	}
}

func (t *treeModel) moveDown() {
	if t.cursor < len(t.flat)-1 {
		t.cursor++
		t.clampScroll()
	}
}

func (t *treeModel) toggle() {
	if n := t.selected(); n != nil {
		n.Toggle()
		t.reindex()
	}
}

func (t *treeModel) expand() {
	if n := t.selected(); n != nil && n.HasChildren() && !n.Expanded {
		n.Expanded = true
		t.reindex()
	}
}

func (t *treeModel) collapse() {
	n := t.selected()
	if n == nil {
		return
	}

	if n.HasChildren() && n.Expanded {
		n.Expanded = false
		t.reindex()

		return
	}

	// If leaf or already collapsed, jump to parent.
	if n.Parent != nil {
		for i, f := range t.flat {
			if f == n.Parent {
				t.cursor = i
				t.clampScroll()

				break
			}
		}
	}
}

func (t *treeModel) expandAll() {
	for _, r := range t.roots {
		r.ExpandAll()
	}

	t.reindex()
}

func (t *treeModel) collapseAll() {
	for _, r := range t.roots {
		r.CollapseAll()
		r.Expanded = true // keep roots expanded
	}

	t.reindex()
}

// clickAt selects the node at the given row. If the click's X column lands
// on the expand/collapse icon (▼/▶), the node is also toggled.
func (t *treeModel) clickAt(x, y int) bool {
	idx := t.offset + y
	if idx < 0 || idx >= len(t.flat) {
		return false
	}

	t.cursor = idx

	n := t.flat[idx]
	if n.HasChildren() {
		// The rendered line layout is:
		//   cursorPrefix(2) + treePrefix + icon(1) + " " + name
		// Icon column = 2 + visible width of treePrefix.
		const cursorPrefixWidth = 2
		iconCol := cursorPrefixWidth + lipgloss.Width(t.treePrefix(n))

		// Toggle if the click is on the icon itself (± 1 col tolerance).
		if x >= iconCol-1 && x <= iconCol+1 {
			n.Toggle()
			t.reindex()
		}
	}

	return true
}

func (t *treeModel) clampScroll() {
	if t.height <= 0 {
		return
	}

	// Keep cursor visible.
	if t.cursor < t.offset {
		t.offset = t.cursor
	}

	if t.cursor >= t.offset+t.height {
		t.offset = t.cursor - t.height + 1
	}

	// Don't scroll past end.
	maxOffset := max(0, len(t.flat)-t.height)
	if t.offset > maxOffset {
		t.offset = maxOffset
	}
}

func (t *treeModel) setSize(width, height int) {
	t.width = width
	t.height = max(0, height-treePadTop)
	t.clampScroll()
}

// view renders the tree panel content (no border).
func (t *treeModel) view() string {
	if len(t.flat) == 0 {
		return treeMuted.Render("  No structs found")
	}

	var sb strings.Builder

	// Blank line for visual separation from the border title.
	for range treePadTop {
		sb.WriteByte('\n')
	}

	end := min(t.offset+t.height, len(t.flat))

	for i := t.offset; i < end; i++ {
		n := t.flat[i]
		line := t.renderNode(n, i == t.cursor)

		sb.WriteString(line)

		if i < end-1 {
			sb.WriteByte('\n')
		}
	}

	// Pad remaining lines if content is shorter than height.
	rendered := end - t.offset
	if padCount := t.height - rendered; padCount > 0 {
		padLine := strings.Repeat(" ", t.width)

		for i := rendered; i < t.height; i++ {
			if i > 0 {
				sb.WriteByte('\n')
			}

			sb.WriteString(padLine)
		}
	}

	return sb.String()
}

func (t *treeModel) renderNode(n *Node, isCursor bool) string {
	prefix := t.treePrefix(n)
	icon := t.nodeIcon(n)

	nameStyle := t.nodeNameStyle(n, isCursor)
	line := fmt.Sprintf("%s%s %s", prefix, icon, nameStyle.Render(n.Name))

	if isCursor && t.focused {
		line = treeCursor.Render("▸ ") + line
	} else {
		line = "  " + line
	}

	// Truncate to width.
	visible := lipgloss.Width(line)
	if visible > t.width && t.width > 0 {
		line = truncateVisible(line, t.width)
	}

	// Pad to width.
	pad := t.width - lipgloss.Width(line)
	if pad > 0 {
		line += strings.Repeat(" ", pad)
	}

	// Apply full-width background highlight on the selected row.
	if isCursor {
		line = t.selectedLineStyle(n).Render(line)
	}

	return line
}

func (t *treeModel) nodeNameStyle(n *Node, isCursor bool) lipgloss.Style {
	nameStyle := treeNormal
	if n.HasChildren() {
		nameStyle = treeNested
	}

	if n.IsRoot {
		nameStyle = treeRoot
	}

	// When this is the focused cursor row, boost the name foreground.
	if isCursor && t.focused {
		if n.IsRoot {
			nameStyle = treeRootHL
		} else if n.HasChildren() {
			nameStyle = treeNestedHL
		} else {
			nameStyle = nameStyle.Bold(true).Foreground(treeHighlightFg)
		}
	}

	return nameStyle
}

func (t *treeModel) selectedLineStyle(n *Node) lipgloss.Style {
	hlStyle := lipgloss.NewStyle().Background(treeHighlight)

	if !t.focused {
		return hlStyle
	}

	// Re-assert the foreground on the outer wrapper so it doesn't
	// clobber the inner styled text.
	switch {
	case n.IsRoot:
		return hlStyle.Foreground(treeRootHL.GetForeground()).Bold(true)
	case n.HasChildren():
		return hlStyle.Foreground(treeNestedHL.GetForeground()).Bold(true)
	default:
		return hlStyle
	}
}

// treePrefix builds the full prefix string for a node, including the vertical
// continuation lines (│) from all ancestor levels. This produces properly
// connected tree lines like:
//
//	├─ Host
//	├─▼ Database
//	│  ├─ Host
//	│  └─ Port
//	└─ Debug
func (t *treeModel) treePrefix(n *Node) string {
	if n.IsRoot || n.Parent == nil {
		return ""
	}

	// Count ancestor depth to size the stack array.
	depth := 0
	for cur := n.Parent; cur != nil && !cur.IsRoot; cur = cur.Parent {
		depth++
	}

	// Use a stack-allocated array for typical depths (≤16); only heap-
	// allocate for very deep trees. Ancestors are stored nearest-parent-
	// first, then iterated in reverse.
	const stackSize = 16

	var stackBuf [stackSize]*Node

	var ancestors []*Node
	if depth <= stackSize {
		ancestors = stackBuf[:0]
	} else {
		ancestors = make([]*Node, 0, depth)
	}

	for cur := n.Parent; cur != nil && !cur.IsRoot; cur = cur.Parent {
		ancestors = append(ancestors, cur)
	}

	// Build prefix from root-side down (reverse order).
	var sb strings.Builder
	sb.Grow(depth*3 + 2) //nolint:mnd // rough estimate: 3 chars per level + connector

	for i := len(ancestors) - 1; i >= 0; i-- {
		if isLastChild(ancestors[i]) {
			sb.WriteString("   ")
		} else {
			sb.WriteString(treeLine.Render("│") + "  ")
		}
	}

	// Connector for this node itself.
	if isLastChild(n) {
		sb.WriteString(treeLine.Render("└─"))
	} else {
		sb.WriteString(treeLine.Render("├─"))
	}

	return sb.String()
}

// isLastChild checks whether n is the last visible child of its parent.
func isLastChild(n *Node) bool {
	if n.Parent == nil {
		return true
	}

	// Walk backwards from end; first visible sibling we find tells us.
	siblings := n.Parent.Children

	for i := len(siblings) - 1; i >= 0; i-- {
		if siblings[i].Visible {
			return siblings[i] == n
		}
	}

	return true
}

func (t *treeModel) nodeIcon(n *Node) string {
	if !n.HasChildren() {
		return " "
	}

	if n.Expanded {
		return treeCollapsed.Render("▼")
	}

	return treeCollapsed.Render("▶")
}

// truncateVisible truncates a (possibly styled) string to fit within maxWidth
// visible columns. It uses exponential trimming to minimise lipgloss.Width
// calls (O(log n) instead of O(n)).
func truncateVisible(s string, maxWidth int) string {
	if maxWidth <= 3 { //nolint:mnd // minimum for ellipsis
		return "…"
	}

	target := maxWidth - 1 // leave room for "…"
	r := []rune(s)

	for len(r) > 0 {
		w := lipgloss.Width(string(r))
		if w <= target {
			break
		}

		// Estimate how many runes to drop (at least 1).
		drop := (w - target) / 2
		if drop < 1 {
			drop = 1
		}

		if drop > len(r) {
			drop = len(r)
		}

		r = r[:len(r)-drop]
	}

	return string(r) + "…"
}
