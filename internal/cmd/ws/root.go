package ws

import (
	"fmt"

	"github.com/0x6d6179/may/internal/config"
	"github.com/0x6d6179/may/internal/factory"
	"github.com/spf13/cobra"
)

func NewCmdWsRoot(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "root",
		Short: "Manage workspace roots",
	}

	cmd.AddCommand(newCmdWsRootAdd(f))
	cmd.AddCommand(newCmdWsRootList(f))
	cmd.AddCommand(newCmdWsRootRm(f))

	return cmd
}

func newCmdWsRootAdd(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "add <name> <path>",
		Short: "Add a workspace root",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			name, path := args[0], args[1]
			cfg.Workspace.Roots = append(cfg.Workspace.Roots, config.WorkspaceRoot{
				Name: name,
				Path: path,
			})

			if err := config.Save(cfg); err != nil {
				return err
			}

			fmt.Fprintln(f.IO.ErrOut, "added root:", name, "->", path)
			return nil
		},
	}
}

func newCmdWsRootList(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List workspace roots",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			for _, r := range cfg.Workspace.Roots {
				fmt.Fprintf(f.IO.ErrOut, "%s\t%s\n", r.Name, r.Path)
			}
			return nil
		},
	}
}

func newCmdWsRootRm(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "rm <name>",
		Short: "Remove a workspace root",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			name := args[0]
			roots := cfg.Workspace.Roots[:0]
			found := false
			for _, r := range cfg.Workspace.Roots {
				if r.Name == name {
					found = true
					continue
				}
				roots = append(roots, r)
			}

			if !found {
				return fmt.Errorf("root %q not found", name)
			}

			cfg.Workspace.Roots = roots
			if err := config.Save(cfg); err != nil {
				return err
			}

			fmt.Fprintln(f.IO.ErrOut, "removed root:", name)
			return nil
		},
	}
}
