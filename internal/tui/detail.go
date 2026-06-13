package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/CognisiveLabs/recall-cli/internal/storage"

	"github.com/charmbracelet/lipgloss"
)

type DetailModel struct {
	command *storage.Command
	width   int
	height  int
	matchCwd bool
}

// NewDetailModel creates an empty detail panel. Call SetCommand to populate it.
func NewDetailModel() DetailModel {
	return DetailModel{}
}

// SetCommand updates the command displayed in the detail panel. Pass nil to show
// the empty-state placeholder. matchCwd controls the workspace indicator.
func (m *DetailModel) SetCommand(cmd *storage.Command, matchCwd bool) {
	m.command = cmd
	m.matchCwd = matchCwd
}

func (m *DetailModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m DetailModel) View() string {
	if m.command == nil || m.command.Pattern == "" {
		return emptyStyle.Render("Select a command to view details")
	}

	c := m.command
	contentWidth := m.width - 6
	if contentWidth < 20 {
		contentWidth = 20
	}

	var sections []string

	// Title
	title := detailTitle.Width(contentWidth).Render(c.Pattern)
	sections = append(sections, title)

	// Description
	if c.Description != "" {
		desc := detailBody.Width(contentWidth).Render(c.Description)
		sections = append(sections, desc)
	}

	// Divider
	divider := lipgloss.NewStyle().
		Foreground(colorMuted).
		Width(contentWidth).
		Render(strings.Repeat("─", min(contentWidth, 40)))
	sections = append(sections, divider)

	// Command to run
	sections = append(sections,
		detailSection.Render("COMMAND")+"\n"+
			detailCode.Render(c.Pattern),
	)

	// Tags
	if c.Tags != "" {
		tags := strings.Split(c.Tags, ",")
		var rendered []string
		for _, t := range tags {
			t = strings.TrimSpace(t)
			if t != "" {
				rendered = append(rendered, tagStyle.Render(t))
			}
		}
		sections = append(sections,
			detailSection.Render("TAGS")+"\n"+
				strings.Join(rendered, " "),
		)
	}

	// Usage
	if c.UsageCount > 0 {
		usageLine := fmt.Sprintf("Used %d time", c.UsageCount)
		if c.UsageCount != 1 {
			usageLine += "s"
		}
		if !c.LastUsedAt.IsZero() {
			usageLine += " · last " + timeAgo(c.LastUsedAt)
		}
		sections = append(sections,
			detailSection.Render("USAGE")+"\n"+
				detailBody.Render(usageLine),
		)
	}

	// Source
	if c.Source != "" {
		label := "local"
		if c.Source != "local" {
			label = c.Source + " (synced)"
		}
		sections = append(sections,
			detailSection.Render("SOURCE")+"\n"+
				sourceStyle.Render(label),
		)
	}

	// Workspace
	if c.WorkspaceFilter != "" {
		wsLine := c.WorkspaceFilter
		if m.matchCwd {
			wsLine += "  " + workspaceBadge.Render("✓ matches cwd")
		}
		sections = append(sections,
			detailSection.Render("WORKSPACE")+"\n"+
				detailBody.Render(wsLine),
		)
	}

	content := strings.Join(sections, "\n\n")

	return lipgloss.NewStyle().
		Width(contentWidth).
		Render(content)
}

// timeAgo returns a human-readable string describing how long ago t occurred
// (e.g. "just now", "5 mins ago", "2 days ago").
func timeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "yesterday"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

// renderStatusBar returns the full-layout key-hint bar that appears at the bottom
// of the TUI when the terminal is wide enough for the split-pane view.
func renderStatusBar(width int, total int) string {
	keys := []struct {
		key  string
		desc string
	}{
		{"enter", "run"},
		{"/", "search"},
		{"t", "tags"},
		{"a", "add"},
		{"e", "edit"},
		{"x", "delete"},
		{"q", "quit"},
	}

	var parts []string
	for _, k := range keys {
		parts = append(parts,
			statusKey.Render(k.key)+" "+statusDesc.Render(k.desc),
		)
	}

	left := strings.Join(parts, statusDesc.Render("  │  "))

	right := statusDesc.Render(fmt.Sprintf("%d commands", total))

	gap := width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}

	bar := statusBar.
		Width(width).
		Background(lipgloss.AdaptiveColor{Light: "#F0F0F0", Dark: "#1A1A1A"}).
		Render(left + strings.Repeat(" ", gap) + right)

	return bar
}

// renderCompactStatusBar returns a trimmed key-hint bar used when the terminal
// is too narrow or short for the full split-pane layout.
func renderCompactStatusBar(width int, total int) string {
	keys := []struct {
		key  string
		desc string
	}{
		{"↵", "run"},
		{"/", "find"},
		{"a", "add"},
		{"q", "quit"},
	}

	var parts []string
	for _, k := range keys {
		parts = append(parts,
			statusKey.Render(k.key)+" "+statusDesc.Render(k.desc),
		)
	}

	left := strings.Join(parts, statusDesc.Render(" │ "))
	right := statusDesc.Render(fmt.Sprintf("%d", total))

	gap := width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}

	return statusBar.
		Width(width).
		Background(lipgloss.AdaptiveColor{Light: "#F0F0F0", Dark: "#1A1A1A"}).
		Render(left + strings.Repeat(" ", gap) + right)
}
