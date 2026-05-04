<div align="center">
 
<img src="logo.png" width="400" alt="Forge Code Logo">

# Forge Code

**AI Coding Agent for Your Terminal**

[![Release](https://img.shields.io/github/v/release/broman0x/forge-code?label=Release&color=blue)](https://github.com/broman0x/forge-code/releases/latest)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Windows%20|%20Linux%20|%20macOS-lightgrey.svg)](https://github.com/broman0x/forge-code)
[![Go](https://img.shields.io/badge/Built%20with-Go-00ADD8.svg)](https://go.dev)

An agentic AI coding assistant that lives in your terminal.
Built with Go for speed and portability.

</div>

---

## What is Forge Code?

Forge Code is an **AI coding agent** that can read, write, and edit your code directly from the terminal. Unlike simple chat-based CLI tools, Forge Code uses an **agentic loop** — it autonomously decides which tools to use, reads your files, makes changes, runs commands, and iterates until the task is complete.

### Key Features

| Feature | Description |
|---------|-------------|
| **Agentic Loop** | AI autonomously reads, writes, edits files and runs commands |
| **15 Built-in Tools** | Read/Write, Web Search, Project Summary, Git Info, and more |
| **Multi-Provider** | Google Gemini, OpenAI, Claude, NVIDIA, Groq, DeepSeek, Ollama |
| **Smart Context** | Tag files directly in your prompt using `@filename` |
| **Process Manager** | Run and manage background tasks with `/ps` |
| **Confirmation Flow** | Dangerous operations require your approval (or use `--yolo`) |
| **Single Binary** | No Node.js, no Python — just one Go executable |
| **REPL + One-shot** | Interactive mode or `forgecode "fix the tests"` |

### Built-in Tools

| Tool | Purpose |
|------|---------|
| `read_file` / `read_many` | Read contents of one or multiple files |
| `write_file` / `edit_file` | Create new files or perform search-and-replace edits |
| `shell` | Execute any terminal command (with confirmation) |
| `grep_search` | Search for patterns across the entire project |
| `symbol_search` | Find functions, classes, and structs using regex |
| `project_summary` | Get a high-level overview of the codebase structure |
| `git_info` | Check current git status and branch information |
| `web_search` / `web_fetch` | Search the internet and fetch documentation/content |
| `todo` | Track and manage tasks within the agent session |

---

## Quick Start

```bash
# Download the binary for your platform, then:
./forgecode --install    # Install to PATH
forgecode                # Launch REPL
```

### One-shot Mode

```bash
forgecode "explain this codebase"
forgecode "add error handling to main.go"
forgecode "write tests for the auth module"
```

## Setup

### Option 1: Google Gemini (Recommended)

```bash
export GEMINI_API_KEY="your-key-from-aistudio.google.com"
forgecode
```

### Option 2: OpenAI

```bash
export OPENAI_API_KEY="your-key-from-platform.openai.com"
forgecode
```

### Option 3: Anthropic Claude

```bash
export ANTHROPIC_API_KEY="your-key-from-console.anthropic.com"
forgecode
```

### Option 4: NVIDIA NIM (Reasoning)

```bash
export NVIDIA_API_KEY="your-key-from-build.nvidia.com"
forgecode
```

### Option 5: DeepSeek

```bash
export DEEPSEEK_API_KEY="your-key"
forgecode
```

### Option 6: Ollama (Free, Local)

```bash
# Install from https://ollama.com
ollama pull llama3
forgecode
```

---

## Slash Commands

| Command | Description |
|---------|-------------|
| `/help` | Show available commands |
| `/model` | Switch AI provider/model |
| `/clear` | Clear screen |
| `/compact` | Clear conversation history |
| `/ps` | List or stop background processes |
| `/usage` | Show session token usage |
| `/save` | Save current session |
| `/resume` | Resume a saved session |
| `/version` | Show version |
| `/exit` | Exit Forge Code |

---

## How It Works

Forge Code uses a sophisticated **agentic architecture** to solve complex tasks:

```
User Prompt
    ↓
AI decides which tools to call
    ↓
┌──────────────────────────────┐
│   Tool Execution Loop        │
│                              │
│   read_file  → understand    │
│   edit_file  → make changes  │
│   shell      → run tests     │
│   grep_search → find patterns│
│   web_search → look up docs  │
│   git_info   → check status  │
│                              │
│   (with confirmation for     │
│    destructive operations)   │
│                              │
└──────────────────────────────┘
    ↓
AI provides final response
```

The AI has full autonomy to chain multiple tool calls until the task is complete, with a safety limit of 25 turns per prompt.

---

## Architecture

```
forge-code/
├── main.go                     # Entry point
├── cmd/
│   ├── root.go                 # REPL, one-shot, slash commands
│   └── install_self.go         # Self-installer
├── internal/
│   ├── agent/
│   │   ├── agent.go            # Agentic loop (core)
│   │   ├── message.go          # Message/ToolCall types
│   │   ├── process.go          # Background process manager
│   │   ├── session.go          # Session persistence
│   │   └── system_prompt.go    # Dynamic prompt builder
│   ├── ai/
│   │   ├── fc_providers.go     # Multi-provider implementation
│   │   └── streaming.go        # Real-time token streaming
│   ├── config/
│   │   └── config.go           # Configuration & API keys
│   ├── tools/
│   │   ├── registry.go         # Tool interface & registry
│   │   ├── read_file.go        # Read file contents
│   │   ├── write_file.go       # Create/overwrite files
│   │   ├── edit_file.go        # Search & replace edits
│   │   ├── shell.go            # Execute shell commands
│   │   ├── list_dir.go         # Browse directories
│   │   ├── grep.go             # High-performance search
│   │   ├── web_search.go       # Search the web
│   │   ├── project_summary.go  # Codebase overview
│   │   └── symbol_search.go    # Find symbols with regex
│   └── ui/
│       ├── banner.go           # Startup visual
│       └── markdown.go         # Terminal MD renderer
└── go.mod
```

---

## Building from Source

```bash
# Requirements: Go 1.25+
git clone https://github.com/broman0x/forge-code.git
cd forge-code
go build -o forgecode .

# Or cross-compile:
GOOS=linux GOARCH=amd64 go build -o forgecode-linux .
GOOS=darwin GOARCH=arm64 go build -o forgecode-mac .
GOOS=windows GOARCH=amd64 go build -o forgecode.exe .
```

---

## File Locations

| Platform | Config | Binary |
|----------|--------|--------|
| **Windows** | `%APPDATA%\ForgeCode\config.json` | `%LocalAppData%\ForgeCode\forgecode.exe` |
| **Linux/macOS** | `~/.config/forgecode/config.json` | `~/.local/bin/forgecode` |

---

## License

MIT — see [LICENSE](LICENSE) for details.
