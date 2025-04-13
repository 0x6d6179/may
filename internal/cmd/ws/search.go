package ws

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/workspace"
	"github.com/spf13/cobra"
)

func NewCmdWsSearch(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "search [query]",
		Short: "Search workspaces by name",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			query := ""
			if len(args) > 0 {
				query = strings.ToLower(args[0])
			}

			workspaces := workspace.List(cfg)
			tw := tabwriter.NewWriter(f.IO.ErrOut, 0, 2, 2, ' ', 0)
			for _, ws := range workspaces {
				if query == "" || strings.Contains(strings.ToLower(ws.Name), query) {
					fmt.Fprintf(tw, "%s\t%s\t%s\n", ws.Name, ws.Root, ws.Path)
				}
			}
			return tw.Flush()
		},
	}
}
