package wt

import (
	"errors"
	"fmt"
	"strings"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/fuzzy"
	"github.com/0x6d6179/may/internal/git"
	"github.com/0x6d6179/may/internal/ui"
	"github.com/charmbracelet/bubbletea"
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

		var cmds []tea.Cmd
		sem := make(chan struct{}, 5)

		for i, wt := range v {
			options[i] = ui.Option[string]{
				Label:       wt.Branch,
				Value:       wt.Path,
				Description: "getting worktree information...",
				Loading:     true,
			}
			f.branchByPath[wt.Path] = wt.Branch

			idx := i
			path := wt.Path

			cmds = append(cmds, func() tea.Msg {
				sem <- struct{}{}
				defer func() { <-sem }()
				info := git.GitOnlyStatus(path)
				return ui.OptionUpdateMsg{Index: idx, Description: info}
			})
		}
		return ui.NewSelectStep(ui.SelectSpec[string]{
			Title:   "select worktree",
			Options: options,
			Height:  10,
			InitCmd: tea.Batch(cmds...),
		}), false, nil
	case string:
		return nil, true, nil
	}
	return nil, true, nil
}

func NewCmdWt(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "wt",
		Aliases: []string{"worktree"},
		Short:   "git worktree manager",
		RunE: func(cmd *cobra.Command, args []string) error {
			runner := &git.Runner{}

			if !git.IsGitRepo(".") {
				return fmt.Errorf("not a git repository — run this from inside a git project")
			}

			var selected string
			var selectedName string

			if len(args) > 0 {
				query := strings.Join(args, " ")
				worktrees, err := git.ListWorktrees(runner)
				if err != nil {
					return err
				}
				if len(worktrees) == 0 {
					return errors.New("no worktrees found")
				}

				var bestMatch git.Worktree
				var bestScore float64

				for _, wt := range worktrees {
					score := fuzzy.Score(query, wt.Branch)
					if score > bestScore {
						bestScore = score
						bestMatch = wt
					}
				}

				if bestScore >= 0.8 {
					selected = bestMatch.Path
					selectedName = bestMatch.Branch
				} else {
					return fmt.Errorf("no worktree found matching %q", query)
				}
			} else {
				flow := &wtFlow{runner: runner}
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
				selectedName = flow.branchByPath[selected]
			}

			fmt.Fprintf(f.IO.ErrOut, "✓ jumped to %s\n", selectedName)
			if f.IO.IsTerminal() {
				fmt.Fprintln(f.IO.ErrOut, "→ run: may shell configure to enable automatic cd")
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
