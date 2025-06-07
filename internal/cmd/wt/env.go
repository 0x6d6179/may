package wt

import (
	"fmt"
	"os"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/git"
	"github.com/spf13/cobra"
)

func NewCmdWtEnv(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "copy .env files from main worktree to current",
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

			if err := git.CopyEnvFiles(mainPath, cwd); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "✓ env files synced\n  - from: %s\n  - to:   %s\n", mainPath, cwd)
			return nil
		},
	}

	return cmd
}
