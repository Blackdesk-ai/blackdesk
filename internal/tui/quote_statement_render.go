package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"blackdesk/internal/domain"
	"blackdesk/internal/ui"
)

const (
	statementColumnGap          = " "
	statementNameColumnGap      = "  "
	statementNameColumnGapExtra = len(statementNameColumnGap) - len(statementColumnGap)
)

func renderQuoteStatementsBoard(section, label, muted lipgloss.Style, stmt domain.FinancialStatement, activeKind domain.StatementKind, activeFreq domain.StatementFrequency, width, height int) string {
	var b strings.Builder
	b.WriteString(renderStatementHeaderBlock(activeKind, activeFreq, muted, width) + "\n\n")

	if len(stmt.Rows) == 0 || len(stmt.Periods) == 0 {
		b.WriteString(muted.Render("Statement data unavailable"))
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}

	dates := min(4, len(stmt.Periods))
	nameWidth := clamp(width/3, 14, 28)
	colWidth := max(10, (width-nameWidth-dates-statementNameColumnGapExtra)/dates)
	b.WriteString(renderStatementHeader(stmt.Periods[:dates], nameWidth, colWidth, muted) + "\n")
	for _, row := range stmt.Rows {
		b.WriteString(renderStatementRow(stmt, row, dates, nameWidth, colWidth, label) + "\n")
	}
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func renderStatementHeader(periods []domain.StatementPeriod, nameWidth, colWidth int, muted lipgloss.Style) string {
	parts := []string{lipgloss.NewStyle().Width(nameWidth).Render(muted.Render("Name"))}
	for _, period := range periods {
		parts = append(parts, lipgloss.NewStyle().Width(colWidth).Align(lipgloss.Right).Render(muted.Render(period.Label)))
	}
	return joinStatementColumns(parts)
}

func renderStatementRow(stmt domain.FinancialStatement, row domain.StatementRow, dates, nameWidth, colWidth int, label lipgloss.Style) string {
	parts := []string{lipgloss.NewStyle().Width(nameWidth).Render(label.Render(row.Label))}
	for i := 0; i < dates; i++ {
		value := "-"
		if i < len(row.Values) && row.Values[i].Present {
			value = formatStatementCell(stmt, row, i)
		}
		parts = append(parts, lipgloss.NewStyle().Width(colWidth).Align(lipgloss.Right).Render(value))
	}
	return joinStatementColumns(parts)
}

func joinStatementColumns(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return parts[0] + statementNameColumnGap + strings.Join(parts[1:], statementColumnGap)
}

func formatStatementCell(stmt domain.FinancialStatement, row domain.StatementRow, idx int) string {
	if idx < 0 || idx >= len(row.Values) || !row.Values[idx].Present {
		return "-"
	}
	valueText := formatStatementValue(row.Values[idx].Value)
	if idx != 0 {
		return valueText
	}
	growthText, ok := formatStatementGrowth(stmt, row.Values, idx)
	if !ok {
		return valueText
	}
	return valueText + " " + growthText
}

func formatStatementGrowth(stmt domain.FinancialStatement, values []domain.StatementValue, idx int) (string, bool) {
	changePct, ok := statementGrowthPercent(stmt, values, idx)
	if !ok {
		return "", false
	}
	text := fmt.Sprintf("(%+.1f%%)", changePct)
	return marketMoveStyle(changePct).Render(text), true
}

func formatStatementValue(v float64) string {
	switch {
	case v == 0:
		return "0"
	case math.Abs(v) >= 1000:
		return ui.FormatCompactInt(int64(v))
	case math.Abs(v) >= 100:
		return fmt.Sprintf("%.1f", v)
	default:
		return fmt.Sprintf("%.2f", v)
	}
}
