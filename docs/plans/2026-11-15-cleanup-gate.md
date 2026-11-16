# Pass plan: a faster gate, with per-leg times

Initiative: dubplate-server
Round: cleanup (pass 2 of 5, under the conductor `conductor-dubplate`)
Pass: cleanup-gate

Design: `/var/home/glw907/Projects/initiatives/dubplate-server/specs/2026-11-15-cleanup-round-design.md`, approved by
Geoff in the Approval Queue on 2026-11-16 (document `20261116-000308-88f4d3`). Read in full: "Calls made here",
"Scope", "How it fails", pass 2's row under "Passes", "Acceptance" (pass 2, lines 208-234, and the every-pass items),
"Standing answers" 13 to 20, and "Backlog verdicts". Where this plan and the design differ, the design wins and the
difference is a stop.

Worktree: `/var/home/glw907/Projects/dubplate/.claude/worktrees/cleanup-gate`, branch `cleanup-gate`, cut from
`origin/master` at `7db541b`. PR base: `master`. Class: `auth-data` by the adapter's path map, since T2 edits
`cmd/dubplate-server/rendered_test.go` (the design says `runner`; see "Decisions for the executor", D1). Executor:
`dubplate-cleanup-gate`, `sonnet` at `medium` with the Fable advisor; T2 goes to an `opus` pair subagent. Design path
minutes: 130. This plan's path: 150 (see "Path"), under the 260-minute park line.

**This plan is parked at planning.** Acceptance item 3 cannot be met by the mechanism the design names, and changing
it is Geoff's call ("Stop at planning", below). Every other task is ready. When Geoff answers, the planner edits item
3 in the design per his choice, copies the new text into "Acceptance" and T2's check, runs the plan-reviewer read,
and launches.

## Stop at planning

**Question for Geoff.** Item 3's pixel clause and its 130 s line can't both be met by changing how browser contexts
open. Which of these should item 3 say?

**What was tried.** Two probes on pavilion against `7db541b` built the design's mechanism: one context per state and
scheme (38), with each width taken by `page.setViewportSize` and the existing `settleCapture`, and the manifest's
link spend moved to each state's first width. Probe 2 also took two screenshots per cell inside the page to test
whether paint was still moving. Probe 2 ran four times: twice as `go test` alone (2a, 2b) and twice as the whole
`make web-rendered` target, tagged vet and lint included (2c, 2d). The probe edits were reverted and the tree is
clean. The sample is small: two base runs and four 38-context runs.

| Measure | Base (190 contexts) | 38 contexts |
| --- | --- | --- |
| `go test -tags rendered` wall time | 194.2 s, 193.8 s | 136.5 s (probe 1); 154.1 s, 153.4 s with the extra in-page shot (2a, 2b) |
| Whole `make web-rendered` wall time | not timed | 155.1 s, 154.5 s with the extra in-page shot (2c, 2d); the shot adds about 17-19 s |
| Capture matrix | about 117 s, 115 s | 58.7 s; about 76 s with the extra shot |
| Before the matrix (build, fixture, non-capture rows) | about 77 s | about 78 s, untouched |
| Distinct capture sets among the runs | 2 sets from 2 runs, 6 of 95 files apart | 2 sets from 4 runs (2a = 2b, 2c = 2d), 2 of 95 files apart (`login-empty/390-dark`, `390-light`) |
| Files that differ from either base run | n/a | 21 of 95, in every 38-context run |
| Pixels per differing file | 2 to 3 | 2 to 6 |
| In-page double shots that differ | n/a | 1 of 190 in 2a (a sheet-only width), 0 of 190 in 2b |
| Spend reconciliation | passes | fails until the manifest moves spend to first-width rows; then passes |

Run 2d exited 2 on one console error, `net::ERR_NETWORK_CHANGED` loading `branding.css`. NetworkManager logged a
DHCPv6 lease renewal on `wlp4s0` at the same moment, so the host's network changed under Chromium. Its captures
match 2c's exactly.

Every differing pixel sits in a three-pixel column on an alert's left rule (x 32-34 at 390, x 432-434 at 1440, and
x 113-115 or 49-51 on the contact sheet), in the login error states and `interstitial/ready`. That holds for the
base-to-38 differences and for both run-to-run differences. Paint is settled within a page, so standing answer 18
("fixed at its cause") has no settle fix to make. The antialiasing of that edge varies with layout history (fresh
at a width or resized to it) and, more rarely, between runs of the same code.

The 130 s line fails on time too. The whole target with the extra shot ran 155 s; without it, about 137 s. About
78 s of the test comes before the matrix, and the plan has no cut there: "Only the way contexts open changes."

**Choices.**

1. *(Recommended.)* Keep the 38-context change and reword item 3: "The rendered leg writes 190 captures. Against
   the base commit's set, and between two runs on the close head, each file matches with no pixel tolerance or
   differs only on the antialiased alert-rule columns the PR lists. The leg opens 38 browser contexts where it
   opened 190, and its elapsed line on pavilion reads under 160 s." This keeps every assertion and halves the
   matrix. It gives up a byte-for-byte match that neither the base nor the new mechanism reproduces run to run.
2. Keep item 3's pixel clause and drop the context change from this pass. The rendered leg stays at about 194 s,
   item 3 moves to a later pass, and the rest of this pass lands. Even unchanged, a zero-tolerance base compare can
   fail: the one pair of base runs measured differed in 6 of 95 files on these same columns.
3. Add a probe task first, about 30 min: try Chromium launch flags that make a resized page paint the way a fresh
   one does, then choose between 1 and 2 on the result. It may find nothing.

## Pre-flight

No step needs Geoff before the launch, apart from the answer above. The pass needs no secret, sign-in, setting,
hook, or permission rule, and it never reaches the box.

Permissions checked: read `~/.local/state/claude-denials.jsonl` on 2026-11-16 against every command in the tasks.
It holds 110 rows, 2026-11-01T00:55 to 2026-11-16T07:43. The last 27 are "Classifier unavailable" rows from
2026-11-16 (this session and an ecxc-ski one): the auto-mode classifier returned no verdict, which is no rule, and
the answer is to re-issue once and do read-only work between tries. The earlier 83 are as pass 1's plan lists them
(`docs/archive/plans/2026-11-15-cleanup-ci.md`, "Pre-flight"), and seven touch this pass, each with its answer. `claude
stop` and `pkill` (Interfere With Workloads): the executor never types either and ends its own background work with
`TaskStop`. A gate or an install piped to `tail` (Destructive or Unwanted Actions): every gate and install runs bare.
`--no-verify` (Safety Bypass Flag): no commit hook exists (pass 1 mechanism 24), and no task passes the flag. A
`go run ...@latest` (Code from External): `deadcode` comes in through `go get -tool` at a pinned version, never
`@latest` at run time. A python heredoc edit of a `.claude/skills` file: this pass edits no skill. `gh workflow run`
(Auto-Mode Bypass): never; the check runs on `pull_request`. A push to the default branch (Modify Shared Resources):
every push is `git push origin cleanup-gate`. No row names `systemd-run --user --scope`, `go get -tool`, `go tool`,
`go mod tidy`, `git diff`, `node`, `npx playwright`, or `gh pr create --draft`.

Reviewed: pending. The `standards:plan-reviewer` read runs once Geoff answers, since the answer rewrites T2's
acceptance; its record goes under "Review record".

Evidence: the mechanism table, the recorded searches, and the probe table above. Each row is a probe this planner
ran on pavilion on 2026-11-15 or 2026-11-16 in this worktree at `7db541b`, or a quoted source line.

Run `npm ci --prefix web` first. A full local gate on master at `7db541b` passed (the conductor's STATUS commit on
top of pass 1's gated merge). The rendered test alone takes about 194 s here.

## What the pass changes

A full gate prints how long each leg took and a total, and the PR body compares a base run with a close run. The
rendered leg opens 38 browser contexts where it opened 190 and keeps all 190 captures and every assertion. Gitleaks
stops reading the installed and generated trees. A scoped run whose change set holds only Go outside the rendered
binary builds the bundle once, skips `svelte-check` and the static web gates, and takes the light lane. Vale reads
only the changed Go files on a scoped run. `gate.sh` accepts every flag `check.sh` accepts. The exported-unused leg
runs `go tool deadcode` in place of its awk block. Four backlog items fold in (#34, #64's header line, #67's
pattern, #69's wording), with #134's header rewording.

## Rules that bind every task

1. **Files.** Only pass 2's row: `scripts/check.sh`, `scripts/gate.sh`, `scripts/vale-comments.sh`, `Makefile`,
   `.gitleaks.toml`, `web/scripts/rendered/**`, `cmd/dubplate-server/rendered_test.go`, `go.mod`, `go.sum`,
   `BACKLOG.md` (closing this pass's items and filing finds), and this plan and its prompt (moved to
   `docs/archive/plans/` at the close). A call site a changed signature forces goes under `## Decisions`; anything
   else is filed with `standards:log-issue` in BACKLOG and the PR body (standing answer 15). Never touch
   `docs/STATUS.md` (the conductor's; wanted text goes in the PR body under `## STATUS text`) or `.claude/settings*`
   ("`.claude/settings*` is never an executor's file: record the rule the pass wants for the close packet and go
   on"). Never touch `.golangci.yml`, `tools/`, `systemd/`, `docs/standing-rules.md`, the adapter, or anything under
   the design's "Out" list.
2. **The box.** The round adds no feature, no migration, and no config key, and it deploys nothing. No task reaches
   the box (`5.78.148.42`), by ssh or otherwise.
3. **Gates.** The fast lane is `bash scripts/gate.sh --scope-changed`, no `--base`. The close is `bash scripts/gate.sh
   --merge`. Each runs bare from the worktree root with the Bash tool's `timeout: 600000`, re-issued on exit 75 until
   it prints `gate exit:`. Never pipe a gate. Never edit a file under `scripts/` while a gate runs it (`check.sh:21-23`).
   Never commit red. A task whose change set holds machinery (`scripts/*`, `Makefile`, `go.mod`, `web/scripts/rendered`)
   runs the full form under the fast lane; that is expected.
4. **Commits.** One commit per task, the task id first in the subject, imperative mood, specific files only (never
   `-A`), ending with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`. Never commit a secret or a `.env`
   value; read each diff by eye before staging. T6's plants are never committed.
5. **Pushes.** Only `git push origin cleanup-gate`. The draft PR opens on T1's push (`gh pr create --draft --base
   master`). After each push, watch CI as a background Bash task, bare: `gh pr checks <n> --watch --fail-fast`, with
   `gh run watch <run id> --exit-status` as the fallback. A red stops the line when it arrives. A red outside this
   pass's files is fixed in this branch and listed under `## Decisions` (standing answer 19).
6. **Records.** PR notes at `<git dir>/pr-notes.md`, first line `launch-base=<sha>`. Each task appends its gate exits,
   timings, mutation results, CI links, and anything found outside scope.
7. **Expected values** are literals from this plan or counts computed independently (`grep -c` over a file, the
   state and scheme lists' lengths read from source), never a figure the change itself printed.
8. **Mutation proofs are run.** Apply each, run the named test or gate, record the failure, revert. One that doesn't
   fail is a finding fixed in the same task.
9. **Smallest thing that meets acceptance.** No abstraction, export, flag, or fallback the design doesn't name. No
   leg gains a tool-presence skip (`command -v || skip`).
10. **Voice.** Shell comments follow the file's register; no em dash in a comment or commit. TypeScript comments
    follow `standards:ts-conventions`, Go comments `standards:go-conventions`. The PR's opening paragraph follows
    `~/.claude/docs/voice/written-for-people.md`.

## Mechanisms

| # | Mechanism | Source or probe | Result |
| --- | --- | --- | --- |
| 1 | No per-leg timing today | `grep -n -iE 'elapsed\|SECONDS\|date \+%s' scripts/check.sh` | prints nothing |
| 2 | The full form runs the Go and web legs as one `make check` | `check.sh:612-615`; `Makefile:149` `check: web vet lint test web-rendered tidy-check` | one header, no split; T1 runs the six targets as six timed legs in that order |
| 3 | Leg headers | `check.sh:584-916` print `== <leg> ==` | the name each elapsed line carries |
| 4 | `gate.sh` flags | `gate.sh:19-44` accepts `--scope-changed`, `--merge`, `--dry-run`, `--base <40-hex>`; anything else prints usage and exits 2 | `--reduced` is refused today (item 4) |
| 5 | `check.sh` flags | `check.sh:159-186`: `--scope-changed`, `--reduced`, `--merge`, `--base`, `--dry-run`, `-h` | the set `gate.sh` must accept, plus `--fresh` |
| 6 | `--fresh` belongs to the runner | `cairn-run-gate:37` "Usage: cairn-run-gate [--fresh] '<gate string>'"; `:13-14` "--fresh reruns the same tree" | `gate.sh` passes it before the gate string, never into `check.sh` |
| 7 | The lane is chosen before classification | `gate.sh:50-54` sets `CAIRN_GATE_LANE=light` only on `--dry-run` | a scoped run with no rendered leg queues heavy (#34) |
| 8 | The light cap | `cairn-run-gate:43-48` "2G high, 3G max, 1G swap" | probe: `golangci-lint run ./...` cold, 11 s, peak 1.57 GiB; `go test ./...` 71 s; `go tool deadcode ./...` 4 s, peak 2,148,667,392 B with 348 `high` events and no `oom` (wrapper `systemd-run --user --scope -q -p MemoryHigh=2G -p MemoryMax=3G -p MemorySwapMax=1G`, reading `memory.peak` and `memory.events`) |
| 9 | A scoped run builds the bundle twice and runs the web checks | `check.sh:584-588` runs `make web`; `Makefile:59-118` builds twice, then `npm run check` and eleven static gates | item 6's target |
| 10 | The Go tests need a current bundle | `Makefile:56-58` "the Go tests that read the embedded bundle need it current before `go test` runs" | a scoped run still builds once |
| 11 | The rendered scope | `check.sh:407-419` `load_rendered_pkg_dirs`; `:441-480` `classify_change_set` | `rendered=skipped` on the scope line when no web path and no rendered-linked package changed |
| 12 | The six prerequisite blocks | `check.sh:218-241`: `go`, `golangci-lint`, `shellcheck`, `ruff`, `gitleaks`, `systemd-analyze`, each `command -v` then exit 1 | one loop that names every missing tool, then exits 1 |
| 13 | Vale reads every Go file | `vale-comments.sh:53-56` lists all tracked Go files; `:25-33` accepts only `--check-vendor`, unknown exits 2; `check.sh:769-772` calls it with no argument | item 7's target |
| 14 | Gitleaks reads ignored trees | `.gitleaks.toml` has no path allowlist (`grep -n -iE 'node_modules\|rendered-captures\|venv' .gitleaks.toml` prints nothing); `check.sh:713-738` full form scans the tree with `--no-git` | item 5's target |
| 15 | A planted secret fails the leg | probe: a fake GitHub-token-shaped string in an uncommitted copy of `docs/standing-rules.md`, `gitleaks dir` | "leaks found: 1", exit 1; file restored |
| 16 | The exported-unused leg | `check.sh:812-879`, awk over `internal/`; `grep -c awk scripts/check.sh` prints 5, four of them in this leg and one (`:791`) in the citation leg | after T6 it prints 1 |
| 17 | `go tool` and `go get -tool` | `go help tool`: "additional tools may be defined in the go.mod of the current module. 'go get -tool' can be used to define additional tools"; `go help get`: "The -tool flag instructs go to add a matching tool line to go.mod for each listed package." | probe: `go get -tool golang.org/x/tools/cmd/deadcode` then `go mod tidy`; `go tool deadcode ./...` runs in 5 s and prints 279 lines |
| 18 | #67's missing pattern | `check.sh:781` `citation_patterns`; probe adding `\bR-[0-9]+[a-z]+[0-9]*` | citation count 192 to 220 |
| 19 | #69's retired term | `grep -n boundary scripts/check.sh` | 6 hits (`:31`, `:60`, `:86`, `:253`, `:622`, `:745`); the design names `:60` and `:622` ("boundary gate"); the other four say "every boundary" and are reworded with them |
| 20 | #134's stale header | `check.sh:127-134`, not `:113-122` as BACKLOG #134 says | "flips to fail at the comment-cleanup task"; "flips to fail once the dead surfaces it names are either consumed or deleted" |
| 21 | #64's header line | `check.sh:3-19` names no photo leg; the leg runs at `:757-767` | one header line names it |
| 22 | Capture contexts | `run.ts:1382` `browser.newContext({ colorScheme: state.scheme, viewport: FIXTURE_VIEWPORT, ...options })`, one per cell from `captureRows` (`:6091-6147`) | 190 contexts: 19 states (`rendered_test.go:271-300`) x 5 widths x 2 schemes |
| 23 | 190 captures, 95 files | `run.ts:6087-6089` "The two native widths (390, 1440) each write their own PNG; the other three compose into one contact sheet per state" | 19 state directories x (4 native PNGs + 1 sheet) = 95 files on disk for 190 cells |
| 24 | The settle is a bound | `run.ts:5968` `CAPTURE_SETTLE_MS = 2_000`; `settleCapture` (`:5981`) waits for no disabled button, finished animations, and the primary fill | probe 2: in-page double shots matched in 189 and 190 of 190 cells |
| 25 | Spend reconciliation | `rendered_test.go:708-731` `checkSpendReconciliation` fatals on mismatch; `renderedCaptureRows` (`:302-318`) declares spend per row | probe 1 failed "manifest declares 23 request-link POSTs, the log shows 7"; moving spend to first-width rows passed |
| 26 | Stale counts | `run.ts:5317` "fifteen", `:5536` "fifteen", `:6080-6082` "fifteen ... 150 captures"; `rendered_test.go:39` "150 fresh browser contexts", `:267`; `mintServer` doc (`:652-656`) "sixty links its six signed-in states" | each corrected in T2 |
| 27 | The rendered time bound | `rendered_test.go:36-41` "378s; about 257s uncontended"; `Makefile:120-126` `RENDERED_TIMEOUT := 18m` | the comment is updated with T2's measured figure; the timeout values stay |

## Recorded searches

| Search | Hits on `7db541b` |
| --- | --- |
| `grep -c 'command -v' scripts/check.sh scripts/vale-comments.sh Makefile` | 8, 1, 2 (six are mechanism 12; node at `check.sh:762` and the leg-internal checks stay) |
| `grep -c 'scripts/checks' scripts/check.sh` | 1 (`:366`, a dead machinery row; the adapter `SKILL.md:194` and `web/scripts/gates/ai-tells.ts:3` also name it and are filed, not edited) |
| `grep -c 'CHECK_' scripts/check.sh` | 3 per variable |
| `grep -c 'make check' scripts/check.sh` | 9 |
| `grep -n 'make check' .claude/settings.json CLAUDE.md docs/standing-rules.md` | `settings.json:10` allows `Bash(make check)`; `CLAUDE.md:77`; standing-rules `:55`, `:105`, `:122-127`; all outside the row, and `make check` keeps working |
| `grep -c awk scripts/check.sh` | 5 |
| `grep -n -iE 'node_modules\|rendered-captures\|venv' .gitleaks.toml` | nothing |
| `grep -c newContext web/scripts/rendered/run.ts` | the capture context is one call (`:1382`); the other contexts belong to non-capture rows and stay |

## Tasks

The path is T1, then T2 (an `opus` pair subagent) beside the session's T3, T4, T5, and T6, then the merge and T7.
T2's files (`web/scripts/rendered/**`, `cmd/dubplate-server/rendered_test.go`) are disjoint from the others'.

### T1: Per-leg times and the base run (20 min, `runner`)

Outcome: every form of `check.sh` prints `== <leg> elapsed <seconds>s ==` after each leg and `== total elapsed
<seconds>s ==` at the end, and the full form runs `make web`, `make vet`, `make lint`, `make test`, `make
web-rendered`, and `make tidy-check` as six timed legs in place of one `make check`. Leg names stay fixed from here to
the close, so item 2 can compare them.

Files: `scripts/check.sh`.

Time with bash `SECONDS` (integer seconds is enough at this scale), set once at the top and per leg. A failing leg
still prints its elapsed line before the script exits, so a red run shows where it stopped. Then run `bash
scripts/gate.sh` (the full form) on the T1 commit. That run is the base run of item 2: it is `7db541b` plus timing
lines and nothing else, and the PR's `## Decisions` says so. Save its log's elapsed lines to the PR notes. Copy
`web/rendered-captures/` to `<git dir>/base-captures/` for T2's compare.

Check: `grep -n -iE 'elapsed|SECONDS' scripts/check.sh` prints the new lines; the base run's log holds one elapsed
line per `== <leg> ==` header (count both with `grep -c`) and one total line; the gate exits 0; `ls <git
dir>/base-captures | wc -l` prints 19. Test: the gate.

### T2: Rendered contexts (40 min, `auth-data`, pair, `opus`)

Outcome: the capture matrix opens one browser context per state and scheme (38) and takes each width inside it with
`page.setViewportSize` and `settleCapture`. All 190 cells and every assertion stay. The script prints `capture
contexts: <n>`, and the Go test fails unless `n` is 38. The stale counts in mechanism 26 are corrected.

Files: `web/scripts/rendered/run.ts` (and any file under `web/scripts/rendered/` the change needs),
`cmd/dubplate-server/rendered_test.go`.

Load `standards:ts-conventions` and `standards:go-conventions` first. Test first (the class is `auth-data`): add the
Go-side check that the script's `capture contexts:` line reads 38, with 38 written as a literal, and run it on the
unchanged matrix. Record the failure (it reads 190) in the PR notes. Then change `captureRows`:

- One context per (state, scheme), sized at the state's first width, then for each width in `CAPTURE_WIDTHS` order:
  `setViewportSize`, `settleCapture`, the secondary-control guard, the overflow floor at 320 and 390, the
  touch-target floor, and the screenshot. The state is stood up once per context, after the fixture roster resets.
- In `renderedCaptureRows`, the link spend each state declares moves to its first-width rows, so
  `checkSpendReconciliation` (mechanism 25) still passes with no change to the check itself.
- Correct the stale counts in mechanism 26 and the time comment at `rendered_test.go:36-41`, with this task's
  measured figure.

Mutation proofs: (a) open a context per cell again and the 38 check fails; (b) drop the spend move and
`checkSpendReconciliation` fails. Record both.

Check: two runs of `make web-rendered` on the T2 commit pass, and their capture sets compare as item 3 says (see
"Acceptance"; the comparison is per Geoff's answer to the stop). The PR notes carry the pixel table against `<git
dir>/base-captures/` (file, differing pixel count, box), produced by a throwaway decoder over
`web/scripts/rendered/png.ts`'s `decodePNG` run from the job's tmp directory, never committed. Test: `make
web-rendered`, then the fast lane. Review: `standards:diff-reviewer` at task scope on the pair's merge (the adapter's
per-task review for `auth-data`).

### T3: The scoped build and the shell cleanups (30 min, `runner`)

Outcome: a scoped run with no web change builds the bundle once and runs no `svelte-check` or static web gate; the
full form is unchanged. The six prerequisite blocks are one check. The header and wording items fold in.

Files: `Makefile`, `scripts/check.sh`.

- `Makefile`: split `web` into `web-build` (the Node, npm, and install checks, then one `npm run build`) and `web`
  (which runs `web-build`, then the second build, the compare, `npm run check`, and the static gates, as today). Add
  `web-build` to `.PHONY`. `check:` still names `web`.
- `check.sh` scoped form (`:584-588`): `make web` when a web path changed, `make web-build` otherwise. The reduced form
  keeps its own build.
- `check.sh:218-241`: one loop over the six names that prints every missing one and exits 1 if any is missing.
- `classify_path` (`:378-386`) calls `nearest_go_pkg_dir` (`:422-433`) in place of its own walk; drop the dead
  `scripts/checks/*` row (`:366`).
- Wording: the header names the photo leg (#64); "boundary" (mechanism 19) becomes "full"; the `CHECK_CITATIONS` and
  `CHECK_EXPORTED_UNUSED` text at `:127-134` says the citation leg gates in round `engine` and the exported-unused
  leg stays a warning (#134). The citation leg's `warn` default stays.

Mutation: rename `golangci-lint` in the loop to a name not on the path and the run exits 1 naming it; record it.
Check: `bash scripts/gate.sh --scope-changed` exits 0; `grep -c 'command -v' scripts/check.sh` prints 3 (8 less the
six blocks plus the loop's one); `grep -c 'scripts/checks' scripts/check.sh` prints 0; `grep -c boundary
scripts/check.sh` prints 0. Test: the fast lane.

### T4: `gate.sh` flags and the lane (20 min, `runner`)

Outcome: `gate.sh` accepts `--reduced` and `--fresh` besides its current flags. A run classifies first: it calls
`bash scripts/check.sh <args> --dry-run`, takes the light lane when the scope line reads `scope=changed` with
`rendered=skipped`, or reads `scope=reduced`, and the heavy lane otherwise. It prints `gate.sh: lane=<lane>` before
it execs `cairn-run-gate`.

Files: `scripts/gate.sh`.

`--fresh` goes to `cairn-run-gate` ahead of the gate string (mechanism 6), never to `check.sh`. A dry run keeps the
light lane. Before the light lane goes in, measure one scoped run's whole command under the light cap with mechanism
8's wrapper and record `memory.peak` and `memory.events`. If it OOMs or its `high` events throttle it past twice its
heavy-lane time, the scoped form keeps the heavy lane, and the PR says so (the prompt's rule).

Check: `bash scripts/gate.sh --reduced --dry-run` exits 0 and prints the reduced scope (item 4); `bash scripts/gate.sh
--bogus` exits 2 with the usage line; a `--scope-changed` run whose change set holds one Go file in a package outside
the rendered binary prints `gate.sh: lane=light` and the scope line `rendered=skipped`, and its log shows `make vet`,
`make lint`, `make test`, no `npm run check`, and no static gate (item 6). Run that proof after T4's commit, with an uncommitted
comment edit in `internal/reject` (outside the rendered binary's `go list -deps` set; confirm with `go list -deps -test
-tags rendered ./cmd/dubplate-server | grep -c internal/reject`, which must print 0), and pass `--base <the 40-hex
HEAD>`. Without `--base` the change set runs from the merge base with master, and by T4 it holds `scripts/check.sh` and
`Makefile`, which are gate machinery (`check.sh:458-461`) and promote the run to the full form. With `--base HEAD` the
set is the one uncommitted file (`check.sh:319` runs `git diff --name-only <base>`, which reads the working tree).
Revert the edit after. Test: the fast lane.

### T5: Vale under scope (15 min, `runner`)

Outcome: `vale-comments.sh` takes Go file paths as arguments and reads only those; with none it reads every tracked
Go file as today. `check.sh`'s scoped form passes the changed Go files, and the full form passes none.

Files: `scripts/vale-comments.sh`, `scripts/check.sh`.

A scoped change set with no Go file runs Vale over nothing and says so in one line; it never falls back to a skip of
the leg in the full form. `--check-vendor` keeps working.

Check: after T5's commit, a `bash scripts/gate.sh --scope-changed --base <the 40-hex HEAD>` run with one uncommitted
Go comment edit shows Vale reading one file (the `--base` keeps the branch's machinery out of the change set, as in T4),
then the edit is reverted; the full form reads `git ls-files '*.go' | wc -l` files. Item 7's planted fault is T6's. Test: the fast lane.

### T6: Gitleaks paths, the citation pattern, and `deadcode` (20 min, `runner`)

Outcome: gitleaks skips `web/node_modules`, `web/rendered-captures`, and `tools/venv/.venv`; the citation leg also
matches `R-6c`-style ids; the exported-unused leg runs `go tool deadcode ./...`, pinned by a `tool` line in `go.mod`,
and prints its output with no filter, as a warning.

Files: `.gitleaks.toml`, `scripts/check.sh`, `go.mod`, `go.sum`.

- `.gitleaks.toml`: a path allowlist for the three trees (item 5).
- `check.sh:781`: add `\bR-[0-9]+[a-z]+[0-9]*` to `citation_patterns` (#67). The leg stays `warn` by default.
- `go get -tool golang.org/x/tools/cmd/deadcode`, `go mod tidy`, stage both files before the gate (pass 1 mechanism
  27). Replace `:812-879` with a leg that runs `go tool deadcode ./...`, prints its whole output, and exits nonzero
  only under `CHECK_EXPORTED_UNUSED=fail`. Keep the leg name.
- Plants, never committed: a fake token in a tracked file and a Vale comment fault in a Go file the branch doesn't
  change. Run `bash scripts/gate.sh --merge` once with both present; it must exit nonzero with the gitleaks leg and
  the Vale leg both red. Record both lines in the PR notes, then remove both and confirm `git status` shows neither.
  If the plant write is refused, that is standing answer S4.

Check: `grep -n -iE 'node_modules|rendered-captures|venv' .gitleaks.toml` prints three lines; `grep -n '^tool
golang.org/x/tools/cmd/deadcode' go.mod` prints one line; `grep -c awk scripts/check.sh` prints 1; the citation count
the leg prints rises from 192 (mechanism 18). Test: the fast lane, then the plant run.

### T7: The close (30 min, `runner`)

Outcome: the close head's full timed run, the `--merge` run, the BACKLOG closes, the review, and a ready PR.

Files: `BACKLOG.md`, this plan and its prompt (`git mv` to `docs/archive/plans/`).

1. Merge T2 if not merged; run the fast lane.
2. Run `bash scripts/gate.sh --merge` on the close head. Its elapsed lines are item 2's second run: the total is
   lower than T1's, and every leg T1 named is named again except "exported identifiers", which T6 replaced (keep its
   name, and the comparison is by name). Copy both runs' per-leg tables into the PR body.
3. Close #34, #64's header clause (leave the rest of #64 open), #67, #69, #104, #105's gate clauses (the
   `.golangci.yml` clause stays open, reworded), and #134. File: the `.golangci.yml` comment, the standing-rules
   wording that names `make check`, the adapter's `scripts/checks/` line, and `ai-tells.ts:3`.
4. Whole-branch `standards:diff-reviewer`, plus `standards:web-auth-security-reviewer` on the `rendered_test.go`
   diff (the class is `auth-data`). One fold on `fix`, then one fresh read.
5. PR body: `Round: cleanup`, `Pass: cleanup-gate`, `Review:`, `Calls:`, each acceptance item's evidence, `##
   Decisions`, `## Planning misses`, the score, and the attribution line. Mark ready once CI is green on the head.

## Standing answers

| # | Stop | Answer |
| --- | --- | --- |
| S1 | The rendered leg's elapsed line is over item 3's bound after T2 | Record the per-cell split from one instrumented run in the PR notes. Cut nothing outside how contexts open. The PR carries a call and doesn't land. |
| S2 | A capture still differs from the base outside item 3's allowance after two cause fixes | Stop; the question names the files and the boxes. Never accept it as the new baseline (standing answer 18). |
| S3 | The scoped form OOMs or throttles under the light cap | Keep the heavy lane for it, say so in the PR, and go on (T4). |
| S4 | A plant write or the red `--merge` is refused | Record the refusal, run the two plants as separate uncommitted leg runs (`gitleaks dir` and `vale-comments.sh <file>`), and list it under `## Decisions`. |
| S5 | CI red outside the pass's files | Fix in the branch, list under `## Decisions` (standing answer 19). |

## Decisions for the executor

- D1: `cmd/dubplate-server/**` maps to `auth-data`, and the design classed the pass `runner`. The map wins for T2;
  the PR says so. The test-first record and the security read follow from it.
- D2: Item 2's "full run on the pass's base" is T1's run, which is the base plus timing lines only; no earlier commit
  prints per-leg times.
- D3: The design's "190 captures" are 190 cells written as 95 files (mechanism 23). Every compare is over the 95.

## Path

| Step | Minutes |
| --- | --- |
| T1 (with the base full run) | 20 |
| T3 + T4 + T5 + T6 in the session, beside T2 (40) | 85 |
| Merge T2, fast lane, per-task review | 10 |
| T7 | 35 |
| Total | 150 `[estimate]` |

## Acceptance

Copied verbatim from the design, lines 208-234, plus the two every-pass items. Item 3 is tagged `[design-gap]`
until Geoff answers the stop.

| # | Item | Task | Evidence |
| --- | --- | --- | --- |
| 1 | A full gate run prints one elapsed-time line per leg and a total. Today: `grep -n -iE 'elapsed\|SECONDS\|date \+%s' scripts/check.sh` prints nothing. | T1 | the grep's lines; a log's elapsed count against its header count |
| 2 | The PR body carries the per-leg times of a full run on the pass's base and of a full run on the close head, both on pavilion. The total is lower, and every leg named in the first run is named in the second, apart from the awk leg that item 8 replaces. Today: no per-leg figure exists; the review measured 10,224 s over 20 runs in `admin-api-1`. | T1, T7 | both tables in the body |
| 3 | `[design-gap]` The rendered leg writes 190 captures that match the base commit's 190 with no pixel tolerance, opens 38 browser contexts where it opened 190, and its elapsed line on pavilion reads under 130 s. Today: about 257 s by the test's own comment, and `web/scripts/rendered/run.ts:6082` still says "150 captures". | T2 | the 38 check; the pixel table; the elapsed line |
| 4 | `bash scripts/gate.sh --reduced --dry-run` exits 0 and prints the reduced scope. Today: exit 2 and the usage line. | T4 | the command's output and exit |
| 5 | `.gitleaks.toml` excludes `web/node_modules`, `web/rendered-captures`, and `tools/venv/.venv`, and a planted fake secret in a tracked file still fails the leg (run once, recorded, removed). Today: `grep -n -iE 'node_modules\|rendered-captures\|venv' .gitleaks.toml` prints nothing. | T6 | the grep; the plant run's gitleaks line |
| 6 | A scoped run whose change set holds one Go file runs vet, golangci-lint, and the Go tests, runs no `svelte-check` and no static web gate, and reports the light lane. Today: it builds the web app twice and queues on the heavy lane (BACKLOG #34). | T3, T4 | the run's log and `gate.sh: lane=light` |
| 7 | A scoped run's Vale leg reads only changed Go files, and the full form still reads every Go file: a planted comment fault in a file the branch doesn't change fails `gate.sh --merge` (run once, recorded, removed). Today: `vale-comments.sh` reads every Go file on every form. | T5, T6 | the scoped run's Vale line; the plant run's Vale line |
| 8 | The exported-unused leg runs `go tool deadcode ./...`, pinned by a `tool` line in `go.mod`, and the awk block is gone. The leg stays a warning, as it is today, and prints what `deadcode` reports with no filter. Today: `grep -c awk scripts/check.sh` prints 5. | T6 | the `go.mod` grep; `grep -c awk` prints 1 (the citation leg) |
| 9 | The check from pass 1 is green on this PR with no change to which legs it runs. | T7 | the CI run link |
| 10 | The PR's check is green on its head. | T7 | `gh pr checks <n>` |
| 11 | `bash scripts/gate.sh --merge` exits 0 on the close head. | T7 | the `gate exit: 0` line |

## Review record

Pending: one `standards:plan-reviewer` read after Geoff's answer.
