package scanner

import "testing"

func TestLanguageForFile(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"main.go", "go"},
		{"app.rb", "ruby"},
		{"server.py", "python"},
		{"index.js", "javascript"},
		{"index.mjs", "javascript"},
		{"app.ts", "typescript"},
		{"component.tsx", "tsx"},
		{"component.jsx", "jsx"},
		{"main.rs", "rust"},
		{"Main.java", "java"},
		{"main.kt", "kotlin"},
		{"App.cs", "csharp"},
		{"main.cpp", "cpp"},
		{"main.c", "c"},
		{"header.h", "c"},
		{"main.swift", "swift"},
		{"index.php", "php"},
		{"Main.scala", "scala"},
		{"app.ex", "elixir"},
		{"app.exs", "elixir"},
		{"Main.hs", "haskell"},
		{"script.lua", "lua"},
		{"deploy.sh", "bash"},
		{"script.bash", "bash"},
		{"query.sql", "sql"},
		{"index.html", "html"},
		{"style.css", "css"},
		{"style.scss", "scss"},
		{"App.vue", "vue"},
		{"App.svelte", "svelte"},
		{"data.json", "json"},
		{"config.yaml", "yaml"},
		{"config.yml", "yaml"},
		{"config.toml", "toml"},
		{"layout.xml", "xml"},
		{"README.md", "markdown"},
		{"doc.mdx", "markdown"},
		{"main.tf", "terraform"},
		{"schema.proto", "protobuf"},
		{"schema.graphql", "graphql"},
		{"schema.gql", "graphql"},
		{"Dockerfile", "Dockerfile"},
		{"Makefile", "Makefile"},
		{"Gemfile", "Gemfile"},
		// Unrecognised extension.
		{"photo.bmp", ""},
		{"random.xyz", ""},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := LanguageForFile(tt.path)
			if got != tt.want {
				t.Errorf("LanguageForFile(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestChunkTypeForFile(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"main_test.go", "test"},
		{"user_spec.rb", "test"},
		{"app.test.ts", "test"},
		{"app.test.js", "test"},
		{"app.spec.ts", "test"},
		{"app.spec.js", "test"},
		{"spec/models/user.rb", "code"}, // dirIs checks immediate parent only
		{"test/main_test.go", "test"},
		{"tests/test_app.py", "test"},
		{"__tests__/App.test.tsx", "test"},
		{"config.yaml", "config"},
		{"docker-compose.yml", "config"},
		{"settings.toml", "config"},
		{"package.json", "config"},
		{"Dockerfile", "config"},
		{"Makefile", "config"},
		{"README.md", "docs"},
		{"docs/guide.mdx", "docs"},
		{"main.go", "code"},
		{"app.ts", "code"},
		{"server.py", "code"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := ChunkTypeForFile(tt.path)
			if got != tt.want {
				t.Errorf("ChunkTypeForFile(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	if !contains("main_test.go", "_test.go") {
		t.Error("expected contains to match suffix")
	}
	if contains("main.go", "_test.go") {
		t.Error("expected contains to not match")
	}
}

func TestDirIs(t *testing.T) {
	if !dirIs("spec", "spec", "test") {
		t.Error("expected dirIs to match 'spec'")
	}
	if dirIs("src", "spec", "test") {
		t.Error("expected dirIs to not match 'src'")
	}
}
