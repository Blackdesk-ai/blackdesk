package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"blackdesk/internal/domain"
	"blackdesk/internal/ui"
)

func buildHeaderMeta(m Model, muted, pos, neg lipgloss.Style, width int) string {
	if width <= 0 {
		return ""
	}
	displayPrice, displayChange, displayChangePercent, _ := displayQuoteLine(m.quote)
	symbolBlock := strings.ToUpper(m.activeSymbol())
	if name := strings.TrimSpace(m.quote.ShortName); name != "" {
		symbolBlock += " " + name
	}
	priceText := ui.FormatMoney(displayPrice)
	changeText := fmt.Sprintf("%+.2f%%", displayChangePercent)

	priceStyle := muted
	changeStyle := muted
	if displayChange > 0 {
		priceStyle = pos
		changeStyle = pos
	} else if displayChange < 0 {
		priceStyle = neg
		changeStyle = neg
	}

	staticWidth := lipgloss.Width(priceText) + lipgloss.Width(changeText) + 6
	symbolWidth := max(1, width-staticWidth)
	if lipgloss.Width(symbolBlock) > symbolWidth {
		symbolBlock = truncateText(symbolBlock, symbolWidth)
	}

	meta := muted.Render(symbolBlock) + muted.Render("   ") + priceStyle.Render(priceText) + muted.Render("   ") + changeStyle.Render(changeText)
	if lipgloss.Width(meta) > width {
		meta = truncateText(meta, width)
	}
	return meta
}

func shortRefresh(ts time.Time) string {
	if ts.IsZero() {
		return "cold start"
	}
	return "refreshed " + ts.Local().Format("15:04:05")
}

func moveArrow(changePercent float64) string {
	switch {
	case changePercent > 0:
		return "▲"
	case changePercent < 0:
		return "▼"
	default:
		return "■"
	}
}

func displayQuoteLine(quote domain.QuoteSnapshot) (price, change, changePercent float64, sessionLabel string) {
	switch quote.MarketState {
	case domain.MarketStatePre:
		if quote.PreMarketPrice > 0 {
			return quote.PreMarketPrice, quote.PreMarketChange, quote.PreMarketChangePerc, ""
		}
		return quote.Price, quote.Change, quote.ChangePercent, "Last close"
	case domain.MarketStateRegular:
		return quote.Price, quote.Change, quote.ChangePercent, ""
	case domain.MarketStatePost:
		if quote.PostMarketPrice > 0 {
			return quote.PostMarketPrice, quote.PostMarketChange, quote.PostMarketChangePerc, ""
		}
		return quote.Price, quote.Change, quote.ChangePercent, "Last close"
	case domain.MarketStateClosed:
		return quote.Price, quote.Change, quote.ChangePercent, ""
	}
	return quote.Price, quote.Change, quote.ChangePercent, "Last"
}
