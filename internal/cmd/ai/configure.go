package ai

import (
	"context"
	"errors"
	"fmt"
	"time"

	mayai "github.com/0x6d6179/may/internal/ai"
	"github.com/0x6d6179/may/internal/config"
	"github.com/0x6d6179/may/internal/factory"
	"github.com/0x6d6179/may/internal/ui"
	"github.com/spf13/cobra"
)

func newCmdConfigure(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "configure",
		Short: "configure ai provider and model",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigure(cmd.Context(), f)
		},
	}
}

func runConfigure(ctx context.Context, f *factory.Factory) error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	provider := cfg.AI.Provider
	if provider == "" {
		provider = "openrouter"
	}
	model := cfg.AI.Model
	if model == "" {
		model = mayai.DefaultModel
	}
	keySet := cfg.AI.APIKey != ""

	fmt.Fprintf(f.IO.ErrOut, "provider: %s\n", provider)
	fmt.Fprintf(f.IO.ErrOut, "model:    %s\n", model)
	if keySet {
		fmt.Fprintln(f.IO.ErrOut, "api key:  set")
	} else {
		fmt.Fprintln(f.IO.ErrOut, "api key:  not set")
	}
	fmt.Fprintln(f.IO.ErrOut, "")

	opts := ui.RunOptions{In: f.IO.In, Out: f.IO.ErrOut}

	apiKey, err := ui.RunInput(opts, ui.InputSpec{
		Title:       "ai api key",
		Placeholder: "sk-or-...",
		Default:     cfg.AI.APIKey,
		Password:    true,
	})
	if errors.Is(err, ui.ErrAborted) {
		return nil
	}
	if err != nil {
		return err
	}

	providerClient := mayai.NewOpenRouterClient("", apiKey)

	fetchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	models, err := ui.RunFlow[[]mayai.ModelInfo](&configureLoadFlow{
		ctx:    fetchCtx,
		client: providerClient,
	}, opts)
	if errors.Is(err, ui.ErrAborted) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("fetch models: %w", err)
	}

	selectOpts := make([]ui.Option[string], len(models))
	for i, m := range models {
		desc := fmt.Sprintf("ctx: %d  prompt: %s  completion: %s", m.ContextLength, m.PromptPrice, m.CompletionPrice)
		selectOpts[i] = ui.Option[string]{
			Label:       m.Name,
			Description: desc,
			Value:       m.ID,
		}
	}

	selectedModel, err := ui.RunSelect(opts, ui.SelectSpec[string]{
		Title:   "choose model",
		Options: selectOpts,
		Height:  15,
	})
	if errors.Is(err, ui.ErrAborted) {
		return nil
	}
	if err != nil {
		return err
	}

	cfg.AI.Provider = "openrouter"
	cfg.AI.APIKey = apiKey
	cfg.AI.Model = selectedModel

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Fprintf(f.IO.ErrOut, "✓ saved: provider=openrouter model=%s\n", selectedModel)
	fmt.Fprintln(f.IO.ErrOut, "→ run: may shell configure to enable the ai alias")
	fmt.Fprintln(f.IO.ErrOut, "→ or set MAY_AI_API_KEY env to override the stored key")
	return nil
}

type configureLoadFlow struct {
	ctx    context.Context
	client *mayai.OpenRouterClient
}

func (f *configureLoadFlow) Start() ui.Step {
	return ui.NewLoading[[]mayai.ModelInfo](ui.LoadingSpec[[]mayai.ModelInfo]{
		Title: "configure",
		Label: "fetching models…",
		Task: func() ([]mayai.ModelInfo, error) {
			return f.client.ListModels(f.ctx)
		},
	})
}

func (f *configureLoadFlow) Next(_ any) (ui.Step, bool, error) {
	return nil, true, nil
}
