package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all keybindings for the TUI.
type KeyMap struct {
	Up          key.Binding
	Down        key.Binding
	Left        key.Binding
	Right       key.Binding
	Toggle      key.Binding
	Tab         key.Binding
	Search      key.Binding
	SearchEnter key.Binding
	SearchEsc   key.Binding
	Backspace   key.Binding
	ExpandAll   key.Binding
	CollapseAll key.Binding
	CopyPath    key.Binding
	Help        key.Binding
	Filter      key.Binding
	Save        key.Binding
	Quit        key.Binding
}

// DefaultKeyMap returns the default keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "collapse/tree"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "expand/detail"),
		),
		Toggle: key.NewBinding(
			key.WithKeys(" ", "enter"),
			key.WithHelp("space", "toggle"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch panel"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		SearchEnter: key.NewBinding(
			key.WithKeys("enter"),
		),
		SearchEsc: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "clear/cancel"),
		),
		Backspace: key.NewBinding(
			key.WithKeys("backspace"),
		),
		ExpandAll: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "expand all"),
		),
		CollapseAll: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "collapse all"),
		),
		CopyPath: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "copy YAML path"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Filter: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "filter by tag"),
		),
		Save: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "save/export"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}
