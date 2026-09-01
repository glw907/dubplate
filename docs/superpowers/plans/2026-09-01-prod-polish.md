# Pass: production cleanup and polish (prod-polish)

**STATUS: PLANNED 2026-09-01, ahead of need (Geoff: "We could plan that pass
now"). Executes after the uploader initiative closes (Pass B through U13).
Geoff's plan approval doubles as the box-write sanction for P4/P5's named
steps.**

**Goal:** Clear the carried production debris, settle the owed rewrite
post-mortem, and create the family's Navidrome accounts, so the estate is
clean before the organizer (beets replacement) work starts.

**Engine consultation:** no engine asks (dubplate builds no cairn-cms
surface; enumeration run 2026-09-01).

**Scope:** BACKLOG items #1–#8 (`#project:prod-polish` plus the #8 watch
item) and whatever `/log-issue` has collected by pass start. Out of scope:
new features, admin API surface (the organizer initiative owns it), the
beets replacement, packaging. #9 (seed-album review) is Geoff's own item and
rides along untracked.

**Budget:** ceiling 1M tokens, checkpoint every 4 tasks. Below six tasks:
dispatch each chain with the Agent tool (no workflow). Implementer
`general-purpose` pinned `model: sonnet`; reviewer `diff-reviewer` (opus).
Repo gate: `bash scripts/check.sh` (P3 uses the pings repo's own gate).
Unit/script tasks invoke `vps-conventions`; Go touches invoke
`go-conventions`.

---

## P0: Pass-start triage (conductor, inline)

- [ ] Sweep `BACKLOG.md` for items logged since this plan was written; fold
      small fixes into the matching task's dispatch notes, split anything
      larger than a small fix to its own pass. State what was folded and
      what was split.
- [ ] #8: `npm view tus-js-client dist-tags` — record the 5.x verdict (stay
      pinned or schedule the upgrade) in the task ledger and close #8.
- [ ] Confirm no Pass B carry-forwards landed in STATUS that belong here.

## P1: Docs truth sweep (#3)

**Files:** `docs/STATUS.md`, `docs/bringup-runbook.md`.

- [ ] Fix the pull-vs-push wording: musicbox-navidrome/dubplate-navidrome is
      a push check driven by the box's navidrome-check timer; sweep both
      docs for the stale "pull" framing (`grep -ni pull docs/`).
- [ ] Gate, commit.

## P2: provision.sh diagnostic fix (#5)

**Files:** `scripts/provision.sh` (or its current home; locate first).

- [ ] `ssh_cidrs_json` malformed input currently exits only the substitution
      subshell, degrading the diagnostic to a bare jq error. Make malformed
      input fail the script with a named, actionable message (match the
      repo's existing error style; the fix pattern is the explicit
      `[ -f ... ]`-style check HISTORY records for the same defect class).
- [ ] Gate, commit.

## P3: pings remove-check --force (#4, cross-repo)

**Repo:** `~/Projects/pings` (its own gate; this task never touches
dubplate).

- [ ] `remove-check` gains `--force`: deletes the check record even when the
      route record is permanently missing, with a confirmation naming what
      is skipped. Test covers the orphaned-record path.
- [ ] Clear the currently undeletable orphaned check with it (a KV write;
      name the check slug in the report).
- [ ] pings gate, commit, push.

## P4: Family Navidrome accounts (#7, box task, Geoff-sanctioned)

**Files:** `scripts/create-family-accounts.sh`, `docs/bringup-runbook.md`
(one section).

**Decision (Geoff, 2026-09-01):** automated from the roster. One account per
roster contributor via Navidrome's admin API, generated passwords written to
a root-only file on the box; Geoff shares credentials through his own
channel; family members change passwords after first login (Navidrome does
not force rotation; the handoff note says so).

- [ ] Script: reads the roster and admin credentials from the box config/env
      (never hardcoded), creates missing accounts idempotently (safe to
      re-run; existing accounts untouched), writes
      `/root/dubplate-family-accounts.txt` mode 0600 with username +
      generated password per new account, and prints a summary naming the
      file. No secrets in argv, logs, or the repo.
- [ ] Runbook section: when to run it, where the file lands, "share, then
      delete the file" as the documented final step.
- [ ] Box step (quiet box): run it once; verify each family member can
      authenticate (one API-login probe per account, then delete nothing —
      the file stays for Geoff).
- [ ] Gate, commit. `web-auth-security-reviewer` reads this task's diff at
      the pass-end fan-out.

## P5: Box tidy and verdicts (#2, #6, box task, Geoff-sanctioned)

- [ ] #2: `systemctl reset-failed` on the box; confirm the two stale
      `musicbox-ping-failure@` units clear and nothing else was in failed
      state. Remove the carried item from STATUS.
- [ ] #6: read-only SELinux check on the state/library paths (`getenforce`,
      `ls -Z` on the dirs the units write); record the verdict (fine as-is,
      or the one labeling fix needed) and close the carried item with a
      stated reason either way.
- [ ] Update STATUS/BACKLOG, commit.

## P6: Go-rewrite post-mortem (#1, conductor-written, docs-only)

- [ ] Write the rewrite post-mortem as the pass's HISTORY entry plus a
      scorecard: spend against the approved ceiling (≈5.0M of 5.5M per
      STATUS), interaction points, what the gates caught, what a later
      initiative (the organizer, which builds on this binary) would be
      wrong to rediscover. Sources: docs/HISTORY.md entries, the archived
      rewrite plan, the model-economy ledger.
- [ ] Close #1, commit.

---

## Pass-end checklist

- [ ] `code-simplifier:code-simplifier` over changed code (P2/P3/P4)
- [ ] Quality gate: `bash scripts/check.sh` green (pings gate for P3)
- [ ] Review gate (proportionate): `diff-reviewer` already ran per task;
      fan out `web-auth-security-reviewer` over P4 only. No UI work, so no
      svelte/a11y/fidelity chain unless triage folded UI fixes in.
- [ ] Ledgers: STATUS (≤60 lines, next starter prompt = organizer
      brainstorm), HISTORY entry, BACKLOG items closed, ROADMAP:
      prod-polish → Done.
- [ ] Archive this plan to `docs/superpowers/archive/plans/`.
- [ ] Commit, push, score both budgets, prep the context clear.
