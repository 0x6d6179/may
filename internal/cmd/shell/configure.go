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

func buildAliasableCommands(disabled map[string]bool) []ui.Option[string] {
	var opts []ui.Option[string]
	for _, cmd := range allCommands {
		if disabled[cmd.name] {
			continue
		}
		opts = append(opts, ui.Option[string]{
			Label:       cmd.name,
			Description: cmd.short,
			Value:       cmd.name,
		})
	}
	return opts
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

			opts := ui.RunOptions{In: f.IO.In, Out: f.IO.ErrOut}

			features, err := SelectFeatures(opts, dev)
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

			cfg, err := f.Config()
			if err != nil {
				return err
			}

			disabled := make(map[string]bool, len(cfg.DisabledCommands))
			for _, name := range cfg.DisabledCommands {
				disabled[name] = true
			}

			cmdAliases, err := selectCommandAliases(opts, cfg.ShellAliasedCommands, disabled)
			if errors.Is(err, ui.ErrAborted) {
				return nil
			}
			if err != nil {
				return err
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

func selectCommandAliases(opts ui.RunOptions, current []string, disabled map[string]bool) ([]string, error) {
	return ui.RunMultiSelect(opts, ui.MultiSelectSpec[string]{
		Title:    "select command aliases  (shell functions for any command)",
		Options:  buildAliasableCommands(disabled),
		Defaults: current,
	})
}

func SelectFeatures(opts ui.RunOptions, dev bool) ([]string, error) {
	options := []ui.Option[string]{
		{Label: "ws alias", Description: "switch workspace and cd", Value: featureWS},
		{Label: "wt alias", Description: "switch worktree and cd", Value: featureWT},
		{Label: "ai alias", Description: "ai assistant shortcut", Value: featureAI},
		{Label: "ai fix", Description: "auto-suggest fix on command error", Value: featureAIFix},
		{Label: "j alias", Description: "smart directory jump with fuzzy matching", Value: featureJ},
		{Label: "sshm alias", Description: "ssh connection manager shortcut", Value: featureSSHM},
		{Label: "completion", Description: "shell tab completion", Value: featureCompletion},
	}

	defaults := []string{featureWS, featureWT, featureAI, featureJ, featureCompletion}

	title := "select features  (core always included)"
	if dev {
		title = "select features  (core + dev always included)"
	}

	return ui.RunMultiSelect(opts, ui.MultiSelectSpec[string]{
		Title:    title,
		Options:  options,
		Defaults: defaults,
	})
}
