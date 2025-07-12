package ai

import (
	"context"
	"encoding/json"
)

type Provider interface {
	ListModels(ctx context.Context) ([]ModelInfo, error)
	ChatComplete(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}

type ModelInfo struct {
	ID              string
	Name            string
	ContextLength   int
	PromptPrice     string
	CompletionPrice string
}

type ChatRequest struct {
	Model          string
	Messages       []Message
	ResponseFormat *ResponseFormat
}

type Message struct {
	Role    string
	Content string
}

type ResponseFormat struct {
	Type       string
	JSONSchema *JSONSchema
}

type JSONSchema struct {
	Name   string
	Strict bool
	Schema json.RawMessage
}

type ChatResponse struct {
	Content string
}
