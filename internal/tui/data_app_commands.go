package tui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"blackdesk/internal/application"
)

func (m Model) loadAllCmd(symbol string) tea.Cmd {
	plan := application.PlanQuoteWorkspaceLoad(
		m.applicationQuoteCenterMode(),
		m.needsTechnicalHistory(symbol),
		m.services.HasStatements(),
		m.services.HasInsiders(),
		m.services.HasOwners(),
	)
	cmds := make([]tea.Cmd, 0, 10)
	if plan.LoadQuote {
		cmds = append(cmds, m.loadQuoteCmd(symbol))
	}
	if plan.LoadWatchlistQuotes {
		cmds = append(cmds, m.loadWatchlistQuotesCmd(symbol))
	}
	if plan.LoadMarketQuotes {
		cmds = append(cmds, m.loadMarketQuotesCmd())
	}
	if plan.LoadHistory {
		cmds = append(cmds, m.loadHistoryCmd(symbol))
	}
	if plan.LoadNews {
		cmds = append(cmds, m.loadNewsCmd(symbol))
	}
	if plan.LoadFundamentals {
		cmds = append(cmds, m.loadFundamentalsCmd(symbol))
	}
	if plan.LoadTechnical {
		cmds = append(cmds, m.loadTechnicalHistoryCmd(symbol))
	}
	if plan.LoadStatement {
		cmds = append(cmds, m.loadStatementCmd(symbol))
	}
	if plan.LoadInsiders {
		cmds = append(cmds, m.loadInsidersCmd(symbol))
	}
	if plan.LoadOwners {
		cmds = append(cmds, m.loadOwnersCmd(symbol))
	}
	if plan.LoadAnalyst {
		cmds = append(cmds, m.loadAnalystRecommendationsCmd(symbol))
	}
	if plan.LoadFilings {
		cmds = append(cmds, m.loadFilingsCmd(symbol))
	}
	if plan.LoadEarnings {
		cmds = append(cmds, m.loadEarningsCmd(symbol))
	}
	cmds = append(cmds, m.loadMarketRiskCmd())
	return tea.Batch(cmds...)
}

const searchDebounceDelay = 250 * time.Millisecond

func (m Model) searchDebounceCmd(query string, id int) tea.Cmd {
	return tea.Tick(searchDebounceDelay, func(time.Time) tea.Msg {
		return searchDebouncedMsg{id: id, query: query}
	})
}

func (m Model) searchCmd(query string, id int) tea.Cmd {
	return func() tea.Msg {
		results, err := m.services.SearchSymbols(m.ctx, query)
		return searchLoadedMsg{id: id, query: query, results: results, err: err}
	}
}

func (m Model) commandPaletteDebounceCmd(query string, id int) tea.Cmd {
	return tea.Tick(searchDebounceDelay, func(time.Time) tea.Msg {
		return commandPaletteDebouncedMsg{id: id, query: query}
	})
}

func (m Model) commandPaletteSearchCmd(query string, id int) tea.Cmd {
	return func() tea.Msg {
		results, err := m.services.SearchSymbols(m.ctx, query)
		return commandPaletteLoadedMsg{id: id, query: query, results: results, err: err}
	}
}

func (m Model) persistCmd() tea.Cmd {
	cfg := m.config
	return func() tea.Msg {
		_ = m.services.SaveConfig(cfg)
		return nil
	}
}

func (m Model) loadFilingsCmd(symbol string) tea.Cmd {
	if m.services == nil || !m.services.HasFilings() {
		return nil
	}
	return func() tea.Msg {
		data, err := m.services.GetFilings(m.ctx, symbol)
		return filingsLoadedMsg{data: data, err: err}
	}
}

func (m Model) loadAnalystRecommendationsCmd(symbol string) tea.Cmd {
	if m.services == nil || !m.services.HasAnalystRecommendations() {
		return nil
	}
	if data, ok := m.cachedAnalystRecommendations(symbol); ok {
		return func() tea.Msg {
			return analystRecommendationsLoadedMsg{data: data, err: nil}
		}
	}
	return func() tea.Msg {
		data, err := m.services.GetAnalystRecommendations(m.ctx, symbol)
		return analystRecommendationsLoadedMsg{data: data, err: err}
	}
}

func (m Model) loadOwnersCmd(symbol string) tea.Cmd {
	if m.services == nil || !m.services.HasOwners() {
		return nil
	}
	if data, ok := m.cachedOwners(symbol); ok {
		return func() tea.Msg {
			return ownersLoadedMsg{data: data, err: nil}
		}
	}
	return func() tea.Msg {
		data, err := m.services.GetOwners(m.ctx, symbol)
		return ownersLoadedMsg{data: data, err: err}
	}
}

func (m Model) loadEarningsCmd(symbol string) tea.Cmd {
	if m.services == nil || !m.services.HasEarnings() {
		return nil
	}
	if data, ok := m.cachedEarnings(symbol); ok {
		return func() tea.Msg {
			return earningsLoadedMsg{data: data, err: nil}
		}
	}
	return func() tea.Msg {
		data, err := m.services.GetEarnings(m.ctx, symbol)
		return earningsLoadedMsg{data: data, err: err}
	}
}

func (m Model) loadCalendarCmd(filter calendarFilterMode) tea.Cmd {
	if m.services == nil || !m.services.HasEconomicCalendar() {
		return nil
	}
	if data, ok := m.cachedCalendar(filter); ok {
		return func() tea.Msg {
			return calendarLoadedMsg{filter: filter, data: data, err: nil}
		}
	}
	start, end := m.calendarRangeForFilter(filter)
	return func() tea.Msg {
		data, err := m.services.GetEconomicCalendar(m.ctx, start, end)
		return calendarLoadedMsg{filter: filter, data: data, err: err}
	}
}

func (m Model) calendarRangeForFilter(filter calendarFilterMode) (time.Time, time.Time) {
	now := m.clock
	if now.IsZero() {
		now = time.Now()
	}
	localNow := now.In(now.Location())
	start := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, localNow.Location())
	switch filter {
	case calendarFilterThisWeek:
		return start, start.AddDate(0, 0, 7)
	default:
		return start, start.AddDate(0, 0, 1)
	}
}

func (m Model) calendarFilterLabel() string {
	switch m.calendarFilter {
	case calendarFilterThisWeek:
		return "This Week"
	default:
		return "Today"
	}
}

func (m Model) globalPageStatusLine() string {
	if !m.globalPageOpen {
		return ""
	}
	switch m.globalPageKind {
	case globalPageCalendar:
		return "Calendar: ←/→ filter • ↑/↓ move • r refresh • Esc close"
	default:
		return strings.TrimSpace(m.status)
	}
}
