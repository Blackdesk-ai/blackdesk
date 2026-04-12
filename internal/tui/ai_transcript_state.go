package tui

import (
	"fmt"
	"strings"
	"time"
)

func (m *Model) pushAIUserMessage(prompt string) {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return
	}
	m.aiMessages = append(m.aiMessages, aiMessage{
		Role:      aiMessageUser,
		Body:      prompt,
		Timestamp: time.Now(),
		Meta:      fmt.Sprintf("%s%s", m.activeAIConnectorLabel(), m.activeAIModelStatus()),
	})
	m.maintainAITranscriptBudget()
	m.aiScroll = 0
}

func (m *Model) pushAIAssistantMessage(body string, runErr error, duration time.Duration) {
	meta := m.activeAIConnectorLabel()
	if duration > 0 {
		meta += " • " + duration.Round(time.Millisecond).String()
	}
	if runErr != nil {
		if strings.TrimSpace(body) == "" {
			body = runErr.Error()
		}
		meta += " • error"
	}
	if strings.TrimSpace(body) == "" {
		return
	}
	m.aiMessages = append(m.aiMessages, aiMessage{
		Role:      aiMessageAssistant,
		Body:      body,
		Timestamp: time.Now(),
		Meta:      meta,
	})
	m.maintainAITranscriptBudget()
	m.aiScroll = 0
}

func (m *Model) scrollAITranscript(step int) {
	m.aiScroll = max(0, m.aiScroll+step)
}

func (m *Model) touchAIContext() {
	m.aiContextRevision++
}

func (m Model) aiContextStatusLine() string {
	if m.aiRunning {
		return "refreshing"
	}
	if m.aiLastContext == "" {
		return "cold"
	}
	if m.aiLastContextRevision == m.aiContextRevision && strings.EqualFold(m.aiLastSymbol, m.activeSymbol()) {
		return "stable"
	}
	return "stale"
}
