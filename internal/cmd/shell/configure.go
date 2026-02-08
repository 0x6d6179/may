package shell

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/0x6d6179/may/internal/config"
	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/ui"
	"github.com/spf13/cobra"
)

var featureKeys = map[string]bool{
	featureWS: true, featureWT: true, featureAI: true,
	featureJ: true, featureSSHM: true, featureAIFix: true,
	featureCompletion: true,
}

var allCommands = []struct {
	name  string
	short string
}{
	{"ws", "workspace management"},
	{"wt", "git worktree manager"},
	{"j", "smart directory jump"},
	{"branch", "list or switch git branches"},
	{"recent", "show recently visited projects"},
	{"open", "open repository in browser"},
	{"ai", "ai assistant"},
	{"stash", "manage git stashes"},
	{"todo", "find todo comments"},
	{"env", "manage .env files"},
	{"run", "run project scripts"},
	{"port", "show or kill processes on a port"},
	{"db", "connect to database from .env urls"},
	{"path", "inspect and debug path"},
	{"ip", "show local and public ip addresses"},
	{"dotfiles", "manage dotfile symlinks"},
	{"weather", "show weather forecast"},
	{"b64", "base64 encode/decode"},
	{"uuid", "generate uuids"},
	{"hash", "hash strings or files"},
	{"jwt", "decode jwt tokens"},
	{"secret", "encrypt or decrypt secrets"},
	{"qr", "generate qr codes"},
	{"id", "git identity management"},
	{"sshm", "ssh connection manager"},
	{"alias", "manage shell command aliases"},
}

var defaultIntegrations = []string{featureWS, featureWT, featureAI, featureJ, featureCompletion}

func buildIntegrationOptions(disabled map[string]bool) []ui.Option[string] {
	opts := []ui.Option[string]{
		{Label: "ai fix", Description: "auto-suggest fix on failed commands", Value: featureAIFix},
		{Label: "completion", Description: "shell tab completion", Value: featureCompletion},
	}

	shellFuncs := []struct{ label, value, desc string }{
		{"ws", featureWS, "workspace management"},
		{"wt", featureWT, "git worktree manager"},
		{"ai", featureAI, "ai assistant"},
		{"j", featureJ, "smart directory jump (+ k to go back)"},
		{"sshm", featureSSHM, "ssh connection manager"},
	}
	for _, sf := range shellFuncs {
		if disabled[sf.value] {
			continue
		}
		opts = append(opts, ui.Option[string]{Label: sf.label, Description: sf.desc, Value: sf.value})
	}

	for _, cmd := range allCommands {
		if featureKeys[cmd.name] || disabled[cmd.name] {
			continue
		}
		opts = append(opts, ui.Option[string]{Label: cmd.name, Description: cmd.short, Value: cmd.name})
	}

	return opts
}

func SelectIntegrations(opts ui.RunOptions, currentFeatures []string, currentAliases []string, disabled map[string]bool, dev bool) ([]string, []string, error) {
	var defaults []string
	if len(currentFeatures) == 0 && len(currentAliases) == 0 {
		defaults = defaultIntegrations
	} else {
		for _, f := range currentFeatures {
			if featureKeys[f] {
				defaults = append(defaults, f)
			}
		}
		defaults = append(defaults, currentAliases...)
	}

	title := "configure shell integrations  (may wrapper always included)"
	if dev {
		title = "configure shell integrations  (may wrapper + dev mode always included)"
	}

	selected, err := ui.RunMultiSelect(opts, ui.MultiSelectSpec[string]{
		Title:    title,
		Options:  buildIntegrationOptions(disabled),
		Defaults: defaults,
	})
	if err != nil {
		return nil, nil, err
	}

	var features []string
	var cmdAliases []string
	for _, v := range selected {
		if featureKeys[v] {
			features = append(features, v)
		} else {
			cmdAliases = append(cmdAliases, v)
		}
	}
	return features, cmdAliases, nil
}

func NewCmdShellConfigure(f *factory.Factory) *cobra.Command {
	var dev bool
	cmd := &cobra.Command{
		Use:   "configure [bash|zsh|fish]",
		Short: "interactively configure shell integration",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sh, err := resolveShell(args)
			if err != nil {
				return err
			}

			profile := ProfileFile(sh)
			if profile == "" {
				return fmt.Errorf("unsupported shell: %s", sh)
			}

			cfg, err := f.Config()
			if err != nil {
				return err
			}

			disabled := make(map[string]bool, len(cfg.DisabledCommands))
			for _, name := range cfg.DisabledCommands {
				disabled[name] = true
			}

			opts := ui.RunOptions{In: f.IO.In, Out: f.IO.ErrOut}
			currentFeatures, _ := ReadFeatures(profile)

			features, cmdAliases, err := SelectIntegrations(opts, currentFeatures, cfg.ShellAliasedCommands, disabled, dev)
			if errors.Is(err, ui.ErrAborted) {
				return nil
			}
			if err != nil {
				return err
			}

			features = append([]string{featureCore}, features...)
			if dev {
				features = append(features, featureDev)
			}

			var devPath string
			if dev {
				devPath, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("get working directory: %w", err)
				}
			}

			var aliases []AliasEntry
			for _, a := range cfg.Aliases {
				aliases = append(aliases, AliasEntry{Name: a.Name, Command: a.Command})
			}
			for _, name := range cmdAliases {
				aliases = append(aliases, AliasEntry{Name: name, Command: name})
			}

			snippet := BuildSnippet(sh, features, devPath, aliases...)
			featuresLabel := strings.Join(features, ",")

			action := "add to"
			if HasBlock(profile) {
				action = "update"
			}

			title := fmt.Sprintf("%s %s?", action, profile)
			if dev {
				title = fmt.Sprintf("%s %s? (dev mode)", action, profile)
			}

			confirmed, err := ui.RunConfirm(opts, ui.ConfirmSpec{Title: title})
			if errors.Is(err, ui.ErrAborted) {
				return nil
			}
			if err != nil {
				return err
			}
			if !confirmed {
				return nil
			}

			cfg.ShellAliasedCommands = cmdAliases
			if err := config.Save(cfg); err != nil {
				return err
			}

			if err := WriteBlock(profile, featuresLabel, snippet); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "✓ configured %s in %s\n", strings.Join(features, ", "), profile)
			fmt.Fprintf(f.IO.ErrOut, "  run: source %s\n", profile)
			return nil
		},
	}
	cmd.Flags().BoolVar(&dev, "dev", false, "prepend cwd to PATH (development mode)")
	return cmd
}
