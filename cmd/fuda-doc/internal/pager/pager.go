package pager

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/arloliu/fuda/cmd/fuda-doc/internal/colors"
)

var (
	helpStyle   = colors.PagerHelpStyle
	statusStyle = colors.PagerStatusStyle
)

// searchMode describes the current search UI state.
type searchMode int

const (
	searchOff    searchMode = iota // no search active
	searchInput                    // user is typing a query
	searchActive                   // matches computed, navigating
)

// Model is the bubbletea model for the pager viewport.
type Model struct {
	viewport viewport.Model
	content  string // original (unhighlighted) content
	title    string
	ready    bool

	// search
	mode        searchMode
	searchBuf   string // text the user is typing
	search      searchState
	highlighted string // content with search highlights applied
}

// New creates a new pager Model with the given content and title.
func New(content, title string) Model {
	return Model{
		content: content,
		title:   title,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		headerHeight := 0
		footerHeight := 1

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-headerHeight-footerHeight)
			m.viewport.SetContent(m.content)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - headerHeight - footerHeight
		}

	case tea.KeyMsg:
		switch m.mode {
		case searchInput:
			return m.updateSearchInput(msg)
		case searchActive:
			return m.updateSearchActive(msg)
		case searchOff:
			return m.updateNormal(msg)
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)

	return m, cmd
}

// updateNormal handles keys when no search is active.
func (m Model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "/":
		m.mode = searchInput
		m.searchBuf = ""

		return m, nil
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)

	return m, cmd
}

// updateSearchInput handles keys while the user is typing a search query.
func (m Model) updateSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.searchBuf == "" {
			m.mode = searchOff
			return m, nil
		}

		m.search.query = m.searchBuf
		m.search.active = true
		m.search.buildMatches(m.content)

		if len(m.search.matches) == 0 {
			// No matches — stay in active mode so the user sees "0/0"
			m.mode = searchActive
			m.viewport.SetContent(m.content)

			return m, nil
		}

		m.mode = searchActive
		m.applyHighlightsAndScroll()

		return m, nil

	case "esc":
		m.mode = searchOff
		m.searchBuf = ""

		return m, nil

	case "backspace":
		if len(m.searchBuf) > 0 {
			m.searchBuf = m.searchBuf[:len(m.searchBuf)-1]
		}

		return m, nil

	default:
		// Accept printable characters.
		if len(msg.String()) == 1 || msg.Type == tea.KeyRunes {
			m.searchBuf += msg.String()
		}

		return m, nil
	}
}

// updateSearchActive handles keys while navigating search results.
func (m Model) updateSearchActive(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "esc":
		m.clearSearch()
		return m, nil

	case "/":
		m.mode = searchInput
		m.searchBuf = m.search.query

		return m, nil

	case "n":
		m.search.nextMatch()
		m.applyHighlightsAndScroll()

		return m, nil

	case "p":
		m.search.prevMatch()
		m.applyHighlightsAndScroll()

		return m, nil
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)

	return m, cmd
}

// applyHighlightsAndScroll refreshes the viewport content with highlights and
// scrolls so the current match is visible.
func (m *Model) applyHighlightsAndScroll() {
	m.highlighted = m.search.highlightContent(m.content)
	m.viewport.SetContent(m.highlighted)

	if line := m.search.currentLine(); line >= 0 {
		// Place the match roughly 1/3 from the top of the viewport.
		target := line - m.viewport.Height/3 //nolint:mnd // visual offset
		if target < 0 {
			target = 0
		}

		m.viewport.SetYOffset(target)
	}
}

// clearSearch resets all search state and restores original content.
func (m *Model) clearSearch() {
	m.mode = searchOff
	m.searchBuf = ""
	m.search = searchState{}
	m.highlighted = ""
	m.viewport.SetContent(m.content)
}

// View implements tea.Model.
func (m Model) View() string {
	if !m.ready {
		return "\n  Loading..."
	}

	return fmt.Sprintf("%s\n%s", m.viewport.View(), m.statusBar())
}

func (m Model) statusBar() string {
	pct := fmt.Sprintf(" %3.f%% ", m.viewport.ScrollPercent()*100) //nolint:mnd // percentage multiplier
	info := statusStyle.Render(pct)

	var mid string

	switch m.mode {
	case searchInput:
		prompt := colors.SearchPromptStyle.Render("/")
		input := colors.SearchInputStyle.Render(m.searchBuf + "█")
		mid = " " + prompt + input

	case searchActive:
		total := len(m.search.matches)
		cur := 0

		if total > 0 {
			cur = m.search.current + 1
		}

		matchInfo := statusStyle.Render(fmt.Sprintf(" [%d/%d] ", cur, total))
		query := colors.SearchPromptStyle.Render("/" + m.search.query)
		mid = " " + query + " " + matchInfo

	case searchOff:
		// no search
	}

	help := m.helpText()
	helpRendered := helpStyle.Render(help)

	gap := strings.Repeat(" ",
		max(0, m.viewport.Width-lipgloss.Width(info)-lipgloss.Width(mid)-lipgloss.Width(helpRendered)))

	return lipgloss.JoinHorizontal(lipgloss.Top, info, mid, gap, helpRendered)
}

func (m Model) helpText() string {
	switch m.mode {
	case searchInput:
		return " enter confirm • esc cancel"
	case searchActive:
		return " n/p next/prev • / new search • esc clear"
	case searchOff:
		return " ↑/↓ scroll • pgup/pgdn page • / search • q quit"
	}

	return ""
}

// Run launches the pager with the given content. Blocks until the user quits.
func Run(content, title string) error {
	m := New(content, title)

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("pager error: %w", err)
	}

	return nil
}
