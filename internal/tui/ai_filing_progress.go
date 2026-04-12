package tui

import (
	"fmt"
	"time"
)

func (m *Model) clearAIFilingRun() {
	m.aiFilingRun = aiFilingRunState{}
	m.aiFilingRunActive = false
}

func (m Model) aiFilingProgressLabel() string {
	if !m.aiFilingRunActive {
		return ""
	}
	total := len(m.aiFilingRun.chunks)
	if total == 0 {
		return ""
	}
	if m.aiFilingRun.synthesizing {
		return fmt.Sprintf("synthesizing from %d chunks", total)
	}
	current := min(total, max(1, m.aiFilingRun.nextChunkIdx+1))
	return fmt.Sprintf("chunk %d/%d", current, total)
}

func (m Model) aiRunningIndicator() string {
	if label := m.aiFilingProgressLabel(); label != "" {
		return label + aiProgressDots(m.clock)
	}
	return aiTypingFrame(m.clock)
}

func aiProgressDots(ts time.Time) string {
	frames := []string{"", ".", "..", "..."}
	if ts.IsZero() {
		return frames[0]
	}
	return frames[ts.Second()%len(frames)]
}
