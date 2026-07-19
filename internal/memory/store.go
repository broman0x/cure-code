package memory

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Store struct {
	Rules []string `json:"rules"`
}

func GetMemoryPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".curecode", "memory.json")
}

func Load() (*Store, error) {
	path := GetMemoryPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return &Store{Rules: []string{}}, nil
	}
	
	var store Store
	if err := json.Unmarshal(data, &store); err != nil {
		return &Store{Rules: []string{}}, nil
	}
	return &store, nil
}

func Save(store *Store) error {
	path := GetMemoryPath()
	os.MkdirAll(filepath.Dir(path), 0755)
	
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func AddRule(rule string) error {
	store, _ := Load()
	store.Rules = append(store.Rules, rule)
	return Save(store)
}

func ClearRules() error {
	store := &Store{Rules: []string{}}
	return Save(store)
}
