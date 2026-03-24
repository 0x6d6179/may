package run

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/spf13/cobra"
)

func NewCmdRun(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [script]",
		Short: "run project scripts",
		RunE: func(cmd *cobra.Command, args []string) error {
			runner, scripts, err := detectProject()
			if err != nil {
				return err
			}

			if len(args) == 0 {
				return listScripts(f, runner, scripts)
			}

			scriptName := args[0]
			return executeScript(runner, scriptName, scripts)
		},
	}

	return cmd
}

func detectProject() (string, map[string]string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", nil, err
	}

	if data, err := os.ReadFile(filepath.Join(wd, "package.json")); err == nil {
		scripts, err := parsePackageJSON(data)
		if err == nil {
			runner := detectNodeRunner(wd)
			return runner, scripts, nil
		}
	}

	if _, err := os.Stat(filepath.Join(wd, "Makefile")); err == nil {
		scripts, err := parseMakefile(filepath.Join(wd, "Makefile"))
		if err == nil {
			return "make", scripts, nil
		}
	}

	if _, err := os.Stat(filepath.Join(wd, "Cargo.toml")); err == nil {
		return "cargo", map[string]string{
			"build": "cargo build",
			"test":  "cargo test",
			"run":   "cargo run",
		}, nil
	}

	if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
		return "go", map[string]string{
			"run":   "go run",
			"test":  "go test",
			"build": "go build",
		}, nil
	}

	if _, err := os.Stat(filepath.Join(wd, "pyproject.toml")); err == nil {
		return "python", map[string]string{
			"install": "poetry install",
			"test":    "poetry run pytest",
			"build":   "poetry build",
		}, nil
	}

	if _, err := os.Stat(filepath.Join(wd, "Taskfile.yml")); err == nil {
		scripts, err := parseTaskfile(filepath.Join(wd, "Taskfile.yml"))
		if err == nil {
			return "task", scripts, nil
		}
	}

	if _, err := os.Stat(filepath.Join(wd, "justfile")); err == nil {
		scripts, err := parseJustfile(filepath.Join(wd, "justfile"))
		if err == nil {
			return "just", scripts, nil
		}
	}

	if _, err := os.Stat(filepath.Join(wd, "Justfile")); err == nil {
		scripts, err := parseJustfile(filepath.Join(wd, "Justfile"))
		if err == nil {
			return "just", scripts, nil
		}
	}

	if _, err := os.Stat(filepath.Join(wd, "deno.json")); err == nil {
		return "deno", map[string]string{
			"run": "deno run",
		}, nil
	}

	if _, err := os.Stat(filepath.Join(wd, "deno.jsonc")); err == nil {
		return "deno", map[string]string{
			"run": "deno run",
		}, nil
	}

	return "", nil, nil
}

func detectNodeRunner(wd string) string {
	if _, err := os.Stat(filepath.Join(wd, "yarn.lock")); err == nil {
		return "yarn"
	}
	if _, err := os.Stat(filepath.Join(wd, "pnpm-lock.yaml")); err == nil {
		return "pnpm"
	}
	if _, err := os.Stat(filepath.Join(wd, "bun.lockb")); err == nil {
		return "bun"
	}
	return "npm"
}

func parsePackageJSON(data []byte) (map[string]string, error) {
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}
	return pkg.Scripts, nil
}

func parseMakefile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scripts := make(map[string]string)
	scanner := bufio.NewScanner(file)
	re := regexp.MustCompile(`^([a-zA-Z0-9_-]+):`)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, ".") || strings.HasPrefix(line, "_") {
			continue
		}
		if match := re.FindStringSubmatch(line); match != nil {
			target := match[1]
			scripts[target] = "make " + target
		}
	}

	return scripts, scanner.Err()
}

func parseTaskfile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scripts := make(map[string]string)
	scanner := bufio.NewScanner(file)
	re := regexp.MustCompile(`^  ([a-zA-Z0-9_-]+):`)

	for scanner.Scan() {
		line := scanner.Text()
		if match := re.FindStringSubmatch(line); match != nil {
			task := match[1]
			scripts[task] = "task " + task
		}
	}

	return scripts, scanner.Err()
}

func parseJustfile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scripts := make(map[string]string)
	scanner := bufio.NewScanner(file)
	re := regexp.MustCompile(`^([a-zA-Z0-9_-]+):`)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		if match := re.FindStringSubmatch(line); match != nil {
			recipe := match[1]
			scripts[recipe] = "just " + recipe
		}
	}

	return scripts, scanner.Err()
}

func listScripts(f *factory.Factory, runner string, scripts map[string]string) error {
	if runner == "" {
		return nil
	}

	out := f.IO.Out
	if len(scripts) == 0 {
		return nil
	}

	runnerLabel := runner + " scripts:"
	if runner == "make" {
		runnerLabel = "make targets:"
	} else if runner == "just" {
		runnerLabel = "just recipes:"
	}

	fmt.Fprintln(out, runnerLabel)

	for name, desc := range scripts {
		fmt.Fprintf(out, "  %s\t%s\n", name, desc)
	}

	return nil
}

func executeScript(runner, scriptName string, scripts map[string]string) error {
	var cmd *exec.Cmd

	switch runner {
	case "make":
		cmd = exec.Command("make", scriptName)
	case "just":
		cmd = exec.Command("just", scriptName)
	case "task":
		cmd = exec.Command("task", scriptName)
	case "go":
		cmd = exec.Command("go", "run", ".")
	case "cargo":
		cmd = exec.Command("cargo", scriptName)
	default:
		cmd = exec.Command(runner, "run", scriptName)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
