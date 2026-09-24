# Rung 5c plan fold record (revision 1 to revision 2), 2026-09-23

Target: `docs/superpowers/plans/2026-09-23-rung-5c-pages-member-verb.md`,
revision 1 at `dba1679`, folded to revision 2 in the working tree. Inputs:
the three fresh-context lens reviews,
`docs/superpowers/research/2026-09-23-rung-5c-plan-review-contract.md`
(C1-C18), `-mechanics.md` (M1-M14), and `-risk.md` (R1-R12). Together
they report 0 blockers, 12 majors, and 32 minors: 44 findings.

## Method

Every finding was checked before it was acted on. The checks run for
this fold:

- `unicode.In(' ', unicode.Zs)` returns true and `unicode.In(0x200D,
  unicode.Cf)` returns true (scratch program, go1.27.1), which confirms
  M1 and M11.
- `strings.ToLower("a@Kexample.com")` returns `"a@kexample.com"`,
  and `net/mail.ParseAddress` accepts the unlowered form. That confirms
  C8's pin: the Kelvin-sign address is accepted.
- `src/encoding/json/v2/arshal.go` carries `//go:build
  goexperiment.jsonv2`, which confirms M4.
- `src/os/user/lookup_stubs.go` falls back to `os.Getenv("USER")` and
  builds from `os.Getuid()`, which confirms M6.
- `cmd/dubplate-server/doctor.go`'s `identitySourceStatus` calls
  `src.Validate` directly. So a keyed error type reaches `doctor` through
  `errors.As`, which settles the mechanism C12 and M8 left open.
- `internal/auth/roster.go`'s `CreateRoster` writes `"members":[]`. That
  confirms the convergent defect C1, M2, and R3 all describe.
- `internal/identity/identity.go`'s `Identity.Valid` accepts any
  `Channel.Valid()` value, which confirms R8.
- The lockout arithmetic in R1 checks out. Five link requests an hour at
  burst five refill a single charge in 12 minutes. Holding 10,000 buckets
  therefore costs about 14 requests a second, and at 20 per /64 per hour
  that is about 2,500 /64s.

## Convergent defects, fixed at the root

**A. The last-admin guard refused the bootstrap** (C1, M2, R3). Revision 1
routed `CreateRoster`'s empty write through the function that held the
no-admin guard. The fix is at the root: `CreateRoster(ctx, path, first
Member, actor, logger)` writes the first admin in one locked create, and
it refuses a non-admin first member. So the guard has no exemption, and
exit gate 11 holds as written, now including the bootstrap. Folded at
R-5c8, Task 6 (new interface, table, and mutations), and Task 7 (the
verb's pre-check collapses into the store's refusal). Task 1 gives
`CreateRoster` a `ctx`, per M2's note.

**B. Fail-closed eviction on all three limiter maps enabled a group
sign-in lockout** (C2, R1). The contract and risk lenses reached the same
design independently. The per-address key is charged only after the
roster lookup, and only for a known address. The per-address map never
evicts an unrefilled bucket: it admits a new key untracked instead, a
backstop that needs a roster larger than 10,000. The client-keyed maps
keep least-recently-used eviction. Under the generic answer, a known
address over its limit answers `requested` and sends nothing. This
changes the security posture Geoff ruled on, since he routed finding 3
"with its fix". So the design is **applied and listed as Ruling 1**, and a
"no" restores revision 1's design. Folded at R-5c3, Task 3 (rewritten),
exit gate 8, and conductor action 4.

**C. The reload's cached stat could come from a different file than its
content** (M3, R2). Every read now opens the file, takes
`(*os.File).Stat` on that descriptor, reads that descriptor, and caches
that stat with that content. A cache filled within one second of the
file's modification time counts as unverified and is re-read on the next
call (git's racily-clean rule). Folded at R-5c5, and at Task 1: its
outcomes, two table rows, the `os.Chtimes` cache-hit rows, a
reviewer-verified criterion for the rename race no test can force, and
mutations 9 and 10.

## Dispositions, one per finding

### Contract lens (C1-C18)

| ID | Sev | Disposition | Where |
| --- | --- | --- | --- |
| C1 | major | Folded (convergent A) | R-5c8, Tasks 6, 7 |
| C2 | major | **Owner fork, Ruling 1** (convergent B, applied to the recommendation) | Rulings, R-5c3, Task 3 |
| C3 | major | Folded: minor 6 disposed. The review's code fix is refused because a global 429 ceiling is a cheaper lockout than the signal it closes. The runbook sentence is carried to 5d | R-5c18, conductor-owed fold 11, Task 0 item 3, Task 10 |
| C4 | major | Folded (with R5): the refusal is the shell's own `http.NotFound`, compared against an unregistered path on status, headers except `Date`, and body; no `Set-Cookie`; the store-failure row added; 5c-ii tracks W3.7 | R-5c12, Tasks 8, 9, outline item 5 |
| C5 | major | Folded: the uid row now uses a `/tmp` fixture the test can write but does not own, and asserts the refusal itself | Task 7 |
| C6 | major | Folded (with R7's account-collision half): the `streamer_username` row records that it becomes a privilege channel at rung 12, and the constraint is carried to rung 12 | Privileged-fields table, Task 10 |
| C7 | minor | Folded: rows added for a malformed roster, an absent directory, the concurrent-bootstrap race, and a failed user lookup | Task 7 |
| C8 | minor | Folded: the Kelvin-sign row is pinned to acceptance, backed by the measurement | Task 2 |
| C9 | minor | Folded at the root: nothing is exported ahead of a cross-package caller. That rule unexports the context accessor, the audit-actor, no-admin, and busy errors, and every per-refusal error | R-5c20, Tasks 1, 2, 6, 8 |
| C10 | minor | Folded: the at-capacity policy is tested on the keyed map in isolation | Task 3 |
| C11 | minor | Folded: the `doctor` criterion is reworded ("prints the roster row exactly once, as any roster failure does") | Task 6 |
| C12 | minor | Folded (with M8): a keyed error type in `cmd/dubplate-server`, read through `errors.As`; relational refusals name both keys | R-5c10, Task 5 |
| C13 | minor | Folded: the errata list now names R3, R1's write path, "The roster", and `spec:1417` | R-5c6, Task 10 |
| C14 | minor | Folded: the outline gains W1.2/W1.4, W3.6, and W4.4/W4.6, the sign-out copy's bound, and the three page routes. The token-log assertion lands in exit gate 12 as a re-run of 5b-ii's test | Outline, exit gate 12, Task 0 item 13 |
| C15 | minor | **Owner fork, Ruling 6**: the reservation overrun is stated at the top, and the close owes section 15's line an erratum | Rulings, Token ceiling, Task 10 |
| C16 | minor | Folded: exit gate 6 is reworded to "two independent roster values, which `flock(2)` treats as two processes". The optional re-exec subprocess test is not taken, because the quoted `flock(2)` text carries the equivalence | Exit gate 6 |
| C17 | minor | Folded (with R11): a 5d runbook row names an invocation that keeps secrets out of argv. Narrowing the verb's load is declined, because the full load is what catches a broken environment before the first admin is written | R-5c14, Task 10 |
| C18 | minor | Folded (convergent C): the cache-hit fixture now uses `os.Chtimes` | Task 1 |

### Mechanics lens (M1-M14)

| ID | Sev | Disposition | Where |
| --- | --- | --- | --- |
| M1 | major | Folded: display names refuse `Zs` except U+0020; the address set is unchanged, since it already refuses space. The ZWJ/ZWNJ half became Ruling 3 | Task 2 |
| M2 | major | Folded (convergent A), plus `ctx` on `CreateRoster` | R-5c8, Tasks 1, 6 |
| M3 | major | Folded (convergent C) | R-5c5, Task 1 |
| M4 | minor | Folded: the experiment gate is recorded, and the null-member-value, invalid-UTF-8, and trailing-data behaviors are added to the table | Global Constraints, R-5c4, Task 1 |
| M5 | minor | Folded (with R6): named constants (`rosterAPIWriteWait` 5s, `memberVerbWriteWait` 30s); the `flock` is taken before the in-process mutex | R-5c19, Tasks 1, 7, 8 |
| M6 | minor | Folded: the actor comes from `user.LookupId(Geteuid)` with a `uid:<n>` fallback, never `$USER` | R-5c14, Task 7 |
| M7 | minor | Folded (with R12's half): the body is form-encoded through `readForm`, fields come from `PostForm`, an `admin` key anywhere in `r.Form` is refused, and responses are JSON | R-5c13, Task 8 |
| M8 | minor | Folded through C12's keyed type. M8's proposed prefix-cutting method is not used: the keyed type removes the convention dependence the method guarded against | R-5c10, Task 5 |
| M9 | minor | Folded: the ceiling is re-derived from 5b-ii's whole-pass 6.5M over twelve tasks (0.54M each), giving 7.40M and a 7.5M ceiling. The double-counted simplifier line is gone, and the Task 6 / Task 7 cut is stated as likely | Token ceiling, Execution model |
| M10 | minor | **Owner fork, Ruling 2** (merged with R7) | Rulings, R-5c6, Task 2 |
| M11 | minor | **Owner fork, Ruling 3** | Rulings, Task 2 |
| M12 | minor | Folded: the outline records the kit-source evidence, names the `reroute` hook as the first candidate to prove, and fixes the routes under `Shell.Register`'s no-slash rule | Outline, R-5c9 |
| M13 | minor | Folded: `adminContact` is added to Task 1's Files and criteria | Task 1 |
| M14 | minor | Folded: the doc comment must explain why a least-recently-used-only check would be wrong | Task 3 |

### Risk lens (R1-R12)

| ID | Sev | Disposition | Where |
| --- | --- | --- | --- |
| R1 | major | **Owner fork, Ruling 1** (convergent B), with the corrected arithmetic in R-5c3 | Rulings, R-5c3, Task 3 |
| R2 | major | Folded (convergent C) | R-5c5, Task 1 |
| R3 | major | Folded (convergent A) | R-5c8, Task 6 |
| R4 | minor | Folded: `AddMember` re-checks the actor's admin flag under the lock; `NewRosterAPI` exports only the gated composition, and the inner handler refuses on its own | R-5c8, Tasks 6, 8, privileged-fields table |
| R5 | minor | Folded (with C4): refresh happens only after the admin check | R-5c12, Tasks 8, 9 |
| R6 | minor | Folded (with M5): lock order, plus the API's own deadline so the `busy` 503 can be reached | R-5c19, Tasks 1, 8 |
| R7 | minor | **Owner fork, Ruling 2** for the character set (the alphanumeric first character is applied as the recommendation). The account-collision carry is folded with C6 | Rulings, R-5c6, privileged-fields table |
| R8 | minor | Folded: `Identity.Valid` refuses `ChannelCommandLine`, and the whole `Channel` comment is rewritten | R-5c7, Task 6 |
| R9 | minor | Folded: three residuals added (a malformed hand edit takes members offline, no promote path, a lockless hand edit inside a write window), and review focus 1 is corrected | Threat model, Review focus |
| R10 | minor | Folded: the pass-end security read now receives the routed findings as a "verify fixed" list, plus pointers to this round's majors | Conductor action 4 |
| R11 | minor | Folded (with C17) | R-5c14, Task 10 |
| R12 | minor | Folded: the header set covers every owned response in the outline; `login`'s registration is proved through the prefixed-link test; Task 4 tests both public-URL forms; the API reads `PostForm` | Outline, Task 4, Task 8 |

## Counts

- **44 findings.** 38 folded; 6 went to Geoff as owner forks (C2, C15,
  M10, M11, R1, R7); 0 refused outright.
- **Rulings: 6.** Rulings 1, 2, 3, and 6 come from those six findings
  after merging overlaps (C2 with R1, M10 with R7). Rulings 4 and 5 come
  from revision 1's conductor-owed folds 3 and 8, which the coordinator
  asked to consolidate as forks. The contract lens had judged both sound
  as applied.

**Why nothing is refused whole.** Every finding survived verification
against the tree, the Go source, or a measurement, and each fix is a
plan edit costing less than the fix round it prevents. Four sub-proposals
were declined, each with a one-line reason at its row:

- **C16's re-exec subprocess test.** The quoted `flock(2)` equivalence
  already proves two processes.
- **C17 and R11's narrowed verb load.** The full load catches a broken
  environment before the first admin is written.
- **M8's prefix-cutting method.** C12's keyed type supersedes it.
- **The 5b-ii review's code fix for minor 6, which C3 surfaced.** A
  global 429 ceiling is a cheaper lockout than the signal it closes.

## Errata owed to ratified documents

This revision changes the meaning of passages in ratified documents. The
fold edited none of them. Task 10 drafts each correction paste-ready.

- **Rung 5 spec, R3** (`spec:322-332`), **R1's write path**
  (`spec:291-292`), **"The roster"** (`spec:1106-1107`), and
  **`spec:1417`'s "an optional streamer username"**: the write path
  requires the field (Ruling 4). Uniqueness also becomes case-insensitive
  at load and at write time (R-5c6).
- **Rung 5 spec, `spec:1417`'s "an admin checkbox"**: dropped (Ruling 5).
- **Rung 5 spec, `spec:1033`**: the public URL must be a bare origin, and
  links carry the prefix (R-5c9).
- **Rung 5 spec, R23 and the 5c row** (`spec:569-575`, `spec:2331`): 5c
  splits into 5c-i at 7.5M and 5c-ii at 10.0M provisional, and the rung
  reservation stands near 36.4M against 15.75M (Ruling 6).
- **Rung 5 spec, R24**: the 5c review shape spans both halves.
- **Web surface standard, section 15**: "its 4.5M ceiling is unchanged by
  this standard" (Ruling 6).
- **5b-ii close (`draft-HISTORY-5b-ii.md`)**, once folded into
  `docs/HISTORY.md`: it routes minor 6 to 5c's fold, and this plan refuses
  the code fix (R-5c18).
- **5b-ii's fold 7**, the one-line-per-source ruling: amended to name keys
  on the line (R-5c10).

The 5b-ii security review's fix text for finding 3 is a research record,
not a ratified document. Ruling 1 records the departure from it, so no
erratum is owed there.

## Unresolved

- The six rulings await Geoff. Ruling 2 also needs the conductor's
  read-only Navidrome username query before Task 2.
- Plan preparation's 1.30M line in the ceiling is an estimate. The
  conductor holds the actual spend for drafting, the three lenses, and
  this fold.
- The box's filesystem timestamp granularity is unmeasured. The
  racily-clean rule covers both cases.

---

## Revision 2.1: the verification read folded (revision 2 to 2.1)

Input: `docs/superpowers/research/2026-09-23-rung-5c-plan-fold-verification.md`
(V1-V12: 0 blockers, 3 majors, 9 minors), read against revision 2 at
`562b9b2`. The coordinator's direction for this round: fold the three
majors, with the Task 7 / Task 8 cut as the default rather than a
fallback (a conductor ruling), and fold the minors where cheap.

### What was checked before acting

- **5b-ii's recorded figure.** `docs/HISTORY.md`'s top entry, "Both
  budgets", records "about **6.4M against the 7.5M ceiling**", not
  revision 2's 6.5M. The entry does not itemize the figure. The window's
  contents (Task 0 through the close, the conductor's turns, simplifiers,
  the security read) are inside it. Plan preparation (the 2026-09-12
  window) and the four architecture reads (recorded after the close) are
  outside it. So revision 2's 0.40M conductor line double-counted the
  conductor's turns, as V3 suspected.
- **Navidrome's username match.** Navidrome v0.63.2's
  `persistence/user_repository.go` builds `user.user_name = ? COLLATE
  NOCASE` for `FindByUsername`, and `model/user.go` comments "FindByUsername
  must be case-insensitive" (read 2026-09-23 through the GitHub API at tag
  `v0.63.2`). That is case-insensitive equality with no `LIKE`, which
  settles review finding 15's worry and retires revision 2's rung 12
  `LIKE` carry. The `user.user_name` column is confirmed by the v0.63.2
  migration `20200819111809_drop_email_unique_constraint.go`. The
  database path comes from `compose.yaml`'s `/srv/music/state/navidrome:/data`
  volume.
- **`r.Form`.** `go doc net/http.Request` says Form "contains the parsed
  form data, including both the URL field's query parameters and the
  PATCH, POST, or PUT form data".
- **`errors.Join`.** `go doc errors.Join` says "A non-nil error returned
  by Join implements the Unwrap() []error method." That confirms V10's
  point that collecting every key needs a walk, not one `errors.As`.
- **`sqlite3 --help`** documents `-readonly` as "open the database
  read-only".

### Dispositions, one per finding

| ID | Sev | Disposition | Where |
| --- | --- | --- | --- |
| V1 | major | Folded: `CreateRoster` routes through the locked read-modify-write function in an absent-file mode, so the first member and the actor pass the same write-path validation as every later write. Adds four invalid-create refusal rows, an actor row, mutation 2 ("skip the candidate validation in `CreateRoster`"), a verb row for `--streamer-username ""`, and exit gate 7's clause | R-5c8, Task 6, Task 7, review focus 5, exit gate 7 |
| V2 | major | Folded (conductor ruling, as directed): the cut at the Task 7 / Task 8 boundary is the default. 5c-i is Task 0, Tasks 1-7, and the close (renumbered Task 8). Revision 2's Tasks 8 and 9 head 5c-ii's outline in full, with the spec row's exit-gate assertions 1, 2, 3, and 5. The in-pass fallback moves to the Task 5 / Task 6 boundary, so the grant and bootstrap never ship without the verb (R-5c20) | The split, fold 1, Execution model, 5c-ii outline head |
| V3 | major | Folded: re-derived from the recorded 6.4M over twelve tasks (0.533M per task, conductor turns included). Planning is counted as spent at launch, outside the ceiling. The 0.40M conductor line is removed as a double count and replaced by 0.15M of architecture reads. Projection 4.72M; ceiling 6.5M; flag 5.2M, a task's worth above the projection. Re-segmented A (Tasks 1-3), B (Tasks 4-7), C (the close) | Token ceiling, fold 10, Execution model |
| V4 | minor | Folded: `errAdminGrantForbidden` is unexported. `RequireAdmin` and `NewRosterAPI` move with revision 2's Task 8 to 5c-ii, where the document route and the serve wiring in `cmd/dubplate-server` are their cross-package callers | R-5c7, R-5c20, 5c-ii Task 1 |
| V5 | minor | Folded (new conductor-owed fold 12): `ChannelHeader` and `ChannelServiceToken` dropped, with a `channelNames` table. Mutation 1 reworked onto `ChannelCookie`. `seam.go` is added to Task 6's Files for the `VerifiedIdentity.Admin` comment, and `Identity.Valid`'s comment is corrected. `spec:1223` erratum added | R-5c7, Task 6, Task 8 corrections |
| V6 | minor | Folded: R-5c7 is reworded ("fails `Identity.Valid`, which 5c-ii's gate requires"); mutation 9 ("skip `Identity.Valid` in the gate") moves with the gate | R-5c7, 5c-ii Task 1 |
| V7 | minor | Folded: the file is aged to T0 before `OpenRoster` and restored to T0; the residual race is renamed "an old stat over new content"; the reviewer criterion now permits the change check's path stat | R-5c5, Task 1 |
| V8 | minor | Folded: one gating line at all three sites ("Ruling 6 gates the launch; Rulings 2, 3, and 4 gate Task 2; Ruling 1 gates Task 3; and Ruling 5 gates Task 6"); Task 6's header names Ruling 5 | Rulings intro, Owner touchpoints, Execution model, Task 6 |
| V9 | minor | Folded. Ruling 2 is narrowed to "may the conductor run one read-only query", and its "no" applies the shape unchecked with a recorded residual. Ruling 5's "no" names the tasks, gate, and table rows it re-writes. Ruling 6's "no" now moves the add-member surface to rung 7, because V9's suggested "no" (the Task 7 / Task 8 cut) became the default | Rulings 2, 5, 6; R-5c6 |
| V10 | minor | Folded. The racily-clean rule is restated as the plan's own one-second rule with no git citation. `systemd-run` is handed to 5d as a claim to prove. `r.Form` is quoted. The Navidrome query is given exactly: `sudo sqlite3 -readonly /srv/music/state/navidrome/navidrome.db "SELECT user_name FROM user;"`, with two named pre-checks and a fallback to Ruling 2's "no". Task 5 now walks the join, with `errors.Join` quoted | R-5c5, R-5c10, R-5c13, R-5c14, R-5c6, Task 5, conductor action 1 |
| V11 | minor | Folded: 5c-ii Task 1's Files add `internal/auth/handlers.go` (the `readForm` lift); Task 4's Files add `cmd/dubplate-server/doctor_test.go` | 5c-ii Task 1, Task 4 |
| V12 | minor | Folded: a distinct unexported actor-not-admin error, plus mutation 6 ("return the no-admin error for a non-admin actor") | Task 6 |

### Counts and refusals

There are 12 findings, all folded, with no owner forks and no refusals.
Every finding held against the tree, the Go source, or the recorded
figure, and each fix was a plan edit. Two proposals were taken in a
different form than proposed, each with a one-line reason:

- **V3's alternative of folding the cut into Ruling 6** was not taken,
  because the coordinator ruled the cut the default.
- **V4's option of keeping `RequireAdmin` unexported in 5c-i** is moot,
  because the gate moved to 5c-ii with its callers.

### Errata owed, added this round

- **Rung 5 spec, `spec:1223`.** Its three assertion channels become two
  constants ("constants land with the source that asserts them"),
  fold 12.
- **Rung 5 spec, the 5c row (`spec:2331`).** Its exit-gate assertions 1,
  2, 3, and 5 move to 5c-ii with the roster API; the 5c-i/5c-ii split
  erratum already listed carries this.
- **The rung 12 carry list.** The `LIKE` question is closed on Navidrome's
  source; the account-collision constraint stays.

### Unresolved after revision 2.1

- **The six rulings.** They await Geoff.
- **Planning spend.** The actual spend for drafting, the three lenses,
  the fold, the verification, this fold, and the prose review is the
  conductor's to record. The rung arithmetic estimates it at 1.5M.
- **The Navidrome query's two pre-checks.** Whether `sqlite3` is
  installed on the box, and whether `-readonly` opens the live WAL-mode
  database while Navidrome holds it, are both unverified. Either failure
  takes Ruling 2's "no" branch.
- **The architecture-read estimate.** The 0.15M line has no recorded
  basis in 5b-ii.
- **5c-ii's size.** It now carries thirteen tasks plus Task 0, past the
  twelve-task line, so its own plan should propose a split.
