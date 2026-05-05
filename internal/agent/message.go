package agent

type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	Thought    string     `json:"thought,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Name       string     `json:"name,omitempty"`
}

type ToolCall struct {
	ID   string                 `json:"id"`
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"arguments"`
}

type Response struct {
	Content      string
	Thought      string
	ToolCalls    []ToolCall
	FinishReason string
	Usage        *UsageStats
}

type StreamEvent struct {
	Type StreamEventType

	Text string

	Thought string

	ToolCall *ToolCall

	Error error

	Usage *UsageStats

	FinishReason string
}

type StreamEventType int

const (
	StreamText StreamEventType = iota

	StreamToolCall

	StreamDone

	StreamError

	StreamThought
)

type UsageStats struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

type SessionUsage struct {
	TotalInputTokens  int `json:"total_input_tokens"`
	TotalOutputTokens int `json:"total_output_tokens"`
	TotalTokens       int `json:"total_tokens"`
	RequestCount      int `json:"request_count"`
}

func (su *SessionUsage) Add(usage *UsageStats) {
	if usage == nil {
		return
	}
	su.TotalInputTokens += usage.InputTokens
	su.TotalOutputTokens += usage.OutputTokens
	su.TotalTokens += usage.TotalTokens
	su.RequestCount++
}

type Task struct {
	Description string `json:"description"`
	Status      string `json:"status"`
}
