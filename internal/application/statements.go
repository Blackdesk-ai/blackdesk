package application

import "blackdesk/internal/domain"

type StatementKindStepResult struct {
	Kind   domain.StatementKind
	Status string
}

type StatementFrequencyStepResult struct {
	Frequency domain.StatementFrequency
	Status    string
}

func StepStatementKind(current domain.StatementKind, step int) StatementKindStepResult {
	kinds := []domain.StatementKind{
		domain.StatementKindIncome,
		domain.StatementKindBalanceSheet,
		domain.StatementKindCashFlow,
	}
	idx := 0
	for i, kind := range kinds {
		if kind == current {
			idx = i
			break
		}
	}
	next := kinds[(idx+step+len(kinds))%len(kinds)]
	return StatementKindStepResult{
		Kind:   next,
		Status: "Statements: " + statementKindLabel(next),
	}
}

func StepStatementFrequency(current domain.StatementFrequency, step int) StatementFrequencyStepResult {
	freqs := []domain.StatementFrequency{
		domain.StatementFrequencyAnnual,
		domain.StatementFrequencyQuarterly,
	}
	idx := 0
	for i, freq := range freqs {
		if freq == current {
			idx = i
			break
		}
	}
	next := freqs[(idx+step+len(freqs))%len(freqs)]
	return StatementFrequencyStepResult{
		Frequency: next,
		Status:    "Statements frequency: " + statementFrequencyLabel(next),
	}
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
