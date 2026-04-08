package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"blackdesk/internal/ui"
)

func renderMarketBoard(m Model, items []marketBoardItem, width int, label lipgloss.Style) string {
	rows := make([]string, 0, len(items))
	labelWidth := clamp(width/3, 9, 14)
	priceWidth := 10
	for _, item := range items {
		quote, ok := m.lookupQuote(item.symbol)
		priceText := "--"
		changeText := "--"
		move := 0.0
		if ok {
			displayPrice, displayChangePercent := marketDisplayQuoteLine(quote)
			priceText = ui.FormatMoney(displayPrice)
			changeText = fmt.Sprintf("%+.2f%%", displayChangePercent)
			move = displayChangePercent
		}
		rows = append(rows, lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.NewStyle().Width(labelWidth).Render(label.Render(item.label)),
			" ",
			lipgloss.NewStyle().Width(priceWidth).Render(colorizeMarketPrice(priceText, move, ok)),
			" ",
			colorizeMarketChange(changeText, move, ok),
		))
	}
	return strings.Join(rows, "\n")
}

func renderMarketBoardHeader(width int, valueLabel string, label lipgloss.Style) string {
	if strings.TrimSpace(valueLabel) == "" {
		valueLabel = "Level"
	}
	labelWidth := clamp(width/3, 9, 14)
	priceWidth := 10
	return fmt.Sprintf("%-*s %-*s %s", labelWidth, "Name", priceWidth, valueLabel, "Chg")
}

func marketBoardContentWidth(width int) int {
	labelWidth := clamp(width/3, 9, 14)
	priceWidth := 10
	changeWidth := len("Chg")
	return labelWidth + 1 + priceWidth + 1 + changeWidth
}

func renderMarketBoardColumns(m Model, items []marketBoardItem, width, cols int) string {
	rows := make([]string, 0, len(items))
	for _, item := range items {
		quote, ok := m.lookupQuote(item.symbol)
		priceText := "--"
		changeText := "--"
		move := 0.0
		if ok {
			displayPrice, displayChangePercent := marketDisplayQuoteLine(quote)
			priceText = ui.FormatMoney(displayPrice)
			changeText = fmt.Sprintf("%+.2f%%", displayChangePercent)
			move = displayChangePercent
		}
		rows = append(rows, lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.NewStyle().Width(13).Render(item.label),
			" ",
			lipgloss.NewStyle().Width(8).Render(colorizeMarketPrice(priceText, move, ok)),
			" ",
			colorizeMarketChange(changeText, move, ok),
		))
	}
	return renderStringColumns(rows, width, cols)
}

func renderCompactMarketLine(m Model, label, muted lipgloss.Style, item marketBoardItem) string {
	quote, ok := m.lookupQuote(item.symbol)
	priceText := "--"
	changeText := "--"
	move := 0.0
	if ok {
		displayPrice, displayChangePercent := marketDisplayQuoteLine(quote)
		priceText = ui.FormatMoney(displayPrice)
		changeText = fmt.Sprintf("%+.2f%%", displayChangePercent)
		move = displayChangePercent
	}
	return fmt.Sprintf("%s %s  %s", label.Render(item.label), colorizeMarketPrice(priceText, move, ok), colorizeMarketChange(changeText, move, ok))
}
