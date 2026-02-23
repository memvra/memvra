package cli

import (
	"fmt"
	"os"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"github.com/memvra/memvra/internal/config"
	"github.com/memvra/memvra/internal/db"
	"github.com/memvra/memvra/internal/memory"
	"github.com/memvra/memvra/internal/scanner"
)

func newUpdateCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Re-scan the project and update the index incrementally",
		Long: `Detect changed files since the last scan and re-index only those files.
Use --force to re-index everything regardless of content hash.`,
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

			gcfg, _ := config.LoadGlobal()

			bar := progressbar.NewOptions(-1,
				progressbar.OptionSetDescription("  Scanning"),
				progressbar.OptionSpinnerType(14),
				progressbar.OptionSetWriter(os.Stderr),
				progressbar.OptionClearOnFinish(),
			)

			result := scanner.Scan(scanner.ScanOptions{
				Root:          root,
				MaxChunkLines: gcfg.Context.ChunkMaxLines,
			})
			_ = bar.Finish()

			var modified, added, skipped int

			for _, sf := range result.Files {
				existing, err := store.GetFileByPath(sf.File.Path)

				if err != nil || force {
					// New file or forced re-index.
					fileID, insertErr := store.UpsertFile(sf.File)
					if insertErr != nil {
						fmt.Fprintf(os.Stderr, "  Warning: could not index %s: %v\n", sf.File.Path, insertErr)
						continue
					}
					_ = store.DeleteChunksByFileID(fileID)
					for _, chunk := range sf.Chunks {
						chunk.FileID = fileID
						_ = store.InsertChunk(chunk)
					}
					if err != nil {
						added++
					} else {
						modified++ // forced re-index of existing file
					}
					continue
				}

				// Check if content changed.
				if existing.ContentHash == sf.File.ContentHash {
					skipped++
					continue
				}

				// File was modified.
				modified++
				fileID, err := store.UpsertFile(sf.File)
				if err != nil {
					continue
				}
				_ = store.DeleteChunksByFileID(fileID)
				for _, chunk := range sf.Chunks {
					chunk.FileID = fileID
					_ = store.InsertChunk(chunk)
				}
			}

			fileCount, _ := store.CountFiles()
			chunkCount, _ := store.CountChunks()

			// Update project counts.
			proj, err := store.GetProject()
			if err == nil {
				proj.FileCount = fileCount
				proj.ChunkCount = chunkCount
				_ = store.UpsertProject(proj)
			}

			fmt.Printf("Modified: %d files\n", modified)
			fmt.Printf("Added:    %d files\n", added)
			fmt.Printf("Skipped:  %d files (unchanged)\n", skipped)
			fmt.Printf("Total:    %d files, %d chunks\n", fileCount, chunkCount)

			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "re-index all files, ignoring content hashes")

	return cmd
}
