# Navidrome removal plan review: contract and criteria lens

Target: `docs/superpowers/plans/2026-10-05-navidrome-removal.md` at `7036ebc` (revision 1).
Spec: `docs/superpowers/specs/2026-10-05-rungs-12-16-replan-design.md` rev 2.1.
Lens: is each acceptance testable and non-vacuous; does each spec requirement map to a task; the four
planning-miss items; whether the pass class fits each task. Every claim was checked against the tree at `7036ebc`
(working tree clean apart from the warm files the plan names).

Counts: **0 blocker, 4 major, 8 minor** correctness findings; **1 minor** over-ceremony note. No owner fork.

## Spec-to-task map

Every requirement in the spec's "The removal pass" section maps to a task. That covers the four `internal/serve`
deletions, the bearer bypass and `statusCapture` retention, the prose sweep, the exported-unused leg, the test
dispositions, the shadow program and its six orphans, the configuration rows, the stale comments, the rename and its
seven surfaces, the `0003` in-place edit and its preconditions, the close list, and Task 0's sweep and baseline. The
roster runbook step and 6d's revision 2.3 sit with the document pass and the next action, as the spec places them.

The eight readings:

| Reading | Verdict |
| --- | --- |
| 1 Exported-unused, "no hit of its own" | Sound. Confirmed at `7036ebc`: the leg's `internal/serve` rows are exactly the nine names listed. Its method needs naming (PC6) |
| 2 Prune `ClassBackground`, `ErrUnavailable` | Sound. The spec's condition triggers it: `docs/superpowers/specs/` holds no rung 9 spec, and no non-shadow production code branches on either name |
| 3 Remove `IntegrationClass.Optional` | Sound. Its escape clause has a false trigger, and its line ranges are short (PC8) |
| 4 Runbook flag lines here | Sound. The sibling plan's coordination section agrees ("two bootstrap commands"). Line 138 is prose rather than a command, so this is not a collision, since sibling Task 6 only verifies |
| 5 IDEAS entry is the document pass's | Sound |
| 6 Comment-only packages take a reader | Literal reading of ruling 6. See OC1 for cost |
| 7 Rephrase "interval driver" | Sound, but it covers only `serve.go`; the test file still fails the grep (PC7) |
| 8 Plan-time box reads | Sound; the re-confirmation is ordered after the archive chore |

No reading needs the owner.

## Correctness findings, ranked by consequence

### PC1 (major): three re-pointed `internal/serve` tests cannot fail on the variant they exist to catch

`docs/superpowers/plans/2026-10-05-navidrome-removal.md:224`, `:227`, `:229`

Once `next` is `http.NotFoundHandler()`, the fallthrough's distinguishing signal is gone. The remaining 404 is
indistinguishable from `Terminator`'s 404, and from the mux's own 404.

- `TestOrderPrefixRequiresSegmentBoundary` (`internal/serve/serve_test.go:160`). Break `Prefix.Owns`
  (`prefix.go:116-121`) to a bare `strings.HasPrefix`. Then `/dubplatefoo` enters the owned mux. `Register` mounts
  only `/dubplate/foo` and `/dubplate/foo/` (`serve.go:103-105`), so the request reaches `Terminator`. It answers
  404 and leaves the spy uninvoked. The re-pointed test passes on the broken variant.
- `TestUnclaimedOwnedPathReachesTheTerminator`. Drop `Terminator`, or route owned-unclaimed paths to `next`. Either
  way the answer is still 404 with no spy reached.
- `TestResponseControllerFlushReachesRealWriter`, "keep the owned half" (`serve_test.go:455-471`).
  `flushingHandler` writes `*dst` only when invoked (`:444-449`). An owned handler that never runs leaves
  `ownedErr == nil`, which is a pass. The plan also names no mutation for it.

The spec requires each re-pointed test to carry a "spy reached" check, or a status a 404 cannot also produce. All
three meet neither. `web-auth-security-reviewer` is told to look for vacuous re-pointed tests first, so this lands
as a close-time fold if it is not fixed now.

**Fold:**
- Segment boundary: delete the shell-level test and record why. Its only observable subject was the fallthrough as
  destination, which is the spec's deletion condition. Name `prefix_test.go`'s `Route` table row for `/dubplatefoo`
  as the proof, with the mutation "`Owns` as bare `HasPrefix`" reported failing there.
- Terminator: keep the test as a behavior pin (an owned unclaimed path answers 404 and invokes no spy). State in the
  dispositions table that no mutation can separate it, because both destinations are 404 by design.
- Flush: add a `reached` flag, or seed `*dst` with a sentinel error. Name the mutation "delete both
  `statusCapture.Flush` and `Unwrap`" (`http.ResponseController` then returns `http.ErrNotSupported`), reported
  failing.

### PC2 (major): the runner arguments contradict the plan's gate lane and reviewer model

`navidrome-removal.md:29-34`, `:48-51`

`~/.claude/workflows/pass-execute.js:262-268` gives `tool` the class default `gateLane: "light"`. Line 359 then
resolves `t.gateLane || a.gateLane || cls.gateLane`, so Tasks 2a, 2b, and 3 run with `CAIRN_GATE_LANE=light` on
the Chromium-launching full gate. That breaks the plan's own "Never `CAIRN_GATE_LANE=light`" (`:31-32`).

`sweep` sets `reviewerModel: "sonnet"` (`:248-253`), and line 569 uses it unless `args.reviewerModel` is set. Tasks
4a and 4b would therefore be reviewed by Sonnet, against the plan's "`diff-reviewer` (`claude-opus-5-5`) per task"
(`:34`). `sweep` also sets `coverageBlocks: false`. A missing state-row test or mutation in 4a or 4b would be
demoted to `batchedNotes` instead of blocking.

**Fold:** add `gateLane: "heavy"` to the execution-mode arguments (`:48-51`). Any value other than `"light"`
keeps the heavy lane. Add `reviewerModel: "claude-opus-5-5"`. Together with PC3, this makes the runner enforce what
the plan says.

### PC3 (major): Task 4b is not a `sweep`

`navidrome-removal.md:22`, `:442`

Task 4b changes three contracts:
- the on-disk roster's accepted key, which feeds the principal and the admin-flag file, through
  `parseRoster`/`RejectUnknownMembers`;
- the roster API's wire codes and its closed field set;
- a required CLI flag.

Its acceptance asks for a nine-row states table, three named mutations, and a fixture-state clause (`:474-496`),
which is the `auth-data` mandate. `sweep`'s mandate says "add no new test unless the plan names one" and "coverage
notes go to nonBlocking" (`pass-execute.js:249-250`). The class therefore under-gates the very proofs the plan asks
for. A misfit class is a plan-review finding (`pass-core`, "Pass class").

**Fold:** set Task 4b to `Pass class: auth-data`. The close already runs `web-auth-security-reviewer` and the live
smoke, so the union does not grow. Scope that reviewer's brief to include `internal/auth/roster.go` and `admin.go`.
Task 4a can stay `sweep` once PC2's `reviewerModel` is set: its one mutation and the hash pin are named in its
criteria, and the runner then reviews it at Opus.

### PC4 (major): the operator-supplied baseline is measured as two different quantities, and its method stalls on malformed values

`navidrome-removal.md:166-170`, `:406`

The spec defines the baseline as "the set of values a default install's operator supplies by hand". Every later
slice states its delta against it, and a nonzero delta needs an owner ruling. The plan's method has three problems.

First, it counts every key `check` reports as required. That includes:
- keys `env/dubplate-server.env.template` already supplies as literals (`DUBPLATE_STATE_DIR`,
  `DUBPLATE_ROSTER_FILE`, `DUBPLATE_IDENTITY_SOURCE`);
- secrets the deploy mints (`DUBPLATE_MAGIC_LINK_SECRET`, `DUBPLATE_SESSION_SECRET`).

The spec's own example treats a generated secret as one "an operator never types". A slice 8 `init` that generates
a secret would then read as a negative delta under one reading and zero under the other.

Second, the method names no handling for a value `check` refuses as malformed. Probed at `7036ebc` with an empty
environment, `check` reports five unset keys. Supplying placeholder values then yields only `invalid email address`
and `unknown identity source` refusals, and no new unset keys. The loop as written stalls, or counts a key twice,
unless it names valid values.

Third, Task 3's ruling 7 criterion compares "the required-key count `check` reports with an empty environment",
which is round one (five keys), against Task 0's iterated total (fourteen or more). The two are different numbers by
construction.

**Fold:**
- Task 0 records the full required-key list, each key with a valid sample value, and classifies each as
  template-literal, generated or minted, or operator-typed. "The count" is the operator-typed class, and all three
  lists are recorded.
- Task 3's criterion re-runs the same iterated method at its commit and compares the key list, not a count.
- The close repeats the comparison after Task 4b.

This is a method call, not a fork: recording all three classes keeps either reading auditable later.

### PC5 (minor): Task 2a's unknown-key proof does not name its registry

`navidrome-removal.md:309-310`

The shadow class is registered in `cmd/dubplate-server` (`main.go:112-113`), never in `internal/config`'s seed. A
test written in `internal/config` against a fixture registry reports `DUBPLATE_SHADOW_ENABLED` as unknown before the
task and after it, so it proves nothing.

**Fold:** the fixture is `newProductionRegistry()` (or the test suite's `productionRegistry` helper). Name the
mutation "re-add `registerShadowIntegration(reg)` to `newProductionRegistry`", reported failing.

### PC6 (minor): Task 1's exported-unused criterion cannot be read from the gate's output

`navidrome-removal.md:262`, `:81-89`

The leg prints only "first twenty" of a sorted list (`scripts/check.sh:615-616`). At `7036ebc` the list holds 127
rows, and the `internal/serve` rows sort after `internal/engine/*` and `internal/harness/*`, so the gate never shows
them. An implementer reading the gate cannot check the criterion.

**Fold:** name the method. Run the leg's block (`check.sh:567-611`) standalone into a temp `work_dir`, then
`grep '^internal/serve' exported.txt`. Diff the result against the nine names Reading 1 lists.

### PC7 (minor): the `serve*.go` prose grep fails on the test file's "sweep driver"

`navidrome-removal.md:256-258`, `:106-108`

`cmd/dubplate-server/serve_test.go:73`, `:567`, and `:593-610` say "sweep driver". Line 599 is a test name,
`TestNewServeShell_SweepDriverStopsWithTheGivenContext`. None of them is a rung or history citation, so the
criterion "prints nothing outside rung and history citations in tests" fails, and Reading 7 covers `serve.go` alone.

**Fold:** extend Reading 7 to `serve_test.go`: rephrase the comments and rename the test (for example
`..._NonceSweepStopsWithTheGivenContext`).

### PC8 (minor): Reading 3's escape clause has a ready false trigger, and its files list is short

`navidrome-removal.md:93-97`, `:281-283`

`IdentitySource.Optional` (`internal/config/registry.go:365`) is a different type with a production use
(`cmd/dubplate-server/auth.go:57`, `doctor.go:508`, `:570`). An implementer told to keep both if it "finds a second
production use" may stop on it.

Task 2a's line ranges also omit four places that carry `IntegrationClass.Optional`:
- `registry.go:240-298`: the cross-class claim loop at `:241` and `:251`, `isOptionalIntegrationKey`, and
  `AddThreshold`'s skip at `:280`;
- `registry.go:289`: the "one exemption" sentence;
- `keys.go:202`: the integration lookup;
- `config.go:23-25`: the package doc's Optional rows.

**Fold:** name `IdentitySource.Optional` as out of scope, and add these ranges to the task's files list.

### PC9 (minor): the rename's grep misses the copy it is meant to retire

`navidrome-removal.md:488-489`

`web/src/lib/copy/deck.ts:138` reads "Used to sign in to the streamer." The pattern `streamer[_ -]?username` does
not match it, so the hint could survive the task with the grep green. Only three `web/` files contain "streamer" at
all (`deck.ts`, `+page.svelte`, `run.ts`).

**Fold:** add `git grep -niw streamer -- web/src web/scripts` returns nothing to Task 4b's acceptance.

### PC10 (minor): the Microsoft A-Z check is assigned to an executor that cannot fetch a page

`navidrome-removal.md:466-468`, `:568`

`dubplate-implementer`'s tools are `Read, Write, Edit, Bash, Grep, Glob` (`.claude/agents/dubplate-implementer.md:4`),
with no WebFetch. The acceptance "the report states what the guide says and the URL read" cannot be met reliably.

**Fold:** the conductor fetches the entry in segment B's pre-flight, and the plan carries the quote. Task 4b then
only applies it.

### PC11 (minor): two leftovers make Task 2b's "case-only" incomplete

`navidrome-removal.md:319-339`

- `corpus/README.md:2292-2302` is the schema contract for `adjudications.json`. It still documents the `album` plus
  `release_mbid` subject that Task 2b makes a decode refusal.
- `internal/harness/analysis/adjudication_test.go:454-497` asserts the re-encoded ledger carries no
  `matcher_commit`. Once the field is deleted, that assertion cannot fail.

**Fold:** add `corpus/README.md` to Task 2b's files, so the contract follows the code. Name the re-encode test's
disposition as "delete: its subject no longer exists".

### PC12 (minor): two relied-on behaviors are recalled rather than quoted

`navidrome-removal.md:484`, `:416`

The states table relies on cobra's unknown-flag error for `--streamer-username`, but the quoted-behaviors table
gives only `MarkFlagRequired`. The `0006` fallback relies on SQLite `RENAME COLUMN` keeping the `UNIQUE COLLATE
NOCASE` constraint. That claim is unquoted, and it matters only if the precondition fails.

**Fold:** quote pflag's unknown-flag error text from source. Quote the SQLite `ALTER TABLE ... RENAME COLUMN`
documentation, or have the conductor quote it when re-cutting 4a.

## Checked and sound

- The quoted behaviors match the tree. These were checked: `go doc` for `AddInsecureBypassPattern`,
  `RejectUnknownMembers`, and `MarkFlagRequired`; `net/http/fs.go:430` and `io/io.go:363` at Go 1.27.1; and
  `config.go:311`, `:315`, `:338`.
- The privileged-fields table is complete for the surfaces touched. `principal()` (`internal/auth/roster.go:816-822`)
  confirms the "no username loads, the principal is the address" row.
- Most named mutations do fail on their broken variant:
  - Task 1: the `NotFoundHandler` swap, the `origin.wrap` drop, the `refuseEchoMethods` drop, the `ReadFrom`
    deletion (`plainReader` has no `WriteTo`, and `statusCapture` promotes no `ReadFrom`), and the `Remote-User`
    skip (`cmd/dubplate-server/serve_test.go:1896-1929` asserts it);
  - Task 2b: `DisallowUnknownFields` (`ledger.go:19-20`);
  - Task 3: the superseded-row deletion;
  - Task 4a: `COLLATE NOCASE`;
  - Task 4b: all three mutations.

  PC1 lists the three that do not.
- The box needs no extra precondition for newly unknown keys. `deploy.sh` re-renders `/etc/dubplate/server.env`
  whole from the template (`scripts/deploy.sh:187-188`, `:420`), and the template carries no shadow or Navidrome
  row.
- Sibling collisions: none beyond those the plan names. The removal plan's `.claude/agents/dubplate-implementer.md`
  and `BACKLOG.md` writes sit outside the document pass's territory.

## Over-ceremony

### OC1 (minor): three architecture readers grade one-comment edits

`navidrome-removal.md:513-516`

`internal/engine/tag`, `internal/engine/tagio`, and `internal/engine/matcher` are touched only in comments.
`matcher` has one line (`quiet.go:5`), and `tagio` is a large package. Each reader reads the whole package's
non-test source. That is a meaningful slice of the 8.0M ceiling for findings that will almost all be filed to the
cleanup inventory, out of scope. Ruling 6 says "every touched package", so the plan is literal and correct.

**Recommendation:** keep the readers as written. If the 6.4M flag is near at the close, these three are the first
to drop, and the drop is recorded in HISTORY as a budget call. No other step reads as ceremony beyond what the
class, the spec, or `CLAUDE.md` requires.
