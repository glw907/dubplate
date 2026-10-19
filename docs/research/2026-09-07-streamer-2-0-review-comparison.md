# Review: is Navidrome the right comparison point for dubplate-server 2.0?

Fresh-context adversarial review of
`docs/specs/2026-09-07-streamer-2-0-design.md` (revision 1).
Lens: the choice of Navidrome as reference, oracle, and subsumed product.
All observations dated 2026-09-07 unless noted.

## Verdict

**Keep Navidrome as the behavioral reference and the primary oracle, and
keep OpenSubsonic first. Add two things the spec does not have: a
machine-readable contract tier above the oracle, and gonic as a narrow
second oracle for browse-by-folder.** Navidrome earns its place on
adoption alone: 23,396 stars against gonic's 2,538 and LMS's 1,673, and
every client that publishes compatibility notes names it first
([topic listing](https://github.com/topics/opensubsonic);
[Symfonium docs](https://docs.symfonium.app/wiki/providers/subsonic-opensubsonic-media-provider-configuration/)).
It also containerizes against a plain directory with no wizard, which
most of the field does not. But a single-server oracle enshrines that
server's choices, and Navidrome publishes a list of the places it
diverges from the Subsonic contract on purpose — including one,
browse-by-folder, that it fills with a *simulated* tree
([navidrome.org](https://www.navidrome.org/docs/developers/subsonic-api/)).
That is precisely the surface the spec plans to derive from real pipeline
paths (spec L146-148), so on that endpoint family Navidrome is the wrong
oracle and gonic, which serves an actual filesystem tree, is the right
one. Separately, two machine-readable contracts already exist — the
OpenSubsonic OpenAPI schema and the official Subsonic 1.16.1 XSD — and
Ampache already runs both in CI as a conformance test. A response can
match Navidrome and still violate the schema; only the schema tier
catches that. Jellyfin is not a reason to reconsider: its OpenSubsonic
bridges (JellySonic, jellysub) are archived and last targeted Jellyfin
10.8, and its own clients still lack gapless playback. Finally, the spec
should stop treating the oracle as a fixed artifact: Navidrome ships
near-monthly and changed transcoding, artwork, and search inside the last
six months, so the harness must pin a version and treat an upgrade as a
ledger event.

## Servers compared

Stars, issues, and commit-author distribution pulled from the GitHub API
2026-09-07. "Bus factor" is the share of the last 100 commits held by the
top author, a proxy, not a rights audit.

| Server | Lang | License | Last release | Cadence | Bus factor | Stars / open issues | OpenSubsonic | Oracle? |
|---|---|---|---|---|---|---|---|---|
| [Navidrome](https://github.com/navidrome/navidrome) | Go | GPL-3.0 | v0.63.2, 2026-07-11 | ~monthly | 1 (88/100 commits, deluan) | 23,396 / 281 (oldest 2020-04) | Broad; adds per release | **Excellent** — one image, `ND_MUSICFOLDER`, SQLite auto-bootstrap, no wizard |
| [gonic](https://github.com/sentriz/gonic) | Go | GPL-3.0 | v0.22.0, 2026-06-08 | sparse tags, trunk active to 2026-09-07 | 1 (90/100, sentriz) | 2,538 / 72 | "most extensions", ≥0.21.0 per Symfonium | **Excellent** — single binary or image, flags for music/db paths |
| [LMS](https://github.com/epoupon/lms) | C++ | GPL-3.0 | v3.80.0, 2026-08-19 | frequent | 1 (100/100, emeric) | 1,673 / 81 | Documented `SUBSONIC.md`; needs legacy auth | **Good** — lightweight, own DB |
| [Supysonic](https://github.com/spl0k/supysonic) | Python | AGPL-3.0 | 0.7.9, 2024-12-15 | ~yearly | 1 (87/100) | 288 / 12 | Targets base spec **1.10.2** only | Good but legacy-tier |
| [Melodee](https://github.com/melodee-project/melodee) | C# .NET 10 | **MIT** | 2.2.0, matrix dated 2026-07-11 | active (1,478 commits since 2024-10) | 1 (one copyright holder) | 65 / 1 | Own [matrix](https://melodee.org/opensubsonic-matrix/), 311 schema tests | Poor as oracle (Postgres + Blazor + staged ingest), **high value as readable prior art** |
| [AudioMuse-AI-MusicServer](https://github.com/NeptuneHub/AudioMuse-AI-MusicServer) | Go | AGPL-3.0 | active 2026-09 | new (2025-09) | 1 | 56 / 0 | Full OpenSubsonic claim | Runs in Docker on `/music`, but AGPL and young |
| [Ampache](https://github.com/ampache/ampache) | PHP | AGPL-3.0 | 8.0.1, 2026-08-17 | active | ~2 | 3,828 / 379 (oldest 2014-04) | **All 87 endpoints**, audited 2026-08-07 | Poor — MySQL, install wizard, LAMP stack |
| [Airsonic-Advanced](https://github.com/airsonic-advanced/airsonic-advanced) | Java | GPL-3.0 | v10.6.0, **2020-04-24** | stalled (last push 2024-04) | 1 | 1,404 / 275 | Base Subsonic; extensions unverified | Marginal; Symfonium calls it "not recommended" |
| [Jellyfin](https://github.com/jellyfin/jellyfin) + [JellySonic](https://github.com/lyarenei/JellySonic) | C# | GPL-2.0 | 10.11.11, 2026-06-06 | very active | many (foundation) | 56,633 / 591 | **Plugin archived 2022**, targeted JF 10.8 | Poor — mandatory wizard, dead bridge |
| [Funkwhale](https://dev.funkwhale.audio/funkwhale/funkwhale) | Python | AGPL-3.0 | 2.0.4, 2026-06-08 | 1-2/yr | unverified | unverified (GitLab) | Partial; Symfonium notes open issues | Poor — Postgres+Redis+Celery+nginx |
| [Koel](https://github.com/koel/koel) | PHP | MIT | 9.11.2, 2026-09-07 | continuous | 1 (75/100) | 17,241 / 17 | **None** | N/A |
| [Black Candy](https://github.com/blackcandy-org/blackcandy) | Ruby | MIT | 3.2.1, 2026-08-26 | active | 1 (96/100) | 4,409 / 33 | **None** | N/A |
| [Music Assistant](https://github.com/music-assistant/server) | Python | Apache-2.0 | 2.10.2, 2026-09-04 | very active | ~3 | 3,040 / 59 | Aggregator, not a server | N/A |
| [Astiga](https://asti.ga/) | — | proprietary SaaS | — | — | — | — | Partial; needs compat mode | **Not self-hostable** |
| Emby / Plex+Plexamp | C# / — | proprietary | — | — | — | — | Third-party shims only ([EmbySonic](https://github.com/trifleneurotic/EmbySonic), [Plexsonic](https://github.com/ClassicOldSong/Plexsonic)) | N/A |

Two structural readings. First, **bus factor 1 is the field's norm**, Navidrome
included; the spec's "subsumes Navidrome" framing is not betting against a
foundation. Second, the OpenSubsonic
[servers page](https://opensubsonic.netlify.app/docs/) (Ampache, Astiga,
EdgeSonic, gonic, LMS, Navidrome, Nextcloud Music, Qm-Music, Radiccio,
Supysonic) lags the ecosystem; Melodee and AudioMuse are absent from it.

## Clients

The roster decides the oracle, so the roster's health matters.

| Client | Platform / license | Maintained | Reference server | Notes |
|---|---|---|---|---|
| [Symfonium](https://docs.symfonium.app/wiki/providers/subsonic-opensubsonic-media-provider-configuration/) | Android, proprietary, paid | v14.1.0 ~2026-04 | **Navidrome ≥0.51.1 and gonic ≥0.21.0 named as the two "actively maintained, fast sync" servers**; LMS via legacy auth; Subsonic and Airsonic "no longer maintained, not recommended" | The ecosystem's only real compatibility matrix; runs an extensions forum |
| [Feishin](https://github.com/jeffvli/feishin) | desktop/web, GPLv3 | v1.15.1, 2026-07 | Navidrome first; also Jellyfin, gonic, LMS, Ampache | Successor to Sonixd; uses Song Lyrics v2 (needs Navidrome ≥0.63) |
| [Supersonic](https://github.com/dweymouth/supersonic) | desktop Go/Fyne, GPLv3 | active, 2,300 stars | Navidrome, Jellyfin, gonic, LMS | Multi-valued artist/genre extensions since 0.6.0 |
| [Substreamer](https://github.com/ghenry22/substreamer) | Android+iOS, GPLv3 | v8.0.91, 2026-07 | Navidrome "recommended, full OpenSubsonic"; gonic "some features limited" | ~68k Play installs (unverified) |
| **Tempus** (eddyizm/tempus) | Android, GPLv3 | v4.26.1, **2026-09-02** | Navidrome | The live fork; **original Tempo is dead** (v3.9.0, Dec 2024, maintainer unresponsive) |
| [Ultrasonic](https://gitlab.com/ultrasonic/ultrasonic) | Android, GPLv3 | v4.9.0, 2026-03 | Subsonic, Airsonic-Adv, Supysonic, Ampache | **Moved to GitLab**; GitHub repo archived 2022 |
| DSub2000 (paroj) | Android, GPLv3 | v5.7.4, 2026-05 | generic Subsonic | Original DSub abandoned 2022 |
| [Amperfy](https://github.com/BLeeEZ/amperfy) | iOS/macOS, GPLv3 | last tag v2.1.1 (2024-08); store shows 2026-08 | Ampache/Subsonic generic | 1.7k stars |
| play:Sub | iOS/CarPlay, paid | 2026-07 | Navidrome, Airsonic | 4.5★ / 424 |
| [airsonic-refix](https://github.com/tamland/airsonic-refix) | web UI, AGPLv3 | 2026-04 | Airsonic-Adv, Navidrome, gonic | Web UI, not an app |
| Cassette (iOS/macOS, MPL-2.0), Nautiline (iOS, paid, Dec 2025), Firmium (Rust+Kotlin) | — | new 2025-26 | Navidrome | Newcomers all target Navidrome |
| **Dead**: Sonixd (archived 2024-08), Audinaut (2022), Sublime Music (2024-12), Aurial (2020) | — | — | — | Do not put these in a roster |
| [Finamp](https://github.com/jmshrv/finamp) (Jellyfin) | Flutter, MPL-2.0 | 1.0.1-beta, 2026-09-02 | Jellyfin native | 4.2k stars; strong, but one app against ~20 Subsonic clients |

Every actively maintained client either names Navidrome as its reference or
names it first among several. That confirms the spec's choice of reference.
It also means a Navidrome-only oracle will bake in Navidrome behavior that
the *clients* never actually depend on, which is what the schema tier and
the second oracle exist to separate.

## Findings, ranked

**1. Add a contract tier above the oracle: the OpenSubsonic OpenAPI schema
and the Subsonic 1.16.1 XSD.** (Spec L101-108, L257-262, prior-art table
L307-323.) The spec's evidence rests entirely on "does dubplate match
Navidrome". Two machine-readable contracts exist:
[`openapi.json`](https://opensubsonic.netlify.app/docs/openapi/) covering
~100 endpoints (JSON only; carries a self-declared "work still ongoing"
caveat), and the official
`http://www.subsonic.org/pages/inc/api/schema/subsonic-rest-api-1.16.1.xsd`.
Ampache already runs exactly this as CI —
`OpenSubsonicSpecConformanceTest.php` validates JSON against the schema and
XML against the 1.16.1 XSD, and a sibling test checksums the vendored spec
copy so a refresh fails CI
([ampache.org](https://ampache.org/api/subsonic/), audit dated 2026-08-07).
**If accepted**: rung 2's deliverable list gains a schema validator beside
the differ; the prior-art table gains rows for the OpenAPI schema, the XSD,
and Ampache's conformance test as a method exemplar; a divergence gains a
third classification (dubplate wrong / Navidrome extra-spec / spec silent).

**2. Navidrome is the wrong oracle for browse-by-folder; use gonic there.**
(Spec L146-148, L101-108, L226-231.) Navidrome documents "no plans to
support browse-by-folder. Endpoints for this functionality (Ex:
`getIndexes`, `getMusicDirectory`) returns a simulated directory tree, using
the format `/Artist/Album/01 - Song.mp3`", and `getIndexes` "doesn't support
shortcuts, nor direct children"
([navidrome.org](https://www.navidrome.org/docs/developers/subsonic-api/)).
The spec commits to a folder view "derived from the paths the pipeline
chose" — a real tree. Diffing a real tree against a synthetic one produces
divergences that are noise on the endpoint where dubplate has the stronger
answer. gonic serves a genuine filesystem tree and is the second server
every client names. **If accepted**: the harness rung runs two oracle
containers, with a per-endpoint-family oracle assignment (gonic authoritative
for `getIndexes` / `getMusicDirectory` / `getMusicFolders`, Navidrome for
ID3 browsing, lists, search, annotation); the ladder's rung 1 study output
gains a column for which oracle governs each endpoint.

**3. Pin the oracle version; an upgrade is a ledger event.** (Spec L102-103,
L256-262.) Navidrome ships near-monthly and changed load-bearing behavior
inside six months: v0.61.0 (2026-03-31) overhauled artwork and moved to
server-managed transcoding, v0.62.0 (2026-06-08) added OpenSubsonic
extensions and playback reporting, v0.63.0 (2026-07-08) rebuilt search on
FTS5
([releases](https://github.com/navidrome/navidrome/releases)). The spec says
findings are "cited to a version" but the harness section does not pin one.
**If accepted**: rung 2 pins an image digest, the divergence ledger records
the oracle version per adjudication, and a version bump is a deliberate pass
that re-runs the ledger rather than a background drift.

**4. Melodee is missing from the prior art, and it is the only MIT entry in
the field.** (Spec L121-125, L307-323.) [Melodee](https://melodee.org/about/)
is .NET 10 / Blazor / Postgres, **MIT**, 1,478 commits since 2024-10, and its
architecture is dubplate's thesis: a staged inbound → staging → storage
ingestion pipeline feeding an OpenSubsonic surface, with 311 schema-validation
tests and a published
[compatibility matrix](https://melodee.org/opensubsonic-matrix/) (2.2.0,
2026-07-11). The spec's license section bars implementers from reading
Navidrome and gonic because they are GPLv3. Melodee carries no such bar. Its
bus factor is 1 and its stack disqualifies it as an oracle, but "the only
readable implementation of our own architecture" is worth a study-rung row.
**If accepted**: prior-art table gains a Melodee row (role: readable
reference implementation of the ingest-pipeline-plus-OpenSubsonic shape;
consulted by study, projection, art/search); the license paragraph notes that
MIT prior art may be read where GPL prior art may only be observed.

**5. The client roster must exclude dead clients and should anchor on
Symfonium.** (Spec L338-340, L262.) The roster is deferred to rung 2, which
is right, but the candidate pool has rotted: Sonixd archived 2024-08,
Audinaut 2022, Sublime Music 2024-12, original DSub 2022, and **Tempo is
effectively dead — v3.9.0 December 2024, the live fork is Tempus (v4.26.1,
2026-09-02)**. Ultrasonic moved to GitLab. Symfonium is the only client
publishing per-server compatibility notes and is already in the prior-art
table (L322); it should be the roster's anchor, with the caveat that it is
proprietary and paid, so the family must actually own a copy for the walk.
**If accepted**: the open item at L338 gains a shortlist and an exclusion
list; rung 2's scripted walk names Symfonium, Feishin, Supersonic,
Substreamer, and Tempus as the live candidates.

**6. Keep OpenSubsonic first; record the Jellyfin evidence so it is not
reopened.** (Spec L36-44.) Jellyfin has 56,633 stars and a real music client
in Finamp (1.0.1-beta, 2026-09-02, MPL-2.0), but its Subsonic bridges are
dead — [JellySonic](https://github.com/lyarenei/JellySonic) is archived,
self-labelled experimental, last compatible with Jellyfin 10.8;
[jellysub](https://github.com/nvllsvm/jellysub) is archived and seeking a
maintainer — and Jellyfin's own apps still lack gapless playback and
ReplayGain. Streamyfin is explicitly video-only. One strong beta client does
not outweigh ~20 maintained Subsonic clients. **If accepted**: the decision's
rejected-alternatives list gains "Jellyfin API instead of or beside
OpenSubsonic", with this evidence and its date.

**7. Pre-adjudicate the Navidrome divergences that are already known, and fix
the one where the spec contradicts the oracle.** (Spec L192-194, L226-231.)
Navidrome documents: it never marks a song played on `stream`, only on
`scrobble` with `submission=true`; `getUser` ignores the username parameter
and `getUsers` returns only the authenticated user; `search2`/`search3` accept
no Lucene queries; `getPlayQueue`'s `current` is a string, not the `int` the
official docs show; `getPlaylists` ignores `username`; IDs are always strings;
no video, ever
([navidrome.org](https://www.navidrome.org/docs/developers/subsonic-api/)).
Most of these match spec choices already. **One does not**: the spec records a
play event "when a stream passes a threshold of the track's length **or** the
client reports a scrobble" (L192-194), so every non-scrobbling client will
produce play counts in dubplate and none in Navidrome. That is a designed
divergence and the harness will report it on every run unless it is a
pre-seeded ledger entry. **If accepted**: rung 1's study output ships a
pre-seeded divergence ledger with these seven entries, and L192-194 gains one
sentence naming the play-marking difference as deliberate.

**8. Name the target extension subset.** (Spec L226-231.) The spec says
extensions are "declared through the extension-discovery endpoint" without
saying which. There are exactly ten real extensions in the spec repo today:
Transcode Offset, API Key Authentication, HTTP form POST, getPodcastEpisode,
Index-based Queue, Playback Reporting, Song Lyrics, Sonic Similarity, Top
Songs by Artist ID, Transcoding
([spec repo](https://github.com/opensubsonic/open-subsonic-api)). The spec
already commits to API-key auth (L59-66) and stubs lyrics (L226-231).
**If accepted**: the API-surface paragraph lists the launch subset explicitly
and states that podcasts, jukebox, sonic similarity, and internet radio return
the not-implemented error, which lets rung 2 measure coverage rather than
infer it.

**9. Expect search-diff noise, not signal.** (Spec L162-171.) dubplate plans
FTS5 with a trigram tokenizer over accent-folded text; Navidrome rebuilt its
own search on FTS5 in v0.61.0 and again in v0.63.0. Two FTS5 implementations
with different tokenizers and different rank tie-breaks will differ on
ordering constantly while agreeing on membership. **If accepted**: the differ
compares search results as a set plus a top-N ordering assertion, not
field-by-field on the full ordered list.

**10. Record that the field's bus factor is uniformly 1.** (Spec L19-29,
L300-306.) Navidrome (88/100 commits by one author), gonic (90/100), LMS
(100/100), Koel (75/100), Black Candy (96/100), Melodee (one copyright
holder). The four adoption criteria will score every candidate the same way
on maintenance, which means the criterion discriminates nothing here and
adoption evidence (stars, client references) is doing the real work.
**If accepted**: the prior-art scoring note says so, so the study rung does
not spend a pass proving it entry by entry.
