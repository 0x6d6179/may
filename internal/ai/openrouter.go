package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

const openRouterBaseURL = "https://openrouter.ai/api/v1"

type OpenRouterClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewOpenRouterClient(baseURL, apiKey string) *OpenRouterClient {
	if baseURL == "" {
		baseURL = openRouterBaseURL
	}
	return &OpenRouterClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type orModelsResponse struct {
	Data []orModel `json:"data"`
}

type orModel struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	ContextLength int       `json:"context_length"`
	Architecture  orArch    `json:"architecture"`
	Pricing       orPricing `json:"pricing"`
}

type orArch struct {
	OutputModalities []string `json:"output_modalities"`
}

type orPricing struct {
	Prompt     string `json:"prompt"`
	Completion string `json:"completion"`
}

func (c *OpenRouterClient) ListModels(ctx context.Context) ([]ModelInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/models", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("list models: unexpected status %d", resp.StatusCode)
	}

	var modelsResp orModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil, fmt.Errorf("decode models response: %w", err)
	}

	var models []ModelInfo
	for _, m := range modelsResp.Data {
		if !isTextOutputModel(m) {
			continue
		}
		models = append(models, ModelInfo{
			ID:              m.ID,
			Name:            m.Name,
			ContextLength:   m.ContextLength,
			PromptPrice:     m.Pricing.Prompt,
			CompletionPrice: m.Pricing.Completion,
		})
	}

	sort.Slice(models, func(i, j int) bool {
		return models[i].Name < models[j].Name
	})

	return models, nil
}

func isTextOutputModel(m orModel) bool {
	for _, mod := range m.Architecture.OutputModalities {
		if mod == "text" {
			return true
		}
	}
	return len(m.Architecture.OutputModalities) == 0
}

type orChatRequest struct {
	Model          string      `json:"model"`
	Messages       []orMessage `json:"messages"`
	ResponseFormat any         `json:"response_format,omitempty"`
}

type orMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type orResponseFormat struct {
	Type       string        `json:"type"`
	JSONSchema *orJSONSchema `json:"json_schema,omitempty"`
}

type orJSONSchema struct {
	Name   string          `json:"name"`
	Strict bool            `json:"strict"`
	Schema json.RawMessage `json:"schema"`
}

type orChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (c *OpenRouterClient) ChatComplete(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	body := orChatRequest{
		Model: req.Model,
	}

	for _, m := range req.Messages {
		body.Messages = append(body.Messages, orMessage(m))
	}

	if req.ResponseFormat != nil {
		rf := &orResponseFormat{Type: req.ResponseFormat.Type}
		if req.ResponseFormat.JSONSchema != nil {
			rf.JSONSchema = &orJSONSchema{
				Name:   req.ResponseFormat.JSONSchema.Name,
				Strict: req.ResponseFormat.JSONSchema.Strict,
				Schema: req.ResponseFormat.JSONSchema.Schema,
			}
		}
		body.ResponseFormat = rf
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	const maxRetries = 3
	var lastErr error

	for attempt := range maxRetries {
		if attempt > 0 {
			delay := time.Duration(attempt) * 500 * time.Millisecond
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}
		c.setHeaders(httpReq)

		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			lastErr = fmt.Errorf("chat complete: %w", err)
			continue
		}

		if isRetryable(resp.StatusCode) {
			resp.Body.Close()
			lastErr = fmt.Errorf("chat complete: unexpected status %d", resp.StatusCode)
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("chat complete: unexpected status %d", resp.StatusCode)
		}

		var chatResp orChatResponse
		if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}

		if len(chatResp.Choices) == 0 {
			return nil, fmt.Errorf("no choices in response")
		}

		return &ChatResponse{Content: chatResp.Choices[0].Message.Content}, nil
	}

	return nil, lastErr
}

func isRetryable(statusCode int) bool {
	switch statusCode {
	case 429, 500, 502, 503, 504:
		return true
	default:
		return false
	}
}

func (c *OpenRouterClient) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Title", "may")
}
