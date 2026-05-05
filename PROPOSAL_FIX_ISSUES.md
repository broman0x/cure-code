# PROPOSAL_FIX_ISSUES.md - CuRe Code Fix Proposal

**Date:** 2026-05-05  
**Based on:** BREAKDOWN.md & ISSUES.md  
**Total Issues:** 16 (original) + 9 (additional) = 25

---

## Executive Summary

This document proposes fixes for all issues identified in CuRe Code. Issues are categorized by severity and priority, with specific code-level implementation suggestions.

| Severity | Count | Status |
|----------|-------|--------|
| **HIGH** | 6 | Needs immediate attention |
| **MEDIUM** | 10 | Important improvements |
| **LOW** | 9 | Nice to have |

---

## Phase 1: Critical Fixes (Must Fix)

### 1. TTY Panic in Non-Interactive Mode (Issues #1, #8)

**Problem:** `go-prompt` requires TTY, causes panic in CI/CD/piped scenarios.

**Files Affected:**
- `main.go`
- `cmd/root.go`

**Proposed Fix:**
```go
// main.go - Add TTY detection
import "golang.org/x/crypto/ssh/terminal"

func isTerminal(f *os.File) bool {
    return terminal.IsTerminal(int(f.Fd()))
}

func runNonInteractive(prompt string) error {
    // Skip showBanner(), skip pauseExit()
    // Process prompt directly
}
```

**Priority:** HIGH - Blocks all CI/CD usage

---

### 2. Sessions Directory Not Created (Issues #3, #15)

**Problem:** `~/.config/curecode/sessions/` never created, session save errors ignored.

**Files Affected:**
- `internal/agent/session.go`
- `cmd/root.go`

**Proposed Fix:**
```go
// session.go - SaveSession()
func SaveSession(...) (string, error) {
    sessionsDir := filepath.Join(configDir, "sessions")
    
    // CREATE DIRECTORY FIRST
    if err := os.MkdirAll(sessionsDir, 0755); err != nil {
        return "", fmt.Errorf("failed to create sessions dir: %w", err)
    }
    // ... rest of save logic
}
```

**Priority:** HIGH - Core feature broken

---

### 3. Model Hangs on Valid Models (Issue #2)

**Problem:** qwen2.5:3b hangs for 35+ seconds then times out.

**Files Affected:**
- `internal/ai/fc_providers.go`

**Proposed Fix:**
```go
// fc_providers.go - OllamaFCProvider
func (o *OllamaFCProvider) SendWithTools(...) (*agent.Response, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()
    
    req, _ := http.NewRequestWithContext(ctx, "POST", o.BaseURL, bytes.NewBuffer(payload))
    // ... rest of request
}
```

**Priority:** HIGH - Valid models don't work

---

## Phase 2: Provider/Model Fixes

### 4. Provider Fallback Broken (Issue #5)

**Problem:** Config's `last_provider` overrides available API keys.

**Proposed Fix:**
```go
func detectAvailableProvider() string {
    providers := []struct {
        name string
        key  string
    }{
        {"gemini", os.Getenv("GEMINI_API_KEY")},
        {"openai", os.Getenv("OPENAI_API_KEY")},
    }
    for _, p := range providers {
        if p.key != "" {
            return p.name
        }
    }
    return ""
}
```

**Priority:** MEDIUM

---

### 5. Ollama Model Auto-Detection (Issue #4)

**Problem:** Default model `llama3` not available.

**Proposed Fix:**
Query `ollama list` on init and select available model.

**Priority:** MEDIUM

---

### 6. Model Tool Support Validation (Issue #6)

**Problem:** No warning when model doesn't support tools.

**Proposed Fix:**
Maintain list of tool-supporting models per provider.

**Priority:** MEDIUM

---

## Phase 3: UX/TTY Fixes

### 7. pauseExit() on All Exits (Issue #10)

**Problem:** Runs on all exits, inappropriate for Linux.

**Proposed Fix:**
```go
func main() {
    defer func() {
        if runtime.GOOS != "windows" {
            return // Skip on Linux/macOS
        }
        pauseExit()
    }()
}
```

**Priority:** LOW

---

### 8. Slash Commands in One-Shot (Issues #11, #12)

**Problem:** Slash commands only work in interactive mode.

**Proposed Fix:**
Parse `/` prefix before processing prompt.

**Priority:** MEDIUM

---

### 9. File Tagging @ in Non-TTY (Issue #9)

**Problem:** `@filename` requires TTY.

**Proposed Fix:**
Parse `@filename` before REPL initialization.

**Priority:** MEDIUM

---

## Phase 4: Enhancements (Already Implemented)

### 10. OpenRouter Support (NEW - Issue #17)

**Status:** ✅ ALREADY IMPLEMENTED

```go
case "openrouter":
    key := os.Getenv("OPENROUTER_API_KEY")
    // Added to fc_providers.go
```

---

### 11. MCP/Plugin Support (Issue #18)

**Status:** ⏳ Not implemented - requires design decision

**Option 1: MCP Client**
```go
type MCPClient struct {
    transport Transport
}
```

**Option 2: Simple Plugin System**
```
~/.config/curecode/plugins/
  ├── my-plugin/
  │   └── tool.go
```

**Priority:** MEDIUM (enhancement)

---

## Phase 5: Minor Fixes

### 12. Version Mismatch (Issue #13)
**Fix:** Update build script with ldflags

### 13. --resume with One-Shot (Issue #15)
**Fix:** Allow combining flags

### 14. --install waits for Enter (Issue #16)
**Fix:** Skip pauseExit for install/uninstall/version

### 15. One-Shot Fails with Help (Issue #14)
**Fix:** Print just error, not full help

---

## Implementation Roadmap

### Sprint 1: Critical (Week 1)
1. TTY detection + fallback (Issues #1, #8)
2. Sessions directory + error handling (Issues #3, #15)
3. Model timeout (Issue #2)

### Sprint 2: Provider Fixes (Week 2)
4. Provider fallback logic (Issue #5)
5. Ollama model auto-detection (Issue #4)
6. Tool support validation (Issue #6)

### Sprint 3: UX Improvements (Week 3)
7. pauseExit() OS detection (Issue #10)
8. Slash commands in one-shot (Issues #11, #12)
9. File tagging in non-TTY (Issue #9)

### Future: Enhancements
10. MCP/Plugin system
11. OpenRouter ✅ DONE

---

## Summary

| Phase | Issues | Priority |
|-------|--------|----------|
| Phase 1 | #1, #2, #3, #8, #15 | CRITICAL |
| Phase 2 | #4, #5, #6 | HIGH |
| Phase 3 | #9, #10, #11, #12 | MEDIUM |
| Phase 4 | #17 (done), #18 | ENHANCEMENT |
| Phase 5 | #13, #14, #16 | LOW |

**Already Implemented:** OpenRouter support