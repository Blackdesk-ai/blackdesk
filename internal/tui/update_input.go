package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) handleMouseMsg(msg tea.MouseMsg) (Model, tea.Cmd) {
	if m.tabIdx != tabAI || m.aiPickerOpen || m.searchMode {
		return m, nil
	}
	if msg.Action != tea.MouseActionPress {
		return m, nil
	}
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		m.scrollAITranscript(3)
	case tea.MouseButtonWheelDown:
		m.scrollAITranscript(-3)
	}
	return m, nil
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd) {
	if next, cmd, handled := m.handleAIPickerKey(msg); handled {
		return next, cmd
	}
	if next, cmd, handled := m.handleAIFocusedKey(msg); handled {
		return next, cmd
	}
	if next, cmd, handled := m.handleSearchKey(msg); handled {
		return next, cmd
	}
	if next, cmd, handled := m.handleHelpKey(msg); handled {
		return next, cmd
	}
	return m.handleGlobalKey(msg)
}

func (m Model) handleGlobalKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	key := msg.String()
	if next, cmd, handled := m.handleGlobalTopLevelKey(key); handled {
		return next, cmd
	}
	if next, cmd, handled := m.handleGlobalNavigationKey(key); handled {
		return next, cmd
	}
	if next, cmd, handled := m.handleGlobalWorkspaceActionKey(key); handled {
		return next, cmd
	}
	return m, nil
}
