package id

import (
	"fmt"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/ui"
	"github.com/spf13/cobra"
)

// NewCmdIdList returns the `id list` subcommand.
func NewCmdIdList(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all git identity profiles and mappings",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			if len(cfg.Git.Profiles) > 0 {
				w := ui.NewTable(f.IO.ErrOut)
				fmt.Fprintln(w, "name\tusername\temail\tgh_user")
				for _, p := range cfg.Git.Profiles {
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.Name, p.Username, p.Email, p.GHUser)
				}
				w.Flush()
			}

			if len(cfg.Git.Mappings) > 0 {
				w := ui.NewTable(f.IO.ErrOut)
				fmt.Fprintln(w, "path\tprofile")
				for _, m := range cfg.Git.Mappings {
					fmt.Fprintf(w, "%s\t%s\n", m.Path, m.Profile)
				}
				w.Flush()
			}

			return nil
		},
	}

	return cmd
}
