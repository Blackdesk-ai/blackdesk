package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"blackdesk/internal/domain"
)

func (m Model) renderQuoteAnalystRecommendationsPage(header, section, label, muted, pos, neg lipgloss.Style, width, height int) string {
	snapshot := m.analystRecommendationsForSymbol(m.activeSymbol())
	listWidth := clamp((width*3)/5, 48, 82)
	previewWidth := max(24, width-listWidth-3)
	bodyHeight := max(8, height-6)

	var b strings.Builder
	b.WriteString(section.Render("ANALYST RECOMMENDATIONS") + "\n\n")
	b.WriteString(m.renderQuoteAnalystSummary(header, muted, width, snapshot) + "\n\n")
	left := lipgloss.NewStyle().Width(listWidth).Render(m.renderQuoteAnalystList(section, muted, pos, neg, listWidth, bodyHeight, snapshot))
	right := lipgloss.NewStyle().Width(previewWidth).Render(m.renderQuoteAnalystPreview(section, label, muted, pos, neg, previewWidth, bodyHeight, snapshot))
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, left, "   ", right))
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func (m Model) renderQuoteAnalystSummary(header, muted lipgloss.Style, width int, snapshot domain.AnalystRecommendationsSnapshot) string {
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

func (m Model) renderQuoteAnalystList(section, muted, pos, neg lipgloss.Style, width, height int, snapshot domain.AnalystRecommendationsSnapshot) string {
	var b strings.Builder
	b.WriteString(section.Render("LATEST UPDATES") + "\n\n")
	items := snapshot.Items
	if len(items) == 0 {
		if m.errAnalyst != nil {
			b.WriteString(m.errAnalyst.Error())
		} else {
			b.WriteString(muted.Render("No analyst recommendation history loaded for the active symbol"))
		}
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#14110D")).
		Background(lipgloss.Color("#E7B66B")).
		Bold(true).
		Width(width).
		MaxWidth(width)

	dateColWidth := 10
	firmColWidth := max(12, width/3)
	actionColWidth := 12
	detailColWidth := max(12, width-dateColWidth-firmColWidth-actionColWidth-6)

	headerLine := fmt.Sprintf(
		"%-*s  %-*s  %-*s  %s",
		dateColWidth, "DATE",
		firmColWidth, "FIRM",
		actionColWidth, "ACTION",
		truncateText("RECOMMENDATION", detailColWidth),
	)
	b.WriteString(muted.Render(truncateText(headerLine, width)) + "\n")
	b.WriteString(muted.Render(strings.Repeat("─", max(12, min(width, dateColWidth+firmColWidth+actionColWidth+detailColWidth+6)))) + "\n")

	visibleRows := max(3, height/2)
	start := 0
	if m.analystSel >= visibleRows {
		start = m.analystSel - visibleRows + 1
	}
	end := min(len(items), start+visibleRows)
	for i := start; i < end; i++ {
		item := items[i]
		line := fmt.Sprintf(
			"%-*s  %-*s  %-*s  %s",
			dateColWidth, truncateText(analystRecommendationDateLabel(item), dateColWidth),
			firmColWidth, truncateText(item.Firm, firmColWidth),
			actionColWidth, truncateText(strings.ToUpper(analystRecommendationActionLabel(item)), actionColWidth),
			truncateText(analystRecommendationDetailLine(item), detailColWidth),
		)
		line = truncateText(line, width)
		if i == m.analystSel {
			b.WriteString(selectedStyle.Render(line) + "\n")
			continue
		}
		b.WriteString(renderAnalystRecommendationRow(line, item, pos, neg) + "\n")
	}
	b.WriteString("\n" + muted.Render("↑/↓ move • palette to switch views"))
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func (m Model) renderQuoteAnalystPreview(section, label, muted, pos, neg lipgloss.Style, width, height int, snapshot domain.AnalystRecommendationsSnapshot) string {
	var b strings.Builder
	b.WriteString(section.Render("PREVIEW") + "\n\n")

	b.WriteString(label.Render("Consensus") + "\n")
	consensusLine := colorizeRecommendationBadge(snapshot.RecommendationKey)
	if consensusLine == "" {
		consensusLine = recommendationBadge(snapshot.RecommendationKey)
	}
	if consensusLine == "" {
		consensusLine = "-"
	}
	b.WriteString(renderWrappedTextBlock(muted, fmt.Sprintf("%s  •  Mean %s  •  %s analysts", consensusLine, formatAnalystMean(snapshot.RecommendationMean), formatAnalystOpinions(snapshot.AnalystOpinions)), width))
	b.WriteString("\n\n" + muted.Render("Targets") + "\n")
	b.WriteString(renderWrappedTextBlock(muted, fmt.Sprintf("Low %s  •  Mean %s  •  High %s", formatMoneyDash(snapshot.TargetLowPrice), formatMoneyDash(snapshot.TargetMeanPrice), formatMoneyDash(snapshot.TargetHighPrice)), width))

	if len(snapshot.Trends) > 0 {
		b.WriteString("\n\n" + muted.Render("Trend") + "\n")
		b.WriteString(renderWrappedTextBlock(muted, analystTrendInterpretation(snapshot), width) + "\n")
		for _, trend := range snapshot.Trends {
			line := fmt.Sprintf("%s: Strong Buy %d, Buy %d, Hold %d, Sell %d, Strong Sell %d", analystTrendPeriodLabel(trend.Period), trend.StrongBuy, trend.Buy, trend.Hold, trend.Sell, trend.StrongSell)
			b.WriteString(renderWrappedTextBlock(muted, line, width) + "\n")
		}
	}

	item, ok := m.currentAnalystRecommendation()
	if !ok {
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}

	b.WriteString("\n" + label.Render("Selected update") + "\n")
	b.WriteString(muted.Render(analystRecommendationDateTimeLabel(item)) + "\n")
	if strings.TrimSpace(item.Firm) != "" {
		b.WriteString("\n" + muted.Render("Firm") + "\n")
		b.WriteString(renderWrappedTextBlock(lipgloss.NewStyle(), item.Firm, width) + "\n")
	}
	b.WriteString("\n" + muted.Render("Action") + "\n")
	b.WriteString(renderWrappedTextBlock(lipgloss.NewStyle(), analystRecommendationActionStyle(item, pos, neg).Render(strings.ToUpper(analystRecommendationActionLabel(item))), width))
	if detail := analystRecommendationDetailLine(item); detail != "-" {
		b.WriteString("\n\n" + muted.Render("Recommendation") + "\n")
		b.WriteString(renderWrappedTextBlock(muted, detail, width))
	}
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func analystRecommendationDateLabel(item domain.AnalystRecommendationItem) string {
	if item.Date.IsZero() {
		return "-"
	}
	return item.Date.Format("2006-01-02")
}

func analystRecommendationDateTimeLabel(item domain.AnalystRecommendationItem) string {
	if item.Date.IsZero() {
		return "Update time unavailable"
	}
	return "Updated " + item.Date.Format("2006-01-02 15:04")
}

func analystRecommendationActionLabel(item domain.AnalystRecommendationItem) string {
	action := strings.TrimSpace(item.Action)
	if action == "" {
		if item.ToGrade != "" && item.FromGrade != "" && !strings.EqualFold(item.ToGrade, item.FromGrade) {
			return "revised"
		}
		if item.ToGrade != "" {
			return "initiated"
		}
		return "updated"
	}
	return strings.ReplaceAll(action, "_", " ")
}

func analystRecommendationActionStyle(item domain.AnalystRecommendationItem, pos, neg lipgloss.Style) lipgloss.Style {
	action := strings.ToLower(strings.TrimSpace(item.Action))
	switch action {
	case "up", "upgrade", "maintains":
		return pos
	case "down", "downgrade":
		return neg
	default:
		return lipgloss.NewStyle()
	}
}

func analystRecommendationDetailLine(item domain.AnalystRecommendationItem) string {
	to := strings.TrimSpace(item.ToGrade)
	from := strings.TrimSpace(item.FromGrade)
	switch {
	case to != "" && from != "" && !strings.EqualFold(to, from):
		return from + " -> " + to
	case to != "":
		return to
	case from != "":
		return from
	default:
		return "-"
	}
}

func renderAnalystRecommendationRow(line string, item domain.AnalystRecommendationItem, pos, neg lipgloss.Style) string {
	action := strings.ToUpper(analystRecommendationActionLabel(item))
	switch {
	case strings.Contains(action, "UP") || strings.Contains(action, "MAINTAIN"):
		return strings.Replace(line, action, pos.Render(action), 1)
	case strings.Contains(action, "DOWN"):
		return strings.Replace(line, action, neg.Render(action), 1)
	default:
		return line
	}
}

func analystTrendPeriodLabel(period string) string {
	switch strings.TrimSpace(period) {
	case "0m":
		return "Now"
	case "-1m":
		return "1m ago"
	case "-2m":
		return "2m ago"
	case "-3m":
		return "3m ago"
	default:
		return period
	}
}

func analystTrendInterpretation(snapshot domain.AnalystRecommendationsSnapshot) string {
	if len(snapshot.Trends) == 0 {
		return ""
	}
	current := snapshot.Trends[0]
	bullish := current.StrongBuy + current.Buy
	bearish := current.Sell + current.StrongSell

	tone := "balanced"
	switch {
	case bullish > bearish+current.Hold/2:
		tone = "positive"
	case bearish > bullish+current.Hold/2:
		tone = "negative"
	case current.Hold >= bullish && current.Hold >= bearish:
		tone = "cautious"
	}

	changeText := "The monthly mix is broadly unchanged."
	if len(snapshot.Trends) > 1 {
		prev := snapshot.Trends[1]
		deltaBullish := bullish - (prev.StrongBuy + prev.Buy)
		deltaBearish := bearish - (prev.Sell + prev.StrongSell)
		switch {
		case deltaBullish > 0 && deltaBearish <= 0:
			changeText = "Bullish ratings improved versus last month."
		case deltaBearish > 0 && deltaBullish <= 0:
			changeText = "Bearish ratings increased versus last month."
		case deltaBullish == 0 && deltaBearish == 0:
			changeText = "The monthly mix is unchanged versus last month."
		default:
			changeText = "The monthly mix shifted only slightly versus last month."
		}
	}

	return fmt.Sprintf("Distribution looks %s: %d bullish, %d hold, %d bearish. %s", tone, bullish, current.Hold, bearish, changeText)
}
