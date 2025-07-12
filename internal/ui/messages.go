package ui

import tea "github.com/charmbracelet/bubbletea"

type stepCompleteMsg struct{ Value any }

type stepAbortMsg struct{}

type asyncFinishedMsg struct {
	Value any
	Err   error
}

type OptionUpdateMsg struct {
	Index       int
	Description string
}

func completeCmd(v any) tea.Cmd {
	return func() tea.Msg { return stepCompleteMsg{Value: v} }
}

func abortCmd() tea.Cmd {
	return func() tea.Msg { return stepAbortMsg{} }
}

func asyncCmd[T any](fn func() (T, error)) tea.Cmd {
	return func() tea.Msg {
		v, err := fn()
		return asyncFinishedMsg{Value: v, Err: err}
	}
}
