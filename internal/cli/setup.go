package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/memvra/memvra/internal/config"
)

func newSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Interactive first-time configuration",
		Long:  "Configure API keys, default LLM model, and embedding provider for Memvra.",
		RunE: func(cmd *cobra.Command, args []string) error {
			reader := bufio.NewReader(os.Stdin)

			fmt.Println("Welcome to Memvra! Let's configure your AI memory layer.")
			fmt.Println()

			cfg := config.DefaultGlobal()

			// Step 1: Choose LLM provider.
			fmt.Println("Which LLM do you primarily use?")
			fmt.Println("  [1] Claude (Anthropic)")
			fmt.Println("  [2] OpenAI (GPT-4o)")
			fmt.Println("  [3] Gemini (Google)")
			fmt.Println("  [4] Ollama (local)")
			fmt.Print("> ")

			choice := readLineBuf(reader)
			switch strings.TrimSpace(choice) {
			case "1":
				cfg.DefaultModel = "claude"
				fmt.Print("Enter your Anthropic API key (or press Enter to set ANTHROPIC_API_KEY later): ")
				if key := readLineBuf(reader); key != "" {
					cfg.Keys.Anthropic = key
				}
			case "2":
				cfg.DefaultModel = "openai"
				fmt.Print("Enter your OpenAI API key (or press Enter to set OPENAI_API_KEY later): ")
				if key := readLineBuf(reader); key != "" {
					cfg.Keys.OpenAI = key
				}
			case "3":
				cfg.DefaultModel = "gemini"
				fmt.Print("Enter your Gemini API key (or press Enter to set GEMINI_API_KEY later): ")
				if key := readLineBuf(reader); key != "" {
					cfg.Keys.Gemini = key
				}
			case "4":
				cfg.DefaultModel = "ollama"
			default:
				fmt.Println("Unrecognized choice; defaulting to claude.")
				cfg.DefaultModel = "claude"
			}

			fmt.Println()

			// Step 2: Choose embedding provider.
			fmt.Println("For embeddings (semantic search), use:")
			fmt.Println("  [1] Local embeddings via Ollama (private, free — requires Ollama)")
			fmt.Println("  [2] OpenAI embeddings (better quality, small cost)")
			fmt.Print("> ")

			embedChoice := readLineBuf(reader)
			switch strings.TrimSpace(embedChoice) {
			case "2":
				cfg.DefaultEmbedder = "openai"
				if cfg.Keys.OpenAI == "" {
					fmt.Print("Enter your OpenAI API key: ")
					cfg.Keys.OpenAI = readLineBuf(reader)
				}
			default:
				cfg.DefaultEmbedder = "ollama"
				fmt.Printf("Ollama host (press Enter for %s): ", cfg.Ollama.Host)
				if host := readLineBuf(reader); host != "" {
					cfg.Ollama.Host = host
				}
			}

			fmt.Println()

			if err := config.SaveGlobal(cfg); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			path, _ := config.GlobalConfigPath()
			fmt.Printf("Configuration saved to %s\n", path)
			fmt.Println("Navigate to any project and run `memvra init` to get started.")

			return nil
		},
	}
}

// readLineBuf reads a trimmed line from a bufio.Reader.
func readLineBuf(r *bufio.Reader) string {
	line, _ := r.ReadString('\n')
	return strings.TrimRight(line, "\r\n")
}
