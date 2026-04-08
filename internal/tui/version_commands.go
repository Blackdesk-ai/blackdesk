package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"blackdesk/internal/updater"
)

func (m Model) shouldCheckForUpdates() bool {
	return strings.TrimSpace(m.appVersion) != ""
}

func (m Model) checkForUpdatesCmd() tea.Cmd {
	currentVersion := m.appVersion
	return func() tea.Msg {
		result, err := updater.Default().Check(m.ctx, currentVersion)
		return versionCheckLoadedMsg{result: result, err: err}
	}
}
