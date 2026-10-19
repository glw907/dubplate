# intake-1 plan fold

**Task-minutes: 995 after the fold and its verification fold (760 before; 990 after the first fold). That is 215
minutes over the design's line of 780 (twice its 390), so the plan goes back to Geoff before launch: "Rulings for
Geoff", ruling 1.** Path about 725 minutes (550 before). No fold changes the approved design, and no design scope
moved out of the pass; the plan's two pre-fold departures from the design's file list are named under the ruling.

Target: `docs/superpowers/plans/2026-10-12-intake-1.md` at `8457b10` (the reviews read `3783060`; the plan text is
the same), and its launch prompt `docs/superpowers/plans/2026-10-12-intake-1-prompt.txt`, which needs no change: it
defers to the plan's conventions for every rule the fold touched. Reviews folded:
`2026-10-12-intake-1-plan-review-contract.md` (B1, M1 to M15, m1 to m13, o1 and o2), `-review-mechanics.md` (B1, M1
to M10, F1 to F19; its IDs appear here prefixed `X`), and `-review-risk.md` (findings 1 to 16, here R1 to R16). Fold
agent: Opus 5.5, under `superpowers:receiving-code-review` and `spec-plan-review`'s "The fold". Every finding was
checked against the worktree at `8457b10` (no code change since `08a6fa1`), the module cache (tusd v2.10.1,
mewkiz/flac v1.0.14, mewkiz/pkg errutil, x/exp, x/sys v0.49.0), Go 1.27.2's source, or a probe in a scratch
directory, before it was acted on.

## Convergent roots

Each root folds once; its rows below point here.

| Root | Findings | Verified against | Folded at |
|---|---|---|---|
| A. The `Detail*` constants | B1, XM3 | `internal/ops/errors.go:66-67` ("Detail, when set, is one of the Detail constants"); `document_test.go` `TestDocument_EnumerationsMatchTheCode` (the enum equals `details.go`'s constants both ways); `errors_test.go` `TestDetailConstants` pins the ten; `deck-coverage.ts` check 1b; `deck.ts` `API_DETAILS`; `opshttp.go:172-175` (`decode` answers bad JSON with no detail) | P4 "Three details" (exactly `store_url`, `extra`, `manifest`; bad JSON keeps the shared detail-free answer); Conventions "Skills" (`web/` gains three deck lines); Layering (the gate imports no `ops`; the seam maps); T6 (manifest errors, not details); T10 outcome, files, fixtures, mutation; T13 (the lines join the copy-read item) |
| B. The `tagio` bound | M4, XM4, XM5, XF4 | `container.go:191` (one break condition, before `make` at `:200`); `Parse`'s callers (`read.go:32`, `write.go:71`, `internal/harness/subsonic/fixture/manifest.go:165`); no sentinel for a bound exists | T5 outcome (split order: type-127 break, bound, past-end break, `make`; `ParseBounded` beside an unchanged `Parse`; sentinels `ErrTagsTooLarge`, `ErrPictureTooLarge`), acceptance, mutation retargeted; T6 classifier; Behaviors |
| C. Cleanup's scope and its lock | M5, XM7, R1 | `memorylocker.go:59-72` (no try-lock; `requestRelease` asks the holder to stop; `ErrLockTimeout` on `ctx.Done()`); spec:588-592 and :815-817 (the general "every byte outside each submission's staged set" against its enumerated list; an `uploading` file's unreceived partial is outside "its received files") | T11 "Cleanup" (the closed list, short lock timeout, no-op `requestRelease`, `ErrLockTimeout` leaves the file); States rows; Behaviors quote; T11 fixtures (unlocked partial untouched, extraction mid-sweep, held lock) and mutation; "Spec readings" 1 |
| D. The `UseIn` mutation | M9, XM6 | `handler.go:56-66` (GET and DELETE gated on config flags, so the 405s survive `UseIn`); `unrouted_handler.go:127-136` (`Tus-Extension`), `:1453-1455` (501) | T9 mutation retargeted to the 501 and `Tus-Extension` rows; T9 fixtures; Behaviors |
| E. Race determinism | m7, XM10 | an unsynchronized pair often serializes on its own, so "exactly one commits" can pass under the mutation | T8 AC 15 fixture (the free-space read is the create's last act before `Update`; the injected read blocks until both reach it), mutation annotation; the same barrier for two receipts |
| F. The state-equals-event helper | m10, XF6, R5 | plan Conventions vs spec:1094-1096 ("latest state event's kind"); `stalled` is event-only (spec:284-286); `at` ties under a frozen test clock and `id` is random; `snapshot.go:100` uses `VACUUM INTO`, and SQLite's VACUUM "may change the ROWIDs" (fetched) | Conventions (latest *state* event by `seq`; one injected clock); T4 owns the helper, adds `seq` with `UNIQUE (submission_id, seq)` (probed), a `stalled` fixture and mutation; P4 events in `seq` order; T8 transitions write `seq` |
| G. The lease-before-recovery fixture | M10, XF7a | T11's fixture named no state recovery would change | T11 AC 27 fixture (the test holds the lease over an orphan directory and a missing-bytes `checking` row; both unchanged) |
| H. "Mail inside the write callback" | contract trailing note, XF7b | `go doc ./internal/engine/store Handle.Update` ("on any context not derived from fn's ... waits for the connection fn's own transaction holds") | T11 AC 28 fixture (the fake mailer's short-deadline `Update` on a fresh context records a wait); Behaviors quote |
| I. The tus refusal rows | M8, XF7c | spec:399-416 (nine rows); T9's fixtures covered AC 7 to AC 10 only, so "skip the floor on PATCH" failed no named test | T9 table test over every row with injected free-space read and store fault, the crossing-floor PATCH row; T11 cancelled-mid-PATCH fixture (AC 11's last clause) |
| J. AC 24's create clause | M13 (part), XF8 (part) | no task named a fixture for "a `statfs` fault at create answers `could-not-check`" | T8 fixture and the `statfs`-order mutation (with m6); acceptance map AC 24 T8 and T11 |
| K. The security reads | o1, R15, XM9 | adapter path map (`SKILL.md:91-93`) classes no `internal/intake/**`; T6 and T7 are `ordinary` Sonnet tasks parsing member bytes; T11's read and the close glob both read T8 to T11 | T7 gets a per-task `diff-reviewer` (Conventions); T11's security read covers the T6+T7 dispatch commit through T11 with a fifth question; Close step 2 hands the close reader the `security-read=` sha and scopes it to what that read did not cover; the path-map gap joins "Found, not done" |
| L. The privileged-field rows | m5, XF11 | `DUBPLATE_INTAKE_FORMAT_ALLOWLIST` is an environment threshold (P3), not store-homed; `collections.admission_format_allowlist` had no row | States: the store-homed row narrowed; threshold row; column row; plus R's optional binary-path row (`FileInfoChanges.Storage` nil), nearly free |
| M. Behaviors quoted, not recalled | M15, XF12, XF13, R10, m13 | `go doc` and source reads listed under "Measures" | Behaviors: re-dated versions (probed 2026-10-12), tusd router and extensions, `memorylocker`, `GetUpload` and `writeInfo`, mewkiz `New`, encoder `Hash` and depths, RFC 9639 section 8.2 MD5 rule, `tagio` break and sentinels and `Read`/`Sniff`, olderStore and the SQLite probes, `Handle.Update` in full, `WithoutCancel`, `ToValidUTF8`, `Statx`, `readDirectoryEnd`, the signal context. The deferred-quote paragraph now covers only behaviors not listed |
| N. Server faults classified as member faults | M12, R3 | mewkiz `frame/frame.go:128`, `:234` return the bit reader's error raw beside `fmt.Errorf` format errors (R3's `errutil` claim is wrong for the decode path: `grep` finds `errutil` only in the encoder files, so the classification problem stands on the raw path); `archive/zip/reader.go:294-325` (`ErrFormat` past the size, `io.ErrUnexpectedEOF` short of it) and `compress/flate` `CorruptInputError`; a directory in place of a staged file fails at the gate's own `ID3` read, never reaching `tagio` | P19 (classify by the gate's own I/O through a recording reader; R3's closed `unpack-failed` list corrected to cover `io.ErrUnexpectedEOF` and `CorruptInputError`, which it would have sent to "could not run" and a retry loop); T6 classifier table, `EIO` fixture, mutations; T7 classification, write-failure and `EIO` fixtures, mutation; T11 AC 24 fixture relabeled, uid 0 skip |
| O. T1's four mutations | XF19, R16 | `marks_test.go`: one `slices.Equal` over the exact set | T1: one mutation run. T1 keeps its per-task review: the Conventions review every `auth-data` task |
| P. Design placements | m1, XF17 | design "Files and interfaces" lists `internal/serve` (the tus mount) and `internal/auth` (`RequireMember`) | P6's companion sentence for the tus mount; Layering (`internal/intake` never imports `internal/serve`) |
| S. Shapes the pair halves share | XM8 | T6 and T7 write `internal/intake/gate` in parallel; T9 and T10 both parse the 32-hex id in `internal/intake` | P18 (T4 lands `policy.go`: `Policy`, the format table, the per-file reason value; T8 owns `parseID`); T4, T6, T9 cross-references |

## Findings

### Contract lens

| ID | Severity | Verified against | Disposition |
|---|---|---|---|
| B1 | blocker | see root A | Folded, root A |
| M1 | major | spec:631 (extras get no row), spec:788 (`get_submission` answers warnings); `0006` had no home for them | Folded: T4 `submission_warnings` (`subject` BLOB, no `kind` column); T8 verdict commit writes them; P4 warnings shape; T11 AC 22 reads the extras warning's subject |
| M2 | major | PCF section 4 (`passb-copy-final.md:120-131`: `{friendly}` and "Next: {next_step}" "both verbatim from the API", chips from the wire `label`); salvage `reject.go:177-187` wire shape | Folded: P4 "A reason on the wire" object on `files[].verdict`, `rejects`, events' `reject`, and warnings; T2's caller list; T10 fixture and mutation; End-to-end item 2 and the smoke compare the literals |
| M3 | major | `config.Load` callers: `backup.go:204`, `check.go:30`, `doctor.go:48`, `member.go:58`, `restore.go:152`, `serve.go:221`; spec:123-125 | Folded: P2 `loadConfig` for the five verbs, `doctor` apart; T3 files, `backup` fixture, mutation |
| M4 | major | see root B | Folded, root B |
| M5 | major | see root C | Folded, root C |
| M6 | major | `Makefile:16-33` (`UNCACHED_TESTS` and the reason); `arch_test.go` `serverUnlinkable` | Folded: P10 (uncached entry; two `serverUnlinkable` prefixes); T6 files; End-to-end item 3; Close step 1's oracle vet line, which an `internal/harness` edit brings (adapter "Full-suite home") |
| M7 | major | `encode_frame.go:59` (`f.Hash(enc.md5sum)`); RFC 9639 section 8.2 (fetched) | Folded: P9 (the test computes the expected key from the samples); T6 fixture; Behaviors quote |
| M8 | major | see root I | Folded, root I |
| M9 | major | see root D | Folded, root D |
| M10 | major | see root G | Folded, root G |
| M11 | major | `notice.go:26-53` filters `Admin` over `Roster.List`; `go doc ./internal/auth Roster.List` ("every live member") | Folded: T11 outcome (the `cmd` closure), AC 28 fixture drives it over a real roster with demoted and tombstoned admins, mutation |
| M12 | major | see root N | Folded, root N |
| M13 | major | each listed gap confirmed unfixtured | Folded as fixture lines: AC 19, AC 11, AC 2 restart and `serve` clauses (T11); AC 24 create (T8, root J); AC 14's seven sources (T10); manifest rows (T6, an absent `files` key answering `no-audio` as Go decodes it); ZIP rows (T7); AC 15's three NULL-cap shapes (T8) |
| M14 | major | plan map and "Out of scope"; P8; T13 | Folded: acceptance map names the three departures; T13 files one `#rung:intake` item naming every carried clause and decision |
| M15 | major | see root M | Folded, root M |
| m1 | minor, optional | see root P | Folded, root P (the planning session's guidance; a sentence) |
| m2 | minor | spec:1086 | Folded: P5 sentence; acceptance map; Close step 6 |
| m3 | minor | `internal/serve/serve.go:283` | Folded: T9 mutation names the fallback target |
| m4 | minor, optional, marked owner fork | `marks_test.go:20-24` ("A key joins it only by a ruling") | Taken, not forked: option a, the mark stays with the PR-body sentence asking Geoff to confirm (P3a), since the PR read is a ruling venue and the plan stops on nothing. T1 widens `storeHomedSensitive`'s comment |
| m5 | minor | see root L | Folded, root L |
| m6 | minor | spec:1059 ("first failing limit only") | Folded: T8 carries the read's result into limit 5's place; fixture and mutation |
| m7 | minor | see root E | Folded, root E |
| m8 | minor, optional | Go request type cannot carry those fields | Folded: free and removes vacuous test code; T8 keeps the uploader, T10's HTTP fixture holds the rest |
| m9 | minor | spec:796 names no number | Folded: P4 2048 bytes; T10 2049-byte fixture |
| m10 | minor | see root F | Folded, root F |
| m11 | minor, optional | accept runs inside one IMMEDIATE `Update`; T8's "skip the duplicate check on the track path" mutation covers the path | Dropped: optional, and not nearly free (a second concurrent fixture) |
| m12 | minor, optional | T3 | Folded: free; the two paths named in the acceptance list |
| m13 | minor, optional | `go list -m -versions` and `npm view`, 2026-10-12: v2.10.1, v1.0.14, 4.3.1 newest | Folded, root M |
| o1 | optional | see root K | Folded, root K |
| o2 | optional | see m8 | Folded with m8 |

### Mechanics lens

| ID | Severity | Verified against | Disposition |
|---|---|---|---|
| XB1 | blocker | the reviewer's `metaflac` reading of the seeds (16000 Hz, mono); `MIN_SAMPLE_RATE` floor 44100; probed `ffmpeg -f lavfi -i sine=...:sample_rate=44100 -ac 2 -sample_fmt s16` gives 44100, 16, 2 | Folded: Close step 3 generates CD-quality tracks on the workstation, long in duration for the restart step; End-to-end item 2 |
| XM1 | major | `olderstore_test.go:69-75` drops tables only; probed DROP COLUMN under the index fails, DROP INDEX first succeeds | Folded: T4 undo list, doc comment, two mutations with their true messages; Behaviors probe |
| XM2 | major | `check_test.go:122-129` `baseEnviron` (no staging; library `/srv/music/library`); `serve_test.go:46-49`; `newServeShellVia` (`serve.go:310`) builds the routes every shell test reaches | Folded: T3 (staging in `baseEnviron`; `serveAuthEnviron` makes sibling temp directories; files); T11 (teardown order, two-shell close rule, files) |
| XM3 | major | see root A | Folded, root A |
| XM4 | major | see root B | Folded, root B |
| XM5 | major | see root B | Folded, root B (two sentinels, mapped before the fallback) |
| XM6 | major | see root D | Folded, root D |
| XM7 | major | see root C | Folded, root C |
| XM8 | major | see root S | Folded, root S |
| XM9 | major | see root K | Folded, root K |
| XM10 | major | see root E | Folded, root E |
| XF1 | minor, marked owner fork | the re-estimate below | Owner ruling: "Rulings for Geoff", ruling 1. Option (b), moving AC 28's stall and mail out, is not offered by the fold as its own move: the brief bars moving design scope to fit a number |
| XF2 | minor | `auth_test.go` near `:521` pins `sensitive` false | Folded: T1 outcome and files |
| XF3 | minor | `tagio`'s three functions take a path (`seam.go`, `container.go:133`, `read.go:17`) | Folded: P1 (tagio takes the joined staged path; `os.Root` for the gate's own opens, writes, removals) |
| XF4 | minor | see root B | Folded, root B (`ParseBounded`, so `internal/harness` is untouched by T5) |
| XF5 | minor | `go doc ./internal/engine/tagio Format` (an `int`, no `String`) | Folded: P3 and P18 (the gate's format table maps names; `serve`'s allowlist check reads it) |
| XF6 | minor | see root F | Folded, root F |
| XF7 | minor | see roots G, H, I | Folded, roots G, H, I |
| XF8 | minor | each gap confirmed | Folded: AC 24 create (root J); AC 21 "no tag is rewritten" (T6 byte-identical fixture); the other-member `uploader` body (T10 fixture and mutation) |
| XF9 | minor | `df`: `/var/home` 58 GB free, `/tmp` 7.7 GB tmpfs; `internal/settings/doc.go` seed rule (the seed re-applies while no admin set the key) | Folded: Close step 3 (scratch under `$HOME`, floor seeded at 1 GiB, then at the 1 TiB ceiling, then back) |
| XF10 | minor | `templatekeys_test.go` `TestRenderedTemplateDoctor` first case exits 0 on a fresh render | Folded: States `doctor` row; T11 files (the device row in that case) |
| XF11 | minor | see root L | Folded, root L |
| XF12 | minor | `go doc ./internal/engine/store Handle.Update` | Folded, root M; T8 Accept is a function over the caller's `*store.Tx` |
| XF13 | minor | `encode_frame.go:400-431` | Folded: P9 (patch a 16-bit file's STREAMINFO depth); Behaviors |
| XF14 | minor | spec:596-599 already quotes `readInt` and `safeCastInt` | Refused: the spec quotes it, the plan adopts the spec's quotes by reference, and every dispatch reads the spec sections its task cites |
| XF15 | minor | `git ls-files` at `8457b10` lists the prompt | Refused: already closed; the planning session committed the prompt in `8457b10` |
| XF16 | minor, optional | `internal/opensubsonic/boundary_test.go` runs `go list -deps` under `exec.CommandContext`; `arch_test.go` uses `go list` too (`:239-245`) | Folded: free while P10 was rewritten for M6; P10 cites `boundary_test.go` |
| XF17 | minor, optional | see root P | Folded, root P (the planning session's guidance) |
| XF18 | minor, optional | spec AC 9 text; tusd extends both deadlines | Folded: free; the acceptance map names it as a strengthening, with M14 |
| XF19 | minor, optional | see root O | Folded, root O |

### Risk lens

| ID | Severity | Verified against | Disposition |
|---|---|---|---|
| R1 | major | see root C | Folded, root C |
| R2 | major | `main.go:88` (`signal.NotifyContext` on SIGTERM); `serve.go:229-234` (deferred `h.Close()` after `serve.Run`); `sweepNonces` is started and never joined | Folded: T11 (a cancellation commits nothing and stays `checking`; teardown order cancel, wait, release, close); States row; P19's last sentence; T11 fixture and mutation; Close step 3 and End-to-end item 2 add the SIGTERM restart |
| R3 | major | see root N | Folded, root N, with the `unpack-failed` list corrected |
| R4 | major | spec:576-579 (extraction writes directly) against spec:383-385 (receipt syncs because the store is durable on another device); spec's reconcile tests presence | Folded: T7 syncs each member and `x/` before the verdict `Update`; T11 reconcile tests size against the row's `size`; truncated-member fixture and mutation. The spec is silent on member sync, so this adds to it without changing it |
| R5 | major | see root F | Folded, root F: the assertion's wording and `seq` both. The lighter `ORDER BY at, rowid` is refused because a snapshot's `VACUUM INTO` may renumber rowids |
| R6 | minor | `unrouted_handler.go:45` (413), `:344`, `:489` (`validateMaxSize`); `ApplyNextOperation` on the cap | Folded: P7 (`MaxSize` at the ceiling); States row; T9 fixture and mutation |
| R7 | minor | no task named the hand-off | Folded: T8 "The hand-off to the worker" (the store is the queue; a capacity-1 wake channel poked after commit by receipt and retry); T11 adds the startup pass and the sweep tick |
| R8 | minor | `handler.go:38`, `unrouted_handler.go:1703-1705`, `opshttp/mount.go:78` | Folded: P7 "The id is tusd's id"; States path-shape row; T9 fixtures and mutation |
| R9 | minor | `filestore.go:130-140` (an unparseable `.info` is a plain error, so 500), `:303-348` (`O_TRUNC`, no sync) | Folded: T8 receipt syncs `.info`; T11 startup pass removes both files of an unreceived upload whose `.info` does not parse; States row; T11 fixture. A torn `.info` needs a power loss, which a restart and so the startup pass follow |
| R10 | minor | see root M | Folded: P14 and T0 (`go list -m -versions`, recorded); Behaviors re-dated |
| R11 | minor | `go doc context.WithoutCancel`; `notice.go:26-53` precedent | Folded: free; T8 receipt |
| R12 | minor, optional | P5 | Folded: free and clearly right (a sentence); "Out of scope" and Close step 6 |
| R13 | minor, optional | T8 used "open" and "non-terminal" for one set | Folded: free; "non-terminal" throughout, an `accepted` reservation in the group-cap fixture |
| R14 | minor, optional | `TakeKindLease` lives under one state directory | Folded: free; T12's runbook gains one sentence. The extra flock is refused (machinery for a misconfiguration the box's unit cannot make) |
| R15 | minor, optional | see root K | Folded, root K |
| R16 | minor, optional | see root O | Folded, root O |

## Spec readings (no erratum)

The ratified spec is not edited, and the adapter's rule is that a frozen ratified spec gets no errata and there is
never an owed list. Each place the plan reads the spec past its literal wording, and why:

1. **Cleanup's scope.** "It removes every byte outside each submission's staged set" (spec:815-816) would, read
   alone, remove an `uploading` submission's unreceived partial tus file (outside "its received files") and a
   `checking` ZIP's `x/`, breaking resume and rejecting a good album. The colon's enumerated list that follows is the
   governing text, and T11 removes exactly that list plus the sweep's orphan `tus/` file.
2. **The malformed row's detail.** "malformed (bad JSON, negative size, duplicate path) → `invalid_request` with a
   detail" (spec:1055): a body that is not JSON is refused by `opshttp`'s shared decoder before the seam, detail-free,
   as on every operation; the two manifest faults carry `manifest`.
3. **ZIP read errors.** "A ZIP that does not open, a member read that errors, and a CRC mismatch answer
   `unpack-failed`" (spec:549) beside L4's "a read error on a staged file ... moves the submission to
   `check-unfinished`" (spec:631-634): the archive is a staged file, so an I/O error reading it is L4's, and
   `unpack-failed` is an error the ZIP format or its decompressor raises (P19).
4. **The stall row.** The spec's sweep row stalls `checking` or `matching`; with no probe, P5 stalls `checking` only,
   named in the PR body.
5. **`0006`'s two additions.** The spec's "Migration `0006`" names three tables. The plan adds
   `submission_events.seq` (an unordered "latest" under a frozen test clock and random ids, and rowids that
   `VACUUM INTO` may renumber) and `submission_warnings` (the warnings `get_submission` returns had no storage, and
   an extras warning names members that get no row). Both are additive and change no table's meaning.

## Rulings for Geoff

1. **Approve the 995 task-minute estimate?** The design says a plan over twice its 390-minute figure (780) comes back
   before launch. The drafted plan sat at 760; the fold adds 230 and its verification fold 5 (the whole-chain
   metadata bound in T5): honest re-estimates where the reviews found missing work (the serve-test environments and
   `loadConfig` in T3, the older-store undo and the warnings table in T4, the tus refusal table in T9, the details
   and reason objects in T10) and the proofs and fixtures that keep correctness-ranked majors from passing vacuously
   (T6, T7, T8, T11). The trims taken are the reviews' named over-ceremony: T1's three extra mutations and the
   duplicated close security read.
   - **Yes (recommended).** Launch at 995 (path about 725). Each task still stops at twice its own estimate, and
     `admin-api-1` ran about 68 percent of its plan (475 of 695).
   - **No.** No cut that keeps the design reaches 780: the fold found no over-ceremony left to trim, so a No means
     Geoff naming about 215 minutes of design scope to move to `intake-2`. The mechanics lens's candidate, AC 28's
     stall event, its admin mail, and the `STALL_AFTER` key (about 25 minutes of T11), is one piece of that and
     leaves the plan about 190 over. The fold takes no cut itself: the brief bars moving design scope to fit a number.
   - **Tokens.** The design's line also names 9.0M tokens (flag 7.2M). The plan carries no separate token estimate:
     `admin-api-1`'s PR #4 score records minutes and gate time, not tokens, so the fold has no measured ratio to scale
     from, and the minutes figure stands in for both halves of the line.
   - **For the same read (informational, not a fork).** The plan builds less than the design's file list in two
     places, each a PR-body sentence: `RequireMember` waits for its first caller, the uploader page in `intake-2`
     (CLAUDE.md, "nothing is exported ahead of a caller"; P6); and the tus mount and its wrapper live in
     `internal/intake`, registered through the shell, not in `internal/serve`, with no `internal/serve` production
     change (P1, P7, P8).

## Measures

- Plan length: 916 lines before, 1199 after the first fold (see "Verification fold" for the final count). Prompt: 79
  lines, unchanged.
- Task-minutes: 760 before, 990 after the first fold, 995 after the verification fold. Per task: T0 25 (25), T1 15
  (15), T2 35 (35), T3 65 (45), T4 80 (55), T5 45 (35), T6 100 (75), T7 70 (45), T8 95 (75), T9 90 (70), T10 85 (60),
  T11 145 (85), T12 35 (35), T13 25 (20), Close 85 (85); the figure in parentheses is the draft's. Path: 550 to about
  725.
- Counts: 77 findings (31 contract, 30 mechanics, 16 risk). 73 folded (R5 in full with its lighter alternative
  refused, R14 with the minimal fold and its flock refused, and the m4 owner fork taken as option a among them); 2 refused (XF14, XF15); 1 dropped (m11); 1 owner ruling (XF1, the clock). Convergent roots: 17.
- New mechanisms folded, each with its source and the defect it answers:
  - `submission_events.seq` under a unique index: probed with `sqlite3` 3.53.4 (the `MAX + 1` insert and the
    refused duplicate); the defect is an unordered "latest" under a frozen test clock and random ids, and SQLite's
    documented rowid renumbering under `VACUUM INTO`, which the store's snapshots run.
  - `submission_warnings`: a plain STRICT table under the store's own invariants; the defect is warnings the spec's
    `get_submission` returns with no storage (contract M1).
  - The olderStore undo list: probed (`DROP INDEX`, then `DROP COLUMN`); the defect is `TestOlderStore` going red
    on `0006`'s column (XM1).
  - The recording reader of P19: a plain `io.Reader` and `io.ReaderAt` wrapper (`checksumReader`,
    `archive/zip/reader.go:294-325`, wraps and hashes a reader the same way, though it records no error; the
    verification read probed a recording wrapper capturing a raw `EIO` through `zip.Reader`); the defect is a server I/O fault rejecting a member's album and deleting its
    bytes (R3, M12).
  - The store-as-queue with a capacity-1 wake channel: a non-blocking send, the stdlib `select` idiom; the defect is
    an unnamed hand-off whose plausible shapes drop or block (R7).
  - The member and `.info` syncs: `internal/atomicfile`'s precedent and the spec's own receipt rationale; the defects
    are R4's silent corruption and R9's stranded upload.
  - `ParseBounded` and the two `tagio` sentinels: `tagio`'s existing sentinel pattern (`ErrMalformedContainer`);
    the defect is a bound the walk could never reach and an error the gate would misclassify (root B).
  - The `MaxSize` ceiling: tusd's own 413 path quoted; the defect is a stranded submission after a cap raise (R6).
  - The uncached cross-compile entry and the two `serverUnlinkable` prefixes: the Makefile's existing mechanism and
    reason; the defect is a cached pass replaying (M6).
- New mechanisms refused: R14's extra flock on staging (a misconfiguration the box's unit cannot make; the runbook
  sentence stands instead); R5's `ORDER BY at, rowid` (implicit, and rowids may move under `VACUUM INTO`); a
  separate concurrent accept fixture (m11).
- Quotes and probes run for this fold: tusd v2.10.1 `handler.go`, `unrouted_handler.go`, `memorylocker.go`,
  `filestore.go`, `config.go` (`x/exp/slog`); mewkiz/flac v1.0.14 `flac.go`, `frame/frame.go`, `encode_frame.go`,
  and an `errutil` grep; x/exp's `slog.Logger` is its own struct; x/sys v0.49.0 `STATX_MNT_ID` and `Statx_t`; Go
  1.27.2 `archive/zip/reader.go`, `compress/flate/inflate.go`; `go doc` on `Handle.Update`, `Roster.List`,
  `tagio.{Format,SupportedFormats,Sniff,Read}`, `context.WithoutCancel`, `strings.ToValidUTF8`, `unix.Statx`; RFC
  9639 section 8.2 and `sqlite.org/lang_vacuum.html` (fetched 2026-10-12); `sqlite3` probes of DROP COLUMN and
  `seq`; `ffmpeg` and `metaflac` on a generated track; `go list -m -versions` and `npm view`; `df`.

## Verification fold

The verification read (`2026-10-12-intake-1-plan-fold-verification.md`, against `d5e8f79`) found 0 blockers, 1
major, 5 minor (correctness), and 3 minor (optional). Each was checked against the tree, or probed, before it was
acted on.

| ID | Severity | Verified against | Disposition |
|---|---|---|---|
| V1 | major | spec:577-582 ("no chain totalling more than four times the picture cap"); `container.go:52-64` and `:238` (every block's `Payload` kept, whatever its type) | Folded: T5 bounds the sum of every block's declared length at four times `MAX_PICTURE_BYTES` from the header, answering `ErrPictureTooLarge`; padding-chain fixture and the "count only picture blocks" mutation; Behaviors quotes `Block.Payload`; P3 names the derived figure. T5 40 to 45 |
| V2 | minor | `doctorprobe_test.go:386-395` runs `newServeCmd` over `doctorEnviron` (`doctor_test.go:978-984`, on `authEnviron`, on `baseEnviron`'s absent `/srv/music` paths) | Folded: T11's files gain `doctorprobe_test.go`, with its sibling temporary directories |
| V3 | minor | T8 receipt reads no free space, so the create's barrier names nothing on the receipt path | Folded: T8 names receipt's own barrier (an injected sync, its last act before the `Update`) and the mutation "read the received set outside receipt's `Update`" |
| V4 | minor | fold record ruling 1; design:52-53; PR #4's score rows (minutes and gate time, no tokens) | Folded: ruling 1 says no design-preserving cut reaches 780, so No means naming about 215 minutes of scope; the token figure is stated with why no estimate is derived; one informational line names the two file-list departures. The plan's clock paragraph says the same |
| V5 | minor | `Makefile:22-27` (the variables hold names; the packages are the recipe's literal list) | Folded: P10 names both edits |
| V6 | minor | `.golangci.yml:17-25`; `check.sh:812-877` (`CHECK_EXPORTED_UNUSED` defaults to `warn`); a scratch module probed with golangci-lint 2.14.0 under `unused` and `exported-fields-are-used: false`: an exported field nothing reads, and one referenced only in a test, both "0 issues" | Folded in part. The lint half is refused as written: the probe shows `unused` does not fail on an unread exported field, so T4's gate stays green; P18 now says so and that `check.sh`'s warn-only leg lists the shapes until T6 and T7 merge. The vocabulary half is folded: T4's helper holds the state set as its own literal, and T8's closed vocabulary lives once in `submission.go`, checked against it |
| V7 | minor, optional | P2 against the States row | Folded: free; the row names all five verbs |
| V8 | minor, optional | design "Files and interfaces"; P6; T4 | Folded: free; "Spec readings" 5 names `seq` and `submission_warnings`, and the ruling carries the informational line (with V4). The fold record's headline now says no fold changes the design and points at the two pre-fold departures |
| V9 | minor, optional | `details.go:3` ("The ten values"), `errors_test.go:174` | Folded: free; T10 names both comments (thirteen) |

The verification read's note on the record's `checksumReader` citation is also taken: "Measures" now calls it a
nearby wrapper, not the pattern, and cites the verification's own `EIO` probe.

- Plan length after this fold: 1222 lines (1199 before it).
- Task-minutes: 990 to 995 (T5 40 to 45); path unchanged at about 725 (T4 still sets the T4 and T5 pair's length).
- Counts: 9 findings; 8 folded in full, 1 folded in part (V6, its lint half refused on the probe). No new owner fork;
  ruling 1 stands with its evidence completed.
- New mechanisms: none beyond V1's bound, which is arithmetic over the header fields the walk already decodes; the
  receipt barrier reuses the create barrier's shape.
