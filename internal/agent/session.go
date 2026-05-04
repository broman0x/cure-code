package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type SessionData struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	History   []Message `json:"history"`
	Tasks     []Task    `json:"tasks,omitempty"`
	WorkDir   string    `json:"work_dir"`
}

func SaveSession(history []Message, tasks []Task, workDir, configDir string) (string, error) {
	sessionDir := filepath.Join(configDir, "sessions")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create session directory: %v", err)
	}

	sessionID := fmt.Sprintf("session-%d", time.Now().Unix())
	sessionPath := filepath.Join(sessionDir, sessionID+".json")

	data := SessionData{
		ID:        sessionID,
		Timestamp: time.Now(),
		History:   history,
		Tasks:     tasks,
		WorkDir:   workDir,
	}

	file, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal session: %v", err)
	}

	if err := os.WriteFile(sessionPath, file, 0644); err != nil {
		return "", fmt.Errorf("failed to write session file: %v", err)
	}

	return sessionID, nil
}

func LoadSession(sessionID, configDir string) ([]Message, []Task, error) {
	sessionPath := filepath.Join(configDir, "sessions", sessionID+".json")
	if !strings.HasSuffix(sessionID, ".json") {
		sessionPath = filepath.Join(configDir, "sessions", sessionID+".json")
	}

	file, err := os.ReadFile(sessionPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read session file: %v", err)
	}

	var data SessionData
	if err := json.Unmarshal(file, &data); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal session: %v", err)
	}

	return data.History, data.Tasks, nil
}

func ListSessions(configDir string) ([]string, error) {
	sessionDir := filepath.Join(configDir, "sessions")
	entries, err := os.ReadDir(sessionDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var sessions []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			sessions = append(sessions, strings.TrimSuffix(entry.Name(), ".json"))
		}
	}
	return sessions, nil
}
