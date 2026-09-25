# Rung 6b-i plan fold record

Target: `docs/superpowers/plans/2026-09-24-rung-6b-i-ops-seam-backup.md`,
revision 1 (`b36871c`) to revision 2 (folded, awaiting verification).
Inputs: the three lens reviews at
`docs/superpowers/research/2026-09-24-rung-6b-i-plan-review-{contract,mechanics,risk}.md`.
Code read at `lane-2-6a` (`aa3b13e`; the branch head `3b962aa` adds only
the 6a close drafts).

## Counts

| Disposition | Count |
| --- | --- |
| Folded (including convergent duplicates, each listed once per ID) | 55 |
| Folded in part, with the rest refused | 3 (M10, R2, R8) |
| Refused outright | 0 |
| Owner fork, consolidated into a ruling | 2 (C1 as Ruling 4, F1 into Ruling 3) |

Sixty findings: C1 to C29, M1 to M14, F1, and R1 to R16. The reviewers'
candidate forks C11, R1, R9, and R10 each had a clearly safe answer inside
6b-i's scope and were folded rather than asked (reasons below).

## What the fold verified before acting

Each load-bearing claim was checked against source or a probe under
`/tmp/claude-1000/fold6bi/`, never inside the repo.

- **C1.** `internal/engine/store` exports only `Inspect`, its options, and
  its report at `aa3b13e`; `openServe` and `serveOptions` are unexported
  (`open.go`). The reviewer's recommended option (a), a `storetest`
  package, cannot reach the unexported create or migration runner from a
  separate package, so it needs the same export as option (b). Ruling 4
  re-casts the fork honestly.
- **C2.** Spec `:578-580`: "pruning never removes the newest successful
  record of a kind." The 10-day record is the newest, so the 100-day
  succeeded record is prunable. Confirmed.
- **C4, M7.** `go doc os/exec.Cmd` at go1.27.1 carries the quoted `Cancel`
  paragraph. Confirmed.
- **C18.** `leaf.create` chmods every leaf directory back to 0700
  (`leaf.go:75-81`). Confirmed.
- **C29.** `busyError`'s doc names the btrfs gap (`lock.go:103`). Confirmed.
- **C9.** `0002_jobs.sql` has nullable `finished_at`, `started_at`, and
  `progress`. Confirmed.
- **M2.** SQLite documents that a FULL or stronger checkpoint blocks new
  writers while it waits; the mechanics probe measured the stall. The
  fold's reading of `snapshot.go` shows 6a's checkpoint runs with the
  caller's busy timeout. Confirmed.
- **R1.** `newestCopy` (`open.go`) counts only `pre-migrate-v<N>.db` and
  `pre-restore-*.db`; Task 5 deletes both after each backup. Confirmed.
- **R5.** `rclone lsjson` on a missing `local` directory exits 3 with
  `directory not found` (rclone v1.74.3, fold probe). Confirmed.
- **R7.** `path/filepath.WalkDir` over a symlink to a two-file directory
  counted 0 files (fold probe; `go doc` says "WalkDir does not follow
  symbolic links"). Confirmed. Note that `io/fs.WalkDir` does follow a
  symlinked root; the plan names the `path/filepath` function.
- **R3, M5.** `rclone copyto x loc:snapshots/x.db` on a `local` remote
  wrote under the working directory (fold probe). Confirmed.
- **R9.** No store identity exists: `0001_ledger.sql` sets only
  `application_id`, and no migration carries a store id.
- **C12.** Go opens files close-on-exec (`os/file_unix.go`,
  `flag|syscall.O_CLOEXEC`). Confirmed.
- **C27.** `rclone backend features` on an `http` remote at
  `http://127.0.0.1:9/` answered both false, exit 0, in 0.27 s with
  nothing listening (fold probe).
- **M10's sweep.** Not taken: see the refusal below.

## Dispositions

### Contract lens

| ID | Disposition | Location in revision 2 |
| --- | --- | --- |
| C1 | **Owner fork, Ruling 4.** The recommended `storetest` needs the same export as exporting serve's open; the ruling asks whether to export it now for tests. | Rulings 4; R-6bi3; Global constraints; exit gate 11 |
| C2 | Folded. Criterion 17 expects both 100-day records deleted; mutation 6 added; an interrupted-record row added. | Task 3 criterion 17, mutations 6 and 11 |
| C3 | Folded with M3. Task 3 defines the body contract; Task 4 binds it in `ops/local`; the import edge stays as the spec draws it. | R-6bi17; Task 3 outcomes; Task 4 Files and outcomes |
| C4 | Folded with M7. `Cancel`'s `Wait` paragraph quoted; precedence ruled (context first, then signal death, then exit code). | R-6bi12; Tasks 4 and 7 tables |
| C5 | Folded with M1. Criterion 9's store half split into fixtures A and B. | Task 1B criteria |
| C6 | Folded with M6 and R14. Two stand-in modes, a pid file, the verb child started with `Setpgid`, an alive check at 500 ms. | Task 4 stand-in; Task 6 criterion 20 |
| C7 | Folded with R2. Own snapshot kept by name, 13 newest of the rest, and a skip unless the run's stamp is strictly newest. | R-6bi15; Task 5 |
| C8 | Folded. The failed-run row runs the real binary in criterion 13's shape against 15 seeded snapshots. | Task 5 failed-run row |
| C9 | Folded. Rows for `finished_at`, `started_at`, `progress`, and `updated_at`; every terminal transition sets `finished_at`. | Privileged fields; Task 3 |
| C10 | Folded, with R4's first bullet. Held-owner fixture for the queued cancel; a refused queued-to-running ends the goroutine before the body. | Task 3 criterion 6, cancel-first row, mutation 10 |
| C11 | Folded with M12, not forked. `forbidden`, `internal`, and the network actor (with the `internal/identity` import) defer to 6b-ii, since no 6b-i criterion needs them. The catalog operation stays: criterion 8 requires it, so exit gate 11 names it as an exception. | R-6bi4; R-6bi5; Task 2; exit gate 11; spec deviations |
| C12 | Folded with M4 and R16. The busy sentinel and a `dup(2)` descriptor export in R-6bi3; `ExtraFiles` and close-on-exec quoted. | R-6bi3; R-6bi12; Task 1A |
| C13 | Folded. State rows for WAL size, free space, and local copies; none fails `doctor`. | R-6bi13; Task 7 table |
| C14 | Folded. A succeeded record's text lives in `progress`. | Privileged fields; Task 5 |
| C15 | Folded with M10 and R12. Release removes the lease file while holding it. | R-6bi2; Task 1A |
| C16 | Folded. The fixture uses a stamp in the layout. | Task 5 criterion 16 |
| C17 | Folded. The terminal record names the same dead owner. | Task 3 criterion 4 |
| C18 | Folded. The rename-failure fixture enters through the unexported runner after create. | Task 1B criteria |
| C19 | Folded. Stamp layout at microsecond resolution. | R-6bi8 |
| C20 | Folded. SIGTERM and SIGHUP rows. | Task 6 criterion 20 |
| C21 | Folded. Metadata read-back in criterion 10; the backend claim moved to "Not verified". | Task 4 criterion 10; Not verified |
| C22 | Folded with M14's first part. Crypt `_REMOTE` unset, plain path, connection string, and empty name rows. | Task 7 |
| C23 | Folded with R7. `rclone-missing` belongs to the body's first phase; the library root has a named channel. | Task 4; R-6bi18; privileged fields |
| C24 | Folded. Deliverables recounted (31). | Token ceiling |
| C25 | Folded. Pointer base moved to `aa3b13e`; Task 0 item 2 pre-maps the renames. | Header; Task 0 |
| C26 | Folded into Ruling 3 with F1. | Ruling 3 |
| C27 | Folded. Loopback URL named and probed. | Task 7 criteria |
| C28 | Folded. Three states driven; verification failure declared reviewer-read. | Task 1B criteria |
| C29 | Folded with M8. No assertion depends on a pid. | R-6bi2; Task 1A criteria |

### Mechanics lens

| ID | Disposition | Location in revision 2 |
| --- | --- | --- |
| M1 | Folded as C5. | Task 1B criteria |
| M2 | Folded. `busy_timeout=0` on the checkpoint's own connection; a writer-latency row and mutation. No spec deviation. | R-6bi6; Task 1B |
| M3 | Folded as C3; the orphan row lives in `backup_test`. | R-6bi17; Task 4 |
| M4 | Folded as C12. | R-6bi3; Task 1A |
| M5 | Folded with R3. The join rule and argv rows; an empty path is accepted. | R-6bi14; Task 4 |
| M6 | Folded as C6. | Task 6 criterion 20 |
| M7 | Folded as C4, plus `rclone-signal-<NAME>`; rclone's exit 130 moved from "Not verified" to measured. | R-6bi12; Task 4 |
| M8 | Folded. The no-record conflict reads fixed text without pids. | R-6bi2; Task 3 |
| M9 | Folded. A not-exist removal after the lock reads free. | R-6bi2; Task 1A |
| M10 | Folded in part: a clean close removes its own lease. **Refused in part (the recovery sweep):** a sweep of every `leases/` file races a lease created but not yet locked and would remove a live process's lease, and the residual it cures is one small file per crash with no open record. | R-6bi2 |
| M11 | Folded. On a busy kind lease, recovery runs before the conflict answer. | Task 3 |
| M12 | Folded as C11. | R-6bi5; Task 2 |
| M13 | Folded into Ruling 2: `--low-level-retries 1`, with the B2 timing as evidence. | Ruling 2; R-6bi9 |
| M14 | Folded. Crypt plain path and connection string (Task 7); the same-instance stage by R11's kind-lease sweep; the orphan's SIGPIPE-bounded lifetime recorded. | Task 7; R-6bi6; R-6bi12; Task 8 |
| F1 | **Owner fork, into Ruling 3.** Recommendation (a): 130 for a cancel by any means, stated in the ruling and the help. | Ruling 3 |

### Risk lens

| ID | Disposition | Location in revision 2 |
| --- | --- | --- |
| R1 | Folded, not forked. Option (a), the backup reference E11 reads, is the clearly safe fold: it keeps ruling 7 and criterion 16 as ruled and restores the guard's reach. It is a spec deviation approved with the plan. Option (b), the regression floor, routes to rung 7. | R-6bi16; Tasks 1B and 5; Task 8 forward items; spec deviations |
| R2 | Folded in part, as C7: the listing and strictly-newest checks, archive age from the run's stamp, a future-stamp row, and the forward-jump residual recorded. **Refused in part:** refusing the run on a behind clock (`clock-behind`) would lose a backup to protect a prune the skip already protects, and the stage-size comparison adds nothing that rclone's own post-transfer check and the verified stage do not already give. | R-6bi15; Task 5; Not verified |
| R3 | Folded with M5. One validator shared by `doctor` and the job (`remote-invalid`); no-colon refused; `purge` and `deletefile` arguments built only from a parsed stamp. | R-6bi14; Tasks 4, 5, 7 |
| R4 | Folded. The verb exits after the runner's close, a failed cancel write falls back to an in-process cancel, and a refused queued-to-running never runs the body. | R-6bi11; Tasks 3 and 6 |
| R5 | Folded. Exit 3 on a listing reads empty; a first-run row; a prune skip or failure shows in `doctor`'s last-success row. Ruling 3's codes unchanged. | R-6bi15; Tasks 5 and 7 |
| R6 | Folded. Criterion 3 carries a dead-owner positive control and mutation 9. | Task 3 |
| R7 | Folded. The root's channel named; `filepath.EvalSymlinks` once; symlink and dangling-symlink rows; `WalkDir` quoted. | R-6bi18; Task 4 |
| R8 | Folded in part: "Not verified" rows for the `local`-only measurement and the box's version, and a rung 16 runbook pin. **Refused in part (a `doctor` rclone version row):** rung 16's provisioning pins the box's rclone, so a version row adds a subprocess and goldens for a fact the provisioning already fixes. | R-6bi7; Not verified; Task 8 |
| R9 | Folded, not forked. Option (b) needs a store identity, which no migration carries and 6b-i adds no migration, so it routes to 6d, which owns create; 6b-i records a runbook line (option a). | Task 8 forward items and runbook |
| R10 | Folded, not forked. The reviewer's recommendation (c): the plan stands at 6b-i, where every box has zero tracks, and option (b) routes to rung 7, the first writer of tracks. | Task 8 forward items |
| R11 | Folded. A backup start holding the kind lease sweeps every `staging/backup-*.db`; recovery no longer touches staging. | R-6bi6; Tasks 1B and 4; spec deviations |
| R12 | Folded as C15 and M9. | R-6bi2; Task 1A |
| R13 | Folded as a record update: the decline names the `COMMIT; BEGIN;` case, and the `SAVEPOINT` check routes to 6c. | Carried items; Task 8 forward items |
| R14 | Folded as C6. | Task 6 criterion 20 |
| R15 | Folded. Runbook line with the `rclone moveto` remedy; no flag. | Task 8 |
| R16 | Folded as C12, and the staging deviation added in its kind-lease form (R11). | R-6bi3; spec deviations |

## The rulings after the fold

1. **Ceiling.** 8.2M execution, flag 6.56M, projection 6.20M (eight
   implementation tasks after Task 1's split). Recommendation: yes.
2. **Features probe.** `rclone backend features --low-level-retries 1`
   under 10 s. Recommendation: yes.
3. **Exit codes.** 0, 1, 69, 75, and 130 for a cancel by any means.
   Recommendation: yes.
4. **Serve's open exported now for tests.** Recommendation: yes; no
   builds a checked-in leaf template.

## Scope change and the ceiling

Task 1 passed four deliverables once the descriptor export, the lease
cleanup, the checkpoint pragma, the backup reference, and serve's open
joined it, so it split into Tasks 1A and 1B (three deliverables each).
Tasks 3, 4, 5, and 7 grew rows but not deliverables, except Task 5, which
gained the reference (three to four). Surfaces left: two sentinels, the
network actor, the `internal/identity` import, and recovery's staging
cleanup. Deliverables: 25 claimed in revision 1 (actually 28) to 31. The
ceiling re-derivation is in the plan's "Token ceiling".

## Owed errata to the rung 6 spec

Numbered here 6bi-E1 to 6bi-E13 to keep clear of 6a's E1 to E11; the
conductor assigns final numbers. Each takes effect with the plan's
approval, and Task 8 drafts the paste-ready text.

- **6bi-E1 (R-6bi2).** Package territory (`:200`, `:202`): the lease
  primitive lives in `internal/engine/store`; recovery policy in
  `ops/local`.
- **6bi-E2 (R-6bi5, R-6bi4).** Vocabulary and privileges (`:538-546`):
  6b-i declares five sentinels and the local actor; `forbidden`,
  `internal`, and the network actor land in 6b-ii with their first caller.
- **6bi-E3 (R-6bi6).** The backup job (`:624-626`): the backup copy is
  staged, verified, sent, and removed; the rename and hand fsync hold for
  the pre-migration copy only.
- **6bi-E4 (R-6bi6).** Layout and erratum E7: a backup start holding the
  kind lease removes every `staging/backup-*.db`, the one write-mode
  removal of staging; 6a handoff item 3's per-instance recovery cleanup is
  superseded.
- **6bi-E5 (R-6bi6).** The backup copy's checkpoint runs with
  `busy_timeout=0` on its own connection, so it never stalls serve's
  writers.
- **6bi-E6 (R-6bi7).** Guards (`:647-649`) and criterion 13: `--max-delete
  0` suffices; the fallback abort is not built.
- **6bi-E7 (R-6bi8).** Constants: `DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP` is
  declared by 6d; the stamp layout is `20060102T150405.000000Z`.
- **6bi-E8 (R-6bi9).** Credentials and `doctor`: the per-backend feature
  check reads `rclone backend features --low-level-retries 1` (Ruling 2's
  yes).
- **6bi-E9 (R-6bi12).** The job model: the transfer child inherits the
  kind lease and the store's shared lock by duplicated descriptors and
  runs in its own process group; a canceled job's outcome is decided by
  the context before the exit status.
- **6bi-E10 (R-6bi2).** The job model: a clean close removes its own
  instance lease file.
- **6bi-E11 (R-6bi14).** Remote layout: destinations join the remote's
  path with `path.Join`, never by string; the value must carry a colon.
- **6bi-E12 (R-6bi15).** Ruling 10 and criterion 15: the prune keeps the
  run's own snapshot by name and the 13 newest others, and skips every
  prune step unless the run's stamp is strictly the newest on the remote.
- **6bi-E13 (R-6bi16).** Layout, erratum E11, and ruling 7: a successful
  backup leaves `snapshots/backup-<stamp>.ref`, which the lost-database
  guard reads as a copy and whose refusal names `restore --snapshot`.

The spec's "Not verified" rows this pass settles: `--max-delete` counting
of `--backup-dir` moves (on `local`), the feature table, `os/exec`'s
`WaitDelay` and `Cancel` wording, rclone's exit codes (7 on
`--max-delete`, 130 on SIGINT mid-sync, 3 on a missing directory listing).

## Second fold (revision 2.1)

Inputs: the fresh-context verification read
(`docs/superpowers/research/2026-09-24-rung-6b-i-plan-fold-verification.md`,
V1 to V10) and the `internal/engine/store` section of the 6a architecture
read (`docs/superpowers/research/2026-09-24-rung-6a-architecture-reads.md`),
taking only the items that shape a signature 6b-i exports or binds to.
Base: master `0b6da0f` (6a merged at `8736528`; the store tree equals
revision 2's `aa3b13e`).

Verified before acting: `syscall.Dup` returns a descriptor with
`FD_CLOEXEC` clear and `fcntl(F_DUPFD_CLOEXEC)` one with it set (fold probe
on go1.27.1, `/tmp/claude-1000/fold6bi/dup/`); `syscall/exec_linux.go`
pass 2 creates each `ExtraFiles` slot without close-on-exec; the `euid`
parameters, `migrateOptions.backup` with `settle` omitting it
(`open.go:261`) and the test helper defaulting it to true
(`migrate_test.go:63-65`), and `migrate`'s `fs.FS` parameter all read as
the architecture read states at `0b6da0f`.

### Verification findings

| ID | Disposition | Location in revision 2.1 |
| --- | --- | --- |
| V1 | Folded. The stand-in appends one line per SIGINT received; the group row asserts exactly one (the runner's `Cancel`), so dropping `Setpgid` fails it by count with no timing window. Task 4 mutation 5 and a new Task 6 mutation 7 name the row. | Task 4 stand-in and mutation 5; Task 6 criterion 20 and mutation 7; review focus 5 |
| V2 | Folded. Duplicates use `fcntl(F_DUPFD_CLOEXEC)`, with the probe and the `exec_linux.go` pass 2 quote; a Task 1A row reads `FD_CLOEXEC` and mutation 7 tries plain `dup(2)`. | R-6bi12; R-6bi3; Task 1A; privileged fields |
| V3 | Folded. One ownership sentence: lease values from the primitive, none taken by the handle, the runner the instance lease's one owner, the descriptor method on the kind-lease value. A Task 1A row proves a handle's open and close create no lease file (mutation 8). | R-6bi2; R-6bi3; Task 1A; privileged fields |
| V4 | Folded. The store is the stamp layout's one home and exports a format and parse pair (Task 1B, `stamp.go`), with UTC conversion and the malformed-form rows. | R-6bi8; R-6bi3; Task 1B |
| V5 | Folded. Ruling 1 now calls 2.1M a projection and states the margin as a number (0.43M, about one task) against the re-derived ceiling. | Ruling 1 |
| V6 | Folded. Ruling 4's "no" has the conductor amend Task 1A's Files before dispatch and name the helper's home; the chmod step is stated as required. | Ruling 4 |
| V7 | Folded. Pointer base `0b6da0f`; Task 0's diff base and rename commit (`f85e7be`) repointed. | Header; Primary inputs 3; Task 0 items 1 and 2 |
| V8 | Folded. The rulings preamble says approval also approves the deviations, and that R-6bi15 and R-6bi16 qualify rulings 10 and 7. | Rulings preamble |
| V9 | Folded (cheap). Precedence rule 1 applies to a nonzero exit or a signal death; an exit 0 continues and the next progress callback observes the cancel. | R-6bi12; Task 4 table |
| V10 | Folded (cheap). The orphan exits at its next stderr write, which at the default log level may be the sync's end; the inherited locks, not a short window, keep it safe. | R-6bi12 |

### Architecture read items taken

| Item | Disposition | Location in revision 2.1 |
| --- | --- | --- |
| 1 (`euid` seam) | Taken. `var geteuid = os.Geteuid`; `inspect` folds into `Inspect`; `openWrite` and `openServe` lose the parameter before Task 1A exports them. Mutation 1. | New Task 1S |
| 2 (`migrateOptions.backup`) | Taken, wired. Production is already the zero value (backup off, prune), but `settle` never passed the field; serve's options carry it through so 6d wires `DUBPLATE_BACKUP_ENABLED` without reshaping the type Ruling 4 may export. The test helper's default flips to the zero value. Mutation 2 and a backup-on serve row. | New Task 1S |
| 3 (`migrate` takes `[]migration`) | Taken. Mutation 3, declared reviewer-read (a double parse only). | New Task 1S |
| 4 (comment corrections) | Taken for everything 6b-i exports: each export's and each returned sentinel's doc names the `errors.Is` callers that now exist. | R-6bi3; Task 1A |
| 7 (one `waitLock`) | Not taken in 6b-i. Its condition fails: 6b-i's leases use the non-blocking `tryLock`, not the poll loop, and write mode's bounded shared wait keeps its signature. Routed to the architecture cleanup pass. | none |
| 8 (`primaryCode` with `sqlite3.SQLITE_*`) | Taken as a conditional rule: new code matches no raw SQLite code and reuses the store's classifiers; a task that must match one lands the helper first. Otherwise routed to the architecture cleanup pass. | Global constraints |
| 5, 6, 9, 10, 11, 12 | Left to the architecture cleanup pass, as directed: `State.String` composition, the duplicate rollback, the DSN base list, the shared preamble, the file split, and the comment cleanup outside 6b-i's exports. | none |

### Scope and the ceiling

Task 1S adds two deliverables (33 in all) and about 0.25M, which re-derives
the projection to 6.45M. The ceiling rises to 8.6M (flag 6.88M, 0.43M above
the projection); the Task 3 split would bring the projection to 6.85M, just
under the flag. Segment A grew to five tasks, so the pass now runs in four
segments of at most four (A: 1S, 1A, 1B, 2; B: 3, 4, 5; C: 6, 7; D: 8), and
segment B's end is the pre-named pass split point.

### Rulings after the second fold

1. Ceiling 8.6M execution, flag 6.88M. Recommendation: yes. No holds 5.9M
   as a total, and the flag fires on plan.
2. The features probe (`rclone backend features --low-level-retries 1`,
   10 s). Recommendation: yes. No builds a static feature table.
3. Exit codes 0, 1, 69, 75, and 130 for a cancel by any means.
   Recommendation: yes. No uses 2, 3, and 4.
4. Serve's open exported now for tests. Recommendation: yes. No builds a
   checked-in leaf template, with the conductor amending Task 1A's Files.

### Errata additions

- **6bi-E9, amended (V2).** The transfer child's inherited descriptors are
  close-on-exec duplicates (`F_DUPFD_CLOEXEC`) handed out by the kind-lease
  value.
- **6bi-E14 (V4).** The store owns the backup stamp layout
  `20060102T150405.000000Z` and exports its format and parse pair.
- **6bi-E15 (V9).** A transfer that exits 0 after a cancel request counts
  as a success; the job observes the cancel at its next progress point.
