package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/broman0x/cure-code/internal/config"
	"github.com/broman0x/cure-code/internal/tools"
	"github.com/broman0x/cure-code/internal/ui"
	"github.com/broman0x/cure-code/internal/version"
	"github.com/broman0x/cure-code/internal/watcher"
	"github.com/fatih/color"
)

// [EN] FunctionCallingProvider defines the interface for AI models that support tool calling.
// [ID] FunctionCallingProvider mendefinisikan antarmuka untuk model AI yang mendukung pemanggilan tool.
type FunctionCallingProvider interface {
	// [EN] Name returns the display name of the provider and model.
	// [ID] Name mengembalikan nama tampilan provider dan model.
	Name() string

	// [EN] SendWithTools sends a prompt and message history to the AI with a set of tool definitions.
	// [ID] SendWithTools mengirim prompt dan riwayat pesan ke AI dengan kumpulan definisi tool.
	SendWithTools(ctx context.Context, systemPrompt string, messages []Message, toolDefs []tools.ToolDefinition) (*Response, error)

	// [EN] SupportsTools returns true if the provider supports tool calling.
	// [ID] SupportsTools mengembalikan true jika provider mendukung pemanggilan tool.
	SupportsTools() bool
}

type StreamingProvider interface {
	FunctionCallingProvider

	SendWithToolsStream(ctx context.Context, systemPrompt string, messages []Message, toolDefs []tools.ToolDefinition) (<-chan StreamEvent, error)

	SupportsStreaming() bool
}

// [EN] Agent is the core entity that manages the conversation state, tool execution, and AI interactions.
// [ID] Agent adalah entitas inti yang mengelola status percakapan, eksekusi tool, dan interaksi AI.
type Agent struct {
	Provider     FunctionCallingProvider
	Tools        *tools.ToolRegistry
	History      []Message
	SystemPrompt string
	WorkDir      string
	Scanner      *bufio.Scanner
	MaxTurns     int
	YOLO         bool

	AlwaysAllow map[string]bool

	ToolCallCount   int
	Usage           SessionUsage
	Tasks           []Task
	ProcMgr         *ProcessManager
	RecentToolCalls []string
	Skills          *SkillRegistry
	Planning        bool
	InterviewMode   bool
	
	FileCacheMu     sync.RWMutex
	FileCache       map[string]string
	FileCacheOrder  []string

	Watcher         *watcher.Watcher

	CompactThreshold int
	RecentSymbols    []string
	Intel            *IntelligenceService
	RepeatTracker    map[string]int

	PermissionMode       string
	AllowedCommandPrefix []string
	ShellSandboxProfile  string

	ModifiedFiles        map[string]bool
	ValidationRan        bool
	VerificationPrompted bool

	CleanupFuncs []func() error
}

// [EN] NewAgent initializes a new AI coding agent with the given provider and working directory.
// [ID] NewAgent menginisialisasi agen coding AI baru dengan provider dan direktori kerja yang diberikan.
func NewAgent(provider FunctionCallingProvider, workDir string) *Agent {
	wsCtx := DetectWorkspace(workDir)
	registry := tools.NewDefaultRegistry(workDir)
	skills := NewSkillRegistry()
	skills.LoadBuiltin()
	skills.LoadFromDir(workDir)

	a := &Agent{
		Provider:     provider,
		Tools:        registry,
		History:      make([]Message, 0),
		WorkDir:      workDir,
		Scanner:      bufio.NewScanner(os.Stdin),
		MaxTurns:     25,
		AlwaysAllow:  make(map[string]bool),
		YOLO:         false,
		ProcMgr:      NewProcessManager(),
		Skills:       skills,
		FileCache:    make(map[string]string),
		FileCacheOrder: make([]string, 0),
		CompactThreshold: 15000, // [EN] Default threshold for compaction | [ID] Ambang batas default untuk pemadatan
		RecentSymbols:    make([]string, 0),
		Intel:            NewIntelligenceService(workDir),
		RepeatTracker:    make(map[string]int),
		PermissionMode:   tools.PermissionModeDefault,
		ShellSandboxProfile: tools.SandboxOff,
		ModifiedFiles:    make(map[string]bool),
	}

	// [EN] Register delegation tool with callback to agent
	// [ID] Daftarkan tool delegasi dengan callback ke agen
	registry.Register(&tools.DelegateTool{
		ProcessPromptFunc: a.SpawnSubAgent,
	})

	w, err := watcher.NewWatcher(workDir, func(path string, content string) {
		a.FileCacheMu.Lock()
		defer a.FileCacheMu.Unlock()
		if _, exists := a.FileCache[path]; exists {
			a.FileCache[path] = content
			color.HiCyan("\n  [Watcher] Detected external changes to %s. Context updated.", path)
		}
	})
	if err == nil {
		a.Watcher = w
		a.Watcher.Start()
		a.CleanupFuncs = append(a.CleanupFuncs, a.Watcher.Close)
	}

	a.SystemPrompt = BuildSystemPrompt(wsCtx, skills.List(), a.getFileCacheCopy(), nil, nil, a.InterviewMode)
	return a
}

func (a *Agent) getFileCacheCopy() map[string]string {
	a.FileCacheMu.RLock()
	defer a.FileCacheMu.RUnlock()
	copy := make(map[string]string)
	for k, v := range a.FileCache {
		copy[k] = v
	}
	return copy
}

// [EN] ProcessPrompt takes a user input, resolves mentions, and starts the agentic loop.
// [ID] ProcessPrompt menerima input pengguna, menyelesaikan mention, dan memulai loop agentic.
func (a *Agent) ProcessPrompt(ctx context.Context, userPrompt string) error {
	defer a.saveState()
	a.ToolCallCount = 0
	a.RecentToolCalls = nil
	a.ModifiedFiles = make(map[string]bool)
	a.ValidationRan = false
	a.VerificationPrompted = false

	processedPrompt := a.ResolveMentions(userPrompt)

	a.History = append(a.History, Message{
		Role:    "user",
		Content: processedPrompt,
	})

	wsCtx := DetectWorkspace(a.WorkDir)
	suggested := a.Intel.SuggestContext(userPrompt, a.History)
	a.SystemPrompt = BuildSystemPrompt(wsCtx, a.Skills.List(), a.getFileCacheCopy(), a.RecentSymbols, suggested, a.InterviewMode)
	a.SystemPrompt += "\n\n## RUNTIME SAFETY POLICY\n" + a.runtimePolicyPrompt()

	// [EN] Check and compact history if needed
	// [ID] Cek dan padatkan riwayat jika diperlukan
	if err := a.checkAndCompact(ctx); err != nil {
		color.Yellow("  [!] Compaction failed: %v", err)
	}

	a.renderTaskList()
	toolDefs := a.Tools.CoreDefinitions()

	sp, canStream := a.Provider.(StreamingProvider)
	if canStream && sp.SupportsStreaming() {
		return a.processWithStreaming(ctx, sp, toolDefs)
	}
	return a.processWithBatch(ctx, toolDefs)
}

func (a *Agent) processWithStreaming(ctx context.Context, sp StreamingProvider, toolDefs []tools.ToolDefinition) error {
	for turn := 0; turn < a.MaxTurns; turn++ {
		spinner := ui.NewSpinner("Thinking")
		spinner.Start()

		eventCh, err := sp.SendWithToolsStream(ctx, a.SystemPrompt, a.History, toolDefs)
		if err != nil {
			spinner.Stop()
			color.Red("\n  Error: %v\n", err)
			return err
		}

		var contentBuf strings.Builder
		var thoughtBuf strings.Builder
		var collectedToolCalls []ToolCall
		var usage *UsageStats
		var finishReason string

		firstText := true
		firstThought := true
		spinnerStopped := false
		startTime := time.Now()

		for event := range eventCh {
			switch event.Type {
			case StreamThought:
				if !spinnerStopped {
					spinner.Stop()
					spinnerStopped = true
				}
				if firstThought {
					cThought := color.New(color.FgHiBlack, color.Bold).SprintFunc()
					fmt.Printf("\n  %s\n", cThought("┌─ THOUGHT"))
					firstThought = false
				}

				lines := strings.Split(event.Thought, "\n")
				for _, line := range lines {
					if line == "" && thoughtBuf.Len() > 0 {
						fmt.Printf("  %s\n", color.HiBlackString("│"))
					} else if line != "" {
						fmt.Printf("  %s %s\n", color.HiBlackString("│"), color.HiBlackString(line))
					}
				}
				thoughtBuf.WriteString(event.Thought)

			case StreamText:
				if !spinnerStopped {
					spinner.Stop()
					spinnerStopped = true
				}
				if firstText {
					if !firstThought {
						fmt.Printf("  %s\n", color.HiBlackString("└─"))
						fmt.Println()
					}
					cAI := color.New(color.FgHiCyan, color.Bold).SprintFunc()
					fmt.Printf("  %s %s\n\n  ", cAI("✦"), cAI("CURE CODE"))
					firstText = false
				}
				fmt.Print(event.Text)
				contentBuf.WriteString(event.Text)

			case StreamToolCall:
				if !spinnerStopped {
					spinner.Stop()
					spinnerStopped = true
				}
				if event.ToolCall != nil {
					collectedToolCalls = append(collectedToolCalls, *event.ToolCall)
				}

			case StreamDone:
				if !spinnerStopped {
					spinner.Stop()
					spinnerStopped = true
				}
				usage = event.Usage
				finishReason = event.FinishReason

			case StreamError:
				spinner.Stop()
				if !strings.Contains(event.Error.Error(), "context canceled") {
					color.Red("\n  Error: %v\n", event.Error)
				}
				return event.Error
			}
		}

		content := contentBuf.String()
		elapsed := time.Since(startTime)

		if content != "" {
			fmt.Println()
		}

		if usage != nil {
			a.Usage.Add(usage)
			a.showUsageInline(usage, elapsed)
		}

		if content != "" || len(collectedToolCalls) > 0 || thoughtBuf.Len() > 0 {
			a.History = append(a.History, Message{
				Role:      "assistant",
				Content:   content,
				Thought:   thoughtBuf.String(),
				ToolCalls: collectedToolCalls,
			})
		}

		if len(collectedToolCalls) == 0 {
			_ = finishReason
			if a.maybeInjectVerificationPrompt() {
				continue
			}
			return nil
		}

		for _, tc := range collectedToolCalls {
			a.ToolCallCount++
			result, err := a.executeToolCall(ctx, tc)
			if err != nil {
				a.History = append(a.History, Message{
					Role:       "tool",
					ToolCallID: tc.ID,
					Name:       tc.Name,
					Content:    "Tool execution cancelled by user.",
				})
				continue
			}

			a.History = append(a.History, Message{
				Role:       "tool",
				ToolCallID: tc.ID,
				Name:       tc.Name,
				Content:    result.Content,
			})
			a.checkLoopDetection(tc)
			if tc.Name == "write_todos" {
				a.syncPlanMD()
			}
		}
		
		// [EN] Save state for web dashboard
		// [ID] Simpan status untuk dashboard web
		a.saveState()
	}

	color.Yellow("\n  [!] Agent reached maximum turn limit (%d). Stopping.", a.MaxTurns)
	return nil
}

func (a *Agent) processWithBatch(ctx context.Context, toolDefs []tools.ToolDefinition) error {
	mdRenderer := ui.NewMarkdownRenderer()

	for turn := 0; turn < a.MaxTurns; turn++ {
		spinner := ui.NewSpinner("Thinking")
		spinner.Start()
		startTime := time.Now()

		resp, err := a.Provider.SendWithTools(ctx, a.SystemPrompt, a.History, toolDefs)
		spinner.Stop()
		elapsed := time.Since(startTime)
		if err != nil {
			if !strings.Contains(err.Error(), "context canceled") {
				color.Red("\n  Error: %v\n", err)
			}
			return err
		}

		if resp.Usage != nil {
			a.Usage.Add(resp.Usage)
			a.showUsageInline(resp.Usage, elapsed)
		}

		if resp.Content != "" {
			a.History = append(a.History, Message{
				Role:      "assistant",
				Content:   resp.Content,
				ToolCalls: resp.ToolCalls,
			})

			cAI := color.New(color.FgHiCyan, color.Bold).SprintFunc()
			cBrand := color.New(color.FgCyan).SprintFunc()
			fmt.Printf("\n  %s %s\n\n", cAI("✦"), cBrand("CuRe Code"))

			rendered := mdRenderer.Render(resp.Content)
			fmt.Println(rendered)
			fmt.Println()
		} else if len(resp.ToolCalls) > 0 {
			a.History = append(a.History, Message{
				Role:      "assistant",
				ToolCalls: resp.ToolCalls,
			})
		}

		if len(resp.ToolCalls) == 0 {
			if a.maybeInjectVerificationPrompt() {
				continue
			}
			return nil
		}

		for _, tc := range resp.ToolCalls {
			a.ToolCallCount++
			result, err := a.executeToolCall(ctx, tc)
			if err != nil {
				a.History = append(a.History, Message{
					Role:       "tool",
					ToolCallID: tc.ID,
					Name:       tc.Name,
					Content:    "Tool execution cancelled by user.",
				})
				continue
			}

			a.History = append(a.History, Message{
				Role:       "tool",
				ToolCallID: tc.ID,
				Name:       tc.Name,
				Content:    result.Content,
			})
		}
	}

	color.Yellow("\n  [!] Agent reached maximum turn limit (%d). Stopping.", a.MaxTurns)
	return nil
}

func (a *Agent) showUsageInline(usage *UsageStats, elapsed time.Duration) {
	cDim := color.New(color.FgHiBlack).SprintFunc()
	fmt.Printf("  %s\n", cDim(fmt.Sprintf(
		"tokens: %d in / %d out | %.1fs",
		usage.InputTokens, usage.OutputTokens, elapsed.Seconds(),
	)))
}

func (a *Agent) executeToolCall(ctx context.Context, tc ToolCall) (*tools.ToolResult, error) {

	argsJSON, _ := json.Marshal(tc.Args)
	callKey := fmt.Sprintf("%s:%s", tc.Name, string(argsJSON))

	// [EN] Pass planning state via context
	// [ID] Teruskan status planning melalui context
	ctx = context.WithValue(ctx, tools.PlanningKey, a.Planning)
	ctx = context.WithValue(ctx, tools.PermissionModeKey, a.PermissionMode)
	ctx = context.WithValue(ctx, tools.AllowedCommandPrefixesKey, a.AllowedCommandPrefix)
	ctx = context.WithValue(ctx, tools.ShellSandboxProfileKey, a.ShellSandboxProfile)

	a.RecentToolCalls = append(a.RecentToolCalls, callKey)
	if len(a.RecentToolCalls) > 5 {
		a.RecentToolCalls = a.RecentToolCalls[1:]
	}

	count := 0
	for _, k := range a.RecentToolCalls {
		if k == callKey {
			count++
		}
	}

	if count >= 3 {
		return &tools.ToolResult{
			Content: fmt.Sprintf("Error: You are in a tool-calling loop. You have already called '%s' with these exact arguments %d times. DO NOT repeat the same call. Try a different approach (e.g., read a larger range, search for something else, or explain the issue).", tc.Name, count),
			IsError: true,
		}, nil
	}

	tool, ok := a.Tools.Get(tc.Name)
	if !ok {
		return &tools.ToolResult{
			Content: fmt.Sprintf("Error: unknown tool '%s'", tc.Name),
			IsError: true,
		}, nil
	}

	if tool.NeedsConfirmation(tc.Args) && !a.shouldAutoAllow(tc) && !a.AlwaysAllow[tc.Name] {
		approved, always := a.confirmTool(tc)
		if !approved {
			color.Yellow("  [X] Cancelled: %s", tc.Name)
			return nil, fmt.Errorf("cancelled by user")
		}
		if always {
			a.AlwaysAllow[tc.Name] = true
		}
	} else if tool.NeedsConfirmation(tc.Args) {
		color.HiBlack("  [AutoAllow] %s approved by mode/rule", tc.Name)
	}

	cDim := color.New(color.FgHiBlack).SprintFunc()
	cTool := color.New(color.FgHiCyan, color.Bold).SprintFunc()
	toolDisplay := a.formatToolHeader(tc)
	fmt.Printf("\n  %s %s %s\n", color.HiCyanString("┌─▶"), cTool(tc.Name), cDim(toolDisplay))

	result, err := tool.Execute(ctx, tc.Args)
	if err != nil {
		fmt.Printf("  %s %s\n", color.RedString("✖"), color.RedString(err.Error()))
		return nil, err
	}

	if result.IsError {
		fmt.Printf("  %s %s\n", color.HiRedString("└─✖"), color.RedString(result.Display))
	} else if result.Display != "" {
		fmt.Printf("  %s %s\n", color.HiBlackString("└─ "), color.HiBlackString(result.Display))
	}

	if result.BackgroundCmd != nil {
		if cmd, ok := result.BackgroundCmd.(*exec.Cmd); ok {
			command, _ := tc.Args["command"].(string)
			pid := a.ProcMgr.Add(command, cmd)
			fmt.Printf("  [OK] Process tracked as ID %d\n", pid)
		}
	}

	if len(result.FilesModified) > 0 {
		for _, p := range result.FilesModified {
			a.ModifiedFiles[p] = true
		}
	}

	if tc.Name == "run_command" {
		if command, ok := tc.Args["command"].(string); ok && looksLikeValidationCommand(command) {
			a.ValidationRan = true
		}
	}

	if tc.Name == "write_todos" && !result.IsError {
		if todosRaw, ok := result.Metadata["todos"]; ok {
			data, _ := json.Marshal(todosRaw)
			var tasks []Task
			if err := json.Unmarshal(data, &tasks); err == nil {
				a.Tasks = tasks
				a.renderTaskList()
				a.syncPlanMD() // [EN] Sync to PLAN.md | [ID] Sinkronisasi ke PLAN.md
			}
		}
	}

	if tc.Name == "search_symbol" && !result.IsError {
		if symbols, ok := result.Metadata["symbols"].([]string); ok {
			a.updateRecentSymbols(symbols)
		}
	}

	// [EN] Capture file contents for Smart Context Re-injection
	// [ID] Tangkap konten file untuk Smart Context Re-injection
	if (tc.Name == "read_file" || tc.Name == "edit_file") && !result.IsError {
		if path, ok := tc.Args["file_path"].(string); ok {
			if tc.Name == "read_file" {
				a.updateFileCache(path, result.Content)
			}
		}
	}

	if tc.Name == "enter_plan_mode" && !result.IsError {
		a.Planning = true
	}
	if tc.Name == "exit_plan_mode" && !result.IsError {
		a.Planning = false
	}

	return result, nil
}

func (a *Agent) renderTaskList() {
	if len(a.Tasks) == 0 {
		return
	}

	cTitle := color.New(color.FgHiCyan, color.Bold).SprintFunc()
	fmt.Println()
	fmt.Printf("  %s\n", cTitle("┌─ TASK PLAN"))
	for _, t := range a.Tasks {
		statusIcon := "[ ]"
		switch t.Status {
		case "in_progress":
			statusIcon = color.HiCyanString("[>]")
		case "completed":
			statusIcon = color.GreenString("[x]")
		case "cancelled":
			statusIcon = color.HiBlackString("[-]")
		case "blocked":
			statusIcon = color.RedString("!")
		}

		msg := fmt.Sprintf("  %s %s %s", color.HiCyanString("│"), statusIcon, t.Description)
		if t.Status == "completed" {
			color.HiBlack(msg)
		} else {
			fmt.Println(msg)
		}
	}
	fmt.Printf("  %s\n", cTitle("└──────────────────────────────────"))
	fmt.Println()
}

func (a *Agent) formatToolHeader(tc ToolCall) string {
	switch tc.Name {
	case "read_file", "write_file", "edit_file":
		if path, ok := tc.Args["file_path"].(string); ok {
			return path
		}
	case "run_command":
		if cmd, ok := tc.Args["command"].(string); ok {
			if len(cmd) > 60 {
				return cmd[:60] + "..."
			}
			return cmd
		}
	case "list_directory":
		if path, ok := tc.Args["path"].(string); ok {
			return path
		}
	case "grep_search":
		if pattern, ok := tc.Args["pattern"].(string); ok {
			return fmt.Sprintf("'%s'", pattern)
		}
	case "glob":
		if pattern, ok := tc.Args["pattern"].(string); ok {
			return pattern
		}
	}
	return ""
}

func (a *Agent) confirmTool(tc ToolCall) (bool, bool) {
	if a.YOLO {
		color.HiRed("  [!] YOLO Mode: Executing tool automatically...")
		return true, false
	}
	cKey := color.New(color.FgCyan).SprintFunc()
	cVal := color.New(color.FgHiWhite).SprintFunc()

	fmt.Printf("\n  %s\n", color.HiYellowString("┌─ TOOL REQUEST"))
	fmt.Printf("  %s %s: %s\n", color.HiYellowString("│"), cKey("tool"), tc.Name)

	for key, val := range tc.Args {
		valStr := fmt.Sprintf("%v", val)

		if tc.Name == "edit_file" && key == "old_string" {
			lines := strings.Split(valStr, "\n")
			if len(lines) > 5 {
				valStr = strings.Join(lines[:5], "\n") + fmt.Sprintf("\n%s ... (+%d more lines)", color.HiYellowString("│"), len(lines)-5)
			}
			fmt.Printf("  %s %s:\n", color.HiYellowString("│"), cKey(key))
			for _, line := range strings.Split(valStr, "\n") {
				fmt.Printf("  %s   %s %s\n", color.HiYellowString("│"), color.RedString("-"), line)
			}
			continue
		}
		if tc.Name == "edit_file" && key == "new_string" {
			lines := strings.Split(valStr, "\n")
			if len(lines) > 5 {
				valStr = strings.Join(lines[:5], "\n") + fmt.Sprintf("\n%s ... (+%d more lines)", color.HiYellowString("│"), len(lines)-5)
			}
			fmt.Printf("  %s %s:\n", color.HiYellowString("│"), cKey(key))
			for _, line := range strings.Split(valStr, "\n") {
				fmt.Printf("  %s   %s %s\n", color.HiYellowString("│"), color.GreenString("+"), line)
			}
			continue
		}

		if len(valStr) > 120 {
			valStr = valStr[:120] + "..."
		}
		fmt.Printf("  %s %s: %s\n", color.HiYellowString("│"), cKey(key), cVal(valStr))
	}

	fmt.Printf("  %s\n", color.HiYellowString("└──────────────────────────────────"))
	fmt.Print("  Allow? [y]es / [n]o / [a]lways > ")

	if !a.Scanner.Scan() {
		return false, false
	}
	input := strings.ToLower(strings.TrimSpace(a.Scanner.Text()))

	switch input {
	case "y", "yes", "":
		return true, false
	case "a", "always":
		color.Green("  [OK] Always allowing %s for this session", tc.Name)
		return true, true
	default:
		return false, false
	}
}

func (a *Agent) shouldAutoAllow(tc ToolCall) bool {
	if a.YOLO || a.PermissionMode == tools.PermissionModeBypass {
		return true
	}

	if a.PermissionMode == tools.PermissionModeAcceptEdits {
		return tc.Name == "write_file" || tc.Name == "edit_file"
	}

	if tc.Name == "run_command" {
		cmd, _ := tc.Args["command"].(string)
		return commandMatchesAllowedPrefix(cmd, a.AllowedCommandPrefix)
	}

	return false
}

func (a *Agent) Close() error {
	var lastErr error
	for _, cleanup := range a.CleanupFuncs {
		if err := cleanup(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

func commandMatchesAllowedPrefix(command string, prefixes []string) bool {
	command = strings.TrimSpace(command)
	for _, p := range prefixes {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.HasPrefix(command, p) {
			return true
		}
	}
	return false
}

func (a *Agent) SetPermissionMode(mode string) error {
	switch mode {
	case tools.PermissionModeDefault, tools.PermissionModeAcceptEdits, tools.PermissionModeBypass:
		a.PermissionMode = mode
		return nil
	default:
		return fmt.Errorf("invalid permission mode: %s", mode)
	}
}

func (a *Agent) AddAllowedCommandPrefix(prefix string) {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return
	}
	a.AllowedCommandPrefix = append(a.AllowedCommandPrefix, prefix)
}

func (a *Agent) SetShellSandboxProfile(profile string) error {
	switch profile {
	case tools.SandboxOff, tools.SandboxWorkspaceWrite, tools.SandboxWorkspaceWriteNoNet, tools.SandboxReadOnly:
		a.ShellSandboxProfile = profile
		return nil
	default:
		return fmt.Errorf("invalid sandbox profile: %s", profile)
	}
}

func (a *Agent) runtimePolicyPrompt() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("- Permission mode: %s\n", a.PermissionMode))
	b.WriteString(fmt.Sprintf("- Shell sandbox profile: %s\n", a.ShellSandboxProfile))
	if len(a.AllowedCommandPrefix) == 0 {
		b.WriteString("- Allowed shell prefixes: (none)\n")
	} else {
		b.WriteString("- Allowed shell prefixes:\n")
		for _, p := range a.AllowedCommandPrefix {
			b.WriteString(fmt.Sprintf("  - %s\n", p))
		}
	}
	b.WriteString("- If run_command is blocked, adapt strategy instead of retrying the same command.\n")
	return b.String()
}

func (a *Agent) maybeInjectVerificationPrompt() bool {
	if a.VerificationPrompted {
		return false
	}
	if len(a.ModifiedFiles) == 0 || a.ValidationRan {
		return false
	}

	// In accept_edits mode without allowed shell prefixes, validation commands are likely blocked.
	if a.PermissionMode == tools.PermissionModeAcceptEdits && len(a.AllowedCommandPrefix) == 0 {
		return false
	}

	a.VerificationPrompted = true
	color.HiYellow("  [Verify] Files were modified but no validation command was detected. Requesting verification before final answer.")
	a.History = append(a.History, Message{
		Role: "system",
		Content: `Before you provide the final answer:
1. Run at least one validation command appropriate to this project (tests/build/lint).
2. If validation cannot be run, explain exactly why.
3. Then provide the final summary.`,
	})
	return true
}

func looksLikeValidationCommand(command string) bool {
	c := strings.ToLower(strings.TrimSpace(command))
	if c == "" {
		return false
	}
	hints := []string{
		"go test", "go vet", "go build",
		"npm test", "npm run test", "npm run build", "npm run lint",
		"pnpm test", "pnpm run test", "pnpm run build", "pnpm run lint",
		"yarn test", "yarn build", "yarn lint",
		"pytest", "cargo test", "cargo clippy", "cargo check",
		"make test", "make build", "make lint", "mvn test", "gradle test",
	}
	for _, h := range hints {
		if strings.HasPrefix(c, h) {
			return true
		}
	}
	return false
}

func (a *Agent) updateFileCache(path, content string) {
	a.FileCacheMu.Lock()
	defer a.FileCacheMu.Unlock()

	// [EN] Keep only the last 3 files in cache to save tokens
	// [ID] Hanya simpan 3 file terakhir di cache untuk menghemat token
	maxCache := 3

	if _, exists := a.FileCache[path]; !exists {
		a.FileCacheOrder = append(a.FileCacheOrder, path)
	}
	a.FileCache[path] = content

	if len(a.FileCacheOrder) > maxCache {
		oldest := a.FileCacheOrder[0]
		a.FileCacheOrder = a.FileCacheOrder[1:]
		delete(a.FileCache, oldest)
	}
}

func (a *Agent) ClearHistory() {
	a.History = make([]Message, 0)
	a.ToolCallCount = 0
	a.RecentSymbols = make([]string, 0)
}

func (a *Agent) updateRecentSymbols(newSymbols []string) {
	// [EN] Add new symbols to the beginning
	// [ID] Tambahkan simbol baru ke bagian awal
	a.RecentSymbols = append(newSymbols, a.RecentSymbols...)

	// [EN] Keep only unique top 20 symbols for spatial context
	// [ID] Simpan hanya 20 simbol teratas yang unik untuk konteks spasial
	seen := make(map[string]bool)
	unique := make([]string, 0)
	for _, s := range a.RecentSymbols {
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		unique = append(unique, s)
		if len(unique) >= 20 {
			break
		}
	}
	a.RecentSymbols = unique
}

func (a *Agent) UsageSummary() string {
	if a.Usage.RequestCount == 0 {
		return "No API requests made this session."
	}
	return fmt.Sprintf(
		"Session: %d requests | %d input tokens | %d output tokens | %d total tokens",
		a.Usage.RequestCount,
		a.Usage.TotalInputTokens,
		a.Usage.TotalOutputTokens,
		a.Usage.TotalTokens,
	)
}

func (a *Agent) SpawnSubAgent(ctx context.Context, task string) (string, error) {
	// [EN] Create a fresh agent for the sub-task
	// [ID] Buat agen baru untuk sub-tugas
	sub := NewAgent(a.Provider, a.WorkDir)
	sub.YOLO = a.YOLO
	sub.PermissionMode = a.PermissionMode
	sub.AllowedCommandPrefix = append([]string{}, a.AllowedCommandPrefix...)
	sub.ShellSandboxProfile = a.ShellSandboxProfile
	
	// [EN] Update system prompt for sub-agent
	sub.SystemPrompt = sub.SystemPrompt + "\n\n## SUB-AGENT ROLE\nYou are a specialized sub-agent spawned to help with a specific task. Focus ONLY on the requested task. When finished, provide a clear summary and then stop. Do not ask for further tasks."

	err := sub.ProcessPrompt(ctx, task)
	if err != nil {
		return "", err
	}

	// [EN] Get the last message from sub-agent history as the result
	if len(sub.History) > 0 {
		lastMsg := sub.History[len(sub.History)-1]
		return lastMsg.Content, nil
	}

	return "No summary provided by sub-agent.", nil
}
func (a *Agent) checkLoopDetection(tc ToolCall) {
	// [EN] Enhanced loop detection (RepeatTracker 2.0)
	// [ID] Deteksi loop tingkat lanjut (RepeatTracker 2.0)
	argKey := fmt.Sprintf("%s:%v", tc.Name, tc.Args)
	a.RepeatTracker[argKey]++

	count := a.RepeatTracker[argKey]
	if count >= 3 {
		color.HiRed("  [!] Loop detected for tool '%s'. Injecting self-correction prompt.", tc.Name)
		
		warning := fmt.Sprintf(`SYSTEM CRITICAL WARNING: You have executed '%s' with the EXACT same parameters %d times.
Your current strategy is failing to produce new information or progress. 

STRATEGY ADJUSTMENT REQUIRED:
1. STOP repeating this tool call.
2. THINK: Why is this not working? Are you looking in the wrong file? Is the search pattern too specific?
3. TRY: Use a different tool (e.g. if grep_search fails, try list_directory or read_file).
4. IF STUCK: Ask the user for help using 'ask_user'.

Failure to change strategy will lead to task termination.`, tc.Name, count)

		a.History = append(a.History, Message{
			Role:    "system",
			Content: warning,
		})
		
		// [EN] Exponential threshold to be even more aggressive if they keep trying
		// [ID] Ambang batas eksponensial untuk lebih agresif jika mereka terus mencoba
		a.RepeatTracker[argKey] = 0 
	}
}

func (a *Agent) syncPlanMD() {
	// [EN] Sync internal task list to PLAN.md for human visibility
	// [ID] Sinkronisasi daftar tugas internal ke PLAN.md untuk visibilitas manusia
	if len(a.Tasks) == 0 {
		return
	}

	var sb strings.Builder
	sb.WriteString("# CuRe Code: Project Implementation Plan\n\n")
	sb.WriteString("This file is automatically maintained by the AI Agent to track autonomous progress. Do not edit manually.\n\n")
	
	sb.WriteString("## Current Task Roadmap\n")
	for i, t := range a.Tasks {
		statusIcon := " "
		switch t.Status {
		case "completed":
			statusIcon = "x"
		case "in_progress":
			statusIcon = "/"
		case "blocked":
			statusIcon = "!"
		case "cancelled":
			statusIcon = "-"
		}
		sb.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, statusIcon, t.Description))
	}

	sb.WriteString(fmt.Sprintf("\n*Last updated: %s*\n", time.Now().Format(time.RFC1123)))

	path := filepath.Join(a.WorkDir, "PLAN.md")
	_ = os.WriteFile(path, []byte(sb.String()), 0644)
	color.HiBlack("  [Sync] Project plan synced to PLAN.md")
}

func (a *Agent) saveState() {
	// [EN] Save current agent state to a JSON file for web synchronization
	// [ID] Simpan status agen saat ini ke file JSON untuk sinkronisasi web
	state := struct {
		ProjectName   string       `json:"project_name"`
		RecentSymbols []string     `json:"recent_symbols"`
		Tasks         []Task       `json:"tasks"`
		HistoryCount  int          `json:"history_count"`
		LastTurnTime  time.Time    `json:"last_turn_time"`
		Usage         SessionUsage `json:"usage"`
		ToolCallCount int          `json:"tool_call_count"`
		IsPlanning    bool         `json:"is_planning"`
		AgentVersion  string       `json:"agent_version"`
	}{
		ProjectName:   filepath.Base(a.WorkDir),
		RecentSymbols: a.RecentSymbols,
		Tasks:         a.Tasks,
		HistoryCount:  len(a.History),
		LastTurnTime:  time.Now(),
		Usage:         a.Usage,
		ToolCallCount: a.ToolCallCount,
		IsPlanning:    a.Planning,
		AgentVersion:  version.Version, // Galileo
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return
	}

	// [EN] We save to ~/.config/curecode/state.json
	// [ID] Kita simpan ke ~/.config/curecode/state.json
	dir := filepath.Dir(config.GetConfigPath())
	os.MkdirAll(dir, 0755)
	
	path := filepath.Join(dir, "state.json")
	_ = os.WriteFile(path, data, 0644)
	a.syncPlanMD() // Also ensure PLAN.md is in sync
}
