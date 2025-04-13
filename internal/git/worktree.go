package git

import (
	"strings"
)

// Worktree represents a git worktree entry.
type Worktree struct {
	Path   string
	Branch string
	IsMain bool
}

// MainWorktreePath returns the path of the main worktree by parsing `git worktree list --porcelain`.
func MainWorktreePath(r *Runner) (string, error) {
	out, err := r.Run("worktree", "list", "--porcelain")
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "worktree ") {
			return strings.TrimPrefix(line, "worktree "), nil
		}
	}
	return "", nil
}

// ListWorktrees parses `git worktree list --porcelain` and returns all worktrees.
func ListWorktrees(r *Runner) ([]Worktree, error) {
	out, err := r.Run("worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	return parseWorktrees(out), nil
}

// parseWorktrees parses the porcelain output of `git worktree list --porcelain`.
func parseWorktrees(out string) []Worktree {
	var worktrees []Worktree
	var current Worktree
	first := true

	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			if current.Path != "" {
				worktrees = append(worktrees, current)
				current = Worktree{}
			}
			continue
		}
		if strings.HasPrefix(line, "worktree ") {
			current.Path = strings.TrimPrefix(line, "worktree ")
			if first {
				current.IsMain = true
				first = false
			}
		} else if strings.HasPrefix(line, "branch ") {
			ref := strings.TrimPrefix(line, "branch ")
			current.Branch = strings.TrimPrefix(ref, "refs/heads/")
		} else if line == "detached" {
			current.Branch = "HEAD"
		}
	}
	if current.Path != "" {
		worktrees = append(worktrees, current)
	}
	return worktrees
}

// IsInsideWorktree returns true if path is a linked worktree (not the main worktree).
func IsInsideWorktree(r *Runner, path string) bool {
	main, err := MainWorktreePath(r)
	if err != nil || main == "" {
		return false
	}
	return path != main
}
