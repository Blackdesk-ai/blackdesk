package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderOverviewRight(section, label, muted lipgloss.Style, width, height int) string {
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
