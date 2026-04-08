package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderNewsCenter(section, label, muted lipgloss.Style, width, height int) string {
	var b strings.Builder
	b.WriteString(section.Render("MARKET WIRE") + " " + muted.Render("↑/↓") + "\n\n")
	if len(m.marketNews) == 0 {
		if m.errMarketNews != nil {
			b.WriteString(renderWrappedTextBlock(muted, "Market wire unavailable: "+m.errMarketNews.Error(), width))
		} else {
			b.WriteString(renderWrappedTextBlock(muted, "Waiting for live market headlines…", width))
		}
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}

	listLines := max(1, height-2)
	start := min(m.marketNewsScroll, max(0, len(m.marketNews)-listLines))
	if m.marketNewsSel < start {
		start = m.marketNewsSel
	}
	if m.marketNewsSel >= start+listLines {
		start = m.marketNewsSel - listLines + 1
	}
	start = max(0, min(start, max(0, len(m.marketNews)-listLines)))
	showTopMore := start > 0
	showBottomMore := len(m.marketNews)-start > listLines
	visibleItems := max(1, listLines-boolToInt(showTopMore)-boolToInt(showBottomMore))
	maxStart := max(0, len(m.marketNews)-visibleItems)
	if start > maxStart {
		start = maxStart
	}
	if m.marketNewsSel < start {
		start = m.marketNewsSel
	}
	if m.marketNewsSel >= start+visibleItems {
		start = m.marketNewsSel - visibleItems + 1
	}
	if start > maxStart {
		start = maxStart
	}
	showTopMore = start > 0
	showBottomMore = len(m.marketNews)-start > visibleItems
	if showTopMore {
		b.WriteString(muted.Render("↑ more") + "\n")
	}
	end := min(len(m.marketNews), start+visibleItems)
	titleWidth := max(24, width)
	titleStyle := lipgloss.NewStyle().Bold(true).Width(titleWidth).MaxWidth(titleWidth)
	newsDot := lipgloss.NewStyle().Foreground(lipgloss.Color("#55C778"))
	for i := start; i < end; i++ {
		item := m.marketNews[i]
		prefix := "  "
		if i == m.marketNewsSel {
			prefix = "▶ "
		} else if _, ok := m.marketNewsFresh[marketNewsIdentity(item)]; ok {
			prefix = newsDot.Render("● ")
		}
		timePrefix := "--:--"
		if !item.Time.IsZero() {
			timePrefix = item.Time.Local().Format("15:04")
		}
		title := strings.Join(strings.Fields(item.Title), " ")
		prefixText := "  " + timePrefix + " "
		lineWidth := max(1, titleWidth-lipgloss.Width(prefixText))
		titleRendered := truncateText(title, lineWidth)
		b.WriteString(titleStyle.Render(prefix+muted.Render(timePrefix+" ")+titleRendered) + "\n")
	}
	if showBottomMore {
		b.WriteString(muted.Render("↓ more") + "\n")
	}
	if m.errMarketNews != nil {
		b.WriteString("\n\n" + muted.Render("Wire may be stale: "+m.errMarketNews.Error()))
	}
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}
