# Changelog

All notable changes to this project will be documented in this file.

## [1.0.2] - 2026-05-05
### Added
- **Agentic Memory (Galileo)**: Implementation of predictive context compaction and high-fidelity summarization to handle long-running sessions.
- **Deep Code Intelligence**: Fuzzy keyword context discovery and persistent tracking of up to 20 unique code symbols.
- **Autonomous Orchestration**: Automated synchronization of internal tasks to `PLAN.md` for real-time progress visibility.
- **Sub-agent Delegation**: Functional `delegate_task` tool with support for passing initial context files to specialized sub-agents.
- **Updated Web Dashboard**: Complete overhaul of the web interface with a clean, minimalist "Technical White" and "Modern Dark" aesthetic. Removed cluttered sidebars in favor of focused, responsive data tables.
- **RepeatTracker 2.0**: Advanced loop detection and strategy-driven self-correction prompts to prevent stuck states.
- **Centralized Versioning**: Global release management via the `internal/version` package.
- **License Update**: Reverted project license to GNU GPLv3 to ensure copyleft protection.

### Changed
- **System Prompt V2**: Context-aware prompt construction with "Spatial Awareness" and "Memory blocks".
- **Tool Feedback**: Upgraded tool results to carry structured `Metadata` for internal agent reasoning.
- **Process Manager**: Improved lifecycle handling for background sub-agents and tool calls.

### Fixed
- **Token Bloat**: Eliminated context window exhaustion via proactive `autoCompact` logic.
- **Bilingual Consistency**: Full audit and standardization of English/Indonesian comments across the entire codebase.

## [1.0.1] - 2026-05-04
### Added
- Initial bilingual support (EN/ID).
- Multi-provider support (Gemini, OpenAI, Claude, NVIDIA, xAI, DeepSeek, Ollama).
- Background process manager (`/ps`).
- Basic REPL and One-shot modes.
- Self-installer for Windows, Linux, and macOS.
