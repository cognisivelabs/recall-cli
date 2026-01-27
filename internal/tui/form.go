package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/CognisiveLabs/recall-cli/internal/storage"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FormSavedMsg struct{}

type FormModel struct {
	inputs  []textinput.Model
	focused int
	err     error
	id      int // If 0, insert. If >0, update.
	store   storage.Storage
}

func NewFormModel(store storage.Storage, existing *storage.Command) FormModel {
	m := FormModel{
		store:  store,
		inputs: make([]textinput.Model, 3),
	}

	var initialCmd, initialDesc, initialTags string
	if existing != nil {
		m.id = existing.ID
		initialCmd = existing.Pattern
		initialDesc = existing.Description
		initialTags = existing.Tags
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = lipgloss.NewStyle()
		t.CharLimit = 100

		switch i {
		case 0:
			t.Placeholder = "Command"
			t.SetValue(initialCmd)
			t.Focus()
			t.Prompt = "CMD > "
		case 1:
			t.Placeholder = "Description"
			t.CharLimit = 200
			t.SetValue(initialDesc)
			t.Prompt = "DESC > "
		case 2:
			t.Placeholder = "Tags (comma separated)"
			t.SetValue(initialTags)
			t.Prompt = "TAGS > "
		}

		m.inputs[i] = t
	}

	// Focus first input if new, or maybe first input if edit too
	return m
}

// ... Init, Update, View (mostly unchanged, except checking ID for button label?) ...

func saveCommand(m FormModel) tea.Msg {
	// Save logic
	// db call removed, usage of m.store
	cmd := m.inputs[0].Value()
	desc := m.inputs[1].Value()
	tags := m.inputs[2].Value()

	targetID := m.id

	// If insert (id == 0), check if pattern already exists to prevent duplicate
	if targetID == 0 {
		existing, err := m.store.GetByPattern(cmd)
		if err == nil && existing != nil {
			targetID = existing.ID // Switch to update mode for this ID
		}
	}

	var err error
	if targetID != 0 {
		// Update
		c := storage.Command{
			ID:          targetID,
			Pattern:     cmd,
			Description: desc,
			Tags:        tags,
		}
		err = m.store.Update(c)
	} else {
		// Insert aka Upsert in our interface currently
		c := storage.Command{
			Pattern:     cmd,
			Description: desc,
			Tags:        tags,
		}
		err = m.store.Upsert(c)
	}

	if err != nil {
		return err
	}

	return func() tea.Msg { return FormSavedMsg{} }
}

func StartForm(store storage.Storage, existing *storage.Command) error {
	p := tea.NewProgram(NewFormModel(store, existing), tea.WithOutput(os.Stderr))
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run form: %w", err)
	}
	return nil
}

func (m FormModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m FormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case error:
		m.err = msg
		return m, nil

	case FormSavedMsg:
		return m, tea.Quit

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "ctrl+s", "ctrl+o":
			return m, func() tea.Msg {
				return saveCommand(m)
			}

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// If error exists, clear it on any navigation
			if m.err != nil {
				m.err = nil
			}

			// If enter on submit button (index 3)
			if s == "enter" && m.focused == len(m.inputs) {
				return m, func() tea.Msg {
					return saveCommand(m)
				}
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focused--
			} else {
				m.focused++
			}

			// 0..len(inputs) -> last is button
			if m.focused > len(m.inputs) {
				m.focused = 0
			} else if m.focused < 0 {
				m.focused = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i < len(m.inputs); i++ {
				if i == m.focused {
					// Set focused state
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
					continue
				}
				// Remove focused state
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input only if focusing an input (not button)
	if m.focused < len(m.inputs) {
		cmd := m.updateInputs(msg)
		return m, cmd
	}

	return m, nil
}

func (m *FormModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func (m FormModel) View() string {
	var b strings.Builder

	b.WriteString("Recall: Save Command\n\n")

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := "\n\n[ Submit ]"
	if m.focused == len(m.inputs) {
		button = "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("[ Submit ]")
	}
	b.WriteString(button)

	if m.err != nil {
		b.WriteString("\n\n")
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Error: " + m.err.Error()))
	}

	b.WriteString("\n\n(tab to navigate, enter to select, ctrl+o to save)")

	return b.String()
}
