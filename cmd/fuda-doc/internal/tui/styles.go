package tui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/arloliu/fuda/cmd/fuda-doc/internal/colors"
)

// panel is the border/layout style builder for each panel.
type panelStyle struct {
	active   lipgloss.Style
	inactive lipgloss.Style
}

var panels = panelStyle{
	active: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colors.Teal),
	inactive: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colors.Muted),
}

// tree rendering styles
var (
	treeLine        = colors.TUITreeLine
	treeCursor      = colors.TUITreeCursor
	treeNormal      = colors.TUITreeNormal
	treeNested      = colors.TUITreeNested
	treeRoot        = lipgloss.NewStyle().Bold(true).Foreground(colors.Teal)
	treeRootHL      = lipgloss.NewStyle().Bold(true).Foreground(colors.BrightTeal)
	treeNestedHL    = lipgloss.NewStyle().Bold(true).Foreground(colors.BrightSage)
	treeHighlight   = colors.TUIHighlightBg
	treeHighlightFg = colors.TUIHighlightFg
	treeMuted       = colors.MutedStyle
	treeCollapsed   = lipgloss.NewStyle().Foreground(colors.Sand)
)

// detail panel styles
var (
	detailTitle = lipgloss.NewStyle().Bold(true).Foreground(colors.Teal).MarginBottom(1)
	detailLabel = colors.LabelStyle
	detailValue = colors.ValueStyle
	detailType  = colors.TypeStyle
	detailDesc  = lipgloss.NewStyle().Foreground(colors.Text)
	detailMuted = colors.MutedStyle
)

// status bar / help
var (
	helpStyle  = colors.TUIHelpStyle
	titleStyle = colors.TUITitleStyle
)

// search
var (
	searchPrompt = colors.SearchPromptStyle
	searchInput  = colors.SearchInputStyle
)
