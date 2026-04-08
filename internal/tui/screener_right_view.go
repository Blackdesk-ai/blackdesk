package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"blackdesk/internal/ui"
)

func (m Model) renderScreenerRight(section, label, muted lipgloss.Style, width, height int) string {
	var b strings.Builder
	item, ok := m.currentScreenerItem()
	result := m.screenerResult
	if strings.TrimSpace(result.Definition.ID) == "" {
		result.Definition = m.currentScreenerDefinition()
	}

	b.WriteString(section.Render("DETAIL") + "\n\n")
	if !ok {
		b.WriteString(renderWrappedTextBlock(muted, "Select a loaded screener to inspect names and filters.", width))
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}

	b.WriteString(ansi.Truncate(item.Symbol, width, "") + "\n")
	if strings.TrimSpace(item.Name) != "" {
		b.WriteString(renderWrappedTextBlock(muted, item.Name, width) + "\n")
	}
	meta := []string{}
	if item.Exchange != "" {
		meta = append(meta, item.Exchange)
	}
	if item.Type != "" {
		meta = append(meta, item.Type)
	}
	if len(meta) > 0 {
		b.WriteString(muted.Render(strings.Join(meta, " • ")) + "\n")
	}
	b.WriteString("\n")
	for _, line := range []string{
		fmt.Sprintf("Price %s", ui.FormatMoney(item.Price)),
		fmt.Sprintf("Move %s", fmt.Sprintf("%+.2f%%", item.ChangePercent)),
		fmt.Sprintf("Volume %s", valueOrDash(ui.FormatCompactInt(item.Volume))),
		fmt.Sprintf("Avg 3M %s", valueOrDash(ui.FormatCompactInt(item.AverageVolume))),
		fmt.Sprintf("Mkt Cap %s", valueOrDash(ui.FormatCompactInt(item.MarketCap))),
	} {
		b.WriteString(renderWrappedLabelLine(label, line, width) + "\n")
	}
	rvValue := valueOrDash(screenerMetricValue(item, "relative_volume"))
	b.WriteString(renderWrappedLabelLine(label, "RV "+colorizeRelativeVolume(rvValue, item, true), width) + "\n")

	if len(item.Metrics) > 0 {
		b.WriteString("\n" + section.Render("SIGNALS") + "\n\n")
		for _, metric := range item.Metrics {
			line := metric.Label + " " + metric.Value
			if metric.Signal != "" {
				line += " (" + metric.Signal + ")"
			}
			b.WriteString(renderWrappedLabelLine(label, line, width) + "\n")
		}
	}

	if len(result.Criteria) > 0 {
		b.WriteString("\n" + section.Render("RULES") + "\n\n")
		for _, criterion := range result.Criteria {
			b.WriteString(renderWrappedTextBlock(muted, "• "+criterion.Statement, width) + "\n")
		}
	}

	b.WriteString("\n" + section.Render("ACTIONS") + "\n\n")
	for _, line := range []string{
		"Enter opens the selected symbol in Quote",
		"a adds the selected symbol to the watchlist",
		"n / p or ← / → changes the active screener",
		"r refreshes the current screen",
	} {
		b.WriteString(renderWrappedTextBlock(muted, line, width) + "\n")
	}
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}
