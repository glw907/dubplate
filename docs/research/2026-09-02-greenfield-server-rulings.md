# Greenfield dubplate-server: the ratified rulings

Brainstorm outcome, 2026-09-02 (Geoff + Claude). This document is the
handoff into the greenfield build's spec-writing session. Every ruling
below is Geoff's, given in session; the spec is written FROM this
document and the inputs listed at the end, then goes to Geoff for
review BEFORE any adversarial review round or plan (process correction
2026-09-02: no review rounds or plans ahead of Geoff's read).

## The rulings

1. **Fresh build.** dubplate-server is built fresh and logically,
   foundation first. It is the universal server of the organizer spec,
   not an evolution of the current code.
2. **Retire the old code.** No old code needs preserving and nothing
   needs to keep running through the build. Salvage is welcome where it
   clears the bar: the uploader is explicitly salvageable (the
   fidelity-passed web components, tus engine, checks, reject copy
   machinery are adoption candidates, admitted on merit under fresh
   review).
3. **Launch bar: everything, Go matcher included.** The server launches
   only when it is the complete organizer-spec server: intake, engine
   with the Go matcher past its conformance bar, store, review/decision
   queue, backup/restore, notifications, admin API, client API, login.
   The family system is not live until then.
4. **Universal, never idiosyncratic.** No machinery specific to one
   estate. The universality tests from the login spec apply to
   everything: no hostname/path/email/name outside config, drivers own
   driver knowledge, tunables are thresholds with defaults, the design
   must make sense with no Cloudflare account.
5. **Exemplary Go.** Truly idiomatic, architecturally beautiful Go
   tools: stdlib idiom as the default vocabulary, exported surfaces
   that read like stdlib APIs, reviewers empowered to fail
   awkward-but-working code. (Memory: idiomatic-beautiful-go.)
6. **Server before TUI, hard boundary.** No TUI or local-binary work
   until the server is fully formed (recorded in the organizer spec).
7. **Login machinery is server purview**; the SSO proxy is a
   streamer-driver capability; magic-link is the verify-identity seam's
   reference implementation with Access demoted to optional hardening.
8. **Extensibility fork (settled earlier the same day):** built-ins +
   lifecycle hooks + the documented API; no plugin ABI.
9. **Uncertain-match custody policy:** strict refusal default,
   owner-queue and contributor-picks as settings, the classical-appeal
   action rides every setting (already in the organizer spec).

## Rulings taken during the build

10. **Rung 9's concurrency model is a per-album lease** (Geoff,
    2026-10-10, at rung 14's planning, as PA:118 schedules it). Not an
    explicit queue. The store already admits jobs through flock leases
    (`internal/engine/store/lease.go:166-168`, one job per kind), and
    a lease per album extends that mechanism rather than adding a
    second one. The album is the verdict unit, so it is the natural
    unit of exclusion. A per-album lease takes no global lock, so an
    interactive session holds nothing a timer import waits on; that
    meets the organizer spec's binding constraint, "an interactive
    local session must not starve timer imports"
    (`specs/2026-09-01-dubplate-organizer-design.md:612-616`). The one
    budget the two still share, MusicBrainz's, is arbitrated by the
    limiter's caller classes (`internal/engine/musicbrainz/limiter.go`),
    whose anti-starvation floor rung 9 re-verifies. An explicit queue
    would have bought a visible order at the cost of a written
    starvation policy. Consequence for rung 14: `/api/v1` publishes no
    queue position, and a start that finds an album's lease held
    answers `conflict` naming the running job, as a second backup does.

## Approach A: two tracks, dependency-merged (ratified)

- **Architecture spine**, foundation first: config core → identity/auth
  (magic-link login built to the hardened requirements below) → ops
  seam + store → intake → pipeline → streamer drivers/SSO → decision
  queue → versioned API.
- **Evidence track**, independent from day one: corpus + harness →
  Go matcher → shadow. The still-running estate is the shadow oracle:
  the new matcher shadows live beets decisions for the whole build,
  which supersedes the organizer spec's 30-day/20-album shadow window
  with something strictly stronger.
- Tracks meet at an integration rung; the launch cutover closes:
  beets DB migration, box swap, retirement of whatever interim estate
  remains.
- Reuse stance (ratified with A): greenfield package tree per the
  ops-seam architecture; proven leaf packages (atomicfile, reject copy
  contract, gate policy, ledger format) are candidates admitted on
  merit; the uploader web components are source material.

## Interim estate state (wind-down executed 2026-09-02)

The box runs Navidrome + cloudflared only, plus the guardian timers
(import, backup, disk-check, navidrome-check) protecting the seeded
library. Retired: FileBrowser (container, compose service, env var,
inbox.907.life DNS + Access app), the uploader's serve unit and nudge
units (disabled; unit files remain deployed as salvage), the uploader
tunnel path rule, both interim Access apps. The tunnel serves
music.907.life → Navidrome only. DUBPLATE_SERVE_SITE_NAME ("907
Music") stays minted in the age store for the new server.
FILEBROWSER_ADMIN_PASSWORD is orphaned in the age store (harmless;
delete at the next registry sweep).

## Inputs the spec must fold

- The organizer spec (2026-09-01-dubplate-organizer-design.md) as
  amended through 2026-09-02: the architecture, success bar, seams,
  custody policy, beauty section. Its ladder is re-derived for
  greenfield (no seam-swap rung; a launch-cutover rung instead).
- The login spec (2026-09-02-magic-link-login-design.md) as SOURCE
  MATERIAL (its build-on-old-serve framing is superseded), plus BOTH
  its adversarial reviews as birth requirements for the auth build:
  - Security review (6 blockers): unconditional identity-header strip
    on ingress for every path including /rest/* and /share/*; trusted
    source pinned to serve's single address (never the bridge /24)
    with ingress-flip-before-ExtAuth ordering; the /api/* collision
    (dubplate owns exactly one prefix); rate limits keyed on the real
    client IP behind a configured trusted proxy; magic links
    GET-render/POST-consume with durable single-use nonces; plus the
    majors (secrets domain separation and rotation overlap, absolute
    session cap and per-email epoch, __Host- cookie stripped before
    proxying, header-spelling strip set, PUBLIC_URL config, SSE/write
    timeout fix, unbounded-map guards, roster validation).
  - Universality review (3 criticals): dubplate's API under its own
    prefix only; the trusted-source wording; roster name vs streamer
    username split (per-contributor streamer_username + collision
    checks) plus identity-provenance in the verify seam, forward-auth
    verify endpoint for adopters with their own proxy, cookie scoping
    for split-host deployments, and the enable-flag discipline.
  Full review texts:
  docs/superpowers/research/2026-09-02-login-spec-reviews/
  (security.md, universality.md).
- Navidrome integration research
  (conversation, 2026-09-02; key facts recorded in the login spec):
  ExtAuth config names, auto-create-on-webapp-path, /rest and /share
  bypass requirements, SPA 401 behavior, SSE via httputil.
- The conformance corpus research
  (2026-09-01-beets-conformance-corpus.md) and the uncertain-match UX
  research (2026-09-01-uncertain-match-ux.md).

## What the fresh session does, in order

1. Read STATUS, this document, the organizer spec, the login spec.
2. Write the greenfield build spec (the initiative map: tracks, rungs
   with magnitude bands, launch bar, auth birth requirements,
   reuse stance, repo tree shape).
3. Present the spec to Geoff for review. Only after his read: the
   adversarial review round, the fold, and rung 1's plan via
   writing-plans.
