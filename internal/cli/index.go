package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/memvra/memvra/internal/config"
	"github.com/memvra/memvra/internal/memory"
	"github.com/memvra/memvra/internal/scanner"
)

// fileStatus describes what happened when a file was upserted.
type fileStatus int

const (
	fileUnchanged fileStatus = iota
	fileAdded
	fileModified
)

// upsertScannedFile indexes a single scanned file: upserts the file record,
// replaces its chunks, and returns the new file ID. force causes re-indexing
// even if the content hash matches.
func upsertScannedFile(store *memory.Store, sf scanner.ScannedFile, force bool) (fileID string, status fileStatus, err error) {
	existing, lookupErr := store.GetFileByPath(sf.File.Path)

	isNew := lookupErr != nil

	if isNew || force {
		fileID, err = store.UpsertFile(sf.File)
		if err != nil {
			return "", fileUnchanged, fmt.Errorf("upsert %s: %w", sf.File.Path, err)
		}
		_ = store.DeleteChunksByFileID(fileID)
		for _, chunk := range sf.Chunks {
			chunk.FileID = fileID
			_ = store.InsertChunk(chunk)
		}
		if isNew {
			return fileID, fileAdded, nil
		}
		return fileID, fileModified, nil
	}

	// Existing file — check hash.
	if existing.ContentHash == sf.File.ContentHash {
		return existing.ID, fileUnchanged, nil
	}

	// Content changed — re-chunk.
	fileID, err = store.UpsertFile(sf.File)
	if err != nil {
		return "", fileUnchanged, fmt.Errorf("upsert %s: %w", sf.File.Path, err)
	}
	_ = store.DeleteChunksByFileID(fileID)
	for _, chunk := range sf.Chunks {
		chunk.FileID = fileID
		_ = store.InsertChunk(chunk)
	}
	return fileID, fileModified, nil
}

// pruneDeletedFile removes a file and its vector embeddings from the store.
func pruneDeletedFile(store *memory.Store, vectors *memory.VectorStore, fileID string) {
	chunks, _ := store.ListChunksByFileID(fileID)
	for _, c := range chunks {
		_ = vectors.DeleteChunkEmbedding(c.ID)
	}
	_ = store.DeleteFile(fileID)
}

// embedFileChunks generates embeddings for all chunks of the given file IDs.
// Returns the count of chunks successfully embedded.
func embedFileChunks(ctx context.Context, store *memory.Store, vectors *memory.VectorStore, embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}, fileIDs []string) int {
	embeddedCount := 0
	const batchSize = 32

	for _, fileID := range fileIDs {
		chunks, err := store.ListChunksByFileID(fileID)
		if err != nil || len(chunks) == 0 {
			continue
		}

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

			vecs, err := embedder.Embed(ctx, texts)
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
	return embeddedCount
}

// refreshProjectCounts updates the file and chunk counts on the project record.
func refreshProjectCounts(store *memory.Store) {
	fileCount, _ := store.CountFiles()
	chunkCount, _ := store.CountChunks()
	proj, err := store.GetProject()
	if err != nil {
		return
	}
	proj.FileCount = fileCount
	proj.ChunkCount = chunkCount
	_ = store.UpsertProject(proj)
}

// ensureInitialized checks that the project has been initialized (.memvra/memvra.db exists).
func ensureInitialized(root string) (string, error) {
	dbPath := config.ProjectDBPath(root)
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return "", fmt.Errorf("Memvra not initialized. Run `memvra init` first")
	}
	return dbPath, nil
}
