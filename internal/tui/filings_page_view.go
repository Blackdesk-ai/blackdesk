package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"blackdesk/internal/domain"
)

func (m Model) renderQuoteFilingsPage(header, section, label, muted, pos, neg lipgloss.Style, width, height int) string {
	snapshot := m.filteredFilingsSnapshot(m.activeSymbol())
	listWidth := clamp((width*3)/5, 46, 76)
	previewWidth := max(24, width-listWidth-3)
	bodyHeight := max(8, height-6)

	var b strings.Builder
	b.WriteString(section.Render("FILINGS") + "\n\n")
	b.WriteString(m.renderQuoteFilingsSummary(header, label, muted, pos, neg, width) + "\n\n")
	left := lipgloss.NewStyle().Width(listWidth).Render(m.renderQuoteFilingsList(section, muted, listWidth, bodyHeight, snapshot))
	right := lipgloss.NewStyle().Width(previewWidth).Render(m.renderQuoteFilingsPreview(section, label, muted, previewWidth, bodyHeight, snapshot))
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, left, "   ", right))
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func (m Model) renderQuoteFilingsSummary(header, label, muted, pos, neg lipgloss.Style, width int) string {
	quote := m.quote
	snapshot := m.filingsForSymbol(m.activeSymbol())
	company := snapshot.CompanyName
	if company == "" {
		company = quote.ShortName
	}
	title := header.Render(strings.ToUpper(m.activeSymbol()))
	if strings.TrimSpace(company) != "" {
		title += muted.Render("  " + company)
	}
	return renderStatusLine(width, title, "")
}

func (m Model) renderQuoteFilingsList(section, muted lipgloss.Style, width, height int, snapshot domain.FilingsSnapshot) string {
	var b strings.Builder
	b.WriteString(renderStatusLine(width, section.Render("RECENT FILINGS"), m.renderFilingsFilterTabs(muted, width)) + "\n\n")
	if len(snapshot.Items) == 0 {
		if m.errFilings != nil {
			b.WriteString(m.errFilings.Error())
		} else {
			b.WriteString(muted.Render("No filings match the current filter for the active symbol"))
		}
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#14110D")).
		Background(lipgloss.Color("#E7B66B")).
		Bold(true).
		Width(width).
		MaxWidth(width)

	dateColWidth := 10
	formColWidth := clamp(width/4, 6, 14)
	titleColWidth := max(12, width-dateColWidth-formColWidth-4)

	headerLine := fmt.Sprintf(
		"%-*s  %-*s  %s",
		dateColWidth, "DATE",
		formColWidth, "FORM",
		truncateText("MEANING", titleColWidth),
	)
	b.WriteString(muted.Render(truncateText(headerLine, width)) + "\n")
	b.WriteString(muted.Render(strings.Repeat("─", max(12, min(width, dateColWidth+formColWidth+titleColWidth+4)))) + "\n")

	visibleRows := max(3, height/2)
	start := 0
	if m.filingsSel >= visibleRows {
		start = m.filingsSel - visibleRows + 1
	}
	end := min(len(snapshot.Items), start+visibleRows)
	for i := start; i < end; i++ {
		item := snapshot.Items[i]
		line := fmt.Sprintf(
			"%-*s  %-*s  %s",
			dateColWidth, filingDateLabel(item),
			formColWidth, truncateText(strings.ToUpper(item.Form), formColWidth),
			truncateText(filingDisplayTitle(item), titleColWidth),
		)
		line = truncateText(line, width)
		if i == m.filingsSel {
			b.WriteString(selectedStyle.Render(line) + "\n")
			continue
		}
		b.WriteString(line + "\n")
	}
	b.WriteString("\n" + muted.Render("←/→ filter • ↑/↓ move • Enter open filing • i analyze in AI"))
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func (m Model) renderQuoteFilingsPreview(section, label, muted lipgloss.Style, width, height int, snapshot domain.FilingsSnapshot) string {
	var b strings.Builder
	b.WriteString(section.Render("PREVIEW") + "\n\n")
	item, ok := m.currentFiling()
	if !ok {
		b.WriteString(muted.Render("Select a filing to inspect it."))
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}

	b.WriteString(label.Render(filingDisplayTitle(item)) + "\n")
	b.WriteString(muted.Render(strings.ToUpper(item.Form)+" • "+filingDateLabel(item)) + "\n")
	if snapshot.CompanyName != "" {
		b.WriteString(muted.Render(truncateText(snapshot.CompanyName, width)) + "\n")
	}
	b.WriteString("\n" + muted.Render("What This Is") + "\n")
	b.WriteString(renderWrappedTextBlock(muted, filingMeaning(item), width))
	b.WriteString("\n\n" + muted.Render("Document") + "\n")
	b.WriteString(renderWrappedTextBlock(muted, filingDocumentLabel(item), width))
	if !item.ReportDate.IsZero() {
		b.WriteString("\n\n" + muted.Render("Reporting Period") + "\n")
		b.WriteString(renderWrappedTextBlock(muted, item.ReportDate.Format("2006-01-02"), width))
	}
	if item.URL != "" {
		b.WriteString("\n\n" + muted.Render("Open") + "\n")
		b.WriteString(renderWrappedTextBlock(muted, "Press Enter to open the original SEC filing in your browser.", width))
	}
	b.WriteString("\n\n" + muted.Render("AI Analysis") + "\n")
	b.WriteString(renderWrappedTextBlock(muted, "Press i to open AI and generate a structured analysis of this selected filing.", width))
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func (m Model) renderFilingsFilterTabs(muted lipgloss.Style, width int) string {
	active := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#14110D")).
		Background(lipgloss.Color("#E7B66B")).
		Bold(true).
		Padding(0, 1)
	inactive := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#D8C9B8")).
		Background(lipgloss.Color("#2A2520")).
		Padding(0, 1)

	tabs := []struct {
		mode  filingsFilterMode
		label string
	}{
		{mode: filingsFilterAll, label: "ALL"},
		{mode: filingsFilterPeriodicReports, label: "10-K/10-Q"},
		{mode: filingsFilterTenK, label: "10-K"},
		{mode: filingsFilterTenQ, label: "10-Q"},
	}

	parts := make([]string, 0, len(tabs))
	for _, tab := range tabs {
		style := inactive
		if tab.mode == m.filingsFilter {
			style = active
		}
		parts = append(parts, style.Render(tab.label))
	}
	return ansi.Truncate(strings.Join(parts, " "), width, "")
}
