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
- Geoff, on the Chocolate Watchband group portrait: "The crop is off ... otherwise it would be
  great." Conductor ruling: each photo carries a fit, `cover` (with the focus point) or `contain`
  (whole, on the film-edge carrier), chosen on the sorting page and copied onto the record at
  close. Flyers, posters, and zine covers default to `contain`. The login page honors both fields.
- Sorting page v6 to v8 (`f204740`, `a9059dd`, `fca3cc2`): the treated login view is the only
  large view, with a focus point (an image-space point) and a fit per photo. The diff-reviewer
  returned `fix` on f204740 (stale write queue, wrong plain-view coordinates) and on a9059dd (a
  failed write retried without limit); the conductor accepted fca3cc2 without a third review,
  because the fix was narrow and its test counts write attempts directly. The store was checked
  before v7: no choice carried a v6-era focus. First pass: 111 choices (86 keep, 19 pass, 6 maybe).

## Requested additions spot-check (2026-09-26)

Housekeeping before the spot-check: `commons:33882251` (the HABS 9:30 Club interior, recovered
this pass with a firm 1990/circa date) removed from `research-b/drops.json`, since it is now a
live candidate in `requests/`. `dolls/commons-103914289.json`'s license evidence, which quoted
`prop=imageinfo extmetadata`, replaced with Commons structured data fetched raw
(`action=wbgetentities&ids=M103914289`): statement P275 resolves to Q19068220 ("Creative Commons
Attribution-ShareAlike 2.0 Generic", confirmed via a second `wbgetentities` call against
wikidata.org), matching the record's existing `CC-BY-SA-2.0` license; P571 (inception,
`+2010-06-01T00:00:00Z`) matches the record's year 2010 exact; qualifier P2093 on the P275
statement names Chris Whippet, matching the record's photographer field.

**Sample.** Pool: the 20 new records across `requests/` (11), `dolls/` (7), and `zines/` (2), all
`commons:` source ids. Mandatory set: every public-domain record, which is the one HABS interior
(`commons:33882251`, `LicenseRef-public-domain`); no flyer or zine-cover format record is among
the 20 (the zines directory holds photographs of zine-fair scenes, not zine covers). Rest: 19.
Method: sort the rest ascending by the hex `sha256("20260926:" + id)` (`id` in the sheet's
`<source>-<key>` form), then take the first `max(5, ceil(0.2 x 19))` = 5. Seed `20260926`, pool
20, sample size 6: `commons-33882251`, `commons-3736388`, `commons-115584322`, `commons-8757805`,
`commons-25532803`, `commons-93206778`.

**Channels.** Each Commons file's `action=raw` wikitext, fetched under the api lock (`api.php`
class of endpoint, 0.25s serial spacing). Laundering and deletion: a batched
`action=query&titles=Commons:Deletion_requests/File:<title>` per sampled file, all `missing`. The
three Flickr-origin accounts among the sample (Yves Lorson `29572161@N00`, Masao Nakagami
`45489981@N00`, and `spector1`/Raj Gupta) checked against
`Commons:Questionable_Flickr_images/Users`, raw wikitext, none present.

**Tally.** 6 match, 0 mismatch, 0 changed-since-grant, 0 unverifiable.

**Matches**, all dated 2026-09-26:

- `commons-33882251`: wikitext `|permission = {{PD-USGov-NPS}}`, `|source =
  https://www.loc.gov/pictures/item/dc0636.photos.030314p`. Matches the record's
  `LicenseRef-public-domain` / `us-government-work` basis and citation; no delete, copyvio,
  no-permission, or no-source template, no open deletion request.
- `commons-3736388` (The Cramps, Belgium): `{{cc-by-2.0}}` plus
  `{{User:FlickreviewR/reviewed-pass|Yves Lorson on permanent vacation|...|06:16, 19 March 2008
  (UTC)|cc-by-2.0}}`. Matches the record's `CC-BY-2.0` and quoted review line.
- `commons-115584322` (Lux Interior and Poison Ivy, Tokyo): `{{cc-by-sa-2.0}}` plus
  `{{FlickreviewR|status=passed|author=masao nakagami|...|reviewdate=2022-02-26
  21:04:35|reviewlicense=cc-by-sa-2.0|reviewer=FlickreviewR 2}}`. Matches the record's
  `CC-BY-SA-2.0` and quoted review line.
- `commons-8757805` (Dr Madd Vibe and Fishbone): `{{cc-by-sa-2.0}}{{flickrreview|Bidgee|2009-12-16}}`.
  Matches the record's `CC-BY-SA-2.0` and quoted old-style review template.
- `commons-25532803` (the clock at 430 King's Road): `|source={{own}}`,
  `|author=[[User:Mark Ahsmann|Mark Ahsmann]]`, `{{self|cc-by-sa-3.0}}`, contemporaneous
  (2011-06-04 photo, matching upload). Matches the record's `CC-BY-SA-3.0`.
- `commons-93206778` (8-Ball Zine Fair): `|source={{own}}`,
  `|author=[[User:Elbowplants|Elbowplants]]`, `{{self|cc-by-sa-4.0}}`, contemporaneous
  (2019-06-22 event and upload). Matches the record's `CC-BY-SA-4.0`.

No record sent the round back for re-verification.

## Requested additions (2026-09-25)

Geoff's mid-sitting requests added 20 candidates to round 1 (131 total): Fishbone 3, Bad Manners 4,
the 9:30 Club HABS interior (recovered, circa 1990), Skip Groff, Tim Armstrong 2, the Cramps 5,
430 King's Road 2, zine culture 2. Not available under a shippable license: the New York Dolls'
original run, Government Issue, Operation Ivy, the Transplants, punk-era Georgetown, the Yesterday &
Today storefront, and zine covers generally (three zine flatlays dropped: the covers are the subject
and are third-party works). These are outreach targets. Conductor ruling: further photo requests
this round queue for round two, because each search cost 0.2M to 1.0M tokens and the lane is past
its 7.5M ceiling (rough estimate 11M).
- Geoff (2026-09-26): "older is generally better for all of these." Round two's search briefs rank
  by era first. The Feelies: Commons holds only three 2016 to 2019 reunion-era frames (Central Park
  2016, Rough Trade Brooklyn 2019, a 2808x547 panorama); queued for round two as an outreach target
  for 1976 to 1991 photos.
- Yellowman: Commons holds only 2007 frames under the size floor and 2018 to 2022 stage shots (CC BY
  3.0); nothing from the 1980s dancehall peak. King Tubby: no freely licensed photo on Commons at all.
  Both are round-two outreach targets (Beth Lesser's reggae archive is the natural first letter).
- Nightmares on Wax: Commons holds two 2006 DJ E.A.S.E. frames at 600x800, under the floor. Geoff
  (2026-09-26): "Newer stuff is OK for Nightmares on Wax," an exception to older-is-better. Round
  two searches Flickr CC BY/BY-SA festival shots for it (no Flickr API key on the workstation, so a
  search agent, not a quick check).

## Artist sweep (2026-09-26)

Geoff asked for "a round of direct artist searches" across the whole library's range. The list is
`artist-sweep-list.md` (about 290 artists and scenes, including eleven adjacent scenes Geoff
approved and added to `vocab.json` at `5c54f6a` and `f8641cb`). A scripted Commons metadata sweep
(two passes, one API request per second after Commons rate-limited the first attempt, 0 skipped)
found candidates for 223 entries and none for 52; a judgment prune removed name collisions (the Gang
of Four, Suicide, Television, the Seattle SuperSonics for the Sonics) and kept 408 of 765. Geoff
chose to verify the top 30 tonight (period photos and his named artists first; the lane cap stays
near 13M); the other 30 of the verify set wait for round two. Candidate AI-retouched derivatives are
refused in favour of the original file.

Outreach targets (no shippable Commons photo): early Devo (only 2003 to 2010 shots), the Mystic
Knights of the Oingo Boingo, Danny Barnes and the Bad Livers, Soul Coughing as a band (only Mike
Doughty solo), the DC go-go scene beyond Chuck Brown, E.U., Junkyard Band, and Little Benny (Trouble
Funk, Hot Cold Sweat, Redds and the Boys, the Go-Go Museum, Don't Mute DC, Moechella), Augustus Pablo,
U-Roy, Jah Shaka, Justin Hinds, Tenor Saw, Dennis Bovell, Louisa Mark, the Beat, ? and the
Mysterians, the Shaggs, the Lyres, the Embarrassment, the rebetiko singers, Lee Quiñones, Spoonie
Gee, the Treacherous Three, the Rock Steady Crew, Blaze Foley, the mento artists, DNA, Goldie, LTJ
Bukem, Heavens to Betsy, Smith & Mighty, X-Ray Spex, the Plugz, and the Chantays.

## Artist sweep additions spot-check (2026-09-26)

Reader: the conductor, inline. Two new candidate directories: `photos/.candidates/sweep-a/`
(3 records) and `photos/.candidates/sweep-b/` (10 records), the first tranche of the artist
sweep's 30-record verify batch.

**Sample.** Pool: 13 records. Mandatory set: every public-domain record — `commons:164908056`
(the Kingsmen, `us-pre-1978-no-notice`), `commons:17282573` (Thelonious Monk,
`rights-holder-dedication`), and `commons:93080929` (Bill Monroe with Jimmy Carter,
`us-government-work`) — 3 records. No flyer or zine-cover format record is in this tranche. Rest:
10. Method: sort the rest ascending by the hex `sha256("20260926b:" + id)` (`id` in the sheet's
`<source>-<key>` form), then take the first `max(5, ceil(0.2 x 10))` = 5. Seed `20260926b`, pool
13, sample size 8: `commons-164908056`, `commons-17282573`, `commons-93080929` (mandatory), plus
`commons-39320066`, `commons-2214791`, `commons-79670861`, `commons-65425038`, `commons-34602246`
(drawn).

**Channels.** Raw channel only, reusing each candidate's saved `raw/` evidence
(`photos/.candidates/sweep-a/raw/`, `photos/.candidates/sweep-b/raw/`) where a fresh fetch would
duplicate it, plus independent fetches this pass added: a batched
`Commons:Deletion_requests/File:<title>` check per sampled file (one request per title, under
the API lock with the 0.25s serial spacing) and a `Commons:Questionable_Flickr_images/Users` raw
check for the two Flickr-origin accounts in the sample (Greg Neate for the Bratmobile photo,
`jonathancharles` for the Kathleen Hanna photo).

**Tally.** 8 match, 0 mismatch, 0 changed-since-grant, 0 unverifiable.

**Matches**, all dated 2026-09-26:

- `commons-164908056` (the Kingsmen, group portrait, 1966): Commons `action=raw` wikitext
  `|permission` field quotes, verbatim, "When *KRLA BEAT* originally published this image, they
  did not include a copyright notice on either the photograph or as part of its caption. The
  publication failed to included a blanket copyright notice on any of its pages," citing
  `https://krlabeat.sakionline.net/issue/5feb66.pdf` (the 5 February 1966 issue). The record's
  `verified` field quotes this line and adds an independent page-render check of the same PDF.
  This is the KRLA Beat no-notice evidence the plan required in `verified`; confirmed present.
- `commons-17282573` (Thelonious Monk, Minton's Playhouse, ca. 1947): license header
  `{{PD-Gottlieb}}`; `Template:PD-Gottlieb` raw quotes "In accordance with the wishes of William
  Gottlieb, the photographs in this collection entered into the public domain on February 16,
  2010." Matches the record's `rights-holder-dedication` basis and quoted fact exactly.
- `commons-93080929` (Bill Monroe with Jimmy Carter, South Lawn, 1980): the page's bare
  `{{PD-USGov}}` tag carries no authorship prose, so the record instead cites Commons structured
  data. `action=wbgetentities&ids=M93080929` confirms statement P6216 (copyright status) =
  Q19652 "public domain," qualified by P459 (determination method) = Q60671452 "work of the
  federal government of the United States" and P1001 (jurisdiction) = Q30 "United States"; P180
  (depicts) = Q170042 "Bill Monroe." Matches the record's `us-government-work` basis and quoted
  fact.
- `commons-39320066` (Kathleen Hanna, Bikini Kill, Olympia Mayday 1991): `{{flickrreview|ww2censor|17:57,
  1 April 2015 (UTC)}}` plus `{{cc-by-sa-2.0}}`. Matches the record's `CC-BY-SA-2.0`.
- `commons-2214791` (Bo Diddley, Long Beach Blues Festival, 1997): `{{self|GFDL|Cc-by-sa-3.0-migrated|Cc-by-sa-2.5,2.0,1.0}}`,
  own-work by Masahiro Sumori. Matches the record's `CC-BY-SA-3.0` (the migrated tag).
- `commons-79670861` (Clifton Chenier, Le Palace, Paris, 1977): `|source={{own}}`,
  `|author=[[User:Lioneldecoster|Lionel Decoster]]`, `{{self|cc-by-sa-4.0}}`. Matches the
  record's `CC-BY-SA-4.0`.
- `commons-65425038` (Art Blakey and the Jazz Messengers, Concertgebouw, 1959): Nationaal
  Archief / Anefo, `{{CC-zero}}`, credited "Herbert Behrens / Anefo." Matches the record's
  `CC0-1.0`.
- `commons-34602246` (Bratmobile, The Charlotte, Leicester, 1994):
  `{{User:FlickreviewR/reviewed-pass|neate photos|...|2014-08-08 08:53:32|cc-by-2.0|}}{{cc-by-2.0}}`.
  Matches the record's `CC-BY-2.0`.

**Laundering and deletion.** A batched `Commons:Deletion_requests/File:<title>` query for all 8
sampled files returned `missing` for every one: no open deletion request. The two Flickr-origin
accounts in the sample (Greg Neate, `jonathancharles`) are both absent from
`Commons:Questionable_Flickr_images/Users`.

No mismatch. The round does not go back for re-verification on this tranche.
- Artist sweep verification (top 30): 13 kept and added to round 1 (144 total); 17 dropped, mostly
  US publicity and trade-ad photos with a bare no-notice tag (the 1957 Johnny Cash Cash Box ad's own
  masthead carries a notice), plus name collisions (a Madrid rock band for Junkyard Band, an MC5
  mural). Calypso photos on a Trinidad and Tobago term (`{{PD-TT-photo}}`) are a round-two research
  item: whether a 1937 Trinidad photo is also public domain in the US.

## Round close (2026-09-26)

Round 1 closed with 144 candidates sorted: 117 keep, 22 pass, 5 maybe. Committed at `e4754e8`: 116
picked pairs (each .webp the exact preview Geoff judged; 25 records carry his focus or fit), 16
waivers (13 artist, 3 landmark), and `rejected.json` with 326 entries (22 passes plus the agents'
non-duplicate drops). `check.mjs integrity` passes on the committed pool; `check.mjs` validates
the new `focus` and `fit` fields since `779d3e7` (diff-reviewer accept).

Returned unpicked: `commons:139197098` (Run-DMC, Long Beach Arena). The page's name-based
second-image flag missed it ("Run-DMC" against "Run-D.M.C."); the close's MusicBrainz step found all
three frames share one MBID. The other two are committed with a waiver; this one stays in
`.candidates/` for Geoff's call. The five maybes also stay in `.candidates/`. Round two should key
the page's second-image flag on MBID where the sheet has one.

MBIDs: 55 of 59 artist names resolved; Chitra Gupt, Deedar Singh Pardesi, Les Ambassadeurs, and
Poison Ivy stay name-keyed.

Balance: every family share is below target (116 photos against a 300-photo aim), which blocks the
done checkpoint, not this commit.

Round two targets: the 30 verify-set candidates not yet checked (`sweep/verify-set.json`); the
outreach list above (early Devo, the Mystic Knights, the Bad Livers, go-go, the New York Dolls,
Government Issue, Op Ivy, the Feelies, Yellowman, King Tubby, and the sweep's 52 empties); the
Trinidad and Tobago public-domain question for 1930s calypso; Nightmares on Wax on Flickr; 1960s
garage and the 1980s revival. Search briefs rank by era first.

Spend: roughly 13M against a 7.5M ceiling raised mid-round. The page rebuilds (review mode, crop
controls, three review cycles) and the mid-sitting searches drove the overrun; a later round should
budget the sorting page as its own task and batch owner requests into the next round's search.
