package adapter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

// openaiAdapter implements LLMAdapter for OpenAI.
type openaiAdapter struct {
	client *openai.Client
}

// NewOpenAI creates an OpenAI adapter. If apiKey is empty, OPENAI_API_KEY is used.
func NewOpenAI(apiKey string) LLMAdapter {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	return &openaiAdapter{
		client: openai.NewClient(apiKey),
	}
}

func (o *openaiAdapter) Info() ModelInfo {
	return ModelInfo{
		Name:               "gpt-4o",
		Provider:           ProviderOpenAI,
		MaxContextWindow:   128000,
		SupportsStreaming:  true,
		EmbeddingDimension: 1536, // text-embedding-3-small
	}
}

func (o *openaiAdapter) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	resp, err := o.client.CreateEmbeddings(ctx, openai.EmbeddingRequestStrings{
		Input: texts,
		Model: openai.SmallEmbedding3,
	})
	if err != nil {
		return nil, fmt.Errorf("openai embed: %w", err)
	}

	result := make([][]float32, len(resp.Data))
	for i, d := range resp.Data {
		result[i] = d.Embedding
	}
	return result, nil
}

func (o *openaiAdapter) Complete(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
	model := req.Model
	if model == "" {
		model = "gpt-4o"
	}

	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}

	messages := []openai.ChatCompletionMessage{}
	if req.SystemPrompt != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: req.SystemPrompt,
		})
	}
	if req.Context != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: fmt.Sprintf("<context>\n%s\n</context>", req.Context),
		})
	}
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: req.UserMessage,
	})

	ch := make(chan StreamChunk, 64)

	if !req.Stream {
		go func() {
			defer close(ch)
			resp, err := o.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
				Model:       model,
				Messages:    messages,
				MaxTokens:   maxTokens,
				Temperature: float32(req.Temperature),
			})
			if err != nil {
				ch <- StreamChunk{Error: fmt.Errorf("openai complete: %w", err)}
				return
			}
			if len(resp.Choices) > 0 {
				ch <- StreamChunk{Text: resp.Choices[0].Message.Content}
			}
		}()
		return ch, nil
	}

	stream, err := o.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: float32(req.Temperature),
		Stream:      true,
	})
	if err != nil {
		close(ch)
		return nil, fmt.Errorf("openai stream: %w", err)
	}

	go func() {
		defer close(ch)
		defer stream.Close()
		for {
			resp, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				ch <- StreamChunk{Error: fmt.Errorf("openai stream recv: %w", err)}
				return
			}
			if len(resp.Choices) > 0 {
				ch <- StreamChunk{Text: resp.Choices[0].Delta.Content}
			}
		}
	}()

	return ch, nil
}
