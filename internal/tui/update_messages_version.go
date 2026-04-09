package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) handleVersionCheckLoaded(msg versionCheckLoadedMsg) (Model, tea.Cmd) {
	if msg.err != nil {
		return m, nil
	}
	m.latestVersion = msg.result.LatestVersion
	m.updateAvailable = msg.result.UpdateAvailable
	return m, nil
}

func (m Model) handleVersionUpgradeLoaded(msg versionUpgradeLoadedMsg) (Model, tea.Cmd) {
	m.upgradeRunning = false
	if msg.err != nil {
		m.status = "Upgrade failed: " + msg.err.Error()
		return m, nil
	}

	if msg.result.AlreadyCurrent {
		m.updateAvailable = false
		m.latestVersion = msg.result.InstalledVersion
		m.status = fmt.Sprintf("Blackdesk %s is already installed", msg.result.InstalledVersion)
		return m, nil
	}

	if msg.result.RestartRequired {
		m.updateAvailable = false
		m.latestVersion = msg.result.InstalledVersion
		m.status = fmt.Sprintf("Updated to %s. Restart Blackdesk to use it.", msg.result.InstalledVersion)
		return m, nil
	}

	m.restartOnQuit = true
	m.status = fmt.Sprintf("Updated to %s. Restarting…", msg.result.InstalledVersion)
	return m, tea.Quit
}
