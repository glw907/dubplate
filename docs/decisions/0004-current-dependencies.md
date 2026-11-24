# 4. Current dependencies

## Status

Accepted

## Context

Geoff ruled on 2026-10-04 that dubplate releases with no dependency debt and uses what its
libraries offer. The first sweep under the rule is
`docs/research/2026-10-04-dependency-sweep.md`. The XML tier's goldens carry libxml2's messages
word for word, so an unplanned libxml2 upgrade breaks them.

## Decision

Development runs on the newest production version of every library, tool, and pinned binary. Pins
stay exact, at the newest version, in `go.mod`, `web/package.json`, `tools/venv/requirements.txt`,
and `.github/workflows/check.yml` (the gate's binaries and their SHA-256 values).

The sweep runs once a milestone, as the first task of the milestone's first spec, over those files
and the pinned binaries: rclone, golangci-lint, shellcheck, ruff, gitleaks, vale, Playwright, and
libxml2. The sweep never covers the GitHub Actions the workflows pin by SHA; `.github/dependabot.yml`
watches those alone, and the sweep takes its pull request. A new capability that replaces
hand-rolled code is taken now or filed to the milestone that first leans on it, and a workaround for
an old version never outlives the bump that fixes it.

Minors and patches land without asking. A major waits for Geoff and is recorded with its unblock
condition. A spec that adopts a library names its newest version.

libxml2 (`/home/linuxbrew/.linuxbrew/bin/xmllint`, 2.15.4) is held by `brew pin libxml2`, so a bump
is deliberate.

## Consequences

The `dubplate-build` skill's "Dependency sweep" bullet holds the sweep's steps, and its "Gate"
section holds the libxml2 bump steps. A sweep with nothing to move is one line in the pull request
body.
