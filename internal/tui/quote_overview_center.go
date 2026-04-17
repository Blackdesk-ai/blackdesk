package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"blackdesk/internal/domain"
	"blackdesk/internal/ui"
)

func (m Model) renderOverviewCenter(header, section, label, muted, pos, neg lipgloss.Style, chartWidth, height int) string {
	var b strings.Builder
	quote := m.quote
	displaySeries := m.displaySeries()
	if m.errQuote != nil {
		b.WriteString(header.Render(strings.ToUpper(m.activeSymbol())) + "\n")
		b.WriteString(neg.Render(m.errQuote.Error()))
		return b.String()
	}
	displayPrice, displayChange, displayChangePercent, sessionLabel := displayQuoteLine(quote)
	changeStyle := pos
	if displayChange < 0 {
		changeStyle = neg
	}
	priceStyle := pos
	if displayChange < 0 {
		priceStyle = neg
	}

	symbolLabel := strings.ToUpper(m.activeSymbol())
	if quote.Symbol != "" {
		symbolLabel = quote.Symbol
	}
	titleLine := header.Render(fmt.Sprintf("%s  ", symbolLabel)) +
		priceStyle.Render(ui.FormatMoney(displayPrice)) +
		header.Render(fmt.Sprintf(" %s", quote.Currency))
	badgeLine := renderBadgeRow(
		descriptorBadge(quote.MarketState, displayChangePercent),
		sessionLabel,
		rangeBadge(yearlyRangePct(m.fundamentals, quote.Price)),
	)
	headerWidth := chartWidth + 4
	if badgeLine != "" && lipgloss.Width(titleLine)+1+lipgloss.Width(badgeLine) <= headerWidth {
		b.WriteString(titleLine + strings.Repeat(" ", max(1, headerWidth-lipgloss.Width(titleLine)-lipgloss.Width(badgeLine))) + badgeLine + "\n")
	} else {
		b.WriteString(titleLine + "\n")
		if badgeLine != "" {
			b.WriteString(badgeLine + "\n")
		}
	}
	b.WriteString(changeStyle.Render(fmt.Sprintf("%+.2f (%+.2f%%)", displayChange, displayChangePercent)) + "\n\n")
	switch m.quoteCenterMode {
	case quoteCenterFundamentals:
		return m.renderOverviewFundamentals(section, label, muted, neg, quote, chartWidth, height, &b)
	case quoteCenterTechnicals:
		return m.renderOverviewTechnicals(section, label, muted, neg, quote, chartWidth, height, &b)
	case quoteCenterSharpe:
		return m.renderOverviewSharpe(section, label, muted, pos, neg, chartWidth, height, &b)
	case quoteCenterStatements:
		return m.renderOverviewStatements(section, label, muted, neg, chartWidth, height, &b)
	case quoteCenterInsiders:
		return m.renderOverviewInsiders(section, label, muted, neg, chartWidth, height, &b)
	case quoteCenterFilings:
		return m.renderOverviewFilings(section, label, muted, neg, chartWidth, height, &b)
	default:
		return m.renderOverviewChart(section, label, muted, neg, chartWidth, height, displaySeries, &b)
	}
}

func (m Model) renderOverviewFundamentals(section, label, muted, neg lipgloss.Style, quote domain.QuoteSnapshot, chartWidth, height int, b *strings.Builder) string {
	boardHeight := max(8, height-6)
	if !strings.EqualFold(m.fundamentals.Symbol, m.activeSymbol()) && m.errFundamentals == nil {
		b.WriteString(muted.Render("Loading " + strings.ToUpper(m.activeSymbol()) + " fundamentals…"))
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}
	b.WriteString(renderQuoteFundamentalsGrid(section, label, muted, quote, m.fundamentals, chartWidth, boardHeight))
	if m.errFundamentals != nil {
		b.WriteString("\n\n" + neg.Render("Fundamentals may be stale: "+m.errFundamentals.Error()))
	}
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func (m Model) renderOverviewTechnicals(section, label, muted, neg lipgloss.Style, quote domain.QuoteSnapshot, chartWidth, height int, b *strings.Builder) string {
	boardHeight := max(8, height-6)
	if len(m.technicalSeries(m.activeSymbol()).Candles) == 0 && m.errTechnicalHistory == nil {
		b.WriteString(muted.Render("Loading " + strings.ToUpper(m.activeSymbol()) + " technicals…"))
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}
	b.WriteString(renderQuoteTechnicalsGrid(section, label, muted, quote, m.technicalSeries(m.activeSymbol()), chartWidth, boardHeight))
	if m.errTechnicalHistory != nil {
		b.WriteString("\n\n" + neg.Render("Technicals may be stale: "+m.errTechnicalHistory.Error()))
	}
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func (m Model) renderOverviewSharpe(section, label, muted, pos, neg lipgloss.Style, chartWidth, height int, b *strings.Builder) string {
	boardHeight := max(8, height-6)
	b.WriteString(renderQuoteSharpeBoard(section, label, muted, pos, neg, chartWidth, boardHeight, m.sharpeRangeIdx, m.sharpeSeries(m.activeSymbol())))
	if m.errSharpeHistory != nil {
		b.WriteString("\n\n" + neg.Render("Risk-adjusted history may be stale: "+m.errSharpeHistory.Error()))
	}
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func (m Model) renderOverviewStatements(section, label, muted, neg lipgloss.Style, chartWidth, height int, b *strings.Builder) string {
	boardHeight := max(8, height-6)
	if (!strings.EqualFold(m.statement.Symbol, m.activeSymbol()) || m.statement.Kind != m.statementKind || m.statement.Frequency != m.statementFreq || len(m.statement.Rows) == 0) && m.errStatement == nil {
		b.WriteString(muted.Render("Loading " + strings.ToUpper(m.activeSymbol()) + " statements…"))
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}
	b.WriteString(renderQuoteStatementsBoard(section, label, muted, m.statement, m.statementKind, m.statementFreq, chartWidth, boardHeight))
	if m.errStatement != nil {
		b.WriteString("\n\n" + neg.Render("Statements may be stale: "+m.errStatement.Error()))
	}
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func (m Model) renderOverviewInsiders(section, label, muted, neg lipgloss.Style, chartWidth, height int, b *strings.Builder) string {
	boardHeight := max(8, height-6)
	b.WriteString(renderQuoteInsidersBoard(section, label, muted, m.insidersForSymbol(m.activeSymbol()), chartWidth, boardHeight))
	if m.errInsiders != nil {
		b.WriteString("\n\n" + neg.Render("Insiders may be stale: "+m.errInsiders.Error()))
	}
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func (m Model) renderOverviewFilings(section, label, muted, neg lipgloss.Style, chartWidth, height int, b *strings.Builder) string {
	snapshot := m.filingsForSymbol(m.activeSymbol())
	b.WriteString(section.Render("FILINGS") + " " + muted.Render("(palette)") + "\n\n")
	if len(snapshot.Items) == 0 {
		if m.errFilings != nil {
			b.WriteString(neg.Render(m.errFilings.Error()))
		} else {
			b.WriteString(muted.Render("No recent SEC filings loaded for the active symbol"))
		}
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}

	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#14110D")).Background(lipgloss.Color("#E7B66B")).Bold(true)
	visibleRows := max(4, height-7)
	start := 0
	if m.filingsSel >= visibleRows {
		start = m.filingsSel - visibleRows + 1
	}
	end := min(len(snapshot.Items), start+visibleRows)
	for i := start; i < end; i++ {
		item := snapshot.Items[i]
		left := fmt.Sprintf("%s  %s", item.Form, filingDateLabel(item))
		right := strings.TrimSpace(item.PrimaryDocDescription)
		if right == "" {
			right = strings.TrimSpace(item.PrimaryDocument)
		}
		line := renderStatusLine(chartWidth, left, right)
		if i == m.filingsSel {
			b.WriteString(selectedStyle.Width(chartWidth).Render(line) + "\n")
			continue
		}
		b.WriteString(line + "\n")
	}
	if m.errFilings != nil {
		b.WriteString("\n" + neg.Render("Filings may be stale: "+m.errFilings.Error()))
	}
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func (m Model) renderOverviewChart(section, label, muted, neg lipgloss.Style, chartWidth, height int, displaySeries domain.PriceSeries, b *strings.Builder) string {
	if len(displaySeries.Candles) == 0 && m.errHistory == nil {
		b.WriteString(muted.Render("Loading " + strings.ToUpper(m.activeSymbol()) + " chart…"))
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}
	chartHeight := max(9, height-12)
	plotWidth := ui.ChartPlotWidth(chartWidth)
	leftPad := strings.Repeat(" ", ui.ChartPlotPad())
	b.WriteString(ui.RenderLineChart(displaySeries.Candles, chartWidth, chartHeight) + "\n")
	b.WriteString(muted.Render(ui.RenderVolumeStrip(displaySeries.Candles, plotWidth)) + "\n")
	b.WriteString(muted.Render(ui.RenderTimeAxis(displaySeries.Candles, plotWidth)) + "\n")
	b.WriteString(leftPad + muted.Render(ui.RenderChartSummary(displaySeries)) + "\n\n")
	b.WriteString(leftPad + section.Render("TIMEFRAMES") + " ")
	selectedTimeframe := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#1C1712")).Background(lipgloss.Color("#E7B66B")).Padding(0, 1)
	for i, item := range ranges {
		text := item.Label
		if i == m.rangeIdx {
			b.WriteString(selectedTimeframe.Render(text) + " ")
			continue
		}
		b.WriteString(muted.Render(text) + " ")
	}
	b.WriteString(section.Render("←/→"))
	b.WriteString("\n\n")
	for _, row := range overviewStatsGrid(label, chartWidth, m.quote, m.fundamentals) {
		b.WriteString(leftPad + row + "\n")
	}
	if m.errHistory != nil {
		b.WriteString("\n" + neg.Render(m.errHistory.Error()))
	}
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}
