package completion

import (
	"errors"
	"fmt"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/spf13/cobra"
)

// NewCmdCompletion returns the `completion` subcommand.
func NewCmdCompletion(f *factory.Factory, rootCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "completion [bash|zsh|fish]",
		Short:                 "Generate shell completion scripts",
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "zsh":
				return rootCmd.GenZshCompletion(f.IO.Out)
			case "bash":
				return rootCmd.GenBashCompletionV2(f.IO.Out, true)
			case "fish":
				return rootCmd.GenFishCompletion(f.IO.Out, true)
			default:
				fmt.Fprintf(f.IO.ErrOut, "unsupported shell: %s\n", args[0])
				return errors.New("unsupported shell")
			}
		},
	}
	return cmd
}
