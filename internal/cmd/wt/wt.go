package wt

import (
	"errors"
	"fmt"
	"time"

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
			worktrees, err := git.ListWorktrees(runner)
			close(done)
			time.Sleep(10 * time.Millisecond)

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
						Filtering(true).
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
