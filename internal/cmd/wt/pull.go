package wt

import (
	"fmt"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/git"
	"github.com/spf13/cobra"
)

func NewCmdWtPull(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pull",
		Short: "fetch origin and merge origin/main",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			runner := &git.Runner{}

			out, err := runner.Run("fetch", "origin")
			if err != nil {
				fmt.Fprintln(f.IO.ErrOut, out)
				return err
			}
			if out != "" {
				fmt.Fprintln(f.IO.ErrOut, out)
			}

			out, err = runner.Run("merge", "origin/main")
			if err != nil {
				fmt.Fprintln(f.IO.ErrOut, out)
				return err
			}
			if out != "" {
				fmt.Fprintln(f.IO.ErrOut, out)
			}

			return nil
		},
	}

	return cmd
}
