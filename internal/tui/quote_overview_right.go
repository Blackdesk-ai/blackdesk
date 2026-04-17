package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderOverviewRight(section, label, muted lipgloss.Style, width, height int) string {
	if m.quoteCenterMode == quoteCenterSharpe {
		return m.renderOverviewSharpeRight(section, label, muted, width, height)
	}
	if m.quoteCenterMode == quoteCenterFilings {
		return m.renderOverviewFilingsRight(section, label, muted, width, height)
	}
	var b strings.Builder
	upside := analystUpsideValue(m.fundamentals, m.quote.Price)
	upsideLine := analystUpsideLine(m.fundamentals, m.quote.Price)
	if upsideLine != "-" {
		upsideLine = marketMoveStyle(upside).Render(upsideLine)
	}
	b.WriteString(section.Render("MARKET HEAT") + "\n\n")
	b.WriteString(renderHeatMeter("Day swing", clampFloat(math.Abs(m.quote.ChangePercent)/4, 0, 1), 10) + "\n")
	b.WriteString(renderHeatMeter("Volume", volumePulse(m.quote), 10) + "\n")
	b.WriteString(renderHeatMeter("52W pos", yearlyRangePct(m.fundamentals, m.quote.Price), 10) + "\n\n")
	b.WriteString(section.Render("ANALYSTS") + "\n\n")
	b.WriteString(fmt.Sprintf("%s %s %s\n", label.Render("Target"), analystTargetLine(m.fundamentals), colorizeRecommendationBadge(m.fundamentals.RecommendationKey)))
	b.WriteString(fmt.Sprintf("%s %s\n", label.Render("Upside"), upsideLine))
	b.WriteString(fmt.Sprintf("%s %s\n", label.Render("Fwd PE"), formatMetricFloat(m.fundamentals.ForwardPE)))
	b.WriteString(fmt.Sprintf("%s %s\n", label.Render("PEG"), formatMetricFloat(pegRatioValue(m.quote, m.fundamentals))))
	b.WriteString(fmt.Sprintf("%s %s\n", label.Render("EPS TTM"), formatMetricFloat(m.fundamentals.EPS)))
	b.WriteString(fmt.Sprintf("%s %s\n", label.Render("Beta"), formatMetricFloat(m.fundamentals.Beta)))
	b.WriteString("\n" + section.Render("AI INSIGHT") + " " + muted.Render("(i)") + "\n\n")
	b.WriteString(m.renderQuoteInsightBlock(muted, width))
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func (m Model) renderOverviewSharpeRight(section, label, muted lipgloss.Style, width, height int) string {
	chartSeries := displaySharpeSeriesForRange(buildSharpeChartSeries(m.sharpeSeries(m.activeSymbol())), ranges[m.sharpeRangeIdx].Range)
	pos := lipgloss.NewStyle().Foreground(lipgloss.Color("#62D394"))
	neg := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF7A73"))
	return renderQuoteSharpePreview(label, muted, pos, neg, width, height, chartSeries)
}

func (m Model) renderOverviewFilingsRight(section, label, muted lipgloss.Style, width, height int) string {
	var b strings.Builder
	b.WriteString(section.Render("FILING PREVIEW") + "\n\n")
	snapshot := m.filingsForSymbol(m.activeSymbol())
	item, ok := m.currentFiling()
	if !ok {
		if m.errFilings != nil {
			b.WriteString(m.errFilings.Error())
		} else {
			b.WriteString(muted.Render("No filing selected"))
		}
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}

	b.WriteString(label.Render(item.Form) + "\n")
	if snapshot.CompanyName != "" {
		b.WriteString(muted.Render(truncateText(snapshot.CompanyName, width)) + "\n")
	}
	if snapshot.CIK != "" {
		b.WriteString(muted.Render("CIK "+snapshot.CIK) + "\n")
	}
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("%s %s\n", label.Render("Filed"), filingDateLabel(item)))
	if !item.ReportDate.IsZero() {
		b.WriteString(fmt.Sprintf("%s %s\n", label.Render("Period"), item.ReportDate.Format("2006-01-02")))
	}
	if !item.AcceptedAt.IsZero() {
		b.WriteString(fmt.Sprintf("%s %s\n", label.Render("Accepted"), item.AcceptedAt.Format("2006-01-02 15:04:05")))
	}
	if item.PrimaryDocDescription != "" {
		b.WriteString("\n" + muted.Render("Document") + "\n")
		b.WriteString(renderWrappedTextBlock(muted, item.PrimaryDocDescription, width))
	}
	if item.PrimaryDocument != "" {
		b.WriteString("\n\n" + muted.Render("File") + "\n")
		b.WriteString(renderWrappedTextBlock(muted, item.PrimaryDocument, width))
	}
	flags := make([]string, 0, 2)
	if item.IsInlineXBRL {
		flags = append(flags, "inline xbrl")
	}
	if item.IsXBRL {
		flags = append(flags, "xbrl")
	}
	if len(flags) > 0 {
		b.WriteString("\n\n" + muted.Render("Flags") + "\n")
		b.WriteString(renderWrappedTextBlock(muted, strings.Join(flags, " • "), width))
	}
	b.WriteString("\n\n" + muted.Render("n/p move • o open filing"))
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}
