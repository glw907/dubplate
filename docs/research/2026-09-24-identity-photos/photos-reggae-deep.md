# Deep search: Jamaican/British reggae, ska, dub, dancehall period photos (1960-1985)

Scope: openly-licensed candidates for the dubplate login-page rotation, Jamaican/British
half only. Wikimedia Commons, Library of Congress, and Nationaal Archief/Anefo are covered
by another agent and excluded here even where they surfaced in results.

## Method notes / source access problems (read before trusting the "empty" list)

- **Openverse API** (`api.openverse.org`) returned normal JSON for the first ~15 queries,
  then started returning a Cloudflare "Just a moment..." JS challenge page for every
  subsequent request (both via `curl` and via the WebFetch tool, from this sandbox's egress
  IP), including on `openverse.org` itself. It never recovered across ~40 minutes of
  spaced-out retries. **Openverse coverage in this report is therefore partial**: only the
  first batch of queries (Studio One, Treasure Isle, King Tubby, Lee Perry, Coxsone Dodd,
  Duke Reid, Skatalites, Toots and the Maytals, Desmond Dekker, Prince Buster, Jimmy Cliff,
  Millie Small, Alton Ellis, Ken Boothe, John Holt) actually ran. Everything past that
  (Britain side entirely, plus Burning Spear, Culture, Congos, Big Youth, U-Roy, I-Roy,
  Dennis Brown, Gregory Isaacs, Peter Tosh, Bunny Wailer, Black Uhuru, Augustus Pablo,
  Yellowman, Eek-a-Mouse, sound systems, Notting Hill Carnival, 2 Tone, Rock Against Racism,
  skinhead scene) was **not** searchable via the Openverse API in this session and had to be
  covered by web search + direct Flickr page reads instead. A follow-up session from a
  different network should re-run Openverse for those terms.
- **Verifying "period" vs. "modern photo of a legacy artist"**: this is the single biggest
  trap in this domain. The great majority of Flickr/Openverse CC hits for Skatalites,
  Toots and the Maytals, Desmond Dekker, Prince Buster, Millie Small, Alton Ellis, Ken
  Boothe, John Holt, Jimmy Cliff, etc. are amateur concert photography from 2000s-2020s
  reunion/legacy tours and festivals (Rototom Sunsplash, club dates, festival stages), or
  album-cover rephotographs and storefront snapshots taken decades after the fact. I opened
  and checked upload/EXIF-adjacent metadata on the most promising-looking of these
  (titles like "1971-John-Holt-Still_in_chains", "Prince Buster's Studio") and confirmed
  they were shot in 2007-2009, not 1971. None of the Openverse batch-1 results qualify as
  a period (1960-1985) photograph. They are listed below under "checked, rejected" rather
  than omitted, so the next pass doesn't re-spend budget on them.
- **UK archival photography of this scene is a closed shop.** Every UK-side institutional
  lead this pass found (Urbanimage/Adrian Boot via London Museum, Dennis Morris via Hackney
  Museum, David Hoffman/British Culture Archive, Syd Shelton for Rock Against Racism, Kate
  Simon) is explicitly "all rights reserved" on the object page, quoted below. This matches
  the well-known reality that the canonical British reggae/2 Tone/RAR photo archives are
  commercially licensed (Getty, Urbanimage, PYMCA, the photographers' own sites) and none of
  them offer a CC or NC carve-out.
- **thereggaemuseum.com**: fetched the homepage and several linked collection pages
  (Collections & Archives, Cultural Artifacts Collection, Artists Archive, Ska Era
  Collection). No copyright notice, Creative Commons badge, public-domain statement, or
  terms-of-use language appears anywhere on the site as fetched. Treat every image on the
  site as all-rights-reserved by default (standard web copyright) unless a specific page
  states otherwise; none checked stated otherwise. Not usable as a source.
- **Europeana and Digital Commonwealth** returned HTTP 403 (Europeana) or a genuine
  "no items found" (Digital Commonwealth, query `reggae OR ska OR notting hill carnival`)
  when fetched directly; DPLA's and Smithsonian's non-API search pages returned empty or
  403 too. These are **inconclusive**, not confirmed-empty — a session with API keys (DPLA,
  Smithsonian `api.data.gov`) or a real browser (Europeana is bot-gated) should re-run them
  before calling the subject truly empty. British Library was not reachable in a useful form
  either (its main search UI didn't yield an image-search endpoint that returned content
  under these queries).
- **Internet Archive** `advancedsearch.php` (JSON, no auth, worked fine) only surfaces record
  scans (7" single labels/sleeves) for "reggae"/"ska jamaica" image-mediatype queries, not
  photographs of people, events, or venues. No candidates from IA.

---

## thereggaemuseum.com (owner's specific ask)

Checked the homepage and the linked gallery/collection pages (Collections & Archives,
Cultural Artifacts Collection, The Reggae Fashion Museum, Artists Archive, Ska Era
Collection, the Beth Lesser feature post). **No Creative Commons, CC0, public-domain, or
explicit reuse/terms-of-use statement was found anywhere on the site.** The site is
"Curated by YardRock TV, Operated by Reggae Arts & Culture Foundation Inc." with no stated
license on any image page fetched. Its Beth Lesser feature post links back to the same
reggae-vibes.com CC BY-NC-ND 4.0 "Witness to Reggae" pages documented below — i.e. the one
open-licensed material associated with the museum's own content is actually hosted
elsewhere, not on thereggaemuseum.com itself. Conclusion: **not a usable open-license source
under current site content.** (Outreach to the museum is being handled by another agent per
the coordinator's note; this is a licensing-facts-only finding.)

---

## Major find: Beth Lesser's "Witness to Reggae" archive (reggae-vibes.com)

**License, quoted from every page checked:** "This work is licensed under a Creative
Commons Attribution-NonCommercial-NoDerivatives 4.0 International License" (CC BY-NC-ND
4.0). Copyright line on every page: "All images © 1982-1988 Beth Lesser." **Class: NC + ND**
— flagged per the brief; a caption-plus-tone treatment likely counts as a derivative under
ND, so legal use may be limited to as-is display with attribution, no cropping/color
treatment, and non-commercial only.

Photographer Beth Lesser and her husband Dave Kingston travelled Toronto-Kingston-Brooklyn
through the 1980s documenting the roots-to-dancehall transition first-hand: sound systems,
studios, deejays, and street/dance scenes. This is the single richest vein of genuinely
period, on-the-ground Jamaican dancehall photography found under any open license in this
pass. Caveat: the on-page images are lazy-loaded/served only up to **479px on the long
edge** (checked raw `<img>`/`srcset` markup: `...jpg 479w` is the largest size served); the
site's download-request form (WeTransfer, gated behind agreeing to the CC terms) may offer
higher resolution but this was not verified. **Fit score capped at 2** for every entry below
purely on the size rule, pending confirmation of the WeTransfer resolution.

Index: https://www.reggae-vibes.com/beth-lesser-witness-to-reggae-photography/ (22 subject
pages, all same license/copyright block). Relevant to the named subject list:

| Subject | Year(s), verified from page | URL | Fit* |
|---|---|---|---|
| King Jammy (Prince Jammy) — studio, mixing desk, sound-system crew, Bobby Digital, Junior Reid, Half Pint, Cocoa Tea, Pinchers | 1984-1987 | https://www.reggae-vibes.com/witness-to-reggae-king-jammy-1/ (+ #2, #3, #4) | 3 (studio/crew shots), size-capped 2 |
| U-Roy — home, Stur Gav Crew, with other performers | 1985-1987 | https://www.reggae-vibes.com/witness-to-reggae-u-roy/ | 3, capped 2 |
| I-Roy — 1984 at Channel One studio; 1985 with David Rodigan and King Jammy; 1987 portrait | 1984-1987 | https://www.reggae-vibes.com/witness-to-reggae-i-roy/ | 3 (Channel One studio shot), capped 2 |
| Gregory Isaacs — with Leroy Smart, Junie; his record store | 1983, one 1986 | https://www.reggae-vibes.com/witness-to-reggae-gregory-isaacs/ | 3, capped 2 |
| Yellowman — with Son Walters, at Volcano Soundsystem, at CKLN-FM | 1983-1984 | https://www.reggae-vibes.com/witness-to-reggae-yellowman/ | 4 (sound-system/crowd shots), capped 2 |
| Youth Promotion / Sugar Minott — sound system, Waterhouse street scenes, Studio One | 1982-1988 (era, not itemized per-photo on this page) | https://www.reggae-vibes.com/witness-to-reggae-sugar-minott-youth-promotion/ | 4 (sound system/street), capped 2 |
| Gemini Sound (dancehall venue) — deejays Welton Irie, Squiddley Ranking, Michael Palmer performing to a crowd | 1984-1985 | https://www.reggae-vibes.com/witness-to-reggae-gemini-sound/ | 4 (crowd/dance), capped 2 |
| Tenor Saw at Youth Promotion, with Sugar Minott, Wayne Smith, General Leon, Pad Anthony | 1982-1988, mostly 1985 | https://www.reggae-vibes.com/witness-to-reggae-tenor-saw/ | 3, capped 2 |

Other pages in the same series exist for artists outside the named subject list (Horace
Ferguson, King Everald & Pinchers, Jah Stitch, Horace Andy, Ernest Wilson, Barry Brown,
Super Cat, Michael Prophet, Barrington Levy, Early B, U Brown) — same license, same 1982-88
window, same 479px cap; not detailed individually here but available at the index URL above
if the rotation wants more dancehall-culture variety.

\* Fit score is capped at 2 for all of these solely by the pixel-size rule in the brief
(under 1000px long edge). Composition-wise several (Gemini Sound, Youth Promotion, Yellowman
at Volcano) would otherwise score 4 for genuine sound-system/crowd energy.

---

## Subject-by-subject results

### Studios: Studio One, Treasure Isle, Black Ark, King Tubby's, Channel One, Randy's, Harry J
- Openverse "Treasure Isle studio Jamaica", "Lee Perry Black Ark", "Duke Reid producer",
  "Channel One studio Kingston", "Harry J studio reggae", "Randy's studio Jamaica" (batch 2,
  never ran due to the Cloudflare block) — **not searched**, follow-up needed.
- Openverse "Studio One Kingston" (batch 1) and "King Tubby" (batch 1) returned only
  Wikimedia-Commons items (excluded, other agent) and modern Flickr concert/tribute photos
  of King Tubby's Hi-Fi as a modern reissue soundsystem brand (Rototom Sunsplash festival,
  2010s, NC-SA) — not period, rejected.
- Black Ark: found one candidate, a print at the Smithsonian's National Museum of African
  American History and Culture, "Photographic print of Lee 'Scratch' Perry's Black Ark
  Studio" (nmaahc.si.edu/object/nmaahc_2015.132.165.1-.2), dated 1978 per the search snippet.
  **Rights unverified** — the object page returned HTTP 403 to WebFetch every time; NMAAHC
  objects require an explicit Open Access / CC0 tag to be usable and most are not. **Do not
  use without a human checking that page directly for an Open Access badge.**
- Beth Lesser's I-Roy page includes a genuine 1984 shot inside Channel One studio (see
  table above) — the one confirmed, licensed studio-interior candidate from this pass.
- Studio One, Treasure Isle, Randy's, Harry J: **empty** after this pass (Studio One partly
  unchecked per above; the rest genuinely turned up nothing beyond Wikipedia text).

### Producers: Coxsone Dodd, Duke Reid, Lee Perry, King Tubby, Prince Jammy, Joe Gibbs, Bunny Lee
- Prince Jammy: covered extensively via Beth Lesser (King Jammy #1-4), NC-ND, 1984-1987.
- Coxsone Dodd, Duke Reid, Lee Perry, King Tubby, Joe Gibbs, Bunny Lee: web search and (for
  Dodd/Reid) Openverse batch 1 turned up nothing but modern statue/mural photos (Coxsone
  Dodd corner-marker photos in Brooklyn/Kingston, CC BY-SA 2.0/4.0, but taken in the 2010s-
  2020s, not period) and all-rights-reserved museum/press material. **Empty** for genuine
  period photos of these five producers.

### Performers: Skatalites, Toots and the Maytals, Desmond Dekker, Prince Buster, Jimmy Cliff, Millie Small, Alton Ellis, Ken Boothe, John Holt
- All eight were queried on Openverse (batch 1, worked). Every hit is either Wikimedia
  Commons (excluded), a modern (2000s-2020s) concert/tribute photo, or a scan of a 2007-2009
  Flickr upload mislabeled with the recording year (verified by opening two of them — see
  Method notes). **No genuine period photo found for any of these eight** in this pass.

### Performers: Burning Spear, Culture, The Congos, Big Youth, U-Roy, I-Roy, Dennis Brown, Gregory Isaacs, Peter Tosh, Bunny Wailer, Black Uhuru, Augustus Pablo, Yellowman, Eek-a-Mouse
- U-Roy, I-Roy, Gregory Isaacs, Yellowman: covered via Beth Lesser (NC-ND, see table).
- Burning Spear, Culture, The Congos, Big Youth, Dennis Brown, Peter Tosh, Bunny Wailer,
  Black Uhuru, Augustus Pablo, Eek-a-Mouse: web search only (Openverse batch 2 blocked).
  Nothing but Getty/Alamy/museum all-rights-reserved hits and Wikipedia discography pages.
  **Empty** for all ten in this pass; Openverse itself was never actually queried for these
  names, so this is not a fully exhausted search — flagging for a retry, not a confirmed
  void.

### Sound systems, dances, and street/dancehall culture in Kingston
- Beth Lesser's Youth Promotion/Sugar Minott and Gemini Sound pages are strong, verified
  hits (NC-ND, 1982-88, genuine dance-floor/sound-system crowd shots). This is the best
  coverage of this subject in the whole search.
- No other open-licensed hits found (Openverse batch 2 for "sound system Kingston Jamaica
  dance", "Jamaica dancehall 1970s", "Kingston street scene 1970s" never ran — blocked).

### Britain: Jah Shaka, Sir Coxsone Outernational, Saxon sound systems
- Jah Shaka: one Flickr candidate (freestylee/5735468203, "Boom!") — checked directly:
  **"All rights reserved"**, and shot/uploaded 2011 regardless. Rejected.
- Sir Coxsone Outernational: London Museum holds "Coxsone International Sound System,"
  1980-1985, by Jean Bernard Sohiez — checked directly: **"© Jean Bernard Sohiez /
  Urbanimage,"** all rights reserved, no CC option; London Museum only offers commercial
  licensing through its Picture Library. Rejected.
- Saxon Studio International: no period photo found under any license.
- **Empty** for all three under open licenses.

### Notting Hill Carnival (1970s/early 1980s)
- The best-known 1970s/79 carnival photography (David Hoffman, via British Culture Archive)
  is explicitly all-rights-reserved ("no usage or reproduction of any kind is permitted
  without prior permission"). PYMCA (Richard Braine, 1975-76) is a commercial agency archive,
  not open-licensed. The only CC hits Wikimedia surfaces are recent-decade carnival photos
  (excluded anyway as Commons, and not period). **Empty**.

### Steel Pulse, Aswad, Matumbi, Misty in Roots, Linton Kwesi Johnson, UB40, Janet Kay/lovers rock
- Steel Pulse/Aswad: press photos credited to Adrian Boot and Dennis Morris (Island Records
  era) — no CC availability found.
- Linton Kwesi Johnson: Hackney Museum holds a 1979 Dennis Morris portrait ("The Dub Poet,
  the Voice of Brixton") — checked directly: **"© Dennis Morris www.dennismorris.com,"** all
  rights reserved. Rejected.
- Matumbi, Misty in Roots, Janet Kay/lovers rock, UB40 (period, i.e. not the 2011 Eva
  Rinaldi concert photos already seen in Openverse batch 1, which are modern and CC BY 2.0
  but not period): **empty**, nothing found under any license.

### 2 Tone: The Specials, The Selecter, The Beat, Madness, Bad Manners, The Bodysnatchers (1979-80 tour)
- Coverage is dominated by named professional photographers (Chalkie Davies, Erica
  Echenberg, Janette Beckman) whose work sits with Getty/Redferns/agencies, plus fan sites
  (Flashbak, Dangerous Minds, Coventry Music Museum) reproducing agency material without a
  reuse license. No CC/PD/NC source located for any of the six acts. **Empty**.

### Rock Against Racism / Victoria Park Carnival 1978, and the original 1969 skinhead/ska scene
- RAR: Syd Shelton is the definitive photographer; Art UK and Hackney Museum hold his and
  others' work, all under copyright/all-rights-reserved (confirmed via web search summaries,
  consistent with the two directly-checked museum pages above). **Empty**.
- 1969 skinhead/ska scene: nothing found beyond a single unrelated Wikimedia file (excluded)
  and SUBCULTZ.com's uncredited/rights-unclear photo essay (not checked for license, appears
  to be a personal fan site with no stated terms — treat as unusable without direct
  verification). **Effectively empty**.

---

## Fifteen best candidates overall

All fifteen are Beth Lesser's "Witness to Reggae" (reggae-vibes.com), because it is the only
source this pass confirmed as both genuinely period (1982-88) and openly licensed (CC
BY-NC-ND 4.0). All are capped at fit 2 for size (on-page images max out at 479px on the long
edge) despite several having 4-worthy composition. Ranked by composition/mood, size caveat
applying to every row:

1. Gemini Sound — deejays working a packed dancehall crowd, 1984-85 — https://www.reggae-vibes.com/witness-to-reggae-gemini-sound/
2. Yellowman @ Volcano Soundsystem, 1984 — https://www.reggae-vibes.com/witness-to-reggae-yellowman/
3. Youth Promotion / Sugar Minott — Waterhouse sound-system and street scenes, 1982-88 — https://www.reggae-vibes.com/witness-to-reggae-sugar-minott-youth-promotion/
4. King Jammy's yard — sound-system truck, crew, mixing desk, 1984-87 — https://www.reggae-vibes.com/witness-to-reggae-king-jammy-1/
5. Tenor Saw at Youth Promotion with Sugar Minott, Wayne Smith, General Leon, 1985 — https://www.reggae-vibes.com/witness-to-reggae-tenor-saw/
6. I-Roy at Channel One studio, 1984 — https://www.reggae-vibes.com/witness-to-reggae-i-roy/
7. I-Roy with David Rodigan and King Jammy, 1985 — https://www.reggae-vibes.com/witness-to-reggae-i-roy/
8. U-Roy with the Stur Gav Crew, 1985-87 — https://www.reggae-vibes.com/witness-to-reggae-u-roy/
9. Gregory Isaacs with Leroy Smart and Junie, 1983 — https://www.reggae-vibes.com/witness-to-reggae-gregory-isaacs/
10. King Jammy #2 studio/mixing-desk interior, 1984-87 — https://www.reggae-vibes.com/witness-to-reggae-king-jammy-2/
11. King Jammy #3 with Junior Reid, Half Pint, Cocoa Tea, Pinchers — https://www.reggae-vibes.com/witness-to-reggae-king-jammy-3/
12. King Jammy #4 — https://www.reggae-vibes.com/witness-to-reggae-king-jammy-4/
13. Yellowman and Dego at CKLN-FM Toronto, 1984 — https://www.reggae-vibes.com/witness-to-reggae-yellowman/
14. Yellowman and Son Walters, Toronto, 1983 — https://www.reggae-vibes.com/witness-to-reggae-yellowman/
15. Gregory Isaacs at his record store, 1986 — https://www.reggae-vibes.com/witness-to-reggae-gregory-isaacs/

No non-Beth-Lesser candidate cleared the license-plus-period bar in this pass. Nothing here
is CC0/PD/BY/BY-SA — the entire yield is NC+ND class.

---

## Subjects that stayed empty after an exhaustive search (this pass)

Fully exhausted (checked multiple sources, high confidence nothing open-licensed and period
exists, or the only leads found were confirmed all-rights-reserved on direct inspection):
- Studio One, Treasure Isle, Randy's, Harry J (studios)
- Coxsone Dodd, Duke Reid, Lee Perry, King Tubby, Joe Gibbs, Bunny Lee (producers, as
  individuals/portraits)
- Skatalites, Toots and the Maytals, Desmond Dekker, Prince Buster, Jimmy Cliff, Millie
  Small, Alton Ellis, Ken Boothe, John Holt (performers — Openverse fully queried, all
  results rejected as non-period)
- Jah Shaka, Sir Coxsone Outernational, Saxon (UK sound systems)
- Notting Hill Carnival 1976/1979/1980
- Steel Pulse, Aswad, Matumbi, Misty in Roots, Linton Kwesi Johnson, UB40, Janet Kay/lovers
  rock
- The Specials, The Selecter, The Beat, Madness, Bad Manners, The Bodysnatchers / 2 Tone
  1979-80
- Rock Against Racism / Victoria Park 1978
- Original 1969 multiracial skinhead/ska scene

Not exhausted — Openverse never got to run these before the block (flag for a retry from a
clean network/IP, or via a real browser session, before calling them empty):
- Burning Spear, Culture, The Congos, Big Youth, Dennis Brown, Peter Tosh, Bunny Wailer,
  Black Uhuru, Augustus Pablo, Eek-a-Mouse
- Generic "sound system Kingston Jamaica dance", "Jamaica dancehall 1970s", "Kingston street
  scene 1970s", "ska Jamaica 1964" queries
- Channel One, Randy's, Harry J as venues specifically (beyond the one Channel One interior
  shot found via Beth Lesser)
- Europeana, DPLA, Smithsonian Open Access, British Library, Digital Commonwealth as
  institutions generally — all returned 403/empty to unauthenticated fetches in this session
  and need either API keys or a real browser to search properly.
