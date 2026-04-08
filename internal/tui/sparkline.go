package tui

import (
	"math"
	"strings"

	"blackdesk/internal/domain"
)

func extractCloses(candles []domain.Candle) []float64 {
	out := make([]float64, 0, len(candles))
	for _, candle := range candles {
		out = append(out, candle.Close)
	}
	return out
}

func sparklineBlock(values []float64, width int) string {
	if width <= 0 {
		return ""
	}
	if len(values) == 0 {
		return strings.Repeat(".", width)
	}
	levels := []rune("▁▂▃▄▅▆▇█")
	if len(values) != width {
		if len(values) == 1 {
			value := values[0]
			values = make([]float64, width)
			for i := range values {
				values[i] = value
			}
		} else {
			reduced := make([]float64, 0, width)
			last := float64(len(values) - 1)
			for i := 0; i < width; i++ {
				pos := float64(i) * last / float64(width-1)
				loIdx := int(math.Floor(pos))
				hiIdx := int(math.Ceil(pos))
				if hiIdx >= len(values) {
					hiIdx = len(values) - 1
				}
				if loIdx == hiIdx {
					reduced = append(reduced, values[loIdx])
					continue
				}
				ratio := pos - float64(loIdx)
				reduced = append(reduced, values[loIdx]+(values[hiIdx]-values[loIdx])*ratio)
			}
			values = reduced
		}
	}
	lo, hi := values[0], values[0]
	for _, v := range values {
		if v < lo {
			lo = v
		}
		if v > hi {
			hi = v
		}
	}
	if hi == lo {
		return strings.Repeat(string(levels[3]), len(values))
	}
	var b strings.Builder
	for _, v := range values {
		idx := int(math.Round((v - lo) / (hi - lo) * float64(len(levels)-1)))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(levels) {
			idx = len(levels) - 1
		}
		b.WriteRune(levels[idx])
	}
	return b.String()
}

func padRight(s string, width int) string {
	r := []rune(s)
	if len(r) >= width {
		return string(r[:width])
	}
	return s + strings.Repeat(" ", width-len(r))
}
