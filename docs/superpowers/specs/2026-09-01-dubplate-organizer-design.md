# dubplate-organizer: initiative design

Status: high-level design, ratified in brainstorm 2026-09-01 (Geoff +
Claude). This spec sets the initiative's architecture and build order.
Each ladder rung gets its own spec and plan when its turn comes; none
exists yet. Preconditions: the uploader initiative ships and the
prod-polish pass completes.

## What the initiative builds

dubplate-organizer extends dubplate into a batteries-included music
library organizer that runs two ways on one shared engine: locally as a
terminal application, or hosted as the estate's server. The entry point
is the beets replacement: a native Go reimplementation of beets'
MusicBrainz autotagging matcher, proven behaviorally equivalent before
it replaces the subprocess, plus the empirically top beets plugins as
built-ins (art fetch/embed, genre, duplicates, tag hygiene,
fingerprinting, ReplayGain).

## The server's purview

`dubplate-server` owns custody, organization, record, and intake of one
shared library for one group. That means: intake and admission, the
organizing engine, the library store, the review queue, backup and
restore, notifications, the admin API, and a documented client API.

The server integrates with, and never owns, five adjacent surfaces.
Each is a narrow seam, optional or swappable; the required-no-defaults
config wall applies to what the server owns, while integrations get
explicit enable flags.

- Streaming: a streamer-driver seam (scan trigger, scan status, album
  deep link, capability discovery). Navidrome, Jellyfin, and generic
  Subsonic are equal peers; Navidrome is the reference driver with the
  richest integration. The server is fully functional with no streamer
  configured, degrading the journey's final step to "imported, scan
  pending" honesty.
- Auth: a verify-identity seam. The server verifies an identity
  assertion; how a client obtains it is the auth provider's business.
  Cloudflare Access is the reference implementation. The seam must
  support non-browser token acquisition (device-login and service
  tokens) so native clients authenticate without a cookie jar.
- Storage: backup targets are rclone remotes, as today.
- Monitoring: the pings dead-man's-switch wiring stays optional.
- Web UI: the embedded page is the reference client, disableable, and
  replaceable by anything that speaks the documented API.

## Client-agnostic API

The HTTP API is the product surface; every user interface is a client
of it. The contract (session creation, tus upload, journey, history,
config, admin verbs) is documented and versioned. It carries no browser
assumptions: token clients get pure bearer semantics with CSRF defenses
scoped to cookie-carrying browsers, errors are machine-readable, and
the copy contract's `friendly_reason` strings travel in payloads so
every client renders the same kind rejection language. tus keeps upload
transport standard; mature tus clients exist for iOS (TUSKit) and
Android (tus-android-client), so native mobile apps are first-class
consumers, not ports.

The client roster the API serves as equals: the embedded web page, the
local TUI (publish and remote admin), third-party web UIs, and native
mobile apps.

## The two binaries

One repo, one engine, two binaries. Each binary owns its requirement
wall outright; shared subcommands are shared cobra constructors, not
duplicates.

`dubplate` is the local organizer. It requires a music directory and
its own state directory, nothing else; web, tusd, JWT, tunnel, and
Access are absent from its dependency graph, and its config philosophy
is minimal with sane defaults. The bare command launches the bubbletea
TUI: browse the library, import with interactive ask-resolution, triage
duplicates, run hygiene tools. Scriptable subcommands: `import` (with
an unattended mode that quarantines asks instead of prompting),
`backup`, `restore`, `doctor`, `check`, `version`. With a remote target
configured, the same TUI drives a server's admin API; review-queue
triage is the first remote job.

`dubplate-server` is today's estate binary renamed, keeping the full
required-no-defaults wall: `serve`, `import` (poll mode and
`--session`), `nudge-consume`, `backup`, `restore`, `check`, `doctor`,
`ping-fail`, `version`, and `beet` passthrough until the seam swaps.

Sync between the rungs is publish-only in this initiative: the local
organizer publishes a finished album to a server through the same
validated intake API as every other client, so the server's custody
guarantees hold because everything enters through the one gate. A
server-to-local read-only mirror is designed for but deferred: the
store schema keeps a remote-origin notion so a mirror can slot in
later without migration. True two-way sync is out of scope.

## Shared code: the ops seam

The operation surface is defined once, and local and remote are two
implementations of that one interface.

- `engine`: matcher, store, pipeline, plugins. A pure library that
  knows nothing about HTTP, terminals, or which binary hosts it.
- `ops`: the operation vocabulary. A Go interface (import, resolve
  ask, history, review queue, doctor) plus request and response types.
  The types are JSON-serializable and are the API's wire types; one
  vocabulary, not two.
- `ops/local`: implements the interface by calling the engine
  in-process. Bare `dubplate` binds this.
- `ops/client`: implements the same interface over HTTP against
  `dubplate-server`. Remote-TUI mode and publish both use it.
- `serve`: exposes the interface as HTTP handlers, marshaling the same
  types `ops/client` unmarshals.
- `tui`: consumes only the `ops` interface and cannot tell local from
  remote.

Parity is enforced, not hoped for: one contract-test suite runs against
both implementations, `ops/local` directly and `ops/client` against an
in-test server, so an operation added to one side fails the suite until
the other side has it. Asymmetries are declared, not discovered: the
interface reports capabilities (interactive ask-resolution is local;
the review queue represents it remotely; filesystem browsing is
local-only), and the TUI renders what is available. The web client
shares contract rather than code: TypeScript consuming the same wire
types and in-payload copy strings, with TS type generation from the Go
contract available as an option if drift bites.

## Beauty and cleanliness as architecture

Design quality is a core architectural property of this initiative, not
a polish pass at the end. Three commitments carry it.

The TUI starts from poplar's design language (the bubbletea email
client at `~/Projects/poplar`): its theme system, spacing and glyph
discipline, and interaction idiom are the reference, adapted rather
than reinvented. TUI work runs under the elm-conventions and
bubbletea-design skills, and no TUI screen is claimed working without
the tui-visual-verify gate.

The web reference client keeps the uploader's ratified design language
(Navidrome's dark ground, Material blue, uppercase actions, flat
cards) and the standing idiomatic-stack mandate: DaisyUI component
classes and theme tokens over hand-rolled utilities, Svelte 5 runes
idiom, with the cairn UI work as the exemplar.

Cleanliness binds the code as well as the pixels: small single-purpose
packages with narrow seams (the ops interface is the pattern), Go per
go-conventions throughout, and un-idiomatic code adjacent to a change
is in scope for that change. A rung that ships working but unclean has
not shipped.

## Settled forks

- Native dependencies: pure-Go core; chromaprint and ffmpeg remain
  optional subprocess probes that degrade gracefully when absent, like
  today's quality gate. No cgo.
- Library store: SQLite through the pure-Go modernc driver. The
  relational model beets proved out survives (albums, tracks, flexible
  attributes such as collection), the single-file snapshot backup
  stays identical, and the binary stays static.
- Migration: a one-shot converter reads the live beets database
  (items, albums, flexible attributes) into the Go store, so import
  history and collection routing survive the swap. The beets database
  is archived in R2, never deleted.
- Adoption: beets stays behind the `internal/beets` subprocess seam
  until the harvested matcher proves conformance, then swaps at that
  one seam.

## Conformance harness

The harness proves the Go matcher reproduces beets' behavior (verdict,
chosen release, distance ranking) before the seam swaps, using a frozen
JSON fixture corpus of roughly 150 to 300 cases in three strata (full
findings and sources:
`docs/superpowers/research/2026-09-01-beets-conformance-corpus.md`):

- Structural, about 40%: input shapes ported from beets' own autotag
  tests (track scrambling, missing and extra tracks, multi-disc,
  various-artists, non-Latin scripts), with expected verdicts derived
  by running real beets, not copied. beets' fixtures are inline Python
  object construction, so the portable asset is the scenario shapes,
  not files.
- Real catalog, about 50%: releases captured once from the MusicBrainz
  web service, covering reissues, remasters, near-duplicate pressings,
  partial rips, and mistagged inputs.
- Adversarial, about 10%: should-reject and should-ask cases that
  prove the boundaries, not just confident matches.

Each case is a self-contained triple: matcher input, frozen
MusicBrainz response, expected beets verdict. A differential runner
drives pinned-version beets as a subprocess and the Go matcher against
the same frozen responses through a local stub server; CI never touches
the live network. beets has no HTTP-cassette infrastructure, so the
project owns its own minimal cassette format (go-vcr style). Freshness
comes from re-capturing a rotating subset quarterly and on beets
version bumps, which separates "we differ from beets" from "beets
changed". MusicBrainz core data is CC0 and the anonymous rate limit is
1 request per second, so capture is cheap and clean.

## Build order

The ladder is risk-ordered; each rung is shippable and the estate keeps
running beets until rung 3 completes.

1. Corpus and harness. Differential rig with beets as the oracle. No
   behavior change.
2. Go matcher. MusicBrainz client (rate-limited, cached), the distance
   and penalty model, album-level matching, behind the harness until
   conformance holds.
3. Store, migrator, pipeline port, and the one-seam swap on the box.
   The initiative's point of no return, taken only on rung 2's proof.
4. Plugin built-ins: art, genre, duplicates, tag hygiene in pure Go;
   fingerprinting and ReplayGain as optional probes.
5. Two-binary split and the local TUI, including publish and the
   ops-seam contract tests.
6. Remote TUI, admin API spine, and the Jellyfin and Subsonic drivers.

Rungs 5 and 6 touch different surfaces than rung 4 and can overlap
with it once the seam has swapped.

## Roadmap revisions this spec carries

Fold into ROADMAP.md at the next pass close (the file carries
unrelated in-flight edits right now):

- Streamer integration: Navidrome, Jellyfin, and generic Subsonic as
  equal peers behind the driver seam; Navidrome moves from "the bonus
  only dubplate can offer" to "reference driver, richest integration".
- The API is client-agnostic by requirement; native mobile clients are
  named consumers.
- Rung-to-rung sync is publish-only this initiative, mirror deferred,
  two-way ruled out.

## Out of scope

The web admin page stays deferred (the API spine keeps it cheap; it
earns its build only if phone triage proves wanted). Multi-tenancy is
out: one library, one group is the product essence. Scrapers and the
sub-10% beets plugin tail stay refused. True two-way sync is out, as
above.
