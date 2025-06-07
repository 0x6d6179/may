package wt

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/git"
	"github.com/spf13/cobra"
)

func NewCmdWtCp(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cp [relpath]",
		Short: "copy a file or directory from main worktree to current",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runner := &git.Runner{}

			mainPath, err := git.MainWorktreePath(runner)
			if err != nil {
				return err
			}

			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			src := filepath.Join(mainPath, args[0])
			dst := filepath.Join(cwd, args[0])

			info, err := os.Stat(src)
			if err != nil {
				return fmt.Errorf("stat %s: %w", src, err)
			}

			if info.IsDir() {
				if err := copyDir(src, dst); err != nil {
					return err
				}
			} else {
				if err := copyFile(src, dst); err != nil {
					return err
				}
			}

			fmt.Fprintf(f.IO.ErrOut, "✓ copied %s → %s\n", src, dst)
			return nil
		},
	}

	return cmd
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(dst), err)
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read %s: %w", src, err)
	}

	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat %s: %w", src, err)
	}

	if err := os.WriteFile(dst, data, info.Mode()); err != nil {
		return fmt.Errorf("write %s: %w", dst, err)
	}
	return nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		return os.WriteFile(target, data, info.Mode())
	})
}
