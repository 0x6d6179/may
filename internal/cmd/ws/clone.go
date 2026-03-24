package ws

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/0x6d6179/may/internal/config"
	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/workspace"
	"github.com/spf13/cobra"
)

func NewCmdWsClone(f *factory.Factory) *cobra.Command {
	var rootName string

	cmd := &cobra.Command{
		Use:   "clone <repo>",
		Short: "clone a repo into a workspace and install dependencies",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			repo := args[0]
			name := repoName(repo)

			if rootName == "" {
				dr := workspace.DefaultRoot(cfg)
				if dr == nil {
					return fmt.Errorf("no default root configured; use --in to specify a root")
				}
				rootName = dr.Name
			}

			root := rootPath(cfg, rootName)
			if root == "" {
				return fmt.Errorf("root %q not found in config", rootName)
			}

			dest := filepath.Join(root, name)
			if _, err := os.Stat(dest); err == nil {
				return fmt.Errorf("%s already exists", dest)
			}

			fmt.Fprintf(f.IO.ErrOut, "cloning %s into %s...\n", repo, dest)

			gitCmd := exec.Command("git", "clone", repo, dest)
			gitCmd.Stdin = os.Stdin
			gitCmd.Stdout = os.Stdout
			gitCmd.Stderr = os.Stderr
			if err := gitCmd.Run(); err != nil {
				return fmt.Errorf("git clone failed: %w", err)
			}

			fmt.Fprintf(f.IO.ErrOut, "✓ cloned %s\n", name)

			installDeps(f, dest)

			fmt.Fprintln(f.IO.Out, dest)
			return nil
		},
	}

	cmd.Flags().StringVar(&rootName, "in", "", "root name to clone into")
	_ = cmd.RegisterFlagCompletionFunc("in", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		cfg, err := f.Config()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		names := make([]string, len(cfg.Workspace.Roots))
		for i, r := range cfg.Workspace.Roots {
			names[i] = r.Name
		}
		return names, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func repoName(repo string) string {
	repo = strings.TrimSuffix(repo, ".git")
	repo = strings.TrimSuffix(repo, "/")

	if i := strings.LastIndex(repo, "/"); i >= 0 {
		return repo[i+1:]
	}
	if i := strings.LastIndex(repo, ":"); i >= 0 {
		return repo[i+1:]
	}
	return repo
}

func rootPath(cfg *config.Config, rootName string) string {
	for _, r := range cfg.Workspace.Roots {
		if r.Name == rootName {
			return r.Path
		}
	}
	return ""
}

type depInstaller struct {
	file    string
	label   string
	cmdFunc func(dir string) *exec.Cmd
}

var installers = []depInstaller{
	{"package.json", "node", func(dir string) *exec.Cmd {
		if _, err := os.Stat(filepath.Join(dir, "bun.lockb")); err == nil {
			return exec.Command("bun", "install")
		}
		if _, err := os.Stat(filepath.Join(dir, "pnpm-lock.yaml")); err == nil {
			return exec.Command("pnpm", "install")
		}
		if _, err := os.Stat(filepath.Join(dir, "yarn.lock")); err == nil {
			return exec.Command("yarn", "install")
		}
		return exec.Command("npm", "install")
	}},
	{"go.mod", "go", func(dir string) *exec.Cmd {
		return exec.Command("go", "mod", "download")
	}},
	{"Cargo.toml", "cargo", func(dir string) *exec.Cmd {
		return exec.Command("cargo", "fetch")
	}},
	{"pyproject.toml", "python", func(dir string) *exec.Cmd {
		if _, err := os.Stat(filepath.Join(dir, "poetry.lock")); err == nil {
			return exec.Command("poetry", "install")
		}
		if _, err := os.Stat(filepath.Join(dir, "uv.lock")); err == nil {
			return exec.Command("uv", "sync")
		}
		return exec.Command("pip", "install", "-e", ".")
	}},
	{"requirements.txt", "pip", func(dir string) *exec.Cmd {
		return exec.Command("pip", "install", "-r", "requirements.txt")
	}},
	{"Gemfile", "bundler", func(dir string) *exec.Cmd {
		return exec.Command("bundle", "install")
	}},
	{"composer.json", "composer", func(dir string) *exec.Cmd {
		return exec.Command("composer", "install")
	}},
	{"pubspec.yaml", "flutter/dart", func(dir string) *exec.Cmd {
		if _, err := exec.LookPath("flutter"); err == nil {
			return exec.Command("flutter", "pub", "get")
		}
		return exec.Command("dart", "pub", "get")
	}},
	{"mix.exs", "elixir", func(dir string) *exec.Cmd {
		return exec.Command("mix", "deps.get")
	}},
}

func installDeps(f *factory.Factory, dir string) {
	for _, inst := range installers {
		if _, err := os.Stat(filepath.Join(dir, inst.file)); err != nil {
			continue
		}

		cmd := inst.cmdFunc(dir)
		if _, err := exec.LookPath(cmd.Args[0]); err != nil {
			fmt.Fprintf(f.IO.ErrOut, "  skipping %s install (%s not found)\n", inst.label, cmd.Args[0])
			continue
		}

		fmt.Fprintf(f.IO.ErrOut, "  installing %s dependencies...\n", inst.label)
		cmd.Dir = dir
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(f.IO.ErrOut, "  ⚠ %s install failed: %s\n", inst.label, err)
		} else {
			fmt.Fprintf(f.IO.ErrOut, "  ✓ %s dependencies installed\n", inst.label)
		}
		return
	}
}
