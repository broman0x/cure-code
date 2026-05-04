package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Skill struct {
	Name        string
	Description string
	Instruction string
}

type SkillRegistry struct {
	skills map[string]Skill
}

func NewSkillRegistry() *SkillRegistry {
	return &SkillRegistry{skills: make(map[string]Skill)}
}

func (r *SkillRegistry) Register(s Skill) {
	r.skills[s.Name] = s
}

func (r *SkillRegistry) Get(name string) (Skill, bool) {
	s, ok := r.skills[name]
	return s, ok
}

func (r *SkillRegistry) List() []Skill {
	var list []Skill
	for _, s := range r.skills {
		list = append(list, s)
	}
	return list
}

func (r *SkillRegistry) LoadBuiltin() {
	r.Register(Skill{
		Name:        "CodeReview",
		Description: "Comprehensive analysis of code quality, bugs, and security.",
		Instruction: `When performing a Code Review:
1. Start by using 'list_directory' to understand the project structure.
2. Read the main files using 'read_file'.
3. Analyze the logic for potential edge cases, bugs, and security vulnerabilities.
4. Evaluate code style, readability, and performance.
5. DO NOT just list suggestions. Actually use 'edit_file' or 'write_file' to implement the improvements immediately.`,
	})

	r.Register(Skill{
		Name:        "TestAndFix",
		Description: "Automated testing and iterative bug fixing.",
		Instruction: `When asked to Test and Fix:
1. Identify existing tests or create new ones using 'write_file'.
2. Run the tests using 'run_command'.
3. Analyze the test output to pinpoint failures.
4. Modify the code using 'edit_file' to fix the bugs.
5. Re-run tests to verify the fix. Repeat until all tests pass.`,
	})
}

func (r *SkillRegistry) LoadFromDir(dir string) error {
	skillsDir := filepath.Join(dir, ".curecode", "skills")
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		return nil
	}

	subdirs, err := os.ReadDir(skillsDir)
	if err != nil {
		return err
	}

	for _, d := range subdirs {
		if !d.IsDir() {
			continue
		}

		skillFile := filepath.Join(skillsDir, d.Name(), "SKILL.md")
		if _, err := os.Stat(skillFile); err == nil {
			skill, err := parseSkillFile(skillFile)
			if err == nil {
				r.Register(skill)
			}
		}
	}
	return nil
}

func parseSkillFile(path string) (Skill, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Skill{}, err
	}

	lines := strings.Split(string(content), "\n")
	skill := Skill{Name: filepath.Base(filepath.Dir(path))}

	inInstruction := false
	var instructionLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			skill.Name = strings.TrimPrefix(trimmed, "# ")
			continue
		}
		if strings.HasPrefix(trimmed, "Description:") {
			skill.Description = strings.TrimSpace(strings.TrimPrefix(trimmed, "Description:"))
			continue
		}
		if strings.HasPrefix(trimmed, "Instruction:") {
			inInstruction = true
			continue
		}
		if inInstruction {
			instructionLines = append(instructionLines, line)
		}
	}

	skill.Instruction = strings.Join(instructionLines, "\n")
	if skill.Description == "" {
		skill.Description = fmt.Sprintf("Custom skill: %s", skill.Name)
	}

	return skill, nil
}
