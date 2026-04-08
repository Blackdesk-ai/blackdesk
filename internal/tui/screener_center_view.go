package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderScreenerCenter(section, label, muted, pos, neg lipgloss.Style, width, height int) string {
	var b strings.Builder
	result := m.screenerResult
	def := m.currentScreenerDefinition()
	if strings.TrimSpace(result.Definition.ID) == "" {
		result.Definition = def
	}

	title := valueOrDash(result.Definition.Name)
	desc := strings.TrimSpace(result.Definition.Description)
	if desc == "" {
		desc = "Predefined Yahoo Finance screener."
	}
	b.WriteString(section.Render(strings.ToUpper(title)) + "\n")
	b.WriteString(renderWrappedTextBlock(muted, desc, width) + "\n\n")

	metaParts := []string{}
	if strings.TrimSpace(result.Definition.Kind) != "" {
		metaParts = append(metaParts, strings.ToUpper(result.Definition.Kind))
	}
	if len(result.Items) > 0 || result.Total > 0 {
		metaParts = append(metaParts, fmt.Sprintf("%d shown / %d total", len(result.Items), max(result.Total, len(result.Items))))
	}
	if !result.UpdatedAt.IsZero() {
		metaParts = append(metaParts, "Updated "+result.UpdatedAt.Local().Format("15:04:05"))
	}
	if len(metaParts) > 0 {
		b.WriteString(muted.Render(strings.Join(metaParts, "  •  ")) + "\n\n")
	}

	if len(result.Items) == 0 {
		if m.errScreener != nil {
			b.WriteString(renderWrappedTextBlock(muted, "Screener unavailable: "+m.errScreener.Error(), width))
		} else if !m.screenerAvailable() {
			b.WriteString(renderWrappedTextBlock(muted, "Active provider does not expose screeners.", width))
		} else {
			b.WriteString(renderWrappedTextBlock(muted, "Press r to load the selected screener.", width))
		}
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}

	metricKey, metricLabel := screenerPrimaryMetric(result)
	b.WriteString(muted.Render(renderScreenerHeader(width, label, metricLabel)) + "\n")

	listHeight := max(4, height-lipgloss.Height(strings.TrimRight(b.String(), "\n"))-1)
	start := min(max(0, m.screenerScroll), max(0, len(result.Items)-listHeight))
	if m.screenerSel < start {
		start = m.screenerSel
	}
	if m.screenerSel >= start+listHeight {
		start = m.screenerSel - listHeight + 1
	}
	start = max(0, min(start, max(0, len(result.Items)-listHeight)))
	if start > 0 {
		b.WriteString(muted.Render("↑ more") + "\n")
		listHeight--
	}
	end := min(len(result.Items), start+listHeight)
	if end < len(result.Items) && listHeight > 0 {
		end--
	}
	for i := start; i < end; i++ {
		b.WriteString(renderScreenerRow(result.Items[i], i == m.screenerSel, width, label, metricKey) + "\n")
	}
	if end < len(result.Items) {
		b.WriteString(muted.Render("↓ more") + "\n")
	}
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}
