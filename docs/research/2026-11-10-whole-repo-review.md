# Whole-repo review, 2026-11-10

Seven fresh-context reviewers read the repo at `36e84ee`: six over code slices and one over the Claude
infrastructure, with a plan read beside them. This file holds their findings, grouped by the pass that takes each.
BACKLOG items #104 to #127 point here by finding id. Each finding was read from source and grepped for callers;
none was proven by a test run. The planner of each pass re-reads its findings against the tree before dispatch.

Line numbers are as of `36e84ee`. Paths are relative to the repo root.

## gate-1

Runner class: `scripts/`, `Makefile`, `web/scripts/`, `.github/`, `.claude/`, `systemd/`, `cloud-init/`.

- **G1. Per-leg timing.** `check.sh` records no elapsed time per leg. The gate took 36% of `admin-api-1`'s wall
  clock (10,224 s over 20 runs) and 32% of `intake-1`'s (10,826 s over 17 runs). Print one elapsed line per leg
  first; G2 to G6 are ordered by reading and need the measurement.
- **G2. The rendered matrix.** `web/scripts/rendered/run.ts:6091-6146` runs 190 cells in sequence, each with a
  roster reset, a new browser context, and a full arrange. `rendered_test.go:37-40` records about 257 s. Arrange
  once per state and scheme, then call `setViewportSize` across the five widths: 38 contexts. The Go manifest
  counts link requests per cell (`rendered_test.go:271-275`) and changes with it. Stale counts: `run.ts:5317`,
  `:5536`, `:6080-6082`, and `rendered_test.go:39` say 15 states and 150 captures; the matrix is 19 and 190.
- **G3. gitleaks walks ignored trees.** The full form scans `-s .` (`check.sh:735`). Add `web/node_modules`,
  `web/rendered-captures`, and `tools/venv/.venv` to `.gitleaks.toml`'s allowlist. The scoped form spawns one
  gitleaks process per changed file (`check.sh:569-576`).
- **G4. A scoped run always builds the web app.** `check.sh:581-589` runs two Vite builds, `svelte-check`, and
  eleven static gates when no `web/` path changed. Split the Makefile into `web-build` and `web`, and call
  `web-build` when `web_changed=0`.
- **G5. Lane choice.** `gate.sh` unsets `CAIRN_GATE_LANE` for every run, so a browserless scoped run queues on
  the heavy lane (BACKLOG #34). Choose the lane from the scope line. Measure golangci-lint against the light
  lane's memory cap first.
- **G6. Vale under scope.** `vale-comments.sh:53-59` reads every Go file on every form. Take the change set under
  `--scope-changed`.
- **G7. PR CI** (BACKLOG #97). A `pull_request` workflow running `check.sh`. The runner needs xmllint 2.15.4, the
  Unidecode wheel, the XSD, rclone, Chromium, and `systemd-analyze`. With CI, the close stops running `--merge`
  twice.
- **G8. `gate.sh:21-43` rejects `--reduced`**, which `check.sh` documents and plans call.
- **G9. `deploy.sh:257,338`.** `rsync -a` to `root@` keeps the sender's numeric uid, so the unit and the binary
  land owned by uid 1000. Add `--chown=root:root`. Not confirmed on the box. The check-mode rsync (`:255`, `:336`)
  compares size and mtime; add `-c`.
- **G10. `provision.sh:98-111`** falls back to `alma-9`, but `cloud-init/user-data.yaml:165-166` hard-codes el-10
  URLs under `set -e`. Fail instead. `provision.sh` applies by default while `deploy.sh` checks by default; align
  on `--apply`. `provision.sh:7` wrongly says it is the only hcloud caller (`deploy.sh:96`).
- **G11. Duplicate work in `check.sh` and the Makefile.** Identical file-list pipelines (`check.sh:782-785`,
  `:823-826`); `classify_path:378-386` re-implements `nearest_go_pkg_dir:422-433`; six `command -v` blocks
  (`:218-241`); a 20-line shebang sniffer (`:663-682`); the dead `scripts/checks/*` pattern (`:366`); `make vet`
  duplicates golangci-lint's `govet`; `web-rendered` runs its own tagged vet and lint (`Makefile:139-140`).
- **G12. The exported-unused leg** is 55 lines of awk (`check.sh:812-879`), name-based and warn-only.
  `golang.org/x/tools/cmd/deadcode` answers for functions. Both staged legs print up to twenty warnings a run.
- **G13. Hardening.** `deploy.sh:377-389` copies the serve unit's hardening by hand; move it to a oneshot unit
  under `systemd/`. Add `systemd-analyze security --offline=yes --threshold=N` as a leg. The serve unit lacks
  `PrivateDevices`, `ProtectKernelLogs`, `ProtectClock`, `ProtectHostname`, `ProtectProc=invisible`,
  `ProcSubset=pid`, `RemoveIPC`, and `SystemCallFilter=~@privileged @resources`. Its `:74-75` says serve runs no
  subprocess, which the streamer's ffmpeg will change.
- **G14. The dependency sweep leaves the pass branch.** A sweep task writes `go.mod` or `package.json`, which
  `check.sh` classes as gate machinery, so later scoped runs go full unless `--base` is passed. Both recent plans
  departed from the adapter's "Pass no `--base`". Run the sweep on master before the branch is cut, and make
  `--base <task start>` the adapter rule.
- **G15. Plan conventions move to `dubplate-pass`.** 18 of 19 "Conventions every task follows" headings are
  identical in the `admin-api-1` and `intake-1` plans, and `intake-1`'s prompt pasted 48 KB into each of 11
  dispatches. Plans cite the adapter; dispatches get a path and line ranges.
- **G16. STATUS carries no fact git knows.** It said the `intake-1` PR was open for four weeks after the merge.
  Drop the head, PR state, and session ids from its contract.
- **G17. Close cleans up.** No step removes a merged pass's worktree or branch. Add `git worktree remove`, the
  local and remote branch delete, and a prune.
- **G18. Allow the mandated commands.** `.claude/settings.json` allows bare `check.sh` and not
  `bash scripts/gate.sh`. Add `gate.sh*`, `npm ci --prefix web`, `make *`, `gofmt *`, and read-only `gh pr` forms.
- **G19. `docs/plans/*-prompt.txt`** still call `review-desk`, which is not on `PATH`.

## Workstation process

Outside this repo: `~/.dotfiles` and the standards plugin, through `dotfiles-pass`.

- **X1. Nested gate.** A nested `cairn-run-gate` call deadlocked for 70 minutes (PR #10). Export a marker and
  refuse a nested call. `pass-core` "Gates" says to wrap the gate, and dubplate's gate already wraps itself; it
  should say "run the adapter's gate command as written".
- **X2. Estimates.** Actual clock ran 1.45 times the design figure in both feature passes, and plan figures ran
  1.3 to 1.7 times actual. One inflated figure crossed the 2x ceiling and cost an overnight wait. Set the plan
  figure at design times 1.5, and let a fold raise it only with a named added task.
- **X3. Review rounds.** `admin-api-1` took three rounds (4 spec lenses, 3 plan lenses, 3 on the re-plan). Review
  a re-plan with the mechanics lens only, and commit the fold record without the lens files.
- **X4. `pass-core` is 36 KB.** An executor loads its conductor, Proton Pass, hosts, and release sections. Split
  it into executor, planner, and conductor references.
- **X5. Rule copies.** The one-executor rule has three copies with three `pgrep` patterns, and the global one
  matches itself. The dependency rule has four copies.
- **X6. Tokens in the score.** Feature-pass transcripts are purged within the month, so record tokens in
  `pass-score` at close.

## cleanup-engine

`internal/engine/{matcher,musicbrainz,tagio,tag}` and `internal/harness`. Runs in the main tree for
`make derive`.

- **E1. tagio's 69 fields are stated five times**: constants (`tagset.go:15-85`), struct (`:128-198`), mapping
  (`:256-326`), and two switches ending in `panic` (`encode.go:55-186`, `read.go:87-224`). Put a typed accessor
  on each mapping row so `Serialize` and the projection walk one table. Fold the six scalar encoders
  (`encode.go:235-317`) and five readers (`read.go:296-350`) into one generic each. About 300 source lines.
- **E2. The MusicBrainz limiter has no caller.** `NewLimiter`, the `Class` constants, `Limiter.Budget`,
  `ErrBudgetExhausted`, `ErrLimiterClosed`, `Budget.Spent`, and `DefaultInterval` appear only in `limiter.go` and
  its test; every client uses `Unmetered()`. 390 lines and a 423-line test. Delete until a rung meters. It also
  has a spacing bug: an idle tick is banked as `spare` (`limiter.go:213-216,289-292`), so two requests can leave
  arbitrarily close together.
- **E3. One matcher component per file.** 20 files hold 3-to-10-line functions behind a seven-parameter
  signature; `Name` equals `Key` in all 21 rows and only tests read either (`score.go:12-16`, `track.go:9-13`).
  Write `DistanceForAlbum` and `DistanceForTrack` as straight-line calls in two files. Two tests restate the
  slice (`score_test.go:20`, `track_test.go:103`).
- **E4. `Tunables` is stringly typed.** Of 25 fields, non-test code outside the matcher assigns seven, all in
  `harness/conform`. Names resolve in four places (`tunables.go:448-493`, `verdict.go:162`, `:179-210`,
  `candidate.go:121-125`). Type `MaxRec` as `map[string]Recommendation` and the actions as `Action`, and parse
  where strings enter. `compilePanicAsError` (`tunables.go:504-512`) recovers a panic as control flow, and both
  pattern lists recompile per candidate (`album_media.go:41-51`, `album_country.go:36-46`).
- **E5. tagio's registry dispatches over one format** (`seam.go:33-56,112-122`). Keep `Format`, `FormatFLAC`,
  `SupportedFormats`, and `Sniff` as plain functions and delete the map. `Read` opens the file twice.
- **E6. Exports with no outside caller.** matcher: `Distance.Update`, `Levenshtein`, `Transliterate`,
  `AssignItems`, `AlbumPairsFor`, `ParseAction`, `ParseRecommendation`, `LegalQuietFallbacks`,
  `LegalNoneRecActions`, `SDEndWords`, `SDReplace`, `SDReplacement`; `ProjectItemFields` and its helpers
  (`info.go:154-259`) belong in `harness/conform`. tagio: `WriteOrder`, `Mapping`, `EmptyString`, `EmptyList`,
  `FrontCover`, `Deviations`, `FieldMapping.Citation`. musicbrainz: `ClientConfig.Timeout`. tag:
  `Options.SearchIDs`.
- **E7. strdist's regex cache** (`strdist.go:146-193`). Hold `*regexp.Regexp` in the row; `SDReplace` is
  `strings.ReplaceAll(s, "&", "and")`.
- **E8. musicbrainz leftovers.** `GroupRelations` (`relations.go:27-38`) serves one read at `pseudo.go:25`.
  `parseDate` (`events.go:59`) returns a month and day nobody reads. `preferredAlias` rebuilds its map per call
  (`tracks.go:70-73`). A cancelled context is retried once (`client.go:353-363`). `io.ReadAll` on the response
  has no cap (`client.go:365`).
- **E9. unidecode.** The checksum block appears four times (`fetch.go:48-52`, `extract.go:22-26,37-41`,
  `table.go:30-34`). `decode` stores four values nothing reads.
- **E10. Harness trims.** `conform/request.go`'s sweep (`:72-190`) is subsumed by tier 1 (`case.go:415`,
  `:505-514`); confirm by mutating `inc=` first. `Case.Extra` and `tiesAwaitingRuling` have no data and no
  writer. `fieldNullability` (`translate.go:70`) is read only by its test. Canonicalization and release-lookup
  parsing each exist three times (`cassette.go:260`, `stub.go:240`, `mbrequest.go:34`); export from `cassette`
  and delete `mbrequest`. `oracle` repeats its scaffold three times. `arch_test.go` (288 lines) holds server and
  engine rules inside `conform`. `crypto/rand.Read` no longer errors (`stack.go:608`, `client.go:107`).
- **E11. Latent harness bug.** Default-mode grading never reads `row.DefectFree.PatternTable`
  (`conform/case.go:582`, `:649-665`). `cmd/harness/main.go:39-41` claims a contact check only `capture` makes.
- **E12. Tests.** `TestWriteOrderDeterministic` (`tagset_test.go:350`) compares a walk with itself.
  `TestDefaultTunables_*` re-type literals (`tunables_test.go:14-160,231`). Two tests check `slices.Clone`.
  `tagio/conform_test.go:347-490` is a staging pipeline where a diff would do. Three weak harness tests:
  `defaults_test.go:156`, `:185`, `mbrequest_test.go:16`.
- **E13. Comments.** matcher is 44% comment lines and musicbrainz 42%. About 250 lines cite pass or ruling ids;
  `conform` carries 50 more. `quiet.go:14-16` cites line numbers that have moved.

## cleanup-store

`internal/engine/store`, `internal/backup`, `internal/settings`, `internal/config`, `internal/atomicfile`.

- **S1. Bug: `--from-file` restore durability.** `StageFile` (`store/restore.go:328-356`) never fsyncs
  `staging/`, and nothing syncs it before `Record` names the stage. The fetch path does (`backup/fetch.go:129-151`).
  A power cut between `Record` and `SwapIn` can lose the directory entry, and every open then refuses. S3 closes
  it.
- **S2. Two packages prune one directory** (`backup/backup.go:174`, `backup/prune.go:24,163-194`,
  `store/snapshot.go:249-254`, `store/migrate.go:135-139`, `store/open.go:99-102`). Add
  `Handle.PrunePreMigrateCopies(before time.Time)` and drop `IsPreMigrateCopy`, `ServeOptions.Backup`, and
  backup's path arithmetic.
- **S3. Sealing a stage is split across packages** (`backup/fetch.go:120,129-151`,
  `store/restore.go:393-407,447`). Give the session one seal step.
- **S4. A restore's intent is modelled twice** (`store/intent.go:35-53,86-109`,
  `store/restore.go:578-593,723-750,764-774`). Make `RestoreIntent` a field of the record. Free while nothing is
  live, since it changes the record's JSON.
- **S5. `Inspect` doubles as a domain reader** (`store/inspect.go:134-168,193-235,408-491`,
  `settings/evaluate.go:27,132-141,169`). `store.SettingsRow` equals `settings.Row`. `Facts.SettingsGeneration`
  has no reader. Cheap form: alias the row and delete the field.
- **S6. Lock acquisition has four entry points** (`store/lock.go:38-78`, `store/open.go:386-418`). Keep `tryLock`
  and one bounded wait. `waitLock` reports the wait even when the parent context was cancelled (`open.go:412`).
- **S7. `atomicfile.WriteFile` is `WriteStream` without the directory fsync** (`atomicfile.go:16-43,55-91`).
  Implement one on the other. fsync-a-directory is written three times.
- **S8. Per-kind parsing five times** (`config/registry.go:287-308`, `config/settings.go:131-147`,
  `config/thresholds.go:64-136`, `settings/evaluate.go:360-373`, `settings/accessor.go:43-62`), and environment
  lookup three (`config/config.go:196`, `settings/evaluate.go:347`, `cmd/dubplate-server/doctor.go:684`). One
  parse method and one exported `config.Lookup`.
- **S9. `Threshold` gains `Floor` and `Ceiling`**, as `Setting` has. `cmd` hand-checks bounds today
  (`store.go:34-53,133-180`, `serve.go:98-126`); `cleanup-ops` removes those.
- **S10. A typed `config.KeyError{Key, Reason}`** from `Load` and the parsers. `doctor` recovers keys with
  `strings.Cut(line, ": ")` today (`doctor.go:475-496,700-706`); `cleanup-ops` removes that.
- **S11. Exports with no production caller**: `config.OwnedWall`, `config.ReservedClasses`,
  `Registry.Superseded`. `store.LeaseState` has three values and its caller tests one.
- **S12. `refusal(StateRefused, "…%v", err)` at 46 sites** drops the cause from the chain. Use `%w` and keep
  `refusal` for the other states.
- **S13. Tests.** Four tests of `slices.Clone`; `thresholds_test.go:37` duplicates `registry_test.go:83-89`;
  `registry_test.go:31-41` compares the registry to its seed tables; `schema_test.go:343-361` fails only on a
  2^-64 coincidence.
- **S14. Comments in `config`**: ten `// ruled, conductor` trailers in `registry.go`, five more in `keys.go` and
  `config.go`, 19 lines about deleted rows (`keys_test.go:40-58`), and "arrives at rung %d" in an
  adopter-facing error (`config.go:322`).

## web-shell-1

`web/src`.

- **W1. The admin page shell.** `routes/members/+page.svelte:78-96,135-178,297-313,418-430` and
  `routes/server-settings/+page.svelte:89-102,113-174,269-285,293-305` share a status machine, a fetch, a load,
  three gate-state branches, and a reauth alert. Extract `loadList(operation, field)`, a signed-in shell or an
  `(app)/+layout.svelte` group, and a write notice. Rendered rows select by id, so ids pass through.
- **W2. One transport core.** `lib/api/post.ts:51-76` and `v1.ts:43-66` each fetch, catch, and classify, with two
  vocabularies and two resolver sets (`outcome.ts:38-57`, `v1.ts:69-96`). `post.ts:64` treats every 403 as
  cross-origin and `v1.ts:59` only a non-JSON 403; keep each. A network failure is `alert-error` on login and
  `alert-warning` on the admin pages (`outcome.ts:19-31`, `v1.ts:95-96`); pick one.
- **W3. `btn btn-sm min-h-11`** at eight sites fights the size model. The identity spec (`:249-250`) takes touch
  targets from `--size-field`.
- **W4. Page scaffolding** (`<main>`, the `h1`, `siteName`, the head title) repeats in five route files.
- **W5. Component-layer rules no page uses** (`components.css:40-62,180-209,374-428`). The same file defers tab
  and list-row rules until a page renders them (`:430-435`); apply one rule to both.
- **W6. State nits.** `login/+page.svelte:33` is a `$derived` over constants. `SiteHeader.svelte:46` holds
  `$state` read once. `members/+page.svelte:109-111,124-128` keeps three `bind:this` refs for fields that have
  ids. Two `$effect`s (`members:120`, `login:56`) belong in their handlers.
- **W7. Tokens.** `opacity-70` at four sites (BACKLOG #19) and `text-[13px]` at two.
- **W8. Dead code.** `deck.ts:84` `SENT_MOMENT_TEXT`; `outcomeClass` and `refusalText` are exported for no
  outside reader.

## cleanup-ops

`internal/{ops,auth,serve,identity,notify}` and `cmd/dubplate-server`. `auth-data`. Follows `cleanup-store`.

- **O1. One handler for `opshttp`'s admit path** (`mount.go:61-80`, `carrier.go:1-43`, `authorize.go:20-230`,
  `opshttp.go:118-127`). The mount verifies, stores the actor in the context, and the handler calls a decorator
  that discards a zero actor and reads the real one back through an interface with one implementation. Write one
  `ServeHTTP` that verifies, routes, checks privilege, decodes, and calls. About 200 of 375 lines, plus 346 test
  lines. It also fixes the order: privilege is checked after the body decode today (`opshttp.go:121-124`).
- **O2. Job operations and the catalog have no caller.** `api.go:35` always passes `ops.NewUnavailableJobs`.
  Delete `ops/unavailable.go`, four hand-written handlers (`opshttp.go:153-187,224-234`), and `Catalog` with its
  helpers (`ops/catalog.go:48-61,117-152`) until a page starts a job (BACKLOG #52). `opshttp/contract_test.go`
  (1,020 lines) drives a wiring production never builds. The operations are in the ratified `openapi.yaml`.
- **O3. Surface ahead of a caller.** `auth/epoch.go:72,98-138,202-215` (`Subscribe`); `auth/seam.go:35-38`
  (`Seam.Read`); `auth/session.go:219` (`VerifySession`); `ops/errors.go:44` (`ErrUnauthenticated`). The bearer
  bypass (`serve/origin.go:54-74`) stays by BACKLOG #54. Unexport what only its package uses: `auth.RecordID`,
  `NoncePath`, `EpochPath`, `MintLink`, `VerifyLink`, `LinkURL`, `MintSession`, `RefreshSession`,
  `AllowConsumption`, `AllowLinkMail`; `serve.Normalize`, `RewriteDocument`, `PeerAddress`, `Listen`;
  `identity.MaxAddressLength`; notify's four sentinels.
- **O4. The ops error vocabulary** (`ops/errors.go:16-62,94-138`) keeps nine codes, nine sentinels, nine
  constructors, and a map. Production matches two sentinels. Make `Code` the `errors.Is` target.
- **O5. Two `Actor` types** (`ops/ops.go:25-80`, `auth/roster.go:152-215`). `MemberActor`'s `clientIP` is always
  empty, so web roster audits log no address. An operator is `local:<user>` in `jobs.actor` and bare `<user>` in
  the roster and settings audits.
- **O6. Bug: `AddMember` re-reads after commit** (`ops/admin/members.go:39-42`). A failed read answers 500 for a
  committed add, and the retry gets `duplicate_address`. Have `Roster.AddMember` return the written member.
- **O7. `statusCapture`** (`serve/serve.go:260-318,363-371`) wraps every request; `logRefusal` can log the status
  it already has.
- **O8. `RequireAdmin` and `CheckRequest`** are one function twice (`auth/admin.go:41-66`,
  `auth/request.go:31-58`).
- **O9. Hand-rolled stdlib.** `serve/clientaddr.go:259-292`; `opshttp.go:203-218`; `auth/handlers.go:527-547`;
  `serve/serve.go:240-258`; `auth/roster.go:235-240`; `auth/store.go:119-144`.
- **O10. Move the web surface out of package main.** `cmd/dubplate-server/web.go:228-860` is handler logic, and
  `serve/{document,csp,branding}.go` (780 lines) is called only from it. Give them one package beside `serve`.
- **O11. One cobra wiring shape.** 15 constructors take `(environ, reg, stdout, stderr)`, and 12 commands repeat
  the root's `SilenceUsage`. Verbs write through `cmd.OutOrStdout()`; tests run the root with `SetArgs`.
- **O12. Probes move to what they probe.** The boundary-probe client (`doctorprobe.go:181-445`) goes beside
  `serve.NewProbeRoute`; the remote probe (`backup.go:402-579`) to `internal/backup`; the table install
  (`unidecode.go:146-296`) to `matcher/unidecode`. `unidecode.go:260-296` copies `store/leaf.go:101-141`.
- **O13. Consume S9 and S10**: delete the hand-checked bounds and `doctor`'s string parsing. Six verbs re-resolve
  `storeBusyTimeout(environ)` after `loadConfig`.
- **O14. Smaller.** `newServeShellVia` and `registrar` exist for one test (`serve.go:306-338`, `web.go:196`).
  SIGHUP is handled in three places (`main.go:88`, `backup.go:296-319`, `restore.go:171`). `thresholdStatus` has
  four identical arms (`doctor.go:535-566`). Test-only seams sit in production types (`ops/local/local.go:78-82`,
  `notify/notify.go:60-71`, `auth/limit.go:265`).
- **O15. Tests.** Declaration mirrors (`auth_test.go:502`, `store_test.go:488-530`, `serve_test.go:127`,
  `intake_test.go:372`); tests that a finished migration stays finished; `templatekeys_test.go:194` tests a Go
  copy of `deploy.sh`'s renderer; `intake_test.go:360` is half-vacuous; literal pins in `ops/errors_test.go` and
  `catalog_test.go:121-197` repeat `opshttp/document_test.go:403`.
- **O16. Comments.** About 75 ids in `serve` and `auth` and 90 lines in `cmd`. `serve/clientaddr.go:1-3` is a
  second package comment. The `ops` tree has none and is the model. This pass flips the citation leg to gating.

## intake-2 additions

- **I1. The gate parses each album file twice.** `gate/file.go:167` calls `tagio.Read(path)` after
  `ParseBounded` at `:109`, a second sniff and an unbounded parse. `tagio.Tags(c)` exists for this.
- **I2. Prefix-duplicate names.** `gate/zip.go:137` and `gate/manifest.go:63` refuse exact duplicates only, so
  `a.flac` beside `a.flac/b.flac` passes.
- **I3. Intake collapses.** The role is derived twice (`submission.go:218`, `accept.go:62`) as an untyped string;
  the id-scan closure appears four times; `accept`'s album branch (`accept.go:138-150`) is unreachable;
  `reject.Symlink` has no producer. BACKLOG #87 and #88 hold the rest.

## The re-plan sitting

Rulings for Geoff, due before `pipeline` is planned.

- **P1. Rung 10 targets a retired estate.** `migration` builds a beets-database migrator and a path-parity
  fixture. beets is retired, and first launch already relocates every collection and re-lays multi-disc albums.
  No spec names how a Navidrome switcher's files enter the library. A verb that admits a folder through the
  pipeline would serve the box's 6.6 GB and a stranger alike.
- **P2. The first played track arrives late.** It needs all of `pipeline` and streamer slices 3 and 5, and
  `integration` is rung 15. Run slices 3 and 5 ahead of 4, 6, and 7, then deploy privately to the box.
- **P3. Scope candidates for after 0.8**: lifecycle hooks in `builtins`, the two non-default custody policies in
  `queues`, and smart playlists in slice 7.

## Leave alone

`matcher/assign.go:149-393` (the LAPJV port); `tagio/container.go:229-330` with `write.go:70-88,121-164`;
`store/write.go:69-227`; `store/swap.go:126-235`; `ops/local/jobs.go:149-236`; `auth/token.go:236-335`;
`intake/gate/zipcount.go`; `intake/submission.go:281-315`; the `subsonic/*` and `opensubsonic/*` packages; the
web markup; and `themes/system.css`, whose duplication was ruled.
