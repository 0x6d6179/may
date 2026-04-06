package env

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/spf13/cobra"
)

func NewCmdEnv(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "manage .env files",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listEnv(f)
		},
	}

	cmdDiff := &cobra.Command{
		Use:   "diff",
		Short: "compare .env vs .env.example",
		RunE: func(cmd *cobra.Command, args []string) error {
			return diffEnv(f)
		},
	}

	cmdGet := &cobra.Command{
		Use:   "get <key>",
		Short: "get value for a key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return getEnv(f, args[0])
		},
	}

	cmdSet := &cobra.Command{
		Use:   "set <key=value>",
		Short: "set or update a key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return setEnv(f, args[0])
		},
	}

	cmd.AddCommand(cmdDiff, cmdGet, cmdSet)
	return cmd
}

func readEnvFile(path string) (map[string]string, error) {
	vars := make(map[string]string)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return vars, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = unquote(value)
		vars[key] = value
	}

	return vars, nil
}

func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func findExampleFile() string {
	for _, name := range []string{".env.example", ".env.template", ".env.sample"} {
		if _, err := os.Stat(name); err == nil {
			return name
		}
	}
	return ".env.example"
}

func listEnv(f *factory.Factory) error {
	vars, err := readEnvFile(".env")
	if err != nil {
		return err
	}

	if len(vars) == 0 {
		fmt.Fprintf(f.IO.ErrOut, "no .env file or empty\n")
		return nil
	}

	var keys []string
	for k := range vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Fprintf(f.IO.Out, "%s=%s\n", k, vars[k])
	}

	return nil
}

func diffEnv(f *factory.Factory) error {
	envVars, err := readEnvFile(".env")
	if err != nil {
		return err
	}

	exampleFile := findExampleFile()
	exampleVars, err := readEnvFile(exampleFile)
	if err != nil {
		return err
	}

	var missingFromEnv []string
	var missingFromExample []string

	for k := range exampleVars {
		if _, ok := envVars[k]; !ok {
			missingFromEnv = append(missingFromEnv, k)
		}
	}

	for k := range envVars {
		if _, ok := exampleVars[k]; !ok {
			missingFromExample = append(missingFromExample, k)
		}
	}

	sort.Strings(missingFromEnv)
	sort.Strings(missingFromExample)

	if len(missingFromEnv) == 0 && len(missingFromExample) == 0 {
		fmt.Fprintf(f.IO.ErrOut, ".env and %s are in sync\n", exampleFile)
		return nil
	}

	if len(missingFromEnv) > 0 {
		fmt.Fprintf(f.IO.Out, "missing from .env:\n")
		for _, k := range missingFromEnv {
			fmt.Fprintf(f.IO.Out, "  %s\n", k)
		}
	}

	if len(missingFromExample) > 0 {
		if len(missingFromEnv) > 0 {
			fmt.Fprintf(f.IO.Out, "\n")
		}
		fmt.Fprintf(f.IO.Out, "missing from %s:\n", exampleFile)
		for _, k := range missingFromExample {
			fmt.Fprintf(f.IO.Out, "  %s\n", k)
		}
	}

	return nil
}

func getEnv(f *factory.Factory, key string) error {
	vars, err := readEnvFile(".env")
	if err != nil {
		return err
	}

	value, ok := vars[key]
	if !ok {
		return fmt.Errorf("key not found: %s", key)
	}

	fmt.Fprintf(f.IO.Out, "%s\n", value)
	return nil
}

func setEnv(f *factory.Factory, keyValue string) error {
	parts := strings.SplitN(keyValue, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format: use key=value")
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	vars, err := readEnvFile(".env")
	if err != nil {
		return err
	}

	vars[key] = value

	var lines []string
	data, err := os.ReadFile(".env")
	if err == nil {
		lines = strings.Split(string(data), "\n")
	}

	var content strings.Builder
	found := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			content.WriteString(line)
			content.WriteString("\n")
			continue
		}

		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == key {
			content.WriteString(key)
			content.WriteString("=")
			content.WriteString(value)
			content.WriteString("\n")
			found = true
		} else {
			content.WriteString(line)
			content.WriteString("\n")
		}
	}

	if !found {
		content.WriteString(key)
		content.WriteString("=")
		content.WriteString(value)
		content.WriteString("\n")
	}

	output := content.String()
	output = strings.TrimSuffix(output, "\n")
	if output != "" {
		output += "\n"
	}

	if err := os.WriteFile(".env", []byte(output), 0o600); err != nil {
		return fmt.Errorf("write .env: %w", err)
	}

	fmt.Fprintf(f.IO.ErrOut, "set %s=%s\n", key, value)
	return nil
}
