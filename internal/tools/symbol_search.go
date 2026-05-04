package tools

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type SearchSymbolTool struct {
	WorkDir string
}

func NewSearchSymbolTool(workDir string) *SearchSymbolTool {
	return &SearchSymbolTool{WorkDir: workDir}
}

func (t *SearchSymbolTool) Name() string {
	return "search_symbol"
}

func (t *SearchSymbolTool) NeedsConfirmation(args map[string]interface{}) bool {
	return false
}

func (t *SearchSymbolTool) Description() string {
	return "Find code symbols (functions, structs, interfaces) in the project. Uses intelligent parsing for Go files."
}

func (t *SearchSymbolTool) ParameterSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "The name of the symbol to find (e.g. 'NewAgent'). If empty, lists all significant symbols.",
			},
		},
	}
}

type symbolMatch struct {
	Name     string
	Type     string
	File     string
	Line     int
	ModTime  time.Time
}

func (t *SearchSymbolTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	query, _ := params["query"].(string)

	var matches []symbolMatch

	// [EN] Step 1: Search in Go files using AST
	// [ID] Langkah 1: Cari di file Go menggunakan AST
	goMatches, err := t.searchGoSymbols(query)
	if err == nil {
		matches = append(matches, goMatches...)
	}

	// [EN] Step 2: Fallback to regex grep for other symbols or if no Go symbols found
	// [ID] Langkah 2: Fallback ke regex grep untuk simbol lain atau jika simbol Go tidak ditemukan
	if len(matches) == 0 {
		grepMatches := t.searchRegexFallback(query)
		matches = append(matches, grepMatches...)
	}

	// [EN] Step 3: Sort by ModTime (newest first)
	// [ID] Langkah 3: Urutkan berdasarkan ModTime (terbaru dulu)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].ModTime.After(matches[j].ModTime)
	})

	if len(matches) == 0 {
		return &ToolResult{
			Content: fmt.Sprintf("No symbols matching '%s' found.", query),
			Display: fmt.Sprintf("[S] No symbols for '%s'", query),
		}, nil
	}

	// [EN] Step 4: Format results
	// [ID] Langkah 4: Format hasil
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d symbol(s) matching '%s':\n\n", len(matches), query))
	
	// Limit to 50 results for context efficiency
	displayCount := len(matches)
	if displayCount > 50 {
		displayCount = 50
	}

	for i := 0; i < displayCount; i++ {
		m := matches[i]
		sb.WriteString(fmt.Sprintf("%s:%d [%s] %s\n", m.File, m.Line, m.Type, m.Name))
	}

	if len(matches) > 50 {
		sb.WriteString(fmt.Sprintf("\n... (truncated %d more results)", len(matches)-50))
	}

	// [EN] Step 5: Extract symbol names for agent memory
	// [ID] Langkah 5: Ekstrak nama simbol untuk memori agen
	symbolNames := make([]string, 0)
	for i := 0; i < displayCount; i++ {
		symbolNames = append(symbolNames, matches[i].Name)
	}

	return &ToolResult{
		Content: sb.String(),
		Display: fmt.Sprintf("[S] Found %d symbols for '%s'", len(matches), query),
		Metadata: map[string]interface{}{
			"symbols": symbolNames,
		},
	}, nil
}

func (t *SearchSymbolTool) searchGoSymbols(query string) ([]symbolMatch, error) {
	var matches []symbolMatch
	fset := token.NewFileSet()
	queryLower := strings.ToLower(query)

	skipDirs := map[string]bool{
		".git": true, "node_modules": true, "vendor": true, "dist": true, "build": true,
	}

	err := filepath.Walk(t.WorkDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			if info != nil && info.IsDir() && skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Parse the file
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(t.WorkDir, path)

		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.FuncDecl:
				if strings.Contains(strings.ToLower(x.Name.Name), queryLower) {
					matches = append(matches, symbolMatch{
						Name:    x.Name.Name,
						Type:    "func",
						File:    relPath,
						Line:    fset.Position(x.Pos()).Line,
						ModTime: info.ModTime(),
					})
				}
			case *ast.TypeSpec:
				if strings.Contains(strings.ToLower(x.Name.Name), queryLower) {
					symbolType := "type"
					switch x.Type.(type) {
					case *ast.StructType:
						symbolType = "struct"
					case *ast.InterfaceType:
						symbolType = "interface"
					}
					matches = append(matches, symbolMatch{
						Name:    x.Name.Name,
						Type:    symbolType,
						File:    relPath,
						Line:    fset.Position(x.Pos()).Line,
						ModTime: info.ModTime(),
					})
				}
			}
			return true
		})

		return nil
	})

	return matches, err
}

func (t *SearchSymbolTool) searchRegexFallback(query string) []symbolMatch {
	patterns := []string{
		fmt.Sprintf(`func\s+.*%s`, query),
		fmt.Sprintf(`type\s+.*%s`, query),
		fmt.Sprintf(`class\s+.*%s`, query),
		fmt.Sprintf(`interface\s+.*%s`, query),
		fmt.Sprintf(`def\s+.*%s`, query),
	}

	fullRegex := "(" + strings.Join(patterns, "|") + ")"
	var matches []symbolMatch

	// Using the system's grep as a fast fallback
	args := []string{"-rnE", fullRegex, "."}
	cmd := exec.Command("grep", args...)
	cmd.Dir = t.WorkDir
	out, _ := cmd.Output()

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 3)
		if len(parts) >= 3 {
			file := parts[0]
			lineNum := 0
			fmt.Sscanf(parts[1], "%d", &lineNum)
			
			info, err := os.Stat(filepath.Join(t.WorkDir, file))
			mtime := time.Now()
			if err == nil {
				mtime = info.ModTime()
			}

			matches = append(matches, symbolMatch{
				Name:    strings.TrimSpace(parts[2]),
				Type:    "regex",
				File:    file,
				Line:    lineNum,
				ModTime: mtime,
			})
		}
	}

	return matches
}
