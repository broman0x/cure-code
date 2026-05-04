package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type WebSearchTool struct{}

func NewWebSearchTool() *WebSearchTool {
	return &WebSearchTool{}
}

func (t *WebSearchTool) Name() string { return "web_search" }

func (t *WebSearchTool) Description() string {
	return "Search the web for information, documentation, or code examples. Returns a list of relevant snippets and URLs."
}

func (t *WebSearchTool) ParameterSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "The search query.",
			},
		},
		"required": []string{"query"},
	}
}

func (t *WebSearchTool) NeedsConfirmation(params map[string]interface{}) bool {
	return false
}

func (t *WebSearchTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	query, ok := params["query"].(string)
	if !ok || query == "" {
		return &ToolResult{Content: "Error: query is required", IsError: true}, nil
	}

	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return &ToolResult{Content: fmt.Sprintf("Error creating request: %v", err), IsError: true}, nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return &ToolResult{Content: fmt.Sprintf("Error performing search: %v", err), IsError: true}, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ToolResult{Content: fmt.Sprintf("Error reading search results: %v", err), IsError: true}, nil
	}

	html := string(body)

	re := regexp.MustCompile(`<a class="result__a" href="([^"]+)">([^<]+)</a>`)
	matches := re.FindAllStringSubmatch(html, 10)

	if len(matches) == 0 {
		return &ToolResult{Content: "No results found for your query.", IsError: false}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Search results for: %s\n\n", query))
	for i, m := range matches {
		link := m[1]
		title := m[2]

		if strings.Contains(link, "uddg=") {
			u, err := url.Parse(link)
			if err == nil {
				link = u.Query().Get("uddg")
			}
		}

		sb.WriteString(fmt.Sprintf("%d. %s\n   URL: %s\n\n", i+1, title, link))
	}

	return &ToolResult{
		Content: sb.String(),
		Display: fmt.Sprintf("[W] Searched '%s' (%d results)", query, len(matches)),
	}, nil
}
