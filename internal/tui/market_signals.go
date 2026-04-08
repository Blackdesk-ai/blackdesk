package tui

import (
	"fmt"
	"math"
	"strings"

	"blackdesk/internal/domain"
)

func colorizeRegimeLabel(text string) string {
	switch text {
	case "low vol bull", "risk on":
		return marketMoveStyle(1).Render(text)
	case "risk off", "high stress", "trend weak":
		return marketMoveStyle(-1).Render(text)
	default:
		return marketMoveStyle(0).Render(text)
	}
}

func marketRegimeLabel(m Model) string {
	return strings.ToUpper(regimeSummary(m))
}

func colorizeBreadthLine(text string) string {
	var up, down int
	if _, err := fmt.Sscanf(text, "%d up / %d down", &up, &down); err != nil {
		return text
	}
	switch {
	case up > down:
		return marketMoveStyle(1).Render(text)
	case up < down:
		return marketMoveStyle(-1).Render(text)
	default:
		return marketMoveStyle(0).Render(text)
	}
}

func renderHeatMeter(label string, value float64, width int) string {
	value = clampFloat(value, 0, 1)
	filled := int(math.Round(value * float64(width)))
	if filled > width {
		filled = width
	}
	bar := strings.Repeat("=", filled) + strings.Repeat(".", max(0, width-filled))
	move := (value - 0.5) * 2
	return fmt.Sprintf("%-10s %s %s", label, marketMoveStyle(move).Render(bar), marketMoveStyle(move).Render(fmt.Sprintf("%3.0f%%", value*100)))
}

func marketPulseGlyph(symbol, active string, move float64) string {
	if !strings.EqualFold(symbol, active) {
		return "·"
	}
	switch {
	case move > 0:
		return "▲"
	case move < 0:
		return "▼"
	default:
		return "■"
	}
}

func candleBadge(move float64) string {
	switch {
	case move > 1:
		return "▲"
	case move < -1:
		return "▼"
	default:
		return "■"
	}
}

func volumePulse(q domain.QuoteSnapshot) float64 {
	if q.AverageVolume == 0 {
		return 0
	}
	return clampFloat(float64(q.Volume)/float64(q.AverageVolume), 0, 2) / 2
}

func yearlyRangePct(f domain.FundamentalsSnapshot, price float64) float64 {
	if price == 0 || f.FiftyTwoWeekHigh <= f.FiftyTwoWeekLow {
		return 0
	}
	return clampFloat((price-f.FiftyTwoWeekLow)/(f.FiftyTwoWeekHigh-f.FiftyTwoWeekLow), 0, 1)
}

func rangeBadge(v float64) string {
	switch {
	case v >= 0.85:
		return "AT HIGHS"
	case v >= 0.6:
		return "UPPER RANGE"
	case v <= 0.2:
		return "NEAR LOWS"
	default:
		return "MID RANGE"
	}
}

func yieldBadge(yield float64) string {
	switch {
	case yield >= 0.03:
		return "YIELD+"
	case yield > 0:
		return "YIELD"
	default:
		return "NO YIELD"
	}
}
