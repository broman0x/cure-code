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
	return "Get a high-level overview of the project structure, including key files and a directory tree. Useful at the start of a session."
}

func (t *ProjectSummaryTool) ParameterSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

func (t *ProjectSummaryTool) NeedsConfirmation(params map[string]interface{}) bool {
	return false
}

func (t *ProjectSummaryTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	var sb strings.Builder

	sb.WriteString("## PROJECT STRUCTURE\n\n")

	tree := t.generateTree(t.workDir, "", 0, 2)
	sb.WriteString(tree)

	sb.WriteString("\n## KEY FILES\n")
	keyFiles := []string{"README.md", "package.json", "go.mod", "requirements.txt", "CURECODE.md", "CODEBASE.md"}
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
		Display: "[I] Generated project summary",
	}, nil
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
	for i, entry := range entries {
		name := entry.Name()

		if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" || name == "__pycache__" {
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
	}
	return sb.String()
}
