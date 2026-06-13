package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/CognisiveLabs/recall-cli/internal/storage"
	"github.com/CognisiveLabs/recall-cli/internal/workspace"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type item struct {
	command  storage.Command
	matchCwd bool
}

func (i item) Title() string       { return i.command.Pattern }
func (i item) Description() string { return i.command.Description }
func (i item) FilterValue() string { return i.command.Pattern + " " + i.command.Description + " " + i.command.Tags }

type ListModel struct {
	list list.Model
}

// NewListModel wraps a Bubble Tea list with recall-specific configuration.
// cwd is used to mark items that match the user's current working directory.
func NewListModel(commands []storage.Command, cwd string) ListModel {
	items := make([]list.Item, len(commands))
	for i, cmd := range commands {
		items[i] = item{
			command:  cmd,
			matchCwd: workspace.Matches(cwd, cmd.WorkspaceFilter),
		}
	}

	delegate := &commandDelegate{}
	l := list.New(items, delegate, 0, 0)
	l.Title = ""
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.FilterInput.Prompt = "  search: "
	l.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(colorAccent)
	l.FilterInput.Cursor.Style = lipgloss.NewStyle().Foreground(colorAccent)
	l.Styles.NoItems = emptyStyle

	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add")),
			key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
			key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "delete")),
		}
	}

	return ListModel{list: l}
}

func (m ListModel) Init() tea.Cmd {
	return nil
}

func (m ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ListModel) View() string {
	return m.list.View()
}

func (m *ListModel) SetSize(width, height int) {
	m.list.SetSize(width, height)
}

// SelectedItem returns the currently highlighted command, or nil if nothing is selected.
func (m ListModel) SelectedItem() *storage.Command {
	if i, ok := m.list.SelectedItem().(item); ok {
		return &i.command
	}
	return nil
}

// SelectedMatchesCwd reports whether the currently selected command's workspace
// filter matches the user's current working directory.
func (m ListModel) SelectedMatchesCwd() bool {
	if i, ok := m.list.SelectedItem().(item); ok {
		return i.matchCwd
	}
	return false
}

// Custom delegate for styled list items
type commandDelegate struct{}

func (d *commandDelegate) Height() int                         { return 3 }
func (d *commandDelegate) Spacing() int                        { return 0 }
func (d *commandDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d *commandDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	width := m.Width() - 4
	if width < 10 {
		width = 10
	}

	isSelected := index == m.Index()

	// Command pattern (title line)
	pattern := i.command.Pattern
	if len(pattern) > width {
		pattern = pattern[:width-1] + "…"
	}

	// Description (second line)
	desc := i.command.Description
	if desc == "" {
		desc = "—"
	}
	if len(desc) > width {
		desc = desc[:width-1] + "…"
	}

	// Third line: tags + source + workspace
	var meta []string
	if i.command.Tags != "" {
		tags := strings.Split(i.command.Tags, ",")
		for _, t := range tags {
			t = strings.TrimSpace(t)
			if t != "" {
				meta = append(meta, tagStyle.Render(t))
			}
		}
	}
	if i.command.Source != "" && i.command.Source != "local" {
		meta = append(meta, sourceStyle.Render("⟨"+i.command.Source+"⟩"))
	}
	if i.command.UsageCount > 0 {
		usageStyle := lipgloss.NewStyle().Foreground(colorDimText)
		meta = append(meta, usageStyle.Render(fmt.Sprintf("↑%d", i.command.UsageCount)))
	}
	if i.matchCwd {
		meta = append(meta, workspaceBadge.Render("● here"))
	}
	metaLine := strings.Join(meta, " ")

	if isSelected {
		patternStyle := lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

		descStyle := lipgloss.NewStyle().
			Foreground(colorHighlight)

		indicator := lipgloss.NewStyle().
			Foreground(colorAccent).
			Render("▸ ")

		fmt.Fprintf(w, "%s%s\n  %s\n  %s",
			indicator,
			patternStyle.Render(pattern),
			descStyle.Render(desc),
			metaLine,
		)
	} else {
		patternStyle := lipgloss.NewStyle().
			Foreground(colorHighlight)

		descStyle := lipgloss.NewStyle().
			Foreground(colorSubtle)

		fmt.Fprintf(w, "  %s\n  %s\n  %s",
			patternStyle.Render(pattern),
			descStyle.Render(desc),
			metaLine,
		)
	}
}
