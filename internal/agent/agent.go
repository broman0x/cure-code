package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/broman0x/cure-code/internal/tools"
	"github.com/broman0x/cure-code/internal/ui"
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
	SendWithTools(systemPrompt string, messages []Message, toolDefs []tools.ToolDefinition) (*Response, error)

	// [EN] SupportsTools returns true if the provider supports tool calling.
	// [ID] SupportsTools mengembalikan true jika provider mendukung pemanggilan tool.
	SupportsTools() bool
}

type StreamingProvider interface {
	FunctionCallingProvider

	SendWithToolsStream(systemPrompt string, messages []Message, toolDefs []tools.ToolDefinition) (<-chan StreamEvent, error)

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
}

// [EN] NewAgent initializes a new AI coding agent with the given provider and working directory.
// [ID] NewAgent menginisialisasi agen coding AI baru dengan provider dan direktori kerja yang diberikan.
func NewAgent(provider FunctionCallingProvider, workDir string) *Agent {
	wsCtx := DetectWorkspace(workDir)
	registry := tools.NewDefaultRegistry(workDir)
	skills := NewSkillRegistry()
	skills.LoadBuiltin()
	skills.LoadFromDir(workDir)

	return &Agent{
		Provider:     provider,
		Tools:        registry,
		History:      make([]Message, 0),
		SystemPrompt: BuildSystemPrompt(wsCtx, skills.List()),
		WorkDir:      workDir,
		Scanner:      bufio.NewScanner(os.Stdin),
		MaxTurns:     25,
		AlwaysAllow:  make(map[string]bool),
		YOLO:         false,
		ProcMgr:      NewProcessManager(),
		Skills:       skills,
	}
}

// [EN] ProcessPrompt takes a user input, resolves mentions, and starts the agentic loop.
// [ID] ProcessPrompt menerima input pengguna, menyelesaikan mention, dan memulai loop agentic.
func (a *Agent) ProcessPrompt(ctx context.Context, userPrompt string) error {
	a.ToolCallCount = 0
	a.RecentToolCalls = nil

	processedPrompt := a.ResolveMentions(userPrompt)

	a.History = append(a.History, Message{
		Role:    "user",
		Content: processedPrompt,
	})

	wsCtx := DetectWorkspace(a.WorkDir)
	a.SystemPrompt = BuildSystemPrompt(wsCtx, a.Skills.List())

	a.renderTaskList()
	toolDefs := a.Tools.Definitions()

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

		eventCh, err := sp.SendWithToolsStream(a.SystemPrompt, a.History, toolDefs)
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
				color.Red("\n  Error: %v\n", event.Error)
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
		}
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

		resp, err := a.Provider.SendWithTools(a.SystemPrompt, a.History, toolDefs)
		spinner.Stop()
		elapsed := time.Since(startTime)
		if err != nil {
			color.Red("\n  Error: %v\n", err)
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

	if tool.NeedsConfirmation(tc.Args) && !a.AlwaysAllow[tc.Name] {
		approved, always := a.confirmTool(tc)
		if !approved {
			color.Yellow("  [X] Cancelled: %s", tc.Name)
			return nil, fmt.Errorf("cancelled by user")
		}
		if always {
			a.AlwaysAllow[tc.Name] = true
		}
	}

	cDim := color.New(color.FgHiBlack).SprintFunc()
	cTool := color.New(color.FgHiCyan, color.Bold).SprintFunc()
	toolDisplay := a.formatToolHeader(tc)
	fmt.Printf("\n  %s %s %s\n", color.HiCyanString("◆"), cTool(tc.Name), cDim(toolDisplay))

	result, err := tool.Execute(ctx, tc.Args)
	if err != nil {
		fmt.Printf("  %s %s\n", color.RedString("✖"), color.RedString(err.Error()))
		return nil, err
	}

	if result.IsError {
		fmt.Printf("  %s %s\n", color.RedString("✖"), color.RedString(result.Display))
	} else if result.Display != "" {
		fmt.Printf("  %s %s\n", color.HiGreenString("✔"), color.HiWhiteString(result.Display))
	}

	if result.BackgroundCmd != nil {
		if cmd, ok := result.BackgroundCmd.(*exec.Cmd); ok {
			command, _ := tc.Args["command"].(string)
			pid := a.ProcMgr.Add(command, cmd)
			fmt.Printf("  [OK] Process tracked as ID %d\n", pid)
		}
	}

	if tc.Name == "write_todos" && !result.IsError {
		var payload struct {
			Todos []Task `json:"todos"`
		}
		if err := json.Unmarshal([]byte(result.Content), &payload); err == nil {
			a.Tasks = payload.Todos
			a.renderTaskList()
		}
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
			statusIcon = color.HiCyanString("➤")
		case "completed":
			statusIcon = color.GreenString("✔")
		case "cancelled":
			statusIcon = color.HiBlackString("✖")
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

func (a *Agent) ClearHistory() {
	a.History = make([]Message, 0)
	a.ToolCallCount = 0
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
