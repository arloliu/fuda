package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/arloliu/fuda/cmd/fuda-doc/internal/docgen"
	"github.com/arloliu/fuda/cmd/fuda-doc/internal/docutil"
)

// panel identifies which panel has focus.
type panel int

const (
	panelTree panel = iota
	panelDetail
	panelYAML
	panelCount // sentinel for wrapping
)

// Model is the top-level bubbletea model for the TUI explorer.
type Model struct {
	docs   []docgen.StructDoc
	tree   treeModel
	detail detailModel
	yaml   yamlModel
	search searchModel
	keys   KeyMap

	focus  panel
	width  int
	height int
	ready  bool

	// Overlay state
	showHelp bool

	// Tag filter state
	filterActive bool
	filterTags   []string // available tags
	filterCursor int      // selection in the tag list
	activeFilter string   // currently active tag filter ("" = none)

	// Export picker state
	exportActive bool
	exportItems  []exportItem
	exportCursor int

	// Flash message (for copy confirmation, save, etc.)
	flash    string
	flashEnd time.Time
}

// New creates a new TUI Model.
func New(docs []docgen.StructDoc) Model {
	roots := BuildTree(docs)

	m := Model{
		docs:   docs,
		tree:   newTreeModel(roots),
		detail: newDetailModel(),
		yaml:   newYAMLModel(),
		search: newSearchModel(),
		keys:   DefaultKeyMap(),
		focus:  panelTree,
	}

	return m
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.recalcLayout()
		m.refreshPanels()

	case tea.MouseMsg:
		return m.handleMouse(msg)

	case tea.KeyMsg:
		if m.search.active {
			return m.handleSearchKey(msg)
		}

		return m.handleKey(msg)
	}

	return m, nil
}

// handleKey processes keys in normal (non-search) mode.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If help overlay is showing, any key dismisses it.
	if m.showHelp {
		m.showHelp = false

		return m, nil
	}

	// If tag filter picker is active, handle it separately.
	if m.filterActive {
		return m.handleFilterKey(msg)
	}

	// If export picker is active, handle it separately.
	if m.exportActive {
		return m.handleExportKey(msg)
	}

	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Search):
		m.search.start()

		return m, nil

	case key.Matches(msg, m.keys.SearchEsc):
		m.handleEsc()

		return m, nil

	case key.Matches(msg, m.keys.Tab):
		m.focus = (m.focus + 1) % panelCount
		m.tree.focused = (m.focus == panelTree)

		return m, nil

	case key.Matches(msg, m.keys.ExpandAll):
		if m.focus == panelTree {
			m.tree.expandAll()
		}

		return m, nil

	case key.Matches(msg, m.keys.CollapseAll):
		if m.focus == panelTree {
			m.tree.collapseAll()
		}

		return m, nil

	case key.Matches(msg, m.keys.Up):
		m.scrollFocused(-1)

		return m, nil

	case key.Matches(msg, m.keys.Down):
		m.scrollFocused(1)

		return m, nil

	case key.Matches(msg, m.keys.Left):
		m.handleNavLeft()

		return m, nil

	case key.Matches(msg, m.keys.Right):
		m.handleNavRight()

		return m, nil

	case key.Matches(msg, m.keys.Toggle):
		if m.focus == panelTree {
			m.tree.toggle()
			m.refreshPanels()
		}

		return m, nil

	case key.Matches(msg, m.keys.CopyPath):
		m.copyYAMLPath()

		return m, nil

	case key.Matches(msg, m.keys.Help):
		m.showHelp = true

		return m, nil

	case key.Matches(msg, m.keys.Filter):
		m.openFilter()

		return m, nil

	case key.Matches(msg, m.keys.Save):
		m.openExport()

		return m, nil
	}

	return m, nil
}

// handleEsc processes the Escape key — clears active filter, then search.
func (m *Model) handleEsc() {
	if m.activeFilter != "" {
		m.activeFilter = ""
		showAll(m.tree.roots)
		m.tree.reindex()
		m.refreshPanels()

		return
	}

	if m.search.hasQuery() {
		m.search.clear()
		showAll(m.tree.roots)
		m.tree.reindex()
		m.refreshPanels()
	}
}

// scrollFocused scrolls the currently focused panel in the given direction
// (negative = up, positive = down).
func (m *Model) scrollFocused(dir int) {
	switch m.focus {
	case panelTree:
		if dir < 0 {
			m.tree.moveUp()
		} else {
			m.tree.moveDown()
		}

		m.refreshPanels()
	case panelDetail:
		if dir < 0 {
			m.detail.scrollUp()
		} else {
			m.detail.scrollDown()
		}
	case panelYAML:
		if dir < 0 {
			m.yaml.scrollUp()
		} else {
			m.yaml.scrollDown()
		}
	case panelCount:
		// sentinel — not a real panel
	}
}

func (m *Model) handleNavLeft() {
	switch m.focus {
	case panelTree:
		m.tree.collapse()
		m.refreshPanels()
	case panelDetail, panelYAML:
		// Switch focus left.
		m.focus--
		if m.focus < 0 {
			m.focus = 0
		}

		m.tree.focused = (m.focus == panelTree)
	case panelCount:
		// sentinel — not a real panel
	}
}

func (m *Model) handleNavRight() {
	switch m.focus {
	case panelTree:
		n := m.tree.selected()
		if n != nil && n.HasChildren() && !n.Expanded {
			m.tree.expand()
			m.refreshPanels()
		} else {
			// Move focus to detail.
			m.focus = panelDetail
			m.tree.focused = false
		}
	case panelDetail:
		m.focus = panelYAML
	case panelYAML, panelCount:
		// nothing to do
	}
}

// handleSearchKey processes keys while the search input is active.
func (m Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.SearchEnter):
		q := m.search.confirm()
		if q == "" {
			showAll(m.tree.roots)
		} else {
			m.search.applyFilter(m.tree.roots)
		}

		m.tree.reindex()
		m.refreshPanels()

		return m, nil

	case key.Matches(msg, m.keys.SearchEsc):
		m.search.cancel()

		return m, nil

	case key.Matches(msg, m.keys.Backspace):
		m.search.backspace()

		return m, nil

	default:
		s := msg.String()
		if len(s) == 1 || msg.Type == tea.KeyRunes {
			m.search.addChar(s)
		}

		return m, nil
	}
}

// handleMouse processes mouse events.
func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Auto-focus the panel under the mouse pointer on any mouse event.
	m.autoFocusPanel(msg.X, msg.Y)

	switch msg.Button { //nolint:exhaustive // only handle wheel and left click
	case tea.MouseButtonWheelUp:
		m.scrollFocused(-1)

	case tea.MouseButtonWheelDown:
		m.scrollFocused(1)

	case tea.MouseButtonLeft:
		if msg.Action != tea.MouseActionPress {
			break
		}

		m.handleClick(msg.X, msg.Y)
	}

	return m, nil
}

// autoFocusPanel determines which panel the mouse pointer is over and
// switches focus accordingly. This enables mouse-wheel scrolling on any
// panel by simply hovering over it.
func (m *Model) autoFocusPanel(x, y int) {
	treeWidth := m.treePanelWidth()
	detailH := m.rightPanelHeight()

	// Account for border (1 char top).
	contentY := y - 1

	if x < treeWidth {
		m.focus = panelTree
		m.tree.focused = true
	} else if contentY < detailH {
		m.focus = panelDetail
		m.tree.focused = false
	} else {
		m.focus = panelYAML
		m.tree.focused = false
	}
}

// handleClick determines which panel was clicked and acts accordingly.
func (m *Model) handleClick(x, y int) {
	treeWidth := m.treePanelWidth()
	rightWidth := m.width - treeWidth
	detailH := m.rightPanelHeight()

	// Account for borders (1 char each side).
	contentY := y - 1 // subtract top border

	if x < treeWidth {
		// Clicked in tree panel.
		m.focus = panelTree
		m.tree.focused = true

		// Subtract top padding so row 0 maps to the first node.
		treeY := contentY - treePadTop
		contentX := x - 1 // subtract left border
		if treeY >= 0 && treeY < m.tree.height {
			if m.tree.clickAt(contentX, treeY) {
				m.refreshPanels()
			}
		}
	} else if x < treeWidth+rightWidth {
		innerY := contentY

		if innerY < detailH {
			m.focus = panelDetail
			m.tree.focused = false
		} else {
			m.focus = panelYAML
			m.tree.focused = false
		}
	}
}

// refreshPanels updates the detail and YAML panels if the cursor changed.
func (m *Model) refreshPanels() {
	n := m.tree.selected()
	m.detail.update(n)
	m.yaml.update(n)
}

// ---------------------------------------------------------------------------
// Copy YAML path (y)
// ---------------------------------------------------------------------------

// copyYAMLPath builds the dotted YAML path of the selected node and copies
// it to the clipboard using OSC 52 escape sequence.
func (m *Model) copyYAMLPath() {
	n := m.tree.selected()
	if n == nil || n.IsRoot {
		return
	}

	// Build YAML path by walking from the selected node to the root.
	// Count depth first, then fill from end to avoid prepend allocations.
	depth := 0
	for cur := n; cur != nil && !cur.IsRoot; cur = cur.Parent {
		if cur.Field != nil {
			depth++
		}
	}

	if depth == 0 {
		return
	}

	parts := make([]string, depth)
	i := depth - 1

	for cur := n; cur != nil && !cur.IsRoot; cur = cur.Parent {
		if cur.Field != nil {
			parts[i] = docutil.YAMLKey(cur.Field)
			i--
		}
	}

	path := strings.Join(parts, ".")
	m.setFlash("Copied: "+path, flashDurationInfo)
}

// ---------------------------------------------------------------------------
// Help overlay (?)
// ---------------------------------------------------------------------------

func (m Model) helpOverlay() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#5eead4")).
		Render("Keyboard Shortcuts")

	rows := []struct{ key, desc string }{
		{"↑/k ↓/j", "Navigate up/down"},
		{"←/h →/l", "Collapse/expand or switch panel"},
		{"Space/Enter", "Toggle expand/collapse"},
		{"Tab", "Cycle panel focus"},
		{"e / w", "Expand / collapse all"},
		{"/ (slash)", "Search fields"},
		{"Esc", "Clear search or filter"},
		{"y", "Copy YAML path of selected field"},
		{"f", "Filter by tag"},
		{"s", "Export (Markdown / YAML / .env)"},
		{"?", "Show/hide this help"},
		{"q / Ctrl+C", "Quit"},
		{"", ""},
		{"Mouse wheel", "Scroll panel under pointer"},
		{"Left click", "Select tree item"},
	}

	var sb strings.Builder

	sb.WriteString(title)
	sb.WriteString("\n\n")

	for _, r := range rows {
		if r.key == "" {
			sb.WriteByte('\n')

			continue
		}

		k := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#d4b97d")).
			Width(16). //nolint:mnd // column width
			Render(r.key)
		sb.WriteString("  " + k + r.desc + "\n")
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#5eead4")).
		Padding(1, 2).
		Render(sb.String())

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// ---------------------------------------------------------------------------
// Tag filter (f)
// ---------------------------------------------------------------------------

const clearFilterLabel = "(clear filter)"

// allFilterTags is the pre-built filter list (clear sentinel + known tags).
// Built once to avoid allocation on every openFilter call.
var allFilterTags = append([]string{clearFilterLabel},
	"env", "default", "validate", "ref", "refFrom", "dsn", "required",
)

func (m *Model) openFilter() {
	m.filterActive = true
	m.filterTags = allFilterTags
	m.filterCursor = 0
}

func (m Model) handleFilterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.SearchEsc):
		m.filterActive = false

		return m, nil

	case key.Matches(msg, m.keys.Up):
		if m.filterCursor > 0 {
			m.filterCursor--
		}

		return m, nil

	case key.Matches(msg, m.keys.Down):
		if m.filterCursor < len(m.filterTags)-1 {
			m.filterCursor++
		}

		return m, nil

	case key.Matches(msg, m.keys.Toggle):
		tag := m.filterTags[m.filterCursor]
		m.filterActive = false

		if tag == clearFilterLabel {
			m.activeFilter = ""
			showAll(m.tree.roots)
		} else {
			m.activeFilter = tag
			applyTagFilter(m.tree.roots, tag)
		}

		m.tree.reindex()
		m.refreshPanels()

		return m, nil
	}

	return m, nil
}

// applyTagFilter marks nodes visible only if they (or descendants) have the
// given tag.
func applyTagFilter(roots []*Node, tag string) {
	for _, root := range roots {
		filterNodeByTag(root, tag)
	}
}

func filterNodeByTag(n *Node, tag string) bool {
	selfMatch := false

	if n.Field != nil {
		if _, ok := n.Field.Tags[tag]; ok {
			selfMatch = true
		}
	}

	childMatch := false

	for _, child := range n.Children {
		if filterNodeByTag(child, tag) {
			childMatch = true
		}
	}

	matched := selfMatch || childMatch || n.IsRoot
	n.Visible = matched

	if childMatch {
		n.Expanded = true
	}

	return selfMatch || childMatch
}

func (m Model) filterOverlay() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#5eead4")).
		Render("Filter by tag")

	var sb strings.Builder

	sb.WriteString(title)
	sb.WriteString("\n\n")

	for i, tag := range m.filterTags {
		cursor := "  "
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("#e2e8f0"))

		if tag == clearFilterLabel {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("#8b9dab")).Italic(true)
		}

		if i == m.filterCursor {
			cursor = "▸ "
			style = style.Bold(true).Foreground(lipgloss.Color("#5eead4"))
		}

		sb.WriteString(cursor + style.Render(tag) + "\n")
	}

	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#8b9dab")).
		Render("space select • esc cancel"))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#5eead4")).
		Padding(1, 3). //nolint:mnd // visual padding
		Render(sb.String())

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// ---------------------------------------------------------------------------
// Save / export (s)
// ---------------------------------------------------------------------------

// exportItem describes one export format option.
type exportItem struct {
	label string
	ext   string
}

var exportFormats = []exportItem{
	{label: "Markdown documentation", ext: ".md"},
	{label: "Default YAML config", ext: ".yaml"},
	{label: ".env.example", ext: ".env.example"},
}

func (m *Model) openExport() {
	m.exportActive = true
	m.exportItems = exportFormats
	m.exportCursor = 0
}

func (m Model) handleExportKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.SearchEsc):
		m.exportActive = false

		return m, nil

	case key.Matches(msg, m.keys.Up):
		if m.exportCursor > 0 {
			m.exportCursor--
		}

		return m, nil

	case key.Matches(msg, m.keys.Down):
		if m.exportCursor < len(m.exportItems)-1 {
			m.exportCursor++
		}

		return m, nil

	case key.Matches(msg, m.keys.Toggle):
		m.exportActive = false
		m.doExport(m.exportItems[m.exportCursor])

		return m, nil
	}

	return m, nil
}

func (m *Model) doExport(item exportItem) {
	root := m.findSelectedRoot()
	if root == nil || root.StructDoc == nil {
		return
	}

	doc := root.StructDoc
	baseName := strings.ToLower(doc.Name)

	var filename string

	switch item.ext {
	case ".md":
		filename = baseName + ".md"
		m.exportMarkdown(filename, doc)
	case ".yaml":
		filename = baseName + ".yaml"
		m.exportYAML(filename, doc)
	case ".env.example":
		filename = baseName + ".env.example"
		m.exportEnvFile(filename, doc)
	}
}

func (m *Model) findSelectedRoot() *Node {
	n := m.tree.selected()
	if n == nil {
		return nil
	}

	root := n
	for root.Parent != nil {
		root = root.Parent
	}

	return root
}

const (
	flashDurationInfo  = 2 * time.Second
	flashDurationError = 3 * time.Second
)

// setFlash sets a temporary status message shown in the status bar.
func (m *Model) setFlash(msg string, d time.Duration) {
	m.flash = msg
	m.flashEnd = time.Now().Add(d)
}

func (m *Model) exportMarkdown(filename string, doc *docgen.StructDoc) {
	f, err := os.Create(filename)
	if err != nil {
		m.setFlash("Error: "+err.Error(), flashDurationError)

		return
	}

	printer := docgen.NewMarkdownPrinter(f)
	printer.Print(doc.Name, doc.Doc, doc.Fields)
	_ = f.Close()

	m.setFlash("Saved: "+filename, flashDurationInfo)
}

func (m *Model) exportYAML(filename string, doc *docgen.StructDoc) {
	f, err := os.Create(filename)
	if err != nil {
		m.setFlash("Error: "+err.Error(), flashDurationError)

		return
	}

	docs := []docgen.StructDoc{*doc}
	_ = docgen.PrintDefaultYAML(docs, f, true)
	_ = f.Close()

	m.setFlash("Saved: "+filename, flashDurationInfo)
}

func (m *Model) exportEnvFile(filename string, doc *docgen.StructDoc) {
	f, err := os.Create(filename)
	if err != nil {
		m.setFlash("Error: "+err.Error(), flashDurationError)

		return
	}

	docs := []docgen.StructDoc{*doc}
	_ = docgen.PrintEnvFile(docs, f)
	_ = f.Close()

	m.setFlash("Saved: "+filename, flashDurationInfo)
}

func (m Model) exportOverlay() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#5eead4")).
		Render("Export / Save")

	var sb strings.Builder

	sb.WriteString(title)
	sb.WriteString("\n\n")

	for i, item := range m.exportItems {
		cursor := "  "
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("#e2e8f0"))

		if i == m.exportCursor {
			cursor = "▸ "
			style = style.Bold(true).Foreground(lipgloss.Color("#5eead4"))
		}

		ext := lipgloss.NewStyle().Foreground(lipgloss.Color("#8b9dab")).Render(" (" + item.ext + ")")
		sb.WriteString(cursor + style.Render(item.label) + ext + "\n")
	}

	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#8b9dab")).
		Render("space select • esc cancel"))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#5eead4")).
		Padding(1, 3). //nolint:mnd // visual padding
		Render(sb.String())

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// ---------------------------------------------------------------------------
// Layout
// ---------------------------------------------------------------------------

const (
	treeWidthFraction = 0.35
	minTreeWidth      = 30
	statusBarHeight   = 1
	borderSize        = 2 // top + bottom border
)

func (m *Model) recalcLayout() {
	treeW := m.treePanelWidth()
	rightW := m.width - treeW

	contentH := m.height - statusBarHeight - borderSize
	if contentH < 1 {
		contentH = 1
	}

	m.tree.setSize(treeW-borderSize, contentH)
	m.tree.focused = (m.focus == panelTree)

	detailH := m.rightPanelHeight()
	yamlH := contentH - detailH

	m.detail.setSize(rightW-borderSize, detailH-borderSize)
	m.yaml.setSize(rightW-borderSize, yamlH-borderSize)
}

func (m *Model) treePanelWidth() int {
	tw := int(float64(m.width) * treeWidthFraction)
	if tw < minTreeWidth {
		tw = minTreeWidth
	}

	if tw > m.width/2 {
		tw = m.width / 2
	}

	return tw
}

func (m *Model) rightPanelHeight() int {
	contentH := m.height - statusBarHeight - borderSize
	// 60% for detail, 40% for YAML preview
	return max(3, contentH*6/10) //nolint:mnd // layout ratio
}

// ---------------------------------------------------------------------------
// View
// ---------------------------------------------------------------------------

// View implements tea.Model.
func (m Model) View() string {
	if !m.ready {
		return "\n  Loading..."
	}

	// Full-screen overlays.
	if m.showHelp {
		return m.helpOverlay()
	}

	if m.filterActive {
		return m.filterOverlay()
	}

	if m.exportActive {
		return m.exportOverlay()
	}

	treeW := m.treePanelWidth()
	rightW := m.width - treeW
	contentH := m.height - statusBarHeight - borderSize
	detailH := m.rightPanelHeight()
	yamlH := contentH - detailH

	// Tree panel
	treeStyle := m.borderStyle(panelTree).Width(treeW - borderSize).Height(contentH)
	treeView := treeStyle.Render(m.tree.view())

	// Detail panel
	detailStyle := m.borderStyle(panelDetail).Width(rightW - borderSize).Height(detailH - borderSize)
	detailView := detailStyle.Render(m.detail.view())

	// YAML panel
	yamlStyle := m.borderStyle(panelYAML).Width(rightW - borderSize).Height(yamlH - borderSize)
	yamlView := yamlStyle.Render(m.yaml.view())

	// Stack detail + yaml vertically
	rightColumn := lipgloss.JoinVertical(lipgloss.Left, detailView, yamlView)

	// Join tree + right horizontally
	main := lipgloss.JoinHorizontal(lipgloss.Top, treeView, rightColumn)

	// Status bar
	status := m.statusBar()

	return lipgloss.JoinVertical(lipgloss.Left, main, status)
}

func (m Model) borderStyle(p panel) lipgloss.Style {
	title := ""

	switch p {
	case panelTree:
		title = " Structs "
	case panelDetail:
		title = " Detail "

		if n := m.tree.selected(); n != nil {
			bc := n.Breadcrumb()
			title = " " + strings.Join(bc, " › ") + " "
		}
	case panelYAML:
		title = " YAML Preview "
	case panelCount:
		// sentinel — not a real panel
	}

	var style lipgloss.Style
	if p == m.focus {
		style = panels.active
	} else {
		style = panels.inactive
	}

	return style.BorderTop(true).BorderBottom(true).BorderLeft(true).BorderRight(true).
		UnsetWidth().UnsetHeight(). // reset so caller can set
		SetString(titleStyle.Render(title))
}

func (m Model) statusBar() string {
	// Flash message (e.g., "Copied: server.host").
	flashView := ""
	if m.flash != "" && time.Now().Before(m.flashEnd) {
		flashView = lipgloss.NewStyle().
			Bold(true).Foreground(lipgloss.Color("#86d9b4")).
			Render(" " + m.flash)
	}

	// Search bar.
	searchView := m.search.view()

	// Help text.
	help := m.helpText()

	// Compose.
	leftPart := searchView
	if flashView != "" {
		if leftPart != "" {
			leftPart += "  " + flashView
		} else {
			leftPart = flashView
		}
	}

	if leftPart != "" {
		gap := strings.Repeat(" ",
			max(0, m.width-lipgloss.Width(leftPart)-lipgloss.Width(help)))

		return lipgloss.JoinHorizontal(lipgloss.Top, leftPart, gap, help)
	}

	gap := strings.Repeat(" ", max(0, m.width-lipgloss.Width(help)))

	return gap + help
}

func (m Model) helpText() string {
	if m.search.active {
		return helpStyle.Render(" enter confirm • esc cancel")
	}

	parts := []string{
		"↑/↓ navigate",
		"space toggle",
		"tab panel",
		"/ search",
		"y copy",
		"f filter",
		"s save",
		"? help",
	}

	if m.search.hasQuery() {
		parts = append(parts, "esc clear search")
	}

	if m.activeFilter != "" {
		parts = append(parts, "esc clear ["+m.activeFilter+"]")
	}

	parts = append(parts, "q quit")

	return helpStyle.Render(" " + strings.Join(parts, " • "))
}

// Run launches the TUI. Blocks until the user quits.
func Run(docs []docgen.StructDoc) error {
	m := New(docs)

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui error: %w", err)
	}

	return nil
}
