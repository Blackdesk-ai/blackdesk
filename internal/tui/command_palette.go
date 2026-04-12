package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"blackdesk/internal/domain"
)

type commandPaletteItemKind int

const (
	commandPaletteItemFunction commandPaletteItemKind = iota
	commandPaletteItemSymbol
)

type commandPaletteItem struct {
	Kind        commandPaletteItemKind
	Title       string
	Subtitle    string
	Description string
	Meta        string
	FunctionID  string
	Symbol      domain.SymbolRef
}

type commandPaletteFunction struct {
	ID          string
	Title       string
	Aliases     []string
	Category    string
	Description string
}

func (m *Model) openCommandPalette() tea.Cmd {
	m.helpOpen = false
	m.searchMode = false
	m.searchInput.Blur()
	m.aiFocused = false
	m.aiInput.Blur()
	m.aiPickerOpen = false
	m.commandPaletteOpen = true
	m.commandInput.SetValue("")
	m.commandInput.Focus()
	m.commandPaletteSymbolItems = nil
	m.commandPaletteDebounceID++
	m.commandPaletteRequestQuery = ""
	m.commandPaletteItems = m.buildCommandPaletteItems("")
	m.commandPaletteIdx = 0
	m.status = "Command palette"
	return nil
}

func (m *Model) closeCommandPalette(status string) {
	m.commandPaletteOpen = false
	m.commandInput.SetValue("")
	m.commandInput.Blur()
	m.commandPaletteItems = nil
	m.commandPaletteIdx = 0
	m.commandPaletteSymbolItems = nil
	m.commandPaletteRequestQuery = ""
	if strings.TrimSpace(status) != "" {
		m.status = status
	}
}

func (m Model) currentCommandPaletteItem() (commandPaletteItem, bool) {
	if len(m.commandPaletteItems) == 0 || m.commandPaletteIdx < 0 || m.commandPaletteIdx >= len(m.commandPaletteItems) {
		return commandPaletteItem{}, false
	}
	return m.commandPaletteItems[m.commandPaletteIdx], true
}

func (m *Model) refreshCommandPaletteItems() {
	query := strings.TrimSpace(m.commandInput.Value())
	m.commandPaletteItems = m.buildCommandPaletteItems(query)
	if len(m.commandPaletteItems) == 0 {
		m.commandPaletteIdx = 0
		return
	}
	if m.commandPaletteIdx >= len(m.commandPaletteItems) {
		m.commandPaletteIdx = len(m.commandPaletteItems) - 1
	}
	if m.commandPaletteIdx < 0 {
		m.commandPaletteIdx = 0
	}
}

func (m Model) buildCommandPaletteItems(query string) []commandPaletteItem {
	items := make([]commandPaletteItem, 0, 24)
	items = append(items, m.commandPaletteFunctionItems(query)...)
	items = append(items, m.commandPaletteSymbolResults(query)...)
	return items
}

func (m Model) commandPaletteFunctionItems(query string) []commandPaletteItem {
	type scoredItem struct {
		item  commandPaletteItem
		score int
	}
	needle := strings.ToLower(strings.TrimSpace(query))
	scored := make([]scoredItem, 0, 16)
	for _, fn := range m.commandPaletteFunctions() {
		score, ok := matchCommandPaletteFunction(fn, needle)
		if !ok {
			continue
		}
		meta := "Function"
		if fn.Category != "" {
			meta += " • " + fn.Category
		}
		scored = append(scored, scoredItem{
			score: score,
			item: commandPaletteItem{
				Kind:        commandPaletteItemFunction,
				Title:       fn.Title,
				Subtitle:    commandPaletteFunctionSubtitle(fn, m.activeSymbol()),
				Description: fn.Description,
				Meta:        meta,
				FunctionID:  fn.ID,
			},
		})
	}
	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].score != scored[j].score {
			return scored[i].score < scored[j].score
		}
		return scored[i].item.Title < scored[j].item.Title
	})
	items := make([]commandPaletteItem, 0, len(scored))
	for _, item := range scored {
		items = append(items, item.item)
	}
	return items
}

func (m Model) commandPaletteSymbolResults(query string) []commandPaletteItem {
	seen := make(map[string]struct{})
	out := make([]commandPaletteItem, 0, 16)
	for _, item := range m.commandPaletteWatchlistResults(query) {
		key := strings.ToUpper(strings.TrimSpace(item.Symbol.Symbol))
		if key == "" {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	for _, ref := range m.commandPaletteSymbolItems {
		key := strings.ToUpper(strings.TrimSpace(ref.Symbol))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, commandPaletteItem{
			Kind:        commandPaletteItemSymbol,
			Title:       strings.ToUpper(strings.TrimSpace(ref.Symbol)),
			Subtitle:    commandPaletteFirstNonEmpty(strings.TrimSpace(ref.Name), "Open symbol in Quote"),
			Description: "Open symbol in Quote and select it in the watchlist.",
			Meta:        commandPaletteSymbolMeta(ref, false),
			Symbol:      ref,
		})
	}
	return out
}

func (m Model) commandPaletteWatchlistResults(query string) []commandPaletteItem {
	needle := strings.ToLower(strings.TrimSpace(query))
	items := make([]commandPaletteItem, 0, min(8, len(m.config.Watchlist)))
	for _, symbol := range m.config.Watchlist {
		name := ""
		if quote, ok := m.watchQuotes[strings.ToUpper(symbol)]; ok {
			name = strings.TrimSpace(quote.ShortName)
		}
		ref := domain.SymbolRef{Symbol: symbol, Name: name}
		if needle != "" && !commandPaletteMatchesSymbol(ref, needle) {
			continue
		}
		items = append(items, commandPaletteItem{
			Kind:        commandPaletteItemSymbol,
			Title:       strings.ToUpper(strings.TrimSpace(ref.Symbol)),
			Subtitle:    commandPaletteFirstNonEmpty(strings.TrimSpace(ref.Name), "Watchlist symbol"),
			Description: "Open symbol in Quote and select it in the watchlist.",
			Meta:        commandPaletteSymbolMeta(ref, true),
			Symbol:      ref,
		})
		if len(items) == 8 {
			break
		}
	}
	return items
}

func commandPaletteMatchesSymbol(ref domain.SymbolRef, needle string) bool {
	if needle == "" {
		return true
	}
	symbol := strings.ToLower(strings.TrimSpace(ref.Symbol))
	name := strings.ToLower(strings.TrimSpace(ref.Name))
	return strings.HasPrefix(symbol, needle) || strings.Contains(symbol, needle) || strings.Contains(name, needle)
}

func commandPaletteSymbolMeta(ref domain.SymbolRef, watchlist bool) string {
	parts := []string{"Symbol"}
	if watchlist {
		parts = append(parts, "watchlist")
	}
	if exchange := strings.TrimSpace(ref.Exchange); exchange != "" {
		parts = append(parts, exchange)
	}
	if typ := strings.TrimSpace(ref.Type); typ != "" {
		parts = append(parts, typ)
	}
	return strings.Join(parts, " • ")
}

func (m Model) commandPaletteFunctions() []commandPaletteFunction {
	activeSymbol := strings.ToUpper(strings.TrimSpace(m.activeSymbol()))
	items := []commandPaletteFunction{
		{ID: "markets", Title: "Markets", Aliases: []string{"market", "dashboard", "home"}, Category: "Workspace", Description: "Open the global market board."},
		{ID: "quote", Title: "Quote", Aliases: []string{"research", "symbol"}, Category: "Workspace", Description: "Open the Quote workspace for the active symbol."},
		{ID: "news", Title: "News", Aliases: []string{"wire", "headlines"}, Category: "Workspace", Description: "Open the market news wire."},
		{ID: "screeners", Title: "Screeners", Aliases: []string{"screener", "scan"}, Category: "Workspace", Description: "Open the screener workspace."},
		{ID: "ai", Title: "AI", Aliases: []string{"assistant", "chat"}, Category: "Workspace", Description: "Open the AI workspace."},
		{ID: "chart", Title: "Chart", Aliases: []string{"price", "quote chart"}, Category: "Quote", Description: fmt.Sprintf("Open the chart view for %s.", activeSymbol)},
		{ID: "fundamentals", Title: "Fundamentals", Aliases: []string{"fundamental", "valuation", "fa"}, Category: "Quote", Description: fmt.Sprintf("Open the fundamentals view for %s.", activeSymbol)},
		{ID: "technicals", Title: "Technicals", Aliases: []string{"technical", "ta"}, Category: "Quote", Description: fmt.Sprintf("Open the technicals view for %s.", activeSymbol)},
	}
	if m.services.HasEconomicCalendar() {
		items = append(items, commandPaletteFunction{
			ID:          "calendar",
			Title:       "Calendar",
			Aliases:     []string{"economic calendar", "macro calendar", "events"},
			Category:    "Workspace",
			Description: "Open the global economic calendar.",
		})
	}
	if m.services.HasStatements() {
		items = append(items, commandPaletteFunction{
			ID:          "statements",
			Title:       "Statements",
			Aliases:     []string{"financials", "income", "balance sheet", "cash flow"},
			Category:    "Quote",
			Description: fmt.Sprintf("Open financial statements for %s.", activeSymbol),
		})
	}
	if m.services.HasInsiders() {
		items = append(items, commandPaletteFunction{
			ID:          "insiders",
			Title:       "Insiders",
			Aliases:     []string{"insider", "ownership"},
			Category:    "Quote",
			Description: fmt.Sprintf("Open insider activity for %s.", activeSymbol),
		})
	}
	if m.services.HasAnalystRecommendations() {
		items = append(items, commandPaletteFunction{
			ID:          "analyst",
			Title:       "Analyst Recommendations",
			Aliases:     []string{"anr", "analysts", "recommendations", "ratings", "upgrades", "downgrades"},
			Category:    "Quote",
			Description: fmt.Sprintf("Open analyst recommendation history for %s.", activeSymbol),
		})
	}
	if m.services.HasFilings() {
		items = append(items, commandPaletteFunction{
			ID:          "filings",
			Title:       "Filings",
			Aliases:     []string{"sec", "10-k", "10-q", "8-k", "edgar"},
			Category:    "Quote",
			Description: fmt.Sprintf("Open recent SEC filings for %s.", activeSymbol),
		})
	}
	if m.services.HasEarnings() {
		items = append(items, commandPaletteFunction{
			ID:          "earnings",
			Title:       "Earnings",
			Aliases:     []string{"results", "eps", "earnings history", "earnings date"},
			Category:    "Quote",
			Description: fmt.Sprintf("Open earnings history and estimates for %s.", activeSymbol),
		})
	}
	return items
}

func commandPaletteFunctionSubtitle(fn commandPaletteFunction, activeSymbol string) string {
	switch fn.Category {
	case "Quote":
		return "Active symbol: " + strings.ToUpper(strings.TrimSpace(activeSymbol))
	default:
		return fn.Description
	}
}

func matchCommandPaletteFunction(fn commandPaletteFunction, needle string) (int, bool) {
	if needle == "" {
		return 10, true
	}
	title := strings.ToLower(fn.Title)
	category := strings.ToLower(fn.Category)
	description := strings.ToLower(fn.Description)
	bestScore := 100
	matched := false
	if title == needle {
		bestScore = 0
		matched = true
	}
	if strings.HasPrefix(title, needle) {
		bestScore = min(bestScore, 1)
		matched = true
	}
	if strings.Contains(title, needle) {
		bestScore = min(bestScore, 2)
		matched = true
	}
	if strings.Contains(category, needle) {
		bestScore = min(bestScore, 3)
		matched = true
	}
	if strings.Contains(description, needle) {
		bestScore = min(bestScore, 4)
		matched = true
	}
	for _, alias := range fn.Aliases {
		value := strings.ToLower(strings.TrimSpace(alias))
		switch {
		case value == needle:
			bestScore = min(bestScore, 0)
			matched = true
		case strings.HasPrefix(value, needle):
			bestScore = min(bestScore, 1)
			matched = true
		case strings.Contains(value, needle):
			bestScore = min(bestScore, 2)
			matched = true
		}
	}
	return bestScore, matched
}

func (m Model) executeCommandPaletteSelection() (Model, tea.Cmd) {
	if item, ok := m.currentCommandPaletteItem(); ok {
		switch item.Kind {
		case commandPaletteItemSymbol:
			return m.openCommandPaletteSymbol(item.Symbol.Symbol)
		case commandPaletteItemFunction:
			return m.executeCommandPaletteFunction(item.FunctionID)
		}
	}
	query := strings.TrimSpace(m.commandInput.Value())
	if symbol, ok := normalizeDirectSearchSymbol(query); ok {
		return m.openCommandPaletteSymbol(symbol)
	}
	return m, nil
}

func (m Model) openCommandPaletteSymbol(symbol string) (Model, tea.Cmd) {
	m.addToWatchlist(symbol)
	m.selectSymbol(symbol)
	m.globalPageOpen = false
	m.closeCommandPalette("Selected " + strings.ToUpper(strings.TrimSpace(symbol)))
	tabCmd := m.setActiveTab(tabQuote)
	m.setQuoteCenterMode(quoteCenterChart)
	return m, tea.Batch(tabCmd, m.persistCmd(), m.loadAllCmd(symbol))
}

func (m Model) executeCommandPaletteFunction(id string) (Model, tea.Cmd) {
	activeSymbol := m.activeSymbol()
	normalizedID := strings.ToLower(strings.TrimSpace(id))
	if normalizedID != "calendar" {
		m.globalPageOpen = false
	}
	switch normalizedID {
	case "markets":
		m.closeCommandPalette("Opened Markets workspace")
		return m, m.setActiveTab(tabMarkets)
	case "calendar":
		m.closeCommandPalette("Opened economic calendar")
		m.globalPageOpen = true
		m.globalPageKind = globalPageCalendar
		m.calendarFilter = calendarFilterToday
		if data, ok := m.cachedCalendar(m.calendarFilter); ok {
			m.calendar = data
			if len(m.calendar.Events) == 0 || m.calendarSel >= len(m.calendar.Events) {
				m.calendarSel = 0
			}
			return m, nil
		}
		m.calendar = domain.EconomicCalendarSnapshot{}
		m.errCalendar = nil
		m.calendarSel = 0
		return m, m.loadCalendarCmd(m.calendarFilter)
	case "quote":
		m.closeCommandPalette("Opened Quote workspace")
		return m, m.setActiveTab(tabQuote)
	case "news":
		m.closeCommandPalette("Opened News workspace")
		return m, m.setActiveTab(tabNews)
	case "screeners":
		m.closeCommandPalette("Opened Screeners workspace")
		return m, m.setActiveTab(tabScreener)
	case "ai":
		m.closeCommandPalette("Opened AI workspace")
		return m, m.setActiveTab(tabAI)
	case "chart":
		m.closeCommandPalette("Opened chart view")
		tabCmd := m.setActiveTab(tabQuote)
		m.setQuoteCenterMode(quoteCenterChart)
		return m, tabCmd
	case "fundamentals":
		m.closeCommandPalette("Opened fundamentals view")
		tabCmd := m.setActiveTab(tabQuote)
		m.setQuoteCenterMode(quoteCenterFundamentals)
		return m, tabCmd
	case "technicals":
		m.closeCommandPalette("Opened technicals view")
		tabCmd := m.setActiveTab(tabQuote)
		m.setQuoteCenterMode(quoteCenterTechnicals)
		if m.needsTechnicalHistory(activeSymbol) {
			return m, tea.Batch(tabCmd, m.loadTechnicalHistoryCmd(activeSymbol))
		}
		return m, tabCmd
	case "statements":
		m.closeCommandPalette("Opened statements view")
		tabCmd := m.setActiveTab(tabQuote)
		m.setQuoteCenterMode(quoteCenterStatements)
		return m, tea.Batch(tabCmd, m.loadStatementCmd(activeSymbol))
	case "insiders":
		m.closeCommandPalette("Opened insiders view")
		tabCmd := m.setActiveTab(tabQuote)
		m.setQuoteCenterMode(quoteCenterInsiders)
		return m, tea.Batch(tabCmd, m.loadInsidersCmd(activeSymbol))
	case "analyst":
		m.closeCommandPalette("Opened analyst recommendations view")
		tabCmd := m.setActiveTab(tabQuote)
		m.setQuoteCenterMode(quoteCenterAnalyst)
		return m, tea.Batch(tabCmd, m.loadAnalystRecommendationsCmd(activeSymbol))
	case "filings":
		m.closeCommandPalette("Opened filings view")
		tabCmd := m.setActiveTab(tabQuote)
		m.setQuoteCenterMode(quoteCenterFilings)
		return m, tea.Batch(tabCmd, m.loadFilingsCmd(activeSymbol))
	case "earnings":
		m.closeCommandPalette("Opened earnings view")
		tabCmd := m.setActiveTab(tabQuote)
		m.setQuoteCenterMode(quoteCenterEarnings)
		return m, tea.Batch(tabCmd, m.loadEarningsCmd(activeSymbol))
	default:
		return m, nil
	}
}

func commandPaletteFirstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
