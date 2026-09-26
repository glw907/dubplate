# Photo curation round 1

Plan: `docs/superpowers/plans/2026-09-25-photo-curation-round-one.md`. Spec: revision 2.3.

## Task 4 checkpoint (search closed, 2026-09-25)

Seven search agents ran in parallel. `check.mjs sheet` over the candidate home wrote 114
candidates with no cross-agent duplicate.

| Agent | Kept | Main drop reasons |
|---|---|---|
| research-a | 38 | one too-small |
| research-b | 27 | 11 license (PDM marks, PD-self), 5 too-small, 11 redundant |
| research-c | 18 | ~22 too-small, ~13 license (implausible own-work claims) |
| hiphop | 21 | 6 too-small, 4 license, 3 PD-self (no basis yet) |
| flyers | 7 | 17 license, 16 too-small |
| research-d1 | 2 | 28 too-small, 40 off-topic |
| research-d2 | 1 | 28 too-small, 41 off-topic |

By family: reggae 22, punk 25, hip-hop 22, everything else 45. By primary subject: performers
53, clubs and venues 20, crowds and dancers 14, streets and neighborhoods 12, sound systems 7,
recording studios 6, record stores 1, pressing plants and cutting lathes 1. By region: North
America 39, Europe 24, no place 23, UK and Ireland 12, Caribbean 8, Latin America 5, one each
MENA, Asia and Pacific, Africa.

Findings:

- The unverified Flickr harvest yielded 3 of 140 rows; later rounds skip keyword harvests.
- The DC Punk Archive's Dig DC flyers are all marked In Copyright; the outreach letter is the
  only route.
- Beth Lesser's archive is 479px web copies; only her letter can reach it.
- Round-two leads: hiphop's 8 vetted, undownloaded candidates (its `manifest.json`);
  research-a's unpursued Skatalites and Notting Hill 1989 frames; the HABS 9:30 Club interior
  (year unknown); reggae record stores and sound systems; reggae's share.

Conductor corrections owed before the sitting:

- Drop the Sun Ra poster (flyers): its embedded 1973 photo has no verified license.
- Add a `self-dedication` public-domain basis (Commons `{{PD-self}}`) and recover the dropped
  PD-self images (1520 Sedgwick Avenue, two Bronx 2009 Kool Herc and Coke La Rock frames, two
  CBGB exteriors).
- Add the `off-topic` reason code; research-d1 and d2 logged off-topic drops as `passed`.
- Add the `sonidero` and `rebetiko` scenes to the everything-else family.
- Tighten `SOURCE_ID_RE`: `commons:<digits>`, `flickr:<digits>`, `archive:<name>:<key>` with
  dots allowed in the name.
- The spot-check tests every public-domain and flyer record, with particular attention to the
  1978 to 1989 no-notice-unregistered basis (the Black Flag flyer is likely Raymond
  Pettibon's), the Gottlieb jazz photos' basis (a donor dedication, not government work), and
  the 1976 Marley press photo's no-notice claim.

Operational lessons for later rounds (to the plan's Global constraints): wrap every download in
`timeout -k 10 150` as well as curl's own `--max-time`, which the sandbox did not always
enforce; kill only PIDs an agent recorded itself, never by pattern (one agent's `pkill -9 -f`
on the shared lock killed every agent's queued downloads); never have a search agent start
downloads it cannot finish within the round.

Spend: roughly 4.8M of the 5.5M ceiling at this checkpoint, the 4.4M flag crossed during the
search.

## Resume after the power loss (2026-09-25)

The conductor session ended mid-question when the laptop lost power. A fresh session resumed from
this log. The sorting page is published at `https://claude.ai/artifact/7CewFiyc6pE1SG1a4pfQPE`
(Task 2's close owed this line). Every Task 6 dispatch names it.

Owner rulings (Geoff, 2026-09-25):

- Ceiling raised to 7.0M; the 80% flag moves to 5.6M.
- Text-only flyers count as public domain under the threshold of originality (Commons
  `{{PD-text}}`), limited to purely typographic flyers with no drawing, logo art, or photo. The
  spot-check views each such flyer to confirm. This recovers the three 1984 to 1988 West Coast
  hip-hop flyers.
- The sorting page shows each photo in the context of the login page: the silver-print
  treatment and film-edge caption in the login page's photo column, beside the form, at desktop
  and phone widths. The crop to that column is the judgment Geoff makes while sorting.
- The sorting page gains a one-at-a-time review mode with yes, no, and maybe (Geoff: "cycle
  through them with a yes/no/maybe for each"). Conductor ruling: a maybe is treated like an
  unsorted candidate at close, staying in `.candidates/` for the next round, never rejected.

## Task 5 spot-check (2026-09-25)

Reader: a fresh-context Opus spot-check reader. Raw evidence is under
`photos/.candidates/_spotcheck/raw/`, and the drawn ID list is in `_spotcheck/sample.json`.

**Sample.** `node tools/photos/check.mjs sheet --out photos/.candidates/_sheet/` wrote 121
candidates and dropped 0 duplicates. The mandatory set holds every record with license
`LicenseRef-public-domain`, a `public_domain` block, or format `flyer` or `zine-cover`: 34
records, which include all 9 flyers. The rest holds 87 records. Method: sort the rest ascending
by the hex `sha256("20260925:" + id)`, then take the first `max(5, ceil(0.2 × 87))` = 18. Seed
`20260925`, pool 121, sample size 52.

**Channels.** Commons evidence came from `index.php?curid=<id>&action=raw`, plus
`Template:LABW` and `Template:MikaVäisänen` raw and one batched `imageinfo` upload-history
query. Commons deletion-request pages were checked for eleven files, and none exists. No
sampled page carries a delete, copyvio, no-permission, or no-source template. Flickr evidence
is each photo model's `"license":<n>` together with the license URL rendered on the page. Every
Flickr owner NSID (94086509@N00, 68537341@N00, 61473396@N04) and every Flickr-origin Commons
account is absent from `Commons:Questionable_Flickr_images` and its `/Users` list. LOC evidence
came from `item/<id>/?fo=json`, reading `item.rights`. Registration checks used the Copyright
Office public-records API.

**Tally.** 36 match, 16 mismatch, 0 changed-since-grant, 0 unverifiable. The Marley record's
upstream provenance could not be reached (archive.ph 429, worthpoint 403). Its Commons evidence
was verified, so it is logged as a match with a note. Every mismatch sends the round back for
re-verification before Task 6, per the spec.

### Mismatches

| Record | Record claims | Evidence shows | Recommended fix |
|---|---|---|---|
| archive-loc-2023867575 | basis `us-government-work` | `item.rights`: "In accordance with the wishes of William Gottlieb, the photographs in this collection entered into the public domain on February 16, 2010" | basis `rights-holder-dedication`, quote that line |
| archive-loc-2023867585 | basis `us-government-work` | same Gottlieb dedication line (LOC date "[1938?]", title "between 1938 and 1948") | basis `rights-holder-dedication` |
| archive-loc-2023867629 | basis `us-government-work` | same Gottlieb dedication line | basis `rights-holder-dedication` |
| archive-loc-2023867716 | basis `us-government-work` | same Gottlieb dedication line | basis `rights-holder-dedication` |
| archive-loc-2023868971 | basis `us-government-work` | same Gottlieb dedication line | basis `rights-holder-dedication` |
| commons-155873916 (Black Flag flyer) | `us-1978-1989-no-notice-unregistered`, artist "Unknown" | Viewed: the flyer carries Raymond Pettibon's drawing (captioned "If you're really Christ, why don't you raise Janie?") and the four-bar logo, so the artist is not unknown. The file is a 2024 phone photo of a wall showing at least six other flyers (Masque, Anti-Nowhere League, Lewd, Punk Picnic, Nazi…), each a separate artwork the record never analyses. The only no-notice evidence is the uploader's `\|permission= The posters appear all to be from the era when they would have needed copyright notices (which they lack)`. The Copyright Office search found no registration for "Black Flag Meat Puppets" (0 hits); the "Pettibon" name search shows a 1987 VA registration (*Mulholland Drive*), so he did register work in the window | drop (`license`): multiple third-party artworks in frame, misattributed artist, unregistered leg unevidenced |
| flickr-2255257225 | `us-1978-1989-no-notice-unregistered`, 1989 | The flyer reads "THURSDAY, MAY 25" and "SATURDAY June 3". Both dates fall on those weekdays only in 1989, after the Berne date of 1 March 1989, when notice stopped being required, so the no-notice basis cannot apply. It also reproduces Cruz Records ad copy and the ALL mascot art | drop (`license`) |
| flickr-2253305393 | `us-1978-1989-no-notice-unregistered`, 1987 | Flickr `"license":4`, `creativecommons.org/licenses/by/2.0/` (the scan only). No notice is visible. The record carries no registration evidence. My Copyright Office search for "Leaping Lizards Productions" returned 7 hits, none 1978 to 1995. The cartoon crowd drawing's source is unknown | re-verify: add the registration search to `verified` and keep only if the conductor accepts the drawing as the promoter's own; else drop |
| flickr-2256260222 | same basis, 1988 | No notice, and no registration evidence in the record. It embeds the Dischord and Enigma label logos and a large cut-and-paste block of a Chinese-language newspaper, a foreign work the US no-notice rule does not reach | drop (`license`) |
| flickr-2274222503 | same basis, 1988 | No notice, and no registration evidence. It embeds the Fist Records logo, a halftone fist image, and the NYHC logo | drop (`license`) |
| flickr-2274349509 | same basis, 1986 | No notice, and no registration evidence. It redraws the Descendents' "Milo" character, published with notice on the band's record sleeves, and the JFA logo | drop (`license`) |
| commons-16927339 | photographer Danny Lyon, year 1973 exact | `Author = {{NARA-Author\|Pickerell, Jim, 1936-, Photographer\|4588217}}`, `Date = 1974-07`. The license `{{PD-USGov-EPA}}` matches | photographer Jim Pickerell, year 1974 |
| commons-17083273 | photographer Danny Lyon | `Author = {{NARA-Author\|Calonius, Erik, Photographer\|8464439}}`, `Date = 1973-05`. The license `{{PD-USGov-EPA}}` matches | photographer Erik Calonius |
| commons-17083311 | photographer Danny Lyon | `Author = {{NARA-Author\|Calonius, Erik, Photographer\|8464439}}`, `Date = 1973-05`. The license `{{PD-USGov-EPA}}` matches | photographer Erik Calonius |
| commons-17032876 | photographer "Unknown, Commons uploader", no plausibility note | `\|author=[[User:Revup67\|Anthony Gliozzo (revup67)]]` and `{{self\|cc-by-sa-3.0}}`. The photo is dated 1984 and was first uploaded 2011-10-19, 27 years later, so the plan's own-work rule applies, and `verified` gives no reason. The account has only these two uploads, both from one 1984 Irving Plaza show | photographer Anthony Gliozzo, and add the plausibility statement (named real person, backstage access, a single personal set), or drop |
| commons-17032958 | same | `\|author=[[User:Revup67\|Anthony Gliozzo (Revup67)]]`, `{{self\|cc-by-sa-3.0}}`, uploaded 2011-10-19 | same as above |

### Matches

The date for every row is 2026-09-25.

- FSA/OWI (Marion Post Wolcott), `us-government-work`: archive-loc-2017801730, -2017801734,
  -2017801737, -2017801843, -2017802165. Evidence: "The contents of the Library of Congress Farm
  Security Administration/Office of War Information Black-and-White Negatives are in the public
  domain".
- commons-139148268 (Kinks): `{{PD-Scan|PD-Pre1978}}` and "they did not include a copyright
  notice on either the photograph or as part of its caption".
- commons-26925953 (Standells): `{{PD-Pre1978}}` and "There are no copyright markings as can be
  seen at the full view link".
- commons-58940407 (Chocolate Watchband): `{{PD-1978}}` and "All versions of and other unscanned
  portions of the back of this photo do not display the copyright notice". This is weak
  provenance, since the source is a blogspot scan with an unknown author.
- commons-168617945 (Marley 1976): `{{PD-1978}}` and "published in 1977 by the US branch of
  Island Records ... No copyright markings on the photo". The no-notice claim rests on the
  uploader's reading of an eBay listing. The listing archive (archive.ph 429) and the worthpoint
  copy (403) were unreachable, so the claim is unconfirmed upstream. No deletion request exists.
  The record's `verified` field quotes only the template and should also quote the no-notice
  line.
- commons-165167895 (Duperly, "before 1922") and commons-165168466 (Valentine & Sons, 1891):
  `{{PD-US-expired|PD-Jamaica-old-1946}}`.
- EPA DOCUMERICA, `{{PD-USGov-EPA}}`: commons-16911725, where the author (Danny Lyon) is right.
  Its `Date = 1974-07` against the record's circa 1973 is a minor correction.
- `{{PD-self}}`, `rights-holder-dedication`, where the uploader is the named author and the
  upload was contemporaneous:
  - commons-2521430: `{{PD-self|author=[[User:Schlamniel|Schlamniel]]}}`, dated 2006-07-21,
    first uploaded 2007-08-05 by Schlamniel.
  - commons-4599589: dated 2008-07-30, uploaded 2008-08-21 by Sbazzone.
  - commons-6163794 and commons-6172376: dated 2009-02-28, uploaded 2009-03-09 by
    Bigtimepeace.
  - commons-7357632: dated 2009-07-21, uploaded 2009-07-21 by Bigtimepeace.
- `{{PD-text}}`, `below-threshold-of-originality`, each flyer viewed:
  - commons-53201399: type, star ornaments, and the printer's tiny imprint only. The imprint
    reads "COLBY POSTER PRINTING CO.", so the record's "Cooley" (copied from Commons) is a
    spelling fix.
  - commons-61657014 (1984-06-16): type and two star bullets only.
  - commons-61686488 (1984-08-10): type and star bullets only.
  All three predate 1 March 1989 and show no notice.
- Commons self-licenses:
  - commons-10429450 `{{self|cc-by-sa-3.0|GFDL}}`. Its `verified` quotes api extmetadata, an
    aggregator field, so re-quote the wikitext.
  - commons-125742880 `{{self|cc-by-sa-4.0}}`.
  - commons-19445529 `{{self|cc-by-3.0}}`.
  - commons-4235509 `{{self|GFDL|cc-by-3.0}}`.
  - commons-55096678 `{{self|cc-by-sa-4.0}}`.
  - commons-172710925 `{{self|cc-by-sa-4.0}}`.
  - commons-38430378 `{{self|cc-by-sa-4.0}}`, with `{{MikaVäisänen}}`, which carries
    `{{PermissionTicket|id=2015020510019074}}` and confirms that Mika Väisänen uploads his own
    pictures from the Mikamote and MikaV accounts. It was dated 1999 and uploaded 2015-02-04,
    16 years apart.
- commons-164948851: `{{cc-by-sa-4.0}}` with `{{PermissionTicket|id=2025050810011748}}`.
- commons-171049547: `{{LABW|...}}`, and `Template:LABW` renders `{{cc-by-4.0|{{{1}}}}}`.
- commons-73912446: `{{CC-zero}}` under `{{Nationaal Archief}}`, credited "Rob Bogaerts / Anefo".
- Flickr-origin Commons files, with FlickreviewR or Flickrreview as the evidence of record:
  - commons-44052504: `{{User:FlickreviewR/reviewed-pass|Sam Howzit|...|2015-10-07 03:32:19|cc-by-2.0|}}`.
  - commons-2583699: `{{User:FlickreviewR/reviewed-pass|DaigoOliva|...|10:04, 19 August 2007 (UTC)|cc-by-sa-2.0}}`.
  - commons-40564624: `{{cc-by-sa-2.0}}` and `{{Flickrreview|INeverCry|2015-05-29}}`.
  - commons-110934684: this crop's page carries no review. Its parent,
    `File:Michael_Schwartz,_Mix_Master_Mike.jpg`, carries
    `{{User:FlickreviewR/reviewed-pass|fabbio|...|10:27, 11 September 2007 (UTC)|cc-by-sa-2.0}}`.
- Flickr direct, where the live page matches the record, so there was no change since the grant:
  - flickr-4075403267: `"license":4` and `creativecommons.org/licenses/by/2.0/`.
  - flickr-21014404506: the same.

### Round-wide findings

- The hiphop agent's `verified` fields never quote a line. They say "Raw wikitext license
  template confirmed ... see raw/ evidence in the hiphop agent directory". This affects 8
  sampled records and likely all 21 of that agent's records. The license itself matched in every
  sampled case, but the plan requires the matched line in `verified`.
- research-a's Commons records quote only the bare license template, never the no-notice or
  authorship lines the basis rests on. The Marley, Watchband, Standells, and Fuzztones records
  show this.
- The `us-1978-1989-no-notice-unregistered` basis failed in every sampled use (6 of 6). No record
  carried registration evidence, and flyers routinely embed third-party logos and character art
  that a flyer's own lack of notice does not free.

## Re-verification (2026-09-25)

The spot-check's mismatches sent the round back. Three agents re-verified every remaining record
by directory, so each `verified` field now quotes the matched source line, and each public-domain
record also quotes the line its basis rests on.

- Dropped: all six flyers resting on the 1978 to 1989 no-notice-unregistered basis (Black Flag
  and five Kopper flyers; embedded third-party art, no registration evidence);
  `commons:188415528` (Flickr source gone before FlickreviewR, no archive capture);
  `commons:150285417` (a wall of third-party Fela posters); `commons:158743918` (the Slits;
  upstream Flickr is CC BY-NC-ND 2.0, no evidence of a freer grant). Conductor ruling:
  `commons:168617945` (Marley 1976) moves to round two as `unavailable`, because its no-notice
  evidence (eBay via archive.ph and worthpoint) answered 429 and 403 twice.
- Corrected: the five Gottlieb records to `rights-holder-dedication`; the Fuzztones credited to
  Anthony Gliozzo and kept (author field and a file description name him, a two-file account from
  one 1984 show); DOCUMERICA credits (Pickerell, Calonius) and dates; the Colby printer imprint;
  own-work plausibility statements for eight research-a records with twenty-year-plus gaps.
- Pool after re-verification: 111 candidates.
- Garage (Geoff asked): the Standells and the Chocolate Watchband (1966, public domain), Q65 on
  the Scheveningen pier (1966, CC0), and the Fuzztones (1984). Round two targets 1960s garage and
  the 1980s revival as a named gap.

## Round open (2026-09-25)

The sorting page gained review mode with yes, no, and maybe inside a login-page mockup
(`46b7e39`, `71b957f`, `15c32cb`, `9aa235e`, `ca41952`; diff-reviewer `fix` then `accept`) and
is published as version 5 at the same URL. Round open encoded 222 files from 111 candidates,
uploaded all 222 as assets in 9 calls, wrote `meta/vocab`, and seeded the test round
(`rounds/test`, `candidates/fixture:1` to `fixture:3`). Round one's 112 writes wait in
`photos/.candidates/_open/seed-{0,1,2}.json` for Geoff's test-round try. The plan's 150-candidate
approval estimate is superseded by these counts.

## Sitting notes (2026-09-25)

- Geoff: "Some of these pics are good, but have cropping issues." The sorting page gains a per-photo
  focus point (`focus: {x, y}` on the choice document), copied onto the committed record at close.
- Geoff, on the Eazy-E flyer: "It's great, but the top is cut off." Conductor ruling: flyers,
  posters, and zine covers are never cropped; they fit whole on the film-edge carrier. The real
  login page needs the same rule and the focus point; route both to the web lane's login-photo
  work and to the visual identity spec's Photography section at round close.
