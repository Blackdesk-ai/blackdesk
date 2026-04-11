package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"blackdesk/internal/domain"
)

const fundamentalsColumnGap = 2

func renderQuoteFundamentalsGrid(section, label, muted lipgloss.Style, quote domain.QuoteSnapshot, fundamentals domain.FundamentalsSnapshot, width, height int) string {
	gap := 2
	financialLeft, financialRight := splitFinancialFundamentalsRows(quoteFundamentalsFinancialRows(fundamentals))
	profitabilityLeft, profitabilityRight := splitFundamentalsRows(quoteFundamentalsProfitabilityRows(quote, fundamentals))
	leftDesired, rightDesired := fundamentalsGridDesiredWidths(quote, fundamentals, profitabilityLeft, profitabilityRight, financialLeft, financialRight)
	if width < 60 {
		return renderQuoteFundamentalsStacked(section, label, muted, quote, fundamentals, width, height)
	}
	leftWidth, rightWidth, gap := fundamentalsGridLayout(width, gap, leftDesired, rightDesired)
	leftCol := renderQuoteFundamentalsCard(section, label, muted, leftWidth, quoteFundamentalsValuationRows(quote, fundamentals), "VALUATION")
	rightCol := strings.Join([]string{
		renderQuoteFundamentalsSplitCard(section, label, muted, rightWidth, profitabilityLeft, profitabilityRight, "PROFITABILITY"),
		renderQuoteFundamentalsSplitCard(section, label, muted, rightWidth, financialLeft, financialRight, "FINANCIALS"),
	}, "\n\n")
	grid := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(leftWidth).MaxWidth(leftWidth).Render(leftCol),
		strings.Repeat(" ", gap),
		lipgloss.NewStyle().Width(rightWidth).MaxWidth(rightWidth).Render(rightCol),
	)
	return clipLines(grid, height)
}

func renderQuoteFundamentalsStacked(section, label, muted lipgloss.Style, quote domain.QuoteSnapshot, fundamentals domain.FundamentalsSnapshot, width, height int) string {
	profitabilityRows := quoteFundamentalsProfitabilityRows(quote, fundamentals)
	financialRows := quoteFundamentalsFinancialRows(fundamentals)
	cards := []string{
		renderQuoteFundamentalsCard(section, label, muted, width, quoteFundamentalsValuationRows(quote, fundamentals), "VALUATION"),
		renderQuoteFundamentalsCard(section, label, muted, width, profitabilityRows, "PROFITABILITY"),
		renderQuoteFundamentalsCard(section, label, muted, width, financialRows, "FINANCIALS"),
	}
	return clipLines(strings.Join(cards, "\n\n"), height)
}

func renderQuoteFundamentalsCard(section, label, muted lipgloss.Style, width int, rows []marketTableRow, title string) string {
	var b strings.Builder
	b.WriteString(section.Render(title) + "\n\n")
	labelWidth, valueWidth := fundamentalsTableWidths(rows, width, 18)
	b.WriteString(muted.Render(renderFundamentalsTableHeader(labelWidth, valueWidth)) + "\n")
	for _, row := range rows {
		b.WriteString(renderQuoteFundamentalsTableRow(row, labelWidth, valueWidth, label) + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func renderQuoteFundamentalsSplitCard(section, label, muted lipgloss.Style, width int, leftRows, rightRows []marketTableRow, title string) string {
	var b strings.Builder
	b.WriteString(section.Render(title) + "\n\n")

	gap := clamp(width/28, 1, 3)
	leftWidth, rightWidth := fundamentalsSplitCardWidths(width, gap, leftRows, rightRows)
	left := renderQuoteFundamentalsTableBlock(muted, label, leftWidth, leftRows)
	right := renderQuoteFundamentalsTableBlock(muted, label, rightWidth, rightRows)
	b.WriteString(lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(leftWidth).MaxWidth(leftWidth).Render(left),
		strings.Repeat(" ", gap),
		lipgloss.NewStyle().Width(rightWidth).MaxWidth(rightWidth).Render(right),
	))
	return strings.TrimRight(b.String(), "\n")
}

func renderQuoteFundamentalsTableBlock(muted, label lipgloss.Style, width int, rows []marketTableRow) string {
	var b strings.Builder
	labelWidth, valueWidth := fundamentalsSplitTableWidths(rows, width)
	b.WriteString(muted.Render(renderFundamentalsSplitTableHeader(labelWidth, valueWidth)) + "\n")
	for _, row := range rows {
		b.WriteString(renderQuoteFundamentalsSplitTableRow(row, labelWidth, valueWidth, label) + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func fundamentalsGridDesiredWidths(quote domain.QuoteSnapshot, fundamentals domain.FundamentalsSnapshot, profitabilityLeft, profitabilityRight, financialLeft, financialRight []marketTableRow) (int, int) {
	leftDesired := fundamentalsTableDesiredWidth(quoteFundamentalsValuationRows(quote, fundamentals), 16)
	rightDesired := max(
		fundamentalsSplitDesiredWidth(profitabilityLeft, profitabilityRight, 3),
		fundamentalsSplitDesiredWidth(financialLeft, financialRight, 3),
	)
	return leftDesired, rightDesired
}

func fundamentalsGridColumnWidths(width, gap, leftDesired, rightDesired int) (int, int) {
	available := max(44, width-gap)
	if leftDesired+rightDesired <= available {
		return leftDesired, rightDesired
	}
	totalDesired := max(1, leftDesired+rightDesired)
	leftWidth := clamp(available*leftDesired/totalDesired, 22, available-22)
	rightWidth := max(22, available-leftWidth)
	return leftWidth, rightWidth
}

func fundamentalsGridLayout(width, baseGap, leftDesired, rightDesired int) (int, int, int) {
	leftWidth, rightWidth := fundamentalsGridColumnWidths(width, baseGap, leftDesired, rightDesired)
	used := leftWidth + rightWidth + baseGap
	slack := max(0, width-used)
	gap := baseGap + min(4, slack/6)
	slack -= gap - baseGap
	leftWidth += slack / 2
	rightWidth += slack - slack/2
	return leftWidth, rightWidth, gap
}

func fundamentalsTableDesiredWidth(rows []marketTableRow, maxLabelWidth int) int {
	labelWidth, valueWidth := fundamentalsTableWidths(rows, max(22, maxLabelWidth+9), maxLabelWidth)
	return max(22, labelWidth+fundamentalsColumnGap+valueWidth)
}

func fundamentalsTableWidths(rows []marketTableRow, width, maxLabelWidth int) (int, int) {
	labelContentWidth := lipgloss.Width("Name")
	valueContentWidth := lipgloss.Width("Value")
	for _, row := range rows {
		labelContentWidth = max(labelContentWidth, lipgloss.Width(row.name))
		valueContentWidth = max(valueContentWidth, lipgloss.Width(row.price))
	}
	labelWidth := clamp(labelContentWidth, 8, max(8, min(maxLabelWidth, width-(8+fundamentalsColumnGap))))
	valueWidth := max(8, width-labelWidth-fundamentalsColumnGap)
	desiredValueWidth := min(valueContentWidth, max(8, width-(8+fundamentalsColumnGap)))
	if desiredValueWidth > valueWidth {
		labelWidth = max(8, labelWidth-(desiredValueWidth-valueWidth))
		valueWidth = max(8, width-labelWidth-fundamentalsColumnGap)
	}
	if labelWidth+fundamentalsColumnGap+valueWidth > width {
		labelWidth = max(8, width-valueWidth-fundamentalsColumnGap)
	}
	if labelWidth+fundamentalsColumnGap+valueWidth > width {
		valueWidth = max(8, width-labelWidth-fundamentalsColumnGap)
	}
	return labelWidth, valueWidth
}

func fundamentalsSplitDesiredWidth(leftRows, rightRows []marketTableRow, gap int) int {
	leftLabelWidth, leftValueWidth := fundamentalsSplitTableWidths(leftRows, 28)
	rightLabelWidth, rightValueWidth := fundamentalsSplitTableWidths(rightRows, 28)
	return leftLabelWidth + fundamentalsColumnGap + leftValueWidth + gap + rightLabelWidth + fundamentalsColumnGap + rightValueWidth
}

func fundamentalsSplitCardWidths(width, gap int, leftRows, rightRows []marketTableRow) (int, int) {
	available := max(24, width-gap)
	leftDesired := fundamentalsSplitBlockDesiredWidth(leftRows)
	rightDesired := fundamentalsSplitBlockDesiredWidth(rightRows)
	if leftDesired+rightDesired <= available {
		return leftDesired, rightDesired
	}
	totalDesired := max(1, leftDesired+rightDesired)
	leftWidth := clamp(available*leftDesired/totalDesired, 12, available-12)
	rightWidth := max(12, available-leftWidth)
	return leftWidth, rightWidth
}

func fundamentalsSplitBlockDesiredWidth(rows []marketTableRow) int {
	labelWidth, valueWidth := fundamentalsSplitTableWidths(rows, 28)
	return labelWidth + fundamentalsColumnGap + valueWidth
}

func fundamentalsSplitTableWidths(rows []marketTableRow, width int) (int, int) {
	labelContentWidth := lipgloss.Width("Name")
	valueContentWidth := lipgloss.Width("Value")
	for _, row := range rows {
		labelContentWidth = max(labelContentWidth, lipgloss.Width(row.name))
		valueContentWidth = max(valueContentWidth, lipgloss.Width(row.price))
	}
	labelWidth := clamp(labelContentWidth, 8, max(8, width-(6+fundamentalsColumnGap)))
	valueWidth := max(6, width-labelWidth-fundamentalsColumnGap)
	desiredValueWidth := min(valueContentWidth, max(6, width-(8+fundamentalsColumnGap)))
	if desiredValueWidth > valueWidth {
		labelWidth = max(8, labelWidth-(desiredValueWidth-valueWidth))
		valueWidth = max(6, width-labelWidth-fundamentalsColumnGap)
	}
	if labelWidth+fundamentalsColumnGap+valueWidth > width {
		labelWidth = max(8, width-valueWidth-fundamentalsColumnGap)
	}
	if labelWidth+fundamentalsColumnGap+valueWidth > width {
		valueWidth = max(6, width-labelWidth-fundamentalsColumnGap)
	}
	return labelWidth, valueWidth
}

func renderFundamentalsSplitTableHeader(labelWidth, valueWidth int) string {
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(labelWidth).Render(ansi.Truncate("Name", labelWidth, "...")),
		strings.Repeat(" ", fundamentalsColumnGap),
		lipgloss.NewStyle().Width(valueWidth).Render(ansi.Truncate("Value", valueWidth, "...")),
	)
}

func renderQuoteFundamentalsSplitTableRow(row marketTableRow, labelWidth, valueWidth int, label lipgloss.Style) string {
	value := ansi.Truncate(colorizeMarketPrice(row.price, row.move, row.styled), valueWidth, "...")
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(labelWidth).Render(label.Render(ansi.Truncate(row.name, labelWidth, "..."))),
		strings.Repeat(" ", fundamentalsColumnGap),
		lipgloss.NewStyle().Width(valueWidth).Render(value),
	)
}
