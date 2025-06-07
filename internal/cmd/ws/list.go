package ws

import (
	"fmt"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/ui"
	"github.com/0x6d6179/may/internal/workspace"
	"github.com/spf13/cobra"
)

func NewCmdWsList(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list all workspaces",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			workspaces := workspace.List(cfg)
			if len(workspaces) == 0 {
				return nil
			}

			tw := ui.NewTable(f.IO.ErrOut)
			for _, ws := range workspaces {
				fmt.Fprintf(tw, "%s\t%s\t%s\n", ws.Name, ws.Root, ws.Path)
			}
			return tw.Flush()
		},
	}
}
