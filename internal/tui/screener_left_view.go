package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"blackdesk/internal/domain"
)

func (m Model) renderScreenerLeft(section, label, muted lipgloss.Style, width, height int) string {
	var b strings.Builder
	b.WriteString(section.Render("SCREENERS") + " " + muted.Render("←/→") + "\n\n")

	rows, selectedRow := screenerSidebarRows(m.screenerDefs, m.screenerIdx, label, muted, width)
	listHeight := max(4, height-10)
	start := 0
	if selectedRow >= listHeight {
		start = selectedRow - listHeight/2
	}
	maxStart := max(0, len(rows)-listHeight)
	if start > maxStart {
		start = maxStart
	}
	end := min(len(rows), start+listHeight)
	if start > 0 {
		b.WriteString(muted.Render("↑ more") + "\n")
	}
	for _, row := range rows[start:end] {
		b.WriteString(row + "\n")
	}
	if end < len(rows) {
		b.WriteString(muted.Render("↓ more") + "\n")
	}

	def := m.currentScreenerDefinition()
	result := m.screenerResult
	if strings.TrimSpace(result.Definition.ID) == "" {
		result.Definition = def
	}
	b.WriteString("\n" + section.Render("DESK") + "\n\n")
	b.WriteString(renderWrappedLabelLine(label, "Universe "+valueOrDash(strings.Title(result.Definition.Kind)), width) + "\n")
	b.WriteString(renderWrappedLabelLine(label, fmt.Sprintf("Loaded %d/%d", len(result.Items), max(result.Total, len(result.Items))), width) + "\n")
	if !result.UpdatedAt.IsZero() {
		b.WriteString(renderWrappedLabelLine(label, "Updated "+result.UpdatedAt.Local().Format("15:04:05"), width) + "\n")
	}
	if m.errScreener != nil {
		b.WriteString("\n" + renderWrappedTextBlock(muted, "Last error: "+m.errScreener.Error(), width))
	}
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func screenerSidebarRows(defs []domain.ScreenerDefinition, selectedIdx int, label, muted lipgloss.Style, width int) ([]string, int) {
	rows := make([]string, 0, len(defs)+8)
	selectedRow := 0
	currentCategory := ""
	for i, def := range defs {
		if def.Category != currentCategory {
			currentCategory = def.Category
			rows = append(rows, label.Render(strings.ToUpper(currentCategory)))
		}
		prefix := "  "
		style := lipgloss.NewStyle()
		if i == selectedIdx {
			prefix = "▶ "
			style = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F3EBDD"))
			selectedRow = len(rows)
		}
		text := prefix + ansi.Truncate(def.Name, max(8, width-2), "")
		rows = append(rows, style.Render(text))
	}
	return rows, selectedRow
}
