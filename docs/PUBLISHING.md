# Publishing Checklist

Use this checklist when preparing the final history-free public repository.

## Repository Setup

1. Create the new GitHub repository at `Blackdesk-ai/blackdesk`.
2. Copy the cleaned working tree into the new repository without old git history.
3. Set the default branch and release permissions for the new repository.
4. Update `https://blackdesk.ai/install` to serve the same installer content as `scripts/install.sh`.

## Metadata Alignment

1. Keep `BLACKDESK_REPO=Blackdesk-ai/blackdesk` aligned across release documentation, install examples, and the website installer.
2. Confirm release archives and checksums match the names expected by `scripts/install.sh`.
3. Recheck `.goreleaser.yml`, README installation instructions, and release workflow files against `Blackdesk-ai/blackdesk`.
4. For versioned production releases, merge a pull request labeled `release`, `release:patch`, `release:minor`, or `release:major`.

## Final Validation

1. Run `gofmt -w` on any final changed Go files.
2. Run `go test ./...`.
3. Run `go vet ./...`.
4. Re-run repository hygiene scans for:
   - non-English text
   - local paths
   - debug dumps
   - placeholder branding or unpublished repository slugs
5. Verify the installer against a tagged release artifact after the first public release is created.

## Public-Facing Review

1. Confirm README, architecture, contributing, security, and license documents are present and current.
2. Confirm the market-data and usage notice remains visible in the README.
3. Confirm no screenshots, examples, or fixtures expose private prompts, local paths, or unpublished infrastructure details.
