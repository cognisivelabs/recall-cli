package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Muted, professional palette
	colorSubtle    = lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#626262"}
	colorHighlight = lipgloss.AdaptiveColor{Light: "#4A4A4A", Dark: "#E2E2E2"}
	colorAccent    = lipgloss.AdaptiveColor{Light: "#5C7C9A", Dark: "#7EB8DA"}
	colorMuted     = lipgloss.AdaptiveColor{Light: "#B0B0B0", Dark: "#4A4A4A"}
	colorWarn      = lipgloss.AdaptiveColor{Light: "#B38B5D", Dark: "#D4A76A"}
	colorSuccess   = lipgloss.AdaptiveColor{Light: "#5D8C6F", Dark: "#7FB895"}
	colorBorder    = lipgloss.AdaptiveColor{Light: "#D0D0D0", Dark: "#3A3A3A"}
	colorActiveBdr = lipgloss.AdaptiveColor{Light: "#8BADC4", Dark: "#5A8BA8"}
	colorDimText   = lipgloss.AdaptiveColor{Light: "#AAAAAA", Dark: "#555555"}
	colorTag       = lipgloss.AdaptiveColor{Light: "#6B8E8A", Dark: "#6B9E97"}
	colorSource    = lipgloss.AdaptiveColor{Light: "#8B7EB8", Dark: "#9B8EC8"}

	// Pane borders
	listPaneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorActiveBdr).
			Padding(0, 1)

	detailPaneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(1, 2)

	// Detail content styles
	detailTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorHighlight).
			MarginBottom(1)

	detailSection = lipgloss.NewStyle().
			Foreground(colorSubtle).
			Bold(true)

	detailBody = lipgloss.NewStyle().
			Foreground(colorHighlight)

	detailCode = lipgloss.NewStyle().
			Foreground(colorAccent).
			Background(lipgloss.AdaptiveColor{Light: "#F0F0F0", Dark: "#1E1E1E"}).
			Padding(0, 1)

	tagStyle = lipgloss.NewStyle().
			Foreground(colorTag).
			Background(lipgloss.AdaptiveColor{Light: "#EAF2F0", Dark: "#1A2A27"}).
			Padding(0, 1).
			MarginRight(1)

	sourceStyle = lipgloss.NewStyle().
			Foreground(colorSource).
			Italic(true)

	workspaceBadge = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	// Status bar
	statusBar = lipgloss.NewStyle().
			Foreground(colorSubtle).
			Padding(0, 1)

	statusKey = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	statusDesc = lipgloss.NewStyle().
			Foreground(colorSubtle)

	// Header
	headerStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true).
			Padding(0, 1)

	headerDim = lipgloss.NewStyle().
			Foreground(colorSubtle).
			Italic(true)

	// Empty state
	emptyStyle = lipgloss.NewStyle().
			Foreground(colorDimText).
			Italic(true).
			Padding(2, 2)
)
