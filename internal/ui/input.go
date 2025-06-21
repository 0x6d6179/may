package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type InputSpec struct {
	Title       string
	Placeholder string
	Default     string
	Password    bool
	Validate    func(string) error
}

type InputModel struct {
	title    string
	input    textinput.Model
	validate func(string) error
	err      error
	width    int
}

func newInput(spec InputSpec) *InputModel {
	ti := textinput.New()
	ti.Placeholder = spec.Placeholder
	ti.SetValue(spec.Default)
	ti.Focus()

	if spec.Password {
		ti.EchoMode = textinput.EchoPassword
		ti.EchoCharacter = '•'
	}

	return &InputModel{
		title:    spec.Title,
		input:    ti,
		validate: spec.Validate,
	}
}

func (m *InputModel) Title() string { return m.title }

func (m *InputModel) FooterHints() []Hint {
	return []Hint{
		{Key: "enter", Desc: "submit"},
		{Key: "esc", Desc: "abort"},
	}
}

func (m *InputModel) SetSize(width, _ int) {
	m.width = width
	m.input.Width = width - 4
}

func (m *InputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *InputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			val := m.input.Value()
			if m.validate != nil {
				if err := m.validate(val); err != nil {
					m.err = err
					return m, nil
				}
			}
			m.err = nil
			return m, completeCmd(val)
		case "esc":
			return m, abortCmd()
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.err = nil
	return m, cmd
}

func (m *InputModel) View() string {
	var b strings.Builder

	b.WriteString(m.input.View())

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(StyleError.Render(m.err.Error()))
	}

	return b.String()
}

func NewInputStep(spec InputSpec) *InputModel {
	return newInput(spec)
}

func RunInput(opts RunOptions, spec InputSpec) (string, error) {
	m := newInput(spec)
	flow := &singleStepFlow{step: m}
	return RunFlow[string](flow, opts)
}
