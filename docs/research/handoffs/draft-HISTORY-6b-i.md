# Draft HISTORY entry: rung 6, pass 6b-i (the ops seam and backup)

Paste-ready for the conductor's ledger fold, once the merge ritual and the
architecture reads land. The bracketed fields need the merge's own commit
ids; everything else is final. Model: the top entry of `docs/HISTORY.md`
(pass 6a) is the format this draft follows.

---

## Rung 6, pass 6b-i (the ops seam and backup): closed 2026-09-25, lane 2

Executed against
`docs/plans/2026-09-24-rung-6b-i-ops-seam-backup.md`
(revision 2.2, approved by Geoff 2026-09-24, all four rulings yes, no
split) in the lane 2 worktree (`.claude/worktrees/lane-2-6b-i`, branch
`lane-2-6b-i`), from `dfe6c4c`. Task 0's staleness read found no
blockers. The plan's full review history, fold rulings, and finding
dispositions live in
`docs/research/2026-09-24-rung-6b-i-plan-fold-record.md`.

Nineteen commits ran across three segments and a simplifier round
(`git log --oneline dfe6c4c..25e5a3e | wc -l` reports 19):

- **Segment A** (Tasks 1S, 1A, 1B, 2, plus a conductor tidy task):
  Task 1S (`4079422`, then fix `418c3d6`) reshaped the store's uid,
  backup-flag, and migrate-source seams ahead of the exports that
  follow. Task 1A (`1d2106a`, `claude-opus-5-5`) exported write mode's
  open, the handle's write method and read path, serve mode's open with
  its nil-logger refusal (Ruling 4's yes), the id function, and the
  lease primitive: `leases/` joins the leaf, an instance lease, a kind
  lease with its busy sentinel, and the kind lease's close-on-exec
  descriptor duplicates for a transfer child. Task 1B (`089afe2`,
  `claude-opus-5-5`) added the backup copy (staged, `VACUUM INTO`,
  verified, checkpointed at `busy_timeout=0`), the stamp format/parse
  pair, and the backup reference that keeps the lost-database guard
  alive; inspect gained the upkeep fields `doctor` reads. A conductor
  tidy task (1T: `5c0a1fc`, then fix `d7e8896`) validated kind-lease
  names and backup-reference remotes, and proved the kind-lease escape
  row actually escapes without the check. Task 2 (`72ec8b7`, then fix
  `367ed53`) added `internal/ops`: the seam interface with the actor as
  a method parameter, the local actor kind (`local:<username>`), five
  of the seven wire sentinels, and the four-operation catalog.
- **Segment B** (Tasks 3-5): Task 3 (`db3189b`, `claude-opus-5-5`) added
  `internal/ops/local`: the runner, recovery (owning conflict policy
  over the store's lease primitive), cancel, 90-day job retention, and
  the job-body contract (`ctx`, a progress callback that is also the
  cancel-poll point). Task 4 (`3bf74fb`, `claude-opus-5-5`, then fix
  `d4b20e9`) added `internal/backup`: the guard fraction (0.5) and
  library-root resolution (`filepath.EvalSymlinks`, once, in the job's
  first phase), the remote validator and its `path.Join` destination
  rule, the transfer child (its own process group, `Cancel`/`WaitDelay`
  set explicitly, the inherited kind-lease and store-lock descriptors),
  and the job that binds `ops/local`'s contract to the backup body. Task
  5 (`6c2d287`) added remote and local pruning after a successful send,
  keeping the run's own snapshot by name and skipping every prune step
  unless the run's stamp is strictly the newest on the remote; a
  conductor tidy task (5T: `d14e43e`) tested four previously-untested
  pruning states and pinned "backup finished" as the job's final
  progress text.
- **Segment C** (Tasks 6-7, plus the simplifier round): Task 6
  (`36d637f`, then two fix rounds `85be4f8` lint findings and `95a5c68`
  a signal-poll spin and untested help-text codes) added the backup
  class, its verb, and the five exit codes (Ruling 3), the verb's own
  signal survival (`context.WithoutCancel`, SIGHUP treated as cancel)
  and its supersession rules. Task 7 landed in two commits (`5600c88`,
  then `67e5e10` exercising `backupDoctorRowsOn` under a flag-on doctor
  run) and added `doctor`'s backup rows: remote validation and the
  store's upkeep facts. The `code-simplifier` round (`25e5a3e`, Opus
  5.5, accepted by review) shared `backup.FinishedMessage`, the store's
  exported copy-name prefixes, and turned `parseSnapshotName` into a
  method on `remoteEntry`.

### The merge

6b-i merges to master at `<merge-sha, filled at the fold>`: `<rebase
outcome>`, a green worktree gate, `git merge --ff-only`, a green gate on
master, `<merged-range diff-reviewer verdict, if the rebase is not a
fast-forward>`.

### The architecture reads

Five `go-architecture-reader` grades run at the merge, each with no plan
in context: `internal/engine/store`, `internal/ops`, `internal/ops/local`,
`internal/backup`, and `cmd/dubplate-server`. `<findings and routing,
filled at the fold>`.

### The gate evidence

`go test -count=1 ./internal/engine/store/...`: **4.976s** (cache
cleared for the whole module first, `go clean -testcache`).
`go test -count=1 ./internal/ops/... ./internal/backup/...`:
**9.105s** wall (`internal/ops` 0.006s, `internal/ops/local` 2.685s,
`internal/backup` 8.298s, the package with the real-rclone-binary
tests). `go test -count=1 ./cmd/dubplate-server/...`: **52.137s**.

`go test ./...` at HEAD (`25e5a3e`), cache cleared first: **2m17.854s**
wall (`internal/harness/conform` at 134.046s dominates, unchanged from
6a; `internal/harness` at 11.598s is the next largest; every touched
package finishes under 9s).

`CAIRN_GATE_LANE=light cairn-run-gate 'bash scripts/check.sh'` at HEAD,
run twice: **exit 0** both times, 102-line log, **1m5.092s** wall on
the second (timed) run. The gitleaks leg scanned about 113.08 MB in
54-54.1s across the two runs and found no leaks. `go vet` and
`golangci-lint run ./...` report 0 issues.

Two staged legs stayed at warning, in the lane-2 tree ahead of the
merge: **326** process citations in Go comments (up from 325 at 6a's
close; `backup.go`'s five and one carried rung-1 citation account for
the growth) and **127** exported-unused `internal/` identifiers (down
from 128; the pass's own exports each gained a production caller: the
verb for `internal/backup` and `internal/ops/local`, `doctor` for the
catalog's read side). Both counts come from `bash scripts/check.sh`'s
own reporting lines, pasted above under "The gate evidence".

### The four rulings and their outcomes

All four ruled **yes** by Geoff on 2026-09-24 (plan `:78-152`):

1. **A 7.4M execution ceiling, planning recorded separately.** Yes: the
   80% flag sits at 5.92M. See "Both budgets" below for the actual.
2. **`doctor` asks rclone itself for backend features**
   (`rclone backend features <remote>: --low-level-retries 1`, a 10 s
   timeout). Yes: landed as R-6bi9 in Task 7 (`5600c88`).
3. **The `backup` verb's five exit codes** (0, 1, 69, 75, 130, the last
   for any means of cancel). Yes: landed as R-6bi10 in Task 6
   (`36d637f`), the help text says "canceled by any means".
4. **Task 1A exports serve mode's open one pass early**, ahead of 6d's
   `serve`. Yes: landed in Task 1A (`1d2106a`) with the nil-logger
   refusal and named as exit gate 11's test-only exception; no
   `storetest` package was needed.

### The resolutions, R-6bi1 through R-6bi18

Each is stated in full in the plan at its own `### R-6biN` heading; this
table names where each landed at HEAD (`25e5a3e`).

| Resolution | What it decides | Landed |
| --- | --- | --- |
| R-6bi1 | Lane 2's package territory against 5c-ii; the overlap is `main.go` and `doctor.go` | Confirmed: Task 0 found no blocking 5c-ii commits since `dfe6c4c`; Tasks 6 and 7 ran last |
| R-6bi2 | The store owns the lease primitive; `ops/local` owns recovery policy | `internal/engine/store` (Task 1A, `1d2106a`); `internal/ops/local/recovery.go` (Task 3, `db3189b`) |
| R-6bi3 | The exported write surface, each with its first caller in this pass | `internal/engine/store` (Tasks 1A `1d2106a`, 1B `089afe2`) |
| R-6bi4 | The actor is a method parameter; 6b-i builds the local kind only | `internal/ops/ops.go` (Task 2, `72ec8b7`) |
| R-6bi5 | Four operations, five of seven sentinels, at 6b-i | `internal/ops` (Task 2, `72ec8b7`) |
| R-6bi6 | The backup copy is staged, verified, sent, and removed; staging cleanup on the kind lease; checkpoint at `busy_timeout=0` | `internal/engine/store` (Task 1B, `089afe2`) |
| R-6bi7 | `--max-delete 0` suffices for the zero-track guard; no fallback abort | `internal/backup` (Task 4, `3bf74fb`) |
| R-6bi8 | Four constants: retention (90d), guard fraction (0.5), snapshot keep (14), archive retention (90d) | `internal/ops/local/jobs.go` (Task 3, `db3189b`); `internal/backup/guard.go`, `prune.go` (Tasks 4 `3bf74fb`, 5 `6c2d287`); the stamp layout in `internal/engine/store` (Task 1B, `089afe2`) |
| R-6bi9 | The backend features check reads `rclone backend features`, not a copied table | `cmd/dubplate-server/doctor.go` (Task 7, `5600c88`) |
| R-6bi10 | Exit codes and `main`'s exit path | `cmd/dubplate-server` (Task 6, `36d637f`) |
| R-6bi11 | The verb survives its own signal; SIGHUP treated as cancel | `cmd/dubplate-server/backup.go` (Task 6, `36d637f`, fix rounds `85be4f8`, `95a5c68`) |
| R-6bi12 | The transfer child's process group and inherited kind-lease/store-lock descriptors | descriptor duplication in `internal/engine/store` (Task 1A, `1d2106a`); process group and `Cancel`/`WaitDelay` in `internal/backup` (Task 4, `3bf74fb`) |
| R-6bi13 | `doctor` reads backup facts through the inspect report; opens nothing for writing | inspect fields in `internal/engine/store` (Task 1B, `089afe2`); `doctor`'s consumption (Task 7, `5600c88`) |
| R-6bi14 | One remote validator; destinations joined with `path.Join`, never string concatenation | `internal/backup` validator (Task 4, `3bf74fb`); `doctor`'s call (Task 7, `5600c88`) |
| R-6bi15 | Pruning never trusts the clock alone; skip is recorded, not failed | `internal/backup/prune.go` (Task 5, `6c2d287`; tidy 5T, `d14e43e`) |
| R-6bi16 | The backup reference keeps the lost-database guard alive after Ruling 7's deletion | `internal/engine/store` (Task 1B, `089afe2`) |
| R-6bi17 | The job body contract (Task 3, no `backup` in view) and its binding (Task 4) | `internal/ops/local` (Task 3, `db3189b`); `internal/backup` (Task 4, `3bf74fb`) |
| R-6bi18 | The library root resolves once with `filepath.EvalSymlinks`, in the job's first phase | `internal/backup/guard.go` (Task 4, `3bf74fb`) |

### The declared coverage gaps

**Five**, all named in the plan's "Not verified in this revision"
(`:2234-2251`), none blocking the exit gate:

1. Whether every rclone backend accepts `--metadata-set`; only the
   `local` backend on tmpfs was probed.
2. Whether `rclone backend features` contacts the provider for every
   backend; Ruling 2's 10 s bound and `--low-level-retries 1` cover the
   case where it does.
3. `--max-delete`'s counting of `--backup-dir` moves was measured on
   the `local` backend only; a copy-only backend (S3, B2) archives by
   server-side copy and delete, untested.
4. The box's rclone version: nothing in 6b-i reads it; the runbook pins
   it at or above v1.74.3 (see "The runbook lines" below).
5. A forward clock jump still purges archive directories early
   (R-6bi15); the safety rule catches only a backward jump.

Five further gaps surfaced in review and are carried, not fixed, since
each is a declared limit rather than a defect:

6. The prune step-failure test proves the failure path takes effect but
   not the "and it is logged" half, since the test wires a discard
   logger.
7. A fixture that chmods a directory 0555 to force a permission-denied
   path fails under a root-run test process, since root bypasses the
   mode bit.
8. A `backup` run that receives SIGINT or SIGTERM and then SIGHUP calls
   the cancel operation twice; the second call is a no-op the runner
   tolerates.
9. `doctor`'s features probe inherits the process's own environment
   rather than the injected one a test supplies; a test that needs a
   different `RCLONE_CONFIG_*` environment for the probe alone cannot
   express it.
10. `PreMigrateSuffix` doubles as the pre-restore suffix; the sharing is
    a naming choice, not a semantic one, and no test conflates the two
    uses.

### What a later pass would be wrong to rediscover

- **R-6bi7's measurement**: rclone v1.74.3 counts a move into
  `--backup-dir` against `--max-delete` as a delete (exit 7, "failed to
  delete 1 files", nothing archived), and a modified file is archived
  under `--max-delete 0` regardless (`Server Side Moves: 1`). The
  fallback abort the spec left open is not needed.
- **R-6bi12's orphan shape**: an orphaned transfer child (parent killed
  outright) keeps its inherited kind-lease and store-lock descriptors
  held until it exits, which for a `stderr`-piped rclone process is
  bounded by the next write to the closed pipe raising SIGPIPE, at the
  latest the sync's own end. `Cmd.Cancel`'s precedence governs every
  outcome: a nonzero exit or any signal death after the job's context is
  done reads canceled; a signal nobody sent through `Cancel` reads
  failed with `rclone-signal-<NAME>`; any other nonzero exit reads
  failed with `rclone-exit-<N>`.
- **The checkpoint's `busy_timeout=0`** (R-6bi6): SQLite's `wal_checkpoint(TRUNCATE)`
  blocks new writers while pending; a probe measured a 1.93s stall at a
  2s busy timeout against a held reader, and no stall at 0. Every
  backup's checkpoint runs on a connection whose busy timeout is set to
  0 for that call alone, so serve's writers are never held behind a
  backup.
- **R-6bi14's remote join**: destinations are built only by
  `path.Join(path, sub)` on the value split at its first colon, never by
  string concatenation. `name:` and `name:dir` and `name:dir/` all give
  a path under the remote's root; `name:/abs` keeps its absolute path.
  The prior string-concatenation rule silently escaped to the backend's
  root on a bare `name:`.
- **The `lsjson` exit 3** (R-6bi15): rclone v1.74.3's `lsjson` on a
  missing directory exits 3, "directory not found"; a listing of an
  absent `snapshots/` or `archive/` reads as empty rather than an error,
  which the first-ever prune (no `archive/` yet) and a fresh remote both
  hit.

### Both budgets

Tokens: **about 5.4M in subagent tokens** before Task 8, the merge, and
the architecture reads (Task 0 about 0.08M, segment A about 2.28M,
segment B about 1.35M, segment C about 1.49M, the simplifier round with
its review about 0.16M), against the **7.4M execution ceiling** with
the 80% flag at **5.92M** (Ruling 1). Geoff approved continuing past
the flag on 2026-09-25. Attended time: **zero attended interactions
after approval**, one combined approval at plan close. Fix rounds: Task
1S one, Task 1T one, Task 2 one, Task 4 one, Task 6 two, Tasks 1A, 1B,
3, 5, 7 none. `<Task 8's own spend, the merge ritual's, and the
architecture reads', filled at the fold>`.

---

## The errata this pass owes

Drafted paste-ready for the conductor from the plan fold record's "Owed
errata" table
(`docs/research/2026-09-24-rung-6b-i-plan-fold-record.md:166-213`),
numbered 6bi-E1 to 6bi-E13. All four rulings are yes, so no erratum text
changes on a "no" branch.

- **6bi-E1 (R-6bi2).** Package territory (`:200`, `:202`): the lease
  primitive lives in `internal/engine/store`; recovery policy lives in
  `ops/local`.
- **6bi-E2 (R-6bi5, R-6bi4).** Vocabulary and privileges (`:538-546`):
  6b-i declares five sentinels and the local actor kind; `forbidden`,
  `internal`, and the network actor land in 6b-ii with their first
  caller.
- **6bi-E3 (R-6bi6).** The backup job (`:624-626`): the backup copy is
  staged, verified, sent, and removed; the rename and hand fsync hold
  for the pre-migration copy only.
- **6bi-E4 (R-6bi6).** Layout and erratum E7: a backup start holding the
  kind lease removes every `staging/backup-*.db`, the one write-mode
  removal of staging; 6a handoff item 3's per-instance recovery cleanup
  is superseded.
- **6bi-E5 (R-6bi6).** The backup copy's checkpoint runs with
  `busy_timeout=0` on its own connection, so it never stalls serve's
  writers.
- **6bi-E6 (R-6bi7).** Guards (`:647-649`) and criterion 13:
  `--max-delete 0` suffices; the fallback abort is not built.
- **6bi-E7 (R-6bi8).** Constants: `DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP`
  is declared by 6d; the stamp layout is `20060102T150405.000000Z`.
- **6bi-E8 (R-6bi9).** Credentials and `doctor`: the per-backend feature
  check reads `rclone backend features --low-level-retries 1` (Ruling
  2's yes).
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
  prune step unless the run's stamp is strictly the newest on the
  remote.
- **6bi-E13 (R-6bi16).** Layout, erratum E11, and ruling 7: a successful
  backup leaves `snapshots/backup-<stamp>.ref`, which the lost-database
  guard reads as a copy and whose refusal names `restore --snapshot`.

The spec's "Not verified" rows this pass settles: `--max-delete`
counting of `--backup-dir` moves (on `local`), the feature table,
`os/exec`'s `WaitDelay` and `Cancel` wording, rclone's exit codes (7 on
`--max-delete`, 130 on SIGINT mid-sync, 3 on a missing directory
listing).

---

## The 6b-i to 6b-ii handoff

1. **The ops interface and codes 6b-ii's projection mirrors**:
   `internal/ops`'s `Seam` interface, its four operations
   (`OpStartBackup` "start_backup", `OpJobStatus` "job_status",
   `OpCancelJob` "cancel_job", `OpCatalog` "catalog"), and the five
   sentinels 6b-i declares (`unavailable`, `not_found`, `conflict`,
   `invalid_request`, `canceled`), each with its wire code.
2. **`forbidden`, `internal`, and the network actor**, which 6b-ii
   declares with their first caller: the network actor is built from
   `identity.Identity` (`internal/identity`'s import lands with 6b-ii,
   not 6b-i, R-6bi4).
3. **The catalog operation's first production caller**: 6b-i exports it
   for the catalog vocabulary (spec criterion 8); 6b-ii's projection is
   its first production caller.
4. **The lease and kind-lease semantics restore must respect**: a clean
   close removes its own instance lease; a kind lease's busy error is
   the sentinel `ops/local` matches with `errors.Is`; recovery tries
   only the owners its own records name, never a blind sweep of
   `leases/`.
5. **The staging rule**: a backup start, once it holds the kind lease,
   removes every `staging/backup-*.db` before staging its own; a
   restore's own staging rule is unbuilt at 6b-i and should not assume
   the backup rule covers it.
6. **The remote layout and join restore reads**: destinations are
   `<name>:<path.Join(path, sub)>` for `sub` in `library`,
   `archive/<stamp>`, `snapshots/<stamp>.db`; a restore reading
   `snapshots/<stamp>.db` back must build the same join, never
   string-concatenate.
7. **The backup reference**, `snapshots/backup-<stamp>.ref`: a restore
   from a snapshot should replace it (the reference names the newest
   successful backup, and a restore establishes a different newest
   state).

---

## Forward items for later rungs

- **To 6d**: a store identity minted at create, so a backup can refuse
  a remote another store writes to (risk R9's option b).
- **To rung 7**: per-collection and per-extension counting in the
  guard; a `--max-delete` sized as `ceil((1 - fraction) × tracks)` with
  an operator flag for a planned mass change; a regression floor on
  remote pruning (risk R10's option b, risk R1's option b).
- **To 6c**: the `SAVEPOINT` check on migration files that also catches
  `COMMIT; BEGIN;` (risk R13; carried from 6a's disposed review items,
  unchanged since 6b-i adds no migration).

---

## The CLAUDE.md Gate line for `rclone`

Draft text only; the conductor edits `CLAUDE.md`'s Gate section, never
this file.

> `rclone` config validation joins `scripts/check.sh` at rung 6b-i's
> merge: the backup remote's name and path part are checked by the same
> validator `doctor` and the `backup` job call
> (`internal/backup`'s remote validator), never a second copy.

## The runbook lines owed to rung 16

Draft text only; the conductor folds these into rung 16's runbook when
it is written.

- `KillMode=mixed` on the backup-capable unit, so SIGTERM reaches the
  verb first and SIGKILL only the remaining process group
  (`systemd.kill(5)`), letting the verb cancel through the job record
  before rclone dies.
- rclone pinned at or above the measured v1.74.3 on the box; nothing in
  6b-i reads the installed version.
- rclone's remote variables (`RCLONE_CONFIG_<NAME>_TYPE` and friends)
  are set by hand until the config-template rows land.
- A crypt remote is recommended.
- One remote path per server.
- The remote value's path part names where the backup lives, bucket
  included (a bucket backend's first path element is the bucket).
- The non-empty wrong-mount gap: a library root that resolves to the
  wrong (but non-empty) mount is not caught by the guard fraction alone.
- With zero tracks, any library deletion fails every run with exit 7
  until tracks exist again; remedied by archiving the file on the
  remote by hand with `rclone moveto`.
- Removing a backup reference by hand
  (`snapshots/backup-<stamp>.ref`) is how an operator deliberately
  starts a fresh store.
- A restore before rung 12 loses stars and playlists.
