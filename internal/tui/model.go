package tui

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/CognisiveLabs/recall-cli/internal/placeholders"
	"github.com/CognisiveLabs/recall-cli/internal/storage"
	"github.com/CognisiveLabs/recall-cli/internal/workspace"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type State int

const (
	StateList State = iota
	StateEdit
	StateTagPicker
)

type LayoutMode int

const (
	LayoutFull    LayoutMode = iota // split pane: list + detail
	LayoutCompact                  // single pane: list only
	LayoutTiny                     // too small to render
)

const (
	minWidth       = 40
	minHeight      = 10
	compactMaxW    = 80
	compactMaxH    = 20
)

type Model struct {
	list      ListModel
	detail    DetailModel
	form      FormModel
	state     State
	ready     bool
	width     int
	height    int
	layout    LayoutMode
	Selected  string
	store     storage.Storage
	cwd       string
	total     int
	allCmds   []storage.Command // unfiltered
	allTags   []string
	activeTag string // "" means show all
	tagCursor int
}

func detectLayout(w, h int) LayoutMode {
	if w < minWidth || h < minHeight {
		return LayoutTiny
	}
	if w < compactMaxW || h < compactMaxH {
		return LayoutCompact
	}
	return LayoutFull
}

func NewModel(store storage.Storage) Model {
	cmds, err := store.List()
	if err != nil {
		cmds = []storage.Command{}
	}

	cwd := workspace.Detect()
	cmds = sortByWorkspace(cmds, cwd)
	tags := storage.CollectTags(cmds)
	total := len(cmds)

	displayCmds := cmds
	if len(displayCmds) == 0 {
		displayCmds = []storage.Command{
			{Pattern: "", Description: "No commands saved yet. Use 'recall add' to get started."},
		}
	}

	return Model{
		list:    NewListModel(displayCmds, cwd),
		detail:  NewDetailModel(),
		state:   StateList,
		store:   store,
		cwd:     cwd,
		total:   total,
		allCmds: cmds,
		allTags: tags,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	if _, ok := msg.(FormSavedMsg); ok {
		m.state = StateList
		m.activeTag = ""
		m = m.reloadList()
		return m, nil
	}

	if m.state == StateEdit {
		var formModel tea.Model
		formModel, cmd = m.form.Update(msg)
		m.form = formModel.(FormModel)
		return m, cmd
	}

	if m.state == StateTagPicker {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "q", "esc", "t":
				m.state = StateList
				return m, nil
			case "up", "k":
				if m.tagCursor > 0 {
					m.tagCursor--
				}
			case "down", "j":
				if m.tagCursor < len(m.allTags) {
					m.tagCursor++
				}
			case "enter":
				if m.tagCursor == 0 {
					m.activeTag = ""
				} else {
					m.activeTag = m.allTags[m.tagCursor-1]
				}
				m.state = StateList
				m = m.applyTagFilter()
				return m, nil
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.list.list.FilterState() == list.Filtering {
			break
		}
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			selected := m.list.SelectedItem()
			if selected != nil && selected.Pattern != "" {
				m.Selected = selected.Pattern
				if selected.ID != 0 {
					m.store.RecordUsage(selected.ID)
				}
				return m, tea.Quit
			}
		case "x":
			selected := m.list.SelectedItem()
			if selected != nil && selected.ID != 0 {
				m.store.Delete(selected.ID)
				m.activeTag = ""
				m = m.reloadList()
			}
		case "e":
			selected := m.list.SelectedItem()
			if selected != nil && selected.ID != 0 {
				m.state = StateEdit
				m.form = NewFormModel(m.store, selected)
			}
		case "a":
			m.state = StateEdit
			m.form = NewFormModel(m.store, nil)
		case "t":
			if len(m.allTags) > 0 {
				m.state = StateTagPicker
				m.tagCursor = 0
				// Position cursor on active tag
				if m.activeTag != "" {
					for i, t := range m.allTags {
						if t == m.activeTag {
							m.tagCursor = i + 1
							break
						}
					}
				}
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layout = detectLayout(msg.Width, msg.Height)
		m.ready = true
		m = m.resizePanes()
	}

	var listModel tea.Model
	listModel, cmd = m.list.Update(msg)
	m.list = listModel.(ListModel)
	cmds = append(cmds, cmd)

	selected := m.list.SelectedItem()
	matchCwd := m.list.SelectedMatchesCwd()
	m.detail.SetCommand(selected, matchCwd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if !m.ready {
		return "\n  Loading..."
	}

	if m.state == StateEdit {
		return m.form.View()
	}

	// Too small — just show a resize message
	if m.layout == LayoutTiny {
		return m.viewTiny()
	}

	// Compact — single pane, no detail
	if m.layout == LayoutCompact {
		return m.viewCompact()
	}

	// Full — split pane
	return m.viewFull()
}

func (m Model) viewTiny() string {
	msg := lipgloss.NewStyle().
		Foreground(colorSubtle).
		Padding(1, 2).
		Render(fmt.Sprintf(
			"Terminal too small (%d×%d)\nNeed at least %d×%d\n\nResize or press q to quit",
			m.width, m.height, minWidth, minHeight,
		))
	return msg
}

func (m Model) viewCompact() string {
	// Minimal header
	header := lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1).
		Render(headerStyle.Render("recall") + m.tagIndicator())

	headerH := lipgloss.Height(header)
	statusH := 1
	bodyH := m.height - headerH - statusH

	// Full-width list, no border to save space
	m.list.SetSize(m.width-2, bodyH)
	listContent := lipgloss.NewStyle().
		Width(m.width).
		Height(bodyH).
		Render(m.list.View())

	status := renderCompactStatusBar(m.width, m.total)

	return lipgloss.JoinVertical(lipgloss.Left, header, listContent, status)
}

func (m Model) viewFull() string {
	// Header
	headerLeft := lipgloss.JoinHorizontal(lipgloss.Center,
		headerStyle.Render("recall"),
		headerDim.Render(" · your terminal memory"),
	)
	headerRight := m.tagIndicator()

	headerGap := m.width - lipgloss.Width(headerLeft) - lipgloss.Width(headerRight) - 4
	if headerGap < 1 {
		headerGap = 1
	}
	header := lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1).
		Render(headerLeft + strings.Repeat(" ", headerGap) + headerRight)

	headerH := lipgloss.Height(header)
	statusH := 1
	bodyH := m.height - headerH - statusH - 1

	listWidth := int(float64(m.width) * 0.40)
	detailWidth := m.width - listWidth - 1

	listContent := listPaneStyle.
		Width(listWidth - 2).
		Height(bodyH - 2).
		Render(m.list.View())

	var rightPane string
	if m.state == StateTagPicker {
		rightPane = detailPaneStyle.
			Width(detailWidth - 2).
			Height(bodyH - 2).
			Render(m.renderTagPicker(detailWidth - 6))
	} else {
		rightPane = detailPaneStyle.
			Width(detailWidth - 2).
			Height(bodyH - 2).
			Render(m.detail.View())
	}

	body := lipgloss.JoinHorizontal(lipgloss.Top, listContent, rightPane)

	status := renderStatusBar(m.width, m.total)

	return lipgloss.JoinVertical(lipgloss.Left, header, body, status)
}

func (m Model) tagIndicator() string {
	if m.activeTag == "" {
		return ""
	}
	return " " + tagStyle.Render(m.activeTag) + " " +
		lipgloss.NewStyle().Foreground(colorDimText).Render("(t)")
}

func (m Model) renderTagPicker(width int) string {
	title := detailTitle.Render("Filter by Tag")

	var rows []string

	// "All" option
	label := "  all"
	if m.tagCursor == 0 {
		if m.activeTag == "" {
			label = lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render("▸ all ✓")
		} else {
			label = lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render("▸ all")
		}
	} else if m.activeTag == "" {
		label = lipgloss.NewStyle().Foreground(colorSuccess).Render("  all ✓")
	}
	rows = append(rows, label)

	for i, t := range m.allTags {
		cursor := i + 1
		entry := "  " + t
		if cursor == m.tagCursor {
			if m.activeTag == t {
				entry = lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render("▸ " + t + " ✓")
			} else {
				entry = lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render("▸ " + t)
			}
		} else if m.activeTag == t {
			entry = lipgloss.NewStyle().Foreground(colorSuccess).Render("  " + t + " ✓")
		} else {
			entry = lipgloss.NewStyle().Foreground(colorHighlight).Render("  " + t)
		}
		rows = append(rows, entry)
	}

	hint := formHint.Render("↑↓ navigate · enter select · esc back")

	return title + "\n\n" + strings.Join(rows, "\n") + "\n\n" + hint
}

func (m Model) reloadList() Model {
	cmdsList, _ := m.store.List()
	m.allCmds = cmdsList
	m.allTags = storage.CollectTags(cmdsList)
	m.total = len(cmdsList)
	if len(cmdsList) == 0 {
		cmdsList = []storage.Command{{Pattern: "", Description: "No commands saved yet."}}
	}
	cmdsList = sortByWorkspace(cmdsList, m.cwd)
	m.list = NewListModel(cmdsList, m.cwd)
	m = m.resizePanes()
	return m
}

func (m Model) applyTagFilter() Model {
	filtered := storage.FilterByTag(m.allCmds, m.activeTag)

	m.total = len(filtered)
	if len(filtered) == 0 {
		filtered = []storage.Command{{Pattern: "", Description: "No commands with tag '" + m.activeTag + "'."}}
	}
	filtered = sortByWorkspace(filtered, m.cwd)
	m.list = NewListModel(filtered, m.cwd)
	m = m.resizePanes()
	return m
}

func (m Model) resizePanes() Model {
	if m.width == 0 || m.layout == LayoutTiny {
		return m
	}

	headerH := 2
	statusH := 1

	if m.layout == LayoutCompact {
		bodyH := m.height - headerH - statusH
		m.list.SetSize(m.width-2, bodyH)
		return m
	}

	// Full layout
	bodyH := m.height - headerH - statusH - 1
	listWidth := int(float64(m.width) * 0.40)
	m.list.SetSize(listWidth-4, bodyH-4)
	detailWidth := m.width - listWidth - 1
	m.detail.SetSize(detailWidth-2, bodyH-2)
	return m
}

func Start(store storage.Storage) (string, error) {
	p := tea.NewProgram(NewModel(store), tea.WithAltScreen(), tea.WithOutput(os.Stderr))
	m, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run tui: %w", err)
	}

	if model, ok := m.(Model); ok && model.Selected != "" {
		cmd := model.Selected
		if placeholders.HasPlaceholders(cmd) {
			// Auto-resolve {{branch}}, {{cwd}}, {{user}}, etc.
			var remaining []placeholders.Placeholder
			cmd, remaining = placeholders.AutoResolve(cmd)
			if len(remaining) > 0 {
				// Still has interactive placeholders — launch resolver
				rm := NewResolvingModelFromParsed(cmd, remaining)
				pr := NewResolverProgram(rm)
				res, err := pr.Run()
				if err != nil {
					return "", fmt.Errorf("failed to resolve placeholders: %w", err)
				}
				if resolved, ok := res.(ResolvingModel); ok && resolved.Done() {
					return resolved.Resolved(), nil
				}
				return "", nil
			}
		}
		return cmd, nil
	}

	return "", nil
}

func sortByWorkspace(cmds []storage.Command, cwd string) []storage.Command {
	if cwd == "" {
		return cmds
	}
	sort.SliceStable(cmds, func(i, j int) bool {
		mi := workspace.Matches(cwd, cmds[i].WorkspaceFilter)
		mj := workspace.Matches(cwd, cmds[j].WorkspaceFilter)
		if mi != mj {
			return mi
		}
		return false
	})
	return cmds
}

