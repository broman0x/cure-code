package tools

import (
	"context"
	"fmt"
	"sync"
)

// [EN] Tool defines the interface that all CuRe Code tools must implement.
// [ID] Tool mendefinisikan antarmuka yang harus diimplementasikan oleh semua tool CuRe Code.
type Tool interface {
	// [EN] Name returns the unique identifier for the tool.
	// [ID] Name mengembalikan pengenal unik untuk tool tersebut.
	Name() string

	// [EN] Description returns a clear explanation of what the tool does.
	// [ID] Description mengembalikan penjelasan jelas tentang apa yang dilakukan tool tersebut.
	Description() string

	// [EN] ParameterSchema returns a JSON schema for the tool's parameters.
	// [ID] ParameterSchema mengembalikan skema JSON untuk parameter tool tersebut.
	ParameterSchema() map[string]interface{}

	// [EN] Execute runs the tool logic with the provided parameters.
	// [ID] Execute menjalankan logika tool dengan parameter yang diberikan.
	Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error)

	// [EN] NeedsConfirmation returns true if the tool requires user approval before execution.
	// [ID] NeedsConfirmation mengembalikan true jika tool memerlukan persetujuan pengguna sebelum eksekusi.
	NeedsConfirmation(params map[string]interface{}) bool
}

type ToolResult struct {
	Content string

	Display string

	IsError bool

	FilesModified []string

	BackgroundCmd interface{}
}

type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// [EN] ToolRegistry manages the collection of available tools for the agent.
// [ID] ToolRegistry mengelola kumpulan tool yang tersedia untuk agen.
type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]Tool
	order []string
}

func NewRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
		order: make([]string, 0),
	}
}

func (r *ToolRegistry) Register(t Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := t.Name()
	if _, exists := r.tools[name]; !exists {
		r.order = append(r.order, name)
	}
	r.tools[name] = t
}

func (r *ToolRegistry) Get(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[name]
	return t, ok
}

func (r *ToolRegistry) All() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Tool, 0, len(r.order))
	for _, name := range r.order {
		if t, ok := r.tools[name]; ok {
			result = append(result, t)
		}
	}
	return result
}

func (r *ToolRegistry) Definitions() []ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	defs := make([]ToolDefinition, 0, len(r.order))
	for _, name := range r.order {
		t := r.tools[name]
		defs = append(defs, ToolDefinition{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters:  t.ParameterSchema(),
		})
	}
	return defs
}

func (r *ToolRegistry) Execute(ctx context.Context, name string, params map[string]interface{}) (*ToolResult, error) {
	t, ok := r.Get(name)
	if !ok {
		return &ToolResult{
			Content: fmt.Sprintf("Error: unknown tool '%s'", name),
			IsError: true,
		}, fmt.Errorf("unknown tool: %s", name)
	}
	return t.Execute(ctx, params)
}

// [EN] NewDefaultRegistry creates a registry and registers all built-in tools.
// [ID] NewDefaultRegistry membuat registry dan mendaftarkan semua tool bawaan.
func NewDefaultRegistry(workDir string) *ToolRegistry {
	r := NewRegistry()

	r.Register(NewReadFileTool(workDir))
	r.Register(NewReadManyFilesTool(workDir))
	r.Register(NewWriteFileTool(workDir))
	r.Register(NewEditFileTool(workDir))
	r.Register(NewShellTool(workDir))
	r.Register(NewListDirTool(workDir))
	r.Register(NewGrepTool(workDir))
	r.Register(NewGlobTool(workDir))
	r.Register(NewAskUserTool())
	r.Register(NewWebFetchTool())
	r.Register(NewWebSearchTool())
	r.Register(NewProjectSummaryTool(workDir))
	r.Register(NewGitInfoTool(workDir))
	r.Register(NewSearchSymbolTool(workDir))
	r.Register(&TodoTool{})

	return r
}
