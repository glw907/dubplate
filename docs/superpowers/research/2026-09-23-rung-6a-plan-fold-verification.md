# Rung 6a plan fold verification

This is a fresh-context verification of the review fold of the rung 6a plan. The target is
`docs/superpowers/plans/2026-09-23-rung-6a-store-foundation.md` at revision 2 (`1d69ab8`). The
fold record is `docs/superpowers/research/2026-09-23-rung-6a-plan-fold-record.md`. The three
reviews and the ratified rung 6 spec (revision 2.2) were read as sources. The verifier had no part
in the fold. Line numbers are revision 2 plan lines unless marked otherwise.

Evidence was re-run on 2026-09-23 with `modernc.org/sqlite v1.59.0` and go1.27.1. The fold probe
binary (`6a-fold-probes/probe`) was re-run, and P1 to P9 all reproduce. One new probe,
`6a-verify-probes/main.go`, re-measured R-6a18 on a warm pool, and the verifier read the quoted
driver sources directly.

**Counts: 0 blockers, 2 majors, 7 minors.** **Verdict: needs a narrow second fold.** Both majors
are one-paragraph edits with no new task, no new deliverable, and no ceiling change.

---

## Question 1: did each blocker and major close at its cited location?

Yes, all of them. Every fix appears where the fold record says it does, and the probe each one
rests on reproduces.

| ID | Closed at | Evidence re-checked |
| --- | --- | --- |
| D1 (blocker) | R-6a10 `:561-584` (the error latch); Task 3 `:1168-1171`, states `:1210-1213`, autocommit criterion `:1225-1233`, mutation 8 `:1269-1270`; the runner's first-error rule in Task 4 `:1308-1317` | P1 re-run: `rows: [1 3]`, `Rollback` answers "no transaction is active" |
| C1, M6 | Task 4 criterion 8, three rows `:1401-1417`; mutation 1 names its row | P8 re-run: `constraint failed: injected (1811)`, `uv 1 objects 1 ledger 1` |
| C2 | Task 6 criterion 15 hot-WAL fixture `:1776-1789`; mutation 6 `:1814-1815` | P4 re-run. The non-`mode=ro` arm keeps `_query_only=1`, the same DSN mutation 6 names, and it changes the hash. |
| C3, M2 | R-6a5 `:449-451`; R-6a6 `:474-496`; Task 6 `:1735-1742` (the header-only fixture is asserted at 4096 bytes) | Mechanics probe D and D2 source read (`6a-review-probes/main.go:110-138`) |
| C4 | Task 1 idempotent create `:829-839` and `:909-910`; R-6a5 `:446-448`; Task 6 listing-unchanged rows `:1730-1734` | Not probe-dependent |
| C5, M4 | Task 6 criterion 14 `:1747-1775` (logger signal, placed store at current, `os.SameFile`, other connections closed) | P6 and P7 re-run (0 of 2000 failed); the VFS deferred-close quote is verified at `lib/sqlite_g_000000000001ffff.go:12569` |
| C6 | Task 5 criterion 12 `:1572-1580`; the 787 parents `:1586-1590` | P5 re-run, 2067, `users.address` and `users.streamer_username` |
| M1 | R-6a17 `:669-691`; Task 2 `:1095-1098`; mutation 6 `:1123-1124` | P2 re-run; see V6 for a misreport of its output |
| M3, D4 | R-6a7 `:507-520`; Task 6 `:1692`, `:1713-1714`, `:1793-1796`; mutations 4 and 9 | Not probe-dependent |
| M5 | R-6a18 `:693-708`; Task 3 `:1202`, `:1234-1237`; Ruling 4 | Re-measured; see question 3 |
| M7 | R-6a3 `:386-400`; Task 7 `:1866-1874`, `:1895`, `:1912-1916`; mutations 5 and 6 | `sqlite.go:307-311` `ParseInt(..., 10, 64)` verified |
| D2 | As C3 and C4, plus the re-stat under the exclusive lock `:833-835` | As above |
| D3 | R-6a5 `:430-437`, `:455`; Task 6 `:1743-1746`; mutation 7 | Not re-probed; the reviewer measured it |
| D5 | Ruling 3 `:85-99`; R-6a5 `:447`; Task 6 `:1797-1800`; mutation 10 | Closed as an owner fork. V2 finds a leak in the "yes" answer. |

## Question 2: did the fold introduce a contradiction or an unbuildable order?

It introduced one real ambiguity (V1) and one small forward dependency (V4). Every other rule stays
consistent across R-6a5, R-6a7, the privileged-fields table, Ruling 2, and Tasks 1, 6, and 7. The
task order stays buildable. The segment boundaries still sit on green commits.

## Question 3: did the fold state any new mechanism from memory?

No load-bearing mechanism rests on memory.

- **R-6a17.** P2 and P3 prove it. The builder produces `file:///…/a%23b%2541%3Fc/s.db?mode=rw…`,
  `busy_timeout` reads back 500, and `mode=rw` on an absent file returns code 14 and creates
  nothing. V6 covers two sentences that misreport P2's listing.
- **R-6a18.** Mechanics probe A (`6a-review-probes/main.go:49-72`) proves it. The verifier re-ran
  it on a warm pool in the plan's Task 3 shape: a 100 ms deadline, `_busy_timeout=400`, and a
  second handle holding `IMMEDIATE`. Five of five runs returned `database is locked (5)` at
  402 ms, with `ctx.Err()` already `deadline exceeded`. The quoted source is real:
  `interruptOnDone` is at `sqlite.go:74`, and `tx.exec` wraps `begin` in it at `tx.go:71`. The
  test comment is at `all_test.go:4593`. One sentence is an inference, not a measurement: "a
  write pool's one connection is replaced after an interrupt" (`:701-703`). Nothing depends on
  it. The next `Exec` on the probe's pool succeeded either way.
- **The D1 error latch.** P1 proves its premise, autocommit after an automatic rollback. The latch
  itself is a design rule, and mutation 8 tests it.
- **The interrupted-create classification.** Probes D and D2 prove the header-only state, and the
  source read confirms their shape: a pre-created file, the write DSN, then begin, rollback, and
  close. The risk review's child probe proves the exit-without-close half.
- **Criterion 8** rests on P8. **Criterion 14** rests on P6 and P7, plus the `flock(2)` and VFS
  quotes. **Criterion 15** rests on P4. All five probes re-ran.
- **Other quotes re-verified:** the `VACUUM INTO` pager-flags comment
  (`lib/sqlite_g_0000000000000003.go:77138`), the `dev_t` masks in `sysmacros.h`, and the
  `magic.h` constants. V7 names a constant the fold missed.

## Question 4: can Geoff rule each ruling on the evidence given?

- **Ruling 1: yes.** The arithmetic checks: 8 × 0.533 = 4.27, plus 0.05, 0.20, 0.10, and 0.30,
  makes 4.92, and 80% of 6.8 is 5.44. V8 is a 0.01M rounding in the split's headroom claim.
- **Ruling 2: yes.** The two answers' effects are stated, and 5d's exit gate is named.
- **Ruling 3: not yet.** The "yes" answer claims to close the silent empty-store path, but a
  zero-byte or header-only file beside a snapshot copy still reaches it (V2). Geoff should rule on
  the closed version.
- **Ruling 4: yes.** The measurement is real and re-confirmed. V8 notes that the "no" answer's
  split cost is not in Ruling 1's figures.

---

## Findings

### V1 (major): a newer store matches two report rows, and the rebuilt proofs hit the wrong one

**Location:** plan `:458` and `:461` (R-6a5); Task 4 `:1379-1381`; Task 6 criterion 13
`:1724-1725`; criterion 15 `:1776-1786`.

**Defect.** The fold added a ledger-completeness rule: "a version at or below `user_version` with
no ledger row" is refused (`:461`, `:1380`). Two newer-store fixtures meet that rule and the
"newer" row (`:458`) at once:

- Criterion 13's newer store has `user_version` set past the head by a raw connection. Its head+1
  version has no ledger row.
- Criterion 15's hot-WAL newer row is "the version change" written by the child.

A real newer store, the spec's "newer snapshot restored under an older binary" (spec `:374-376`),
carries ledger rows for versions the binary has no file for. Those rows meet Task 4's
revision 1 row "a row for a version with no binary file → refused" (`:1379`). Neither the table
nor the runner says which check runs first. An implementer who applies the ledger check first
reports "refused, naming version N" where the spec and criterion 15 expect "newer, naming both
versions". The test then fails, or it gets rewritten to the implementation. The older hot-WAL row
has the same problem: a child that lowers `user_version` leaves a ledger row above it, which is
"refused".

**Proposed fold.**

1. State the precedence once in R-6a5 and Task 4: "newer" is decided on `user_version` alone,
   before the ledger is read, because the binary cannot judge versions above its head. The ledger
   checks run only when `user_version` is at or below the head.
2. Name the two fixtures' writes:
   - Older: the child applies a real pending migration through the runner on a sub-`fs.FS`.
   - Newer: the child sets `user_version` to head+1 and inserts that version's ledger row, as a
     newer binary would.

### V2 (major): Ruling 3's "yes" still lets serve replace a lost catalog through the interrupted-create path

**Location:** Ruling 3 `:85-99`; R-6a5 `:447`, `:449`, `:451`; R-6a6 `:474-493`; Task 6
`:1797-1800`.

**Defect.** Ruling 3's "yes" guards only the case where no database file exists. R-6a6 lets serve
complete any zero-byte file with no `-wal`, or any header-only unmarked file, with no condition.
Filesystem repair, a failed copy, or `: > store.db` commonly leaves a zero-length file. A lost
database of that shape beside `pre-migrate-v*.db` copies becomes an empty store without a word.
That is D5's hazard through a second door. After 6b-i, the next backup uploads the empty store,
which is the consequence Ruling 3 cites. The ruling text tells Geoff the path is closed, so he
would rule on an incomplete guarantee.

**Proposed fold.** Under "yes", the same `snapshots/` read also guards completion. A genuine
interrupted create never has a copy beside it, because a store at version 0 takes no snapshot
(R-6a11 `:606-607`). Add that clause to Ruling 3's "what yes builds" text and to R-6a5's zero-byte
and header-only rows. Extend Task 6's Ruling 3 criterion with a zero-byte database beside
`snapshots/pre-migrate-v1.db`: serve refuses and inspect reports "refused".

### V3 (minor): an unmarked file with `user_version` above 0 and no schema object has no row

**Location:** R-6a5 `:451`, `:454`; criterion 13's foreign fixture `:1726-1727`.

**Defect.** Row `:451` needs `user_version` 0, and row `:454` needs a schema object. A file with
`application_id` 0, `user_version` above 0, and no schema object matches neither row. That is C3's
own suggested fixture ("`user_version` set but no `application_id`") if the fixture is built
without a table.

**Proposed fold.** Change row `:454` to read "`application_id` 0 with any schema object or a
nonzero `user_version`: not a store".

### V4 (minor): Task 4 uses the inspect DSN that Task 2 assigns to Task 6

**Location:** Task 2 DSN table `:1023` ("inspect (Task 6)"); Task 4 `:1330-1331`; R-6a17 `:691`.

**Defect.** The fold moved the snapshot's verification open onto the inspect DSN, which lands two
tasks later. The order is buildable, since a DSN is one string, but it leaves Task 4's implementer
to write another task's surface.

**Proposed fold.** Land the inspect class in Task 2's builder, and include it in criterion 4's
readback.

### V5 (minor): criterion 14 hangs on a log record whose timing is unstated

**Location:** Task 6 `:1689-1690` ("logging holders at an interval"), `:1755-1757`; R-6a14
`:636-647`; the Global Constraints universality line `:281-283`.

**Defect.** Both of criterion 14's rows now order on "the holders record". Neither the plan nor
R-6a14 says when the first record is emitted or what the interval's value is. If the first record
waits one interval, each row takes that long under `-race`. An unstated interval also breaks the
global rule that every constant states its value.

**Proposed fold.** State in R-6a14 that the first holders record is emitted on the first failed
poll. Name the interval constant with its value.

### V6 (minor): R-6a17 misreports P2's output

**Location:** `:681-682`, `:687-688`.

**Defect.** P2's re-run prints `base: [a#b%41?c d1.db] inner: [s.db]`. `d1.db` is P1's file, so
"the only entry beside the directory was the directory itself" is false as worded. After the
close, no `-wal` is left beside the file, so "creating `-wal` beside it" is not what P2 shows. The
conclusion stands: the file lands at the exact path, and nothing appears at a truncated path.

**Proposed fold.** Restate the result as "no entry at a truncated path (`a`), and the database at
the exact path in WAL mode".

### V7 (minor): the filesystem probe misses the in-kernel AFS magic

**Location:** Task 1 `:894-896`.

**Defect.** `magic.h` carries two AFS constants: `AFS_SUPER_MAGIC 0x5346414F` (line 7) and
`AFS_FS_MAGIC 0x6B414653` (line 56). The fold took only the first. The in-kernel kAFS client
reports the second; this comes from the verifier's knowledge of `fs/afs/super.c` and was not
probed.

**Proposed fold.** Refuse `AFS_FS_MAGIC` too, or name it as a declared gap. Either costs one line.

### V8 (minor): the split-cost headroom claims are off by a rounding, and Ruling 4's "no" is unpriced

**Location:** `:218-219`; `:1833-1835`; `:2114-2115`; Ruling 4 `:110-114`.

**Defect.** The flag's headroom is 0.52M, and the pre-named Task 6 split is "about 0.53M, which
the flag's headroom covers". After that split, the projection is 5.45M, just past the 5.44M flag.
Ruling 4's "no" also splits Task 3 for about the same cost, which appears in neither ruling.

**Proposed fold.** Say that either split brings the projection to the flag, and that the
conductor asks at that boundary under the 80% rule. Or add about 0.5M to Ruling 4's "no" text.

### V9 (minor): serve's completion of a present but partial leaf has no step in the order

**Location:** R-6a5 `:446` ("the leaf completed and then created"), `:448`; Task 6 order
`:1658-1659`.

**Defect.** Task 6's order runs the idempotent leaf create only when the leaf is absent. R-6a5
promises that serve completes a present leaf that lacks `store.lock`, `db/`, `snapshots/`, or
`staging/`. No step does that.

**Proposed fold.** Serve runs Task 1's idempotent create whenever any leaf name is missing. It runs
after the owner check and under the state-directory owner rule.

---

## Not raised

These were checked and need no fold, or were not introduced by the fold.

- R-6a7 says the intent check runs "before any other step", while Task 6's order runs the owner
  check and the filesystem probe first. Revision 1 had the same wording, and the fold did not
  introduce it.
- The fold record summarizes P5's message without the `constraint failed:` prefix, but the plan
  quotes it exactly.
- The finding counts reconcile: 13 + 16 + 17 = 46, which is 38 folded, 4 folded in part, 1
  refused, and 3 owner forks.
