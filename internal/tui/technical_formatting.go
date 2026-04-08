package tui

import (
	"fmt"
	"math"

	"blackdesk/internal/ui"
)

func formatSignedPercentRatio(v float64) string {
	if v == 0 {
		return "-"
	}
	return fmt.Sprintf("%+.2f%%", v*100)
}

func formatMetricSigned(v float64) string {
	if v == 0 {
		return "-"
	}
	return fmt.Sprintf("%+.2f", v)
}

func distanceLabel(price, level float64) string {
	if price == 0 || level == 0 {
		return "-"
	}
	return fmt.Sprintf("%s %s", relationLabel(price, level), formatSignedPercentRatio(relativeMove(price, level)))
}

func percentFromUnit(v float64) string {
	if v == 0 {
		return "-"
	}
	return fmt.Sprintf("%.0f%%", v*100)
}

func formatCompactFloat(v float64) string {
	if v == 0 {
		return "-"
	}
	return ui.FormatCompactInt(int64(math.Round(v)))
}

func formatRatio(v float64) string {
	if v == 0 {
		return "-"
	}
	return fmt.Sprintf("%.2fx", v)
}

func formatProbability(v float64) string {
	if v == 0 {
		return "-"
	}
	if v < 0.001 {
		return "<0.1%"
	}
	return fmt.Sprintf("%.1f%%", v*100)
}
