# curecode.md - Project Instructions

This file provides instructions for AI coding agents working on the CuRe Code project.

## Project Overview

CuRe Code is an AI coding agent CLI built with Go. It uses an agentic loop architecture where the AI autonomously calls tools to complete coding tasks.

## Architecture

- **Entry point**: `main.go` -> `cmd/root.go`
- **Agent loop**: `internal/agent/agent.go` - processes user prompts through the agentic cycle. Supports both batch and real-time streaming.
- **Process Manager**: `internal/agent/process.go` - manages background tasks started by the agent.
- **Tools**: `internal/tools/` - 15 built-in tools (read/write, web search, git info, symbol search, etc.).
- **AI Providers**: `internal/ai/fc_providers.go` & `streaming.go` - supports Gemini, OpenAI, Claude, NVIDIA NIM, Groq, DeepSeek, and Ollama.
- **Session Management**: `internal/agent/session.go` - handles persisting and resuming conversation history.
- **UI**: `internal/ui/` - terminal rendering (markdown, spinner, banner).

## Coding Conventions

1. **Go style**: Follow standard Go conventions (`gofmt`).
2. **Bilingual Documentation**: All exported structs, functions, and complex logic MUST have bilingual comments (English and Indonesian).
   - Use `// [EN] ...` for English.
   - Use `// [ID] ...` for Indonesian.
3. **Emoji-free UI**: Do NOT use emojis in UI outputs or logs. Use ASCII markers like `[T]`, `[OK]`, `[!]`.
4. **Error handling**: Always handle errors - never ignore with `_`.
5. **Concurrency**: Use Goroutines for performance-heavy tasks (e.g., `grep_search`).
6. **Tool interface**: All tools must implement `Name()`, `Description()`, `ParameterSchema()`, `Execute()`, `NeedsConfirmation()`.

## Key Design Decisions

- **Streaming Responses**: Token-by-token output for a responsive UX.
- **Session Persistence**: Conversations are saved as JSON files in the config directory.
- **Git Safety**: Agent is aware of git status and warns before making changes in "dirty" or non-git repos.
- **Task Tracking**: Built-in TODO system to track plan progress within a session.

## Adding a New Tool

1. Create `internal/tools/your_tool.go`.
2. Implement the `Tool` interface.
3. Register in `NewDefaultRegistry()` in `registry.go`.

## Adding a New AI Provider

1. Implement `StreamingProvider` interface in `internal/ai/fc_providers.go`.
2. Add to `CreateFCProvider` factory in `internal/ai/fc_providers.go`.
3. Add to model switch menu in `cmd/root.go`.
