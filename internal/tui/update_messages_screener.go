package tui

import (
	"fmt"

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
