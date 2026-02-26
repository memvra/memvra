package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/memvra/memvra/internal/db"
	"github.com/memvra/memvra/internal/memory"
)

func TestStripAnsi_Colors(t *testing.T) {
	input := "\x1b[32mhello\x1b[0m world"
	got := stripAnsi(input)
	want := "hello world"
	if got != want {
		t.Errorf("stripAnsi colors: got %q, want %q", got, want)
	}
}

func TestStripAnsi_CarriageReturn(t *testing.T) {
	input := "line one\r\nline two\r\n"
	got := stripAnsi(input)
	want := "line one\nline two"
	if got != want {
		t.Errorf("stripAnsi CR: got %q, want %q", got, want)
	}
}

func TestStripAnsi_CursorMovement(t *testing.T) {
	input := "\x1b[2J\x1b[Hhello\x1b[1A"
	got := stripAnsi(input)
	want := "hello"
	if got != want {
		t.Errorf("stripAnsi cursor: got %q, want %q", got, want)
	}
}

func TestStripAnsi_CollapseBlankLines(t *testing.T) {
	input := "line one\n\n\n\n\nline two"
	got := stripAnsi(input)
	want := "line one\n\nline two"
	if got != want {
		t.Errorf("stripAnsi collapse: got %q, want %q", got, want)
	}
}

func TestStripAnsi_PlainText(t *testing.T) {
	input := "hello world"
	got := stripAnsi(input)
	if got != input {
		t.Errorf("stripAnsi plain: got %q, want %q", got, input)
	}
}

func TestStripAnsi_Empty(t *testing.T) {
	got := stripAnsi("")
	if got != "" {
		t.Errorf("stripAnsi empty: got %q, want %q", got, "")
	}
}

func TestStripAnsi_BoldAndUnderline(t *testing.T) {
	input := "\x1b[1mbold\x1b[0m \x1b[4munderline\x1b[0m"
	got := stripAnsi(input)
	want := "bold underline"
	if got != want {
		t.Errorf("stripAnsi bold/underline: got %q, want %q", got, want)
	}
}

func TestStripAnsi_256Color(t *testing.T) {
	input := "\x1b[38;5;196mred\x1b[0m"
	got := stripAnsi(input)
	want := "red"
	if got != want {
		t.Errorf("stripAnsi 256color: got %q, want %q", got, want)
	}
}

func setupWrapTestStore(t *testing.T) *memory.Store {
	t.Helper()
	root := t.TempDir()
	dbDir := filepath.Join(root, ".memvra")
	os.MkdirAll(dbDir, 0o755)
	database, err := db.Open(filepath.Join(dbDir, "memvra.db"))
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return memory.NewStore(database)
}

func TestBuildWrapContext_WithSessions(t *testing.T) {
	store := setupWrapTestStore(t)

	store.InsertSessionReturningID(memory.Session{
		Question:        "Implementing JWT auth",
		ResponseSummary: "Created auth middleware with RS256",
		ModelUsed:       "claude",
	})

	got := buildWrapContext(store)
	if !strings.Contains(got, "Implementing JWT auth") {
		t.Error("expected session question in context")
	}
	if !strings.Contains(got, "Created auth middleware") {
		t.Error("expected session summary in context")
	}
	if !strings.Contains(got, "(claude)") {
		t.Error("expected model name in context")
	}
	if !strings.Contains(got, "previous AI sessions") {
		t.Error("expected preamble text")
	}
}

func TestBuildWrapContext_WithDecisions(t *testing.T) {
	store := setupWrapTestStore(t)

	store.InsertMemory(memory.Memory{
		Content:    "Use PostgreSQL for persistence",
		MemoryType: memory.TypeDecision,
		Importance: 0.8,
	})

	got := buildWrapContext(store)
	if !strings.Contains(got, "Key Decisions") {
		t.Error("expected decisions header")
	}
	if !strings.Contains(got, "Use PostgreSQL") {
		t.Error("expected decision content")
	}
}

func TestBuildWrapContext_WithTodos(t *testing.T) {
	store := setupWrapTestStore(t)

	store.InsertMemory(memory.Memory{
		Content:    "Add rate limiting to auth endpoints",
		MemoryType: memory.TypeTodo,
		Importance: 0.6,
	})

	got := buildWrapContext(store)
	if !strings.Contains(got, "TODOs") {
		t.Error("expected TODOs header")
	}
	if !strings.Contains(got, "rate limiting") {
		t.Error("expected todo content")
	}
}

func TestBuildWrapContext_Empty(t *testing.T) {
	store := setupWrapTestStore(t)

	got := buildWrapContext(store)
	if got != "" {
		t.Errorf("expected empty context for fresh project, got: %q", got)
	}
}

func TestBuildWrapContext_ContinuePrompt(t *testing.T) {
	store := setupWrapTestStore(t)

	store.InsertSessionReturningID(memory.Session{
		Question:  "test session",
		ModelUsed: "gemini",
	})

	got := buildWrapContext(store)
	if !strings.Contains(got, "continue from where the previous session left off") {
		t.Error("expected continue prompt at end")
	}
}

func TestWrapCmd_RequiresArgs(t *testing.T) {
	cmd := newWrapCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for no args")
	}
}

func TestWrapCmd_Help(t *testing.T) {
	cmd := newWrapCmd()
	cmd.SetArgs([]string{"--help"})
	// Should not error on --help.
	_ = cmd.Execute()
	if cmd.Use != "wrap <tool> [tool-args...]" {
		t.Errorf("unexpected Use: %q", cmd.Use)
	}
}
