package tools

import (
	"context"
	"fmt"
	"os/exec"
)

type GitInfoTool struct {
	workDir string
}

func NewGitInfoTool(workDir string) *GitInfoTool {
	return &GitInfoTool{workDir: workDir}
}

func (t *GitInfoTool) Name() string { return "get_git_info" }

func (t *GitInfoTool) Description() string {
	return "Get information about the current git repository, including status, branch, and recent diffs."
}

func (t *GitInfoTool) ParameterSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"status", "diff", "branch", "log"},
				"description": "The git action to perform. Defaults to 'status'.",
			},
		},
	}
}

func (t *GitInfoTool) NeedsConfirmation(params map[string]interface{}) bool {
	return false
}

func (t *GitInfoTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	action := "status"
	if a, ok := params["action"].(string); ok && a != "" {
		action = a
	}

	var args []string
	switch action {
	case "status":
		args = []string{"status", "--short"}
	case "diff":
		args = []string{"diff"}
	case "branch":
		args = []string{"branch", "-a"}
	case "log":
		args = []string{"log", "-n", "5", "--oneline"}
	default:
		return &ToolResult{Content: "Error: unknown action", IsError: true}, nil
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = t.workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &ToolResult{
			Content: fmt.Sprintf("Error executing git %s: %v\n%s", action, err, string(out)),
			IsError: true,
		}, nil
	}

	result := string(out)
	if result == "" {
		result = "No output (repository might be clean)."
	}

	return &ToolResult{
		Content: result,
		Display: fmt.Sprintf("[G] git %s", action),
	}, nil
}
