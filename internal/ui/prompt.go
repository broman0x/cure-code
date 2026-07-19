package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type PromptResult struct {
	Text     string
	Canceled bool
}

type Suggestion struct {
	Text        string
	Description string
}

type promptModel struct {
	textarea      textarea.Model
	err           error
	result        PromptResult
	lastKeystroke time.Time
	suggestions   []Suggestion
	suggestionIdx int
	completer     func(string) []Suggestion
}

func initialPromptModel(completer func(string) []Suggestion) promptModel {
	ti := textarea.New()
	ti.Placeholder = "Type your prompt (Alt+Enter for newline, @ to tag files)..."
	ti.Focus()
	ti.Prompt = "  cure > "
	ti.CharLimit = 0
	ti.SetWidth(100)
	ti.SetHeight(2) // At least 2 lines to show multiline capability
	ti.ShowLineNumbers = false
	// We MUST enable InsertNewline so that pastes containing \n are not cropped by textarea!
	ti.KeyMap.InsertNewline.SetEnabled(true)

	return promptModel{
		textarea:  ti,
		err:       nil,
		completer: completer,
	}
}

func (m promptModel) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, tea.EnableBracketedPaste)
}

func (m promptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		now := time.Now()
		isBurst := false
		if !m.lastKeystroke.IsZero() && now.Sub(m.lastKeystroke) < 25*time.Millisecond {
			isBurst = true
		}
		m.lastKeystroke = now

		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.result.Canceled = true
			return m, tea.Quit
		case tea.KeyUp:
			if len(m.suggestions) > 0 {
				m.suggestionIdx--
				if m.suggestionIdx < 0 {
					m.suggestionIdx = len(m.suggestions) - 1
				}
				return m, nil
			}
		case tea.KeyDown:
			if len(m.suggestions) > 0 {
				m.suggestionIdx++
				if m.suggestionIdx >= len(m.suggestions) {
					m.suggestionIdx = 0
				}
				return m, nil
			}
		case tea.KeyTab, tea.KeyEnter:
			if len(m.suggestions) > 0 {
				val := m.textarea.Value()
				idx := strings.LastIndexAny(val, " \n\t")
				word := val
				if idx != -1 {
					word = val[idx+1:]
				}

				// Send backspaces to delete the current word
				for i := 0; i < len(word); i++ {
					m.textarea, _ = m.textarea.Update(tea.KeyMsg{Type: tea.KeyBackspace})
				}
				
				// Insert the selected suggestion
				m.textarea.InsertString(m.suggestions[m.suggestionIdx].Text)
				m.suggestions = nil
				return m, nil
			}
			if msg.Type == tea.KeyEnter {
				if msg.Alt || isBurst {
					m.textarea, cmd = m.textarea.Update(msg)
					return m, cmd
				}
				val := strings.TrimSpace(m.textarea.Value())
				if val == "" {
					return m, nil
				}
				m.result.Text = val
				return m, tea.Quit
			}
		}

	case error:
		m.err = msg
		return m, nil
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)

	if m.completer != nil {
		// Just use a simple regex or string split on the current line to get the last word
		val := m.textarea.Value()
		idx := strings.LastIndexAny(val, " \n\t")
		word := val
		if idx != -1 {
			word = val[idx+1:]
		}

		if word != "" {
			m.suggestions = m.completer(word)
			if m.suggestionIdx >= len(m.suggestions) {
				m.suggestionIdx = 0
			}
		} else {
			m.suggestions = nil
		}
	}

	return m, tea.Batch(cmds...)
}

func (m promptModel) View() string {
	if m.result.Canceled {
		return ""
	}
	if m.result.Text != "" {
		val := m.result.Text
		if len(val) > 250 || strings.Contains(val, "\n") {
			preview := val
			if len(preview) > 50 {
				preview = preview[:47] + "..."
			}
			preview = strings.ReplaceAll(preview, "\n", " ")
			// Dim text for the char count
			return fmt.Sprintf("  \033[36mcure >\033[0m %s \033[90m[#%d chars]\033[0m\n", preview, len(val))
		}
		return fmt.Sprintf("  \033[36mcure >\033[0m %s\n", val)
	}
	return m.textarea.View()
}

func RunPrompt(completer func(string) []Suggestion) (PromptResult, error) {
	p := tea.NewProgram(initialPromptModel(completer))
	m, err := p.Run()
	if err != nil {
		return PromptResult{}, err
	}
	finalModel := m.(promptModel)
	return finalModel.result, nil
}
