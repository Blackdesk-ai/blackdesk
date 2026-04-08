package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"blackdesk/internal/domain"
)

func tabIdentity(idx int) string {
	switch idx {
	case tabScreener:
		return "screener workspace"
	case tabQuote:
		return "quote and research"
	case tabNews:
		return "market news wire"
	case tabAI:
		return "AI workspace"
	default:
		return "global market board"
	}
}

func (m Model) renderTickerTape(muted, pos, neg lipgloss.Style, width int) string {
	return ""
}

func renderBadgeRow(parts ...string) string {
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			continue
		}
		out = append(out, "["+part+"]")
	}
	return strings.Join(out, " ")
}

func (m Model) displaySeries() domain.PriceSeries {
	series := m.series
	if len(series.Candles) == 0 {
		return series
	}
	displayPrice, _, _, _ := displayQuoteLine(m.quote)
	if displayPrice == 0 {
		return series
	}
	series.Candles = append([]domain.Candle(nil), series.Candles...)
	last := series.Candles[len(series.Candles)-1]
	if last.Close == displayPrice {
		return series
	}
	pointTime := m.quote.RegularMarketTime
	if pointTime.IsZero() || pointTime.Before(last.Time) {
		pointTime = last.Time
	}
	series.Candles = append(series.Candles, domain.Candle{
		Time:   pointTime,
		Open:   last.Close,
		High:   maxFloat(last.Close, displayPrice),
		Low:    minFloat(last.Close, displayPrice),
		Close:  displayPrice,
		Volume: 0,
	})
	return series
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func descriptorBadge(state domain.MarketState, move float64) string {
	switch {
	case move >= 2:
		return "RISK ON"
	case move <= -2:
		return "SELL PRESSURE"
	}
	switch state {
	case domain.MarketStatePre:
		return "PRE"
	case domain.MarketStatePost:
		return "AFTER HOURS"
	case domain.MarketStateRegular:
		return ""
	case domain.MarketStateClosed:
		return "CLOSED"
	default:
		return "IDLE"
	}
}

func sessionBadge(q domain.QuoteSnapshot) string {
	switch q.MarketState {
	case domain.MarketStatePre:
		return "PRE"
	case domain.MarketStatePost:
		return "POST"
	case domain.MarketStateRegular:
		return ""
	case domain.MarketStateClosed:
		return "CLOSED"
	default:
		return "IDLE"
	}
}

func (m Model) appendDeskFooter(content string, section lipgloss.Style, width, height int) string {
	lines := splitLines(strings.TrimRight(content, "\n"))
	sparkWidth := max(12, width-2)
	footer := []string{
		"",
		section.Render("DESK"),
		descriptorBadge(m.quote.MarketState, m.quote.ChangePercent),
		sparklineBlock(extractCloses(m.series.Candles), sparkWidth),
	}
	if len(lines)+len(footer) > height {
		lines = lines[:max(0, height-len(footer))]
	}
	lines = append(lines, footer...)
	return clipLines(strings.Join(lines, "\n"), height)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
