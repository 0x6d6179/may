package ws

import (
	"errors"
	"fmt"
	"time"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/workspace"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

func NewCmdWs(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ws",
		Short: "Workspace management",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !f.IO.IsTerminal() {
				fmt.Fprintln(f.IO.ErrOut, "not a terminal")
				return errors.New("not a terminal")
			}

			cfg, err := f.Config()
			if err != nil {
				return err
			}

			done := make(chan struct{})
			go func() {
				frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
				i := 0
				for {
					select {
					case <-done:
						fmt.Fprint(f.IO.ErrOut, "\r                    \r")
						return
					default:
						fmt.Fprintf(f.IO.ErrOut, "\r%s loading...", frames[i%len(frames)])
						i++
						time.Sleep(80 * time.Millisecond)
					}
				}
			}()
			workspaces := workspace.List(cfg)
			close(done)
			time.Sleep(10 * time.Millisecond)

			if len(workspaces) == 0 {
				fmt.Fprintln(f.IO.ErrOut, "no workspaces configured")
				return nil
			}

			options := make([]huh.Option[string], len(workspaces))
			for i, ws := range workspaces {
				options[i] = huh.NewOption(ws.Name, ws.Path)
			}

			var selected string
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Filtering(true).
						Title("Select workspace").
						Options(options...).
						Value(&selected),
				),
			).WithHeight(10)

			if err := form.Run(); err != nil {
				return err
			}

			fmt.Fprintln(f.IO.Out, selected)
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
