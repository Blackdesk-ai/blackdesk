package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"blackdesk/internal/providers"
	"blackdesk/internal/storage"
)

func TestCommandPaletteCalendarOpensGlobalCalendarPage(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.clock = time.Date(2026, 4, 12, 12, 0, 0, 0, time.UTC)
	model.commandPaletteOpen = true
	model.commandInput.Focus()
	model.commandPaletteItems = []commandPaletteItem{
		{Kind: commandPaletteItemFunction, FunctionID: "calendar", Title: "Calendar"},
	}

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updated.(Model)
	if !m.globalPageOpen || m.globalPageKind != globalPageCalendar {
		t.Fatalf("expected global calendar page, got open=%v kind=%d", m.globalPageOpen, m.globalPageKind)
	}
	if m.calendarFilter != calendarFilterToday {
		t.Fatalf("expected today filter, got %d", m.calendarFilter)
	}
	if cmd == nil {
		t.Fatal("expected economic calendar load command")
	}
}

func TestGlobalCalendarViewRendersLoadedEvents(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.width = 140
	model.height = 40
	model.clock = time.Date(2026, 4, 12, 12, 0, 0, 0, time.UTC)
	model.globalPageOpen = true
	model.globalPageKind = globalPageCalendar
	model.calendarFilter = calendarFilterThisWeek
	model.calendar = sampleEconomicCalendarSnapshot()

	view := model.View()
	if !strings.Contains(view, "CALENDAR") {
		t.Fatal("expected calendar section")
	}
	if !strings.Contains(view, "IMPORTANT EVENTS") || !strings.Contains(view, "PREVIEW") {
		t.Fatal("expected calendar list and preview")
	}
	if !strings.Contains(view, "Consumer Price Index YoY") || !strings.Contains(view, "Retail Sales MoM") {
		t.Fatal("expected economic events in view")
	}
	if !strings.Contains(view, "TODAY") || !strings.Contains(view, "THIS WEEK") {
		t.Fatal("expected calendar filter tabs")
	}
	if strings.Contains(view, "SYMBOLS") || strings.Contains(view, "\nNEWS\n") {
		t.Fatal("expected global calendar page to hide workspace sidebars")
	}
}

func TestGlobalCalendarUsesArrowNavigationInsteadOfWatchlist(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.globalPageOpen = true
	model.globalPageKind = globalPageCalendar
	model.config.Watchlist = []string{"AAPL", "MSFT"}
	model.config.ActiveSymbol = "AAPL"
	model.selectedIdx = 0
	model.calendar = sampleEconomicCalendarSnapshot()

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	m := updated.(Model)
	if m.selectedIdx != 0 {
		t.Fatalf("expected watchlist selection unchanged, got %d", m.selectedIdx)
	}
	if m.calendarSel != 1 {
		t.Fatalf("expected calendar selection to move, got %d", m.calendarSel)
	}
}

func TestCalendarPreviewOmitsMissingExpectationAndActual(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.width = 140
	model.height = 40
	model.globalPageOpen = true
	model.globalPageKind = globalPageCalendar
	model.calendarFilter = calendarFilterThisWeek
	model.calendar = sampleEconomicCalendarSnapshot()
	model.calendar.Events[0].Actual = ""
	model.calendar.Events[0].ConsensusEstimate = ""

	view := model.View()
	if strings.Contains(view, "Expectation -") || strings.Contains(view, "Actual -") {
		t.Fatal("expected preview to omit missing values instead of rendering dashes")
	}
	if !strings.Contains(view, "Prior 3.2%") {
		t.Fatal("expected preview to keep available values")
	}
}
