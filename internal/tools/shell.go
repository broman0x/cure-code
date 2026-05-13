package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type ShellTool struct {
	workDir string
}

const (
	PermissionModeDefault     = "default"
	PermissionModeAcceptEdits = "accept_edits"
	PermissionModeBypass      = "bypass"
)

const (
	SandboxOff                  = "off"
	SandboxWorkspaceWrite       = "workspace-write"
	SandboxWorkspaceWriteNoNet  = "workspace-write-no-network"
	SandboxReadOnly             = "read-only"
)

func NewShellTool(workDir string) *ShellTool {
	return &ShellTool{workDir: workDir}
}

func (t *ShellTool) Name() string { return "run_command" }

func (t *ShellTool) Description() string {
	return `Execute a shell command in the user's terminal and return the output.
Use this for running tests, installing dependencies, checking git status, building projects, etc.
Commands are executed in the project's working directory.
The command will be terminated if it runs longer than 60 seconds.
For long-running commands, either pass background=true or suffix the command with '&'.`
}

func (t *ShellTool) ParameterSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "The shell command to execute.",
			},
			"background": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to run the command in the background. Use true for long-running processes like servers or watchers.",
			},
		},
		"required": []string{"command"},
	}
}

func (t *ShellTool) validateCommand(command string) string {
	// [EN] Basic security parser to flag dangerous shell patterns.
	// [ID] Parser keamanan dasar untuk menandai pola shell yang berbahaya.
	dangerousPatterns := []struct {
		pattern string
		reason  string
	}{
		{";", "Command chaining (;) detected"},
		{"&&", "Logical AND (&&) detected"},
		{"||", "Logical OR (||) detected"},
		{"|", "Piping (|) detected"},
		{"`", "Backticks (`) detected"},
		{"$(", "Command substitution $() detected"},
		{">", "Output redirection (>) detected"},
		{"<", "Input redirection (<) detected"},
	}

	var warnings []string
	for _, p := range dangerousPatterns {
		if strings.Contains(command, p.pattern) {
			warnings = append(warnings, p.reason)
		}
	}

	if len(warnings) > 0 {
		return "Security Warning: " + strings.Join(warnings, ", ")
	}
	return ""
}

func (t *ShellTool) NeedsConfirmation(params map[string]interface{}) bool {
	return true
}

func (t *ShellTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	command, ok := getStringParam(params, "command")
	if !ok || command == "" {
		return &ToolResult{Content: "Error: command is required", IsError: true}, nil
	}

	requestedBackground, _ := params["background"].(bool)
	command = strings.TrimSpace(command)
	command, impliedBackground := normalizeBackgroundCommand(command)
	isBackground := requestedBackground || impliedBackground
	permissionMode := getPermissionMode(ctx)
	allowedPrefixes := getAllowedCommandPrefixes(ctx)
	sandboxProfile := getSandboxProfile(ctx)

	if permissionMode == PermissionModeAcceptEdits && isLikelyNonEditCommand(command) && !isAllowedByPrefix(command, allowedPrefixes) {
		return &ToolResult{
			Content: "Permission mode 'accept_edits' blocks non-edit shell commands unless they match an allowed prefix.",
			Display: "[X] Blocked by permission mode: accept_edits",
			IsError: true,
		}, nil
	}

	if errText := enforceSandboxProfile(command, sandboxProfile, t.workDir); errText != "" {
		return &ToolResult{
			Content: errText,
			Display: fmt.Sprintf("[X] Blocked by sandbox profile '%s'", sandboxProfile),
			IsError: true,
		}, nil
	}

	var cmd *exec.Cmd
	var execCtx context.Context
	var cancel context.CancelFunc

	if isBackground {
		execCtx = ctx
	} else {
		execCtx, cancel = context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
	}

	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(execCtx, "cmd", "/C", command)
	} else {
		cmd = exec.CommandContext(execCtx, "sh", "-c", command)
	}

	cmd.Dir = t.workDir

	if isBackground {
		err := cmd.Start()
		if err != nil {
			return &ToolResult{Content: fmt.Sprintf("Error starting background command: %v", err), IsError: true}, nil
		}
		backgroundNote := ""
		if impliedBackground && !requestedBackground {
			backgroundNote = " (auto-detected from trailing &)"
		}
		return &ToolResult{
			Content:       fmt.Sprintf("Background process started (PID %d)%s.", cmd.Process.Pid, backgroundNote),
			Display:       fmt.Sprintf("[X] Started background: %s", truncateStr(command, 60)),
			BackgroundCmd: cmd,
		}, nil
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	var output strings.Builder
	if stdout.Len() > 0 {
		output.WriteString(stdout.String())
	}
	if stderr.Len() > 0 {
		if output.Len() > 0 {
			output.WriteString("\n")
		}
		output.WriteString(stderr.String())
	}

	result := output.String()
	const maxLen = 30000
	if len(result) > maxLen {
		result = result[:maxLen] + "\n... (output truncated)"
	}

	if err != nil {
		exitErr, isExitError := err.(*exec.ExitError)
		if isExitError {
			hint := ""
			if exitErr.ExitCode() == -1 && strings.Contains(command, "pkill -f") {
				hint = "\n\nHint: 'pkill -f' can match its own invocation. Try a safer pattern, e.g. pkill -f \"[n]ode index.js\"."
			}
			return &ToolResult{
				Content: fmt.Sprintf("Command exited with code %d.\n\n%s%s", exitErr.ExitCode(), result, hint),
				Display: fmt.Sprintf("[!] Command exited with code %d: %s", exitErr.ExitCode(), truncateStr(command, 60)),
				IsError: true,
			}, nil
		}
		if execCtx.Err() == context.DeadlineExceeded {
			return &ToolResult{
				Content: fmt.Sprintf("Command timed out after 60 seconds.\n\n%s", result),
				Display: fmt.Sprintf("[T] Command timed out: %s", truncateStr(command, 60)),
				IsError: true,
			}, nil
		}
		return &ToolResult{
			Content: fmt.Sprintf("Error executing command: %v\n\n%s", err, result),
			IsError: true,
		}, nil
	}

	if result == "" {
		result = "(no output)"
	}

	display := fmt.Sprintf("[X] Ran: %s", truncateStr(command, 60))
	if warning := t.validateCommand(command); warning != "" {
		display = warning + "\n  " + display
		result = "WARNING: " + warning + "\n\n" + result
	}

	return &ToolResult{
		Content: result,
		Display: display,
	}, nil
}

func normalizeBackgroundCommand(command string) (string, bool) {
	trimmed := strings.TrimSpace(command)
	if !strings.HasSuffix(trimmed, "&") || strings.HasSuffix(trimmed, "&&") {
		return trimmed, false
	}
	withoutAmp := strings.TrimSpace(strings.TrimSuffix(trimmed, "&"))
	if withoutAmp == "" {
		return trimmed, false
	}
	return withoutAmp, true
}

func getPermissionMode(ctx context.Context) string {
	if v, ok := ctx.Value(PermissionModeKey).(string); ok && v != "" {
		return v
	}
	return PermissionModeDefault
}

func getAllowedCommandPrefixes(ctx context.Context) []string {
	if v, ok := ctx.Value(AllowedCommandPrefixesKey).([]string); ok {
		return v
	}
	return nil
}

func getSandboxProfile(ctx context.Context) string {
	if v, ok := ctx.Value(ShellSandboxProfileKey).(string); ok && v != "" {
		return v
	}
	return SandboxOff
}

func isAllowedByPrefix(command string, prefixes []string) bool {
	trimmed := strings.TrimSpace(command)
	for _, p := range prefixes {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.HasPrefix(trimmed, p) {
			return true
		}
	}
	return false
}

func isLikelyNonEditCommand(command string) bool {
	c := strings.TrimSpace(strings.ToLower(command))
	if c == "" {
		return false
	}

	editPrefixes := []string{
		"go test", "go fmt", "go vet", "npm test", "npm run test", "pnpm test", "yarn test",
		"pytest", "cargo test", "cargo fmt", "cargo clippy", "make test", "make fmt",
	}
	for _, p := range editPrefixes {
		if strings.HasPrefix(c, p) {
			return false
		}
	}
	return true
}

func enforceSandboxProfile(command string, profile string, workDir string) string {
	switch profile {
	case SandboxOff:
		return ""
	case SandboxReadOnly:
		if isLikelyWriteCommand(command) {
			return "Sandbox(read-only) blocked this command because it appears to modify files."
		}
	case SandboxWorkspaceWrite:
		if p := firstDangerousAbsolutePath(command, workDir); p != "" {
			return fmt.Sprintf("Sandbox(workspace-write) blocked write-like path outside workspace: %s", p)
		}
	case SandboxWorkspaceWriteNoNet:
		if looksLikeNetworkCommand(command) {
			return "Sandbox(workspace-write-no-network) blocked this command because it appears to use network access."
		}
		if p := firstDangerousAbsolutePath(command, workDir); p != "" {
			return fmt.Sprintf("Sandbox(workspace-write-no-network) blocked write-like path outside workspace: %s", p)
		}
	default:
		return fmt.Sprintf("Unknown sandbox profile: %s", profile)
	}
	return ""
}

func isLikelyWriteCommand(command string) bool {
	c := " " + strings.ToLower(command) + " "
	writeSignals := []string{
		" rm ", " mv ", " cp ", " touch ", " mkdir ", " rmdir ", " chmod ", " chown ",
		" sed -i", " perl -i", " tee ", " truncate ", " dd ",
	}
	for _, s := range writeSignals {
		if strings.Contains(c, s) {
			return true
		}
	}
	return strings.Contains(command, ">") || strings.Contains(command, ">>")
}

func looksLikeNetworkCommand(command string) bool {
	c := " " + strings.ToLower(command) + " "
	netSignals := []string{" curl ", " wget ", " ssh ", " scp ", " rsync ", " nc ", " netcat "}
	for _, s := range netSignals {
		if strings.Contains(c, s) {
			return true
		}
	}
	return strings.Contains(c, " http://") || strings.Contains(c, " https://")
}

func firstDangerousAbsolutePath(command string, workDir string) string {
	if !isLikelyWriteCommand(command) {
		return ""
	}

	cleanWorkDir := filepath.Clean(workDir)
	for _, tok := range strings.Fields(command) {
		if !strings.HasPrefix(tok, "/") {
			continue
		}
		p := strings.Trim(tok, "\"'`,;")
		if p == "/" || p == "" {
			return p
		}
		if strings.HasPrefix(p, "/tmp/") || p == "/tmp" {
			continue
		}
		clean := filepath.Clean(p)
		if !strings.HasPrefix(clean, cleanWorkDir+"/") && clean != cleanWorkDir {
			return clean
		}
	}
	return ""
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
