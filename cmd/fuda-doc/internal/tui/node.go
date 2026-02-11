package tui

import (
	"github.com/arloliu/fuda/cmd/fuda-doc/internal/docgen"
	"github.com/arloliu/fuda/cmd/fuda-doc/internal/docutil"
)

// Node represents a single item in the struct tree. It wraps a FieldInfo
// (or a top-level struct) and adds UI state such as expand/collapse and
// search visibility.
type Node struct {
	// Display
	Name       string // display name (field name or struct name)
	StructDoc  *docgen.StructDoc
	Field      *docgen.FieldInfo
	Children   []*Node
	Parent     *Node
	Depth      int
	IsRoot     bool // true for top-level struct nodes
	NestedType string

	// UI state
	Expanded bool
	Visible  bool // search filtering
}

// BuildTree converts a slice of StructDocs into a forest of Nodes.
func BuildTree(docs []docgen.StructDoc) []*Node {
	roots := make([]*Node, 0, len(docs))

	for i := range docs {
		doc := &docs[i]
		root := &Node{
			Name:      doc.Name,
			StructDoc: doc,
			IsRoot:    true,
			Depth:     0,
			Expanded:  true,
			Visible:   true,
		}
		root.Children = buildChildren(doc.Fields, root, 1)
		roots = append(roots, root)
	}

	return roots
}

func buildChildren(fields []docgen.FieldInfo, parent *Node, depth int) []*Node {
	nodes := make([]*Node, 0, len(fields))

	for i := range fields {
		f := &fields[i]

		if !docutil.IsExported(f.Name) {
			continue
		}

		n := &Node{
			Name:       f.Name,
			Field:      f,
			Parent:     parent,
			Depth:      depth,
			NestedType: f.NestedType,
			Expanded:   true,
			Visible:    true,
		}

		if len(f.Nested) > 0 {
			n.Children = buildChildren(f.Nested, n, depth+1)
		}

		nodes = append(nodes, n)
	}

	return nodes
}

// Flatten returns the visible, expanded nodes in tree order for rendering.
// If dst is provided, it reuses the backing array to reduce allocations.
func Flatten(roots []*Node, dst ...[]*Node) []*Node {
	var result []*Node
	if len(dst) > 0 && dst[0] != nil {
		result = dst[0][:0] // reuse backing array
	}

	for _, root := range roots {
		flattenNode(root, &result)
	}

	return result
}

func flattenNode(n *Node, out *[]*Node) {
	if !n.Visible {
		return
	}

	*out = append(*out, n)

	if !n.Expanded {
		return
	}

	for _, child := range n.Children {
		flattenNode(child, out)
	}
}

// HasChildren returns true if the node has child nodes.
func (n *Node) HasChildren() bool {
	return len(n.Children) > 0
}

// Toggle flips the expand/collapse state.
func (n *Node) Toggle() {
	if n.HasChildren() {
		n.Expanded = !n.Expanded
	}
}

// Breadcrumb returns the path from root to this node as a slice of names.
func (n *Node) Breadcrumb() []string {
	// Count depth first to allocate once.
	depth := 0
	for cur := n; cur != nil; cur = cur.Parent {
		depth++
	}

	path := make([]string, depth)
	i := depth - 1

	for cur := n; cur != nil; cur = cur.Parent {
		path[i] = cur.Name
		i--
	}

	return path
}

// ExpandAll recursively expands the node and all descendants.
func (n *Node) ExpandAll() {
	n.Expanded = true
	for _, c := range n.Children {
		c.ExpandAll()
	}
}

// CollapseAll recursively collapses the node and all descendants.
func (n *Node) CollapseAll() {
	n.Expanded = false
	for _, c := range n.Children {
		c.CollapseAll()
	}
}
