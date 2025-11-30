package alias

import (
	"fmt"
	"os/exec"

	"github.com/0x6d6179/may/internal/config"
	"github.com/0x6d6179/may/internal/factory"
	"github.com/spf13/cobra"
)

var reservedNames = map[string]bool{
	"ws": true, "wt": true, "ai": true, "id": true, "shell": true,
	"update": true, "init": true, "help": true, "completion": true,
	"j": true, "branch": true, "br": true, "todo": true, "stash": true,
	"env": true, "open": true, "run": true, "port": true, "path": true,
	"ip": true, "b64": true, "base64": true, "uuid": true, "hash": true,
	"jwt": true, "secret": true, "dotfiles": true, "weather": true,
	"db": true, "recent": true, "alias": true, "configure": true,
	"sshm": true, "ssh": true, "k": true, "commands": true,
	"cmd": true, "cmds": true, "qr": true,
}

func NewCmdAlias(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alias",
		Short: "manage shell command aliases",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listAliases(f)
		},
	}

	cmd.AddCommand(newCmdAliasAdd(f))
	cmd.AddCommand(newCmdAliasRm(f))
	cmd.AddCommand(newCmdAliasList(f))

	return cmd
}

func newCmdAliasList(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list all aliases",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listAliases(f)
		},
	}
}

func listAliases(f *factory.Factory) error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	if len(cfg.Aliases) == 0 {
		fmt.Fprintln(f.IO.ErrOut, "no aliases configured")
		return nil
	}

	for _, a := range cfg.Aliases {
		fmt.Fprintf(f.IO.Out, "%s → %s\n", a.Name, a.Command)
	}
	return nil
}

func newCmdAliasAdd(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "add <name> <command>",
		Short: "add a shell alias",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			command := args[1]

			if err := checkConflicts(name); err != nil {
				return err
			}

			cfg, err := f.Config()
			if err != nil {
				return err
			}

			for _, a := range cfg.Aliases {
				if a.Name == name {
					return fmt.Errorf("alias %q already exists — remove it first with: may alias rm %s", name, name)
				}
			}

			cfg.Aliases = append(cfg.Aliases, config.Alias{Name: name, Command: command})
			if err := config.Save(cfg); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "✓ added alias: %s → %s\n", name, command)
			fmt.Fprintln(f.IO.ErrOut, "  run: may shell configure to update shell integration")
			return nil
		},
	}
}

func newCmdAliasRm(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "rm <name>",
		Aliases: []string{"remove", "delete"},
		Short:   "remove a shell alias",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfg, err := f.Config()
			if err != nil {
				return err
			}

			found := false
			filtered := cfg.Aliases[:0]
			for _, a := range cfg.Aliases {
				if a.Name == name {
					found = true
					continue
				}
				filtered = append(filtered, a)
			}

			if !found {
				return fmt.Errorf("alias %q not found", name)
			}

			cfg.Aliases = filtered
			if err := config.Save(cfg); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "✓ removed alias: %s\n", name)
			fmt.Fprintln(f.IO.ErrOut, "  run: may shell configure to update shell integration")
			return nil
		},
	}
}

func checkConflicts(name string) error {
	if reservedNames[name] {
		return fmt.Errorf("%q conflicts with a built-in may command", name)
	}

	if _, err := exec.LookPath(name); err == nil {
		return fmt.Errorf("%q conflicts with an installed binary on PATH", name)
	}

	return nil
}
