package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/CognisiveLabs/recall-cli/internal/placeholders"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ResolvingModel struct {
	originalCommand string
	currentCommand  string
	pending         []placeholders.Placeholder

	textInput textinput.Model
	listInput list.Model

	isOptions bool
	err       error
	done      bool
}

func (m ResolvingModel) Done() bool      { return m.done }
func (m ResolvingModel) Resolved() string { return m.currentCommand }

// NewResolverProgram wraps a ResolvingModel in a Bubble Tea program that writes to stderr.
func NewResolverProgram(m ResolvingModel) *tea.Program {
	return tea.NewProgram(m, tea.WithOutput(os.Stderr))
}

// NewResolvingModel builds a resolver starting from the raw command string.
// Auto-resolves {{branch}}, {{cwd}}, etc. first, then prompts the user for the rest.
func NewResolvingModel(cmd string) ResolvingModel {
	// Auto-resolve first, then only prompt for remaining
	resolved, remaining := placeholders.AutoResolve(cmd)
	m := ResolvingModel{
		originalCommand: cmd,
		currentCommand:  resolved,
		pending:         remaining,
		textInput:       textinput.New(),
	}
	m.setupNextInput()
	return m
}

// NewResolvingModelFromParsed creates a resolver when auto-resolution has already happened
// upstream (e.g. in the TUI or `recall run`). remaining contains only the placeholders
// that still need user input.
func NewResolvingModelFromParsed(cmd string, remaining []placeholders.Placeholder) ResolvingModel {
	m := ResolvingModel{
		originalCommand: cmd,
		currentCommand:  cmd,
		pending:         remaining,
		textInput:       textinput.New(),
	}
	m.setupNextInput()
	return m
}

// setupNextInput prepares the UI for the next unresolved placeholder.
// Sets isOptions=true and builds a list picker for "options:" types; otherwise
// creates a plain text input. Sets done=true when all placeholders are resolved.
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

		delegate := list.NewDefaultDelegate()
		delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
		delegate.Styles.NormalTitle = lipgloss.NewStyle().Foreground(colorHighlight)

		l := list.New(items, delegate, 40, 10)
		l.Title = "Select: " + p.Key
		l.Styles.Title = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
		l.SetShowStatusBar(false)
		l.SetFilteringEnabled(false)
		l.SetShowHelp(false)
		m.listInput = l
	} else {
		m.isOptions = false
		m.textInput = textinput.New()
		m.textInput.Placeholder = "enter value"
		m.textInput.Prompt = p.Key + " ▸ "
		m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
		m.textInput.Cursor.Style = lipgloss.NewStyle().Foreground(colorAccent)
		m.textInput.TextStyle = lipgloss.NewStyle().Foreground(colorHighlight)
		m.textInput.Focus()
		m.textInput.SetValue("")
	}
}

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

	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		}
	}

	if m.isOptions {
		if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "enter" {
			selected := m.listInput.SelectedItem()
			if selected != nil {
				val := selected.FilterValue()
				m.resolveCurrent(val)
			}
			if m.done {
				return m, tea.Quit
			}
			return m, nil
		}

		m.listInput, cmd = m.listInput.Update(msg)
		return m, cmd
	} else {
		if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "enter" {
			val := m.textInput.Value()
			m.resolveCurrent(val)
			if m.done {
				return m, tea.Quit
			}
			return m, nil
		}
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
}

// resolveCurrent substitutes val into the current command string for the first
// pending placeholder, then advances to the next one.
func (m *ResolvingModel) resolveCurrent(val string) {
	p := m.pending[0]
	m.currentCommand = placeholders.Replace(m.currentCommand, p, val)
	m.pending = m.pending[1:]
	m.setupNextInput()
}

func (m ResolvingModel) View() string {
	if m.done {
		return ""
	}

	remaining := fmt.Sprintf("%d remaining", len(m.pending))

	// Highlight placeholders in command preview
	preview := m.currentCommand
	for _, p := range m.pending {
		preview = strings.Replace(preview,
			p.FullMatch,
			lipgloss.NewStyle().Foreground(colorWarn).Bold(true).Render(p.FullMatch),
			1,
		)
	}

	var content string
	if m.isOptions {
		content = m.listInput.View()
	} else {
		content = m.textInput.View()
	}

	title := headerStyle.Render("recall · resolve placeholders")
	counter := statusDesc.Render(remaining)

	return lipgloss.NewStyle().Padding(1, 2).Render(
		title + "  " + counter + "\n\n" +
			detailSection.Render("COMMAND") + "\n" +
			detailCode.Render(preview) + "\n\n" +
			content + "\n\n" +
			formHint.Render("enter confirm · esc cancel"),
	)
}
