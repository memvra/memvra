// Package scanner handles project directory scanning, tech-stack detection,
// and file chunking for Memvra.
package scanner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// TechStack holds the auto-detected project profile.
type TechStack struct {
	ProjectName      string   `json:"project_name"`
	Language         string   `json:"language"`
	Framework        string   `json:"framework"`
	FrameworkVersion string   `json:"framework_version,omitempty"`
	Database         string   `json:"database,omitempty"`
	Frontend         string   `json:"frontend,omitempty"`
	TestFramework    string   `json:"test_framework,omitempty"`
	Architecture     string   `json:"architecture_pattern,omitempty"`
	CI               string   `json:"ci,omitempty"`
	EntryPoints      []string `json:"entry_points,omitempty"`
	DetectedPatterns []string `json:"detected_patterns,omitempty"`
	FileCount        int      `json:"file_count"`
	ChunkCount       int      `json:"chunk_count"`
}

// ToJSON serialises the tech stack as a JSON string for storage.
func (ts TechStack) ToJSON() string {
	b, _ := json.Marshal(ts)
	return string(b)
}

// TechStackFromJSON parses a stored JSON string.
func TechStackFromJSON(s string) (TechStack, error) {
	var ts TechStack
	err := json.Unmarshal([]byte(s), &ts)
	return ts, err
}

// DetectTechStack inspects the project root and returns a best-effort profile.
func DetectTechStack(root string) TechStack {
	ts := TechStack{
		ProjectName: filepath.Base(root),
	}

	// Helper: check file existence.
	has := func(names ...string) bool {
		for _, n := range names {
			if _, err := os.Stat(filepath.Join(root, n)); err == nil {
				return true
			}
		}
		return false
	}

	// Read file content helper.
	read := func(name string) string {
		b, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			return ""
		}
		return string(b)
	}

	// ----- Language / Framework detection -----

	switch {
	case has("Gemfile", "Gemfile.lock"):
		ts.Language = "Ruby"
		gemfile := read("Gemfile")
		switch {
		case strings.Contains(gemfile, "rails"):
			ts.Framework = "Rails"
			ts.TestFramework = "RSpec"
			if has("config/routes.rb") {
				ts.EntryPoints = append(ts.EntryPoints, "config/routes.rb")
			}
		case strings.Contains(gemfile, "sinatra"):
			ts.Framework = "Sinatra"
		}
		if strings.Contains(gemfile, "rspec") {
			ts.TestFramework = "RSpec"
		}
		if strings.Contains(gemfile, "minitest") {
			ts.TestFramework = "Minitest"
		}
		if strings.Contains(gemfile, "sidekiq") {
			ts.DetectedPatterns = append(ts.DetectedPatterns, "background-jobs")
		}
		if strings.Contains(gemfile, "acts_as_tenant") {
			ts.DetectedPatterns = append(ts.DetectedPatterns, "multi-tenant")
		}

	case has("package.json"):
		pkgJSON := read("package.json")
		ts.Language = "JavaScript/TypeScript"
		switch {
		case strings.Contains(pkgJSON, `"next"`):
			ts.Framework = "Next.js"
		case strings.Contains(pkgJSON, `"react"`):
			ts.Framework = "React"
		case strings.Contains(pkgJSON, `"vue"`):
			ts.Framework = "Vue.js"
		case strings.Contains(pkgJSON, `"express"`):
			ts.Framework = "Express"
		case strings.Contains(pkgJSON, `"fastify"`):
			ts.Framework = "Fastify"
		case strings.Contains(pkgJSON, `"nest"`):
			ts.Framework = "NestJS"
		}
		if has("tsconfig.json") {
			ts.Language = "TypeScript"
		}
		if strings.Contains(pkgJSON, `"jest"`) {
			ts.TestFramework = "Jest"
		} else if strings.Contains(pkgJSON, `"vitest"`) {
			ts.TestFramework = "Vitest"
		}

	case has("go.mod"):
		ts.Language = "Go"
		goMod := read("go.mod")
		switch {
		case strings.Contains(goMod, "github.com/gin-gonic/gin"):
			ts.Framework = "Gin"
		case strings.Contains(goMod, "github.com/labstack/echo"):
			ts.Framework = "Echo"
		case strings.Contains(goMod, "github.com/gofiber/fiber"):
			ts.Framework = "Fiber"
		default:
			ts.Framework = "stdlib"
		}
		if has("cmd") {
			ts.EntryPoints = append(ts.EntryPoints, "cmd/")
		}

	case has("Cargo.toml"):
		ts.Language = "Rust"
		cargo := read("Cargo.toml")
		if strings.Contains(cargo, "actix") {
			ts.Framework = "Actix"
		} else if strings.Contains(cargo, "axum") {
			ts.Framework = "Axum"
		}

	case has("pyproject.toml", "requirements.txt", "setup.py"):
		ts.Language = "Python"
		req := read("requirements.txt") + read("pyproject.toml")
		switch {
		case strings.Contains(req, "django"):
			ts.Framework = "Django"
		case strings.Contains(req, "fastapi"):
			ts.Framework = "FastAPI"
		case strings.Contains(req, "flask"):
			ts.Framework = "Flask"
		}
		if strings.Contains(req, "pytest") {
			ts.TestFramework = "pytest"
		}

	case has("pom.xml", "build.gradle", "build.gradle.kts"):
		ts.Language = "Java/Kotlin"
		ts.Framework = "Spring Boot"
		if has("build.gradle.kts") {
			ts.Language = "Kotlin"
		}
	}

	// ----- Database detection -----
	allContent := read("Gemfile") + read("package.json") +
		read("docker-compose.yml") + read("docker-compose.yaml") +
		read(".env.example") + read("config/database.yml")

	switch {
	case strings.Contains(allContent, "postgresql") || strings.Contains(allContent, "postgres") || strings.Contains(allContent, "pg"):
		ts.Database = "PostgreSQL"
	case strings.Contains(allContent, "mysql"):
		ts.Database = "MySQL"
	case strings.Contains(allContent, "sqlite"):
		ts.Database = "SQLite"
	case strings.Contains(allContent, "mongodb") || strings.Contains(allContent, "mongoose"):
		ts.Database = "MongoDB"
	case strings.Contains(allContent, "redis"):
		ts.Database = "Redis"
	}

	// ----- CI detection -----
	switch {
	case has(".github/workflows"):
		ts.CI = "GitHub Actions"
	case has(".circleci/config.yml"):
		ts.CI = "CircleCI"
	case has(".gitlab-ci.yml"):
		ts.CI = "GitLab CI"
	case has("Jenkinsfile"):
		ts.CI = "Jenkins"
	}

	// ----- Architecture -----
	if has("config/routes.rb") && !has("app/views") {
		ts.Architecture = "API + SPA"
	} else if has("config/routes.rb") {
		ts.Architecture = "MVC (Monolith)"
	}

	return ts
}
