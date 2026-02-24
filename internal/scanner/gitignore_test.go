package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHardIgnore(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"node_modules", true},
		{"vendor", true},
		{".git", true},
		{"dist", true},
		{".memvra", true},
		{"__pycache__", true},
		{"src", false},
		{"internal", false},
		{"cmd", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HardIgnore(tt.name)
			if got != tt.want {
				t.Errorf("HardIgnore(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestSkipFile(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"photo.png", true},
		{"image.jpg", true},
		{"archive.zip", true},
		{"binary.exe", true},
		{"Gemfile.lock", true},
		{"package-lock.json", true},
		{"yarn.lock", true},
		{"go.sum", true},
		{"main.go", false},
		{"app.ts", false},
		{"README.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SkipFile(tt.name)
			if got != tt.want {
				t.Errorf("SkipFile(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestIgnoreMatcher_NoGitignore(t *testing.T) {
	m := NewIgnoreMatcher("/tmp/memvra-nonexistent-dir")
	if m.Match("anything.go") {
		t.Error("expected no-op matcher to accept all files")
	}
}

func TestIgnoreMatcher_WithGitignore(t *testing.T) {
	dir := t.TempDir()
	gitignoreContent := "*.log\nbuild/\n"
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(gitignoreContent), 0o644); err != nil {
		t.Fatal(err)
	}

	m := NewIgnoreMatcher(dir)
	if !m.Match("debug.log") {
		t.Error("expected .log files to be ignored")
	}
	if !m.Match("build/output.js") {
		t.Error("expected build/ dir to be ignored")
	}
	if m.Match("main.go") {
		t.Error("expected main.go to NOT be ignored")
	}
}
