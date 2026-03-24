package initcmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/0x6d6179/may/internal/cmd/shell"
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

type wizardState struct {
	cfg             *config.Config
	profileName     string
	aiProvider      string
	shellName       string
	shellFeats      []string
	shellCmdAliases []string
	profilePath     string
	disableCmds     []string
	setupShell      bool
	setupCmds       bool
}

func runInit(f *factory.Factory) error {
	if !f.IO.IsErrTerminal() {
		fmt.Fprintln(f.IO.ErrOut, "not a terminal")
		return errors.New("not a terminal")
	}

	state := &wizardState{
		cfg: &config.Config{},
	}

	detected := detectWorkspaceDirs()

	if err := setupWorkspaceRoots(f, state.cfg, detected); err != nil {
		return nil
	}

	profileName, err := setupGitIdentity(f, state.cfg)
	if err != nil {
		return nil
	}
	state.profileName = profileName

	if err := setupMappings(f, state.cfg, profileName); err != nil {
		return nil
	}

	if err := setupAI(f, state); err != nil {
		return nil
	}

	if err := setupShellIntegration(f, state); err != nil {
		return nil
	}

	if err := setupCommands(f, state); err != nil {
		return nil
	}

	if err := reviewAndConfirm(f, state); err != nil {
		return nil
	}

	return applyAndSave(f, state)
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

	return candidates
}

func setupWorkspaceRoots(f *factory.Factory, cfg *config.Config, detected []string) error {
	opts := ui.RunOptions{In: f.IO.In, Out: f.IO.ErrOut}

	for {
		selectOpts := make([]ui.Option[string], 0, len(detected)+1)
		for _, p := range detected {
			label := p
			if _, err := os.Stat(p); err != nil {
				label = p + " (create)"
			}
			selectOpts = append(selectOpts, ui.Option[string]{Label: label, Value: p})
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

		if _, err := os.Stat(rootPath); err != nil {
			create, cerr := ui.RunConfirm(opts, ui.ConfirmSpec{
				Title:   fmt.Sprintf("create %s?", rootPath),
				Default: true,
			})
			if errors.Is(cerr, ui.ErrAborted) {
				return cerr
			}
			if cerr != nil {
				return cerr
			}
			if !create {
				continue
			}
			if merr := os.MkdirAll(rootPath, 0o755); merr != nil {
				fmt.Fprintf(f.IO.ErrOut, "  failed to create directory: %s\n", merr)
				continue
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

var aiProviders = []ui.Option[string]{
	{Label: "openrouter", Value: "openrouter"},
	{Label: "cerebras", Value: "cerebras"},
	{Label: "custom", Value: "custom"},
}

var providerDefaults = map[string]struct{ baseURL, model string }{
	"openrouter": {"https://openrouter.ai/api/v1", "inception/mercury-2"},
	"cerebras":   {"https://api.cerebras.ai/v1", "llama-4-scout-17b-16e-instruct"},
}

func setupAI(f *factory.Factory, state *wizardState) error {
	opts := ui.RunOptions{In: f.IO.In, Out: f.IO.ErrOut}

	doSetup, err := ui.RunConfirm(opts, ui.ConfirmSpec{Title: "set up ai?", Default: true})
	if errors.Is(err, ui.ErrAborted) {
		return err
	}
	if err != nil {
		return err
	}
	if !doSetup {
		return nil
	}

	provider, err := ui.RunSelect(opts, ui.SelectSpec[string]{
		Title:   "ai provider",
		Options: aiProviders,
	})
	if errors.Is(err, ui.ErrAborted) {
		return err
	}
	if err != nil {
		return err
	}
	state.aiProvider = provider

	var baseURL, defaultModel string
	if defs, ok := providerDefaults[provider]; ok {
		baseURL = defs.baseURL
		defaultModel = defs.model
	}

	if provider == "custom" {
		baseURL, err = ui.RunInput(opts, ui.InputSpec{Title: "base url"})
		if errors.Is(err, ui.ErrAborted) {
			return err
		}
		if err != nil {
			return err
		}
	}

	apiKey, err := ui.RunInput(opts, ui.InputSpec{Title: "ai api key", Password: true})
	if errors.Is(err, ui.ErrAborted) {
		return err
	}
	if err != nil {
		return err
	}

	model, err := ui.RunInput(opts, ui.InputSpec{
		Title:   "model (enter to keep default)",
		Default: defaultModel,
	})
	if errors.Is(err, ui.ErrAborted) {
		return err
	}
	if err != nil {
		return err
	}
	if model == "" {
		model = defaultModel
	}

	state.cfg.AI.Provider = provider
	state.cfg.AI.BaseURL = baseURL
	state.cfg.AI.APIKey = apiKey
	state.cfg.AI.Model = model

	return nil
}

func setupShellIntegration(f *factory.Factory, state *wizardState) error {
	opts := ui.RunOptions{In: f.IO.In, Out: f.IO.ErrOut}

	doSetup, err := ui.RunConfirm(opts, ui.ConfirmSpec{Title: "set up shell integration?", Default: true})
	if errors.Is(err, ui.ErrAborted) {
		return err
	}
	if err != nil {
		return err
	}
	if !doSetup {
		return nil
	}

	shellName, err := shell.DetectShell()
	if err != nil {
		shellName, err = ui.RunSelect(opts, ui.SelectSpec[string]{
			Title: "select shell",
			Options: []ui.Option[string]{
				{Label: "bash", Value: "bash"},
				{Label: "zsh", Value: "zsh"},
				{Label: "fish", Value: "fish"},
			},
		})
		if errors.Is(err, ui.ErrAborted) {
			return err
		}
		if err != nil {
			return err
		}
	}

	feats, cmdAliases, err := shell.SelectIntegrations(opts, nil, nil, nil, false)
	if errors.Is(err, ui.ErrAborted) {
		return err
	}
	if err != nil {
		return err
	}

	feats = append([]string{"core"}, feats...)

	state.setupShell = true
	state.shellName = shellName
	state.shellFeats = feats
	state.shellCmdAliases = cmdAliases
	state.profilePath = shell.ProfileFile(shellName)

	return nil
}

var toggleableCommands = []struct{ Name, Short, Group string }{
	{"ws", "workspace management", "nav"},
	{"wt", "git worktree manager", "nav"},
	{"j", "smart directory jump", "nav"},
	{"branch", "list or switch git branches", "nav"},
	{"recent", "show recently visited projects", "nav"},
	{"open", "open repository in browser", "nav"},
	{"ai", "ai assistant", "ai"},
	{"stash", "manage git stashes", "git"},
	{"todo", "find todo comments", "git"},
	{"env", "manage .env files", "git"},
	{"run", "run project scripts", "project"},
	{"port", "show or kill processes on a port", "project"},
	{"db", "connect to database from .env urls", "project"},
	{"path", "inspect and debug path", "system"},
	{"ip", "show local and public ip addresses", "system"},
	{"dotfiles", "manage dotfile symlinks", "system"},
	{"weather", "show weather forecast", "system"},
	{"b64", "base64 encode/decode", "encode"},
	{"uuid", "generate uuids", "encode"},
	{"hash", "hash strings or files", "encode"},
	{"jwt", "decode jwt tokens", "encode"},
	{"secret", "encrypt or decrypt secrets", "encode"},
	{"qr", "generate qr codes", "encode"},
	{"id", "git identity management", "meta"},
	{"sshm", "ssh connection manager", "meta"},
	{"alias", "manage shell command aliases", "meta"},
}

func setupCommands(f *factory.Factory, state *wizardState) error {
	opts := ui.RunOptions{In: f.IO.In, Out: f.IO.ErrOut}

	customize, err := ui.RunConfirm(opts, ui.ConfirmSpec{
		Title:   "customize enabled commands?",
		Default: false,
	})
	if errors.Is(err, ui.ErrAborted) {
		return err
	}
	if err != nil {
		return err
	}
	if !customize {
		return nil
	}

	state.setupCmds = true

	options := make([]ui.Option[string], len(toggleableCommands))
	defaults := make([]string, len(toggleableCommands))
	for i, cmd := range toggleableCommands {
		options[i] = ui.Option[string]{
			Label:       cmd.Name,
			Description: fmt.Sprintf("(%s) %s", cmd.Group, cmd.Short),
			Value:       cmd.Name,
		}
		defaults[i] = cmd.Name
	}

	selected, err := ui.RunMultiSelect(opts, ui.MultiSelectSpec[string]{
		Title:    "enable or disable commands",
		Options:  options,
		Defaults: defaults,
	})
	if errors.Is(err, ui.ErrAborted) {
		return err
	}
	if err != nil {
		return err
	}

	selectedSet := make(map[string]bool, len(selected))
	for _, s := range selected {
		selectedSet[s] = true
	}

	var disabled []string
	for _, cmd := range toggleableCommands {
		if !selectedSet[cmd.Name] {
			disabled = append(disabled, cmd.Name)
		}
	}
	state.disableCmds = disabled

	return nil
}

func reviewAndConfirm(f *factory.Factory, state *wizardState) error {
	opts := ui.RunOptions{In: f.IO.In, Out: f.IO.ErrOut}

	fmt.Fprintln(f.IO.ErrOut, "")
	fmt.Fprintf(f.IO.ErrOut, "  workspace roots: %d configured\n", len(state.cfg.Workspace.Roots))
	fmt.Fprintf(f.IO.ErrOut, "  git identity: %s\n", state.profileName)

	if state.cfg.AI.Provider != "" {
		fmt.Fprintf(f.IO.ErrOut, "  ai: %s / %s\n", state.cfg.AI.Provider, state.cfg.AI.Model)
	} else {
		fmt.Fprintln(f.IO.ErrOut, "  ai: skipped")
	}

	if state.setupShell {
		fmt.Fprintf(f.IO.ErrOut, "  shell: %s — %s\n", state.shellName, strings.Join(state.shellFeats, ", "))
	} else {
		fmt.Fprintln(f.IO.ErrOut, "  shell: skipped")
	}

	if state.setupCmds && len(state.disableCmds) > 0 {
		fmt.Fprintf(f.IO.ErrOut, "  commands: %d disabled\n", len(state.disableCmds))
	} else {
		fmt.Fprintln(f.IO.ErrOut, "  commands: all enabled")
	}

	fmt.Fprintln(f.IO.ErrOut, "")

	confirmed, err := ui.RunConfirm(opts, ui.ConfirmSpec{Title: "save and apply?", Default: true})
	if errors.Is(err, ui.ErrAborted) {
		return err
	}
	if err != nil {
		return err
	}
	if !confirmed {
		return errors.New("aborted")
	}

	return nil
}

func applyAndSave(f *factory.Factory, state *wizardState) error {
	if state.setupCmds {
		state.cfg.DisabledCommands = state.disableCmds
	}

	if err := config.Save(state.cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	home, _ := os.UserHomeDir()
	configFile := filepath.Join(home, ".config", "may", "config.yaml")
	fmt.Fprintf(f.IO.ErrOut, "✓ config saved to %s\n", configFile)

	if state.setupShell && state.profilePath != "" {
		var aliases []shell.AliasEntry
		for _, a := range state.cfg.Aliases {
			aliases = append(aliases, shell.AliasEntry{Name: a.Name, Command: a.Command})
		}
		for _, name := range state.shellCmdAliases {
			aliases = append(aliases, shell.AliasEntry{Name: name, Command: name})
		}
		if len(state.shellCmdAliases) > 0 {
			state.cfg.ShellAliasedCommands = state.shellCmdAliases
		}

		snippet := shell.BuildSnippet(state.shellName, state.shellFeats, "", aliases...)
		featuresLabel := strings.Join(state.shellFeats, ",")

		if err := shell.WriteBlock(state.profilePath, featuresLabel, snippet); err != nil {
			return fmt.Errorf("write shell block: %w", err)
		}

		fmt.Fprintf(f.IO.ErrOut, "✓ shell integration written to %s\n", state.profilePath)
		fmt.Fprintf(f.IO.ErrOut, "  run: source %s\n", state.profilePath)
	}

	return nil
}
