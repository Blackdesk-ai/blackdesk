package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"blackdesk/internal/providers"
	"blackdesk/internal/storage"
)

func TestAITabRendersAssistantMarkdown(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.width = 140
	model.height = 42
	model.tabIdx = tabAI
	model.aiMessages = []aiMessage{
		{Role: aiMessageAssistant, Body: "# Setup\n\n- Momentum\n- Trend\n\n| Metric | Value |\n| --- | --- |\n| RSI | 64 |", Timestamp: time.Now()},
	}

	view := ansi.Strip(model.View())
	if !strings.Contains(view, "SETUP") || !strings.Contains(view, "Momentum") || !strings.Contains(view, "Trend") {
		t.Fatal("expected markdown headings and lists to render in assistant transcript")
	}
	if !strings.Contains(view, "RSI") || !strings.Contains(view, "Value: 64") {
		t.Fatal("expected markdown table content to render in assistant transcript")
	}
	if !strings.Contains(view, "• Momentum") || !strings.Contains(view, "• Trend") {
		t.Fatal("expected markdown list items to render with terminal bullets")
	}
	if strings.Contains(view, "| Metric | Value |") {
		t.Fatal("expected markdown table syntax to be rendered, not shown raw")
	}
}

func TestAITabRendersInlineMarkdownInNormalView(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.width = 140
	model.height = 42
	model.tabIdx = tabAI
	model.aiMessages = []aiMessage{
		{Role: aiMessageAssistant, Body: "Price action is **strong** and see [link](https://example.com/test) with `code`.", Timestamp: time.Now()},
	}

	view := ansi.Strip(model.View())
	if !strings.Contains(view, "Price action is strong") || !strings.Contains(view, "link") || !strings.Contains(view, "https://example.com/test") || !strings.Contains(view, "code") {
		t.Fatal("expected normal AI mode to simplify inline markdown markers")
	}
	if strings.Contains(view, "**") || strings.Contains(view, "[link]") || strings.Contains(view, "`code`") {
		t.Fatal("expected markdown markers to be stripped in normal view")
	}
}

func TestRenderMarkdownTranscriptWrapsWideLines(t *testing.T) {
	lines := renderMarkdownTranscript("| Section | Indicator | Value | Interpretation |\n| --- | --- | --- | --- |\n| Trend | SMA 20 / 50 / 200 | 26.76 / 29.64 / 39.8 | Price remains below the key moving averages and the trend stays weak across multiple windows. |", 48)
	if len(lines) < 4 {
		t.Fatal("expected wrapped markdown output")
	}
	for i, line := range lines {
		if lipgloss.Width(line) > 48 {
			t.Fatalf("line %d exceeds wrap width: got %d want <= 48\n%q", i+1, lipgloss.Width(line), line)
		}
	}
}

func TestRenderMarkdownTranscriptCompactsBareURLs(t *testing.T) {
	lines := renderMarkdownTranscript("Sources: https://stockanalysis.com/stocks/msft/ https://www.microsoft.com/en-us/Investor/earnings/fy-2026-q2/performance", 120)
	view := ansi.Strip(strings.Join(lines, "\n"))
	if !strings.Contains(view, "stockanalysis.com/stocks/msft") {
		t.Fatal("expected stockanalysis URL to be compacted")
	}
	if !strings.Contains(view, "microsoft.com/en-us/Investor") {
		t.Fatal("expected microsoft URL to be compacted")
	}
	if strings.Contains(view, "https://stockanalysis.com/stocks/msft/") || strings.Contains(view, "https://www.microsoft.com/en-us/Investor/earnings/fy-2026-q2/performance") {
		t.Fatal("expected raw URLs to stay hidden behind compact labels")
	}
}

func TestNormalizeMarkdownInputSeparatesGluedURLs(t *testing.T) {
	lines := renderMarkdownTranscript("Sources: StockAnalysis:https://stockanalysis.com/stocks/msft/,https://finviz.com/quote.ashx?t=MSFT Microsoft IR webcasthttps://www.microsoft.com/en-us/Investor/earnings/fy-2026-q2/performance", 160)
	view := ansi.Strip(strings.Join(lines, "\n"))
	if !strings.Contains(view, "StockAnalysis:") {
		t.Fatal("expected source label to remain visible")
	}
	if !strings.Contains(view, "stockanalysis.com/stocks/msft") || !strings.Contains(view, "finviz.com/quote.ashx") || !strings.Contains(view, "microsoft.com/en-us/Investor") {
		t.Fatal("expected glued URLs to be separated and compacted")
	}
}
