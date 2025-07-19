package ws

import (
	"errors"
	"fmt"
	"strings"

	"github.com/0x6d6179/may/internal/config"
	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/fuzzy"
	"github.com/0x6d6179/may/internal/git"
	"github.com/0x6d6179/may/internal/ui"
	"github.com/0x6d6179/may/internal/workspace"
	"github.com/charmbracelet/bubbletea"
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

		var cmds []tea.Cmd
		sem := make(chan struct{}, 5)

		for i, ws := range v {
			options[i] = ui.Option[string]{
				Label:       ws.Name,
				Value:       ws.Path,
				Description: "getting workspace information...",
				Loading:     true,
			}
			f.nameByPath[ws.Path] = ws.Name

			idx := i
			path := ws.Path

			cmds = append(cmds, func() tea.Msg {
				sem <- struct{}{}
				defer func() { <-sem }()
				info := git.ShortStatus(path)
				return ui.OptionUpdateMsg{Index: idx, Description: info}
			})
		}
		return ui.NewSelectStep(ui.SelectSpec[string]{
			Title:   "select workspace",
			Options: options,
			Height:  10,
			InitCmd: tea.Batch(cmds...),
		}), false, nil
	case string:
		return nil, true, nil
	}
	return nil, true, nil
}

func NewCmdWs(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ws",
		Aliases: []string{"workspace", "mayspace", "space"},
		Short:   "workspace management",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			var selected string
			var selectedName string

			if len(args) > 0 {
				query := strings.Join(args, " ")
				workspaces := workspace.List(cfg)
				if len(workspaces) == 0 {
					return errors.New("no workspaces configured")
				}

				var bestMatch workspace.Workspace
				var bestScore float64

				for _, ws := range workspaces {
					score := fuzzy.Score(query, ws.Name)
					if score > bestScore {
						bestScore = score
						bestMatch = ws
					}
				}

				if bestScore >= 0.8 {
					selected = bestMatch.Path
					selectedName = bestMatch.Name
				} else {
					return fmt.Errorf("no workspace found matching %q", query)
				}
			} else {
				flow := &wsFlow{cfg: cfg}
				var runErr error
				selected, runErr = ui.RunFlow[string](flow, ui.RunOptions{
					In: f.IO.In, Out: f.IO.ErrOut,
				})
				if errors.Is(runErr, ui.ErrAborted) {
					return nil
				}
				if runErr != nil {
					return runErr
				}
				selectedName = flow.nameByPath[selected]
			}

			fmt.Fprintf(f.IO.ErrOut, "✓ workspace switched to %s\n", selectedName)
			if f.IO.IsTerminal() {
				fmt.Fprintln(f.IO.ErrOut, "→ run: may shell configure to enable automatic cd")
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
