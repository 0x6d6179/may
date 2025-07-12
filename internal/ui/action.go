package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SafetyLevel string

const (
	SafetySafe        SafetyLevel = "safe"
	SafetyCautious    SafetyLevel = "cautious"
	SafetyRisky       SafetyLevel = "risky"
	SafetyDestructive SafetyLevel = "destructive"
)

func SafetyStyle(level SafetyLevel) lipgloss.Style {
	switch level {
	case SafetySafe:
		return StyleSafetySafe
	case SafetyCautious:
		return StyleSafetyCautious
	case SafetyRisky:
		return StyleSafetyRisky
	case SafetyDestructive:
		return StyleSafetyDestructive
	default:
		return StyleDescription
	}
}

type ActionSpec struct {
	Command   string
	Safety    SafetyLevel
	Breakdown string
}

type ActionResult string

const (
	ActionApprove  ActionResult = "approve"
	ActionReject   ActionResult = "reject"
	ActionModify   ActionResult = "modify"
	ActionReprompt ActionResult = "reprompt"
)

var actionLabels = []struct {
	result ActionResult
	label  string
}{
	{ActionApprove, "[approve]"},
	{ActionReject, "[reject]"},
	{ActionModify, "[modify]"},
	{ActionReprompt, "[reprompt]"},
}

type ActionModel struct {
	command   string
	safety    SafetyLevel
	breakdown string
	cursor    int
	width     int
}

func NewActionStep(spec ActionSpec) *ActionModel {
	return &ActionModel{
		command:   spec.Command,
		safety:    spec.Safety,
		breakdown: spec.Breakdown,
	}
}

func (m *ActionModel) Title() string { return "" }

func (m *ActionModel) FooterHints() []Hint {
	return []Hint{
		{Key: "tab/←/→", Desc: "switch"},
		{Key: "enter", Desc: "select"},
		{Key: "esc", Desc: "reject"},
	}
}

func (m *ActionModel) SetSize(width, _ int) {
	m.width = width
}

func (m *ActionModel) Init() tea.Cmd {
	return nil
}

func (m *ActionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "right", "l":
			m.cursor = (m.cursor + 1) % len(actionLabels)
		case "left", "h":
			m.cursor = (m.cursor - 1 + len(actionLabels)) % len(actionLabels)
		case "enter":
			return m, completeCmd(actionLabels[m.cursor].result)
		case "esc":
			return m, completeCmd(ActionReject)
		}
	}
	return m, nil
}

func (m *ActionModel) View() string {
	var b strings.Builder

	b.WriteString(StyleTitle.Render(AILogo + " may ai"))
	b.WriteString(" wants to run:")
	b.WriteString("\n")
	b.WriteString(StyleSelected.Render(m.command))
	b.WriteString("\n")

	safetyStyle := SafetyStyle(m.safety)
	b.WriteString(StyleDescription.Render("- safety: "))
	b.WriteString(safetyStyle.Render(string(m.safety)))
	b.WriteString("\n")

	b.WriteString(StyleDescription.Render(wrapWithPrefix("- ", m.breakdown, m.width)))
	b.WriteString("\n")

	if m.safety == SafetyRisky || m.safety == SafetyDestructive {
		b.WriteString("\n")
		warning := "⚠ this command is potentially destructive and may cause irreversible changes"
		if m.safety == SafetyDestructive {
			warning = "⚠ this command is destructive and may cause irreversible damage"
		}
		b.WriteString(StyleWarning.Render(wrapText(warning, m.width)))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	for i, a := range actionLabels {
		if i > 0 {
			b.WriteString("  ")
		}
		if i == m.cursor {
			b.WriteString(StyleSelected.Render(a.label))
		} else {
			b.WriteString(StyleHint.Render(a.label))
		}
	}

	return b.String()
}

func wrapText(text string, width int) string {
	if width <= 0 || len(text) <= width {
		return text
	}
	return wrapWithPrefix("", text, width)
}

func wrapWithPrefix(prefix, text string, width int) string {
	if width <= 0 {
		return prefix + text
	}

	indent := strings.Repeat(" ", len(prefix))
	maxLine := width - len(prefix)
	if maxLine <= 0 {
		return prefix + text
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return prefix
	}

	var lines []string
	current := words[0]

	for _, word := range words[1:] {
		if len(current)+1+len(word) > maxLine {
			lines = append(lines, current)
			current = word
		} else {
			current += " " + word
		}
	}
	lines = append(lines, current)

	var b strings.Builder
	for i, line := range lines {
		if i == 0 {
			b.WriteString(prefix + line)
		} else {
			b.WriteString("\n" + indent + line)
		}
	}
	return b.String()
}

func RunAction(opts RunOptions, spec ActionSpec) (ActionResult, error) {
	m := NewActionStep(spec)
	flow := &singleStepFlow{step: m}
	return RunFlow[ActionResult](flow, opts)
}
