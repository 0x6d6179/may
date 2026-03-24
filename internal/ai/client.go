package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/0x6d6179/may/internal/config"
)

const DefaultModel = "inception/mercury-2"

const commitSystemPrompt = `Generate a conventional commit message for the following git diff.
Format: <type>(<scope>): <description>
Types: feat, fix, docs, style, refactor, perf, test, build, ci, chore
Rules: imperative mood, max 50 chars, no period at end.
Return JSON with keys: commit_message, alt1, alt2, alt3 (4 different options).`

var commitResponseSchema = json.RawMessage(`{
	"type": "object",
	"properties": {
		"commit_message": {"type": "string"},
		"alt1": {"type": "string"},
		"alt2": {"type": "string"},
		"alt3": {"type": "string"}
	},
	"required": ["commit_message", "alt1", "alt2", "alt3"],
	"additionalProperties": false
}`)

type CommitMessages struct {
	Primary string
	Alt1    string
	Alt2    string
	Alt3    string
}

type Client struct {
	provider Provider
	model    string
}

func NewProviderFromConfig(cfg *config.AIConfig) Provider {
	baseURL := ""
	if cfg.Provider == "" && cfg.BaseURL != "" {
		baseURL = cfg.BaseURL
	}
	return NewOpenRouterClient(baseURL, cfg.APIKey)
}

func NewClientFromConfig(cfg *config.AIConfig) *Client {
	model := cfg.Model
	if model == "" {
		model = DefaultModel
	}
	return &Client{
		provider: NewProviderFromConfig(cfg),
		model:    model,
	}
}

type commitPayload struct {
	CommitMessage string `json:"commit_message"`
	Alt1          string `json:"alt1"`
	Alt2          string `json:"alt2"`
	Alt3          string `json:"alt3"`
}

func (c *Client) GenerateCommitMessages(ctx context.Context, diff string) (*CommitMessages, error) {
	req := ChatRequest{
		Model: c.model,
		Messages: []Message{
			{Role: "system", Content: commitSystemPrompt},
			{Role: "user", Content: diff},
		},
		ResponseFormat: &ResponseFormat{
			Type: "json_schema",
			JSONSchema: &JSONSchema{
				Name:   "commit_messages",
				Strict: true,
				Schema: commitResponseSchema,
			},
		},
	}

	resp, err := c.provider.ChatComplete(ctx, req)
	if err != nil {
		return nil, err
	}

	var payload commitPayload
	if err := json.Unmarshal([]byte(resp.Content), &payload); err != nil {
		return nil, fmt.Errorf("parse commit payload: %w", err)
	}

	return &CommitMessages{
		Primary: payload.CommitMessage,
		Alt1:    payload.Alt1,
		Alt2:    payload.Alt2,
		Alt3:    payload.Alt3,
	}, nil
}
