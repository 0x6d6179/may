package wt

import (
	"errors"
	"fmt"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/git"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

func NewCmdWt(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wt",
		Short: "Git worktree manager",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !f.IO.IsTerminal() {
				fmt.Fprintln(f.IO.ErrOut, "not a terminal")
				return errors.New("not a terminal")
			}

			runner := &git.Runner{}
			worktrees, err := git.ListWorktrees(runner)
			if err != nil {
				return err
			}
			if len(worktrees) == 0 {
				fmt.Fprintln(f.IO.ErrOut, "no worktrees found")
				return nil
			}

			options := make([]huh.Option[string], len(worktrees))
			for i, wt := range worktrees {
				options[i] = huh.NewOption(wt.Branch, wt.Path)
			}

			var selected string
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("Select worktree").
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
