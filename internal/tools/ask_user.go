package tools

import (
	"context"
	"fmt"
)

type AskUserTool struct{}

func NewAskUserTool() *AskUserTool { return &AskUserTool{} }

func (t *AskUserTool) Name() string { return "ask_user" }

func (t *AskUserTool) Description() string {
	return "Ask the user a question when you need clarification or more information to proceed. Use this when instructions are ambiguous."
}

func (t *AskUserTool) ParameterSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"question": map[string]interface{}{
				"type":        "string",
				"description": "The question to ask the user.",
			},
		},
		"required": []string{"question"},
	}
}

func (t *AskUserTool) NeedsConfirmation(params map[string]interface{}) bool { return false }

func (t *AskUserTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	question, ok := getStringParam(params, "question")
	if !ok || question == "" {
		return &ToolResult{Content: "Error: question is required", IsError: true}, nil
	}

	return &ToolResult{
		Content: fmt.Sprintf("QUESTION_FOR_USER: %s", question),
		Display: fmt.Sprintf("[?] %s", question),
	}, nil
}
