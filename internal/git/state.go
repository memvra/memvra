// Package git captures the working state of a git repository
// for inclusion in exported context files.
package git

import (
	"os/exec"
	"strings"
)

// WorkingState captures the current git status and diff summary.
type WorkingState struct {
	Branch    string
	Modified  []string // unstaged changes
	Staged    []string // staged changes
	Untracked []string // new files
	DiffStat  string   // git diff --stat output
}

// IsEmpty returns true if there is no git state to report.
func (ws WorkingState) IsEmpty() bool {
	return ws.Branch == "" && len(ws.Modified) == 0 &&
		len(ws.Staged) == 0 && len(ws.Untracked) == 0
}

// HasChanges returns true if there are any modified, staged, or untracked files.
func (ws WorkingState) HasChanges() bool {
	return len(ws.Modified) > 0 || len(ws.Staged) > 0 || len(ws.Untracked) > 0
}

// ChangedFiles returns all files with any kind of change (staged + modified, deduplicated).
func (ws WorkingState) ChangedFiles() []string {
	seen := make(map[string]bool)
	var files []string
	for _, f := range ws.Staged {
		if !seen[f] {
			seen[f] = true
			files = append(files, f)
		}
	}
	for _, f := range ws.Modified {
		if !seen[f] {
			seen[f] = true
			files = append(files, f)
		}
	}
	return files
}

// CaptureWorkingState runs git commands in the given directory and returns
// the current working state. All errors are swallowed — if git is not
// installed or the directory is not a repo, an empty WorkingState is returned.
func CaptureWorkingState(dir string) WorkingState {
	var ws WorkingState

	ws.Branch = gitOutput(dir, "rev-parse", "--abbrev-ref", "HEAD")

	porcelain := gitOutput(dir, "status", "--porcelain")
	if porcelain != "" {
		for _, line := range strings.Split(porcelain, "\n") {
			if len(line) < 3 {
				continue
			}
			x, y := line[0], line[1]
			path := strings.TrimSpace(line[2:])
			if path == "" {
				continue
			}

			if x == '?' && y == '?' {
				ws.Untracked = append(ws.Untracked, path)
			} else {
				if x == 'A' || x == 'M' || x == 'D' || x == 'R' || x == 'C' {
					ws.Staged = append(ws.Staged, path)
				}
				if y == 'M' || y == 'D' {
					ws.Modified = append(ws.Modified, path)
				}
			}
		}
	}

	ws.DiffStat = gitOutput(dir, "diff", "--stat")

	return ws
}

// gitOutput runs a git command and returns trimmed stdout.
// Returns "" on any error.
func gitOutput(dir string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
