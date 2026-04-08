package tui

import "strings"

func (m Model) activeAIConnectorID() string {
	return strings.TrimSpace(m.config.AIConnector)
}

func (m Model) activeAIConnectorLabel() string {
	if strings.TrimSpace(m.activeAIConnectorID()) == "" {
		return "Not selected"
	}
	if m.services != nil {
		if item, ok := m.services.LookupAIConnector(m.activeAIConnectorID()); ok {
			return item.Label
		}
	}
	return strings.ToUpper(m.activeAIConnectorID())
}

func (m Model) currentAIModels() []string {
	if strings.TrimSpace(m.activeAIConnectorID()) == "" {
		return nil
	}
	return append([]string(nil), m.aiModels[strings.ToLower(m.activeAIConnectorID())]...)
}

func (m *Model) syncAIModelSelection() {
	if strings.TrimSpace(m.activeAIConnectorID()) == "" {
		m.aiModelIdx = 0
		m.config.AIModel = ""
		return
	}
	models := m.currentAIModels()
	if len(models) == 0 {
		m.aiModelIdx = 0
		m.config.AIModel = ""
		return
	}
	current := strings.TrimSpace(m.config.AIModel)
	if current == "" {
		m.aiModelIdx = 0
		m.config.AIModel = models[0]
		return
	}
	for i, model := range models {
		if strings.EqualFold(model, current) {
			m.aiModelIdx = i
			m.config.AIModel = models[i]
			return
		}
	}
	m.aiModelIdx = 0
	m.config.AIModel = models[0]
}

func (m *Model) cycleAIModel(step int) {
	models := m.currentAIModels()
	if len(models) == 0 {
		return
	}
	m.syncAIModelSelection()
	m.aiModelIdx = (m.aiModelIdx + step + len(models)) % len(models)
	m.config.AIModel = models[m.aiModelIdx]
	m.status = "AI model: " + m.config.AIModel
}

func (m *Model) cycleAIConnector(step int) {
	items := m.services.ListAIConnectors()
	if len(items) == 0 {
		return
	}
	idx := -1
	current := m.activeAIConnectorID()
	for i, item := range items {
		if strings.EqualFold(item.ID, current) {
			idx = i
			break
		}
	}
	if idx < 0 {
		if step < 0 {
			idx = len(items) - 1
		} else {
			idx = 0
		}
	} else {
		idx = (idx + step + len(items)) % len(items)
	}
	m.config.AIConnector = items[idx].ID
	m.config.AIModel = ""
	m.status = "AI connector: " + items[idx].Label
	if m.aiOutput == "" {
		m.aiErr = nil
	}
}

func (m Model) activeAIModelStatus() string {
	model := strings.TrimSpace(m.config.AIModel)
	if model == "" {
		return ""
	}
	return " • " + model
}

func (m Model) activeAIStatusLabel() string {
	model := strings.TrimSpace(m.config.AIModel)
	if model == "" {
		return "No model selected"
	}
	return model
}
