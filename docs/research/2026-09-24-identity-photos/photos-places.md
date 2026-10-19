# Places photo research — dubplate login rotation

Wikimedia Commons search, category walks, and license verification for
recording-studio and landmark-location photos to add to the login page's
silver-print rotation. Method: `action=query` category-members and
`list=search` against the Commons API, `imageinfo` with `extmetadata`,
1.5s+ sleep between calls, User-Agent header sent on every request.

License classes shippable: PD, CC0, CC BY, CC BY-SA. NC licenses recorded
separately and flagged, never shipped. Uploader-vs-photographer mismatches
("own work" claim but description credits someone else) flagged as
provenance risk.

Fit score 1-5: atmosphere/texture/signage/people in frame score high; flat
daytime building shots 2-3; plaques alone 2; anything under 1000px long
edge caps at 2.

---

## Jamaica

Commons has almost no photography of the actual studio buildings named in the
brief. Targeted searches for "Randy's Records", "Channel One studio
Kingston", "Black Ark studio", "Federal Records Jamaica", "Dynamic Sounds
Jamaica", "Harry J studio Jamaica", "Treasure Isle studio Kingston", "13
Brentford Road", and "King Tubby Waterhouse" returned zero on-topic hits —
these buildings are essentially undocumented on Commons. Studio One's street
(Brentford Road) only returns London's Brentford. Recorded as "nothing
usable" below.

**Best finds:**

1. **Bass Odyssey Sound System speaker column, Tropical Hut, St. Mary, May
   2012.jpg** — https://commons.wikimedia.org/wiki/File:Bass_Odyssey_Sound_System_speaker_column,_Tropical_Hut,_St._Mary,_May_2012.jpg
   — exterior, stacked speaker boxes at a live dance. CC BY 3.0. Photographer
   Yaniq Walford, own work, credit matches uploader. 5184x3456. **Fit 5** —
   exactly the speaker-box/sound-system atmosphere the brief calls out.
2. **Corner Coxsone Dodd Kingston Jamaica.jpg** — https://commons.wikimedia.org/wiki/File:Corner_Coxsone_Dodd_Kingston_Jamaica.jpg
   — street corner named for the Studio One founder. CC BY-SA 2.0. Christina
   Xu, 2007, via Flickr (dubdem-adjacent account). 3008x2000. **Fit 4** —
   street signage/place-name connection to Studio One, good light and grain.
3. **Vinyl pressing machine.jpg** — https://commons.wikimedia.org/wiki/File:Vinyl_pressing_machine.jpg
   — interior, a Jamaican record-pressing plant. CC BY-SA 2.0. Christina Xu,
   2007. 3008x2000. **Fit 4** — interior machinery texture, on-theme for a
   dub label's home turf even without a named studio.
4. **Street dance Kingston Jamaica passa passa.jpg** — https://commons.wikimedia.org/wiki/File:Street_dance_Kingston_Jamaica_passa_passa.jpg
   — night street dancehall crowd, Passa Passa street session. CC BY-SA 2.5.
   nickolette, 2007. 1600x1066. **Fit 5** — crowd, dancehall atmosphere,
   exactly the "sound-system and dancehall sites" brief.
5. **Culture Yard series** (Trench Town Culture Yard, Bob Marley's boyhood
   home) — 11 photos by YardEdge, CC BY 2.0, 2010-03-13, e.g.
   https://commons.wikimedia.org/wiki/File:Culture_Yard_(4546721487).jpg —
   exteriors/interiors of the yard, bar, library, shop, kids playing. All
   570px long edge, **capped at fit 2** by size floor despite strong
   atmosphere; would need upscaling to use.
6. **Hope Road Jamaika.jpg / Hope Road Jamaika2.jpg** — https://commons.wikimedia.org/wiki/File:Hope_Road_Jamaika.jpg
   — street view near the Bob Marley Museum/Tuff Gong on Hope Road. CC BY-SA
   4.0. Hitachi-Hu, 2019, own work. 2048x1536. **Fit 3** — flat daytime
   street shot, but genuine location value.
7. **Bob Marley Museum 2006 1–4.jpg** — https://commons.wikimedia.org/wiki/File:Bob_Marley_Museum_2006_1.jpg
   (and _2, _3, _4) — museum exterior/grounds. CC BY 2.0. Credited to
   "Dubdem • Música • Design • Sound System" via Flickr. 476–650px.
   **Fit 2**, capped by size.
8. **Half way tree clock.jpg** — https://commons.wikimedia.org/wiki/File:Half_way_tree_clock.jpg
   — the Half Way Tree clock-tower landmark. CC BY-SA 3.0. Raw9345, 2011, own
   work. 1600x1200. **Fit 3**.
9. **Valentine and Sons – At Half Way Tree, Jamaica, 1891** — https://commons.wikimedia.org/wiki/File:Valentine_and_Sons_-_At_Half_Way_Tree,_Jamaica,_1891_3085032083.jpg
   — historic PD image (Caribbean Photo Archive via Flickr Commons). 2048x1448.
   **Fit 3** — the brief allows any year for places.
10. **A. Duperly & Sons – Jubilee Market, Orange Street, Kingston, Jamaica**
    — https://commons.wikimedia.org/wiki/File:A._Duperly_%26_Sons_-_Jubilee_Market,_Orange_Street,_Kingston,_Jamaica_8433582139.jpg
    — pre-1922, Public Domain. 1885x1234. **Fit 3** — historic Orange
    Street ("Beat Street") market scene, predates the reggae era but a
    genuine place record.
11. **House at Chelsea Av. Half Way Tree, Jamaica 1988.jpg** — https://commons.wikimedia.org/wiki/File:House_at_Chelsea_Av._Half_Way_Tree,_Jamaica_1988.jpg
    — CC BY-SA 4.0. De Geo, 1988. 1736x1180. **Fit 2** — residential, flat
    documentary shot.
12. **Tuff Gong 1.jpg** — https://commons.wikimedia.org/wiki/File:Tuff_Gong_1.jpg
    — CC BY 2.0, "dubdem sound system" credit via Flickr, 2006. 768x394.
    **Fit 2**, capped by size (and below 1000px cap rule).

**Provenance flags:** none of the above claim "own work" while crediting a
different photographer in the description — no mismatch found in this batch.
The Culture Yard and Bob Marley Museum sets both trace to Flickr accounts
(YardEdge; Dubdem/dubdem sound system) rather than direct Commons uploaders
claiming personal authorship, so the CC BY / CC BY-SA 2.0 licenses are
Flickr-sourced and machine-verified by Commons' own FlickreviewR bot in each
case — treat as verified, not a risk flag.

**Nothing usable:** Studio One (13 Brentford Road building itself), Treasure
Isle, Black Ark (and its ruins), King Tubby's studio/Waterhouse, Channel One,
Tuff Gong's actual studio building (only the Hope Road street and unrelated
merch/museum shots), Harry J, Randy's Records, Federal Records, Dynamic
Sounds — zero on-topic Commons results for any of these specific buildings.

---

## Clubs (owner priority: reggae/dub/dancehall/ska and punk/protopunk cores first)

Searched hardest and first per owner instruction. Interiors, exteriors,
marquees, and crowds, any year.

### Reggae/dub/dancehall/ska core

- **Jamaican dancehall/sound-system venues**: covered above under Jamaica —
  best are Bass Odyssey Sound System speaker column (fit 5), Street dance
  Kingston Jamaica passa passa (fit 5), and Corner Coxsone Dodd (fit 4). No
  further dedicated venue building (no Channel One, no dedicated dancehall
  lawn) turned up beyond those.
- **Notting Hill Carnival sound systems**: the Category tree
  (Category:Notting Hill Carnival 1974–2026, 20+ year subcats) holds
  thousands of CC BY-SA 2.0/4.0 photos, mostly mas-band costumes and crowds
  rather than confirmed speaker rigs by filename alone. One confirmed
  sound-system-truck shot: **This is the City, Jamaican sound system —
  geograph.org.uk — 7208771.jpg** — https://commons.wikimedia.org/wiki/File:This_is_the_City,_Jamaican_sound_system_-_geograph.org.uk_-_7208771.jpg
  — CC BY-SA 2.0, A J Paxton, 2022, 2592x1936. **Fit 5.** Also usable:
  **Notting Hill Carnival 2007 004.jpg** (float in parade), CC BY 2.0,
  Dominic Alves, 2484x1944, fit 3. A full sweep of the year subcats for
  speaker-stack floats specifically would need individual image review
  beyond this pass's budget — flagging as a rich, only partially mined
  vein.

### Punk/protopunk core

1. **9:30 Club (930 F Street NW, original location)**
   - **FIRST FLOOR REAR ROOM, SAFE DEPOSIT ROOM (Note: this was the 9:30
     Club during the early 1980s) — Atlantic Building, 930 F Street NW,
     HABS DC,WASH,655A-5.tif** — https://commons.wikimedia.org/wiki/File:FIRST_FLOOR_REAR_ROOM,_SAFE_DEPOSIT_ROOM_(Note-_This_was_the_9-30_Club_during_the_early_1980s)_-_Atlantic_Building,_930_F_Street,_Northwest,_Washington,_District_of_Columbia,_DC_HABS_DC,WASH,655A-5.tif
     — Public Domain (HABS/Library of Congress). 3563x5000. **Fit 3** —
     architectural documentation, not a gig shot, but the actual original
     room, rare and PD.
   - **9-30Club WashingtonDC 1990.jpeg** / **9-30Club WashingtonDC interior
     1990.jpeg** — Public Domain, John McWilliams, Library of Congress,
     Summer 1990. 457x640 / 456x640. **Fit 2**, capped by size — but these
     are exterior/interior of the actual original-location club at the
     time it operated.
   - **Nightclub 930 Closed on a June summer night.jpg** — https://commons.wikimedia.org/wiki/File:Nightclub_930_Closed_on_a_June_summer_night.jpg
     — CC BY-SA 4.0, Terps2008, 2015, 6016x4000. **Fit 5** — moody lit
     marquee at night (current V Street location).
   - **9-30 Club Hall of Records.jpg** — CC BY-SA 4.0, VinylStyle, 2016,
     4067x2288. **Fit 4** — interior with record-wall décor.
   - **9-30 Club (34590147581).jpg** / **9 30 Club (55256118542).jpg** —
     CC BY 2.0 / CC BY 4.0, Paul Sableman (2017) / ajay_suresh (2026),
     4756x3171 / 5833x3889. **Fit 3** exterior daytime.

2. **CBGB (315 Bowery)**
   - **CBGB club facade.jpg** — https://commons.wikimedia.org/wiki/File:CBGB_club_facade.jpg
     — CC BY-SA 3.0, Adicarlo, 2005, 1600x1200. **Fit 5** — the real
     Bowery storefront with the awning, while open.
   - **CBGB closed.JPG** — CC BY 2.5, Noeticsage, 2006-10-16, 2816x2112.
     **Fit 4** — facade the day it closed, "CBGB CLOSED" signage.
   - **CBGB the day after.JPG** — CC BY-SA 3.0, Romanontheprowl,
     2006-10-16, 3456x2304. **Fit 4** — crowd/memorial atmosphere at the
     facade.
   - **CBGB, Manhattan (exterior view).JPG** — Public Domain, Schlamniel,
     2006, 1632x1224. **Fit 4**.
   - **CBGB Awning — Rock and Roll Hall of Fame.jpg** —
     https://commons.wikimedia.org/wiki/File:CBGB_Awning_-_Rock_and_Roll_Hall_of_Fame_(2014-12-30_11.15.07_by_Sam_Howzit).jpg
     — CC BY 2.0, Sam Howzit, 2014, 5616x3744. **Fit 4** — the actual
     preserved awning, though now a museum artifact rather than in situ.
   - **ExCBGB.jpg** — Public Domain, Sbazzone, 2008, 2033x1920. **Fit 3**
     — post-closure facade.
   - Provenance flag: **CBGB '73 (Manhattan, New York) P001.jpg**, credited
     "own work" by Leonard J. DeFrancisci but uploaded 2017 with a 1973
     date — plausibly the photographer's own decades-old negative scanned
     later (consistent, not necessarily a mismatch), but worth a second
     look before use given the "own work" + retroactive date combination.
   - CBGB 2009 interior series (Gabba Gabba Hey, Grand Piano, Skull and
     Ramones, etc., CC BY 2.0, Angie Garrett, 2009, ~2816x1880): titles and
     décor (Ramones art, pinball) suggest these may be the Rock and Roll
     Hall of Fame's CBGB room reconstruction rather than the original
     Bowery site — flag as unverified location before use.

3. **924 Gilman Street, Berkeley** — the best-documented punk venue found.
   Photographer Jonathan Botkin's images explicitly state "sent directly to
   me by the photographer... for inclusion on Wikimedia," CC BY-SA 4.0,
   spanning 2018–2025:
   - **2025-05-03 Gilman Exterior 0008a.jpg** — https://commons.wikimedia.org/wiki/File:2025-05-03_Gilman_Exterior_0008a.jpg
     — 4682x3312. **Fit 5** — current exterior/signage.
   - **2025-05-03 Gilman Exterior 0013a.jpg / 0843a.jpg** — similar,
     3743x3912 / 5186x2724. **Fit 4-5**.
   - **924 Gilman Street.JPG** — CC BY-SA 3.0, Coro, 2009, 2535x1534.
     **Fit 4** exterior.
   - **924 Gilman Street (4)-(7).jpg** — CC BY 2.0, milesgehm, 2007,
     2592x1944 each. **Fit 3-4**, exterior/street context.
   - Interior show shots (2018–2024, various bands) also CC BY-SA 4.0 with
     the same explicit-permission provenance — good backup stock if more
     variety is wanted.

4. **Grande Ballroom, Detroit (MC5/Stooges venue)**
   - **Grande Ballroom Interior 2010.jpg** — https://commons.wikimedia.org/wiki/File:Grande_Ballroom_Interior_2010.jpg
     — CC BY-SA 3.0, Albert duce, 2010, 4256x2832. **Fit 5** — decayed,
     atmospheric interior ruin, exactly the mood the brief wants.
   - **Grande Ballroom Detroit 2009.jpg** — CC BY 3.0, Albert duce, 2009,
     5616x3744. **Fit 4** exterior ruin.
   - **Grande Ballroom window detail.jpg** — CC BY-SA 4.0, Brianleecox,
     2017, 3024x4032. **Fit 4** — decay/texture detail.
   - **Grande Ballroom Detroit Rock Music Landmark on Grand River.jpg** —
     CC BY-SA 3.0, Mike Russell, 2008, 1959x1007. **Fit 4** — exterior with
     landmark signage.

5. **The Rat (Kenmore Square, Boston)** — thin: **YoungSnakesRat.jpg** —
   CC BY-SA 3.0, David Henry, 1981, 1354x1050. Band performing, presumably
   inside the club. **Fit 4** for atmosphere and period authenticity, though
   unverified as interior-of-The-Rat specifically vs. a general period shot
   attributed to it by the source site.

6. **100 Club (Oxford Street, London)**
   - **The 100 Club Oxford Street, London 2023-04-26.jpg** —
     https://commons.wikimedia.org/wiki/File:The_100_Club_Oxford_Street,_London_2023-04-26.jpg
     — CC BY 2.0, Roy Katzenberg, 4458x3304. **Fit 4** exterior/signage.
   - **100-club-oxford-st-london.jpg** — CC BY-SA 3.0, Le Deluge, 2010,
     1968x3253. **Fit 4** exterior/signage, portrait framing.
   - **Peter Doherty, 100 Club, London (52122518571).jpg** — CC BY 2.0,
     Drew de F Fawkes, 2022, 12000x9000 (huge). **Fit 4** interior gig shot.

7. **The Masque (Hollywood)** — **nothing usable**; every "Masque" search
   result was an unrelated African/ceremonial mask object.

8. **The Roxy (Neal Street/Covent Garden, London)** — the club itself is
   gone; **41-43 Neal Street — Speedo The Roxy (2599362142).jpg** — CC BY
   2.0, Emanuel Berglund, 2008, 2592x1944. **Fit 3** — current building at
   the site, shopfront signage visible.

9. **Max's Kansas City** — site marker only: **This Was Max's Kansas City
   (2962732199).jpg** and **Max's Kansas City.jpg** — both CC BY 2.0, Eden
   Janine and Jim, 2008, ~1273x1908. **Fit 3** — building/plaque at the
   former site.

10. **Mudd Club** — **Facciata del Mudd Club @ NYC.jpg** —
    https://commons.wikimedia.org/wiki/File:Facciata_del_Mudd_Club_%40_NYC.jpg
    — CC BY-SA 3.0, Wickkey, 2012, 2736x3648. **Fit 3-4** — the building
    facade.

11. **Marquee Club (Wardour Street)** — original venue demolished, only
    plaque/site photos: **The site of the Marquee Club, 90 Wardour Street,
    Soho, London.jpg** and **Keith Moon... performed here at the site of
    the Marquee Club...jpg** — both CC BY-SA 4.0, Spudgun67, 2019. **Fit 2**
    (plaque rule).

**Cavern Club (Liverpool)** — richly documented, mostly interior:
- **The Cavern of the Cavern Club, 2009.jpg** — https://commons.wikimedia.org/wiki/File:The_Cavern_of_the_Cavern_Club,_2009.jpg
  — CC BY 2.0, Hens Zimmerman, 4000x3000. **Fit 5** — dark brick-tunnel
  interior, strong atmosphere.
- **The Cavern Club interior 2, Mathew Street, Liverpool, 2010.jpg** — CC
  BY 2.0, Jennifer Boyer, 2848x4288. **Fit 4**.
- **The Cavern Club stage 2, Mathew Street, Liverpool, 2008.jpg** — CC BY
  2.0, George M. Groutas, 3872x2592. **Fit 4**.
- **The Cavern Club wall, Mathew Street, Liverpool, 2011.jpg** — CC BY 2.0,
  Ian Ransley, 4288x2848. **Fit 4** — brick tunnel texture.

### Other clubs searched (jazz, mod, rock; secondary per owner ranking)

- **Village Vanguard**: **Village Vanguard 2018.jpg** (CC BY-SA 4.0, Alwin
  Tong, 4032x3024, fit 4, night entrance sign); **The Village Vanguard at
  night 1976.jpg** (CC BY-SA 2.0, Tom Marcello, 1024x679, fit 4, moody
  night shot); historic Gottlieb/LOC portraits from 1947 are PD but mostly
  under 1000px, capped at fit 2.
- **Minton's Playhouse**: **Minton's Playhouse.jpg** (CC BY-SA 2.0,
  3264x2448, fit 3-4, exterior); the famous 1947 Gottlieb photo of Monk,
  McGhee, Eldridge, and Hill inside the club is PD but 928x968, **capped
  at fit 2** by the size floor.
- **Apollo Theater (Harlem)**: **The Apollo at Night 05/06/07.jpg** (CC
  BY-SA 4.0, Bohemian Baltimore, 2019, ~4048x3036) — **fit 5**, lit
  marquee at night; **Apollo neon sign.jpg** (CC BY-SA 2.0, All-Nite
  Images, 2013, 3456x4608, fit 4).
- **Roundhouse (Camden)**: **Interior, Camden Roundhouse.jpg** (CC BY-SA
  4.0, Stephencdickson, 2013, 3088x2056, fit 4); **The Roundhouse.jpg**
  (Public Domain, AlanHancock, 2010, 2560x1920, fit 3 exterior).
- **Hammersmith Palais**: **Hammersmith Palais de Danse.jpg** / **(left)**
  / **(right)** — CC BY-SA 2.0, Ewan-M, 2012, ~2980x1987, fit 3-4
  exterior; a 1919 PD dancefloor photo exists but is 415x303, capped at 2.
- **Flamingo Club (Wardour Street)**: site plaque only — **Site of The
  Flamingo Club 33-37 Wardour Street...jpg**, CC BY-SA 4.0, Spudgun67,
  2017, 1500x2000, **fit 2**.
- **Scene Club (Soho)** and **Eel Pie Island Hotel/ballroom** —
  **nothing usable**: Eel Pie Island returns only nature/boathouse
  geograph shots of the island generally (the hotel that hosted the 1960s
  R&B scene burned down in 1971); a 1900 postcard of the hotel exists but
  is 739x468, capped at 2 and pre-dates the music history entirely.
- **Fillmore East (Second Avenue, NYC)**: **Fillmore East Exterior.jpg**
  (CC BY-SA 3.0, Fletcher6, 2010, 3008x2000, fit 4); **Site of Fillmore
  East 105 Second Avenue...jpg** (CC BY-SA 4.0, Spudgun67, 2016, fit 2,
  plaque).
- **Whisky a Go Go (Sunset Strip)**: **Whisky A Go Go Sunset Strip.jpg**
  (CC BY-SA 2.0, Ken Lund, 2015, 4000x3000, fit 4); **Marquee outside
  Whiskey a Go Go on the Sunset Strip, June 2, 2009.JPG** — note license
  is **CC BY 1.0** (an early, rarely-used CC version; verify it resolves
  to a compatible attribution-only license before shipping), 2572x1928.
- **Maxwell's (Hoboken)**: the "11.7.11Maxwell'sByLuigiNovi" series, CC BY
  3.0, Nightscream, 2011, 2048x1536 each. **Fit 3** exterior/street level.
- **Birdland**: current club — **Birdland Jazz Club (55283714363).jpg**,
  CC BY 4.0, Ajay Suresh, 2026, 7690x4326, fit 4; a pre-1952 PD postcard
  of the original club's interior/entrance exists (**Birdland club
  interior.jpg**, 1207x770, fit 3; **Birdland club entrance.jpg**, 585x689,
  capped at 2).
- **Five Spot (Cooper Square)**, **Tractor Tavern (Ballard)**, **Milestone
  Club (Charlotte)**, **Dischord House (Arlington)** — **nothing usable**;
  zero on-topic Commons results for any of these four.

---

## UK (non-club places)

Half Way Tree and other Jamaica-adjacent UK-diaspora sites are under
Jamaica above. Additional finds:

- **Rough Trade (original Talbot Road/Notting Hill shop)**:
  **Rough Trade, Notting Hill, W11.jpg** — CC BY-SA 2.0, Ewan-M, 2011,
  2635x1976. **Fit 4** — shopfront exterior. **Rough Trade Records, Talbot
  Road — geograph.org.uk — 1279495.jpg** — CC BY-SA 2.0, 640x480, **fit
  2**, capped by size.
- 2 Tone Village (Coventry), Horizon Studios (Coventry), the Specials'
  plaques, blue plaques for reggae/ska artists, Brixton, Handsworth,
  Ladbroke Grove/All Saints Road, Brighton seafront, Margate — **not
  searched this pass**; time and call budget went to the clubs core per
  owner priority. Flag as an open follow-up if the rotation still needs
  UK depth beyond Carnival/Roundhouse/Cavern/100 Club/Rough Trade above.

---

## US (non-club places)

- **Chess Records (2120 South Michigan Avenue, Chicago)**
  - **Chess Records Studio.jpg** — https://commons.wikimedia.org/wiki/File:Chess_Records_Studio.jpg
    — CC BY-SA 2.0, Zol87, 2009, 1944x2592. **Fit 4** — exterior.
  - **Chess Records (2639758382).jpg** — CC BY 2.0, Reading Tom, 2006,
    2592x1944. **Fit 3-4** exterior.
- **Stax Records (Memphis)** — Bubba73's 2022 set (CC BY-SA 3.0, own work,
  ~30 numbered images, 4016x6016/6016x4016) covers the rebuilt
  Stax Museum building from multiple angles at very high resolution.
  Representative: **Stax Records, Memphis, TN, US (03).jpg** — https://commons.wikimedia.org/wiki/File:Stax_Records,_Memphis,_TN,_US_(03).jpg
  — **fit 4** exterior. Also: **Stax Records Memphis 20ch mixing board.jpg**
  and **Stax Records Memphis Courtney recording studio 1.jpg** — CC BY 2.0,
  Mr. Littlehand, 2013, 4000x3000. **Fit 4-5** — interior control-room and
  studio gear, strong atmosphere.
- **FAME Studios / Muscle Shoals Sound (Muscle Shoals, AL)** — the best US
  studio coverage found:
  - **FAME Recording Studios sign, Muscle Shoals, Alabama.jpg** — CC BY
    2.0, Ralph Daily, 2013, 4287x2750. **Fit 4** exterior signage.
  - **FAME Recording Studios - Studio A control room with board.jpg** —
    CC BY 2.0, Ralph Daily, 2013, 4288x2848. **Fit 5** — interior control
    room, board visible, strong atmosphere.
  - **FAME Recording Studios - Through these doors.jpg** — CC BY 2.0,
    Ralph Daily, 2013, 4288x2848. **Fit 4** interior corridor.
  - **FAME Recording Studios, Muscle Shoals, Alabama LCCN2010641236.tif**
    (and 2010640534/641237/641238) — Public Domain, Carol M. Highsmith/
    Library of Congress, 2010, ~6144x4096. **Fit 4** — professional
    exterior documentation, PD and huge.
  - **Rick Hall, founder of FAME Recording Studios, sitting in the FAME
    studio...LCCN2010641239.tif** (and 641240/641241) — Public Domain,
    Highsmith, 2010, 6144x4096. **Fit 3** — portrait-in-studio rather than
    a pure place shot, but PD and on-topic.
  - **Muscle Shoals Sound Studio, main area, 2025.jpg** and **recording
    booth, 2025.jpg** — CC BY 4.0, Peter K Burian, 3754x2400/3534x2400.
    **Fit 4-5** interior.
  - **Muscle Shoals Sound Studio Exterior.jpg** / **Interior.jpg** — CC
    BY-SA 4.0, A1bi, 2022, 3351x2535/4032x3024. **Fit 4**.
- **Sun Studio (Memphis)** — search errored on a transient network fault
  and was not re-run this pass; **flag as unchecked**, likely productive
  given the Chess/Stax/FAME hit rate on similar queries.
- CBGB, 9:30 Club, Grande Ballroom, Village Vanguard, Minton's, Apollo,
  Fillmore East, Whisky a Go Go, Maxwell's, Birdland — see Clubs section
  above.
- Rudy Van Gelder's studio (Englewood Cliffs/Hackensack), the Village
  Vanguard's neighborhood, Preservation Hall, Clarksdale juke joints
  (Red's, Ground Zero), the Crossroads marker, Dockery Farms, Maxwell
  Street, CBGB's facade (done above), Gilman (done above), Moondog's
  corner — **not searched this pass**; budget went to the clubs core and
  the studio trio (Chess/Stax/FAME) per owner priority.

---

## Record shops

- **Rough Trade** (Notting Hill original shop) — see UK section above,
  fit 4.
- **Dub Vendor, Daddy Kool, Honest Jon's, Randy's, Rockers International**
  — **not searched this pass**; time went to the clubs core per owner's
  re-prioritization mid-task. Flag as open follow-up.

---

## Summary of "nothing usable" (zero on-topic Commons results)

Studio One building, Treasure Isle, Black Ark, King Tubby's/Waterhouse,
Channel One, Tuff Gong's studio building, Harry J, Randy's Records,
Federal Records, Dynamic Sounds, Dischord House, The Masque, Scene Club,
Eel Pie Island's actual hotel/ballroom, Five Spot, Tractor Tavern, Milestone
Club.


