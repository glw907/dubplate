# dubplate status

## Current state

The greenfield dubplate-server build is the one active initiative, past
its pivot: rung 1 (foundation) landed 2026-09-02, freezing the pre-pivot
`cmd/`/`internal/`/`web/`/`e2e/` trees under `_salvage/` (tag
`pre-greenfield`, read-only) and standing up `cmd/dubplate-server` with
`internal/config`'s owned wall, `check`, and `doctor`. The family system
is NOT live and will not be until the new server launches complete, Go
matcher included. The box (5.78.148.42) runs Navidrome + cloudflared +
the four guardian timers only; FileBrowser, the uploader serve, its
Access apps, and the inbox hostname are all retired (2026-09-02).
Remote: github.com/glw907/dubplate (private).

## Rules for a new session in this repo

1. The rulings doc (research/2026-09-02-greenfield-server-rulings.md)
   carries the ratified rulings and approach A; read it first.
2. Process: the build runs in session-scale arcs per CLAUDE.md's
   "Build process" section: one session per arc, no clears inside
   one, plans just-in-time, workflow mode a standing grant.
3. Governing architecture: the ratified build spec
   (specs/2026-09-02-greenfield-server-build.md, revision 2.1) owns
   the 17-rung ladder; the organizer spec still governs purview, API
   shape, and the ops seam. The login spec is source material; its
   two reviews (research/2026-09-02-login-spec-reviews/) are auth
   birth requirements, and the build spec wins every conflict.
4. Process rule (Geoff, 2026-09-02): specs go to Geoff for review
   BEFORE any adversarial review round or plan.
5. The box is production for Geoff's own listening; quiet-box check
   before writes; the library and beets DB stay protected by the
   timers and R2 backups throughout the build.

## Open items (carried)

- Rewrite post-mortem still owed.
- Two albums in /srv/music/review/2026-08-31/geoff/ await manual review.
- Age store: FILEBROWSER_ADMIN_PASSWORD orphaned (delete at next
  sweep). Box: stale failed units, cosmetic; `systemctl reset-failed`.

## Pinned versions (looked up 2026-08-30)

| Component | Image / package | Tag / version |
|---|---|---|
| Navidrome | `deluan/navidrome` | `0.63.2` |
| cloudflared | `cloudflare/cloudflared` | `2026.8.2` |
| beets (box, via uv) | PyPI `beets` | `2.13.1` |

### Next action (evidence-arc session)

Pass 2b is executing from its plan (revision 2, committed 415c4d9;
14 tasks, 61 synthetic cases, no owner touchpoints). Conductor seam
ruling: machinery tasks 1-3 first, then a staleness check of case
tasks 4-13 against the shipped machinery before dispatching them.
Case sweep: families 1-7 all accepted (45 cases committed; latest
6a335af). Family 2's "combined" exclusivity and family 3's misflag
request-log witness: ruled and recorded in the corrections doc
(research/2026-09-02-rung-2b-staleness-corrections.md, which wins
over the plan on its named points and carries all dispatch
corrections). Next: batch C = plan tasks 10-13 (families 8-15,
16 cases; task 13 uses the pre-task's zero-candidate machinery),
then Task 14 closes the pass (inventory read off the tree,
determinism run with wall-time breakdown). After close: post-mortem,
2c plan, overnight captures per the recorded grants.
Overnight grants (Geoff, 2026-09-02, ~9h away): 2b ceiling raised to
12M if needed; after 2b closes, proceed into 2c fully - plan
just-in-time, run the live real-catalog captures (anonymous 1 req/s),
author strata, and leave the first adjudication batch PREPARED and
WAITING for his morning (never self-adjudicate). Extension: continue
until a substantive input-gate; at such a gate, pivot to other tasks
that advance the build without input rather than idling.
Spend: 2a ≈1.75M/5M; 2b ≈3.5M (ceiling now 12M).
