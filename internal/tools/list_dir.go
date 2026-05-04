package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type ListDirTool struct {
	workDir string
}

func NewListDirTool(workDir string) *ListDirTool {
	return &ListDirTool{workDir: workDir}
}

func (t *ListDirTool) Name() string { return "list_directory" }

func (t *ListDirTool) Description() string {
	return "List files and directories in a given path. Returns file names, sizes, and types. Use this to explore project structure."
}

func (t *ListDirTool) ParameterSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "The directory path to list (relative to working directory or absolute). Defaults to '.' if omitted.",
			},
			"recursive": map[string]interface{}{
				"type":        "boolean",
				"description": "If true, list contents recursively. Defaults to false.",
			},
		},
		"required": []string{},
	}
}

func (t *ListDirTool) NeedsConfirmation(params map[string]interface{}) bool {
	return false
}

func (t *ListDirTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	dirPath := "."
	if p, ok := getStringParam(params, "path"); ok && p != "" {
		dirPath = p
	}

	recursive, _ := getBoolParam(params, "recursive")

	absPath := t.resolvePath(dirPath)

	info, err := os.Stat(absPath)
	if err != nil {
		return &ToolResult{
			Content: fmt.Sprintf("Error: %v", err),
			IsError: true,
		}, nil
	}
	if !info.IsDir() {
		return &ToolResult{
			Content: fmt.Sprintf("Error: '%s' is not a directory", dirPath),
			IsError: true,
		}, nil
	}

	var entries []string
	maxEntries := 500

	if recursive {
		count := 0
		filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if count >= maxEntries {
				return filepath.SkipAll
			}

			name := info.Name()
			if strings.HasPrefix(name, ".") && name != "." {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if name == "node_modules" || name == "vendor" || name == "__pycache__" {
				return filepath.SkipDir
			}

			relPath, _ := filepath.Rel(absPath, path)
			if relPath == "." {
				return nil
			}

			prefix := "[F]"
			suffix := fmt.Sprintf(" (%s)", formatSize(info.Size()))
			if info.IsDir() {
				prefix = "[D]"
				suffix = "/"
			}

			entries = append(entries, fmt.Sprintf("%s %s%s", prefix, relPath, suffix))
			count++
			return nil
		})
	} else {
		dirEntries, err := os.ReadDir(absPath)
		if err != nil {
			return &ToolResult{
				Content: fmt.Sprintf("Error reading directory: %v", err),
				IsError: true,
			}, nil
		}

		sort.Slice(dirEntries, func(i, j int) bool {
			di, dj := dirEntries[i].IsDir(), dirEntries[j].IsDir()
			if di != dj {
				return di
			}
			return dirEntries[i].Name() < dirEntries[j].Name()
		})

		for _, entry := range dirEntries {
			if len(entries) >= maxEntries {
				break
			}
			name := entry.Name()
			prefix := "[F]"
			suffix := ""

			if entry.IsDir() {
				prefix = "[D]"
				suffix = "/"
			} else {
				info, err := entry.Info()
				if err == nil {
					suffix = fmt.Sprintf(" (%s)", formatSize(info.Size()))
				}
			}
			entries = append(entries, fmt.Sprintf("%s %s%s", prefix, name, suffix))
		}
	}

	if len(entries) == 0 {
		return &ToolResult{
			Content: fmt.Sprintf("Directory '%s' is empty.", dirPath),
			Display: fmt.Sprintf("[L] %s (empty)", dirPath),
		}, nil
	}

	result := fmt.Sprintf("Contents of %s:\n\n%s", dirPath, strings.Join(entries, "\n"))
	if len(entries) >= maxEntries {
		result += fmt.Sprintf("\n\n... (truncated at %d entries)", maxEntries)
	}

	return &ToolResult{
		Content: result,
		Display: fmt.Sprintf("[L] Listed %s (%d entries)", dirPath, len(entries)),
	}, nil
}

func (t *ListDirTool) resolvePath(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(t.workDir, p)
}

func formatSize(size int64) string {
	switch {
	case size < 1024:
		return fmt.Sprintf("%dB", size)
	case size < 1024*1024:
		return fmt.Sprintf("%.1fKB", float64(size)/1024)
	case size < 1024*1024*1024:
		return fmt.Sprintf("%.1fMB", float64(size)/(1024*1024))
	default:
		return fmt.Sprintf("%.1fGB", float64(size)/(1024*1024*1024))
	}
}
