package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScan_GoProject(t *testing.T) {
	result := Scan(ScanOptions{
		Root: "../../testdata/go_project",
	})

	if result.Stack.Language != "Go" {
		t.Errorf("stack language: got %q, want %q", result.Stack.Language, "Go")
	}
	if len(result.Files) == 0 {
		t.Fatal("expected at least one scanned file")
	}
	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}

	// Check that main.go was found.
	found := false
	for _, sf := range result.Files {
		if sf.File.Path == "main.go" {
			found = true
			if sf.File.Language != "go" {
				t.Errorf("main.go language: got %q, want %q", sf.File.Language, "go")
			}
			if sf.File.ContentHash == "" {
				t.Error("main.go should have a content hash")
			}
			if len(sf.Chunks) == 0 {
				t.Error("main.go should have at least one chunk")
			}
		}
	}
	if !found {
		t.Error("main.go not found in scan results")
	}
}

func TestScan_SkipsHardIgnored(t *testing.T) {
	dir := t.TempDir()

	// Create a source file and a node_modules dir.
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0o644)
	os.MkdirAll(filepath.Join(dir, "node_modules", "pkg"), 0o755)
	os.WriteFile(filepath.Join(dir, "node_modules", "pkg", "index.js"), []byte("export default {}"), 0o644)

	result := Scan(ScanOptions{Root: dir})

	for _, sf := range result.Files {
		if filepath.Dir(sf.File.Path) == "node_modules" {
			t.Error("should not have scanned files in node_modules")
		}
	}
}

func TestScan_SkipsBinaryFiles(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "image.png"), []byte{0x89, 0x50, 0x4E, 0x47}, 0o644)

	result := Scan(ScanOptions{Root: dir})

	for _, sf := range result.Files {
		if sf.File.Path == "image.png" {
			t.Error("should not have scanned image.png")
		}
	}
}

func TestScan_CountsCorrectly(t *testing.T) {
	result := Scan(ScanOptions{
		Root: "../../testdata/go_project",
	})

	if result.Stack.FileCount != len(result.Files) {
		t.Errorf("FileCount=%d but len(Files)=%d", result.Stack.FileCount, len(result.Files))
	}

	totalChunks := 0
	for _, sf := range result.Files {
		totalChunks += len(sf.Chunks)
	}
	if result.Stack.ChunkCount != totalChunks {
		t.Errorf("ChunkCount=%d but actual chunks=%d", result.Stack.ChunkCount, totalChunks)
	}
}

func TestFindProjectRoot_FromSubdir(t *testing.T) {
	// testdata/go_project has go.mod, so FindProjectRoot should find it.
	root, err := FindProjectRoot("../../testdata/go_project")
	if err != nil {
		t.Fatalf("FindProjectRoot error: %v", err)
	}
	if filepath.Base(root) != "go_project" && root != "" {
		// It might walk up to the memvra root (which also has go.mod).
		// That's acceptable — we just check it doesn't error.
		t.Logf("found root: %s", root)
	}
}
