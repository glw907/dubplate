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
Accepted: 2b-1 (cec3c12; echo, freeze verb, schema v2 re-freeze) and
2b-2 (6090b6c + 182a5b4, one fix round; synth builder, authoring
loop, weights validation). Next: 2b-3 on opus (tie detector,
proximity flag, claims checker, component tool; plan sanctions a
3a/3b split if it grows). Then the batched simplifier (explicitly:
consolidate maximalReleaseInclude and canonicalizeQuery duplicated
across synth/capture), then the staleness check of case tasks 4-13.
Spend: 2a closed ≈1.75M/5M; 2b ≈1.35M of 6M.
