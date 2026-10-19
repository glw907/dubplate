# Conformance corpus for the beets matcher harvest

Research report, 2026-09-01 (Sonnet research agent, web + beets repo
survey), commissioned during the dubplate-organizer brainstorm. The
recommendation section is folded into the initiative spec
(2026-09-01-dubplate-organizer-design.md); this file keeps the
findings and sources.

## Recommendation

Build a frozen JSON fixture corpus of ~150-300 cases, each a
self-contained {matcher input, frozen MusicBrainz response(s),
expected beets verdict} triple, in three strata:

1. Structural/synthetic (~40%, ~60-120 cases): generated from the
   shapes beets' own unit tests encode (test/autotag/test_match.py,
   test_distance.py, test_hooks.py): track-order scrambling,
   missing/extra tracks, multi-disc medium splits, various-artists,
   data-source-priority penalties, string-distance edges (non-Latin
   scripts, "The" articles, feat./remix suffixes). Port the input
   shapes; derive expected verdicts by running real beets against
   them, never by copying.
2. Real catalog (~50%, ~80-150 cases): actual releases pulled once
   from the MusicBrainz WS (or selected via the JSON data dumps /
   canonical dataset), covering reissues, remasters,
   bootleg-vs-official duplicates, partial rips, and mistagged inputs
   (swapped artist/title, wrong year, transliterated names). Freeze
   each release's WS response as a fixture at capture time.
3. Adversarial/should-reject (~10%, ~15-30 cases): near-duplicate
   releases (same album, different pressing/label/catalog number),
   decoy candidates, garbage tags; proves the "ask" and "reject"
   boundaries, not just confident-match.

Freezing: a small capture tool hits the real WS once per case
(1 req/sec anonymous limit; proper User-Agent required) and writes raw
JSON to testdata fixtures. Both beets-under-test (subprocess, pinned
2.13.1, MB calls redirected to a local stub) and the Go matcher run
against the same frozen responses; no live network in CI, ever. beets
has no HTTP-cassette infrastructure to borrow, so the project owns its
own minimal cassette format (go-vcr style). Freshness: re-capture a
rotating subset quarterly and on beets version bumps, so behavioral
drift is caught without letting MusicBrainz data drift cause false
mismatches per run.

## How beets tests its own autotagger

Matcher tests live under test/autotag/ in beetbox/beets:
test_match.py (track/album assignment), test_distance.py (the
Distance/penalty math), test_hooks.py (TrackInfo/AlbumInfo/AlbumMatch
behavior). Inputs are Item/TrackInfo/AlbumInfo objects constructed
inline in Python, not recorded MB responses. The MusicBrainz plugin is
tested via polyfactory-style factory objects
(test/plugins/factories/musicbrainz.py), not HTTP fixtures; no
VCR/cassette or responses-mocked payloads exist in the repo. beets'
fixtures are therefore not reusable as data files from Go; the
portable asset is the scenario shapes.

## Public corpora (no audio required)

- MusicBrainz Canonical Data (CC0, ~23M rows): realistic
  artist/title strings and reissue clustering; not raw enough alone.
- MusicBrainz JSON/Postgres data dumps (CC0 core, twice weekly,
  5.4M+ releases, 38.7M+ recordings): the real source for the
  real-catalog stratum offline; WS response shapes still need
  reconstruction for inc= combinations.
- MSD-to-MusicBrainz mapping (~250K MSD IDs to 370K MBIDs, 2016):
  bulk sampling only, not curated for hard cases.
- ListenBrainz canonical/MBID mapping: explicitly unstable over time;
  pin a dated dump if used.
- AcoustID public dump (~10M fingerprint-to-MBID links): only
  relevant if AcoustID-assisted matching is added.
- No public Picard test corpus or MusicBrainz test-database subset
  designed for tagger conformance was found.

## Recording and replaying MB responses

MusicBrainz WS: 1 req/sec anonymous (higher for identified clients),
User-Agent required, core data CC0. Go has mature VCR-style tooling
(dnaeon/go-vcr, seborama/govcr).

## Sources

beets test/autotag: https://github.com/beetbox/beets/tree/master/test/autotag
(test_match.py, test_distance.py, test/plugins/test_musicbrainz.py);
MusicBrainz Canonical Data: https://musicbrainz.org/doc/Canonical_MusicBrainz_data
(announcement: https://blog.metabrainz.org/2023/06/12/new-dataset-musicbrainz-canonical-metadata/);
JSON dumps: https://musicbrainz.org/doc/Development/JSON_Data_Dumps;
DB download: https://musicbrainz.org/doc/MusicBrainz_Database/Download;
data license: https://musicbrainz.org/doc/About/Data_License;
rate limiting: https://musicbrainz.org/doc/MusicBrainz_API/Rate_Limiting;
MSD mapping: https://github.com/keunwoochoi/MSD-to-MB-mapping;
ListenBrainz mapping: https://listenbrainz.readthedocs.io/en/latest/developers/mapping.html;
AcoustID: https://acoustid.org/webservice;
go-vcr: https://github.com/dnaeon/go-vcr
