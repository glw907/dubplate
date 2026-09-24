# Rung 6a plan fold record

This record folds three lens reviews of the rung 6a plan
(`docs/superpowers/plans/2026-09-23-rung-6a-store-foundation.md`, revision
1 at `11f11c2`) into revision 2 of the same file. It gives one
disposition per finding, the owner forks as numbered rulings, the owed
errata against the ratified rung 6 spec (revision 2.2), and the ceiling
change. The reviews are the contract lens (C1 to C13), the mechanics lens
(M1 to M7 and m1 to m9, which the dispatch counted as M1 to M16), and the
risk lens (D1 to D17).

## Summary

- **Findings:** 46. **Folded:** 38. **Folded in part, the rest refused:**
  4. **Refused:** 1. **Owner fork:** 3 findings, which became 2 rulings.
- **Rulings for Geoff:** 4 (the drafter's 2, plus D5 and C7/M5).
- **Deliverable counts:** unchanged. Task 6's four are re-cut. No task
  passes four.
- **Ceiling:** 6.4M becomes 6.8M execution; the projection moves from
  4.62M to 4.92M; the flag moves from 5.12M to 5.44M.
- **Owed errata:** 11 (the drafter's 6, two of them broadened, and 5
  new, two of them conditional on a ruling).

### Why the refusals are few

Anthropic's Claude Code guidance warns that a reviewer asked for gaps
will report some even when the work is sound, and that chasing every
finding leads to over-engineering. This fold tested each finding against
that warning. Most findings name a state the plan left without a report,
or a proof that cannot fail as written. The repo's planning-miss rule
requires both closed before approval, and each fold costs a table row or
a fixture line, so refusing them would cost more than folding. Where a
fix went further than its risk, the fold kept the named state and refused
the extra mechanism (C11, m5, D7, D15). The one whole refusal (D12) is a
deferral to the rung whose caller can meet the state.

### How each finding was verified

The fold did not take a finding's evidence on trust where a fix would
rest on it. It re-ran or extended the probes in
`/tmp/claude-1000/-var-home-glw907-Projects-dubplate/82a7130f-dc63-4c92-b308-9af3d3bb0dde/scratchpad/6a-fold-probes/`
(`main.go`, 2026-09-23, `modernc.org/sqlite v1.59.0`, go1.27.1). Results
the revision cites:

| Probe | What it measured |
| --- | --- |
| P1 | D1 reproduced: after `database or disk is full (13)`, a later insert returned nil, `Rollback` answered "no transaction is active", and the table held rows 1 and 3. |
| P2 | M1's fix: `(&url.URL{Scheme: "file", Path: p}).String()` under a directory named `a#b%41?c` produced `a%23b%2541%3Fc`, opened the exact file with `mode=rw`, converted it to WAL, and created nothing beside the directory. |
| P3 | C9's fix: `mode=rw` on an absent file returned `unable to open database file (14)` and created nothing. |
| P4 | C2's fixture: a child that wrote with `wal_autocheckpoint(0)` and exited through `os.Exit` left a 12392-byte `-wal`. A `mode=ro` read saw all three rows with the main file's hash unchanged. The same read without `mode=ro` changed the hash and removed the `-wal`. |
| P5 | C6's message: `UNIQUE constraint failed: users.address (2067)`, and `users.streamer_username` for a case-folded duplicate. |
| P6 | C5's descriptor check: `os.Stat` of `/proc/self/fd/N` matched the renamed file by `os.SameFile`, while `readlink` returned the new name. |
| P7 | M4's ordering: converting the test's own shared lock to `LOCK_EX|LOCK_NB` succeeded 2000 of 2000 times against a tight `LOCK_NB` poller. |
| P8 | M6's fixture: a `BEFORE INSERT` trigger with `RAISE(ABORT, 'injected')` failed the ledger insert after `PRAGMA user_version = 2` with code 1811. After the rollback, `user_version` read 1 and the table and trigger were gone. |
| P9 | m2's fixture: `/dev/shm` is tmpfs (`0x1021994`), and the computed `00:1b:<inode>` matched the process's own `/proc/locks` line. |

The fold also read these sources directly:

- `internal/config/config.go`'s package comment ("resolves at its point
  of use rather than at Load") and `thresholdStatus` in `doctor.go`
  (M7).
- `tx.go`'s `newTx` (m1).
- `/usr/include/bits/sysmacros.h` for the `dev_t` split (m2).
- `magic.h` for the Ceph and AFS constants (m9).
- The `VACUUM INTO` pager-flags comment at
  `lib/sqlite_g_0000000000000003.go:77138` (m5).
- The deferred-close comment at `lib/sqlite_g_000000000001ffff.go:12569`
  (M4).
- `TestInMemoryDBSurvivesContextCancel`'s statement that an interrupted
  file-backed connection is discarded (C7).

## Convergent roots, folded once

| Root | Finding IDs | Where revision 2 fixes it |
| --- | --- | --- |
| The interrupted create is wider than a zero-byte file | C3, D2, M2 | R-6a5's table gains a row for a header-only file of any size, and R-6a6 is restated. Task 6 adds the header-only and zero-byte-with-`-wal` rows. |
| A partly created leaf | C4, D2 (leaf half), m3 (leaf half) | Task 1's create is idempotent, and inspect and write never create the lock file. R-6a5 names the missing-lock and missing-database states. Task 6 adds the listing-unchanged rows. |
| Staging cleanup's lock mode and owner | D4, M3 | R-6a7: serve mode alone cleans, holding the exclusive lock, in step 3. Write mode never cleans. Task 6 changes the rows and mutation 4. Task 8 carries the 6b-i stager rule. |
| Criterion 8's failure point | C1, M6 | Task 4's criterion 8 becomes three rows: the trigger after the runner's version write (P8), the in-file version write, and the hash mismatch. |
| The conversion proof's ordering | C5, M4 | Task 6's criterion 14 becomes two rows ordered by the logger signal. The test converts its own shared lock (P7), checks descriptors by `os.SameFile` (P6), places a store at a different version, and closes the other connections. |
| The context against the busy wait | C7, M5 | R-6a18 and Ruling 4. Task 3's row is corrected, and the connection discard is recorded. |
| Nothing may create a database except the pre-create | C9, D8, D16, D17 (creation row) | R-6a17: writing DSNs carry `mode=rw`, and classification and snapshot verification use `mode=ro`. The privileged-fields table names SQLite's create flag and restore's future channel. |
| Ledger and `user_version` disagreement | C13, D11 | R-6a5's table and Task 4's ledger check and rows cover a missing row and a row above `user_version`. |
| Retention's own names only | D10, m7 | R-6a11 (pattern-only, `keep` of at least 1), Task 4's rows, and a privileged-fields row. |

## Dispositions

### Contract lens (C1 to C13)

| ID | Rank | Disposition | Where, or why |
| --- | --- | --- | --- |
| C1 | major | Folded (root: criterion 8) | Task 4's criterion 8 now has three rows. Mutation 1 fires on the trigger row. |
| C2 | major | Folded | Task 6's criterion 15 uses a hot-WAL fixture (P4), and mutation 6 fires on it. |
| C3 | major | Folded (root: interrupted create) | R-6a5 and R-6a6. Task 6 adds the header-only row and the zero-byte-with-`-wal` row. |
| C4 | major | Folded (root: partial leaf) | Task 1's idempotent create. R-6a5's rows. Task 6's present-leaf, no-database rows assert an unchanged listing. |
| C5 | major | Folded (root: conversion) | Task 6's criterion 14 is two rows: the logger-signal order, the placed store at a different version, and the `os.SameFile` check. |
| C6 | major | Folded | Task 5's criterion 12: the duplicate differs in every column but the address, and the test asserts 2067 and `users.address` (P5). The 787 rows name their parents present. |
| C7 | minor | Owner fork: Ruling 4 (with M5) | The connection-discard fact is folded into R-6a18 and Task 3's autocommit fixture. |
| C8 | minor | Folded | Task 6: serve's step 1 shared wait is bounded by the migration wait and logs holders. It is an owed erratum (E8). |
| C9 | minor | Folded (root: no create) | R-6a17 (`mode=rw`, P3). Task 4 verifies the copy through `mode=ro` and asserts its version and marker. The privileged-fields creation row is updated. |
| C10 | minor | Folded | The new "Spec deviations that take effect on approval" list under the rulings. |
| C11 | minor | Folded in part | The ownership assertion joins Task 8's declared gaps. **Refused:** the inspect-class `query_only` readback, because Task 6's hot-WAL row proves `mode=ro` by behavior, which a pragma readback cannot. |
| C12 | minor | Folded | A global constraint allows no `t.Parallel()` in the store package. Task 1 restates it. |
| C13 | minor | Folded (root: ledger) | Task 4's ledger check, states, and rows. |

### Mechanics lens (M1 to M7, m1 to m9)

| ID | Rank | Disposition | Where, or why |
| --- | --- | --- | --- |
| M1 | major | Folded | R-6a17's one escaping builder (P2). Task 2 adds the `a#b%41?c` row and mutation 6. |
| M2 | major | Folded (root: interrupted create) | As C3. The header-only fixture is built the way probe D2 was and asserted to be 4096 bytes. |
| M3 | major | Folded (root: staging) | As D4. |
| M4 | major | Folded (root: conversion) | As C5. The deferred-close precondition is quoted and made a fixture state. |
| M5 | major | Owner fork: Ruling 4 | The measured behavior is stated in R-6a18. Task 3's row is corrected, with a pinning row under yes. |
| M6 | major | Folded (root: criterion 8) | As C1. The trigger shape is re-proved by P8. |
| M7 | major | Folded | R-6a3: the store row carries the range refusal. The floor is 1 ms. The threshold row shows the parsed value, as `thresholdStatus` does. "A load error" is removed, and Task 7's rows are corrected. |
| m1 | minor | Folded | The maintenance DSN gains `_txlock=immediate`. Task 2 adds a row proving it and mutation 8. |
| m2 | minor | Folded | Task 1 quotes the `dev_t` split from `sysmacros.h`. Waiter (`->`) lines are skipped. The own-pid fixture is on `/dev/shm` (P9). The btrfs gap is named in Task 1 and Task 8. |
| m3 | minor | Folded (roots: partial leaf, D3) | R-6a5 maps every state to the closed set. A ledger disagreement and an unreadable file are "refused" with their reason. |
| m4 | minor | Folded | R-6a10 gives the callback the marked `ctx` and documents the closure limit. Task 3 adds a row and mutation 9. |
| m5 | minor | Folded in part | `VACUUM INTO ?` binds the target. The pager-flags inheritance is quoted. The rollback-journal copy goes to the 6b-ii handoff. **Refused:** a hand `Sync` of the copy, and a firing mutation for `_synchronous=FULL`. The quote shows the copy already inherits `FULL`, and no test can observe durability, so the mutation is a reviewer-read row. |
| m6 | minor | Folded | Task 2 adds the 517 row, so mutation 4 fires. |
| m7 | minor | Folded (root: retention) | As D10. |
| m8 | minor | Folded | R-6a14: "every writing class". |
| m9 | minor | Folded | Task 1 refuses `CEPH_SUPER_MAGIC` and `AFS_SUPER_MAGIC` (read from `magic.h`). It is part of erratum E6. |

### Risk lens (D1 to D17)

| ID | Rank | Disposition | Where, or why |
| --- | --- | --- | --- |
| D1 | blocker | Folded | R-6a10: after a recorded error, no statement reaches SQLite (P1). Task 3 adds the outcome, two state rows, the autocommit criterion, and mutation 8. `IOERR`, `NOMEM`, and `INTERRUPT` are reviewer-read rows. Task 4 states the runner's first-error rule. |
| D2 | major | Folded (roots: interrupted create, partial leaf) | As C3 and C4. The create is idempotent and re-stats under the exclusive lock (Task 1). |
| D3 | major | Folded | R-6a5: absent comes from `os.Stat` alone, and no SQLite code maps to absent. Every other error is "refused". Task 6 adds the 0000-mode row and mutation 7. |
| D4 | major | Folded (root: staging) | R-6a7, Task 6, and Task 8's 6b-i handoff line. It is erratum E7. |
| D5 | major | Owner fork: Ruling 3 | The plan carries both answers' rows in R-6a5 and Task 6, with mutation 10 under yes. |
| D6 | minor | Folded | Task 5 pins each embedded migration's SHA-256 in a test table. Mutation 8 appends a newline. |
| D7 | minor | Folded in part | Task 4: the runner stops at a file's first error. A `Commit` that answers "no transaction is active" after a clean apply is a hard error. The diff-reviewer checks new files for transaction control, and a `commit` mid-file row proves the error. **Refused:** a gate test that scans migration text for those keywords. A trigger body's `BEGIN ... END` would false-positive, and tokenizing SQL correctly costs more than the risk the commit check already bounds. |
| D8 | minor | Folded (root: no create) | Task 4 verifies through `mode=ro` and asserts the copy's `user_version` and marker. |
| D9 | minor | Folded | R-6a3's ceiling is 2147483647 ms. Task 7 adds the `1000h` row and mutation 6. |
| D10 | minor | Folded (root: retention) | R-6a11, Task 4's rows and mutation 10, and the privileged-fields row. |
| D11 | minor | Folded (root: ledger) | As C13. |
| D12 | minor | Refused at 6a (deferred to 6b-i) | At 6a the checkpoint runs only under the exclusive lock, so the busy row cannot occur. The first caller that can meet it is 6b-i's backup under a shared lock, and Task 8's handoff carries the rule (read the row; busy is logged, never a failure). |
| D13 | minor | Folded | R-6a14 renames the constant "the post-checkpoint WAL size". R-6a11 states what the free-space check covers and names the mid-migration disk-full. |
| D14 | minor | Folded | Task 1: a create requires the effective uid to own the state directory. It is erratum E9. |
| D15 | minor | Folded in part | Task 1 probes `db/` when it exists. `statfs` follows a symlink, so a symlinked `db/` on NFS is caught. It is part of erratum E6. **Refused:** refusing a symlinked leaf or `db/` through `os.Lstat`. A symlink is a legitimate way to put the store on a larger local disk, and the probe already sees the target's type. |
| D16 | minor | Folded (root: no create) | R-6a5 and Task 6: classification runs through the inspect DSN. Task 6 adds a foreign rollback-journal row with unchanged bytes, and mutation 8. |
| D17 | minor | Folded | The privileged-fields table names restore (6b-ii) as the later creation channel, and the leaf and each subdirectory at 0700. Task 1 makes each directory 0700. |

## Rulings for Geoff

These are the plan's "Rulings for Geoff" section, stated here for the
record. The fold rules none of them.

1. **A 6.8M execution ceiling for 6a, with planning recorded separately?**
   (The drafter's Ruling 1, re-derived.) Recommend yes.
2. **`doctor` and `check` exit 0 for absent, older, and current, and
   nonzero otherwise?** (The drafter's Ruling 2.) Recommend yes.
3. **`serve` refuses to create a store beside a snapshot copy when the
   database is gone?** (D5.) Recommend yes.
4. **A caller's deadline bounds only the in-process wait, and the busy
   timeout alone bounds the cross-process wait?** (M5, fork A, with C7.)
   Recommend yes. No would split Task 3.

## Owed errata

The close (Task 8) drafts the paste-ready text. Each item changes the
ratified spec's meaning, or a row in the spec fold record, once the plan
is approved. The fold did not edit the spec.

| # | Source | Spec location | The change |
| --- | --- | --- | --- |
| E1 | Drafter 1, stands | The spec fold record's `main.go` contested row | The migration wait registers at 6d with `serve`, not at 6a (R-6a3). |
| E2 | Drafter 2, broadened by C3/D2/M2 | Criterion 13 (`:938`) | A zero-byte file and a header-only unmarked file are interrupted creates that serve completes. "Not a store" holds for the inspect and write modes (R-6a6). |
| E3 | Drafter 3, stands | "Package territory" (`:213-216`) | "6a exports the open" means the inspect mode, its options, and its report (R-6a2). |
| E4 | Drafter 4, re-figured | The ceiling table row (`:1098`) | 6a's figure follows Ruling 1's outcome: 6.8M execution plus planning, under yes. |
| E5 | Drafter 5, stands | Paths (`:480-487`) | The path refusal lands as schema CHECKs at 6a, and the Go validator lands with rung 7's first writer (R-6a9). |
| E6 | Drafter 6, broadened by m9/D15 | Filesystem probe (`:256-258`) | The probe also refuses SMB, Ceph, and AFS. It probes `db/` when it exists, else the leaf. |
| E7 | New, D4/M3 | Layout (`:235-236`) | Only serve mode, holding the exclusive lock with no intent record, removes staging leftovers. Write mode never does (R-6a7). |
| E8 | New, C8 | Serve procedure step 1 (`:351`) | Step 1's shared wait is bounded by the migration wait and logs the holders, as step 2 does. |
| E9 | New, D14 | Ownership (`:238-242`) | A create requires the effective uid to own the state directory, and otherwise refuses naming that owner. |
| E10 | New, M5/C7, conditional on Ruling 4 | The write method's contract (`:331-332`) | Under yes: the busy timeout alone bounds the wait on another process's write lock, and a context does not shorten it (measured). Under no: a per-call busy timeout replaces this. |
| E11 | New, D5, conditional on Ruling 3 | Layout and the serve procedure's create | Under yes: serve refuses a create when `snapshots/` holds a copy, naming it and `restore --from-file`, and inspect reports it as refused. Under no: no erratum. |

The drafter's six candidates all stand. E2 and E6 are broadened, and E4
takes the new figure.

## Ceiling change

The fold adds about two dozen named states and reworks three proofs
(criteria 8, 14, and 15). The work lands mostly in Tasks 3, 4, and 6, and
adds no task. The plan's arithmetic gains a 0.30M line, about half a task
at the measured 0.533M rate. The projection moves from 4.62M to 4.92M. To
keep the flag about one task above the projection, the recommended
ceiling moves from 6.4M to 6.8M, and the flag from 5.12M to 5.44M.

Task 6 was weighed for a split. Its count stays at four: staging cleanup
joins the serve mode, and the shared classification takes its place. The
plan pre-names the split instead: serve mode moves to its own task if
the implementer cannot land all four in one green gate. That split costs
about 0.53M, which the flag's headroom covers.
