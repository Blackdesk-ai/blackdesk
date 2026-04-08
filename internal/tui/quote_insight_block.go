package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderQuoteInsightBlock(muted lipgloss.Style, width int) string {
	bodyWidth := max(18, width)
	activeSymbol := strings.ToUpper(m.activeSymbol())
	switch {
	case m.aiQuoteInsightRunning && strings.EqualFold(m.aiQuoteInsightSymbol, activeSymbol):
		return renderWrappedTextBlock(muted, aiTypingFrame(m.clock), bodyWidth)
	case strings.EqualFold(m.aiQuoteInsightSymbol, activeSymbol) && strings.TrimSpace(m.aiQuoteInsight) != "":
		text := renderWrappedTextBlock(lipgloss.NewStyle(), m.aiQuoteInsight, bodyWidth)
		if !m.aiQuoteInsightUpdated.IsZero() {
			text += "\n\n" + muted.Render("Updated "+m.aiQuoteInsightUpdated.Local().Format("15:04"))
		}
		return text
	case strings.EqualFold(m.aiQuoteInsightSymbol, activeSymbol) && m.aiQuoteInsightErr != nil:
		return renderWrappedTextBlock(muted, "Unavailable: "+m.aiQuoteInsightErr.Error(), bodyWidth)
	default:
		return renderWrappedTextBlock(muted, "Press i to generate AI insight.", bodyWidth)
	}
}
