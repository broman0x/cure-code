package tools

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

type SearchExtraToolsTool struct {
	registry *ToolRegistry
}

func NewSearchExtraToolsTool(registry *ToolRegistry) *SearchExtraToolsTool {
	return &SearchExtraToolsTool{registry: registry}
}

func (t *SearchExtraToolsTool) Name() string { return "search_extra_tools" }

func (t *SearchExtraToolsTool) Description() string {
	return "Search deferred tools by intent. Use this when core tools are not enough and you need additional capabilities."
}

func (t *SearchExtraToolsTool) ParameterSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "What capability you are looking for.",
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of tools to return (default 5).",
			},
		},
		"required": []string{"query"},
	}
}

func (t *SearchExtraToolsTool) NeedsConfirmation(params map[string]interface{}) bool {
	return false
}

func (t *SearchExtraToolsTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	query, ok := getStringParam(params, "query")
	if !ok || strings.TrimSpace(query) == "" {
		return &ToolResult{Content: "Error: query is required", IsError: true}, nil
	}

	limit := 5
	if v, ok := getIntParam(params, "limit"); ok && v > 0 && v <= 20 {
		limit = v
	}

	query = strings.ToLower(strings.TrimSpace(query))
	type scored struct {
		def   ToolDefinition
		score int
	}
	var candidates []scored

	for _, def := range t.registry.DeferredDefinitions() {
		score := toolRelevanceScore(def, query)
		if score == 0 {
			continue
		}
		candidates = append(candidates, scored{def: def, score: score})
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].score == candidates[j].score {
			return candidates[i].def.Name < candidates[j].def.Name
		}
		return candidates[i].score > candidates[j].score
	})

	if len(candidates) == 0 {
		return &ToolResult{
			Content: "No deferred tools matched your query.",
			Display: "[I] No extra tools matched",
		}, nil
	}

	if len(candidates) > limit {
		candidates = candidates[:limit]
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Deferred tool matches for: %s\n\n", query))
	for i, c := range candidates {
		b.WriteString(fmt.Sprintf("%d. %s\n   %s\n\n", i+1, c.def.Name, c.def.Description))
	}
	b.WriteString("Use execute_extra_tool with 'tool_name' and 'arguments' to run one of these tools.")

	return &ToolResult{
		Content: b.String(),
		Display: fmt.Sprintf("[I] Found %d deferred tool matches", len(candidates)),
		Metadata: map[string]interface{}{
			"matches": len(candidates),
		},
	}, nil
}

func toolRelevanceScore(def ToolDefinition, query string) int {
	name := strings.ToLower(def.Name)
	desc := strings.ToLower(def.Description)

	score := 0
	if strings.Contains(name, query) {
		score += 8
	}
	if strings.Contains(desc, query) {
		score += 4
	}

	for _, tok := range strings.Fields(query) {
		if len(tok) < 2 {
			continue
		}
		if strings.Contains(name, tok) {
			score += 3
		}
		if strings.Contains(desc, tok) {
			score++
		}
	}
	return score
}
