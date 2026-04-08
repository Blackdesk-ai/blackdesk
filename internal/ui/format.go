package ui

import (
	"fmt"
	"time"
)

func FormatMoney(v float64) string {
	return fmt.Sprintf("%.2f", v)
}

func FormatCompactInt(v int64) string {
	sign := ""
	if v < 0 {
		sign = "-"
		v = -v
	}
	switch {
	case v >= 1_000_000_000_000:
		return fmt.Sprintf("%s%.2fT", sign, float64(v)/1_000_000_000_000)
	case v >= 1_000_000_000:
		return fmt.Sprintf("%s%.2fB", sign, float64(v)/1_000_000_000)
	case v >= 1_000_000:
		return fmt.Sprintf("%s%.2fM", sign, float64(v)/1_000_000)
	case v >= 1_000:
		return fmt.Sprintf("%s%.2fK", sign, float64(v)/1_000)
	default:
		return fmt.Sprintf("%s%d", sign, v)
	}
}

func FormatPercent(v float64) string {
	return fmt.Sprintf("%.2f%%", v)
}

func FormatTimestamp(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Local().Format("2006-01-02 15:04")
}
