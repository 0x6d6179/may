package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type MultiSelectSpec[T comparable] struct {
	Title    string
	Options  []Option[T]
	Defaults []T
	Height   int
}

type MultiSelectModel[T comparable] struct {
	title      string
	options    []Option[T]
	checked    []bool
	cursor     int
	offset     int
	maxVisible int
	width      int
}

func newMultiSelect[T comparable](spec MultiSelectSpec[T]) *MultiSelectModel[T] {
	height := spec.Height
	if height <= 0 {
		height = len(spec.Options)
		if height > 10 {
			height = 10
		}
		if height == 0 {
			height = 5
		}
	}

	checked := make([]bool, len(spec.Options))
	for _, def := range spec.Defaults {
		for i, opt := range spec.Options {
			if opt.Value == def {
				checked[i] = true
				break
			}
		}
	}

	return &MultiSelectModel[T]{
		title:      spec.Title,
		options:    spec.Options,
		checked:    checked,
		maxVisible: height,
	}
}

func (m *MultiSelectModel[T]) Title() string { return m.title }

func (m *MultiSelectModel[T]) FooterHints() []Hint {
	return []Hint{
		{Key: "↑/↓", Desc: "navigate"},
		{Key: "space", Desc: "toggle"},
		{Key: "enter", Desc: "confirm"},
		{Key: "esc", Desc: "abort"},
	}
}

func (m *MultiSelectModel[T]) SetSize(width, height int) {
	m.width = width
	if height > 0 {
		visible := height - 6
		if visible < 3 {
			visible = 3
		}
		m.maxVisible = visible
	}
}

func (m *MultiSelectModel[T]) Init() tea.Cmd {
	return nil
}

func (m *MultiSelectModel[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.moveCursor(-1)
		case "down", "j":
			m.moveCursor(1)
		case " ", "x":
			if len(m.options) > 0 {
				m.checked[m.cursor] = !m.checked[m.cursor]
			}
		case "enter":
			return m, completeCmd(m.values())
		case "esc":
			return m, abortCmd()
		}
	}
	return m, nil
}

func (m *MultiSelectModel[T]) moveCursor(delta int) {
	if len(m.options) == 0 {
		return
	}
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = len(m.options) - 1
	}
	if m.cursor >= len(m.options) {
		m.cursor = 0
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+m.maxVisible {
		m.offset = m.cursor - m.maxVisible + 1
	}
}

func (m *MultiSelectModel[T]) values() []T {
	var result []T
	for i, opt := range m.options {
		if m.checked[i] {
			result = append(result, opt.Value)
		}
	}
	return result
}

func (m *MultiSelectModel[T]) View() string {
	var b strings.Builder

	end := m.offset + m.maxVisible
	if end > len(m.options) {
		end = len(m.options)
	}

	if m.offset > 0 {
		b.WriteString(StyleHint.Render(fmt.Sprintf("  ↑ %d more", m.offset)))
		b.WriteString("\n")
	}

	for i := m.offset; i < end; i++ {
		opt := m.options[i]

		var checkbox string
		if m.checked[i] {
			checkbox = StyleSelected.Render("[x]") + " "
		} else {
			checkbox = "[ ] "
		}

		if i == m.cursor {
			b.WriteString(StyleSelector.Render("→ "))
			b.WriteString(checkbox)
			b.WriteString(StyleSelected.Render(opt.Label))
		} else {
			b.WriteString("  ")
			b.WriteString(checkbox)
			b.WriteString(opt.Label)
		}

		if opt.Description != "" {
			b.WriteString("  ")
			b.WriteString(StyleDescription.Render(opt.Description))
		}

		b.WriteString("\n")
	}

	remaining := len(m.options) - end
	if remaining > 0 {
		b.WriteString(StyleHint.Render(fmt.Sprintf("  ↓ %d more", remaining)))
	}

	return b.String()
}

func NewMultiSelectStep[T comparable](spec MultiSelectSpec[T]) *MultiSelectModel[T] {
	return newMultiSelect(spec)
}

func RunMultiSelect[T comparable](opts RunOptions, spec MultiSelectSpec[T]) ([]T, error) {
	m := newMultiSelect(spec)
	flow := &singleStepFlow{step: m}
	return RunFlow[[]T](flow, opts)
}
