package tools

import (
	"context"
	"fmt"
)

type ExecuteExtraTool struct {
	registry *ToolRegistry
}

func NewExecuteExtraTool(registry *ToolRegistry) *ExecuteExtraTool {
	return &ExecuteExtraTool{registry: registry}
}

func (t *ExecuteExtraTool) Name() string { return "execute_extra_tool" }

func (t *ExecuteExtraTool) Description() string {
	return "Execute a deferred tool by name. Use this after discovering candidates with search_extra_tools."
}

func (t *ExecuteExtraTool) ParameterSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"tool_name": map[string]interface{}{
				"type":        "string",
				"description": "Deferred tool name to execute.",
			},
			"arguments": map[string]interface{}{
				"type":        "object",
				"description": "Arguments object for the deferred tool call.",
			},
		},
		"required": []string{"tool_name"},
	}
}

func (t *ExecuteExtraTool) NeedsConfirmation(params map[string]interface{}) bool {
	return false
}

func (t *ExecuteExtraTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	toolName, ok := getStringParam(params, "tool_name")
	if !ok || toolName == "" {
		return &ToolResult{Content: "Error: tool_name is required", IsError: true}, nil
	}

	if toolName == t.Name() || toolName == "search_extra_tools" {
		return &ToolResult{Content: "Error: recursive execution is not allowed", IsError: true}, nil
	}

	if !t.registry.IsDeferred(toolName) {
		return &ToolResult{
			Content: fmt.Sprintf("Error: '%s' is not a deferred tool. Call it directly instead.", toolName),
			IsError: true,
		}, nil
	}

	args := map[string]interface{}{}
	if raw, ok := params["arguments"]; ok && raw != nil {
		if cast, ok := raw.(map[string]interface{}); ok {
			args = cast
		} else {
			return &ToolResult{Content: "Error: arguments must be an object", IsError: true}, nil
		}
	}

	target, ok := t.registry.Get(toolName)
	if !ok {
		return &ToolResult{Content: fmt.Sprintf("Error: unknown tool '%s'", toolName), IsError: true}, nil
	}

	if target.NeedsConfirmation(args) {
		return &ToolResult{
			Content: fmt.Sprintf("Error: deferred tool '%s' requires confirmation and cannot be called through execute_extra_tool.", toolName),
			IsError: true,
		}, nil
	}

	result, err := target.Execute(ctx, args)
	if err != nil {
		return nil, err
	}

	if result.Display == "" {
		result.Display = fmt.Sprintf("[I] Deferred tool executed: %s", toolName)
	}
	return result, nil
}
