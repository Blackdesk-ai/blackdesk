package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const maxNavigationHistory = 64

func (m Model) currentNavigationSnapshot() navigationSnapshot {
	return navigationSnapshot{
		tabIdx:             m.tabIdx,
		quoteCenterMode:    m.quoteCenterMode,
		globalPageOpen:     m.globalPageOpen,
		globalPageKind:     m.globalPageKind,
		rangeIdx:           m.rangeIdx,
		sharpeRangeIdx:     m.sharpeRangeIdx,
		statisticsRangeIdx: m.statisticsRangeIdx,
		statementKind:      m.statementKind,
		statementFreq:      m.statementFreq,
		filingsFilter:      m.filingsFilter,
		calendarFilter:     m.calendarFilter,
		aiFullscreen:       m.aiFullscreen,
		activeSymbol:       strings.ToUpper(strings.TrimSpace(m.activeSymbol())),
	}
}

func (m *Model) pushNavigationSnapshot(snapshot navigationSnapshot) {
	if snapshot.activeSymbol == "" && len(m.config.Watchlist) > 0 {
		snapshot.activeSymbol = strings.ToUpper(strings.TrimSpace(m.activeSymbol()))
	}
	if len(m.navigationStack) > 0 && navigationSnapshotEqual(m.navigationStack[len(m.navigationStack)-1], snapshot) {
		return
	}
	m.navigationStack = append(m.navigationStack, snapshot)
	if len(m.navigationStack) > maxNavigationHistory {
		m.navigationStack = append([]navigationSnapshot(nil), m.navigationStack[len(m.navigationStack)-maxNavigationHistory:]...)
	}
}

func navigationSnapshotEqual(a, b navigationSnapshot) bool {
	return a.tabIdx == b.tabIdx &&
		a.quoteCenterMode == b.quoteCenterMode &&
		a.globalPageOpen == b.globalPageOpen &&
		a.globalPageKind == b.globalPageKind &&
		a.rangeIdx == b.rangeIdx &&
		a.sharpeRangeIdx == b.sharpeRangeIdx &&
		a.statisticsRangeIdx == b.statisticsRangeIdx &&
		a.statementKind == b.statementKind &&
		a.statementFreq == b.statementFreq &&
		a.filingsFilter == b.filingsFilter &&
		a.calendarFilter == b.calendarFilter &&
		a.aiFullscreen == b.aiFullscreen &&
		strings.EqualFold(a.activeSymbol, b.activeSymbol)
}

func isNavigationBackKey(msg tea.KeyMsg) bool {
	switch msg.String() {
	case "ctrl+backspace", "ctrl+h":
		return true
	default:
		return false
	}
}

func (m Model) restorePreviousNavigation() (Model, tea.Cmd, bool) {
	if len(m.navigationStack) == 0 {
		return m, nil, false
	}

	snapshot := m.navigationStack[len(m.navigationStack)-1]
	m.navigationStack = append([]navigationSnapshot(nil), m.navigationStack[:len(m.navigationStack)-1]...)

	prevSymbol := strings.ToUpper(strings.TrimSpace(m.activeSymbol()))
	if snapshot.activeSymbol != "" {
		m.selectSymbol(snapshot.activeSymbol)
		m.ensureWatchlistSelectionVisible()
	}

	m.rangeIdx = snapshot.rangeIdx
	m.sharpeRangeIdx = snapshot.sharpeRangeIdx
	m.statisticsRangeIdx = snapshot.statisticsRangeIdx
	m.statementKind = snapshot.statementKind
	m.statementFreq = snapshot.statementFreq
	m.filingsFilter = snapshot.filingsFilter
	m.calendarFilter = snapshot.calendarFilter
	m.globalPageOpen = snapshot.globalPageOpen
	m.globalPageKind = snapshot.globalPageKind
	m.aiFullscreen = snapshot.aiFullscreen
	m.helpOpen = false
	m.commandPaletteOpen = false
	m.searchMode = false
	m.aiPickerOpen = false

	tabCmd := m.setActiveTab(snapshot.tabIdx)
	m.setQuoteCenterMode(snapshot.quoteCenterMode)

	cmds := make([]tea.Cmd, 0, 3)
	if tabCmd != nil {
		cmds = append(cmds, tabCmd)
	}

	activeSymbol := m.activeSymbol()
	symbolChanged := !strings.EqualFold(prevSymbol, activeSymbol)
	if snapshot.tabIdx == tabQuote && symbolChanged {
		cmds = append(cmds, m.loadAllCmd(activeSymbol))
	}

	if snapshot.globalPageOpen && snapshot.globalPageKind == globalPageCalendar {
		if data, ok := m.cachedCalendar(snapshot.calendarFilter); ok {
			m.calendar = data
			m.errCalendar = nil
		} else {
			cmds = append(cmds, m.loadCalendarCmd(snapshot.calendarFilter))
		}
	}

	if snapshot.tabIdx == tabQuote {
		switch snapshot.quoteCenterMode {
		case quoteCenterSharpe:
			if m.needsSharpeHistory(activeSymbol) {
				cmds = append(cmds, m.loadSharpeHistoryCmd(activeSymbol))
			}
		case quoteCenterStatistics:
			if m.needsStatisticsHistory(activeSymbol) {
				cmds = append(cmds, m.loadStatisticsHistoryCmd(activeSymbol))
			}
		case quoteCenterStatements:
			cmds = append(cmds, m.loadStatementCmd(activeSymbol))
		case quoteCenterInsiders:
			cmds = append(cmds, m.loadInsidersCmd(activeSymbol))
		case quoteCenterOwners:
			cmds = append(cmds, m.loadOwnersCmd(activeSymbol))
		case quoteCenterAnalyst:
			cmds = append(cmds, m.loadAnalystRecommendationsCmd(activeSymbol))
		case quoteCenterFilings:
			cmds = append(cmds, m.loadFilingsCmd(activeSymbol))
		case quoteCenterEarnings:
			cmds = append(cmds, m.loadEarningsCmd(activeSymbol))
		case quoteCenterTechnicals:
			if m.needsTechnicalHistory(activeSymbol) {
				cmds = append(cmds, m.loadTechnicalHistoryCmd(activeSymbol))
			}
		}
	}

	m.status = "Navigation back"
	return m, tea.Batch(cmds...), true
}
