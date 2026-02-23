package scanner

import (
	"strings"
)

const (
	DefaultMaxLines = 150
	DefaultOverlap  = 10
)

// RawChunk holds a slice of a source file before it is persisted.
type RawChunk struct {
	Content   string
	StartLine int // 1-based
	EndLine   int // 1-based, inclusive
	ChunkType string
}

// ChunkFile splits the file content into overlapping chunks.
// chunkType should be one of "code", "config", "test", "docs".
func ChunkFile(content, chunkType string, maxLines int) []RawChunk {
	if maxLines <= 0 {
		maxLines = DefaultMaxLines
	}

	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return nil
	}

	// For markdown files, prefer splitting by headings.
	if chunkType == "docs" {
		return chunkMarkdown(lines, maxLines)
	}

	return chunkByLines(lines, chunkType, maxLines, DefaultOverlap)
}

// chunkByLines performs simple line-based chunking with overlap.
func chunkByLines(lines []string, chunkType string, maxLines, overlap int) []RawChunk {
	total := len(lines)
	if total <= maxLines {
		return []RawChunk{{
			Content:   strings.Join(lines, "\n"),
			StartLine: 1,
			EndLine:   total,
			ChunkType: chunkType,
		}}
	}

	var chunks []RawChunk
	start := 0

	for start < total {
		end := start + maxLines
		if end > total {
			end = total
		}

		chunks = append(chunks, RawChunk{
			Content:   strings.Join(lines[start:end], "\n"),
			StartLine: start + 1,
			EndLine:   end,
			ChunkType: chunkType,
		})

		// Advance by maxLines minus overlap.
		advance := maxLines - overlap
		if advance <= 0 {
			advance = maxLines
		}
		start += advance

		// Avoid a tiny final chunk — merge into previous.
		if total-start < overlap && start < total {
			last := &chunks[len(chunks)-1]
			last.Content = strings.Join(lines[last.StartLine-1:total], "\n")
			last.EndLine = total
			break
		}
	}

	return chunks
}

// chunkMarkdown splits markdown by headings (## or ###).
func chunkMarkdown(lines []string, maxLines int) []RawChunk {
	var chunks []RawChunk
	var current []string
	startLine := 1

	flush := func(endLine int) {
		if len(current) == 0 {
			return
		}
		chunks = append(chunks, RawChunk{
			Content:   strings.Join(current, "\n"),
			StartLine: startLine,
			EndLine:   endLine,
			ChunkType: "docs",
		})
		current = nil
	}

	for i, line := range lines {
		lineNum := i + 1
		isHeading := strings.HasPrefix(line, "## ") || strings.HasPrefix(line, "### ")

		if isHeading && len(current) > 0 {
			flush(lineNum - 1)
			startLine = lineNum
		}

		current = append(current, line)

		// Force-split if chunk is getting too large.
		if len(current) >= maxLines {
			flush(lineNum)
			startLine = lineNum + 1
		}
	}

	flush(len(lines))
	return chunks
}
