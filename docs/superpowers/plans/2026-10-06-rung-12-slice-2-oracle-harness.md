# Rung 12, Slice 2 (the Oracle Harness) Implementation Plan

Status: **revision 1.2, approved by Geoff 2026-10-07; its `spec-plan-review` is folded twice.** Ruling 1 is answered
yes. Every other decision is a reading below, taken on its merits and reported.

History:

- Revision 1 (`f62a151`): drafted.
- Revision 1.1: folds the three lens reviews (contract, mechanics, and risk) per
  `docs/superpowers/research/2026-10-06-rung-12-slice-2-plan-fold.md`. The 2a proof no longer grades Navidrome
  against gonic under the committed ledger: the differ is proven on a same-source perturbation table, and the
  real oracle pair is graded against a test-only ledger and pinned as a golden report. Ids pair by fixture
  identity, never by position. The checker classifies status, content, and schema itself before kin-openapi runs,
  and reports only the envelope branch the body claims. Every task gate runs on the heavy lane, because
  `--scope-changed` runs `make check` whenever the change set touches `internal/harness/`. 6d merges first, and
  this branch waits for it. The oracle configuration the live probes measured (the SELinux mount label, gonic's
  paths and multi-value settings, Navidrome's first admin, token auth) is now stated. The Markdown transcription
  proof moves to 2b and the search comparison to slice 4. The ceiling is re-derived from actual spend: 8.5M.
- Revision 1.2: folds the fold verification read
  (`docs/superpowers/research/2026-10-06-rung-12-slice-2-plan-fold-verification.md`, V1 to V12) per the fold
  record's "Second fold". The library digest is recorded per source, and a later source joins a session through
  Task 4's writer. A `contract` finding and a `contract` row each name one source. The perturbation table drives
  the differ alone, and a named grader test kills the skip-contract mutation. The ceiling prices seven Sonnet build
  tasks: 9.0M.

**Spec:** the 0.8 streamer spec (`docs/superpowers/specs/2026-09-07-streamer-2-0-design.md`, revision 3, "the
spec"; it still says 2.0), slice 2 at `:732-736`, under the re-plan spec
(`docs/superpowers/specs/2026-10-05-rungs-12-16-replan-design.md`, revision 2.1, "the re-plan"), ruling 5 at
`:48-54`.

**What this pass is.** It builds the evidence machinery every later rung 12 slice is graded by. The protocol
layer gets its first two packages, a client and a conformance checker, under a gate-enforced import boundary. The
harness side gets a byte-deterministic fixture library, the pinned Navidrome and gonic containers as
workstation-only behavior oracles, a recorded session format, a divergence ledger that cannot rot, and a grader
that runs the contract tier and the oracle differ over a session.

**What slice 3 consumes, and so what 2a must prove.** The spec's oracle tier "diffs dubplate's responses against"
the oracles (`:274-276`), and its contract tier validates every JSON response (`:261-263`). Slice 3 is the first
slice with a dubplate side. It needs a differ that reports exactly one finding per real difference, a contract
tier whose only pass is a real pass, a ledger whose rows match only the pair they were adjudicated on, and a
recorded oracle session over the same library bytes a slice 3 test builds, which that test joins by adding dubplate
as a source through Task 4's writer. 2a proves each directly:

- **The differ** on a same-source perturbation table that drives the oracle-tier differ alone: one real recorded
  body against itself, so the unperturbed baseline is zero findings and each perturbation yields exactly one
  (Task 8).
- **The contract tier** on a hand-authored document with every kin-openapi silent-pass path (Task 7), and on the
  real oracle bodies, whose exact violation lists are asserted (Task 7, harness side).
- **The ledger** on fixture ledgers and fixture findings (Task 6).
- **The whole pipeline over real servers** by grading the seed session's Navidrome responses against gonic's under
  a test-only ledger, with the full report pinned as a committed golden (Task 8). The spec never asks for an
  oracle-against-oracle grade, so it is a test of the grader and never a source of committed ledger rows.

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

**Ceiling (2a):** 9.0M execution tokens, flag 7.2M (80%). Derivation from actual spend: the removal pass spent
about 3.7M on seven tasks and its close (`docs/HISTORY.md:191`), about 0.45M a task with the close folded in. 2a
prices seven Sonnet build tasks, Tasks 1 to 7, at 0.5M (3.5M), a 0.4M premium on each live task, Tasks 3 and 5 (0.8M), the Opus
Task 8 at 1.5M, one fix round a segment plus one (1.0M), and a close of seven architecture reads, one simplifier
round with its review, one vacuity read, and two fold tasks (2.2M). The fold moved scope both ways (identity
pairing, the branch filter, the golden, and the library digest in; the transcription proof and the search rule
out), and the net is priced above. The sum is 3.5 + 0.8 + 1.5 + 1.0 + 2.2 = 9.0M; revision 1.1 priced six
Sonnet tasks and so left one task out. Planning and plan review are recorded outside it.
**Ceiling (2b):** 4.0M, flag 3.2M, re-derived at 2b's Task 0 from 2a's actual rate.
**Lane sub-ceiling:** lane 3's sub-ceiling is 2a's ceiling, then 2b's. The conductor reconciles both against the
phase's combined ceiling at each checkpoint.
**Checkpoint interval:** each segment boundary (three tasks apart at most) and the close. STATUS's lane 3 row is
written at each.
**Segments (2a):** Task 0 (conductor) first. A: Tasks 1, 2, 3. B: Tasks 4, 5, 6. C: Tasks 7 and 8. Then the
close's reviews, the merge ritual, and Task 9. Every boundary lands on a green full gate.
**Tasks (2a):** ten. Task 0, eight build tasks, and the close.

**Executor:** `dubplate-implementer`, `model: sonnet`, effort `high`, for Tasks 1 to 7. Task 8 runs on
`model: opus` from the start, because the grader's identity pairing is novel correctness-critical logic this
plan specifies only by outcome. A second `fix` verdict on Task 5 is the one upshift trigger the conductor weighs
first, since its failure mode is a reference that resolves to the wrong entity and still records.
**Per-task gate:** `cairn-run-gate 'bash /var/home/glw907/Projects/dubplate/.claude/worktrees/lane-3-slice-2/scripts/check.sh --scope-changed'`,
on the heavy lane, re-issued on exit 75 until it prints `gate exit:`. The heavy lane is required. On a branch the
change set is everything since `git merge-base master HEAD` (`scripts/check.sh:183`), so from Task 1 on it
holds an `internal/harness/` path, and the conform suite runs. The rendered leg is skipped only inside the
conform-skipped branch:

```
if [ "$mode" = "scope-changed" ] && [ "$conform_affected" -eq 0 ]; then
  scope_conform=skipped
  if [ "$rendered_affected" -eq 0 ]; then
    scope_rendered=skipped
```

(`scripts/check.sh:270-275`). Otherwise the script runs `make check` (`:364-367`), and `Makefile:134` reads
`check: web vet lint test web-rendered tidy-check`. Every task gate therefore launches Chromium, and a browser
gate belongs under the machine's heavy lock (`~/.claude/docs/pass-gate-economy.md`). Decoupling the rendered
scope from the conform scope is one condition in `scripts/check.sh`, which 6d contests; it is filed for after 6d
(see "Not verified at planning"). The comment-only fix round takes `check.sh --reduced`. Segment boundaries and
the pass end take the full form (`check.sh` with no flag), and the merge ritual takes `check.sh --merge`.
**Execution mode:** `pass-execute` by name, one invocation per segment, sequential, with `repo` set to the
worktree path, `implementer: "dubplate-implementer"`, `reviewer: "diff-reviewer"`,
`reviewerModel: "claude-opus-5-5"`, `maxFix: 1`, `stopOnEscalate: true`, `passClass: "tool"`, `gate` set to the
per-task string, and `reducedGate` set to the `--reduced` form. No `gateLane` is set, so every run takes the heavy
lane.
**Shared heavy lock:** every gate queues on the machine's heavy lock beside lane 1 and the cairn-cms session. The
conductor follows `~/.claude/docs/pass-gate-economy.md` ("Sessions sharing the heavy lock coordinate") and never
re-runs a full gate on an unchanged commit.

## Sources

Code and spec lines are cited at `869915e` on master, re-checked at `d2f8f5d`. Every commit since `0cb33ae`
touched only `docs/`, so the code pointers hold. Library behavior is quoted at the version this plan adopts.

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
- The 6d plan, revision 2.5 (`docs/superpowers/plans/2026-09-26-rung-6d-identity-serve-store.md`): "Where it
  runs" (`:432-435`), Task 0's sweep (`:951-957`), and "Contested surfaces" (`:1694-1708`), whose last sentence
  (`:1707-1708`) names this slice's `go.mod` addition as the one expected concurrent edit and sets when it may
  land.
- The navidrome-removal plan (`docs/superpowers/plans/2026-10-05-navidrome-removal.md`) for the gate and
  execution header lines.
- The three lens reviews of revision 1 and their probes (`docs/superpowers/research/2026-10-06-rung-12-slice-2-plan-review-contract.md`,
  `-mechanics.md`, `-risk.md`). The mechanics lens ran both pinned oracles live under podman; its measurements
  are cited as "the probe" below.

### What `internal/harness` already holds, and what this pass reuses

| Existing piece | At `0cb33ae` | Disposition |
|---|---|---|
| `internal/harness/cassette` | `cassette.go:1-24`: verbatim response files, never re-encoded, because `encoding/json` compacts and HTML-escapes a `json.RawMessage`; canonical path and query with sorted pairs, never decoded and re-encoded | The lesson is reused and the package is not. Its `Case` schema is the matcher's (beets items, an expected record; `cassette.go:41-57`), and `canonicalizePathQuery` (`:260`) is unexported and keeps every parameter, while a Subsonic session must drop the auth parameters. Task 4 states both rules in its own format. |
| `internal/harness/stub` | `stub.go:1-6`: replays a case over loopback and records misses | Not needed in 2a. The grader reads recorded responses from files. A later slice that wants a fake oracle over HTTP may adopt its shape. |
| `internal/harness/conform/arch_test.go` | `:16-65` and `:138-172`: a tree-walk invariant (the engine never imports the harness) with a proof that the check fires on a violating fixture | Task 2's boundary leg follows this pattern, including the fire-on-fixture proof. Task 7's no-link assertion joins this file, beside `TestEngineNeverImportsHarness`, with its own fire-on-fixture proof. |
| `internal/harness/oracle` and `make derive` | `oracle.go:1-10`, `Makefile:136-146`: the beets oracle runs outside the gate under the `oracle` build tag | The pattern is declined for this pass. The live container path ships as `cmd/harness` verbs, which the gate compiles, vets, and lints like `capture` (`cmd/harness/main.go:7-11`). A build tag would hide that code from `golangci-lint` and from `make derive`'s beets-only scope. `cmd/harness/freeze_derive_test.go` is `//go:build oracle` in package `main`, so the close vets that tag once (Task 9). |
| `internal/harness/analysis/ledger.go` | `:10-25`: strict JSON reads, unknown keys rejected | The rule is reused through `encoding/json/v2`'s `json.RejectUnknownMembers(true)`, as `internal/auth/roster.go:340` already does. |
| `cmd/harness` | `main.go:53-66`: the verb registration block | Tasks 3, 5, and 8 each add one verb. No other lane edits this block today. |
| `internal/engine/tagio` | `write.go:34-39`: `Write(path, ts, pictures, mode)` rewrites a FLAC's tags and pictures. `planBlocks` keeps an existing `VORBIS_COMMENT` vendor string and calls `freshVendor` only when the container has none (`write.go:122-128`); `freshVendor` stamps the main module's build version (`:239-245`) | Task 1 calls `Write` to tag the fixture library, from seeds that always carry a `VORBIS_COMMENT` block, so the built bytes never depend on the building binary. The import runs harness to engine, the allowed direction. |
| `internal/harness/oracle/testdata/template.flac` | 42 bytes, a header with no audio frames | Declined as a fixture seed. Later slices stream the fixture library, so its seeds carry real audio. |

### Library and stdlib behavior this plan relies on, quoted

| Source | Quote | Relied on by |
|---|---|---|
| `openapi3filter.ValidateResponse`, kin-openapi v0.149.0 | "ValidateResponse is used to validate the given input according to previous loaded OpenAPIv3 spec. If the input does not match the OpenAPIv3 spec, a non-nil error will be returned." | Task 7 |
| `openapi3filter.Options`, same (`options.go:22-26`) | "MultiError bool" (no doc line); "Set IncludeResponseStatus so ValidateResponse fails on response status not defined in OpenAPI spec" | Task 7: the checker sets both |
| `openapi3filter/validate_response.go`, same | Returns `nil` without validating for a `HEAD` request; for status 301, 304, 307, and 308 ("These status codes will never be validated"); when `responses.Len() == 0`; for an undeclared status with no `default` response unless `IncludeResponseStatus` is set ("By default, status that is not documented is allowed"), in which case it returns `&ResponseError{Input: input, Reason: "status is not supported"}` (`:54-57`); when the response has no content entry (`:99-102`); and when the media type's `Schema == nil` (`:114-117`) | Task 7: the checker resolves status, content, and schema itself and never reads `nil` as valid |
| `routers/legacy.NewRouter`, same | "If the given OpenAPIv3 document has servers, router will use them." | Task 7 |
| `openapi3/server.go:145-190`, same | A `{var}` segment in a server URL consumes input up to the next `/`. | Task 7: the pinned file's only server is `{url}`, so a full loopback URL leaves `//127.0.0.1:…/rest/…` unmatched. The probe confirmed that clearing `doc.Servers` and routing by path finds `/rest/getAlbum`. |
| `openapi3/schema.go` and `openapi3/errors.go`, same | A `oneOf` failure collects one error per failing branch, `validationErrors = multiErrorForOneOf{}` (`schema.go:2284`), appended in `OneOf` order as `validationErrors = append(validationErrors, err)` (`:2304`), skipping a branch a discriminator excludes (`:2294`); the `SchemaError` with `SchemaField: "oneOf"` then sets `e.Origin = fmt.Errorf("doesn't match schema due to: %w", validationErrors)` (`:2326`); and `func (meo multiErrorForOneOf) Unwrap() error { return MultiError(meo) }` (`errors.go:57-58`). With no discriminator and no branch matched, entry `i` is `OneOf[i]` | Task 7: the branch filter descends the `oneOf` error's `Origin` rather than flattening every leaf with `errors.As` |
| The pinned bundled OpenAPI file (`cb54c03c…ed46bfb6`, 453,720 bytes, fetched 2026-10-06) | `openapi` 3.0.0, `info.version` 1.16.1, 87 paths, all of the form `/rest/<endpoint>`, none with `.view`; each JSON `subsonic-response` operation's `200` response has only `application/json` content, and its schema is a success-or-failure union on `status`, which holds for all ten seed endpoints. Fourteen operations' `200` responses carry non-JSON content (`stream`, `download`, `getCoverArt`, and `hls.m3u8` among them), and `getTranscodeDecision` (POST) is not such a union (the verification read) | Tasks 4 and 7: the canonical endpoint drops a `.view` suffix before routing; the checker filters violations to the branch the body claims, falling back to every leaf when it reads no `status` |
| OpenSubsonic `content/en/docs/api-reference.md:20-25,55-59` at `bed1688` (Apache-2.0) | "`t` … The authentication token computed as md5(password + salt)"; "Use a salt length of at least six characters"; "Treat the strings as UTF-8 encoded"; the worked vector: password `sesame`, salt `c19b2d`, token `26719a1196d2a940705a59634eb18eab`; `f` takes "xml", "json", and "jsonp" | Task 2 |
| `encoding/json/v2.Unmarshal` | "a nil interface value of an empty interface type is initialized with a zero Go bool, string, float64, map[string]any, or []any" | Task 8: decoding into `any` turns every number into a `float64`, so the grader reads numbers through `encoding/json/jsontext` tokens and keeps the literal |
| `encoding/xml.Decoder.Token` | "Token expands self-closing elements such as <br> into separate start and end elements"; tokens are "properly nested and matched" | Task 8: `<a/>` and `<a></a>` compare equal |
| `go help list` | "-deps … iterate over not just the named packages but also all their dependencies"; "-test … report not only the named packages but also their test binaries"; a test variant's path carries a suffix | Task 2: a listed path is cut at its first space before the prefix check |
| `os/exec.Cmd.Start` (`$GOROOT/src/os/exec/exec.go:713-718`, go1.27.1) | `if c.ctx != nil { select { case <-c.ctx.Done(): return c.ctx.Err() …` | Task 3: a teardown command built on the cancelled run context never starts |
| `context.WithoutCancel` | "WithoutCancel returns a derived context that points to the parent context and is not canceled when parent is canceled." | Task 3: teardown runs under it, with its own timeout |
| Navidrome docs, `navidrome/website` `content/en/docs/usage/configuration/options.md` (Apache-2.0) | `ND_ENABLEINSIGHTSCOLLECTOR` default `true` (`:145`), `ND_AGENTS` default `"deezer,lastfm,listenbrainz"` (`:157`), `ND_DEEZER_ENABLED`, `ND_LASTFM_ENABLED`, `ND_LISTENBRAINZ_ENABLED` each default `true` (`:171,226,232`), `ND_SCANNER_SCANONSTARTUP` (`:258`); "Any configuration option can be set as an environment variable, just add … the prefix `ND_`" (`:125`) | Task 3: telemetry and every metadata agent off, so oracle responses never depend on the network |
| Navidrome docs, `content/en/docs/developers/dev-environment.md:39-41` (Apache-2.0) | "This option will always create an `admin` user with the specified password … `DevAutoCreateAdminPassword = "password"`" | Task 3: the first admin. The probe confirmed `ND_DEVAUTOCREATEADMINPASSWORD` creates `admin` and logs the password. |
| gonic `README.md` | "the default login is **admin**/**admin**" (`:31`); `GONIC_MUSIC_PATH` (`:54`), `GONIC_LISTEN_ADDR` (`:60`), `GONIC_SCAN_AT_START_ENABLED` (`:65`); `GONIC_MULTI_VALUE_ARTIST` and `GONIC_MULTI_VALUE_ALBUM_ARTIST` default `none`, "gonic will not attempt to do any multi value processing" (`:72-74,92`); "newer salt and token auth" (`:26`) | Task 3 |
| The probe (mechanics lens, 2026-10-06) | On this workstation (SELinux enforcing) a bare `:ro` bind mount gave both containers `open /music: permission denied`, and `:ro,z` scanned 3 of 3. gonic exits 1 unless `/playlists` exists; the image sets `GONIC_DB_PATH=/data/gonic.db`, `GONIC_PODCAST_PATH=/podcasts`, `GONIC_CACHE_PATH=/cache`, and `GONIC_PLAYLISTS_PATH=/playlists`. Both oracles answered token auth with `"status":"ok"`. An unreadable library still ends in "Scan completed" with zero tracks. Pulled by digest, `podman image inspect --format '{{.Digest}}'` returns that digest. | Task 3, Task 5 |
| `podman-run(1)`, `podman-port(1)`, `podman-rm(1)`, podman 5.8.7 man pages | `-p`: "Host port does not have to be specified (e.g. podman run -p 127.0.0.1::80). If it is not, the container port is randomly assigned a port on the host. Use podman port to see the actual mapping: podman port $CONTAINER $CONTAINERPORT."; `--label, -l=key=value`: "Add metadata to a container."; `--tmpfs=fs`: "Mount a temporary filesystem (tmpfs) mount into a container"; `podman port`: "look up the public-facing port that is NAT-ed to the private-port"; `podman rm --volumes, -v`: "Remove anonymous volumes associated with the container. This does not include named volumes created with podman volume create, or the --volume option of podman run and podman create."; `podman-ps(1)`: `podman ps --filter label=app=frontend` | Task 3: the loopback publish, the assigned port, the run label, gonic's tmpfs paths, and teardown's volume removal. The verification read proved the publish, port, label filter, and volume removal live. |

Task 0 re-quotes each row against the base commit and the adopted versions (the podman man page row against the
workstation's podman), and quotes `podman-run(1)`'s `z` option and `podman-image-inspect(1)`'s `Digest` and
`RepoDigests` fields. A changed quote is a note naming the new
text, and a removed behavior is a blocker.

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
| `josephburnett/jd` v2 (MIT) | Declined | A generic JSON diff cannot pair ids across servers or tell an order finding from a value finding (survey `:171`). |
| `delucks/go-subsonic`, `supersonic-app/go-subsonic` (GPL-3.0) | Declined | A test binary linking GPL code is a derivative (survey `:172`). |
| `mdlayher/gosubsonic` (MIT) | Declined | Last push 2017-11-19. Token auth is `md5(password + salt)` over `net/http` (survey `:173`). |
| Ampache `SubsonicSpecConformanceTest.php` (AGPL-3.0) | Method studied, never read for code | Its `KNOWN_DEVIATIONS` list is asserted to still fail, so it cannot rot (survey `:161`). Task 6's staleness rule is that idea, reimplemented. |
| Melodee test scripts (MIT) | Read | A curl walk with token auth (survey `:162`). Nothing to adopt in Go. |
| `mewkiz/flac`, `go.senan.xyz/taglib`, modernc sqlite, goose, Picard tables, chromaprint | Not consulted | No task here decodes audio, writes tags beyond `tagio`, or touches a database. |

**kin-openapi adoption evidence (2026-10-06).** License MIT (GitHub API `license.spdx_id`). Newest version
v0.149.0, 2026-08-28 (`proxy.golang.org/.../@latest`); twelve tagged releases from 2026-05-08 to 2026-08-28; last
push 2026-10-04; 103 open issues and PRs; 3,298 stars; not archived; 481 importers of `openapi3filter` (survey
`:166`). The survey's scratch module loaded the bundled file with no `doc.Validate` error and caught a missing
required array and a numeric `id`. The `openapi3filter` plus `routers/legacy` import set pulls six non-stdlib
modules: `go-openapi/jsonpointer`, `go-openapi/swag/jsonname`, `oasdiff/yaml`, `oasdiff/yaml3`,
`santhosh-tekuri/jsonschema/v6`, and `golang.org/x/text`. `go mod tidy` also adds `kr/pretty v0.3.1 // indirect`
through `oasdiff/yaml3`'s tests. `routers/gorillamux` is not imported, so `gorilla/mux` stays out of the build. The
module is v0.x, so a minor bump may break; the `dependency-upgrade` skill reads each bump's changelog as for any
other dependency.

**Indirect versions (Reading R11).** kin-openapi's minimums are `go-openapi/jsonpointer v0.22.5` and
`go-openapi/swag/jsonname v0.25.5`; `go list -m -u` shows `v1.0.2` and `v0.29.2`. Evidence on 2026-10-06:
`jsonpointer` v1.0.2 is the proxy's latest (2026-09-24), Apache-2.0, pushed 2026-10-01, 0 open issues, 53 stars,
not archived; `swag` (the `jsonname` submodule's repository) v0.29.2 is latest (2026-09-04), Apache-2.0, pushed
2026-10-01, 1 open issue, 222 stars, not archived. Both are `go-openapi` organization modules. With both bumped,
the probe's scratch module built and the validation results were identical. Task 0 adds each module's importer
count from pkg.go.dev to this record.

## Readings this plan takes

Each is settled here, with its reason. The review may contest any of them.

- **R1. The worktree serves this harness.** `CLAUDE.md:103-104` says "Corpus and harness work never moves to a
  worktree." The rule's recorded reason is the beets venv. The parallel-build approach's checklist item 5
  (`:255`) says `tools/venv/.venv/` is gitignored, so a worktree "has the README and the requirements and no
  venv", and "`make derive` … is deliberately outside `make check`, so the lane's gate goes green while the only
  leg exercising the beets oracle never ran and never complained. **Policy: corpus and harness work never moves to
  a worktree**." This harness uses no venv, no beets, no `corpus/`, and no `make derive`. The re-plan's ruling 5
  places this slice in its own worktree and keeps the main-tree fallback only "if the slice 2 plan finds that its
  harness needs the main tree" (`:48-54`); it does not. One residue of the reason applies:
  `cmd/harness/freeze_derive_test.go` is `//go:build oracle` in package `main`, Tasks 3, 5, and 8 add files to that
  package, and the gate never builds the tag. `go vet -tags oracle` compiles tagged code without the venv, so the
  close runs it once (Task 9). What a worktree lacks that this lane needs is `web/node_modules` (the gate's web
  leg), which the launch checklist provisions. The images live in the user's container store, and the Playwright
  browsers and the Unidecode wheel in the user cache, so every tree shares them. The mechanics lens's probe also
  confirmed that gitleaks' `.claude/worktrees/.*` allowlist hides nothing scanned from inside a worktree. The
  conductor queues a `CLAUDE.md` edit that scopes the rule to the matcher corpus and the beets oracle, naming the
  `go vet -tags oracle` check.
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
  date, sha256, the spec repo commit). The conformance package embeds nothing, accepts the document from its
  caller, and tests itself on a document its own `testdata/` holds, written for those tests. No file under
  `internal/opensubsonic/` reads a path outside that subtree, so the package extracts after 0.8 with its tests
  intact, and the module's own license question (spec `:947-948`, owed to slice 1) does not touch this pass.
- **R4. The client builds only what 2a calls.** It issues a request, signs it with the legacy token scheme, and
  returns the verbatim response with the envelope fields a 2a caller reads: `status`, `serverVersion`, and the
  error `code` and `message`. Typed per-endpoint wire types arrive with slice 3's handler, their first consumer.
  The `apiKey` scheme has no oracle (spec `:286-287`) and no dubplate endpoint before slice 6, so it waits for
  slice 6. Both pinned oracles accept token auth (the probe), so the plaintext `p=enc:` form is not built.
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
- **R9. 6d merges first, and this branch waits for it.** 6d is on the launch path (the spine: 6, 14, 7, 9), and
  this harness's first grading consumer is slice 3. 6d's own text lets this slice's `go.mod` addition land "before
  Task 0 or after 6d's merge" (6d `:1707-1708`), and kin-openapi cannot land ahead of the code that imports it,
  because `tidy-check` drops a requirement nothing imports (the probe confirmed it). So 2a builds alongside 6d in
  its worktree, holds its merge-ready branch until lane 1 closes, then rebases onto master and merges in a quiet
  `go.mod` window. 2a never commits in the main tree while the lane registry shows a lane 1 executor, which
  `CLAUDE.md:107-110` forbids. Task 0's sweep follows the same rule (Task 0). The later rebase is cheap: the
  branch adds new paths plus `go.mod`, `go.sum`, three `cmd/harness` verb files, one line each in
  `cmd/harness/main.go`, and one test in `arch_test.go`.
- **R10. Pins follow slice 1's first-pin-wins rule.** Slice 1's Task 0 pins Navidrome, gonic, and the
  OpenSubsonic specification commit, and adopts any pin this lane made first (slice 1 plan, settled item 5). Task 0
  here does the same from the other side. Slice 1 runs its own probe oracles and builds its own scratch fixture,
  so 2a hands it nothing and waits on it for nothing.
- **R11. Dependency picks take the newest production version.** `CLAUDE.md` ("Current dependencies") lands minors
  and patches without asking and has a rung plan that adopts a library name its newest version with license,
  maintenance, and adoption evidence (above). `go-openapi/jsonpointer` v0.22.5 to v1.0.2 is taken as non-major:
  under Go modules a major change is a `/vN` import path for N of 2 or more, v0 and v1 share one path, `go list
  -m -u` offers v1.0.2 as an ordinary upgrade, and semver's 0.y range carries no compatibility promise for v1.0.0
  to break. The probe built and validated identically with it. Task 0 re-reads `jsonpointer`'s v1.0.0 release
  notes; a stated break in an API kin-openapi calls is a held major with its unblock condition.
- **R12. The 2a proof grades the grader, never the oracles.** The committed ledger
  (`internal/harness/subsonic/data/ledger.json`) carries only rows with `dubplate` as one source, because the
  spec's oracle tier compares dubplate with an oracle (`:274-276`) and a row about two oracles would later match a
  dubplate comparison on the same path and hide a real divergence. In 2a it carries the spec's two declared
  classes (`:285-287`) as `owed` rows. The Navidrome-against-gonic grade of the seed session runs in a gate test
  against a test-only ledger under `grade/testdata/`, and its full report is a committed golden. Its unledgered
  findings stay in the golden as recorded evidence; they are never filed as ledger rows.

## Ruling for Geoff

**1. Does slice 2 run as two passes, 2a now and 2b after slice 1, under ceilings of 9.0M (flag 7.2M) and 4.0M
(flag 3.2M)?** Recommendation: **yes.** Reason: the slice is twelve build tasks, past the ten-task line, and
four of them (the XML rule, the seeded ledger rows with the per-endpoint oracle assignment, the encoding-quirk
cases, and the roster walk's content) consume slice 1 deliverables that do not yet exist. Planning them now would
plan against guesses, the plan-staleness defect this build has paid for. What yes builds: 2a merges on its own
once 6d has closed, its tests become the "test owed" targets slice 1's ledger names, and 2b's plan revision is
authored against slice 1's record. What no builds: one pass whose last segment stalls until slice 1 closes and
Geoff rules its forks, holding the worktree and the lane open meanwhile, with four tasks planned against guesses.
Revision 1 asked for 9.5M; revision 1.1 re-derived 2a from actual spend, and revision 1.2 corrects that sum to
9.0M (see "Ceiling").

## Global constraints

- **Where it runs.** Worktree `/var/home/glw907/Projects/dubplate/.claude/worktrees/lane-3-slice-2`, branch
  `lane-3-slice-2`, cut from master's tip at launch (`git worktree add` changes no main-tree file). Lane 1 (6d)
  keeps the main tree.
- **Territory (exclusive to this lane).** New: `internal/opensubsonic/**` and `internal/harness/subsonic/**`.
  Shared but uncontested: `cmd/harness/main.go`'s registration block (one line per verb) and three new verb files
  in `cmd/harness/`; `internal/harness/conform/arch_test.go` (one new test and its proof, Task 7). Read-only:
  `internal/engine/tagio`.
- **Contested with lane 1.** `go.mod` and `go.sum` only, for kin-openapi, under R9. This lane never edits another
  file on 6d's contested list (6d `:1694-1708`), `scripts/check.sh` among them, and it leaves `Makefile` alone too.
- **Contested with slice 1.** The oracle pins (R10) and, in 2b, the Android emulator install, which slice 1 makes
  and records and this lane reuses (slice 1 plan, settled item 9). Ports and `podman` containers never collide:
  each container publishes to a host port podman assigns, and its name and label are unique to the run. Queued
  edits that land on master after the merge, through the conductor: `CLAUDE.md` (R1's wording; the gate paragraph
  naming the new invariant tests and, at 2b, any new workstation prerequisite), `docs/STATUS.md`, and
  `docs/HISTORY.md`.
- **Clean room.** The protocol packages cite the OpenSubsonic specification or an observation recorded in a
  session, never GPL source (spec `:316-318`, `:341-347`). No implementer opens Navidrome's or gonic's source
  tree. Configuration comes from Navidrome's Apache-2.0 docs site and gonic's README.
- **The boundary.** Nothing under `internal/opensubsonic/` imports a module-internal path outside that subtree,
  tests included, and no file there reads a path outside it. Task 2 makes the import half a failing test.
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
- **Credentials.** No oracle credential is written to the repo, a session, a report, or a file the harness
  writes. Any container log line the harness prints passes through a redaction of the run's password.
- **Forward hooks.** No task lands a forward hook from spec `:767-855`. Every task is neutral; the harness reads
  no catalog, roster, or store surface.

## Lane launch (the conductor, before Task 1)

From parallel-build approach `:249-258` and `CLAUDE.md:99-110`:

1. **Reverse edges cleared.** 6d's territory (its "Contested surfaces") and this lane's are disjoint except
   `go.mod`/`go.sum`. Slice 1 is a document lane writing under `docs/` only, and it neither reads nor waits on
   anything 2a builds.
2. **Contested surfaces written** to STATUS's lane 3 row: `go.mod`/`go.sum`, with R9's rule (6d merges first).
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
- **The sweep.** If the lane registry shows a lane 1 executor, or 6d's Task 0 sweep has landed on master within
  the day, this pass records that sweep as its own in one line and runs none. Otherwise the conductor runs the
  sweep over `go.mod`, `web/package.json`, `tools/venv/requirements.txt`, and the pinned binaries in the main tree
  and lands it on master before cutting the worktree, or records one line when nothing moves. The last sweep is
  `docs/superpowers/research/2026-10-05-dependency-sweep.md`. Either way, the merge ritual (Task 9) re-reads the
  newest versions of kin-openapi and its indirects before the rebase.
- The worktree is cut from master's tip and the launch checklist is complete.
- Pins resolved and recorded in the task's report: kin-openapi's newest version (v0.149.0 on 2026-10-06; a newer
  release supersedes it, with the quotes table re-read), and its indirects at their newest (R11); the bundled
  OpenAPI build's sha256 and size on the day of pinning. For Navidrome and gonic, the first pin wins (R10): if
  slice 1's Task 0 has handed over its pins, this task adopts them; otherwise it pins each image as a `sha256`
  digest read from the registry and the conductor hands the list to slice 1's lane. The probe resolved both on
  2026-10-06: `docker.io/deluan/navidrome:0.64.2` (the tag carries no `v`; `ghcr.io/navidrome/navidrome:0.64.2`
  resolves to the same digest) at `sha256:38dc2727…a44ee52`, and `docker.io/sentriz/gonic:v0.22.0` (the tag
  carries the `v`) at `sha256:516fd964…be4c8b1d49`. Both are the newest releases (Navidrome v0.64.2 of
  2026-09-24; gonic v0.22.0 of 2026-06-08).
- Every `file:line` in this plan re-resolved against the worktree base. A moved pointer is a note; a pointer whose
  content is gone is a blocker.
- **The OpenAPI pin's timing against slice 1.** The spec has slice 1 re-read OpenSubsonic PRs #237, #162, #258,
  and #261 before the harness pins a spec build (`:951-955`); slice 1's Task 1 does that read. If its record has
  not reached this lane when Task 7 starts, Task 7 pins the build at the specification commit already pinned, and
  2b's Task 0 re-pins if the re-read moves anything. The pin is replaceable by design, and a re-pin re-runs the
  golden.

**Acceptance.** The report lists each pin with its evidence, each re-resolved pointer, and the added quotes
(`podman-run(1)`'s `z`, `podman-image-inspect(1)`'s `Digest` and `RepoDigests`). No gate.

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
- Committed audio seeds are short real encoded FLAC, each carrying a `VORBIS_COMMENT` block (the `flac` encoder
  writes one by default), so `tagio.Write` keeps the seed's vendor string and the built bytes never depend on the
  building binary. Image seeds are generated, never downloaded (for example `ffmpeg -f lavfi -i testsrc`). Each
  seed carries its sha256, and a README beside them records the command that produced each one. Committed fixture
  bytes stay under 512 KiB.
- A builder writes the library to a target directory by copying seeds and tagging each track through
  `tagio.Write`. Two builds into two directories are byte-identical, file for file. The builder also returns the
  library digest: a sha256 over the sorted list of each built file's relative path, size, and sha256. Task 5
  records it.

**States table.** Absent manifest: an error naming the path. Empty (no albums): refused, "manifest declares no
albums". Malformed or carrying an unknown member: an error with the JSON location. Many bad (several tracks with a
missing seed, a duplicate path, a path escaping the root, a seed whose sha256 differs, or a seed with no
`VORBIS_COMMENT` block): every one named in one error, and nothing written to the target.

**Test mandate and fixture state.** The determinism proof builds the committed manifest twice and compares every
file's bytes and the two library digests; it asserts the file count equals the manifest's track count plus its
image files, so an empty build cannot pass. A round-trip test reads every built track back through `tagio.Read`,
compares each manifest tag, and asserts each track's vendor string equals its seed's. A size test sums the
committed seed directory and fails at 512 KiB. **Mutations:** drop one tag write (the round trip fails); stamp the
build time into a file (determinism fails); skip the seed checksum (the bad-seed case passes and the test fails);
accept a seed with no `VORBIS_COMMENT` block (the refusal test fails).

### Task 2: The protocol client and the import boundary

**Files:** `internal/opensubsonic/client/` (new), the boundary test (below). **Hook:** neutral.

**Outcomes.**
- The client issues one request for a named endpoint with given parameters and a chosen format (`json` or
  `xml`), against a base URL that may carry a sub-path. It sends `u`, `t`, `s`, `v`, `c`, and `f`, with a fresh
  salt per request from `crypto/rand`, at least six characters (api-reference `:55-57`).
- It returns the HTTP status, the headers, the body bytes verbatim, and the parsed envelope for either format:
  `status`, `serverVersion`, and, on failure, the error `code` and `message` (R4). A failed envelope is a typed
  error the caller matches with `errors.As`.
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
  the fixture library (built by Task 1 into a temporary directory) mounted read-only with the shared SELinux label
  (`:ro,z`; the probe measured a bare `:ro` as unreadable on this workstation). Each container publishes its
  listen port to loopback with no fixed host port (`-p 127.0.0.1::<port>`), and the stack reads the assigned one
  with `podman port`. Containers carry a run-unique name and a per-run `--label`, so a later run can name orphans
  a killed process left.
- An image reference without a `@sha256:` digest is refused before podman runs. The digests live in one committed
  file the stack reads.
- Navidrome runs with insights and the Deezer, Last.fm, and ListenBrainz agents off and a scan on startup; its
  first admin is `admin` with a per-run password through `ND_DEVAUTOCREATEADMINPASSWORD` (quotes table). gonic
  runs with a scan at start, `GONIC_MULTI_VALUE_ARTIST=multi` and `GONIC_MULTI_VALUE_ALBUM_ARTIST=multi` (so the
  featured-artist track records gonic's behavior, never a configuration default), and a `--tmpfs` for each of
  `/data`, `/cache`, `/podcasts`, and `/playlists`. gonic keeps its documented `admin`/`admin` login, since no
  variable sets it; the container is per run and loopback-only. The credentials rule in "Global constraints"
  holds: log tails the harness prints redact the run's password.
- Readiness means a `ping` through the Task 2 client answers `ok`, and `getScanStatus` reports `scanning` false
  with `count` equal to the manifest's track count. Readiness has a deadline. A finished scan with any other count
  is a failure, because an unreadable library still ends in "Scan completed" (the probe).
- Teardown removes both containers and their anonymous volumes (`podman rm -f -v`) on every path: success,
  failure, and context cancellation. It runs under `context.WithoutCancel(ctx)` with its own timeout, because a
  command built on the cancelled context never starts (quotes table).
- `harness oracle-up` builds the fixture, starts the stack, prints each oracle's base URL and the run's
  credentials to standard output only (a roster walk types them into a client), and holds until interrupted,
  then tears down. Its callers are this task's live
  evidence run and every conductor roster walk from 2b on. It takes no base-path flag: slice 1 measures the
  sub-path base URL on its own probe oracles, and `ND_BASEURL` and `GONIC_PROXY_PREFIX` stay unset until a caller
  needs them.

**States table.** `podman` absent from `PATH`: an error naming it. A tag-only image reference: refused, naming the
image. Digest file absent, empty, or malformed: an error naming the file. Readiness past the deadline: an error
carrying each container's last log lines, redacted. A scan that never finishes: a timeout error naming the oracle.
A scan finished with a count other than the manifest's: an error naming the oracle and both counts. Both oracles
failing: both named in one error.

**Test mandate and fixture state.** Gate tests drive the stack against a fake container runtime (a helper process
re-executed from the test binary), proving the argv (the `:ro,z` mount, the loopback publish with no fixed host
port, the run label, the telemetry, agent, and multi-value variables, gonic's four tmpfs paths, the digest), the
refusals, the deadline, the count check, the redaction, and teardown on cancellation. The fake must observe a
removal call carrying `-v` for every container it started, including after the test cancels the run's context.
The live path runs once on the workstation as the task's evidence: the implementer runs `harness oracle-up`, pings
both oracles, and quotes the output (password redacted) in its report. **Mutations:** omit `:ro` on the mount;
drop `ND_DEEZER_ENABLED=false`; build teardown on the run context (the cancellation proof sees no removal); drop
`-v` from teardown; accept a tag-only reference; accept `count: 0` from a finished scan; print a log tail
unredacted. Each must fail a named test.

## Segment B

### Task 4: The session format

**Files:** `internal/harness/subsonic/session/` (new). **Hook:** neutral.

**Outcomes.**
- A session is one directory: a metadata file and one verbatim body file per (request, source) pair. Bodies are
  never re-encoded (`cassette.go:5-13`).
- The metadata records, per source, the library digest it served (Task 1), its image digest when it is an oracle,
  and its reported `serverVersion`, plus the resolution map (each symbolic key with the id it resolved to on each source), and the ordered requests. Each
  request names its endpoint in canonical form (the `.view` suffix dropped), its format, and its non-auth
  parameters sorted byte-wise, with entity references in symbolic form (a manifest key, never an oracle id). Each
  response names its source, HTTP status, `Content-Type`, and body file.
- No auth parameter (`u`, `p`, `t`, `s`, `apiKey`) is ever written to a session. The writer refuses a request that
  carries one, and the loader refuses a session that holds one.
- **A session's `serverVersion` has one channel.** The recorder sets it from the oracle's own envelope (Task 5).
  Because committed metadata can be hand-edited, the loader cross-checks each source's recorded `serverVersion`
  against every one of its response bodies' envelopes (the JSON `subsonic-response.serverVersion`, the XML
  attribute) and refuses a mismatch.
- A loader reads and validates a session; a writer writes one atomically.
- **A source joins a session through the writer.** The writer adds one source to a loaded session (its library
  digest, any image digest, `serverVersion`, resolutions, and its responses to the session's existing requests) and
  writes the result atomically to a directory its caller names. It refuses a source already present and a response
  to a request the session does not hold. Task 5's recorder adds each oracle this way. Slice 3 joins dubplate the
  same way: its test builds the library from the committed manifest, answers the committed seed session's
  requests, and adds dubplate as a source to a copy of that session, which Task 8 then grades against an oracle.

**States table.** Absent directory or metadata file: an error naming it. Empty request list: refused. Malformed or
unknown-member metadata: an error with the location. Many bad (missing body files, duplicate request keys, a
response for a source the metadata does not declare, a request carrying an auth parameter, a body file nothing
references, a source with an empty `serverVersion` or library digest, a `serverVersion` its bodies contradict, a
resolution-map key no request uses): every one named in one error. A joining source already present, or one
answering a request the session lacks: refused, naming the source and the request.

**Test mandate and fixture state.** Tests run on hand-built sessions in `testdata`, at least one valid session
with two sources and both formats, plus one per bad state. The version cross-check needs a session whose metadata
version differs from its bodies' envelope version. A round trip writes and reloads a session and compares bytes. A
join test adds a third source to the valid session, reloads the result, and asserts that source's library digest;
it also asserts both join refusals. **Mutations:** allow `t` through the writer (the auth refusal test fails); stop flagging orphan bodies; read
bodies through a JSON decode-and-marshal (the byte round trip fails); skip the version cross-check (the hand-edit
case loads and the test fails); let a joining source answer a request the session lacks (the join refusal test
fails).

### Task 5: `harness record-oracle` and the seed session

**Files:** `cmd/harness/recordoracle.go` (new verb), one line in `cmd/harness/main.go`, a session spec and the
seed session under `internal/harness/subsonic/data/sessions/seed/`. **Hook:** neutral.

**Outcomes.**
- The verb reads a session spec (the requests in order, with symbolic references), starts the stack, issues each
  request against each oracle through the client, and writes a session through Task 4's writer, adding each oracle
  as a source with the library digest from Task 1's builder.
- A symbolic reference resolves against responses already in the session. An album key resolves through
  `getArtists`, then `getArtist`, then a match on the manifest's album artist and title; a folder key resolves
  through `getIndexes` and `getMusicDirectory`. Each resolution lands in the session's resolution map. A reference
  that no earlier response reaches, or that matches more than one entity, refuses the run, naming the key and the
  oracle.
- The session's oracle digests come from `podman image inspect` on the running images, never from the digest
  file, and the recorder refuses unless the pinned digest appears in the image's `Digest` or `RepoDigests` (the
  fields Task 0 quotes). Its `serverVersion` comes from the oracle's own envelope, and an absent or empty one
  refuses the recording, naming the oracle.
- The seed session covers R8's ten endpoints in JSON and XML against both oracles, with `getArtist`, `getAlbum`,
  `getSong`, and `getMusicDirectory` requested for the same-title pair, the compilation, and a multi-disc album.
- The implementer runs the verb once live and commits the seed session.

**States table.** Spec absent, empty, or malformed: an error naming the file. An unresolvable or ambiguous
reference: refused before anything is written. One oracle failing mid-run: nothing written, the request and
oracle named. A spec request carrying an auth parameter: refused. A running image whose inspect output carries no
pinned digest: refused, naming the image and both digests. An absent or empty `serverVersion`: refused.

**Test mandate and fixture state.** The gate tests run the verb against a fake stack serving canned responses, so
resolution and refusal are proven offline. The ambiguity proof needs the same-title pair present in the canned
responses, and the unreachable proof needs a reference to an album absent from them. The digest proof needs a
fake inspect output whose `RepoDigests` holds the pin beside a different `Digest`, plus one holding no pin. The
version proof needs a fake envelope carrying a version string that appears nowhere else, and asserts the session
holds that exact string. The committed seed session must load through Task 4's loader with every request
answered by both oracles, and a seed test asserts both oracles report `duration` above zero for every song (the
offline evidence that the seeds decode). **Mutations:** resolve on title alone (the same-title pair resolves to
the wrong album and the test fails); take the digest from the digest file; accept an inspect output with no pin;
write a partial session when an oracle fails; accept an empty `serverVersion`.

### Task 6: The divergence ledger

**Files:** `internal/harness/subsonic/grade/` (new; the ledger half and the `Finding` type),
`internal/harness/subsonic/data/ledger.json`. **Hook:** neutral.

**Outcomes.**
- **The `Finding` type, defined here and consumed unchanged by Tasks 7 and 8.** A finding carries the request,
  the sources, a kind, a path, and both values. A `contract` finding carries one source, the one whose body the
  checker did not pass, and that body's value. The kinds are `value`, `type`, `presence`, `order`, `identity`
  (an id that pairs with two ids, or an entity that pairs with none), `missing-source`, `parse`, and `contract`.
- **The path grammar.** A path is `/`-separated segment names over the comparable tree Task 8 builds, where an
  XML attribute and a child element map to the same segment name as the JSON key. Each node keeps its XML form,
  attribute or element, and a form mismatch at one path is one `type` finding. A row's pattern is an endpoint
  (canonical form), a format (`json`, `xml`, or `any`), and a path in which `*` may stand only for one array-index
  segment. A pattern names one field path; no other wildcard exists.
- **Rows.** Each row carries every column of slice 1's Markdown ledger, so 2b's transcription is mechanical (slice
  1 plan, settled item 6 and Task 7): an id, the endpoint or area, the class, the sources it was adjudicated on,
  the oracle's behavior, dubplate's intended behavior, and an evidence pointer (a spec section or a slice 1
  record line). It adds what a machine needs: a state, `active` or `owed`. An `active` row carries a pattern, the
  session that exercises it, and one oracle version per non-dubplate source it was adjudicated against. An `owed` row names the slice
  whose plan builds its test, since a row about dubplate's own behavior has nothing to match until that slice
  serves a response. The report counts `owed` rows by slice.
- **A row matches only its own pair.** The grader applies a row only when its sources equal the graded pair. A
  gate test asserts every row in the committed `data/ledger.json` names `dubplate` as one source (R12).
- **Classes and what each suppresses.** `by-construction` (an oracle cannot grade this surface) suppresses every
  oracle-tier kind in its scope (`value`, `type`, `presence`, `order`, `identity`). `unoracled` (no oracle covers
  it) suppresses the same. `contract` (one source's response violates the OpenAPI file) names that source beside
  its pair and suppresses the `contract` kind at its path only for a finding from that source, so an oracle's
  `contract` row never masks the same violation in dubplate's body. `volatile` (the value varies per run or per server) suppresses `value` only, so type and
  presence are still compared. No class suppresses `parse` or `missing-source`.
- **The committed ledger** holds the spec's two declared classes (`:285-287`) as `owed` rows: Navidrome's folder
  view diverges by construction for `getIndexes`, `getMusicDirectory`, and `getMusicFolders` (owed to slice 3,
  the first slice with a dubplate side), and the `apiKey` path is unoracled (owed to slice 6). 2b adds slice 1's
  pre-seeded rows, and slice 3 adds the first `active` rows from its first dubplate session.
- **No active row can rot.** Grading a session fails on an `active` row that names the session and the graded
  pair and matches no finding in it, naming the row. The loader refuses an `active` row whose session directory is
  absent under `data/sessions/`. A gate test grades every session any committed `active` row names, so a new
  session joins staleness checking without an edit.
- **The oracle version is privileged.** Each of an `active` row's versions must be non-empty and equal the
  session's recorded `serverVersion` for its source. Only a ledger edit sets it, and the grader refuses a mismatch, so an
  oracle bump forces the deliberate re-adjudication pass the spec requires (`:279-280`). The session's version has
  its own single channel (Task 4).
- The report prints each row's match count, so an over-broad row shows on review.

**States table.** Absent ledger: an error (the ledger is required). Empty (`[]`): valid, and every finding then
reports unledgered. Malformed or unknown-member: an error with the location. Many bad (duplicate ids, unknown
class, an `active` row with an empty pattern, no session, an absent session directory, or an empty version, a
pattern using `*` outside an array-index segment, a `contract` row naming no source or one outside its pair, an
`owed` row naming no slice, a version mismatch, a stale row):
every one named in one error.

**Test mandate and fixture state.** Tests use fixture ledgers against a hand-built findings list holding at least
one match per class and one finding per kind. The staleness proof needs an `active` row that matches nothing in a
non-empty findings list, so an empty list cannot make it pass. The pair proof needs a row adjudicated on one pair
and a finding from another pair at the same path, plus one case where an oracle-source `contract` row leaves the
same-path `contract` finding from dubplate unledgered. The every-named-session test is proven on a fixture ledger
naming a fixture session directory, since the committed ledger holds no `active` row in 2a. **Mutations:** skip
the staleness check; compare versions by prefix; accept an empty version; let a `volatile` row suppress a type
change; let an `owed` row suppress a finding; match a row regardless of its pair; match a `contract` row on its
pair without its source; let `*` match a named segment.

## Segment C

### Task 7: The contract tier

**Files:** `internal/opensubsonic/conformance/` (new, with its own `testdata/`),
`internal/harness/subsonic/data/openapi/` (the pinned file, its `LICENSE`, and a provenance file), the pinned
document's loader and its real-body test in `internal/harness/subsonic/grade/`, `go.mod` and `go.sum`
(kin-openapi and its indirects at R11's versions), one test and its proof in
`internal/harness/conform/arch_test.go`. **Hook:** neutral.

**Outcomes.**
- The checker loads an OpenAPI document from bytes its caller supplies, clears its servers, and validates a
  response for a named endpoint, method, status, content type, and body. It routes by path (the quotes table
  explains why a full URL fails). It sets `MultiError` and `IncludeResponseStatus`.
- **It classifies before kin-openapi runs.** It first tests for a `HEAD` request and a 301, 304, 307, or 308
  status, as kin itself does before it reads the route (`validate_response.go:23-36`), and reports those as no schema. It then
  resolves the route, then the status (`Operation.Responses.Status(code)`, then `Default()`), then the content
  entry (`Content.Get(mime)`), then the schema, and returns one of five outcomes: valid; violations; no route for the endpoint; an undeclared status; or
  no schema (an empty content entry, a nil schema, a `HEAD` request, or a 301, 304, 307, or 308 status, each of
  which kin-openapi returns `nil` for without validating). `ValidateResponse` runs only when a schema exists. Only
  the first outcome is a pass, and no outcome is read from kin's error text.
- **Violations name the branch the body claims.** Every JSON `subsonic-response` operation's `200` schema in the
  pinned file is a success-or-failure union, so kin reports the failed branch's leaves too (the probe). A
  flattening `errors.As` walk loses which branch a leaf came from, so the checker descends each `SchemaError`
  whose `SchemaField` is `"oneOf"` through its `Origin` (quotes table), keeps the entry whose `OneOf` schema's
  `status` enum holds the body's `status` (every entry when the body has no readable `status`), and
  deduplicates by JSON pointer and reason. Each violation carries its JSON pointer and
  reason.
- The harness side loads the pinned file and verifies its sha256 before handing it to the checker. A mismatch
  fails, printing both sums.
- **The server never links the checker or the harness.** The new test in `arch_test.go` asserts that `go list
  -deps ./cmd/dubplate-server` names no `internal/harness` path, no `internal/opensubsonic/conformance`, and no
  `github.com/getkin/` path. Its matcher takes the listing as input, and a second test feeds it a canned listing
  carrying a `github.com/getkin/kin-openapi/...` path and an `internal/harness` path and asserts both are named,
  as `TestHarnessImportCheckFiresOnItsOwnFixture` does. The `go list` parsing from Task 2 is not shared across the
  boundary; this test carries its own.
- kin-openapi enters `go.mod` at the version Task 0 pinned, with its indirects at R11's versions, each named in
  the task's report.

**States table.** Document absent or failing its checksum: an error naming the file and both sums. A document
that fails to load: the loader's error. An empty body: a violation, never a pass. A body that is not JSON: a
violation naming the parse error. A wrong `Content-Type`: a violation naming it. Many violations: all of them,
branch-filtered and deduplicated.

**Test mandate and fixture state.** The conformance package's proofs run on a minimal OpenAPI document written for
these tests under `conformance/testdata/` (never copied from the pinned file): one path whose `200` response is a
success-or-failure union on `status` with a required array and string ids, one path whose declared status has an
empty content entry, one whose media type has no schema, and one declaring a 301. Its proofs: a valid body
passes; the required array removed fails at its pointer; a numeric `id` fails at its pointer; a body with two
sibling scalars numeric (`id` and `name` on one object) reports exactly those two pointers and no failure-branch
leaf; an endpoint absent from the document reports no route; an undeclared status reports an undeclared status;
the empty content, the nil schema, the 301, and a `HEAD` request each report no schema; an empty body, a
non-JSON body, and a wrong content type each report a violation. The harness-side test validates the seed
session's real `getAlbum` JSON bodies against the pinned file and asserts each oracle's exact violation list. The
probe measured Navidrome's as `/album/userRating` ("number must be at least 1", value `0`) and gonic's as
`/song/0/replayGain` ("Value is not nullable"); the task's report confirms or names the recorded lists, and the
repaired body (those violations removed) asserts valid. **Mutations:** drop `MultiError` (the two-sibling body
reports one and the test fails); unset `IncludeResponseStatus` and skip the pre-resolution (the undeclared status
reads valid); treat no schema as valid; drop the branch filter (a failure-branch leaf appears); route by full URL
(every proof reports no route); feed the no-link matcher a prefix typo (the canned violation passes and the test
fails).

### Task 8: The grader and `harness grade`

**Files:** `internal/harness/subsonic/grade/` (the differ half), `cmd/harness/grade.go` (new verb), one line in
`cmd/harness/main.go`, the seed grade's test ledger and golden report under `grade/testdata/`. **Hook:** neutral.
**Model:** opus.

**Outcomes.**
- The grader takes a session, a ledger, the pinned document, the committed fixture manifest (Task 1), and a pair
  of sources to compare, and produces one report holding two tiers. It refuses to compare two sources whose
  recorded library digests differ, naming both. The contract tier runs every JSON response of both sources through
  the checker; any outcome but valid is a `contract` finding carrying that response's source. The oracle tier diffs the two sources' responses for each request in the same format.
- Both formats decode into one comparable tree. JSON numbers keep their literal through `jsontext` tokens, and a
  string versus a number is a finding, since clients break on it. XML elements, attributes, and text map to the
  tree under Task 6's path grammar, and a self-closing element equals its expanded form. A JSON singleton where the
  other side has an array is a finding, never normalized away.
- **Ids pair by fixture identity, never by position.** Ids are opaque per server. The grader seeds the
  correspondence from the session's resolution map, then pairs every other id-bearing entity through its manifest
  identity read from the response (an artist by name, an album by album artist and title, a song by album, disc,
  and track number, or by album and title when the track carries no track number). An id that pairs with two ids, or an entity that matches no manifest identity or several, is an
  `identity` finding, never a positional guess. Ids paired inside a `by-construction` row's scope are used only
  inside that scope, so a folder view's divergent ids never leak into the album and song endpoints.
- **Order is its own assertion.** Arrays compare as sets of paired elements, except on a named table of
  order-significant paths kept in the grade package, each entry citing the OpenSubsonic response doc or a seed
  observation. The table starts with an album's song order. A reordered order-significant array is one `order`
  finding.
- Each finding carries the request, the path, the kind, both values, and the matching ledger row or
  "unledgered". The report prints each row's match count and is deterministic. The verb writes it to a file and
  exits nonzero on any unledgered finding or stale row.
- **The seed grade.** A gate test grades the committed seed session, Navidrome against gonic, under the test-only
  ledger `grade/testdata/seed-ledger.json`. That ledger holds the `by-construction` rows for Navidrome's folder
  view against gonic, the two `contract` rows Task 7 recorded (each naming the oracle whose body it excuses), and the `volatile` rows the pair needs, each with
  its reason. The test asserts: the report equals the committed golden `grade/testdata/seed-report.golden` byte
  for byte; a second grade yields the same bytes; a `by-construction` finding is present at a named path in the
  compilation's `getMusicDirectory` (the task's report names it); both `contract` findings are present at their
  paths; and every test-ledger row has a match count of at least one. Unledgered findings stay in the golden as
  recorded evidence of how the two oracles differ; they are never filed as ledger rows (R12). The golden
  regenerates only under the test's `-update` flag, and the task's report summarizes its findings by kind.

**States table.** A request with a response from only one source: a `missing-source` finding. An empty body on
either side: a finding. A body that fails to parse: a `parse` finding naming the format and the error. Sources
recorded over different library digests: refused. Many findings: all of them, sorted.

**Test mandate and fixture state.** A perturbation table drives the oracle-tier differ directly, without the
contract tier, on a same-source pair: one real seed-session body in each format, from one oracle, compared against
itself. The unperturbed pair asserts exactly zero findings. The JSON body must be a `getAlbum`, since the order
table's one entry is an album's song array, and that body fails the schema (Task 7), so a baseline through the full
grader cannot be zero. Each
row applies one change and asserts exactly one finding of the expected kind at the expected path: a changed
value, a missing field, an extra field, a string-for-number swap, a singleton-for-array swap, a reordered
order-significant array (one `order` finding and no `identity` finding), an id that pairs twice, and an XML
attribute moved to a child element (one `type` finding). `TestGradeIdenticalBodiesReportOnlyContract` runs the
full grader over a hand-built session whose two sources hold the seed session's real Navidrome `getAlbum` JSON body
byte for byte, and asserts zero oracle-tier findings and exactly the `contract` violations Task 7 recorded for that
body, once per source. The mismatched-digest proof grades a hand-built session whose two sources carry different
library digests and asserts the refusal names both. **Mutations:** decode numbers into `float64`; pair ids by position (the
reorder row reports a cascade); treat arrays as unordered everywhere; skip the contract tier when the oracle tier
is clean (`TestGradeIdenticalBodiesReportOnlyContract` fails; the seed grade cannot kill it, since the oracle
pair's oracle tier is never clean); drop the library-digest check (the mismatched-digest proof fails); leak an id pair out of a `by-construction` scope.

## Close (2a)

### Task 9: Reviews, merge, and the ledgers

The conductor runs these in order. Every reviewer dispatch asks for correctness gaps ranked by consequence, not
polish, and a fold may refuse a finding with a one-line reason.

1. One `code-simplifier:code-simplifier` dispatch over the pass's changed Go, pasting
   `.claude/instructions/simplifier-brief.md` verbatim, then one `diff-reviewer` read of its diff.
2. One fresh-context vacuity read (`claude-opus-5-5`) over the `grade` package's tests, the one package this plan
   specifies only by outcome. Its single question: which test could pass while the behavior it names is broken?
   One fold task if it finds any.
3. `go vet -tags oracle ./cmd/harness/... ./internal/harness/...` in the worktree, once, since the gate never
   builds that tag (R1).
4. The merge, after lane 1 closes (R9): the merge token; the newest versions of kin-openapi and its indirects
   re-read; `git rebase master` (resolve `go.mod` and `go.sum` by taking master's files and re-adding kin-openapi
   and R11's indirects at their newest, then `go mod tidy`); `check.sh --merge` in the worktree;
   `git merge --ff-only lane-3-slice-2` on master; the full gate on master; one `diff-reviewer` over the merged
   range if the rebase was not a fast-forward.
5. One `go-architecture-reader` per touched Go package, seven in all: `internal/opensubsonic/client`,
   `internal/opensubsonic/conformance`, `internal/harness/subsonic/fixture`, `.../stack`, `.../session`,
   `.../grade`, and `cmd/harness`. One fold task for accepted findings.
6. Queued repo edits in one batch on master: `CLAUDE.md` (R1's wording with the `go vet -tags oracle` check; the
   gate paragraph naming the boundary test and the no-link test), HISTORY's 2a entry (what landed, what the gate
   and reviewers caught, what slice 3 would be wrong to rediscover, the golden's finding counts by kind, the
   throughput baseline, both budgets scored), and STATUS (lane 3's row, 2b's wait on slice 1).
7. Worktree teardown, or the worktree held for 2b if slice 1 is close; the registry row says which.

**Acceptance.** Master is green with the merged work. The staged exported-unused leg reports no hit in the six new
packages. `go list -deps ./cmd/dubplate-server | grep -E 'internal/(harness|opensubsonic)|getkin'` prints
nothing. Every committed session loads through Task 4's loader, which refuses an auth parameter. `git grep -n
'\.\./' -- internal/opensubsonic` prints nothing. The oracle-tag vet is clean.

## Pass 2b (sketch; authored in full at its Task 0)

2b starts when slice 1's record lands. Its Task 0 re-reads that record and re-derives its ceiling, and its plan
revision replaces this sketch. It names these slice 1 deliverables as inputs, and each blocks only its own task.

| 2b task | Slice 1 input it waits on | Spec source |
|---|---|---|
| 10. The XML rule: validate XML per slice 1's choice (a legacy form against the official XSD, or a derived XSD), through `xmllint`, which the test names and fails without. The XSD carries no license text (survey `:272`), so it is fetched into the user cache and verified by sha256, never committed, unless slice 1's record settles its license. A new workstation prerequisite joins `CLAUDE.md`'s gate paragraph. | The XML validation rule | `:269-271`, `:939-940` |
| 11. The seeded ledger and the oracle assignment: slice 1's pre-seeded divergence rows transcribed into Task 6's machine form, through a Markdown-to-rows converter proven by a round trip on slice 1's actual table (moved from 2a, its first caller), each `active` where a recorded session exercises it and `owed` otherwise, and the per-endpoint oracle assignment (contract only, Navidrome, gonic, or the roster) as data the grader reads when one side is dubplate. The assignment's first non-test consumer is slice 3's grade, so 2b builds the data and its loader, and the verb's dubplate-side flag waits for slice 3. | The pre-seeded ledger; the per-endpoint oracle assignment | `:728-730` |
| 12. Encoding quirks: a session and a perturbation row for each quirk slice 1 enumerates (the attribute-heavy XML form, the wrapper, singleton versus array, JSONP). | The encoding-quirk list | `:883` |
| 13. The roster walk: a walk script per roster client (Feishin and Rufin on the desktop under the isolated configuration homes slice 1 sets up, Tempus in the emulator slice 1 installs), each step stating the observation to record, run once against `harness oracle-up` to prove the script, with the record committed under `docs/superpowers/research/`. Amperfy, which the workstation cannot drive, is listed for rung 16's attended gate (re-plan `:140-141`). | The flows each client exercises (slice 1's client-demands record) | `:732-734`; roster audit `:16-19` |
| Re-pin, if slice 1's PR re-read moved the OpenAPI build, and regenerate the seed golden. | The re-read of PRs #237, #162, #258, #261 | `:951-955` |

The search comparison (a set plus a top-N ordering assertion, spec `:287-289`) is built by slice 4's grader work,
its first consumer, on the order table and per-path comparison Task 8 leaves.

**One slice 1 owner fork reaches 2a's work.** Slice 1 brings "folder browsing at launch" to Geoff (its Task 7).
If he rules it out, gonic loses its folder-view role but keeps `getMusicFolders` (spec `:276-278`), the seed
session's `getIndexes` and `getMusicDirectory` rows become reference observations, and 2b's Task 0 retires the
`by-construction` ledger rows that grade them. 2a builds the spec's design as written, since gonic's place in the
stack is one image and one set of variables, and waiting on the ruling would stall segment A.

## Not verified at planning

- Whether `tagio.Write` writes a multi-valued artist tag in the shape Navidrome and gonic split. Task 1's round
  trip proves the write. Task 5's live recording, with gonic's multi-value settings on, shows how each oracle
  reads it.
- The exact paths of the seed session's `by-construction` and `contract` findings. The probe ran on a three-track
  scratch fixture; Task 5's recording and Task 7's real-body test name them.
- The conform suite's and the rendered leg's run time on every task gate. The first task's gate measures them.
  After 6d, the conductor files a `--scope-changed` refinement for `scripts/check.sh` that decouples the rendered
  scope from the conform scope, since nothing in this harness changes what the browser sees.
