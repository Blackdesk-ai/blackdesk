# Security Policy

## Supported Scope

Security issues are relevant for:

- secrets or credential exposure
- local path leakage
- unsafe command execution behavior in AI connector integrations
- unsafe handling of external content from providers or feeds
- release or packaging issues that could affect end users

## Reporting

Until the public repository is live, report security issues privately to the maintainer through a private channel.
Do not open a public issue for a suspected vulnerability before it is reviewed.

## Repository Expectations

Blackdesk aims to keep the public repository free of:

- committed secrets
- local debug dumps
- private prompts or machine-specific paths
- unpublished infrastructure details

## Operational Notes

- AI debug dumps are disabled by default
- provider payloads should be normalized before reaching the UI
- release metadata should stay aligned with the public repository `Blackdesk-ai/blackdesk`
