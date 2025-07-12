package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/0x6d6179/may/internal/config"
)

type chatResponseBody struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func buildChatResponse(primary, alt1, alt2, alt3 string) []byte {
	payload := map[string]string{
		"commit_message": primary,
		"alt1":           alt1,
		"alt2":           alt2,
		"alt3":           alt3,
	}
	inner, _ := json.Marshal(payload)

	outer := chatResponseBody{}
	outer.Choices = append(outer.Choices, struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}{})
	outer.Choices[0].Message.Content = string(inner)

	data, _ := json.Marshal(outer)
	return data
}

func newTestClient(baseURL, apiKey, model string, httpClient *http.Client) *Client {
	provider := NewOpenRouterClient(baseURL, apiKey)
	provider.httpClient = httpClient
	return &Client{provider: provider, model: model}
}

func newTestClientFromConfig(baseURL, apiKey, model string, httpClient *http.Client) *Client {
	cfg := &config.AIConfig{BaseURL: baseURL, APIKey: apiKey, Model: model}
	c := NewClientFromConfig(cfg)
	if or, ok := c.provider.(*OpenRouterClient); ok {
		or.httpClient = httpClient
	}
	return c
}

func TestGenerateCommitMessages_ParseResponse(t *testing.T) {
	wantPrimary := "feat(auth): add JWT login"
	wantAlt1 := "feat(auth): implement JWT authentication"
	wantAlt2 := "feat: add JWT-based login flow"
	wantAlt3 := "feat(login): introduce JWT support"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(buildChatResponse(wantPrimary, wantAlt1, wantAlt2, wantAlt3))
	}))
	defer srv.Close()

	client := newTestClient(srv.URL, "test-key", "test-model", srv.Client())

	ctx := context.Background()
	msgs, err := client.GenerateCommitMessages(ctx, "diff --git a/main.go")
	if err != nil {
		t.Fatalf("GenerateCommitMessages: unexpected error: %v", err)
	}

	if msgs.Primary != wantPrimary {
		t.Errorf("Primary = %q; want %q", msgs.Primary, wantPrimary)
	}
	if msgs.Alt1 != wantAlt1 {
		t.Errorf("Alt1 = %q; want %q", msgs.Alt1, wantAlt1)
	}
	if msgs.Alt2 != wantAlt2 {
		t.Errorf("Alt2 = %q; want %q", msgs.Alt2, wantAlt2)
	}
	if msgs.Alt3 != wantAlt3 {
		t.Errorf("Alt3 = %q; want %q", msgs.Alt3, wantAlt3)
	}
}

func TestGenerateCommitMessages_HTTPError(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL, "test-key", "test-model", srv.Client())

	ctx := context.Background()
	_, err := client.GenerateCommitMessages(ctx, "some diff")
	if err == nil {
		t.Fatal("GenerateCommitMessages with 500: expected error, got nil")
	}

	if callCount != 2 {
		t.Errorf("server called %d times; want 2", callCount)
	}
}

func TestGenerateCommitMessages_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL, "test-key", "test-model", srv.Client())

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := client.GenerateCommitMessages(ctx, "some diff")
	if err == nil {
		t.Fatal("GenerateCommitMessages with timeout context: expected error, got nil")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "context") && !strings.Contains(errStr, "deadline") && !strings.Contains(errStr, "timeout") {
		t.Errorf("error %q does not mention context/deadline/timeout", errStr)
	}
}

func TestGenerateCommitMessages_EmptyChoices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices":[]}`))
	}))
	defer srv.Close()

	client := newTestClient(srv.URL, "test-key", "test-model", srv.Client())

	ctx := context.Background()
	_, err := client.GenerateCommitMessages(ctx, "diff")
	if err == nil {
		t.Fatal("GenerateCommitMessages with empty choices: expected error, got nil")
	}
}

func TestGenerateCommitMessages_MalformedPayload(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices":[{"message":{"content":"not-json"}}]}`))
	}))
	defer srv.Close()

	client := newTestClient(srv.URL, "test-key", "test-model", srv.Client())

	ctx := context.Background()
	_, err := client.GenerateCommitMessages(ctx, "diff")
	if err == nil {
		t.Fatal("GenerateCommitMessages with malformed payload: expected error, got nil")
	}
}
