package id

import (
	"errors"
	"fmt"
	"os"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/git"
	"github.com/spf13/cobra"
)

// NewCmdIdSwitch returns the `id switch` subcommand.
func NewCmdIdSwitch(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "switch <profile-name>",
		Short: "Switch the git identity for the current directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]

			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			cfg, err := f.Config()
			if err != nil {
				return err
			}

			var found *struct{ username, email string }
			for _, p := range cfg.Git.Profiles {
				if p.Name == profileName {
					found = &struct{ username, email string }{p.Username, p.Email}
					break
				}
			}
			if found == nil {
				return fmt.Errorf("profile %q not found", profileName)
			}

			if !git.IsGitRepo(cwd) {
				return errors.New("not a git repository")
			}

			runner := &git.Runner{}
			if err := git.SetLocalIdentity(runner, cwd, found.username, found.email); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "switched to %s (%s <%s>)\n", profileName, found.username, found.email)
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			cfg, err := f.Config()
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			names := make([]string, len(cfg.Git.Profiles))
			for i, p := range cfg.Git.Profiles {
				names[i] = p.Name
			}
			return names, cobra.ShellCompDirectiveNoFileComp
		},
	}

	return cmd
}
