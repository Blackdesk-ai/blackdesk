package tui

import (
	"context"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"blackdesk/internal/application"
	"blackdesk/internal/buildinfo"
	"blackdesk/internal/domain"
)

func NewModel(ctx context.Context, deps Dependencies) Model {
	now := time.Now()
	ti := textinput.New()
	ti.Placeholder = "Search ticker or company"
	ti.CharLimit = 40
	ti.Width = 32
	command := textinput.New()
	command.Placeholder = "Search functions or symbols"
	command.CharLimit = 80
	command.Width = 64
	ai := textinput.New()
	ai.Placeholder = "Ask the selected local AI about the market, this symbol, or the current app state (Esc to close)"
	ai.CharLimit = 1000
	ai.Width = 96

	rangeIdx := 2
	for i, item := range ranges {
		if item.Range == deps.Config.DefaultRange && item.Interval == deps.Config.DefaultInterval {
			rangeIdx = i
			break
		}
	}

	selectedIdx := 0
	for i, symbol := range deps.Config.Watchlist {
		if strings.EqualFold(symbol, deps.Config.ActiveSymbol) {
			selectedIdx = i
			break
		}
	}

	services := deps.Services
	if services == nil {
		services = application.NewServicesWithFilings(deps.Registry, deps.AgentRegistry, deps.ConfigStore, deps.FilingsProvider)
	}
	screenerDefs := services.Screeners()

	return Model{
		ctx:                    ctx,
		services:               services,
		marketRiskProvider:     deps.MarketRiskProvider,
		config:                 deps.Config,
		workspaceRoot:          deps.WorkspaceRoot,
		selectedIdx:            selectedIdx,
		rangeIdx:               rangeIdx,
		status:                 "Loading market data…",
		clock:                  now,
		lastAutoRefresh:        now,
		lastMarketNews:         now,
		appVersion:             buildinfo.NormalizedVersion(),
		searchInput:            ti,
		commandInput:           command,
		aiInput:                ai,
		aiModels:               make(map[string][]string),
		watchQuotes:            make(map[string]domain.QuoteSnapshot),
		technicalCache:         make(map[string]domain.PriceSeries),
		statementCache:         make(map[string]domain.FinancialStatement),
		insiderCache:           make(map[string]domain.InsiderSnapshot),
		filingsCache:           make(map[string]domain.FilingsSnapshot),
		marketOpinionHistory:   make(map[string]domain.PriceSeries),
		marketOpinionHistoryAt: make(map[string]time.Time),
		screenerDefs:           append([]domain.ScreenerDefinition(nil), screenerDefs...),
		statementKind:          domain.StatementKindIncome,
		statementFreq:          domain.StatementFrequencyAnnual,
	}
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		m.loadAllCmd(m.activeSymbol()),
		tickCmd(time.Second),
	}
	if m.shouldCheckForUpdates() {
		cmds = append(cmds, m.checkForUpdatesCmd())
	}
	if m.tabIdx == tabNews {
		cmds = append(cmds, m.loadMarketNewsCmd(), m.loadMarketQuotesCmd())
	}
	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.MouseMsg:
		return m.handleMouseMsg(msg)
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	case quoteLoadedMsg:
		return m.handleQuoteLoaded(msg)
	case quotesLoadedMsg:
		return m.handleQuotesLoaded(msg)
	case historyLoadedMsg:
		return m.handleHistoryLoaded(msg)
	case technicalHistoryLoadedMsg:
		return m.handleTechnicalHistoryLoaded(msg)
	case newsLoadedMsg:
		return m.handleNewsLoaded(msg)
	case marketNewsLoadedMsg:
		return m.handleMarketNewsLoaded(msg)
	case marketRiskLoadedMsg:
		return m.handleMarketRiskLoaded(msg)
	case screenerLoadedMsg:
		return m.handleScreenerLoaded(msg)
	case fundamentalsLoadedMsg:
		return m.handleFundamentalsLoaded(msg)
	case statementLoadedMsg:
		return m.handleStatementLoaded(msg)
	case insidersLoadedMsg:
		return m.handleInsidersLoaded(msg)
	case filingsLoadedMsg:
		return m.handleFilingsLoaded(msg)
	case searchDebouncedMsg:
		return m.handleSearchDebounced(msg)
	case searchLoadedMsg:
		return m.handleSearchLoaded(msg)
	case commandPaletteDebouncedMsg:
		return m.handleCommandPaletteDebounced(msg)
	case commandPaletteLoadedMsg:
		return m.handleCommandPaletteLoaded(msg)
	case tickMsg:
		return m.handleTick(msg)
	case aiResponseLoadedMsg:
		return m.handleAIResponseLoaded(msg)
	case aiMarketOpinionLoadedMsg:
		return m.handleAIMarketOpinionLoaded(msg)
	case aiQuoteInsightPreparedMsg:
		return m.handleAIQuoteInsightPrepared(msg)
	case aiQuoteInsightLoadedMsg:
		return m.handleAIQuoteInsightLoaded(msg)
	case aiModelsLoadedMsg:
		return m.handleAIModelsLoaded(msg)
	case aiContextPreparedMsg:
		return m.handleAIContextPrepared(msg)
	case aiFilingAnalysisPreparedMsg:
		return m.handleAIFilingAnalysisPrepared(msg)
	case versionCheckLoadedMsg:
		return m.handleVersionCheckLoaded(msg)
	case versionUpgradeLoadedMsg:
		return m.handleVersionUpgradeLoaded(msg)
	}
	return m, nil
}

func tickCmd(delay time.Duration) tea.Cmd {
	return tea.Tick(delay, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
