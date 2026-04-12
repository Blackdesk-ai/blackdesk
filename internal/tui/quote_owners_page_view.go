package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"blackdesk/internal/domain"
	"blackdesk/internal/ui"
)

func (m Model) renderQuoteOwnersPage(header, section, label, muted, pos, neg lipgloss.Style, width, height int) string {
	snapshot := m.ownersForSymbol(m.activeSymbol())
	listWidth := clamp((width*3)/5, 48, 86)
	previewWidth := max(24, width-listWidth-3)
	bodyHeight := max(8, height-6)

	var b strings.Builder
	b.WriteString(section.Render("OWNERS") + "\n\n")
	b.WriteString(m.renderQuoteOwnersSummary(header, muted, width, snapshot) + "\n\n")
	left := lipgloss.NewStyle().Width(listWidth).Render(m.renderQuoteOwnersList(section, muted, pos, neg, listWidth, bodyHeight, snapshot))
	right := lipgloss.NewStyle().Width(previewWidth).Render(m.renderQuoteOwnersPreview(section, label, muted, pos, neg, previewWidth, bodyHeight, snapshot))
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, left, "   ", right))
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func (m Model) renderQuoteOwnersSummary(header, muted lipgloss.Style, width int, snapshot domain.OwnershipSnapshot) string {
	company := snapshot.CompanyName
	if company == "" {
		company = m.quote.ShortName
	}
	title := header.Render(strings.ToUpper(m.activeSymbol()))
	if strings.TrimSpace(company) != "" {
		title += muted.Render("  " + company)
	}
	return renderStatusLine(width, title, "")
}

func (m Model) renderQuoteOwnersList(section, muted, pos, neg lipgloss.Style, width, height int, snapshot domain.OwnershipSnapshot) string {
	var b strings.Builder
	b.WriteString(section.Render("TOP HOLDERS") + "\n\n")
	items := m.ownerItemsForSymbol(snapshot.Symbol)
	if len(items) == 0 {
		if m.errOwners != nil {
			b.WriteString(m.errOwners.Error())
		} else {
			b.WriteString(muted.Render("No ownership data loaded for the active symbol"))
		}
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#14110D")).
		Background(lipgloss.Color("#E7B66B")).
		Bold(true).
		Width(width).
		MaxWidth(width)

	typeColWidth := 12
	pctColWidth := 9
	dateColWidth := 10
	holderColWidth := max(12, width-typeColWidth-pctColWidth-dateColWidth-6)

	headerLine := fmt.Sprintf(
		"%-*s  %-*s  %-*s  %s",
		typeColWidth, "TYPE",
		holderColWidth, "HOLDER",
		pctColWidth, "% HELD",
		truncateText("DATE", dateColWidth),
	)
	b.WriteString(muted.Render(truncateText(headerLine, width)) + "\n")
	b.WriteString(muted.Render(strings.Repeat("─", max(12, min(width, typeColWidth+holderColWidth+pctColWidth+dateColWidth+6)))) + "\n")

	visibleRows := max(3, height/2)
	start := 0
	if m.ownersSel >= visibleRows {
		start = m.ownersSel - visibleRows + 1
	}
	end := min(len(items), start+visibleRows)
	for i := start; i < end; i++ {
		item := items[i]
		line := fmt.Sprintf(
			"%-*s  %-*s  %-*s  %s",
			typeColWidth, truncateText(strings.ToUpper(item.kind), typeColWidth),
			holderColWidth, truncateText(item.holder.Name, holderColWidth),
			pctColWidth, truncateText(formatOwnershipPercent(item.holder.PercentHeld), pctColWidth),
			truncateText(ownershipDateLabel(item.holder), dateColWidth),
		)
		line = truncateText(line, width)
		if i == m.ownersSel {
			b.WriteString(selectedStyle.Render(line) + "\n")
			continue
		}
		b.WriteString(renderOwnershipRow(line, item, pos, neg) + "\n")
	}
	b.WriteString("\n" + muted.Render("↑/↓ move • palette to switch views"))
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func (m Model) renderQuoteOwnersPreview(section, label, muted, pos, neg lipgloss.Style, width, height int, snapshot domain.OwnershipSnapshot) string {
	var b strings.Builder
	b.WriteString(section.Render("PREVIEW") + "\n\n")

	b.WriteString(label.Render("Breakdown") + "\n")
	b.WriteString(renderWrappedTextBlock(muted, fmt.Sprintf("Institutions %s  •  Float %s  •  Insiders %s", formatOwnershipPercent(snapshot.Summary.InstitutionsPercentHeld), formatOwnershipPercent(snapshot.Summary.InstitutionsFloatPercentHeld), formatOwnershipPercent(snapshot.Summary.InsidersPercentHeld)), width))
	b.WriteString("\n\n" + muted.Render("Institution count") + "\n")
	b.WriteString(renderWrappedTextBlock(muted, ui.FormatCompactInt(int64(snapshot.Summary.InstitutionsHoldingCount)), width))
	b.WriteString("\n\n" + muted.Render("Coverage") + "\n")
	b.WriteString(renderWrappedTextBlock(muted, fmt.Sprintf("%d institutions  •  %d funds", len(snapshot.Institutions), len(snapshot.Funds)), width))

	item, ok := m.currentOwnerItem()
	if !ok {
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}

	b.WriteString("\n\n" + label.Render("Selected holder") + "\n")
	b.WriteString(renderWrappedTextBlock(ownerKindStyle(item.kind, pos, neg), strings.ToUpper(item.kind), width))
	b.WriteString("\n" + muted.Render(ownershipDateTimeLabel(item.holder)) + "\n")
	if strings.TrimSpace(item.holder.Name) != "" {
		b.WriteString("\n" + muted.Render("Name") + "\n")
		b.WriteString(renderWrappedTextBlock(lipgloss.NewStyle(), item.holder.Name, width))
	}
	b.WriteString("\n\n" + muted.Render("Stake") + "\n")
	b.WriteString(renderWrappedTextBlock(muted, fmt.Sprintf("%s of shares outstanding", formatOwnershipPercent(item.holder.PercentHeld)), width))
	b.WriteString("\n\n" + muted.Render("Position") + "\n")
	b.WriteString(renderWrappedTextBlock(muted, fmt.Sprintf("%s shares  •  %s value", ui.FormatCompactInt(item.holder.Shares), formatOwnershipValue(item.holder.Value)), width))
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func ownershipDateLabel(item domain.OwnershipHolder) string {
	if item.ReportDate.IsZero() {
		return "-"
	}
	return item.ReportDate.Format("2006-01-02")
}

func ownershipDateTimeLabel(item domain.OwnershipHolder) string {
	if item.ReportDate.IsZero() {
		return "Report date unavailable"
	}
	return "Reported " + item.ReportDate.Format("2006-01-02")
}

func ownerKindStyle(kind string, pos, neg lipgloss.Style) lipgloss.Style {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "institution":
		return pos
	case "fund":
		return neg
	default:
		return lipgloss.NewStyle()
	}
}

func renderOwnershipRow(line string, item ownerListItem, pos, neg lipgloss.Style) string {
	kind := strings.ToUpper(item.kind)
	return strings.Replace(line, kind, ownerKindStyle(item.kind, pos, neg).Render(kind), 1)
}

func formatOwnershipPercent(v float64) string {
	return ui.FormatPercent(v * 100)
}

func formatOwnershipValue(v int64) string {
	if v == 0 {
		return "$0"
	}
	return "$" + ui.FormatCompactInt(v)
}
