package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"

	"blackdesk/internal/domain"
)

func TestStatementRendererAddsExtraGapBeforeFirstPeriodColumn(t *testing.T) {
	mutedStyle := lipgloss.NewStyle()
	header := renderStatementHeader([]domain.StatementPeriod{
		{Label: "FY 2025"},
		{Label: "FY 2024"},
	}, len("Name"), len("FY 2025"), mutedStyle)
	if header != "Name  FY 2025 FY 2024" {
		t.Fatalf("expected extra gap before first period column, got %q", header)
	}

	row := renderStatementRow(domain.FinancialStatement{
		Frequency: domain.StatementFrequencyAnnual,
		Periods: []domain.StatementPeriod{
			{Label: "FY 2025"},
		},
	}, domain.StatementRow{
		Label: "Metric",
		Values: []domain.StatementValue{
			{Value: 0, Present: true},
		},
	}, 1, len("Metric"), len("0"), lipgloss.NewStyle())
	if row != "Metric  0" {
		t.Fatalf("expected extra gap before first value column, got %q", row)
	}
}
