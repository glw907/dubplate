# dubplate-organizer: initiative design

Status: high-level design, ratified in brainstorm 2026-09-01 (Geoff +
Claude), revised the same day to fold two adversarial reviews
(`docs/superpowers/research/2026-09-01-organizer-spec-reviews/`) and the
owner rulings they forced. This spec sets the initiative's success bar,
architecture, and build order. Each ladder rung gets its own spec and
plan when its turn comes; none exists yet. Preconditions: the uploader
initiative ships and the prod-polish pass completes.

## The success bar

The initiative's bar is one sentence: **at least as good as beets for
core tagging**. Nothing else in the effort matters if that fails.

The gate is asymmetric, and the asymmetry is the point:

- Zero cases, across the conformance corpus and the shadow period,
  where beets reaches the correct verdict and Go reaches an incorrect
  one. That count is the gate, not a percentage.
- Divergences where Go is defensibly equal or better are adjudicated
  case by case and recorded as known-acceptable divergences with the
  reasoning written down. The record is an artifact of the harness, not
  a note in a commit message.

**A false confident match is the unforgivable failure class.** A wrong
reject or a spurious ask costs an admin some attention. A confident
match against the wrong release writes wrong tags, moves files to a
wrong path, and tells the contributor the album landed. That corrupts
custody, the one thing the server exists to hold.

If the bar cannot be cleared, the honest outcome is staying on the
beets subprocess indefinitely. Staged adoption makes that survivable:
beets stays behind the `internal/beets` seam until the Go matcher earns
the swap, so a failed matcher costs the harness and matcher rungs, not
the estate.

## What the initiative builds

dubplate-organizer extends dubplate into a batteries-included music
library organizer that runs two ways on one shared engine: hosted as
the estate's server, or locally as a terminal application. The entry
point is the beets replacement: a native Go reimplementation of beets'
MusicBrainz autotagging matcher, held to the bar above before it
replaces the subprocess, plus a built-in set covering the work the
estate's most-used beets plugins do today.

**The server rungs deliver first**, because the estate has immediate
use for them and the local organizer has none yet. Every rung through
the swap improves the running box; the local organizer follows. One
exception is deliberate: the admin API spine may advance ahead of the
two-binary split if remote triage proves wanted, because the spine is
itself a server deliverable.

## The server's purview

`dubplate-server` owns custody, organization, record, and intake of one
shared library for one group: intake and admission, the organizing
engine, the library store, the review queue, backup and restore,
notifications, the admin API, and a documented client API.

The server integrates with, and never owns, five adjacent surfaces.
Each is a narrow seam, optional or swappable. The required-no-defaults
config wall applies to what the server owns; integrations get explicit
enable flags.

- **Streaming**: a streamer-driver seam (scan trigger, scan status,
  album deep link, driver capability reporting). The seam ships with
  the Navidrome reference driver only. Jellyfin and generic Subsonic
  are supported by design and each earns its build on demand: classic
  Subsonic has no album deep link and no portable scan verbs, and
  Jellyfin's asynchronous `/Library/Refresh` has no scan-status
  equivalent, so a partial driver reports the part it satisfies.
- **Auth**: a verify-identity seam. The server verifies an identity
  assertion; how a client obtains it is the auth provider's business.
  Cloudflare Access is the reference implementation. The seam supports
  non-browser token acquisition, so native clients authenticate without
  a cookie jar.
- **Storage**: backup targets are rclone remotes, as today.
- **Monitoring**: the pings dead-man's-switch wiring stays optional.
- **Web UI**: the embedded page is the reference client, disableable,
  and replaceable by anything that speaks the documented API.

With no streamer configured, custody and organization function
normally: files are admitted, tagged, and laid out, and the record is
complete. The journey terminates at **"imported, no player
configured"**. That state is honest copy, and it is a vocabulary
amendment to the uploader spec's ratified journey states (uploading,
checking, matching, importing, in-library, pushed-back, queued,
stalled). The streamer-seam work owns filing that amendment, and it
owns the config-wall change that today makes the Navidrome environment
variables required with no default: those variables become driver
config.

## Client-agnostic API

The HTTP API is the product surface; every user interface is a client
of it.

**Audience.** One group's estate plus its own clients. There is no
third-party developer audience today and this spec does not fund one.
The world-facing property the API claims is narrower and real:
**nothing in the contract assumes a particular client.** Token clients
get pure bearer semantics with CSRF defenses scoped to cookie-carrying
browsers, errors are machine-readable, and no verb needs a browser to
call it.

**Versioning.** The contract document, URL versioning (`/api/v1`), an
additive-within-version compatibility policy, and the authorization
model below are named deliverables of the admin-API rung. Two
mechanisms a reader will otherwise conflate: *internal parity* is
lockstep, because `ops/local` and `ops/client` implement one Go
interface at one version in one repo, released together and enforced by
the contract-test suite; *external compatibility* is
additive-within-version, so a published `/api/v1` verb keeps its shape,
new fields and verbs may appear, and removals or semantic changes
require `/api/v2`. The two do not fight, because the wire types are a
published projection of the Go interface rather than the interface
itself.

**Authorization.** Admin is an `AdminEmail` match, with no role system.
Admin verbs accept header-only identity assertions and reject
cookie-derived identity, which preserves ROADMAP's wall against admin
verbs reachable from the browser origin. Service tokens authenticate
machines only and are limited to unattended server-to-server verbs;
they carry no email, so any contributor-attributed operation requires a
user assertion that does.

**Copy in payloads.** The `internal/reject` copy contract's
`friendly_reason` strings travel in payloads and are the default
rendering, so every client shows the same kind rejection language.
Clients may render their own. Copy revisions are additive and never
contract-breaking. Localization is out of scope.

**Transport.** tus keeps upload transport standard, and tus clients
exist for iOS (TUSKit) and Android (tus-android-client), so nothing in
the transport precludes a native client. Transport is the easy part.
The hard parts are non-browser token acquisition, background upload
continuation, and a reference auth implementation issuing credentials
an outside adopter does not have. None is solved here, and no mobile
client is a deliverable of this initiative.

## The two binaries

One repo, one engine, two binaries. Each binary owns its requirement
wall outright; shared subcommands are shared cobra constructors, not
duplicates.

`dubplate` is the local organizer. It requires a music directory and
its own state directory, nothing else; web, tusd, JWT, tunnel, and
Access are absent from its dependency graph. The bare command launches
the bubbletea TUI: browse the library, import with interactive
ask-resolution, triage duplicates, run hygiene tools. Scriptable
subcommands: `import` (with an unattended mode that quarantines asks
instead of prompting), `publish`, `backup`, `restore`, `doctor`,
`check`, `version`. With a remote target configured, the same TUI
drives a server's admin API.

`dubplate-server` is today's estate binary renamed, keeping the full
required-no-defaults wall: `serve`, `import` (poll mode and
`--session`), `nudge-consume`, `backup`, `restore`, `check`, `doctor`,
`ping-fail`, `version`, and `beet` passthrough until the repair verbs
replace it. The rename is a funded deliverable of the split rung, not a
clause: it touches the systemd units, `scripts/deploy.sh`, the `/etc`
and `/opt` paths, the pings check slugs, the runbook, and STATUS, and
it ships with a box-write plan. The box installs only
`dubplate-server`; the local binary on the box is unsupported, which is
what keeps the name `dubplate` from shadowing the server's historical
identity there.

**Config.** `internal/config` stays the one home for defaults, per
CLAUDE.md. The binaries differ in which surfaces are *required*, never
in where defaults live: the server requires everything it owns, the
local binary requires a music directory and a state directory.
Engine-level settings (path formats, collection routing, match
thresholds) resolve identically in both.

**Publish and custody.** Sync between rungs is publish-only in this
initiative. The local organizer publishes a finished album to a server
through the same validated intake API as every other client. Publish is
copy, not move, with an optional prune, so a laptop-to-box workflow
never silently deletes the local copy and never silently doubles disk
without saying so. The local store records publish provenance (target
server, session ID, server outcome), which makes a re-publish
detectable locally rather than leaving the server's duplicate detection
as the only guard, and which gives a server push-back of a locally
finished album a defined local record: the publish outcome.

The custody claim is therefore stated precisely: **contributions enter
through the one gate.** `restore` is the admin bulk path and sits
outside that claim explicitly. Bulk migration between rungs remains
backup and restore; publish is the per-album path. A server-to-local
read-only mirror is out of scope and designs its own schema when it
earns its build, so this spec adds no column for it. True two-way sync
is out.

## Shared code: the ops seam

The operation surface is defined once; local and remote are two
implementations of that one interface.

- `engine`: matcher, tag I/O, store, pipeline, built-ins. A pure
  library that knows nothing about HTTP, terminals, or which binary
  hosts it.
- `ops`: the operation vocabulary. A Go interface plus request and
  response types, JSON-serializable and the source of the API's wire
  types; one vocabulary, not two.
- `ops/local`: calls the engine in-process. Bare `dubplate` binds it.
- `ops/client`: the same interface over HTTP against
  `dubplate-server`. Remote-TUI mode and publish both use it.
- `serve`: the interface as HTTP handlers, marshaling the same types
  `ops/client` unmarshals.
- `tui`: consumes only the `ops` interface.

**The job model exists from birth.** Long-running operations (import,
publish) return a job handle. Progress is polled, per the uploader's
settled ruling about cloudflared's event-stream buffering and idle cut;
a local implementation may layer push on the same abstraction.
Cancellation is an explicit verb, not a reliance on `ctx` unwinding
in-process. It is settled before the interface exists because
retrofitting it changes the signature of the interface's most important
operation.

**Ask resolution is two named operations**, not one with an asterisk:

1. *The local interactive session.* The TUI holds an import mid-flight,
   presents candidates, takes a choice or a manual MBID or an edit,
   recomputes, repeats.
2. *The decision queue.* `list pending`, `answer`, and `defer`, backed
   by persisted decision records with candidates, distances, and the
   penalty breakdown serialized in the store. This is the server-side
   equivalent of beets' interactive chooser: asynchronous, persisted,
   resumable. The remote TUI renders beets' chooser UX against it.

Contributors never get a held session. The uploader's ratified
push-back flow is their equivalent, and it stays that way.

**Capability asymmetries are named in this spec or a successor spec,
never declared ad hoc in code.** That rule is what keeps capability
reporting from becoming an escape hatch any future operation can use to
opt out of parity. Named today: filesystem browsing is local-only, and
the two ask-resolution operations above are distinct rather than one
operation with two flavors.

Parity is enforced, not hoped for: one contract-test suite runs against
both implementations, `ops/local` directly and `ops/client` against an
in-test server, so an operation added to one side fails the suite until
the other side has it. The web client shares contract rather than code:
TypeScript consuming the same wire types and in-payload copy strings,
with TS type generation from the Go contract available if drift bites.

**Two capability mechanisms, two names.** *Streamer driver
capabilities* report what a configured streamer supports and belong to
the streamer seam. *Ops operation availability* reports which
operations an implementation offers and belongs to the ops seam. Each
is specified by the rung that builds its surface. They never share a
term.

## Uncertain-match custody policy

Who resolves an upload the matcher is not confident about is a policy,
not the architecture. The engine carries the spectrum; config picks the
point on it. The survey of comparable tools (full findings:
`docs/superpowers/research/2026-09-01-uncertain-match-ux.md`) found
that no multi-user system grants non-admin contributors match-correction
rights: Plex locks Fix Match to the server admin by explicit policy,
Jellyfin has no metadata-edit permission below admin, and the tools
with real interactive pickers (Lidarr, Roon, Picard, beets itself) are
single-owner software. The estate's ratified refusal flow is the field
consensus, not an outlier.

The policy enum, with the machinery shared across all three settings
(the matcher's ranked candidates and the decision-queue records already
exist):

- `strict` (the shipped default, and the estate's setting): push-back
  with the ratified kind copy and escapes; uncertain material never
  enters the main library on a contributor's judgment.
- `owner-queue`: uncertain uploads park as decision records in the
  admin queue instead of pushing back; the contributor sees "waiting
  for a decision" honesty.
- `contributor-picks`: the push-back page offers the matcher's own
  top candidates as a bounded choice. A documented opt-in for
  high-trust groups; never the default, because no surveyed multi-user
  tool trusts contributors with correction and the survey's operator
  evidence runs the same way.

One targeted remedy rides every setting. Classical, compilation, and
box-set material mismatches structurally: MetaBrainz's forums document
the coverage gaps, and beets has open false-negative issues on
classical releases even with correct embedded MusicBrainz IDs. For
that class, refusal is a genre tax on honest contributions. The
push-back page therefore gains an appeal action ("I think these tags
are right") that routes the album to the owner decision queue instead
of forcing the private-lane demotion. Decision rights stay with the
owner; the contributor gains an appeal, not authority. The appeal
action amends the uploader's ratified push-back surface, and the
decision-queue rung owns filing that amendment alongside its build.

## Extensibility

The fork is settled: **no in-process plugin architecture, and no plugin
ABI.** Three extension surfaces cover the need.

- **Built-ins** for the proven core, compiled in.
- **Lifecycle hooks**: configured external commands at named pipeline
  events (post-import, on-reject, on-review, post-scan). The hook
  receives a JSON payload on stdin sharing the ops wire-type
  vocabulary. Hooks are optional and degrade gracefully; an absent or
  failing hook never fails an import.
- **The documented client API** for anything tool-shaped.

Revisit a plugin system only on demonstrated demand. The hook seam
pre-establishes the process boundary such a system would need, so
deferring costs nothing structural.

### Selecting the built-ins

The set derives from the beets plugin adoption survey
(`2026-08-31-beets-plugin-usage.md` in the dotfiles research directory,
n=87 published configs). The survey's ranking is the signal, and its
sampling caveat travels with it: dotfiles publishers are power users,
so absolute counts overstate the casual user. **The exact list is
re-derived at the built-ins rung's planning**, against the survey and
against what the estate actually uses. This spec fixes the policy, not
the roster. Two policies fix now:

- **Scrapers are refused on API fragility and maintenance grounds.**
  This is a policy judgment, not a popularity threshold, and the survey
  does not support reading it as one: `lyrics`, the canonical scraper,
  ranks in the survey's top tier.
- **Fingerprinting and ReplayGain are probe-backed, not in the box.**
  They are present when chromaprint and ffmpeg are installed and absent
  otherwise, like today's quality gate. Conformance is proven for the
  no-chromaprint configuration; a fingerprint-assisted corpus stratum
  is future work.

## Beauty and cleanliness

Design quality is an architectural property of this initiative, not a
polish pass at the end. The spec separates what a gate can fail from
what a person judges.

**Gated properties**, each with a check that can go red: Go per
go-conventions; TUI work under elm-conventions and bubbletea-design,
with no TUI screen claimed working without the tui-visual-verify gate;
web work under svelte-conventions and ts-conventions with the
DaisyUI-and-runes idiom mandate; uploader UI work under the
visual-fidelity chain; the linters and `scripts/check.sh` on every
commit; the reviewer fan-out at pass end.

**Judgment properties**, held by review rather than by a gate: small
single-purpose packages with narrow seams (the ops interface is the
pattern); a TUI that starts from poplar's design language (the
bubbletea email client at `~/Projects/poplar`) rather than reinventing
a theme system, spacing discipline, and interaction idiom; and a web
reference client keeping the uploader's ratified design language
(Navidrome's dark ground, Material blue, uppercase actions, flat
cards).

**The adjacency rule is bounded.** Un-idiomatic code adjacent to a
change is in scope only when it sits in the same file or the same
package, or when a reviewer nominates it. Anything larger becomes a
filed follow-up, so the mandate cannot become the
accretion-by-adjacency failure mode the pass-sizing rule names.

## Settled forks

- **Native dependencies**: pure-Go core, no cgo. chromaprint and
  ffmpeg remain optional subprocess probes that degrade gracefully.
- **Tag I/O is a named subsystem with its own ladder rung**, ahead of
  the pipeline port. Writing tags is the pipeline's primary side
  effect and Go has no equivalent of beets' MediaFile layer, so this is
  not an implementation detail of a later rung. Server scope is
  FLAC-only, matching today's gate policy (`AllowedExt: [".flac"]`), so
  the swap needs pure-Go FLAC tag writing and nothing else.
  Mixed-format support is required only by the later local organizer,
  whose laptop libraries are not FLAC-only. Its named candidate is a
  WASM-compiled taglib driven through wazero, preserving both the
  no-cgo rule and the static binary, decided by a spike inside the
  tag-I/O rung.
- **Library store**: SQLite through the pure-Go modernc driver. The
  schema is designed at the swap rung; beets' relational shape is an
  input to that design, not a commitment. A migrations mechanism ships
  with the store from its first version. The single-file snapshot
  backup mechanism stays identical; its contents change.
- **Migration**: a one-shot converter reads the live beets database
  into the Go store. The beets database is archived in R2, never
  deleted.
- **Adoption**: beets stays behind the `internal/beets` subprocess seam
  until the Go matcher clears the success bar, then swaps at that one
  seam.

### Migrator criteria

The converter is specified by its failure modes, which are the parts
that silently corrupt:

- beets stores `items.path` and `albums.artpath` as BLOB, not TEXT,
  precisely so non-UTF8 filenames survive. Never round-trip a path
  through a Go string. The corpus deliberately includes non-Latin
  material, so this is real estate content, not an edge case.
- Singletons carry `album_id` NULL, not 0.
- beets has no import-history table. What survives is current library
  state plus `added` timestamps. The estate's per-session history lives
  in `internal/ledger`, whose `LibraryPaths` point at beets-database
  paths, so the ledger must be reconciled or every contributor's
  journey history breaks its links at the swap.
- **The swap requires a no-move assertion.** The on-disk layout is the
  real artifact. A dry-run diff proves the Go pipeline's computed paths
  equal the on-disk layout before anything writes. A one-character
  difference from `config/beets-config.yaml`'s query-prefixed path
  formats re-lays-out the entire library on the first post-swap run.

### Rollback

At the swap, snapshot the beets database to R2. beets stays installed
through a deprecation window. Reverting means restoring that snapshot
and re-importing post-swap sessions from the ledger's session records.
The files are on disk and the session records name them, so the
re-import is mechanical.

## Conformance harness

The harness proves the Go matcher reproduces beets' behavior before the
seam swaps, using a frozen JSON fixture corpus (full findings and
sources:
`docs/superpowers/research/2026-09-01-beets-conformance-corpus.md`).

**A case is a quadruple: {matcher input, frozen MusicBrainz responses,
beets config, expected verdict}.** Config is part of the case because a
beets verdict is a pure function of it: `strong_rec_thresh`,
`medium_rec_thresh`, `rec_gap_thresh`, per-penalty `max_rec` overrides,
the `distance_weights` table, and `preferred`, `ignored`, `required`,
and `ignored_media` each move verdicts on unchanged input. The primary
oracle config is the estate's `config/beets-config.yaml`, because the
estate is what the swap risks. A stock-config stratum is secondary, and
it is what says anything about a stranger adopter.

Strata, sized as **initial estimates from the research report and
refined at the harness rung**: roughly 150 to 300 cases, about 40%
structural, 50% real catalog, 10% adversarial.

- Structural: input shapes ported from beets' own autotag tests (track
  scrambling, missing and extra tracks, multi-disc, various-artists,
  non-Latin scripts), with expected verdicts derived by running real
  beets, never copied. beets' fixtures are inline Python object
  construction, so the portable asset is the scenario shapes. **Hand
  porting them is the harness rung's largest manual cost**, and the
  rung's plan sizes it explicitly.
- Real catalog: releases captured once from the MusicBrainz web
  service, covering reissues, remasters, near-duplicate pressings,
  partial rips, and mistagged inputs.
- Adversarial: should-reject and should-ask cases that prove the
  boundaries, not just confident matches.

**Acceptance.** Verdict agreement is necessary and not sufficient,
because verdict is a thresholded function of a continuous distance and
two implementations can agree on a corpus while still diverging the
first time real material sits near a threshold. Acceptance therefore
includes numeric distance agreement within a stated epsilon and stable
candidate ordering: the Go matcher preserves MusicBrainz candidate
iteration order and sorts stably, as beets does. Ordering matters
beyond the harness, because the uploader's ratified synchronous match
probe shows the contributor what it matched, so the seam already
carries a candidate-and-confidence contract and the presented list is
part of what conformance covers.

**Mechanics: two jobs on two cadences.** Frozen expected verdicts run
in the everyday fast gate, with no Python in the Go gate. The live
beets differential runs as a separate job on beets version bumps only,
driving pinned beets as a subprocess against a local stub server. The
two freshness axes move independently, which is the only way they
separate "we differ from beets" from "beets changed": a version bump
runs new beets against unchanged responses, and a periodic re-capture
runs unchanged beets against new responses. Cassette tooling: adopt
go-vcr unless the corpus's fixture shape demands otherwise. (The
earlier reasoning here was a non sequitur. beets lacking cassette
infrastructure has no bearing on whether a Go library has it.)

**The oracle has an exit.** At the swap, beets ceases to be the oracle
and the corpus becomes dubplate's own regression suite. Divergences
from newer beets after the swap are adjudicated under the success bar
rather than followed automatically, because conformance to a specific
beets version's quirks stops being a goal once dubplate owns the
behavior.

**MusicBrainz client**, a rung-2 requirement rather than an
implementation detail: a registered application identity and a proper
User-Agent, since the anonymous limit is 1 request per second and
identified clients get more; and a shared rate limiter with per-caller
budgets where a user-present probe outranks a batch import, because the
uploader's ratified journey holds a contributor on the page through
those lookups while timer imports and other contributors share the same
limiter.

## Build order

Rungs are sequential, with one executor, per the workstation's
one-executor-per-worktree rule. There is no rung overlap. Rungs 1
through 4 are internal milestones that retire estimation risk: they
change no user-visible behavior and the estate keeps running beets
throughout. The first user-visible change is rung 5, which is also the
irreversible one, which is why rungs 3 and 4 sit in front of it.

Magnitude bands are relative to each other, carry no time units, and
are refined by each rung's own plan. Every rung's plan carries a token
ceiling per workstation discipline.

1. **Corpus and harness** (L). The quadruple corpus, the frozen-verdict
   fast gate, the version-bump differential job, the capture tool.
2. **Go matcher** (L). The MusicBrainz client above, the distance and
   penalty model, album-level matching, and **beets' full tunable
   matching surface**: thresholds, distance weights, preferred,
   ignored, required. Behind the harness until the bar holds.
3. **Tag I/O** (M). Pure-Go FLAC tag writing for the server scope, plus
   the wazero-taglib spike that decides the mixed-format path the local
   organizer needs later.
4. **Shadow on the box** (S). The Go matcher runs in shadow against
   real uploads on the live box, with beets still deciding. Both
   verdicts are logged and compared under the estate config. This is
   the only evidence from real contributor material rather than a
   curated corpus, and it costs nothing irreversible.
5. **Store, migrator, pipeline port, and the swap** (L). The
   initiative's point of no return. Gate: the success bar's asymmetric
   criterion over the corpus **and** a clean shadow period. Beyond the
   port itself, this rung delivers the schema migrations mechanism, the
   migrator against its named failure modes, the no-move dry-run
   assertion, a rollback procedure exercised rather than merely
   written, and three things worth naming:
   - *Repair verbs*: query, retag against a named MusicBrainz release,
     and move, in the admin library surface. They exist because
     retiring `beet` passthrough deletes the admin's entire ad-hoc
     repair toolkit. **The passthrough retires only when they exist**,
     never before.
   - *A concurrency model*: the estate's single import lock is the
     current model, and the ops seam multiplies writers against it. The
     candidates are a per-album lease or an explicit queue with a
     documented starvation policy. The decision belongs to this rung's
     spec and is flagged here rather than settled. Binding constraint:
     an interactive local session must not starve timer imports.
   - *Engine extraction*: notification, roster, and monitoring become
     optional engine seams here, while the pipeline is being rewritten
     anyway. Today's `internal/importer` mails, reads the roster, and
     pings a dead-man's switch, none of which means anything on a
     laptop with one user, and unpicking it later is strictly more
     expensive.
6. **Built-ins and lifecycle hooks** (M). The built-in set re-derived
   per the policy above, plus the hook seam.
7. **Two-binary split and the local TUI** (L). Includes the funded
   `dubplate-server` rename with its box-write plan, the ops-seam
   contract tests, and publish.
8. **Remote TUI and the admin API spine** (M). The contract document,
   `/api/v1`, the compatibility policy, and the authorization model.
   **The spine may advance ahead of rung 7** if remote triage proves
   wanted, because it is a server deliverable and the server delivers
   first. When it advances, it carries a thin read-only ops slice
   (`history`, `doctor`) that proves the ops seam end to end before the
   split designs around it, and the remote TUI follows the spine.

The remote TUI's first job is triaging the server-owned review queue
plus operational verbs (`doctor`, `history`). **The review queue is
defined once, here**: it is the admin-only side door for timer-path
material, per the uploader addendum that abolished contributor-facing
review. It is a low-volume admin backstop, not a contributor surface,
and it is not the decision queue described in the ops seam. The demand
argument for the remote TUI is retiring ssh as the admin interface, not
review volume.

Not yet assigned to a rung: the streamer-driver seam extraction itself,
with its config-wall change and its journey vocabulary amendment. It is
a server deliverable and gets its rung when it is scheduled.

## Roadmap revisions this spec carries

Fold into ROADMAP.md at the next pass close (the file carries unrelated
in-flight edits right now):

- Streamer integration ships with the Navidrome reference driver only;
  Jellyfin and Subsonic drivers are supported by design and each earns
  its build on demand. Navidrome moves from "the bonus only dubplate
  can offer" to "reference driver, richest integration".
- That demotion runs against ROADMAP's product-essence line and
  CLAUDE.md's identical framing, both of which read "a shared Navidrome
  server". Reconcile the essence wording in the same edit.
- The API carries no client assumptions by requirement; the audience is
  one group's estate plus its own clients.
- Sync between rungs is publish-only this initiative, mirror out of
  scope, two-way ruled out.
- Publish and backup/restore split: publish is the per-album path
  through intake, backup and restore is the bulk and admin path. This
  supersedes reading "migration between rungs is backup/restore" as the
  only path while leaving it true for bulk.
- Auth generalizes from "Access-authenticated via the cloudflared
  access token flow" to a verify-identity seam with Access as the
  reference implementation.

## Out of scope

- The web admin page stays deferred. The API spine keeps it cheap, and
  it earns its build only if phone triage proves wanted.
- Multi-tenancy. One library, one group is the product essence.
- In-process plugins and a plugin ABI, per the extensibility fork.
- Scrapers, on maintenance grounds.
- True two-way sync, and the server-to-local mirror.
- Localization of the copy contract.
- Native mobile clients as a deliverable of this initiative.

Packaging remains sequenced after this initiative per ROADMAP, so it
templates the final two-binary shape. This spec carries that as an
inherited input and does not re-argue it.
