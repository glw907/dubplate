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

1. **Approve the 6a plan and rule its four rulings.** Each is recommended yes:
   - a 6.8M ceiling;
   - `doctor` and `check` exit 0 only for a store that is absent, older, or current;
   - serve refuses to create a store beside a snapshot copy;
   - the busy timeout alone bounds the cross-process wait.

   Eleven spec errata land on approval.
2. **Accept 5c-i's flag crossing?** The pass came to about 5.8M against the 5.2M flag. The
   security fold drove it (0 critical, 0 major, five minors folded). The conductor ran the close
   within the ceiling rather than hold it for a sitting.
3. **The principal-collision rule (m1).** Uniqueness now keys on the effective principal. A username
   may equal another member's address only when that member has a username of its own. Should the
   rule be stricter, with every address and username in one shared namespace? The recommendation is
   no, because the current rule already guarantees that no two members share one principal.
4. **Routed findings** (`research/2026-09-23-beets-2-14-and-disc-layout-findings.md`):
   - **beets 2.14 fixes the "ft" pattern.** Bumping the pin moves the oracle mid-shadow, so this is a
     `dependency-upgrade` decision.
   - **Disc numbers in file names.** The default layout omits them, so multi-disc albums interleave
     in one folder. This is an owner fork, and a fix is cheapest before rung 16.

## Next action

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

- **No deploy until 5d.** The env template selects `magic-link` without its nine required keys.
  `newProductionRegistry` in `main.go` is contested.
- **Carried to 5c-ii:**
  - the admin gate (R-5c12) and the roster API (R-5c13);
  - the serve wiring and the `web/` tree;
  - the rejoin sentences (`CLAUDE.md:140`, `scripts/check.sh:5`, `Makefile:33`);
  - refusing invisible code points on the web write path (security n5).
- **Carried to 5d:**
  - the "`doctor` exits 0" row;
  - `ReadWritePaths`;
  - the template's nine keys;
  - the trusted-proxy setting;
  - the runbook row for running `member add` without secrets in argv;
  - the runbook sentences on kind mode's enumeration bound and the mail-ceiling lockout lever.
- **Carried to rung 12:** no principal may map a non-admin member onto a streamer admin account.
  Navidrome v0.63.2 matches usernames case-insensitively.
- **Carried to rung 6:** row 9's `identity.go` half of the errata (`Identity.UserID`), now unblocked.
- **Carried to rung 7:** the uploader accepts `.zip`, `.tar.gz`, and similar archives
  (uploader spec, `d74543d`).
- **Coverage gaps and staged legs.** 86 declared coverage gaps, none blocking. The staged gate legs
  warn today: 325 process citations and 128 exported-unused symbols. Configuration homes routes tasks
  into rungs 6, 7, 9, 12, 14, 16 and the 2.0 list. Fork B ruled 1.0 FLAC-only, and the group non-FLAC
  question is carried.
- **Leftovers:** four-knob wall registration; `AdminEmail`'s contact row; five 3b items; `profile_hashes`;
  `Case.Extra`; `slices.Clone` (rung 9); the corrupted untracked 2026-09-04 evidence doc; two review
  albums on the box; `FILEBROWSER_ADMIN_PASSWORD`; stale failed units; `DUBPLATE_SITE_NAME` re-mint.
- **Pinned (2026-08-30):**
  - Navidrome `0.63.2`
  - cloudflared `2026.8.2`
  - beets `2.13.1`
  - mediafile `0.17.0`
  - unidecode `1.4.0`
  - lap `0.5.13`

  Spend: about 61.6M, plus 9.1M (the 2026-09-12 window), 6.4M (5b-ii), and about 8.3M for the
  2026-09-23 window of the 12M cap.
