package branch

import (
	"errors"
	"fmt"
	"strings"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/fuzzy"
	"github.com/0x6d6179/may/internal/git"
	"github.com/spf13/cobra"
)

func NewCmdBranch(f *factory.Factory) *cobra.Command {
	var (
		all    bool
		delete string
	)

	cmd := &cobra.Command{
		Use:     "branch [name]",
		Aliases: []string{"br"},
		Short:   "list or switch git branches",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !git.IsGitRepo(".") {
				return errors.New("not a git repository")
			}

			runner := &git.Runner{Dir: "."}

			if delete != "" {
				_, err := runner.Run("branch", "-d", delete)
				if err != nil {
					return fmt.Errorf("delete branch %s: %w", delete, err)
				}
				fmt.Fprintf(f.IO.ErrOut, "deleted branch %s\n", delete)
				return nil
			}

			if len(args) == 0 {
				current, err := git.CurrentBranch(runner)
				if err != nil {
					return err
				}

				listArgs := []string{"branch", "--list", "--sort=-committerdate"}
				if all {
					listArgs = append(listArgs, "--all")
				}

				output, err := runner.Run(listArgs...)
				if err != nil {
					return err
				}

				lines := strings.Split(strings.TrimSpace(output), "\n")
				for _, line := range lines {
					if line == "" {
						continue
					}
					line = strings.TrimPrefix(line, "* ")
					if line == current {
						fmt.Fprintf(f.IO.Out, "* %s\n", line)
					} else {
						fmt.Fprintf(f.IO.Out, "  %s\n", line)
					}
				}
				return nil
			}

			query := strings.Join(args, " ")
			listArgs := []string{"branch", "--list", "--sort=-committerdate"}
			if all {
				listArgs = append(listArgs, "--all")
			}

			output, err := runner.Run(listArgs...)
			if err != nil {
				return err
			}

			lines := strings.Split(strings.TrimSpace(output), "\n")
			if len(lines) == 0 {
				return errors.New("no branches found")
			}

			var bestMatch string
			var bestScore float64

			for _, line := range lines {
				if line == "" {
					continue
				}
				branchName := strings.TrimPrefix(strings.TrimSpace(line), "* ")
				score := fuzzy.Score(query, branchName)
				if score > bestScore {
					bestScore = score
					bestMatch = branchName
				}
			}

			if bestScore < 0.8 {
				return fmt.Errorf("no branch matching %q", query)
			}

			_, err = runner.Run("checkout", bestMatch)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "switched to %s\n", bestMatch)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&all, "all", "a", false, "include remote branches")
	cmd.Flags().StringVarP(&delete, "delete", "d", "", "delete a branch")

	return cmd
}
