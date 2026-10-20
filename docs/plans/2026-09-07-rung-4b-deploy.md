# Rung 4, Pass 4b (The Deployment) Implementation Plan

**Revision 1, 2026-09-07.** Drafted against the rung 4 spec at revision 2
(`docs/specs/2026-09-05-rung-4-shadow.md`), whose six owner
rulings Geoff ratified as recommended on 2026-09-07, so nothing in this
plan waits on a ruling. Band **S**. The pass runs unattended to the
sitting boundary and then **stays open**: ruling 6 makes the first deploy
an owner sitting, and that sitting is a later attended session.

**Revision 2, 2026-09-07.** Folded from two fresh-context adversarial
reviews, recorded at
`docs/research/2026-09-07-rung-4b-plan-review.md` with the
merged do-not-re-litigate list. One blocker, eight majors, thirteen
minors; none rejected. The load-bearing corrections, in the order they
change what an executor does. **Revision 1 told the implementer that
omitting the host argument aborts; on this workstation it silently
resolves the production box**, because `HCLOUD_TOKEN` and `hcloud` are
both present, so the address is now mandatory and every command that
takes one is quoted in full. **Tasks 3, 4 and 5 do issue ssh**, to the
unroutable documentation address, which the implementer agent's blanket
prohibition forbade and revision 1 described as contacting nothing; the
prohibition is now scoped to the box, the cost is named, and the
expected connection errors are named as expected. **Task 4's helper
contract could not be satisfied**: one render-and-install helper called
twice cannot render both templates before the first remote step, so R5
now splits into a render pass and a per-template install step. **Two of
Task 4's three mutation proofs stated results the workstation's own
environment contradicts**, one because `deploy.sh` re-sources
`~/.local/secrets` and one because `DUBPLATE_STATE_DIR` is already
exported there. **Task 2 is no longer orderable after Task 4**, which
depends on the file Task 2 creates. **Task 0 regains the spec's first
staleness blocker**, whether rung 1b has landed, with a 1b plan and
inventory sitting untracked in this tree as this pass opens. **The
sitting's apply is a full-estate converge**, so its blast radius is now
listed and a compose and Navidrome check follows it. Two resolutions
join the list, R15 scoping the ssh prohibition and R16 declining a
host-resolution guard, and R1 records `StateDirectory=` as a third
declined option. Nine citations, one prior-art count, one resolution
count, and one grep's case sensitivity are corrected.

**Conductor rulings, 2026-09-07, under the unattended-window grant.** Six
items are ruled and applied as written: R15 (the ssh prohibition scoped
to the box, with the quoted `deploy.sh --check` against `192.0.2.1`
allowed), R16 (no host-resolution guard added to `deploy.sh`), R5 (the
render pass over both templates before `deploy.sh`'s first step, plus a
per-template install step), R1 (the leaf grant
`/var/lib/dubplate/shadow` with a two-level pre-create, `StateDirectory=`
declined), the ceiling branch (the 3M ceiling stands, and a second `fix`
verdict on any task is the conductor's decision at the 2.4M flag), and
R13 (the `Makefile` rides the `systemd/`, `scripts/deploy.sh`, `env/`
ownership row, now written into
`docs/specs/2026-09-03-parallel-build-approach.md`'s section
3 table). R2, R3, R4, R6, R7, R8, R10, R11, R12 and R14 retired at the
4b close (`docs/HISTORY.md`'s "The ten retired resolutions"), R12
declined with no file change. Task 0 ran on 2026-09-07 with no blockers;
its one finding, the gitleaks red in
`docs/research/2026-09-07-rung-1b-plan-review.md`, is fixed
in this commit.

> **For agentic workers:** workflow mode per CLAUDE.md's Build process
> section, the `~/.claude/workflows/pass-execute.js` chain per task. The
> implementer is `cairn-implementer` instructed to read and follow
> `.claude/agents/dubplate-implementer.md`; the reviewer is
> `diff-reviewer`; the gate is `bash scripts/check.sh`, green at every
> task end. The conductor consumes reports and never reads diffs inline.
> Subagents never dispatch agents: the `code-simplifier` dispatch and the
> fresh-context unit read are the conductor's. The conductor owns
> `docs/STATUS.md`, `docs/HISTORY.md`, `ROADMAP.md`, the lane registry,
> the contested-surface rows, and the batch commit of anything a task
> drafts into them. **This pass runs in the MAIN TREE.** The lane
> registry records lane-1 as IDLE pending this plan, lane-2 as BLOCKED,
> and the doc lane as IDLE, so no second executor is live.

**This pass writes to the box exactly once, and not in this window.**
Every task below finishes without a single write to `5.78.148.42`. The
deploy, the timer enable, the SELinux verifications, the journal check,
and the recorded deploy date all belong to the sitting. One read-only
exception runs in this window, the conductor's own `deploy.sh --check`
against the box (Resolution R11); the Box contact rule below is the
authority on it.

## Primary inputs

Read these before Task 1, in this order.

1. **`docs/research/2026-09-07-rung-4b-planning-inventory.md`,
   in full.** It is this plan's argument, read off the tree at HEAD
   `c98a9166b38c6ec1c6bdc23ca280d1bb1c58526a`. Its thirteen-item
   **surprises list is binding**: where the tree contradicts the spec,
   this plan follows the tree, and Task 8 drafts the spec correction. Do
   not re-derive what it settles; cite it, and verify the line if you
   carry one forward.
2. **`docs/specs/2026-09-05-rung-4-shadow.md`, revision 2.**
   The sections this pass implements: "The deployment" (`:1147-1435`),
   "Config keys" (`:1436-1484`), the pass-split table's 4b row
   (`:1632`), the contested-surface rows (`:1639-1661`), and Process
   (`:1800-1848`).
3. **`docs/HISTORY.md:385-481`**, the "4a to 4b handoff" and the forward
   carries. Six numbered handoff items are this plan's inputs, stated
   rather than derived. The 4a-ii entry above them (`:5-384`) is the
   record of what landed.
4. **`~/.claude/skills/vps-conventions/SKILL.md`, in full**, before the
   unit files, `deploy.sh`, or anything that places a file or a directory
   on the box.
5. **`CLAUDE.md`**, in full: the Gate section, the Stack rules, the
   build-for-2.0 rule, the prior-art rule, and the Safety rails.
6. The tree, read-only: `systemd/dubplate-serve.service` (the hardening
   template, `:96-125`), `systemd/dubplate-disk-check.service` (the
   oneshot shape that takes `network-online` without `docker.service`),
   `systemd/dubplate-navidrome-check.timer` (the timer this pass's timer
   matches exactly), `systemd/dubplate-ping-fail@.service`,
   `scripts/deploy.sh` in full, `env/dubplate.env.template`,
   `Makefile:1-30`, `cmd/dubplate-server/main.go`,
   `internal/config/{keys,config,thresholds}.go`,
   `internal/engine/shadow/{store,config}.go`.
7. **`~/.claude/skills/go-conventions/SKILL.md` before any Go.** Task 5
   is the only task writing Go, and it writes about ten lines of it.

## Goal

Deploy the shadow runner. Land the systemd unit and timer that run
`dubplate-server shadow observe` on the box under a minimal write grant,
the greenfield environment template that configures it, the `deploy.sh`
edits that install it, and the two build stamps that make the runner's
own unstamped-commit refusal observable instead of silent.

Leave the tree in a state where one attended sitting can run
`deploy.sh --check`, read every diff, apply, enable the timer by hand,
and watch the first pass land. **Nothing here writes to the box or
enables anything on it**, and the one read-only reach is the conductor's,
under the Box contact rule.

## Owner touchpoints: exactly one, the deploy sitting

Ruling 6 (`spec:345-371`) makes the first deploy of greenfield code an
owner sitting rather than a conductor action. It is this pass's only
attended item, it happens after every unattended task has landed, and it
is a **later session**.

### What the sitting runs, as a checklist

The order is the spec's quiet-box procedure (`spec:1365-1391`), with the
inventory's corrections applied. Read-only ssh at every step until step 4.

1. **Confirm the box is quiet.** No import or backup run in flight, and
   none due inside the window. The import timer fires `*:0/15` with 300
   seconds of jitter (`systemd/dubplate-import.timer:10-11`); backup
   fires daily at 11:30 UTC with 600 seconds of jitter
   (`systemd/dubplate-backup.timer:9-10`). Disk-check shares import's
   schedule and navidrome-check fires `*:0/5`; both are harmless.
2. **Confirm the worktree is clean and on `master`.** `git describe
   --tags --always --dirty` is the version stamp, so a dirty tree ships a
   binary stamped `-dirty` and tells MusicBrainz so on every request
   (Resolution R9).
3. **Run `bash scripts/deploy.sh --check 5.78.148.42` and read every
   step's diff.** Sixteen steps print, the two new ones being the
   `/var/lib/dubplate/shadow` pre-create and the
   `/etc/dubplate/server.env` render. Expect `would run:` lines and three
   `rsync -avzn` diffs. **The new env step prints `would change (contents
   differ; diff not shown, it's secrets)`**, because
   `/etc/dubplate/server.env` does not exist on the box yet; `would write
   N variables` is the unreachable-box wording (`deploy.sh:308`) and
   cannot appear against a live box. **Three diffs decide whether to
   apply**, and they are the reason this step is attended rather than
   skimmed: the `/opt/dubplate` `rsync --delete` diff (`:201`), which
   pushes whatever `compose.yaml` and `config/` currently hold; the
   incumbent `/etc/dubplate/env` step's verdict (`:297`), which re-renders
   from the age store; and the compose step's `would run:` line.
4. **Apply, knowing what the apply re-runs.** `bash scripts/deploy.sh
   5.78.148.42`. **This is a full-estate converge, not a two-step
   addition.** Beside the pass's own two new steps it re-runs eleven
   pre-existing ones: the `/opt/dubplate` rsync with `--delete`, the units
   rsync, the tmpfiles install plus `systemd-tmpfiles --create`, the
   binary rsync and chmod, the smoke test, the incumbent env render and
   install, `contributors.yaml`, the four directory steps, `docker compose
   up -d`, the four guardian timers' `enable --now`, and the beets
   reinstall via uv. The smoke test asserts a non-`dev` matcher stamp and
   aborts the whole deploy if it is missing, so a stamp failure stops here
   rather than thirty days later.
5. **Check the stack survived the compose step.** `docker compose ps` and
   one Navidrome reachability check. `~/.claude/skills/vps-conventions/SKILL.md:90`
   records that a compose step following a network change reattaches live
   containers without their service aliases and silently killed
   Navidrome's container-name DNS on 2026-09-01. Nothing in this pass
   changes compose networking, which is why this is a check rather than a
   `--force-recreate`; it costs a minute and it is the one step that would
   catch an unrelated `config/` change riding along on the converge.
6. **Enable the timer by hand.** `systemctl enable --now
   dubplate-shadow.timer`. This is deliberately not in `TIMER_UNITS`
   (`spec:1311-1317`): an adopter's deploy should not silently start
   observing, and **this command is the act that starts the launch
   clock**.
7. **Verify the first pass in the journal.** `journalctl -u
   dubplate-shadow.service -n 50 --no-pager` after the first `*:0/5`
   window plus up to 60 seconds of jitter. A pass that observes nothing
   is a success and writes `started-at` (HISTORY item 6).
8. **Verify both SELinux contexts.** `ls -Zd /var/lib/dubplate
   /var/lib/dubplate/shadow /var/lib/musicbox`, expecting `var_lib_t` on
   each, and `getenforce` for the record.
   `~/.claude/skills/vps-conventions/SKILL.md:43-45` asks for this once,
   on the first real deploy. The `/var/lib/musicbox` half **closes an
   item carried open since the rewrite pass**, recorded at
   `docs/HISTORY.md:3082-3083` (the spec cites `:2327-2328`, which is
   rung 3a adjudication text; inventory surprise 13).
9. **Record the store's sizing baseline.** The root filesystem's free
   space (`df -h /`), and whether `dubplate-disk-check` watches it. The
   store sits on root, not on the 50 GB `/srv/music` volume
   (`spec:1765-1778`), and this figure is the trigger the retention knob
   waits on (HISTORY forward carry 8).
10. **Record the deploy date as a first-class figure.** Read
    `started-at` off the box to get the clock's own value rather than
    the wall clock of the sitting, and write the date into the close;
    the 720h floor is a gate the launch reads, not a date derived from
    it.
11. **Read the operator rule once, out loud.** After the sitting, a
    hand-run shadow verb goes through `systemctl start
    dubplate-shadow.service` or `systemd-run` carrying the unit's own
    properties, **never the bare binary as root**, because the read-only
    guarantee is the unit's and not the binary's (HISTORY item 5,
    `spec:1387-1391`).

Two owner options belong in the same sitting, neither blocking. A
deliberate backlog feed into the inbox raises the observed count and also
raises the excluded multi-album-directory count (`spec:1730-1741`). And
the rollback is two verbs, not one: `systemctl disable --now` stops
observing and preserves the clock, while removing `/var/lib/dubplate`
discards the evidence and restarts the floor from zero, which is the
owner's explicit decision and never a tidy-up step (`spec:1393-1417`).

### What the unattended tasks must leave ready for it

- Both unit files in `systemd/`, in `UNIT_FILES`, and green under the
  gate's `systemd-analyze verify` leg.
- `env/dubplate-server.env.template` complete, with every literal filled
  and `DUBPLATE_ADMIN_EMAIL` the one store lookup.
- All six `deploy.sh` edits landed, each with its `--check` branch.
- The two stamps reaching the binary through `make build-linux`, and
  `dubplate-server version` printing both rows.
- One conductor-run `deploy.sh --check` against the box, read-only, with
  its output recorded in the task ledger (Resolution R11).
- The fresh-context unit read complete, with every finding either fixed
  or recorded with a reason.
- `docs/STATUS.md` rewritten to point at the sitting, naming the exact
  commands in order, so the sitting session needs no reconstruction.

## The sitting boundary

**Unattended, this window: Tasks 0 through 8.** The staleness pre-task,
the two units with the manifest edit, the environment template and the
header amendment, the three `deploy.sh` manifest-and-directory edits, the
render pass and install step with their literal column, the stamps, the
provisional-mark retirement, the fresh-context unit read, and the handoff
STATUS write.

**Task 8 is the boundary task.** It writes the handoff and leaves the
pass OPEN. It does not score the pass, does not write a HISTORY entry,
and does not record a deploy date, because none of those figures exists
yet.

**Attended, a later session: Task 9.** The sitting checklist above, then
the pass close carrying `started-at` and the deploy date as first-class
figures.

The pass ledger stays open across the boundary. `docs/STATUS.md`'s next
action after Task 8 is the sitting, not the next rung.

## Box contact rule

**No box write of any kind before the sitting.** Not a `mkdir`, not a
`systemctl`, not an rsync, not a file install.

**The address argument is mandatory, and omitting it is the hazard
rather than a stop.** `scripts/deploy.sh:99-109` aborts only when
`HCLOUD_TOKEN` is empty **or** `hcloud` is absent. On this workstation
neither holds: `hcloud` is on `PATH` and `~/.local/secrets` carries
`HCLOUD_TOKEN`, which `deploy.sh:93-96` sources unconditionally. So a
`bash scripts/deploy.sh` typed with no address **resolves the production
box through `hcloud server ip musicbox` and applies the full deploy to
it**. Revision 1 said the opposite. Every command in this plan that runs
the script is therefore quoted in full, address included, and a task
copies the string rather than reconstructing it.

**Read-only `deploy.sh --check` against the box is allowed, and only the
conductor runs it.** The inventory read every branch of `--check` and
proved it makes no box write: the three rsyncs use `-n`, `deploy_file`
and the env step read with `ssh cat`, `run_remote` issues no ssh at all,
and the one extra probe is an informational `test -d` (inventory section
1). The conductor runs it once after Task 5 and records the output
(Resolution R11).

**The implementer's substitute is `--check` against the reserved
documentation address, and that command does issue ssh.** The exact
string, copied verbatim into every task that needs it:

`bash scripts/deploy.sh --check 192.0.2.1`

`192.0.2.1` is TEST-NET-1, unroutable by definition. The run exercises
the render pass and every step's check branch, and **it attempts and
fails six TCP connections on the way**: three `remote_reachable` probes
(`deploy.sh:119-121`, reached from `deploy_file` at `:153`, the env step
at `:301-302`, and the trigger probe at `:239-249`) and three `rsync -e
ssh` calls (`:201`, `:214`, `:277`). Each pays the `ConnectTimeout=10`
set at `:70`, so the whole run costs about a minute. **The reachability
probes print `can't reach root@192.0.2.1 ...` and the three rsync steps
print rsync's own connection errors**; both are expected output, not a
broken edit. The rsyncs are guarded with `|| true` (`:200-202`,
`:214-215`, `:219`), so the run completes. **Never pass the box's address
to that command.**

**The implementer's ssh prohibition is scoped to the box** (Resolution
R15). `.claude/agents/dubplate-implementer.md:35` reads as a blanket ban;
this plan scopes it to `5.78.148.42` and sanctions ssh to the reserved
documentation address, and the dispatch says so in those words. No other
address is sanctioned.

**One precondition the implementer checks before blaming its own edit:**
the render pass resolves all 42 incumbent template rows before any step
and exits 1 listing every miss, so an incomplete `~/.local/secrets`
aborts the run for reasons unrelated to this pass. Task 0's check 11
verifies this before the first dispatch, so an abort here is a BLOCKED
report, never a workaround.

## Spec

`docs/specs/2026-09-05-rung-4-shadow.md` (revision 2), and
behind it `specs/2026-09-02-greenfield-server-build.md` (revision 2.1,
ratified), ladder entry 4. 4b builds deliverable 6 of the spec's six, the
deployment. Deliverables 1 through 5 landed in 4a.

## Token ceiling: 3M

Re-derived from 4a-ii's measured rate, per the spec's instruction that
4b's figure comes from 4a's actual report rather than from the spec's own
proposal of 1.5M (`spec:1662-1672`).

**The basis, read off `docs/HISTORY.md:468-480` rather than asserted.**
4a-ii spent about 2.7M against a 3M ceiling over five tasks: staleness
pre-task plus chain-args extraction 0.26M, the chain 1.32M across 16
agents, two fix and review rounds 0.49M, simplify plus its review 0.18M,
the close drafter 0.16M, and the conductor's own turns about 0.2M. Two
rates come out of that: **0.26M per task including its review** (the
chain alone), and 0.54M per task all-in.

**The all-in rate does not transfer and the plan says so.** 4a-ii's tasks
were novel Go against a corpus, with two fix rounds and a simplify pass
over new packages. 4b's tasks are two unit files, one environment
template, four shell edits, a Makefile variable, about ten lines of Go,
and a comment sweep. Seven tasks at 0.54M would project 3.8M, which is
plainly wrong for this profile. **The chain rate is the honest base.**

The arithmetic:

| Item | Figure | Basis |
| --- | --- | --- |
| Six implementer chains (Tasks 1 to 6) | 1.56M | 0.26M each, 4a-ii's measured chain rate |
| Task 0, the staleness pre-task | 0.15M | inside 4a-ii's 0.26M line, taken alone |
| Task 7, the fresh-context unit read | 0.15M | one zero-context agent, two short files, one diff, one skill |
| Task 8, the handoff drafter | 0.16M | 4a-ii's close drafter, and this one is smaller |
| The conductor's own turns | 0.20M | 4a-ii's measured figure |
| Plan prep: the inventory and this draft | 0.40M | this window's own spend, counted rather than hidden |
| One fix round, allowed for | 0.26M | Task 4 is the one task where a `fix` verdict is plausible |
| **Total** | **2.88M** | |

**The headroom is 4%, and one fix round is half what the basis pass
spent.** 4a-ii spent 0.49M across two fix-and-review rounds
(`docs/HISTORY.md:470-471`); this table allows 0.26M across one. **A
second `fix` verdict on any task therefore crosses the ceiling**, and
that is a planned branch rather than a surprise: it is the conductor's
decision at the 2.4M flag, taken with the task ledger in hand, and the
options are to accept the breach, drop the fix and record the finding for
the sitting, or split the remaining tasks. **Alternative recorded:** raise
the ceiling to 3.25M now. Declined because the projection is a rate
carried from a different task profile, and a stated branch reads more
honestly at the close than headroom bought in advance. **Ruled, conductor
2026-09-07: accepted.** The 3M ceiling stands; a second `fix` verdict on
any task is the conductor's decision at the 2.4M flag.

**Ceiling 3M.** It sits above both the spec's 1.5M and the inventory's
2.5M, and the reasons are stated rather than shaded. The spec's figure
predates 4a's measurement. The inventory's figure excluded the plan prep
and allowed for no fix round. And the pass carries three items the
spec's pass-split row does not name: the Makefile edit, a fifth
provisional mark outside `internal/engine/shadow`, and the two-level
pre-create the grant decision forces (inventory surprises 5, 6, and 1).

**Flag at 80%, 2.4M: finish the task in flight, write STATUS, and one
combined question routes to the conductor**, under Geoff's grant for this
window. No mid-pass ask goes to Geoff; a discovered question is a finding
to report.

**The sitting's own spend sits outside this ceiling.** Task 9 runs in a
later session with its own budget, and this figure covers the unattended
half only.

**Checkpoint: every four tasks**, the default. That puts one checkpoint
after Task 4, which is the end of the `deploy.sh` group and the point
where the fix-round allowance either fired or did not.

## Global Constraints

- **Main tree, one executor.** The lane registry (`docs/STATUS.md:31-36`)
  is the live-executor check, never `pgrep -f Projects/dubplate`, which
  matches every lane once a worktree exists. All three lanes are idle or
  blocked at this plan's writing.
- **Territory, stated as a wall.** This pass writes `systemd/` (two new
  files), `scripts/deploy.sh`, `env/` (one new file plus one header
  amendment), `Makefile` (one variable and one recipe), and
  `cmd/dubplate-server/{main.go,main_test.go}` plus comment-only edits in
  `cmd/dubplate-server/shadow.go` and `internal/engine/shadow/`. It
  writes **nothing** under `corpus/`, `internal/harness/`,
  `internal/config/`, `internal/engine/` beyond three comment lines,
  `cloud-init/`, `config/`, `compose.yaml`, `tools/`, `web/`, or
  `_salvage/`. A task that believes it needs one of those stops and
  reports.
- **`vps-conventions` before the units, `deploy.sh`, or anything that
  places a file or directory on the box.** Read the skill in full. The
  rules that bind this pass: a `ReadWritePaths=` entry must exist on the
  host before the unit starts, or startup fails with "Failed to set up
  mount namespacing"; every service unit including helpers and templates
  carries `NoNewPrivileges`, `PrivateTmp`, and `ProtectSystem=strict`
  with explicit `ReadWritePaths=`; timers carry `RandomizedDelaySec`;
  `Persistent=true` fires missed runs at boot; deploy steps stay
  idempotent with a `--check` branch each; rsync exits 23 on a missing
  source, so a manifest edit lands in the same commit as the files it
  names; a unit's write targets are **argued statically, never assumed**.
- **`go-conventions` before any Go.** Task 5 only. The em dash is banned
  in Go comments with no opt-out, and vale scans a brand-new comment
  before it is staged.
- **Comments carry their reason, on the estate's own precedent.** Every
  unit in `systemd/` explains its own choices: the serve unit spends
  fifteen lines on why it does not take the network-online pairing, every
  timer names the retired `music-*` timer it mirrors and why its jitter
  is sized as it is, and every unit omitting `ReadWritePaths` says in a
  comment why it writes nothing. **A new unit that states a directive
  without its reason is below the estate's bar**, and the fresh-context
  read will say so.
- **Universality.** No estate hostname, path, email, catalog value, or
  personal name outside `internal/config` and the estate's own
  operational record. `env/`, `systemd/`, and `scripts/deploy.sh` **are**
  that operational record (`spec:1429-1434`), so estate paths belong in
  them and nowhere else. The runner itself stays adopter-optional behind
  `DUBPLATE_SHADOW_ENABLED`.
- **Secrets: none minted, and that is a checked statement.** No greenfield
  template row is a secret, `secret-set.sh` is not run, and
  `~/.dotfiles/secrets/registry.md` gains no row. CLAUDE.md's rule that
  "adding a name there requires minting it" reads against the incumbent
  manifest, whose every row is a store lookup. The greenfield template's
  new row state, the literal, is a value that will never be minted
  because it is not a secret (`spec:1282-1291`). **The plan states this
  conclusion out loud rather than leaving the rule to be read as
  requiring a mint.**
- **No value from `~/.local/secrets` is ever printed, logged, committed,
  or quoted in a report.** One task reads one non-secret name from it
  (Resolution R7) and copies it into a tracked file; every other name
  stays a name.
- **The gate is `bash scripts/check.sh`, green at every task end.** Never
  commit red, specific files only. This pass's live legs are
  `systemd-analyze verify` (which picks up both new unit files by glob,
  `check.sh:137`, with no wiring), shellcheck over `scripts/*`, gitleaks
  over the tracked worktree, vale over Go comment prose, and the full
  `make check`. A gate run is about three and a half minutes; that is
  expected, not a hang.
- **The gate proves the units parse and proves nothing about the grant.**
  `systemd-analyze verify` catches a bad section, a misspelled directive,
  and a bad `OnCalendar`. It cannot tell a minimal `ReadWritePaths` from
  a generous one. That gap is exactly what Task 7 exists to close, and no
  task may claim a grant is minimal on the strength of a green gate.
- **`.gitleaksignore` rider.** Its single line pins a finding to
  `internal/config/keys_test.go` at line 66, by line number. No task in
  this pass touches that file, so the rider is a guard rather than a
  task. **No task, comment, or commit message quotes the pinned fixture
  literal.**
- **Module dependencies: none added.** `go.mod` is unchanged by this
  pass.
- **Assertion discipline, at the spec's bar.** "The observation
  mechanism's own honesty gets one mutation proof per checker"
  (`spec:1831-1832`). This pass adds two checkers, the smoke assertion
  and the version verb's test, and both get a named mutation. It also
  adds three artifacts no test can check at all: the two unit files and
  the environment template. **Where no mutation exists, the plan says so
  and names what holds the property instead**, rather than dressing a
  gate pass as a proof.

## Contested surfaces

Lane-launch checklist item 2 requires a written owner for every surface a
lane touches. The rows 4b touches, verified against
`docs/specs/2026-09-03-parallel-build-approach.md`:

**The line numbers are HEAD `c98a916`, and the row names are what to
search on.** That spec carries an uncommitted fifteen-line addition in
this worktree, which moves every row below `:177` by fourteen or fifteen
lines. Match the row's own leading cell, never the number alone.

| Surface this pass writes | Row | Note |
| --- | --- | --- |
| `systemd/`, `scripts/deploy.sh`, `env/` | `PA:184`, the ``systemd/`, `scripts/deploy.sh`, `env/`` row | Names this pass and its exact deliverable count: "rung 4 pass 4b adds two units, one template, and six deploy edits" |
| the state directory layout | `PA:178`, the "the state directory layout" row | "rung 4, first to deploy one, and its unit's `ReadWritePaths` depends on it". Resolution R1 is this row's decision |
| `cmd/dubplate-server` | `PA:182`, the `cmd/dubplate-server` row | **A narrower touch than the row anticipates**: this pass edits `newVersionCmd`'s `RunE`, not the `AddCommand` block, and adds no file |
| `internal/engine/shadow` | `PA:180`, the `internal/engine/shadow` row | Comment text only, three lines in three files, no code path |
| `docs/` ledgers | `PA:177`, the ``docs/` ledgers` row | Conductor-owned; a task drafts and never commits |
| `Makefile` | **no row exists** | Resolution R13 rides the `systemd/`, `scripts/deploy.sh`, `env/` row as the build-and-deploy surface. Ruled, conductor 2026-09-07: accepted |

**No second lane is live.** `docs/STATUS.md:31-36` records lane-1 as IDLE
pending this plan, lane-2 as BLOCKED until Phase 0b and 0c, and the doc
lane as IDLE with one queued candidate. 4b takes the main tree with no
contention, and Task 0 re-verifies that before the first dispatch.

## The 2.0 forward hooks: every task neutral

CLAUDE.md's build-for-2.0 rule requires each task to state whether it
lands a forward hook from
`docs/specs/2026-09-07-streamer-2-0-design.md:354-391` or is
neutral. **All fourteen hooks were read in full, plus a grep of that spec
for `systemd`, `deploy`, `/var/lib`, `state directory`, `ldflags`,
`stamp`, and `environment`. No hook names deploy, systemd, a build stamp,
or an environment file.** The hooks bind rung 6 schema work, rungs 7
through 9, and rung 12, which matches HISTORY forward carry 9.

Two adjacent mentions were checked and neither reaches 4b. The 2.0 spec's
body at `:325` says the streamer's state lives under the existing state
directory, which is a rung 6 and rung 12 consequence of that directory
existing rather than a claim on its path or its grant. And `:240`
mentions an environment file in the context of an app-password key, which
is rung 6's schema work.

**Every task below states "2.0 hook: neutral", and the statement is a
finding rather than a formality.** A task that discovers otherwise
reports it so the conductor can route it into the hooks list.

## Prior art

CLAUDE.md's prior-art rule requires the plan to name which entries of the
2.0 spec's prior-art table (`:401-423`) and its recorded declines
(`:425-428`) it consulted, and to adopt or decline each with a reason.
Both were read in full. **This is a systemd, environment-file, and
shell-script pass; twenty-two of the twenty-three entries are not
applicable, and the plan says so per entry rather than skipping the
section.**

| Entry | 4b's verdict |
| --- | --- |
| OpenSubsonic specification and extension list | Not applicable. 4b serves no API. |
| OpenSubsonic OpenAPI schema and Subsonic 1.16.1 XSD | Not applicable. No harness input in this pass. |
| Navidrome (oracle, ideas and lessons) | Not applicable. 4b deploys no streamer and reads no Navidrome behavior. |
| gonic | Not applicable, same reason. |
| Melodee | Not applicable. No ingest pipeline here. |
| `go.senan.xyz/taglib` | Not applicable. 4b reads no tags; 4a's FLAC reader is untouched. |
| `tetratelabs/wazero` | Not applicable. No WebAssembly. |
| `modernc.org/sqlite` | Not applicable. No database in this pass. |
| `ncruces/go-sqlite3` | Not applicable, same reason. |
| `mewkiz/flac` | Not applicable. No audio decode. |
| Picard tag-mapping tables | Not applicable. No tag mapping. |
| `golang.org/x/image`, `gen2brain/webp` | Not applicable. No images. |
| `golang.org/x/time/rate` | Not applicable. The shadow's pacing is 4a's limiter and is unchanged. |
| `crypto/aes`, `cipher.NewGCM` | Not applicable. Nothing encrypted at rest here. |
| Schema migrations (hand-rolled) | Not applicable. No schema. |
| Stream-through transcode cache | Not applicable. |
| **`internal/atomicfile`** | **Consulted and adopted as-is, unchanged.** Its verdict is "extend it with a streaming write rather than inventing a second temp-file convention". The shadow store already writes through it (the import at `store.go:14`, the first call at `:102`), and 4b neither extends it nor adds a second convention. The connection is real: the unit's `ReadWritePaths` grant is what makes `atomicfile`'s same-directory rename property hold on the box, which 4a declared as an untested gap. |
| M3U read and write; Subsonic XML and JSON dual encoding | Not applicable. |
| Chromaprint via `fpcalc` | Not applicable. No fingerprinting, and 4b adds no subprocess. |
| Symfonium compatibility notes | Not applicable. |
| beets `convert`; MusicBee device sync | Not applicable. |
| `xrash/smetrics` (already declined) | Not applicable. |
| `pressly/goose`, `djherbis/fscache` (already declined) | Not applicable. Named because both declines were recorded against rungs this pass does not touch. |

**One entry outside that table was consulted and adopted, and it is the
one that matters here:** the estate's own units, which
`~/.claude/skills/vps-conventions/SKILL.md` and
`systemd/dubplate-serve.service:96-125` supply as the hardening template.
Copying a working sibling is this pass's whole prior-art story.

## Resolutions this plan takes rather than discovers

Every reading the inputs left open, each with the recommendation applied
and the alternative recorded. **Each is marked "provisional, conductor
ruling owed"**, so the fold at the close is a search rather than a
re-reading, and a reversal is a named edit rather than a rediscovery.

### R1: the grant is the leaf, and the pre-create makes both levels

**Ruled, conductor 2026-09-07: accepted as written.** Inventory surprise 1: the handoff
makes `<state dir>/shadow` the `ReadWritePaths` argument
(`docs/HISTORY.md:389-394`) while the spec's deploy edit 3 pre-creates
`/var/lib/dubplate` (`spec:1318-1322`), and
`~/.claude/skills/vps-conventions/SKILL.md:39-42` makes those mutually
exclusive as written, because a leaf grant needs the leaf pre-created.

**Applied:** the unit grants `ReadWritePaths=/var/lib/dubplate/shadow`,
and the pre-create step makes both levels in one `run_remote`, chowning
both to `music:music`. The unit can never write beside its own store,
`OpenStore`'s own `MkdirAll` becomes a no-op on the box rather than the
thing that creates the root, and the sitting's SELinux check reads both
levels.

**Why the leaf.** `SKILL.md`'s Gate rule asks for a write set argued
statically rather than assumed, and the narrowest set that satisfies the
store's proven footprint is the store's own root.
`TestStoreWritesNothingOutsideItsRoot`
(`internal/engine/shadow/store_test.go:191-249`) is the static argument,
and it proves the leaf and not the parent. The estate's own precedent for
taking a deliberately leaf-only grant is `dubplate-backup.service:32-39`.

**First alternative recorded:** grant the parent and pre-create the
parent only, letting `OpenStore` create `shadow/` on the first pass. One
pre-create line, one wider grant. **The fold if ruled that way:** one
directive value, one `mkdir -p` argument, and one line of the pre-create
comment. Rungs 5 and 6 will want `/var/lib/dubplate` writable for their
own subdirectories, and each will grant its own leaf under the same rule.

**Second alternative recorded, and declined: `StateDirectory=`.**
`~/.claude/skills/vps-conventions/SKILL.md:18` names `StateDirectory=` /
`/var/lib/<x>` as the persistent home in its own lifetimes table, and it
is the one option that dissolves the `:39-42` pre-exist constraint
entirely, because systemd creates every level, owns them as `User=`, and
grants the write implicitly. **Declined for two reasons.** The store's
path is config-driven through `DUBPLATE_STATE_DIR`, which has no default
(`internal/config/config.go:217`), so a `StateDirectory=dubplate/shadow`
directive would hard-code in the unit what the binary reads from its
environment, and the two could diverge silently. And `/var/lib/dubplate`
becomes shared across rungs 5, 6 and 12, where a per-unit
`StateDirectory` owning the tree stops being the right model. The
estate's own precedent agrees: all four existing directory grants use
pre-create plus `ReadWritePaths` (`deploy.sh:336-338`, `:357-359`,
`:370-372`, `:387-389`). **Recorded here so Task 7 grades a considered
decline rather than an omission.**

### R2: five provisional marks retire, and `flac.go:62` stays

**Retired at the 4b close** (`docs/HISTORY.md`'s "The ten retired
resolutions"), because it widened the task past
the four the dispatch brief named. Inventory surprise 5: the brief says
four, and `grep -rni provisional internal/ cmd/` returns six.
**The grep is what establishes the count**, and it must be
case-insensitive: two of the six read "Provisional" with a capital P, so
`grep -rn` sees four. `docs/STATUS.md:47-48` corroborates without
counting marks; it records all six of rung 4's owner rulings ratified and
the retirement mechanical.

**Applied:** five retire and one stays, each judged on its own text.
`internal/engine/shadow/rescore.go:28` and `match.go:20` carry "ruling 3,
provisional" and ruling 3 is ratified. `run.go:388` carries "ruling 2,
provisional" and ruling 2 is ratified. `cmd/dubplate-server/shadow.go:22`
and `:37` carry "Provisional, owner ruling 4 open" and "Provisional,
owner ruling 5 open", and rulings 4 and 5 are ratified.
**`internal/engine/shadow/flac.go:62` stays exactly as written**: it
names no ruling number, its "provisional" describes the reader's lifetime
rather than a pending decision, and `docs/HISTORY.md:442-447` carries the
absorption as forward carry 5. Deleting it would delete the pointer rung
8 needs.

**The ruling citation survives in every retirement.** The word goes; the
"(ruling 3)" stays, because the citation is what tells a reader where the
choice was made.

### R3: the version verb prints two labeled rows

**Retired at the 4b close.** The spec says `dubplate-server
version` prints both rows and cites `main.go:56-64` for the current
one-line print; at HEAD those lines are `newRootCmd`'s command
construction and `newVersionCmd` is `main.go:72-81` (inventory surprise
3). Nothing in the spec or the tree says what the two-row output looks
like, and the smoke assertion has to parse it.

**Applied:** two labeled rows in `doctor.go`'s own `%-32s %s` column
shape, labeled `version` and `matcher-commit`, in that order. The label
gives the smoke assertion a stable field to select on, and the shape
matches the only other table this binary prints. **Nothing parses the
current output**: a grep of the repo finds no test and no caller reading
`version`'s stdout, so no consumer breaks.

**Alternative recorded:** keep the bare version on line one for
compatibility and append an unlabeled second line. **The fold if ruled
that way:** one format string, one test expectation, and the smoke
assertion's field selector.

### R4: the smoke assertion runs inside the single remote command

**Retired at the 4b close.** `run_remote` prints `would run:
<command>` under `--check` and issues the ssh otherwise, which is what
stops a step's check line from drifting from what apply runs
(`deploy.sh:129-137`). An assertion written in the local shell around a
captured `run_remote` output would break that property and would have
nothing to assert under `--check`.

**Applied:** the assertion is part of the one remote command string, so
the check branch prints the whole thing verbatim and the apply branch
runs it. The remote command prints both rows for the operator to read and
exits non-zero when the matcher-commit field is empty or `dev`, which
`set -euo pipefail` turns into a whole-deploy abort. The step keeps its
existing `echo "== ..."` announcement.

**Alternative recorded:** a second step that captures and asserts
locally. **The fold if ruled that way:** one step gains a check branch of
its own and the no-drift property is argued in a comment instead of held
by construction.

### R5: the render is a pass over both templates; the install is a per-template step

**Ruled, conductor 2026-09-07: accepted as written, revised at revision 2.** Deploy
edit 4 lifts the render loop and the install block out of the script's
two blocks. Revision 1 called for one helper doing both, called twice,
while also requiring both renders to finish before the first remote step.
**Those two cannot both hold.** An install is a remote step
(`deploy.sh:314-315`) and the script's first remote step is the repo
config step at `:199-208`, so a combined helper invoked at two script
positions renders its second template after five announced steps and
after the first install. The current script satisfies the property only
because the render (`:171-194`) and the install (`:291-316`) sit five
announced steps apart (`:199`, `:212`, `:230`, `:273`, `:288`).

**Applied: the refactor is two pieces, not one.**

- **A render pass** runs over both templates before `deploy.sh:196`'s
  first announcement, resolving every row into memory and accumulating
  one combined missing list. On any miss it exits 1 naming every missing
  row with the template it came from, before any remote step. That
  preserves the property the current script has and its comment states,
  that a workstation missing a value aborts before touching the box
  (`:190-194`).
- **An install-and-check step** runs at each template's own position,
  parameterized on the rendered lines and the remote path.
  **`/etc/dubplate/env` keeps its position at `:297`**, and the
  `server.env` step lands immediately after it, so the two env steps read
  as a pair and the announced count reaches sixteen.

**The constraints on both pieces, which are acceptance criteria and not
an implementation:** the render stays in memory, so no decrypted value
touches local disk or a log at any point under either branch; each
install step's check branch keeps the existing "diff not shown, it's
secrets" wording; and each install branch keeps `install -D -m 600 -o
root -g root /dev/stdin`, reading straight off stdin with no staged copy
on the box.

**Alternative recorded:** keep one combined helper and accept that the
second render runs mid-script. **The fold if ruled that way:** R5's
first bullet is deleted, and the half-deployed-box failure mode is
accepted and named in the helper's own comment.

### R6: `/etc/dubplate/server.env` is 0600 root:root, exactly like the incumbent's

**Retired at the 4b close.** The unit runs `User=music`, and a
reviewer may read 0600 root:root as a file the service cannot read.

**Applied:** the same mode and ownership as `/etc/dubplate/env`, for the
same reason, **stated in a comment so nobody widens it**: systemd reads
`EnvironmentFile=` as PID 1 before it drops privileges, which is why
every existing `User=music` oneshot already reads a 0600 root:root file
(`dubplate-import.service:15` against `deploy.sh:314-315`). Widening it
to `root:music 0640` would expose the file to the music user for no
gain.

### R7: the masthead literal is copied, never invented

**Retired at the 4b close.** The spec rules `DUBPLATE_SITE_NAME`
a literal carrying "the estate's masthead text" (`spec:1278`), on two
grounds: a bare `NAME=` aborts the render because the age store carries
the pre-rename `DUBPLATE_SERVE_SITE_NAME`, and re-keying the store would
put one string in two places. **The spec does not quote the string, and
neither does this plan.**

**Applied:** the implementer copies the value verbatim from the age
store's existing `DUBPLATE_SERVE_SITE_NAME`, which
`~/.dotfiles/secrets/registry.md` records as non-secret masthead text
minted 2026-09-01, and writes it into the tracked template. It does not
print the value in its report or in the commit message, and it invents no
masthead. **If the name is absent from the store, the task stops and
reports BLOCKED**, because a guessed masthead would ship to the box and
appear on every page rung 7 builds.

### R8: the floor and the budget ship as literals matching their registered defaults

**Retired at the 4b close.** `DUBPLATE_SHADOW_FLOOR` and
`DUBPLATE_SHADOW_BUDGET` are registered thresholds with defaults that
stand when unset (`thresholds.go:55-96`), so the template could omit them
entirely. The spec names them literals (`spec:1280`).

**Applied:** both ship as literals, `720h` and `60`, matching the
registered defaults exactly. The template is the estate's operational
record, and a reader of the box's env file should be able to see the
operative floor and cap without reading Go. **The duplication is real and
the template's own comment names it**: the values live in two places, and
a change to either default must change both.

**Alternative recorded:** omit both rows and let the defaults stand.
**The fold if ruled that way:** two template rows deleted, and the ten-row
count becomes eight.

### R9: stamping `version` changes what the box tells MusicBrainz

**Not a decision, a consequence nobody wrote down.** Inventory surprise
12: `cmd/dubplate-server/shadow.go:215` builds the MusicBrainz User-Agent
as `dubplate-server/%s ( %s )` from `version` and the admin email, so an
unstamped build identifies itself as `dubplate-server/dev` on every
request. Neither the spec nor the handoff mentions it.

**Recorded here so the sitting can observe it**, and so a dirty worktree
at deploy time is understood to ship a `-dirty` User-Agent rather than
merely a cosmetic version string. Task 5's acceptance names it and the
sitting's step 2 checks for it.

### R10: `TIMER_UNITS` keeps its list and gains a sentence

**Retired at the 4b close.** The spec calls edit 2 "unchanged"
(`spec:1311`). Read literally, that is a no-op and the file records
nothing about why the shadow timer is absent.

**Applied:** the list is unchanged and the comment above it gains the
shadow timer's own sentence beside the serve unit's, with its own reason:
the shadow has no exposure, so the reason is not exposure, it is that an
adopter's deploy should not silently start observing and that the enable
is the act that starts the launch clock. Without the sentence, the next
reader has to decide whether the absence is a decision or an oversight,
which is exactly the question the serve unit's own eleven-line comment
already answers for itself.

### R11: the conductor runs `--check`, not the implementer

**Retired at the 4b close.** The exit gate says "`deploy.sh
--check` clean" (`spec:1632`), and the implementer agent is forbidden any
ssh.

**Applied:** the conductor runs `bash scripts/deploy.sh --check
5.78.148.42` once, after Task 5's review accepts and before Task 7's
fresh-context read, and records the output in the ledger. Tasks prove
their own shell edits against the unroutable address instead (Box contact
rule). A `--check` failure at that point is a conductor decision: a fix
task, or a finding recorded for the sitting.

### R12: doctor prints neither stamp, and that stays true

**Declined at the 4b close, no file change**; filed for whichever rung
next touches `cmd/dubplate-server/doctor.go`. Inventory surprise 11:
`doctor.go:24-72` derives four row groups from `internal/config`'s tables
and mentions neither `version` nor `matcherCommit`.

**Applied:** declined. The scope row names only `dubplate-server version`
(`spec:1632`), and adding doctor rows would be an unnamed deliverable in
a pass whose whole discipline is that its surface is small. **Filed as a
note** for whichever rung next touches `doctor.go`: an operator debugging
a refusal reads the stamp from `version`, and a doctor row would save
that hop.

### R13: the `Makefile` rides the deploy row

**Ruled, conductor 2026-09-07: accepted.** Inventory surprise 6: edit 6
touches three files, and no ownership row names `Makefile`.

**Applied:** `Makefile` is read as part of the `systemd/`,
`scripts/deploy.sh`, `env/` surface (`PA:184`), because the edit exists
only to let `deploy.sh` pass values through `build-linux` and has no
independent purpose. **The conductor has written the row** in
`docs/specs/2026-09-03-parallel-build-approach.md`'s section 3
ownership table.

### R14: the spec's stale citations are corrected at the close, never in flight

**Retired at the 4b close**, corrections queued for the conductor's next
spec-touching batch, the spec left unedited by this pass. Five citations
in the spec point
at lines that have moved or at work already landed: the nine-step count
against fourteen announced steps (surprise 2), `main.go:56-64` for the
version verb (surprise 3), the threshold `Kind` work rung 4a already
landed (surprise 4), `dubplate-import.service:14` for the
`EnvironmentFile` line at `:15` (surprise 7), and
`docs/HISTORY.md:2327-2328` for the open `/var/lib/musicbox` SELinux
check that lives at `:3082-3083` (surprise 13).

**Applied:** every task follows the tree and cites the tree. **No task
edits the spec.** Task 8 drafts eight corrections, one line each, for the
conductor's batch: these five, the R1 grant correction, the pass-split
row's understated edit 6, and the handoff item that reads as a grant on
the store root.

### R15: the implementer's ssh prohibition is scoped to the box

**Ruled, conductor 2026-09-07: accepted as written.** The one quoted
`deploy.sh --check` against `192.0.2.1` is allowed, because a plan scoping
an agent
file's hard rule is a decision that outlives this pass.
`.claude/agents/dubplate-implementer.md:35` reads "**The box is
production.** No `ssh`, no deploy, no write to `5.78.148.42`, ever." Read
as a blanket ban on ssh, it forbids the one command Tasks 3, 4 and 5 use
to prove their shell edits, because `deploy.sh --check` issues ssh at six
call sites whatever address it is given.

**Applied:** the prohibition is scoped to `5.78.148.42`, and ssh to
`192.0.2.1` is sanctioned for the exact command the Box contact rule
quotes. The dispatch says so in those words, so the implementer does not
have to reconcile the two itself. No other address is sanctioned, and the
sentence's other three clauses (no deploy, no write, the box is
production) are untouched.

**Alternative recorded:** amend the agent file instead. Declined here
because `.claude/agents/` is outside this pass's territory wall and the
file serves every future pass, so the edit belongs to whoever next owns
that file rather than to a deploy pass. **The fold if ruled that way:**
this resolution becomes a one-line agent-file edit and the Box contact
rule's scoping paragraph is deleted.

### R16: no host-resolution guard is added to `deploy.sh`

**Ruled, conductor 2026-09-07: accepted.** No host-resolution guard is
added to `deploy.sh`. The blocker behind the Box
contact rule's inversion has a mechanical fix available: make `deploy.sh`
refuse to apply when `HOST` was resolved by `hcloud` lookup rather than
passed as an argument, so a bare `bash scripts/deploy.sh` can never reach
the box unannounced.

**Applied: declined for this pass.** It would be a seventh deploy edit
the spec's scope row does not name (`spec:1632`), it changes behavior for
every adopter and every future deploy rather than only for this pass's
exposure, and the pass's own discipline is that its surface stays small
and named. The exposure is closed instead by making the address mandatory
in every quoted command, which costs nothing and holds for exactly this
pass's tasks.

**Recorded as a note for whichever pass next edits the resolution
block**: the lookup path exists for a provisioning flow that has not run
since the box was built, and a guard on it is cheap. **The fold if ruled
the other way:** one task gains one `deploy.sh` deliverable, a flag or a
sentinel distinguishing an argument from a lookup, with its own `--check`
branch and its own mutation proof.

---

## Task 0: The staleness pre-task

Model: `claude-opus-5`, zero context. **Not a chain task**: the conductor
dispatches it before the first implementer dispatch, rules each blocker,
and only then dispatches Task 1. Its output is blockers with evidence,
never fixes, and it commits nothing.

Every rung 4 pass opens with this (`spec:1674-1727`). Plan staleness is
the build's named dominant defect class.

**Files:** none written. Read-only against the tree at HEAD.

**The checks, 4b's subset of the spec's list plus this plan's own
exposure.** The first four are blockers; a "no" stops the launch.

1. **Has rung 1b landed?** The spec's own first blocker
   (`spec:1685-1690`), carried here rather than dropped: 1b changes the
   wall's shape and re-expresses ruling 1's option (a) in its registry's
   own terms. **The exposure is live rather than theoretical.** At this
   plan's writing the tree carries an untracked
   `docs/plans/2026-09-07-rung-1b-config-registry.md` and an
   untracked `docs/research/2026-09-07-rung-1b-planning-inventory.md`,
   so 1b is being planned in parallel. A "yes" means Task 2's ten-row
   template and Task 0's check 9 are both written against a shape that
   has moved, and the conductor rules before Task 1. Check 9 corroborates
   this answer and never substitutes for it.
2. **Is the lane registry consistent with running here, and is the
   primary tree free?** The registry is the check, never `pgrep`.
3. **Is the tree clean at `master` and green?** `bash scripts/check.sh`
   passes before the first task edits anything, so a red gate later is
   this pass's own doing.
4. **Is `.gitleaksignore:1` still
   `internal/config/keys_test.go:generic-api-key:66`, and is line 66
   still the row it names?** No task touches that file, so this is a
   guard rather than a task, but a red gitleaks leg would stop every
   task's commit.
5. Have `systemd/`, `scripts/deploy.sh`, `Makefile`, or
   `env/dubplate.env.template` moved since the inventory was written
   against HEAD `c98a916`? Report any diff.
6. Does `Makefile`'s `build-linux` still set no `-ldflags`, and does
   `dubplate-server version` still print one line? A "yes" to both is the
   expected state and is what makes Task 5 real work.
7. Are `main.version` and `main.matcherCommit` still package-level `var`s
   at `main.go:31` and `:44`, both `"dev"`, with the refusal live at
   `cmd/dubplate-server/shadow.go:203-205` and `:303-305`?
8. Is `StoreDir` still `filepath.Join(stateDir, "shadow")`
   (`internal/engine/shadow/store.go:41-47`), and does
   `TestStoreWritesNothingOutsideItsRoot` still pass? Resolution R1 rests
   on both.
9. Do the six `DUBPLATE_SHADOW_*` rows still register as the inventory's
   section 4 describes, with the three paths required-when-enabled and an
   explicit `DUBPLATE_SHADOW_ENABLED=false` loading clean?
10. Does the incumbent's import timer still fire `*:0/15` and its
    quiescence default still sit at ten minutes? The five-minute shadow
    interval depends on both.
11. **Does the render loop resolve every row of
    `env/dubplate.env.template` against this workstation's own
    environment?** Simulate the loop's `value="${!name:-}"` lookup over
    all 42 rows and report the resolved count plus the names of any
    misses, **values never printed**. This is the precondition Tasks 3, 4
    and 5 each rest on: an incomplete `~/.local/secrets` aborts
    `deploy.sh --check` before any step, and without this check the abort
    surfaces as three separate BLOCKED reports after three dispatches. It
    held when this plan was written, so a miss here is drift.
12. Does `~/.local/secrets` carry `DUBPLATE_SERVE_SITE_NAME` and
    `DUBPLATE_ADMIN_EMAIL`, **by name only, values never printed**?
    Resolution R7 and the greenfield template's one store lookup depend
    on the first and the second respectively. **Report separately whether
    `DUBPLATE_STATE_DIR` is present**, since Task 4's mutations 2 and 3
    both turn on it being exported.
13. Do the five retireable provisional marks still sit where the
    inventory found them, and does `flac.go:62` still read as a rung 8
    handoff rather than a ruling mark? Use `grep -rni`; two of the six
    carry a capital P.
14. **Does the box still match what this plan assumes?**
    (`spec:1722-1725`) Four enabled timers, `dubplate-serve.service` and
    the nudge units shipped and disabled, per `docs/HISTORY.md` and
    `docs/STATUS.md` **read as documents, never by contacting the box**.
    4b is the pass whose apply re-enables `TIMER_UNITS`, so a drifted
    record is a finding the sitting needs.

**Acceptance criteria:**

- Every check answered yes or no with a `file:line` or a command output,
  never a recollection.
- Any drift from the inventory's HEAD reported as a diff, with the
  affected plan task named.
- No file written, no commit, no ssh.

**2.0 hook: neutral.** A read-only staleness check lands no hook.

**Independent:** n/a. Nothing dispatches until its blockers are ruled.

---

## Task 1: The two units, their comments, and the `UNIT_FILES` manifest edit

Model: `sonnet`. Depends on Task 0.

**The manifest edit is in this task by necessity, not by convenience.**
rsync exits 23 on a missing source and aborts the whole deploy, so a
`UNIT_FILES` entry naming a file that does not exist yet breaks every
deploy between the two commits (`SKILL.md:70-74`, `spec:1307-1310`). The
files and the manifest land together or not at all.

**Files:**
- Create: `systemd/dubplate-shadow.service`
- Create: `systemd/dubplate-shadow.timer`
- Modify: `scripts/deploy.sh:56-64` (the `UNIT_FILES` array, two entries)

**Outcomes:**

The service unit is a `Type=oneshot`, `User=music` unit that runs
`/usr/local/bin/dubplate-server shadow observe` under the extended
hardening set, reading `/etc/dubplate/server.env`, writing only the
shadow's own store, and failing into the ping-fail chain.

- `[Unit]`: `Description=`, `OnFailure=dubplate-ping-fail@shadow.service`,
  `After=network-online.target`, `Wants=network-online.target`.
- `[Service]`: `Type=oneshot`, `User=music`,
  `EnvironmentFile=/etc/dubplate/server.env`,
  `ExecStart=/usr/local/bin/dubplate-server shadow observe`,
  `ReadWritePaths=/var/lib/dubplate/shadow` (Resolution R1).
- The hardening set carried verbatim from
  `systemd/dubplate-serve.service:99-125`: `NoNewPrivileges`,
  `PrivateTmp`, `ProtectSystem=strict`, `ProtectHome`,
  `ProtectKernelTunables`, `ProtectKernelModules`,
  `ProtectControlGroups`, `RestrictNamespaces`, `RestrictRealtime`,
  `RestrictSUIDSGID`, `LockPersonality`, `MemoryDenyWriteExecute`,
  `SystemCallFilter=@system-service`, `SystemCallArchitectures=native`,
  empty `AmbientCapabilities` and `CapabilityBoundingSet`, `UMask=0027`.
- `RestrictAddressFamilies=AF_INET AF_INET6 AF_UNIX`, **without
  `AF_NETLINK`**.

The timer is the estate's fourth timer shape, matching
`systemd/dubplate-navidrome-check.timer:8-10` on the three schedule
directives (`OnCalendar=*:0/5`, `RandomizedDelaySec=60`,
`Persistent=true`), naming its own service in an explicit `Unit=`, and
carrying `[Install] WantedBy=timers.target`.

**Five comments the units must carry, each with its citation.** The
estate's own bar is that a directive states its reason; a bare directive
set fails this task even with a green gate.

1. **Why the network-online pairing is taken.** Five sibling oneshots
   carry it and the serve unit spends fifteen lines saying why it does
   not (`dubplate-serve.service:12-26`), so the pairing is a decision in
   this estate rather than a default. The shadow's reason: a `*:0/5`
   timer otherwise fires before the network is up after a reboot and
   turns a boot into a failure ping.
   `systemd/dubplate-disk-check.service:11-12` is the precedent for
   taking `network-online` **without** `docker.service`, which is this
   unit's shape, because the shadow reaches MusicBrainz and touches no
   container.
2. **Why `AF_NETLINK` is excluded.** The serve unit's own comment
   (`:116-124`) records that netlink joined its set for a CPython
   subprocess whose glibc `getaddrinfo` opens a `NETLINK_ROUTE` socket,
   and that "The Go process itself never needed it, since Go's own
   resolver has its own fallback." The shadow runs no subprocess, so the
   family stays denied.
3. **Why the grant is one leaf path.** Name
   `TestStoreWritesNothingOutsideItsRoot` as the static argument, name
   the pre-create step in `deploy.sh` as the thing that must run first,
   and name the failure mode by its message, "Failed to set up mount
   namespacing".
4. **The operator rule** (HISTORY handoff item 5, `spec:1387-1391`): a
   hand-run shadow verb goes through `systemctl start
   dubplate-shadow.service` or `systemd-run` carrying the same
   properties, **never the bare binary as root**, because the read-only
   guarantee is the unit's and not the binary's.
5. **That `OnFailure=` is wired and inert, with the mechanism.** No
   `DUBPLATE_PING_SHADOW_URL` is minted, the whole `DUBPLATE_PING_*`
   class is reserved for rung 9 (`internal/config/keys.go:116-125`), and
   an unset URL no-ops rather than erroring, which
   `systemd/dubplate-serve.service:28-34` is the estate's precedent for.
   **State the mechanism precisely** (inventory surprise 8): the helper
   reads `EnvironmentFile=/etc/dubplate/env` at
   `dubplate-ping-fail@.service:11`, the incumbent's file, and execs the
   frozen `/usr/local/bin/dubplate` at `:20`, so a URL minted at rung 9
   would have to land in the incumbent manifest and could never live in
   the greenfield template. **Say plainly what this buys: no monitoring
   at rung 4.** The stamp assertion at deploy and the journal check at
   the sitting are what catch a shadow that observes nothing.

The timer's comment states three things. **The five-minute interval's
reason**, with its citation, the incumbent's quiescence window, rather
than leaving a reader to infer it. **Why `Persistent=true`**: a missed
pass is not "at most a later capture", because the incumbent sweeps its
run reports to the newest thirty and fires `*:0/15`, so the oracle signal
for every album imported during an outage longer than roughly seven and a
half hours is gone permanently (`spec:1208-1216`). **And why a
five-minute interval cannot overlap its own predecessor**, which is
HISTORY handoff item 4 (`docs/HISTORY.md:411-415`): the wire cap is sixty
tokens and the pace is `2 * musicbrainz.DefaultInterval`, so sixty tokens
at two seconds a token bounds one pass at two minutes, comfortably inside
the timer's window. Without that sentence a reader has no way to know the
interval was chosen against a measured ceiling rather than picked.

**Acceptance criteria:**

- Both files exist, are picked up by `check.sh:137`'s glob with no
  wiring, and `systemd-analyze verify` reports nothing beyond the
  expected missing-executable line the gate filters.
- `ReadWritePaths=` names exactly one path, the store's leaf root.
- `RestrictAddressFamilies=` omits `AF_NETLINK` and the omission carries
  its reason.
- The hardening set matches `dubplate-serve.service:99-125` directive for
  directive, minus `AF_NETLINK`. A reviewer can diff the two blocks.
- The timer's `OnCalendar`, `RandomizedDelaySec` and `Persistent` match
  `dubplate-navidrome-check.timer:8-10` exactly, and its `Unit=` names
  this pass's own service. **The fourth directive cannot match the
  sibling and must not**, which is why the criterion names three.
- All five comments are present with their citations, and every citation
  resolves to the line it names at HEAD.
- The timer's comment carries all three of its required statements, the
  pass-duration bound included.
- `UNIT_FILES` carries both new names and the array is otherwise
  byte-identical.
- `bash scripts/check.sh` green.

**Mutation proofs this task reports:**

1. Misspell one hardening directive, for example `ProtectSystm=strict`,
   and run `bash scripts/check.sh` → the systemd leg fails on that line,
   proving the new files are inside the gate's glob rather than merely
   present. Restore and prove restored.
2. Change `OnCalendar=*:0/5` to a syntactically invalid value → the same
   leg fails, proving the timer is verified too.
3. **A mutation that deliberately does not fire, and the plan says so.**
   Widen `ReadWritePaths=` to `/var/lib` and run the gate → **it stays
   green**. `systemd-analyze verify` cannot tell a minimal grant from a
   generous one. Report the green result as a finding, not a pass. **What
   holds the property instead** is the static argument in the unit's own
   comment plus Task 7's fresh-context read.

**2.0 hook: neutral.** No hook names systemd, a unit, or a timer; the
fourteen were read in full.

**Independent: no.** It writes `scripts/deploy.sh`, which Tasks 3, 4 and 5
also write.

**Deliverable count: 3** (the service unit, the timer, the manifest
edit).

---

## Task 2: The greenfield environment template and the incumbent header amendment

Model: `sonnet`. Depends on Task 0.

Two deliverables that are one decision: a new manifest for the greenfield
binary's wall, and one sentence in the incumbent's header naming its
sibling. Both files coexist until rung 16.

**Files:**
- Create: `env/dubplate-server.env.template`
- Modify: `env/dubplate.env.template:1-3` (the header's first block)

**Outcomes:**

The new template carries **exactly ten rows**: four owned wall rows and
six `DUBPLATE_SHADOW_*` rows. Nine are literals and one is a store
lookup.

| Row | State | Value |
| --- | --- | --- |
| `DUBPLATE_LIBRARY_ROOT` | literal | `/srv/music/library` |
| `DUBPLATE_STATE_DIR` | literal | `/var/lib/dubplate` |
| `DUBPLATE_SITE_NAME` | literal | copied from the store, Resolution R7 |
| `DUBPLATE_ADMIN_EMAIL` | store lookup, bare `NAME=` | already minted |
| `DUBPLATE_SHADOW_ENABLED` | literal | `true` |
| `DUBPLATE_SHADOW_OBSERVE_DIR` | literal | `/srv/music/inbox` |
| `DUBPLATE_SHADOW_ORACLE_STATE_DIR` | literal | `/var/lib/musicbox` |
| `DUBPLATE_SHADOW_REVIEW_DIR` | literal | `/srv/music/review` |
| `DUBPLATE_SHADOW_FLOOR` | literal | `720h`, Resolution R8 |
| `DUBPLATE_SHADOW_BUDGET` | literal | `60`, Resolution R8 |

**The header of the new file names all three row states**, so a later
rung does not re-derive them (`spec:1339-1348`): a bare `NAME=` resolves
from the workstation age store or the render aborts; a `NAME=value` line
is a **literal**, a value that will never be minted because it is not a
secret; and a commented pending row is a name whose value arrives later,
on the `config.WallRow.Pending` precedent. It also states the no-mint
conclusion explicitly, with the rename as the reason.

**The comments each row group owes, each with its reason:**

- `DUBPLATE_STATE_DIR` is `/var/lib/dubplate` and **not** the incumbent's
  `/var/lib/musicbox`, because a shared directory would grant the shadow
  unit write over the incumbent's lock, tag logs, and run reports, which
  is exactly the grant the read-only guarantee exists to withhold
  (`spec:1226-1237`). It is a literal rather than a bare row because
  `DUBPLATE_STATE_DIR` is already exported into `deploy.sh`'s one flat
  shell namespace, carrying the incumbent's value, and a bare row would
  silently inherit it. **The mechanism, stated precisely so nobody hunts
  for it in the wrong place**: the incumbent template's own bare
  `DUBPLATE_STATE_DIR=` row forces the name into `~/.local/secrets`, and
  `deploy.sh:93-96` sources that file before the render runs, so the
  indirect expansion at `:182` finds a value. The render loop itself
  exports nothing; it appends to an array.
- `DUBPLATE_LIBRARY_ROOT` and `DUBPLATE_SITE_NAME` are literals because
  the age store carries their pre-rename names, `DUBPLATE_LIBRARY_DIR`
  and `DUBPLATE_SERVE_SITE_NAME` (`internal/config/keys.go:165-166`), so
  a bare row aborts the render. **`DUBPLATE_LIBRARY_ROOT` exists to
  satisfy `config.Load`'s wall (`config.go:216`) and is read by nothing
  in `internal/engine/shadow`**, which a grep confirms (inventory
  surprise 10). Say so, so a later reader does not hunt for the consumer.
- `DUBPLATE_SHADOW_ENABLED=true` is load-bearing in a way the others are
  not: with the flag false or absent, `cfg.Shadow` stays nil and `shadow
  observe` refuses naming the flag. And **a `Required` or `Optional` row
  set beside a false or absent flag is a load error naming the flag**
  (`config.go:190-192`), so the six rows only ever ship together. An
  adopter turning the integration off sets the flag to an explicit
  `false` and removes the other five; that path loads clean and is tested
  (`config_test.go:307-310`).
- The three path rows are required-when-enabled, and each names what the
  path is: the inbox the incumbent consumes, where the incumbent's
  `report-*.json` and `import-*.log` live, and where a skipped album is
  filed.
- The floor and the budget name their registered defaults and their
  duplication (Resolution R8).

**The incumbent header amendment** is one edit to
`env/dubplate.env.template:1-3`, whose sentence already anticipates this
rung ("this manifest evolves at rung 4"). It names the sibling file, says
which binary reads which, and states that the greenfield binary cannot
read this one: of the 39 `DUBPLATE_*` rows here, two are owned rows the
greenfield wall accepts and every other is a reserved, superseded, or
unknown name that `rejectNamespace` turns into a load error
(`config.go:173-204`), while two rows the greenfield binary requires are
absent entirely, having been renamed. **Change nothing else in the
header**; the render contract, the no-real-values sentence, and the
pre-pivot `Load` semantics all stay as written.

**Acceptance criteria:**

- The new template has exactly ten assignment rows, and a grep for a
  seventh `DUBPLATE_SHADOW_*` or an eleventh row returns nothing.
- Nine rows carry a right-hand value and exactly one,
  `DUBPLATE_ADMIN_EMAIL`, is bare.
- No value in the file is a secret, and no name is added to
  `~/.dotfiles/secrets/registry.md`. The report states this as a checked
  fact.
- `DUBPLATE_SITE_NAME`'s value is copied from the store and appears
  nowhere in the report or the commit message. **The evidence is an exit
  status, never the value**: the implementer compares the template's
  rendered value against the store's `DUBPLATE_SERVE_SITE_NAME` and
  reports only whether the comparison succeeded, and the reviewer re-runs
  the same comparison rather than eyeballing the literal in the diff.
  Without that, the criterion is unverifiable from a diff and R7's
  guessed-masthead failure mode is undetectable.
- The header names all three row states and states the no-mint
  conclusion.
- The incumbent header's first block names the sibling; the other three
  blocks are byte-identical.
- `bash scripts/check.sh` green, including the gitleaks leg over the new
  tracked file.

**Mutation proofs this task reports:**

1. **The one real checker in reach is gitleaks.** Nothing in the gate
   validates a template's row count or its values, so the plan does not
   pretend otherwise. What is proven: the file is inside the gitleaks
   sweep. Temporarily replace one literal with a string shaped like a
   credential and run the gate → gitleaks reds. Restore, prove restored,
   and **never commit the mutated line**.
2. **A deliberately unfired one, recorded as a gap.** A wrong path value,
   for example `/srv/music/inbox` mistyped, is caught by no test in this
   repo. **What holds it instead** is the conductor's `--check` run
   (Resolution R11), Task 7's read, and the sitting's step 7 journal
   check, where a wrong observe directory shows up as a pass that
   observes nothing.

**2.0 hook: neutral.** The 2.0 spec's one environment-file mention
(`:240`) is a rung 6 app-password key, not this template.

**Independent of Tasks 1, 3, 5 and 6; Task 4 depends on it.** Its file
territory is `env/` alone, which no other task writes, so the parallel
mode's territory check is satisfied. **Ordering is a different
constraint**: Task 4's second install step names
`env/dubplate-server.env.template`, and the render loop reads that file
directly (`deploy.sh:175-194`) under `set -euo pipefail`, so running Task
2 after Task 4 aborts every `deploy.sh --check` and every apply while the
gate stays green, because `scripts/check.sh` never invokes `deploy.sh`.
Task 2 runs before Task 4.

**Deliverable count: 2** (the new template, the header amendment).

---

## Task 3: The pre-create step and the timer-list comment

Model: `sonnet`. Depends on Task 1 (same file).

Deploy edits 2 and 3, together because they are the two `deploy.sh`
changes that place nothing new in the render path.

**Files:**
- Modify: `scripts/deploy.sh` (a new directory step, and the comment
  above `TIMER_UNITS` at `:396-407`)

**Outcomes:**

**The pre-create step** makes `/var/lib/dubplate` and
`/var/lib/dubplate/shadow` and chowns both to `music:music`, in one
`run_remote` call, in the shape of the existing `/var/lib/musicbox` step
at `:357-359`. It is placed beside the other three directory steps, not
at the end of the script, so the four read as one group. Its comment
follows `:350-356`'s own shape and states four things: that
`dubplate-shadow.service` names the leaf in `ReadWritePaths` and systemd
requires it to exist before the unit starts or startup fails with "Failed
to set up mount namespacing"; that the unit's own `OpenStore` cannot
create it, because startup fails before the process runs; that both
levels are made because the grant is the leaf (Resolution R1); and that
`music:music` matches the unit's `User=`.

**The `TIMER_UNITS` comment** gains the shadow timer's own sentence
beside the serve unit's, per Resolution R10. `TIMER_UNITS` itself is
unchanged and the report says so explicitly, so a reviewer does not go
looking for a missing edit.

**Acceptance criteria:**

- The new step announces itself with an `echo "== ... =="` line like every
  other step, bringing the announced count from fourteen to fifteen.
- It is a single `run_remote` call, so its `--check` line and its apply
  command cannot drift.
- Under `--check` it prints `would run: mkdir -p ...` and issues no ssh.
- `TIMER_UNITS`'s value is byte-identical, and its comment names the
  shadow timer with its own reason.
- `shellcheck` clean via `bash scripts/check.sh`.
- `bash scripts/deploy.sh --check 192.0.2.1` runs to completion, prints
  the new step's `would run:` line, and contacts no box.

**Mutation proofs this task reports:**

1. Introduce an unquoted variable in the new step → shellcheck reds the
   gate, proving the new lines are inside the shellcheck leg. Restore and
   prove restored.
2. Replace the `run_remote` call with a bare `ssh` → the `bash
   scripts/deploy.sh --check 192.0.2.1` run stops printing `would run:`
   and attempts a connection, proving the check branch is held by
   `run_remote` rather than by the step's own text. Restore and prove restored. **Run this
   against the unroutable address only.**

**2.0 hook: neutral.**

**Independent: no.** Same file as Tasks 1, 4 and 5.

**Deliverable count: 2** (the pre-create step, the timer comment).

---

## Task 4: The render pass, the install step, and the literal column

Model: `sonnet`. Depends on Task 3 (same file) **and on Task 2**, whose
`env/dubplate-server.env.template` this task's second install step and
all three of its mutations read. **The one task in this pass where a
`fix` verdict is plausible**, because it rewrites a block whose
in-memory-only property is the reason the block exists.

Deploy edits 4 and 5, together because they are one refactor of the same
two blocks.

**Files:**
- Modify: `scripts/deploy.sh:171-194` (the render loop),
  `scripts/deploy.sh:291-316` (the env install block), and the call sites

**Outcomes:**

**The refactor is two pieces, not one helper doing both** (Resolution
R5, revised at revision 2). A single render-and-install helper called at
two script positions cannot satisfy the render-before-any-remote-step
property, because an install is itself a remote step
(`deploy.sh:314-315`) and the first remote step is the repo config step
at `:199-208`.

- **A render pass over both templates**, running before `deploy.sh:196`'s
  first announcement. It resolves `env/dubplate.env.template` and
  `env/dubplate-server.env.template` into memory, accumulates **one**
  combined missing list, and on any miss exits 1 naming every missing row
  with the template it came from, before any ssh.
- **An install-and-check step per template**, parameterized on the
  rendered lines and the remote path.
  **`/etc/dubplate/env` keeps its position at `:297`** and the
  `/etc/dubplate/server.env` step lands immediately after it.

**The properties both pieces must keep, all of them load-bearing:**

- **In memory only.** No decrypted value touches local disk or a log, at
  any point, under either branch. Each install branch keeps `install -D
  -m 600 -o root -g root /dev/stdin` reading straight off stdin, with no
  staged copy on the box (Resolution R6).
- **The check branch shows no content.** It keeps the existing wording:
  `no changes (contents match; diff not shown, it's secrets)`, `would
  change (contents differ; diff not shown, it's secrets)`, and `can't
  reach <target> to diff; would write N variables` when
  `remote_reachable` fails.
- **Each step announces its own remote path and its own variable count**,
  so the operator reading `--check` sees two distinct env steps rather
  than one.

**The literal column** is the third row state. A template line carrying a
value on its right-hand side is used verbatim; a bare `NAME=` must still
resolve from the environment or be reported missing. It is backward
compatible by construction: no tracked line in the incumbent template
carries a right-hand value, verified by grep, and the loop reads only the
name today (`:181-182`). Both template headers name all three states,
which Task 2 already wrote for the greenfield file.

**One edge the plan names rather than leaving to discovery:** a literal
value may itself contain `=`. The name is everything before the first
`=`, and the literal is everything after it. State this in the loop's own
comment.

**Acceptance criteria:**

- One render pass over both templates, one parameterized install step
  with two call sites, and no remaining copy of the original render loop
  or install block.
- Both templates render before `deploy.sh:196`'s first announcement, and
  `/etc/dubplate/env`'s install step is still at `:297` with the
  `server.env` step immediately after it.
- A bare `NAME=` with no exported value still lands in the missing list
  and still exits 1 before any remote step.
- A `NAME=value` line renders verbatim and is never looked up in the
  environment, **even when a variable of that name is exported**. That
  last clause is the one that matters: `DUBPLATE_STATE_DIR` is present in
  `~/.local/secrets`, which `deploy.sh:93-96` sources into the script's
  one flat namespace, so the indirect expansion at `:182` would resolve
  it to the incumbent's `/var/lib/musicbox`. The greenfield literal must
  win.
- A literal containing `=` keeps everything after the first `=`.
- The announced step count reaches sixteen, both env steps announcing
  their own path and count.
- `bash scripts/deploy.sh --check 192.0.2.1` prints both env steps, each
  with its own variable count, the greenfield one reading `would write 10
  variables`. **That wording is correct here and only here**, because
  `192.0.2.1` is unreachable; against a live box the same step prints
  `would change (contents differ; ...)`.
- `shellcheck` clean; `bash scripts/check.sh` green.

**Mutation proofs this task reports:**

1. Add a bogus bare row to `env/dubplate-server.env.template`, a name no
   store carries, and run `bash scripts/deploy.sh --check 192.0.2.1` →
   the render exits 1 naming that row and its template, before any step
   prints. Restore and prove restored. **Unsetting the name in a subshell
   does not work and is not the mutation**: `deploy.sh:93-96` re-sources
   `~/.local/secrets` before the render runs and restores anything the
   calling shell unset, so that version of the proof reports green and
   proves nothing.
2. Change the greenfield template's `DUBPLATE_STATE_DIR` literal to a
   bare `NAME=` and run `bash scripts/deploy.sh --check 192.0.2.1` →
   **the rendered value becomes the incumbent's `/var/lib/musicbox`,
   silently, with no error**,
   which is precisely the failure the literal column exists to prevent.
   Report the rendered count and the observed behavior, restore, and
   prove restored. This is the pass's most valuable mutation: it
   demonstrates the flat-namespace hazard the spec argues from.
3. Remove the literal branch entirely and run `bash scripts/deploy.sh
   --check 192.0.2.1` →
   **eight of the nine literal rows are reported missing, and
   `DUBPLATE_STATE_DIR` is not among them**, because it alone resolves
   from `~/.local/secrets` to the incumbent's value. **The silent ninth is
   the point of the mutation, not an anomaly in it**: it is mutation 2's
   hazard reproduced without touching the template. A nine-name failure
   list would mean the literal branch is still in force somewhere.
   Restore and prove restored.

**2.0 hook: neutral.**

**Independent: no.** Same file as Tasks 1, 3 and 5, and it also depends
on Task 2's template existing.

**Deliverable count: 2** (the render pass and the parameterized install
step with its second call site, the literal column).

---

## Task 5: The two stamps, the Makefile plumbing, the version verb, and the smoke assertion

Model: `sonnet`. Depends on Task 4 (same file).

Deploy edit 6, which the pass-split row describes as one file and which
the tree says is three (inventory surprise 6). **Without this task the
runner's own refusal is unobservable for thirty days**: an unstamped
binary deploys clean, every timer pass refuses, and the `OnFailure=`
chain fires into a slug whose URL is unset and therefore no-ops (HISTORY
handoff item 2).

**Files:**
- Modify: `Makefile:6-7` (a flags variable the `build-linux` recipe
  consumes)
- Modify: `scripts/deploy.sh:273-289` (compute the two values, pass them
  to `make build-linux`, assert in the smoke step)
- Modify: `cmd/dubplate-server/main.go:72-81` (`newVersionCmd`'s `RunE`)
- Modify: `cmd/dubplate-server/main_test.go` (append)

**Outcomes:**

`deploy.sh` computes `git describe --tags --always --dirty` and `git
rev-parse HEAD` against `$REPO_ROOT`, and passes both through `make
build-linux` into `-X main.version=` and `-X main.matcherCommit=` against
package `main` in `./cmd/dubplate-server`. The `Makefile` gains a
variable the recipe consumes, **defaulting to empty**, so a bare `make
build-linux` from a workstation still produces a `dev`-stamped binary and
the refusal keeps firing for workstation builds, which is exactly what
`matcherCommit`'s own comment documents (`main.go:33-43`). The `build`
and `install` targets are unchanged.

`dubplate-server version` prints two labeled rows, `version` and
`matcher-commit`, in `doctor.go`'s `%-32s %s` column shape (Resolution
R3).

The smoke step at `:288-289` gains its assertion, inside the single
`run_remote` command string (Resolution R4). It prints both rows for the
operator and exits non-zero when the matcher-commit field is empty or
`dev`, which `set -euo pipefail` turns into a whole-deploy abort. Under
`--check` the whole command prints as a `would run:` line, unchanged in
shape.

**One consequence the task records rather than discovers** (Resolution
R9): `cmd/dubplate-server/shadow.go:215` builds the MusicBrainz
User-Agent from `version`, so stamping it changes what the box sends on
every request, from `dubplate-server/dev` to the describe string. A dirty
worktree at deploy time therefore ships a `-dirty` User-Agent. Note it in
the smoke step's comment.

**Acceptance criteria:**

- `make build-linux` with no arguments still builds and still stamps
  `dev` for both values; a test or a direct `strings` check proves it.
- `make build-linux` with the flags variable set stamps both. The proof
  is direct: the workstation is Linux and the cross-compiled artifact is
  a `linux/amd64` binary, so the implementer runs
  `./dubplate-server-linux-amd64 version` locally and asserts both rows
  carry the passed values.
- `dubplate-server version` prints exactly two rows, labeled and in the
  stated order, and its test asserts both by setting the package
  variables and restoring them.
- The smoke assertion lives inside one `run_remote` string; the `--check`
  output shows the whole command on a `would run:` line.
- `bash scripts/deploy.sh --check 192.0.2.1` runs `make build-linux` with
  the flags, and the resulting local binary prints two non-`dev` rows.
- `go vet`, `golangci-lint`, and vale clean on the new Go comment;
  `shellcheck` clean; `bash scripts/check.sh` green.

**Mutation proofs this task reports:**

1. Drop the `matcher-commit` row from `newVersionCmd`'s output → the new
   test fails on the missing row. Restore and prove restored.
2. Have `newVersionCmd` print `matcherCommit` where `version` belongs and
   vice versa → the test fails on both rows, proving the assertion reads
   the values rather than the labels. Restore and prove restored.
3. Remove the flags variable from the `build-linux` recipe, then run
   `bash scripts/deploy.sh --check 192.0.2.1` and run the built binary →
   it prints `dev` for
   both rows, which is what the smoke assertion would abort on at a real
   deploy. **This is the assertion's own fires-proof**, taken locally
   because the assertion itself can only run against the box. Restore and
   prove restored.
4. Set the flags variable to stamp `matcherCommit=dev` explicitly and run
   the built binary → the row prints `dev`, confirming the assertion's
   target condition is reachable and not merely theoretical.

**2.0 hook: neutral.** No hook names a build stamp or a version surface;
the fourteen were read in full and the spec was grepped for `ldflags` and
`stamp`.

**Independent: no.** Same file as Tasks 1, 3 and 4.

**Deliverable count: 4** (the Makefile variable, the deploy computation
and pass-through, the two-row version verb with its test, the smoke
assertion).

---

## Task 6: Retire the five ratified provisional marks

Model: `sonnet`. Depends on Task 0. **Mechanical, comment text only.**

Five edits across four files, each retiring one word and keeping its
ruling citation, plus one comment that must survive untouched because
rung 8 reads it (Resolution R2).

**Files:**
- Modify: `internal/engine/shadow/rescore.go:28`
- Modify: `internal/engine/shadow/match.go:20`
- Modify: `internal/engine/shadow/run.go:388`
- Modify: `cmd/dubplate-server/shadow.go:22` and `:37`
- **Do not touch:** `internal/engine/shadow/flac.go:62`

**Outcomes:**

Each of the five loses the word "provisional" or the clause "owner ruling
N open" and keeps the ruling number, because the citation is what tells a
reader where the choice was made. `rescore.go:28` and `match.go:20` carry
ruling 3; `run.go:388` carries ruling 2; `shadow.go:22` carries ruling 4
and `:37` carries ruling 5. All five rulings were ratified as recommended
on 2026-09-07. Nothing else in any of the five paragraphs changes; the
arguments they state are what the 4a contracts quote.

`flac.go:62` stays exactly as written. It names no ruling number, its
"provisional" describes the reader's lifetime rather than a pending
decision, and `docs/HISTORY.md:442-447` carries the absorption as forward
carry 5.

**Acceptance criteria:**

- `grep -rni provisional internal/ cmd/` returns exactly one line,
  `internal/engine/shadow/flac.go:62`, and that line is byte-identical to
  its state at HEAD.
- Each of the five paragraphs still names its ruling number.
- No code path, no test, and no signature changes. `git diff --stat`
  shows comment lines only.
- Vale clean on all five edited comments; no em dash introduced.
- `bash scripts/check.sh` green, `go test ./...` unchanged in pass count.

**Mutation proofs this task reports:**

1. **This task changes no behavior, so it names no behavioral mutation
   and the plan says so rather than inventing one.** What is proven
   instead: introduce an em dash into one edited comment and run `bash
   scripts/check.sh` → the vale leg reds, proving the edited comments are
   inside the prose gate. Restore and prove restored.
2. Confirm by `git diff` that no non-comment line moved, and report the
   diff's line count. A behavioral change in this task is a finding.

**2.0 hook: neutral.**

**Independent: yes.** Its five files are disjoint from every other task's
except `cmd/dubplate-server/shadow.go`, which no other task in this pass
writes (Task 5 writes `main.go` and `main_test.go`). The conductor may
run it at any point after Task 0.

**Deliverable count: 1** (the five retirements).

---

## Task 7: The fresh-context unit read

**Conductor-dispatched, not a chain task.** Model: `claude-opus-5`, zero
context, no implementer, no commit. It is the first of the spec's two
named review exceptions for rung 4 (`spec:1821-1830`), and it exists
because `systemd-analyze verify` proves the units parse and proves
nothing about whether the grant is the minimum.

**Runs after Tasks 1 through 6 have landed and after the conductor's own
`deploy.sh --check` (Resolution R11), and before Task 8.** It runs before
any box write, so it belongs on the unattended side of the boundary.

**What the conductor dispatches.** One agent with no prior context,
given:

- `systemd/dubplate-shadow.service` and `systemd/dubplate-shadow.timer`,
  in full.
- Every `deploy.sh` diff this pass produced, plus `Makefile`'s.
- `~/.claude/skills/vps-conventions/SKILL.md`, in full, as the standard.
- `systemd/dubplate-serve.service` as the hardening template, and
  `systemd/dubplate-disk-check.service` and
  `systemd/dubplate-navidrome-check.timer` as the oneshot and timer
  shapes.
- `internal/engine/shadow/store.go` and its `store_test.go` walk test, as
  the static argument the grant rests on.

**The questions it answers, each with evidence:**

1. Is `ReadWritePaths=/var/lib/dubplate/shadow` the **minimum** grant for
   what the code actually writes, argued from the store's own footprint
   rather than assumed?
2. Does anything the unit runs write outside that path, read a path
   `ProtectSystem=strict` or `ProtectHome=yes` would hide, or need a
   directive the set does not carry?
3. Is the hardening set complete against
   `dubplate-serve.service:99-125`, and is the `AF_NETLINK` exclusion
   correct for a process that runs no subprocess?
4. Does every `ReadWritePaths` entry exist on the host before the unit
   starts, given the pre-create step as written?
5. Is the timer's `Persistent=true` plus `RandomizedDelaySec=60`
   combination right for a `*:0/5` schedule on a box that reboots?
6. Do the `deploy.sh` edits keep idempotence and a `--check` branch per
   step, stage nothing through a fixed `/tmp` name, and keep the
   in-memory-only property for both env renders?
7. Does any comment state a directive without its reason?

**What the implementer tasks leave for it.** Nothing hidden and nothing
half-done: all six deploy edits landed, both units complete with their
five comments, and the conductor's `--check` output recorded so the agent
reads what the script actually prints rather than predicting it. **No
task may pre-empt this read by asserting its own grant is minimal**; the
tasks state the argument and this agent grades it.

**What happens to its findings.** They route to the conductor, who either
dispatches a fix task through the normal chain or records the finding
with a reason. **A finding about the grant is not optional**: the sitting
applies whatever this read leaves standing, and a wrong grant is a
production write the rollback cannot undo.

**2.0 hook: neutral.**

**Independent:** n/a. It is a gate, not a task, and it runs once all six
tasks have landed.

---

## Task 8: The handoff to the sitting

Model: `sonnet`. Depends on Tasks 0 through 7. **This is the boundary
task. It leaves the pass OPEN.**

**Files:**
- Modify: `docs/STATUS.md` (drafted for the conductor's batch commit)
- Draft, handed to the conductor in the task return: the spec
  corrections, the filed notes, and the provisional-resolution list

**What it does not do**, stated because the close ritual's shape invites
it: it writes no `docs/HISTORY.md` entry, scores neither budget, records
no deploy date, and claims no pass close. Those figures do not exist
until the sitting has run, and a HISTORY entry written now would have to
be rewritten with the numbers that matter.

**Behavior contract:**

- **STATUS's next action is the sitting, written as the checklist the
  sitting session runs**, in order, with the exact commands: the
  quiet-box confirmation, the clean-tree check, `bash scripts/deploy.sh
  --check 5.78.148.42`, the apply as `bash scripts/deploy.sh
  5.78.148.42` with its eleven re-run steps named, `docker compose ps`
  plus the Navidrome check, `systemctl enable --now
  dubplate-shadow.timer`, the journal check, both `ls -Zd` verifications,
  `df -h /`, and the two dates. **Every `deploy.sh` invocation carries
  its address**, because a bare one resolves the box through `hcloud` and
  applies. A fresh session must be able to run it without reconstructing
  anything.
- **STATUS records what is ready**, item by item: both units in the tree
  and in `UNIT_FILES`, the ten-row template, all six deploy edits, both
  stamps reaching the binary, the conductor's `--check` output, and Task
  7's findings with their disposition.
- **STATUS stays present tense and under 60 lines.** Anything historical
  moves to HISTORY at the sitting's close, not now. The lane registry
  entry is updated to record lane-1 as holding an open pass rather than
  idle.
- **The sixteen resolutions are listed with their markers**, fifteen of
  them provisional and R9 a recorded consequence rather than a decision,
  each naming the file and the line its application sits on, so the
  conductor's fold is a search. R1's grant decision is named first,
  because it is the one the sitting physically applies.
- **The drafted spec corrections, one line each** for the conductor's
  batch, none of them edited by this task:
  - `spec:1304-1306`, "nine idempotent steps" against fourteen announced
    at the inventory's HEAD and sixteen after this pass.
  - `spec:1357-1359`, `main.go:56-64` for the version verb, which is
    `main.go:72-81`.
  - `spec:1463-1472`, the threshold `Kind` work assigned to rung 4, which
    pass 4a already landed (`thresholds.go:10-18`, `:74-96`,
    `doctor.go:111-128`). Nothing was owed at 4b.
  - `spec:1166`, `dubplate-import.service:14` for the `EnvironmentFile`
    line, which is `:15`; line 14 is `User=music`.
  - `spec:1243-1245`, `docs/HISTORY.md:2327-2328` for the open
    `/var/lib/musicbox` SELinux check, which lives at `:3082-3083`. The
    claim is right and the pointer is not.
  - `spec:1318-1322`, deploy edit 3's pre-create of `/var/lib/dubplate`
    against the handoff's leaf grant, resolved here as R1: the grant is
    the leaf and the pre-create makes both levels.
  - `spec:1632`, the pass-split row's "deploy.sh's six edits", which
    understates edit 6 by two files (`Makefile` and
    `cmd/dubplate-server/main.go`).
  - `docs/HISTORY.md:389-394`, the handoff's item 1, which reads as a
    grant on the store root without saying the pre-create must make it.
- **Three notes filed, none of them this pass's work:** doctor rows for
  the two stamps, for whichever rung next touches `doctor.go` (R12); the
  retention knob's trigger, which is the report's store-size row against
  the free-space baseline the sitting records (HISTORY forward carry 8);
  and rung 9's re-verification of the limiter's caller-priority policy
  against a live measurement, which arrives only after the deploy
  (forward carry 6).
- **The declared gaps this pass adds**, on the same criterion 4a used, a
  surface this pass builds that no test here can reach:
  1. **The grant's minimality has no mechanical check.** Held by the
     unit's static argument and Task 7's read.
  2. **The environment template's values have no checker.** A wrong path
     surfaces at the sitting's journal check, not in the gate.
  3. **The smoke assertion has never run against a box.** Its fires-proof
     is Task 5's local mutation 3; the real run is the sitting's step 4.
  4. **The unit has never started.** `systemd-analyze verify` proves it
     parses; nothing proves it starts, which is the sitting's step 7.

  **The count is four, and it joins the running total of thirty** (rung
  3's twenty-one plus 4a's nine), so rung 15 reads the total and never
  re-counts it.
- **Spend so far is reported against the 3M ceiling**, as a running
  figure rather than a close, with the sitting's own spend named as
  separate.

**Acceptance criteria:**

- STATUS is under 60 lines, present tense, and its next action is the
  sitting checklist in runnable order.
- Every readiness item is stated as a fact with its evidence, not as an
  intention.
- The sixteen resolutions each name a file and a line.
- The eight corrections and three notes are drafted, and **no spec file
  is edited by this task**.
- No HISTORY entry, no budget score, no deploy date.

**Mutation proofs:** none. This task writes prose for the conductor and
changes no behavior. It reports that plainly rather than manufacturing an
assertion.

**2.0 hook: neutral.**

**Independent: no.** It depends on every other task's outcome.

**Deliverable count: 4** (the STATUS rewrite with the sitting checklist,
the resolution list, the eight corrections, the four declared gaps with
their running total).

---

## Task 9: The sitting, and the pass close (done, 2026-09-08)

**Attended, a later session.** Not a chain task and not this window's
work. Ruling 6 makes the first deploy of greenfield code an owner
sitting, and this task is that sitting plus the close it makes possible.

**Files:**
- Modify: `docs/STATUS.md`, `docs/HISTORY.md`, `ROADMAP.md` if the
  initiative row moves (all conductor-committed)

**What runs:** the eleven-step checklist in "Owner touchpoints" above, in
order, with read-only ssh until step 4.

**What the close records, and these are the figures that make the sitting
worth its own task:**

- **The deploy date, as a first-class figure**, read off the box's own
  `started-at` rather than off the sitting's wall clock. `started-at` is
  written on the first pass that completes all three steps without a
  fatal error, **including a pass that observes nothing**, so the clock
  starts at the deploy (HISTORY handoff item 6, Resolution R11 of the 4a
  plan).
- **Both SELinux verdicts**, with the `ls -Zd` output. The
  `/var/lib/musicbox` half closes the item carried open at
  `docs/HISTORY.md:3082-3083` since the rewrite pass, and the close says
  so explicitly so the carry can be struck.
- **The free-space baseline** and whether `dubplate-disk-check` watches
  the root filesystem.
- **The first pass's journal output**: whether it observed anything, what
  it excluded, and the store's own size row.
- **The full HISTORY entry in the house shape**: what landed, what the
  gate caught, and what a later pass would be wrong to rediscover.
- **Both budgets scored**: tokens against the 3M ceiling, and attended
  interactions as two counts. **The sitting itself is one execution
  sitting**, and it is a planned one rather than a miss; a planning miss
  is an ambiguity that surfaced after approval which a planning question
  would have caught.
- **The rung 4 close**, since 4b is rung 4's second and last pass, with
  the handoff to whichever rung the phase plan takes next.

**2.0 hook: neutral.**

**Independent: no.** It is the pass's last item and depends on all of it.

---

## Execution model

Workflow mode, the `~/.claude/workflows/pass-execute.js` chain per task.
The plan names the mode, which is the opt-in, and marks its independent
tasks.

**Per task:** `cairn-implementer`, instructed to read and follow
`.claude/agents/dubplate-implementer.md`, which commits the task's
specific files once `bash scripts/check.sh` is green; then
`diff-reviewer` (`claude-opus-5`) against the task's acceptance criteria;
then the gate, inside the chain and never in the main loop. One
re-dispatch on `fix`; a second `fix` is the conductor's decision. The
conductor dispatches `code-simplifier` over Task 5's Go after its review
accepts, and over nothing else, because Task 5 is the only task writing
Go.

**Models.** `sonnet` for every implementer task. `claude-opus-5` for Task
0, for every `diff-reviewer`, and for Task 7. **No upshift is planned.**
This pass writes about ten lines of Go and four shell edits, and none of
it is novel correctness-critical logic the plan leaves unspecified. **One
named runner-up:** Task 4, upshifted to `opus` only if its first review
returns `fix`, because the render pass is the one place where a subtle
rewrite could keep the gate green while breaking the in-memory-only
property.

**Dependencies and order, stated exactly:**

- **Task 0 runs first, in fresh context, before any implementer
  dispatch.** Its blockers are ruled by the conductor before Task 1.
- **Task 1 → Task 3 → Task 4 → Task 5.** All four write
  `scripts/deploy.sh`, so they are strictly serial and the order is the
  file's own: the manifest, then the directory step, then the render
  refactor, then the build stamps. Task 5 also needs Task 4's render pass
  and install step in place, because both touch the region around the
  binary step.
- **Task 2 → Task 4**, an ordering edge rather than a territory one.
  Task 2 writes only `env/` and contends with nobody, but Task 4's second
  install step names the template Task 2 creates, and the render reads
  that file directly under `set -euo pipefail`. Run out of order, every
  `deploy.sh --check` and every apply aborts while `scripts/check.sh`
  stays green, because the gate never invokes `deploy.sh`. **The full
  order is Task 0 → Task 2 → Task 1 → Task 3 → Task 4 → Task 5**, with
  Task 6 free.
- **Task 6 depends on Task 0 alone** and is **independent**: its five
  files are disjoint from every other task's, and nothing reads its
  output. The conductor may run it at any point, and pass-execute's
  parallel mode may take it, with one caveat named rather than
  discovered: **one tree runs one executor**, because the module gate
  runs over `./...` and two whole-repo legs besides. Parallelism here
  means reordering freedom, not two live executors in this tree.
- **The conductor's `deploy.sh --check` runs after Task 5**, and its
  output feeds Task 7.
- **Tasks 1 to 6 plus the `--check` → Task 7.**
- **Tasks 0 to 7 → Task 8.**
- **Task 8 → the boundary. Task 9 is a later attended session.**

**Checkpoint:** every four tasks, which lands one checkpoint after Task
4. At each checkpoint, write STATUS with the task ledger, the decisions
taken, the spend, and the next task, then continue.

**Review scaling.** Rung 4 is not one of the three full
adversarial-cycle rungs (5, 12 and 16), so review is the in-chain
`diff-reviewer` per task plus the pass-end gate. **The spec's two named
exceptions are both this pass's**, and both are honored: the
fresh-context read of the unit files and the deploy path is Task 7, and
the deploy sitting is Task 9.

**Runaway guard armed at launch**, per the standing requirement.
**Subagents never dispatch agents.** No implementer runs
`code-simplifier`, and no implementer runs an ssh of any kind.

**Budget:** 3M ceiling for the unattended half, flag at 2.4M routed to
the conductor, checkpoint every four tasks, and the sitting's spend
counted separately in a later session. Both budgets are scored at Task 9,
not at Task 8.
