package sshm

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Connection struct {
	Name       string `yaml:"name"`
	Host       string `yaml:"host"`
	User       string `yaml:"user"`
	Port       int    `yaml:"port,omitempty"`
	KeyPath    string `yaml:"key,omitempty"`
	ProxyJump  string `yaml:"proxy_jump,omitempty"`
	RemoteCmd  string `yaml:"remote_cmd,omitempty"`
	ExtraFlags string `yaml:"extra_flags,omitempty"`
}

type store struct {
	Connections []Connection `yaml:"connections"`
}

func storagePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".config/may/ssh.yaml"
	}
	return filepath.Join(home, ".config", "may", "ssh.yaml")
}

func loadStore() (*store, error) {
	path := storagePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &store{}, nil
		}
		return nil, fmt.Errorf("read ssh config: %w", err)
	}

	var s store
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse ssh config: %w", err)
	}
	return &s, nil
}

func saveStore(s *store) error {
	path := storagePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal ssh config: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("write ssh config: %w", err)
	}

	return os.Rename(tmp, path)
}

func findConnection(s *store, name string) (*Connection, int) {
	for i := range s.Connections {
		if s.Connections[i].Name == name {
			return &s.Connections[i], i
		}
	}
	return nil, -1
}
