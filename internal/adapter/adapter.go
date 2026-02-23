// Package adapter provides a unified interface for LLM providers and embedders.
package adapter

import (
	"context"
	"fmt"
)

// Provider name constants.
const (
	ProviderClaude = "claude"
	ProviderOpenAI = "openai"
	ProviderGemini = "gemini"
	ProviderOllama = "ollama"
)

// StreamChunk is a single token or error delivered during streaming.
type StreamChunk struct {
	Text  string
	Error error
}

// CompletionRequest holds the parameters for a completion call.
type CompletionRequest struct {
	SystemPrompt string
	Context      string
	UserMessage  string
	Model        string
	MaxTokens    int
	Temperature  float64
	Stream       bool
}

// ModelInfo describes the capabilities of a model.
type ModelInfo struct {
	Name               string
	Provider           string
	MaxContextWindow   int
	SupportsStreaming   bool
	EmbeddingDimension int // 0 if not an embedding model
}

// LLMAdapter is the common interface all provider adapters implement.
type LLMAdapter interface {
	// Complete sends a prompt and streams the response.
	Complete(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error)

	// Embed generates embeddings for a batch of texts.
	Embed(ctx context.Context, texts []string) ([][]float32, error)

	// Info returns metadata about the adapter/model.
	Info() ModelInfo
}

// New constructs the LLMAdapter for the named provider.
//
//   - provider: "claude", "openai", "gemini", "ollama"
//   - embedModel: embedding model name (used by Ollama; ignored by others)
//   - apiKey: provider API key (empty = read from env in the concrete adapter)
//   - ollamaHost: base URL for the Ollama server (used only when provider == "ollama")
func New(provider, embedModel, apiKey, ollamaHost string) (LLMAdapter, error) {
	switch provider {
	case ProviderClaude:
		return NewClaude(apiKey), nil
	case ProviderOpenAI:
		return NewOpenAI(apiKey), nil
	case ProviderGemini:
		return NewGemini(apiKey), nil
	case ProviderOllama:
		host := ollamaHost
		if host == "" {
			host = "http://localhost:11434"
		}
		model := embedModel
		if model == "" {
			model = "nomic-embed-text"
		}
		return NewOllama(host, model), nil
	default:
		return nil, fmt.Errorf("adapter: unknown provider %q; valid providers: claude, openai, gemini, ollama", provider)
	}
}
