# Memvra

> Your AI finally remembers your project.

Memvra is a developer CLI tool that provides persistent, project-aware memory for AI coding assistants. It solves the core problem that AI agents (Claude, GPT, Gemini, etc.) are stateless across sessions — forcing developers to repeatedly re-explain their project architecture, conventions, constraints, and past decisions every time they start a new conversation.

## Features

- **Auto-indexes** your project — tech stack, architecture, folder structure, conventions
- **Remembers** decisions, constraints, and context across sessions
- **Retrieves** relevant context semantically when you ask a question
- **Injects** optimized context into any LLM call transparently
- **Exports** to CLAUDE.md, .cursorrules, or plain markdown
- **Local-first** — all data stays on your machine

## Installation

```bash
# macOS (Homebrew)
brew tap memvra/tap
brew install memvra

# Linux
curl -fsSL https://get.memvra.dev | sh

# Go install
go install github.com/memvra/memvra@latest
```

## Quick Start

```bash
# First-time setup (API keys, embedding provider)
memvra setup

# Initialize in your project
cd /path/to/your/project
memvra init

# Ask a question with full project context
memvra ask "How should I implement the document upload endpoint?"

# Store a decision
memvra remember "We use JWT auth, not Devise — API-only mode"

# Check what Memvra knows
memvra status
memvra context

# Export to CLAUDE.md
memvra export --format claude > CLAUDE.md
```

## Commands

| Command | Description |
|---------|-------------|
| `memvra setup` | Interactive first-time configuration |
| `memvra init` | Initialize and index the current project |
| `memvra ask "<question>"` | Ask with full project context injected |
| `memvra remember "<statement>"` | Store a decision, convention, or constraint |
| `memvra forget` | Remove specific memories |
| `memvra context` | View the current project context |
| `memvra status` | Show project stats |
| `memvra update` | Re-index changed files |
| `memvra export` | Export context to other tool formats |

## Configuration

Global config: `~/.config/memvra/config.toml`
Project config: `.memvra/config.toml`

Supported LLM providers: Claude, OpenAI, Gemini, Ollama (local)

## Development

```bash
# Build
make build

# Test
make test

# Install locally
make install
```

Requires Go 1.22+ with CGO enabled (for SQLite).

## License

MIT
