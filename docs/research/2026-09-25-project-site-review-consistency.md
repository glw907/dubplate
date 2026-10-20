# Project site spec review: consistency lens

Target: `docs/superpowers/specs/2026-09-25-project-site-design.md` at `e91f528` (revision 1).
Lens: does the draft contradict a ratified document, and does each citation say what the draft
claims? Checked against CLAUDE.md, ROADMAP.md, the visual identity spec, the group-scale
principles spec, the photo curation spec, the streamer spec (`2026-09-07-streamer-2-0-design.md`,
revision 2.4), the greenfield build spec, the working tree, the cairn-cms repo, and the project
memory files.

Verdict: the draft is broadly sound. The versions (0.x launch, 0.8 streamer), the group and member
vocabulary, the reject-contract framing, the forward-hooks rule, the MIT license, and the cairn
citations (`docs/admin/README.md`, `create-cairn-site`, the docs reset) all check out. Three
findings are major: two sequencing preconditions that ratified documents set and the draft
omits, and one misattributed initiative. The rest are minor inventory and citation fixes.

## Findings, ranked by consequence

### K1 (major): Phase B omits the give-back license precondition

Location: `2026-09-25-project-site-design.md:132-134`.

Phase B's preconditions are "0.8 has shipped, that a stranger can reach the source or the
binaries, that the guides describe the tagged 0.8 release, and that the domain is ruled and
registered." ROADMAP's `give-back` entry (`ROADMAP.md:158-166`) sets a precondition on exactly
this event: the Unidecode path-1 pass (the binary stops embedding the GPL-derived table) is "a
small pass, due before the first release strangers can install (with `packaging`), plus a
lawyer's confirmation of the distribution reading before that release." A public site whose
"Set up a server" page tells a stranger how to install is that release's front door.

Fold: add a fourth Phase B precondition: the give-back path-1 pass has landed and the lawyer's
confirmation of the distribution reading is recorded, citing `ROADMAP.md` `give-back`.

### K2 (major): `oss-pair` is miscited, and distribution sequencing conflicts with "publishes with 0.8"

Location: `2026-09-25-project-site-design.md:141-143`.

The draft says "The dubplate repo is private today. Open-sourcing is the `oss-pair` entry's
question and distribution is `packaging`'s." ROADMAP's `oss-pair` (`ROADMAP.md:220-227`, Someday)
is about publishing "the uploader and the pings Worker as documentation-linked siblings" plus a
validated-intake write-up. It isn't the question of making the dubplate repo public. No ratified
document owns that question today; the principles spec only says "The repository keeps its MIT
license" (principles spec, "What these principles can and cannot do"), and `LICENSE` is MIT while
`gh` reports the repo PRIVATE.

Sequencing also conflicts. ROADMAP says `packaging` "Runs after `dubplate-organizer`"
(`ROADMAP.md:190-191`) and the header orders it after the TUI and local binary (`ROADMAP.md:8-9`).
The greenfield build spec puts "the local `dubplate` binary, the TUI in both modes, the two-binary
split" after launch (`2026-09-02-greenfield-server-build.md:1001-1004`). If Phase B waits on
`packaging` for distribution, owner ruling 5 ("It publishes with 0.8") can't hold. Note that 0.8
itself already implies a public release channel: the streamer spec's `dubplate upgrade`
"downloads a release, verifies its checksum against the signed release manifest"
(`2026-09-07-streamer-2-0-design.md:366-367`).

Fold: correct the citation, and route the question explicitly.

**OWNER FORK: where repo visibility and 0.8 distribution are decided.**
- (a) 0.8 ships its own minimal public channel: signed release binaries that `dubplate upgrade`
  already needs, plus the public source. `packaging` keeps the richer installers (cloud-init
  generation, desktop heads). Phase B stays at 0.8.
- (b) Phase B waits for `packaging`, which waits on the organizer. The site publishes later than
  0.8, and ruling 5 is amended.
- (c) A new ROADMAP entry owns "open the dubplate repo" separately from `oss-pair`, with (a) or
  (b) for binaries.

Recommendation: (a), recorded as a new line in the `packaging` or `streamer-0-8` entry. 0.8's
upgrade path already needs a public signed release, so it is the smallest change that keeps
ruling 5.

### K3 (major): the configuration reference depends on forward hook 18, which rung 14 owns

Location: `2026-09-25-project-site-design.md:59`, `:79-83`, `:126-127`.

The configuration reference "builds from the config key registry" and Phase A builds "the three
generators." A reference page needs a description per key. The registry has none today
(`internal/config/registry.go` carries no description or sensitive field). The streamer spec
assigns that field to hook 18: "A description and a sensitive mark on every registered config
key, rung 14" (`2026-09-07-streamer-2-0-design.md:518-520`). The CLAUDE.md refinement says "a 0.8
forward hook is named in the rung plan and built by the rung that first consumes it, and nothing
is exported ahead of a caller." Phase A's timing is tied to cairn's docs reset, not to the rung
ladder, so the generator may be written before rung 14. It would then either land hook 18 outside
its rung or render a page with no descriptions. The draft's general "weighs each generator
against the 0.8 spec's forward-hooks list" (`:82-83`) doesn't surface this concrete dependency.

Fold: name hook 18 in the Pages section as the configuration generator's input. State that the
generator lands after rung 14 or, if Phase A runs first, that the generator becomes hook 18's
first consumer and the Phase A plan coordinates with the rung 14 lane. Do not leave it to the
plan to discover.

### K4 (minor): `internal/reject` doesn't exist yet in the greenfield tree

Location: `2026-09-25-project-site-design.md:79-81`.

"The reject page builds from `internal/reject`, whose wording is already contract." The contract
is ratified (CLAUDE.md, "Reject wording is contract"), but the package lives only under
`_salvage/internal/reject/`. The greenfield tree lists it as planned
(`2026-09-02-greenfield-server-build.md:434`) and rung 7 rebuilds it. Salvage is never imported.

Fold: say the page builds from the greenfield `internal/reject` that rung 7 lands, so the
generator is sequenced after rung 7.

### K5 (minor): the theme file inventory misnames two files

Location: `2026-09-25-project-site-design.md:90-91`, `:108-110`.

- "The app's component layer (`surface.css`, `system.css`) probably doesn't [travel]." Both are
  token files. `system.css` is the `dubplate-system` theme, "the third stamped value," repeating
  the composed themes' declarations plus the ink and surface values. `surface.css` is "the
  surface tokens: a fourth layer" carrying the printed-shadow and keyline values the identity
  spec pins (`web/src/lib/themes/surface.css:1-9`). Leaving them out drops the follow-the-OS
  theme and the hard printed shadow, which contradicts ruling 2's "the theme tokens can't drift."
- "The theme tokens live in `web/src/lib/themes/` ... with the font files they reference." The
  woff2 files live in `web/static/fonts/`, beside `FONTS.md` and `licenses/*-OFL.txt`. The OFL
  requires its license text to travel with redistributed fonts.

Fold: list the traveling set as `dubplate.css`, `label.css`, `label-night.css`, `inks.css`,
`surface.css`, `system.css`, `fonts.css`, and `web/static/fonts/` with its `licenses/` and
`FONTS.md`. Name the component layer by its actual home, or drop the sentence.

### K6 (minor): command naming and the citation for `init` and `upgrade`

Location: `2026-09-25-project-site-design.md:55-56`, `:60`, `:81-82`, `:85-86`.

The draft cites ROADMAP `packaging` for `dubplate init` and `dubplate upgrade`. ROADMAP records
the move, but the definition lives in the streamer spec, "Hosting and setup"
(`2026-09-07-streamer-2-0-design.md:360-370`), and slice 8 (`:445-446`). The names also don't match
the built binary. The server's cobra root is `dubplate-server` (`cmd/dubplate-server/main.go:131`),
and today's verbs are spelled `dubplate-server member` (principles spec, `:96`). The bare
`dubplate` name is reserved for the future local binary (streamer spec `:175`; build spec
`:1002`). The streamer spec itself uses `dubplate init` loosely, so this is an upstream
inconsistency the site will inherit, not one it created. The "Command reference" generated "from
the cobra command tree" (`:81-82`) must name which binary's tree, since 0.9 adds a second one.

Fold: cite the streamer spec for both commands. Add one line saying the guide uses whatever name
the 0.8 server binary ships under (today `dubplate-server`), and that the command reference
covers the server binary until the local binary exists. Optionally file a note against the
streamer spec to settle `dubplate init` against `dubplate-server init`.

### K7 (minor): the is/isn't page omits the legal note the principles spec routes to the website

Location: `2026-09-25-project-site-design.md:53`, `:64-68`.

The principles spec's landing table has the row "'dubplate is, dubplate isn't' and the legal note
| README; website | new." Its legal note is its own section ("A legal note": "Nothing in this
spec, the README, or the website is legal advice"). The draft's page carries only "The 'dubplate
is' and 'dubplate isn't' lines," and its move into `docs/guide/` names only those lines.

Fold: the "What dubplate is, and isn't" page carries the legal note too, and the move into
`docs/guide/` takes both.

### K8 (minor): photos cite the superseded ruling and miss the curation spec's website hooks

Location: `2026-09-25-project-site-design.md:147-149`, `:62`.

The draft points to "the redistribution-permitting licenses the visual identity spec already
requires." The photo curation spec is the later ruling: "Where this spec differs from it, this
spec is the later ruling" (`2026-09-25-photo-curation-design.md:18-20`). It also anticipates this
site: "the future website can browse the same records" (`:11-12`) and each slug "later serves as
the website's path for the photo" (`:130-131`). The draft vendors only tokens and guides, so
`photos/` and the CREDITS manifest have no sync path, and the About page's "photo credits" has no
stated source. Records with `status: pack` (NC or dubplate-only permission, `:178`) must never
reach the site.

Fold: cite the photo curation spec. Either add `photos/` (picked records only) and the generated
CREDITS to the vendored set under the same lock, or state that photos are out of scope for the
first publication. Name the exclusion of `pack.json` records.

### K9 (minor): Phase A declares an initiative with no ROADMAP entry

Location: `2026-09-25-project-site-design.md:128`.

"Phase A is its own initiative, outside the rung ladder." CLAUDE.md's ledger rule makes
`ROADMAP.md` the initiative queue, managed by `/log-project`, and ROADMAP has no project-site
entry at `e91f528`.

Fold: add a line that the spec's approval opens a `project-site` entry in ROADMAP (Planned),
triggered by the close of cairn's docs reset.

## Checked and consistent

- Versions: "publishes with 0.8 and describes 0.8," "doesn't document the 0.x pre-release,"
  and the Navidrome credit "before 0.8" match the version ruling and memory.
- Vocabulary: "group," "member," and "invited member" throughout; no "family" outside examples.
  No coined metaphors or record-shop puns.
- The principles spec's statement that "The website publishes it when 0.8 ships" matches the
  draft's claim, and the share line's week and month tie to ruling 8 as the draft says.
- ROADMAP `packaging` does record `dubplate init` and `dubplate upgrade` moving into 0.8
  (`ROADMAP.md:191-193`); K6 concerns the better citation and the binary name only.
- The comparison page's no-disparagement rule and the Navidrome credit don't conflict with the
  own-visual-identity ruling, which concerns the look, not the credit.
- The cairn trial's citations exist: `cairn-cms/docs/admin/README.md` and the
  `create-cairn-site` package; cairn's STATUS shows the docs reset in progress (pass 2a next).
- Lane rules: `docs/guide/` as a document lane and the generators as a code lane match CLAUDE.md.
