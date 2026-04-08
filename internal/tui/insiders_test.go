package tui

import (
	"testing"

	"blackdesk/internal/domain"
)

func TestInsiderSummaryRowsUseVisibleTransactionsForFlow(t *testing.T) {
	rows := insiderSummaryRows(sampleInsiderSnapshot("AAPL"))
	if len(rows) != 5 {
		t.Fatalf("expected 5 insider summary rows, got %+v", rows)
	}
	if rows[1].price != "1.10K" {
		t.Fatalf("expected buy shares to be derived from transactions, got %+v", rows[1])
	}
	if rows[2].price != "750" {
		t.Fatalf("expected sell shares to be derived from transactions, got %+v", rows[2])
	}
	if rows[3].price != "350" {
		t.Fatalf("expected net shares to be derived from transactions, got %+v", rows[3])
	}
	if rows[4].price != "2.10M" {
		t.Fatalf("expected held to keep provider total, got %+v", rows[4])
	}
}

func TestColorizeInsiderTransactionMetricUsesOnlyBuySellSignals(t *testing.T) {
	if got := colorizeInsiderTransactionMetric("10", domain.InsiderTransaction{Action: "Buy"}); got != marketMoveStyle(1).Render("10") {
		t.Fatal("expected buy transactions to color metrics green")
	}
	if got := colorizeInsiderTransactionMetric("10", domain.InsiderTransaction{Action: "Sale"}); got != marketMoveStyle(-1).Render("10") {
		t.Fatal("expected sale transactions to color metrics red")
	}
	if got := colorizeInsiderTransactionMetric("10", domain.InsiderTransaction{Action: "Exercise"}); got != "10" {
		t.Fatal("expected non buy or sell actions to remain unstyled")
	}
}
