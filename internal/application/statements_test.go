package application

import (
	"testing"

	"blackdesk/internal/domain"
)

func TestStepStatementKindCyclesAcrossKinds(t *testing.T) {
	result := StepStatementKind(domain.StatementKindIncome, 1)
	if result.Kind != domain.StatementKindBalanceSheet || result.Status != "Statements: balance sheet" {
		t.Fatalf("unexpected statement kind result: %+v", result)
	}

	result = StepStatementKind(domain.StatementKindIncome, -1)
	if result.Kind != domain.StatementKindCashFlow || result.Status != "Statements: cash flow" {
		t.Fatalf("unexpected wrapped statement kind result: %+v", result)
	}
}

func TestStepStatementFrequencyCyclesAcrossFrequencies(t *testing.T) {
	result := StepStatementFrequency(domain.StatementFrequencyAnnual, 1)
	if result.Frequency != domain.StatementFrequencyQuarterly || result.Status != "Statements frequency: quarterly" {
		t.Fatalf("unexpected statement frequency result: %+v", result)
	}

	result = StepStatementFrequency(domain.StatementFrequencyAnnual, -1)
	if result.Frequency != domain.StatementFrequencyQuarterly {
		t.Fatalf("unexpected wrapped statement frequency result: %+v", result)
	}
}
