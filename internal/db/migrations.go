package db

import (
	"database/sql"
	"fmt"
)

// migrations is an ordered list of SQL migration statements.
// Each entry is applied once in order. New migrations are appended at the end.
var migrations = []string{
	// Migration 0: initial schema
	`CREATE TABLE IF NOT EXISTS project (
		id           TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
		name         TEXT NOT NULL,
		root_path    TEXT NOT NULL,
		tech_stack   TEXT NOT NULL DEFAULT '{}',
		architecture TEXT,
		conventions  TEXT,
		file_count   INTEGER DEFAULT 0,
		chunk_count  INTEGER DEFAULT 0,
		created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,

	`CREATE TABLE IF NOT EXISTS files (
		id            TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
		path          TEXT NOT NULL UNIQUE,
		language      TEXT,
		last_modified DATETIME,
		content_hash  TEXT,
		indexed_at    DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,

	`CREATE TABLE IF NOT EXISTS chunks (
		id         TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
		file_id    TEXT NOT NULL REFERENCES files(id) ON DELETE CASCADE,
		content    TEXT NOT NULL,
		start_line INTEGER,
		end_line   INTEGER,
		chunk_type TEXT DEFAULT 'code',
		embedding  BLOB,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,

	`CREATE TABLE IF NOT EXISTS memories (
		id            TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
		content       TEXT NOT NULL,
		memory_type   TEXT NOT NULL,
		importance    REAL DEFAULT 0.5,
		embedding     BLOB,
		source        TEXT,
		related_files TEXT,
		created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,

	`CREATE TABLE IF NOT EXISTS sessions (
		id               TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
		question         TEXT NOT NULL,
		context_used     TEXT,
		response_summary TEXT,
		model_used       TEXT,
		tokens_used      INTEGER,
		created_at       DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,

	`CREATE INDEX IF NOT EXISTS idx_memories_type   ON memories(memory_type)`,
	`CREATE INDEX IF NOT EXISTS idx_chunks_file      ON chunks(file_id)`,
	`CREATE INDEX IF NOT EXISTS idx_sessions_created ON sessions(created_at DESC)`,
	`CREATE INDEX IF NOT EXISTS idx_files_path       ON files(path)`,

	// Migration 1: migration tracking table
	`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    INTEGER PRIMARY KEY,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,
}

// applyMigrations runs any migrations that have not yet been applied.
func applyMigrations(conn *sql.DB) error {
	// Ensure the migration tracking table exists first.
	if _, err := conn.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    INTEGER PRIMARY KEY,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	for i, stmt := range migrations {
		var count int
		row := conn.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, i)
		if err := row.Scan(&count); err != nil {
			return fmt.Errorf("check migration %d: %w", i, err)
		}
		if count > 0 {
			continue
		}

		if _, err := conn.Exec(stmt); err != nil {
			return fmt.Errorf("apply migration %d: %w", i, err)
		}

		if _, err := conn.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, i); err != nil {
			return fmt.Errorf("record migration %d: %w", i, err)
		}
	}

	return nil
}

// applyVectorTables creates the sqlite-vec virtual tables.
// Called separately after the vec extension is confirmed loaded.
func applyVectorTables(conn *sql.DB, dimension int) error {
	stmts := []string{
		fmt.Sprintf(`CREATE VIRTUAL TABLE IF NOT EXISTS vec_chunks USING vec0(
			id TEXT PRIMARY KEY,
			embedding float[%d]
		)`, dimension),
		fmt.Sprintf(`CREATE VIRTUAL TABLE IF NOT EXISTS vec_memories USING vec0(
			id TEXT PRIMARY KEY,
			embedding float[%d]
		)`, dimension),
	}

	for _, stmt := range stmts {
		if _, err := conn.Exec(stmt); err != nil {
			return fmt.Errorf("create vector table: %w", err)
		}
	}

	return nil
}
