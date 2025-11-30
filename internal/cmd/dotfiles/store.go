package dotfiles

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type State struct {
	RepoPath      string   `yaml:"repo_path"`
	RemoteURL     string   `yaml:"remote_url,omitempty"`
	MigratedFiles []string `yaml:"migrated_files,omitempty"`
}

func statePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".config/may/dotfiles.yaml"
	}
	return filepath.Join(home, ".config", "may", "dotfiles.yaml")
}

func loadState() (*State, error) {
	path := statePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &State{}, nil
		}
		return nil, fmt.Errorf("read dotfiles state %s: %w", path, err)
	}

	var s State
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse dotfiles state %s: %w", path, err)
	}
	return &s, nil
}

func saveState(s *State) error {
	path := statePath()

	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal dotfiles state: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("write dotfiles state tmp: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename dotfiles state tmp: %w", err)
	}

	return nil
}
