package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"blackdesk/internal/domain"
)

func renderStatementHeaderBlock(activeKind domain.StatementKind, activeFreq domain.StatementFrequency, muted lipgloss.Style, width int) string {
	left := renderStatementKindTabs(activeKind) + " " + muted.Render("←/→")
	right := renderStatementFrequencyTabs(activeFreq) + " " + muted.Render("[/]")
	if lipgloss.Width(left)+2+lipgloss.Width(right) <= width {
		gap := max(2, width-lipgloss.Width(left)-lipgloss.Width(right))
		return left + strings.Repeat(" ", gap) + right
	}
	return strings.Join([]string{
		left,
		lipgloss.NewStyle().Width(width).Align(lipgloss.Right).Render(right),
	}, "\n")
}

func renderStatementKindTabs(activeKind domain.StatementKind) string {
	activeStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#1C1712")).Background(lipgloss.Color("#E7B66B")).Padding(0, 1)
	idleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#A6A29D")).Padding(0, 1)
	kindTabs := []struct {
		kind  domain.StatementKind
		label string
	}{
		{kind: domain.StatementKindIncome, label: "Income"},
		{kind: domain.StatementKindBalanceSheet, label: "Balance Sheet"},
		{kind: domain.StatementKindCashFlow, label: "Cash Flow"},
	}
	parts := make([]string, 0, len(kindTabs))
	for _, tab := range kindTabs {
		style := idleStyle
		if tab.kind == activeKind {
			style = activeStyle
		}
		parts = append(parts, style.Render(tab.label))
	}
	return strings.Join(parts, " ")
}

func renderStatementFrequencyTabs(activeFreq domain.StatementFrequency) string {
	activeStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#1C1712")).Background(lipgloss.Color("#E7B66B")).Padding(0, 1)
	idleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#A6A29D")).Padding(0, 1)
	freqTabs := []struct {
		freq  domain.StatementFrequency
		label string
	}{
		{freq: domain.StatementFrequencyAnnual, label: "Annual"},
		{freq: domain.StatementFrequencyQuarterly, label: "Quarterly"},
	}
	parts := make([]string, 0, len(freqTabs))
	for _, tab := range freqTabs {
		style := idleStyle
		if tab.freq == activeFreq {
			style = activeStyle
		}
		parts = append(parts, style.Render(tab.label))
	}
	return strings.Join(parts, " ")
}
