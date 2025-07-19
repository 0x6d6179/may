package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config is the top-level configuration for may.
type Config struct {
	Workspace WorkspaceConfig `yaml:"workspace"`
	Git       GitConfig       `yaml:"git"`
	AI        AIConfig        `yaml:"ai"`
	Aliases   []Alias         `yaml:"aliases,omitempty"`
}

// WorkspaceConfig holds workspace root definitions.
type WorkspaceConfig struct {
	DefaultRoot string          `yaml:"default_root"`
	Roots       []WorkspaceRoot `yaml:"roots"`
}

// WorkspaceRoot is a named filesystem path that may manages.
type WorkspaceRoot struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

// GitConfig holds git identity profiles and path-to-profile mappings.
type GitConfig struct {
	DefaultProfile string    `yaml:"default_profile"`
	Profiles       []Profile `yaml:"profiles"`
	Mappings       []Mapping `yaml:"mappings"`
}

// Profile is a named git identity.
type Profile struct {
	Name     string `yaml:"name"`
	Username string `yaml:"username"`
	Email    string `yaml:"email"`
	GHUser   string `yaml:"gh_user"`
}

// Mapping associates a filesystem path with a git profile.
type Mapping struct {
	Path    string `yaml:"path"`
	Profile string `yaml:"profile"`
}

// AIConfig holds settings for the AI backend.
type AIConfig struct {
	Provider string `yaml:"provider"`
	BaseURL  string `yaml:"base_url,omitempty"`
	APIKey   string `yaml:"api_key"`
	Model    string `yaml:"model"`
}

type Alias struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
}

// Load reads the config from disk, applies env overrides, and expands tilde paths.
// If the config file does not exist, an empty Config is returned without error.
func Load() (*Config, error) {
	path := configPath()

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}

	if err := expandPaths(&cfg); err != nil {
		return nil, fmt.Errorf("expand paths: %w", err)
	}

	if key := os.Getenv("MAY_AI_API_KEY"); key != "" {
		cfg.AI.APIKey = key
	}

	return &cfg, nil
}

// Save writes cfg to the config path atomically.
func Save(cfg *Config) error {
	path := configPath()

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("write config tmp: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename config tmp: %w", err)
	}

	return nil
}

func configPath() string {
	if p := os.Getenv("MAY_CONFIG"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".config/may/config.yaml"
	}
	return filepath.Join(home, ".config", "may", "config.yaml")
}

func expandPaths(cfg *Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	expand := func(p string) string {
		if p == "~" {
			return home
		}
		if strings.HasPrefix(p, "~/") {
			return home + p[1:]
		}
		return p
	}

	for i := range cfg.Workspace.Roots {
		cfg.Workspace.Roots[i].Path = expand(cfg.Workspace.Roots[i].Path)
	}

	for i := range cfg.Git.Mappings {
		cfg.Git.Mappings[i].Path = expand(cfg.Git.Mappings[i].Path)
	}

	return nil
}
