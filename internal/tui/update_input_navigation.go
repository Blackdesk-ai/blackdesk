package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) handleGlobalNavigationKey(key string) (Model, tea.Cmd, bool) {
	switch key {
	case "down":
		return m.handleWorkspaceVerticalNavigation(1)
	case "up":
		return m.handleWorkspaceVerticalNavigation(-1)
	case "right":
		if next, cmd, handled := m.handleStatisticsRangeNavigation(1); handled {
			return next, cmd, true
		}
		if next, cmd, handled := m.handleFilingsFilterNavigation(1); handled {
			return next, cmd, true
		}
		if next, cmd, handled := m.handleScreenerDefinitionKey(1); handled {
			return next, cmd, true
		}
		if next, cmd, handled := m.handleStatementKindNavigation(1); handled {
			return next, cmd, true
		}
		return m.handleTimeframeNavigation(1)
	case "left":
		if next, cmd, handled := m.handleStatisticsRangeNavigation(-1); handled {
			return next, cmd, true
		}
		if next, cmd, handled := m.handleFilingsFilterNavigation(-1); handled {
			return next, cmd, true
		}
		if next, cmd, handled := m.handleScreenerDefinitionKey(-1); handled {
			return next, cmd, true
		}
		if next, cmd, handled := m.handleStatementKindNavigation(-1); handled {
			return next, cmd, true
		}
		return m.handleTimeframeNavigation(-1)
	case "]":
		if next, cmd, handled := m.handleStatementFrequencyNavigation(1); handled {
			return next, cmd, true
		}
		return m, nil, true
	case "[":
		if next, cmd, handled := m.handleStatementFrequencyNavigation(-1); handled {
			return next, cmd, true
		}
		return m, nil, true
	default:
		return m, nil, false
	}
}
