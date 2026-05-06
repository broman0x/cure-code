# ADR-001: Enhanced Session Persistence with Rich Metadata

**Date:** 2026-05-06  
**Status:** Proposed  
**Author:** Ev3lynx

---

## 1. Summary

Enhance cure-code session JSON structure to include rich metadata (provider, model, tokens, timestamps) similar to Hermes/pi-mono architecture, while maintaining tool tracking capabilities.

---

## 2. Current State

### Existing Session Structure
```json
{
  "id": "session-1746523200",
  "timestamp": "2026-05-06T11:33:20Z",
  "history": [...],
  "tasks": [...],
  "work_dir": "/home/ev3lynx/pysockserver"
}
```

### Issues
- No AI provider/model information stored
- No token usage tracking
- No session start/end timestamps
- Per-message timestamps missing
- Session ID format not human-readable

---

## 3. Proposed Enhanced Structure

```json
{
  "session_id": "session_20260506_143200_a4f2d1c3",
  "timestamp": "2026-05-06T14:32:00Z",
  "last_active": "2026-05-06T14:32:45Z",
  "history": [
    {
      "role": "user",
      "content": "...",
      "timestamp": "2026-05-06T14:32:00Z"
    },
    {
      "role": "assistant", 
      "content": "...",
      "timestamp": "2026-05-06T14:32:05Z",
      "tool_calls": [...]
    }
  ],
  "tasks": [...],
  "work_dir": "/home/ev3lynx/pysockserver",
  "metadata": {
    "provider": "ollama",
    "model": "granite4:latest",
    "start_time": "2026-05-06T14:32:00Z",
    "end_time": "2026-05-06T14:32:45Z",
    "total_tokens": 12450,
    "tool_call_count": 5,
    "version": "1.0.3"
  }
}
```

---

## 4. Field-by-Field Changes

| Current Field | New Field | Type | Description |
|--------------|-----------|------|-------------|
| `id` | `session_id` | string | Human-readable format |
| `timestamp` | `timestamp` | timestamp | ISO 8601 |
| - | `last_active` | timestamp | Last interaction time |
| `History[].role` | `History[].role` | string | Unchanged |
| `History[].content` | `History[].content` | string | Unchanged |
| - | `History[].timestamp` | timestamp | Per message time |
| - | `History[].tool_calls` | array | Already exists |
| `Tasks` | `Tasks` | array | Unchanged |
| `WorkDir` | `work_dir` | string | Unchanged |
| - | `metadata` | object | NEW - Rich metadata |
| - | `metadata.provider` | string | AI provider name |
| - | `metadata.model` | string | Model name |
| - | `metadata.start_time` | timestamp | Session start |
| - | `metadata.end_time` | timestamp | Session end |
| - | `metadata.total_tokens` | int | Token usage |
| - | `metadata.tool_call_count` | int | Tools invoked |
| - | `metadata.version` | string | cure-code version |

---

## 5. Session ID Format

**Current:**
```
session-{unix_timestamp}
session-1746523200
```

**New (Hermes-style):**
```
session_{YYYYMMDD}_{HHMMSS}_{short_id}
session_20260506_143200_a4f2d1c3
```

---

## 6. Rationale

1. **Rich Metadata**: Match Hermes/pi-mono for provider/model tracking
2. **Token Usage**: Enable usage analytics and billing tracking
3. **Human-readable ID**: Easier session identification
4. **Per-message Timestamps**: Timeline reconstruction
5. **Tool Tracking**: Already exists, preserve it

---

## 7. Implementation Plan

### Phase 1: Structure Update
- [ ] Update `SessionData` struct in `session.go`
- [ ] Add `Metadata` struct
- [ ] Change session ID format to Hermes-style
- [ ] Add per-message timestamps

### Phase 2: Metadata Collection
- [ ] Capture provider/model at session start
- [ ] Track token usage from AI responses
- [ ] Record tool_call_count
- [ ] Add version info

### Phase 3: Backward Compatibility
- [ ] Load legacy session format
- [ ] Migrate old sessions on access

---

## 8. Related Systems

| System | Session Format | Reference |
|--------|--------------|-----------|
| Hermes | `session_{date}_{time}_{id}.json` | Headquarters doc |
| pi-mono | JSON with metadata | github.com/badlogic/pi-mono |
| cure-code | Current JSON | This ADR |

---

## 9. Open Questions

- [ ] Should we auto-migrate legacy sessions?
- [ ] Maximum session size limit?
- [ ] Session expiry/TTL policy?

---

**End of ADR**