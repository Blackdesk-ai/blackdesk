package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderNewsRight(section, label, muted lipgloss.Style, width, height int) string {
	var b strings.Builder
	b.WriteString(section.Render("STORY") + "\n\n")
	item, ok := m.currentMarketNews()
	if !ok {
		if m.errMarketNews != nil {
			b.WriteString(renderWrappedTextBlock(muted, "Market wire unavailable: "+m.errMarketNews.Error(), width))
		} else {
			b.WriteString(renderWrappedTextBlock(muted, "Waiting for live market headlines…", width))
		}
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Width(max(20, width)).MaxWidth(max(20, width))
	b.WriteString(renderWrappedTextBlock(titleStyle, strings.Join(strings.Fields(item.Title), " "), width) + "\n\n")
	meta := []string{item.Publisher}
	if !item.Time.IsZero() {
		meta = append(meta, item.Time.Local().Format("Mon 15:04"))
		meta = append(meta, newsTimeLabel(item.Time))
	}
	if host := newsHost(item.URL); host != "" {
		meta = append(meta, host)
	}
	b.WriteString(muted.Render(truncateText(strings.Join(meta, " • "), max(20, width))) + "\n\n")

	summary := strings.TrimSpace(item.Summary)
	if summary == "" {
		summary = "Summary unavailable from the feed. Open the story for full article detail."
	}
	b.WriteString(renderWrappedTextBlock(lipgloss.NewStyle().Width(max(20, width)).MaxWidth(max(20, width)), summary, width))
	if strings.TrimSpace(item.URL) != "" {
		b.WriteString("\n\n" + muted.Render(truncateText(compactURLLabel(item.URL), max(20, width))))
		b.WriteString("\n\n\n" + muted.Render("Press o to open in browser"))
	}
	if m.errMarketNews != nil {
		b.WriteString("\n\n" + muted.Render("Wire may be stale: "+m.errMarketNews.Error()))
	}
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}
