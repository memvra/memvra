package adapter

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
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
		SupportsStreaming:   true,
		EmbeddingDimension: 768, // text-embedding-004
	}
}

// ---------- Embedding types ----------

type geminiEmbedRequest struct {
	Model   string             `json:"model"`
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

// ---------- Completion types ----------

// geminiGenerateRequest is the request body for the Gemini generateContent API.
type geminiGenerateRequest struct {
	Contents         []geminiContent         `json:"contents"`
	SystemInstruction *geminiContent          `json:"systemInstruction,omitempty"`
	GenerationConfig *geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerationConfig struct {
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	Temperature     float64 `json:"temperature,omitempty"`
}

// geminiGenerateResponse is the response from the Gemini generateContent API.
type geminiGenerateResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
	Error      *geminiError      `json:"error,omitempty"`
}

type geminiCandidate struct {
	Content geminiContent `json:"content"`
}

type geminiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (g *geminiAdapter) Complete(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
	model := req.Model
	if model == "" {
		model = "gemini-2.0-flash"
	}

	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}

	// Build the system instruction if provided.
	var sysInstruction *geminiContent
	systemText := req.SystemPrompt
	if req.Context != "" {
		systemText += fmt.Sprintf("\n\n<context>\n%s\n</context>", req.Context)
	}
	if systemText != "" {
		sysInstruction = &geminiContent{
			Parts: []geminiPart{{Text: systemText}},
		}
	}

	genReq := geminiGenerateRequest{
		Contents: []geminiContent{
			{
				Role:  "user",
				Parts: []geminiPart{{Text: req.UserMessage}},
			},
		},
		SystemInstruction: sysInstruction,
		GenerationConfig: &geminiGenerationConfig{
			MaxOutputTokens: maxTokens,
			Temperature:     req.Temperature,
		},
	}

	body, err := json.Marshal(genReq)
	if err != nil {
		return nil, fmt.Errorf("gemini complete marshal: %w", err)
	}

	ch := make(chan StreamChunk, 64)

	if !req.Stream {
		// Non-streaming: use generateContent endpoint.
		url := fmt.Sprintf(
			"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
			model, g.apiKey,
		)

		go func() {
			defer close(ch)
			text, err := g.doGenerate(ctx, url, body)
			if err != nil {
				ch <- StreamChunk{Error: err}
				return
			}
			ch <- StreamChunk{Text: text}
		}()
		return ch, nil
	}

	// Streaming: use streamGenerateContent endpoint with SSE.
	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s",
		model, g.apiKey,
	)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("gemini stream request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	go func() {
		defer close(ch)

		resp, err := g.client.Do(httpReq)
		if err != nil {
			ch <- StreamChunk{Error: fmt.Errorf("gemini stream: %w", err)}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			ch <- StreamChunk{Error: fmt.Errorf("gemini stream: status %d: %s", resp.StatusCode, respBody)}
			return
		}

		// Gemini SSE: each event is "data: {json}\n\n".
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")

			var genResp geminiGenerateResponse
			if err := json.Unmarshal([]byte(data), &genResp); err != nil {
				ch <- StreamChunk{Error: fmt.Errorf("gemini stream decode: %w", err)}
				return
			}

			if genResp.Error != nil {
				ch <- StreamChunk{Error: fmt.Errorf("gemini api error %d: %s", genResp.Error.Code, genResp.Error.Message)}
				return
			}

			for _, cand := range genResp.Candidates {
				for _, part := range cand.Content.Parts {
					if part.Text != "" {
						ch <- StreamChunk{Text: part.Text}
					}
				}
			}
		}
		if err := scanner.Err(); err != nil {
			ch <- StreamChunk{Error: fmt.Errorf("gemini stream scan: %w", err)}
		}
	}()

	return ch, nil
}

// doGenerate makes a non-streaming generateContent call and returns the text.
func (g *geminiAdapter) doGenerate(ctx context.Context, url string, body []byte) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("gemini complete request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("gemini complete: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("gemini complete: status %d: %s", resp.StatusCode, respBody)
	}

	var genResp geminiGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return "", fmt.Errorf("gemini complete decode: %w", err)
	}

	if genResp.Error != nil {
		return "", fmt.Errorf("gemini api error %d: %s", genResp.Error.Code, genResp.Error.Message)
	}

	var parts []string
	for _, cand := range genResp.Candidates {
		for _, part := range cand.Content.Parts {
			if part.Text != "" {
				parts = append(parts, part.Text)
			}
		}
	}
	return strings.Join(parts, ""), nil
}
