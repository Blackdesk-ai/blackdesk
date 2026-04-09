# FAQ

## What is Blackdesk?

Blackdesk is a local-first, keyboard-first market research terminal.
It brings quotes, technicals, news, fundamentals, statements, insiders, screeners, and local AI connectors into one desk.

## Is Blackdesk a trading platform?

No.
Blackdesk is a research tool, not a brokerage or execution platform.
It does not provide investment advice.

## Which platforms are supported?

Blackdesk targets:

- macOS
- Linux
- Windows

Release artifacts and install channels may become available at different times depending on packaging rollout.

## How do I install it?

Primary install path:

```bash
curl -fsSL https://blackdesk.ai/install | bash
```

You can also use the local installer script from the repository:

```bash
./scripts/install.sh
```

## How do I update Blackdesk?

Check for a newer published release:

```bash
blackdesk upgrade --check
```

Upgrade the installed binary:

```bash
blackdesk upgrade
```

## How do I see the available CLI commands?

Use any of these:

```bash
blackdesk -h
blackdesk --help
blackdesk ?
```

Print the installed version:

```bash
blackdesk -v
blackdesk --version
```

## How do I start using it?

A simple first-run flow:

1. Launch `blackdesk`.
2. Press `/` to search for a symbol.
3. Press `Enter` to load the selected result.
4. Use `c`, `f`, `t`, `s`, and `h` in `Quote`.
5. Use `Tab` or `1-5` to move across workspaces.
6. Open `AI` and ask for a summary of the active setup.

## Where can I find the keyboard shortcuts?

See `docs/KEYBOARD_SHORTCUTS.md`.

## Where can I find the product guide?

See `docs/USER_GUIDE.md`.

## What data source does Blackdesk use?

Yahoo Finance is the first market-data adapter.
It is treated as unofficial and fallible, so some data may be delayed, incomplete, or temporarily unavailable.

## Does Blackdesk send my data to a hosted AI service?

Blackdesk is built around local AI connector workflows.
Connector behavior depends on the local CLI you configure and the runtime behind that CLI.

## Does AI see screener data?

No.
Screener workspace data is intentionally excluded from AI context payloads.

## Can I use multiple AI connectors?

Yes.
Blackdesk supports connector selection in the AI workspace.
Model discovery depends on what each local CLI exposes.

## Why is a feature unavailable in my current view?

Some actions are workspace-specific.
For example:

- AI picker opens from the `AI` workspace
- symbol-focused views live in `Quote`
- story navigation depends on whether you are in `News` or quote-specific news

## Why is a symbol not loading correctly?

Possible causes include:

- provider instability
- temporary rate limiting
- missing or partial upstream data
- symbol-specific coverage gaps

Try refreshing with `r`, searching again, or checking later.

## Why are some install commands shown publicly but not live yet?

Blackdesk exposes an intended public install surface, but some package-manager channels require separate publishing infrastructure, repository feeds, or external review before they become live.

## How do I report a bug or security issue?

- for product or usage issues, open a GitHub issue once the public repository is live
- for security issues, follow `SECURITY.md`
