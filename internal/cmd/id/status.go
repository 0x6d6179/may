package id

import (
	"fmt"
	"os"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/git"
	"github.com/0x6d6179/may/internal/identity"
	"github.com/spf13/cobra"
)

// NewCmdIdStatus returns the `id status` subcommand.
func NewCmdIdStatus(f *factory.Factory) *cobra.Command {
	var apply bool
	var quiet bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "show or apply the git identity for the current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			if !git.IsGitRepo(cwd) {
				return nil
			}

			cfg, err := f.Config()
			if err != nil {
				return err
			}

			profile, ok := identity.ResolveProfile(cfg, cwd)
			if !ok {
				if apply {
					return nil
				}
				return nil
			}

			if apply {
				runner := &git.Runner{}
				if err := git.SetLocalIdentity(runner, cwd, profile.Username, profile.Email); err != nil {
					return err
				}
				if !quiet {
					fmt.Fprintf(f.IO.ErrOut, "✓ applied: %s <%s>\n", profile.Username, profile.Email)
				}
				return nil
			}

			if !quiet {
				fmt.Fprintf(f.IO.ErrOut, "name:    %s\nemail:   %s\nprofile: %s\n", profile.Username, profile.Email, profile.Name)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&apply, "apply", false, "Apply the resolved identity to the local git config")
	cmd.Flags().BoolVar(&quiet, "quiet", false, "Suppress all output")

	return cmd
}
