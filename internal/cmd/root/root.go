package root

import (
	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/version"
	"github.com/spf13/cobra"
)

// NewCmdRoot returns the root cobra command for the may CLI.
func NewCmdRoot(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "may",
		Short:         "Personal productivity toolkit",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.Version = version.Version

	// Subcommand groups — each is replaced by a real implementation in its own package.
	cmd.AddCommand(wsGroup())
	cmd.AddCommand(wtGroup())
	cmd.AddCommand(commitGroup())
	cmd.AddCommand(idGroup())
	cmd.AddCommand(shellGroup())
	cmd.AddCommand(completionGroup())
	cmd.AddCommand(updateGroup())

	return cmd
}

func wsGroup() *cobra.Command {
	return &cobra.Command{Use: "ws", Short: "Workspace management"}
}

func wtGroup() *cobra.Command {
	return &cobra.Command{Use: "wt", Short: "Git worktree manager"}
}

func commitGroup() *cobra.Command {
	return &cobra.Command{Use: "commit", Short: "AI conventional commit"}
}

func idGroup() *cobra.Command {
	return &cobra.Command{Use: "id", Short: "Git identity management"}
}

func shellGroup() *cobra.Command {
	return &cobra.Command{Use: "shell", Short: "Shell integration"}
}

func completionGroup() *cobra.Command {
	return &cobra.Command{Use: "completion", Short: "Shell completion scripts"}
}

func updateGroup() *cobra.Command {
	return &cobra.Command{Use: "update", Short: "Self-update"}
}
