package export

import (
	"fmt"
	"strings"

	"github.com/memvra/memvra/internal/memory"
)

// MarkdownExporter renders context as generic markdown.
type MarkdownExporter struct{}

func (e *MarkdownExporter) Export(data ExportData) (string, error) {
	ts := data.Stack
	proj := data.Project

	var b strings.Builder
	fmt.Fprintf(&b, "# %s — Project Context\n\n", proj.Name)

	b.WriteString(renderGitStateMarkdown(data.GitState))
	b.WriteString(renderSessionsMarkdown(data.Sessions))

	fmt.Fprintf(&b, "## Tech Stack\n\n")
	if ts.Language != "" {
		fmt.Fprintf(&b, "| Language | %s |\n", ts.Language)
	}
	if ts.Framework != "" {
		fmt.Fprintf(&b, "| Framework | %s |\n", ts.Framework)
	}
	if ts.Database != "" {
		fmt.Fprintf(&b, "| Database | %s |\n", ts.Database)
	}
	if ts.Architecture != "" {
		fmt.Fprintf(&b, "| Architecture | %s |\n", ts.Architecture)
	}
	if ts.TestFramework != "" {
		fmt.Fprintf(&b, "| Tests | %s |\n", ts.TestFramework)
	}
	b.WriteString("\n")

	for _, section := range []struct {
		heading string
		mt      memory.MemoryType
	}{
		{"Architectural Decisions", memory.TypeDecision},
		{"Coding Conventions", memory.TypeConvention},
		{"Constraints", memory.TypeConstraint},
		{"Notes", memory.TypeNote},
		{"TODOs", memory.TypeTodo},
	} {
		b.WriteString(memorySection(section.heading, section.mt, data.Memories))
	}

	return b.String(), nil
}
