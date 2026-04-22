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

func impliedReturnScoreStyle(score float64) lipgloss.Style {
	style := lipgloss.NewStyle()
	switch {
	case score > 0.05:
		return style.Foreground(lipgloss.Color("#62D394"))
	case score >= 0:
		return style.Foreground(lipgloss.Color("#A6A29D"))
	default:
		return style.Foreground(lipgloss.Color("#FF7A73"))
	}
}

func colorizeImpliedReturnScore(text string, score float64, styled bool) string {
	if !styled || text == "--" {
		return text
	}
	return impliedReturnScoreStyle(score).Render(text)
}

func impliedSharpeScoreStyle(score float64) lipgloss.Style {
	style := lipgloss.NewStyle()
	switch {
	case score > 1:
		return style.Foreground(lipgloss.Color("#62D394")).Bold(true)
	case score >= 0.5:
		return style.Foreground(lipgloss.Color("#62D394"))
	case score >= 0:
		return style.Foreground(lipgloss.Color("#A6A29D"))
	default:
		return style.Foreground(lipgloss.Color("#FF7A73"))
	}
}

func colorizeImpliedSharpeScore(text string, score float64, styled bool) string {
	if !styled || text == "--" {
		return text
	}
	return impliedSharpeScoreStyle(score).Render(text)
}

func colorizeMarketChange(text string, move float64, styled bool) string {
	if !styled || text == "--" {
		return text
	}
	return marketMoveStyle(move).Render(text)
}
