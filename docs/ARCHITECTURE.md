# Architecture

Blackdesk is a local-first Go TUI for market research.
The project uses Bubble Tea for the terminal UI, but the long-term architecture is not classic MVC.

## Recommended Shape

The project is moving toward a layered, port-and-adapter structure:

- presentation: Bubble Tea update/view logic and reusable TUI components
- application: use cases and orchestration services
- domain: normalized market and AI-facing models
- ports: contracts for market data, news, storage, and AI runtimes
- adapters: Yahoo, RSS, local AI CLIs, filesystem storage, and in-memory cache
- bootstrap: dependency wiring and runtime composition

Bubble Tea remains MVU in the presentation layer.
The goal is to keep everything below that layer modular and replaceable.

## Current Repository Layout

```text
main.go
internal/
  bootstrap/
  application/
  agents/
  tui/
  buildinfo/
  domain/
  providers/
  storage/
  ui/
docs/
```

## Current Responsibilities

- `main.go`: process entry point and program startup
- `internal/bootstrap`: runtime dependency construction
- `internal/application`: use cases and orchestration services
- `internal/tui`: Bubble Tea model, workspace behavior, state reducers, and rendering helpers
- `internal/domain`: normalized types shared across the application
- `internal/providers`: provider interfaces plus concrete adapters
- `internal/agents`: local AI connector discovery and execution
- `internal/storage`: config and cache implementations
- `internal/ui`: rendering helpers shared across views

## What The Refactor Already Achieved

- runtime wiring is no longer owned by `main.go`
- fetch planning and orchestration have been moved into `internal/application`
- AI request preparation and execution are no longer concentrated in one root file
- workspace rendering and input handling are split across smaller modules inside `internal/tui`
- the TUI still uses Bubble Tea's MVU style, but the rest of the code now follows a clearer layered structure

## Remaining Structural Debt

- `internal/tui` is much cleaner than before, but a few medium-sized helpers remain
- provider interfaces still live next to concrete provider code instead of a dedicated `ports` package
- the current package layout is modular, but not yet a full ports-and-adapters directory split

## Package Layout Notes

`internal/tui` and `internal/application` are intentionally kept as flat Go packages for now.
That is a pragmatic trade-off:

- the TUI layer shares one central Bubble Tea model and many helpers depend on that shared state directly
- the application layer exposes closely related use-case helpers that are still cheap to navigate as one package
- splitting them into many subdirectories would create new packages, more import churn, and more cross-package plumbing

If either package starts growing in a way that creates real ownership boundaries, the next step should be subpackages by domain, not arbitrary folder splits.

That means the current layout should be read as an intentional Go design choice, not as unfinished organization.

## Provider Design Rules

- every provider adapter must normalize external payloads into internal domain types
- UI code must never depend on Yahoo-specific response fields
- auth, crumb, retry, and rate-limit handling belong inside the Yahoo adapter only
- market-wide news aggregation must remain separate from UI rendering concerns

## AI Connector Rules

- the TUI must depend on connector abstractions, not one CLI
- AI context must be built from normalized Blackdesk state
- screener data must never be included in AI context payloads
- local debug artifacts must remain opt-in, not default behavior

## Market Data And Usage Boundary

- Blackdesk is a research and workflow tool, not an execution platform
- unofficial or third-party data sources must be treated as fallible and non-authoritative
- user-facing documentation should keep a visible usage notice that the app does not provide investment advice
- engineering choices should prefer resilience and clear failure modes over pretending that upstream data is complete or guaranteed

## Practical Guidance

When making future changes:

1. keep use-case decisions in `internal/application`
2. keep the TUI focused on state, key handling, and rendering
3. keep external payload normalization inside provider and agent packages
4. update this document when the package boundaries change in a meaningful way
