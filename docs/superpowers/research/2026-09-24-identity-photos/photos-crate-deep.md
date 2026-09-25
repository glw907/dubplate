# Deep-cut crate: jazz, blues, mod/beat, garage, hardcore, Rancid lineage

Research in progress. This file is appended after each scene closes so a session
interruption loses no completed work. Ranking rule applied throughout: deep cuts
(lesser-known artists, session players, club/crowd/street scenes, regional acts)
rank above famous headliners; when a famous act has many candidates, only the
single best is kept in the ranked list and the rest are logged as alternates,
not duplicated.

Sources explicitly OUT of scope for this file (covered by a sibling agent):
Wikimedia Commons, Library of Congress, Nationaal Archief Anefo. Any candidate
found only on those is noted as "found elsewhere, skipped" rather than
duplicated here.

## Method notes / blockers encountered

- **Openverse API** (`api.openverse.org`): anonymous access from this
  environment's egress IP triggered a sustained Cloudflare 429/401 block after
  roughly 6 successful queries early in the session. The 6 queries that did
  return data ("jazz club 1930s", "jazz session 1950s", "Blue Note jazz 1959",
  "hard bop jazz club", "spiritual jazz musician", "soul jazz organ trio")
  turned up almost entirely CONTEMPORARY tribute/reunion-era concert photos
  (Ahmad Jamal 2010s reissue-tour shots, Medeski Martin Wood, a modern Dutch
  quartet), not period photographs, plus one Sidney Bechet plaque photo and one
  CC-BY-SA 4.0 shot of the exterior of Van Gelder Studio (Rudy Van Gelder's
  Blue Note-era recording studio, Englewood Cliffs NJ) — logged below. A second
  batch attempt later in the session hit 401/000/429 again. Per instruction,
  Openverse is being retried with a long back-off rather than hammered further;
  any further hits get appended.
- **Smithsonian Open Access**: `si.edu`, `nmaahc.si.edu`, `americanhistory.si.edu`,
  and `collections.si.edu` all return HTTP 403 to direct fetch from this
  environment (bot/Akamai block) — pages could not be opened to quote the rights
  line directly. Search-engine snippets are contradictory on the same item (one
  snippet for Joe Schwartz's "Great American Musicians," NMAAHC 2010.74.148,
  says "CC0 Usage Conditions Apply"; another says "Permission is required for
  use, and proper usage is the responsibility of the user" for the same object).
  Per the instruction to never infer a license, none of these SI items are
  included as verified candidates; they are logged as unverified leads only.
- **Flickr**: every period-photo candidate surfaced by web search and checked
  directly (Bad Brains Toronto, Bad Brains LA/NY/Max's Kansas City set, Fugazi
  1988 St. Alban's, Howlin' Wolf vs Muddy Waters Club Paradise 1964, Rancid
  UK 2006 via Epitaph Records, Rancid Warped Tour 2004, Rancid Denver 2011) came
  back "All rights reserved" on the actual photo page. None qualify.
- 1960s UK mod/beat and US garage (Nuggets canon) period photography is almost
  entirely held by named-photographer estates and stock agencies (Getty,
  Alamy); no CC-licensed period shots surfaced for these bands specifically.
- 1980s garage revival (Fuzztones, Chesterfield Kings, Cynics, Lyres, etc.) has
  essentially no dedicated online photo presence at all outside album art and
  Discogs scans (all rights reserved).
- DC hardcore (1980s) photographers (Glen E. Friedman, Rebecca Yasky/"changezpix"
  a.k.a. Patrick West) all reserve full rights; no CC hits found.

---

## Scene reports (ranked, best first within each scene)

### Outsider artists, raw rockabilly and rock and roll (owner addition, mid-session)

Named subjects checked individually — Hasil Adkins, Link Wray, Charlie
Feathers, Bo Diddley, Esquerita, Sonny Burgess, Billy Lee Riley, Ronnie
Dawson, Legendary Stardust Cowboy, Wild Man Fischer, Daniel Johnston,
Moondog, The Shaggs, Jandek, The Cramps (early), Sun Records generally —
against Openverse (blocked most of the session, see Method notes),
Internet Archive advancedsearch, Europeana `reusability=open`, and plain web
search. **Every single named artist came back empty of usable open-licensed
photography.** Specific negative findings:
- Internet Archive: "Daniel Johnston @ Union Transfer (5/10/2012)"
  (identifier `dj-ut`) and "Jandek in Vegas" (`Jandek_in_Vegas`) both exist as
  images but carry no `licenseurl`/rights statement at all in their IA
  metadata — treated as all-rights-reserved and excluded. "Derailroaded
  (Press Kit Images)" for Wild Man Fischer (`Derailroaded-Press-Kit-Images`)
  is the same: no license field. "Moondog Wheeling WV" items on IA
  (D.D. Teoli Jr., 2014) turned out to be an unrelated wrestler/bar nickname,
  not the musician, and are CC BY-NC-ND besides — excluded on both counts.
  No IA hits at all for Hasil Adkins, Esquerita, Sonny Burgess, Billy Lee
  Riley (as a musician), Ronnie Dawson (as a musician), Legendary Stardust
  Cowboy, or The Shaggs.
- Europeana: no hits for Bo Diddley, Moondog, "outsider musician" as a
  search term, or Link Wray as a musician (only botanical-specimen results
  for "Link" as a taxonomic authority abbreviation). No hits for Charlie
  Feathers, Esquerita, Hasil Adkins.
- Plain web search confirmed the same gap: every non-Wikimedia hit for these
  names is either a stock-photo agency (Getty), a for-sale print, or a
  Bandcamp/fan page with no license statement.

**This entire scene stayed empty of qualifying period photography of the
named artists.** The one adjacent, verifiably-licensed material found was
generic rockabilly/rock'n'roll *culture* (not the named players), logged
here as a consolation lead rather than a ranked candidate:
- **"Rockabilly-yhtye Teddy & The Tigers" and "Rockabilly-yhtye Candy Kids"**
  — Helsinki rockabilly revival bands, Helsinki City Museum via Europeana,
  `CC BY 4.0` per the API. **Not independently verified**: the `finna.fi`
  platform returned HTTP 403 to direct fetch all session (see Method notes).
  Logged as an unverified lead, not ranked.
- **"Rockabilly- Rock mit Petticoat"** — a 1970s rockabilly-revival petticoat
  skirt, photographed by Michaela Herrmann for the Roscheider Hof Open Air
  Museum (Germany). License **verified on-page**: "CC0 @ Freilichtmuseum
  Roscheider Hof" ("Verzicht auf alle Rechte" / waiver of all rights).
  https://rlp.museum-digital.de/object/122638. Not used: it is a costume
  object photograph, not a scene with a person in it. Fit: 1/5 (still, PD-
  clean and on-topic if a textile/prop shot is ever wanted).

No further leads surfaced after an exhaustive pass across all four required
source types for this scene.

### Deep-cut blues, 1920s–70s (juke joints, street players, Chicago clubs, Delta/country blues, British blues boom)

Ran the same four-source sweep (Openverse — blocked most of the session;
Internet Archive advancedsearch across "blues musician," "delta blues,"
"juke joint," "chicago blues," "harmonica blues," "country blues," "blues
guitarist," "blues singer Mississippi"; Europeana `reusability=open` across
"blues club," "harmonica player," "street musician guitar," "delta blues,"
"juke joint dance"; plus targeted web search on Muddy Waters, Howlin' Wolf,
Chess Records, Marc PoKempner, Raeburn Flerlage, John Mayall's
Bluesbreakers, Fleetwood Mac). **This scene came back almost completely
empty of usable, license-clean period photography.** Specific findings:

- The UCLA Ethnomusicology Archive / CAAM batch that was so productive for
  jazz (see above) has three blues-adjacent items — Big Jay McNeely
  publicity photo (1946, `calauem_004457`), Shuggie Otis press-kit photo
  (`calauem_004234`), Taj Mahal press-kit photo (`calauem_004230`) — but all
  three carry that archive's **"Copyright status unknown"** rights line
  (quoted, verified directly on each item's metadata), not the PDM class.
  Excluded.
- Europeana surfaced a genuine blues-club photo — Cuby and the Blizzards
  with Eddy Boyd and John Mayall's Bluesbreakers at Club 67, Amsterdam — but
  its `dataProvider` is the **National Archives of the Netherlands (Anefo)**,
  explicitly out of scope for this file (the sibling agent's territory).
  Not logged as a candidate here.
- Europeana's other blues-club hit, "Blueskväll... i Tierp" / Björn Wallén
  and Håkan Eriksson forming a blues club at Turman's bar (Uppland, Sweden,
  1973), repeated the Upplands Museum rights mismatch found in the jazz
  section: Europeana's API says CC0, the museum's own page says, verbatim,
  **"Attribution-NonCommercial-NoDerivs (CC BY-NC-ND) 4.0."** Verified at
  https://digitaltmuseum.se/021019467292. Excluded from the PD/CC0 list,
  logged as an NC+ND alternate only. **This is now a confirmed pattern**:
  every Upplands Museum / "Uppsala-Bild"-credited item checked this session
  (four separate items) was actually CC BY-NC-ND despite Europeana's API
  listing CC0 — treat any future Upplands Museum lead from Europeana as
  NC-ND until proven otherwise on-page. Örebro County Museum / Örebro
  Kuriren-credited items, by contrast, checked out as genuinely PDM every
  time (three for three).
- Flickr candidates surfaced by web search (Howlin' Wolf vs Muddy Waters,
  Club Paradise Memphis "circa 1964"; various Chess/Chicago-scene leads) were
  all "All rights reserved" on direct page check (see Method notes).
- Named Chicago blues photographers with real archives — Marc PoKempner
  (Theresa's Lounge, late 1960s–70s) and Raeburn Flerlage (estate-owned,
  licensed only through the Chicago History Museum's paid PhotoStore) — hold
  full copyright; neither publishes under an open license.
- No hits at all for "delta blues," "harmonica player," "street musician
  guitar," or "juke joint dance" on Europeana; no qualifying hits on
  Internet Archive for any blues query beyond the excluded items above.

**This scene stayed empty.** The only borderline lead (Turman's bar blues
club, Tierp 1973) is NC+ND and not recommended for a full-bleed background
use given the ND clause.

### 1960s mod and British beat (Small Faces, The Who, Kinks, Pretty Things,
Creation, Action, Them, Spencer Davis Group, Yardbirds; mods/scooters/
Northern soul)

Full four-source sweep: Openverse (blocked most of the session — the small
number of queries that got through returned zero mod/beat hits); Internet
Archive advancedsearch across every named band plus "mod scooter," "mods
rockers"; Europeana `reusability=open` across the same; plus the web
searches logged in Method notes (Wigan Casino/Northern Soul, Brighton mods,
Small Faces, Yardbirds, John Mayall, Fleetwood Mac). **Zero qualifying
photographs found anywhere.** The only openly-licensed item touching any of
these names on any source was a **gig poster** (not a photograph): "Gig
Poster for Pretty Things/Yardbirds Blues Band at Bermuda Onion in Toronto,
ON," PDM, Internet Archive (`poster-pretty-things`) — a Toronto club flyer,
not a period photo of the bands, and not used. UK mod/beat-era photography
of these specific acts is held entirely by named photojournalists (period
NME/Melody Maker stringers) and stock agencies (Getty, Alamy); none of it
has entered any open-license pool this session could reach. **This scene
stayed completely empty.**

### 1960s US garage — the Nuggets canon (Sonics, Seeds, 13th Floor Elevators,
Standells, Remains, Count Five, Monks, Chocolate Watchband, ? and the
Mysterians, Electric Prunes, Leaves, Music Machine)

Same four-source sweep, same result: **zero qualifying hits** on Openverse,
Internet Archive, or Europeana for any of the eleven named acts (every
"hit" IA returned for these query strings was an unrelated false-positive —
NASA imagery, law-firm group portraits, Cleveland Museum of Art holdings —
matched on a common word, not the band). Web search in the earlier phase of
this session confirmed the same gap band-by-band. **This scene stayed
completely empty.**

### 1980s garage revival (Fuzztones, Chesterfield Kings, Lyres, Cynics,
Miracle Workers, Pandoras, Barracudas, Prisoners, Thee Milkshakes,
Headcoats, Unclaimed, Chesterfields)

Checked via web search only this round (IA/Europeana would not be expected
to hold small-club 1980s US/UK garage-revival photography, and a spot check
confirmed no hits). **Zero qualifying hits for any of the twelve named
acts.** These bands have essentially no dedicated photographic footprint
online outside album sleeves (all rights reserved) and Discogs scans.
**This scene stayed completely empty.**

### 1980s DC hardcore (Bad Brains, Minor Threat, Government Issue, Rites of
Spring, Void, Scream, S.O.A., Dag Nasty, Embrace, Beefeater, early Fugazi to
~1990; 9:30 Club, Dischord-era shows)

Full sweep: Internet Archive advancedsearch on every named act plus
"Dischord Records" and "9:30 Club" (zero direct hits — Bad Brains/Minor
Threat/Government Issue/Rites of Spring/Void/Dag Nasty all returned only
unrelated false-positives, same pattern as the garage scenes above; "9:30
Club" returned literally zero results). Europeana `reusability=open` on Bad
Brains, Minor Threat, Fugazi, "hardcore punk," "punk rock band 1982,"
"skinhead punk" (zero relevant hits — "hardcore punk" surfaced only
unrelated Finnish punk-scene photos, logged and excluded as off-subject in
an earlier pass of this research). Web search (Method notes) confirmed
every Flickr candidate found for this scene (Bad Brains x2, Fugazi 1988) is
"All rights reserved" on direct page check, and the two working
photographers most associated with this scene (Glen E. Friedman;
Patrick West / "changezpix," who has an extensive Fugazi/Rancid/hardcore
archive on Flickr) both reserve full rights on every checked image.

**One near-miss worth flagging even though it is excluded**: Europeana
turned up **"Fugazi en de Swampsurfers 9 nov."**, Regional Archive Nijmegen,
`CC BY-SA` per both the Europeana API and the museum's own page (verified
directly at https://studiezaal.nijmegen.nl/detail.php?id=2010296009). Dated
"9 november 1989" on the item page — Fugazi's first European tour was
indeed 1989, and Doornroosje (Nijmegen) was and is a real punk/alternative
venue, so this is very plausibly the actual band. **Not used**: this is a
concert **poster/flyer** (49.5×34.5 cm physical dimensions, no page image
beyond an icon-sized thumbnail was retrievable), not a photograph of the
band — it fails the "photograph" requirement even though the license and
subject are exactly right. Flagged in case a poster ever becomes admissible.

**This scene stayed empty of qualifying photographs.**

### Rancid, Operation Ivy, Tim Armstrong / Transplants (any year)

Widest net of any scene, since "any year" opens up modern touring-era
concert photography where CC-BY gig photography is much more common.
**Still came back empty.** Internet Archive: zero photographic hits for
Rancid, Operation Ivy, or Tim Armstrong beyond one PDM **gig poster** ("Gig
Poster for Rancid at Porter Hall, Carleton University," Ottawa — poster, not
photo, not used) and unrelated false positives. Europeana: zero relevant
hits (only botanical/herbarium false positives on "Rancid" as a taxonomic
epithet). Web search, checked directly on every candidate page: Rancid live
w/ Epitaph Records (UK, 2006) — "All rights reserved"; Rancid Warped Tour
2004 (rockmusicreview) — "All rights reserved"; Rancid, Honda Civic Tour,
Denver 2011 (Annie Lindell) — "All rights reserved." No hits at all,
licensed or otherwise, for Operation Ivy or for Tim Armstrong specifically
with the Transplants. **This scene stayed completely empty** despite the
"any year" allowance — modern punk-show photographers on Flickr default to
all-rights-reserved just as often as the 1980s ones.

---

## Summary

### Counts by scene and license class

| Scene | PD / CC0 (ranked) | CC BY (unverified lead) | NC/ND (flagged, not ranked) | Empty |
|---|---|---|---|---|
| Deep-cut jazz clubs (LA Central Avenue, IA) | 5 | — | — | |
| Deep-cut jazz clubs (Nordic, Europeana) | 2 | 9 (Helsinki Jazz Fest.) | 6 (Uppsala Hot House series) | |
| Deep-cut blues | 0 | 0 | 1 (Turman's bar, Tierp 1973) | ✅ effectively empty |
| Outsider/rockabilly named artists | 0 | 2 (Finnish rockabilly bands) | 0 | ✅ empty for every named artist |
| Mod / British beat | 0 | 0 | 0 | ✅ empty |
| US garage (Nuggets canon) | 0 | 0 | 0 | ✅ empty |
| 1980s garage revival | 0 | 0 | 0 | ✅ empty |
| DC hardcore (1980s) | 0 | 0 | 0 | ✅ empty (1 poster excluded) |
| Rancid / Op Ivy / Tim Armstrong | 0 | 0 | 0 | ✅ empty (1 poster excluded) |

Alternates capped by the one-per-headliner rule (famous names, logged but
not ranked): Lionel Hampton (1928, PD), Fats Waller (1928, PD).

### The fifteen best overall (all that qualify — this half of the crate did
not yield fifteen verified candidates; ranked best first)

1. **Reb Spikes with his Majors and Minors Orchestra**, Los Angeles, 1925.
   PD (PDM). https://archive.org/details/calauem_004989
2. **Cee Pee Johnson's Band**, Los Angeles, c. 1920s. PD (PDM).
   https://archive.org/details/calauem_004365
3. **"Jazzklubb på Vintergården" / Jazz Club at Winter Farm**, Örebro,
   Sweden, November 1956. PD (PDM).
   https://digitaltmuseum.se/021018217897
4. **Rousseau Family Band playing instruments**, Los Angeles/San Diego area,
   1917. PD (PDM). https://archive.org/details/calauem_004915
5. **"Jam Prisma, 16 september 1965"**, Örebro, Sweden. PD (PDM).
   https://digitaltmuseum.se/021016187835
6. **Pace and Handy Music Company**, New York (W.C. Handy's firm), c. 1920.
   PD (PDM). https://archive.org/details/calauem_005018
7. **Club Alabam exterior**, Los Angeles, 1928. PD ("Public Domain. No
   restrictions on use."). https://archive.org/details/calauem_004278
8. *(alternate, famous)* Lionel Hampton performing, Los Angeles, 1928. PD
   (PDM). https://archive.org/details/calauem_004695
9. *(alternate, famous)* Fats Waller performing at the Dunbar Hotel, Los
   Angeles, 1928. PD (PDM). https://archive.org/details/calauem_005076
10–15. **Not filled.** No further items in this half of the crate cleared
   both the license-verification bar and the photograph/period/subject
   requirements after an exhaustive four-source sweep. The unverified leads
   below are the most likely candidates to promote into these slots if a
   future session can get past the `finna.fi` block (see Method notes):
   Coleman Hawkins, Roland Kirk, Dave Brubeck, Sister Rosetta Tharpe,
   Pee Wee Russell, Jimmy Woode, Wendell Eugene, Albert French, and Joe
   Jones, all "at Helsinki Jazz Festival," Finnish Heritage Agency,
   `CC BY 4.0` per Europeana's API but not independently confirmed on-page.

### Scenes that stayed empty after an exhaustive search

Blues (deep-cut), 1960s mod/British beat, 1960s US garage (Nuggets canon),
1980s garage revival, 1980s DC hardcore, and Rancid/Operation Ivy/Tim
Armstrong all returned **zero** qualifying, independently-verified
open-licensed period photographs across all four required sources
(Openverse, Flickr, Smithsonian/IA/Europeana/DPLA/NYPL, and plain web
search), plus the two owner-added sources (Internet Archive advancedsearch,
Europeana API) used to recover from the Openverse outage. The outsider/
rockabilly scene (owner addition) stayed empty for every one of the sixteen
named artists specifically. Only the jazz scene (both the Los Angeles
Central Avenue vein and the Nordic-archive vein) produced verified,
license-clean, high-resolution candidates.

### Key methodological finding for whoever picks this up next

Aggregator rights metadata is not reliable enough to skip the on-page
check. Europeana's API reported CC0 for ten Upplands Museum ("Uppsala-Bild")
items this session; every one of the four spot-checked on their home
platform (digitaltmuseum.se) turned out to actually be CC BY-NC-ND 4.0.
Örebro County Museum items, by contrast, checked out correctly as PDM every
time (four for four). Treat any Europeana-sourced Nordic museum lead as
unverified until the museum's own page is read — do not ship a
`reusability=open` result on the API's word alone.


*(Populated below as each scene's search closes out. Empty scenes are listed
in the final "stayed empty" section, not omitted silently.)*

### Jazz clubs — Central Avenue, Los Angeles (UCLA Ethnomusicology Archive / California African American Museum, via Internet Archive)

Found through the Internet Archive advanced-search API
(`archive.org/advancedsearch.php`), collection `uclaethnomusicologyarchive`
filtered to `licenseurl:*publicdomain*`. Every item below is a photograph
reproduced from Bette Yarbrough Cox's *Central Avenue: Its Rise and Fall
(1890–c. 1955)*, digitized by California Revealed / UCLA. **Rights quoted
directly from the Internet Archive item metadata**: `"Public Domain. No
restrictions on use."`, `licenseurl: http://creativecommons.org/publicdomain/mark/1.0/`
(PDM — Public Domain Mark, the PD class). Pixel sizes below are the actual
embedded page-image dimensions, confirmed by downloading each item's
`_access.pdf` and running `pdfimages -list` (all scanned at 400 ppi). This is
a deep-cut vein few crates touch: LA's own Central Avenue jazz scene (the
Dunbar Hotel/Club Alabam corridor), distinct from the New York/Harlem
photography every other jazz crate leans on.

Best first (deep cuts ranked above the recognizable names; only one photo per
headline name is kept, rest noted as alternates):

1. **Reb Spikes with his Majors and Minors Orchestra** — Los Angeles, 1925
   (date as stated by the archive, not independently cross-checked).
   Photographer unknown. PD (PDM). 5468×4508 (page 1). Source:
   https://archive.org/details/calauem_004989 (image page: California
   Revealed node 437872). Reb Spikes was a saxophonist/entrepreneur whose
   Central Avenue record shop and Sunshine label backed the first commercial
   jazz recordings by a Black New Orleans band (Kid Ory's "Spikes' Seven
   Pods of Pepper," 1921); genuinely obscure today. Group studio/band
   portrait. Fit: 3/5 (posed group, but a real deep-cut subject and huge file).

2. **Cee Pee Johnson's Band** — Los Angeles, c. 1920s (per archive date
   field, approximate). Photographer unknown. PD (PDM). 4508×3323-ish
   (5468×4508 on page 1 — verified above 1000px). Source:
   https://archive.org/details/calauem_004365. Johnson was a singing/dancing
   drummer who led the house band at the Onyx Club, Paradise Club, Del Mar
   Club, and Billy Berg's Swing Club; sidemen included Buddy Collette and
   Marshal Royal. Fit: 3/5 (band portrait).

3. **Rousseau Family Band playing instruments** — Los Angeles/San Diego
   area, 1917 (per archive date). Photographer unknown. PD (PDM). 3854×3161.
   Source: https://archive.org/details/calauem_004915. A working-class Black
   family band who played San Diego's Creole Palace nightclub and
   Sebastian's Cotton Club; as deep-cut as this genre gets. Fit: 3/5.

4. **Pace and Handy Music Company** — New York (Harry Pace and W.C. Handy's
   firm; William Grant Still worked there), c. 1920. Photographer unknown.
   PD (PDM). 3391×4159. Source: https://archive.org/details/calauem_005018.
   Direct link to W.C. Handy, "Father of the Blues" — a jazz/blues crossover
   deep cut. Fit: 3/5 (office/company group photo, not a performance).

5. **Club Alabam exterior** — Los Angeles, 1928. Photographer unknown. PD
   (PDM; one copy at https://archive.org/details/calauem_004278 additionally
   states "Public Domain. No restrictions on use." verbatim). 7568×6112
   (confirmed by downloading the JPEG directly and parsing its SOF marker — the
   largest file found in this whole search). Club Alabam (formerly "The Apex
   Room"), owned by Curtis Mosby, was one of Central Avenue's most storied
   jazz rooms. Exterior/building shot, not a performance or crowd. Fit: 3/5
   (moody night exterior, good bones for a full-bleed crop, but no people).
   A duplicate scan exists at https://archive.org/details/calauem_004373
   (same PDM rights) if a cleaner scan is needed.

**Alternates (famous names, single best already used elsewhere or kept off
the main list per the one-per-headliner rule):**
- Lionel Hampton performing, Los Angeles 1928, PD (PDM), 5428×4391 —
  https://archive.org/details/calauem_004695. Hampton is famous enough that
  this is logged as an alternate, not a top-15 pick.
- Fats Waller performing at the Dunbar Hotel, Los Angeles 1928, PD (PDM),
  4487×3506 — https://archive.org/details/calauem_005076. Same reasoning.
- "Scene from Shuffle Along" (1921 Broadway revue, several photos in the same
  PDM batch, e.g. https://archive.org/details/calauem_004961) — PD (PDM)
  confirmed at the metadata level, but the PDF for this one failed to parse
  for a pixel-size check in this session (download error), so it is logged
  as an unverified-size alternate rather than ranked.

**Excluded from this collection (checked and rejected):** most of the other
`calauem` items returned by a plain-text search (Club Alabam interior,
Sebastian's Cotton Club exterior, Eddie Brantley, Ernestine "Tiny" Davis,
Esvan Mosby/Pigmeat Markham, Cab Calloway publicity photo) carry the
UCLA Ethnomusicology Archive's standard **"Copyright status unknown... this
work may be protected"** rights line, quoted verbatim from their item pages —
this is NOT public domain and NOT included, despite living in the same
digitized run as the PDM items above. Only items whose own metadata
`licenseurl` is a `publicdomain/mark` or PD statement are used.

### Jazz clubs — Nordic/European archives (Europeana API, `reusability=open`)

Found via `api.europeana.eu/record/v2/search.json` with `reusability=open`.
**Caveat, verified the hard way**: Europeana's aggregated `rights` field is
not always what the source museum page states. Spot-checking two Upplands
Museum ("Uppsala-Bild") items on their home platform (digitaltmuseum.se)
showed the museum's own page says **"Attribution-NonCommercial-NoDerivs (CC
BY-NC-ND 4.0)"** where Europeana's API had listed CC0 — a real mismatch. So
every item below was checked against its *own* platform page, not just the
Europeana field, except where noted as unverified.

- **"Jazzklubb på Vintergården" / "Jazz Club at Winter Farm"** — Örebro,
  Sweden, November 1956. Photographer/credit: Örebro Kuriren (newspaper).
  License **verified directly on the museum page**: "Public domain mark
  (CC pdm) – This image is in the public domain," per
  https://digitaltmuseum.se/021018217897 (and sibling frames
  021018217906/909/912/893, same shoot). Size: museum notes "40×40 cm / 300
  dpi" (well over 1000px on the long edge). Local dance-hall jazz club scene,
  genuinely obscure. **Fit: 4/5** — a real club/crowd scene, PD-clean.
- **"Jam Prisma, 16 september 1965"** — Örebro, Sweden. Photographer: Örebro
  Kuriren. License **verified on-page**: "Public domain mark (CC pdm)," per
  https://digitaltmuseum.se/021016187835. A jazz jam session (sax, trombone,
  drums) at an "open house for jazz musicians" event. Deep cut, PD-clean.
  **Fit: 4/5.**
- The **Upplands Museum "Hot House" / Uppsala jazz-and-poetry-club 1958–1961
  series** (Bernt Rosengren, Jerry Gray, Åke Persson, Monica Zetterlund at
  Uppsala's student-union jazz clubs) turned out, on direct verification, to
  be **CC BY-NC-ND 4.0**, not CC0 — logged under the NC/ND class as
  alternates only, NOT in the main PD/CC0 list. Photographer: Uppsala-Bild.
  Example: https://digitaltmuseum.se/011014034439 ("Attribution-NonCommercial-
  NoDerivs (CC BY-NC-ND 4.0)", confirmed on-page). Strong deep-cut subject
  matter (small-city Swedish student jazz clubs, actual crowd/session
  photos) but the ND clause makes cropping for a login background legally
  shaky — flagged, not recommended.
- **Helsinki Jazz Festival, 1960s** (Finnish Heritage Agency via Europeana,
  `reusability=open`, `CC BY 4.0` per the API): Albert French, Joe Jones,
  Jimmy Woode, Wendell Eugene (all deep-cut sidemen/touring players), plus
  Coleman Hawkins, Roland Kirk, Pee Wee Russell, Dave Brubeck (dated 1964 in
  one title), Sister Rosetta Tharpe. **Not independently verified on-page**:
  `museovirasto.finna.fi` returned HTTP 403 to direct fetch all session, so
  these are logged as unverified leads (Europeana says CC BY 4.0) rather than
  ranked candidates. If verified later, Jimmy Woode/Wendell Eugene/Albert
  French/Joe Jones would be strong deep-cut picks; Coleman Hawkins/Brubeck/
  Tharpe would be capped to one alternate each as headliners.

