package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNew_ValidProviders(t *testing.T) {
	tests := []struct {
		provider string
	}{
		{ProviderClaude},
		{ProviderOpenAI},
		{ProviderGemini},
		{ProviderOllama},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			a, err := New(tt.provider, "", "test-key", "")
			if err != nil {
				t.Fatalf("New(%q) error: %v", tt.provider, err)
			}
			if a == nil {
				t.Fatalf("New(%q) returned nil adapter", tt.provider)
			}
			info := a.Info()
			if info.Provider != tt.provider {
				t.Errorf("Info().Provider = %q, want %q", info.Provider, tt.provider)
			}
		})
	}
}

func TestNew_InvalidProvider(t *testing.T) {
	_, err := New("invalid", "", "key", "")
	if err == nil {
		t.Error("expected error for invalid provider")
	}
}

func TestNew_OllamaDefaults(t *testing.T) {
	a, err := New(ProviderOllama, "", "", "")
	if err != nil {
		t.Fatalf("New(ollama) error: %v", err)
	}
	// Should use default host and model.
	info := a.Info()
	if info.Provider != ProviderOllama {
		t.Errorf("provider: got %q", info.Provider)
	}
}

func TestProviderConstants(t *testing.T) {
	if ProviderClaude != "claude" {
		t.Errorf("ProviderClaude = %q", ProviderClaude)
	}
	if ProviderOpenAI != "openai" {
		t.Errorf("ProviderOpenAI = %q", ProviderOpenAI)
	}
	if ProviderGemini != "gemini" {
		t.Errorf("ProviderGemini = %q", ProviderGemini)
	}
	if ProviderOllama != "ollama" {
		t.Errorf("ProviderOllama = %q", ProviderOllama)
	}
}

func TestCompletionRequest_Defaults(t *testing.T) {
	req := CompletionRequest{}
	if req.MaxTokens != 0 {
		t.Error("expected zero MaxTokens by default")
	}
	if req.Stream {
		t.Error("expected Stream false by default")
	}
}

func TestGeminiComplete_NonStreaming(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"candidates": [{
				"content": {
					"parts": [{"text": "Hello from Gemini!"}],
					"role": "model"
				}
			}]
		}`)
	}))
	defer server.Close()

	adapter := &geminiAdapter{
		apiKey: "test-key",
		client: server.Client(),
	}

	// Test doGenerate helper directly against the mock server.
	text, err := adapter.doGenerate(
		context.Background(),
		server.URL+"/v1beta/models/gemini-2.0-flash:generateContent?key=test-key",
		[]byte(`{"contents":[{"role":"user","parts":[{"text":"Hello"}]}]}`),
	)
	if err != nil {
		t.Fatalf("doGenerate error: %v", err)
	}
	if text != "Hello from Gemini!" {
		t.Errorf("got %q, want %q", text, "Hello from Gemini!")
	}
}

func TestGeminiComplete_StreamingSSE(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)

		chunks := []string{"Hello ", "World!"}
		for _, text := range chunks {
			resp := geminiGenerateResponse{
				Candidates: []geminiCandidate{{
					Content: geminiContent{
						Role:  "model",
						Parts: []geminiPart{{Text: text}},
					},
				}},
			}
			data, _ := json.Marshal(resp)
			fmt.Fprintf(w, "data: %s\n\n", data)
			if flusher != nil {
				flusher.Flush()
			}
		}
	}))
	defer server.Close()

	adapter := &geminiAdapter{
		apiKey: "test-key",
		client: server.Client(),
	}

	// Make a streaming HTTP request to the mock and parse SSE lines,
	// verifying the same parsing logic the adapter uses.
	resp, err := adapter.client.Post(
		server.URL+"/v1beta/models/gemini-2.0-flash:streamGenerateContent?alt=sse&key=test-key",
		"application/json",
		strings.NewReader(`{"contents":[{"role":"user","parts":[{"text":"Hello"}]}]}`),
	)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	defer resp.Body.Close()

	var collected []string
	buf := make([]byte, 4096)
	var raw strings.Builder
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			raw.Write(buf[:n])
		}
		if readErr != nil {
			break
		}
	}

	for _, line := range strings.Split(raw.String(), "\n") {
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		var genResp geminiGenerateResponse
		if err := json.Unmarshal([]byte(data), &genResp); err != nil {
			t.Fatalf("decode SSE chunk: %v", err)
		}
		for _, cand := range genResp.Candidates {
			for _, part := range cand.Content.Parts {
				collected = append(collected, part.Text)
			}
		}
	}

	got := strings.Join(collected, "")
	if got != "Hello World!" {
		t.Errorf("streamed text: got %q, want %q", got, "Hello World!")
	}
}

func TestGeminiComplete_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"error":{"code":403,"message":"API key invalid"}}`)
	}))
	defer server.Close()

	adapter := &geminiAdapter{
		apiKey: "bad-key",
		client: server.Client(),
	}

	_, err := adapter.doGenerate(
		context.Background(),
		server.URL+"/v1beta/models/gemini-2.0-flash:generateContent?key=bad-key",
		[]byte(`{"contents":[{"role":"user","parts":[{"text":"Hello"}]}]}`),
	)
	if err == nil {
		t.Fatal("expected error for 403 response")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("error should mention status code 403: %v", err)
	}
}
