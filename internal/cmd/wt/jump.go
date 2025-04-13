package wt

import (
	"fmt"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/git"
	"github.com/spf13/cobra"
)

func NewCmdWtJump(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jump",
		Short: "Jump to the main worktree",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			runner := &git.Runner{}

			mainPath, err := git.MainWorktreePath(runner)
			if err != nil {
				return err
			}

			fmt.Fprintln(f.IO.Out, mainPath)
			return nil
		},
	}

	return cmd
}
