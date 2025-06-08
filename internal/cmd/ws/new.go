package ws

import (
	"fmt"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/workspace"
	"github.com/spf13/cobra"
)

func NewCmdWsNew(f *factory.Factory) *cobra.Command {
	var rootName string

	cmd := &cobra.Command{
		Use:   "new [name]",
		Short: "create a new workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			name := args[0]

			if rootName == "" {
				dr := workspace.DefaultRoot(cfg)
				if dr == nil {
					return fmt.Errorf("no default root configured; use --in to specify a root")
				}
				rootName = dr.Name
			}

			path, err := workspace.Create(cfg, name, rootName)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "✓ workspace created: %s\n", name)
			fmt.Fprintln(f.IO.Out, path)
			return nil
		},
	}

	cmd.Flags().StringVar(&rootName, "in", "", "root name to create workspace under")
	cmd.RegisterFlagCompletionFunc("in", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		cfg, err := f.Config()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		names := make([]string, len(cfg.Workspace.Roots))
		for i, r := range cfg.Workspace.Roots {
			names[i] = r.Name
		}
		return names, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}
