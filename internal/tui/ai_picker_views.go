package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderAILeft(section, muted lipgloss.Style, width, height int) string {
	if m.aiPickerOpen {
		return clipLines(m.renderAIPickerGuide(section, muted, width), height)
	}
	return clipLines(m.renderAISetupBlock(section, muted, width), height)
}

func (m Model) renderAIPickerGuide(section, muted lipgloss.Style, width int) string {
	var b strings.Builder
	b.WriteString(section.Render("GUIDE") + "\n\n")
	switch m.aiPickerStep {
	case aiPickerStepConnector:
		b.WriteString(renderWrappedTextBlock(muted, "Step 1 of 2", width) + "\n\n")
		b.WriteString(renderWrappedTextBlock(muted, "Choose the AI provider with ↑/↓. Press Enter to continue to model selection.", width) + "\n\n")
		b.WriteString(renderWrappedTextBlock(muted, "Esc closes setup without changing the chat view.", width))
	case aiPickerStepModel:
		b.WriteString(renderWrappedTextBlock(muted, "Step 2 of 2", width) + "\n\n")
		b.WriteString(renderWrappedTextBlock(muted, "Choose a model for the selected provider with ↑/↓. Press Enter to confirm and return to chat.", width) + "\n\n")
		b.WriteString(renderWrappedTextBlock(muted, "Left arrow goes back to provider selection.", width))
	}
	return strings.TrimRight(b.String(), "\n")
}

func (m Model) renderAISetupBlock(section, muted lipgloss.Style, width int) string {
	var b strings.Builder
	b.WriteString(section.Render("SETUP") + "\n\n")
	b.WriteString("Connector\n")
	b.WriteString(muted.Render(truncateText(m.activeAIConnectorLabel(), width)) + "\n\n")
	b.WriteString("Model\n")
	b.WriteString(muted.Render(truncateText(valueOrDash(strings.TrimSpace(m.config.AIModel)), width)) + "\n\n")
	b.WriteString(section.Render("STATE") + "\n\n")
	b.WriteString(truncateText(fmt.Sprintf("Messages  %d", len(m.aiMessages)), width) + "\n")
	if m.aiCompactedMessages > 0 {
		b.WriteString(truncateText(fmt.Sprintf("Summary   %d compacted", m.aiCompactedMessages), width) + "\n")
	}
	b.WriteString(truncateText(fmt.Sprintf("Symbol    %s", m.activeSymbol()), width) + "\n")
	b.WriteString(truncateText(fmt.Sprintf("Context   %s", m.aiContextStatusLine()), width) + "\n")
	if label := m.aiFilingProgressLabel(); label != "" {
		b.WriteString(truncateText(fmt.Sprintf("Filing    %s", label), width) + "\n")
	}
	if m.aiLastRequestTruncation.hasAny() {
		b.WriteString(truncateText(fmt.Sprintf("Request   %s", m.aiLastRequestTruncation.summaryLine()), width) + "\n")
	}
	if m.aiRunning {
		b.WriteString(truncateText("Running   yes", width) + "\n")
	}
	for _, line := range []string{
		". focus/send in AI",
		"c open config",
		"↑/↓ scroll chat",
		"x clear",
	} {
		b.WriteString("\n" + muted.Render(truncateText(line, width)))
	}
	return strings.TrimRight(b.String(), "\n")
}

func (m Model) renderAIPicker(section, muted lipgloss.Style, width, height int) string {
	var b strings.Builder
	if m.aiPickerStep == aiPickerStepConnector {
		b.WriteString(section.Render("AI PROVIDER") + "\n\n")
		items := m.services.ListAIConnectors()
		if len(items) == 0 {
			b.WriteString(renderWrappedTextBlock(muted, "No AI providers available.", width))
			return clipLines(strings.TrimRight(b.String(), "\n"), height)
		}
		selected := m.activeAIConnectorID()
		for _, item := range items {
			prefix := "  "
			if strings.EqualFold(item.ID, selected) {
				prefix = "▶ "
			}
			b.WriteString(truncateText(prefix+item.Label, width) + "\n")
		}
		b.WriteString("\n" + renderWrappedTextBlock(muted, "↑/↓ move • Enter continue • Esc close", width))
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}

	b.WriteString(section.Render("AI MODEL") + "\n\n")
	b.WriteString(truncateText("Provider: "+m.activeAIConnectorLabel(), width) + "\n")
	if m.aiModelBusy {
		b.WriteString("\n" + renderWrappedTextBlock(muted, "Loading models from local connector...", width))
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}
	if m.aiModelErr != nil {
		b.WriteString("\n" + renderWrappedTextBlock(muted, "Model discovery: "+m.aiModelErr.Error(), width))
	}
	models := m.currentAIModels()
	if len(models) == 0 {
		b.WriteString("\n" + renderWrappedTextBlock(muted, "No discovered models. Connector may not expose model listing; saved/default model will be used.", width))
		b.WriteString("\n\n" + renderWrappedTextBlock(muted, "Enter confirm • ← back • Esc close", width))
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}
	b.WriteString("\n")
	selected := strings.TrimSpace(m.config.AIModel)
	limit := min(len(models), max(3, height-6))
	start := 0
	if m.aiModelIdx >= limit {
		start = m.aiModelIdx - limit + 1
	}
	for i := start; i < start+limit && i < len(models); i++ {
		prefix := "  "
		if strings.EqualFold(models[i], selected) {
			prefix = "▶ "
		}
		b.WriteString(truncateText(prefix+models[i], width) + "\n")
	}
	b.WriteString("\n" + renderWrappedTextBlock(muted, "↑/↓ move • Enter confirm • ← back • Esc close", width))
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func (m Model) renderAIPickerCenter(section, label, muted lipgloss.Style, width, height int) string {
	var b strings.Builder
	if m.aiPickerStep == aiPickerStepConnector {
		b.WriteString(section.Render("AI SETUP") + "\n\n")
		b.WriteString(renderWrappedTextBlock(muted, "Select a provider first. After you confirm, Blackdesk will show the available models for that provider.", width))
		b.WriteString("\n\n")
	} else {
		b.WriteString(section.Render("AI SETUP") + "\n\n")
		b.WriteString(renderWrappedTextBlock(muted, "Select a model for the chosen provider. Confirming returns you to the chat.", width))
		b.WriteString("\n\n")
	}
	pickerHeight := max(6, height-lipgloss.Height(strings.TrimRight(b.String(), "\n"))-2)
	b.WriteString(m.renderAIPicker(section, muted, width, pickerHeight))
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}
