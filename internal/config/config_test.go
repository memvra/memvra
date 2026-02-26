package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultGlobal(t *testing.T) {
	cfg := DefaultGlobal()

	if cfg.DefaultModel != "claude" {
		t.Errorf("default model: got %q, want %q", cfg.DefaultModel, "claude")
	}
	if cfg.DefaultEmbedder != "ollama" {
		t.Errorf("default embedder: got %q, want %q", cfg.DefaultEmbedder, "ollama")
	}
	if cfg.Context.MaxTokens != 8000 {
		t.Errorf("max tokens: got %d, want 8000", cfg.Context.MaxTokens)
	}
	if cfg.Context.ChunkMaxLines != 150 {
		t.Errorf("chunk max lines: got %d, want 150", cfg.Context.ChunkMaxLines)
	}
	if cfg.Context.SimilarityThreshold != 0.3 {
		t.Errorf("similarity threshold: got %f, want 0.3", cfg.Context.SimilarityThreshold)
	}
	if cfg.Context.TopKChunks != 10 {
		t.Errorf("top k chunks: got %d, want 10", cfg.Context.TopKChunks)
	}
	if cfg.Context.TopKMemories != 5 {
		t.Errorf("top k memories: got %d, want 5", cfg.Context.TopKMemories)
	}
	if !cfg.Output.Stream {
		t.Error("stream should default to true")
	}
	if !cfg.Output.Color {
		t.Error("color should default to true")
	}
	if cfg.Extraction.Enabled {
		t.Error("extraction should default to disabled")
	}
	if cfg.Extraction.MaxExtracts != 3 {
		t.Errorf("max extracts: got %d, want 3", cfg.Extraction.MaxExtracts)
	}
	if !cfg.Summarization.Enabled {
		t.Error("summarization should default to enabled")
	}
	if cfg.Context.TopKSessions != 3 {
		t.Errorf("top k sessions: got %d, want 3", cfg.Context.TopKSessions)
	}
	if !cfg.AutoExport.Enabled {
		t.Error("auto export should default to enabled")
	}
	if len(cfg.AutoExport.Formats) != 4 {
		t.Errorf("auto export formats: got %d, want 4", len(cfg.AutoExport.Formats))
	}
	if cfg.Ollama.Host != "http://localhost:11434" {
		t.Errorf("ollama host: got %q", cfg.Ollama.Host)
	}
	if cfg.Ollama.EmbedModel != "nomic-embed-text" {
		t.Errorf("ollama embed model: got %q", cfg.Ollama.EmbedModel)
	}
}

func TestProjectDBPath(t *testing.T) {
	got := ProjectDBPath("/home/user/project")
	want := filepath.Join("/home/user/project", ".memvra", "memvra.db")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestProjectConfigDirPath(t *testing.T) {
	got := ProjectConfigDirPath("/home/user/project")
	want := filepath.Join("/home/user/project", ".memvra")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLoadProject_NoFile(t *testing.T) {
	cfg, err := LoadProject(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return zero-value config with no error.
	if cfg.DefaultModel != "" {
		t.Errorf("expected empty default model, got %q", cfg.DefaultModel)
	}
}

func TestSaveAndLoadProject(t *testing.T) {
	dir := t.TempDir()
	cfg := ProjectConfig{
		DefaultModel: "openai",
		Project:      ProjectMeta{Name: "testproj"},
		AlwaysInclude: []string{"README.md"},
	}

	if err := SaveProject(dir, cfg); err != nil {
		t.Fatalf("SaveProject: %v", err)
	}

	loaded, err := LoadProject(dir)
	if err != nil {
		t.Fatalf("LoadProject: %v", err)
	}
	if loaded.DefaultModel != "openai" {
		t.Errorf("default model: got %q, want %q", loaded.DefaultModel, "openai")
	}
	if loaded.Project.Name != "testproj" {
		t.Errorf("project name: got %q, want %q", loaded.Project.Name, "testproj")
	}
}

func TestLoad_MergesProjectOverrides(t *testing.T) {
	dir := t.TempDir()

	// Save a project config with model override.
	SaveProject(dir, ProjectConfig{DefaultModel: "openai"})

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DefaultModel != "openai" {
		t.Errorf("expected project override 'openai', got %q", cfg.DefaultModel)
	}
}

func TestLoadGlobal_EnvOverrides(t *testing.T) {
	// Set env vars and verify they override config.
	os.Setenv("ANTHROPIC_API_KEY", "test-key-123")
	defer os.Unsetenv("ANTHROPIC_API_KEY")

	cfg, err := LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal: %v", err)
	}
	if cfg.Keys.Anthropic != "test-key-123" {
		t.Errorf("expected env override, got %q", cfg.Keys.Anthropic)
	}
}

func TestGlobalConfigPath(t *testing.T) {
	path, err := GlobalConfigPath()
	if err != nil {
		t.Fatalf("GlobalConfigPath: %v", err)
	}
	if !filepath.IsAbs(path) {
		t.Errorf("expected absolute path, got %q", path)
	}
	if filepath.Base(path) != "config.toml" {
		t.Errorf("expected config.toml, got %q", filepath.Base(path))
	}
}
