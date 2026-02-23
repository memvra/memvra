package adapter

import "context"

// Embedder is a narrower interface for components that only need embedding,
// not full chat completion. An LLMAdapter satisfies this interface.
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}
