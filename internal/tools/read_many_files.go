package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ReadManyFilesTool struct {
	workDir string
}

func NewReadManyFilesTool(workDir string) *ReadManyFilesTool {
	return &ReadManyFilesTool{workDir: workDir}
}

func (t *ReadManyFilesTool) Name() string { return "read_many_files" }

func (t *ReadManyFilesTool) Description() string {
	return "Read the contents of multiple files at once. Useful for understanding relationships between files."
}

func (t *ReadManyFilesTool) ParameterSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_paths": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "List of file paths to read.",
			},
		},
		"required": []string{"file_paths"},
	}
}

func (t *ReadManyFilesTool) NeedsConfirmation(params map[string]interface{}) bool {
	return false
}

func (t *ReadManyFilesTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	paths, ok := params["file_paths"].([]interface{})
	if !ok {
		return &ToolResult{Content: "Error: file_paths must be an array", IsError: true}, nil
	}

	var sb strings.Builder
	var displays []string

	for _, p := range paths {
		filePath, ok := p.(string)
		if !ok || filePath == "" {
			continue
		}

		absPath := t.resolvePath(filePath)
		data, err := os.ReadFile(absPath)
		if err != nil {
			sb.WriteString(fmt.Sprintf("--- Error reading %s: %v ---\n\n", filePath, err))
			continue
		}

		lines := strings.Split(string(data), "\n")
		sb.WriteString(fmt.Sprintf("--- File: %s (%d lines) ---\n", filePath, len(lines)))
		for i, line := range lines {
			sb.WriteString(fmt.Sprintf("%d: %s\n", i+1, line))
		}
		sb.WriteString("\n")
		displays = append(displays, filepath.Base(filePath))
	}

	return &ToolResult{
		Content: sb.String(),
		Display: fmt.Sprintf("[R] Read %d files: %s", len(displays), strings.Join(displays, ", ")),
	}, nil
}

func (t *ReadManyFilesTool) resolvePath(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(t.workDir, p)
}
