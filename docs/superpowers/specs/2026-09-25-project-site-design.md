# The dubplate project site

Design spec. Revision 1, 2026-09-25, drafted from a brainstorm with Geoff. Status: **awaiting
adversarial review, then Geoff's read.** Two owner rulings stay open (see "Open rulings").

dubplate needs a public home where a stranger learns what it is, decides whether it fits their
group, and sets up a server. This spec fixes who the site serves, where it lives, which pages it
carries, who owns each page's content, and how the site stays in step with dubplate's code and
look. The site is a standard cairn site in its own repo. dubplate owns the theme tokens and the
guides, and the site carries a pinned, hash-checked copy of both. Building the site doubles as a
trial of cairn's documentation for a brand-new site, so the build waits until cairn's docs reset
is complete. The site goes public when dubplate 0.8 ships, and every page describes 0.8.

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

## Readers

The primary reader is the person who sets up and runs a server for their group. They're
comfortable with a terminal and a VPS, and they haven't read dubplate's code or specs. The
secondary readers are a member who has been invited to a group, and a person deciding whether
dubplate fits their group at all. Every page names which of these readers it serves.

## Where the site lives

The site is a standard cairn site in its own repo, `~/Projects/dubplate-site`. It's created with
`npx create-cairn-site` and deployed to Cloudflare the way cairn's admin track describes. Keeping
it out of the dubplate repo means cairn's publish flow, which commits edits to the site repo's
main branch, never touches dubplate's tree or its lanes.

## Pages

Each page has one owner. Site-owned pages are edited in cairn's editor. dubplate-owned pages are
written in dubplate and vendored into the site, where cairn's editor can't change them. Generated
pages are built from dubplate's code and vendored the same way.

| Page | Owner | Reader | Carries |
|---|---|---|---|
| Home | site | all | What dubplate is in one paragraph, the look, and one path each for setting up a server and for an invited member |
| What dubplate is, and isn't | dubplate | deciding | The "dubplate is" and "dubplate isn't" lines |
| How dubplate compares | site | deciding | Where dubplate fits among other self-hosted music servers, and when another one fits better |
| Set up a server | dubplate | operator | What you need, installing, `dubplate init`, the first admin, and the first member |
| Run your server | dubplate | operator | Members, collections, backups and restore, and `dubplate upgrade` |
| For members | dubplate | member | Signing in, adding music, and listening with an OpenSubsonic app |
| Why an album was turned away | generated | member | Every reject reason, in the wording the uploader and the emails use |
| Configuration reference | generated | operator | Every config key |
| Command reference | generated | operator | Every command and flag |
| Release notes | site | operator | One cairn post per release |
| About | site | deciding | Who makes dubplate, the MIT license, the stance on buying music, and photo credits |

The "is" and "isn't" lines live today in the group-scale principles spec
(`2026-09-25-group-scale-principles-design.md`, "dubplate is, and dubplate isn't"), which says the
website publishes them when 0.8 ships. They change when a ruling changes, such as the share
link's week and month, so their source moves into `docs/guide/` and the principles spec points at
it. That move happens in Phase A, after the principles spec's own rulings settle.

The comparison page helps a reader choose, and it never disparages another project (Geoff,
2026-09-25). It describes each project in its own terms: what it's built for, who it suits, and
what it does well. It then says plainly when that project is the better fit. Every claim about
another project cites that project's own documentation and carries the date it was checked, so
the page is re-verified before each release that touches it. The page credits Navidrome, which
served the group's library before 0.8, by name. It compares purposes and features, never quality,
and it makes no claim about another project's code, security, or maintainers. The Phase A plan
picks the projects from the ones a person choosing a self-hosted music server is likely to weigh.

The three generated pages exist so they can't drift from the code. The reject page builds from
`internal/reject`, whose wording is already contract across the uploader and the emails. The
configuration reference builds from the config key registry, and the command reference builds
from the cobra command tree. Each needs a small generator in dubplate. The Phase A plan weighs
each generator against the 0.8 spec's forward-hooks list, as every dubplate plan does.

`dubplate init` and `dubplate upgrade` are named here because ROADMAP moves both into 0.8
(`packaging`). The guide describes whatever 0.8 ships under those names.

## Keeping the site in step with dubplate

dubplate owns two things the site shows. The theme tokens live in `web/src/lib/themes/`, as
DaisyUI 5 `@plugin "daisyui/theme"` blocks, with the font files they reference. The guides live
in a new `docs/guide/` directory, written to the Google developer documentation style. A cairn
site's theme redeclares the same DaisyUI theme blocks over its chassis, so the token files move
as files, with no translation step.

The site repo carries a sync script, run as `npm run sync-dubplate -- <tag>`. It exports the
owned files from a dubplate release tag in a local dubplate checkout, then writes them into a
vendor directory in the site. It also writes `dubplate.lock.json`, which records the tag, the
commit, and each file's SHA-256. The vendored files and the lock are committed to the site repo.
The dubplate repo is private, so the site's build never fetches from it.

The site's gate fails when a vendored file's hash doesn't match the lock. That catches a hand
edit to a vendored file. Drift runs one way only: when the site needs a token dubplate lacks,
the token is added in dubplate first, and the site picks it up at the next sync. dubplate's
release checklist gains a step that bumps the site's pin, so the guide on the site always matches
a tagged release.

The Phase A plan decides exactly which theme files travel. The palette, ink, and font files
travel. The app's component layer (`surface.css`, `system.css`) probably doesn't, because it
styles the app's own markup.

## The cairn trial

The site's build is a trial of cairn's docs for a brand-new site. The build follows cairn's
admin track literally, starting from `docs/admin/README.md`, as a stranger would. It doesn't
borrow setup knowledge from the other cairn sites on this workstation. Each friction point is
logged in a trial record in the site repo, with the step, the doc page, what happened, and what
the docs led the builder to expect. Each place the builder steps off the documented path is
logged the same way. The findings go to cairn through the engine consultation path
(`engine-consult`). The trial record is a deliverable of Phase A in its own right.

## Sequence

The build waits for its trigger, the close of cairn's docs reset. Nothing is built before then.

Phase A builds the site: the cairn trial, the scaffold, the sync script and its gate, the three
generators, the landing page, and a preview deploy on a `workers.dev` address marked `noindex`.
Phase A is its own initiative, outside the rung ladder, in its own repo. The `docs/guide/` files
and the generators are the only part that touches the dubplate tree, and they land as a lane
under dubplate's lane rules.

Phase B publishes the site. Its preconditions are that 0.8 has shipped, that a stranger can reach
the source or the binaries, that the guides describe the tagged 0.8 release, and that the domain
is ruled and registered.

## Open rulings

- **The domain.** It's `dubplate.dev` or `dubplate.studio`. Cloudflare Registrar sells both at
  standard tier, `.dev` at $12.20 a year and `.studio` at $31.20 a year (authoritative check,
  2026-09-25). Due before Phase B registers it, and the purchase is confirmed with Geoff then.
- **How a stranger gets the software.** The dubplate repo is private today. Open-sourcing is the
  `oss-pair` entry's question and distribution is `packaging`'s, and this spec decides neither.
  Due before Phase B.

## Out of scope

The site doesn't host downloads, a demo server, or a members' area. It doesn't document the 0.x
pre-release. Photographs from the photo curation work may appear on the site, but only under the
redistribution-permitting licenses the visual identity spec already requires.
