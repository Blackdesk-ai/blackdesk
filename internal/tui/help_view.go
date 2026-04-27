package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type helpOverlayEntry struct {
	key  string
	desc string
}

type helpOverlaySection struct {
	title   string
	entries []helpOverlayEntry
}

func renderHelpOverlay(section, label, muted lipgloss.Style, width, height int) string {
	sections := []helpOverlaySection{
		{"GLOBAL", []helpOverlayEntry{
			{"?", "Toggle this help page"},
			{"/", "Open symbol search"},
			{"Ctrl+K", "Open command palette"},
			{"Ctrl+⌫", "Back to previous page"},
			{"Tab", "Cycle tabs"},
			{"1-5", "Jump to tab"},
			{"q", "Quit"},
		}},
		{"SCREENER", []helpOverlayEntry{
			{"↑ / ↓", "Navigate screener results"},
			{"← / →", "Change screener preset"},
			{"n / p", "Next / previous preset"},
			{"a", "Add symbol to watchlist"},
			{"Enter", "Open selected symbol in Quote"},
			{"r", "Refresh screener"},
		}},
		{"SEARCH", []helpOverlayEntry{
			{"Enter", "Open query or selected result"},
			{"↑ / ↓", "Navigate results"},
			{"Ctrl+A", "Add symbol to watchlist"},
			{"Esc", "Close search"},
		}},
		{"COMMANDS", []helpOverlayEntry{
			{"Ctrl+K", "Open command palette"},
			{"↑ / ↓", "Navigate matches"},
			{"Enter", "Open selected item"},
			{"Esc", "Close command palette"},
		}},
		{"MARKETS", []helpOverlayEntry{
			{"i", "Generate AI market insight"},
			{"r", "Refresh market data"},
		}},
		{"CALENDAR", []helpOverlayEntry{
			{"← / →", "Switch Today / This Week"},
			{"↑ / ↓", "Navigate events"},
			{"r", "Refresh economic calendar"},
			{"Esc", "Close calendar"},
		}},
		{"QUOTE", []helpOverlayEntry{
			{"↑ / ↓", "Navigate watchlist symbols"},
			{"c", "Chart view"},
			{"f", "Fundamentals view"},
			{"t", "Technicals view"},
			{"s", "Statements view"},
			{"h", "Insiders view"},
			{"← / →", "Chart / RA / Stats / Stmts / Filings"},
			{"[ / ]", "Change statement frequency"},
			{"n", "Next news story"},
			{"p", "Scroll company description"},
			{"o", "Open news URL in browser"},
			{"d", "Delete symbol from watchlist"},
			{"i", "Generate AI quote insight"},
			{"r", "Refresh symbol data"},
		}},
		{"NEWS", []helpOverlayEntry{
			{"↑ / ↓", "Navigate stories"},
			{"n / p", "Next / previous story"},
			{"o", "Open story in browser"},
			{"r", "Refresh news feed"},
		}},
		{"AI", []helpOverlayEntry{
			{".", "Focus input or send prompt"},
			{"Type", "Start typing to prompt AI"},
			{"Enter", "Send prompt when input is focused"},
			{"c", "Open AI config"},
			{"↑ / ↓", "Scroll transcript"},
			{"f", "Toggle fullscreen"},
			{"x", "Clear conversation"},
		}},
		{"AI PICKER", []helpOverlayEntry{
			{"↑ / ↓", "Cycle connectors / models"},
			{"← / →", "Switch connector & model step"},
			{"Enter", "Confirm selection"},
			{"Esc  .", "Close picker"},
		}},
	}

	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#E7B66B")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#D8C9B8"))
	titleStyle := section

	bodyHeight := max(1, height-2)
	columnGap := 3
	columnsCount := chooseHelpColumnCount(sections, width, bodyHeight, columnGap)
	colWidth := max(28, (width-(columnGap*(columnsCount-1)))/columnsCount)
	keyColW := 12
	sectionSpacing := 1
	if columnsCount >= 3 {
		sectionSpacing = 0
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("KEYBOARD SHORTCUTS") + "\n\n")

	columns := make([][]string, columnsCount)
	columnHeights := make([]int, columnsCount)
	for _, sec := range sections {
		target := minIndex(columnHeights)
		columns[target] = append(columns[target], label.Render(sec.title))
		for _, e := range sec.entries {
			columns[target] = append(columns[target], renderHelpEntryLine(keyStyle, descStyle, e.key, e.desc, keyColW))
		}
		if sectionSpacing > 0 {
			columns[target] = append(columns[target], "")
		}
		columnHeights[target] += helpSectionHeight(sec, sectionSpacing)
	}

	maxLines := 0
	for _, col := range columns {
		if len(col) > maxLines {
			maxLines = len(col)
		}
	}
	for i := range columns {
		for len(columns[i]) < maxLines {
			columns[i] = append(columns[i], "")
		}
	}

	columnStyle := lipgloss.NewStyle().Width(colWidth)
	gap := strings.Repeat(" ", columnGap)

	var out strings.Builder
	out.WriteString(b.String())
	for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
		for colIdx := 0; colIdx < columnsCount; colIdx++ {
			out.WriteString(columnStyle.Render(columns[colIdx][lineIdx]))
			if colIdx < columnsCount-1 {
				out.WriteString(gap)
			}
		}
		out.WriteString("\n")
	}

	return clipLines(strings.TrimRight(out.String(), "\n"), height)
}

func renderHelpEntryLine(keyStyle, descStyle lipgloss.Style, key, desc string, keyColW int) string {
	keyPad := key + strings.Repeat(" ", max(0, keyColW-lipgloss.Width(key)))
	return keyStyle.Render(keyPad) + descStyle.Render(desc)
}

func helpSectionHeight(sec helpOverlaySection, spacing int) int {
	return 1 + len(sec.entries) + spacing
}

func chooseHelpColumnCount(sections []helpOverlaySection, width, bodyHeight, gap int) int {
	const minColWidth = 38
	maxCols := max(1, min(3, (width+gap)/(minColWidth+gap)))
	if maxCols == 1 {
		return 1
	}

	for cols := 1; cols <= maxCols; cols++ {
		colWidth := (width - (gap * (cols - 1))) / cols
		if colWidth < minColWidth {
			continue
		}
		heights := make([]int, cols)
		spacing := 1
		if cols >= 3 {
			spacing = 0
		}
		for _, sec := range sections {
			heights[minIndex(heights)] += helpSectionHeight(sec, spacing)
		}
		if maxInt(heights) <= bodyHeight {
			return cols
		}
	}
	return maxCols
}

func minIndex(values []int) int {
	if len(values) == 0 {
		return 0
	}
	bestIdx := 0
	bestVal := values[0]
	for i := 1; i < len(values); i++ {
		if values[i] < bestVal {
			bestIdx = i
			bestVal = values[i]
		}
	}
	return bestIdx
}

func maxInt(values []int) int {
	best := 0
	for _, value := range values {
		if value > best {
			best = value
		}
	}
	return best
}
