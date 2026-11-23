# Standing rules

Detail that CLAUDE.md summarizes.

## Build for 0.8

**Build for 0.8 (Geoff, 2026-09-07, re-keyed 2026-09-25, re-planned
2026-10-05).** The streamer is the built-in OpenSubsonic streamer, slices 1 to 8
of `docs/specs/2026-09-07-streamer-2-0-design.md` (the spec
still says 2.0). Every other spec weighs its surface against that
spec's forward-hooks list and states per task whether it lands a hook or is
neutral; a store schema built without 0.8 in view is rework. Hook 6
retires and hook 5 is renamed to `username`. Route newly discovered hooks
into that list; never build 0.8 features early. The same spec carries a prior-art list (the
OpenSubsonic specification, Navidrome and gonic for behavior only under
GPL, taglib via wazero, modernc sqlite, goose, mewkiz/flac, Picard's tag
tables, chromaprint); a spec names which entries it consulted and
adopts or declines each with a reason, so no task rebuilds what a
license-compatible library already does well. An adoption is judged on
license, code quality against the house bar, maintenance (release
cadence, open-issue age, bus factor), and adoption (importers, who
depends on it), with the evidence recorded in the spec, never on
familiarity alone. **Refined (Geoff, 2026-09-12):** a 0.8 forward hook is
named in the spec of the task that first consumes it and built there,
and nothing is exported ahead of a caller.

The streamer spec's hook numbers are stable: a retired hook stays in place,
marked retired, so no hook renumbers.

## Current dependencies

**Current dependencies (Geoff, 2026-10-04).** Development runs on the
newest production version of every library, tool, and pinned binary, so
dubplate releases with no dependency debt and uses what its libraries
offer. Pins stay exact, at the newest version, and live in `go.mod`,
`web/package.json`, `tools/venv/requirements.txt`, and
`.github/workflows/check.yml` (the gate's binaries and their SHA-256
values). The sweep runs once a milestone, as the first task of the milestone's first
spec, over those files and the pinned binaries (rclone, golangci-lint,
shellcheck, ruff, gitleaks, vale, Playwright, libxml2); a sweep with nothing to move
is one line in the PR body. The sweep never covers the GitHub Actions the
workflows pin by SHA. `.github/dependabot.yml` watches those alone and opens
one grouped pull request a month, which the sweep takes through
`standards:dependency-upgrade`. The milestone's first task also runs
`make fuzz FUZZTIME=10m`. The skill's survey is the point: a new
capability that replaces hand-rolled code is ruled take-now or filed to
the milestone that first leans on it, and a workaround for an old version
never outlives the bump that fixes it. Minors and patches land without
asking; a major waits for Geoff and is recorded with its unblock
condition. A spec that adopts a library names its newest version.
libxml2 (`/home/linuxbrew/.linuxbrew/bin/xmllint`, 2.15.4) is held
by `brew pin libxml2`, so a bump is deliberate: `brew unpin libxml2 &&
brew upgrade libxml2`, regenerate the XML goldens, `brew pin libxml2`,
then upgrade the eight held dependents.
The last sweep: `docs/research/2026-10-04-dependency-sweep.md`.

## Gate legs

`bash scripts/check.sh` before every commit: `make check` (`web`, `vet`,
`lint`, `test`, `web-rendered`, `tidy-check`) over the whole module (`./...`), shellcheck over `scripts/*`,
ruff (`check` and `format
--check`) over every tracked `*.py` (the oracle shim and its derivation
siblings under `internal/harness/oracle/`, that package's FLAC template
generator, tagio's golden-FLAC generator under
`internal/engine/tagio/testdata/`, the matcher's featuring-golden generator
`internal/engine/matcher/testdata/gen_featuring_golden.py`, and
`tools/unidecode-table/generate.py`, now the oracle that re-derives the
transliteration table's pin on a Unidecode bump, never a source of a
committed table),
the photo pool integrity check (`node tools/photos/check.mjs integrity`),
gitleaks over
the tracked worktree (`--no-git`; the corpus under `corpus/cases/` is the repo's
first large data directory), vale (Go comment prose), a pattern-filtered `systemd-analyze verify` (hard-required:
a missing `systemd-analyze` fails the gate), and the state-dir agreement
leg, which lives in `scripts/check-state-dir.sh` (the full gate runs it, it
runs alone in seconds, and shellcheck covers it with the rest of
`scripts/*`): the server env template, every `dubplate-server` unit, and
`deploy.sh`'s `DUBPLATE_SERVER_LEAVES` name the same state directory and
leaves; every granted leaf is in that array and every array entry is
granted; the serve unit is granted the `store` leaf; and the leg proves
`deploy.sh`'s pre-create loop iterates that array. A second grant class admits
one non-leaf entry, on the serve unit alone: the music volume's directory,
which must be the common parent of the template's `DUBPLATE_STAGING_DIR` and
`DUBPLATE_LIBRARY_ROOT` (one mount keeps `rename(2)` from failing with `EXDEV`
between them). The serve unit must carry it, any other non-leaf grant still
fails, and the leg checks that `deploy.sh` has one pre-create line for each of
the two rows.
Two import-graph tests hold the OpenSubsonic layer's boundaries, and a green
`go test ./...` runs both. The import-boundary test
(`internal/opensubsonic/boundary_test.go`) fails if anything under
`internal/opensubsonic` pulls in a module-internal package outside it, tests
included. The no-link test (`internal/harness/conform/arch_test.go`) fails if
`go list -deps ./cmd/dubplate-server` names a package under `internal/harness`
or `internal/opensubsonic/conformance`, or any `github.com/getkin` path, so the
server never links the harness, the contract checker, or kin-openapi; the
protocol client stays linkable.
Three workstation prerequisites sit outside the gate, because no test
opens a network connection: `npm ci --prefix web`,
`make unidecode-wheel` (the verified Unidecode 1.4.0 wheel in the user
cache, which the unidecode tests and the install verb's tests read), and
`make subsonic-xsd` (the pinned Subsonic 1.16.1 XSD in the user cache,
read by the harness's XML tier and never committed). The XML tier also
needs `xmllint` at libxml 2.15.4 on `PATH`, held by `brew pin libxml2`
because its messages land in goldens; eight installed formulae depend on
it and are held with it (chafa, ffmpeg, gdk-pixbuf, imagemagick, librsvg,
libxkbcommon, shared-mime-info, xdotool).
The `web/` leg runs today, first in `make check`'s prerequisite chain:
Node and npm required and named on failure, a stale or drifted
`npm ci --prefix web` refused, the SvelteKit static bundle built, then
`svelte-check` under strict TypeScript and the static web gates (no
server-only file, no hardcoded color, no external origin, theme
variable completeness, theme contrast, and the built bundle's stamped
theme values). Since 2026-09-25, the leg also runs the rendered
(browser) gates: a Playwright 1.65.1 script under a `rendered`-tagged Go
test, two builds compared byte for byte, and first paint per theme, so
every full gate launches Chromium and runs on the heavy lane (about 106s,
about 560 MiB peak when first measured). The `store` work added an `rclone`-dependent leg: `internal/backup`'s
tests exercise the real `rclone` binary (`internal/backup`'s share of
`go test ./...` runs about 8.3s of it). A check either
gates or does not exist; never guard one with `command -v || skip`.
Never commit red. The PR check
(`.github/workflows/check.yml`) runs `bash scripts/check.sh` on every pull
request, so CI is the full-suite home, and `bash scripts/gate.sh --merge`
runs locally before a pull request is marked ready. The process-citation leg
gates: a shipped Go comment under `cmd` or `internal` that cites a rung,
task, resolution, or plan line fails the run. A wider pattern (contract,
ruling, fold, pass, and conductor tags, and a name) applies everywhere
except `cmd/dubplate-server`, `internal/serve`, `internal/auth`, and
`internal/intake`, whose comments BACKLOG #156 clears. The exported-unused leg
stays a warning, and it runs `go tool deadcode`, which reports a function
unreachable from any `main`.
The govulncheck leg runs `go tool govulncheck ./...` in every form but
`--reduced` and fails on any known vulnerability the module's code can
reach. It reads the Go vulnerability database, so it is the gate's one
network read, and it fails when the network is down. Two tools are
pinned by `tool` lines in `go.mod` and run through `go tool`: `deadcode`
and `govulncheck`. Every host and CI run the same version, and CI needs
no install step for either.
Two checks run in CI and stay out of the local gate. The race detector
(`make test-race`, `go test -race -count=1 -timeout 30m ./...`) runs in
`check.yml` as the "Run the race detector" step after the gate, because
the module takes over ten minutes under `-race`. A timed fuzz run
(`.github/workflows/fuzz.yml`, `make fuzz FUZZTIME=1m`) starts on each
pull request that touches `internal/engine/tagio`, `internal/intake/gate`,
the workflow, or the `Makefile`, and uploads any crashing input as an
artifact. The fuzz targets' seeds run in every `go test`. A crashing
input joins the seed corpus under the package's `testdata/fuzz` once its
traced cause is fixed. `make fuzz` runs each target for `FUZZTIME`
(default `30s`).
