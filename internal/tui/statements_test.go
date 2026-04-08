package tui

import (
	"strings"
	"testing"
	"time"

	"blackdesk/internal/domain"
)

func TestStatementGrowthPercentQuarterlyUsesComparableQuarterLastYear(t *testing.T) {
	stmt := domain.FinancialStatement{
		Frequency: domain.StatementFrequencyQuarterly,
		Periods: []domain.StatementPeriod{
			{Label: "2024-09-28", EndDate: time.Date(2024, time.September, 28, 0, 0, 0, 0, time.UTC)},
			{Label: "2024-06-29", EndDate: time.Date(2024, time.June, 29, 0, 0, 0, 0, time.UTC)},
			{Label: "2024-03-30", EndDate: time.Date(2024, time.March, 30, 0, 0, 0, 0, time.UTC)},
			{Label: "2023-12-30", EndDate: time.Date(2023, time.December, 30, 0, 0, 0, 0, time.UTC)},
			{Label: "2023-09-30", EndDate: time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC)},
		},
	}
	values := []domain.StatementValue{
		{Value: 110, Present: true},
		{Value: 95, Present: true},
		{Value: 90, Present: true},
		{Value: 85, Present: true},
		{Value: 100, Present: true},
	}

	changePct, ok := statementGrowthPercent(stmt, values, 0)
	if !ok {
		t.Fatal("expected quarterly comparison to find same quarter last year")
	}
	if changePct != 10 {
		t.Fatalf("expected yearly quarterly growth of 10%%, got %.2f%%", changePct)
	}
}

func TestFormatStatementCellQuarterlyUsesYearAgoPeriod(t *testing.T) {
	stmt := domain.FinancialStatement{
		Frequency: domain.StatementFrequencyQuarterly,
		Periods: []domain.StatementPeriod{
			{Label: "2024-09-28", EndDate: time.Date(2024, time.September, 28, 0, 0, 0, 0, time.UTC)},
			{Label: "2024-06-29", EndDate: time.Date(2024, time.June, 29, 0, 0, 0, 0, time.UTC)},
			{Label: "2024-03-30", EndDate: time.Date(2024, time.March, 30, 0, 0, 0, 0, time.UTC)},
			{Label: "2023-12-30", EndDate: time.Date(2023, time.December, 30, 0, 0, 0, 0, time.UTC)},
			{Label: "2023-09-30", EndDate: time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC)},
		},
	}
	row := domain.StatementRow{
		Label: "Total Revenue",
		Values: []domain.StatementValue{
			{Value: 110_000_000, Present: true},
			{Value: 95_000_000, Present: true},
			{Value: 90_000_000, Present: true},
			{Value: 85_000_000, Present: true},
			{Value: 100_000_000, Present: true},
		},
	}

	cell := formatStatementCell(stmt, row, 0)
	if !strings.Contains(cell, "(+10.0%)") {
		t.Fatalf("expected quarterly cell growth to use same quarter last year, got %q", cell)
	}
}

func TestStatementGrowthInsightQuarterlyUsesYearAgoPeriod(t *testing.T) {
	stmt := domain.FinancialStatement{
		Frequency: domain.StatementFrequencyQuarterly,
		Periods: []domain.StatementPeriod{
			{Label: "2024-09-28", EndDate: time.Date(2024, time.September, 28, 0, 0, 0, 0, time.UTC)},
			{Label: "2024-06-29", EndDate: time.Date(2024, time.June, 29, 0, 0, 0, 0, time.UTC)},
			{Label: "2024-03-30", EndDate: time.Date(2024, time.March, 30, 0, 0, 0, 0, time.UTC)},
			{Label: "2023-12-30", EndDate: time.Date(2023, time.December, 30, 0, 0, 0, 0, time.UTC)},
			{Label: "2023-09-30", EndDate: time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC)},
		},
		Rows: []domain.StatementRow{
			{
				Key:   "TotalRevenue",
				Label: "Total Revenue",
				Values: []domain.StatementValue{
					{Value: 110, Present: true},
					{Value: 95, Present: true},
					{Value: 90, Present: true},
					{Value: 85, Present: true},
					{Value: 100, Present: true},
				},
			},
		},
	}

	row, ok := statementGrowthInsight("Revenue YoY (Q)", stmt, "TotalRevenue")
	if !ok {
		t.Fatal("expected quarterly insight to be available")
	}
	if row.Value != "+10.0%" {
		t.Fatalf("expected quarterly insight to use same quarter last year, got %+v", row)
	}
}
