package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func renderStringColumns(rows []string, width, cols int) string {
	if len(rows) == 0 {
		return ""
	}
	cols = max(1, min(cols, len(rows)))
	if cols == 1 {
		return strings.Join(rows, "\n")
	}

	gap := 2
	colWidth := max(18, (width-gap*(cols-1))/cols)
	rowsPerCol := (len(rows) + cols - 1) / cols
	rendered := make([]string, 0, cols)
	for col := 0; col < cols; col++ {
		start := col * rowsPerCol
		if start >= len(rows) {
			break
		}
		end := min(len(rows), start+rowsPerCol)
		rendered = append(rendered, lipgloss.NewStyle().Width(colWidth).MaxWidth(colWidth).Render(strings.Join(rows[start:end], "\n")))
	}
	parts := make([]string, 0, len(rendered)*2-1)
	for i, col := range rendered {
		if i > 0 {
			parts = append(parts, strings.Repeat(" ", gap))
		}
		parts = append(parts, col)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func renderCompactStringColumns(rows []string, width, cols int) string {
	if len(rows) == 0 {
		return ""
	}
	cols = max(1, min(cols, len(rows)))
	if cols == 1 {
		return strings.Join(rows, "\n")
	}

	gap := 2
	colWidth := max(10, (width-gap*(cols-1))/cols)
	rowsPerCol := (len(rows) + cols - 1) / cols
	rendered := make([]string, 0, cols)
	for col := 0; col < cols; col++ {
		start := col * rowsPerCol
		if start >= len(rows) {
			break
		}
		end := min(len(rows), start+rowsPerCol)
		rendered = append(rendered, lipgloss.NewStyle().Width(colWidth).MaxWidth(colWidth).Render(strings.Join(rows[start:end], "\n")))
	}
	parts := make([]string, 0, len(rendered)*2-1)
	for i, col := range rendered {
		if i > 0 {
			parts = append(parts, strings.Repeat(" ", gap))
		}
		parts = append(parts, col)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}
