package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) handleAIWorkspaceActionKey(key string) (Model, tea.Cmd, bool) {
	if m.tabIdx != tabAI {
		return m, nil, false
	}
	switch key {
	case "f":
		m.aiFullscreen = !m.aiFullscreen
		if m.aiFullscreen {
			m.status = "AI fullscreen enabled"
		} else {
			m.status = "AI fullscreen disabled"
		}
		return m, nil, true
	case "r":
		prompt := strings.TrimSpace(m.aiInput.Value())
		if prompt == "" || m.aiRunning {
			return m, nil, true
		}
		m.pushAIUserMessage(prompt)
		m.aiInput.SetValue("")
		m.aiRunning = true
		m.aiErr = nil
		m.status = "Refreshing AI context…"
		return m, m.prepareAIContextCmd(prompt), true
	case "x":
		m.aiOutput = ""
		m.aiErr = nil
		m.aiDuration = 0
		m.aiMessages = nil
		m.aiConversationSummary = ""
		m.aiCompactedMessages = 0
		m.status = "Cleared AI response"
		return m, nil, true
	default:
		return m, nil, false
	}
}
