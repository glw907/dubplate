# Dubplate alignment plan: fold record

**Design change: none.** No finding changes the approved design
(`docs/superpowers/specs/2026-10-11-dubplate-alignment-design.md`) or any of Geoff's five rulings. The dependency
sweep (C7, m7) applies the adapter's standing 2026-10-04 rule rather than adding scope to the design, and the class
raise to `auth-data` follows pass-core's "the design's class is a floor".

Target: `docs/superpowers/plans/2026-10-11-dubplate-alignment.md` and its launch prompt
`docs/superpowers/plans/2026-10-11-dubplate-alignment-prompt.txt`, as of `c58caf2`. Reviews:
`docs/superpowers/research/2026-10-11-dubplate-alignment-plan-review-{contract,mechanics,risk}.md` (ids below: C for
contract, CO for its over-ceremony list, B, M, and m for mechanics, R for risk). The fold followed
`superpowers:receiving-code-review` and the `spec-plan-review` fold rules. The planning session settled ten points
before the fold (cited "PS-1" to "PS-11" below, in the order the fold brief gives them); this record carries them as
dispositions.

## Probes and quotes behind the new mechanisms

- **The `--base` rule.** `scripts/gate.sh` is unmatched: `classify_path`'s last arm is `*) echo unmatched ;;`
  (`check.sh:399`), and `check.sh:458-459` sets `fallback_reason="$f is $class"` for `machinery | unmatched`. Without
  `--base`, a branch measures from `git merge-base master HEAD` (`check.sh:307`). A supplied base must be a full hex
  id that resolves to itself and is an ancestor of `HEAD`, or the run falls back to the full form
  (`check.sh:268-294`). `.gitignore`, `go.mod`, `go.sum`, and `web/package*.json` are machinery rows
  (`check.sh:367-368`).
- **The clock file's extra line.** `claude-clock-stop` reads `task=`, `start=`, and `estimate=` each with `sed -n
  's/^<key>=//p' | head -n 1`, so a fourth `base=` line is ignored (read from `~/.local/bin/claude-clock-stop`).
- **Docs-only halves.** `docs/* | .claude/* | ... | ROADMAP.md | IDEAS.md | BACKLOG.md | CLAUDE.md | LICENSE` echo
  `noeffect` (`check.sh:395-397`); the header says "gitleaks only" (`check.sh:109-110`). T2, T3, T4, and T6 touch only
  those paths once `.gitignore` moves to T0.
- **The light lane for a dry run.** `cairn-run-gate`'s header: "CAIRN_GATE_LANE=light is for a gate that launches no
  browser"; `check.sh:75-76`: a dry run "runs no leg, takes no lock". `--dry-run` works with any form
  (`check.sh:75`), and `--merge` prints `scope=full+history` (`check.sh:537`).
- **The comment-only filter.** Probed on a scratch diff (`~/.claude/jobs/d7d4f1c0/tmp/fold/probe.diff`): it flagged
  all 14 planted code or directive lines (`*p = v`, a removed `return nil`, `//go:build oracle`, `// @ts-ignore`, a
  hunk line `---flag-at-col-0`, CSS `* { color: red; }` and `// not a CSS comment`, `# shellcheck disable=SC2034`,
  `#!/bin/sh`, `local x=1`, `-++counter`, `+--counter`, `# noqa: E501`, a Python call) and passed the 6 plain
  comments. An inline `awk` program piped from `git diff` runs under the worktree guard; `awk -f <file>` and `bash
  --pretty-print` are refused, so the plan uses neither. The heredoc residual is closed for the one shell file T7
  touches: `grep -n -E '<<-?[[:space:]]*["'\'']?[A-Za-z_]' scripts/provision.sh` prints nothing (its `<<<` lines are
  here-strings).
- **The joined retired-phrase and lane checks.** Run on the planning head: the joined grep finds "lane registry" in
  the adapter (wrapped at `SKILL.md:18-19`), in CLAUDE.md, and in STATUS, and "post-mortem" twice in ROADMAP; the lane
  check prints "every lane", "Document lanes", "photo lane", "dubplate lanes", and others, and keeps "heavy lane".
- **The inventory check.** Run on the planning head as written in the plan: the row count prints 70, and the check
  prints 56 `MISSING` lines (every `B` row plus the `S` row I4); the 14 rows it passes are the carried ones. A first
  draft matched `^\| I` and extracted nothing because the table is indented; the row count is what caught it, and the
  plan now matches `^ *\| I`.
- **End-to-end check 4.** Run over the plan's own Path section (the format T2 copies), it prints the twelve names in
  the derived order.
- **`check-ignore -v`.** The main repo's `.git/info/exclude:18` already ignores the outreach doc, and a worktree
  shares it, so `-q` would pass vacuously; `-v` names the deciding pattern, and `gitignore(5)` ranks a `.gitignore`
  file above `info/exclude`.
- **Facts the fold verified against the tree.** `serve.go:64-67` and `:140-148` already carry the serve-routing
  lesson; `doctor_test.go:289-292` builds the doctor counts from `reg.OwnedWall()`, `reg.ReservedClasses()`,
  `reg.IntegrationClasses()`, and `reg.Thresholds()`; cobra v1.10.2's `UsageFunc` writes through `c.OutOrStderr()`
  (`command.go:454`), which returns `c.getOut(os.Stderr)` (`:398-399`); `OpenServe` and `applySettingsSeed` are wired
  (`cmd/dubplate-server/serve.go:300`, `:391`), so the "To 6d" carry is landed; `claude agents --json` returns `name`
  and `cwd` per live session; the 14a fold's C14 BACKLOG item is at `2026-10-10-rung-14a-spec-fold.md:91`; the old
  plan's PS note is at `2026-10-10-rung-14a-admin-api.md:829-833`; `(14,9)` is HARD at
  `reverse-edge-completion.md:37`; the display-name field still exists (`internal/auth/handlers.go:218`).

## Dispositions

### Contract lens

| Id | Disposition | Where the revision carries it |
|---|---|---|
| C1 | Folded (PS-7). T3 copies the TypeScript 7 item from its own base; T2 only removes it. | T2 Active bullet; T3 inventory row I66 |
| C2 | Folded (PS-7). T2 builds a removal map; the play-queue requirement stays in ROADMAP. | T2 "The removal map"; acceptance `grep -c -i 'up next'`; row I70 |
| C3 | Folded (PS-7). The plan carries the traced inventory, with every HISTORY pointer followed to its leaf. | T3 "The inventory" (70 rows) and its check |
| C4 | Folded (PS-10) at the root with M8: joined-text retired-phrase check and a lane check over STATUS, ROADMAP, CLAUDE.md, and the adapter. The extra `conductor` and `\bowed\b` greps are refused: "conductor ruling" is already in the phrase list, and "owed" is ordinary English that legitimately survives (owed ledger rows, an owed recording). | Conventions "Retired-phrase check", "Lane check"; end-to-end check 2 |
| C5 | Folded (PS-10), convergent with M9 and R9: every diff check reads `<base>..HEAD` and asserts a non-empty diff; T6's frozen-spec check is `git diff --numstat` expecting `1 0` or `2 0`. | Conventions "Diff checks"; T5, T6, T7 acceptance |
| C6 | Folded (PS-7). The prompt's acceptance greps `CH:171-172`, `PS:145-147`, `2026-10-10-rung-14a-admin-api.md:829`, and `admin-api-1-plan`; the outcome names the PS edit's source and leaves the member cap's scope to the re-plan. | T3 prompt bullet and acceptance |
| C7 | Not a fork (PS-1a). The adapter's standing rule binds and only Geoff can waive it, so the sweep runs as T1. | T1 |
| C8 | Folded (PS-4): T0 proves `--merge --dry-run`, and names the wrapper in the adapter's Full-suite home and Close lines and CLAUDE.md's Gate line. A plain `--dry-run` proof is dropped (the scoped dry run already exercises the wrapper's dry-run path). The `Bash(bash scripts/gate.sh*)` allow rule is refused: the session runs in auto mode (R's settings check), so no rule is needed, and a broad prefix rule widens what runs unprompted. | T0 |
| C9 | Folded. A "duplicate" whose grep finds nothing is classed keep, named in "Found, not done". | T4 dry run |
| C10 | Folded (PS-6), convergent with M5 and R6: T4 returns the table and index; the session writes them to the git dir. | Conventions "Records"; T4 |
| C11 | Folded (PS-10), convergent with M8: `grep -n -i -E '\brungs? [0-9]'`, applied to ROADMAP (T2) and CLAUDE.md (T4). | Conventions "Names"; T2, T4 acceptance |
| C12 | Folded. Each T7 destination carries a named token, grepped for arrival. The `\*` exception is gone: the new filter allows `*` nowhere. | T7 outcome tokens and "Arrival" |
| C13 | Folded (PS-11). T7's dispatch names `ts-conventions` and `vps-conventions`. | Conventions "Pairs"; prompt |
| C14 | Folded (PS-11). Current state says nothing deploys and gives the box's state; STATUS:6-8 go to rows I7 to I9; STATUS:58 is retired as T6's; `make derive` carries a rung tag (row I10). | T3 STATUS bullet; rows I5 to I10; "Retired, with no row" |
| C15 | Folded at the root with C3 and R4: the plan states the inventory, so the subagent no longer sets its own floor. | T3 inventory check |
| C16 | Folded. Check 3 asserts the three headings; check 4 is a command, probed. | End-to-end checks 3 and 4 |
| C17 | Folded (PS-11): (14,9) cites `reverse-edge-completion.md:37`; the rung 7 sentence is line 8; "Owed records" is `:1233-1262`, its title kept because `:165` and `:877` cite it; "Other records" gets a disposition (BACKLOG #8 to T3, row I67); the adapter's `--merge` lines are `:92-95`. | Path item 1; T0; T6; T3 |
| C18 | Folded (PS-11). The header leads with `auth-data` and carries `live-account`. | Pass class |
| C19 | Folded. The prompt keeps the override sentence, extends it to the auto-memory, and says "Write no auto-memory"; the Conventions say the session writes none. The prompt file is committed. | Prompt; Conventions "The main checkout stays untouched" |
| C20 | Folded in part: T3 is 55 minutes on `opus` (PS-7). Moving the planning prompt to T2 is refused: the prompt carries T3's own dispositions (the CH and PS edits), so one author keeps them consistent, and Opus at 55 minutes covers the load. | T3 estimate; Pairs |
| C21 | Folded: the settings check asserts `14 [] True`; the adapter-slot read is a grep for eight labels; the rung 7 spec change is bounded by hunk ranges; a T5 Python edit is a `#` comment the filter covers; `photo-older-is-better` names VI's "Curation rules" as its alternative home. | T6 acceptance; T5; T4 Keep list |
| CO1 | Folded (PS-4): the Pi 5 copy is part of T0. | T0 |
| CO2 | Refused: the checklist glob is mandatory, and T7's `vite.config.ts` comment brings both web reviewers whatever T5 does, so moving the CSS lessons to the spec saves no dispatch and moves a lesson away from the token it governs. | (unchanged) |
| CO3 | Kept as the reviewer recommends. | T7 acceptance; Close step 1 |

### Mechanics lens

| Id | Disposition | Where the revision carries it |
|---|---|---|
| B1 | Folded (PS-4). The fast lane passes `--base <base>` on this branch, with `check.sh`'s lines quoted; T0's acceptance expects the fallback line its dry run really prints; `check.sh` is untouched. The rule is branch-local: the adapter keeps "Pass no `--base`", since a later branch's merge base excludes the wrapper. The first live proof that `--base` narrows is the pair 1 post-merge run, which must print `scope=changed`. | Conventions "The base"; T0 acceptance; T3 "Pair" |
| M1 | Folded (PS-4): an allowlist of `--scope-changed`, `--merge`, `--dry-run`, and `--base <40-hex>`, exit 2 otherwise, no trailing space. `--reduced` is left out as the brief directs (no task uses it). | T0 outcome and acceptance |
| M2 | Folded (PS-2): the conducting session or Geoff launches, from the main checkout; the agents check is the reviewer's `jq` form. | Launch "Who launches" |
| M3 | Folded (PS-4): estimates hold a gate (T0 45, T1 60, T5 40, T7 60, close 70); T1 merged into T0; the wrapper uses the light lane for a dry run. | Clock estimate; T0 |
| M4 | Folded (PS-5), verified against `check.sh:395-397`: docs-only halves skip `npm ci` and their gate. | Conventions "Docs-only pair halves" |
| M5 | Folded (PS-6), convergent with C10 and R6; the stamp file is replaced by `find -newermt "@<start>"`. | Conventions "Records"; T4 acceptance |
| M6 | Folded (PS-7), convergent with C1 and R2. | T3 row I66 |
| M7 | Not a fork (PS-1b): the flagged lines are an EEPROM address and street pricing, so all three files are committed byte for byte; a classifier refusal is a stop. | T0; Conventions "Stops" |
| M8 | Folded (PS-10) at the root with C4 and C11. | Conventions; T2 acceptance |
| M9 | Folded (PS-10), convergent with C5 and R9. | Conventions "Diff checks"; T5 |
| M10 | Folded (PS-10): a per-language filter, probed; the heredoc residual is closed for `provision.sh` by a grep. | Conventions "Diff checks"; T7 acceptance |
| M11 | Folded (PS-10): T7 greps added lines with the `check.sh:781` pattern, T5 and T7 grep for the em dash, and `gofmt -l` runs before the gate; T7's text says plain paragraphs, since gofmt rewrites doc-comment lists. | T7 outcome and acceptance; T5 acceptance |
| m1 | Folded (PS-11): the serve lesson is dropped as carried at `serve.go:64-67` and `:140-148`; the doctor lesson is dropped as carried at `doctor_test.go:289-292` and `:22-33`; the cobra fact stays, quoted from `command.go`. T7 no longer edits `internal/serve` or `doctor_test.go`; the pass stays `auth-data` through `main_test.go`. | T7 |
| m2 | Folded (PS-11) in T0; the settings rule is refused as under C8. | T0 |
| m3 | Folded (PS-11), as under C17. | T6 |
| m4 | Folded, as under C5. | T6 acceptance |
| m5 | Folded, as under C21. | T6 acceptance |
| m6 | Folded: T5's and T7's reports name their worktree and branch, and the score sums `--records` over all three. | Conventions "Records"; Close step 5 |
| m7 | Not a fork (PS-1a), as C7. | T1 |
| m8 | Folded: the session id is `$CLAUDE_CODE_SESSION_ID`; T4 finds CLAUDE.md sections by content; one plain command per write, or the Write tool; `git merge --no-edit`; each pair checks `git merge-base --is-ancestor <dispatch sha> HEAD` first. | T0; T4; Conventions "Records", "Commits", "Pairs" |

### Risk lens

| Id | Disposition | Where the revision carries it |
|---|---|---|
| R1 | Folded (PS-3), with one placement change: the `.gitignore` entry and its comment land in T0, not T4, because `.gitignore` is gate machinery (`check.sh:367`): in T4 it would turn a docs-only half into a full gate and force the pair's post-merge run to the full form, while T0's gate is the full form already. The dry run gains "extract, then delete", and `unattended-window-2026-09-07` and `photo-round-one-closed` take it. The reviewer's "not ignored" premise is partly wrong: the main repo's `info/exclude:18` already ignores the file locally; the repo still records the ruling nowhere, which is what the entry fixes. The five maybes go to an open decision (row I4); the `-gpu host` contradiction goes to row I8. | T0; T4 dry run; T3 rows I4, I8 |
| R2 | Folded (PS-7), convergent with C1 and M6. | T3 row I66 |
| R3 | Folded (PS-7), convergent with C2. The play-queue line is kept in ROADMAP rather than moved to IDEAS, since `own-visual-identity.md` already says the requirement lives on ROADMAP and T2 owns that file alone. | T2 removal map and acceptance |
| R4 | Folded (PS-7) at the root with C3 and C15: the plan carries the inventory with tokens; T3 runs on `opus`; `grep -c 'HISTORY' BACKLOG.md` may not grow; streamer items name their slice; STATUS:58 is retired as T6's. | T3 |
| R5 | Folded (PS-8): STATUS's next action is the ordered post-merge list, starting with the pull, and the PR body repeats it. | T3 STATUS bullet; Close step 5 |
| R6 | Folded (PS-6), convergent with C10 and M5; `test -s` precedes every check that reads a file which could be empty. | Conventions "Records"; T4 acceptance; retired-phrase check |
| R7 | Folded (PS-9): a Resume convention, and the prompt's resume entry. Checking a pair worktree's warm state after a restart is a stop rather than a `git -C` read, since the guard refuses `git -C` into another tree. | Conventions "Resume"; prompt |
| R8 | Folded (PS-7): the "facts not carried" column; the post-merge step backs up with `tar`, re-lists, writes the index before deleting, and keeps any file newer than T4's start. The named facts are carried: the Svelte exemplar paths into CLAUDE.md (T4), the Go acceptance dimension into the adapter's pair line (T6), the never-drop-an-open-decision line into the adapter's close checklist (T6), the `pgrep` self-match into the executor check (T4, T6). | T4; T6; T3 post-merge list |
| R9 | Folded (PS-10), convergent with C5 and M9; the CSS arm allows only one-line `/* */`. | Conventions "Diff checks" |
| R10 | Folded (PS-10), convergent with M10: directive comments are rejected and `*` is never a comment marker. | Conventions "Diff checks" |
| R11 | Folded (PS-11): `claude agents --json` first, then `pgrep -f` with a bracket class (`bluefin-admin.md:111-118`), then `git status`; STATUS's in-flight line says it is visible on the branch only. | T4 Safety rails; T6 adapter; T0 STATUS line |
| R12 | Folded (PS-11), convergent with m8. | Conventions "Pairs" |
| R13 | Folded: the close reruns `git ls-files .claude/agent-memory` after merging `origin/master`. | Close step 1 |
| R14 | Folded (PS-3): the launch check expects empty output, and the stale out-of-scope and post-merge lines are gone. The close's STATUS line reads as open until merge, and the `admin-api-1` planning session replaces it with its own in-flight line. | Launch; Out of scope; Close step 4; T3 prompt |
| R15 | Folded, as C8. | T0 acceptance |
| R16 | Folded lightly: the class stays (pass-core's floor), and the security reviewer's dispatch confirms the comment-only proof and reads the comments for false security claims. T7 no longer touches `internal/serve`, which narrows that review to one test file. | Close step 3 |

## Measures

- Findings: 60 (contract 21 plus 3 over-ceremony; mechanics 1 blocker, 11 major, 8 minor; risk 16).
- Folded in full: 51. Folded in part: 4 (C4, C8, C20, and R1, whose entry moved to T0). Review forks settled as
  tasks: 3 (C7 and m7 as T1, M7 inside T0; PS-1 rules each a question with one correct answer). Refused: 1 (CO2),
  plus the settings allow rule inside C8 and m2. Kept as recommended: 1 (CO3). Owner forks: 0. Design changes: 0.
- New mechanisms introduced by the fold, probed or quoted above: the `--base` convention, the wrapper allowlist and
  dry-run lane, the docs-only gate skip, the comment-only filter, the joined retired-phrase and lane checks, the
  inventory check, end-to-end check 4, and the `check-ignore -v` proof. Two were not probed by this fold (corrected
  after the verification read, finding 10): the git-dir records rest on precedent, not a probe here (two isolated
  cairn sessions, `lean-cutover` and `engine-pass-b`, wrote `pass-task-clock` into their worktree git dir with one
  plain `printf ... > <path>`, the second with a `base=` line, as the verification read reports), and T0's clock write
  is this pass's first live proof, with a refusal there a stop before any work; the resume procedure is unprobed and
  rests on the git commands it names.
- Plan length: 510 lines before, 786 after the first fold, 830 after the verification fold; the inventory table is 70
  of the added lines. Launch prompt: 44 lines before, 58 after the first fold, 62 after the verification fold.
- Clock estimate: 305 task-minutes and about 210 on the path before; 450 task-minutes and about 335 on the path
  after, because every gate-bearing task now holds one recorded gate run and the dependency sweep is added. The
  verification fold leaves it unchanged (bounding T1 removes work rather than adding it).

## Verification fold

The verification read (`docs/superpowers/research/2026-10-11-dubplate-alignment-plan-fold-verification.md`, at
`582d47e`) found no blocker, one major, and ten minors, and no design change or owner fork. Each is folded; none
needed a new verification read by the verifier's own judgment.

| Id | Disposition | Where the revision carries it |
|---|---|---|
| V1 (major) | Folded. T1 is bounded: every take-now item and every held major is a BACKLOG item with a `#rung:` tag, not taken; no tripwire, cron, or cloud routine; no ROADMAP line; nothing outside the worktree; the gate is local and there is no CHANGELOG. `BACKLOG.md` joins T1's files, and the latest record is `2026-10-10-dependency-sweep.md` (held majors at `:120-121`, read). T1 commits before the first pair, so T3 branches from T1's BACKLOG and treats its items as existing; the typescript 7 held major and row I66 are one item. Acceptance adds an empty ROADMAP diff and a `#rung:` on each added line. | T1; prompt's T1 sentence |
| V2 | Folded. Both one-line block arms use `\/\*([^*]\|\*+[^*\/])*\*+\/`, so a line must be exactly one comment. Probed: on the verifier's `p.diff` the new filter flags `/* x */ y := 2 /* z */`, `/* x */ return nil /* y */`, and `/* a */ --color-primary: red; /* b */`, which the `.*` form passed, and it still flags 14 of 14 on the first probe diff. T7's `vite.config.ts` comment is a `//` comment. | Conventions "Diff checks"; T7 |
| V3 | Folded. The T4 and T6 grep commands, the lane check, and end-to-end check 4 sit on single lines in fenced blocks. | End-to-end check 4; Conventions "Lane check"; T4 and T6 acceptance |
| V4 | Folded. The Navidrome lesson's arrival token is `Music Library`, which prints 0 in `stack/stack.go` today (`grep -c -F`); `X-ND-Authorization` is already there twice. | T7 |
| V5 | Folded. Resume rewrites `start=`; reads task commits with `git log --no-merges <launch base>..HEAD ^origin/master` from a `launch-base=` line the session writes first into the PR notes; judges a pair branch by whether it holds its `T<n>:` commit (from the session's own tree, no `git -C`), a branch without one being a stop; and the prompt now opens with the existence check, then EnterWorktree by `path` or `name`. | Conventions "Resume", "Records"; prompt opening |
| V6 | Folded. The `find -newermt` output is read, not required empty: each path goes into the PR notes and is kept by the post-merge re-list, and it is a stop only for a path T4 classed delete. | T4 acceptance |
| V7 | Folded. The fixed floor of 30 is gone; the `#rung:` count must be nonzero, and the inventory check proves coverage. | T3 acceptance |
| V8 | Folded. The lane check reads `\b([[:alnum:]]+[ -])?lanes?\b` and excludes `^(heavy\|light\|fast)[ -]lanes?$`. Probed: it keeps "heavy-lane" and "fast lane", still prints "Document lanes" and "every lane", and prints 10 lines for today's adapter. | Conventions "Lane check" |
| V9 | Folded. T7's `diff-reviewer` bullet says the session runs it after the merge, never the pair. | T7 acceptance |
| V10 | Folded. The Measures line no longer claims probes for the git-dir records or the resume procedure; it cites the cairn precedent the verifier named and calls the resume procedure unprobed. The Records convention cites the same precedent. | Measures; Conventions "Records" |
| V11 | Folded, each part: T0 records the three originals' `stat -c %Y` before the copy into the PR notes and compares after; the admin-api-1 planning prompt cites this plan's archive path; the launch check also matches `.name == "dubplate-alignment"` (probed with a substituted name against live `claude agents --json`); the Pi 5 direction restores "relayed by the conducting session"; T2 writes "up next" in lower case and its acceptance grep is case-sensitive, matching row I70. | T0; T3 prompt bullet; Launch; T2 |

Verification fold measures: 11 findings, 11 folded, 0 refused, 0 owner forks, 0 design changes.
