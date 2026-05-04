package agent

import (
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
