package shell

import (
	"fmt"
	"os"
	"path/filepath"
)

// detectShell returns the current shell from $SHELL, or an error if unknown.
func detectShell() (string, error) {
	shell := os.Getenv("SHELL")
	switch filepath.Base(shell) {
	case "zsh":
		return "zsh", nil
	case "bash":
		return "bash", nil
	case "fish":
		return "fish", nil
	default:
		if shell == "" {
			return "", fmt.Errorf("$SHELL not set; specify shell explicitly")
		}
		return "", fmt.Errorf("unknown shell %q; pass shell name explicitly", filepath.Base(shell))
	}
}

// resolveShell returns the explicit shell if provided, else autodetects.
func resolveShell(args []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}
	return detectShell()
}
