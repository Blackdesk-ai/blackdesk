package tui

import (
	"errors"

	tea "github.com/charmbracelet/bubbletea"

	"blackdesk/internal/domain"
)

var sharpeHistoryRanges = []string{"5y"}
var statisticsHistoryRanges = []string{"5y", "3y", "2y", "1y"}
var statistics10YHistoryRanges = []string{"10y"}
var statisticsMaxHistoryRanges = []string{"10y", "7y", "5y", "3y", "2y", "1y"}

func (m Model) loadQuoteCmd(symbol string) tea.Cmd {
	return func() tea.Msg {
		q, err := m.services.GetQuote(m.ctx, symbol)
		return quoteLoadedMsg{symbol: symbol, quote: q, err: err}
	}
}

func (m Model) loadHistoryCmd(symbol string) tea.Cmd {
	current := ranges[m.rangeIdx]
	return func() tea.Msg {
		series, err := m.services.GetHistory(m.ctx, symbol, current.Range, current.Interval)
		return historyLoadedMsg{symbol: symbol, series: series, err: err}
	}
}

func (m Model) loadTechnicalHistoryCmd(symbol string) tea.Cmd {
	return func() tea.Msg {
		series, err := m.services.GetHistory(m.ctx, symbol, "2y", "1d")
		return technicalHistoryLoadedMsg{symbol: symbol, series: series, err: err}
	}
}

func (m Model) loadSharpeHistoryCmd(symbol string) tea.Cmd {
	return func() tea.Msg {
		series, err := m.loadSharpeHistory(symbol)
		return sharpeHistoryLoadedMsg{series: series, err: err}
	}
}

func (m Model) loadStatisticsHistoryCmd(symbol string) tea.Cmd {
	return func() tea.Msg {
		series, err := m.loadStatisticsHistory(symbol)
		return sharpeHistoryLoadedMsg{series: series, err: err}
	}
}

func (m Model) loadStatisticsHistory(symbol string) (domain.PriceSeries, error) {
	spec := m.statisticsRangeSpec()
	return m.loadHistoryWithFallback(symbol, spec.Ranges, "statistics history unavailable")
}

func (m Model) loadSharpeHistory(symbol string) (domain.PriceSeries, error) {
	return m.loadHistoryWithFallback(symbol, sharpeHistoryRanges, "sharpe history unavailable")
}

func (m Model) loadHistoryWithFallback(symbol string, rangeKeys []string, fallbackErr string) (domain.PriceSeries, error) {
	var lastErr error
	for _, rangeKey := range rangeKeys {
		series, err := m.services.GetHistory(m.ctx, symbol, rangeKey, "1d")
		if err == nil {
			return series, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = errors.New(fallbackErr)
	}
	return domain.PriceSeries{}, lastErr
}

func (m Model) loadNewsCmd(symbol string) tea.Cmd {
	return func() tea.Msg {
		items, err := m.services.GetNews(m.ctx, symbol)
		return newsLoadedMsg{items: items, err: err}
	}
}

func (m Model) loadFundamentalsCmd(symbol string) tea.Cmd {
	if data, ok := m.cachedFundamentals(symbol); ok {
		return func() tea.Msg {
			return fundamentalsLoadedMsg{symbol: symbol, data: data, err: nil}
		}
	}
	return func() tea.Msg {
		data, err := m.services.GetFundamentals(m.ctx, symbol)
		return fundamentalsLoadedMsg{symbol: symbol, data: data, err: err}
	}
}

func (m Model) loadStatementCmd(symbol string) tea.Cmd {
	if !m.services.HasStatements() {
		return nil
	}
	kind := m.statementKind
	frequency := m.statementFreq
	if data, ok := m.cachedStatement(symbol, kind, frequency); ok {
		return func() tea.Msg {
			return statementLoadedMsg{symbol: symbol, data: data, err: nil}
		}
	}
	return func() tea.Msg {
		data, err := m.services.GetStatement(m.ctx, symbol, kind, frequency)
		return statementLoadedMsg{symbol: symbol, data: data, err: err}
	}
}

func (m Model) loadInsidersCmd(symbol string) tea.Cmd {
	if !m.services.HasInsiders() {
		return nil
	}
	if data, ok := m.cachedInsiders(symbol); ok {
		return func() tea.Msg {
			return insidersLoadedMsg{data: data, err: nil}
		}
	}
	return func() tea.Msg {
		data, err := m.services.GetInsiders(m.ctx, symbol)
		return insidersLoadedMsg{data: data, err: err}
	}
}
