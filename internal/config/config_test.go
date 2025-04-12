package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_MissingFile(t *testing.T) {
	t.Setenv("MAY_CONFIG", "/nonexistent/path/config.yaml")

	cfg, err := Load()

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.AI.APIKey != "" || cfg.Git.DefaultProfile != "" || len(cfg.Workspace.Roots) != 0 {
		t.Errorf("expected zero-value Config, got %+v", cfg)
	}
}

func TestLoad_ValidYAML(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")

	content := `
workspace:
  default_root: personal
  roots:
    - name: personal
      path: /home/user/personal
    - name: work
      path: /home/user/work
git:
  default_profile: personal
  profiles:
    - name: personal
      username: jdoe
      email: jdoe@personal.dev
      gh_user: jdoe
  mappings:
    - path: /home/user/work
      profile: work
ai:
  base_url: https://api.cerebras.ai/v1
  api_key: sk-secret
  model: llama-3.3-70b
`
	if err := os.WriteFile(cfgFile, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MAY_CONFIG", cfgFile)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Workspace.DefaultRoot != "personal" {
		t.Errorf("DefaultRoot = %q, want %q", cfg.Workspace.DefaultRoot, "personal")
	}
	if len(cfg.Workspace.Roots) != 2 {
		t.Errorf("len(Roots) = %d, want 2", len(cfg.Workspace.Roots))
	}
	if cfg.Workspace.Roots[0].Name != "personal" {
		t.Errorf("Roots[0].Name = %q, want %q", cfg.Workspace.Roots[0].Name, "personal")
	}
	if cfg.Git.DefaultProfile != "personal" {
		t.Errorf("DefaultProfile = %q, want %q", cfg.Git.DefaultProfile, "personal")
	}
	if len(cfg.Git.Profiles) != 1 {
		t.Errorf("len(Profiles) = %d, want 1", len(cfg.Git.Profiles))
	}
	if cfg.Git.Profiles[0].Email != "jdoe@personal.dev" {
		t.Errorf("Profiles[0].Email = %q, want %q", cfg.Git.Profiles[0].Email, "jdoe@personal.dev")
	}
	if cfg.AI.BaseURL != "https://api.cerebras.ai/v1" {
		t.Errorf("AI.BaseURL = %q", cfg.AI.BaseURL)
	}
	if cfg.AI.APIKey != "sk-secret" {
		t.Errorf("AI.APIKey = %q, want %q", cfg.AI.APIKey, "sk-secret")
	}
	if cfg.AI.Model != "llama-3.3-70b" {
		t.Errorf("AI.Model = %q, want %q", cfg.AI.Model, "llama-3.3-70b")
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")

	content := `
ai:
  api_key: original-key
  model: gpt-4
`
	if err := os.WriteFile(cfgFile, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MAY_CONFIG", cfgFile)
	t.Setenv("MAY_AI_API_KEY", "testkey123")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.AI.APIKey != "testkey123" {
		t.Errorf("AI.APIKey = %q, want %q", cfg.AI.APIKey, "testkey123")
	}
	if cfg.AI.Model != "gpt-4" {
		t.Errorf("AI.Model = %q, want %q", cfg.AI.Model, "gpt-4")
	}
}

func TestLoad_TildeExpansion(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")

	content := `
workspace:
  roots:
    - name: code
      path: ~/Workspaces
git:
  mappings:
    - path: ~/work
      profile: work
`
	if err := os.WriteFile(cfgFile, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MAY_CONFIG", cfgFile)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	if len(cfg.Workspace.Roots) == 0 {
		t.Fatal("expected at least one workspace root")
	}
	rootPath := cfg.Workspace.Roots[0].Path
	if strings.HasPrefix(rootPath, "~") {
		t.Errorf("tilde not expanded: %q", rootPath)
	}
	if rootPath != filepath.Join(home, "Workspaces") {
		t.Errorf("Roots[0].Path = %q, want %q", rootPath, filepath.Join(home, "Workspaces"))
	}

	if len(cfg.Git.Mappings) == 0 {
		t.Fatal("expected at least one git mapping")
	}
	mappingPath := cfg.Git.Mappings[0].Path
	if strings.HasPrefix(mappingPath, "~") {
		t.Errorf("tilde not expanded in mapping: %q", mappingPath)
	}
	if mappingPath != filepath.Join(home, "work") {
		t.Errorf("Mappings[0].Path = %q, want %q", mappingPath, filepath.Join(home, "work"))
	}
}

func TestSave_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "subdir", "config.yaml")
	t.Setenv("MAY_CONFIG", cfgFile)

	cfg := &Config{
		AI: AIConfig{
			BaseURL: "https://api.example.com",
			APIKey:  "save-test-key",
			Model:   "test-model",
		},
		Git: GitConfig{
			DefaultProfile: "default",
		},
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	if _, err := os.Stat(cfgFile); err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() after Save() error: %v", err)
	}

	if loaded.AI.BaseURL != cfg.AI.BaseURL {
		t.Errorf("AI.BaseURL = %q, want %q", loaded.AI.BaseURL, cfg.AI.BaseURL)
	}
	if loaded.AI.APIKey != cfg.AI.APIKey {
		t.Errorf("AI.APIKey = %q, want %q", loaded.AI.APIKey, cfg.AI.APIKey)
	}
	if loaded.AI.Model != cfg.AI.Model {
		t.Errorf("AI.Model = %q, want %q", loaded.AI.Model, cfg.AI.Model)
	}
	if loaded.Git.DefaultProfile != cfg.Git.DefaultProfile {
		t.Errorf("Git.DefaultProfile = %q, want %q", loaded.Git.DefaultProfile, cfg.Git.DefaultProfile)
	}
}

func TestLoad_MAY_CONFIG(t *testing.T) {
	dir := t.TempDir()
	customPath := filepath.Join(dir, "custom-config.yaml")

	content := `
ai:
  model: custom-model
  api_key: custom-key
`
	if err := os.WriteFile(customPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MAY_CONFIG", customPath)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.AI.Model != "custom-model" {
		t.Errorf("AI.Model = %q, want %q", cfg.AI.Model, "custom-model")
	}
	if cfg.AI.APIKey != "custom-key" {
		t.Errorf("AI.APIKey = %q, want %q", cfg.AI.APIKey, "custom-key")
	}
}
