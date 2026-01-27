package tui

import (
	"github.com/CognisiveLabs/recall-cli/internal/storage"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type item struct {
	command storage.Command
}

func (i item) Title() string       { return i.command.Pattern }
func (i item) Description() string { return i.command.Description }
func (i item) FilterValue() string { return i.command.Pattern + " " + i.command.Description }

type ListModel struct {
	list list.Model
}

func NewListModel(commands []storage.Command) ListModel {
	items := make([]list.Item, len(commands))
	for i, cmd := range commands {
		items[i] = item{command: cmd}
	}

	// Default delegate
	delegate := list.NewDefaultDelegate()

	l := list.New(items, delegate, 0, 0)
	l.Title = "Recall"
	l.SetShowStatusBar(false)

	// Add custom help keys
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "delete")),
			key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
		}
	}
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "delete")),
			key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
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

func (m ListModel) SelectedItem() *storage.Command {
	if i, ok := m.list.SelectedItem().(item); ok {
		return &i.command
	}
	return nil
}
