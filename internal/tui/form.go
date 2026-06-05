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
	id      int
	store   storage.Storage
}

var (
	formTitle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true).
			MarginBottom(1)

	formFocusedPrompt = lipgloss.NewStyle().
				Foreground(colorAccent)

	formBlurredPrompt = lipgloss.NewStyle().
				Foreground(colorMuted)

	formButton = lipgloss.NewStyle().
			Foreground(colorSubtle).
			Background(lipgloss.AdaptiveColor{Light: "#E8E8E8", Dark: "#2A2A2A"}).
			Padding(0, 3)

	formButtonActive = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}).
				Background(colorAccent).
				Padding(0, 3).
				Bold(true)

	formError = lipgloss.NewStyle().
			Foreground(colorWarn)

	formHint = lipgloss.NewStyle().
			Foreground(colorDimText).
			Italic(true)
)

func NewFormModel(store storage.Storage, existing *storage.Command) FormModel {
	m := FormModel{
		store:  store,
		inputs: make([]textinput.Model, 4),
	}

	var initialCmd, initialDesc, initialTags, initialWorkspace string
	if existing != nil {
		m.id = existing.ID
		initialCmd = existing.Pattern
		initialDesc = existing.Description
		initialTags = existing.Tags
		initialWorkspace = existing.WorkspaceFilter
	}

	labels := []struct {
		prompt      string
		placeholder string
		value       string
		charLimit   int
	}{
		{"command   │ ", "e.g. kubectl logs -f deploy/{{service}}", initialCmd, 200},
		{"describe  │ ", "What does this command do?", initialDesc, 200},
		{"tags      │ ", "comma separated, e.g. k8s, debug", initialTags, 100},
		{"workspace │ ", "e.g. ~/work/billing-*  (optional)", initialWorkspace, 200},
	}

	for i, l := range labels {
		t := textinput.New()
		t.Prompt = l.prompt
		t.Placeholder = l.placeholder
		t.SetValue(l.value)
		t.CharLimit = l.charLimit
		t.Cursor.Style = lipgloss.NewStyle().Foreground(colorAccent)

		if i == 0 {
			t.Focus()
			t.PromptStyle = formFocusedPrompt
			t.TextStyle = lipgloss.NewStyle().Foreground(colorHighlight)
		} else {
			t.PromptStyle = formBlurredPrompt
			t.TextStyle = lipgloss.NewStyle().Foreground(colorSubtle)
		}

		m.inputs[i] = t
	}

	return m
}

func saveCommand(m FormModel) tea.Msg {
	cmd := m.inputs[0].Value()
	desc := m.inputs[1].Value()
	tags := m.inputs[2].Value()
	ws := m.inputs[3].Value()

	targetID := m.id

	if targetID == 0 {
		existing, err := m.store.GetByPattern(cmd)
		if err == nil && existing != nil {
			targetID = existing.ID
		}
	}

	var err error
	if targetID != 0 {
		c := storage.Command{
			ID:              targetID,
			Pattern:         cmd,
			Description:     desc,
			Tags:            tags,
			WorkspaceFilter: ws,
		}
		err = m.store.Update(c)
	} else {
		c := storage.Command{
			Pattern:         cmd,
			Description:     desc,
			Tags:            tags,
			WorkspaceFilter: ws,
		}
		err = m.store.Upsert(c)
	}

	if err != nil {
		return err
	}

	return FormSavedMsg{}
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

		case "ctrl+s":
			return m, func() tea.Msg {
				return saveCommand(m)
			}

		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			if m.err != nil {
				m.err = nil
			}

			if s == "enter" && m.focused == len(m.inputs) {
				return m, func() tea.Msg {
					return saveCommand(m)
				}
			}

			if s == "up" || s == "shift+tab" {
				m.focused--
			} else {
				m.focused++
			}

			if m.focused > len(m.inputs) {
				m.focused = 0
			} else if m.focused < 0 {
				m.focused = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i < len(m.inputs); i++ {
				if i == m.focused {
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = formFocusedPrompt
					m.inputs[i].TextStyle = lipgloss.NewStyle().Foreground(colorHighlight)
					continue
				}
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = formBlurredPrompt
				m.inputs[i].TextStyle = lipgloss.NewStyle().Foreground(colorSubtle)
			}

			return m, tea.Batch(cmds...)
		}
	}

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

	// Title
	action := "Save Command"
	if m.id > 0 {
		action = "Edit Command"
	}
	b.WriteString(formTitle.Render("recall · "+action) + "\n\n")

	// Inputs
	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		b.WriteRune('\n')
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	// Submit button
	b.WriteString("\n\n")
	if m.focused == len(m.inputs) {
		b.WriteString("  " + formButtonActive.Render("Save"))
	} else {
		b.WriteString("  " + formButton.Render("Save"))
	}

	// Error
	if m.err != nil {
		b.WriteString("\n\n")
		b.WriteString("  " + formError.Render("Error: "+m.err.Error()))
	}

	// Hint
	b.WriteString("\n\n")
	b.WriteString("  " + formHint.Render("tab/↑↓ navigate · enter select · ctrl+s save · esc cancel"))

	return lipgloss.NewStyle().Padding(1, 3).Render(b.String())
}
