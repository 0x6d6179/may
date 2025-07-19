package j

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/fuzzy"
	"github.com/spf13/cobra"
)

const fuzzyThreshold = 0.6

func NewCmdJ(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "j <path>",
		Short: "smart directory jump",
		Long:  "jump to a directory using exact or fuzzy matching on each path segment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]

			resolved, err := resolve(target)
			if err != nil {
				return err
			}

			fmt.Fprintln(f.IO.Out, resolved)
			return nil
		},
	}
	return cmd
}

func resolve(target string) (string, error) {
	if filepath.IsAbs(target) {
		cleaned := filepath.Clean(target)
		if isDir(cleaned) {
			return cleaned, nil
		}
		return "", fmt.Errorf("not a directory: %s", cleaned)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	if target == "." {
		return cwd, nil
	}

	segments := strings.Split(filepath.ToSlash(target), "/")
	current := cwd

	for _, seg := range segments {
		if seg == "" {
			continue
		}

		next, err := resolveSegment(current, seg)
		if err != nil {
			return "", err
		}
		current = next
	}

	return current, nil
}

func resolveSegment(base, seg string) (string, error) {
	if seg == ".." {
		parent := filepath.Dir(base)
		if parent == base {
			return base, nil
		}
		return parent, nil
	}

	if seg == "." {
		return base, nil
	}

	exact := filepath.Join(base, seg)
	if isDir(exact) {
		return exact, nil
	}

	return fuzzyMatchDir(base, seg)
}

func fuzzyMatchDir(base, query string) (string, error) {
	entries, err := os.ReadDir(base)
	if err != nil {
		return "", fmt.Errorf("read directory %s: %w", base, err)
	}

	type candidate struct {
		name  string
		score float64
	}

	var candidates []candidate
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		score := fuzzy.Score(query, name)
		if score >= fuzzyThreshold {
			candidates = append(candidates, candidate{name: name, score: score})
		}
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("no directory matching %q in %s", query, base)
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	return filepath.Join(base, candidates[0].name), nil
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
