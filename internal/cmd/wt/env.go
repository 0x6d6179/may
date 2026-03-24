package wt

import (
	"fmt"
	"os"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/git"
	"github.com/spf13/cobra"
)

func NewCmdWtEnv(f *factory.Factory) *cobra.Command {
	var reverse bool

	cmd := &cobra.Command{
		Use:   "env",
		Short: "sync .env files between main worktree and current",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			runner := &git.Runner{}

			mainPath, err := git.MainWorktreePath(runner)
			if err != nil {
				return err
			}

			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			src, dst := mainPath, cwd
			direction := "main → current"
			if reverse {
				src, dst = cwd, mainPath
				direction = "current → main"
			}

			if err := git.CopyEnvFiles(src, dst); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "✓ env files synced (%s)\n", direction)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&reverse, "reverse", "r", false, "copy from current worktree to main")

	return cmd
}
