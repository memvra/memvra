-- Project metadata
CREATE TABLE IF NOT EXISTS project (
    id           TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    name         TEXT NOT NULL,
    root_path    TEXT NOT NULL,
    tech_stack   TEXT NOT NULL DEFAULT '{}',   -- JSON
    architecture TEXT,                          -- JSON: detected patterns, entry points
    conventions  TEXT,                          -- JSON: user-defined coding conventions
    file_count   INTEGER DEFAULT 0,
    chunk_count  INTEGER DEFAULT 0,
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Source file index
CREATE TABLE IF NOT EXISTS files (
    id            TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    path          TEXT NOT NULL UNIQUE,         -- Relative path from project root
    language      TEXT,
    last_modified DATETIME,
    content_hash  TEXT,                         -- SHA256 for change detection
    indexed_at    DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Content chunks with embeddings
CREATE TABLE IF NOT EXISTS chunks (
    id         TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    file_id    TEXT NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    content    TEXT NOT NULL,
    start_line INTEGER,
    end_line   INTEGER,
    chunk_type TEXT DEFAULT 'code',             -- code, comment, config, test, docs
    embedding  BLOB,                            -- Vector stored as blob for sqlite-vec
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Persistent memories (decisions, conventions, constraints)
CREATE TABLE IF NOT EXISTS memories (
    id            TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    content       TEXT NOT NULL,
    memory_type   TEXT NOT NULL,                -- decision, convention, constraint, note, todo
    importance    REAL DEFAULT 0.5,             -- 0.0 to 1.0, affects retrieval priority
    embedding     BLOB,
    source        TEXT,                         -- 'user' (manual) or 'extracted' (from session)
    related_files TEXT,                         -- JSON array of file paths
    created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Session history
CREATE TABLE IF NOT EXISTS sessions (
    id               TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    question         TEXT NOT NULL,
    context_used     TEXT,                      -- JSON: which chunks/memories were injected
    response_summary TEXT,                      -- Brief summary of the AI response
    model_used       TEXT,
    tokens_used      INTEGER,
    created_at       DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Virtual table for vector similarity search (sqlite-vec)
-- NOTE: These are created conditionally in Go code after the extension loads.

-- Indexes
CREATE INDEX IF NOT EXISTS idx_memories_type   ON memories(memory_type);
CREATE INDEX IF NOT EXISTS idx_chunks_file      ON chunks(file_id);
CREATE INDEX IF NOT EXISTS idx_sessions_created ON sessions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_files_path       ON files(path);
