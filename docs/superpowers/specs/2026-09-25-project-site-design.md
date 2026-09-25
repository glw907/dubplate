# The dubplate project site

Design spec. Revision 2, 2026-09-25, drafted from a brainstorm with Geoff. Status: **awaiting
Geoff's read.** Four owner rulings stay open (see "Open rulings"). Revision 1 (`e91f528`) went
through five review lenses (contract, drift, consistency, mechanics, prior art) under
`docs/superpowers/research/2026-09-25-project-site-review-*.md`. The fold record is
`docs/superpowers/research/2026-09-25-project-site-fold.md`.

Revision 2 changelog: folds the five reviews, defines what the sync reads and what each gate
proves, and adds the Credits, Get help, and Bring your library pages (owner rulings 8 and 9).

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
`npx create-cairn-site` and deployed to Cloudflare the way cairn's docs describe. Keeping it out
of the dubplate repo means cairn's publish flow, which commits edits to the site repo's main
branch, never touches dubplate's tree or its lanes.

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
| Credits | dubplate and generated | all | The software dubplate builds on or learned from, and the photo credits |
| About | site | deciding | Who makes dubplate, the MIT license, and the stance on buying music |

The "is" and "isn't" lines and the legal note live today in the group-scale principles spec
(`2026-09-25-group-scale-principles-design.md`, "dubplate is, and dubplate isn't" and "A legal
note"). Its landing table routes both to the website. They change when a ruling changes, such as
the share link's week and month, so their source moves into `docs/guide/`, and the principles
spec points at it. The move happens in Phase A, once the principles spec is ratified. After the
move, no second copy of the lines remains in the principles spec.

Set up a server opens with "What you need", a linkable first section that Home and the is/isn't
page both link to. It lists the VPS, the domain, the SMTP provider, the backup remote, and
storage sized in albums rather than bare gigabytes. The guide ends with configuring and testing a
backup, because the server's store is the library's record. Restore stays in Run your server.

Run your server carries a security and privacy section. It says what the server exposes and how
to harden it, and it lists every outbound call the server makes, such as MusicBrainz lookups, mail,
and the backup remote. It also carries a short "Leaving dubplate" section that says what is on
disk and in what layout.

For members lists a small set of OpenSubsonic apps tested against 0.8, by platform, each with the
date tested and a link to the app. Members sign in by magic link and mint an app password per
device (streamer spec, "Identity: app passwords minted from the roster account"), so the page
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

Get help names where an operator asks for help and what to include in a report. It gives a
security contact with a stated scope, and it exists from the first public day. The channels it
names follow open rulings 11 and 12. Its troubleshooting list covers the common failures (logs,
mail not arriving), which is why the site carries no separate FAQ.

## Credits

The Credits page has two parts, and neither is site-owned, so neither drifts from what dubplate
ships.

The software credits are a dubplate-owned guide, `docs/guide/credits.md`. It credits beets as
instrumental: the matcher adopts beets' matching behavior, and pinned beets is the oracle it is
measured against (greenfield build spec, "The relationship with beets outlives the launch").
beets also ran the group's imports on the interim estate. It credits Navidrome as an inspiration
and as the server the group ran before 0.8. It says Navidrome and gonic are studied for behavior
and never copied, as the streamer spec's prior-art table requires. It then names the other
projects the tagged release builds on or learned from, drawn from that table's adopted rows and
from `go.mod`. At revision 2 that set includes the OpenSubsonic specification, MusicBrainz,
Picard's tag-mapping tables, Python Unidecode (the source of the transliteration table), the Go
modules `go.mod` requires directly (cobra, go-mail, `golang.org/x/time`, and modernc SQLite), the
web stack (SvelteKit, Svelte, DaisyUI, and Tailwind), and the three OFL fonts named in
`web/static/fonts/FONTS.md`. The page credits each project with its license, checked when the
page is written. dubplate's gate fails when a direct requirement in `go.mod` or a face in
`FONTS.md` goes unnamed in the guide, so a new dependency can't ship uncredited.

The photo credits are generated. The photo curation spec gives every picked photo a record under
`photos/` with the photographer, the licensor's requested credit, the source title, the license,
and the source URL, and it assigns the CREDITS generator to rung 7 (photo curation spec, "Out of
scope"). The site's photo credits are that generator's output, vendored through the export
manifest. The site never keeps a second list. The site shows only picked photos from dubplate's
pool, served at their slug, so every photo on the site has a credit by construction. A record with
`status: pack` never reaches the site, because the pack holds NC and dubplate-only grants. A
photographer who grants a license by letter lands as a picked record like any other, so their
credit appears the same way. Until rung 7 lands the generator, the site carries no photos.

## Keeping the site in step with dubplate

dubplate owns three kinds of file the site shows: the theme tokens, the guides, and the generated
references with the photo credits. dubplate lists every exported path in one export manifest.
The manifest settles which files travel, in a place dubplate's own gate can check.

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
reads `--cairn-<status>-ink` where dubplate declares `--ink-<status>` (mechanics review, M1). So
the site owns one adapter file, outside the lock. It sets the default and dark themes, renames the
theme toggle's two names, and maps cairn's font and ink roles onto dubplate's. The adapter holds
only references such as `var(--ink-info)`, never values, so the one-way drift rule still holds.

The site repo carries a sync script, run as `npm run sync-dubplate -- <ref>`, against a local
dubplate checkout. The dubplate repo is private, so the site's build never fetches from it.

### What each check proves

The sync reads only git objects. It resolves the ref to a commit and exports the manifest's paths
from that commit, never from the working tree, so a dirty or wrong checkout can't leak in. It
accepts any commit in Phase A, since no dubplate release tag exists yet, and records whether the
ref is a release tag. It refuses a lightweight tag, and it refuses when the lock already names the
same tag at a different commit. It stages the export and swaps it in whole, writing the lock last,
so an interrupted sync leaves the old tree and the old lock together. The lock, `dubplate.lock.json`,
records the ref, the commit, whether it is a release, and each file's SHA-256. The vendored files
and the lock are committed to the site repo.

The site's lock check proves that the vendored tree is exactly what a sync wrote. It fails when
the lock is missing or doesn't parse, when a class the manifest declares is empty, when a locked
file is missing, when an unlocked file sits in the vendor tree, and when a hash doesn't match. It
runs in the build that produces the deployed site, so no path deploys a mismatched tree: not an
editor commit, a sync pushed without the gate, or a push from another machine. The vendor
directories sit outside every content directory the site's cairn config names, so the editor
never lists a vendored page. If cairn's template offers no build hook for the check, that is a
trial finding for cairn, not a local patch. The lock doesn't prove provenance, because someone
could edit a file and its hash together. A `--verify` mode of the sync, run on the workstation,
re-exports the locked commit and diffs it.

dubplate's gate proves that what it exports is current and true to the code at every commit.
Every manifest path exists. The generated pages are committed, and a gate leg regenerates them and
fails on any diff, so no commit carries a stale reference. A guide leg fails when a command, flag,
or config key named in code formatting in `docs/guide/` is absent from the command tree or the
key registry, which the generators already walk. The same leg fails on the box's address, the
group's hostnames, an email address outside an allow-list, or a workstation home path in any
exported file.
The credits check in "Credits" is the fourth leg.

Every vendored and generated page shows "Describes dubplate `<ref>`" from the lock, and a page
from a non-release commit says it previews an unreleased version. A stale pin is then visible to
the reader. The Phase B deploy refuses a lock whose ref isn't a release tag. The step that bumps
the site's pin belongs in the tool that cuts dubplate releases, which prints the sync command for
the new tag. Until that tool exists, dubplate's release checklist carries the step, and the Phase
A dubplate lane creates that checklist.

The site describes one version, the latest release. An operator on an older release reads the
changelog's upgrade notes, which cover every version, and once the source is public (open ruling
11) the site links to the guides at each earlier tag.

Drift runs one way only. When the site needs a token dubplate lacks, the token is added in
dubplate first, and the site picks it up at the next sync.

## Generated pages

The generated pages exist so they can't drift from the code, and each has a source and a rung.

The reject page builds from the greenfield `internal/reject`, the reject copy contract that rung 7
adopts verbatim from the string authority (greenfield build spec, rung 7). Today the package
exists only under `_salvage/`, which is never imported, and its reasons are a `switch` with no
enumerator. The generator needs the reasons as an enumerable set, with a canonical wording for
each parameterized reason, so rung 7's plan receives that need with the generator as its named
consumer. The reject generator lands after rung 7.

The configuration reference covers every key the config registry declares. It needs a
description per key, which forward hook 18 adds at rung 14 (streamer spec, "Forward hooks"), so
the configuration generator is one of hook 18's consumers and lands after rung 14. Settings held
in the other configuration homes are covered by Run your server's prose.

The command reference walks the server binary's cobra tree from its root. The binary is
`dubplate-server` today (`cmd/dubplate-server/main.go`), and the page names whatever 0.8 ships.
The `cmd/harness` tree is excluded. The Phase A plan weighs `github.com/spf13/cobra/doc` as a
prior-art row.

The registry constructor and the root command are unexported in the server's command package, so
the generators live beside them, and nothing is exported for them. The Phase A plan picks the
form. Each generator is also weighed against the 0.8 forward-hooks list, as every dubplate plan
is.

The guides call the setup commands `init` and `upgrade` as the streamer spec defines them
("Hosting and setup"). The streamer spec writes `dubplate init`, and the server binary is named
`dubplate-server` today. The guide uses the name 0.8 ships under.

## The cairn trial

The site's build is a trial of cairn's docs for a brand-new site. The build follows cairn's docs
literally, as a stranger would, starting from the site operator's and site designer's entry
points as the docs reset publishes them. It doesn't borrow setup knowledge from the other cairn
sites on this workstation. The adapter, the vendored-page route, the sync script, and the lock
check are code, so those steps are expected to leave the operator track, and they aren't logged
as friction by default.

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
  exists at the pinned commit, and a placeholder that says which rung it waits for otherwise.
- Guide links written for the dubplate repo resolve to site routes.
- The rendered pages meet WCAG AA contrast in both themes, checked by the site's gate over its
  built CSS, since the site's own markup doesn't follow dubplate's ink rule.
- The trial record is committed, and its findings are filed through `engine-consult`.
- The is/isn't lines and the legal note have moved into `docs/guide/`, once the principles spec is
  ratified.

Phase B publishes the site. Its preconditions are these:

- 0.8 has shipped, and the lock names the 0.8 release tag.
- A stranger can get the software, as open ruling 11 decides.
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
    standard tier, `.dev` at $12.20 a year and `.studio` at $31.20 a year (authoritative check,
    2026-09-25). Due before Phase B registers it, and the purchase is confirmed with Geoff then.
11. **How a stranger gets the software at 0.8.** The dubplate repo is private today, and no
    ratified document owns opening it. ROADMAP's `oss-pair` covers only the uploader and the
    pings Worker. ROADMAP sequences `packaging` after the organizer, which runs after launch, so
    waiting on `packaging` would move Phase B past 0.8. Recommendation: (a), because 0.8's
    `dubplate upgrade` already downloads a release and verifies it against a signed release
    manifest (streamer spec, "Hosting and setup").
    - (a) 0.8 ships its own minimal public channel: the public source and the signed release
      binaries `upgrade` needs. `packaging` keeps the richer installers. Phase B stays at 0.8,
      and ROADMAP records the channel in `streamer-0-8`.
    - (b) Phase B waits for `packaging`. The site publishes after 0.8, and owner ruling 5 is
      amended.
12. **A community channel at 0.8.** Get help always carries a security contact, and a public issue
    tracker follows the source when it opens. The fork is whether dubplate also opens a community
    channel, such as a forum or chat, which costs Geoff's time to tend. Recommendation: no at 0.8;
    Get help says support is best-effort through the issue tracker.
    - Yes: Get help names the channel, and someone tends it.
    - No: the issue tracker and the security contact are the only channels.
13. **A "Coming later" list.** A reader with an MP3 library needs to know that other formats are
    planned for 1.0, not refused. Recommendation: yes, one undated list on the is/isn't page. It
    answers the fit question without a date.
    - Yes: the is/isn't page carries the list, owned in `docs/guide/` with the rest of the page.
    - No: the site says nothing about later versions.

## Out of scope

The site doesn't host downloads, a demo server, or a members' area. A demo server would
contradict the "isn't" line about a public music site. The site doesn't document the 0.x
pre-release. Photographs appear on the site only from dubplate's picked pool, under the shipped
license classes the photo curation spec sets, which is the later ruling over the visual identity
spec.
