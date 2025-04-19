package id

import (
	"errors"
	"fmt"

	"github.com/0x6d6179/may/internal/config"
	"github.com/0x6d6179/may/internal/factory"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// NewCmdIdAdd returns the `id add` subcommand.
func NewCmdIdAdd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new git identity profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !f.IO.IsTerminal() {
				fmt.Fprintln(f.IO.ErrOut, "not a terminal")
				return errors.New("not a terminal")
			}

			cfg, err := f.Config()
			if err != nil {
				return err
			}

			var name, username, email, ghUser string

			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Profile name").
						Value(&name),
					huh.NewInput().
						Title("Git username").
						Value(&username),
					huh.NewInput().
						Title("Git email").
						Value(&email),
					huh.NewInput().
						Title("GitHub username").
						Value(&ghUser),
				),
			)

			if err := form.Run(); err != nil {
				return err
			}

			cfg.Git.Profiles = append(cfg.Git.Profiles, config.Profile{
				Name:     name,
				Username: username,
				Email:    email,
				GHUser:   ghUser,
			})

			if err := config.Save(cfg); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "added profile %q\n", name)
			return nil
		},
	}

	return cmd
}
