# Project site spec: fold record

Target: `docs/superpowers/specs/2026-09-25-project-site-design.md`, revision 1 (`e91f528`) to
revision 2. Reviews folded: contract (C1 to C8), drift (D1 to D10), consistency (K1 to K9),
mechanics (M1 to M10), and prior art (P1 to P10), all under
`docs/superpowers/research/2026-09-25-project-site-review-*.md`. The fold also takes Geoff's two
additions of 2026-09-25 as owner rulings 8 and 9 (the Credits page).

Totals: 47 findings. 44 folded, 4 of them with a part refused (C2, D4, D6, P4). 0 refused whole. 3 became owner
forks (K2, P8, and P1, whose page is folded and whose channel question is the fork). The spec now carries four open rulings, numbered 10 to 13.

## Facts verified before folding

- `web/src/lib/themes/` holds seven files, and the fonts sit in `web/static/fonts/` with
  `FONTS.md` and three OFL texts in `licenses/`.
- `internal/reject` exists only under `_salvage/internal/reject/`. The greenfield build spec
  lists it as planned and assigns the reject copy contract to rung 7.
- The server binary's cobra root is `Use: "dubplate-server"` (`cmd/dubplate-server/main.go:131`).
- `git tag -l` lists only `pre-greenfield`.
- `npm view create-cairn-site` returns a 404 on 2026-09-25.
- ROADMAP `give-back` makes path 1 and a lawyer's confirmation "due before the first release
  strangers can install".
- ROADMAP `oss-pair` covers the uploader and the pings Worker, not the dubplate repo.
- Streamer spec hook 18 adds "a description and a sensitive mark on every registered config key,
  rung 14". Its `dubplate upgrade` "downloads a release, verifies its checksum against the signed
  release manifest".
- `daisyui/theme/index.js` gates the `:root` selector on `isDefault` and the dark rule on
  `prefersdark`.
- The photo curation spec's picked records carry `photographer`, `credit`, `source_title`,
  `license`, and `source_url`. It assigns "the CREDITS generator" to rung 7 and says `pack`
  records ship nothing. Letter grants land as records whose `source_url` points at the grant.
- `go.mod` directly requires cobra, go-mail, `golang.org/x/time`, and modernc SQLite.
- The greenfield build spec says the initiative "succeeds or fails on adopting beets'
  battle-tested matching behavior".
- The principles spec's "isn't" list carries "A public music site", and its "A legal note"
  section and landing table route the legal note to the website.

## The shared root

C1, C3, C5, C6, C7, D3, D5, D7, M4, and M7 describe one defect: revision 1 named a lock and a
gate without saying what the sync reads, what the lock proves, and what dubplate's gate proves.
Revision 2's "What each check proves" section fixes it at the root. The sync reads git objects
listed in a dubplate-owned export manifest. The site's lock check proves the vendored tree is
exactly what a sync wrote, and it names its limit (no provenance, covered by a workstation
`--verify`). dubplate's gate proves the exported files are current and true to the code.
C2, D1, and D2 share a second root, freshness against releases and against the code, which the
same section folds.

## Dispositions

| IDs | Disposition | Where |
|---|---|---|
| C1, D7 | Folded. The lock check names every failure (missing lock, empty class, missing file, unlisted file, hash mismatch) and its provenance limit; the sync stages and writes the lock last. C1's "every class non-empty" became "every class the manifest declares", since generated sources arrive at rungs 7 and 14. | What each check proves |
| C2, D2 | Folded. The guide leg checks every backticked command, flag, and key against the command tree and the registry. The reviewer's owner fork is not a fork: ruling 2 already promises the docs can't drift, and option (b) is the cheapest design that keeps that promise. Option (c), executable guide steps, is refused: it overlaps `packaging` and costs far more than the renamed-name drift it adds cover for. | What each check proves |
| C3, D3, M4 | Folded. Generated pages are committed in dubplate with a regenerate-and-diff leg; the sync reads git objects only and refuses a lightweight or moved tag. | What each check proves; Generated pages |
| C4 | Folded as the Phase A done list and the Phase B precondition list. The is/isn't move's signal is the principles spec's ratification. | Sequence |
| C5, D5 | Folded. The lock check runs in the deploy build; vendor directories sit outside every cairn content directory; a missing build hook is a cairn trial finding. | What each check proves |
| C6 | Folded. dubplate owns the export manifest, and its gate checks every path exists. | Keeping the site in step |
| C7, K5 | Folded. All seven theme files travel, with the fonts, `FONTS.md`, and the OFL texts, served at `/fonts/`. K5's correction stands: `surface.css` and `system.css` are token files, and revision 1's "component layer" sentence is dropped. | Keeping the site in step |
| C8, D9 | Folded. Comparison dates must be under six months old at Phase B, and the site's gate checks them before each later site release. | Sequence |
| D1, M7 | Folded. Every vendored page shows the ref from the lock; the sync accepts any commit in Phase A and records whether it is a release; Phase B refuses a non-release lock; the pin bump belongs in the release tool, with a checklist step until it exists. | What each check proves; Sequence |
| D4 | Folded in part. The font location and travel list are corrected (with C7), and Phase A acceptance adds a WCAG AA contrast check over the site's built CSS in both themes. Refused: locking the DaisyUI and Tailwind versions to dubplate's, because the site builds on cairn's chassis and its own contrast check covers the rendered result directly. | Keeping the site in step; Sequence |
| D6 | Folded in part. dubplate's guide leg fails on the box address, group hostnames, email addresses outside an allow-list, and home paths in any exported file. Refused: denying `docs/superpowers` and `_salvage` path citations, which leak nothing sensitive and would fail the token files' existing comments for no risk removed. | What each check proves |
| D8, P7 | Folded as the spec's own call, not a fork: dubplate owns a changelog with upgrade notes, vendored through the manifest, and each site post links to it. Rulings 2 and 3 already make dubplate the owner of product facts. Home's one paragraph stays site-owned and links to the is/isn't page for the fit facts, in place of rendering the vendored lines. | Pages |
| D10 | Folded as the spec's call, not a fork: latest version only, with the ref banner; the changelog's upgrade notes cover older versions; earlier guides are linked once the source is public (ruling 11). | What each check proves |
| K1 | Folded as a Phase B precondition citing ROADMAP `give-back`. | Sequence |
| K2 | Citation corrected (`oss-pair` doesn't own the question). Owner fork: open ruling 11, with the distribution options and a recommendation. | Open rulings |
| K3, M3 | Folded. The configuration generator is a consumer of hook 18 and lands after rung 14. M3's scope fork is taken, not asked: the reference covers every key the registry declares, which is what hook 18 describes, and the other homes are Run your server's prose. | Generated pages |
| K4, M2 | Folded. The reject page builds from rung 7's `internal/reject`; the generator lands after rung 7 and needs an enumerable reason set, routed to rung 7's plan (see owed items). | Generated pages |
| K6, M10 | Folded. The command reference walks the server binary's tree (`dubplate-server` today), excludes `cmd/harness`, and the plan weighs `cobra/doc`. The generators live beside the unexported constructors so nothing is exported ahead of a caller. The `init` and `upgrade` citations move to the streamer spec. | Generated pages |
| K7 | Folded. The is/isn't page carries the legal note, and the move takes both. | Pages |
| K8 | Folded into the Credits page: the photo curation spec is cited as the later ruling, photo credits are rung 7's CREDITS output vendored through the manifest, and `pack` records never reach the site. | Credits; Out of scope |
| K9 | Folded. Approval opens a `project-site` entry in ROADMAP (owed action below). | Sequence |
| M1 | Folded. The token files move unedited, and the site owns one reference-only adapter for the default and dark flags, the toggle names, the font roles, and the ink names. `dubplate.css` joins the travel list. | Keeping the site in step |
| M5 | Folded. The trigger adds "`create-cairn-site` is published on npm". | Sequence |
| M6 | Folded. The trial follows the operator's and designer's entry points as the reset publishes them, and code steps aren't logged as friction by default. | The cairn trial |
| M8 | Folded as a Phase A acceptance item: `X-Robots-Tag: noindex` on `workers.dev` hosts only, never a `robots.txt` bar. | Sequence |
| M9 | Folded as Phase A acceptance: guide links resolve to site routes, and the sitemap lists every route. | Sequence |
| P1, P9 | Folded: a site-owned Get help page with a troubleshooting list and a security contact from day one; no separate FAQ (P9's own recommendation). The channel choice is an owner fork: open ruling 12. | Pages; Open rulings |
| P2 | Folded. For members carries a short tested app list with dates and the app-password sign-in step. | Pages |
| P3 | Folded. A dubplate-owned Bring your library guide, and a "Leaving dubplate" section in Run your server. | Pages |
| P4 | Folded in part. Run your server carries a security and privacy section with the outbound-call list. Refused: generating the outbound-call list, which is a short list a guide review holds true at far less cost than a generator. | Pages |
| P5 | Folded. Set up a server ends with configuring and testing a backup. | Pages |
| P6 | Folded. "What you need" is the guide's first, linkable section, linked from Home and the is/isn't page, with storage sized in albums. | Pages |
| P8 | Owner fork: open ruling 13. | Open rulings |
| P10 | Folded. Home names its three screenshots, and Out of scope says why there is no demo server. | Pages; Out of scope |

## Owner rulings 8 and 9: the Credits page

Geoff's two additions become the Credits page. Its software half is a dubplate-owned guide,
`docs/guide/credits.md`. beets and Navidrome are credits of judgment, not dependencies, so no
generator can write their lines. The drift risk is a new dependency shipping uncredited, so
dubplate's gate fails when a direct `go.mod` requirement or a `FONTS.md` face goes unnamed. The
set is drawn from the streamer spec's prior-art table and `go.mod`, never invented. The photo half
is generated: rung 7's CREDITS generator reads the curation records, and the site vendors its
output. About's "photo credits" line moved to Credits. Navidrome's credit lives only on Credits;
the comparison page's Navidrome entry keeps one sentence of lineage and links there.

## Owed errata and routing

- **Principles spec (erratum).** Revision 2 moves the legal note to `docs/guide/` alongside the
  is/isn't lines, which revision 1 already moved. Once the principles spec is ratified and the
  move lands, its "dubplate is, and dubplate isn't" and "A legal note" sections become pointers to
  `docs/guide/`, and its landing table row names the guide as the source. That changes a spec
  that is meant to be ratified, so it is recorded here and not edited.
- **Streamer spec (note).** It writes `dubplate init` and `dubplate upgrade`, while the server
  binary is `dubplate-server` and the bare `dubplate` name is reserved for the local binary. The
  command names should be settled there.
- **Streamer spec hook 18 (routing).** The configuration reference generator is a new consumer of
  hook 18, beside the settings surface and `init`.
- **Rung 7 plan (routing).** The reject generator needs the reason set enumerable, with a
  canonical wording for each parameterized reason. The Credits page needs rung 7's CREDITS output
  at a stable path the export manifest can list.
- **ROADMAP (action at approval).** Open a `project-site` entry under Planned, triggered by the
  close of cairn's docs reset and the publication of `create-cairn-site`.

## Second fold

Target: revision 2 (`73b429f`) to revision 2.1. Input: the fold verification,
`docs/superpowers/research/2026-09-25-project-site-fold-verification.md` (V1 to V11). Each
finding was checked against the spec and its sources before folding. Totals: 11 findings, 11
folded, 0 refused, and no new owner fork.

| Finding | Disposition | Where |
|---|---|---|
| V1 | Folded. The guide leg reads a dubplate-owned planned-names list, kept beside the export manifest and not exported, with the rung for each entry. A planned name passes while absent, and a stale entry fails. The list is empty at the 0.8 tag, which joins the Phase B preconditions. One clause says how the leg tells a command, flag, or key span from other code spans (binary-name prefix, `--name` shape, and the `DUBPLATE_` prefix or the store-homed key shape quoted from `internal/config` by the Phase A plan). | What each check proves; Phase B |
| V2 | Folded. The ROADMAP `project-site` entry carries two owed dubplate lanes, triggered by the rung 7 and rung 14 merges, each adding its generator's regenerate-and-diff coverage. Phase A builds either generator itself if it runs after that rung. | Sequence |
| V3 | Folded. Open ruling 11 splits into (a) public signed binaries only, source private; (b) binaries and public source; and (c) wait for `packaging`, amending ruling 5. It states that opening the source is irreversible and names (b)'s costs: the `oss-pair` maintainership precondition and a pre-publication scrub (`CLAUDE.md` names the box's address). Recommendation (a), reasoned from ruling 5, `oss-pair`, `packaging`'s sequence, and `give-back`, with the principles spec's MIT line named as its cost. Not ruled. Ruling 12 and the earlier-version links now say what happens under each option. | Open rulings 11 and 12; What each check proves |
| V4 | Folded. The lock records the classes the manifest declares at the locked commit, so the empty-class check has an input. | What each check proves |
| V5 | Folded. A release tag is an annotated tag named `v<major>.<minor>.<patch>` with no pre-release suffix, and the release checklist cuts tags to that shape. | What each check proves |
| V6 | Folded. The picked `.webp` files are a fourth exported kind, joining the manifest when rung 7 lands the CREDITS generator, from the same commit as its output. | Keeping the site in step |
| V7 | Folded, decided in the spec. A letter grant under a shipped class credits as a picked record. A pack grantor gets a name-only line generated from its `pack.json` record's `photographer` and `credit` fields, with no image, once the pack ships. Reason: ruling 9 covers anyone who lets dubplate use licensed photos, and the curation spec keeps a pack candidate as "a full record without an image" (`2026-09-25-photo-curation-design.md:213-214`), so the credit comes from the same records with no second list. The need routes to rung 7's plan. | Credits |
| V8 | Folded. The credits leg also checks `web/package.json`'s runtime dependencies, which cost nothing until rung 7 adds `tus-js-client`. The claim narrows to "a new Go module, bundled web package, or font". Unidecode, fetched at setup under give-back path 1, is named by rule instead of gated. | Credits |
| V9 | Folded. The leak leg fails on IPv4 literals outside the RFC 5737 documentation ranges and on hostnames and email addresses outside an allow-list of example and upstream domains, so it names no site value itself. The site repo is stated as private, with opening it following ruling 11. The verifier's "read literal deny entries from secrets" alternative is not taken, since the pattern forms cover the named leaks. | What each check proves; Where the site lives |
| V10 | Folded. The is/isn't and legal-note move is a Phase B precondition, and the README links to the guide and holds no copy. | Pages; Phase B |
| V11 | Folded. The Phase A done list gains the changelog with its entry format, the release checklist, and the proven comparison-date check. | Phase A |

Also folded at the conductor's direction:

- **Credited set.** The hand list at revision 2 lines 128-132 is trimmed to beets and Navidrome by
  name, plus a pointer to the classes the guide covers: the prior-art table's adopted rows, the
  data services the server queries, the source of derived data, and the gate-checked set. A hand
  list goes stale on the next `go.mod` change.
- **MusicBrainz.** Verified in use: the client lives in `internal/engine/musicbrainz` (and
  `internal/harness/mbrequest`). It isn't a prior-art row, so the spec now credits it as a data
  service the server queries, citing that package, instead of implying the table supplies it.
- **Domain prices.** Ruling 10 cites its source: Cloudflare Registrar's domain-check API
  (`POST /accounts/{account_id}/registrar/domain-check`), run by the conductor on 2026-09-25,
  which returned `dubplate.studio` registrable at standard tier for $31.20 registration and
  renewal, and `dubplate.dev` for $12.20.

Owed routing added by this fold: the rung 7 plan also receives the pack name-only credit need and
the picked-image export class; the ROADMAP `project-site` entry, at approval, carries the two
merge-triggered generator lanes.

## Prose fold

Target: revision 2.1 (`44f94c9`) to revision 2.2. Input: the prose and facts review,
`docs/superpowers/research/2026-09-25-project-site-prose-review.md` (B1, W1 to W8, S1 to S11).
Each finding was checked against the spec and its sources before folding. Totals: 20 findings,
20 folded, 0 refused, and no new owner fork. No open ruling's options or recommendation changed.

Facts verified for B1, on 2026-09-25 against `53f9dc3`:

- `web/package.json` has no `dependencies` key. All twelve packages, Svelte, SvelteKit, daisyUI,
  and Tailwind among them, are `devDependencies`.
- `go.mod` has four direct requirements and thirteen indirect ones.
- `go version -m` on a Linux build of `./cmd/dubplate-server` lists 14 `dep` modules, indirect
  ones included (`github.com/spf13/pflag`, `modernc.org/libc`, `golang.org/x/crypto`,
  `github.com/google/uuid`, `github.com/dustin/go-humanize`).
- `go list -deps -f '{{with .Module}}{{.Path}}{{end}}' ./cmd/dubplate-server` lists the same 14
  without a build. With `GOOS=windows` it adds `github.com/inconshreveable/mousetrap`, so the
  linked set is per target.
- `go list -m all` lists 38 modules besides the main one, 24 of which the Linux binary never links
  (`golang.org/x/tools`, `modernc.org/ccgo/v4`, and others). It over-reports, so it isn't the input.
- The bundler is rolldown (Vite 8). Its `PluginContext` declares `getModuleIds():
  IterableIterator<string>`, "Get all module ids in the current module graph"
  (`web/node_modules/rolldown/dist/shared/define-config-C7HCoqY8.d.mts`). Tailwind and daisyUI
  enter through `@import 'tailwindcss'` and `@plugin "daisyui"` in `web/src/app.css`, not through
  a JS import, so the generator also reads those directives.

| Finding | Disposition | Where |
|---|---|---|
| B1 | Folded by widening. The credits guide names the notable projects and the font faces by hand, and the credits leg checks only the faces. A new generated third-party notices list covers every module the binary links, per release target from `go list -deps`, and every package the web bundle ships, from rolldown's `getModuleIds()` plus the stylesheet's `@import` and `@plugin` packages, each with its license text. The regenerate leg keeps it current. The "can't ship uncredited" claim and the `tus-js-client` sentence are gone. Reason for widening over narrowing: MIT and BSD notices must travel with a shipped binary, so the list is owed at 0.8 anyway, and the regenerate leg already exists. | Credits; What each check proves |
| W1 | Folded. dubplate's gate names four legs, one paragraph each: export (paths exist, plus the leak check over every exported file), regenerate, guide, and credits. | What each check proves |
| W2 | Folded. The kinds list is six items and names the fonts with `FONTS.md` and the OFL texts, and the changelog at `CHANGELOG.md` in the dubplate repo root (none exists yet; the path is the spec's call). | Keeping the site in step |
| W3 | Folded. The site repo starts private and opens only if dubplate's source opens (ruling 11, option b), stated as a condition, not a ruling. | Where the site lives |
| W4 | Folded. Verified against ROADMAP's header: `packaging` follows the TUI and the local binary, which follow the server build's launch. Ruling 11's context sentence is corrected; its options and recommendation are unchanged. | Open ruling 11 |
| W5 | Folded. The repeated binary-name paragraph is deleted, and its `init` and `upgrade` sentence joins the command reference paragraph. The guide leg reads a command span by the root command's name or a binary name on the planned-names list, which carries the 0.8 name. Ruling 11 says "0.8's `upgrade` command". | Generated pages; What each check proves; Open ruling 11 |
| W6 | Folded. The photo-credits paragraph splits into the picked-photo paragraph and the pack paragraph, and "no second list" appears once. | Credits |
| W7 | Folded. Page titles in running prose are quoted and carry a noun ("The "Set up a server" guide"). | Pages |
| W8 | Folded. The placeholder is site-side. An owed generator lane declares the page's class in the export manifest, and the next sync replaces the site's placeholder. | Phase A; Sequence |
| S1 | Folded. Verified in `web/src/lib/themes/inks.css` (no `--ink-success`) and the identity spec's ink table: the adapter maps success onto `--ink-accent`. | Keeping the site in step |
| S2 | Folded in "What each check proves", not in ruling 3's text, since the owner rulings keep Geoff's wording. | What each check proves |
| S3 | Folded. A 0.x tag of the release shape is a release tag; Phase B requires the 0.8 tag. | What each check proves |
| S4 | Folded. "no exported enumerator" (verified: `_salvage/internal/reject/reject.go` has the unexported `codeToReason` map). | Generated pages |
| S5 | Folded. The source is cited as the streamer spec's "License" section, which says implementers "never copy or transliterate GPL code" of Navidrome and gonic. It's the ratified record, so it's cited in place of `CLAUDE.md`. | Credits |
| S6 | Folded. | What each check proves |
| S7 | Folded. | The cairn trial |
| S8 | Folded. | Open ruling 10 |
| S9 | Folded as ordering only. Rulings 12 and 13 now list options before the recommendation, as ruling 11 does. No wording changed. | Open rulings 12 and 13 |
| S10 | Folded. The sentence splits, and the second says the photo curation spec governs where the two specs differ. | Out of scope |
| S11 | Folded. The "exported from the same commit" clause is cut, since the sync exports every path from one commit. | Keeping the site in step |

Also folded at the conductor's direction: the group-scale principles spec is ratified (its header
reads "Revision 3, 2026-09-25. Status: ratified (Geoff, 2026-09-25)"). The is/isn't and
legal-note move no longer waits on ratification, so it leaves the Phase B preconditions and joins
the Phase A done list, landing in Phase A's dubplate lane with the other guides. The principles
spec itself is not edited by this fold.

## Owner ruling on C2

Geoff ruled option (b) for C2 on 2026-09-25. The guide leg already carried its first half. The
second half, a site-side check that the lock's pin is the newest release, was missing: the spec
refused only a non-release ref. Revision 2.3 adds it in "What each check proves". The site's gate
fails when the newest release-notes post names a version newer than the lock's tag, since the
site's build can't read the private repo's tags.
