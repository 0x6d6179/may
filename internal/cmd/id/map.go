package id

import (
	"fmt"
	"os"

	"github.com/0x6d6179/may/internal/config"
	"github.com/0x6d6179/may/internal/factory"
	"github.com/spf13/cobra"
)

// NewCmdIdMap returns the `id map` subcommand.
func NewCmdIdMap(f *factory.Factory) *cobra.Command {
	var path string
	var profileName string

	cmd := &cobra.Command{
		Use:   "map",
		Short: "map a directory path to a git identity profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			if path == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				path = cwd
			}

			cfg, err := f.Config()
			if err != nil {
				return err
			}

			cfg.Git.Mappings = append(cfg.Git.Mappings, config.Mapping{
				Path:    path,
				Profile: profileName,
			})

			if err := config.Save(cfg); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "✓ mapped %s → %s\n", path, profileName)
			return nil
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "Directory path to map (default: current directory)")
	cmd.Flags().StringVar(&profileName, "profile", "", "Profile name to map to")
	_ = cmd.MarkFlagRequired("profile")

	cmd.RegisterFlagCompletionFunc("profile", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) { //nolint:errcheck
		cfg, err := f.Config()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		names := make([]string, len(cfg.Git.Profiles))
		for i, p := range cfg.Git.Profiles {
			names[i] = p.Name
		}
		return names, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}
