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

var aliasableCommands = []ui.Option[string]{
	{Label: "branch", Description: "list or switch git branches", Value: "branch"},
	{Label: "recent", Description: "show recently visited projects", Value: "recent"},
	{Label: "open", Description: "open repository in browser", Value: "open"},
	{Label: "stash", Description: "manage git stashes", Value: "stash"},
	{Label: "todo", Description: "find todo comments", Value: "todo"},
	{Label: "env", Description: "manage .env files", Value: "env"},
	{Label: "run", Description: "run project scripts", Value: "run"},
	{Label: "port", Description: "show or kill processes on a port", Value: "port"},
	{Label: "db", Description: "connect to database from .env urls", Value: "db"},
	{Label: "path", Description: "inspect and debug path", Value: "path"},
	{Label: "ip", Description: "show local and public ip addresses", Value: "ip"},
	{Label: "dotfiles", Description: "manage dotfile symlinks", Value: "dotfiles"},
	{Label: "weather", Description: "show weather forecast", Value: "weather"},
	{Label: "b64", Description: "base64 encode/decode", Value: "b64"},
	{Label: "uuid", Description: "generate uuids", Value: "uuid"},
	{Label: "hash", Description: "hash strings or files", Value: "hash"},
	{Label: "jwt", Description: "decode jwt tokens", Value: "jwt"},
	{Label: "secret", Description: "encrypt or decrypt secrets", Value: "secret"},
	{Label: "qr", Description: "generate qr codes", Value: "qr"},
	{Label: "id", Description: "git identity management", Value: "id"},
	{Label: "alias", Description: "manage shell command aliases", Value: "alias"},
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

			cmdAliases, err := selectCommandAliases(opts, cfg.ShellAliasedCommands)
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

func saveConfig(cfg *config.Config) error {
	return config.Save(cfg)
}

func selectCommandAliases(opts ui.RunOptions, current []string) ([]string, error) {
	defaults := current
	return ui.RunMultiSelect(opts, ui.MultiSelectSpec[string]{
		Title:    "select command aliases  (shell functions for any command)",
		Options:  aliasableCommands,
		Defaults: defaults,
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
