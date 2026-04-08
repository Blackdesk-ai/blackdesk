package tui

func (m *Model) cycleNewsSelection() {
	if len(m.news) == 0 {
		return
	}
	m.newsSelected = (m.newsSelected + 1) % len(m.news)
}

func (m *Model) cycleMarketNewsSelection(step int) {
	if len(m.marketNews) == 0 {
		m.marketNewsSel = 0
		m.marketNewsScroll = 0
		return
	}
	m.marketNewsSel = (m.marketNewsSel + step + len(m.marketNews)) % len(m.marketNews)
	if m.marketNewsFresh != nil && m.marketNewsSel < len(m.marketNews) {
		delete(m.marketNewsFresh, marketNewsIdentity(m.marketNews[m.marketNewsSel]))
	}
	m.ensureMarketNewsSelectionVisible()
}

func (m *Model) ensureMarketNewsSelectionVisible() {
	if len(m.marketNews) == 0 {
		m.marketNewsSel = 0
		m.marketNewsScroll = 0
		return
	}
	if m.marketNewsSel < 0 {
		m.marketNewsSel = 0
	}
	if m.marketNewsSel >= len(m.marketNews) {
		m.marketNewsSel = len(m.marketNews) - 1
	}
	visibleRows := m.marketNewsVisibleRows()
	maxStart := max(0, len(m.marketNews)-visibleRows)
	if m.marketNewsScroll > maxStart {
		m.marketNewsScroll = maxStart
	}
	if m.marketNewsSel < m.marketNewsScroll {
		m.marketNewsScroll = m.marketNewsSel
	}
	if m.marketNewsSel >= m.marketNewsScroll+visibleRows {
		m.marketNewsScroll = m.marketNewsSel - visibleRows + 1
	}
}
