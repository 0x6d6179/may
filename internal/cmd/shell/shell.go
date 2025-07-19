package shell

import (
	"github.com/0x6d6179/may/internal/factory"
	"github.com/spf13/cobra"
)

// NewCmdShell returns the shell group command.
func NewCmdShell(f *factory.Factory, rootCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shell",
		Short: "shell integration",
	}

	cmd.AddCommand(NewCmdShellInit(f))
	cmd.AddCommand(NewCmdShellCompletion(f, rootCmd))
	cmd.AddCommand(NewCmdShellAlias(f))
	cmd.AddCommand(NewCmdShellInstall(f))
	cmd.AddCommand(NewCmdShellConfigure(f))

	return cmd
}
