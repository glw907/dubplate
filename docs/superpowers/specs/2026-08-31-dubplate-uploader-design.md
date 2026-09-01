# dubplate uploader: design

**Status: revised after three adversarial reviews (product, technical, stack),
approved by Geoff 2026-08-31 (framework decision included: standalone
Svelte 5).** Brainstormed with Geoff 2026-08-31; supersedes
`2026-08-30-upload-check-design.md`. Implements as its own pass after the Go
rewrite; the plan and its review happen at that pass's kickoff.

## Product frame

The estate's product essence: a group of family or friends building a
well-organized library for a shared Navidrome server. Contribution is the
group's verb; this page is the one user-facing surface the estate owns and
members judge the system by it. The bar is well-polished. The adopter goal
binds: site specifics enter through config only.

## Decisions

Carried from 2026-08-30: FileBrowser retires; Cloudflare Access is the only
login; validation happens at upload time with the emails' friendly wording;
uploads are resumable: tus via `tus-js-client` pinned at 4.3.x (the only
serious client; stable since 2025-01 with a stalled 5.0 pre-release — a
watch item, not a blocker). The transport seam is thinner than first
assumed: 4.3.0 already speaks the IETF resumable-upload draft behind a
`protocol` option, so migrating off classic tus is a config flag on the
same client — do not over-abstract the seam.

Ratified 2026-08-31: bespoke UI over a tus core; the contributor sees their
album's journey; the uploader is `dubplate serve` in the one binary; uploads
nudge the import; no group feed, own uploads only; **the page carries a
collection picker** (Geoff, 2026-08-31) — Navidrome ≥0.58 multi-library
maps collections to library-root subtrees, the session metadata carries the
choice, the importer routes via a beets flexible attribute, and the picker
is hidden entirely when only one collection is configured, so
single-collection estates never see it. Stack: Svelte + DaisyUI
(see Development shape), built statically and embedded via `go:embed`; Node
is a workstation build-time dependency only.

Review-forced (2026-08-31): **the album is the verdict unit** — one failing
track rejects the batch in-page before anything lands in the inbox,
preserving the pipeline's deliberate all-or-nothing album rule; **a session
ID minted at upload creation is the journey's spine** (see Architecture);
**a small append-only contributions ledger exists** (JSON or SQLite in the
state dir, keyed by verified email) because run reports plus filesystem
cannot show a contributor their successes after beets renames them — this is
history for the uploader themself, not a group feed.

## The experience

One page. Desktop gets drag-and-drop of a ZIP or an album folder. **iPad
gets ZIP or multi-file selection, presented as a first-class route, never a
degraded one** — iOS Safari has no folder selection at any version, so the
page detects the platform and words the path accordingly; album grouping
comes from the session, not from a directory the browser cannot supply.

Per-file verdicts render as the stream checks run during upload, in
`internal/reject` wording. The journey states are honest about every real
outcome:

- **uploaded → checking → tagging → in your library** (the happy path, with
  a link into Navidrome once the scan has actually run),
- **needs a human look** — the album reached review; the copy says who looks
  (the admin), that this is normal for rare or unusual releases, and shows
  the 60-day clock,
- **couldn't be taken in** — rejected, with per-file reasons and what to do
  next,
- **waiting on your other upload** — a second upload by the same contributor
  queues behind the first, said plainly rather than silently starving.

Kind-copy rules, enforced in `internal/reject`: every `Reason` carries
plain-language copy and a `NextStep`; there is no raw-text fallthrough; and
stage-aware truth — "the admin has what you sent" renders only for material
that actually reached the box, never for stream rejections that did not.
An email not in the roster gets a warm "ask the admin to add you" page.
Revisit shows the contributor's own history from the ledger, including
resumable in-flight uploads (resume anchors server-side per verified email;
Web Storage fingerprints do not survive an iPad reload).

## Architecture

- **`dubplate serve`** — a long-running subcommand serving the embedded app
  and the API. **Reachability is a named mechanism, not an assumption**: the
  tunnel's cloudflared runs inside the compose network, so ingress cannot
  reach a host-loopback port. The pass chooses at plan time between
  `extra_hosts: ["host.docker.internal:host-gateway"]` on cloudflared with
  ingress at `http://host.docker.internal:<port>`, or running serve as a
  compose service; either way the ingress change is an API PUT (a deploy
  step this pass owns) and the firewall file is checked for consequences.
  serve binds loopback/bridge only and **rejects any request lacking the
  Access assertion header** — otherwise any process on the box forges
  identity by setting the header.
- **Identity** — verify the `Cf-Access-Jwt-Assertion` JWT properly: JWKS
  from the team-domain certs endpoint selected by `kid` (Access rotates
  signing keys every six weeks with a seven-day overlap, so the verifier
  caches with refresh-on-unknown-kid, rate-limited), validate `iss` against
  the team domain, pinned `aud`, `exp`/`nbf`, RS256 allowlist. `dubplate
  doctor` gains a probe of the certs endpoint and configured AUD. The
  verified email keys the ledger and the roster lookup; **the import
  pipeline's directory-name key is untouched** — `internal/config` gains an
  email-to-directory-name index instead, so the identity flip needs no
  pipeline change.
- **Session ID as the spine** — minted at tus upload creation, carried in
  upload metadata alongside the chosen collection, and stamped into the batch directory name
  (`inbox/<user>/<sessionid>__<album>/`) so it survives staging; the
  importer's run report carries per-session provenance (uploader, source
  dir, outcome, and the imported library paths resolved from the beets
  database after the run — the tag log records only skips and beets renames
  everything, so the database, not the filesystem, is the success record).
  The run report is written on **every** exit path, held-lock and panic
  included; absence of a report is a monitoring event, not a journey state.
- **Uploads** — embedded tusd (`tusd/v2/pkg/handler`, maintained, designed
  for embedding; its pre-create and pre-finish hooks are the metadata and
  assembly seams). **Chunk size 5–16 MB, configurable**: the binding limit
  is Cloudflare's ~100-second origin-response window against the
  contributor's upstream bandwidth, not the 100 MB body cap; an over-sized
  chunk 524s and retries from the same offset forever on exactly the
  residential links the page serves. `Upload-Expires` is set and a sweep
  reclaims abandoned partials — orphaned tus data would eat the free-space
  floor and fail imports closed. Server timeouts are configured to not
  fight tusd's own deadline management. Completed uploads assemble through
  `internal/archive` isolation into the session-named inbox directory.
- **Validation** — stream checks in-request (ZIP central directory, FLAC
  STREAMINFO), verdict per file, decision per album; `internal/gate`'s full
  decode stays the deep check in the import. Intake enforces the same size
  caps and symlink rules. Validator and verifier stay behind the prior
  draft's seams.
- **Journey transport: polling is the contract, a live stream is the
  enhancement.** The page is correct on 5-second polling of a journey
  endpoint (an import takes minutes; polling is immune to proxy behavior).
  SSE may be layered behind the same client interface only after **a
  pre-pass spike proves it through the real tunnel and Access**: cloudflared
  has documented event-stream buffering and a ~100-second idle cut, so the
  spike must demonstrate `text/event-stream` + `X-Accel-Buffering: no` +
  `no-transform`, a ≤20-second heartbeat, `Last-Event-ID` replay, and a
  sane "session expired — reload to sign in" state (the stream rides the
  Access cookie; expiry mid-journey must not render as a bare error).
- **The nudge** — serve knows when an upload session finishes, so the
  trigger names the settled session and the nudged run bypasses the
  mtime-quiescence heuristic for that session (`dubplate import --session
  <id>`); the 10-minute quiescence gate exists only because nothing
  previously knew when uploads were done. Mechanically: a trigger file in a
  directory both units may write under their hardening
  (`RuntimeDirectory=`/`MakeDirectory=` decided at plan time),
  `PathChanged=` semantics, coalescing accepted because the timer remains
  the backstop.

## Development shape

The app lives in `web/`, compiled by Vite to static assets the Go build
embeds; `make check` gains the web build and Svelte gates; deploy.sh orders
vite before go. Dev loop: Vite dev server proxying to a local serve.

**Framework — decided by Geoff 2026-08-31: Svelte 5 standalone** (no
SvelteKit): runes, components, svelte-conventions, and DaisyUI, compiled by
Vite into `web/dist` and embedded — no adapter, no SPA-fallback routing in
the Go file server (one `index.html`, one asset dir), no Kit migration
surface. DaisyUI is framework-agnostic CSS and carries most of the review
infrastructure. Two embed details still pre-decided at plan time:
`Cache-Control` for hashed immutable assets (embed.FS carries no mtimes),
and the Vite base path agreed between the dev proxy and embedded serving.

serve is the estate's first internet-facing unit: its systemd hardening is
written explicitly in the plan (not by reference), reviewed by
web-auth-security-reviewer alongside the JWT surface. Visual gates per the
design bar: Navidrome reference screenshots, fresh-context visual
verification, Geoff's before/after before family exposure.

## Pipeline changes this pass owns

The out-of-scope wall moved deliberately: the importer gains (1) per-session
run-report provenance with beets-database-resolved library paths, written on
every exit path; (2) the session-scoped `--session` import mode bypassing
quiescence; (3) nothing else. The Go rewrite lands the run-report structure
so this pass only extends it. Review filing, notification emails, and the
gate itself are unchanged.

## Pre-pass gates

1. The SSE spike through the real tunnel (or the decision to ship
   polling-only, which the design survives).
2. A timed large-chunk PATCH through the tunnel to validate the chunk-size
   config default.
3. iOS Safari picker behavior verified on the actual iPad before the copy
   is written.

## Prior art (scouted 2026-08-31)

No Svelte-native resumable-upload UI exists; the bespoke plan is confirmed,
not redundant. Rulings for the pass:

- **Adopt** `tus-js-client` directly, not through Uppy: 13 KB gz vs 35 KB
  plus Preact and lodash in the embedded bundle, and Uppy's file-state
  machine is replaced by our album-as-verdict-unit logic anyway. Pre-decide
  the Vite browser-entry/export-condition config: tus-js-client carries Node
  path dependencies and three of its last four releases were bundler fixes.
- **Crib, don't depend**: shadcn-svelte-extras' `FileDropZone`
  decomposition (Root / Trigger / DragOverlay, page-wide drag overlay,
  paste-to-upload) copied in and restyled to DaisyUI; Uppy 6's headless
  naming shape as the contract for the runes store.
- **Reject** svelte-file-dropzone (Svelte-3-era internals), svelte-filepond
  (dead adapter), svelte-dropzone-runes (one maintainer, nine stars).
- **DaisyUI `steps`** carries the linear journey spine
  (step-primary/success/error, data-content glyphs); the three branch
  outcomes render as an alert beside a frozen spine, never a fifth step.
  `steps` ships zero ARIA — add `aria-current="step"` and a live region —
  and horizontal steps overflow iPad portrait: `steps-vertical
  sm:steps-horizontal` or an overflow wrapper.

## Out of scope

Browsing or managing inbox files; the review flow's admin side; playback; a
group activity feed (revisit after the family has lived with the page).

## Sequencing

Builds on the Go rewrite's packages and config wall; runs as the pass after
the rewrite; packaging follows and inherits a finished intake. The rewrite
plan is amended (same day) so T9's run report carries the provenance
structure this spec depends on.
