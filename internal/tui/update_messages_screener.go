package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"blackdesk/internal/application"
)

func (m Model) handleScreenerLoaded(msg screenerLoadedMsg) (Model, tea.Cmd) {
	m.errScreener = msg.err
	result := application.ApplyScreenerLoad(application.ScreenerLoadInput{
		CurrentResult: m.screenerResult,
		SelectedIndex: m.screenerSel,
		UserTriggered: msg.userTriggered,
		Data:          msg.data,
		Err:           msg.err,
	})
	m.screenerLoaded = m.screenerLoaded || result.Loaded
	m.screenerResult = result.Result
	m.screenerSel = result.SelectedIndex
	m.status = result.Status
	m.ensureScreenerSelectionVisible()
	return m, nil
}

func (m Model) handleSearchLoaded(msg searchLoadedMsg) (Model, tea.Cmd) {
	if !m.searchMode || msg.id != m.searchRequestID || msg.query != m.searchRequestQuery {
		return m, nil
	}
	m.searchItems = msg.results
	m.searchIdx = 0
	if msg.err != nil {
		m.status = msg.err.Error()
	} else if len(msg.results) == 0 {
		m.status = "No search results"
	} else {
		m.status = fmt.Sprintf("%d search results", len(msg.results))
	}
	return m, nil
}

func (m Model) handleSearchDebounced(msg searchDebouncedMsg) (Model, tea.Cmd) {
	if !m.searchMode || msg.id != m.searchDebounceID {
		return m, nil
	}
	query := strings.TrimSpace(m.searchInput.Value())
	if query == "" || query != msg.query || query == m.searchRequestQuery {
		return m, nil
	}
	m.status = "Searching…"
	m.searchRequestID++
	m.searchRequestQuery = query
	return m, m.searchCmd(query, m.searchRequestID)
}
