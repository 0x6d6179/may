package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type ConfirmSpec struct {
	Title   string
	Default bool
}

type ConfirmModel struct {
	title string
	value bool
	width int
}

func newConfirm(spec ConfirmSpec) *ConfirmModel {
	return &ConfirmModel{
		title: spec.Title,
		value: spec.Default,
	}
}

func (m *ConfirmModel) Title() string { return m.title }

func (m *ConfirmModel) FooterHints() []Hint {
	return []Hint{
		{Key: "tab/←/→", Desc: "switch"},
		{Key: "enter", Desc: "confirm"},
		{Key: "esc", Desc: "abort"},
	}
}

func (m *ConfirmModel) SetSize(width, _ int) {
	m.width = width
}

func (m *ConfirmModel) Init() tea.Cmd {
	return nil
}

func (m *ConfirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "left", "right", "h", "l":
			m.value = !m.value
		case "y", "Y":
			m.value = true
		case "n", "N":
			m.value = false
		case "enter":
			return m, completeCmd(m.value)
		case "esc":
			return m, abortCmd()
		}
	}
	return m, nil
}

func (m *ConfirmModel) View() string {
	var b strings.Builder

	yes := "[yes]"
	no := "[no]"

	if m.value {
		b.WriteString(StyleSelected.Render(yes))
		b.WriteString("  ")
		b.WriteString(StyleHint.Render(no))
	} else {
		b.WriteString(StyleHint.Render(yes))
		b.WriteString("  ")
		b.WriteString(StyleSelected.Render(no))
	}

	return b.String()
}

func RunConfirm(opts RunOptions, spec ConfirmSpec) (bool, error) {
	m := newConfirm(spec)
	flow := &singleStepFlow{step: m}
	return RunFlow[bool](flow, opts)
}
