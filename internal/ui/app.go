package ui

import (
	"fmt"
	"io"

	tea "github.com/charmbracelet/bubbletea"
)

type Step interface {
	tea.Model
	Title() string
	FooterHints() []Hint
	SetSize(width, height int)
}

type Flow interface {
	Start() Step
	Next(result any) (Step, bool, error)
}

type RunOptions struct {
	In  io.Reader
	Out io.Writer
}

type App struct {
	flow    Flow
	step    Step
	result  any
	err     error
	width   int
	quitting bool
}

func newApp(flow Flow, width int) *App {
	step := flow.Start()
	step.SetSize(width, 0)
	return &App{
		flow:  flow,
		step:  step,
		width: width,
	}
}

func (a *App) Init() tea.Cmd {
	return a.step.Init()
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			a.err = ErrAborted
			a.quitting = true
			return a, tea.Quit
		}

	case stepCompleteMsg:
		next, done, err := a.flow.Next(msg.Value)
		if err != nil {
			a.err = err
			a.quitting = true
			return a, tea.Quit
		}
		if done {
			a.result = msg.Value
			a.quitting = true
			return a, tea.Quit
		}
		a.step = next
		a.step.SetSize(a.width, 0)
		return a, a.step.Init()

	case stepAbortMsg:
		a.err = ErrAborted
		a.quitting = true
		return a, tea.Quit

	case tea.WindowSizeMsg:
		a.width = msg.Width
		if a.width > 80 {
			a.width = 80
		}
		a.step.SetSize(a.width, msg.Height)
	}

	var cmd tea.Cmd
	updated, cmd := a.step.Update(msg)
	if s, ok := updated.(Step); ok {
		a.step = s
	}
	return a, cmd
}

func (a *App) View() string {
	if a.quitting {
		return ""
	}
	return RenderFrame(
		a.step.Title(),
		a.step.View(),
		a.step.FooterHints(),
		a.width,
	)
}

func RunFlow[T any](flow Flow, opts RunOptions) (T, error) {
	var zero T
	width := TermWidth()
	app := newApp(flow, width)

	var progOpts []tea.ProgramOption
	if opts.Out != nil {
		progOpts = append(progOpts, tea.WithOutput(opts.Out))
	}
	if opts.In != nil {
		progOpts = append(progOpts, tea.WithInput(opts.In))
	}

	p := tea.NewProgram(app, progOpts...)
	m, err := p.Run()
	if err != nil {
		return zero, fmt.Errorf("run flow: %w", err)
	}

	a := m.(*App)
	if a.err != nil {
		return zero, a.err
	}

	if a.result == nil {
		return zero, nil
	}

	v, ok := a.result.(T)
	if !ok {
		return zero, fmt.Errorf("unexpected result type %T", a.result)
	}
	return v, nil
}

type singleStepFlow struct {
	step Step
}

func (f *singleStepFlow) Start() Step {
	return f.step
}

func (f *singleStepFlow) Next(_ any) (Step, bool, error) {
	return nil, true, nil
}
