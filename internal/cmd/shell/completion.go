package shell

import (
	"errors"
	"fmt"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/spf13/cobra"
)

func NewCmdShellCompletion(f *factory.Factory, rootCmd *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:                   "completion [bash|zsh|fish]",
		Short:                 "generate shell completion scripts",
		Args:                  cobra.MaximumNArgs(1),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			shell, err := resolveShell(args)
			if err != nil {
				return err
			}
			switch shell {
			case "zsh":
				return rootCmd.GenZshCompletion(f.IO.Out)
			case "bash":
				return rootCmd.GenBashCompletionV2(f.IO.Out, true)
			case "fish":
				return rootCmd.GenFishCompletion(f.IO.Out, true)
			default:
				fmt.Fprintf(f.IO.ErrOut, "unsupported shell: %s\n", shell)
				return errors.New("unsupported shell")
			}
		},
	}
}
