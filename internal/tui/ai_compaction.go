package tui

import "strings"

func (m *Model) maintainAITranscriptBudget() {
	if len(m.aiMessages) == 0 {
		return
	}
	start := compactedRecentMessageStart(m.aiMessages)
	if start <= 0 {
		return
	}
	older := append([]aiMessage(nil), m.aiMessages[:start]...)
	summary := summarizeAIMessages(older)
	if summary != "" {
		m.aiConversationSummary = mergeAIConversationSummary(m.aiConversationSummary, summary, aiMaxSummaryChars)
		m.aiCompactedMessages += len(older)
	}
	m.aiMessages = append([]aiMessage(nil), m.aiMessages[start:]...)
}

func compactedRecentMessageStart(messages []aiMessage) int {
	if len(messages) == 0 {
		return 0
	}
	total := 0
	kept := 0
	start := len(messages)
	for i := len(messages) - 1; i >= 0; i-- {
		msgChars := aiCompactionMessageChars(messages[i])
		if kept >= aiRecentHistoryMinMsgs && total+msgChars > aiMaxRecentHistoryChars {
			break
		}
		total += msgChars
		kept++
		start = i
	}
	return start
}

func aiCompactionMessageChars(msg aiMessage) int {
	body := strings.TrimSpace(msg.Body)
	if body == "" {
		return 0
	}
	return len([]rune(body)) + 24
}

func summarizeAIMessages(messages []aiMessage) string {
	if len(messages) == 0 {
		return ""
	}
	lines := make([]string, 0, len(messages))
	for _, msg := range messages {
		body := collapseAIWhitespace(msg.Body)
		if body == "" {
			continue
		}
		limit := aiSummaryUserChars
		label := "User"
		if msg.Role == aiMessageAssistant {
			label = "AI"
			limit = aiSummaryAssistantChars
		}
		lines = append(lines, "- "+label+": "+truncateRunes(body, limit))
	}
	return strings.Join(lines, "\n")
}

func mergeAIConversationSummary(existing, incoming string, limit int) string {
	lines := make([]string, 0, 64)
	lines = append(lines, nonEmptyAILines(existing)...)
	lines = append(lines, nonEmptyAILines(incoming)...)
	if len(lines) == 0 || limit <= 0 {
		return ""
	}
	kept := make([]string, 0, len(lines))
	chars := 0
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		add := len([]rune(line))
		if len(kept) > 0 {
			add++
		}
		if chars+add > limit {
			break
		}
		kept = append([]string{line}, kept...)
		chars += add
	}
	if len(kept) == 0 {
		return ""
	}
	out := strings.Join(kept, "\n")
	if len(kept) == len(lines) {
		return out
	}
	prefix := "… older conversation summarized …"
	if len([]rune(prefix))+1+len([]rune(out)) <= limit {
		return prefix + "\n" + out
	}
	return out
}

func nonEmptyAILines(input string) []string {
	raw := strings.Split(strings.TrimSpace(input), "\n")
	lines := make([]string, 0, len(raw))
	for _, line := range raw {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

func collapseAIWhitespace(input string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(input)), " ")
}
