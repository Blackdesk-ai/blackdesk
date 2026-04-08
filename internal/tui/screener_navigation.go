package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"blackdesk/internal/application"
)

func (m *Model) cycleScreener(step int) {
	plan := application.PlanScreenerAdvance(application.ScreenerAdvanceInput{
		Available:    m.screenerAvailable(),
		Definitions:  m.screenerDefs,
		CurrentIndex: m.screenerIdx,
		Step:         step,
	})
	m.screenerIdx = plan.NextIndex
}

func (m *Model) cycleScreenerSelection(step int) {
	if len(m.screenerResult.Items) == 0 {
		m.screenerSel = 0
		m.screenerScroll = 0
		return
	}
	m.screenerSel = (m.screenerSel + step + len(m.screenerResult.Items)) % len(m.screenerResult.Items)
	m.ensureScreenerSelectionVisible()
}

func (m *Model) ensureScreenerSelectionVisible() {
	if len(m.screenerResult.Items) == 0 {
		m.screenerSel = 0
		m.screenerScroll = 0
		return
	}
	if m.screenerSel < 0 {
		m.screenerSel = 0
	}
	if m.screenerSel >= len(m.screenerResult.Items) {
		m.screenerSel = len(m.screenerResult.Items) - 1
	}
	visibleRows := m.screenerVisibleRows()
	maxStart := max(0, len(m.screenerResult.Items)-visibleRows)
	if m.screenerScroll > maxStart {
		m.screenerScroll = maxStart
	}
	if m.screenerSel < m.screenerScroll {
		m.screenerScroll = m.screenerSel
	}
	if m.screenerSel >= m.screenerScroll+visibleRows {
		m.screenerScroll = m.screenerSel - visibleRows + 1
	}
}

func (m *Model) cycleProfileScroll() {
	if strings.TrimSpace(m.fundamentals.Description) == "" {
		m.profileScroll = 0
		return
	}

	textStyle := lipgloss.NewStyle().Width(max(18, m.bottomPanelProfileWidth())).MaxWidth(max(18, m.bottomPanelProfileWidth()))
	profileText := textStyle.Render(m.fundamentals.Description)
	bodyHeight := max(1, m.bottomPanelHeight()-2)
	maxOffset := max(0, len(splitLines(profileText))-bodyHeight)
	if maxOffset == 0 || m.profileScroll >= maxOffset {
		m.profileScroll = 0
		return
	}
	m.profileScroll = min(maxOffset, m.profileScroll+max(1, m.bottomPanelHeight()/3))
}
