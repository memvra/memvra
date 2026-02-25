package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRemoveManagedBlock_FullHook(t *testing.T) {
	// When our hook is the entire file, removing the block should leave only the shebang.
	result := removeManagedBlock(hookScript)
	result = strings.TrimSpace(result)
	if result != "#!/bin/sh" {
		t.Errorf("expected only shebang remaining, got:\n%s", result)
	}
}

func TestRemoveManagedBlock_PreservesOtherHooks(t *testing.T) {
	content := "#!/bin/sh\necho 'pre-existing hook'\n" + hookMarker + "\n" +
		"# Auto-update Memvra index after each commit.\n" +
		"if command -v memvra >/dev/null 2>&1; then\n" +
		"  memvra update --quiet 2>/dev/null &\n" +
		"fi\n" +
		"echo 'post-hook work'\n"

	result := removeManagedBlock(content)
	if strings.Contains(result, hookMarker) {
		t.Error("managed block marker should be removed")
	}
	if strings.Contains(result, "memvra update") {
		t.Error("memvra update line should be removed")
	}
	if !strings.Contains(result, "pre-existing hook") {
		t.Error("pre-existing hook should be preserved")
	}
	if !strings.Contains(result, "post-hook work") {
		t.Error("lines after the managed block should be preserved")
	}
}

func TestRemoveManagedBlock_NoBlock(t *testing.T) {
	content := "#!/bin/sh\necho 'some other hook'\n"
	result := removeManagedBlock(content)
	if result != content {
		t.Errorf("content should be unchanged when no managed block exists")
	}
}

func TestHookInstallUninstall(t *testing.T) {
	dir := t.TempDir()

	// Create a fake .git/hooks directory.
	hooksDir := filepath.Join(dir, ".git", "hooks")
	os.MkdirAll(hooksDir, 0o755)

	hookPath := filepath.Join(hooksDir, "post-commit")

	// Install: write fresh hook.
	if err := os.WriteFile(hookPath, []byte(hookScript), 0o755); err != nil {
		t.Fatalf("write hook: %v", err)
	}

	// Verify the hook file exists and contains our marker.
	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("read hook: %v", err)
	}
	if !strings.Contains(string(data), hookMarker) {
		t.Error("hook should contain marker after install")
	}
	if !strings.Contains(string(data), "memvra update") {
		t.Error("hook should contain 'memvra update'")
	}

	// Verify file is executable.
	info, _ := os.Stat(hookPath)
	if info.Mode()&0o100 == 0 {
		t.Error("hook should be executable")
	}

	// Uninstall: remove managed block.
	content := string(data)
	cleaned := strings.TrimSpace(removeManagedBlock(content))
	if cleaned == "" || cleaned == "#!/bin/sh" {
		// Nothing left — remove the file.
		os.Remove(hookPath)
	}

	if _, err := os.Stat(hookPath); !os.IsNotExist(err) {
		t.Error("hook file should be removed when only memvra content existed")
	}
}

func TestHookAppendToExisting(t *testing.T) {
	dir := t.TempDir()
	hooksDir := filepath.Join(dir, ".git", "hooks")
	os.MkdirAll(hooksDir, 0o755)

	hookPath := filepath.Join(hooksDir, "post-commit")

	// Write a pre-existing hook.
	existing := "#!/bin/sh\necho 'lint check'\n"
	os.WriteFile(hookPath, []byte(existing), 0o755)

	// Simulate append (same logic as hook install when file already exists).
	data, _ := os.ReadFile(hookPath)
	content := string(data)
	if !strings.Contains(content, hookMarker) {
		appended := content + "\n" + hookMarker + "\n" +
			"# Auto-update Memvra index after each commit.\n" +
			"if command -v memvra >/dev/null 2>&1; then\n" +
			"  memvra update --quiet 2>/dev/null &\n" +
			"fi\n"
		os.WriteFile(hookPath, []byte(appended), 0o755)
	}

	// Verify both sections exist.
	final, _ := os.ReadFile(hookPath)
	finalStr := string(final)
	if !strings.Contains(finalStr, "lint check") {
		t.Error("original hook content should be preserved")
	}
	if !strings.Contains(finalStr, hookMarker) {
		t.Error("memvra marker should be present")
	}

	// Uninstall — should preserve the original hook.
	cleaned := strings.TrimSpace(removeManagedBlock(finalStr))
	os.WriteFile(hookPath, []byte(cleaned+"\n"), 0o755)

	restored, _ := os.ReadFile(hookPath)
	restoredStr := string(restored)
	if !strings.Contains(restoredStr, "lint check") {
		t.Error("original hook content should survive uninstall")
	}
	if strings.Contains(restoredStr, hookMarker) {
		t.Error("memvra marker should be gone after uninstall")
	}
}
