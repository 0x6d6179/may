package git

import (
	"os"
	"testing"
)

func TestBranchSanitize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"feature/my-auth", "feature-my-auth"},
		{"feat//double", "feat-double"},
		{"feat!!!", "feat"},
		{"main", "main"},
		{"/leading", "leading"},
		{"trailing/", "trailing"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := BranchSanitize(tt.input)
			if got != tt.want {
				t.Errorf("BranchSanitize(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsGitRepo(t *testing.T) {
	t.Run("non-repo", func(t *testing.T) {
		dir := t.TempDir()
		if IsGitRepo(dir) {
			t.Errorf("IsGitRepo(%q) = true; want false", dir)
		}
	})

	t.Run("repo", func(t *testing.T) {
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("os.Getwd: %v", err)
		}
		if !IsGitRepo(cwd) {
			t.Errorf("IsGitRepo(%q) = false; want true", cwd)
		}
	})
}

func TestParseWorktrees(t *testing.T) {
	input := `worktree /path/to/main
HEAD abc123
branch refs/heads/main

worktree /path/to/linked
HEAD def456
branch refs/heads/feature-foo

worktree /path/to/detached
HEAD ghi789
detached
`
	got := parseWorktrees(input)

	if len(got) != 3 {
		t.Fatalf("parseWorktrees: got %d worktrees; want 3", len(got))
	}

	if got[0].Path != "/path/to/main" {
		t.Errorf("worktrees[0].Path = %q; want %q", got[0].Path, "/path/to/main")
	}
	if got[0].Branch != "main" {
		t.Errorf("worktrees[0].Branch = %q; want %q", got[0].Branch, "main")
	}
	if !got[0].IsMain {
		t.Error("worktrees[0].IsMain = false; want true")
	}

	if got[1].Path != "/path/to/linked" {
		t.Errorf("worktrees[1].Path = %q; want %q", got[1].Path, "/path/to/linked")
	}
	if got[1].Branch != "feature-foo" {
		t.Errorf("worktrees[1].Branch = %q; want %q", got[1].Branch, "feature-foo")
	}
	if got[1].IsMain {
		t.Error("worktrees[1].IsMain = true; want false")
	}

	if got[2].Path != "/path/to/detached" {
		t.Errorf("worktrees[2].Path = %q; want %q", got[2].Path, "/path/to/detached")
	}
	if got[2].Branch != "HEAD" {
		t.Errorf("worktrees[2].Branch = %q; want %q", got[2].Branch, "HEAD")
	}
	if got[2].IsMain {
		t.Error("worktrees[2].IsMain = true; want false")
	}
}
