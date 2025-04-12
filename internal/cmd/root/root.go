package root

import (
	"fmt"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/version"
	"github.com/spf13/cobra"
)

func NewCmdRoot(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "may",
		Short:         "Personal productivity toolkit",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.Version = version.Version

	cmd.AddCommand(placeholderGroup("ws", "Workspace management"))
	cmd.AddCommand(placeholderGroup("wt", "Git worktree manager"))
	cmd.AddCommand(placeholderGroup("commit", "AI conventional commit"))
	cmd.AddCommand(placeholderGroup("id", "Git identity management"))
	cmd.AddCommand(placeholderGroup("shell", "Shell integration"))
	cmd.AddCommand(placeholderGroup("completion", "Shell completion scripts"))
	cmd.AddCommand(placeholderGroup("update", "Self-update"))

	return cmd
}

func placeholderGroup(use, short string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("%s: not yet implemented", use)
		},
	}
}
