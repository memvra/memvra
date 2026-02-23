package context

import (
	"context"
	"strings"

	"github.com/memvra/memvra/internal/memory"
	"github.com/memvra/memvra/internal/scanner"
)

// BuildOptions controls how context is assembled.
type BuildOptions struct {
	Question            string
	MaxTokens           int
	TopKChunks          int
	TopKMemories        int
	SimilarityThreshold float64
	ExtraFiles          []string // paths to always include
}

// BuiltContext is the result of a context build operation.
type BuiltContext struct {
	SystemPrompt string
	ContextText  string
	TokensUsed   int
	ChunksUsed   int
	MemoriesUsed int
}

// Builder assembles token-budget-aware prompts from project memory.
type Builder struct {
	store        *memory.Store
	orchestrator interface {
		Retrieve(ctx context.Context, query string, opts memory.RetrieveOptions) (*memory.RetrievalResult, error)
	}
	formatter *Formatter
	tokenizer *Tokenizer
}

// NewBuilder creates a Builder.
func NewBuilder(
	store *memory.Store,
	orchestrator interface {
		Retrieve(ctx context.Context, query string, opts memory.RetrieveOptions) (*memory.RetrievalResult, error)
	},
	formatter *Formatter,
	tokenizer *Tokenizer,
) *Builder {
	return &Builder{
		store:        store,
		orchestrator: orchestrator,
		formatter:    formatter,
		tokenizer:    tokenizer,
	}
}

// Build constructs the context for the given question within the token budget.
func (b *Builder) Build(ctx context.Context, opts BuildOptions) (*BuiltContext, error) {
	if opts.MaxTokens == 0 {
		opts.MaxTokens = 8000
	}
	if opts.TopKChunks == 0 {
		opts.TopKChunks = 10
	}
	if opts.TopKMemories == 0 {
		opts.TopKMemories = 5
	}
	if opts.SimilarityThreshold == 0 {
		opts.SimilarityThreshold = 0.3
	}

	remaining := opts.MaxTokens
	var contextSections []string

	// --- Step 1: Project profile (always included) ---
	proj, err := b.store.GetProject()
	if err != nil {
		proj = memory.Project{Name: "unknown"}
	}
	ts, _ := scanner.TechStackFromJSON(proj.TechStack)

	// --- Step 2: Conventions + constraints (always included) ---
	conventions, _ := b.store.ListMemories(memory.TypeConvention)
	constraints, _ := b.store.ListMemories(memory.TypeConstraint)
	decisions, _ := b.store.ListMemories(memory.TypeDecision)

	systemPrompt := b.formatter.FormatSystemPrompt(proj, ts, conventions, constraints)

	// --- Step 3: Retrieve semantically relevant content ---
	retrieval, _ := b.orchestrator.Retrieve(ctx, opts.Question, memory.RetrieveOptions{
		TopKChunks:          opts.TopKChunks,
		TopKMemories:        opts.TopKMemories,
		SimilarityThreshold: opts.SimilarityThreshold,
	})

	// --- Step 4: Decision block ---
	if len(decisions) > 0 {
		block := b.formatter.FormatMemories(memory.TypeDecision, decisions)
		tokens := b.tokenizer.Count(block)
		if tokens <= remaining {
			contextSections = append(contextSections, block)
			remaining -= tokens
		}
	}

	// --- Step 5: Fill remaining budget with retrieved chunks ---
	chunksUsed := 0
	memoriesUsed := 0

	if retrieval != nil {
		// Add relevant memories first.
		for _, m := range retrieval.Memories {
			if m.MemoryType == memory.TypeConvention || m.MemoryType == memory.TypeConstraint {
				continue // Already in system prompt.
			}
			block := "- " + m.Content + "\n"
			tokens := b.tokenizer.Count(block)
			if tokens <= remaining {
				contextSections = append(contextSections, block)
				remaining -= tokens
				memoriesUsed++
			}
		}

		// Add relevant chunks.
		for _, c := range retrieval.Chunks {
			// Resolve file path from the file record.
			filePath := ""
			if file, err := b.store.GetFileByID(c.FileID); err == nil {
				filePath = file.Path
			}
			block := b.formatter.FormatChunk(c, filePath)
			tokens := b.tokenizer.Count(block)
			if tokens <= remaining {
				contextSections = append(contextSections, block)
				remaining -= tokens
				chunksUsed++
			} else if remaining > 100 {
				// Truncate the chunk to fit.
				truncated := b.tokenizer.Truncate(c.Content, remaining-50)
				c.Content = truncated
				block = b.formatter.FormatChunk(c, filePath)
				contextSections = append(contextSections, block)
				remaining = 0
				chunksUsed++
				break
			} else {
				break
			}
		}
	}

	contextText := strings.Join(contextSections, "\n")
	tokensUsed := opts.MaxTokens - remaining

	return &BuiltContext{
		SystemPrompt: systemPrompt,
		ContextText:  contextText,
		TokensUsed:   tokensUsed,
		ChunksUsed:   chunksUsed,
		MemoriesUsed: memoriesUsed,
	}, nil
}
