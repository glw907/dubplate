# Dubplate alignment plan review: contract and criteria lens

Target: `docs/superpowers/plans/2026-10-11-dubplate-alignment.md` at `61e6af5` (branch `worktree-alignment-plan`).
Read in full: the plan, the design (`specs/2026-10-11-dubplate-alignment-design.md`), the audit
(`research/2026-10-11-dubplate-alignment-audit.md`), `pass-core`, the `dubplate-pass` adapter, the format model
(`~/.dotfiles/docs/superpowers/plans/2026-10-11-infra-sweep-1-guard.md`), and the planning prompt. Every cited repo
line was read at the cited location; the Path to 0.8 was checked against the dependency matrix, the reverse-edge note,
the re-plan's "The ladder after this spec", the 14a spec, the rung 7 spec, and the streamer spec's hooks. Paths below
are relative to the repo unless they start with `~`. The design's five rulings were not reviewed.

## Findings, ranked by consequence

### 1. Major: T2 and T3 run as a pair, but T3 needs T2's output

- **Location:** plan:209-210, plan:247, plan:222.
- **Defect:** T2 says the TypeScript 7 recheck "moves to BACKLOG as one item tagged `#rung:integration` (T3 owns
  BACKLOG, so T2 hands the text to T3 through the session: T2's report quotes it)", and T3's outcome lists "T2's
  TypeScript 7 item". The two run concurrently as a pair, so T2's report arrives only after T3 has returned. T3
  either stalls, invents the text, or omits the item; T3's acceptance never checks for it.
- **Fold:** T3 reads the source text itself from the base commit (`ROADMAP.md:111-113`, "recheck TypeScript 7
  (7.0.2 at the time) ... sveltejs/language-tools #2891") and writes the BACKLOG item. Add to T3's acceptance:
  `grep -n 'TypeScript 7' BACKLOG.md` prints one line tagged `#rung:integration`.

### 2. Major: T2 deletes a Geoff ruling whose only home is ROADMAP

- **Location:** plan:206-208 ("the streamer's settled list (`:78-110`) collapse into pointers to the organizer and
  streamer specs").
- **Defect:** `ROADMAP.md:102-109` carries "The web player needs a strong play-queue interface (Geoff, 2026-09-24):
  now playing, up next, reorder, and add from album, track, or filter results. It is designed from a prior-art
  study ...". A repo-wide grep for "play-queue interface" or "strong play" finds it only in `ROADMAP.md`; the
  streamer spec's "The 2.1 list" (`:1092`) names "the first-party web player" and nothing more. Collapsing the
  range into a pointer deletes the ruling. The plan never requires a home check before a collapse.
- **Fold:** T2 checks each sentence of `:49-65` and `:78-113` against the spec it points to (a `grep` of a
  distinctive phrase), and moves any sentence with no home into its new home before deleting it. The play-queue
  ruling goes to `IDEAS.md` or the streamer spec's 2.1 list as a one-paragraph edit (a living reference, edited in
  place under the new Errata rule). T2's report lists each moved sentence and its destination.

### 3. Major: T3 resolves only the first hop of the HISTORY pointers

- **Location:** plan:241-244.
- **Defect:** T3 reads "Moved out of STATUS at this close" at `HISTORY.md:316, 511, 588`, "Routed items" at `:443`,
  and the cookie test. The section at `:588` is itself a pointer list: `HISTORY.md:626` reads "To `give-back` and
  `packaging`: the carries listed in HISTORY, re-plan close, 'Moved out of STATUS'", and `:627` reads "Photo lane,
  after lane 1: three owed checks and a retest (HISTORY, re-plan close, 'Moved out of STATUS')". Those land at
  `HISTORY.md:697-708`, which hold live carries (`tools/photos/check.test.mjs` fails 2 of 121; the integrity leg
  reads `photos/` from disk; Ruling 3's photo erratum to rung 7; the give-back and packaging carries, which point
  again at "HISTORY transliteration, 'Routed'"). `:612-621` also routes "To rung 7" carries via "HISTORY 5c-iv 'To
  rung 7'". Once STATUS stops pointing at HISTORY, nothing reaches these items.
- **Fold:** state the rule as transitive: follow each pointer until it reaches an item, and either write the item
  into BACKLOG with its `#rung:` tag or record it as already carried (name the rung 7 spec line, the ROADMAP entry,
  or the BACKLOG number). Add `HISTORY.md:612-627` and `:697-708` to the cited list, and cite the cookie test at
  `HISTORY.md:802-803`. The side-by-side map in T3's Test covers each terminal item, not each STATUS line.

### 4. Major: the end-to-end grep is blind to wrapped phrases and passes on the adapter today

- **Location:** plan:29-31 (end-to-end check 2), reused by T2, T3, T4, and T6 acceptance.
- **Defect:** the grep is line-oriented, and the phrases wrap. The current adapter reads "check the lane" at
  `SKILL.md:18` and "registry for a live executor" at `:19`, so check 2 prints nothing for the adapter before the
  pass starts: the check is vacuous there. Run against the tree at `61e6af5`, check 2 finds only `STATUS.md:26`,
  `CLAUDE.md:174`, and `ROADMAP.md:215, 306`. T4's and T6's `lane` greps cover CLAUDE.md and the adapter, but STATUS
  and ROADMAP get no such cover, and `ROADMAP.md:276` ("Two owed dubplate lanes") survives T2 as written ("Other
  Planned entries keep their substance").
- **Fold:** join lines before matching, for example `tr '\n' ' ' < FILE | grep -o -i -E 'lane[[:space:]]+registry|...'`,
  and add a second grep over STATUS and ROADMAP for `\blanes?\b` (excluding `fast lane`, `heavy lane`, `light lane`),
  `conductor`, and `\bowed\b`. Run the joined grep against the current adapter once at T0 to prove it now finds the
  wrapped "lane registry" (the fixture that shows it can fail).

### 5. Major: T5's comment-only proof has no base, so it passes on any committed change

- **Location:** plan:375-379.
- **Defect:** the prose says "against the task base", but the command is `git diff -U0 -- web/src/lib/themes | ...`,
  which compares the working tree to the index. After the subagent commits, it prints nothing whatever the commit
  changed, and "prints nothing" is the pass condition. A changed token value would pass. T6's "`git diff --stat`
  shows one changed line in each of the two frozen specs" (plan:426-427) has the same missing base (it fails
  noisily rather than passing, but cannot be run as written), and T7 writes `<task base>` without saying what it is.
- **Fold:** define the task base once in Conventions (the session commit the pair was dispatched from, recorded
  before dispatch as `base=$(git rev-parse HEAD)` and passed in the dispatch text), and write every diff proof as
  `git diff -U0 "$base"..HEAD -- <paths>`. For T6, use `git diff --numstat "$base"..HEAD -- <spec>` and expect
  `1 0` or `2 0` per frozen spec (a header under a heading may need a blank line). Prove T5's grep can fail once by
  running it against a scratch edit of a token value.

### 6. Major: T3's planning prompt acceptance omits the edits it exists to carry

- **Location:** plan:255-256, plan:268, plan:277-278.
- **Defect:** with the owed list retired, the admin-api-1 planning prompt is the only carrier for the two
  living-reference edits (CH:171-172 and the PS note). Acceptance checks only "the 14a spec path, `auth-data`, the
  three lenses, and the push target". The "PS member-cap note" is also underspecified: its source is the old plan,
  `plans/2026-10-10-rung-14a-admin-api.md:831-833` ("PS gains a note that the member cap landed at 14a, and an
  erratum to PS:145-147"), and the member cap itself is that plan's addition, not the 14a spec's.
- **Fold:** acceptance greps the prompt for `CH:171-172`, `PS:145-147`, and `2026-10-10-rung-14a-admin-api.md:829`,
  and for the EnterWorktree name `admin-api-1-plan`. T3's outcome names the PS edit's source lines and says the
  re-plan decides whether the member cap stays in scope.

### 7. Major, OWNER FORK: the adapter's per-pass dependency sweep is neither run nor waived

- **Location:** absent from the plan; the rule is `.claude/skills/dubplate-pass/SKILL.md:45-49`: "Every pass opens
  with a `dependency-upgrade` sweep ... A sweep with nothing to move is one line in the PR body." The audit (section 3)
  lists it under items to keep as "Geoff's 2026-10-04 ruling".
- **Defect:** the plan is silent, so the executor either skips a standing ruling unrecorded or runs a sweep that can
  pull `go.mod`, `web/package.json`, and pinned-binary bumps (and libxml2 golden regeneration) into a docs-and-process
  PR.
- **Options:** (a) T0 opens with the sweep, as the adapter says, and its result goes in the PR body (adds roughly 15
  to 60 minutes and can widen the diff into `auth-data` dependencies); (b) Out of scope gains "the dependency sweep:
  this pass changes no dependency, and `admin-api-1` opens with one", recorded as one PR-body line.
- **Recommendation:** (b). The pass touches no dependency, the next pass sweeps within the day, and a mixed PR costs
  Geoff's read more than the delay costs currency. Because the rule is Geoff's, the waiver is his to grant.

### 8. Minor: T0's proofs skip the two forms the close depends on

- **Location:** plan:169-171, plan:28, plan:482.
- **Defect:** T0 proves only `--scope-changed --dry-run` and `--scope-changed`. The close's only full-suite proof is
  `gate.sh --merge`, first run 300 minutes later. `check.sh:75` says "--dry-run works with any form" and prints
  `scope=full+history` for `--merge` (`check.sh:537`). T0 also leaves the adapter's Full-suite home line ("Run one
  `bash scripts/check.sh` on the close head", `SKILL.md:66-67`) and Close checklist line 91 naming the bare form,
  which the plan's own convention forbids ("Never fall back to a bare `bash scripts/check.sh`").
- **Fold:** add `bash scripts/gate.sh --merge --dry-run` (prints `scope=full+history` and `gate exit: 0`) and
  `bash scripts/gate.sh --dry-run` (prints `scope=full`). Widen T0's adapter edit to lines 58-69 and 91-95. Consider
  adding `Bash(bash scripts/gate.sh*)` to `.claude/settings.json` in T6, beside the existing `Bash(bash scripts/check.sh)`.

### 9. Minor: T4's conditional duplicates have no "not found" report

- **Location:** plan:332-335.
- **Defect:** six memory files are "duplicate" "each only if `grep` finds its ruling in the named spec". The absent
  state has no class: keep, move into CLAUDE.md, or move into the spec. The executor will choose ad hoc.
- **Fold:** "If the grep finds nothing, the file is classed keep, with the missing ruling named in the reason, and the
  PR body's Found, not done list carries it for the conducting session."

### 10. Minor: T4 writes outside its worktree against the plan's own convention

- **Location:** plan:118-121 ("No task writes, moves, or deletes a file outside its worktree") against plan:338-340
  (T4 "writes the proposed index ... to `$CLAUDE_JOB_DIR/tmp/MEMORY.md.proposed`") and plan:135-137 (a pair
  subagent puts PR notes in its report).
- **Defect:** T4 is a pair subagent, so its direct job-dir write breaks both rules, and close check 2 then reads a
  file whose writer is ambiguous.
- **Fold:** T4 returns the dry-run table and the proposed index in its final report; the session writes both to the
  job directory, as it does for every other pair's notes.

### 11. Minor: T2's rung-number grep is case-sensitive and misses plurals

- **Location:** plan:218-219.
- **Defect:** `grep -n -E 'rung [0-9]+' ROADMAP.md` does not match `### Rung 12: the built-in streamer`
  (`ROADMAP.md:67`), "Rung 1 (foundation)", or "rungs 7 and 9", so a surviving heading or sentence passes.
- **Fold:** `grep -n -i -E '\brungs? [0-9]' ROADMAP.md`. Apply the same grep to CLAUDE.md in T4, whose "Remaining
  rung numbers become names" (plan:312) has no check at all; the expected survivor is the one Versions sentence about
  older documents.

### 12. Minor: T7's "moved, not dropped" promise is self-graded

- **Location:** plan:468-470.
- **Defect:** the section's deletion is checked, but arrival is checked only by "`go doc` ... print the moved text"
  for two of nine destinations, with no phrase named. The other seven rest on the subagent's own report.
- **Fold:** one grep per destination for a distinctive phrase of its lesson, for example `GONIC_PROXY_PREFIX` in
  `internal/harness/subsonic/stack/stack.go`, `artist-set` in `session/session.go`, `two seconds` or `2 s` in
  `fixture/doc.go`, `golden grades the grader` in `grade/finding.go`, `paths.relative` in `web/vite.config.ts`,
  `errexit` in `scripts/provision.sh`, `SetOut` in a comment in `cmd/dubplate-server/main_test.go`, and `Thresholds`
  in `doctor_test.go`. Also: the comment-only filter's `\*` exception admits a Go statement beginning `*p = ...`;
  restrict `\*` to `.ts` and CSS.

### 13. Minor: dispatch text omits skills CLAUDE.md makes mandatory for T7's files

- **Location:** plan:145-146.
- **Defect:** T7 edits a comment in `web/vite.config.ts` (CLAUDE.md:100-101: "Invoke `svelte-conventions` and
  `ts-conventions` before writing components or TS") and in `scripts/provision.sh` (CLAUDE.md:107-108: "invoke the
  `vps-conventions` skill BEFORE touching ... deploy/provision scripts"). The dispatch names only `go-conventions`.
- **Fold:** T7's dispatch names `ts-conventions` and `vps-conventions` too.

### 14. Minor: T3's "Current state" misstates the box and drops three STATUS lines

- **Location:** plan:229-230; `STATUS.md:5-11, 58`.
- **Defect:** "nothing deployed" contradicts `STATUS.md:9-10` ("dubplate-server binary `d1e93e9`, the serve unit
  installed and disabled"); STATUS's own words are "nothing deploys". Lines 6-8 (slice 3's inputs, the 52 owed ledger
  rows, the Tempus walk result) have no assigned destination, and `:58` ("two dead `docker compose` entries ... for
  Geoff to drop") would become a BACKLOG item under "one per distinct item" although T6 removes them in the same PR.
- **Fold:** write "nothing deploys"; route `:6-8` to BACKLOG `#rung:streamer` or retire them in the map; mark `:58`
  "done by T6". Give `make derive` its tag (`#rung:pipeline` or whichever rung next runs the oracle).

### 15. Minor: T3's BACKLOG floor is set by the executor

- **Location:** plan:276-277.
- **Defect:** "`grep -c '#rung:' BACKLOG.md` is at least the number of distinct items at `STATUS.md:42-60` (the
  subagent states its count)". BACKLOG has 0 `#rung:` lines today, so the check is not vacuous, but the subagent
  sets its own bar, and `grep -c` counts lines, not items.
- **Fold:** the plan states the floor (a read of `STATUS.md:42-60` finds about 40 items, about 12 of them existing
  BACKLOG numbers), and the map in T3's report is the proof; the count is a sanity check.

### 16. Minor: two end-to-end checks are met before the pass starts or have no command

- **Location:** plan:32, plan:33-34.
- **Defect:** check 3 (`wc -l < docs/STATUS.md` prints 60 or less) passes on today's STATUS, which is exactly 60
  lines (`STATUS.md` ends at line 60). Check 4 has no command.
- **Fold:** check 3 adds `grep -c '^## ' docs/STATUS.md` prints 3 and the three headings are Current state, Next
  action, Open decisions. Check 4 extracts the backticked names between `## Path to 0.8` and `## Active` and compares
  their order to the plan's twelve items.

### 17. Minor: citations that will mislead the executor

- plan:50-52 calls `(14,9)` a matrix hard edge. The matrix (`rung-dependency-analysis.md:58`) marks row 9, column 14
  `S`; the HARD class is `research/2026-09-04-reverse-edge-completion.md` "(14,9) | **HARD**", which the 14a spec's
  owner ruling 1 cites. The Path's sources (plan:41-44) should add that note; its (12,11), (14,12), and (14,13) cells
  agree with the derived order.
- plan:413: the rung 7 spec's "It goes to Geoff before any plan is written" ends on line 8, not line 7.
- plan:415: the Owed records section starts at `:1233`, and its "Other records" (`:1257-1260`, BACKLOG #8's
  `tus-js-client` note) gets no disposition.
- plan:166 cites adapter line 93 for the `--merge` sentence, which spans `:92-95`.

### 18. Minor: the pass-class header names the wrong class first and omits `live-account`

- **Location:** plan:13-17.
- **Defect:** "**Pass class:** `runner`" is followed by "so T7 is `auth-data` and the pass takes `auth-data`". The
  session pushes and opens a PR on GitHub, which pass-core flags `live-account` ("a run that can touch a real
  Cloudflare, GitHub, or npm account"); the format model sets the flag for the same act.
- **Fold:** "**Pass class:** `auth-data` (T7 by path; the design's `runner` is the floor), with the `live-account`
  flag for the push and PR."

### 19. Minor: the executor loads the old memory index at launch

- **Location:** plan:98-99 (launch from `~/Projects/dubplate`).
- **Defect:** a session started in the main checkout loads `~/.claude/projects/-var-home-glw907-Projects-dubplate/memory/MEMORY.md`,
  whose lines 9-10 and 18 teach `pass-execute` and two-lane mechanics. Its own memory writes would also land in the
  directory T4's `find -newer` check watches.
- **Fold:** the launch prompt (not yet in the tree: `2026-10-11-dubplate-alignment-prompt.txt` is absent at
  `61e6af5`) carries the planning prompt's override sentence ("the design overrides them") and tells the session to
  write no auto-memory. The verification read confirms the prompt file is committed before the launch block runs.

### 20. Minor: T3's estimate invites a clock stop

- **Location:** plan:36-37, plan:282.
- **Defect:** T3 rewrites STATUS, resolves the HISTORY chain (finding 3), writes about 40 BACKLOG items, writes the
  dispositions, and drafts a planning prompt, at 45 minutes; the pair's clock file takes 45 and stops at 90. T2 is
  lighter.
- **Fold:** move the admin-api-1 planning prompt to T2 (a new file, still disjoint) and set T3 to 55 minutes.

### 21. Minor: small gaps

- T6's settings proof prints the allow list for a read; make it assert: no entry contains `docker compose` or
  `systemctl restart`, `systemctl reset-failed` is present, and the count drops by exactly three (17 to 14).
- T6's slot read (plan:428-429) can be a grep for the eight bold labels.
- T6 says "No other line changes" in the rung 7 spec with no check; `git diff --numstat` on it bounds the change.
- T5 may edit one `internal/harness/oracle/*.py` docstring, which no proof covers.
- T4 keeps `photo-older-is-better`, a curation rule that can execute in the visual identity spec's "Curation rules"
  (`:277`); the plan allows T4 to correct a class, so name that home as the alternative.

## Over-ceremony, ranked by cost

1. **T1 as its own task** (10 minutes, one commit, one clock cycle). It is one deliverable, and pass-core says "Fold
   a single-deliverable task into a neighbor." Fold it into T0, which the session also runs inline and which already
   precedes the first pair.
2. **Two `web/` close reviewers for comment-only edits.** The adapter's glob brings `svelte-reviewer` and
   `daisyui-a11y-reviewer` for T5's CSS comments and T7's `vite.config.ts` comment: two Opus dispatches at the close.
   The glob is mandatory, so the cheap fold is upstream: T5 may send all six DaisyUI lessons to the visual identity
   spec, which CLAUDE.md:103-106 says "every `web/` task reads". That removes the CSS diff but not T7's one TS comment,
   so the saving is only the review's scope. Low priority.
3. **The task-scope `diff-reviewer` on T7.** Required by pass-core for `auth-data` and cheap against a comment-only
   diff; keep it. The close's `go vet -tags oracle` line is also cheap; keep it.

## Coverage of the design

Every item in the design's "What changes" list maps to a task: STATUS (T3), CLAUDE.md (T0, T4), the adapter (T0,
T6), frozen-spec headers and dispositions (T6, T3), lessons home (T7), memory (T4, T5, with the prune handed to the
conducting session as the planning prompt directed), the rung 7 gate (T6), permissions (T6), and ladder naming
(T2). The four end-to-end checks restate the design's, with the auto-memory grep moved to the proposed index. The
derived Path to 0.8 holds: every hard edge in the matrix and the reverse-edge note points backward in the list,
ruling 2 puts `intake` after `admin-api-1`, `admin-api-2` follows the first collection-row writer (14a spec, "The
split"), streamer slices 3 onward wait on `pipeline` (re-plan, "Slices 3 onward start after rung 9 lands the
catalog"), `queues` precedes slice 7 (streamer spec hook 16, "rung 13"), and 15, 16, 17 run serially.

## Verdict

The plan is sound in shape and its Path to 0.8 is correctly derived. Every design item has a task, the risk
classes follow the path map, and every task carries an estimate. It is not yet ready to launch. Three findings
would lose content without any check firing: T3 cannot receive T2's TypeScript 7 text inside a concurrent pair, T2's
collapse deletes Geoff's play-queue ruling, and T3 stops one hop short in the HISTORY pointer chain. Two proofs pass
vacuously: the headline grep misses wrapped phrases (and passes on the adapter today), and T5's comment-only diff
has no base. The planning prompt's acceptance does not check the two edits it exists to carry. The dependency-sweep
waiver is Geoff's to grant. All of these fold in place without changing the design, and the remaining findings are
minor tightening.
