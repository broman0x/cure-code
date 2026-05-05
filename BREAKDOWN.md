# BREAKDOWN.md - CuRe Code Architecture Analysis

**Date:** 2026-05-05  
**Repository:** https://github.com/broman0x/cure-code  
**Local Path:** `/home/ev3lynx/.openclaw/workspace-gh0st/dev/cure-code/`  
**Language:** Go 1.25.4  
**License:** GPLv3  

---

## What is CuRe Code?

CuRe Code is a **terminal-only AI coding agent** written in Go. It's a single binary (no Python/Node dependencies) that autonomously reads/writes code, runs commands, and iterates until tasks complete via an agentic loop.

Think "Claude Code but written in Go" — focused purely on local coding workflows.

---

## Core Architecture

```
cure-code/
├── main.go                          # Entry point → cmd.Execute()
│   └── defer pauseExit()           # Windows-only: waits for Enter on exit
│
├── cmd/                             # CLI layer (cobra + go-prompt)
│   ├── root.go                      # REPL, one-shot, slash commands, provider setup
│   └── install_self.go              # Self-installer (--install/--uninstall)
│
├── internal/
│   ├── agent/                       # CORE: Agentic loop & session management
│   │   ├── agent.go                 # Main agentic loop (ProcessPrompt, tool calling)
│   │   ├── message.go               # Message/ContentBlock/ToolCall types
│   │   ├── process.go               # Background process manager (/ps command)
│   │   ├── session.go               # Session persistence (save/load JSON)
│   │   ├── system_prompt.go         # Dynamic system prompt builder
│   │   ├── context.go               # Proactive context suggestions
│   │   ├── intelligence.go         # Smart file suggestions
│   │   ├── mentions.go              # @filename tagging logic
│   │   ├── skills.go                # Skills management
│   │   ├── compaction.go           # Context compaction (like Hermes compression)
│   │   └── delegate.go             # Sub-agent delegation (from tools/)
│   │
│   ├── ai/                          # Multi-provider AI support
│   │   ├── fc_providers.go         # Provider factory (CreateFCProvider)
│   │   └── streaming.go            # Real-time token streaming per provider
│   │
│   ├── tools/                       # 17 built-in tools
│   │   ├── registry.go             # Tool interface & registry
│   │   ├── read_file.go            # Read file contents
│   │   ├── read_many_files.go      # Batch read multiple files
│   │   ├── write_file.go           # Create/overwrite files
│   │   ├── edit_file.go            # Search & replace edits
│   │   ├── shell.go                # Execute shell commands (with confirmation)
│   │   ├── list_dir.go             # Browse directories
│   │   ├── grep.go                 # High-performance regex search
│   │   ├── symbol_search.go        # Find functions/classes/structs via regex
│   │   ├── project_summary.go      # Codebase overview
│   │   ├── git_info.go             # Git status/branch info
│   │   ├── web_search.go          # Search the web (DuckDuckGo?)
│   │   ├── web_fetch.go            # Fetch URL content
│   │   ├── todo.go                 # Task tracking within session
│   │   ├── ask_user.go             # Ask user for input during agentic loop
│   │   ├── plan_mode.go            # PLAN.md synchronization
│   │   └── delegate.go            # Spawn sub-agents
│   │
│   ├── config/
│   │   └── config.go               # Configuration & API key management (JSON)
│   │
│   └── ui/
│       ├── banner.go                # Startup ASCII art
│       ├── markdown.go              # Terminal markdown renderer
│       └── (spinner, colors, etc.) # Terminal UI elements
│
├── web/                             # Web UI? (possibly unfinished)
│
├── go.mod                           # Go 1.25.4 dependencies
├── go.sum
├── .env.example                     # API key templates
├── config.json (global)             # ~/.config/curecode/config.json
└── .env (global)                    # ~/.config/curecode/.env
```

---

## Key Components Deep Dive

### 1. Entry Point (`main.go`)

```go
func main() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("CRITICAL PANIC:", r)
        }
        pauseExit()  // BLOCKS: waits for Enter
    }()
    if err := cmd.Execute(); err != nil {
        fmt.Println("[!] Error:", err)
        os.Exit(1)
    }
}
```

**Issues:**
- `pauseExit()` runs on ALL exits (even success)
- Panic recovery prints but doesn't re-raise
- Windows-centric design (Enter to close window)

---

### 2. CLI Layer (`cmd/root.go`)

**Commands:**
- `curecode` → Interactive REPL (requires TTY)
- `curecode "prompt"` → One-shot mode
- `curecode --resume <id>` → Resume session in REPL
- `curecode --install` → Install to PATH
- `curecode --version` → Show version

**Flow:**
```
User Input
    ↓
createAgent() → Load config, create provider (Gemini/OpenAI/Ollama)
    ↓
runREPL() or runOneShot()
    ↓
agent.ProcessPrompt() → Agentic loop
    ↓
Tool calls → Read/Write/Shell/Grep
    ↓
Response → Streaming output
    ↓
Save session (if requested)
```

**Issues:**
- `runOneShot()` still requires TTY (go-prompt initializes)
- `--resume` forces REPL, can't combine with one-shot
- Provider fallback broken (config overrides env vars)

---

### 3. Agentic Loop (`internal/agent/agent.go`)

The core intelligence. Processes user prompts through:

1. **System Prompt Builder** (`system_prompt.go`) — Injects tool definitions, context, memory
2. **Provider Call** — Sends messages to AI (Gemini/OpenAI/Claude/Ollama)
3. **Response Parsing** — Extracts text + tool calls
4. **Tool Execution** — Runs tools (read_file, shell, grep, etc.)
5. **Loop** — Feeds tool results back to AI (max 25 turns)
6. **Final Response** — Streams final answer to user

**Tool Interface:**
```go
type Tool interface {
    Name() string
    Description() string
    ParameterSchema() string  // JSON schema
    Execute(params map[string]interface{}) (string, error)
    NeedsConfirmation() bool  // Dangerous operations
}
```

**Max Turns:** 25 per prompt (hardcoded safety limit)

---

### 4. Multi-Provider Support (`internal/ai/`)

**Supported Providers:**
| Provider | Model Examples | Streaming |
|----------|---------------|-----------|
| Gemini | gemini-2.5-flash, gemini-2.5-pro | ✓ |
| OpenAI | gpt-4o-mini, gpt-4o | ✓ |
| Claude | claude-sonnet-4-20250514 | ✓ |
| NVIDIA | nemotron-3-super-120b-a12b | ✓ |
| xAI (Grok) | grok-2-1212 | ✓ |
| DeepSeek | deepseek-coder | ✓ |
| Ollama | llama3, qwen2.5, etc. | ✓ |
| Together | Meta-Llama-3.1-70B | ✓ |
| Mistral | mistral-large-latest | ✓ |

**Provider Factory:**
```go
func CreateFCProvider(providerType, model string) (FunctionCallingProvider, error) {
    switch providerType {
    case "gemini":
        return &GeminiProvider{model: model}, nil
    case "openai":
        return &OpenAIProvider{model: model}, nil
    // ... etc.
    }
}
```

**Issues:**
- No model validation on creation (Issue #7)
- Ollama model "llama3" hardcoded but may not exist (Issue #4)
- Some models don't support tools (Issue #6)

---

### 5. Tools (`internal/tools/`)

**17 Built-in Tools:**

| Tool | Purpose | Confirmation |
|------|----------|--------------|
| `read_file` | Read file contents | No |
| `read_many_files` | Batch read multiple files | No |
| `write_file` | Create/overwrite files | **Yes** |
| `edit_file` | Search & replace edits | **Yes** |
| `shell` | Execute shell commands | **Yes** |
| `list_dir` | Browse directories | No |
| `grep` | Regex search across project | No |
| `symbol_search` | Find functions/classes/structs | No |
| `project_summary` | Codebase overview | No |
| `git_info` | Git status/branch | No |
| `web_search` | Search the web | No |
| `web_fetch` | Fetch URL content | No |
| `todo` | Task tracking | No |
| `ask_user` | Ask user for input | No |
| `plan_mode` | PLAN.md sync | No |
| `delegate` | Spawn sub-agents | No |

**Tool Registry:**
```go
func NewDefaultRegistry() *Registry {
    r := &Registry{}
    r.Register(&ReadFileTool{})
    r.Register(&WriteFileTool{})
    // ... 15 more
    return r
}
```

---

### 6. Memory System (Corrected Analysis)

**README Claims ("Agentic Memory V1"):**
- ✗ Persistent symbol tracking → PARTIALLY implemented (code exists but broken)
- ✗ File tree awareness → NOT PERSISTENT (rebuilt each session)
- ✗ State sync → NOT IMPLEMENTED (state.json has zeros/nulls)

**What Actually Exists:**

#### 6.1 Symbol Tracking (`RecentSymbols`)
```go
// agent.go
type Agent struct {
    RecentSymbols []string  // Slice of recently seen symbols
}

func (a *Agent) updateRecentSymbols(newSymbols []string) {
    a.RecentSymbols = append(newSymbols, a.RecentSymbols...)
    // Deduplicate, keep top 20
}
```

**Problem:** `state.json` shows `"recent_symbols": []` (ALWAYS EMPTY)

**Root Cause:** `updateRecentSymbols()` only called from `search_symbol` tool result — if tool never used, symbols never tracked.

**Used in:**
- `BuildSystemPrompt()` → Injected into system prompt
- `summarizeHistory()` (compaction) → Referenced in prompt but always empty

---

#### 6.2 State Persistence (`state.json`)
**Saved to:** `.curecode/state.json` in **WORKING DIRECTORY** (NOT `~/.config/curecode/`)
**Code Location:** `agent.go` lines 783-791

```go
func (a *Agent) saveState() {
    dir := filepath.Join(a.WorkDir, ".curecode")  // WRONG: should be config dir
    os.MkdirAll(dir, 0755)
    path := filepath.Join(dir, "state.json")
    _ = os.WriteFile(path, data, 0644)  // Error IGNORED
}
```

**state.json contents (from cure-code project dir):**
```json
{
  "project_name": "cure-code",
  "recent_symbols": [],        // EMPTY - symbol tracking broken
  "tasks": null,                // NULL - task tracking not saved
  "history_count": 1,           // But no actual history
  "last_turn_time": "2026-05-05T19:27:51+07:00",
  "usage": {
    "TotalInputTokens": 0,      // ZERO - usage tracking broken
    "TotalOutputTokens": 0,     // ZERO
    "TotalTokens": 0,            // ZERO
    "RequestCount": 0             // ZERO
  },
  "tool_call_count": 0,         // ZERO - tool tracking broken
  "is_planning": false,
  "agent_version": "1.0.2"
}
```

**Issues:**
1. Saved to WRONG location (working dir, not config dir) → Issue #18
2. Error ignored on save → Issue #19
3. All usage metrics are ZERO → Issue #20
4. `recent_symbols` always empty → Issue #17
5. Not persistent across working directories

---

#### 6.3 Session Management (`internal/agent/session.go`)

**Advertised:** "Session Persistence: Conversations are saved as JSON files in the config directory."

**Actual Code:**
```go
func SaveSession(history []Message, tasks []Task, workDir, configDir string) (string, error) {
    sessionDir := filepath.Join(configDir, "sessions")
    if err := os.MkdirAll(sessionDir, 0755); err != nil {
        return "", fmt.Errorf("failed to create session directory: %v", err)
    }
    // ... save to sessionDir/session-<timestamp>.json
}
```

**Problem:** From stress testing: `~/.config/curecode/sessions/` DOESN'T EXIST

**Where called:** `cmd/root.go` line 238:
```go
id, _ := agent.SaveSession(ag.History, ag.Tasks, ag.WorkDir, configDir)  // Error IGNORED
color.HiBlack("  Session auto-saved as %s", id)  // Prints even on error!
```

**Issues:**
- Session save error ignored (`_`) → Issue #19
- Sessions directory never created (despite `os.MkdirAll()`)
- `/save` and `--resume` commands broken

---

#### 6.4 Compaction System (`compaction.go`)
**Inspired by:** Claude Code's context management

**How it works:**
1. `checkAndCompact()` monitors token usage
2. When threshold exceeded → calls `summarizeHistory()`
3. AI generates "High-Fidelity Memory Block"
4. History replaced with summary + last 8 messages

**Issues:**
- Uses `a.Usage.TotalInputTokens` which is always 0 → Compaction NEVER TRIGGERS
- References `a.RecentSymbols` which is always empty → No spatial context in summary
- Token counting not implemented for most providers

---

#### 6.5 Skills System (`skills.go`)
**Implementation:**
```go
type SkillRegistry struct {
    skills map[string]Skill  // In-memory ONLY
}

func (r *SkillRegistry) LoadFromDir(dir string) error {
    // Loads from .curecode/skills/ directory
    // Loaded FRESH each session
}
```

**Issues:**
- No persistence across sessions (in-memory only)
- No tracking of which skills were used
- No learning or personalization
- Built-in skills hardcoded (CodeReview, TestAndFix)

---

### 7. Session Management (Legacy Section - See Section 6.3)

**Save Format:** JSON file in `~/.config/curecode/sessions/<session_id>.json`

**Contents:**
- Message history (role, content, tool calls)
- Task list (TODO items)
- Session metadata (timestamp, model, provider)

**Issues:**
- Sessions directory NOT created (Issue #3, #19)
- Resume functionality broken
- No session listing command
- Error ignored on save

---

### 7. Configuration (`internal/config/config.go`)

**Format:** JSON (NOT YAML)

**Location:**
- Linux/macOS: `~/.config/curecode/config.json`
- Windows: `%APPDATA%\CuReCode\config.json`

**Structure:**
```json
{
  "language": "en",
  "first_run": false,
  "last_model": "llama3",
  "last_provider": "ollama",
  "install_path": "",
  "version": "2.0.0"
}
```

**API Keys:** Stored in `~/.config/curecode/.env` (dotenv format)

**Issues:**
- Version mismatch (binary v1.0.2, config defaults to v2.0.0)
- `last_provider` overrides env vars (Issue #5)
- No validation of config values

---

## Interesting Design Choices

### 1. Bilingual Code Comments
All exported functions have dual comments:
```go
// [EN] main is the entry point of the application.
// [ID] main adalah titik masuk aplikasi.
func main() { ... }
```
**Author is Indonesian** ("bromanprjkt" = "broman project").

### 2. No Emojis in UI
Strict ASCII-only output:
- `[T]` for thinking
- `[OK]` for success
- `[!]` for warnings/errors

Per project rules in `CURECODE.md`.

### 3. Go 1.25+ Required
Very recent Go version (1.25.4). May limit compatibility with older systems.

### 4. Single Binary Philosophy
- No Python/Node dependencies
- All tools built-in (no MCP-like extensibility)
- Cross-compiles easily (Linux/macOS/Windows)

### 5. Agentic Loop Limitations
- Max 25 turns per prompt (hardcoded)
- No memory across sessions (unlike Hermes Redis/PostgreSQL)
- Simple JSON file persistence

---

## Comparison to OpenClaw/Ghostclaw

| Aspect | CuRe Code | OpenClaw/Ghostclaw |
|--------|-----------|-------------------|
| **Language** | Go (single binary) | Python (FastAPI, PydanticAI) |
| **Scope** | Terminal-only coding agent | Multi-platform (CLI, Discord, Telegram, etc.) |
| **Agent Model** | Single agent + sub-agents | Multi-agent (scout, coworker, vtwin) |
| **Memory** | JSON files (BROKEN - Issues #17-23) | Redis + PostgreSQL + LanceDB (hybrid) |
| **Tools** | 17 built-in coding tools | Extensible via MCP, plugins, skills |
| **Providers** | 9 AI providers (cloud + local) | OpenRouter + direct APIs |
| **Extensibility** | None (hardcoded tools) | High (MCP servers, skills, plugins) |
| **Use Case** | Local coding tasks | General-purpose orchestration, RAG, infra |
| **Session State** | JSON files (wrong location - Issue #18) | Hybrid (files + Redis + PostgreSQL) |
| **Context Mgmt** | Compaction (BROKEN - tokens always 0) | Compression + warm cache + summarization |
| **Symbol Tracking** | PARTIAL (always empty - Issue #17) | ✅ Working |
| **Cross-Session Memory** | ❌ None (in-memory only) | ✅ Yes (Redis/PostgreSQL) |

---

## Issues Summary (from ISSUES.md)

**23 total issues** found during stress testing:

| # | Issue | Severity | Category |
|---|-------|----------|----------|
| 1 | CRITICAL PANIC: no such device or address | HIGH | TTY |
| 2 | Hangs on valid model (qwen2.5:3b) | HIGH | Provider |
| 3 | No sessions directory created | HIGH | Session |
| 4 | Ollama llama3 model not found | MEDIUM | Provider |
| 5 | Provider fallback broken | MEDIUM | Provider |
| 6 | Model doesn't support tools | MEDIUM | Provider |
| 7 | No model validation on startup | LOW | Provider |
| 8 | TTY required even for one-shot | HIGH | TTY |
| 9 | File tagging @ fails in non-TTY | MEDIUM | TTY |
| 10 | pauseExit() on all exits | LOW | TTY |
| 11 | Slash commands only in interactive | MEDIUM | UX |
| 12 | Slash commands treated as prompts | LOW | UX |
| 13 | Version mismatch | LOW | Config |
| 14 | One-shot fails but prints help | LOW | Config |
| 15 | --resume broken with one-shot | MEDIUM | Config |
| 16 | --install success waits for Enter | LOW | Config |
| **17** | **FALSE ADVERTISING: "Agentic Memory V1" NOT FULLY IMPLEMENTED** | **HIGH** | **Memory/Documentation** |
| **18** | **`state.json` Saved to WRONG LOCATION (working dir)** | **HIGH** | **Memory** |
| **19** | **Session Saving BROKEN (ERROR IGNORED)** | **HIGH** | **Session/Memory** |
| **20** | **`state.json` Shows ZEROS for Usage** | **MEDIUM** | **Memory/Usage Tracking** |
| **21** | **No Real Memory Backend** | **MEDIUM** | **Architecture** |
| **22** | **Compaction References `RecentSymbols` But They're Empty** | **LOW** | **Memory/Compaction** |
| **23** | **Skills System is IN-MEMORY ONLY** | **LOW** | **Memory/Skills** |

**Priority Fixes:**
1. Fix TTY panic (Issue #1) — blocks all non-interactive use
2. Fix model hanging (Issue #2) — valid models don't work
3. Fix FALSE ADVERTISING (Issue #17) — remove claims or implement
4. Fix wrong state location (Issue #18) — save to config dir
5. Fix session save (Issue #19) — stop ignoring errors
6. Fix one-shot TTY requirement (Issue #8) — main use case broken

---

## Potential Improvements

### For CuRe Code Authors:
1. **Fix TTY dependency** — Allow non-interactive mode without go-prompt
2. **Validate models on startup** — Check if model exists before trying to use it
3. **Create sessions directory** — Make session persistence actually work
4. **Remove pauseExit() on Linux** — Clean exit for terminal users
5. **Add MCP support** — Extensibility like Hermes/Claude Code
6. **Fix memory backend** — Implement REAL persistent memory (Redis/SQLite)
7. **Fix usage tracking** — Make token counting work for ALL providers
8. **Stop ignoring errors** — Never use `_` for error returns

### For Your Evaluation (mr zer0):
- Good for **simple local coding tasks** where you want a single binary
- Lacks the **multi-agent orchestration** of OpenClaw
- No **MCP extensibility** (hardcoded tools only)
- **Memory system is BROKEN** (not just basic — actually broken)
- False advertising: README promises features that don't work
- Could be useful as a **lightweight coding agent** for quick tasks (IF fixed)

---

## File Locations (Installed)

```
Binary:        /home/ev3lynx/.local/bin/curecode  (7.7MB)
Config:        ~/.config/curecode/config.json
Env:           ~/.config/curecode/.env
Sessions:      ~/.config/curecode/sessions/  (DOESN'T EXIST - broken)
State (wrong): .curecode/state.json in WORKING DIR (not persistent)
Source:        /home/ev3lynx/.openclaw/workspace-gh0st/dev/cure-code/
```

**Key Insight:** `state.json` is saved to `.curecode/` in the CURRENT WORKING DIRECTORY, NOT in `~/.config/curecode/`. This means:
- State is NOT persistent across different working directories
- State is LOST if you delete the project directory
- Clutters project dirs with `.curecode/` folder

---

## Memory Backend Verdict

**CuRe Code's "Agentic Memory V1" is MISLEADING:**

| Feature | Advertised | Actual | Status |
|---------|------------|--------|--------|
| Symbol tracking | ✅ Persistent | ❌ Always empty (`[]`) | BROKEN |
| File tree awareness | ✅ Persistent | ❌ Rebuilt each session | FAKE |
| State sync | ✅ Working | ❌ Zeros/nulls in state.json | BROKEN |
| Session persistence | ✅ JSON files | ❌ Directory never created | BROKEN |
| Cross-session memory | ✅ Implied | ❌ None (in-memory only) | MISSING |
| Usage tracking | ✅ Implied | ❌ All zeros | BROKEN |

**Compared to OpenClaw/Hermes:**
- OpenClaw: Redis + PostgreSQL + LanceDB (real persistent memory)
- CuRe Code: Broken JSON files + in-memory only (no persistence)

---

## PR #26 Analysis - "Fix: 25 Issues Proposal + OpenRouter Provider Support"

**PR Status:** OPEN  
**Author:** Ev3lynx727 (mr zer0's GitHub)  
**Branch:** `fix/25-issues-proposal`  
**Files Changed:** 3 files (+640 additions, 0 deletions)

### Files in PR:
| File | Changes | Description |
|------|---------|-------------|
| `ISSUES.md` | +351 lines | All 25 identified issues documented |
| `PROPOSAL_FIX_ISSUES.md` | +279 lines | Comprehensive fix roadmap (5 phases) |
| `internal/ai/fc_providers.go` | +10 lines | ✅ OpenRouter API support ADDED |

### PR Summary:
```
## Summary
This PR adds comprehensive issue tracking and proposed fixes for CuRe Code.

### Changes
1. **OpenRouter Provider Support** - Added OpenRouter API support
2. **ISSUES.md** - Documents all identified issues (25 total)
3. **PROPOSAL_FIX_ISSUES.md** - Comprehensive fix proposal with roadmap

### Issues Covered
- CRITICAL: TTY panic, model hang, sessions directory
- MEDIUM: Provider fallback, Ollama auto-detect
- LOW: pauseExit(), version mismatch
```

---

## Fix Roadmap (from PROPOSAL_FIX_ISSUES.md)

### Phase 1: Critical Fixes (Week 1 - MUST FIX)
| # | Issue | Priority |
|---|-------|----------|
| 1 | TTY Panic in Non-Interactive Mode (#1, #8) | HIGH |
| 3 | Sessions Directory Not Created (#3, #15) | HIGH |
| 2 | Model Hangs on Valid Models (#2) | HIGH |

**Key Fix:** TTY detection with `terminal.IsTerminal()`, fallback to non-interactive mode.

---

### Phase 2: Provider/Model Fixes (Week 2)
| # | Issue | Priority |
|---|-------|----------|
| 5 | Provider Fallback Broken (#5) | MEDIUM |
| 4 | Ollama Model Auto-Detection (#4) | MEDIUM |
| 6 | Model Tool Support Validation (#6) | MEDIUM |

**Key Fix:** Query `ollama list` on init, validate tool support per model.

---

### Phase 3: UX/TTY Fixes (Week 3)
| # | Issue | Priority |
|---|-------|----------|
| 10 | pauseExit() on All Exits (#10) | LOW |
| 11, 12 | Slash Commands in One-Shot (#11, #12) | MEDIUM |
| 9 | File Tagging @ in Non-TTY (#9) | MEDIUM |

**Key Fix:** OS detection for pauseExit(), parse `/` prefix before REPL init.

---

### Phase 4: Enhancements (Already Implemented)
| # | Issue | Status |
|---|-------|--------|
| 17 | OpenRouter Support | ✅ ALREADY IMPLEMENTED |
| 18 | MCP/Plugin Support | ⏳ Not implemented |

**OpenRouter Code (ALREADY IN PR):**
```go
// fc_providers.go
case "openrouter":
    key := os.Getenv("OPENROUTER_API_KEY")
    if key == "" {
        return nil, fmt.Errorf("OPENROUTER_API_KEY not found")
    }
    if modelName == "" {
        modelName = "anthropic/claude-3.5-sonnet"
    }
    return NewGenericOpenAIFCProvider(key, modelName, "https://openrouter.ai/api/v1", "OpenRouter"), nil
```

---

### Phase 5: Minor Fixes (LOW Priority)
- #13 Version Mismatch → Update build script with `-ldflags`
- #15 `--resume` with One-Shot → Allow combining flags
- #16 `--install` waits for Enter → Skip pauseExit for flags
- #14 One-Shot Fails with Help → Print just error

---

## Updated Issue Count

**TOTAL: 25 Issues** (16 original + 9 additional from PR)

| Severity | Count | Issues |
|----------|-------|--------|
| **HIGH** | 6 | #1, #2, #3, #17, #18, #19 |
| **MEDIUM** | 12 | #4, #5, #6, #8, #15, #20, #21, + 4 more in PR |
| **LOW** | 7 | #7, #9, #10, #11, #12, #13, #14, #22, #23 |

---

**Analysis by:** Hermes Agent (hy3-preview)  
**Date:** 2026-05-05  
**Session:** Fresh startup, post-stress-test + memory audit + PR #26 analysis
