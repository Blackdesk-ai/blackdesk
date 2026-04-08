package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"blackdesk/internal/domain"
	"blackdesk/internal/providers"
	"blackdesk/internal/storage"
)

func TestHeaderTabsShowMarketsQuoteNewsAndAI(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.width = 120
	model.height = 32

	view := model.View()
	if !strings.Contains(view, "Markets") || !strings.Contains(view, "Quote") || !strings.Contains(view, "News") || !strings.Contains(view, "AI") {
		t.Fatal("expected all top-level tabs")
	}
	if strings.Contains(view, "Fundamentals") || strings.Contains(view, "Technicals") {
		t.Fatal("expected removed tabs to be absent from header")
	}
}

func TestDefaultViewShowsMarketDashboard(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.width = 140
	model.height = 40
	model.watchQuotes["SPY"] = domain.QuoteSnapshot{Symbol: "SPY", Price: 512.30, ChangePercent: 0.84}
	model.watchQuotes["QQQ"] = domain.QuoteSnapshot{Symbol: "QQQ", Price: 442.18, ChangePercent: 1.11}
	model.watchQuotes["TLT"] = domain.QuoteSnapshot{Symbol: "TLT", Price: 92.44, ChangePercent: -0.34}

	view := model.View()
	if !strings.Contains(view, "Global Market Board") {
		t.Fatal("expected market dashboard center")
	}
	if !strings.Contains(view, "MARKET PULSE") {
		t.Fatal("expected market dashboard right rail")
	}
	if !strings.Contains(view, "GLOBAL SNAPSHOT") {
		t.Fatal("expected market dashboard left rail")
	}
	if !strings.Contains(ansi.Strip(view), "AI INSIGHT (i)") {
		t.Fatal("expected market AI insight key hint in sidebar")
	}
	if !strings.Contains(view, "AI INSIGHT") {
		t.Fatal("expected market AI insight block")
	}
	if !strings.Contains(view, "REGIONS") {
		t.Fatal("expected market dashboard bottom section")
	}
	if strings.Count(view, "SECTORS") < 2 {
		t.Fatal("expected market dashboard sectors panels")
	}
}

func TestMarketsBoardsUseSessionDisplayQuotes(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.watchQuotes["SPY"] = domain.QuoteSnapshot{
		Symbol:              "SPY",
		Price:               512.30,
		ChangePercent:       0.84,
		MarketState:         domain.MarketStatePre,
		PreMarketPrice:      520.55,
		PreMarketChangePerc: 2.18,
	}
	model.watchQuotes["^VIX"] = domain.QuoteSnapshot{
		Symbol:               "^VIX",
		Price:                14.10,
		ChangePercent:        -0.50,
		MarketState:          domain.MarketStatePost,
		PostMarketPrice:      15.25,
		PostMarketChangePerc: 1.40,
	}
	model.watchQuotes["DX-Y.NYB"] = domain.QuoteSnapshot{
		Symbol:              "DX-Y.NYB",
		Price:               104.10,
		ChangePercent:       0.12,
		MarketState:         domain.MarketStatePre,
		PreMarketPrice:      104.55,
		PreMarketChangePerc: 0.43,
	}
	model.watchQuotes["TLT"] = domain.QuoteSnapshot{
		Symbol:               "TLT",
		Price:                92.44,
		ChangePercent:        -0.34,
		MarketState:          domain.MarketStatePost,
		PostMarketPrice:      91.80,
		PostMarketChangePerc: -1.02,
	}

	board := ansi.Strip(renderMarketBoard(model, marketUSBoard[:1], 32, lipgloss.NewStyle()))
	if !strings.Contains(board, "520.55") {
		t.Fatal("expected markets board to use premarket price")
	}
	if !strings.Contains(board, "+2.18%") {
		t.Fatal("expected markets board to use premarket move")
	}

	pressure := ansi.Strip(marketPressureLine(model))
	if !strings.Contains(pressure, "VIX +1.40%") {
		t.Fatal("expected market pressure to use postmarket VIX move")
	}
	if !strings.Contains(pressure, "USD +0.43%") {
		t.Fatal("expected market pressure to use premarket dollar move")
	}
	if !strings.Contains(pressure, "TLT -1.02%") {
		t.Fatal("expected market pressure to use postmarket TLT move")
	}
}

func TestMarketOpinionBlockShowsThinkingAndCachedText(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.aiMarketOpinionRunning = true
	model.clock = time.Date(2026, 4, 3, 12, 0, 2, 0, time.UTC)

	block := ansi.Strip(model.renderMarketOpinionBlock(lipgloss.NewStyle(), 36))
	if !strings.Contains(block, "thinking") {
		t.Fatal("expected thinking indicator while market opinion runs")
	}

	model.aiMarketOpinionRunning = false
	model.aiMarketOpinion = "Tone is mildly risk-off as defensives lead while volatility stays elevated."
	model.aiMarketOpinionUpdated = time.Date(2026, 4, 3, 12, 4, 0, 0, time.UTC)

	block = ansi.Strip(model.renderMarketOpinionBlock(lipgloss.NewStyle(), 36))
	if !strings.Contains(block, "mildly risk-off") {
		t.Fatal("expected cached market opinion text")
	}
	if !strings.Contains(block, "Updated ") {
		t.Fatal("expected market opinion updated timestamp")
	}
}

func TestMarketsRefreshAndInsightKeysAreSeparated(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.tabIdx = tabMarkets
	for i, symbol := range marketDashboardSymbols {
		if i >= 12 {
			break
		}
		model.watchQuotes[strings.ToUpper(symbol)] = domain.QuoteSnapshot{Symbol: symbol, Price: 1}
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	model = updated.(Model)
	if model.pendingMarketOpinionRefresh {
		t.Fatal("expected refresh key to leave market AI opinion idle")
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	model = updated.(Model)
	if !model.pendingMarketOpinionRefresh {
		t.Fatal("expected insight key to queue market AI opinion refresh")
	}
}
