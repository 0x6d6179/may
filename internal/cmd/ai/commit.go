package ai

import (
	"github.com/0x6d6179/may/internal/cmd/commit"
	"github.com/0x6d6179/may/internal/factory"
	"github.com/spf13/cobra"
)

func newCmdCommit(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "commit",
		Short: "ai conventional commit",
		RunE: func(cmd *cobra.Command, args []string) error {
			return commit.RunCommit(cmd.Context(), f)
		},
	}
}
