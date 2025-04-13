package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// Runner executes git commands via the system git binary.
type Runner struct {
	Dir string // working directory; empty = inherit process CWD
}

// Run executes git with the given arguments in r.Dir.
// Returns trimmed combined stdout+stderr output.
func (r *Runner) Run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.Dir
	out, err := cmd.CombinedOutput()
	result := strings.TrimSpace(string(out))
	if err != nil {
		return result, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return result, nil
}

// RunInDir executes git with the given arguments, overriding Dir for this call only.
func (r *Runner) RunInDir(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	result := strings.TrimSpace(string(out))
	if err != nil {
		return result, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return result, nil
}
