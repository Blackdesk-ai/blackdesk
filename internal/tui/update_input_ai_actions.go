package tui

import tea "github.com/charmbracelet/bubbletea"

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
		return m, nil, true
	case "x":
		m.aiOutput = ""
		m.aiErr = nil
		m.aiDuration = 0
		m.aiMessages = nil
		m.aiConversationSummary = ""
		m.aiCompactedMessages = 0
		m.aiLastRequestTruncation = aiRequestTruncation{}
		m.status = "Cleared AI response"
		return m, nil, true
	default:
		return m, nil, false
	}
}
