package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type SearchSymbolTool struct {
	WorkDir string
}

func NewSearchSymbolTool(workDir string) *SearchSymbolTool {
	return &SearchSymbolTool{WorkDir: workDir}
}

func (t *SearchSymbolTool) Name() string {
	return "search_symbol"
}

func (t *SearchSymbolTool) NeedsConfirmation(args map[string]interface{}) bool {
	return false
}

func (t *SearchSymbolTool) Description() string {
	return "Find code symbols (functions, classes, interfaces, structs) across the project using regex patterns."
}

func (t *SearchSymbolTool) ParameterSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "The name or partial name of the symbol to find.",
			},
			"include_patterns": map[string]interface{}{
				"type":        "string",
				"description": "Optional glob patterns for file inclusion (e.g. *.go, *.ts).",
			},
		},
		"required": []string{"query"},
	}
}

func (t *SearchSymbolTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	query, ok := params["query"].(string)
	if !ok {
		return nil, fmt.Errorf("query is required")
	}

	include, _ := params["include_patterns"].(string)

	patterns := []string{
		fmt.Sprintf(`func\s+%s`, query),
		fmt.Sprintf(`type\s+%s\s+(struct|interface)`, query),
		fmt.Sprintf(`class\s+%s`, query),
		fmt.Sprintf(`interface\s+%s`, query),
		fmt.Sprintf(`def\s+%s`, query),
		fmt.Sprintf(`export\s+(const|let|var|function|class|interface)\s+%s`, query),
		fmt.Sprintf(`fn\s+%s`, query),
	}

	fullRegex := "(" + strings.Join(patterns, "|") + ")"

	args := []string{"-rnE", fullRegex, "."}
	if include != "" {
		args = append([]string{"--include", include}, args...)
	}

	cmd := exec.Command("grep", args...)
	cmd.Dir = t.WorkDir
	out, err := cmd.Output()
	if err != nil && len(out) == 0 {
		return &ToolResult{
			Content: fmt.Sprintf("No symbols matching '%s' found.", query),
		}, nil
	}

	return &ToolResult{
		Content: string(out),
	}, nil
}
