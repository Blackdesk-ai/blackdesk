package tui

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"blackdesk/internal/domain"
	"blackdesk/internal/ui"
)

func (m Model) renderQuoteFullscreenPage(header, section, label, muted, pos, neg lipgloss.Style, width, height int) string {
	switch m.quoteCenterMode {
	case quoteCenterAnalyst:
		return m.renderQuoteAnalystRecommendationsPage(header, section, label, muted, pos, neg, width, height)
	case quoteCenterEarnings:
		return m.renderQuoteEarningsPage(header, section, label, muted, pos, neg, width, height)
	default:
		return m.renderQuoteFilingsPage(header, section, label, muted, pos, neg, width, height)
	}
}

func (m Model) renderQuoteEarningsPage(header, section, label, muted, pos, neg lipgloss.Style, width, height int) string {
	snapshot := m.earningsForSymbol(m.activeSymbol())
	listWidth := clamp((width*3)/5, 46, 76)
	previewWidth := max(24, width-listWidth-3)
	bodyHeight := max(8, height-6)

	var b strings.Builder
	b.WriteString(section.Render("EARNINGS") + "\n\n")
	b.WriteString(m.renderQuoteEarningsSummary(header, muted, width) + "\n\n")
	left := lipgloss.NewStyle().Width(listWidth).Render(m.renderQuoteEarningsList(section, muted, pos, neg, listWidth, bodyHeight, snapshot))
	right := lipgloss.NewStyle().Width(previewWidth).Render(m.renderQuoteEarningsPreview(section, label, muted, pos, neg, previewWidth, bodyHeight, snapshot))
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, left, "   ", right))
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func (m Model) renderQuoteEarningsSummary(header, muted lipgloss.Style, width int) string {
	snapshot := m.earningsForSymbol(m.activeSymbol())
	company := snapshot.CompanyName
	if company == "" {
		company = m.quote.ShortName
	}
	title := header.Render(strings.ToUpper(m.activeSymbol()))
	if strings.TrimSpace(company) != "" {
		title += muted.Render("  " + company)
	}
	return renderStatusLine(width, title, "")
}

func (m Model) renderQuoteEarningsList(section, muted, pos, neg lipgloss.Style, width, height int, snapshot domain.EarningsSnapshot) string {
	var b strings.Builder
	b.WriteString(section.Render("EARNINGS HISTORY") + "\n\n")
	items := snapshot.Items
	if len(items) == 0 {
		if m.errEarnings != nil {
			b.WriteString(m.errEarnings.Error())
		} else {
			b.WriteString(muted.Render("No earnings data loaded for the active symbol"))
		}
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#14110D")).
		Background(lipgloss.Color("#E7B66B")).
		Bold(true).
		Width(width).
		MaxWidth(width)

	dateColWidth := 12
	typeColWidth := 9
	detailColWidth := max(12, width-dateColWidth-typeColWidth-4)

	headerLine := fmt.Sprintf(
		"%-*s  %-*s  %s",
		dateColWidth, "DATE",
		typeColWidth, "TYPE",
		truncateText("DETAIL", detailColWidth),
	)
	b.WriteString(muted.Render(truncateText(headerLine, width)) + "\n")
	b.WriteString(muted.Render(strings.Repeat("─", max(12, min(width, dateColWidth+typeColWidth+detailColWidth+4)))) + "\n")

	visibleRows := max(3, height/2)
	start := 0
	if m.earningsSel >= visibleRows {
		start = m.earningsSel - visibleRows + 1
	}
	end := min(len(items), start+visibleRows)
	for i := start; i < end; i++ {
		item := items[i]
		if i == m.earningsSel {
			line := fmt.Sprintf(
				"%-*s  %-*s  %s",
				dateColWidth, truncateText(earningsItemDateLabel(item), dateColWidth),
				typeColWidth, truncateText(strings.ToUpper(item.Kind), typeColWidth),
				truncateText(earningsItemDetailLine(item), detailColWidth),
			)
			line = truncateText(line, width)
			b.WriteString(selectedStyle.Render(line) + "\n")
			continue
		}
		b.WriteString(renderEarningsListRow(item, pos, neg, dateColWidth, typeColWidth, detailColWidth, width) + "\n")
	}
	b.WriteString("\n" + muted.Render("↑/↓ move • palette to switch views"))
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func (m Model) renderQuoteEarningsPreview(section, label, muted, pos, neg lipgloss.Style, width, height int, snapshot domain.EarningsSnapshot) string {
	var b strings.Builder
	b.WriteString(section.Render("PREVIEW") + "\n\n")
	item, ok := m.currentEarningsItem()
	if !ok {
		b.WriteString(muted.Render("Select an earnings entry to inspect it."))
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}

	b.WriteString(label.Render(earningsPreviewTitle(item)) + "\n")
	b.WriteString(muted.Render(earningsItemDateWindowLabel(item)) + "\n")
	if snapshot.CompanyName != "" {
		b.WriteString(muted.Render(truncateText(snapshot.CompanyName, width)) + "\n")
	}

	if item.Kind == "upcoming" {
		b.WriteString("\n" + muted.Render("Consensus") + "\n")
		b.WriteString(renderWrappedTextBlock(muted, fmt.Sprintf("EPS %s  •  Revenue %s", earningsEPSWindowLabel(item), earningsRevenueWindowLabel(item)), width))
		if summary, ok := earningsTrendSummary(snapshot, pos, neg); ok {
			b.WriteString("\n\n" + muted.Render("Trend") + "\n")
			b.WriteString(renderWrappedTextBlock(lipgloss.NewStyle(), summary, width))
		}
		if len(snapshot.Estimates) > 0 {
			b.WriteString("\n\n" + muted.Render("Estimate Trend") + "\n")
			for _, estimate := range snapshot.Estimates {
				line := fmt.Sprintf("%s  EPS %s  Rev %s", earningsPeriodLabel(estimate.Period), formatEarningsValue(estimate.EPSAverage), formatRevenueValue(estimate.RevenueAverage))
				b.WriteString(renderWrappedTextBlock(muted, line, width) + "\n")
			}
		}
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}

	b.WriteString("\n" + muted.Render("EPS") + "\n")
	b.WriteString(renderWrappedTextBlock(muted, fmt.Sprintf("Estimate %s  •  Actual %s", formatEarningsValue(item.EPSEstimate), formatEarningsValue(item.EPSActual)), width))
	if summary, ok := earningsTrendSummary(snapshot, pos, neg); ok {
		b.WriteString("\n\n" + muted.Render("Trend") + "\n")
		b.WriteString(renderWrappedTextBlock(lipgloss.NewStyle(), summary, width))
	}
	b.WriteString("\n\n" + muted.Render("Surprise") + "\n")
	b.WriteString(renderWrappedTextBlock(muted, earningsSurpriseLabel(item, pos, neg), width))
	if item.EPSDifference != 0 {
		b.WriteString("\n\n" + muted.Render("Difference") + "\n")
		b.WriteString(renderWrappedTextBlock(muted, formatSignedEarningsValue(item.EPSDifference), width))
	}
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func earningsItemDateLabel(item domain.EarningsItem) string {
	if !item.WindowStart.IsZero() {
		return item.WindowStart.Format("2006-01-02")
	}
	if !item.QuarterEnd.IsZero() {
		return item.QuarterEnd.Format("2006-01-02")
	}
	return "-"
}

func earningsItemDateWindowLabel(item domain.EarningsItem) string {
	if item.Kind == "upcoming" {
		if item.WindowStart.IsZero() && item.WindowEnd.IsZero() {
			return "Upcoming earnings window unavailable"
		}
		if item.WindowEnd.IsZero() || item.WindowStart.Equal(item.WindowEnd) {
			return "Expected on " + item.WindowStart.Format("2006-01-02")
		}
		return "Expected between " + item.WindowStart.Format("2006-01-02") + " and " + item.WindowEnd.Format("2006-01-02")
	}
	if item.QuarterEnd.IsZero() {
		return "Reported quarter"
	}
	return "Quarter ended " + item.QuarterEnd.Format("2006-01-02")
}

func earningsItemDetailLine(item domain.EarningsItem) string {
	if item.Kind == "upcoming" {
		return "Consensus " + earningsEPSWindowLabel(item)
	}
	if item.SurprisePercent > 0 {
		return "Beat " + ui.FormatPercent(item.SurprisePercent*100)
	}
	if item.SurprisePercent < 0 {
		return "Miss " + ui.FormatPercent(math.Abs(item.SurprisePercent*100))
	}
	if item.EPSActual != 0 {
		return "Reported " + formatEarningsValue(item.EPSActual)
	}
	return "Reported"
}

func renderEarningsListRow(item domain.EarningsItem, pos, neg lipgloss.Style, dateColWidth, typeColWidth, detailColWidth, width int) string {
	dateCol := fmt.Sprintf("%-*s", dateColWidth, truncateText(earningsItemDateLabel(item), dateColWidth))
	typeCol := fmt.Sprintf("%-*s", typeColWidth, truncateText(strings.ToUpper(item.Kind), typeColWidth))
	detail := renderEarningsDetailCell(item, pos, neg, detailColWidth)
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		dateCol,
		"  ",
		typeCol,
		"  ",
		lipgloss.NewStyle().MaxWidth(max(1, width-dateColWidth-typeColWidth-4)).Render(detail),
	)
}

func renderEarningsDetailCell(item domain.EarningsItem, pos, neg lipgloss.Style, width int) string {
	text := truncateText(earningsItemDetailLine(item), width)
	switch {
	case strings.HasPrefix(text, "Beat "):
		return pos.Render("Beat") + text[len("Beat"):]
	case strings.HasPrefix(text, "Miss "):
		return neg.Render("Miss") + text[len("Miss"):]
	default:
		return text
	}
}

func earningsPreviewTitle(item domain.EarningsItem) string {
	if item.Kind == "upcoming" {
		return "Next earnings"
	}
	return "Reported quarter"
}

func earningsEPSWindowLabel(item domain.EarningsItem) string {
	if item.EPSEstimate == 0 && item.EPSLow == 0 && item.EPSHigh == 0 {
		return "-"
	}
	if item.EPSLow != 0 || item.EPSHigh != 0 {
		return fmt.Sprintf("%s (%s to %s)", formatEarningsValue(item.EPSEstimate), formatEarningsValue(item.EPSLow), formatEarningsValue(item.EPSHigh))
	}
	return formatEarningsValue(item.EPSEstimate)
}

func earningsRevenueWindowLabel(item domain.EarningsItem) string {
	if item.RevenueAverage == 0 && item.RevenueLow == 0 && item.RevenueHigh == 0 {
		return "-"
	}
	if item.RevenueLow != 0 || item.RevenueHigh != 0 {
		return fmt.Sprintf("%s (%s to %s)", formatRevenueValue(item.RevenueAverage), formatRevenueValue(item.RevenueLow), formatRevenueValue(item.RevenueHigh))
	}
	return formatRevenueValue(item.RevenueAverage)
}

func earningsPeriodLabel(period string) string {
	switch strings.TrimSpace(period) {
	case "0q":
		return "This qtr"
	case "+1q":
		return "Next qtr"
	case "0y":
		return "This year"
	case "+1y":
		return "Next year"
	default:
		return strings.TrimSpace(period)
	}
}

func formatEarningsValue(v float64) string {
	if v == 0 {
		return "-"
	}
	return fmt.Sprintf("%.2f", v)
}

func formatSignedEarningsValue(v float64) string {
	if v == 0 {
		return "-"
	}
	return fmt.Sprintf("%+.2f", v)
}

func formatRevenueValue(v float64) string {
	if v == 0 {
		return "-"
	}
	return ui.FormatCompactInt(int64(math.Round(v)))
}

func earningsSurpriseLabel(item domain.EarningsItem, pos, neg lipgloss.Style) string {
	if item.SurprisePercent > 0 {
		return pos.Render("Beat " + ui.FormatPercent(item.SurprisePercent*100))
	}
	if item.SurprisePercent < 0 {
		return neg.Render("Miss " + ui.FormatPercent(math.Abs(item.SurprisePercent*100)))
	}
	if item.EPSActual != 0 {
		return "In line"
	}
	return "-"
}

type earningsTrendPoint struct {
	label string
	value float64
}

func earningsTrendSummary(snapshot domain.EarningsSnapshot, pos, neg lipgloss.Style) (string, bool) {
	points := make([]earningsTrendPoint, 0, 5)
	reported := make([]domain.EarningsItem, 0, len(snapshot.Items))
	for _, item := range snapshot.Items {
		if item.Kind != "reported" || item.QuarterEnd.IsZero() || item.EPSActual == 0 {
			continue
		}
		reported = append(reported, item)
	}
	sort.SliceStable(reported, func(i, j int) bool {
		return reported[i].QuarterEnd.Before(reported[j].QuarterEnd)
	})
	if len(reported) > 3 {
		reported = reported[len(reported)-3:]
	}
	for _, item := range reported {
		points = append(points, earningsTrendPoint{
			label: item.QuarterEnd.Format("Jan 2006"),
			value: item.EPSActual,
		})
	}
	for _, item := range snapshot.Items {
		if item.Kind == "upcoming" && item.EPSEstimate != 0 {
			points = append(points, earningsTrendPoint{
				label: "Next est",
				value: item.EPSEstimate,
			})
			break
		}
	}
	if len(points) < 2 {
		return "", false
	}
	values := make([]float64, 0, len(points))
	for _, point := range points {
		values = append(values, point.value)
	}
	label, style := earningsTrendDirection(values, pos, neg)
	return style.Render(label), true
}

func earningsTrendDirection(values []float64, pos, neg lipgloss.Style) (string, lipgloss.Style) {
	const tolerance = 0.01
	if len(values) < 2 {
		return "EPS trend is flat", lipgloss.NewStyle()
	}

	positiveMoves := 0
	negativeMoves := 0
	for i := 1; i < len(values); i++ {
		delta := values[i] - values[i-1]
		if delta > tolerance {
			positiveMoves++
		}
		if delta < -tolerance {
			negativeMoves++
		}
	}
	switch {
	case positiveMoves > 0 && negativeMoves > 0:
		return "EPS trend is mixed", lipgloss.NewStyle()
	case positiveMoves > 0:
		return "EPS trend is rising", pos
	case negativeMoves > 0:
		return "EPS trend is falling", neg
	default:
		return "EPS trend is flat", lipgloss.NewStyle()
	}
}
