# Rung 12 slice 2b plan review: mechanics and feasibility

Target: `docs/superpowers/plans/2026-10-07-rung-12-slice-2b-oracle-harness.md` at `0ae2b54` (revision 1), worktree
`.claude/worktrees/lane-3-slice-2`. Lens: does every mechanism behave as the plan states. Each finding was checked by
a probe in a scratch directory outside the repo, by `go doc` or the Go source, by a tool's own output, or by a
fetched upstream document. Nothing in the repo was changed except this file. No container was started.

**Counts:** 0 blockers, 6 majors, 12 minors. No owner forks.

The plan holds up well overall. The task order builds, and every pin and count the quotes table states reproduces.
The light-lane claim holds with headroom. The majors are mechanisms the plan describes wrongly or leaves open, and
each one would cost a fix round or quietly drop evidence.

## Verified as stated (no finding)

- **The XSD pin.** HTTP 200, 32,775 bytes, sha256 `9baee72f...0387`. The element types by position match `:205`.
  `album` is `AlbumWithSongsID3` at `xsd:19`, `Child` at `:311` and `:387`, and `AlbumID3` at `:319` and `:393`.
  `entry` is `NowPlayingEntry` at `:284` and `Child` at `:350` and `:367`. The XSD uses only `complexType`,
  `sequence`, one `choice`, `complexContent`/`extension`, `attribute`, and `simpleType`, with no `ref=`, no
  `xs:any`, and no groups, so a hand parse is feasible. `Genre` and `Lyrics` are `mixed="true"`.
- **The OpenAPI bundle** (`cb54c03c...`) has 195 schemas and 87 paths, all under `/rest/`. `Child` has 54 properties,
  `AlbumID3` 29, `ArtistID3` 10, and `Playlist` 13. The three `...WithSongs`/`WithAlbums` types and `NowPlayingEntry`
  are `allOf`. Every endpoint Task 5 names is in the path set.
- **xmllint on `PATH` inside the gate.** `systemd-run --user --scope -p MemoryMax=3G bash -c 'command -v xmllint'`
  (the same wrapper `cairn-run-gate` uses) resolves `/home/linuxbrew/.linuxbrew/bin/xmllint`, version `21504`. The
  scope inherits the caller's environment.
- **xmllint's exit codes are documented**, which resolves "Not verified at planning" item 1 (`:810-813`). libxml2
  `v2.15.4` `doc/xmllint.xml`, DIAGNOSTICS: 0 "No error", 1 "Unclassified", 2 "Error in DTD", 3 "Validation error",
  4 "Document isn't well-formed or couldn't be read", 5 "Error in schema compilation", 6 "Error writing output", and
  so on. The runner's rule (`:445`) holds.
- **The error-line behavior.** Every schema error is reported at its element's start-tag line. That covers attribute
  errors, required-but-missing attributes, and `Missing child element(s)`, all checked on a multi-line `bookmarks`
  probe. Line numbers past 65,535 are reported exactly (`big.xml:70004`). An empty or malformed file exits 4 with a
  `parser error` line, and a missing file exits 4 with `Can't open`.
- **The light lane's 3G cap** (`:816-817`). Measured under the light lane's own scope (`MemoryHigh=2G MemoryMax=3G
  MemorySwapMax=1G`) with `memory.peak`: `go test -count=1 ./...`, the conform suite included, peaked at 605 MB in
  175 s with zero `high`/`max`/`oom` events. A cold `golangci-lint run ./...` (empty lint cache) peaked at 1.30 GB in
  26 s. `make web`'s gates launch no browser (no `playwright`/`chromium` under `web/scripts/gates/`). R26 holds:
  `go list -deps -test ./internal/harness/conform` names no package under `internal/harness/subsonic` or
  `cmd/harness`, so Go's test cache also keeps the 170 s conform suite cached across this lane's tasks.
- **Navidrome's admin sees a new library.** The same multi-library page at `b0fe572`, `:30`, says "Admin users
  automatically have access to all libraries". That settles half of "Not verified" item 4 (`:818-819`). The gonic
  README quotes at `v0.22.0` (`:104`, `:122`) are exact.
- **Task order.** No task consumes what a later task makes. The dependency list at `:114-116` is right. Minor M15
  notes that Task 6 also touches the seed golden.

## Majors

### M1 (major): `Cmd.Output` truncates xmllint's stderr, so a body with many violations silently loses them

**Location:** quotes table `:213`, Task 3 runner `:443-445`, states `:457` ("Many errors: all of them").

The plan quotes `Cmd.Output` populating `ExitError.Stderr` as the mechanism. `go doc os/exec.ExitError` says: "If the
error output is long, Stderr may contain only a prefix and suffix of the output, with the middle replaced with text
about the number of omitted bytes ... Users with other needs should redirect Cmd.Stderr as needed."
`exec.go:1037` sets `prefixSuffixSaver{N: 32 << 10}`. Probe: a `randomSongs` body with 1,000 misspelled attributes
gives `exit=3 stderrBytes=65566 validityLines=434 omitted=true`. That is 566 violations dropped, and no error
says so. A whole-library `search3` or `getAlbumList2 size=500` body crosses this line easily, since gonic adds
`title` and `album` on every ID3 album.

**Fold:** in Task 3, set `cmd.Stderr` to a `bytes.Buffer` and parse that. Never read `ExitError.Stderr`. Correct
the `:213` quote. Add a proof with more than 64 KiB of errors (500 songs) that asserts every one is reported, and a
mutation that reads `ExitError.Stderr` (the proof fails).

### M2 (major): libxml2 auto-upgrades nightly, and the exact pin then turns every gate red

**Location:** R18 `:268-270`, stop rule `:161-162`, xmllint adoption `:240-244`, close step 6 `:796-797`.

`uupd.timer` runs daily at about 04:10 with its Brew module enabled. `/etc/uupd/config.json` disables only
distrobox, and the journal shows `"msg":"Brew module" ... "Enabled":true`. Cellar directories were rewritten on
10-01, 10-02, 10-04, and 10-06. `brew info libxml2` reports "Installed (as dependency)" (of ffmpeg, imagemagick, and
others), and `~/.dotfiles/bluefin/Brewfile` has no `libxml2` line.

So any upstream libxml2 release lands overnight. The R18 refusal then fails every gate that grades XML. That
includes this lane's gates, stopping an unattended run that spans 04:10 (`:161`). After the merge it also includes
master's full gate for every lane. A future `brew autoremove` could also remove the binary outright. The plan treats
the pin as stable between sweeps, but nothing on this workstation holds it there.

**Fold:** in Task 2 (or the lane launch), install libxml2 on request, add `brew "libxml2"` to the Brewfile under the
workstation's manifest rule, and run `brew pin libxml2`. `brew pin --help`: "Pin the specified package, preventing
it from being upgraded when issuing the brew upgrade formula or cask command". The refusal message names
`brew unpin libxml2 && brew upgrade libxml2` as the sweep's deliberate bump path. Add the pin to the close's
CLAUDE.md prerequisite line. This is a workstation method call, not an owner fork.

### M3 (major): "required on every row" pulls 27 test-ledger rows and about 32 test literals into Task 6, unnamed

**Location:** Task 6 `:561-566`, states `:579-581`, Files `:557-558`.

`divergence` and `test` become required on every row, and an `owed` row may not carry a `class`. But
`grade/testdata/seed-ledger.json` holds 27 active oracle-pair rows (volatile, contract, by-construction) that grade
the grader. The tests also build Rows inline: 22 `Row{` literals in `ledger_test.go`, 6 in `grade_test.go`, and 4 in
`cmd/harness/grade_test.go`. Several are owed rows with a class, for example `TestApplyOwedRowSuppressesNothing`.
None of these files is in Task 6's Files. The record's ten classes describe how dubplate differs. They have no
natural value for a Navidrome-against-gonic `server-type-ping` row, so the implementer would invent 27 values.

**Fold:** require `divergence` and `test` on `owed` rows only, the rows transcribed from the record. Keep them
optional on `active` rows until slice 3 adjudicates one against a dubplate session. Or name every file and the value
a test-only row takes. Either way, list `grade/testdata/seed-ledger.json`, `grade_test.go`, and
`cmd/harness/grade_test.go` in Task 6's Files.

### M4 (major): the assignment's 87 entries and every row's `slice` have no named source past section 3's table

**Location:** Task 5 `:518-524`, R13 `:250-254`, Task 6 `:567`, `:576-577`.

The study record's section 3 table covers about 60 endpoints. Several of its verdicts are split, while the
assignment holds one oracle per endpoint:

- `stream`: Both direct, Navidrome for transcode.
- `getOpenSubsonicExtensions`: Navidrome for the carve-out, contract only for the list.
- `getCoverArt`: Both for shape, contract for WebP.
- `star`/`unstar`: Both, "gonic's missing-id leniency excepted".
- `getPlayQueue`: Navidrome for the empty answer, contract for the rest.
- The share endpoints: Navidrome for shape, the spec for the bounds.

About 25 of the 87 paths appear nowhere in the table, among them `getVideos`, the podcast, chat, jukebox,
bookmark, and user-admin endpoints, `search` (v1), `getAvatar`, `hls`, and `getCaptions`. The "slice that first
serves it" column has no source at all, since section 3 gives no slices.

R13 then derives each ledger row's `slice` from that column, and its rule leaves rows undecided:

- L-06 names three endpoints.
- L-34's endpoints are `not-served-0.8`, which is not a slice.
- L-37 has no endpoint and is missing from R13's no-endpoint list.
- The agreement rows' slices have no rule (A-01, A-02, A-09, A-11, A-12 say only "Harness").

The plan calls `slice` privileged and "proved by the transcription test" (`:576-577`). The test cannot prove a
value the plan never fixed. This is the 5b-ii ambiguity class.

**Fold:**
- Give the unlisted endpoints a rule: contract only and `not-served-0.8` unless the spec's family list at `:657-661`
  serves them.
- Name a source for slice per family, the re-plan ladder.
- Pick the representation for a split verdict: a primary oracle plus a note, or per-mode entries.
- List the slice for each row R13's rule leaves open.

### M5 (major): the projection's line mapping and namespace handling do not behave as stated

**Location:** quotes table `:207-210`, R20 `:275-278`, Task 3 `:446-450`, proofs `:466-469`.

1. **gonic's bodies are indented.** 30 of gonic's 31 seed XML bodies span 2 to 60 lines, while all of Navidrome's
   are one line. A `RawToken` loop that passes whitespace `CharData` through `EncodeToken` under `Indent` emits it beside
   the encoder's own newlines. Probe on `getArtist` from gonic: 16 start tags became 41 lines, with blank lines
   between tags. With whitespace dropped it is 21 lines, because end tags take lines too. "One start tag per line"
   is not the same as "line k is the k-th element", and `Lyrics` and `Genre` are mixed content whose text can carry
   newlines. Every location proof in Task 3 (`:469`) and Task 4 (`:506-507`) uses a one-line or hand-written body.
   An implementer who maps line k to element k therefore passes every proof and still misplaces every gonic
   violation in the golden.
2. **`Encoder` rewrites a prefix.** `marshal.go:744-748` writes a non-empty `Name.Space` as `xmlns="<Space>"`, and
   `RawToken` leaves the prefix in `Space`. Probe: a valid prefixed body (`<sub:subsonic-response
   xmlns:sub="http://subsonic.org/restapi" ...>`) "validates" as written. Projected, it becomes `<subsonic-response
   xmlns="sub" xmlns:_xmlns="xmlns" _xmlns:sub="...">` and fails with "No matching global declaration available for
   the validation root". The `:207` quote ("keeps the body's namespace declarations as written") does not survive
   the re-encode. Task 7 treats a prefixed body as a real case (`:613-614`). Neither oracle and no Go encoder emits
   one today, so this half is minor on its own.

**Fold:**
- Drop whitespace-only `CharData` from the projection.
- Build the line-to-ordinal table from the bytes actually written: flush and count newlines before each start tag,
  or use a small own writer that emits `prefix:local` and the attributes as read.
- Add a Task 3 proof that places an error on the second of two `album` elements in a real indented gonic body.
- Add "prefixed body" to the states table: it validates, or it is declared unsupported with its reason.

### M6 (major): Task 9's Navidrome library creation names a tool the implementer lacks and skips a step

**Location:** Task 9 `:683-686`, conductor ruling `:153-154`.

`dubplate-implementer` carries Read, Write, Edit, Bash, Grep, and Glob, and no browser tool. "Observes that request
once in the browser's network log" therefore means writing a Playwright script (Playwright 1.63.0 and its Chromium
are in `web/node_modules`). The script would log in to Navidrome's UI, open Settings, then Libraries, then "+"
(multi-library page `:46-57`), and capture the traffic. The plan does not say so.

The plan also records only the creation request. Navidrome's native API needs the login exchange first, and that
exchange has to be observed and recorded too. Slice 1 did create a library "through its admin API" (`[servers]`
F2, `:31`) but recorded no method or path.

The fallback, "unmeasured" (`:153-154`), costs the primary oracle's `musicFolderId` evidence. Navidrome is the
primary for every album-list endpoint (study record section 3). Slice 3 consumes exactly this evidence (A-04, L-19).

**Fold:** state the route: a headless Playwright script run once by the implementer outside the gate, recording
both the login exchange and the creation request (method, path, headers' names, body shape, response). The stack
replays both. Also name the container paths, disjoint siblings such as `/collections/<name>` with the first as
`ND_MUSICFOLDER`, since the docs refuse a path that would "overlap with other libraries" (`:150`). Task 9's fake
oracle server then serves both routes.

## Minors

- **M7. Two-collection fixture details** (Task 9 `:675-679`, Task 10 `:717-725`).
  - `fixture.Load` reads seeds from `filepath.Join(filepath.Dir(path), "seeds")` (`manifest.go:83`). A manifest
    under `data/fixture/collections/` finds no seeds, so either put the two manifests beside `manifest.json` or let
    `Load` take the seeds directory.
  - The seed manifest carries no `genres` tag on any track, so "a genre the collection manifests carry" (`:725`)
    requires the collection manifests to add one. Otherwise `getSongsByGenre` is empty everywhere and the filter
    rows prove nothing. tagio supports `genres` (`encode.go:123`).
  - The collection key's "exactly the collection manifest's album artists" (`:717-718`) fails for the untagged
    album. The manifest says `(untagged)`, Navidrome lists `[Unknown Artist]`, and gonic lists `Unknown Artist`
    (seed `006-*.json`). Name the `albumartist` tag as the set and keep `Loose files` out of both collections.
  - Computing the combined digest means `grade` builds each library into a temp directory (`Build` is the only
    digest source). It is cheap (the fixture package tests run in 0.11 s) but worth one line.
- **M8. `xmllint --version` writes to stderr, not stdout** (`:443-444`). Probe: stdout empty, stderr `xmllint: using
  libxml version 21504`, exit 0. Quote it in the table. A fake that writes stdout, checked by a runner that reads
  stdout, passes its test while the real gate refuses.
- **M9. "`format: any` covers both" is false for one-element lists** (Task 4 `:490-492`). JSON paths always index a
  list (`pointerPath` gives `album/song/0/title`). `tree.go:249-257` indexes an XML name only when it repeats, so a
  one-song album's XML path is `album/song/title`, and a `*` segment matches only digits (`ledger.go:isIndex`). Drop
  the claim, or say the two paths agree only for lists of two or more.
- **M10. Task 7 makes 20 test-ledger rows stale** (`:606-608`, Files `:602-603`). `seed-ledger.json` has
  `server-type-*` and `server-version-*` volatile rows on ten endpoints. `TestSeedGrade` fails on `ErrStaleRow`
  before its golden compare. `TestGradeNamedSessionsReportsAStaleRow`'s "live" row is `ping`
  `subsonic-response/type` (`grade_test.go:470`). Name both files and re-point the live row at a still-volatile
  path such as `song/created`.
- **M11. The E5 proof is vacuous as written** (`:613`): "`&` against `&` in JSON" compares identical bytes. It
  presumably means `&` against `&`.
- **M12. Residual elements left in the projection mask later violations** (`:437-438`). Probe on gonic `getArtist`:
  after `roles: This element is not expected`, xmllint reported nothing on the following `album` or its
  attributes. Any non-residual error after a residual element (a bad type, a missing required attribute) is hidden.
  The body never passes, but the golden under-reports. Dropping each residual element from the projected document
  once it is recorded lets xmllint see the rest, and it also removes the need for the dedupe rule at `:449-450`.
- **M13. No XML-tier state for a non-XML answer** (Task 4 states `:501-502`). A binary body, a non-XML content type,
  or a non-200 status has no stated report. The JSON tier reports non-`Valid` outcomes at `envelopePath`
  (`grade.go:contractTier`). Mirror that.
- **M14. Task 0's rebase is not a fast-forward** (`:336-337`). The branch carries `0ae2b54` (this plan) ahead of
  master. Also, if the sweep moves `web/` dependencies, the worktree needs `npm ci --prefix web` after the rebase, or
  Task 1's first gate refuses a drifted install.
- **M15. Task 6's report change contradicts Task 5** (`:574` against `:532-533`). "Owed agreements by slice" in every
  report rewrites the seed golden, which Task 5 says an oracle pair never touches. Either print agreements only when
  the assignment is consulted, or list the seed golden in Task 6.
- **M16. Task 8's assertion contradicts itself** (`:658-659`): "the test asserts each one present", yet "an absent
  prediction is ... never a failure". Say the test asserts each prediction as recorded.
- **M17. Pass-execute's `tool` class defaults `gateLane` to `"light"`** (`pass-execute.js`, `PASS_CLASSES.tool`,
  and line 359: `t.gateLane || a.gateLane || cls.gateLane`). Task 1 must carry `gateLane: "heavy"` and its own full
  `gate` string in its task entry. If the conductor leaves either off, Task 1's Chromium gate runs under the 3G cap.
- **M18. Scope and fidelity gaps.**
  - Ladder entry 2 (spec `:885`) and the study record's "Routed" list route "the known-deviation list" to slice 2,
    and the scope table (`:82-93`) does not mention it. It is 2a's `contract` rows with the stale-row check, plus
    L-20 owed to slice 3; one row in the table says so.
  - L-51's `test` text ("only if fork C is yes") is stale under ruling C, and the fidelity exemption (`:590-591`)
    covers only the three `dubplate` texts. Name L-51's `test` text as ruled too.

## Over-ceremony

Nothing of consequence. Task 1's three quoted gate runs cost one extra heavy gate. Go's test cache absorbs the
conform suite on every scoped harness gate. Task 3's Opus upshift is justified by M5.

## Planning-miss checklist (CLAUDE.md)

- **Behavior quoted, not recalled:** mostly met. M1 (the `Output` quote omits truncation), M5 (the
  `RawToken`/`Encoder` pairing), and M8 (stderr) are the misses.
- **Every state names its report:** M13 (non-XML answer) and M5 (indented body, prefixed body) are missing.
- **Privileged fields name every channel:** met for the pins, the assignment, and the credential channels. M4 shows
  the `slice` channel names a value source that does not exist.
- **Proofs name their fixture state:** M5 (location proofs only on one-line bodies), M11 (identical bytes), and M7
  (no genre, the untagged artist) are the misses.
