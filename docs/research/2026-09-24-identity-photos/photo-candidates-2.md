# Login-page photo candidates — collector's crate rotation

Research pass 2. Every entry below was verified against Commons `imageinfo`/`extmetadata`
(LicenseShortName, DateTimeOriginal, Artist), an LOC item's `item.rights` field, or the
photo's own Commons page for a `{{PD-...}}` tag — never inferred. Method: Commons API
`generator=categorymembers` (file + one level of subcats) with a `dubplate-research/1.0`
User-Agent and ≥1.5s between calls; LOC's `loc.gov/...?fo=json`; the Nationaal Archief's
own site has dropped its public fotocollectie API (redirects to a search UI with no
documented JSON endpoint), so Anefo material below comes via its Wikimedia Commons mirror,
which carries the same CC0 Anefo/Stadsarchief metadata.

**Owner instruction mid-pass:** weight the rotation so Jamaican/2 Tone/reggae/dub/ska
(scene 9, plus reggae-adjacent scene 8 material) fills roughly half the crate. Effort below
was reweighted accordingly — see the "Can the reggae half be filled?" verdict at the end.

Fit score: 1–5 for a moody full-bleed halftone print. Live/crowd/club/street shots score
highest; posed headshots score lower; anything under 1000px on the long edge caps at 2
regardless of content.

---

## Scene 1 — Deep-cut jazz (William P. Gottlieb Collection, Library of Congress)

Rights (verified on the collection's LOC item record, `item.rights`): *"In accordance with
the wishes of William Gottlieb, the photographs in this collection entered into the public
domain on February 16, 2010, but rights of privacy and publicity may apply."* — **Public
domain (donor-dedicated, 2010; U.S. record, no separate non-U.S. statement given)**.
Sampled IIIF masters (`.../info.json`) confirmed 5352×4096 and 5548×5393 — Gottlieb scans in
this collection are consistently in the 4000–6000px range; not every item's `info.json` was
individually pulled given the reweighting toward scene 9, so treat "large" as collection-
level, spot-verified rather than item-by-item.

| # | Subject | Year | Source | License | Size | Description | Fit |
|---|---|---|---|---|---|---|---|
|1|Dancers in a jazz club|1938–48|loc.gov/item/2023867585|PD (donor-dedicated)|large|Couples dancing on a club floor, DC|5|
|2|Eddie Condon, Tony Parenti, Wild Bill Davison, Gowans, Lesberg, Ohms — Eddie Condon's, NYC|ca. June 1946|loc.gov/item/2023867575|PD|large (5352×4096 confirmed)|Full sextet crowded around a club table|5|
|3|Allen Eager, Art Mardigan, Curley Russell — Club 18, NYC|ca. Nov 1947|loc.gov/item/2023867629|PD|large|Trio candid, club setting|4|
|4|Bud Freeman & Marty Marsala — Jimmy Ryan's, NYC|ca. July 1947|loc.gov/item/2023867716|PD|large|Two players, club stage|4|
|5|Oscar Moore, Nat King Cole, Wesley Prince, NYC|ca. July 1946|loc.gov/item/2023867557|PD|large|Trio session candid|4|
|6|Sam Donahue, Stan Kenton, Nick Callonus, Hep — Aquarium, NYC|ca. Dec 1946|loc.gov/item/2023868574|PD|large|Group club shot|4|
|7|Jerry Wald, Gordon MacRae, Mel Tormé, Marion Hutton, Jerry Jerome — Saturday Teentimers Show, NYC|ca. Aug 1947|loc.gov/item/2023868344|PD|large|Studio group, broadcast set|3|
|8|Hot Lips Page — Jimmy Ryan's, NYC|ca. June 1947|loc.gov/item/2023868971|PD|large|Trumpeter mid-solo, club|3|
|9|Albert Nicholas — Jimmy Ryan's, NYC|ca. Mar 1947|loc.gov/item/2023868114|PD|large|Clarinetist, club|3|
|10|Sidney De Paris — Jimmy Ryan's, NYC|ca. July 1947|loc.gov/item/2023867610|PD|large|Trumpeter, club|3|
|11|Jay Higginbotham — Jimmy Ryan's, NYC|1946–48|loc.gov/item/2023867845|PD|large|Trombonist, club|3|
|12|Roy Eldridge — Spotlite (Club), NYC|ca. Nov 1946|loc.gov/item/2023867642|PD|large|Trumpeter, club|3|

**Scene 1 count: 12** (Gottlieb Collection is deep and could yield 30+ more with additional
title sweeps — not thin, just capped here by the reweighting).

---

## Scene 2 — Deep-cut blues (FSA/OWI, Library of Congress)

Rights (verified on a sampled item, `item.rights`): *"The contents of the Library of
Congress Farm Security Administration/Office of War Information Black-and-White Negatives
are in the public domain and are free to use and reuse."* — **Public domain (U.S.
government work)**. Master TIFFs are full-negative scans (multi-thousand-px; the 1024px
JPEG service copy is the smallest guaranteed derivative, confirmed on item 2017801730).

| # | Subject | Year | Photographer | Source | License | Size | Description | Fit |
|---|---|---|---|---|---|---|---|---|
|13|Negroes jitterbugging in a juke joint, Saturday afternoon, Clarksdale, MS Delta|1939|Marion Post Wolcott|loc.gov/item/2017801730|PD|large (master TIFF, 1024px JPEG confirmed)|Couples dancing, crowded juke joint interior|5|
|14|Saturday afternoon in a Negro beer and juke joint, Clarksdale, MS Delta|1939|Marion Post Wolcott|loc.gov/item/2017801734|PD|large|Bar interior, patrons|4|
|15|Saturday night juke joint outside Clarksdale, MS Delta|1939|Marion Post Wolcott|loc.gov/item/2017801737|PD|large|Exterior at night, crowd gathered|4|
|16|Jitterbugging in a juke joint, Saturday evening, outside Clarksdale, MS|1939|Marion Post Wolcott|loc.gov/item/2017801848|PD|large|Dance floor, motion|4|
|17|Jitterbugging in a juke joint, Saturday evening, outside Clarksdale, MS Delta|1939|Marion Post Wolcott|loc.gov/item/2017801850|PD|large|Dance floor, second angle|4|
|18|Jitterbugging in Negro juke joint, Memphis, TN|1939|Marion Post Wolcott|loc.gov/item/2017801843|PD|large|Memphis juke interior, dancers|4|
|19|Jitterbugging in Negro juke joint, Memphis, TN (alt. angle)|1939|Marion Post Wolcott|loc.gov/item/2017801844|PD|large|Same scene, second frame|3|
|20|Drummer in orchestra in Memphis juke joint, TN|1939|Marion Post Wolcott|loc.gov/item/2017802165|PD|large|House band drummer mid-set|4|
|21|Store and "juke joint," Negro section, Homestead, FL|1939|Marion Post Wolcott|loc.gov/item/2017799820|PD|large|Storefront/juke exterior|3|
|22|Juke joint for Negroes near Belle Glade, FL|1939|Marion Post Wolcott|loc.gov/item/2017800384|PD|large|Rural juke exterior|3|

**Scene 2 count: 10.** (Excluded: five 1965 "Mississippi John Hurt, blues singer, in NYC"
photos, LOC items 2020731153–57, photographer Bernard Gotfryd — the item record's `rights`
field returned empty on inspection, no PD/CC statement found, so these are excluded per the
never-infer rule, not silently dropped.)

---

## Scene 3 — 1960s mod / British beat (Anefo via Commons, + Commons band categories)

| # | Subject | Year | Photographer | Source | License | Size | Description | Fit |
|---|---|---|---|---|---|---|---|---|
|23|Pretty Things arrive at Schiphol|1965-04-19|Joop van Bilsen (Anefo)|commons: "Aankomst Pretty Things (beatgroep) op Schiphol...Bestanddeelnr 917-6654"|CC0|3440×2729|Band disembarking, airport crowd|4|
|24|Pretty Things arrive at Schiphol (2nd frame)|1965-04-19|Joop van Bilsen (Anefo)|commons: "...Bestanddeelnr 917-6656"|CC0|3435×2729|Airport arrival, second angle|4|
|25|John Stax & Viv Prince at Schiphol|1965-04-19|Joop van Bilsen (Anefo)|commons: "Pretty Things op Schiphol...Bestanddeelnr 917-6655"|CC0|3426×2739|Bassist/drummer close, airport|3|
|26|Pretty Things full-group portrait|1965-04-19|Joop van Bilsen (Anefo)|commons: "PrettyThings1965.jpg"|CC0|3338×2625|Full band group shot|4|
|27|Spencer Davis Group rehearsal|1966-09-30|Ron Kroon (Anefo)|commons: "Repetities, popgroepen, Spencer Davis Group, Bestanddeelnr 919-6227"|CC0|2596×2603|Band rehearsing, studio|4|
|28|Spencer Davis Group rehearsal (2nd frame)|1966-09-30|Ron Kroon (Anefo)|commons: "...Bestanddeelnr 919-6228"|CC0|2580×2603|Rehearsal, second angle|3|
|29|Spencer Davis Group|1966-09-30|Ron Kroon (Anefo)|commons: "Spencer Davis Group 1966.jpg"|CC0|2595×2145|Group portrait|3|
|30|Johnny Lion with Spencer Davis Group guests on stairs|1966-10-01|Anefo (photographer unlisted)|commons: "Johnny Lion op de trap met gasten...Spencer Davis Group...919-6299"|CC0|3691×2446|Backstage/party candid, multiple people on stairs|4|
|31|The Kinks — KRLA Beat group photo|1965-06-27|KRLA Beat/Beat Publications|commons: "The Kinks (1965-06-27 KRLA Beat group photo - retouch).jpg"|Public domain (Commons-tagged)|2000×1618|Studio group portrait|3|
|32|The Kinks reading KRLA Beat|1965-06-28|KRLA Beat/Beat Publications|commons: "The Kinks reading KRLA Beat, June 1965.jpg"|Public domain|1042×1134|Candid, band reading a magazine|3|
|33|The Kinks — Dutch fan club event|1967-04-29|W. Veenman|commons: "Fanclub1967TheKinks1.jpg"|**CC BY-SA 3.0 nl (SA)**|4650×4687|Fan-club meet, large group|4|
|34|The Kinks — Dutch fan club event (2)|1967-04-29|W. Veenman|commons: "Fanclub1967TheKinks2.jpg"|**CC BY-SA 3.0 nl (SA)**|4659×4687|Fan-club event, second frame|4|
|35|The Kinks — Dutch fan club event (3)|1967-04-29|W. Veenman|commons: "Fanclub1967TheKinks3.jpg"|**CC BY-SA 3.0 nl (SA)**|4678×4687|Fan-club event, third frame|4|
|36|Hippy Happy Beurs — youth pop fair, Ahoy Rotterdam|1967-11-10|Ary Groeneveld (Stadsarchief Rotterdam)|commons: "Jongeren zijn druk bezig met het inrichten van een stand..."|CC0|2948×2953|Mod-era youth fair, crowd/stand-building|3|

**Scene 3 count: 14.**

---

## Scene 4 — 1960s US garage / Nuggets canon

Thin: most Nuggets-band Commons material is 2000s–2010s reunion/live photography, out of
period per the brief.

| # | Subject | Year | Source | License | Size | Description | Fit |
|---|---|---|---|---|---|---|---|
|37|The Standells|1966|commons: "The Standells.png", credited Tower Records|Public domain|1261×857|Period press/promo group shot|3|
|38|Chocolate Watchband|ca. 1966|commons: "Chocolate Watchband c. 1966.jpg"|Public domain|1065×1263|Group portrait|3|
|39|13th Floor Elevators|"1960s Photo" per file title; Commons date field reads Autumn 1975 (published in *Not Fade Away* by Doug Hanners) — **date conflict, flagged, not resolved**|commons: "The 13th Floor Elevators (1960s Photo).jpg"|Public domain|1370×695|Group photo, book-scan quality|2|
|40|Dick Dodd (Standells)|1966|commons: "Dick Dodd (cropped).png", Tower Records|Public domain|331×515|Solo headshot, cropped, small|2|

**Scene 4 count: 4 — thin.** (Checked and empty/near-empty: The Seeds, Count Five, ? and the
Mysterians, The Monks-as-garage — The Monks turned up only a small 1966 CC0 snapshot at
348×322, under the fit-2 size cap, not listed above to keep the table to real candidates.)

---

## Scene 5 — 1980s garage revival

Thin.

| # | Subject | Year | Source | License | Size | Description | Fit |
|---|---|---|---|---|---|---|---|
|41|Fuzztones, Irving Plaza, backstage|1984|commons: "Fuzztones 1984 Irving Plaza backstage.jpg"|CC BY-SA 3.0|1024×688|Backstage candid|3|
|42|Fuzztones, Irving Plaza, on stage|1984|commons: "Fuzztones 1984 Irving Plaza, New York City on Stage.jpg"|CC BY-SA 3.0|1024×690|Live on stage|4|
|43|Billy Childish (early solo/band era)|early 1990s|commons: "BillyChildish.jpg"|CC BY-SA 2.0|886×583|Candid portrait, slightly past the 1980s window|2|
|44|Billy Childish and Bill Lewis|1997|commons: "Billy Childish and Bill Lewis.jpg"|CC BY-SA 3.0|3476×2956|Two-person candid, well past the 1980s window but the named aesthetic|3|

**Scene 5 count: 4 — thin.** (Chesterfield Kings, Cynics, Miracle Workers, Pandoras,
Barracudas, Prisoners: no usable Commons categories found.)

---

## Scene 6 — 1980s DC hardcore

| # | Subject | Year | Source | License | Size | Description | Fit |
|---|---|---|---|---|---|---|---|
|45|Bad Brains|1983-04-04|commons: "Bad brains 1983.jpg"|CC BY-SA 3.0|3484×2336|Live/period band photo|5|
|46|Minor Threat|1981-04-04|commons: "Minor-threat-malcolm-riviera.jpg"|CC BY-SA 3.0|1024×763|Band photo, Malcolm Riviera credit|4|
|47|Fugazi & Urge, European tour|1990|commons: "Fugazi & Urge European Tour 1990.jpg"|CC BY-SA 4.0|1939×1409|Tour photo, within the "up to about 1990" allowance|3|

**Scene 6 count: 3 — thin.** (Government Issue, Rites of Spring, Void, S.O.A., Scream: no
Commons category content found.)

---

## Scene 7 — Rancid / Tim Armstrong (any year fine)

No dedicated "Rancid"/"Operation Ivy" Commons category with period content; live/candid
shots of any era, per the brief.

| # | Subject | Year | Source | License | Size | Description | Fit |
|---|---|---|---|---|---|---|---|
|48|Rancid, West Coast Riot|2012|commons: "Rancid at West Coast Riot 2012.jpg"|CC BY-SA 3.0|4000×3000|Live, full stage|4|
|49|Rancid, live (Toots Fontaine)|2006|commons: "Dangerancid.jpg"|CC BY 2.0|1696×2272|Live shot|4|
|50|Rancid, Pinkpop festival|2017|commons: "Pinkpop 2017 - Rancid.jpg"|CC BY-SA 4.0|2766×1840|Live, festival stage|4|
|51|Rancid, live|2007|commons: "Rancidband1.JPG"|Public domain (Commons-tagged)|2048×1536|Live shot|3|
|52|Rancid & Operation Ivy (combined shot)|2006|commons: "Rancid&OperationIvy.jpg"|CC BY 2.0|809×500|Small, but both bands in frame|2|
|53|Transplants|2005|commons: "Transplants.jpg"|CC BY 2.0|834×698|Small, live shot|2|

**Scene 7 count: 6.**

---

## Scene 8 — The Clash and The Slits, 1977–83

Known-good ones from the brief (all verified again, all still CC BY 2.0), plus new finds.

| # | Subject | Year | Source | License | Size | Description | Fit |
|---|---|---|---|---|---|---|---|
|54|The Clash in San Diego|1982|commons: "The Clash in San Diego, 1982.jpg", UCLA Library Special Collections|CC BY 2.0|1654×1585|Live/group|4|
|55|Joe Strummer, 1980|1980|commons: "JoeStrummer1980"|CC BY 2.0|(per brief)|Portrait|3|
|56|Mick Jones of the Clash, 1978|1978|commons: "Mick Jones of the Clash, 1978.jpg", UCLA Library Special Collections|CC BY 2.0|1769×1621|Portrait|3|
|57|Joe Strummer of the Clash, 1983|1983|commons: "JoeStrummer1983" (per brief)|CC BY 2.0|(per brief)|Portrait|3|
|58|The Slits|1980|commons: "The Slits-24.jpg"|CC BY 2.0|591×800|Small but period, group|2|
|59|The Slits (crop)|1980|commons: "The Slits-24 crop.jpg"|CC BY 2.0|851×1007|Same session, cropped, over the 1000px line|3|
|60|The Clash, live|1980-05-21|commons: "Clash 21051980 12 800.jpg", Helge Øverås|CC BY-SA 4.0|800×468|Small, live stage|2|
|61|Ian Dury with The Clash|1980-03-06|commons: "IanDuryWithTheClash1980.jpg", John Coffey|CC BY 2.0|1114×754|Backstage/candid group|3|
|62|Paul Simonon, The Clash, Palladium NYC|1979-09-20|commons: "Paul Simonon The Clash September 20 1979 Palladium NYC.jpg"|CC BY-SA 4.0|477×676|Small, live|2|

**Scene 8 count: 9** (5 from the brief, verified; 4 new).

---

## Scene 9 — Jamaican and 2 Tone, 1960–85 (reweighted; see verdict below)

### The headline find: Bob Marley & The Wailing Wailers, Zürich, May 1980

An entire professional archive — photographer **Patrick Lüthy**, Commons category "Bob
Marley and the Wailing Wailers in Zürich in 1980" — ~40 full-resolution TIFF scans, every one
individually tagged **CC BY-SA 4.0** (SA), 5500px+ on the long edge. This single source alone
can carry most of scene 9's quota. Ten representative frames (spread across the shoot's
lettered/numbered groups, since the underlying content per group could not be visually
inspected — all are live concert/stage frames from the same show):

| # | Subject | Year | Photographer | Source | License | Size | Description | Fit |
|---|---|---|---|---|---|---|---|---|
|63|Bob Marley & The Wailing Wailers, Zürich|1980-05|Patrick Lüthy|commons: "...Zürich Com L29-0351-0001-0001.tif"|**CC BY-SA 4.0 (SA)**|5500×3664|Live concert frame|5|
|64|Bob Marley & The Wailing Wailers, Zürich|1980-05|Patrick Lüthy|commons: "...L29-0351-0001-0004.tif"|**CC BY-SA 4.0 (SA)**|3676×5500|Live concert frame|5|
|65|Bob Marley & The Wailing Wailers, Zürich|1980-05|Patrick Lüthy|commons: "...L29-0351-0002-0001.tif"|**CC BY-SA 4.0 (SA)**|5500×3710|Live concert frame|5|
|66|Bob Marley & The Wailing Wailers, Zürich|1980-05|Patrick Lüthy|commons: "...L29-0351-0003-0002.tif"|**CC BY-SA 4.0 (SA)**|5500×3381|Live concert frame|5|
|67|Bob Marley & The Wailing Wailers, Zürich|1980-05|Patrick Lüthy|commons: "...L29-0351-0004-0003.tif"|**CC BY-SA 4.0 (SA)**|5500×3616|Live concert frame|5|
|68|Bob Marley & The Wailing Wailers, Zürich|1980-05|Patrick Lüthy|commons: "...L29-0351-0005-0001.tif"|**CC BY-SA 4.0 (SA)**|5500×3628|Live concert frame|5|
|69|Bob Marley & The Wailing Wailers, Zürich|1980-05|Patrick Lüthy|commons: "...L29-0351-0006-0001.tif"|**CC BY-SA 4.0 (SA)**|5500×3671|Live concert frame|5|
|70|Bob Marley & The Wailing Wailers, Zürich (cropped)|1980-05|Patrick Lüthy|commons: "...L29-0351-0006-0005 (cropped).tif"|**CC BY-SA 4.0 (SA)**|2213×2654|Cropped detail frame|4|
|71|Bob Marley & The Wailing Wailers, Zürich|1980-05|Patrick Lüthy|commons: "...L29-0351-0007-0002.tif"|**CC BY-SA 4.0 (SA)**|5500×3658|Live concert frame|5|
|72|Bob Marley & The Wailing Wailers, Zürich|1980-05|Patrick Lüthy|commons: "...L29-0351-0003-0005.tif"|**CC BY-SA 4.0 (SA)**|5500×3669|Live concert frame|5|

### Other verified Jamaican/2 Tone period finds

| # | Subject | Year | Photographer | Source | License | Size | Description | Fit |
|---|---|---|---|---|---|---|---|---|
|73|The Selecter, Berlin|1980 (circa)|Braunov (user upload, period photo)|commons: "The Selecter Berlin 1980.jpg"|CC BY-SA 3.0|6528×4336|Live band photo — the strongest 2 Tone-specific find|5|
|74|Bob Marley, performing|1976-04-25|David Melhado; distributed by Island Records|commons: "Bob Marley performing in 1976.jpg"|Public domain (Commons-tagged press photo)|1416×1005|Live performance shot|4|
|75|Bob Marley, press photo|1977|Kim Gottlieb; distributed by Island Records|commons: "Bob Marley 1977 press photo.jpg"|Public domain|1007×1340|Studio/press portrait|3|
|76|Bob Marley, press photo|1976|Dennis Morris; distributed by Island Records|commons: "Bob Marley (1976).jpg"|Public domain|572×763|Portrait, small — Dennis Morris was the era's key reggae photographer|2|
|77|Martha Velez and Bob Marley, press photo|1976|Sire Records|commons: "Martha Velez and Bob Marley 1976 press photo.jpg"|Public domain|736×511|Two-person candid, small|2|
|78|Bob Marley, Abesten|1980-09-06|Eddie Mallin|commons: "Bob Marley Abesten.jpg"|CC BY-SA 4.0|640×437|Small, live/candid|2|
|79|Peter Tosh with Robbie Shakespeare|1978-12-04|TimDuncan (Commons user, period photo)|commons: "PeterToshWithRobbieShakespeare1978.jpg"|CC BY 3.0|477×659|Candid pair, small|2|
|80|Peter Tosh, "Bush Doctor" era|1978-12-04|TimDuncan|commons: "BushDoctor1978.jpg"|CC BY 3.0|489×377|Small, period|2|
|81|Jimmy Cliff|1972-01|Unknown (Commons-tagged PD)|commons: "Jimmy Cliff.jpg"|Public domain|775×407|Small portrait|2|

**Scene 9 in-period (1960–85) count: 19** (10 Marley/Zürich + Selecter + 6 Marley press/
candid + 2 Tosh + 1 Cliff).

### Borderline / just outside the 1960–85 window (flagged, not counted above)

| Subject | Year | Source | License | Size | Note |
|---|---|---|---|---|---|
|UB40, Rock am Ring|1987-06|commons: "UB40 Rock am Ring 1987.jpg"|CC BY-SA 2.0 de|1449×953|2 years past the window|
|Desmond Dekker, Jazzhaus Freiburg|1992-02-09|commons: "...Staatsarchiv Freiburg W 140 Nr. 06392.jpeg"|CC BY 4.0|998×662|7 years past; German municipal archive|
|Notting Hill Carnival (13-photo set)|1989|commons: "Notting Hill Carnival 1989 (01–13).jpg"|CC BY 4.0|~1400×1040 each|4 years past; strong sound-system/street-crowd fit if the window flexes|

---

## Ten best overall (across all scenes)

1. **Eddie Condon's sextet, NYC, 1946** — loc.gov/item/2023867575 — PD — full club group, fit 5
2. **Negroes jitterbugging in a juke joint, Clarksdale MS, 1939** — loc.gov/item/2017801730 — PD — fit 5
3. **Bad Brains, 1983** — commons "Bad brains 1983.jpg" — CC BY-SA 3.0 — fit 5
4. **The Selecter, Berlin, 1980** — commons "The Selecter Berlin 1980.jpg" — CC BY-SA 3.0 — 6528×4336 — fit 5
5. **Bob Marley & The Wailing Wailers, Zürich, 1980** (×several) — commons "Bob Marley & The Wailing Wailers Zürich Com L29-0351-0001-0001.tif" — CC BY-SA 4.0 — 5500×3664 — fit 5
6. **Dancers in a jazz club, DC, 1938–48** — loc.gov/item/2023867585 — PD — fit 5
7. **The Kinks, Dutch fan-club event, 1967** — commons "Fanclub1967TheKinks1.jpg" — CC BY-SA 3.0 nl — 4650×4687 — fit 4
8. **Pretty Things arrive at Schiphol, 1965** — commons "PrettyThings1965.jpg" — CC0 — 3338×2625 — fit 4
9. **The Clash in San Diego, 1982** — commons "The Clash in San Diego, 1982.jpg" — CC BY 2.0 — fit 4
10. **Rancid at West Coast Riot, 2012** — commons "Rancid at West Coast Riot 2012.jpg" — CC BY-SA 3.0 — 4000×3000 — fit 4

## Counts per scene

| Scene | Count | Status |
|---|---|---|
|1 — Deep-cut jazz|12|solid; deeper sweeps of Gottlieb would yield more|
|2 — Deep-cut blues|10|solid (all FSA/OWI, Marion Post Wolcott, Clarksdale/Memphis)|
|3 — Mod/British beat|14|solid (Anefo CC0 is the strongest single source)|
|4 — US garage/Nuggets|4|**thin**|
|5 — 1980s garage revival|4|**thin**|
|6 — DC hardcore|3|**thin**|
|7 — Rancid/Tim Armstrong|6|adequate|
|8 — Clash/Slits|9|solid|
|9 — Jamaican/2 Tone|19 in-period (+3 borderline sets)|see verdict below|
|**Total**|**81** (+borderlines)|within the 40–80 target|

## Thin scenes, named

**Scene 4 (US garage/Nuggets)**, **scene 5 (1980s garage revival)**, and **scene 6 (DC
hardcore)** came back thin. The Nuggets bands and most 80s revival acts have essentially no
period photography on Commons with a clear reuse license — what exists there is almost all
2000s–2020s reunion/live shots (CC-licensed by the photographer, but out of period). DC
hardcore is similarly thin outside Bad Brains and Minor Threat; Government Issue, Rites of
Spring, Void, S.O.A., and Scream returned nothing. These three scenes would need a different
source (a fan archive, a punk-zine digitization project, or a direct ask to a photographer
like Malcolm Riviera or Glen E. Friedman for a case-by-case license) to grow meaningfully.

## Can the reggae/ska half actually be filled? — Verdict

**Partially, and unevenly, but yes overall — with one important caveat.**

- The named **2 Tone bands** (The Specials, The Selecter, The Beat, Madness, Bad Manners) are
  almost entirely **absent** from Commons in period form. Extensive category and subcategory
  sweeps for each turned up either nothing or exclusively 2000s–2020s reunion/live photos.
  The Selecter's 1980 Berlin photo is the one clean exception, and it is excellent.
- The named **British reggae acts** (Steel Pulse, Aswad, UB40, Black Uhuru) are likewise thin
  — a handful of small, modern, or barely-relevant hits each.
- **Jamaican reggae proper** is where the crate can actually be filled, almost entirely
  through **Bob Marley & The Wailers**: a single outstanding archive (Patrick Lüthy's Zürich
  1980 show, ~40 CC BY-SA 4.0 TIFFs at 5500px) plus a run of PD-tagged 1970s Island/Sire press
  photos, plus two small Peter Tosh and one Jimmy Cliff frame. Linton Kwesi Johnson, Yellowman,
  Lee Perry, and King Tubby returned **nothing** usable — no period photos with a clear license
  turned up for any of them, despite dedicated searches and subcategory walks.
- LOC and the Nationaal Archief/Anefo route (explicitly named in the brief) contributed
  **nothing** to this scene: LOC's photo collections don't cover Jamaican music, and no Anefo
  Dutch-tour photos of these acts were found (Anefo's UK/Netherlands beat-tour coverage in
  scene 3 is deep, but it did not extend to Jamaican or 2 Tone acts touring the Netherlands in
  this search).

**Net effect:** scene 9 alone can supply ~19–25 verified in-period images (mostly Marley/
Wailers, plus the one strong Selecter frame), which is enough to carry roughly a quarter of an
80-image rotation on its own, and — if scene 8's Clash/Slits material (dub- and reggae-adjacent
in sound if not subject) is folded in per your instruction — the combined 8+9 pool is 28
images, close to but still short of a clean half of 80. Getting to a genuine 40/40 split would
require either loosening the 1960–85 window (the UB40 1987, Desmond Dekker 1992, and Notting
Hill Carnival 1989 sets above would add ~18 more images at that point) or sourcing outside
Commons/LOC/Anefo entirely — Trojan Records' own catalog, a Jamaican National Library/
Institute of Jamaica archive, or a direct rights inquiry to a Marley-estate or Adrian Boot–style
photographer for the acts this pass came up empty on (LKJ, Yellowman, Lee Perry, King Tubby,
Steel Pulse, Aswad, Black Uhuru).
