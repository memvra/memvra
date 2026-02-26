package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestIsEmpty_ZeroValue(t *testing.T) {
	var ws WorkingState
	if !ws.IsEmpty() {
		t.Error("zero-value WorkingState should be empty")
	}
}

func TestIsEmpty_BranchOnly(t *testing.T) {
	ws := WorkingState{Branch: "main"}
	if ws.IsEmpty() {
		t.Error("WorkingState with branch should not be empty")
	}
}

func TestIsEmpty_WithFiles(t *testing.T) {
	ws := WorkingState{Modified: []string{"file.go"}}
	if ws.IsEmpty() {
		t.Error("WorkingState with modified files should not be empty")
	}
}

func TestHasChanges(t *testing.T) {
	tests := []struct {
		name string
		ws   WorkingState
		want bool
	}{
		{"empty", WorkingState{}, false},
		{"branch only", WorkingState{Branch: "main"}, false},
		{"modified", WorkingState{Modified: []string{"f.go"}}, true},
		{"staged", WorkingState{Staged: []string{"f.go"}}, true},
		{"untracked", WorkingState{Untracked: []string{"f.go"}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ws.HasChanges(); got != tt.want {
				t.Errorf("HasChanges() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChangedFiles_Deduplicates(t *testing.T) {
	ws := WorkingState{
		Staged:   []string{"a.go", "b.go"},
		Modified: []string{"b.go", "c.go"},
	}
	files := ws.ChangedFiles()
	if len(files) != 3 {
		t.Errorf("expected 3 unique files, got %d: %v", len(files), files)
	}
}

func TestCaptureWorkingState_NonGitDir(t *testing.T) {
	dir := t.TempDir()
	ws := CaptureWorkingState(dir)
	if !ws.IsEmpty() {
		t.Errorf("expected empty state for non-git dir, got: %+v", ws)
	}
}

func TestCaptureWorkingState_CleanRepo(t *testing.T) {
	dir := initTestRepo(t)

	ws := CaptureWorkingState(dir)
	if ws.Branch == "" {
		t.Error("expected branch name in git repo")
	}
	if ws.HasChanges() {
		t.Errorf("expected clean state, got: %+v", ws)
	}
}

func TestCaptureWorkingState_UntrackedFile(t *testing.T) {
	dir := initTestRepo(t)

	os.WriteFile(filepath.Join(dir, "newfile.go"), []byte("package main"), 0o644)

	ws := CaptureWorkingState(dir)
	if len(ws.Untracked) != 1 || ws.Untracked[0] != "newfile.go" {
		t.Errorf("expected 1 untracked file 'newfile.go', got: %v", ws.Untracked)
	}
}

func TestCaptureWorkingState_StagedFile(t *testing.T) {
	dir := initTestRepo(t)

	os.WriteFile(filepath.Join(dir, "staged.go"), []byte("package main"), 0o644)
	gitCmd(t, dir, "add", "staged.go")

	ws := CaptureWorkingState(dir)
	if len(ws.Staged) != 1 || ws.Staged[0] != "staged.go" {
		t.Errorf("expected 1 staged file 'staged.go', got: %v", ws.Staged)
	}
}

func TestCaptureWorkingState_ChangedFile(t *testing.T) {
	dir := initTestRepo(t)

	// Create, add, commit a file.
	fpath := filepath.Join(dir, "tracked.go")
	os.WriteFile(fpath, []byte("package main"), 0o644)
	gitCmd(t, dir, "add", "tracked.go")
	gitCmd(t, dir, "commit", "-m", "add tracked")

	// Modify it. Due to git's stat cache, fast modifications may be reported
	// as either staged (M in X) or working-tree-modified (M in Y). Both mean
	// the file has changes — use ChangedFiles() which merges both.
	os.WriteFile(fpath, []byte("package main\n// changed\n// extra line"), 0o644)

	ws := CaptureWorkingState(dir)

	changed := ws.ChangedFiles()
	if len(changed) != 1 || changed[0] != "tracked.go" {
		t.Errorf("expected 1 changed file 'tracked.go', got: %v", changed)
	}
	if ws.DiffStat == "" {
		t.Error("expected non-empty DiffStat for changed file")
	}
}

func TestCaptureWorkingState_MixedState(t *testing.T) {
	dir := initTestRepo(t)

	// Create and commit a file.
	os.WriteFile(filepath.Join(dir, "base.go"), []byte("package main"), 0o644)
	gitCmd(t, dir, "add", "base.go")
	gitCmd(t, dir, "commit", "-m", "init")

	// Modify committed file.
	os.WriteFile(filepath.Join(dir, "base.go"), []byte("package main\n// edited\n// extra"), 0o644)

	// Stage a new file.
	os.WriteFile(filepath.Join(dir, "new.go"), []byte("package new"), 0o644)
	gitCmd(t, dir, "add", "new.go")

	// Create an untracked file.
	os.WriteFile(filepath.Join(dir, "untracked.txt"), []byte("hello"), 0o644)

	ws := CaptureWorkingState(dir)

	// base.go should appear as changed (staged or modified depending on timing).
	changed := ws.ChangedFiles()
	found := map[string]bool{}
	for _, f := range changed {
		found[f] = true
	}
	if !found["base.go"] {
		t.Errorf("expected base.go in changed files, got: %v", changed)
	}
	if !found["new.go"] {
		t.Errorf("expected new.go in changed files, got: %v", changed)
	}

	// Untracked should always be correct.
	if len(ws.Untracked) != 1 || ws.Untracked[0] != "untracked.txt" {
		t.Errorf("expected 1 untracked 'untracked.txt', got: %v", ws.Untracked)
	}

	if !ws.HasChanges() {
		t.Error("expected HasChanges() to be true")
	}
}

// initTestRepo creates a temp dir with a git repo and an initial commit.
func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitCmd(t, dir, "init")
	gitCmd(t, dir, "config", "user.email", "test@test.com")
	gitCmd(t, dir, "config", "user.name", "Test")

	// Need at least one commit for branch to exist.
	os.WriteFile(filepath.Join(dir, ".gitkeep"), []byte(""), 0o644)
	gitCmd(t, dir, "add", ".gitkeep")
	gitCmd(t, dir, "commit", "-m", "initial")

	return dir
}

func gitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}
