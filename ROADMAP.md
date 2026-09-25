# ROADMAP

> Strategic initiatives. Managed by `/log-project`. Issues tracked in `BACKLOG.md`;
> uncommitted features and member requests in `IDEAS.md`.
> Product essence: a group of family or friends building a well-organized
> library for a shared music server. Sequencing ruling (Geoff, 2026-09-02,
> supersedes 2026-09-01): the greenfield dubplate-server build is the one
> active initiative; it launches complete (Go matcher included), TUI and
> local binary follow, `packaging` after that. The group's system is not
> live until launch.

## Active

### Greenfield dubplate-server `dubplate-organizer`
The universal shared-library server, built fresh and foundation-first on
a 17-rung ladder per the ratified build spec
(docs/superpowers/specs/2026-09-02-greenfield-server-build.md, revision
2.1), which supersedes the organizer spec's own build order while the
organizer spec continues to govern purview, API shape, and the ops seam
(docs/superpowers/specs/2026-09-01-dubplate-organizer-design.md, amended
through 2026-09-02, folded per "Roadmap revisions this spec carries"
below). Two tracks dependency-merge at rung 15: the architecture spine
(foundation, auth, ops seam + store, intake, pipeline, streamer
drivers/SSO, decision queue, versioned API) and an independent evidence
track (corpus/harness, Go matcher, shadow against the live estate as
oracle), closing with the launch cutover (rung 16: beets DB migration,
box swap, interim retirement) and close-out (rung 17). This launch is a
0.x pre-release the group uses; 0.8 is the first production release
(Geoff, 2026-09-25). Launch bar: the
complete server, Go matcher past the two-tier conformance bar, a clean
shadow record, layout parity. Old code is salvage tier, frozen at the
`pre-greenfield` tag under `_salvage/`. Server before TUI is a hard
boundary: the organizer spec's rung 10 (two-binary split, local TUI) and
the remote-TUI half of its rung 11 move past launch, unscheduled on the
new ladder. Rung 1 (foundation) through rung 8 (tag I/O: the writer
core and reader, PICTURE and the conformance instrument, and the
shadow's absorption and repoint) are landed, alongside rung 4 (the
shadow), rung 1b, rung 3, and rung 5's passes 5a and 5b-i (identity,
auth, config contract, session, and the two state records). Next: rung
5's pass 5b-ii (roster, limiters, notify, origin check, verify seam,
handlers).

Folded from the organizer spec's "Roadmap revisions this spec carries"
(2026-09-01-dubplate-organizer-design.md):
- Streamer integration ships with the Navidrome reference driver only
  (rung 12). Target systems are Navidrome and Jellyfin (trimmed from
  "Jellyfin and Subsonic" by Geoff, 2026-09-03: living Subsonic use
  is mostly Navidrome itself, Airsonic-Advanced is stale with no
  theming surface, and headless servers' users live in clients);
  a Jellyfin driver builds on demand, and the documented driver seam
  stays universal, so a Subsonic-family driver remains possible if
  real demand appears. Navidrome is the reference driver, not "the
  bonus only dubplate can offer": this page's product-essence line
  already reads generically ("a shared music server") and needs no
  further edit.
- The built-ins roster is set by policy and probe-backed demotions
  (rung 11), superseding any fixed plugin roster or adoption-band
  ship criterion; the survey backing it is
  `docs/superpowers/research/2026-08-31-beets-plugin-usage.md`.
- The API carries no client assumptions by requirement.
- Sync between rungs is publish-only; two-way sync is out of scope.
  (2026-09-12: the laptop-to-server mirror of a whole collection, one
  writer per collection, and the read-only checkout are in, scheduled
  with the TUI; configuration homes spec.)
- Publish (per-album, through intake) and backup/restore (bulk and
  admin) are separate paths; migration between rungs is not
  backup/restore-only.
- Auth generalizes to a verify-identity seam (rung 5); magic-link is
  the reference implementation, Cloudflare Access an optional
  forward-auth front (rung 12).

## Planned

### dubplate-server 0.8: the built-in streamer `streamer-0-8`
(Re-keyed from "2.0"/`streamer-2-0`, Geoff, 2026-09-25: 0.8 is the first production release.)
Subsume Navidrome so the estate is one clean integrated system (Geoff,
2026-09-07; primary driver: one binary a stranger can set up). Brainstormed
2026-09-07; spec at `docs/superpowers/specs/2026-09-07-streamer-2-0-design.md`
(revision 2.1, ratified 2026-09-07 after four adversarial reviews under
`docs/superpowers/research/2026-09-07-streamer-2-0-review-*.md`). Settled so
far: OpenSubsonic API first, first-party web player later; the pipeline's
catalog is the index (no scanner admits music; a reconcile walk only
reports drift); app passwords minted per device from the roster account;
ffmpeg transcoding under a content-keyed cache with a size cap; playlists,
starred, timestamped play history, and the saved queue at launch; art
resized on demand from a pipeline-admitted blob; FTS5 trigram search over
folded text. Group features join the target (Geoff, 2026-09-24; spec revision 2.3):
report a problem into the decision queue, `added_by` and album notes, collaborative
playlists, and smart playlists, with forward hooks 16 (rung 13) and 17 (rungs 7 and 9).
Hosting and setup join too (Geoff, 2026-09-24; revision 2.4): HTTPS without
Cloudflare, a general Navidrome import, arm64 builds, `dubplate init`, and
`dubplate upgrade`, with hooks 18 (rung 14) and 19 (rung 12).
Build shape A: an oracle harness (Navidrome and dubplate
against one library, responses diffed and adjudicated) and a written
Navidrome study first, then vertical slices each walked by a named client
roster. Sequenced after launch (rung 17); the driver seam stays and
Navidrome remains the escape hatch through 0.8's first months. Navidrome
is GPLv3: study it for behavior, never copy; its permissive dependencies
(taglib via wazero, x/image, fscache) are ours to use. **Standing rule
while the ladder builds:** each rung's plan weighs its surface against
the 0.8 spec's forward-hooks list (content hash and folded sort key on
the catalog row, the art blob recorded at admit, play history as events)
and says per task whether it lands a hook or is neutral, and names which prior-art entries it consulted (rung 6: modernc sqlite, goose; rung 7: mewkiz/flac for the intake integrity check; rung 8: taglib via wazero and Picard's tag tables). 2.1 list:
scrobbling (ListenBrainz, Last.fm), ratings, artist images, transcode
pre-warming per profile, the web player, metrics, an edit-distance
search rerank. The web player needs a strong play-queue interface (Geoff,
2026-09-24): now playing, up next, reorder, and add from album, track, or
filter results. It is designed from a prior-art study, as library filtering
was (`research/2026-09-24-library-filtering-prior-art.md`), in the visual
identity spec's language, and it builds on 0.8's saved per-member play queue
(`streamer-2-0-design.md:297-309`) so a queue carries between clients.

### dubplate-server 0.9: the paired local client `local-client-0-9`
The 0.9 feature (Geoff, 2026-09-25 ruling on version numbering): a local client
paired with dubplate-server. Scope is unspecified; nothing else about it is
settled yet.

### dubplate-server 1.0: the format ladder `formats-1-0`
(Re-keyed from "1.1"/`formats-1-1`, Geoff, 2026-09-25: the format ladder is 1.0's
feature; 0.8 and 0.9 stay FLAC-only.)
Ruled by Geoff, 2026-09-08 (rung 8 spec, owner ruling 5, closing fork
B): 0.8 and 0.9 ship FLAC-only, and the other formats target the 1.0
release, sequenced after launch and after `streamer-0-8`. The ladder, each
step its own pass in the shape of rung 8 (a per-container writer behind
the seam, a golden container, a corpus stratum derived through beets,
and a conformance run against the ledger): MP3 (ID3v2.4, the recorded
version) with Ogg Vorbis and Opus riding the same pass; then M4A (AAC
and ALAC, one container and one iTunes-atom tag model); then WAV and
AIFF by transcode to FLAC at intake, which waits on the ffmpeg seam
and may therefore reorder against 0.8. Everything else stays a reject
whose message names the supported set and suggests a lossless
conversion to FLAC where one exists. Rung 7's reject copy is written to
that message shape now, whichever step has landed. Open input: whether
anyone in the group has ever wanted to upload a non-FLAC file, which decides
whether the MP3 pass moves ahead of the cutover.

### Give back: permissive Go modules `give-back`
Geoff, 2026-09-25: where dubplate fills a gap in the Go ecosystem, its reusable layer is
authored so it can ship as its own permissively licensed module. Each candidate keeps a
gate-enforced import boundary now (`go list -deps` shows no `github.com/glw907/dubplate/` path)
and publishes only after a release has proven it. The evidence is in
`docs/superpowers/research/2026-09-25-give-back-survey.md` and
`2026-09-25-streamer-prior-art-survey.md`.
- **OpenSubsonic protocol layer** (decided): wire types generated from the Apache-2.0 OpenAPI
  file, a request-parsing handler over a backend interface, a client, and a conformance checker.
  Built with 0.8 and published after 0.8 ships. Clean-room record: behavior cites the OpenSubsonic
  spec or observation, never GPL source.
- **Matcher** (a beets autotag scoring port, MIT with beets' and `lap`'s notices), together with
  the beets translation half of `internal/engine/musicbrainz` and its conformance suite. Blocker:
  the embedded transliteration table derives from Python Unidecode (GPL-2.0-or-later), so the
  module ships a `func(rune) (string, bool)` seam and the generator, never the table.
- **FLAC tag I/O** (`tagio`, MIT): a crash-safe Vorbis comment and PICTURE writer on mediafile's
  field table, which no permissive Go library offers. Boundary work: cut `internal/atomicfile`,
  vendor its test fixtures, rewrite process citations.
- **Conformance corpus** (its own repo): strip MusicBrainz's `tags` and `genres` arrays
  (CC BY-NC-SA) first, and rewrite the README for strangers.
- Declined: a standalone MusicBrainz client (`musicbrainzws2` covers it), `atomicfile`, the
  store's migrations, and the replay stub.
- **dubplate stays MIT (Geoff, 2026-09-25; path 1):** the binary stops embedding the
  Unidecode-derived table. It carries a pinned checksum of the Unidecode 1.4.0 package instead;
  at setup it downloads that package, verifies it, and builds the table into the state directory
  with the generator's extraction logic ported to Go, with an offline path from a local file and
  a `doctor` row. The table stays byte-identical, so conformance is unchanged. The matcher reads
  it through the same `func(rune) (string, bool)` seam the published library needs. A small pass,
  due before the first release strangers can install (with `packaging`), plus a lawyer's
  confirmation of the distribution reading before that release.

### Production cleanup and polish `prod-polish`
Largely superseded by the greenfield ruling (2026-09-02): group accounts
and live-use intake move to the new server's launch; what survives is the
debris cleanup (stale failed units, doc drift) and the owed rewrite
post-mortem, foldable into the greenfield build's early passes. Intake for
live findings: `/log-issue` → `BACKLOG.md` with `#project:prod-polish`.
Plan (committed 2026-09-01, ahead of need):
`docs/superpowers/plans/2026-09-01-prod-polish.md`.
Related: #1, #2, #3, #4, #5, #6, #7, #8

### (folded into Active) dubplate as a music library organizer
The 2026-09-01 initiative entry is folded into the Active greenfield
entry above; its settled decisions live in the organizer spec. The
two-binary shape, remote-capable TUI, and plugin built-ins all stand,
sequenced behind the server per the 2026-09-02 rulings.

### Packaging `packaging`
Make the estate easy for a stranger to set up from its ingredients (Navidrome,
beets or its harvested successor, the binaries, their Cloudflare account, any
VPS). Opens with a brainstorm; context staged in agent memory (distribution
model: binary-as-installer vs clone floor; cloud-init generation; existing-box
bootstrap; pre-setup-VPS option map). **Runs after `dubplate-organizer`**
(resequenced 2026-09-01, was after `uploader`) so it packages the final
two-binary shape rather than an intermediate. Five setup pieces moved into 0.8
(2026-09-24; re-keyed from "2.0", Geoff, 2026-09-25): the TLS listener, the Navidrome import, arm64 builds, `dubplate
init`, and `dubplate upgrade`; packaging builds on them rather than owning them. The `oss-pair` write-up below
doubles as its demand test.
Carried option (Geoff, 2026-09-03): evaluate a Wails desktop app as a
third head on the shared engine: the existing Svelte UI in each OS's
own webview (native shell, single small binary per platform; macOS
notarized DMG, Windows signed installer, Linux Flatpak pinning
WebKitGTK), runnable embedded (engine in-process, no server) or as a
client of a hosted estate over the versioned API. Sequencing is firm:
the bubbletea TUI local organizer builds first after the server; the
Wails head waits beyond that.
Configuration brief for the TUI (connection profiles, the machine
settings and credential files, the attach overlap report and its four
verbs) and for every settings surface:
`docs/superpowers/specs/2026-09-11-configuration-homes-design.md`
(revision 3, 2026-09-12).

### Photo curation for the web surface `photo-curation`
Curate about 300 openly licensed photographs of music scenes, show flyers, and zine covers for
the login rotation, the credits gallery, the empty library, and the error pages. The design is
`docs/superpowers/specs/2026-09-25-photo-curation-design.md` (revision 2.2, approved 2026-09-25),
which supersedes the visual identity spec's curation rules where they differ; the active plan is
`docs/superpowers/plans/2026-09-25-photo-curation-round-one.md`. Inputs:
`docs/superpowers/research/2026-09-24-identity-photos/`. Owner touchpoints: the sorting sittings
and sending the outreach letters. Separate from the rung ladder; rung 7 consumes it.

## Someday

### Open-source the pair `oss-pair`
Publish the uploader and the pings Worker as documentation-linked siblings —
a validated-intake drop target and a dead-man's-switch monitor, proven here
first. Preconditions: months of group use, the uploader's auth/validator
seams, and a clear-eyed maintainership decision. First deliverable,
independent of any code release: the validated-intake pattern write-up (what
each layer catches that the others cannot; 907.life is a natural home) —
zero maintainership cost, and the packaging demand test.

## Done

### Uploader `uploader`
Pass A+B built and hardened the full uploader (46-finding fold, fidelity
chain passed, simplified, deployed); superseded pre-launch by the
greenfield ruling. The code is salvage tier for the new server (web
components, tus engine, checks, reject contract); the deployment is
retired. Spec/plan archived under docs/superpowers/. Closed: 2026-09-02.

### Go rewrite `go-rewrite`
The box-side shell layer became one `dubplate` binary (import, backup, beet,
ping-fail, checks, doctor, restore), carrying the musicbox → dubplate estate
rename. All four jobs green on supervised runs, overnight soak confirmed,
shell layer and msmtp deleted everywhere.
Completed: 2026-08-31 (post-mortem owed, tracked as #1)
