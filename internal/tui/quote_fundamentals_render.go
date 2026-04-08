package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"blackdesk/internal/domain"
)

func renderQuoteFundamentalsGrid(section, label, muted lipgloss.Style, quote domain.QuoteSnapshot, fundamentals domain.FundamentalsSnapshot, width, height int) string {
	gap := 2
	colWidth := max(24, (width-gap)/2)
	financialTop, financialBottom := splitFinancialFundamentalsRows(quoteFundamentalsFinancialRows(fundamentals))
	leftCol := strings.Join([]string{
		renderQuoteFundamentalsCard(section, label, muted, colWidth, quoteFundamentalsValuationRows(quote, fundamentals), "VALUATION"),
		renderQuoteFundamentalsCard(section, label, muted, colWidth, financialTop, "FINANCIALS"),
	}, "\n\n")
	rightCol := strings.Join([]string{
		renderQuoteFundamentalsCard(section, label, muted, colWidth, quoteFundamentalsProfitabilityRows(fundamentals), "PROFITABILITY"),
		renderQuoteFundamentalsCard(section, label, muted, colWidth, financialBottom, "FINANCIALS"),
	}, "\n\n")
	grid := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(colWidth).MaxWidth(colWidth).Render(leftCol),
		strings.Repeat(" ", gap),
		lipgloss.NewStyle().Width(colWidth).MaxWidth(colWidth).Render(rightCol),
	)
	return clipLines(grid, height)
}

func renderQuoteFundamentalsCard(section, label, muted lipgloss.Style, width int, rows []marketTableRow, title string) string {
	var b strings.Builder
	b.WriteString(section.Render(title) + "\n\n")
	b.WriteString(muted.Render(renderFundamentalsTableHeader(width)) + "\n")
	for _, row := range rows {
		b.WriteString(renderQuoteFundamentalsTableRow(row, width, label) + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}
