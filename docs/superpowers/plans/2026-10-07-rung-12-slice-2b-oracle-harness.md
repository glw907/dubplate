# Rung 12, Slice 2b (the Oracle Harness, Second Pass) Implementation Plan

Status: **revision 1, drafted 2026-10-07; awaiting its `spec-plan-review` and Geoff's approval.** Owner forks A, B,
and C from slice 1 are ruled (below). Every other decision is a reading below, taken on its merits and reported.

History:

- Revision 1: drafted from slice 1's records and 2a's merged code. It replaces the "Pass 2b (sketch; authored in full
  at its Task 0)" section of the 2a plan, which now points here.

**Parent plan:** `docs/superpowers/plans/2026-10-06-rung-12-slice-2-oracle-harness.md` (revision 1.2, "the 2a plan").
Its readings R1 to R12, its global constraints, and its clean-room rule stay in force for this pass; this plan adds
readings R13 to R26. Task numbers here are 2b's own, and commits name them as "rung 12 slice 2b Task N".

**Spec:** the 0.8 streamer spec (`docs/superpowers/specs/2026-09-07-streamer-2-0-design.md`, revision 3.1, "the
spec"; it still says 2.0), ladder entry 2 at `:880-887`, "Evidence" at `:314-362`, under the re-plan
(`docs/superpowers/specs/2026-10-05-rungs-12-16-replan-design.md`, revision 2.1, "the re-plan"). The sketch cited
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
  and on every seed XML body through a golden (Tasks 3 and 4).
- **The transcription** by a byte-exact round trip on slice 1's committed tables and an accounting test over all 66
  ids (Task 6).
- **The assignment** on a hand-built three-source session (Task 5).
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
`docs/superpowers/research/2026-10-06-slice-1-study-record.md`, section 19), relayed by the conductor on 2026-10-07.

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
on the heavy lane. From Task 2 on, a task whose change set reaches none of the rendered leg's inputs launches no
browser (Task 1's outcome), so it runs with `gateLane: "light"` (`~/.claude/docs/pass-gate-economy.md`: "A gate that
launches no browser takes the light lane"). If a light gate is killed at its 3G memory cap, the conductor moves the
rest of that segment to the heavy lane and records it; that is never a task failure. Task 11 takes `check.sh
--reduced`. Segment boundaries and the pass end take the full form (`check.sh`, heavy lane), and the merge ritual
takes `check.sh --merge`. A boundary reuses the last task's full gate when `HEAD` has not moved since it ran.
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
- recording Navidrome's second library as unmeasured if it cannot be created without reading Navidrome's source
  (Task 9);
- moving a segment to the heavy lane, and reordering segments C and D if a live task is blocked.

**These stop the run (STATUS written, then stop):**
- the hard ceiling;
- a new executor in the main tree editing `scripts/check.sh`, `Makefile`, or `cmd/harness/main.go` (the contested
  surfaces), or any warm change in this worktree the conductor did not dispatch;
- a pinned input that no longer verifies: the XSD's sha256 at its URL, the pinned oracle digests, or xmllint's
  version on `PATH`;
- a merged OpenSubsonic PR or a NOTICE file upstream that changes a ruling's premise (Task 0 or Task R);
- a gate red on a commit whose diff does not explain it, after one re-run;
- an `escalate` verdict the conductor cannot resolve inside the task's files.

**Owner-gated steps, batched at the end in one STATUS question:** enabling VT-x (it unblocks Tempus's walk, Task 11),
and any finding that would reopen fork A, B, or C. Nothing deploys, and the box is never contacted.

## Sources

Code and spec lines are cited at `c4b7afd` on master, where this worktree is level. Library and tool behavior is
quoted at the version this plan uses, and the commands below were run on the workstation on 2026-10-07.

- Slice 1's records under `docs/superpowers/research/`, cited by the study record's aliases: the study record
  (`2026-10-06-slice-1-study-record.md`, "[record]"), the divergence ledger ("[ledger]"), and the contract
  ("[contract]"), client-demands ("[clients]"), base-URL ("[base-url]"), server-behavior ("[servers]"), evidence-pins
  ("[pins]"), and prior-art-cells ("[cells]") records. Also `2026-10-05-client-roster-audit.md`.
- The spec: "Evidence" (`:314-362`), "The protocol layer is a separable module" (`:371-408`), "License"
  (`:421-427`), ladder entry 2 (`:880-887`), the forward hooks (`:930-1025`), "Prior art" (`:1026-1064`).
- The re-plan: the rung 12 ladder text and the roster walk (`:113-141`); the operator delta (`:137-138`).
- 2a's merged code under `internal/harness/subsonic/`, `internal/opensubsonic/`, and `cmd/harness/`; HISTORY's 2a
  entry (`docs/HISTORY.md:5-142`); BACKLOG #22 and #27 (`BACKLOG.md:85-89`, `:104-107`).
- The dependency sweep of 2026-10-07 (`docs/superpowers/research/2026-10-07-dependency-sweep.md`).

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
| The Subsonic 1.16.1 XSD at `https://subsonic.org/pages/inc/api/schema/subsonic-rest-api-1.16.1.xsd` | Fetched 2026-10-07: HTTP 200, 32,775 bytes, sha256 `9baee72fa24087190ad22be4f7674d628acc7a3e0fb65160d74115b166df0387`, the pin [contract] recorded. `targetNamespace="http://subsonic.org/restapi"`, `elementFormDefault="qualified"`, `attributeFormDefault="unqualified"`; the root element `subsonic-response` has type `sub:Response`, whose payload is one `xs:choice` (`xsd:9-56`). No `xs:any` and no `xs:anyAttribute` | Tasks 2, 3 |
| The same XSD, element types by position | `album` is `sub:AlbumWithSongsID3` in `Response` (`:19`), `sub:Child` in `albumList` (`:387`) and `searchResult2` (`:311`), and `sub:AlbumID3` in `albumList2` (`:393`) and `searchResult3` (`:319`); `entry` is `sub:NowPlayingEntry` in `nowPlaying` (`:284`) and `sub:Child` in playlists (`:350`, `:367`) | Task 3: types resolve by position |
| `xmllint` (libxml2 2.15.4, `/home/linuxbrew/.linuxbrew/bin/xmllint`, first on `PATH`) | `xmllint --version` prints `xmllint: using libxml version 21504`. `/usr/bin/xmllint` is libxml2 2.12.10 (`21210`). Man page (`--schema SCHEMA`): "Use a W3C XML Schema file named SCHEMA for validation"; `--noout`: "Suppress output". Observed on a seed body: each error is one stderr line `<file>:<line>: Schemas validity error : Element '{http://subsonic.org/restapi}artist', attribute 'sortName': The attribute 'sortName' is not allowed.`, then `<file> fails to validate`, exit 3. After an unexpected child (`Element '{...}roles': This element is not expected. Expected is ( {...}album )`) no later sibling of that parent is reported. An unreadable file exits 4 (`Can't open`), and a malformed one exits 4 with a `parser error` line. Navidrome's seed XML bodies are one line, so every error reads line 1 | Task 3. The installed man page carries no exit-status table, so exit codes 3 and 4 are observed, not documented (see "Not verified at planning") |
| `encoding/xml.Decoder.RawToken` (go1.27.1) | "RawToken is like Decoder.Token but does not verify that start and end elements match and does not translate name space prefixes to their corresponding URLs." | Task 3: the projection keeps the body's namespace declarations as written |
| `encoding/xml.Decoder.Token` | "Each of the Name structures contained in the Token has the Space set to the URL identifying its name space when known." "Token expands self-closing elements such as <br> into separate start and end elements" | Tasks 3, 4 |
| `encoding/xml.Encoder.Indent` | "Indent sets the encoder to generate XML in which each element begins on a new indented line" | Task 3: one start tag per line, so an xmllint line names one element |
| `encoding/xml.Encoder.EncodeToken` | "Callers that create an Encoder and then invoke EncodeToken directly, without using Encode or EncodeElement, need to call Flush when finished to ensure that the XML is written to the underlying writer." | Task 3 |
| `os.UserCacheDir` | "On Unix systems, it returns $XDG_CACHE_HOME ... if non-empty, else $HOME/.cache." "If the location cannot be determined (for example, $HOME is not defined) or the path in $XDG_CACHE_HOME is relative, then it will return an error." | Task 2 |
| `os/exec.LookPath` | "LookPath searches for an executable named file in the directories named by the PATH environment variable"; "on success the result is an absolute path" | Task 3: the version refusal names the binary's path |
| `os/exec.Cmd.Output` | "Any returned error will usually be of type *ExitError. If c.Stderr was nil and the returned error is of type *ExitError, Output populates the Stderr field" | Task 3 |
| `os.ProcessState.ExitCode` | "ExitCode returns the exit code of the exited process, or -1 if the process hasn't exited or was terminated by a signal." | Task 3: only 0 and 3 are validation outcomes |
| kin-openapi v0.149.0 `openapi3` | `Loader.LoadFromData`: "LoadFromData loads a spec from a byte array"; `Schema` carries `AllOf SchemaRefs` and `Properties Schemas`; `SchemaRef` carries `Ref string` and `Value *Schema`; `Components.Schemas`; `Paths.Map`: "Note: iteration on Go maps is not ordered." | Tasks 3, 5 |
| The pinned OpenAPI bundle (`cb54c03c...`) | 195 component schemas, among them `Child` (54 properties), `AlbumID3` (29), `ArtistID3` (10), `Playlist` (13), and `AlbumID3WithSongs`, `ArtistWithAlbumsID3`, `PlaylistWithSongs` as `allOf` compositions; 87 paths, each `/rest/<endpoint>` | Tasks 3, 5 |
| gonic `README.md` at `v0.22.0` | "gonic supports multiple music folders" (`:104`); `export GONIC_MUSIC_PATH="my albums->/path/to/albums,my compilations->/path/to/compilations" # with aliases` (`:122`) | Task 9 |
| `navidrome/website` `content/en/docs/usage/features/multi-library.md` at `b0fe572` (Apache-2.0) | "When Navidrome starts, it automatically creates a default library using your `MusicFolder` configuration. This becomes "Library 1"" (`:26`); "Creating Additional Libraries" through the admin UI's "+" button with a name and a path (`:46-57`); "The new library will automatically begin scanning" (`:62`); a path must not "overlap with other libraries" (`:150`) | Task 9. The docs name only the UI; [servers] F2 created a library "through its admin API" by observation |
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
binaries, so each pass's sweep re-reads it.

## Readings this plan takes

Each is settled here, with its reason. The review may contest any of them.

- **R13. Every transcribed row lands `owed`.** An `active` row needs a session with a dubplate source (R12), and no
  dubplate response exists until slice 3. The sketch's "each `active` where a recorded session exercises it" cannot
  apply to any row. Each row's `slice` is the first slice its "Test owed" cell names; a cell that names only the
  harness or the contract tier takes the slice the assignment names for the row's endpoint; a row with no single
  endpoint (L-01 to L-05, L-07, L-08, L-14 to L-16, L-23) takes slice 3, the first slice that serves an envelope.
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
  golden diff, and a bump goes through the dependency sweep.
- **R19. The projection resolves an element's type by its position.** The XSD names `album`, `entry`, and `artist`
  with different types under different parents, and the OpenAPI additions differ by type. Keying on the element name,
  as the prototype does, drops `AlbumID3`'s additions from a `Child`-typed `album` and keeps `Child`'s, so every
  `getAlbumList` body would fail for the wrong reason.
- **R20. Errors are located by element ordinal.** xmllint reports a line, and an oracle's body may be one line. The
  projection writes one start tag per line and keeps each line's document-order ordinal in the original body;
  `xmlrule` reports ordinals and `grade` turns an ordinal into a path under its own tree rule. The path grammar then
  lives in one package.
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
  and `xmllint` at the pinned version on `PATH`. Each test that needs them fails naming the missing piece and the
  command that supplies it; none skips (CLAUDE.md, "A check either gates or does not exist").
- **Credentials.** No session body carries a credential parameter (Task 7), in addition to 2a's request rule.
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
   `ls /dev/kvm`.
6. **Sub-ceiling** assigned.
7. **AC power**, and the throughput baseline: the conductor's one full gate in the worktree, timed and quoted (2a's
   was about 586 s).

## Task 0: Sweep, pins, and the PR re-read (conductor)

Through the `dependency-upgrade` skill, since implementers carry no Skill tool.

**Outcomes.**
- **The sweep** over `go.mod` (kin-openapi now included), `web/package.json`, `tools/venv/requirements.txt`, and
  the pinned binaries, run in the main tree and landed on master before segment A, then the worktree rebased onto it
  (a fast-forward, since the branch has no commits yet). Or one line when nothing moves. The 2026-10-07 sweep is the
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

### Task 3: The scoped projection and the xmllint runner

**Files:** `internal/harness/subsonic/xmlrule/` (the projection, the runner, their `testdata/`). **Hook:** neutral.
**Model:** opus.

**Outcomes.**
- **The schemas.** From the XSD: each complexType's attribute and child-element names, following `xs:extension`;
  the type of each child element at its position; and the `Response` choice's payload elements. From the pinned
  OpenAPI document: each component schema's property names, following `$ref` and `allOf`. An XSD type maps to the
  OpenAPI schema of the same name, or through a short rename table in the package, each entry citing both names
  (`AlbumWithSongsID3` and `AlbumID3WithSongs` among them; the `Response` envelope and `Error` have their counterparts
  named there too). A type's additions are its OpenAPI property names minus its XSD names.
- **The projection** walks a body from the root, resolving each element's type by its position (R19), and drops from
  each element the attributes and child elements its type's additions name. A name declared at its position stays.
  A name in neither the XSD type nor the OpenAPI schema stays too, and is a **residual**: every residual in the body is
  reported, since xmllint stops at a parent's first unexpected child (quotes table). `status` and `version` on the
  envelope are never dropped.
- **Declared unvalidated.** A body whose payload element is not in the `Response` choice (for example
  `openSubsonicExtensions` and `lyricsList`, [contract] Q3) is reported as declared unvalidated, naming the element.
  The list comes from the XSD, never from a hand-kept table. It is never a pass.
- **The runner** checks `xmllint` once per process: resolved through `PATH`, its `--version` libxml version equal to
  the pinned `21504` (R18). It validates the projected document with `--noout --schema <cached XSD>` under
  `exec.CommandContext`. Exit 0 is valid and exit 3 is violations; any other exit is a tool failure carrying stderr.
- **Error location** (R20). The projected document has one start tag per line, and the runner keeps each line's
  ordinal in the original body's document order. Every xmllint error line becomes a violation at that ordinal, with
  the attribute name when the message names one and the message text after `Schemas validity error : `. An error
  whose line maps to no element is a violation at ordinal 0, never dropped. A residual that xmllint also reports on
  the same element and attribute yields one violation.
- **The result** of a body is one of: valid; violations (each an ordinal, an optional attribute, a message, and
  whether it came from the residual scan or xmllint); declared unvalidated; or a tool failure.

**States table.** XSD not cached or failing its sum: Task 2's errors. OpenAPI document lacking a schema the rename
table names: an error at construction naming it. `xmllint` absent: an error naming it. Wrong version: an error
naming both versions and the resolved path. Empty body: a violation, "empty body". Malformed XML: a violation carrying
the parse error. Payload outside the choice: declared unvalidated. Many errors: all of them, sorted by ordinal.
xmllint exiting other than 0 or 3: a tool failure.

**Test mandate and fixture state.** Unit proofs use small XSD and OpenAPI documents written for the tests under
`xmlrule/testdata/` (one element name with two types at two positions; an extension chain; a rename). Real-pin proofs
use the cached XSD and the committed OpenAPI bundle:
- The additions reproduce slice 1's measurement: 49 names over the envelope, `error`, `Child`, the album pair, the
  artist pair, and the playlist pair ([contract] Q3). The report names any count that differs, with the cause.
- A real seed `getAlbum` XML body from each oracle, with its residual attributes removed by hand, validates.
- A hand-written `getAlbumList` body whose `album` elements carry a `Child` addition (`mediaType`) validates; this
  is the proof the prototype's scope fails.
- A misspelled attribute on a `song` (`titel`) is one residual violation at that song's ordinal.
- An error on the second of two `song` elements in a one-line body lands on the second song's ordinal.
- `status` and `version` survive the projection of a real envelope.
- An `openSubsonicExtensions` body is declared unvalidated.
- A fake `xmllint` (a helper process re-executed from the test binary) reporting another version is refused, and one
  exiting 4 is a tool failure.

**Mutations:** key the scope on the element name (the `getAlbumList` proof fails); drop names globally (the envelope
proof fails); skip the residual scan (the misspelling proof passes and the test fails); read exit 3 as valid; locate
errors by projected line without the ordinal table (the second-song proof fails); accept any xmllint version; treat
declared unvalidated as valid.

## Segment B

### Task 4: The XML contract tier in the grader

**Files:** `internal/harness/subsonic/grade/` (`contract.go`, `grade.go`, `tree.go`, their tests), `cmd/harness/grade.go`,
the regenerated `grade/testdata/seed-report.golden`. **Hook:** neutral.

**Outcomes.**
- The contract tier runs every XML response of both graded sources through `xmlrule`, beside the JSON check. A
  violation is a `contract` finding with format `xml` and the violating source, at the path its ordinal maps to.
- **The ordinal maps to the tree's path** under `tree.go`'s rule: a name repeated among one element's children is an
  array and gets an index, a single child does not, and an attribute is a member. So a `contract` row's pattern names
  one path in both tiers, and `format: any` covers both.
- A declared-unvalidated body is no finding. The report counts them by endpoint and source in its header.
- A tool failure ends the grade with an error and no report.
- `grade.New` reads the XSD through Task 2's verified read; `harness grade` uses the cache by default.
- The seed golden regenerates. The task's report gives the XML contract findings by oracle and message class, and
  sets them beside the vendor residuals [contract] Q3 measured (Navidrome's `lastModified` on `artists` and
  `songCount` on `directory`; gonic's `title` and `album` attributes on ID3 `album`, `artistImageUrl` as an element,
  `albumCount` on a non-ID3 `artist`, the missing `lastModified` on `indexes`), naming each as present or absent.

**States table.** XSD absent or bad: `grade.New` fails with Task 2's error. An XML body that does not parse: one
`contract` finding (as for JSON) and the oracle tier's `parse` finding, both kept. Many violations: all, sorted.

**Test mandate and fixture state.** `TestGradeIdenticalBodiesReportOnlyContract` gains an XML twin: a hand-built
session whose two sources hold the seed's real Navidrome `getAlbum` XML body byte for byte, asserting zero oracle-tier
findings and exactly `xmlrule`'s violations for that body, once per source, at their tree paths. A path proof places
one violation on the second of two `song` elements and asserts the indexed path. A row proof shows a `contract` row
with format `xml` suppressing its finding and not the same path's JSON finding. **Mutations:** skip XML in the
contract tier (the twin fails); map an ordinal to a path by element name alone (the indexed-path proof fails); count
declared unvalidated as valid and drop it from the header; continue past a tool failure.

### Task 5: The oracle assignment and the area list

**Files:** `internal/harness/subsonic/data/assignment.json` (new), its loader and the grader's use of it in
`grade/`, `data/ledger.json` (three rows retired). **Hook:** neutral.

**Outcomes.**
- **Every endpoint in the pinned document has one entry** (87 on 2026-10-07): its oracle (`both`, meaning Navidrome
  primary and gonic secondary; `navidrome`; `gonic`; or `contract-only`), the slice that first serves it or
  `not-served-0.8` for an endpoint that answers error 0 at launch, and a reason with an evidence pointer. The values
  are the study record's section 3 table under the rulings: `getIndexes` and `getMusicDirectory` contract only, served
  by slice 3 (ruling C); `getMusicFolders` gonic, shape only; `getTranscodeDecision` and `getTranscodeStream` contract
  only and not served (ruling A); `getNowPlaying`, `scrobble`, `getUser`, `getShares` and the other share endpoints
  Navidrome; `download`, `getScanStatus`, `startScan`, `getLyrics`, `tokenInfo` contract only.
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
  names ruling C and R16.

**The assignment's one channel.** Only `data/assignment.json` sets an endpoint's oracle; nothing in the grader keeps a
second list.

**States table.** File absent: an error naming it. Empty: refused. Malformed or with an unknown member: the location.
Many bad (an endpoint twice, an unknown oracle value, an empty slice or reason, an album shape naming no manifest
album or one that lacks the shape): every one in one error.

**Test mandate and fixture state.** A gate test asserts the assignment's endpoint set equals the pinned document's
path set with `/rest/` cut, so a re-pin that adds an endpoint fails until it is assigned. A gate test checks each
album shape against the manifest by computation (more than one disc; a `Various Artists` album artist; a track with
two artist values; a name outside ASCII). The grading proof is a hand-built session with sources `dubplate`,
`navidrome`, and `gonic` and real seed bodies on `getAlbum` (both), `getIndexes` (contract only), and
`getMusicFolders` (gonic): grading dubplate against Navidrome reports `getAlbum`'s oracle findings, skips the other
two with their counts, and still reports `getIndexes`'s contract findings; grading dubplate against gonic compares
`getMusicFolders`. **Mutations:** consult the assignment for an oracle pair (the seed golden changes and
`TestSeedGrade` fails); skip the contract tier on an unassigned request; treat an unmapped endpoint as `both`; accept
an album shape the manifest lacks.

### Task 6: The ledger's machine form and the transcription

**Files:** `grade/ledger.go` and its tests, `data/ledger.json` (rewritten), `data/assignment.json` (its `agreements`
section), a transcription test in `grade/`. **Hook:** neutral.

**Outcomes.**
- **The row gains the record's columns** (R14): `divergence`, one of the record's ten classes (`declared: folder
  view`, `declared: apiKey path`, `by-design`, `contract-deviation`, `oracle-wrong`, `oracle-lenient`,
  `oracles-disagree`, `feature-absent`, `no-oracle`, `unspecified-shape`), required on every row; and `test`, the
  "Test owed" text, required on every row. The machine `class` is required on an `active` row and absent on an `owed`
  one. An `owed` row's `sources` is `dubplate` plus each oracle the record's "Oracle" column names (`Both` names two),
  so up to three; an `active` row still names exactly one pair.
- **The transcription.** `data/ledger.json` holds L-01 to L-54 as `owed` rows, with `slice` set per R13, except the
  retired rows below. The `agreements` section of the assignment holds A-01 to A-12 (R15), each with its area,
  shared behavior, evidence, test, and slice. The rows the rulings touch take the ruled text in the `dubplate` column:
  L-34 (not declared at 0.8, ruling A), L-51 (an empty `getIndexes` and error 70 from `getMusicDirectory`, graded
  against the contract and the spec, ruling C), and L-54 (the launch subset without `transcoding`, ruling A).
- **Retired, each with its ruling:** L-35 (ruling A), L-52 (ruling C), and 2a's `apikey-unoracled` row, which L-10
  replaces.
- **The report** counts `owed` rows and owed agreements by slice.

**The privileged columns.** `slice`, `divergence`, and the ruled `dubplate` texts have one channel each: a ledger
edit, proved by the transcription test. No code computes them at load time.

**States table.** As 2a's ledger, plus: a row with no `divergence` or an unknown one, an empty `test`, an `owed` row
carrying a `class`, an `active` row without one, an `owed` row with more than three sources or none of them
`dubplate`, and an agreement with an empty field: every one named in one error.

**Test mandate and fixture state.** The transcription test reads both Markdown tables from
`docs/superpowers/research/2026-10-06-slice-1-divergence-ledger.md`:
- **Round trip.** Parsing each table into rows and rendering the rows back yields the record's table bytes exactly,
  for the 54-row table and the 12-row agreement table. The proof needs the real tables, so an empty parse cannot pass:
  it asserts 54 and 12 rows.
- **Accounting.** Every id in the two tables is a committed ledger row, an agreement entry, or a retired id with its
  ruling; nothing else is in the ledger.
- **Fidelity.** Each committed row's transcribed columns equal the record's, except the three ruled `dubplate` texts,
  which the test names with their rulings.

The rendering lives in the test file, so no export exists without a caller. `TestCommittedLedger` and
`TestCommittedLedgerRowsAreLive` keep passing (every row names `dubplate`; no row is `active`). **Mutations:** map
`Both` to one oracle (the round trip fails); drop the `test` column (the round trip fails); let an `owed` row carry a
machine class; omit L-35 from the retired list (accounting fails); accept a fourth source.

## Segment C

### Task 7: Encoding quirks in the differ, and the credential guard

**Files:** `grade/differ.go`, `grade/tree.go`, their tests; the credential scan in `grade/`; its call in
`cmd/harness/recordoracle.go`; the regenerated seed golden. **Hook:** neutral.

**Outcomes.**
- **E1 and L-08.** The envelope's `serverVersion` and `type` compare by presence and type, never by value, since each
  names its own server ([contract] E1). A failed envelope's `error/message` compares the same way, and its `code` by
  value ([ledger] L-08). Each rule sits in one named table in the grade package beside the order table, citing its
  row.
- **E6.** JSON `null` against an empty container is one `type` finding; `null` against an absent member is one
  `presence` finding; `{}` against `{"playlist":[]}` is one `presence` finding at `playlist`; `[""]` against `[]` is
  one `presence` finding at index 0.
- **E5, E10, and the XML envelope.** Bodies that differ only in escaping (`&` against `&` in JSON; `&#34;`
  against `&quot;` in XML), in the presence of an XML declaration, or in a namespace prefix against the default
  namespace, yield zero findings. A singleton against an array stays one finding (2a).
- **L-07, the credential scan.** Every string leaf of every body from either graded source that holds a URL whose
  query carries `u`, `p`, `t`, `s`, or `apiKey` is one `contract` finding for that source at that leaf's path, naming
  the parameters and never their values. A query parameter whose name merely ends in one of them (`format`, `size`)
  never fires.
- **The body-side credential guard.** `harness record-oracle` runs the same scan over every body before it writes a
  session and refuses, naming the request and source and writing nothing, when any body carries a credential URL. A
  session therefore never commits an echoed credential (R21).
- The seed golden regenerates; the report gives the change in findings by kind (the `serverVersion` and `type` value
  findings leave).

**The credential channels.** A credential can reach a session through a request parameter (2a's writer and loader
refuse it) or through a response body (this guard refuses it). No third channel writes a session.

**States table.** A URL leaf that does not parse: no finding (it carries no query). A body with several credential
URLs: one finding per leaf. A refused recording: nothing written, the request and source named.

**Test mandate and fixture state.** `TestDifferPerturbations` gains one row per rule above, each on a real seed body
(the E6 rows need a seed body holding an empty container or a nullable member; the task names which), asserting
exactly the expected findings and kinds, or zero. The credential proof perturbs a real Navidrome `getArtist` body's
`artistImageUrl` (it carries `?size=600`) by appending `&u=admin&t=x&s=y`, asserting one finding naming `u`, `t`, and
`s` and no value; the unperturbed body yields none. The recorder proof uses 2a's fake stack serving such a body and
asserts the refusal and an empty output directory. **Mutations:** compare `serverVersion` by value (the golden and a
named row fail); match parameter names by suffix (the `size` proof fires); print a credential's value; let the
recorder write before scanning.

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
  the mixed empty containers. Each is named as present (with its path) or absent (with the recorded body), and the test
  asserts each one present. An absent prediction is recorded evidence, never a failure.

**States table.** As 2a's Task 5: the spec absent, empty, or malformed; an unresolvable key; an oracle failing
mid-run; a body carrying a credential (Task 7). Each refuses with nothing written.

**Test mandate and fixture state.** The committed session loads through the loader with every request answered by
both oracles. The golden test asserts byte equality, a second grade's identical bytes, and every test-ledger row's
count of at least one. **Mutations:** delete one response file (the loader refuses); add a test-ledger row matching
nothing (the stale-row assertion fails).

### Task 9: Two collections in the fixture and the stack

**Files:** `internal/harness/subsonic/data/fixture/collections/` (two new manifests), `fixture/`, `stack/`,
`cmd/harness/oracleup.go`, `cmd/harness/recordoracle.go` (the stack call only), their tests. **Hook:** neutral.

**Outcomes.**
- **Two small manifests** (three or four albums each) reuse the committed seed files and add no binary seed. Their
  paths and albums are disjoint, so no root nests in another (CLAUDE.md, "A collection is a store row ... with no two
  roots nested"). One album artist appears in both, so a filtered artist list differs from the unfiltered one in a way
  a test can name. The seed manifest is unchanged, byte for byte (R23), and still builds the library digest the seed
  session records.
- **The stack takes an ordered list of collections**, each a name and a manifest, built into one root each and mounted
  read-only with `:ro,z`. A one-collection list behaves exactly as today, so the seed recording path is unchanged.
  gonic receives every collection through `GONIC_MUSIC_PATH` with each name as its alias (quotes table). Navidrome's
  first collection is its `MusicFolder`; each later one is created through the request Navidrome's web UI sends
  when an admin adds a library (quotes table, the multi-library page). The implementer observes that request once in
  the browser's network log against a throwaway container and records its method, path, body shape, and response in
  the task's report, so the stack rests on a recorded observation and no one opens Navidrome's source.
- **Readiness** on both oracles: `getMusicFolders` lists exactly one folder per collection, and `getScanStatus`
  reports `scanning` false with `count` equal to the total tracks across collections.
- **A multi-collection library digest** is a sha256 over each collection's name and library digest, in list order. A
  one-collection list's digest is that library's own digest, so the seed's recorded digest still matches.
- `harness oracle-up` and `harness record-oracle` take the collection list by flag, defaulting to the seed library.
- **The live evidence:** the implementer runs `harness oracle-up` with both collections and quotes each oracle's
  `getMusicFolders` (password redacted).

**States table.** An empty collection list: refused. Two collections with one name: refused. Overlapping paths across
collections: refused, naming both. Navidrome's library creation answering an error: an error naming the status and
the collection. A folder count other than the collection count at readiness: an error naming the oracle and both
counts. A scan count short of the total: 2a's count error.

**Test mandate and fixture state.** Gate tests drive the stack against 2a's fake container runtime: the argv carries
both mounts with `:ro,z` and gonic's aliased `GONIC_MUSIC_PATH`; the fake answers the library creation and asserts its
body; readiness refuses a folder count of one. A fixture test builds both manifests twice and asserts identical bytes
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
  the collection manifest's album artists. No match, or more than one, refuses the run, naming the key and the oracle.
  The resolution lands in the session's resolution map like every other key.
- **The session** requests, in JSON against both oracles: `getMusicFolders` (and in XML), then the ten
  `musicFolderId` endpoints of [record] section 4 (`getIndexes`, `getArtists`, `getAlbumList`, `getAlbumList2`,
  `getRandomSongs`, `getSongsByGenre`, `getStarred`, `getStarred2`, `search2`, `search3`), each with the parameter
  absent, set to each collection, and set to the unknown literal `999`. Each request carries the parameters its
  endpoint needs and no more (an alphabetical `type` for the album lists, a `size` at least the total track count for
  `getRandomSongs`, an empty `query` for the searches, a genre the collection manifests carry for `getSongsByGenre`).
- **The grader takes the session's manifests.** `Config` names an ordered collection list, and the grader refuses a
  session whose recorded library digest differs from the list's combined digest, naming both. `harness grade` takes
  the list by flag, defaulting to the seed.
- A gate test grades the session, Navidrome against gonic, under a test-only ledger and pins the golden. It also
  asserts L-19's recorded behavior directly on the bodies: for `999`, Navidrome answers error 70 on all ten and gonic
  answers `ok` with an empty result; and A-04's: for each collection id, each oracle's filtered `getArtists` lists only
  that collection's artists.

**States table.** As Task 8, plus: an ambiguous or unmatched collection key, refused before anything is written; a
grade given the wrong manifests, refused with both digests.

**Test mandate and fixture state.** The resolution proof uses canned responses in which the two folders' artist sets
differ by the shared artist, and a refusal case in which they are equal. The digest-check proof grades the seed session
with the collection list and asserts the refusal. **Mutations:** resolve a collection by folder position (the canned
reversed order fails); skip the digest check; resolve on artist overlap instead of set equality (the shared-artist case
resolves twice).

### Task 11: The roster walk

**Files:** `docs/roster-walk.md` (new: the method and one script per client), the run record
`docs/superpowers/research/2026-10-07-slice-2b-roster-walk.md` (dated the day it runs), screenshots under the
record's own directory only if the record cites them. **Hook:** neutral. **Class:** `docs`. **Gate:**
`check.sh --reduced`.

**Outcomes.**
- **The method** is [base-url]'s, written once: each client under an isolated configuration home inside its Flatpak
  sandbox, keyring talk names denied, Rufin on a private Xvfb display, Feishin through its debugging port, and Geoff's
  Rufin profile hashed before and after with no change allowed. Credentials typed into a client come from
  `harness oracle-up`'s output and never reach the record.
- **One script per client**, each step stating the action, the expected observation, and what to record. The steps
  cover the flows [clients] section 2 names for that client: add the server as the OpenSubsonic type, log in, browse
  artists by ID3, open the multi-disc album and the compilation, play a track, show cover art, open the folder view
  (an empty or minimal view is the expected 0.8 answer under ruling C), search, star a track, and create a playlist.
- **Feishin 1.17.0 and Rufin 0.16.6** run once against each oracle `harness oracle-up` starts, and the record gives
  each step's pass or fail with its observation. A step that fails on an oracle is evidence about that oracle, recorded
  and not fixed.
- **Tempus 4.28.0** gets its script from source reading ([clients] section 2d), marked unproven. Its walk is deferred
  with the unblock condition: `ls /dev/kvm` succeeds after Geoff enables VT-x, then the SDK, emulator, and image
  install per [pins] "Client installs" with their manifest lines, then the script runs. The conductor checks
  `/dev/kvm` before this task; if it now exists, the Tempus walk runs here.
- **Amperfy 2.1.1** is listed for rung 16's attended gate (re-plan `:140-141`), with its script from [clients]
  section 2c.

**States table.** A client that will not start: recorded with its error, and the walk continues with the other.
Geoff's Rufin profile hash changed: the task stops and reports, since the isolation failed. An oracle-up that fails:
the task stops; it is 2a's verb, so this is a blocker.

**Acceptance.** The record names every step's outcome for both desktop clients on both oracles, the profile hashes
before and after, and the Tempus and Amperfy deferrals with their conditions. The `diff-reviewer` reads the record
against [clients] section 2's flow list for coverage.

## Close

### Task 12: Reviews, merge, and the ledgers

The conductor runs these in order. Every reviewer dispatch asks for correctness gaps ranked by consequence, not
polish, and a fold may refuse a finding with a one-line reason.

1. One `code-simplifier:code-simplifier` dispatch over the pass's changed Go, pasting
   `.claude/instructions/simplifier-brief.md` verbatim, then one `diff-reviewer` read of its diff.
2. One fresh-context vacuity read (`claude-opus-5-5`) over the tests of `xmlrule` and the tests Tasks 4 to 7 added in
   `grade`, the code this plan specifies only by outcome. Its single question: which test could pass while the
   behavior it names is broken? One fold task if it finds any.
3. `go vet -tags oracle ./cmd/harness/... ./internal/harness/...` in the worktree, once (CLAUDE.md, "Lanes").
4. The merge: the merge token; `git rebase master`; `check.sh --merge` in the worktree; `git merge --ff-only
   lane-3-slice-2` on master; the full gate on master; one `diff-reviewer` over the merged range if the rebase was not a
   fast-forward.
5. One `go-architecture-reader` per touched Go package: `internal/harness/subsonic/xmlrule`, `.../grade`,
   `.../fixture`, `.../stack`, `.../session` if Task 7 or 10 touched it, and `cmd/harness`. One fold task for accepted
   findings.
6. Queued repo edits in one batch on master: `CLAUDE.md` (the gate paragraph's prerequisites gain `make subsonic-xsd`
   and `xmllint` at libxml 2.15.4; the pinned binaries gain libxml2; the scoped gate's rendered-leg rule as Task 1
   left it); the streamer spec's revision note marking forks A to C ruled; BACKLOG #22 and #27 closed; HISTORY's 2b
   entry (what landed, what the gates and reviewers caught, what slice 3 would be wrong to rediscover, each golden's
   counts by kind, both budgets scored); STATUS (lane 3's row, the Tempus deferral, slice 3's readiness).
7. Worktree teardown, and the registry row says so.

**Acceptance.** Master is green with the merged work. The staged exported-unused leg reports no hit in `xmlrule` or
in the identifiers this pass added. `go list -deps ./cmd/dubplate-server | grep -E 'internal/(harness|opensubsonic)|getkin'`
prints nothing. `git diff master...lane-3-slice-2 -- go.mod go.sum internal/opensubsonic` is empty before the
rebase. Every committed session loads. No file under `internal/` carries the XSD's bytes. The oracle-tag vet is clean.

## Not verified at planning

- **xmllint's exit codes.** 3 for validation errors and 4 for an unreadable or malformed input were observed on
  libxml2 2.15.4; the installed man page (Fedora's 2.12) has no exit-status table. Task 0 quotes libxml2 2.15.4's own
  documentation or source comment for them if one exists; Task 3's runner treats anything but 0 and 3 as a tool
  failure, which holds either way.
- **Which `xmllint` `cairn-run-gate` resolves.** The gate's scope inherits `PATH`, which should give Homebrew's 2.15.4
  ahead of `/usr/bin`'s 2.12.10. Task 2's first gate shows it, and the version refusal names it if not.
- **The light lane's 3G cap.** Whether a scoped gate running the whole `go test ./...`, conform suite included, fits
  under it. Task 2's gate measures it; the fallback is the heavy lane.
- **Navidrome's library-creation request** and whether its first admin sees a new library without a grant. Task 9
  observes the first; its readiness check catches the second.
- **gonic's folder names under aliases** in `getMusicFolders`. Task 9's live evidence shows them; collection keys
  resolve by artist set, not by name, so either answer works.
- **The empty-states answers.** [contract] E6 predicts them at the probe fixture; the seed library is a different
  fixture, and Task 8's report reconciles each.
- **Whether every slice 1 projection count reproduces from the bundled file**, since the prototype read the
  specification repository's per-schema files. Task 3's count proof names any difference.
- **Whether Feishin's debugging port and a private Xvfb display still work** for this session's driver. Slice 1 used
  both on 2026-10-07; Task 11 states the method it used.
