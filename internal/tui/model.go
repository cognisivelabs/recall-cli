package tui

import (
	"fmt"
	"os"

	"github.com/CognisiveLabs/recall-cli/internal/placeholders"
	"github.com/CognisiveLabs/recall-cli/internal/storage"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type State int

const (
	StateList State = iota
	StateEdit
)

type Model struct {
	list     ListModel
	detail   DetailModel
	form     FormModel
	state    State
	ready    bool
	width    int
	height   int
	Selected string
	store    storage.Storage
}

func NewModel(store storage.Storage) Model {
	// Load actual commands
	cmds, err := store.List()
	if err != nil {
		// Log error? For TUI app, maybe simple default?
		// We can initialize with error message item
		cmds = []storage.Command{}
	}
	if len(cmds) == 0 {
		cmds = []storage.Command{
			{Pattern: "", Description: "No commands found. Try 'recall save'!"},
		}
	}

	return Model{
		list:   NewListModel(cmds),
		detail: NewDetailModel(),
		state:  StateList,
		store:  store,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// 1. Handle Global Messages (State Transitions)
	if _, ok := msg.(FormSavedMsg); ok {
		m.state = StateList
		m.state = StateList
		// Reload commands from DB
		cmdsList, _ := m.store.List()
		if len(cmdsList) == 0 {
			cmdsList = []storage.Command{{Pattern: "", Description: "No commands found."}}
		}
		m.list = NewListModel(cmdsList)
		m.list.SetSize(int(float64(m.width)*0.35), m.height)
		return m, nil
	}

	// 2. State-Specific Logic
	if m.state == StateEdit {
		var formModel tea.Model
		formModel, cmd = m.form.Update(msg)
		m.form = formModel.(FormModel)
		return m, cmd
	}

	// 3. List Mode Logic
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			selected := m.list.SelectedItem()
			if selected != nil {
				m.Selected = selected.Pattern
				return m, tea.Quit
			}
		case "x": // Delete
			selected := m.list.SelectedItem()
			if selected != nil && selected.ID != 0 {
				m.store.Delete(selected.ID)
				// Reload
				cmdsList, _ := m.store.List()
				if len(cmdsList) == 0 {
					cmdsList = []storage.Command{{Pattern: "", Description: "No commands found."}}
				}
				m.list = NewListModel(cmdsList)
				m.list.SetSize(int(float64(m.width)*0.35), m.height)
			}
		case "e": // Edit
			selected := m.list.SelectedItem()
			if selected != nil && selected.ID != 0 {
				m.state = StateEdit
				m.form = NewFormModel(m.store, selected)
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		listWidth := int(float64(msg.Width) * 0.35)
		detailWidth := msg.Width - listWidth - 4

		m.list.SetSize(listWidth, msg.Height)
		m.detail.SetSize(detailWidth, msg.Height)
	}

	// Forward msg to List component
	var listModel tea.Model
	listModel, cmd = m.list.Update(msg)
	m.list = listModel.(ListModel)
	cmds = append(cmds, cmd)

	// Update Detail based on selection
	selected := m.list.SelectedItem()
	if selected != nil {
		content := fmt.Sprintf("# %s\n\n%s\n\n**Run:** `%s`", selected.Pattern, selected.Description, selected.Pattern)
		m.detail.SetContent(content)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	if m.state == StateEdit {
		return m.form.View()
	}

	// Style split
	listView := lipgloss.NewStyle().
		Width(m.list.list.Width()).
		Height(m.height).
		MarginRight(1).
		Render(m.list.View())

	detailView := lipgloss.NewStyle().
		Width(m.detail.width).
		Height(m.height).
		Render(m.detail.View())

	return lipgloss.JoinHorizontal(lipgloss.Top, listView, detailView)
}

func Start(store storage.Storage) (string, error) {
	// Critical: Write TUI to Stderr so that Stdout can be used independently for the selected command.
	p := tea.NewProgram(NewModel(store), tea.WithAltScreen(), tea.WithOutput(os.Stderr))
	m, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run tui: %w", err)
	}

	if model, ok := m.(Model); ok && model.Selected != "" {
		cmd := model.Selected
		// Check for placeholders
		if placeholders.HasPlaceholders(cmd) {
			// Run resolver TUI
			rm := NewResolvingModel(cmd)
			pr := tea.NewProgram(rm, tea.WithOutput(os.Stderr))
			res, err := pr.Run()
			if err != nil {
				return "", fmt.Errorf("failed to resolve placeholders: %w", err)
			}
			if resolved, ok := res.(ResolvingModel); ok && resolved.done {
				return resolved.currentCommand, nil
			}
			return "", nil // Cancelled
		}
		return cmd, nil
	}

	return "", nil
}
