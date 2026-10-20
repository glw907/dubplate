# Rung 6d plan review: domain risk lens

Target: `docs/superpowers/plans/2026-09-26-rung-6d-identity-serve-store.md` at `5afe46b` (revision 1).
Lens: data integrity, auth, and failure. The review read the plan, the rung 6 spec (`:700-880`), the
6c HISTORY handoff, the 6c architecture reads, the approved 5d plan (serve unit, sitting steps
12-14), and the code the plan rewrites: `internal/auth/{roster,seam,handlers,admin,nonce,session}.go`,
`internal/engine/store/{open,write,leaf}.go`, `internal/settings/{apply,accessor,evaluate}.go`,
`internal/config/settings.go`, `cmd/dubplate-server/auth.go`.

**Counts: 0 blockers, 3 major, 10 minor. No owner forks.** Plan line numbers below are the plan
file's own.

## What is sound

- **R-6d1's ordering is safe.** The 5d sitting leaves the serve unit disabled and unrouted, starts it
  once on loopback (5d plan step 14), and never creates a store: `OpenWrite` refuses an absent store,
  so no pre-6d store or snapshot can exist on the box. The real 6d upgrade therefore moves a file
  roster into a fresh store. It never migrates a populated store. The migration and restore-of-older
  cases are hardening for later upgrades, not for this one.
- **Atomicity.** The import, the seed apply, and the reactivation each run in one `Update`. That is
  one `IMMEDIATE` transaction that "commits only when fn returns nil and no statement fn issued
  through tx failed" (`write.go:44-63`). A crash leaves either none of the write or all of it. Two
  concurrent imports serialize on the write lock.
- **Next-operation nonce retention is safe.** Each nonce record carries its own sweep time, the later
  of `consumedAt+retention` and the token's `expiresAt` (`nonce.go:160-166`). Lowering the retention
  live cannot open a replay window.
- **A live-changed cap cannot revive an expired session.** `VerifySession` checks the cap first and
  then the signed `ExpiresAt` (`session.go:219-233`), and `ExpiresAt` was clamped to the cap at mint
  or refresh. Raising the cap re-admits only sessions whose signed expiry is still ahead. The two
  correlated rules are therefore not load-bearing at read time: `clampExpiry` and `Insert` enforce
  both invariants themselves.
- **Lookup keeps the kind-answer guard.** R-6d10 makes a store error "returned as itself, never as
  not-found". That preserves `handlers.go:210-222`, which refuses to hand the kind answer to a member
  when the roster is unavailable.
- **The `sensitive` erratum (R-6d6) weakens nothing.** No setting carries `Sensitive` today (grep:
  only the field and its two branches). CH rules that "a settings key never holds a credential". A
  closed-set string cannot carry a secret either.
- **Every reader fails closed.** R-6d9's rule that a failed read is never defaulted is right. The
  admin gate turns any seam error into a 404 without touching the cookie (`admin.go:116-128`).
- **The update paths have deadlines.** The cookie path has one (`admin.go:255`) and so does the CLI
  path (`member.go:95`), so the `ctx` change in R-6d10 inherits `Update`'s deadline rule.
- **Handoff item 6 cannot strand `serve` at 6d.** No channel writes an `admin` row, and a snapshot
  from a later binary is refused as newer before the seed runs.
- **Rulings 2 and 3: agree with both recommendations.** With no `Bump` caller and no removal path,
  the address-keyed epoch cannot diverge from `UserID`. Two stores sharing one remote first becomes a
  live risk when backup runs unattended, which starts at rung 16.

## Findings, ranked by consequence

### Major

**M1. `AddMember`'s two in-transaction authorization rules are dropped from R-6d10 and Task 3.**
Location: plan `:360-367` (R-6d10) and `:641-655` (Task 3's states).
Defect: today's write path enforces two rules under the lock. First, a cookie-channel actor must hold
the admin flag in the copy read under the lock (`roster.go:752-755`, `addTo` `:796-805`). Second,
"the roster written must name at least one admin" (`errNoAdmin`, `roster.go:151`, enforced at
`:460`). R-6d10 lists the reads inside the callback as duplicate, tombstone, and username only. Task
3's state table has no row for either rule, and the security-review checklist (`:467-477`) does not
name them. An implementer following the plan literally loses both:
- the in-transaction admin re-check, which is the guard a demotion path (rung 14 or the revoke verb)
  will rely on;
- the no-admin refusal. A first `member add` without `--admin` then succeeds, and the store is left
  with no admin. Task 4's doctor row "members but no admin" (`:701`) assumes that state can happen.

Fold: add both rules to R-6d10's in-callback list, and add three Task 3 rows:
- a cookie-channel actor whose live row is not an admin is refused, with nothing written;
- a cookie-channel actor with no live row is refused;
- a command-line add of a non-admin to a table with no admin is refused with `errNoAdmin`.

Add two named mutations: read the actor's admin flag before the `Update`, and drop the no-admin
check. Add one checklist line to the security reads.

**M2. `member import` loses the roster file's writable-by-others refusal.**
Location: plan `:729-731` (Task 5 Files: "`parseRoster`, `validateRoster`") and `:742-753` (states).
Defect: the import is the one channel besides the CLI grant that sets `users.admin`, and it takes the
flag from a file. Today `readRoster` refuses a file that is group- or other-writable
(`roster.go:318-320`, `errRosterWritableByOthers`), and 5d's plan names that refusal as the roster's
integrity control (5d plan `:349`). Task 5 reuses only `parseRoster` and `validateRoster`, which do
no mode check, and its state table has no mode row. Another local account that can write the file
could then add itself as an admin through an operator's import.
Fold: the import opens and stats the file the way `readRoster` does. It refuses a mode with
`0o022` set, naming the file and the mode, and it hashes the same bytes it parses. Add the state row
"the file writable by group or others: refused naming the file and its mode", with its own fixture
and a named mutation that drops the check.

**M3. Task 8's "restart `serve`" would start the unit 5d keeps disabled and stopped.**
Location: plan `:880-882` (Task 8 outcomes); see also R-6d1 `:256-266`.
Defect: the 5d unit is installed disabled. 5d's `deploy.sh` "never enables or starts" it (5d plan
`:520`, `:1001`), and the unit stays that way until the rung 16 cutover. `systemctl restart` on a
stopped unit starts it. If Task 8 adds a plain restart, every later deploy leaves `serve` running
on the production box without a ruling. It creates the store, holds the shared lock indefinitely
(which blocks `restore`), and answers logins against an empty `users` table. A related ordering
gap: the plan does not say that `check` runs after the env is re-rendered and the new binary is
installed. With the stale env, the superseded `DUBPLATE_ROSTER_FILE` fails `check` and halts the
deploy. The reverse order would let a bad env through.
Fold: specify `systemctl try-restart` (it restarts only an active unit), preceded by the owner-run
`check`. Order the steps as: env render, then binary install, then `check`, then halt or
`try-restart`. `deploy.sh --check` shows `try-restart`. Add a named mutation: replace `try-restart`
with `restart`, and a `--check` golden fails.

### Minor

**m1. The import admits usernames and display names the write path refuses.**
Location: `:729-731`, `:742-753`, and the privileged-fields row for `users.streamer_username` at
`:242`.
The loader's `validate` checks only that `streamer_username` is ASCII-printable
(`roster.go:582-585`) and allows a blank display name. `validateWriteCandidate` enforces
`streamerUsernamePattern` and a non-blank name (`:602-617`). Every other channel writes the narrow
shape, and rung 12 injects the username into the proxy header. NOCASE agreement still holds because
the loose check is ASCII too. Fold: run each import entry through `validateWriteCandidate` and report
each refused entry, or state that the looser shape is intended and why. Add a state row either way.

**m2. R-6d4 can strand `serve` on a keep row the seed would heal.**
Location: `:293-298`.
R-6d4 refuses a stored keep row outside its bounds before any migration. But `outcome` heals an
out-of-bounds `env-seed` row that the env re-seeds or removes (`evaluate.go:275-280` doc). Only
`serve` seeds, so a later binary that narrows the keep's bounds would refuse every start, and no
in-product channel could repair it. Fold: take the keep from `EvaluateFacts`'s outcome for the key.
Refuse only when that outcome refuses; otherwise pass the post-seed value in force. Add one state row
to Task 1.

**m3. The upgrade runbook deletes the only rollback input.**
Location: `:885-888`.
5d keeps three `.prev` copies as the rollback until the next attended deploy (5d plan step 15). The
spec says a rollback past 6d "needs the old template and the old file" (spec `:834-836`). Deleting
the file right after `member list` removes it. Fold: the runbook keeps the file until the `.prev`
copies are superseded, then deletes it.

**m4. States not named in the plan.**
Planning-miss item 2.
- Task 1 (`:540-551`) lacks "leaf present and empty". That is the state every fresh box reaches,
  because `deploy.sh` pre-creates the leaf and `create` must accept it (`leaf.go:70-99`).
- Task 4's members row (`:697-703`) and `Facts` lack "store older than `0006`". There the
  `member_imports` table is absent and the row must read "no import" instead of failing the
  query.

Fold: one row each.

**m5. Tombstoned rows against the `UNIQUE` constraints.**
Location: `:360-367`, `:651`, `:751`.
`users.address` and `users.streamer_username` are `UNIQUE` across tombstoned rows (`0003:3,5`). A
new member whose username equals a tombstoned row's username (a different address) passes a
live-rows-only collision read and then fails on the constraint as a generic store error. The plan
also does not say whether an import entry on a tombstoned address reactivates or refuses. Neither
case can happen at 6d, but both are cheap to state. Fold: the username collision read covers
tombstoned rows and refuses by name. The import refuses a tombstoned address, naming it, and leaves
reactivation to the verb.

**m6. The settle loop has no overall bound.**
Location: `:414-424`.
The case for no bound assumes each round needs another process to win a lock conversion. It also
assumes that `settle` returning nil guarantees a store at head. If `settle` ever returns nil without
reaching head, the loop spins with no lock contention and logs a warning every pass. Fold: bound the
loop's total elapsed time by `MigrationWait`, which does not refuse a legitimate operator sequence
inside the wait, and add a test with a stub that never converges.

**m7. Read ordering in the link handlers is not pinned down.**
Location: `:342-350`, `:846-857`.
- Under the generic answer, a failed settings read must not make responses differ by membership.
  Read the answer (and the body cap) before `Lookup`.
- In `consume`, the lifetime and cap reads must happen before `Nonces.Insert`. Otherwise a failed
  read burns the link without minting a session. R-6d9's "before any write" covers this in words,
  but no criterion proves it.

Fold: add a Task 7 row, "lifetime read fails in consume: 500, and the nonce is not consumed (the same
link then succeeds)", with a named mutation that moves the read after the insert.

**m8. `OpenNonceStore`'s `linkExpiryCeiling` must stay the bound, never the live value.**
Location: `:833-843`.
`mtimeFallback` uses `max(retention, linkExpiryCeiling)` so that an unreadable record outlives every
token that could have produced it (`nonce.go:69-76`, `:286-291`). Task 7 turns values into readers,
and the constant becomes registry `Ceiling` data. Fold: one line saying the argument stays the
registered ceiling of `DUBPLATE_MAGIC_LINK_EXPIRY`, never the value in force.

**m9. Task 6 does not prove that `check` still fails on an auth refusal.**
Location: `:799-809`.
R-6d8 moves the auth bounds out of the load, which fails every verb, into the seed evaluation. The
deploy gate then depends on 6c's Ruling 2: `check` exits nonzero when the seed would refuse.
Task 6 proves the `backup` half but not the `check` half. Fold: add the row "`check` with an
out-of-bounds auth env value: exits nonzero naming the key".

**m10. Import-record concurrency and the spec's restore-report line.**
Location: `:376-385`, `:769-770`, `:414-420`.
- The plan never says the "record present" check runs inside the `Update` callback. The named
  mutation covers only the record write. Fold: state it, and add a test that runs two imports
  concurrently (exactly one succeeds). A schema singleton on `member_imports` would be cheaper
  still.
- The spec's "the report says to run `member import`" after a pre-6d restore (spec `:715-716`) is
  not carried into R-6d16. No pre-6d store exists anywhere (see "What is sound"), so a fold may
  refuse this half with that reason.

## The four planning-miss items

1. **Quoted behaviors.** Mostly present. Missing: the `UNIQUE` constraints' reach over tombstoned
   rows (m5), and `readRoster`'s mode refusal, which the import would silently drop (M2).
   `systemd-run -p EnvironmentFile=` is not quoted, but 5d's sitting already uses it successfully.
2. **States named.** The gaps are in M1, m2, m4, m5, m7, and m9.
3. **Privileged-field channels.** The `users.admin` row (`:241`) is right about who may set the
   flag. It is incomplete about what the write must re-check (M1) and what protects the import's
   source (M2). The `users.streamer_username` row admits a wider shape through the import (m1).
4. **Fixture states.** Adequate. The Task 7 limiter proof (`:851`) should state the request count
   that trips limit A but not limit B, or the "uses A until restart" case can pass vacuously. The
   raised-cap case needs a session whose signed `ExpiresAt` is later than its age at the lowered
   cap.
