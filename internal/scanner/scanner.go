package scanner

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/memvra/memvra/internal/memory"
)

// ScanResult holds the output of a full project scan.
type ScanResult struct {
	Stack  TechStack
	Files  []ScannedFile
	Errors []error
}

// ScannedFile pairs a file record with its chunks.
type ScannedFile struct {
	File   memory.File
	Chunks []memory.Chunk // FileID is empty here; set after the file is stored.
}

// ScanOptions controls scanner behaviour.
type ScanOptions struct {
	Root         string
	MaxChunkLines int
	ExcludeGlobs []string
}

// Scan walks the project tree, hashes files, and splits them into chunks.
// It does NOT write to the database — that is the caller's responsibility.
func Scan(opts ScanOptions) ScanResult {
	root := opts.Root
	maxLines := opts.MaxChunkLines
	if maxLines == 0 {
		maxLines = DefaultMaxLines
	}

	ignore := NewIgnoreMatcher(root)
	stack := DetectTechStack(root)

	var result ScanResult
	result.Stack = stack

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, err)
			return nil // Skip unreadable entries.
		}

		// Get relative path.
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		if rel == "." {
			return nil
		}

		// Skip hard-ignored directories.
		if d.IsDir() {
			if HardIgnore(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip files by name/extension.
		if SkipFile(d.Name()) {
			return nil
		}

		// Skip files matched by .gitignore.
		if ignore.Match(rel) {
			return nil
		}

		lang := LanguageForFile(path)
		if lang == "" {
			return nil // Not a recognised source file.
		}

		chunkType := ChunkTypeForFile(rel)

		// Read and hash the file.
		content, err := os.ReadFile(path)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("read %s: %w", rel, err))
			return nil
		}

		hash := fmt.Sprintf("%x", sha256.Sum256(content))

		info, _ := d.Info()
		var modTime time.Time
		if info != nil {
			modTime = info.ModTime()
		}

		sf := ScannedFile{
			File: memory.File{
				Path:         rel,
				Language:     lang,
				LastModified: modTime,
				ContentHash:  hash,
			},
		}

		rawChunks := ChunkFile(string(content), chunkType, maxLines)
		for _, rc := range rawChunks {
			sf.Chunks = append(sf.Chunks, memory.Chunk{
				Content:   rc.Content,
				StartLine: rc.StartLine,
				EndLine:   rc.EndLine,
				ChunkType: rc.ChunkType,
			})
		}

		result.Files = append(result.Files, sf)
		return nil
	})

	if err != nil {
		result.Errors = append(result.Errors, err)
	}

	stack.FileCount = len(result.Files)
	for _, sf := range result.Files {
		stack.ChunkCount += len(sf.Chunks)
	}
	result.Stack = stack

	return result
}

// ScanFile scans a single file and returns a ScannedFile.
// relPath is relative to root. Returns nil if the file should be skipped
// (binary, gitignored, unrecognised language, etc).
func ScanFile(root, relPath string, maxChunkLines int, ignore *IgnoreMatcher) (*ScannedFile, error) {
	if maxChunkLines == 0 {
		maxChunkLines = DefaultMaxLines
	}

	name := filepath.Base(relPath)

	// Check hard-ignore on every directory component.
	for _, part := range strings.Split(filepath.Dir(relPath), string(filepath.Separator)) {
		if HardIgnore(part) {
			return nil, nil
		}
	}

	if SkipFile(name) {
		return nil, nil
	}

	if ignore != nil && ignore.Match(relPath) {
		return nil, nil
	}

	lang := LanguageForFile(relPath)
	if lang == "" {
		return nil, nil
	}

	absPath := filepath.Join(root, relPath)
	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", relPath, err)
	}

	hash := fmt.Sprintf("%x", sha256.Sum256(content))

	info, err := os.Stat(absPath)
	var modTime time.Time
	if err == nil {
		modTime = info.ModTime()
	}

	chunkType := ChunkTypeForFile(relPath)
	rawChunks := ChunkFile(string(content), chunkType, maxChunkLines)

	sf := &ScannedFile{
		File: memory.File{
			Path:         relPath,
			Language:     lang,
			LastModified: modTime,
			ContentHash:  hash,
		},
	}

	for _, rc := range rawChunks {
		sf.Chunks = append(sf.Chunks, memory.Chunk{
			Content:   rc.Content,
			StartLine: rc.StartLine,
			EndLine:   rc.EndLine,
			ChunkType: rc.ChunkType,
		})
	}

	return sf, nil
}

// FindProjectRoot walks up from startDir looking for a project root marker.
func FindProjectRoot(startDir string) (string, error) {
	markers := []string{".git", "go.mod", "package.json", "Gemfile", "Cargo.toml",
		"pyproject.toml", "requirements.txt", "pom.xml", "build.gradle"}

	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}

	for {
		for _, marker := range markers {
			if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
				return dir, nil
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root.
			return startDir, nil // Fall back to cwd.
		}
		dir = parent
	}
}
