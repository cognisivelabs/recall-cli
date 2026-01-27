package tui

import (
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type DetailModel struct {
	content  string
	width    int
	renderer *glamour.TermRenderer
}

func NewDetailModel() DetailModel {
	r, _ := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(80),
	)
	return DetailModel{
		renderer: r,
	}
}

func (m *DetailModel) SetContent(content string) {
	m.content = content
}

func (m *DetailModel) SetSize(width, height int) {
	m.width = width
	// Update renderer wrap
	m.renderer, _ = glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(width),
	)
}

func (m DetailModel) View() string {
	if m.content == "" {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("No command selected.")
	}

	out, err := m.renderer.Render(m.content)
	if err != nil {
		return "Error rendering markdown"
	}
	return out
}
