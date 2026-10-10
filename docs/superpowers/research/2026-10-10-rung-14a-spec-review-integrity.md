# Rung 14a spec review, lens C: data integrity and failure risk

Target: `docs/superpowers/specs/2026-10-10-rung-14a-admin-api-design.md` at `9d730d7`. Code read at the same commit:
`internal/auth/{epoch,roster,seam,session,store}.go`, `internal/settings/{accessor,evaluate}.go`,
`internal/engine/store/write.go`, migrations `0004` and `0005`, `cmd/dubplate-server/member.go`,
`internal/backup/doc.go`, R6 "The epoch" and its criteria, CH:162-173 and CH:576-586, HISTORY R-6d2.

Findings are ranked by consequence. The ruled points (per-album lease, 0.8-callers-only scope, the 14a pages,
seed-only `DUBPLATE_ADMIN_EMAIL`, the 14a/14b split) are not reopened.

## Findings

### C1. major. `reset_setting` puts the registry default live, not the next start's value, and skips the correlated rules

Location: spec:129-130, spec:123-124, AC10 (spec:369-371).

Defect. `reset_setting` deletes the `admin` row. A running `serve` sees the generation bump and reloads. With no row,
`Accessor.value` returns `s.Default` (`internal/settings/accessor.go:113-115`), not the environment's value. When the
serving environment sets the key to X, the live value is the default until the next start re-seeds X. A
live-applied threshold then runs at a value nobody chose. Second, only `set_setting` validates "every correlated
rule over the post-write set". A reset can leave a pair that breaks a correlated rule, such as a lifetime above the
absolute cap. The accessor validates each key alone (`:117`), so the broken pair applies live. At the next start,
`evaluate` refuses (`evaluate.go:217-238`) and `serve` does not come up. A deploy restarts it, so a web reset can
take the server down at the next deploy.

Fold. Give both writes one validator: run the seed rule (`evaluate`) over the post-write rows and the serving
environment, and refuse the write when it refuses. Reset then writes what the next start would write. When the
serving environment sets the key, reset replaces the `admin` row with an `env-seed` row carrying the environment's
value. Otherwise it deletes the row. The live value and the next start's value then agree, and no web write can
make the next start refuse. Add one AC: a reset that would break a correlated rule answers `invalid_request` and
changes nothing, and after a reset the live read equals the value `evaluate` puts in force.

### C2. major, OWNER FORK. A web write over an environment-seeded key arms a start refusal

Location: spec:123-125, spec:131-133, AC11 (spec:372-373).

Defect. R6's seed table refuses `serve`'s start for an `admin` row whose environment value differs (R6:809;
`evaluate.go:292-294`). Suppose the environment file sets a key to X, which seeds an `env-seed` row. An admin
then sets Y on the settings page. The write commits with a warning in the answer, and the next restart refuses.
That restart may be a deploy, a reboot, or the unattended upgrade path. Recovery then needs box access, because
the settings page is down with `serve`. This affects every key the operator seeded, which is every key the
template sets. The warning is computed against the environment `serve` started with. If the operator has edited
the environment file since, the warning can be wrong in either direction.

Options:
- (a) Refuse the write when the serving environment sets the key to a different value. The settings page shows
  such a key as "set in the server's environment file" with its value, and offers no edit. C1's validator does
  this with no extra code.
- (b) Keep the spec as written: commit the write and warn in the answer.
- (c) Change the seed rule so that an `admin` row wins with a `doctor` warning. This reopens R6's ratified table.

Recommendation: (a). It keeps the seed rule as ratified, removes the warning machinery from AC11, and makes a
start refusal reachable only from the box. An `env-seed` key whose value the admin writes back unchanged still
commits.

### C3. major. A removed member's old sessions can verify again after a re-add, because only the removal bumps

Location: spec:191-209, AC15 (spec:382-384).

Defect. The spec relies on one bump at removal, which `CheckSession` compares against the session anchor
(`epoch.go:225-241`). That record can be lost or outrun in two ways.

1. The epoch is a file under the auth leaf (`internal/auth/store.go:12-25`). A backup carries only the library
   and the store snapshot (`internal/backup/doc.go:1-9`). After a restore to a rebuilt box, the store holds the
   tombstone but no epoch record exists. The session secret comes from the age store, so it survives. Re-adding
   the address revives the old id (`roster.go:415-417`), and every pre-removal cookie verifies again. A cookie
   left on a shared laptop is the case CH guards.
2. A magic link can be consumed between the bump and the commit. Both run inside one write callback, and a
   cross-process `member remove` widens the gap. Consumption writes no store row, so the write lock does not block
   it. The new session's anchor is later than the epoch. The tombstone refuses that session while the member is
   removed, and the re-add revives it.

Fold. The reactivation branch of `addUser` (`roster.go:415-417`) also bumps the address's epoch, inside the same
callback and in the same order as removal. A re-add then revokes every credential anchored before it, whether or
not the removal's record survived. Keep the removal bump: it also refuses an outstanding magic link for its whole
expiry. Extend AC15: a re-add with the removal's epoch record deleted still refuses a pre-removal session. This
also covers C11's re-run bootstrap path.

### C4. major (over-ceremony with integrity risk). The resolving verb's older-store write duplicates the rollback recovery R-6d2 designed

Location: spec:157-168, AC14 (spec:379-381), the planning-miss states (spec:307-308).

Defect. `Handle.Update` refuses any write whose `user_version` differs from the binary's (`write.go:44-63`). To
write an older store, the verb needs a new open and write mode that bypasses that guard. Newer code would then
write an older schema. The store's most protective check would gain an exception, plus a matrix of five states
and a fixture of older stores. R-6d2 (`docs/HISTORY.md:411-416`) keeps an older store unmigrated precisely so
that "the rollback binary" can still run it. Both refusals the verb targets already have a recovery without the
verb:
- An environment contradiction is fixed in the environment. The refusal itself says "unset X, or set it to Y"
  (`evaluate.go:293`). The spec says this refusal cannot be fixed "through the environment alone" (spec:160), and
  the code contradicts that.
- An `admin` row outside a narrowed bound passes the previous release's bound. Roll back, reset the row from the
  page or with the previous release's verb, and upgrade. No older store can hold an `admin` row before 14a,
  because 14a adds no migration. Every store that can reach this state was therefore written by a binary that has
  the verb.

Fold. The verb opens through `store.OpenWrite`, at the binary's version only. Each seed-rule refusal message
names its own recovery. For a current store that is `settings reset`. For an older store it is redeploying the
previous release and then `settings reset`. Drop "older with the settings table" and "older without it" from the
states, and drop the older-store clause of AC14. If the owner wants the convenience anyway, three constraints
apply. The mode takes `OpenWrite`'s store lock and restore-intent check. It reads the version inside the
`IMMEDIATE` transaction, so a concurrent `serve` migration cannot land between the check and the write. And it
pins the settings tables' shape for every version it accepts, so a later migration that alters them narrows the
accepted range.

### C5. minor. A settings write does not re-read the actor's admin flag inside the transaction

Location: spec:123-130, spec:187.

Defect. Removal and demotion re-read the actor's flag inside the write (spec:187), as `addUser` does
(`roster.go:379-389`). The settings writes say nothing. The opshttp authorizer reads `Admin` once, when the
request is verified (`seam.go:132-141`). An admin demoted or removed by a concurrent CLI verb can therefore still
commit a sensitive write that is already in flight. That makes demotion non-final for one request, on the most
dangerous writes.

Fold. `set_setting` and `reset_setting` re-read the actor's live admin row inside their transaction, keyed on the
`UserID`. Extend AC16's "an actor demoted before the write is refused" to the settings writes.

### C6. minor. `ErrEpochBackward` rolls a removal back

Location: spec:203-205.

Defect. `Bump` returns `ErrEpochBackward` when the stored epoch is later than now (`epoch.go:152-153`). A remove,
re-add, and remove sequence across a backward clock step therefore makes "a bump failure rolls the removal back"
refuse the removal. The member then cannot be removed until the clock catches up. The stored, later epoch
already revokes everything anchored at or before now.

Fold. The removal treats `ErrEpochBackward` as satisfied, and any other bump error rolls back. Add one test.

### C7. minor. The gate's coverage of a sensitive reset is unstated, and so is its behavior when the window cannot be read

Location: spec:137-148, AC12 (spec:374-376).

Defect. AC12 names "a sensitive write, a removal, and a demotion". `reset_setting` on a sensitive key changes the
live value (C1), and the notice already treats it as a write (spec:152). Whether the gate covers it is left open.
The window is itself a store-homed setting. If its row fails to read, the spec does not say whether the gate
refuses or falls back to a default.

Fold. State that a reset of a sensitive key takes the gate. State that an unreadable window refuses the write,
following the seam's "never a default" rule (`seam.go:104-105`). Add both states to the planning-miss list.

### C8. minor. The notice's position relative to the commit is implied, not stated

Location: spec:150-155, AC13.

Defect. "A mail failure ... leaves the write committed" implies sending after the commit, but the spec does not
say so. A send inside the callback would hold the store's write lock across SMTP, and every other writer in both
processes would wait on the mail server. A crash between the commit and the send drops the notice silently.

Fold. The notice runs after `Update` returns nil, outside the callback, under its own deadline. A crash before the
send loses the notice, and the `settings_audit` row stays the durable record. State both.

### C9. minor (over-ceremony). The command-line channel's notice

Location: spec:152-155.

Defect. "On any channel" makes `settings reset` load the SMTP configuration and the roster, and send mail, even
with `serve` stopped. That adds failure paths to the recovery tool. The spec exempts the seed rule because only an
operator with box access can edit the environment (spec:154-155). The command line rests on the same proof
(spec:147-148).

Fold. The command-line channel sends no notice, for the same reason. Its audit row is the record. If C4's fold is
taken, the verb runs only against a current store, so this saves less, but it still drops mail wiring from a
recovery verb.

### C10. minor. A CLI removal or demotion leaves only a line on the operator's terminal

Location: spec:187-189.

Defect. "The log record the add path writes" goes to the verb's stderr (`member.go:96`), which is the operator's
terminal, not the journal. A web removal is journaled through `serve`, but a CLI removal or demotion leaves no
durable record of who acted. The tombstone row records when, not who. R6:909-910 rules out a member-audit table,
so the spec should state the consequence.

Fold. Accept this and state it, or have the runbook run member verbs so that stderr lands in the journal. Either
costs one line. A table is not proposed.

### C11. minor. Re-running the bootstrap step reactivates a removed admin with the admin flag

Location: spec:239-243.

Defect. The runbook step is `member add <DUBPLATE_ADMIN_EMAIL> --admin`. On a tombstoned address, `addUser`
reactivates the row with every field from the candidate, the flag included (`roster.go:415-417`). An operator who
re-runs the bootstrap after a restore to a rebuilt box re-grants an admin someone removed. That is the outcome D9's
stated reason exists to prevent.

Fold. The runbook step runs only when `member list` shows no admin. With C3's fold, any such re-add at least
revokes old sessions.

### C12. minor. The outcome of `reset_setting` and `settings reset` on an `env-seed` row or on no row is unstated

Location: spec:129-130, spec:306-308.

Defect. The states list names "no row" and "an `env-seed` row", but no rule gives their outcome. Deleting an
`env-seed` row would put the default live (C1) and write a misleading audit row.

Fold. Either case answers `invalid_request` with "nothing to reset" and writes nothing: no row, no audit, no
generation bump, and no notice.

### C13. minor. Overrides survive a removal and return on a re-add

Location: spec:185-186, the checklist (spec:222-229).

Defect. Removal deletes `member_preferences` rows only. `member_overrides` (`0005_routed.sql`) stays on the
tombstoned id, and a re-add under the old id revives it, including `may_upload` and `may_create_collections`.
Neither the spec nor CH:576-586 says whether this is intended.

Fold. State the rule. Keeping the rows is reasonable, because they are the admin's data and are restored with the
member. If they should not return, removal deletes them beside the preferences.

### C14. minor. A restore from a snapshot taken before a removal reinstates the member silently

Location: spec:191-209 (no restore rule).

Defect. A restore on the same box replaces the store and leaves the auth leaf alone. A member removed after the
snapshot was taken comes back live. Their old cookies stay refused by the surviving epoch, but a fresh magic link
signs them in. This is inherent to a restore, but nothing tells the operator it happened.

Fold. Add a runbook line now: after a restore, re-run any removals made since the snapshot. Record a cheap check
in BACKLOG for rung 16. That check would list each live member whose epoch record is later than the snapshot's
stamp. It needs no new state, and it is not proposed for 14a.

### C15. minor. `list_settings` reports a restart-applied key's new row as the value in force

Location: spec:112-113, spec:175-176, AC9.

Defect. After a write to one of the four rate-limit keys, the store row holds the new value, but the limiter
built at start still enforces the old one. "Value in force" then reports the wrong value until the next restart.

Fold. For a `restart` key, the read answers the running value and the pending stored value as separate fields.
Alternatively it labels the stored value "applies at the next start". The settings page shows whichever form is
chosen.

## Checked and sound

- Last-admin guard (spec:211-216). `Update` holds one `IMMEDIATE` transaction (`write.go:44-63`). The guard runs
  inside it after the write, and the actor's flag is re-read there. Two concurrent self-demotions in two processes
  therefore serialize, and the second is refused.
- Removal and bump ordering (spec:203-207). The bump runs after the tombstone and the guard and before the commit.
  A failed commit after the bump only signs out a live member. The bump is durable before any subscriber runs
  (`epoch.go:92-102`), and `CheckSession` reads the file on every verification (`seam.go:123`), so a CLI bump
  reaches a running `serve` at its next request. Cross-process serialization of the read, compare, and write in
  `advance` holds only because every 14a bump runs under the store write lock. C3's re-add bump keeps that
  property.
- Settings write atomicity. The row, the generation bump, and the `settings_audit` row share one transaction.
  `0004`'s schema carries every audit field the spec names. `0005` leaves the settings tables unchanged.
- The `NetworkActor` re-key. It has no production caller (`grep` at `9d730d7`), so no stored job carries the old
  form.

## Summary

There are fifteen findings: zero blockers, four major, and eleven minor. One major is an owner fork. The top
three:

- C1. A reset puts the registry default live instead of the next start's value, and skips the correlated rules,
  so a web reset can make the next start refuse.
- C2 (OWNER FORK; recommend refusing). A web write over an environment-seeded key commits with a warning and arms
  a start refusal that only box access can clear. Running C1's single seed-rule validator over every write closes
  C1 and C2 together.
- C3. Only the removal bumps the epoch. The record lives outside the backup and can be outrun between the bump and
  the commit, so a re-add after a rebuilt-box restore revives every pre-removal session. Bumping on reactivation
  closes it in a few lines.

C4 is the main over-ceremony item. The older-store write mode in the resolving verb duplicates the rollback
recovery R-6d2 already designed. It also needs a bypass of the store's version guard, and dropping it removes
the spec's riskiest new write path.
