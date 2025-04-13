package git

import (
	"regexp"
	"strings"
)

var (
	reInvalidChars    = regexp.MustCompile(`[^a-zA-Z0-9_.\-]`)
	reConsecutiveDash = regexp.MustCompile(`-{2,}`)
)

// BranchSanitize transforms a branch name into a git-safe identifier:
// 1. Replace '/' with '-'
// 2. Remove chars not in [a-zA-Z0-9_.-]
// 3. Collapse consecutive '-' into single '-'
// 4. Trim leading/trailing '-'
func BranchSanitize(branch string) string {
	s := strings.ReplaceAll(branch, "/", "-")
	s = reInvalidChars.ReplaceAllString(s, "")
	s = reConsecutiveDash.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// CurrentBranch returns the name of the current git branch.
func CurrentBranch(r *Runner) (string, error) {
	return r.Run("rev-parse", "--abbrev-ref", "HEAD")
}
