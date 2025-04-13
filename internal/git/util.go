package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
