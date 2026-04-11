package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"blackdesk/internal/domain"
	"blackdesk/internal/ui"
)

func overviewStatsGrid(label lipgloss.Style, width int, quote domain.QuoteSnapshot, fundamentals domain.FundamentalsSnapshot) []string {
	type statCell struct {
		label string
		value string
	}

	cols := [][]statCell{
		{
			{label: "Open", value: ui.FormatMoney(quote.Open)},
			{label: "Day low", value: ui.FormatMoney(quote.DayLow)},
			{label: "52W low", value: formatMoneyDash(fundamentals.FiftyTwoWeekLow)},
		},
		{
			{label: "Prev", value: ui.FormatMoney(quote.PreviousClose)},
			{label: "Day high", value: ui.FormatMoney(quote.DayHigh)},
			{label: "52W high", value: formatMoneyDash(fundamentals.FiftyTwoWeekHigh)},
		},
		{
			{label: "Avg vol", value: ui.FormatCompactInt(quote.AverageVolume)},
			{label: "Volume", value: ui.FormatCompactInt(quote.Volume)},
			{label: "Mkt cap", value: ui.FormatCompactInt(quote.MarketCap)},
		},
	}

	labelWidths := []int{9, 10, 10}
	gap := clamp(width/16, 3, 6)
	totalLabelWidth := 0
	for _, labelWidth := range labelWidths {
		totalLabelWidth += labelWidth
	}
	valueWidth := max(7, (width-totalLabelWidth-gap*(len(cols)-1))/len(cols))

	out := make([]string, 0, len(cols[0]))
	for rowIdx := range cols[0] {
		parts := make([]string, 0, len(cols))
		for colIdx, col := range cols {
			cell := col[rowIdx]
			labelCell := lipgloss.NewStyle().Width(labelWidths[colIdx]).Render(label.Render(cell.label))
			valueCell := lipgloss.NewStyle().Width(valueWidth).Render(cell.value)
			parts = append(parts, labelCell+valueCell)
		}
		out = append(out, strings.Join(parts, strings.Repeat(" ", gap)))
	}
	return out
}

func renderQuoteTechnicalsGrid(section, label, muted lipgloss.Style, quote domain.QuoteSnapshot, series domain.PriceSeries, width, height int) string {
	gap := 2
	colWidth := max(24, (width-gap)/2)
	snapshot := buildTechnicalSnapshot(quote, series)
	leftCol := strings.Join([]string{
		renderQuoteTechnicalsCard(section, label, muted, colWidth, technicalMomentumRows(snapshot), "MOMENTUM"),
		renderQuoteTechnicalsCard(section, label, muted, colWidth, technicalTrendRows(snapshot), "TREND"),
	}, "\n\n")
	rightCol := strings.Join([]string{
		renderQuoteTechnicalsCard(section, label, muted, colWidth, technicalStatRows(snapshot), "STAT EDGE"),
		renderQuoteTechnicalsCard(section, label, muted, colWidth, technicalVolatilityRows(snapshot), "VOLATILITY"),
		renderQuoteTechnicalsCard(section, label, muted, colWidth, technicalVolumeRows(snapshot), "VOLUME"),
	}, "\n\n")
	grid := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(colWidth).MaxWidth(colWidth).Render(leftCol),
		strings.Repeat(" ", gap),
		lipgloss.NewStyle().Width(colWidth).MaxWidth(colWidth).Render(rightCol),
	)
	return clipLines(grid, height)
}

func renderQuoteTechnicalsCard(section, label, muted lipgloss.Style, width int, rows []marketTableRow, title string) string {
	var b strings.Builder
	b.WriteString(section.Render(title) + "\n\n")
	b.WriteString(muted.Render(renderMarketTableHeaderWithLabels(width, "Name", "Value", "Signal", label)) + "\n")
	for _, row := range rows {
		b.WriteString(renderQuoteMetricTableRow(row, width, label) + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func renderFundamentalsTableHeader(labelWidth, valueWidth int) string {
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(labelWidth).Render(ansi.Truncate("Name", labelWidth, "...")),
		strings.Repeat(" ", fundamentalsColumnGap),
		lipgloss.NewStyle().Width(valueWidth).Render(ansi.Truncate("Value", valueWidth, "...")),
	)
}

func renderQuoteFundamentalsTableRow(row marketTableRow, labelWidth, valueWidth int, label lipgloss.Style) string {
	value := ansi.Truncate(colorizeMarketPrice(row.price, row.move, row.styled), valueWidth, "...")
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(labelWidth).Render(label.Render(ansi.Truncate(row.name, labelWidth, "..."))),
		strings.Repeat(" ", fundamentalsColumnGap),
		lipgloss.NewStyle().Width(valueWidth).Render(value),
	)
}

func renderQuoteMetricTableRow(row marketTableRow, width int, label lipgloss.Style) string {
	labelWidth := clamp(width/2, 8, 14)
	signalWidth := max(6, width-labelWidth-10)
	signal := ansi.Truncate(colorizeMarketChange(row.chg, row.move, row.styled), signalWidth, "")
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(labelWidth).Render(label.Render(row.name)),
		" ",
		lipgloss.NewStyle().Width(8).Render(colorizeMarketPrice(row.price, row.move, row.styled)),
		" ",
		lipgloss.NewStyle().Width(signalWidth).Render(signal),
	)
}
