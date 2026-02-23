package adapter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	anthropic "github.com/liushuangls/go-anthropic/v2"
)

// claudeAdapter implements LLMAdapter for Anthropic Claude.
type claudeAdapter struct {
	client *anthropic.Client
}

// NewClaude creates a Claude adapter. If apiKey is empty, ANTHROPIC_API_KEY is used.
func NewClaude(apiKey string) LLMAdapter {
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	return &claudeAdapter{
		client: anthropic.NewClient(apiKey),
	}
}

func (c *claudeAdapter) Info() ModelInfo {
	return ModelInfo{
		Name:               "claude-sonnet-4-6",
		Provider:           ProviderClaude,
		MaxContextWindow:   200000,
		SupportsStreaming:  true,
		EmbeddingDimension: 0, // Claude does not provide embeddings
	}
}

func (c *claudeAdapter) Embed(_ context.Context, _ []string) ([][]float32, error) {
	return nil, errors.New("claude adapter: embeddings not supported; use openai or ollama for embeddings")
}

func (c *claudeAdapter) Complete(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
	model := req.Model
	if model == "" {
		model = "claude-sonnet-4-6"
	}

	// Build the user message, prepending the injected context if present.
	userContent := req.UserMessage
	if req.Context != "" {
		userContent = fmt.Sprintf("<context>\n%s\n</context>\n\n%s", req.Context, req.UserMessage)
	}

	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}

	messages := []anthropic.Message{
		{
			Role:    anthropic.RoleUser,
			Content: []anthropic.MessageContent{anthropic.NewTextMessageContent(userContent)},
		},
	}

	ch := make(chan StreamChunk, 64)

	if !req.Stream {
		// Non-streaming fallback.
		go func() {
			defer close(ch)
			resp, err := c.client.CreateMessages(ctx, anthropic.MessagesRequest{
				Model:     anthropic.Model(model),
				Messages:  messages,
				MaxTokens: maxTokens,
				System:    req.SystemPrompt,
			})
			if err != nil {
				ch <- StreamChunk{Error: fmt.Errorf("claude complete: %w", err)}
				return
			}
			if len(resp.Content) > 0 {
				ch <- StreamChunk{Text: resp.Content[0].GetText()}
			}
		}()
		return ch, nil
	}

	// Streaming — the library uses a callback-based API.
	go func() {
		defer close(ch)

		streamReq := anthropic.MessagesStreamRequest{
			MessagesRequest: anthropic.MessagesRequest{
				Model:     anthropic.Model(model),
				Messages:  messages,
				MaxTokens: maxTokens,
				System:    req.SystemPrompt,
			},
			OnContentBlockDelta: func(delta anthropic.MessagesEventContentBlockDeltaData) {
				if delta.Delta.Type == anthropic.MessagesContentTypeTextDelta {
					ch <- StreamChunk{Text: delta.Delta.GetText()}
				}
			},
		}

		_, err := c.client.CreateMessagesStream(ctx, streamReq)
		if err != nil && !errors.Is(err, io.EOF) {
			ch <- StreamChunk{Error: fmt.Errorf("claude stream: %w", err)}
		}
	}()

	return ch, nil
}
