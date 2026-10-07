# Rung 12, Slice 1 (The Study) Implementation Plan

Status: **draft revision 1, 2026-10-06.** It awaits its `spec-plan-review` lens and then Geoff's approval gate.
It carries one ruling for Geoff (Ruling 1, below). Every other fork it touches is settled here with its reason or
is routed, by name, to the slice that owns it.

**Spec:** the 0.8 streamer spec (`docs/superpowers/specs/2026-09-07-streamer-2-0-design.md`, revision 3, "the
spec"; it still says "2.0"). Its ladder entry 1 (`:724-731`) is this pass's scope list, and its "Open items"
(`:917-955`) name the questions owed to it.
**Re-plan:** `docs/superpowers/specs/2026-10-05-rungs-12-16-replan-design.md`, revision 2.1, approved by Geoff
2026-10-05 ("the re-plan"). Ruling 5 (`:48-54`) opens this slice now as a document lane, ahead of rung 12's place on
the ladder.
**Parallel lanes:** lane 1 (main tree, master; next is 6d revision 2.3) and slice 2 (the harness, a code lane in its
own worktree, plan authored just in time beside this one). This lane shares the main tree with lane 1.

**Pass class:** `docs`. The pass writes research records under `docs/superpowers/research/` and one spec revision.
It writes no Go, TypeScript, Svelte, Python, shell, unit, or configuration, so no code class fits. The class table
(`pass-core`) gives `docs` "the docs tier" as its gate and "the register chain" as its review; this repo's docs tier
is the full gate (see "Per-task gate"), and the register chain here is narrowed as "Review" says.
**Executor:** a `general-purpose` agent at `model: sonnet` for the evidence tasks (1 to 6), because they need
network reads, `git clone`, and `podman`, and `dubplate-implementer`'s test-first, commit-per-task shape does not fit
a staged research record. Task 7 (the synthesis) runs at `model: claude-opus-5-5`, effort `high`, because it makes
the spec-level readings. Each landing step is a `general-purpose` agent at `model: sonnet`.
**Per-task gate:** none inside an evidence task, which commits nothing. Each landing step (A-land, B-land) runs
`cairn-run-gate 'bash /var/home/glw907/Projects/dubplate/scripts/check.sh'`, re-issued on exit 75 until it prints
`gate exit:`, before its one commit. The full form follows the re-plan document pass's precedent
(`plans/2026-10-05-replan-document-pass.md`, "Per-task gate"): `check.sh --reduced` is "never the final gate on
master", and every commit here lands on master. The gate is not browserless, so `CAIRN_GATE_LANE=light` is not
allowed. The leg this lane can actually redden is gitleaks over the worktree, which reads every new record.
**Review:** `diff-reviewer` at Sonnet once per task file at its landing, against that task's acceptance criteria,
blocking only on an unmet outcome, a factual error against a cited source, or a breach of the clean-room rules below.
Task 8 adds one fresh-context fact read at `claude-opus-5-5`. The spec revision takes one `prose-voice-reviewer`
read (the spec prose rule: once per revision, facts and contradictions first). Research records take no register
review. No `code-simplifier` runs (a `docs` pass never takes one).
**Ceiling:** 6.0M execution tokens, flag 4.8M (80%). Planning and the plan-review lens are recorded outside it. The
spec bands the study M (`:724`). The weight sits in source reading: two servers, four clients, Melodee, and the
contract repository, at pinned commits. The runaway guard is armed per research dispatch at the memory file's ~2MB
mark (`pass-execute-dispatch-mechanics.md`).
**Checkpoint interval:** each segment boundary and the close. STATUS is written at each.
**Segments:** A (Task 0, then Tasks 1 to 6 in parallel, then A-land) and B (Task 7, Task 8, then Task 9, the close,
which carries B-land). Segment A holds seven tasks, past the three-to-four default. The override is the disjoint
Files seam the workstation rule names: Tasks 1 to 6 each write one new file and share nothing, and A-land is the one
green commit that bounds them.
**Tasks:** ten (Task 0, six evidence tasks, the synthesis, the fact read, and the close).
**Execution mode:** segment A's evidence tasks dispatch through the Agent tool in parallel, one message, after
Task 0 lands its pins. The `pass-execute` runner is built around an implementer that commits each task under a gate,
and these tasks stage outside the repo and commit nothing, so the runner's chain does not fit. This is a departure
from the workstation rule that independent tasks run through `pass-execute`, and the plan-review lens should grade
it. Segment B is three sequential dispatches.
**Dependency sweep:** this lane edits no dependency manifest, so the pass-opening `dependency-upgrade` sweep
(`CLAUDE.md`, "Current dependencies") does not apply. A sweep here would also collide on `go.mod` with lane 1 and
slice 2, whose own sweeps own it. The record carries one line saying so. The rule's intent does apply to what the
study reads: Task 0 pins each study subject at its newest production release, so no finding describes a version
the harness will not run.
**Operator-supplied values:** delta zero. The lane adds no key, threshold, or setup step (re-plan `:137-138`).
**0.8 forward hooks:** every task is neutral. The study builds nothing. A hook the study discovers is routed into
the spec's forward-hooks list at Task 9 with the rung that first consumes it (`CLAUDE.md`, "Build for 0.8"), and is
never built here.

Sources, cited at `0cb33ae` on master:

- The spec: ladder entry 1 (`:724-731`), entry 2 (`:732-736`), "License" (`:341-347`), "The protocol layer is a
  separable module" and its clean-room record (`:298-328`), "Evidence" (`:258-289`), "Catalog projection"
  (`:374-376`), "Streaming and transcoding" (`:407-445`), "API surface" (`:640-673`), "Prior art" (`:856-891`), and
  "Open items" (`:917-955`). Revision 3's items 10, 12, and 13 (`:43-50`).
- The re-plan: ruling 5 (`:48-54`), rung 12 (`:113-141`), "Preserving the archive (now)" (`:180-194`), document
  change 2's open-items clause (`:384-386`), and "Where the carried items land" (`:410-445`), whose row at `:427`
  routes the API base URL here.
- The re-plan's fold record (`research/2026-10-05-rungs-12-16-replan-fold.md:38,127`) and its mechanics review M15
  (`research/2026-10-05-rungs-12-16-replan-review-mechanics.md:284-301`), which frame the base-URL measurement.
- The streamer prior-art survey (`research/2026-09-25-streamer-prior-art-survey.md`): "Copyleft servers: study for
  behavior only" (`:131-151`), "Conformance suites and test drivers" (`:153-173`), the open-PR row (`:261`), and "Not
  verified" (`:263-280`).
- The client roster audit (`research/2026-10-05-client-roster-audit.md`), which rules the roster and models the
  pinned-commit source reading this plan asks for.
- `docs/STATUS.md` "Open items" and "Owed" (`:39`, `:46-48`).
- `cmd/dubplate-server/serve.go:168-177`, the `DUBPLATE_SERVE_PREFIX` threshold (default `/dubplate`), which is the
  sub-path the base-URL task measures.

---

## The clean-room rules

The spec governs this lane's most important constraint. Its "License" section (`:343-346`):

> Navidrome and gonic are GPLv3. The study rung reads them for ideas and for lessons, positive and negative, and
> records what it learns in its own document (Geoff, 2026-09-07). Implementers build from that record and from the
> OpenSubsonic contract, and never copy or transliterate GPL code. Melodee is MIT, so its code may be reused as well
> as read.

The survey adds (`:133`): "Each line cites the LICENSE the GitHub API reports. None may be copied or transliterated."
The protocol layer's clean-room clause (`:316-318`) makes this record the citation source: "Every behavior the layer
implements cites the OpenSubsonic specification, its Apache-2.0 documentation, or an observation recorded by the
harness, never GPL source. The study rung's record is the citation source."

So every record in this lane holds to these rules, and each task's acceptance checks them:

1. **GPL and AGPL source is cited, never reproduced.** A claim from Navidrome, gonic, Ampache, or a GPL client
   cites `repo@commit:path:line` and states the behavior in dubplate's own words. No code, no pseudo-code that
   follows the source's structure, and no identifier list copied as a schema. No fenced block in any record may hold
   text taken from a copyleft repository.
2. **Permissive and documentation sources may be quoted.** The OpenSubsonic specification and OpenAPI file
   (Apache-2.0), Navidrome's documentation repository `navidrome/website` (Apache-2.0, survey `:149-151`), and Melodee
   (MIT) may be quoted briefly with their citation. Melodee code may be reused later; this lane still records ideas,
   not code.
3. **Observation beats reading.** Where a local oracle can show a behavior, the record states the observed request
   and response shape in its own words, with the oracle version and the fixture state.
4. **Every claim carries its source.** A behavior stated from memory is a finding against the task. A claim that
   could not be confirmed is marked unconfirmed and says what would confirm it.

## Settled here, with reasons

These have a clearly correct answer, so the plan takes them and records why.

1. **A local Navidrome and gonic run is sanctioned, on the workstation only.** The re-plan (`:127-128`) keeps "the
   pinned Navidrome and gonic containers ... in the harness as workstation-only behavior oracles, never on the box,
   under the streamer spec's GPL behavior-only rule", and the spec's revision 3 item 12 (`:48-49`) says the same.
   Running a GPL server and observing it copies nothing. The box is out of bounds: Navidrome is retired from it and
   it runs only dubplate-server. Evidence therefore comes from three channels, and each record names which one backs
   each claim: published documentation, pinned source reading, and observation of a local `podman` container bound
   to `127.0.0.1`.
2. **The study does not open the retirement archive.** The re-plan gives the archive one reader, rung 16's import
   "against the verified off-box copy of the archive's database" (`:157-159`), and says it "holds members'
   addresses and listening history" (`:193-194`). Nothing in slice 1's scope needs real state. The local oracles
   serve a scratch fixture library (Task 0). If a study question turns out to need the real database (the
   Navidrome schema version the import must read is the likeliest), the record routes it to slice 8 as an open row,
   and no agent in this lane decrypts the `.age` file.
3. **The fixture library is scratch, built from synthetic audio.** It is built under the session scratchpad from the
   repo's committed FLAC fixtures or `ffmpeg`-generated tones with tags set, and never from the group's music. Its
   recipe is recorded in Task 0's file so any later task rebuilds it. Slice 2 builds the real fixture library;
   this one is a probe.
4. **"Cited to a version and a harness test" reads as "a version now, a test row owed".** The spec (`:724-725`)
   wrote that sentence when the harness would follow the study. Ruling 5 runs them side by side, so no harness test
   exists to cite yet. Each behavior claim cites its version now, and the pre-seeded divergence ledger (Task 7)
   carries a "test owed" column naming the harness test or slice that proves it. The claim is fully cited the day
   that test lands.
5. **Slice 1 proposes the pins and slice 2 adopts them, whichever lands first wins.** Task 0 picks the newest
   production release of Navidrome, gonic, and the OpenSubsonic specification commit. If slice 2 has already pinned
   any of them when Task 0 runs, Task 0 adopts slice 2's pin and records a newer release only as a note. One pin per
   subject, never two.
6. **The pre-seeded divergence ledger is a Markdown table with fixed columns.** Slice 2 builds the ledger's machine
   form, and its plan transcribes this table into it. The columns are named in Task 7 so the transcription is
   mechanical.
7. **Slice 1 writes the spec's answers as revision 3.1.** The re-plan document pass named slice 1 as the spec's only
   other writer for the open items owed to the study rung (`plans/2026-10-05-replan-document-pass.md`, Task 2's
   row). Task 9 marks each item answered with a pointer to the record, fills the prior-art cells, and leaves any
   owner fork marked "answered pending ruling" until Geoff rules.
8. **Out of scope, with owners.** "Which credential a stream authenticates with before slice 6" is owed to the slice 5
   plan (`:949-950`). The app-password key's name is slice 6's (`:927`). The report path is slice 7's and the admin fix
   path rung 13's (`:929-933`). The base-URL decision is slice 3's: this lane measures, and slice 3 decides with a
   build-spec erratum and brings it to Geoff's gate (re-plan `:427`; fold record `:127`).

## Rulings

1. **May the base-URL measurement drive real clients on the workstation, starting now?** The yes form installs
   Feishin and the Android emulator with Tempus (recorded in their tier manifests the same session, with popups,
   sounds, and badges off at install, per the calm-desktop rule) and walks them, plus Rufin, against a local
   Navidrome served under `/dubplate`. Rufin is Geoff's daily client, so it runs under an isolated configuration
   home, and no server entry is added to his own profile. If Rufin cannot be isolated, it is measured by source
   alone and the record says so. Amperfy cannot be driven from Linux and is measured by source alone; the re-plan
   walks it at rung 16's attended gate (`:140-141`). The no form measures all four clients by source reading.
   **Recommend yes.** Source reading misses runtime URL handling: trailing-slash joins, a path dropped by a URL
   library, a stream or cover URL built by a second code path. Slice 2's roster walk needs the same installs, so
   the marginal cost is the walk itself.

### Rulings the record will bring at the close

These cannot be answered until the evidence exists. Task 7 brings each with a recommendation, and Task 9 batches them
into one checkpoint question with the spec revision.

- Whether 0.8 answers the `transcoding` extension (`:412-419`, `:941-942`). This is scope, so it goes to Geoff.
- Whether folder browsing ships at launch (`:925-926`). Also scope.
- The protocol module's license (`:327-328`, `:947-948`). Geoff's 2026-09-07 reasons for MIT over Apache-2.0
  (`:346-347`) likely carry over. The open fact is whether the module would carry the Apache-2.0 OpenAPI file and
  what NOTICE text, if any, the OpenSubsonic repository requires.

Every other open item is a method or architecture reading the conductor settles from the evidence (`CLAUDE.md`,
"Process proportionality"), recorded with its reason.

---

## Planning-miss prevention, adapted to a study

The repo's four checks (`CLAUDE.md`, "Planning-miss prevention") apply in this form:

1. **Quoted, never recalled.** Every behavior a record relies on is quoted from a document or cited to source at a
   pinned commit. This plan states no server or client behavior as fact. Where it names a likely behavior, it is a
   question for a task.
2. **Every question reaches a named outcome.** Each question a task owns ends in one of three states: answered with
   evidence; partly answered, with the missing piece named; or not answerable in this lane, routed to a named slice.
   A question that silently drops out of a record is a finding.
3. **Privileged inputs name their channels.** Two questions touch one: CORS (which origins and whether credentials
   ever cross) and the unauthenticated `getOpenSubsonicExtensions` carve-out. The record names every channel a
   client uses to send credentials (query, header, body) for each.
4. **Each observation names its fixture state.** An observation of a transcode says whether the cache entry was
   filling or complete, of a browse whether the library was scanned and populated, of a client whether it was logged
   in. An observation without its state proves nothing and is a finding.

---

## Tasks

Every evidence task (1 to 6) writes its draft to the lane's staging directory, `<conductor scratchpad>/slice-1/`,
which the conductor names in each dispatch, and commits nothing. A-land moves the drafts into
`docs/superpowers/research/` and commits them. Nothing warm sits in the shared tree during a segment, so lane 1's
live-executor check never meets this lane's files.

Every record opens with its date, the pins it read (from Task 0's file), and the question list it owns, and ends
with a "Routed" section listing anything it hands to another slice.

### Task 0: Pin the evidence base and stand up the probe oracles

**Deliverable:** `docs/superpowers/research/2026-10-06-slice-1-evidence-pins.md`.
**Settles for:** every later task (the pins they cite); slice 2 (the oracle image pins and the specification commit,
under settled item 5).

**Outcome.** One file names, with a date and a source for each: the newest production release of Navidrome and
gonic with their container image digests; the OpenSubsonic specification repository's commit, and the extension
count at that commit; the newest release tag and its commit for each roster client (Feishin, Rufin, Amperfy, Tempus)
and for Melodee; the `navidrome/website` commit; and each repository's LICENSE file URL as the GitHub API reports
it. It records the fixture recipe (settled item 3) and the run recipe for both oracles: `podman`, `127.0.0.1` only,
the fixture mounted read-only, and Navidrome configured to serve under `/dubplate` for Task 3. The run recipe
quotes Navidrome's documented option for a base path from `navidrome/website`, and gonic's from its README, at the
pinned commits, before relying on either. If Ruling 1 is yes, the file records the client installs and their
manifest entries.

**Constraints.** A release is a tagged release the project marks as production; a pre-release, nightly, or `-rc`
tag is skipped. If slice 2 has pinned a subject already, adopt its pin (settled item 5). No container binds a public
interface. The box is never contacted.

**Acceptance.**

- Every pin carries its source URL and the date it was read.
- Both oracles started from the recorded recipe, answered `ping` at their own root and Navidrome at `/dubplate`,
  and listed the fixture's albums. The record states each response's status, not just "worked".
- The base-path option is quoted with its citation, or the record says it does not exist and Task 3 is re-scoped
  before dispatch.
- The dependency-sweep line and the operator-delta line are present.
- Prior art consulted: Navidrome and gonic (adopted as oracles, spec `:868-869`), the OpenSubsonic specification
  (adopted as the contract, `:866`). Nothing adopted or declined anew.

### Task 1: The contract, the encoding quirks, and the XML rule

**Deliverable:** `docs/superpowers/research/2026-10-06-slice-1-contract.md`.
**Settles for:** slice 2 (the XML validation rule its harness implements; the quirk list its tests cover); slice 3
(the response layer's encoding); the protocol module (the wire types it hand-writes).

**Outcome.** The record answers, from the specification at the pinned commit:

- The extension list, each with its endpoints, its version, and what a server must do to declare it. This is the
  input to Task 7's launch extension subset (`:672-673`).
- The encoding quirks the prior-art table asks the study to enumerate (`:883`): attribute-heavy XML, the JSON
  wrapper, singleton-versus-array in JSON, and JSONP. For each, what the specification says and what Navidrome and
  gonic do, observed on the probe oracles. Any other quirk found is added.
- The XML validation rule (`:266-271`, `:939-940`): the evidence for a legacy XML form validated against the
  official XSD and for a derived XSD that admits the OpenSubsonic attributes, with what each costs the harness and
  what each proves. Ampache's method is studied for the idea only (AGPL, survey `:137`, `:161`).
- The state of OpenSubsonic PRs #162, #237, #258, and #261 and any newer PR touching music folders, collections, or
  library browsing (`:951-955`; survey `:261`). For #237, what its definition of "collection" would mean for
  dubplate's word.
- The error-code list and the text of error 0, 50, and 70 as the specification states them.

**Acceptance.** Every specification claim quotes the file and line at the pinned commit. Every oracle claim names
version and fixture state. The XML rule has a recommended reading with its reason, for Task 7 to settle. Prior art
consulted: the OpenSubsonic specification and the OpenAPI and XSD row (adopted, `:866-867`); Ampache (study the
method only, survey `:137`, `:161`). Neutral on hooks.

### Task 2: The client roster's demands

**Deliverable:** `docs/superpowers/research/2026-10-06-slice-1-client-demands.md`.
**Settles for:** Task 7's empty-success set and launch extension subset; slice 3 (browse views, `musicFolderId`);
slice 4 (whether any client sends `Accept: image/webp` on `getCoverArt`, survey "Not verified" `:267-268`); slice 5
(the AAC container and the `transcoding` extension).

**Outcome.** For each roster client at its pinned commit, read in source the way the roster audit did, the record
states:

- Every endpoint the client calls, and for each whether the client fails, degrades, or carries on when the server
  answers error 0. The endpoints whose failure breaks a client's normal use are the empty-success candidates
  (`:670-672`).
- Every OpenSubsonic extension the client checks for, and what it does with and without it, the `transcoding`
  extension above all (`:412-419`).
- Whether it browses by ID3, by folder, or both, and whether it passes `musicFolderId`, and on which endpoints. This
  is evidence for the folder-browsing item (`:925-926`).
- Which stream formats and containers its player accepts, read from the player library it uses, as evidence for the
  AAC container item (fragmented MP4 or ADTS, `:427-431`, `:945-946`).
- Whether it runs in a browser context that needs CORS (`:650-655`), and how it sends credentials (query, header,
  or body; planning-miss check 3).
- Whether it sends `Accept: image/webp` on `getCoverArt`.

The Symfonium compatibility notes (`:885`) and Melodee's published compatibility matrix are read as secondary inputs,
cited, never as a substitute for source.

**Acceptance.** One table per question, a row per client, each cell cited to `repo@commit:path:line` or marked
"not found in source" with what was searched. Clean-room rule 1 holds: the clients are GPL, so no client code
appears in the record. Prior art consulted: Symfonium notes (adopted as a study input, `:885`); Feishin's type
schemas (study for required fields, survey `:165`). The GPL Go Subsonic clients stay declined (survey `:172`) and are
not used as probes. Neutral on hooks.

### Task 3: The API base URL against a sub-path

**Deliverable:** `docs/superpowers/research/2026-10-06-slice-1-base-url.md`.
**Settles for:** slice 3, which decides between the sub-prefix mount and `/rest` at the root and writes the build-spec
erratum (re-plan `:427`). This task measures and recommends nothing binding.

**Outcome.** For each roster client, a verdict on whether it works with a server URL that carries a path
(`http://127.0.0.1:<port>/dubplate`), each verdict one of works, breaks (with the failing request), or unmeasured
(with the reason). The record covers, per client:

- How the client joins its configured URL with `/rest/<endpoint>`: with and without a trailing slash, and with the
  `.view` suffix if it uses one.
- Every second code path that builds a URL (stream, download, cover art, share), since a client can get the API
  right and a stream URL wrong.
- Any open or closed issue in the client's tracker about sub-path or reverse-proxy base URLs, cited.
- Under Ruling 1's yes: the live walk against Navidrome at `/dubplate` (login, browse one album, play one track,
  load its cover), with the requests the oracle logged. Under no, source reading only.

It also records what Navidrome and gonic document for serving under a sub-path, quoted from their documentation, so
slice 3 can see how the field treats the question.

**Acceptance.** Four client rows, each with a verdict and evidence; a live verdict states the fixture state (scanned,
logged in) and the oracle's logged request paths. Rufin's live walk used an isolated configuration home or did not
run. No setting in Geoff's own client profiles changed. Clients installed under Ruling 1 are in their tier
manifests. Prior art consulted: Navidrome documentation (Apache-2.0, quotable). Neutral on hooks.

### Task 4: Where the specification is silent, what Navidrome and gonic do

**Deliverable:** `docs/superpowers/research/2026-10-06-slice-1-server-behavior.md`. This is the spec's "written
record" (`:724-726`) in its server half.
**Settles for:** Task 7's per-endpoint oracle assignment and divergence ledger; slices 3, 5, 6, and 7 (each section
names its slice); the protocol module's clean-room citations (`:316-318`).

**Outcome.** For each question below, the record states what each oracle does, by observation on the probe where the
probe can show it and by documentation or cited source otherwise, and then the lesson for dubplate in its own words:

- How each answers an endpoint it does not implement, and which endpoints it answers with an empty success. This is
  the server side of the empty-success set.
- Which endpoints honor `musicFolderId` (`:374-376`).
- Range on a transcode: complete versus filling entries, `Accept-Ranges`, `estimateContentLength`, and `timeOffset`
  (`:436-443`; S finding 4). The probe observes both cache states.
- How Navidrome implements the `transcoding` extension, and how a posted `ClientInfo` meets its own transcoding
  configuration (`:412-419`).
- What CORS headers each sends on a simple request and on a preflight, and whether credentials ever cross (`:650-655`).
- `search3` with an empty query and with a query under three characters; paging over the whole library.
- The play queue's edges and `scrobble`'s `submission` semantics, checked against the spec's statements at
  `:563-569`.
- The folder view: how Navidrome synthesizes its tree and how gonic serves real paths (`:276-280`).
- The smart playlist `.nsp` format, quoted from `navidrome/website`'s smart-playlist page (survey `:150-151`). This
  answers the `.nsp` item (`:934`).
- The area list of Navidrome's `server/subsonic/e2e` suite and gonic's golden fixtures (survey `:135-136`, `:163`),
  as a list of what their authors chose to test.

The ideas and lessons half asks, for each area, what each server's design teaches dubplate positively and
negatively, stated as dubplate design notes.

**Acceptance.** Every claim names its channel (documentation, source, observation), its version, and for an
observation its fixture state. Clean-room rule 1 holds throughout; the fact read in Task 8 samples this file first.
Each section ends with the slice it feeds. Prior art consulted: Navidrome and gonic (oracles, read for ideas, never
transliterated); `navidrome/website` (quotable); `djherbis/stream` is named only where Navidrome's filling-entry
behavior rests on it, and its adoption stays slice 5's race test (`:445-447`). Neutral on hooks.

### Task 5: Melodee's intake pipeline

**Deliverable:** `docs/superpowers/research/2026-10-06-slice-1-melodee-intake.md`.
**Settles for:** the spec's one study question (`:725-728`); its ideas feed rung 7 (intake), rung 9 (admission),
and rung 13 (the decision queue), each named per idea.

**Outcome.** The record compares Melodee's ingest pipeline (staging, validation, admission into the library) with
dubplate's intake-as-the-one-door model, read at the pinned Melodee release. It states what Melodee does at each
stage, cited, and for each a lesson for dubplate, positive or negative, with the rung it bears on. It also fills the
prior-art table's Melodee quality cell (`:870`, "study rung") against the house bar, with the evidence.

**Constraints.** Melodee is MIT, so brief quotation is allowed. The record still states ideas, and any code worth
reusing is named by path for the rung that would reuse it, never copied here.

**Acceptance.** Each stage cited to `melodee@commit:path:line`; each lesson names its rung; the quality cell carries
evidence. Prior art consulted: Melodee (adopted as readable and reusable prior art, `:870`). Neutral on hooks; an
idea that implies a hook is listed under "Routed" for Task 9.

### Task 6: The prior-art table's open cells

**Deliverable:** `docs/superpowers/research/2026-10-06-slice-1-prior-art-cells.md`.
**Settles for:** the spec's "Prior art" table (`:856-891`), whose open cells "are the study rung's deliverable,
scored on CLAUDE.md's four criteria with evidence per cell" (`:858-860`).

**Outcome.** For every cell that reads "URL from study rung" or "study rung" (Navidrome, gonic, Melodee, wazero,
`modernc.org/sqlite`, `ncruces/go-sqlite3`, `mewkiz/flac`, Picard's tables, `x/image` and `gen2brain/webp`,
`x/time/rate`), the LICENSE URL as the GitHub API or the project's own source reports it. For each row whose version
or adoption figure the spec dates, the current figure with its date, so Task 9 can refresh the row under the
current-dependencies rule. A row whose verdict the new evidence would change is flagged for Task 7, never changed
here.

**Acceptance.** One row per entry, each license cell a URL that resolves, each figure dated and sourced. The survey's
already-verified URLs (`:135-136`) are re-checked, not copied. Prior art consulted: every table row listed.
Neutral on hooks.

### A-land: Land segment A

**Outcome.** The seven drafts move from the staging directory into `docs/superpowers/research/`, one
`diff-reviewer` per file reads it against its task's acceptance, any `fix` is folded once by the task's own agent,
and one commit lands them on master after the full gate is green.

**Constraints.** Runs at a lane 1 segment boundary: lane 1's last task is committed and `git status` shows no warm
change outside these files (the document pass's "Sharing the main tree" rules apply as written). Stage the seven
files by name, never `git add -A`. A red gate whose failing leg sits outside these files is not this lane's failure:
the agent reports and stops without committing. STATUS is written at the checkpoint.

### Task 7: The synthesis

**Deliverables:** `docs/superpowers/research/2026-10-06-slice-1-study-record.md` (the study record) and
`docs/superpowers/research/2026-10-06-slice-1-divergence-ledger.md` (the pre-seeded ledger).
**Model:** `claude-opus-5-5`, effort `high`.
**Settles for:** every later slice. This is the file the spec, slice 2, and each slice plan cite.

**Outcome.** The study record is the index and the decisions. For each item the spec owes the study rung, it gives
the answer, the evidence (by pointer into the segment A records), and its disposition: settled by the conductor with
the reason, or an owner fork with a recommendation. The items:

- The empty-success endpoint set, measured against the roster (`:670-672`).
- The launch extension subset, measured against the extension list (`:672-673`).
- The per-endpoint oracle assignment: contract only, Navidrome, gonic, or the roster, per endpoint the spec's launch
  surface names (`:657-661`).
- The `musicFolderId` endpoint list (`:375`).
- The XML validation rule (`:939-940`).
- The `transcoding` extension (`:941-942`): owner fork.
- Range on a filling transcode (`:943-944`).
- The CORS policy (`:945`).
- The AAC container (`:946`).
- The protocol module's license and NOTICE (`:947-948`): owner fork.
- The word "collection" and the PR watch (`:951-955`).
- Folder browsing at launch (`:925-926`): owner fork.
- The `.nsp` format (`:934`).
- The API base URL: the measurement summary only, handed to slice 3.
- Any new forward hook, with the rung that first consumes it.

The divergence ledger has one row per known divergence, with these columns: id, endpoint or area, divergence class
(the spec's two declared classes, `:285-287`, or a new one), oracle, the oracle's behavior, dubplate's intended
behavior, evidence pointer, and the test owed (the harness test or slice that proves it, under settled item 4).

**Acceptance.** Every item above has a row with one of the three outcome states (planning-miss check 2). Every
settled disposition gives its reason in one or two sentences. Every owner fork names both answers' consequences and
a recommendation. No claim in the record lacks a pointer into a segment A record or a primary source. Neutral on
hooks; the record lists any it routes.

### Task 8: The fact read

**Outcome.** One fresh-context `claude-opus-5-5` reader, with no plan in context, reads the study record and the
ledger, follows a sample of their pointers to the segment A records and from there to the primary sources, and
returns correctness gaps ranked by consequence, never polish. It checks, in order: a claim its source does not
support; a breach of the clean-room rules (any text in a record that tracks a GPL file's structure); an observation
with no fixture state; an item with no outcome state. The server-behavior record is sampled first.

**Constraints.** Read-only. The fold (inside Task 9) may refuse a finding whose fix costs more than the risk it
removes, with a one-line reason (`CLAUDE.md`, "Findings are ranked, and a fold may refuse").

### Task 9: The close (B-land)

**Outcome.**

- The fact read's findings are folded into Task 7's two files, with the fold record appended to the study record.
- The spec moves to revision 3.1. Each open item answered here is marked answered with a pointer to the study
  record. The owner forks are marked "answered pending ruling". The prior-art cells take Task 6's URLs and dated
  figures, and a row whose verdict Task 7 changed takes the change with its reason. New hooks join the
  forward-hooks list, numbered after 21. The revision-3.1 changelog lists each change.
- One `prose-voice-reviewer` read of the spec revision, facts and contradictions first, folded.
- Each landed file gets its `diff-reviewer` read.
- STATUS: the slice 1 lane row reads closed, the "Slice 1" and "Owed" carries are removed, and the owner question
  bundles Ruling 1's outcome (if it was not ruled at approval) with the forks the record brings. HISTORY takes the
  pass entry: what landed, what the fact read caught, and what a later slice would be wrong to rediscover.
- One commit on master after the full gate, under A-land's constraints.

**Acceptance.** `grep -n "owed to the study rung" docs/superpowers/specs/2026-09-07-streamer-2-0-design.md` returns
only lines that also say answered, or lines the revision-3.1 changelog explains. Every "URL from study rung" cell is
filled. STATUS stays at or under 60 lines. The tokens spent are recorded against the ceiling in HISTORY.

## Contested surfaces

| Surface | Other writer | Rule |
|---|---|---|
| The streamer spec | Slice 2's plan may cite or amend its "Evidence" section | Task 9 edits it only after reading its current state on master; a slice 2 edit landed first is merged, never overwritten |
| Oracle image pins and the specification commit | Slice 2 | Settled item 5: the first pin wins |
| `docs/STATUS.md`, `docs/HISTORY.md` | Lane 1's close | One ledger writer at a time; Task 9 waits for any in-flight lane 1 close |
| The heavy gate lock | Lane 1 and slice 2 | `cairn-run-gate` serializes; A-land and B-land each run one gate |
| Workstation ports and `podman` | Slice 2's harness containers | The probe binds its own ports, recorded in Task 0, and stops its containers at A-land |
