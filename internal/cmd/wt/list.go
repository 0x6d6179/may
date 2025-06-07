package wt

import (
	"fmt"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/git"
	"github.com/0x6d6179/may/internal/ui"
	"github.com/spf13/cobra"
)

func NewCmdWtList(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all worktrees",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			runner := &git.Runner{}

			worktrees, err := git.ListWorktrees(runner)
			if err != nil {
				return err
			}

			w := ui.NewTable(f.IO.ErrOut)
			fmt.Fprintln(w, "path\tbranch\tmain")
			for _, wt := range worktrees {
				main := ""
				if wt.IsMain {
					main = "✓"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n", wt.Path, wt.Branch, main)
			}
			w.Flush()

			return nil
		},
	}

	return cmd
}
