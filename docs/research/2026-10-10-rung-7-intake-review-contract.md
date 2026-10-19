# Rung 7 spec review, lens: contract and criteria

Target: `docs/superpowers/specs/2026-10-10-rung-7-intake-design.md` at `9fc7c53`, revision 1. Reviewer: fresh
context, the contract-and-criteria lens. It asks four things. Is every promise testable? Does each acceptance
criterion name its fixture and its failure? Can any criterion pass vacuously? Does the spec close the four
planning-miss items? Code and dependencies were read at `9fc7c53`: tusd v2.10.1 from the module cache, `tagio`, and
`store`. Spec line numbers below are `spec:N`. The settled owner and conductor rulings are not reopened.

Ranked by consequence. Correctness gaps come first, then over-ceremony. No blocker: every gap below has a fold that
the spec can absorb in revision 2 without a new owner ruling.

## Correctness gaps

### C1 (major). A completed upload can be stranded, unreceived, and the client reports success

- **Location:** spec:359-361 (`PreFinishResponseCallback`), spec:707-708 (startup recovery leaves `uploading` to tus
  resumption), spec:956 (the states table), AC 26.
- **Defect:** Receipt happens only in `PreFinishResponseCallback`, and tusd never calls it a second time. Suppose the
  callback returns an error, or `serve` stops between the last byte reaching disk and the callback committing (a
  deploy restart is enough). tusd's PATCH on an upload whose offset already equals its size answers at once, without
  finishing it again: `if !info.SizeIsDeferred && info.Offset == info.Size { resp.Header["Upload-Offset"] = ...;
  ... return }` (`unrouted_handler.go`, `PatchFile`, about line 810). The finish path that calls the callback is
  `finishUploadIfComplete` (`:993-1010`), and it runs only from the PATCH that writes the last bytes. tus-js-client's
  retry HEADs the upload, sees the offset equal to the length, and reports success. The file row stays unreceived,
  the submission never reaches `checking`, and the sweep expires it after `RESUME_WINDOW`. The member was told it
  uploaded. Startup recovery re-queues `checking` and `matching` only.
- **Fold:** Make receipt one idempotent function: mark the file received, rename it out of `tus/`, remove its `.info`,
  and move the submission to `checking` when it is the last file. The callback calls it. Startup recovery, and the
  sweep, also call it for every unreceived file of an `uploading` submission whose `.info` shows offset equal to size.
  Add a row to the states table: "tus upload complete, receipt not recorded: received at the next startup or sweep".
  Add an AC whose fixture is a submission with one file complete in `tus/` (`.info` offset equal to size) and its row
  unreceived; after a restart the submission reaches `checking`.

### C2 (major). The duplicate refusal cannot meet its own criterion, and the reroute skips it

- **Location:** spec:530-536, AC 22, spec:586 (`route_unmatched` from `pushed-back`).
- **Defect:** The check runs at gate pass, during the move to the next state. On the main route that state is
  `matching`, not `accepted`, and the check compares only against `accepted` submissions. Take two identical albums,
  A and B, on the main route. A passes the gate and moves to `matching`. B then passes the gate, finds no `accepted`
  twin, and moves to `matching` too. Both probes return `ActionApply`, so both reach `accepted`. The sentence "of two
  identical albums in flight, the second to commit is refused" is false under the stated mechanism. AC 22's race
  ("exactly one is accepted") would fail against the code the spec describes, so the plan would build the wrong
  mechanism first. Separately, `route_unmatched` moves a submission to the unmatched collection with no duplicate
  check against that collection.
- **Fold:** Compare at gate pass against every submission in the target collection that has `album_key` set and is
  not terminal (`matching`, `match-unavailable`, `pushed-back`, `accepted`, and a stalled one). Run the same check in
  the transaction that changes a submission's collection (`route_unmatched`, the reroute). AC 22 gains a fixture of
  two identical submissions that both pass the gate while neither is `accepted`; the second is refused at gate pass.

### C3 (major). `stalled` has no exit, and `match-unavailable` never ends

- **Location:** spec:254-266 (the states table), spec:707-719 (recovery and the sweep), AC 26.
- **Defect:** Three things go wrong with `stalled`. Recovery re-queues `checking` and `matching` only, so a stalled
  submission is never resumed after a restart. No transition out of `stalled` is named, so the worker that finishes
  late has no valid move. The state also forgets whether the submission was checking or matching, which recovery
  needs. "Mails every live admin once" is not tested across two sweeps, and AC 26 runs one. `match-unavailable` has a
  similar gap: the only way out is a member's retry, and the sweep never expires it. A member who never comes back
  holds a reservation, which counts against quota and the floor, and staged bytes, with no end.
- **Fold:** Make the stall an event, on `rerouted`'s precedent (spec:266), and keep the state at `checking` or
  `matching`. Recovery and the worker then need no new transitions, and "once" becomes "no prior stall event", which
  is durable across restarts. PCF section 8's member copy reads the event. Have the sweep expire `match-unavailable`
  under the `pushed-back` hold. An automatic re-probe on the sweep interval is also cheap: MusicBrainz outages are
  transient, and the member's retry stays available. AC 26 runs the sweep twice and asserts one mail.

### C4 (major). The `uploading` state has no member exit and no resume contract

- **Location:** spec:303-304, spec:362-363, spec:378, spec:416-417, spec:688-689, spec:760-761.
- **Defect:** `cancel_submission` is valid only from `pushed-back` ("the states named there"). The only cancel
  during upload is a tus DELETE, which needs a tus upload to exist. Suppose a submission was created but no tus
  creation followed: the page closed, or a laptop slept. It holds a reservation and a `too-many-sessions` slot for
  the full 7-day `RESUME_WINDOW`. Three such submissions lock the member out of uploading for a week. The second gap
  is resume after a reload. The spec relies on `storeFingerprintForResuming` "so a reloaded page can find an
  unfinished upload". A browser cannot reopen a file, so the member must pick the files again. The spec does not say
  how re-picked files bind to the existing submission and its file ids. Without that, a re-pick calls
  `create_submission` again, which takes a second reservation and a second slot. No AC covers reload or resume.
- **Fold:** Make `cancel_submission` valid from `uploading`, `pushed-back`, and `match-unavailable`. Name the resume
  contract: "Your uploads" offers "resume" on an `uploading` submission, and the page matches the re-picked files to
  `get_submission`'s file list by relative path and size. It sends the existing `submission` and `file` ids as tus
  metadata and takes no second reservation. Name the report for a re-pick that does not match. Add an AC with a
  fixture of a submission with one file half sent: a reload and re-pick resume it from its offset on the same
  submission, and use and session count are unchanged.

### C5 (major). Track and disc numbers that are present but malformed get the wrong report

- **Location:** spec:521-528, AC 21, spec:542-543 (`tagio`'s sentinels).
- **Defect:** `tagio.Read` gives a present but non-numeric value as 0, not as absent. `readInt` returns `&n` from
  `safeCastInt`, which "yield[s]" 0 when there is no numeric prefix (`internal/engine/tagio/read.go:296-303`,
  `encode.go:380-390`). A vinyl rip tagged `A1`, `A2`, `B1` therefore reads as track 0 three times. It is refused as
  `duplicate-track-position` rather than `missing-track-number`, so the member is told the wrong fix. The same
  happens to disc 0. "Missing disc reads as disc 1" does not cover disc 0. The spec also leaves three states
  unreported:
  - an album whose tracks start at 2: "a gap" does not say whether numbering must start at 1;
  - a disc gap (discs 1 and 3) or a `DISCTOTAL` mismatch: `tagio` has `Disctotal` (`tagset.go:163`), and no check
    reads it;
  - tracks that disagree on `TRACKTOTAL`.

  spec:543 sends "a `tagio` error other than its two sentinel errors" to `could-not-check`. It never says what the two
  sentinels, `ErrMalformedContainer` and `ErrUnsupportedContainer`, report.
- **Fold:** Read a track number at or below 0 as missing (`missing-track-number`), and a disc at or below 0 as
  disc 1. Quote `safeCastInt` in the behaviors list. Tracks run 1 to n on each disc, and discs run 1 to m; any break,
  a `DISCTOTAL` mismatch, or disagreeing totals is `incomplete-album`. `ErrMalformedContainer` maps to `corrupt` and
  `ErrUnsupportedContainer` maps to `bad-extension`. AC 21 gains one fixture per case, an `A1`/`B1` album among them.

### C6 (major). The ZIP true-up can raise a reservation past the member's quota unchecked

- **Location:** spec:404-406, spec:415-421, spec:434, AC 14.
- **Defect:** A ZIP's reservation is taken from its archive size. The true-up "can raise it", and before extraction
  only the album cap (`too-large`) and the floor are checked. Take a member with a 1 GiB `quota_bytes` who uploads a
  900 MB ZIP whose members add up to 4 GiB, under the album cap. They end with 4 GiB reserved, and the over-quota
  check never runs. Owner ruling 3 rests on "an admin's per-member cap means what it says". This state (cap set,
  true-up over it) has no named report.
- **Fold:** Re-run limits 3 and 4 against the extracted total, inside the write transaction that does the true-up and
  before extraction. If either fails, the submission is rejected with `over-quota`. AC 14 gains a fixture of a member
  with `quota_bytes` set, uploading a deflated ZIP whose archive size fits under the cap and whose extracted total
  does not.

### C7 (major). Ten reject codes and three tus-path states have no acceptance criterion

- **Location:** spec:966-1060 against spec:413-422, spec:461-471, spec:493-499, spec:920-960.
- **Defect:** No AC produces these reject codes:
  - `uploads-off`, `too-many-sessions`, and the group-total `over-quota`;
  - `too-many-files`, `path-too-deep`, and `no-audio` (both the empty-manifest and the ZIP case);
  - `tags-too-large`, `picture-too-large`, and `missing-track-number`;
  - `bad-extension` (manifest and sniff);
  - `no-space` at a tus creation and before a PATCH.

  No AC covers the tus create where `Upload-Length` differs from the declared size, though it is the one link
  between a file's bytes and its reservation (spec:910). No AC covers the 409 for a file already started, or the 401
  on the tus mount. AC 18 also misses the sample-count mismatch against `NSamples` and the "bytes hashed equal file
  size" check. The criteria list is long but uneven, and a code with no producing test can drift unseen.
- **Fold:** Keep the per-case criteria and add one structural criterion. A test enumerates `internal/reject`'s set
  and asserts that each blocking code has at least one gate, create, or tus-path fixture producing it. A coverage map
  in the test names the fixture, and a code added without a fixture fails the build. `symlink` is exempt, being rung
  9's, with the reason written in the map. Also add one AC for `Upload-Length` against the declared size: a fixture
  creation declaring one byte more than the manifest, refused before `.info` is written.

### C8 (major). The race criteria can pass on scheduling luck

- **Location:** AC 13, AC 22, spec:394-396, spec:361 (the last-file move).
- **Defect:** "Two concurrent creates ... exactly one succeeds" holds by construction only if the check and the
  insert run inside one `store.Handle.Update` callback. That function "runs fn inside one IMMEDIATE transaction on the
  write pool" (`internal/engine/store/write.go:44-47`). An implementation that reads use on a read connection and
  then inserts in `Update` is racy, but two goroutines in a test will usually serialize and pass anyway. The same
  holds for AC 22's race. It also holds for an unnamed case: two files of one submission finishing at once must move
  it to `checking` exactly once (spec:361).
- **Fold:** Make it structural. The quota and floor check, the duplicate check, and the receipt function each take
  the `*store.Tx` that `Update` hands its callback, so none can run outside the write transaction, and the compiler
  proves the property. Quote `Handle.Update`'s doc in the behaviors list. Keep the concurrent tests as smoke tests and
  state that they back the type-level proof. Add the last-file receipt to the states table, with "exactly one move to
  `checking`" as its report.

### C9 (major). AC 9 can pass without the wrapper chain it exists to prove

- **Location:** AC 9, spec:340-347.
- **Defect:** The risk is that the shell's `statusCapture`, or the rung 7 wrapper, hides
  `ResponseController.SetReadDeadline` and lets the 30-second `ReadTimeout` cut off a long PATCH. AC 9's fixture is "a
  test server with a short `ReadTimeout`". A bare tusd handler on `httptest` passes that test while the real chain
  fails. On the box, the failure would cut every upload longer than 30 seconds, which is the core feature.
- **Fold:** The fixture mounts the tus subtree through `Shell.Register` and the shell's full middleware, using the
  production server constructor with only `ReadTimeout` lowered. Add the mutation proof: the test fails when the rung
  7 wrapper drops `Unwrap`.

### C10 (major). The refusal wire shapes on the tus path, and for a create with many reasons, are unnamed

- **Location:** spec:336-337, spec:429, spec:442, spec:683, spec:929, spec:933.
- **Defect:** The tus path has three refusals: `no-space` before a PATCH, `could-not-check` when `statfs` fails, and
  a declared-size mismatch at creation ("`invalid_request`-shaped tus error"). None names an HTTP status or a body. The
  page must turn each into PCF copy, and tus-js-client retries some statuses and gives up on others, so the status
  decides whether the member sees a refusal or a silent retry loop. `create_submission` answers "a refusal carrying
  its reject code" (singular, spec:683). The states table says every failing file is listed with its own reason
  (spec:933), and 14a's error body carries one `error.detail` (spec:679). The order between the manifest checks
  (spec:461-471) and the limit checks (spec:413-422) is also unstated. For example, does `uploads-off` short-circuit
  `bad-extension`?
- **Fold:** Name a tus-path refusal as a 4xx status that tus-js-client does not retry (its `shouldRetry` reads the
  status; the plan quotes it), with a JSON body carrying the reject code. Name the create refusal body: album-level
  codes, plus a list of (path, code) pairs, under 14a's error envelope as an additive field. Name the order:
  malformed input (`invalid_request`), then every per-file manifest reason, then album-level manifest reasons, then
  limits 1 to 5, the first failing limit ending the check.

### C11 (major). AC 15's "byte for byte" collides with W8.2, and its authority is ambiguous

- **Location:** spec:646-660, AC 15.
- **Defect:** Two of the nine adopted strings name FLAC: "This isn't a FLAC file, and the library keeps lossless
  FLAC." and "This FLAC is below CD quality." (`_salvage/internal/reject/reject.go:50,56`). W8.2's leak list names
  "FLAC", and its only exceptions are MusicBrainz on the push-back screen and "collection" (W:1124-1135). The spec
  applies W8.2 to the new strings (spec:656-657), so the register read flags the adopted ones while AC 15 requires
  them unchanged. Both gates cannot pass. `bad-extension` "gains its parameter" and names the supported set and a
  conversion to FLAC (spec:658-659), so it cannot be byte-identical either. "The authority" is two sources, the
  salvage file and PCF section 4, and a Go test cannot import `_salvage` (it is invisible to the go tool and retired
  at launch). A golden copied from the implementation would be vacuous.
- **Fold:** Add a W8.2 exception for the format reasons (`bad-extension`, `below-cd-quality`). A member must learn
  the format's name to convert a file, which is the same reason the MusicBrainz exception exists. Record it as a W
  erratum. Name PCF section 4 as the single authority, and in AC 15 exclude `bad-extension`'s parameterized wording
  from byte equality. Make the byte check a one-time comparison that the task's diff-reviewer records against PCF,
  not a permanent test.

### C12 (major). Config load's `EvalSymlinks` preempts the startup table's "absent" row and breaks every verb

- **Location:** spec:107-108, spec:162, AC 1, spec:170 (`doctor`).
- **Defect:** `filepath.EvalSymlinks` fails on a path that does not exist (on the workstation,
  `EvalSymlinks("/nonexistent/staging")` gives `lstat /nonexistent: no such file or directory`). Its doc does not say
  so, and the spec does not quote it. When config load runs it, a missing staging directory is a load error. The
  startup check's "staging directory absent" report, which names `deploy.sh`, never fires. Every verb that loads
  config (`doctor`, the user verbs, `unidecode install`) then fails on a fresh box before `deploy.sh` has run, and
  `doctor` cannot report the condition it exists to report.
- **Fold:** Config load compares the two paths lexically (`filepath.Clean`, then `filepath.Rel` with an `IsLocal`
  test both ways). The symlink-resolved nesting comparison moves into the startup check after both paths are proven
  to exist, as a new row: "staging and library nest after symlinks: start refused". AC 1 gains a fixture where the
  staging path is absent: config loads, and `serve` refuses with the `deploy.sh` message.

### C13 (minor). AC 2 covers four of the startup table's seven rows

- **Location:** spec:160-168, AC 2.
- **Defect:** These rows have no AC: staging absent, not a directory or not writable, and library absent. "`statx`
  returns no mount id" is a state separate from a `statx` error. `statx(2)` "may fail to return fields that were
  requested", so the test is `stx_mask` lacking `STATX_MNT_ID`. The spec names no detection rule and quotes nothing
  for it. "An injected `Stat`" also implies a test-only function variable, which the architecture reader flags.
- **Fold:** Split the check into a pure function over two stat results, tested table-driven over every row, and a
  thin syscall reader. Quote the `stx_mask` sentence. AC 2 lists all seven rows.

### C14 (minor). Several fixtures as written are vacuous or impossible

- **AC 5:** "a pre-existing root that would nest under the allocated one". The allocated root comes from a fresh id,
  so no fixture can predict it unless the id source is injected. Test the no-nesting check directly with two literal
  roots, and keep AC 5's create path for the owner and the single-creation properties.
- **AC 13:** "a create far above any plausible cap succeeds" with the cap NULL runs into `MAX_ALBUM_BYTES` and the
  real floor. Restate it as several submissions whose sum exceeds what a set cap would allow, all accepted, with
  `statfs` injected.
- **AC 14:** the true-up needs a deflated ZIP whose extracted total differs from its archive size. With a stored ZIP
  the two are nearly equal, and the assertion barely tests anything.
- **AC 23:** "the same fixture without the fault passes" leaves the fixture unnamed. Name a passing two-track album.
- **AC 25:** "each action from any other state answers `conflict`" has one fixture, a `pushed-back` submission. Make it
  a table of states against actions with the expected transition or `conflict` in each cell, after C3 and C4 fix the
  valid sets.
- **AC 4:** the out-of-step fixture proves the checker works, not that the state machine keeps state and events in
  step. Run the same assertion at the end of every journey test (AC 22, 24, 25, 26).
- spec:964's claim that every criterion names its fixture is false for AC 1, 15, 23, and 29. Fix the line or the
  criteria.

### C15 (minor). "Every staging write goes through `os.Root`" is false for tusd's filestore

- **Location:** spec:120-123, spec:930.
- **Defect:** tusd's filestore writes `tus/` with plain paths: `createFile` calls `os.OpenFile(path,
  os.O_CREATE|os.O_WRONLY|os.O_TRUNC, filePerm)` (`pkg/filestore/filestore.go:323-335`), from `NewUpload`
  (`:109-111`). The promise cannot be tested true. Because of `O_TRUNC`, a second creation for the same file id
  truncates its bytes, which is why "file already started: 409" matters. One state is still unnamed. A creation whose
  response was lost is retried by the client with a POST, and that POST meets 409, so the member is stuck on that
  file until it expires.
- **Fold:** Narrow the promise to rung 7's own writes, and note that tusd's paths are server-minted hex ids, checked
  by the wrapper's shape test and set by the create callback. Name the lost-response state. Recommended: allow a
  re-creation when the file is unreceived and its tus lock can be taken (no PATCH in flight). It restarts that file
  from zero. Otherwise answer 409.

### C16 (minor). Smaller unnamed states

- A ZIP with two members whose names are equal, or equal after `Clean`: `O_EXCL` fails, and no report is named.
  Recommended: `unsafe-path`. AC 16 also omits "a directory entry with content".
- A store error during the verdict (spec:543): the `rejected` write cannot commit either. Name the report: the
  submission stays `checking`, and recovery or the stall event picks it up.
- The sweep's idle clock (spec:712) is unnamed. tus PATCHes write no store row, so `updated_at` does not move during
  a long resumed upload. Name the clock: the wrapper stamps `updated_at` before each PATCH, or the sweep reads the
  `.info` file's mtime. Say that the sweep takes the file's `memorylocker` lock before it deletes.
- "Every open submission's unwritten reserved bytes" (spec:431) has no computation. The store knows only whether a
  file is received, not its tus offset. Name the conservative reading: count each unreceived file in full. Say that
  it double-counts bytes already written, which only makes the check stricter.
- Re-validation at accept (spec:592-595) and `rerouted` have no AC, though the spec says how its test reaches it. Add
  one.
- The privileged-field table (spec:903-916) leaves out `submission_files.role`, `submissions.container`,
  `declared_bytes`, and `submission_files.size`. Recommended: `role` is derived by the server from the extension and
  is never a request field. `container`, `declared_bytes`, and `size` come from the manifest, and `size` is bound by
  `Upload-Length` (C7).
- A batch of more albums than `SESSIONS_PER_MEMBER`: the spec says 3 is "enough for a batch to pipeline" but names
  no page behavior for the fourth album. Recommended: the page holds the extra creates until a slot frees and never
  shows `too-many-sessions` for the member's own batch. Add it to AC 30's states.

### C17 (minor). AC 3 asks for a behavior test of a shell leg

- **Location:** AC 3.
- **Defect:** CLAUDE.md says the scripts carry "no behavior tests of its own". "Proven by editing a fixture copy" is
  either a committed test, which breaks that rule, or a one-time proof, which the AC does not say.
- **Fold:** State it as a one-time mutation proof recorded in the task report.

## Over-ceremony

### O1 (minor). Two untested platform implementations, and a test that is compiled but never run

- **Location:** spec:183-196, AC 32.
- **Cost:** The darwin `_unix.go` device check and free-space read, and the Windows `GetDiskFreeSpaceEx` read, are
  code no test runs at 0.8. The plan still has to research them (darwin's `int32` `Dev`, `uint32` `Bsize`), review
  them, and maintain them. Hook 23 allows "a stated refusal there". AC 32's "unit test of the refusal function built
  under `windows` ... compiled, not run" proves only that it compiles, which the cross-compile already proves.
- **Fold:** Ship `_linux.go` and one `_other.go` (`//go:build !linux`) that refuses both the device check and the
  free-space read with the stated message. Drop AC 32's test claim and keep the cross-compile. The 0.9 home install
  (a Raspberry Pi, which runs Linux) needs nothing more. A darwin or Windows target writes its counterpart when one
  is planned.

### O2 (minor). The racing tests as the proof of atomicity

- **Location:** AC 13, AC 22.
- **Cost:** Under C8's fold, the type-level proof carries the property. Concurrent tests that try to force an
  interleaving (barriers, retries, `-count=N`) add clock time to every gate for little extra assurance. Keep one plain
  concurrent smoke test per property, and add no harness for forcing an interleaving.

## Planning-miss items: verdict

- **Behaviors quoted.** Strong overall: tusd, mewkiz/flac, `archive/zip`, `os.Root`, and the syscalls are quoted with
  file and line. Four gaps: `EvalSymlinks` on an absent path (C12), `stx_mask` (C13), `safeCastInt`'s zero (C5), and
  `Handle.Update` (C8).
- **Every state names its report.** The states table is thorough, but these states have none: stranded receipt (C1),
  in-flight duplicates (C2), stalled recovery and `match-unavailable` idle (C3), cancel during upload and resume
  (C4), malformed position tags (C5), true-up over quota (C6), and the tus-path wire shapes (C10).
- **Privileged fields.** Complete for the fields that matter. C16 adds four rows that need little argument.
- **Fixture state per proof.** Named for most criteria. AC 5, 9, 13, 14, 23, and 25 need the corrections in C9 and
  C14.

## OWNER FORK

None. C11's W8.2 exception for the format reasons is a change to a copy standard. It is proposed as the clearly
correct answer, by analogy with the MusicBrainz exception, and is not raised as a fork. If the conductor judges that
changing W's exception list needs Geoff, the options are:

- (a) add the exception (recommended);
- (b) redraft the two adopted strings without "FLAC", for example "This file isn't in the lossless format the library
  keeps", which tells the member less about how to fix it.
