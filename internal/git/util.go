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

// CopyEnvFiles copies .env* files from src to dst recursively.
// Skips .git, node_modules, and other common non-project directories.
func CopyEnvFiles(src, dst string) error {
	skipDirs := map[string]bool{
		".git":         true,
		"node_modules": true,
		"vendor":       true,
		".next":        true,
		"dist":         true,
		"build":        true,
		"target":       true,
		"__pycache__":  true,
		".venv":        true,
	}

	copied := 0
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		name := info.Name()
		if !strings.HasPrefix(name, ".env") {
			return nil
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return nil
		}

		dstPath := filepath.Join(dst, rel)
		if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Dir(dstPath), err)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		// force 0o600 for .env files since they typically contain secrets
		if err := os.WriteFile(dstPath, data, 0o600); err != nil {
			return fmt.Errorf("write %s: %w", dstPath, err)
		}

		copied++
		return nil
	})

	if err != nil {
		return err
	}

	if copied == 0 {
		return fmt.Errorf("no .env files found in %s", src)
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
	var branch string
	if strings.HasPrefix(branchLine, "## ") {
		b := strings.TrimPrefix(branchLine, "## ")
		if idx := strings.Index(b, "..."); idx != -1 {
			b = b[:idx]
		}
		b = strings.TrimPrefix(b, "No commits yet on ")
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
