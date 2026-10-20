# Rung 12, Slice 2b (the Oracle Harness, Second Pass) Implementation Plan

Status: **revision 1.3, approved by Geoff 2026-10-07 at revision 1.2; Ruling 1 answered yes (one pass at 9.0M).** Owner forks A,
B, and C from slice 1 are ruled (below). Every other decision is a reading below, taken on its merits and
reported.

History:

- Revision 1: drafted from slice 1's records and 2a's merged code. It replaces the "Pass 2b (sketch; authored in full
  at its Task 0)" section of the 2a plan, which now points here.
- Revision 1.1: folds the three lens reviews (contract C1 to C15 and O1, mechanics M1 to M18, risk DR1 to DR14) per
  `docs/research/2026-10-07-rung-12-slice-2b-plan-fold.md`. The assignment's values and every ledger
  row's slice are fixed here; the XML tier binds each body to its endpoint; the projection locates errors from the
  bytes it writes and drops residuals before xmllint runs; Navidrome's library request is the conductor's observation;
  Task 11 walks gonic only for the folder view; libxml2 is pinned in Homebrew.
- Revision 1.2: a second, narrow fold of the verification read
  (`docs/research/2026-10-07-rung-12-slice-2b-plan-fold-verification.md`, V1 to V9), recorded in the fold
  record's "Second fold". The conductor replays Navidrome's library request before segment C, and Task 9 gains a
  degraded form, so a failure before or during Task 9 never stops the run (V1). Task 5's three-source session carries
  one `serverVersion` per source (V2). Ruling 1's split delta and fallback claim are corrected (V5). Task 6's Files
  name the assignment loader (V7), and the conductor's step states its real trade-off (V6).
- Revision 1.3: lands the second fold's four owed items (V3, V4, V8, V9) before execution.

**Parent plan:** `docs/plans/2026-10-06-rung-12-slice-2-oracle-harness.md` (revision 1.2, "the 2a plan").
Its readings R1 to R12, its global constraints, and its clean-room rule stay in force for this pass; this plan adds
readings R13 to R28. Task numbers here are 2b's own, and commits name them as "rung 12 slice 2b Task N".

**Spec:** the 0.8 streamer spec (`docs/specs/2026-09-07-streamer-2-0-design.md`, revision 3.1, "the
spec"; it still says 2.0), ladder entry 2 at `:880-887`, "Evidence" at `:314-362`, under the re-plan
(`docs/specs/2026-10-05-rungs-12-16-replan-design.md`, revision 2.1, "the re-plan"). The sketch cited
revision 3's line numbers; revision 3.1 moved them, and every spec pointer below is revision 3.1's.

**What this pass is.** It finishes the evidence machinery 2a started, consuming slice 1's study record. It adds the
XML half of the contract tier, the divergence ledger's 54 rows in machine form, the per-endpoint oracle assignment,
the differ's encoding-quirk coverage, two more recorded oracle sessions (empty states and two music folders), a
two-collection fixture, and the first scripted roster walk. It also lands two routed chores: BACKLOG #27, the gate
scope that runs Chromium for harness-only changes, and BACKLOG #22, `harness capture`'s overwrite.

**What slice 3 consumes, and so what 2b must prove.** Slice 3 is the first slice with a dubplate side. It needs an
XML contract tier whose only pass is a real pass, a ledger whose every slice 1 row is present and owed to the right
slice, an assignment that never compares dubplate with an oracle the study ruled unfair, a differ that treats each
encoding quirk as the study classified it, and recorded oracle evidence for the `musicFolderId` filter. 2b proves
each directly:

- **The XML tier** on hand-written bodies that separate the correct projection from the prototype's name-keyed one,
  on bodies that bind the wrong payload to an endpoint, and on every seed XML body through a golden (Tasks 3 and 4).
- **The transcription** by a byte-exact round trip on slice 1's committed tables, an accounting test over all 66
  ids, and a recomputation of every row's slice (Task 6).
- **The assignment** by a value test that pins every entry against the study record's section 3 and refuses a
  `contract-only` value the record does not grant, and on a hand-built three-source session whose bodies differ
  wherever a comparison is claimed (Task 5).
- **The quirks** on a perturbation table over real seed bodies (Task 7).
- **The recordings** through the existing loader, grader, and goldens (Tasks 8 and 10).

**What this pass is not.** It touches no production surface: no wire type, handler, endpoint, or config key. It adds
no Go module dependency, so `go.mod` and `go.sum` stay untouched. It runs no container in the gate, and the gate opens
no network connection. Its delta to the set of values a default install's operator supplies by hand is zero (re-plan
`:137-138`).

**Pass class:** `tool` for Tasks 1 to 10, `docs` for Task 11. Everything is Go and data that ship nowhere, plus one
shell edit to the gate and one walk record. No task touches auth, sessions, the store, or `web/`. The class gives an
Opus `diff-reviewer` per task, `go-conventions` as the mandate, and a `go-architecture-reader` per touched package at
the close. As in 2a, every task names its mutations and the implementer reports each one failing.

## Rulings taken (Geoff, 2026-10-07)

Geoff accepted slice 1's three study forks with the study's recommendation on each (study record
`docs/research/2026-10-06-slice-1-study-record.md`, section 19), relayed by the conductor on 2026-10-07.

- **A. No `transcoding` extension at 0.8** (study record section 6). `getTranscodeDecision` and `getTranscodeStream`
  are not served at 0.8 and answer error 0. Effect here: the assignment marks both contract only and not served
  (Task 5); ledger rows L-34 and L-54 take the ruled text, and L-35 (AAC in a transcode decision, "only if fork A is
  yes") retires (Task 6). No 2b session records a transcoding endpoint.
- **B. The protocol module is MIT; its conformance checker loads the OpenAPI file as caller input and never vendors
  it** (study record section 10, which also answers the fork's NOTICE half: the OpenSubsonic repository carries no
  NOTICE file, so Apache-2.0 section 4(d) attaches nothing). Effect here: 2a's design already holds it, since
  `internal/opensubsonic/conformance` embeds nothing and the pinned document lives under harness data. 2b keeps the
  module that way. The XML tier lives in the harness, so the module gains no external binary and no XSD (R17). The
  sketch's Task 10 asked whether slice 1 settled the XSD's license; it did not (no license text), and this ruling's
  caller-input rule applies to the XSD too: it lives in the user cache, verified by sha256, never committed (Task 2).
  The NOTICE watch item is Task 0's pin re-read. The module's README line ("fetch the OpenAPI file yourself") is owed
  at extraction after 0.8, not here.
- **C. No folder browsing at launch** (study record section 12). gonic loses its folder-view role and keeps
  `getMusicFolders` (spec `:342-345`). The seed session's `getIndexes` and `getMusicDirectory` rows become reference
  observations: they stay in the seed and its golden, and no dubplate grade ever compares them with an oracle. Effect
  here: Task 5 retires the three `by-construction` rows in `data/ledger.json` (`folder-view-getindexes`,
  `folder-view-getmusicdirectory`, `folder-view-getmusicfolders`) and moves their job to the assignment, which marks
  `getIndexes` and `getMusicDirectory` contract only and `getMusicFolders` gonic only. Ledger row L-51 takes the
  ruled text, and L-52 ("only if fork C is yes") retires (Task 6). The test-only `by-construction` rows in
  `grade/testdata/seed-ledger.json` stay, since they grade the grader on an oracle pair (R12).

## Rulings for Geoff

1. **Run 2b as one eleven-task pass at 9.0M (flag 7.2M)?** Recommendation: **yes.** **Ruled yes by Geoff,
   2026-10-07.** The four segments already put a
   green full gate and a STATUS checkpoint every three tasks, which bounds an unattended run without a second close.
   The live tasks (9 to 11) share the worktree and the `grade` package with the rest. The two whose evidence rests on
   Navidrome's unquoted library API now carry written degraded forms the conductor takes alone (Task 9's
   first-collection Navidrome, Task 10's gonic-only session), triggered before or during Task 9, so they no longer
   carry an unbounded tail.
   - **Yes builds:** this plan as written: segments A to D, one close (Task 12), one merge, one set of architecture
     reads.
   - **No builds:** 2b ends after Task 8 with its own close and merge (about 6.5M at the same rates: Task 0, eight
     build tasks, the conditional Task R, the close); Tasks 9 to 11 become a pass 2c with its own plan revision, review,
     ceiling (about 3.3M: three build tasks, Task 11's 0.3M premium, and a second close at about 1.3M), and merge.
     Estimate against estimate, the two sum to about 9.8M against this pass's 8.55M: about 1.3M more, which is the
     second close. Each would also carry its own fix-round headroom above its estimate, as 9.0M does above 8.55M.
     Slice 3 waits for 2c, since it consumes the two-collection evidence.

## The scope, re-derived against the sketch

The sketch named four tasks and a re-pin. The records show three of them were the wrong size, and they route six
items the sketch did not list.

| Sketch item | What the records show | 2b task |
|---|---|---|
| 10. The XML rule | Settled as the legacy form with a scoped projection (study record section 5; contract record Q3). The prototype keys the scope on the element name, and the XSD declares one name with several types (`album` is `AlbumID3` under `albumList2` at `xsd:393` and `Child` under `albumList` at `xsd:387`), so a faithful port fails exactly the bodies it should pass. xmllint reports a one-line body's errors all on line 1 (observed on the seed, below). Larger than one task. | 2 (the pinned XSD), 3 (projection and runner), 4 (the tier in the grader) |
| 11. Seeded ledger and oracle assignment | The ledger's ten classes are not 2a's four machine classes, its "Oracle" column names up to two oracles where a machine row holds one pair, and 12 of its 66 rows are agreements, not divergences. Under R12 no row can be `active` before slice 3 serves a dubplate response, so every row lands `owed`. Larger than one task. | 5 (assignment), 6 (transcription) |
| 12. Encoding quirks | Most of E1 to E12 are dubplate's own by-design answers (JSONP declined, error 0 for an unknown endpoint, `v` required), graded against the spec and never an oracle, so recording the oracles' answers buys slice 3 nothing. What the differ itself must handle is E1, E5, E6, E10, L-07, and L-08. | 7 |
| 13. The roster walk | Slice 1 already walked Feishin and Rufin live and recorded a method that works (base-URL record, "Isolation and the profile hashes"). Tempus still has no emulator: `ls /dev/kvm` fails and `/proc/cpuinfo` shows no `vmx` (checked 2026-10-07). | 11, with Tempus deferred |
| Re-pin and golden | On 2026-10-07 `main` is still `bed1688`, none of PRs #162, #237, #258, #261, #278 has merged, the bundle's sha256 is unchanged (`cb54c03c...`), and no NOTICE file exists. | Task 0 re-checks; Task R runs only if something moved |
| Not in the sketch | 2a's carry: lyrics, `nowPlaying`, and radio requests (HISTORY, 2a, "Routed items"). | 8 |
| Not in the sketch | A two-collection fixture (spec ladder entry 2, `:885-886`; study record "Routed"). | 9, 10 |
| Not in the sketch | The area list (server-behavior record section 10.3) and the agreement rows (ledger A-01 to A-12). | 5, 6 |
| Not in the sketch | The credential-echo scan (ledger L-07) and the message rule (L-08). | 7 |
| Not in the sketch | BACKLOG #27 and #22 (HISTORY, 2a, "Routed items"). | 1 |
| Not in the sketch | The known-deviation list "with `lastScan` and the oracle nulls" (spec ladder entry 2, `:885`; [record] "Routed"). 2a's `contract` class with its stale-row check is the mechanism; Task 8's test ledger carries gonic's nulls as `contract` rows; L-20's `lastScan` is owed to slice 3 (Task 6). | 6, 8 |

**One record conflict, resolved here.** The study record's section 4 routes "the two-collection fixture that proves
the filter" to slice 3, while its own "Routed" list and the spec's ladder entry 2 (`:885-886`, revision 3.1, the later
and governing text) route it to slice 2. This plan follows the spec (R22).

## Ceiling, checkpoints, segments, and execution

**Ceiling (2b):** 9.0M execution tokens, flag 7.2M (80%). The sketch's 4.0M priced four tasks; the re-derived scope
is eleven. Derivation from 2a's actual spend (HISTORY, 2a, "Budgets"): about 6.0M in all, of which Task 0 took about
0.15M and the close (simplifier and its review, the vacuity fold, seven architecture reads and their fold, the merge,
the ledgers) about 1.2M, leaving about 4.64M over eight build tasks with one fix round among them: 0.58M a task, live
and Opus tasks included. 2b prices eleven build tasks at that rate (6.4M), a 0.3M premium on Task 11 (driving two GUI
clients through screenshots), Task 0 at 0.25M (it adds the PR re-read), a conditional Task R at 0.3M, and a close at
1.3M (six or seven architecture reads, the simplifier, one vacuity read, folds, and the merge). The sum is
6.4 + 0.3 + 0.25 + 0.3 + 1.3 = 8.55M; 9.0M leaves one more fix round. Planning and plan review are recorded outside it.
Revision 1.1 keeps the figure: its folds add proofs to Tasks 3, 5, and 6 and one conductor observation before Task 9,
and they remove Task 3's dedupe rule and half of Task 11's GUI walks.
**Lane sub-ceiling:** lane 3's sub-ceiling is this ceiling.
**Checkpoint interval:** every four tasks at most, at each segment boundary, and at the close. STATUS's lane 3 row is
written at each.
**Segments:** Task 0 (conductor), then Task R if Task 0 calls for it. A: Tasks 1, 2, 3. B: Tasks 4, 5, 6. C: Tasks
7, 8, 9. D: Tasks 10 and 11. Then the close, Task 12. Every boundary lands on a green full gate.
**Dependencies.** Task 4 needs Tasks 2 and 3. Task 6 needs Task 5. Task 8 needs Tasks 4 and 7. Task 10 needs Task 9.
Tasks 1, 2, 5, 9, and 11 depend on nothing earlier in this pass but share one tree and the `grade` package with their
neighbors, so every task runs in order in the one worktree; no task is marked for a parallel chain.

**Executor:** `dubplate-implementer`, `model: sonnet`, effort `high`, for every task but Task 3. Task 3 runs on
`model: opus` from the start: the projection's position-keyed type resolution and its error location are novel
correctness-critical logic this plan specifies only by outcome. A second `fix` verdict on Task 9 or Task 10 is the
upshift the conductor weighs first, since a live stack's failure mode is a recording that looks complete and is not.
**Per-task gate:** `cairn-run-gate 'bash /var/home/glw907/Projects/dubplate/.claude/worktrees/lane-3-slice-2/scripts/check.sh --scope-changed'`,
re-issued on exit 75 until it prints `gate exit:`. Task 1 changes the gate itself, so its own gate is the full form
on the heavy lane, and its task entry carries both `gate` (the full form) and `gateLane: "heavy"`, because
`pass-execute`'s `tool` class defaults the lane to light (`pass-execute.js`: `PASS_CLASSES.tool.gateLane` is
`"light"`, and the lane resolves as `t.gateLane || a.gateLane || cls.gateLane`). From Task 2 on, a task whose change
set reaches none of the rendered leg's inputs launches no browser (Task 1's outcome), so it runs with `gateLane:
"light"` (`~/.claude/docs/pass-gate-economy.md`: "A gate that launches no browser takes the light lane"). The
mechanics review measured the fit: `go test -count=1 ./...` under the light lane's scope peaked at 605 MB and a cold
`golangci-lint run ./...` at 1.30 GB, under the 3G cap. If a light gate is still killed at the cap, the conductor
moves the rest of that segment to the heavy lane and records it; that is never a task failure. Task 11 takes
`check.sh --reduced`. Segment boundaries and the pass end take the full form (`check.sh`, heavy lane), and the merge
ritual takes `check.sh --merge`. A boundary reuses the last task's full gate when `HEAD` has not moved since it ran.
**Execution mode:** `pass-execute` by name, one invocation per segment, sequential, with `repo` set to the worktree
path, `implementer: "dubplate-implementer"`, `reviewer: "diff-reviewer"`, `reviewerModel: "claude-opus-5-5"`,
`maxFix: 1`, `stopOnEscalate: true`, `passClass: "tool"` (`"docs"` for Task 11), `gate` set to the per-task string,
`reducedGate` set to the `--reduced` form, and `gateLane` set per task as above.
**Shared heavy lock:** every heavy gate queues beside any other session's, under `pass-gate-economy.md`'s
coordination rules.

### The unattended run (Geoff, 2026-10-07: a pass should handle a 10-hour unattended run)

The conductor runs segments A to D and the close without an owner touchpoint, holding the sleep inhibitor and the
runaway guard per `~/.claude/docs/unattended-work-guards.md`.

**The hard ceiling.** At the flag (7.2M) the conductor lets the in-flight task's chain finish, writes STATUS with the
resume prompt, finishes the current segment, and starts no new segment. At 9.0M it stops after the in-flight chain
returns, writes STATUS, and stops. STATUS is also written at each segment boundary and before any stop.

**The conductor rules alone, and records each ruling in its notes and later in HISTORY:**
- accept, one re-dispatch on `fix`, an upshift to Opus, or a split of a task whose second `fix` names a seam;
- the content of any golden, since a golden records oracle evidence and grades the grader (R12);
- a live oracle behavior that contradicts a plan expectation, when the remedy stays inside the task's files and
  touches no ruling (as 2a's errata E1 and E2 did); the remedy is recorded as an erratum at the plan's end;
- a request an oracle answers outside the envelope, which the recorder cannot hold: it drops from the session spec
  with its reason (Task 8, Task 10);
- recording Navidrome's second library as unmeasured if the conductor's step before segment C cannot observe or
  replay the creation request without reading Navidrome's source, or if Task 9's live evidence cannot be met on
  Navidrome with two libraries; Task 9 and Task 10 then take their degraded forms, which are written below and need no
  further ruling;
- moving a segment to the heavy lane, and reordering segments C and D if a live task is blocked.

**These stop the run (STATUS written, then stop):**
- the hard ceiling;
- a new executor in the main tree editing `scripts/check.sh`, `Makefile`, or `cmd/harness/main.go` (the contested
  surfaces), or any warm change in this worktree the conductor did not dispatch;
- a pinned input that no longer verifies: the XSD's sha256 at its URL, the pinned oracle digests, or xmllint's
  version on `PATH`;
- a merged OpenSubsonic PR or a NOTICE file upstream that changes a ruling's premise (Task 0 or Task R);
- any hash of a real client profile or the host keyring that changes across Task 11's walk (the isolation failed);
- a gate red on a commit whose diff does not explain it, after one re-run;
- an `escalate` verdict the conductor cannot resolve inside the task's files.

**Owner-gated steps, batched at the end in one STATUS question:** enabling VT-x (it unblocks Tempus's walk, Task 11),
and any finding that would reopen fork A, B, or C. Nothing deploys, and the box is never contacted.

## Sources

Code and spec lines are cited at `c4b7afd` on master, where this worktree is level. Library and tool behavior is
quoted at the version this plan uses, and the commands below were run on the workstation on 2026-10-07.

- Slice 1's records under `docs/research/`, cited by the study record's aliases: the study record
  (`2026-10-06-slice-1-study-record.md`, "[record]"), the divergence ledger ("[ledger]"), and the contract
  ("[contract]"), client-demands ("[clients]"), base-URL ("[base-url]"), server-behavior ("[servers]"), evidence-pins
  ("[pins]"), and prior-art-cells ("[cells]") records. Also `2026-10-05-client-roster-audit.md`.
- The spec: "Evidence" (`:314-362`), "The protocol layer is a separable module" (`:371-408`), "License"
  (`:421-427`), ladder entry 2 (`:880-887`), the forward hooks (`:930-1025`), "Prior art" (`:1026-1064`).
- The re-plan: the rung 12 ladder text and the roster walk (`:113-141`); the operator delta (`:137-138`).
- 2a's merged code under `internal/harness/subsonic/`, `internal/opensubsonic/`, and `cmd/harness/`; HISTORY's 2a
  entry (`docs/HISTORY.md:5-142`); BACKLOG #22 and #27 (`BACKLOG.md:85-89`, `:104-107`).
- The dependency sweep of 2026-10-07 (`docs/research/2026-10-07-dependency-sweep.md`).

### What 2a built that this pass extends

| Piece | At `c4b7afd` | 2b's use |
|---|---|---|
| `grade/ledger.go` | `Row` carries `id`, `area`, `class` (four machine classes), `sources` (one or two), `oracle`, `dubplate`, `evidence`, `state`, `source`, `slice`, `pattern`, `session`, `versions`; `problems` requires a known `class` on every row | Task 6 adds the record's class and test columns and relaxes `class` and `sources` for `owed` rows |
| `grade/grade.go` | `contractTier` checks JSON only (`req.Format != client.JSON` skips XML); `Grade(ctx, dir, pair)`; `Config` names one manifest | Task 4 adds XML; Task 5 adds the assignment; Task 10 lets `Config` name several manifests |
| `grade/tree.go` | XML decode: attributes and children become members by local name, a name repeated among one element's children becomes an array, namespace declarations are dropped (`:197-271`) | Task 4 maps an element ordinal to a path under this same rule |
| `grade/differ.go`, `differ_test.go` | `TestDifferPerturbations` drives the differ alone on real seed bodies | Task 7 extends the table |
| `session` | The writer and loader refuse a credential in request parameters | Task 7 adds the body-side guard at the recorder |
| `cmd/harness/recordoracle.go` | Symbolic `artist:`, `album:`, `song:`, `folder:` keys (`:37`) | Task 10 adds `collection:` |
| `stack` | One library mounted at `/music`, readiness on the track count | Task 9 mounts several |
| `cmd/harness/fetchunidecode.go`, `Makefile:148-156` | A cache-filling verb and its make target, verified by sha256, a no-op when the cache verifies | Task 2 follows the same shape |
| `data/ledger.json` | Four `owed` rows: three folder-view `by-construction` rows and `apikey-unoracled` | Task 5 retires three; Task 6 replaces the fourth with L-10 |

### Library, tool, and stdlib behavior this plan relies on, quoted

| Source | Quote or observation | Relied on by |
|---|---|---|
| The Subsonic 1.16.1 XSD at `https://subsonic.org/pages/inc/api/schema/subsonic-rest-api-1.16.1.xsd` | Fetched 2026-10-07: HTTP 200, 32,775 bytes, sha256 `9baee72fa24087190ad22be4f7674d628acc7a3e0fb65160d74115b166df0387`, the pin [contract] recorded. `targetNamespace="http://subsonic.org/restapi"`, `elementFormDefault="qualified"`, `attributeFormDefault="unqualified"`; the root element `subsonic-response` has type `sub:Response`, whose payload is one `<xs:choice minOccurs="0" maxOccurs="1">` (`xsd:9-56`, the choice at `:12`), so the schema accepts an `ok` envelope with no payload, with any payload in the choice, or with `error`, whatever the endpoint. No `xs:any` and no `xs:anyAttribute` | Tasks 2, 3 (the endpoint binding exists because of `minOccurs="0"`) |
| The same XSD, element types by position | `album` is `sub:AlbumWithSongsID3` in `Response` (`:19`), `sub:Child` in `albumList` (`:387`) and `searchResult2` (`:311`), and `sub:AlbumID3` in `albumList2` (`:393`) and `searchResult3` (`:319`); `entry` is `sub:NowPlayingEntry` in `nowPlaying` (`:284`) and `sub:Child` in playlists (`:350`, `:367`) | Task 3: types resolve by position |
| `xmllint` (libxml2 2.15.4, `/home/linuxbrew/.linuxbrew/bin/xmllint`, first on `PATH`) | `xmllint --version` writes `xmllint: using libxml version 21504` to **stderr** and nothing to stdout, exit 0 (observed 2026-10-07). `/usr/bin/xmllint` is libxml2 2.12.10 (`21210`). Man page (`--schema SCHEMA`): "Use a W3C XML Schema file named SCHEMA for validation"; `--noout`: "Suppress output". Observed on a seed body: each error is one stderr line `<file>:<line>: Schemas validity error : Element '{http://subsonic.org/restapi}artist', attribute 'sortName': The attribute 'sortName' is not allowed.`, then the trailer `<file> fails to validate`; a valid file prints `<file> validates`. Every schema error is reported at its element's start-tag line, attribute and missing-child errors included. After an unexpected child (`Element '{...}roles': This element is not expected. Expected is ( {...}album )`) no later sibling of that parent is reported, nor anything inside one. Navidrome's seed XML bodies are one line; 30 of gonic's 31 are indented over 2 to 60 lines | Task 3 |
| libxml2 `v2.15.4` `doc/xmllint.xml`, DIAGNOSTICS | Exit 0 "No error", 1 "Unclassified", 2 "Error in DTD", 3 "Validation error", 4 "Document isn't well-formed or couldn't be read", 5 "Error in schema compilation", 6 "Error writing output". `--nonet`: "Do not use the Internet to fetch DTDs or entities. Has no effect with libxml2 2.15 or later." The 2.15.0 NEWS: "The built-in HTTP client and support for LZMA compression were removed." | Task 3: only 0 and 3 are validation outcomes; the pinned build cannot open a connection, so the gate's no-network rule holds without a flag |
| Homebrew on this workstation | `uupd.timer` runs daily with its Brew module enabled (journal: `"msg":"Brew module" ... "Enabled":true`), and that module runs `brew update` then `brew upgrade -y` (`ublue-os/uupd` `drv/brew/brew.go:50`, `:63`, on `main`, 2026-10-07). `brew info libxml2`: "Installed (as dependency)"; the Brewfile has no `libxml2` line. `brew pin --help`: "Pin the specified package, preventing it from being upgraded when issuing the brew upgrade formula or cask command. See also unpin. Note: Other packages which depend on newer versions of a pinned formula might not install or run correctly." | Lane launch: libxml2 is pinned (R18) |
| `encoding/xml.Decoder.RawToken` (go1.27.1) | "RawToken is like Decoder.Token but does not verify that start and end elements match and does not translate name space prefixes to their corresponding URLs." | Task 3 |
| `encoding/xml.Decoder.Token` | "Each of the Name structures contained in the Token has the Space set to the URL identifying its name space when known." "Token expands self-closing elements such as <br> into separate start and end elements" | Tasks 3, 4 |
| `encoding/xml.Encoder.Indent` | "Indent sets the encoder to generate XML in which each element begins on a new indented line" | Task 3. One start tag per line is not "line k is element k": end tags take lines, whitespace character data passed through adds lines, and text can hold a newline (next two rows) |
| `encoding/xml` encoder source (go1.27.1) | A start tag with a non-empty `Name.Space` is written as `xmlns="<Space>"` (`marshal.go:744-748`), so a `RawToken` prefix such as `sub` becomes `xmlns="sub"`; character data goes through `escapeText(p, t, false)` (`marshal.go:224-225`), and `escapeText` escapes a newline only "If escapeNewline is true" (`xml.go:1922`) | Task 3: the line table counts the newlines actually written, and a prefixed body has its own state |
| `encoding/xml.Encoder.EncodeToken` | "Callers that create an Encoder and then invoke EncodeToken directly, without using Encode or EncodeElement, need to call Flush when finished to ensure that the XML is written to the underlying writer." | Task 3 |
| `os.UserCacheDir` | "On Unix systems, it returns $XDG_CACHE_HOME ... if non-empty, else $HOME/.cache." "If the location cannot be determined (for example, $HOME is not defined) or the path in $XDG_CACHE_HOME is relative, then it will return an error." | Task 2 |
| `os/exec.LookPath` | "LookPath searches for an executable named file in the directories named by the PATH environment variable"; "on success the result is an absolute path" | Task 3: the version refusal names the binary's path |
| `os/exec.ExitError.Stderr` | "If the error output is long, Stderr may contain only a prefix and suffix of the output, with the middle replaced with text about the number of omitted bytes. Stderr is provided for debugging, for inclusion in error messages. Users with other needs should redirect Cmd.Stderr as needed." The mechanics review measured the cut: 1,000 errors gave 65,566 stderr bytes and 434 of them survived | Task 3: the runner sets `Cmd.Stderr` to its own buffer and never parses `ExitError.Stderr` |
| `net/url.ParseQuery`, `URL.Query` | `ParseQuery` "always returns a non-nil map containing all the valid query parameters found; err describes the first decoding error encountered, if any", and "Settings containing a non-URL-encoded semicolon are considered invalid". `URL.Query` "silently discards malformed value pairs. To check errors use ParseQuery." | Task 7: the scan reads `ParseQuery`'s map and treats its error as a finding |
| The pinned OpenAPI bundle, one path probed (`/rest/getAlbum`) | The 200 response's only content is `application/json`, schema `GetAlbumResponse`, whose `subsonic-response` is a `oneOf` of `GetAlbumSuccessResponse` and `SubsonicFailureResponse`; `GetAlbumSuccessResponse` is an `allOf` of `SubsonicSuccessResponse` and an object with property `album` and `required: ["album"]` | Task 3: the endpoint binding reads each path's payload member this way |
| The seed session | All 62 recorded XML responses carry `application/xml` | Task 3's content-type check |
| `os.ProcessState.ExitCode` | "ExitCode returns the exit code of the exited process, or -1 if the process hasn't exited or was terminated by a signal." | Task 3: only 0 and 3 are validation outcomes |
| kin-openapi v0.149.0 `openapi3` | `Loader.LoadFromData`: "LoadFromData loads a spec from a byte array"; `Schema` carries `AllOf SchemaRefs` and `Properties Schemas`; `SchemaRef` carries `Ref string` and `Value *Schema`; `Components.Schemas`; `Paths.Map`: "Note: iteration on Go maps is not ordered." | Tasks 3, 5 |
| The pinned OpenAPI bundle (`cb54c03c...`) | 195 component schemas, among them `Child` (54 properties), `AlbumID3` (29), `ArtistID3` (10), `Playlist` (13), and `AlbumID3WithSongs`, `ArtistWithAlbumsID3`, `PlaylistWithSongs` as `allOf` compositions; 87 paths, each `/rest/<endpoint>` | Tasks 3, 5 |
| gonic `README.md` at `v0.22.0` | "gonic supports multiple music folders" (`:104`); `export GONIC_MUSIC_PATH="my albums->/path/to/albums,my compilations->/path/to/compilations" # with aliases` (`:122`) | Task 9 |
| `navidrome/website` `content/en/docs/usage/features/multi-library.md` at `b0fe572` (Apache-2.0) | "When Navidrome starts, it automatically creates a default library using your `MusicFolder` configuration. This becomes "Library 1"" (`:26`); "Admin users automatically have access to all libraries" (`:30`); "Creating Additional Libraries" through the admin UI's "+" button with a name and a path (`:46-57`); "The new library will automatically begin scanning" (`:62`); "Ensure the path doesn't overlap with other libraries" (`:150`) | Task 9. The docs name only the UI; [servers] F2 created a library "through its admin API" by observation and recorded no method or path |
| `fixture.Load` (`fixture/manifest.go:67-83`) | "The seeds are read from a seeds directory beside it" (`filepath.Join(filepath.Dir(path), "seeds")`); a manifest's `tags` is a `tagio.TagSet`, which writes `genres` (`internal/engine/tagio/encode.go:123`); the seed manifest sets no `genres` on any track | Task 9 |
| [base-url] "Isolation and the profile hashes" | Feishin and Rufin ran under `flatpak run --command=sh` with `HOME` and `XDG_*` exported inside the sandbox, keyring talk names denied, Feishin's `--password-store=basic`, Rufin on a private Xvfb display; the real Rufin profile hashed identically before and after | Task 11 |

Task 0 re-quotes each row against the base commit and the versions in use. A changed quote is a note naming the new
text, and a removed behavior is a blocker.

## Prior art consulted (CLAUDE.md, build-for-0.8)

| Entry | Verdict here | Reason |
|---|---|---|
| OpenSubsonic specification and OpenAPI bundle (Apache-2.0) | Adopt, as 2a pinned them | The contract. Task 0 re-reads the PRs and the NOTICE watch. |
| Subsonic 1.16.1 XSD (no license text) | Adopt as a fetched input in the user cache, pinned by sha256, never committed | Ruling B's caller-input rule and [contract] Q3, "The XSD's license". |
| libxml2 `xmllint` 2.15.4 (MIT) | Adopt as a workstation tool, pinned by version | Evidence below. |
| `jacoelho/xsd` v0.6.3 (MIT, pure Go) | Declined | Three stars, one maintainer, created 2026-01-06, v0.x with three releases in the week before 2026-10-04. No adoption evidence beside libxml2's. Revisit when the module extraction after 0.8 wants a pure-Go XML tier. |
| `lestrrat-go/libxml2`, `terminalstatic/go-xsd-validate` | Declined | cgo bindings to the same libxml2: they add a C toolchain and headers to the gate for a binary the workstation already runs. |
| `moolekkari/validatexml-go` | Declined | No SPDX license (`NOASSERTION`), last push 2025-11-24. |
| Ampache `SubsonicSpecConformanceTest.php` (AGPL-3.0) | Method only, as in 2a | Validate only what an authority covers and record the rest as a declared, self-expiring gap ([contract] Q3). Task 4's declared-unvalidated count is that idea. |
| Slice 1's projection prototype ([contract] appendix, this lane's own) | Method adopted, with one correction | Its per-type scoping is right; its name-keyed scope is wrong for `albumList`, `searchResult2`, and `nowPlaying` (R19). |
| Navidrome and gonic (GPL-3.0) | Oracles, behavior only, at 2a's digests | Their docs (Apache-2.0 site, gonic's README) are read for configuration; no source tree is opened. |
| Feishin, Rufin, Tempus, Amperfy (GPL-3.0) | Roster clients, run and read at the [pins] versions | Never copied. |
| `mewkiz/flac`, `go.senan.xyz/taglib`, modernc sqlite, goose, Picard tables, chromaprint, Melodee | Not consulted | No task decodes audio, writes tags beyond `tagio`, touches a database, or reads intake. |

**xmllint adoption evidence (2026-10-07).** libxml2 is the GNOME project's XML library, MIT-licensed. Its newest tag
is `v2.15.4` (`gh api repos/GNOME/libxml2/tags`), and Homebrew's `libxml2` stable is 2.15.4, the build first on
`PATH`. It is the validator Ampache's conformance test and slice 1's measurement both used, and it ships in every
mainstream Linux distribution (Fedora's own is 2.12.10 at `/usr/bin/xmllint` here). It joins CLAUDE.md's pinned
binaries, so each pass's sweep re-reads it, and Homebrew holds it with `brew pin` between sweeps (R18).

## Readings this plan takes

Each is settled here, with its reason. The review may contest any of them.

- **R13. Every transcribed row lands `owed`.** An `active` row needs a session with a dubplate source (R12), and no
  dubplate response exists until slice 3. The sketch's "each `active` where a recorded session exercises it" cannot
  apply to any row. A row's `slice` is fixed by one mechanical rule and one list. **The rule:** the first `Slice N`
  token in the "Test owed" cell, matched without regard to case, is the slice; a rung is never a slice. **The list**
  holds every row whose cell carries no such token, with its slice and the reason: the earliest slice R27 assigns
  among the row's endpoints, or slice 3, the first slice that serves an envelope, for a row with no endpoint or one
  not served at 0.8.

  | Slice | Rows without a slice token |
  |---|---|
  | 3 | L-01 to L-05, L-07, L-08, L-14 to L-16, L-23, L-37 (no single endpoint); L-12, L-54 (`getOpenSubsonicExtensions`); L-27 (`getAlbumList2`); L-34 (its endpoints answer error 0 at 0.8, first served in slice 3); L-51 (ruled text); A-01, A-02, A-11 (no endpoint); A-04 (the ten `musicFolderId` endpoints); A-12 (the empty-success set) |
  | 4 | L-24 to L-26 (`search3`); A-05 (`search3`) |
  | 5 | L-17 (`download`); A-09 (`stream`, direct-path ranges) |

  L-35 and L-52 retire (Task 6). L-02 sits in the list because "rung 16" is not a slice. Every other row takes its
  token: L-06 takes the "slice 4" inside its citation, which is also the earliest of its three endpoints' slices;
  L-19 takes slice 3 and L-43 slice 7. The transcription test recomputes each slice from the record's cell, the
  rule, and this list.
- **R14. The record's ten classes are kept beside the machine class.** The ten ([ledger] "Columns and classes") say
  why dubplate differs; the four machine classes say what a row suppresses. Mapping ten onto four would lose the
  first, and a machine class chosen before slice 3 adjudicates a row against a real session is a guess. So a row
  carries the record's class as `divergence` always, and the machine `class` only once it is `active`.
- **R15. Agreement rows live in the assignment, not the ledger.** A-01 to A-12 describe where dubplate matches both
  oracles; they suppress nothing, and an `active` agreement row would go stale on its first clean grade. The record
  itself moved them out of its divergence table ([ledger] "Agreement rows").
- **R16. The assignment replaces the folder-view `by-construction` rows.** Under ruling C no dubplate grade compares
  `getIndexes` or `getMusicDirectory` with an oracle, and `getMusicFolders` is gonic's alone. An assignment that
  never pairs dubplate with the wrong oracle is the general form of what those rows did per endpoint.
- **R17. The XML tier lives in the harness.** It execs a binary and reads an unlicensed XSD from the user cache.
  Ruling B keeps the protocol module single-licensed and self-contained, so `internal/opensubsonic/conformance` stays
  the JSON checker. The new package is `internal/harness/subsonic/xmlrule`.
- **R18. xmllint is pinned by its libxml version.** Its messages land in goldens, and two libxml2 builds sit on this
  workstation's `PATH`. The tier refuses any version but the pinned one, so a drift fails loudly instead of as a
  golden diff, and a bump goes through the dependency sweep. Nothing held the version before this plan: `uupd`
  runs `brew upgrade -y` daily (quotes table), so an upstream release would land overnight and red every lane's gate.
  The lane launch therefore records libxml2 in the Brewfile and runs `brew pin libxml2`; the refusal message names
  the sweep's bump path (`brew unpin libxml2 && brew upgrade libxml2`, then the goldens, then `brew pin libxml2`, then a
  `brew upgrade` of the held dependents). `brew pin --help` warns that "Other packages which depend on newer versions
  of a pinned formula might not install or run correctly", and `brew uses --installed libxml2` names eight on this
  workstation: chafa, ffmpeg, gdk-pixbuf, imagemagick, librsvg, libxkbcommon, shared-mime-info, and xdotool. While the
  pin holds they stay on the builds they have; the upgrade after the re-pin brings them level.
- **R19. The projection resolves an element's type by its position.** The XSD names `album`, `entry`, and `artist`
  with different types under different parents, and the OpenAPI additions differ by type. Keying on the element name,
  as the prototype does, drops `AlbumID3`'s additions from a `Child`-typed `album` and keeps `Child`'s, so every
  `getAlbumList` body would fail for the wrong reason.
- **R20. Errors are located by element ordinal.** xmllint reports a line, and an oracle's body may be one line. The
  projection writes each start tag on a new line and records, for each line that opens an element, that element's
  document-order ordinal in the original body, counting the newlines it actually wrote (end tags, text, and a newline
  inside text all move the count; quotes table). `xmlrule` reports ordinals and `grade` turns an ordinal into a path
  under its own tree rule. The path grammar then lives in one package. Under that rule a name that occurs once among
  its siblings is not indexed, so an XML path and a JSON path agree only for a list of two or more; a `contract` row
  for a one-element list in XML names `format: xml` and the unindexed path.
- **R21. 2b records only what a 2b check consumes.** The empty-states session (Task 8) feeds the contract tier's
  predicted flags and the differ's empty-container rules. The `f` forms, the missing `v`, unknown endpoints, binary
  endpoint errors, and the unauthenticated extensions call are dubplate's by-design answers graded against the spec
  (L-01, L-05, L-06, L-12, L-14); slice 3 records them as their first consumer. L-07's real-body proof is a
  perturbation, never a recorded echo, because recording gonic's `getArtistInfo2` would commit a credential.
- **R22. The two-collection fixture is built here.** The spec's ladder entry 2 (`:885-886`) and the record's "Routed"
  list put it in slice 2; slice 3 then grades its filter against oracle evidence that already exists.
- **R23. The seed library stays byte-identical.** Slice 3 joins the seed session by building the committed manifest
  and matching its library digest, so the two collections are separate manifests and the seed manifest never changes.
- **R24. No oracle version moves in 2b.** A bump is "a deliberate pass that re-runs the ledger" (spec `:345-346`). If
  Task 0 finds a newer Navidrome or gonic release, it is recorded as held with that reason.
- **R25. The roster walk records what a member sees.** Request captures exist in [base-url] for Feishin and Rufin, and
  each later slice asks whether a flow works in the app. The walk script names each step's expected observation and
  records pass or fail with a screenshot path; it adds no logging proxy.
- **R26. The gate's rendered scope stops following the conform scope** (BACKLOG #27). Nothing under
  `internal/harness/` reaches the browser, so the scoped gate drops Chromium for this lane's tasks.
- **R27. The assignment's values come from three named sources, and a split verdict takes its narrow value.**
  Section 3 of [record] names 51 of the bundle's 87 endpoints. The empty-success set ([record] section 1; spec
  `:810-814`) adds `getInternetRadioStations`, `getPodcasts`, and `getNewestPodcasts`. Spec `:809` says every other
  endpoint answers error 0, so the other 33 (`search`, `getVideos`, the podcast writes, chat, jukebox, bookmarks, user
  administration, `hls.m3u8`, and the rest, computed against the pinned paths) are `not-served-0.8` and contract only,
  joined by `getTranscodeDecision` and `getTranscodeStream` under ruling A. That is 52 served and 35 not served.
  - **Oracle.** Section 3's value, under the rulings: `getIndexes` and `getMusicDirectory` contract only (ruling C);
    `getMusicFolders` gonic, with "shape only" in the reason. The info family keeps "Both": the family row is
    specific, and both oracles answer the empty success. The three empty-success additions are contract only
    (section 3's cross-cutting row). A verdict that splits one endpoint takes the value fair for every request form of
    it, and its reason keeps section 3's wider grant word for word for the serving slice: `stream` is `navidrome`
    (gonic only for direct play); `getOpenSubsonicExtensions`, `getPlayQueue`, and `savePlayQueue` are contract only
    (Navidrome only for the carve-out or the empty answer); the share endpoints are `navidrome`. A verdict whose
    exception a ledger row already names keeps its oracle: `getCoverArt` both (L-39), `star` and `unstar` both
    (L-15), the playlist endpoints both (L-50).
  - **Slice.** The ladder slice whose text names the surface (spec `:871-924`), else slice 3, the first slice that
    serves an envelope. Slice 3: the system endpoints, ID3 and folder browsing, `getMusicFolders`, `getGenres`, the
    album-info pair, the info family, `getAlbumList`, `getAlbumList2`, `getRandomSongs`, `getSongsByGenre`,
    `getLyrics`, and the three empty-success additions. Slice 4: `search2`, `search3`, `getCoverArt`. Slice 5: `stream`,
    `download`. Slice 6: `getUser`, `tokenInfo`, the four share endpoints. Slice 7: `getStarred`, `getStarred2`,
    `star`, `unstar`, `scrobble`, `getNowPlaying`, the five playlist endpoints, `getPlayQueue`, `savePlayQueue`.
- **R28. A `Both` row activates as one row per pair.** 2a's `Row.applies` matches an `active` row only on the graded
  pair, so slice 3 activates L-01 by committing `L-01/navidrome` and `L-01/gonic`, each with a two-source `sources`
  drawn from the record's set. The transcription's accounting accepts that form, and its fidelity check binds owed rows
  only (Task 6), so activation never edits the dated record or weakens the test.

## Global constraints

The 2a plan's global constraints hold unchanged ("Where it runs", the clean room, the boundary, no export ahead of a
caller, Go conventions, the four states, determinism, credentials, forward hooks). Added or changed for 2b:

- **Territory.** This lane's: `internal/harness/subsonic/**` (the new `xmlrule` package included),
  `cmd/harness/` (new verb files, the registration block, `capture.go`, `synth.go`, `recordoracle.go`,
  `oracleup.go`, `grade.go`, and their tests), and `docs/roster-walk.md`. Contested, checked at Task 0 and at each
  boundary: `scripts/check.sh` (Task 1) and `Makefile` (Task 2), both edited only while the lane registry shows no
  other code executor. Read-only: `internal/opensubsonic/**`, `internal/engine/tagio`.
- **No module dependency.** `go.mod` and `go.sum` stay as master has them.
- **The prerequisites the gate gains.** After Task 2 the gate needs the XSD in the user cache (`make subsonic-xsd`)
  and `xmllint` at the pinned version on `PATH`, held by `brew pin`. Each test that needs them fails naming the missing
  piece and the command that supplies it; none skips (CLAUDE.md, "A check either gates or does not exist").
- **Credentials.** No session body carries a credential parameter (Task 7), in addition to 2a's request rule; the
  session writer and loader enforce both.
- **Forward hooks.** Every task is neutral on the hooks at spec `:930-1025`. Task 10 records oracle evidence for hook
  22 (the integer music folder id) and builds none of it; slice 3 owns the hook.

## Lane launch (the conductor, before Task 1)

The worktree is held from 2a, level with master at `c4b7afd`.

1. **Reverse edges cleared.** Lane 1 is idle; if rung 14's executor has started in the main tree, `scripts/check.sh`
   and `Makefile` are written to STATUS as contested and Tasks 1 and 2 wait for a quiet window.
2. **Contested surfaces written** to STATUS's lane 3 row.
3. **Runaway guard armed**, with the sleep inhibitor sized to the run's end.
4. **Lane registry row** in STATUS: lane 3, the worktree, the branch, this plan, the sub-ceiling.
5. **Provisioning.** No venv (2a R1). `web/node_modules` is present from 2a; the first gate refuses it if stale.
   Recorded with versions: `podman`, `xmllint` (resolved path and libxml version), the Feishin and Rufin Flatpaks, and
   `ls /dev/kvm`. libxml2 is held (R18): a `brew "libxml2"` line lands in `~/.dotfiles/bluefin/Brewfile` under the
   workstation's manifest rule, and `brew pin libxml2` runs. The conductor quotes `brew list --pinned` naming it, and
   `brew info libxml2`; if that still reads "Installed (as dependency)", it runs `brew install libxml2` and quotes the
   result, so the formula no longer hangs on its dependents. Done at launch on 2026-10-07: `brew install libxml2` moved
   it to "Installed (on request)", `brew pin libxml2` ran, `brew list --pinned` names libxml2, `xmllint` resolves to
   `/home/linuxbrew/.linuxbrew/Cellar/libxml2/2.15.4/bin/xmllint` reporting libxml version 21504, and the Brewfile line
   landed at dotfiles `c25573b`. The pin holds eight installed dependents on their current builds (chafa, ffmpeg,
   gdk-pixbuf, imagemagick, librsvg, libxkbcommon, shared-mime-info, xdotool; `brew pin --help` warns that packages
   needing a newer pinned formula "might not install or run correctly"), so the Brewfile line's comment names them.
6. **Sub-ceiling** assigned.
7. **AC power**, and the throughput baseline: the conductor's one full gate in the worktree, timed and quoted (2a's
   was about 586 s).

## Task 0: Sweep, pins, and the PR re-read (conductor)

Through the `dependency-upgrade` skill, since implementers carry no Skill tool.

**Outcomes.**
- **The sweep** over `go.mod` (kin-openapi now included), `web/package.json`, `tools/venv/requirements.txt`, and
  the pinned binaries, run in the main tree and landed on master before segment A, then the worktree rebased onto it.
  The branch already carries the plan commits, which touch only `docs/`, so the rebase replays them; that is expected,
  never a stop signal. If the sweep moved `web/` dependencies, `npm ci --prefix web` runs in the worktree after the
  rebase. Or one line when nothing moves. The 2026-10-07 sweep is the
  baseline. libxml2 is read as a pinned binary for the first time.
- **Oracle and client versions** read: Navidrome, gonic (R24: a newer release is held), Feishin, Rufin, Tempus. On
  2026-10-07 each is at its [pins] version and the newest production release.
- **The PR re-read** (spec ladder entry 2; [record] section 11): #162, #237, #258, #261, and #278, their state and any
  merge; the repository's `main` commit; the bundle's sha256 at its URL; whether a `NOTICE` file exists (ruling B's
  watch). If none moved, one line in the report and Task R does not run. If one moved, Task R runs, and a merged PR
  that touches a ruling's premise stops the run.
- **The rulings landed:** STATUS's owner questions drop forks A to C, and the spec's "answered pending ruling" text is
  queued for the close's batch.
- **Quotes re-read** per the table above, through one `haiku` research dispatch, since the conductor reads no
  source itself.
- Every `file:line` here re-resolved against the worktree base.

**Acceptance.** The report lists each version with its evidence, each re-resolved pointer, and the PR states. No
gate.

## Task R (conditional): Re-pin the OpenAPI bundle

Runs only when Task 0 finds the bundle's sha256 moved. **Files:** `internal/harness/subsonic/data/openapi/`
(`openapi.json`, `LICENSE` if it changed, `PROVENANCE.md`), the pinned sum in `grade/contract.go`, and every golden
under `grade/testdata/`. **Hook:** neutral. **Class:** `tool`.

**Outcomes.** The new bundle, its license, and its provenance replace the old together. Each golden regenerates
under its test's `-update` flag. The report lists every finding count that moved, by golden and kind, and every
contract finding that appeared or vanished.

**States and proofs.** `TestPinnedOpenAPISumMatchesTheCommittedFile` passes on the new sum; `TestSeedGetAlbumContract`
asserts the new exact violation lists. **Mutation:** leave the old sum in `contract.go` (the load refuses).

## Segment A

### Task 1: The rendered scope, and capture's overwrite guard

**Files:** `scripts/check.sh`; `cmd/harness/capture.go`, `cmd/harness/synth.go`, and their tests. **Hook:** neutral.
**Gate:** the full form, heavy lane.

**Outcomes.**
- **BACKLOG #27.** Under `--scope-changed`, the rendered leg runs exactly when the change set reaches one of the
  rendered leg's inputs (the list in `derive_rendered_affected`, `scripts/check.sh:241-256`), whatever the conform
  scope. A change set that reaches `internal/harness/` and no rendered input runs `make web`, `vet`, `lint`, the whole
  `go test ./...` (the conform suite included), and `tidy-check`, skips `make web-rendered`, and says so in one line.
  The scope header line prints both scopes. The full, merge, and reduced forms are unchanged.
- **BACKLOG #22.** `harness capture --out` and `harness synth --out` refuse through one shared guard, keeping synth's
  rule (an existing `case.json` refuses, naming the path; `synth.go:48-52`), so capture never drops an existing
  case's `Input`, `Profile`, `EvalDate`, or `Expected`. Synth's comment that capture has a seed-only contract becomes
  true.

**States table.** `--scope-changed` with no merge base: unchanged (the full-scope fallback with its warning). A change
set touching `web/` or a serve-shell file: the rendered leg runs. Only `docs/`: neither suite narrows more than today.
Capture or synth into a directory with no `case.json`: proceeds. Into one holding `case.json`: refused, naming it, and
no byte of the directory changes.

**Test mandate and fixture state.** Shell carries no behavior tests (CLAUDE.md, "Shell"). The evidence is three runs
quoted in the report: a scratch commit touching one `internal/harness/` file shows the header with `conform=full
rendered=skipped` and no Chromium process; one touching a `web/` file shows `rendered=full`; the full form's header is
unchanged. `shellcheck` is clean. A Go test runs capture into a case directory written by synth, asserts the refusal,
and compares the directory's bytes before and after. **Mutations:** drop the guard from capture (the test fails);
let the guard accept an existing `case.json` with an empty `Expected` (the test fails).

### Task 2: The pinned XSD in the user cache

**Files:** `internal/harness/subsonic/xmlrule/` (new: the pin, the cache path, the verified read),
`cmd/harness/fetchxsd.go` (new verb `fetch-subsonic-xsd`), one line in `cmd/harness/main.go`, a `subsonic-xsd`
target in `Makefile`. **Hook:** neutral.

**Outcomes.**
- One pin, in one place: the URL, the size (32,775 bytes), and the sha256 (`9baee72f...0387`). No other channel sets
  them, and no committed file carries the XSD's bytes.
- The cache path is `<os.UserCacheDir()>/dubplate/subsonic-rest-api-1.16.1.xsd`.
- The verified read returns the file's bytes only when its sha256 matches the pin. Every consumer reads through it.
- The verb follows `fetch-unidecode`'s shape (`cmd/harness/fetchunidecode.go`): a no-op making no request when the
  cache verifies; otherwise one fetch, verified before an atomic write. The make target runs it. Its comment says why
  it sits outside `check`, as `unidecode-wheel`'s does.
- The implementer runs `make subsonic-xsd` once and quotes the result.

**States table.** Cache directory unresolvable: the `os.UserCacheDir` error, wrapped. File absent: an error naming
the path and `make subsonic-xsd`. Present with the wrong sum: an error naming both sums and the path. HTTP non-200:
an error carrying the status, nothing written. Fetched bytes with the wrong sum or size: refused, nothing written.

**Test mandate and fixture state.** Tests set `XDG_CACHE_HOME` to a temporary directory. The fetch tests run against
an `httptest.Server` serving canned bytes, through the same seam `fetch-unidecode`'s tests use. The verified-read
proof needs a cached file with one byte changed. A test asserts the verb is registered. **Mutations:** skip the sum
check on read (the corrupted-cache proof passes and the test fails); write before verifying (the bad-download proof
finds a file); fetch when the cache verifies (a test counting requests fails).

### Task 3: The scoped projection, the endpoint binding, and the xmllint runner

**Files:** `internal/harness/subsonic/xmlrule/` (the projection, the binding, the runner, their `testdata/`). **Hook:**
neutral. **Model:** opus.

**Outcomes.**
- **The schemas.** From the XSD: each complexType's attribute and child-element names, following `xs:extension`;
  the type of each child element at its position; and the `Response` choice's payload elements. From the pinned
  OpenAPI document: each component schema's property names, following `$ref` and `allOf`. An XSD type maps to the
  OpenAPI schema of the same name, or through a short rename table in the package, each entry citing both names
  (`AlbumWithSongsID3` and `AlbumID3WithSongs` among them; the `Response` envelope and `Error` have their counterparts
  named there too). A type's additions are its OpenAPI property names minus its XSD names. An XSD type with no
  OpenAPI counterpart and no rename entry (`Captions`, `AudioTrack`, `VideoConversion`, `ArtistInfoBase` at the pin)
  has no additions, so any extra name on it is a residual.
- **The endpoint binding** (the XSD's choice is `minOccurs="0"`, quotes table). Each request is checked against its
  endpoint, its HTTP status, and its content type, read from the pinned OpenAPI document as the quotes table's
  `getAlbum` probe shows: the success schema's payload member and whether it is required. An `ok` body must carry that
  payload element as the envelope's child, and nothing else in the choice; it may carry none only when the success
  schema names no payload (`ping` and the writes). A `failed` body must carry `error`. The status must be one the
  operation declares. The content type must be `application/xml` or `text/xml`. Each mismatch is a violation at
  ordinal 1, the envelope. A non-envelope success the operation declares (a binary body) is not an XML-tier input.
- **The projection** walks a body from the root, resolving each element's type by its position (R19). Each
  attribute or child element its type's additions name is dropped. A name declared at its position stays. A name in
  neither the XSD type nor the OpenAPI schema is a **residual**: it is reported as a violation from the residual scan
  and then dropped from the projected document, so xmllint validates everything after it (xmllint stops at a
  parent's first unexpected child, quotes table). `status` and `version` on the envelope are never dropped.
  Whitespace-only character data is not copied.
- **Declared unvalidated.** A request whose endpoint's payload element is not in the `Response` choice (for example
  `getOpenSubsonicExtensions` and `getLyricsBySongId`, [contract] Q3) is reported as declared unvalidated, naming the
  element, after the binding has run. The list comes from the XSD and the binding, never from a hand-kept table. It is
  never a pass.
- **The runner** checks `xmllint` once per process: resolved through `PATH`, its libxml version read from stderr
  (quotes table) and equal to the pinned `21504` (R18). It validates the projected document with `--noout --schema
  <cached XSD>` under `exec.CommandContext`, with `Cmd.Stderr` set to its own buffer (quotes table, `ExitError`).
  Exit 0 is valid and exit 3 is violations; any other exit is a tool failure carrying stderr. Under exit 3 only lines
  of the form `<file>:<line>: Schemas validity error : ` are violations and the `<file> fails to validate` trailer is
  expected. An unmatched line that directly follows a validity-error line continues that error's message (a long
  message wraps onto a context line), and joins it; an unmatched line with no error line before it is a tool failure.
- **Error location** (R20). The line-to-ordinal table counts the newlines the projection actually wrote. Every
  xmllint error line becomes a violation at that line's ordinal, with the attribute name when the message names one
  and the message text after `Schemas validity error : `. An error whose line maps to no element is a violation at
  ordinal 0, never dropped.
- **A prefixed body.** A body whose elements carry a namespace prefix (no oracle emits one, and the Go encoder never
  writes one, quotes table) is one violation at ordinal 1 naming the prefix as unsupported by the projection. It is
  never a pass and never a silent skip.
- **The result** of a body is one of: valid; violations (each an ordinal, an optional attribute, a message, and
  whether it came from the binding, the residual scan, or xmllint); declared unvalidated; or a tool failure.

**States table.** XSD not cached or failing its sum: Task 2's errors. OpenAPI document lacking a schema the rename
table names: an error at construction naming it. `xmllint` absent: an error naming it. Wrong version: an error
naming both versions, the resolved path, and the bump path. Empty body: a violation, "empty body". Malformed XML: a
violation carrying the parse error. Wrong or missing payload, `failed` without `error`, an undeclared status, or a
non-XML content type: a binding violation at ordinal 1. Payload outside the choice: declared unvalidated. A prefixed
body: the unsupported violation. Many errors, past 64 KiB of stderr included: all of them, sorted by ordinal. xmllint
exiting other than 0 or 3, or printing under 3 an unrecognized line with no validity-error line before it: a tool
failure. An unrecognized line directly after a validity-error line: part of that error's message.

**Test mandate and fixture state.** Unit proofs use small XSD and OpenAPI documents written for the tests under
`xmlrule/testdata/`, with invented type names only, so no file there excerpts the Subsonic XSD (ruling B): one
element name with two types at two positions; an extension chain; a rename. Real-pin proofs use the cached XSD and
the committed OpenAPI bundle:
- **The additions** are pinned per type as a golden under `xmlrule/testdata/`. The report reconciles the golden with
  slice 1's measurement (49 names over the envelope, `error`, `Child`, the album pair, the artist pair, and the
  playlist pair, [contract] Q3) and names the cause of any difference.
- A real seed `getAlbum` XML body from each oracle, with its residual attributes removed by hand, validates.
- A hand-written `getAlbumList` body whose `album` elements carry a `Child` addition (`mediaType`) validates; an
  `album` there carrying `isCompilation`, an `AlbumID3` addition `Child` lacks, is one residual violation. Together
  they separate position-keyed scoping from both a name-keyed scope and a union over every type sharing the name.
- **Behind a residual.** An `album` whose first child is an unexpected element, followed by a `song` with a
  non-integer `track` and a misspelled attribute (`titel`): three violations, the first and third from the residual
  scan and the second from xmllint.
- An error on the second of two `song` elements in a one-line body lands on the second song's ordinal, and an error
  on the second of two `album` elements in a real indented gonic body lands on that album's ordinal. A body with a
  multi-line text node before a violating element locates it correctly.
- `status` and `version` survive the projection of a real envelope.
- **The binding:** a `getAlbum` body with no payload, a `getAlbum` body carrying `<artists>`, and an `ok` body carrying
  `<error>` are each a binding violation; a `ping` body with no payload validates.
- An `openSubsonicExtensions` body is declared unvalidated. A prefixed body is the unsupported violation.
- A fake `xmllint` (a helper process re-executed from the test binary, writing its version line to stderr) reporting
  another version is refused; one exiting 4 is a tool failure; one emitting more than 64 KiB of error lines yields
  every one of them.

**Mutations:** key the scope on the element name (the `getAlbumList` proof fails); union additions across a name's
types (the `isCompilation` proof fails); drop names globally (the envelope proof fails); keep residuals in the
projected document (the `track` error behind a residual vanishes); skip the residual scan (`titel` is not reported);
locate errors by projected line without the written-newline table (the indented-body proof fails); read the version
from stdout (the fake's version is never seen); parse `ExitError.Stderr` (the 64 KiB proof fails); read exit 3 as
valid; accept any xmllint version; drop the binding (the three binding proofs fail); treat declared unvalidated as
valid.

## Segment B

### Task 4: The XML contract tier in the grader

**Files:** `internal/harness/subsonic/grade/` (`contract.go`, `grade.go`, `tree.go`, their tests), `cmd/harness/grade.go`,
the regenerated `grade/testdata/seed-report.golden`. **Hook:** neutral.

**Outcomes.**
- The contract tier runs every XML response of both graded sources through `xmlrule` with its endpoint, status, and
  content type, beside the JSON check. A violation is a `contract` finding with format `xml` and the violating source,
  at the path its ordinal maps to; ordinals 0 and 1 map to the envelope.
- **The ordinal maps to the tree's path** under `tree.go`'s rule: a name repeated among one element's children is an
  array and gets an index, a single child does not, and an attribute is a member. A JSON path always indexes a list,
  so the two tiers name one path only for a list of two or more (R20); the plan claims nothing wider.
- A declared-unvalidated body is no finding. The report counts them by endpoint and source in its header.
- A tool failure ends the grade with an error and no report.
- `grade.New` reads the XSD through Task 2's verified read; `harness grade` uses the cache by default.
- The seed golden regenerates. The task's report gives the XML contract findings by oracle and message class, and
  sets them beside the vendor residuals [contract] Q3 measured (Navidrome's `lastModified` on `artists` and
  `songCount` on `directory`; gonic's `title` and `album` attributes on ID3 `album`, `artistImageUrl` as an element,
  `albumCount` on a non-ID3 `artist`, the missing `lastModified` on `indexes`), naming each as present or absent.

**States table.** XSD absent or bad: `grade.New` fails with Task 2's error. An XML body that does not parse: one
`contract` finding (as for JSON) and the oracle tier's `parse` finding, both kept. A binding violation: one `contract`
finding at the envelope. Many violations: all, sorted.

**Test mandate and fixture state.** `TestGradeIdenticalBodiesReportOnlyContract` gains an XML twin: a hand-built session
whose two sources hold the seed's real gonic `getAlbum` XML body byte for byte, with gonic's `serverVersion` in both
sources' metadata (it carries the `title` and `album` residuals on the ID3 `album`, [contract] Q3), asserting zero
oracle-tier findings and a literal, non-empty list of `contract` findings, once per source, at their tree paths. A path
proof places one violation on the second of two `song` elements and asserts the indexed path; a singleton proof places
one on the only `song` of a one-song album and asserts the unindexed path. A row proof shows a `contract` row with
format `xml` suppressing its finding and not the same path's JSON finding. **Mutations:** skip XML in the contract tier
(the twin's literal list fails); map an ordinal to a path by element name alone (the indexed-path proof fails); index a
singleton (the singleton proof fails); count declared unvalidated as valid and drop it from the header; continue past a
tool failure.

### Task 5: The oracle assignment and the area list

**Files:** `internal/harness/subsonic/data/assignment.json` (new), its loader and the grader's use of it in
`grade/`, `data/ledger.json` (three rows retired), `grade/ledger_test.go` (`TestCommittedLedger`'s owed counts).
**Hook:** neutral.

**Outcomes.**
- **Every endpoint in the pinned document has one entry** (87 on 2026-10-07): its oracle (`both`, meaning Navidrome
  primary and gonic secondary; `navidrome`; `gonic`; or `contract-only`), the slice that first serves it or
  `not-served-0.8` for an endpoint that answers error 0 at launch, and a reason with an evidence pointer. The values
  are R27's: section 3 under the rulings, the empty-success additions, the 35 not-served endpoints, the narrow value
  for each split verdict with section 3's wider grant kept in its reason, and the slice per R27's list.
- **The area list** ([servers] section 10.3) as entries: multi-user isolation, per-endpoint folder filtering, and
  permission refusals, each with its slice; and the album shapes compilation, multi-disc, collaboration credit, and
  Unicode, each naming the seed manifest album that carries it. The empty album is recorded as not applicable, with
  the reason that an album is admitted only with its tracks.
- **The grader reads the assignment when a graded source is `dubplate`.** A request whose endpoint the assignment
  does not grant to the pair's oracle skips the oracle tier, keeps the contract tier, and is counted in the report
  under "not assigned to <oracle>" by endpoint. An endpoint the assignment lacks (a deliberate unknown-name probe) is
  graded contract only and counted under "unmapped". An oracle pair never consults the assignment, so the seed golden
  is unchanged by this task.
- **Ruling C's retirement.** The three folder-view `by-construction` rows leave `data/ledger.json`; the commit message
  names ruling C and R16. `TestCommittedLedger`'s owed-by-slice counts follow.

**The assignment's one channel.** Only `data/assignment.json` sets an endpoint's oracle; nothing in the grader keeps a
second list. The value test below is a proof against the record, not a second source the grader reads.

**States table.** File absent: an error naming it. Empty: refused. Malformed or with an unknown member: the location.
Many bad (an endpoint twice, an unknown oracle value, an empty slice or reason, a `not-served-0.8` entry whose oracle
is not `contract-only`, an album shape naming no manifest album or one that lacks the shape): every one in one error.

**Test mandate and fixture state.**
- **Coverage.** The assignment's endpoint set equals the pinned document's path set with `/rest/` cut, so a re-pin
  that adds an endpoint fails until it is assigned.
- **Values.** One test holds section 3's 51 verdicts as a literal the implementer transcribes and the diff-reviewer
  reads against the record, and asserts each entry's oracle and slice per R27. Every `contract-only` entry must be on
  a named list: section 3's contract-only rows, ruling C's two, the three split verdicts narrowed to contract only, the
  three empty-success additions, and the 35 not served. A new `contract-only` value therefore needs a test edit. The
  not-served set equals the 87 paths minus R27's 52 served endpoints.
- **Album shapes**, by computation against the manifest: more than one disc; a `Various Artists` album artist; a track
  with two artist values; a name outside ASCII.
- **The grading proof** is a hand-built session with sources `dubplate`, `navidrome`, and `gonic` over real seed bodies.
  The session loader refuses a body whose `serverVersion` differs from its source's metadata (`session.go:369-377`, run
  by `Write` and `Load`), so every `dubplate` body's envelope `serverVersion` is rewritten to the one value the
  `dubplate` source's metadata names. Its fixture state: dubplate's `getAlbum` and `getIndexes` bodies are gonic's, so
  they differ from Navidrome's; dubplate's `getMusicFolders` body is Navidrome's, so it differs from gonic's; dubplate's
  `getIndexes` body is perturbed so the contract tier reports one known violation (a required member removed); and one
  request goes to an endpoint the assignment lacks, with differing bodies. Grading dubplate against Navidrome reports a
  non-zero count of `getAlbum` oracle findings, zero oracle-tier findings and a skip count of one each for `getIndexes`
  and `getMusicFolders`, the known `getIndexes` contract finding, and the unmapped request under "unmapped" with zero
  oracle-tier findings. Grading dubplate against gonic reports a non-zero count of `getMusicFolders` oracle findings.
  Both non-zero counts are asserted below the envelope, since until Task 7 the rewritten `serverVersion` differs from
  both oracles' by value.

**Mutations:** consult the assignment for an oracle pair (the seed golden changes and `TestSeedGrade` fails); skip the
contract tier on an unassigned request (the known `getIndexes` finding vanishes); treat an unmapped endpoint as `both`
(the unmapped request gains oracle findings); flip `getTranscodeDecision` to `navidrome` (the values test fails); mark
`getArtist` `contract-only` (the values test fails); accept a `not-served-0.8` entry with an oracle; accept an album
shape the manifest lacks.

### Task 6: The ledger's machine form and the transcription

**Files:** `grade/ledger.go` and its tests, `data/ledger.json` (rewritten), `data/assignment.json` (its `agreements`
section), Task 5's assignment loader in `grade/` and its tests (the agreements' validation), a transcription test in
`grade/`, and the inline `Row` literals that build `owed` rows with a class (`grade/ledger_test.go`,
`grade/grade_test.go`, `cmd/harness/grade_test.go`). **Hook:** neutral.

**Outcomes.**
- **The row gains the record's columns** (R14): `divergence`, one of the record's ten classes (`declared: folder
  view`, `declared: apiKey path`, `by-design`, `contract-deviation`, `oracle-wrong`, `oracle-lenient`,
  `oracles-disagree`, `feature-absent`, `no-oracle`, `unspecified-shape`), and `test`, the "Test owed" text. The
  loader requires both on every `owed` row; an `active` row may omit them, so the 27 active rows of
  `grade/testdata/seed-ledger.json` that grade the grader stay as they are. A committed-ledger test requires both on
  every row of `data/ledger.json`. The machine `class` is required on an `active` row and absent on an `owed` one. An
  `owed` row's `sources` is `dubplate` plus each oracle the record's "Oracle" column names (`Both` names two), so up
  to three; an `active` row still names exactly one pair (R28).
- **The transcription.** `data/ledger.json` holds L-01 to L-54 as `owed` rows, with `slice` set per R13, except the
  retired rows below. The `agreements` section of the assignment holds A-01 to A-12 (R15), each with its area,
  shared behavior, evidence, test, and slice. The rows the rulings touch take the ruled text: in the `dubplate`
  column L-34 (not declared at 0.8, ruling A), L-51 (an empty `getIndexes` and error 70 from `getMusicDirectory`,
  graded against the contract and the spec, ruling C), and L-54 (the launch subset without `transcoding`, ruling A);
  in the `test` column L-51 ("Slice 3: an empty `getIndexes` and error 70 from `getMusicDirectory`, asserted against
  the spec", ruling C).
- **Retired, each with its ruling:** L-35 (ruling A), L-52 (ruling C), and 2a's `apikey-unoracled` row, which L-10
  replaces.

**The privileged columns.** `slice`, `divergence`, and the ruled texts have one channel each: a ledger edit, proved
by the transcription test. No code computes them at load time.

**States table.** As 2a's ledger, plus: an `owed` row with no `divergence` or an unknown one, an `owed` row with an
empty `test`, an `owed` row carrying a `class`, an `active` row without one, an `owed` row with more than three
sources or none of them `dubplate`, and an agreement with an empty field: every one named in one error.

**Test mandate and fixture state.** The transcription test reads both Markdown tables from
`docs/research/2026-10-06-slice-1-divergence-ledger.md`:
- **Round trip.** Parsing each table into rows and rendering the rows back yields the record's table bytes exactly,
  for the 54-row table and the 12-row agreement table. The proof needs the real tables, so an empty parse cannot pass:
  it asserts 54 and 12 rows.
- **Accounting, one way.** Every id in the two tables is a committed row (by its id, or as `<id>/<oracle>` under
  R28), an agreement entry, or a retired id with its ruling. The ledger may carry other ids, so slice 3 can add a newly
  found divergence.
- **Fidelity, on `owed` rows.** Each committed `owed` row's transcribed columns equal the record's, except the four
  ruled texts, which the test names with their rulings. An `active` row derived from a record row names a pair drawn
  from that row's sources.
- **Slice.** Each row's and each agreement's `slice` equals R13's recomputation from the record's cell, the rule, and
  the list. `TestCommittedLedger`'s owed-by-slice map takes the recomputed counts.

The rendering lives in the test file, so no export exists without a caller. `TestCommittedLedger` and
`TestCommittedLedgerRowsAreLive` keep passing (every row names `dubplate`; no row is `active`). **Mutations:** map
`Both` to one oracle (the round trip fails); drop the `test` column (the round trip fails); let an `owed` row carry a
machine class; omit L-35 from the retired list (accounting fails); accept a fourth source; move one row's `slice`
(the slice check fails). Two activation proofs run on a fixture ledger: an `active` row `L-01/navidrome` naming
`dubplate` and `navidrome` passes accounting and fidelity; an `active` row for a record row that names one oracle,
activated against the other oracle, fails fidelity.

## Segment C

### Task 7: Encoding quirks in the differ, and the credential guard

**Files:** `grade/differ.go`, `grade/tree.go`, their tests; the credential scan in `internal/harness/subsonic/session/`
(beside 2a's request rule) and its tests; `cmd/harness/recordoracle.go`'s test; the regenerated
seed golden; `grade/testdata/seed-ledger.json` (20 rows retire); `grade/grade_test.go` (the stale-row test's live
row). **Hook:** neutral.

**Outcomes.**
- **E1 and L-08.** The envelope's `serverVersion` and `type` compare by presence and type, never by value, since each
  names its own server ([contract] E1). A failed envelope's `error/message` compares the same way, and its `code` by
  value ([ledger] L-08). Each rule sits in one named table in the grade package beside the order table, citing its
  row. The 20 `server-type-*` and `server-version-*` `volatile` rows of `grade/testdata/seed-ledger.json` then match
  nothing and retire, the commit message citing E1, and `TestGradeNamedSessionsReportsAStaleRow`'s live row
  (`ping` `subsonic-response/type`, `grade_test.go:470`) moves to a path that stays volatile (`song/created`).
- **E6.** JSON `null` against an empty container is one `type` finding; `null` against an absent member is one
  `presence` finding; `{}` against `{"playlist":[]}` is one `presence` finding at `playlist`; `[""]` against `[]` is
  one `presence` finding at index 0.
- **E5, E10, and the XML envelope.** Bodies that differ only in escaping (the escape `\u0026` against a literal `&`
  in JSON, [contract] E10; `&#34;` against `&quot;` in XML), in the presence of an XML declaration, or in a namespace
  prefix against the default namespace, yield zero findings. A singleton against an array stays one finding (2a).
- **L-07, the credential scan.** Every string leaf of every body that holds a URL whose query carries `u`, `p`, `t`,
  `s`, or `apiKey` is one finding at that leaf's path, which the `session` scan returns and the writer and loader turn
  into a refusal. The finding names the parameters and never their values. The query is read with `ParseQuery`, whose map holds every valid parameter
  even when it errs (quotes table); a query it cannot parse cleanly is itself a finding, so the scan fails closed. The
  scan separates a leaf's query from its path by cutting at the first `?` (`strings.Cut` semantics) and parses only
  the remainder (`url.ParseQuery` semantics), never `url.Parse` on the whole leaf, since `url.Parse` errs on a bad
  percent escape in the path before the query is read. A
  query parameter whose name merely ends in one of them (`format`, `size`) never fires.
- **The body-side credential guard.** The scan lives in the `session` package beside 2a's request-parameter rule, and
  both the session writer and the loader run it over every body: the writer refuses, naming the request and source
  and writing nothing, and the loader refuses a committed session that carries one. `harness record-oracle` inherits
  it through the writer, and slice 3's dubplate recorder will too. `grade` has no L-07 finding of its own: every
  session reaches `Grade` through the loader, whose refusal makes such a finding unreachable, so the writer's and the
  loader's refusal is the L-07 check. A session therefore never commits an echoed credential (R21).
- The seed golden regenerates; the report gives the change in findings by kind (the `serverVersion` and `type` value
  findings leave).

**The credential channels.** A credential can reach a session through a request parameter (2a's writer and loader
refuse it) or through a response body (the same writer and loader refuse it). Every session is written through the
writer, so no third channel exists.

**States table.** A leaf with no query: no finding. A query `ParseQuery` rejects (`?u=a&t=x;s=y`): one finding naming
the parse failure and every credential name it did read. A leaf with a bad path escape and a credential parameter in
its query (`/rest/x%zz?p=secret`): still refused, naming `p`. A body with several credential URLs: one finding per leaf. A
refused recording: nothing written, the request and source named. A committed session carrying one: the loader
refuses, naming the file.

**Test mandate and fixture state.** `TestDifferPerturbations` gains one row per rule above, each on a real seed body,
asserting exactly the expected findings and kinds, or zero. The E6 fixtures: gonic's `015` body (it carries
`replayGain: null`) for the two `null` rows, gonic's `008` body (it carries `releaseTypes: [""]`) for the `[""]` row,
and a perturbation for the `{}` against `{"playlist":[]}` row. The credential proof perturbs a real Navidrome
`getArtist` body's `artistImageUrl` (it carries `?size=600`) by appending `&u=admin&t=x&s=y`, asserting one finding
naming `u`, `t`, and `s` and no value; the unperturbed body yields none; a malformed-query row (`;s=y`) yields its
finding; a bad-path-escape row (`/rest/x%zz?p=secret`) yields one finding naming `p`. The proof calls the `session`
scan directly, since `Grade` cannot reach a body that carries a credential. The recorder proof uses 2a's fake stack serving such a body and asserts the refusal and an empty output
directory; a loader proof refuses a hand-written session carrying one. **Mutations:** compare `serverVersion` by value
(the golden and a named row fail); match parameter names by suffix (the `size` proof fires); read the query with
`URL.Query` (the malformed-query row is silent); parse the whole leaf with `url.Parse` (the bad-path-escape row is
silent); print a credential's value; let the writer write before scanning;
drop the loader's scan (the loader proof fails).

### Task 8: The empty-states session

**Files:** `internal/harness/subsonic/data/sessions/empty-states.spec.json` and the recorded
`data/sessions/empty-states/`, `grade/testdata/empty-states-ledger.json`, `grade/testdata/empty-states-report.golden`,
and their test. **Hook:** neutral.

**Outcomes.**
- The spec requests, in JSON and XML against both oracles: `getLyrics` (artist and title of one seed track, which
  carries no lyrics), `getLyricsBySongId` (a `song:` key), `getNowPlaying` (idle), `getInternetRadioStations` (none
  configured), `getPlaylists` (none), and `getStarred2` (nothing starred). These are 2a's carry (lyrics, `nowPlaying`,
  radio) and the E6 empty containers.
- The implementer records it live with `harness record-oracle` over the seed library and commits it. A request either
  oracle answers outside the envelope drops from the spec with its reason, by the conductor's ruling.
- A gate test grades it, Navidrome against gonic, under the test-only ledger, and pins the report as a golden. The
  test ledger holds only rows the pair needs, each with its reason, and each matches at least once.
- **The study's predicted flags.** The task's report sets the golden beside [contract] E6 and [ledger] L-03: gonic's
  `structuredLyrics: null`, Navidrome's idle `nowPlaying` without `entry`, gonic's error 70 for `getNowPlaying`, and
  the mixed empty containers. Each is named as present (with its path) or absent (with the recorded body). The test
  pins each prediction's recorded state, present at its path or absent, as the report gives it, so an absent
  prediction is a pinned fact, never a red test.
- **The known-deviation list's oracle nulls.** The test ledger carries gonic's `null` members as `contract` rows,
  each matching at least once; that is the mechanism the scope table names, and `lastScan` waits for slice 3 (L-20).

**States table.** As 2a's Task 5: the spec absent, empty, or malformed; an unresolvable key; an oracle failing
mid-run; a body carrying a credential (Task 7). Each refuses with nothing written.

**Test mandate and fixture state.** The committed session loads through the loader with every request answered by
both oracles. The golden test asserts byte equality, a second grade's identical bytes, and every test-ledger row's
count of at least one. **Mutations:** delete one response file (the loader refuses); add a test-ledger row matching
nothing (the stale-row assertion fails).

### Task 9: Two collections in the fixture and the stack

**Before segment C launches (the conductor).** The conductor observes Navidrome's library creation once, inline: a
one-time observation with no script to maintain, at the cost of GUI work in the conductor and a re-observation owed by
R24's next deliberate Navidrome bump. It drives Playwright MCP against the Navidrome container `harness oracle-up`
starts at the pinned digest: log in as the first admin, open the libraries page, add a library with "+" (the
multi-library page, `:46-57`) at a container path that exists and does not overlap `/music`, and read the network log.
It records the login exchange and the creation request (method, path, header names, body shape, status, response body
shape; no credential value). **Then it replays them:** against a fresh `harness oracle-up`, `curl` sends the recorded
login and creation with the credentials that run prints, and the conductor quotes the creation's status, then
`getMusicFolders` and `getScanStatus` once the new library's scan finishes (`:62`), passwords redacted. The second
path holds no tracks (the container mounts only `/music`), so the quote settles the request and the folder listing,
not how `count` sums over two libraries that both hold tracks; that is Task 9's live evidence. The record of both runs
lands in `docs/research/2026-10-07-navidrome-library-request.md`, committed before segment C launches; the
step needs nothing from Tasks 7 or 8. No one opens Navidrome's source. If the observation cannot be made or the replay
fails, Navidrome's second library is ruled unmeasured (the conductor's ruling), and Tasks 9 and 10 are dispatched in
their degraded forms.

**Files:** two new manifests beside the seed's `manifest.json` in `internal/harness/subsonic/data/fixture/`, so
`fixture.Load` finds the shared `seeds/` directory unchanged (quotes table); `fixture/`, `stack/`,
`cmd/harness/oracleup.go`, `cmd/harness/recordoracle.go` (the stack call only), their tests. **Hook:** neutral.

**Outcomes.**
- **Two small manifests** (three or four albums each) reuse the committed seed files and add no binary seed. Their
  paths and albums are disjoint, so no root nests in another (CLAUDE.md, "A collection is a store row ... with no two
  roots nested"). One album artist appears in both, so a filtered artist list differs from the unfiltered one in a way
  a test can name. Every track in both carries one `genres` value shared across the collections (the seed manifest
  sets none, quotes table), so `getSongsByGenre` has something to filter. Neither carries the untagged album (`Loose
  files`), whose artist each oracle names differently. The seed manifest is unchanged, byte for byte (R23), and still
  builds the library digest the seed session records.
- **The stack takes an ordered list of collections**, each a name and a manifest, built into one root each and mounted
  read-only with `:ro,z`. A one-collection list behaves exactly as today, mounted at `/music`, so the seed recording
  path is unchanged. A longer list mounts each collection at `/collections/<name>`, disjoint siblings, since Navidrome
  refuses a library path that overlaps another (quotes table). gonic receives every collection through
  `GONIC_MUSIC_PATH` with each name as its alias (quotes table). Navidrome's first collection is its `MusicFolder`;
  each later one is created by replaying the conductor's recorded login exchange and creation request over HTTP. The
  first admin needs no grant: "Admin users automatically have access to all libraries" (quotes table).
- **Readiness** on both oracles: `getMusicFolders` lists exactly one folder per collection, and `getScanStatus`
  reports `scanning` false with `count` equal to the total tracks across collections.
- **A multi-collection library digest** is a sha256 over each collection's name and library digest, in list order. A
  one-collection list's digest is that library's own digest, so the seed's recorded digest still matches.
- `harness oracle-up` and `harness record-oracle` take the collection list by flag, defaulting to the seed library.
- **The live evidence:** the implementer runs `harness oracle-up` with both collections and quotes each oracle's
  `getMusicFolders` and finished `getScanStatus` (password redacted).
- **The degraded form**, taken when the conductor rules Navidrome's second library unmeasured: before Task 9, by
  dispatching it in this form; or during it, when the replay answers an error or Navidrome's readiness never sees one
  folder per collection and the total count. In that second case the implementer reports the live behavior with its
  quote and does not work around it, and the conductor re-dispatches Task 9 in this form. Navidrome mounts the first
  collection only, as its `MusicFolder`, and creates no library; gonic still receives every collection. Readiness checks
  each oracle against its own collections: one folder and the first collection's tracks on Navidrome, one folder per
  collection and the total on gonic. Each oracle's source records the digest of the collections it mounted, under the
  rule above. The replay, its fake routes, and their proof drop; a gate test asserts Navidrome's argv carries one mount.

**States table.** An empty collection list: refused. Two collections with one name: refused. Overlapping paths across
collections: refused, naming both. Navidrome's library creation answering an error: an error naming the status and
the collection. A folder count other than the collection count at readiness: an error naming the oracle and both
counts. A scan count short of the total: 2a's count error.

**Test mandate and fixture state.** Gate tests drive the stack against 2a's fake container runtime: the argv carries
both mounts with `:ro,z` at disjoint paths and gonic's aliased `GONIC_MUSIC_PATH`; the fake oracle serves both
recorded routes, the login and the creation, and asserts the creation's body; readiness refuses a folder count of
one. A fixture test builds both manifests twice and asserts identical bytes
and digests, and asserts the seed manifest's digest is unchanged. **Mutations:** mount a collection without `:ro`;
omit gonic's alias; accept a folder count of one; compute the combined digest without the names (a test swapping two
names' contents fails).

## Segment D

### Task 10: The music-folders session

**Files:** `cmd/harness/recordoracle.go` (the `collection:` key), `grade/` (`Config` naming several manifests, the
digest check), `cmd/harness/grade.go` (the manifest flag), `data/sessions/music-folders.spec.json` and
`data/sessions/music-folders/`, `grade/testdata/music-folders-ledger.json`, `grade/testdata/music-folders-report.golden`,
and their tests. **Hook:** neutral (records evidence for hook 22; slice 3 builds it).

**Outcomes.**
- **A `collection:<name>` key resolves** to the music folder id whose `getArtists`, filtered by that id, lists exactly
  the set of the collection manifest's `albumartist` tag values. No match, or more than one, refuses the run, naming
  the key and the oracle. The resolution lands in the session's resolution map like every other key.
- **The session** requests, in JSON against both oracles: `getMusicFolders` (and in XML), then the ten
  `musicFolderId` endpoints of [record] section 4 (`getIndexes`, `getArtists`, `getAlbumList`, `getAlbumList2`,
  `getRandomSongs`, `getSongsByGenre`, `getStarred`, `getStarred2`, `search2`, `search3`), each with the parameter
  absent, set to each collection, and set to the unknown literal `999`. Each request carries the parameters its
  endpoint needs and no more (an alphabetical `type` for the album lists, a `size` at least the total track count for
  `getRandomSongs`, an empty `query` for the searches, a genre the collection manifests carry for `getSongsByGenre`).
- **The grader takes the session's manifests.** `Config` names an ordered collection list, and the grader refuses a
  session whose recorded library digest differs from the list's combined digest, naming both. It builds each
  collection into a temporary directory to compute its digest, since `Build` is the only digest source. `harness
  grade` takes the list by flag, defaulting to the seed.
- A gate test grades the session, Navidrome against gonic, under a test-only ledger and pins the golden. It also
  asserts L-19's recorded behavior directly on the bodies: for `999`, Navidrome answers error 70 on all ten and gonic
  answers `ok` with an empty result; and A-04's: for each collection id, each oracle's filtered `getArtists` lists only
  that collection's artists.
- **The gonic-only form**, taken only if Navidrome's second library was ruled unmeasured before or during Task 9: the
  session records gonic alone for the filter requests, and no pair golden exists. The test asserts A-04 and L-19's gonic
  half on the bodies. L-19's Navidrome half stays slice 1's observation ([servers] section 2), recorded as an erratum at
  the plan's end.

**States table.** As Task 8, plus: an ambiguous or unmatched collection key, refused before anything is written; a
grade given the wrong manifests, refused with both digests.

**Test mandate and fixture state.** The resolution proof uses canned responses in which the two folders' artist sets
differ by the shared artist, and a refusal case in which they are equal. The digest-check proof grades the seed session
with the collection list and asserts the refusal. **Mutations:** resolve a collection by folder position (the canned
reversed order fails); skip the digest check; resolve on artist overlap instead of set equality (the shared-artist case
resolves twice).

### Task 11: The roster walk

**Files:** `docs/roster-walk.md` (new: the method and one script per client), the run record
`docs/research/2026-10-07-slice-2b-roster-walk.md` (dated the day it runs), and the screenshots or DOM
captures it cites, under the record's own directory. **Hook:** neutral. **Class:** `docs`, whose in-chain bar is facts
and outcomes; `pass-execute`'s `docs` bar says "Prose register belongs to the register chain, not this review", and
no register chain runs on an evidence record. **Gate:** `check.sh --reduced`.

**Outcomes.**
- **The method** is [base-url]'s, written once: each client under an isolated configuration home inside its Flatpak
  sandbox, keyring talk names denied, Rufin on a private Xvfb display, Feishin through its debugging port. Every tree
  [base-url] hashed is hashed before and after with no change allowed: both apps' real `~/.var/app/<id>/` trees
  (except `cache/` and Flatpak's `.ld.so/` linker cache, which [base-url] explains) and the host GNOME keyring files.
  Credentials typed into a client come from `harness oracle-up`'s output and never reach the record. A committed
  screenshot is cropped to the client window and read for host paths before staging.
- **One script per client**, each step stating the action, the expected observation, and what to record. The steps
  cover the flows [clients] section 2 names for that client: add the server as the OpenSubsonic type, log in, browse
  artists by ID3, open the multi-disc album and the compilation, play a track, show cover art, open the folder view
  (an empty or minimal view is the expected 0.8 answer under ruling C), search, star a track, and create a playlist.
- **Feishin 1.17.0 and Rufin 0.16.6** each walk the whole script once against Navidrome, and run the folder-view
  step against gonic too, the one step whose expected answer the oracles change on the 0.8 surface. 2b's deliverable is
  the scripts, which slices 3 to 8 walk against dubplate. The record gives each step's pass or fail with its
  observation, and every claim cites a screenshot or a DOM capture over Feishin's debugging port. A step that fails on
  an oracle is evidence about that oracle, recorded and not fixed.
- **Tempus 4.28.0** gets its script from source reading ([clients] section 2d), marked unproven. Its walk is deferred
  with the unblock condition: `ls /dev/kvm` succeeds after Geoff enables VT-x, then the SDK, emulator, and image
  install per [pins] "Client installs" with their manifest lines, then the script runs. The conductor checks
  `/dev/kvm` before this task; if it now exists, the Tempus walk runs here.
- **Amperfy 2.1.1** is listed for rung 16's attended gate (re-plan `:140-141`), with its script from [clients]
  section 2c.

**States table.** A client that will not start: recorded with its error, and the walk continues with the other. Any
real-profile or keyring hash changed: the run stops (the stop rules), since the isolation failed. An oracle-up that
fails: the task stops; it is 2a's verb, so this is a blocker.

**Acceptance.** The record names every step's outcome for both desktop clients against Navidrome and the folder-view
step against gonic, each citing its capture, the hashes before and after, and the Tempus and Amperfy deferrals with
their conditions. The `diff-reviewer` reads the record against [clients] section 2's flow list for coverage and checks
each pass claim against its capture.

## Close

### Task 12: Reviews, merge, and the ledgers

The conductor runs these in order. Every reviewer dispatch asks for correctness gaps ranked by consequence, not
polish, and a fold may refuse a finding with a one-line reason.

1. One `code-simplifier:code-simplifier` dispatch over the pass's changed Go, pasting
   `.claude/instructions/simplifier-brief.md` verbatim, then one `diff-reviewer` read of its diff.
2. One fresh-context vacuity read (`claude-opus-5-5`) over the tests of `xmlrule`, the tests Tasks 4 to 7 added in
   `grade`, and the credential-scan tests in `session`, the code this plan specifies only by outcome. Its single
   question: which test could pass while the behavior it names is broken? One fold task if it finds any.
3. `go vet -tags oracle ./cmd/harness/... ./internal/harness/...` in the worktree, once (CLAUDE.md, "Lanes").
4. The merge: the merge token; `git rebase master`; `check.sh --merge` in the worktree; `git merge --ff-only
   lane-3-slice-2` on master; the full gate on master; one `diff-reviewer` over the merged range if the rebase was not a
   fast-forward.
5. One `go-architecture-reader` per touched Go package: `internal/harness/subsonic/xmlrule`, `.../grade`,
   `.../fixture`, `.../stack`, `.../session`, and `cmd/harness`. One fold task for accepted
   findings.
6. Queued repo edits in one batch on master: `CLAUDE.md` (the gate paragraph's prerequisites gain `make subsonic-xsd`
   and `xmllint` at libxml 2.15.4, held by `brew pin libxml2`, with its eight held dependents named (chafa, ffmpeg,
   gdk-pixbuf, imagemagick, librsvg, libxkbcommon, shared-mime-info, xdotool); the pinned binaries gain libxml2 with
   its unpin, bump, regenerate, re-pin, and dependents-upgrade path; the scoped gate's rendered-leg rule as Task 1 left it); the streamer spec's revision
   note marking forks A to C ruled, with `:342-345` amended so gonic is no longer authoritative for `getIndexes` and
   `getMusicDirectory` (ruling C); BACKLOG #22 and #27 closed; HISTORY's 2b entry (what landed, what the gates and
   reviewers caught, each golden's counts by kind, both budgets scored, and what slice 3 would be wrong to
   rediscover, among it how a row on a `contract-only` endpoint closes, since L-17, L-20, L-21, and L-51 can never
   activate against an oracle); STATUS (lane 3's row, the Tempus deferral, slice 3's readiness).
7. Worktree teardown, and the registry row says so.

**Acceptance.** Master is green with the merged work. The staged exported-unused leg reports no hit in `xmlrule` or
in the identifiers this pass added. `go list -deps ./cmd/dubplate-server | grep -E 'internal/(harness|opensubsonic)|getkin'`
prints nothing. `git diff master...lane-3-slice-2 -- go.mod go.sum internal/opensubsonic` is empty before the
rebase. Every committed session loads. No file under `internal/` carries the XSD's bytes, and no `.xsd` file under
`xmlrule/testdata/` names a type of the real XSD (`AlbumWithSongsID3`, `NowPlayingEntry`). `brew list --pinned` names
libxml2. The oracle-tag vet is clean.

## Not verified at planning

Revision 1.1 closed four items from revision 1's list: xmllint's exit codes are documented (quotes table, DIAGNOSTICS);
the mechanics review resolved `xmllint` inside the gate's own `systemd-run` scope to Homebrew's 2.15.4; it measured the
light lane's fit (header); and the multi-library page answers the first admin's access (quotes table).

- **Navidrome's library-creation request.** The conductor observes and replays it before Task 9. Task 9's and Task
  10's degraded forms cover a failure before or during Task 9.
- **Navidrome's `getScanStatus` count over two libraries that both hold tracks.** No quote covers it, and the
  conductor's replay cannot show it. Task 9's live evidence shows it, and a count other than the total triggers the
  degraded form.
- **gonic's folder names under aliases** in `getMusicFolders`. Task 9's live evidence shows them; collection keys
  resolve by artist set, not by name, so either answer works.
- **The empty-states answers.** [contract] E6 predicts them at the probe fixture; the seed library is a different
  fixture, and Task 8's report reconciles each.
- **Whether every slice 1 projection count reproduces from the bundled file**, since the prototype read the
  specification repository's per-schema files. Task 3's count proof names any difference.
- **Whether Feishin's debugging port and a private Xvfb display still work** for this session's driver. Slice 1 used
  both on 2026-10-07; Task 11 states the method it used.

## Errata (filled during execution)

A conductor ruling on a live behavior that contradicts this plan lands here with its task and remedy, as do Task 9's
and Task 10's degraded forms if they are taken.

- **E1 (Task 1, `9f878c3`).** The no-Chromium evidence is the scope header (`conform=full rendered=skipped` for a
  harness-only change, `conform=skipped rendered=full` for a `web/package.json` change) plus the code path, which
  never reaches `make web-rendered` when the rendered scope is skipped. The runs were stopped after the header, so no
  process watch backs it. The conductor accepts the reading; segment B's light-lane gates are the live confirmation.
- **E2 (Task 3, `ca2991e`).** Three readings the plan left open, taken by the implementer and accepted on review.
  A fourth violation source, `Document`, covers an empty, malformed, or prefixed body, which none of the plan's three
  sources names. The endpoint binding holds a set of payload members, each with its required flag, since
  `jukeboxControl` declares two optional payloads; every other endpoint has one required payload, as the plan says.
  The payload members exclude the envelope's own names (the XSD root type's attributes plus its additions), or `ping`
  would read `version` and `serverVersion` as required payloads. A tool failure is a returned error wrapping
  `ErrToolFailure`, never a `Result` outcome.
- **E3 (Task 3, observed on review).** Past line 65535 of a projected document, xmllint 21504 reports an error one
  line too high, so the ordinal lookup can land on the next element. No recorded body comes near that size. Filed as
  BACKLOG #28.
- **E4 (segment B pre-flight, at `f4ba980`).** Six amendments, passed to the implementers in segment B's criteria.
  Task 4 handles Task 3's two non-fatal returns: `xmlrule.ErrUnknownEndpoint` grades as the JSON tier's `no-route`
  outcome finding, and `xmlrule.ErrNotXML` (a binary endpoint answered with a non-XML body) yields no XML contract
  finding. Only `ErrToolFailure`, `ErrNotCached`, and `ErrChecksum` end the grade; both rows join the states table
  with a proof. Task 4's XML twin extends the hand-session helper to take format, content type, and `serverVersion`,
  since `handSessionAt` hard-codes JSON and Navidrome's version. Task 4's tests are `contract_test.go`,
  `grade_test.go`, and `differ_test.go` (no `tree_test.go` exists). Task 5's Files gain `grade.Config`'s assignment
  field (`grade.go`), `cmd/harness/main.go`'s `subsonicPaths` and `defaultSubsonicPaths`, and `cmd/harness/grade.go`.
  Task 6's inline-literal change is `grade/ledger_test.go`'s `owedRow` helper alone. Task 6's agreement round trip
  renders the record's five agreement columns (id, Area, Behavior shared, Evidence, Test owed), with the slice derived.
- **E5 (segment C pre-flight, at `5d00313`).** Thirteen amendments, passed to the implementers in segment C's
  criteria. The conductor's Navidrome observation (`docs/research/2026-10-07-navidrome-library-request.md`)
  succeeded in both runs, so Task 9 runs in its normal form; the record drove the UI with the repo's Playwright 1.63.0
  because the Playwright MCP server found no Chrome, and the replay needed only the bearer header.
  Task 7: `TestGradeNamedSessionsReportsAStaleRow`'s live row sits at `grade/grade_test.go:741-748` and moves to
  Endpoint `getSong` with path `subsonic-response/song/created`, or it goes stale. `TestSeedGrade`'s hardcoded totals
  (`grade_test.go:83-92`) join the Files and move with the 124 retired envelope value findings (expected 1645 findings,
  976 unledgered, value 245; the report's change by kind confirms). The E5, E10, XML-envelope, and E6 rows already
  hold on the current differ and land as pinned rows with mutations; only E1 and L-08 change code. The `session` scan
  carries its own walker over JSON string leaves and XML attribute values and element text, since `grade` imports
  `session`. The recorder proof adds a payload hook to `cannedOracle`'s options in `cmd/harness/recordoracle_test.go`
  and asserts an absent or empty output directory, since `Write` refuses before `MkdirAll`. For `?u=a&t=x;s=y` the
  finding names `u` only, because `ParseQuery` drops the `t=x;s=y` pair whole; the states table says so.
  Task 8: each spec entry carries one format, so every request is listed once per format, and JSON `getArtists`,
  `getArtist`, and `getAlbum` entries precede `getLyricsBySongId` so its `song:` key resolves; those requests are part
  of the recorded session. The test ledger carries gonic's `null` against Navidrome's absent member as
  `by-construction` rows scoped to each path, since a `contract` row suppresses only the contract finding.
  Task 9: `cmd/harness/recordoracle.go`'s scope is `recordConfig`, the library build and digest in
  `runRecordOracle`, and the `startOracles` signature with its test fake, not the stack call only. The collection list
  is a repeated `--collection name=manifest` flag on `oracle-up` and `record-oracle`, mutually exclusive with the
  existing `--manifest`, which stays the one-collection default; a name that is empty, starts with `.`, or contains
  `,`, `->`, `=`, or `/` is refused (a states row). `stack.Start` runs Navidrome's ping, then the login (polled within
  the readiness timeout), then each creation, then the folder-count and scan wait on both oracles. The fake
  `oracleServer` gains `getMusicFolders`, `/auth/login`, and `/api/library` routes, and the existing stack tests take
  the folder check, the one-collection path included. Housekeeping: the "What 2a built" row for `data/ledger.json` now
  holds 52 owed rows and none active, and the `fixture.Load` cite reads `manifest.go:70-86`.
- **E6 (segment C, paused 2026-10-07 at Geoff's ask).** Tasks 7 (`55d6398`) and 8 (`f5c916d`) were accepted with no
  fix rounds; the run was stopped before Task 9 wrote anything, so segment C resumes at Task 9 in its normal form
  under E5. Task 7 took two readings on review: the loader's refusal names the directory, request, endpoint, and source
  but not the body file (the file follows from the request and source), and one states row's name ("an escaped query
  separator is still read") claims a property its body does not exercise. Task 7's scan treats every string leaf
  containing `?` as a URL and refuses one whose query `ParseQuery` rejects, as the plan's fail-closed rule says, so a
  title such as "Is This It? 50% off" would refuse a session; filed as BACKLOG #29 for slice 3. Task 8 recorded 30
  spec entries with no request dropped; every prediction is pinned, the two absent ones as absent.
- **E7 (segment D pre-flight, at `7c012e7`).** Task 9 was accepted with no fix round; its live run held the normal
  form (Navidrome listed folders 1 "Music Library" and 2 "second", gonic 0 "first" and 1 "second", both counted 22),
  so Task 10 runs in its normal form. Navidrome names its first folder "Music Library", so nothing resolves a folder by
  name. Ten amendments, passed to the implementers in segment D's criteria.
  Task 10: a session spec carries only requests, so the collections come from the recording command:
  `record-oracle --spec music-folders.spec.json --collection first=<collection-a.json> --collection
  second=<collection-b.json> --out .../music-folders`, and those two names are the ones the grader passes. The
  `collection:` key joins `refKinds`; the recorder gains each collection's `albumartist` set (a `fixture.Load` per
  manifest) and resolves the key itself, from the `getMusicFolders` answer and one filtered `getArtists` per folder id,
  never from spec order. The expected sets are A = {Alpha Artist, Delta Duo, Borealis Trio, Zoë Ångström} and B =
  {Alpha Artist, Carmine Ensemble, Various Artists, The Meridians}. If an oracle lists a set differently (compilation
  handling or a leading article), the implementer quotes the live list and the resolver matches instead on the artists
  that sit under exactly one collection and that both oracles list verbatim; the shared artist alone never resolves,
  so the overlap mutation still fails. That reading needs no further ruling and is reported as an erratum. `harness
  grade`'s collection flag is new: a repeated `--collection name=manifest`, defaulting to the seed, mutually exclusive
  with nothing (grade has no manifest flag today). `grade.New` builds every collection; the library the grader reads
  tracks from is their union in list order, and the assignment loads as today, independent of the list. The new digest
  check uses `fixture.CombineDigests` with the recorder's names; it is separate from the existing source-to-source
  comparison.
  Task 11: [clients] section 2 is a per-client endpoint table with no flow list, so `docs/roster-walk.md` defines the
  flow list itself (the outcome's list is the checklist), cross-checked against [clients] 2a and 2b and the folder-view
  demand in section 4; the reviewer reads coverage against that list. No host Xvfb exists, so Rufin's display follows
  [base-url]'s throwaway Xvfb container, and the record names the image and its digest. Feishin's real
  `~/.var/app/org.jeffvli.feishin` tree now exists (slice 1 created it) and is hashed as found; the `cache/` and
  `.ld.so/` exceptions apply to both apps. `/dev/kvm` is still absent, so the Tempus walk stays deferred.
