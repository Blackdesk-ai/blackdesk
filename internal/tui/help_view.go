package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func renderHelpOverlay(section, label, muted lipgloss.Style, width, height int) string {
	type helpEntry struct {
		key  string
		desc string
	}
	type helpSection struct {
		title   string
		entries []helpEntry
	}

	sections := []helpSection{
		{"GLOBAL", []helpEntry{
			{"?", "Toggle this help page"},
			{"/", "Open symbol search"},
			{"Ctrl+K", "Open command palette"},
			{".", "Focus AI composer"},
			{"Tab", "Cycle tabs"},
			{"1-5", "Jump to tab"},
			{"q", "Quit"},
		}},
		{"SCREENER", []helpEntry{
			{"↑ / ↓", "Navigate screener results"},
			{"← / →", "Change screener preset"},
			{"n / p", "Next / previous screener"},
			{"a", "Add selected symbol to watchlist"},
			{"Enter", "Open selected symbol in Quote"},
			{"r", "Refresh screener"},
		}},
		{"SEARCH", []helpEntry{
			{"Enter", "Submit query or select result"},
			{"↑ / ↓", "Navigate results"},
			{"Ctrl+A", "Add symbol to watchlist"},
			{"Esc", "Close search"},
		}},
		{"COMMANDS", []helpEntry{
			{"Ctrl+K", "Open command palette"},
			{"↑ / ↓", "Navigate matches"},
			{"Enter", "Open selected function or symbol"},
			{"Esc", "Close command palette"},
		}},
		{"MARKETS", []helpEntry{
			{"i", "Generate AI market insight"},
			{"r", "Refresh market data"},
		}},
		{"QUOTE", []helpEntry{
			{"↑ / ↓", "Navigate watchlist symbols"},
			{"c", "Chart view"},
			{"f", "Fundamentals view"},
			{"t", "Technicals view"},
			{"s", "Statements view"},
			{"h", "Insiders view"},
			{"← / →", "Change timeframe / statement kind"},
			{"[ / ]", "Change statement frequency"},
			{"n", "Next news story"},
			{"p", "Scroll company description"},
			{"o", "Open news URL in browser"},
			{"d", "Delete symbol from watchlist"},
			{"i", "Generate AI quote insight"},
			{"r", "Refresh symbol data"},
		}},
		{"NEWS", []helpEntry{
			{"↑ / ↓", "Navigate stories"},
			{"n / p", "Next / previous story"},
			{"o", "Open story in browser"},
			{"r", "Refresh news feed"},
		}},
		{"AI", []helpEntry{
			{".", "Focus input / send prompt"},
			{"c", "Open connector & model picker"},
			{"↑ / ↓", "Scroll transcript"},
			{"f", "Toggle fullscreen"},
			{"r", "Re-run with fresh context"},
			{"x", "Clear conversation"},
		}},
		{"AI PICKER", []helpEntry{
			{"↑ / ↓", "Cycle connectors / models"},
			{"← / →", "Switch connector & model step"},
			{"Enter", "Confirm selection"},
			{"Esc  .", "Close picker"},
		}},
	}

	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#E7B66B")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#D8C9B8"))
	titleStyle := section

	colWidth := max(32, (width-4)/2)
	keyColW := 14

	var b strings.Builder
	b.WriteString(titleStyle.Render("KEYBOARD SHORTCUTS") + "\n\n")

	col1 := make([]string, 0, 40)
	col2 := make([]string, 0, 40)
	target := &col1
	for i, sec := range sections {
		if i == 5 {
			target = &col2
		}
		*target = append(*target, label.Render(sec.title))
		for _, e := range sec.entries {
			*target = append(*target, renderHelpEntryLine(keyStyle, descStyle, e.key, e.desc, keyColW))
		}
		*target = append(*target, "")
	}

	for len(col1) < len(col2) {
		col1 = append(col1, "")
	}
	for len(col2) < len(col1) {
		col2 = append(col2, "")
	}

	var out strings.Builder
	out.WriteString(b.String())
	for i := range col1 {
		left := lipgloss.NewStyle().Width(colWidth).Render(col1[i])
		right := col2[i]
		out.WriteString(left + right + "\n")
	}

	return clipLines(strings.TrimRight(out.String(), "\n"), height)
}

func renderHelpEntryLine(keyStyle, descStyle lipgloss.Style, key, desc string, keyColW int) string {
	keyPad := key + strings.Repeat(" ", max(0, keyColW-lipgloss.Width(key)))
	return keyStyle.Render(keyPad) + descStyle.Render(desc)
}
