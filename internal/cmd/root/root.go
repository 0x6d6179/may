package root

import (
	"github.com/0x6d6179/may/internal/cmd/shell"
	"github.com/0x6d6179/may/internal/cmd/ws"
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
	cmd.AddCommand(ws.NewCmdWs(f))
	cmd.AddCommand(wtGroup())
	cmd.AddCommand(commitGroup())
	cmd.AddCommand(idGroup())
	cmd.AddCommand(shell.NewCmdShell(f))
	cmd.AddCommand(completionGroup())
	cmd.AddCommand(updateGroup())

	return cmd
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

func completionGroup() *cobra.Command {
	return &cobra.Command{Use: "completion", Short: "Shell completion scripts"}
}

func updateGroup() *cobra.Command {
	return &cobra.Command{Use: "update", Short: "Self-update"}
}
