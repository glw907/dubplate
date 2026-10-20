# Slice 1, Task 6: the prior-art table's open cells

Date read: 2026-10-07 (UTC, about 04:50 to 05:40). Staged outside the repo; lands at A-land as
`docs/research/2026-10-06-slice-1-prior-art-cells.md`.

**Pins read** (from `2026-10-06-slice-1-evidence-pins.md`): Navidrome v0.64.2 (`1011457`), gonic v0.22.0 (`5ae99e8`),
the OpenSubsonic specification at `bed1688` (still `main` HEAD on 2026-10-07), Melodee v2.2.0 (`d91342e`). Task 0's
LICENSE table is cited where a cell names one of its subjects; each of its rows was re-fetched here, not copied. Every
other subject below is pinned at the version this record names, read from the Go module proxy or a release tag.

**Questions this record owns.**

1. For each open license cell (Navidrome, gonic, Melodee, wazero, `modernc.org/sqlite`, `ncruces/go-sqlite3`,
   `mewkiz/flac`, Picard's tag tables, `x/image` and `gen2brain/webp`, `x/time/rate`), what URL does the project's own
   source or the GitHub API report, and does it resolve.
2. For Melodee's open "Quality" cell, what grade does a source reading earn against the house bar.
3. For every row whose version or adoption figure the spec dates, what is the current figure, dated.
4. Does any evidence here change a row's verdict (flagged for Task 7, never changed here).

All four reach a named outcome. Question 4 yields five flags (see "Flags for Task 7").

## Method and evidence channels

Three channels back every claim, and each row names which: the GitHub or GitLab API and release pages (read
2026-10-07), the Go module proxy and `pkg.go.dev` (read 2026-10-07), and pinned source read from shallow clones under
`work/task-6/clones/` (a workstation-scratch location that did not land in the repo). Clones were read, never built or run. No oracle was started, so no container, port, or fixture
state applies. The box was not contacted.

Measures, defined once so the figures compare:

- **License URL**: the URL `gh api repos/<repo>/license` reports as `html_url`, fetched with `curl -L` and required to
  return 200. For GitLab and Google-hosted modules, the project's own LICENSE URL, fetched the same way. The SPDX id is
  confirmed against the first lines of the LICENSE file in the clone.
- **Importers**: the "Known importers" figure on the `pkg.go.dev` imported-by tab of the named package. Six of the
  spec's figures match this measure to the digit today (wazero 714, `ncruces/go-sqlite3` 96, `mewkiz/flac` 96,
  `x/image/draw` 1,948, taglib 11, `djherbis/stream` 7), so the spec used it.
- **Releases in 12 months**: non-prerelease GitHub releases published on or after 2025-10-07, or, for modules that
  publish tags only, tags with a proxy timestamp on or after that date.
- **Commit concentration**: the GitHub contributors API's per-login commit counts on the default branch.

## The table, cell by cell

The first block is the open cells. The "Verdict on the cell" column says whether the spec's current wording stands.

### Servers and the contract (rows with open or re-checked cells)

| Entry | License cell (URL, resolves 2026-10-07) | Quality | Maintenance | Adoption | Verdict on the cell |
|---|---|---|---|---|---|
| Navidrome | GPL-3.0, [LICENSE](https://github.com/navidrome/navidrome/blob/master/LICENSE) (200; GNU GPL "Version 3, 29 June 2007" on its first line in the clone at v0.64.2) | n/a, read for behavior only | 15 releases in 12 months; v0.64.2 on 2026-09-24; `master` 24 commits past that tag, newest 2026-10-06; 115 open issues and 170 open PRs, oldest open issue #141 (2020-04-04); 251 contributors, with `deluan` at 4,154 of 5,065 commits (82%) and dependabot at 283 | 24,007 stars; 57,505,552 Docker Hub pulls for `deluan/navidrome`; 11,359 downloads of the v0.64.2 release assets | Fill. "Very active, near-monthly" stands. |
| gonic | GPL-3.0, [LICENSE](https://github.com/sentriz/gonic/blob/master/LICENSE) (200; same GPL v3 text) | n/a, read for behavior only | 5 releases in 12 months; v0.22.0 on 2026-06-08; `master` 43 commits past that tag, newest 2026-10-06; 163 commits since 2025-10-07, busiest in May 2026 (40); 54 open issues and 17 open PRs, oldest open issue #44 (2020-03-11); 57 contributors, `sentriz` at 1,155 of 1,379 commits (84%) | 2,563 stars; 2,121,055 Docker Hub pulls for `sentriz/gonic`; 880 downloads of the v0.22.0 release assets | Fill. "Active" stands, with the note that the newest release is four months old while `master` moves weekly. The "shares taglib's bus factor" remark stands: the author holds 84% of commits. |
| Melodee | MIT, [LICENSE](https://github.com/melodee-project/melodee/blob/main/LICENSE) (200; "MIT License", copyright 2024 Steven Hildreth) | Graded below | See below: bursty, no commit since 2026-07-12 | 66 stars; 1 open issue, 0 open PRs; GHCR pull counts need a package-read token and release assets are empty, so no download figure exists; publishes [a compatibility matrix page](https://melodee.org/opensubsonic-matrix/) (200) | Fill license. Quality graded below. Maintenance wording "Active" needs the qualifier below. |
| OpenSubsonic specification and extension list | Apache-2.0, [LICENSE](https://raw.githubusercontent.com/opensubsonic/open-subsonic-api/main/LICENSE) (200). Re-checked, filled by the reviews. | n/a | 34 commits in 12 months, uneven (9 in March 2026, 1 in August); newest 2026-08-07; 10 extensions at `bed1688` | 251 stars; 21 open PRs (the survey counted 19) | Stands. One correction to the Verdict wording: see flag 1. |
| OpenAPI schema and Subsonic 1.16.1 XSD | OpenAPI: Apache-2.0, [LICENSE](https://github.com/opensubsonic/open-subsonic-api/blob/main/LICENSE) (200). XSD: the file at `subsonic.org/pages/inc/api/schema/subsonic-rest-api-1.16.1.xsd` (200) carries no license or copyright text; the only hits for either word are the schema's own `License` element. | n/a | Unchanged | Unchanged | Stands. The "no license text, open" cell is re-confirmed. |

### Go libraries

| Entry | License cell (URL, resolves 2026-10-07) | Quality | Maintenance | Adoption | Verdict on the cell |
|---|---|---|---|---|---|
| `tetratelabs/wazero` | Apache-2.0, [LICENSE](https://github.com/wazero/wazero/blob/main/LICENSE) (200). The repository moved to the `wazero` organization; the old `tetratelabs` URL now redirects to this one, and the module path in `go.mod` is still `github.com/tetratelabs/wazero`. A one-line `NOTICE` file exists ("Copyright 2020-2023 wazero authors"). | High, confirmed at v1.12.0: the README claims Wasm Core 1.0 and 2.0 compliance, a compiler and an interpreter runtime, and "zero dependencies", while `go.mod` requires `golang.org/x/sys` alone; Go floor 1.25.0 | v1.12.0 (tag 2026-05-28, release 2026-05-29) is still the newest; 3 releases in 12 months; 35 commits in the last six months; 32 open issues and 12 open PRs, oldest issue #985 (2022-12-31); 84 contributors, `mathetake` at 966 of 1,982 (49%) and `codefromthecrypt` at 518 (26%) | 714 importers (matches the spec); 6,410 stars; Navidrome's `go.mod` requires v1.12.0 | Fill. Cite the `wazero/wazero` URL and record the move. |
| `modernc.org/sqlite` | BSD-3-Clause, [LICENSE](https://gitlab.com/cznic/sqlite/-/raw/master/LICENSE) (200; three-clause text with "Neither the name", 26 lines). Its generated `LICENSE-3RD-PARTY.md` lists what a linked binary carries: MIT (humanize, musl, sqlite-vec when imported), BSD-3-Clause (uuid, bigfft, x/sys, the modernc modules), and SQLite itself as public domain (3.53.4). It says MPL-2.0 appears only among module-graph entries compiled into nothing. | Transpiled C, unchanged; the house bar applies to the `database/sql` layer only. Not re-graded here. | v1.60.1 on 2026-09-29 (the spec says v1.59.0, 2026-09-15); 35 tags in 12 months of 171 total; seven `retract` entries in `go.mod` (the newest, v1.42.0, "Accidentaly broken"); the changelog shows a same-week fix for an upstream SQLite 3.53.3 journal-rollback corruption bug, shipped as 3.53.4. `go 1.26.0` floor. Downstream `go.mod` files must pin the same `modernc.org/libc` the module pins (v1.77.1); the repo's `go.mod` does. | 6,070 importers on `modernc.org/sqlite` (the spec says 3,518; see the figures table); neither Navidrome nor gonic requires it (Navidrome requires `mattn/go-sqlite3`; gonic requires `ncruces` by default and `mattn` under a build tag) | Fill. Verdict stands. The rung 6 trigram and WAL conditions are untouched by this evidence. |
| `ncruces/go-sqlite3` | MIT, [LICENSE](https://github.com/ncruces/go-sqlite3/blob/main/LICENSE) (200; "MIT License", copyright 2023 Nuno Cruces). The embedded SQLite build lives in a separate module, `ncruces/go-sqlite3-wasm`, which `go.mod` requires; its license is [MIT-0](https://github.com/ncruces/go-sqlite3-wasm/blob/main/LICENSE) (200). | Confirmed: the README states it wraps a Wasm build of SQLite translated to Go with `wasm2go`, and that Go and `x/sys` are the only required dependencies; the `go.mod` shows a few small `ncruces` helper modules besides. Go floor 1.26.0 | v0.35.6 on 2026-09-23 is the newest; 23 releases in 12 months (25 tags); `master` commit 2026-09-29; 5 open issues, 0 open PRs, oldest issue #53 (2024-01-19); `ncruces` at 1,081 of 1,170 commits (92%) | 96 importers (matches the spec); 1,115 stars; gonic's `deps/deps.go` imports its driver in the default build (`//go:build !nowasm`), and `deps_nowasm.go` swaps in `mattn` | Fill. "gonic uses it" is confirmed and sharpened: it is gonic's default driver. Verdict stands. |
| `mewkiz/flac` | Unlicense, [LICENSE](https://github.com/mewkiz/flac/blob/master/LICENSE) (200; "free and unencumbered software released into the public domain"). Its two direct dependencies: `icza/bitio` (Apache-2.0) and `mewkiz/pkg` (Unlicense), per the GitHub API. | Confirmed at v1.0.14: top-level package plus `frame` and `meta`; the `frame` package checks each frame's CRC-16 and reports a mismatch as an error, and exposes a `Hash` method documented for use "in conjunction with" the STREAMINFO MD5 to verify integrity; the changelog for v1.0.14 says parsing was hardened against corrupt or malicious files | v1.0.14 on 2026-08-19 is the newest; 1 tag in 12 months (15 total; v1.0.13 was 2025-07-11); 9 open issues, oldest #19 (2018-04-16); `mewmew` at 279 of 330 commits (85%) | 96 importers (matches the spec); 363 stars | Fill. Verdict stands. The CRC and MD5 evidence supports the "intake integrity check" role. |
| Picard tag-mapping tables | CC0-1.0, [LICENSE](https://github.com/metabrainz/picard-docs/blob/main/LICENSE) in `metabrainz/picard-docs` (200). The README and `conf.py` also state CC0 1.0. The table is `appendices/tag_mapping.rst`, rendered at [Appendix A, v3.0](https://picard-docs.musicbrainz.org/en/v3.0/appendices/tag_mapping.html) (200). The Picard code repository (`metabrainz/picard`) is GPL-2.0 and is not the source. | n/a. The rst is generated and carries a "Vorbis" column, which covers FLAC. | Pin: `picard-docs@35f376e` (2026-10-05); the `v3.0` branch and `main` point at that same commit; Picard 3.0 shipped 2026-10-04 after three release candidates | The canonical mapping; 16 stars on the docs repository; Picard itself 5,271 stars | Fill, with a citation rule: see flag 4. |
| `golang.org/x/image`, `gen2brain/webp` | x/image: BSD-3-Clause, [LICENSE](https://github.com/golang/image/blob/master/LICENSE) (200; a `PATENTS` file sits beside it). webp: MIT, [LICENSE](https://github.com/gen2brain/webp/blob/main/LICENSE) (200), plus [`lib/LICENSE.libwebp`](https://github.com/gen2brain/webp/blob/main/lib/LICENSE.libwebp) (200; Google's BSD-style libwebp text). webp's one direct dependency, `ebitengine/purego`, is Apache-2.0. | x/image: `draw` offers `CatmullRom` ("slow but usually looks best") and `ApproxBiLinear` ("reasonable speed and quality"); `go 1.26.0` floor. webp: confirmed, `README.md` says libwebp compiled to Wasm and transpiled with `wasm2go`, tries a dynamic library through `purego` first, and offers a `nodynamic` tag to skip it; `go 1.23` | x/image v0.46.0 (2026-09-08) is the newest, 15 tags in 12 months. webp v0.6.4 (2026-07-01) is the newest tag; `main` is one README commit past it (2026-08-15); 5 tags in 12 months; 1 open issue; `gen2brain` at 66 of 67 commits | x/image/draw 1,948 importers (matches); webp 32 importers (the spec says 18); Navidrome requires webp v0.6.4 and x/image v0.46.0, and its `core/artwork/resize.go` calls `webp.Encode` | Fill. Verdict stands. |
| `golang.org/x/time/rate` | BSD-3-Clause, [LICENSE](https://github.com/golang/time/blob/master/LICENSE) (200; `PATENTS` beside it) | Exemplar stands: stdlib-style package, Go Authors header, 32 test functions in `rate/rate_test.go` and `rate/sometimes_test.go`; `go 1.26.0` floor, no requirements | v0.16.0 on 2026-08-19 is the newest; 2 tags in 12 months; the repo's `go.mod` already pins it (line 11) | 16,473 importers (the spec says 14,348); Navidrome requires v0.16.0 and uses it in four packages | Fill. Verdict stands. |

### Rows whose license cell was filled by the reviews, re-checked for dated figures only

| Entry | License URL re-check | Figures |
|---|---|---|
| `go.senan.xyz/taglib` | [LICENSE](https://github.com/sentriz/go-taglib/blob/master/LICENSE) 200, LGPL-2.1 per the API | v0.14.0 (2026-07-24) is still the newest; 23 tags in total, 5 in 12 months; 11 importers, unchanged. Its `go.mod` still requires wazero at the untagged `v1.11.1-0.20260428013916-2bbd517b7633`. Navidrome replaces the module with its own fork (`go.mod:6`); gonic requires v0.12.0. |
| `djherbis/stream` | [LICENSE](https://github.com/djherbis/stream/blob/master/LICENSE) 200, MIT | v1.5.1 (2026-08-29) is still the newest; 2 tags in 12 months; 7 importers, unchanged. |

## Melodee's quality cell (source reading at v2.2.0, `d91342e`)

**Grade: broad and well tested in volume, below the house bar in structure, assertion discipline, and auth.** The
reading below is of Melodee's whole-tree quality. Task 5's record owns the intake pipeline in detail; nothing here
restates it.

Strengths, each from source:

- **Size and test volume.** 1,417 C# files, about 271,000 lines, of which 376 test files hold about 103,000 lines (38%).
  `tests/Melodee.Tests.OpenSubsonic` is a per-endpoint-area project of 3,353 lines with a hand-built schema validator
  (`SubsonicSchemaValidator.cs`) and a 45-test `ComprehensiveSchemaValidationTests.cs`.
- **Guard clauses.** 275 `Guard.Against` calls under `src/`; `ValidateTokenAsync` guards its three inputs on entry
  (`UserAuthenticationService.cs:183-185`).
- **Token auth shape.** The MD5-of-password-plus-salt check is computed from a secret the server can recover: an
  `OpenSubsonicSecretProtected` field is unprotected, or the older reversible-encrypted password is decrypted, then
  compared case-insensitively (`UserAuthenticationService.cs:208-230`). The same constraint binds dubplate's
  app-password storage, so this is a confirming pattern, stated here in dubplate's words.
- **Range serving.** `MediaRetrievalController.cs` answers 206 with a bounded stream and 416 for an unsatisfiable range
  (lines 136-152, 259-263, 281-305), so a reader can compare against the spec's Range rules.
- **AI-agent guides.** `AGENTS.md` points agents at per-area instruction files
  under `.github/instructions/`. The history shows 94 commits by a Copilot bot account (6% of the total).

Defects, each from source:

- **Authentication can be skipped by request shape.** `ControllerBase.cs:118-123` sets `requiresAuth = false` when the
  request is local, when the path contains `/images/` (`HttpContextRequestExtensions.cs:8-11`, a substring test on the
  path), or when `Request.Scheme://Request.Host` equals the configured base URL. `Request.Host` is built from the
  client's `Host` header. When the flag is false, `AuthenticateSubsonicApiAsync` returns success and loads the user named
  by the request's own `u` parameter with no password or token check (`OpenSubsonicApiService.cs:1704-1720`).
  **Unconfirmed by observation.** Melodee needs Postgres and a .NET build, and a clone is untrusted data, so nothing was
  run. What would confirm it: a request to a running Melodee carrying a `Host` header equal to its configured base URL,
  `u=<admin>`, and no credential. If the reading holds, the lesson for dubplate is that a trust decision never keys on a
  client-supplied header, and the auth carve-outs the spec names (the unauthenticated extensions call) need an
  enumerated allow-list, never a path pattern.
- **Vacuous assertions in the compatibility tests.** The streaming tests accept any of several statuses: the plain
  stream test expects "OK, NotFound, or BadRequest", the range test expects "PartialContent, OK, NotFound, or
  BadRequest", and the invalid-id test accepts OK as well (`StreamingEndpointTests.cs:13-48`). The comments say the test
  database holds no audio. 38 `BeOneOf` assertions appear in the project, and 10 `Skip =` markers. A test that passes
  when the server returns an error proves nothing; this is the repo's own "vacuous assertions" rule met in the wild, and
  it is the case for the harness's fixture library holding real audio.
- **A 3,933-line service file.** `OpenSubsonicApiService.cs` holds the protocol's whole behavior in one class; the
  controllers beside it total 2,507 lines.
- **Duplicated request parsing and stray code.** `ControllerBase.cs:87-110` repeats ten parameter reads for form and for
  query requests; line 40 holds a lone `;`; line 171 writes "User authenticated" to the trace after the auth decision,
  whether or not the user is authenticated.
- **Provenance is unaudited.** The two Navidrome mentions in `src/` are comments saying a `jwt` parameter and the JWT
  flow follow "Navidrome semantics" and were "seen in the Navidrome source code" (`OpenSubsonicApiService.cs:1749`,
  `ApiRequest.cs:21`). Melodee is MIT, but whether any code was derived from Navidrome's GPL source was not audited, and
  the spec's "its code may be reused" depends on it (flag 2).

Maintenance, from the commit history (GitHub API, 2025-10-07 onward, by month): October 4, November 2, December 251,
January 408, February 13, March 3, April 5, May 50, June 20, July 9. The last commit is 2026-07-12, 87 days before this
read. Releases: v1.7.6 and v1.8.0 in January 2026; v2.0.0, v2.0.1, v2.1.0, v2.1.4, v2.2.0 from 2026-05-01 to 2026-07-11.
Concentration: `sphildreth` 1,383 of 1,478 commits (94%), three contributors in all. The spec's "1,478 commits since
2024-10" is exact today, so nothing has landed since the spec was written. The database is PostgreSQL 17 per `compose.yml`
and the README, as the spec says; the DecentDB the docs mention serves the local MusicBrainz store (test names
`DecentDBMusicBrainzRepositoryTests`), so the "Postgres rules it out as an oracle" cell stands.

## Dated figures to refresh at Task 9

Each row: the spec's figure, the current figure, the source and date. "Same" means unchanged. The spec's figures date from the 2026-09-25
survey.

| Row | Spec figure | Current (2026-10-07) | Source |
|---|---|---|---|
| Navidrome stars | 23,396 | 24,007 | GitHub API |
| gonic stars | 2,538 | 2,563 | GitHub API |
| Melodee stars | 65 | 66 | GitHub API |
| Melodee commits | 1,478 since 2024-10 | 1,478, none since 2026-07-12 | contributors and commits API |
| OpenSubsonic extensions | 10 | 10 at `bed1688` | Task 0's record |
| taglib | v0.14.0, 2026-07-24; 11 importers | Same | module proxy, `pkg.go.dev` |
| wazero | v1.12.0, 2026-05-29; 714 importers | Same (tag 2026-05-28 15:27 UTC, release 2026-05-29) | proxy, `pkg.go.dev`, releases API |
| `modernc.org/sqlite` | v1.59.0, 2026-09-15; 3,518 importers | v1.60.1, 2026-09-29; 6,070 importers | proxy, `pkg.go.dev` |
| `ncruces/go-sqlite3` | 96 importers; (survey) v0.35.6 | Same | proxy, `pkg.go.dev` |
| `mewkiz/flac` | v1.0.14, 2026-08-19; 96 importers | Same | proxy, `pkg.go.dev` |
| `x/image` | v0.46.0, 2026-09-08; 1,948 importers | Same | proxy, `pkg.go.dev` (`x/image/draw`) |
| `gen2brain/webp` | v0.6.4; 18 importers | v0.6.4 (2026-07-01); 32 importers | proxy, `pkg.go.dev` |
| `x/time/rate` | 14,348 importers | 16,473 importers; v0.16.0 (2026-08-19) | proxy, `pkg.go.dev` |
| `djherbis/stream` | v1.5.1, 2026-08-29 | Same; 7 importers | proxy, `pkg.go.dev` |
| Navidrome and gonic maintenance wording | "near-monthly"; "Active" | 15 and 5 releases in 12 months | releases API |

The `modernc.org/sqlite` importer jump (3,518 to 6,070 in twelve days) is larger than the other movements. Its cause is
unknown: the measure matches the spec's for six other rows, so a change of definition is unlikely, and a stale survey
read is the likelier reading. Task 9 should write the current figure and its date and not carry a trend claim.

## Flags for Task 7

1. **OpenSubsonic row, Verdict wording.** The spec says "never vendor it, because the NOTICE clause would attach." The
   repository holds no `NOTICE` file (a recursive tree search of `main` for `notice`, `copyright`, and `authors` found
   nothing, and `LICENSE` carries the unfilled appendix boilerplate with no copyright line). Apache-2.0 section 4(d)
   attaches only to a work that includes a NOTICE file, so no NOTICE text attaches today. Sections 4(a) to 4(c) still
   require the license text and modification notices on redistribution. This bears on ruling "the protocol module's
   license" (the open fact the plan names: "what NOTICE text, if any, the OpenSubsonic repository requires": none
   today). Carried as a fact for Task 7, not decided here.
2. **Melodee row, Verdict wording.** "Adopt as readable and reusable prior art" stands for ideas. The reuse half needs
   a qualifier: the auth shortcuts and the vacuous assertions argue against lifting code from the protocol and test
   layers, and the provenance of Melodee's code against GPL servers is unaudited. Recommendation to weigh: keep Melodee
   as read-for-ideas prior art, and require a per-file provenance read before any reuse.
3. **Melodee row, Maintenance wording.** "Active" becomes "bursty: 659 commits in December and January, 21 from
   February to April, 79 from May to July, none since 2026-07-12."
4. **Picard row, citation rule.** The docs repository is CC0, but its generator `tag_mapping.py` carries a GPL-2.0-or-later
   header, and the rst is generated from it. Cite the rst and the rendered v3.0 page, never the Python file, and pin
   `picard-docs@35f376e`. The spec's "cited to a pinned doc version" is satisfied by that pin.
5. **modernc row, a standing condition.** Downstream modules must pin `modernc.org/libc` to the version `modernc.org/sqlite`
   pins (the maintainers' own rule, issue #177 on their tracker). The repo's `go.mod` matches today (v1.77.1). A bump
   that moves one without the other is the failure; rung 6's plan should say so.

No verdict changes. Flags 1 to 3 change wording, flag 4 adds a rule, and flag 5 adds a condition.

## Re-checked URLs

Every URL in the table above returned 200 on 2026-10-07 with `curl -sL`. The survey's two already-verified URLs
(Navidrome and gonic, survey lines 135-136) and the spec's existing URLs (OpenSubsonic LICENSE in both raw and blob
forms, taglib's LICENSE) were fetched fresh, and the first lines of each LICENSE file in the clones matched the SPDX
id the API reports. The one URL that changed is wazero's: the spec's source family named `tetratelabs/wazero`, and
that path now redirects to `wazero/wazero`.

## Prior art consulted

Every table row listed in Task 6's outcome: Navidrome, gonic, Melodee, wazero, `modernc.org/sqlite`,
`ncruces/go-sqlite3`, `mewkiz/flac`, Picard's tables, `x/image`, `gen2brain/webp`, `x/time/rate`, plus the dated
rows (taglib, `djherbis/stream`, the OpenSubsonic rows). Nothing is adopted or declined anew; each verdict stands, with
the wording flags above. Clean-room rule 1 holds: Navidrome and gonic appear only through LICENSE files, `go.mod`
requirements, release and API data, and two file-level facts (`deps/deps.go` and `deps_nowasm.go` selecting a driver by
build tag, `core/artwork/resize.go` calling a webp function); no copyleft code or code-shaped text is reproduced. The
Melodee reading quotes nothing beyond a few words of its test names and comments, and Melodee is MIT.

## Routed

- **Task 7:** the five flags above.
- **Task 9:** the figures table, and the license URLs for every open cell.
- **Slice 2 (harness):** Melodee's vacuous streaming assertions are the counter-example for the harness's "no
  silent-success fakes" rule and argue for audio in the fixture library; Task 0's fixture already holds it.
- **Task 1 (contract):** Melodee accepts a `jwt` request parameter that its source attributes to Navidrome (comment at
  `ApiRequest.cs:21`); whether Navidrome documents or sends it is Task 1's and Task 4's to confirm against observation.
- **Slice 3 and slice 6 (auth):** the Melodee `Host`-header finding is the case against any carve-out keyed on a client
  header or a path substring.
- **Task 5:** Melodee's intake pipeline stays Task 5's; this record graded the tree as a whole and cites no intake file.

## Not done

- The Melodee auth-bypass reading is not observed (untrusted clone, Postgres and .NET build, outside this task's
  authority to run). It stands as a source reading with the confirming request named.
- Melodee's container pull count: GHCR needs a package-read token this lane does not hold.
- No oracle was started, so no container needed removal.
