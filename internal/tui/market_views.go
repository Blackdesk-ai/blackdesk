package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderMarketCenter(header, section, label, muted, pos, neg lipgloss.Style, width, height int) string {
	var b strings.Builder
	b.WriteString(header.Render("Global Market Board") + "\n")
	b.WriteString(muted.Render("Cross-asset pulse across US equities, futures, credit, commodities, FX, and regions") + "\n\n")
	b.WriteString(renderMarketCenterGrid(section, label, muted, m, width))
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func (m Model) renderMarketLeft(section, label, muted lipgloss.Style, width, height int) string {
	var b strings.Builder
	b.WriteString(section.Render("GLOBAL SNAPSHOT") + "\n\n")
	b.WriteString(renderMarketSnapshotLine(label, "Regime", marketRiskLine(m.marketRisk), width) + "\n")
	b.WriteString(renderMarketSnapshotLine(label, "Breadth", colorizeBreadthLine(marketBreadthLine(m)), width) + "\n")
	b.WriteString(renderMarketSnapshotLine(label, "Pressure", marketPressureLine(m), width) + "\n")
	b.WriteString(renderMarketSnapshotLine(label, "Focus", focusAssetLine(m), width) + "\n")
	b.WriteString("\n" + section.Render("VOLATILITY") + "\n\n")
	b.WriteString(muted.Render(renderMarketTableHeaderWithValueLabel(width, "Level", label)) + "\n")
	for _, row := range marketBoardRows(m, marketVolBoard) {
		b.WriteString(renderMarketTableRow(row, width, label) + "\n")
	}
	b.WriteString("\n" + section.Render("AI INSIGHT") + " " + muted.Render("(i)") + "\n\n")
	b.WriteString(m.renderMarketOpinionBlock(muted, width))
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func (m Model) renderMarketOpinionBlock(muted lipgloss.Style, width int) string {
	bodyWidth := max(18, width)
	switch {
	case m.aiMarketOpinionRunning:
		return renderWrappedTextBlock(muted, aiTypingFrame(m.clock), bodyWidth)
	case strings.TrimSpace(m.aiMarketOpinion) != "":
		text := renderWrappedTextBlock(lipgloss.NewStyle(), m.aiMarketOpinion, bodyWidth)
		if !m.aiMarketOpinionUpdated.IsZero() {
			text += "\n\n" + muted.Render("Updated "+m.aiMarketOpinionUpdated.Local().Format("15:04"))
		}
		return text
	case m.aiMarketOpinionErr != nil:
		return renderWrappedTextBlock(muted, "Unavailable: "+m.aiMarketOpinionErr.Error(), bodyWidth)
	case m.hasSufficientMarketOpinionData():
		return renderWrappedTextBlock(muted, "Press i to generate AI insight.", bodyWidth)
	default:
		return renderWrappedTextBlock(muted, "Waiting for market board data.", bodyWidth)
	}
}

func (m Model) renderMarketRight(section, label, muted lipgloss.Style, width, height int) string {
	var b strings.Builder
	b.WriteString(section.Render("CROSS-ASSET PULSE") + "\n\n")
	b.WriteString(renderHeatMeter("US Equity", marketBasketScore(m, marketUSBoard), 10) + "\n")
	b.WriteString(renderHeatMeter("Credit", marketBasketScore(m, marketRatesBoard), 10) + "\n")
	b.WriteString(renderHeatMeter("Commods", marketBasketScore(m, marketMacroBoard[:3]), 10) + "\n")
	b.WriteString(renderHeatMeter("Regions", marketBasketScore(m, marketRegionBoard), 10) + "\n\n")
	b.WriteString("\n" + section.Render("DESK") + "\n\n")
	b.WriteString(fmt.Sprintf("%s %s\n", label.Render("US lead"), bestMarketMoveLine(m, marketUSBoard)))
	b.WriteString(fmt.Sprintf("%s %s\n", label.Render("US lag"), worstMarketMoveLine(m, marketUSBoard)))
	b.WriteString(fmt.Sprintf("%s %s\n", label.Render("Best sector"), bestMarketMoveLine(m, marketSectorBoard)))
	b.WriteString(fmt.Sprintf("%s %s\n", label.Render("Worst sector"), worstMarketMoveLine(m, marketSectorBoard)))
	b.WriteString(fmt.Sprintf("%s %s\n", label.Render("Region"), bestMarketMoveLine(m, marketRegionBoard)))
	b.WriteString(fmt.Sprintf("%s %s\n", label.Render("Macro"), bestMarketMoveLine(m, marketMacroBoard)))
	b.WriteString(fmt.Sprintf("%s %s\n", label.Render("2Y/10Y spread"), curve2s10sLine(m)))
	b.WriteString(fmt.Sprintf("%s %s\n", label.Render("Vol curve"), volatilityCurveLine(m)))
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func (m Model) renderMarketBottom(section, label, muted lipgloss.Style, width, height int) string {
	gap := 2
	colWidth := max(20, (width-gap*3)/4)
	leftSectors := marketSectorBoard[:len(marketSectorBoard)/2]
	rightSectors := marketSectorBoard[len(marketSectorBoard)/2:]
	panels := []string{
		lipgloss.NewStyle().Width(colWidth).MaxWidth(colWidth).Height(height).Render(m.renderMarketSimpleTablePanel(section, label, muted, colWidth, height, "REGIONS", marketBoardRows(m, marketRegionBoard))),
		lipgloss.NewStyle().Width(colWidth).MaxWidth(colWidth).Height(height).Render(m.renderMarketSimpleTablePanel(section, label, muted, colWidth, height, "COUNTRIES", marketBoardRows(m, marketCountryBoard))),
		lipgloss.NewStyle().Width(colWidth).MaxWidth(colWidth).Height(height).Render(m.renderMarketSimpleTablePanel(section, label, muted, colWidth, height, "SECTORS", marketBoardRows(m, leftSectors))),
		lipgloss.NewStyle().Width(colWidth).MaxWidth(colWidth).Height(height).Render(m.renderMarketSimpleTablePanel(section, label, muted, colWidth, height, "SECTORS", marketBoardRows(m, rightSectors))),
	}
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		panels[0],
		strings.Repeat(" ", gap),
		panels[1],
		strings.Repeat(" ", gap),
		panels[2],
		strings.Repeat(" ", gap),
		panels[3],
	)
}

func (m Model) renderMarketSimpleTablePanel(section, label, muted lipgloss.Style, width, height int, title string, rows []marketTableRow) string {
	var b strings.Builder
	b.WriteString(section.Render(title) + "\n\n")
	if len(rows) == 0 {
		b.WriteString(muted.Render("Market board still loading"))
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}
	b.WriteString(muted.Render(renderMarketTableHeader(width, label)) + "\n")
	for _, row := range rows {
		b.WriteString(renderMarketTableRow(row, width, label) + "\n")
	}
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func renderMarketCenterGrid(section, label, muted lipgloss.Style, m Model, width int) string {
	gap := 2
	colWidth := max(24, (width-gap)/2)
	leftCol := strings.Join([]string{
		renderMarketBoardCard(section, label, muted, m, colWidth, marketSectionBlock{title: "INDEX FUTURES", valueLabel: "Level", items: marketFuturesBoard}),
		renderMarketBoardCard(section, label, muted, m, colWidth, marketSectionBlock{title: "FX", valueLabel: "Rate", items: marketFXBoard}),
	}, "\n\n")
	rightCol := strings.Join([]string{
		renderMarketBoardCard(section, label, muted, m, colWidth, marketSectionBlock{title: "YIELDS", valueLabel: "Yield", items: marketYieldBoard}),
		renderMarketBoardCard(section, label, muted, m, colWidth, marketSectionBlock{title: "CREDIT", valueLabel: "Level", items: marketRatesBoard}),
		renderMarketBoardCard(section, label, muted, m, colWidth, marketSectionBlock{title: "COMMODITIES", valueLabel: "Level", items: marketMacroBoard}),
	}, "\n\n")
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(colWidth).MaxWidth(colWidth).Render(leftCol),
		strings.Repeat(" ", gap),
		lipgloss.NewStyle().Width(colWidth).MaxWidth(colWidth).Render(rightCol),
	)
}

func renderMarketBoardCard(section, label, muted lipgloss.Style, m Model, width int, block marketSectionBlock) string {
	var b strings.Builder
	b.WriteString(section.Render(block.title) + "\n\n")
	b.WriteString(muted.Render(renderMarketBoardHeader(width, block.valueLabel, label)) + "\n")
	b.WriteString(renderMarketBoard(m, block.items, width, label))
	return b.String()
}
