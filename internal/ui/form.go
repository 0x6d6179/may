package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type InputField struct {
	Key         string
	Label       string
	Placeholder string
	Default     string
	Password    bool
	Validate    func(string) error
}

type FormSpec struct {
	Title  string
	Fields []InputField
}

type FormModel struct {
	title  string
	fields []InputField
	inputs []textinput.Model
	focus  int
	errs   []error
	width  int
}

func newForm(spec FormSpec) *FormModel {
	inputs := make([]textinput.Model, len(spec.Fields))
	errs := make([]error, len(spec.Fields))

	for i, f := range spec.Fields {
		ti := textinput.New()
		ti.Placeholder = f.Placeholder
		ti.SetValue(f.Default)
		if f.Password {
			ti.EchoMode = textinput.EchoPassword
			ti.EchoCharacter = '•'
		}
		inputs[i] = ti
	}

	if len(inputs) > 0 {
		inputs[0].Focus()
	}

	return &FormModel{
		title:  spec.Title,
		fields: spec.Fields,
		inputs: inputs,
		errs:   errs,
	}
}

func (m *FormModel) Title() string { return m.title }

func (m *FormModel) FooterHints() []Hint {
	return []Hint{
		{Key: "tab", Desc: "next field"},
		{Key: "shift+tab", Desc: "prev field"},
		{Key: "enter", Desc: "submit"},
		{Key: "esc", Desc: "abort"},
	}
}

func (m *FormModel) SetSize(width, _ int) {
	m.width = width
	for i := range m.inputs {
		m.inputs[i].Width = width - 4
	}
}

func (m *FormModel) Init() tea.Cmd {
	if len(m.inputs) > 0 {
		return textinput.Blink
	}
	return nil
}

func (m *FormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			return m, m.focusNext()
		case "shift+tab", "up":
			return m, m.focusPrev()
		case "enter":
			if m.focus == len(m.inputs)-1 {
				return m, m.submit()
			}
			return m, m.focusNext()
		case "esc":
			return m, abortCmd()
		}
	}

	if m.focus < len(m.inputs) {
		var cmd tea.Cmd
		m.inputs[m.focus], cmd = m.inputs[m.focus].Update(msg)
		m.errs[m.focus] = nil
		return m, cmd
	}
	return m, nil
}

func (m *FormModel) submit() tea.Cmd {
	hasErr := false
	for i, f := range m.fields {
		if f.Validate != nil {
			if err := f.Validate(m.inputs[i].Value()); err != nil {
				m.errs[i] = err
				hasErr = true
			}
		}
	}
	if hasErr {
		return nil
	}

	result := make(map[string]string, len(m.fields))
	for i, f := range m.fields {
		result[f.Key] = m.inputs[i].Value()
	}
	return completeCmd(result)
}

func (m *FormModel) focusNext() tea.Cmd {
	m.inputs[m.focus].Blur()
	m.focus++
	if m.focus >= len(m.inputs) {
		m.focus = 0
	}
	m.inputs[m.focus].Focus()
	return textinput.Blink
}

func (m *FormModel) focusPrev() tea.Cmd {
	m.inputs[m.focus].Blur()
	m.focus--
	if m.focus < 0 {
		m.focus = len(m.inputs) - 1
	}
	m.inputs[m.focus].Focus()
	return textinput.Blink
}

func (m *FormModel) View() string {
	var b strings.Builder

	for i, f := range m.fields {
		if i > 0 {
			b.WriteString("\n")
		}

		label := f.Label
		if label == "" {
			label = f.Key
		}

		if i == m.focus {
			b.WriteString(StyleSelected.Render(label))
		} else {
			b.WriteString(StyleHint.Render(label))
		}
		b.WriteString("\n")
		b.WriteString(m.inputs[i].View())

		if m.errs[i] != nil {
			b.WriteString("\n")
			b.WriteString(StyleError.Render(m.errs[i].Error()))
		}
	}

	return b.String()
}

func NewFormStep(spec FormSpec) *FormModel {
	return newForm(spec)
}

func RunForm(opts RunOptions, spec FormSpec) (map[string]string, error) {
	m := newForm(spec)
	flow := &singleStepFlow{step: m}
	return RunFlow[map[string]string](flow, opts)
}
