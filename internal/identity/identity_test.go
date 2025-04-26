package identity

import (
	"testing"

	"github.com/0x6d6179/may/internal/config"
)

func TestLongestPrefix(t *testing.T) {
	mappings := []config.Mapping{
		{Path: "/home/user/Work", Profile: "work"},
		{Path: "/home/user/Work/oss", Profile: "personal"},
	}

	tests := []struct {
		name string
		path string
		want string // profile name, or "" for no match
	}{
		{
			name: "longer prefix wins",
			path: "/home/user/Work/oss/repo",
			want: "personal",
		},
		{
			name: "shorter prefix matches",
			path: "/home/user/Work/other",
			want: "work",
		},
		{
			name: "no match",
			path: "/home/other",
			want: "",
		},
		{
			name: "exact match",
			path: "/home/user/Work",
			want: "work",
		},
		{
			name: "exact match longer",
			path: "/home/user/Work/oss",
			want: "personal",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := LongestPrefix(mappings, tc.path)
			if tc.want == "" {
				if m != nil {
					t.Errorf("LongestPrefix(..., %q) = %+v; want nil", tc.path, m)
				}
			} else {
				if m == nil {
					t.Fatalf("LongestPrefix(..., %q) = nil; want profile %q", tc.path, tc.want)
				}
				if m.Profile != tc.want {
					t.Errorf("LongestPrefix(..., %q).Profile = %q; want %q", tc.path, m.Profile, tc.want)
				}
			}
		})
	}
}

func TestLongestPrefix_EmptyMappings(t *testing.T) {
	m := LongestPrefix(nil, "/home/user/Work")
	if m != nil {
		t.Errorf("LongestPrefix(nil, ...) = %+v; want nil", m)
	}
}

func TestLongestPrefix_DoesNotMutateInput(t *testing.T) {
	mappings := []config.Mapping{
		{Path: "/a", Profile: "a"},
		{Path: "/a/b", Profile: "ab"},
	}
	original := make([]config.Mapping, len(mappings))
	copy(original, mappings)

	LongestPrefix(mappings, "/a/b/c")

	for i := range mappings {
		if mappings[i] != original[i] {
			t.Errorf("LongestPrefix mutated input at index %d: got %+v; want %+v", i, mappings[i], original[i])
		}
	}
}

func TestResolveProfile_NoMatch(t *testing.T) {
	cfg := &config.Config{}
	_, ok := ResolveProfile(cfg, "/some/path")
	if ok {
		t.Error("ResolveProfile on empty config returned ok=true; want false")
	}
}

func TestResolveProfile_MatchFound(t *testing.T) {
	cfg := &config.Config{
		Git: config.GitConfig{
			Profiles: []config.Profile{
				{Name: "work", Username: "jdoe", Email: "jdoe@work.com"},
				{Name: "personal", Username: "jdoe-oss", Email: "jdoe@personal.dev"},
			},
			Mappings: []config.Mapping{
				{Path: "/home/user/Work", Profile: "work"},
				{Path: "/home/user/Work/oss", Profile: "personal"},
			},
		},
	}

	tests := []struct {
		name        string
		path        string
		wantProfile string
		wantOK      bool
	}{
		{
			name:        "work path",
			path:        "/home/user/Work/project",
			wantProfile: "work",
			wantOK:      true,
		},
		{
			name:        "oss path wins over work",
			path:        "/home/user/Work/oss/repo",
			wantProfile: "personal",
			wantOK:      true,
		},
		{
			name:   "unrelated path",
			path:   "/tmp/random",
			wantOK: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			profile, ok := ResolveProfile(cfg, tc.path)
			if ok != tc.wantOK {
				t.Fatalf("ResolveProfile(..., %q) ok=%v; want %v", tc.path, ok, tc.wantOK)
			}
			if tc.wantOK {
				if profile == nil {
					t.Fatalf("ResolveProfile(..., %q) returned nil profile; want %q", tc.path, tc.wantProfile)
				}
				if profile.Name != tc.wantProfile {
					t.Errorf("profile.Name = %q; want %q", profile.Name, tc.wantProfile)
				}
			}
		})
	}
}

func TestResolveProfile_MappingExistsButProfileMissing(t *testing.T) {
	cfg := &config.Config{
		Git: config.GitConfig{
			Profiles: []config.Profile{},
			Mappings: []config.Mapping{
				{Path: "/home/user/Work", Profile: "work"},
			},
		},
	}
	_, ok := ResolveProfile(cfg, "/home/user/Work/project")
	if ok {
		t.Error("ResolveProfile returned ok=true when profile name has no matching profile entry")
	}
}
