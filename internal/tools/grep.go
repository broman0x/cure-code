package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

type GrepTool struct {
	workDir string
}

func NewGrepTool(workDir string) *GrepTool {
	return &GrepTool{workDir: workDir}
}

func (t *GrepTool) Name() string { return "grep_search" }

func (t *GrepTool) Description() string {
	return `Search for a text pattern or regex in files within a directory.
Returns matching lines with file paths and line numbers.
Useful for finding function definitions, imports, error messages, or any text pattern in the codebase.`
}

func (t *GrepTool) ParameterSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "The text or regex pattern to search for.",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "The directory or file path to search in. Defaults to '.' (working directory).",
			},
			"include": map[string]interface{}{
				"type":        "string",
				"description": "File glob pattern to include (e.g., '*.go', '*.js'). If empty, searches all files.",
			},
			"case_sensitive": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether the search is case-sensitive. Defaults to true.",
			},
		},
		"required": []string{"pattern"},
	}
}

func (t *GrepTool) NeedsConfirmation(params map[string]interface{}) bool {
	return false
}

func (t *GrepTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	pattern, ok := getStringParam(params, "pattern")
	if !ok || pattern == "" {
		return &ToolResult{Content: "Error: pattern is required", IsError: true}, nil
	}

	searchPath := "."
	if p, ok := getStringParam(params, "path"); ok && p != "" {
		searchPath = p
	}

	includeGlob, _ := getStringParam(params, "include")
	caseSensitive := true
	if cs, ok := getBoolParam(params, "case_sensitive"); ok {
		caseSensitive = cs
	}

	absPath := t.resolvePath(searchPath)

	flags := ""
	if !caseSensitive {
		flags = "(?i)"
	}
	re, err := regexp.Compile(flags + pattern)
	if err != nil {

		escaped := regexp.QuoteMeta(pattern)
		re, _ = regexp.Compile(flags + escaped)
	}

	type match struct {
		file    string
		lineNum int
		line    string
	}

	var matches []match
	var mu sync.Mutex
	maxMatches := 100

	type task struct {
		path string
		info os.FileInfo
	}
	taskCh := make(chan task, 100)
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	numWorkers := runtime.NumCPU() * 2
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range taskCh {
				select {
				case <-ctx.Done():
					return
				default:
					f, err := os.Open(t.path)
					if err != nil {
						continue
					}
					scanner := bufio.NewScanner(f)
					lineNum := 0
					for scanner.Scan() {
						lineNum++
						line := scanner.Text()
						if re.MatchString(line) {
							relPath, _ := filepath.Rel(absPath, t.path)
							mu.Lock()
							if len(matches) < maxMatches {
								matches = append(matches, match{
									file:    relPath,
									lineNum: lineNum,
									line:    truncateStr(strings.TrimSpace(line), 200),
								})
								if len(matches) >= maxMatches {
									cancel()
								}
							}
							mu.Unlock()
						}
					}
					f.Close()
				}
			}
		}()
	}

	skipDirs := map[string]bool{
		".git": true, "node_modules": true, "vendor": true,
		"__pycache__": true, ".next": true, "dist": true,
		"build": true, ".idea": true, ".vscode": true,
	}

	go func() {
		defer close(taskCh)
		filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			select {
			case <-ctx.Done():
				return filepath.SkipAll
			default:
				if info.IsDir() {
					if skipDirs[info.Name()] {
						return filepath.SkipDir
					}
					return nil
				}

				if includeGlob != "" {
					matched, _ := filepath.Match(includeGlob, info.Name())
					if !matched {
						return nil
					}
				}

				if info.Size() > 1024*1024 || isBinaryExt(filepath.Ext(path)) {
					return nil
				}
				taskCh <- task{path: path, info: info}
				return nil
			}
		})
	}()

	wg.Wait()

	if len(matches) == 0 {
		return &ToolResult{
			Content: fmt.Sprintf("No matches found for pattern '%s' in %s", pattern, searchPath),
			Display: fmt.Sprintf("[S] No matches for '%s'", truncateStr(pattern, 40)),
		}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d match(es) for '%s':\n\n", len(matches), pattern))
	for _, m := range matches {
		sb.WriteString(fmt.Sprintf("%s:%d: %s\n", m.file, m.lineNum, m.line))
	}
	if len(matches) >= maxMatches {
		sb.WriteString(fmt.Sprintf("\n... (results truncated at %d matches)", maxMatches))
	}

	return &ToolResult{
		Content: sb.String(),
		Display: fmt.Sprintf("[S] Found %d matches for '%s'", len(matches), truncateStr(pattern, 30)),
	}, nil
}

func (t *GrepTool) resolvePath(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(t.workDir, p)
}

func isBinaryExt(ext string) bool {
	binExts := map[string]bool{
		".exe": true, ".bin": true, ".dll": true, ".so": true, ".dylib": true,
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".ico": true,
		".mp3": true, ".mp4": true, ".avi": true, ".mov": true,
		".zip": true, ".tar": true, ".gz": true, ".rar": true,
		".pdf": true, ".doc": true, ".docx": true,
		".woff": true, ".woff2": true, ".ttf": true, ".eot": true,
		".o": true, ".a": true, ".pyc": true, ".class": true,
	}
	return binExts[strings.ToLower(ext)]
}
