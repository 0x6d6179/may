package ws

import (
	"errors"
	"fmt"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/ui"
	"github.com/0x6d6179/may/internal/workspace"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

func NewCmdWs(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ws",
		Short: "workspace management",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !f.IO.IsTerminal() {
				fmt.Fprintln(f.IO.ErrOut, "not a terminal")
				return errors.New("not a terminal")
			}

			cfg, err := f.Config()
			if err != nil {
				return err
			}

			stop := ui.Spinner(f.IO.ErrOut, "loading...")
			workspaces := workspace.List(cfg)
			stop()

			if len(workspaces) == 0 {
				fmt.Fprintln(f.IO.ErrOut, "no workspaces configured")
				return nil
			}

			nameByPath := make(map[string]string, len(workspaces))
			options := make([]huh.Option[string], len(workspaces))
			for i, ws := range workspaces {
				options[i] = huh.NewOption(ws.Name, ws.Path)
				nameByPath[ws.Path] = ws.Name
			}

			ui.Header(f.IO.ErrOut, "select workspace")
			var selected string
			form := ui.NewForm(
				huh.NewGroup(
					ui.NewSelect[string]().
						Options(options...).
						Value(&selected),
				),
			)

			if err := form.Run(); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "✓ workspace switched to %s\n", nameByPath[selected])
			if !f.IO.IsTerminal() {
				fmt.Fprintln(f.IO.Out, selected)
			}
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			cfg, err := f.Config()
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			workspaces := workspace.List(cfg)
			names := make([]string, len(workspaces))
			for i, ws := range workspaces {
				names[i] = ws.Name
			}
			return names, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.AddCommand(NewCmdWsNew(f))
	cmd.AddCommand(NewCmdWsList(f))
	cmd.AddCommand(NewCmdWsSearch(f))
	cmd.AddCommand(NewCmdWsRoot(f))

	return cmd
}
