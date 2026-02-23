// Package context builds token-budget-aware prompts from project memory.
package context

import (
	"fmt"

	tiktoken "github.com/pkoukk/tiktoken-go"
)

// Tokenizer wraps tiktoken for approximate token counting.
type Tokenizer struct {
	enc *tiktoken.Tiktoken
}

// NewTokenizer creates a Tokenizer using the cl100k_base encoding
// (used by GPT-4 and Claude — a good approximation for all providers).
func NewTokenizer() (*Tokenizer, error) {
	enc, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return nil, fmt.Errorf("tokenizer: get encoding: %w", err)
	}
	return &Tokenizer{enc: enc}, nil
}

// Count returns the approximate number of tokens in s.
func (t *Tokenizer) Count(s string) int {
	return len(t.enc.Encode(s, nil, nil))
}

// Truncate truncates s to at most maxTokens tokens, returning the result.
func (t *Tokenizer) Truncate(s string, maxTokens int) string {
	tokens := t.enc.Encode(s, nil, nil)
	if len(tokens) <= maxTokens {
		return s
	}
	// Decode the truncated token slice back to a string.
	return t.enc.Decode(tokens[:maxTokens])
}
