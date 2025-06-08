package root

import (
	"github.com/0x6d6179/may/internal/cmd/commit"
	"github.com/0x6d6179/may/internal/cmd/completion"
	"github.com/0x6d6179/may/internal/cmd/id"
	initcmd "github.com/0x6d6179/may/internal/cmd/init"
	"github.com/0x6d6179/may/internal/cmd/shell"
	"github.com/0x6d6179/may/internal/cmd/update"
	"github.com/0x6d6179/may/internal/cmd/ws"
	"github.com/0x6d6179/may/internal/cmd/wt"
	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/version"
	"github.com/spf13/cobra"
)

func NewCmdRoot(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "may",
		Short:         "personal productivity toolkit",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.Version = version.Version

	cmd.AddCommand(ws.NewCmdWs(f))
	cmd.AddCommand(wt.NewCmdWt(f))
	cmd.AddCommand(commit.NewCmdCommit(f))
	cmd.AddCommand(id.NewCmdId(f))
	cmd.AddCommand(shell.NewCmdShell(f, cmd))
	cmd.AddCommand(update.NewCmdUpdate(f))
	cmd.AddCommand(initcmd.NewCmdInit(f))

	// hidden backward-compat completion command
	completionCmd := completion.NewCmdCompletion(f, cmd)
	completionCmd.Hidden = true
	cmd.AddCommand(completionCmd)

	// lowercase cobra's built-in help command
	cmd.InitDefaultHelpCmd()
	for _, c := range cmd.Commands() {
		if c.Name() == "help" {
			c.Short = "help about any command"
		}
	}

	return cmd
}
