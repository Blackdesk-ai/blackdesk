package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) handleMouseMsg(msg tea.MouseMsg) (Model, tea.Cmd) {
	_ = msg
	return m, nil
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd) {
	if next, cmd, handled := m.handleCommandPaletteKey(msg); handled {
		return next, cmd
	}
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
	if next, cmd, handled := m.handleAIComposerEntryKey(msg); handled {
		return next, cmd
	}
	return m, nil
}
