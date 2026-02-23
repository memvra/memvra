// Package export renders Memvra context into formats compatible with other tools.
package export

import (
	"fmt"

	"github.com/memvra/memvra/internal/memory"
	"github.com/memvra/memvra/internal/scanner"
)

// ExportData is passed to every Exporter.
type ExportData struct {
	Project  memory.Project
	Stack    scanner.TechStack
	Memories []memory.Memory
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
