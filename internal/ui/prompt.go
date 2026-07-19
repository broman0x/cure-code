package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type PromptResult struct {
	Text     string
	Canceled bool
}

type promptModel struct {
	textarea textarea.Model
	err      error
	result   PromptResult
}

func initialPromptModel() promptModel {
	ti := textarea.New()
	ti.Placeholder = "Type your prompt (Alt+Enter for newline, @ to tag files)..."
	ti.Focus()
	ti.Prompt = "  cure > "
	ti.CharLimit = 0
	ti.SetWidth(100)
	ti.SetHeight(1) 
	ti.ShowLineNumbers = false
	ti.KeyMap.InsertNewline.SetEnabled(false) 

	return promptModel{
		textarea: ti,
		err:      nil,
	}
}

func (m promptModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m promptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.result.Canceled = true
			return m, tea.Quit
		case tea.KeyEnter:
			if msg.Alt {
				m.textarea.InsertString("\n")
				return m, nil
			}
			val := strings.TrimSpace(m.textarea.Value())
			if val == "" {
				return m, nil
			}
			m.result.Text = val
			return m, tea.Quit
		}

	case error:
		m.err = msg
		return m, nil
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m promptModel) View() string {
	if m.result.Canceled || m.result.Text != "" {
		return ""
	}
	return m.textarea.View()
}

func RunPrompt() (PromptResult, error) {
	p := tea.NewProgram(initialPromptModel())
	m, err := p.Run()
	if err != nil {
		return PromptResult{}, err
	}
	finalModel := m.(promptModel)
	return finalModel.result, nil
}
