package tools

import (
	"context"
	"fmt"
	"github.com/broman0x/cure-code/internal/mcp"
)

type MCPTool struct {
	client     *mcp.Client
	serverName string
	tool       mcp.Tool
}

func NewMCPTool(client *mcp.Client, serverName string, tool mcp.Tool) *MCPTool {
	return &MCPTool{
		client:     client,
		serverName: serverName,
		tool:       tool,
	}
}

func (t *MCPTool) Name() string {
	// Prefix with server name to avoid conflicts if multiple servers have the same tool name
	return fmt.Sprintf("%s_%s", t.serverName, t.tool.Name)
}

func (t *MCPTool) Description() string {
	if t.tool.Description != "" {
		return fmt.Sprintf("[MCP %s] %s", t.serverName, t.tool.Description)
	}
	return fmt.Sprintf("Execute tool '%s' on MCP server '%s'", t.tool.Name, t.serverName)
}

func (t *MCPTool) ParameterSchema() map[string]interface{} {
	return t.tool.InputSchema
}

func (t *MCPTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	output, isError, err := t.client.CallTool(t.tool.Name, params)
	
	if err != nil {
		return &ToolResult{
			Content: fmt.Sprintf("Error calling MCP tool %s: %v", t.tool.Name, err),
			IsError: true,
			Display: fmt.Sprintf("[!] MCP %s error: %v", t.tool.Name, err),
		}, nil
	}
	
	display := fmt.Sprintf("[MCP] Executed %s_%s", t.serverName, t.tool.Name)
	return &ToolResult{
		Content: output,
		IsError: isError,
		Display: display,
	}, nil
}

func (t *MCPTool) NeedsConfirmation(params map[string]interface{}) bool {
	// We default to false for MCP tools for now, unless we can infer it's dangerous
	return false
}
