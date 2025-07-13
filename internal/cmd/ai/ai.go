package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	mayai "github.com/0x6d6179/may/internal/ai"
	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/ui"
	"github.com/spf13/cobra"
)

const maxValidationRetries = 3

func osName() string {
	switch runtime.GOOS {
	case "darwin":
		return "macOS"
	case "linux":
		return "Linux"
	case "windows":
		return "Windows"
	default:
		return runtime.GOOS
	}
}

func freeformSystemPrompt(shell string) string {
	return fmt.Sprintf(`You are a CLI assistant running on %s with %s. The user describes what they want to do. Suggest a single shell command.

Rules:
- Output MUST be a single line. Chain multiple commands with && or ; or | as appropriate for %s syntax.
- NEVER output multiline commands or newlines in the command string.
- NEVER suggest optional follow-up steps. Only suggest exactly what the user asked for.
- The breakdown must only describe what the suggested command does. Do not mention anything not in the command.

Return JSON with keys:
- "command": the shell command (single line, proper %s syntax)
- "safety": one of "safe", "cautious", "risky", "destructive"
  - safe: read-only or harmless (ls, cat, echo, pwd, which, grep)
  - cautious: modifying but recoverable (git commit, cp, mkdir, touch)
  - risky: potentially dangerous, hard to reverse (sudo, rm files, chmod, chown)
  - destructive: irreversible damage (rm -rf, force push, drop table, mkfs, dd)
- "breakdown": one sentence describing only what the command does

Only return the JSON, nothing else.`, osName(), shell, shell, shell)
}

var freeformSchema = json.RawMessage(`{
	"type": "object",
	"properties": {
		"command": {"type": "string"},
		"safety": {"type": "string", "enum": ["safe", "cautious", "risky", "destructive"]},
		"breakdown": {"type": "string"}
	},
	"required": ["command", "safety", "breakdown"],
	"additionalProperties": false
}`)

type commandSuggestion struct {
	Command   string `json:"command"`
	Safety    string `json:"safety"`
	Breakdown string `json:"breakdown"`
}

func NewCmdAi(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ai [prompt]",
		Aliases: []string{"aioh"},
		Short:   "ai assistant",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return runFreeform(cmd.Context(), f, strings.Join(args, " "))
		},
	}

	cmd.AddCommand(newCmdConfigure(f))
	cmd.AddCommand(newCmdCommit(f))
	cmd.AddCommand(newCmdFix(f))

	return cmd
}

func runFreeform(ctx context.Context, f *factory.Factory, prompt string) error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	if cfg.AI.APIKey == "" {
		return fmt.Errorf("ai is not configured — run: may ai configure")
	}

	provider := mayai.NewProviderFromConfig(&cfg.AI)
	model := cfg.AI.Model
	if model == "" {
		model = mayai.DefaultModel
	}

	shell := detectShell()
	cwd, _ := os.Getwd()

	userMsg := buildUserMessage(prompt, shell, cwd)

	opts := ui.RunOptions{In: f.IO.In, Out: f.IO.ErrOut}

	currentPrompt := userMsg
	var suggestion *commandSuggestion

	for attempt := range maxValidationRetries {
		suggestion, err = fetchSuggestion(ctx, provider, model, shell, currentPrompt, opts)
		if err != nil {
			return err
		}

		missing := findMissingCommands(suggestion.Command)
		if len(missing) == 0 {
			break
		}

		if attempt == maxValidationRetries-1 {
			fmt.Fprintf(f.IO.ErrOut, "no suitable command found for this system\n")
			return nil
		}

		currentPrompt = userMsg + fmt.Sprintf(
			"\nthe command %q does not exist on this system. suggest an alternative using commands that are available.",
			strings.Join(missing, ", "),
		)
	}

	return handleSuggestion(ctx, f, provider, model, shell, prompt, suggestion, opts)
}

func fetchSuggestion(ctx context.Context, provider mayai.Provider, model, shell, prompt string, opts ui.RunOptions) (*commandSuggestion, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	result, err := ui.RunFlow[any](&freeformLoadFlow{
		ctx:      timeoutCtx,
		provider: provider,
		model:    model,
		shell:    shell,
		prompt:   prompt,
	}, opts)
	if errors.Is(err, ui.ErrAborted) {
		return nil, ui.ErrAborted
	}
	if err != nil {
		return nil, fmt.Errorf("ai request failed: %w", err)
	}

	switch v := result.(type) {
	case *commandSuggestion:
		return v, nil
	case error:
		return nil, fmt.Errorf("ai request failed: %w", v)
	}
	return nil, fmt.Errorf("unexpected response")
}

func handleSuggestion(ctx context.Context, f *factory.Factory, provider mayai.Provider, model, shell, originalPrompt string, suggestion *commandSuggestion, opts ui.RunOptions) error {
	for {
		action, err := ui.RunAction(opts, ui.ActionSpec{
			Command:   suggestion.Command,
			Safety:    ui.SafetyLevel(suggestion.Safety),
			Breakdown: suggestion.Breakdown,
		})
		if errors.Is(err, ui.ErrAborted) {
			return nil
		}
		if err != nil {
			return err
		}

		switch action {
		case ui.ActionApprove:
			if needsConfirmation(suggestion.Safety) {
				confirmed, cerr := ui.RunConfirm(opts, ui.ConfirmSpec{
					Title:   "are you sure you want to run this?",
					Default: false,
				})
				if errors.Is(cerr, ui.ErrAborted) {
					return nil
				}
				if cerr != nil {
					return cerr
				}
				if !confirmed {
					continue
				}
			}
			return executeCommand(f, suggestion.Command)

		case ui.ActionReject:
			return nil

		case ui.ActionModify:
			modified, err := ui.RunInput(opts, ui.InputSpec{
				Title:   "modify command",
				Default: suggestion.Command,
			})
			if errors.Is(err, ui.ErrAborted) {
				return nil
			}
			if err != nil {
				return err
			}
			return executeCommand(f, modified)

		case ui.ActionReprompt:
			extra, err := ui.RunInput(opts, ui.InputSpec{
				Title:       "additional context",
				Placeholder: "be more specific…",
			})
			if errors.Is(err, ui.ErrAborted) {
				return nil
			}
			if err != nil {
				return err
			}

			newPrompt := originalPrompt + "\n" + extra
			suggestion, err = fetchSuggestion(ctx, provider, model, shell, newPrompt, opts)
			if errors.Is(err, ui.ErrAborted) {
				return nil
			}
			if err != nil {
				return err
			}
		}
	}
}

func needsConfirmation(safety string) bool {
	return safety == string(ui.SafetyRisky) || safety == string(ui.SafetyDestructive)
}

func executeCommand(f *factory.Factory, command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = f.IO.Out
	cmd.Stderr = f.IO.ErrOut
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}
	return nil
}

func findMissingCommands(command string) []string {
	var missing []string
	seen := make(map[string]bool)

	for _, name := range extractCommands(command) {
		if seen[name] {
			continue
		}
		seen[name] = true
		if _, err := exec.LookPath(name); err != nil {
			missing = append(missing, name)
		}
	}

	return missing
}

func extractCommands(command string) []string {
	var cmds []string

	segments := splitShellOperators(command)
	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		if name := extractPrimaryCommand(seg); name != "" {
			cmds = append(cmds, name)
		}
	}

	return cmds
}

func splitShellOperators(command string) []string {
	result := []string{command}
	for _, op := range []string{"&&", "||", "|", ";"} {
		var next []string
		for _, s := range result {
			next = append(next, strings.Split(s, op)...)
		}
		result = next
	}
	return result
}

func extractPrimaryCommand(segment string) string {
	for _, token := range strings.Fields(segment) {
		if strings.Contains(token, "=") {
			continue
		}
		token = strings.TrimLeft(token, "(")
		if token == "sudo" {
			continue
		}
		return token
	}
	return ""
}

func detectShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return "bash"
	}
	base := shell[strings.LastIndex(shell, "/")+1:]
	switch base {
	case "zsh", "bash", "fish":
		return base
	default:
		return "bash"
	}
}

func buildUserMessage(prompt, shell, cwd string) string {
	var b strings.Builder
	b.WriteString(prompt)
	b.WriteString(fmt.Sprintf("\n\n[context: shell=%s, cwd=%s]", shell, cwd))

	history := lastShellHistory(shell, 5)
	if history != "" {
		b.WriteString(fmt.Sprintf("\n[recent commands:\n%s]", history))
	}

	return b.String()
}

func lastShellHistory(shell string, n int) string {
	var cmd *exec.Cmd
	switch shell {
	case "zsh":
		cmd = exec.Command("zsh", "-c", fmt.Sprintf("fc -l -%d 2>/dev/null | tail -%d", n, n))
	case "bash":
		cmd = exec.Command("bash", "-c", fmt.Sprintf("history %d 2>/dev/null | tail -%d", n, n))
	case "fish":
		cmd = exec.Command("fish", "-c", fmt.Sprintf("builtin history | head -%d", n))
	default:
		return ""
	}

	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

type freeformLoadFlow struct {
	ctx      context.Context
	provider mayai.Provider
	model    string
	shell    string
	prompt   string
}

func (f *freeformLoadFlow) Start() ui.Step {
	return ui.NewLoading[*commandSuggestion](ui.LoadingSpec[*commandSuggestion]{
		Title:         "may ai",
		Label:         nextLoaderMessage(),
		Task:          func() (*commandSuggestion, error) { return f.callAI() },
		CustomSpinner: &ui.AISpinner,
	})
}

func (f *freeformLoadFlow) Next(_ any) (ui.Step, bool, error) {
	return nil, true, nil
}

func (f *freeformLoadFlow) callAI() (*commandSuggestion, error) {
	req := mayai.ChatRequest{
		Model: f.model,
		Messages: []mayai.Message{
			{Role: "system", Content: freeformSystemPrompt(f.shell)},
			{Role: "user", Content: f.prompt},
		},
		ResponseFormat: &mayai.ResponseFormat{
			Type: "json_schema",
			JSONSchema: &mayai.JSONSchema{
				Name:   "command_suggestion",
				Strict: true,
				Schema: freeformSchema,
			},
		},
	}

	resp, err := f.provider.ChatComplete(f.ctx, req)
	if err != nil {
		return nil, err
	}

	var suggestion commandSuggestion
	if err := json.Unmarshal([]byte(resp.Content), &suggestion); err != nil {
		return nil, fmt.Errorf("parse suggestion: %w", err)
	}

	return &suggestion, nil
}
