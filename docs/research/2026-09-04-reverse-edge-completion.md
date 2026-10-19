# Reverse-edge completion: the fifteen inverted cells

Phase 0a of the ratified parallel-build approach. Written 2026-09-04 against the parallel-approach spec (revision 2), the rung dependency analysis, the build spec (revision 2.1), the rung 3 spec (revision 2), and the organizer spec.

## Method, and the reading of NONE

A cell (A,B) asks what rung A needs from, imposes on, or contends with rung B when A runs before or concurrently with B. The forward matrix asked only the ladder-order question, so a forward `—` is not evidence for the reverse cell.

NONE is granted only when the intersection of what A builds and what B builds is empty across five surfaces: Go packages, config keys, serve/mux surfaces, corpus and harness artifacts, and the box. **No cell earns NONE.** Fourteen are SOFT and one is HARD. A uniform wall of SOFT would be useless to a conductor, so the SOFT cells carry a band:

- **SOFT-A**: a real ownership or artifact edge with no written owner. Write the row before the inversion runs.
- **SOFT-B**: a scheduling exclusion the record already implies somewhere. Restated exactly so a lane launch can check it.
- **SOFT-C**: conservative on silence. The record does not answer the question; the cell is SOFT until it does.

## Four couplings that apply to every pair

Stated once here so no cell re-derives them.

1. **The gate couples every lane in both trees.** `scripts/check.sh` runs `make check` over `./...`, shellcheck over `scripts/*`, ruff over tracked Python, `gitleaks detect --no-git -s .` over the whole worktree, vale, `docker compose config`, and a pattern-filtered `systemd-analyze verify`. A worktree gets its own checkout of every tracked file, so a second code lane pays the full 192 MB gitleaks scan too. This is uniform and is not re-scored per cell.
2. **`.gitleaksignore` is one line.** It contains exactly `internal/config/keys_test.go:generic-api-key:66`. Any lane that needs an allowed finding appends to it, and rung 1b must re-fingerprint it. Two lanes appending concurrently produce a non-fast-forward rebase, which by the merge ritual triggers a `diff-reviewer` re-review of the merged result. Cheap, but it is the one file every lane can touch.
3. **The config wall is discharged by Phase 0b, and only the config wall.** `internal/config/keys.go` holds `ownedWall` as a slice literal today, which is exactly what rung 1b converts to a registry. Key collisions between any two rungs are therefore not scored per cell.
4. **`cmd/dubplate-server/main.go` has the config wall's problem and no registry.** `newRootCmd` calls `cmd.AddCommand(...)` once, at line 48, listing every subcommand. Rung 4 adds a shadow subcommand, rung 5 wires `serve`, and rungs 6, 9, and 13 add verbs. `doctor.go` has the same shape: `rowStatuses` is one function every rung extends. Phase 0b fixes the keys file and leaves this untouched. It is a general finding, and it bites hardest on (5,4), the only pair in phase 1 where both rungs add a subcommand.

## Summary

| Cell | Class | Basis |
|---|---|---|
| (5,2) | SOFT-B | No artifact edge either way. Shared gate cost and the one-line `.gitleaksignore`. |
| (5,3) | SOFT-B | No artifact edge. Lane A and lane T share one worktree slot for rung 3's whole four-pass window, and rung 5 is L band. |
| (5,4) | SOFT-A | `env/dubplate.env.template` is rung 4's by its own header, and `deploy.sh` hard-fails on any name with no minted value. Plus the `AddCommand` block and the state-directory layout. |
| (5,6) | SOFT-C | Rung 5 needs no store, by deliberate design. The verified-identity type that `internal/ops` carries has no written home, and the nonce and epoch state has no written migration story. |
| (5,7) | SOFT-A | Rung 5's login pages are the build's first rendered surface. The embedded static asset and font delivery path, and the one concrete theme variable block those pages need, have no owner. |
| (5,9) | SOFT-A | Three current documents disagree on who builds `internal/notify`. |
| (8,3) | SOFT-A | Rung 8's conformance target is `corpus/analysis/file_tags.json` and its fixture is `internal/harness/oracle/testdata/template.flac`, both inside territory rung 3 holds. The rung 3 spec asserts lane T writes neither. |
| (8,4) | SOFT-B | Rung 4's deploy is merge point M1 and the phase boundary. Rung 8's applied-field divergences have no assigned ledger. |
| (14,7) | SOFT-A | Phase 2 runs lanes V and I concurrently and both write `internal/serve`. The additive policy is stated over fields and verbs, and a journey enum value is neither. |
| (14,9) | **HARD** | Rung 14's wire types are a published projection of the ops job model, which rung 9's concurrency ruling shapes. The blocking artifact is the ruling. |
| (14,12) | SOFT-A | Rung 12 owns `internal/serve` from its opening with no concurrent serve-writing lane. Rung 14 must terminate the owned prefix so the later fallthrough cannot capture an unknown API path. |
| (14,13) | SOFT-C | Whether rung 13 exposes its queue verbs over `/api/v1`, and who writes those handlers, is unwritten. |
| (12,13) | SOFT-C | Same silence, read from the other side. Phase 3 declares rung 12 the only serve-writing lane, which holds only if rung 13 writes no serve. |
| (12,10) | SOFT-B | Box exclusion. Already written in the parallel spec; restated as a launch-checkable sentence. |
| (12,11) | SOFT-A | The post-scan hook event is rung 11's to name and rung 12's driver to fire. Inverted, rung 12 either steals the naming or ships without the firing. |

---

## The cells

### (5,2) rung 5 before or concurrent with rung 2 — SOFT-B

**What rung 5 builds.** `internal/auth`, the `internal/serve` shell and mux and ingress strip, `internal/notify`, Go-templated login and consumption pages, durable nonces and a per-email epoch in the state directory, the roster loader, and a large block of config keys registered through rung 1b's registry.

**What rung 2 built.** `corpus/`, `internal/harness/{cassette,stub,analysis,oracle,synth,mbrequest}`, `cmd/harness`, `tools/venv`, the Python oracle shims, and the MusicBrainz capture tool with its registered application identity.

**The intersection.** Package territory is empty. Config keys are empty and discharged by 0b regardless. Rung 5 spends no MusicBrainz budget, needs no venv, and reads no corpus artifact. Rung 2 needs no HTTP surface. The dependency analysis states the forward half plainly: "Rungs 2, 3, and 4 contribute nothing to auth."

Three files remain. `.gitleaksignore` is the sharpest, because rung 5 is the rung most likely in the whole build to need an allowed finding: it ships HMAC secrets, token serialization, and test fixtures full of secret-shaped strings. `Makefile` and `scripts/check.sh` are one-time contention that rung 2 has already taken and rung 5 does not add to. `docs/` is conductor-only and batched.

The residue is gate cost. Rung 2 is closed, so the live version of this cell is the standing adjudication stream, which commits dispositions into `corpus/analysis/*.json` from a document lane while lane A runs. Those are small commits into territory rung 5 never reads.

**Paste-ready row:**

| Surface | Owner | Consumers |
|---|---|---|
| `.gitleaksignore` | rung 1b re-fingerprints it; thereafter the lane that needs an allowed finding appends, and names the finding in its pass report | every lane; a concurrent append is a non-fast-forward rebase and triggers the merge ritual's step 5 |

### (5,3) rung 5 before or concurrent with rung 3 — SOFT-B

**The intersection.** Rung 3 builds `internal/engine/{matcher,musicbrainz}`, `internal/harness/conform`, and, across 3b through 3d, `corpus/`, `internal/harness/oracle`, `internal/harness/{cassette,stub}`, and `internal/harness/analysis`. Rung 3 computes no applied field, reads and writes no file, and has no HTTP surface: its conformance runs are offline, with no venv, no network, and no Python. Rung 5 touches none of that. The rung 3 spec's own claim holds on inspection: "rung 5's `internal/serve` work touches nothing rung 3 owns".

Both rungs are blocked on rung 1b for their config registration, which is a shared prerequisite rather than an edge between them. Rung 3's staleness pre-task already carries the check; rung 5's plan needs the same one.

**What survives is the slot.** The rung 3 allocation table puts rung 3 in the primary tree from 2d's close through 3d and gives the worktree to "lane A or T", one at a time. Rung 5 is L band and splits into several passes. If lane A holds the slot for the whole rung, lane T never runs in phase 1 and rung 8 slips past M1. Nothing in the record says the slot passes at pass boundaries rather than rung boundaries, and the merge ritual is per lane close, which a pass close satisfies.

**Paste-ready constraint sentence:**

> The second code slot is one worktree, held by lane A or lane T and never both. The slot passes at a pass close, not at a rung close: an L-band lane releases the worktree at each of its own pass merges, so a multi-pass rung cannot monopolize phase 1's second slot.

**One cross-cutting item this cell surfaces and does not resolve.** The gate lock is specified as "only one `check.sh` runs per tree at a time", which with a worktree permits two concurrent runs on one workstation. The claim that the lock "fixes `golangci-lint` cache contention" does not survive that, because the lint cache is per user and shared across worktrees. The record is silent. Any two-code-lane pair inherits this; it is named here once rather than in five cells.

### (5,4) rung 5 before or concurrent with rung 4 — SOFT-A

This is the highest-priority cell of the five rung 5 rows, and the record actively hides the edge.

**What rung 4 builds.** `internal/engine/shadow` and its subcommand, the divergence log in the state directory, the adjudication artifact, a hardened systemd unit with pre-created `ReadWritePaths`, its environment file, and the `deploy.sh` path rung 1 repointed. Rung 4 deploys to the box and starts the calendar clock.

**Three intersections, all real.**

First, `env/dubplate.env.template`. Its own header says "this manifest evolves at rung 4" and "Every NAME below must resolve to a non-empty value at render time or deploy.sh fails". `deploy.sh` line 191 confirms it: a missing value aborts the render, listing what to mint. Rung 5 introduces `DUBPLATE_MAGIC_LINK_SECRET`, `DUBPLATE_SESSION_SECRET`, `DUBPLATE_PUBLIC_URL`, and an SMTP transport. If lane A adds those names to the manifest before rung 16 mints them, rung 4's deploy fails, and rung 4's deploy is the merge point that ends phase 1. The build spec puts minting at rung 16, so the collision is between two ratified statements rather than an oversight by either.

The repo already carries the pattern that resolves it. `config.WallRow` has a `Pending` field, and `ownedWall` carries `{Name: "identity source", Pending: "pending (arrives at rung 5)"}` for precisely this reason: a row present in the table whose requirement has not arrived.

Second, `cmd/dubplate-server/main.go`'s single `AddCommand` block and `doctor.go`'s `rowStatuses`. Rung 4 adds a shadow subcommand and a shadow doctor row. Rung 5 wires `serve` and adds doctor rows for the trusted-proxy mode and the roster collision checks. Both edit the same two call sites, in different trees, concurrently.

Third, the state directory. Rung 4's divergence log is "durable in the state directory; the store arrives later". Rung 5's consumed nonces and per-email epoch are likewise durable in the state directory. Both phrasings are deliberate decouplings from rung 6, and neither names a layout owner. The systemd consequence is concrete: a `ReadWritePaths` entry must exist on the host before the unit starts, and `deploy.sh` pre-creates it, so rung 4's unit needs to know which subtrees exist.

**Paste-ready rows:**

| Surface | Owner | Consumers |
|---|---|---|
| `env/dubplate.env.template` | **rung 4** (the file's own header: "this manifest evolves at rung 4") | rung 5 adds no name it has not minted; a name whose value arrives later lands as a commented pending row on the `config.WallRow.Pending` precedent, never as a live `NAME=` line, because `deploy.sh` fails the whole render on one unresolved name. Rung 16 mints the auth secrets through the age-store flow. |
| `cmd/dubplate-server` root command and `doctor` row set | rung 1's skeleton; each rung adds its own file and registers through it | 4, 5, 6, 9, 13. The `AddCommand` block at `main.go:48` and `rowStatuses` in `doctor.go` are the `keys.go` failure mode one level up, and Phase 0b does not cover them. |
| The state directory layout | **rung 4** (first to deploy one, and its unit's `ReadWritePaths` depends on it) | rung 5's nonces and per-email epoch; rung 6's store supersedes neither |

### (5,6) rung 5 before or concurrent with rung 6 — SOFT-C

**Argued for NONE and refused.** The forward direction is a written SOFT, and the analysis answers the reverse half cleanly for the store: rung 5's durable nonces and per-email epoch live in the state directory, rung 4's divergence log likewise, and both phrasings are "deliberate decouplings from rung 6's SQLite store". Rung 5 genuinely needs no store, no migrations, and no ops vocabulary.

Two silences stop it short of NONE.

The verified-identity type has no written home. Rung 5 births the verify-identity seam and the verified identity carries its assertion channel through it, so the organizer's rule that admin verbs reject cookie-derived identity is enforceable at the seam. The analysis names "the identity type that ops request types may carry" as the second coupling. If that type lives in `internal/auth`, then `internal/ops` imports `internal/auth`, which puts a server surface underneath the vocabulary package that both implementations share. No document picks the package.

The nonce and epoch state has no migration story. Rung 6 ships SQLite with migrations that run forward on open. Whether rung 5's state-directory records move into the store when it arrives, and which rung pays for that, is unwritten. Answering it "no" is a legitimate outcome and would be worth writing down, because a later reader will otherwise assume the store absorbed them.

**Paste-ready constraint sentence:**

> Rung 5's spec names the package that owns the verified-identity type, and states whether the durable nonce and per-email epoch records stay in the state directory after rung 6's store lands. `internal/ops` may not import `internal/auth`.

The serve half is already governed and is restated here only so a lane launch can check it: "No two lanes write `internal/serve` until rung 5 closes." Rung 6's contract suite therefore runs `httptest` against its own handler until then, which is the cost the analysis already priced.

### (5,7) rung 5 before or concurrent with rung 7 — SOFT-A

**What the fold already discharged.** Phase 0c pulls the theme-variable contract, the two-axis match and palette model, and the embedded font delivery mechanism out of rung 7 into a standard, precisely so rung 5's login pages are authored against the contract every surface binds to. The contested table awards the roster loader to rung 5 and `internal/notify` to rung 5. Rung 7's ladder entry still says "roster loading with the birth-requirement validation", so the build spec text is unamended and the award lives only in the parallel spec's table; rung 7's plan must read the table rather than the ladder.

**What 0c does not discharge, and this is the cell's substance.** A documented contract names variables. It ships no variable values, no woff2 file, and no serving path. Rung 5's two pages are the first rendered user-path surface in the build, and they need three things nobody owns:

- The `go:embed` static asset path in `internal/serve`. The build spec's tree puts "embedded web" inside `internal/serve`, which rung 5 owns, so rung 5 births it and rung 7 later embeds a Svelte bundle through it.
- The woff2 font files themselves. 0c defines the mechanism; the estate's configured match is navidrome, whose typography is Roboto.
- One concrete theme variable block. Rung 7 owns the roster. Rung 5's pages render unstyled without at least one block, and rung 5 must not author a roster entry.

Rung 7's third theming tier, "an optional operator CSS file served after the bundle", and its second tier, "operator branding as config-rendered CSS variables at serve time", are both serve-side mechanisms in rung 5's shell.

One more clause worth writing while the notify row is open. Rung 5 births `internal/notify` to send magic links. Rung 7's reject emails and rung 9's notifications consume it. A notify seam designed as a magic-link mailer is the predictable defect. `_salvage/internal/notify` exists as the adoption candidate, under fresh review at the exemplary-Go bar.

**Paste-ready rows:**

| Surface | Owner | Consumers |
|---|---|---|
| The embedded static asset path (`go:embed` in `internal/serve`), the embedded woff2 fonts, and the config-rendered CSS variable injection point | **rung 5** (it births the serve shell and ships the first rendered pages) | rung 7 embeds the built `web/` bundle and the operator CSS file through the same path; rung 7 authors the roster |
| The concrete theme variable block rung 5's login pages render against | **Phase 0c's standard ships one default block**; rung 5 authors no roster entry | rung 7's roster supersedes it without touching rung 5's templates, because the pages bind to variable names only |

Amend the existing notify row's owner cell to read: rung 5, **as a general send seam rather than a magic-link mailer**, adopted from `_salvage/internal/notify` on merit.

### (5,9) rung 5 before or concurrent with rung 9 — SOFT-A

**The finding is a live contradiction, not a scheduling question.** Three current documents assign `internal/notify` differently:

- The build spec's ladder entry 9 (revision 2.1, ratified) gives rung 9 "the notifications and monitoring seams as optional integrations, built immediately ahead of the pipeline that consumes them".
- The dependency analysis's Wave B table gives lane P (rung 9) the territory `internal/engine/pipeline`, `internal/notify`.
- The parallel spec's contested-surface table (revision 2, the ownership authority) gives `internal/notify` to rung 5, "it must send mail to deliver magic links, and its deliverables already carry SMTP send", with rung 9 demoted to a consumer.

The contested table is the newest and is the document whose stated job is ownership, so it governs. The other two are errata. This is worth stating explicitly because rung 9's plan will otherwise be drafted against a ratified ladder entry that tells it to build a package another rung already shipped.

The organizer spec's rationale supports the table rather than the ladder. It put notification, roster, and monitoring together on one rung because "unpicking it later is strictly more expensive". Rung 5 building the send seam satisfies that reasoning, and leaves rung 9 with monitoring, the pings dead-man's-switch wiring, and the integration.

**The rest of the intersection is empty.** Rung 5 needs nothing from `internal/engine/pipeline`. It does not render journey states; the honest terminal state "imported, no player configured" is rung 9's addition to rung 7's vocabulary, reached through rung 12's driver-declared post-login destination. Rung 9's rate-limiter priority re-verification is MusicBrainz budget, which rung 5 never spends. Rung 9's concurrency ruling shapes the ops job model, which rung 5 does not consume.

**Paste-ready constraint sentence:**

> `internal/notify` is rung 5's, per the contested-surface table. The build spec's ladder entry 9 and the dependency analysis's Wave B territory column are errata on this point: rung 9 builds the monitoring seam and integrates the mail seam it does not own.

### (8,3) rung 8 before or concurrent with rung 3 — SOFT-A

**The forward SOFT is now discharged, and the reverse edge is a different one.** The forward cell 3 to 8 was soft on both rungs wanting one `AppliedFields` type, with the residual cost named as reconciling the corpus applied-tag record with the matcher's output type. The rung 3 spec closes that: "Rung 3 grades no applied field and computes none", and the type is pinned by rung 2's case format with rungs 8 and 9 as its consumers. There is no matcher output type to reconcile. The type axis is clear in both directions.

**What replaces it is territory, and the record makes a claim that does not hold.** The rung 3 spec's tree-allocation table concludes: "rung 3 holds the primary tree from 2d's close through 3d, and lanes A and T take the worktree for that whole window, one at a time. Neither writes `internal/harness`, `corpus/`, or `internal/engine/{matcher,musicbrainz}`."

The rung 8 handoff written into `corpus/README.md` at 2d's close says otherwise about lane T's reach. Rung 8's conformance target is `corpus/analysis/file_tags.json`, twelve subjects over the 69-key writable set. Its fixture is `internal/harness/oracle/testdata/template.flac`, the 42-byte silent untagged blob, and the proof shape is "writing these fields and reading them back the same way `derive_file_tags.py` did". So rung 8's grading instrument reaches into two paths rung 3 owns during 3b through 3d, and no document says where that instrument lives.

Read is safe. A worktree carries its own checkout of both paths, reads see committed state, and a rebase brings 3b's `internal/harness/oracle` additions in. Write is not safe, and there are two write cases the record does not rule out. If rung 8 needs additional derived subjects beyond the twelve, that is `make derive` under the venv, writing `corpus/analysis/file_tags.json`, which is main-tree work by the standing policy that corpus and harness work never moves to a worktree. And if rung 8's conformance runner is judged a harness artifact by the same reasoning that put rung 3's tier 0 decoder in `internal/harness/conform`, it lands in territory rung 3 holds.

One schema note. 3c moves the single `wantManifestSchemaVersion` constant if the `Case.Extra` change lands. Rung 8 reads `file_tags.json` and `corpus/manifest.json`'s `mediafile` pin rather than case bodies, so its exposure is the manifest rather than the cassette, but a lane T that pins a schema version of its own would collide.

**Paste-ready rows:**

| Surface | Owner | Consumers |
|---|---|---|
| Rung 8's applied-field conformance instrument | **rung 8, under `internal/engine/tagio` with its own `testdata`**, read-only against `corpus/analysis/file_tags.json` and `internal/harness/oracle/testdata/template.flac` | rung 15's re-run. Rung 8 pins no schema version of its own and never calls `Reconcile`. |
| Additional derived `file_tags.json` subjects | **main tree, venv-bound, `make derive`**; not lane T's to take from a worktree | if rung 8 needs them, they are scheduled ahead of rung 3 taking the primary tree, or as a named task in the rung 3 lane |

**Paste-ready constraint sentence:**

> Lane T runs beside rung 3 only while its plan states that its conformance instrument is read-only against `corpus/` and `internal/harness`. A rung 8 task that would derive a new corpus record moves to the primary tree.

### (8,4) rung 8 before or concurrent with rung 4 — SOFT-B

**Argued for NONE.** Package territory is disjoint: `internal/engine/tagio` against `internal/engine/shadow`, `systemd/`, `env/`, and `scripts/deploy.sh`. Rung 8 spends no MusicBrainz budget, touches the box not at all, and needs no venv at its committed scope. Rung 4 writes no tags: the shadow is read-only with respect to the library, it decides nothing, and the applied fields in its comparison come from live beets on the box rather than from a Go writer. Rung 3 computes no applied field, so the shadow's Go side has none either and the comparison is verdict-level. Both add subcommands, which is the cross-cutting `AddCommand` item rather than a pair-specific edge. That is a genuinely thin intersection.

**Two things stop it at SOFT.**

Rung 4's deploy is merge point M1 and the phase boundary. A lane open mid-pass at M1 has to close through the merge ritual or be carried into phase 2 with a written per-lane handoff, and the binary rung 4 deploys is built from whatever master holds. Rung 8 is pure library with no wiring until rung 9, so a landed but unwired `tagio` ships as unreachable code inside the deployed shadow binary. Harmless, and worth one sentence in rung 4's plan rather than a discovery at deploy time.

The adjudication artifact has no written home. The build spec says shadow divergences are adjudicated "into the same known-acceptable-divergence artifact the corpus uses", and rung 4's deliverables name "the divergence log ... and the adjudication artifact". The corpus's ledger is `corpus/analysis/adjudications.json`, whose proposed owner row says lane E1 authors and rung 3 writes only schema additions through re-derivation. Rung 4 has no row. Ruling B additionally requires every shadow divergence to be stamped with the matcher commit, which is a schema addition to that ledger. Separately, the record never says where rung 8's applied-field disagreements go when the resolution is an owner question; rung 3 built a tier 2 pending-row escape for exactly that situation, and nothing grants rung 8 one.

**Paste-ready rows:**

| Surface | Owner | Consumers |
|---|---|---|
| `corpus/analysis/adjudications.json` and the divergence schema | lane E1 authored it; **rung 4 adds the shadow divergence rows and the matcher-commit stamp** as a schema addition, never by hand-editing a derived record | rungs 3, 8, and 15 read it; `disposition` stays owner-only |

**Paste-ready constraint sentence:**

> Any lane open at M1 closes through the merge ritual before rung 4 deploys, or is carried into phase 2 with a written per-lane handoff. Rung 4's plan states that the deployed binary carries landed but unwired packages from other lanes, and that none is reachable from the shadow subcommand.

### (14,7) rung 14 before or concurrent with rung 7 — SOFT-A

**The inversion is the point of the re-sequencing**, so the question is what it costs. Three findings.

**Phase 2 runs two concurrent serve-writing lanes.** The parallel spec's serialization list reads: "No two lanes write `internal/serve` until rung 5 closes, and rung 12 owns it thereafter without a concurrent serve-writing lane." The exclusivity clause is scoped to rung 12's window. Phase 2 pairs lane V with lane I, and the contested row names "6, 7, 12, 14 add files at the declared extension point" as serve consumers. Rung 7 needs tus transport handlers and the embedded uploader served, so it writes serve. Rung 14 writes serve. Both at once, in phase 2, with only the extension-point convention between them.

That convention is not yet a mechanism. Rung 5's spec is required to "declare the extension point every later consumer adds a file at", and rung 5's spec does not exist. If the extension point turns out to be a shared registration list in one file, phase 2 has reproduced `keys.go` inside `internal/serve` with two Go lanes writing it concurrently. The requirement needs the same shape rung 1b gives the wall: registration from the consumer's own file.

**The additive-within-version policy does not cover enum values.** It is stated as "a published `/api/v1` verb keeps its shape, new fields and verbs may appear, and removals or semantic changes require `/api/v2`". The journey vocabulary is rung 7's, and rung 9 adds one state to it. A new value in a published enum field is neither a new field nor a new verb. The contract document has to rule on enum extension, and rung 14 is the rung that writes the contract document.

**The reject copy travels in payloads.** The organizer spec makes `internal/reject`'s `friendly_reason` strings part of the wire surface and the default rendering for every client. Rung 7 owns that contract, adopted verbatim from the ratified string authority under a register gate. Rung 14 publishes the payload shape that carries it, ahead of the rung that owns the strings. Copy revisions are stated as additive and never contract-breaking, so this is cheap, and it needs saying once so a copy change is never routed through a code review.

Verified sound, and named so a later round does not re-litigate: rung 14's authorization model already accommodates rung 7's contributor surface. Admin is an `AdminEmail` match with header-only assertions rejecting cookie-derived identity, which leaves contributor-scope verbs on cookie identity by construction.

**Paste-ready rows:**

| Surface | Owner | Consumers |
|---|---|---|
| The `internal/serve` extension point mechanism | **rung 5's spec declares it**, and it must let a consumer register from its own file rather than by editing a shared list | 6, 7, 12, 14. Phase 2 runs lanes V and I as two concurrent serve-writers, so a shared registration file is a merge conflict by design. |
| The journey vocabulary's wire projection | rung 7 owns the vocabulary; **rung 14's contract document rules whether enum extension is additive within a version** | rung 9 adds one state; rung 13 adds the queue states |

### (14,9) rung 14 before or concurrent with rung 9 — **HARD**

**The blocking artifact is rung 9's concurrency-model ruling**, not rung 9's code.

Rung 6 to 14 is already a hard forward edge: the wire types are a published projection of the ops interface, so the API cannot be authored before the interface exists. The job model is part of that interface "from its first version", and cancellation is an explicit verb rather than a reliance on `ctx` unwinding, settled before the interface exists "because retrofitting it changes the signature of the interface's most important operation". Rung 9's ruling picks between a per-album lease and an explicit queue with a documented starvation policy, and that choice shapes the job model those wire types project.

The parallel spec already recognized this and mitigated it without classifying it: "Rung 9's concurrency ruling is taken at phase 2's opening checkpoint, before lane V drafts its plan ... a ruling arriving after V lands leaves `/api/v1` taking a breaking change before it has a consumer or carrying a documented breach into launch."

That is a hard inbound edge with a scheduled discharge. Classifying it SOFT would let a conductor launch lane V on the strength of "the additive policy makes later verbs cheap", which is true of verbs and false of changes to existing types. The spec says so in the same paragraph.

**Earliest point the artifact exists:** phase 2's opening checkpoint. It is an owner touchpoint the build spec already names ahead of time, so honoring the edge costs no new attended moment.

**The residual is soft and stays soft.** Rung 9's repair verbs (query, retag against a named MusicBrainz release, move) are launch-bar admin surface that rung 14 must publish. They are new verbs, which the additive policy covers exactly.

**Paste-ready constraint sentence:**

> Lane V does not draft its plan before rung 9's concurrency-model ruling is taken. The ruling is scheduled at phase 2's opening checkpoint. This is a hard inbound edge on a ruling rather than on code, and the additive-within-version policy does not discharge it, because the policy is cheap for later verbs and not for later changes to existing types.

### (14,12) rung 14 before or concurrent with rung 12 — SOFT-A

**Argued and refused for NONE.** Rung 14's inbound is rungs 5 and 6, not 12: the session vocabulary and non-browser token acquisition sit on rung 5's seam, with the supersession list awarding the acquisition half to 14, and the ops interface is rung 6's. Rung 14 needs no driver, no proxy, and no Navidrome. Two things remain.

**Serve exclusivity.** Rung 12 owns `internal/serve` from its opening, and phase 3 states it is "the only serve-writing lane in the phase". Rung 14 writes serve. The proposed schedule separates them into phases 2 and 3, so the constraint costs nothing, and it should be written because it is the rule that survives a schedule change.

**Owned-prefix termination, which is the substantive finding.** The birth requirement is that dubplate's entire HTTP surface including the API lives under exactly one configured prefix, the streamer owns everything else, and "the auth boundary is decided by configuration, never by mux-miss". Rung 12 is the root mux's fallthrough rather than a mounted subtree. So the semantics of a mux miss change when rung 12 lands: before, an unmatched path is a 404 from dubplate; after, it is proxied to the streamer.

Rung 14 authors `/api/v1` before that fallthrough exists. Any route behavior that leans on 404-on-miss changes meaning later: an unknown verb under `/api/v1`, a trailing-slash redirect, a method-not-allowed path. An unknown path under the owned prefix must 404 from dubplate and must never reach the streamer, and rung 14 is the lane that has to build the prefix as terminating rather than falling through. Rung 5's spec owns the root-level composition order and names which point the driver's proxy occupies, so the requirement lands there and rung 14 implements against it.

**Paste-ready rows and sentence:**

| Surface | Owner | Consumers |
|---|---|---|
| The owned prefix's terminating behavior | **rung 5's spec declares the composition order and the fallthrough slot; rung 14 builds `/api/v1` as terminating within the prefix** | rung 12's proxy occupies the fallthrough and must be unreachable from any path under the owned prefix, including unknown API verbs |

> Lane V (rung 14) closes before lane X (rung 12) opens. Rung 12 owns `internal/serve` from its opening and no other serve-writing lane may be live beside it.

### (14,13) rung 14 before or concurrent with rung 13 — SOFT-C

**Argued for NONE.** Rung 13 builds persisted decision records, the queue verbs, the custody-policy enum, the admin review queue, and the appeal action on rung 7's push-back page. Rung 14 needs none of it: the decision-record payload serializes rung 3's candidate and penalty types (a hard 3 to 13 edge), and adding queue verbs to a published `/api/v1` is exactly what the additive policy makes cheap. Rung 13's `web/` reach contends with lane I, not with lane V.

The one thing that looks like an edge and is not: rung 14's authorization model has to accommodate the contributor resolution verbs, which are present only under the custody-policy settings that enable them. That is the ops operation-availability mechanism, which is rung 6's deliverable and a known hard 6 to 14 edge. It is not a 13 edge, and the organizer spec's rule that capability asymmetries are named in a spec rather than declared ad hoc in code is what keeps it that way.

**The silence that stops it at NONE.** Nobody has written whether rung 13 exposes its queue verbs over `/api/v1`. Its ladder entry names "the queue verbs (`list pending`, `answer`, `defer`) with a minimal CLI answering path", and the organizer spec says the minimal CLI exists "so the queue is usable before the remote TUI renders it", which reads as ops plus CLI with no HTTP at launch. Against that, launch-bar part one names the admin API as a completeness component, and the remote TUI (post-launch) consumes `ops/client` over HTTP. The contested serve row lists consumers 6, 7, 12, and 14, and omits 13.

If rung 13 writes no serve, this cell is NONE and (12,13) is NONE with it. That is one sentence in rung 13's spec, and it resolves two cells.

**Paste-ready constraint sentence:**

> Rung 13's spec states whether the queue verbs are exposed over `/api/v1` at launch. If they are, rung 13 joins the `internal/serve` consumer list, lane N inherits the serve exclusivity rule, and phase 3's claim that rung 12 is its only serve-writing lane needs amending. Until it is stated, treat rung 13 as a serve writer.

### (12,13) rung 12 before or concurrent with rung 13 — SOFT-C

**Argued for NONE.** Rung 12's driver seam covers scan trigger, scan status, album deep link, capability reporting, and a driver-declared post-login destination, plus the ingress reverse proxy and the Navidrome box configuration. Rung 13's queues live above the ops seam. The scan trigger is fired by rung 9, not 13. The appeal action is in `web/`, where rung 12 has no surface, because rung 12's user-facing UI is Navidrome's own. Rung 12's admin identity needs come from rung 5's provenance. On artifacts, the intersection is genuinely empty.

**The same silence.** Phase 3 proposes rung 12 alongside one of rungs 10, 11, or 13, and declares rung 12 the phase's only serve-writing lane. If rung 13 writes serve, pairing 12 with 13 breaks the phase's own rule, and 13 is one of only three candidates for that slot. The cell is therefore SOFT until rung 13's spec answers, and it collapses to NONE the moment it does.

**Paste-ready constraint sentence:**

> Until rung 13's spec states that its queue verbs are ops-and-CLI only, lane X (rung 12) pairs with rung 10 or rung 11 in phase 3, never with rung 13.

### (12,10) rung 12 before or concurrent with rung 10 — SOFT-B

**The record already answers this** and the answer needs restating as something a lane launch can check. The parallel spec's M12 erratum establishes that rung 12 writes the box, contrary to the build spec's box-safety section: it sets Navidrome's default UI theme to its built-in Gruvbox Dark via config, verifies the externalized-auth config names against the pinned Navidrome version (0.63.2), and carries provisioning choreography with a first-admin `doctor` preflight. Rung 10's first no-move diff needs read-only access to the live library listing, with R2 snapshots as the alternative source. The serialization list already says rung 12 "must not overlap rung 10's diff window".

**On artifacts the intersection is empty.** Rung 10 builds the migrator against the four failure modes, captures the legacy ledger format as a documented fixture while `_salvage/` is legible, and runs the corpus-driven path-parity fixture. Rung 12 needs none of it. Rung 12's Navidrome writes touch Navidrome's own database and configuration, never the library, so the exclusion is about a restart producing an inconsistent read during rung 10's listing rather than about corruption.

Worth noting for rung 10's plan rather than for this cell: rung 10 to 17 is the ladder's only latest-bound gate, and rung 12 does not move it.

**Paste-ready constraint sentence:**

> Rung 12 carries the quiet-box check. Lane X does not write the box while rung 10's read-only diff window is open, and rung 10 names its window's open and close in its pass report so the exclusion is checkable rather than assumed.

### (12,11) rung 12 before or concurrent with rung 11 — SOFT-A

**The forward cell states the edge from the other side.** "The post-scan hook event is named at rung 11 and fired by rung 12's driver. Cost: the event contract is defined once, wired later." That sentence presumes rung 11 runs first. Inverted, the presumption fails.

Rung 11 owns the lifecycle hook seam and its four named events: post-import, on-reject, on-review, and post-scan. Post-scan is the only one whose sole producer is rung 12's driver. If rung 12 runs first, it either names the event itself, which takes rung 11's ownership of the hook vocabulary, or it ships the driver without the firing and rung 11 wires it afterward. Both are defensible and they are different work, so the choice belongs in writing before either lane launches. Phase 3 makes the inversion plausible: rungs 10, 11, and 13 run two at a time and rung 12 takes a slot alongside one of them, so 12 running before 11 is a live configuration rather than a hypothetical.

**The rest of the intersection is empty, and one part is empty by an explicit rule.** Rung 11's built-ins write metadata through rung 8's tag writer and consume rung 6's per-track metadata storage decision, neither of which is rung 12's. Rung 11's probe-backed built-ins need chromaprint and ffmpeg, and rung 11 touches the box not at all. The capability question is settled by name: streamer driver capabilities belong to the streamer seam and ops operation availability belongs to the ops seam, and the organizer spec rules that "They never share a term." Rung 11 owns neither.

**Paste-ready row:**

| Surface | Owner | Consumers |
|---|---|---|
| The lifecycle hook event names and their JSON payload vocabulary | **rung 11** | rung 9 fires post-import, on-reject, and on-review; rung 12's driver fires post-scan. If rung 12 lands first it ships the driver without the post-scan firing, and rung 11 wires it; rung 12 does not name the event. |

---

## What the results permit and forbid in phase 1

Read against the tree as it stands: rung 2 closed at `e73b97e`, 215 corpus cases, `internal/harness/oracle` and `corpus/analysis/file_tags.json` committed, rung 3 taking the primary tree, and lane 2 recorded in `docs/STATUS.md` as blocked until Phase 0a, 0b, and 0c land.

**Can rung 5's lane run beside rung 3's passes? Yes, once four things are true.**

1. Phase 0a (this document), 0b (rung 1b's registry), and 0c (the design-language standard) have landed. All three are already blockers on rung 3's own staleness pre-task, and 0c is what rung 5's login pages are authored against.
2. Lane A takes the worktree and lane T does not hold it. The slot passes at a pass close rather than a rung close, so rung 5's L band does not consume phase 1's whole second slot.
3. The (5,4) rows are written before lane A launches, not before rung 4 deploys. Rung 4 follows immediately at the merge point, and the `env/dubplate.env.template` render is a hard failure rather than a warning.
4. Rung 5's spec carries what four cells push into it: the root-level composition order and the extension-point mechanism, the fallthrough slot the driver's proxy occupies, the package that owns the verified-identity type, and the disposition of the nonce and epoch state after rung 6.

Lane A writes `internal/auth`, `internal/serve`, `internal/notify`, its own config registration file, and its templates and embedded assets. It writes nothing under `internal/harness`, `corpus/`, or `internal/engine`. The split is clean in both directions, and no cell found an artifact rung 5 needs from rung 3 or rung 3 needs from rung 5.

**Can rung 8's lane run beside 3b through 3d? Conditionally, and 3a is the better pairing.**

Permitted when rung 8's plan states that its conformance instrument lives under `internal/engine/tagio` with its own `testdata`, and reads `corpus/analysis/file_tags.json` and `internal/harness/oracle/testdata/template.flac` without writing either. Reads from a worktree are of committed state and are safe.

Forbidden when rung 8 needs a new derived subject beyond the twelve `file_tags.json` carries. That is `make derive` under the venv writing `corpus/analysis/file_tags.json`, which is main-tree work under the standing policy, and lane T's worktree has no venv, because `tools/venv/.venv` is gitignored and `git worktree add` checks out tracked files only.

The pairing preference falls out of the rung 3 pass split rather than from a preference. Pass 3a writes `internal/engine/matcher` and `internal/harness/conform`, and touches neither `internal/harness/oracle` nor `corpus/`. Pass 3b writes both. So **lane T beside 3a and lane A beside 3b through 3d** is the allocation with the smallest collision surface, and it also fits the slot rule: rung 8 is M band and fits in one pass, while rung 5 is L band and wants the longer window.

**Forbidden in phase 1, in every configuration.**

- Lane A and lane T live at the same time. There is one worktree and one second code slot.
- Any lane writing `corpus/`, `internal/harness`, or `internal/engine/{matcher,musicbrainz}` from the worktree.
- Lane A adding a live `NAME=` row to `env/dubplate.env.template` for a secret that has not been minted, at any point before rung 16.
- A second executor in the primary tree. The live-executor check is the lane registry, never `pgrep -f Projects/dubplate`, which matches every lane once a worktree exists.

**Forbidden in phase 2, carried forward from these results.**

- Launching lane V before rung 9's concurrency-model ruling. This is the one HARD cell.
- Lanes V and I both writing a shared registration list inside `internal/serve`. They are both serve writers by phase 2's own table, and the extension point that keeps them apart is a rung 5 spec deliverable that does not exist yet.

**Forbidden in phase 3.**

- Pairing lane X (rung 12) with rung 13 until rung 13's spec states that its queue verbs are ops-and-CLI only.
- Rung 12 writing the box inside rung 10's read-only diff window.

**Two documents owe a correction as a result of this analysis.** The `internal/notify` ownership contradiction, where the build spec's ladder entry 9 and the dependency analysis's Wave B territory column both contradict the contested-surface table that governs. And the rung 3 spec's tree-allocation claim that lanes A and T write no `internal/harness` or `corpus/`, which is true of lane A and unproven of lane T, given where the rung 8 handoff puts rung 8's conformance target.
