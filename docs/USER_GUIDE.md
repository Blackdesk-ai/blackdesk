# User Guide

Blackdesk is a keyboard-first market research desk.
This guide explains how to get around the product, what each workspace is for, and the main flows a new user should learn first.

## Getting Started

Install Blackdesk:

```bash
curl -fsSL https://blackdesk.ai/install | bash
```

Run it:

```bash
blackdesk
```

Useful CLI commands:

- `blackdesk -v` or `blackdesk --version` prints the installed version
- `blackdesk -h` or `blackdesk --help` shows CLI help

Useful maintenance commands:

- `blackdesk upgrade --check` checks whether a newer published release is available
- `blackdesk upgrade` upgrades the installed binary to the latest published release

## Status Bar

The bottom status bar always shows:

- the active market-data source
- the active AI model or unset state
- the current app version

If a newer release is available, the version segment changes from `vCurrent` to `vCurrent -> vLatest`.

## Core Navigation

The product is organized into five workspaces:

1. `Markets`
2. `Quote`
3. `News`
4. `Screeners`
5. `AI`

Use these global controls everywhere:

- `Tab` cycles workspaces
- `1-5` jumps directly to a workspace
- `Ctrl+K` opens the command palette
- `/` opens symbol search
- `?` opens help
- `q` quits the app

## First Run Flow

If you are new to Blackdesk, this is the fastest way to understand the desk:

1. Open `Quote`.
2. Press `/` and search for a symbol.
3. Press `Enter` to load the selected result.
4. Move through `c`, `f`, `t`, `s`, and `h` to inspect the active symbol.
5. Press `Tab` to jump into `Markets`, `News`, and `Screeners`.
6. Open `AI` and ask for a summary of the active setup.

## Search

Search is the fastest way to load a symbol into the desk.

- `/` opens search
- type a ticker or company name
- results appear after a short pause while typing
- `Enter` on a ticker-like query loads that symbol immediately
- `↑ / ↓` moves through results
- `Enter` on a selected result loads that symbol into the desk
- `Ctrl+A` adds the selected result to the watchlist
- `Esc` closes search

Selecting a result updates the active symbol and refreshes the quote workspace.

## Command Palette

The command palette is the global launcher for functions and symbols.

- `Ctrl+K` opens the palette
- type a function name or symbol
- `↑ / ↓` moves through matches
- `Enter` opens the selected function or symbol
- `Esc` closes the palette

Common palette functions include:

- `Calendar`
- `Chart`
- `Fundamentals`
- `Technicals`
- `Statements`
- `Insiders`
- `Equity Research` or `er`
- `Analyst Recommendations` or `anr`
- `Filings`
- `Earnings`

Use `/` when you want the fastest path to a symbol.
Use `Ctrl+K` when you want a broader launcher across workspaces and symbol results.

## Workspaces

## Markets

`Markets` is the broad tape view.
Use it before drilling into one name.

What it is for:

- watching cross-asset movement
- checking breadth and regional context
- spotting regime shifts before opening a specific quote

Main controls:

- `i` generates an AI market insight
- `r` refreshes market data

## Calendar

`Calendar` is the global macro and economic events page.
It is separate from `Quote` on purpose.

What it is for:

- tracking high-importance economic releases
- scanning the current day quickly
- seeing the next week of macro catalysts

Main controls:

- `← / →` switches `Today` and `This Week`
- `↑ / ↓` moves through events
- `r` refreshes the calendar
- `Esc` closes the page

## Quote

`Quote` is the active symbol workflow.
This is where Blackdesk spends most of its time.

What it is for:

- chart and timeframe review
- fundamentals and company context
- technicals and trend state
- statements and insiders
- analyst recommendation review from the command palette
- filings and earnings review from the command palette
- symbol-specific news and AI insight

Main controls:

- `↑ / ↓` moves through the watchlist
- `c` opens chart view
- `f` opens fundamentals
- `t` opens technicals
- `s` opens statements
- `h` opens insiders
- `← / →` changes chart timeframe or statement kind
- `[ / ]` changes statement frequency
- `n` moves to the next quote news story
- `p` scrolls the company description
- `o` opens the selected quote news story in the browser
- `d` removes the selected symbol from the watchlist
- `i` generates AI insight for the active symbol
- `r` refreshes symbol data

Additional Quote pages:

- `Analyst Recommendations` is opened from the command palette or by typing `anr` and uses a fullscreen research layout for broker rating changes, recommendation trend, and current consensus targets.
- `Filings` is opened from the command palette and uses a fullscreen research layout for recent SEC filings.
- `Earnings` is opened from the command palette and uses the same fullscreen layout for reported quarters, upcoming estimates, and EPS trend context.

Quote fundamentals note:

- `QARP Score` means `Quality at a Reasonable Price`.
- It is a fast composite signal shown under the `Valuation` block in `Fundamentals`.
- Blackdesk calculates it as `Earnings Yield x ROIC`.
- `Earnings Yield` is derived from trailing PE when available, with EPS over price as fallback.
- `ROIC` is return on invested capital.
- If either `Earnings Yield` or `ROIC` is negative, the score is forced negative so weak fundamentals cannot display as a strong positive result.
- The value is displayed as a plain number instead of a percent. Example: `1.35` means the multiplied result scaled for readability in the UI.
- The value is color-coded by threshold:
- `< 0.5` weak / overvalued
- `0.5 - 0.8` fair
- `0.8 - 1.2` good
- `1.2 - 1.5` very good
- `> 1.5` rare / opportunity
- `R40` is Blackdesk's `Revenue Growth + Profit Margin` readout shown directly under `QARP Score`.
- It is displayed as a percent and color-coded by threshold:
- `< 15%` weak
- `15% - 25%` mediocre
- `25% - 40%` good
- `40% - 60%` very good
- `> 60%` exceptional
- Use it as a shortcut for balancing quality and price, not as a standalone investment verdict.

Global pages from the command palette:

- `Calendar` opens a fullscreen economic calendar with `Today` and `This Week` filters for high-importance global events.

## News

`News` is the market-wide wire.
It is separate from quote-specific news on purpose.

What it is for:

- scanning the current market narrative
- reading headlines without leaving the desk
- tracking broad moves not tied to one symbol

Main controls:

- `↑ / ↓` navigates stories
- `n / p` goes to next or previous story
- `o` opens the selected story in the browser
- `r` refreshes the feed

## Screeners

`Screeners` is for discovery.
Use it when you want to find candidates instead of analyze one name you already know.

What it is for:

- finding movers
- rotating through predefined discovery presets
- sending interesting names into the watchlist or quote workflow

Main controls:

- `↑ / ↓` navigates screener results
- `← / →` changes screener preset
- `n / p` moves to next or previous screener
- `a` adds the selected symbol to the watchlist
- `Enter` opens the selected result in `Quote`
- `r` refreshes the screener

## AI

`AI` is the desk-aware chat workspace.
It uses local connectors and Blackdesk context rather than raw upstream payloads.

What it is for:

- summarizing the active setup
- asking for market context
- reviewing fundamentals, technicals, or statement trends
- running a structured `Equity Research` memo for the active symbol from the command palette
- switching between local connectors and available models

Command palette AI functions:

- `Equity Research` opens `AI`, injects a full structured investment-research prompt for the active symbol, and starts the AI run immediately.
- It uses the current Blackdesk app context, including market regime, quote data, fundamentals, technicals, statements, news, and derived metrics available in the AI snapshot.

Main controls:

- `.` focuses the AI input and sends the prompt when already focused
- start typing to focus the AI input
- `Enter` sends the prompt when the input is focused
- `c` opens connector and model selection from the AI workspace
- `↑ / ↓` scrolls the transcript
- `f` toggles AI fullscreen
- `x` clears the conversation

## AI Picker

The AI picker only opens inside the `AI` workspace.

Controls:

- `↑ / ↓` cycles connectors or models
- `← / →` switches between connector and model steps
- `Enter` confirms the current selection
- `Esc` or `.` closes the picker

## Watchlist Workflow

The watchlist is central to the desk.

Typical flow:

1. Search for a symbol with `/`.
2. Add it to the watchlist.
3. Move through names with `↑ / ↓` in `Quote`.
4. Use `d` to remove symbols you no longer track.

## Browser Actions

Blackdesk stays terminal-first, but some flows intentionally jump out to the browser:

- `o` opens the selected news story

Use this for full articles and source pages after triaging headlines inside the desk.

## Help Inside The App

Press `?` at any time to open the built-in shortcut overlay.
Press `?`, `Esc`, or `q` again to close it.

## Product Boundaries

Blackdesk is a research tool.
It is not a brokerage, execution venue, or investment-advice product.

- market data can be delayed or incomplete
- provider behavior can change
- always verify critical decisions against authoritative sources
