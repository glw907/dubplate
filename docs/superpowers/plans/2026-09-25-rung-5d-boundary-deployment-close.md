# Rung 5, Pass 5d (The Boundary, the Deployment, and the Close) Implementation Plan

Status: **revision 2.2, 2026-09-25; approved (Geoff, 2026-09-26): Ruling 1 yes (5d-i 7.0M, 5d-ii 3.5M), Ruling 2 yes (the attended deploy sitting).** Revision 2.2 folds a prose-and-facts review
(`docs/superpowers/research/2026-09-25-rung-5d-plan-prose-review.md`). The review history, one
disposition per finding for the earlier folds, is `docs/superpowers/research/2026-09-25-rung-5d-plan-fold.md`.
Geoff reviews this revision at the plan-approval gate. 5d-i launches in lane 1 (the main tree) only
after 5c-iv closes and 6c merges to master (conductor decision, revision 2.1); Task 0 re-verifies every
pointer against the tree they leave.

**The fold adds one task (Task 4D, the deploy script's safety), which takes the pass past ten tasks,
so 5d runs as two passes at the named seam** (the ten-task line, rung 6 spec `:1149`, applied here to
rung 5): **5d-i** is Task 0 and Tasks 1 to 6, the
build; **5d-ii** is Tasks 7 to 11, the review, the sitting, and the rung close. The two halves are
file-disjoint at the seam: 5d-ii writes Go only through Task 9's folds.

Token ceilings (Ruling 1; arithmetic in "Token ceiling"): **5d-i 7.0M, flag 5.6M; 5d-ii 3.5M, flag
2.8M.** Checkpoint interval: STATUS at every segment boundary, at most three tasks apart. Segments:
5d-i Task 0 (alone, before any segment), **A** (Tasks 1, 2, 3), **B** (Tasks 4D, 5, 6), then the 5d-i close; 5d-ii **C** (Tasks 7, 8, the
adversarial cycle, Task 9), **Sitting** (Task 10, attended, under Ruling 2), **Close** (Task 11).
Planning spend is recorded against the rung, outside both ceilings.

Sources. Code pointers were resolved at `4c766c3` on master and re-checked by the fold at `70ce7be`,
with warm uncommitted edits from other lanes under `cmd/dubplate-server/web.go`, `internal/config`, and
`internal/serve/document.go`; nothing this plan cites was among the lines they touch. Box facts were
read read-only on 2026-09-25, once by the drafter and once by the fold (below, "Box facts").

- The rung 5 spec (`docs/superpowers/specs/2026-09-08-rung-5-identity-auth.md`, "the spec" below): the
  pass-split table's 5d row (`:2519`), R6 (`:363-392`), R12 (`:454-462`), R21 and R22 (`:566-577`),
  R27 (`:612-624`), "The trust boundary" (`:795-903`), the owned-prefix route list (`:915-918`),
  "Integration classes" and "The environment template" (`:1635-1686`), "The state directory"
  (`:1690-1733`), "Deployment" (`:1737-1779`), the threat model (`:1950-2009`), "Errata the close
  applies" (`:2013-2463`), the staleness pre-task (`:2626-2700`), risks 10 and 11 (`:2751-2762`), and
  "Process" (`:2799-2837`).
- The build spec (`docs/superpowers/specs/2026-09-02-greenfield-server-build.md`, BS): ladder entry 5
  (`:707-714`) and the forward-auth requirements (`:504-514`).
- `docs/STATUS.md`'s carried "No deploy until 5d" item, and its "To rung 12" and "To rung 14" items.
- `docs/HISTORY.md`: 5a's forward carries (findings 22 and 23, R-G, R-I; `:3857-3885`, `:4045-4056`),
  5b-ii's disposition (major 5, minors 9 and 14; `:2244-2253`, `:2405-2413`), 5c-i's runbook carries
  (`:1489`, `:1729-1731`), and 5c-ii's declared gaps (`:555-608`).
- The 5c-iii plan (`docs/superpowers/plans/2026-09-25-rung-5c-iii-pages-fonts-rendered-gates.md`): its
  scope wall (`:60-65`) and Task 11's handoff to 5d (`:1768-1777`).
- The rung 6 spec (`docs/superpowers/specs/2026-09-23-rung-6-ops-seam-store.md`): the 6c lane
  condition (`:1062`, `:1173-1176`), 6d's criteria 13 and 14 (`:1115-1117`), the settings metadata
  (`:729-736`), and "Not verified" (`:1263-1265`).
- The streamer spec (`docs/superpowers/specs/2026-09-07-streamer-2-0-design.md`, which still says 2.0
  for the 0.8 streamer): "Forward hooks" (`:683-768`) and "Prior art" (`:770-806`).
- The two login-spec reviews (`docs/superpowers/research/2026-09-02-login-spec-reviews/security.md`,
  SEC, and `universality.md`, UNI), and the rung 5 security review records under
  `docs/superpowers/research/` (5a, 5b-i, 5b-ii, 5c-i, 5c-ii, 5c-iii), the do-not-re-litigate list.
- The three plan lens reviews, `docs/superpowers/research/2026-09-25-rung-5d-plan-review-{contract,
  mechanics,risk}.md`, and their "Verified sound" sections, which join the do-not-re-litigate list.
- The translit plan (`docs/superpowers/plans/2026-09-25-translit-table-at-setup.md`, approved), which
  also edits `scripts/deploy.sh` and `cmd/dubplate-server/doctor.go` from lane 2.

Box facts (read-only, 2026-09-25). systemd 257; SELinux enforcing; `music` is uid 2000;
`/var/lib/dubplate` and its only leaf `shadow` are `music` mode 0755 (`var_lib_t`); nothing listens on
8080; `/etc/dubplate/server.env` carries the ten rung 4 rows and no `DUBPLATE_IDENTITY_SOURCE`; the
installed binary is `pre-greenfield-321-gbe27434`; the timers `dubplate-import`, `-backup`,
`-disk-check`, `-navidrome-check`, and `-shadow` are all enabled. **Two drifts the apply will carry:**
the box's beets is **2.13.1** while `deploy.sh` pins `BEETS_VERSION="2.14.1"` (`deploy.sh:32`, adopted
at `367564c`), and the box's `config/contributors.yaml` and `config/reject-email.txt` differ from master
(the "family" to "group" copy, `53ecd0a`); `config/roster.example.json` exists only on master, so the
`--delete` rsync adds it; `compose.yaml` matches.

What this pass is. Rung 5's last pass, in two halves. It gives the serve shell its two remaining
configuration surfaces, gives `doctor` rows whose status comes from a probe of the running server,
makes `deploy.sh` safe to hand an unattended agent (check by default, `--apply` to write, a clean tree,
every local build before the first remote write), writes the deployment (the serve unit, its grant,
`deploy.sh`'s pre-creates and clean web build, the environment template's rows, and a generalized
state-dir check), writes the operator runbook and the trust-boundary documentation, maps all 50 SEC
and UNI findings to where each was absorbed, runs the rung's full fresh-context adversarial cycle over
the whole subsystem, and closes the rung. Under Ruling 2's yes it ends with the first production
deploy of the greenfield server, as one attended sitting with a written rollback.

What this pass is not. It builds no streamer driver, reverse proxy, fallthrough, SSO header, or
roster sync (rung 12). It mounts no API and no bearer channel (rung 14). It builds no revoke verb and
does not close review minor 14 (epoch monotonicity across processes), which the rung 6 spec gives to
6d's criterion 8. It takes none of STATUS's "To rung 12" items (the streamer-admin mapping, the shared
origin M2, Ruling 7's availability question, the `streamer_username` re-validation, the header step's
fallthrough reconciliation, the streamer database in a restore) and none of the "To rung 14" items (the
projection's mount, the production authorizer, the bearer channel, Ruling 3's 403, security L5 to L7).
It does not take L5 (a roster-POST throttle) or L6 (an n5 whole-roster doctor row), which stay
unassigned in STATUS. It adds no Go module dependency and no web dependency. **It edits nothing under
`internal/config`** (see "Global constraints"). It writes nothing under `web/`, `internal/auth`,
`internal/identity`, `internal/notify`, `internal/engine`, `internal/ops`,
`internal/backup`, `compose.yaml`, `config/` (the example roster already landed at 5b-ii), `corpus/`,
`tools/`, or `_salvage/`. It edits `docs/bringup-runbook.md` only to add `--apply` to its `deploy.sh`
apply invocations (Task 4D). **Nothing deploys and nothing writes the box before Task 10**, and Task
10 exists only under Ruling 2's yes.

> For agentic workers: workflow mode per CLAUDE.md's "Build process", the pass-execute chain per task,
> sequential within each segment. The implementer is `dubplate-implementer`. The conductor consumes
> structured reports and never reads a diff, a test log, or a gate transcript inline. Subagents never
> dispatch agents, so `code-simplifier`, the two adversarial lenses, the prose reader, and the
> architecture reads are the conductor's dispatches. The conductor owns `docs/STATUS.md`,
> `docs/HISTORY.md`, `ROADMAP.md`, the lane registry, and the contested-surface rows. **This pass runs
> in lane 1, the main tree.** The live-executor check is the lane registry, never `pgrep -f
> Projects/dubplate`. **No implementer dispatch ever passes `--apply` to `scripts/deploy.sh`**; every
> dispatch that may run it says so in its prompt (R-5d16).

---

## Rulings

Two yes/no questions, each with a recommendation and what each answer builds. Everything else this
plan decides is in "Resolutions this plan takes", each with its reason, because each has one
architecturally correct answer (Geoff, 2026-09-25: "You don't need to ask questions for the sake of
ceremony"). Revision 2 asked a third question, whether the forward-auth verify endpoint moves to rung
12; revision 2.1 takes it as R-5d17, because CLAUDE.md's forward-hook rule settles it. The rung 5
spec's rulings, the 5a through 5c-iv plans' rulings, and the rung 6 spec's rulings stand and are not
reopened.

**1. Do 5d-i and 5d-ii run under these ceilings: 5d-i 7.0M (flag 5.6M), 5d-ii 3.5M (flag 2.8M)?**
Recommendation: yes.

Reason: the split itself is not the question. It follows from the ten-task line (rung 6 spec `:1149`,
applied here to rung 5): Task 4D makes eleven tasks, so the pass splits at the seam revision 1 already
named, where the build tasks and the review-and-close tasks are file-disjoint. "Token ceiling" derives
each half's projection from 5c-i's per-task actual (the most recent rung 5 Go pass with a security read
and fold) plus named adders; each ceiling is its projection divided by 0.8, rounded up to the build's
0.5M grain, so each flag sits at or just above its projection and marks an overrun rather than firing
on plan. The spec's 3M (R23) predates every rung 5 actual and is superseded by the re-derivation R23
itself requires ("Each ceiling is re-derived from the previous pass's actual report"). What yes
builds: the two passes under these ceilings, the 5d-i close between them (a HISTORY entry, a STATUS
handoff naming 5d-ii's resume prompt, the simplifier round, and the two architecture reads), and 5d-ii
starting as a fresh session. What no builds: the same two passes under whatever ceilings Geoff names.

**2. Does 5d end with an attended sitting that deploys the greenfield server to the production box
for the first time?** Recommendation: yes.

A yes makes these changes on the box, and no others:

- The shadow's binary, `/usr/local/bin/dubplate-server`, is replaced by one built at Task 9's commit.
  Every matcher change since the installed build (`be27434`) goes live in the shadow, including the
  retirement of the matcher's workaround for beets 2.13.1's featuring-credit bug.
- The importer's beets moves from 2.13.1 to 2.14.1, the version the matcher is now tested against.
- `/etc/dubplate/server.env` gains the identity source, two newly minted HMAC keys, the SMTP settings
  and password, the roster path, the public URL, and trusted proxy set to off. Every `dubplate-server`
  verb reads this file, the shadow included, so the shadow holds the keys and the password too. That
  residual is carried to rung 16.
- `/opt/dubplate/config/` picks up the "family" to "group" wording in `contributors.yaml` and
  `reject-email.txt`, and an unused `roster.example.json`.
- `/var/lib/dubplate/auth` and `/var/lib/dubplate/roster` are created, and the roster holds one admin.
- `dubplate-server-serve.service` is installed disabled, started once on 127.0.0.1 so `doctor` can
  probe it, and stopped. Nothing routes to it; the tunnel still serves Navidrome only.

The shadow timer stops for the apply, which runs in the gap between two imports, and one manual
shadow pass runs before the timer restarts. The rollback restores the previous binary, `server.env`,
`/etc/dubplate/env`, and beets 2.13.1 together, and removes the unit. It leaves behind the wording
change, the two directories, and the two minted keys in the age store.

Why now: the rung's exit gate asks for `doctor` exiting 0 against the rendered environment, which
cannot pass until the two HMAC keys exist. The shadow reads the same environment, so any binary built
after 5b-i breaks it unless `server.env` carries the full sign-in set with real values. Without this
deploy, no shadow update and no other deploy, the transliteration pass's included, reaches the box
before rung 16. Minting needs Geoff's 1Password approval, so it belongs in an attended sitting.

What no builds: Task 6 writes the two keys as commented pending rows, Task 10 does not exist, the
exit gate's `doctor`-exits-0 half moves to rung 16, and the no-deploy carry moves to rung 16 with a
STATUS line naming the shadow's dependency on it.

---

## Primary inputs

The implementer reads, per task, only what the task's Files and Outcomes name, plus:

- `CLAUDE.md` in full (stack rules, the gate, safety rails, "Facts agents get wrong").
- `go-conventions` before any Go; `vps-conventions` before `systemd/`, `scripts/deploy.sh`, or
  `scripts/check.sh`; `golang-spf13-cobra` before any command wiring; `writing-voice` (Google developer
  documentation standard) before the runbook.
- `.claude/instructions/simplifier-brief.md`, pasted verbatim into every `code-simplifier` dispatch.

---

## Goal

After 5d, the serve shell reads a trusted proxy's client-address header only when an operator
configured one and only from that proxy's address; `doctor` reports the live client-IP mode, the web
surface's startup resolution, the configured bind addresses against the host's interfaces, and a live
boundary probe that dials the running listener and says plainly when none answers; `deploy.sh` checks
by default, writes only under `--apply`, refuses a dirty tree, and finishes every local build before
its first remote write; the repo carries a hardened greenfield serve unit whose grant names exactly the
auth leaf and the roster leaf, and a `deploy.sh` that pre-creates every greenfield leaf from one array
and installs the unit without enabling it; a state-dir check that derives its expectations from every
greenfield unit and proves the pre-create loop reads the array; a template whose render passes
`doctor`; an operator runbook; a disposition table covering all 50 SEC and UNI findings; a folded
two-lens adversarial review; the rung's errata applied; and, under Ruling 2's yes, the greenfield binary
and unit on the box with the admin bootstrapped, `doctor` green, and the shadow still passing.

---

## Owner touchpoints

The two rulings, at the plan-approval gate. Under Ruling 2's yes, the Task 10 sitting: Geoff approves
the one 1Password read that mints the two secrets; supplies his own display name and streamer username
for the admin bootstrap (the streamer username should equal his Navidrome login, since rung 12 maps
it); approves stopping the shadow timer; and approves the apply after reading `deploy.sh --check`'s
step list against the expected set, which names the beets 2.13.1 to 2.14.1 upgrade and every
`config/` file the apply carries (derived from `git diff --name-only <box stamp> HEAD`, Task 10 step
6). What would become a touchpoint: any Go module or web
dependency; any change to `compose.yaml` or the tunnel; any write to the box outside Task 10; any
`--check` line outside the expected set; a finding from either adversarial lens the fold would refuse
on a correctness-critical claim (the conductor's overrule then takes CLAUDE.md's `xhigh`-then-`fable`
path and records it).

---

## Token ceiling

Basis. 5c-i's recorded actual (`docs/HISTORY.md:1692-1698`): about 5.3M in subagent tokens across
seven build tasks, which amortizes Task 0, the fix rounds, the simplifier round, one register gate, one
pass-end `web-auth-security-reviewer` read, and its two-task fold. That is **0.76M per build task**,
review and fold included. 5c-i is the right basis: the most recent rung 5 pass whose surface is Go in
`cmd/dubplate-server` and `internal/serve` plus a security read, which is 5d's shape. 5c-ii's 1.15M per
task carried a new web tree and three reviewers, which 5d does not; 6b-ii-a's 0.65M carried no security
read.

**5d-i (Task 0, Tasks 1, 2, 3, 4D, 5, 6).**

1. Six build tasks at 0.76M: **4.56M**. Task 0 is inside the rate, as it was for 5c-i.
2. The heavy gate lane on every task (Chromium, about 3-4 min and about 560 MiB per full gate since
   5c-iii Task 3a), which lengthens each fix round's wait and report: **+0.30M**, 5c-iii's own adder
   for the same cause.
3. The fold's added criteria in Tasks 1, 2, 3, 5, and 6 (about eighteen new rows, mutations, and
   proofs, the wire-level probe proof and the `/dev/shm` namespace evidence the largest): **+0.25M**,
   about one fix round's worth.
4. The 5d-i close: the simplifier round and its `diff-reviewer`, two `go-architecture-reader`
   dispatches (`internal/serve`, `cmd/dubplate-server`), and the HISTORY and STATUS handoff drafts:
   **+0.25M**.

Projection 5.36M. 5.36 / 0.8 = 6.70, rounded up to **7.0M**; the flag at 80% is **5.6M**, 0.24M
above the projection.

**5d-ii (Tasks 7 to 11).**

1. Two document tasks (7, 8) at 0.76M: **1.52M**. Task 9's fold is inside the per-task rate, as
   5c-i's fold was.
2. The second adversarial lens (universality), which 5c-i did not run: **+0.35M**.
3. Both lenses read the whole rung 5 subsystem (four packages, `cmd/dubplate-server`'s auth and serve
   files, the unit, and `web/`'s three pages), not one pass's diff: **+0.25M**.
4. Task 11's rung close plus the sitting's evidence subagents under Ruling 2: **+0.50M** (revision
   1's adder, which also carried the architecture reads that now sit in 5d-i).
5. The sitting's added steps (the timer window, the `.prev` copies, the manual shadow pass, the names-only
   env diff, the rollback rehearsal read): **+0.10M**.

Projection 2.72M. 2.72 / 0.8 = 3.40, rounded up to **3.5M**; the flag is **2.8M**.

Together: projection 8.08M against revision 1's 6.72M (the 1.36M difference is Task 4D at 0.76M,
the fold's criteria at 0.25M, the split's second close at 0.25M, and the sitting's steps at 0.10M);
ceilings 10.5M against revision 1's 8.5M. The two ceilings together exceed the spec's 3M line for 5d by
7.5M, so rung 5's reservation moves from about 41M (5c-ii's close) to about 48.5M.

Deliverable count: Tasks 1, 2, 3, 5 carry four each; Tasks 4D and 6 carry three; Tasks 7, 8, 9,
11 carry three or four; Task 10 is a conductor-run sitting. None passes four.

---

## Global constraints

- **Lane 1, the main tree, on master**, after 5c-iv's close is committed, **after 6c has merged to
  master** (conductor decision, revision 2.1: 6c's Task 5 edits `doctor.go`, `check.go`, and their
  tests, and 5d-i's Tasks 1 to 3 and 6 edit `doctor.go`, `doctor_test.go`, `check_test.go`, and
  `main.go`, never `check.go` itself, so 5d-i builds on the tree 6c leaves rather than racing it), and the lane registry shows lane 1 free with this pass named. No task reads, writes, or runs a command inside
  `.claude/worktrees/`. The spec's "Worktree, released at the close" for 5d is superseded by the lane
  registry; the close records it.
- **`internal/config` is untouched, stated plainly for the 6c lane condition.** The rung 6 spec runs 6c
  in lane 2 "only if 5c and 5d leave `internal/config` untouched" (`:1062`). 5d declares its integration
  class in its own file in `cmd/dubplate-server` (CLAUDE.md: "Config keys are declared through the
  registry in each rung's own file"), through the existing `AddIntegrationClass` (`internal/config/
  registry.go:213`), with no reserved class to retire (`keys.go` carries no trusted-proxy class; read at
  `4c766c3`). `config.IntegrationClass` has no value-validation hook (`internal/config/keys.go:184-190`),
  so value validation lives in `cmd/dubplate-server`'s resolver, which serve startup and `doctor` both
  call, on R-5cii12's precedent for the theme keys. 6c adds a per-key value validator for store-homed
  settings, built from bounds data (rung 6 spec `:729-736`), which does not fit a list-valued
  integration key, so 5d's checks stay in `cmd/dubplate-server`'s resolver; moving them onto a
  registry hook is 6d's or the architecture cleanup pass's, never 5d's. `internal/config` stays byte-identical across both halves, checked at each close by `git
  diff --stat <5d base>..HEAD -- internal/config` answering empty.
- **Package territory, stated as a wall.** This pass writes: `internal/serve` (`clientaddr.go`,
  `strip.go`, a new probe file, and their tests); `cmd/dubplate-server` (a new `proxy.go`, `serve.go`,
  `doctor.go`, a new doctor-probe file, `main.go`'s registration line, and their tests, plus
  `templatekeys_test.go` and the shared test fixtures in `check_test.go`); `systemd/` (one new unit);
  `scripts/deploy.sh`; `scripts/check.sh`; `docs/bringup-runbook.md` (the `--apply` invocations only);
  `env/dubplate-server.env.template`; `docs/server-runbook.md` (new); the disposition table and review
  records under `docs/superpowers/research/`; the close drafts; and, at the 5d-ii close only, the errata
  targets. A task that believes it needs anything else stops and reports.
- **No Go module or web dependency.** `go.mod`, `go.sum`, `web/package.json`, and
  `web/package-lock.json` stay byte-identical. `golang.org/x/net/http/httpguts` is declined for header-
  name validation (R26 fixed rung 5's dependency count at two); a token check against RFC 9110's `tchar`
  set is a few lines.
- **No deploy before Task 10, enforced by the tool from Task 4D on.** No task writes the box. Tasks 1,
  2, and 3 never run `scripts/deploy.sh` (until Task 4D lands, a bare invocation applies). From Task 4D
  on, `deploy.sh` with no flag or with `--check` reports every step without applying, and only
  `--apply` writes (R-5d16). An implementer may run `bash scripts/deploy.sh --check 192.0.2.1`,
  always spelled with the flag and the documentation address, and never `--apply`; every such dispatch
  says so in its prompt. **Implementer runs never target the real box**: `dubplate-implementer.md`
  forbids `ssh` to `5.78.148.42` and any network use (its hard rules, `:43-44`). `192.0.2.1` routes
  nowhere, and it proves every property Tasks 4D, 5, and 6 claim: the local build lines, the step
  headers, the render, and `run_remote`'s "would run" lines print whether or not a host answers
  (`deploy.sh:133-137`), `remote_reachable`'s BatchMode probe (`:123-125`) fails and reports "can't
  reach" in place of a diff, and each `rsync -n` dry run ends in `|| true` (`:254-255`, `:267`, `:341`).
  The one real-box `--check` belongs to Task 10 step 6, run by the conductor. **Two stated exceptions
  to the implementer's no-network rule, named in every dispatch that runs `deploy.sh`:** the `ssh` and
  `rsync -n` connection attempts to `192.0.2.1`, each bounded by `ConnectTimeout=10` (`:74`), which
  cannot reach a host, and `npm ci
  --prefix web`'s reads from the npm registry, which fetch exactly the packages
  `web/package-lock.json` pins. The box stays read-only otherwise (`ssh -o BatchMode=yes` for a
  conductor read). The incumbent `/usr/local/bin/dubplate` and its units are never touched.
- **Every `deploy.sh` evidence run happens in a detached scratch worktree at the task's commit**, never
  in the main tree. R-5d13's clean-tree refusal fires on any modified tracked file, in both modes, and
  the main tree is shared with the document lanes, so a run there prints the refusal instead of the
  steps whenever any lane holds a warm edit. The implementer commits the task first, then runs `git
  worktree add --detach <scratchpad>/5d-deploy <sha>` (outside `.claude/worktrees/`), runs every
  `deploy.sh` evidence command from that directory, makes Task 4D's deliberate scratch edits there, and
  removes it with `git worktree remove --force <scratchpad>/5d-deploy` afterward. A run that shows a
  defect is fixed in the main tree by a new commit and re-run in a fresh worktree at that commit. Task
  10 uses the same form at Task 9's commit.
- **Assertion discipline at the repo's bar.** No vacuous assertion, no expectation derived from the code
  under test. Every hard-coded expectation cites the spec, a quoted man page or `go doc`, or a measured
  probe. Every test names the fixture state it needs and asserts that state before the property. Every
  `doctor` exit-1 assertion is the green baseline (`doctorEnviron`, Task 2) plus exactly one
  perturbation, so no exit 1 can come from an unrelated row. Every named mutation is proved to fail the
  gate and then reverted.
- **Skills before code**, as "Primary inputs". No process citation in a shipped comment. The em dash is
  banned in Go comments.
- **Universality.** No estate hostname, address, or path in Go code or a Go fixture; fixtures use
  `example.test` and the documentation ranges (192.0.2.0/24, 198.51.100.0/24, 203.0.113.0/24,
  2001:db8::/32). Estate values live in the template, the unit, and `deploy.sh`, which already carry
  them.
- **The gate is `cairn-run-gate 'bash /var/home/glw907/Projects/dubplate/scripts/check.sh'`**, re-issued
  on exit 75 until it prints `gate exit:`, **in the default heavy lane for every task: never set
  `CAIRN_GATE_LANE=light`.** Every full gate launches Chromium and runs about 3-4 minutes at about 560
  MiB. Evidence runs use `-count=1` and report wall time. Specific files only. Commit footer
  `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>` for an `opus` task, the dispatch's footer
  otherwise.

---

## Privileged fields and their channels

| Field | Channels that may set it | What refuses every other channel |
| --- | --- | --- |
| The trusted-proxy set and header name | `DUBPLATE_TRUSTED_PROXY_ENABLED`, `DUBPLATE_TRUSTED_PROXY_ADDRS`, `DUBPLATE_TRUSTED_PROXY_HEADER`, read once from the environment at serve startup through Task 1's resolver; `doctor` reads the same resolver. | No request component reaches the resolver. A value the resolver refuses stops serve before any listener opens (`newServeCmd`'s own "refuses at startup" contract, `serve.go:162-171`). |
| The reported client IP (R27) | `RemoteAddr` always. The configured header only when the peer address, compared as an unmapped `netip.Addr`, is in the configured set, and only its rightmost hop that parses (as an address, or as an address and port whose address is kept) and is not itself a configured proxy. | A header from any other peer is never read (`clientaddr.go:107-109`). A malformed rightmost untrusted hop ends the scan and yields the peer address, so no hop left of it is ever reached (Task 1). Group 2 strips every forwarded-address header from an untrusted peer (`strip.go:455-466`). The proxy must overwrite or append the configured header; a header it passes through unchanged is client-chosen (Task 7 states the operator's duty). |
| The peer address (the only input to a trust decision) | `RemoteAddr` alone (`PeerAddress`, `clientaddr.go:68-74`). | No header, no configuration. Unchanged by 5d. |
| The front exemption (one identity header a front may keep) | None at 5d. `newServeShell` passes `serve.FrontExemption{}`. | A Task 1 test asserts the shell's strip carries the zero exemption, so every identity spelling is stripped on every path from every peer. |
| The boundary probe's answer | The probe route, answering only a peer whose `PeerAddress`, parsed and unmapped as a `netip.Addr`, is loopback or one of the configured bind addresses **and is not in the trusted-proxy set**. Every other peer, and every method other than `GET` from a permitted peer, meets `serve.Terminator()`'s 404. | The answer carries counts only, never a header value or an address. `ClientIP` is never consulted, so no forwarded header can make a peer local. **Residual, recorded:** a same-host proxy that the operator did not configure as a trusted proxy forwards internet traffic from loopback, and such traffic reaches the route and learns the counts of its own headers that survived the strip (Task 7, Task 8's residue list). |
| The public URL (origin check, every link's base, HSTS) | `DUBPLATE_PUBLIC_URL`, a template literal rendered into `server.env` by `deploy.sh`. | No request component; `validateMagicLinkSource` refuses a non-HTTPS non-loopback value at `Load`. |
| The two state leaves' ownership and mode | `deploy.sh` (as root on the box) creating each with owner `music`, mode 0700. | `auth.Open` refuses a leaf wider than 0700 (`internal/auth/store.go:53-71`, `dirPerm = 0o700` at `:32`); the roster loader refuses a group- or other-writable file (`roster.go:318-319`). |
| The two HMAC secrets and the SMTP password | The workstation age store, rendered by `deploy.sh` into `/etc/dubplate/server.env` (0600 root:root). | No other channel. No value appears in the repo, a log, a test, a report, or a ledger; checks are name-only. **Residual, recorded:** every `dubplate-server` verb shares `server.env`, so the five-minute shadow (which parses member uploads) holds both HMAC keys and the SMTP password in its environment, readable by any process running as `music`; carried to rung 16 (`LoadCredential=` or a dedicated uid for serve). |

---

## Resolutions this plan takes

Each is the plan's own decision with its reason. The adversarial lenses may challenge any of them.

**R-5d1. The trusted-proxy class's shape.** Name `trusted proxy`; flag `DUBPLATE_TRUSTED_PROXY_ENABLED`;
Required `DUBPLATE_TRUSTED_PROXY_ADDRS` (comma-separated addresses) and `DUBPLATE_TRUSTED_PROXY_HEADER`
(one header name). Declared in a new `cmd/dubplate-server/proxy.go`, registered from
`newProductionRegistry` (`main.go:108-117`). Reason: R22 names exactly this shape (an enable flag
plus the proxy addresses and the client-address header, required when enabled); the file follows
`backup.go`'s precedent (`:40-54`).

**R-5d2. Address grammar.** Each entry is trimmed and parsed with `netip.ParseAddr`, then unmapped. `go
doc net/netip.ParseAddr`: "The string s can be in dotted decimal ("192.0.2.1"), IPv6 ("2001:db8::68"),
or IPv6 with a scoped addressing zone". `go doc net/netip.Addr.Unmap`: "if ip is an IPv6 address
wrapping an IPv4 address, it returns the wrapped IPv4 address." Unmapping matters because `go doc
net/netip.Addr.String` warns "unlike package net's IP.String method, IPv4-mapped IPv6 addresses format
with a "::ffff:" prefix", while `net.IP.String` (the form `RemoteAddr` carries) prints "dotted decimal
... if ip is an IPv4 or IP4-mapped IPv6 address". Refused, each naming the entry: an empty entry, a
value that does not parse (a hostname), a network in CIDR form (reason: "an address, not a network";
the spec pins trust to observed addresses, never a subnet, `:839-840`), an address with a zone
(`fe80::1%eth0` parses with zone `eth0`, measured by the mechanics lens, so the refusal is explicit),
and an unspecified address. The states: flag absent or false with no Required key set is mode
direct; flag true with `ADDRS` absent or empty is `Load`'s own missing-key report for an enabled class
(`config.go:408-427` at `593d7fd`; `lookupEnviron`'s rule makes an empty value unset); flag false or
absent with `ADDRS` or `HEADER` set is `Load`'s "set without ... the trusted proxy integration's enable
flag" report (`config.go:301` at `593d7fd`); a flag `strconv.ParseBool` refuses is `Load`'s own error (`integrationEnabled`'s doc
comment: "A value strconv.ParseBool refuses is reported as an error rather than silently read as off");
a malformed entry is refused naming it; many bad entries are joined with `errors.Join` so every one is
named at once. A duplicate entry collapses into the set and is not an error.

**R-5d3. Header grammar.** The header name must be an RFC 9110 token (letters, digits, and
`` !#$%&'*+-.^_`|~ ``), since `go doc net/textproto.CanonicalMIMEHeaderKey` says a key that "contains a
space or invalid header field bytes as defined by RFC 9112 ... is returned without modifications",
which would leave a malformed name uncanonicalized and unmatched. Refused by name, with a reason each:
`Cookie`, `Origin`, and `Sec-Fetch-Site` (group 3's keep set, `group3KeepSet`, `strip.go:199-203`;
adding one to group 2 would strip it from every untrusted peer and break sessions and the origin
check); `Host`, `Connection`, `Content-Length`, `Transfer-Encoding`, and `Authorization` (end-to-end or
framing headers no front reports an address in); `Forwarded` (RFC 7239's `for=` grammar is not an
address list, so every hop would fail R-5d4's parse); and any identity spelling group 1 strips
(`identityExtras`, `strip.go:112-135`, both forms through `identitySpellings`, `:210`), since group 1
would delete it before `ClientIP` reads it. Comparison is case-insensitive through canonical form, and
the `_` spelling of a refused dash name is refused too. The same states as R-5d2.

**R-5d4. The hop rule (review minor 9, SEC finding 22).** `ClientIP` scans hops right to left as today
(`clientaddr.go:121-135`). Each non-empty hop is parsed with `netip.ParseAddr`; a hop that fails is
parsed with `netip.ParseAddrPort` (`go doc net/netip.ParseAddrPort`: "both the address and the port
must be numeric") and its `Addr()` kept, so a proxy that appends a port (`203.0.113.9:4711`,
`[2001:db8::1]:443`) does not collapse every client onto one limiter key. A hop with a zone is
malformed. A configured proxy is skipped; the first hop that is not a proxy is returned in unmapped
`netip.Addr.String` form; **a hop that fails both parses ends the scan and returns the peer address**
(a bracketed address with no port, `[2001:db8::1]`, is such a hop), so an attacker who plants a
malformed value cannot make the scan continue leftward into values they chose. `TrustedProxyConfig`'s
set becomes `map[netip.Addr]struct{}` (5a's forward carry 1), and `PeerAddress` is compared after the
same parse and unmap. A `RemoteAddr` that does not parse is never trusted.

**R-5d5. Doctor's four new rows print unconditionally, and both line-count tests take a named
constant.** R6 requires the plan to say which way the collision goes. The rows are `client ip`, `web`,
`bind addresses`, and `boundary probe`, printed after the backup rows. They are not registry
declarations, so registering them through the registry would invent a family for four fixed rows; the
store row set this precedent (`storeRowLines`, `doctor_test.go:217-235`). `TestDoctorCmd` and
`TestDoctorCmd_LineCountFollowsRegistry` each add a `probeRowLines` constant equal to 4, hard-coded,
never derived from the rows' own slice. When `Load` fails and doctor holds no `*config.Config`
(`doctor.go:35-36`), `client ip` reads `not checked` and adds no failure of its own, as the web row
does.

**R-5d6. The web row runs serve's own web resolution.** STATUS's "the theme keys' doctor row" and the
5c-ii erratum ("`doctor` prints the configured value without judging it; that gap is carried to 5d")
close by calling `newWebRoutes` (`web.go:218`), the exact function serve startup calls
(`serve.go:303-306`), with the same inputs, reading the bundle and never writing. It judges the theme
match, the palette, the branding value, and the operator CSS file in one place, so doctor and serve
cannot disagree. Doctor passes a logger that discards, because `newWebRoutes` writes one record when a
branding neutral drops the theme-color meta (its doc comment: "logger takes the one record written when
a branding neutral leaves the documents without a theme-color meta"), and that record must not land on
doctor's stdout. A refusal reads `invalid: <error>` and fails doctor. When the prefix or the site name
cannot be resolved (so `Load` already failed), the row reads `not checked` and adds no failure of its
own. Reason: a per-key validator map would be a second copy of R-5cii12's checks; the security
review's minor 12 class ("`doctor` green for sign-in-breaking configurations") is exactly a doctor that
checks less than startup does.

**R-5d7. The bind-addresses row.** For each configured address (`serve.SplitBindAddrs`,
`bindaddrs.go:37-56`), parsed and unmapped as a `netip.Addr`: `present` when it equals the IP of an
interface address from `net.InterfaceAddrs` (`go doc net.InterfaceAddrs`: "returns a list of the
system's unicast interface addresses") or is a loopback address; `not on any interface` otherwise,
which fails doctor. Reason: R20's `IP_FREEBIND` makes the bind succeed for an absent address, so
the listener raises no error and the spec gives this row the job (`:1601-1607`). A loopback address
counts present because Linux routes 127.0.0.0/8 to `lo` while `InterfaceAddrs` lists only
`127.0.0.1/8`. **`InterfaceAddrs` reads the table over netlink** (`net/interface_linux.go`,
`interfaceAddrTable`: `syscall.NetlinkRIB(syscall.RTM_GETADDR, syscall.AF_UNSPEC)`), which the unit's
`RestrictAddressFamilies=AF_INET AF_INET6 AF_UNIX` denies; measured by two lenses under that
restriction: `route ip+net: netlinkrib: address family not supported by protocol`. The unit's family
set is not widened, since serve itself never needs netlink. The five states: bind addresses unset
is the registered default (loopback) and `present`; empty or malformed is `SplitBindAddrs`'s refusal as
`invalid: ...`, failing; one absent address fails naming it; many absent addresses are all named on
the one row; **`InterfaceAddrs` returns an error**, and the row reads `not checked: <error>` and fails,
naming `RestrictAddressFamilies` as the likely cause, since the operator asked for the check. The
status function takes `InterfaceAddrs`'s result (addresses and error) as arguments, and doctor calls
`net.InterfaceAddrs` once, so every state is testable without a seam. Doctor is run without the unit's
family restriction (Task 7, Task 10).

**R-5d8. The boundary probe observes through a local-only probe route.** The spec leaves the probe's
observation unstated. A route under the owned prefix, named `boundary-probe`, answers `GET` with a
fixed JSON object of three counts: identity headers that reached it (group 1), forwarded-address
headers (group 2), and a probe-owned sentinel header the request listed in `Connection` (group 3), with
`Cache-Control: no-store`. **The route checks the peer first and the method second, and both refusals
answer `serve.Terminator()`** (`internal/serve/prefix.go:130-134`, `http.NotFound`). It keeps
`exactPath` and does not use `methodOnly`, because `methodOnly` answers 405 with `Allow: <method>`
(`cmd/dubplate-server/serve.go:380-389`), which would tell a stranger the route exists. **The peer
rule:** `PeerAddress(r)`, parsed and unmapped as a `netip.Addr`, must be loopback (`IsLoopback`) or
equal one of the configured bind addresses (each parsed and unmapped the same way, since
`SplitBindAddrs` returns strings verbatim and `net.ParseIP` accepts `::ffff:127.0.0.1`), **and must
not be in the trusted-proxy set**. `ClientIP` is never consulted. Counting: the route counts each
name with `r.Header.Values(name)`; `go doc net/http.Header.Values`: "It is case insensitive;
textproto.CanonicalMIMEHeaderKey is used to canonicalize the provided key." That matches what arrives,
because `go doc net/textproto.Reader.ReadMIMEHeader` says "The returned map m maps
CanonicalMIMEHeaderKey(key) to a sequence of values", so `Remote_User` arrives as `Remote_user`, a key
distinct from `Remote-User` (`CanonicalMIMEHeaderKey` treats `_` as an ordinary token byte; measured by
two lenses). The probe client builds its request with `Header.Add` per spelling, one request builder
shared by doctor and every proof. Group 3's evidence: the sentinel's zero count means the strip ran
only if Go's server does not itself remove a `Connection`-listed header; the mechanics lens measured
that it does not ("a Go server receives both without removing any hop-by-hop header"), and Task 3's
wire-level no-strip proof re-proves it in the gate. The probe client and the route live in one
`internal/serve` file, so the header lists stay unexported and the client can never send a set the
route does not count. `doctor` dials each configured bind address on the configured port with
`net.Dialer.LocalAddr` set to that same address (`go doc net.Dialer`: "LocalAddr is the local address to
use when dialing an address ... If nil, a local address is automatically chosen"), so the peer the shell
sees is known. The dial and the response each carry their own 2-second bound. Reason: the spec's R6
wants a live dial that "cannot fail for any deployment reason" to be replaced by one that can; only a
response the running process computes after its own strip can tell doctor whether the strip ran.
Excluding the trusted set keeps a same-host proxy that the operator did configure from exposing the
route, and closes BS:510-512's "from a non-trusted source" for the probe's own dial. The residual (a
same-host proxy the operator did not configure) is recorded in the privileged-fields table; a secret or
second gate on the route costs more machinery than a count leak is worth.

**R-5d9. The probe row's states, each named.** Per address: `refused` (all three counts zero; since
the permitted peer is never a trusted proxy, group 2 must be zero too); `no listener` only when the
dial is refused (`ECONNREFUSED`) or the dial itself times out, which is **not** a failure, since the
runbook runs doctor before the unit starts and the sitting runs it with the unit stopped; `not probed
(not on any interface)` when the dial's local bind fails because the address is absent (measured:
`bind: cannot assign requested address`), adding no failure beyond R-5d7's; `not probed: every local
address is a trusted proxy` when the dial address is itself in the trusted set, which is not a failure;
`passed N identity header(s)`, `passed N forwarded-address header(s)`, or `passed the
Connection-listed sentinel`, failing; `unexpected answer: <status>` for any non-200 or unparseable
body, and `unexpected answer: timeout` for a connection that accepts and never answers within the
response bound, failing. The row ends with `proves this server's strip, not a proxy's` (R6: "it cannot
prove an adopter's proxy strips one"). When the bind configuration is invalid the row reads `not
probed` and adds no failure beyond R-5d7's.

**R-5d10. The unit's name, type, and body.** `systemd/dubplate-server-serve.service`: it names the
binary and the verb, and it is distinct by more than one letter from the incumbent
`dubplate-serve.service`, which stays untouched until rung 16 (spec `:1739-1742`). `Type=exec`: `man
systemd.service`, "systemctl start command lines for exec services will report failure when the
service's binary cannot be invoked successfully", which `Type=simple` does not. The body inherits the
shadow unit's hardening block verbatim minus nothing (the shadow already dropped `AF_NETLINK`,
`dubplate-shadow.service:84-89`), `User=music`, `UMask=0027`,
`EnvironmentFile=/etc/dubplate/server.env`, `ExecStart=/usr/local/bin/dubplate-server serve`,
`Restart=on-failure`, `RestartSec=5`, `StartLimitIntervalSec=300` with `StartLimitBurst=30` (the
2026-09-01 ruling the incumbent records), `OnFailure=dubplate-ping-fail@server-serve.service` wired and
inert (a slug of its own, so a rung 9 mint pages this unit's check and never the incumbent uploader's
`serve` slug; the helper maps it to `DUBPLATE_PING_SERVER_SERVE_URL`,
`_salvage/cmd/dubplate/pingfail.go:64`), `After=docker.service` without `network-online.target` (R20),
`TimeoutStopSec=30`, which covers serve's own 25-second drain (`shutdownGrace`,
`internal/serve/server.go:52-56`: "A supervising unit's own stop timeout ... is the real backstop"),
and `ReadWritePaths=/var/lib/dubplate/auth /var/lib/dubplate/roster` with no `-` or `+` prefix. `man
systemd.exec`: paths "may be prefixed with "-", in which case they will be ignored when they do not
exist", which would turn a missing pre-create into an `EROFS` at first write instead of the namespace
failure (the mechanics lens measured the `-` case starting cleanly). An `[Install]` section with
`WantedBy=multi-user.target` exists so a later `systemctl enable` works, and `deploy.sh` never enables
it (risk 10).

**R-5d11. The generalized state-dir check (R12).** `deploy.sh` gains one array naming every greenfield
leaf under `DUBPLATE_LIB_DIR`, `DUBPLATE_SERVER_LEAVES=(shadow auth roster)`, and **one pre-create loop
that iterates `"${DUBPLATE_SERVER_LEAVES[@]}"`** and runs `install -d -o music -g music -m 0700` for
each leaf, **skipping `shadow` by name** with a comment saying why: the shadow leaf's existing line
(`deploy.sh:503-505`, `mkdir -p` plus `chown`) stays, because the live shadow leaf is mode 0755 on the
box and `install -d -m 0700` re-modes an existing directory. `install -d`'s behavior is measured, not
recalled (GNU coreutils 9.10 on the workstation, 2026-09-25): on an existing 0755 directory `install -d
-m 0700` leaves it 0700; on a missing parent it creates the parent with the default mode (0755 under
umask 0022) and only the leaf with `-m`. The loop runs after the shadow line, which creates and chowns
`/var/lib/dubplate`. `scripts/check.sh`'s state-dir leg: (1) selects every unit under `systemd/` whose
`ExecStart=` names `/usr/local/bin/dubplate-server`, and for each `ReadWritePaths=` entry asserts no
`-` or `+` prefix, that the entry is the template's `DUBPLATE_STATE_DIR` or a leaf directly under it,
and that the leaf is in `DUBPLATE_SERVER_LEAVES`; (2) asserts `deploy.sh` carries the loop line that
iterates `"${DUBPLATE_SERVER_LEAVES[@]}"` and that its body runs `install -d`; (3) asserts the shadow
pre-create line naming `$DUBPLATE_SHADOW_DIR` is present; (4) asserts `DUBPLATE_LIB_DIR` equals the
template's `DUBPLATE_STATE_DIR` (today's second half). Each failure names the unit, the entry, or the
line, and the missing piece. The four states: no greenfield unit found fails (a selector that
matches nothing must not pass); a unit with no `ReadWritePaths=` passes for that unit (a read-only unit
is legal); a malformed entry fails naming it; many bad entries are all reported before the leg fails.
Reason: data beats a third hard-coded pair (R12), the loop-reads-the-array assertion is what makes
"failing when either leaf is removed from `deploy.sh`" true of the pre-create and not only of the
declaration, and 6d adds the `store` leaf by one array word (rung 6 spec criterion 14).

**R-5d12. `systemd-analyze` becomes hard-required.** `scripts/check.sh:631-643` guards the verify leg
with `command -v systemd-analyze ... skipping`, which CLAUDE.md forbids ("never guard one with `command
-v || skip`"). The guard is removed and `systemd-analyze` joins the hard-required tool list at the top
of the script (`check.sh:138-157`, beside go, golangci-lint, shellcheck, ruff, gitleaks), so a missing
tool fails by name before any leg runs and a re-added guard in the leg has no job. Reason: the exit
gate's first assertion is "`systemd-analyze verify` clean on the new unit".

**R-5d13. `deploy.sh` ships only committed source, built before its first remote write.** After the
render (`deploy.sh:240-247`) and before the first remote write (the repo-config rsync, `:252`),
`deploy.sh` refuses, naming the paths, when `git status --porcelain --untracked-files=no` is non-empty
(a tracked edit anywhere, since `git describe --dirty` at `:334` would otherwise stamp a reviewed
commit's id on unreviewed code) or when `git status --porcelain -- web/` is non-empty (an untracked file
under `web/` would be embedded); then runs `npm ci --prefix web`, `npm run build --prefix web` (`vite
build`, `web/package.json:8`; `make web` would run the gates a second time), and `make build-linux`
with the stamp (`Makefile:14-15`). The binary's rsync stays where it is. Nothing Node or Playwright
ships to the box: the rendered leg is a workstation gate only. The refusal and the builds run in both
modes, on the binary step's own precedent (`deploy.sh:317-320`). A dirty `web/` or a failed build
therefore aborts before anything reaches the box. `web/build/placeholder.txt` survives a build, so a
gate run leaves `git status -- web/` clean (checked by the mechanics lens). Reason: 5c-ii's handoff
("the binary embeds whatever `web/build` holds"), 5c-iv's Task 11 row, and the shared main tree.

**R-5d14. The template's rows.** `DUBPLATE_ROSTER_FILE=/var/lib/dubplate/roster/roster.json` and
`DUBPLATE_PUBLIC_URL=https://music.907.life` as literals (not secrets, never minted; the hostname is the
tunnel's own, `config/cloudflared-tunnel-config.json`, and it carries no path per R-5c9). The four
non-secret SMTP names (`DUBPLATE_SMTP_HOST`, `_PORT`, `_USER`, `_FROM`) and `DUBPLATE_SMTP_PASSWORD` as
**bare rows**, correcting the spec's "literal" for the four: the age store already holds all five under
these exact names (the registry's note on "NEW non-secret values (SMTP host/port/user/from ...)"), so
"none is ever minted" is stale, a bare row keeps a personal address out of a repo that goes public at
0.8, and the flat-namespace hazard that forced `DUBPLATE_STATE_DIR` to a literal
(`env/dubplate-server.env.template:38-53`) does not apply, because the incumbent and the greenfield
server use the one transport with the same values. `DUBPLATE_MAGIC_LINK_SECRET` and
`DUBPLATE_SESSION_SECRET` as bare rows under Ruling 2's yes, commented pending rows under its no.
`DUBPLATE_MAGIC_LINK_SECRET_PREVIOUS` and `DUBPLATE_SESSION_SECRET_PREVIOUS` as commented rows (Optional
keys, used only across a rotation). `DUBPLATE_TRUSTED_PROXY_ENABLED=false` as an explicit literal,
because proxyless operation is "chosen, never fallen into" (BS:278-280); the estate's own values (the
cloudflared container's address, which Docker allocates dynamically on `musicbox_default` today, and
`Cf-Connecting-Ip`) wait for the cutover, which must first pin that container's address in
`compose.yaml`. The header's "No row here is a secret" paragraph is rewritten. The serve bind and port
rows are not added: the unit stays disabled and loopback is the right default until the cutover binds
the bridge gateway.

**R-5d15. The runbook's home.** A new `docs/server-runbook.md`, the greenfield server's operator
runbook, beside the incumbent's `docs/bringup-runbook.md`, which gains only R-5d16's `--apply` on its
apply invocations. Reason: the spec requires a runbook and trust-boundary documentation
(`:1771-1774`, `:884-903`, `:2865-2872`), the incumbent file documents a retiring estate, and a section
inside it would be deleted with it at rung 16.

**R-5d16. `deploy.sh` checks by default and writes only under `--apply`.** Today no flag means apply
(`deploy.sh:81-95`: `check_only=false` unless `--check`). From Task 4D, no flag and `--check` both
report every step without applying, `--apply` applies, and both flags together are refused. The usage
comment (`:18-24`) and the run banner name the mode. Reason: this pass is the first to put
`deploy.sh` in implementers' hands, implementer subagents hold BatchMode root ssh to the box (the risk
lens reached it), and one dropped flag would turn a read-only check into M1's estate-wide apply on
production, unattended. CLAUDE.md's "a rule lives where it executes" asks for the strongest form: the
tool makes the dropped-flag failure safe. The alternative (handing every `--check` run to the
conductor, on the translit plan's precedent at its `:1283`) is not taken because each task's `--check`
evidence is its own acceptance and the conductor never reads a transcript; it would also leave the
tool's default dangerous for every later lane. **Residual, recorded:** a deliberate `--apply` by an
implementer is prevented only by the dispatch prompt. Cross-lane note: translit's staged deploy runs
`deploy.sh --apply` once this lands (Task 0 reports whether translit has deployed; the contested-surface
row carries it).

**R-5d17. The forward-auth verify endpoint moves from 5d to rung 12.** Revision 2 asked this as a
ruling; it is not an owner fork, because CLAUDE.md's forward-hook rule (2026-09-12) settles it: a hook
"is named in the rung plan and built by the rung that first consumes it, and nothing is exported ahead
of a caller." Reason: nothing consumes the endpoint before rung 12. No adopter exists before launch
(rung 16), and the estate reaches serve through the tunnel as ingress, never through a forward-auth
proxy. Three facts corroborate the rule's answer. The endpoint's answer is driver knowledge the spec's
own "Universality" section gives to rung 12 ("The header name, the injection rules, the bypass paths,
the logout wiring, and the post-login destination live with the driver at rung 12"), so a 5d endpoint
would pick its answer header with no driver to agree with. STATUS carries "no non-admin maps onto a
streamer admin" to rung 12, and a verify endpoint in front of a streamer that auto-creates accounts
from a trusted header opens that path before its guard exists. An unguarded verify endpoint is an
identity oracle (R21's own words), so the move keeps the one new anonymous-reachable identity surface
out of the first production deploy. Rung 12 keeps the full adversarial cycle (CLAUDE.md, "Review
scales to rung risk"), so the endpoint is not reviewed less by moving. What it builds: no verify
endpoint, no forward-auth verify integration class, and no keys (rung 1b's contract forbids a
reserved class whose keys the declaring rung will not build). STATUS's "To rung 12" item gains the
endpoint and its forward-auth doctor proof. Task 8 marks every SEC or UNI finding absorbed only by R21
as **carried to rung 12**, never absorbed. Task 7 states the forward-auth topology as a future
integration with the requirements BS:504-514 fixes. **The errata** (recorded by the close, never an
edit to a ratified document): the spec's 5d row's first clause (`:2519`); 5b-i's moved assertion "a
verify subrequest from such a peer is refused"; BS ladder entry 5's "the forward-auth verify endpoint
with its own trust-boundary documentation" (`BS:707-714`); and BS:510-513's "`dubplate doctor` proves
the boundary by verifying that a spoofed identity header from a non-trusted source is rejected", for
the forward-auth topology (5d's probe proves serve's own strip). Revision 2's Task 4V, the endpoint
built at 5d, is removed with its ceiling variant and its privileged-fields row.

---

## Review focus

5d is auth-risk and deploy-risk, so it keeps the spec's full fresh-context adversarial cycle (R24,
`:2815-2825`): two independent lenses over the whole rung 5 subsystem, `web-auth-security-reviewer`
for security and one `opus` universality reader, each given the spec's threat model (`:1950-2009`)
and Task 8's disposition table, each grading the implementation against the claims rather than against
taste, and each asked for correctness gaps ranked by consequence, not polish. The six focus points the
lenses must each address by name:

1. A configured header changing a trust decision. Every trust decision reads the peer address;
   the client-address header affects only rate-limit keys and logs, from a configured proxy only, and a
   malformed hop cannot steer the scan (Tasks 1 and 3).
2. An identity header surviving ingress on any path, from any peer, in either spelling, including
   through a `Connection` listing (Task 3's probe and the in-process strip tests).
3. `doctor` green on a configuration serve would refuse, or green while the listener passes a
   spoofed header (Tasks 2 and 3).
4. A grant wider or narrower than the writes. The unit's `ReadWritePaths` against every write the
   serve process makes, the leaves' modes against `auth.Open`'s refusal, and the namespace failure when
   a pre-create is missing (Task 5).
5. A secret in the repo, a log, a report, a ledger, or a rendered file with the wrong mode (Tasks 6
   and 10).
6. A deploy that breaks the live shadow, starts the new unit, or writes the box unattended,
   including the rollback's completeness (Tasks 4D, 5, 6, and 10).

---

## Task 0: The staleness pre-task

Model: `opus`, fresh context, dispatched before any implementer. Reads this plan against the tree
5c-iv leaves. **Its output is blockers with evidence, not fixes.** It edits no tracked file. A moved
pointer whose content survives is a note; a pointer whose content is gone is a blocker.

1. Lane 1 free and the tree clean of warm changes the conductor did not author; the lane registry
   names this pass; **5c-iv's close is committed on master, and 6c has merged to master** (conductor
   decision, revision 2.1; a 6c still running is a blocker, never a note). Report HEAD, 5c-iv's close
   commit, and 6c's merge commit.
2. 5c-iv's handoff to 5d (5c-iii plan Task 11, `:1768-1777`, and 5c-iv's HISTORY entry): every
   item it names for 5d, and whether any lands outside this plan's tasks. The clean-build row
   (R-5d13) must be among them.
3. What 5c-iii and 5c-iv changed in 5d's files: `cmd/dubplate-server/serve.go` (the shell
   construction at `:236-315`, in particular the zero `proxyConfig` at `:244`, the strip at `:245`,
   `clientIP` at `:283`, and `methodOnly` at `:380-389`), `cmd/dubplate-server/web.go`
   (`newWebRoutes`'s signature at `:218`, its logger parameter, and its side-effect freedom),
   `internal/serve/` (`clientaddr.go`, `strip.go:112-135,199-215,455-475`, `serve.go:84-116`,
   `prefix.go:130-134`, `server.go:52-56`), `cmd/dubplate-server/doctor.go` with `doctor_test.go` (the
   row order at `:38-127`, the constants `storeRowLines` and `backupDoctorRowsOff`, `doctorEnviron` at
   `:908-916`), and `check_test.go`'s `baseEnviron` and `authEnviron` (whether either sets
   `DUBPLATE_SERVE_PORT`).
4. What lane 2 merged since `4c766c3`: 6b-ii-b (restore) and the translit pass each may have edited
   `doctor.go`, `doctor_test.go`, `main.go`, `scripts/deploy.sh` (translit's staged transition, which
   moves the `/var/lib/dubplate` pre-create block ahead of the binary step and stages the binary), and
   `scripts/check.sh`. Report every doctor row and line-count constant they added, whether translit's
   `unidecode/` leaf is pre-created by `deploy.sh` and granted to any unit (if so it joins
   `DUBPLATE_SERVER_LEAVES` with its own mode argued), any new unit under `systemd/` whose `ExecStart=`
   names `dubplate-server` (R-5d11 must cover it), and **whether translit's deploy has run on the box**
   (the binary stamp), since R-5d16 changes its apply command. **If translit has merged**, report the
   exact `--check` lines its deploy steps print (read from its `deploy.sh`, never run against the box):
   the staged `.new` binary rsync, the smoke test on the staged path, the `systemd-run` table install
   (a PyPI download from the box), the `mv` swap, and the `/var/lib/dubplate/unidecode` leaf's
   pre-create. That report fills Task 10 step 6's "if translit merged" block.
5. 6c as merged (`docs/superpowers/plans/2026-09-25-rung-6c-routed-schema-settings-home.md`,
   revision 2.2), and whether `internal/config` has changed since `4c766c3`. 6c's value validator,
   built for store-homed settings, changes nothing here (see "Global constraints"); report its shape. 6c's Task 5 edits `doctor.go`,
   `check.go`, and their tests, so **re-read, as 6c left them:** `doctor.go`'s row order (the backup
   rows' position, after which R-5d5 prints the four new rows, and any settings rows 6c added),
   `doctor_test.go`'s line-count constants (`storeRowLines`, `backupDoctorRowsOff`, and any 6c added,
   beside which `probeRowLines` lands), and `check.go` (every fixture or helper Tasks 1 to 3 and 6
   reuse). Every pointer into those files that 6c moved is re-cited in the report.
6. `scripts/check.sh`: the hard-required tool list (`:138-157`), the systemd leg (`:631-643`), the
   state-dir leg (`:645-658`), and the heavy gate's current wall time and peak memory.
7. `scripts/deploy.sh`: the argument parse (`:81-95`), `UNIT_FILES` (`:59`), the render
   (`:240-247`), the first remote write (`:252`), the stamp and binary step (`:334-366`), the env
   install (`:407-414`), the state-dir pre-creates (`:436-505`), `DUBPLATE_LIB_DIR` (`:502`), and
   `TIMER_UNITS` (`:526`, confirming `dubplate-shadow.timer` is not in it).
8. `env/dubplate-server.env.template` still carries the ten rows read at `4c766c3` and nothing of the
   magic-link set but `DUBPLATE_IDENTITY_SOURCE`.
9. Age-store names, name-only: `DUBPLATE_SMTP_HOST`, `_PORT`, `_USER`, `_FROM`, and `_PASSWORD` are
   present in `~/.local/secrets`; `DUBPLATE_MAGIC_LINK_SECRET` and `DUBPLATE_SESSION_SECRET` are absent.
   Never print a value.
10. Box, read-only (`ssh -o BatchMode=yes root@5.78.148.42`), each compared with "Box facts":
    systemd 257; `music` uid 2000; `/var/lib/dubplate` holding only `shadow` (plus `unidecode` if
    translit deployed) and both modes; nothing on 8080; `/etc/dubplate/server.env`'s row names; the
    installed binary's `version` stamp; every timer in `TIMER_UNITS` plus `dubplate-shadow.timer`
    enabled; the installed beets version (`runuser -u music -- env HOME=/srv/music
    /srv/music/.local/bin/beet version`); `md5sum` of `/opt/dubplate/compose.yaml` and every file under
    `/opt/dubplate/config/` against master's; and the row **names** of `/etc/dubplate/env` against the
    names master's `env/dubplate.env.template` renders (never a value). Any drift beyond "Box facts"
    is reported for the Task 10 expected set.
11. Every other `file:line` this plan cites resolves to the content described, including the spec
    pointers in "Sources" and `internal/auth/store.go:32,53-71`, `internal/auth/roster.go:318-319`,
    `internal/auth/handlers.go:189-197`, `internal/config/registry.go:213`,
    `internal/config/keys.go:184-190`, `internal/config/config.go:301,408-427` (at `593d7fd`),
    `cmd/dubplate-server/main.go:108-117`, `cmd/dubplate-server/backup.go:40-54`, and
    `_salvage/cmd/dubplate/pingfail.go:64`.
12. The spec's staleness questions (`:2641-2700`) that still apply to 5d: the state-dir check still
    hard-coded; `doctor` still deriving registry rows; `.gitleaksignore` still empty or its appends named.

---

# Pass 5d-i: the build

## Segment A

## Task 1: The trusted-proxy integration class and the address grammar

Model: `opus` (the trust input to every rate-limit key; novel validation the plan specifies by outcome).

Files:
- Create: `cmd/dubplate-server/proxy.go`, `cmd/dubplate-server/proxy_test.go`
- Modify: `internal/serve/clientaddr.go`, `internal/serve/clientaddr_test.go`, `internal/serve/strip.go`
  (the proxy set's type only), `internal/serve/strip_test.go` as the type change requires,
  `cmd/dubplate-server/serve.go` (the shell's construction), `cmd/dubplate-server/main.go` (one
  registration line), `cmd/dubplate-server/serve_test.go`

Outcomes:
- The `trusted proxy` integration class per R-5d1, registered from `newProductionRegistry`.
- A resolver in `proxy.go` that reads the class from a `*config.Config` (`IntegrationEnabled`,
  `IntegrationValue`, as `backup.go:61-62` does) and returns a `serve.TrustedProxyConfig`, applying
  R-5d2 and R-5d3 and joining every refusal. Serve startup calls it before any listener opens; Task 2's
  doctor row calls the same function.
- `TrustedProxyConfig` keyed by `netip.Addr`, `PeerAddress` compared after parse and unmap, and
  `ClientIP` applying R-5d4. The exported surface does not grow except as the new resolver needs;
  `NewTrustedProxyConfig` takes parsed addresses.
- `newServeShell` builds the strip and every handler's `ClientIP` from the resolved configuration,
  keeps `serve.FrontExemption{}`, and its doc comment's "No trusted proxy is configured yet" paragraph
  (`serve.go:239-242`) is replaced by the real contract.

Acceptance criteria:
- A table test over the resolver covers each R-5d2 and R-5d3 state by name: flag absent (mode direct),
  flag `false` (direct), flag `true` with each Required key absent and empty (the `Load` report), flag
  absent with `ADDRS` set (the "set without" report), flag `yes-please` (the `ParseBool` report), each
  refused address shape (hostname, `192.0.2.0/24`, `fe80::1%eth0`, `0.0.0.0`, `::`, an empty entry
  from `a,,b`), each refused header (`Cookie`, `Forwarded`, `Remote-User`, `Remote_User`, `X Bad`), a
  two-bad-entries value naming both, and a valid `192.0.2.10, ::ffff:198.51.100.7` with `Cf-Connecting-Ip`
  yielding a set whose second member reads `198.51.100.7`.
- `ClientIP` rows, each with its fixture peer and header stated first: a trusted peer with
  `203.0.113.9, 192.0.2.10` (proxy rightmost) yields `203.0.113.9`; a trusted peer with
  `203.0.113.9, garbage` yields the peer; a trusted peer with `garbage, 203.0.113.9` yields
  `203.0.113.9`; a trusted peer with `203.0.113.9:4711` yields `203.0.113.9`; a trusted peer with
  `[2001:db8::1]:443` yields `2001:db8::1`; a trusted peer with `[2001:db8::1]` (no port) yields the
  peer; a trusted peer with a zoned hop yields the peer; an untrusted peer with any header yields the
  peer; a trusted peer with an IPv4-mapped hop yields dotted form; two header lines read as one chain
  (`clientaddr.go:96-101`).
- A shell built by `newServeShell` from an environment enabling the class, driven through `httptest`
  with `RemoteAddr` set to the configured proxy, keys the link-request limiter on the header's client:
  the per-client limit trips for that client and not for a second client behind the same proxy; the
  same request from an unconfigured peer keys on the peer. Fixture state: a fixture roster and the
  environment's own mailer (no stub is needed: `allowLinkClient` runs before the form parse and before
  any send, `internal/auth/handlers.go:189-197`), the limiter's per-client limit set to 1, each POST
  carrying `Origin` equal to the fixture public URL and `Sec-Fetch-Site: same-origin` so the origin
  check admits it, and a body whose address does not parse, so the request consumes the client's bucket
  and sends nothing.
- The shell's strip carries the zero front exemption: a spy route registered on the returned shell
  (legal before `Handler()`, `internal/serve/serve.go:84-90`) receives a request from the configured
  proxy bearing `Remote-User` without it.
- `bash scripts/check.sh` green, heavy lane.

Named mutations:
1. Continue the scan past a malformed hop → the `203.0.113.9, garbage` row returns `203.0.113.9` and
   fails.
2. Skip `Unmap` in the resolver → the `::ffff:198.51.100.7` resolver row's second member reads the
   mapped form, and a limiter row keyed from peer `198.51.100.7` (configured as `::ffff:198.51.100.7`)
   falls back to the peer and fails.
3. Drop the CIDR refusal → the `192.0.2.0/24` row fails.
4. Build the shell with the zero `TrustedProxyConfig` again → the limiter row fails.
5. Drop the `ParseAddrPort` second parse → the `203.0.113.9:4711` row yields the peer and fails.

0.8 hook: lands hook 19's client-IP half (a direct mode that trusts no forwarded header, chosen
explicitly), since this class is its first consumer; the listener half stays rung 12's. Hook 18
(description and sensitive mark on every key) is neutral: `IntegrationClass` carries neither, and the
field is rung 14's to add. Deliverable count: 4 (the class, the resolver, the netip set with the
hop rule, the shell wiring).

---

## Task 2: Doctor's client-IP, web, and bind-address rows, and the green baseline

Model: `sonnet`. Depends on Task 1.

Files:
- Create: `cmd/dubplate-server/doctorprobe.go` (the three offline rows' status functions and Task 3's
  row), `cmd/dubplate-server/doctorprobe_test.go`
- Modify: `cmd/dubplate-server/doctor.go` (the four rows' print calls after the backup rows and the
  exit condition), `cmd/dubplate-server/doctor_test.go` (`probeRowLines`, R-5d5),
  `cmd/dubplate-server/check_test.go` (the shared fixture's port)

Outcomes:
- `client ip`: `direct` or `trusted proxy: <header> from <n> address(es)`, or `invalid: <joined
  refusal>` (failing), from Task 1's resolver; `not checked` when `Load` failed (R-5d5). It names no
  address, since a proxy address is deployment topology a terminal need not show; the count is enough
  to see the mode.
- `web` per R-5d6, with a discarding logger.
- `bind addresses` per R-5d7's five states.
- `boundary probe` prints `not probed` in this task (Task 3 replaces it), so the line count is final
  here.
- `newDoctorCmd`'s exit condition (`doctor.go:129`) gains the three rows' failures; its doc comment's
  exit-status sentence lists them.
- The green baseline. Every environment a doctor test passes to `newDoctorCmd` carries
  `DUBPLATE_SERVE_PORT` set to a port the test binary reserved once (listen on `127.0.0.1:0`, record
  the port, close) and never listens on, so no doctor test dials the host's real `127.0.0.1:8080`
  after Task 3. A serve test that asserts the registered default port keeps an environment without it.
  `doctorEnviron` (`doctor_test.go:908-916`) is the green baseline every exit-1 row perturbs once.

Acceptance criteria:
- Baseline: one test runs doctor on `doctorEnviron` and asserts exit 0 with every row's status
  stated, including `client ip` `direct`, `web` ok, `bind addresses` `127.0.0.1 present`, and
  `boundary probe` `not probed`.
- Each row's states, each by its own test, each the baseline plus exactly one perturbation, with the
  fixture stated first: `client ip` trusted (exit 0), and invalid (a CIDR entry, exit 1); `web` an
  unknown theme match value (exit 1, naming the accepted set), a missing operator CSS file (exit 1),
  a neutral branding value (exit 0, and the line count unchanged, proving the web row's log record did
  not reach stdout), and `not checked` when `DUBPLATE_SITE_NAME` is unset (exit 1 from `Load`, and the
  row's own text proves it added no second reason); `client ip` `not checked` in that same `Load`
  failure; `bind addresses` with `127.0.0.2` present (loopback, exit 0), `192.0.2.1` failing (exit 1),
  `127.0.0.1,,::1` invalid (exit 1), and, through the status function, an `InterfaceAddrs` error
  yielding `not checked: <error>` naming `RestrictAddressFamilies`, failing.
- `TestDoctorCmd` and `TestDoctorCmd_LineCountFollowsRegistry` pass with `probeRowLines = 4`
  hard-coded; `TestDoctorCmd_AbsentStoreCreatesNothing` still passes (the web row writes nothing).
- `bash scripts/check.sh` green, heavy lane.

Named mutations:
1. Print the web row without calling `newWebRoutes` (the old unjudged value) → the unknown-theme row
   exits 0 and fails.
2. Treat a loopback address as absent → the `127.0.0.2` row fails.
3. Derive `probeRowLines` from the rows' own slice → replace with the literal and show the derivation
   mutation (one row dropped) passes the derived count and fails the literal one.
4. Swallow the `InterfaceAddrs` error as an empty list → the error row reads `not on any interface`
   and fails.

0.8 hook: neutral. The 0.8 slice 6 admin credential list names "a `doctor` row [that] names the gap
in the default mode"; `client ip` is that row's present form, and 0.8 extends its text, never its
mechanism. Deliverable count: 4 (three rows and the green baseline).

---

## Task 3: The boundary probe route and its doctor row

Model: `opus` (a live-dial proof that must be able to fail; the spec rejected an in-process probe for
exactly that reason).

Files:
- Create: `internal/serve/probe.go` (the route handler, the probe client's request builder, and the
  header set they share), `internal/serve/probe_test.go`
- Modify: `cmd/dubplate-server/serve.go` (mount the route), `cmd/dubplate-server/doctorprobe.go`,
  `cmd/dubplate-server/doctorprobe_test.go`

Outcomes:
- The route and client per R-5d8, the row's states per R-5d9. The probe request carries every group 1
  spelling in both forms, every group 2 name, and the sentinel listed in `Connection`, each added with
  `Header.Add`, from one request builder that doctor and every proof use.
- `doctor` dials each configured bind address on the configured port.
- The route's peer rule reads the same resolved bind list serve listens on and the same resolved
  trusted-proxy set, passed at construction; the route never reads configuration itself.

Acceptance criteria:
- Wire-level, through the shell: the route mounted on the real shell (`NewShell` with the real
  `Strip`) served by `httptest.NewServer`, probed by the shared builder from `127.0.0.1`: all counts
  zero.
- Wire-level, without the strip: the route handler alone served by `httptest.NewServer`, probed by
  the same builder: the identity count **equals** the number of distinct canonical names the builder
  sent (each `_` spelling arrives as its own canonical key, R-5d8), the forwarded count equals
  group 2's distinct canonical count, and the sentinel count is 1. This proves the route sees what the
  strip removes over the same parser the live dial crosses, and that Go's server removes no
  `Connection`-listed header, so a zero through the shell is evidence.
- End to end in `cmd/dubplate-server`: the real shell from `newServeShell` served by `serve.Run` on
  `127.0.0.1` at a port the test reserved, `doctor` run against `doctorEnviron` naming that port, and
  the row reads `refused` with exit 0. Fixture state: the listener answering, asserted by one plain
  `GET` to the route before doctor runs.
- No listener: the baseline's closed port → the row reads `no listener` and doctor exits 0.
- A listener that did not strip: a test server on the port answering the route's JSON with a
  nonzero identity count → the row reads `passed N identity header(s)` and doctor exits 1.
- A wedged listener: a test server that accepts and never answers → `unexpected answer: timeout`,
  exit 1, within the response bound.
- An absent bind address: `DUBPLATE_SERVE_BIND_ADDRS=192.0.2.1` → the probe row reads `not probed
  (not on any interface)` and the only failure named is the bind row's.
- A trusted local address: trusted proxy enabled with `127.0.0.1` as a proxy and `127.0.0.1` as
  the bind address → the row reads `not probed: every local address is a trusted proxy`, adding no
  failure.
- Peer and method, each 404 byte-equal to an unknown path's, by handler call with `RemoteAddr` set:
  a `GET` from `198.51.100.7` (not local); a `POST` from `198.51.100.7`; a `POST` from `127.0.0.1`; a
  `GET` from `127.0.0.1` when `127.0.0.1` is a configured trusted proxy; and a `GET` from a configured
  non-local trusted proxy `192.0.2.10` whose client-address header names `127.0.0.1`. No 404 carries
  an `Allow` header.
- `bash scripts/check.sh` green, heavy lane.

Named mutations:
1. Mount the route outside the strip (a root-level handler) → the end-to-end row reads a nonzero count
   and fails.
2. Report `no listener` as a failure → the no-listener row exits 1 and fails.
3. Accept any peer at the route → the `198.51.100.7` rows get 200 and fail.
4. Decide the peer through `ClientIP` instead of `PeerAddress` → the `192.0.2.10`-naming-`127.0.0.1`
   row gets 200 and fails.
5. Drop the trusted-set exclusion → the trusted-loopback row gets 200 and fails.
6. Count by indexing the header map with the list's spellings as written → the wire-level no-strip
   equality row undercounts (every `_` spelling reads zero) and fails.
7. Make the strip delete only the dash forms of group 1 while `identitySpellings` still returns both
   → the end-to-end row reads `passed N identity header(s)` and fails.

0.8 hook: neutral. Deliverable count: 4 (the route, the client, the row, the proofs).

---

## Segment B

## Task 4D: The deploy script's safety

Model: `sonnet`. Depends on Task 0 only; runs first in Segment B because Task 5 edits the same script.

Files:
- Modify: `scripts/deploy.sh`, `docs/bringup-runbook.md` (the `deploy.sh` apply invocations only)

Outcomes:
- R-5d16: no flag and `--check` report; `--apply` applies; both flags together are refused naming
  both; the usage comment and the banner (`deploy.sh:249`, "(--check: nothing will be applied)") name
  the mode.
- R-5d13, moved from Task 5: the two refusals, `npm ci --prefix web`, `npm run build --prefix web`,
  and `make build-linux` with the stamp run after the render and before the repo-config rsync, in both
  modes. The binary's rsync, the smoke test, and every later step stay in their order.
- `docs/bringup-runbook.md`'s apply invocations gain `--apply`; its `--check` invocations are unchanged.
- The header comment names every remote write the apply performs, so a reader of the script sees the
  whole-estate scope Task 10 must plan around (config rsync with `--delete`, both env installs,
  `contributors.yaml`, compose `up -d`, `enable --now` on `TIMER_UNITS`, the beets install).

Acceptance criteria:
- `shellcheck` clean.
- Every run below happens in the detached scratch worktree at the task's commit ("Global
  constraints"), and the dispatch names the two network exceptions.
- The default is safe, proved without a writable target: `bash scripts/deploy.sh 192.0.2.1` (a
  documentation address that routes nowhere, so no mode can write anything) with no flag prints the
  check banner and reports each step as "would"; the same with `--check` prints the same; `--apply
  --check` is refused before the render. Each output's banner line and exit status are in the report.
- The refusal precedes every remote write: in the scratch worktree, with a scratch edit to a
  tracked Go file, and separately with an untracked file under `web/src/`, `bash scripts/deploy.sh
  --check 192.0.2.1` refuses naming the path, and its output contains no `== repo config` line; both
  scratch edits are removed before the next run.
- The local builds precede every remote write: in the clean scratch worktree, `bash
  scripts/deploy.sh --check 192.0.2.1` prints the `npm ci`, web build, and `make build-linux` lines
  before `== repo config`; the report quotes the full step-header list. (The real box is never the
  target here; Task 10 step 6 is its one real `--check`.)
- `bash scripts/check.sh` green, heavy lane.

Named mutations:
1. Restore apply as the no-flag default → the no-flag `192.0.2.1` run prints the apply banner and
   fails the criterion (safe: the target routes nowhere).
2. Move the clean-tree refusal after the repo-config rsync → the dirty-tree run prints `== repo config`
   before the refusal and fails.

0.8 hook: neutral. Deliverable count: 3 (the default flip, the refusal and build ordering, the
runbook lines).

---

## Task 5: The serve unit, the leaves, and the state-dir check

Model: `sonnet`. Depends on Task 1 (the binary the unit runs must carry the class) and Task 4D (the
same script).

Files:
- Create: `systemd/dubplate-server-serve.service`
- Modify: `scripts/deploy.sh`, `scripts/check.sh`

Outcomes:
- The unit per R-5d10, with a comment per inherited block stating its reason, on the shadow unit's
  precedent, and a derivation comment above `ReadWritePaths=` naming the two leaves' sources
  (`auth.StoreDir`, `internal/auth/store.go:36-39`, and the template's roster row's directory).
- `deploy.sh`: the unit joins `UNIT_FILES` (rsynced and daemon-reloaded, never enabled or started, and
  `TIMER_UNITS` unchanged); `DUBPLATE_SERVER_LEAVES` and its one pre-create loop per R-5d11, with the
  comment naming every unit that depends on each leaf and why `shadow` is skipped.
- `scripts/check.sh`: the generalized state-dir leg per R-5d11 and the hard-required `systemd-analyze`
  per R-5d12.
- Evidence, recorded in the task report, not a gate leg. A user-level transient unit
  (`systemd-run --user --wait --collect`) carrying the new unit's hardening properties **minus
  `User=`** (a user manager cannot switch users), with the two scratch leaves **and a copy of the built
  binary under `/dev/shm/<unique>`**. Scratch under `/tmp` or `/var/tmp` is excluded because
  `PrivateTmp=` gives the unit its own (`man systemd.exec`: "/tmp/ and /var/tmp/ ... not shared"), so an
  existing leaf there fails with the missing-leaf text; the home directory is excluded because
  `ProtectHome=` makes "/home/, /root, and /run/user ... inaccessible", so the absent-leaf case
  succeeds. The mechanics lens measured all three placements on this workstation (systemd 259). **The
  positive half runs first in that same location:** `dubplate-server member add --admin` against a
  scratch roster path, then `dubplate-server serve` for a few seconds against a fixture environment;
  the roster file appears and the auth leaf gains its subdirectories. **Then one leaf is removed** and
  the same run fails with systemd's own text: "Failed to set up mount namespacing: <path>: No such file
  or directory" and "status=226/NAMESPACE" (`man systemd.exec`'s exit-code table: "226 EXIT_NAMESPACE
  Failed to set up mount, UTS, or IPC namespacing"). The box repeats the positive half in Task 10.

Acceptance criteria:
- `systemd-analyze verify` over `systemd/*` clean modulo the expected box-only `ExecStart` paths
  (the leg's existing filter, `check.sh:634`).
- The state-dir leg passes the real tree and fails by name on each of eight seeds: `roster` removed
  from `DUBPLATE_SERVER_LEAVES`; `auth` removed; a `-/var/lib/dubplate/roster` entry; an entry outside
  the state directory (`/var/lib/other/x`); a nested entry (`/var/lib/dubplate/auth/nonces`); every
  unit's `ExecStart` pointed elsewhere (no greenfield unit found); the loop iterating a literal `auth
  roster` list instead of the array; and the shadow pre-create line deleted.
- `shellcheck` clean on both scripts.
- `bash scripts/deploy.sh --check 192.0.2.1`, from the scratch worktree at the task's commit (never
  the real box, "Global constraints"), reports the two new pre-creates (the `install -d` "would run"
  lines for `auth` and `roster`, and no such line for `shadow`), the units step header naming the new
  unit (the header prints every `UNIT_FILES` entry, `deploy.sh:264`), and the web build; the report
  quotes those lines. A dry-run `rsync` against an unreachable host lists no files, so Task 10 step 6's
  real `--check` is where the unit shows as added.
- The hard requirement is proved, not mutated: with `systemd-analyze` hidden (a shim directory of
  symlinks to every `/usr/bin` entry except `systemd-analyze`, with `/usr/bin` and `/bin` dropped from
  `PATH` and every other entry kept), `bash scripts/check.sh` fails naming the tool before any leg runs.
- The evidence run's two outcomes are in the report with their exit codes, systemd's text, and the
  `/dev/shm` placement.
- `bash scripts/check.sh` green, heavy lane.

Named mutations: the eight state-dir seeds above, each failing the gate by name.

0.8 hook: neutral. 6d adds the `store` leaf by one array word (rung 6 spec criterion 14) and
retires `roster` (erratum 8). Deliverable count: 4 (the unit, the leaves and loop, the state-dir
leg with the hard requirement, the namespace evidence).

---

## Task 6: The environment template and the rendered-doctor proof

Model: `sonnet`. Depends on Tasks 1, 2, 3 (doctor's final row set) and Task 5.

Files:
- Modify: `env/dubplate-server.env.template`, `cmd/dubplate-server/templatekeys_test.go`

Outcomes:
- The rows per R-5d14, each with a comment giving its row state and reason, and the header's
  secret paragraph rewritten.
- A gate-resident test that parses the greenfield template the way `deploy.sh`'s `render_template`
  does (`deploy.sh:183-205`'s documented states: a literal keeps its value, including a literal
  containing `=`; a bare row takes a fixture value; a blank or commented row is omitted; a line with
  no `=` fails the test), each state asserted by its own table row. **Fixture values replace the value
  of a template row that exists and never append a row**, so deleting a row leaves no substitute. The
  test overrides exactly the rows `Load` and `doctor` stat or open (the task quotes which, from
  `doctor.go` and the rows' status functions: at least the state directory, the roster file, and the
  bundle), and sets `DUBPLATE_SERVE_PORT` to the baseline's closed port. The fixture roster names one
  admin at mode 0600. The test runs `doctor` against that environment and asserts exit 0 with every
  row's status stated, `boundary probe` reading `no listener`. Fixture values are obviously not
  credentials (`example.test` addresses, a 32-byte repeated-letter secret) and trip no gitleaks rule.
- The test's doc comment says that Task 10 step 6's `deploy.sh --check` is the render's only real
  proof and this test is its parity model.

Acceptance criteria:
- The rendered-doctor test exits 0 with the fixture roster present, and exits 1 naming the `roster`
  row when the roster file is absent (the runbook's "doctor before the bootstrap" state, stated in the
  test's name).
- Deleting the `DUBPLATE_ROSTER_FILE` row fails the test naming the magic-link source's missing key
  (nothing substitutes for it); deleting the trusted-proxy flag row leaves it passing (the class is off
  by default), which the test asserts deliberately so a future reader does not mistake the explicit
  `false` for a requirement.
- Under Ruling 2's no: the two secrets are commented, and the test instead asserts exit 1 naming
  exactly those two keys.
- `TestTemplateKeysClassifiedAgainstProductionRegistry` and `TestTemplateKeysCoverEveryOwnedNowRow`
  pass unchanged.
- Under Ruling 2's yes, `bash scripts/deploy.sh --check 192.0.2.1`, from the scratch worktree at the
  task's commit, aborts at the render naming exactly the two
  secrets and nothing else, before any remote step; the report quotes the name list. (Under no, it
  renders and reports every step.) From here until the Task 10 mint, `--check` stops at the render.
- `bash scripts/check.sh` green, heavy lane.

Named mutations: delete the `DUBPLATE_ROSTER_FILE` row from the template (the test fails naming
the key, proving no appended fixture row substitutes); drop the fixture roster's admin flag (the roster
row fails, exit 1).

0.8 hook: neutral. Deliverable count: 3 (the rows, the header, the rendered-doctor proof).

---

## The 5d-i close (conductor's actions)

1. One `code-simplifier` round over Tasks 1 to 6's Go and shell, pasting
   `.claude/instructions/simplifier-brief.md` verbatim, reviewed by one `diff-reviewer`, gate green.
2. Two `go-architecture-reader` dispatches (`internal/serve`, `cmd/dubplate-server`), no plan in
   context; findings routed to the architecture cleanup pass, or to Task 9 if one is a correctness gap.
3. `git diff --stat <5d base>..HEAD -- internal/config` empty.
4. The 5d-i HISTORY entry and a STATUS handoff carrying 5d-ii's resume prompt, both budgets scored,
   the lane registry updated; close the session.

---

# Pass 5d-ii: the review, the sitting, and the close

A fresh session, from the STATUS resume prompt. Its first action re-reads this plan from "Pass 5d-ii"
and the 5d-i HISTORY entry; no Task 0 re-run, since 5d-i's close is its base.

## Segment C

## Task 7: The server runbook and the trust-boundary documentation

Model: `sonnet`, Google developer documentation standard (`writing-voice`). Independent of Tasks 1 to 6
by files; it describes their behavior, so it runs after them.

Files:
- Create: `docs/server-runbook.md` (R-5d15)

Outcomes, one section per read, each opening with a one-paragraph brief:
- **Install order** (spec `:1771-1774`): install, render the environment, run `dubplate-server member add
  --admin` as the service user through `systemd-run` carrying the unit's properties (the shadow unit's
  operator rule, `dubplate-shadow.service:42-46`), run `doctor`, start the unit. **Doctor is run with the
  unit's filesystem and privilege properties but without `RestrictAddressFamilies`**, and the section
  says why in one sentence: the bind-addresses row reads the interface table over netlink, which that
  restriction denies (R-5d7).
- **The doctor rows**: what each new row proves and cannot, `no listener` before start, every probe
  state of R-5d9, and the boundary probe's limit; and 6c's settings rows, including that "env `<v>`
  seeds at serve's next start" is not yet true until 6d wires the seed into `serve` (the carry 6c's
  plan hands to 5d and 6d).
- **Mail**: two worked SMTP examples, one with no Cloudflare account (a mailbox provider's submission
  port with STARTTLS) and one through Cloudflare Email Service's SMTP submission endpoint (R7), and the
  outbound-465 note; the mail-ceiling lockout lever and the kind-mode per-client bound (5c-i's carry,
  `HISTORY:1729-1731`).
- **The trust boundary**: the peer address and the reported client IP (R27); trusted-proxy
  configuration with the Authelia and Miniflux corroboration the spec names (`:894-901`); **the proxy's
  duty to overwrite or append the configured header**, since a header it passes through unchanged is
  client-chosen; that every group 2 name, not only the configured one, survives from a trusted peer
  (nothing reads them at 5d); a port-bearing hop's handling (R-5d4); **the probe route's residual under
  a same-host proxy the operator did not configure as trusted** (R-5d8); the HTTPS precondition (R38);
  and the forward-auth topology stated as a future integration built at rung 12 (R-5d17), with its
  requirements from BS:504-514: the proxy strips the published set before the verify subrequest and
  before the streamer, and the streamer trusts only that proxy.
- **Browser and presentation notes**: Ruling 6's old-browser `Referrer-Policy` residual; a branding
  color applies in both color schemes (identity spec amendment A4).
- **Backup rows** carried from 6a and 6b-i: `KillMode=mixed` for a unit that runs rclone, rclone
  v1.74.3 or later, a crypt remote, and the exit-7 remedy (`HISTORY:884-887`).

Acceptance criteria: every section exists with its brief; every command in it was run by the
implementer against the workstation or a fixture and matches the binary's `--help` (the report lists
each); no estate hostname or personal address appears; one `prose-voice-reviewer` read (the
conductor's dispatch) and its fold; `bash scripts/check.sh` green (the gate reads no Markdown, so this
proves the commit is clean).

0.8 hook: neutral. Deliverable count: 3 (the runbook, the trust-boundary section, the prose
fold).

---

## Task 8: The SEC and UNI disposition table

Model: `opus`. Independent of every code task by files, but its `file:line` and test-name citations
for the proxy, hop, and probe findings do not exist until Tasks 1 to 3 land, and the 5d-i simplifier
round moves lines after that. **It runs after the 5d-i close**, in Segment C, and must land before the
adversarial cycle. If the conductor drafts it earlier as a document lane, it re-greps every citation
at the 5d-ii base before the cycle and the report says so.

Files:
- Create: `docs/superpowers/research/2026-09-25-rung-5-sec-uni-disposition.md`

Outcomes: one row per finding, all 50 (SEC's 6 blockers, 13 majors, and 12 minors; UNI's 3
criticals, 6 highs, 7 mediums, and 3 lows, per the spec's count at `:2003-2006`): the finding's own id
and a one-line statement; the birth-requirement clause or ruling that absorbed it (a spec pointer); the
code and the test that prove it at HEAD (`file:line` and test name, grepped, never recalled); and a
status of absorbed, absorbed with a recorded residual (the residual stated), or carried (to which rung,
citing the carry). **Every finding absorbed only by R21 (the verify endpoint) is carried to rung 12,
never absorbed** (R-5d17), UNI's High "the forward-auth shape ... is never designed" among
them. A closing section lists the residue in one place, including the plan's own recorded residuals:
the probe route under an unconfigured same-host proxy, the HMAC keys and SMTP password in every
`music`-uid verb's environment (carried to rung 16), group 2 surviving from a trusted peer (carried to
rung 12: strip every group 2 name but the configured header from a trusted peer), the SMTP transport
not exercised on the box (carried to rung 16), and a deliberate `deploy.sh --apply` by an agent
prevented only by its dispatch.

Acceptance criteria: the row count is 50 and each review text's own finding count matches its
rows; every cited test exists (`go test -list` output in the report); every residual cites where it is
recorded; no finding is marked absorbed without a test or a "by design" pointer, and **every by-design
pointer cites the spec clause or ruling that ratified the design**.

0.8 hook: neutral. Deliverable count: 3 (the table, the residue list, the verification).

---

## The adversarial cycle (conductor's actions, between Tasks 8 and 9)

1. Two fresh-context lenses, dispatched in one message, neither given this plan:
   - **Security**: `web-auth-security-reviewer` (`claude-opus-5-5`, effort `high`) over the whole rung 5
     subsystem at HEAD: `internal/serve`, `internal/auth`, `internal/identity`, `internal/notify`,
     `cmd/dubplate-server`'s serve, auth, member, doctor, and proxy files, the new unit, `deploy.sh`'s
     new steps and default, the template, and `web/`'s three pages.
   - **Universality**: one `opus` reader (effort `high`), the lens that caught UNI's three criticals,
     over the same tree plus the runbook, asking whether a stranger with no Cloudflare account, a
     different mail provider, a different proxy, or IPv6 can install and operate it from the runbook
     alone, and whether any estate value leaks outside `internal/config`, the template, the unit, and
     `deploy.sh`.
   Each receives the spec's threat model (`:1950-2009`), Task 8's table, the six review-focus points,
   and the list of prior review records, including the three plan lenses' "Verified sound" sections,
   as the do-not-re-litigate baseline (a finding on a "verified sound" item needs new evidence). Each
   returns correctness gaps ranked by consequence, with `file:line` and a reproduction where one is
   possible.
2. Both reports are committed as
   `docs/superpowers/research/2026-09-25-rung-5d-review-{security,universality}.md`.

## Task 9: The adversarial fold

Model: `opus` (the fold rules on security findings; CLAUDE.md lets a fold refuse a finding whose fix
costs more than the risk it removes).

Files: whatever each accepted finding's fix touches, inside this pass's territory wall; the merged
disposition record `docs/superpowers/research/2026-09-25-rung-5d-review-fold.md`.

Outcomes: every finding from both lenses gets exactly one disposition, folded (with the commit and
the test that fails without it) or refused (with a one-line reason), in the merged record, which
becomes the rung's final do-not-re-litigate list. A finding outside the territory wall is carried to
the rung that owns the surface, named in the record. A fix that would touch `internal/config` is
carried, never made (the 6c lane condition). **A fold that touches `deploy.sh` cannot be exercised by
`--check` before the Task 10 mint** (the render aborts on the two unminted names under Ruling 2's yes),
so it is proved by `shellcheck`, a `192.0.2.1` run from the scratch worktree where the change
precedes the render, and Task 10
step 6's `--check` read, and the record says which.

Acceptance criteria: the record covers every finding in both reports (the counts match); each fold
has a test that fails with the fix reverted, or the proof named above for a `deploy.sh` fold; `bash
scripts/check.sh` green, heavy lane. **If the lenses return more than about six findings needing
code**, the conductor splits the fold into two tasks at a file-disjoint seam rather than growing one
task, and says so at the checkpoint.

0.8 hook: neutral. Deliverable count: up to 4.

---

## Task 10 (under Ruling 2's yes): The first-deploy sitting

**Attended; conductor-run with Geoff present. No implementer.** Every box write happens here and
nowhere else. **The record keeps statuses, exit codes, row names, and step headers, never a member
field, a secret, or an address** (Geoff's address and Navidrome login stay out of every ledger).

Why the order matters. `deploy.sh --apply` swaps the binary (`:337-343`, or translit's staged
`mv`) before it installs `server.env` (`:407-414`); the old binary refuses the new env's keys and the
new binary refuses the old env (`Load` requires the selected source's keys and rejects every unknown
`DUBPLATE_*` name), so a shadow firing between the two, or an abort between them, fails every five
minutes unmonitored ("no monitoring at rung 4", `dubplate-shadow.service:57`). "Redeploy the
previous commit" is no rollback: every commit from 5b-i to 5d renders `DUBPLATE_IDENTITY_SOURCE` without
secrets. The sitting therefore stops the shadow timer, keeps the working binary and both env files as `.prev`
copies, and restores all three together if anything fails. `dubplate-shadow.timer` is not in `TIMER_UNITS`
(`deploy.sh:526`), so the apply's `systemctl enable --now` leaves it stopped. The import timer is
**not** stopped: the apply's `enable --now` would restart it, and because it is `Persistent=true` (`man
systemd.timer`: "the service unit is triggered immediately if it would have been triggered at least
once during the time when the timer was inactive") a stop across a quarter-hour would fire an import
right beside the beets install. The apply is timed against its schedule instead. A pause of a few
minutes costs the shadow's evidence floor nothing: the floor measures the window from the first
successful pass (`internal/engine/shadow/run.go:89-92`), and a gap is counted only when an oracle report
is swept before the shadow consumes it (`reconcile.go:49,78-81`), which a short attended pause with the
next import timed away does not cause.

1. Preconditions on the workstation: master green at Task 9's commit; the lane registry shows no
   other executor in lane 1 and lane 2 not mid-merge; AC power. **Every `deploy.sh` command in this
   sitting runs from a detached scratch worktree at Task 9's commit** (`git worktree add --detach
   <scratchpad>/5d-deploy <sha>`, "Global constraints"), because the main tree is shared with the
   document lanes and R-5d13 refuses any modified tracked file. `git status --porcelain
   --untracked-files=no` in that worktree is empty. `make build-linux` runs there once now, so step 3
   has a binary, and the worktree is removed after step 15 or after the rollback.
2. The mint (one 1Password approval, Geoff's touch): for each of `DUBPLATE_MAGIC_LINK_SECRET` and
   `DUBPLATE_SESSION_SECRET`, `openssl rand -base64 48 | ~/.dotfiles/scripts/secrets/secret-set.sh
   NAME --stdin` (48 random bytes, 64 characters, above the 32-byte floor, `secretLengthFloor`,
   `cmd/dubplate-server/auth.go:71`; `--stdin` reads with `IFS= read -rsp`, `secret-set.sh:120`, and
   keeps the value out of argv, which `--value` would not). Each is documented in
   `~/.dotfiles/secrets/registry.md` with its scope (the greenfield server's HMAC keys, box-only,
   rendered by `deploy.sh` from the age store; not a Worker secret, so the worker-only exception for
   cairn per-site keys does not apply and neither name joins `sync.sh`'s `WORKER_SECRETS`) and its
   rotation note (the `_PREVIOUS` overlap doubles the count at rotation, spec `:1662-1664`); then
   `sync.sh --verify`. No value is printed.
3. The local doctor, pre-deploy: the template rendered against `~/.local/secrets` into a file of
   rendered rows only, with the scratch state directory, scratch roster (one admin), and the closed port
   **substituted** for the template's rows and the closed `DUBPLATE_SERVE_PORT` **added** (the
   template has no port row, R-5d14), one `NAME=value` per line, then `mapfile -t rows < rendered; env
   -i PATH="$PATH" HOME="$HOME" "${rows[@]}" <worktree>/dubplate-server-linux-amd64 doctor`, so a value
   with a space or a glob character stays one word (never a shell that sourced `~/.local/secrets`, whose roughly 40 incumbent
   `DUBPLATE_*` names, `DUBPLATE_STATE_DIR=/var/lib/musicbox` among them, would trip `Load`'s namespace
   rejection or win over the rendered values). Every row read; exit 0. This proves the estate's real
   SMTP and URL values pass the magic-link source's validation before anything reaches the box. Values
   are never printed; the rendered file lives in the scratchpad and is deleted after.
4. Box reads for the expected set (read-only): Task 0 item 10's reads repeated now, each compared
   with Task 0's report.
5. The quiet-box check (read-only): no `dubplate-*` unit activating (`systemctl list-units
   --state=activating 'dubplate-*'`); no upload in progress. The import-boundary timing is taken at step
   7, directly before the apply, never here.
6. `bash scripts/deploy.sh --check 5.78.148.42` from the scratch worktree, the pass's one real-box
   `--check`, timed with `time`: renders with no unresolved name; the report is read in full against
   **the expected set**, and any line outside it is a stop. **The expected set is derived, not
   recalled:** `git diff --name-only <box stamp> HEAD -- config/ compose.yaml systemd/ env/`, where
   `<box stamp>` is the commit in the installed binary's `version` (`be27434` at "Box facts"), lists
   every repo file the apply can change; each listed path maps to its expected line below, and Task 0
   item 10's md5 reads add any box-side drift the diff cannot see. At `1eeea40` that diff lists
   `config/contributors.yaml`, `config/reject-email.txt`, `config/roster.example.json`, and
   `env/dubplate-server.env.template`; 5d adds `systemd/dubplate-server-serve.service`.
   - Changes expected: the local clean-tree check and builds; under `config/`, every path the diff
     lists, among them `contributors.yaml` and `reject-email.txt` (the `53ecd0a` copy, named to Geoff)
     and the new `roster.example.json` (an example file nothing on the box reads); the unit rsync adding
     `dubplate-server-serve.service` and any other unit the diff lists, and no unit it does not;
     `/etc/dubplate/env` "would change" only if the names-only diff from step 4 explains it (otherwise
     a stop); the binary (the stamp at Task 9's commit); `/etc/dubplate/server.env` (the new rows); the
     `auth` and `roster` pre-creates; the beets install, which upgrades the box from 2.13.1 to the
     pinned 2.14.1 (named to Geoff as a shadow-evidence event: the incumbent importer moves to the
     oracle pin the matcher was re-derived against at `367564c`, and the apply also ships every matcher
     change since `be27434` into the shadow, the retirement of the workaround for beets 2.13.1's
     featuring-credit bug among them; Ruling 2).
   - If translit merged (Task 0 item 4's report selects this block, with the exact lines it
     quoted): the staged `.new` binary rsync in place of the direct binary rsync; the smoke test on the
     staged path; the `systemd-run` table install, which downloads the pinned wheel from PyPI on the
     box; the `mv` swap of the staged binary; and the `/var/lib/dubplate/unidecode` leaf's pre-create.
     If translit's deploy already ran on the box, its leaf and table are no-ops instead.
   - No-ops expected: `compose.yaml` unless the diff lists it; the tmpfiles fragment; the shadow,
     uploads, backup-staging, and musicbox state directories; compose `up -d` (no recreate); `enable
     --now` on `TIMER_UNITS` (already active).
   Geoff approves stopping the shadow timer and the apply.
7. The timing gate, then the window opens. Directly before the window, after Geoff's approval.
   The import timer (`systemd/dubplate-import.timer`) carries `OnCalendar=*:0/15`,
   `RandomizedDelaySec=300`, and `Persistent=true`, so each quarter-hour's import fires at a random
   point up to five minutes after the boundary. The window qualifies only when all of these hold:
   `dubplate-import.service` is inactive; the timer's last trigger (`systemctl show
   dubplate-import.timer -p LastTriggerUSec`) is at or after the most recent quarter-hour boundary, so
   this quarter-hour's import has already fired; the last `dubplate-shadow.service` run started after
   that import finished (its `ExecMainStartTimestamp` against the import's `ExecMainExitTimestamp`),
   so the shadow has consumed its report; and the next boundary is at least **twice step 6's measured
   wall time, and never less than ten minutes**, away. The apply repeats step 6's local builds and
   turns each remote diff into a write, the binary upload and compose the slowest, and its beets step
   runs near the end (`deploy.sh:530-557`), so the whole apply must finish before the boundary. **If
   step 6's wall time exceeds 7.5 minutes, no window qualifies: stop and report.** If a window fails
   any other condition, wait for the next quarter-hour with the shadow timer still running; never
   improvise. When it passes: `systemctl stop dubplate-shadow.timer`; confirm `dubplate-shadow.service`
   is not activating; on the box, `cp -p /usr/local/bin/dubplate-server{,.prev}`, `cp -p
   /etc/dubplate/server.env{,.prev}`, and `cp -p /etc/dubplate/env{,.prev}` (each copy keeps mode 0600
   and root ownership); record the `.prev` binary's `version` stamp, and whether step 6 showed
   `/etc/dubplate/env` changing.
8. `bash scripts/deploy.sh --apply 5.78.148.42` from the scratch worktree, at once after step 7.
9. One manual shadow pass: `systemctl start dubplate-shadow.service`; read its `Result` and the
   journal's summary line; it must be `success`.
10. Box verification: the two leaves exist as `music:music 0700`, with their SELinux context read
    once (`ls -Zd`; `var_lib_t` expected, vps-conventions); the shadow leaf still 0755;
    `systemctl is-enabled dubplate-server-serve.service` answers `disabled`; the new binary's stamp is
    Task 9's commit.
11. The window closes: `systemctl start dubplate-shadow.timer` (being `Persistent=true`, it may fire
    at once, which is harmless: the service is a oneshot and runs serially); read its next run's
    `Result` after it fires.
12. The admin bootstrap: `dubplate-server member add <DUBPLATE_ADMIN_EMAIL> --admin --name
    <Geoff's display name> --streamer-username <Geoff's Navidrome login>`, run as `music` through
    `systemd-run --wait --pipe --uid=music --gid=music -p EnvironmentFile=/etc/dubplate/server.env -p
    ProtectSystem=strict -p ReadWritePaths=/var/lib/dubplate/roster` (`member add` opens only the
    roster, `member.go:120-124`), then `member list` the same way; the record keeps the exit codes and
    the member count only.
13. Doctor against the rendered environment, the same way with `-p
    ReadWritePaths=/var/lib/dubplate/auth` added and without `RestrictAddressFamilies` (R-5d7): exit 0,
    every row read; `boundary probe` reads `no listener`. 6c's settings rows (one per registered
    setting, after the backup rows) each read the default with "store holds no settings yet", since
    the box has no store. A row reading "env `<v>` seeds at serve's next start" is a stop: between 6c
    and 6d `serve` seeds nothing, and the template sets none of the five keys, so such a row means a
    hand-set value.
14. One supervised start: `systemctl start dubplate-server-serve.service`; `doctor` again, with
    `boundary probe` reading `refused` and `bind addresses` reading `127.0.0.1 present`; the journal
    shows the listener and no error; `systemctl stop dubplate-server-serve.service`. The unit stays
    disabled; nothing routes to it (the tunnel still serves Navidrome only).
15. Lift the carry: the conductor edits STATUS's "No deploy until 5d" item to record what deployed,
    that the three `.prev` copies stay on the box as the rollback until the next attended deploy replaces it,
    and what the next deploy may carry (its command is `deploy.sh --apply`).

If a step fails. Before step 7, every step is read-only: stop, and the failure becomes a finding.
From step 7 on, **the rollback is part of the sitting, never an improvised fix**, and runs in this
order: if `dubplate-server-serve.service` is running or looping, `systemctl stop
dubplate-server-serve.service` and `systemctl reset-failed dubplate-server-serve.service` (with
`Restart=on-failure` and `StartLimitBurst=30`, a failing start loops); `cp -p` all three `.prev` files (the binary,
`server.env`, and `/etc/dubplate/env`) back over the live ones; if step 6 showed `/etc/dubplate/env`
changing, rerun the compose step (`deploy.sh:508`'s `docker compose --env-file /etc/dubplate/env ...
up -d`) so the stack reads the restored file; `/usr/local/bin/dubplate-server version` shows the `.prev` stamp recorded at step
7; `rm /etc/systemd/system/dubplate-server-serve.service` and `systemctl daemon-reload`; one manual
`systemctl start dubplate-shadow.service` reading `success`; `systemctl start dubplate-shadow.timer`.
**Whenever the `.prev` binary is restored, beets returns to 2.13.1 with it**, before the manual shadow
pass and while the shadow timer is still stopped: `runuser -u music -- env HOME=/srv/music uv tool
install --force beets==2.13.1`, run with no `dubplate-import.service` activating. The reason: the
`.prev` binary (`be27434`) predates `367564c` and still carries the matcher's workaround for beets
2.13.1's featuring-credit bug, so the matcher and the importer's beets move as a pair, and a shadow scoring that emulator
against 2.14.1 decisions would put featuring-title divergences into launch-gating evidence. If the
apply stopped before its beets step, the command reinstalls the version already there. The
downgrade needs no copy of the library database (`/srv/music/state/library.db`): 2.14.1 leaves it
readable by 2.13.1, since both carry the same nine migrations (`beets/library/migrations.py` is
byte-identical), the same fixed fields (97 item, 43 album), and 2.14.1's one storage change binds
paths as `bytes` instead of `memoryview`, both SQLite BLOBs (2.14.1 changelog, "Other changes"; read
from both PyPI sdists, 2026-09-25). The `auth` and
`roster` leaves stay, since nothing reads them, and the two minted keys stay in the age store. The `config/` copy stays: it is the repo's intended
state, and the change is member-facing wording ("family" to "group"), not behavior. If the first import run after a successful apply fails, the
remedy is this whole rollback, binary and beets together, taken only with Geoff's word; beets never
moves alone. The failure, and the step it stopped at, becomes a finding for a follow-up task.

---

## Task 11: The rung 5 close

Model: `sonnet`.

Files: the close drafts under `docs/superpowers/research/handoffs/`; the errata targets named below.

Outcomes:
- **The 5d-ii HISTORY entry and the rung 5 summary**: what landed per task across both halves, the gate
  evidence, the adversarial cycle's counts and dispositions, the sitting's record (under Ruling 2's
  yes, statuses only), declared coverage gaps (the SMTP transport not exercised on the box, carried to
  rung 16), both budgets scored, and "What a later pass would be wrong to rediscover" (at least: the
  hop rule stops at a malformed hop and keeps a port-bearing hop's address; doctor's web row is serve's
  own resolution; the probe route decides on `PeerAddress` minus the trusted set; the state-dir leg
  reads `DUBPLATE_SERVER_LEAVES` and proves the loop iterates it; `deploy.sh` checks by default; the
  apply swaps the binary before the env, so a deploy that changes the env's key set stops the shadow
  timer first; the unit is installed disabled; the estate's trusted-proxy values wait for the
  cloudflared address pin).
- **The STATUS draft**: present tense, rung 5 closed, the lane registry line, the next action, the
  carries 5d hands on: to the cutover (rung 16): the trusted-proxy estate values and the `compose.yaml`
  address pin for cloudflared, the serve bind to the bridge gateway, enabling the unit, the secrets in
  every `music`-uid verb's environment, and exercising SMTP from the box; to rung 12 (R-5d17): the verify
  endpoint and its forward-auth doctor proof (BS:510-513), and stripping every group 2
  name but the configured header from a trusted peer; to 6d or the architecture cleanup pass: moving
  5d's value checks onto a registry hook.
- **The errata**, each dated and pasted where the spec's "Errata the close applies" section says,
  skipping any already applied (grep first): the build spec's ladder entry 5 (page stack), ladder
  entry 7 (gate rejoin and roster), B12 (proxy hygiene), B4 (tokens and logged URLs), and every other
  paste-ready block in that section. Plus the owed errata the fold record lists, never an edit to a
  ratified document beyond the pasted erratum: the rung 5 spec's 5d row as executed (two passes in lane
  1, not a worktree; the example-roster clause already dropped); the SMTP rows bare, not literal
  (R-5d14); the secrets minted at 5d (Ruling 2's yes) or carried (no); the verify endpoint moved
  (R-5d17), with BS ladder entry 5 (`BS:707-714`) and BS:510-513's forward-auth doctor proof;
  the owned-prefix route list (`spec:915-918`) gaining the `boundary-probe` route; R6's "dials the
  service's own configured bind address over loopback" (`spec:369-370`) becoming "dials each configured
  bind address from that address"; "Deployment"'s "`deploy.sh` gains the unit and the two pre-creates,
  and nothing else" (`spec:1765`) gaining R-5d13's clean-tree refusal and local build and R-5d16's
  check-by-default; R23's reservation moved to about 48.5M; and the pass-split table's 5d exit gate's
  "the unit starting ... failing with a named mount-namespacing error" recorded as Task 5's `/dev/shm`
  workstation evidence plus Task 10's box start.
- **A conductor edit of `CLAUDE.md`, routed by the STATUS draft** (CLAUDE.md is conductor-owned and
  outside the implementer's wall). Once Task 4D and Task 5 land, CLAUDE.md describes the old tool, and
  agents read it ("a rule lives where it executes"). The draft names each line and its new text:
  "Safety rails"' deploy line ("`--check` first for anything risky") becomes "checks by default;
  `--apply` writes, and only from a clean committed tree"; the gate paragraph's state-dir leg ("the
  env template, the shadow unit, and `deploy.sh` name the same state directory") becomes the
  generalized leg over every greenfield unit and `DUBPLATE_SERVER_LEAVES` (R-5d11); and its
  "pattern-filtered `systemd-analyze verify`" gains "hard-required" (R-5d12).
- **The contested-surface rows** released: `internal/serve` and `internal/auth` pass to rung 12's
  ownership per the spec's table (`:2542-2543`).
- A `go-architecture-reader` per Go package Task 9 touched (conductor's dispatch), none if it touched
  none.

Acceptance criteria: every carried item in STATUS's "No deploy until 5d" line is closed or routed
by name; `git diff --stat <5d base>..HEAD -- internal/config` is empty; `bash scripts/check.sh` green.

Deliverable count: 4 (HISTORY, STATUS, errata, contested surfaces).

---

## The exit gate

`bash scripts/check.sh` green on master, plus each assertion and its proving task. The spec's 5d row
(`:2519`) is the source; each clause maps here.

1. `systemd-analyze verify` clean on the new unit, and the tool hard-required (Task 5).
2. The generalized state-dir check failing when either leaf is removed from `deploy.sh`'s array or its
   pre-create loop (Task 5).
3. The unit starting with the roster leaf writable and failing with systemd's named mount-namespacing
   error when its pre-create is removed (Task 5's `/dev/shm` evidence; Task 10's box start under
   Ruling 2's yes).
4. `doctor` reporting the live client-IP mode, the bound addresses, and a boundary-probe row that
   dials the running listener and reports plainly when none answers (Tasks 2 and 3).
5. `deploy.sh --check` rendering with no unresolved name and `doctor` exiting 0 against the rendered
   environment (Task 6's gate-resident proof; Task 10's steps 3, 6, and 13 under Ruling 2's yes).
6. The disposition table covering all 50 findings (Task 8).
7. The adversarial cycle's findings folded or refused with reasons (Task 9).
8. `internal/config` untouched (both closes' check).

**Deferred, by name:** the verify endpoint and its forward-auth doctor proof (R-5d17); the
estate's trusted-proxy values and the cloudflared address pin (rung 16); minor 14 (6d criterion 8); L5
and L6 (unassigned); **under Ruling 2's no, clause 5's `doctor`-exits-0 half**, carried to rung 16 with
the live shadow's dependency on it.

---

## Contested surfaces

| Surface | Owner | This pass's position | Contention |
| --- | --- | --- | --- |
| Lane 1 (the main tree) | the lane registry | both halves | 5c-iv must close first; the conductor writes the order |
| `cmd/dubplate-server/doctor.go`, `doctor_test.go`, `check.go`, `check_test.go`, `main.go` | shared | Tasks 1, 2, 3 (and Task 6's fixtures) | **6c merges to master before 5d-i's executor launches** (conductor decision, revision 2.1): 6c's Task 5 edits `doctor.go`, `check.go`, and their tests, so the two never run concurrently on these files, and Task 0 item 5 re-reads them as 6c left them. Lane 2's 6b-ii-b and translit (doctor rows, a verb, registration lines, exit-0 doctor fixtures) may still merge during 5d; a non-fast-forward merge triggers the merge ritual's re-review. Task 0 reports what merged |
| `scripts/deploy.sh` | shared | Tasks 4D, 5 | translit's staged transition (lane 2), whose apply command becomes `--apply` once Task 4D lands (the conductor writes this on the registry row for lane 2); 6d later edits the leaf array |
| `docs/bringup-runbook.md` | rung 4's file | Task 4D (the `--apply` invocations only) | none; retired at rung 16 |
| `scripts/check.sh` | shared | Task 5 | every lane; the state-dir leg and the hard-required list are 5d's |
| `env/dubplate-server.env.template` | rung 4's file; rung 5 adds rows | Task 6 | 6d removes the roster row |
| `systemd/` | shared | Task 5 (one new file) | 6d edits this unit's grant |
| `internal/serve` | rung 5 until this close, then rung 12 | Tasks 1, 3 | none open (PA:80) |
| `internal/config` | rung 1's contract | untouched | 6c's registry edits merge before 5d-i launches (revision 2.1); 5d leaves it byte-identical |
| The box | production | Task 10 only | the quiet-box check and the timer window; no other lane deploys during 5d |
| `docs/` ledgers | conductor | the 5d-i close and Task 11 drafts only | every lane |

---

## The conductor's actions

Not tasks; no implementer performs them.

1. Before Task 0: confirm 6c has merged to master (revision 2.1); record planning spend against the rung; write the lane registry's lane 1 row and the
   contested-surface rows (including the `deploy.sh` `--apply` note for lane 2); arm the runaway guard;
   AC power; confirm the heavy gate string.
2. Before Task 1: rule Task 0's blockers.
3. At each segment boundary: STATUS (task ledger, decisions, spend, next task); the 80% flag check.
4. At the 5d-i close: the simplifier round, the two architecture reads, the ledgers; close the session.
5. 5d-ii opens as a fresh session from the resume prompt.
6. After Task 7: the `prose-voice-reviewer` read and its fold inside Task 7.
7. Between Tasks 8 and 9: the adversarial cycle as written above.
8. Task 10 under Ruling 2's yes, attended.
9. At Task 11: apply the `CLAUDE.md` edit its STATUS draft routes; any architecture read Task 9's
   packages need; score both budgets; close the session.

---

## Execution model

Workflow mode: the pass-execute chain per task, **sequential**, one invocation per segment. Every
implementer task runs `dubplate-implementer`, then `diff-reviewer` (`claude-opus-5-5`, effort `high`),
then the gate, with one re-dispatch on `fix` and a second `fix` as the conductor's decision.

| Pass | Segment | Tasks | The boundary's own work |
| --- | --- | --- | --- |
| 5d-i | before A | 0 | the conductor rules Task 0's blockers. |
| 5d-i | **A** | 1, 2, 3 | STATUS. |
| 5d-i | **B** | 4D, 5, 6 | STATUS; the 5d-i close. |
| 5d-ii | **C** | 7, 8 (re-grepped at the 5d-ii base if drafted earlier), the adversarial cycle, 9 | STATUS; the two lenses; the fold. |
| 5d-ii | **Sitting** | 10 (Ruling 2's yes) | attended |
| 5d-ii | **Close** | 11 | the ledgers |

Each boundary sits on a task's own green committed gate.

Models. `sonnet` at effort `high` by default. Upshifts to `opus`, unconditional: Task 0 (fresh
context), Task 1, Task 3, Task 8, Task 9, and the universality lens.

Dependencies, stated exactly: 6c merged to master before Task 0 (conductor decision, revision
2.1). Task 0 before any task. Task 1 → Task 2 → Task 3. Task 4D → Task 5;
Task 1 → Task 5. Tasks 2, 3, 5 → Task 6. Tasks 1 to 6 → Task 7. Task 8 depends on nothing but must
precede the cycle. The lenses → Task 9 → Task 10 → Task 11. **Tasks 4D and 2 are independent by files**
but run sequentially because one tree runs one executor.

---

## Prior art consulted

Per CLAUDE.md, from the streamer spec's list (`:770-806`) and the rung 5 prior-art study
(`docs/superpowers/research/2026-09-08-rung-5-auth-prior-art.md`).

| Entry | Verdict for 5d, with reason |
| --- | --- |
| stdlib `net/netip` | **Adopt** for the proxy set, the hop parse (`ParseAddr`, then `ParseAddrPort`), and the probe's peer rule: value-type addresses with `Unmap` give the canonical comparison `net.IP` does not (R-5d2, R-5d4, R-5d8, SEC finding 22). |
| stdlib `net.Dialer` with `LocalAddr`, `net.InterfaceAddrs` | **Adopt** for the probe's known source and the bind row (R-5d7, R-5d8), with `InterfaceAddrs`'s netlink dependency named. |
| `golang.org/x/net/http/httpguts` | **Decline**: a new module dependency against R26's count of two, for a token check of a few lines. |
| `golang.org/x/time/rate` | Already adopted at 5b (R26); 5d changes only the key it receives. |
| Authelia (Apache-2.0) forwarded-headers documentation | **Adopt as behavior and citation** for the runbook's trust-boundary section (spec `:894-899`). No code read. |
| Miniflux `AUTH_PROXY_HEADER` with `TRUSTED_REVERSE_PROXY_NETWORKS` | **Adopt as shape corroboration** for R22's class; **decline** its network (CIDR) form, since the spec pins trust to addresses, never a subnet. |
| `coreos/go-systemd` (sd_notify) | **Decline**: `Type=exec` reports an exec failure without a readiness protocol, and adding a dependency for readiness no consumer needs is ahead of a caller. |
| Navidrome's reverse-proxy authentication (GPL-3.0) | **Not consulted**: the verify endpoint moved to rung 12 (R-5d17), which consults it for behavior only; never read for code. |
| Caddy `forward_auth`, Traefik `forwardAuth`, nginx `auth_request` documentation | **Not consulted**: rung 12 quotes each for the method it sends when it builds the verify endpoint (R-5d17). |
| OpenSubsonic, taglib, wazero, modernc sqlite, goose, mewkiz/flac, Picard, chromaprint | **Neutral**: none touches 5d's surface. |

---

## Not verified in this revision

- Whether translit's `unidecode/` leaf needs a pre-create in `DUBPLATE_SERVER_LEAVES` or is written by
  its install verb as root; Task 0 item 4 reads translit's landed `deploy.sh`.
- The exact text `systemd-run --user` prints on the box's systemd 257 for the namespace failure (the
  workstation's 259 was measured); the box repeats only the positive half, so the exact wording on 257
  is never asserted.
- The box's `install -d` (AlmaLinux coreutils) against the workstation's measured 9.10 behavior; Task 10
  step 10 reads the leaves' modes after the apply.
- Whether `validateSMTPFrom` accepts the age store's `DUBPLATE_SMTP_FROM` value; Task 10 step 3 proves
  it before the deploy, and a refusal there is a finding, not a deploy.
- Which rows `Load` and `doctor` stat or open for Task 6's overrides; Task 6 quotes them.
- The heavy gate's wall time after 5c-iv, which sets each task's `-timeout`.
