# admin-api-1 plan review: data integrity and failure risk

Target: `docs/superpowers/plans/2026-10-11-admin-api-1.md` at `9c58bcc`. Lens: crash, concurrency (the CLI verbs
while `serve` runs), partial writes, upgrade, rollback, restore, privilege retention, privileged-field channels, task
ordering, and the unattended execution. Code read at the worktree head (`9c58bcc`, code unchanged since `551bf4f`).

**Counts:** 0 blockers, 7 majors, 12 minors. 15 are tagged `correctness`, 4 `optional`. No owner fork.

**What holds.** I checked these and found no gap:

- Cross-process revocation. `EpochStore.CheckSession` reads the record file on every call (`epoch.go:225-241`,
  `readNormalized` does `os.ReadFile`). `Roster.Lookup` reads the users table on every call (`roster.go:249-266`), and
  `MagicLinkSeam.VerifyIdentity` runs both on every request (`seam.go:96-129`). A `member remove` from the CLI while
  `serve` runs is therefore seen at the next request.
- The CLI's write mode coexists with `serve`. `OpenWrite` holds the store lock shared (`open.go:120-121`).
- Bumps from the CLI and from `serve` serialize, because every bump this pass adds runs inside `Update`'s IMMEDIATE
  transaction. `EpochStore.mu` is in-process only (`epoch.go:69-72`), and no production `Bump` caller exists today
  (`grep '\.Bump(' internal cmd` prints nothing).
- A cookie-channel reactivation already returns without the admin flag. `checkAddition` refuses `Admin` off the
  command line (`roster.go:330-331`), and the reactivation writes `admin = ?` from the candidate (`:416`).
- Rollback is safe. Orphan rows never refuse a start: `evaluate` collects them and never refuses on them
  (`evaluate.go:195-201`). The pre-14a binary already handles `admin`-source rows (`evaluate.go:262-264`, `:290-292`).
  A rollback that meets a `DUBPLATE_REAUTH_WINDOW` row therefore starts.
- No API operation can be reached cross-site with the cookie. Every `opshttp` operation is POST with
  `application/json` (`opshttp.go:83-99`), and the cookie is `SameSite=Lax` (`session.go:181`).
- Ordering never leaves a committed task less secure than master. The old `roster` add, with no re-auth gate, is
  master's existing state, and it retires inside the same PR (T11) before anything merges.

## Majors

### M1. Resume skips the review of a task that is committed but not yet reviewed (correctness)

- **Location:** plan:154-161 ("Resume"), plan:123-126 ("Per-task review"), plan:509-513 (T6's security read),
  plan:675-679 (Close step 2).
- **Defect:** Resume "resumes at the first task whose acceptance fails and never re-applies a committed task". The
  per-task `diff-reviewer` runs after the commit, and so does T6's `web-auth-security-reviewer` read. If the session
  crashes, or the API drops, between a commit and its review, the task's acceptance passes on resume. The session moves
  on, and an `auth-data` commit (the roster's epoch ordering, the gate, the mount) ships with no task-scope read.
  pass-core's Close has the whole-branch review "give any unread `auth-data` task its read", but plan Close step 2 does
  not say so, and nothing records which tasks went unread.
- **Fold:** Each review appends one PR-notes line, `review T<n>=<verdict> <sha>`, and T6 also appends
  `security-read=<verdict> <sha>`. On resume, a committed `auth-data` task with no such line is the task in flight,
  at its review step. Close step 2 names every task with no recorded verdict for the whole-branch reviewer to read.

### M2. A pair cut short by a crash can lose a committed half, and the resume rule is ambiguous (correctness)

- **Location:** plan:157-160 (Resume), plan:130-135 (Records), plan:140-148 (Dispatches).
- **Defect:** "a branch with its `T<n>:` commit is merged" can be read as "is already merged, skip it" as well as "merge
  it now". On the first reading, a T4 branch whose half was never merged is silently dropped, and the later tasks build
  without T4. A fresh session also has no way to find the branch names. They reach the PR notes only through the
  subagent's final report, which a crash mid-pair never delivers. The dispatch sha is likewise only in the clock file.
- **Fold:** Before dispatch, append `dispatch T<a>+T<b> sha=<40-hex>` to the PR notes. On resume, find the candidate
  branches with `git worktree list --porcelain` and keep each one where `git merge-base --is-ancestor <dispatch sha>
  <branch>` holds. Then classify each candidate:
  - It carries its `T<n>:` commit and `git merge-base --is-ancestor <branch> HEAD` holds: done.
  - It carries its `T<n>:` commit but is not an ancestor of `HEAD`: merge it, run the fast lane with `--base <dispatch
    sha>`, run the review, and push.
  - It has no `T<n>:` commit: a stop, as the plan already says.

### M3. T4 does not require the validator's rows to be read inside the write transaction (correctness)

- **Location:** plan:417-437 (T4).
- **Defect:** T4 puts the row write, `seed_value`, the generation bump, and the audit row "inside one `Update`".
  The pre-write and post-write evaluations are not placed there. The ready-made entry point, `settings.Evaluate`,
  "reads through h's read pool" (`evaluate.go:110-121`), outside any transaction. If the validator reads its rows
  there, two concurrent writes can each pass: two admins, or `settings reset` on the CLI while `serve` writes. Take two
  keys under one correlated rule, say `DUBPLATE_MAGIC_LINK_EXPIRY` up and `DUBPLATE_NONCE_RETENTION` down. Each write
  passes against the pre-image the other has not changed, both commit, and the next `serve` start refuses the store.
  That is exactly the state spec:228 says a write "cannot add". Clearing it then needs box access. The old value and
  `seed_value` carry the same risk. `Apply` is the in-tree precedent for doing it right: "The rows are read inside the
  write's own transaction" (`apply.go:26-28`, `:33-44`). `Update`'s doc confines its read-inside rule to "the rows the
  write changes", which does not obviously cover the other keys a correlated rule names.
- **Fold:** T4's outcome states that both evaluations run `evaluate` over rows read through `tx` inside the `Update`
  callback, on `Apply`'s precedent, and that the old value and `seed_value` come from the same read. Add a mutation:
  read the rows through `Evaluate` before `Update`. Its fixture is two writes, each valid alone, that jointly break
  `nonceRetentionCoversExpiry`, queued behind a second connection's `BEGIN IMMEDIATE` (AC16's technique). Exactly one
  commits.

### M4. AC16's settings-write actor clause can pass vacuously through the mount (correctness)

- **Location:** plan:466 (T5 acceptance names "AC16's settings-write actor clause"), plan:471-481 (T5 fixtures),
  plan:482-485 (T5 mutations), P9 (plan:213-216).
- **Defect:** No T5 fixture names the state this clause needs: an actor demoted, or removed, after verification and
  before the write. Through the mount, the verifier reads `Admin` from the roster on each request (`seam.go:140`), so
  the authorizer refuses a demoted actor with `forbidden` first. A test driven that way passes even when
  `internal/ops/admin` passes no in-write guard to the settings functions. Their guard parameter is "absent on the
  command line" by design (P9), so a nil guard on the cookie path compiles and runs. No mutation covers it. T3's roster
  fixture covers member writes only.
- **Fold:** Add a T5 fixture that calls the admin seam directly, with an admin-privileged cookie actor whose row a
  direct store write has demoted (and one it has tombstoned). `set_setting` and `reset_setting` each answer `forbidden`
  and write nothing. Add the mutation "the admin seam passes no guard to the settings write".

### M5. Nothing proves a committed settings write reaches a running `serve` (correctness)

- **Location:** plan:435-437 (T4 mutations), plan:552-561 (T7 fixtures), spec:289-290 ("with `serve` stopped or
  running").
- **Defect:** `serve`'s `Accessor` caches rows and reloads only when `settings_generation` changes
  (`accessor.go:93-110`). T4 mutates "bump the generation on a no-op" but never "skip the bump on a committed
  write". AC10 asserts the bump only for `set_setting`. If a reset, or `settings reset` on the CLI while `serve` runs,
  commits without the bump, the running server keeps serving the old value until a restart. That could be a sensitive
  key the admin believes they reset. The smoke runs `settings reset` only with `serve` stopped (plan:688-689), and no
  fixture holds a serve-mode handle open.
- **Fold:** Add a T4 mutation, "skip the generation bump on a committed reset". Its fixture is an `Accessor` over a
  second handle that reads the key first, then reads the new value after the reset. Add one T7 fixture that runs
  `settings reset` while a serve-mode handle (shared lock) stays open, and that handle's accessor then reads the
  default.

### M6. `internal/ops/admin` cannot classify the roster's refusals, and no task gives it a way (correctness)

- **Location:** plan:388-389 (T3 outcome), plan:453-454 and plan:462-464 (T5 outcome and Files), P7 (plan:200-205).
- **Defect:** Every roster refusal is unexported: `errNoAdmin`, `errActorNotAdmin`, `errMemberNotFound`,
  `errDuplicateAddress`, and the field refusals (`roster.go:39-105`). Only `additionOutcome`, in-package, maps them
  (`admin.go:312-337`). T5 must turn them into `forbidden`, `not_found`, `conflict` with `last_admin`, and P7's add
  details, but T5's Files list no `internal/auth` file, and T3 exports nothing for it. Two outcomes follow, and both
  are bad:
  - T5's subagent edits outside its files and lands in a contested-file stop.
  - Or it maps by fallthrough, and `last_admin` or `forbidden` turns into `internal`. AC16's `conflict` and `last_admin`
    assertion catches that, but only after an `opus` dispatch is spent.
- **Fold:** T3 adds the exported classification the seam needs (exported sentinels, or one exported outcome function
  over P7's detail names) and names T5's `internal/ops/admin` as its production caller. T3's Files and acceptance say
  so.

### M7. A store restore is a privileged-field channel that neither table names (correctness)

- **Location:** plan:270-278 (privileged-field table, which adopts spec:471-481 by reference), plan:572-576 (T8
  runbook line).
- **Defect:** A restore brings back the users and settings tables as they were at the snapshot. The auth leaf is not
  backed up (spec:339, `internal/backup/doc.go:1-9`). A restore after a demotion therefore sets a demoted member's
  admin flag again. After a removal it clears a tombstone. After a reset it brings back an `admin` settings row. The
  table says the admin flag is set by "the CLI grant alone", which is false once a restore runs. T8's runbook line
  tells the operator to re-run removals only, so a demotion made since the snapshot is silently undone. That is the
  "a demoted member keeps access" case this lens asks about.
- **Fold:** Add a "store restore" row to the plan's privileged-field table, naming the admin flag set, the tombstone
  cleared, and `admin` settings rows. T8's runbook line reads "re-run any removal or demotion made since the
  snapshot". This widens the spec's sentence by two words and changes no design.

## Minors

### m1. T6's security read has no verdict or re-read rule (correctness)

- **Location:** plan:509-513.
- **Defect:** Its findings "join the per-task review's one fix commit". It is unclear who re-reads that commit, and
  whether an escalation from this reviewer is a stop. Without a rule, an auth fix can land with only the
  `diff-reviewer` re-read.
- **Fold:** The security reviewer re-reads the fix commit. Its blocker or major findings count as a `fix`, and its
  escalation is a stop under "Stops".

### m2. The re-add bump's failure policy is not stated (correctness)

- **Location:** plan:378-379.
- **Defect:** "In the same callback and order" does not say whether a non-backward bump error rolls the re-add back,
  or whether `ErrEpochBackward` commits it. The mutations cover the removal only.
- **Fold:** State the removal's policy for the re-add, and add the mutation "roll back a re-add on
  `ErrEpochBackward`".

### m3. The smoke's 5-minute window can fail the gated steps falsely (correctness)

- **Location:** plan:683-689.
- **Defect:** The window is seeded at `5m`, and the other steps run inside the wait. Those steps include a sensitive
  write, add, remove, re-add, and demote, each gated (D10). Any gated step reached more than 5 minutes after its
  session's sign-in answers `reauthentication_required`, and that reads as a defect. The result is a misattributed fix
  chain or a stop.
- **Fold:** Mint the stale session A first and start the wait. Run the gated steps from a session B signed in right
  before them. Use A only after the wait.

### m4. A T0 bump can move a library whose behaviors the plan quotes (correctness)

- **Location:** plan:295-313 (T0), plan:236-263, spec:43-70.
- **Defect:** The kin-openapi silent passes, the 2020-12 fallback, and the five formats are quoted at v0.149.0. The
  sqlite busy behavior rests on `modernc.org/sqlite` v1.60.1. A minor bump in T0 lands without re-quoting either, and
  the contract tests' "every other silent pass fails" list could then miss a new one.
- **Fold:** If T0 moves kin-openapi, `modernc.org/sqlite`, or go-mail, the task that relies on it re-quotes the
  behavior at the new version in the PR notes before its code lands: T6 for kin-openapi, T3 and T5 for sqlite, T5 for
  go-mail.

### m5. The window rule can block a start after an upgrade, and the States table does not name it (correctness)

- **Location:** plan:194-199 (P6), plan:289 (States).
- **Defect:** `DUBPLATE_SESSION_ABSOLUTE_CAP` has floor `1s` (`auth.go:146-155`). A box whose cap is at or below `2h`
  refuses to start after the upgrade under the default window. A cap at or below `5m` can never start, because the
  window's floor is `5m`.
- **Fold:** Add a States row: on `serve` start after an upgrade with the cap at or below the window, the refusal names
  both keys and both remedies (lower the window, or raise the cap). No code change.

### m6. T4's refusal fixture omits thresholds and integration-class keys (correctness)

- **Location:** plan:431-434.
- **Defect:** It covers only "a wall key" and "an unregistered key". A lookup over the wrong family could write an
  orphan `admin` row for a threshold, and that row would then show a value the server never reads.
- **Fold:** Add one threshold key and one integration-class key to the fixture. Each answers `invalid_request` and
  writes nothing.

### m7. T2's Files miss the test declaration sites that its panics break (correctness)

- **Location:** plan:356-358.
- **Defect:** `Add*` panics on a missing description or a zero mark. That fires in
  `internal/settings/{accessor,apply,helpers}_test.go`, in
  `cmd/dubplate-server/{auth,check,doctorprobe,doctor,store,unidecode}_test.go`, and in the `internal/config` tests
  (all found by grep for `AddSetting|NewRegistry|config.Setting{`). A subagent that keeps to the listed files goes
  red, and a red is a stop.
- **Fold:** List these files. They stay disjoint from T1's.

### m8. The launch prompt detects a resume by the worktree directory alone (correctness)

- **Location:** plan:53-56, and the prompt's first paragraph.
- **Defect:** The directory can be gone while the branch `worktree-admin-api-1` still exists locally or on origin.
  A fresh session then calls `EnterWorktree` with the same name, collides with the branch, or redoes the work. A crash
  before the planning-branch merge also leaves no `launch-base` line for Resume to read.
- **Fold:** Resume when either the directory or the branch exists (`git rev-parse --verify`, `git ls-remote --heads
  origin`), and stop when the branch exists without the directory. On resume, if `launch-base` is absent, check `git
  merge-base --is-ancestor worktree-admin-api-1-plan HEAD` and finish the launch steps.

### m9. The busy States row overstates "nothing written" (correctness)

- **Location:** plan:285.
- **Defect:** A removal whose bump ran before a deadline or busy failure at commit answers `busy`, yet the epoch has
  moved. The member is still live but signed out (spec:352-353). A test written to the row's literal text could assert
  the session still verifies.
- **Fold:** The row says "no store row written; a removal's bump may already have revoked the target's sessions".

### m10. The notice's deadline should not inherit the request's cancellation (optional)

- **Location:** plan:450-452.
- **Defect:** If "its own deadline" derives from `r.Context()`, a client that disconnects cancels the send. The
  notice is then lost, though logged.
- **Fold:** Name `context.WithoutCancel(r.Context())` with a timeout as the send's context.

### m11. The fast lane's `--base` departs from the adapter without saying so (optional)

- **Location:** plan:87-97, against dubplate-pass SKILL.md:71 ("Pass no `--base`").
- **Defect:** pass-core says the adapter wins where it disagrees. The plan's reason is sound and the code supports it
  (`check.sh:268-293` validates the base; `:364-370` and `:458-459` promote machinery files to the full form). But the
  departure is not named, so a reviewer or a resumed session may "correct" it.
- **Fold:** One sentence naming the adapter line it departs from, and a "Found, not done" item to amend the adapter.

### m12. Close step 3 multiplies screenshots in the main loop (optional, over-ceremony)

- **Location:** plan:39-40, plan:680-681.
- **Defect:** Every refusal state on both pages, in both palettes, at two widths, comes to about 40 or more
  screenshots read by the Sonnet session. That is roughly 20 to 30 minutes and a large token cost. The refusal states
  differ in copy, not in layout per palette or width.
- **Fold:** Read each page's default state in all four palette and width combinations, and each refusal state once,
  in one palette at phone width.

## Considered and not raised

- T6's extra `web-auth-security-reviewer` read over T1 to T6. It costs one Opus read. I keep it, because T9 to T11
  build the pages on that API, and a defect there is cheapest to catch before they do.
- AC16's timing. Correct code, guarded inside the transaction, refuses whatever the interleaving. The mutant fails
  only when the second demotion starts within the hold, and the recorded ten-of-ten runs are the stated proof, so a
  wait-until-blocked hook is not worth adding.
- A CLI verb run as bare root, which writes a root-owned epoch file that `serve` cannot read. The runbook already
  forbids running the binary as bare root (runbook:173-176), and a failed read fails closed with a 500.
