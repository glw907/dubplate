# Dubplate alignment plan review: data integrity and failure risk

Reviewer lens: data integrity and failure risk (what a crash, a concurrent process, a partial write, an interrupted
session, or a merge can lose or corrupt). For this pass, "data" is mostly knowledge: rulings, open items, owed work,
lessons, and memory. Target: `docs/superpowers/plans/2026-10-11-dubplate-alignment.md` at `61e6af5` (cited as
"plan"). Also read in full: the design, the audit, the launch prompt
(`docs/superpowers/plans/2026-10-11-dubplate-alignment-prompt.txt`), `docs/STATUS.md`, `ROADMAP.md`, `BACKLOG.md`,
`docs/standing-rules.md`, `.claude/settings.json`, the adapter, every file under `.claude/agent-memory/`, all 42
files of the project auto-memory (read only), the HISTORY sections STATUS points into (`HISTORY.md:291, 316, 443,
511, 588, 697-701, 802`), the rung 7 spec's head and "Owed records", and `pass-core`.

Two facts observed on disk change the plan's premises:

- The main checkout's `.claude/agent-memory/dubplate-implementer/` was settled at 10:40 today: `MEMORY.md` is
  byte-identical to `master` (`cmp` against `git show master:...` passes) and the four untracked siblings are gone.
  The plan still assumes the modified file (plan:22-24, :95, :498-500).
- `docs/superpowers/research/2026-09-04-visual-collaborator-outreach.md` sits untracked in the main checkout and is
  not ignored. The only record that it must stay uncommitted is a memory file the plan deletes (finding 1).

## Findings, ranked by consequence

### 1. Major: a privacy ruling lives only in a memory file classed for deletion

- **Location:** plan:325-329 (session-log class) and plan:498-501 (post-merge (a) and (b)).
- **Defect:** `unattended-window-2026-09-07.md` is classed "delete, session or window log", but it is the only home of
  Geoff's ruling "the visual-collaborator outreach doc stays uncommitted (personal emails)". That doc is one of the
  two untracked 2026-09-04 research notes, and post-merge step (a) tells the conducting session to "settle" exactly
  those untracked files, after step (b) has deleted the ruling. A conductor that settles by committing puts personal
  email addresses into git history. No acceptance check reads session logs for live rulings: the dry-run table's
  only test is row count (plan:349).
- **Fold:** T4's classing rule gains a fourth class, "extract then delete": a session log whose body holds a ruling
  still in force. For this file, add the path to `.gitignore` in T4 (a rule lives where it executes; the filename is
  not sensitive) and name it in post-merge step (a). T4's report quotes every ruling it finds in a log file with
  its new home, or "moot" with the reason. The same read surfaces the photo log's "five maybes wait in
  `photos/.candidates/`" (STATUS keeps only the one Run-D.M.C. frame) and STATUS:8's "emulator needs `-gpu host`",
  which contradicts `docs/roster-walk.md:35` (`--disable-gpu`); route both to the `#rung:streamer` roster-walk item.

### 2. Major: the TypeScript 7 recheck can vanish between two concurrent pair halves

- **Location:** plan:209-210 ("T3 owns BACKLOG, so T2 hands the text to T3 through the session: T2's report quotes
  it") and plan:247 ("and T2's TypeScript 7 item").
- **Defect:** T2 and T3 run at the same time (plan:222, :281). T3 cannot receive text from T2's report, which exists
  only after T2 returns. T2 deletes ROADMAP:111-113; if T3 does not invent the item on its own, Geoff's 2026-09-26
  ruling ("Before the 0.8 release: recheck TypeScript 7") is in neither file. Neither task's acceptance greps for it.
- **Fold:** T3 copies the item from ROADMAP:111-113 in its own base (the pair branches before T2 lands), and T3's
  acceptance adds `grep -n 'TypeScript 7' BACKLOG.md` prints one line tagged `#rung:integration`.

### 3. Major: T2's trim drops a Geoff product requirement that has no other home

- **Location:** plan:204-207 ("the streamer's settled list (`:78-110`) collapse[s] into pointers to the organizer and
  streamer specs") and plan:330 (`own-visual-identity` classed duplicate).
- **Defect:** ROADMAP:103-109 holds "The web player needs a strong play-queue interface (Geoff, 2026-09-24): now
  playing, up next, reorder, and add from album, track, or filter results. It is designed from a prior-art study...
  in the visual identity spec's language." The streamer spec has the saved play queue (`:689`, `:702`) and puts the
  web player on the 2.1 list (`:1094-1095`), but no "up next" or reorder requirement; `grep -rn -i 'up next'` over
  `docs/` and `IDEAS.md` finds only an unrelated research line. `own-visual-identity.md` says "the queue UI is a 2.1
  web-player requirement on ROADMAP", so both copies go in the same pass. Other ROADMAP-only sentences at risk in
  the same trim: "Server before TUI is a hard boundary" with the organizer spec's rung 10 and the remote-TUI half of
  rung 11 moved past launch (the build spec carries these at `:358` and `:1082`, so they survive), and the 2026-09-12
  sync addendum at ROADMAP:57-59.
- **Fold:** T2's report carries a removal map like T3's: every sentence it deletes maps to a destination confirmed by
  `grep` (file and line) or is kept. The web-player queue requirement moves to `IDEAS.md` or stays as one line in the
  Planned section. T2's acceptance adds `grep -rn -i 'up next' ROADMAP.md IDEAS.md` prints one line.

### 4. Major: the STATUS rewrite is checked only by the writer's own count

- **Location:** plan:239-249 and the acceptance at plan:275-280.
- **Defect:** STATUS:42-60 packs about 37 distinct items into terse handles (R-6d20, R-6d4, UNI-H2, L-20, L-21,
  L-51, L-17, `CAP_NET_BIND_SERVICE`, `duplicate_principal`, criterion 14's `runuser` clauses, "group 2 in
  trailers"), and STATUS:60 resolves through a pointer chain: HISTORY:588 points onward to "HISTORY 5c-iv 'To rung
  7'", "9b's five nits (HISTORY 5c-iv)", "HISTORY, re-plan close, 'Moved out of STATUS'" (HISTORY:697-701, which
  points again to "HISTORY transliteration, 'Routed'"), and `research/*-architecture-reads.md`. T3 names only the
  first hop. The acceptance cannot catch a dropped item: `grep -c '#rung:' BACKLOG.md` must exceed a number "the
  subagent states", and tagging existing items raises the count with no new item written. `grep -n 'HISTORY'
  docs/STATUS.md` passes by deleting pointers, not by resolving them, and nothing greps BACKLOG for a leftover HISTORY
  pointer. The deleted memory `ledger-fold-must-keep-list.md` records this failure: "two Sonnet folds in one night...
  compressed STATUS to fit the cap and dropped Geoff's open questions". T3 runs as a Sonnet pair at medium effort.
- **Fold:** the session, before dispatch, writes the item inventory (one line per item, with its distinctive token)
  into T3's dispatch; T3's acceptance greps each token in BACKLOG or the new STATUS and prints any miss (an empty
  inventory fails the check). T3 follows every HISTORY pointer to its leaf, and `grep -n 'HISTORY' BACKLOG.md`
  shows no new line. Keep each streamer item's slice number in its text ("streamer slice 6"), since `#rung:streamer`
  alone loses which slice consumes it. STATUS:58's two `docker compose` entries are disposed as "done by T6", not
  filed. Upshift T3 to `opus` (pass-core allows a per-task upshift).

### 5. Major: post-merge duties have no owner who will run them, and the main checkout pull is missing

- **Location:** plan:498-503 and plan:231-234.
- **Defect:** steps (a) to (e) belong to "the conducting session", which is unnamed and may be closed when Geoff
  merges. They live only in the PR body. STATUS's next action names only the admin-api-1 launch, so a fresh session
  following STATUS skips the auto-memory pruning, and the retired-process memories (`greenfield-arc-sessions`: "Workflow
  mode is a STANDING grant") load into the admin-api-1 planning session. No step pulls the main checkout, yet (e)'s
  launch line runs `cat docs/superpowers/plans/2026-10-11-admin-api-1-planning-prompt.txt` there, and
  `worktree.baseRef: "head"` would branch the planning session from the pre-alignment `master`, which is pass-core's
  recorded failure ("six repos still ran the pre-cutover process a day after their lean PRs merged").
- **Fold:** STATUS's next action is the ordered post-merge list: `git -C ~/Projects/dubplate pull --ff-only` under
  pass-core's conditions; back up and prune the auto-memory (finding 8); trash the Pi 5 originals; then the
  admin-api-1 launch. Drop step (a) (already settled) and the stale clauses at plan:22-24.

### 6. Major: the PR notes, the only record of where knowledge went, live in ephemeral scratch

- **Location:** plan:135-137, plan:338-340, plan:31.
- **Defect:** T3's dispositions, T4's dry run and proposed index, T5's triage, and T7's lessons table accumulate in
  `$CLAUDE_JOB_DIR/tmp/` (here `~/.claude/jobs/<session id>/`). A crash or API drop that ends the session, followed
  by a fresh session, gets a new job directory and loses them. Those records are the PR body, and the PR body is
  the history under lean ruling 8. T4 also writes `MEMORY.md.proposed` there itself, but the launch prompt forbids
  pairs to write in `~/.claude/` ("none in `~/.claude/`"), so T4 either breaks the prompt or never writes it. End-to-end
  check 2 over a missing file prints only to stderr and "prints nothing" on stdout: a vacuous pass.
- **Fold:** keep the notes at `$(git rev-parse --absolute-git-dir)/pass-notes.md` in the pass worktree's git dir
  (it survives a session restart, is never committed, and matches the clock file's home). T4 returns the proposed
  index in its report and the session writes it. Check 2 over the index begins with `test -s <file>`.

### 7. Major: no resume procedure for a crash or API drop mid-pair

- **Location:** plan:147-151 (Stops) and the launch prompt.
- **Defect:** the Stops convention covers a red, a conflict, a clock stop, and a pair that returns uncommitted. It
  does not cover a session that dies (API drop past the `/loop` wake-up, a crash) or a pair killed mid-turn. A fresh
  session has no rule for unmerged pair branches, a pair worktree with warm edits (global CLAUDE.md: "Warm
  uncommitted code at dispatch time is a stop-and-investigate signal"), or the lost notes. Rerunning T3 or T7 after a
  partial merge appends BACKLOG items or doc comments twice. The deleted `pass-execute-dispatch-mechanics.md` holds
  the recovery rule ("Check the tree first... warm edits from a standalone agent mean resume that agent with
  SendMessage").
- **Fold:** one Resume paragraph. Done tasks are the `T<n>:` subjects in `git log master..HEAD`. For each pair,
  `git worktree list` and `git branch --no-merged`: a committed pair branch merges, and a pair worktree with warm
  edits is resumed by `SendMessage` in the same session, or else removed and the task rerun from the current head.
  The notes file (finding 6) is checked for each merged task's record and rebuilt from the commit if missing. A stop
  commits and pushes STATUS, so Geoff can read the stop off the branch.

### 8. Major: the "duplicate" class tests a clause's presence, not the coverage of the memory's why

- **Location:** plan:293-298, :318-319 ("naming the repo file and line that now carries it, confirmed by `grep`"),
  :330-335.
- **Defect:** a `grep` hit for one clause certifies the deletion of a whole memory. `idiomatic-svelte-daisyui.md`
  holds the exemplar Geoff named (`~/Projects/cairn-cms/src/lib/admin-toolkit/` and `src/lib/components/`: "imitating
  the exemplar beats interpreting the rule") and "un-idiomatic code adjacent to a change is in scope". The planned
  clause ("the stack's idiom wins over a finding's literal fix text") comes from the index line, not the file body.
  `idiomatic-beautiful-go.md` holds "diff-reviewers may return fix verdicts on awkward-but-working code" and
  "dispatch language... should name 'idiomatic, architecturally beautiful Go' as an acceptance dimension"; that rule
  executes in the adapter's pair and reviewer dispatches, which T6 rewrites without it.
  `ledger-fold-must-keep-list.md` (retired) holds a lesson for every future STATUS rewrite.
  `session-2026-09-12-infra-round.md` holds the `pgrep -f` self-match gotcha that the new executor check walks into
  (finding 11). The deletion is permanent: the memory directory is outside every repo, and the post-merge step
  deletes rather than trashes.
- **Fold:** the dry-run table gains a column, "facts not carried", and a row whose column is non-empty is kept or
  extracted first. Carry the Svelte exemplar paths into CLAUDE.md's Web bullet. Carry the Go acceptance dimension into
  the adapter's pair-subagent line (T6). Add one line to the adapter's STATUS close item: a STATUS rewrite moves
  items to BACKLOG and never drops an open decision. Post-merge: `tar czf` the memory directory to
  `~/.local/share/dubplate/` before any deletion, write the new `MEMORY.md` first, then delete. Re-list the
  directory first: any file added or modified after T4's stamp is kept and its index line carried, since another
  session's auto-memory may write there between T4 and the merge.

### 9. Major: T5's comment-only proof is vacuous after commit

- **Location:** plan:375-379.
- **Defect:** `git diff -U0 -- web/src/lib/themes | ...` names no base, so after the commit it diffs the working tree
  against the index, prints nothing, and passes whatever the commit changed. The filter also accepts `//` lines,
  which CSS treats as garbage before the next declaration rather than as a comment, and `*` lines, which match a
  universal selector.
- **Fold:** `git diff -U0 <task base> HEAD -- web/src/lib/themes`, preceded by a check that the diff is non-empty when
  a theme lesson landed there, and a CSS filter of `^[+-]\s*(/\*|\*/|\*\s)` only. The web leg's theme gates are a
  second net, but the proof should not lean on them.

### 10. Minor: T7's comment-only proof has two holes

- **Location:** plan:466-467.
- **Defect:** `^[+-]\s*\*` passes a Go or TS pointer write (`*p = v`) as a comment. A comment line can change
  behavior: `//go:build`, `//go:embed`, `//go:generate`, `//nolint`, `# shellcheck disable=`, `// @ts-ignore`. None of
  T7's target files carries a build constraint today (checked), so the risk is small, but the proof claims more than
  it shows.
- **Fold:** narrow the filter to `^[+-]\s*(//|#|/\*|\*/|\*\s|\*$)`, add `git diff -U0 <base> HEAD | grep -E
  '^[+-]\s*(//go:|//nolint|# shellcheck|// @ts-)'` prints nothing, and require a non-empty diff.

### 11. Minor: the new executor check self-matches and cannot see pairs

- **Location:** plan:313-314 (CLAUDE.md) and plan:387-389 (adapter).
- **Defect:** `pgrep -f <worktree path>` matches the calling shell whenever the path is in the command text
  (`~/.claude/docs/bluefin-admin.md:112-117`), so it always reports a live executor. It also cannot see Agent-tool pair
  subagents, which run inside the parent process. And STATUS's in-flight line exists only on the pass branch, so a
  session reading `master`'s STATUS sees no pass in flight.
- **Fold:** name `claude agents --json` (as the launch block at plan:97 already does) as the primary check, and write
  the pgrep with a character class (`pgrep -f 'worktrees/dubplate-alignmen[t]'`). STATUS's line says it is visible on
  the branch only.

### 12. Minor: pair base is assumed, not verified

- **Location:** plan:142-146.
- **Defect:** if the Agent tool resolves `worktree.baseRef: "head"` against the main checkout rather than the session
  worktree, a pair starts from `master` without `scripts/gate.sh`, T0's STATUS line, or T0's CLAUDE.md edit. T4 then
  conflicts with T0 on CLAUDE.md:82-92 at merge. It fails safe as a stop, but late.
- **Fold:** the dispatch names the session's head SHA, and the pair's first act is `git merge-base --is-ancestor <sha>
  HEAD` or stop.

### 13. Minor: the close does not re-run the deletion checks after merging `origin/master`

- **Location:** plan:481-485.
- **Defect:** a concurrent session on `master` that still dispatches `site-implementer` (the old adapter names it) can
  add a lesson under `.claude/agent-memory/site-implementer/`. Merging that is not a conflict (the branch deleted
  other files), so the file survives with no reader.
- **Fold:** after the merge, re-run `git ls-files .claude/agent-memory` (prints nothing) and route any new lesson as
  T5 did.

### 14. Minor: the launch check and the close leave stale states

- **Location:** plan:95, plan:493-494.
- **Defect:** plan:95 expects `" M .claude/agent-memory/dubplate-implementer/MEMORY.md"`, but `git status` now prints
  nothing, so a literal launcher stops. Close step 4 sets STATUS to "PR open for Geoff's read", which is false the
  moment the PR merges.
- **Fold:** plan:95 expects empty output. The close's STATUS line reads "merged: the post-merge list below runs next"
  in present tense after merge, or the post-merge list (finding 5) rewrites it.

### 15. Minor: `--merge` is first exercised at the close

- **Location:** plan:169-171.
- **Fold:** T0's acceptance adds `bash scripts/gate.sh --merge --dry-run`, so the close's form is proven before any pair
  runs.

### 16. Minor (over-ceremony): T7's `auth-data` class buys a security review of comments

- **Location:** plan:13-16, :464, :486-489.
- **Cost:** the class is forced by the path map, and the task-scope `diff-reviewer` read is cheap insurance for the
  comment-only proof. The close's `web-auth-security-reviewer` over comment-only edits in `internal/serve/` and two
  test files reviews no behavior. Keep the class (pass-core makes it a floor), but have the security reviewer's
  dispatch confirm the comment-only proof and stop there. The real knowledge-loss risk in this pass sits in T3 and
  T4, which are `ordinary` and `runner`; findings 4 and 8 put the scrutiny there instead.

## Concurrency and settings, checked and sound

- **Merging `origin/master` at the close.** T3 rewrites STATUS whole and T2 rewrites ROADMAP heavily, so any concurrent
  edit to either conflicts, and a conflict is a stop (plan:481). Nothing merges silently into a rewritten ledger.
- **The heavy gate lane.** `cairn-run-gate` serializes heavy legs across sessions, `check.sh` takes a per-tree `flock`
  for `tidy-check`, and exit 75 re-issues. Pairs in separate worktrees do not share a tree.
- **`.claude/settings.json`.** The removed rules (`docker compose config*`, the box `docker compose ... restart`, box
  `systemctl restart`) match no command any script or the gate runs today: the compose leg is retired (BACKLOG #11),
  and the box runs no compose stack. Removing an allow rule can only add a prompt, never break a run. The session's
  `defaultMode` is `auto`, so `bash scripts/gate.sh` needs no new allow rule.
- **Main checkout at the next pull.** The implementer directory is already clean, so the pull deletes the two tracked
  files without refusing. Only the pull's owner is missing (finding 5).
- **The auto-memory dry run.** The listing matches: 41 files plus `MEMORY.md`, and the index names each exactly once.
  The plan's four classes cover all 41.
- **T1.** The three Pi 5 reports hold prices and public links, with no personal data. The `cmp` and `stat` checks
  prove a byte-for-byte copy with the originals untouched.

## OWNER FORK

None. Every finding is a mechanism or method call. Finding 1 applies a ruling Geoff already made.

## Verdict

The plan's shape is sound for this lens: one committing executor per worktree, disjoint pair files, conflicts as
stops, a full-gate close after merging `origin/master`, and no deletion by the pass itself. Its weak point is the
knowledge it moves, which is the pass's real payload. The acceptance checks prove that removals happened (greps that
print nothing, counts the writer states) but not that each item arrived somewhere. In concrete cases they would pass
while losing a Geoff ruling. Those cases are the outreach-doc privacy ruling, the TypeScript 7 recheck, the web
player's queue requirement, and the Svelte exemplar. The records that would let anyone audit the moves sit in
scratch that a session restart discards, and the post-merge pruning and pull have no owner who is sure to run them.
None of this blocks the design. Before launch, fold findings 1 to 9: inventory-driven presence checks for T2, T3,
and T4, a "facts not carried" column, notes in the git dir, a resume paragraph, the post-merge list moved into
STATUS with the pull first and a backup before deletion, and based diffs in the comment proofs. With those folds,
the plan is safe to launch unattended.
