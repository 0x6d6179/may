package wt

import (
	"errors"
	"fmt"
	"os"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/git"
	"github.com/spf13/cobra"
)

func NewCmdWtRm(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm [branch]",
		Short: "Remove a worktree",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runner := &git.Runner{}

			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			worktrees, err := git.ListWorktrees(runner)
			if err != nil {
				return err
			}

			var targetBranch, targetPath string

			if len(args) == 0 {
				// Default: branch of current worktree
				branch, err := git.CurrentBranch(runner)
				if err != nil {
					return err
				}
				targetBranch = branch

				for _, wt := range worktrees {
					if wt.Branch == targetBranch {
						targetPath = wt.Path
						break
					}
				}
			} else {
				targetBranch = args[0]
				for _, wt := range worktrees {
					if wt.Branch == targetBranch {
						targetPath = wt.Path
						break
					}
				}
			}

			if targetPath == "" {
				return fmt.Errorf("worktree for branch %q not found", targetBranch)
			}

			if cwd == targetPath {
				fmt.Fprintln(f.IO.ErrOut, "cannot remove current worktree")
				return errors.New("cannot remove current worktree")
			}

			if _, err := runner.Run("worktree", "remove", "--force", targetPath); err != nil {
				return err
			}

			// Best-effort branch deletion
			runner.Run("branch", "-D", targetBranch)

			if _, err := runner.Run("worktree", "prune"); err != nil {
				fmt.Fprintf(f.IO.ErrOut, "warning: worktree prune: %v\n", err)
			}

			fmt.Fprintf(f.IO.ErrOut, "removed worktree: %s\n", targetPath)
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			runner := &git.Runner{}
			worktrees, err := git.ListWorktrees(runner)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			branches := make([]string, 0, len(worktrees))
			for _, wt := range worktrees {
				if wt.Branch != "" {
					branches = append(branches, wt.Branch)
				}
			}
			return branches, cobra.ShellCompDirectiveNoFileComp
		},
	}

	return cmd
}
