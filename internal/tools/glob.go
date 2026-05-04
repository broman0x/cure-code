package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type GlobTool struct {
	workDir string
}

func NewGlobTool(workDir string) *GlobTool {
	return &GlobTool{workDir: workDir}
}

func (t *GlobTool) Name() string { return "glob" }

func (t *GlobTool) Description() string {
	return "Find files matching a glob pattern. Returns a list of matching file paths."
}

func (t *GlobTool) ParameterSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "Glob pattern (e.g., '*.go', '*.ts').",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Directory to search in. Defaults to '.'.",
			},
		},
		"required": []string{"pattern"},
	}
}

func (t *GlobTool) NeedsConfirmation(params map[string]interface{}) bool { return false }

func (t *GlobTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	pattern, ok := getStringParam(params, "pattern")
	if !ok || pattern == "" {
		return &ToolResult{Content: "Error: pattern is required", IsError: true}, nil
	}
	searchPath := "."
	if p, ok := getStringParam(params, "path"); ok && p != "" {
		searchPath = p
	}
	absPath := t.resolvePath(searchPath)
	skipDirs := map[string]bool{
		".git": true, "node_modules": true, "vendor": true, "__pycache__": true,
	}
	var matches []string
	filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || len(matches) >= 500 {
			return nil
		}
		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		name := info.Name()
		matched, _ := filepath.Match(pattern, name)
		if !matched && strings.HasPrefix(pattern, "**") {
			suffix := strings.TrimPrefix(pattern, "**/")
			matched, _ = filepath.Match(suffix, name)
		}
		if matched {
			relPath, _ := filepath.Rel(absPath, path)
			matches = append(matches, relPath)
		}
		return nil
	})
	if len(matches) == 0 {
		return &ToolResult{Content: fmt.Sprintf("No files matching '%s'", pattern)}, nil
	}
	return &ToolResult{
		Content: fmt.Sprintf("Found %d file(s):\n\n%s", len(matches), strings.Join(matches, "\n")),
		Display: fmt.Sprintf("[S] Found %d files matching '%s'", len(matches), pattern),
	}, nil
}

func (t *GlobTool) resolvePath(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(t.workDir, p)
}
