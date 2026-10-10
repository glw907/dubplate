# Rung 14a plan fold

Target: `docs/superpowers/plans/2026-10-10-rung-14a-admin-api.md`, revision 1 at `dd161b3`, folded to revision 2.
Reviews: `2026-10-10-rung-14a-plan-review-contract.md` (PA), `-mechanics.md` (B), and `-domain.md` (D). Spec:
revision 2.1, unchanged. Each finding was checked against the tree at `dd161b3` before its disposition; the checks
that decided a disposition are named in its row.

The fold rule (CLAUDE.md, "Findings are ranked, and a fold may refuse"; Anthropic's best-practices guidance, "Chasing
every finding leads to over-engineering") refuses a finding whose fix costs more than the risk it removes. New
machinery folds only with a named source that runs it and a measured defect.

## Convergent defects, fixed at the root

| Root | IDs | Fix |
|---|---|---|
| Task 10's retirement reaches surfaces outside its files and 14a's territory | B1, PA-2, D14 | Task 10 split into 10a (members page), 10b (settings page), 10c (retirement, `auth-data`); spec deviation 2; lane launch step 3; owed spec erratum 1 |
| The close's code folds land after gates, and the close runs three full gates | PA-3, B5, D15 | Task 11 reordered: every read, one fold, the simplifier, rebase, one close gate (`--merge`), `--ff-only`, the on-master gate |
| Hook 18's mark conversion has no proof that existing sensitivity survives | PA-4, B9, D4 | Task 2's production-registry pin and its mutation; per-family zero-mark fixtures |
| The member cap has no carrier into `internal/auth`, and the roster's new dependencies are unshaped | PA-6, B3, B4 | P11; the accessor quote; Task 4's files, interface, and states |
| AC16's concurrency proof names no ordering point | PA-8, B8 | Task 4's fixture: a fixed hold far inside the busy timeout, ten-of-ten mutant runs, no store hook |
| The bootstrap row is unreachable, and its order against the cap read is open | PA-13a, D10 | P11 (bootstrap first, reads no cap); the states row and Task 4's fixture restated reachably |
| The window's correlated rule cites a non-strict precedent and its fixtures are missing | PA-12 (window items), D9 | P6 strict, `durationAtLeast` quoted; the `serve`-start states row; Task 6's fixtures and mutation |
| The PS sentence on a command-line cap raise goes unrecorded | PA-16b, D11 | spec deviation 1's last sentences; Task 11's PS erratum |
| The scoped gate's cumulative base and double web build | PA-17, B12 | filed to BACKLOG at boundary A (lane launch step 8); no plan mechanism |

## Dispositions

### Lens P-A (contract and criteria)

| ID | Sev. | Disposition |
|---|---|---|
| PA-1 | major | **Folded.** Verified `session.go:76-86` (a refresh moves `IssuedAt`, not `AnchoredAt`). P2 names the anchor; privileged-field row added; AC12's refresh clause, its fixture, and the `IssuedAt` mutation moved to Task 7; Task 6's mutation restated. |
| PA-2 | major | **Folded** with B1. Verified `run.ts:866`, `:3098`, `rendered_test.go:218-219`, `:860`; no vitest in `web/package.json`. Tasks 10a-10c list the files; page proofs restated as rendered rows (global constraints). Owed spec erratum 1. |
| PA-3 | major | **Folded** with B5, D15. Task 11 reordered. `internal/identity`'s architecture read dropped (doc-comment change only), with the reason in step 2. |
| PA-4 | major | **Folded** with B9, D4. Verified three `Sensitive: true` settings (`auth.go:126-155`). Task 2 outcome, acceptance, fixture, mutation; Tasks 4 and 6 extend the pin. |
| PA-5 | minor | **Folded.** P5: the sign-in link opens a new tab; Task 10b's stale-session row proves the edit survives and Retry commits. |
| PA-6 | minor | **Folded** with B3, B4 (P11, Task 4). |
| PA-7 | minor | **Folded.** Task 4: both bump-failure fixtures and the "commit despite a non-backward bump error" mutation. |
| PA-8 | minor | **Folded** with B8, in the conventional form: a fixed hold far inside the busy timeout, mutant failing ten of ten. The `DBStats.WaitCount` form is refused: it counts waits for the pool's connection, not SQLite's busy wait on a lock another connection holds, so it would not observe this fixture. |
| PA-9 | minor | **Folded.** Task 7: no cookie, removed member, refresh, self-removal fixtures; "answer a gate refusal with 500" mutation. |
| PA-10 | minor | **Folded.** Task 7 logs `principal` as its own field and its acceptance cites AC6's log clause; Task 1 adds the principal accessor (Task 6's notice also reads it). |
| PA-11 | minor | **Folded.** Task 5 states AC13's command-line and seed-rule clauses hold by construction (no mailer parameter; only `internal/ops/admin` sends). Task 7 carries AC21's first clause with the live non-admin fixture. |
| PA-12 | minor | **Folded** (window items with D9). Task 6: erroring window reader, `busy` on both write kinds with the `ErrBusy` mutation, equal-to-cap and one-second-below windows, a blocking mailer with a concurrent writer. |
| PA-13 | minor | **Folded.** (a) with D10: the unreachable mutation dropped, the bootstrap row restated over cap-many tombstones and an unreadable cap. (b) Task 8's `nonceRetentionCoversExpiry` reset fixture (verified at `auth.go:255-259`). |
| PA-14 | minor | **Folded.** Task 5's test-first #25 fixture; Task 8's "(required)" assertion. |
| PA-15 | minor | **Folded.** Task 7's V4 test asserts each job schema's property set equals the Go type's JSON fields. |
| PA-16 | minor | **Folded.** (a) Task 11 step 9's STATUS bullet carries the `internal/serve` question as an open decision (verified in the fold's conductor-owed list). (b) with D11. |
| PA-17 | minor | **Folded** with B12 as a BACKLOG filing at boundary A. A `check.sh` base override is new gate machinery with no measured defect on this pass (the cost is priced), so it is filed, not built. |

### Lens P-B (mechanics and feasibility)

| ID | Sev. | Disposition |
|---|---|---|
| B1 | blocker | **Folded** (root above). Verified `deck-coverage.ts:155-167` (`admin_not_grantable`, the 21-outcome floor), `admin_test.go:75`, `:265`, `serve_test.go:810-1208`. Re-priced: +2 tasks. |
| B2 | major | **Folded.** Verified `testOnlyReduces: true` for `engine-logic` (`pass-execute.js:310-316`) and `--reduced` skipping the rendered leg (`check.sh:261-263`). Tasks 10a and 10b set `reducedGate` to the per-task gate; 10c is `auth-data`, which never reduces a test-only round. |
| B3 | major | **Folded** with PA-6. Verified `internal/auth` imports no settings or config package, and `accessor.go:20-21`. P11; the quote; Task 4. |
| B4 | major | **Folded.** Verified `NewRoster` callers (`member.go:97`, `:133`, `serve.go:324`, and five test files) and `auth.Open`'s behavior (`store.go:49-58`: it creates an absent leaf, so "leaf absent" is not a refusal). P11 keeps the epoch store off reads; Task 4 lists the callers and moves `member add`'s auth-leaf opening in; states row added. |
| B5 | major | **Folded** with PA-3. Verified HISTORY:229's precedent. `check.sh --merge` serves as both the close gate and the in-tree merge gate, after the rebase, as CLAUDE.md's ritual orders. |
| B6 | minor | **Folded.** Verified `membersRowStatus`'s non-failing advisory (`doctor.go:223`) and the print loop (`:160-164`). P9 and Task 9 drop the new warning state; the row's states join the states table. Removes a mechanism. |
| B7 | minor | **Folded.** P12; the empty-string quote; Task 6's fixtures and "drop the length check" mutation. The all-zero id answers `invalid_request` (`identity.go:205-208`: the zero value is never resolved). |
| B8 | minor | **Folded** with PA-8. The store wait hook is refused (the spec fold refused a barrier hook; no measured defect). |
| B9 | minor | **Folded** with PA-4. |
| B10 | minor | **Folded.** Lane launch step 6 uses Task 0's timed gate. |
| B11 | minor | **Folded.** Lane launch steps 1 and 2 run before Task 0. |
| B12 | minor | **Folded** with PA-17 as a BACKLOG filing. Verified `check.sh:332-335` and `:371-374`. |
| B13 | minor | **Folded.** Verified the runner passes only `{ model }` (`pass-execute.js:857`); the effort claim is dropped. |

### Lens P-C (domain risk)

| ID | Sev. | Disposition |
|---|---|---|
| D1 | major | **Folded.** Task 11 step 1 widens the boundary-B read to Task 3's authorizer and Task 4's roster and epoch code, with its four questions; step 2's read covers 8 and 10a-10c. |
| D2 | major | **Folded.** One rule: after the fix round, one more dispatch (Opus upshift or seam split), then stop. The close's red smoke, page defect, close gate, and on-master gate each have a rule. |
| D3 | major | **Folded.** Four ceiling rules in "The hard ceiling": the B read belongs to segment B, the close runs to the hard ceiling, the merge ritual starts only with 0.5M left and never stops between `--ff-only` and the on-master gate, STATUS names the step. |
| D4 | major | **Folded** with PA-4. |
| D5 | major | **Folded.** Verified spec:244 ("audits the same way"). Task 5 outcome, acceptance, and two mutations; Task 8 asserts the CLI reset's row and mutates it. |
| D6 | minor | **Folded.** Verified `main.go:99-108` and `admin.go:265`. Task 4 lists `main.go` and the `admin.go` call site; global constraints name the registration line. |
| D7 | minor | **Folded.** Task 4: "bump before the guard" with the refused self-removal fixture; "count live members outside the `Update`" with the concurrent-adds fixture. |
| D8 | minor | **Refused.** Task 6's AC16 settings-write clause already drives the only cookie-channel caller (`internal/ops/admin`) through the real guard, so a nil guard there fails that test; a defensive nil check in `internal/settings` adds code for a path no caller reaches (go-conventions: no defensive nil checks). |
| D9 | minor | **Folded** with PA-12. Verified `durationAtLeast` refuses only `l < s` (`auth.go:276-291`). |
| D10 | minor | **Folded** with PA-13a. |
| D11 | minor, fork | **Refused as an owner fork; folded as a record.** The spec settles it: the fold's C9 folded "the command line sends no notice" (erratum 5) and AC13 says a seed-rule change mails no one, so option (b) would reopen the approved spec. Task 11 writes the PS erratum (option a). |
| D12 | minor | **Folded.** States row; Task 4 outcome, fixture, and mutation. |
| D13 | minor | **Folded.** Two smoke steps (CLI removal while `serve` runs; the cap on both channels); the `5m` wait runs under the other steps; HISTORY routes the three sandboxed verbs to rung 16. |
| D14 | minor | **Folded.** Task 10a names the last-admin sequence and the roster reset; Task 10b reaches the stale state by intercepting the first write with a JSON 403, the harness's existing interception pattern, so its shape stays rung 7's. |
| D15 | minor | **Folded** with PA-3: Task 11 step 6 re-runs the smoke's auth steps after a non-fast-forward rebase that touches the auth or serve paths. |

## Owed spec errata (not edited here; Task 11 step 9 writes them)

1. **Spec, "Contested surfaces with the rung 7 lane":** the `web/` row and the `internal/auth/admin.go` row grant 14a
   only appended rows and the handlers' removal. Retiring the roster API also edits in place the harness's members
   rows and helpers, deck-coverage's roster outcome source, the `RequireAdmin` table's test host, and, until the
   removal, the roster API's `AddMember` call site. The plan records it as spec deviation 2.

PS also gains the erratum to PS:145-147 (spec deviation 1), from the fold's C9; it is a PS note, not a spec erratum.

## Rulings for Geoff

None. D11 was the only finding marked a fork, and the approved spec settles it.

## Counts

45 findings: 1 blocker, 13 majors, 31 minors. 43 folded (convergent IDs counted each), 2 refused (D8; D11 as a
fork, its record folded). No owner forks.

## Measures

- **New-mechanism findings:** 4 proposed, 0 folded as built machinery.
  - Refused: the store-level wait hook or `WaitCount` poll (PA-8, B8), the conventional timed hold taken instead; the
    fail-closed guard check (D8).
  - Filed, not built: the `check.sh` per-task base override and the double web build (PA-17, B12).
  - One mechanism removed: doctor's new warning state (B6).
- **Plan length:** 609 lines before, 819 after (+34%). The growth is the Task 10 split (three tasks for one), the
  fixtures and mutations the majors named, and the quoted behaviors; every addition is a proof, a file, or a state a
  task needs.
- **Ceiling:** 10.5M to 11.8M (flag 8.4M to 9.4M). Two more build tasks at 0.6M each, about 0.2M of added fixtures,
  and 0.1M less overhead (the `internal/identity` read dropped).
- **Clock:** about 6 hours to about 7 (fix rounds toward 7.5 to toward 8.5). Two more tasks at about 22 minutes, a
  third boundary gate, and a close about 6 minutes shorter (one full gate removed, two smoke steps added, the `5m`
  wait overlapped). The lane-launch baseline gate, outside the estimate, is removed.
