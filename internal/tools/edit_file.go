package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type EditFileTool struct {
	workDir string
}

func NewEditFileTool(workDir string) *EditFileTool {
	return &EditFileTool{workDir: workDir}
}

func (t *EditFileTool) Name() string { return "edit_file" }

func (t *EditFileTool) Description() string {
	return `Make edits to an existing file using a search and replace approach.
Specify the exact text to find (old_string) and what to replace it with (new_string).
The old_string must match exactly — including whitespace and indentation.
For creating new files, use the write_file tool instead.`
}

func (t *EditFileTool) ParameterSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "The path of the file to edit.",
			},
			"old_string": map[string]interface{}{
				"type":        "string",
				"description": "The exact text to search for in the file. Must match exactly including whitespace.",
			},
			"new_string": map[string]interface{}{
				"type":        "string",
				"description": "The text to replace old_string with. Use empty string to delete the matched text.",
			},
		},
		"required": []string{"file_path", "old_string", "new_string"},
	}
}

func (t *EditFileTool) NeedsConfirmation(params map[string]interface{}) bool {
	return true
}

func (t *EditFileTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	if planning, ok := ctx.Value(PlanningKey).(bool); ok && planning {
		return &ToolResult{
			Content: "Error: Cannot edit files while in Plan Mode. Please finish your exploration and design, then use 'exit_plan_mode' to apply changes.",
			IsError: true,
		}, nil
	}

	filePath, ok := getStringParam(params, "file_path")
	if !ok || filePath == "" {
		return &ToolResult{Content: "Error: file_path is required", IsError: true}, nil
	}

	oldString, ok := getStringParam(params, "old_string")
	if !ok {
		return &ToolResult{Content: "Error: old_string is required", IsError: true}, nil
	}

	newString, _ := getStringParam(params, "new_string")

	absPath := t.resolvePath(filePath)

	data, err := os.ReadFile(absPath)
	if err != nil {
		return &ToolResult{
			Content: fmt.Sprintf("Error reading file: %v", err),
			IsError: true,
		}, nil
	}

	content := string(data)

	count := strings.Count(content, oldString)

	if count == 0 {
		return &ToolResult{
			Content: fmt.Sprintf("Error: old_string not found in %s. Make sure the text matches exactly including whitespace and indentation.", filePath),
			IsError: true,
		}, nil
	}

	if count > 1 {
		return &ToolResult{
			Content: fmt.Sprintf("Error: old_string found %d times in %s. Please provide a more specific match that occurs exactly once.", count, filePath),
			IsError: true,
		}, nil
	}

	newContent := strings.Replace(content, oldString, newString, 1)

	if err := os.WriteFile(absPath, []byte(newContent), 0644); err != nil {
		return &ToolResult{
			Content: fmt.Sprintf("Error writing file: %v", err),
			IsError: true,
		}, nil
	}

	oldLines := len(strings.Split(oldString, "\n"))
	newLines := len(strings.Split(newString, "\n"))

	return &ToolResult{
		Content:       fmt.Sprintf("Successfully edited %s. Replaced %d line(s) with %d line(s).", filePath, oldLines, newLines),
		Display:       fmt.Sprintf("[E] Edited %s (-%d/+%d lines)", filepath.Base(filePath), oldLines, newLines),
		FilesModified: []string{absPath},
	}, nil
}

func (t *EditFileTool) resolvePath(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(t.workDir, p)
}
