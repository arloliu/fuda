// Package colors provides a shared color theme for the fuda-doc CLI using
// lipgloss CompleteAdaptiveColor. Each color adapts to the terminal's light or
// dark background and degrades gracefully across TrueColor, ANSI256, and basic
// ANSI color profiles.
package colors

import "github.com/charmbracelet/lipgloss"

// Theme colors — a modern, low-contrast palette inspired by Catppuccin /
// GitHub Primer.  Every color is a lipgloss.CompleteAdaptiveColor so it works
// correctly on light and dark backgrounds and across all terminal color
// profiles (TrueColor, ANSI 256, ANSI 16).
//
// Design principles:
//   - Soft, desaturated tones — easy on the eyes for long reading sessions.
//   - No purple / magenta — replaced by warm slate and teal accents.
//   - Low overall contrast — text layers feel cohesive rather than flashy.

// Teal is the primary accent, used for headers, YAML keys, and sub-section titles.
var Teal = lipgloss.CompleteAdaptiveColor{
	Light: lipgloss.CompleteColor{TrueColor: "#0d9488", ANSI256: "30", ANSI: "6"},
	Dark:  lipgloss.CompleteColor{TrueColor: "#5eead4", ANSI256: "79", ANSI: "14"},
}

// BrightTeal is a more vivid/saturated teal used for focused root items.
var BrightTeal = lipgloss.CompleteAdaptiveColor{
	Light: lipgloss.CompleteColor{TrueColor: "#047857", ANSI256: "29", ANSI: "6"},
	Dark:  lipgloss.CompleteColor{TrueColor: "#99f6e4", ANSI256: "122", ANSI: "14"},
}

// Sage is used for field names — a calm, muted green.
var Sage = lipgloss.CompleteAdaptiveColor{
	Light: lipgloss.CompleteColor{TrueColor: "#4d7c6f", ANSI256: "65", ANSI: "2"},
	Dark:  lipgloss.CompleteColor{TrueColor: "#86d9b4", ANSI256: "115", ANSI: "10"},
}

// BrightSage is a more vivid sage used for focused nested struct items.
var BrightSage = lipgloss.CompleteAdaptiveColor{
	Light: lipgloss.CompleteColor{TrueColor: "#2d6a4f", ANSI256: "29", ANSI: "2"},
	Dark:  lipgloss.CompleteColor{TrueColor: "#b5f0d3", ANSI256: "157", ANSI: "10"},
}

// Sand is a warm neutral for labels and ordered-list markers.
var Sand = lipgloss.CompleteAdaptiveColor{
	Light: lipgloss.CompleteColor{TrueColor: "#92713a", ANSI256: "137", ANSI: "3"},
	Dark:  lipgloss.CompleteColor{TrueColor: "#d4b97d", ANSI256: "180", ANSI: "11"},
}

// Sky is a gentle blue for type annotations.
var Sky = lipgloss.CompleteAdaptiveColor{
	Light: lipgloss.CompleteColor{TrueColor: "#3b82a0", ANSI256: "67", ANSI: "4"},
	Dark:  lipgloss.CompleteColor{TrueColor: "#7ec8e3", ANSI256: "110", ANSI: "12"},
}

// Slate is used for section title boxes — a cool blue-gray instead of purple.
var Slate = lipgloss.CompleteAdaptiveColor{
	Light: lipgloss.CompleteColor{TrueColor: "#475569", ANSI256: "60", ANSI: "8"},
	Dark:  lipgloss.CompleteColor{TrueColor: "#94a3b8", ANSI256: "146", ANSI: "7"},
}

// Text is the primary foreground for values and normal emphasis.
var Text = lipgloss.CompleteAdaptiveColor{
	Light: lipgloss.CompleteColor{TrueColor: "#334155", ANSI256: "239", ANSI: "0"},
	Dark:  lipgloss.CompleteColor{TrueColor: "#e2e8f0", ANSI256: "253", ANSI: "15"},
}

// Muted is used for dim/secondary text (comments, dividers, help text).
var Muted = lipgloss.CompleteAdaptiveColor{
	Light: lipgloss.CompleteColor{TrueColor: "#78909c", ANSI256: "245", ANSI: "8"},
	Dark:  lipgloss.CompleteColor{TrueColor: "#8b9dab", ANSI256: "247", ANSI: "7"},
}

// SubtleBg is used for status bar backgrounds in the pager.
var SubtleBg = lipgloss.CompleteAdaptiveColor{
	Light: lipgloss.CompleteColor{TrueColor: "#e8edf2", ANSI256: "254", ANSI: "7"},
	Dark:  lipgloss.CompleteColor{TrueColor: "#2d3748", ANSI256: "237", ANSI: "8"},
}

// HighlightBg is used for the selected/cursor row background in the TUI tree.
var HighlightBg = lipgloss.CompleteAdaptiveColor{
	Light: lipgloss.CompleteColor{TrueColor: "#d1e5df", ANSI256: "254", ANSI: "7"},
	Dark:  lipgloss.CompleteColor{TrueColor: "#1e3a4a", ANSI256: "236", ANSI: "8"},
}

// HighlightFg is used for the focused cursor item's text foreground.
var HighlightFg = lipgloss.CompleteAdaptiveColor{
	Light: lipgloss.CompleteColor{TrueColor: "#065f46", ANSI256: "23", ANSI: "6"},
	Dark:  lipgloss.CompleteColor{TrueColor: "#f0fdfa", ANSI256: "231", ANSI: "15"},
}

// SearchHighlightFg is the foreground color for search match highlights.
var SearchHighlightFg = lipgloss.CompleteAdaptiveColor{
	Light: lipgloss.CompleteColor{TrueColor: "#1a1a2e", ANSI256: "0", ANSI: "0"},
	Dark:  lipgloss.CompleteColor{TrueColor: "#1a1a2e", ANSI256: "0", ANSI: "0"},
}

// SearchHighlightBg is the background color for search match highlights.
var SearchHighlightBg = lipgloss.CompleteAdaptiveColor{
	Light: lipgloss.CompleteColor{TrueColor: "#fbbf24", ANSI256: "220", ANSI: "11"},
	Dark:  lipgloss.CompleteColor{TrueColor: "#fbbf24", ANSI256: "220", ANSI: "11"},
}

// SearchCurrentFg is the foreground color for the current/active search match.
var SearchCurrentFg = lipgloss.CompleteAdaptiveColor{
	Light: lipgloss.CompleteColor{TrueColor: "#1a1a2e", ANSI256: "0", ANSI: "0"},
	Dark:  lipgloss.CompleteColor{TrueColor: "#1a1a2e", ANSI256: "0", ANSI: "0"},
}

// SearchCurrentBg is the background color for the current/active search match.
var SearchCurrentBg = lipgloss.CompleteAdaptiveColor{
	Light: lipgloss.CompleteColor{TrueColor: "#f97316", ANSI256: "208", ANSI: "3"},
	Dark:  lipgloss.CompleteColor{TrueColor: "#fb923c", ANSI256: "208", ANSI: "3"},
}

// --- YAML syntax highlight colors ------------------------------------------

// YAMLKey is used for YAML mapping keys — a soft lavender-blue, easy to scan.
var YAMLKey = lipgloss.CompleteAdaptiveColor{
	Light: lipgloss.CompleteColor{TrueColor: "#6366f1", ANSI256: "62", ANSI: "4"},
	Dark:  lipgloss.CompleteColor{TrueColor: "#a5b4fc", ANSI256: "141", ANSI: "12"},
}

// YAMLString is used for YAML string values — a warm sand/amber tone.
var YAMLString = lipgloss.CompleteAdaptiveColor{
	Light: lipgloss.CompleteColor{TrueColor: "#92713a", ANSI256: "137", ANSI: "3"},
	Dark:  lipgloss.CompleteColor{TrueColor: "#e2c08d", ANSI256: "186", ANSI: "11"},
}

// YAMLPunct is used for YAML punctuation (colons, dashes) — dim/muted.
var YAMLPunct = lipgloss.CompleteAdaptiveColor{
	Light: lipgloss.CompleteColor{TrueColor: "#94a3b8", ANSI256: "247", ANSI: "8"},
	Dark:  lipgloss.CompleteColor{TrueColor: "#64748b", ANSI256: "244", ANSI: "8"},
}

// ---------------------------------------------------------------------------
// Pre-built lipgloss styles
// ---------------------------------------------------------------------------

// Bold creates a bold style with the given foreground color.
func Bold(fg lipgloss.TerminalColor) lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(fg)
}

// Styles commonly used across the ASCII printer and pager.
var (
	HeaderStyle     = lipgloss.NewStyle().Bold(true).Foreground(Teal)
	SectionStyle    = lipgloss.NewStyle().Bold(true).Foreground(Slate)
	SubsectionStyle = lipgloss.NewStyle().Bold(true).Foreground(Teal)
	FieldStyle      = lipgloss.NewStyle().Bold(true).Foreground(Sage)
	TypeStyle       = lipgloss.NewStyle().Foreground(Sky)
	LabelStyle      = lipgloss.NewStyle().Foreground(Sand)
	ValueStyle      = lipgloss.NewStyle().Foreground(Text)
	MutedStyle      = lipgloss.NewStyle().Foreground(Muted)
	DimStyle        = lipgloss.NewStyle().Faint(true)
	CyanStyle       = lipgloss.NewStyle().Foreground(Teal)

	PagerStatusStyle = lipgloss.NewStyle().
				Foreground(Muted).
				Background(SubtleBg).
				PaddingLeft(1).
				PaddingRight(1)
	PagerHelpStyle = lipgloss.NewStyle().Foreground(Muted)

	SearchHighlightStyle = lipgloss.NewStyle().
				Foreground(SearchHighlightFg).
				Background(SearchHighlightBg).
				Bold(true)
	SearchCurrentStyle = lipgloss.NewStyle().
				Foreground(SearchCurrentFg).
				Background(SearchCurrentBg).
				Bold(true)
	SearchPromptStyle = lipgloss.NewStyle().
				Foreground(Sand).
				Bold(true)
	SearchInputStyle = lipgloss.NewStyle().
				Foreground(Text)

	// TUI-specific styles
	TUITitleStyle  = lipgloss.NewStyle().Bold(true).Foreground(Teal).PaddingLeft(1).PaddingRight(1)
	TUITreeLine    = lipgloss.NewStyle().Foreground(Muted)
	TUITreeCursor  = lipgloss.NewStyle().Bold(true).Foreground(Teal)
	TUITreeNormal  = lipgloss.NewStyle().Foreground(Text)
	TUITreeNested  = lipgloss.NewStyle().Bold(true).Foreground(Sage)
	TUIHelpStyle   = lipgloss.NewStyle().Foreground(Muted)
	TUIHighlightBg = HighlightBg
	TUIHighlightFg = HighlightFg

	// YAML syntax styles
	YAMLKeyStyle   = lipgloss.NewStyle().Foreground(YAMLKey)
	YAMLValueStyle = lipgloss.NewStyle().Foreground(YAMLString)
	YAMLPunctStyle = lipgloss.NewStyle().Foreground(YAMLPunct)
)
