package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// [EN] IntelligenceService provides advanced code intelligence features.
// [ID] IntelligenceService menyediakan fitur kecerdasan kode tingkat lanjut.
type IntelligenceService struct {
	WorkDir string
}

func NewIntelligenceService(workDir string) *IntelligenceService {
	return &IntelligenceService{WorkDir: workDir}
}

// [EN] SuggestContext suggests relevant files based on the current user query and history.
// [ID] SuggestContext menyarankan file yang relevan berdasarkan query pengguna dan riwayat saat ini.
func (s *IntelligenceService) SuggestContext(query string, history []Message) []string {
	suggestions := make([]string, 0)
	queryLower := strings.ToLower(query)
	
	// [EN] Keywords to ignore
	// [ID] Kata kunci untuk diabaikan
	ignore := map[string]bool{"the": true, "and": true, "for": true, "how": true, "what": true}

	// [EN] Scan workspace for files matching query keywords
	// [ID] Pindai workspace untuk file yang cocok dengan kata kunci query
	words := strings.Fields(queryLower)
	filepath.Walk(s.WorkDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		rel, _ := filepath.Rel(s.WorkDir, path)
		if strings.Contains(rel, ".git") || strings.Contains(rel, "node_modules") {
			return nil
		}

		relLower := strings.ToLower(rel)
		for _, word := range words {
			if len(word) < 3 || ignore[word] {
				continue
			}
			if strings.Contains(relLower, word) {
				suggestions = append(suggestions, rel)
				break
			}
		}

		if len(suggestions) >= 5 {
			return filepath.SkipDir // Stop early if we have enough
		}
		return nil
	})

	return suggestions
}

// [EN] GetWorkspaceOverview returns a concise summary of the codebase.
// [ID] GetWorkspaceOverview mengembalikan ringkasan ringkas dari codebase.
func (s *IntelligenceService) GetWorkspaceOverview() string {
	// [EN] Quick scan of project to find main entry points and packages
	// [ID] Pemindaian cepat proyek untuk mencari entry point dan paket utama
	var sb strings.Builder
	sb.WriteString("Project Overview:\n")
	
	entries, _ := os.ReadDir(s.WorkDir)
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			sb.WriteString(fmt.Sprintf("- Package: %s\n", entry.Name()))
		}
	}
	
	return sb.String()
}
