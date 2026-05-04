package tools

import (
	"context"
	"fmt"
)

// [EN] DelegateTool allows the agent to spawn a sub-agent for a specific task.
// [ID] DelegateTool memungkinkan agen untuk membuat sub-agen untuk tugas tertentu.
type DelegateTool struct {
	// [EN] We need a way to call the agent's ProcessPrompt.
	// [ID] Kita butuh cara untuk memanggil ProcessPrompt milik agen.
	ProcessPromptFunc func(ctx context.Context, prompt string) (string, error)
}

func (t *DelegateTool) Name() string {
	return "delegate_task"
}

func (t *DelegateTool) Description() string {
	return "Delegate a specific sub-task to a specialized sub-agent. Useful for research, auditing code, or implementing isolated features without bloating the main conversation context."
}

func (t *DelegateTool) ParameterSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"task": map[string]interface{}{
				"type":        "string",
				"description": "The specific task for the sub-agent (e.g. 'Research how the authentication middleware works').",
			},
			"context_files": map[string]interface{}{
				"type":        "array",
				"description": "Optional list of files to provide as initial context to the sub-agent.",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
		},
		"required": []string{"task"},
	}
}

func (t *DelegateTool) NeedsConfirmation(args map[string]interface{}) bool {
	return true // [EN] Always confirm delegation | [ID] Selalu konfirmasi delegasi
}

func (t *DelegateTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	task, _ := args["task"].(string)
	contextFiles, _ := args["context_files"].([]interface{})
	
	if t.ProcessPromptFunc == nil {
		return nil, fmt.Errorf("ProcessPromptFunc not initialized")
	}

	// [EN] We wrap the task to instruct the sub-agent to be concise and return a summary.
	// [ID] Kita bungkus tugasnya untuk menginstruksikan sub-agen agar ringkas dan mengembalikan ringkasan.
	subPrompt := fmt.Sprintf("SUB-AGENT TASK: %s\n\nPlease execute this task and provide a final summary of your findings or implementation. Be concise.", task)

	// [EN] Pass context files if any
	// [ID] Lewati file konteks jika ada
	if len(contextFiles) > 0 {
		subPrompt = fmt.Sprintf("CONTEXT FILES: %v\n\n%s", contextFiles, subPrompt)
	}

	result, err := t.ProcessPromptFunc(ctx, subPrompt)
	if err != nil {
		return &ToolResult{
			Content: fmt.Sprintf("Sub-agent failed: %v", err),
			IsError: true,
		}, nil
	}

	return &ToolResult{
		Content: fmt.Sprintf("Sub-agent completed task. Summary:\n%s", result),
		Display: "Sub-agent finished delegated task.",
		Metadata: map[string]interface{}{
			"sub_agent_task": task,
		},
	}, nil
}
