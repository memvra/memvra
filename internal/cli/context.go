package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/memvra/memvra/internal/config"
	"github.com/memvra/memvra/internal/db"
	"github.com/memvra/memvra/internal/memory"
	"github.com/memvra/memvra/internal/scanner"
)

func newContextCmd() *cobra.Command {
	var section string
	var export bool

	cmd := &cobra.Command{
		Use:   "context",
		Short: "View the current project context that would be injected into LLM calls",
		Long: `Print a human-readable summary of what Memvra knows about this project.

Sections: profile, decisions, conventions, constraints, notes, todos

Examples:
  memvra context
  memvra context --section decisions
  memvra context --export`,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := findRoot()
			if err != nil {
				return err
			}

			dbPath := config.ProjectDBPath(root)
			if _, err := os.Stat(dbPath); os.IsNotExist(err) {
				return fmt.Errorf("Memvra not initialized. Run `memvra init` first")
			}

			database, err := db.Open(dbPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer database.Close()

			store := memory.NewStore(database)

			proj, err := store.GetProject()
			if err != nil {
				return err
			}

			ts, _ := scanner.TechStackFromJSON(proj.TechStack)

			out := &strings.Builder{}

			if section == "" || section == "profile" {
				fmt.Fprintf(out, "# Project Context: %s\n\n", proj.Name)
				fmt.Fprintf(out, "## Profile\n\n")
				fmt.Fprintf(out, "- **Language:** %s\n", ts.Language)
				if ts.Framework != "" {
					fmt.Fprintf(out, "- **Framework:** %s\n", ts.Framework)
				}
				if ts.Database != "" {
					fmt.Fprintf(out, "- **Database:** %s\n", ts.Database)
				}
				if ts.Architecture != "" {
					fmt.Fprintf(out, "- **Architecture:** %s\n", ts.Architecture)
				}
				if ts.TestFramework != "" {
					fmt.Fprintf(out, "- **Tests:** %s\n", ts.TestFramework)
				}
				if ts.CI != "" {
					fmt.Fprintf(out, "- **CI:** %s\n", ts.CI)
				}
				if len(ts.DetectedPatterns) > 0 {
					fmt.Fprintf(out, "- **Patterns:** %s\n", strings.Join(ts.DetectedPatterns, ", "))
				}
				fmt.Fprintf(out, "- **Files indexed:** %d (%d chunks)\n", proj.FileCount, proj.ChunkCount)
				fmt.Fprintf(out, "- **Last updated:** %s\n\n", proj.UpdatedAt.Format(time.RFC3339))
			}

			// Memory sections.
			typeSections := []struct {
				label string
				mt    memory.MemoryType
			}{
				{"Decisions", memory.TypeDecision},
				{"Conventions", memory.TypeConvention},
				{"Constraints", memory.TypeConstraint},
				{"Notes", memory.TypeNote},
				{"TODOs", memory.TypeTodo},
			}

			for _, ts := range typeSections {
				key := strings.ToLower(ts.label)
				if section != "" && section != key && section != "profile" {
					// Only filter when a non-profile section is requested.
					if section != key {
						continue
					}
				}
				if section == "profile" {
					continue
				}

				memories, err := store.ListMemories(ts.mt)
				if err != nil {
					continue
				}
				if len(memories) == 0 {
					continue
				}

				fmt.Fprintf(out, "## %s\n\n", ts.label)
				for _, m := range memories {
					fmt.Fprintf(out, "- %s\n", m.Content)
				}
				fmt.Fprintln(out)
			}

			fmt.Print(out.String())
			if export {
				// Write to context.md as well.
				ctxPath := config.ProjectConfigDirPath(root) + "/context.md"
				if err := os.WriteFile(ctxPath, []byte(out.String()), 0o644); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not write context.md: %v\n", err)
				} else {
					fmt.Fprintf(os.Stderr, "Context exported to %s\n", ctxPath)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&section, "section", "s", "", "Show only a specific section: profile, decisions, conventions, constraints, notes, todos")
	cmd.Flags().BoolVar(&export, "export", false, "Also write context to .memvra/context.md")

	return cmd
}
