package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const systemPrompt = `Generate a conventional commit message for the following git diff.
Format: <type>(<scope>): <description>
Types: feat, fix, docs, style, refactor, perf, test, build, ci, chore
Rules: imperative mood, max 50 chars, no period at end.
Return JSON with keys: commit_message, alt1, alt2, alt3 (4 different options).`

// Client is an OpenAI-compatible HTTP client for generating commit messages.
type Client struct {
	BaseURL    string
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

// CommitMessages holds the AI-generated commit message options.
type CommitMessages struct {
	Primary string
	Alt1    string
	Alt2    string
	Alt3    string
}

type chatRequest struct {
	Model          string    `json:"model"`
	Messages       []message `json:"messages"`
	ResponseFormat any       `json:"response_format"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type       string     `json:"type"`
	JSONSchema jsonSchema `json:"json_schema"`
}

type jsonSchema struct {
	Name   string    `json:"name"`
	Strict bool      `json:"strict"`
	Schema schemaObj `json:"schema"`
}

type schemaObj struct {
	Type       string              `json:"type"`
	Properties map[string]typeProp `json:"properties"`
	Required   []string            `json:"required"`
}

type typeProp struct {
	Type string `json:"type"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type commitPayload struct {
	CommitMessage string `json:"commit_message"`
	Alt1          string `json:"alt1"`
	Alt2          string `json:"alt2"`
	Alt3          string `json:"alt3"`
}

// GenerateCommitMessages calls the AI backend to produce 4 conventional commit options.
func (c *Client) GenerateCommitMessages(ctx context.Context, diff string) (*CommitMessages, error) {
	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}

	reqBody := chatRequest{
		Model: c.Model,
		Messages: []message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: diff},
		},
		ResponseFormat: responseFormat{
			Type: "json_schema",
			JSONSchema: jsonSchema{
				Name:   "commit_messages",
				Strict: true,
				Schema: schemaObj{
					Type: "object",
					Properties: map[string]typeProp{
						"commit_message": {Type: "string"},
						"alt1":           {Type: "string"},
						"alt2":           {Type: "string"},
						"alt3":           {Type: "string"},
					},
					Required: []string{"commit_message", "alt1", "alt2", "alt3"},
				},
			},
		},
	}

	msgs, err := c.doRequest(ctx, reqBody)
	if err != nil {
		time.Sleep(1 * time.Second)
		msgs, err = c.doRequest(ctx, reqBody)
		if err != nil {
			return nil, err
		}
	}

	return msgs, nil
}

func (c *Client) doRequest(ctx context.Context, reqBody chatRequest) (*CommitMessages, error) {
	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/chat/completions", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var chatResp chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	var payload commitPayload
	if err := json.Unmarshal([]byte(chatResp.Choices[0].Message.Content), &payload); err != nil {
		return nil, fmt.Errorf("parse commit payload: %w", err)
	}

	return &CommitMessages{
		Primary: payload.CommitMessage,
		Alt1:    payload.Alt1,
		Alt2:    payload.Alt2,
		Alt3:    payload.Alt3,
	}, nil
}
