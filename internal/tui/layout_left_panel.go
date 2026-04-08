package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"blackdesk/internal/ui"
)

func (m Model) renderLeftPanel(section, muted, pos, neg lipgloss.Style, width, height int) string {
	if m.tabIdx == tabMarkets {
		return m.renderMarketLeft(section, lipgloss.NewStyle().Foreground(lipgloss.Color("245")), muted, width, height)
	}
	if m.tabIdx == tabScreener {
		return m.renderScreenerLeft(section, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#D8C9B8")), muted, width, height)
	}
	if m.tabIdx == tabNews {
		return m.renderNewsLeft(section, muted, width, height)
	}
	if m.tabIdx == tabAI {
		return m.renderAILeft(section, muted, width, height)
	}

	var b strings.Builder
	b.WriteString(section.Render("SYMBOLS") + " " + muted.Render("↑/↓") + "\n")
	b.WriteString("\n")
	tickerWidth := clamp(width/3, 7, 12)
	moveWidth := clamp(width/3, 11, 16)
	priceWidth := max(7, width-tickerWidth-moveWidth)
	b.WriteString(muted.Render(padRight("Ticker", tickerWidth)+padRight("Price", priceWidth)+"Change") + "\n")
	b.WriteString("\n")

	listLines := max(1, height-4)
	start := 0
	if m.selectedIdx >= listLines {
		start = m.selectedIdx - listLines + 1
	}
	start = max(start, m.watchlistScroll)

	indicatorLines := 0
	if start > 0 {
		indicatorLines++
	}
	if len(m.config.Watchlist)-start > listLines-indicatorLines {
		indicatorLines++
	}

	visibleRows := max(1, listLines-indicatorLines)
	maxStart := max(0, len(m.config.Watchlist)-visibleRows)
	if start > maxStart {
		start = maxStart
	}

	showTopMore := start > 0
	showBottomMore := len(m.config.Watchlist)-start > visibleRows
	visibleRows = max(1, listLines-boolToInt(showTopMore)-boolToInt(showBottomMore))
	maxStart = max(0, len(m.config.Watchlist)-visibleRows)
	if start > maxStart {
		start = maxStart
	}
	if m.selectedIdx < start {
		start = m.selectedIdx
	}
	if m.selectedIdx >= start+visibleRows {
		start = m.selectedIdx - visibleRows + 1
	}
	if start > maxStart {
		start = maxStart
	}
	showTopMore = start > 0
	showBottomMore = len(m.config.Watchlist)-start > visibleRows

	if showTopMore {
		b.WriteString(muted.Render("↑ more") + "\n")
	}

	end := min(len(m.config.Watchlist), start+visibleRows)
	for i := start; i < end; i++ {
		symbol := m.config.Watchlist[i]
		prefix := "  "
		if i == m.selectedIdx {
			prefix = "▶ "
		}
		priceText := "--"
		moveText := "--"
		priceStyle := muted
		moveStyle := muted
		quote, ok := m.watchQuotes[strings.ToUpper(symbol)]
		if ok {
			displayPrice, _, displayChangePercent, _ := displayQuoteLine(quote)
			priceText = ui.FormatMoney(displayPrice)
			moveText = moveArrow(displayChangePercent) + " " + fmt.Sprintf("%+.2f%%", displayChangePercent)
			if displayChangePercent > 0 {
				priceStyle = pos
				moveStyle = pos
			} else if displayChangePercent < 0 {
				priceStyle = neg
				moveStyle = neg
			}
		}
		leftPart := prefix + padRight(symbol, max(1, tickerWidth-len([]rune(prefix))))
		pricePart := priceStyle.Render(padRight(priceText, priceWidth))
		baseWidth := len([]rune(prefix)) + max(1, tickerWidth-len([]rune(prefix))) + priceWidth
		moveWidth := max(4, width-baseWidth)
		b.WriteString(leftPart + pricePart + moveStyle.Render(padRight(moveText, moveWidth)) + "\n")
	}
	if showBottomMore {
		b.WriteString(muted.Render("↓ more") + "\n")
	}
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}
