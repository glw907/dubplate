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
Pass 2b is CLOSED (HISTORY has the ledger; commits cec3c12 →
2219b19): all machinery, 61 structural cases, 68 total, everything
re-deriving byte-exact; ≈5.8M of 6M spent, the 12M grant untouched.
Next per the overnight grants: plan pass 2c just-in-time on the
quiet tree (real-catalog stratum ~80-150 live-captured cases,
adversarial stratum seeded from known-oracle-bug classes,
stock-config stratum per ruling 3, #6358 alias verification,
file-level applied-tag derivation for rung 8, version-bump
differential doc, re-capture chore doc, oracle-bug ledger and
adjudication formats), then execute: overnight captures at anonymous
1 req/s, strata authoring, the FIRST adjudication batch prepared and
waiting for Geoff (never self-adjudicated).
Overnight grants (Geoff, 2026-09-02, ~9h away): 2b ceiling raised to
12M if needed; after 2b closes, proceed into 2c fully - plan
just-in-time, run the live real-catalog captures (anonymous 1 req/s),
author strata, and leave the first adjudication batch PREPARED and
WAITING for his morning (never self-adjudicate). Extension: continue
until a substantive input-gate; at such a gate, pivot to other tasks
that advance the build without input rather than idling.
Spend: 2a ≈1.75M/5M; 2b ≈3.5M (ceiling now 12M).
