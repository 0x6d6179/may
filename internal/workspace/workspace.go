package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/0x6d6179/may/internal/config"
)

// Workspace represents a directory managed under a named root.
type Workspace struct {
	Name string
	Path string
	Root string
}

// List returns all workspaces found across all configured roots.
// For each root, every subdirectory in root.Path is a workspace.
func List(cfg *config.Config) []Workspace {
	var workspaces []Workspace
	for _, root := range cfg.Workspace.Roots {
		entries, err := os.ReadDir(root.Path)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			if strings.HasPrefix(e.Name(), ".") {
				continue
			}
			workspaces = append(workspaces, Workspace{
				Name: e.Name(),
				Path: filepath.Join(root.Path, e.Name()),
				Root: root.Name,
			})
		}
	}
	return workspaces
}

// Resolve returns the absolute path of the first workspace matching name
// across all roots (ordered as in config).
func Resolve(cfg *config.Config, name string) (string, error) {
	for _, root := range cfg.Workspace.Roots {
		path := filepath.Join(root.Path, name)
		info, err := os.Stat(path)
		if err == nil && info.IsDir() {
			return path, nil
		}
	}
	return "", fmt.Errorf("workspace %q not found", name)
}

// Create creates a new workspace directory under the named root.
// If the directory already exists, it is adopted silently.
func Create(cfg *config.Config, name, rootName string) (string, error) {
	var root *config.WorkspaceRoot
	for i := range cfg.Workspace.Roots {
		if cfg.Workspace.Roots[i].Name == rootName {
			root = &cfg.Workspace.Roots[i]
			break
		}
	}
	if root == nil {
		return "", fmt.Errorf("root %q not found in config", rootName)
	}
	path := filepath.Join(root.Path, name)
	if err := os.MkdirAll(path, 0o755); err != nil {
		return "", fmt.Errorf("create workspace %q: %w", path, err)
	}
	return path, nil
}

// DefaultRoot returns the root whose Name matches cfg.Workspace.DefaultRoot,
// or nil if not found.
func DefaultRoot(cfg *config.Config) *config.WorkspaceRoot {
	for i := range cfg.Workspace.Roots {
		if cfg.Workspace.Roots[i].Name == cfg.Workspace.DefaultRoot {
			return &cfg.Workspace.Roots[i]
		}
	}
	return nil
}
