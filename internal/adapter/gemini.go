package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

// geminiAdapter implements LLMAdapter for Google Gemini via the REST API.
type geminiAdapter struct {
	apiKey string
	client *http.Client
}

// NewGemini creates a Gemini adapter. If apiKey is empty, GEMINI_API_KEY is used.
func NewGemini(apiKey string) LLMAdapter {
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
	}
	return &geminiAdapter{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

func (g *geminiAdapter) Info() ModelInfo {
	return ModelInfo{
		Name:               "gemini-2.0-flash",
		Provider:           ProviderGemini,
		MaxContextWindow:   1000000,
		SupportsStreaming:  true,
		EmbeddingDimension: 768, // text-embedding-004
	}
}

// geminiEmbedRequest is the request body for the Gemini embedding API.
type geminiEmbedRequest struct {
	Model   string            `json:"model"`
	Content geminiEmbedContent `json:"content"`
}

type geminiEmbedContent struct {
	Parts []geminiEmbedPart `json:"parts"`
}

type geminiEmbedPart struct {
	Text string `json:"text"`
}

type geminiEmbedResponse struct {
	Embedding struct {
		Values []float32 `json:"values"`
	} `json:"embedding"`
}

func (g *geminiAdapter) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	const model = "text-embedding-004"
	baseURL := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:embedContent?key=%s",
		model, g.apiKey,
	)

	results := make([][]float32, 0, len(texts))
	for _, text := range texts {
		body, err := json.Marshal(geminiEmbedRequest{
			Model: "models/" + model,
			Content: geminiEmbedContent{
				Parts: []geminiEmbedPart{{Text: text}},
			},
		})
		if err != nil {
			return nil, fmt.Errorf("gemini embed marshal: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("gemini embed request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := g.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("gemini embed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("gemini embed: status %d", resp.StatusCode)
		}

		var result geminiEmbedResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("gemini embed decode: %w", err)
		}
		results = append(results, result.Embedding.Values)
	}

	return results, nil
}

func (g *geminiAdapter) Complete(_ context.Context, _ CompletionRequest) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 1)
	ch <- StreamChunk{Error: errors.New("gemini adapter: streaming completion not yet implemented")}
	close(ch)
	return ch, nil
}
