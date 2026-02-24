package db

import (
	"path/filepath"
	"testing"
)

func TestOpen_CreatesDatabase(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer database.Close()

	if err := database.Ping(); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestOpen_CreatesParentDirs(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "subdir", "nested", "test.db")
	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer database.Close()
}

func TestOpen_TablesExist(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer database.Close()

	tables := []string{"project", "files", "chunks", "memories", "sessions", "schema_migrations"}
	for _, table := range tables {
		var count int
		err := database.Conn().QueryRow(
			`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, table,
		).Scan(&count)
		if err != nil {
			t.Fatalf("query table %q: %v", table, err)
		}
		if count != 1 {
			t.Errorf("table %q not found", table)
		}
	}
}

func TestOpen_MigrationsRecorded(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer database.Close()

	var count int
	err = database.Conn().QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count)
	if err != nil {
		t.Fatalf("query migrations: %v", err)
	}
	if count != len(migrations) {
		t.Errorf("expected %d migrations recorded, got %d", len(migrations), count)
	}
}

func TestOpen_Idempotent(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	// Open twice — should not fail on re-open.
	db1, err := Open(dbPath)
	if err != nil {
		t.Fatalf("first Open: %v", err)
	}
	db1.Close()

	db2, err := Open(dbPath)
	if err != nil {
		t.Fatalf("second Open: %v", err)
	}
	defer db2.Close()

	// Verify tables still exist.
	var count int
	db2.Conn().QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='project'`).Scan(&count)
	if count != 1 {
		t.Error("project table missing after re-open")
	}
}

func TestOpen_VectorTables(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer database.Close()

	// sqlite-vec tables might not exist if the extension isn't loaded,
	// but we should at least not crash.
	for _, table := range []string{"vec_chunks", "vec_memories"} {
		var count int
		database.Conn().QueryRow(
			`SELECT COUNT(*) FROM sqlite_master WHERE name=?`, table,
		).Scan(&count)
		// Just log — don't fail if vec extension is unavailable.
		t.Logf("vector table %q exists: %v", table, count > 0)
	}
}

func TestConn_ReturnsNonNil(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer database.Close()

	if database.Conn() == nil {
		t.Error("Conn() returned nil")
	}
}

func TestClose(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	if err := database.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// After close, Ping should fail.
	if err := database.Ping(); err == nil {
		t.Error("expected Ping to fail after Close")
	}
}
