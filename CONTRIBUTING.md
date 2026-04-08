# Contributing

Blackdesk is still evolving, but contributions are welcome once the public repository is created.

## Project Priorities

- keep the application local-first
- keep market data and AI runtimes behind replaceable abstractions
- keep the TUI dense, keyboard-first, and resilient under partial provider failure
- never couple the UI directly to Yahoo-specific payloads or one AI CLI implementation

## Before You Change Code

- read `README.md`
- read `docs/ARCHITECTURE.md`
- read `docs/OPEN_SOURCE_PREPARATION_PLAN.md`
- read `CODE_OF_CONDUCT.md`

## Development Expectations

- keep everything in English: code comments, tests, commit messages, and documentation
- prefer small, reviewable pull requests
- add or update tests when behavior changes
- keep provider-specific parsing, auth, and retry logic inside adapter packages
- normalize external payloads before they reach the TUI or AI context
- do not add screener data to AI context payloads
- respect the current layer boundary:
  `internal/tui` for presentation, `internal/application` for use cases
- do not split packages into more folders unless the new folder also represents a real Go package boundary with clear ownership
- keep the README market-data and usage notice accurate when public-facing behavior changes

## Repository Hygiene

- do not commit local machine artifacts, debug dumps, or editor state
- do not commit secrets, tokens, account identifiers, or unpublished infrastructure details
- do not introduce placeholder personal branding in release or packaging metadata
- assume contributions are submitted under the repository license unless agreed otherwise

## Recommended Workflow

1. Make the smallest change that solves the problem cleanly.
2. Run `gofmt -w` on changed Go files.
3. Run `go test ./...`.
4. Update public documentation when behavior or architecture changes.

## Scope Guidance

Good contribution areas:

- provider hardening
- fixture-based tests
- workspace polish and cleanup
- AI connector abstraction improvements
- resilience and error handling
- public documentation
- narrowly scoped package-boundary improvements with clear dependency benefits

Changes that should be discussed before implementation:

- major data-model changes
- new provider families
- release-process changes
- workflow changes for AI context or security-sensitive behavior
