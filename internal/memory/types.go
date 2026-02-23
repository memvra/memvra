// Package memory defines types for Memvra's persistent memory store.
package memory

import "time"

// MemoryType classifies a stored memory.
type MemoryType string

const (
	TypeDecision   MemoryType = "decision"
	TypeConvention MemoryType = "convention"
	TypeConstraint MemoryType = "constraint"
	TypeNote       MemoryType = "note"
	TypeTodo       MemoryType = "todo"
)

// ValidMemoryType returns true if t is a recognised memory type.
func ValidMemoryType(t MemoryType) bool {
	switch t {
	case TypeDecision, TypeConvention, TypeConstraint, TypeNote, TypeTodo:
		return true
	}
	return false
}

// Memory is a single stored memory record.
type Memory struct {
	ID           string     `json:"id"`
	Content      string     `json:"content"`
	MemoryType   MemoryType `json:"memory_type"`
	Importance   float64    `json:"importance"`
	Source       string     `json:"source"` // "user" or "extracted"
	RelatedFiles []string   `json:"related_files,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Project holds the top-level project record stored in SQLite.
type Project struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	RootPath     string    `json:"root_path"`
	TechStack    string    `json:"tech_stack"`    // JSON blob
	Architecture string    `json:"architecture"`  // JSON blob
	Conventions  string    `json:"conventions"`   // JSON blob
	FileCount    int       `json:"file_count"`
	ChunkCount   int       `json:"chunk_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// File represents an indexed source file.
type File struct {
	ID           string    `json:"id"`
	Path         string    `json:"path"`
	Language     string    `json:"language"`
	LastModified time.Time `json:"last_modified"`
	ContentHash  string    `json:"content_hash"`
	IndexedAt    time.Time `json:"indexed_at"`
}

// Chunk is a content slice of a File.
type Chunk struct {
	ID        string    `json:"id"`
	FileID    string    `json:"file_id"`
	Content   string    `json:"content"`
	StartLine int       `json:"start_line"`
	EndLine   int       `json:"end_line"`
	ChunkType string    `json:"chunk_type"` // code, config, test, docs
	CreatedAt time.Time `json:"created_at"`
}

// Session records a single memvra ask interaction.
type Session struct {
	ID              string    `json:"id"`
	Question        string    `json:"question"`
	ContextUsed     string    `json:"context_used"`      // JSON
	ResponseSummary string    `json:"response_summary"`
	ModelUsed       string    `json:"model_used"`
	TokensUsed      int       `json:"tokens_used"`
	CreatedAt       time.Time `json:"created_at"`
}

// Stats summarises what's stored for a project.
type Stats struct {
	ProjectName string
	TechStack   string
	FileCount   int
	ChunkCount  int
	Memories    map[MemoryType]int
	Sessions    int
	LastUpdated time.Time
	DBSizeBytes int64
}
