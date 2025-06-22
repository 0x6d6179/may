package id

import (
	"errors"
	"fmt"

	"github.com/0x6d6179/may/internal/config"
	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/ui"
	"github.com/spf13/cobra"
)

// NewCmdIdAdd returns the `id add` subcommand.
func NewCmdIdAdd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "add a new git identity profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !f.IO.IsTerminal() {
				fmt.Fprintln(f.IO.ErrOut, "not a terminal")
				return errors.New("not a terminal")
			}

			cfg, err := f.Config()
			if err != nil {
				return err
			}

			opts := ui.RunOptions{In: f.IO.In, Out: f.IO.ErrOut}
			result, err := ui.RunForm(opts, ui.FormSpec{
				Title: "add identity",
				Fields: []ui.InputField{
					{Key: "name", Label: "profile name"},
					{Key: "username", Label: "git username"},
					{Key: "email", Label: "git email"},
					{Key: "gh_user", Label: "github username"},
				},
			})
			if errors.Is(err, ui.ErrAborted) {
				return nil
			}
			if err != nil {
				return err
			}

			cfg.Git.Profiles = append(cfg.Git.Profiles, config.Profile{
				Name:     result["name"],
				Username: result["username"],
				Email:    result["email"],
				GHUser:   result["gh_user"],
			})

			if err := config.Save(cfg); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "✓ added profile: %s\n", result["name"])
			return nil
		},
	}

	return cmd
}
