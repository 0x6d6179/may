package wt

import (
	"errors"
	"fmt"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/git"
	"github.com/0x6d6179/may/internal/ui"
	"github.com/spf13/cobra"
)

type wtFlow struct {
	runner       *git.Runner
	branchByPath map[string]string
}

func (f *wtFlow) Start() ui.Step {
	return ui.NewLoading(ui.LoadingSpec[[]git.Worktree]{
		Title: "select worktree",
		Label: "loading...",
		Task:  func() ([]git.Worktree, error) { return git.ListWorktrees(f.runner) },
	})
}

func (f *wtFlow) Next(result any) (ui.Step, bool, error) {
	switch v := result.(type) {
	case []git.Worktree:
		if len(v) == 0 {
			return nil, false, errors.New("no worktrees found")
		}
		f.branchByPath = make(map[string]string, len(v))
		options := make([]ui.Option[string], len(v))
		for i, wt := range v {
			options[i] = ui.Option[string]{Label: wt.Branch, Value: wt.Path}
			f.branchByPath[wt.Path] = wt.Branch
		}
		return ui.NewSelectStep(ui.SelectSpec[string]{
			Title:   "select worktree",
			Options: options,
			Height:  10,
		}), false, nil
	case string:
		return nil, true, nil
	}
	return nil, true, nil
}

func NewCmdWt(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wt",
		Short: "git worktree manager",
		RunE: func(cmd *cobra.Command, args []string) error {
			runner := &git.Runner{}

			flow := &wtFlow{runner: runner}
			selected, err := ui.RunFlow[string](flow, ui.RunOptions{
				In: f.IO.In, Out: f.IO.ErrOut,
			})
			if errors.Is(err, ui.ErrAborted) {
				return nil
			}
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "✓ jumped to %s\n", flow.branchByPath[selected])
			if f.IO.IsTerminal() {
				fmt.Fprintln(f.IO.ErrOut, "→ shell integration not active · run: eval \"$(may shell init)\"")
			}
			if !f.IO.IsTerminal() {
				fmt.Fprintln(f.IO.Out, selected)
			}
			return nil
		},
	}

	cmd.AddCommand(NewCmdWtNew(f))
	cmd.AddCommand(NewCmdWtRm(f))
	cmd.AddCommand(NewCmdWtClose(f))
	cmd.AddCommand(NewCmdWtList(f))
	cmd.AddCommand(NewCmdWtJump(f))
	cmd.AddCommand(NewCmdWtEnv(f))
	cmd.AddCommand(NewCmdWtPull(f))
	cmd.AddCommand(NewCmdWtCp(f))

	return cmd
}
