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
