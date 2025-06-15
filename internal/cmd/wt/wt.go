package wt

import (
	"fmt"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/git"
	"github.com/0x6d6179/may/internal/ui"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

func NewCmdWt(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wt",
		Short: "git worktree manager",
		RunE: func(cmd *cobra.Command, args []string) error {
			runner := &git.Runner{}

			stop := ui.Spinner(f.IO.ErrOut, "loading...")
			worktrees, err := git.ListWorktrees(runner)
			stop()

			if err != nil {
				return err
			}
			if len(worktrees) == 0 {
				fmt.Fprintln(f.IO.ErrOut, "no worktrees found")
				return nil
			}

			branchByPath := make(map[string]string, len(worktrees))
			options := make([]huh.Option[string], len(worktrees))
			for i, wt := range worktrees {
				options[i] = huh.NewOption(wt.Branch, wt.Path)
				branchByPath[wt.Path] = wt.Branch
			}

			ui.Header(f.IO.ErrOut, "select worktree")
			var selected string
			form := ui.NewForm(
				huh.NewGroup(
					ui.NewSelect[string]().
						Title("select worktree").
						Options(options...).
						Value(&selected),
				),
			).WithHeight(10)

			if err := form.Run(); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "✓ jumped to %s\n", branchByPath[selected])
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
