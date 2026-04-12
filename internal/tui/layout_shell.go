package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading terminal…"
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F3EBDD"))
	brandStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F3EBDD"))
	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F3EBDD"))
	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#D8C9B8"))
	tabStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#8E7F71"))
	activeTabStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#1C1712")).Background(lipgloss.Color("#E7B66B")).Padding(0, 1)
	frameStyle := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("#4E4033")).
		Padding(0, 1)
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("#9F907E"))
	pos := lipgloss.NewStyle().Foreground(lipgloss.Color("#62D394"))
	neg := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF7A73"))
	rootStyle := lipgloss.NewStyle().Padding(0, 1, 0, 1)

	viewportWidth := max(20, m.width-2)
	headerLeft := brandStyle.Render(brandHeaderWordmark) + "  " + renderHeaderTabs(tabStyle, activeTabStyle, m.tabIdx)
	headerRight := buildHeaderMeta(m, muted, pos, neg, max(8, viewportWidth-lipgloss.Width(headerLeft)-1))
	topSpacer := strings.Repeat(" ", viewportWidth)
	top := renderHeader(headerStyle, viewportWidth, headerLeft, headerRight)

	frameX := frameStyle.GetHorizontalFrameSize()
	frameBX := frameStyle.GetHorizontalBorderSize()
	frameY := frameStyle.GetVerticalFrameSize()
	totalWidth := viewportWidth
	leftTotal := clamp(totalWidth/4, 30, 36)
	rightTotal := clamp(totalWidth/4, 28, 36)
	centerTotal := totalWidth - leftTotal - rightTotal
	if centerTotal < 46 {
		centerTotal = 46
		leftTotal = max(24, totalWidth-centerTotal-rightTotal)
		rightTotal = totalWidth - centerTotal - leftTotal
	}
	centerWidth := max(36, centerTotal-frameX)
	topSpacerH := lipgloss.Height(topSpacer)
	topH := lipgloss.Height(top)
	statusH := 1
	lineH := 1
	availableH := max(16, m.height-topSpacerH-topH-statusH-lineH)
	mainTotalHeight := max(12, int(float64(availableH)*0.68))
	bottomTotalHeight := max(4, availableH-mainTotalHeight)
	fullHeightQuoteCenter := m.tabIdx == tabQuote && (m.quoteCenterMode == quoteCenterStatements || m.quoteCenterMode == quoteCenterInsiders)
	fullscreenQuotePage := m.tabIdx == tabQuote && m.quoteCenterMode == quoteCenterFilings
	fullHeightNews := m.tabIdx == tabNews || m.tabIdx == tabScreener
	if m.commandPaletteOpen || m.tabIdx == tabAI || fullscreenQuotePage || fullHeightQuoteCenter || fullHeightNews {
		mainTotalHeight = availableH
		bottomTotalHeight = 0
	}
	mainHeight := max(8, mainTotalHeight-frameY)
	bottomHeight := max(0, bottomTotalHeight-frameY)

	mainRow := ""
	if m.helpOpen {
		helpH := mainHeight + bottomTotalHeight
		if m.commandPaletteOpen || m.tabIdx == tabAI || fullscreenQuotePage || fullHeightQuoteCenter || fullHeightNews {
			helpH = mainHeight
		}
		mainRow = frameStyle.Width(viewportWidth - frameBX).Height(helpH).Render(
			renderHelpOverlay(sectionStyle, lipgloss.NewStyle().Foreground(lipgloss.Color("245")), muted, viewportWidth-frameX, helpH-2),
		)
		mainRow = lipgloss.NewStyle().Width(viewportWidth).MaxWidth(viewportWidth).Render(mainRow)
	} else if m.commandPaletteOpen {
		mainRow = frameStyle.Width(viewportWidth - frameBX).Height(mainHeight).Render(
			m.renderCommandPalette(sectionStyle, labelStyle, muted, viewportWidth-frameX, mainHeight-2),
		)
	} else if fullscreenQuotePage {
		mainRow = frameStyle.Width(viewportWidth - frameBX).Height(mainHeight).Render(
			m.renderQuoteFilingsPage(headerStyle, sectionStyle, labelStyle, muted, pos, neg, viewportWidth-frameX, mainHeight-2),
		)
	} else if m.tabIdx == tabAI && m.aiFullscreen && !m.aiPickerOpen {
		mainRow = frameStyle.Width(viewportWidth - frameBX).Height(mainHeight).Render(
			m.renderCenterPanel(headerStyle, sectionStyle, labelStyle, muted, pos, neg, viewportWidth-frameX, mainHeight-2),
		)
	} else {
		left := frameStyle.Width(leftTotal - frameBX).Height(mainHeight).Render(m.renderLeftPanel(sectionStyle, muted, pos, neg, leftTotal-frameX, mainHeight-2))
		center := frameStyle.Width(centerTotal - frameBX).Height(mainHeight).Render(m.renderCenterPanel(headerStyle, sectionStyle, labelStyle, muted, pos, neg, centerWidth-4, mainHeight-2))
		right := frameStyle.Width(rightTotal - frameBX).Height(mainHeight).Render(m.renderRightPanel(sectionStyle, labelStyle, muted, rightTotal-frameX, mainHeight-2))
		mainRow = lipgloss.JoinHorizontal(lipgloss.Top, left, center, right)
	}
	mainRow = lipgloss.NewStyle().Width(viewportWidth).MaxWidth(viewportWidth).Render(mainRow)

	bottom := ""
	if !m.helpOpen && !m.commandPaletteOpen && !fullscreenQuotePage && m.tabIdx != tabAI && !fullHeightQuoteCenter && !fullHeightNews {
		bottom = frameStyle.Width(viewportWidth - frameBX).Height(bottomHeight).Render(m.renderBottomPanel(sectionStyle, labelStyle, muted, viewportWidth-frameX, bottomHeight-2))
	}

	statusText := m.statusText()
	lineText := m.status
	if m.helpOpen {
		lineText = muted.Render("Press ? or Esc to close help")
		statusText = ""
	} else if m.commandPaletteOpen {
		lineText = muted.Render("Command palette: Enter open • ↑/↓ move • Esc close")
		statusText = ""
	} else if m.searchMode {
		lineText = m.searchInput.View()
		statusText = ""
	} else if m.aiPickerOpen {
		lineText = muted.Render("AI picker: " + m.activeAIStatusLabel())
		statusText = ""
	} else if m.aiFocused {
		lineText = m.aiInput.View()
		statusText = ""
	} else {
		lineText = muted.Render(lineText)
	}

	status := lipgloss.NewStyle().
		Width(viewportWidth).
		MaxWidth(viewportWidth).
		Render(ansi.Truncate(statusText, viewportWidth, ""))
	line := lipgloss.NewStyle().Width(viewportWidth).MaxWidth(viewportWidth).Render(
		renderStatusLine(viewportWidth, lineText, m.renderStatusMeta(muted, pos.Bold(true))),
	)

	parts := []string{topSpacer, top, mainRow}
	if bottom != "" {
		parts = append(parts, bottom)
	}
	parts = append(parts, line, status)
	content := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return rootStyle.Width(m.width).MaxWidth(m.width).Render(content)
}
