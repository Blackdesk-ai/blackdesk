package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"blackdesk/internal/domain"
)

func (m Model) renderGlobalFullscreenPage(header, section, label, muted, pos, neg lipgloss.Style, width, height int) string {
	switch m.globalPageKind {
	case globalPageCalendar:
		return m.renderCalendarPage(header, section, label, muted, pos, neg, width, height)
	default:
		return ""
	}
}

func (m Model) renderCalendarPage(header, section, label, muted, pos, neg lipgloss.Style, width, height int) string {
	listWidth := clamp((width*3)/5, 50, 84)
	previewWidth := max(24, width-listWidth-3)
	bodyHeight := max(8, height-6)

	var b strings.Builder
	b.WriteString(section.Render("CALENDAR") + "\n\n")
	left := lipgloss.NewStyle().Width(listWidth).Render(m.renderCalendarList(section, muted, listWidth, bodyHeight))
	right := lipgloss.NewStyle().Width(previewWidth).Render(m.renderCalendarPreview(section, label, muted, previewWidth, bodyHeight))
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, left, "   ", right))
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func (m Model) renderCalendarList(section, muted lipgloss.Style, width, height int) string {
	var b strings.Builder
	b.WriteString(renderStatusLine(width, section.Render("IMPORTANT EVENTS"), m.renderCalendarFilterTabs(muted, width)) + "\n\n")
	if len(m.calendar.Events) == 0 {
		switch {
		case m.errCalendar != nil:
			b.WriteString(m.errCalendar.Error())
		case !m.calendar.StartDate.IsZero():
			b.WriteString(muted.Render("No high-importance events returned for this range"))
		default:
			b.WriteString(muted.Render("Loading economic calendar…"))
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
	timeColWidth := 5
	eventColWidth := max(14, width-dateColWidth-timeColWidth-4)
	headerLine := fmt.Sprintf(
		"%-*s  %-*s  %s",
		dateColWidth, "DATE",
		timeColWidth, "TIME",
		truncateText("EVENT", eventColWidth),
	)
	b.WriteString(muted.Render(truncateText(headerLine, width)) + "\n")
	b.WriteString(muted.Render(strings.Repeat("─", max(12, min(width, dateColWidth+timeColWidth+eventColWidth+4)))) + "\n")

	visibleRows := max(3, height/2)
	start := 0
	if m.calendarSel >= visibleRows {
		start = m.calendarSel - visibleRows + 1
	}
	end := min(len(m.calendar.Events), start+visibleRows)
	for i := start; i < end; i++ {
		item := m.calendar.Events[i]
		line := fmt.Sprintf(
			"%-*s  %-*s  %s",
			dateColWidth, truncateText(calendarEventDateLabel(item), dateColWidth),
			timeColWidth, calendarEventTimeLabel(item),
			truncateText(calendarEventListLabel(item), eventColWidth),
		)
		line = truncateText(line, width)
		if i == m.calendarSel {
			b.WriteString(selectedStyle.Render(line) + "\n")
			continue
		}
		b.WriteString(line + "\n")
	}

	b.WriteString("\n" + muted.Render("←/→ filter • ↑/↓ move • r refresh • Esc close"))
	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func (m Model) renderCalendarPreview(section, label, muted lipgloss.Style, width, height int) string {
	var b strings.Builder
	b.WriteString(section.Render("PREVIEW") + "\n\n")
	item, ok := m.currentCalendarEvent()
	if !ok {
		b.WriteString(muted.Render("Select an economic event to inspect it."))
		return clipLines(strings.TrimRight(b.String(), "\n"), height)
	}

	b.WriteString(renderWrappedTextBlock(label, item.Event, width) + "\n")
	b.WriteString(muted.Render(calendarEventMetaLine(item)) + "\n")

	if item.Period != "" {
		b.WriteString("\n" + muted.Render("Period") + "\n")
		b.WriteString(renderWrappedTextBlock(muted, item.Period, width))
	}

	b.WriteString("\n\n" + muted.Render("Data") + "\n")
	b.WriteString(renderWrappedTextBlock(muted, calendarEventDataLine(item), width))
	if item.RevisedFrom != "" {
		b.WriteString("\n" + renderWrappedTextBlock(muted, "Revised from "+item.RevisedFrom, width))
	}

	if item.Description != "" {
		b.WriteString("\n\n" + muted.Render("Description") + "\n")
		b.WriteString(renderWrappedTextBlock(muted, item.Description, width))
	}

	return clipLines(strings.TrimRight(b.String(), "\n"), height)
}

func (m Model) renderCalendarFilterTabs(muted lipgloss.Style, width int) string {
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
		mode  calendarFilterMode
		label string
	}{
		{mode: calendarFilterToday, label: "TODAY"},
		{mode: calendarFilterThisWeek, label: "THIS WEEK"},
	}

	parts := make([]string, 0, len(tabs))
	for _, tab := range tabs {
		style := inactive
		if tab.mode == m.calendarFilter {
			style = active
		}
		parts = append(parts, style.Render(tab.label))
	}
	return ansi.Truncate(strings.Join(parts, " "), width, "")
}

func calendarSummaryRangeLabel(snapshot domain.EconomicCalendarSnapshot, filter calendarFilterMode) string {
	start := snapshot.StartDate
	end := snapshot.EndDate
	if start.IsZero() || end.IsZero() || !end.After(start) {
		return strings.ToUpper(calendarFilterModeLabel(filter))
	}
	lastDay := end.Add(-time.Nanosecond)
	if filter == calendarFilterToday || sameDate(start, lastDay) {
		return start.Format("Mon 2006-01-02")
	}
	return start.Format("Mon 2006-01-02") + " to " + lastDay.Format("Mon 2006-01-02")
}

func calendarFilterModeLabel(filter calendarFilterMode) string {
	switch filter {
	case calendarFilterThisWeek:
		return "This Week"
	default:
		return "Today"
	}
}

func calendarEventDateLabel(item domain.EconomicCalendarEvent) string {
	if !item.EventAt.IsZero() {
		return item.EventAt.In(time.Local).Format("2006-01-02")
	}
	if item.Date.IsZero() {
		return "-"
	}
	return item.Date.Format("2006-01-02")
}

func calendarEventTimeLabel(item domain.EconomicCalendarEvent) string {
	value := strings.TrimSpace(item.EventTime)
	if value == "" {
		return "-"
	}
	if len(value) > 5 && strings.Contains(value, ":") {
		return value[:5]
	}
	return value
}

func calendarEventListLabel(item domain.EconomicCalendarEvent) string {
	if item.CountryCode == "" {
		return item.Event
	}
	return "[" + item.CountryCode + "] " + item.Event
}

func calendarEventMetaLine(item domain.EconomicCalendarEvent) string {
	parts := make([]string, 0, 3)
	if !item.EventAt.IsZero() {
		local := item.EventAt.In(time.Local)
		parts = append(parts, local.Format("Mon 2006-01-02"))
		parts = append(parts, local.Format("15:04 MST"))
	} else if !item.Date.IsZero() {
		parts = append(parts, item.Date.Format("Mon 2006-01-02"))
	}
	if item.EventTime != "" && item.EventAt.IsZero() {
		parts = append(parts, item.EventTime)
	}
	if item.CountryCode != "" {
		parts = append(parts, item.CountryCode)
	}
	if len(parts) == 0 {
		return "Global economic event"
	}
	return strings.Join(parts, " • ")
}

func calendarEventDataLine(item domain.EconomicCalendarEvent) string {
	parts := make([]string, 0, 3)
	if value := strings.TrimSpace(item.Actual); value != "" {
		parts = append(parts, "Actual "+value)
	}
	if value := strings.TrimSpace(item.ConsensusEstimate); value != "" {
		parts = append(parts, "Expectation "+value)
	}
	if value := strings.TrimSpace(item.Prior); value != "" {
		parts = append(parts, "Prior "+value)
	}
	if len(parts) == 0 {
		return "Yahoo did not provide release values for this event."
	}
	return strings.Join(parts, "  •  ")
}

func sameDate(left, right time.Time) bool {
	ly, lm, ld := left.Date()
	ry, rm, rd := right.Date()
	return ly == ry && lm == rm && ld == rd
}
