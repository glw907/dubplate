# dubplate status

## Current state

Rungs 1b, 3, 4, 8, and rung 5's 5a, 5b-i, 5b-ii, and **5c-i** are closed. 5c-i closed on master in
the 2026-09-23 overnight window (HISTORY, top entry). Its seven tasks were accepted, and all eight
exit-gate assertions pass. The security read found no critical and no major finding, and its minors
are folded. The spend is about 5.8M against the 6.5M ceiling, which crossed the 5.2M flag. There were
zero execution sittings. The window's document lanes applied the rung 6 errata (`b406f93`) and
brought the **6a plan to revision 2.2** (`plans/2026-09-23-rung-6a-store-foundation.md`), which
awaits approval. R4's exclusion denominator is still unread, and owed before rung 15.

## Morning question for Geoff (one combined question)

1. **Approve the 6a plan and rule its four rulings** (each recommended yes): the 6.8M ceiling;
   `doctor`/`check` exit 0 only when the store is absent, older, or current; serve refuses a create
   beside a snapshot copy; the busy timeout alone bounds the cross-process wait. Eleven errata follow.
2. **Accept 5c-i's flag crossing?** The pass came to about 5.8M against the 5.2M flag. The
   security fold drove it (0 critical, 0 major, five minors folded). The conductor ran the close
   within the ceiling rather than hold it for a sitting.
3. **The principal-collision rule (m1).** Uniqueness now keys on the effective principal. A username
   may equal another member's address only when that member has a username of its own. Should the
   rule be stricter, with every address and username in one shared namespace? The recommendation is
   no, because the current rule already guarantees that no two members share one principal.
4. **Routed findings** (`research/2026-09-23-beets-2-14-and-disc-layout-findings.md`):
   beets 2.14 fixes the "ft" pattern, but a pin bump moves the oracle mid-shadow (a
   `dependency-upgrade` decision). The default layout omits disc numbers, so multi-disc albums
   interleave in one folder: an owner fork, cheapest to fix before rung 16.

## Next action

0. The nine rung 5 spec corrections are in flight (HISTORY 5c-i entry, "The spec corrections this
   pass owes"). They apply before 5c-ii's plan is drafted.
1. On Geoff's answers: open lane 2 on 6a in a worktree, and draft 5c-ii's plan for lane 1.
   5c-ii opens with the admin gate and the roster API.
2. The architecture cleanup pass joins the staged comment-cleanup task
   (`research/2026-09-24-rung-5c-i-architecture-reads.md`, plus the 5b-ii reads).

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (`/var/home/glw907/Projects/dubplate`, master) | CLOSED at 5c-i's close; reopens on 5c-ii once its plan is approved |
| 2 | worktree removed, branch deleted | CLOSED; reopens on 6a once Geoff approves it |

## Open items (carried)

- No deploy until 5d: the env template selects `magic-link` without its nine keys; `main.go`'s
  `newProductionRegistry` is contested.
- To 5c-ii: the admin gate (R-5c12), the roster API (R-5c13), serve wiring, `web/`, the rejoin
  sentences (`CLAUDE.md:140`, `scripts/check.sh:5`, `Makefile:33`), security n5 on the web write path.
- To 5d: the "`doctor` exits 0" row, `ReadWritePaths`, the nine keys, the trusted proxy, and the
  runbook rows (`member add` without secrets in argv, kind mode's enumeration bound, the mail-ceiling lever).
- To rung 12: no principal maps a non-admin onto a streamer admin (Navidrome 0.63.2 matches names
  case-insensitively). To rung 6: errata row 9's `identity.go` half. To rung 7: archive uploads (`d74543d`).
- 86 declared coverage gaps, none blocking. Staged legs warn: 325 process citations, 128
  exported-unused. Fork B: 1.0 is FLAC-only; the group non-FLAC question is carried.
  Configuration homes routes tasks into rungs 6, 7, 9, 12, 14, 16 and the 2.0 list.
- Leftovers: four-knob wall registration; `AdminEmail`'s contact row; five 3b items; `profile_hashes`;
  response-status extension; `Case.Extra`; `slices.Clone` (rung 9); the corrupted untracked 2026-09-04 evidence doc; two review albums on
  the box; `FILEBROWSER_ADMIN_PASSWORD`; stale failed units; `DUBPLATE_SITE_NAME` re-mint.
- Pinned (2026-08-30): Navidrome `0.63.2`, cloudflared `2026.8.2`, beets `2.13.1`, mediafile `0.17.0`,
  unidecode `1.4.0`, lap `0.5.13`. Spend: about 61.6M plus 9.1M (2026-09-12 window),
  6.4M (5b-ii), and about 8.3M (2026-09-23 window, 12M cap; conductor's sum of subagent tokens).
  5b-i's overage still sits on R23.
