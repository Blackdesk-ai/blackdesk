package tui

import (
	"fmt"
	"strings"

	"blackdesk/internal/domain"
	"blackdesk/internal/ui"
)

func formatMetricFloat(v float64) string {
	if v == 0 {
		return "-"
	}
	return fmt.Sprintf("%.2f", v)
}

func recommendationBadge(key string) string {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "":
		return ""
	case "strong_buy":
		return "STRONG BUY"
	case "buy":
		return "BUY"
	case "hold":
		return "HOLD"
	case "underperform":
		return "UNDERPERFORM"
	case "sell":
		return "SELL"
	default:
		return strings.ReplaceAll(strings.ToUpper(key), "_", " ")
	}
}

func recommendationMove(key string) float64 {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "strong_buy":
		return 2
	case "buy":
		return 1
	case "hold":
		return 0
	case "underperform", "sell":
		return -1
	default:
		return 0
	}
}

func colorizeRecommendationBadge(key string) string {
	badge := recommendationBadge(key)
	if badge == "" {
		return ""
	}
	return marketMoveStyle(recommendationMove(key)).Render("[" + badge + "]")
}

func formatAnalystMean(v float64) string {
	if v == 0 {
		return "-"
	}
	return fmt.Sprintf("%.2f / 5", v)
}

func formatAnalystOpinions(v int) string {
	if v <= 0 {
		return "-"
	}
	return fmt.Sprintf("%d", v)
}

func analystTargetLine(f domain.FundamentalsSnapshot) string {
	if f.TargetMeanPrice == 0 {
		return "-"
	}
	return formatMoneyDash(f.TargetMeanPrice)
}

func analystUpsideLine(f domain.FundamentalsSnapshot, price float64) string {
	if price == 0 || f.TargetMeanPrice == 0 {
		return "-"
	}
	return fmt.Sprintf("%s vs px", ui.FormatPercent((f.TargetMeanPrice-price)/price*100))
}

func formatMoneyDash(v float64) string {
	if v == 0 {
		return "-"
	}
	return ui.FormatMoney(v)
}

func percentDash(v float64) string {
	if v == 0 {
		return "-"
	}
	return ui.FormatPercent(v * 100)
}

func formatOptionalPercent(v float64, ok bool) string {
	if !ok {
		return "-"
	}
	return ui.FormatPercent(v * 100)
}

func formatOptionalScaledFloat(v float64, ok bool, scale float64) string {
	if !ok {
		return "-"
	}
	return fmt.Sprintf("%.2f", v*scale)
}

func yearlyRangeText(f domain.FundamentalsSnapshot, price float64) string {
	if f.FiftyTwoWeekLow == 0 && f.FiftyTwoWeekHigh == 0 {
		return "-"
	}
	if price == 0 || f.FiftyTwoWeekHigh <= f.FiftyTwoWeekLow {
		return fmt.Sprintf("%s to %s", ui.FormatMoney(f.FiftyTwoWeekLow), ui.FormatMoney(f.FiftyTwoWeekHigh))
	}
	pos := (price - f.FiftyTwoWeekLow) / (f.FiftyTwoWeekHigh - f.FiftyTwoWeekLow)
	return fmt.Sprintf("%s to %s (%s)", ui.FormatMoney(f.FiftyTwoWeekLow), ui.FormatMoney(f.FiftyTwoWeekHigh), ui.FormatPercent(clampFloat(pos, 0, 1)*100))
}

func relationLabel(price, level float64) string {
	if price == 0 || level == 0 {
		return "-"
	}
	if price > level {
		return "above"
	}
	if price < level {
		return "below"
	}
	return "at"
}

func clampFloat(v, low, high float64) float64 {
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}
