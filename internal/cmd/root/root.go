package root

import (
	"github.com/0x6d6179/may/internal/cmd/commit"
	"github.com/0x6d6179/may/internal/cmd/shell"
	"github.com/0x6d6179/may/internal/cmd/ws"
	"github.com/0x6d6179/may/internal/cmd/wt"
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

	cmd.AddCommand(ws.NewCmdWs(f))
	cmd.AddCommand(wt.NewCmdWt(f))
	cmd.AddCommand(commit.NewCmdCommit(f))
	cmd.AddCommand(idGroup())
	cmd.AddCommand(shell.NewCmdShell(f))
	cmd.AddCommand(completionGroup())
	cmd.AddCommand(updateGroup())

	return cmd
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
