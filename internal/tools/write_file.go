package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type WriteFileTool struct {
	workDir string
}

func NewWriteFileTool(workDir string) *WriteFileTool {
	return &WriteFileTool{workDir: workDir}
}

func (t *WriteFileTool) Name() string { return "write_file" }

func (t *WriteFileTool) Description() string {
	return "Create a new file or overwrite an existing file with the provided content. Use this for creating new files. For modifying existing files, prefer the edit_file tool."
}

func (t *WriteFileTool) ParameterSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "The path of the file to write (relative to working directory or absolute).",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "The complete content to write to the file.",
			},
		},
		"required": []string{"file_path", "content"},
	}
}

func (t *WriteFileTool) NeedsConfirmation(params map[string]interface{}) bool {
	return true
}

func (t *WriteFileTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	filePath, ok := getStringParam(params, "file_path")
	if !ok || filePath == "" {
		return &ToolResult{Content: "Error: file_path is required", IsError: true}, nil
	}

	content, ok := getStringParam(params, "content")
	if !ok {
		return &ToolResult{Content: "Error: content is required", IsError: true}, nil
	}

	absPath := t.resolvePath(filePath)

	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &ToolResult{
			Content: fmt.Sprintf("Error creating directory: %v", err),
			IsError: true,
		}, nil
	}

	_, existErr := os.Stat(absPath)
	isNew := os.IsNotExist(existErr)

	if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
		return &ToolResult{
			Content: fmt.Sprintf("Error writing file: %v", err),
			IsError: true,
		}, nil
	}

	action := "Updated"
	if isNew {
		action = "Created"
	}

	lineCount := len(strings.Split(content, "\n"))
	return &ToolResult{
		Content:       fmt.Sprintf("Successfully %s file: %s (%d lines)", strings.ToLower(action), filePath, lineCount),
		Display:       fmt.Sprintf("[W] %s %s (%d lines)", action, filepath.Base(filePath), lineCount),
		FilesModified: []string{absPath},
	}, nil
}

func (t *WriteFileTool) resolvePath(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(t.workDir, p)
}
