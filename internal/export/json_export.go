package export

import (
	"encoding/json"

	"github.com/memvra/memvra/internal/memory"
)

// JSONExporter renders ExportData as structured JSON.
type JSONExporter struct{}

type jsonOutput struct {
	Project    jsonProject              `json:"project"`
	Stack      jsonStack                `json:"stack"`
	Memories   map[string][]jsonMemory `json:"memories"`
}

type jsonProject struct {
	Name       string `json:"name"`
	FileCount  int    `json:"file_count"`
	ChunkCount int    `json:"chunk_count"`
}

type jsonStack struct {
	Language     string   `json:"language,omitempty"`
	Framework    string   `json:"framework,omitempty"`
	Database     string   `json:"database,omitempty"`
	Architecture string   `json:"architecture,omitempty"`
	TestFramework string  `json:"test_framework,omitempty"`
	CI           string   `json:"ci,omitempty"`
	Patterns     []string `json:"patterns,omitempty"`
}

type jsonMemory struct {
	ID         string  `json:"id"`
	Content    string  `json:"content"`
	Importance float64 `json:"importance"`
	Source     string  `json:"source"`
}

func (e *JSONExporter) Export(data ExportData) (string, error) {
	ts := data.Stack
	proj := data.Project

	out := jsonOutput{
		Project: jsonProject{
			Name:       proj.Name,
			FileCount:  proj.FileCount,
			ChunkCount: proj.ChunkCount,
		},
		Stack: jsonStack{
			Language:      ts.Language,
			Framework:     ts.Framework,
			Database:      ts.Database,
			Architecture:  ts.Architecture,
			TestFramework: ts.TestFramework,
			CI:            ts.CI,
			Patterns:      ts.DetectedPatterns,
		},
		Memories: groupMemoriesByType(data.Memories),
	}

	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b) + "\n", nil
}

func groupMemoriesByType(memories []memory.Memory) map[string][]jsonMemory {
	groups := make(map[string][]jsonMemory)
	for _, m := range memories {
		key := string(m.MemoryType)
		groups[key] = append(groups[key], jsonMemory{
			ID:         m.ID,
			Content:    m.Content,
			Importance: m.Importance,
			Source:     m.Source,
		})
	}
	// Return nil map as empty object in JSON.
	if len(groups) == 0 {
		return map[string][]jsonMemory{}
	}
	return groups
}

