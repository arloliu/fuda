package tui

import (
	"strings"

	"github.com/arloliu/fuda/cmd/fuda-doc/internal/docutil"
)

// searchModel handles search input and tree filtering.
type searchModel struct {
	active bool
	query  string
	buf    string // text being typed
}

func newSearchModel() searchModel {
	return searchModel{}
}

// start enters search input mode.
func (s *searchModel) start() {
	s.active = true
	s.buf = s.query // pre-fill with previous query
}

// cancel exits search input without applying.
func (s *searchModel) cancel() {
	s.active = false
	s.buf = ""
}

// confirm applies the current buffer as the search query.
func (s *searchModel) confirm() string {
	s.active = false
	s.query = s.buf

	return s.query
}

// clear resets the search entirely.
func (s *searchModel) clear() {
	s.active = false
	s.query = ""
	s.buf = ""
}

// backspace removes the last character from the input buffer.
func (s *searchModel) backspace() {
	if len(s.buf) > 0 {
		s.buf = s.buf[:len(s.buf)-1]
	}
}

// addChar appends a character to the input buffer.
func (s *searchModel) addChar(ch string) {
	s.buf += ch
}

// hasQuery returns true if there's an active search filter.
func (s *searchModel) hasQuery() bool {
	return s.query != ""
}

// applyFilter marks nodes as visible/invisible based on the current query.
// It ensures that ancestors of matching nodes remain visible and expanded.
func (s *searchModel) applyFilter(roots []*Node) {
	if s.query == "" {
		// Show everything.
		showAll(roots)

		return
	}

	lower := strings.ToLower(s.query)

	for _, root := range roots {
		filterNode(root, lower)
	}
}

// showAll recursively marks all nodes as visible.
func showAll(nodes []*Node) {
	for _, n := range nodes {
		n.Visible = true
		showAll(n.Children)
	}
}

// filterNode returns true if this node or any descendant matches the query.
func filterNode(n *Node, query string) bool {
	// Check if this node itself matches.
	selfMatch := nodeMatches(n, query)

	// Check children.
	childMatch := false

	for _, child := range n.Children {
		if filterNode(child, query) {
			childMatch = true
		}
	}

	matched := selfMatch || childMatch
	n.Visible = matched

	// Auto-expand nodes that have matching descendants.
	if childMatch {
		n.Expanded = true
	}

	return matched
}

// containsFold reports whether s contains substr using a case-insensitive match
// without allocating a lowered copy of s.
func containsFold(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), substr)
}

// nodeMatches checks whether a node matches the search query. It checks the
// field name, YAML key, env tag, type, and description.
// The query must already be lowercased.
func nodeMatches(n *Node, query string) bool {
	if containsFold(n.Name, query) {
		return true
	}

	if n.Field != nil {
		if containsFold(docutil.YAMLKey(n.Field), query) ||
			containsFold(n.Field.Type, query) ||
			containsFold(n.Field.Description, query) {
			return true
		}

		if v := n.Field.Tags["env"]; v != "" && containsFold(v, query) {
			return true
		}

		if v := n.Field.Tags["ref"]; v != "" && containsFold(v, query) {
			return true
		}
	}

	if n.StructDoc != nil && containsFold(n.StructDoc.Doc, query) {
		return true
	}

	return false
}

// view renders the search bar content.
func (s *searchModel) view() string {
	if s.active {
		return searchPrompt.Render("/") + searchInput.Render(s.buf+"█")
	}

	if s.hasQuery() {
		return searchPrompt.Render("/"+s.query) + " " + detailMuted.Render("(esc to clear)")
	}

	return ""
}
