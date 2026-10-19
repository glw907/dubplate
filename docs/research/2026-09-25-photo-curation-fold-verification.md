# Photo curation spec: fold verification

Target: `docs/superpowers/specs/2026-09-25-photo-curation-design.md`, revision 2 at `881ad32`
(cited as PC). Fold record: `2026-09-25-photo-curation-fold.md`. Reviews: C, M, D, K.
Reader: fresh context, no part in the fold. Probes ran on 2026-09-25 in
`/tmp/claude-1000/verify-photo/`, outside the repo.

Counts: 0 blocker, 2 major, 7 minor. Findings are ranked by consequence.

## 1. Did each blocker and major close?

All three blockers closed where the record cites them.

- **C1**: closed at PC:28-61. It adds plus or minus 3 point bands, waivers in
  `photos/meta/waivers.json`, a waiver proposal after two empty targeted rounds, and a **Done**
  that requires a clean balance report and Geoff's declaration. The review's pool-floor number
  was not adopted. That holds, because done is Geoff's declaration and ~300 is his target.
- **C2**: closed at PC:181-198 with `check.mjs`, which has integrity and balance modes and a
  failing-fixture proof. The gate leg is deferred as a queued edit because lane 1 owns
  `scripts/check.sh`. That reason holds, and the conductor runs the check by hand before each
  photo commit in the meantime.
- **D1**: closed at PC:172-176 (`photos/.candidates/`, durable and gitignored) and PC:291-293
  (a keep applies only to a file that matches its hash).

Majors:

- **C3, C5, C6, C9, M2, M4, M5, M6, D2-D5, K1-K6**: closed at the cited locations. The refused
  parts all hold: C5's closed artist list, C6's closed landmark list, D4's temp-and-rename step,
  D5's perceptual hash, and D3's reconciliation line.
- **C4**: closed at PC:40-41 (15%, waivable).
- **C7**: closed at PC:154. Its enforcement stops at the top-level license (V1).
- **C8**: closed only in part. The fold record does not say so (V1).
- **M1**: closed at PC:227-234. The "on any reading" claim is overstated (V8).
- **M3**: routed to Ruling 1. The ruling's evidence is misstated (V2).
- **D6**: folded with `archive_url` made optional ("when one succeeds", PC:149). D6 asked for it
  to be required on picked and pack records. The weakening is defensible, but the record lists
  D6 as fully folded (V7).
- **D7**: deferred to the queued edit, the same deferral as C2. It holds.

## 2. Contradictions and process order

- No check-before-its-object inversion was found. Agents hash at download, the Opus sample runs
  before the sort, and the round close runs keep, encode, integrity, then commit.
- Four order and consistency defects remain: V3 (the commit mechanic's race), V5 (the MBID
  step), V6 (pack records), and V9 (the embed pattern). V1's `CC-PDM-1.0` case is a
  contradiction between two rules of the license field.

## 3. New mechanisms from memory

| Mechanism | Backing |
|---|---|
| curl UA / python-requests 403 | M2 probe and the Wikimedia policy quote |
| MusicBrainz 1 req/s, UA, 503 | M11 quote |
| Artifact 25-per-call, outside-working-directory | the Artifact tool contract ("up to 25 ... in one call"; "a file outside the working directory ... each go in a call of their own") |
| `git add` then `git commit -- <paths>` | the fold's probe, reproduced here on git 2.55.0; it hides a race (V3) |
| cwebp `-metadata icc` through a PNG intermediate | unprobed by M9 and M10. Probed here: a 21012-byte Adobe RGB profile survives `magick → PNG → cwebp -metadata icc`, and EXIF `Copyright` is dropped as stated. The same probe found V4. |
| SPDX IDs | the SPDX 3.29.0 list, fetched here: `CC-BY-2.0`, `CC-BY-SA-4.0`, `CC0-1.0`, and `LAL-1.3` exist; `CC-BY-SA-3.0-NL` does not. The list also carries `CC-PDM-1.0`, which the spec's rule then routes into the `public_domain` ban (V1). |
| Wayback archiving | unprobed by the fold. Probed here (below and V7). |
| Node check script | stdlib only; nothing to probe |

The Wayback probe: an anonymous `GET https://web.archive.org/save/<url>` first drew `429`. A retry
about 20 seconds later drew `302` with `location: https://web.archive.org/web/<timestamp>/<url>`,
under both a browser UA and a descriptive UA. The mechanism works, and it is rate-limited, so
"when one succeeds" is the right hedge.

## 4. The owner fork, and settled items

Ruling 1 cannot be ruled well on the evidence as written (V2). The settled items were checked.
The C1 bands, C4's 15%, D6's grant record, D8's git history, and hip-hop's "every era" are each
stated in the spec body or in Changes, so approving the spec approves them. Each is waivable or
has one clearly dominant option. D8 comes closest to a risk-acceptance fork: a withdrawn image
stays in public history. The alternative (LFS or R2 for about 300 files) is plainly
disproportionate, and PC:238-240 states the tradeoff where Geoff will read it. No settled item
needs to become a question.

## 5. Proportionality

The spec stays proportionate. Its weight sits in licensing provenance, which is the substance of
shipping third-party images inside an MIT binary. Scripts carry the counting rules, so they cost
Geoff no attended time. One trim is worth considering: the serialized MusicBrainz step
(PC:102-106) exists only to key the one-image-per-artist rule. A normalized name catches nearly
the same duplicates, and Geoff sees every candidate in one grid. Dropping MBIDs from this pass,
or leaving them optional, removes a rate-limited network step and V5 together. This is not a
finding; it is the fold's call.

## Findings

### V1. The shipped-class and public-domain proof rules miss the artwork license, the US status, and `CC-PDM-1.0`

- Severity: major.
- Location: PC:151-154, PC:184-186.
- Defect: C8 and C7 were folded for one harm: a picked image whose license does not permit
  shipping. Three paths still reach it.
  1. `public_domain` records "the US status" (PC:152) but never requires it to *be* public
     domain in the US. C8 asked for exactly that constraint. A record reading "under copyright
     in the US" passes integrity.
  2. `artwork` carries its own license (PC:154). The integrity check tests only "a picked
     record's license" against the shipped classes (PC:185-186). The `artwork` object has no
     `public_domain` proof either, although public domain (US no-notice or non-renewal) is the
     most likely basis for a period flyer. An NC-licensed flyer artwork under a CC BY scan
     passes, and that is C7's own example.
  3. The license list uses "the SPDX ID where one exists" (PC:151), and SPDX 3.29.0 has
     `CC-PDM-1.0` (probed). A Public Domain Mark item therefore takes `CC-PDM-1.0`, which makes
     `public_domain` "forbidden" (PC:152). The riskiest class then ships with no basis at all.

  Every flyer is in the Opus sample, but that reader checks the quote, not the PD basis.
- Fold: one sentence in PC:152 and one in PC:154. The US status must be public domain. The
  public-domain class always takes `LicenseRef-public-domain`, and a mark such as `CC-PDM-1.0`
  goes into `verified` as evidence. `artwork.license` comes from the shipped classes, and it
  carries `public_domain` under the same rule. Add an NC-artwork flyer to the check's fixtures.

### V2. Ruling 1's evidence misstates what "yes" saves, so Geoff cannot weigh it

- Severity: major.
- Location: PC:357-363.
- Defect: the Commons size query was re-run here on 137 names from the research reports, and
  113 resolved, matching M3's 113. Their long edges fall into four bands:

  | Long edge | Candidates |
  |---|---|
  | 1600px or more | 62 |
  | 1200-1599px | 15 |
  | 1000-1199px | 6 |
  | under 1000px | 30 |

  "Yes" at 1200px saves 15 of the 51 sub-1600 candidates and still drops 36 (32% of the set).
  The ruling says "no" "drops close to half ... and thins the period record", which implies
  "yes" keeps the period record. The two period examples it cites, Bundesarchiv at 568×608 and
  Tropenmuseum at 700×536, fall under 1000px, so they are dropped under either answer. M3
  offered a floor of 1000 to 1200, but the choice of 1200 over 1000 moves only 6 files. "A
  login column that renders them without upscaling" is also a promise to rung 7 with no stated
  column size. The identity spec gives no pixel size for the column. At DPR 2, a full-height
  column likely needs more device pixels than 1600 anyway.
- Fold: replace the evidence with the table above. State that "yes" saves 15 files, that the
  cited Bundesarchiv and Tropenmuseum examples need an archive's larger scan under either answer,
  and that 1000px would save 6 more. Drop "without upscaling" from the Yes branch, or reword it
  to "stored without upscaling; rung 7 sizes the column".

### V3. `git add` then `git commit -- <paths>` reopens the race D2 closed

- Severity: minor. The window is small, and a failed or empty photo commit exposes it.
- Location: PC:298-301.
- Defect: probed on git 2.55.0. After `git add -- p2`, another lane's plain `git commit`
  commits `p2` along with its own file. D2 banned a separate `git add` for that reason, and the
  fold's correction reintroduced it to handle untracked paths. With `git add -N -- p3`
  (intent-to-add), the other lane's plain commit leaves `p3` out, and `git commit -- p3` then
  commits it with full content. A withdrawal's deletion commits through `git add -- p` or
  `git commit -- p` alike.
- Fold: "stages new paths with `git add -N -- <paths>` only, then `git commit -- <paths>`".

### V4. The pinned encode stores EXIF-rotated photos sideways

- Severity: minor. The defect is visible and cheap to fix, but it silently corrupts committed
  files and their `width` and `height`.
- Location: PC:164-166.
- Defect: probed on a 2400×1200 JPEG with EXIF Orientation 6. `magick <src> -resize
  '1600x1600>' tmp.png` then `cwebp -q 80 -m 6 -metadata icc` gives a 1600×800 WebP with
  orientation `Undefined`, so a portrait photo is stored as a sideways landscape. ImageMagick
  does not auto-orient during `-resize`, and cwebp drops the orientation tag with the rest of
  the EXIF.
- Fold: `magick <source> -auto-orient -resize '1600x1600>' <tmp>.png`, and the same for
  thumbnails.

### V5. The MBID step and the duplicate-artist rule have no place in the round order

- Severity: minor.
- Location: PC:102-106, PC:189-190, PC:286-301.
- Defect: the serialized MBID lookup appears in none of the five round-close steps. Integrity
  "flags" a second image of one artist or landmark (PC:189). If "flags" means fail, a keep
  stalls the commit until a waiver question reaches Geoff, although waivers are bundled at
  checkpoints (PC:57-58). If it means warn, the rule is unenforced. Adding MBIDs after a commit
  can also expose duplicates that name keys missed.
- Fold: run the MBID step in round-close step 3, before integrity. Rule that a `Y` on a
  candidate the sorting page marks as a second image of an artist or landmark writes the waiver
  entry, so integrity can fail hard without a mid-close question. Alternatively, drop MBIDs
  (see section 5).

### V6. Pack candidates never reach `pack.json` in the round order

- Severity: minor.
- Location: PC:176-179, PC:256, PC:286-296.
- Defect: agents set `pack`, and the page hides pack candidates, but no step moves them into
  `photos/meta/pack.json` or strips their `screen`. Agents cannot write the one shared file
  safely in parallel. D9's durable-storage point for pack files is silently dropped. A re-download
  checked against `source_sha256` at packaging time is an adequate answer, but the record does
  not state it.
- Fold: step 3 adds "each pack candidate goes into `pack.json` with `screen` stripped". Note in
  the fold record that pack files are re-fetched and hash-checked at packaging.

### V7. The Wayback mechanism is unnamed, and D6's weakening is undisclosed

- Severity: minor.
- Location: PC:149, PC:220-222, PC:268-270; fold record, row C9/D6.
- Defect: `archive_url` became optional, but the record lists D6 as fully folded. For a
  Flickr-direct candidate, the grant-time license rests on "an archive capture" (PC:221). When
  capture fails, which is plausible given the `429` probed here, the only grant-time evidence is
  the `verified` quote.
- Fold: one clause. Capture is an anonymous `GET web.archive.org/save/<source_url>`, retried
  after a `429`, and the snapshot URL is the `302` `location`. A Flickr-direct record without a
  capture rests on its raw `verified` quote and `retrieved_at`. Mark D6 as folded in part.

### V8. "On any reading" ignores the render-time crop and tint

- Severity: minor.
- Location: PC:232-234.
- Defect: M1 named the login column's render-time crop as a modification that CC 4.0 requires
  to be indicated. The silver-print CSS filter is another. The CREDITS line names only the
  resize and the WebP conversion, yet the spec says this meets the duty "on any reading".
- Fold: "CREDITS says the image was resized, converted to WebP, and is displayed cropped and
  toned." Rung 7 owns the wording.

### V9. The embed pattern note contradicts its own package placement

- Severity: minor. This is a note for rung 7.
- Location: PC:202-204.
- Defect: the embedding package "must sit inside `photos/`", and `go doc embed` says patterns
  are relative to the package directory and may not contain `..`. From inside `photos/`, the
  pattern `photos/*.webp photos/*.json` names `photos/photos/...`.
- Fold: "a pattern such as `*.webp *.json`".

## Process note

The size re-query sent Geoff's email address to the Commons API in its User-Agent string. That
was an error by this reader. No other probe sent it.
