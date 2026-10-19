# Rung 12 slice 2b plan review: domain risk

Target: `docs/superpowers/plans/2026-10-07-rung-12-slice-2b-oracle-harness.md` at `0ae2b54` (revision 1), worktree
`lane-3-slice-2`. Lens: what goes wrong for slice 3 onward, which grades dubplate's real streamer with this machinery.
I report correctness gaps ranked by consequence, then over-ceremony ranked by cost. Line numbers are the plan's
unless a file is named.

**Verdict.** The plan is sound in its main lines. Owed rows match nothing (`grade/ledger.go` `Apply` skips them), so
landing all 54 rows `owed` (R13) excuses no divergence. The gate prerequisites fail loudly. The XSD stays out of the
repo, and the gate opens no connection. Three majors remain: the XML tier is blind to the endpoint, the transcription
tests freeze the ledger against slice 3's first activation, and the 87-entry assignment has no check in the direction
that removes oracle coverage. The rest are minor.

Counts: 0 blockers, 3 majors, 11 minors (10 on correctness, 1 on cost). No owner forks.

## Majors

### DR1. The XML tier passes any envelope for any endpoint (major)

**Location.** `:440-445` (Task 3, "Declared unvalidated" and the runner), `:488-494` (Task 4), `:204` (the XSD
quote).

**Defect.** The XSD's `Response` type is `<xs:choice minOccurs="0" maxOccurs="1">` (pinned XSD, line 12, checked
against the sha256 `9baee72f...` copy in slice 1's scratch). The quotes table cites this choice but leaves out
`minOccurs="0"`, which carries the weight here. xmllint therefore accepts three kinds of body for every endpoint:

- an `ok` envelope with no payload;
- an `ok` envelope carrying any payload in the choice, for example `<artists>` answering `getAlbum`;
- an `ok` envelope carrying `<error>`.

Task 3's "declared unvalidated" test reads the body's payload element, not the endpoint's. So a wrong payload name
outside the choice gives no finding at all (`:493`). Unlike the JSON checker, the tier never sees the HTTP status or
the content type. The JSON checker resolves the route, checks the status, checks the declared content type, and
validates against the path's schema (`internal/opensubsonic/conformance/conformance.go:92-117`). The oracle tier
catches a missing payload on an endpoint it compares. For an endpoint the assignment marks `contract-only`, though,
the XML tier is the only grade an XML answer gets. Those endpoints are `getIndexes`, `getMusicDirectory`,
`getScanStatus`, `startScan`, `download`, `getLyrics`, `tokenInfo`, the empty-success set, and the extensions list.
So slice 3 can serve `getIndexes?f=xml` with no `indexes` element, or with HTTP 500, and pass.

**Fold.** Bind each XML body to its endpoint from the pinned OpenAPI document:

- The path's success schema names its payload member under `subsonic-response`, and whether that member is required.
- An `ok` body must carry that element. It may omit it only when the schema has no payload (`ping` and the writes).
- A `failed` body must carry `error`.
- The HTTP status must be one the operation declares. The content type must be an XML type (`application/xml` or
  `text/xml`, with the L-06 caveat recorded).
- Each mismatch is a violation at ordinal 1, the envelope.
- "Declared unvalidated" keys on the endpoint's payload not being in the choice, never on what the body happens to
  carry.

Task 3 gains three proofs: a `getAlbum` body with no payload, a `getAlbum` body carrying `<artists>`, and an `ok` body
carrying `<error>`. Each must be a violation. Task 4 gains one mutation: drop the payload binding, and the proof
fails. The derivation reuses the kin-openapi document Task 3 already loads, so the cost is one function and four
proofs.

### DR2. The transcription tests freeze the ledger, and slice 3's first activation breaks them (major)

**Location.** `:564-566` (an active row "still names exactly one pair"), `:583-596` (accounting and fidelity).

**Defect.** Accounting says "nothing else is in the ledger", and fidelity says each committed row's columns "equal
the record's". 2a's `Row.applies` requires an active row to name exactly the graded pair and one session
(`grade/ledger.go`, `applies`). So a `Both` row such as L-01, L-05, or L-14 must become two active rows, one per
oracle pair, when slice 3 activates it. Activation therefore does one of three things:

- splits the row into new ids, which fails accounting;
- narrows `sources` from three names to two, which fails fidelity against "Both";
- adds a row for a newly found divergence, which fails accounting.

Slice 3 then edits a dated research record or weakens the test under deadline. Either outcome removes the guard 2b
built.

**Fold.** Scope the two checks now, with no new field and no export:

- **Accounting runs one way.** Every record id is a committed row (or a row whose id starts with that id, for
  example `L-01/navidrome`), an agreement, or a retired id. The ledger may carry other ids.
- **Fidelity runs on `owed` rows.** An `active` row derived from a record row may narrow `sources` to a pair drawn
  from the record's set. The fidelity test asserts that subset rule instead of equality.

Add one sentence telling slice 3 how a `Both` row activates: one row per pair, with the id suffixed by the oracle.
Add two mutations: a split active row passes, and an active row naming an oracle outside the record's set fails.

### DR3. The assignment has no check in the direction that drops oracle coverage (major)

**Location.** `:517-533`, `:537-553` (Task 5).

**Defect.** The assignment is now the only channel that decides whether dubplate meets an oracle. A request on an
endpoint marked `contract-only` skips the oracle tier for good. The report counts the skip but nothing fails (`:529-532`).
Task 5's tests check two things: that the endpoint set equals the document's paths, and that the album shapes match.
No test ties the 87 values to the study record's section 3 table, the way Task 6 ties the ledger to its record. A
single wrong `contract-only` on `getArtist`, `search3`, or `getPlaylist` would silently remove the oracle comparison
for that endpoint from every later slice. The four-value enum also drops section 3's qualifiers:

- `getMusicFolders` is "shape only";
- `getPlayQueue` and `savePlayQueue` go to Navidrome "for the empty answer";
- smart playlists go to Navidrome only;
- `stream` transcode goes to Navidrome and `stream` direct to both;
- `getOpenSubsonicExtensions` has a Navidrome carve-out and is contract only for the list.

Most of these losses err strict, which is loud. The `contract-only` direction is the silent one.

**Fold.** Add one gate test in the harmful direction only. Every endpoint section 3's table assigns an oracle must
not be `contract-only` in the assignment. Every `contract-only` entry must be one of the following, with the
exceptions listed in the test so that a new one needs a test edit:

- a section 3 contract-only row;
- a member of section 1's empty-success set;
- an endpoint not served at 0.8;
- an endpoint ruling C names.

The test can hold section 3's endpoint-to-oracle list as a literal transcribed by the implementer and read by the
diff-reviewer against the record. Each qualifier stays word for word in the entry's `reason`, so slice 3 reads it
where it grades.

## Minors

### DR4. The body-side credential guard sits in one verb, not in the session writer (minor)

**Location.** `:620-627`. **Defect.** 2a's request-parameter rule lives in the session writer and the loader
(`:194`). The new body scan runs only in `harness record-oracle`, and "No third channel writes a session" holds only
until slice 3 adds a dubplate recorder. A committed session is also never re-scanned on load. **Fold.** Put the scan
in `session` beside 2a's rule, in both the writer and the loader. `grade`'s L-07 finding calls the same function
(`grade` already imports `session`). `record-oracle` then inherits it. The recorder proof stays, and one loader proof
is added.

### DR5. Three proofs can pass on a body with nothing to find (minor)

**Location.** `:504-510` (Task 4's XML twin), `:547-551` (Task 5's grading proof). **Defect.** [contract] Q3 lists
Navidrome's post-projection residuals: `lastModified` on `artists`, `songCount` on `directory`, the `scanStatus`
fields, and `lyricsList`. None is on `getAlbum`. Navidrome's seed `getAlbum` XML is therefore likely valid, so the
twin's "exactly `xmlrule`'s violations" is an empty list. The "skip XML" mutation then passes the twin, and only the
seed golden catches it. Task 5's "still reports `getIndexes`'s contract findings" passes vacuously if the body
standing in for dubplate's has none. **Fold.** Name a fixture state for each proof:

- The twin uses gonic's `getAlbum` XML, which carries `title` and `album` on the ID3 `album` ([contract] Q3).
- The `getIndexes` stand-in is gonic's body, which lacks the required `lastModified`.

Each proof asserts a literal, non-empty expected list instead of one computed by calling `xmlrule`.

### DR6. L-51 keeps pre-ruling test text, and rows on contract-only endpoints can never activate (minor)

**Location.** `:569-572`, `:590-591`, `:250-254`. **Defect.** Fidelity exempts only the three ruled `dubplate`
texts. L-51's `test` therefore still says "Harness: folder rows graded by gonic, only if fork C is yes", and spec
`:342-345` still makes gonic authoritative for `getIndexes` and `getMusicDirectory`. Slice 3 reads both, and both
contradict the assignment. Rows whose endpoint is `contract-only` (L-17, L-20, L-21, L-51) can never become `active`,
because the assignment skips the oracle tier and an active row would go stale. They stay owed forever and count
against slice 3. **Fold.**

- L-51's `test` takes ruled text: "Slice 3: an empty `getIndexes` and error 70 from `getMusicDirectory`, asserted
  against the spec". Fidelity names it.
- The close's spec batch amends `:342-345`.
- Add one line routing to slice 3 the question of how a row on a contract-only endpoint closes.

### DR7. Text content with newlines shifts the line-to-ordinal table (minor)

**Location.** `:275-278` (R20), `:446-450`. **Defect.** "One start tag per line" holds only if no character data
holds a newline. Lyrics text and multi-line descriptions do. If the table assumes line N is ordinal N, an error is
silently placed on the wrong element, and a `contract` row may then suppress the wrong finding. **Fold.** Build the
table from the lines actually written: count the newlines emitted before each start tag. Add one proof: a body with a
multi-line text node before a violating element.

### DR8. The runner's argv should carry `--nonet` (minor)

**Location.** `:443-445`. **Defect.** A body with a DOCTYPE, or a schema-location hint, can make libxml2 try a fetch.
The gate promises no connection (`:41-42`). **Fold.** Pass `--nonet` to xmllint and assert the flag in the fake
xmllint's argv check.

### DR9. Any `brew upgrade` on the workstation reds every lane's gate (minor)

**Location.** `:268-270` (R18), `:159-162`. **Defect.** The version pin is exact (`21504`), and `go test ./...` in
every lane's gate now execs xmllint. Bluefin's `brew-upgrade.timer` is disabled here, so nothing upgrades on a timer.
A manual `brew upgrade` in any repo's sweep still bumps libxml2, since other formulae depend on it. That reds master
and every lane at once, and it mid-run trips this plan's own stop rule. **Fold.** Run `brew pin libxml2` and record
it in the Homebrew tier's manifest with the reason. The dependency sweep unpins, bumps, and re-pins together with the
goldens. Add one line to the close's CLAUDE.md batch.

### DR10. The known-deviation list is routed to slice 2 but is in neither half's scope (minor)

**Location.** `:82-93`. **Defect.** The study record's "Routed" list sends "the known-deviation list with `lastScan`
and the oracle nulls" to slice 2. Spec ladder entry 2 (`:885-886`) names it too. The scope table drops it without a
word. **Fold.** One row in the table naming where it lands. 2a's `contract` class with the stale-row rule is the
mechanism, Task 8's test ledger carries gonic's nulls as `contract` rows, and L-20's `lastScan` lands in slice 3.

### DR11. Task 3's test XSDs must not excerpt the Subsonic XSD (minor)

**Location.** `:460-461`, `:806`. **Defect.** "Small XSD documents written for the tests" invites pasting a real
`Child` type, which would commit unlicensed bytes (ruling B). The close's check, "no file carries the XSD's bytes",
reads as whole-file only. **Fold.** Test schemas use invented type names. The close's acceptance greps
`xmlrule/testdata` for the real XSD's type names (`AlbumWithSongsID3`, `NowPlayingEntry`) and finds none.

### DR12. The walk hashes only Rufin's real profile (minor)

**Location.** `:751-754`, `:770`. **Defect.** [base-url] hashed both apps' real trees and the GNOME keyring files.
Feishin's real `~/.var/app/org.jeffvli.feishin` now exists. Screenshots can also show the host's home path or window
chrome. **Fold.** Hash every tree [base-url] hashed. A changed hash stops the run, not only the task. Crop committed
screenshots to the client window and read them for host paths before staging.

### DR13. Three plan texts an implementer cannot act on as written (minor)

- `:613`: "`&` against `&` in JSON" compares a character with itself; the escape was lost. Restore it as `&`
  against `&`, as in [contract] E10.
- `:656-659`: "the test asserts each one present" contradicts "an absent prediction is recorded evidence, never a
  failure". Say that the test asserts the predictions the recording confirms, and that the report lists the rest.
- `:153-154` with `:726-732`: if Navidrome's second library is ruled unmeasured, Task 10's pairwise grade cannot run.
  Its digest check refuses one source, and its A-04 assertion needs both. Name the fallback: a gonic-only session
  with its own A-04 assertion and no pairwise golden, with Navidrome's L-19 evidence taken from a one-library run.

## Over-ceremony, by cost

### DR14. Task 11 walks every flow on both oracles (minor, about 0.3M to 0.4M)

**Location.** `:759-761`. The walk's job in 2b is to prove the method and the scripts that slices 3 to 8 run against
dubplate. Feishin and Rufin already walked Navidrome in slice 1, and a gonic walk records evidence about gonic, which
no roster check consumes. **Fold.** Walk each client once against Navidrome and run gonic only for the steps where
the oracles differ by record (the folder view, under ruling C). That cuts the GUI-driven half of the most expensive
task.

## Considered and not raised

- **Owed rows excusing divergences.** `Apply` counts owed rows and matches nothing with them, so R13 cannot suppress a
  finding.
- **Retiring the three folder-view rows (ruling C).** They were `owed` and suppressed nothing. The assignment carries
  their job, and the coverage slice 3 needs is a spec assertion, not an oracle comparison (see DR6 for the text).
- **Prerequisites missing at gate time.** Every test that needs them fails by name, with no skip (`:307-309`). A
  wrong xmllint on `PATH` is refused (R18), and the light lane's PATH is listed as unverified with a loud failure
  mode.
- **Gate network.** Uses `httptest`, a fake container runtime, and canned responses throughout (DR8 aside).
- **Determinism.** Goldens grade committed sessions, xmllint is pinned, and keys resolve by artist set, not by folder
  order.
- **Task 8 records lyrics and radio, which the assignment grades contract only.** This is slightly off R21, but it is
  one session and 2a's routed carry, so cutting it saves little.
- **The unattended-run rules.** Ceiling arithmetic, stop rules, and owner batching are proportionate. The close's
  1.3M is inside 9.0M.
