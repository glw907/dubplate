# Standing rules

Detail that CLAUDE.md summarizes. Verbatim from CLAUDE.md as of the lean cutover.

## Build for 0.8

**Build for 0.8 (Geoff, 2026-09-07, re-keyed 2026-09-25, re-planned
2026-10-05).** Rung 12 is the built-in OpenSubsonic streamer, slices 1 to 8
of `docs/superpowers/specs/2026-09-07-streamer-2-0-design.md` (the spec
still says 2.0). Every other rung plan weighs its surface against that
spec's forward-hooks list and states per task whether it lands a hook or is
neutral; a rung 6 schema built without 0.8 in view is rework. Hook 6
retires and hook 5 is renamed to `username`. Route newly discovered hooks
into that list; never build 0.8 features early. The same spec carries a prior-art list (the
OpenSubsonic specification, Navidrome and gonic for behavior only under
GPL, taglib via wazero, modernc sqlite, goose, mewkiz/flac, Picard's tag
tables, chromaprint); a rung plan names which entries it consulted and
adopts or declines each with a reason, so no rung rebuilds what a
license-compatible library already does well. An adoption is judged on
license, code quality against the house bar, maintenance (release
cadence, open-issue age, bus factor), and adoption (importers, who
depends on it), with the evidence recorded in the plan, never on
familiarity alone. **Refined (Geoff, 2026-09-12):** a 0.8 forward hook is
named in the rung plan and built by the rung that first consumes it,
and nothing is exported ahead of a caller.

## Current dependencies

**Current dependencies (Geoff, 2026-10-04).** Development runs on the
newest production version of every library, tool, and pinned binary, so
dubplate releases with no dependency debt and uses what its libraries
offer. Pins stay exact, at the newest version. Every pass opens with a
`dependency-upgrade` sweep over `go.mod`, `web/package.json`,
`tools/venv/requirements.txt`, and the pinned binaries (rclone, golangci-lint, Playwright, libxml2); a sweep with nothing to move
is one line in the pass record. The skill's survey is the point: a new
capability that replaces hand-rolled code is ruled take-now or filed to
the rung that first leans on it, and a workaround for an old version
never outlives the bump that fixes it. Minors and patches land without
asking; a major waits for Geoff and is recorded with its unblock
condition. A rung plan that adopts a library names its newest version.
libxml2 (`/home/linuxbrew/.linuxbrew/bin/xmllint`, 2.15.4) is held
by `brew pin libxml2`, so a bump is deliberate: `brew unpin libxml2 &&
brew upgrade libxml2`, regenerate the XML goldens, `brew pin libxml2`,
then upgrade the eight held dependents.
The last sweep: `docs/superpowers/research/2026-10-04-dependency-sweep.md`.

## Gate legs

`bash scripts/check.sh` before every commit — `make check` (vet,
golangci-lint, test, tidy-check) over the whole module (`./...`;
`_salvage/` is invisible to the go tool), shellcheck over `scripts/*`,
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
first large data directory), vale (Go comment prose, `_salvage/`
excluded), a pattern-filtered `systemd-analyze verify` (hard-required:
a missing `systemd-analyze` fails the gate), and the state-dir agreement
leg, which lives in `scripts/check-state-dir.sh` (the full gate runs it, it
runs alone in seconds, and shellcheck covers it with the rest of
`scripts/*`): the server env template, every `dubplate-server` unit, and
`deploy.sh`'s `DUBPLATE_SERVER_LEAVES` name the same state directory and
leaves; every granted leaf is in that array and every array entry is
granted; the serve unit is granted the `store` leaf; and the leg proves
`deploy.sh`'s pre-create loop iterates that array.
Two import-graph tests hold the OpenSubsonic layer's boundaries, and a green
`go test ./...` runs both. The import-boundary test
(`internal/opensubsonic/boundary_test.go`) fails if anything under
`internal/opensubsonic` pulls in a module-internal package outside it, tests
included. The no-link test (`internal/harness/conform/arch_test.go`) fails if
`go list -deps ./cmd/dubplate-server` names a package under `internal/harness`
or `internal/opensubsonic/conformance`, or any `github.com/getkin` path, so the
server never links the harness, the contract checker, or kin-openapi; the
protocol client stays linkable.
Three workstation prerequisites sit outside the gate, because the gate
never opens a network connection: `npm ci --prefix web`,
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
theme values). Since rung 5c-iii Task 3a, the leg also runs the rendered
(browser) gates: a Playwright 1.64.0 script under a `rendered`-tagged Go
test, two builds compared byte for byte, and first paint per theme, so
every full gate launches Chromium and runs on the heavy lane (about 106s,
about 560 MiB peak at 3a). Rung 6b-i's merge adds an `rclone`-dependent leg: `internal/backup`'s
tests exercise the real `rclone` binary (`internal/backup`'s share of
`go test ./...` runs about 8.3s of it). A check either
gates or does not exist; never guard one with `command -v || skip`.
Never commit red. The process-citation leg
(no rung, task, resolution, or plan-line citation in shipped Go comments)
and the exported-unused leg (an exported `internal/` symbol whose only
references are its own file and `_test.go` files) are staged: they warn
today and gate at the comment-cleanup task.

## Lessons the code cannot carry

Durable facts harvested from the retired per-pass ledger, for the places where
no test or comment sits next to the code they guard.

**Oracle harness (`internal/harness/subsonic`).**
- Navidrome's later libraries come from `POST /auth/login`, then `POST
  /api/library` with the `X-ND-Authorization` bearer. The stack leaves the first
  library named "Music Library", so nothing may resolve a folder by name.
- A `collection:` key in a session spec resolves by exact artist-set equality
  against the oracle's own answers, never by name or order.
- gonic's `getMusicDirectory` without `id`, or with `id=0`, answers error 10.
  Navidrome serves a per-track cover that differs from the album cover. gonic
  splits an album held in disc subfolders into two albums (Borealis in the seed).
  Tracks under one second report no duration on either oracle, so a fixture that
  exercises duration needs 2 s or more.
- The seed lacks lyrics, `nowPlaying`, and radio requests, so the contract flags
  that depend on them stay owed to a later recording.
- `go-openapi/jsonpointer` v1.0.2 drops `swag/jsonname`, so `go.mod` carries
  jsonpointer and no `swag`.
- gonic cannot serve under a sub-path on its own: `GONIC_PROXY_PREFIX` only suits
  a stripping proxy.
- The golden grades the grader, not the oracles. A ledger row never names an
  oracle-only difference; only a row naming `dubplate` as one side is live.

**Build for 0.8.** The streamer spec's hook numbers are stable: a retired hook
stays in place, marked retired, so no hook renumbers.

**Serve routing.** Serve routes through one origin-wrapped mux with a 404
default, and the CSRF proof covers every subtree registered on it. A new subtree
is covered only if it registers on that mux.

**Settings.** A consumer reads a setting before it enters `Handle.Update`, never
inside the callback. The rule that every read a write decides on runs inside the
callback governs the rows the write changes; a setting read just before the write
is the value in force for that operation.

**Web.** Prerendered documents locate themselves from wherever they are served
because `paths.relative` stays at its default with no `reroute` hook and no
fallback page. The cost is that every route must answer its exact path: a
trailing slash gets the terminator's 404, not the document.

**Matcher and corpus.**
- The corpus never grades the MusicBrainz retry path: the stub answers 200 or 404
  only and `cassette.Request` carries no status field. The response-status
  extension stays declined until a case genuinely needs a 404 or 503.
- An absent release `status` is a plain, non-pseudo release in the client; beets
  drops it as a candidate. No corpus lookup omits one (303 of 303 carry it).

**Go tests.**
- Cobra's usage writer falls back to `os.Stderr` when only `SetErr` is wired, so a
  command test must wire `SetOut` to observe a leaked usage block.
- `doctor` row expectations derive from config's own tables (owned, reserved,
  `Thresholds()`), never from a literal count.

**Shell.** Bash cannot re-arm `errexit` inside a command-substitution subshell, so
a `set -e` placed there does nothing. Enforce best-effort calls by guarding with
`if` at the definition.
