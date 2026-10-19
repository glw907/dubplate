# intake-1 plan fold verification

Reader: fresh context (Opus 5.5), no part in the fold. Target: `docs/plans/2026-10-12-intake-1.md` at
`d5e8f79`, read against the fold record `2026-10-12-intake-1-plan-fold.md`, the three reviews, the design, and the
spec (revision 2). Pass class `auth-data`. Every cited code pointer named below was re-read at `d5e8f79` or in the
module cache (tusd v2.10.1, mewkiz/flac v1.0.14, x/exp, x/sys v0.49.0, Go 1.27.2).

**Counts: 0 blockers, 1 major, 5 minor (correctness), 3 minor (optional).** Every blocker and major from the three
reviews closed at the location the fold record cites. One major is new to this read: the spec's whole-chain metadata
bound is narrowed to a picture total in T5. The rest are small file-list, wording, and ruling-evidence gaps.

## Probes and reads run for this verification

- `sqlite3` 3.53.4: a STRICT `submission_events` with `seq INTEGER NOT NULL` and `UNIQUE (submission_id, seq)` takes
  `INSERT ... SELECT ..., COALESCE(MAX(seq), 0) + 1 ... WHERE submission_id = ?` (1, 2 for one submission, 1 for
  another) and refuses a second row at one `seq` ("UNIQUE constraint failed: submission_events.submission_id,
  submission_events.seq").
- `sqlite3` 3.53.4, the olderStore undo: on a STRICT `collections` with the CHECKed `intake_route` column, its unique
  index, and a referencing `submissions` table, `DROP COLUMN` first fails ("error in index collections_intake_route
  after drop column"); `DROP INDEX`, then `DROP COLUMN`, then re-adding both succeeds.
- Go 1.27.2 scratch module: a `zip.Reader` over a `ReaderAt` that returns `syscall.EIO` mid-member surfaces the raw
  `EIO` (not `ErrFormat`, not `io.ErrUnexpectedEOF`) for both Store and Deflate members, and a recording wrapper
  captures it. P19's "classify by the gate's own I/O" is needed and works.
- `ffmpeg -f lavfi -i sine=frequency=440:sample_rate=44100:duration=5 -ac 2 -sample_fmt s16 t1.flac`: `metaflac`
  reads 44100, 16, 2, a nonzero MD5, and 220500 samples; `metaflac --set-tag=TRACKNUMBER=1` lands.
- Read: `internal/ops/errors.go:66-67`, `details.go` (ten constants), `errors_test.go:174-189`,
  `document_test.go:389-421`, `deck-coverage.ts` check 1b, `deck.ts:175-186`, `opshttp.go` `decode`;
  `tagio/container.go:133-257` and `Parse`'s callers; `memorylocker.go:59-80`; tusd `handler.go:30-70`,
  `unrouted_handler.go:44-50`, `:120-140`, `:1445-1458`, `:1280-1300`, `:1698-1708`, `filestore.go:70-100`,
  `:128-142`, `:300-312`, `datastore.go:58-64`, `config.go:192-199`, `error.go:22-36`; x/exp `slog.Logger` (a struct
  in both cached versions); `frame/frame.go:120-130`, `:176-182`, `:228-236`, `encode_frame.go:55-62`, `flac.go:65-76`,
  `encode.go:70-80`; `archive/zip/reader.go:290-330`, `compress/flate` `CorruptInputError`; `main.go:84-92`,
  `serve.go:210-240`, `:294-310`, `:401`; `olderstore_test.go:20-80`; `invariant_test.go` invariants 1 to 7;
  `Makefile:16-33`; `arch_test.go:210-251`; `check_test.go:122-171`, `serve_test.go:46-90`, `:682-690`, `:1913-1923`,
  `:2095-2125`, `doctor_test.go:978-984`, `doctorprobe_test.go:380-400`, `serve_store_test.go:341-372`,
  `auth_test.go:407-470`; `go doc ./internal/engine/store Handle.Update`; `.golangci.yml`; `check.sh:812-877`;
  `_salvage/internal/reject/reject.go:171-187`; PCF section 4 (`passb-copy-final.md:118-137`); the launch prompt.

## Question 1: did each blocker and major close where the fold record says?

Yes, every one, with one residual noted under finding 2.

- **Contract B1, mechanics XM3 (root A).** P4 "Three details" (plan:296-304), Conventions "Skills" (:144-147),
  Layering (:136-137), T6 (:772-774), T10 outcome, files, fixtures, and mutation (:966-993), T13 (:1135). Matches
  `errors.go:66-67`, the document test's both-ways enum, `TestDetailConstants`, and deck check 1b.
- **M1.** `submission_warnings` in T4 (:693-695), the verdict commit (:880-881), P4's warnings shape (:286-287), AC 22
  at T11 (:1054-1056). Checked against invariants 1 to 7 by reading: no `*_path` column, a single BLOB `id`, no `kind`
  column, no `_events` name, no name-and-value pair, no cascade onto `users` or `collections`.
- **M2.** P4 "A reason on the wire" (:278-282) matches the salvage `reasonWire` (`code`, `friendly`, `next_step`,
  `label`) and PCF section 4's "verbatim from the API"; T10 fixture and mutation (:987-992); the smoke (:1169-1171).
- **M3.** P2's `loadConfig` (:232-235) covers every production `config.Load` caller (`check.go:30`, `restore.go:152`,
  `member.go:58`, `serve.go:221`, `backup.go:204`; `doctor.go:48` apart). T3 fixture and mutation (:667, :673).
- **M4, XM4, XM5 (root B).** T5's split order and `ParseBounded` beside an unchanged `Parse` (:746-752) match
  `container.go:191` and `:200`; the two sentinels feed T6's classifier (:780-783). See finding 1 for the bound's
  scope, which this root left narrowed.
- **M5, XM7, R1 (root C).** The closed cleanup list (:1007-1014) matches spec:813-817's enumerated list; the
  `memorylocker` quote (:441-443) matches `memorylocker.go:59-72` (no try-lock, `requestRelease` called, then
  `ErrLockTimeout` on `ctx.Done()`); fixtures (:1067-1070).
- **M6.** P10 (:359-367), the uncached entry and the two `serverUnlinkable` prefixes. The prefixes work: the arch test
  greps `go list -deps ./cmd/dubplate-server` output (`arch_test.go:240-250`), so external module paths match. See
  finding 5 for one wording slip.
- **M7.** P9 (:355-357), T6 fixture (:807-808).
- **M8 (root I).** T9's refusal table (:946-950), T11's cancelled-mid-PATCH fixture (:1063-1064).
- **M9, XM6 (root D).** T9 mutation (:956-957). Confirmed: `handler.go` routes GET and DELETE on config flags alone,
  so the 405s survive `UseIn`; `NewUnroutedHandler` adds `creation-defer-length` only with a length deferrer, and a
  deferred length without one is `ErrNotImplemented`.
- **M10, XF7a (root G).** T11 AC 27 (:1071-1072); the test holds the lease, so no earlier serve's recovery consumed
  the rows.
- **M11.** T11 AC 28's real-roster closure and its mutation (:1078-1079, :1087).
- **M12, R3 (root N).** P19 (:404-414), T6 classifier and `EIO` fixture (:780-783, :808-810), T7 (:829-832, :846-847),
  T11 AC 24 (:1057-1059). Probed (zip passes `EIO` raw).
- **M13.** Each listed gap has its fixture line: T11 (:1051-1052, :1061-1064), T8 (:908-909, :898-900), T10 (:987),
  T6 (:805-807), T7 (:843-844).
- **M14.** Acceptance map's three departures (:41-44), T13's carried-clause item (:1136-1139).
- **M15 (root M).** Behaviors (:416-537) now quote `Statx`, `readDirectoryEnd`, `memorylocker`, `tagio`'s error
  contract, `ToValidUTF8`, the encoder's `Hash`, and `x/exp/slog` (P7, P14).
- **XB1.** Close step 3's generated CD-quality tracks (:1165-1168), probed above; E2E item 2 (:49-54).
- **XM1.** T4's undo list, doc comment, and mutation messages (:697-699, :731-733), probed above.
- **XM2.** T3's `baseEnviron` and `serveAuthEnviron` (:652-656), T11's teardown and two-shell rule (:1027-1031).
  Every `newServeShell` test I traced reaches `serveAuthEnviron` (directly, through `rosterAPIEnviron`, or through
  `renderedEnviron`); `serve_store_test.go:364`'s replaced library root is its own `t.TempDir()`, a sibling on one
  mount. One `serve` start does not: finding 2.
- **XM8 (root S).** P18 (:398-403); T6 and T7 split files (:817-818); `parseID` in T8 (:870), called from T9 (:930).
- **XM9 (root K).** Per-task review on T7 (:148-153, :840); T11's security read over the T6+T7 dispatch to T11
  (:1036-1045); Close step 2 scoped (:1158-1160); the path-map gap in Close step 6 (:1186-1187).
- **XM10 (root E).** T8's barrier (:900-904). The create barrier is deterministic for the named mutation. The
  receipt half is not: finding 3.
- **R2.** T11 (:1003-1006, :1027-1031), States (:565), fixture and mutation (:1065-1066, :1088), Close step 3 and E2E
  item 2. Matches `main.go:88` and `serve.go:229-234`; `sweepNonces` is started at `serve.go:401` and never joined.
- **R4.** T7 member and `x/` sync (:834-838), reconcile by size (:1015), fixture and mutation (:1074-1075, :1089).
- **R5 (root F).** `seq` in T4 (:688-692), Conventions (:126-130), P4 events in `seq` order (:284), T8 (:868-869).
  Probed above. `snapshot.go:100` runs `VACUUM INTO`, so the rowid alternative's refusal stands.

## Findings, ranked by consequence

### 1. Major (correctness): T5 narrows the spec's whole-chain bound to a picture total, leaving other blocks unbounded

- **Location.** Plan:746-752 (T5 outcome: "a Vorbis comment block, a picture block, and the chain's picture total";
  "the chain total is the picture one"), :757 (acceptance), :459-463 (Behaviors); P3's table (:245-246) names no
  chain figure. Spec:577-582: "no chain totalling more than four times the picture cap: `tags-too-large` and
  `picture-too-large`", checked "from each 4-byte header before the allocation".
- **Defect.** The spec bounds the whole chain at four times `MAX_PICTURE_BYTES`. The plan bounds only picture blocks'
  total and names no ratio. `parseBlocks` allocates and keeps every block's payload (`Block.Payload`,
  `container.go:52-64`, `:238`), whatever its type, so a crafted FLAC of many padding or application blocks, each
  under 16 MiB and the whole under the album cap, allocates and retains its full metadata size (up to gigabytes) in
  `ParseBounded`. `tagio.Read` then repeats the walk through the unbounded `Parse`. This is the "allocate without
  bound" case T11's security read asks about, and the plan's text steers the executor away from the spec's rule. The
  fold rewrote this bullet (root B) and kept the narrowed wording. The original draft carried it too.
- **Proposed fold.** T5: the third bound is the sum of every block's declared length, checked from the header before
  `make`, refused over four times the picture cap with `ErrPictureTooLarge` (spec step 2). Add a fixture, a chain of
  padding blocks each under both caps whose sum crosses four times the picture cap, refused before allocation. Add a
  mutation: count only picture blocks in the total (that fixture passes, so the test fails). Behaviors gains one
  sentence quoting spec step 2's ratio and noting that `Block.Payload` keeps every block. About 5 minutes on T5.

### 2. Minor (correctness): `TestDoctorBoundaryProbe_EndToEnd` runs the real `serve` over directories T11's device check refuses

- **Location.** Plan:1046-1049 (T11 files); `cmd/dubplate-server/doctorprobe_test.go:386-395` runs `newServeCmd`
  over `doctorEnviron` (`doctor_test.go:978-984`, built on `authEnviron`, then `baseEnviron`, `check_test.go:122-129`).
- **Defect.** After T3, `baseEnviron` carries `/srv/music/library` and `/srv/music/staging`, neither of which exists on
  the workstation. T11's startup device check refuses a start with staging absent (:1051-1052), so this test goes red
  in T11. That is the residual of XM2: the fold made `serveAuthEnviron` create sibling temporary directories, but this
  one test starts `serve` through `doctorEnviron`. `doctorprobe_test.go` is in no task's file list. The Opus executor
  can fix a red in its own package, but an unlisted file costs a fix cycle on an unattended run.
- **Proposed fold.** T11's files gain `doctorprobe_test.go`. The test points the library root and staging at sibling
  directories under one `t.TempDir()`, or `doctorEnviron` does, if T11's `doctor_test.go` pins tolerate it.

### 3. Minor (correctness): the "same barrier" for two receipts has no read to block on

- **Location.** Plan:903-904 ("the same barrier holds two receipts for the 'two files finish at once' clause").
- **Defect.** The create barrier is the injected free-space read, the create's last act before its `Update`. Receipt
  reads no free space: it syncs, then runs one `Update` (:875-879). So "the same barrier" names nothing on the receipt
  path, and the clause's proof is back to scheduler luck. That is XM10's defect for its second half.
- **Proposed fold.** Name the receipt's barrier: an injected sync (the receipt's last act before its `Update`) that
  blocks until both receipts reach it. The mutation is "read the received set outside the `Update`" (both receipts
  then see one file outstanding, and neither moves the submission, or both do).

### 4. Minor (correctness): ruling 1's "No" names no cut that reaches the line, and the design's token figure goes unaddressed

- **Location.** Fold record:156-166; plan:58-66; design:52-53 ("9.0M tokens, flag 7.2M, about 6.5 hours of
  execution; a plan over twice that comes back to you").
- **Defect.** The "No" branch offers AC 28's stall and mail, about 25 minutes, which leaves the plan at about 965,
  still 185 over the 780 line. A Geoff who answers No has no named option that meets the line, and the record does not
  say that none exists short of about 210 minutes of design scope. The design's line also names tokens, and neither
  the plan nor the record gives a token estimate, so Geoff cannot tell whether that half of the line is crossed.
- **Proposed fold.** Add two sentences to ruling 1. First, no design-preserving cut reaches 780, so "No" means naming
  about 210 minutes of scope to move, with the stall-and-mail cut as one piece of it. Second, give a token estimate
  with its basis (`admin-api-1`'s recorded tokens against its minutes, scaled), or state that none is recorded and the
  minutes figure stands in.

### 5. Minor (correctness): P10 puts the package on the wrong Makefile variable

- **Location.** Plan:363-364 ("T6 adds the test and its package to `UNCACHED_TESTS` and `UNCACHED_RE`");
  `Makefile:20-27`.
- **Defect.** `UNCACHED_TESTS` and `UNCACHED_RE` hold test names. The packages are the literal list in the recipe's
  `go test -count=1 -v -run ... ./internal/harness/conform ./internal/opensubsonic`. An executor who follows the
  sentence adds the name but not `./internal/intake/gate`, and the recipe fails closed ("did not run and pass"). That
  costs a red, not coverage.
- **Proposed fold.** "T6 adds the test's name to `UNCACHED_TESTS` and `UNCACHED_RE` and `./internal/intake/gate` to
  the recipe's package list."

### 6. Minor (correctness): shapes landed ahead of their consumers meet the field linter, and the state vocabulary has no stated home

- **Location.** Plan:700-701 (T4 lands P18's `policy.go` and the state-equals-event helper), :870 (T8: "`kind` is
  validated against the closed vocabulary"), :733 (T4's "count `stalled` as a state" mutation); `.golangci.yml:17-25`
  (`unused` with `exported-fields-are-used: false`); `check.sh:812-877` (the exported-no-consumer leg, warn by default).
- **Defect.** First, T4 lands `Policy`'s exported fields, the format table, and the per-file reason value with their
  first readers in T6 and T7. A field nothing reads fails `unused` at T4's gate unless T4's own tests read it. Second,
  T4's helper must tell state kinds from `stalled` before T8 defines the closed vocabulary. The plan does not say
  where the set lives, so T4 and T8 may each write a copy.
- **Proposed fold.** T4's tests read each `Policy` field (or T4 lands only what its own tests exercise, with T6 adding
  the rest; T6 is the only pair half that edits `policy.go`). The helper holds the state set as a literal in
  `stateevent_test.go`, which is the right independent expected value. T8's vocabulary lands in `submission.go` and is
  checked against that literal.

### 7. Minor (optional): the States nesting row lists two verbs where P2 lists five

- **Location.** Plan:572 ("`serve`, `check` | staging and the library root nest lexically | refused") against P2
  (:232-235: `serve`, `check`, `backup`, `restore`, `member`).
- **Defect.** The row predates M3's widening and now reads narrower than the decision. No test follows from it.
- **Proposed fold.** "`serve`, `check`, `backup`, `restore`, `member`."

### 8. Minor (optional): "no finding changes the approved design" is imprecise, and two schema additions are not listed with the spec readings

- **Location.** Fold record:5 and "Spec readings" (:135-152); plan P6 (:316-324); design "Files and interfaces"
  (:25-27); T4 (:688-695).
- **Defect.** No fold changes the design (question 4). Before the fold, though, the plan already built less than the
  design's file list in two places: no `internal/auth` `RequireMember`, and the tus mount in `internal/intake`, not
  `internal/serve`. Contract m1 named both, and the fold folded them as PR-body sentences. That is defensible, since
  CLAUDE.md's "nothing is exported ahead of a caller" outranks the list. But Geoff reads the plan before launch for
  ruling 1, and the record tells him nothing touches the design. Also, `seq` and `submission_warnings` add to the
  spec's three-table `0006` without changing its meaning, and the "Spec readings" list omits them.
- **Proposed fold.** One informational line under "Rulings for Geoff", not a fork: the two file-list departures and
  why. Two lines in "Spec readings": `seq` and `submission_warnings`, additive, with their defects.

### 9. Minor (optional): the detail-count comments

- **Location.** `internal/ops/details.go:3` ("The ten values"), `errors_test.go:174` ("pins the ten wire details").
- **Defect.** T10 adds three constants, so both comments go stale.
- **Proposed fold.** T10's outcome names both comments (thirteen).

## Question 2: contradictions, unbuildable orders, estimates

- **Contradiction with the spec.** Finding 1 (the chain bound). No other plan-to-spec contradiction was introduced.
  The four "Spec readings" read correctly against spec:549, :631-634, :813-817, and :1055, and P5's
  `checking`-only stall is named in the acceptance map and the PR body.
- **Between tasks and decisions.** Finding 7 (the States row) only. P4's details and Layering agree, since the gate
  returns manifest errors and the T10 seam maps them. T6's classifier agrees with P19 and T11 on cancellation. The
  cleanup list, the reconcile rule, and the States rows agree. The launch prompt defers to the plan for every rule the
  fold touched.
- **Order and pairs.** Every pair builds on what earlier tasks land. T4's `policy.go` needs `internal/reject` (T2,
  earlier) and `tagio.Format` (shipped). T6 and T7 touch disjoint files and share T4's shapes. T8 precedes T9 and T10
  (`parseID`, receipt, the wake channel). T9 and T10 write disjoint files (T10 alone touches `cmd/`). `internal/ops`,
  `settings`, and `store` do not reach `internal/serve` (`go list -deps`), so P8's in-package test cannot cycle.
  Residuals: finding 2 (an unlisted test file for T11) and finding 6 (lint at T4).
- **Estimates.** The arithmetic holds: the per-task sum is 990, the path is 725, the pair maxima are right, and the
  per-task deltas sum to the fold's 230. Each delta names work of about its size. T11 at 145 is the one estimate at
  risk. It carries the largest added surface (the cleanup list with locking, cancellation and teardown, reconcile by
  size, torn `.info`, the address closure, about a dozen fixture groups, eleven mutations) plus a security read that
  now spans T6 to T11, including both untrusted-input parsers. Its two-times stop is 290, and `admin-api-1` ran at
  about 68 percent, so I do not raise it as a finding.

## Question 3: is every new mechanism quoted or probed?

Yes, with the probes above added to the fold's.

- **`seq`, its ordering, and the unique refusal.** Probed again here. The `VACUUM INTO` rowid reason is
  `snapshot.go:100` plus the fetched SQLite page.
- **`submission_warnings`.** Checked here against invariants 1 to 7 by reading. The fold record states it under "the
  store's own invariants" without a probe; T4's invariant test will prove it.
- **The closed cleanup list.** Quoted from spec:813-817 and `memorylocker.go`. The lock timeout's value is left as
  "short", which is an executor's call.
- **The SIGTERM teardown order.** Quoted (`main.go:88`, `serve.go:229-234`, `:401`).
- **The gate's I/O-fault classification.** Quoted (`frame/frame.go:128`, `:234`; `archive/zip` `checksumReader`;
  `flate.CorruptInputError`) and probed here for `zip`. The record calls `checksumReader` "the pattern" for the
  recording reader, which is loose: it wraps and hashes, it does not record errors. The mechanism is a plain wrapper,
  so nothing rests on the citation.
- **The reject wording on the wire.** Quoted (salvage `reasonWire`, PCF section 4).
- **The three `Detail` constants and their deck lines.** Quoted (`errors.go:66-67`, the document test, `TestDetailConstants`,
  deck check 1b). The deck gate's check 2 (no straight apostrophe or quote) also binds the three new lines; T10's
  executor meets it at the `web/` leg.
- **The capacity-1 wake channel.** Stated as the standard non-blocking `select` send. It is a language idiom, not a
  library behavior, so no quote is owed.
- **`MaxSize` at the ceiling.** Quoted (`unrouted_handler.go:45`, `validateMaxSize`). `FileInfoChanges.ID` is honored
  by `filestore.NewUpload` (`filestore.go:82-84`) and validated by `validateUploadId` (`unrouted_handler.go:1738-1755`),
  which supports P7's "the id is tusd's id".

## Question 4: can Geoff rule the fork, and does any fold change the design?

**The fork.** Yes, on the evidence given. The record gives the number (990 against 780), where each of the 230 added
minutes went, the trims taken, the path (725), the per-task stop, and the `admin-api-1` actual-to-plan ratio. That is
enough for a Yes. A No is under-specified: the one cut it names (about 25 minutes) leaves the plan about 185 over the
line, and the design's token half of the line is not addressed (finding 4). Two sentences close both gaps.

**The design.** No fold changes it:

- **Outcome.** Unchanged.
- **Files and interfaces.** Fold additions stay inside lines the design already names: `0006`'s `seq` and
  `submission_warnings` sit within "migration `0006`"; the three `Detail` constants and their deck lines sit within
  "`internal/ops` and `opshttp` (the submission operations and their contract rows)", since the deck gate makes copy
  lines part of a detail; `loadConfig` sits within the config line.
- **Out of scope.** Unchanged. The deck lines are drafted copy filed for `intake-2`'s copy read, as T2's new reasons
  already were. They are not a PCF redraft and not a page.
- **End-to-end check.** Kept and strengthened: CD-quality tracks so the pass verdict can occur, wording compared with
  the literals, and a SIGTERM restart beside `kill -9`.
- **Risk class.** `auth-data`, unchanged.

The two pre-fold departures from the design's file list (`RequireMember`, the tus mount's package) are documented and
justified by a standing rule that outranks the list. They are not fold changes. Finding 8 asks that Geoff see them in
the same pre-launch read.
