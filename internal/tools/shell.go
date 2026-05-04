package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type ShellTool struct {
	workDir string
}

func NewShellTool(workDir string) *ShellTool {
	return &ShellTool{workDir: workDir}
}

func (t *ShellTool) Name() string { return "run_command" }

func (t *ShellTool) Description() string {
	return `Execute a shell command in the user's terminal and return the output.
Use this for running tests, installing dependencies, checking git status, building projects, etc.
Commands are executed in the project's working directory.
The command will be terminated if it runs longer than 60 seconds.`
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

func (t *ShellTool) NeedsConfirmation(params map[string]interface{}) bool {
	return true
}

func (t *ShellTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	command, ok := getStringParam(params, "command")
	if !ok || command == "" {
		return &ToolResult{Content: "Error: command is required", IsError: true}, nil
	}

	isBackground, _ := params["background"].(bool)

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
		return &ToolResult{
			Content:       fmt.Sprintf("Background process started (PID %d).", cmd.Process.Pid),
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
			return &ToolResult{
				Content: fmt.Sprintf("Command exited with code %d.\n\n%s", exitErr.ExitCode(), result),
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

	return &ToolResult{
		Content: result,
		Display: fmt.Sprintf("[X] Ran: %s", truncateStr(command, 60)),
	}, nil
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
