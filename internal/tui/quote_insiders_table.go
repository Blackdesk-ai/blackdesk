package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"blackdesk/internal/domain"
	"blackdesk/internal/ui"
)

type insiderTableColumn struct {
	title string
	width int
	align lipgloss.Position
	style lipgloss.Style
}

func renderInsiderTransactionHeader(width int, items []domain.InsiderTransaction, label, muted lipgloss.Style) string {
	columns := insiderTransactionColumns(width, items, label, muted)
	return renderInsiderTableHeader(columns, muted)
}

func renderInsiderTransactionRow(item domain.InsiderTransaction, width int, items []domain.InsiderTransaction, label, muted lipgloss.Style) string {
	showRole := insidersAnyTransactionRole(items)
	showOwn := insidersShowOwnership(items)
	showValue := insidersAnyTransactionValue(items)
	columns := insiderTransactionColumns(width, items, label, muted)
	values := make([]string, 0, len(columns))
	dateText := "-"
	if !item.Date.IsZero() {
		dateText = item.Date.Format("2006-01-02")
	}
	values = append(values, dateText, valueOrDash(strings.TrimSpace(item.Insider)))
	if showRole {
		values = append(values, valueOrDash(strings.TrimSpace(item.Relation)))
	}
	if showOwn {
		values = append(values, valueOrDash(strings.TrimSpace(item.Ownership)))
	}
	sharesText := "-"
	if item.Shares != 0 {
		sharesText = ui.FormatCompactInt(item.Shares)
		sharesText = colorizeInsiderTransactionMetric(sharesText, item)
	}
	values = append(values, sharesText)
	if showValue {
		valueText := "-"
		if item.Value != 0 {
			valueText = ui.FormatCompactInt(item.Value)
			valueText = colorizeInsiderTransactionMetric(valueText, item)
		}
		values = append(values, valueText)
	}
	return renderInsiderTableRow(columns, values)
}

func renderInsiderTableHeader(columns []insiderTableColumn, muted lipgloss.Style) string {
	if len(columns) == 0 {
		return ""
	}
	parts := make([]string, 0, len(columns))
	for _, col := range columns {
		parts = append(parts, lipgloss.NewStyle().Width(col.width).Align(col.align).Render(muted.Render(col.title)))
	}
	return strings.Join(parts, "  ")
}

func renderInsiderTableRow(columns []insiderTableColumn, values []string) string {
	if len(columns) == 0 || len(values) == 0 {
		return ""
	}
	parts := make([]string, 0, len(columns))
	for i, col := range columns {
		value := ""
		if i < len(values) {
			value = values[i]
		}
		parts = append(parts, col.style.Width(col.width).Align(col.align).Render(ansi.Truncate(value, col.width, "")))
	}
	return strings.Join(parts, "  ")
}

func insiderTransactionColumns(width int, items []domain.InsiderTransaction, label, muted lipgloss.Style) []insiderTableColumn {
	showRole := insidersAnyTransactionRole(items)
	showOwn := insidersShowOwnership(items)
	showValue := insidersAnyTransactionValue(items)
	dateWidth := 10
	sharesWidth := 8
	valueWidth := 8
	roleWidth := clamp(width/6, 16, 26)
	ownWidth := 8
	fixed := dateWidth + sharesWidth
	gaps := 2
	if showRole {
		fixed += roleWidth
	}
	if showOwn {
		fixed += ownWidth
	}
	if showValue {
		fixed += valueWidth
	}
	available := max(24, width-fixed-gaps*(2+boolToInt(showRole)+boolToInt(showOwn)+boolToInt(showValue)))
	insiderWidth := min(available, clamp(width/6, 18, 30))
	if showRole {
		roleWidth += max(0, available-insiderWidth)
	} else {
		insiderWidth = available
	}
	cols := []insiderTableColumn{
		{title: "Date", width: dateWidth, align: lipgloss.Left, style: muted},
		{title: "Insider", width: insiderWidth, align: lipgloss.Left, style: label},
	}
	if showRole {
		cols = append(cols, insiderTableColumn{title: "Role", width: roleWidth, align: lipgloss.Left, style: muted})
	}
	if showOwn {
		cols = append(cols, insiderTableColumn{title: "Own", width: ownWidth, align: lipgloss.Left, style: muted})
	}
	cols = append(cols, insiderTableColumn{title: "Shares", width: sharesWidth, align: lipgloss.Right, style: lipgloss.NewStyle()})
	if showValue {
		cols = append(cols, insiderTableColumn{title: "Value", width: valueWidth, align: lipgloss.Right, style: lipgloss.NewStyle()})
	}
	return cols
}

func insidersAnyTransactionRole(items []domain.InsiderTransaction) bool {
	for _, item := range items {
		if strings.TrimSpace(item.Relation) != "" {
			return true
		}
	}
	return false
}

func insidersShowOwnership(items []domain.InsiderTransaction) bool {
	seen := make(map[string]struct{})
	for _, item := range items {
		ownership := strings.TrimSpace(item.Ownership)
		if ownership == "" {
			continue
		}
		seen[strings.ToLower(ownership)] = struct{}{}
	}
	if len(seen) == 0 {
		return false
	}
	if len(seen) > 1 {
		return true
	}
	_, onlyDirect := seen["direct"]
	return !onlyDirect
}

func insidersAnyTransactionValue(items []domain.InsiderTransaction) bool {
	for _, item := range items {
		if item.Value != 0 {
			return true
		}
	}
	return false
}

func colorizeInsiderTransactionMetric(text string, item domain.InsiderTransaction) string {
	action := strings.ToLower(strings.TrimSpace(item.Action))
	if action == "" {
		action = strings.ToLower(strings.TrimSpace(item.Text))
	}
	switch {
	case strings.Contains(action, "buy"), strings.Contains(action, "purchase"):
		return marketMoveStyle(1).Render(text)
	case strings.Contains(action, "sale"), strings.Contains(action, "sell"):
		return marketMoveStyle(-1).Render(text)
	default:
		return text
	}
}
