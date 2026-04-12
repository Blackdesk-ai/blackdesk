package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderAICenter(section, muted lipgloss.Style, width, height int) string {
	var b strings.Builder
	b.WriteString(section.Render("CHAT") + "\n\n")
	lines := m.renderAITranscriptLines(width)
	if m.aiRunning {
		assistantStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F3EBDD")).Background(lipgloss.Color("#3A3028")).Padding(0, 1)
		lines = append(lines, assistantStyle.Render("AI")+" "+muted.Render(aiTypingFrame(m.clock)), "")
	}
	if len(lines) == 0 {
		b.WriteString(renderWrappedTextBlock(muted, "No messages yet. Press . or start typing to ask the selected local AI about the current desk context.", width))
		if m.aiErr != nil {
			b.WriteString("\n\n" + renderWrappedTextBlock(muted, "Last error: "+m.aiErr.Error(), width))
		}
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}
	visible := max(1, height-2)
	maxScroll := max(0, len(lines)-visible)
	offset := max(0, maxScroll-min(m.aiScroll, maxScroll))
	for _, line := range lines[offset:min(len(lines), offset+visible)] {
		b.WriteString(line + "\n")
	}
	if len(lines) == 0 {
		b.WriteString(muted.Render("Transcript unavailable"))
	}
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func (m Model) renderAITranscriptLines(width int) []string {
	if len(m.aiMessages) == 0 {
		return nil
	}
	lines := make([]string, 0, len(m.aiMessages)*6)
	userStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#1C1712")).Background(lipgloss.Color("#E7B66B")).Padding(0, 1)
	assistantStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F3EBDD")).Background(lipgloss.Color("#3A3028")).Padding(0, 1)
	bodyWidth := max(24, width-2)
	for i := range m.aiMessages {
		msg := m.aiMessages[i]
		badge := userStyle.Render("You")
		if msg.Role == aiMessageAssistant {
			badge = assistantStyle.Render("AI")
		}
		lines = append(lines, badge)
		wrapped := splitLines(lipgloss.NewStyle().Width(bodyWidth).Render(strings.TrimSpace(msg.Body)))
		if msg.Role == aiMessageAssistant {
			wrapped = renderMarkdownTranscript(msg.Body, bodyWidth)
		}
		lines = append(lines, wrapped...)
		lines = append(lines, "")
	}
	return lines
}

func aiTypingFrame(ts time.Time) string {
	frames := []string{"thinking   ", "thinking.  ", "thinking.. ", "thinking..."}
	if ts.IsZero() {
		return frames[0]
	}
	return frames[(ts.Second())%len(frames)]
}
