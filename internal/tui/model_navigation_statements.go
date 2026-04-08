package tui

import "blackdesk/internal/domain"

func (m *Model) cycleStatementKind(step int) {
	kinds := []domain.StatementKind{
		domain.StatementKindIncome,
		domain.StatementKindBalanceSheet,
		domain.StatementKindCashFlow,
	}
	idx := 0
	for i, kind := range kinds {
		if kind == m.statementKind {
			idx = i
			break
		}
	}
	m.statementKind = kinds[(idx+step+len(kinds))%len(kinds)]
}

func (m *Model) cycleStatementFrequency(step int) {
	freqs := []domain.StatementFrequency{
		domain.StatementFrequencyAnnual,
		domain.StatementFrequencyQuarterly,
	}
	idx := 0
	for i, freq := range freqs {
		if freq == m.statementFreq {
			idx = i
			break
		}
	}
	m.statementFreq = freqs[(idx+step+len(freqs))%len(freqs)]
}

func statementKindLabel(kind domain.StatementKind) string {
	switch kind {
	case domain.StatementKindBalanceSheet:
		return "balance sheet"
	case domain.StatementKindCashFlow:
		return "cash flow"
	default:
		return "income"
	}
}

func statementFrequencyLabel(freq domain.StatementFrequency) string {
	if freq == domain.StatementFrequencyQuarterly {
		return "quarterly"
	}
	return "annual"
}
