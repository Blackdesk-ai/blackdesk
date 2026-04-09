package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"blackdesk/internal/domain"
)

func marketRiskLine(risk domain.MarketRiskSnapshot) string {
	if !risk.Available || risk.Min >= risk.Max {
		return "N/A"
	}
	barWidth := risk.Max - risk.Min
	if barWidth <= 0 {
		return "N/A"
	}
	filled := risk.Score - risk.Min
	if filled < 0 {
		filled = 0
	}
	if filled > barWidth {
		filled = barWidth
	}
	bar := strings.Repeat("=", filled) + strings.Repeat(".", max(0, barWidth-filled))
	label := strings.ToUpper(strings.TrimSpace(risk.Label))
	if label == "" {
		label = marketRiskStanceLabel(risk)
	}
	style := marketRiskMeterStyle(risk.Score)
	return style.Render(bar) + " " + style.Render(fmt.Sprintf("%s %s", formatMarketRiskScore(risk.Score), label))
}

func marketRiskBiasText(risk domain.MarketRiskSnapshot) string {
	if !risk.Available || risk.Min >= risk.Max {
		return "N/A"
	}
	switch stance := marketRiskStanceKey(risk); stance {
	case "neutral":
		return "Neutral"
	default:
		return strings.Title(marketRiskStrength(risk)) + " " + strings.ReplaceAll(stance, "_", "-")
	}
}

func marketRiskStanceKey(risk domain.MarketRiskSnapshot) string {
	switch {
	case risk.Score > 0:
		return "risk_on"
	case risk.Score < 0:
		return "risk_off"
	default:
		return "neutral"
	}
}

func marketRiskStanceLabel(risk domain.MarketRiskSnapshot) string {
	switch marketRiskStanceKey(risk) {
	case "risk_on":
		return "RISK ON"
	case "risk_off":
		return "RISK OFF"
	default:
		return "NEUTRAL"
	}
}

func marketRiskStrength(risk domain.MarketRiskSnapshot) string {
	maxAbs := max(-risk.Min, risk.Max)
	if maxAbs <= 0 {
		return "neutral"
	}
	absScore := risk.Score
	if absScore < 0 {
		absScore = -absScore
	}
	if absScore == 0 {
		return "neutral"
	}
	ratio := float64(absScore) / float64(maxAbs)
	switch {
	case ratio <= 0.25:
		return "mild"
	case ratio <= 0.5:
		return "moderate"
	case ratio <= 0.75:
		return "strong"
	default:
		return "extreme"
	}
}

func formatMarketRiskScore(score int) string {
	if score == 0 {
		return "0"
	}
	return fmt.Sprintf("%+d", score)
}

func renderMarketSnapshotLine(label lipgloss.Style, key, value string, width int) string {
	keyWidth := 8
	if width <= keyWidth+1 {
		return ansi.Truncate(label.Render(key)+" "+value, max(1, width), "")
	}
	left := lipgloss.NewStyle().Width(keyWidth).Render(label.Render(key))
	line := left + " " + value
	return ansi.Truncate(line, max(1, width), "")
}

func marketRiskMeterStyle(score int) lipgloss.Style {
	switch {
	case score > 1:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#62D394"))
	case score < -1:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF7A73"))
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#E7B66B"))
	}
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
