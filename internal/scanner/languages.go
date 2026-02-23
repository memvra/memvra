package scanner

import "path/filepath"

// LanguageForFile returns the language name for a given file path.
// Returns "" if the extension is not recognised.
func LanguageForFile(path string) string {
	ext := filepath.Ext(path)
	switch ext {
	case ".go":
		return "go"
	case ".rb":
		return "ruby"
	case ".py":
		return "python"
	case ".js", ".mjs", ".cjs":
		return "javascript"
	case ".ts", ".mts", ".cts":
		return "typescript"
	case ".tsx":
		return "tsx"
	case ".jsx":
		return "jsx"
	case ".rs":
		return "rust"
	case ".java":
		return "java"
	case ".kt":
		return "kotlin"
	case ".cs":
		return "csharp"
	case ".cpp", ".cc", ".cxx":
		return "cpp"
	case ".c":
		return "c"
	case ".h", ".hpp":
		return "c"
	case ".swift":
		return "swift"
	case ".php":
		return "php"
	case ".scala":
		return "scala"
	case ".ex", ".exs":
		return "elixir"
	case ".hs":
		return "haskell"
	case ".lua":
		return "lua"
	case ".sh", ".bash", ".zsh":
		return "bash"
	case ".sql":
		return "sql"
	case ".html", ".htm":
		return "html"
	case ".css":
		return "css"
	case ".scss", ".sass":
		return "scss"
	case ".vue":
		return "vue"
	case ".svelte":
		return "svelte"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".toml":
		return "toml"
	case ".xml":
		return "xml"
	case ".md", ".mdx":
		return "markdown"
	case ".tf":
		return "terraform"
	case ".proto":
		return "protobuf"
	case ".graphql", ".gql":
		return "graphql"
	case ".dockerfile", "":
		name := filepath.Base(path)
		if name == "Dockerfile" || name == "Makefile" || name == "Gemfile" {
			return name
		}
	}
	return ""
}

// ChunkTypeForFile returns the chunk classification for a file.
func ChunkTypeForFile(path string) string {
	name := filepath.Base(path)
	dir := filepath.Dir(path)

	// Test files.
	switch {
	case contains(name, "_test.go", "_spec.rb", ".test.ts", ".test.js", ".spec.ts", ".spec.js"):
		return "test"
	case dirIs(dir, "spec", "test", "tests", "__tests__"):
		return "test"
	// Config files.
	case filepath.Ext(name) == ".yaml" || filepath.Ext(name) == ".yml" ||
		filepath.Ext(name) == ".toml" || filepath.Ext(name) == ".json" ||
		name == "Dockerfile" || name == "Makefile":
		return "config"
	// Docs.
	case filepath.Ext(name) == ".md" || filepath.Ext(name) == ".mdx":
		return "docs"
	}
	return "code"
}

func contains(s string, substrings ...string) bool {
	for _, sub := range substrings {
		if len(s) >= len(sub) && (s[len(s)-len(sub):] == sub) {
			return true
		}
	}
	return false
}

func dirIs(dir string, names ...string) bool {
	base := filepath.Base(dir)
	for _, n := range names {
		if base == n {
			return true
		}
	}
	return false
}
