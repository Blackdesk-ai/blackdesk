package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) handleAIPickerKey(msg tea.KeyMsg) (Model, tea.Cmd, bool) {
	if !m.aiPickerOpen {
		return m, nil, false
	}
	switch msg.String() {
	case ".", "esc":
		m.aiPickerOpen = false
		m.aiModelBusy = false
		m.aiPickerStep = aiPickerStepConnector
		m.status = "Closed AI picker"
		return m, nil, true
	case "up":
		if m.aiPickerStep == aiPickerStepConnector {
			m.cycleAIConnector(-1)
			return m, m.persistCmd(), true
		}
		m.cycleAIModel(-1)
		return m, m.persistCmd(), true
	case "down":
		if m.aiPickerStep == aiPickerStepConnector {
			m.cycleAIConnector(1)
			return m, m.persistCmd(), true
		}
		m.cycleAIModel(1)
		return m, m.persistCmd(), true
	case "left":
		if m.aiPickerStep == aiPickerStepModel {
			m.aiPickerStep = aiPickerStepConnector
			m.aiModelBusy = false
			m.aiModelErr = nil
			m.status = "Select AI connector"
		}
		return m, nil, true
	case "right", "enter":
		if msg.String() == "right" {
			updated, cmd := m.handleAIPickerRight()
			return updated, cmd, true
		}
		updated, cmd := m.handleAIPickerEnter()
		return updated, cmd, true
	default:
		return m, nil, true
	}
}

func (m Model) handleAIPickerRight() (Model, tea.Cmd) {
	if m.aiPickerStep == aiPickerStepConnector {
		m.aiPickerStep = aiPickerStepModel
		m.aiModelErr = nil
		if len(m.currentAIModels()) > 0 {
			m.syncAIModelSelection()
			m.status = fmt.Sprintf("Select model for %s", m.activeAIConnectorLabel())
			return m, nil
		}
		m.aiModelBusy = true
		m.status = "Loading AI models…"
		return m, m.loadAIModelsCmd(m.activeAIConnectorID())
	}
	m.cycleAIModel(1)
	return m, m.persistCmd()
}

func (m Model) handleAIPickerEnter() (Model, tea.Cmd) {
	if m.aiPickerStep == aiPickerStepConnector {
		return m.handleAIPickerRight()
	}
	m.aiPickerOpen = false
	m.aiPickerStep = aiPickerStepConnector
	m.status = fmt.Sprintf("AI model: %s%s", m.activeAIConnectorLabel(), m.activeAIModelStatus())
	return m, m.persistCmd()
}

func (m Model) handleAIFocusedKey(msg tea.KeyMsg) (Model, tea.Cmd, bool) {
	if !m.aiFocused {
		return m, nil, false
	}
	switch msg.String() {
	case "esc":
		m.aiFocused = false
		m.aiInput.Blur()
		return m, nil, true
	case ".", "enter":
		prompt := strings.TrimSpace(m.aiInput.Value())
		if prompt == "" || m.aiRunning {
			return m, nil, true
		}
		m.aiFocused = false
		m.aiInput.Blur()
		m.pushAIUserMessage(prompt)
		m.aiInput.SetValue("")
		m.aiRunning = true
		m.aiErr = nil
		m.status = "Refreshing AI context…"
		return m, m.prepareAIContextCmd(prompt), true
	default:
		var cmd tea.Cmd
		m.aiInput, cmd = m.aiInput.Update(msg)
		return m, cmd, true
	}
}

func (m Model) handleAIComposerEntryKey(msg tea.KeyMsg) (Model, tea.Cmd, bool) {
	if m.tabIdx != tabAI || m.aiPickerOpen || m.aiFocused {
		return m, nil, false
	}
	if msg.String() == "." {
		m.aiFocused = true
		m.aiInput.Focus()
		m.status = "AI composer focused"
		return m, nil, true
	}
	if msg.Type != tea.KeyRunes || len(msg.Runes) == 0 {
		return m, nil, false
	}
	m.aiFocused = true
	m.aiInput.Focus()
	m.status = "AI composer focused"
	var cmd tea.Cmd
	m.aiInput, cmd = m.aiInput.Update(msg)
	return m, cmd, true
}
