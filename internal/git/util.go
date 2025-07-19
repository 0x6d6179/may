package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// IsGitRepo returns true if path is inside a git repository.
func IsGitRepo(path string) bool {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

// RepoName returns the base name of the repository root directory.
func RepoName(r *Runner) (string, error) {
	top, err := r.Run("rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return filepath.Base(top), nil
}

// CopyEnvFiles copies .env and .env.local from src to dst if they exist.
// Returns nil if neither file exists.
func CopyEnvFiles(src, dst string) error {
	for _, name := range []string{".env", ".env.local"} {
		srcPath := filepath.Join(src, name)
		dstPath := filepath.Join(dst, name)

		data, err := os.ReadFile(srcPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("read %s: %w", srcPath, err)
		}

		info, err := os.Stat(srcPath)
		if err != nil {
			return fmt.Errorf("stat %s: %w", srcPath, err)
		}

		if err := os.WriteFile(dstPath, data, info.Mode()); err != nil {
			return fmt.Errorf("write %s: %w", dstPath, err)
		}
	}
	return nil
}

func getProjectLanguage(path string) string {
	langs := []string{}

	if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
		langs = append(langs, "go")
	}
	if _, err := os.Stat(filepath.Join(path, "package.json")); err == nil {
		if _, err := os.Stat(filepath.Join(path, "tsconfig.json")); err == nil {
			langs = append(langs, "typescript")
		} else {
			langs = append(langs, "node")
		}
	}
	if _, err := os.Stat(filepath.Join(path, "Cargo.toml")); err == nil {
		langs = append(langs, "rust")
	}
	if _, err := os.Stat(filepath.Join(path, "pyproject.toml")); err == nil {
		langs = append(langs, "python")
	} else if _, err := os.Stat(filepath.Join(path, "requirements.txt")); err == nil {
		langs = append(langs, "python")
	}
	if _, err := os.Stat(filepath.Join(path, "composer.json")); err == nil {
		langs = append(langs, "php")
	}
	if _, err := os.Stat(filepath.Join(path, "Gemfile")); err == nil {
		langs = append(langs, "ruby")
	}
	_, errPom := os.Stat(filepath.Join(path, "pom.xml"))
	_, errGradle := os.Stat(filepath.Join(path, "build.gradle"))
	if errPom == nil || errGradle == nil {
		langs = append(langs, "java")
	}

	if len(langs) > 0 {
		return strings.Join(langs, ", ")
	}
	return ""
}

func timeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%d mins ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%d hours ago", int(d.Hours()))
	case d < 48*time.Hour:
		return "yesterday"
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%d days ago", int(d.Hours()/24))
	case d < 365*24*time.Hour:
		return fmt.Sprintf("%d months ago", int(d.Hours()/(24*30)))
	default:
		return fmt.Sprintf("%d years ago", int(d.Hours()/(24*365)))
	}
}

// ShortStatus returns a compact status string for the given repo path.
func ShortStatus(path string) string {
	lang := getProjectLanguage(path)

	r := &Runner{Dir: path}

	lastCommit := ""
	if logOut, err := r.Run("log", "-1", "--format=%cr"); err == nil && logOut != "" {
		lastCommit = strings.TrimSpace(logOut)
	}

	lastTouched := ""
	if info, err := os.Stat(path); err == nil {
		lastTouched = timeAgo(info.ModTime())
	}

	out, err := r.Run("status", "--short", "--branch")
	if err != nil {
		var parts []string
		if lang != "" {
			parts = append(parts, lang)
		}
		parts = append(parts, "non-git")
		if lastTouched != "" {
			parts = append(parts, lastTouched)
		}
		return strings.Join(parts, " · ")
	}

	lines := strings.Split(out, "\n")
	if len(lines) == 0 || lines[0] == "" {
		var parts []string
		if lang != "" {
			parts = append(parts, lang)
		}
		parts = append(parts, "clean")
		if lastCommit != "" {
			parts = append(parts, lastCommit)
		}
		return strings.Join(parts, " · ")
	}

	branchLine := lines[0]
	branch := "HEAD"
	if strings.HasPrefix(branchLine, "## ") {
		b := strings.TrimPrefix(branchLine, "## ")
		if idx := strings.Index(b, "..."); idx != -1 {
			b = b[:idx]
		}
		if strings.HasPrefix(b, "No commits yet on ") {
			b = strings.TrimPrefix(b, "No commits yet on ")
		}
		branch = strings.TrimSpace(b)
	} else {
		branch = "no branch"
	}

	changes := 0
	for _, l := range lines[1:] {
		if strings.TrimSpace(l) != "" {
			changes++
		}
	}

	status := fmt.Sprintf("%s · clean", branch)
	if changes > 0 {
		status = fmt.Sprintf("%s · %d uncommitted", branch, changes)
	}

	var parts []string
	if lang != "" {
		parts = append(parts, lang)
	}
	parts = append(parts, status)
	if lastCommit != "" {
		parts = append(parts, lastCommit)
	}

	return strings.Join(parts, " · ")
}

func GitOnlyStatus(path string) string {
	r := &Runner{Dir: path}

	lastCommit := ""
	if logOut, err := r.Run("log", "-1", "--format=%cr"); err == nil && logOut != "" {
		lastCommit = strings.TrimSpace(logOut)
	}

	out, err := r.Run("status", "--short", "--branch")
	if err != nil {
		return ""
	}

	lines := strings.Split(out, "\n")
	if len(lines) == 0 || lines[0] == "" {
		var parts []string
		parts = append(parts, "clean")
		if lastCommit != "" {
			parts = append(parts, lastCommit)
		}
		return strings.Join(parts, " · ")
	}

	changes := 0
	for _, l := range lines[1:] {
		if strings.TrimSpace(l) != "" {
			changes++
		}
	}

	var parts []string
	if changes == 0 {
		parts = append(parts, "clean")
	} else {
		parts = append(parts, fmt.Sprintf("%d uncommitted", changes))
	}
	if lastCommit != "" {
		parts = append(parts, lastCommit)
	}

	return strings.Join(parts, " · ")
}
