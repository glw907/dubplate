# Dubplate alignment plan: fold verification

**Target:** `docs/plans/2026-10-11-dubplate-alignment.md` and
`docs/superpowers/plans/2026-10-11-dubplate-alignment-prompt.txt` at `582d47e`, against the fold record
(`research/2026-10-11-dubplate-alignment-plan-fold.md`), the three reviews, the approved design, `pass-core`, the
`dubplate-pass` adapter, and `spec-plan-review` ("The verification read"). This reader had no part in the fold.
Scratch probes are in `~/.claude/jobs/d7d4f1c0/tmp/verify/` (`ci-probe.sh`, `filt.sh`, `p.diff`, `arrival.sh`,
`cites.sh`, `clockhist.sh`). No gate ran beyond `--dry-run` scope prints.

## Closure of the blockers and majors

Every blocker and major closed at the location the fold record cites:

- **B1** (the unmatched wrapper widens every fast lane): Conventions "The base" (plan:126-136) and T0's acceptance
  (plan:270-272) now expect the fallback. Probe from this isolated worktree:
  `bash scripts/check.sh --scope-changed --base c58caf2b50f05d52c0fcaf6aff31a47aed05f9d0 --dry-run` printed
  `check.sh: scope=changed base=c58caf2b... source=caller rendered=skipped ... files=6`, so a literal base narrows,
  and the guard accepts a plain `bash <script> --base <40-hex>` call. Every cited `check.sh` line holds (`:75-76`,
  `:109-110`, `:127-130`, `:268-294`, `:307`, `:367-368`, `:395-397`, `:399`, `:458-459`, `:537`, `:781`), and
  `cairn-run-gate:335` is `exec bash -c "$2"`.
- **M1** (`$*` into `bash -c`): T0 allowlist, plan:236-239. **M2** (launcher): plan:97-112. **M3** (estimates):
  plan:41-44 (the sums check: 450 task-minutes, a 335-minute path). **M4** (docs-only halves): plan:137-141. An
  all-`noeffect` change set prints `scope=changed ... rendered=skipped` (a dry run on today's branch, whose 9 changed
  files are all under `docs/`), so the pair 1 post-merge `scope=changed` expectation holds.
- **M5, C10, R6** (pair writes outside its worktree; records in scratch): Records, plan:196-203. **M6, C1, R2** (the
  TypeScript 7 item): row I66 (plan:452). **M7**: settled in T0 (plan:253-260).
- **M8, C4, C11** (wrapped phrases, "lanes", case-blind rung grep): plan:144-150, :188. The joined check on today's
  tree finds "lane registry" in CLAUDE.md, the adapter, and STATUS, and "post-mortem" twice in ROADMAP; the lane
  check prints "Document lanes", "photo lane", "dubplate lanes", and others.
- **M9, C5, R9** (diffs with no base): plan:151-152; T5 at :628-629; T6 at :689-691. **M10, R10** (filter holes):
  plan:155-175. Rerunning the plan's filter text over the fold's `probe.diff` flags exactly 14 lines; the heredoc grep
  over `provision.sh` prints nothing (4 `<<<` here-strings). **M11**: plan:735-740.
- **C2, R3** (play-queue requirement): removal map and `up next` (plan:320-326, :337, row I70). **C3, C15, R4**
  (HISTORY chain; STATUS checked by its writer): the 70-row inventory (plan:371-462). Probe: the check prints 56
  `MISSING` lines (55 `B`, one `S`, which is I4) and the row count prints 70, matching the fold record. **C6**:
  plan:506-508. **C7, m7**: T1. **R1**: T0 `.gitignore` (plan:248-252) plus T4's "extract, then delete" class
  (plan:573-574). `git check-ignore -v` behaves as stated: in a scratch repo where `info/exclude` and a worktree
  `.gitignore` both match, it printed `.gitignore:2:docs/secret.md` (`ci-probe.sh`). **R5**: plan:350-365. **R7**:
  plan:221-227. **R8**: plan:554-562 and :355-359.

The audited line citations hold (`cites.sh`): `serve.go:64-67` and `:140-148`, `doctor_test.go:22-33` and `:289-292`,
`write.go:64`, `provision.sh:123/143/161`, `main_test.go:24`, VI `:226/277/314`, `roster-walk.md:35`,
`handlers.go:218`, `auth.ts:26`, build spec `:1143`, parallel-build `:3`, the 14a fold `:91`, the old plan
`:829-833`, `reverse-edge-completion.md:37`, the ROADMAP ranges, adapter `:58-69`, CLAUDE.md `:70` and `:89-92`,
`serve.go:300/391`, and the rung 7 spec's `:8`, `:165`, `:877`, `:1233-1263`. T6's settings check prints
`17 [three rules] True` today, and the eight-label grep prints 8. The auto-memory directory holds 41 files plus
`MEMORY.md`, and T4's provisional classes name all 41. `find -newermt "@<now-60>"` printed 0 lines, and `@1000`
printed 43.

## Findings, ranked by consequence

### 1. Major: T1 hands an unattended Sonnet session a skill whose required outputs conflict with the plan's boundaries

- **Location:** plan:284-301 (T1); prompt.txt:22-24.
- **Defect:** T1 runs "the `dependency-upgrade` skill" with no reconciliation against this pass. The skill
  requires, per sweep: a refactor decision, where "Take now: ... lands in this upgrade's own task" (code edits,
  possibly in `auth-data` paths); "File: A ROADMAP line"; "Held majors ... and a scheduled tripwire (the `schedule`
  skill) for each machine-detectable condition"; "CI on a pushed branch"; and a CHANGELOG entry. Two majors are held
  today with machine-detectable conditions (`2026-10-10-dependency-sweep.md:120-121`: typescript 7.0.2 and
  @types/node 26). A scheduled cloud routine is a write outside the worktree, which the plan's Out of scope and the
  prompt forbid ("never write any repo other than..."), and it is a likely auto-mode refusal, which is a stop. A
  take-now refactor widens a process PR into behavior. T1's Files list names none of ROADMAP, BACKLOG, or code, so the
  executor meets a conflict with no rule. Also, T1 cites `2026-10-04-dependency-sweep.md` as the record to sit
  beside, but the latest is `2026-10-10-dependency-sweep.md`, and the skill says to extend an existing record for
  the same window in place.
- **Fold:** add to T1's outcome: "In this pass every take-now item is filed, not taken: a BACKLOG item tagged
  `#rung:<name>` (T1 runs before T3's pair, so BACKLOG is free). Held majors are recorded with their unblock
  conditions; no tripwire is scheduled (the PR body lists them for the conducting session). The full-suite home is
  the local `bash scripts/gate.sh` (no CI); dubplate has no CHANGELOG. The prior record is
  `2026-10-10-dependency-sweep.md`." Add `BACKLOG.md` to T1's Files. The prompt's T1 sentence gains "file every
  take-now item; schedule nothing".

### 2. Minor: the comment-only filter still admits a code line wrapped in two one-line block comments

- **Location:** plan:162 (Go arm), plan:164 (CSS arm), and the claim at plan:169-175.
- **Defect:** the one-line `/* */` arms test `^[ \t]*\/\*.*\*\/[ \t]*$`, and `.*` spans a closing `*/`. Probe
  (`filt.sh` over `p.diff`): `/* x */ y := 2 /* z */` and `/* x */ return nil /* y */` in a `.go` file, and
  `/* a */ --color-primary: red; /* b */` in a theme `.css` file, all pass. The last is a token value change in
  exactly the files T5 edits. Separately, a TSDoc block (`/** doc */`, ` * line`) in `vite.config.ts` is flagged,
  so T7 must use `//` there, which T7's text does not say.
- **Fold:** replace `.*` with `([^*]|\*+[^*\/])*\*+` in both arms, so the line holds exactly one comment, and add "a
  `//` comment" to T7's `vite.config.ts` bullet (plan:718-719). The existing options already use `//` comments.

### 3. Minor: two check commands wrap inside a code span at a hyphen

- **Location:** plan:592-593 (`'from memory\|parallel-` / `build-approach\|at T14'`); plan:683-684
  (`(Gate|Fast lane|Full-` / `suite home|...`); also plan:36, whose regex holds a backtick that ends the code span.
- **Defect:** a reader that joins the wrap with a space gets `parallel- build-approach`, which never matches the
  CLAUDE.md:62 path, so that alternative passes vacuously; and `Full- suite home`, which makes the T6 count 7, not 8,
  a false red on an unattended pair. Lines that wrap at a space are harmless.
- **Fold:** keep each of these two commands on one line, or put them in fenced blocks, as the inventory check is.

### 4. Minor: one T7 arrival token is already in its destination

- **Location:** plan:703 and :741.
- **Defect:** `arrival.sh` on today's tree: `X-ND-Authorization` already appears twice in `stack/stack.go` (a code
  comment at `:541` and the header set at `:549`), so the Navidrome lesson's arrival check passes before T7 writes
  anything. (`hook numbers are stable` is present once today, but only inside the section T7 deletes, so its check
  is sound.) Every other token prints 0 today.
- **Fold:** use a token from the lesson's own claim, for example `by name` or `Music Library`, after first
  confirming `grep -c` prints 0 for it at the base, or have T7 drop the lesson as carried if `stack.go` already says it.

### 5. Minor: the Resume paragraph omits three mechanics

- **Location:** plan:221-227; prompt.txt:3-10.
- **Defect:** (a) The clock file survives a crash with its old `start=`, so the resumed session's first tool call
  trips `claude-clock-stop` (elapsed past twice the estimate). pass-core says "A resume rewrites `start=`", and the
  paragraph does not. (b) "`git log --format=%s master..HEAD`, subjects starting `T<n>:`" misreads after Close
  step 1 merges `origin/master`, since every dubplate pass uses `T<n>:` subjects and the local `master` is stale. (c)
  "A pair worktree with uncommitted work is a stop", but the session cannot read another worktree's status (the
  guard refuses `git -C`), so the test as written cannot run. Also, the prompt's first sentence ("Before any other
  act, call EnterWorktree with the name") precedes its resume check.
- **Fold:** "rewrite `start=` to now"; read task commits with `git log --first-parent --format=%s <launch
  base>..HEAD`; "a pair branch with no `T<n>:` commit is a stop (its worktree may be warm)"; and open the prompt
  with the existence check, then EnterWorktree by `path` or by `name`.

### 6. Minor: T4's `find -newermt` check names no report for a non-empty result, and other sessions share the directory

- **Location:** plan:597-599.
- **Defect:** the auto-memory directory is shared by every dubplate session, worktree sessions included (this
  session's context names the same `memory/` path). A conducting or planning session that writes a memory during the
  pair turns this "prints nothing" into a red with no stated handling, on an unattended run. The post-merge step
  already keeps any newer file (plan:356-357).
- **Fold:** "Each printed path is listed in the PR notes as written during the pass and kept by the post-merge
  re-list; it is not a stop unless the path is one T4 classed delete."

### 7. Minor: T3's `#rung:` floor can contradict "related rows may share one item"

- **Location:** plan:505 against plan:377-378.
- **Defect:** about 46 inventory rows carry a rung tag, but the slice 3, 6, and 8 and `cutover` rows merge naturally
  into a few items each, so a faithful merge can land under 30 tagged lines and stop the pair.
- **Fold:** drop the fixed 30, since the inventory check is the real proof, or state "at least 30 unless the map
  shows merged rows".

### 8. Minor: the lane check flags a hyphenated lane name

- **Location:** plan:148-150.
- **Defect:** `\b([[:alnum:]-]+ )?lanes?\b` matches the bare `lane` in "heavy-lane", which the exclusion then misses;
  the plan itself writes "heavy-lane lock wait" (plan:44), so a T6 or T4 writer copying that phrase gets a false red.
- **Fold:** exclude `(heavy|light|fast)[ -]lanes?`, with the leading group widened to `([[:alnum:]]+[ -])?`.

### 9. Minor: T7's `diff-reviewer` acceptance reaches the pair verbatim

- **Location:** plan:747; prompt.txt:50-51.
- **Defect:** the pair receives its task text verbatim, including "`diff-reviewer` ... returns `accept`". The prompt
  gives that read to the session after the merge. A `general-purpose` pair may dispatch it itself, so it runs twice,
  or it treats the acceptance as unmet.
- **Fold:** append "(run by the session after the merge)" to that bullet.

### 10. Minor: the fold record claims probes for two mechanisms it does not show

- **Location:** fold.md:144-146.
- **Defect:** the record lists "the git-dir records" and "the resume procedure" as "each probed or quoted above",
  but its probe section has no entry for either. Precedent does exist: two isolated cairn sessions wrote
  `pass-task-clock` into their worktree git dir with one plain `printf ... > <cairn-cms>/.git/worktrees/<name>/pass-task-clock`
  (`lean-cutover`; `engine-pass-b`, whose file also carried a `base=` line). This read also confirmed that a
  read pipeline over a git-dir file runs under the guard.
- **Fold:** cite that precedent in the record. T0's clock write is the first live proof, and a refusal there is a stop
  before any work.

### 11. Minor: small points

- plan:278-279: "the originals' `stat -c %Y` is unchanged" has no baseline. Record the three values before the copy.
  Today they are the Oct 10 mtimes shown by `ls -la reports/`.
- plan:481 and :769: the planning prompt tells the `admin-api-1` planner "this plan is the model", but the Close
  archives the plan to `docs/archive/plans/`. The prompt should cite the archive path.
- plan:109: the launch check matches on `cwd` only. Add `or .name == "dubplate-alignment"`, so that a session still
  in the main checkout, or one that died before EnterWorktree, is also caught.
- plan:253: the fold dropped the pre-fold provenance "relayed by the conducting session" from Geoff's Pi 5
  direction. The design's Out of scope gives `reports/` to the conducting session to settle with Geoff, so restore
  the source.
- Row I70 (plan:456) greps `up next` case-sensitively, while T2's acceptance (plan:337) is case-insensitive. A
  rewrite that starts "Up next" passes T2 and fails the post-merge inventory check. Use the same case rule in both.

## New mechanisms

| Mechanism | Evidence |
|---|---|
| Wrapper allowlist, no trailing space | `cairn-run-gate:335` quoted; T0's refusal acceptance proves it at run time |
| `--base` rule | `check.sh` lines quoted; probed here (`source=caller`, `scope=changed`) |
| Docs-only pair rule | `check.sh:395-397` quoted; probed here (an all-`noeffect` set prints `scope=changed rendered=skipped`) |
| Comment-only `awk` filter | Fold probe reproduced (14 of 14). One residual hole, finding 2 |
| Line-joined phrase and lane checks | Probed here on today's tree. Residual: hyphen wrap (none today) and finding 8 |
| STATUS inventory check | Probed here: 70 rows, 56 `MISSING` |
| `find -newermt` | Probed here (0 and 43 lines) |
| `check-ignore -v` | Probed here in a scratch repo with a worktree |
| PR notes in the git dir | Not probed by the fold; precedent cited in finding 10 |
| Resume paragraph | Not probed; finding 5 |
| Dependency-sweep task | Quoted from the adapter and standing rules; the skill's outputs are not reconciled (finding 1) |

## Launch prompt against the plan

The two agree on task ids (T0 and T1 inline, then T2+T3, T4+T6, T5+T7), the models (session `sonnet` `medium`,
pairs `sonnet` `medium`, T3 `opus`), the six dispatch rules, the per-pair skills, the gate rule (`gate.sh`,
`--base <base>`, `timeout: 600000`, re-issue on 75, never bare `check.sh`), the PR-notes location, the stops
(including the classifier refusal), the post-merge boundary, and the close. The only gaps are findings 5 (resume
order) and 9 (who runs T7's `diff-reviewer`).

## Design and leanness

No fold addition changes the approved design. The class raise follows pass-core's floor. The `.gitignore` entry
records an existing Geoff ruling inside the repo and does not touch the main checkout. The sweep applies the
adapter's standing rule (`SKILL.md:45-49`, `standing-rules.md:32-34`), so the fold's "not a fork" reading holds:
running it is the default, and only a waiver would be Geoff's call. Finding 1 bounds what the sweep may do. The plan
holds no implementation code beyond check commands. Its growth from 510 to 786 lines is mostly the 70-row inventory,
which carries the pass's real payload. The added ceremony (docs-only skips, literal bases, the git-dir records) saves
gate time or removes a data-loss path, so none costs more than the risk it removes. **OWNER FORK:** none.

## Verdict

No blocker remains. Every blocker and major from the three lenses closed at its cited location. The new gate
mechanisms (the `--base` rule, the docs-only skip, the joined checks, the inventory check, `find -newermt`, and
`check-ignore -v`) are quoted or were reproduced here against today's tree, and the launch prompt matches the plan.
One major remains for the fold agent: T1 runs the `dependency-upgrade` skill unbounded, and in an unattended run
that skill would schedule tripwires outside the worktree or take refactors into a process PR. The remaining ten
findings are minor. They cover one hole in the comment filter, two commands that wrap at a hyphen, one arrival
token that is already present, and gaps in the resume mechanics and in the reports of three checks. Each folds in a
line or two, and none needs another verification read.
