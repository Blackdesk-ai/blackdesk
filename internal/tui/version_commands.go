package tui

import (
	"os"
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

func (m Model) upgradeCmd() tea.Cmd {
	currentVersion := m.appVersion
	return func() tea.Msg {
		executablePath, err := os.Executable()
		if err != nil {
			return versionUpgradeLoadedMsg{err: err}
		}
		result, err := updater.Default().Upgrade(m.ctx, executablePath, currentVersion, "")
		return versionUpgradeLoadedMsg{result: result, err: err}
	}
}

func (m Model) ShouldRestartOnQuit() bool {
	return m.restartOnQuit
}
