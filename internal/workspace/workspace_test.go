package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/0x6d6179/may/internal/config"
)

func TestCreate_ExistingDir(t *testing.T) {
	root := t.TempDir()
	existing := filepath.Join(root, "myws")
	if err := os.Mkdir(existing, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{
			Roots: []config.WorkspaceRoot{
				{Name: "default", Path: root},
			},
		},
	}

	got, err := Create(cfg, "myws", "default")
	if err != nil {
		t.Fatalf("Create existing dir: unexpected error: %v", err)
	}
	if got != existing {
		t.Errorf("Create returned %q; want %q", got, existing)
	}
}

func TestCreate_NewDir(t *testing.T) {
	root := t.TempDir()
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{
			Roots: []config.WorkspaceRoot{
				{Name: "default", Path: root},
			},
		},
	}

	want := filepath.Join(root, "newws")
	got, err := Create(cfg, "newws", "default")
	if err != nil {
		t.Fatalf("Create new dir: unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("Create returned %q; want %q", got, want)
	}

	info, err := os.Stat(got)
	if err != nil {
		t.Fatalf("workspace dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("workspace path %q is not a directory", got)
	}
}

func TestCreate_RootNotFound(t *testing.T) {
	cfg := &config.Config{}
	_, err := Create(cfg, "myws", "nonexistent")
	if err == nil {
		t.Error("Create with unknown root: expected error, got nil")
	}
}

func TestList_EmptyRoots(t *testing.T) {
	cfg := &config.Config{}
	workspaces := List(cfg)
	if len(workspaces) != 0 {
		t.Errorf("List on empty config = %v; want empty slice", workspaces)
	}
}

func TestList_NonexistentRoot(t *testing.T) {
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{
			Roots: []config.WorkspaceRoot{
				{Name: "missing", Path: "/nonexistent/path/that/does/not/exist"},
			},
		},
	}
	workspaces := List(cfg)
	if len(workspaces) != 0 {
		t.Errorf("List with nonexistent root = %v; want empty slice", workspaces)
	}
}

func TestList_PopulatedRoot(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"alpha", "beta", "gamma"} {
		if err := os.Mkdir(filepath.Join(root, name), 0o755); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}
	// Add a file to make sure it's excluded
	if err := os.WriteFile(filepath.Join(root, "not-a-dir.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{
			Roots: []config.WorkspaceRoot{
				{Name: "test", Path: root},
			},
		},
	}

	workspaces := List(cfg)
	if len(workspaces) != 3 {
		t.Fatalf("List returned %d workspaces; want 3", len(workspaces))
	}
	for _, ws := range workspaces {
		if ws.Root != "test" {
			t.Errorf("workspace %q has root %q; want %q", ws.Name, ws.Root, "test")
		}
		want := filepath.Join(root, ws.Name)
		if ws.Path != want {
			t.Errorf("workspace %q path = %q; want %q", ws.Name, ws.Path, want)
		}
	}
}

func TestDefaultRoot_Found(t *testing.T) {
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{
			DefaultRoot: "work",
			Roots: []config.WorkspaceRoot{
				{Name: "personal", Path: "/home/user/personal"},
				{Name: "work", Path: "/home/user/work"},
			},
		},
	}

	root := DefaultRoot(cfg)
	if root == nil {
		t.Fatal("DefaultRoot = nil; want non-nil")
	}
	if root.Name != "work" {
		t.Errorf("DefaultRoot.Name = %q; want %q", root.Name, "work")
	}
	if root.Path != "/home/user/work" {
		t.Errorf("DefaultRoot.Path = %q; want %q", root.Path, "/home/user/work")
	}
}

func TestDefaultRoot_NotFound(t *testing.T) {
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{
			DefaultRoot: "missing",
			Roots: []config.WorkspaceRoot{
				{Name: "personal", Path: "/home/user/personal"},
			},
		},
	}

	root := DefaultRoot(cfg)
	if root != nil {
		t.Errorf("DefaultRoot = %+v; want nil", root)
	}
}

func TestDefaultRoot_EmptyConfig(t *testing.T) {
	cfg := &config.Config{}
	root := DefaultRoot(cfg)
	if root != nil {
		t.Errorf("DefaultRoot on empty config = %+v; want nil", root)
	}
}

func TestResolve_Found(t *testing.T) {
	root := t.TempDir()
	wsDir := filepath.Join(root, "myproject")
	if err := os.Mkdir(wsDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{
			Roots: []config.WorkspaceRoot{
				{Name: "default", Path: root},
			},
		},
	}

	got, err := Resolve(cfg, "myproject")
	if err != nil {
		t.Fatalf("Resolve: unexpected error: %v", err)
	}
	if got != wsDir {
		t.Errorf("Resolve = %q; want %q", got, wsDir)
	}
}

func TestResolve_NotFound(t *testing.T) {
	root := t.TempDir()
	cfg := &config.Config{
		Workspace: config.WorkspaceConfig{
			Roots: []config.WorkspaceRoot{
				{Name: "default", Path: root},
			},
		},
	}

	_, err := Resolve(cfg, "nonexistent")
	if err == nil {
		t.Error("Resolve nonexistent workspace: expected error, got nil")
	}
}
