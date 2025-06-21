package commit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/0x6d6179/may/internal/ai"
	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/git"
	"github.com/0x6d6179/may/internal/ui"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func NewCmdCommit(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "commit",
		Short: "ai conventional commit",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !term.IsTerminal(int(os.Stdin.Fd())) {
				fmt.Fprintln(f.IO.ErrOut, "not a terminal")
				return errors.New("not a terminal")
			}

			cfg, err := f.Config()
			if err != nil {
				return err
			}

			opts := ui.RunOptions{In: f.IO.In, Out: f.IO.ErrOut}
			runner := &git.Runner{}

			staged, _ := runner.Run("diff", "--cached")
			diff := staged

			if diff == "" {
				unstaged, _ := runner.Run("diff")
				if unstaged == "" {
					fmt.Fprintln(f.IO.ErrOut, "nothing to commit")
					return errors.New("nothing to commit")
				}

				stageAll, err := ui.RunConfirm(opts, ui.ConfirmSpec{Title: "stage all changes?"})
				if errors.Is(err, ui.ErrAborted) {
					return nil
				}
				if err != nil {
					return err
				}

				if stageAll {
					if _, err := runner.Run("add", "-A"); err != nil {
						return err
					}
					diff, _ = runner.Run("diff", "--cached")
				} else {
					diff = unstaged
				}
			}

			aiClient := &ai.Client{
				BaseURL: cfg.AI.BaseURL,
				APIKey:  cfg.AI.APIKey,
				Model:   cfg.AI.Model,
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer cancel()

			result, err := ui.RunFlow[any](&commitLoadFlow{
				ctx:    ctx,
				diff:   diff,
				client: aiClient,
			}, opts)
			if errors.Is(err, ui.ErrAborted) {
				return nil
			}
			if err != nil {
				return err
			}

			var msgs *ai.CommitMessages
			var aiErr error
			switch v := result.(type) {
			case *ai.CommitMessages:
				msgs = v
			case error:
				aiErr = v
			}

			var selectOpts []ui.Option[string]
			if aiErr == nil && msgs != nil {
				selectOpts = []ui.Option[string]{
					{Label: msgs.Primary, Value: msgs.Primary},
					{Label: msgs.Alt1, Value: msgs.Alt1},
					{Label: msgs.Alt2, Value: msgs.Alt2},
					{Label: msgs.Alt3, Value: msgs.Alt3},
					{Label: "Enter custom message", Value: "__custom__"},
					{Label: "Abort", Value: "__abort__"},
				}
			} else {
				selectOpts = []ui.Option[string]{
					{Label: "Enter custom message", Value: "__custom__"},
					{Label: "Abort", Value: "__abort__"},
				}
			}

			selected, err := ui.RunSelect(opts, ui.SelectSpec[string]{
				Title:   "choose commit message",
				Options: selectOpts,
			})
			if errors.Is(err, ui.ErrAborted) {
				return nil
			}
			if err != nil {
				return err
			}

			if selected == "__abort__" {
				return nil
			}

			if selected == "__custom__" {
				custom, err := ui.RunInput(opts, ui.InputSpec{Title: "commit message"})
				if errors.Is(err, ui.ErrAborted) {
					return nil
				}
				if err != nil {
					return err
				}
				selected = custom
			}

			out, err := runner.Run("-c", "user.name=may", "-c", "user.email=ryana@ryanaque.com", "commit", "-m", selected)
			if err != nil {
				return err
			}
			_ = out

			fmt.Fprintf(f.IO.ErrOut, "✓ committed: %s\n", selected)
			return nil
		},
	}
}

type commitLoadFlow struct {
	ctx    context.Context
	diff   string
	client *ai.Client
}

func (f *commitLoadFlow) Start() ui.Step {
	return ui.NewLoading[*ai.CommitMessages](ui.LoadingSpec[*ai.CommitMessages]{
		Title: "commit",
		Label: "generating commit messages…",
		Task: func() (*ai.CommitMessages, error) {
			return f.client.GenerateCommitMessages(f.ctx, f.diff)
		},
	})
}

func (f *commitLoadFlow) Next(_ any) (ui.Step, bool, error) {
	return nil, true, nil
}
