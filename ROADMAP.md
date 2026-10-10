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
2.1, with the 2026-10-05 re-plan errata), which supersedes the organizer
spec's own build order while the organizer spec continues to govern
purview, API shape, and the ops seam
(docs/superpowers/specs/2026-09-01-dubplate-organizer-design.md, amended
through 2026-09-02, folded per "Roadmap revisions this spec carries"
below). Two tracks dependency-merge at rung 15: the architecture spine
(foundation, auth, ops seam + store, intake, pipeline, the built-in
streamer, decision queue, versioned API) and an independent evidence
track (corpus/harness, Go matcher), which ends at the corpus and the
conformance bar. The ladder closes with the first-launch sitting (rung
16: the beets DB migration and the Navidrome import from the retirement
archive) and close-out (rung 17). The first launch is 0.8.0 on the
built-in streamer (Geoff, 2026-10-05, superseding the 2026-09-25 plan for
a 0.x pre-release): the group listens through dubplate's own OpenSubsonic
streamer from the first day, with no shadow record in the launch bar and
no box swap, because the incumbent stack was retired on 2026-10-05. Launch
bar: the complete server, streamer included, with the Go matcher past the
two-tier conformance bar and layout parity. Old code is salvage tier,
frozen at the `pre-greenfield` tag under `_salvage/`. Server before TUI
is a hard boundary: the organizer spec's rung 10 (two-binary split, local
TUI) and the remote-TUI half of its rung 11 move past launch,
unscheduled on the new ladder. Rung 1 (foundation) through rung 8 (tag
I/O: the writer core and reader, PICTURE and the conformance instrument)
are landed, alongside rung 4 (whose shadow program the removal pass
deleted on 2026-10-05; its history is the `pre-greenfield` tag and git),
rung 1b, rung 3, rung 5 (closed 2026-10-05), and rung 6 (6d closed
2026-10-07). Rung 12's slice 1 (the Navidrome study) closed 2026-10-07, and slice 2
(the oracle harness) closed 2026-10-10 with pass 2b, beside the ladder;
slice 3, the first with a dubplate side, is ready to plan. The spine
continues with rung 14.

Folded from the organizer spec's "Roadmap revisions this spec carries"
(2026-09-01-dubplate-organizer-design.md):
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
  the reference implementation. (Cloudflare Access as an optional
  forward-auth front moved to `IDEAS.md`, 2026-10-05.)

### Rung 12: the built-in streamer
(Keyed `streamer-2-0`, re-keyed `streamer-0-8` by Geoff, 2026-09-25: 0.8 is the first production release.
Folded into the greenfield initiative as rung 12, Geoff, 2026-10-05: the first launch is 0.8.0
on this streamer, so it is a rung of the ladder, not a release after it.)
Subsume Navidrome so the estate is one clean integrated system (Geoff,
2026-09-07; primary driver: one binary a stranger can set up). Brainstormed
2026-09-07; spec at `docs/superpowers/specs/2026-09-07-streamer-2-0-design.md`
(revision 3, 2026-10-05; ratified 2026-09-07 after four adversarial reviews under
`docs/superpowers/research/2026-09-07-streamer-2-0-review-*.md`). Rung 12 is the spec's
eight slices; the shadow-and-flip slice is deleted. Slices 1 (the study) and 2
(the harness) start now, ahead of rung 12's turn on the ladder (re-plan
design, `docs/superpowers/specs/2026-10-05-rungs-12-16-replan-design.md`). Settled so
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
roster. dubplate integrates with no external streamer: the driver seam, hook 6,
and the Navidrome and Jellyfin targets are retired (Geoff, 2026-10-05; the entry
moved to `IDEAS.md`). Navidrome is GPLv3: study it for behavior, never copy; its permissive dependencies
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

**Before the 0.8 release (Geoff, 2026-09-26):** recheck TypeScript 7 (7.0.2 at the time).
The bump waits on a svelte-check release that supports it; 4.7.6 does not, and `--tsgo`
hits rootDir errors (TypeScript #30693, sveltejs/language-tools #2891). No hurry before then.

## Planned

### dubplate-server 0.9: the paired local client `local-client-0-9`
The 0.9 feature (Geoff, 2026-09-25 ruling on version numbering): a local client
paired with dubplate-server. Its core is the home copy server (Geoff, 2026-10-10).
A dubplate install on a machine on a member's home network keeps a one-way synced
copy of the whole library and serves it to that member's music apps over the same
OpenSubsonic API. At home the apps stream lossless from the house; away from it they
use the server. It must run on a current Raspberry Pi (Geoff, 2026-10-10): arm64,
the library on an attached USB or NVMe drive, and a memory and CPU budget that 0.9's
plan measures on the then-current model. It builds on 0.8's arm64 builds (Promoted in
`IDEAS.md`), and it never transcodes, since it serves the stored FLAC untouched. This supersedes the organizer spec's out-of-scope line for "the
server-to-local mirror" (`2026-09-01-dubplate-organizer-design.md:704`), which stays as
written. It sits inside the group-scale principles: the copy serves the member's own
devices, which those principles name as intended use, and it is not a link between two
catalogs. Rung 9's plan weighs whether a sync source (a change feed over the catalog,
keyed on hook 1's audio keys) is a 0.8 forward hook, and routes it to the streamer
spec's list if so. Everything else, including the pairing credential, the sync
transport, and how an app picks the home copy or the server, is settled when 0.9 is
planned.

### dubplate-server 1.0: the format ladder `formats-1-0`
(Re-keyed from "1.1"/`formats-1-1`, Geoff, 2026-09-25: the format ladder is 1.0's
feature; 0.8 and 0.9 stay FLAC-only.)
Ruled by Geoff, 2026-09-08 (rung 8 spec, owner ruling 5, closing fork
B): 0.8 and 0.9 ship FLAC-only, and the other formats target the 1.0
release, sequenced after the 0.8.0 launch, which the rung 12 streamer carries. The ladder, each
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
- **OpenSubsonic protocol layer** (decided): hand-written wire types, validated against the
  Apache-2.0 OpenAPI file (conductor reading, 2026-09-25, reversible: the prior-art survey found
  `oapi-codegen` output unfit for production types and `ogen` failing on the full spec, and the
  streamer spec already hand-rolls the dual encoding), a request-parsing handler over a backend interface, a client, and a conformance checker.
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
  confirmation of the distribution reading before that release. **Landed 2026-10-03**
  (HISTORY, the transliteration entry). Routed here for the module's publication:
  - The lawyer reads the four Unidecode-derived items the repo still carries, not an absolute
    claim: `internal/engine/matcher/testdata/case-mapping-sweep.tsv` (56 rows computed through
    Unidecode 1.4.0), the pinned lookup rows in `internal/engine/matcher/unidecode`'s tests,
    `tools/unidecode-table/generate.py` (imports the GPL library), and `Lookup`'s port of
    `_get_repl_str`'s rules.
  - Git history: the `.tbl` stays in every commit before the pass's Task 2. Whether history is
    rewritten before the repo goes public is the same lawyer's question (the fresh scrubbed
    repo planned at 0.8 may answer it).
  - The transliterator holder is process-wide, installed once by the main package. A published
    matcher library should weigh passing the seam explicitly instead.

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
Make the estate easy for a stranger to set up from its ingredients (beets
or its harvested successor, the binaries, their Cloudflare account, any
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
Round one closed on 2026-09-26 with 116 photos committed to `photos/`, and Geoff ruled the pool
enough for 0.8.0, so later rounds are continuing work that never blocks a release. Each later
round starts from the targets in `docs/superpowers/research/photo-curation-rounds/round-1.md`
(Round close) and the artist list beside it, ranks period photos first, and budgets the sorting
page as its own task. The outreach letters (drafted under the identity-photos research) wait
until dubplate has a public product to point to (Geoff, 2026-09-26); most of the outreach list's
gaps, such as early Devo, go-go, and the 1980s dancehall archives, are reachable only that way.

### The project site `project-site`
dubplate's public project home for strangers: a standard cairn site in its own repo
(`~/Projects/dubplate-site`) that vendors the theme tokens, the guides, and the generated
references from a pinned dubplate release, checked against a lock file. Spec:
`docs/superpowers/specs/2026-09-25-project-site-design.md` (revision 2.5, approved 2026-09-25).
**Trigger:** cairn's docs reset has closed and `create-cairn-site` is published on npm; the
build doubles as a stranger's trial of cairn's docs. Phase A builds to a `noindex` preview and
moves the principles spec's is/isn't lines and legal note into `docs/guide/` (its erratum then
lands). Two owed dubplate lanes: the reject-page generator when rung 7 merges, and the
configuration-reference generator when rung 14 merges. Phase B publishes with 0.8, beside public
signed binaries and a fresh public source repository from a scrubbed 0.8 snapshot. Open: the
domain, `dubplate.dev` or `dubplate.studio`, due before Phase B.

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
