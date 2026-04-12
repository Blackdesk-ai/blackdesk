package tui

import "strings"

type aiRequestTruncation struct {
	ContextPayload      bool
	ConversationSummary bool
	ConversationHistory bool
	FilingText          bool
	FinalPrompt         bool
}

func (t aiRequestTruncation) activeLabels() []string {
	labels := make([]string, 0, 5)
	if t.ContextPayload {
		labels = append(labels, "app context")
	}
	if t.ConversationSummary {
		labels = append(labels, "summary")
	}
	if t.ConversationHistory {
		labels = append(labels, "history")
	}
	if t.FilingText {
		labels = append(labels, "filing text")
	}
	if t.FinalPrompt {
		labels = append(labels, "final prompt")
	}
	return labels
}

func (t aiRequestTruncation) hasAny() bool {
	return len(t.activeLabels()) > 0
}

func (t aiRequestTruncation) summaryLine() string {
	if !t.hasAny() {
		return "full"
	}
	return "clipped: " + strings.Join(t.activeLabels(), ", ")
}
