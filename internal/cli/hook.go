package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// hookMarker identifies the memvra-managed section inside a hook script.
const hookMarker = "# memvra:managed"

// hookScript is the shell snippet injected into the post-commit hook.
const hookScript = `#!/bin/sh
` + hookMarker + `
# Auto-update Memvra index after each commit.
if command -v memvra >/dev/null 2>&1; then
  memvra update --quiet 2>/dev/null &
fi
`

func newHookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hook",
		Short: "Manage git hooks for automatic re-indexing",
		Long: `Install or remove a post-commit git hook that automatically runs
'memvra update' after each commit, keeping the index in sync.`,
	}

	cmd.AddCommand(
		newHookInstallCmd(),
		newHookUninstallCmd(),
		newHookStatusCmd(),
	)

	return cmd
}

func newHookInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install the post-commit hook",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := findRoot()
			if err != nil {
				return err
			}

			gitDir := filepath.Join(root, ".git")
			if _, err := os.Stat(gitDir); os.IsNotExist(err) {
				return fmt.Errorf("no .git directory found in %s", root)
			}

			hooksDir := filepath.Join(gitDir, "hooks")
			if err := os.MkdirAll(hooksDir, 0o755); err != nil {
				return fmt.Errorf("create hooks directory: %w", err)
			}

			hookPath := filepath.Join(hooksDir, "post-commit")

			// If the hook file already exists, check whether we already manage it.
			if data, err := os.ReadFile(hookPath); err == nil {
				content := string(data)
				if strings.Contains(content, hookMarker) {
					fmt.Println("Hook already installed.")
					return nil
				}

				// Existing hook from another tool — append our snippet.
				appended := content + "\n" + hookMarker + "\n" +
					"# Auto-update Memvra index after each commit.\n" +
					"if command -v memvra >/dev/null 2>&1; then\n" +
					"  memvra update --quiet 2>/dev/null &\n" +
					"fi\n"
				if err := os.WriteFile(hookPath, []byte(appended), 0o755); err != nil {
					return fmt.Errorf("append to hook: %w", err)
				}
				fmt.Println("Appended memvra hook to existing post-commit hook.")
				return nil
			}

			// No existing hook — write a fresh one.
			if err := os.WriteFile(hookPath, []byte(hookScript), 0o755); err != nil {
				return fmt.Errorf("write hook: %w", err)
			}

			fmt.Println("Installed post-commit hook. Memvra will auto-update after each commit.")
			return nil
		},
	}
}

func newHookUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Remove the post-commit hook",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := findRoot()
			if err != nil {
				return err
			}

			hookPath := filepath.Join(root, ".git", "hooks", "post-commit")

			data, err := os.ReadFile(hookPath)
			if os.IsNotExist(err) {
				fmt.Println("No post-commit hook found.")
				return nil
			}
			if err != nil {
				return fmt.Errorf("read hook: %w", err)
			}

			content := string(data)
			if !strings.Contains(content, hookMarker) {
				fmt.Println("No memvra hook found in post-commit.")
				return nil
			}

			// Remove our managed lines.
			cleaned := removeManagedBlock(content)
			cleaned = strings.TrimSpace(cleaned)

			if cleaned == "" || cleaned == "#!/bin/sh" {
				// Nothing left — remove the entire file.
				if err := os.Remove(hookPath); err != nil {
					return fmt.Errorf("remove hook: %w", err)
				}
				fmt.Println("Removed post-commit hook.")
				return nil
			}

			// Other hook content remains — write it back.
			if err := os.WriteFile(hookPath, []byte(cleaned+"\n"), 0o755); err != nil {
				return fmt.Errorf("write hook: %w", err)
			}
			fmt.Println("Removed memvra section from post-commit hook (other hooks preserved).")
			return nil
		},
	}
}

func newHookStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check if the post-commit hook is installed",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := findRoot()
			if err != nil {
				return err
			}

			hookPath := filepath.Join(root, ".git", "hooks", "post-commit")
			data, err := os.ReadFile(hookPath)
			if os.IsNotExist(err) {
				fmt.Println("Not installed.")
				return nil
			}
			if err != nil {
				return fmt.Errorf("read hook: %w", err)
			}

			if strings.Contains(string(data), hookMarker) {
				fmt.Println("Installed.")
			} else {
				fmt.Println("Not installed (post-commit hook exists but has no memvra section).")
			}
			return nil
		},
	}
}

// removeManagedBlock removes the memvra-managed lines from a hook script.
// It strips from the hookMarker line through the trailing "fi" of our block.
func removeManagedBlock(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	inBlock := false

	for _, line := range lines {
		if strings.Contains(line, hookMarker) {
			inBlock = true
			continue
		}
		if inBlock {
			// Our block ends after the "fi" line.
			if strings.TrimSpace(line) == "fi" {
				inBlock = false
				continue
			}
			continue
		}
		result = append(result, line)
	}

	return strings.Join(result, "\n")
}
