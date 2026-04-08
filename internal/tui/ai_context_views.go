package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderAIRight(section, label, muted lipgloss.Style, width, height int) string {
	return clipLines(m.renderAIContextBlock(section, label, muted, width), height)
}

func (m Model) renderAIContextBlock(section, label, muted lipgloss.Style, width int) string {
	var b strings.Builder
	b.WriteString(section.Render("CONTEXT") + "\n\n")
	statements := m.statementBundleForSymbol(m.activeSymbol())
	statementInsights := aiStatementInsights(statements)
	statementSummary := fmt.Sprintf("%d/6 loaded", statements.loadedCount())
	for _, line := range []string{
		fmt.Sprintf("Symbol %s", m.activeSymbol()),
		fmt.Sprintf("Provider %s", m.statusMetaMarketSource()),
		fmt.Sprintf("Market %s", m.statusMetaMarketSource()),
		fmt.Sprintf("AI %s", m.activeAIConnectorLabel()),
		fmt.Sprintf("Model %s", valueOrDash(strings.TrimSpace(m.config.AIModel))),
		fmt.Sprintf("Watchlist %d", len(m.config.Watchlist)),
		fmt.Sprintf("News %d", len(m.news)),
		fmt.Sprintf("Candles %d", len(m.series.Candles)),
		fmt.Sprintf("Statements %s", statementSummary),
		fmt.Sprintf("Stmt insights %d", len(statementInsights)),
		fmt.Sprintf("Range %s", ranges[m.rangeIdx].Label),
	} {
		b.WriteString(renderWrappedLabelLine(label, line, width) + "\n")
	}
	b.WriteString("\n" + section.Render("FLOW") + "\n\n")
	b.WriteString(renderWrappedTextBlock(muted, "Blackdesk only resends normalized context when the selected symbol or app snapshot changes. Otherwise it sends the transcript only.", width))
	if m.aiDuration > 0 {
		b.WriteString("\n\n" + muted.Render(truncateText("Last run: "+m.aiDuration.Round(time.Millisecond).String(), width)))
	}
	return strings.TrimRight(b.String(), "\n")
}

func renderWrappedTextBlock(style lipgloss.Style, text string, width int) string {
	return strings.TrimRight(style.Width(max(1, width)).Render(text), "\n")
}

func renderWrappedLabelLine(label lipgloss.Style, text string, width int) string {
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return truncateText(parts[0], width)
	}
	return renderWrappedTextBlock(
		lipgloss.NewStyle(),
		label.Render(parts[0])+" "+strings.Join(parts[1:], " "),
		width,
	)
}
