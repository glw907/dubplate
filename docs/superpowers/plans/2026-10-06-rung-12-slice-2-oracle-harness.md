# Rung 12, Slice 2 (the Oracle Harness) Implementation Plan

Status: **draft revision 1, 2026-10-06; awaits `spec-plan-review`, then Geoff's approval gate.** One ruling
is open (Ruling 1). Every other decision is a reading below, taken on its merits and reported.

**Spec:** the 0.8 streamer spec (`docs/superpowers/specs/2026-09-07-streamer-2-0-design.md`, revision 3, "the
spec"; it still says 2.0), slice 2 at `:732-736`, under the re-plan spec
(`docs/superpowers/specs/2026-10-05-rungs-12-16-replan-design.md`, revision 2.1, "the re-plan"), ruling 5 at
`:48-54`.

**What this pass is.** It builds the evidence machinery every later rung 12 slice is graded by. The protocol
layer gets its first two packages, a client and a conformance checker, under a gate-enforced import boundary. The
harness side gets a byte-deterministic fixture library, the pinned Navidrome and gonic containers as
workstation-only behavior oracles, a recorded session format, a divergence ledger that cannot rot, and a grader
that runs the contract tier and the oracle differ over a session. The proof that all of it works is the
oracle-against-oracle grade of a seed session: Navidrome and gonic disagree on the folder view by construction,
and the grader must find exactly the disagreements the ledger declares.

**What this pass is not.** It touches no production surface. It builds no wire types, no handler, no endpoint,
and no config key. It runs no container in the gate, and the gate opens no network connection. Its delta to the
set of values a default install's operator supplies by hand is zero (re-plan `:137-138`).

**Pass class:** `tool`. Everything here is Go that ships nowhere: a workstation harness and two `internal/`
packages that `cmd/dubplate-server` never links in this pass. No task touches auth, sessions, the store, or
`web/`, so no task takes `auth-data`. The client computes a Subsonic token, but it does so as a test client
against throwaway oracle credentials, and nothing server-side verifies one. The class gives an Opus
`diff-reviewer` per task, `go-conventions` as the mandate, and a `go-architecture-reader` per touched package at
the close. This plan raises the mandate in one respect: every task names its mutations, and the implementer
reports each one failing, because a harness whose tests pass vacuously poisons six slices of evidence.

**The split (Ruling 1).** The slice needs twelve build tasks. Four of them consume deliverables only slice 1 can
produce, so the plan cuts at that seam. **Pass 2a** (this plan's Tasks 0 to 9) builds everything slice 1 does not
gate and merges to master. **Pass 2b** (Tasks 10 to 13, sketched under "Pass 2b") consumes slice 1's record and
gets its own plan revision when slice 1 lands. Slice 1 needs nothing from 2a: it runs its own probe oracles on a
scratch fixture and cites each behavior to a version now, with a "test owed" column in its ledger (slice 1 plan,
settled items 3 and 4). 2a's tests are where many of those owed tests land.

**Ceiling (2a):** 9.5M execution tokens, flag 7.6M (80%). Derivation: the removal pass priced 8.0M for seven
tasks and 6d 11.5M for ten. 2a has eight build tasks, two of them novel (Tasks 3 and 8), and its close runs seven
architecture reads, one simplifier round, and one vacuity read. Planning and plan review are recorded outside it.
**Ceiling (2b):** 4.0M, flag 3.2M, re-derived at 2b's Task 0 from 2a's actual rate.
**Lane sub-ceiling:** lane 3's sub-ceiling is 2a's ceiling, then 2b's. The conductor reconciles both against the
phase's combined ceiling at each checkpoint.
**Checkpoint interval:** each segment boundary (three tasks apart at most) and the close. STATUS's lane 3 row is
written at each.
**Segments (2a):** Task 0 (conductor) first. A: Tasks 1, 2, 3. B: Tasks 4, 5, 6. C: Tasks 7 and 8. Then the
close's reviews, the merge ritual, and Task 9. Every boundary lands on a green full gate.
**Tasks (2a):** ten. Task 0, eight build tasks, and the close.

**Executor:** `dubplate-implementer`, `model: sonnet`, effort `high`, for Tasks 1 to 7. Task 8 runs on
`model: opus` from the start, because the grader's id correspondence is novel correctness-critical logic this
plan specifies only by outcome. A second `fix` verdict on Task 5 is the one upshift trigger the conductor weighs
first, since its failure mode is a reference that resolves to the wrong entity and still records.
**Per-task gate:** `CAIRN_GATE_LANE=light cairn-run-gate 'bash /var/home/glw907/Projects/dubplate/.claude/worktrees/lane-3-slice-2/scripts/check.sh --scope-changed'`,
re-issued on exit 75 until it prints `gate exit:`. The light lane holds because no task touches `web/`,
`internal/serve/`, `internal/auth/`, `internal/identity/`, or the serve shell's files, so `--scope-changed` skips
the rendered leg and launches no browser (`scripts/check.sh:26-40`). Every task touches `internal/harness/`, so
the matcher's conform suite runs on every task gate; that cost is accepted. The comment-only fix round takes
`check.sh --reduced`. Segment boundaries and the pass end take the full form on the heavy lane (`check.sh` with
no flag), and the merge ritual takes `check.sh --merge`.
**Execution mode:** `pass-execute` by name, one invocation per segment, sequential, with `repo` set to the
worktree path, `implementer: "dubplate-implementer"`, `reviewer: "diff-reviewer"`,
`reviewerModel: "claude-opus-5-5"`, `maxFix: 1`, `stopOnEscalate: true`, `passClass: "tool"`,
`gateLane: "light"`, `gate` set to the per-task string, and `reducedGate` set to the `--reduced` form.
**Shared heavy lock:** the boundary gates queue on the machine's heavy lock beside lane 1 and the cairn-cms
session. The conductor follows `~/.claude/docs/pass-gate-economy.md` ("Sessions sharing the heavy lock
coordinate") and never re-runs a full gate on an unchanged commit.

## Sources

Code and spec lines are cited at `869915e` on master. Every commit since `0cb33ae` touched only `docs/`, so the
code pointers hold at both. Library behavior is quoted at the version this plan adopts.

- The slice 1 plan (`docs/superpowers/plans/2026-10-06-rung-12-slice-1-study.md`, revision 1.1 at `4db7c23`):
  settled items 3 to 6 and 9 (the probe fixture, "a test owed", the first-pin-wins rule, the Markdown ledger's
  fixed columns, the emulator install), Task 0's pin handoff, Task 1 (the XML rule, the quirks, the PR re-read), and
  Task 7 (the oracle assignment, the ledger columns, and the folder-browsing owner fork).
- The spec: "Evidence" (`:258-289`), "The protocol layer is a separable module" (`:298-328`), "License"
  (`:341-347`), "API surface" (`:640-673`), slices 1 and 2 (`:724-736`), the slice note (`:763-765`), "Prior art"
  (`:856-891`), and "Open items" (`:917-955`).
- The re-plan: ruling 5 (`:48-54`), the rung 12 ladder text (`:113-141`), the roster walk (`:140-141`), and
  "Where the carried items land" (`:410-447`; the API base URL row is `:427`).
- `docs/superpowers/research/2026-09-25-streamer-prior-art-survey.md` ("the survey"): the brief (`:9-25`), finding
  1 (`:54-59`), the OpenAPI rows (`:101-103`), section 3, "Conformance suites and test drivers" (`:155-175`), and
  "Not verified" (`:262-276`).
- `docs/superpowers/research/2026-10-05-client-roster-audit.md` ("the roster audit"), "The roster" (`:8-22`).
- `docs/superpowers/specs/2026-09-03-parallel-build-approach.md`: the lane-launch checklist (`:249-258`).
- The repo `CLAUDE.md`: "Lanes" (`:99-110`), "Planning-miss prevention" (`:127-134`), "Findings are ranked"
  (`:136-144`), "Current dependencies" (`:55-68`), and the build-for-0.8 paragraph.
- The 6d plan, committed revision 2.3 at `4b2cddf` (`git show HEAD:docs/superpowers/plans/2026-09-26-rung-6d-identity-serve-store.md`):
  its header shape and "Contested surfaces" (`:1526-1539`), which names this slice's `go.mod` addition as the one
  expected concurrent edit and sets when it may land.
- The navidrome-removal plan (`docs/superpowers/plans/2026-10-05-navidrome-removal.md`) for the gate and
  execution header lines.

### What `internal/harness` already holds, and what this pass reuses

| Existing piece | At `0cb33ae` | Disposition |
|---|---|---|
| `internal/harness/cassette` | `cassette.go:1-24`: verbatim response files, never re-encoded, because `encoding/json` compacts and HTML-escapes a `json.RawMessage`; canonical path and query with sorted pairs, never decoded and re-encoded | The lesson is reused and the package is not. Its `Case` schema is the matcher's (beets items, an expected record; `cassette.go:41-57`), and `canonicalizePathQuery` (`:260`) is unexported and keeps every parameter, while a Subsonic session must drop the auth parameters. Task 4 states both rules in its own format. |
| `internal/harness/stub` | `stub.go:1-6`: replays a case over loopback and records misses | Not needed in 2a. The grader reads recorded responses from files. A later slice that wants a fake oracle over HTTP may adopt its shape. |
| `internal/harness/conform/arch_test.go` | `:16-65` and `:138-172`: a tree-walk invariant (the engine never imports the harness) with a proof that the check fires on a violating fixture | Task 2's boundary leg follows this pattern, including the fire-on-fixture proof. Task 7's no-link assertion joins this file, beside `TestEngineNeverImportsHarness`. |
| `internal/harness/oracle` and `make derive` | `oracle.go:1-10`, `Makefile:136-146`: the beets oracle runs outside the gate under the `oracle` build tag | The pattern is declined for this pass. The live container path ships as `cmd/harness` verbs, which the gate compiles, vets, and lints like `capture` (`cmd/harness/main.go:7-11`). A build tag would hide that code from `golangci-lint` and from `make derive`'s beets-only scope. |
| `internal/harness/analysis/ledger.go` | `:10-25`: strict JSON reads, unknown keys rejected | The rule is reused through `encoding/json/v2`'s `json.RejectUnknownMembers(true)`, as `internal/auth/roster.go:340` already does. |
| `cmd/harness` | `main.go:53-66`: the verb registration block | Tasks 3, 5, and 8 each add one verb. No other lane edits this block today. |
| `internal/engine/tagio` | `write.go:34-39`: `Write(path, ts, pictures, mode)` rewrites a FLAC's tags and pictures | Task 1 calls it to tag the fixture library. The import runs harness to engine, the allowed direction. |
| `internal/harness/oracle/testdata/template.flac` | 42 bytes, a header with no audio frames | Declined as a fixture seed. Later slices stream the fixture library, so its seeds carry real audio. |

### Library and stdlib behavior this plan relies on, quoted

| Source | Quote | Relied on by |
|---|---|---|
| `openapi3filter.ValidateResponse`, kin-openapi v0.149.0 | "ValidateResponse is used to validate the given input according to previous loaded OpenAPIv3 spec. If the input does not match the OpenAPIv3 spec, a non-nil error will be returned." | Task 7 |
| `openapi3filter.Options`, same | "MultiError bool" (no doc line); "Set IncludeResponseStatus so ValidateResponse fails on response status not defined in OpenAPI spec" | Task 7: many-bad reports every violation, never the first only |
| `openapi3filter/validate_response.go:96-112`, same | When the response has no content entry, "An operation does not contains a validation schema for responses with this status code" and it returns nil. A content type with no entry returns a `ResponseError` naming the MIME type. | Task 7: "no schema" is a reported outcome, never a silent pass |
| `routers/legacy.NewRouter`, same | "If the given OpenAPIv3 document has servers, router will use them." | Task 7 |
| `openapi3/server.go:145-190`, same | A `{var}` segment in a server URL consumes input up to the next `/`. | Task 7: the pinned file's only server is `{url}`, so a full loopback URL leaves `//127.0.0.1:…/rest/…` unmatched. The checker routes by path. |
| The pinned bundled OpenAPI file (`cb54c03c…ed46bfb6`, 453,720 bytes, fetched 2026-10-06) | `openapi` 3.0.0, `info.version` 1.16.1, 87 paths, all of the form `/rest/<endpoint>`, none with `.view`; each `200` response has only `application/json` content. | Tasks 4 and 7: the canonical endpoint drops a `.view` suffix before routing |
| OpenSubsonic `content/en/docs/api-reference.md:20-25,55-59` at `bed1688` (Apache-2.0) | "`t` … The authentication token computed as md5(password + salt)"; "Use a salt length of at least six characters"; "Treat the strings as UTF-8 encoded"; the worked vector: password `sesame`, salt `c19b2d`, token `26719a1196d2a940705a59634eb18eab`; `f` takes "xml", "json", and "jsonp" | Task 2 |
| `encoding/json/v2.Unmarshal` | "a nil interface value of an empty interface type is initialized with a zero Go bool, string, float64, map[string]any, or []any" | Task 8: decoding into `any` turns every number into a `float64`, so the grader reads numbers through `encoding/json/jsontext` tokens and keeps the literal |
| `encoding/xml.Decoder.Token` | "Token expands self-closing elements such as <br> into separate start and end elements"; tokens are "properly nested and matched" | Task 8: `<a/>` and `<a></a>` compare equal |
| `go help list` | "-deps … iterate over not just the named packages but also all their dependencies"; "-test … report not only the named packages but also their test binaries"; a test variant's path carries a suffix | Task 2: a listed path is cut at its first space before the prefix check |
| Navidrome docs, `navidrome/website` `content/en/docs/usage/configuration/options.md` (Apache-2.0) | `ND_ENABLEINSIGHTSCOLLECTOR` default `true` (`:145`), `ND_AGENTS` default `"deezer,lastfm,listenbrainz"` (`:157`), `ND_DEEZER_ENABLED`, `ND_LASTFM_ENABLED`, `ND_LISTENBRAINZ_ENABLED` each default `true` (`:171,226,232`), `ND_SCANNER_SCANONSTARTUP` (`:258`) | Task 3: telemetry and every metadata agent off, so oracle responses never depend on the network |
| gonic `README.md` | "the default login is **admin**/**admin**" (`:31`); `GONIC_MUSIC_PATH` (`:54`), `GONIC_SCAN_AT_START_ENABLED` (`:65`), `GONIC_LISTEN_ADDR` (`:60`) | Task 3 |

Task 0 re-quotes each row against the base commit and the adopted versions. A changed quote is a note naming the
new text, and a removed behavior is a blocker.

## Prior art consulted (CLAUDE.md, build-for-0.8)

| Entry | Verdict here | Reason |
|---|---|---|
| OpenSubsonic specification (Apache-2.0) | Adopt as the contract | Every behavior the protocol packages implement cites it (spec `:316-318`). |
| OpenSubsonic bundled OpenAPI build (Apache-2.0) | Adopt, pinned by sha256 and committed with its license (Reading R3) | The contract tier's input (spec `:261-263`; survey `:101-102`). |
| Subsonic 1.16.1 XSD (no license text) | Deferred to 2b | The XML rule is slice 1's (spec `:939-940`). |
| Navidrome (GPL-3.0) | Adopt as an oracle, behavior only, pinned by digest | Spec `:273-280`. Implementers read its Apache-2.0 docs site and its HTTP responses, never its source. |
| gonic (GPL-3.0) | Adopt as the folder-view oracle, behavior only, pinned by digest | Spec `:276-278`. Its README is read for configuration only. |
| `getkin/kin-openapi` v0.149.0 (MIT) | Adopt as a module dependency of the conformance package (Reading R2) | Evidence below. |
| `pb33f/libopenapi-validator` (MIT) | Recorded as the fallback | Survey `:167`. The proxy lists v0.15.2 as latest on 2026-10-06, resolved from a branch ref; 2b's Task 0 re-reads it if kin-openapi fails. |
| `santhosh-tekuri/jsonschema/v6` (Apache-2.0) | Declined as a direct dependency | It validates schemas without routing a response (survey `:168`). kin-openapi carries it transitively. |
| `josephburnett/jd` v2 (MIT) | Declined | A generic JSON diff cannot express search as a set plus top-N order (survey `:171`). |
| `delucks/go-subsonic`, `supersonic-app/go-subsonic` (GPL-3.0) | Declined | A test binary linking GPL code is a derivative (survey `:172`). |
| `mdlayher/gosubsonic` (MIT) | Declined | Last push 2017-11-19. Token auth is `md5(password + salt)` over `net/http` (survey `:173`). |
| Ampache `SubsonicSpecConformanceTest.php` (AGPL-3.0) | Method studied, never read for code | Its `KNOWN_DEVIATIONS` list is asserted to still fail, so it cannot rot (survey `:161`). Task 6's staleness rule is that idea, reimplemented. |
| Melodee test scripts (MIT) | Read | A curl walk with token auth (survey `:162`). Nothing to adopt in Go. |
| `mewkiz/flac`, `go.senan.xyz/taglib`, modernc sqlite, goose, Picard tables, chromaprint | Not consulted | No task here decodes audio, writes tags beyond `tagio`, or touches a database. |

**kin-openapi adoption evidence (2026-10-06).** License MIT (GitHub API `license.spdx_id`). Newest version
v0.149.0, 2026-08-28; twelve tagged releases from 2026-05-08 to 2026-08-28; last push 2026-10-04; 103 open issues
and PRs; 3,298 stars; 481 importers of `openapi3filter` (survey `:166`). The survey's scratch module loaded the
bundled file with no `doc.Validate` error and caught a missing required array and a numeric `id`. The
`openapi3filter` plus `routers/legacy` import set pulls six non-stdlib modules: `go-openapi/jsonpointer`,
`go-openapi/swag/jsonname`, `oasdiff/yaml`, `oasdiff/yaml3`, `santhosh-tekuri/jsonschema/v6`, and
`golang.org/x/text` (measured with `go list -deps` in a scratch module). `routers/gorillamux` is not imported, so
`gorilla/mux` stays out of the build. The module is v0.x, so a minor bump may break; the `dependency-upgrade`
skill reads each bump's changelog as for any other dependency.

## Readings this plan takes

Each is settled here, with its reason. The review may contest any of them.

- **R1. The worktree serves this harness.** `CLAUDE.md:103-104` says "Corpus and harness work never moves to a
  worktree." The rule's reason is the beets venv: `tools/venv/.venv/` is untracked, so a worktree has none, and
  `make derive` would skip silently (parallel-build approach `:255`). This harness uses no venv, no beets, no
  `corpus/`, and no `make derive`. What a worktree lacks that it does need is `web/node_modules` (the gate's web
  leg), which the launch checklist provisions. The images live in the user's container store and the Playwright
  browsers and Unidecode wheel in the user cache, so every tree shares them. Ruling 5's fallback ("takes the next
  free slot instead") is therefore not taken. The conductor queues a `CLAUDE.md` edit that scopes the rule to the
  matcher corpus and the beets oracle.
- **R2. The conformance package carries kin-openapi.** The spec asks this plan to weigh it (`:732-736`). The
  checker is the module's distinctive offering, since no conformance suite exists anywhere (survey `:155`), and
  its value is routing a response to its schema, which only the dependency does. Keeping the checker in its own
  package means a consumer that imports the client or (from slice 3) the wire types never builds kin-openapi. The
  alternative, a checker that takes a validator interface and leaves the wiring in the harness, ships a checker
  that checks nothing. Task 7 asserts that `cmd/dubplate-server` never links the checker or kin-openapi.
- **R3. The pinned OpenAPI build is committed under harness data, and the checker takes it as input.** The gate
  opens no network connection, so the file is either committed or a third workstation prerequisite. Apache-2.0
  permits redistribution with the license text, and `opensubsonic/open-subsonic-api` carries no NOTICE file (its
  root listing on 2026-10-06 shows `LICENSE` and no `NOTICE`). The file lands at
  `internal/harness/subsonic/data/openapi/` beside a copy of that `LICENSE` and a provenance file (URL, fetch
  date, sha256, the spec repo commit). The conformance package embeds nothing and accepts the document from its
  caller, so the module's own license question (spec `:947-948`, owed to slice 1) does not touch this pass.
- **R4. The client builds only what 2a calls.** It issues a request, signs it with the legacy token scheme, and
  returns the verbatim response with its parsed envelope. Typed per-endpoint wire types arrive with slice 3's
  handler, their first consumer. The `apiKey` scheme has no oracle (spec `:286-287`) and no dubplate endpoint
  before slice 6, so it waits for slice 6. The plaintext `p=enc:` form is built only if an oracle refuses token
  auth (Task 3 finds out).
- **R5. Package names and homes.** The protocol layer is `internal/opensubsonic/`, the protocol's own name, with
  `client` and `conformance` as its first packages. The streamer harness is `internal/harness/subsonic/`, with
  `fixture`, `stack`, `session`, and `grade`, and its data under `internal/harness/subsonic/data/`. The subtree keeps
  the streamer harness apart from the matcher's packages while staying under `internal/harness`, so
  `TestEngineNeverImportsHarness` covers it unchanged, and so does the removal pass's acceptance that `go list
  -deps ./cmd/dubplate-server` names no `internal/harness` path (Task 7 turns that into a standing test). The names
  are reversible; the architecture reads grade them.
- **R6. The import boundary is a Go test, never a `scripts/check.sh` leg.** The spec asks for a gate leg that
  fails (`:312-314`). A test under `go test ./...` is one, and it follows the repo's existing invariant pattern.
  It also keeps this lane out of `scripts/check.sh`, which 6d contests. Read literally, the spec's "no
  `github.com/glw907/dubplate/` path" would fail on the layer's own packages, so the leg asserts that every
  module-internal dependency lies inside `internal/opensubsonic/`.
- **R7. Podman is the container runtime.** It is the workstation's runtime (`podman version 5.8.7`). The stack
  names it on failure and accepts no other. A contributor with only Docker is an `IDEAS.md` line at the close if a
  reviewer raises it.
- **R8. The seed session covers what slice 3 grades first.** It records the system and browse endpoints in both
  formats against both oracles: `ping`, `getLicense`, `getOpenSubsonicExtensions`, `getMusicFolders`,
  `getIndexes`, `getMusicDirectory`, `getArtists`, `getArtist`, `getAlbum`, and `getSong`. Each later slice
  records its own endpoints, as their first consumer.
- **R9. One dependency sweep per window, and kin-openapi lands when 6d allows.** Two lanes bumping `go.mod`
  collide at the merge. 6d revision 2.3 contests `go.mod` during its own Task 0 sweep and says this slice's
  addition "lands before Task 0 or after 6d's merge" (6d `:1537-1539`). So Task 0's sweep runs on master in the
  main tree, never in the worktree, and only while no 6d sweep is in flight; if 6d's sweep has landed within the
  day, this pass records it as its own in one line. The worktree branches from the post-sweep commit, and the
  branch's only `go.mod` change is kin-openapi. kin-openapi cannot land on master ahead of the branch, because
  `tidy-check` drops a requirement nothing imports. Task 9's merge therefore takes a window 6d's rule allows.
- **R10. Pins follow slice 1's first-pin-wins rule.** Slice 1's Task 0 pins Navidrome, gonic, and the
  OpenSubsonic specification commit, and adopts any pin this lane made first (slice 1 plan, settled item 5). Task 0
  here does the same from the other side. Slice 1 runs its own probe oracles and builds its own scratch fixture,
  so 2a hands it nothing and waits on it for nothing.

## Ruling for Geoff

**1. Does slice 2 run as two passes, 2a now and 2b after slice 1, under ceilings of 9.5M (flag 7.6M) and 4.0M
(flag 3.2M)?** Recommendation: **yes.** Reason: the slice is twelve build tasks, past the ten-task line, and
four of them (the XML rule, the seeded ledger rows with the per-endpoint oracle assignment, the encoding-quirk
cases, and the roster walk's content) consume slice 1 deliverables that do not yet exist. Planning them now would
plan against guesses, the plan-staleness defect this build has paid for. What yes builds: 2a merges on its own,
its tests become the "test owed" targets slice 1's ledger names, and 2b's plan revision is authored against slice
1's record. What no builds: one pass whose last segment stalls until slice 1 closes and Geoff rules its forks,
holding the worktree and the lane open meanwhile, with four tasks planned against guesses.

## Global constraints

- **Where it runs.** Worktree `/var/home/glw907/Projects/dubplate/.claude/worktrees/lane-3-slice-2`, branch
  `lane-3-slice-2`, cut from master after Task 0's sweep commit. Lane 1 (6d) keeps the main tree.
- **Territory (exclusive to this lane).** New: `internal/opensubsonic/**` and `internal/harness/subsonic/**`.
  Shared but uncontested: `cmd/harness/main.go`'s registration block (one line per verb) and three new verb files
  in `cmd/harness/`; `internal/harness/conform/arch_test.go` (one new test, Task 7). Read-only: `internal/engine/tagio`.
- **Contested with lane 1.** `go.mod` and `go.sum` only, for kin-openapi, under R9's timing. This lane never edits
  another file on 6d's contested list (6d `:1526-1539`), `scripts/check.sh` among them, and it leaves `Makefile`
  alone too.
- **Contested with slice 1.** The oracle pins (R10) and, in 2b, the Android emulator install, which slice 1 makes
  and records and this lane reuses (slice 1 plan, settled item 9). Ports and `podman` containers never collide:
  each run here picks its loopback ports at start and names its containers per run. Queued edits that land on master
  after the merge, through the conductor: `CLAUDE.md` (R1's wording; the gate paragraph naming the new invariant
  tests and, at 2b, any new workstation prerequisite), `docs/STATUS.md`, and `docs/HISTORY.md`.
- **Clean room.** The protocol packages cite the OpenSubsonic specification or an observation recorded in a
  session, never GPL source (spec `:316-318`, `:341-347`). No implementer opens Navidrome's or gonic's source
  tree. Configuration comes from Navidrome's Apache-2.0 docs site and gonic's README.
- **The boundary.** Nothing under `internal/opensubsonic/` imports a module-internal path outside that subtree,
  tests included. Task 2 makes this a failing test.
- **No export ahead of a caller.** Every exported identifier in a new package has a non-test caller in this pass,
  and the staged exported-unused leg reports no hit in the new packages at the close.
- **Go** follows `go-conventions`: `ctx` first, every subprocess under `exec.CommandContext`, wrapped errors with
  sentinels matched by `errors.Is`, `encoding/json/v2` with `json.RejectUnknownMembers(true)` for every committed
  file this pass defines, atomic writes through `internal/atomicfile` for every file a verb writes, table-driven
  tests under the assertion-discipline rules.
- **Every check reports four states.** For each loader and check below, the task's states table names the report
  for an absent input, an empty one, a malformed one, and one with many bad entries. Many-bad reports every bad
  entry in one error (`errors.Join`), never the first only.
- **Determinism.** Every report and every file a verb writes is byte-identical across runs over the same input:
  sorted keys, sorted findings, no wall-clock field unless the file's schema names one.
- **Forward hooks.** No task lands a forward hook from spec `:767-855`. Every task is neutral; the harness reads
  no catalog, roster, or store surface.

## Lane launch (the conductor, before Task 0's sweep closes)

From parallel-build approach `:249-258` and `CLAUDE.md:99-110`:

1. **Reverse edges cleared.** 6d's territory (its "Contested surfaces") and this lane's are disjoint except
   `go.mod`/`go.sum`. Slice 1 is a document lane writing under `docs/` only, and it neither reads nor waits on
   anything 2a builds.
2. **Contested surfaces written** to STATUS's lane 3 row: `go.mod`/`go.sum`, with R9's rule.
3. **Runaway guard armed** for this lane's transcript directory, per `~/.claude/docs/unattended-work-guards.md`.
4. **Lane registry row** in STATUS: lane 3, the worktree path, the branch, the plan path, the sub-ceiling.
5. **Provisioning.** No venv (R1). In the worktree: `npm ci --prefix web` (the gate's web leg refuses without it).
   Workstation tools checked present and recorded with versions: `podman`, `ffmpeg`, `flac`. `make
   unidecode-wheel`'s cache is shared and already verified.
6. **Sub-ceiling** assigned (above).
7. **AC power**, and the throughput baseline: the conductor's one full gate in the fresh worktree, timed and
   quoted (Task 0 takes no gate of its own).

## Task 0: Sweep, pins, and pointer re-resolution (conductor)

The conductor runs it through the `dependency-upgrade` skill, since implementers carry no Skill tool.

**Outcomes.**
- The sweep over `go.mod`, `web/package.json`, `tools/venv/requirements.txt`, and the pinned binaries lands on
  master in the main tree (R9), or is one line in the pass record when nothing moves. The last sweep is
  `docs/superpowers/research/2026-10-05-dependency-sweep.md`.
- The worktree is cut from the post-sweep master and the launch checklist is complete.
- Pins resolved and recorded in the task's report: kin-openapi's newest version (v0.149.0 on 2026-10-06; a newer
  release supersedes it, with the quotes table re-read); the bundled OpenAPI build's sha256 and size on the day of
  pinning. For Navidrome (`deluan/navidrome`, newest release v0.64.2 of 2026-09-24), gonic (`sentriz/gonic`, newest
  release v0.22.0 of 2026-06-08), and the specification commit, the first pin wins (R10): if slice 1's Task 0 has
  handed over its pins, this task adopts them; otherwise it pins each image as a `sha256` digest read from the
  registry and the conductor hands the list to slice 1's lane.
- Every `file:line` in this plan re-resolved against the worktree base. A moved pointer is a note; a pointer whose
  content is gone is a blocker.
- **The OpenAPI pin's timing against slice 1.** The spec has slice 1 re-read OpenSubsonic PRs #237, #162, #258,
  and #261 before the harness pins a spec build (`:951-955`); slice 1's Task 1 does that read. If its record has
  not reached this lane when Task 7 starts, Task 7 pins the build at the specification commit already pinned, and
  2b's Task 0 re-pins if the re-read moves anything. The pin is replaceable by design, and a re-pin re-runs the
  ledger.

**Acceptance.** The report lists each pin with its evidence and each re-resolved pointer. No gate.

## Segment A

### Task 1: The fixture library

**Files:** `internal/harness/subsonic/fixture/` (new), `internal/harness/subsonic/data/fixture/` (manifest and
seeds). **Hook:** neutral.

**Outcomes.**
- A committed manifest declares the library: artists, albums, discs, tracks, every tag value, every relative path,
  and which albums carry a `cover.jpg`, an embedded picture, both, or neither. Paths are data in the manifest,
  never computed, because rung 7 owns the layout (hook 12) and has not landed it.
- The manifest covers the shapes the browse views and the folder view must meet: a single-disc album; a
  multi-disc album in disc subfolders and one with disc-prefixed filenames in one folder; a compilation under a
  various-artists album artist; a track with a featured artist carried as a multi-valued artist tag; two albums
  with the same title by different artists; one artist with several albums; names with diacritics and with
  non-Latin script; sort tags that differ from display names; release MBIDs on some albums and none on others; and
  a track missing every optional tag.
- Committed audio seeds are short real encoded FLAC (decodable by `ffmpeg`), so later slices can stream them.
  Each seed and image seed carries its sha256, and a README beside them records the command that produced it.
  Committed fixture bytes stay under 512 KiB.
- A builder writes the library to a target directory by copying seeds and tagging each track through
  `tagio.Write`. Two builds into two directories are byte-identical, file for file.

**States table.** Absent manifest: an error naming the path. Empty (no albums): refused, "manifest declares no
albums". Malformed or carrying an unknown member: an error with the JSON location. Many bad (several tracks with a
missing seed, a duplicate path, a path escaping the root, or a seed whose sha256 differs): every one named in one
error, and nothing written to the target.

**Test mandate and fixture state.** The determinism proof builds the committed manifest twice and compares every
file's bytes; it asserts the file count equals the manifest's track count plus its image files, so an empty build
cannot pass. A round-trip test reads every built track back through `tagio.Read` and compares each manifest tag.
**Mutations:** drop one tag write (the round trip fails); stamp the build time into a file (determinism fails);
skip the seed checksum (the bad-seed case passes and the test fails).

### Task 2: The protocol client and the import boundary

**Files:** `internal/opensubsonic/client/` (new), the boundary test (below). **Hook:** neutral.

**Outcomes.**
- The client issues one request for a named endpoint with given parameters and a chosen format (`json` or
  `xml`), against a base URL that may carry a sub-path. It sends `u`, `t`, `s`, `v`, `c`, and `f`, with a fresh
  salt per request from `crypto/rand`, at least six characters (api-reference `:55-57`).
- It returns the HTTP status, the headers, the body bytes verbatim, and the parsed envelope for either format:
  `status`, `version`, `type`, `serverVersion`, `openSubsonic`, and, on failure, the error `code` and `message`.
  A failed envelope is a typed error the caller matches with `errors.As`.
- Its doc comments cite the OpenSubsonic docs for every behavior. It imports nothing module-internal.
- **The boundary leg.** A test lists `go list -deps -test ./internal/opensubsonic/...`, cuts each line at its
  first space, and fails naming every module-internal path outside `internal/opensubsonic/`. The parsing half
  takes the listing as input, so a second test feeds it a canned listing with one violating path and asserts the
  failure names it. A failing `go list` fails the test with its output.

**States table.** Non-200 HTTP: an error carrying the status. An empty body: "empty response". A body that parses
as neither envelope: an error naming the format and the first bytes. An envelope with `status="failed"`: the typed
error with code and message. A JSON envelope missing the `subsonic-response` wrapper: malformed, never a success.

**Test mandate and fixture state.** The token test reproduces the spec's worked vector: password `sesame`, salt
`c19b2d`, token `26719a1196d2a940705a59634eb18eab`. Envelope tests run against an `httptest.Server` serving
hand-written JSON and XML envelopes in both states, including the attribute-heavy XML form. **Mutations:** hash
`salt + password` (the vector fails); reuse one salt across two requests (a test issuing two requests compares
salts); accept a bare JSON object without the wrapper (the malformed case passes and the test fails); drop the
prefix filter in the boundary parser (the canned violation passes and the test fails).

### Task 3: The oracle stack and `harness oracle-up`

**Files:** `internal/harness/subsonic/stack/` (new), `cmd/harness/oracleup.go` (new verb), one line in
`cmd/harness/main.go`. **Hook:** neutral.

**Outcomes.**
- The stack starts the pinned Navidrome and gonic images under podman, each through `exec.CommandContext`, with
  the fixture library (built by Task 1 into a temporary directory) mounted read-only, each on a loopback port
  chosen at start, under container names unique to the run.
- An image reference without a `@sha256:` digest is refused before podman runs. The digests live in one committed
  file the stack reads, and the session records them (Task 5).
- Navidrome runs with insights and the Deezer, Last.fm, and ListenBrainz agents off and a scan on startup;
  gonic with a scan at start. The admin credential is generated per run and never written to disk. Navidrome's
  first-admin path is resolved from its docs site or by observing its HTTP behavior, never from its source, and
  the task records which.
- Readiness means a `ping` through the Task 2 client answers `ok` and the scan has finished, read through
  `getScanStatus`. Readiness has a deadline. Teardown removes both containers on every path: success, failure,
  and context cancellation.
- Whether each oracle accepts token auth is recorded. If one refuses it, the client gains the `p=enc:` form in
  this task, since the stack is then its caller (R4).
- `harness oracle-up` builds the fixture, starts the stack, prints each oracle's base URL and the run's
  credential, and holds until interrupted, then tears down. Its callers are this task's live evidence run and
  every conductor roster walk from 2b on. It takes no base-path flag: slice 1 measures the sub-path base URL on
  its own probe oracles, and `ND_BASEURL` and `GONIC_PROXY_PREFIX` stay unset until a caller needs them.

**States table.** `podman` absent from `PATH`: an error naming it. A tag-only image reference: refused, naming the
image. Digest file absent, empty, or malformed: an error naming the file. Readiness past the deadline: an error
carrying each container's last log lines. A scan that never finishes: a timeout error naming the oracle. Both
oracles failing: both named in one error.

**Test mandate and fixture state.** Gate tests drive the stack against a fake container runtime (a helper process
re-executed from the test binary), proving the argv (read-only mount, loopback publish, the telemetry and agent
variables, the digest), the refusals, the deadline, and teardown on cancellation. The fake must observe a removal
call for every container it started. The live path runs once on the workstation as the task's evidence: the
implementer runs `harness oracle-up`, pings both oracles, and quotes the output in its report. **Mutations:** omit
`:ro` on the mount; drop `ND_DEEZER_ENABLED=false`; skip teardown on the cancellation path; accept a tag-only
reference. Each must fail a named test.

## Segment B

### Task 4: The session format

**Files:** `internal/harness/subsonic/session/` (new). **Hook:** neutral.

**Outcomes.**
- A session is one directory: a metadata file and one verbatim body file per (request, oracle) pair. Bodies are
  never re-encoded (`cassette.go:5-13`).
- The metadata records the fixture manifest's sha256, each oracle's image digest and reported `serverVersion`,
  and the ordered requests. Each request names its endpoint in canonical form (the `.view` suffix dropped), its
  format, and its non-auth parameters sorted byte-wise, with entity references in symbolic form (a manifest key,
  never an oracle id). Each response names its oracle, HTTP status, `Content-Type`, and body file.
- No auth parameter (`u`, `p`, `t`, `s`, `apiKey`) is ever written to a session. The writer refuses a request that
  carries one.
- A loader reads and validates a session; a writer writes one atomically.

**States table.** Absent directory or metadata file: an error naming it. Empty request list: refused. Malformed or
unknown-member metadata: an error with the location. Many bad (missing body files, duplicate request keys, a
response for an oracle the metadata does not declare, a request carrying an auth parameter, a body file nothing
references): every one named in one error.

**Test mandate and fixture state.** Tests run on hand-built sessions in `testdata`, at least one valid session
with two oracles and both formats, plus one per bad state. A round trip writes and reloads a session and compares
bytes. **Mutations:** allow `t` through the writer (the auth refusal test fails); stop flagging orphan bodies; read
bodies through a JSON decode-and-marshal (the byte round trip fails).

### Task 5: `harness record-oracle` and the seed session

**Files:** `cmd/harness/recordoracle.go` (new verb), one line in `cmd/harness/main.go`, a session spec and the
seed session under `internal/harness/subsonic/data/sessions/seed/`. **Hook:** neutral.

**Outcomes.**
- The verb reads a session spec (the requests in order, with symbolic references), starts the stack, issues each
  request against each oracle through the client, and writes a session through Task 4's writer.
- A symbolic reference resolves against responses already in the session. An album key resolves through
  `getArtists`, then `getArtist`, then a match on the manifest's album artist and title; a folder key resolves
  through `getIndexes` and `getMusicDirectory`. A reference that no earlier response reaches, or that matches more
  than one entity, refuses the run, naming the key and the oracle.
- The session's oracle digests come from `podman image inspect` on the running images, never from the spec file.
  Its `serverVersion` comes from the oracle's own envelope.
- The seed session covers R8's ten endpoints in JSON and XML against both oracles, with `getArtist`, `getAlbum`,
  `getSong`, and `getMusicDirectory` requested for the same-title pair, the compilation, and a multi-disc album.
- The implementer runs the verb once live and commits the seed session.

**States table.** Spec absent, empty, or malformed: an error naming the file. An unresolvable or ambiguous
reference: refused before anything is written. One oracle failing mid-run: nothing written, the request and
oracle named. A spec request carrying an auth parameter: refused.

**Test mandate and fixture state.** The gate tests run the verb against a fake stack serving canned responses, so
resolution and refusal are proven offline. The ambiguity proof needs the same-title pair present in the canned
responses, and the unreachable proof needs a reference to an album absent from them. The committed seed session
must load through Task 4's loader with every request answered by both oracles. **Mutations:** resolve on title
alone (the same-title pair resolves to the wrong album and the test fails); take the digest from the spec; write
a partial session when an oracle fails.

### Task 6: The divergence ledger

**Files:** `internal/harness/subsonic/grade/` (new; the ledger half), `internal/harness/subsonic/data/ledger.json`.
**Hook:** neutral.

**Outcomes.**
- Each ledger row carries every column of slice 1's Markdown ledger, so 2b's transcription is mechanical (slice 1
  plan, settled item 6 and Task 7): an id, the endpoint or area, the class, the oracle, the oracle's behavior,
  dubplate's intended behavior, and an evidence pointer (a spec section or a slice 1 record line). It adds what a
  machine needs: a state, `active` or `owed`. An `active` row carries a scope the grader matches (format and a
  path pattern into the response), the session that exercises it, and the oracle version it was adjudicated
  against. An `owed` row names the slice whose plan builds its test, since a row about dubplate's own behavior
  (a range on a filling transcode, say) has nothing to match until that slice serves a response. The report
  counts `owed` rows by slice.
- Classes: `by-construction` (an oracle cannot grade this surface), `unoracled` (no oracle covers it), `contract`
  (an oracle's response violates the OpenAPI file), and `volatile` (the value varies per run or per server;
  presence and type are still compared).
- The committed ledger seeds the spec's two declared classes (`:285-287`): Navidrome's folder view diverges by
  construction for `getIndexes`, `getMusicDirectory`, and `getMusicFolders`, and the `apiKey` path is unoracled.
  It adds the `volatile` rows the seed session needs (`serverVersion`, `type`, and any timestamp field). 2b adds
  slice 1's pre-seeded rows.
- **No active row can rot.** Grading a session fails on an `active` row that names the session and matches no
  finding in it, naming the row.
- **The oracle version is privileged.** An `active` row's version must equal the session's recorded
  `serverVersion` for that oracle. Only a ledger edit sets it, and the grader refuses a mismatch, so an oracle bump
  forces the deliberate re-adjudication pass the spec requires (`:279-280`). The session's own version is set only
  by the recorder from the oracle's envelope (Task 5).

**States table.** Absent ledger: an error (the ledger is required). Empty (`[]`): valid, and every finding then
reports unledgered. Malformed or unknown-member: an error with the location. Many bad (duplicate ids, unknown
class, an `active` row with an empty scope or no session, an `owed` row naming no slice, a version mismatch, a
stale row): every one named in one error.

**Test mandate and fixture state.** Tests use fixture ledgers against a hand-built findings list holding at least
one match per class. The staleness proof needs an `active` row that matches nothing in a non-empty findings list,
so an empty list cannot make it pass. A transcription proof converts a three-row Markdown table in slice 1's
column order into rows and back. **Mutations:** skip the staleness check; compare versions by prefix; let a
`volatile` row suppress a type change; let an `owed` row suppress a finding.

## Segment C

### Task 7: The contract tier

**Files:** `internal/opensubsonic/conformance/` (new), `internal/harness/subsonic/data/openapi/` (the pinned
file, its `LICENSE`, and a provenance file), `go.mod` and `go.sum` (kin-openapi), one test in
`internal/harness/conform/arch_test.go`. **Hook:** neutral.

**Outcomes.**
- The checker loads an OpenAPI document from bytes its caller supplies and validates a JSON response body for a
  named endpoint. It routes by path (the quotes table explains why a full URL fails), and returns one of four
  outcomes: valid; violations, each with its JSON path and reason, all of them; no route for the endpoint; or no
  JSON schema for the status and content type. Only the first is a pass.
- The harness verifies the pinned file's sha256 before handing it to the checker. A mismatch fails, printing both
  sums.
- **The server never links the checker or the harness.** The new test in `arch_test.go` asserts that `go list
  -deps ./cmd/dubplate-server` names no `internal/harness` path, no `internal/opensubsonic/conformance`, and no
  `github.com/getkin/` path. The
  harness `go list` parsing from Task 2 is not shared across the boundary; this test carries its own.
- kin-openapi enters `go.mod` at the version Task 0 pinned.

**States table.** Document absent or failing its checksum: an error naming the file and both sums. A document
that fails to load: the loader's error. An empty body: a violation, never a pass. A body that is not JSON: a
violation naming the parse error. A wrong `Content-Type`: a violation naming it. Many violations: all of them.

**Test mandate and fixture state.** Proof responses are hand-built from the seed session's real `getAlbum` body:
the untouched body passes; one with the required `song` array removed and one with a numeric `id` each fail, and
the assertion checks the reported path. The no-route outcome needs an endpoint absent from the file. The
no-schema outcome needs a status the file does not declare. **Mutations:** drop `MultiError` (the two-violation
body reports one and the test fails); treat no schema as valid; route by full URL (every proof reports no route).

### Task 8: The grader and `harness grade`

**Files:** `internal/harness/subsonic/grade/` (the differ half), `cmd/harness/grade.go` (new verb), one line in
`cmd/harness/main.go`. **Hook:** neutral. **Model:** opus.

**Outcomes.**
- The grader takes a session, the ledger, the pinned document, and a pair of sources to compare, and produces one
  report holding two tiers. The contract tier runs every JSON response through the checker. The oracle tier diffs
  the two sources' responses for each request in the same format.
- Both formats decode into one comparable tree. JSON numbers keep their literal through `jsontext` tokens, and a
  string versus a number is a finding, since clients break on it. XML elements, attributes, and text map to the
  tree, and a self-closing element equals its expanded form. A JSON singleton where the other side has an array is
  a finding, never normalized away.
- Ids are opaque per server. The grader pairs ids by the position they occupy in corresponding responses and
  reports a conflict when one id pairs with two. Request parameters compare by symbolic key.
- Search endpoints compare as a set plus a top-N ordering assertion (spec `:287-289`). No search endpoint is in
  the seed session, so the rule is proven on a fixture pair.
- Each finding carries the request, the path, both values, and the matching ledger row or "unledgered". The report
  is deterministic. The verb writes it to a file and exits nonzero on any unledgered finding or stale row.
- **The gate test grades the seed session**, Navidrome against gonic. It asserts at least one contract-tier
  validation ran per oracle, at least one `by-construction` folder-view finding, zero unledgered findings, and zero
  stale rows. Every `contract` or extra `volatile` row this needs is added to the ledger with its reason, and the
  task's report lists each.

**States table.** A request with a response from only one source: a finding naming the missing source. An empty
body on either side: a finding. A body that fails to parse: a finding naming the format and the error. Many
findings: all of them, sorted.

**Test mandate and fixture state.** A perturbation table starts from one real seed-session pair and applies one
change per row: a changed value, a missing field, an extra field, a string-for-number swap, a singleton-for-array
swap, a reordered order-significant array, an id that pairs twice, and an XML attribute moved to a child element.
Each row asserts exactly one finding at the expected path. The unperturbed pair asserts zero findings outside the
ledgered classes. **Mutations:** decode numbers into `float64`; treat arrays as unordered everywhere; skip the
contract tier when the oracle tier is clean.

## Close (2a)

### Task 9: Reviews, merge, and the ledgers

The conductor runs these in order. Every reviewer dispatch asks for correctness gaps ranked by consequence, not
polish, and a fold may refuse a finding with a one-line reason.

1. One `code-simplifier:code-simplifier` dispatch over the pass's changed Go, pasting
   `.claude/instructions/simplifier-brief.md` verbatim, then one `diff-reviewer` read of its diff.
2. One fresh-context vacuity read (`claude-opus-5-5`) over the gate tests in the six new packages and
   `arch_test.go`. Its single question: which test could pass while the behavior it names is broken? One fold
   task if it finds any.
3. The full gate in the worktree, then the merge ritual, in a window 6d's rule allows (R9: before 6d's Task 0
   starts, or after 6d merges): the merge token; `git rebase master` (resolve `go.mod`
   and `go.sum` by taking master's files and re-adding kin-openapi at the pinned version, then `go mod tidy`);
   `check.sh --merge` in the worktree; `git merge --ff-only lane-3-slice-2` on master; the full gate on master;
   one `diff-reviewer` over the merged range if the rebase was not a fast-forward.
4. One `go-architecture-reader` per touched Go package, seven in all: `internal/opensubsonic/client`,
   `internal/opensubsonic/conformance`, `internal/harness/subsonic/fixture`, `.../stack`, `.../session`,
   `.../grade`, and `cmd/harness`. One fold task for accepted findings.
5. Queued repo edits in one batch on master: `CLAUDE.md` (R1's wording; the gate paragraph naming the boundary
   test and the no-link test), HISTORY's 2a entry (what landed, what the gate and reviewers caught, what slice 3
   would be wrong to rediscover, the throughput baseline, both budgets scored), and STATUS (lane 3's row, 2b's
   wait on slice 1).
6. Worktree teardown, or the worktree held for 2b if slice 1 is close; the registry row says which.

**Acceptance.** Master is green with the merged work. The staged exported-unused leg reports no hit in the six new
packages. `go list -deps ./cmd/dubplate-server | grep -E 'internal/(harness|opensubsonic)|getkin'` prints
nothing. `git grep -nE '[?&](t|s|p|apiKey)=' -- internal/harness/subsonic/data` prints nothing.

## Pass 2b (sketch; authored in full at its Task 0)

2b starts when slice 1's record lands. Its Task 0 re-reads that record and re-derives its ceiling, and its plan
revision replaces this sketch. It names these slice 1 deliverables as inputs, and each blocks only its own task.

| 2b task | Slice 1 input it waits on | Spec source |
|---|---|---|
| 10. The XML rule: validate XML per slice 1's choice (a legacy form against the official XSD, or a derived XSD), through `xmllint`, which the test names and fails without. The XSD carries no license text (survey `:272`), so it is fetched into the user cache and verified by sha256, never committed, unless slice 1's record settles its license. A new workstation prerequisite joins `CLAUDE.md`'s gate paragraph. | The XML validation rule | `:269-271`, `:939-940` |
| 11. The seeded ledger and the oracle assignment: slice 1's pre-seeded divergence rows transcribed into Task 6's machine form, each `active` where a recorded session exercises it and `owed` otherwise, and the per-endpoint oracle assignment (contract only, Navidrome, gonic, or the roster) as data the grader reads when one side is dubplate. The assignment's first non-test consumer is slice 3's grade, so 2b builds the data and its loader, and the verb's dubplate-side flag waits for slice 3. | The pre-seeded ledger; the per-endpoint oracle assignment | `:728-730` |
| 12. Encoding quirks: a session and a perturbation row for each quirk slice 1 enumerates (the attribute-heavy XML form, the wrapper, singleton versus array, JSONP). | The encoding-quirk list | `:883` |
| 13. The roster walk: a walk script per roster client (Feishin and Rufin on the desktop under the isolated configuration homes slice 1 sets up, Tempus in the emulator slice 1 installs), each step stating the observation to record, run once against `harness oracle-up` to prove the script, with the record committed under `docs/superpowers/research/`. Amperfy, which the workstation cannot drive, is listed for rung 16's attended gate (re-plan `:140-141`). | The flows each client exercises (slice 1's client-demands record) | `:732-734`; roster audit `:16-19` |
| Re-pin, if slice 1's PR re-read moved the OpenAPI build. | The re-read of PRs #237, #162, #258, #261 | `:951-955` |

**One slice 1 owner fork reaches 2a's work.** Slice 1 brings "folder browsing at launch" to Geoff (its Task 7).
If he rules it out, gonic loses its folder-view role but keeps `getMusicFolders` (spec `:276-278`), the seed
session's `getIndexes` and `getMusicDirectory` rows become reference observations, and 2b's Task 0 retires the
`by-construction` ledger rows that grade them. 2a builds the spec's design as written, since gonic's place in the
stack is one image and one set of variables, and waiting on the ruling would stall segment A.

## Not verified at planning

- Navidrome's first-admin path for an unattended start. The docs read here name no environment variable for it.
  Task 3 resolves it from the docs site or by observation.
- Whether gonic accepts the token scheme. Task 3 records it (R4).
- The Navidrome image's tag spelling for v0.64.2 on Docker Hub, and both digests. Task 0 resolves them.
- Whether `tagio.Write` writes a multi-valued artist tag in the shape Navidrome and gonic split. Task 1's round
  trip proves the write. Task 5's live recording shows how each oracle reads it.
- The conform suite's run time on every task gate. The first task's gate measures it. If it dominates, the
  conductor files a `--scope-changed` refinement for `scripts/check.sh` after 6d.
