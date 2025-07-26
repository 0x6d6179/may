package stash

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/git"
	"github.com/spf13/cobra"
)

func NewCmdStash(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stash",
		Short: "manage git stashes",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !git.IsGitRepo(".") {
				return errors.New("not in a git repository")
			}

			if len(args) == 0 {
				return listStashes(f)
			}

			subcommand := args[0]
			switch subcommand {
			case "save":
				var message string
				if len(args) > 1 {
					message = strings.Join(args[1:], " ")
				}
				return saveStash(f, message)
			case "pop":
				var index int
				if len(args) > 1 {
					idx, err := strconv.Atoi(args[1])
					if err != nil {
						return fmt.Errorf("invalid stash index: %s", args[1])
					}
					index = idx
				} else {
					index = 0
				}
				return popStash(f, index)
			case "drop":
				if len(args) < 2 {
					return errors.New("drop requires a stash index")
				}
				index, err := strconv.Atoi(args[1])
				if err != nil {
					return fmt.Errorf("invalid stash index: %s", args[1])
				}
				return dropStash(f, index)
			case "show":
				var index int
				if len(args) > 1 {
					idx, err := strconv.Atoi(args[1])
					if err != nil {
						return fmt.Errorf("invalid stash index: %s", args[1])
					}
					index = idx
				} else {
					index = 0
				}
				return showStash(f, index)
			default:
				return fmt.Errorf("unknown stash subcommand: %s", subcommand)
			}
		},
	}
	return cmd
}

func listStashes(f *factory.Factory) error {
	runner := &git.Runner{}
	output, err := runner.Run("stash", "list")
	if err != nil {
		return err
	}

	if output == "" {
		fmt.Fprintln(f.IO.ErrOut, "no stashes")
		return nil
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		formatted := formatStashLine(line)
		fmt.Fprintln(f.IO.Out, formatted)
	}

	return nil
}

func saveStash(f *factory.Factory, message string) error {
	runner := &git.Runner{}

	if message == "" {
		message = "wip"
	}

	_, err := runner.Run("stash", "push", "-m", message)
	if err != nil {
		return err
	}

	fmt.Fprintf(f.IO.ErrOut, "stashed: %s\n", message)
	return nil
}

func popStash(f *factory.Factory, index int) error {
	runner := &git.Runner{}
	_, err := runner.Run("stash", "pop", fmt.Sprintf("stash@{%d}", index))
	if err != nil {
		return err
	}

	fmt.Fprintf(f.IO.ErrOut, "popped stash@{%d}\n", index)
	return nil
}

func dropStash(f *factory.Factory, index int) error {
	runner := &git.Runner{}
	_, err := runner.Run("stash", "drop", fmt.Sprintf("stash@{%d}", index))
	if err != nil {
		return err
	}

	fmt.Fprintf(f.IO.ErrOut, "dropped stash@{%d}\n", index)
	return nil
}

func showStash(f *factory.Factory, index int) error {
	runner := &git.Runner{}
	output, err := runner.Run("stash", "show", "-p", fmt.Sprintf("stash@{%d}", index))
	if err != nil {
		return err
	}

	fmt.Fprintln(f.IO.Out, output)
	return nil
}

func formatStashLine(line string) string {
	re := regexp.MustCompile(`^stash@\{(\d+)\}:\s+(.+)$`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 3 {
		return line
	}

	index := matches[1]
	rest := matches[2]

	timeRe := regexp.MustCompile(`^WIP on (.+?):\s+(.+)$|^On (.+?):\s+(.+)$`)
	timeMatches := timeRe.FindStringSubmatch(rest)

	if len(timeMatches) >= 3 {
		if timeMatches[1] != "" {
			branch := timeMatches[1]
			message := timeMatches[2]
			return fmt.Sprintf("%-3s  on %s: %s", index, branch, message)
		} else if timeMatches[3] != "" {
			branch := timeMatches[3]
			message := timeMatches[4]
			return fmt.Sprintf("%-3s  on %s: %s", index, branch, message)
		}
	}

	return fmt.Sprintf("%-3s  %s", index, rest)
}
