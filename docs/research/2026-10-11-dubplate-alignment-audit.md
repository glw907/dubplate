# Dubplate alignment audit (2026-10-11)

A read-only audit of dubplate at `8269d27` (master, after the lean cutover's step-3 PR #1) against
the lean pass process. An Opus reviewer ran it for the dotfiles conducting session. Paths are
relative to the repo unless they start with `~`. Abbreviations: BS is
`docs/superpowers/specs/2026-09-02-greenfield-server-build.md`; PA is
`docs/superpowers/specs/2026-09-03-parallel-build-approach.md`; rung7 is
`docs/superpowers/specs/2026-10-10-rung-7-intake-design.md`; the 14a plan is
`docs/superpowers/plans/2026-10-10-rung-14a-admin-api.md`; the adapter is
`.claude/skills/dubplate-pass/SKILL.md`.

**Short answer.** The ladder as product sequencing is still sound and should stay. Almost all of
the process machinery around it predates the lean process and contradicts it. Step 3 removed the
agents and scripts. It did not remove the doctrine, which still lives in CLAUDE.md pointers, the
adapter, STATUS, the frozen specs, and the auto-loaded project memory.

## 1. The build ladder

### (a) Product sequencing: keep

- Rungs ordered by dependency (BS:673-688), with the graph in
  `research/2026-09-03-rung-dependency-analysis.md`. Product order, not process; ruling 2 reaches
  only process steps.
- The launch bar (BS:162) and the serial constraints in PA §2 (PA:132-144): 5→12 ingress before
  ExtAuth, 6→7/9/14 store first, box quiet windows, the MusicBrainz rate budget.
- Phases (BS:1163-1168): reshape into ROADMAP milestones. They stop being session units; "no clears
  inside a phase" and phase-span sessions retire (pass-core: "One session executes.").
- Size bands M/S/L, an L rung splitting into passes at its own planning (BS:683-686): keep as pass
  sizing.
- Token ceiling and checkpoint interval in every plan header (BS:685-686; PA:245): retire.
  pass-core gives each task "a clock estimate in minutes", and the Score table has no token row.

### (b) Process machinery

| Element | Where | Verdict | Lean basis |
|---|---|---|---|
| Lanes, the cap raised to three code lanes | PA:146-150, PA:241; adapter:29-31 | Reshape into concurrent passes, each its own session, worktree, and PR; pairs inside a pass | pass-core Execute: "One session executes." Ruling 10: "independent items run in pairs by default." Geoff ruled up to three concurrent passes, 2026-10-11 |
| Lane registry as the executor check | PA:248; STATUS:26-31; CLAUDE.md:174-176; adapter:18-22 | Retire | Global CLAUDE.md: "verify none is already working it: `pgrep -f <worktree path>`, `git status`"; pass-core: "STATUS records the session's name and id." |
| Contested-surface ownership table (about 50 rows) | PA:183-236; rung7:973-993; 14a plan:355-358 | Retire as a standing table; each design names its files | pass-core step 1: the design names "files and interfaces touched" |
| Merge ritual (merge token, rebase, `--ff-only`, on-master gate, re-review) | PA:166-177; 14a plan:122-124 | Retire; keep only: with two PRs open, the second merges master and reruns the close gate | pass-core: "A pass that continues an existing branch merges..."; "Geoff reads the PR and merges." |
| Conductor and conducting loop (thin conductor, stop list) | PA:240-246; 14a plan:113-147 | Retire | pass-core Execute and Clock stop; ruling 13 |
| Phase post-mortems and per-lane handoffs | BS:1170-1172; PA:247 | Retire | Ruling 8: "HISTORY and per-plan post-mortems stop growing." |
| Checkpoint bundles of batched questions | BS:1157-1158 | Reshape into the design and the PR; product adjudication batches stay | Ruling 4 |
| Workflow mode as a standing grant; per-task implementer to diff-reviewer chain; "workflow mode at six or more tasks"; "both budgets scored" | BS:1147-1149, 1177-1179 | Retire | Global CLAUDE.md: Workflow runs only on Geoff's opt-in; pass-core: diff-reviewer on `auth-data` tasks and at close |
| Just-in-time plans | BS:1173-1177; adapter:32-33 | Keep: plan staleness is the dominant defect class; a fresh read found "five blockers before any dispatch" (adapter:26-28) | A plan per pass, after the design |
| Rung spec and plan weight | rung7 spec 1,275 lines; 14a plan 869 lines; their reviews, folds, and verifications total 5,903 lines | Reshape | pass-core: "a design of a page or less"; "spec ... at the length precision needs"; plan tasks carry "never implementation code" |
| Rung 7 spec "goes to Geoff before any plan is written" | rung7:7; STATUS:21 | Retire | Ruling 4: "He never needs to read a plan or a spec." |
| Errata owed across many documents | rung7:1233-1259 (eight documents); STATUS:33-40; PA:5-7; 35 errata mentions in BS; CLAUDE.md:30-33 | Retire | Ruling 8; "A lesson lands where it executes." |

Spec and plan weight: cut the ceiling and clock sections (rung7:1265-1271; 14a plan:36-69), the
segment and unattended-run rules, sequencing tables that assume lanes, and the "Rulings for Geoff"
blocks. Keep spec precision on `auth-data` surfaces (rung7:997-1090). Fold records become a revision
line in the spec plus the PR body.

Errata rule going forward: frozen, ratified specs (BS, PA, R6, the 14A spec once 14a lands) get no
errata; a superseded spec gets at most a one-line header pointer. Living references read by later
passes (SS forward hooks, PCF copy contract, VI, CH, W) are edited in place, as a task inside the
pass that changes them. There is never an owed list.

## 2. Remaining misalignments, ranked by consequence

1. **The project auto-memory teaches the retired process.** The index
   `~/.claude/projects/-var-home-glw907-Projects-dubplate/memory/MEMORY.md` is 44 lines over 42
   files: retired process (`greenfield-arc-sessions` line 4, `pass-execute-*` 9-10,
   `two-lane-run-mechanics` 18, `ledger-fold-must-keep-list` 40) and eleven session and window logs
   (lines 12, 14-15, 19, 25-26, 28, 30, 33, 39, 42-44). CLAUDE.md:22-24 leans on it ("Two standing
   bars, from memory"), and memory never reaches pair subagents. pass-core: "Memory holds only a
   preference or a ruling's why that no file can." Fix: delete the process and session files; move
   product rulings that live only there into CLAUDE.md (version numbering, group-not-family, plain
   terms, own visual identity, idiomatic Go and Svelte, see-it-before).
2. **The rung 7 gate contradicts ruling 4** (STATUS:21; rung7:7). Fix: a one-page design brought to
   Geoff; the spec stays as agent input; his eight owner rulings are already in it (rung7:29-51).
3. **CLAUDE.md:58-64 points sessions at the retired machinery**, naming BS Process and PA as the
   canonical record (BS:1180-1182 says CLAUDE.md carries "the operational form"). Fix: phase names
   plus "Process: `dubplate-pass`", keeping the product touchpoints.
4. **The lane registry is still the executor check** (CLAUDE.md:174-176; adapter:18-22;
   STATUS:26-31). Fix: the global per-worktree `pgrep` plus `git status`; session name and id in
   STATUS.
5. **The errata-owed pattern is live** (STATUS:33-40: eight documents of rung 7 errata, the rung
   14a spec's eight owed errata, the roster-walk pin). Fix: apply the errata rule; only living-
   reference edits become tasks in the pass that consumes them; close the rest as moot, one line
   each.
6. **STATUS content breaks pass-core.** It is exactly 60 lines, but the next action is stale
   (STATUS:14, "Once the lean-cutover PR merges"), STATUS:42-60 is a backlog of future-rung items,
   and STATUS:4, 11, 56, 57, and 60 point into the frozen HISTORY. Fix: move rung items to ROADMAP
   or BACKLOG, pull cited items out of HISTORY, rewrite the next action.
7. **The adapter names the wrong pair agent** (adapter:53, `site-implementer`, whose definition of
   done runs `npm ci`, `svelte-kit sync`, and `npm run check`; `~/.claude/agents/site-implementer.md:14-27`).
   Fix: `general-purpose` with the adapter's fast lane, setup, and `go-conventions` or the web
   skills named in the prompt.
8. **Agent memory.** `.claude/agent-memory/dubplate-implementer/` belongs to an agent deleted at
   `8269d27`: `MEMORY.md` and `feedback_test-echoes-code.md` are tracked, `MEMORY.md` is modified,
   and four files are untracked (the 14a plan:351-352 flagged them as warm). Dispositions:
   test-echoes-code becomes one clause at CLAUDE.md:86-88; gate-vanished is deleted (pass-core
   covers exit 75); the photo and network notes are deleted (the round is closed); then the
   directory goes. `.claude/agent-memory/site-implementer/` holds 11 tracked lessons
   (`memory: project`): the DaisyUI lessons belong in the visual identity spec or the `web/` review
   dispatch, and `commit-only-named-paths` is already pass-core and global git policy.
9. **`docs/standing-rules.md` is partly a side doc.** Its first three sections are reached
   (CLAUDE.md:38, 44, 76; adapter:14, 49). Its "Lessons the code cannot carry" (113-170) is reached
   by nothing. Fix: move each lesson home: package docs (harness 119-136), mux and `Handle.Update`
   doc comments (141-148), a `svelte.config.js` comment (150-153), a cobra test helper (163-164), a
   `deploy.sh` or `provision.sh` comment (168-170). Trim the duplicated 0.8 rule (CLAUDE.md, the
   adapter, standing-rules, ROADMAP:100-103, memory).
10. **HISTORY is frozen but still referenced** (12,041 lines; frozen header at line 1; last write
    `28e0963`). CLAUDE.md:50 still calls it the "per-pass ledger, read on demand". Fix: mark it a
    frozen archive.
11. **A post-mortem is still owed on paper** (BACKLOG:9-11, #1; ROADMAP:215, 306). Fix: close #1
    under ruling 8.
12. **Stale CLAUDE.md lines:** 30-33 (errata-avoidance prose; one line is enough) and 53-56 (the
    "migrated ... at T14" note).
13. **ROADMAP:14-114 is long narrative**, the Active section carries landed status (39-47), and
    there is a Done section (292-306) beyond Active, Planned, and Someday. Fix: trim Active entries
    to goal and next rung.
14. **Old wording:** `scripts/check.sh:31, 41, 253` says "boundary gate". `.claude/settings.json`
    allows retired docker compose and box `systemctl restart` commands (Geoff ruled them removed,
    2026-10-11).
15. **Process citations in shipped Go comments** (about 40 "ruled, conductor" citations, for example
    `internal/config/registry.go:35`, `internal/auth/epoch.go:65`). The staged leg (`check.sh:781`;
    standing-rules:107-111) only warns. An existing chore; it stays queued.

## 3. Dubplate-specific items to keep

- Product touchpoints: corpus adjudication batches, the rung 7 uploader sitting, the first-launch
  sitting (BS:1150-1158; rung7 owner ruling 4).
- Mutation proofs are run, never narrated (adapter:23-25): "two of six such mutations ... did not
  fire."
- Planning-miss prevention's four gaps (adapter:34-38) and the read against shipped code
  (adapter:26-28).
- Just-in-time plans (adapter:32-33).
- Production box rules: the quiet-box check and read-only ssh (CLAUDE.md:176-177).
- Main-tree-only work for the venv, `make derive`, and `corpus/` (adapter:71-77), since a
  worktree's gate goes green without the beets leg (PA:251).
- The gate lock (`check.sh:194-209`): `tidy-check` mutates `go.mod`.
- The `--merge` history gitleaks scan (adapter:92-95) and the oracle-tagged `go vet`
  (adapter:66-69).
- The per-pass dependency sweep (adapter:45-49), Geoff's 2026-10-04 ruling.
- The 0.8 forward-hook and prior-art rules (CLAUDE.md:35-38).
- Refused features, facts agents get wrong, and the visual-identity review dispatch
  (CLAUDE.md:138-170; adapter:50-52).

## 4. Not checked

The contents of the untracked `reports/` and `research_notes/` directories, and of two untracked
research files dated 2026-09-04.
