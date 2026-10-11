# ROADMAP

> Strategic initiatives. Managed by `/log-project`. Issues tracked in `BACKLOG.md`;
> uncommitted features and member requests in `IDEAS.md`.
> Product essence: a group of family or friends building a well-organized
> library for a shared music server. Sequencing ruling (Geoff, 2026-09-02,
> supersedes 2026-09-01): the greenfield dubplate-server build is the one
> active initiative; it launches complete (Go matcher included), TUI and
> local binary follow, `packaging` after that. The group's system is not
> live until launch.

## Path to 0.8

The order of the remaining work on the greenfield ladder, derived from the dependency analysis
(`docs/superpowers/research/2026-09-03-rung-dependency-analysis.md`), the reverse-edge note
(`docs/superpowers/research/2026-09-04-reverse-edge-completion.md`), the re-plan
(`docs/superpowers/specs/2026-10-05-rungs-12-16-replan-design.md`), the intake and `admin-api-1` specs, and the
lean pass process ruling that puts `intake` after `admin-api-1`. A rung is named, and a pass is `<name>-<n>`.

Landed: `foundation`, `corpus`, `matcher`, `shadow` (its program was deleted on 2026-10-05), `auth`, `store`,
`tag-io`, and `streamer-1` to `streamer-3`.

1. **`admin-api-1`** (14a). The first slice of the versioned API and the admin surface. Inbound: `auth` and `store`,
   both landed; the edge from `pipeline` was discharged by Geoff's per-album lease ruling in its spec. Carried work:
   its planning prompt. STATUS names it as the next pass.
2. **`intake`** (rung 7). The uploader and the intake path, as `intake-1` to `intake-3`, ending in Geoff's uploader
   sitting. Inbound: `store`, landed; it follows `admin-api-1` by ruling, so no `intake` segment runs beside it.
   Carried work: the intake spec's "Owed records", plus its BACKLOG items.
3. **`admin-api-2`** (14b). The second slice of the API, the part that writes a collection row. Inbound: `intake`,
   which creates the default collection. Carried work: BACKLOG items tagged `#rung:admin-api`.
4. **`pipeline`** (rung 9). The admission pipeline and the catalog, large enough to split into passes at its own
   planning. Inbound: `store` and `tag-io`, landed; it follows `intake` for the handoff record and the `accepted` state,
   which would otherwise cost a stub producer. It shares no files with `admin-api-2`, so the two may run together.
5. **`streamer`** (rung 12). Slices 3 to 8 of the built-in OpenSubsonic streamer, as `streamer-4` onward (a slice may
   take more than one pass). Inbound: `pipeline`'s catalog. It goes first of the rungs that open after `pipeline`
   because it is the longest and 0.8 ships on it. Slice 7 waits on `queues`. A soft edge ties it to `builtins`, which
   names a post-scan hook event the streamer fires; items 5 to 9 open together, so the order stands.
6. **`queues`** (rung 13). The decision queue. Inbound: `matcher`, `store`, and `pipeline`. It goes ahead of
   `migration` and `builtins` because streamer slice 7 consumes it.
7. **`migration`** (rung 10). Migration and layout parity. Inbound: `corpus`, `store`, and `pipeline`.
8. **`builtins`** (rung 11). The built-in plugin roster. Inbound: `store`, `tag-io`, and `pipeline`; it is
   independent of `migration` and `queues`.
9. **`library-web-1`**, a web pass outside the ladder. The album page, the library grid and list, filtering, and buy
   links, which the intake spec moved to a web pass after `pipeline`. Inbound: `pipeline`'s catalog.
10. **`integration`** (rung 15). The whole-system pass against the launch bar. Inbound: every component rung.
11. **`cutover`** (rung 16). The first launch, 0.8.0, ending in Geoff's first-launch sitting. Inbound: `integration`.
12. **`closeout`** (rung 17). Clean-up after launch. Inbound: `cutover`, and `migration` latest, because the legacy
    ledger fixture is captured before `_salvage/` is deleted.

Items whose edges are met may run at once, up to three passes (ruling 1 of the lean pass process); the numbers set the
start order and are rewritten when the order changes. Items 5 to 9 open together once `pipeline` lands, and items 10 to
12 are strictly serial. Carried work for each item lives in BACKLOG, under items tagged `#rung:<name>`, except where
an item names another home.

Frozen documents keep the old rung numbers, and this table bridges them to the names.

| Old rung | Name | Landed passes (old ids) | New pass ids |
| --- | --- | --- | --- |
| 1 | `foundation` | 1, 1b | none owed |
| 2 | `corpus` | 2a to 2d | none owed |
| 3 | `matcher` | 3a to 3d | none owed |
| 4 | `shadow` | 4a, 4b | none owed |
| 5 | `auth` | 5a to 5d | none owed |
| 6 | `store` | 6a to 6d | none owed |
| 7 | `intake` | none | `intake-1` to `intake-3` (7a to 7c) |
| 8 | `tag-io` | 8a to 8c | none owed |
| 9 | `pipeline` | none | `pipeline-1` onward, split at planning |
| 10 | `migration` | none | `migration-1` onward |
| 11 | `builtins` | none | `builtins-1` onward |
| 12 | `streamer` | slices 1, 2a, 2b | `streamer-1` to `streamer-3` landed; `streamer-4` onward (slices 3 to 8) |
| 13 | `queues` | none | `queues-1` onward |
| 14 | `admin-api` | none | `admin-api-1` (14a), `admin-api-2` (14b) |
| 15 | `integration` | none | `integration-1` onward |
| 16 | `cutover` | none | `cutover-1` onward |
| 17 | `closeout` | none | `closeout-1` onward |
| none | `library-web-1` | none | `library-web-1`, outside the ladder |

## Active

### Greenfield dubplate-server `dubplate-organizer`
The universal shared-library server, built fresh and foundation-first on the 17-rung ladder of the ratified build
spec (`docs/superpowers/specs/2026-09-02-greenfield-server-build.md`, revision 2.1, with the 2026-10-05 re-plan
errata). The organizer spec (`docs/superpowers/specs/2026-09-01-dubplate-organizer-design.md`) continues to govern
purview, API shape, and the ops seam. The first launch is 0.8.0 on the built-in streamer (Geoff, 2026-10-05): the
group listens through dubplate's own OpenSubsonic streamer from the first day. Launch bar: the complete server,
streamer included, with the Go matcher past the two-tier conformance bar and layout parity. Old code is salvage tier,
frozen at the `pre-greenfield` tag under `_salvage/`. Server before TUI is a hard boundary. Next: `admin-api-1`. The
order of the rest is "Path to 0.8".

### The built-in streamer `streamer`
Subsume Navidrome so the estate is one clean integrated system (Geoff, 2026-09-07): one binary a stranger can set up.
The spec is `docs/superpowers/specs/2026-09-07-streamer-2-0-design.md` (revision 3.1, 2026-10-07; the file still says
2.0), and it carries the forward hooks and the prior-art list that every rung plan weighs its surface against.
Navidrome is GPLv3: study it for behavior, never copy. Slices 1 to 2b have landed; next: slice 3, which waits on
`pipeline`'s catalog. The web player needs a strong play-queue interface (Geoff, 2026-09-24): now playing, up next,
reorder, and add from album, track, or filter results, designed from a prior-art study in the visual identity spec's
language, and building on 0.8's saved per-member play queue. The order of the work is "Path to 0.8".

## Planned

### dubplate-server 0.9: the paired local client `local-client-0-9`
The 0.9 feature (Geoff, 2026-09-25 ruling on version numbering): a local client
paired with dubplate-server. It ships the home install, in two modes (Geoff,
2026-10-10). Standalone: dubplate runs at home on the LAN with no VPS, which Geoff
expects many home audiophiles to need alone. Home copy: described below. It runs on
any likely home audiophile setup: native Linux on arm64 and amd64, a multi-arch
container image (Synology, Unraid, TrueNAS, QNAP), macOS, and Windows. macOS and
Windows verification is reasoned and disclosed, since the workstation cannot run them.
0.8 is built ready for it through the streamer spec's forward hooks 23 (no platform
assumption in the core) and 24 (an install that needs no public host); 0.8 builds
neither mode. The home copy server (Geoff, 2026-10-10):
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
catalogs. The `pipeline` plan weighs whether a sync source (a change feed over the catalog,
keyed on hook 1's audio keys) is a 0.8 forward hook, and routes it to the streamer
spec's list if so. Everything else, including the pairing credential, the sync
transport, and how an app picks the home copy or the server, is settled when 0.9 is
planned.
The Pi 5 stereo client research (2026-10-10) is the live input for that plan; its current decision is
`docs/superpowers/research/2026-10-10-pi5-client/proven-build.md`.

### dubplate-server 1.0: the format ladder `formats-1-0`
(Re-keyed from "1.1"/`formats-1-1`, Geoff, 2026-09-25: the format ladder is 1.0's
feature; 0.8 and 0.9 stay FLAC-only.)
Ruled by Geoff, 2026-09-08 (the `tag-io` spec, owner ruling 5, closing fork
B): 0.8 and 0.9 ship FLAC-only, and the other formats target the 1.0
release, sequenced after the 0.8.0 launch, which the `streamer` rung carries. The ladder, each
step its own pass in the shape of `tag-io` (a per-container writer behind
the seam, a golden container, a corpus stratum derived through beets,
and a conformance run against the ledger): MP3 (ID3v2.4, the recorded
version) with Ogg Vorbis and Opus riding the same pass; then M4A (AAC
and ALAC, one container and one iTunes-atom tag model); then WAV and
AIFF by transcode to FLAC at intake, which waits on the ffmpeg seam
and may therefore reorder against 0.8. Everything else stays a reject
whose message names the supported set and suggests a lossless
conversion to FLAC where one exists. The `intake` reject copy is written to
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
  Apache-2.0 OpenAPI file (reading, 2026-09-25, reversible: the prior-art survey found
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

### PR test CI `pr-test-ci`
Add a workflow that runs `bash scripts/check.sh` on `pull_request`. Until it lands, the full-suite home is the
local close gate (ruling 11, lean pass process spec). It lands after `admin-api-1` (ruling 3, alignment design).

### Production cleanup and polish `prod-polish`
Largely superseded by the greenfield ruling (2026-09-02): group accounts
and live-use intake move to the new server's launch; what survives is the
debris cleanup (stale failed units, doc drift), foldable into the greenfield build's early passes. Intake for
live findings: `/log-issue` → `BACKLOG.md` with `#project:prod-polish`.
Plan (committed 2026-09-01, ahead of need):
`docs/superpowers/plans/2026-09-01-prod-polish.md`.
Related: #2, #3, #4, #5, #6, #7, #8

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
and sending the outreach letters. Separate from the ladder; `intake` consumes it.
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
lands). Two generators are owed to dubplate: the reject-page generator when `intake` merges, and the
configuration-reference generator when `admin-api` merges. Phase B publishes with 0.8, beside public
signed binaries and a fresh public source repository from a scrubbed 0.8 snapshot. Open: the
domain, `dubplate.dev` or `dubplate.studio`, due before Phase B.

## Someday

### Open-source the pair `oss-pair`
Publish the uploader and the pings Worker as documentation-linked siblings:
a validated-intake drop target and a dead-man's-switch monitor, proven here
first. Preconditions: months of group use, the uploader's auth/validator
seams, and a clear-eyed maintainership decision. First deliverable,
independent of any code release: the validated-intake pattern write-up (what
each layer catches that the others cannot; 907.life is a natural home),
which costs no maintainership and doubles as the packaging demand test.
