# Photo curation spec review: contract and criteria

Target: `docs/superpowers/specs/2026-09-25-photo-curation-design.md` at `67d4adf`.
Lens: is every promise testable, can any criterion pass vacuously, is the completion condition
well defined, and do the record fields carry enough to check every stated rule.
Line numbers refer to the spec at `67d4adf`.

Counts: 2 blocker, 7 major, 6 minor. Findings are ranked by consequence within each severity.

## Blockers

### C1. The completion condition is unreachable as written, and it can also be met trivially

- **Severity:** blocker.
- **Location:** spec:174-175, with 21-23, 25-36, and 38-41.
- **Defect:** "Rounds continue until the picked pool meets the shares and the coverage floor" is
  the pass's only exit, and it fails in two ways.
  - **Unreachable.** The shares are exact percentages with no tolerance (30/20/20/30). Any pool
    size the pass is likely to reach makes an exact hit a coincidence. The coverage floor has 18
    cells (3 families by 6 subjects). Some cells may have no openly licensed supply at all. Two
    punk images of *sound systems* is the clearest case, since punk had no sound-system culture.
    One empty cell keeps the pass open forever. The "no cap, a round that turns up strong images
    keeps going" rule (22-23) makes this worse, because every added image moves every share.
  - **Trivially reached.** "Geoff's picks can override them" (36) with no record of the override
    means any state can be declared done. The shares then check nothing.
- **Fold:**
  1. Give each share a tolerance band, for example plus or minus 3 points of the picked pool.
  2. State the exit as three conditions:
     - the picked count is at least 250 (or another number Geoff names);
     - every share is within its band, or carries a recorded waiver;
     - every floor cell is at least 2, or carries a recorded waiver.
  3. Define a waiver as one line in a committed `photos/waivers.json`, naming the cell or share,
     the landmark sweeps tried, and Geoff's ruling. A floor cell earns a waiver proposal after two
     targeted rounds come back empty.
  4. Bundle the waiver proposals into one checkpoint question.
- **OWNER FORK:** the pool floor number and the tolerance width. I recommend 250 and plus or
  minus 3 points.

### C2. No checker covers any rule that spans records, yet the pass says balance "can be counted"

- **Severity:** blocker.
- **Location:** spec:8-10, 133-136, 171.
- **Defect:** The only validation named is `schema.json` under `ajv`, and that checks one record
  at a time. Every balance and uniqueness rule spans records:
  - family shares, the region cap, the coverage floor, and the posed share;
  - one image per headliner and one per landmark;
  - no duplicate `source_sha256` or `source_url` (the rejected-record dedupe at 129-130 depends on
    this);
  - slug-to-file agreement.

  The artifact's count bar (171) is the only counter, and it shows family shares only. The done
  condition (C1) cannot be evaluated from anything the spec builds. The "prove it can fail"
  standard also has nothing to attach to.
- **Fold:** Add a workstation-only report script next to the schema, for example
  `photos/tools/count.mjs` or a `jq` program, run after each round next to `ajv`. It prints each
  share with its band, the region maximum, the floor grid, the posed share, and duplicate
  artists, landmarks, and hashes. It exits nonzero on any breach that has no waiver. This adds no
  Go, so the no-Go rationale at 134 still holds. Prove the script can fail by running it once on
  a fixture with a duplicate artist and an empty floor cell. The artifact's bar should read the
  same numbers the script produces. The consuming rung's gate leg can then port the script rather
  than re-derive the rules.

## Majors

### C3. The count base is unstated for records with several scenes or several subjects

- **Severity:** major.
- **Location:** spec:25-41, 52-58, 60-63.
- **Defect:** A record has one primary scene plus any number of secondary scenes, and those may
  span families. It also has one or more subjects. The spec does not say which scene sets the
  family for the shares, or whether a record with three subjects fills three floor cells.
  Counting by any subject lets a single image fill a whole family's floor row, which makes the
  floor pass vacuously. The region cap has the same gap for records without a place, such as
  still lifes and flyers of unknown venue. They are either in the denominator or not.
- **Fold:** State the rules:
  - family comes from the primary scene only;
  - a floor cell counts a record only when the subject is its first-listed (primary) subject, or
    add a `primary_subject` field;
  - the region cap's denominator is picked records that carry a place;
  - no-place records are counted and reported separately.

### C4. The posed-portrait share is not a number, so the rule it claims to check passes vacuously

- **Severity:** major.
- **Location:** spec:40-41.
- **Defect:** "Posed portraits stay a small share of shots, which is how the identity spec's rule
  is checked." No number means no check, and the sentence claims a check it does not provide. The
  base is also unclear, because flyers and zine covers have a shot value that means nothing for
  this rule.
- **Fold:** "Posed stays at or below N% of picked records whose format is photograph." The count
  script reports it.
- **OWNER FORK:** N. I recommend 15%.

### C5. The one-image-per-headliner rule cannot be enforced from the record

- **Severity:** major.
- **Location:** spec:43-44, 81-84.
- **Defect:** No field says who counts as a "headline artist". The mechanism has two holes:
  - The MusicBrainz ID is only "where one exists", so an artist without an ID escapes the lookup.
  - A free-text name is no key: "Bad Brains" and "The Bad Brains" are two keys.

  As written, the rule either applies to nobody or relies on judgment the record does not
  capture.
- **Fold:** Apply the limit to every named artist. That rule is simpler, it holds the deep-cuts
  intent, and it needs no headliner flag. Key each artist by MusicBrainz ID. With no ID, key by a
  normalized name taken from a closed `artists` list in `vocab.json`, the same closed-list
  discipline the other facets use. Allow a second image only through a waiver entry (C1).
- **Alternative:** add `headliner: true` per artist. I recommend against it, because it moves a
  judgment into every agent's record.

### C6. The landmark rule and its "on purpose" exception have no key and no recorded override

- **Severity:** major.
- **Location:** spec:43-44, 74-75.
- **Defect:** Landmarks are exempt from the closed vocabulary (48-49) and keyed by name, with a
  Wikidata ID only "where one exists". Two agents can therefore write "King Tubby's studio" and
  "Tubby's Dromilly Avenue studio" as two landmarks. The exception "unless Geoff picks a second
  on purpose" leaves no trace, so a checker cannot tell a deliberate second image from a
  duplicate that slipped through.
- **Fold:**
  - Put landmarks in a closed list in `vocab.json`, with a slug, a name, and an optional
    Wikidata ID. An agent proposes a new landmark in its round report, and the conductor adds it.
  - Record deliberate seconds in `waivers.json`.
  - The count script flags any landmark slug with more than one picked record that has no
    waiver.

### C7. A flyer's two licenses cannot be recorded, so the artist-license rule cannot be checked

- **Severity:** major.
- **Location:** spec:111, 114, 143-145.
- **Defect:** A flyer or zine cover photographed or scanned by someone else carries two rights:
  the artwork's and the reproduction's. The record has one `photographer`, one `license`, and one
  `verified`. The spec says the license "must come from its artist, not only from the
  photographer who scanned or shot it" (144-145), and the fields cannot show that this was done.
  A flyer record that carries only the scanner's CC BY would validate.
- **Fold:** Require an `artwork` object when the format is flyer or zine cover. It carries the
  artist, the license, `verified` (quote and date), and a source URL. Add a schema conditional
  (`if format in [flyer, zine cover] then required: [artwork]`). Keep the top-level fields for
  the reproduction. If a flat photo of a flat artwork carries no separate rights in the source's
  jurisdiction, record that basis explicitly rather than leaving the field empty.

### C8. Nothing restricts a picked record's license to the shipped classes, and public-domain proof is free text

- **Severity:** major.
- **Location:** spec:114-117, 140-141.
- **Defect:** `license` accepts any SPDX identifier, so a `picked` record under CC BY-NC
  validates. The shipped-class rule (140-141) is prose only. `public_domain` is described as "the
  basis, the jurisdiction, and the US status" with no shape. A record can therefore carry
  `"public_domain": "old"` and pass. Public domain is the riskiest class, because its proof rests
  on dates and facts about publication and renewal, not on a quoted license. Nothing requires
  `public_domain` when `license` is `public-domain`, and nothing forbids it otherwise.
- **Fold:** In `schema.json`:
  1. When `status` is `picked`, `license` is an enum of the shipped classes:
     - `CC0-1.0`;
     - the CC BY and CC BY-SA version identifiers you accept;
     - `LAL-1.3`, the SPDX identifier for the Free Art License;
     - `public-domain`.
  2. When `license` is `public-domain`, `public_domain` is required, with this shape:
     - `basis` from an enum, for example: published before 1930; US federal work; no notice,
       1929-1977; not renewed, 1929-1963; author died more than 70 years ago (life+70);
     - `jurisdiction`;
     - `us_status`, a const for public domain in the US;
     - the fact the basis rests on, as a first-publication year or the author's death year.
  3. Forbid `public_domain` otherwise.

### C9. The spot-check has no sample size, and a license granted by letter has no verifiable source

- **Severity:** major.
- **Location:** spec:117, 158-159, 189.
- **Defect:**
  - "Spot-checks a sample" can mean one record, and it records nothing, so it can pass
    vacuously.
  - `verified` stores a quote and a date but not who checked it or whether the quote came from
    the source page or a later re-check.
  - "A reply that grants a license is verified like any other source" (189) cannot happen as
    written. Verification means quoting the license from `source_url`, and a private email has no
    public page. A redistributor of an MIT project then has no way to confirm the grant.
- **Fold:**
  - Fix the sample size: every public-domain and every flyer record, plus 20% of the rest, with
    at least 5.
  - Write the reader's per-record result into a committed round log. Each row carries the slug, a
    match or mismatch, and the date.
  - For outreach, the letter asks the grantor to post the license on a page they control (a
    Flickr license field, a line on their site). That page becomes `source_url`. If the grantor
    only answers by email, store the grant text, dated, under `outreach/grants/`, and set
    `source_url` to that file with a `grant: email` marker.

## Minors

### C10. Rejected records contradict the consuming rung's slug check, and the dedupe key is unnamed

- **Severity:** minor.
- **Location:** spec:128-136.
- **Defect:**
  - A `rejected` record is committed with no image, but the future gate "checks every slug
    against both files". Every rejected record will fail that check.
  - "Later rounds do not find it again" names no key to match on.
  - Records are written "with status unset" (152), but the table lists only three values.
- **Fold:**
  - State that the slug check covers `picked` records only.
  - Name the dedupe key: `source_url` plus `source_sha256`.
  - Make `status` optional in the schema, and required for any committed record.

### C11. The claim that re-encoding is a technical modification is scoped to CC 4.0, but the pool holds CC 3.0 records

- **Severity:** minor. The defect is a correctness gap in a claim, and the risk is low.
- **Location:** spec:124-126, 202-204.
- **Defect:** The spec's own example is `CC-BY-SA-3.0-NL`, and CC 3.0 has no Section 2(a)(4)
  technical-modification clause. CC 4.0 also still asks the licensee to indicate a modification
  in the attribution.
- **Fold:**
  - Cite Creative Commons' guidance that a format change is not an adaptation under any version,
    or scope the claim per version.
  - Add a `modified: "resized and re-encoded to WebP"` note that CREDITS can print.
  - Confirm the example is a real SPDX identifier. `CC-BY-3.0-NL` is listed; I am not sure
    `CC-BY-SA-3.0-NL` is. Use `LicenseRef-` for anything off the list.

### C12. The screening criteria cannot be checked from the record

- **Severity:** minor.
- **Location:** spec:161-163.
- **Defect:** "Too small for a 1600px display" and "can crop to the tall column" leave no
  evidence behind.
- **Fold:**
  - Record `source_width` and `source_height`.
  - Require a long edge of at least 1600 for the shipped file, or record the upscale rule if one
    is intended.
  - Leave the crop judgment to Geoff's sort.

### C13. The year has no representation for unknown or approximate dates

- **Severity:** minor.
- **Location:** spec:77-79.
- **Defect:** Many period photos and flyers are dated "circa" or not at all. Without a
  convention, an agent will guess, and the derived decade will be wrong. A public-domain basis
  also depends on the publication year, which can differ from the year the image was made (C8).
- **Fold:** Use `year` plus `year_precision`, one of `exact`, `circa`, or `decade`.

### C14. No field can express the error-surface criterion, and the surfaces have no floor

- **Severity:** minor.
- **Location:** spec:86-88.
- **Defect:** "A club's final night" is not expressible in the vocabulary. The error and empty
  surfaces also have no floor, so the pass can close with no error-page photo at all.
- **Fold:**
  - Add a `closed-venue` subject, or drop the final-night clause.
  - Add floor cells: error at least N and empty at least N, reported by the count script.

### C15. The sort page cannot mark a candidate for the pack

- **Severity:** minor.
- **Location:** spec:130-131, 167-172.
- **Defect:** The page offers keep or pass only, so the path that sets `pack` is undefined. It is
  also unclear whether NC candidates appear on the page at all.
- **Fold:** Either the agent sets `pack` at screening and the page excludes those candidates, or
  the page adds a third key. State which.
