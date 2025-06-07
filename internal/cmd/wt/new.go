package wt

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/git"
	"github.com/0x6d6179/may/internal/ui"
	"github.com/spf13/cobra"
)

func NewCmdWtNew(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new [branch]",
		Short: "create a new worktree",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runner := &git.Runner{}

			repoName, err := git.RepoName(runner)
			if err != nil {
				return err
			}

			sanitized := git.BranchSanitize(args[0])

			cfg, err := f.Config()
			if err != nil {
				return err
			}

			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			var matchingRootPath string
			matchLen := 0
			for _, root := range cfg.Workspace.Roots {
				if strings.HasPrefix(cwd, root.Path) && len(root.Path) > matchLen {
					matchingRootPath = root.Path
					matchLen = len(root.Path)
				}
			}
			if matchingRootPath == "" {
				return fmt.Errorf("current directory is not under a configured workspace root")
			}

			shadowPath := filepath.Join(matchingRootPath, ".worktrees", repoName, sanitized)

			stop := ui.Spinner(f.IO.ErrOut, "creating worktree...")
			_, err = runner.Run("worktree", "add", "-b", args[0], shadowPath)
			stop()
			if err != nil {
				return err
			}

			mainPath, err := git.MainWorktreePath(runner)
			if err != nil {
				return err
			}

			if err := git.CopyEnvFiles(mainPath, shadowPath); err != nil {
				fmt.Fprintf(f.IO.ErrOut, "warning: could not copy env files: %v\n", err)
			}

			if !f.IO.IsTerminal() {
				fmt.Fprintln(f.IO.Out, shadowPath)
			}
			fmt.Fprintf(f.IO.ErrOut, "✓ worktree created: %s\n  - location: %s\n", args[0], shadowPath)
			return nil
		},
	}

	return cmd
}
