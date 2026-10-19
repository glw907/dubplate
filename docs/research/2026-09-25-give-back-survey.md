# Give-back survey: publishing dubplate components as Go libraries

Evidence date: 2026-09-25. Every license below was read on that date from a LICENSE file, package metadata (PyPI JSON,
installed `*.dist-info/METADATA` in `tools/venv/.venv`, CPAN metadata, the GitHub API, `pkg.go.dev`), or the
publisher's own terms page, with the URL given in place. Importer counts are `pkg.go.dev` "Imported by" figures read the
same day. Nothing here is legal advice: where a question turns on law, the section states the facts and names the
question the owner or a lawyer decides.

## Brief

Four candidates were surveyed beside the OpenSubsonic layer the owner already ruled on. The matcher is the strongest
give-back. No Go port of beets' autotag scoring exists, and the package already imports nothing from dubplate. One
blocker stands in the way. The matcher embeds a transliteration table derived from Python Unidecode 1.4.0, and that
table is distributed under GPL-2.0-or-later with no permissive grant. So the matcher cannot ship that table under MIT or
Apache-2.0 on any license its authors have granted. No permissive replacement reproduces Unidecode bit for bit, and the
matcher must match beets bit for bit. The workable path is to publish the matcher with a pluggable transliterator and no
table, and leave the table where its license allows. The same GPL table sits inside dubplate today, whose repo already
carries an MIT `LICENSE` file; that is an open question for the owner whatever happens to the matcher. The corpus can
ship once folksonomy tags and genres are stripped from 138 cases, and it carries no personal data. The FLAC tag writer
fills a real gap: no Go library does crash-safe Vorbis-comment and PICTURE writes to beets' field table. The MusicBrainz
client does not fill a gap on its own, because `go.uploadedlobster.com/musicbrainzws2` (MIT) covers the wire; its
beets-translation layer belongs with the matcher.

## Verdicts

| Candidate | Verdict | License | Reason for the license | Boundary cut (`go list -deps`) |
|---|---|---|---|---|
| Unidecode table (`unidecode-1.4.0.tbl`) | Decline to publish under a permissive license | n/a (GPL-2.0-or-later as distributed) | Its only granted license is GPLv2+ | n/a |
| Matcher (`internal/engine/matcher`) | Publish after the table blocker is fixed | MIT | Its upstream, beets, is MIT, and dubplate's own `LICENSE` is MIT | None in non-test code; remove the `go:embed` of the table; the conformance suite lives outside the package |
| MusicBrainz translation and client (`internal/engine/musicbrainz`) | Publish inside the matcher module; decline as a standalone client | MIT | Ports beets' MIT MusicBrainz plugin | Imports `internal/engine/matcher`, so it moves with it; one test reads `corpus/` |
| Corpus (`corpus/`) | Publish after stripping tags and genres | Harness code MIT; data carries its sources' terms | MusicBrainz core data is CC0; MIT and Apache-2.0 are software licenses | Separate repo or module; `internal/harness/{cassette,conform}` read it |
| FLAC tag I/O (`internal/engine/tagio`) | Publish, with boundary work | MIT | Its field table ports mediafile, which is MIT | Imports `internal/atomicfile`; tests read `corpus/analysis/file_tags.json` and `internal/harness/oracle/testdata/template.flac` |

## The Unidecode transliteration table

### How the table is derived

`tools/unidecode-table/generate.py` imports the pinned `unidecode` package from `tools/venv/.venv`. It walks every data
module the package ships and writes each codepoint's replacement string into a binary file. It also bakes in a SHA-256
digest over every codepoint from 0 to 0x10FFFF. `internal/engine/matcher/translit.go` embeds the result
(`//go:embed unidecode-1.4.0.tbl`, 252 KB). `TestUnidecodeTableFullSweep` proves the Go decoder reproduces the digest.
The table is therefore a mechanical, complete extraction of Unidecode 1.4.0's data.

### What the licenses say

- **Python Unidecode 1.4.0.** PyPI metadata gives `License: GPL` and the classifier "GNU General Public License v2 or
  later (GPLv2+)" ([PyPI JSON](https://pypi.org/pypi/Unidecode/1.4.0/json)). The installed
  `Unidecode-1.4.0.dist-info/LICENSE` is the GPL version 2 text. The project README states: "Original character
  transliteration tables: Copyright 2001, Sean M. Burke, all rights reserved." It then licenses "Python code and later
  additions" under GPL-2.0-or-later ([README.rst](https://raw.githubusercontent.com/avian2/unidecode/master/README.rst)).
  Its FAQ declines a permissive relicense and points readers who need one to `text-unidecode`.
- **Perl Text::Unidecode 1.30** (Sean M. Burke, 2016-11-26). The module's LICENSE section reads: "Copyright (c) 2001,
  2014, 2015, 2016 Sean M. Burke. Unidecode is distributed under the Perl Artistic License, namely: This library is free
  software; you can redistribute it and/or modify it under the same terms as Perl itself"
  ([source](https://fastapi.metacpan.org/v1/source/SBURKE/Text-Unidecode-1.30/lib/Text/Unidecode.pm), lines 813-825).
  CPAN metadata records the license as `perl_5`, meaning Artistic 1.0 or GPL
  ([metacpan](https://metacpan.org/pod/Text::Unidecode)).

### The plain answer

Unidecode 1.4.0 is distributed under GPL-2.0-or-later, and no permissive grant covers it. Burke's original tables carry
Perl's terms (Artistic 1.0 or GPL), which are not MIT or Apache-2.0 either. The 1.4.0 table also includes the Python
maintainers' later additions, which carry only the GPL. So an embedded table derived from 1.4.0 cannot ship under MIT or
Apache-2.0 on any license its authors have granted.

The questions a lawyer or the owner decides are these. Is a character-mapping table copyrightable expression, or
unprotected fact? Is an extracted, re-encoded table a derivative work of the GPL package? The survey takes no position on
either.

### What this means for dubplate today

- The repo carries an MIT `LICENSE` file (added in `d4e15db` and `f64f74e`, 2026-09-07). The ROADMAP `oss-pair` entry
  still treats open-sourcing as undecided.
- The GitHub repo is private (`gh repo view`: `"visibility":"PRIVATE"`). The binary runs on the owner's own box.
- The `dubplate-server` binary embeds the GPL-derived table through `package matcher`.
- The question for the owner is whether an MIT-labeled repo and binary may carry a GPL-derived data file. It matters the
  moment either is distributed to anyone else. The options are to relabel the repo, to isolate the table with its own
  license notice, or to replace it. Each one trades against beets conformance.

### Replacement options

None of these reproduces Unidecode 1.4.0 bit for bit, and the matcher's conformance to beets requires exactly that.

| Option | License (verified) | Derived from the GPL original? | Bit-exact with Unidecode 1.4.0? |
|---|---|---|---|
| `mozillazg/go-unidecode` v0.2.0 | MIT ([LICENSE](https://raw.githubusercontent.com/mozillazg/go-unidecode/master/LICENSE)) | README says "Inspired by python-unidecode". Its `table/x000.go`… files mirror Python Unidecode's `x000.py`… module layout, and its changelog reads "use empty string instead of `[?]`", which is Unidecode's historical placeholder. That evidence points to derivation from Python Unidecode data | No. The changelog records its own data edits; last push 2023-05-14 |
| `rainycape/unidecode` | Apache-2.0 ([LICENSE](https://raw.githubusercontent.com/rainycape/unidecode/master/LICENSE)) | `table.txt` carries `/* x000 */` section headers, the Unidecode module layout; the repo states no source | No proof; last push 2016-03-08 |
| `gosimple/unidecode` | Apache-2.0 (GitHub API) | README: "Fork of rainycape/unidecode" | No proof; last push 2021-11-08 |
| `fiam/gounidecode` | No license (GitHub API returns none) | Not examined | Not usable without a license |
| AnyAscii | ISC ([LICENSE](https://raw.githubusercontent.com/anyascii/anyascii/master/LICENSE)); Go port in the same repo | README: "an alternative to (and inspired by) Unidecode"; mappings "based on popular existing romanization systems" | No, by design: it ships a `unidecode/unidecode.tsv` to compare against and claims different results |
| ICU or CLDR transforms (`Any-Latin; Latin-ASCII`) | Unicode License V3 ([unicode.org](https://www.unicode.org/license.txt); CLDR's [LICENSE](https://raw.githubusercontent.com/unicode-org/cldr/main/LICENSE), `SPDX-License-Identifier: Unicode-3.0`) | No | No. The transforms are rule-based and script-aware. Go has no pure-Go ICU transliterator, so this means cgo ICU or a new rule engine |
| `text-unidecode` (Python) | Artistic License, per PyPI classifiers, which also list GPL and GPLv2+ ([PyPI JSON](https://pypi.org/pypi/text-unidecode/json)) | Yes, from Burke's Perl tables | No; it predates the Python project's later additions |

### A path that keeps conformance

Publish the matcher with a transliterator seam and no table. The function the matcher needs is `func(rune) (string,
bool)`, which `translit.go`'s `transliterateRune` already has. The published module would ship three things:

- the seam;
- the table-file decoder and format, which are dubplate's own code;
- `generate.py`, which a user runs against their own installed Unidecode to produce the table.

Conformance to beets then holds for anyone who supplies a Unidecode 1.4.0 table. A permissive default, such as AnyAscii,
could ship for callers who want no GPL data, documented as not beets-conformant. The owner decides whether dubplate itself
keeps embedding the table and under what notice.

## The matcher

### What it ports and the upstream licenses

| Upstream | What the Go code reproduces | License (verified) | Notice obligation |
|---|---|---|---|
| beets 2.14.1 (`beets/autotag`, `beetsplug/musicbrainz.py`) | Scoring, distance, recommendation, candidate admission, the MusicBrainz translation | MIT, "Copyright (c) 2010-2016 Adrian Sampson" (installed `beets-2.14.1.dist-info/licenses/LICENSE`; [repo LICENSE](https://raw.githubusercontent.com/beetbox/beets/master/LICENSE); PyPI `License: MIT`) | MIT requires the copyright and permission notice in "all copies or substantial portions of the Software" |
| `lap` 0.5.13 (`lap.lapjv`) | `solveLAP` in `assign.go`: "The three routines below are that solver's own, in its own order" | BSD-2-Clause, "Copyright (c) 2012-2025, Tomas Kazmar" (installed `lap-0.5.13.dist-info/licenses/LICENSE`; PyPI `BSD-2-Clause`; [repo](https://github.com/gatagat/lap)). Its METADATA credits a public-domain Pascal implementation by A. Volgenant | Source redistribution retains the notice; binary redistribution reproduces it in documentation |
| `jellyfish` 1.2.1 (`levenshtein_distance`) | `Levenshtein` in `translit.go` | MIT, "Copyright (c) 2015 James Turk" (installed dist-info) | The same MIT notice, if the owner judges the port a copy |
| Unidecode 1.4.0 | The embedded table | GPL-2.0-or-later | See the section above |

Whether a behavior port that follows upstream routines line by line counts as a "copy" is a legal question. Carrying
beets' and lap's notices in a `NOTICE` or `THIRD_PARTY` file costs nothing and removes the question.

### Existing Go ports

No Go port of beets' autotag scoring was found. GitHub searches for `beets language:go`, `autotag musicbrainz
language:go`, and `musicbrainz tagger language:go` returned these nearest neighbors:

- `sentriz/wrtag` (GPL-3.0, active, 437 stars). A Go MusicBrainz tagger that describes itself as "similar to … Beets …
  but aims to be simpler". It does not claim beets conformance, and its GPL-3.0 license makes it unusable in a permissive
  module.
- `llehouerou/waves` and `aunefyren/autotaggerr` (both GPL-3.0).
- `Ambrevar/demlo` (MIT, last push 2018).

None reproduces beets' distance model.

### Boundary check

`go list -deps` over `./internal/engine/matcher` lists no non-standard package but the matcher itself, and its test
dependencies add none. The cuts are:

- Replace the `go:embed` of the table with the seam described above.
- The conformance proof is `internal/harness/conform`, which imports the matcher and reads `corpus/`. A published matcher
  without that suite loses its evidence. So the suite and a stripped corpus need to publish too, or the matcher's README
  cites dubplate's suite.
- Package testdata (`lapjv-optima.json`, `featuring-golden.json`, `python-word-ranges.json`, `case-mapping-sweep.tsv`)
  is self-contained. `gen_featuring_golden.py` and a stray `__pycache__` directory sit beside it.

The gate check for the new module: `go list -deps ./...` shows no `github.com/glw907/dubplate/` path, and no `go:embed`
names a `.tbl` file.

## The MusicBrainz translation and client

### What dubplate's package does

Per `internal/engine/musicbrainz/doc.go`, the package has four parts:

- It translates a decoded WS/2 release into the matcher's `AlbumInfo`. This is a port of beets' MusicBrainz plugin,
  including pseudo-release merging and alias selection.
- It builds beets' Lucene album search query.
- It issues release lookups and release searches (`Client.Release`, `SearchReleases`, `AlbumForID`) with retry.
- It paces every request against one shared budget through a three-class priority limiter with an anti-starvation
  floor (`limiter.go`).

### Existing Go clients

| Library | License | Latest | Importers | Rate limiting | Coverage |
|---|---|---|---|---|---|
| `go.uploadedlobster.com/musicbrainzws2` (Philipp Wolfer) | MIT (`pkg.go.dev`) | v0.19.0, 2026-06-09; the repo log shows commits on 2026-09-22 | 8 | Reactive only: resty retries 429 and 5xx up to five times and honors `Retry-After`. It has no proactive one-request-per-second pacer ([client.go](https://git.sr.ht/~phw/go-musicbrainzws2/blob/main/client.go)) | Lookup, browse, and search for all entities, plus submissions and collections. JSON. Labeled beta |
| `michiwend/gomusicbrainz` | MIT (GitHub API; README badge) | Pseudo-version from 2018-10-12; last push 2023-05-15 | 29 | Sets a User-Agent. No pacing code was found in `gomusicbrainz.go` | Search and lookup, no browse (README). XML |
| `mgdigital/go-musicbrainz`, `rustedshader/musicbrainz-go` | MIT | 2025-06-02 and 2024-12-30 | 0 stars and 1 star | Not examined | Not examined |
| `sentriz/wrtag`'s client | GPL-3.0 | Active | n/a | n/a | Unusable in a permissive module |

MusicBrainz's own terms require at most one call per second per application and a meaningful User-Agent
([MusicBrainz API](https://musicbrainz.org/doc/MusicBrainz_API)).

### Is there a gap?

For the wire, no: `musicbrainzws2` is maintained, MIT, and broader than dubplate's client. Two gaps remain:

- **The beets translation layer.** It is only useful beside the matcher, so it belongs in the matcher module as a
  subpackage.
- **Proactive pacing.** `musicbrainzws2` has none. dubplate's class-priority limiter solves a problem most callers do not
  have, and a plain `golang.org/x/time/rate` limiter covers the rest. It does not justify its own module.

### Boundary check

`internal/engine/musicbrainz` imports `internal/engine/matcher`, and nothing else from dubplate. It moves with the matcher
module, so the matcher dependency stays in-module. `release_test.go:43` reads
`corpus/cases/media-baseline/responses/001.json` and needs a local fixture copy. The limiter and client could split into
their own subpackage so the translation can also run over `musicbrainzws2`'s types. That split is optional.

## The corpus

### What it holds

`corpus/` is 106 MB with 216 cases. Of these, 138 (the seven smoke cases and 131 `real-*` cases) carry verbatim
MusicBrainz WS/2 responses captured live. The rest are synthetic releases built by `harness synth`. Each case holds:

- `case.json`, with input items, the request log, and the expected beets record;
- `claims.json`;
- `responses/NNN.json`.

The `real-*` cases were selected by quota from the MusicBrainz catalog (`corpus/README.md`, "The real-catalog stratum",
and the selection dossier). They were not drawn from a member's library.

### MusicBrainz data licensing

- Core data is CC0. The core entities are "Areas, Artists, Events, Genres, Instruments, Labels, Mediums, Places,
  Recordings, Release Groups, Releases, Series, Works, Relationships & URLs, CD Stubs"
  ([MusicBrainz Database](https://musicbrainz.org/doc/MusicBrainz_Database);
  [Data License](https://musicbrainz.org/doc/About/Data_License)).
- Supplementary data is CC BY-NC-SA 3.0. It covers "user submitted annotations, tags (including genre associations) and
  ratings, derived statistics, search indexes, edit history, non-personal user data" (the same pages).
- The API page says: "Non-commercial use of this web service is free". It defers to the Data License for the data
  ([MusicBrainz API](https://musicbrainz.org/doc/MusicBrainz_API)).

The capture requests `inc=…+genres+tags`, as `case.json`'s `path_query` shows. A key census of all 359 response files
found these results:

- `tags` is non-empty in 8,601 places and `genres` in 8,291. Both are supplementary (CC BY-NC-SA 3.0), and every
  non-empty one sits in the 138 captured cases.
- `aliases` is non-empty in 13,132 places. The core list names entities rather than fields, so whether aliases count as
  core is listed under "Not verified".
- Search responses carry a relevance `score` and a `count`. Whether a search `score` is "search index" data is also
  listed under "Not verified".
- No `annotation`, `rating`, or `user-*` field appears.

### Can it be republished?

Yes, after stripping, and with one derivation cost:

- Strip every `tags` and `genres` array from the captured responses. That would make the non-commercial and share-alike
  terms moot for what remains.
- Stripping changes bytes the README calls "verbatim". If pinned beets reads genres under either profile, every
  affected case needs re-deriving with `make derive` and its record re-frozen. If it does not, a byte-identical
  re-derivation proves stripping is safe. This is a plan-time check.
- The synthetic cases and the analyses are dubplate's own work, and the owner licenses them.
- MIT and Apache-2.0 are software licenses and fit a data set poorly. The owner decides whether the dubplate-authored
  data takes CC0, matching MusicBrainz core, or rides the code license.

### Personal and site-specific content

A scan of `corpus/` for home and mount paths, the owner's handle, the estate's `musicbox` name, and email addresses
found nothing personal:

- No `path` value in any `case.json` is non-empty.
- `corpus/profiles/*.yaml` use `{{DUBPLATE_LIBRARY_DIR}}` and `{{DUBPLATE_LIBRARY_DB}}` placeholders.
- Hits on a surname match MusicBrainz artist names and composer sort names, which are public catalog data.
- The only `/media/` hits are third-party URL relations, in
  `real-reissue-songs-key-of-life-2013{,-stock}/responses/001.json` and
  `real-noise-grateful-dead-de1996/responses/001.json`.

The site-specific material is editorial:

- `corpus/profiles/estate.yaml` is named for and reduced from the estate's beets config. It holds no secrets or paths.
- `corpus/README.md` cites internal rulings, task numbers, and the owner by first name. A public copy wants a rewritten
  README.

### Boundary check

`internal/harness/cassette` and `internal/harness/conform` read the corpus. `conform` imports the matcher and the
musicbrainz package. The corpus needs its own repo or module, because every importer of a Go module downloads the
module's testdata too. The conformance harness either travels with it or with the matcher.

## FLAC tag I/O

### What `tagio` does

Per `internal/engine/tagio/doc.go` and `deviations.go`, the package reads and writes FLAC tags. It projects 69 writable
fields named after mediafile 0.17.0's fields, with Vorbis keys cited to mediafile. It parses and writes PICTURE blocks,
reserves 8 KiB of padding, and rewrites through `internal/atomicfile` so a crash never leaves a half-written file. It
has a format seam that carries one row, FLAC. `deviations.go` lists every deliberate divergence from mediafile and
mutagen behavior.

### Existing Go libraries

| Library | License | Latest | Importers | Writes Vorbis comments and PICTURE? | Field table |
|---|---|---|---|---|---|
| `mewkiz/flac` (+ `meta`) | Unlicense (GitHub API) | v1.0.14, 2026-08-19 | 96 | Decodes both. Its encoder encodes whole streams; no in-place metadata rewrite (rung 8 spec) | None |
| `go-flac/go-flac` + `flacvorbis` + `flacpicture` | Apache-2.0 (GitHub API) | v1 module paths last tagged 2023-05-22; repos pushed 2025-08-06, 2024-08-02, 2024-10-22 | 35, 27, 24 (v1 paths; the rung 8 spec counts 8, 6, 4 on the v2 paths) | Yes, marshal both. The rung 8 spec found the in-place save truncates and shifts audio with no temp file | None; raw key/value |
| `dhowden/tag` | BSD-2-Clause (GitHub API) | Pseudo-version 2024-04-17; last push 2024-07-30 | 351 | Read-only. Issue #116 (unbounded allocation in `readPictureBlock`) is open and fix PR #115 is unmerged, per the GitHub API today | Its own small interface |
| `go.senan.xyz/taglib` | LGPL-2.1 (`pkg.go.dev`) | v0.14.0, 2026-07-24 | 11 | Yes: `WriteTags`, `WriteImage`, `WriteImageOptions` (`pkg.go.dev` index); TagLib compiled to Wasm | TagLib's property map, not mediafile's |
| `gcottom/audiometa` | MIT (GitHub API) | Last push 2024-12-29 | Not read | Not examined | Not examined |

Upstream licenses behind `tagio`'s behavior:

- mediafile 0.17.0 is MIT (`License-Expression: MIT`, installed dist-info). The 69-field table follows it, so its notice
  applies on the same terms as beets'.
- mutagen 1.48.1 is GPL-2.0-or-later (`License-Expression: GPL-2.0-or-later`, installed dist-info). `tagio` cites it as
  a behavior reference (`container.go:344-346`, `deviations.go`). The repo's golden-FLAC generator under
  `internal/engine/tagio/testdata/` runs mutagen at development time. Whether mutagen-written fixture files carry
  anything of mutagen's license is a question for the owner. The files are program output, not mutagen code.
- Picard's tag-mapping tables are CC0 1.0, per the rung 8 spec's survey (not re-read here).

### Is there a real gap?

Yes. No permissively licensed Go library writes FLAC Vorbis comments and PICTURE blocks both crash-safely and to beets'
field table. `go-flac` writes but fails the crash-safety bar. `mewkiz/flac` and `dhowden/tag` do not rewrite tags. The
one library that writes safely is LGPL-2.1 and follows TagLib's field map.

### Boundary check

`go list -deps ./internal/engine/tagio` shows one dubplate import: `internal/atomicfile`. The cuts are:

- Move `atomicfile`'s `WriteStream` into the new module as an internal package, or publish it as a small package of its
  own.
- `conform_test.go` reads `../../../corpus/analysis/file_tags.json`, `../../../corpus/cases/%s/case.json`, and
  `../../harness/oracle/testdata/template.flac`. Copy the needed fixtures into the module's testdata.
- Package comments cite rungs and rulings, such as "ruling 14", "R12", and "rung 9". They need rewriting for outside
  readers. The staged process-citation gate leg will force the same rewrite anyway.

The gate check for the new module is the same as the matcher's: no `github.com/glw907/dubplate/` path in
`go list -deps ./...`.

## Not verified

- Whether a character-mapping table is copyrightable, and whether the extracted table is a derivative work. These are
  legal questions.
- Whether `mozillazg/go-unidecode`'s and `rainycape/unidecode`'s tables came from Python Unidecode or from Burke's Perl
  tables. The layout evidence is circumstantial, and neither repo names a source.
- Whether the Python Unidecode 1.4.0 PyPI web page shows the same license text. That page failed to render. The PyPI JSON
  API and the installed metadata were read instead.
- Whether MusicBrainz treats artist `aliases` and search-result `score` fields as core or supplementary. The licensing
  pages list entities and categories, not response fields.
- Whether pinned beets 2.14.1 reads `genres` or `tags` from the response under the `estate`, `stock`, or `locale`
  profile. This decides whether stripping forces re-derivation.
- `pkg.go.dev` importer counts for `go.senan.xyz/wrtag` (the scrape returned an implausible figure, so it is omitted).
- Coverage and pacing for `mgdigital/go-musicbrainz`, `rustedshader/musicbrainz-go`, and `gcottom/audiometa`.
- Whether `go.senan.xyz/taglib`'s writes are crash-safe.
- The Picard tag tables' CC0 license was taken from the rung 8 spec
  (`docs/specs/2026-09-07-rung-8-tag-io.md`), not re-read today.
- Go's module download size limits for a module carrying the 106 MB corpus.
