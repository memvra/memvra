package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/memvra/memvra/internal/config"
	"github.com/memvra/memvra/internal/db"
	"github.com/memvra/memvra/internal/memory"
	"github.com/memvra/memvra/internal/scanner"
)

// ANSI color helpers.
var (
	cReset  = "\033[0m"
	cGreen  = "\033[32m"
	cYellow = "\033[33m"
	cRed    = "\033[31m"
	cCyan   = "\033[36m"
	cBold   = "\033[1m"
	cDim    = "\033[2m"
)

func disableColors() {
	cReset, cGreen, cYellow, cRed, cCyan, cBold, cDim = "", "", "", "", "", "", ""
}

func newDiffCmd() *cobra.Command {
	var (
		filesOnly    bool
		memoriesOnly bool
		sessionsOnly bool
		since        string
		noScan       bool
	)

	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Show changes since the last update",
		Long: `Compare the current project state against the Memvra index.

Shows three sections:
  - File index changes (added, modified, deleted files)
  - New memories since the last update
  - New sessions since the last update

Examples:
  memvra diff
  memvra diff --files-only
  memvra diff --since 24h
  memvra diff --no-scan`,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := findRoot()
			if err != nil {
				return err
			}

			dbPath, err := ensureInitialized(root)
			if err != nil {
				return err
			}

			database, err := db.Open(dbPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer database.Close()

			store := memory.NewStore(database)
			gcfg, _ := config.LoadGlobal()

			if !gcfg.Output.Color || os.Getenv("NO_COLOR") != "" {
				disableColors()
			}

			proj, err := store.GetProject()
			if err != nil {
				return err
			}

			anchor := proj.UpdatedAt
			if since != "" {
				dur, err := parseDuration(since)
				if err != nil {
					return fmt.Errorf("invalid --since value %q: %w", since, err)
				}
				anchor = time.Now().Add(-dur)
			}

			showAll := !filesOnly && !memoriesOnly && !sessionsOnly
			showFiles := showAll || filesOnly
			showMemories := showAll || memoriesOnly
			showSessions := showAll || sessionsOnly

			if noScan && !filesOnly {
				showFiles = false
			}

			if showFiles {
				result := scanner.Scan(scanner.ScanOptions{
					Root:          root,
					MaxChunkLines: gcfg.Context.ChunkMaxLines,
				})

				allDBFiles, err := store.ListFiles()
				if err != nil {
					return fmt.Errorf("list indexed files: %w", err)
				}

				dbFilesByPath := make(map[string]memory.File, len(allDBFiles))
				for _, f := range allDBFiles {
					dbFilesByPath[f.Path] = f
				}

				scannedPaths := make(map[string]struct{}, len(result.Files))
				for _, sf := range result.Files {
					scannedPaths[sf.File.Path] = struct{}{}
				}

				var added, modified []string
				for _, sf := range result.Files {
					dbFile, exists := dbFilesByPath[sf.File.Path]
					if !exists {
						added = append(added, sf.File.Path)
					} else if dbFile.ContentHash != sf.File.ContentHash {
						modified = append(modified, sf.File.Path)
					}
				}

				var deleted []string
				for _, dbFile := range allDBFiles {
					if _, found := scannedPaths[dbFile.Path]; !found {
						deleted = append(deleted, dbFile.Path)
					}
				}

				printFileDiff(added, modified, deleted)
			}

			if showMemories {
				memories, err := store.ListMemoriesSince(anchor)
				if err != nil {
					return fmt.Errorf("list recent memories: %w", err)
				}
				printMemoryDiff(memories, anchor)
			}

			if showSessions {
				sessions, err := store.ListSessionsSince(anchor)
				if err != nil {
					return fmt.Errorf("list recent sessions: %w", err)
				}
				printSessionDiff(sessions, anchor)
			}

			fmt.Println()
			return nil
		},
	}

	cmd.Flags().BoolVar(&filesOnly, "files-only", false, "only show file index changes")
	cmd.Flags().BoolVar(&memoriesOnly, "memories-only", false, "only show memory changes")
	cmd.Flags().BoolVar(&sessionsOnly, "sessions-only", false, "only show session changes")
	cmd.Flags().StringVar(&since, "since", "", "override time anchor (e.g. 24h, 7d, 2h30m)")
	cmd.Flags().BoolVar(&noScan, "no-scan", false, "skip filesystem scan (show only memory/session changes)")

	return cmd
}

func printFileDiff(added, modified, deleted []string) {
	total := len(added) + len(modified) + len(deleted)
	fmt.Printf("\n%s=== File Index ===%s\n", cBold, cReset)

	if total == 0 {
		fmt.Printf("  %s(no changes)%s\n", cDim, cReset)
		return
	}

	sort.Strings(added)
	sort.Strings(modified)
	sort.Strings(deleted)

	for _, p := range added {
		fmt.Printf("  %s+ %s%s\n", cGreen, p, cReset)
	}
	for _, p := range modified {
		fmt.Printf("  %s~ %s%s\n", cYellow, p, cReset)
	}
	for _, p := range deleted {
		fmt.Printf("  %s- %s%s\n", cRed, p, cReset)
	}

	fmt.Printf("\n  %d added, %d modified, %d deleted\n", len(added), len(modified), len(deleted))
}

func printMemoryDiff(memories []memory.Memory, since time.Time) {
	fmt.Printf("\n%s=== Memories (since %s) ===%s\n", cBold, since.Format("2006-01-02 15:04"), cReset)

	if len(memories) == 0 {
		fmt.Printf("  %s(none)%s\n", cDim, cReset)
		return
	}

	grouped := make(map[memory.MemoryType][]memory.Memory)
	typeOrder := []memory.MemoryType{
		memory.TypeDecision, memory.TypeConvention,
		memory.TypeConstraint, memory.TypeNote, memory.TypeTodo,
	}
	for _, m := range memories {
		grouped[m.MemoryType] = append(grouped[m.MemoryType], m)
	}

	for _, t := range typeOrder {
		mems, ok := grouped[t]
		if !ok || len(mems) == 0 {
			continue
		}
		fmt.Printf("  %s%s%s (%d)\n", cCyan, capitalize(string(t))+"s", cReset, len(mems))
		for _, m := range mems {
			preview := m.Content
			if len(preview) > 80 {
				preview = preview[:77] + "..."
			}
			fmt.Printf("    %s+ %s%s\n", cGreen, preview, cReset)
		}
	}
}

func printSessionDiff(sessions []memory.Session, since time.Time) {
	fmt.Printf("\n%s=== Sessions (since %s) ===%s\n", cBold, since.Format("2006-01-02 15:04"), cReset)

	if len(sessions) == 0 {
		fmt.Printf("  %s(none)%s\n", cDim, cReset)
		return
	}

	for _, sess := range sessions {
		ts := sess.CreatedAt.Format("Jan 02 15:04")
		question := sess.Question
		if len(question) > 70 {
			question = question[:67] + "..."
		}
		fmt.Printf("  %s[%s]%s %s\n", cDim, ts, cReset, question)

		if sess.ResponseSummary != "" {
			summary := sess.ResponseSummary
			if len(summary) > 100 {
				summary = summary[:97] + "..."
			}
			fmt.Printf("    %s%s%s\n", cDim, summary, cReset)
		}
	}

	fmt.Printf("\n  %d session%s\n", len(sessions), pluralS(len(sessions)))
}

// parseDuration extends time.ParseDuration to support "d" (day) units.
func parseDuration(s string) (time.Duration, error) {
	if d, err := time.ParseDuration(s); err == nil {
		return d, nil
	}
	if strings.HasSuffix(s, "d") {
		numStr := strings.TrimSuffix(s, "d")
		var n int
		if _, err := fmt.Sscan(numStr, &n); err != nil {
			return 0, fmt.Errorf("invalid day count %q", numStr)
		}
		return time.Duration(n) * 24 * time.Hour, nil
	}
	return 0, fmt.Errorf("unrecognised duration format %q", s)
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func pluralS(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
