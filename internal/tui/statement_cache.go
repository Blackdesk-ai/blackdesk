package tui

import (
	"strings"

	"blackdesk/internal/domain"
)

func (m Model) needsStatement(symbol string) bool {
	if strings.TrimSpace(symbol) == "" {
		return false
	}
	data, ok := m.cachedStatement(symbol, m.statementKind, m.statementFreq)
	if !ok {
		return true
	}
	return len(data.Rows) == 0
}

func statementCacheKey(symbol string, kind domain.StatementKind, freq domain.StatementFrequency) string {
	return strings.ToUpper(strings.TrimSpace(symbol)) + "|" + string(kind) + "|" + string(freq)
}

func (m Model) cachedStatement(symbol string, kind domain.StatementKind, freq domain.StatementFrequency) (domain.FinancialStatement, bool) {
	if strings.TrimSpace(symbol) == "" {
		return domain.FinancialStatement{}, false
	}
	if strings.EqualFold(m.statement.Symbol, symbol) && m.statement.Kind == kind && m.statement.Frequency == freq && len(m.statement.Rows) > 0 {
		return m.statement, true
	}
	if m.statementCache == nil {
		return domain.FinancialStatement{}, false
	}
	data, ok := m.statementCache[statementCacheKey(symbol, kind, freq)]
	if !ok || len(data.Rows) == 0 {
		return domain.FinancialStatement{}, false
	}
	return data, true
}

func (m *Model) cacheStatement(data domain.FinancialStatement) {
	if strings.TrimSpace(data.Symbol) == "" || data.Kind == "" || data.Frequency == "" || len(data.Rows) == 0 {
		return
	}
	if m.statementCache == nil {
		m.statementCache = make(map[string]domain.FinancialStatement)
	}
	m.statementCache[statementCacheKey(data.Symbol, data.Kind, data.Frequency)] = data
}

func (m Model) missingAIStatements(symbol string) []statementRequest {
	if strings.TrimSpace(symbol) == "" {
		return nil
	}
	if !m.services.HasStatements() {
		return nil
	}
	missing := make([]statementRequest, 0, len(aiStatementRequests))
	for _, req := range aiStatementRequests {
		if _, ok := m.cachedStatement(symbol, req.kind, req.frequency); ok {
			continue
		}
		missing = append(missing, req)
	}
	return missing
}
