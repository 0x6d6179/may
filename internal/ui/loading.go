package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type LoadingSpec[T any] struct {
	Title         string
	Label         string
	Task          func() (T, error)
	CustomSpinner *spinner.Spinner
}

type LoadingModel[T any] struct {
	title   string
	label   string
	task    func() (T, error)
	spinner spinner.Model
	done    bool
	err     error
	width   int
}

func NewLoading[T any](spec LoadingSpec[T]) *LoadingModel[T] {
	s := spinner.New()
	if spec.CustomSpinner != nil {
		s.Spinner = *spec.CustomSpinner
	} else {
		s.Spinner = spinner.MiniDot
	}

	return &LoadingModel[T]{
		title:   spec.Title,
		label:   spec.Label,
		task:    spec.Task,
		spinner: s,
	}
}

func (m *LoadingModel[T]) Title() string { return "" }

func (m *LoadingModel[T]) FooterHints() []Hint {
	return nil
}

func (m *LoadingModel[T]) SetSize(width, _ int) {
	m.width = width
}

func (m *LoadingModel[T]) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, asyncCmd(m.task))
}

func (m *LoadingModel[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case asyncFinishedMsg:
		m.done = true
		if msg.Err != nil {
			m.err = msg.Err
			return m, completeCmd(msg.Err)
		}
		return m, completeCmd(msg.Value)

	case spinner.TickMsg:
		if m.done {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *LoadingModel[T]) View() string {
	if m.done {
		return ""
	}
	return m.spinner.View() + " " + StyleTitle.Render(m.title) + " " + m.label
}
