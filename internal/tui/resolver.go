package tui

import (
	"fmt"

	"github.com/CognisiveLabs/recall-cli/internal/placeholders"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ResolvingModel handles the step-by-step resolution of placeholders
type ResolvingModel struct {
	originalCommand string
	currentCommand  string
	pending         []placeholders.Placeholder

	// Input types
	textInput textinput.Model
	listInput list.Model

	isOptions bool
	err       error
	done      bool
}

func NewResolvingModel(cmd string) ResolvingModel {
	pending := placeholders.Parse(cmd)
	m := ResolvingModel{
		originalCommand: cmd,
		currentCommand:  cmd,
		pending:         pending,
		textInput:       textinput.New(),
	}
	m.setupNextInput()
	return m
}

func (m *ResolvingModel) setupNextInput() {
	if len(m.pending) == 0 {
		m.done = true
		return
	}

	p := m.pending[0]
	if p.Type == "options" {
		m.isOptions = true
		items := make([]list.Item, len(p.Options))
		for i, opt := range p.Options {
			items[i] = optionItem(opt)
		}

		l := list.New(items, list.NewDefaultDelegate(), 0, 0)
		l.Title = "Select " + p.Key
		l.SetShowStatusBar(false)
		l.SetFilteringEnabled(false)
		m.listInput = l
	} else {
		m.isOptions = false
		m.textInput.Placeholder = "Enter " + p.Key
		m.textInput.Prompt = p.Key + " > "
		m.textInput.Focus()
		m.textInput.SetValue("")
	}
}

// Simple list item adapter
type optionItem string

func (o optionItem) FilterValue() string { return string(o) }
func (o optionItem) Title() string       { return string(o) }
func (o optionItem) Description() string { return "" }

func (m ResolvingModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m ResolvingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.done {
		return m, tea.Quit
	}

	var cmd tea.Cmd

	// Quit keys
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		}
	}

	if m.isOptions {
		// List update
		if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "enter" {
			selected := m.listInput.SelectedItem()
			if selected != nil {
				val := selected.FilterValue()
				m.resolveCurrent(val)
			}
			return m, nil
		}

		m.listInput, cmd = m.listInput.Update(msg)
		return m, cmd
	} else {
		// Text update
		if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "enter" {
			val := m.textInput.Value()
			m.resolveCurrent(val)
			return m, nil
		}
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
}

func (m *ResolvingModel) resolveCurrent(val string) {
	p := m.pending[0]
	// Replace ALL instances or just one? Let's say one for now, or all if same name?
	// The parser found instances. If specific logic needed, use parser.
	// Simple replacement:
	m.currentCommand = placeholders.Replace(m.currentCommand, p, val)

	// Remove this pending, and any others that were identical?
	// For simplicity, just pop first. User might have {{file}} twice and want different values?
	// Usually repeated placeholder means same value.
	// Let's just pop the head for now.
	m.pending = m.pending[1:]
	m.setupNextInput()
}

func (m ResolvingModel) View() string {
	if m.done {
		return fmt.Sprintf("Ready: %s", m.currentCommand)
	}

	var content string
	if m.isOptions {
		content = m.listInput.View()
	} else {
		content = m.textInput.View()
	}

	return fmt.Sprintf(
		"Resolving Placeholders...\n\nCommand: %s\n\n%s",
		highlightPlaceholders(m.currentCommand),
		content,
	)
}

func highlightPlaceholders(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(s)
}
