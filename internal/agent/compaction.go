package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// [EN] checkAndCompact checks if the history tokens exceed the threshold and triggers compaction.
// [ID] checkAndCompact mengecek apakah token riwayat melebihi ambang batas dan memicu pemadatan.
func (a *Agent) checkAndCompact(ctx context.Context) error {
	// [EN] Conservative estimate for tool result growth per turn (inspired by Claude Code)
	// [ID] Estimasi konservatif untuk pertumbuhan hasil tool per putaran (terinspirasi oleh Claude Code)
	const toolGrowthEstimate = 10000
	
	currentUsage := a.Usage.TotalInputTokens
	predictiveUsage := currentUsage + toolGrowthEstimate

	// [EN] Use a context-aware threshold. If we don't have usage stats yet, use history length as fallback.
	// [ID] Gunakan ambang batas yang sadar konteks. Jika belum ada statistik penggunaan, gunakan panjang riwayat sebagai cadangan.
	if len(a.History) < 15 {
		return nil
	}

	if predictiveUsage < a.CompactThreshold && currentUsage < a.CompactThreshold {
		return nil
	}

	color.HiBlack("  [Memory] Context pressure detected (%d tokens). Predictive usage: %d", currentUsage, predictiveUsage)
	color.HiBlack("  [Memory] Condensing history to maintain reasoning quality...")

	// [EN] Select messages to compact (everything except last 8 messages to keep immediate flow)
	// [ID] Pilih pesan untuk dipadatkan (semua kecuali 8 pesan terakhir untuk menjaga alur langsung)
	keepCount := 8
	if len(a.History) <= keepCount {
		return nil
	}

	toCompact := a.History[:len(a.History)-keepCount]
	toKeep := a.History[len(a.History)-keepCount:]

	// [EN] Generate high-fidelity summary
	// [ID] Buat ringkasan fidelitas tinggi
	summary, err := a.summarizeHistory(ctx, toCompact)
	if err != nil {
		return err
	}

	// [EN] Rebuild history with a "Context Injection" pattern
	// [ID] Bangun ulang riwayat dengan pola "Context Injection"
	newHistory := []Message{
		{
			Role: "user",
			Content: fmt.Sprintf(`[SYSTEM NOTIFICATION: CONTEXT CONDENSED]
The earlier part of this conversation has been summarized to save space. 
IMPORTANT: Maintain the technical decisions and progress described below.

### Summary of Previous Context:
%s

### End of Summary. 
Please acknowledge and continue with the current task based on the messages below.`, summary),
		},
		{
			Role:    "assistant",
			Content: "Context absorbed. I have integrated the summarized history into my active memory. I am ready to proceed with the next steps.",
		},
	}

	a.History = append(newHistory, toKeep...)
	
	// [EN] Reset usage stats for the new compressed context (approximate)
	// [ID] Reset statistik penggunaan untuk konteks terkompresi yang baru (perkiraan)
	a.Usage.TotalInputTokens = len(summary) / 4 // Rough estimate
	
	color.Green("  [Memory] Success: History compacted. Reasoning headroom restored.")

	return nil
}

func (a *Agent) summarizeHistory(ctx context.Context, messages []Message) (string, error) {
	// [EN] Inject spatial context into the summarization prompt
	// [ID] Suntikkan konteks spasial ke dalam prompt peringkasan
	spatialInfo := ""
	if len(a.RecentSymbols) > 0 {
		spatialInfo = fmt.Sprintf("\nRecently tracked symbols: %s", strings.Join(a.RecentSymbols, ", "))
	}

	compactPrompt := fmt.Sprintf(`You are a Context Manager. Your goal is to condense a conversation history into a "High-Fidelity Memory Block".
This block will be used by another AI (you) to continue a complex coding task.

Strictly capture:
1. Technical Roadmap: What has been achieved and what is the immediate goal?
2. Architecture & Patterns: Any specific design decisions made (e.g. "using Gin for API", "implemented repository pattern").
3. Files & Changes: Which files were modified and why? List key changes.
4. Resolved Roadblocks: What errors were encountered and how were they fixed?
5. State Info: %s

Respond with a structured, concise, yet technically dense summary.`, spatialInfo)

	// [EN] We use the provider to generate the summary
	// [ID] Kita menggunakan provider untuk membuat ringkasan
	resp, err := a.Provider.SendWithTools(ctx, compactPrompt, messages, nil)
	if err != nil {
		return "", fmt.Errorf("summarization request failed: %v", err)
	}

	return resp.Content, nil
}
