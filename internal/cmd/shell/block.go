package shell

import (
	"fmt"
	"os"
	"strings"
)

const (
	blockOpen  = "# >>> may >>>"
	blockClose = "# <<< may <<<"
)

func hasBlock(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), blockOpen)
}

func writeBlock(path, features, content string) error {
	_, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", path, err)
	}

	block := buildBlock(features, content)

	if os.IsNotExist(err) {
		return os.WriteFile(path, []byte(block+"\n"), 0644)
	}

	if hasBlock(path) {
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		updated := replaceBlock(string(data), block)
		return os.WriteFile(path, []byte(updated), 0644)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "\n%s\n", block)
	return err
}

func removeBlock(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	updated := stripBlock(string(data))
	return os.WriteFile(path, []byte(updated), 0644)
}

func buildBlock(features, content string) string {
	var b strings.Builder
	b.WriteString(blockOpen + "\n")
	b.WriteString("# managed by: may shell configure\n")
	if features != "" {
		b.WriteString("# features: " + features + "\n")
	}
	b.WriteString(content)
	if !strings.HasSuffix(content, "\n") {
		b.WriteString("\n")
	}
	b.WriteString(blockClose)
	return b.String()
}

func replaceBlock(profile, newBlock string) string {
	start := strings.Index(profile, blockOpen)
	if start == -1 {
		return profile + "\n" + newBlock + "\n"
	}
	end := strings.Index(profile, blockClose)
	if end == -1 {
		return profile[:start] + newBlock + "\n"
	}
	end += len(blockClose)
	return profile[:start] + newBlock + profile[end:]
}

func stripBlock(profile string) string {
	start := strings.Index(profile, blockOpen)
	if start == -1 {
		return profile
	}
	end := strings.Index(profile, blockClose)
	if end == -1 {
		return profile
	}
	end += len(blockClose)
	if start > 0 && profile[start-1] == '\n' {
		start--
	}
	return profile[:start] + profile[end:]
}
