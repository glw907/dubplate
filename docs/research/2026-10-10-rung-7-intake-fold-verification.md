# Rung 7 intake spec: fold verification, revision 2

Target: `docs/superpowers/specs/2026-10-10-rung-7-intake-design.md` at `5cf5241` (cited `spec:N`). Fold record:
`research/2026-10-10-rung-7-intake-fold.md`. Reviews: `research/2026-10-10-rung-7-intake-review-*.md`. The verifier
had no part in the fold. Code was read on the tree at `5cf5241`, tusd v2.10.1 and mewkiz/flac v1.0.14 from the module
cache, archive/zip from Go 1.27.2's `GOROOT`, and `tus-js-client` 4.3.1 from its npm tarball. Three probes ran under
the session scratchpad and were removed afterward.

Counts: 1 blocker, 1 major, 5 minors. Findings are ranked by consequence; polish is left out.

## Findings

### B1 (blocker). A ZIP's staged bytes have no defined home after extraction, so restarts expire good ZIPs and a ZIP in `check-unfinished` loses its archive

Location: `spec:536-541`, `spec:773-778`, `spec:271`, `spec:62-64`; AC 24 (`spec:1142-1144`) and AC 27
(`spec:1153-1157`).

The fold added three rules that do not compose for a ZIP:

- "The archive is removed after the verdict commits, and the startup pass and the sweep remove any archive left
  behind past `checking`" (`spec:540-541`, repeated at `:778`).
- `check-unfinished` is a verdict out of `checking` (`spec:290`), so it is "past `checking`". "Any failure removes the
  directory" (`spec:537`) also removes `x/`. A ZIP in `check-unfinished` then has neither its archive nor its
  members, so `retry_submission` has nothing to re-run. That breaks conductor ruling 5 ("keeps its bytes") and the
  PCF copy it cites ("already uploaded and waiting").
- Reconcile expires "a non-terminal submission whose received files are missing" (`spec:774-775`). The archive's
  `submission_files` row has `role` `archive` and `received_at` set (`spec:244`), and its bytes are removed by
  design. Read literally, every ZIP in `matching`, `pushed-back`, `match-unavailable`, or `accepted` expires at the
  next `serve` start, which includes every deploy. An `accepted` ZIP waiting for rung 9 is lost each time.

No criterion catches either case. AC 24's fixture is "a passing two-track album" with no ZIP. AC 27's fixtures are
empty staging and orphans, never a healthy ZIP past its verdict.

Proposed fold: name a ZIP's staged bytes per state in one sentence. The archive is the staged bytes through
`checking` and `check-unfinished`. From the first verdict that consumed the extraction (`matching`, `accepted`,
`rejected`) on, the staged bytes are the `x/` members. Only then is the archive removed. Reconcile tests a ZIP
against that set, not against the archive row. Cleanup's "past `checking`" becomes "past `checking`, other than
`check-unfinished`". AC 24 gains a ZIP fixture: the fault, then retry, reaches the verdict. AC 27 gains an `accepted`
ZIP whose archive is gone and whose members are present, and it survives the restart.

### M1 (major). The end-of-central-directory entry count does not bound `zip.NewReader`

Location: `spec:524-526`; AC 18 (`spec:1117`); states row `spec:1023`; fold record `:206`.

The fold answers mec-M5's ZIP half by reading the entry count from the end-of-central-directory record and refusing a
count over `MAX_FILES` before `zip.NewReader` runs. `archive/zip` does not stop at that count. `Reader.init` reads
headers "until we encounter a bad one", and it checks the count only modulo 65536:
`if uint16(len(r.File)) != uint16(end.directoryRecords) { // only compare 16 bits here`
(`$GOROOT/src/archive/zip/reader.go:139-161`). The claim in the fold record, that `readDirectoryEnd` "reads the same
record", is true. It does not make that record authoritative.

Probe: a stored ZIP of 196,609 entries (65,536 × 3 + 1), 21 MB, with its zip64 end record patched to claim 1 entry.
The pre-check reads 1 and passes. `zip.NewReader` returns nil error, `len(r.File)` is 196,609, and it allocated 53 MiB.
At the 4 GiB album cap, the same construction holds tens of millions of entries, so the bomb mec-M5 measured stays
open.

Proposed fold: before `zip.NewReader`, count central-directory headers by streaming from the directory offset with
the same stop rule `Reader.init` uses: read each 46-byte header, skip its name, extra, and comment lengths, and stop
at the first non-header signature. Refuse the `MAX_FILES + 1`th header with `too-many-files`. Memory stays constant.
AC 18 gains a fixture whose end record understates its entry count. The planning-miss list quotes `Reader.init`'s
loop, not only `readDirectoryEnd`.

### m1 (minor). Receipt's store-error status is unnamed, and AC 10 can be read to forbid the right one

Location: `spec:383-389`, the refusals table (`spec:402-411`), AC 10 (`spec:1088`).

When receipt's `Update` fails on a store error, the callback's status decides what the member sees. A 5xx (a plain
error gives 500 through `sendError`, `unrouted_handler.go:1282-1289`) is retried, the retry's HEAD answers the full
offset, and the sweep finishes receipt. That is the B1 repair working. A 400 or 403 ends the client with a failure
for bytes that are safely on disk. The table has no row for this state, and AC 10's "No callback refusal answers
500" pushes an implementer toward a 4xx. Fold: add the row "receipt: store error | 503 | `could-not-check`
(retried)". Scope AC 10's sentence to definitive refusals.

### m2 (minor). The client needs `endpoint` beside `uploadUrl`

Location: `spec:418-420`.

On a HEAD 404, tus-js-client creates only when `options.endpoint` is set. Otherwise it fails with "tus: unable to
resume upload (new upload cannot be created without an endpoint)" (`lib/upload.js:685-693` in 4.3.1). With
`uploadUrl` alone, every file's first upload fails. The spec says the client "creates only after a 404" without naming
the option. Fold: "`endpoint` is the tus base path, and `uploadUrl` is the base path plus the file id".

### m3 (minor). AC 26's "worker verdict after a cancel" has no reachable fixture

Location: AC 26 (`spec:1151`), states row `spec:1042`, transition table `spec:295`.

`cancel_submission` is valid from `uploading`, `check-unfinished`, `match-unavailable`, and `pushed-back`. The worker
and the probe run only in `checking` and `matching`, so no member action can change state under them. The criterion
either passes vacuously or needs a direct store write. Fold: drop the clause, or name the race that can occur, which
is a receipt finishing after a cancel. AC 11 already covers that race.

### m4 (minor). Two separate-tracks states have no named report

Location: `spec:517`, `spec:596-598`, `spec:615-616`.

The spec says "A separate-tracks submission carries no extras" but names no reject for a manifest or ZIP member
that is an extra. It also does not say that a submission whose passing tracks are all refused `already-in-library`
ends `rejected`. The transition `checking` to `rejected` exists, but the accept `Update` must choose it. Fold: name
both. Either an extra is refused with a stated code (`invalid_request` at create, `unsafe-path` or a new code for a
ZIP member), or it is dropped with a warning. All duplicates means `rejected`. Add one fixture for each case to AC 22.

### m5 (minor). The ID3 fork omits that the reference decoder accepts an ID3v2 prefix

Location: `spec:1219-1227`.

Probe on a generated 16-bit FLAC. With an ID3v2 block prepended, `flac -t` exits 0 with "WARNING, ID3v2 tag found.
This is non-standard and strongly discouraged". mewkiz/flac also decodes all 88,200 samples, so only `tagio.Sniff`
refuses the file. With an ID3v1 record appended, `flac -t` exits 1 (`LOST_SYNC`) and mewkiz fails with
"frame.Frame.parseHeader: invalid sync-code", as the spec says. The fork paragraph is accurate but leaves out the
first fact, which bears on option (c) for the prefix case. Fold: add one sentence saying the reference decoder
accepts a prefixed file with a warning and refuses a trailing ID3v1 record. Option (c) would also need `tagio`'s
writer to handle the prefix before rung 9 could write tags.

## Question 1: did each blocker and major close?

| Finding | Closed at | Verdict |
|---|---|---|
| int-B1 stranded upload (with con-C1, con-C15) | `spec:369-389`, `:418-420`, AC 11-12 | closed. tusd derives the offset from the data file (`filestore.go:166`) and calls the finish callback only from the completing PATCH (`unrouted_handler.go:993-1019`). tus-js-client takes `uploadUrl` first (`:432-437`) and reports success on a full offset (`:724-727`). Replay by the startup pass and the sweep closes the stuck case. See m1 and m2. |
| int-B2 duplicate race (con-C2, mec-M1) | `spec:607-619`, transition rows `:291`, `:294`, AC 23 | closed. The check runs in the `accepted`-entering `Update` on all four paths, against the target collection. |
| con-C3, con-C4, int-M2 dead ends | `spec:280-298`, `:426-429`, AC 12, 26, 28 | closed. Items 2 and 4 of int-M2 were refused with reasons that hold. See m3. |
| con-C5 track numbers | `spec:578-589`, AC 21 | closed |
| con-C6, int-M5 ZIP quota | `spec:454-462`, AC 16 | closed |
| con-C7 uncovered codes | AC 10, AC 17 | closed |
| con-C8 race by luck | `spec:444-446`, AC 15 | closed |
| con-C9 AC 9 chain | AC 9 | closed |
| con-C10 wire shapes | `spec:395-411`, `:466-468`, `:750-751` | closed, apart from m1 |
| con-C11 PCF and W8.2 | conductor ruling 6, `spec:720-727`, AC 17 | closed |
| con-C12 `EvalSymlinks` | `spec:119-124`, AC 1 | closed |
| mec-M2, cst-M1 staging key | `spec:111-117`, AC 1 | closed |
| mec-M3, int-M4 single process | `spec:338-344`, AC 27 | closed. `TakeKindLease` is an exclusive non-blocking flock on `leases/kind-<kind>`. It accepts `intake` (`validKind`, `lease.go:152-163`), refuses only a handle with no store lock (`OpenServe` holds one, `open.go:201-203`), and wraps `ErrLeaseBusy` with the pids `/proc/locks` names (`lease.go:174-200`). |
| mec-M4 cross-compile | `spec:188-205`, AC 34 | closed |
| mec-M5 memory bombs | `spec:551-556`; `spec:524-526` | FLAC half closed: `parseBlocks` decodes the 24-bit length from the 4-byte header before `make([]byte, length)` (`container.go:178-199`). Frame bytes after the chain decode as type 127 and break first, so the cap must sit after the invalid-type test. ZIP half **not closed** (M1). |
| mec-M6, int-M1 re-extraction | `spec:523`, `:536-541`, AC 19 | closed for a crash in `checking`. The archive's lifecycle in later states is B1 above. |
| int-M3 restore | `spec:773-776`, AC 27 | closed for empty staging. It overreaches for ZIPs (B1 above). |
| int-M6 sync | `spec:383-385` | closed for tus files. Extracted members are not synced before the verdict (`spec:536-538`). Rung 9's `file_hash` re-verify (`spec:879`) detects a torn member, so the silent case the lens named is closed. |
| cst-M2 unit grant | `spec:144-161`, AC 3 | closed |
| cst-M3 held state | conductor ruling 5, `spec:271`, AC 24 | closed for loose files. **Not closed for ZIPs** (B1 above). |
| cst-M4 | owner ruling 7 | ruled |
| cst-M5 picker | `spec:665-669` | closed |

## Question 2: contradictions and segment order

- The transition table (`spec:287-298`) agrees with the state table, the operations (`spec:755-756`), the probe
  table (`spec:693-698`), and the sweep (`spec:787-791`). The one conflict is AC 26's cancel race (m3).
- Separate tracks against "the album is the verdict unit". Inside the spec, the rule is scoped to albums
  (`spec:221`, `:584-591`). Steps 5 to 7 are skipped for tracks (`spec:596`). The CLAUDE.md line is listed as owed
  (`spec:1199`). No text conflicts.
- `check-unfinished`, `could-not-check`, and `source-damaged` agree with the reject list (`spec:729-733`), AC 17, AC
  20, and AC 24. The exception is the ZIP case in B1.
- Segment order builds. A's migration, config, and reject set come before B's gate and limit functions, which come
  before C's tus, receipt, worker, and sweep. A passing album that reaches `matching` in 7a has no probe until 7b. The
  plan assigns AC 25 and AC 27's `matching` half to 7b.

## Question 3: mechanisms stated from memory

| Mechanism | Evidence | Verdict |
|---|---|---|
| Bytes kept in `tus/`, HEAD answers the full offset | `filestore.go:130-166`; `unrouted_handler.go:993-1019` | verified |
| `uploadUrl` resumption, HEAD first, create after a 404 | `lib/upload.js:432-437`, `:668-698`, `:724-727`; retry rule `:1076-1081` | verified, with the `endpoint` requirement unstated (m2) |
| `FileInfoChanges.ID` and the empty `MetaData` | `datastore.go:55-75` | verified |
| End-of-central-directory count bounds `NewReader` | `reader.go:139-161`; probe | **fails** (M1) |
| `tagio`'s block bound before allocation | `container.go:178-199` | feasible as stated |
| The `intake` kind lease | `lease.go:148-200`; `open.go:201-203`; `local.go:227` | verified |
| `O_TRUNC` on create; no `Sync` | `filestore.go:311-313`, `:324` | verified |

## Question 4: can the owner rule the ID3 fork?

Yes. The three options, their costs, and the recommendation are clear, and the decoder behavior they rest on is
accurate. Adding the one fact in m5 makes the prefix case's option (c) fairly stated. It does not change the
recommendation.

## Question 5: Geoff's four rulings of 2026-10-10

1. "Unlisted music": correct. Owner ruling 1 (`spec:30-33`), the collection (`spec:650-661`), AC 5 and AC 22, the
   PCF erratum (`spec:1196-1197`), and no leftover "unmatched" or "Private Recordings" outside the errata.
2. "Library": correct. Owner ruling 5 (`spec:41`), `spec:643`, AC 5.
3. Separate tracks: correct. The kind is chosen at upload with a per-submission default of album (`spec:42-45`,
   `:215-225`, `:829`). Verdicts are per track (`spec:596-598`), tracks land in "Unlisted music" with no probe, and
   AC 22 asserts that no MusicBrainz request is made. Loose-track matching is deferred until after 0.8
   (`spec:883`), and the non-album path goes to rung 9 (`spec:878`).
4. A rejected submission is final and re-sent whole: correct. Owner ruling 7 (`spec:46-47`) and the PCF erratum
   (`spec:1195`). `rejected` is terminal in the transition table. Fix-names cancels and returns to selection
   (`spec:703-704`). `check-unfinished` is not a rejection, so its retry is consistent with the ruling.
