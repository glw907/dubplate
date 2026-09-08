# Adversarial review: the 2.0 spec's prior art, scored for adoption

Fresh-context review of
`docs/superpowers/specs/2026-09-07-streamer-2-0-design.md` revision 1,
against CLAUDE.md's "Build for 2.0" four criteria (license, code quality
against the house bar, maintenance, adoption). Everything below was
observed 2026-09-07 and is cited. Line numbers are that spec's.

## Verdict

The table's shape is right and most of its cells survive scoring, but two
of its license strings are wrong, and one of them is wrong in the
direction that matters: `go.senan.xyz/taglib` is **LGPL-2.1 in whole**,
not an "MIT wrapper" around a separately-licensed blob, so line 312 and
the open item at lines 344-346 both rest on a false premise, and the
MPL-1.1 escape the spec is counting on is not reachable through that
module at all. `mewkiz/flac` is Unlicense, not MIT. Beyond the licenses,
three entries do not earn their place: `djherbis/fscache` is a
single-maintainer library with seven importers whose only serious user
had to land the exact fix 2.0's design depends on and then pin an
untagged commit; `pressly/goose` buys a Go-function migration mechanism
this project has no use for at the price of a ten-driver dependency
graph, against a product goal of one binary a stranger can set up; and
`xrash/smetrics` duplicates `internal/engine/matcher/strdist.go` and
`translit.go`, which this repo already built and pinned. `modernc.org/sqlite`,
`golang.org/x/image`, `gen2brain/webp`, `mewkiz/flac`, wazero (as
taglib's transitive dependency only), and the three documentation
entries all clear the bar. Seven entries a Subsonic server in Go needs
are missing from the table entirely, of which the ReplayGain/R128 one is
a forward hook, not a library, and is therefore the most expensive to
discover late.

## The table, scored

All cells observed 2026-09-07. "Importers" is pkg.go.dev's known-importer
count. Code quality is judged against `~/.claude/skills/go-conventions/SKILL.md`
(stdlib idiom, no interface-first design, no gratuitous concurrency, doc
comments that state a contract); generated or transpiled code is judged
on its hand-written layer only.

| Entry | License | Code quality vs house bar | Maintenance | Adoption | Verdict |
|---|---|---|---|---|---|
| OpenSubsonic spec + extension list | Apache-2.0 ([LICENSE](https://raw.githubusercontent.com/opensubsonic/open-subsonic-api/main/LICENSE)) — spec text, not code | n/a | Live; per-extension versioning, 10 extensions, `getOpenSubsonicExtensions` discovery ([docs](https://opensubsonic.netlify.app/docs/extensions/)) | The de facto contract; every modern server | **Adopt.** Condition: if the OpenAPI schema is ever vendored to generate types, Apache-2.0's NOTICE clause attaches to a repo that declined Apache at line 124. Read it; do not vendor it. |
| Navidrome | GPL-3.0 | n/a (behavior only) | Very active | The reference server | **Adopt as oracle**, behavior only, per line 122. Its `go.mod` is public fact and citable; its source is not. |
| gonic | GPL-3.0, 2.5k stars, 55 open issues | n/a (behavior only) | Active | Smaller second reference | **Adopt as oracle**, behavior only. Note the coupling: gonic's author is also taglib's author and its sole significant maintainer, so "gonic for behavior" and "taglib as a dependency" share one bus factor. |
| `go.senan.xyz/taglib` | **LGPL-2.1, whole module.** [LICENSE](https://github.com/sentriz/go-taglib/blob/master/LICENSE) is the LGPL text; README: "This project is licensed under the GNU Lesser General Public License v2.1"; GitHub reports `spdx_id: LGPL-2.1`. No NOTICE, no per-blob license, no MPL-1.1 election anywhere in the repo | Mostly good: `sync.OnceValues` for the runtime, no speculative interfaces. Three smells: the wazero compilation cache is hardcoded to `os.TempDir()` (`taglib.go:356`); the host module answers `__cxa_allocate_exception`/`__cxa_throw` by `panic(...)` and no `recover()` exists anywhere in the 693-line file; `binaryPath` is an ldflags-injected package global that `clear()`s the embedded blob | v0.14.0 tagged 2026-07-24; 28 commits in 12 months; 3 contributors; 2 real open issues, oldest ~8 months. Depends on wazero at an **untagged pseudo-version** (`v1.11.1-0.20260428…`), not a release | 11 importers, of which three are the author's own projects. Navidrome depends on it but as of 2026-09-05 through **its own fork**, `deluan/go-taglib`, via a `replace` in navidrome's `go.mod` | **Decline as written; defer the scoped decision to rung 8.** Line 312's "MIT wrapper" is false and the whole license argument has to be re-run. Three live options for rung 8 to weigh: (a) accept LGPL-2.1 on the combined binary and document how §6 relinking is satisfied (the `-X binaryPath=` override covers the wasm half, not the Go half); (b) build TagLib's wasm under **MPL-1.1** and write an MIT wrapper in-house; (c) decline entirely — see F1. |
| `tetratelabs/wazero` | Apache-2.0 | High; the largest and best-supported pure-Go wasm runtime | v1.12.0, 2026-05-29; 82 contributors; 59 commits/12mo; 28 open issues, oldest 2022-12-31 | 714 importers | **Adopt only as taglib's transitive dependency**, never as a direct one. It is not an independent adoption decision, and it disappears if taglib does. Note the compiler engine covers linux/amd64+arm64 and macOS arm64; elsewhere it falls back to an interpreter wazero itself calls up to 10x slower. |
| `modernc.org/sqlite` | BSD-3 | Machine-transpiled C amalgamation; the house bar cannot apply and should not be claimed. The hand-written `database/sql` layer is unremarkable and fine | v1.58.0, 2026-09-01; near-weekly releases; tracks SQLite 3.53.4. Hosted on GitLab; effectively one maintainer (cznic) — thin bus factor against a long, unbroken shipping record | 3,518 importers (FerretDB among them) | **Adopt with a named condition at rung 6.** Two things must be proven by a test in the gate, not assumed: (1) FTS5's **trigram** tokenizer is present (FTS5 itself is compiled in per the build flags, but no source confirms trigram specifically), which line 164's whole search design rests on; (2) WAL under the server's real reader/writer mix, against open issue [cznic/sqlite#115](https://gitlab.com/cznic/sqlite/-/issues/115), "SQLITE_BUSY on concurrent SELECT in WAL mode", which the reporter says does not reproduce under mattn. Record `ncruces/go-sqlite3` as the named fallback (see F7). |
| `pressly/goose` | MIT | Fine | v3.28.0, 2026-09-02; very active | 1,725 importers; Navidrome uses v3.27.3 | **Decline for rung 6; hand-roll.** goose's `go.mod` direct-requires ClickHouse, mysql, pgx, mssql, moby api+client, libsql, vertica, ydb, mymysql and modernc/sqlite. Module-graph pruning keeps unused packages out of the *binary*, so this is not a size cost — it is a supply-chain and `go.sum` cost, paid forever, on a project whose stated goal is one binary a stranger can set up. What it buys over `embed.FS` + `PRAGMA user_version` (roughly 40 lines, [worked example](https://gluer.org/blog/sqlites-user_version-pragma-for-schema-versioning/)) is Go-function migrations, which hook 7 (line 296) does not ask for. `golang-migrate` is strictly worse here: heavier still (cloud SDKs and VCS clients in its direct requires), SQL-files-only, last release 2025-11-29. |
| `mewkiz/flac` | **Unlicense**, not MIT — line 316 is wrong. Public-domain-equivalent, so permissive enough, but say so accurately | Idiomatic: clean `frame`/`meta` package split, decoder + seek API | v1.0.14, 2026-08-19; 15 contributors; 9 open issues; low but live commit volume on a 2012 codebase | 96 importers | **Adopt** for rung 7's integrity check. It supplies what `internal/engine/shadow/flac.go` deliberately does not: audio-frame decode, so the STREAMINFO MD5 can actually be verified. Fix the license string. |
| Picard tag-mapping tables | CC0 1.0 ([picard-docs](https://github.com/metabrainz/picard-docs)) | n/a (documentation) | Current, versioned with Picard (v3.0 latest) | The canonical cross-container mapping | **Adopt.** Cite a pinned doc version, as the matcher pins unidecode 1.4.0. |
| `golang.org/x/image` | BSD-3 | The exemplar; Go team | v0.45.0, 2026-08-11 | 1,948 importers on `draw` | **Adopt.** One note for the art slice: Navidrome's own switch off `disintegration/imaging` chose `draw.BiLinear`, not `CatmullRom`, for thumbnail speed. Because line 158's cache makes the encode a once-per-variant cost, CatmullRom is affordable here and probably right — but make that a stated choice with a measurement, not a default. `disintegration/imaging` is out regardless: last release 2019-11-16, last commit ~2020-12-18. libvips bindings (`govips`) are ~6x faster but reintroduce cgo and a system library, which the estate's static-build posture rules out. |
| `gen2brain/webp` | MIT | Build-tag-heavy (purego / nodynamic / transpiled paths) and largely generated, but the seam is honest and cgo is never required | v0.6.4; active | 18 importers, including Navidrome's `core/artwork` | **Adopt.** `x/image/webp` is still decode-only; the encoding proposal [golang/go#45121](https://github.com/golang/go/issues/45121) has been open since 2021-03-19. Navidrome's measured result for exactly this pipeline: ~74% smaller at equal quality for ~25% slower encode, paid once because cached. |
| `djherbis/fscache` | MIT | 2016-era Go: `Cache` interface, pluggable `Handler`/`haunter`. Interface-first, which the house bar warns against, and larger than the problem | Last **tag** v0.11, 2021-04-25. Dormant 2023-11 → 2026-08, then a four-day burst. 4 open issues, opened 2016, 2016, 2018, 2022. One committer of record | **7 importers.** Navidrome pins an **untagged** commit (`v0.10.2-0.20260829235704-6d85d5878c22`, no `replace`), and that commit is *Navidrome's own maintainer's* PR #22, "Add CloseWithError so readers learn a write failed" | **Decline as a launch dependency; defer to 2.0 rung 5 behind a named bar.** The bar: write the four-case concurrency table first — two readers on one fill, writer fails mid-stream, last reader disconnects, entry evicted while read — and let it decide. The evidence says the library does not yet ship a release containing the semantics line 189-192 depends on, and that its one serious user is effectively maintaining it. See F2. |
| `xrash/smetrics` | MIT | Plain functions, no ceremony; fine | Small, essentially finished; 44 commits | 105 importers | **Decline: the repo already covers it.** `internal/engine/matcher/strdist.go` implements beets' string distance (`StringDist`, `Distance.AddString`) and `translit.go` carries the pinned unidecode 1.4.0 table with a digest test. A 2.1 rerank that used smetrics instead would fold text two ways in one system — the exact defect the folded-sort-key hook (line 286) exists to prevent. |
| Chromaprint / `fpcalc` | Chromaprint's own source is MIT, but it links FFmpeg, so the distributed library is **LGPL-2.1+**; built against FFTW3 instead it becomes **GPL** | n/a | Stable upstream | The AcoustID standard | **Adopt for 2.1 with a condition: subprocess only.** Invoked as a separate executable under `exec.CommandContext` — as ffmpeg already is at line 68 — no linking obligation attaches. Never link libchromaprint. Provisioning must pin the FFT backend so a GPL build never lands. |
| Symfonium compatibility notes | Documentation | n/a | Live forum + wiki | The best public record of client-side divergence | **Adopt** as a study input. Concretely useful already: its compatibility switches name `search3`-with-empty-query and the `download`-instead-of-`stream` fallback as the two places servers break clients ([docs](https://docs.symfonium.app/wiki/providers/subsonic-opensubsonic-media-provider-configuration/)). |

## Findings, ranked

**F1 — Line 312 and lines 344-346: the taglib license is stated wrong,
and the stated escape does not exist.** The spec calls it an "MIT
wrapper" over a dual-licensed TagLib and parks the variant question as a
non-blocking open item for rung 8 ("MPL-1.1 is the safe choice"). The
module is LGPL-2.1 in whole — the wrapper's own Go source, not just the
blob — so no MPL election is available through it, and rung 8 cannot
"confirm which variant the Go binding ships" because the binding ships
one. This is not a documentation fix. An MIT project statically
embedding an LGPL-2.1 module distributes a combined work under LGPL-2.1
§6, whose relinking requirement Go's static-only distribution model has
no settled answer for. The blob is 388,336 bytes, committed in-tree and
`go:embed`ed. Rung 8's plan must open with the license decision, and
lines 344-346 must be re-graded from "open item, blocks nothing" to a
blocking ruling. Cheapest alternative if it is declined: the repo's own
`internal/engine/shadow/flac.go` already walks FLAC metadata blocks by
type; reading block type 6 (PICTURE) for line 155's art fallback is a
few dozen lines on top of it, and `REPLAYGAIN_*` comments are already in
the map it builds. Writing tags for other containers is the real gap,
and belongs to rung 8's decision, not 2.0's.

**F2 — Lines 188-192: the transcode design's hardest requirement rests
on the table's weakest dependency.** "Its output streams into the cache
entry through `djherbis/fscache`, and the client reads the entry as it
fills, so a second listener shares the encode" is the one genuinely
subtle piece of the streaming slice, and the entry backing it has 7
importers, one maintainer, no tagged release since 2021, and a
three-year dormancy broken only when Navidrome's maintainer landed the
partial-write signal himself in August 2026. Adopting it means pinning
an untagged commit — every cost of a dependency, none of the
maintenance. The stdlib idiom for this is within reach: a `sync.Map` of
in-flight encodes, `io.MultiWriter` into a temp file plus the attached
readers, `os.Rename` on completion (the pattern `internal/atomicfile`
already models), `http.ServeContent` on the hit path. That is
comparable in size to `internal/engine/musicbrainz/limiter.go`, which
this repo wrote rather than imported for a similar reason.

**F3 — Lines 163-165: FTS5 trigram support is assumed, not evidenced.**
The search design specifies "an FTS5 index … using the trigram
tokenizer" and the table's SQLite row carries no cell proving
`modernc.org/sqlite` has it. FTS5 is compiled in; trigram specifically
is unconfirmed in any source found. Rung 6's plan must land a test that
creates a `tokenize='trigram'` table and queries it, in the gate, before
line 164 is treated as settled. The same rung owes the WAL check
against cznic/sqlite#115.

**F4 — Line 315: goose is adopted for a mechanism hook 7 does not
request.** Line 296-298 asks only that "the store's migration mechanism
is chosen with [schema evolution] in view". Embedded numbered SQL plus
`PRAGMA user_version` satisfies that in about forty lines and adds no
dependency graph. Adopting goose because Navidrome does is familiarity,
which CLAUDE.md's adoption paragraph names as the thing evidence is
supposed to replace.

**F5 — Line 320: smetrics duplicates an existing pinned surface.** See
the table row. The 2.1 rerank should be specified against
`matcher.StringDist`, and the table row should be struck rather than
deferred.

**F6 — Line 316: `mewkiz/flac` is Unlicense, not MIT.** Second license
error in one table. Both suggest the cells were filled from memory. The
study rung's deliverable should require each license cell to cite the
repository's LICENSE file by URL.

**F7 — Seven entries a Subsonic server in Go needs, missing from the
table.** Scored on the same four criteria:

| Missing entry | License | Quality | Maintenance | Adoption | Verdict |
|---|---|---|---|---|---|
| `ncruces/go-sqlite3` — the fourth SQLite option | MIT | Well-documented VFS design | Active | 96 importers; **gonic uses it** | **Record as the named fallback** if F3's trigram or WAL check fails. Real SQLite compiled to wasm rather than transpiled C, so upstream FTS5 arrives verbatim. Higher memory per connection. The other two alternatives are settled: `mattn/go-sqlite3` (MIT, 24,316 importers, Navidrome's choice) requires cgo and forfeits `CGO_ENABLED=0`; `crawshaw.io/sqlite` is dead (last push 2024-04-21) and superseded by `zombiezen.com/go/sqlite` (ISC, fastest on queries in [cvilsmeier/go-sqlite-bench](https://github.com/cvilsmeier/go-sqlite-bench), but 128 importers and last tagged 2025-05-23, and built on modernc anyway). |
| **ReplayGain / R128 gain** — not a library, a forward hook | n/a | n/a | n/a | OpenSubsonic's `Child` carries `replayGain`; Navidrome shipped R128 in v0.53.0 | **Add to the forward-hooks list (line 282-298) as an eighth hook, rung 6 schema + rung 7 admit.** Track and album gain/peak, plus Opus `R128_TRACK_GAIN`, must land on the catalog row at admit. Discovered at 2.0 instead, it means re-reading every file in the library. Highest-value omission on this page. |
| Lyrics (LRC / SYLT / USLT) | n/a | n/a | n/a | OpenSubsonic Song Lyrics extension, `structuredLyrics`; Navidrome reads sidecar `.lrc`/`.txt` | **Hand-roll; add as a 2.1 row.** No Go LRC library is worth adopting. Line 229 stubs lyrics at launch, which is fine, but the table names no source and the 2.1 list (line 328-333) omits lyrics entirely. |
| M3U playlist read/write | Candidates all MIT-ish | Tiny, IPTV-oriented | `sherif-fanous/m3u` and `pawanpaudel93/go-m3u-parser` released 2025; `grafov/m3u8` is archived (→ Eyevinn/hls-m3u8) and is HLS, the wrong format | Negligible | **Decline all; hand-roll.** `#EXTM3U` is line-oriented and about forty lines. It belongs on the table because line 242 needs it: importing Navidrome's exported playlists at cutover. |
| Subsonic XML/JSON dual encoding | n/a | n/a | n/a | Navidrome hand-rolls `server/subsonic/responses` | **Declined-and-hand-rolled, recorded explicitly.** Line 222-224 says "XML and JSON from the same typed response structs", which `encoding/xml` and `encoding/json` do not make free: attribute-heavy XML, the `subsonic-response` wrapper, singleton-vs-array in JSON, and JSONP (`f=jsonp&callback=`). No library; the study rung should enumerate the quirks so the harness tests them. |
| Failed-auth rate limiting (`golang.org/x/time/rate`) | BSD-3 | Exemplar | Go team | 14,348 importers | **Adopt, or state stdlib.** Line 204-205 requires per-account and per-address limits and the table names nothing. `internal/engine/musicbrainz/limiter.go` does **not** cover it — that is class-priority fairness over one shared wire budget, a different problem from per-key token buckets. |
| App-password encryption at rest | n/a | n/a | n/a | n/a | **Name stdlib and decline `x/crypto`.** Line 199-201 needs reversible encryption because the legacy salted-token scheme needs the plaintext; `crypto/aes` + `cipher.NewGCM` is the whole answer. Saying so in the table stops a rung reaching for a KDF library it does not need. |

**F8 — Line 148 and the art blob: `internal/atomicfile` covers part of
this and should be said to.** `atomicfile.WriteFile` is bytes-only. The
content-addressed art blob (line 154) and any cache entry want a
*streaming* atomic write — temp file, `io.Copy`, `Sync`, `Rename`. That
is a small extension to an existing package, and naming it keeps a rung
from inventing a second temp-file convention.

**F9 — Line 325: the not-adopted list is accurate but incomplete.**
Navidrome's `go.mod` also carries `pocketbase/dbx`, `lestrrat-go/jwx/v3`
and `santhosh-tekuri/jsonschema`, all in the same "architecture 2.0 chose
against" category. Minor; add them so the sentence is exhaustive.

**F10 — Line 336-337: "None blocks the study rung" no longer holds.**
F1 makes the taglib license a rung-8 blocking ruling rather than an open
item, and F3 makes the trigram check a rung-6 gate item. Both are owner-
visible only as budget, not as taste, so neither needs a sitting — but
the spec should stop calling them non-blocking.
