# The dubplate project site

Design spec. Revision 2.4, 2026-09-25, drafted from a brainstorm with Geoff. Status: **awaiting
Geoff's read.** Three owner rulings stay open (see "Open rulings"). Revision 1 (`e91f528`) went
through five review lenses (contract, drift, consistency, mechanics, prior art) under
`docs/superpowers/research/2026-09-25-project-site-review-*.md`. The fold record is
`docs/superpowers/research/2026-09-25-project-site-fold.md`.

Revision 2 changelog: folds the five reviews, defines what the sync reads and what each gate
proves, and adds the Credits, Get help, and Bring your library pages (owner rulings 8 and 9).

Revision 2.1 changelog: folds the fold verification (V1 to V11): planned names in the guide leg,
owners for the late generators, open ruling 11 split three ways, and the exported photos.

Revision 2.2 changelog: folds the prose review, adds a generated third-party notices list that the
gate keeps current, and moves the is/isn't lines into Phase A now that the principles spec is
ratified.

Revision 2.3 changelog: records Geoff's ruling of option (b) on contract finding C2 and adds its
missing half, a site gate check that the lock's pin is the newest release.

Revision 2.4 changelog: records Geoff's ruling on ruling 11, public binaries and public source at
0.8, bounded by a fresh scrubbed repository and a stated contribution policy.

dubplate needs a public home where a stranger learns what it is, decides whether it fits their
group, and sets up a server. This spec fixes who the site serves, where it lives, which pages it
carries, who owns each page's content, and how the site stays in step with dubplate's code and
look. The site is a standard cairn site in its own repo. dubplate owns the theme tokens, the
guides, and the generated references, and the site carries a hash-locked copy of them taken from
a dubplate commit. Building the site doubles as a trial of cairn's documentation for a brand-new
site, so the build waits until cairn's docs reset is complete. The site goes public when dubplate
0.8 ships, and every page describes 0.8.

## Owner rulings

Geoff made these rulings on 2026-09-25, in order.

1. The site is dubplate's public project home, for strangers. It isn't a page for one group or
   a docs-only site.
2. It's a standard cairn site. The sync between it and dubplate is designed so the docs and the
   theme tokens can't drift.
3. dubplate owns the tokens and the guides. The site vendors a pinned snapshot from a dubplate
   release tag, checked against a lock file.
4. The build is a trial of cairn's docs and tools for a new site, and it waits until cairn's docs
   are complete.
5. The site is built early and published later. It publishes with 0.8 and describes 0.8.
6. The site map in "Pages" is accepted in outline, subject to this spec's adversarial review.
7. The site compares dubplate to a few other solutions, without disparaging anyone.
8. The site has a credits section, since beets was instrumental to the build and Navidrome was an
   inspiration.
9. The credits section also carries photo credits, for anyone who lets dubplate use licensed
   photos.

## Readers

The primary reader is the person who sets up and runs a server for their group. They're
comfortable with a terminal and a VPS, and they haven't read dubplate's code or specs. The
secondary readers are a member who has been invited to a group, and a person deciding whether
dubplate fits their group at all. Every page names which of these readers it serves.

## Where the site lives

The site is a standard cairn site in its own repo, `~/Projects/dubplate-site`. It's created with
`npx create-cairn-site` and deployed to Cloudflare the way cairn's docs describe. The site repo
starts private and opens when dubplate's public source repository does, at Phase B (ruling 11).
Keeping it out of the dubplate repo means cairn's publish flow, which commits edits to the site
repo's main branch, never touches dubplate's tree or its lanes.

## Pages

Each page has one owner. Site-owned pages are edited in cairn's editor. dubplate-owned pages are
written in dubplate's `docs/guide/` and vendored into the site. Generated pages are built from
dubplate's code, committed in dubplate, and vendored the same way. The site never edits a
vendored page (see "What each check proves").

| Page | Owner | Reader | Carries |
|---|---|---|---|
| Home | site | all | What dubplate is in one paragraph, three screenshots (the library, an album page, and the uploader turning an album away), and one path each for setting up a server and for an invited member |
| What dubplate is, and isn't | dubplate | deciding | The "dubplate is" and "dubplate isn't" lines and the legal note |
| How dubplate compares | site | deciding | Where dubplate fits among other self-hosted music servers, and when another one fits better |
| Set up a server | dubplate | operator | What you need, installing, `init`, the first admin, the first member, and the first backup |
| Bring your library | dubplate | operator | The Navidrome import, a folder of existing albums, and what gets turned away and why |
| Run your server | dubplate | operator | Members, collections, restoring a backup, `upgrade`, security and privacy, and leaving dubplate |
| For members | dubplate | member | Signing in, adding music, and listening with an OpenSubsonic app, including the tested app list |
| Why an album was turned away | generated | member | Every reject reason, in the wording the uploader and the emails use |
| Configuration reference | generated | operator | Every config key the registry declares |
| Command reference | generated | operator | Every command and flag of the server binary |
| Release notes | dubplate and site | operator | dubplate's changelog with upgrade notes, and one cairn post per release that links to it |
| Get help | site | all | Where to ask, what to include in a report, a short troubleshooting list, and the security contact |
| Credits | dubplate and generated | all | The software dubplate builds on or learned from, the third-party notices, and the photo credits |
| About | site | deciding | Who makes dubplate, the MIT license, and the stance on buying music |

The "is" and "isn't" lines and the legal note live today in the group-scale principles spec
(`2026-09-25-group-scale-principles-design.md`, "dubplate is, and dubplate isn't" and "A legal
note"). Its landing table routes both to the website. They change when a ruling changes, such as the
share link's week and month, so their source moves into `docs/guide/`, and the principles spec
points at it. The principles spec is ratified (revision 3, 2026-09-25), so the move lands in Phase
A's dubplate lane with the other guides. After the move, no second copy of the lines remains in the
principles spec. The README, which the principles spec's landing table also names, links to the
guide and holds no copy.

The "Set up a server" guide opens with "What you need", a linkable first section that Home and the
is/isn't page both link to. It lists the VPS, the domain, the SMTP provider, the backup remote, and
storage sized in albums rather than bare gigabytes. The guide ends with configuring and testing a
backup, because the server's store is the library's record. Restore stays in "Run your server".

The "Run your server" guide carries a security and privacy section. It says what the server exposes
and how to harden it, and it lists every outbound call the server makes, such as MusicBrainz
lookups, mail, and the backup remote. It also carries a short "Leaving dubplate" section that says
what is on disk and in what layout.

The "For members" guide lists a small set of OpenSubsonic apps tested against 0.8, by platform, each
with the date tested and a link to the app. Members sign in by magic link and mint an app password
per device (streamer spec, "Identity: app passwords minted from the roster account"), so the page
says how an app signs in. An untested app never joins the list.

The comparison page helps a reader choose, and it never disparages another project. It describes
each project in its own terms: what it's built for, who it suits, and what it does well. It then
says plainly when that project is the better fit. Every claim about another project cites that
project's own documentation and carries the date it was checked. It compares purposes and
features, never quality, and it makes no claim about another project's code, security, or
maintainers. The Phase A plan picks the projects, starting from the five the prior-art review
recommends (Navidrome, gonic, Jellyfin, Funkwhale, and Lyrion Music Server). The Navidrome entry
says in one sentence that the group's library ran on Navidrome before 0.8 and links to Credits,
which carries the credit itself.

The "Get help" page names where an operator asks for help and what to include in a report. It gives
a security contact with a stated scope, and it exists from the first public day. It states the
contribution policy that ruling 11 sets, and its channels follow open ruling 12. Its
troubleshooting list covers the common failures (logs, mail
not arriving), which is why the site carries no separate FAQ.

## Credits

The Credits page has three parts, and none is site-owned, so none drifts from what dubplate
ships.

The software credits are a dubplate-owned guide, `docs/guide/credits.md`. It credits beets as
instrumental: the matcher adopts beets' matching behavior, and pinned beets is the oracle it is
measured against (greenfield build spec, "The relationship with beets outlives the launch").
beets also ran the group's imports on the interim estate. It credits Navidrome as an inspiration
and as the server the group ran before 0.8. It says Navidrome and gonic are read for ideas and
never copied, as the streamer spec's "License" section requires of both GPL projects. It then
names by hand the other projects that shaped dubplate: the adopted rows of that spec's prior-art
table, the data services the server queries (MusicBrainz, whose client is
`internal/engine/musicbrainz`), the source of derived data (Python Unidecode, per ROADMAP
`give-back`), and the font faces in `web/static/fonts/FONTS.md`. The guide is a curated
acknowledgment, so the gate checks only that it names every face in `FONTS.md`.

The third-party notices are generated, and they are the complete list. They cover every Go module
the server binary links and every npm package whose code the web bundle ships, each with its version
and the license text read from the package's own files. The Go half comes from `go list -deps -f
'{{with .Module}}{{.Path}}{{end}}' ./cmd/dubplate-server`, run once per release target's `GOOS` and
`GOARCH`, because the linked set depends on the target: a Windows build adds
`github.com/inconshreveable/mousetrap`. On 2026-09-25 that command listed the same 14 modules that
`go version -m` read back from a built binary, including indirect modules such as
`github.com/spf13/pflag` and `modernc.org/libc`. The full module graph from `go list -m all` isn't
the input, because on Linux it also lists 24 modules that the binary never links, such as
`golang.org/x/tools`. The web half comes from the bundle's module graph at build time, through the
bundler's `getModuleIds()` plugin hook (rolldown's `PluginContext`), plus the packages the
stylesheet pulls in through `@import` and `@plugin` (`tailwindcss` and `daisyui` in
`web/src/app.css`). The input is that graph and never `web/package.json`, because every package
there is a `devDependency`, including Svelte, SvelteKit, and daisyUI, whose code still ships. The
Phase A plan proves both halves against a real build. The notices list travels with each release and
through the export manifest, and the Credits page links to it.

The photo credits are generated too. The photo curation spec gives every picked photo a record
under `photos/` with the photographer, the licensor's requested credit, the source title, the
license, and the source URL, and it assigns the CREDITS generator to rung 7 (photo curation spec,
"Out of scope"). The site's photo credits are that generator's output, vendored through the
export manifest, and the site keeps no second list. The site shows only picked photos from
dubplate's pool, served at their slug, so every photo on the site has a credit by construction.

A record with `status: pack` never reaches the site as an image, because the pack holds NC and
dubplate-only grants. A photographer who grants a shipped license class by letter lands as a
picked record like any other, so their credit appears the same way. A pack grantor also lets
dubplate use licensed photos (owner ruling 9). Once the pack ships, the generator's output carries
a name-only line for each record in `pack.json`, built from its `photographer` and `credit`
fields, with no image, so that credit also comes from the pack's own records. Rung 7's plan
receives that need beside the CREDITS path. Until rung 7 lands the generator, the site carries no
photos.

## Keeping the site in step with dubplate

dubplate owns six kinds of file the site shows:

- The theme tokens.
- The font files, with `FONTS.md` and the OFL texts.
- The guides in `docs/guide/`.
- The changelog, `CHANGELOG.md` at the dubplate repo root, with its upgrade notes.
- The generated references: the reject, configuration, and command references, the third-party
  notices, and the photo credits.
- The picked photos, whose `.webp` files join the manifest as a class when rung 7 lands the
  CREDITS generator.

dubplate lists every exported path in one export manifest. The manifest settles which files
travel, in a place dubplate's own gate can check.

The theme files that travel are all seven in `web/src/lib/themes/`: `dubplate.css` (radius,
border, depth, and the font stacks), `label.css`, `label-night.css`, `system.css`, `inks.css`,
`surface.css`, and `fonts.css`. The font files travel from `web/static/fonts/` with `FONTS.md`
and the OFL texts in `licenses/`, because the OFL requires its license to travel with the fonts.
`fonts.css` loads the faces from the root path `/fonts/`, so the site serves them at `/fonts/`.

The token files move unedited, but they don't theme a cairn site on their own. dubplate's theme
blocks declare no `default` or `prefersdark` flag, and daisyUI applies a block at `:root` only
under those flags (`daisyui/theme/index.js`). dubplate's app avoids the gap by stamping
`data-theme` on the server. The cairn template also names its own themes, reads
`--font-display` and `--font-body` where dubplate declares `--font-sans` and `--font-serif`, and
reads `--cairn-<status>-ink` where dubplate declares `--ink-<status>`, with `--ink-accent` also
serving success (mechanics review, M1; visual identity spec, the ink table). So
the site owns one adapter file, outside the lock. It sets the default and dark themes, renames the
theme toggle's two names, and maps cairn's font and ink roles onto dubplate's. The adapter holds
only references such as `var(--ink-info)`, never values, so the one-way drift rule still holds.

The site repo carries a sync script, run as `npm run sync-dubplate -- <ref>`, against a local
dubplate checkout. The site's build never fetches from dubplate, which stays private until Phase B.

### What each check proves

The sync reads only git objects. It resolves the ref to a commit and exports the manifest's paths
from that commit, never from the working tree, so a dirty or wrong checkout can't leak in. A
release tag is an annotated tag named `v<major>.<minor>.<patch>` with no pre-release suffix, and
the release checklist cuts tags to that shape. A 0.x tag of that shape is a release tag. Owner
ruling 3's release-tag pin binds from Phase B, which requires the 0.8 tag. The sync accepts any
commit in Phase A, since no dubplate release tag exists yet, and records whether the ref is a
release tag. It refuses a
lightweight tag, and it refuses when the lock already names the same tag at a different commit. It
stages the export and swaps it in whole, writing the lock last, so an interrupted sync leaves the
old tree and the old lock together. The lock, `dubplate.lock.json`, records the ref, the commit,
whether it is a release, the classes the manifest declares at that commit, and each file's
SHA-256. The vendored files
and the lock are committed to the site repo.

The site's lock check proves that the vendored tree is exactly what a sync wrote. It fails when the
lock is missing or doesn't parse, when a class the manifest declares is empty, when a locked file is
missing, when an unlocked file sits in the vendor tree, and when a hash doesn't match. It runs in
the build that produces the deployed site, so no path deploys a mismatched tree: not an editor
commit, a sync pushed without the gate, or a push from another machine. The vendor directories sit
outside every content directory the site's cairn config names, so the editor never lists a vendored
page. If cairn's template offers no build hook for the check, the gap is filed as a trial finding
for cairn and is never patched locally. The lock doesn't prove provenance, because someone could
edit a file and its hash together. A `--verify` mode of the sync, run on the workstation, re-exports
the locked commit and diffs it.

dubplate's gate proves that what it exports is current and true to the code at every commit. It
does so in four legs.

The export leg reads every file the export manifest lists. It fails when a listed path doesn't
exist. It also fails on an IPv4 literal outside the RFC 5737 documentation ranges, a hostname or
an email address outside an allow-list of example and upstream domains, or a workstation home path
in any exported file. It names no site value itself, since a deny-list of the box's address would
be one more site-specific value outside `internal/config`.

The regenerate leg covers the generated pages and the third-party notices, which are committed. It
regenerates each one and fails on any diff, so no commit carries a stale reference or an
incomplete notices list. The web half of the notices regenerates from the bundle build that the
gate's web leg already runs.

The guide leg fails when a command, flag, or config key named in code formatting in `docs/guide/`
is absent from the command tree or the key registry, which the generators already walk. It reads
a code span as a command when it starts with the root command's name or with a binary name on the
planned-names list, as a flag by its `--name` shape, and as a key by the env wall's `DUBPLATE_`
prefix or the store-homed key shape, which the Phase A plan quotes from `internal/config`. The
guides are drafts against the planned 0.8 surface, so the leg also reads a dubplate-owned
planned-names list, kept beside the export manifest and not exported. The list carries the 0.8
binary name and every planned command, flag, and key, and each entry names the rung that lands
it. A planned name passes while it is absent, and the leg fails on an entry whose name now exists,
so every code-formatted name is either real or listed. The list is empty at the 0.8 release tag.

The credits leg fails when a face in `web/static/fonts/FONTS.md` goes unnamed in
`docs/guide/credits.md`. Completeness for code comes from the notices list, which the regenerate
leg keeps current.

Every vendored and generated page shows "Describes dubplate `<ref>`" from the lock, and a page
from a non-release commit says it previews an unreleased version. A stale pin is then visible to
the reader. The Phase B deploy refuses a lock whose ref isn't a release tag. The site's gate
also fails when the newest release-notes post names a version newer than the lock's tag, so a
release announced on the site can't go out with an older guide. The site's build can't read
dubplate's tags, since it fetches nothing from dubplate, so the post is the signal the gate
reads. Geoff
ruled this pair of checks, option (b) of contract finding C2, on 2026-09-25. The step that bumps
the site's pin belongs in the tool that cuts dubplate releases, which prints the sync command for
the new tag. Until that tool exists, dubplate's release checklist carries the step, and the Phase
A dubplate lane creates that checklist.

The site describes one version, the latest release. An operator on an older release reads the
changelog's upgrade notes, which cover every version. The public source repository (ruling 11)
holds the guides at each earlier tag, and the site links there.

Drift runs one way only. When the site needs a token dubplate lacks, the token is added in
dubplate first, and the site picks it up at the next sync.

## Generated pages

The generated pages exist so they can't drift from the code, and each has a source and a rung.

The reject page builds from the greenfield `internal/reject`, the reject copy contract that rung 7
adopts verbatim from the string authority (greenfield build spec, rung 7). Today the package exists
only under `_salvage/`, which is never imported, and its reasons are a `switch` with no exported
enumerator. The generator needs the reasons as an enumerable set, with a canonical wording for each
parameterized reason, so rung 7's plan receives that need with the generator as its named consumer.
The reject generator lands after rung 7.

The configuration reference covers every key the config registry declares. It needs a
description per key, which forward hook 18 adds at rung 14 (streamer spec, "Forward hooks"), so
the configuration generator is one of hook 18's consumers and lands after rung 14. Settings held
in the other configuration homes are covered by Run your server's prose.

The command reference walks the server binary's cobra tree from its root. The binary is
`dubplate-server` today (`cmd/dubplate-server/main.go`), and the page names whatever 0.8 ships.
The guides call the setup commands `init` and `upgrade` as the streamer spec defines them
("Hosting and setup"), under the binary name 0.8 ships, which the planned-names list carries until
then.
The `cmd/harness` tree is excluded. The Phase A plan weighs `github.com/spf13/cobra/doc` as a
prior-art row.

The registry constructor and the root command are unexported in the server's command package, so
the generators live beside them, and nothing is exported for them. The Phase A plan picks the
form. Each generator is also weighed against the 0.8 forward-hooks list, as every dubplate plan
is.

## The cairn trial

The site's build is a trial of cairn's docs for a brand-new site. The build follows cairn's docs
literally, as a stranger would, starting from the site operator's and site designer's entry points
as the docs reset publishes them. It doesn't borrow setup knowledge from the other cairn sites on
this workstation. The adapter, the vendored-page route, the sync script, and the lock check are
code, so those steps are expected to leave the documented operator path, and they aren't logged as
friction by default.

The trial record lives in the site repo. It has one entry for every docs page the build follows,
including pages that caused no friction, so an empty record can't read as a clean trial. Each
entry names the step, the page, what happened, and what the docs led the builder to expect. Each
place the builder steps off the documented path is logged the same way. The findings go to cairn
through the engine consultation path (`engine-consult`).

## Sequence

The build waits for its trigger: cairn's docs reset has closed, and `create-cairn-site` is
published on npm. Today `npm view create-cairn-site` returns 404, and cairn's roadmap treats
publication as its own decision, so the docs reset alone doesn't fire the trigger. Nothing is
built before then. When the spec is approved, ROADMAP gains a `project-site` entry under Planned
with that trigger.

Phase A builds the site. It is its own initiative, outside the rung ladder, in its own repo. The
`docs/guide/` files, the export manifest, the generators, and dubplate's gate legs are the only
part that touches the dubplate tree, and they land as a lane under dubplate's lane rules. Phase A
is done when:

- The site builds and deploys to a `workers.dev` address that answers with `X-Robots-Tag: noindex`
  on `workers.dev` hosts only, checked by fetching a page. The address isn't barred in
  `robots.txt`, because a crawler barred there never reads the header.
- The sync and the lock check exist, and the Phase A plan proves each failure the lock check
  names can fire.
- dubplate's export manifest and its four gate legs exist, each proven able to fail.
- Every page in the site map exists as a route and is listed in the sitemap. The guides are
  drafts against the planned 0.8 surface. A generated page is real output wherever its source
  exists at the pinned commit. Otherwise the site carries its own placeholder that says which rung
  the page waits for.
- The is/isn't lines and the legal note have moved into `docs/guide/`, and the principles spec
  points at them.
- Guide links written for the dubplate repo resolve to site routes.
- The rendered pages meet WCAG AA contrast in both themes, checked by the site's gate over its
  built CSS, since the site's own markup doesn't follow dubplate's ink rule.
- The trial record is committed, and its findings are filed through `engine-consult`.
- dubplate's changelog exists with its entry format, and the release checklist exists.
- The site gate's comparison-date check exists and is proven able to fail.

The reject and configuration generators land after rungs 7 and 14, which Phase A may precede. The
ROADMAP `project-site` entry carries two owed dubplate lanes, each triggered by a merge: the reject
generator when rung 7 merges and the configuration generator when rung 14 merges. Each lane adds
its generator's regenerate-and-diff coverage and declares the page's class in the export
manifest. The next sync then replaces the site's placeholder. If Phase A runs after
either rung, its own lane builds that generator.

Phase B publishes the site. Its preconditions are these:

- 0.8 has shipped, and the lock names the 0.8 release tag. The planned-names list is empty.
- The signed 0.8 binaries are published. The public source repository exists as a fresh
  repository from a scrubbed snapshot at the 0.8 tag, and a history-aware secret scan and a read
  for production specifics both pass on that snapshot (ruling 11). The site repository is public.
- The give-back path 1 pass has landed, and the lawyer's confirmation of the distribution reading
  is recorded (ROADMAP, `give-back`). The public setup guide is that release's front door.
- All three generated pages carry real output.
- The app list in For members is tested against 0.8.
- Every comparison claim's checked date is under six months old. The site's gate checks the dates
  before each later site release too, since other projects change on their own schedules.
- The security contact on Get help is live.
- The domain is ruled and registered. The production domain serves without `noindex`, and the
  `workers.dev` address keeps it.

## Open rulings

These continue the owner rulings' numbering.

10. **The domain.** It's `dubplate.dev` or `dubplate.studio`. Cloudflare Registrar sells both at
    standard tier, `.dev` at $12.20 a year and `.studio` at $31.20 a year for registration and
    renewal. The source is the Registrar's domain-check API
    (`POST /accounts/{account_id}/registrar/domain-check`), run by the conductor on 2026-09-25,
    which returned both as registrable. The ruling is due before Phase B registers the domain,
    and Geoff confirms the purchase then.
11. **How a stranger gets the software at 0.8. Ruled (Geoff, 2026-09-25): public binaries and
    public source.** 0.8 publishes signed release binaries, which the `upgrade` command already
    needs (streamer spec, "Hosting and setup"), and a public source repository under the MIT
    license. Every project on the comparison page publishes its source, and the ratified
    principles spec is written as though the source is reachable ("Anyone with the source can
    remove any limit this spec describes"). Opening the source is irreversible, so two conditions
    bound what is published.
    - The public repository is a fresh repository created from a scrubbed snapshot at the 0.8
      tag. The working repository and its history are never published. Its history carries
      production specifics (`CLAUDE.md` names the box's address) and years of working notes, and
      the gate's gitleaks leg scans only the working tree (`--no-git`), so no check has ever read
      the history.
    - The "Get help" page states the contribution policy: support is best-effort, issues are
      welcome, and a pull request starts with a discussion. The refused-features list in dubplate's
      `CLAUDE.md` is the written answer to requests the project won't take. This is the
      maintainership decision ROADMAP's `oss-pair` asks for, taken for the server.

    The site repository opens at the same time. Both `give-back` obligations (path 1 and the
    lawyer's confirmation) still apply before the release. Give-back's modules publish as their
    own repositories, as that entry already plans.
12. **A community channel at 0.8.** Get help always carries a security contact, and the public
    source brings a public issue tracker (ruling 11). The fork is whether dubplate
    also opens a community channel, such as a forum or chat, which costs Geoff's time to tend.
    - Yes: Get help names the channel, and someone tends it.
    - No: the issue tracker and the security contact are the only channels.

    Recommendation: no at 0.8. The issue tracker is the channel, under the contribution policy
    that ruling 11 puts on Get help.
13. **A "Coming later" list.** A reader with an MP3 library needs to know that other formats are
    planned for 1.0, not refused.
    - Yes: the is/isn't page carries the list, owned in `docs/guide/` with the rest of the page.
    - No: the site says nothing about later versions.

    Recommendation: yes, one undated list on the is/isn't page. It answers the fit question
    without a date.

## Out of scope

The site doesn't host downloads, a demo server, or a members' area. A demo server would
contradict the "isn't" line about a public music site. The site doesn't document the 0.x
pre-release. Photographs appear on the site only from dubplate's picked pool, under the shipped
license classes the photo curation spec sets. The photo curation spec is the later ruling, so it
governs where the visual identity spec differs.
