package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ReadFileTool struct {
	workDir string
}

func NewReadFileTool(workDir string) *ReadFileTool {
	return &ReadFileTool{workDir: workDir}
}

func (t *ReadFileTool) Name() string { return "read_file" }

func (t *ReadFileTool) Description() string {
	return "Read the contents of a file. Supports optional start_line and end_line parameters to read specific line ranges. Use this to understand code before making edits."
}

func (t *ReadFileTool) ParameterSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "The path to the file to read (relative to working directory or absolute).",
			},
			"start_line": map[string]interface{}{
				"type":        "integer",
				"description": "Optional. Start line number (1-indexed). If omitted, reads from the beginning.",
			},
			"end_line": map[string]interface{}{
				"type":        "integer",
				"description": "Optional. End line number (1-indexed, inclusive). If omitted, reads to the end.",
			},
		},
		"required": []string{"file_path"},
	}
}

func (t *ReadFileTool) NeedsConfirmation(params map[string]interface{}) bool {
	return false
}

func (t *ReadFileTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	filePath, ok := params["file_path"].(string)
	if !ok || filePath == "" {
		return &ToolResult{Content: "Error: file_path is required", IsError: true}, nil
	}

	absPath := t.resolvePath(filePath)

	data, err := os.ReadFile(absPath)
	if err != nil {
		return &ToolResult{
			Content: fmt.Sprintf("Error reading file: %v", err),
			IsError: true,
		}, nil
	}

	lines := strings.Split(string(data), "\n")
	totalLines := len(lines)

	startLine := 1
	endLine := totalLines

	if v, ok := getIntParam(params, "start_line"); ok && v > 0 {
		startLine = v
	}
	if v, ok := getIntParam(params, "end_line"); ok && v > 0 {
		endLine = v
	}

	if startLine > totalLines {
		startLine = totalLines
	}
	if endLine > totalLines {
		endLine = totalLines
	}
	if startLine > endLine {
		startLine = endLine
	}

	selectedLines := lines[startLine-1 : endLine]

	var sb strings.Builder
	for i, line := range selectedLines {
		sb.WriteString(fmt.Sprintf("%d: %s\n", startLine+i, line))
	}

	header := fmt.Sprintf("File: %s (lines %d-%d of %d)\n\n", filePath, startLine, endLine, totalLines)

	return &ToolResult{
		Content: header + sb.String(),
		Display: fmt.Sprintf("[R] Read %s (lines %d-%d of %d)", filepath.Base(filePath), startLine, endLine, totalLines),
	}, nil
}

func (t *ReadFileTool) resolvePath(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(t.workDir, p)
}

func getIntParam(params map[string]interface{}, key string) (int, bool) {
	v, ok := params[key]
	if !ok {
		return 0, false
	}
	switch val := v.(type) {
	case int:
		return val, true
	case float64:
		return int(val), true
	case int64:
		return int(val), true
	}
	return 0, false
}

func getStringParam(params map[string]interface{}, key string) (string, bool) {
	v, ok := params[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func getBoolParam(params map[string]interface{}, key string) (bool, bool) {
	v, ok := params[key]
	if !ok {
		return false, false
	}
	b, ok := v.(bool)
	return b, ok
}
