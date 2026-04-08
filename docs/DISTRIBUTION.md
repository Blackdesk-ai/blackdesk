# Distribution Strategy

Blackdesk can support several install channels, but they do not all have the same maintenance cost or automation path.

This document defines the recommended rollout order and the GitHub Actions model behind it.

## Current Live Channels

These distribution paths are live today:

1. GitHub Releases
2. `curl -fsSL https://blackdesk.ai/install | bash`
3. `blackdesk upgrade`

Other package-manager commands may appear on the website as intended install surface, but they should not be treated as live until their corresponding feed or package publication is actually in place.

## Recommended Release Model

Use pull requests for release intent, not direct pushes to the default branch.

- every pull request runs CI
- every non-draft pull request builds release-preview artifacts
- a merged pull request with a release label creates the next version tag
- the production release workflow runs from the tag, not from a branch push

Supported release labels:

- `release:patch`
- `release:minor`
- `release:major`
- `release`

`release` defaults to a patch release.

## How Production Releases Work

The current release flow is:

1. Open a pull request to `main`.
2. Add one of these labels:
   - `release`
   - `release:patch`
   - `release:minor`
   - `release:major`
3. Merge the pull request.
4. The `Tag Release From PR` workflow creates the version tag and publishes the GitHub Release.
5. `blackdesk.ai/install` resolves the latest published GitHub Release and installs the matching asset for the local platform.
6. installed binaries can move to the latest published release with `blackdesk upgrade`

Version behavior:

- first release with `release:minor` becomes `v0.1.0`
- `release` behaves like `release:patch`
- later releases increment from the latest existing `v*` tag

## Manual Release Recovery

If a tag already exists but the release workflow needs to be rerun:

1. use the `Release` workflow through `workflow_dispatch`
2. provide an explicit existing tag such as `v0.1.0`

The manual workflow is intentionally restricted to an explicit existing tag.
It does not create a new version on its own and it does not run against arbitrary branch heads.

## Distribution Priority

Start with the channels that fit a Go binary project naturally and can be kept reliable:

1. GitHub Releases
2. `curl -fsSL https://blackdesk.ai/install | bash`
3. Homebrew tap
4. Scoop bucket
5. Chocolatey
6. AUR
7. Nix package or flake

## Channel Notes

### GitHub Releases

This is the source of truth for versioned artifacts.
All other package-manager installs should ultimately resolve to release assets produced from tagged builds.

### `curl | bash`

The intended public command is:

```bash
curl -fsSL https://blackdesk.ai/install | bash
```

The `blackdesk.ai/install` endpoint should serve the same logic as `scripts/install.sh`, with `Blackdesk-ai/blackdesk` baked in.

### In-App Upgrade

Installed binaries can update themselves with:

```bash
blackdesk upgrade --check
blackdesk upgrade
```

The in-app updater resolves the latest GitHub Release, downloads the platform-matching archive, verifies its checksum, and replaces the current executable.
The TUI status bar also shows the current version and surfaces `vCurrent -> vLatest` when a newer release is available.

### Homebrew

This is a strong fit for Blackdesk.
The website install surface currently presents:

```bash
brew install blackdesk
```

That command is part of the intended public install surface.
Keep GitHub Releases and `blackdesk.ai/install` as the source of truth until the formula is published and validated.

### Scoop

This is a good Windows-native channel for a Go CLI/TUI.
It can be automated once a dedicated scoop bucket repository exists.

### Chocolatey

This is also a reasonable Windows distribution channel, but it adds more packaging and publishing overhead than Scoop.
It should be treated as secondary to Scoop unless Windows enterprise distribution is a priority.

### AUR

The website install surface currently presents:

```bash
yay -S blackdesk
```

The actual AUR package name has to match the package that is published.
Keep GitHub Releases and `blackdesk.ai/install` as the source of truth until the AUR package is live.

### Nix

Nix support is reasonable, but it should be treated as a packaging track of its own.
It is worth doing after GitHub Releases, Homebrew, and Scoop are stable.

### npm, bun, pnpm, yarn

These are not natural first-class distribution channels for a Go binary.
Supporting them cleanly usually means maintaining a JavaScript wrapper package that downloads the correct release asset during install.
That can be added later, but it should not be the first packaging priority.

### Official Homebrew core, official Arch repositories, and similar channels

These depend on external maintainers, ecosystem rules, and separate review processes.
They should be considered optional later-stage distribution, not day-one release requirements.
