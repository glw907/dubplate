# dubplate uploader: implementation plan

**STATUS: APPROVED by Geoff 2026-08-31 ("the plan looks good") after two
adversarial rounds (32 findings + 13 verification findings, see Review
provenance). Executes after the Go rewrite pass closes (T14 included).**

**Goal:** Build the estate's one user-facing surface — `dubplate serve` plus a
bespoke Svelte 5 / DaisyUI upload page replacing FileBrowser — with album-unit
validation at upload, an honest journey to "in your library", a collections
picker, and kind rejection copy, per the approved design spec.

**Spec:** `~/.dotfiles/docs/superpowers/specs/2026-08-31-dubplate-uploader-design.md`
(approved by Geoff 2026-08-31 after three adversarial reviews). The plan
argues from the spec; executors read both.

**Budget:** ceiling 5M tokens across both halves (checkpoint every 4 tasks;
score each half separately at its close).

**Execution — two workflow invocations with an attended gap (review-forced,
M17):**

- **Pass A (backend):** pass-execute, sequential, U1–U8, then U12a (box
  task, Geoff-sanctioned). Ends with serve deployed and reachable through
  Access, the JSON API verified end to end, FileBrowser still running.
- **Between passes (attended + spikes):** post-A gates 1–2 run against the
  real serve; gate 3 and Phase P (prototyping, Geoff's ratification) follow.
  Outcomes recorded in "Gate outcomes" below before Pass B starts.
- **Pass B (web):** pass-execute, sequential, U9–U11, then U12b (box task)
  and U13. Starts only after ratification is recorded.

Runs after the Go rewrite pass closes (T14 included — paths assume the
renamed repo `~/Projects/dubplate` and `DUBPLATE_*` env). Implementer
`general-purpose` pinned `model: sonnet`; reviewer `diff-reviewer` (opus).
Repo gate: `bash scripts/check.sh` (gains the web build in U9). Web tasks
invoke `svelte-conventions` + `ts-conventions`; unit tasks `vps-conventions`.
Pass-end fan-out: `svelte-reviewer`, `daisyui-a11y-reviewer`,
`web-auth-security-reviewer`, plus the visual-fidelity chain.

## Global constraints (every task)

- Go tasks: read the `go-conventions` skill first; module
  `github.com/glw907/dubplate`; every subprocess under `exec.CommandContext`;
  atomic writes via `internal/atomicfile`; table-driven tests under the
  assertion-discipline rules; no interfaces until a second in-tree consumer
  exists.
- Web tasks: Svelte 5 runes, standalone (NO SvelteKit), DaisyUI 5,
  `tus-js-client` pinned 4.3.x, Vite to static assets embedded via
  `go:embed`. Node is workstation build-time only.
- **Upgrade-resilient display code** (Geoff, 2026-08-31): stock DaisyUI 5
  components and semantic classes carry the design; custom CSS only where
  DaisyUI has no primitive, isolated in one `custom.css` with a comment
  naming the gap; never style or depend on DaisyUI internal class names;
  Svelte 5 runes idiom only, no legacy stores or `$:`; components are
  display-only over the runes store so upgrades touch state logic in one
  place. Each pinned web dependency carries a one-line upgrade note
  (tus-js-client 5.0-pre is a named watch item).
- Research constraints bind the web tasks:
  `~/.dotfiles/docs/superpowers/research/2026-08-31-uploader-prior-art.md`
  and `2026-08-31-uploader-ux-practices.md` (same dir). Hard rules: never
  pre-read a whole file client-side (iOS Safari memory ceiling
  ~100–200MB); ZIP is the primary path on every platform, folder-select a
  feature-detected desktop/Android extra hidden on iOS; a rejection
  renders attached to the album it concerns, never only out-of-band; no
  false "done" states — a stalled pipeline says "still working", never
  completion; one bad file never forces re-uploading files the server
  already holds (path-specific honesty in U10).
- Site-identifying values enter through config only (adopter goal); reject
  wording is contract (`internal/reject`), shared verbatim between page
  and emails; the album is the verdict unit — never admit a partial album,
  and a partial session directory never appears under the inbox root.
- The importer's directory-name key is untouched; identity mapping happens
  in config (U1). U1's config extension is this pass's COMPLETE field
  inventory — a later task discovering a missing field escalates instead
  of adding one.
- Journey truth comes from the ledger, never the filesystem; run reports
  are monitoring artifacts, not the history store (plan-time decision
  below).
- No live-box access in U1–U11. U12a/U12b name their box steps explicitly
  and run with Geoff's sanction (production box; every write through
  deploy.sh or a named, quiet-box step).

## Gates (renumbered; review-forced resequencing, M2)

Gates 1–2 need a real tus endpoint and SSE emitter behind the real tunnel
and Access — which is exactly what Pass A ships. They therefore run
**after Pass A, against the deployed serve**, not before anything exists.

1. **SSE spike (post-A):** against U7's dedicated probe endpoint
   (Pass A ships it precisely so this gate has a real target; see U7),
   with the field-proven mitigation set applied from the start
   (research 2026-08-31: cloudflared buffers GET-opened SSE —
   cloudflared#1449 — and its HTTP/2 path cuts idle streams near 60s):
   the probe stream is **POST-opened**, `text/event-stream` +
   `no-transform`, heartbeat comment every **15s**, flush after every
   write; prove incremental delivery, `Last-Event-ID` replay, and the
   "session expired — reload to sign in" state through the real tunnel
   and Access. If it misbehaves with all four mitigations in place,
   record polling-only and stop chasing (the design survives it; U11
   ships polling alone and the client seam stays).
2. **Chunk timing AND resumability (post-A):** timed large-chunk PATCHes
   through the tunnel from a residential link against the deployed serve
   — field reports converge on 5–10MB chunks (confirming the 8MB
   default; 16MB risks the 100s window on slow uplinks) — AND verify
   **offset persistence**: interrupt an upload mid-chunk and confirm
   tus HEAD reports real partial progress through the tunnel (a
   community report claims proxy buffering can hide offsets until
   request completion — unverified; this gate settles it). Confirm the
   zone's request-body cap once against the chunk size.
3. **iPad picker check (pre-Phase-P):** on real iPadOS Safari. Route
   decided by the 2026-08-31 infra research: the household iPad over
   USB to the workstation via usbmuxd + ios-webkit-debug-proxy (free;
   real console/network visibility during real uploads — Geoff's
   testing role is plugging the cable in; he remains final eye and
   design taste only). Farms are the fallback: BrowserStack's free tier
   is ruled out (file injection paywalled); LambdaTest ~$15/mo only
   after verifying large-file injection. Playwright WebKit is CI smoke
   for form JS only, never iOS truth (documented file-input
   divergence). The probe instrument is published (claude.ai artifact
   "Dubplate Picker Probe"): what the picker can select (multi-FLAC?
   ZIP? third-party providers?); feeds the prototyping brief and U10's
   copy.

## Gate outcomes (filled before Pass B starts)

- Gate 1 (SSE): _unrecorded_
- Gate 2 (chunk): _unrecorded_
- Gate 3 (iPad): **CLOSED documentarily 2026-08-31, no device needed**
  (two research rounds; artifacts in the research dir). Rulings for
  U10: (a) accept values are extension-first, never bare `audio/*` —
  WebKit bug 242110 (audio/* treated as video/*, files greyed out) is
  confirmed unfixed through iOS 18.3.1 (Gutenberg #70119, May 2025) and
  no Safari 18.x release note remediates it; (b) ZIP selection through
  the Files picker works in production (WeTransfer's own mobile-web
  docs; the famous iOS ZIP problem is download-side, never selection);
  (c) multi-select exists behind the "⋯ → Select" gesture — Apple
  documents it, NO shipping uploader's help does, so dubplate's copy
  coaching the gesture exceeds prior art; (d) `webkitdirectory` is
  absent (WebKit 271705), folder-select stays hidden on iOS; (e) no
  picker-side size ceiling exists — chunked resumable uploads are the
  production norm with no documented iOS byte cap. The published probe
  artifact remains a courtesy runtime confirmation, gating nothing.

## Phase P: UI/UX prototyping (post-gates, before U9–U11)

Run under the `visual-fidelity` skill's method; the ratified prototype is
the fidelity chain's reference for U13. Geoff's ratification is the one
attended gate in this phase, and Pass B does not start without it.

**P1 — Reference capture.** Screenshot Navidrome as the family actually
uses it (desktop wide, iPad portrait, phone; the theme in use), plus the
current FileBrowser flow as the "before" set. Stored under the research
dir; captured before any mockup exists — never build from a verbal
description.

**P2 — Journey storyboard.** Every screen and state as a storyboard before
any component: the idle page with the spec published up front (accepted
formats, the CD-quality floor, size caps, in the Bandcamp educational
register — why lossless matters here); per-device selection (desktop
whole-page drop + big picker button; iPad/iPhone native-picker-first with
multi-FLAC and ZIP copy per gate 3); uploading (per-file rows + batch bar,
honest server-acknowledged progress, path-specific verdict timing per
U10); checking/tagging (staged indicator, "safe to close this tab"
honesty: closing pauses, returning resumes); the four terminal branches
(in your library with a link that works; needs a human look with who
looks and the 60-day clock; couldn't be taken in with the per-track
checklist, one calm summary, honest retry per path; waiting — both copy
variants, own-upload and library-busy); the history view (terminal states
checkable from outside the flow, in-flight resumables listed); the
roster-absent warm page. Device matrix per screen: desktop / iPad
portrait (the steps-overflow case) / phone.

**P3 — Mockups and ratification.** Static mockups built directly in
DaisyUI 5 markup (they double as the component vocabulary and prove the
stock-components constraint); light and dark; the device matrix rendered
as real screenshots. The ten adopted patterns and the anti-pattern list
are the storyboard checklist — each anti-pattern gets an explicit "where
we avoid it" note. Geoff reviews on his actual devices and ratifies;
ratified captures are the U13 reference set.

## Plan-time decisions (ratified here so no task re-litigates)

- **Reachability**: serve runs on the host under systemd. cloudflared
  gains `extra_hosts: ["host.docker.internal:host-gateway"]`; the compose
  network subnet is pinned in `compose.yaml` so the gateway IP is stable;
  serve binds the configured `Serve.BindAddrs` list (loopback + the
  gateway IP), never `0.0.0.0`. Ingress update is an API PUT
  (config/cloudflared-tunnel-config.json is the record). Defense in
  depth: the Hetzner firewall admits only tcp/22 inbound regardless.
- **Access enforcement**: every request must carry a valid
  `Cf-Access-Jwt-Assertion` (JWKS by `kid`, cached, refresh-on-unknown-kid
  rate-limited to one refetch per 5 minutes; `iss` = team domain, pinned
  `aud`, `exp`/`nbf`, RS256 only). Authorization is separate and explicit:
  every session-scoped operation checks the verified email against the
  session's owner (U5/U6).
- **Session identity**: session IDs are 128 bits from `crypto/rand`,
  base32 lowercase (`^[a-z2-7]{26}$`), validated at every ingress that
  accepts one (API paths, tus metadata, trigger filenames). Multi-file
  sessions are created explicitly (`POST /api/session`); a lone ZIP
  upload without a session ID mints one implicitly at tus pre-create.
- **Session sidecar**: at promotion, serve writes `session.json`
  (`{version, session_id, email, dir_name, album, collection,
  created_at}`) into the session directory BEFORE the atomic rename into
  the inbox; it is the collection's transport to the importer and the
  session key's durable home.
- **Ledger**: append-only JSONL at `state/contributions.jsonl`,
  multi-writer (serve AND the importer) via `O_APPEND`, every record
  written in ONE `write(2)` call with no short-write loop (the actual
  interleave guarantee for concurrent appenders; asserted in a test) +
  fsync, records ≤ 4096 bytes. **One appender per state** (review-forced,
  N6): the importer appends `tagging` (carrying the beets-resolved
  `LibraryPaths`), `review`, and `rejected`; serve appends the upload-side
  states, `stalled`, and — on observing Navidrome scan completion —
  `in-library`. History and journey therefore read the same rows and can
  never disagree. Run reports remain per-run monitoring artifacts
  (`report-<runid>.json`, retention `Thresholds.ReportRetention`) and the
  journey NEVER reads them.
- **Nudge trigger directory**: derived from `Paths.RunDir` (the literal
  `/run/dubplate` appears only in the tmpfiles fragment, which deploy.sh
  renders and installs to `/etc/tmpfiles.d/` and activates with
  `systemd-tmpfiles --create`). The path unit uses
  `DirectoryNotEmpty=` (level-triggered — `PathChanged=` loses triggers
  written during a run); the consumer is a Go subcommand,
  `dubplate nudge-consume`, that owns the whole loop IN-PROCESS
  (review-forced, N3 — an exit-and-retry consumer cannot observe lock
  contention and level-triggering would spin it against a left-behind
  file): it reads the trigger dir, acquires the import lock via
  `lock.AcquireBlocking` under a bounded context timeout, runs each
  session's import in-process under the held lock, deletes each trigger
  only after its import succeeds, and exits nonzero only on real
  failure; the nudge service sets `StartLimitIntervalSec=0`. The
  15-minute import timer stays the backstop.
- **Embedded asset caching**: hashed `/assets/*` get
  `Cache-Control: private, max-age=31536000, immutable` (responses ride
  an authenticated transport — never `public`); `index.html` gets
  `no-cache`. Vite `base` stays `/`; embed home is `web/embed.go`
  (`package web`, `//go:embed all:dist`) with a committed non-dot
  `web/dist/placeholder.txt` so Pass A builds before Vite ever runs.
- **Chunk default**: `UploadChunkMB` default 8 pending gate 2.

---

## Pass A — backend

### U1: Config extension and identity index

**Files:** `internal/config/config.go`, `internal/config/contributors.go`,
tests alongside; `~/.dotfiles/secrets/registry.md` + age store (minting).

**Produces:** the pass's complete config-field inventory, added in one
task: `Serve` (BindAddrs []string, Port, AccessTeamDomain, AccessAUD,
PublicNavidromeURL — all required), `Upload` (ChunkMB, MaxAlbumBytes,
PartialTTL, MaxSessionsPerUser — thresholds, defaulted), `Paths` gains
`UploadsDir` (the tus store, required), `LedgerPath` (defaulted from
StateDir), `Thresholds` gains `StallMinutes` (defaulted, and the default
MUST exceed the import timer interval plus the quiescence window so the
stall state cannot fire on a healthy backstop path — state the arithmetic
in a comment) and `ReportRetention` (defaulted, 30 run-report files —
thresholds live here, not as constants). Roster gains
`DirNameFor(email string) (string, bool)`; the importer's directory-name
key is untouched.

**Criteria:** rewrite-T2 rules exactly (required fields join the
`ErrMissing` `errors.Join` list; env through the single values getter;
`DUBPLATE_*` names). Every new env NAME this task introduces is minted
into the age store and registry in THIS task (the secrets flow's
same-session rule); U12a only renders them. Tests: unknown email absent;
roster round-trip; every new required field missing → named in the joined
error; StallMinutes default arithmetic asserted against the timer +
quiescence constants. **This is the complete inventory** — U2–U13
escalate rather than add a field.

### U2: Access JWT verifier and doctor probes

**Files:** `internal/access/access.go`, tests; `cmd/dubplate/doctor.go`
(extend).

**Produces:**

```go
func NewVerifier(teamDomain, aud string) *Verifier
func (v *Verifier) Verify(ctx context.Context, token string) (email string, err error)
```

**Criteria:** JWKS from the team-domain certs endpoint, key by `kid`,
cached; refresh-on-unknown-kid rate-limited to one refetch per 5 minutes
(test: a burst of unknown-kid requests triggers exactly one refetch);
validates `iss`, pinned `aud`, `exp`/`nbf`; alg allowlist RS256 only
(`alg=none` and HS256 rejected — test each). All against `httptest` JWKS
fixtures with generated keys. Doctor gains: certs-endpoint + AUD probe;
UploadsDir writability; ledger appendability; run-dir presence;
PublicNavidromeURL syntactic sanity; UploadsDir and InboxDir on the same
filesystem (the atomic-promotion precondition — N13). Written for the
`web-auth-security-reviewer` audience.

### U3: Contributions ledger

**Files:** `internal/ledger/ledger.go`, tests.

**Produces:**

```go
type State string // the one journey vocabulary: "uploading","uploaded","checking","tagging","in-library","review","rejected","queued","stalled"
func ValidState(s State) bool
func Open(path string) (*Ledger, error)
type Event struct { SessionID, Email, Album, Collection string; State State; At time.Time; Detail string; LibraryPaths []string }
func (l *Ledger) Append(e Event) error
func (l *Ledger) History(email string) ([]Event, error)   // newest first
func (l *Ledger) Session(id string) ([]Event, error)
```

**Criteria:** append-only JSONL, `O_APPEND` + fsync, **multi-writer**
(serve and the importer are both sanctioned appenders); every record is
written in a single `write(2)` with no short-write loop — the interleave
guarantee for concurrent appenders to a regular file — asserted in a
test. A record that would exceed 4096 bytes degrades deterministically
(N8): `LibraryPaths` collapses to the album root plus a track count,
`Detail` truncates next; a record is never split and the terminal event
is never rejected — tested with real path lengths from a multi-disc
release, not a synthetic short fixture. `Append` rejects an invalid
`State` (exhaustive test over the vocabulary, `stalled` included). A
torn final line (crash mid-write) is skipped with a logged warning
(truncated fixture test). Missing file = empty history. The `State` type
defined HERE is the vocabulary `internal/journey` (U6) and the web store
(U9) reuse — one vocabulary end to end. The one-appender-per-state map
from the plan-time decisions is documented on the type.

### U4: Stream validators and the reject wire format

**Files:** `internal/intake/intake.go`, tests with fixture ZIPs and
FLACs; `internal/reject/reject.go` (extend), the deployed email template
asset + its renderer test.

**Produces:**

```go
type FileVerdict struct { Name string; OK bool; Reason reject.Reason }
func CheckZIP(r io.ReaderAt, size int64, p gate.Policy) ([]FileVerdict, error)
func CheckFLAC(r io.Reader, name string, p gate.Policy) (FileVerdict, error)
type AlbumVerdict struct { OK bool; Files []FileVerdict }
func Decide(files []FileVerdict) AlbumVerdict
```

**Criteria:** ZIP central-directory walk (no full extraction) checks
names, sizes against caps, symlink/absolute/traversal entries — three
new enum values are explicitly authorized here and nowhere else:
`ReasonUnsafePath` (these checks), `ReasonNoSpace` and
`ReasonTooManySessions` (U5's admission control), each with friendly +
next-step copy written in this task; FLAC STREAMINFO parse checks magic,
sample rate, bits-per-sample
against the same CD-quality floor as `internal/gate` (shared constants,
not duplicates). `Decide` implements album-as-verdict-unit. The deep
check (`flac -t`) stays the import's job (doc comment). `internal/reject`
gains `func (r Reason) NextStep() string` (every Reason returns non-empty
from both `Friendly` and `NextStep` — exhaustive test) and
`MarshalJSON` emitting
`{"code":"below-cd-quality","friendly":"...","next_step":"..."}` — the
wire format U6 serves and U11 renders verbatim (round-trip + golden
tests; a bare-integer encode is the failure this exists to prevent). A
rendered rejection email and the API JSON carry the same `Friendly` and
`NextStep` strings (cross-asset test). Fixtures: truncated FLAC,
below-CD FLAC, traversal ZIP entry, clean album ZIP.

### U5: Upload engine — tusd, sessions, completion, promotion

**Files:** `internal/serve/upload.go`, `internal/serve/session.go`,
tests; `go.mod` gains `github.com/tus/tusd/v2`.

**Produces:** tusd mounted at `/files/` (U7 routing and U9's dev proxy
agree); `POST /api/session` (album, collection, expected_files) →
session ID; `POST /api/session/<id>/finish`;
`type Session struct { ID, Email, Album, Collection string; ExpectedFiles int }`;
promotion into `inbox/<dirname>/<sessionid>__<albumslug>/`.

**Criteria:**

- **Identity and authorization (review-forced, B3):** pre-create stamps
  the verified email into the upload's metadata; pre-PATCH and
  pre-terminate hooks reject any request whose verified email differs —
  as 404, never 403 (no existence oracle). Session lookups enforce the
  same ownership. Cross-user PATCH and cross-user session access both
  have denial tests. Session IDs follow the plan-time format and are
  validated at every ingress.
- **Input sanitization (B4):** `Album` is normalized to a bounded slug
  (length cap, allowlist charset, no leading dot, non-empty) before any
  path use; `Collection` must be a member of `config.Collections` or
  pre-create rejects. Table-driven tests: `../`, absolute paths, NUL,
  unicode-normalization pairs, unconfigured collection.
- **Session completion (B2):** a multi-file session completes when
  `uploads_done == ExpectedFiles` or on explicit `/finish`; a lone ZIP
  completes at its pre-finish; an upload past `ExpectedFiles` is rejected
  at pre-create (N13). Only at completion does `Decide` run over
  the session's accumulated verdicts; an abandoned session expires via
  `PartialTTL` without promotion. Tests: a partial session NEVER appears
  under the inbox root; an abandoned session leaves no inbox trace; the
  N+1st upload is refused.
- **Atomic promotion (M9, B6):** uploads assemble into
  `<UploadsDir>/<sessionid>/` (ZIPs through `internal/archive.Extract` —
  isolation, caps); at completion-with-OK-verdict, `session.json` is
  written into the directory, then ONE `os.Rename` moves it to the inbox
  path (same filesystem — UploadsDir's placement makes this true; state
  it in config docs). The rename, not a lock, is the ordering guarantee
  (doc comment states the argument; a test proves no partial inbox
  state).
- **Admission control (M13):** pre-create rejects when
  `archive.FreeSpace(UploadsDir)` minus the declared `Upload-Length`
  would cross `FreeFloorGB`, and when the caller already has
  `MaxSessionsPerUser` in-flight sessions — with a `reject.Reason` and
  kind copy, not a bare 500.
- **Mechanics:** chunk size honors `Upload.ChunkMB`; `Upload-Expires`
  stamped; the sweep (frozen-clock seam) deletes expired partials on a
  ticker. HTTP timeouts: `ReadHeaderTimeout` on the tus routes,
  `ReadTimeout`/`WriteTimeout` zero THERE (tusd owns transfer deadlines),
  ordinary timeouts elsewhere; a slow-PATCH test proves a trickling
  upload is not cut.

### U6: Journey model and polling endpoint

**Files:** `internal/journey/journey.go`, `internal/serve/api.go`, tests.

**Produces:**

```go
type LiveSession struct { SessionID, Email string; State ledger.State; Files []intake.FileVerdict }
type LiveSet map[string]LiveSession
type Journey struct { SessionID string; State ledger.State; Detail string; LibraryLink string; Files []intake.FileVerdict }
func Resolve(sessionID string, l *ledger.Ledger, live LiveSet) (Journey, error)
```

`GET /api/journey/<sessionid>` (5s polling contract), `GET /api/history`,
`GET /api/config` is U7's. `LiveSession`/`LiveSet` are concrete types
owned HERE by `internal/journey` (N2 — `serve` constructs them and
passes down; the dependency runs serve → journey → ledger, one
direction, no interface). This task also adds
`func Held(path string) (bool, error)` to `internal/lock` via `fcntl`
`F_GETLK` (authorized here, N10) — the non-destructive probe the
`queued` library-busy variant needs; serve never takes the import lock
to test it.

**Criteria:** journey composes from the ledger plus serve's live session
state ONLY (run reports are monitoring artifacts). Authorization: a
journey or history read is scoped to the verified email; someone else's
session ID returns 404 (test). **The `in-library` ledger event is
appended by SERVE, not the importer** (N6): the importer's terminal
append is `tagging` (carrying `LibraryPaths`); serve polls Subsonic
`getScanStatus` after observing a `tagging` event and appends
`in-library` only when a scan has completed after the import finished
(stubbed-Navidrome test) — so `/api/history` and `/api/journey` read
the same rows and can never disagree, and a history link can never 404
(M4). `LibraryLink` is built from
`Serve.PublicNavidromeURL` plus Navidrome's search route for the album
title (the exact route verified against the pinned Navidrome version in
the task report; a deep album link is permitted only if resolvable from
ledger data without new config). `queued` has two copy variants from one
state (M6): an unfinished earlier session for the same email → "waiting
on your other upload"; the import lock held by another run → "waiting for
the library to finish its current import" — both tested. A session with
no terminal event past `Thresholds.StallMinutes` gets a `stalled` ledger
event appended by serve (a first-class vocabulary member, N4 — it is
both the de-dupe record and what U11 renders) AND one admin notification
(a `notify.Mailer` send; the ledger event is the de-dupe, so a restart
cannot re-send — test), making U11's "the admin has been notified" copy
true (M5). SSE lands here in Pass B only if gate 1 passed, behind the
same client interface, by promoting U7's probe endpoint to the journey
stream.

### U7: serve command — routing, embed, enforcement

**Files:** `cmd/dubplate/serve.go`, `internal/serve/serve.go`,
`web/embed.go`, `web/dist/placeholder.txt`, tests.

**Criteria:** serves `web.Assets` (`//go:embed all:dist` in `web/embed.go`
— the placeholder ships now so Pass A builds; U9's Vite build replaces
dist content and must keep the placeholder or update the embed pattern in
the same change); Cache-Control per plan-time decision (`private`, not
`public`). Enforcement: every route requires a valid Access assertion;
**failure shape is per surface** (M10, m5): document requests from a
roster-absent email get the warm 200 HTML "ask the admin to add you"
page; `/api/*` and `/files/*` get machine-readable JSON — 401 for a
missing/expired assertion (the client renders reload-to-sign-in), 403
with the warm copy for roster-absent. A dev mode exists ONLY as
`--dev-identity <email>`: refuses to start unless every BindAddr is
loopback, off by default, inert on non-loopback (all three tested) — the
documented dev loop runs through it, so no implementer ever improvises a
bypass. `GET /api/config` returns collections, chunk MB, caps, the
review-clock days (from `Thresholds.RetentionDays` — U11's "60-day
clock" source, N11), and the accepted-format spec text derived from
`gate.Policy`'s constants plus `Upload.MaxAlbumBytes` (generated, never
hand-written — the page's single source; "who looks" copy says "the
admin", no new field). Serves `GET /api/stream/probe` (N1): a
deliberately minimal SSE endpoint — correct headers, ≤20s heartbeat,
`Last-Event-ID` replay, no journey semantics, ~40 lines — existing only
as gate 1's measurement target; Pass B promotes it (U6) or deletes it
(U12b). Binds exactly `Serve.BindAddrs` via a `net.ListenConfig.Control`
that sets `IP_FREEBIND` (N7 — binding the not-yet-existing bridge
gateway IP at boot succeeds; test binds an absent address), rejecting
any unlisted address (test). At startup, serve verifies UploadsDir and
InboxDir share a filesystem and refuses to start otherwise (N13 — the
atomic-promotion precondition, checked where it matters, not only in
doctor). Writes the nudge
trigger file (the session ID, validated format) to the run-dir-derived
trigger directory ONLY after promotion's rename returns; a nudge write
failure logs and never fails the upload (timer backstop). Graceful
shutdown drains in-flight tus PATCHes on SIGTERM (in-flight-request
test).

### U8: Importer session mode, terminal events, nudge units

**Files:** `internal/importer/importer.go` (extend),
`cmd/dubplate/import.go` (flag), `systemd/dubplate-import-nudge.path`,
`systemd/dubplate-import-nudge.service`, `config/tmpfiles-dubplate.conf`,
tests.

**Criteria:** `dubplate import --session <id>` processes exactly that
session's directory, bypassing quiescence for it; it reads
`session.json` and passes the collection to beets
(`--set collection=<c>`, default `Collections[0]` when absent) — the
picker's transport (B6). The run report gains `SessionID` and
`Collection` per album (schema version bump named in the report), becomes
per-run files `report-<runid>.json` swept at `Thresholds.ReportRetention`
(M1) — and stays a monitoring artifact. **The importer appends its
terminal ledger events** — `tagging` (with beets-resolved
`LibraryPaths`), `review`, `rejected`; `in-library` belongs to serve
(N6) — for sessions it settles, timer runs included (test: a timer run
and a `--session` run both append; ledger history survives ten
consecutive runs). Nudge: `DirectoryNotEmpty=` on the trigger dir
(level-triggered, M8); the service runs `dubplate nudge-consume`, which
owns the loop in-process per the plan-time decision (N3): read the
trigger dir, `lock.AcquireBlocking` under a bounded context timeout, run
each session's import in-process under the held lock, delete each
trigger only after its import succeeds, exit nonzero only on real
failure (delete-ordering and lock-contention unit tests);
`StartLimitIntervalSec=0` on the service. Both units carry the full
hardening block (`NoNewPrivileges=yes`, `PrivateTmp=yes`,
`ProtectSystem=strict`, explicit `ReadWritePaths=` for the trigger dir,
inbox, staging, state, library — enumerated, `systemd-analyze verify`
green). The tmpfiles fragment ships here; U12a installs it. The import
timer remains the backstop, untouched.

### U12a: Ship serve (box task, Geoff-sanctioned)

**Files:** `systemd/dubplate-serve.service`, `scripts/deploy.sh`,
`compose.yaml` (pinned subnet + cloudflared extra_hosts),
`config/cloudflared-tunnel-config.json`, `env/` template,
`docs/bringup-runbook.md`, `cloud-init/user-data.yaml` (if the box needs
nothing new, say so rather than editing).

**Box steps, box quiet:** deploy binary + units; install the tmpfiles
fragment to `/etc/tmpfiles.d/dubplate.conf` and run
`systemd-tmpfiles --create` on it (a `--check` run verifies fragment AND
directory); start serve; ingress API PUT adding the uploader hostname →
`http://host.docker.internal:<port>`; create the Access application and
record its AUD (runbook documents every step for a stranger — the
adopter-path rule, M15); verify the JSON API end to end through Access
from a real client. **FileBrowser stays up** — it retires in U12b.

**Criteria:** `dubplate-serve.service` hardening written in full: the U8
baseline PLUS `ProtectHome=yes`, `ProtectKernelTunables=yes`,
`ProtectKernelModules=yes`, `ProtectControlGroups=yes`,
`RestrictNamespaces=yes`, `RestrictRealtime=yes`,
`RestrictSUIDSGID=yes`, `LockPersonality=yes`,
`MemoryDenyWriteExecute=yes`, `SystemCallFilter=@system-service`,
`SystemCallArchitectures=native`, `AmbientCapabilities=` (empty),
`CapabilityBoundingSet=` (empty), `UMask=0027`,
`RestrictAddressFamilies=AF_INET AF_INET6 AF_UNIX`,
`ReadWritePaths=` exactly the uploads dir, trigger dir, state dir, inbox;
`Restart=on-failure`, `RestartSec=5`, `StartLimitIntervalSec=0`,
`After=docker.service` — boot resilience comes from U7's `IP_FREEBIND`
listener (there is no such service directive as `IPFreeBind=`, N7; the
restart settings are the documented backstop, not the fix);
`TimeoutStopSec=15` (a mid-PATCH SIGKILL is safe — tus resumes — say so
in a comment). Gate on `systemd-analyze verify` AND
`systemd-analyze security dubplate-serve.service` reporting the
installed systemd's "OK" band — record the exact label and score the
tool prints in the task report rather than a remembered number (N12).
This deploy also publishes U7's `/api/stream/probe` as gate 1's named
target. deploy.sh orders the
web build before the Go cross-compile; manifest edits land with the unit
they ship. Env rendering only — names were minted in U1. Rollback
documented AND complete: stop serve, `systemctl disable --now
dubplate-import-nudge.path` (a rolled-back binary has no `--session`
flag, M16), ingress PUT back, tmpfiles fragment removal optional.

## Pass B — web (starts only after gates + ratification recorded)

### U9: Web scaffold and upload store

**Files:** `web/package.json`, `web/vite.config.ts`,
`web/svelte.config.js`, `web/src/main.ts`, `web/src/App.svelte`,
`web/src/lib/store/upload.svelte.ts`, `web/src/lib/tus.ts`,
`web/src/lib/platform.ts`, store unit tests (vitest), `Makefile` +
`scripts/check.sh` (web gate wiring).

**Interfaces — produces:** the runes store every U10/U11 component
consumes: per-file `pending → uploading → ok | failed` and batch states
using `ledger.State`'s vocabulary verbatim; `progressBytes` (transfer)
and `ackedBytes` (server-confirmed tus offset) kept distinct. Consumes
U7's `/api/config` and the tus endpoint at `/files/`.

**Criteria:** Vite + Svelte 5 standalone builds to `web/dist` with
`base: "/"`; the embed placeholder moves to `web/public/placeholder.txt`
so every Vite build re-emits it into `dist` (N9 — Vite's default
`emptyOutDir` would otherwise delete U7's committed copy and break the
Go build); tus-js-client
browser-entry/export-condition config proven by a production build with
no Node shims; the gate gains `npm ci`, `svelte-check`, vitest, and the
production build inside `make check`. The store is UI-free and fully
unit-tested: state transitions; the honest-progress split; resume
anchoring (persists only the tus upload URL — no IndexedDB/localStorage
file caching; "re-select the file after a full browser close" is a store
state, not a surprise); **the retry ladder (M14)**: explicit
`retryDelays` starting from the field-proven default
`[0, 3000, 5000, 10000, 20000]`; on repeated 502/504/524 at the same
offset, halve
`chunkSize` down to a 1MB floor and surface a "your connection is slow —
using smaller pieces" note; when the budget is exhausted, a distinct
honest stalled state with manual retry (downshift ladder unit-tested
against a stubbed failing endpoint); **401 handling (M10)**: a 401 or an
HTML content-type where JSON was expected renders the reload-to-sign-in
state, and tus retries do not consume their budget on 401. Chunk size
and caps come only from `/api/config`. `platform.ts` is feature
detection (`webkitdirectory` probe) plus the minimal iOS-Safari check
the copy switch needs. Dev loop documented: Vite dev server proxying
`/api` and `/files` to a local `dubplate serve --dev-identity`.

### U10: Upload UI — selection, verdicts, collection picker

**Files:** `web/src/lib/components/DropZone.svelte`,
`DropOverlay.svelte`, `PickerButton.svelte`, `CollectionPicker.svelte`,
`SpecCard.svelte`, `AlbumCard.svelte`, `FileRow.svelte`, component
tests.

**Interfaces — consumes:** U9's store exclusively; components are
display-only (no fetch, no tus calls in components).

**Criteria:** desktop: whole-page drop target (crib
shadcn-svelte-extras' Root / Trigger / DragOverlay decomposition,
restyled to stock DaisyUI classes only) with paste-to-upload; every
platform: the primary control is a large native-picker button —
drag-drop is enhancement, never required. ZIP is the primary path
everywhere; folder-select renders only where `webkitdirectory`
feature-detects (hidden, not disabled, on iOS Safari — WebKit 271705
confirms it unimplemented); **accept values are extension-first, never
bare `audio/*`** (a documented WebKit UTI bug greys out audio files in
the Files picker for MIME-wildcard accepts; extension lists are the
field workaround — research 2026-08-31, probe confirms when run);
iPad/iPhone copy presents multi-FLAC selection and "zip it first" as
first-class per gate 3, and explicitly surfaces the **"⋯ → Select"
multi-select gesture** (iPadOS hides multi-select behind it —
discoverability is a copy job, not an assumption).
`SpecCard` publishes formats, the CD-quality floor, and caps up front
from `/api/config`, in the educational register. No client-side
full-file reads ever. **Verdict timing is path-honest (M7):** on the
multi-file path, per-file verdict rows genuinely stream as each upload's
checks run, and retry re-sends only failing files; on the ZIP path the
copy says up front "we check the tracks once the zip finishes
uploading", verdicts render as one batch at 100%, and a failed ZIP's
retry is honest that it re-uploads the archive. All rejection strings
come from the API wire format (U4) — never client-side lookup tables.
CollectionPicker renders only when `/api/config` lists more than one
collection. Layout: single-column full-width cards below `md:`,
row/table above; every interactive control carries DaisyUI's default
`btn` sizing or explicit `min-h-11 min-w-11` (44px) — checked in the
a11y review pass. Keyboard path complete: picker, per-row actions, and
retry all operable without a pointer.

### U11: Journey UI — steps spine, outcomes, history

**Files:** `web/src/lib/components/Journey.svelte`,
`JourneySteps.svelte`, `OutcomeAlert.svelte`, `History.svelte`,
`web/src/lib/announce.ts`, component tests.

**Interfaces — consumes:** U9's store (journey via 5s polling of
`/api/journey/<sessionid>`; SSE only if gate 1 passed, same client
interface).

**Criteria:** DaisyUI `steps` spine
(`step-primary`/`step-success`/`step-error`, `data-content` glyphs),
`steps-vertical sm:steps-horizontal` (iPad portrait never overflows);
`aria-current="step"`. The three branch outcomes render as a DaisyUI
alert beside the frozen spine — never a fifth step. One polite
`aria-live` region exists empty from first render (`announce.ts` owns
it): announcements at ≥25% progress deltas and on state changes only;
completion is a distinct final announcement. All progress animation
behind `prefers-reduced-motion` (instant jumps when reduced). Honesty:
transfer vs server-acknowledged progress render distinctly; backgrounding
copy states closing/switching away pauses and returning resumes; an
optional "keep my screen awake" toggle uses Screen Wake Lock where
supported, presented as exactly that. The stall state renders "taking
longer than expected — the admin has been notified" — true because U6
sends it. `History.svelte` renders `/api/history`: unambiguous terminal
states checkable without an active flow; in-flight resumables first with
a resume affordance; review items show who looks and the 60-day clock;
queued shows the applicable copy variant. Every rejection string comes
from the API.

### U12b: Cutover and FileBrowser retirement (box task, Geoff-sanctioned)

**Files:** `compose.yaml` (remove FileBrowser),
`config/cloudflared-tunnel-config.json`, `scripts/deploy.sh` manifests,
`env/` template (FileBrowser vars retired), `docs/bringup-runbook.md`.

**Box steps:** deploy the full binary with embedded page; Geoff verifies
the page on his actual devices (desktop + iPad + phone) through Access;
only then: FileBrowser's compose service removed, its ingress hostname
retired (or repointed), its env names retired from template + age store
in the same commit. Rollback: redeploy previous binary, restore
FileBrowser service + ingress from git.

**Criteria:** the retirement commit removes every FileBrowser reference
(compose, env template, deploy manifests, runbook, STATUS pin table) in
one commit — `grep -ri filebrowser` afterwards returns only HISTORY;
manifest edits land with the deletion (rsync exit-23 rule). If gate 1
failed (polling-only shipped), this task also deletes `/api/stream/probe`
— nothing measurement-only survives the pass.

### U13: Pass-end verification and family exposure

**Criteria:** reviewer fan-out (`svelte-reviewer`,
`daisyui-a11y-reviewer` — the 44px and live-region criteria are named
checks in its dispatch, `web-auth-security-reviewer` over U2/U5/U6/U7);
the visual-fidelity chain against the Phase P ratified reference
(fresh-context `visual-verifier`, never the building context); Geoff's
before/after on his devices precedes any family member getting the URL;
ledgers updated (STATUS, HISTORY, ROADMAP move). No inherited leftovers:
anything the rewrite pass's close-out owed (vale-comments gate,
tidy-check) is that pass's business and does not ride here.

## Review provenance

Adversarially reviewed at authoring (opus agent, 2026-08-31): 6 blockers,
17 majors, 9 minors — all folded. The load-bearing folds: the session ID
now travels to the importer via the `session.json` sidecar and comes back
through importer-appended terminal ledger events (B1/B6/M1); multi-file
session completion is defined and promotion is a single atomic rename, so
a partial album can never reach the inbox (B2/M9); authorization is
explicit — ownership checks on every session-scoped operation, 404 shape,
`crypto/rand` session IDs (B3); album/collection inputs are
sanitized/validated (B4); the embed home is `web/embed.go` with a
committed placeholder (B5); gates 1–2 resequenced to post-Pass-A against
the real serve (M2); the config inventory gained UploadsDir,
PublicNavidromeURL, BindAddrs-as-list, StallMinutes and lost its
self-contradictory clause (M3/M12); `in-library` waits for scan
COMPLETION (M4); the stall state's admin notification exists and is
de-duped (M5); `queued` covers the global-lock case (M6); verdict timing
and retry promises are path-honest (M7); the nudge uses
`DirectoryNotEmpty=` with delete-after-success (M8); dev mode is an
explicit loopback-only flag and API auth failures are machine-readable
(M10); the reject wire format is specified with golden tests (M11); the
serve unit gained the full hardening set, boot-ordering fixes, and a
security-score gate (M12); admission control exists (M13); the client
retry/downshift ladder is specified (M14); the adopter runbook and doctor
probes are owned (M15); the tmpfiles fragment is installed, and rollback
disables the nudge path unit (M16); the pass split at U8 with the
attended gap (M17). Minors m1–m9 folded as written (secret minting in U1;
numeric criteria; email-template ownership + `ReasonUnsafePath`; RunDir
derivation; API-vs-document failure shapes; no single-consumer interface;
ledger State type; hygiene sections; `private` cache).

A verification re-review by the same opus agent scored all 32 folds
FOLDED and surfaced 13 seam defects (N1–N13), folded in the same
session: gate 1 gained a real target (U7's `/api/stream/probe`, deployed
by U12a, deleted by U12b if unused); `LiveSession`/`LiveSet` moved into
`internal/journey`, breaking the import cycle; the nudge consumer became
`dubplate nudge-consume`, an in-process loop that can actually observe
lock contention and cannot spin the level-triggered path unit; `stalled`
joined the ledger vocabulary as a first-class state; `ReasonNoSpace` and
`ReasonTooManySessions` were authorized in U4; `in-library` became
serve's append (scan-completion-observed) while the importer appends
`tagging` with `LibraryPaths`, so history and journey read identically;
the fake `IPFreeBind=` directive was replaced by a real `IP_FREEBIND`
listener control in U7 with restart settings demoted to backstop; the
ledger's overflow rule degrades `LibraryPaths` deterministically with a
single-`write(2)` guarantee and a multi-disc test; the embed placeholder
moved to `web/public/`; `lock.Held` (F_GETLK) was authorized for the
queued probe; `/api/config` gained the review clock and a generated spec
text; retention moved to `Thresholds.ReportRetention` and the security
score gate records the installed tool's own label; `ExpectedFiles`
overrun is rejected and the same-filesystem precondition is checked at
serve startup and in doctor. The verifier's closing guidance — a
coordinator read of these edits suffices; no third adversarial round —
was followed.

## Self-review notes

Spec coverage: product frame/adopter wall → global constraints + M15 fold
(runbook); tus client + protocol seam → U9; tusd server, session spine,
Upload-Expires sweep, chunk config, admission → U5; stream validation +
album verdict + kind-copy wire contract → U4; ledger (multi-writer
decision) → U3; journey states incl. queued-both-variants, stall,
stage-aware truth, scan-completion honesty → U6/U11; roster-absent warm
page (document requests) + API failure shapes → U7; reachability, JWKS
detail, doctor probes → decisions + U2/U12a; email→dirname index → U1;
nudge (--session, sidecar, DirectoryNotEmpty, tmpfiles, timer backstop) →
U8/U12a; development shape (web/, vite, gate, dev-identity proxy,
Cache-Control, base, embed home) → U9/U7; explicit hardening + security
score → U8/U12a; visual gates → Phase P + U13; prior-art rulings →
U9/U10/U11; FileBrowser retirement ordering → U12b. Out-of-scope wall
holds: no group feed, no admin review UI, no playback.

Type consistency: `ledger.State` is the one journey vocabulary (U3 → U6 →
U9); `intake.FileVerdict` + `reject.Reason` (with `MarshalJSON`) are the
one rejection vocabulary end to end; `Session{ID,Email,Album,Collection,
ExpectedFiles}` matches the sidecar schema and the ledger Event fields;
`LiveSession`/`LiveSet` are concrete journey-owned types, no interface,
dependency serve → journey → ledger one direction. Every task's
"consumes" exists before it runs in the stated order; Pass B consumes
only Pass-A-landed surfaces plus gate/ratification outcomes recorded in
this file.
