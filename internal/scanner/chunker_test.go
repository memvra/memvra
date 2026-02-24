package scanner

import (
	"strings"
	"testing"
)

func TestChunkFile_SmallFile(t *testing.T) {
	content := "line1\nline2\nline3"
	chunks := ChunkFile(content, "code", 150)
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if chunks[0].Content != content {
		t.Errorf("content mismatch: got %q", chunks[0].Content)
	}
	if chunks[0].StartLine != 1 || chunks[0].EndLine != 3 {
		t.Errorf("line range: got %d-%d, want 1-3", chunks[0].StartLine, chunks[0].EndLine)
	}
	if chunks[0].ChunkType != "code" {
		t.Errorf("chunk type: got %q, want %q", chunks[0].ChunkType, "code")
	}
}

func TestChunkFile_LargeFile(t *testing.T) {
	lines := make([]string, 300)
	for i := range lines {
		lines[i] = "x"
	}
	content := strings.Join(lines, "\n")

	chunks := ChunkFile(content, "code", 150)
	if len(chunks) < 2 {
		t.Fatalf("expected at least 2 chunks for 300 lines, got %d", len(chunks))
	}
	// First chunk should start at line 1.
	if chunks[0].StartLine != 1 {
		t.Errorf("first chunk start: got %d, want 1", chunks[0].StartLine)
	}
	// Last chunk should end at line 300.
	last := chunks[len(chunks)-1]
	if last.EndLine != 300 {
		t.Errorf("last chunk end: got %d, want 300", last.EndLine)
	}
}

func TestChunkFile_DefaultMaxLines(t *testing.T) {
	content := "one\ntwo\nthree"
	chunks := ChunkFile(content, "code", 0)
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk with default max lines, got %d", len(chunks))
	}
}

func TestChunkFile_EmptyContent(t *testing.T) {
	chunks := ChunkFile("", "code", 150)
	// Empty string split produces [""] which is 1 line — should produce 1 chunk.
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk for empty content, got %d", len(chunks))
	}
}

func TestChunkFile_Markdown(t *testing.T) {
	content := `# Title

Intro paragraph.

## Section One

Some content here.
More content.

## Section Two

Other content.

### Subsection

Details.`

	chunks := ChunkFile(content, "docs", 150)
	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks for markdown with headings, got %d", len(chunks))
	}
	for _, c := range chunks {
		if c.ChunkType != "docs" {
			t.Errorf("expected chunk type 'docs', got %q", c.ChunkType)
		}
	}
}

func TestChunkFile_MarkdownForceSplit(t *testing.T) {
	// Create a markdown section longer than maxLines.
	var lines []string
	lines = append(lines, "## Big Section")
	for i := 0; i < 20; i++ {
		lines = append(lines, "content line")
	}
	content := strings.Join(lines, "\n")

	chunks := ChunkFile(content, "docs", 10)
	if len(chunks) < 2 {
		t.Fatalf("expected force-split on large markdown section, got %d chunks", len(chunks))
	}
}

func TestChunkByLines_Overlap(t *testing.T) {
	lines := make([]string, 50)
	for i := range lines {
		lines[i] = "x"
	}

	chunks := chunkByLines(lines, "code", 30, 10)
	if len(chunks) < 2 {
		t.Fatalf("expected at least 2 chunks, got %d", len(chunks))
	}
	// Second chunk should start overlapping with the first.
	if chunks[1].StartLine <= chunks[0].EndLine-10 {
		t.Errorf("expected overlap: chunk[0] ends at %d, chunk[1] starts at %d",
			chunks[0].EndLine, chunks[1].StartLine)
	}
}

func TestChunkByLines_TinyFinalChunk(t *testing.T) {
	// 155 lines with maxLines=150, overlap=10: the tail is 15 lines, overlap is 10.
	// The tiny tail (5 lines < overlap) should be merged into the previous chunk.
	lines := make([]string, 155)
	for i := range lines {
		lines[i] = "x"
	}
	chunks := chunkByLines(lines, "code", 150, 10)
	last := chunks[len(chunks)-1]
	if last.EndLine != 155 {
		t.Errorf("last chunk should cover through line 155, got %d", last.EndLine)
	}
}
