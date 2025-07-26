package recent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/spf13/cobra"
)

const maxRecent = 20

type entry struct {
	Path      string    `json:"path"`
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
}

func NewCmdRecent(f *factory.Factory) *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "recent",
		Short: "show recently visited projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, err := loadRecent()
			if err != nil {
				return err
			}

			if len(entries) == 0 {
				fmt.Fprintln(f.IO.ErrOut, "no recent projects")
				return nil
			}

			count := limit
			if count > len(entries) {
				count = len(entries)
			}

			for i := 0; i < count; i++ {
				e := entries[i]
				ago := timeAgo(e.Timestamp)
				fmt.Fprintf(f.IO.Out, "  %s  %s  %s\n", ago, e.Name, e.Path)
			}
			return nil
		},
	}
	cmd.Flags().IntVarP(&limit, "limit", "n", 10, "number of recent projects to show")
	return cmd
}

func RecordVisit(path, name string) error {
	entries, _ := loadRecent()

	filtered := make([]entry, 0, len(entries))
	for _, e := range entries {
		if e.Path != path {
			filtered = append(filtered, e)
		}
	}

	filtered = append([]entry{{
		Path:      path,
		Name:      name,
		Timestamp: time.Now(),
	}}, filtered...)

	if len(filtered) > maxRecent {
		filtered = filtered[:maxRecent]
	}

	return saveRecent(filtered)
}

func recentPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "may", "recent.json")
}

func loadRecent() ([]entry, error) {
	path := recentPath()
	if path == "" {
		return nil, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var entries []entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, nil
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})

	return entries, nil
}

func saveRecent(entries []entry) error {
	path := recentPath()
	if path == "" {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}

func timeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		return fmt.Sprintf("%dm ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		return fmt.Sprintf("%dh ago", h)
	default:
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	}
}
