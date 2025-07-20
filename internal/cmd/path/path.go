package pathcmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/spf13/cobra"
)

func NewCmdPath(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "path",
		Short: "inspect and debug path",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listPath(f)
		},
	}

	cmd.AddCommand(
		newCmdPathDupes(f),
		newCmdPathFind(f),
	)

	return cmd
}

func listPath(f *factory.Factory) error {
	pathStr := os.Getenv("PATH")
	if pathStr == "" {
		fmt.Fprintf(f.IO.Out, "PATH is empty\n")
		return nil
	}

	entries := filepath.SplitList(pathStr)
	for i, entry := range entries {
		fmt.Fprintf(f.IO.Out, "%d. %s\n", i+1, entry)
	}

	return nil
}

func newCmdPathDupes(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "dupes",
		Short: "find duplicate path entries",
		RunE: func(cmd *cobra.Command, args []string) error {
			return findDupes(f)
		},
	}
}

func findDupes(f *factory.Factory) error {
	pathStr := os.Getenv("PATH")
	if pathStr == "" {
		fmt.Fprintf(f.IO.Out, "PATH is empty\n")
		return nil
	}

	entries := filepath.SplitList(pathStr)
	seen := make(map[string][]int)

	for i, entry := range entries {
		seen[entry] = append(seen[entry], i+1)
	}

	hasDupes := false
	for entry, indices := range seen {
		if len(indices) > 1 {
			hasDupes = true
			fmt.Fprintf(f.IO.ErrOut, "duplicate: %s (entries ", entry)
			for i, idx := range indices {
				if i > 0 {
					fmt.Fprintf(f.IO.ErrOut, ", ")
				}
				fmt.Fprintf(f.IO.ErrOut, "%d", idx)
			}
			fmt.Fprintf(f.IO.ErrOut, ")\n")
		}
	}

	if !hasDupes {
		fmt.Fprintf(f.IO.Out, "no duplicate path entries found\n")
	}

	return nil
}

func newCmdPathFind(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "find <name>",
		Short: "find command in path entries",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return findInPath(f, args[0])
		},
	}
}

func findInPath(f *factory.Factory, name string) error {
	pathStr := os.Getenv("PATH")
	if pathStr == "" {
		fmt.Fprintf(f.IO.ErrOut, "PATH is empty\n")
		return nil
	}

	entries := filepath.SplitList(pathStr)
	var matches []string

	for _, entry := range entries {
		fullPath := filepath.Join(entry, name)
		if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
			if isExecutable(fullPath) {
				matches = append(matches, fullPath)
			}
		}
	}

	if len(matches) == 0 {
		fmt.Fprintf(f.IO.ErrOut, "%s: not found in PATH\n", name)
		return nil
	}

	for _, match := range matches {
		fmt.Fprintf(f.IO.Out, "%s\n", match)
	}

	return nil
}

func isExecutable(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}

	mode := stat.Mode()
	if mode.IsDir() {
		return false
	}

	_, err = exec.LookPath(path)
	return err == nil
}
