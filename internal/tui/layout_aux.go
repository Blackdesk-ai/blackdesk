package tui

import (
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"blackdesk/internal/ui"
)

func (m Model) renderOverviewBottom(section, muted lipgloss.Style, width, height int) string {
	newsWidth := max(28, (width-2)*2/3)
	profileWidth := max(20, width-2-newsWidth)
	if newsWidth+profileWidth > width {
		profileWidth = max(20, width-newsWidth)
	}

	newsPanel := lipgloss.NewStyle().
		Width(newsWidth).
		MaxWidth(newsWidth).
		Height(height).
		Render(m.renderNewsPanel(section, muted, newsWidth, height))
	profilePanel := lipgloss.NewStyle().
		Width(profileWidth).
		MaxWidth(profileWidth).
		Height(height).
		Render(m.renderProfilePanel(section, muted, profileWidth, height))
	return lipgloss.JoinHorizontal(lipgloss.Top, newsPanel, "  ", profilePanel)
}

func (m Model) renderNewsPanel(section, muted lipgloss.Style, width, height int) string {
	var b strings.Builder
	b.WriteString(section.Render("NEWS") + " " + muted.Render("(n)") + "\n\n")
	if len(m.news) == 0 {
		if m.errNews != nil {
			b.WriteString(muted.Render("News unavailable: " + m.errNews.Error()))
			return b.String()
		}
		b.WriteString(muted.Render("No news loaded"))
		return b.String()
	}
	maxItems := max(1, (height-2)/2)
	startItem := 0
	if maxItems > 0 && m.newsSelected >= maxItems {
		startItem = m.newsSelected - maxItems + 1
	}
	titleStyle := lipgloss.NewStyle().Width(max(16, width-2)).MaxWidth(max(16, width-2))
	for i := startItem; i < len(m.news) && i < startItem+maxItems; i++ {
		item := m.news[i]
		prefix := "  "
		if i == m.newsSelected {
			prefix = "> "
		}
		b.WriteString(titleStyle.Render(prefix+item.Title) + "\n")
		b.WriteString(muted.Render("  "+item.Publisher+" • "+ui.FormatTimestamp(item.Time)) + "\n")
	}
	if m.errNews != nil {
		b.WriteString(muted.Render("News may be stale: " + m.errNews.Error()))
	}
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func (m Model) renderProfilePanel(section, muted lipgloss.Style, width, height int) string {
	var b strings.Builder
	b.WriteString(section.Render("PROFILE") + " " + muted.Render("(p)") + "\n\n")
	if strings.TrimSpace(m.fundamentals.Description) == "" {
		if m.errFundamentals != nil {
			b.WriteString(muted.Render("Profile unavailable: " + m.errFundamentals.Error()))
			return b.String()
		}
		b.WriteString(muted.Render("No company profile loaded"))
		return b.String()
	}

	textStyle := lipgloss.NewStyle().Width(max(18, width)).MaxWidth(max(18, width))
	profileText := textStyle.Render(m.fundamentals.Description)
	if m.errFundamentals != nil {
		profileText += "\n\n" + muted.Render("Profile may be stale: "+m.errFundamentals.Error())
	}

	indicatorLines := 0
	if m.profileScroll > 0 {
		indicatorLines = 1
	}
	bodyHeight := max(1, height-2-indicatorLines)
	wrappedLines := splitLines(profileText)
	maxOffset := max(0, len(wrappedLines)-bodyHeight)
	offset := min(max(0, m.profileScroll), maxOffset)
	if offset > 0 {
		b.WriteString(muted.Render("↑ more") + "\n")
	}

	visible := slices.Clone(wrappedLines[offset:min(len(wrappedLines), offset+bodyHeight)])
	if len(visible) < bodyHeight {
		visible = append(visible, make([]string, bodyHeight-len(visible))...)
	}
	if maxOffset > 0 && offset < maxOffset && bodyHeight > 0 {
		visible[len(visible)-1] = muted.Render("↓ more")
	}
	b.WriteString(strings.Join(visible, "\n"))
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func (m Model) renderSearchPanel(section, muted lipgloss.Style, height int) string {
	var b strings.Builder
	b.WriteString(section.Render("SEARCH RESULTS") + "\n\n")
	if len(m.searchItems) == 0 {
		b.WriteString(muted.Render("Type a symbol/company in the status bar, Enter to search, Esc to close"))
		return b.String()
	}

	maxItems := max(3, (height-1)/2)
	for i, item := range m.searchItems {
		if i >= maxItems {
			break
		}
		prefix := "  "
		if i == m.searchIdx {
			prefix = "> "
		}
		b.WriteString(prefix + item.Symbol)
		if item.Name != "" {
			b.WriteString(" | " + item.Name)
		}
		b.WriteString("\n")
		metaParts := make([]string, 0, 2)
		if item.Exchange != "" {
			metaParts = append(metaParts, item.Exchange)
		}
		if item.Type != "" {
			metaParts = append(metaParts, item.Type)
		}
		if len(metaParts) > 0 {
			b.WriteString(muted.Render("  "+strings.Join(metaParts, " • ")) + "\n")
		}
	}
	return strings.TrimRight(b.String(), "\n")
}
