package tui

import (
	"strings"

	"blackdesk/internal/application"
)

func (m *Model) selectSymbol(symbol string) {
	prev := strings.TrimSpace(m.config.ActiveSymbol)
	m.config = application.SetActiveSymbol(m.config, symbol)
	if !strings.EqualFold(prev, strings.TrimSpace(m.config.ActiveSymbol)) {
		m.touchAIContext()
	}
	m.errTechnicalHistory = nil
	m.errStatement = nil
	m.errInsiders = nil
}

func (m *Model) addToWatchlist(symbol string) {
	prev := len(m.config.Watchlist)
	state := application.AddWatchlistSymbol(m.config, m.selectedIdx, m.watchlistScroll, m.watchlistVisibleRows(), symbol)
	m.config = state.Config
	m.selectedIdx = state.SelectedIndex
	m.watchlistScroll = state.Scroll
	if len(m.config.Watchlist) != prev {
		m.touchAIContext()
	}
}

func (m *Model) removeSelectedWatchlistSymbol() {
	prev := len(m.config.Watchlist)
	state := application.RemoveWatchlistSymbol(m.config, m.selectedIdx, m.watchlistScroll, m.watchlistVisibleRows())
	m.config = state.Config
	m.selectedIdx = state.SelectedIndex
	m.watchlistScroll = state.Scroll
	if len(m.config.Watchlist) != prev {
		m.touchAIContext()
	}
	m.errTechnicalHistory = nil
	m.errStatement = nil
	m.errInsiders = nil
}

func (m *Model) ensureWatchlistSelectionVisible() {
	if len(m.config.Watchlist) == 0 {
		m.watchlistScroll = 0
		return
	}
	if m.selectedIdx < 0 {
		m.selectedIdx = 0
	}
	if m.selectedIdx >= len(m.config.Watchlist) {
		m.selectedIdx = len(m.config.Watchlist) - 1
	}

	visibleRows := m.watchlistVisibleRows()
	maxStart := max(0, len(m.config.Watchlist)-visibleRows)
	if m.watchlistScroll > maxStart {
		m.watchlistScroll = maxStart
	}
	if m.selectedIdx < m.watchlistScroll {
		m.watchlistScroll = m.selectedIdx
	}
	if m.selectedIdx >= m.watchlistScroll+visibleRows {
		m.watchlistScroll = m.selectedIdx - visibleRows + 1
	}
}
