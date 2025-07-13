package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	mayai "github.com/0x6d6179/may/internal/ai"
	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/ui"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

func fixSystemPrompt() string {
	return fmt.Sprintf(`You are a CLI assistant running on %s. The user ran a shell command that failed. Suggest a corrected command.

Return JSON with a single key:
- "command": the corrected shell command

Only return the JSON, nothing else.`, osName())
}

var fixSchema = json.RawMessage(`{
	"type": "object",
	"properties": {
		"command": {"type": "string"}
	},
	"required": ["command"],
	"additionalProperties": false
}`)

type fixSuggestion struct {
	Command string `json:"command"`
}

func newCmdFix(f *factory.Factory) *cobra.Command {
	var exitCode int

	cmd := &cobra.Command{
		Use:   "fix [failed command]",
		Short: "suggest a fix for a failed command",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return runFix(cmd.Context(), f, strings.Join(args, " "), exitCode)
		},
	}

	cmd.Flags().IntVar(&exitCode, "exit-code", 1, "exit code of the failed command")

	return cmd
}

func runFix(ctx context.Context, f *factory.Factory, failedCmd string, exitCode int) error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	if cfg.AI.APIKey == "" {
		return nil
	}

	provider := mayai.NewProviderFromConfig(&cfg.AI)
	model := cfg.AI.Model
	if model == "" {
		model = mayai.DefaultModel
	}

	prompt := fmt.Sprintf("command: %s\nexit code: %d", failedCmd, exitCode)

	timeoutCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	suggestion, err := fetchFixSuggestion(timeoutCtx, provider, model, prompt, f)
	if err != nil {
		return err
	}

	if suggestion.Command == failedCmd {
		return nil
	}

	opts := ui.RunOptions{In: f.IO.In, Out: f.IO.ErrOut}
	result, err := runFixPrompt(opts, suggestion.Command)
	if err != nil {
		return nil
	}

	if result {
		cmd := exec.Command("sh", "-c", suggestion.Command)
		cmd.Stdout = f.IO.Out
		cmd.Stderr = f.IO.ErrOut
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command failed: %w", err)
		}
	}

	return nil
}

func fetchFixSuggestion(ctx context.Context, provider mayai.Provider, model, prompt string, f *factory.Factory) (*fixSuggestion, error) {
	s := spinner.New()
	s.Spinner = ui.AISpinner

	p := tea.NewProgram(
		&fixLoaderModel{
			spinner:  s,
			provider: provider,
			model:    model,
			prompt:   prompt,
		},
		tea.WithOutput(f.IO.ErrOut),
	)

	m, err := p.Run()
	if err != nil {
		return nil, err
	}

	loader := m.(*fixLoaderModel)
	if loader.err != nil {
		return nil, loader.err
	}

	return loader.result, nil
}

type fixLoaderModel struct {
	spinner  spinner.Model
	provider mayai.Provider
	model    string
	prompt   string
	result   *fixSuggestion
	err      error
	done     bool
}

func (m *fixLoaderModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			req := mayai.ChatRequest{
				Model: m.model,
				Messages: []mayai.Message{
					{Role: "system", Content: fixSystemPrompt()},
					{Role: "user", Content: m.prompt},
				},
				ResponseFormat: &mayai.ResponseFormat{
					Type: "json_schema",
					JSONSchema: &mayai.JSONSchema{
						Name:   "fix_suggestion",
						Strict: true,
						Schema: fixSchema,
					},
				},
			}
			resp, err := m.provider.ChatComplete(context.Background(), req)
			if err != nil {
				return fixDoneMsg{err: err}
			}
			var s fixSuggestion
			if err := json.Unmarshal([]byte(resp.Content), &s); err != nil {
				return fixDoneMsg{err: fmt.Errorf("parse fix: %w", err)}
			}
			return fixDoneMsg{suggestion: &s}
		},
	)
}

type fixDoneMsg struct {
	suggestion *fixSuggestion
	err        error
}

func (m *fixLoaderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case fixDoneMsg:
		m.done = true
		m.result = msg.suggestion
		m.err = msg.err
		return m, tea.Quit
	case spinner.TickMsg:
		if m.done {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.err = ui.ErrAborted
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *fixLoaderModel) View() string {
	if m.done {
		return ""
	}
	return m.spinner.View() + " " + ui.StyleTitle.Render("may ai") + " " + nextLoaderMessage() + "\n"
}

type fixPromptModel struct {
	command string
	value   bool
}

func (m *fixPromptModel) Init() tea.Cmd { return nil }

func (m *fixPromptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			m.value = true
			return m, tea.Quit
		case "n", "N", "enter", "esc":
			m.value = false
			return m, tea.Quit
		case "ctrl+c":
			m.value = false
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *fixPromptModel) View() string {
	prefix := ui.StyleTitle.Render(ui.AILogo + " may ai")
	cmd := ui.StyleSelected.Render(m.command)
	hint := ui.StyleHint.Render("[y/N]")
	return prefix + " did you mean: " + cmd + " " + hint + "\n"
}

func runFixPrompt(opts ui.RunOptions, command string) (bool, error) {
	m := &fixPromptModel{command: command}

	var progOpts []tea.ProgramOption
	if opts.Out != nil {
		progOpts = append(progOpts, tea.WithOutput(opts.Out))
	}
	if opts.In != nil {
		progOpts = append(progOpts, tea.WithInput(opts.In))
	}

	p := tea.NewProgram(m, progOpts...)
	result, err := p.Run()
	if err != nil {
		return false, err
	}

	return result.(*fixPromptModel).value, nil
}
