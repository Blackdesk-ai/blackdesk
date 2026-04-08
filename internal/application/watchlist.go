package application

import (
	"strings"

	"blackdesk/internal/storage"
)

type WatchlistState struct {
	Config        storage.Config
	SelectedIndex int
	Scroll        int
}

func AddWatchlistSymbol(cfg storage.Config, selectedIndex, scroll, visibleRows int, symbol string) WatchlistState {
	out := WatchlistState{
		Config:        cloneConfig(cfg),
		SelectedIndex: selectedIndex,
		Scroll:        scroll,
	}
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return normalizeWatchlistState(out, visibleRows)
	}

	for i, item := range out.Config.Watchlist {
		if strings.EqualFold(item, symbol) {
			out.SelectedIndex = i
			out.Config.ActiveSymbol = symbol
			return normalizeWatchlistState(out, visibleRows)
		}
	}

	out.Config.Watchlist = append([]string{symbol}, out.Config.Watchlist...)
	if len(out.Config.Watchlist) > storage.MaxWatchlistItems {
		out.Config.Watchlist = out.Config.Watchlist[:storage.MaxWatchlistItems]
	}
	out.Config.ActiveSymbol = symbol
	out.SelectedIndex = 0
	out.Scroll = 0
	return normalizeWatchlistState(out, visibleRows)
}

func RemoveWatchlistSymbol(cfg storage.Config, selectedIndex, scroll, visibleRows int) WatchlistState {
	out := WatchlistState{
		Config:        cloneConfig(cfg),
		SelectedIndex: selectedIndex,
		Scroll:        scroll,
	}
	if len(out.Config.Watchlist) <= 1 {
		return normalizeWatchlistState(out, visibleRows)
	}
	if out.SelectedIndex < 0 {
		out.SelectedIndex = 0
	}
	if out.SelectedIndex >= len(out.Config.Watchlist) {
		out.SelectedIndex = len(out.Config.Watchlist) - 1
	}

	out.Config.Watchlist = append(out.Config.Watchlist[:out.SelectedIndex], out.Config.Watchlist[out.SelectedIndex+1:]...)
	if out.SelectedIndex >= len(out.Config.Watchlist) {
		out.SelectedIndex = len(out.Config.Watchlist) - 1
	}
	if len(out.Config.Watchlist) > 0 {
		out.Config.ActiveSymbol = out.Config.Watchlist[out.SelectedIndex]
	} else {
		out.Config.ActiveSymbol = ""
	}
	return normalizeWatchlistState(out, visibleRows)
}

func SetActiveSymbol(cfg storage.Config, symbol string) storage.Config {
	out := cloneConfig(cfg)
	out.ActiveSymbol = strings.ToUpper(strings.TrimSpace(symbol))
	return out
}

func SetDefaultRange(cfg storage.Config, timeRange, interval string) storage.Config {
	out := cloneConfig(cfg)
	out.DefaultRange = strings.TrimSpace(timeRange)
	out.DefaultInterval = strings.TrimSpace(interval)
	return out
}

func cloneConfig(cfg storage.Config) storage.Config {
	out := cfg
	out.Watchlist = append([]string(nil), cfg.Watchlist...)
	return out
}

func normalizeWatchlistState(state WatchlistState, visibleRows int) WatchlistState {
	total := len(state.Config.Watchlist)
	if total == 0 {
		state.SelectedIndex = 0
		state.Scroll = 0
		return state
	}
	if state.SelectedIndex < 0 {
		state.SelectedIndex = 0
	}
	if state.SelectedIndex >= total {
		state.SelectedIndex = total - 1
	}
	if visibleRows <= 0 {
		visibleRows = 1
	}
	maxStart := maxInt(0, total-visibleRows)
	if state.Scroll > maxStart {
		state.Scroll = maxStart
	}
	if state.Scroll < 0 {
		state.Scroll = 0
	}
	if state.SelectedIndex < state.Scroll {
		state.Scroll = state.SelectedIndex
	}
	if state.SelectedIndex >= state.Scroll+visibleRows {
		state.Scroll = state.SelectedIndex - visibleRows + 1
	}
	return state
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
