package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

type TodoTool struct{}

func (t *TodoTool) Name() string {
	return "write_todos"
}

func (t *TodoTool) Description() string {
	return "Maintain a list of subtasks for multi-step requests. This helps you plan your work and keeps the user informed of your progress."
}

func (t *TodoTool) ParameterSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"todos": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"description": map[string]interface{}{
							"type":        "string",
							"description": "Technical description of the subtask.",
						},
						"status": map[string]interface{}{
							"type": "string",
							"enum": []string{"pending", "in_progress", "completed", "cancelled", "blocked"},
						},
					},
					"required": []string{"description", "status"},
				},
			},
		},
		"required": []string{"todos"},
	}
}

func (t *TodoTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	todosRaw, ok := args["todos"]
	if !ok {
		return nil, fmt.Errorf("missing todos parameter")
	}

	todosJSON, err := json.Marshal(todosRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal todos: %v", err)
	}

	return &ToolResult{
		Content: string(todosJSON),
		Display: "[T] Updated task list",
		Metadata: map[string]interface{}{
			"todos": todosRaw,
		},
	}, nil
}

func (t *TodoTool) NeedsConfirmation(params map[string]interface{}) bool {
	return false
}
