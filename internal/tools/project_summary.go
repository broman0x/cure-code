package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ProjectSummaryTool struct {
	workDir string
}

func NewProjectSummaryTool(workDir string) *ProjectSummaryTool {
	return &ProjectSummaryTool{workDir: workDir}
}

func (t *ProjectSummaryTool) Name() string { return "get_project_summary" }

func (t *ProjectSummaryTool) Description() string {
	return "Get a high-level overview of the project structure, including entry points, key files, and directory tree. Useful for orienting yourself in a new codebase."
}

func (t *ProjectSummaryTool) ParameterSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"depth": map[string]interface{}{
				"type":        "integer",
				"description": "Depth of the directory tree (default 2)",
			},
		},
	}
}

func (t *ProjectSummaryTool) NeedsConfirmation(params map[string]interface{}) bool {
	return false
}

func (t *ProjectSummaryTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	depth := 2
	if d, ok := params["depth"].(float64); ok {
		depth = int(d)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Project Summary: %s\n\n", filepath.Base(t.workDir)))

	// [EN] Find entry points
	// [ID] Cari titik masuk
	entryPoints := t.findEntryPoints()
	if len(entryPoints) > 0 {
		sb.WriteString("## Entry Points\n")
		for _, ep := range entryPoints {
			sb.WriteString(fmt.Sprintf("- %s\n", ep))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Directory Structure\n")
	tree := t.generateTree(t.workDir, "", 0, depth)
	sb.WriteString(tree)

	sb.WriteString("\n## Key Config Files\n")
	keyFiles := []string{"README.md", "package.json", "go.mod", "requirements.txt", "CURECODE.md", "CODEBASE.md", ".env.example", "Makefile", "Dockerfile"}
	foundKey := false
	for _, f := range keyFiles {
		if _, err := os.Stat(filepath.Join(t.workDir, f)); err == nil {
			sb.WriteString(fmt.Sprintf("- %s\n", f))
			foundKey = true
		}
	}
	if !foundKey {
		sb.WriteString("None found at root.\n")
	}

	return &ToolResult{
		Content: sb.String(),
		Display: "Project summary generated.",
	}, nil
}

func (t *ProjectSummaryTool) findEntryPoints() []string {
	var entries []string
	filepath.Walk(t.workDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			if info != nil && info.IsDir() && (info.Name() == "vendor" || info.Name() == "node_modules" || strings.HasPrefix(info.Name(), ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		name := info.Name()
		if name == "main.go" || name == "index.js" || name == "app.py" || name == "manage.py" || name == "server.js" {
			rel, _ := filepath.Rel(t.workDir, path)
			entries = append(entries, rel)
		}
		return nil
	})
	return entries
}

func (t *ProjectSummaryTool) generateTree(root, indent string, depth, maxDepth int) string {
	if depth > maxDepth {
		return ""
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return ""
	}

	var sb strings.Builder
	count := 0
	for i, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" || name == "__pycache__" || name == "bin" {
			continue
		}

		isLast := i == len(entries)-1
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		sb.WriteString(indent + connector + name)
		if entry.IsDir() {
			sb.WriteString("/")
		}
		sb.WriteString("\n")

		if entry.IsDir() && depth < maxDepth {
			newIndent := indent + "│   "
			if isLast {
				newIndent = indent + "    "
			}
			sb.WriteString(t.generateTree(filepath.Join(root, name), newIndent, depth+1, maxDepth))
		}
		count++
		if count > 30 && depth == 0 {
			sb.WriteString(indent + "└── ... (more files truncated)\n")
			break
		}
	}
	return sb.String()
}
