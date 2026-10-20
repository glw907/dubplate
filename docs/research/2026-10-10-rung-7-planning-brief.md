# Rung 7 (intake and the uploader): planning brief (2026-10-10)

This brief prepares the planning lane for rung 7, which Geoff approved on 2026-10-10 to run beside rung 14a. It
was written against master at `e6fdf2b`. Every claim about the tree cites a file and line read at that commit, and
every version was read from its registry on 2026-10-10. The planning session should still re-read any fact it
builds on (CLAUDE.md, "Planning-miss prevention").

Abbreviations: BS is the build spec (`specs/2026-09-02-greenfield-server-build.md`), RP the re-plan
(`specs/2026-10-05-rungs-12-16-replan-design.md`), SS the streamer spec (`specs/2026-09-07-streamer-2-0-design.md`),
CH the configuration homes spec (`specs/2026-09-11-configuration-homes-design.md`), R6 the rung 6 spec
(`specs/2026-09-23-rung-6-ops-seam-store.md`), PA the parallel build approach
(`specs/2026-09-03-parallel-build-approach.md`), RE the reverse-edge analysis
(`research/2026-09-04-reverse-edge-completion.md`), PS the group-scale principles spec
(`specs/2026-09-25-group-scale-principles-design.md`), VI the visual identity spec
(`specs/2026-09-24-visual-identity-design.md`), W the web surface standard (`specs/2026-09-12-web-surface-standard.md`),
UD the salvage-era uploader design (`specs/2026-08-31-dubplate-uploader-design.md`), OD the organizer spec
(`specs/2026-09-01-dubplate-organizer-design.md`), PCF the reject and uploader string authority
(`research/passb-copy-final.md`), and the 14 brief the rung 14 planning brief
(`research/2026-10-10-next-pass-planning-brief.md`). Paths sit under `docs/` unless they start at the
repo root.

Rung 9's concurrency model was ruled on 2026-10-10: a per-album lease. The ruling is not yet in the tree at
`e6fdf2b` (no spec or ledger names it), so the planning session records it as an erratum before either lane's plan
cites it. Rung 7 does not depend on it.

## 1. Scope

### The ladder entry

BS:798-818: "tus transport, the gate policy, roster loading with the birth-requirement validation, journey records,
the reject copy contract adopted verbatim from the string authority with register-gated changes, and the uploader
reference web client built from the salvaged components under fresh review." Format acceptance is configuration: an
allowlist in the gate policy, FLAC-only for this estate, "BOUNDED by rung 8's tag I/O support", so "the gate never
admits a format the server cannot tag" (BS:802-811). "The album is the verdict unit: one failing track rejects the
batch, and no partial album is ever admitted" (BS:812-814).

The entry has been trimmed twice since. Rung 5's close moved the `web/` gate leg and the roster to rung 5: "Rung 7
consumes it" (BS:869-878). The 2026-09-24 erratum retired the visual-fidelity chain, so the before/after now means
"Geoff's design review of dubplate's own design, rendered and read first, with no reference screenshots" (BS:20-25).
The 2026-09-07 erratum retired the streamer-matching theme roster (BS:5-18), and VI's two palettes, `label` and
`label-night`, replaced it (VI:52). RP:107 leaves rung 7 "unchanged in scope", and RP:143-144 moves every attended
web check, "rung 7's uploader before/after among them", onto the workstation or `ssh -L` until slice 8 opens a
public port.

BS also leaves two research items to this rung. The allowlist's shipped default "derives from what household
uploaders actually hold, researched at rung 7 within the tag I/O bound" (BS:329-331). How DaisyUI's built-ins fit the
two-axis theme model "is a rung 7 design detail" (BS:865-866), which VI's palettes may already settle.

### Routed items, with their origins

From the configuration homes spec (CH:717-724, "Rung 7 and 9 pipeline"), split by first consumer:

1. The audio key computed by decode at the gate, with the file hash beside it (CH:420-427). An all-zero header MD5
   "is absent, never a value". Album sameness is the ordered set of track audio keys, and "Re-upload refusal is
   album-keyed" (CH:433-435). The refusal answers only over collections the member can see (CH:651).
2. The intake caps: "upload size, per-album file count, decompressed total, path depth, metadata and picture size"
   (CH:722).
3. The same-device check for staging, library, and collection roots (CH:472-475, CH:724).
4. Quota enforcement as "a reservation taken before bytes move, with a true-up at completion", plus a hard
   free-space floor before every write (CH:554-558).
5. The default-collection fallback: "The uploader resolves the preference when the page loads and re-validates at
   write", falling back to the admin-owned default and saying so on screen (CH:560-565).
6. The template grammar and validator, the component sanitizer, root-confined writes, the injectivity check, and the
   break-link rule (CH:306, CH:719-720). Their first caller is the layout write, which is rung 9's (D9).

From the rung 6 spec:

7. "No default collection row: the first rung that writes a collection (rung 7 or rung 10) creates it, owned by the
   first admin" (R6:185-186). The root-allocation function and the no-nesting check "land at rung 7 or 9" (R6:181).
8. The Go-side path validator "lands with rung 7's first writer" (R6:567), and "Rung 7's gate rules on non-UTF-8
   filenames and canonicalizes symlinks once" (R6:562-563).

From the principles spec (all twelve rulings ruled, 4 to 12 as recommended, PS:3):

9. The source question: seven options, asked once per upload with a per-album override, carried as "an intake API
   field, not a page-only control" (PS:510-533). Ruling 6 made it required, so intake refuses a missing answer in
   reject wording (PS:752-756; acceptance row A9 at PS:627).
10. `added_by` from the session's member on the uploader channel (PS:346), with "No client field, API parameter, or
    admin form" writing it (PS:355-356; row A10 at PS:628).
11. Buy links under ruling 7: "rung 7, if a rung 7 plan task can absorb it without growing the pass", and the
    uploader captures the optional store link from the source question whichever way that goes (PS:757-764).
12. The route inventory, "The next rung that registers a route" (PS:581; row A2 at PS:620). No inventory test exists
    at HEAD (grep for `route inventory`, `routeInventory`, and `authClass` over `cmd` and `internal` is empty).

From the visual identity spec and the 5c passes:

13. VI's "Open for rung 7" list (VI:350-372): the phone header, error alerts as a tinted panel, the library list
    view, library filtering, the album page with its run-out inscription naming who added it, phone layouts, and
    motion. The catalog line's exact format is rung 7's (VI:243-244). The empty library pairs a photo with "No albums
    yet. Upload the first one." (VI:274-275).
14. 5c-iv's "To rung 7" list (HISTORY:1505-1522): the outline, dash, and ghost button text fix; 5c-iii's three latent
    items; the favicon paths named twice; the styled 404 with the admin gate's refusal through the same handler,
    under a `web-auth-security-reviewer` read (also HISTORY:1366 and 5c-iv-E3 at HISTORY:1483-1487); the theme toggle
    following an OS scheme change; a Continue link and the header wordmark's target "once a member page exists"; the
    masked "Sign in as a***@" line; and the phone gutter `clamp()`. HISTORY:1445-1448 adds the sleeve, grain, rim,
    no-art sleeve, Permanent Marker, the photo column, W12.4's remaining rules, W4.10's 44px promotion, and both
    palettes' placeholder opacity.
15. 5c-iii's svelte S13 and the keyline observation, "recorded for the rung 7 uploader's attended design review"
    (HISTORY:2477-2478).
16. W's own rung 7 items: W mentions rung 7 45 times. Two are structural. W7.19 names the uploader's four
    surfaces rung 7 owns (the drop target's states, the picker route as the iPad path, a multi-file progress
    surface, and a collection picker that hides at one collection). W11.5 makes a fresh-context coherence read
    "rung 7's uploader gate" (W:1298-1310). The plan sweeps the rest by grep.

From the photo and project-site specs:

17. The photo loader as a leaf package, porting `tools/photos/check.mjs`'s two modes into a Go gate leg
    (`specs/2026-09-25-photo-curation-design.md:236-247`), with the photo column, error-page photos, empty-library
    photos, and CREDITS generation "to rung 7 or later" (`:400-403`).
18. An enumerable reject-reason set with "a canonical wording for each parameterized reason", whose named consumer
    is the project site's reject-page generator (`specs/2026-09-25-project-site-design.md:301-306`), and the CREDITS
    generator with name-only lines for pack grantors (`:170-184`).

From STATUS and HISTORY:

19. STATUS:59 points at HISTORY's slice 1 close, whose carries read "guard counting, `--max-delete`, archive
    uploads, 5c-iii's S13/a11y S7, reject-reason set, CREDITS, picked images at stable paths, 6c's `member_quota`
    reservations, 5c-iv's handoff" (HISTORY:560-562). Their sources: the backup guard's per-collection and
    per-extension counting, the `--max-delete` sizing, and a remote-pruning regression floor (6b-i,
    HISTORY:3741-3744); the `member_quota` reservations and the default-collection preference (6c,
    HISTORY:1870-1873); archive uploads, from 4a's coverage gap that "An archive upload is invisible to the capture"
    (HISTORY:8494-8497).
20. The STATUS open item "First migration after `0005`" (STATUS:45-46): R-6d4's keep limit and the older-store helper's
    table map. It binds whichever of rung 7 and rung 14a migrates first.

From the organizer and uploader specs, which the greenfield inherits:

21. The user-present match probe. OD:541-546 has the uploader hold "a contributor on the page through those lookups",
    and the strict custody policy pushes back an album the matcher cannot place (OD:295-318, UD:231-250). The
    rung 4 shadow's budget class is "deferred behind rung 7's user-present probe" (HISTORY:9370-9371).
22. The journey vocabulary (PA:182, "rung 7 | rung 9 adds one state"). UD:260 lists "uploading, checking, matching,
    importing, in-library, pushed-back, queued, stalled", and UD:261-262 retires rejection emails for uploader
    sessions. Rung 13 later adds "the appeal action on rung 7's push-back page" (RE:263, OD:330-335).

From the Melodee study (`research/2026-10-06-slice-1-melodee-intake.md`): six lessons carry tests owed to rung 7. L2
(tag completeness is a pre-check, `:334`), L3 (disc and track are a pair, `:342`), L4 ("A check that cannot run is a
rejection, never a pass", `:350-353`), L7 (never merge into an existing album, `:371-376`), L9 (a closed reason set,
`:383`), and L12 ("The verdict belongs at the door", `:407-411`).

From IDEAS: no member request concerns uploading. The Requests idea (`IDEAS.md:21-24`) "Builds on intake's album
matching", so the submission's audio keys should stay queryable, but nothing is owed to it.

Three carries look misrouted and the plan should rule each out in one line. The backup guard's counting and
`--max-delete` sizing (item 19) belong to whatever first writes catalog tracks, which is rung 9's admission, not
rung 7. The archive-upload carry came from a shadow capture gap, and the shadow is retired (RP:126-130). Uploading
a ZIP stays in scope on its own evidence (D4).

### Forward hooks (SS:945-1040) against rung 7's surface

RP:110-111 names hooks 1, 12, 17, and 21 for rung 7.

| Hook | Verdict | Why |
|---|---|---|
| 1, audio key and file hash (SS:950-956) | lands here | computed at the gate; the columns "migrate in with rungs 7 and 9" (SS:990-995), so rung 7's submission rows carry them and rung 9 copies them to `tracks` |
| 17, the uploading member (SS:1012-1016) | lands here | the submission records the session's `UserID`; rung 9 carries it to the album row |
| 21, the source answer (SS:1027-1032) | lands here | an intake API field, required |
| 4, event-shaped tables (SS:961-962) | lands here | journey records are events and carry `at` and `kind` (store invariant 3, `internal/engine/store/invariant_test.go:329-346`) |
| 12, portable layout and `cover.jpg` (SS:982-984) | re-route to rung 9 | rung 7 lays out no library file; the sanitizer's first caller is rung 9's layout write (D9) |
| 5, 8, 9, 15 | consumed | `UserID`, the `writer` column, the single-writer transaction, and the collection row's fields; private stays unsettable (SS:1004-1005) |
| 22, integer wire id (SS:1033-1040) | neutral | "a rung that creates collections before slice 3 is neutral, because slice 3's migration backfills" |
| 18, description and sensitive mark (SS:1017-1019) | neutral, unless 14a lands first | if 14a's registration refuses a key without a description, rung 7's new keys carry one |
| 2, 3, 11, 13, 14, 16, 19, 20 | neutral | each names rung 8, 9, 13, slice 7, or slice 8 |

## 2. Contested surfaces with rung 14a

PA:173: "A contested surface without a written owner blocks the second lane." Each row below goes into the lane
registry before lane I's code launches.

| Surface | Where | Proposed owner |
|---|---|---|
| Migration `0006`, the older-store map, R-6d4's keep limit | `internal/engine/store/migrations/` (highest is `0005_routed.sql`); `cmd/dubplate-server/olderstore_test.go:28-33`; STATUS:45-46 | the first lane to merge takes `0006` and the STATUS item; the other renumbers on rebase and adds its map row |
| `serve` wiring | `cmd/dubplate-server/serve.go:345-367`, registration through `Shell.Register` (`internal/serve/serve.go:84`) | each lane adds one call to a helper defined in its own file |
| The web document table | `webDocuments` at `cmd/dubplate-server/web.go:104-108` (one row per page) and `webRoutes.register` at `:185-195` | each lane appends its own rows; a rebase conflict here is mechanical |
| `web/` shared modules | `web/src/lib/api/post.ts` (BACKLOG #23 at `:65`), `web/src/lib/copy/deck.ts`, `web/scripts/gates/deck-coverage.ts:150-170` (BACKLOG #21), the rendered harness `web/scripts/rendered/run.ts`, `web/src/lib/components/SiteHeader.svelte` | lane I owns `web/` (PA:113, PA:210); 14a's pages land as consumers, and whichever lane first touches `post.ts` takes #23 |
| `internal/auth` | PA:206: "rung 5, created by it; no other lane writes it" | needs a conductor ruling: rung 7 adds a member-level gate beside `RequireAdmin` and the styled 404's hook, and 14a adds removal and the epoch bump |
| `internal/serve` | PA:138 and PA:205: rung 12 owns it "without a concurrent serve-writing lane" | needs a conductor ruling: the styled 404 edits `Shell.terminated` (`internal/serve/serve.go:248-258`), and 14a may terminate `/api/v1` there |
| The journey vocabulary's wire projection | RE:212 and RE:223 | rung 7 owns the vocabulary; 14a's contract document rules that a new enum value is additive |
| Config registration shape | hook 18 (14 brief, Task 3) | if 14a merges first, rung 7's keys carry its description field |
| The route inventory (PS:581, A2) | no test at HEAD | the first lane to register a route lands the inventory; the other adds rows |
| Dependency manifests | `go.mod`, `web/package.json` | one lane runs the pass-opening sweep; rung 7 adds tusd, `mewkiz/flac`, and `tus-js-client` |

Two surfaces are rung 7's alone but sit in shared territory. The staging leaf touches `systemd/`, `scripts/deploy.sh`,
and the state-dir leg (PA:203 lists rung 7 among those files' consumers). Retiring the `intake` reserved class
edits `internal/config/keys.go:50-57`, which follows the rung 5 precedent of retiring a class in the commit that
declares its replacement (PA, the `internal/config/keys.go` edits row).

Sequencing: rung 7's JSON endpoints belong under `/api/v1` once it exists. If 14a's contract-and-mount task merges
before 7a's submission API task, rung 7 adds contract rows. Otherwise rung 7 mounts its own subtree, and 14a adopts
it (D12).

## 3. Prior art and dependencies

Each entry is judged on license, code quality against the house bar, maintenance, and adoption (CLAUDE.md, "Build
for 0.8").

**`github.com/tus/tusd/v2` v2.10.1** (proxy `@latest`, 2026-09-14; GitHub release 2026-09-16). It is MIT
(`LICENSE.txt`, "Copyright (c) 2013-2017 Transloadit Ltd and Contributors"). Five releases fall in the last twelve
months (v2.9.0 to v2.10.1). The repo has 3,892 stars, 85 open issues, and a last push on 2026-10-07. pkg.go.dev
lists 71 known importers of `pkg/handler`. Its top contributor holds 813 commits, a bus factor the Transloadit
backing offsets. The salvage uploader adopted v2.10.0 (`git show pre-greenfield:go.mod`, line 9) through `filestore`
and `memorylocker` (`_salvage/internal/serve/upload.go:28-30`). A scratch module importing `handler`, `filestore`,
and `filelocker` pulls in only `golang.org/x/exp` and `github.com/tus/lockfile` beyond tusd itself. Three findings
bear on quality:

- `handler.Config.Logger` is a `golang.org/x/exp/slog` logger (`pkg/handler/config.go:9`, `:59-60`), a type
  distinct from `log/slog`'s. Issue #1159, "slog: replace golang.org/x/exp/slog with log/slog", has been open since
  2024-07-24, and `main` still imports `x/exp/slog`. The plan rules between a forwarding handler and a discard
  logger with dubplate's own logging in the hook callbacks.
- `DefaultCorsConfig` allows every origin (`AllowOrigin: regexp.MustCompile(".*")`, `config.go:143-145`), and GET
  downloads are mounted unless `DisableDownload` is set (`:30-32`). The mount must disable CORS, downloads, and
  concatenation, since the shell's origin check is the one CSRF defense (`internal/serve/origin.go:1-35`).
- `PreUploadCreateCallback`, `PreFinishResponseCallback`, and `PreUploadTerminateCallback` (`config.go:65-87`) are where authorization, the quota
  reservation, and the gate attach.

Verdict: adopt. tusd is the protocol's reference server, and tus-js-client is tested against it. A hand-rolled tus
core would re-derive offsets, locking, and resume semantics that tusd ships.

**`tus-js-client` 4.3.1** (npm `latest`, published 2025-01-16), MIT, 2,627 stars, 77 open
issues, last push 2026-09-30. It is the newest 4.3.x and the salvage pin (`_salvage/web/package.json`). `next` is
5.0.0-pre3 (2026-09-30), which updates BACKLOG #8's watch item from pre2. Pin 4.3.1 exactly, per CLAUDE.md's 4.3.x
rule. Its runtime dependencies are seven small modules (`is-stream`, `js-base64`, `url-parse`, `buffer-from`,
`combine-errors`, `lodash.throttle`, `proper-lockfile`).

**`github.com/mewkiz/flac` v1.0.14** (2026-08-19, still the newest on 2026-10-10). Unlicense. The SS prior-art row
already rules "Adopt for rung 7. It supplies audio-frame decode, which the repo's own FLAC reader deliberately
lacks" (SS:1062), and the rung 8 spec routes decode here (`specs/2026-09-07-rung-8-tag-io.md:997`). The plan records
that consult in one line. It adds `icza/bitio` v1.1.0 (Apache-2.0), `mewkiz/pkg`, and `mewpkg/term`, and
`go mod tidy` picks `mewkiz/pkg` at a 2025-04-17 pseudo-version where the newest is 2026-07-03, so the sweep weighs
that indirect.

**Salvage**, admitted on merit (BS:508-516):

- `_salvage/internal/intake/intake.go` (564 lines with tests): a ZIP central-directory walk "that never extracts a
  member" and a STREAMINFO parse (`:1-8`). Adopt the walk's shape for D4.
- `_salvage/internal/gate/gate.go`: decode integrity by `exec.CommandContext(ctx, flacPath, "-t", path)` and
  `metaflac` (`:116-123`). Decline. The serve unit records "serve runs no subprocess"
  (`systemd/dubplate-server-serve.service:68`), and `mewkiz/flac` replaces both.
- `_salvage/internal/journey/journey.go`: coupled to the retired ledger and a scan trigger (`:1-14`). Decline the
  structure and keep the state names as input to D7.
- `_salvage/internal/reject/reject.go`: an `iota` reason set with a `Friendly` switch (`:17-60`) and the wire
  `label` PCF adds (PCF:20-23). Adopt the strings verbatim and restructure the set as enumerable (D11).
- `_salvage/internal/serve/upload.go`: the tusd engine and session grouping (7,553 lines across the package). It
  reads Cloudflare Access identity (`:6-11`), which is retired, so adopt the session-grouping idea only.
- `_salvage/web/src/lib/tus.ts` (386 lines): a driver interface with a retry ladder and a chunk downshift on 502,
  504, and 524 (`:12-25`). 524 is Cloudflare's status, and the tunnel is deleted (CLAUDE.md, Cloudflare), so the
  plan rules that branch. Also `preflight.ts`, `platform.ts`, `wakeLock.svelte.ts`, the upload store, and 13
  components with tests under `_salvage/web/src/lib/components/` (`DropZone`, `JourneySteps`, `CollectionPicker`,
  and others). They predate SvelteKit static mode and VI, so each is source material, never a copy.
- Salvage defaults: a 2 GiB album cap, a 5 GB free-space floor, and three sessions per member
  (`_salvage/internal/config/config.go:153`, `:176-181`). These are inputs to the defaults research, not defaults.

**Stdlib:** `archive/zip` for the walk, `crypto/md5` for the audio key over decoded samples, `crypto/sha256` for the
file hash, and `golang.org/x/text/unicode/norm` (already an indirect at v0.42.0, newest v0.43.0) if the non-UTF-8
policy normalizes.

**The other CLAUDE.md prior-art entries.** OpenSubsonic, Navidrome, gonic, taglib via wazero, Picard's tables,
chromaprint, goose, and modernc sqlite are neutral for rung 7 beyond the store it already uses. The plan says so in
one line. The Melodee study is consulted prior art (section 1).

**The pass-opening sweep** (whichever lane runs it). Registry reads on 2026-10-10 show these movers.

- `go.mod`: `golang.org/x/sys` v0.48.0 to v0.49.0, the indirect `x/crypto` v0.57.0 to v0.58.0 (as the 14 brief
  found), and the indirect `x/text` v0.42.0 to v0.43.0.
- `web/package.json`: `playwright` 1.63.0 to 1.64.0 (also a pinned browser binary) and `vite` 8.3.3 to 8.3.4.
- Majors, which wait for Geoff: `typescript` 6.0.3 to 7.0.2, and `@types/node` 24.19.1 to 26.6.5. The second
  tracks the Node runtime's major, not a release cadence, so it may not be a bump at all.

The other `web/` pins are current.

## 4. Facts the plan relies on, verified at `e6fdf2b`

Tree shape:

- No `internal/intake`, `internal/reject`, or `internal/engine/pipeline` exists. `internal/` holds `atomicfile`,
  `auth`, `backup`, `config`, `engine`, `harness`, `identity`, `notify`, `opensubsonic`, `ops`, `serve`, and
  `settings`.
- `go.mod` requires neither tusd nor `mewkiz/flac`. `internal/engine/tagio` parses FLAC containers and decodes no
  frame (rung 8 spec `:997`).
- `web/src/routes/` holds `+error.svelte`, `+layout.svelte`, `+layout.ts`, `login`, and `members`. No member-level
  page exists.
- `photos/` holds 233 entries (records and images), and `tools/photos/` holds `check.mjs`, its test, and the sorter.

Store:

- `collections` (`internal/engine/store/migrations/0003_catalog.sql:12-34`) carries `owner BLOB NOT NULL REFERENCES
  users (id)`, `visibility` in `('group', 'private')`, `writer` in `('server', 'local')`, `layout_template TEXT NOT
  NULL DEFAULT ''`, and a per-collection `admission_format_allowlist TEXT`. `root_path` is a BLOB under the
  relative-path CHECKs (`:15-22`). Nothing in `internal/` or `cmd/` inserts a collection.
- `tracks` (`:46-62`) has no audio-key or file-hash column.
- `member_overrides` (`0005_routed.sql:1-8`) has `quota_bytes`, `may_upload`, and `upload_concurrency_cap`.
  `member_preferences` (`:10-14`) has only `collapse_duplicates`, with no default-collection column. `member_quota`
  (`:16-20`) has `charged_bytes`, and `group_quota` (`:25-30`) is a singleton with `physical_total_bytes`.
- Store invariants (`internal/engine/store/invariant_test.go`): every table is STRICT with BLOB paths (invariant 1,
  `:219-233`), a table with a `kind` column carries NOT NULL `at` and `kind` (invariant 3, `:329-346`), and a
  non-exempt table's primary key is one BLOB `id` (invariant 6, `:397-416`).
- Writes go through `Handle.Update` (`internal/engine/store/write.go:64`), and ids come from `store.NewID`
  (`id.go:11`). `TakeKindLease` (`lease.go:187`) admits one job of a kind.
- The older-store helper's `migrationTables` maps versions 2 to 5 (`cmd/dubplate-server/olderstore_test.go:28-33`).
- No same-device check exists for staging and library roots. The `st.Dev` reads at HEAD are in the store's lock and
  restore code and `internal/backup/restore.go:563`.

Configuration:

- The wall rows are `DUBPLATE_LIBRARY_ROOT`, `DUBPLATE_STATE_DIR`, `DUBPLATE_SITE_NAME`, and `DUBPLATE_ADMIN_EMAIL`
  (`internal/config/keys.go:16-19`).
- The `intake` reserved class, rung 7, holds `DUBPLATE_UPLOADS_DIR` and `DUBPLATE_TRIGGER_DIR`
  (`keys.go:50-57`). Setting either is a load error naming the class (`internal/config/config.go:14-18`).
- No intake, quota, or upload key is registered. The registered `DUBPLATE_*` setting keys at HEAD cover auth,
  backup, branding, jobs, links, nonces, the operator CSS, serve, sessions, the store, and the theme.

Auth, identity, and serve:

- `RequireAdmin` (`internal/auth/admin.go:110-135`) verifies through `cfg.Seam.VerifyIdentity` and answers every
  refusal with `http.NotFound`. No member-level gate is exported. `MagicLinkSeam.VerifyIdentity` is at
  `internal/auth/seam.go:107`.
- `identity.UserID` is `[16]byte` (`internal/identity/identity.go:199`), and `Identity.Valid` is at `:228`.
- `Shell.Register(name string, h http.Handler)` (`internal/serve/serve.go:84-99`) panics on a duplicate, a late
  registration, or a widening name. `Shell.terminated` answers unclaimed paths with `http.NotFoundHandler`
  (`:248-258`).
- The origin check admits a state-changing request carrying `Sec-Fetch-Site` or `Origin` once
  `http.CrossOriginProtection` passes it (`internal/serve/origin.go:113-124`), so a same-origin tus PATCH passes.
  No bearer bypass is registered (`:44`).
- The saved styled-404 patch (`research/patches/2026-09-26-styled-404-fallback.patch`) no longer applies.
  `git apply --check` fails on `cmd/dubplate-server/web.go:169` and `internal/serve/serve.go:31`, and it targets
  `web/svelte.config.js`, which does not exist. It is reference material only.

Engine and mail:

- `tag.Album(ctx, src *musicbrainz.Client, items []matcher.ItemFields, tun, compat, evalDate, opts)` returns a
  `*Proposal` (`internal/engine/tag/tag.go:66-67`).
- The limiter's classes are `ClassUserPresent`, `ClassBatch`, and `ClassBackground`
  (`internal/engine/musicbrainz/limiter.go:16-21`). No non-test caller outside `limiter.go` names
  `ClassUserPresent`, so the probe would be its first.
- `notify.New(cfg Config) *Mailer` (`internal/notify/notify.go:76`) is the mail seam.

Deployment:

- The serve unit grants `ReadWritePaths=/var/lib/dubplate/auth /var/lib/dubplate/store`
  (`systemd/dubplate-server-serve.service:43`) under `ProtectSystem=strict`. `deploy.sh` names the same leaves in
  `DUBPLATE_SERVER_LEAVES=(auth store)` (`scripts/deploy.sh:282`). A staging leaf joins the unit, that array, and
  the state-dir leg together (CLAUDE.md, "Gate").
- Unverified, and owed before the staging location is chosen: whether `/var/lib/dubplate` and the box's library
  root `/srv/music/library` (STATUS:13) share a device. CH:472-475 requires it. A read-only check of the box's mounts
  answers it.

## 5. Questions for Geoff

1. **What shape should the uploader before/after sitting take?** Recommendation: one attended sitting at the end
   of rung 7's last pass, on the workstation (RP:143-144). Before it, the lane renders and reads every state, and
   W11.5's fresh-context coherence read passes. At the sitting, you drive a real upload. That means a FLAC album as
   files, the same album as a ZIP, one reject, and one push-back, at desktop and phone widths in both palettes. You
   rule VI's three unruled items (phone header, error alerts, list view) in the same sitting. The "before" is VI's
   committed mockups and the 5c pages as they stand, since a greenfield uploader has no earlier build to compare.
   Alternative: two shorter sittings, one after the uploader pass and one after the visual pass. That catches a
   wrong direction earlier at the cost of a second attended moment, and PA:140 counts this as one of two in the
   build.
2. **Does the unmatched lane survive into 0.8?** The 2026-09-01 addendum gave members an escape for albums
   MusicBrainz cannot place: a fenced collection with a minimal tag form and no matching (UD:251-258, PCF's "Private
   Recordings" copy). No greenfield spec carries it. "Private" now means visibility, which stays unsettable until
   slice 3 (SS:1004-1005), and the source question offers "My own recording" (PS:518), which strict custody always
   pushes back. Recommendation: keep it, renamed so it does not collide with private visibility. Rung 7 builds the
   form and the routing choice, and rung 9 admits without a match. Alternative: drop it at 0.8 and let rung 13's
   appeal action (OD:330-335) carry unmatched albums to the owner's queue.
3. **Do the album page, the library grid and list, filtering, and buy links move to a web pass after rung 9?** No
   album can be admitted before rung 9, and VI:361-365 says the filtering study maps each facet "to the rung 9
   column that supplies it". Recommendation: yes. Rung 7 ships the uploader as the signed-in member's landing page
   and the empty-library state with its photo, which is what every box shows until rung 9. A web pass after rung 9
   builds the rest. Ruling 7's own condition ("without growing the pass") lets buy-link rendering move with the album
   page, while the store link is still captured at upload (PS:763-764). Alternative: rung 7 builds those pages
   against fixture data, which adds about two web tasks and renders pages no member can reach before rung 9.
4. **What is a member's default storage quota?** CH:548-558 binds quotas at three levels in the settings table but
   sets no default. Recommendation: no per-member cap by default, a hard free-space floor always, and admins set caps
   later through rung 14's settings. A group of family or friends is the product's shape, and the floor is what keeps
   the box up. Alternative: a fixed default, such as a share of the disk, which needs a number from you.

Carried, not blocking: STATUS:30's Fork B question (whether anyone in the group wants to upload a non-FLAC file).
Rung 7's reject copy is written to the shape ROADMAP:133-137 sets, which names the supported set and suggests a
lossless conversion, so either answer leaves rung 7 unchanged.

## 6. Decisions the plan will take

These are method and architecture calls, each with its evidence.

- **D1. Adopt tusd v2.10.1 behind the shell, configured closed.** CORS disabled, downloads and concatenation off,
  the experimental IETF protocol off, `RespectForwardedHeaders` false, and `MaxSize` from a registered setting.
  Termination stays on, behind an owner check in `PreUploadTerminateCallback`, so a member can cancel. The plan
  rules the logger bridge (section 3), and it rules `memorylocker` against `filelocker` on whether more than one
  `serve` process can hold the staging directory.
- **D2. Decode in process with `mewkiz/flac`.** The gate decodes every track, checks frame CRCs, computes the audio
  key as the MD5 of the decoded samples, and streams a SHA-256 file hash. It compares the header MD5 only when the
  header's value is nonzero (CH:420-427). No subprocess runs (serve unit `:68`).
- **D3. A check that cannot run rejects the album** in reject wording (Melodee L4), with a fixture proving the
  faulted path rejects and the clean path admits.
- **D4. Accept ZIP uploads.** PCF:39-41 ratified "a .zip or its files", and CH:722's "decompressed total" cap
  presumes archives. The verdict reads the central directory without extracting a member, refuses absolute,
  symlinked, and escaping member paths, and caps the decompressed total before any byte is written.
- **D5. Split rung 7 into three passes** (section 7). BS:684-685 lets an M rung that outgrows one pass split at its
  own planning, and the routed items above roughly triple the ladder entry.
- **D6. The user-present match probe is rung 7's.** The journey's `matching` and `pushed-back` states are rung 7's
  vocabulary, rung 13 amends "rung 7's push-back page", and the probe is `ClassUserPresent`'s first caller. The
  probe stores the matcher's proposal on the submission, so rung 9 admits from it and never re-matches.
- **D7. Rung 7 defines only the journey states it emits:** uploading, checking, matching, pushed-back, accepted
  (waiting for admission), and stalled. Rung 9 adds its states additively under 14a's enum rule. That follows the
  2026-09-12 refinement ("nothing is exported ahead of a caller"). Journey records form an event table with `at`
  and `kind` (hook 4).
- **D8. The hook ledger is the table in section 1.** Hook 1 lands on rung 7's submission-file rows, and rung 9
  copies the keys onto `tracks`. Hook 12 is routed to rung 9 with an SS erratum.
- **D9. The default collection, and no template machinery.** Rung 7 creates the default collection,
  server-written, admin-owned, and group-visible, with its root allocated from its id, the no-nesting prefix check,
  and the Go path validator (R6:181, R6:185-186, R6:567, CH:322-325). It leaves `layout_template` empty. The
  template grammar, sanitizer, and injectivity check wait for rung 9's layout write. The default-collection
  preference and the per-upload override wait for the first verb that can make a second collection (14b), because
  W7.19's picker hides at one collection.
- **D10. Quota by reservation.** The reservation is taken at tus creation from `Upload-Length`, trued up at
  completion, and released on reject or abandonment. The free-space floor is checked before every write
  (CH:554-558). The `may_upload` and `upload_concurrency_cap` overrides apply at creation.
- **D11. One enumerable reject set.** The authority's strings are adopted verbatim (BS:508-511) as an enumerable set
  with a canonical wording per parameterized reason, which serves the site generator (project-site spec `:301-306`).
  New reasons are drafted to PCF's rules: the source answer missing (A9), the album already present (CH:435), quota,
  each cap, a check that could not run, and the format reason in ROADMAP:133-137's shape. A `prose-voice-reviewer`
  read gates them, as it gated 5c-iv's deck (HISTORY, 5c-iv-E2). Uploader sessions get no reject email
  (UD:261-262).
- **D12. Mounts.** The tus endpoint is its own subtree, registered from a rung 7 file. The JSON endpoints go under
  `/api/v1` if 14a's mount has merged, and otherwise under their own subtree, which 14a's contract adopts.
- **D13. A member gate beside `RequireAdmin`,** keeping its indistinguishable-refusal discipline for pages that are
  member-only. This waits on the `internal/auth` ownership ruling in section 2.
- **D14. A `doctor` row for the same-device rule** over staging, the library root, and every collection root. The
  staging leaf's location follows the read-only mount check in section 4.
- **D15. Pin `tus-js-client` 4.3.1 exactly;** BACKLOG #8 now watches 5.0.0-pre3.
- **D16. Retire the `intake` reserved class** in the commit that registers rung 7's keys, moving its two names to
  superseded rows.
- **D17. Defaults by research.** The allowlist ships FLAC-only within the tagio bound (BS:802-811), with the
  household-format research BS:329-331 asks for recorded in the plan. The CD-quality floor (44.1 kHz, 16-bit, PCF:39
  and the salvage constants at `_salvage/internal/serve/upload.go:48-51`) becomes a registered threshold. Each cap's
  default cites its basis.

## 7. Proposed segmentation, ceiling, and clock (under the recommended answers)

### 7a, the intake core

| Task | Outcome | Pass class |
|---|---|---|
| 0 | Dependency sweep (or one line if 14a ran it); tusd and `mewkiz/flac` adopted with the consult recorded | `tool` |
| 1 | `internal/reject`: the enumerable set, verbatim strings, new reasons through the register read | `engine-logic` |
| 2 | Migration: submissions, submission files with audio key and file hash, journey events; the older-store map row; R-6d4's keep limit | `auth-data` |
| 3 | The default collection: root allocation, the no-nesting check, the Go path validator | `auth-data` |
| 4 | Member gate, the tus mount (D1), the staging leaf (unit, `deploy.sh`, state-dir leg), the `intake` class retired | `auth-data` |
| 5 | Quota reservation and true-up, the free-space floor, the upload overrides | `auth-data` |
| 6 | Gate policy: allowlist, CD floor, caps, the ZIP walk, unsafe paths, the non-UTF-8 rule, could-not-run rejection | `engine-logic` |
| 7 | Decode integrity, audio key, file hash, album-keyed duplicate refusal over visible collections | `engine-logic` |
| 8 | Submission API: source answer (A9), store link, `added_by` (A10), journey read; the route inventory if first; the `doctor` same-device row | `auth-data` |

Segments: A (0 to 3), B (4 to 6), C (7 and 8). The close runs the simplifier, one `go-architecture-reader` per touched
package (`reject`, `intake`, `store`, `auth`, `serve`, `config`, `cmd/dubplate-server`), and a
`web-auth-security-reviewer` read after B and after C.

### 7b, matching and the uploader

| Task | Outcome | Pass class |
|---|---|---|
| 1 | The user-present probe: `tag.Album` under `ClassUserPresent`, the proposal stored, strict push-back, the journey states | `engine-logic` |
| 2 | The unmatched lane's form and routing (question 2), if kept | `engine-logic` |
| 3 | The uploader page: W7.19's four surfaces, `tus-js-client`, the source question with per-album override | `engine-logic` (web) |
| 4 | The journey and push-back pages with fix paths, on polling (UD:135-137) | `engine-logic` (web) |
| 5 | The signed-in landing: the Continue link, the header wordmark target, the empty-library state | `engine-logic` (web) |
| 6 | Render and read every new state at five widths in both palettes, then the fold | `engine-logic` (web) |

### 7c, visual authoring and the sitting

| Task | Outcome | Pass class |
|---|---|---|
| 1 | The photo loader leaf package, the Go port of `check.mjs`'s two modes, the CREDITS generator with pack lines | `tool` |
| 2 | The photo column on login, error-page photos, the empty-library photo | `engine-logic` (web) |
| 3 | The styled 404 with the admin gate's refusal through the same handler, and the error page's onward link | `auth-data` |
| 4 | 5c's carries: button text fix, latent items, toggle, masked sign-in line, gutter clamp, favicon paths, W12.4, W4.10, placeholder opacity | `engine-logic` (web) |
| 5 | VI's items that need no catalog: phone header, error alerts, motion, the sleeve set | `engine-logic` (web) |
| 6 | Render and read, W11.5's coherence read, the fold | `engine-logic` (web) |
| 7 | Geoff's before/after sitting and the fold of its rulings | attended |

7b and 7c close with `svelte-reviewer` and `daisyui-a11y-reviewer` against VI, a `web-auth-security-reviewer` read
over 7c's Task 3, and architecture reads of the touched Go packages.

### Ceiling

The derivation uses recent spend. 2b spent about 7.4M over twelve tasks, close included, about 0.62M a task. 6d's
chain ran about 0.58M a task inside the chain, with about 4.3M of overhead outside it on a 9.5M pass (the 14 brief,
section 6, from HISTORY 2b and 6d "Budgets"). Web passes run hotter at the fold. 5c-iv's segment E spent about 6.0M,
of which the render-and-read folds R1, R2, R3, and R3b took about 2.34M, and HISTORY records the cause as "pass
growth: the plan's single Task 10 fold became 10b, 10c, R1, R2, R3, and R3b" (HISTORY:1531-1543).

- 7a: nine tasks at 0.6M come to about 5.4M. Overhead is about 2.6M: an architecture fold near half of 6d's 2.59M,
  seven architecture reads at about 0.65M, two security reads at about 0.3M, pre-flights at about 0.2M, and the
  simplifier. That totals about 8.0M. **Proposed ceiling 8.5M, flag 6.8M.**
- 7b: six tasks at 0.6M come to about 3.6M, plus a 1.5M fold reserve for Task 6 and about 1.0M of reviewers,
  architecture reads, and the simplifier, for about 6.1M. **Proposed ceiling 6.5M, flag 5.2M.**
- 7c: six build tasks at 0.6M come to about 3.6M, plus a 2.0M fold reserve on 5c-iv's evidence, about 0.4M for
  W11.5's k=3 read, and about 0.9M of reviewers and close, for about 6.9M. **Proposed ceiling 7.5M, flag 6.0M.**

Rung 7 totals about 22.5M across the three ceilings. Planning spend (one rung 7 spec, three plans, and their
`spec-plan-review` rounds) sits outside, as 6d's did. Question 3 answered "build them now" adds about 1.5M to 7c.

### Clock

2b's overnight run gives about 20 minutes per task inside the chain and about 9 minutes per boundary full gate
(HISTORY 2b, "Budgets"). A close with its reviewer reads took about 1.5 hours.

- 7a: nine tasks (about 3 hours), three boundary gates (about 27 minutes), and a close (about 1.5 hours) come to
  **about 5 hours.** `auth-data` fix rounds of the kind 6d saw push it toward 6.5.
- 7b: six tasks (about 2 hours), two gates, about 1 hour of render-and-read fold, and a close come to **about 5
  hours.**
- 7c: six build tasks (about 2 hours), two gates, about 1.5 hours of folds, the coherence read (about 30 minutes),
  and a close come to **about 6 hours,** excluding the sitting itself.

That is about 16 hours of execution for the rung, excluding owner pauses and the planning sessions. 7a can run
beside 14a once section 2's two ownership rulings are written. 7b and 7c are serial behind 7a, and 7c's sitting is the
rung's single attended moment.
