package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/broman0x/cure-code/internal/agent"
	"github.com/broman0x/cure-code/internal/tools"
)

var (
	_ agent.StreamingProvider = (*GeminiFCProvider)(nil)
	_ agent.StreamingProvider = (*OpenAIFCProvider)(nil)
	_ agent.StreamingProvider = (*AnthropicFCProvider)(nil)
	_ agent.StreamingProvider = (*OllamaFCProvider)(nil)
)

// [EN] GeminiFCProvider implements the FunctionCallingProvider for Google's Gemini models.
// [ID] GeminiFCProvider mengimplementasikan FunctionCallingProvider untuk model Gemini Google.
type GeminiFCProvider struct {
	ApiKey string
	Model  string
	Client *http.Client
}

func NewGeminiFCProvider(apiKey, model string) *GeminiFCProvider {
	return &GeminiFCProvider{
		ApiKey: apiKey,
		Model:  model,
		Client: &http.Client{Timeout: 180 * time.Second},
	}
}

func (g *GeminiFCProvider) Name() string        { return "Gemini (" + g.Model + ")" }
func (g *GeminiFCProvider) SupportsTools() bool { return true }

func (g *GeminiFCProvider) SendWithTools(systemPrompt string, messages []agent.Message, toolDefs []tools.ToolDefinition) (*agent.Response, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", g.Model, g.ApiKey)

	reqBody := g.buildRequest(systemPrompt, messages, toolDefs)

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal error: %v", err)
	}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, truncateBody(body))
	}

	return g.parseResponse(body)
}

func (g *GeminiFCProvider) buildRequest(systemPrompt string, messages []agent.Message, toolDefs []tools.ToolDefinition) map[string]interface{} {

	contents := make([]map[string]interface{}, 0)

	for _, msg := range messages {
		switch msg.Role {
		case "user":
			contents = append(contents, map[string]interface{}{
				"role":  "user",
				"parts": []map[string]interface{}{{"text": msg.Content}},
			})

		case "assistant":
			parts := make([]map[string]interface{}, 0)
			if msg.Content != "" {
				parts = append(parts, map[string]interface{}{"text": msg.Content})
			}
			for _, tc := range msg.ToolCalls {
				parts = append(parts, map[string]interface{}{
					"functionCall": map[string]interface{}{
						"name": tc.Name,
						"args": tc.Args,
					},
				})
			}
			if len(parts) > 0 {
				contents = append(contents, map[string]interface{}{
					"role":  "model",
					"parts": parts,
				})
			}

		case "tool":
			contents = append(contents, map[string]interface{}{
				"role": "user",
				"parts": []map[string]interface{}{
					{
						"functionResponse": map[string]interface{}{
							"name":     msg.Name,
							"response": map[string]interface{}{"content": msg.Content},
						},
					},
				},
			})
		}
	}

	funcDecls := make([]map[string]interface{}, 0, len(toolDefs))
	for _, td := range toolDefs {
		funcDecls = append(funcDecls, map[string]interface{}{
			"name":        td.Name,
			"description": td.Description,
			"parameters":  td.Parameters,
		})
	}

	reqBody := map[string]interface{}{
		"contents": contents,
		"systemInstruction": map[string]interface{}{
			"parts": []map[string]interface{}{{"text": systemPrompt}},
		},
		"tools": []map[string]interface{}{
			{"functionDeclarations": funcDecls},
		},
		"generationConfig": map[string]interface{}{
			"maxOutputTokens": 8192,
			"temperature":     0.7,
		},
	}

	return reqBody
}

func (g *GeminiFCProvider) parseResponse(body []byte) (*agent.Response, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse error: %s", truncateBody(body))
	}

	if errObj, ok := raw["error"]; ok {
		if errMap, ok := errObj.(map[string]interface{}); ok {
			return nil, fmt.Errorf("API error: %v", errMap["message"])
		}
	}

	resp := &agent.Response{}

	candidates, ok := raw["candidates"].([]interface{})
	if !ok || len(candidates) == 0 {
		return nil, fmt.Errorf("empty response from API")
	}

	candidate := candidates[0].(map[string]interface{})
	content, ok := candidate["content"].(map[string]interface{})
	if !ok {
		return resp, nil
	}

	parts, ok := content["parts"].([]interface{})
	if !ok {
		return resp, nil
	}

	for _, part := range parts {
		partMap, ok := part.(map[string]interface{})
		if !ok {
			continue
		}

		if text, ok := partMap["text"].(string); ok {
			resp.Content += text
		}

		if fc, ok := partMap["functionCall"].(map[string]interface{}); ok {
			name, _ := fc["name"].(string)
			args := make(map[string]interface{})
			if a, ok := fc["args"].(map[string]interface{}); ok {
				args = a
			}
			resp.ToolCalls = append(resp.ToolCalls, agent.ToolCall{
				ID:   fmt.Sprintf("%s-%d", name, time.Now().UnixMilli()),
				Name: name,
				Args: args,
			})
		}
	}

	if fr, ok := candidate["finishReason"].(string); ok {
		resp.FinishReason = fr
	}
	if len(resp.ToolCalls) > 0 {
		resp.FinishReason = "tool_calls"
	}

	if um, ok := raw["usageMetadata"].(map[string]interface{}); ok {
		resp.Usage = &agent.UsageStats{}
		if v, ok := um["promptTokenCount"].(float64); ok {
			resp.Usage.InputTokens = int(v)
		}
		if v, ok := um["candidatesTokenCount"].(float64); ok {
			resp.Usage.OutputTokens = int(v)
		}
		if v, ok := um["totalTokenCount"].(float64); ok {
			resp.Usage.TotalTokens = int(v)
		}
	}

	return resp, nil
}

// [EN] OpenAIFCProvider implements the FunctionCallingProvider for OpenAI's models.
// [ID] OpenAIFCProvider mengimplementasikan FunctionCallingProvider untuk model OpenAI.
type OpenAIFCProvider struct {
	ApiKey string
	Model  string
	Client *http.Client
}

func NewOpenAIFCProvider(apiKey, model string) *OpenAIFCProvider {
	return &OpenAIFCProvider{
		ApiKey: apiKey,
		Model:  model,
		Client: &http.Client{Timeout: 180 * time.Second},
	}
}

func (o *OpenAIFCProvider) Name() string        { return "OpenAI (" + o.Model + ")" }
func (o *OpenAIFCProvider) SupportsTools() bool { return true }

func (o *OpenAIFCProvider) SendWithTools(systemPrompt string, messages []agent.Message, toolDefs []tools.ToolDefinition) (*agent.Response, error) {
	url := "https://api.openai.com/v1/chat/completions"

	oaiMsgs := []map[string]interface{}{
		{"role": "system", "content": systemPrompt},
	}

	for _, msg := range messages {
		switch msg.Role {
		case "user":
			oaiMsgs = append(oaiMsgs, map[string]interface{}{
				"role": "user", "content": msg.Content,
			})
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
						"id":   tc.ID,
						"type": "function",
						"function": map[string]interface{}{
							"name":      tc.Name,
							"arguments": string(argsJSON),
						},
					})
				}
				m["tool_calls"] = tcs
			}
			oaiMsgs = append(oaiMsgs, m)
		case "tool":
			oaiMsgs = append(oaiMsgs, map[string]interface{}{
				"role":         "tool",
				"tool_call_id": msg.ToolCallID,
				"content":      msg.Content,
			})
		}
	}

	oaiTools := make([]map[string]interface{}, 0)
	for _, td := range toolDefs {
		oaiTools = append(oaiTools, map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        td.Name,
				"description": td.Description,
				"parameters":  td.Parameters,
			},
		})
	}

	reqBody := map[string]interface{}{
		"model":    o.Model,
		"messages": oaiMsgs,
		"tools":    oaiTools,
	}

	if strings.Contains(o.Model, "nvidia/") {
		reqBody["reasoning_budget"] = 16384
		reqBody["chat_template_kwargs"] = map[string]interface{}{"enable_thinking": true}
	}

	payload, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.ApiKey)

	resp, err := o.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("OpenAI error (%d): %s", resp.StatusCode, truncateBody(body))
	}

	return o.parseResponse(body)
}

func (o *OpenAIFCProvider) parseResponse(body []byte) (*agent.Response, error) {
	var raw struct {
		Choices []struct {
			Message struct {
				Content   *string `json:"content"`
				ToolCalls []struct {
					ID       string `json:"id"`
					Function struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
		Usage *struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse error: %v", err)
	}
	if raw.Error != nil {
		return nil, fmt.Errorf("OpenAI error: %s", raw.Error.Message)
	}
	if len(raw.Choices) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	choice := raw.Choices[0]
	resp := &agent.Response{FinishReason: choice.FinishReason}

	if choice.Message.Content != nil {
		resp.Content = *choice.Message.Content
	}

	for _, tc := range choice.Message.ToolCalls {
		args := make(map[string]interface{})
		json.Unmarshal([]byte(tc.Function.Arguments), &args)
		resp.ToolCalls = append(resp.ToolCalls, agent.ToolCall{
			ID:   tc.ID,
			Name: tc.Function.Name,
			Args: args,
		})
	}

	if raw.Usage != nil {
		resp.Usage = &agent.UsageStats{
			InputTokens:  raw.Usage.PromptTokens,
			OutputTokens: raw.Usage.CompletionTokens,
			TotalTokens:  raw.Usage.TotalTokens,
		}
	}

	return resp, nil
}

type GenericOpenAIFCProvider struct {
	ApiKey       string
	Model        string
	BaseURL      string
	ProviderName string
	Client       *http.Client
}

func NewGenericOpenAIFCProvider(apiKey, model, baseURL, providerName string) *GenericOpenAIFCProvider {
	return &GenericOpenAIFCProvider{
		ApiKey:       apiKey,
		Model:        model,
		BaseURL:      baseURL,
		ProviderName: providerName,
		Client:       &http.Client{Timeout: 180 * time.Second},
	}
}

func (o *GenericOpenAIFCProvider) Name() string        { return o.ProviderName + " (" + o.Model + ")" }
func (o *GenericOpenAIFCProvider) SupportsTools() bool { return true }

func (o *GenericOpenAIFCProvider) SendWithTools(systemPrompt string, messages []agent.Message, toolDefs []tools.ToolDefinition) (*agent.Response, error) {
	url := strings.TrimSuffix(o.BaseURL, "/") + "/chat/completions"

	oaiMsgs := []map[string]interface{}{
		{"role": "system", "content": systemPrompt},
	}

	for _, msg := range messages {
		switch msg.Role {
		case "user":
			oaiMsgs = append(oaiMsgs, map[string]interface{}{
				"role": "user", "content": msg.Content,
			})
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
						"id":   tc.ID,
						"type": "function",
						"function": map[string]interface{}{
							"name":      tc.Name,
							"arguments": string(argsJSON),
						},
					})
				}
				m["tool_calls"] = tcs
			}
			oaiMsgs = append(oaiMsgs, m)
		case "tool":
			oaiMsgs = append(oaiMsgs, map[string]interface{}{
				"role":         "tool",
				"tool_call_id": msg.ToolCallID,
				"content":      msg.Content,
			})
		}
	}

	oaiTools := make([]map[string]interface{}, 0)
	for _, td := range toolDefs {
		oaiTools = append(oaiTools, map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        td.Name,
				"description": td.Description,
				"parameters":  td.Parameters,
			},
		})
	}

	reqBody := map[string]interface{}{
		"model":    o.Model,
		"messages": oaiMsgs,
		"tools":    oaiTools,
	}

	if strings.Contains(o.Model, "nvidia/") || o.ProviderName == "NVIDIA" {
		reqBody["reasoning_budget"] = 16384
		reqBody["chat_template_kwargs"] = map[string]interface{}{"enable_thinking": true}
	}

	payload, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.ApiKey)

	resp, err := o.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s error (%d): %s", o.ProviderName, resp.StatusCode, truncateBody(body))
	}

	return o.parseResponse(body)
}

func (o *GenericOpenAIFCProvider) parseResponse(body []byte) (*agent.Response, error) {
	var raw struct {
		Choices []struct {
			Message struct {
				Content   *string `json:"content"`
				ToolCalls []struct {
					ID       string `json:"id"`
					Function struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
		Usage *struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse error: %v", err)
	}
	if raw.Error != nil {
		return nil, fmt.Errorf("%s error: %s", o.ProviderName, raw.Error.Message)
	}
	if len(raw.Choices) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	choice := raw.Choices[0]
	resp := &agent.Response{FinishReason: choice.FinishReason}

	if choice.Message.Content != nil {
		resp.Content = *choice.Message.Content
	}

	for _, tc := range choice.Message.ToolCalls {
		args := make(map[string]interface{})
		json.Unmarshal([]byte(tc.Function.Arguments), &args)
		resp.ToolCalls = append(resp.ToolCalls, agent.ToolCall{
			ID:   tc.ID,
			Name: tc.Function.Name,
			Args: args,
		})
	}

	if raw.Usage != nil {
		resp.Usage = &agent.UsageStats{
			InputTokens:  raw.Usage.PromptTokens,
			OutputTokens: raw.Usage.CompletionTokens,
			TotalTokens:  raw.Usage.TotalTokens,
		}
	}

	return resp, nil
}

// [EN] AnthropicFCProvider implements the FunctionCallingProvider for Anthropic's Claude models.
// [ID] AnthropicFCProvider mengimplementasikan FunctionCallingProvider untuk model Claude Anthropic.
type AnthropicFCProvider struct {
	ApiKey string
	Model  string
	Client *http.Client
}

func NewAnthropicFCProvider(apiKey, model string) *AnthropicFCProvider {
	return &AnthropicFCProvider{
		ApiKey: apiKey,
		Model:  model,
		Client: &http.Client{Timeout: 180 * time.Second},
	}
}

func (c *AnthropicFCProvider) Name() string        { return "Claude (" + c.Model + ")" }
func (c *AnthropicFCProvider) SupportsTools() bool { return true }

func (c *AnthropicFCProvider) SendWithTools(systemPrompt string, messages []agent.Message, toolDefs []tools.ToolDefinition) (*agent.Response, error) {
	url := "https://api.anthropic.com/v1/messages"

	claudeMsgs := make([]map[string]interface{}, 0)
	for _, msg := range messages {
		switch msg.Role {
		case "user":
			claudeMsgs = append(claudeMsgs, map[string]interface{}{
				"role": "user",
				"content": []map[string]interface{}{
					{"type": "text", "text": msg.Content},
				},
			})
		case "assistant":
			content := make([]map[string]interface{}, 0)
			if msg.Content != "" {
				content = append(content, map[string]interface{}{
					"type": "text", "text": msg.Content,
				})
			}
			for _, tc := range msg.ToolCalls {
				content = append(content, map[string]interface{}{
					"type":  "tool_use",
					"id":    tc.ID,
					"name":  tc.Name,
					"input": tc.Args,
				})
			}
			claudeMsgs = append(claudeMsgs, map[string]interface{}{
				"role": "assistant", "content": content,
			})
		case "tool":
			claudeMsgs = append(claudeMsgs, map[string]interface{}{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type":        "tool_result",
						"tool_use_id": msg.ToolCallID,
						"content":     msg.Content,
					},
				},
			})
		}
	}

	claudeTools := make([]map[string]interface{}, 0)
	for _, td := range toolDefs {
		claudeTools = append(claudeTools, map[string]interface{}{
			"name":         td.Name,
			"description":  td.Description,
			"input_schema": td.Parameters,
		})
	}

	reqBody := map[string]interface{}{
		"model":      c.Model,
		"max_tokens": 8192,
		"system":     systemPrompt,
		"messages":   claudeMsgs,
		"tools":      claudeTools,
	}

	payload, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.ApiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Claude error (%d): %s", resp.StatusCode, truncateBody(body))
	}

	return c.parseResponse(body)
}

func (c *AnthropicFCProvider) parseResponse(body []byte) (*agent.Response, error) {
	var raw struct {
		Content []struct {
			Type  string                 `json:"type"`
			Text  string                 `json:"text,omitempty"`
			ID    string                 `json:"id,omitempty"`
			Name  string                 `json:"name,omitempty"`
			Input map[string]interface{} `json:"input,omitempty"`
		} `json:"content"`
		StopReason string `json:"stop_reason"`
		Error      *struct {
			Message string `json:"message"`
		} `json:"error"`
		Usage *struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse error: %v", err)
	}
	if raw.Error != nil {
		return nil, fmt.Errorf("Claude error: %s", raw.Error.Message)
	}

	resp := &agent.Response{FinishReason: raw.StopReason}

	for _, block := range raw.Content {
		switch block.Type {
		case "text":
			resp.Content += block.Text
		case "tool_use":
			resp.ToolCalls = append(resp.ToolCalls, agent.ToolCall{
				ID:   block.ID,
				Name: block.Name,
				Args: block.Input,
			})
		}
	}

	if raw.Usage != nil {
		resp.Usage = &agent.UsageStats{
			InputTokens:  raw.Usage.InputTokens,
			OutputTokens: raw.Usage.OutputTokens,
			TotalTokens:  raw.Usage.InputTokens + raw.Usage.OutputTokens,
		}
	}

	return resp, nil
}

// [EN] OllamaFCProvider implements the FunctionCallingProvider for local models running via Ollama.
// [ID] OllamaFCProvider mengimplementasikan FunctionCallingProvider untuk model lokal yang berjalan via Ollama.
type OllamaFCProvider struct {
	BaseURL string
	Model   string
	Client  *http.Client
}

func NewOllamaFCProvider(model string) *OllamaFCProvider {
	host := os.Getenv("OLLAMA_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("OLLAMA_PORT")
	if port == "" {
		port = "11434"
	}
	return &OllamaFCProvider{
		BaseURL: fmt.Sprintf("http://%s:%s/api/chat", host, port),
		Model:   model,
		Client:  &http.Client{Timeout: 300 * time.Second},
	}
}

func (o *OllamaFCProvider) Name() string        { return "Ollama (" + o.Model + ")" }
func (o *OllamaFCProvider) SupportsTools() bool { return true }

func (o *OllamaFCProvider) SendWithTools(systemPrompt string, messages []agent.Message, toolDefs []tools.ToolDefinition) (*agent.Response, error) {

	ollamaMsgs := []map[string]interface{}{
		{"role": "system", "content": systemPrompt},
	}
	for _, msg := range messages {
		switch msg.Role {
		case "user":
			ollamaMsgs = append(ollamaMsgs, map[string]interface{}{
				"role": "user", "content": msg.Content,
			})
		case "assistant":
			m := map[string]interface{}{"role": "assistant", "content": msg.Content}
			if len(msg.ToolCalls) > 0 {
				tcs := make([]map[string]interface{}, 0)
				for _, tc := range msg.ToolCalls {
					tcs = append(tcs, map[string]interface{}{
						"function": map[string]interface{}{
							"name":      tc.Name,
							"arguments": tc.Args,
						},
					})
				}
				m["tool_calls"] = tcs
			}
			ollamaMsgs = append(ollamaMsgs, m)
		case "tool":
			ollamaMsgs = append(ollamaMsgs, map[string]interface{}{
				"role": "tool", "content": msg.Content,
			})
		}
	}

	ollamaTools := make([]map[string]interface{}, 0)
	for _, td := range toolDefs {
		ollamaTools = append(ollamaTools, map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        td.Name,
				"description": td.Description,
				"parameters":  td.Parameters,
			},
		})
	}

	reqBody := map[string]interface{}{
		"model":    o.Model,
		"messages": ollamaMsgs,
		"tools":    ollamaTools,
		"stream":   false,
	}

	payload, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", o.BaseURL, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Ollama error (%d): %s", resp.StatusCode, truncateBody(body))
	}

	return o.parseResponse(body)
}

func (o *OllamaFCProvider) parseResponse(body []byte) (*agent.Response, error) {
	var raw struct {
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

	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse error: %v", err)
	}

	resp := &agent.Response{
		Content:      strings.TrimSpace(raw.Message.Content),
		FinishReason: "stop",
	}

	for _, tc := range raw.Message.ToolCalls {
		resp.ToolCalls = append(resp.ToolCalls, agent.ToolCall{
			ID:   fmt.Sprintf("%s-%d", tc.Function.Name, time.Now().UnixMilli()),
			Name: tc.Function.Name,
			Args: tc.Function.Arguments,
		})
	}

	if len(resp.ToolCalls) > 0 {
		resp.FinishReason = "tool_calls"
	}

	return resp, nil
}

func CreateFCProvider(pType, modelName string) (agent.FunctionCallingProvider, error) {
	pType = strings.ToLower(pType)

	switch pType {
	case "gemini":
		key := os.Getenv("GEMINI_API_KEY")
		if key == "" {
			return nil, fmt.Errorf("GEMINI_API_KEY not found")
		}
		if modelName == "" {
			modelName = "gemini-2.5-flash"
		}
		return NewGeminiFCProvider(key, modelName), nil

	case "openai", "chatgpt":
		key := os.Getenv("OPENAI_API_KEY")
		if key == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY not found")
		}
		if modelName == "" {
			modelName = "gpt-4o-mini"
		}
		return NewOpenAIFCProvider(key, modelName), nil

	case "claude", "anthropic":
		key := os.Getenv("ANTHROPIC_API_KEY")
		if key == "" {
			return nil, fmt.Errorf("ANTHROPIC_API_KEY not found")
		}
		if modelName == "" {
			modelName = "claude-sonnet-4-20250514"
		}
		return NewAnthropicFCProvider(key, modelName), nil

	case "ollama":
		if modelName == "" {
			modelName = "llama3"
		}
		return NewOllamaFCProvider(modelName), nil

	case "nvidia":
		key := os.Getenv("NVIDIA_API_KEY")
		if key == "" {
			return nil, fmt.Errorf("NVIDIA_API_KEY not found")
		}
		if modelName == "" {
			modelName = "nvidia/nemotron-3-super-120b-a12b"
		}
		return NewGenericOpenAIFCProvider(key, modelName, "https://integrate.api.nvidia.com/v1", "NVIDIA"), nil

	case "groq":
		key := os.Getenv("GROQ_API_KEY")
		if key == "" {
			return nil, fmt.Errorf("GROQ_API_KEY not found")
		}
		if modelName == "" {
			modelName = "llama-3.1-70b-versatile"
		}
		return NewGenericOpenAIFCProvider(key, modelName, "https://api.groq.com/openai/v1", "Groq"), nil

	case "deepseek":
		key := os.Getenv("DEEPSEEK_API_KEY")
		if key == "" {
			return nil, fmt.Errorf("DEEPSEEK_API_KEY not found")
		}
		if modelName == "" {
			modelName = "deepseek-coder"
		}
		return NewGenericOpenAIFCProvider(key, modelName, "https://api.deepseek.com", "DeepSeek"), nil

	case "together":
		key := os.Getenv("TOGETHER_API_KEY")
		if key == "" {
			return nil, fmt.Errorf("TOGETHER_API_KEY not found")
		}
		if modelName == "" {
			modelName = "meta-llama/Meta-Llama-3.1-70B-Instruct-Turbo"
		}
		return NewGenericOpenAIFCProvider(key, modelName, "https://api.together.xyz/v1", "Together"), nil

	case "mistral":
		key := os.Getenv("MISTRAL_API_KEY")
		if key == "" {
			return nil, fmt.Errorf("MISTRAL_API_KEY not found")
		}
		if modelName == "" {
			modelName = "mistral-large-latest"
		}
		return NewGenericOpenAIFCProvider(key, modelName, "https://api.mistral.ai/v1", "Mistral"), nil

	default:
		return nil, fmt.Errorf("unknown provider: %s", pType)
	}
}

func truncateBody(body []byte) string {
	s := string(body)
	if len(s) > 500 {
		return s[:500] + "..."
	}
	return s
}
