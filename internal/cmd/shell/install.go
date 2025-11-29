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

func ProfileFile(shell string) string {
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
	var dev bool
	cmd := &cobra.Command{
		Use:   "install [bash|zsh|fish]",
		Short: "add shell integration to your profile",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			shell, err := resolveShell(args)
			if err != nil {
				return err
			}

			profile := ProfileFile(shell)
			if profile == "" {
				return fmt.Errorf("unsupported shell: %s", shell)
			}

			var lines []string
			if dev {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("get working directory: %w", err)
				}
				switch shell {
				case "fish":
					lines = append(lines, fmt.Sprintf("set -gx PATH %s $PATH", cwd))
				default:
					lines = append(lines, fmt.Sprintf("export PATH=%q:$PATH", cwd))
				}
			}
			lines = append(lines, fmt.Sprintf(`eval "$(may shell init %s)"`, shell))

			block := strings.Join(lines, "\n")

			// Check if already present
			if data, err := os.ReadFile(profile); err == nil {
				if strings.Contains(string(data), fmt.Sprintf(`eval "$(may shell init %s)"`, shell)) {
					fmt.Fprintf(f.IO.ErrOut, "✓ already configured in %s\n", profile)
					return nil
				}
			}

			title := fmt.Sprintf("add to %s?", profile)
			if dev {
				title = fmt.Sprintf("add dev PATH + init to %s?", profile)
			}

			opts := ui.RunOptions{In: f.IO.In, Out: f.IO.ErrOut}
			confirm, err := ui.RunConfirm(opts, ui.ConfirmSpec{
				Title: title,
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

			if _, err2 := fmt.Fprintf(file, "\n%s\n", block); err2 != nil {
				return fmt.Errorf("write %s: %w", profile, err2)
			}

			fmt.Fprintf(f.IO.ErrOut, "✓ added to %s\n", profile)
			fmt.Fprintf(f.IO.ErrOut, "  run: source %s\n", profile)
			return nil
		},
	}
	cmd.Flags().BoolVar(&dev, "dev", false, "prepend cwd to PATH for development")
	return cmd
}
