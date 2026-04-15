package tui

import "github.com/charmbracelet/lipgloss"

func marketMoveStyle(move float64) lipgloss.Style {
	style := lipgloss.NewStyle()
	switch {
	case move > 0:
		return style.Foreground(lipgloss.Color("#62D394"))
	case move < 0:
		return style.Foreground(lipgloss.Color("#FF7A73"))
	default:
		return style.Foreground(lipgloss.Color("#A6A29D"))
	}
}

func qarpScoreStyle(score float64) lipgloss.Style {
	displayScore := score * 100
	style := lipgloss.NewStyle()
	switch {
	case displayScore > 1.5:
		return style.Foreground(lipgloss.Color("#62D394")).Bold(true)
	case displayScore > 1.2:
		return style.Foreground(lipgloss.Color("#62D394"))
	case displayScore >= 0.8:
		return style.Foreground(lipgloss.Color("#E7B66B"))
	case displayScore >= 0.5:
		return style.Foreground(lipgloss.Color("#A6A29D"))
	default:
		return style.Foreground(lipgloss.Color("#FF7A73"))
	}
}

func colorizeMarketPrice(text string, move float64, styled bool) string {
	if !styled || text == "--" {
		return text
	}
	return marketMoveStyle(move).Render(text)
}

func colorizeQARPScore(text string, score float64, styled bool) string {
	if !styled || text == "--" {
		return text
	}
	return qarpScoreStyle(score).Render(text)
}

func r40ScoreStyle(score float64) lipgloss.Style {
	displayScore := score * 100
	style := lipgloss.NewStyle()
	switch {
	case displayScore > 60:
		return style.Foreground(lipgloss.Color("#62D394")).Bold(true)
	case displayScore > 40:
		return style.Foreground(lipgloss.Color("#62D394"))
	case displayScore >= 25:
		return style.Foreground(lipgloss.Color("#D8C9B8"))
	case displayScore >= 15:
		return style.Foreground(lipgloss.Color("#E7B66B"))
	default:
		return style.Foreground(lipgloss.Color("#FF7A73"))
	}
}

func colorizeR40Score(text string, score float64, styled bool) string {
	if !styled || text == "--" {
		return text
	}
	return r40ScoreStyle(score).Render(text)
}

func colorizeMarketChange(text string, move float64, styled bool) string {
	if !styled || text == "--" {
		return text
	}
	return marketMoveStyle(move).Render(text)
}
