package wt

import (
	"errors"
	"fmt"
	"os"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/git"
	"github.com/spf13/cobra"
)

func NewCmdWtClose(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close",
		Short: "Close the current worktree and return to main",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			runner := &git.Runner{}

			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			if !git.IsInsideWorktree(runner, cwd) {
				fmt.Fprintln(f.IO.ErrOut, "not inside a worktree")
				return errors.New("not inside a worktree")
			}

			mainPath, err := git.MainWorktreePath(runner)
			if err != nil {
				return err
			}

			branch, err := git.CurrentBranch(runner)
			if err != nil {
				return err
			}

			// Print main path first so shell wrapper can cd before cleanup
			fmt.Fprintln(f.IO.Out, mainPath)

			if _, err := runner.Run("worktree", "remove", "--force", cwd); err != nil {
				return err
			}

			// Best-effort branch deletion
			runner.Run("branch", "-D", branch)

			if _, err := runner.Run("worktree", "prune"); err != nil {
				fmt.Fprintf(f.IO.ErrOut, "warning: worktree prune: %v\n", err)
			}

			mainRunner := &git.Runner{Dir: mainPath}
			if _, err := mainRunner.Run("pull"); err != nil {
				fmt.Fprintf(f.IO.ErrOut, "warning: pull: %v\n", err)
			}

			return nil
		},
	}

	return cmd
}
