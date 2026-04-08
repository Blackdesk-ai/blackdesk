package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) loadAIModelsCmd(connectorID string) tea.Cmd {
	if strings.TrimSpace(connectorID) == "" {
		return func() tea.Msg {
			return aiModelsLoadedMsg{connectorID: "", err: fmt.Errorf("no AI connector selected")}
		}
	}
	if m.services == nil {
		return func() tea.Msg {
			return aiModelsLoadedMsg{connectorID: connectorID, err: fmt.Errorf("AI registry unavailable")}
		}
	}
	return func() tea.Msg {
		models, err := m.services.AIModels(m.ctx, connectorID)
		return aiModelsLoadedMsg{connectorID: connectorID, models: models, err: err}
	}
}
