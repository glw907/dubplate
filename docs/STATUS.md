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
   (docs/superpowers/specs/2026-09-02-greenfield-server-build.md,
   revision 2.1) owns the 17-rung ladder and supersedes the organizer
   spec's own build order; the organizer spec
   (2026-09-01-dubplate-organizer-design.md) still governs purview, API
   shape, and the ops seam. The login spec
   (2026-09-02-magic-link-login-design.md) is source material; its two
   reviews (research/2026-09-02-login-spec-reviews/) are auth birth
   requirements.
4. Process rule (Geoff, 2026-09-02): specs go to Geoff for review
   BEFORE any adversarial review round or plan.
5. The box is production for Geoff's own listening; quiet-box check
   before writes; the library and beets DB stay protected by the
   timers and R2 backups throughout the build.

## Open items (carried)

- Rewrite post-mortem still owed.
- Two albums in /srv/music/review/2026-08-31/geoff/ await manual review.
- Age store: FILEBROWSER_ADMIN_PASSWORD orphaned (delete at next sweep);
  DUBPLATE_SERVE_SITE_NAME ("907 Music") stays minted for the new server.
- Stale failed units on the box; cosmetic; `systemctl reset-failed`.

## Pinned versions (looked up 2026-08-30)

| Component | Image / package | Tag / version |
|---|---|---|
| Navidrome | `deluan/navidrome` | `0.63.2` |
| cloudflared | `cloudflare/cloudflared` | `2026.8.2` |
| beets (box, via uv) | PyPI `beets` | `2.13.1` |

### Next action

Rung 1 is closed. Write rung 2's plan (corpus and harness) via
writing-plans, against the build spec's ladder entry 2. Settled (do not
re-brainstorm): the rulings doc, approach A, launch bar, universality,
exemplary Go, reuse stance, auth birth requirements, the 17-rung
ladder, all three review-round rulings.
