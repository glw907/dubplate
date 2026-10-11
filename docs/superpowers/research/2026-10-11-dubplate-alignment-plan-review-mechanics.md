# Dubplate alignment plan review: mechanics and feasibility

**Target:** `docs/superpowers/plans/2026-10-11-dubplate-alignment.md` at `61e6af5` (unchanged at `c58caf2`, which
adds only the launch prompt; the prompt is read here too). **Lens:** does every mechanism behave as stated. Every
claim below is probed or quoted from source. Scratch probes live in
`~/.claude/jobs/d7d4f1c0/tmp/review-mechanics/` (`classify-probe.sh`, `wrapper-probe.sh`, `split-probe.sh`,
`sample.diff`). No gate was run beyond one `--dry-run` scope print.

## Findings, ranked by consequence

### B1. blocker: `scripts/gate.sh` is an unmatched path, so every fast lane on the branch runs the full form, and T0's acceptance fails as written

**Location:** plan:169-171 (T0 acceptance), plan:109-115 (Gate convention), and every task acceptance that says "the
fast lane".

**Defect.** `check.sh`'s class table (`scripts/check.sh:362-397`) lists `scripts/check.sh`,
`scripts/check-state-dir.sh`, `scripts/vale-comments.sh`, and `scripts/checks/*` as machinery, and `scripts/deploy.sh`
and `scripts/provision.sh` as ops. `scripts/gate.sh` matches no row, so it is `unmatched`, and an unmatched path
promotes the run to the full form (`check.sh:441-444`: `machinery | unmatched) fallback_reason="$f is $class"; return
1`). Probe (check.sh's own `classify_path`, extracted and run over the plan's paths):

```
scripts/gate.sh -> unmatched
scripts/provision.sh -> ops
web/vite.config.ts -> web
ROADMAP.md -> noeffect
```

Without `--base`, the change set is measured from `git merge-base master HEAD` (`check.sh:307`), so T0's
`scripts/gate.sh` stays in the change set of every later run on the session branch and on every pair branch (pairs
branch from the session's commits, and their merge base with `master` still predates T0). Consequences:

1. T0's acceptance "prints a `check.sh: scope=changed` line" cannot pass. The dry run prints
   `check.sh: scope=full fallback="scripts/gate.sh is unmatched" ...` (`check.sh:531`). An executor at `sonnet`/`medium`
   meeting a failed acceptance on the first task either stops or "fixes" `check.sh`, which the plan puts out of scope.
2. Every "fast lane" in T1 to T7 and the post-merge runs is the full gate with the rendered Chromium leg on the heavy
   lane. Recorded dubplate runs (`~/.local/state/cairn-run-gate/runs.jsonl`, lean-cutover): scoped runs of 9.7, 13.8,
   and 14.9 minutes with 153 to 313 s of lock wait; the `--merge` run 19.1 minutes with 644 s of lock wait.

**Proposed fold.** Keep `check.sh` untouched (it fails closed correctly for a future edit to the wrapper). Instead:

- T0's acceptance expects the fallback: the dry run prints `scope=full fallback="scripts/gate.sh is unmatched"` and
  `gate exit: 0`; T0's one real gate is the full form, which also exercises the wrapper end to end.
- From T1 on, the fast lane passes `--base <full 40-hex sha of the commit the task started from>`, a literal the
  session records in the clock file at task start (`check.sh:236-253` accepts only a full id that is an ancestor of
  HEAD, else it falls back to full, so a wrong value costs time, never coverage). A pair's base is the dispatch commit;
  the session's post-merge run uses the same pre-dispatch commit, which covers both halves. T0's adapter edit replaces
  "Pass no `--base`" (adapter:61-62) with this rule for the wrapper case, or the plan states it as a pass-local
  convention. The close is unaffected: `--merge` is the full form.
- Alternative if the fold is refused: drop "scope=changed" from T0's acceptance and raise every estimate by one full
  gate (see M3).

### M1. major: the wrapper's `$*` forwards arbitrary words into `bash -c`, which reopens the hole the Bash guard closes

**Location:** plan:158-162 (T0 outcome); the planning probe described at plan:111-113.

**Defect.** `cairn-run-gate` runs the gate string as `exec bash -c "$2"` (`~/.local/bin/cairn-run-gate:335`). A
wrapper body `exec cairn-run-gate "bash scripts/check.sh $*"` turns any argument into shell. Probe (`printf`
standing in for `cairn-run-gate`, then the same `bash -c` shape):

```
[bash scripts/check.sh --dry-run; echo injected]
ran check.sh
injected
```

The model the plan cites, `~/.dotfiles/scripts/gate.sh`, takes no arguments, and its header says why it passes the
guard: "a no-argument wrapper passes". The plan's `$*` version passed the guard in the planning probe only because
the guard cannot see inside the script. A `runner`-class gate tool that executes `bash scripts/gate.sh '<anything>'`
as shell defeats the guard's purpose. The probe also shows the no-argument call yields the gate string
`bash scripts/check.sh ` with a trailing space, so its receipt key differs from the adapter's `'bash scripts/check.sh'`.

**Proposed fold.** T0's outcome: the wrapper accepts only `--scope-changed`, `--reduced`, `--merge`, `--dry-run`, and
`--base` followed by a 40-hex value (a `case` allowlist; anything else exits 2 with the usage line and starts
nothing), then builds the string from the validated words as `"bash scripts/check.sh${*:+ $*}"` (probe:
`[bash scripts/check.sh]`, `[bash scripts/check.sh --merge]`). T0's acceptance adds: `bash scripts/gate.sh --bogus`
exits 2 and prints no `gate exit:`.

### M2. major: the launch block cannot run from a worktree-isolated session, and the plan does not name who runs it

**Location:** plan:84-105.

**Defect.** The planning session is itself worktree-isolated (`claude agents --json` lists `dubplate-align-plan` with
cwd `.../worktrees/alignment-plan`). From such a session the guard refuses the block's first form. Probe:

```
git -C /var/home/glw907/Projects/dubplate status --porcelain --untracked-files=no
-> "this command redirects git to the shared checkout via -C. Refusing to run it"
```

pass-core says "The planning session starts it from the repo's main checkout" (pass-core:55-56), but this planning
session cannot. The other checks hold when run from a non-isolated shell: today `master` and `origin/master` are both
`8269d27` (0 behind, 0 ahead), `worktree-alignment-plan` equals `origin/worktree-alignment-plan` (`c58caf2`), and
`claude agents --json` prints a JSON array with `name`, `cwd`, `status`, and `state` per live session, as the comment
assumes.

**Proposed fold.** The launch section names the launcher: the conducting session (`c747c352`, cwd `~/.dotfiles`, not
isolated) or Geoff's terminal. If the planning session must launch, it first leaves isolation (`ExitWorktree` with
`keep`) and the plan says so. Make the agents check mechanical:
`claude agents --json | jq -r '.[] | select(.cwd | startswith("/var/home/glw907/Projects/dubplate")) | "\(.name) \(.cwd)"'`
prints only the planning session, and `git -C ~/Projects/dubplate worktree list` shows no `dubplate-alignment` entry.

### M3. major: estimates do not hold a full gate, so T0 and T1 are likely to trip the clock stop

**Location:** plan:36-37, plan:174, plan:190.

**Defect.** `claude-clock-stop` fires once elapsed minutes exceed twice the estimate
(`~/.local/bin/claude-clock-stop`: `[ "$elapsed" -gt $((estimate * 2)) ]`). T1 is 10 minutes (stop at 20) and T0 is 15
(stop at 30). Under B1 each runs a full heavy gate: 10 to 19 minutes plus a machine-lock wait of up to 11 minutes in
the recorded runs, while `engine-pass-b` (cairn) is live on the same heavy lane today. T0 also runs its dry run
through the heavy lane, which takes the same machine lock (the planning probe ran on `light`: its record shows
`"lane":"light"`). A stop here is an attended event caused by the plan, not the work.

**Proposed fold.** Fold T1 into T0 (pass-core:33: "Fold a single-deliverable task into a neighbor"; T1 is one copy
and one ROADMAP line), so one gate covers both; set the merged task to 35 minutes. Let the wrapper set
`CAIRN_GATE_LANE=light` when `--dry-run` is among its arguments (a dry run "runs no leg, takes no lock", check.sh:72-73).
With B1's `--base` fold in place, size each pair's estimate for two gates serialized on the heavy lock plus the
session's merged run.

### M4. major (over-ceremony, cost): pair subagents on docs-only diffs each run a gate that checks nothing they changed

**Location:** plan:142-151 (Pairs), plan:220, 278, 348, 427.

**Defect.** T2, T3, T4, and T6 change only `ROADMAP.md`, `docs/`, `BACKLOG.md`, `CLAUDE.md`, and `.claude/`, all
class `noeffect` ("docs, .claude, ... the top-level markdown files: gitleaks only", check.sh:105-106). pass-core's Pairs
section asks each subagent only to run setup, commit, and never push; the session merges and runs the fast lane
(pass-core:74-78). The plan adds a gate per subagent, so each docs pair costs two `npm ci` runs and three heavy-lane
gates (two serialized on the machine lock, one after the merge), against one gate if the session ran the two tasks
serially. Under B1 all three are full gates.

**Proposed fold.** A pair subagent whose diff is entirely `noeffect` skips `npm ci` and the gate and commits; its
acceptance greps still run. The session's post-merge fast lane is the gate for both halves. T5 and T7 (web and Go
paths) keep their own gate. This removes about four gate runs from the critical path.

### M5. major: T4 must write outside its worktree, which the launch prompt forbids a pair to do

**Location:** plan:338-340 and :344-348 (T4), plan:135-137 (PR notes), prompt.txt:29-31.

**Defect.** T4 is a pair subagent, but its outcome writes `$CLAUDE_JOB_DIR/tmp/MEMORY.md.proposed` and appends to
the PR notes, and its acceptance greps that file and runs `find ... -newer` against a stamp "written at task start".
The prompt tells each pair it "writes nothing outside it: ... none in `~/.claude/`", and `$CLAUDE_JOB_DIR` is under
`~/.claude/jobs/` (this session's: `/var/home/glw907/.claude/jobs/d7d4f1c0`). The Conventions also say a pair "puts
it in its final report and the session appends it". The stamp file has no stated home.

**Proposed fold.** T4 returns the dry-run table and the proposed index text in its final report. The session writes
`$CLAUDE_JOB_DIR/tmp/MEMORY.md.proposed` and runs check 2 over it after the pair returns. Replace the stamp with the
clock file's start epoch: `find ~/.claude/projects/-var-home-glw907-Projects-dubplate/memory -newermt "@<start>"`
prints nothing (probe: with a start one minute ago it printed 0 lines; with `@1000` it printed all 43 entries, so the
form discriminates).

### M6. major: T3 cannot receive T2's TypeScript 7 text, because the two run at once

**Location:** plan:209-210 (T2 hands the text to T3 "through the session: T2's report quotes it"), plan:247.

**Defect.** T2 and T3 are dispatched together and return together, so T3 never sees T2's report. Either T3 omits the
item, or the session writes it into `BACKLOG.md` after the merge, an untracked edit in no task's commit and outside
both acceptances.

**Proposed fold.** T3 writes the BACKLOG item itself from `ROADMAP.md:111-113` at its base (the text is three lines and
the file is readable there), tagged `#rung:integration`; T2 only removes it from ROADMAP. No handoff.

### M7. major: the T1 copy may be refused by the auto-mode classifier, and the reports may carry personal data

**Location:** plan:176-190 (T1).

**Defect.** Copying the three reports out of `~/Projects/dubplate/reports/` into scratch was denied in this review
by the auto-mode classifier: "Permission for this action was denied by the Claude Code auto mode classifier. Reason:
[PII Data Handling]". A count-only grep for email, phone, and address patterns found 2 hits in `hardware.md` and 1
in `verified parts.md`, 0 in `proven build.md` (this review did not read them further). The executing session runs
in auto mode, so T1's `cp` may be refused the same way, and a byte-for-byte commit pushes whatever those lines are to
`glw907/dubplate`. The plan names no report for the refused state.

**OWNER FORK.** Options: (a) commit the three files byte for byte as directed; (b) commit them after removing personal
lines, dropping the `cmp` acceptance for the edited files; (c) commit `proven-build.md` only (0 hits; it is "the
current decision") and leave the evidence files uncommitted. **Recommendation:** (c) unless Geoff confirms the hits
are harmless, with T1's refusal state named: a classifier denial is a stop (STATUS question, `notify-send`), never a
workaround through another tool.

### M8. major: check 2's grep passes vacuously on wrapped lines and misses "lanes"

**Location:** plan:29-31 (end-to-end check 2), plan:217-220 (T2 acceptance).

**Defect.** The pattern is matched per line, and the repo wraps prose. Run on the current tree, check 2 finds no hit
in the adapter even though the adapter still carries the registry rule, split across lines: adapter:18-19 "check the
lane / registry". ROADMAP:276 "Two owed dubplate lanes" matches neither check 2 nor any T2 grep. T2's
`grep -n -E 'rung [0-9]+'` is case-sensitive: it found 21 lines today, against 26 with `-i`, missing "Rung 12: the
built-in streamer" (ROADMAP:67) and "Rung 12's slice 1" (:44); it also misses "rungs 7 and 9" (:87).

**Proposed fold.** Run check 2 a second time over joined text: `tr '\n' ' ' < <file> | tr -s ' ' | grep -o -i -E
'<same pattern>'`. Add a `lanes?` grep to T2 and T3 with the same heavy, light, or fast exception T4 and T6 use. Make
T2's number grep `grep -n -i -E 'rungs? [0-9]+'`.

### M9. major: T5's comment-only grep has no base, so after the commit it passes vacuously

**Location:** plan:375-378.

**Defect.** The prose says "against the task base", but the command is `git diff -U0 -- web/src/lib/themes | ...`,
which compares the working tree with the index. Once the task commits, it prints nothing whatever the commit changed.

**Proposed fold.** `git diff -U0 <task base> -- web/src/lib/themes | ...`, as T7 already writes it, with the base
being the dispatch commit.

### M10. major: T7's comment-only grep misses some code lines in each language

**Location:** plan:466-467.

**Defect.** Probe on a scratch diff (`sample.diff`) through the plan's exact pipeline. It caught `return nil` to
`return err`, a code line with a trailing comment, a TS option change, and a shell `local` change. It let these code
lines through:

- a Go line starting with `*` (pointer store `*p = v`), because `\*` is allowed for every language;
- a shell or TS line starting with `--` or `++` (`--flag-at-col-0 \`), because `^[+-][^+-]` drops any line whose
  second character is `+` or `-` to skip the `+++`/`---` file headers;
- a `#` line inside a shell heredoc or a Go raw string.

**Proposed fold.** Per-language filtering, keyed on the file from the `diff --git` header: Go allows only `//`,
shell only `#` (not `#!`), TS `//`, `/*`, or `*`; skip only the real `+++ `/`--- ` header lines. The probe
`split-probe.sh` (an 11-line `awk`) flags all three misses above and passes the comment lines. A residual stays for a
`#` line inside a heredoc; T7 touches `provision.sh` only above `ssh_cidrs_json` (`:123`), so the diff-reviewer read
covers it.

### M11. major: no T7 check proves "no new process citation", and gofmt can redden a doc comment edit

**Location:** plan:129-131, plan:471.

**Defect.** The process-citation leg only warns (`CHECK_CITATIONS=warn`, check.sh:127-130), so "the gate's
process-citation leg warns on one" proves nothing at `gate exit: 0`. The vale leg lints only `.go` comments
(`.vale.ini`: `[*.go]`), so an em dash in a shell or TS comment passes the gate. Separately, golangci-lint runs
`gofmt` as a formatter (`.golangci.yml`), and gofmt rewrites doc-comment lists, headings, and code blocks since Go
1.19. A moved lesson written as a `- item` list fails the gate as a formatting diff.

**Proposed fold.** Add to T7's acceptance: (1) `git diff -U0 <task base> -- '*.go' | grep '^+' | grep -E 'ruled,
conductor|Resolution R|rung [0-9]|Task [0-9]|\(C[0-9]|plans/.*\.md:[0-9]|_salvage/.*:[0-9]|Ruled, '` prints nothing
(the leg's own pattern, check.sh:781); (2) `git diff -U0 <task base> | grep '^+' | grep -c $'\xe2\x80\x94'` (U+2014) prints 0 for T5
and T7; (3) `gofmt -l` on the touched `.go` files prints nothing before the gate runs.

### m1. minor: T7's serve and doctor lessons already live in the code, so moving them adds an auth-data edit for nothing

**Location:** plan:448-449, :454-456.

**Defect.** `Shell.Register` and `Shell.Handler` already say what the serve lesson says: "A path under the owned
prefix that no registered subtree claims ... answers 404" (`internal/serve/serve.go:64-67`) and "it judges only a
path normalization has admitted and covers every registered subtree" (`:140-148`). Also, `serve.go:59` is a struct
field (`mux: http.NewServeMux()`) and `origin.go:78` is the bypass mux (`bypass: http.NewServeMux()`); neither
holds a doc comment. The `doctor` lesson is largely carried by `storeRowLines` and `membersRowLines` in
`cmd/dubplate-server/doctor_test.go:22-33`, which state the opposite emphasis: "It is a literal, so a row dropped from
doctor changes the count a test expects". A comment saying "never from a literal count" beside it would contradict it.

**Proposed fold.** T7 drops the serve lesson as a duplicate, citing `serve.go:64-67` and `:140-148`, and reconciles
the doctor lesson against `doctor_test.go:22-33` before writing anything. The cobra fact stays (verified:
`UsageFunc` writes to `c.OutOrStderr()`, which is `c.getOut(os.Stderr)`, cobra `command.go:398-399, :451-454`). Quote
that source in T7's text, per the adapter's planning-miss rule. The pass stays `auth-data` through
`cmd/dubplate-server/main_test.go`.

### m2. minor: CLAUDE.md's Gate section keeps prescribing the bare form

**Location:** plan:163-164 (T0 edits only the Shell bullet), plan:286-314 (T4 does not list the Gate section).

**Defect.** `CLAUDE.md:70` still reads "`bash scripts/check.sh` before every commit". After this pass, the next
session reads that, while the adapter says to use the wrapper and the plan says "Never fall back to a bare
`bash scripts/check.sh`". `.claude/settings.json` allows `Bash(bash scripts/check.sh)` and has no rule for the
wrapper.

**Proposed fold.** T0 changes the Gate section's opening to name `bash scripts/gate.sh`, and says it runs `check.sh`
through `cairn-run-gate`. T6 may add `Bash(bash scripts/gate.sh*)` beside the existing rule.

### m3. minor: the rung 7 spec citations are off, and the retitle leaves two dangling references

**Location:** plan:413-417.

**Defect.** "It goes to Geoff before any plan is written" is on line 8, not line 7. "Owed records" starts at `:1233`
(errata to `:1257`, "Other records" to `:1262`), not `:1240`. The spec cites the section by name at `:165`
("owe amendments ("Owed records")") and `:877` ("the strings "Owed records" lists for redraft"), and "No other line
changes" leaves both pointing at a heading that no longer exists. "Other records" (the BACKLOG #8 version fix) has no
disposition.

**Proposed fold.** Cite `:8` and `:1233-1262`. Either keep the heading and add a one-line note under it, or allow the
two references to change. T3 takes "Other records"' BACKLOG #8 fix, since it owns BACKLOG.

### m4. minor: "one changed line" in each frozen spec fails when a blank separator line is added

**Location:** plan:426-427.

**Defect.** Under the parallel-build spec's status paragraph (`:3`), a new line with no blank line before it joins that
paragraph in Markdown, so a correct edit adds two lines (one blank). `git diff --stat` then shows 2 insertions.

**Proposed fold.** `git diff --numstat <base> -- <spec>` shows 0 deletions and at most 2 insertions, one of them blank.

### m5. minor: T6's settings check is a read, not a test

**Location:** plan:425-426.

**Defect.** The command prints all 17 rules (probe output today includes `Bash(docker compose config*)` and the two
`ssh root@5.78.148.42 ... restart` rules), and a person has to scan them.

**Proposed fold.** `python3 -I -c 'import json; a = json.load(open(".claude/settings.json"))["permissions"]["allow"];
print([r for r in a if "docker compose" in r or "systemctl restart" in r])'` prints `[]`, and
`"Bash(ssh root@5.78.148.42 systemctl reset-failed *)" in a` is true.

### m6. minor: the score misses pair gate time

**Location:** plan:507.

**Defect.** `--records` sums by toplevel and branch: "prints one line summing, for that toplevel and branch, the gate
time" (cairn-run-gate header). Pair runs happen in other worktrees on other branches, so the session's
`--records <toplevel> <branch>` omits them. Probe (records do score wrapper-built strings): `cairn-run-gate --records
.../alignment-plan worktree-alignment-plan` printed `gate 1s (1 runs), lock wait 1s`, counting the planning probe's
`bash scripts/check.sh --scope-changed --dry-run`.

**Proposed fold.** Each pair's final report names its worktree path and branch; the close sums `--records` over the
session's and each pair's.

### m7. minor: the dependency sweep is neither run nor waived

**Location:** adapter:45-47 ("Every pass opens with a `dependency-upgrade` sweep"); the plan never mentions one.

**OWNER FORK.** Options: (a) run the sweep at the pass's start, which can move `go.mod` or `web/package-lock.json`
(both machinery, so full gates) into a process-only pass; (b) waive it with one line in the PR body and open
`admin-api-1`, the next code pass, with the sweep. **Recommendation:** (b), since this pass ships no behavior and the
sweep's rule exists for passes that build on dependencies.

### m8. minor: small mechanics the plan leaves implicit

- T0's STATUS line needs the session id. It is in `$CLAUDE_CODE_SESSION_ID` in a background session (present in this
  session's environment). Name it.
- The adapter line refs T0 and T6 cite (`:58-64`, `:93`, `:18-22`, `:29-31`, `:32-33`, `:43-44`, `:53-54`, `:75-77`)
  all hold. T0's CLAUDE.md edit can shift T4's later refs (`:101-102`, `:135`, `:170`, `:174-177`) by a line; T4
  should find them by content.
- Compound shell commands that write outside the worktree were refused by the guard in this review ("too complex to
  verify"), while a plain `echo ... >> "$CLAUDE_JOB_DIR/tmp/..."` passed. The PR-notes convention should say "one
  plain command per write, or the Write tool".
- The session's merges of the pair branches should say `git merge --no-edit <branch>`.
- Nested dispatch: pass-core asserts that `isolation: "worktree"` from a session already inside an `EnterWorktree`
  worktree branches from that session's `HEAD`, citing `worktree.baseRef: "head"`
  (`~/.claude/docs/claude-tooling.md:65`), and the plan does not probe it. A one-line check after each dispatch
  closes it: `git merge-base --is-ancestor <dispatch commit> <pair branch>` succeeds.

## Verified as stated

- The wrapper premise. `check.sh` parses `--scope-changed`, `--reduced`, `--merge`, `--base <sha>`, and `--dry-run` in
  any order, rejects unknown arguments with exit 2 (`check.sh:156-184`), and a dry run takes no lock
  (`:201-208`). Through `$*` each flag arrives as one word (probe: `[bash scripts/check.sh --scope-changed
  --dry-run]`). `cairn-run-gate` keys state by gate string and cwd, and the wrapper `cd`s to the repo root, so the key
  holds whatever the caller's cwd. Exit 75 stays re-issue, and `--records` scores the runs (probe above).
- Pair file sets are disjoint, generated outputs included. `web/build/` is ignored except the committed placeholder
  (`git ls-files web/build` lists only `placeholder.txt`). The theme gates blank CSS comments before parsing
  (`g1.ts:143-145`, `g2.ts:140-155`), so T5's comments cannot change a gated value. T0's adapter and CLAUDE.md
  edits precede T4 and T6 sequentially.
- Cited files and lines: the package docs in `stack/stack.go`, `session/session.go`, `grade/finding.go`, and
  `fixture/doc.go` (no `go:build` tag; `go doc ./internal/harness/subsonic/fixture` prints the package doc);
  `Handle.Update` at `write.go:64`; `ssh_cidrs_json` at `provision.sh:123` with callers at `:143` and `:161`;
  `web/vite.config.ts` exporting `kitOptions` with `adapter: adapter()` at `:68` (no `svelte.config.js` exists);
  `## Component layer` at VI:226 and `## Copy` at VI:314; `## Process` at build spec `:1143`; parallel-build status
  at `:3`; every STATUS range (`:9-11`, `:21-24`, `:26-31`, `:33-40`, `:42-60`) and ROADMAP range (`:14-65`, `:67-113`,
  `:111-113`, `:117-142`, `:214-215`, `:221-225`, `:292-306`); `standing-rules.md` "Lessons" at `:113-170` with each
  subsection where T7 says; the 14a spec fold's "Owed errata" items 1 to 8 and "Conductor-owed".
- The auto-memory directory holds 41 files plus `MEMORY.md`, and the plan's four provisional classes cover all 41.
  `site-implementer` holds 11 lessons plus `MEMORY.md`. The three Pi 5 originals exist under the names T1 gives.
- `go vet -tags oracle ./cmd/harness/... ./internal/harness/...` exits 0 on the current tree.
- `git merge --no-edit worktree-alignment-plan` from a branch at `master` is a fast-forward (`master` is an ancestor).
  Command substitution does not re-evaluate the prompt's backticks.
- Each task carries a clock estimate. The plan quotes the guard probe. It quotes neither the class table (B1) nor
  cobra (m1).

## Verdict

The plan's structure is sound, and its citations are accurate almost everywhere. Its central mechanism, the gate
wrapper, needs three changes before launch. `scripts/gate.sh` is an unmatched path, so T0's own acceptance fails and
every later "fast lane" becomes a full browser gate (B1). The `$*` body hands arbitrary words to `bash -c` (M1). The
launch block cannot run from the isolated planning session that would launch it (M2). Together with the estimates
(M3) and the per-subagent docs gates (M4), B1 makes an unattended clock stop on T0 or T1 likely. Fold B1 with a
recorded `--base`, harden the wrapper with an allowlist, name the launcher, fold T1 into T0, and drop gates on
docs-only pair halves. Then repair the four acceptance greps that pass vacuously or leak code changes (M8 to M11),
re-route T4's job-dir writes and T3's TypeScript 7 handoff (M5, M6), and get Geoff's call on the Pi 5 files (M7).
With those folds the plan is executable as written.
