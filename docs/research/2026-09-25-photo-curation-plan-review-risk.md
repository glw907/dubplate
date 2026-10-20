# Photo curation round one: plan review, domain risk lens

Reviewed: `docs/superpowers/plans/2026-09-25-photo-curation-round-one.md` at `1240bfa`, against the
spec (revision 2.2), `docs/STATUS.md`, `scripts/check.sh`, `.gitleaks.toml`, and the research under
`docs/research/2026-09-24-identity-photos/`. Reviewer: Opus 5.5, domain risk lens,
2026-09-25.

**Verdict:** approve after six Major folds. Nothing is a blocker. Every image reaches the repo only
after Geoff's keep and the Task 5 spot-check. Every letter reaches a recipient only after Geoff sends
it. The Majors are the places where those two human gates are weakest: Geoff cannot see a false
claim inside a letter he did not research, or a laundered license behind a clean-looking template.

Counts: 0 Blocker, 6 Major, 9 Minor, 3 Notes. One OWNER FORK (PD7).

State at review time: Task 0 has already run (`fae78c3`: `.gitignore`, `.gitleaks.toml`, the STATUS
lane row, ROADMAP, the identity spec pointer). The allowlist therefore lands before any candidate
exists, so the gitleaks-ordering risk the brief names is closed. `$DUBPLATE_HARNESS_CONTACT` is set
in an agent shell. The repo is private today.

## Major

### PD1. Letters can carry unverified claims and a malformed contact

- **Severity:** Major.
- **Where:** plan lines 158 to 170 (Task 3).
- **Defect:** Task 3's acceptance checks the letter's shape but never its facts. The research it
  drafts from marks its own weak spots:
  - Beth Lesser's CC BY-NC-ND 4.0 archive comes from a "search summary, no single canonical URL
    surfaced ... verify the exact hosting page" (`photographer-outreach.md:33`). The natural letter
    ("you already share your archive under CC BY-NC-ND") would state an unverified fact about the
    recipient.
  - Lesser's contact route is "verify a direct contact before drafting". No direct contact exists
    in the research.
  - Charlie Phillips's archive address is `curator@www.charliephillipsarchive.com`
    (`photographer-outreach.md:83`). The `www.` inside an email domain is almost certainly a
    scraping error, so a letter sent there bounces or goes astray.
  - Living status is recorded once, from 2026-09-24 searches. A letter to a photographer who has
    died, sent in Geoff's name, is the worst reputational outcome this pass can produce.

  Task 3 says "each with its URL", but nothing requires those URLs to resolve today.
- **Fold:** add three items to Task 3's acceptance:
  - Every factual claim a letter makes about the recipient cites a source the drafter fetched raw
    this session. A claim without one is cut.
  - Each contact route is fetched this session and quoted from the recipient's own page. Living
    status is re-checked, with the source recorded.
  - Every image URL in a letter returns 200 and still shows the image, with the date checked.

  The draft header records all three. A route that cannot be confirmed marks the draft `blocked`
  in `outreach/README.md`, not `drafted`.

### PD2. Letters do not have to say what CC BY lets third parties do, and they promise a pack that does not exist

- **Severity:** Major.
- **Where:** plan lines 162 to 166.
- **Defect:** the acceptance forbids "non-commercial" and requires "free, open-source, MIT". It
  never requires the letter to say plainly that CC BY 4.0 lets anyone reuse the image, including
  commercially, once it ships in an MIT project. A photographer who grants CC BY believing the
  grant covers dubplate only has given uninformed consent. That consent is the most likely source
  of a later dispute and a `withdrawn` record, which is the spec's costliest path.

  The letter also "offers the optional photo pack as the fallback". The spec puts the pack's
  packaging at rung 7 or later, and no distribution channel, audience, or NC mechanism is defined.
  The letter would promise a named third party something undesigned.
- **Fold:** the acceptance requires one plain sentence that CC BY 4.0 lets anyone reuse and adapt
  the image, commercially too, with credit. It describes the pack only as a planned, optional
  download kept apart from the free app, with no date. It makes no promise of prominence, a
  linkback, or a timeline beyond what the spec's CREDITS rule already delivers.

### PD3. The license check cannot see laundering or pending deletion

- **Severity:** Major.
- **Where:** plan lines 49 to 51 (review focus 1), lines 189 to 196 (Task 4 brief), and lines 201
  to 209 (Task 5).
- **Defect:** Task 4's brief carries the spec's own-work rule ("own work credits someone else").
  That rule catches only the crudest case. The common failures on Commons and Flickr pass it:
  - **Flickrwashing.** An account posts others' press or album photos under CC BY. `research-d`
    re-verifies 140 unverified Flickr rows, which is exactly this population.
  - **Implausible "own work".** An upload claims "own work" for a 1979 concert photo, made by an
    account whose other uploads show it cannot be the photographer.
  - **Pending deletion.** A Commons page carries `{{delete}}`, `{{copyvio}}`,
    `{{No permission since}}`, `{{No source since}}`, or an open deletion request.

  The Task 5 reader compares quotes against the source. A laundered license quotes perfectly, so
  the spot-check scores it a match.
- **Fold:** add to the Task 4 brief and Task 5's checklist, recorded in `verified`:
  - Every Commons wikitext is scanned for deletion, no-permission, and no-source templates, and
    any hit is dropped with reason `license`.
  - Every Flickr-origin candidate checks its account against Commons' "Questionable Flickr images"
    list and looks at the account's other uploads for mixed authorship.
  - Every "own work" claim on a work dated more than about 20 years before its upload names why
    the uploader is plausibly the author. Otherwise the candidate is dropped.

  This fold adds no new schema.

### PD4. The public-domain bases list has no content and no evidence standard

- **Severity:** Major.
- **Where:** plan line 106 (Task 1's "public-domain bases") and line 205.
- **Defect:** Task 1 asks a Sonnet implementer to produce the closed list of public-domain bases,
  but no task says what the list contains. The spec's US-PD tightening makes this list the
  load-bearing control. The high-risk bases are exactly the ones an implementer might add without
  a qualifier:
  - US publication without notice, from 1978 to 1989. It carries a five-year cure by registration.
  - US not-renewed. Photographs are rarely registered or renewed, so "no renewal found" is weak.
  - Foreign works whose US copyright URAA restored on 1996-01-01. Most European and Jamaican
    photographs from after 1929 fall here. Bundesarchiv and Anefo images are licensed, not PD.
  - Flickr Commons "No known copyright restrictions". It is a statement, not a basis.

  The DC hardcore flyers from 1980 to 1985 will mostly arrive claiming "no notice", which is the
  1978 to 1989 case.
- **Fold:** Task 1's acceptance names the list. At minimum: pre-1931 publication (a rolling
  95-year term), US federal government work, US publication before 1978 without notice, and US
  publication 1978 to 1989 without notice with no registration within five years. Each base names
  the fact it requires. "No known copyright restrictions" and PDM-only are explicitly not bases.
  The check fails a `public_domain` whose base is outside the list. The flyers agent's brief says
  a no-notice claim needs the notice-free published copy, the scan itself, as its evidence.

### PD5. Photo commits on shared master move lane 1's scoped gate base and its STATUS

- **Severity:** Major.
- **Where:** plan lines 28 to 33, line 199 (Task 4's "STATUS updated"), and lines 217 to 229
  (Task 6).
- **Defect:** on master, `check.sh --scope-changed` sets `change_base=HEAD~1`
  (`scripts/check.sh:166-167`). Suppose a photo commit lands between a lane 1 implementer's commit
  and that task's gate run. Lane 1's scoped gitleaks then scans the photo commit's files, and lane
  1's own commit goes unscanned. A round-close commit of more than 100 files forces a whole-tree
  scan, which costs time but is safe. The Task 1 commit, at about 15 files, is the silent case.

  A per-task `diff-reviewer` reading `HEAD~1` or a task range picks up the interleaved photo files
  the same way. Encoded pairs moved into `photos/` before their commit are also untracked, not
  ignored, so they join every lane 1 scoped change set while they wait.

  Task 4 and Task 6 edit `docs/STATUS.md`. So does lane 1's unattended conductor, at each of its
  segment boundaries. Task 0's clean-tree check covers Task 0 only.
- **Fold:**
  - The photo lane commits only at a lane 1 segment boundary, or while lane 1 is idle, confirmed
    from the lane registry. Encode-and-move plus commit in Task 6 runs as one short step with no
    gap.
  - Task 4's and Task 6's checkpoints write the round log and the plan's Later rounds section, not
    STATUS. The photo lane changes STATUS only through its one registry row, and only when
    `git status -- docs/STATUS.md` is clean.

  If the photo lane and lane 1 run under one conductor, say so in the plan header. The ordering
  then holds by construction.

### PD6. Raw Flickr HTML will blow the token ceiling

- **Severity:** Major.
- **Where:** plan lines 36 and 37, lines 177 to 182, and line 23 (the 3.5M ceiling).
- **Defect:** the plan mandates raw `curl` of Flickr page HTML. A Flickr photo page is several
  hundred KB of HTML and script. An agent that reads it into context to find the license link
  spends roughly 50K to 100K tokens per candidate. `research-d` alone has 140 rows. Together with
  Flickr-origin rows in the other research files, that is several million tokens, which breaks the
  ceiling in Task 4 by itself.

  With extraction instead, the ceiling is plausible. By my estimate: Task 4 about 1.2M to 2.0M
  across six agents, Task 1 and Task 2 about 0.5M, Task 3 about 0.2M, Task 5 about 0.3M, and the
  conductor about 0.3M, for about 2.5M to 3.3M total. That is tight but fits.
- **Fold:** the Global constraints and the Task 4 brief say to never read a raw page into context.
  Save it under the agent's candidate directory. Extract the license anchor, and the dimensions
  where present, with `grep` or `sed`. Quote the matched line into `verified`. For Commons, use
  `action=raw` for the license quote and the `imageinfo` API (`iiprop=size|sha1|url`) for
  dimensions, which rejects an under-1000px file before download. Each agent reports its own
  spend. Arm the runaway guard on all six agents: per the dispatch-mechanics memory, about 2MB for
  research readers.

## Minor

### PD7. Personality rights and identifiable minors are unaddressed (OWNER FORK)

- **Severity:** Minor. OWNER FORK.
- **Where:** plan lines 189 to 196, and lines 137 to 142 (the sorting page captions).
- **Defect:** CC licenses expressly do not license publicity or privacy rights. Commons marks
  these with `{{Personality rights}}`, and one research note already quotes "rights of privacy and
  publicity may apply" (`photo-candidates-2.md:26`). Crowd and dancer shots are a floor-grid
  subject. All-ages DC hardcore shows put identifiable teenagers in frame. Nothing records the
  concern or shows it to Geoff. The legal risk of decorative use in a private group server is low,
  but a future public website that browses the records raises it.
- **Recommendation:** no blanket exclusion. Agents quote any personality-rights or consent
  template into `verified` and put "identifiable non-performer, apparent minor" in the `screen`
  note. The sorting page caption shows the flag, and Geoff decides per image. Geoff's alternative
  is a hard rule that drops close shots of identifiable apparent minors who are not performers.

### PD8. Institutions often do not own the copyright they are asked to license

- **Severity:** Minor.
- **Where:** plan line 154 and lines 162 to 166.
- **Defect:** the DC Punk Archive's holdings are largely donated by photographers who kept
  copyright, the research notes. Saah and Connolly are examples. The British Culture Archive's
  People's Archive is contributed by the public. A letter asking the institution to "license CC BY"
  asks it for a right it may not hold.
- **Fold:** the two institution letters ask who holds rights in the named images and whether the
  institution will forward the request, then ask for CC BY from whoever holds it.

### PD9. Attended time is undercounted

- **Severity:** Minor.
- **Where:** plan lines 214 to 216 ("the one expected attended touchpoint").
- **Defect:** round one actually pulls Geoff in several times:
  - He reads, approves, and sends five letters.
  - Each Artifact asset upload call is approval-gated ("one approval covers the call"). About 200
    thumbnails in batches of 25 is about 8 prompts, and about 16 if full-size files go up too
    (PD12). In an unattended run, each prompt stalls the conductor.
  - The floor-waiver bundle, although it is round two or later.
- **Fold:** name the letters' read as a second touchpoint, batched into the same message as the
  "sitting ready" notice. In Task 2's conductor-first step, confirm the upload permission rule and
  the asset store's count and size limits from `artifact-capabilities`. If uploads prompt, run them
  attended at the start of the sitting message and do not scatter them. Sitting size is
  reasonable: 150 to 250 thumbnails at Y/N speed is a single short sitting, and resume-by-device is
  already specified.

### PD10. Wikimedia and Flickr bulk download etiquette is unspecified

- **Severity:** Minor.
- **Where:** plan lines 34 and 35, and line 174.
- **Defect:** six parallel agents on Geoff's residential IP will fetch originals from
  `upload.wikimedia.org` and Flickr. Wikimedia's robot policy and its 2025 tightening against bulk
  crawlers rate-limit, and can block, heavy downloaders. Flickr blocks scrapers.
- **Fold:** each agent downloads serially, honors 429 and `Retry-After`, and stops rather than
  retrying in a loop.

### PD11. ImageMagick decodes untrusted downloads

- **Severity:** Minor.
- **Where:** plan lines 42 and 43.
- **Defect:** zine scans may arrive as PDF, which hands the file to the Ghostscript delegate, and
  flyers may arrive as SVG. Both are historical ImageMagick attack paths.
- **Fold:** accept only JPEG, PNG, TIFF, and WebP, checked with `file` or `magick identify` before
  encoding. Pass an explicit coder prefix (`jpeg:<src>`). Anything else is dropped with reason
  `license` or a new `format` code.

### PD12. Geoff judges on 600px thumbnails only

- **Severity:** Minor.
- **Where:** plan lines 214 and 215 (Task 6 uploads thumbnails and the list).
- **Defect:** the spec says `Space` opens the full-size file, but Task 6 uploads only thumbnails.
- **Fold:** also upload the 1600px encode, or state that `Space` shows the thumbnail enlarged and
  that focus judgment rests on the agent's screen.

### PD13. MusicBrainz name-to-ID can mis-assign

- **Severity:** Minor.
- **Where:** plan lines 220 and 221.
- **Defect:** common band names resolve to several artists. A wrong MBID breaks the one-per-artist
  dedupe silently.
- **Fold:** record an MBID only on a single top-score match whose area or active years fit the
  record. Otherwise keep the normalized name.

### PD14. Flyers embed third-party content

- **Severity:** Minor.
- **Where:** plan lines 186 and 187.
- **Defect:** a flyer artist's grant does not cover a band photo or logo pasted into the flyer.
- **Fold:** the flyers agent notes an embedded photograph in `screen` and treats a flyer built
  around another photographer's image as needing that photographer's license too, or drops it.

### PD15. Original downloads can be huge, and thumbnail renditions are not hash-stable

- **Severity:** Minor.
- **Where:** plan line 194.
- **Defect:** archive originals on Commons (Nationaal Archief, Bundesarchiv, LoC TIFFs) can be 50
  to 200MB each. A Commons thumbnail URL, if used instead, can re-render with different bytes. That
  breaks the `source_sha256` re-fetch check that the spec's packaging and round close rely on.
- **Fold:** download the original when it is under about 40MB. Above that, take a named rendition
  of at least 1600px and record in the record that `source_sha256` covers that rendition.

## Notes

- **N1.** Task 0 ran at `fae78c3` before this review. Its acceptance items appear satisfied (the
  ignore entry, the allowlist with a comment, the STATUS row, and the carried line at STATUS:54).
  The plan should mark Task 0 done so no executor re-runs it.
- **N2.** Withdrawn images stay in git history (spec, ratified). That holds while the repo is
  private. Before the repo goes public, which the stranger-setup goal implies, a DMCA notice to
  GitHub against a history blob would require a history rewrite on master. Carry this to the
  release checklist. It is not a round-one change.
- **N3.** The tree carries warm changes this lane did not author (`web/package.json`,
  `web/package-lock.json`, the group-scale spec and fold). The plan's `git commit -- <paths>` form
  protects them. Task 1's implementer must not stage anything outside its Files list.
