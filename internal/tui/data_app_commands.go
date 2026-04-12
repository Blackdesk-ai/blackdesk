package tui

import (
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
	if plan.LoadFilings {
		cmds = append(cmds, m.loadFilingsCmd(symbol))
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
