package tui

import "blackdesk/internal/domain"

func (m Model) screenerAvailable() bool {
	if len(m.screenerDefs) > 0 {
		return true
	}
	return m.services.HasScreeners()
}

func (m Model) currentScreenerDefinition() domain.ScreenerDefinition {
	if len(m.screenerDefs) == 0 {
		return domain.ScreenerDefinition{}
	}
	idx := min(max(0, m.screenerIdx), len(m.screenerDefs)-1)
	return m.screenerDefs[idx]
}

func (m Model) currentScreenerItem() (domain.ScreenerItem, bool) {
	if len(m.screenerResult.Items) == 0 {
		return domain.ScreenerItem{}, false
	}
	idx := min(max(0, m.screenerSel), len(m.screenerResult.Items)-1)
	return m.screenerResult.Items[idx], true
}
