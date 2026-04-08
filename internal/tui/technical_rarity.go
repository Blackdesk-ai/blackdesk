package tui

import (
	"math"

	"github.com/charmbracelet/lipgloss"
)

func priceZLabel(v float64) string {
	switch {
	case v == 0:
		return "-"
	case v >= 2:
		return "stretch up"
	case v <= -2:
		return "stretch down"
	case v >= 1:
		return "rich"
	case v <= -1:
		return "cheap"
	default:
		return "normal"
	}
}

func rarityStyle(z float64) lipgloss.Style {
	style := lipgloss.NewStyle()
	switch rarityLevel(z) {
	case 3:
		return style.Foreground(lipgloss.Color("#FF7A73"))
	case 2:
		return style.Foreground(lipgloss.Color("#F2A65A"))
	case 1:
		return style.Foreground(lipgloss.Color("#E7B66B"))
	default:
		return style.Foreground(lipgloss.Color("#A6A29D"))
	}
}

func rarityLevel(z float64) int {
	absZ := math.Abs(z)
	switch {
	case absZ >= 3:
		return 3
	case absZ >= 2:
		return 2
	case absZ >= 1:
		return 1
	default:
		return 0
	}
}

func rarityLabel(z float64) string {
	switch {
	case z == 0:
		return "-"
	case rarityLevel(z) >= 3:
		return "extreme"
	case rarityLevel(z) == 2:
		return "high"
	case rarityLevel(z) == 1:
		return "elevated"
	default:
		return "normal"
	}
}
