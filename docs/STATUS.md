# dubplate status

## Current state

The greenfield dubplate-server build is the one active initiative
(Geoff's rulings, 2026-09-02: docs/superpowers/research/
2026-09-02-greenfield-server-rulings.md). The family system is NOT
live and will not be until the new server launches complete, Go
matcher included. The box (5.78.148.42) runs Navidrome + cloudflared +
the four guardian timers only; FileBrowser, the uploader serve, its
Access apps, and the inbox hostname are all retired (2026-09-02).
Existing `internal/` + `web/` code is salvage tier: adoption material
for the fresh build, retired at launch. Remote:
github.com/glw907/dubplate (private).

## Rules for a new session in this repo

1. Read the rulings doc before anything: it carries the ratified
   rulings (fresh build, universal, exemplary Go, launch bar,
   server-before-TUI), approach A (spine + evidence tracks), the
   reuse stance, and the input list for the spec.
2. Governing architecture: the organizer spec
   (docs/superpowers/specs/2026-09-01-dubplate-organizer-design.md,
   amended through 2026-09-02). The login spec
   (2026-09-02-magic-link-login-design.md) is source material; its two
   reviews (research/2026-09-02-login-spec-reviews/) are auth birth
   requirements.
3. Process rule (Geoff, 2026-09-02): specs go to Geoff for review
   BEFORE any adversarial review round or plan.
4. The box is production for Geoff's own listening; quiet-box check
   before writes; the library and beets DB stay protected by the
   timers and R2 backups throughout the build.

## Open items (carried)

- Organizer spec awaits Geoff's ratification read (committed 80f0f5f
  plus amendments; he may fold it into the greenfield spec read).
- Rewrite post-mortem still owed (prod-polish residue).
- Two albums in /srv/music/review/2026-08-31/geoff/ await manual review.
- FILEBROWSER_ADMIN_PASSWORD orphaned in the age store; delete at the
  next registry sweep. DUBPLATE_SERVE_SITE_NAME ("907 Music") stays
  minted for the new server.
- Stale musicbox-ping-failure@ failed units on the box; cosmetic;
  `systemctl reset-failed` clears.

## Pinned versions (looked up 2026-08-30)

| Component | Image / package | Tag / version |
|---|---|---|
| Navidrome | `deluan/navidrome` | `0.63.2` |
| cloudflared | `cloudflare/cloudflared` | `2026.8.2` |
| beets (box, via uv) | PyPI `beets` | `2.13.1` |

### Next starter prompt (greenfield spec session)

> **Goal.** Write the greenfield dubplate-server build spec and take it
> through Geoff's review to rung 1's plan.
>
> **Settled (do not re-brainstorm):** everything in the rulings doc --
> approach A, launch bar, universality, exemplary Go, reuse stance,
> auth birth requirements.
>
> **Approach.** Read the rulings doc, the organizer spec, and the login
> spec + reviews. Write the spec. Present to Geoff BEFORE any review
> round. Then adversarial reviews, fold, and writing-plans for rung 1.
