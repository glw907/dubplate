# intake-1 plan review: mechanics and feasibility

**Target:** `docs/superpowers/plans/2026-10-12-intake-1.md` at `3783060` (worktree `/var/home/glw907/Projects/dubplate/.claude/worktrees/intake-1-plan`). Nothing in the code changed between `08a6fa1` and HEAD (`git diff --stat 08a6fa1 HEAD` touches only the plan), so the plan's pointers were checked against the code as it stands now. Reviewer: the `plan-reviewer` agent (Opus, medium); saved by the planning session.

**Method.** I read each cited file and line in the repo, the tusd v2.10.1 and mewkiz/flac v1.0.14 sources in `~/go/pkg/mod`, and the Go 1.27.2 `archive/zip` source. I also ran some probes:
- SQLite 3.53.4 DROP COLUMN and DROP INDEX behavior;
- `filepath.Rel` and `filepath.IsLocal`;
- a cross-compile of `tagio` and `ops`;
- the stream format of the tone seeds;
- free space on the workstation.

**Verdict.** Most citations are accurate. A few line numbers have drifted by one or two lines, which the plan already tolerates by re-finding pointers by content. The defects are in mechanisms, proofs, and files the plan leaves out. One blocker (the smoke album) and ten major findings follow, ranked by consequence.

---

## Blocker

### B1. The smoke album cannot reach `matching`: the tone seeds are below CD quality
- **Severity / tag:** blocker, `correctness`.
- **Where:** plan:43-44 (the end-to-end check) and plan:889-891 (Close step 3).
- **Defect:** the smoke builds its album "from the tone seeds (`internal/harness/subsonic/data/fixture/seeds/`)".
  - `metaflac` on `tone-220.flac`, `tone-330.flac`, and `tone-440.flac` reports a 16000 Hz sample rate, 16 bits, and **1 channel**.
  - `DUBPLATE_INTAKE_MIN_SAMPLE_RATE` has both its default and its floor at 44100 (plan:228), so no setting can admit these files.
  - Every track is therefore rejected `below-cd-quality`. The step "read `matching` with every file passing" cannot pass, and neither can the design's one end-to-end check.
- **Fold:** generate the smoke tracks at 44.1 kHz, 16-bit stereo. `ffmpeg` and `flac` are installed (`/home/linuxbrew/.linuxbrew/bin/`), so for example `ffmpeg -f lavfi -i sine=frequency=440:sample_rate=44100:duration=600 -ac 2 -sample_fmt s16 t1.flac` works. Alternatively, reuse T6's Go generator. Tag the tracks with `metaflac`.
  - Make the "check runs over a second" album long in duration, not in bytes: decode time scales with sample count, and a FLAC of a sine wave compresses to almost nothing.

---

## Major

### M1. `olderStore` cannot undo `ALTER TABLE collections ADD COLUMN intake_route`
- **Severity / tag:** major, `correctness`.
- **Where:** plan:381 and plan:540-541 / 568-571 (T4). Code: `cmd/dubplate-server/olderstore_test.go:69-75`.
- **Defect:** the plan says the older-store helper only "needs a `6:` row, referencing tables first". But the helper only runs `DROP TABLE <name>` for each version above the target.
  - `0006` adds a column and a unique index to `collections`, a table from version 3.
  - For targets 3, 4, and 5, `collections` survives the cut with `intake_route` and its index still in place. Re-migration then fails on "duplicate column name" (or "index already exists"), so `TestOlderStore` goes red however the `6:` row is written.
  - The T4 mutation note ("drop the `6:` row ... fails 'table already exists'") also mis-predicts the failure message.
- **Probe (sqlite3 3.53.4):**
  - `ALTER TABLE c DROP COLUMN intake_route` while the unique index exists fails with "error in index c_route after drop column".
  - `DROP INDEX` first, then `DROP COLUMN` (the column's own CHECK is fine), then re-adding the column all succeed.
- **Fold:** T4 extends the helper with a per-version undo list beside `migrationTables`, for example `6: {"DROP INDEX <idx>", "ALTER TABLE collections DROP COLUMN intake_route"}`. Update the helper's doc comment and the mutation's expected message, and add about 10 minutes to T4.

### M2. Making staging required, and adding the device check, breaks every cmd test environment; no task lists the helpers
- **Severity / tag:** major, `correctness`.
- **Where:** T3 Files (plan:512-515) and T11 Files (plan:802-804).
- **Defect, part 1 (T3):** `DUBPLATE_STAGING_DIR` becomes a required owned wall row, so `config.Load` refuses every environment that lacks it.
  - The shared base is `baseEnviron()` (`cmd/dubplate-server/check_test.go:122-129`). It is reached through `authEnviron` and `serveAuthEnviron` (`serve_test.go:46`) and other helpers.
  - There are 84 call sites. `check_test.go` and `doctor_test.go` are in neither T3's nor T11's file list.
- **Defect, part 2 (T11):** `serve` runs the startup device check while it builds the shell (tus registration happens in `newServeShellVia`, `serve.go:310`).
  - `baseEnviron` points the library root at `/srv/music/library`, which does not exist on the workstation (`ls`: no such file).
  - So every test that builds a serve shell would be refused at start: `serve_test`, `routes_test` (`TestRouteInventory` builds the real shell), `apimount_test`, `serve_store_test`, `restore_test`, and the rendered Playwright leg (`renderedEnviron` goes through `serveAuthEnviron`).
  - The `intake` kind lease adds a second constraint: two shells on one state directory in one process will collide unless the lease is released on close.
- **Fold:**
  - T3 adds a sibling `DUBPLATE_STAGING_DIR=/srv/music/staging` to `baseEnviron` and lists `check_test.go` and `doctor_test.go`.
  - T11 makes `serveAuthEnviron` create a temp library root and staging directory side by side under one `t.TempDir()`, so they share a filesystem and mount. It lists `serve_test.go` and states that the lease is released when the store handle closes.
  - Re-estimate both tasks (see F1).

### M3. P4 says no Detail constant is added, but T6, T10, and the spec all emit details
- **Severity / tag:** major, `correctness`.
- **Where:** plan:266, against plan:605 (T6: `invalid_request`, detail `extra`) and plan:766 (T10: detail `store_url`). The spec's states table also has "malformed ... `invalid_request` with a detail" (spec:1055, 1058, 1078).
- **Defect:** three existing checks force these to be real constants with copy:
  - `ops.Error.Detail` "is one of the Detail constants" (`internal/ops/errors.go:66`).
  - `TestDocument_EnumerationsMatchTheCode` holds the openapi detail enum equal to the `Detail*` constants in both directions (`opshttp/document_test.go:392-428`), and `errors_test.go:179-188` pins the set.
  - The `deck-coverage` web gate (check 1b, `web/scripts/gates/deck-coverage.ts`) requires an `API_DETAILS` line in `web/src/lib/copy/deck.ts` for every `Detail*` constant.
- So either the details exist as constants, which means a `web/` copy edit the plan forbids ("`web/` is touched only by T0's one pin", plan:131-132), or they cannot be emitted at all.
- **Fold:** T10 adds `DetailExtra`, `DetailStoreURL`, and one manifest-malformed detail to `details.go`, `errors_test.go`, and `openapi.yaml`. It also adds their `deck.ts` lines, drafted to the register and filed in T13's `intake-2` copy-read item.
  - Amend P4's sentence and the `web/` line in Conventions to match.
  - T6 then returns its detail through `ops`, which is safe: `internal/ops` imports no `store` and cross-compiles for windows (probed).

### M4. T5's truncated-file fixture never reaches the bound, and the mutation that should catch it cannot fail
- **Severity / tag:** major, `correctness`.
- **Where:** plan:362-365, 584-596.
- **Defect:** `container.go:191` breaks the walk silently on `blockType == 127 || bodyOffset+length > size`, and this happens before `make` at `:200`.
  - A block whose declared length runs past a truncated file therefore breaks out of the walk before any bound placed "after the type-127 test" can run. The fixture gets `ErrMalformedContainer` (or nothing at all) under both placements of the bound.
  - On a file that is not truncated, `ReadAt` succeeds, so "bound before allocation" and "bound after allocation" return the same error. The ordering cannot be observed.
- **Fold:** state that T5 splits `:191`'s condition into three steps in this order: the type-127 break, then the bound check (returning the bound's error), then the past-end-of-file break. The truncated fixture then returns the bound's error. The mutation "move the bound below the past-end-of-file test (or below `make`)" then returns malformed, so the test fails as intended.

### M5. T6 would classify T5's bound error as "could not run"
- **Severity / tag:** major, `correctness`.
- **Where:** plan:610-611.
- **Defect:** T6 maps "a `tagio` error other than its two sentinels" to "could not run" (which leads to `check-unfinished`).
  - T5's new bound error is a third error class. As written, a file with an oversized tag or picture block lands in `check-unfinished`, never `tags-too-large` or `picture-too-large`, and a retry loops on it forever.
- **Fold:** T5 exports one sentinel (or a typed error that names the bound that tripped) with the gate as its named caller. T6 maps it to the two reject codes before applying the could-not-run fallback. Quote the existing sentinels: `ErrMalformedContainer` (`container.go:35`) and `ErrUnsupportedContainer` (`seam.go:96-104`).

### M6. T9's `FileStore.UseIn` mutation does not fail on the 405 rows
- **Severity / tag:** major, `correctness`.
- **Where:** plan:743-744.
- **Defect:** in tusd's router (`handler.go:56-66`), GET reaches `GetFile` only when `!config.DisableDownload`, and DELETE reaches `DelFile` only when `UsesTerminater && !DisableTermination`. Both flags are set either way, so composing with `UseIn` still answers 405, and the mutation survives.
  - The behavior `UseIn` does change: the length deferrer turns `Upload-Defer-Length: 1` from 501 (`unrouted_handler.go:1453-1455`, "`lengthIsDeferred && !handler.composer.UsesLengthDeferrer` → `ErrNotImplemented`") into an accepted request.
  - It also adds `concatenation` and `creation-defer-length` to the `Tus-Extension` header (`:127-136`).
- **Fold:** name the killing row: AC 7's "501 to `Upload-Defer-Length`", plus an OPTIONS `Tus-Extension` equal to exactly `creation,creation-with-upload`.

### M7. `memorylocker` interrupts the lock holder; cleanup's "leave a held one for the next sweep" relies on behavior the plan never quotes
- **Severity / tag:** major, `correctness`.
- **Where:** plan:780-781 (T11).
- **Defect:** `memoryLock.Lock` (`pkg/memorylocker/memorylocker.go:59-72`) calls `entry.requestRelease()` on the current holder, then waits for `lockReleased` or `ctx.Done()` (returning `handler.ErrLockTimeout`).
  - So cleanup taking the lock asks tusd's in-flight PATCH to stop. There is no non-blocking try; "leaving a held one" only happens on cleanup's own context timeout.
  - This is a library behavior the task relies on, and the plan does not quote it.
- **Fold:** quote it in "Behaviors this plan relies on". State three things:
  - cleanup locks only uploads of terminal or closed submissions, where interrupting the upload is intended;
  - it passes a short context and treats `ErrLockTimeout` as "leave it for the next sweep";
  - its own `requestRelease` is a no-op, because cleanup holds the lock only for the removal.

### M8. Pair halves need types from each other
- **Severity / tag:** major, `correctness`.
- **Where:** plan:637-638 (T6+T7) and plan:746 (T9+T10).
- **Defect:**
  - T7's walk needs the caps (`MAX_FILES`, the album cap) and a per-file reason type. Its separate-tracks extras need the audio-extension rule. All of these are what T6 defines in `gate.go` and `manifest.go`, in the same package, at the same time.
  - T9's wrapper and T10's seam both parse the 32-hex wire id inside `internal/intake`.
  - Two parallel worktrees will each declare their own copies, and the merge then fails to compile (duplicate identifiers) or ships two divergent types.
- **Fold:**
  - Fix the shared shapes before the pairs run. T2 owns the per-file reason value in `internal/reject`, and T4 owns the gate's policy struct and the extension rule beside `path.go`. Alternatively, T7 takes scalar limits and returns `[]reject.FileReason`.
  - T8 owns `parseSubmissionID` (`submission.go`), and both T9 and T10 consume it.

### M9. The two untrusted-input parsers get the lightest review in the pass
- **Severity / tag:** major, `correctness`.
- **Where:** T6 (plan:621), T7 (plan:654), and the security read range (plan:795).
- **Defect:** the FLAC decode and the ZIP walk run on member-supplied bytes. T7 hand-ports `readDirectoryEnd`'s offset arithmetic, which is binary parsing of an attacker-controlled archive.
  - Both tasks are `ordinary`, run on Sonnet, and get no per-task review.
  - T11's security read covers only T8 to T11.
  - The adapter's `auth-data` path map (`.claude/skills/dubplate-pass/SKILL.md:91-93`) does not include `internal/intake/**`, so the close review's `web-auth-security-reviewer` glob does not load for the gate, the wrapper, or the seam either.
- **Fold:** give T7 (at least `zipcount.go` and the extraction) a per-task `diff-reviewer`, extend the security read's range to T6 through T11, and file the gap in the adapter's path map (`internal/intake/**`) under "Found, not done".

### M10. AC 15's race proof can pass under its own mutation
- **Severity / tag:** major, `correctness`.
- **Where:** plan:700-701 and 711 (T8).
- **Defect:** "two concurrent creates ... exactly one committing". Under the mutation "read the limits outside the `Update`", two goroutines often run one after the other anyway: the second reads after the first commits. The test then still sees one commit, and the mutation survives at random.
  - The same applies to "two files finishing at once move the submission once".
- **Fold:** name a deterministic interleaving. Use a test hook that blocks after the limit read until both creates have read, or run both limit reads first and then both writes in the mutated shape. Record which proof kills the mutation.

---

## Minor

### F1. OWNER FORK: the honest re-estimate likely crosses the 780-minute line
- **Severity / tag:** minor, `correctness`.
- **Where:** plan:51-57.
- **Defect:** the plan sits at 760, 20 minutes under the stop line. M1 (about +10 on T4), M2 (about +10 on T3 and about +25 on T11 for the serve-test environments and the lease), and M3 (about +10 on T10 for the details, contract rows, and deck lines) move it to roughly 815. The 85 minutes for T11, which includes the security read, was already tight.
- **Options:**
  - (a) Carry the re-estimate back to Geoff before launch, as the design's budget line requires.
  - (b) Move scope to `intake-2`. The cleanest cut is AC 28's stall event and admin mail plus the `STALL_AFTER` key, which leaves the sweep's expiries in place.
  - (c) Keep 760 and cite `admin-api-1`'s under-run (475 actual against 695 planned) as the basis.
- **Recommendation:** (a), with (b) offered as the fallback if Geoff wants to stay inside the line. (c) treats a measured history as licence to understate known work.

### F2. T1's file list misses `auth_test.go`
- **Severity / tag:** minor, `correctness`.
- **Where:** plan:465.
- **Defect:** `cmd/dubplate-server/auth_test.go:521` (`TestRegisterAuthConfig_Settings`) pins `sensitive: false` for `DUBPLATE_UNKNOWN_ADDRESS_ANSWER`. T1's gate goes red until it changes.
- **Fold:** add `auth_test.go` to T1's Files. `backup_test.go` pins no marks; I checked.

### F3. The tagio API takes paths, not an `os.Root`
- **Severity / tag:** minor, `correctness`.
- **Where:** plan:197 and 606.
- **Defect:** `tagio.Sniff`, `Parse`, and `Read` all open a path string (`seam.go:83`, `container.go:133`, `read.go:17`), so the per-file checks cannot run "over an `os.Root`" as written. The spec asks for `os.Root` only on writes, renames, and removals (spec:141).
- **Fold:** say that tagio calls take `filepath.Join(staging, <server-minted name>)`, while `os.Root` covers writes, removals, and the decode's open.

### F4. The `Parse` signature change touches files T5 does not list
- **Severity / tag:** minor, `correctness`.
- **Where:** plan:585.
- **Defect:** "a bound `Parse` takes from its caller" changes `Parse`, which has three callers: `read.go:32`, `write.go:71`, and `internal/harness/subsonic/fixture/manifest.go:165`. None is in T5's Files. Touching `internal/harness` also triggers the adapter's close obligation to run `go vet -tags oracle` (SKILL.md, "Full-suite home").
- **Fold:** add a bounded entry point instead (for example `ParseBounded(path, Bounds)`) that leaves `Parse` unchanged, or list the three files and the oracle vet line.

### F5. `tagio.Format` has no name
- **Severity / tag:** minor, `correctness`.
- **Where:** plan:240-244 (P3).
- **Defect:** `Format` is an `int` with no `String` method (`go doc` on it lists none), so the value "flac" cannot be compared with `SupportedFormats()` or with `collections.admission_format_allowlist` without a mapping.
- **Fold:** T5 adds `Format.String()`, with the allowlist check as its named caller, or `intake.go` holds the map. Name which.

### F6. The state-equals-event helper has two owners
- **Severity / tag:** minor, `correctness`.
- **Where:** plan:116-119, 559, and 693.
- **Defect:** the conventions say T4 proves the shared assertion can fail, but T8 lists the helper among its own files. T4 runs first and needs the helper to exist.
- **Fold:** T4 owns the helper (for example `internal/intake/stateevent_test.go`), and T8 reuses it.

### F7. Three proofs name no fixture state their mutation would change
- **Severity / tag:** minor, `correctness`.
- **Where:** plan:814, 824-825, and 745.
- **Defects and folds:**
  - **AC 27, "refused at the lease before any recovery runs":** the fixture needs a row that the startup pass would change, written after the first `serve` has started; otherwise the first serve's own recovery already consumed it. Name that row.
  - **"Mail inside the write callback":** the `Send` fake must observe the open transaction (for example by attempting a store read that would block), or the mutation is invisible. Name that fake.
  - **"Skip the floor on PATCH":** this needs an injected free-space read that crosses the floor between the create and the PATCH. T9's fixture list does not have one.

### F8. Some spec criteria have no fixture in any task
- **Severity / tag:** minor, `correctness`.
- **Where:** plan:806-813 and 762-767.
- **Gaps:**
  - AC 24's "A `statfs` fault at create answers `could-not-check`" appears only in T8's outcome text, with no fixture.
  - AC 21's "no tag is rewritten" has no fixture.
  - T10's fixtures do not list a create body that names another member as `uploader`, which the T10 mutation "read the uploader from the body" needs.
- **Fold:** add one fixture line for each.

### F9. The smoke's free-space floor values are unspecified
- **Severity / tag:** minor, `correctness`.
- **Where:** plan:887-893.
- **Defect:**
  - The default floor is 5 GiB and the smallest allowed floor is 1 GiB, so the first two smoke creates also need the scratch filesystem to hold more than the floor.
  - Raising the floor at the restart works through the environment seed only while no admin row exists (the seed rule, `internal/settings/doc.go:5-7`), and only up to the 1 TiB ceiling.
  - The workstation's `/tmp` has about 6.25 GB available; `/var/home` has 64 GB.
- **Fold:** name the scratch filesystem. Seed `DUBPLATE_INTAKE_FREE_SPACE_FLOOR` at its 1 GiB minimum for the first steps, then restart with a seeded value above that filesystem's free space.

### F10. Doctor's exit status for an absent staging directory is not stated
- **Severity / tag:** minor, `correctness`.
- **Where:** plan:435 and 803-804.
- **Defect:** `TestRenderedTemplateDoctor`'s first case requires exit 0 on a freshly rendered template (`templatekeys_test.go:317-330`). On the workstation `/srv/music/staging` is absent, so the device row must not fail `doctor`. The store's "absent" row follows the same pattern.
- **Fold:** add to the states row that the device row reports the absence without failing `doctor`.

### F11. The privileged-field table mis-describes one key and omits a channel
- **Severity / tag:** minor, `correctness`.
- **Where:** plan:424.
- **Defect:** the row "each `DUBPLATE_INTAKE_*` setting: as every store-homed setting" covers `DUBPLATE_INTAKE_FORMAT_ALLOWLIST`, which is an environment-only threshold, not a store-homed setting. The table also has no row for `collections.admission_format_allowlist`, which the effective allowlist reads.
- **Fold:** add the threshold row ("the environment alone"). Add the column row ("no channel at rung 7; a direct store write", on the spec's `member_overrides` precedent).

### F12. The quoted `Update` behavior is imprecise
- **Severity / tag:** minor, `correctness`.
- **Where:** plan:390-392.
- **Defect:** the plan quotes `Update` as refusing "a nested `Update`". `go doc` says something narrower: a nested `Update` on `fn`'s context fails, but one "on any context not derived from fn's ... waits for the connection fn's own transaction holds until that context ends". Accept is "the `Update` that moves a submission to `accepted`" (plan:687), but the separate-tracks verdict commit is also an `Update`, so it must not call accept as a second `Update`.
- **Fold:** quote the doc precisely, and say accept is a function over the caller's `*store.Tx`.

### F13. 18-bit and 32-bit fixtures cannot come from the encoder
- **Severity / tag:** minor, `correctness`.
- **Where:** plan:298-303 and 626.
- **Defect:** `encodeFrameHeaderBitsPerSample` handles only 0, 8, 12, 16, 20, and 24 bits, and errors with "not yet implemented" otherwise (`encode_frame.go:400-431`).
- **Fold:** quote this, and say those two fixtures patch the STREAMINFO depth of a 16-bit file. The gate refuses them before decode, so the frames never need to match.

### F14. The track-tag projection T6 relies on is unquoted
- **Severity / tag:** minor, `correctness`.
- **Where:** plan:610 and 629.
- **Defect:** "`A1`/`A2`/`B1` answers `missing-track-number`" depends on `tagio.readInt`'s safe_cast projection: `^[+-]?[0-9]+`, so a non-matching value reads 0 (`read.go:294-303`, `encode.go:372`). An executor who does not know this may add its own parsing.
- **Fold:** quote it.

### F15. The launch prompt is untracked at this commit
- **Severity / tag:** minor, `correctness`.
- **Where:** plan:61-64 and 81-82.
- **Defect:** `docs/plans/2026-10-12-intake-1-prompt.txt` exists in the worktree but is not in `git ls-files` at `3783060`. The launch reads it with `git show worktree-intake-1-plan:...`, which would fail.
- **Fold:** commit and push it before the "plan ready to launch" notification.

### F16. Wrong precedent cited for the cross-compile test
- **Severity / tag:** minor, `optional`.
- **Where:** plan:304-307 (P10).
- **Defect:** `internal/harness/conform/arch_test.go` walks imports with `go/parser`; it does not shell out to `go list -deps`.
- **Fold:** cite `internal/opensubsonic/boundary_test.go:58` instead, which runs `go list -deps -test ./...` under `exec.CommandContext`.

### F17. The tus mount's move out of `internal/serve` is not flagged
- **Severity / tag:** minor, `optional`.
- **Where:** plan:284-297 (P7, P8).
- **Defect:** the design lists "`internal/serve` (the tus mount and its wrapper)". The plan puts both in `internal/intake` and changes no `internal/serve` production code, but flags only P6's deferral as a departure from the design.
- **Fold:** add one sentence to the PR body beside P6's. Constraint to state: `internal/intake` must never import `internal/serve`, since P8's in-package test imports `intake`. There is no cycle today; I checked with `go list -deps`.

### F18. AC 9's test lowers more than the spec says
- **Severity / tag:** minor, `optional`.
- **Where:** plan:293-297.
- **Defect:** spec AC 9 says "only `ReadTimeout` lowered", and the plan lowers `WriteTimeout` as well. This is a stronger test, since tusd extends both deadlines (`unrouted_handler.go:180-185` and `926-931`), but it departs from the criterion's text.
- **Fold:** record it as a deliberate strengthening in the PR notes.

### F19. Over-ceremony in T1, about 10 minutes and one reviewer dispatch
- **Severity / tag:** minor, `optional`.
- **Where:** plan:467-471.
- **Defect:** T1 runs four mutations against one `slices.Equal` pin. Any one mark revert proves the pin can fail; the other three re-prove the same comparison. Its per-task `diff-reviewer` reads a four-line mark diff.
- T1 is also a single-deliverable task, and the checklist would fold it into a neighbor (T3 edits the same `marks_test.go` pin). The design names it "the pass's first task", though, so keep it separate.
- **Fold:** one mutation run. Let the per-task review of T3 cover T1's commit, or keep T1's review and name the trade.

---

## Verified as cited
- tusd v2.10.1: the read and write deadline extension (`unrouted_handler.go:180-185`, `926-931`); `NewError`'s body and content type (`error.go:25-37`); `sendError`'s 500 and its empty HEAD body (`:1288-1297`); `absFileURL` with an absolute `BasePath`; `UseIn` composes no locker (`filestore.go:73-79`); `FileInfoChanges` (`datastore.go:61-75`); `NetworkTimeout` defaults to 60s; the callback signatures (`config.go:74`, `:80`).
- mewkiz/flac v1.0.14: the `New` doc, the ID3v2 skip, `Frame.Hash`'s depth switch, frame CRC-16 verification, and `Close` updating the MD5 on a seeker.
- `archive/zip`: `Reader.init` reads headers until the first bad one and compares only 16 bits of the count, `checksumReader` refuses an overrun with `ErrFormat`, and a directory with content gives `ErrFormat`.
- Store: the migration, fixture, hash-pin, `binaryHead`, and future-row mechanics; the invariants; the relative-path CHECK refusing exactly the AC 6 set except a backslash; `TakeKindLease`; the `Update` doc; `NewID`; `snapshot.go:198`.
- Config: the `claim` panic, `requireDescribed`, `Load` requiring every owned row, the superseded-key error, `KindString`'s closed `Values`, and the `BIND_ADDRS` precedent.
- Other code: `opshttp`'s `maxBody` and `decode`, `parseUserID`, `CheckRequest`, `apiVerifier`, `Shell.Register`, `newHTTPServer` and `readTimeout`, the origin check, `statusCapture.Unwrap`, and the T1 target marks.
- Workstation: `mid3v2 --delete-all` (mutagen `_tools/mid3v2.py:394-395`), and `x/sys` v0.49.0's `Statx_t.Mnt_id` and `STATX_MNT_ID`.
- Other docs: T13's CH, SS, and W8.2 line cites; BACKLOG #55, #56, #57, and #71.
- `filepath.Rel` plus `IsLocal`, probed: equal paths give `"."` (local, so refused); `/srv/music/lib` against `/srv/music/library` gives `"../library"` (not local, so accepted). P2's lexical check and T3's `HasPrefix` mutation both hold.
