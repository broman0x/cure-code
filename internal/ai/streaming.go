package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/broman0x/cure-code/internal/agent"
	"github.com/broman0x/cure-code/internal/tools"
)

func (g *GeminiFCProvider) SupportsStreaming() bool { return true }

func (g *GeminiFCProvider) SendWithToolsStream(ctx context.Context, systemPrompt string, messages []agent.Message, toolDefs []tools.ToolDefinition) (<-chan agent.StreamEvent, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s", g.Model, g.ApiKey)

	reqBody := g.buildRequest(systemPrompt, messages, toolDefs)
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal error: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %v", err)
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, truncateBody(body))
	}

	ch := make(chan agent.StreamEvent, 32)
	go g.readGeminiSSE(resp, ch)
	return ch, nil
}

func (g *GeminiFCProvider) readGeminiSSE(resp *http.Response, ch chan<- agent.StreamEvent) {
	defer close(ch)
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 256*1024), 1024*1024)

	var usage *agent.UsageStats

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var raw map[string]interface{}
		if err := json.Unmarshal([]byte(data), &raw); err != nil {
			continue
		}

		if um, ok := raw["usageMetadata"].(map[string]interface{}); ok {
			usage = &agent.UsageStats{}
			if v, ok := um["promptTokenCount"].(float64); ok {
				usage.InputTokens = int(v)
			}
			if v, ok := um["candidatesTokenCount"].(float64); ok {
				usage.OutputTokens = int(v)
			}
			if v, ok := um["totalTokenCount"].(float64); ok {
				usage.TotalTokens = int(v)
			}
		}

		candidates, ok := raw["candidates"].([]interface{})
		if !ok || len(candidates) == 0 {
			continue
		}

		candidate, ok := candidates[0].(map[string]interface{})
		if !ok {
			continue
		}

		content, ok := candidate["content"].(map[string]interface{})
		if !ok {
			continue
		}

		parts, ok := content["parts"].([]interface{})
		if !ok {
			continue
		}

		for _, part := range parts {
			partMap, ok := part.(map[string]interface{})
			if !ok {
				continue
			}

			if text, ok := partMap["text"].(string); ok && text != "" {
				ch <- agent.StreamEvent{Type: agent.StreamText, Text: text}
			}

			if fc, ok := partMap["functionCall"].(map[string]interface{}); ok {
				name, _ := fc["name"].(string)
				args := make(map[string]interface{})
				if a, ok := fc["args"].(map[string]interface{}); ok {
					args = a
				}
				ch <- agent.StreamEvent{
					Type: agent.StreamToolCall,
					ToolCall: &agent.ToolCall{
						ID:   fmt.Sprintf("%s-%d", name, time.Now().UnixMilli()),
						Name: name,
						Args: args,
					},
				}
			}
		}

		if fr, ok := candidate["finishReason"].(string); ok && fr != "" {
			ch <- agent.StreamEvent{
				Type:         agent.StreamDone,
				FinishReason: fr,
				Usage:        usage,
			}
			return
		}
	}

	ch <- agent.StreamEvent{Type: agent.StreamDone, Usage: usage, FinishReason: "stop"}
}

func (o *OpenAIFCProvider) SupportsStreaming() bool { return true }

func (o *OpenAIFCProvider) SendWithToolsStream(ctx context.Context, systemPrompt string, messages []agent.Message, toolDefs []tools.ToolDefinition) (<-chan agent.StreamEvent, error) {
	url := "https://api.openai.com/v1/chat/completions"

	oaiMsgs := []map[string]interface{}{
		{"role": "system", "content": systemPrompt},
	}
	for _, msg := range messages {
		switch msg.Role {
		case "user":
			oaiMsgs = append(oaiMsgs, map[string]interface{}{"role": "user", "content": msg.Content})
		case "assistant":
			m := map[string]interface{}{"role": "assistant"}
			if msg.Content != "" {
				m["content"] = msg.Content
			}
			if len(msg.ToolCalls) > 0 {
				tcs := make([]map[string]interface{}, 0)
				for _, tc := range msg.ToolCalls {
					argsJSON, _ := json.Marshal(tc.Args)
					tcs = append(tcs, map[string]interface{}{
						"id": tc.ID, "type": "function",
						"function": map[string]interface{}{"name": tc.Name, "arguments": string(argsJSON)},
					})
				}
				m["tool_calls"] = tcs
			}
			oaiMsgs = append(oaiMsgs, m)
		case "tool":
			oaiMsgs = append(oaiMsgs, map[string]interface{}{
				"role": "tool", "tool_call_id": msg.ToolCallID, "content": msg.Content,
			})
		}
	}

	oaiTools := make([]map[string]interface{}, 0)
	for _, td := range toolDefs {
		oaiTools = append(oaiTools, map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name": td.Name, "description": td.Description, "parameters": td.Parameters,
			},
		})
	}

	reqBody := map[string]interface{}{
		"model": o.Model, "messages": oaiMsgs, "tools": oaiTools,
		"stream": true, "stream_options": map[string]interface{}{"include_usage": true},
	}

	payload, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.ApiKey)

	resp, err := o.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("OpenAI error (%d): %s", resp.StatusCode, truncateBody(body))
	}

	ch := make(chan agent.StreamEvent, 32)
	go o.readOpenAISSE(resp, ch)
	return ch, nil
}

func (o *OpenAIFCProvider) readOpenAISSE(resp *http.Response, ch chan<- agent.StreamEvent) {
	defer close(ch)
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 512*1024)

	tcAccum := make(map[int]*agent.ToolCall)
	var usage *agent.UsageStats

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk struct {
			Choices []struct {
				Delta struct {
					Content   *string `json:"content"`
					ToolCalls []struct {
						Index    int    `json:"index"`
						ID       string `json:"id"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls"`
				} `json:"delta"`
				FinishReason *string `json:"finish_reason"`
			} `json:"choices"`
			Usage *struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			} `json:"usage"`
		}

		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		if chunk.Usage != nil {
			usage = &agent.UsageStats{
				InputTokens:  chunk.Usage.PromptTokens,
				OutputTokens: chunk.Usage.CompletionTokens,
				TotalTokens:  chunk.Usage.TotalTokens,
			}
		}

		if len(chunk.Choices) == 0 {
			continue
		}

		delta := chunk.Choices[0].Delta

		if delta.Content != nil && *delta.Content != "" {
			ch <- agent.StreamEvent{Type: agent.StreamText, Text: *delta.Content}
		}

		for _, tc := range delta.ToolCalls {
			if _, ok := tcAccum[tc.Index]; !ok {
				tcAccum[tc.Index] = &agent.ToolCall{ID: tc.ID, Name: tc.Function.Name}
			}
			existing := tcAccum[tc.Index]
			if tc.ID != "" {
				existing.ID = tc.ID
			}
			if tc.Function.Name != "" {
				existing.Name = tc.Function.Name
			}

			if existing.Args == nil {
				existing.Args = map[string]interface{}{"_raw": tc.Function.Arguments}
			} else {
				existing.Args["_raw"] = existing.Args["_raw"].(string) + tc.Function.Arguments
			}
		}

		if chunk.Choices[0].FinishReason != nil {

			for _, tc := range tcAccum {
				if rawArgs, ok := tc.Args["_raw"].(string); ok {
					parsed := make(map[string]interface{})
					json.Unmarshal([]byte(rawArgs), &parsed)
					tc.Args = parsed
				}
				ch <- agent.StreamEvent{Type: agent.StreamToolCall, ToolCall: tc}
			}
			ch <- agent.StreamEvent{
				Type: agent.StreamDone, FinishReason: *chunk.Choices[0].FinishReason, Usage: usage,
			}
			return
		}
	}

	for _, tc := range tcAccum {
		if rawArgs, ok := tc.Args["_raw"].(string); ok {
			parsed := make(map[string]interface{})
			json.Unmarshal([]byte(rawArgs), &parsed)
			tc.Args = parsed
		}
		ch <- agent.StreamEvent{Type: agent.StreamToolCall, ToolCall: tc}
	}
	ch <- agent.StreamEvent{Type: agent.StreamDone, Usage: usage, FinishReason: "stop"}
}

func (o *GenericOpenAIFCProvider) SupportsStreaming() bool { return true }

func (o *GenericOpenAIFCProvider) SendWithToolsStream(ctx context.Context, systemPrompt string, messages []agent.Message, toolDefs []tools.ToolDefinition) (<-chan agent.StreamEvent, error) {
	url := strings.TrimSuffix(o.BaseURL, "/") + "/chat/completions"

	oaiMsgs := []map[string]interface{}{
		{"role": "system", "content": systemPrompt},
	}
	for _, msg := range messages {
		switch msg.Role {
		case "user":
			oaiMsgs = append(oaiMsgs, map[string]interface{}{"role": "user", "content": msg.Content})
		case "assistant":
			m := map[string]interface{}{"role": "assistant"}
			if msg.Content != "" {
				m["content"] = msg.Content
			}
			if len(msg.ToolCalls) > 0 {
				tcs := make([]map[string]interface{}, 0)
				for _, tc := range msg.ToolCalls {
					argsJSON, _ := json.Marshal(tc.Args)
					tcs = append(tcs, map[string]interface{}{
						"id": tc.ID, "type": "function",
						"function": map[string]interface{}{"name": tc.Name, "arguments": string(argsJSON)},
					})
				}
				m["tool_calls"] = tcs
			}
			oaiMsgs = append(oaiMsgs, m)
		case "tool":
			oaiMsgs = append(oaiMsgs, map[string]interface{}{
				"role": "tool", "tool_call_id": msg.ToolCallID, "content": msg.Content,
			})
		}
	}

	oaiTools := make([]map[string]interface{}, 0)
	for _, td := range toolDefs {
		oaiTools = append(oaiTools, map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name": td.Name, "description": td.Description, "parameters": td.Parameters,
			},
		})
	}

	reqBody := map[string]interface{}{
		"model": o.Model, "messages": oaiMsgs, "tools": oaiTools,
		"stream": true,
	}

	if strings.Contains(o.Model, "nvidia/") || o.ProviderName == "NVIDIA" {
		reqBody["reasoning_budget"] = 16384
		reqBody["chat_template_kwargs"] = map[string]interface{}{"enable_thinking": true}
	}

	payload, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.ApiKey)

	resp, err := o.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("%s error (%d): %s", o.ProviderName, resp.StatusCode, truncateBody(body))
	}

	ch := make(chan agent.StreamEvent, 32)
	go o.readSSE(resp, ch)
	return ch, nil
}

func (o *GenericOpenAIFCProvider) readSSE(resp *http.Response, ch chan<- agent.StreamEvent) {
	defer close(ch)
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 512*1024)

	tcAccum := make(map[int]*agent.ToolCall)
	var usage *agent.UsageStats

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk struct {
			Choices []struct {
				Delta struct {
					Content          *string `json:"content"`
					ReasoningContent *string `json:"reasoning_content"`
					Thought          *string `json:"thought"`
					ToolCalls        []struct {
						Index    int    `json:"index"`
						ID       string `json:"id"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls"`
				} `json:"delta"`
				FinishReason *string `json:"finish_reason"`
			} `json:"choices"`
			Usage *struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			} `json:"usage"`
		}

		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		if chunk.Usage != nil {
			usage = &agent.UsageStats{
				InputTokens:  chunk.Usage.PromptTokens,
				OutputTokens: chunk.Usage.CompletionTokens,
				TotalTokens:  chunk.Usage.TotalTokens,
			}
		}

		if len(chunk.Choices) == 0 {
			continue
		}

		delta := chunk.Choices[0].Delta
		if delta.ReasoningContent != nil && *delta.ReasoningContent != "" {
			ch <- agent.StreamEvent{Type: agent.StreamThought, Thought: *delta.ReasoningContent}
		}
		if delta.Thought != nil && *delta.Thought != "" {
			ch <- agent.StreamEvent{Type: agent.StreamThought, Thought: *delta.Thought}
		}
		if delta.Content != nil && *delta.Content != "" {
			ch <- agent.StreamEvent{Type: agent.StreamText, Text: *delta.Content}
		}

		for _, tc := range delta.ToolCalls {
			if _, ok := tcAccum[tc.Index]; !ok {
				tcAccum[tc.Index] = &agent.ToolCall{ID: tc.ID, Name: tc.Function.Name}
			}
			existing := tcAccum[tc.Index]
			if tc.ID != "" {
				existing.ID = tc.ID
			}
			if tc.Function.Name != "" {
				existing.Name = tc.Function.Name
			}
			if existing.Args == nil {
				existing.Args = map[string]interface{}{"_raw": tc.Function.Arguments}
			} else {
				existing.Args["_raw"] = existing.Args["_raw"].(string) + tc.Function.Arguments
			}
		}

		if chunk.Choices[0].FinishReason != nil {
			for _, tc := range tcAccum {
				if rawArgs, ok := tc.Args["_raw"].(string); ok {
					parsed := make(map[string]interface{})
					json.Unmarshal([]byte(rawArgs), &parsed)
					tc.Args = parsed
				}
				ch <- agent.StreamEvent{Type: agent.StreamToolCall, ToolCall: tc}
			}
			ch <- agent.StreamEvent{
				Type: agent.StreamDone, FinishReason: *chunk.Choices[0].FinishReason, Usage: usage,
			}
			return
		}
	}

	for _, tc := range tcAccum {
		if rawArgs, ok := tc.Args["_raw"].(string); ok {
			parsed := make(map[string]interface{})
			json.Unmarshal([]byte(rawArgs), &parsed)
			tc.Args = parsed
		}
		ch <- agent.StreamEvent{Type: agent.StreamToolCall, ToolCall: tc}
	}
	ch <- agent.StreamEvent{Type: agent.StreamDone, Usage: usage, FinishReason: "stop"}
}

func (c *AnthropicFCProvider) SupportsStreaming() bool { return true }

func (c *AnthropicFCProvider) SendWithToolsStream(ctx context.Context, systemPrompt string, messages []agent.Message, toolDefs []tools.ToolDefinition) (<-chan agent.StreamEvent, error) {
	url := "https://api.anthropic.com/v1/messages"

	claudeMsgs := make([]map[string]interface{}, 0)
	for _, msg := range messages {
		switch msg.Role {
		case "user":
			claudeMsgs = append(claudeMsgs, map[string]interface{}{
				"role": "user", "content": []map[string]interface{}{{"type": "text", "text": msg.Content}},
			})
		case "assistant":
			content := make([]map[string]interface{}, 0)
			if msg.Content != "" {
				content = append(content, map[string]interface{}{"type": "text", "text": msg.Content})
			}
			for _, tc := range msg.ToolCalls {
				content = append(content, map[string]interface{}{
					"type": "tool_use", "id": tc.ID, "name": tc.Name, "input": tc.Args,
				})
			}
			claudeMsgs = append(claudeMsgs, map[string]interface{}{"role": "assistant", "content": content})
		case "tool":
			claudeMsgs = append(claudeMsgs, map[string]interface{}{
				"role": "user", "content": []map[string]interface{}{
					{"type": "tool_result", "tool_use_id": msg.ToolCallID, "content": msg.Content},
				},
			})
		}
	}

	claudeTools := make([]map[string]interface{}, 0)
	for _, td := range toolDefs {
		claudeTools = append(claudeTools, map[string]interface{}{
			"name": td.Name, "description": td.Description, "input_schema": td.Parameters,
		})
	}

	reqBody := map[string]interface{}{
		"model": c.Model, "max_tokens": 8192, "system": systemPrompt,
		"messages": claudeMsgs, "tools": claudeTools, "stream": true,
	}

	payload, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.ApiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("Claude error (%d): %s", resp.StatusCode, truncateBody(body))
	}

	ch := make(chan agent.StreamEvent, 32)
	go c.readAnthropicSSE(resp, ch)
	return ch, nil
}

func (c *AnthropicFCProvider) readAnthropicSSE(resp *http.Response, ch chan<- agent.StreamEvent) {
	defer close(ch)
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 512*1024)

	var currentToolID, currentToolName string
	var toolArgsBuf strings.Builder
	var usage *agent.UsageStats

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")

		var event struct {
			Type  string `json:"type"`
			Index int    `json:"index"`
			Delta struct {
				Type        string `json:"type"`
				Text        string `json:"text"`
				PartialJSON string `json:"partial_json"`
			} `json:"delta"`
			ContentBlock struct {
				Type string `json:"type"`
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"content_block"`
			Usage *struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			} `json:"usage"`
		}

		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		switch event.Type {
		case "content_block_start":
			if event.ContentBlock.Type == "tool_use" {
				currentToolID = event.ContentBlock.ID
				currentToolName = event.ContentBlock.Name
				toolArgsBuf.Reset()
			}

		case "content_block_delta":
			if event.Delta.Type == "text_delta" && event.Delta.Text != "" {
				ch <- agent.StreamEvent{Type: agent.StreamText, Text: event.Delta.Text}
			}
			if event.Delta.Type == "input_json_delta" {
				toolArgsBuf.WriteString(event.Delta.PartialJSON)
			}

		case "content_block_stop":
			if currentToolID != "" {
				args := make(map[string]interface{})
				json.Unmarshal([]byte(toolArgsBuf.String()), &args)
				ch <- agent.StreamEvent{
					Type: agent.StreamToolCall,
					ToolCall: &agent.ToolCall{
						ID: currentToolID, Name: currentToolName, Args: args,
					},
				}
				currentToolID = ""
				currentToolName = ""
			}

		case "message_delta":
			if event.Usage != nil {
				if usage == nil {
					usage = &agent.UsageStats{}
				}
				usage.OutputTokens = event.Usage.OutputTokens
				usage.TotalTokens = usage.InputTokens + usage.OutputTokens
			}

		case "message_start":
			if event.Usage != nil {
				usage = &agent.UsageStats{
					InputTokens: event.Usage.InputTokens,
				}
			}

		case "message_stop":
			ch <- agent.StreamEvent{Type: agent.StreamDone, Usage: usage, FinishReason: "stop"}
			return
		}
	}

	ch <- agent.StreamEvent{Type: agent.StreamDone, Usage: usage, FinishReason: "stop"}
}

func (o *OllamaFCProvider) SupportsStreaming() bool { return true }

func (o *OllamaFCProvider) SendWithToolsStream(ctx context.Context, systemPrompt string, messages []agent.Message, toolDefs []tools.ToolDefinition) (<-chan agent.StreamEvent, error) {
	ollamaMsgs := []map[string]interface{}{
		{"role": "system", "content": systemPrompt},
	}
	for _, msg := range messages {
		switch msg.Role {
		case "user":
			ollamaMsgs = append(ollamaMsgs, map[string]interface{}{"role": "user", "content": msg.Content})
		case "assistant":
			m := map[string]interface{}{"role": "assistant", "content": msg.Content}
			if len(msg.ToolCalls) > 0 {
				tcs := make([]map[string]interface{}, 0)
				for _, tc := range msg.ToolCalls {
					tcs = append(tcs, map[string]interface{}{
						"function": map[string]interface{}{"name": tc.Name, "arguments": tc.Args},
					})
				}
				m["tool_calls"] = tcs
			}
			ollamaMsgs = append(ollamaMsgs, m)
		case "tool":
			ollamaMsgs = append(ollamaMsgs, map[string]interface{}{"role": "tool", "content": msg.Content})
		}
	}

	ollamaTools := make([]map[string]interface{}, 0)
	for _, td := range toolDefs {
		ollamaTools = append(ollamaTools, map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name": td.Name, "description": td.Description, "parameters": td.Parameters,
			},
		})
	}

	reqBody := map[string]interface{}{
		"model": o.Model, "messages": ollamaMsgs, "tools": ollamaTools, "stream": true,
	}

	payload, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", o.BaseURL, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("Ollama error (%d): %s", resp.StatusCode, truncateBody(body))
	}

	ch := make(chan agent.StreamEvent, 32)
	go o.readOllamaStream(resp, ch)
	return ch, nil
}

func (o *OllamaFCProvider) readOllamaStream(resp *http.Response, ch chan<- agent.StreamEvent) {
	defer close(ch)
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)

	for decoder.More() {
		var chunk struct {
			Message struct {
				Content   string `json:"content"`
				ToolCalls []struct {
					Function struct {
						Name      string                 `json:"name"`
						Arguments map[string]interface{} `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"message"`
			Done bool `json:"done"`
		}

		if err := decoder.Decode(&chunk); err != nil {
			break
		}

		if chunk.Message.Content != "" {
			ch <- agent.StreamEvent{Type: agent.StreamText, Text: chunk.Message.Content}
		}

		for _, tc := range chunk.Message.ToolCalls {
			ch <- agent.StreamEvent{
				Type: agent.StreamToolCall,
				ToolCall: &agent.ToolCall{
					ID:   fmt.Sprintf("%s-%d", tc.Function.Name, time.Now().UnixMilli()),
					Name: tc.Function.Name,
					Args: tc.Function.Arguments,
				},
			}
		}

		if chunk.Done {
			ch <- agent.StreamEvent{Type: agent.StreamDone, FinishReason: "stop"}
			return
		}
	}

	ch <- agent.StreamEvent{Type: agent.StreamDone, FinishReason: "stop"}
}
