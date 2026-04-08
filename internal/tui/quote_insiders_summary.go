package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"blackdesk/internal/domain"
	"blackdesk/internal/ui"
)

func renderQuoteInsidersBoard(section, label, muted lipgloss.Style, snapshot domain.InsiderSnapshot, width, height int) string {
	var b strings.Builder
	b.WriteString(renderInsiderSummaryCard(section, label, muted, width, snapshot) + "\n\n")
	b.WriteString(renderInsiderTransactionsCard(section, label, muted, width, snapshot.Transactions))
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func renderInsiderSummaryCard(section, label, muted lipgloss.Style, width int, snapshot domain.InsiderSnapshot) string {
	rows := insiderSummaryRows(snapshot)
	var b strings.Builder
	b.WriteString(section.Render("INSIDER FLOW") + "\n\n")
	if len(rows) == 0 {
		b.WriteString(muted.Render("Insider activity unavailable for the active symbol"))
		return b.String()
	}
	b.WriteString(muted.Render(renderFundamentalsTableHeader(width)) + "\n")
	for _, row := range rows {
		b.WriteString(renderQuoteFundamentalsTableRow(row, width, label) + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func renderInsiderTransactionsCard(section, label, muted lipgloss.Style, width int, items []domain.InsiderTransaction) string {
	var b strings.Builder
	b.WriteString(section.Render("RECENT TRANSACTIONS") + "\n\n")
	if len(items) == 0 {
		b.WriteString(muted.Render("No recent insider transactions reported"))
		return b.String()
	}
	b.WriteString(renderInsiderTransactionHeader(width, items, label, muted) + "\n")
	maxRows := max(4, min(8, width/5))
	if maxRows > len(items) {
		maxRows = len(items)
	}
	for _, item := range items[:maxRows] {
		b.WriteString(renderInsiderTransactionRow(item, width, items, label, muted) + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func insiderSummaryRows(snapshot domain.InsiderSnapshot) []marketTableRow {
	activity := snapshot.PurchaseActivity
	if len(snapshot.Transactions) > 0 {
		activity = summarizeInsiderTransactions(snapshot)
	}
	if len(snapshot.Transactions) == 0 && len(snapshot.Roster) == 0 && activity.TotalInsiderShares == 0 && activity.BuyShares == 0 && activity.SellShares == 0 {
		return nil
	}
	period := strings.TrimSpace(activity.Period)
	if period == "" {
		period = "6m"
	}
	netMove := float64(activity.NetShares)
	return []marketTableRow{
		{name: "Period", price: strings.ToUpper(period), chg: "", move: 0, styled: false},
		{name: "Buy shares", price: ui.FormatCompactInt(activity.BuyShares), chg: fmt.Sprintf("%d tx", activity.BuyTransactions), move: float64(activity.BuyShares), styled: activity.BuyShares > 0},
		{name: "Sell shares", price: ui.FormatCompactInt(activity.SellShares), chg: fmt.Sprintf("%d tx", activity.SellTransactions), move: -float64(activity.SellShares), styled: activity.SellShares > 0},
		{name: "Net shares", price: ui.FormatCompactInt(activity.NetShares), chg: percentDash(activity.NetPercentInsiderShares), move: netMove, styled: activity.NetShares != 0},
		{name: "Held", price: ui.FormatCompactInt(activity.TotalInsiderShares), chg: fmt.Sprintf("%d tx", activity.NetTransactions), move: 0, styled: false},
	}
}

func summarizeInsiderTransactions(snapshot domain.InsiderSnapshot) domain.InsiderPurchaseActivity {
	activity := snapshot.PurchaseActivity
	activity.BuyShares = 0
	activity.BuyTransactions = 0
	activity.SellShares = 0
	activity.SellTransactions = 0
	activity.NetShares = 0
	activity.NetTransactions = 0
	activity.NetPercentInsiderShares = 0
	for _, item := range snapshot.Transactions {
		switch strings.TrimSpace(item.Action) {
		case "Buy":
			activity.BuyShares += item.Shares
			activity.BuyTransactions++
		case "Sale":
			activity.SellShares += item.Shares
			activity.SellTransactions++
		}
	}
	activity.NetShares = activity.BuyShares - activity.SellShares
	activity.NetTransactions = activity.BuyTransactions + activity.SellTransactions
	return activity
}
