package id

import (
	"fmt"
	"os"

	"github.com/0x6d6179/may/internal/config"
	"github.com/0x6d6179/may/internal/factory"
	"github.com/spf13/cobra"
)

// NewCmdIdUnmap returns the `id unmap` subcommand.
func NewCmdIdUnmap(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unmap [path]",
		Short: "Remove a directory-to-profile mapping",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := ""
			if len(args) > 0 {
				path = args[0]
			}
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

			filtered := cfg.Git.Mappings[:0]
			removed := false
			for _, m := range cfg.Git.Mappings {
				if m.Path == path {
					removed = true
					continue
				}
				filtered = append(filtered, m)
			}

			if !removed {
				fmt.Fprintf(f.IO.ErrOut, "no mapping found for %s\n", path)
				return nil
			}

			cfg.Git.Mappings = filtered

			if err := config.Save(cfg); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "removed mapping for %s\n", path)
			return nil
		},
	}

	return cmd
}
