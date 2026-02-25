package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/memvra/memvra/internal/config"
	"github.com/memvra/memvra/internal/export"
	"github.com/memvra/memvra/internal/memory"
	"github.com/memvra/memvra/internal/scanner"
)

// formatToFilename maps an export format name to the file it should be written to.
func formatToFilename(format string) string {
	switch format {
	case "claude":
		return "CLAUDE.md"
	case "cursor":
		return ".cursorrules"
	case "markdown":
		return "PROJECT_CONTEXT.md"
	case "json":
		return "memvra-context.json"
	default:
		return ""
	}
}

// autoExportFilenames returns the filenames that auto-export would generate
// for the given config.
func autoExportFilenames(cfg config.AutoExportConfig) []string {
	var names []string
	for _, f := range cfg.Formats {
		if name := formatToFilename(f); name != "" {
			names = append(names, name)
		}
	}
	return names
}

// autoExport regenerates all configured export files in the project root.
// It is best-effort: failures are logged to stderr but never abort the caller.
func autoExport(root string, store *memory.Store) {
	gcfg, _ := config.Load(root)
	if !gcfg.AutoExport.Enabled || len(gcfg.AutoExport.Formats) == 0 {
		return
	}

	proj, err := store.GetProject()
	if err != nil {
		return
	}
	ts, _ := scanner.TechStackFromJSON(proj.TechStack)

	memories, err := store.ListMemories("")
	if err != nil {
		return
	}

	data := export.ExportData{
		Project:  proj,
		Stack:    ts,
		Memories: memories,
	}

	var exported []string
	for _, format := range gcfg.AutoExport.Formats {
		exporter, ok := export.Get(format)
		if !ok {
			continue
		}
		output, err := exporter.Export(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  warn: auto-export %s failed: %v\n", format, err)
			continue
		}

		filename := formatToFilename(format)
		if filename == "" {
			continue
		}
		outPath := filepath.Join(root, filename)
		if err := os.WriteFile(outPath, []byte(output), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "  warn: write %s failed: %v\n", filename, err)
			continue
		}
		exported = append(exported, filename)
	}

	if len(exported) > 0 {
		fmt.Fprintf(os.Stderr, "  auto-exported: %s\n", strings.Join(exported, ", "))
	}
}
