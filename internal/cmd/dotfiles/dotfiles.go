package dotfiles

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/spf13/cobra"
)

func NewCmdDotfiles(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dotfiles",
		Short: "manage dotfile symlinks",
	}

	cmd.AddCommand(newCmdLink(f))
	cmd.AddCommand(newCmdUnlink(f))
	cmd.AddCommand(newCmdStatus(f))

	return cmd
}

func newCmdLink(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "link <dotfiles-dir>",
		Short: "symlink dotfiles from a directory to home",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			src, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}

			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}

			entries, err := os.ReadDir(src)
			if err != nil {
				return fmt.Errorf("read %s: %w", src, err)
			}

			linked := 0
			for _, e := range entries {
				name := e.Name()
				if !strings.HasPrefix(name, ".") {
					continue
				}
				if name == ".git" || name == ".gitignore" {
					continue
				}

				srcPath := filepath.Join(src, name)
				dstPath := filepath.Join(home, name)

				if _, err := os.Lstat(dstPath); err == nil {
					fmt.Fprintf(f.IO.ErrOut, "  skip %s (already exists)\n", name)
					continue
				}

				if err := os.Symlink(srcPath, dstPath); err != nil {
					fmt.Fprintf(f.IO.ErrOut, "  error %s: %v\n", name, err)
					continue
				}

				fmt.Fprintf(f.IO.ErrOut, "  linked %s → %s\n", name, srcPath)
				linked++
			}

			fmt.Fprintf(f.IO.ErrOut, "✓ linked %d dotfiles\n", linked)
			return nil
		},
	}
}

func newCmdUnlink(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "unlink <dotfiles-dir>",
		Short: "remove symlinks created by link",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			src, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}

			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}

			entries, err := os.ReadDir(src)
			if err != nil {
				return fmt.Errorf("read %s: %w", src, err)
			}

			removed := 0
			for _, e := range entries {
				name := e.Name()
				if !strings.HasPrefix(name, ".") {
					continue
				}

				dstPath := filepath.Join(home, name)
				target, err := os.Readlink(dstPath)
				if err != nil {
					continue
				}

				srcPath := filepath.Join(src, name)
				if target != srcPath {
					continue
				}

				if err := os.Remove(dstPath); err != nil {
					fmt.Fprintf(f.IO.ErrOut, "  error %s: %v\n", name, err)
					continue
				}

				fmt.Fprintf(f.IO.ErrOut, "  unlinked %s\n", name)
				removed++
			}

			fmt.Fprintf(f.IO.ErrOut, "✓ unlinked %d dotfiles\n", removed)
			return nil
		},
	}
}

func newCmdStatus(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "show dotfile locations",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}

			dotfiles := []string{
				".bashrc", ".bash_profile", ".zshrc", ".zprofile",
				".gitconfig", ".vimrc", ".tmux.conf",
				".config/fish/config.fish",
			}
			if runtime.GOOS == "darwin" {
				dotfiles = append(dotfiles, ".config/karabiner")
			}

			for _, name := range dotfiles {
				path := filepath.Join(home, name)
				info, err := os.Lstat(path)
				if err != nil {
					fmt.Fprintf(f.IO.Out, "  ✗ %s (missing)\n", name)
					continue
				}

				if info.Mode()&os.ModeSymlink != 0 {
					target, _ := os.Readlink(path)
					fmt.Fprintf(f.IO.Out, "  ↗ %s → %s\n", name, target)
				} else {
					fmt.Fprintf(f.IO.Out, "  ● %s (regular file)\n", name)
				}
			}
			return nil
		},
	}
}
