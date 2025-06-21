package initcmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/0x6d6179/may/internal/config"
	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/ui"
	"github.com/spf13/cobra"
)

func NewCmdInit(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "interactive first-run setup wizard",
		Long:  "bootstrap ~/.config/may/config.yaml by detecting workspace directories, collecting git identity info, and setting up an AI key.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(f)
		},
	}
	return cmd
}

func runInit(f *factory.Factory) error {
	if !f.IO.IsTerminal() {
		fmt.Fprintln(f.IO.ErrOut, "not a terminal")
		return errors.New("not a terminal")
	}

	cfg := &config.Config{}
	cfg.AI.BaseURL = "https://api.cerebras.ai/v1"
	cfg.AI.Model = "llama-4-scout-17b-16e-instruct"

	detected := detectWorkspaceDirs()

	if err := setupWorkspaceRoots(f, cfg, detected); err != nil {
		return nil //nolint:nilerr // user cancelled
	}

	profile, err := setupGitIdentity(f, cfg)
	if err != nil {
		return nil //nolint:nilerr // user cancelled
	}

	if err := setupMappings(f, cfg, profile); err != nil {
		return nil //nolint:nilerr // user cancelled
	}

	if err := setupAIKey(f, cfg); err != nil {
		return nil //nolint:nilerr // user cancelled
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	home, _ := os.UserHomeDir()
	configFile := filepath.Join(home, ".config", "may", "config.yaml")
	fmt.Fprintln(f.IO.ErrOut, "")
	fmt.Fprintf(f.IO.ErrOut, "✓ config saved to %s\n", configFile)
	fmt.Fprintln(f.IO.ErrOut, "")
	fmt.Fprintln(f.IO.ErrOut, "to activate may in your shell, add this to your ~/.zprofile:")
	fmt.Fprintln(f.IO.ErrOut, `  eval "$(may shell init zsh)"`)
	fmt.Fprintln(f.IO.ErrOut, "")
	fmt.Fprintln(f.IO.ErrOut, "then restart your shell or run:")
	fmt.Fprintln(f.IO.ErrOut, "  source ~/.zprofile")

	return nil
}

func detectWorkspaceDirs() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	candidates := []string{
		filepath.Join(home, "Workspaces"),
		filepath.Join(home, "Projects"),
		filepath.Join(home, "Code"),
		filepath.Join(home, "Developer"),
		filepath.Join(home, "src"),
	}

	var found []string
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			found = append(found, p)
		}
	}
	return found
}

func setupWorkspaceRoots(f *factory.Factory, cfg *config.Config, detected []string) error {
	opts := ui.RunOptions{In: f.IO.In, Out: f.IO.ErrOut}

	for {
		selectOpts := make([]ui.Option[string], 0, len(detected)+1)
		for _, p := range detected {
			selectOpts = append(selectOpts, ui.Option[string]{Label: p, Value: p})
		}
		selectOpts = append(selectOpts, ui.Option[string]{Label: "Enter custom path", Value: "__custom__"})

		rootPath, err := ui.RunSelect(opts, ui.SelectSpec[string]{
			Title:   "primary workspace root",
			Options: selectOpts,
		})
		if errors.Is(err, ui.ErrAborted) {
			return err
		}
		if err != nil {
			return err
		}

		if rootPath == "__custom__" {
			rootPath, err = ui.RunInput(opts, ui.InputSpec{
				Title: "workspace root path",
				Validate: func(s string) error {
					if s == "" {
						return errors.New("path cannot be empty")
					}
					if _, err := os.Stat(s); err != nil {
						return fmt.Errorf("path does not exist: %s", s)
					}
					return nil
				},
			})
			if errors.Is(err, ui.ErrAborted) {
				return err
			}
			if err != nil {
				return err
			}
		}

		defaultName := filepath.Base(rootPath)
		rootName, err := ui.RunInput(opts, ui.InputSpec{
			Title:       "name for this root",
			Placeholder: defaultName,
		})
		if errors.Is(err, ui.ErrAborted) {
			return err
		}
		if err != nil {
			return err
		}
		if rootName == "" {
			rootName = defaultName
		}

		cfg.Workspace.Roots = append(cfg.Workspace.Roots, config.WorkspaceRoot{
			Name: rootName,
			Path: rootPath,
		})

		if cfg.Workspace.DefaultRoot == "" {
			cfg.Workspace.DefaultRoot = rootName
		}

		addAnother, err := ui.RunConfirm(opts, ui.ConfirmSpec{
			Title: "add another root?",
		})
		if errors.Is(err, ui.ErrAborted) {
			return err
		}
		if err != nil {
			return err
		}
		if !addAnother {
			break
		}
	}
	return nil
}

func setupGitIdentity(f *factory.Factory, cfg *config.Config) (string, error) {
	opts := ui.RunOptions{In: f.IO.In, Out: f.IO.ErrOut}

	result, err := ui.RunForm(opts, ui.FormSpec{
		Title: "git identity",
		Fields: []ui.InputField{
			{Key: "name", Label: "full name"},
			{Key: "email", Label: "email"},
			{Key: "gh_user", Label: "github username (optional)"},
			{Key: "profile", Label: "profile name", Placeholder: "personal"},
		},
	})
	if errors.Is(err, ui.ErrAborted) {
		return "", err
	}
	if err != nil {
		return "", err
	}

	profileName := result["profile"]
	if profileName == "" {
		profileName = "personal"
	}

	cfg.Git.Profiles = append(cfg.Git.Profiles, config.Profile{
		Name:     profileName,
		Username: result["name"],
		Email:    result["email"],
		GHUser:   result["gh_user"],
	})
	cfg.Git.DefaultProfile = profileName

	return profileName, nil
}

func setupMappings(f *factory.Factory, cfg *config.Config, profileName string) error {
	opts := ui.RunOptions{In: f.IO.In, Out: f.IO.ErrOut}

	for _, root := range cfg.Workspace.Roots {
		mapIt, err := ui.RunConfirm(opts, ui.ConfirmSpec{
			Title: fmt.Sprintf("map %q to the %q profile?", root.Path, profileName),
		})
		if errors.Is(err, ui.ErrAborted) {
			return err
		}
		if err != nil {
			return err
		}
		if mapIt {
			cfg.Git.Mappings = append(cfg.Git.Mappings, config.Mapping{
				Path:    root.Path,
				Profile: profileName,
			})
		}
	}
	return nil
}

func setupAIKey(f *factory.Factory, cfg *config.Config) error {
	opts := ui.RunOptions{In: f.IO.In, Out: f.IO.ErrOut}

	apiKey, err := ui.RunInput(opts, ui.InputSpec{
		Title:    "ai api key (optional, leave blank to skip)",
		Password: true,
	})
	if errors.Is(err, ui.ErrAborted) {
		return err
	}
	if err != nil {
		return err
	}
	cfg.AI.APIKey = apiKey
	return nil
}
