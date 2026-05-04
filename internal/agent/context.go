package agent

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type WorkspaceContext struct {
	WorkDir       string
	ProjectName   string
	GitBranch     string
	GitDirtyCount int
	Languages     []string
	HasGit        bool
	FileTree      string
}

func DetectWorkspace(workDir string) *WorkspaceContext {
	ctx := &WorkspaceContext{
		WorkDir:     workDir,
		ProjectName: filepath.Base(workDir),
	}

	if _, err := os.Stat(filepath.Join(workDir, ".git")); err == nil {
		ctx.HasGit = true
		ctx.GitBranch = getGitBranch(workDir)
		ctx.GitDirtyCount = getGitDirtyCount(workDir)
	}

	ctx.Languages = detectLanguages(workDir)
	ctx.FileTree = generateFileTree(workDir)

	return ctx
}

func getGitBranch(dir string) string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func detectLanguages(dir string) []string {
	langFiles := map[string]string{
		"go.mod":           "Go",
		"package.json":     "JavaScript/TypeScript",
		"requirements.txt": "Python",
		"pyproject.toml":   "Python",
		"Cargo.toml":       "Rust",
		"pom.xml":          "Java",
		"build.gradle":     "Java/Kotlin",
		"build.gradle.kts": "Kotlin",
		"Gemfile":          "Ruby",
		"composer.json":    "PHP",
		"CMakeLists.txt":   "C/C++",
		"Makefile":         "C/C++",
	}

	seen := make(map[string]bool)
	var langs []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if lang, ok := langFiles[name]; ok {
			if !seen[lang] {
				seen[lang] = true
				langs = append(langs, lang)
			}
		}

		if strings.HasSuffix(name, ".csproj") && !seen["C#"] {
			seen["C#"] = true
			langs = append(langs, "C#")
		}

		if name == "Package.swift" && !seen["Swift"] {
			seen["Swift"] = true
			langs = append(langs, "Swift")
		}
	}

	return langs
}

func getGitDirtyCount(dir string) int {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return 0
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return 0
	}
	return len(lines)
}
func generateFileTree(dir string) string {
	// [EN] Generate a lightweight 2-level deep file tree
	// [ID] Buat pohon file ringan sedalam 2 level
	var sb strings.Builder
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	count := 0
	for _, entry := range entries {
		name := entry.Name()
		if name[0] == '.' && name != ".env" {
			continue
		}
		if name == "vendor" || name == "node_modules" {
			continue
		}

		if entry.IsDir() {
			sb.WriteString(fmt.Sprintf("  %s/\n", name))
			subEntries, _ := os.ReadDir(filepath.Join(dir, name))
			subCount := 0
			for _, sub := range subEntries {
				if subCount > 10 {
					sb.WriteString("    ...\n")
					break
				}
				subName := sub.Name()
				if subName[0] == '.' {
					continue
				}
				if sub.IsDir() {
					sb.WriteString(fmt.Sprintf("    %s/\n", subName))
				} else {
					sb.WriteString(fmt.Sprintf("    %s\n", subName))
				}
				subCount++
			}
		} else {
			sb.WriteString(fmt.Sprintf("  %s\n", name))
		}
		count++
		if count > 20 {
			sb.WriteString("  ...\n")
			break
		}
	}
	return sb.String()
}
