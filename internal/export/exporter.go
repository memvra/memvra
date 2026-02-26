// Package export renders Memvra context into formats compatible with other tools.
package export

import (
	"fmt"
	"strings"

	"github.com/memvra/memvra/internal/git"
	"github.com/memvra/memvra/internal/memory"
	"github.com/memvra/memvra/internal/scanner"
)

// ExportData is passed to every Exporter.
type ExportData struct {
	Project  memory.Project
	Stack    scanner.TechStack
	Memories []memory.Memory
	Sessions []memory.Session
	GitState git.WorkingState
}

// Exporter renders ExportData to a string in a specific format.
type Exporter interface {
	Export(data ExportData) (string, error)
}

// registry maps format names to Exporter implementations.
var registry = map[string]Exporter{
	"claude":   &ClaudeMDExporter{},
	"cursor":   &CursorRulesExporter{},
	"markdown": &MarkdownExporter{},
	"json":     &JSONExporter{},
}

// Get returns the Exporter registered under name, and whether it was found.
func Get(name string) (Exporter, bool) {
	e, ok := registry[name]
	return e, ok
}

// ValidFormats returns the list of supported export format names.
func ValidFormats() []string {
	formats := make([]string, 0, len(registry))
	for k := range registry {
		formats = append(formats, k)
	}
	return formats
}

// memorySection renders memories of the given type as a markdown list block.
func memorySection(heading string, memType memory.MemoryType, memories []memory.Memory) string {
	var items []memory.Memory
	for _, m := range memories {
		if m.MemoryType == memType {
			items = append(items, m)
		}
	}
	if len(items) == 0 {
		return ""
	}
	out := fmt.Sprintf("## %s\n\n", heading)
	for _, m := range items {
		out += fmt.Sprintf("- %s\n", m.Content)
	}
	out += "\n"
	return out
}

// renderGitStateMarkdown renders the git working state as a markdown section.
func renderGitStateMarkdown(gs git.WorkingState) string {
	if gs.IsEmpty() || !gs.HasChanges() {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "## Work in Progress\n\n")

	if gs.Branch != "" {
		fmt.Fprintf(&b, "**Branch:** `%s`\n\n", gs.Branch)
	}

	if len(gs.Staged) > 0 {
		b.WriteString("**Staged for commit:**\n")
		for _, f := range gs.Staged {
			fmt.Fprintf(&b, "- `%s`\n", f)
		}
		b.WriteString("\n")
	}

	if len(gs.Modified) > 0 {
		b.WriteString("**Modified (unstaged):**\n")
		for _, f := range gs.Modified {
			fmt.Fprintf(&b, "- `%s`\n", f)
		}
		b.WriteString("\n")
	}

	if len(gs.Untracked) > 0 {
		b.WriteString("**New files (untracked):**\n")
		for _, f := range gs.Untracked {
			fmt.Fprintf(&b, "- `%s`\n", f)
		}
		b.WriteString("\n")
	}

	if gs.DiffStat != "" {
		fmt.Fprintf(&b, "**Change summary:**\n```\n%s\n```\n\n", gs.DiffStat)
	}

	return b.String()
}

// renderSessionsMarkdown renders recent sessions as a markdown section.
// Sessions are assumed newest-first; they are reversed to chronological order.
func renderSessionsMarkdown(sessions []memory.Session) string {
	if len(sessions) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("## Recent Activity\n\n")

	for i := len(sessions) - 1; i >= 0; i-- {
		s := sessions[i]
		ts := s.CreatedAt.Format("2006-01-02 15:04")
		model := ""
		if s.ModelUsed != "" {
			model = " (" + s.ModelUsed + ")"
		}
		fmt.Fprintf(&b, "**[%s]%s** %s\n", ts, model, s.Question)
		if s.ResponseSummary != "" {
			fmt.Fprintf(&b, "%s\n", s.ResponseSummary)
		}
		b.WriteString("\n")
	}

	return b.String()
}

// renderGitStatePlainText renders git state in plain-text format for .cursorrules.
func renderGitStatePlainText(gs git.WorkingState) string {
	if gs.IsEmpty() || !gs.HasChanges() {
		return ""
	}

	var b strings.Builder
	b.WriteString("# Work in Progress\n")

	if gs.Branch != "" {
		fmt.Fprintf(&b, "Current branch: %s\n", gs.Branch)
	}

	if len(gs.Staged) > 0 {
		b.WriteString("Staged for commit:\n")
		for _, f := range gs.Staged {
			fmt.Fprintf(&b, "  - %s\n", f)
		}
	}

	if len(gs.Modified) > 0 {
		b.WriteString("Modified (unstaged):\n")
		for _, f := range gs.Modified {
			fmt.Fprintf(&b, "  - %s\n", f)
		}
	}

	if len(gs.Untracked) > 0 {
		b.WriteString("New files (untracked):\n")
		for _, f := range gs.Untracked {
			fmt.Fprintf(&b, "  - %s\n", f)
		}
	}

	b.WriteString("\n")
	return b.String()
}

// renderSessionsPlainText renders recent sessions in plain-text format.
func renderSessionsPlainText(sessions []memory.Session) string {
	if len(sessions) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("# Recent Activity\n")

	for i := len(sessions) - 1; i >= 0; i-- {
		s := sessions[i]
		ts := s.CreatedAt.Format("2006-01-02 15:04")
		model := ""
		if s.ModelUsed != "" {
			model = " (" + s.ModelUsed + ")"
		}
		fmt.Fprintf(&b, "[%s]%s %s\n", ts, model, s.Question)
		if s.ResponseSummary != "" {
			fmt.Fprintf(&b, "  %s\n", s.ResponseSummary)
		}
	}

	b.WriteString("\n")
	return b.String()
}
