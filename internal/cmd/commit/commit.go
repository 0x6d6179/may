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
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func NewCmdCommit(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "commit",
		Short: "AI conventional commit",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !term.IsTerminal(int(os.Stdin.Fd())) {
				fmt.Fprintln(f.IO.ErrOut, "not a terminal")
				return errors.New("not a terminal")
			}

			cfg, err := f.Config()
			if err != nil {
				return err
			}

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
				if err := huh.NewForm(
					huh.NewGroup(
						huh.NewConfirm().
							Title("Stage all changes?").
							Value(&stageAll),
					),
				).WithHeight(10).Run(); err != nil {
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
				if err := huh.NewForm(
					huh.NewGroup(
						huh.NewInput().
							Title("Commit message:").
							Value(&custom),
					),
				).WithHeight(10).Run(); err != nil {
					return err
				}
				selected = custom
			}

			out, err := runner.Run("-c", "user.name=may", "-c", "user.email=ryana@ryanaque.com", "commit", "-m", selected)
			if err != nil {
				return err
			}

			fmt.Fprintln(f.IO.ErrOut, out)
			return nil
		},
	}
}

func selectCommitMessage(msgs *ai.CommitMessages, aiErr error) (string, error) {
	var selected string

	if aiErr != nil || msgs == nil {
		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Choose commit message").
					Options(
						huh.NewOption("Enter custom message", "__custom__"),
						huh.NewOption("Abort", "__abort__"),
					).
					Value(&selected),
			),
		).WithHeight(10).Run(); err != nil {
			return "", err
		}
		return selected, nil
	}

	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose commit message").
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
	).WithHeight(10).Run(); err != nil {
		return "", err
	}

	return selected, nil
}
