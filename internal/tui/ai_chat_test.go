package tui

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"blackdesk/internal/providers"
	"blackdesk/internal/storage"
)

func TestAITabRendersTranscriptAndComposer(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.width = 140
	model.height = 42
	model.tabIdx = tabAI
	model.aiMessages = []aiMessage{
		{Role: aiMessageUser, Body: "Summarize AAPL setup", Meta: "Codex", Timestamp: time.Now()},
		{Role: aiMessageAssistant, Body: "AAPL is above its short-term trend with supportive breadth.", Meta: "Codex • 1.2s", Timestamp: time.Now()},
	}
	model.aiInput.SetValue("Compare it with MSFT")

	view := ansi.Strip(model.View())
	if !strings.Contains(view, "CHAT") {
		t.Fatal("expected AI chat panel")
	}
	if !strings.Contains(view, "You") || !strings.Contains(view, "AI") {
		t.Fatal("expected simplified chat speaker badges")
	}
	if !strings.Contains(view, "Summarize AAPL setup") {
		t.Fatal("expected transcript content")
	}
	if strings.Contains(view, "Codex") || strings.Contains(view, "1.2s") {
		t.Fatal("expected transcript metadata to stay hidden")
	}
	if strings.Contains(view, "PROMPT") {
		t.Fatal("expected AI bottom prompt section to be removed")
	}
	if !strings.Contains(view, "SETUP") || !strings.Contains(view, "CONTEXT") {
		t.Fatal("expected AI sidebars")
	}
}

func TestAITabContextSidebarHidesProviderAndMarketRows(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.width = 140
	model.height = 42
	model.tabIdx = tabAI

	view := ansi.Strip(model.View())
	if strings.Contains(view, "Provider test") {
		t.Fatal("expected provider row to be hidden from context sidebar")
	}
	if strings.Contains(view, "Market test") {
		t.Fatal("expected market row to be hidden from context sidebar")
	}
	if !strings.Contains(view, "Symbol") || !strings.Contains(view, "Model") {
		t.Fatal("expected other context rows to remain visible")
	}
}

func TestAITabShowsTypingIndicatorWhileRunning(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.width = 120
	model.height = 30
	model.tabIdx = tabAI
	model.aiRunning = true
	model.clock = time.Date(2026, 4, 3, 12, 0, 2, 0, time.UTC)
	model.aiMessages = []aiMessage{{Role: aiMessageUser, Body: "ping", Timestamp: time.Now()}}

	view := ansi.Strip(model.View())
	if !strings.Contains(view, "AI") || !strings.Contains(view, "thinking") {
		t.Fatal("expected animated AI typing indicator while running")
	}
	if strings.Index(view, "ping") > strings.Index(view, "thinking") {
		t.Fatal("expected typing indicator to render after the latest user message")
	}
	if strings.Contains(view, "Running Codex") {
		t.Fatal("expected static connector running text to be removed")
	}
}

func TestAITabUsesSameMainFrameGridAsQuote(t *testing.T) {
	quoteModel := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	quoteModel.width = 140
	quoteModel.height = 42
	quoteModel.tabIdx = tabQuote

	aiModel := quoteModel
	aiModel.tabIdx = tabAI

	quoteLines := strings.Split(ansi.Strip(quoteModel.View()), "\n")
	aiLines := strings.Split(ansi.Strip(aiModel.View()), "\n")
	if len(quoteLines) < 3 || len(aiLines) < 3 {
		t.Fatal("expected full frame output")
	}
	if quoteLines[2] != aiLines[2] {
		t.Fatalf("expected AI main frame grid to match quote tab\nquote: %q\nai:    %q", quoteLines[2], aiLines[2])
	}
}

func TestAITabViewStaysWithinViewport(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.width = 110
	model.height = 36
	model.tabIdx = tabAI
	model.config.AIModel = "gpt-5.4-mini-with-a-very-long-model-name-to-force-wrapping"
	model.aiDuration = 1350 * time.Millisecond

	lines := strings.Split(ansi.Strip(model.View()), "\n")
	if len(lines) != model.height {
		t.Fatalf("expected %d viewport lines, got %d", model.height, len(lines))
	}
	for i, line := range lines {
		if lipgloss.Width(line) > model.width {
			t.Fatalf("line %d exceeds viewport width: got %d want <= %d\n%q", i+1, lipgloss.Width(line), model.width, line)
		}
	}
}

func TestAITabClearsTranscriptWithX(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.tabIdx = tabAI
	model.aiMessages = []aiMessage{{Role: aiMessageUser, Body: "hello"}}
	model.aiConversationSummary = "- User: old context"
	model.aiCompactedMessages = 6
	model.aiOutput = "old"

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m := updated.(Model)
	if len(m.aiMessages) != 0 {
		t.Fatal("expected transcript to clear")
	}
	if m.aiOutput != "" {
		t.Fatal("expected AI output to clear")
	}
	if m.aiConversationSummary != "" || m.aiCompactedMessages != 0 {
		t.Fatal("expected AI summary state to clear too")
	}
}

func TestAITabFTogglesFullscreen(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.tabIdx = tabAI

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	m := updated.(Model)
	if !m.aiFullscreen {
		t.Fatal("expected f to enable AI fullscreen on AI tab")
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	m = updated.(Model)
	if m.aiFullscreen {
		t.Fatal("expected f to disable AI fullscreen on second press")
	}
}

func TestAIFullscreenHidesSidebars(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.width = 140
	model.height = 42
	model.tabIdx = tabAI
	model.aiFullscreen = true
	model.aiMessages = []aiMessage{{Role: aiMessageAssistant, Body: "fullscreen body", Timestamp: time.Now()}}

	view := ansi.Strip(model.View())
	if strings.Contains(view, "SETUP") || strings.Contains(view, "CONTEXT") {
		t.Fatal("expected AI fullscreen to hide sidebars")
	}
	if !strings.Contains(view, "CHAT") || !strings.Contains(view, "fullscreen body") {
		t.Fatal("expected AI fullscreen to keep chat visible")
	}
}

func TestAITabAutoscrollsToLatestMessages(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.width = 110
	model.height = 18
	model.tabIdx = tabAI
	for i := 1; i <= 12; i++ {
		model.aiMessages = append(model.aiMessages, aiMessage{
			Role: aiMessageAssistant,
			Body: fmt.Sprintf("message %02d", i),
			Meta: "Codex",
		})
	}

	view := ansi.Strip(model.View())
	if !strings.Contains(view, "message 12") {
		t.Fatal("expected latest message to be visible at bottom anchor")
	}
	if strings.Contains(view, "message 01") {
		t.Fatal("expected oldest message to be out of view when transcript overflows")
	}
}

func TestAITabArrowScrollMovesThroughTranscript(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.width = 110
	model.height = 18
	model.tabIdx = tabAI
	for i := 1; i <= 12; i++ {
		model.aiMessages = append(model.aiMessages, aiMessage{
			Role: aiMessageAssistant,
			Body: fmt.Sprintf("message %02d", i),
			Meta: "Codex",
		})
	}

	initialView := ansi.Strip(model.View())
	var m Model
	updated := tea.Model(model)
	for i := 0; i < 12; i++ {
		updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyUp})
		m = updated.(Model)
	}
	if m.aiScroll == 0 {
		t.Fatal("expected AI scroll offset to increase after scrolling up")
	}
	view := ansi.Strip(m.View())
	if view == initialView {
		t.Fatal("expected transcript view to change after scrolling up")
	}

	for m.aiScroll > 0 {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = updated.(Model)
	}
	if m.aiScroll != 0 {
		t.Fatalf("expected AI scroll offset to return to latest message, got %d", m.aiScroll)
	}
	if !strings.Contains(ansi.Strip(m.View()), "message 12") {
		t.Fatal("expected latest message to be visible again after scrolling back down")
	}
}

func TestAIContextStatusTracksRevisionAndRunningState(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})

	if status := model.aiContextStatusLine(); status != "cold" {
		t.Fatalf("expected cold status by default, got %q", status)
	}

	model.aiLastContext = "{\"symbol\":\"AAPL\"}"
	model.aiLastSymbol = model.activeSymbol()
	model.aiContextRevision = 3
	model.aiLastContextRevision = 3
	if status := model.aiContextStatusLine(); status != "stable" {
		t.Fatalf("expected stable status for matching revisions, got %q", status)
	}

	model.touchAIContext()
	if status := model.aiContextStatusLine(); status != "stale" {
		t.Fatalf("expected stale status after context revision changes, got %q", status)
	}

	model.aiRunning = true
	if status := model.aiContextStatusLine(); status != "refreshing" {
		t.Fatalf("expected refreshing status while AI is running, got %q", status)
	}
}

func TestAIResponseKeepsContextStaleIfAppChangedDuringRun(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.aiContextRevision = 6
	model.aiRunning = true

	updated, _ := model.handleAIResponseLoaded(aiResponseLoadedMsg{
		output:          "done",
		contextSent:     "{\"symbol\":\"AAPL\"}",
		contextRevision: 5,
		symbol:          model.activeSymbol(),
	})

	if status := updated.aiContextStatusLine(); status != "stale" {
		t.Fatalf("expected stale status when app context advanced during run, got %q", status)
	}
}

func TestPassiveDataLoadsDoNotInvalidateAIContextStatus(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.aiLastContext = "{\"symbol\":\"AAPL\"}"
	model.aiLastSymbol = model.activeSymbol()
	model.aiContextRevision = 4
	model.aiLastContextRevision = 4

	updated, _ := model.handleQuoteLoaded(quoteLoadedMsg{
		symbol: model.activeSymbol(),
		quote:  model.quote,
	})
	if status := updated.aiContextStatusLine(); status != "stable" {
		t.Fatalf("expected passive quote load to keep stable status, got %q", status)
	}

	updated, _ = updated.handleNewsLoaded(newsLoadedMsg{})
	if status := updated.aiContextStatusLine(); status != "stable" {
		t.Fatalf("expected passive news load to keep stable status, got %q", status)
	}
}
