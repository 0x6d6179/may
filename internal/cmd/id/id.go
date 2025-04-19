package id

import (
	"github.com/0x6d6179/may/internal/factory"
	"github.com/spf13/cobra"
)

// NewCmdId returns the id command group and registers all subcommands.
func NewCmdId(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "id",
		Short: "Git identity management",
	}

	cmd.AddCommand(NewCmdIdStatus(f))
	cmd.AddCommand(NewCmdIdSwitch(f))
	cmd.AddCommand(NewCmdIdAdd(f))
	cmd.AddCommand(NewCmdIdMap(f))
	cmd.AddCommand(NewCmdIdUnmap(f))
	cmd.AddCommand(NewCmdIdList(f))

	return cmd
}
