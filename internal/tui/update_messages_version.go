package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) handleVersionCheckLoaded(msg versionCheckLoadedMsg) (Model, tea.Cmd) {
	if msg.err != nil {
		return m, nil
	}
	m.latestVersion = msg.result.LatestVersion
	m.updateAvailable = msg.result.UpdateAvailable
	return m, nil
}
