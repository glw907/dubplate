# Rung 14a plan review: lens P-B, mechanics and feasibility

Target: `docs/superpowers/plans/2026-10-10-rung-14a-admin-api.md` at `fae3de3` (revision 1). Spec: revision 2.1.
Reviewer: fresh context, lens P-B. Findings are ranked by consequence. Claims were checked against source at `fae3de3`
or run in a scratch module (go1.27.2).

## Claims verified, no finding

- **The scoped gate accumulates from the merge base.** `scripts/check.sh:177-193` takes `git merge-base master HEAD`
  on any branch other than master, then diffs the base against the working tree and adds untracked files. The change
  set is therefore cumulative over the pass. Master ledger commits made from the main tree do not move that base,
  because the branch is never rebased mid-pass.
- **The rendered leg runs from Task 1 on.** `check.sh:246` matches `internal/identity/*`. Task 2's own files
  (`cmd/dubplate-server/serve.go`, `web.go`, `auth.go`) also match `:247-250`. Moving Task 1's comment edit later
  would save the rendered leg on Task 1 alone, which is not worth a fold.
- **The conform suite joins at Task 9.** Task 9 edits `internal/engine/store/inspect.go`, which matches `check.sh:229`.
  From Task 9 on, every scoped gate also runs `internal/harness/conform`, about 130-160 s per HISTORY. The plan's
  22-minute task rate absorbs this.
- **The lane settings are right.** `PASS_CLASSES.tool.gateLane` is `"light"` (`pass-execute.js:340`), and a task's
  `gateLane` overrides it (`:446`). The heavy override on Tasks 2 and 9 is therefore correct. `reviewer`,
  `reviewerModel`, `maxFix`, `stopOnEscalate`, and a per-task `model` are all real runner arguments (`:847-857`,
  `:990`).
- **`hex.DecodeString` accepts uppercase.** `"ABCD"`, `"abcd"`, and `"AbCd"` each decode to `[171 205] <nil>`, and each
  re-encodes to `"abcd"`. The quotes for `strings.CutPrefix`, `Handle.Update`, and `ErrBusy` (`conn.go:34-36`) match
  the source. B7 covers one edge case.
- **The worktree layout satisfies CLAUDE.md's lane rules.**
  - The pass needs no venv, no `make derive`, and no beets.
  - The conform suite reads tracked `corpus/cases`, and the rendered leg needs nothing gitignored (`env/rendered/`
    has no reader).
  - `.gitleaks.toml` allowlists `.claude/worktrees/.*`, so the main tree's scan skips the worktree.
  - `.git/info/exclude:11` hides the worktree from `vale-comments.sh`'s untracked listing, and the go tool ignores
    dot directories.
  - The user-cache prerequisites (the unidecode wheel, the XSD, and the Playwright browsers) are shared across trees.
  - `.github/` is absent at `fae3de3`, so there is no CI shadow.
- **Tasks 1 to 9 build in the order given.** No task consumes an export a later task produces. `api.go` is created in
  Task 4, after Task 2 lands hook 18, and Tasks 6 and 7 extend it.

## Findings

### B1. Blocker: Task 10 cannot finish inside its files or its territory, and the stop rules end the run there

**Location:** plan `:524-547`, plus `:103-105` (territory).

**Defect.** Retiring the roster API breaks or rewrites surfaces that Task 10 does not list, several of them in rung
7's append-only territory:

1. **The deck-coverage gate fails.** `web/scripts/gates/deck-coverage.ts` parses `internal/auth/admin.go`'s
   `outcomeX = "..."` constants. It hard-requires `admin_not_grantable` and `duplicate_username` (`:160`) and at least
   21 parsed outcomes (`:167`). P8 drops `admin_not_grantable`, and the handler removal takes the roster outcome
   constants (`admin.go:35-44`) with it, so the `web` leg goes red. The gate also requires every `TITLES` key to
   appear in `REQUIRED_TITLES` (`:59-75`), so the new page's title fails it as well.
2. **Deck coverage goes stale.** The page now keys on the `error.detail` enumeration, which lives in `internal/ops`.
   No gate checks that enumeration's copy, so AC19's "each refusal's own copy" loses its gate.
3. **`RequireAdmin`'s tests sit on the deleted handler.** `internal/auth/admin_test.go` builds its whole refusal table
   on `NewRosterAPI` (`:75`, `:174-244`, `:265`, `:345-384`). Retiring the handler means rehosting the table.
4. **`serve_test.go` breaks.** It drives `/roster` in about ten tests (`:810-1208`).
5. **The rendered harness needs in-place edits.** The members flows in `web/scripts/rendered/run.ts` intercept
   `**/roster` (`:3098`), and the console allowlist regex names `roster` (`:866`). `rendered_test.go:218-219` carries
   `members/malformed-roster/*` rows built on the old JSON shape. Each needs an edit in place, not an appended row.

The plan's rule that "an escalate the conductor cannot resolve inside the task's files" stops the run would fire at
the last build task. Task 10 is also far above the 0.6M per-task price: two pages, a client module, the retirement,
a rehosted auth test table, and harness rewrites.

**Proposed fold.**

1. Split Task 10 into three tasks:
   - **10a, the members page onto `/api/v1`** (`engine-logic`): the client module, the page, the deck rows, and the
     existing members rendered rows and `roster` patterns rewritten.
   - **10b, the server settings page** (`engine-logic`): the route, the `web.go` row, the `serve.go` registration,
     the navigation entry, `TITLES` plus `REQUIRED_TITLES`, and appended rendered rows.
   - **10c, the roster API's retirement** (`auth-data`): the `admin.go` handlers and outcome constants removed;
     `admin_test.go`'s `RequireAdmin` table rehosted on a stub handler, keeping every refusal row with a mutation
     proof; the `serve_test.go` roster rows turned into 404 proofs; and deck-coverage's outcome source moved to the
     detail enumeration (or `openapi.yaml`'s enum), with its floor counts re-derived.
2. List every file named above in its task.
3. Record a contested-surfaces erratum: 14a edits the existing members rows in `run.ts` and `rendered_test.go`, and the
   deck-coverage outcome source, in place. Rung 7 is only planning today, so this is a coordination note, not a fork.
4. Re-price the pass: two more tasks put the ceiling near 11.3M (flag 9.0M) and add about 45 minutes of clock.

### B2. Major: Task 10's `engine-logic` fix rounds reduce to `--reduced`, which never runs the tests the fix touched

**Location:** plan `:66-67` and `:531`.

**Defect.** The plan sets `reducedGate` to `check.sh --reduced`. Under `engine-logic`, `testOnlyReduces` is `true`
(`pass-execute.js:310-316`), so a test-only fix round takes the reduced gate. `--reduced` skips the rendered leg
(`check.sh:261-263`), `svelte-check`, and the static web gates. It runs only `npm run build` and `go test` over the
changed packages, without the `rendered` tag (`:377-406`). A fix to a rendered row or a component test would be
accepted on a gate that never ran it.

**Proposed fold.** Set Task 10's `reducedGate` (each of 10a, 10b, and 10c after B1) to the per-task gate string, so
its fix rounds never reduce.

### B3. Major: the member cap has no carrier into `internal/auth`, and the accessor's read rule is not quoted

**Location:** plan Task 4, `:356-389`, and P7.

**Defect.** `internal/auth` imports neither `internal/settings` nor `internal/config` today. Task 4 has the add path
"count ... inside the add transaction" and refuse when "a cap that cannot be read". Its constructor change names only
"the epoch store and a clock" (`:375`). Nothing says how the roster obtains the cap, on either channel (`serve`'s
`add_member` or the CLI's `member add`).

The settings accessor's contract also goes unquoted: "A setting read before a write enters [store.Handle.Update] is
the value in force for that operation; read it outside the callback" (`internal/settings/accessor.go:20-21`). An
implementer could import `internal/settings` into `internal/auth`, or read the accessor inside `Update` through the
read pool. Either choice ends in a fix round on an `auth-data` task.

**Proposed fold.** State the carrier on the existing precedent: an `auth.Setting[int]` built in `cmd/dubplate-server`
from the accessor (`cmd/dubplate-server/auth.go:499-516`). `serve.go` and `member.go` both pass it (`member.go` already
holds `reg`, `:38`). The cap is read before `Update` and the live count is taken inside it. Quote the accessor's rule
in "Behaviors this plan relies on". Name the fixture for an unreadable cap: the setting function returns an error and
nothing is written.

### B4. Major: Task 4's constructor change makes every CLI roster verb open the auth leaf, with no states named

**Location:** plan `:356-357` and `:375`.

**Defect.**

1. **New CLI failure states have no named report.** `auth.NewRoster(h)` is called from `member add` and `member list`
   (`member.go:97`, `:133`) and from `serve.go:324`. If the constructor takes the epoch store, both CLI verbs must now
   open the auth store, and so meet `auth.Open`'s leaf checks: mode at most `0700`, owner, and writable record
   directories (`doctor.go:240-300`). That is new CLI behavior with new failure states (the leaf absent, a wrong mode,
   a foreign owner), and none names its report. Task 4 says `member.go` changes "for the roster constructor's new
   arguments only". Task 8 later says the verbs "open the auth leaf", but that is Task 4's change.
2. **`member list` gains an auth-leaf dependency.** A read-only verb would now depend on a writable auth leaf for no
   reason.
3. **The file list is short.** Task 4's list omits the other `NewRoster` callers: `roster_test.go` (10 sites),
   `admin_test.go`, `seam_test.go`, `serve_test.go`, and `rosterreset_test.go`.

**Proposed fold.** Decide the shape in the plan. The recommendation is that the epoch store and clock reach only the
write paths, whether through the write methods or a write-capable roster that `member list` never builds. Name each
state `member add` meets when it opens the auth leaf, with its message. Move the auth-leaf opening for `member add`
into Task 4's outcomes. List the test files.

### B5. Major: the close order leaves fixes ungated and runs three full gates

**Location:** plan `:551-595`.

**Defect.**

1. **The architecture-read fold lands after the merge.** Step 8's reads and their "one fold task" come after step 7's
   merge ritual, so that fold's Go lands on master with no merge gate and no `diff-reviewer`. HISTORY already records
   an earlier pass running these reads before the merge for this reason ("Deviations": "The architecture reads ran
   before the merge, not after it as the plan orders").
2. **Two read steps have no fold.** Step 4's UI reviewers and step 6's render-and-read have no fold task, so any fix
   either skips the step 3 gate or forces another full gate.
3. **The close runs three full gates.** Step 3 runs `check.sh` in the worktree, step 7 runs `check.sh --merge` in the
   worktree, and step 7 runs `check.sh` on master again. Each takes about 8-10 minutes on the heavy lane.

**Proposed fold.** Reorder the close:

1. Run every read first: the second security read, the UI reviewers, the architecture reads (no plan in context),
   the live smoke, and the render-and-read.
2. Run one combined fold task.
3. Run the simplifier.
4. Run `check.sh --merge` in the worktree as the close gate. It is the full gate plus the history scan; on red,
   revert the simplifier first.
5. Rebase, then `--ff-only`, then `check.sh` on master.
6. Run the docs fold, which is documentation only.

This removes one full gate and closes the post-merge code hole.

### B6. Minor: Task 9's doctor row overbuilds a warning state and leaves its states unnamed

**Location:** plan `:502-522`.

**Defect.**

1. **No new warning state is needed.** Every row function already returns `(status string, failed bool)`, and
   non-failing advisory rows exist today ("no store yet; start serve, then member add --admin", `doctor.go:223`). A
   new tri-state would ripple through every row, and the row functions are a contested surface rung 7 adds to.
2. **The row is printed before the store is read.** `DUBPLATE_ADMIN_EMAIL` is a wall row printed by the generic
   `OwnedWall` loop (`doctor.go:64-66`), before `storeRow` runs `Inspect` (`:99`). The task never says how the row
   gets the roster standing. Rows print at the end (`:165`), so the status can be resolved after the store row with
   print order unchanged, but the plan should say so.
3. **Some states have no named report (planning-miss item 2).** The plan names none for these:
   - `Load` refused the environment, which leaves the address unknown;
   - the store is absent, unvouched, or older than `UsersVersion`;
   - `Inspect` fails to read the standing;
   - the cap row when the cap cannot be read.

**Proposed fold.**

- Render the warning as the row's status text with `failed=false`.
- Name how the row reads the standing.
- Add a states line per case above, reusing the members row's "not read" vocabulary.

### B7. Minor: the `user_id` re-encoding rule alone accepts the empty string

**Location:** plan `:186-188` and the Task 6 mutations at `:440-442`.

**Defect.** `hex.DecodeString("")` returns `[] <nil>` and re-encodes to `""`, so "refuses any input that is not its
own lowercase re-encoding" admits `""` unless the length check is separate. `"000...0"` (32 zeros) also passes; it is
the `identity.UserID` zero value, which `identity.go:205-208` defines as "never resolved". The states table names
"wrong length", but no fixture or mutation pins either case.

**Proposed fold.** Add an empty-string fixture and an all-zero fixture to Task 6, with the expected answer for each:
`invalid_request` for the empty string, and `not_found` or `invalid_request` for all zeros, the plan's choice. Add
"drop the length check" as a named mutation.

### B8. Minor: AC16's concurrency proof names no synchronization point

**Location:** plan `:383-387`.

**Defect.** To fail against "guard moved outside the transaction", demotion B must read the admin count before the
first connection commits. Then it must wait on the write lock, which `Handle.Update` bounds "by the busy timeout
alone". The fixture names neither part:

- how the test knows B has reached the lock wait before it commits A;
- the busy timeout's relation to the hold time.

Without the first, the mutation can pass. Without the second, the test can answer `busy` under load, and the plan
lets the conductor rule a flake alone.

**Proposed fold (planning-miss item 4).** Name the fixture's ordering. Either B's guard read is observed before A
commits, through a store-level wait hook or a documented bounded poll, or the test runs the mutation variant
explicitly. Set the busy timeout well above the hold.

### B9. Minor: Task 2 converts `Setting.Sensitive` from `bool` to the closed mark without proving existing values survive

**Location:** plan `:306-324`.

**Defect.** Every existing store-homed setting already carries `Sensitive` (`internal/config/settings.go:76`). Task 6
gates re-authentication and the notice on that mark. A key flipped from sensitive to not during the type change
removes the gate silently. The named mutations cover only a zero mark and an empty description.

**Proposed fold.** Add a table test pinning every existing setting's mark, built from the pre-change values at the
task's base commit. Add "flip one existing key's mark" as a mutation.

### B10. Minor: the lane-launch throughput baseline repeats Task 0's full gate on the same content

**Location:** plan `:258` and `:282`.

**Defect.** Task 0 runs `check.sh` (full, heavy) on master. The worktree is cut from that commit, so the tree content
is identical. Receipts are keyed by repo toplevel, so the receipt does not carry across, and the baseline costs about
8-10 minutes for no new evidence. The first per-task gate already proves the worktree's `npm ci`.

**Proposed fold.** Use Task 0's timed full gate as the throughput baseline.

### B11. Minor: the live-executor check runs after Task 0 has written master

**Location:** plan `:246-264`.

**Defect.** The lane launch is headed "before Task 1", but Task 0 already edits `go.mod` and `web/package.json` in the
main tree. The main tree holds untracked agent-memory and research files the conductor did not author. CLAUDE.md's
safety rail runs the live-executor check before any executor touches the repo.

**Proposed fold.** Run steps 1 and 5 (the live-executor check and the guards) before Task 0.

### B12. Minor: once conform is in scope, the scoped gate builds the web bundle twice

**Location:** `check.sh:332-335` and `:371-374`; `Makefile:170`.

**Defect.** In `--scope-changed`, when both the conform and rendered scopes are full (Tasks 9 and 10, and their fix
rounds), `check.sh` runs `make web` and then `make check`. `make check` runs the `.PHONY` `web` target again. That
means four bundle builds and two `svelte-check` runs per gate, about 1-2 minutes wasted per run. More broadly, the
cumulative change set means every task from 2 to 8 runs nearly the full gate (only conform and the full gitleaks
scan narrow). That is a `check.sh` design limit against pass-core's blast-radius rule, not a plan error, and the plan
priced it.

**Proposed fold.** No plan change. File both items to BACKLOG at the first checkpoint:

- the double `make web`;
- a git-derived per-task base (for example, the runner's base SHA through an environment variable) for the scoped
  form.

### B13. Minor: the implementer's "effort high" cannot be set through `pass-execute`

**Location:** plan `:56`.

**Defect.** The runner passes only `{ model }` to the implementer dispatch (`pass-execute.js:857`), and
`dubplate-implementer`'s frontmatter carries no effort. No argument in the plan sets it.

**Proposed fold.** Drop the effort claim, or record it as the session default the dispatch inherits.

## Proportionality

- **Task classes fit, except in Task 10.** Task 2's `tool` class is acceptable once B9's proof lands, because the mark
  is the re-authentication gate's input. Task 10's retirement half belongs in `auth-data`, which is B1's 10c.
- **The per-task gate is the cumulative scoped gate.** No narrower git-derived tier exists today, so the gate is
  acceptable and correctly priced (B12).
- **The close and the lane launch carry avoidable gate cost.** The over-ceremony is three full gates at the close
  (B5) and a duplicate baseline gate (B10), together about 16-20 minutes of heavy-lane clock.
- **No owner fork.** The rung 7 territory edits in B1 are a coordination erratum while rung 7 is still planning.

## Summary

Thirteen findings: one blocker, four major, eight minor. No owner fork.

1. **B1 (blocker).** Task 10's roster API retirement breaks the deck-coverage gate, which hard-requires
   `admin_not_grantable` and 21 outcomes parsed from `admin.go`. It also strands `RequireAdmin`'s test table and
   rewrites the existing members rendered rows, all outside its file list and in rung 7's append-only territory. The
   unattended run would stop at the last build task. Split the task into three, with a contested-surfaces erratum.
2. **B2 (major).** Task 10's `engine-logic` fix rounds reduce to `check.sh --reduced`, which skips the rendered leg
   and `svelte-check`. A test-only fix could pass a gate that never ran it.
3. **B3 and B4 (major).** The member cap has no carrier into `internal/auth`, and the accessor's "read it outside the
   callback" rule is unquoted. Task 4's constructor change silently makes `member add` and `member list` open the
   auth leaf, with no named failure states and an incomplete file list.

B5 (major) reorders the close, which closes a post-merge code fold and saves one full gate. The plan's empirical
claims hold: the cumulative scoped gate, the rendered leg from Task 1 on, the light-lane default and its heavy
overrides, uppercase hex, and the worktree layout.
