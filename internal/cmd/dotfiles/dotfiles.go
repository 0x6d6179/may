package dotfiles

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/ui"
	"github.com/spf13/cobra"
)

var skipNames = map[string]bool{
	".git": true, ".DS_Store": true, ".Trash": true,
	".cache": true, ".npm": true, ".cargo": true,
	".rustup": true, ".gradle": true, ".m2": true,
	".ivy2": true, ".sbt": true, ".vscode": true,
	".config": true,
}

func NewCmdDotfiles(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dotfiles",
		Short: "manage dotfile symlinks",
	}

	cmd.AddCommand(newCmdMigrate(f))
	cmd.AddCommand(newCmdLink(f))
	cmd.AddCommand(newCmdUnlink(f))
	cmd.AddCommand(newCmdPull(f))
	cmd.AddCommand(newCmdPush(f))
	cmd.AddCommand(newCmdStatus(f))

	return cmd
}

func isGitHubURL(s string) bool {
	return strings.HasPrefix(s, "https://github.com/") ||
		strings.HasPrefix(s, "git@github.com:")
}

func gitRun(repoPath string, args ...string) error {
	cmdArgs := append([]string{"-C", repoPath}, args...)
	cmd := exec.Command("git", cmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, string(out))
	}
	return nil
}

func copyFileOrDir(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		if err := os.MkdirAll(dst, 0o755); err != nil {
			return err
		}
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if err := copyFileOrDir(filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())); err != nil {
				return err
			}
		}
		return nil
	}

	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(src)
		if err != nil {
			return err
		}
		return os.Symlink(target, dst)
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func symlinkDotfiles(f *factory.Factory, repoPath, home string) (int, error) {
	entries, err := os.ReadDir(repoPath)
	if err != nil {
		return 0, fmt.Errorf("read %s: %w", repoPath, err)
	}

	linked := 0
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, ".") {
			continue
		}
		if skipNames[name] {
			continue
		}

		srcPath := filepath.Join(repoPath, name)
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
	return linked, nil
}

func newCmdMigrate(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate [dest-dir]",
		Short: "move dotfiles into a managed repo and symlink them",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := ui.RunOptions{In: f.IO.In, Out: f.IO.ErrOut}

			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}

			defaultDest := filepath.Join(home, "dotfiles")
			if len(args) > 0 {
				defaultDest = args[0]
			}

			entries, err := os.ReadDir(home)
			if err != nil {
				return fmt.Errorf("read home: %w", err)
			}

			var names []string
			for _, e := range entries {
				name := e.Name()
				if !strings.HasPrefix(name, ".") {
					continue
				}
				if skipNames[name] {
					continue
				}
				names = append(names, name)
			}

			if len(names) == 0 {
				fmt.Fprintf(f.IO.ErrOut, "no dotfiles found in %s\n", home)
				return nil
			}

			options := make([]ui.Option[string], len(names))
			for i, n := range names {
				options[i] = ui.Option[string]{Label: n, Value: n}
			}

			selected, err := ui.RunMultiSelect(opts, ui.MultiSelectSpec[string]{
				Title:    "select dotfiles to migrate",
				Options:  options,
				Defaults: names,
			})
			if err != nil {
				if errors.Is(err, ui.ErrAborted) {
					return nil
				}
				return err
			}

			if len(selected) == 0 {
				return nil
			}

			destInput, err := ui.RunInput(opts, ui.InputSpec{
				Title:   "destination directory",
				Default: defaultDest,
			})
			if err != nil {
				if errors.Is(err, ui.ErrAborted) {
					return nil
				}
				return err
			}

			dest := destInput
			if strings.HasPrefix(dest, "~/") {
				dest = filepath.Join(home, dest[2:])
			} else if dest == "~" {
				dest = home
			}
			dest, err = filepath.Abs(dest)
			if err != nil {
				return err
			}

			if info, err := os.Stat(filepath.Join(dest, ".git")); err == nil && info.IsDir() {
				return fmt.Errorf("already a git repo, use link instead")
			}

			if err := os.MkdirAll(dest, 0o755); err != nil {
				return fmt.Errorf("create dest: %w", err)
			}

			initCmd := exec.Command("git", "init", dest)
			initCmd.Stdout = io.Discard
			initCmd.Stderr = io.Discard
			if out, err := initCmd.CombinedOutput(); err != nil {
				return fmt.Errorf("git init: %w\n%s", err, string(out))
			}

			for _, name := range selected {
				srcPath := filepath.Join(home, name)
				dstPath := filepath.Join(dest, name)

				if err := copyFileOrDir(srcPath, dstPath); err != nil {
					fmt.Fprintf(f.IO.ErrOut, "  error copying %s: %v\n", name, err)
					continue
				}

				if err := os.RemoveAll(srcPath); err != nil {
					fmt.Fprintf(f.IO.ErrOut, "  error removing %s: %v\n", name, err)
					continue
				}

				if err := os.Symlink(dstPath, srcPath); err != nil {
					fmt.Fprintf(f.IO.ErrOut, "  error symlinking %s: %v\n", name, err)
					continue
				}

				fmt.Fprintf(f.IO.ErrOut, "  ✓ migrated %s\n", name)
			}

			if err := gitRun(dest, "add", "-A"); err != nil {
				fmt.Fprintf(f.IO.ErrOut, "  warning: git add failed: %v\n", err)
			} else if err := gitRun(dest, "commit", "-m", "initial dotfiles"); err != nil {
				fmt.Fprintf(f.IO.ErrOut, "  warning: git commit failed: %v\n", err)
			}

			doPush, err := ui.RunConfirm(opts, ui.ConfirmSpec{
				Title:   "push to remote?",
				Default: false,
			})
			if err != nil {
				if errors.Is(err, ui.ErrAborted) {
					doPush = false
				} else {
					return err
				}
			}

			remoteURL := ""
			if doPush {
				remoteChoice, err := ui.RunSelect(opts, ui.SelectSpec[string]{
					Title: "existing or new?",
					Options: []ui.Option[string]{
						{Label: "existing", Value: "existing"},
						{Label: "create new", Value: "create new"},
						{Label: "skip", Value: "skip"},
					},
				})
				if err != nil {
					if errors.Is(err, ui.ErrAborted) {
						remoteChoice = "skip"
					} else {
						return err
					}
				}

				switch remoteChoice {
				case "existing":
					url, err := ui.RunInput(opts, ui.InputSpec{
						Title: "remote URL",
					})
					if err != nil {
						if !errors.Is(err, ui.ErrAborted) {
							return err
						}
						break
					}
					if url != "" {
						remoteURL = url
						if err := gitRun(dest, "remote", "add", "origin", url); err != nil {
							fmt.Fprintf(f.IO.ErrOut, "  warning: add remote failed: %v\n", err)
						} else {
							if err := gitRun(dest, "push", "-u", "origin", "main"); err != nil {
								if err2 := gitRun(dest, "push", "-u", "origin", "master"); err2 != nil {
									fmt.Fprintf(f.IO.ErrOut, "  warning: push failed: %v\n", err)
								}
							}
						}
					}

				case "create new":
					defaultName := filepath.Base(dest)
					repoName, err := ui.RunInput(opts, ui.InputSpec{
						Title:   "repository name",
						Default: defaultName,
					})
					if err != nil {
						if !errors.Is(err, ui.ErrAborted) {
							return err
						}
						break
					}
					if repoName != "" {
						ghPath, lookErr := exec.LookPath("gh")
						if lookErr != nil {
							fmt.Fprintf(f.IO.ErrOut, "  gh CLI not found — install it to create repos automatically\n")
							fmt.Fprintf(f.IO.ErrOut, "  run manually: gh repo create %s --private --source=%s --push\n", repoName, dest)
						} else {
							ghCmd := exec.Command(ghPath, "repo", "create", repoName, "--private", "--source="+dest, "--push")
							ghCmd.Stdout = io.Discard
							ghCmd.Stderr = io.Discard
							if out, err := ghCmd.CombinedOutput(); err != nil {
								fmt.Fprintf(f.IO.ErrOut, "  warning: gh repo create failed: %v\n%s\n", err, string(out))
							} else {
								remoteURL = "https://github.com/" + repoName
							}
						}
					}
				}
			}

			state := &State{
				RepoPath:      dest,
				RemoteURL:     remoteURL,
				MigratedFiles: selected,
			}
			if err := saveState(state); err != nil {
				fmt.Fprintf(f.IO.ErrOut, "  warning: save state failed: %v\n", err)
			}

			fmt.Fprintf(f.IO.ErrOut, "✓ dotfiles migrated to %s\n", dest)
			return nil
		},
	}
}

func newCmdLink(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "link <path-or-url>",
		Short: "symlink dotfiles from a local dir or clone from github",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := ui.RunOptions{In: f.IO.In, Out: f.IO.ErrOut}
			arg := args[0]

			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}

			if isGitHubURL(arg) {
				repoName := filepath.Base(strings.TrimSuffix(arg, ".git"))
				defaultDest := filepath.Join(home, repoName)

				destInput, err := ui.RunInput(opts, ui.InputSpec{
					Title:   "clone destination",
					Default: defaultDest,
				})
				if err != nil {
					if errors.Is(err, ui.ErrAborted) {
						return nil
					}
					return err
				}

				dest := destInput
				if strings.HasPrefix(dest, "~/") {
					dest = filepath.Join(home, dest[2:])
				} else if dest == "~" {
					dest = home
				}
				dest, err = filepath.Abs(dest)
				if err != nil {
					return err
				}

				cloneCmd := exec.Command("git", "clone", arg, dest)
				cloneCmd.Stdout = io.Discard
				cloneCmd.Stderr = io.Discard
				if out, err := cloneCmd.CombinedOutput(); err != nil {
					return fmt.Errorf("git clone: %w\n%s", err, string(out))
				}

				linked, err := symlinkDotfiles(f, dest, home)
				if err != nil {
					return err
				}

				remoteURL := ""
				if info, err := os.Stat(filepath.Join(dest, ".git")); err == nil && info.IsDir() {
					remoteURL = arg
				}

				state := &State{
					RepoPath:  dest,
					RemoteURL: remoteURL,
				}
				if err := saveState(state); err != nil {
					fmt.Fprintf(f.IO.ErrOut, "  warning: save state failed: %v\n", err)
				}

				fmt.Fprintf(f.IO.ErrOut, "✓ linked %d dotfiles from %s\n", linked, dest)
				return nil
			}

			src, err := filepath.Abs(arg)
			if err != nil {
				return err
			}

			linked, err := symlinkDotfiles(f, src, home)
			if err != nil {
				return err
			}

			state := &State{
				RepoPath: src,
			}
			if err := saveState(state); err != nil {
				fmt.Fprintf(f.IO.ErrOut, "  warning: save state failed: %v\n", err)
			}

			fmt.Fprintf(f.IO.ErrOut, "✓ linked %d dotfiles\n", linked)
			return nil
		},
	}
}

func newCmdUnlink(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "unlink",
		Short: "restore dotfiles from repo and remove symlinks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := ui.RunOptions{In: f.IO.In, Out: f.IO.ErrOut}

			state, err := loadState()
			if err != nil {
				return err
			}

			if state.RepoPath == "" {
				return fmt.Errorf("no dotfiles repo configured — run: may dotfiles migrate")
			}

			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}

			restored := 0
			for _, name := range state.MigratedFiles {
				dstPath := filepath.Join(home, name)

				info, err := os.Lstat(dstPath)
				if err != nil {
					continue
				}
				if info.Mode()&os.ModeSymlink == 0 {
					continue
				}

				srcPath := filepath.Join(state.RepoPath, name)
				if _, err := os.Stat(srcPath); err != nil {
					fmt.Fprintf(f.IO.ErrOut, "  skip %s (not found in repo)\n", name)
					continue
				}

				if err := os.Remove(dstPath); err != nil {
					fmt.Fprintf(f.IO.ErrOut, "  error removing symlink %s: %v\n", name, err)
					continue
				}

				if err := copyFileOrDir(srcPath, dstPath); err != nil {
					fmt.Fprintf(f.IO.ErrOut, "  error restoring %s: %v\n", name, err)
					continue
				}

				fmt.Fprintf(f.IO.ErrOut, "  ✓ restored %s\n", name)
				restored++
			}

			clearState, err := ui.RunConfirm(opts, ui.ConfirmSpec{
				Title:   "clear dotfiles state?",
				Default: true,
			})
			if err != nil {
				if !errors.Is(err, ui.ErrAborted) {
					return err
				}
				clearState = false
			}

			if clearState {
				if err := saveState(&State{}); err != nil {
					fmt.Fprintf(f.IO.ErrOut, "  warning: clear state failed: %v\n", err)
				}
			}

			fmt.Fprintf(f.IO.ErrOut, "✓ unlinked %d dotfiles\n", restored)
			return nil
		},
	}
}

func newCmdPull(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pull",
		Aliases: []string{"update"},
		Short:   "pull latest changes and resync symlinks",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			state, err := loadState()
			if err != nil {
				return err
			}

			if state.RepoPath == "" {
				return fmt.Errorf("no dotfiles repo configured — run: may dotfiles migrate")
			}

			if state.RemoteURL == "" {
				return fmt.Errorf("no remote configured — only git repositories can be pulled")
			}

			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}

			if err := gitRun(state.RepoPath, "pull"); err != nil {
				return err
			}

			entries, err := os.ReadDir(state.RepoPath)
			if err != nil {
				return fmt.Errorf("read repo: %w", err)
			}

			for _, e := range entries {
				name := e.Name()
				if !strings.HasPrefix(name, ".") {
					continue
				}
				if name == ".git" {
					continue
				}

				srcPath := filepath.Join(state.RepoPath, name)
				dstPath := filepath.Join(home, name)

				if info, err := os.Lstat(dstPath); err == nil {
					if info.Mode()&os.ModeSymlink != 0 {
						target, _ := os.Readlink(dstPath)
						if target == srcPath {
							continue
						}
					}
					continue
				}

				if err := os.Symlink(srcPath, dstPath); err != nil {
					fmt.Fprintf(f.IO.ErrOut, "  error linking %s: %v\n", name, err)
					continue
				}
				fmt.Fprintf(f.IO.ErrOut, "  ✓ linked %s\n", name)
			}

			fmt.Fprintf(f.IO.ErrOut, "✓ pulled and resynced\n")
			return nil
		},
	}
	return cmd
}

func newCmdPush(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "push",
		Short: "stage all changes and push to remote",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			state, err := loadState()
			if err != nil {
				return err
			}

			if state.RepoPath == "" {
				return fmt.Errorf("no dotfiles repo configured — run: may dotfiles migrate")
			}

			if state.RemoteURL == "" {
				return fmt.Errorf("no remote configured")
			}

			if err := gitRun(state.RepoPath, "add", "-A"); err != nil {
				return err
			}

			if err := gitRun(state.RepoPath, "push"); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "✓ pushed dotfiles\n")
			return nil
		},
	}
}

func newCmdStatus(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "show dotfile locations",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}

			state, err := loadState()
			if err != nil {
				return err
			}

			if state.RepoPath != "" {
				fmt.Fprintf(f.IO.Out, "repo:   %s\n", state.RepoPath)
			}
			if state.RemoteURL != "" {
				fmt.Fprintf(f.IO.Out, "remote: %s\n", state.RemoteURL)
			}

			if len(state.MigratedFiles) > 0 {
				fmt.Fprintf(f.IO.Out, "\n")
				for _, name := range state.MigratedFiles {
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
			}

			dotfiles := []string{
				".bashrc", ".bash_profile", ".zshrc", ".zprofile",
				".gitconfig", ".vimrc", ".tmux.conf",
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

			if state.RepoPath == "" {
				fmt.Fprintf(f.IO.Out, "\nno state — run: may dotfiles migrate\n")
			}

			return nil
		},
	}
}
