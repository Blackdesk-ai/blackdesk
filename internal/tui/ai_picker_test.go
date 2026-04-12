package tui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"blackdesk/internal/agents"
	"blackdesk/internal/application"
	"blackdesk/internal/providers"
	"blackdesk/internal/storage"
)

func TestAIPickerUsesCenterSetup(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.services = application.NewServices(nil, agents.NewRegistry(), nil)
	model.width = 120
	model.height = 40
	model.tabIdx = tabAI
	model.aiPickerOpen = true
	model.aiPickerStep = aiPickerStepConnector
	model.aiModels = map[string][]string{"codex": {"gpt-5.4", "gpt-5.4-mini"}}
	model.config.AIConnector = "codex"
	model.config.AIModel = "gpt-5.4"

	view := model.View()
	if !strings.Contains(view, "AI SETUP") || !strings.Contains(view, "AI PROVIDER") {
		t.Fatal("expected centered AI provider step to render")
	}
	if !strings.Contains(view, "GUIDE") || !strings.Contains(view, "CONTEXT") {
		t.Fatal("expected guide in left sidebar and context in right sidebar")
	}
	if strings.Contains(view, "CHAT") {
		t.Fatal("expected chat to be replaced while picker is open")
	}
	if !strings.Contains(view, "Codex") {
		t.Fatal("expected provider list in picker")
	}
}

func TestAISetupShowsFilingProgressState(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.width = 120
	model.height = 40
	model.tabIdx = tabAI
	model.aiFilingRun = aiFilingRunState{
		chunks:       []filingTextChunk{{Index: 1, Total: 4}, {Index: 2, Total: 4}, {Index: 3, Total: 4}, {Index: 4, Total: 4}},
		nextChunkIdx: 2,
	}
	model.aiFilingRunActive = true

	view := ansi.Strip(model.View())
	if !strings.Contains(view, "Filing    chunk 3/4") {
		t.Fatalf("expected filing progress in AI setup block, got %q", view)
	}
}

func TestTypingOnAITabFocusesAIComposer(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.tabIdx = tabAI

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m := updated.(Model)
	if m.tabIdx != tabAI {
		t.Fatal("expected typing to stay on AI tab")
	}
	if !m.aiFocused {
		t.Fatal("expected typing to focus AI composer")
	}
	if m.aiInput.Value() != "h" {
		t.Fatalf("expected typed key to enter AI composer, got %q", m.aiInput.Value())
	}
}

func TestDotFocusesAIComposerOnAITab(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.tabIdx = tabAI

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'.'}})
	m := updated.(Model)
	if !m.aiFocused {
		t.Fatal("expected dot to focus AI composer on AI tab")
	}
	if m.aiInput.Value() != "" {
		t.Fatalf("expected dot shortcut not to be inserted in AI input, got %q", m.aiInput.Value())
	}
}

func TestDotSubmitsFocusedAIComposerWithoutTypingDot(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.tabIdx = tabAI
	model.aiFocused = true
	model.aiInput.Focus()
	model.aiInput.SetValue("summarize aapl")

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'.'}})
	m := updated.(Model)
	if m.aiFocused {
		t.Fatal("expected dot to blur AI composer after submit")
	}
	if m.aiInput.Value() != "" {
		t.Fatalf("expected prompt to clear after dot submit, got %q", m.aiInput.Value())
	}
	if !m.aiRunning {
		t.Fatal("expected dot to start AI run when composer is focused")
	}
	if len(m.aiMessages) == 0 || m.aiMessages[len(m.aiMessages)-1].Body != "summarize aapl" {
		t.Fatal("expected dot submit to append user message")
	}
	if cmd == nil {
		t.Fatal("expected dot submit to prepare AI context")
	}
}

func TestCOpensAIPickerOnlyOnAITab(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.tabIdx = tabQuote

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m := updated.(Model)
	if m.aiPickerOpen {
		t.Fatal("expected AI picker to stay closed outside AI tab")
	}
	if cmd != nil {
		t.Fatal("expected no model-loading command outside AI tab")
	}

	model.tabIdx = tabAI
	updated, cmd = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m = updated.(Model)
	if !m.aiPickerOpen {
		t.Fatal("expected AI picker to open on AI tab")
	}
	if m.aiPickerStep != aiPickerStepConnector {
		t.Fatal("expected picker to start on connector step")
	}
	if cmd != nil {
		t.Fatal("expected no model-loading command until connector is confirmed")
	}
}

func TestCStillSwitchesQuoteToChart(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.tabIdx = tabQuote
	model.quoteCenterMode = quoteCenterFundamentals

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m := updated.(Model)
	if m.aiPickerOpen {
		t.Fatal("expected AI picker to stay closed on quote tab")
	}
	if m.quoteCenterMode != quoteCenterChart {
		t.Fatal("expected c to keep quote chart shortcut on quote tab")
	}
}

func TestAIStartsUnselectedWithoutFallbackConnector(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.services = application.NewServices(nil, agents.NewRegistry(), nil)

	if model.activeAIConnectorID() != "" {
		t.Fatalf("expected no active AI connector on clean/default start, got %q", model.activeAIConnectorID())
	}
	if model.activeAIConnectorLabel() != "Not selected" {
		t.Fatalf("expected 'Not selected' AI label, got %q", model.activeAIConnectorLabel())
	}
}

func TestStatusMetaTextUsesAIModelOrExplicitUnsetLabel(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})

	if got := model.statusMetaText(); !strings.Contains(got, "AI: No model selected") {
		t.Fatalf("expected explicit unset AI model label, got %q", got)
	}

	model.config.AIConnector = "codex"
	model.config.AIModel = "gpt-5.4-mini"
	got := model.statusMetaText()
	if !strings.Contains(got, "AI: gpt-5.4-mini") {
		t.Fatalf("expected AI model in status meta text, got %q", got)
	}
	if strings.Contains(got, "Codex") {
		t.Fatalf("expected connector label to stay hidden from status meta text, got %q", got)
	}
}

func TestAIPickerStatusLineUsesModelLabel(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.width = 120
	model.height = 30
	model.tabIdx = tabAI
	model.aiPickerOpen = true

	view := ansi.Strip(model.View())
	if !strings.Contains(view, "AI picker: No model selected") {
		t.Fatalf("expected explicit unset model label in picker status line, got %q", view)
	}

	model.config.AIConnector = "codex"
	model.config.AIModel = "gpt-5.4"
	view = ansi.Strip(model.View())
	if !strings.Contains(view, "AI picker: gpt-5.4") {
		t.Fatalf("expected selected model in picker status line, got %q", view)
	}
	if strings.Contains(view, "AI picker: Codex") {
		t.Fatalf("expected connector label to stay hidden from picker status line, got %q", view)
	}
}

func TestAIPickerDoesNotRenderOnNonAITabs(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.width = 120
	model.height = 40
	model.tabIdx = tabMarkets
	model.aiPickerOpen = true

	view := model.View()
	if strings.Contains(view, "AI TARGET") || strings.Contains(view, "AI PROVIDER") || strings.Contains(view, "AI MODEL") {
		t.Fatal("expected AI picker UI to stay hidden outside AI tab")
	}
}
