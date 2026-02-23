package cli

import (
	"context"
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
Re-generates embeddings for modified/added files and prunes deleted files.
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
			vectors := memory.NewVectorStore(database)
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

			// Track file IDs that were added or modified so we can re-embed them.
			changedFileIDs := make([]string, 0)

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
					changedFileIDs = append(changedFileIDs, fileID)
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

				// File was modified — re-chunk it.
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
				changedFileIDs = append(changedFileIDs, fileID)
			}

			// Prune files that are no longer on disk.
			var deleted int
			allDBFiles, err := store.ListFiles()
			if err == nil {
				scannedPaths := make(map[string]struct{}, len(result.Files))
				for _, sf := range result.Files {
					scannedPaths[sf.File.Path] = struct{}{}
				}
				for _, dbFile := range allDBFiles {
					if _, found := scannedPaths[dbFile.Path]; !found {
						// Remove vector embeddings for each chunk first.
						chunks, _ := store.ListChunksByFileID(dbFile.ID)
						for _, c := range chunks {
							_ = vectors.DeleteChunkEmbedding(c.ID)
						}
						_ = store.DeleteFile(dbFile.ID)
						deleted++
					}
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
			fmt.Printf("Deleted:  %d files\n", deleted)
			fmt.Printf("Skipped:  %d files (unchanged)\n", skipped)
			fmt.Printf("Total:    %d files, %d chunks\n", fileCount, chunkCount)

			// Re-embed changed/added chunks.
			if len(changedFileIDs) == 0 {
				return nil
			}
			embedder := buildEmbedder(gcfg)
			if embedder == nil {
				return nil
			}

			embBar := progressbar.NewOptions(-1,
				progressbar.OptionSetDescription("  Generating embeddings"),
				progressbar.OptionSpinnerType(14),
				progressbar.OptionSetWriter(os.Stderr),
				progressbar.OptionClearOnFinish(),
			)

			embeddedCount := 0
			for _, fileID := range changedFileIDs {
				chunks, err := store.ListChunksByFileID(fileID)
				if err != nil || len(chunks) == 0 {
					continue
				}

				const batchSize = 32
				for i := 0; i < len(chunks); i += batchSize {
					end := i + batchSize
					if end > len(chunks) {
						end = len(chunks)
					}
					batch := chunks[i:end]

					texts := make([]string, len(batch))
					for j, c := range batch {
						texts[j] = c.Content
					}

					vecs, err := embedder.Embed(context.Background(), texts)
					if err != nil {
						break
					}
					for j, vec := range vecs {
						if j >= len(batch) {
							break
						}
						if err := vectors.UpsertChunkEmbedding(batch[j].ID, vec); err == nil {
							embeddedCount++
						}
					}
				}
			}
			_ = embBar.Finish()

			if embeddedCount > 0 {
				fmt.Printf("%d chunks re-embedded\n", embeddedCount)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "re-index all files, ignoring content hashes")

	return cmd
}
