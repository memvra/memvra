package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fsnotify/fsnotify"

	"github.com/memvra/memvra/internal/scanner"
)

func TestShouldIgnoreEvent(t *testing.T) {
	dir := t.TempDir()
	ignore := scanner.NewIgnoreMatcher(dir)

	tests := []struct {
		rel  string
		want bool
	}{
		{"main.go", false},
		{"src/app.go", false},
		{"node_modules/pkg/index.js", true},
		{".git/HEAD", true},
		{".memvra/memvra.db", true},
		{"vendor/lib/lib.go", true},
	}

	for _, tt := range tests {
		got := shouldIgnoreEvent(tt.rel, ignore)
		if got != tt.want {
			t.Errorf("shouldIgnoreEvent(%q) = %v, want %v", tt.rel, got, tt.want)
		}
	}
}

func TestAddWatchDirs_SkipsIgnored(t *testing.T) {
	dir := t.TempDir()

	// Create a normal dir and a hard-ignored dir.
	os.MkdirAll(filepath.Join(dir, "src"), 0o755)
	os.MkdirAll(filepath.Join(dir, "node_modules", "pkg"), 0o755)
	os.MkdirAll(filepath.Join(dir, ".git", "objects"), 0o755)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatalf("NewWatcher: %v", err)
	}
	defer watcher.Close()

	ignore := scanner.NewIgnoreMatcher(dir)
	if err := addWatchDirs(watcher, dir, ignore); err != nil {
		t.Fatalf("addWatchDirs: %v", err)
	}

	watchList := watcher.WatchList()
	watched := make(map[string]bool)
	for _, p := range watchList {
		rel, _ := filepath.Rel(dir, p)
		watched[rel] = true
	}

	if !watched["."] {
		t.Error("root directory should be watched")
	}
	if !watched["src"] {
		t.Error("src/ should be watched")
	}
	if watched["node_modules"] || watched[filepath.Join("node_modules", "pkg")] {
		t.Error("node_modules should not be watched")
	}
	if watched[".git"] || watched[filepath.Join(".git", "objects")] {
		t.Error(".git should not be watched")
	}
}
