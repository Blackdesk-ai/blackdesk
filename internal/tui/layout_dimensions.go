package tui

import "github.com/charmbracelet/lipgloss"

func (m Model) bottomPanelProfileWidth() int {
	width := max(20, m.width-lipgloss.NewStyle().Padding(0, 1, 0, 1).GetHorizontalFrameSize()-lipgloss.NewStyle().Border(lipgloss.ThickBorder()).Padding(0, 1).GetHorizontalFrameSize())
	newsWidth := max(28, (width-2)*2/3)
	profileWidth := max(20, width-2-newsWidth)
	if newsWidth+profileWidth > width {
		profileWidth = max(20, width-newsWidth)
	}
	return profileWidth
}

func (m Model) watchlistVisibleRows() int {
	height := m.leftPanelContentHeight()
	listLines := max(1, height-3)
	if len(m.config.Watchlist) <= listLines {
		return len(m.config.Watchlist)
	}
	return max(1, listLines-2)
}

func (m Model) marketNewsVisibleRows() int {
	height := m.mainPanelHeight()
	listLines := max(1, height-2)
	if len(m.marketNews) <= listLines {
		return len(m.marketNews)
	}
	return max(1, listLines-2)
}

func (m Model) leftPanelContentHeight() int {
	topSpacerH := 1
	topH := 1
	statusH := 1
	lineH := 1
	availableH := max(16, m.height-topSpacerH-topH-statusH-lineH)
	mainTotalHeight := max(12, int(float64(availableH)*0.68))
	frameY := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 2).GetVerticalFrameSize()
	mainHeight := max(8, mainTotalHeight-frameY)
	return max(1, mainHeight-2)
}

func (m Model) mainPanelHeight() int {
	topH := 1
	statusH := 1
	lineH := 1
	availableH := max(16, m.height-topH-statusH-lineH)
	mainTotalHeight := max(12, int(float64(availableH)*0.64))
	frameY := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 2).GetVerticalFrameSize()
	return max(8, mainTotalHeight-frameY) - 2
}

func (m Model) bottomPanelHeight() int {
	topH := 1
	statusH := 1
	lineH := 1
	availableH := max(16, m.height-topH-statusH-lineH)
	mainTotalHeight := max(12, int(float64(availableH)*0.64))
	bottomTotalHeight := max(8, availableH-mainTotalHeight)
	frameY := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 2).GetVerticalFrameSize()
	return max(4, bottomTotalHeight-frameY) - 2
}

func (m Model) screenerVisibleRows() int {
	height := m.mainPanelHeight()
	listLines := max(1, height-5)
	if len(m.screenerResult.Items) <= listLines {
		return len(m.screenerResult.Items)
	}
	return max(1, listLines-2)
}
