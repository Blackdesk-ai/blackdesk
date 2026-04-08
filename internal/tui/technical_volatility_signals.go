package tui

func hvBaseline(hv63, hv252 float64) float64 {
	switch {
	case hv63 > 0 && hv252 > 0:
		return (hv63 + hv252) / 2
	case hv63 > 0:
		return hv63
	default:
		return hv252
	}
}

func hvLabel(hv21, hv63, hv252 float64) string {
	baseline := hvBaseline(hv63, hv252)
	switch {
	case hv21 == 0:
		return "-"
	case baseline == 0:
		return "unscaled"
	}
	ratio := hv21 / baseline
	switch {
	case ratio <= 0.8:
		return "compressed"
	case ratio <= 1.15:
		return "normal"
	case ratio <= 1.35:
		return "elevated"
	default:
		return "high vol"
	}
}

func hvMove(hv21, hv63, hv252 float64) float64 {
	baseline := hvBaseline(hv63, hv252)
	switch {
	case hv21 == 0 || baseline == 0:
		return 0
	}
	ratio := hv21 / baseline
	switch {
	case ratio <= 0.8:
		return 1
	case ratio <= 1.15:
		return 0
	default:
		return -1
	}
}

func hvRankLabel(v float64, ok bool) string {
	return hvPercentileLabel(v, ok)
}

func hvRankMove(v float64, ok bool) float64 {
	return hvPercentileMove(v, ok)
}

func hvPercentileLabel(v float64, ok bool) string {
	if !ok {
		return "-"
	}
	switch {
	case v < 10:
		return "very low"
	case v < 30:
		return "low"
	case v <= 70:
		return "normal"
	case v <= 90:
		return "high"
	default:
		return "very high"
	}
}

func hvPercentileMove(v float64, ok bool) float64 {
	if !ok {
		return 0
	}
	switch {
	case v < 10:
		return 1
	case v < 30:
		return 0.5
	case v <= 70:
		return 0
	case v <= 90:
		return -0.5
	default:
		return -1
	}
}
