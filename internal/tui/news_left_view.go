package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderNewsLeft(section, muted lipgloss.Style, width, height int) string {
	var b strings.Builder
	label := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#D8C9B8"))
	liveLabel := "STALE"
	if !m.marketNewsUpdated.IsZero() && time.Since(m.marketNewsUpdated) <= 45*time.Second {
		liveLabel = "LIVE"
	}

	b.WriteString(section.Render("DESK") + "\n\n")
	b.WriteString(fmt.Sprintf("%s %s\n", label.Render("Status"), liveLabel))
	b.WriteString(fmt.Sprintf("%s %d\n", label.Render("Stories"), len(m.marketNews)))
	if !m.marketNewsUpdated.IsZero() {
		b.WriteString(fmt.Sprintf("%s %s\n", label.Render("Updated"), m.marketNewsUpdated.Local().Format("15:04:05")))
	}
	b.WriteString("\n" + section.Render("MARKET SNAP") + "\n\n")
	for _, item := range []marketBoardItem{
		{label: "S&P 500", symbol: "SPY"},
		{label: "Nasdaq 100", symbol: "QQQ"},
		{label: "VIX", symbol: "^VIX"},
		{label: "10Y", symbol: "^TNX"},
		{label: "Gold", symbol: "GC=F"},
		{label: "Oil", symbol: "CL=F"},
	} {
		b.WriteString(renderCompactMarketLine(m, label, muted, item) + "\n")
	}

	usedLines := lipgloss.Height(strings.TrimRight(b.String(), "\n"))
	sourceRows := max(1, height-usedLines-3)
	b.WriteString("\n" + section.Render("SOURCE MIX") + "\n\n")
	sourceMix := marketNewsSourceRows(m.marketNewsSources, m.marketNews)
	sourceLines := make([]string, 0, len(sourceMix))
	for _, row := range sourceMix {
		sourceLines = append(sourceLines, fmt.Sprintf("%s %s", label.Render(compactSourceName(row.name)), row.price))
	}
	sourceCols := 1
	if len(sourceLines) > sourceRows {
		sourceCols = min(max(2, (len(sourceLines)+sourceRows-1)/sourceRows), max(1, width/14))
	}
	if sourceCols <= 1 {
		b.WriteString(strings.Join(sourceLines[:min(len(sourceLines), sourceRows)], "\n"))
	} else {
		b.WriteString(renderCompactStringColumns(sourceLines, width, sourceCols))
	}
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}
