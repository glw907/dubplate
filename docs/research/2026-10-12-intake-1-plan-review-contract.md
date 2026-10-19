# intake-1 plan review: contract and criteria

Reviewer: fresh context, contract-and-criteria lens. Read: the plan at `3783060`, the design, the spec (revision 2), both
fold records, the `dubplate-pass` adapter's planning-miss rule, and the code each finding cites. Plan line numbers are
`docs/superpowers/plans/2026-10-12-intake-1.md` at `3783060`.

Counts: 1 blocker, 14 major, 13 minor, and 2 over-ceremony items (both minor). One finding is an OWNER FORK (m4).

## Blocker

### B1. P4 adds no `Detail*` constant, but the spec's details `store_url` and `extra` need one (correctness)

- **Location.** Plan `:266` ("No `Detail*` constant is added"), against T6 `:605` (detail `extra`), T10 `:766-767`
  (`invalid_request` detail `store_url`), and the spec's states rows `spec:1055`, `:1058`, and `:1078`.
- **Defect.** In 14a's contract, `Error.Detail` "is one of the Detail constants" (`internal/ops/errors.go:66-67`).
  `TestDocument_EnumerationsMatchTheCode` holds the OpenAPI `detail` enum equal to `details.go`'s constants both ways
  (`internal/ops/opshttp/document_test.go:392-409`). `deck-coverage.ts` fails on any constant that has no
  `API_DETAILS` line (`web/scripts/gates/deck-coverage.ts:9-11`, `web/src/lib/copy/deck.ts:170-186`). T10's
  `store_url` fixture and T6's `extra` fixture can pass only by breaking that contract (a detail outside the enum,
  which the document test refuses) or by dropping the spec's detail. The spec's malformed row ("`invalid_request`
  with a detail", `spec:1055`) has the same problem. A dispatched T10 hits this contradiction and stops, and on an
  unattended night that stop costs the run.
- **Fold.** Replace P4's sentence. T10 adds `DetailStoreURL = "store_url"`, `DetailExtra = "extra"`, and one detail
  for the malformed manifest (a duplicate path, a negative size), extends the `openapi.yaml` detail enum, and adds
  one `API_DETAILS` line for each in `web/src/lib/copy/deck.ts`, drafted to PCF's rules and added to T2's
  `prose-voice-reviewer` read. Amend the convention "`web/` is touched only by T0's one pin" (`:132`) to cover this
  edit. T10's file list gains `internal/ops/details.go` and `deck.ts`.

## Major

### M1. Warnings have no storage (correctness)

- **Location.** P4 `:258` (`warnings` (codes)); T4 `:531-541`; T8's verdict commit `:683-686`; AC 22 at T11 `:808`.
- **Defect.** `get_submission` answers `warnings`, and AC 22 needs "the extras warning" listing the dropped members.
  Extras get no row (`spec:631`). `0006` has no warnings column, and `submission_events.kind` is a closed vocabulary
  of states plus `stalled`. No task writes a warning anywhere, so the completeness warnings, and the extras
  warning's member names, cannot be read back. The spec has the same gap; the plan inherits it.
- **Fold.** In T4, add a `submission_warnings` table (`id`, `submission_id`, `code`, `subject` BLOB for a member
  name, `at`). Do not name a column `kind`, which keeps invariant 3 out of it. T8's verdict commit writes the
  warnings, P4's `warnings` becomes `[{code, subject}]`, and the T11 AC 22 fixture reads the extras warning through
  `get_submission`.

### M2. Reject wording has no channel on the wire, so the end-to-end check's "contract wording" cannot be tested (correctness)

- **Location.** Plan `:43-46`, Close step 3 `:891-892`, P4 `:256-263`.
- **Defect.** Every verdict on the wire is a bare code: `files[].verdict`, `events[].reason`, and `rejects`. The
  check says to read "`id3-tags` in its contract wording" through `/api/v1`, and no operation returns wording. PCF
  section 4 assumes the wire verdict carries a `label` ("Label fallback when a file's own wire verdict carries no
  `label`", `passb-copy-final.md:130-131`). T2's `label`, friendly sentence, and next step also have no production
  reader in 7a apart from the tus refusal bodies. That conflicts with "No export ahead of a caller" (`:126`).
- **Fold (recommended).** Each failing file and rejecting event in P4 carries `{code, label, reason, next_step}` from
  `internal/reject`. The change is additive under 14A's version policy, and it gives T2's fields their caller. The
  smoke then compares the wording with the set's literal. **Alternative.** Restate the check to read the code alone,
  and rely on T2's literal test for the wording. This is a wire-shape call, so the PR body should name it.

### M3. P2 narrows AC 1's "config load refuses" to three commands (correctness)

- **Location.** P2 `:213-216`; T3 `:505-506`, `:513-514`.
- **Defect.** AC 1 says "Config load refuses" a nested pair. P2 calls `resolveStaging` only from `serve`, `doctor`,
  and `check`. `backup` (`cmd/dubplate-server/backup.go:204`) and `restore` (`restore.go:152`) also call
  `config.Load` and act on the library root. The spec's own reason for the rule is that `internal/backup` syncs the
  whole library tree (`spec:123-125`). The narrowing is not stated as a decision.
- **Fold.** `backup` and `restore` also call `resolveStaging` after `Load`, or every command loads through one shared
  helper. T3 gains a fixture: `backup` refuses a nesting environ.

### M4. T5's truncated-block fixture cannot tell where the bound sits (correctness)

- **Location.** T5 `:584-596`.
- **Defect.** `parseBlocks` breaks on `blockType == blockInvalidType || bodyOffset+length > size`, one condition at
  `internal/engine/tagio/container.go:191`. In a file truncated just past the header, the past-the-end clause
  breaks the walk before a bound placed "after the type-127 test" is reached. The chain then ends without the
  last-block flag, so the result is `ErrMalformedContainer`. A correct bound and the "move the bound after the
  allocation" mutation therefore give the same error, and a test that asserts the bound's error fails on correct
  code.
- **Fold.** State the placement: split the condition so the order is type 127 (break), then the bound (error),
  then past the end (break), then `make`. The truncated fixture then reaches the bound, and the mutation answers
  malformed. The alternative is a full-length oversized block whose allocation `testing.AllocsPerRun` or a
  `runtime.MemStats` delta measures. The split is cheaper.

### M5. `memorylocker`'s lock behavior is recalled, not quoted, and the plan describes it wrongly (correctness)

- **Location.** T11 `:780-781`; spec `:815-817`.
- **Defect.** `memorylocker` has no try-lock. `Lock(ctx, requestRelease)` on a held lock calls the holder's
  `requestRelease`, which tells an in-flight PATCH to stop. It then blocks until the lock is released or `ctx` ends,
  and returns `handler.ErrLockTimeout` (`pkg/memorylocker/memorylocker.go:59-72`, v2.10.1). So "taking each
  upload's lock and leaving a held one for the next sweep" needs a bounded `ctx`, and it always interrupts the
  holder. No fixture covers a held lock.
- **Fold.** Quote `Lock` in "Behaviors" and name the cleanup's lock timeout. State that cleanup touches only bytes
  outside the staged set (terminal or orphaned), so the interruption is intended. Add a T11 fixture: a lock held by
  the test whose `requestRelease` records the call and does not release. The first cleanup leaves the file, and the
  next sweep, after release, removes it.

### M6. P10's cross-compile test can replay a stale pass from the test cache (correctness)

- **Location.** P10 `:304-308`.
- **Defect.** A `go build` subprocess is invisible to `go test`'s cache. The Makefile handles this case for the
  precedent P10 cites: `UNCACHED_TESTS` runs `TestServerNeverLinksTheHarnessOrChecker` with `-count=1` because "a
  cached pass can replay after the source it checks has changed" (`Makefile:16-33`). If the gate package does not
  import `staging`, a change that breaks the windows build of `staging` leaves the gate test cached and green, so
  AC 34 passes without being checked. "It needs no gate-machinery edit" is wrong.
- **Fold.** Add the test to `UNCACHED_TESTS` and `UNCACHED_RE`. T6 already runs the full gate. Also add `filelocker`
  and `lockfile` to the existing `go list -deps` arch test, so AC 33's server clause gates every run instead of
  being one command at close (`:48`, `:740-742`).

### M7. AC 20's audio-key check uses the helper under test (correctness)

- **Location.** T6 `:625-631`; P9 `:298-303`.
- **Defect.** mewkiz's encoder fills STREAMINFO's MD5 by calling `f.Hash(enc.md5sum)`
  (`mewkiz/flac@v1.0.14/encode_frame.go:59`). That is the same `Frame.Hash` the gate uses for the audio key. So
  "`audio_key` equals its header MD5" over a generated fixture cannot catch a `Hash`-level or depth-handling defect,
  which breaks CLAUDE.md's rule that an expected value is never built with the helper the code under test calls.
  The `file_hash` check against `sha256sum` is independent; the key check is not.
- **Fold.** The test computes the expected MD5 itself from the generator's sample slice: interleaved, signed,
  little-endian, at the byte-rounded width that the FLAC format specifies for STREAMINFO's MD5. It asserts the
  audio key against that value, and keeps the header equality as a second assertion.

### M8. The tus-path refusal rows have no fixture, and one mutation names no failing test (correctness)

- **Location.** T9 `:734-745`; the states row `:430`; spec `:399-416`.
- **Defect.** No task's fixture list carries these rows:
  - the floor on create and PATCH (403 `no-space`);
  - a failed `statfs` (503 `could-not-check`);
  - receipt into a closed submission (403 `conflict`);
  - receipt on a store error (503).

  T9 lists fixtures only for AC 7 to AC 10. "Skip the floor on PATCH" therefore fails no named test. AC 11's last
  clause ("its bytes are removed" after a cancel) belongs to no task. T8 proves "changes nothing" at the function
  level, and nothing proves the 403 or the cleanup.
- **Fold.** T9 adds a table test over every row of "Refusals on the tus path", with an injected free-space read and
  an injected store fault. Each row asserts the status, `text/plain`, and the body prefix `<code>: `. T11 adds the
  cancelled-mid-PATCH fixture and asserts that the bytes are gone after one sweep.

### M9. T9's `FileStore.UseIn` mutation does not fail the rows it names (correctness)

- **Location.** T9 `:743-744`.
- **Defect.** tusd routes GET only when `!config.DisableDownload`, and DELETE only when `!config.DisableTermination`
  (`pkg/handler/handler.go:59-64`). The 405 rows hold whatever the composer contains. With `UseIn`, the length
  deferrer is composed, so AC 7's `Upload-Defer-Length` row stops answering 501. Under the plan's own rule, a
  mutation that does not fail sends the executor after a test that cannot be written.
- **Fold.** Retarget the mutation to "(the `Upload-Defer-Length` 501 row fails)".

### M10. T11's lease-before-recovery proof can pass on an empty fixture (correctness)

- **Location.** T11 `:814`, mutation `:823`.
- **Defect.** "Refused at the lease before any recovery runs" names no staging state that recovery would change. On
  empty staging, recovery changes nothing, so the "run recovery before taking the lease" mutation passes.
- **Fold.** The test holds `TakeKindLease("intake")` itself. Staging holds an orphaned `submissions/<hex>/`
  directory and a `checking` row whose bytes are missing. After the refused start, both are unchanged.

### M11. In AC 28, the test's fake decides "none to a removed admin" (correctness)

- **Location.** Layering `:122-123` (cmd injects "the live-admin addresses"); T11 `:819-821`.
- **Defect.** The sweep in `internal/intake` mails whatever addresses the injected function returns. The test's fake
  supplies that list, so the fake, not the code under test, excludes the removed admin. The real filter on
  `Roster.List` and `Admin` lives in `cmd`'s closure, and no fixture tests it.
- **Fold.** Test the `cmd`-side closure against a real roster that holds a demoted admin and a tombstoned admin, or
  have the AC 28 test drive that closure.

### M12. AC 24's "tagio non-sentinel error" fixture never reaches `tagio` (correctness)

- **Location.** T11 `:810-811`.
- **Defect.** "A staged path replaced by a directory" fails at step 1's own `ID3` read through the staging root
  before `tagio.Sniff` runs, and both routes end in `check-unfinished`. So the `tagio` error mapping is never
  exercised, and a mutation that maps every `tagio` error to `corrupt` survives. Also, "a staged file made
  unreadable" (chmod) passes vacuously under root.
- **Fold.** T6 unit-tests the classifier as a table: `ErrMalformedContainer` maps to `corrupt`,
  `ErrUnsupportedContainer` to `bad-extension`, and a wrapped `*fs.PathError` to could-not-run. Relabel the
  directory fixture as the read-fault case. The test skips with a stated reason when it runs as uid 0.

### M13. Owned criteria and states rows without a named fixture (adapter planning-miss gaps b and d)

- **Location.** T6 `:625-631`, T7 `:655-659`, T8 `:695-710`, T11 `:806-822`.
- **Defect.** The plan owns each of these, and no task names a fixture for it:
  - **AC 19, T11.** The restart clauses: a ZIP with an injected stop mid-extraction, and one with a stop after
    extraction but before the verdict commits.
  - **AC 11, T11.** The restart clause.
  - **AC 2, T11.** `serve`'s half.
  - **AC 24.** The create clause, "a `statfs` fault at create answers `could-not-check`". No task owns it: T8 omits
    AC 24, and T11's fixtures cover extraction only.
  - **AC 14, T10.** "Each of the seven values is stored".
  - **The manifest states rows, T6.** An empty manifest (`no-audio`); a source that is absent, empty, or unknown;
    the malformed cases; a `files` key that is absent rather than empty.
  - **The ZIP rows, T7.** A ZIP that does not open, a CRC mismatch, and a ZIP with no audio member.
  - **AC 15's "NULL" cap, T8.** The fixture should name both an absent `member_overrides` row and a NULL column,
    and also an absent `member_quota` row.
- **Fold.** Add each fixture to its task's list, with the reason it would fail.

### M14. The acceptance map is not fully honest about what the plan leaves undone, and the deferred clauses are not filed (correctness)

- **Location.** "Out of scope" `:24-32`; acceptance map `:34-38`; T13 `:864-868`.
- **Defect.** The plan does not state these departures or omissions:
  - It drops AC 22's "the fake MusicBrainz transport records no request" and substitutes "no `matching` event".
  - P8 lowers both timeouts, where AC 9 says "only `ReadTimeout` lowered".
  - AC 24's create clause belongs to no task (M13).

  The plan also splits these off for `intake-2`, and T13 files none of them for its planner:
  - AC 11's client halves, AC 23's two race clauses, AC 26's `route_unlisted` column, and AC 27's `matching` half;
  - P5's `matching` stall and P5's re-queue of `matching`;
  - P6's `RequireMember`;
  - P4 and P16's missing `route` field.
- **Fold.** The map lists each departure in one line. T13 files one `#rung:intake` BACKLOG item that names every
  carried clause and decision.

### M15. Library behaviors are quoted late, or not at all (adapter planning-miss gap a)

- **Location.** `:410-412`.
- **Defect.** The adapter requires every quote "before approval", and the spec says "The plan quotes before writing"
  (`spec:1014-1016`). Three quotes are deferred to execution: `statx` with `STATX_MNT_ID` (T5), `readDirectoryEnd`
  (T7), and memorylocker's `Lock` (M5). Four behaviors are relied on with no quote at all:
  - `tagio.Read`, `Parse`, and `Sniff`'s error contract, which the spec names explicitly;
  - `strings.ToValidUTF8`;
  - mewkiz's encoder hashing through `Frame.Hash` (M7);
  - `golang.org/x/exp/slog` becoming a direct requirement once the mount builds its handler. P14 names only two
    requirements.

  Read at `3783060`, ready to paste: `go doc ./internal/engine/tagio Read` says it "hands back
  ErrUnsupportedContainer unchanged; a recognized format still hands back ErrMalformedContainer unchanged from
  Parse". `Sniff` says "A file shorter than four bytes, or one whose leading bytes match no registered row, answers
  ErrUnsupportedContainer", and returns other read errors wrapped (`seam.go:91-96`). `Parse` takes a path, so T5's
  bound changes its signature or adds a function.
- **Fold.** Add these quotes to "Behaviors" before approval.

## Minor

- **m1. Design departures, stated explicitly** (optional). One decision changes the approved design: **P6**
  (`:277-283`) drops the design's `internal/auth` `RequireMember`. Keep the drop: CLAUDE.md's "nothing is exported
  ahead of a caller" outranks the design, and the PR body already names it, so it is not a stop. One more placement
  differs from the design and is not flagged: **P1 and P7** put the tus mount and wrapper in `internal/intake`, where
  the design's list says `internal/serve`. No behavior changes; give it a PR-body sentence like P6's. P15 reads the
  design's "checked" faithfully. P4 and P16's missing `route` field matches the design's "unlisted route's form" out
  of scope. No other decision changes the design.
- **m2. P5 changes the spec's sweep row** (`:273-274`; `spec:1086` stalls `checking` or `matching`) (correctness). It
  is harmless before launch, but it is a change to the spec's states contract. Record it in the PR body, and file
  it under M14.
- **m3. AC 9's mutation target** (`:743`) (correctness). If the wrapper wraps no writer, "remove the wrapper's
  `Unwrap`" is vacuous. Name `statusCapture.Unwrap` (`internal/serve/serve.go:283`) as the fallback target.
- **m4. OWNER FORK: P3a's Sensitive mark** (`:245-247`) (optional). `marks_test.go:20-24` says "A key joins it only
  by a ruling", and the floor's mark is taken by analogy.
  - Option a: mark it now and ask for confirmation in the PR body.
  - Option b: leave it `NotSensitive` and file it for a ruling.

  Recommendation: option a, since the PR read is a ruling venue. Either way, `storeHomedSensitive`'s comment, which
  speaks only of session and link lifetimes, needs widening for T1's four keys and the floor.
- **m5. The privileged-field row for `DUBPLATE_INTAKE_*`** (`:424`) (correctness) gives store-homed channels, but
  `DUBPLATE_INTAKE_FORMAT_ALLOWLIST` is environment-only (T3 `:509-510`). Split the row.
- **m6. The create order on a `statfs` failure** (`:673-676`) (correctness). Reading free space before the `Update`
  means a failed read answers `could-not-check` even when limits 1 to 4 would refuse first, against "first failing
  limit only" (`spec:1059`). Carry the read error into limit 5's position.
- **m7. AC 15's determinism** (`:700-702`, `:711`) (correctness). "Read the limits outside the `Update`" fails only
  under a lucky interleaving. Name a barrier, such as an injected free-space read that blocks until both creates
  have read.
- **m8. T8's half of AC 13** (`:709-710`) (optional) is vacuous for the collection, state, reservation, and role: the
  Go request type cannot carry them. Keep "uploader from the caller" with two members, and leave the rest to T10's
  HTTP fixture.
- **m9. The `store_url` length bound** (`spec:796`) (correctness) has no number and no fixture. Name one (2048) and
  add an over-length fixture to T10.
- **m10. The state-equals-event helper** (`:693`) (correctness). T4's AC 4 fixture needs it before T8, which lists it
  as its own file. Move it to T4.
- **m11. The track-path accept race** (optional). With `CHECK_WORKERS` up to 8, two workers can accept twin tracks at
  once. A cheap concurrent fixture in T8 proves that the `Update` serializes them.
- **m12. T3's sibling fixture** (`:520`) (optional). Name `/srv/music/lib` and `/srv/music/library` in the acceptance
  list, not only in the mutation.
- **m13. The dependency date** (`:342`) (optional). "Still the latest release (proxy 2026-09-14)" is a month old. It
  still holds today: tusd v2.10.1, mewkiz v1.0.14, and tus-js-client 4.3.1 are each the newest, per `go list -m
  -versions` and `npm view`. Re-date it.

## Over-ceremony

- **o1** (optional). Two `web-auth-security-reviewer` reads cover the same T8 to T11 range: T11's security read, and
  the close's checklist glob. Scope the close read to commits after the `security-read=` sha. That saves about one
  Opus read, roughly 15 to 20 minutes plus its tokens.
- **o2** (optional). The rest of T8's half of AC 13 (m8) is test code with no failure mode. It costs a few minutes.

## Mutations checked against their tests

These hold as written:

- **T1:** all four marks.
- **T2:** all three mutations.
- **T3:** `pipeline`, `HasPrefix` once m12 names the fixture, `EvalSymlinks`, and the template.
- **T4:** all five.
- **T5:** `st_dev` and `stx_mask`.
- **T6:** the sniff order, the `TAG` order, 18-bit, the zero MD5, disc 0, collect-all, and `Statfs`.
- **T7:** all four.
- **T8:** the reservation count, the charge, the true-up, the event, and the compare-and-set.
- **T10:** all five.
- **T11:** the ZIP-archive pair, the idle clock, the repeat mail, and the inventory.

These do not hold: T5's bound placement (M4), T9's `UseIn` (M9) and the floor on PATCH (M8), T11's
recovery-before-lease (M10), and AC 15's limits outside the `Update` (m7).

T6's cancelled-context mutation and T11's "mail inside the write callback" each need their detection named. For T6,
name the fixture's pre-cancelled context. For T11, the fake mailer attempts an `Update`, which a nested `Update` would
refuse.
