package export

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/memvra/memvra/internal/memory"
	"github.com/memvra/memvra/internal/scanner"
)

func sampleExportData() ExportData {
	return ExportData{
		Project: memory.Project{
			Name:       "testapp",
			FileCount:  10,
			ChunkCount: 50,
		},
		Stack: scanner.TechStack{
			Language:         "Go",
			Framework:        "Gin",
			Database:         "PostgreSQL",
			Architecture:     "API + SPA",
			TestFramework:    "testing",
			DetectedPatterns: []string{"background-jobs"},
		},
		Memories: []memory.Memory{
			{ID: "1", Content: "Use PostgreSQL", MemoryType: memory.TypeDecision, Importance: 0.8, Source: "user"},
			{ID: "2", Content: "Use camelCase", MemoryType: memory.TypeConvention, Importance: 0.7, Source: "user"},
			{ID: "3", Content: "Never store secrets in code", MemoryType: memory.TypeConstraint, Importance: 0.9, Source: "user"},
			{ID: "4", Content: "Interesting observation", MemoryType: memory.TypeNote, Importance: 0.5, Source: "extracted"},
			{ID: "5", Content: "Fix auth flow", MemoryType: memory.TypeTodo, Importance: 0.6, Source: "extracted"},
		},
	}
}

func TestGet_ValidFormats(t *testing.T) {
	for _, name := range []string{"claude", "cursor", "markdown", "json"} {
		exp, ok := Get(name)
		if !ok {
			t.Errorf("Get(%q) returned false", name)
		}
		if exp == nil {
			t.Errorf("Get(%q) returned nil exporter", name)
		}
	}
}

func TestGet_InvalidFormat(t *testing.T) {
	_, ok := Get("invalid")
	if ok {
		t.Error("expected Get('invalid') to return false")
	}
}

func TestValidFormats(t *testing.T) {
	formats := ValidFormats()
	if len(formats) < 4 {
		t.Errorf("expected at least 4 formats, got %d", len(formats))
	}
}

func TestClaudeMDExporter(t *testing.T) {
	data := sampleExportData()
	exp, _ := Get("claude")
	result, err := exp.Export(data)
	if err != nil {
		t.Fatalf("Export error: %v", err)
	}

	checks := []string{
		"testapp",
		"Project Profile",
		"Go",
		"Gin",
		"PostgreSQL",
		"Architectural Decisions",
		"Use PostgreSQL",
		"Coding Conventions",
		"Use camelCase",
		"Constraints",
		"Never store secrets in code",
		"Notes",
		"TODOs",
		"Memvra",
	}
	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("claude export missing %q", check)
		}
	}
}

func TestCursorRulesExporter(t *testing.T) {
	data := sampleExportData()
	exp, _ := Get("cursor")
	result, err := exp.Export(data)
	if err != nil {
		t.Fatalf("Export error: %v", err)
	}

	checks := []string{
		"testapp",
		"AI Rules",
		"Go",
		"Gin",
		"PostgreSQL",
		"Architectural Decisions",
		"Coding Conventions",
		"Constraints",
		"Memvra",
	}
	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("cursor export missing %q", check)
		}
	}
}

func TestMarkdownExporter(t *testing.T) {
	data := sampleExportData()
	exp, _ := Get("markdown")
	result, err := exp.Export(data)
	if err != nil {
		t.Fatalf("Export error: %v", err)
	}

	checks := []string{
		"testapp",
		"Tech Stack",
		"Go",
		"Gin",
		"Architectural Decisions",
		"Use PostgreSQL",
	}
	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("markdown export missing %q", check)
		}
	}
}

func TestJSONExporter(t *testing.T) {
	data := sampleExportData()
	exp, _ := Get("json")
	result, err := exp.Export(data)
	if err != nil {
		t.Fatalf("Export error: %v", err)
	}

	// Verify it's valid JSON.
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("JSON export is invalid JSON: %v", err)
	}

	// Check structure.
	if parsed["project"] == nil {
		t.Error("missing 'project' key")
	}
	if parsed["stack"] == nil {
		t.Error("missing 'stack' key")
	}
	if parsed["memories"] == nil {
		t.Error("missing 'memories' key")
	}

	// Check project fields.
	proj := parsed["project"].(map[string]interface{})
	if proj["name"] != "testapp" {
		t.Errorf("project name: got %v", proj["name"])
	}
}

func TestJSONExporter_EmptyMemories(t *testing.T) {
	data := ExportData{
		Project: memory.Project{Name: "empty"},
		Stack:   scanner.TechStack{},
	}
	exp, _ := Get("json")
	result, err := exp.Export(data)
	if err != nil {
		t.Fatalf("Export error: %v", err)
	}

	var parsed map[string]interface{}
	json.Unmarshal([]byte(result), &parsed)
	memories := parsed["memories"].(map[string]interface{})
	if len(memories) != 0 {
		t.Errorf("expected empty memories object, got %d keys", len(memories))
	}
}

func TestMemorySection(t *testing.T) {
	memories := []memory.Memory{
		{Content: "Use React", MemoryType: memory.TypeDecision},
		{Content: "Use Vue", MemoryType: memory.TypeDecision},
		{Content: "A note", MemoryType: memory.TypeNote},
	}

	result := memorySection("Decisions", memory.TypeDecision, memories)
	if !strings.Contains(result, "## Decisions") {
		t.Error("missing heading")
	}
	if !strings.Contains(result, "Use React") {
		t.Error("missing first decision")
	}
	if !strings.Contains(result, "Use Vue") {
		t.Error("missing second decision")
	}
	if strings.Contains(result, "A note") {
		t.Error("should not contain notes in decision section")
	}
}

func TestMemorySection_Empty(t *testing.T) {
	result := memorySection("Decisions", memory.TypeDecision, nil)
	if result != "" {
		t.Errorf("expected empty string for no memories, got %q", result)
	}
}
