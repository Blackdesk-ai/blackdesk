package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"blackdesk/internal/domain"
	"blackdesk/internal/ui"
)

func renderScreenerHeader(width int, label lipgloss.Style, metricLabel string) string {
	prefixWidth, colWidths := screenerColumnWidths(width)
	if strings.TrimSpace(metricLabel) == "" {
		metricLabel = "Metric"
	}
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(prefixWidth).Render(""),
		" ",
		lipgloss.NewStyle().Width(colWidths[0]).Render(label.Render(ansi.Truncate("Symbol", colWidths[0], "..."))),
		" ",
		lipgloss.NewStyle().Width(colWidths[1]).Render(label.Render(ansi.Truncate("Price", colWidths[1], "..."))),
		" ",
		lipgloss.NewStyle().Width(colWidths[2]).Render(label.Render(ansi.Truncate("Chg", colWidths[2], "..."))),
		" ",
		lipgloss.NewStyle().Width(colWidths[3]).Render(label.Render(ansi.Truncate("Volume", colWidths[3], "..."))),
		" ",
		lipgloss.NewStyle().Width(colWidths[4]).Render(label.Render(ansi.Truncate(metricLabel, colWidths[4], "..."))),
		" ",
		lipgloss.NewStyle().Width(colWidths[5]).Render(label.Render(ansi.Truncate("RV", colWidths[5], "..."))),
	)
}

func renderScreenerRow(item domain.ScreenerItem, selected bool, width int, label lipgloss.Style, metricKey string) string {
	prefixWidth, colWidths := screenerColumnWidths(width)

	prefix := "  "
	if selected {
		prefix = "▶ "
	}
	symbol := padRight(ansi.Truncate(item.Symbol, colWidths[0], ""), colWidths[0])
	priceText := ui.FormatMoney(item.Price)
	changeText := fmt.Sprintf("%+.2f%%", item.ChangePercent)
	volumeText := ui.FormatCompactInt(item.Volume)
	rvText := screenerMetricValue(item, "relative_volume")
	if rvText == "" {
		rvText = "-"
	}
	focusText := screenerMetricValue(item, metricKey)
	if focusText == "" {
		focusText = "-"
	}

	row := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(prefixWidth).Render(prefix),
		" ",
		lipgloss.NewStyle().Width(colWidths[0]).Render(label.Render(symbol)),
		" ",
		lipgloss.NewStyle().Width(colWidths[1]).Render(ansi.Truncate(colorizeMarketPrice(priceText, item.ChangePercent, true), colWidths[1], "...")),
		" ",
		lipgloss.NewStyle().Width(colWidths[2]).Render(ansi.Truncate(colorizeMarketChange(changeText, item.ChangePercent, true), colWidths[2], "...")),
		" ",
		lipgloss.NewStyle().Width(colWidths[3]).Render(ansi.Truncate(volumeText, colWidths[3], "...")),
		" ",
		lipgloss.NewStyle().Width(colWidths[4]).Render(ansi.Truncate(focusText, colWidths[4], "...")),
		" ",
		lipgloss.NewStyle().Width(colWidths[5]).Render(ansi.Truncate(colorizeRelativeVolume(rvText, item, true), colWidths[5], "...")),
	)
	if selected {
		return lipgloss.NewStyle().Bold(true).Render(row)
	}
	return row
}

func screenerColumnWidths(width int) (int, [6]int) {
	const (
		prefixWidth = 2
		gapCount    = 6
		colCount    = 6
	)

	available := max(colCount, width-prefixWidth-gapCount)
	base := max(1, available/colCount)
	remainder := max(0, available-base*colCount)

	widths := [6]int{base, base, base, base, base, base}
	for i := 0; i < remainder && i < len(widths); i++ {
		widths[i]++
	}
	return prefixWidth, widths
}

func colorizeRelativeVolume(text string, item domain.ScreenerItem, styled bool) string {
	if !styled || text == "-" || text == "--" || item.AverageVolume <= 0 {
		return text
	}
	move := float64(item.Volume)/float64(item.AverageVolume) - 1
	return marketMoveStyle(move).Render(text)
}

func screenerPrimaryMetric(result domain.ScreenerResult) (string, string) {
	for _, item := range result.Items {
		for _, metric := range item.Metrics {
			if metric.Key == "relative_volume" {
				continue
			}
			if strings.TrimSpace(metric.Value) == "" {
				continue
			}
			return metric.Key, metric.Label
		}
	}
	for _, item := range result.Items {
		if value := screenerMetricValue(item, "relative_volume"); strings.TrimSpace(value) != "" {
			return "relative_volume", "RV"
		}
	}
	return "", "Metric"
}

func screenerMetricValue(item domain.ScreenerItem, key string) string {
	if strings.TrimSpace(key) != "" {
		for _, metric := range item.Metrics {
			if metric.Key == key {
				return strings.TrimSpace(metric.Value)
			}
		}
		return ""
	}
	for _, metric := range item.Metrics {
		if metric.Key == "relative_volume" {
			continue
		}
		if strings.TrimSpace(metric.Value) == "" {
			continue
		}
		return metric.Value
	}
	return strings.TrimSpace(screenerMetricValue(item, "relative_volume"))
}
