package shell

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/ui"
	"github.com/spf13/cobra"
)

func profileFile(shell string) string {
	home, _ := os.UserHomeDir()
	switch shell {
	case "zsh":
		return filepath.Join(home, ".zprofile")
	case "bash":
		return filepath.Join(home, ".bashrc")
	case "fish":
		return filepath.Join(home, ".config", "fish", "config.fish")
	default:
		return ""
	}
}

func NewCmdShellInstall(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "install [bash|zsh|fish]",
		Short: "add shell integration to your profile",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			shell, err := resolveShell(args)
			if err != nil {
				return err
			}

			profile := profileFile(shell)
			if profile == "" {
				return fmt.Errorf("unsupported shell: %s", shell)
			}

			line := fmt.Sprintf(`eval "$(may shell init %s)"`, shell)

			// Check if already present
			if data, err := os.ReadFile(profile); err == nil {
				if strings.Contains(string(data), line) {
					fmt.Fprintf(f.IO.ErrOut, "✓ already configured in %s\n", profile)
					return nil
				}
			}

			opts := ui.RunOptions{In: f.IO.In, Out: f.IO.ErrOut}
			confirm, err := ui.RunConfirm(opts, ui.ConfirmSpec{
				Title: fmt.Sprintf("add to %s?", profile),
			})
			if errors.Is(err, ui.ErrAborted) {
				return nil
			}
			if err != nil {
				return err
			}

			if !confirm {
				return nil
			}

			file, err2 := os.OpenFile(profile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err2 != nil {
				return fmt.Errorf("open %s: %w", profile, err2)
			}
			defer file.Close()

			if _, err2 := fmt.Fprintf(file, "\n%s\n", line); err2 != nil {
				return fmt.Errorf("write %s: %w", profile, err2)
			}

			fmt.Fprintf(f.IO.ErrOut, "✓ added to %s\n", profile)
			fmt.Fprintf(f.IO.ErrOut, "  run: source %s\n", profile)
			return nil
		},
	}
}
