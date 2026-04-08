package tui

import "github.com/charmbracelet/lipgloss"

func (m Model) renderCenterPanel(header, section, label, muted, pos, neg lipgloss.Style, chartWidth, height int) string {
	switch m.tabIdx {
	case tabMarkets:
		return m.renderMarketCenter(header, section, label, muted, pos, neg, chartWidth, height)
	case tabScreener:
		return m.renderScreenerCenter(section, label, muted, pos, neg, chartWidth, height)
	case tabQuote:
		return m.renderOverviewCenter(header, section, label, muted, pos, neg, chartWidth, height)
	case tabNews:
		return m.renderNewsCenter(section, label, muted, chartWidth, height)
	case tabAI:
		if m.aiPickerOpen {
			return m.renderAIPickerCenter(section, label, muted, chartWidth, height)
		}
		return m.renderAICenter(section, muted, chartWidth, height)
	default:
		return m.renderMarketCenter(header, section, label, muted, pos, neg, chartWidth, height)
	}
}

func (m Model) renderRightPanel(section, label, muted lipgloss.Style, width, height int) string {
	switch m.tabIdx {
	case tabMarkets:
		return m.renderMarketRight(section, label, muted, width, height)
	case tabScreener:
		return m.renderScreenerRight(section, label, muted, width, height)
	case tabQuote:
		return m.renderOverviewRight(section, label, muted, width, height)
	case tabNews:
		return m.renderNewsRight(section, label, muted, width, height)
	case tabAI:
		return m.renderAIRight(section, label, muted, width, height)
	default:
		return m.renderMarketRight(section, label, muted, width, height)
	}
}

func (m Model) renderBottomPanel(section, label, muted lipgloss.Style, width, height int) string {
	if m.searchMode {
		return m.renderSearchPanel(section, muted, height)
	}
	if m.tabIdx == tabAI && m.aiPickerOpen {
		return m.renderAIPicker(section, muted, width, height)
	}
	switch m.tabIdx {
	case tabMarkets:
		return m.renderMarketBottom(section, label, muted, width, height)
	case tabQuote:
		return m.renderOverviewBottom(section, muted, width, height)
	default:
		return m.renderMarketBottom(section, label, muted, width, height)
	}
}
