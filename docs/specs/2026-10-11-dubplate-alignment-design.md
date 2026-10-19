# Dubplate alignment with the lean pass process (design)

**Status:** approved by Geoff, 2026-10-11, in the dotfiles conducting session (`c747c352`), with his
rulings below. It runs before the rung 14a re-plan (lean pass process ruling 12). Evidence: the
read-only audit `docs/research/2026-10-11-dubplate-alignment-audit.md`, which cites
file and line for every item. The lean process: `~/.dotfiles/docs/specs/2026-10-10-lean-pass-process-design.md`
and `~/.claude/skills/pass-core/SKILL.md`.

## Outcome

A fresh dubplate session loads only the lean process: no lanes, lane registry, conductor rules,
errata ledgers, phase post-mortems, or workflow-mode grants, from CLAUDE.md, the adapter, STATUS, or
the project's auto-memory. The ladder keeps its product sequencing and gains logical names. The
rung 14a re-plan can then start on aligned ground.

## Geoff's rulings (2026-10-11)

1. **Concurrency.** Up to three dubplate passes may run at once, each its own session, worktree,
   and PR. This restates his 2026-10-05 "up to three code lanes" as passes.
2. **Rung 7** (intake) runs strictly after rung 14a.
3. **PR test CI** lands after rung 14a. The local full gate stays the full-suite home until then.
4. **Permissions.** Remove the retired docker compose and box `systemctl restart` allow rules from
   `.claude/settings.json`.
5. **Ladder naming.** Names become identities, and numbering means order only. Each rung keeps a
   stable subject name (`foundation`, `corpus`, `matcher`, `shadow`, `auth`, `store`, `intake`,
   `tag-io`, `pipeline`, `migration`, `builtins`, `streamer`, `queues`, `admin-api`,
   `integration`, `cutover`, `closeout`). The work inside a rung is numbered passes: `admin-api-1`
   replaces 14a, and `streamer-1` to `streamer-3` replace slices 1, 2a, and 2b. The ROADMAP carries
   one numbered list, "Path to 0.8", in run order. It is the only place a number means order, and
   it is rewritten when the order changes. The run order is derived from
   `docs/research/2026-09-03-rung-dependency-analysis.md` and the current state, with
   rung 7 after 14a per ruling 2. Frozen specs keep their old numbers; one old-to-new mapping table
   in the ROADMAP bridges them, and no errata sweep follows.

## What changes

These are the audit's tasks, adjusted by the rulings. Minutes are rough.

- **STATUS** rewritten to pass-core: current state, next action (the `admin-api-1` re-plan, with
  its launch line), and open decisions. The lane registry table and the Owed list go, open items
  move to ROADMAP or BACKLOG, and pointers into the frozen HISTORY go. About 25.
- **CLAUDE.md:** "Build process" becomes the phase names plus a pointer to `dubplate-pass`. The
  ledgers line marks HISTORY a frozen archive. The two standing bars are stated outright, no longer
  "from memory". The executor check becomes the global per-worktree `pgrep` plus `git status`.
  Stale lines go (the errata-avoidance prose, the T14 migration note). The new rung names are used
  throughout. About 25.
- **Adapter** (`.claude/skills/dubplate-pass/SKILL.md`): the registry goes; lanes become up to
  three concurrent passes (ruling 1); the product touchpoints are restated; the pair agent becomes
  `general-purpose`, given the adapter's setup, fast lane, and `go-conventions` or the web skills,
  in place of `site-implementer`. The errata rule: frozen specs get no errata, and living
  references are edited in place by the pass that changes them, never through an owed list. About 20.
- **Frozen specs:** a one-line superseded header on the build spec's Process section and on the
  parallel-build spec. A disposition line for each STATUS errata item, with the living-reference
  edits moved into the rung 7 and `admin-api-1` designs. About 15.
- **Lessons to their executing homes:** `docs/standing-rules.md` "Lessons the code cannot carry"
  moves to package docs, doc comments, and script comments, and duplicates of the 0.8 rule are
  trimmed. About 30.
- **Memory:** the deleted `dubplate-implementer` agent's tracked memory is removed (its one live
  lesson becomes a CLAUDE.md clause), and the tracked `site-implementer` memory is triaged. The
  project auto-memory under `~/.claude/projects/-var-home-glw907-Projects-dubplate/memory/` is
  pruned of retired-process and session-log files, with a dry run shown first. Product rulings
  that live only there move into CLAUDE.md. BACKLOG #1 (the owed post-mortem) closes under lean
  ruling 8. About 35.
- **Rung 7's gate:** "goes to Geoff before any plan is written" is replaced by the lean step. A
  one-page design is brought to Geoff when rung 7's turn comes. The spec stays as agent input.
  About 5.
- **Permissions** per ruling 4. About 5.
- **Ladder naming** per ruling 5: the "Path to 0.8" list and the mapping table in ROADMAP. ROADMAP's
  Active entries are trimmed to goal and next step, and its Done section goes. About 30.

## Out of scope

Rung work itself (`admin-api-1` is the next pass). PR test CI (ruling 3). The roughly 40 process
citations in shipped Go comments, which stay a queued chore. Untracked files in the main checkout
(`reports/`, `research_notes/`, two 2026-09-04 research notes, and four untracked agent-memory
files) are the conductor's to settle with Geoff; the pass session never touches the main checkout.

## End-to-end check

The full gate (`bash scripts/check.sh`) passes on the close head. A grep over CLAUDE.md, the
adapter, STATUS, ROADMAP, and the auto-memory index finds none of "lane registry",
"pass-execute", "conducting loop", "conductor ruling", "errata owed", "workflow mode", or
"post-mortem" as live instruction. STATUS is 60 lines or fewer. "Path to 0.8" lists every
remaining rung by name, in an order consistent with the dependency analysis.

## Risk

`runner` (it edits `.claude/`), so plan review runs all three lenses. There is no PR CI, so the
local full gate is the full-suite home. About 3 hours of clock.
