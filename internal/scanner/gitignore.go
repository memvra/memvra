package scanner

import (
	"os"
	"path/filepath"

	gitignore "github.com/sabhiram/go-gitignore"
)

// IgnoreMatcher wraps a gitignore pattern matcher.
type IgnoreMatcher struct {
	gi *gitignore.GitIgnore
}

// NewIgnoreMatcher loads .gitignore from the project root.
// If no .gitignore file is found, the matcher accepts everything.
func NewIgnoreMatcher(root string) *IgnoreMatcher {
	path := filepath.Join(root, ".gitignore")
	if _, err := os.Stat(path); err != nil {
		return &IgnoreMatcher{}
	}
	gi, err := gitignore.CompileIgnoreFile(path)
	if err != nil {
		return &IgnoreMatcher{}
	}
	return &IgnoreMatcher{gi: gi}
}

// Match returns true if the given relative path should be ignored.
func (m *IgnoreMatcher) Match(relPath string) bool {
	if m.gi == nil {
		return false
	}
	return m.gi.MatchesPath(relPath)
}

// hardIgnored contains paths that are always skipped regardless of .gitignore.
var hardIgnored = map[string]bool{
	"node_modules":     true,
	"vendor":           true,
	".git":             true,
	"dist":             true,
	"build":            true,
	".memvra":          true,
	"__pycache__":      true,
	".venv":            true,
	"venv":             true,
	"env":              true,
	".bundle":          true,
	"tmp":              true,
	"log":              true,
	"coverage":         true,
	".nyc_output":      true,
	"target":           true, // Rust/Java build output
}

// HardIgnore returns true if the directory name is always excluded.
func HardIgnore(name string) bool {
	return hardIgnored[name]
}

// SkipExtensions contains file extensions we never index.
var skipExtensions = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
	".svg": true, ".ico": true, ".webp": true,
	".pdf": true, ".doc": true, ".docx": true,
	".zip": true, ".tar": true, ".gz": true, ".tgz": true, ".rar": true,
	".exe": true, ".bin": true, ".dll": true, ".so": true, ".dylib": true,
	".lock": true, // Gemfile.lock, package-lock.json, yarn.lock, etc.
	".sum":  true, // go.sum
	".min.js": true,
	".map":    true,
}

// SkipFile returns true for files we should never index.
func SkipFile(name string) bool {
	ext := filepath.Ext(name)
	if skipExtensions[ext] {
		return true
	}
	// Lock files by full name.
	switch name {
	case "Gemfile.lock", "package-lock.json", "yarn.lock", "go.sum",
		"Cargo.lock", "composer.lock", "poetry.lock", "Pipfile.lock":
		return true
	}
	return false
}
