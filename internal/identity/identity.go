package identity

import (
	"sort"
	"strings"

	"github.com/0x6d6179/may/internal/config"
)

// LongestPrefix finds the mapping with the longest matching path prefix.
// More specific (longer) paths take priority. Returns nil if no mapping matches.
func LongestPrefix(mappings []config.Mapping, path string) *config.Mapping {
	sorted := make([]config.Mapping, len(mappings))
	copy(sorted, mappings)
	sort.Slice(sorted, func(i, j int) bool {
		return len(sorted[i].Path) > len(sorted[j].Path)
	})
	for i := range sorted {
		if strings.HasPrefix(path, sorted[i].Path) {
			return &sorted[i]
		}
	}
	return nil
}

// ResolveProfile finds the best-matching profile for a given path using longest-prefix.
func ResolveProfile(cfg *config.Config, path string) (*config.Profile, bool) {
	m := LongestPrefix(cfg.Git.Mappings, path)
	if m == nil {
		return nil, false
	}
	for i := range cfg.Git.Profiles {
		if cfg.Git.Profiles[i].Name == m.Profile {
			return &cfg.Git.Profiles[i], true
		}
	}
	return nil, false
}
