package ws

import (
	"errors"
	"fmt"

	"github.com/0x6d6179/may/internal/config"
	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/ui"
	"github.com/0x6d6179/may/internal/workspace"
	"github.com/spf13/cobra"
)

type wsFlow struct {
	cfg        *config.Config
	nameByPath map[string]string
}

func (f *wsFlow) Start() ui.Step {
	return ui.NewLoading(ui.LoadingSpec[[]workspace.Workspace]{
		Title: "select workspace",
		Label: "loading...",
		Task:  func() ([]workspace.Workspace, error) { return workspace.List(f.cfg), nil },
	})
}

func (f *wsFlow) Next(result any) (ui.Step, bool, error) {
	switch v := result.(type) {
	case []workspace.Workspace:
		if len(v) == 0 {
			return nil, false, errors.New("no workspaces configured")
		}
		f.nameByPath = make(map[string]string, len(v))
		options := make([]ui.Option[string], len(v))
		for i, ws := range v {
			options[i] = ui.Option[string]{Label: ws.Name, Value: ws.Path}
			f.nameByPath[ws.Path] = ws.Name
		}
		return ui.NewSelectStep(ui.SelectSpec[string]{
			Title:   "select workspace",
			Options: options,
			Height:  10,
		}), false, nil
	case string:
		return nil, true, nil
	}
	return nil, true, nil
}

func NewCmdWs(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ws",
		Short: "workspace management",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			flow := &wsFlow{cfg: cfg}
			selected, err := ui.RunFlow[string](flow, ui.RunOptions{
				In: f.IO.In, Out: f.IO.ErrOut,
			})
			if errors.Is(err, ui.ErrAborted) {
				return nil
			}
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "✓ workspace switched to %s\n", flow.nameByPath[selected])
			if f.IO.IsTerminal() {
				fmt.Fprintln(f.IO.ErrOut, "→ shell integration not active · run: eval \"$(may shell init)\"")
			}
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
