package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	scrolled      bool
	completer     func(string) []Suggestion
	pendingPastes map[string]string
}

func initialPromptModel(completer func(string) []Suggestion) promptModel {
	ti := textarea.New()
	ti.Placeholder = "Type your prompt (Alt+Enter for newline, @ to tag files)..."
	ti.Focus()
	ti.Prompt = "  cure > "
	ti.CharLimit = 0
	ti.SetWidth(100)
	ti.SetHeight(1)
	ti.MaxHeight = 10
	ti.ShowLineNumbers = false
	ti.KeyMap.InsertNewline.SetEnabled(true)

	return promptModel{
		textarea:      ti,
		err:           nil,
		completer:     completer,
		pendingPastes: make(map[string]string),
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
		if msg.Paste {
			pasted := string(msg.Runes)
			if len(pasted) > 80 || strings.Contains(pasted, "\n") {
				placeholder := fmt.Sprintf("[Pasted Content %d chars]", len(pasted))
				m.pendingPastes[placeholder] = pasted
				m.textarea, cmd = m.textarea.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(placeholder), Paste: true})
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			}
			m.textarea, cmd = m.textarea.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}

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
				m.scrolled = true
				m.suggestionIdx--
				if m.suggestionIdx < 0 {
					m.suggestionIdx = len(m.suggestions) - 1
				}
				return m, nil
			}
		case tea.KeyDown:
			if len(m.suggestions) > 0 {
				m.scrolled = true
				m.suggestionIdx++
				if m.suggestionIdx >= len(m.suggestions) {
					m.suggestionIdx = 0
				}
				return m, nil
			}
		case tea.KeyTab, tea.KeyEnter:
			isEnter := msg.Type == tea.KeyEnter
			if isEnter && (msg.Alt || isBurst) {
				m.textarea, cmd = m.textarea.Update(msg)
				return m, cmd
			}

			if len(m.suggestions) > 0 && (!isEnter || (isEnter && m.scrolled)) {
				val := m.textarea.Value()
				idx := strings.LastIndexAny(val, " \n\t")
				word := val
				if idx != -1 {
					word = val[idx+1:]
				}

				for i := 0; i < len(word); i++ {
					m.textarea, _ = m.textarea.Update(tea.KeyMsg{Type: tea.KeyBackspace})
				}
				
				m.textarea.InsertString(m.suggestions[m.suggestionIdx].Text + " ")
				m.suggestions = nil
				m.scrolled = false
				return m, nil
			}

			if isEnter {
				val := strings.TrimSpace(m.textarea.Value())
				
				// Rehydrate large pastes
				for placeholder, original := range m.pendingPastes {
					val = strings.ReplaceAll(val, placeholder, original)
				}

				if val == "" {
					return m, nil
				}
				m.result.Text = val
				return m, tea.Quit
			}
		case tea.KeyRunes, tea.KeySpace, tea.KeyBackspace:
			m.scrolled = false
		}

	case error:
		m.err = msg
		return m, nil
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)

	if m.completer != nil {
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
		if len(val) > 80 || strings.Contains(val, "\n") {
			preview := val
			if len(preview) > 50 {
				preview = preview[:47] + "..."
			}
			preview = strings.ReplaceAll(preview, "\n", " ")
			return fmt.Sprintf("  \033[36mcure >\033[0m %s \033[90m[#%d chars]\033[0m\n", preview, len(val))
		}
		return fmt.Sprintf("  \033[36mcure >\033[0m %s\n", val)
	}
	view := m.textarea.View()
	if len(m.suggestions) > 0 {
		var sb strings.Builder
		sb.WriteString("\n")

		maxTextLen := 0
		maxDescLen := 0
		for _, s := range m.suggestions {
			if len(s.Text) > maxTextLen {
				maxTextLen = len(s.Text)
			}
			if len(s.Description) > maxDescLen {
				maxDescLen = len(s.Description)
			}
		}

		suggestionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Background(lipgloss.Color("236")).Padding(0, 1).Width(maxTextLen + 2)
		suggestionDescStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Background(lipgloss.Color("234")).Padding(0, 1).Width(maxDescLen + 2)
		selectedSuggestionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("232")).Background(lipgloss.Color("43")).Padding(0, 1).Width(maxTextLen + 2)
		selectedDescStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Background(lipgloss.Color("236")).Padding(0, 1).Width(maxDescLen + 2)

		displayCount := 8
		startIdx := 0
		if m.suggestionIdx >= displayCount {
			startIdx = m.suggestionIdx - displayCount + 1
		}
		endIdx := startIdx + displayCount
		if endIdx > len(m.suggestions) {
			endIdx = len(m.suggestions)
		}

		for i := startIdx; i < endIdx; i++ {
			s := m.suggestions[i]
			var tStyle, dStyle lipgloss.Style
			if i == m.suggestionIdx {
				tStyle = selectedSuggestionStyle
				dStyle = selectedDescStyle
			} else {
				tStyle = suggestionStyle
				dStyle = suggestionDescStyle
			}

			if maxDescLen == 0 {
				sb.WriteString("  " + tStyle.Render(s.Text) + "\n")
			} else {
				row := lipgloss.JoinHorizontal(lipgloss.Left, tStyle.Render(s.Text), dStyle.Render(s.Description))
				sb.WriteString("  " + row + "\n")
			}
		}
		view += sb.String()
	}
	return view
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
