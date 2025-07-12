package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SelectSpec[T comparable] struct {
	Title   string
	Options []Option[T]
	Height  int
	InitCmd tea.Cmd
}

type SelectModel[T comparable] struct {
	title      string
	options    []Option[T]
	filtered   []int
	cursor     int
	offset     int
	maxVisible int
	filtering  bool
	filter     textinput.Model
	width      int
	selected   bool
	spinner    spinner.Model
	initCmd    tea.Cmd
}

func newSelect[T comparable](spec SelectSpec[T]) *SelectModel[T] {
	height := spec.Height
	if height <= 0 {
		height = 10
	}

	ti := textinput.New()
	ti.Prompt = "/ "
	ti.PromptStyle = StyleHint
	ti.TextStyle = lipgloss.NewStyle().Foreground(lavender)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = StyleHint

	indices := make([]int, len(spec.Options))
	for i := range spec.Options {
		indices[i] = i
	}

	return &SelectModel[T]{
		title:      spec.Title,
		options:    spec.Options,
		filtered:   indices,
		maxVisible: height,
		filter:     ti,
		spinner:    s,
		initCmd:    spec.InitCmd,
	}
}

func (m *SelectModel[T]) Title() string { return m.title }

func (m *SelectModel[T]) FooterHints() []Hint {
	if m.filtering {
		return []Hint{
			{Key: "esc", Desc: "clear filter"},
			{Key: "enter", Desc: "select"},
		}
	}
	return []Hint{
		{Key: "↑/↓", Desc: "navigate"},
		{Key: "/", Desc: "filter"},
		{Key: "enter", Desc: "select"},
		{Key: "esc", Desc: "abort"},
	}
}

func (m *SelectModel[T]) SetSize(width, _ int) {
	m.width = width
	m.filter.Width = width - 4
}

func (m *SelectModel[T]) Init() tea.Cmd {
	return tea.Batch(m.initCmd, m.spinner.Tick)
}

func (m *SelectModel[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case OptionUpdateMsg:
		if msg.Index >= 0 && msg.Index < len(m.options) {
			m.options[msg.Index].Description = msg.Description
			m.options[msg.Index].Loading = false
		}
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		switch {
		case m.filtering:
			return m.updateFiltering(msg)
		default:
			return m.updateNormal(msg)
		}
	}
	return m, nil
}

func (m *SelectModel[T]) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.moveCursor(-1)
	case "down", "j":
		m.moveCursor(1)
	case "enter":
		if len(m.filtered) > 0 {
			m.selected = true
			idx := m.filtered[m.cursor]
			return m, completeCmd(m.options[idx].Value)
		}
	case "/":
		m.filtering = true
		m.filter.Focus()
		return m, textinput.Blink
	case "esc":
		return m, abortCmd()
	}
	return m, nil
}

func (m *SelectModel[T]) updateFiltering(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.filtering = false
		m.filter.Reset()
		m.filter.Blur()
		m.applyFilter()
		return m, nil
	case "enter":
		if len(m.filtered) > 0 {
			m.selected = true
			idx := m.filtered[m.cursor]
			return m, completeCmd(m.options[idx].Value)
		}
		return m, nil
	case "up", "ctrl+p":
		m.moveCursor(-1)
		return m, nil
	case "down", "ctrl+n":
		m.moveCursor(1)
		return m, nil
	}

	var cmd tea.Cmd
	m.filter, cmd = m.filter.Update(msg)
	m.applyFilter()
	return m, cmd
}

func (m *SelectModel[T]) applyFilter() {
	query := strings.ToLower(m.filter.Value())
	if query == "" {
		m.filtered = make([]int, len(m.options))
		for i := range m.options {
			m.filtered[i] = i
		}
	} else {
		m.filtered = m.filtered[:0]
		for i, opt := range m.options {
			if strings.Contains(strings.ToLower(opt.Label), query) {
				m.filtered = append(m.filtered, i)
			}
		}
	}
	m.cursor = 0
	m.offset = 0
}

func (m *SelectModel[T]) moveCursor(delta int) {
	if len(m.filtered) == 0 {
		return
	}
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = len(m.filtered) - 1
	}
	if m.cursor >= len(m.filtered) {
		m.cursor = 0
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+m.maxVisible {
		m.offset = m.cursor - m.maxVisible + 1
	}
}

func (m *SelectModel[T]) View() string {
	var b strings.Builder

	if m.filtering {
		b.WriteString(m.filter.View())
		b.WriteString("\n")
	}

	end := m.offset + m.maxVisible
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	visible := m.filtered[m.offset:end]
	for i, idx := range visible {
		absIdx := m.offset + i
		opt := m.options[idx]

		if absIdx == m.cursor {
			b.WriteString(StyleSelector.Render("→ "))
			b.WriteString(StyleSelected.Render(opt.Label))
		} else {
			b.WriteString("  ")
			b.WriteString(opt.Label)
		}

		if opt.Description != "" || opt.Loading {
			b.WriteString(StyleHint.Render(" · "))
			var desc strings.Builder
			if opt.Loading {
				desc.WriteString(m.spinner.View())
				desc.WriteString(" ")
			}
			desc.WriteString(opt.Description)
			b.WriteString(StyleDescription.Render(desc.String()))
		}

		if i < len(visible)-1 {
			b.WriteString("\n")
		}
	}

	if len(m.filtered) == 0 {
		b.WriteString(StyleHint.Render("no matches"))
	}

	return b.String()
}

func NewSelectStep[T comparable](spec SelectSpec[T]) *SelectModel[T] {
	return newSelect(spec)
}

func RunSelect[T comparable](opts RunOptions, spec SelectSpec[T]) (T, error) {
	m := newSelect(spec)
	flow := &singleStepFlow{step: m}
	return RunFlow[T](flow, opts)
}
