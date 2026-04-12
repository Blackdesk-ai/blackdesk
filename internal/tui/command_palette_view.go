package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func (m Model) renderCommandPalette(section, label, muted lipgloss.Style, width, height int) string {
	bodyHeight := max(8, height-6)
	leftWidth := clamp((width*3)/5, 40, 68)
	rightWidth := max(20, width-leftWidth-3)

	selected, ok := m.currentCommandPaletteItem()

	var b strings.Builder
	b.WriteString(section.Render("COMMAND PALETTE") + "\n\n")
	b.WriteString(m.commandInput.View() + "\n\n")
	b.WriteString(muted.Render(m.commandPaletteSummaryLine()) + "\n\n")

	left := lipgloss.NewStyle().Width(leftWidth).Render(m.renderCommandPaletteResults(section, muted, leftWidth, bodyHeight))
	right := lipgloss.NewStyle().Width(rightWidth).Render(m.renderCommandPalettePreview(section, label, muted, rightWidth, bodyHeight, selected, ok))
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, left, "   ", right))
	return clipLines(b.String(), height)
}

func (m Model) renderCommandPaletteResults(section, muted lipgloss.Style, width, height int) string {
	var b strings.Builder
	b.WriteString(section.Render("RESULTS") + "\n\n")
	if len(m.commandPaletteItems) == 0 {
		b.WriteString(muted.Render("No matching functions or symbols"))
		return clipLines(b.String(), height)
	}

	windowStart := 0
	visibleRows := max(1, height-2)
	if m.commandPaletteIdx >= visibleRows {
		windowStart = m.commandPaletteIdx - visibleRows + 1
	}
	windowEnd := min(len(m.commandPaletteItems), windowStart+visibleRows)

	activeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#14110D")).
		Background(lipgloss.Color("#E7B66B")).
		Bold(true).
		Width(width).
		MaxWidth(width)
	functionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#62D394")).Bold(true)
	symbolStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F3EBDD")).Bold(true)
	metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#9F907E"))

	for i := windowStart; i < windowEnd; i++ {
		item := m.commandPaletteItems[i]
		prefix := "  "
		titleStyle := symbolStyle
		if item.Kind == commandPaletteItemFunction {
			titleStyle = functionStyle
		}
		if i == m.commandPaletteIdx {
			prefix = "› "
			line := renderStatusLine(width, prefix+item.Title, item.Meta)
			b.WriteString(activeStyle.Render(line) + "\n")
			continue
		}
		metaText := strings.TrimSpace(item.Meta)
		metaWidth := lipgloss.Width(metaText)
		if metaText != "" {
			metaWidth = min(metaWidth, max(10, width/3))
		}
		leftWidth := width
		if metaText != "" {
			leftWidth = max(1, width-metaWidth-1)
		}
		left := titleStyle.Render(padRight(ansi.Truncate(prefix+item.Title, leftWidth, "..."), leftWidth))
		if metaText == "" {
			b.WriteString(left + "\n")
			continue
		}
		meta := metaStyle.Render(ansi.Truncate(metaText, metaWidth, "..."))
		spacer := strings.Repeat(" ", max(1, width-lipgloss.Width(left)-lipgloss.Width(meta)))
		b.WriteString(left + spacer + meta + "\n")
	}
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func (m Model) renderCommandPalettePreview(section, label, muted lipgloss.Style, width, height int, item commandPaletteItem, ok bool) string {
	var b strings.Builder
	b.WriteString(section.Render("PREVIEW") + "\n\n")
	if !ok {
		b.WriteString(muted.Render("Use Ctrl+K to open, type to filter, Enter to run, Esc to close."))
		return clipLines(b.String(), height)
	}

	b.WriteString(label.Render(item.Title) + "\n")
	if item.Subtitle != "" {
		b.WriteString(muted.Render(truncateText(item.Subtitle, width)) + "\n")
	}
	if item.Meta != "" {
		b.WriteString(muted.Render(truncateText(item.Meta, width)) + "\n")
	}
	if item.Description != "" {
		b.WriteString("\n" + renderWrappedTextBlock(muted, item.Description, width))
	}

	switch item.Kind {
	case commandPaletteItemFunction:
		if aliases := m.commandPaletteAliases(item.FunctionID); aliases != "" {
			b.WriteString("\n\n" + muted.Render("Aliases") + "\n")
			b.WriteString(renderWrappedTextBlock(muted, aliases, width))
		}
	case commandPaletteItemSymbol:
		if item.Symbol.Exchange != "" || item.Symbol.Type != "" {
			b.WriteString("\n\n" + muted.Render("Instrument") + "\n")
			b.WriteString(renderWrappedTextBlock(muted, strings.TrimSpace(fmt.Sprintf("%s %s", item.Symbol.Exchange, item.Symbol.Type)), width))
		}
	}

	b.WriteString("\n\n" + muted.Render("Enter open • ↑/↓ move • Esc close"))
	return clipLines(b.String(), height)
}

func (m Model) commandPaletteSummaryLine() string {
	functions := 0
	symbols := 0
	for _, item := range m.commandPaletteItems {
		switch item.Kind {
		case commandPaletteItemFunction:
			functions++
		case commandPaletteItemSymbol:
			symbols++
		}
	}
	parts := make([]string, 0, 2)
	if functions > 0 {
		parts = append(parts, fmt.Sprintf("%d functions", functions))
	}
	if symbols > 0 {
		parts = append(parts, fmt.Sprintf("%d symbols", symbols))
	}
	if len(parts) == 0 {
		return "Type a function or symbol"
	}
	return strings.Join(parts, " • ")
}

func (m Model) commandPaletteAliases(id string) string {
	for _, fn := range m.commandPaletteFunctions() {
		if fn.ID == id {
			return strings.Join(fn.Aliases, ", ")
		}
	}
	return ""
}
