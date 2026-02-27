package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/memvra/memvra/internal/config"
	"github.com/memvra/memvra/internal/db"
	"github.com/memvra/memvra/internal/memory"
	"github.com/memvra/memvra/internal/scanner"
)

func newPruneCmd() *cobra.Command {
	var (
		olderThanDays int
		keepLatest    int
		dryRun        bool
	)

	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Remove old sessions to reduce database size",
		Long: `Prune old session records from the .memvra database.

By default, keeps the latest 100 sessions. Use flags to customise:

  memvra prune                    # keep latest 100 sessions
  memvra prune --older-than 30    # delete sessions older than 30 days
  memvra prune --keep 50          # keep only the latest 50 sessions
  memvra prune --dry-run          # preview what would be deleted`,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := scanner.FindProjectRoot(".")
			if err != nil {
				return fmt.Errorf("find project root: %w", err)
			}

			dbPath := config.ProjectDBPath(root)
			database, err := db.Open(dbPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer func() { _ = database.Close() }()

			store := memory.NewStore(database)

			before, _ := store.CountSessions()

			if dryRun {
				fmt.Printf("Current sessions: %d\n", before)
				if olderThanDays > 0 {
					fmt.Printf("Would delete sessions older than %d days\n", olderThanDays)
				} else {
					fmt.Printf("Would keep latest %d sessions\n", keepLatest)
				}
				return nil
			}

			var pruned int
			if olderThanDays > 0 {
				pruned, err = store.PruneSessions(olderThanDays)
			} else {
				pruned, err = store.PruneSessionsKeepLatest(keepLatest)
			}
			if err != nil {
				return err
			}

			after, _ := store.CountSessions()
			fmt.Printf("Pruned %d sessions (%d → %d)\n", pruned, before, after)
			return nil
		},
	}

	cmd.Flags().IntVar(&olderThanDays, "older-than", 0, "Delete sessions older than N days")
	cmd.Flags().IntVar(&keepLatest, "keep", 100, "Keep only the latest N sessions")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview what would be pruned without deleting")

	return cmd
}
