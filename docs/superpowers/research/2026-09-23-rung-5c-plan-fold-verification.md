# Rung 5c plan fold verification (revision 2), 2026-09-23

Target: `docs/superpowers/plans/2026-09-23-rung-5c-pages-member-verb.md`
at `562b9b2` (revision 2), read against the fold record
(`2026-09-23-rung-5c-plan-fold-record.md`), the three lens reviews, and
the tree at HEAD. Fresh context; no part in the fold.

## Verdict

**Not ready for the approval sitting yet. It needs one narrow second fold
on three majors, each a plan edit of a few lines.** The fold is
substantially sound. All twelve majors from the lenses closed at their
cited locations. The three convergent defects no longer occur as
described: `CreateRoster` writes the first admin in one locked create;
per-address charging happens only after the lookup, for known addresses;
and every read fstats the descriptor it reads. The quoted mechanisms I
spot-checked match `go doc` and the Go source: `TokensAt`, the
`lookup_stubs.go` fallback, `url.Error`'s `%q` format,
`CrossOriginProtection`'s header-less allowance and `originCheck`'s
tightening, `Shell.Register`'s two patterns, and the terminator's
`http.NotFound`. The three majors are these. The bootstrap write skips
the write-path validation that Ruling 4 promises on both write paths.
The fallback cut at Task 6 / Task 7 breaks the plan's own new rule
R-5c20. And the ceiling's 0.1M margin, under a 6.0M flag, makes a
mid-pass owner question close to certain unless the cut is decided at
launch. Geoff can rule all six rulings on the evidence given. Two of
the "no" branches (Rulings 5 and 6) do not yet name a buildable edit.

## Findings

### 1. MAJOR: `CreateRoster`'s first member bypasses the write-path validation, so the bootstrap admin can lack a streamer username

**Location:** R-5c8 (plan:698-707); Task 6 outcomes and state table
(plan:1785-1789, :1806-1810); Task 7 outcomes (plan:1896-1906).

**Defect:** Task 6 specifies that `CreateRoster(ctx, path, first, actor,
logger)` "requires `first.Admin`, audits on `ChannelCommandLine`,
re-checks the file's absence under the lock". It does not say that
`first` passes Task 2's write-path rules: the required, shaped
`streamer_username`, the display-name set, and the address set. Nor
does it say that `actor` passes the command-line actor rule. Those rules
live in `AddMember` (Task 2, plan:1339-1354), and R-5c8 routes the
invariant through "the one internal read-modify-write function", which
the create does not use, since there is no disk copy. Cobra's required
flag checks presence, not content, so `member add --admin a@x --name A
--streamer-username ""` reaches `CreateRoster` with an empty username.
The loader keeps the field optional (R-5c6), so the file loads. The most
important member on the roster then has no username, which is what
Ruling 4's "yes" promises cannot happen ("both write paths refusing a
missing username by name") and what rung 6's import refuses. No table
row and no mutation covers an invalid first member.

**Evidence:** `internal/auth/roster.go:147-166` (today's `CreateRoster`
runs no member validation, since it writes none). Task 6's create rows
test only admin versus non-admin.

**Proposed fix:** have `CreateRoster` validate `first` and `actor` with
the same write-path candidate check `AddMember` runs. The cleanest form
routes the create through the locked read-modify-write function, with an
"absent is allowed" mode, which makes "no exemption" true by
construction. Add Task 6 rows for a create with a missing, empty, or
misshapen streamer username, and with an invalid display name, each
refused with nothing created. Add the mutation "skip the candidate check
in `CreateRoster` → those rows create a file".

### 2. MAJOR: the fallback cut at Task 6 / Task 7 contradicts R-5c20

**Location:** Token ceiling (plan:354-360); Execution model (plan:2430-2435);
R-5c20 (plan:933-944).

**Defect:** the plan calls the Task 6 / Task 7 cut "likely" and says
"nothing argues against" it, because Tasks 1-6 ship "a grant whose
callers exist in tests". R-5c20, which this revision adds, exports a
symbol only when a caller outside its package calls it or branches on
it, and it names `CreateRoster` as exported because the verb calls it.
If the cut lands after Task 6, `CreateRoster`, the command-line grant
path through `AddMember`, and `ErrRosterExists`'s production reader all
ship with test-only callers. That is the exact shape Geoff's 2026-09-12
refinement forbids, and the staged exported-unused leg flags it. The
split section's own argument ("5c-i ships its own callers",
plan:222-226) holds only if the verb stays.

**Evidence:** `grep -rn CreateRoster` shows only `roster.go` and
`roster_test.go` today. The verb (Task 7) is its first production
caller.

**Proposed fix:** move the fallback cut to the Task 7 / Task 8 boundary.
Tasks 1-7 keep the verb, so every export has its caller. Tasks 8-9 (the
gate, the API, and the wiring) and exit gates 1, 2, 3, and 5 open 5c-ii.
Task 7 depends only on Task 6, so the order still builds. Restate the
sentence at both sites.

### 3. MAJOR: a 0.1M margin under a 6.0M flag makes a mid-pass owner question near-certain, so decide the cut at launch

**Location:** Token ceiling (plan:340-360); Execution model (plan:2430-2441).

**Defect:** the arithmetic is correct as stated: 5.40 + 0.10 + 0.20 +
1.30 + 0.40 = 7.40M. But its behavior is not what the plan says.

- **The flag fires in segment C.** 5b-ii closed at about 6.4M and
  crossed its 6.0M flag, which produced an owner question
  (`docs/HISTORY.md:657-661`). 5c-i projects 1.0M higher, with the
  1.30M preparation line already spent at launch. The 6.0M flag is then
  crossed around Task 8. Under the workstation rule, that is "finish
  the task, write STATUS, and ask one combined question", an execution
  sitting the plan scores as zero.
- **The re-derivation comes too late.** It is scheduled at segment A.
  Preparation is the one line the conductor already holds as an actual,
  and it includes drafting a 2,569-line plan, three lenses, this fold,
  this verification, a prose review, and possibly a second fold. If
  that exceeds 1.4M, the projection passes 7.5M before Task 1 runs.
- **The basis is loose.** The basis is 6.5M, while the 5b-ii close
  records about 6.4M (`draft-HISTORY-5b-ii.md:657`). The 0.40M
  conductor line may double-count if 5b-ii's whole-pass figure already
  included conductor turns and its architecture reads. The plan's list
  of what the 6.5M covers leaves both unstated.

**Proposed fix:** replace line 3 with the preparation actual at launch,
and decide the cut there, not at segment A. With the margin this thin,
make the Task 7 / Task 8 cut (finding 2) the default, since it
re-derives to about 6.3M with the flag above the projection. Keep one
pass only if the launch projection clears the flag by a task's worth. Or
fold the choice into Ruling 6: "yes" also raises 5c-i to a ceiling whose
80% flag sits above 7.4M. State whether the 6.4M figure includes
conductor turns, preparation, and architecture reads.

### 4. MINOR: `RequireAdmin` and `ErrAdminGrantForbidden` are exported with no cross-package caller in this pass

**Location:** R-5c20 (plan:940-944); Task 8 Produces (plan:2028-2039);
Task 9 (plan:2160-2165).

**Defect:** Task 9 mounts only `NewRosterAPI`, which is "already gated",
so `RequireAdmin`'s only callers in 5c-i are `NewRosterAPI` and the
tests, all inside `internal/auth`. Its cross-package caller is 5c-ii's
document route. The verb "branches on none" and prints error text, so
`ErrAdminGrantForbidden` has no reader outside the package either. Being
"a rename of an existing export" is not a caller under R-5c20's own
wording.

**Proposed fix:** keep `requireAdmin` and the grant error unexported in
5c-i. 5c-ii exports the gate when its document route mounts it. The
handoff in Task 10 names the gate either way.

### 5. MINOR: the channel set stays incoherent after `ChannelCommandLine` (the architecture-read notes)

**Location:** R-5c7 (plan:658-683); Task 6 Files and mutation 1
(plan:1754-1759, :1846).

**Defect:** Task 6 settles half of the architecture read's point about
`identity.go:68-69`: it rewrites the `Channel` comment. It leaves the
following:

- **The same policy statement in `internal/auth/seam.go:52-55`**
  ("An admin verb rejects an identity whose Channel is ChannelCookie;
  the one exception is the add-member handler"). After 5c-i, no admin
  verb takes an `Identity` at all. `seam.go` is in no task's Files.
- **`ChannelHeader` and `ChannelServiceToken`**, which have no producer.
  Their only references are `roster_test.go` and the `String` and
  `Valid` switches. The 5b-ii read recommends dropping them until their
  sources land. Revision 2 keeps both, and Task 6's mutation 1 depends
  on `ChannelServiceToken`. That leaves four constants: one that a
  verifier emits, two that nothing emits, and one that is not an
  assertion at all, under a type documented as "how a verified identity
  was asserted".
- **`Identity.Valid`'s doc comment**, which still promises "a normalized
  address" while the code checks only for a non-empty value
  (`identity.go:129-133`). Task 6 edits that method.

**Proposed fix (conductor fold, not a ruling):** drop `ChannelHeader`
and `ChannelServiceToken` in Task 6. `identity.go` is already in its
Files. Move the roster tests to `ChannelCookie`, and adopt the read's
`channelNames` table form. Mutation 1 becomes "admit the flag on
`ChannelCookie` → the refusal rows write". Add `seam.go` to Task 6's
Files, so the `Admin` field comment states only that the flag comes from
the roster. Fix the `Identity.Valid` doc comment in the same edit.
`spec:1223` lists the three channels descriptively; a one-line erratum
("constants land with the source that asserts them") covers the drop.

### 6. MINOR: R-5c7's "no verifier can emit it" is a structural claim, and no mutation proves the gate calls `Identity.Valid`

**Location:** R-5c7 (plan:674-681); Task 8 outcomes and mutations
(plan:2043-2045, :2121-2135).

**Defect:** `Identity.Valid` refusing `ChannelCommandLine` protects the
web edge only where a caller runs it. The 5b-ii read found that no
production code calls it today (the seam builds an `Identity` and never
validates it). The gate's "require `Identity.Valid()`" is the one new
caller, and no named mutation drops it.

**Proposed fix:** reword R-5c7 to "an identity carrying it fails
`Identity.Valid`, which `RequireAdmin` requires". Add a Task 8 mutation:
"skip `Identity.Valid` in the gate → a test identity on
`ChannelCommandLine` with `Admin` reaches the inner handler". The test
injects that identity through a stub `Seam`.

### 7. MINOR: the racily-clean cache-hit fixture cannot pass as literally written, and the residual race is misnamed

**Location:** Task 1 acceptance criteria (plan:1216-1227); R-5c5
(plan:602-604).

**Defect:**

- **The fixture.** The file is freshly written, so its modification
  time falls inside the one-second window when the cache fills, and the
  cache is unverified from the start. "Restore its modification time …
  to a moment more than one second before the cache was filled" leaves
  a modification time that differs from the cached one. The stat check
  then reloads, and the "old answer persists" row fails. It passes only
  if the file was aged with `os.Chtimes` before the open and is
  restored to that same value.
- **The race's name.** With fstat before the read, the residual race is
  an in-place write between the two, which gives an old stat over new
  content. That is harmless, and it costs one reload, as the plan says.
  The plan calls it "a new stat over old content", which is the
  dangerous direction.
- **The reviewer criterion.** "No path calls `os.Stat(path)` … for a
  cached read" can be read as banning the path stat that the change
  check needs.

**Proposed fix:** "`os.Chtimes` the file to T0, more than one second in
the past, before `OpenRoster`. Rewrite it in place, then restore T0."
Rename the race "an old stat over new content". Reword the criterion:
"the cached stat is always the fstat of the descriptor that was read;
the change check may stat the path".

### 8. MINOR: the gating lists omit Ruling 4, and Ruling 5 actually gates Task 6

**Location:** plan:36-38, :305-306, :2425-2428.

**Defect:** Task 2 depends on "Rulings 2, 3, and 4" (plan:1300-1301),
and Ruling 4 also shapes Tasks 7 and 8, but none of the three gating
lists names Ruling 4. A "no" on Ruling 5 reverses Task 6's rule that
only the command line may set the flag (R-5c13, plan:823-825). So Ruling
5 gates Task 6, not only Task 8.

**Proposed fix:** "Rulings 2, 3, and 4 before Task 2; Ruling 1 before
Task 3; Ruling 5 before Task 6". Use the same line at all three sites.

### 9. MINOR: three rulings' "no" branches are not clean buildable edits

**Location:** Rulings 2, 5, and 6 (plan:57-72, :95-106, :108-118).

**Defect:**

- **Ruling 6.** "No means naming a cut" hands the design back to Geoff.
  It offers an example it "does not propose".
- **Ruling 5.** A "no" reopens Task 6's grant rule, exit gate 11, the
  privileged-fields table, and threat-model claim 1. The branch names
  the consequence, but not the tasks and gates it would rewrite.
- **Ruling 2.** It bundles one genuine owner question with two
  mechanics: the read-only production query, which needs Geoff under
  CLAUDE.md's box rule, and the first-character rule and load-versus-
  write scope. Its "no" builds `[A-Za-z0-9._-]` at load, the variant
  both the mechanics and risk lenses flagged as stranding records.

**Proposed fix:**

- **Ruling 6:** name the default "no": the Task 7 / Task 8 cut (finding
  2), with 5c-i held near 6.3M and the difference moved to 5c-ii.
- **Ruling 5:** "no" re-plans Tasks 6 and 8 and exit gate 11 before
  Task 6.
- **Ruling 2:** narrow the question to "may the conductor run one
  read-only query of the box's Navidrome usernames?" Its "no" applies
  the recommended write-only shape without the check, and records the
  unverified fit as a residual.

The other three rulings (1, 3, 4) are genuine owner choices. Their
evidence is sufficient, and their "no" branches are buildable.

### 10. MINOR: four new mechanisms are stated from memory

**Location:** R-5c5 (plan:606-612); R-5c14 (plan:872-874); R-5c13
(plan:832-836); R-5c6 (plan:652-656); R-5c10 (plan:761-764).

**Defect:**

- **Git's racily-clean rule** is named but not quoted. The plan's
  one-second window is its own rule and needs no git citation, or it
  quotes `Documentation/technical/racy-git.txt`.
- **`systemd-run --uid=music -p EnvironmentFile=...`** is an unquoted
  example. Whether a transient unit accepts that property, and whether
  it needs `--pipe --wait` to return output, is unverified. 5d should
  inherit it as a claim to prove, not a recipe.
- **`r.Form` holding both the query and the body** is relied on and not
  quoted. `go doc net/http.Request.Form` carries the sentence.
- **The Navidrome username query** has no stated mechanism.
  `sqlite3 -readonly` against a database in WAL mode, and an API call
  that needs admin credentials, behave differently, and "read-only" has
  to be true of the one that is chosen.

A fifth item is a misapplied quote rather than a recalled mechanism.
Task 5's "collects keys with `errors.As` over the join" quotes that
`errors.As` finds "the first error" in the tree. So a single call
returns one key, and collecting every key means walking the join's
`Unwrap() []error`. Mutation 1 would catch the error, but the text
invites it.

**Proposed fix:** quote each one, or mark it as the owning task's
pre-check. Name the Navidrome command in the conductor's action 1.
Say "walk the join and `errors.As` each member".

### 11. MINOR: two Files lists miss a file their task edits

**Location:** Task 8 Files (plan:2022-2023) and R-5c13 (plan:829-831);
Task 4 Files and criteria (plan:1570-1580, :1650-1652).

**Defect:**

- **Task 8.** `readForm` is a method on `*handlers` that reads
  `h.cfg.BodyCap` (`handlers.go:321`). To reuse "the same path",
  `admin.go` would have to lift it into a function in `handlers.go`,
  which Task 8's Files do not list. Duplicating it instead invites the
  simplifier to undo it.
- **Task 4.** Its `doctor` criterion ("proved through the existing
  doctor harness") lands in `cmd/dubplate-server/doctor_test.go`, which
  Task 4's Files do not list.

**Proposed fix:** add `internal/auth/handlers.go` (the `readForm` lift)
to Task 8, and `cmd/dubplate-server/doctor_test.go` to Task 4.

### 12. MINOR: Task 6's actor-not-admin refusal has two readings

**Location:** Task 6 outcomes (plan:1795-1799); Task 7 (plan:1928-1929).

**Defect:** Task 6 says the refusal uses "`ErrAdminGrantForbidden`'s
sibling refusal (the no-admin error's text names the rule, 'the actor is
not an admin')". Task 7 quotes the no-admin error's text as "the roster
must name at least one admin". So it is unclear whether this is one
error with two texts or two errors.

**Proposed fix:** name a distinct unexported actor-not-admin error. Task
8 already maps it to 500 `failed`.

## Question-by-question

1. **Majors closed.**
   - **Convergent A** (C1, M2, R3) closes at R-5c8 and Task 6, apart
     from finding 1's validation gap.
   - **Convergent B** (C2, R1) closes at R-5c3, Task 3, and exit gate 8.
     The failure scenario (an unknown-address flood filling the
     per-address map) no longer occurs, because unknown addresses never
     reach that map.
   - **Convergent C** (M3, R2) closes at R-5c5 and Task 1. Stat and
     content come from one descriptor. Finding 7 is a fixture and
     wording fix, not a reopening.
   - **The other majors** close at their cited sites: C3 (fold 11 and
     R-5c18), C4 (the Task 8 refusal table against `http.NotFound`, and
     the Task 9 comparison with an unregistered path through the
     tightened `originCheck`), C5 (the `/tmp` fixture), C6 (the table
     row and the rung 12 carry), and M1 (the display-name carve-outs).
2. **Fold-introduced contradictions:** finding 2 (the cut against
   R-5c20), finding 4 (exports without a caller), finding 8 (the gating
   lists), and finding 11 (Files). The segment and dependency order
   builds as written.
3. **From memory:** finding 10. Every other new mechanism I checked is
   quoted or measured correctly.
4. **Rulings:** all six can be ruled on the evidence given, and all are
   genuine owner choices. Rulings 4 and 5 amend ratified spec text,
   which is why they are Geoff's rather than the conductor's. The "no"
   branches of Rulings 2, 5, and 6 need finding 9's edits.
5. **Ceiling:** the arithmetic is correct. The 0.1M margin should make
   the cut the launch-time default, at Task 7 / Task 8 rather than Task
   6 / Task 7 (findings 2 and 3).
6. **Channel set:** `ChannelCommandLine` settles the comment at
   `identity.go:68-69`. It does not settle the set, as finding 5
   describes.
