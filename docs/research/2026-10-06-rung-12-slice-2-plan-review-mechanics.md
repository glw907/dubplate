# Rung 12 slice 2a plan review: mechanics and feasibility

Target: `docs/superpowers/plans/2026-10-06-rung-12-slice-2-oracle-harness.md` at `f62a151` (draft revision 1).
Lens: does every mechanism behave as stated? Every claim below was checked by a probe under the session scratchpad
(never in the repo tree), or by quoting docs or source. The probes ran both oracles live under podman on this
workstation (images pulled by digest; they stay in the user's container store for the lane).

Counts: 3 High, 5 Medium, 5 Low; 1 over-ceremony note; 1 OWNER FORK (inside M11).

## Verified as the plan states (no fold)

- **kin-openapi v0.149.0 is the newest release.** `proxy.golang.org/.../@latest` gives `v0.149.0`, 2026-08-28. GitHub
  API: MIT, pushed 2026-10-04, 103 open issues and PRs, 3,298 stars, not archived. Twelve releases from v0.138.0
  (2026-05-08) to v0.149.0 (2026-08-28).
- **The bundled OpenAPI file** fetched today: 453,720 bytes, sha256 `cb54c03c…ed46bfb6`, `openapi` 3.0.0,
  `info.version` 1.16.1, 87 paths, one server `{url}`. `doc.Validate` passes. Full-URL routing through
  `legacy.NewRouter` gives "no matching operation was found". Setting `doc.Servers = nil` and routing by path finds
  `/rest/getAlbum`. `/rest/getAlbum.view` finds no route, so Task 4's `.view` drop is needed. `application/json;
  charset=utf-8` is accepted. An empty body, a non-JSON body, and `text/xml` each give a `ResponseError`. The file
  holds no `?t=`/`&s=`/`&p=`/`apiKey=` string, so Task 9's acceptance grep stays clean.
- **Every stdlib and spec quote holds:** `encoding/json/v2.Unmarshal` (the `float64` sentence), `xml.Decoder.Token`
  (self-closing expansion, nesting), the token vector (`md5("sesamec19b2d")` = `26719a1196d2a940705a59634eb18eab`),
  and api-reference `bed1688` (salt "at least six characters", UTF-8, `f` values).
- **`go list -deps -test` format.** The probe output is `ex.com/m/a`, `ex.com/m/b`, `ex.com/m/a [ex.com/m/a.test]`,
  and `ex.com/m/a.test`, so cutting at the first space and checking the prefix works.
- **R1's premise.** The rule's stated reason is the venv. Parallel-build approach `:255` says: "`make derive` ... is
  deliberately outside `make check`, so the lane's gate goes green while the only leg exercising the beets oracle never
  ran ... **Policy: corpus and harness work never moves to a worktree**". This harness uses no venv. M12 names the one
  residue of the rule that still applies.
- **The gate runs in a worktree.** HISTORY records green full worktree gates for lanes 5a, 5b, 6a, 6b-i, 6b-ii, and
  6c, the last one after the rendered leg landed (5c-iii 3a). The worktree needs `npm ci --prefix web`. The Playwright
  browsers, the Unidecode wheel, `rclone`, and the Go cache are user-level and shared. One question was whether
  gitleaks' `.claude/worktrees/.*` allowlist hides a leak inside a worktree. A probe placed a fake token under a
  `.claude/worktrees/lane-x/` path and scanned with `-s .` and `-s <file>` from inside it, using the repo's config.
  Both scans reported "leaks found: 1", rc=1. The allowlist matches the scanned paths, which are relative, so it hides
  nothing inside the worktree.
- **The `go.mod` timing.** `go mod tidy` drops a requirement nothing imports: the scratch tidy removed every unimported
  require. So R9 holds that kin-openapi cannot land ahead of the branch. 6d `:1537-1539` reads as the plan quotes it.
- **The image references.** The Navidrome tag is `0.64.2` with no `v`. Both `docker.io/deluan/navidrome:0.64.2` and
  `ghcr.io/navidrome/navidrome:0.64.2` resolve to `sha256:38dc2727…a44ee52`, and `:v0.64.2` is "manifest unknown".
  gonic is `docker.io/sentriz/gonic:v0.22.0` → `sha256:516fd964…be4c8b1d49`, and `:0.22.0` is unknown. When the image
  is pulled by digest, `podman image inspect --format '{{.Digest}}'` returns that same digest. Both are newest
  (GitHub latest: v0.64.2 2026-09-24; v0.22.0 2026-06-08).
- **The cited `file:line` pointers resolve at `f62a151`.** These were checked: `cassette.go:1-14,41,260`,
  `arch_test.go` (the invariant at `:138`), `Makefile:134-146`, `analysis/ledger.go:10-14`, `cmd/harness/main.go:53-66`,
  `tagio/write.go:34-39`, `auth/roster.go:340`, and `template.flac` at 42 bytes. Small drift only.
- **Task order builds.** Task 7 needs Task 5's seed session and Task 8 needs Tasks 6 and 7, and the segments respect
  both dependencies.

## Correctness gaps, ranked by consequence

### M1. High. Every per-task gate launches Chromium on the light lane

**Where:** plan `:54-60`, "the light lane holds because no task touches `web/` … so `--scope-changed` skips the
rendered leg and launches no browser (`scripts/check.sh:26-40`)".

**Defect:** `check.sh` skips the rendered leg only when the conform suite is also skipped. On a branch, the change set
is everything since the merge base, so from Task 1 on, every task gate is conform-affected and runs the whole of
`make check`, the rendered leg included. Each task gate therefore launches a browser under the light lane's 3G cap
and its own lock. That bypasses the machine-wide heavy lock that lane 1 and the cairn session queue on, which is
exactly what `pass-gate-economy.md` built the lock to prevent ("so parallel chains never run two browser gates at
once").

**Evidence:** `scripts/check.sh:183` sets `change_base="$(git merge-base master HEAD …)"`. The code at `:270-275`
reads:

```
if [ "$mode" = "scope-changed" ] && [ "$conform_affected" -eq 0 ]; then
  scope_conform=skipped
  if [ "$rendered_affected" -eq 0 ]; then
    scope_rendered=skipped
```

Then `:364-367`: when the conform suite is not skipped, the script runs `make check`. `Makefile:134` defines
`check: web vet lint test web-rendered tidy-check`.

**Fold:** drop `CAIRN_GATE_LANE=light` and `gateLane: "light"` from the per-task gate and run the task gates on the
heavy lane. Correct the header sentence. Decoupling the rendered scope from the conform scope is one condition in
`check.sh`, which 6d contests, so file it with the existing "conform suite run time" item (`:624-625`) for after 6d.

### M2. High. Task 8's oracle-against-oracle gate test cannot pass honestly with four classes

**Where:** plan `:16-17` ("Navidrome and gonic disagree on the folder view by construction, and the grader must find
exactly the disagreements the ledger declares"), Task 6's classes at `:476-478`, and Task 8's gate test at `:551-554`
(zero unledgered findings; only `contract` or extra `volatile` rows may be added).

**Defect:** the two oracles disagree well beyond the folder view, on stable behavior that is neither per-run nor a
schema violation. `volatile` compares presence and type, so it cannot absorb a field one side omits. The implementer
must either misclassify dozens of rows or fail the test. The plan names no class for "the oracles disagree".

**Evidence:** a live probe on a three-track fixture, both oracles answering the same JSON request:

- `ping`: `version` is `"1.16.1"` from Navidrome and `"1.15.0"` from gonic. That is an envelope field the plan's
  volatile list (`serverVersion`, `type`, timestamps) does not name.
- `getArtists`: the index `name` is `"A"` from Navidrome and `"a"` from gonic, and `ignoredArticles` is
  `"The El La Los …"` against `""`. Navidrome alone sends `lastModified`, `musicBrainzId`, and `sortName`. `roles`
  is `[maincredit, albumartist, artist]` from Navidrome and `[albumartist, artist]` from gonic.
- `getAlbum` key paths: 21 are Navidrome-only, such as `album/sortName`, `album/userRating`, `song[]/bitDepth`,
  `song[]/samplingRate`, `song[]/mediaType`, and `song[]/works`. 4 are gonic-only: `album/album`, `album/title`,
  `album/playCount`, and `song[]/isVideo`.

**Fold:** add a fifth class, `oracle-split` (the oracles disagree; the row records both behaviors and dubplate's
intended one, and its scope is a path pattern). Restate the proof as "every finding ledgered, and every folder-view
finding `by-construction`". Expect the seed ledger to carry tens of rows, and price that into Task 8. Also state that
id pairing skips any subtree a `by-construction` row covers. Otherwise the folder view's divergent ids create
"pairs with two" conflicts that leak into the album and song endpoints.

### M3. High. Task 7's proof base, "the untouched body passes", is false for the real seed body

**Where:** plan `:526-528`.

**Defect:** the real `getAlbum` bodies from both pinned oracles violate the pinned schema. Built as written, the test
fails at its first assertion.

**Evidence:** kin-openapi v0.149.0, validated against the bundled file, routed by path:

- Navidrome: `Error at "/album": … Error at "/userRating": number must be at least 1` (value `0`; schema
  `minimum: 1, maximum: 5`).
- gonic: `Error at "/song/0/replayGain": … Value is not nullable`.
- The same Navidrome body with `"userRating":0` removed: `err=false`.

**Fold:** the proof asserts the untouched body's exact violation list, which then seeds Task 8's `contract` rows for
each oracle. The "passes" base becomes that body with the recorded violations repaired. A fixture state that cannot
pass vacuously needs a non-empty expected list.

### M4. Medium. The `MultiError` mutation is not killable with the named body, and violation reports carry branch noise

**Where:** plan `:114` (MultiError for many-bad) and `:529-530` ("drop `MultiError` (the two-violation body reports
one and the test fails)"), where the two-violation body is "`song` removed" plus "numeric `id`".

**Defect:**

1. Without `MultiError`, kin still reports both of those violations, so the mutation survives.
2. Every 200 response's schema is a success-or-failure union. Each violation report therefore also carries the
   failed branch's leaves, and `MultiError` duplicates some leaves. "Every violation with its JSON path" would report
   noise.

**Evidence:** leaf `SchemaError`s were walked with `errors.As`, each reported as `JSONPointer()` plus `Reason`.

| Body | `MultiError` | Leaves |
|---|---|---|
| `song` removed + `album.id` numeric | on | 4: `album/id`, `album/song`, `status: not one of ["failed"]`, `error: property missing` |
| same | off | 3: `album/id`, `album/song`, `status…` (both real violations still reported) |
| `album.id` + `song[0].id` numeric | on | 5, with `song/0/id` twice |
| `album.id` + `album.name` numeric | on | `album/id`, `album/name`, plus the two failed-branch leaves |
| same | off | `album/id` and `status…` only, so the mutation dies |

**Fold:**

- The checker reads the body's `status` and reports only the matching branch's leaves, deduplicated by pointer and
  reason.
- The two-violation proof body puts two type violations on sibling scalars of one object (`album.id` and
  `album.name` numeric).
- The required-`song` case stays as its own single-violation row.

### M5. Medium. Readiness passes on an empty library

**Where:** plan `:383-384` ("the scan has finished, read through `getScanStatus`") and the states table `:393-396`.

**Defect:** a scan over an unreadable or not-yet-scanned library reports finished with zero tracks. Readiness then
passes, and every recorded response describes an empty library. The state where the scan has not started also reads
`scanning: false`.

**Evidence:** in the first probe run the mount was unreadable (see M6). Navidrome logged `Scanner: Target folder does
not exist … permission denied` and then `Scan completed`. gonic logged `finished scan … +0/0 tracks (1 err)`. After
the fix, `getScanStatus` gives `{"scanning":false,"count":3,…}` on Navidrome and `{"scanning":false,"count":3}` on
gonic.

**Fold:** readiness means `scanning == false` and `count` equal to the manifest's track count, before the deadline.
Add a states row: "scan finished with a count other than the manifest's: an error naming the oracle and both counts".
Add a mutation: accept `count: 0`.

### M6. Medium. On this workstation, a bare read-only mount is unreadable, and gonic needs four paths present

**Where:** plan `:374-376` and the argv test at `:398-400`.

**Defect and evidence:**

- The workstation runs SELinux `Enforcing`. With `-v "$L:/music:ro"`, both containers got `open /music: permission
  denied`. With `-v "$L:/music:ro,z"`, both scanned 3 of 3 tracks.
- gonic exits 1 with `checking playlist directory: path does not exist, please provide one` unless `/playlists`
  exists. The image env sets `GONIC_DB_PATH=/data/gonic.db`, `GONIC_PODCAST_PATH=/podcasts`,
  `GONIC_CACHE_PATH=/cache`, and `GONIC_PLAYLISTS_PATH=/playlists`. `--tmpfs` for `/data`, `/cache`, `/podcasts`, and
  `/playlists` fixed it.

**Fold:** the argv test asserts `:ro,z` on the library mount and gonic's four tmpfs paths. The "omit `:ro`" mutation
stands as written.

### M7. Medium. Closes "Not verified": Navidrome's first admin, gonic's credential, and token auth

**Where:** plan `:380-382`, `:386-387`, and `:618-620`. Also R4 at `:186-187`.

**Evidence:**

- **Navidrome.** Its Apache-2.0 docs site, `content/en/docs/developers/dev-environment.md:39-41`, says: "This option
  will always create an `admin` user with the specified password … `DevAutoCreateAdminPassword = "password"`". The
  options page `:125` says "Any configuration option can be set as an environment variable, just add … the prefix
  `ND_`". Live, `ND_DEVAUTOCREATEADMINPASSWORD` created user `admin` with the log line `Creating initial admin user.
  This should only be used for development purposes!! … password=s3cretpw user=admin`. The `user create` CLI page
  says "On a new install, create the first admin in the web UI" and takes no password flag, so the CLI is no path.
- **gonic.** README `:31-32` says "the default login is **admin**/**admin**. password can then be changed from the web
  interface". No environment variable sets it.
- **Token auth.** Both oracles accept `u`/`t`/`s` token auth: `ping` returned `"status":"ok"` from Navidrome
  (`0.64.2 (10114574)`) and from gonic (`0.22.0`), and a wrong token returned code 40 from Navidrome. gonic README
  `:26` says "newer salt and token auth".

**Defect:** "the admin credential is generated per run and never written to disk" cannot hold. gonic has no
unattended way to set its password. Navidrome prints the password in its container log, and podman keeps the env in
on-disk container config. The readiness-failure path (`:394-395`, "each container's last log lines") would print the
credential.

**Fold:**

- Navidrome takes `ND_DEVAUTOCREATEADMINPASSWORD` with a per-run value and user `admin`.
- gonic keeps `admin`/`admin`, bound to loopback in a throwaway container.
- Reword the credential rule: "never written to the repo, a session, or a report". Log tails redact the password.
- R4's `p=enc:` branch closes now: no oracle refuses tokens. Drop the conditional from Task 3.

### M8. Medium. gonic ignores the fixture's multi-valued artist unless configured

**Where:** Task 1 `:319-320` (the featured artist is a multi-valued artist tag) and Task 3's environment list `:379`.

**Evidence:** gonic README `:72-74,92`: `GONIC_MULTI_VALUE_ARTIST` and `GONIC_MULTI_VALUE_ALBUM_ARTIST` default to
`none` ("gonic will not attempt to do any multi value processing"). `multi` reads the fields Picard and beets set.

**Fold:** Task 3 sets both variables to `multi`, and the argv test asserts them. Otherwise every gonic response that
touches the featured track records a configuration artifact as oracle behavior.

### M9. Low. The checker must resolve the response object itself, not read kin's error text

**Where:** plan `:115` and Task 7's four outcomes at `:511-513`.

**Evidence:** in `openapi3filter/validate_response.go` at v0.149.0, an undeclared status with
`IncludeResponseStatus` returns `&ResponseError{Input: input, Reason: "status is not supported"}` (`:54-57`). That
error has no sentinel, and without the option the call returns nil. Empty content returns nil (`:99-102`). A nil
schema returns nil (`:114-117`). The plan's cite `:96-112` has drifted slightly.

**Fold:** the checker looks up `route.Operation.Responses.Status(code)` and the content entry before it calls
`ValidateResponse`, and classifies "no schema" from that lookup. It never matches on `Reason` strings.

### M10. Low. Conformance tests would read files across the boundary they defend

**Where:** R3 `:180-181` (the pinned file under `internal/harness/subsonic/data/openapi/`) and Task 7 `:526` (proofs
from the seed session's body).

**Defect:** `internal/opensubsonic/conformance`'s tests need the pinned document and a real body. Both live under
`internal/harness`. Task 2's leg checks imports, so it passes. The layer meant for extraction after 0.8 (spec
`:312-315`) would still ship tests that reach outside its subtree.

**Fold:** keep the pinned file, its `LICENSE`, and its provenance under the protocol subtree (for example
`internal/opensubsonic/conformance/openapi/`), where the harness reads it, which is the allowed direction. Copy the
proof `getAlbum` body into `conformance/testdata/`.

### M11. Low. kin-openapi's indirect dependencies are not at their newest (contains an OWNER FORK)

**Where:** plan `:151-154` and Task 7 `:520`.

**Evidence:** `go list -m -u` shows `go-openapi/jsonpointer v0.22.5 [v1.0.2]` and `go-openapi/swag/jsonname v0.25.5
[v0.29.2]`. Tidy also adds `kr/pretty v0.3.1 // indirect`, which `go mod why` traces to `oasdiff/yaml3.test`. With
both bumped, the scratch module builds and the validation probe gives identical results.

**OWNER FORK:** under the repo's rule, `jsonpointer` going from v0 to v1 may count as a major.

- (a) Take v1.0.2 now. The import path is unchanged and the bump is verified.
- (b) Hold at kin's minimum and record the unblock condition.

Recommendation: (a), asked in the batched checkpoint. Task 7 names each indirect version it lands.

### M12. Low. The worktree never builds the oracle-tagged `cmd/harness` test

**Where:** R1 `:162-169`.

**Evidence:** `cmd/harness/freeze_derive_test.go` is `//go:build oracle` in package `main`, and it declares
`venvOptions` and `copyCaseDir`. Tasks 3, 5, and 8 add files to that package. The gate never builds the tag, and
`make derive` cannot run without the venv. This is the one part of the "never moves to a worktree" rule that still
applies here.

**Fold:** once at the close, the conductor runs `go vet -tags oracle ./cmd/harness/... ./internal/harness/...`. It
compiles the tagged code without the venv. Add the check to R1's queued `CLAUDE.md` wording.

### M13. Low. Task 0's sweep would write into lane 1's tree while 6d runs

**Where:** R9 `:209-211` (the sweep "runs on master in the main tree … only while no 6d sweep is in flight").

**Defect:** after 6d's own Task 0, its executor still owns the main tree. A conductor sweep commit there would put a
second executor in that tree, which `CLAUDE.md:107-110` forbids. The same holds for Task 9's `--ff-only` merge, and
R9 already windows that.

**Fold:** the sweep runs only when the lane registry shows lane 1 idle. Otherwise this pass adopts 6d's sweep record
in one line. State that the finished 2a branch may wait in its worktree for 6d's merge.

## Over-ceremony, ranked by cost

- **O1. The rendered leg on every task gate**, about 106s plus heavy-lock queueing, eight times. This is the cost
  side of M1. Nothing in 2a can change what the browser sees. The fix is the one-condition `check.sh` decoupling,
  which 6d contests and so is filed for after 6d. Until then the cost is accepted. No other ceremony in the plan costs
  enough to cut.

## Planning-miss checklist

1. **Behaviors quoted.** The quotes are accurate. Two things the tests rely on were inferred rather than probed, and
   both are wrong: `MultiError` changes the reported count for the named body (M4), and real oracle bodies pass the
   schema (M3).
2. **States.** The readiness check is missing the "finished with the wrong count" state (M5).
3. **Privileged fields.** The oracle version (set only from the envelope, by the recorder) and the digests (read from
   `podman image inspect`) each name their channel, and the digest read was verified. The credential's channels are
   misstated (M7).
4. **Fixture states.** Task 7's base fixture is wrong (M3), Task 3's readiness check can pass vacuously (M5), and
   Task 8's oracle pair needs a ledger taxonomy that can hold its real findings (M2).
