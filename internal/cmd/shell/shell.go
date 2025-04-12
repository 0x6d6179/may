package shell

import (
	"github.com/0x6d6179/may/internal/factory"
	"github.com/spf13/cobra"
)

// NewCmdShell returns the shell group command.
func NewCmdShell(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shell",
		Short: "Shell integration",
	}

	cmd.AddCommand(NewCmdShellInit(f))

	return cmd
}
