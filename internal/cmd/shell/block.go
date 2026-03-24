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

func HasBlock(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), blockOpen)
}

func WriteBlock(path, features, content string) error {
	_, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", path, err)
	}

	block := buildBlock(features, content)

	if os.IsNotExist(err) {
		return os.WriteFile(path, []byte(block+"\n"), 0644)
	}

	if HasBlock(path) {
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

func ReadFeatures(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	const prefix = "# features: "
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, prefix) {
			raw := strings.TrimPrefix(line, prefix)
			parts := strings.Split(raw, ",")
			features := make([]string, 0, len(parts))
			for _, p := range parts {
				if p = strings.TrimSpace(p); p != "" {
					features = append(features, p)
				}
			}
			return features, nil
		}
	}
	return nil, nil
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
