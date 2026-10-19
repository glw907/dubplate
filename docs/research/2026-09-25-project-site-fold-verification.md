# Project site spec: fold verification

Target: `docs/superpowers/specs/2026-09-25-project-site-design.md` at `73b429f` (revision 2),
against revision 1 (`e91f528`), the fold record
`docs/research/2026-09-25-project-site-fold.md`, and the five review files. Line
numbers below are revision 2's.

## Verdict

The fold is sound in substance. Every major finding closed at a real location, the shared-root
reading (one "What each check proves" section) is the right fix, and the new mechanisms I could
check against source are quoted correctly. Three defects are worth a fold before Geoff's read:
the new guide leg contradicts Phase A's own draft guides (V1), the two late generators have no
owner between Phase A and Phase B (V2), and open ruling 11 bundles an irreversible decision its
evidence doesn't cover (V3). The rest are minor and could equally go to the Phase A plan.

## 1. Did each major finding close?

| Finding | Closed at | Note |
|---|---|---|
| C1 | 182-191 | Closed. The "class the manifest declares" check needs a data source on the site side (V4). |
| C2, D2 | 195-197 | Closed, but the leg as written fails on Phase A's own drafts (V1). |
| C3, M4 | 195-196, 220-225 | Closed. |
| C4 | 272-288 | Closed. |
| D1, M7 | 173-177, 202-207 | Closed. "Release tag" is undefined (V5). |
| D3 | 173-180 | Closed. |
| D4 | 152-166, 284-285 | Closed in part; the refusal is reasoned and the site-side contrast check covers the risk. |
| D5 | 185-189 | Closed. The waymark template already has a `prebuild` script (`templates/waymark/package.json:10`), so a build hook exists. |
| D6 | 197-199 | Closed in part. D6's "state whether the site repo is public" is still unanswered (see V9). |
| K1 | 294-295 | Closed. |
| K2 | 311-321 | Closed as a fork; the fork itself needs a split (V3). |
| K3, M3 | 227-230 | Closed for placement; ownership after Phase A is open (V2). |
| M1 | 158-166 | Closed. |
| M2, K4 | 220-225 | Closed; the routing to rung 7 is in the fold record's owed list. |
| M5 | 263-267 | Closed. |
| P1 to P4 | 64-114, ruling 12 | Closed. |

## 2 to 5. Findings, ranked by consequence

### V1 (major): the guide leg fails on the guides Phase A is required to write

Location: spec:195-197 against spec:280-281 and spec:67-68.

The guide leg fails "when a command, flag, or config key named in code formatting in
`docs/guide/` is absent from the command tree or the key registry". Phase A's done list says
"The guides are drafts against the planned 0.8 surface". Those drafts must name `init`,
`upgrade`, and `import navidrome` (Set up a server, Run your server, Bring your library), and
the streamer spec lands all three at its hosting slice (`2026-09-07-streamer-2-0-design.md:445-446`).
The same holds for any store-homed key a 0.8 rung adds. If Phase A runs before 0.8, which is
the expected order, the leg is red on the first guide commit, and "never commit red" blocks
the lane. The fold introduced this: revision 1 had no guide leg.

Fix: the leg reads one dubplate-owned list of planned names, each with the rung that lands it.
A planned name passes while it is absent. Once it exists in the tree or registry, it must come
off the list, so the leg fails on a stale entry. The list is empty at the 0.8 release tag, and
that joins the Phase B preconditions. This also keeps the check non-vacuous, since every
backticked name is either real or on the list. State in one clause how the leg tells a
command, flag, or key span from other code spans (for example `dubplate-server ...` prefixes,
`--flag` shape, and the registry's key names), or the Phase A plan meets that question as a
planning miss.

### V2 (major): the reject and config generators have no owner after Phase A

Location: spec:220-230, spec:281-282, spec:296.

The reject generator "lands after rung 7" and the config generator "after rung 14". Phase A
may close before either rung, and it then ships placeholders. Phase B requires "All three
generated pages carry real output". No lane, rung, or ROADMAP line is given the work in
between. Rung 7's plan receives only the enumerability need, and hook 18 is built by rung 14
for the settings surface, not the page. As written, nothing fires when rung 7 or rung 14
merges, and the gap would surface only at the Phase B check.

Fix: the ROADMAP `project-site` entry (already owed at approval) carries two owed dubplate
lanes, each triggered by a merge: the reject generator when rung 7 merges and the config
generator when rung 14 merges. Each lane adds its generator's regenerate-and-diff coverage and
replaces the placeholder. If Phase A runs after either rung, its own lane builds that
generator. One sentence in Sequence says so.

### V3 (major): open ruling 11 bundles opening the source with the release channel

Location: spec:311-321.

Option (a) is "the public source and the signed release binaries `upgrade` needs". The only
evidence given is that `upgrade` already verifies a signed release manifest, which argues for
public binaries, not public source. Opening the repo is irreversible, and ROADMAP's nearest
precedent (`oss-pair`, `ROADMAP.md:220-224`) makes "a clear-eyed maintainership decision" a
precondition of publishing even the smaller pair. The repo today also carries production
specifics in tracked files (`CLAUDE.md` names the box address), so "public source" has a
cleanup cost the ruling doesn't show. Other parts of the spec lean on the source half: ruling
12's issue tracker "follows the source when it opens", and older guides are linked "once the
source is public" (spec:209-211).

Fix: split (a) into (a1) public signed binaries, source private for now, and (a2) public
binaries and public source. Name the maintainership question and the pre-publication scrub as
costs of (a2). Say what Get help and the older-version links do under (a1): the security
contact is the only channel, and no earlier-version links. Geoff can then rule the channel and
the source separately, on evidence for each.

### V4 (minor): the site can't know which classes the manifest declares

Location: spec:182-184.

The lock check fails "when a class the manifest declares is empty". The manifest lives in
dubplate, and the site sees only the lock, which records "the ref, the commit, whether it is a
release, and each file's SHA-256" (spec:178-179). Nothing carries the class list across, so
the check has no input and passes vacuously.

Fix: the lock records the manifest's declared classes, or the sync vendors the manifest itself
as a locked file.

### V5 (minor): "release tag" has no definition

Location: spec:175-177, spec:203-204, spec:292.

The sync "records whether the ref is a release tag", and the Phase B deploy refuses a lock
that isn't one, but no predicate says what counts. Rung 16 cuts a 0.x pre-release, and whether
it is a "release tag" decides whether a pre-0.8 lock passes the deploy check. The only
constraint stated is refusing a lightweight tag.

Fix: one sentence, for example "an annotated tag named `v<major>.<minor>.<patch>` with no
pre-release suffix". The release checklist Phase A creates cuts tags to that shape.

### V6 (minor): the photo images themselves aren't exported

Location: spec:139-141 against spec:148-149.

The site "shows only picked photos from dubplate's pool, served at their slug", but the export
names three kinds (tokens, guides, generated references with the photo credits). The `.webp`
files and their records aren't among them, so the site has no locked path to the images whose
credits it vendors.

Fix: add the picked images as a class that joins the manifest when rung 7 lands, from the
same commit as the CREDITS output, so image and credit can't disagree.

### V7 (minor): the letter-grant credit is overstated, against ruling 9's wording

Location: spec:142-144, spec:38-39.

"A photographer who grants a license by letter lands as a picked record like any other." Only
a grant under a shipped class does. The curation spec sends a dubplate-only permission to
`pack.json` (`2026-09-25-photo-curation-design.md:181-182`), and the letters explicitly accept
that outcome (line 382). Such a grantor "lets dubplate use licensed photos" (ruling 9) but gets
no site credit, because `pack` records never reach the site.

Fix: say that grants under a shipped class appear, and state what happens to a pack grantor.
The likely answer is a name-only credit line, which shows no photo. Take it as the spec's call
unless the pack's own CREDITS shape (rung 7 or later) forbids it.

### V8 (minor): the credits leg overclaims

Location: spec:133-134.

"A new dependency can't ship uncredited" holds for Go modules and fonts only. The web bundle
ships inside the binary, and rung 7 adds `tus-js-client` as a runtime dependency (CLAUDE.md,
Stack rules). After give-back path 1, the Unidecode package is fetched at setup
(`ROADMAP.md:159-163`), outside `go.mod`. Neither is checked.

Fix: extend the leg to `web/package.json`'s `dependencies` (it has none today, so the addition
costs nothing until rung 7), or narrow the sentence to "a new Go module or font".

### V9 (minor): the leak leg's deny-list is itself a leak, and the site repo's visibility is unstated

Location: spec:197-199, spec:50-53.

A gate that fails on "the box's address" and "the group's hostnames" must name them in a
tracked file, which is a site-specific value outside `internal/config` (CLAUDE.md's first
paragraph) and is published if ruling 11 opens the source. D6's request to state whether the
site repo is public also went unanswered.

Fix: prefer pattern and allow-list forms: any IPv4 literal outside the RFC 5737 documentation
ranges, and any hostname outside an allow-list of example and upstream domains. Where a literal
deny entry is needed, read it from the workstation's secrets at gate time. Add one line: the
site repo stays private until Phase B, then follows ruling 11.

### V10 (minor): one Phase A done item waits on another spec's ratification

Location: spec:287-288, spec:79-84.

The is/isn't move is a Phase A done item "once the principles spec is ratified". That spec has
rulings 4 to 12 open (`2026-09-25-group-scale-principles-design.md:3`), so Phase A could stay
open on it indefinitely. The principles landing table also routes the lines to the README
(line 585), a second copy this spec doesn't address.

Fix: move the item to the Phase B preconditions, where it has to hold anyway. Name the README
in the erratum: it links to the guide, and it holds no copy of the lines.

### V11 (minor): two new obligations have no creating task

Location: spec:74, spec:298-299.

The dubplate changelog (the D8 decision) and the site gate's comparison-date check are both
new. Neither is in the Phase A done list, while the release checklist is (spec:206-207).

Fix: add both to the Phase A done list: the changelog file exists with an entry format, and
the date check is proven able to fail.

## Question 3: mechanisms checked against source

Verified: the seven theme files and the font layout (`web/src/lib/themes/`, `web/static/fonts/`
with three OFL texts); `daisyui/theme/index.js` gating on `isDefault` and `prefersdark` (lines
14-15, 23, 45); `go.mod`'s four direct requirements; `FONTS.md`'s three faces; the curation
spec's record fields, `pack` handling, letter-grant `source_url`, and rung 7's CREDITS
assignment (lines 18, 156-174, 181-182, 391-396, 400); hook 18's wording (streamer spec
518-519); the `init` and `upgrade` citations ("Hosting and setup", 360-366); the app-password
section title (83); the prior-art table's adopted rows (OpenSubsonic, Picard, `x/time/rate`,
modernc) and the Navidrome and gonic GPL line (185-187); the beets section (greenfield build
spec 185-186); `newRootCmd` and `newProductionRegistry` unexported in `cmd/dubplate-server`;
`cmd/harness` as a second tree; ROADMAP `give-back` and `oss-pair`. On the cairn side:
`--font-display` and `--font-body` in `templates/waymark/src/chassis/tokens.css:72-73`, the
`cairn-dark|cairn` cookie match in `templates/waymark/src/app.html:12`, `--cairn-*-ink` in the
template theme, `create-cairn-site` at `0.0.0`, and a `hooks.server.ts` and `robots.txt` route
for the `noindex` header. The configuration scope matches the homes spec: the registry declares
the env wall and the store-homed server settings (`2026-09-11-configuration-homes-design.md:31-35,
109`).

Not verifiable here: the Registrar prices in ruling 10. They carry over unchanged from revision 1
and cite no source file. MusicBrainz is credited, though it isn't a row in the prior-art table the
spec says the set is "drawn from". That is harmless, but the sentence should say "and the data
services it queries".

## Question 4: the rulings and the settled-by-fold items

The four settled-by-fold items are sound calls, not buried forks. The guide name check follows
from ruling 2's promise, and D2's reviewer agreed. The config scope matches what the registry
actually declares, and member preferences and machine settings aren't operator configuration.
dubplate owning the changelog follows from rulings 2 and 3, and it keeps the site's voice. The
latest-version-only decision costs one link and was the reviewer's own recommendation.

Rulings 10, 12, and 13 are genuine forks of taste, money, or Geoff's time, and each carries
enough evidence to rule. Ruling 11 is a genuine fork but needs the split in V3. One small fork
is buried: whether pack grantors are credited (V7). The spec can take the name-only answer
itself.

## Question 5: altitude and bloat

The spec stays at design altitude. "What each check proves" reads as acceptance criteria, and
the few implementation details it keeps (staging then swapping, writing the lock last) each
guard a named failure. Ranked by cost against protection:

1. Minor: spec:128-132 enumerates the current credited set (four Go modules, the web stack,
   three fonts). The credits leg enforces completeness, so the list protects nothing, and it goes
   stale on the next `go.mod` change. Replace it with the rule and the two source files.
2. Trivial: the command form `npm run sync-dubplate -- <ref>` (spec:168) and the
   unexported-constructor paragraph (spec:237-240) are plan detail, but short and harmless.

Nothing else is worth cutting.
