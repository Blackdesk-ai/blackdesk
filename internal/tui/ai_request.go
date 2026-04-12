package tui

import (
	"encoding/json"
	"strings"

	"blackdesk/internal/domain"
)

type RequestEnvelope struct {
	Prompt         string
	SystemPrompt   string
	ContextPayload string
	ActiveSymbol   string
}

func (m Model) buildAIRequest(prompt string) (RequestEnvelope, error) {
	ctxPayload, err := json.MarshalIndent(m.aiContextSnapshot(), "", "  ")
	if err != nil {
		return RequestEnvelope{}, err
	}
	payload := truncateRunes(string(ctxPayload), aiMaxContextChars)
	activeSymbol := m.activeSymbol()

	var b strings.Builder
	b.WriteString(strings.TrimSpace(aiSystemPromptTemplate))
	b.WriteString("\n\n")
	b.WriteString("<blackdesk_context_update>\n")
	b.WriteString(payload)
	b.WriteString("\n</blackdesk_context_update>\n\n")
	if summary := strings.TrimSpace(m.aiConversationSummary); summary != "" {
		b.WriteString("<conversation_summary>\n")
		b.WriteString(truncateRunes(summary, aiMaxSummaryChars))
		b.WriteString("\n</conversation_summary>\n\n")
	}
	b.WriteString("<conversation>\n")
	history := m.aiMessages
	if len(history) > 0 {
		last := history[len(history)-1]
		if last.Role == aiMessageUser && strings.TrimSpace(last.Body) == strings.TrimSpace(prompt) {
			history = history[:len(history)-1]
		}
	}
	historyBlock := make([]string, 0, len(history))
	historyChars := 0
	for i := len(history) - 1; i >= 0; i-- {
		msg := history[i]
		role := "assistant"
		if msg.Role == aiMessageUser {
			role = "user"
		}
		entry := "[" + role + "] " + truncateRunes(strings.TrimSpace(msg.Body), aiMaxMessageChars)
		entry += "\n\n"
		if historyChars+len([]rune(entry)) > aiMaxRecentHistoryChars {
			break
		}
		historyBlock = append([]string{entry}, historyBlock...)
		historyChars += len([]rune(entry))
	}
	for _, entry := range historyBlock {
		b.WriteString(entry)
	}
	b.WriteString("</conversation>")

	return RequestEnvelope{
		Prompt:         prompt,
		SystemPrompt:   truncateRunes(b.String(), aiMaxPromptChars),
		ContextPayload: payload,
		ActiveSymbol:   activeSymbol,
	}, nil
}

func truncateRunes(input string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(input)
	if len(runes) <= limit {
		return input
	}
	if limit <= 3 {
		return string(runes[:limit])
	}
	return string(runes[:limit-3]) + "..."
}

func compactAINews(items []domain.NewsItem, limit int) []domain.NewsItem {
	if len(items) == 0 || limit <= 0 {
		return nil
	}
	if len(items) > limit {
		items = items[:limit]
	}
	out := make([]domain.NewsItem, 0, len(items))
	for _, item := range items {
		item.Title = truncateRunes(strings.TrimSpace(item.Title), aiNewsTitleChars)
		item.Summary = truncateRunes(strings.TrimSpace(item.Summary), aiNewsSummaryChars)
		item.Publisher = truncateRunes(strings.TrimSpace(item.Publisher), aiNewsPublisherChars)
		item.URL = truncateRunes(strings.TrimSpace(item.URL), aiNewsURLChars)
		out = append(out, item)
	}
	return out
}
