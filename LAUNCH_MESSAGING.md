# Memvra Launch Messaging Kit

## Core One-Liners

**Tagline:** Your AI finally remembers your project.

**Elevator pitch:** Memvra gives AI coding assistants persistent memory of your project — architecture, decisions, and session history. Not just session migration, but long-term knowledge infrastructure.

**Value prop:** Stop re-explaining your project to AI tools. Memvra builds a vector-searchable knowledge base that remembers your conventions, constraints, and past decisions.

**Technical hook:** Local-first vector search + SQLite + MCP protocol for persistent AI memory.

**Differentiation:** Unlike session migration tools, Memvra maintains an active knowledge base over time — not just for switching tools, but for every session with every AI.

---

## Competitive Positioning

**Session migration tools (e.g., cli-continues):**
- Tactical: Resume your most recent session in a different AI tool
- One-time use: Grabs existing session files when you hit rate limits
- Reactive: Only useful when switching tools

**Memvra (persistent project memory):**
- Strategic: Builds and maintains a knowledge base over time
- Always-on infrastructure: Vector search, decision tracking, auto-export
- Proactive: Every session benefits from accumulated project knowledge
- Not just "continue" — remembers *why* decisions were made

**Key difference:** Memvra isn't just about switching tools — it's about never having to re-explain your project in the first place.

---

## Problem Statement

**The frustration:**
Every AI coding session starts from zero. You spend the first 10 minutes re-explaining your tech stack, architecture, conventions, and past decisions. Hit your token limit with Claude? Switch to Gemini and start over. Open Cursor IDE? It has no idea what you just discussed with Claude Code.

**The impact:**
- Wasted time re-explaining context
- Inconsistent advice across AI tools
- Lost momentum when switching tools
- Context debt accumulates across sessions

---

## Solution (How Memvra Works)

Memvra is a CLI tool that:
1. **Auto-indexes** your project (tech stack, file chunks, architecture)
2. **Remembers** decisions, conventions, constraints across sessions
3. **Retrieves** relevant context using vector search
4. **Auto-exports** to every AI tool's format (CLAUDE.md, .cursorrules, PROJECT_CONTEXT.md, JSON)
5. **Works with any LLM** — Claude, Cursor, Gemini, Ollama, GPT, etc.

### The magic moment:
```bash
# Morning: Working with Claude Code
# Hit token limit or quota...

# Afternoon: Switch to Gemini
memvra wrap gemini
> continue

# Gemini picks up exactly where Claude left off
```

All context files are auto-generated and kept in sync. No manual updates.

---

## Key Features (Condensed)

- ✅ **MCP server** — Claude Code & Cursor call Memvra tools automatically
- ✅ **Wrap mode** — Inject context into any CLI tool (Gemini, Aider, Ollama)
- ✅ **Auto-export** — Generates CLAUDE.md, .cursorrules, PROJECT_CONTEXT.md, JSON on every change
- ✅ **Semantic search** — Vector-based retrieval using SQLite + sqlite-vec
- ✅ **Session history** — AI tools see your recent conversations
- ✅ **Git integration** — Tracks staged/unstaged files, auto-updates on commits
- ✅ **100% local** — All data in `.memvra/` on your machine
- ✅ **Zero config** — `memvra setup && memvra init` and you're done

---

## Platform-Specific Hooks

### r/ClaudeAI
**Title:** I built an MCP server that gives Claude Code persistent memory of your entire project

**Hook:** Tired of re-explaining your architecture every session? Memvra is an MCP server that builds a vector-searchable knowledge base of your project — decisions, conventions, code structure, and session history. Claude automatically calls `memvra_save_progress`, `memvra_remember`, and `memvra_get_context`. Unlike session migration tools, Memvra maintains long-term project memory that improves every session.

**Why this matters here:** Claude users understand MCP protocol and crave better context management.

---

### r/cursor
**Title:** Give Cursor (and every other AI tool) persistent memory of your project

**Hook:** Memvra builds a knowledge base of your project that auto-syncs to `.cursorrules`, `CLAUDE.md`, and other formats. It's not just about switching tools — it's infrastructure that remembers your decisions, conventions, and architecture across every session. Vector search retrieves relevant code automatically. Works with Cursor, Claude Code, Gemini, and any AI tool.

**Why this matters here:** Cursor users often juggle multiple AI tools and hate losing context.

---

### r/LocalLLaMA
**Title:** Memvra — Persistent memory for AI assistants, 100% local, works with Ollama

**Hook:** Built with SQLite + sqlite-vec for vector search. All embeddings run locally via Ollama (`nomic-embed-text`). No cloud, no accounts, no telemetry. MIT licensed. Context stays in `.memvra/` on your machine.

**Why this matters here:** Privacy, local-first, and open-source are top priorities.

---

### r/ChatGPTCoding
**Title:** Stop re-explaining your project to every AI tool

**Hook:** Memvra gives ChatGPT/Claude/Gemini/Cursor a shared memory of your project. Switch between tools seamlessly — they all read from auto-generated context files (CLAUDE.md, .cursorrules, PROJECT_CONTEXT.md).

**Why this matters here:** These users switch between AI tools frequently and feel the context loss pain.

---

### r/programming
**Title:** Memvra – Persistent memory infrastructure for AI coding assistants (vector search + MCP + SQLite)

**Hook:** A CLI tool that builds a long-term knowledge base for AI assistants. Uses vector search (SQLite + sqlite-vec) for semantic code retrieval, MCP protocol for native Claude/Cursor integration, and auto-exports to every AI tool's format. Not just session migration — it's persistent project memory that improves over time. Built with Go, fully local, MIT licensed.

**Why this matters here:** Technical audience appreciates architecture details and open-source credentials.

---

### Hacker News
**Title:** Show HN: Memvra – Persistent memory for AI coding assistants

**Opening comment (post this as first comment):**

Hey HN! I'm Mohit, and I built Memvra to solve a problem I kept hitting: AI coding assistants are stateless. Every new session with Claude or Cursor means re-explaining my project architecture, conventions, and past decisions.

Memvra builds a persistent knowledge base of your project:
- Auto-indexes your codebase (tech stack, file chunks, semantic embeddings)
- Stores decisions, conventions, constraints, and session history
- Uses vector search (SQLite + sqlite-vec) for semantic code retrieval
- Auto-exports to CLAUDE.md, .cursorrules, PROJECT_CONTEXT.md, JSON
- MCP server integration for Claude Code and Cursor (they call Memvra tools automatically)
- Wrap mode for CLI tools like Gemini, Aider, Ollama

Unlike session migration tools (which just resume your last conversation), Memvra maintains long-term project memory. It's infrastructure that improves every session, not just tool switches.

Example: `memvra remember "We use JWT with RS256"` → Vector embedded, exported to all formats, retrieved semantically when relevant. Every AI session benefits from accumulated knowledge.

Tech stack: Go, SQLite, sqlite-vec for vector search, supports Claude/OpenAI/Gemini/Ollama.

MIT licensed: https://github.com/memvra/memvra

Happy to answer questions about the architecture, MCP integration, or vector search approach!

**Why this matters here:** HN loves technical depth, local-first tools, and developer productivity hacks.

---

### Dev.to
**Title:** How I Built Persistent Memory for AI Coding Assistants (Vector Search + MCP + SQLite)

**Opening:**
If you've used Claude Code, Cursor, or ChatGPT for coding, you've felt this pain: every session starts from zero. You re-explain your tech stack, architecture, and conventions. Hit your token limit? Switch to another AI tool and start over.

I built **Memvra** to solve this. It's a CLI tool that gives AI assistants a persistent memory of your project using vector search, SQLite, and the Model Context Protocol (MCP).

**Structure:**
1. The problem (context loss between sessions/tools)
2. How Memvra works (indexing, vector search, auto-export)
3. Code examples (setup, remember, ask, wrap)
4. Technical deep dive (SQLite schema, MCP tools, vector retrieval)
5. Demo GIF showing workflow
6. Installation & next steps

**Why this matters here:** Dev.to readers love technical walkthroughs with code examples.

---

### Product Hunt
**Tagline:** Your AI coding assistant finally remembers your project

**Thumbnail text:** Stop re-explaining your code every session

**First comment (as maker):**
Hey Product Hunt! 👋

I'm launching Memvra — a CLI tool that gives AI coding assistants persistent memory across sessions.

🔥 The problem: Every time you start a new Claude/Cursor/ChatGPT session, you re-explain your project. Switch tools mid-session? Start over.

✨ How Memvra fixes it:
- Auto-indexes your project (tech stack, architecture, files)
- Remembers decisions, conventions, constraints
- Auto-exports to CLAUDE.md, .cursorrules, PROJECT_CONTEXT.md, JSON
- Works with any AI tool (Claude, Cursor, Gemini, Ollama, ChatGPT)

🎯 The magic moment:
Hit your Claude token limit → `memvra wrap gemini` → type "continue" → Gemini picks up exactly where Claude left off.

💾 100% local. No cloud. MIT licensed.

Try it: https://github.com/memvra/memvra

What's your biggest frustration with AI coding tools?

**Why this matters here:** PH users want a clear value prop, visual demo, and community engagement.

---

## Technical Highlights (For HN/r/programming)

- **Vector search:** sqlite-vec extension with cosine similarity (768-dim embeddings)
- **Embedders:** Supports Ollama (nomic-embed-text), OpenAI (text-embedding-3-small), Gemini (text-embedding-004)
- **MCP protocol:** 8 tools exposed via stdio transport (save_progress, remember, get_context, search, etc.)
- **Smart chunking:** Language-aware code splitting (respects function/class boundaries)
- **Token budget:** Context builder fits chunks + memories within configurable token limit (default 8K)
- **Git integration:** Tracks working tree state, auto-updates on post-commit hook
- **Zero external dependencies:** Single binary with embedded SQLite + CGO

---

## Call to Action

**For all platforms:**

🔗 GitHub: https://github.com/memvra/memvra
📦 Install: `brew install memvra/tap/memvra`
📖 Docs: https://github.com/memvra/memvra#readme

**Engagement prompts:**
- What AI coding tool do you use most?
- How do you currently handle context between sessions?
- Would you use this with your current workflow?
- What feature would make this a must-have for you?

---

## Demo Script (For Recording)

**Setup:**
```bash
# Terminal with clean prompt, modern theme
cd ~/demo-project

# 1. Show the problem
memvra status  # "No project initialized"

# 2. Setup
memvra setup  # (pre-configured, quick)
memvra init   # Watch it scan files

# 3. Add a decision
memvra remember "We use PostgreSQL for JSONB support"

# 4. Ask a question
memvra ask "How should I implement the user preferences endpoint?"

# 5. Show the auto-exported files
ls -la | grep -E "CLAUDE.md|.cursorrules|PROJECT_CONTEXT.md"
cat CLAUDE.md  # Show first few lines

# 6. The magic: switch tools
memvra wrap gemini
> continue implementing the user preferences endpoint

# Gemini starts with full context — no re-explaining needed
```

**Key moments to capture:**
- Files being scanned during `init`
- Context being auto-exported
- Asking a question and getting a contextual answer
- Switching to another tool and typing "continue"

**Length:** 60-90 seconds max for GIF, 2-3 minutes for video

---

## Timing Strategy

**Phase 1: Quick wins (This week)**
- Reddit posts to r/ClaudeAI, r/cursor, r/LocalLLaMA
- Dev.to article

**Phase 2: Major launch (Next week)**
- Hacker News (weekday morning, 8-10am ET)
- Monitor HN comments, engage immediately

**Phase 3: Community (Ongoing)**
- Product Hunt (after HN traction)
- Share HN/Reddit discussions on LinkedIn
- Engage with feedback, iterate

---

## FAQ (Pre-written Answers)

**Q: How is this different from just using CLAUDE.md or .cursorrules?**
A: Those files are static and manually maintained. Memvra auto-generates them on every change, includes session history + git state, and uses vector search to inject relevant code chunks dynamically.

**Q: Does this send my code to the cloud?**
A: No. Everything runs locally. Your code and embeddings stay in `.memvra/` on your machine. Only LLM API calls go to the cloud (if using Claude/OpenAI/Gemini), but you can use Ollama for 100% local operation.

**Q: What if I already have a .cursorrules file?**
A: Memvra won't overwrite it by default. You can disable auto-export for specific formats in config, or merge your manual rules with Memvra's output.

**Q: Can I use this with [tool X]?**
A: If it reads CLAUDE.md, .cursorrules, PROJECT_CONTEXT.md, or JSON — yes. If it's an MCP-compatible tool — yes. If it's a CLI tool — use `memvra wrap`.

**Q: How much disk space does it use?**
A: SQLite DB size depends on project size. Typical: 5-20MB for a medium project. Check with `memvra status`.

**Q: Can I edit the exported files manually?**
A: Yes, but Memvra will regenerate them on the next change. Better to use `memvra remember` to store decisions, or disable auto-export and use `memvra export` manually.

---

## Social Proof Ideas (Future)

- GitHub stars milestone posts
- User testimonials
- "Built with Memvra" showcase
- Integration guides (Aider, Gemini CLI, etc.)
- Performance benchmarks (context retrieval speed)

---

## Hashtags (If needed)

#AI #DevTools #ClaudeAI #Cursor #LocalLLM #OpenSource #MCP #VectorSearch #SQLite #DeveloperProductivity
