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
	"github.com/charmbracelet/huh"
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

			ui.Header(f.IO.ErrOut, "commit")
			runner := &git.Runner{}

			staged, _ := runner.Run("diff", "--cached")
			diff := staged

			if diff == "" {
				unstaged, _ := runner.Run("diff")
				if unstaged == "" {
					fmt.Fprintln(f.IO.ErrOut, "nothing to commit")
					return errors.New("nothing to commit")
				}

				var stageAll bool
				if err := ui.NewForm(
					huh.NewGroup(
						huh.NewConfirm().
							Title("stage all changes?").
							Value(&stageAll),
					),
				).Run(); err != nil {
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

			msgs, aiErr := aiClient.GenerateCommitMessages(ctx, diff)

			selected, err := selectCommitMessage(msgs, aiErr)
			if err != nil {
				return err
			}

			if selected == "__abort__" {
				return nil
			}

			if selected == "__custom__" {
				var custom string
				if err := ui.NewForm(
					huh.NewGroup(
						huh.NewInput().
							Title("commit message").
							Value(&custom),
					),
				).Run(); err != nil {
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

func selectCommitMessage(msgs *ai.CommitMessages, aiErr error) (string, error) {
	var selected string

	if aiErr != nil || msgs == nil {
		if err := ui.NewForm(
			huh.NewGroup(
				ui.NewSelect[string]().
					Title("choose commit message").
					Options(
						huh.NewOption("Enter custom message", "__custom__"),
						huh.NewOption("Abort", "__abort__"),
					).
					Value(&selected),
			),
		).Run(); err != nil {
			return "", err
		}
		return selected, nil
	}

	if err := ui.NewForm(
		huh.NewGroup(
			ui.NewSelect[string]().
				Title("choose commit message").
				Options(
					huh.NewOption(msgs.Primary, msgs.Primary),
					huh.NewOption(msgs.Alt1, msgs.Alt1),
					huh.NewOption(msgs.Alt2, msgs.Alt2),
					huh.NewOption(msgs.Alt3, msgs.Alt3),
					huh.NewOption("Enter custom message", "__custom__"),
					huh.NewOption("Abort", "__abort__"),
				).
				Value(&selected),
		),
	).Run(); err != nil {
		return "", err
	}

	return selected, nil
}
