# Photo curation design: consistency review

Target: `docs/superpowers/specs/2026-09-25-photo-curation-design.md` at `67d4adf` (cited below as PC).
Lens: consistency with ratified documents, and accuracy of the spec's own citations and of its
"Changes to the identity spec" list. Findings are ranked by consequence. Polish is omitted.

Read against: `CLAUDE.md`; the visual identity spec (VI, `2026-09-24-visual-identity-design.md`,
rev 1.2); `ROADMAP.md` `photo-curation`; the web surface standard (W); `nc-licensing.md` and
`photographer-outreach.md`; the 5c-iii plan (`2026-09-25-rung-5c-iii-pages-fonts-rendered-gates.md`);
`docs/STATUS.md`; `LICENSE`; the memory files named in the brief; the SPDX license list
(`spdx/license-list-data`, fetched 2026-09-25).

Counts: 0 blocker, 6 major, 9 minor.

## Verified accurate

These claims check out and need no fold:

- Beckman is rank 10 in `photographer-outreach.md:21`, and her 1980s NYC hip-hop work is on
  record (`:119`). Lesser, the DC Punk Archive, and Phillips are ranks 1 to 3 (`:12-14`).
- The research has zero hip-hop candidates (`grep` over every candidate report), so "barely
  covered it" (PC:155-156) is fair.
- The MIT license exists (`LICENSE`, added `d4e15db`) and is ruled ("dubplate is MIT (Geoff,
  2026-09-07)", streamer spec `:185`).
- The shipped license classes (PC:140-141) match VI:283-285 exactly. The aggregator rule and the
  "own work" rule (PC:142-144) match VI:291-294.
- "No Go" (PC:133-136) fits the rule that nothing is exported ahead of a caller: no package reads
  the records yet, and the claim that a new package would join the `./...` gate is correct.
- The four surfaces (PC:86-88) match VI:242-244, and the closed-club error photo matches VI:263.
- The pool target (100 or more to about 300) and "no cap" are consistent with VI ruling 14.

## Major

### K1. The photo column has three owners across three documents

- Severity: major.
- Location: PC:12-13, PC:193-194; 5c-iii plan `:150`; ROADMAP `:185`.
- Defect: PC routes the photo column to "the rung that first displays a photo" and never names
  it. The 5c-iii plan's Ruling 3, now in execution, says the opposite at `:150`: "Task 8's login
  page as the single column, with the two-column split routed to the curation pass." The same
  plan's handoff (`:1774`, `:1815`) sends "the photo column under Ruling 3" to rung 7. ROADMAP
  `:185` says "rung 7 and 5c-iii consume it", but Ruling 3 means 5c-iii displays no photo.
- Fold: name rung 7 in PC as the consuming rung, which matches the 5c-iii handoff. Record
  errata:
  - 5c-iii `:150`: "routed to rung 7, which consumes the curation pass's records."
  - ROADMAP `:185`: "rung 7 consumes it."

### K2. The spec's legal basis for re-encoding misses most of the candidate set

- Severity: major.
- Location: PC:124-126 and PC:203-204.
- Defect: PC:124 says "Resizing and re-encoding is a technical modification under CC 4.0, not an
  adaptation". The research's candidate licenses are mostly 2.0 and 3.0. A tally of license
  strings across the candidate reports finds about 130 at 2.0 or 3.0 and about 100 at 4.0; the
  spec's own example is `CC-BY-SA-3.0-NL`. The 4.0 carve-out (§2(a)(4)) does not exist in those
  versions. The 2.0 and 3.0 texts allow only modifications "technically necessary" to use another
  medium or format. Shrinking a file to save binary size is a choice, not a necessity.
- Defect: the Changes entry "It remains unmodified in content, only resized and re-encoded"
  (PC:204) is not accurate either. Lossy re-encoding changes the pixels.
- Defect: `nc-licensing.md` Q6 says that for adapted material, the modification notice is a
  license requirement and "easy to omit".
- Fold: keep the WebP choice and drop the "not an adaptation" claim. Add one requirement: every
  shipped image's CREDITS line says "resized and re-encoded from the source", and the record
  carries that fact (a fixed `modified` note, or derive it from `source_sha256` differing from
  `sha256`). That makes the question moot on every license version. The share-alike duty stays
  scoped to the one file either way, which is the VI:249 concern. Reword the Changes bullet to
  say the file is modified.

### K3. The license field's own example is not an SPDX identifier

- Severity: major.
- Location: PC:114.
- Defect: PC:114 reads "An SPDX identifier with its version, such as `CC-BY-SA-3.0-NL`". SPDX has
  `CC-BY-3.0-NL` but no `CC-BY-SA-3.0-NL`. Its CC-BY-SA ports are only `-2.0-UK`, `-2.1-JP`,
  `-3.0-AT`, `-3.0-DE`, and `-3.0-IGO`. The candidate reports include CC BY-SA 3.0 NL (4 hits,
  mostly Nationaal Archief) and other ports.
- Consequence: under a closed enum drawn from SPDX, an agent either fails validation or relabels
  a ported license as the unported one. A relabel ships a wrong license name in CREDITS, which is
  an attribution defect.
- Fold: define `license` as the SPDX identifier where one exists, else a
  `LicenseRef-CC-BY-SA-3.0-NL`-style ref. SPDX's own escape is the `LicenseRef-` prefix. List the
  admitted refs in `vocab.json`. For public domain, SPDX `CC-PDM-1.0` or `CC0-1.0` covers the
  marked cases, and `public-domain` covers the rest.

### K4. A flyer's two license chains do not fit a one-license record

- Severity: major.
- Location: PC:108-121 against PC:144-145 and VI:278-279.
- Defect: VI rules that a flyer "is artwork, so its license must come from the artist, not only
  the photographer", and PC repeats the rule. PC's record holds one `photographer`, one `credit`,
  and one `license`. PC:111 overloads `photographer` to mean "the artist" for a flyer. A
  photograph of a flyer can carry two rights holders: the artist, and the photographer when the
  shot is not a faithful 2D scan. The schema cannot record both, so the rule that PC:144 restates
  cannot be checked from the record.
- Fold: add an optional `artwork` object (artist, license, license_url, credit, verified) that is
  required when `format` is `flyer` or `zine cover`. Keep `photographer` and `license` for the
  image of it, and allow a value of "faithful scan, no separate right" with its basis. CREDITS
  then emits both lines.

### K5. The subject facet no longer supplies the film-edge caption's subject

- Severity: major.
- Location: PC:60-63 and PC:110 against VI:253-255 and VI:289-291.
- Defect: VI's caption contract says the top edge prints "the subject, band, and year" and the
  manifest carries "title, subject, place, year". The manifest generates the captions. PC
  redefines `subject` as a closed, multi-valued category ("sound systems; record stores; …"),
  and it adds a free-text `title`. Read literally, the consuming rung prints "SOUND SYSTEMS" where
  VI meant "Channel One sound system, Kingston". Changes bullet 4 (PC:205) says subjects "join"
  as categories. It does not say that VI's `subject` field changed meaning.
- Fold: state which record fields fill each film-edge slot: `title` for the subject slot,
  `artists` names for the band slot (empty when none), and `year`. Add a fifth Changes bullet:
  VI's free-text subject is now `title`, and `subject` is a facet.

### K6. The first letters drop the lead the owner added

- Severity: major.
- Location: PC:179-181 against VI:360-362 and `photographer-outreach.md:334-341`.
- Defect: VI's closing paragraph names the curation pass's strongest reggae leads as "Beth Lesser
  and the British Culture Archive's People's Archive". The second lead is the owner's own addition
  (outreach `:334`, "Added by the owner"), rated medium-high, with "exactly the scenes… no open
  source has" (skinhead, 2 Tone, and mod, which PC puts in the reggae family and in "everything
  else"). PC's first four letters omit it without a reason, and PC does not list the omission
  among its changes.
- Fold: add the British Culture Archive to the first letters. The alternative is to state why it
  waits and list that in Changes. The first is the obvious answer, since the archive's stated
  terms ("No usage or reproduction of any kind without prior permission") make a letter the only
  route to those images.

## Minor

### K7. The schema check is not a gate, and it needs an unrecorded dependency

- Severity: minor, bordering on major.
- Location: PC:133.
- Defect: "Validation is `schema.json`, run with `ajv` on the workstation after each round." The
  closed-vocabulary guarantee (PC:49-50) rests entirely on that run. Two conflicts follow:
  - CLAUDE.md: "A check either gates or does not exist." Under this design, about 300 records
    commit without a gate.
  - `ajv` is not in `web/package.json` or its lockfile. PC:195 forbids touching `web/`, so the
    run is an unpinned `npx`, which is a new dependency with no adoption record.
- Fold, a light OWNER FORK:
  - Option (a), recommended: rename the step an agent pre-commit check rather than "validation",
    and say plainly that the consuming rung's gate leg is the first enforcement. This costs
    nothing.
  - Option (b): add a `check.sh` leg now, with `ajv` pinned as a web devDependency after lane 1
    releases `web/`.
  - Either way, pin the tool and version the pass uses.

### K8. Rejected and pack records break the future gate leg as written

- Severity: minor.
- Location: PC:128-136.
- Defect: `rejected` and `pack` records keep a record "and no image". PC:136 has the consuming
  rung check "every slug against both files", which fails on every such record. PC also leaves
  two things open:
  - It does not say whether these records live in `photos/`. If they do, a `photos/*.json` embed
    ships them in the binary.
  - PC:151-152 has agents write records with "status unset", but the table's three values do not
    include unset.
- Fold: say where non-picked records live (for example `photos/candidates/`, excluded from the
  embed), and say the pair check applies to `picked` only. Allow status to be absent only in
  scratch.

### K9. The ROADMAP entry needs an erratum

- Severity: minor.
- Location: ROADMAP `:178-185`.
- Defect: the entry says "Curate 100 or more…", "rung 7 and 5c-iii consume it" (see K1), and
  "Owner touchpoint: outreach approvals". PC adds a larger attended touchpoint: Geoff sorts every
  candidate on the sorting page, across rounds that aim at about 300 keeps.
- Fold: update the entry when PC is approved, covering the target, the consumer, and both
  touchpoints (sorting sittings, and reading and sending letters).

### K10. The Changes list omits several departures from VI

- Severity: minor.
- Location: PC:197-207.
- Defect: beyond K2, K5, and K6, PC departs from VI in these places without listing them:
  - **Owner ruling 7** (VI:33-35) names the core as reggae plus punk. PC says it changes only
    the Photography section, but its shares change ruling 7's core as well.
  - **Hip-hop's era.** VI:274 says "early hip-hop and go-go". PC:56-57 says "every era". That
    widens the family and loosens VI ruling 5's "period photograph".
  - **Outsider artists** (VI:273) is missing from PC's "everything else" list (PC:57-58).
  - **"DC hardcore flyers first"** (VI:278) is dropped.
  - **The manifest.** VI:289 says "Every shipped photo has a manifest entry". PC uses one file
    per image, which is compatible but differs.
  - **New rules framed as standing ones.** PC:43-44 says "The deep-cuts rules stand", then adds
    a per-landmark cap that VI does not have. The region cap (PC:34-35) and the coverage floor
    are also new.
- Fold: list each of these, and have the VI pointer erratum cover ruling 7 as well as the
  Photography section.

### K11. Two references to the retired contact sheet remain

- Severity: minor.
- Location: PC:10 and PC:36.
- Defect: "picks every image from contact sheets" and "running counts on the contact sheet" are
  left over from before `67d4adf` moved picking to a scored list and a sorting page (PC:167-171).
- Fold: say "the sorting page" in both places.

### K12. The coverage floor names a subject that does not exist

- Severity: minor.
- Location: PC:38-40.
- Defect: "a landmark" appears in the floor's subject list. A landmark is a Place attribute
  (PC:74-75), not a Subject term (PC:61-63).
- Fold: say "and at least two with a landmark".

### K13. The portrait check cannot be counted

- Severity: minor.
- Location: PC:40-41.
- Defect: "Posed portraits stay a small share of shots, which is how the identity spec's rule is
  checked." No number is given, which contradicts PC's own premise that balance "can be counted
  rather than judged by eye" (PC:9).
- Fold: give a ceiling, for example at most 10% of picked images with shot `posed`.

### K14. The letter voice and the research template conflict

- Severity: minor.
- Location: PC:185-186.
- Defect 1: the outreach template frames the use as "non-commercial" and offers "permission for
  this specific, credited, non-commercial use" (`photographer-outreach.md:307-314`). A grant in
  those terms lands in the pack, never in the binary (VI:285-287). PC asks for CC BY 4.0 but does
  not tell the drafter to drop the template's framing.
- Defect 2: "to be published under the MIT license" is accurate on the license. The repo is
  private (`gh repo view`), and no publication is ruled beyond ROADMAP's `packaging` and the
  Someday `oss-pair`.
- Fold: letters describe dubplate as free and open-source (MIT) and never as non-commercial. Keep
  "to be published" without a date, per `no-calendar-schedule`.

### K15. Smaller gaps

- Severity: minor.
- **The future website** (PC:9, PC:93) is not ruled in ROADMAP, IDEAS, or any spec. Drop it, or
  cite a ruling. A stable slug stands on its own merits.
- **Scratch storage** (PC:128) should name a durable, gitignored location, such as
  `.superpowers/photos/`, which follows VI's precedent for mockups. The session scratchpad does
  not survive between sessions, and candidates wait across rounds.
- **The lane registry.** PC commits to master in the main tree while lane 1 (5c-iii) is open
  there. The lane-launch checklist in CLAUDE.md applies: registry entry, runaway guard, and
  sub-ceiling. PC does not mention it.
- **Image dimensions.** W11.10 lists "no intrinsic image dimensions" as a known failure. Adding
  `width` and `height` to the record costs nothing now. The consuming rung can compute them
  instead, so this is optional.

## Errata owed on approval

- **VI:** the pointer PC promises, plus ruling 7 (the core), Curation rules (VI:269-279),
  Licensing's "manifest entry" and field list (VI:289-291), and Photography's "the shipped file
  stays unmodified" (VI:248-249).
- **ROADMAP `photo-curation`:** the target, the consumer, and the touchpoints (K1, K9).
- **5c-iii plan `:150`:** the two-column split goes to rung 7, not to the curation pass (K1).
