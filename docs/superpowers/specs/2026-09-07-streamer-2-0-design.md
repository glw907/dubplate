# dubplate-server 2.0: the built-in streamer

Status: **DRAFT, brainstorm settled.** Revision 1, 2026-09-07. Every
decision below was taken with Geoff in the 2026-09-07 brainstorm; the
open items are listed in their own section and none blocks the study
rung. ROADMAP entry: `streamer-2-0`, sequenced after the greenfield
ladder's launch (rung 17).

This document governs the initiative's purview, its architecture, and
the forward hooks the current build owes it. It does not amend the
greenfield build spec
(`docs/superpowers/specs/2026-09-02-greenfield-server-build.md`), the
launch bar, or the 17-rung ladder. Where it names a current rung
(5, 6, 7, 8, 12, 16), it asks that rung's plan to weigh a hook, never
to build a 2.0 feature early.

## Purpose

dubplate-server 2.0 subsumes Navidrome. After it, the estate is one
binary a stranger can set up: intake, tagging, the catalog, and the
streaming server that the family's apps connect to. The primary driver
is that single integrated system (Geoff, 2026-09-07). Owning the
listening experience and the single source of truth follow from it
rather than motivating it.

Navidrome's one missing feature is intake and tagging, which dubplate
already owns. 2.0 adds the half Navidrome does well, streaming to
OpenSubsonic clients, on top of the catalog the pipeline already
maintains.

## Decisions

Each decision names its alternatives and the reason the alternative
lost, so a later reader does not reopen it without new evidence.

### Client surface: OpenSubsonic API first

2.0 launches with the OpenSubsonic API and no first-party player. The
API is what makes phones, cars, and desktop apps work, and it is a
specified contract the harness can test the way rung 3 tested the
matcher. A web player follows on the 2.1 list, built on the same API,
so the browser is one more client rather than a privileged one.
Rejected: web player first (the family listens in apps), both at launch
(largest scope with no proportional gain).

### Index: the catalog is the index

The store rung 6 builds, filled from the beets database at rung 16, is
the only catalog. The streaming index is a projection of it. No scanner
admits music. A reconcile walk runs on a timer and reports drift (a
file missing, a hash that no longer matches) as doctor rows and an
admin notice; it never admits or removes anything. Intake stays the one
door. Consequence, accepted: "point 2.0 at a folder and play" is not a
feature; a pre-existing library enters through the pipeline, as rung
16's migration already does for the box. Rejected: a scanner beside the
catalog (two truths, the thing 2.0 exists to remove), scanner only with
the catalog as overlay (Navidrome's architecture with its friction).

### Identity: app passwords minted from the roster account

The roster stays the only account store. A user signs in by magic link
(rung 5) and mints a named app password per device from a device page,
shown once and revocable alone. One stored secret serves both the
legacy Subsonic salted-token scheme and the OpenSubsonic API-key
scheme. Rejected: one Subsonic password per user (a lost phone rotates
every device), API keys only (client support is uneven).

### Transcoding: ffmpeg under a content-keyed cache

ffmpeg runs as a managed subprocess under `exec.CommandContext`. A
content-addressed cache keyed on (track hash, profile) sits between the
transcoder and the client, filled on first play, never pre-warmed at
launch, evicted least-recently-played under a configured size cap.
Geoff chose the cache on the numbers: a full cache of one profile costs
about a tenth of the library's disk, and the pipeline's content hash
makes invalidation a delete on retag rather than a heuristic. Rejected:
no cache (repeat plays and seeks pay the encode every time), direct
streaming only (lossless on cellular is unusable).

### Per-user state at launch

Starred, playlists, play history, and the saved queue ship at launch.
Play history is an append-only event table from day one; play counts,
recently played, and most played are views over it. Ratings and
external scrobbling wait for 2.1. Rejected: full Navidrome parity at
launch (scrobbling brings credentials and rate limits into the launch),
playlists only (starred and play counts are cheap and missed on day
one).

### Relation to the current ladder

The ladder finishes to launch with Navidrome as the reference driver.
Rung 12 is trimmed to the minimum that makes Navidrome usable behind
dubplate's login (SSO and roster sync); its theming polish is dropped.
The driver seam stays, so 2.0's streamer arrives as one more driver and
the cutover is a configuration flip per estate. Rejected: folding 2.0
into the ladder before launch (launch slides behind the largest
subsystem on a system that is not live), replacing rung 12 with the
built-in streamer (loses the escape hatch).

### Evidence: Navidrome as the oracle, plus a client roster

Navidrome and dubplate serve the same fixture library; a recorded
session issues every launch endpoint to both, and a differ compares
responses field by field with a ledger of adjudicated divergences. A
roster of the apps the family uses walks each slice with screenshots.
Rejected: the roster alone (a client hides most of the API), the spec
alone (silent where clients depend on Navidrome's de facto behavior).

### Build shape: evidence first, then vertical slices

The study rung and the harness rung open the initiative and touch no
production surface. The streamer then lands in slices that each reach a
real client end to end. Rejected: slices first with evidence
retrofitted (every early slice is re-verified once the harness exists,
the plan-staleness defect class this repo has already paid for), a port
of Navidrome's architecture (license risk and the scanner model).

### Local mirroring: one writer per collection

The server stands on its own. Its intake, catalog, and streamer are the
complete product, and the launch estate has no local binary anywhere.
When a person runs the future local dubplate binary (the organizer
spec's post-launch two-binary split), the local catalog is authoritative
for the collections it owns and mirrors them to the server: a push of
catalog rows plus the content-addressed files the server lacks, keyed on
hashes and idempotent. Collections the server owns are written only by
the server's intake. Every collection has exactly one writer, so no
merge rule exists. Rejected: bidirectional sync (a conflict model, the
class of problem that sinks sync products), server as writer with the
local binary as a cache (inverts the local workflow Geoff described).
Collection ownership is therefore a catalog attribute, forward hook 8.

### License

dubplate is MIT (Geoff, 2026-09-07). Navidrome and gonic are GPLv3 and
are studied for behavior only; no implementer reads their code. Their
permissive dependencies are ours to adopt. Apache-2.0 was considered
and declined: its patent grant and NOTICE clause buy nothing for this
project and add friction for a stranger forking the estate.

## Architecture

### Catalog projection

The catalog has one row per admitted track: content hash, path, release
and recording MBIDs, the tags as applied, and the collection. A
projector derives the OpenSubsonic entities (artist, album, song,
genre, music folder) into tables of their own inside the same SQLite
file. The projector is a pure function of the catalog: a full rebuild
produces byte-identical tables and gets a determinism test, and every
catalog event (admit, retag, move, remove) applies its delta in the
same transaction that commits the event.

Identity comes from the pipeline. An album is a release MBID within a
collection; two pressings are two albums; a compilation is whatever
MusicBrainz says, with the album artist from the release credit;
multi-disc releases are one album with disc numbers. Entity ids derive
from MBIDs where they exist and from the content hash otherwise, so ids
survive a rebuild and a starred list never dangles. Collections map to
Subsonic music folders, which yields multi-library as a consequence of
the mapping. Clients that browse by directory get a folder view derived
from the paths the pipeline chose.

### Art

Art has one origin, the pipeline. The admitted cover image is stored as
a content-addressed blob with its hash on the catalog row; embedded
pictures read through taglib are the fallback. The cover-art endpoint
resizes on demand with `golang.org/x/image`, encodes WebP when the
client accepts it and JPEG otherwise, and writes each variant into the
same content-keyed cache the transcoder uses under (art hash, size,
format). Variants are immutable, served with a one-year cache lifetime
and an ETag. A missing cover returns a generated placeholder, never a
404. Artist images are on the 2.1 list.

### Search

The search endpoint answers from an FTS5 index the projector maintains
beside the entity tables, using the trigram tokenizer over
accent-folded text. Folding reuses the matcher's unidecode table
through the folded sort key stored on the catalog row. Ranking is
prefix match, then FTS5 rank, then the matched artist's own albums.
Results carry the same ids as browsing. An edit-distance rerank is held
for 2.1 if trigram recall proves thin.

### Streaming and transcoding

The stream endpoint has two paths. The direct path serves the library
file through the standard library's range-aware file serving, so
seeking and resume are the client's ordinary byte-range requests and
the server holds no per-stream state. The transcode path runs when the
request names a format or bitrate cap or the device's profile sets one.
Profiles resolve in order from the request, the device, and the estate
default. Each profile is a named ffmpeg argument template held in
configuration, with a small shipped set and no user-supplied argument
strings.

On a cache miss, ffmpeg starts under the request's context with a
per-track timeout; its output streams into the cache entry through
`djherbis/fscache`, and the client reads the entry as it fills, so a
second listener shares the encode. On a hit the entry serves like a
static file, ranges included. A disconnect kills ffmpeg only when no
other reader is attached. The cache enforces its size cap, drops every
variant of a track the catalog removes or retags, and reports size and
hit rate as doctor rows. One knob bounds concurrent ffmpeg processes
(default two); requests past it queue in arrival order. Play history
records an event when a stream passes a threshold of the track's length
or the client reports a scrobble, whichever comes first.

### Identity and app passwords

App passwords are long random strings the server generates. The stored
secret is encrypted at rest with a server key from the estate's secrets
manifest, because the legacy salted-token scheme needs the plaintext to
verify. The deprecated plaintext and hex password parameters are
refused. Failed authentications are rate-limited per account and per
address; every mint, first use from a new address, and revoke is an
event. The Subsonic username is a roster handle: short, ASCII, unique,
distinct from the login email. Roles map from the roster (admin or
member); the API's finer flags derive from the role. Share links are
scoped tokens for one album or playlist with an expiry, served on a
public path with no account.

### Per-user state

Four tables keyed by roster user and entity id: starred (one row per
user, entity, timestamp), playlists (a header plus ordered entries,
mutated by whole-list replace in one transaction), the play queue (one
row per user with ordered ids, position, and client), and play history
(append-only events: user, song, client, timestamp, how reported). No
row references a path; a catalog removal cascades to the rows that
named it.

### API surface

One handler set on the standard library mux under the Subsonic path,
answering XML and JSON from the same typed response structs, with
implemented extensions declared through the extension-discovery
endpoint. Launch implements the endpoints the client roster exercises:
system, browsing (ID3 and folder views), album lists, search, media
(stream, download, cover art; lyrics stubbed), annotation (star,
unstar, scrobble), playlists, play queue, and read-only user. Every
other endpoint returns the spec's not-implemented error. The harness
records which endpoints each client calls, so the launch list is
measured. Every endpoint is one function with one test table.

### Cutover and operations

The streamer is one more driver behind rung 12's seam, selected by one
configuration value. Its first weeks run as a shadow: Navidrome stays
the family's server, 2.0 serves the same library on a second hostname,
the harness runs on a timer against the live box, and the roster walks
it in place. The flip is a configuration change and a deploy in one
owner sitting. Navidrome's container stays defined but stopped for a
named period and is removed at a dated close-out. Starred lists and
playlists built in Navidrome are imported once through an export the
harness reads; differences are reported, never silently merged.

After the flip the box runs the dubplate binary, ffmpeg, cloudflared,
and the timers. Provisioning adds ffmpeg and the cache directory. The
streamer's state lives under the existing state directory, so backup
covers it. Doctor gains rows for ffmpeg's version, cache size against
cap, index rebuild parity, and the last drift count. Configuration adds
an integration class in rung 4's shape: enabled, cache cap, concurrent
encodes, the profile set, and the app-password key, each required when
enabled.

## The ladder

1. **Study.** A written record of Navidrome's and gonic's behavior where
   the spec is silent, cited to a version and to a harness test, no
   code quoted; and the prior-art table scored on the four adoption
   criteria.
2. **Harness.** The fixture library, the Navidrome container, the
   recorded session, the differ, the divergence ledger, and the client
   roster's scripted walk.
3. **Catalog projection and browsing.**
4. **Art and search.**
5. **Streaming and transcoding.**
6. **Identity and app passwords.**
7. **Per-user state.**
8. **Shadow period, flip, and close-out.** Retires Navidrome and records
   the post-mortem.

Each slice is a pass with its own plan, ceiling, oracle diff, and
roster walk, under the conducting loop in CLAUDE.md. Rungs 1 and 2
touch no production surface and may run as a document-and-corpus lane
beside the current ladder's later rungs.

## Forward hooks the current build owes

Each hook names the rung whose plan weighs it. A rung plan states per
task whether it lands a hook or is neutral (CLAUDE.md, "Build for
2.0").

1. **Content hash on the catalog row** (rung 6). The cache key, the
   fallback entity id, and the drift check all read it.
2. **Folded sort key on the catalog row** (rung 6, computed at rung 7's
   admit). The projector and the search index agree on folding without
   re-deriving it.
3. **Art blob recorded at admit** (rung 6 schema, rung 7 intake). The
   admitted cover stored content-addressed with its hash on the row.
4. **Play history as timestamped events** (rung 6 schema shape, if any
   listening record lands before 2.0). Never a bare counter.
5. **A roster handle** (rung 5). Short, ASCII, unique, distinct from
   the login email; the Subsonic username.
6. **The driver seam stays universal** (rung 12). The built-in streamer
   must fit as a driver, and rung 12 builds no more of the Navidrome
   driver than SSO and roster sync.
7. **Schema migrations** (rung 6). 2.0 guarantees the schema evolves
   across versions; the store's migration mechanism is chosen with that
   in view.
8. **Collection ownership** (rung 6). Each collection records its
   writer (this server, or a named local binary), so mirroring has one
   writer per collection from the first schema.

## Prior art

The study rung scores each entry on the four adoption criteria in
CLAUDE.md (license, code quality against the house bar, maintenance,
adoption) with evidence per cell. The columns are fixed here; the cells
are the study rung's deliverable.

| Entry | Role | Consulted by |
|---|---|---|
| OpenSubsonic specification and extension list | The contract | every 2.0 rung |
| Navidrome (GPLv3, behavior only) | The oracle and the reference | study, harness |
| gonic (GPLv3, behavior only) | A smaller Subsonic server; its author wrote taglib's Go binding | study |
| `go.senan.xyz/taglib` (MIT wrapper; TagLib under LGPL-2.1 or MPL-1.1, as WebAssembly under wazero) | Tag read and write for every container | rung 8 now; 2.0 art fallback |
| `tetratelabs/wazero` (Apache-2.0) | Runs the WebAssembly | with taglib |
| `modernc.org/sqlite` (BSD-3) | cgo-free SQLite | rung 6 now |
| `pressly/goose` (MIT) | Schema migrations | rung 6 now |
| `mewkiz/flac` (MIT) | Pure-Go FLAC decoder; the intake integrity check | rung 7 now |
| Picard's tag mapping tables (documentation) | Which tag names each container uses | rung 8 now |
| `golang.org/x/image` (BSD-3), `gen2brain/webp` (MIT) | Art resize and encode | 2.0 art |
| `djherbis/fscache` (MIT) | Stream-through file cache | 2.0 transcoding |
| `xrash/smetrics` (MIT) | Edit-distance rerank | 2.1 search |
| Chromaprint via `fpcalc` | Acoustic fingerprinting for untagged uploads | 2.1 |
| Symfonium's server-compatibility notes | Where servers diverge in practice | study |

Navidrome's remaining dependencies (chi, jwtauth, viper, notify, cron,
extism, mpvipc, squirrel, prometheus) are architecture 2.0 chose
against or already covers, and are not adopted.

## The 2.1 list

Scrobbling (ListenBrainz, Last.fm); ratings; artist images; transcode
pre-warming per profile; the first-party web player; metrics; an
edit-distance search rerank; acoustic fingerprinting for untagged
uploads; jukebox mode, if ever.

## Open items

None blocks the study rung.

- The client roster: Which apps the family installs decides the
  roster; answered at the harness rung.
- taglib's license variant: MPL-1.1 is the safe choice for an MIT
  project; the rung 8 plan confirms which variant the Go binding ships
  and records it.
- The app-password key's origin: The estate's secrets manifest and
  the age store, per CLAUDE.md's secrets rules; the exact name is the
  identity slice's.
- **The shadow period's length and the Navidrome retention window.**
  Owner rulings at the cutover slice.
- **Whether the study rung runs beside the current ladder or after
  launch.** A document lane is permitted by the standing grant; the
  conductor decides on the lane registry's state at the time.
