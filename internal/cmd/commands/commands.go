package commands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/0x6d6179/may/internal/cmd/shell"
	"github.com/0x6d6179/may/internal/config"
	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/ui"
	"github.com/spf13/cobra"
)

type cmdMeta struct {
	Name         string
	Short        string
	Group        string
	ShellFeature string
}

var registry = []cmdMeta{
	{Name: "ws", Group: "nav", Short: "workspace management", ShellFeature: "ws"},
	{Name: "wt", Group: "nav", Short: "git worktree manager", ShellFeature: "wt"},
	{Name: "j", Group: "nav", Short: "smart directory jump", ShellFeature: "j"},
	{Name: "branch", Group: "nav", Short: "list or switch git branches"},
	{Name: "recent", Group: "nav", Short: "show recently visited projects"},
	{Name: "open", Group: "nav", Short: "open repository in browser"},
	{Name: "ai", Group: "ai", Short: "ai assistant", ShellFeature: "ai"},
	{Name: "stash", Group: "git", Short: "manage git stashes"},
	{Name: "todo", Group: "git", Short: "find todo comments"},
	{Name: "env", Group: "git", Short: "manage .env files"},
	{Name: "run", Group: "project", Short: "run project scripts"},
	{Name: "port", Group: "project", Short: "show or kill processes on a port"},
	{Name: "db", Group: "project", Short: "connect to database from .env urls"},
	{Name: "path", Group: "system", Short: "inspect and debug path"},
	{Name: "ip", Group: "system", Short: "show local and public ip addresses"},
	{Name: "dotfiles", Group: "system", Short: "manage dotfile symlinks"},
	{Name: "weather", Group: "system", Short: "show weather forecast"},
	{Name: "b64", Group: "encode", Short: "base64 encode/decode"},
	{Name: "uuid", Group: "encode", Short: "generate uuids"},
	{Name: "hash", Group: "encode", Short: "hash strings or files"},
	{Name: "jwt", Group: "encode", Short: "decode jwt tokens"},
	{Name: "secret", Group: "encode", Short: "encrypt or decrypt secrets"},
	{Name: "qr", Group: "encode", Short: "generate qr codes"},
	{Name: "id", Group: "meta", Short: "git identity management"},
	{Name: "sshm", Group: "meta", Short: "ssh connection manager", ShellFeature: "sshm"},
	{Name: "alias", Group: "meta", Short: "manage shell command aliases"},
}

var groupOrder = []string{"nav", "ai", "git", "project", "system", "encode", "meta"}

var groupTitles = map[string]string{
	"nav":     "workspace & navigation",
	"ai":      "ai",
	"git":     "git utilities",
	"project": "project tools",
	"system":  "system & path",
	"encode":  "encode / decode",
	"meta":    "identity & meta",
}

func NewCmdCommands(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "commands",
		Short:   "manage may commands",
		Aliases: []string{"cmd", "cmds"},
	}
	cmd.AddCommand(newCmdCommandsList(f))
	cmd.AddCommand(newCmdCommandsConfigure(f))
	return cmd
}

func newCmdCommandsList(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list all commands and their status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			disabled := toSet(cfg.DisabledCommands)
			for _, g := range groupOrder {
				fmt.Fprintf(f.IO.ErrOut, "\n  %s\n", groupTitles[g])
				for _, m := range registry {
					if m.Group != g {
						continue
					}
					if disabled[m.Name] {
						fmt.Fprintf(f.IO.ErrOut, "  ○ %-14s  %s\n", m.Name, m.Short)
					} else {
						fmt.Fprintf(f.IO.ErrOut, "  ● %-14s  %s\n", m.Name, m.Short)
					}
				}
			}
			fmt.Fprintln(f.IO.ErrOut)
			return nil
		},
	}
}

func newCmdCommandsConfigure(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "configure",
		Short: "enable or disable commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			disabled := toSet(cfg.DisabledCommands)

			opts := ui.RunOptions{In: f.IO.In, Out: f.IO.ErrOut}
			options := make([]ui.Option[string], len(registry))
			defaults := make([]string, 0, len(registry))
			for i, m := range registry {
				options[i] = ui.Option[string]{
					Label:       m.Name,
					Description: fmt.Sprintf("(%s) %s", m.Group, m.Short),
					Value:       m.Name,
				}
				if !disabled[m.Name] {
					defaults = append(defaults, m.Name)
				}
			}

			selected, err := ui.RunMultiSelect(opts, ui.MultiSelectSpec[string]{
				Title:    "enable or disable commands",
				Options:  options,
				Defaults: defaults,
			})
			if errors.Is(err, ui.ErrAborted) {
				return nil
			}
			if err != nil {
				return err
			}

			selectedSet := toSet(selected)

			var newlyDisabled, newlyEnabled []cmdMeta
			for _, m := range registry {
				wasDisabled := disabled[m.Name]
				nowDisabled := !selectedSet[m.Name]
				if !wasDisabled && nowDisabled {
					newlyDisabled = append(newlyDisabled, m)
				}
				if wasDisabled && !nowDisabled {
					newlyEnabled = append(newlyEnabled, m)
				}
			}

			var newDisabled []string
			for _, m := range registry {
				if !selectedSet[m.Name] {
					newDisabled = append(newDisabled, m.Name)
				}
			}
			cfg.DisabledCommands = newDisabled
			if err := config.Save(cfg); err != nil {
				return err
			}

			for _, m := range newlyDisabled {
				if m.ShellFeature == "" {
					continue
				}
				if err := removeShellFeature(m, cfg, f); err != nil {
					fmt.Fprintf(f.IO.ErrOut, "  ! could not update shell: %v\n", err)
				}
			}

			for _, m := range newlyEnabled {
				if m.ShellFeature == "" {
					continue
				}
				if err := offerAddShellFeature(m, opts, cfg, f); err != nil {
					fmt.Fprintf(f.IO.ErrOut, "  ! could not update shell: %v\n", err)
				}
			}

			fmt.Fprintf(f.IO.ErrOut, "✓ saved\n")
			return nil
		},
	}
}

func removeShellFeature(m cmdMeta, cfg *config.Config, f *factory.Factory) error {
	sh, err := shell.DetectShell()
	if err != nil {
		return nil
	}
	profile := shell.ProfileFile(sh)
	if !shell.HasBlock(profile) {
		return nil
	}
	features, err := shell.ReadFeatures(profile)
	if err != nil {
		return err
	}
	features = removeFrom(features, m.ShellFeature)
	snippet := shell.BuildSnippet(sh, features, "", buildAliasEntries(cfg)...)
	if err := shell.WriteBlock(profile, strings.Join(features, ","), snippet); err != nil {
		return err
	}
	fmt.Fprintf(f.IO.ErrOut, "  ✓ removed %s from %s\n", m.ShellFeature, profile)
	return nil
}

func offerAddShellFeature(m cmdMeta, opts ui.RunOptions, cfg *config.Config, f *factory.Factory) error {
	sh, err := shell.DetectShell()
	if err != nil {
		return nil
	}
	profile := shell.ProfileFile(sh)
	if !shell.HasBlock(profile) {
		return nil
	}
	confirmed, err := ui.RunConfirm(opts, ui.ConfirmSpec{
		Title: fmt.Sprintf("add %s to shell integration?", m.Name),
	})
	if errors.Is(err, ui.ErrAborted) || !confirmed {
		return err
	}
	if err != nil {
		return err
	}
	features, err := shell.ReadFeatures(profile)
	if err != nil {
		return err
	}
	if !sliceContains(features, m.ShellFeature) {
		features = append(features, m.ShellFeature)
	}
	snippet := shell.BuildSnippet(sh, features, "", buildAliasEntries(cfg)...)
	if err := shell.WriteBlock(profile, strings.Join(features, ","), snippet); err != nil {
		return err
	}
	fmt.Fprintf(f.IO.ErrOut, "  ✓ added %s to %s\n", m.ShellFeature, profile)
	return nil
}

func buildAliasEntries(cfg *config.Config) []shell.AliasEntry {
	entries := make([]shell.AliasEntry, 0, len(cfg.Aliases)+len(cfg.ShellAliasedCommands))
	for _, a := range cfg.Aliases {
		entries = append(entries, shell.AliasEntry{Name: a.Name, Command: a.Command})
	}
	for _, name := range cfg.ShellAliasedCommands {
		entries = append(entries, shell.AliasEntry{Name: name, Command: name})
	}
	return entries
}

func toSet(slice []string) map[string]bool {
	s := make(map[string]bool, len(slice))
	for _, v := range slice {
		s[v] = true
	}
	return s
}

func removeFrom(slice []string, val string) []string {
	out := slice[:0]
	for _, s := range slice {
		if s != val {
			out = append(out, s)
		}
	}
	return out
}

func sliceContains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}
