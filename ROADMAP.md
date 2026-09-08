# ROADMAP

> Strategic initiatives. Managed by `/log-project`. Issues tracked in `BACKLOG.md`.
> Product essence: a group of family or friends building a well-organized
> library for a shared music server. Sequencing ruling (Geoff, 2026-09-02,
> supersedes 2026-09-01): the greenfield dubplate-server build is the one
> active initiative; it launches complete (Go matcher included), TUI and
> local binary follow, `packaging` after that. The family system is not
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
box swap, interim retirement) and close-out (rung 17). Launch bar: the
complete server, Go matcher past the two-tier conformance bar, a clean
shadow record, layout parity. Old code is salvage tier, frozen at the
`pre-greenfield` tag under `_salvage/`. Server before TUI is a hard
boundary: the organizer spec's rung 10 (two-binary split, local TUI) and
the remote-TUI half of its rung 11 move past launch, unscheduled on the
new ladder. Rung 1 (foundation: the repo pivot, the config core,
`check`/`doctor` scaffolds) landed 2026-09-02. Next: rung 2's plan
(corpus and harness).

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
- Sync between rungs is publish-only; mirror and two-way sync are out
  of scope.
- Publish (per-album, through intake) and backup/restore (bulk and
  admin) are separate paths; migration between rungs is not
  backup/restore-only.
- Auth generalizes to a verify-identity seam (rung 5); magic-link is
  the reference implementation, Cloudflare Access an optional
  forward-auth front (rung 12).

## Planned

### dubplate-server 2.0: the built-in streamer `streamer-2-0`
Subsume Navidrome so the estate is one clean integrated system (Geoff,
2026-09-07; primary driver: one binary a stranger can set up). Brainstormed
2026-09-07; spec at `docs/superpowers/specs/2026-09-07-streamer-2-0-design.md`
(revision 1). Settled so
far: OpenSubsonic API first, first-party web player later; the pipeline's
catalog is the index (no scanner admits music; a reconcile walk only
reports drift); app passwords minted per device from the roster account;
ffmpeg transcoding under a content-keyed cache with a size cap; playlists,
starred, timestamped play history, and the saved queue at launch; art
resized on demand from a pipeline-admitted blob; FTS5 trigram search over
folded text. Build shape A: an oracle harness (Navidrome and dubplate
against one library, responses diffed and adjudicated) and a written
Navidrome study first, then vertical slices each walked by a named client
roster. Sequenced after launch (rung 17); the driver seam stays and
Navidrome remains the escape hatch through 2.0's first months. Navidrome
is GPLv3: study it for behavior, never copy; its permissive dependencies
(taglib via wazero, x/image, fscache) are ours to use. **Standing rule
while the ladder builds:** each rung's plan weighs its surface against
the 2.0 spec's forward-hooks list (content hash and folded sort key on
the catalog row, the art blob recorded at admit, play history as events)
and says per task whether it lands a hook or is neutral, and names which prior-art entries it consulted (rung 6: modernc sqlite, goose; rung 7: mewkiz/flac for the intake integrity check; rung 8: taglib via wazero and Picard's tag tables). 2.1 list:
scrobbling (ListenBrainz, Last.fm), ratings, artist images, transcode
pre-warming per profile, the web player, metrics, an edit-distance
search rerank.


### Production cleanup and polish `prod-polish`
Largely superseded by the greenfield ruling (2026-09-02): family accounts
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
two-binary shape rather than an intermediate. The `oss-pair` write-up below
doubles as its demand test.
Carried option (Geoff, 2026-09-03): evaluate a Wails desktop app as a
third head on the shared engine: the existing Svelte UI in each OS's
own webview (native shell, single small binary per platform; macOS
notarized DMG, Windows signed installer, Linux Flatpak pinning
WebKitGTK), runnable embedded (engine in-process, no server) or as a
client of a hosted estate over the versioned API. Sequencing is firm:
the bubbletea TUI local organizer builds first after the server; the
Wails head waits beyond that.

## Someday

### Open-source the pair `oss-pair`
Publish the uploader and the pings Worker as documentation-linked siblings —
a validated-intake drop target and a dead-man's-switch monitor, proven here
first. Preconditions: months of family use, the uploader's auth/validator
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
