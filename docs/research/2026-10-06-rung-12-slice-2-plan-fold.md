# Rung 12 slice 2a plan fold (revision 1 to 1.1)

Target: `docs/superpowers/plans/2026-10-06-rung-12-slice-2-oracle-harness.md`, revision 1 at `f62a151`, folded to
revision 1.1 (uncommitted at writing; the conductor commits). Reviews folded:

- contract lens, `2026-10-06-rung-12-slice-2-plan-review-contract.md` (C1 to C16; over-ceremony CO1 to CO3 below,
  the lens's O1 to O3);
- mechanics lens, `2026-10-06-rung-12-slice-2-plan-review-mechanics.md` (M1 to M13; MO1 below, the lens's O1);
- risk lens, `2026-10-06-rung-12-slice-2-plan-review-risk.md` (R1 to R13; RO1 and RO2, the lens's O1 and O2; F1).

Each finding was verified before it was acted on. Checked directly for this fold: `scripts/check.sh:183`,
`:270-275`, and `:364-367` with `Makefile:134` (M1); `internal/engine/tagio/write.go:122-128` and `:239-245` (C5,
R8: `planBlocks` keeps an existing vendor and stamps `freshVendor` only when no `VORBIS_COMMENT` exists, so R8's
reading is the precise one); `$GOROOT/src/os/exec/exec.go:713-718` and `go doc context.WithoutCancel` (R6); the
parallel-build approach `:255` and the re-plan ruling 5 `:48-54` (the worktree question); the 6d plan revision 2.4
`:426-429`, `:944-950`, and `:1694-1697` (R9's timing; revision 1 cited revision 2.3's `:1526-1539`);
`docs/HISTORY.md:191` (CO3's 3.7M actual); and the proxy and GitHub API records for `go-openapi/jsonpointer` and
`go-openapi/swag` (M11). The live-probe measurements (M3, M4, M5, M6, M7, M2's field lists) are taken as the
mechanics lens recorded them.

## The convergent blocker: C1, R1, M2

All three lenses found that revision 1's gate proof graded Navidrome against gonic under the committed ledger and
demanded zero unledgered findings. The mechanics probe measured the two oracles disagreeing well past the folder
view (21 Navidrome-only and 4 gonic-only key paths in `getAlbum`, index-name case, `ignoredArticles`, `roles`,
the envelope `version`). No class held "present on one oracle only", so the proof could pass only by misfiling
rows, and a misfiled row naming Navidrome and a path would later match a dubplate-against-Navidrome grade and hide
a real divergence.

**Re-derived from the spec.** Slice 2 delivers the fixture library, the oracles, the contract tier, the session,
the differ, and the ledger (spec `:732-736`). The oracle tier compares dubplate with an oracle (`:274-276`), and
slice 3 is the first slice with a dubplate side. What slice 3 consumes is a differ with exact findings, a contract
tier with no silent pass, a ledger whose rows match only their own pair, and a session over library bytes a later
test can rebuild. Revision 1.1 proves each directly (plan, "What slice 3 consumes"):

- the differ on a same-source perturbation table (C1's proposal): zero-finding baseline, exactly one finding per
  row;
- the contract tier on a hand-authored document covering every kin silent-nil path, plus the real bodies' exact
  violation lists (M3);
- the ledger on fixture ledgers;
- the real pipeline by grading the seed session's oracle pair under a test-only ledger in `grade/testdata/`,
  pinned as a committed golden report (C1), with named findings asserted and every test row's match count at
  least one.

The committed ledger takes only rows with `dubplate` as a source (R1's pair scoping, enforced by a gate test), and
in 2a holds the spec's two declared classes as `owed` rows. Plan Reading R12 records this.

**Refused within the fold:** M2's and R1's fifth class, `oracle-split`. Reason: once the oracle-pair differences
live in the golden as unledgered evidence, no ledger row needs the class, and adding it would put a class into the
committed taxonomy that only an oracle-against-oracle grade can use.

## Dispositions

One line per disposition; duplicates listed together.

| IDs | Disposition | Where |
|---|---|---|
| C1, R1, M2 | Folded (convergent; above). R1's match counts and one-field patterns folded; M2's id-pairing scope rule folded; the `oracle-split` class refused (above). | Header "What slice 3 consumes", R12, Task 6 (pair rule, classes, committed rows, match counts), Task 8 (perturbation table, seed grade, golden) |
| C2, R2 | Folded. Ids pair through the session's resolution map, then manifest identity; never by position. Order is a separate assertion on a named table. | Task 4 (resolution map), Task 5, Task 8 |
| C3, M9, R12 | Folded. The checker resolves route, status, content, and schema itself, sets `MultiError` and `IncludeResponseStatus`, and reports five outcomes; the silent-nil paths are quoted. | Quotes table, Task 7 |
| C4, M4 | Folded. Two-sibling-scalar proof body; violations filtered to the body's `status` branch and deduplicated. | Task 7 |
| C5, R8 | Folded through R8's mechanism: seeds must carry a `VORBIS_COMMENT` block, and the round trip asserts the seed's vendor. C5's library digest folded (recorded per session; the grader refuses mixed digests). C5's "verbs refuse a non-`(devel)` binary" refused: R8's rule removes the binary dependence at its root. | Existing-pieces table, Task 1, Task 4, Task 8 |
| C6 | Folded. Loader refuses an `active` row naming an absent session; a gate test grades every session a committed `active` row names. | Task 6 |
| C7 | Folded. `Finding` type, kind enum, path grammar, and per-class suppression defined in Task 6 and consumed unchanged. | Task 6 |
| C8 | Folded. The loader cross-checks `serverVersion` against every body's envelope. | Task 4 |
| R5 | Folded. Empty `serverVersion` refuses; an `active` row's empty version refuses; the recorder requires the pin in `Digest` or `RepoDigests`, quoted from `podman-image-inspect(1)` at Task 0. | Task 0, Task 5, Task 6 |
| C9 | Folded. Fixture states named for the digest and version mutations. | Task 5 |
| C10 | Folded. The no-link matcher takes a listing and gets a canned-violation proof. | Task 7 |
| C11, R11 | Folded. The vacuous grep is replaced by "every committed session loads through Task 4's loader", which refuses auth parameters. | Task 4, Task 9 acceptance |
| C12, M7 | Folded. Navidrome's first admin through `ND_DEVAUTOCREATEADMINPASSWORD`; gonic keeps `admin`/`admin` on a per-run loopback container; the credential rule reworded; log tails redacted. M7 closes R4's `p=enc:` branch (both oracles take tokens). | Quotes table, Global constraints, R4, Task 3, "Not verified" |
| C13, R9 | Folded. Size cap as a test; decodability evidenced by `duration > 0` in the seed session; image seeds generated, command recorded. | Task 1, Task 5 |
| M3, C14 | Folded. Real bodies' exact violation lists asserted (Navidrome `/album/userRating`, gonic `/song/0/replayGain`); the repaired body asserts valid; the package's own proofs use a hand-authored document. | Task 7 |
| M10, R3 | Folded through R3's approach: conformance tests use their own `testdata/` document and bodies, and `git grep '\.\./' -- internal/opensubsonic` prints nothing. M10's move of the pinned file under the protocol subtree refused: R3's in-package document keeps the extractable module free of the 453 KB Apache-2.0 file and its license copy, and the harness side still exercises the real file. | R3, Global constraints, Task 7, Task 9 acceptance |
| C15 | Folded through the golden: the seed grade runs twice and compares bytes. | Task 8 |
| C16, M13, R4, F1 | Folded per the conductor's resolution: 6d merges first; 2a builds alongside in its worktree and holds its branch until lane 1 closes and a quiet `go.mod` window exists, because 6d is on the launch path. Task 0's sweep runs in the main tree only with no lane 1 executor; otherwise 6d's sweep is recorded as this pass's. | R9, Global constraints, Lane launch, Task 0, Task 9 |
| CO1, R10 (a), RO2 (part) | Folded. The Markdown transcription proof moves to 2b Task 11, its first caller. | Task 6, 2b table |
| CO2, R10 (b), RO2 (part) | Folded. The search set-plus-top-N rule moves to slice 4, its first consumer. | Task 8, 2b sketch |
| R10 (c) | Folded. The envelope exposes only `status`, `serverVersion`, and the error code and message, which 2a callers read. | R4, Task 2 |
| CO3 | Folded as a re-derivation from actual spend: 8.5M, flag 6.8M. The lens's ~7M is not taken, because the fold added scope (identity pairing, branch filter, golden, library digest) and Task 8 is novel on Opus. | Header "Ceiling", Ruling 1 |
| M1, MO1 | Folded per the conductor's resolution: every task gate runs on the heavy lane; the header quotes `check.sh:183`, `:270-275`, `:364-367`, and `Makefile:134`; no new gate lane. The decoupling of the rendered scope is filed for after 6d. | Header "Per-task gate", "Execution mode", "Not verified" |
| M5 | Folded. Readiness requires `count` equal to the manifest's track count; new states row and mutation. | Task 3 |
| M6, R13 | Folded. `:ro,z` mount and gonic's four `--tmpfs` paths, asserted in the argv test. R13's "not verified" is closed by M6's measurement. | Quotes table, Task 3 |
| M8 | Folded. gonic's two multi-value variables set to `multi` and asserted. | Quotes table, Task 3 |
| M11 | Folded per the conductor's resolution, not a ruling: newest versions with evidence (Reading R11). `jsonpointer` v0 to v1 read as non-major under Go module semantics, with a Task 0 release-notes check that turns a stated API break into a held major. | Prior-art evidence, R11, Task 0, Task 7 |
| M12 | Folded. `go vet -tags oracle` once at the close, and in R1's queued `CLAUDE.md` wording. | R1, Task 9 |
| R6 | Folded. Teardown under `context.WithoutCancel` with its own timeout, `podman rm -f -v`, a per-run label; the `exec` behavior quoted. | Quotes table, Task 3 |
| R7 | Folded. Publish with no fixed host port; read it with `podman port`. | Task 3, Global constraints |
| RO1 | Folded. The vacuity read is scoped to the `grade` package, the one package specified only by outcome. | Task 9 |

## The worktree question

Verified against the recorded reason. `CLAUDE.md:103-104` states the rule without a reason; the parallel-build
approach's checklist item 5 (`:255`) gives it: the venv is gitignored, so a worktree lacks it, and `make derive`
sits outside `make check`, so "the lane's gate goes green while the only leg exercising the beets oracle never ran
and never complained." This harness uses no venv, no beets, no `corpus/`, and no `make derive`, and the re-plan's
ruling 5 places the slice in its own worktree unless the plan finds it needs the main tree (`:48-54`). The reason
does not apply, so the lane stays in its worktree. The one residue, the `//go:build oracle` test in `cmd/harness`'s
package `main` (M12), is answered by `go vet -tags oracle` at the close, which needs no venv. No ruling.

## Owner forks

None new. M11's fork and the risk lens's F1 were settled by the conductor's resolutions above, recorded in the plan
as Readings R11 and R9. Ruling 1 stands, with its 2a ceiling moved from 9.5M to 8.5M.

## Errata owed to ratified documents (not edited here)

- **Spec `:312-314`.** "`go list -deps` over the layer's packages shows no `github.com/glw907/dubplate/` path"
  fails on the layer's own packages when read literally. Owed wording: no module-internal path outside the layer's
  subtree. (Carried from revision 1's R6.)
- **Parallel-build approach `:255`.** The policy sentence "corpus and harness work never moves to a worktree"
  should name its scope: the matcher corpus and the beets oracle, whose gate leg needs the venv. `CLAUDE.md`'s
  matching edit is queued in Task 9.
- **Re-plan ruling 5 (`:48-54`).** "Its `kin-openapi` addition rebases onto the removal pass's dependency sweep"
  is overtaken by sequencing: under R9 it rebases onto master after 6d's close, past 6d's own sweep.

One observation for slice 1's record (not an erratum): the real `getAlbum` bodies of both pinned oracles violate
the pinned OpenAPI schema (M3), so the spec's contract tier will carry oracle-side `contract` rows from slice 3 on.

## Measures

- **New-mechanism findings folded (with the source that runs it and the defect it answers):** identity pairing
  (C2, R2; the recorder's resolution map; the probe's index-order and case differences); the violation branch
  filter and dedup (M4; kin's `SchemaError` leaves; the probe's leaf table); the checker's pre-resolution (C3, M9,
  R12; kin's `Responses.Status`, `Default`, `Content.Get`; the quoted nil paths); the test-only ledger and golden
  (C1; the Go `-update` golden convention; the probe's field lists); pair scoping and match counts (R1; the grader;
  the misfiled-row defect); the library digest (C5; Task 1's builder; the measured version stamp); the
  readiness count (M5; `getScanStatus`; the probe's zero-track scan); log redaction (M7; the probe's password log
  line); teardown under `WithoutCancel` (R6; `exec.Cmd.Start`'s quoted check); `podman port` (R7; podman); the
  serverVersion cross-check (C8; the envelope the client parses). Eleven folded.
- **New-mechanism findings refused:** the `oracle-split` class (M2, R1; no row needs it once the golden holds the
  pair's differences); the binary-version refusal (C5; R8's seed rule removes the cause); the pinned-file move
  (M10; R3's in-package document answers the same defect). Three refused.
- **Dispositions:** 49 finding IDs in 31 dispositions; all folded, with the three sub-proposals above refused; no
  owner fork.
- **Plan lines:** 625 before, 826 after.
- **Ceiling:** 2a from 9.5M (flag 7.6M) to 8.5M (flag 6.8M), re-derived from actual spend; 2b unchanged at 4.0M
  (flag 3.2M). Tasks unchanged at ten (Task 0, eight build tasks, the close); segments unchanged (A: 1 to 3; B: 4
  to 6; C: 7 and 8). No further split.

## Second fold (revision 1.1 to 1.2)

Source: `docs/superpowers/research/2026-10-06-rung-12-slice-2-plan-fold-verification.md` (V1 to V3 major, V4 to V12
minor), verdict "needs second fold". Scope was the verification findings only. Checked directly for this fold:
kin-openapi v0.149.0 in the module cache (`openapi3/schema.go:2284`, `:2294`, `:2304`, `:2326`;
`openapi3/errors.go:57-58`; `openapi3filter/validate_response.go:23-36`); the 6d plan at revision 2.5 (`:432-435`,
`:951-957`, `:1694-1708`); and the podman 5.8.7 man pages (`podman-run(1)` `-p`, `--label`, `--tmpfs`;
`podman-port(1)`; `podman-rm(1)` `--volumes`; `podman-ps(1)` `--filter label=`). No live probe was needed.

| ID | Disposition | Where |
|---|---|---|
| V1 | Folded, keeping the check. The library digest is recorded per source, a source joins a session through Task 4's writer (refusing a duplicate source or a response to an absent request), Task 5 adds each oracle that way, and slice 3 joins dubplate to a copy of the seed session the same way. The mutation is killed by a hand-built session whose two sources carry different digests. | Header, Task 4, Task 5, Task 8 |
| V2 | Folded. A `contract` finding carries one source; a `contract` row names the source it excuses and matches only that source's findings. The pair proof gains the oracle-row-against-dubplate case, with its mutation. | Task 6, Task 8 |
| V3 | Folded. The perturbation table drives the oracle-tier differ alone. `TestGradeIdenticalBodiesReportOnlyContract` runs the full grader over two sources holding Navidrome's real `getAlbum` body and asserts zero oracle-tier findings and exactly Task 7's recorded violations, once per source. It kills the skip-contract mutation. | Header, Task 8 |
| V4 | Fixed. kin's `oneOf` grouping quoted; the filter descends the `oneOf` error's `Origin`. | Quotes table, Task 7 |
| V5 | Fixed. Nodes keep their XML form; a form mismatch is one `type` finding. | Task 6, Task 8 |
| V6 | Fixed. The committed manifest is a grader input; a song with no track number pairs by album and title. | Task 8 |
| V7 | Fixed. An `active` row carries one version per non-dubplate source. | Task 6 |
| V8 | Fixed. Both sentences scoped to JSON `subsonic-response` operations, with the no-`status` fallback kept. | Quotes table, Task 7 |
| V9 | Fixed. Seven Sonnet tasks at 0.5M: 3.5 + 0.8 + 1.5 + 1.0 + 2.2 = 9.0M, flag 7.2M. Ruling 1 updated. | Ceiling, Ruling 1 |
| V10 | Fixed. 6d re-cited at revision 2.5. | Sources, R9, Global constraints |
| V11 | Fixed. The podman behaviors quoted from the local man pages; Task 0 re-quotes them. | Quotes table |
| V12 | Fixed. The checker tests `HEAD` and the four 3xx statuses before routing, as kin does. | Task 7 |

**Refused:** none. **Owed:** none. **New mechanism:** none beyond the verification's own narrow folds (the join
path through the existing writer, and one source on a `contract` finding and row). The cost of the full-report
golden's churn through slice 8 (the verification's optional note) is accepted as stated there.

**Measures.** Twelve findings, all folded. Ceiling: 2a from 8.5M (flag 6.8M) to 9.0M (flag 7.2M), an arithmetic
correction, not a scope change; 2b unchanged at 4.0M. Tasks unchanged at ten (Task 0, eight build tasks, the
close); segments unchanged.
