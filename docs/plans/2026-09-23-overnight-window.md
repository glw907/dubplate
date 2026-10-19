# Overnight window, 2026-09-23 to 2026-09-24

This window executes pass 5c-i and runs two document lanes beside it. It
launches from a fresh Opus 5.5 session at effort `medium` and ends by
**05:00 AKDT on 2026-09-24**, when Geoff is up. Geoff ruled its scope,
cap, and end time on 2026-09-23.

## Scope

- **Lane 1, code, the main tree.** Execute pass 5c-i from
  `plans/2026-09-23-rung-5c-pages-member-verb.md` (revision 2.1, approved
  2026-09-23, all six rulings yes). Task 0 first, then segments A (Tasks
  1-3), B (Tasks 4-7, then the simplifier, the mail-copy register gate,
  and the pass-end `web-auth-security-reviewer`), and C (the close). The
  execution ceiling is 6.5M with the flag at 5.2M, against a 4.72M
  projection.
- **Document lane 1.** Apply the thirteen errata the ratified rung 6 spec
  owes (`specs/2026-09-23-rung-6-ops-seam-store.md`, its errata section)
  to the documents they name. It never touches a file lane 1 writes.
- **Document lane 2.** Draft rung 6's pass 6a plan from the ratified rung
  6 spec, then review it with the `spec-plan-review` skill (three lenses,
  one fold, one verification read, the prose review last). It stops at
  the approval gate: Geoff approves 6a in the morning.

Out of scope: the 5c-ii plan (it waits for 5c-i's merge), any deploy (the
no-deploy carry holds until 5d), any write to the box, and the
architecture cleanup pass (after 5c-i, in its own session).

## Budget and stop rules

- **Cap 12M for the window**, across all lanes and the conductor.
- **At 80% (9.6M)**, finish the task in flight, write STATUS, and stop
  starting new work.
- **At 05:00 AKDT**, stop at the next clean boundary: a committed task on
  a green gate. Write STATUS and one combined morning question.
- **A stopped lane never stops the others.** `pass-execute` now halts a
  segment at any task whose final status is not accepted
  (`stopOnEscalate`, `~/.dotfiles` `ce70b5f`). The conductor rules a
  second `fix` or an escalation itself when the plan and the rulings
  settle it, and otherwise parks the lane, records the question, and
  keeps the document lanes running.

## Standing grants for this window

- Workflow mode for execution and plan prep, per CLAUDE.md.
- Read-only ssh to the box during reviews; no writes, no deploy.
- The conductor may apply its own rulings where a plan names a
  conductor-owed decision, recording each in STATUS.
- One code lane only (lane 1). Document lanes commit only their own
  files; only the conductor commits ledger files, path-limited.

## Guards

- The runaway guard (`claude-wf-guard <transcript-dir> <tier> <run-id>`)
  on every workflow, `implementer` tier for code and `writer` for docs.
- The sleep inhibitor is tool-held by the activity lease; verify with
  `systemd-inhibit --list` at launch. The laptop stays on AC.
- A heartbeat check at least every 30 minutes (dynamic `/loop` or
  ScheduleWakeup) re-arms any guard the harness culled.

## Morning report

STATUS carries, present tense: 5c-i's task ledger with each verdict and
commit, spend against the ceiling and the window cap, the security
read's verdict, the errata applied, the 6a plan's review state and its
rulings for Geoff, and one combined question if any lane parked.
