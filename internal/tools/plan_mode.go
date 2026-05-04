package tools

import (
	"context"
)

// [EN] EnterPlanModeTool switches the agent to planning state.
// [ID] EnterPlanModeTool mengalihkan agen ke status planning.
type EnterPlanModeTool struct{}

func (t *EnterPlanModeTool) Name() string { return "enter_plan_mode" }
func (t *EnterPlanModeTool) Description() string {
	return "Enter Plan Mode to explore the codebase and design an approach before making changes. File modifications are restricted in this mode."
}
func (t *EnterPlanModeTool) ParameterSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}
func (t *EnterPlanModeTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	return &ToolResult{
		Content: "Entered Plan Mode. You are now in a read-only exploration phase. Focus on understanding the codebase and designing a plan. DO NOT write or edit any files yet.",
		Display: "Entered Plan Mode (Read-Only)",
	}, nil
}
func (t *EnterPlanModeTool) NeedsConfirmation(params map[string]interface{}) bool { return false }

// [EN] ExitPlanModeTool switches the agent back to execution state.
// [ID] ExitPlanModeTool mengalihkan agen kembali ke status eksekusi.
type ExitPlanModeTool struct{}

func (t *ExitPlanModeTool) Name() string { return "exit_plan_mode" }
func (t *ExitPlanModeTool) Description() string {
	return "Exit Plan Mode and return to execution mode to apply the designed changes."
}
func (t *ExitPlanModeTool) ParameterSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}
func (t *ExitPlanModeTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	return &ToolResult{
		Content: "Exited Plan Mode. You are now in Execution Mode. You can proceed with making changes as per your plan.",
		Display: "Exited Plan Mode (Execution Enabled)",
	}, nil
}
func (t *ExitPlanModeTool) NeedsConfirmation(params map[string]interface{}) bool { return false }
