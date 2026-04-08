# Blackdesk Open Source Preparation Plan

This document tracks the public-repository preparation work that has already landed and the final work that remains before the new history-free repository is created.

## Goals

- publish a clean, professional, English-only repository
- remove internal-only notes, local machine artifacts, and historical assumptions
- refactor the codebase toward a modular architecture that supports new providers and AI connectors without turning the TUI into a monolith
- keep provider payloads and CLI-specific behavior isolated behind internal abstractions

## Architecture Direction

Do not force a classic MVC structure onto a Bubble Tea application.
Blackdesk already uses a message-driven TUI framework, so the better target is:

- presentation: Bubble Tea models, views, keymaps, rendering helpers
- application: use cases and orchestration services
- domain: normalized market models and domain rules
- ports: interfaces for market data, news, AI runtimes, and storage
- adapters: Yahoo, RSS, local AI CLIs, filesystem config, cache implementations
- bootstrap: runtime wiring and dependency composition

The presentation layer keeps Bubble Tea's MVU style.
Everything below that layer should stay modular and replaceable.

## What Has Already Been Done

- runtime composition lives in `internal/bootstrap`
- orchestration use cases live in `internal/application`
- the TUI runtime has been split into smaller workspace-oriented modules inside `internal/tui`
- AI request preparation, AI runtime execution, market/news refresh planning, watchlist actions, and screener flows are no longer concentrated in one monolithic file
- local AI debug dumps are opt-in instead of default behavior
- local path leakage and old branding references have been removed from tracked files
- the repository is English-only across code, tests, comments, and public documentation

## Current Repository Shape

```text
main.go
internal/
  bootstrap/
  application/
  domain/
  providers/
  agents/
  storage/
  ui/
  tui/
```

In the current repository, `internal/tui` and `internal/application` are intentionally kept as flat packages.
That is a deliberate Go trade-off, not an incomplete cleanup.
Further folder splits should happen only when they create real package boundaries, not just cosmetic structure.

## Remaining Work Before the New Public Repository

### Final runtime polish

- keep shrinking medium-sized helpers only where the split improves clarity
- avoid churn-only refactors that do not improve maintainability
- keep `internal/tui` and `internal/application` flat unless a future change creates a genuine package boundary

### Provider and port cleanup

- decide whether provider and AI contracts should move into a dedicated `ports` package
- if that change lands, keep adapters and retries isolated in adapter-facing packages

### Release hardening

- align installer and release metadata with `Blackdesk-ai/blackdesk`
- create the new repository without old git history
- copy the cleaned working tree into that repository
- re-run the full hygiene checks after the new repository is initialized
- keep the README usage notice explicit about unofficial data and non-advisory scope

### Coverage and resilience

- add more fixture-based coverage around Yahoo normalization and edge cases
- keep market-wide news parsing diversified and resilient under source failure

## Public Repository Checklist

- no tracked local machine artifacts
- no tracked private prompts or debug dumps
- no personal paths, internal references, or placeholder repository slugs
- no non-English code comments, tests, or documentation
- no internal-only planning files at the repository root
- no user-facing documentation that depends on unpublished infrastructure
- a visible market-data and non-investment-advice notice in public-facing documentation
- Apache-2.0 is selected and included before the public repository is created
