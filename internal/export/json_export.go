package export

import (
	"encoding/json"
	"time"

	"github.com/memvra/memvra/internal/memory"
)

// JSONExporter renders ExportData as structured JSON.
type JSONExporter struct{}

type jsonOutput struct {
	GitState   *jsonGitState            `json:"work_in_progress,omitempty"`
	Sessions   []jsonSession            `json:"recent_activity,omitempty"`
	Project    jsonProject              `json:"project"`
	Stack      jsonStack                `json:"stack"`
	Memories   map[string][]jsonMemory  `json:"memories"`
}

type jsonGitState struct {
	Branch    string   `json:"branch,omitempty"`
	Staged    []string `json:"staged,omitempty"`
	Modified  []string `json:"modified,omitempty"`
	Untracked []string `json:"untracked,omitempty"`
	DiffStat  string   `json:"diff_stat,omitempty"`
}

type jsonSession struct {
	Timestamp string `json:"timestamp"`
	Question  string `json:"question"`
	Summary   string `json:"summary,omitempty"`
	Model     string `json:"model,omitempty"`
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

	if !data.GitState.IsEmpty() && data.GitState.HasChanges() {
		out.GitState = &jsonGitState{
			Branch:    data.GitState.Branch,
			Staged:    data.GitState.Staged,
			Modified:  data.GitState.Modified,
			Untracked: data.GitState.Untracked,
			DiffStat:  data.GitState.DiffStat,
		}
	}

	if len(data.Sessions) > 0 {
		for i := len(data.Sessions) - 1; i >= 0; i-- {
			s := data.Sessions[i]
			out.Sessions = append(out.Sessions, jsonSession{
				Timestamp: s.CreatedAt.Format(time.RFC3339),
				Question:  s.Question,
				Summary:   s.ResponseSummary,
				Model:     s.ModelUsed,
			})
		}
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

