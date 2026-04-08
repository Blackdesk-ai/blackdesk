package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func renderMarketTableHeader(width int, label lipgloss.Style) string {
	return renderMarketTableHeaderWithValueLabel(width, "Price", label)
}

func renderMarketTableHeaderWithValueLabel(width int, valueLabel string, label lipgloss.Style) string {
	return renderMarketTableHeaderWithLabels(width, "Name", valueLabel, "Chg", label)
}

func renderMarketTableHeaderWithLabels(width int, nameLabel, valueLabel, signalLabel string, label lipgloss.Style) string {
	labelWidth := clamp(width/2, 8, 14)
	return fmt.Sprintf("%-*s %-8s %s", labelWidth, nameLabel, valueLabel, signalLabel)
}

func renderMarketTableRow(row marketTableRow, width int, label lipgloss.Style) string {
	labelWidth := clamp(width/2, 8, 14)
	signalWidth := max(6, width-labelWidth-10)
	signal := ansi.Truncate(colorizeMarketChange(row.chg, row.move, row.styled), signalWidth, "")
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(labelWidth).Render(label.Render(row.name)),
		" ",
		lipgloss.NewStyle().Width(8).Render(colorizeMarketPrice(row.price, row.move, row.styled)),
		" ",
		lipgloss.NewStyle().Width(signalWidth).Render(signal),
	)
}

func renderFundamentalsTableRow(row marketTableRow, width int) string {
	labelWidth := clamp(width/2, 8, 16)
	valueWidth := max(8, width-labelWidth-1)
	value := colorizeMarketPrice(row.price, row.move, row.styled)
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(labelWidth).Render(row.name),
		" ",
		lipgloss.NewStyle().Width(valueWidth).Render(value),
	)
}
