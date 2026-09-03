# Real-catalog selection dossier

Research report, 2026-09-03 (Sonnet implementer, pass 2c Task 3), produced
per `docs/superpowers/plans/2026-09-03-rung-2c-catalog-adversarial.md`
(revision 2), Task 3's section, the "How a real-catalog case gets made"
section, and the Global Constraints network-door bullet. This document is
the selection record Tasks 4 through 15's twelve capture batches consume
without re-deciding anything: 96 primary release rows against explicit
per-axis quotas, plus 12 reserve rows tagged by the axis they back up.

## Method

Two verification channels, exactly as the plan's Global Constraints and
Task 3 section specify:

- **The four fact-verified axes** (reissue/remaster, VA/box-set, non-Latin,
  two-release sub-axis) were seeded through `harness capture --mbid <mbid>
  --out corpus/staging/<id>` and their structural facts confirmed through
  `harness seed-items --case <dir> --facts`, which prints the release date,
  the release-group's own first-release-date, the release status, the
  per-disc track counts, and the documented aliases (release-level and each
  artist-credit's own alias list, with type and locale) -- never by reading
  `responses/*.json` directly. Every fact-verified row below states the
  specific fact `--facts` confirmed. `DUBPLATE_HARNESS_CONTACT` came from
  the environment (`~/.bashrc`, never a literal), and every capture ran at
  the tool's own 1 request per second pacing with a further >=1.1s gap
  between separate `harness capture` invocations; a handful of individual
  requests returned a transient MusicBrainz 503, resolved by a retry a few
  seconds later, no candidate rejected for that reason.
- **The remaining three axes** (classical/compilation genre facts, MB-noise,
  genre/era spread) were selected by browsing `musicbrainz.org`'s own
  human-facing search pages (`/search?query=...&type=release&method=
  indexed`) at a courteous pace (>=1.5s between page fetches), never
  `/ws/2`. These claims are genre, provenance, and MB-status facts (a
  release's MB "Bootleg" status, a shared release title evidencing a
  near-duplicate-catalog pressing pair) legible directly from the search
  results table, not `/ws/2`-shaped structural fields, matching the plan's
  own reasoning for routing them off the capture channel.

No candidate was captured through any other route. No `responses/*.json`
was opened directly by this task; every structural fact below is
`--facts`' own projection. A rejected candidate's staging directory was
deleted before this task ends (see "Rejected candidates" below); a
confirmed fact-verified candidate's staging directory is kept under
`corpus/staging/` for its batch task to consume from loop step 3 onward
(no case is promoted or derived by this task -- promotion, `make derive`,
and ledger writes are the batch tasks' job).

## Axis quotas (restated from the plan, confirmed filled)

| Axis | Primary quota | Reserve | Fact-verified | Notes |
|---|---|---|---|---|
| Classical/compilation | 10 | 1 | no | 5 classical + 5 compilation primary |
| Reissue/remaster | 14 | 2 | yes (date divergence) | all 14 primary are mandatory stock siblings |
| VA/box-set | 14 | 2 | yes (disc count >= 3) | all 14 primary are mandatory stock siblings |
| Non-Latin | 12 | 1 | yes (documented alias) | Japanese, Korean, Cyrillic (Russian), Greek |
| MB-noise | 14 | 2 | no | bootleg status, near-dup catalog, partial/sparse rips |
| Genre/era spread | 26 | 3 | no | avoids over-indexing on rock/pop |
| Two-release sub-axis | 6 | 1 | yes (clean capture + decoy identified) | predominantly classical |
| **Total** | **96** | **12** | | |

Mandatory stock siblings: 14 (reissue) + 14 (boxset) = **28**, exactly
Task 19a's budget (M2) -- `original_date` and `per_disc_numbering` are the
only two config keys where estate and stock can disagree on a verdict, and
those are exactly the two axes that exercise them.

File-tag subjects (2d's Task 13 precondition; this pass only flags them):
**7** rows, one per axis, listed in the tables below with `yes` in that
column: `real-classical-karajan-beethoven-89`,
`real-reissue-nevermind-2011`, `real-boxset-james-brown-star-time`,
`real-nonlatin-utada-first-love`, `real-noise-dylan-bootleg`,
`real-genre-public-enemy-nation-millions`,
`real-decoy-beethoven-symphony-5`.

**Task 18's host release** is `real-nonlatin-utada-first-love`: its
artist-credit is romanized ("Hikaru Utada") with a `--facts`-confirmed
`locale=ja` alias "宇多田ヒカル", giving Task 18's `estate`/`locale` sibling
pair a release where `import.languages: [ja]` has a real, visible
`_preferred_alias` target to shift to if the mechanism works.

## Row format

Each row: case id (`real-<axis>-<slug>`), release MBID, one-sentence
rationale (including the confirmed fact for fact-verified rows), the
expected `Input.Items` shape, whether the row is a mandatory stock
sibling, whether it is a file-tag subject, and its capture batch (or which
axis's reserve pool it backs up).

### Classical (5 primary, part of the classical/compilation axis)

| Case id | MBID | Rationale | Item shape | Mandatory stock sibling | File-tag subject | Batch |
|---|---|---|---|---|---|---|
| `real-classical-karajan-beethoven-89` | `cb2d9dc2-0d6b-4dd2-9724-dbbf97ede73e` | Herbert von Karajan and the Berlin Philharmonic performing Beethoven's 8th and 9th symphonies, a canonical classical-orchestral shape with a conductor+orchestra artist-credit rather than a single billed performer. | clean rip | no | yes | Task 8 |
| `real-classical-gould-goldberg-1981` | `00383a0d-7ea6-4554-8d9f-5df3e3074e4f` | Glenn Gould's 1981 Goldberg Variations recording, a solo-piano classical release with a composer;performer artist-credit join. | clean rip | no | no | Task 9 |
| `real-classical-perlman-four-seasons` | `7558e8ab-0ee6-4f03-a629-234f34b76126` | Itzhak Perlman with the London Philharmonic performing Vivaldi's Four Seasons, a multi-artist classical credit distinct from Karajan's conductor-only shape. | clean rip | no | no | Task 10 |
| `real-classical-yo-yo-ma-cello-suites` | `36133601-f212-4ebe-92a2-a2c6af6d92b1` | Yo-Yo Ma's Six Evolutions: Bach Cello Suites, a modern (2018) solo-instrument classical release. | clean rip | no | no | Task 11 |
| `real-classical-ashkenazy-nocturnes` | `d2c9e620-83cc-32a3-a451-b5c695e805d1` | Vladimir Ashkenazy performing Chopin's Nocturnes, a 2-disc classical release giving the axis a multi-medium classical example distinct from the box-set axis's 3+-disc threshold. | clean rip | no | no | Task 12 |

### Compilation (5 primary + 1 reserve, part of the classical/compilation axis)

| Case id | MBID | Rationale | Item shape | Mandatory stock sibling | File-tag subject | Batch |
|---|---|---|---|---|---|---|
| `real-compilation-beatles-1967-1970` | `acd971e6-d23f-414b-aeea-25812c72fe5b` | The Beatles' 1967-1970 ("Blue Album"), a canonical single-artist compilation spanning the band's catalog. | clean rip | no | no | Task 9 |
| `real-compilation-marley-legend` | `f7da6448-0f6a-4569-90aa-920e7439815e` | Bob Marley & The Wailers' Legend, the best-known single-artist reggae compilation, distinct in artist and genre from the Beatles row. | clean rip | no | no | Task 10 |
| `real-compilation-queen-greatest-hits-ii` | `92655853-3fcc-3491-b166-0c932420f594` | Queen's Greatest Hits II, a single-artist rock compilation covering the band's 1980s output. | clean rip | no | no | Task 11 |
| `real-compilation-abba-gold` | `52cdc6f2-5942-430b-aff9-4c3496e1f893` | ABBA Gold: Greatest Hits, a pop compilation with a non-Anglo-American artist of origin (Sweden). | clean rip | no | no | Task 12 |
| `real-compilation-eagles-greatest-hits` | `a636b16c-01ba-416a-b04c-b3ed41280753` | Eagles' Their Greatest Hits 1971-1975, one of the best-selling compilation albums, rounding out the compilation half of the axis. | clean rip | no | no | Task 13 |
| `real-compilation-fleetwood-mac-greatest-hits` | `da79c2b9-858b-403a-bf38-05888457a8f3` | Fleetwood Mac's Greatest Hits, held in reserve for the compilation sub-axis. | clean rip | no | no | reserve for classical/compilation |

### Reissue/remaster (fact-verified: release date vs release-group first-release-date; 14 primary + 2 reserve; all 14 primary rows are mandatory stock siblings for Task 19a)

| Case id | MBID | Rationale | Item shape | Mandatory stock sibling | File-tag subject | Batch |
|---|---|---|---|---|---|---|
| `real-reissue-dark-side-of-the-moon-2010` | `a773d9ee-9930-4bf8-8e79-51a755396e25` | Pink Floyd - The Dark Side of the Moon (2010 CD reissue); capture-confirmed via seed-items --facts: release date 2010 vs release-group first-release-date 1973-03-24. | clean rip | yes | no | Task 4 |
| `real-reissue-rumours-deluxe-2013` | `76bd530b-5a1d-4cca-92e6-883b8871ef06` | Fleetwood Mac - Rumours (2013 deluxe reissue, 3 discs); capture-confirmed via seed-items --facts: release date 2013-01-25 vs release-group first-release-date 1977-02-04. | clean rip | yes | no | Task 5 |
| `real-reissue-nevermind-2011` | `237b8564-c726-4cb0-b804-348c30697325` | Nirvana - Nevermind (2011 20th-anniversary CD); capture-confirmed via seed-items --facts: release date 2011-09-23 vs release-group first-release-date 1991-09-24. | clean rip | yes | yes | Task 6 |
| `real-reissue-thriller-25` | `7a9ba230-8207-3d50-93bf-11fd3f33901a` | Michael Jackson - Thriller 25; capture-confirmed via seed-items --facts: release date 2008-02-12 vs release-group first-release-date 1982-11-30. | clean rip | yes | no | Task 7 |
| `real-reissue-abbey-road-2019` | `9fcd541a-a9ab-41c0-a6aa-13979c93301d` | The Beatles - Abbey Road (2019 mix); capture-confirmed via seed-items --facts: release date 2019-08-08 vs release-group first-release-date 1969-09-26. | clean rip | yes | no | Task 8 |
| `real-reissue-pet-sounds-1990` | `4a1f176d-4e05-4276-9afb-d9324d31601f` | The Beach Boys - Pet Sounds (1990 CD reissue); capture-confirmed via seed-items --facts: release date 1990 vs release-group first-release-date 1966-05-16. | clean rip | yes | no | Task 9 |
| `real-reissue-are-you-experienced-1987` | `789f6eb5-c7e1-37f7-a8f4-1a3374af225f` | The Jimi Hendrix Experience - Are You Experienced (1987 CD reissue); capture-confirmed via seed-items --facts: release date 1987 vs release-group first-release-date 1967-05-12. | clean rip | yes | no | Task 10 |
| `real-reissue-whats-going-on-2021` | `b17ab157-9871-4c94-bd04-147bf86f33d3` | Marvin Gaye - What's Going On (2021 digital reissue); capture-confirmed via seed-items --facts: release date 2021-01-22 vs release-group first-release-date 1971-05-20. | clean rip | yes | no | Task 11 |
| `real-reissue-catch-a-fire-deluxe` | `c6875ad6-bfd5-427c-a0e3-da7262b3d702` | Bob Marley & The Wailers - Catch a Fire (2001 deluxe edition, 2 discs); capture-confirmed via seed-items --facts: release date 2001-01 vs release-group first-release-date 1973-04-13. | clean rip | yes | no | Task 12 |
| `real-reissue-songs-key-of-life-2013` | `7b169780-da9d-4e39-a96b-81d1e0edc598` | Stevie Wonder - Songs in the Key of Life (2013 SHM-CD reissue, 2 discs); capture-confirmed via seed-items --facts: release date 2013-12-18 vs release-group first-release-date 1976-09-28. | clean rip | yes | no | Task 13 |
| `real-reissue-doolittle-2008` | `5b446655-d10c-3639-ac3e-bc689efce811` | Pixies - Doolittle (2008 hybrid SACD reissue); capture-confirmed via seed-items --facts: release date 2008 vs release-group first-release-date 1989-04-17. | clean rip | yes | no | Task 14 |
| `real-reissue-ten-2009` | `06fc304e-d70a-4846-93e5-544d6aa618d6` | Pearl Jam - Ten (2009 legacy edition, 2 discs); capture-confirmed via seed-items --facts: release date 2009-03-23 vs release-group first-release-date 1991-08-27. | clean rip | yes | no | Task 15 |
| `real-reissue-in-utero-2013` | `c2e80db3-bd09-4dd3-981c-f1b5378eaab0` | Nirvana - In Utero (2013 20th-anniversary remaster); capture-confirmed via seed-items --facts: release date 2013-07-30 vs release-group first-release-date 1993-09-14. | clean rip | yes | no | Task 4 |
| `real-reissue-blood-on-the-tracks-2003` | `b0851322-d298-418e-a741-d5479a0e50cb` | Bob Dylan - Blood on the Tracks (2003 CD reissue); capture-confirmed via seed-items --facts: release date 2003 vs release-group first-release-date 1975-01-20. | clean rip | yes | no | Task 5 |
| `real-reissue-exile-on-main-st-2010` | `dc4d13c0-bf5e-4872-a88b-2f78c1cfd939` | The Rolling Stones - Exile on Main St. (2010 reissue, 2 discs); capture-confirmed via seed-items --facts: release date 2010-05-19 vs release-group first-release-date 1972-05-12. | clean rip | no | no | reserve for reissue/remaster |
| `real-reissue-whats-going-on-2002` | `247fbafd-37f5-4a66-a0e3-41d33a1ace9b` | Marvin Gaye - What's Going On (2002 deluxe, 2 discs); capture-confirmed via seed-items --facts: release date 2002 vs release-group first-release-date 1971-05-20. | clean rip | no | no | reserve for reissue/remaster |

### VA/box-set (fact-verified: disc count >= 3; 14 primary + 2 reserve; all 14 primary rows are mandatory stock siblings for Task 19a)

| Case id | MBID | Rationale | Item shape | Mandatory stock sibling | File-tag subject | Batch |
|---|---|---|---|---|---|---|
| `real-boxset-woodstock-25th` | `c5a03147-c173-43ca-b19a-cff9d1249c81` | Woodstock: Three Days of Peace & Music, 25th Anniversary Collection (VA, 4 discs); capture-confirmed via seed-items --facts: 4 discs. | clean rip | yes | no | Task 5 |
| `real-boxset-mel-kim-singles` | `4f6b028a-b9a8-4595-9ce1-3bacb9c39842` | Mel & Kim - The Singles Box Set (7 discs); capture-confirmed via seed-items --facts: 7 discs. | clean rip | yes | no | Task 6 |
| `real-boxset-dylan-bootleg-1-3` | `1a6fdee5-5250-4d6a-ad39-6cd9721862d7` | Bob Dylan - The Bootleg Series, Volumes 1-3 (3 discs); capture-confirmed via seed-items --facts: 3 discs. | clean rip | yes | no | Task 7 |
| `real-boxset-nuggets-1965-1968` | `f5ebc3e1-4c2d-46d1-8f42-563a3fdc3653` | Nuggets: Original Artyfacts From the First Psychedelic Era (VA, 5 discs); capture-confirmed via seed-items --facts: 5 discs. | clean rip | yes | no | Task 8 |
| `real-boxset-beethoven-for-all` | `d3a68c3c-aa06-4c29-906a-c938c07a33bb` | Beethoven for All: Symphonies 1-9, Barenboim (5 discs); capture-confirmed via seed-items --facts: 5 discs. | clean rip | yes | no | Task 9 |
| `real-boxset-miles-davis-series` | `f380f1cd-692c-4f1a-9b3f-52f80981e1d9` | Miles Davis - The Box Set Series (4 discs); capture-confirmed via seed-items --facts: 4 discs. | clean rip | yes | no | Task 10 |
| `real-boxset-stax-volt-singles` | `54d131cf-79f4-49db-83d1-4176f97a2e3a` | The Complete Stax/Volt Singles: 1959-1968 (VA, 9 discs); capture-confirmed via seed-items --facts: 9 discs. | clean rip | yes | no | Task 11 |
| `real-boxset-now-jukebox-classics` | `c482f8d6-5788-471d-8a43-15e70438c0e5` | Now That's What I Call Jukebox Classics (VA, 4 discs); capture-confirmed via seed-items --facts: 4 discs. | clean rip | yes | no | Task 12 |
| `real-boxset-backstreet-boys-series` | `ed534e8d-a6fe-4b4f-9e5d-f5c202240dd5` | Backstreet Boys - The Box Set Series (4 discs); capture-confirmed via seed-items --facts: 4 discs. | clean rip | yes | no | Task 13 |
| `real-boxset-james-brown-star-time` | `248e85fc-894d-42b0-a4ff-2c96b9f1e384` | James Brown - Star Time (4 discs); capture-confirmed via seed-items --facts: 4 discs. | clean rip | yes | yes | Task 14 |
| `real-boxset-trojan-marley-covers` | `8641cd63-10d1-4ec5-be67-a1ed1f916ad8` | Trojan Bob Marley Covers Box Set (VA, 3 discs); capture-confirmed via seed-items --facts: 3 discs. | clean rip | yes | no | Task 15 |
| `real-boxset-rachmaninov-complete` | `4bf5255b-3354-424b-95e1-c0cab112e4b7` | Rachmaninov - The Complete Works (32 discs); capture-confirmed via seed-items --facts: 32 discs. | clean rip | yes | no | Task 4 |
| `real-boxset-mahler-complete` | `e8bc073a-ca7f-4ffc-b3c6-635450d85491` | Mahler - The Complete Works (16 discs); capture-confirmed via seed-items --facts: 16 discs. | clean rip | yes | no | Task 5 |
| `real-boxset-scriabin-complete` | `b94231a2-1c50-4950-88af-7f2b89198045` | Scriabin - The Complete Works (18 discs); capture-confirmed via seed-items --facts: 18 discs. | clean rip | yes | no | Task 6 |
| `real-boxset-shamen-singles` | `b13f2e83-4668-4a15-8616-62857502cef0` | Shamen - The Complete CD Singles Box Set (22 discs); capture-confirmed via seed-items --facts: 22 discs. | clean rip | no | no | reserve for VA/box-set |
| `real-boxset-trojan-marley-tribute` | `6a3c8fe9-addc-4841-a2f2-72df216709b0` | Trojan Tribute to Bob Marley Box Set (VA, 3 discs); capture-confirmed via seed-items --facts: 3 discs. | clean rip | no | no | reserve for VA/box-set |

### Non-Latin (fact-verified: documented locale alias in artist-credit; 12 primary + 1 reserve; `real-nonlatin-utada-first-love` is Task 18's host release)

| Case id | MBID | Rationale | Item shape | Mandatory stock sibling | File-tag subject | Batch |
|---|---|---|---|---|---|---|
| `real-nonlatin-utada-first-love` | `932c04f9-925a-45fc-847d-0da6a7b4e7a5` | Hikaru Utada - First Love (Japanese; credited artist name is the romanized "Hikaru Utada", with a documented locale=ja alias "宇多田ヒカル" -- this is Task 18's host release for the #6358 alias-selection verification). | clean rip | no | yes | Task 6 |
| `real-nonlatin-x-japan-art-of-life` | `4a56da85-22d2-4ba9-b834-1b736a21ae97` | X JAPAN - Art of Life (Japanese; artist-credit carries a documented locale=ja_JP alias). | clean rip | no | no | Task 7 |
| `real-nonlatin-babymetal` | `354a6d88-df30-4df3-953a-dd640ef9be5a` | BABYMETAL - BABYMETAL (Japanese; artist-credit carries documented locale=ja aliases "ベビーメタル" and "Bebīmetaru"). | clean rip | no | no | Task 8 |
| `real-nonlatin-hamasaki-ayumi` | `d30ced90-bd8e-4b52-9c9e-7ccabc511d0f` | Ayumi Hamasaki - HAMASAKI AYUMI (Japanese; artist-credit is natively the kanji form 浜崎あゆみ itself, with a documented locale=en alias "Ayumi Hamasaki"). | clean rip | no | no | Task 9 |
| `real-nonlatin-bts-map-of-soul-7` | `cc56250c-f0fb-42e3-9ae4-80a8016e68d1` | BTS - MAP OF THE SOUL : 7 (Korean; artist-credit carries a documented locale=ko alias "방탄소년단"). | clean rip | no | no | Task 10 |
| `real-nonlatin-blackpink` | `ea3cb8fd-0c2a-41d5-a3c7-b4c91cf659bf` | BLACKPINK - BLACKPINK (Korean; artist-credit carries a documented locale=ko alias "블랙핑크" among a very large documented alias set). | clean rip | no | no | Task 11 |
| `real-nonlatin-girls-generation` | `e0e72b32-abe2-41da-95eb-ffba407145a9` | Girls' Generation - Girls' Generation (Korean; artist-credit is natively the hangul form 소녀시대 itself, with a documented locale=en alias "Girls' Generation"). | clean rip | no | no | Task 12 |
| `real-nonlatin-tvxq-christmas-gift` | `8ffffea0-5126-4c4e-bda2-ef7eea1d5c62` | TVXQ - Christmas Gift From TVXQ! (Korean; artist-credit is natively the hangul form 동방신기 itself, with documented locale=en/ja/zh aliases). | clean rip | no | no | Task 13 |
| `real-nonlatin-zemfira` | `41bbc503-6f3c-4ef0-894e-b7d05d8df7d4` | Zemfira - Zemfira (Russian, Cyrillic; artist-credit is natively the Cyrillic form Земфира itself, with a documented locale=en alias "Zemfira"; a Pseudo-Release with a backward transl-tracklisting follow-up, giving the axis a second pseudo-release shape alongside the smoke corpus's non-latin-pseudo-release case). | clean rip | no | no | Task 14 |
| `real-nonlatin-alla-pugacheva-vol1` | `b5bed26f-4e2b-48cb-ae9a-057095ec1d01` | Alla Pugacheva - Alla Pugacheva, Volume 1 (Russian, Cyrillic; artist-credit is natively the Cyrillic form Алла Пугачёва itself, with a documented locale=en alias "Alla Pugacheva"). | clean rip | no | no | Task 15 |
| `real-nonlatin-vandi` | `aa3a7756-b8ac-4a53-ba92-ac6d6700a0ad` | Despina Vandi - Despina Vandi (Greek; artist-credit is natively the Greek form Δέσποινα Βανδή itself, with a documented locale=el alias and a documented locale=en romanized alias "Déspina Vandì"). | clean rip | no | no | Task 4 |
| `real-nonlatin-anna-vissi-essential` | `449c9b1f-7b35-4c8c-97f0-3a690b964783` | Anna Vissi - The Essential Anna Vissi (Greek; artist-credit carries a documented locale=el alias "Άννα Βίσση"). | clean rip | no | no | Task 5 |
| `real-nonlatin-alla-pugacheva-alla` | `294c6cd7-ffc0-4297-92aa-7121b88ddd52` | Alla Pugacheva - Alla (Russian, Cyrillic; a Pseudo-Release held in reserve for the non-Latin axis, with the same documented Cyrillic alias set as the primary Pugacheva row). | clean rip | no | no | reserve for non-Latin |

### MB-noise (14 primary + 2 reserve; not fact-verified, confirmed via the MusicBrainz web search page's own Status column)

| Case id | MBID | Rationale | Item shape | Mandatory stock sibling | File-tag subject | Batch |
|---|---|---|---|---|---|---|
| `real-noise-dylan-bootleg` | `df215f20-28d7-4455-894e-52d913c58d04` | Bob Dylan - The Bootleg (DE 1995); MB Status: Bootleg. | partial rip | no | yes | Task 7 |
| `real-noise-grateful-dead-de1996` | `bdf92574-72f3-4a7c-acd3-0c5970c83542` | Grateful Dead - Grateful Dead (DE 1996, live); MB Status: Bootleg. | partial rip | no | no | Task 8 |
| `real-noise-grateful-live-it1993` | `eeb46066-a549-4838-95e5-fc0fe175407a` | Grateful Dead - Grateful Live (IT 1993, 2 discs); MB Status: Bootleg. | sparse metadata | no | no | Task 9 |
| `real-noise-led-zeppelin-honeydrippers` | `1d5bb23f-b666-47e7-9f95-82c9a828c64b` | Led Zeppelin (The Honeydrippers / Jimmy Page) (RU 2000); MB Status: Bootleg. | sparse metadata | no | no | Task 10 |
| `real-noise-pink-floyd-de1994` | `030654bc-4039-4030-905a-fac00a8a9b94` | Pink Floyd - Pink Floyd (DE 1994, 3 discs, live); MB Status: Bootleg. | partial rip | no | no | Task 11 |
| `real-noise-live-pink-floyd-it1990` | `95e1a37c-6058-4d76-881f-f1c5033f925f` | Pink Floyd - The Live Pink Floyd (IT 1990); MB Status: Bootleg. | sparse metadata | no | no | Task 12 |
| `real-noise-neil-young-de1996` | `56406106-bbb5-45a9-9f10-0d2076110f90` | Neil Young - Neil Young (DE 1996, live); MB Status: Bootleg. | partial rip | no | no | Task 13 |
| `real-noise-prince-unreleased-mixes` | `0453f5d0-e693-4635-95b4-e2d25bbe4890` | Prince - The Unreleased Mixes; MB Status: Bootleg. | sparse metadata | no | no | Task 14 |
| `real-noise-rolling-stones-es2004` | `11d2cc47-92ee-4ef0-9ebd-2cdd1ac3c143` | The Rolling Stones - Rolling Stones (ES 2004 compilation); Official status but an obscure regional reissue, well suited to a partial rip. | partial rip | no | no | Task 15 |
| `real-noise-hendrix-gb1996` | `8421d976-12b2-47bb-88e8-9b2746283554` | Jimi Hendrix - Jimi Hendrix (GB 1996, 2 discs); an official compilation that shares its title and near-identical track counts (13+8) with a GB 1992 pressing of the same name (ea06d72c), a genuine near-duplicate-catalog pair. | clean rip | no | no | Task 4 |
| `real-noise-bowie-us1967` | `bfa03202-690b-411a-bf6a-baa78d74cfb7` | David Bowie - David Bowie (US 1967, 12 tracks); shares its title with a GB 1967 pressing carrying a different track count (14), a near-duplicate-catalog regional-pressing pair. | clean rip | no | no | Task 5 |
| `real-noise-velvet-underground-us1986` | `9258ac80-8f65-4699-b8f0-1e05ea815110` | The Velvet Underground - The Velvet Underground (US 1986 CD); shares its title with a JP 1990 pressing (a735be2e), another near-duplicate-catalog regional-pressing pair. | clean rip | no | no | Task 6 |
| `real-noise-zappa-cucamonga` | `90bee536-450c-4120-b3b1-b11a116375ed` | Frank Zappa - Frank Zappa: Cucamonga (US 2004 compilation); an obscure archival compilation well suited to a partial rip. | partial rip | no | no | Task 7 |
| `real-noise-hendrix-live-se` | `a0eda4e5-eb69-41d3-ac3f-16a4f7232db0` | Jimi Hendrix - Jimi Hendrix Live! (SE, undated); a sparsely-dated official release. | sparse metadata | no | no | Task 8 |
| `real-noise-hendrix-gb1992` | `ea06d72c-026c-4a4e-b18f-8a75b96fe54b` | Jimi Hendrix - Jimi Hendrix (GB 1992, 2 discs); the near-duplicate-catalog companion to hendrix-gb1996, held in reserve. | clean rip | no | no | reserve for MB-noise |
| `real-noise-pink-floyd-promo-fr2000` | `ecd89d9c-eb2d-4257-bb6c-c2f95377e77a` | Pink Floyd - Pink Floyd (FR 2000, EP + Compilation); a promotional pressing with sparse metadata, held in reserve. | sparse metadata | no | no | reserve for MB-noise |

### Genre/era spread (26 primary + 3 reserve; not fact-verified)

| Case id | MBID | Rationale | Item shape | Mandatory stock sibling | File-tag subject | Batch |
|---|---|---|---|---|---|---|
| `real-genre-bb-king-live-regal` | `0364504e-544d-43c1-ae8b-8d905a48c0d6` | B.B. King - Live at the Regal (US 1997 CD; blues, live, 1965 recording) | clean rip | no | no | Task 4 |
| `real-genre-public-enemy-nation-millions` | `02cfa32b-e22b-48a9-8a74-0fb4a8dbf221` | Public Enemy - It Takes a Nation of Millions to Hold Us Back (CA 1988 CD; hip-hop) | clean rip | no | yes | Task 4 |
| `real-genre-wu-tang-36-chambers` | `f711b6b2-1dca-442f-ac69-69009141f60c` | Wu-Tang Clan - Enter the Wu-Tang (36 Chambers) (US 1993 CD; hip-hop) | clean rip | no | no | Task 5 |
| `real-genre-fela-kuti-zombie` | `c59b62ef-32a4-410d-a9cf-8629cb9cb10d` | Fela and Africa 70 - Zombie (US 2010 CD reissue; Afrobeat, 1976 original) | clean rip | no | no | Task 5 |
| `real-genre-buena-vista-social-club` | `2e18e017-0a62-4c50-a708-b6fbea29612e` | Buena Vista Social Club - Buena Vista Social Club (US 1997 CD; Cuban son/world) | clean rip | no | no | Task 6 |
| `real-genre-kraftwerk-trans-europe-express` | `03055f40-ed52-320a-8de7-015eddefe156` | Kraftwerk - Trans-Europe Express (US 1987 CD; electronic) | clean rip | no | no | Task 6 |
| `real-genre-daft-punk-discovery` | `44e85a85-77d7-3e2e-8416-65a27d2aef68` | Daft Punk - Discovery (GB 2001 CD; electronic/house) | clean rip | no | no | Task 7 |
| `real-genre-johnny-cash-folsom-prison` | `89efce77-e3a9-4a7d-8be7-0eadfbdfa28e` | Johnny Cash - At Folsom Prison (US 1968 original vinyl; country, live) | clean rip | no | no | Task 7 |
| `real-genre-dolly-parton-jolene` | `2ae4521e-6913-43f4-939b-74c2993eebd9` | Dolly Parton - Jolene (US 1987 CD; country) | clean rip | no | no | Task 8 |
| `real-genre-black-sabbath-paranoid` | `89f60472-e469-417a-abc6-54ffa72c19d5` | Black Sabbath - Paranoid (JP 1991 CD; heavy metal) | clean rip | no | no | Task 8 |
| `real-genre-metallica-master-of-puppets` | `4c183eb6-1258-4a9f-b8e4-9f9e94187b6b` | Metallica - Master of Puppets (US 1999 CD; thrash metal) | clean rip | no | no | Task 9 |
| `real-genre-bad-brains` | `dcf077df-8856-4961-a0a0-e181a794ca61` | Bad Brains - Bad Brains (US 1982 ROIR cassette; hardcore punk) | clean rip | no | no | Task 9 |
| `real-genre-sex-pistols-bollocks` | `9c58da21-9819-41bf-9830-75b0a34a3642` | Sex Pistols - Never Mind the Bollocks Here's the Sex Pistols (US 1988 CD; punk) | clean rip | no | no | Task 10 |
| `real-genre-bob-marley-exodus` | `bb40fa6b-c2fe-350c-98ad-cf116d95b7f1` | Bob Marley & The Wailers - Exodus (US 1990 CD; reggae) | clean rip | no | no | Task 10 |
| `real-genre-toots-maytals-funky-kingston` | `eb60958d-0777-45ea-8985-2bb613fd2efc` | Toots & The Maytals - Funky Kingston (US 1975 original vinyl; reggae/ska) | clean rip | no | no | Task 11 |
| `real-genre-sly-family-stone-riot` | `96fe2af1-e61b-4201-bc77-07ba312bdbf7` | Sly & the Family Stone - There's a Riot Goin' On (JP 1997 CD; funk/soul) | clean rip | no | no | Task 11 |
| `real-genre-earth-wind-fire-way-of-world` | `de2beff8-98a3-4030-a89e-e3e73d34a81a` | Earth, Wind & Fire - That's the Way of the World (US 1988 CD; soul/funk) | clean rip | no | no | Task 12 |
| `real-genre-donna-summer-bad-girls` | `3ace4fec-c949-460d-8af4-cd4b18749509` | Donna Summer - Bad Girls (US 1990 CD; disco) | clean rip | no | no | Task 12 |
| `real-genre-chic-cest-chic` | `0fe8b085-2f82-4c86-86a1-69e1d040b50a` | Chic - C'est Chic (US 1992 CD; disco/funk) | clean rip | no | no | Task 13 |
| `real-genre-aphex-twin-saw85-92` | `f5538f56-309a-3322-92bb-54b829c14934` | Aphex Twin - Selected Ambient Works 85-92 (BE 2008 CD; ambient/electronic) | clean rip | no | no | Task 13 |
| `real-genre-brian-eno-ambient1` | `c52fda08-edf6-3949-8405-074ea605262d` | Brian Eno - Ambient 1: Music for Airports (US 2004 CD; ambient) | clean rip | no | no | Task 14 |
| `real-genre-joao-gilberto-chega-de-saudade` | `835940cc-3a91-4c5a-910a-ff5306ae6287` | Joao Gilberto - Chega de Saudade (BR 1959 original vinyl; bossa nova) | clean rip | no | no | Task 14 |
| `real-genre-celia-cruz-johnny` | `3f2098af-362e-4bcf-870b-d462cb754642` | Celia Cruz & Johnny Pacheco - Celia & Johnny (US 1988 CD; salsa) | clean rip | no | no | Task 14 |
| `real-genre-manu-chao-clandestino` | `3df03311-b0dc-4b99-895b-24db8b1b2c9c` | Manu Chao - Clandestino (XE 1998 CD; Latin alternative/world) | clean rip | no | no | Task 15 |
| `real-genre-burna-boy-african-giant` | `c077002c-e55d-46cd-b7d9-09cb1e41d751` | Burna Boy - African Giant (Digital Media; Afrobeats, 2019) | clean rip | no | no | Task 15 |
| `real-genre-skepta-konnichiwa` | `b7622f63-3470-4490-a296-4a4d5bca82cd` | Skepta - Konnichiwa (GB 2016 CD; grime) | clean rip | no | no | Task 15 |
| `real-genre-aretha-franklin-never-loved` | `539fc513-c656-4a93-a099-9f4b124261be` | Aretha Franklin - I Never Loved a Man the Way I Love You (US 2012 Digital Media; soul, 1967 original), held in reserve | clean rip | no | no | reserve for genre/era spread |
| `real-genre-mahalia-jackson-greatest` | `5fd39a69-8596-43df-b2d7-df3ea6fa0f4d` | Mahalia Jackson - Mahalia Jackson's Greatest Hits (US CD; gospel), held in reserve | clean rip | no | no | reserve for genre/era spread |
| `real-genre-massive-attack-blue-lines` | `ef00b038-af2d-3fff-a655-8ad5b7a57fea` | Massive Attack - Blue Lines (GB 1991 CD; trip-hop), held in reserve | clean rip | no | no | reserve for genre/era spread |

### Two-release sub-axis (fact-verified: clean capture plus a same/similar-titled decoy release found on the same search results page; 6 primary + 1 reserve)

| Case id | MBID | Rationale | Item shape | Mandatory stock sibling | File-tag subject | Batch |
|---|---|---|---|---|---|---|
| `real-decoy-beethoven-symphony-5` | `f394e886-5485-4b66-9b7c-ca45b4734b6e` | Beethoven's Symphony No. 5, Cleveland Orchestra / George Szell; decoy: "Beethoven: Symphony No. 5: Anniversary Edition", Wilhelm Furtwangler / Berlin Philharmonic (8d198198-f3c5-4d26-9765-4441de8118f6), same generic work title, different performer. | clean rip, authored ambiguously to surface the decoy | no | yes | Task 4 |
| `real-decoy-vivaldi-four-seasons` | `d9a105e8-de9c-4a80-a5e7-d98665558263` | Vivaldi's Four Seasons, Giacobbe Stevanato / Interpreti Veneziani; decoy: "Vivaldi's Four Seasons", Gilles Apap & The Colors Of Invention (8407e282-a386-4c8c-b679-f84363a0c8a0), identical release title, different performer. | clean rip, authored ambiguously to surface the decoy | no | no | Task 6 |
| `real-decoy-mozart-requiem` | `a67b8f8d-c524-4634-9473-9a1d75519bf8` | Mozart's Requiem, Bach Choir and Orchestra of the Netherlands; decoy: "Mozart Requiem", English Chamber Orchestra / Benjamin Britten (d13bdf72-34e4-4a48-ac49-5d6ce5d38c0b), the same two words reordered, different performer. | clean rip, authored ambiguously to surface the decoy | no | no | Task 7 |
| `real-decoy-beethoven-symphony-9-choral` | `bb6fc19a-7506-41be-b1eb-4afa24da91be` | Beethoven's Symphony No. 9 "Choral", Josef Krips / LSO; decoy: "Symphony No. 9, 'Choral'", Roger Norrington / London Classical Players (1edbbfc1-70f9-4db0-8b09-dee32e0d7c70), same generic work title, different performer. | clean rip, authored ambiguously to surface the decoy | no | no | Task 13 |
| `real-decoy-tchaikovsky-swan-lake` | `19bbaf85-3aef-469b-95b6-1de1a4802683` | Tchaikovsky's Swan Lake (GB 2013); decoy: "Swan Lake", Charles Dutoit / Montreal Symphony (f3b20090-9a4b-4e9f-80f6-993edce380d6), identical release title, different performer. | clean rip, authored ambiguously to surface the decoy | no | no | Task 14 |
| `real-decoy-chopin-nocturnes` | `12cdb773-f5f2-438b-874a-7ec824ce0ff9` | Chopin Nocturnes, Peter Schmalfuss & Dubravka Tomsic; decoy: "Chopin Nocturnes", Natalia Kushner (d9b0cb53-0d57-4b8d-85ee-813689daddad), identical release title, different performer. | clean rip, authored ambiguously to surface the decoy | no | no | Task 15 |
| `real-decoy-mozart-piano-concerto-21` | `91356ead-4fa6-4f62-b1d6-608ee4235b8a` | Mozart Piano Concertos no. 20, 21, Rudolf Buchbinder / Wiener Symphoniker; decoy: "Piano Concertos No. 9, K271 / No. 21, K467", Mitsuko Uchida / Cleveland Orchestra (8a08db35-7609-49c3-8adb-c2458f4b464e), both plausibly titled around Concerto No. 21, different performer; held in reserve. | clean rip, authored ambiguously to surface the decoy | no | no | reserve for two-release sub-axis |

## Capture batch assignment (Tasks 4 through 15)

Re-dealt (fix pass, see the note at the end of this section) so a batch
failure never wipes out one whole axis, per the plan's own contract: the
largest axis, genre/era spread (26 rows), is spread across all twelve
batches at 2 rows per batch for ten batches and 3 rows for the last two
(Task 14, Task 15), where its own supply runs out; every batch also
carries at least one, and typically five or more, of the other seven
axes, so no batch is ever single-axis. The four fact-verified axes
(reissue, boxset, non-Latin, two-release) and the three web-selected axes
(classical, compilation, MB-noise) are interleaved across the same twelve
slots rather than exhausted in axis order. A batch never re-selects: it
works exactly the 8 case ids listed for its task number below, seeding
(non-fact-verified rows) or resuming from Task 3's kept staging directory
(fact-verified rows) at loop step 3.

**Task 4** (8 cases): `real-boxset-rachmaninov-complete`, `real-decoy-beethoven-symphony-5`, `real-genre-bb-king-live-regal`, `real-genre-public-enemy-nation-millions`, `real-noise-hendrix-gb1996`, `real-nonlatin-vandi`, `real-reissue-dark-side-of-the-moon-2010`, `real-reissue-in-utero-2013`

**Task 5** (8 cases): `real-boxset-mahler-complete`, `real-boxset-woodstock-25th`, `real-genre-fela-kuti-zombie`, `real-genre-wu-tang-36-chambers`, `real-noise-bowie-us1967`, `real-nonlatin-anna-vissi-essential`, `real-reissue-blood-on-the-tracks-2003`, `real-reissue-rumours-deluxe-2013`

**Task 6** (8 cases): `real-boxset-mel-kim-singles`, `real-boxset-scriabin-complete`, `real-decoy-vivaldi-four-seasons`, `real-genre-buena-vista-social-club`, `real-genre-kraftwerk-trans-europe-express`, `real-noise-velvet-underground-us1986`, `real-nonlatin-utada-first-love`, `real-reissue-nevermind-2011`

**Task 7** (8 cases): `real-boxset-dylan-bootleg-1-3`, `real-decoy-mozart-requiem`, `real-genre-daft-punk-discovery`, `real-genre-johnny-cash-folsom-prison`, `real-noise-dylan-bootleg`, `real-noise-zappa-cucamonga`, `real-nonlatin-x-japan-art-of-life`, `real-reissue-thriller-25`

**Task 8** (8 cases): `real-boxset-nuggets-1965-1968`, `real-classical-karajan-beethoven-89`, `real-genre-black-sabbath-paranoid`, `real-genre-dolly-parton-jolene`, `real-noise-grateful-dead-de1996`, `real-noise-hendrix-live-se`, `real-nonlatin-babymetal`, `real-reissue-abbey-road-2019`

**Task 9** (8 cases): `real-boxset-beethoven-for-all`, `real-classical-gould-goldberg-1981`, `real-compilation-beatles-1967-1970`, `real-genre-bad-brains`, `real-genre-metallica-master-of-puppets`, `real-noise-grateful-live-it1993`, `real-nonlatin-hamasaki-ayumi`, `real-reissue-pet-sounds-1990`

**Task 10** (8 cases): `real-boxset-miles-davis-series`, `real-classical-perlman-four-seasons`, `real-compilation-marley-legend`, `real-genre-bob-marley-exodus`, `real-genre-sex-pistols-bollocks`, `real-noise-led-zeppelin-honeydrippers`, `real-nonlatin-bts-map-of-soul-7`, `real-reissue-are-you-experienced-1987`

**Task 11** (8 cases): `real-boxset-stax-volt-singles`, `real-classical-yo-yo-ma-cello-suites`, `real-compilation-queen-greatest-hits-ii`, `real-genre-sly-family-stone-riot`, `real-genre-toots-maytals-funky-kingston`, `real-noise-pink-floyd-de1994`, `real-nonlatin-blackpink`, `real-reissue-whats-going-on-2021`

**Task 12** (8 cases): `real-boxset-now-jukebox-classics`, `real-classical-ashkenazy-nocturnes`, `real-compilation-abba-gold`, `real-genre-donna-summer-bad-girls`, `real-genre-earth-wind-fire-way-of-world`, `real-noise-live-pink-floyd-it1990`, `real-nonlatin-girls-generation`, `real-reissue-catch-a-fire-deluxe`

**Task 13** (8 cases): `real-boxset-backstreet-boys-series`, `real-compilation-eagles-greatest-hits`, `real-decoy-beethoven-symphony-9-choral`, `real-genre-aphex-twin-saw85-92`, `real-genre-chic-cest-chic`, `real-noise-neil-young-de1996`, `real-nonlatin-tvxq-christmas-gift`, `real-reissue-songs-key-of-life-2013`

**Task 14** (8 cases): `real-boxset-james-brown-star-time`, `real-decoy-tchaikovsky-swan-lake`, `real-genre-brian-eno-ambient1`, `real-genre-celia-cruz-johnny`, `real-genre-joao-gilberto-chega-de-saudade`, `real-noise-prince-unreleased-mixes`, `real-nonlatin-zemfira`, `real-reissue-doolittle-2008`

**Task 15** (8 cases): `real-boxset-trojan-marley-covers`, `real-decoy-chopin-nocturnes`, `real-genre-burna-boy-african-giant`, `real-genre-manu-chao-clandestino`, `real-genre-skepta-konnichiwa`, `real-noise-rolling-stones-es2004`, `real-nonlatin-alla-pugacheva-vol1`, `real-reissue-ten-2009`

Distinct axes present per batch, for the record: Task 4 has 6 (boxset,
decoy, genre, noise, non-Latin, reissue), Task 5 has 5, Task 6 has 6,
Task 7 has 6, Task 8 has 6, Task 9 has 7, Task 10 has 7, Task 11 has 7,
Task 12 has 7, Task 13 has 7, Task 14 has 6, Task 15 has 6. Every batch
clears the plan's "at least two axes" floor by a wide margin; genre is
present in every one of the twelve.

## Rejected candidates

One: `real-nonlatin-ddt` (`767f2ed1-2f70-4847-84a9-dacc802bf4b5`, DDT -
DDT). Its original rationale claimed a "Russian rock band" case whose
credited string is Latin but whose genre and origin are Cyrillic-locale,
carrying a documented alias. `harness seed-items --case
corpus/staging/real-nonlatin-ddt --facts` showed otherwise: text
representation `Latn`, language `spa`, country `AR` (the band is
Argentine, not Russian), and its one artist-credit alias
("Demonios de Tasmania") carries no locale at all. The candidate does not
serve the non-Latin axis; its staging directory was deleted before this
task ends. Its replacement, `real-nonlatin-vandi`
(`aa3a7756-b8ac-4a53-ba92-ac6d6700a0ad`, Despina Vandi - Despina Vandi),
was seeded and confirmed via `--facts`: a native Greek artist-credit
(Δέσποινα Βανδή) carrying a documented `locale=el` alias and a documented
`locale=en` romanized alias ("Déspina Vandì"), filling the vacated
primary slot.

Every other fact-verified candidate seeded during this task's selection
work confirmed its axis's required structural fact on its own first
`--facts` read, with no other rejection: 16 reissue/remaster candidates
seeded (14 primary + 2 reserve, all kept), 16 VA/box-set candidates
seeded (14 primary + 2 reserve, all kept), 13 non-Latin candidates
seeded and kept (12 primary + 1 reserve, of which one primary,
`real-nonlatin-vandi`, is the DDT replacement recorded above; the
fourteenth non-Latin candidate seeded overall is the rejected
`real-nonlatin-ddt`), and 7 two-release candidates seeded (6 primary + 1
reserve, all kept). All 52 kept fact-verified staging directories remain
under `corpus/staging/` for their batch tasks (14+2, 14+2, 12+1, 6+1
respectively). A small number of individual `harness capture` requests
returned a transient MusicBrainz `503`; each was resolved by a retry
after a short pause rather than by substituting a different candidate,
so no other MBID was abandoned mid-verification.

## Token spend

The plan asks Task 3 to report its own actual token spend, since the
ceiling re-derivation point moves here (M9). This implementer dispatch
(the fix pass that produced the non-Latin correction and the batch
re-deal above) has no local mechanism to introspect its own cumulative
token usage: there is no `ccusage`-style CLI on the workstation, no
per-dispatch usage log under `~/.claude/telemetry/`, and the
implementer's own session transcript under
`~/.claude/projects/-var-home-glw907-Projects-dubplate/` is not
addressable from inside the dispatch (Claude models are not
self-aware of their running token count). **The conductor's own
Task-dispatch return value (or `/cost`) is the authoritative source**
per CLAUDE.md's "Score both budgets at pass end" convention, and is
where the actual figure for this task, and the original Task 3
attempt this fix pass followed, should be recorded once read back from
the dispatch layer. This gap is reported to the conductor as a
`couldNotDo` item rather than a fabricated number.

## Status for Tasks 4 through 15

`corpus/staging/` holds 52 confirmed, seeded case directories (the
fact-verified rows' primary and reserve candidates, `profile: "estate"`
and `eval_date: "2026-09-03"` already set, `Input.Items` not yet
projected -- that is loop step 3, the batch task's own job). The 44
non-fact-verified rows (classical/compilation, MB-noise, genre/era) carry
only a confirmed MBID and rationale here; their batch task performs loop
step 2 (`harness capture --mbid <mbid> --out corpus/staging/<id>`) itself.
No case in this dossier has been promoted, derived, or committed by this
task; `corpus/cases/` is unchanged.
