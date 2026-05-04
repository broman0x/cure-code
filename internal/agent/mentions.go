package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

var mentionRegex = regexp.MustCompile(`@([\w\.\-/]+)`)

func (a *Agent) ResolveMentions(input string) string {
	mentions := mentionRegex.FindAllStringSubmatch(input, -1)
	if len(mentions) == 0 {
		return input
	}

	var resolvedContent strings.Builder
	resolvedContent.WriteString(input)
	resolvedContent.WriteString("\n\n---\n")
	resolvedContent.WriteString("Context from mentions:\n")

	resolvedCount := 0
	for _, m := range mentions {
		path := m[1]
		fullPath := filepath.Join(a.WorkDir, path)

		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		if info.IsDir() {
			resolvedContent.WriteString(fmt.Sprintf("\n[Folder: @%s]\n", path))
			entries, err := os.ReadDir(fullPath)
			if err == nil {
				for i, entry := range entries {
					if i >= 50 {
						resolvedContent.WriteString("... (truncated)\n")
						break
					}
					name := entry.Name()
					if entry.IsDir() {
						name += "/"
					}
					resolvedContent.WriteString(fmt.Sprintf("- %s\n", name))
				}
			}
			resolvedCount++
			color.HiBlack("  [Mention] Resolved folder: @%s", path)
		} else {
			resolvedContent.WriteString(fmt.Sprintf("\n[File: @%s]\n", path))
			content, err := os.ReadFile(fullPath)
			if err == nil {
				text := string(content)
				lines := strings.Split(text, "\n")
				if len(lines) > 500 {
					text = strings.Join(lines[:500], "\n") + "\n... (truncated)"
				}
				resolvedContent.WriteString("```\n")
				resolvedContent.WriteString(text)
				resolvedContent.WriteString("\n```\n")
			}
			resolvedCount++
			color.HiBlack("  [Mention] Resolved file: @%s", path)
		}
	}

	if resolvedCount == 0 {
		return input
	}

	return resolvedContent.String()
}
