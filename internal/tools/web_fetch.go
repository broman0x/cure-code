package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type WebFetchTool struct{}

func NewWebFetchTool() *WebFetchTool {
	return &WebFetchTool{}
}

func (t *WebFetchTool) Name() string { return "web_fetch" }

func (t *WebFetchTool) Description() string {
	return "Fetch the content of a web page. Useful for reading documentation or external resources."
}

func (t *WebFetchTool) ParameterSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"url": map[string]interface{}{
				"type":        "string",
				"description": "The URL to fetch.",
			},
		},
		"required": []string{"url"},
	}
}

func (t *WebFetchTool) NeedsConfirmation(params map[string]interface{}) bool {
	return false
}

func (t *WebFetchTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	url, ok := params["url"].(string)
	if !ok || url == "" {
		return &ToolResult{Content: "Error: url is required", IsError: true}, nil
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return &ToolResult{Content: fmt.Sprintf("Error creating request: %v", err), IsError: true}, nil
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return &ToolResult{Content: fmt.Sprintf("Error fetching URL: %v", err), IsError: true}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &ToolResult{Content: fmt.Sprintf("Error: received status code %d", resp.StatusCode), IsError: true}, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ToolResult{Content: fmt.Sprintf("Error reading response: %v", err), IsError: true}, nil
	}

	return &ToolResult{
		Content: string(body),
		Display: fmt.Sprintf("[W] Fetched %s (%d bytes)", url, len(body)),
	}, nil
}
