# Plan: `structure-keys`

Initiative: dubplate-server
Round: structure
Pass: structure-keys (pass 2 of 3)

Design: `~/Projects/initiatives/dubplate-server/specs/2026-11-17-structure-design.md`: "Calls made here"
("Settings live in the package itself", "Nine registration lines", "The snapshot count is read from the
settings package"), the `structure-keys` section, "How it fails", "Proof that nothing changed", the acceptance
tables, the standing answers, and "Backlog verdicts" (#124, #138, #139, #156). Kit:
`/var/home/glw907/Projects/initiatives/dubplate-server/plans/2026-11-17-structure/` (`structure-proof.sh`,
`zz_registrydump_test.go`, `README.md`). Neither kit file is ever committed.

Class: `auth-data` (`cmd/dubplate-server`, `internal/auth`, `internal/serve`, `internal/webapp`, and
`internal/intake` are on the adapter's `auth-data` map). The adapter skill edit in T10 is `runner`. The pass
takes `auth-data`.

Worktree: `/var/home/glw907/Projects/dubplate/.claude/worktrees/structure-keys`, branch `structure-keys`, cut
from `origin/master` at `5d1e6e9` (PR #18, `structure-base`, merged), with `c5aa8a5` merged in at planning
(BACKLOG and STATUS only, no Go, so every base value below holds). PR base: `master`. Push target:
`glw907/dubplate`.

## Pre-flight

No step needs Geoff. The pass reads no secret, writes no Claude setting, adds no public surface, adds no
setting, and never reaches `5.78.148.42`.

- `npm ci --prefix web` in the worktree before the first gate. Check: `web/node_modules/.bin/svelte-check`
  exists.
- The kit folder is readable. Check: `ls` on the folder lists the two kit files.
- Port 18431 on 127.0.0.1 has no listener before each proof run. Check: `ss -ltn 'sport = :18431'` prints
  only its header.

Permissions checked: `~/.local/state/claude-denials.jsonl` was read at planning. Its dubplate rows refuse
`sed -i` with absolute paths in a pair agent's worktree, a Python in-place rewrite, an awk script run from a job
tmp folder, a `curl` download, and `git remote -v | sed`. This pass needs none of them. Every move is an Edit
or Write in the executor's own session, never a subagent's, and no pass step pipes `git remote`. The Bash
PreToolUse hook `claude-block-pass` refuses some commands that carry a bracket class (at planning it refused the
acceptance greps inline, and they ran from a script file). A refusal carrying the Proton Pass or
`pass-sitting` message, on a command that touches no Proton Pass, is rerun verbatim from a script under
`$CLAUDE_JOB_DIR/tmp` as `bash <script>`. Any other refusal is a stop, never routed round with another tool.

Reviewed: one `standards:plan-reviewer` read (opus, medium); see "Plan review fold".

Evidence: the mechanism table, the searches, and the base values below, run at `5d1e6e9`.

## Mechanism table

| # | Mechanism | Probe or source, and what it showed |
| --- | --- | --- |
| M1 | The registry keeps one slice per family | `internal/config/registry.go`, `type Registry struct`: `ownedWall`, `reservedClasses`, `integrationClasses`, `thresholds`, `identitySources`, `settings`, `correlatedRules`, and the `superseded` map, each appended by its own `Add*`. Order is only within a family, so moving one package's whole call earlier or later changes nothing unless another package adds to the same family between them. The kit's dump prints the seven slices in that order (`zz_registrydump_test.go` lines 33-39) and does not print `superseded`. |
| M2 | The nine-line order keeps every family's order | Families each package adds today: `serve` thresholds (bind addrs, port, prefix); `auth` a wall row, an identity source, a superseded name, ten settings, two rules; the store two thresholds and two settings (job retention, then snapshot keep); `webapp` four thresholds; `backup` one integration class and three settings; the trusted proxy one integration class; `api` one setting and one rule; `intake` a wall row, two superseded names, settings, and a threshold. With job retention moved to `ops/local` and registered before the store, the settings slice reads auth, job retention, snapshot keep, backup, reauth window, intake, as today; the thresholds slice reads serve, store, webapp, intake; integration classes read backup, then the proxy. The design's scratch probe agrees ("the registry prints exactly as it does today"). The dump after each task is the proof. |
| M3 | A duplicate key panics at registration | `registry.go:118` `claim(family, key)` panics naming both families. `serve` registering twice adds disjoint keys in two families, so no claim repeats. |
| M4 | `Load` runs an identity source's `Validate` and joins its error | `IdentitySource.Validate func(values map[string]string, environ []string) error`; `Load` calls it whenever the source is selected. `validateMagicLinkSource` (`cmd/dubplate-server/auth.go` ~352) joins, in order, the secret length, the public URL, the SMTP from address, the SMTP port, then `validateServePort(environ)`. The joined order is what `doctor` prints, so the port check keeps its place in the join. |
| M5 | The serve port check | `cmd/dubplate-server/serve.go:107` `validateServePort(environ)` calls `config.ParseInt(environ, servePortThreshold)` and refuses below 1024 with "%d is below %d: the serve unit withholds CAP_NET_BIND_SERVICE, so it cannot bind a privileged port" and above 65535 with "%d is above %d". Its one production caller is `validateMagicLinkSource`. It moves to `internal/serve/config.go` as an exported function; `internal/auth/config.go` calls it, so `auth` gains a production import of `serve`. |
| M6 | That edge closes a test cycle, which `go build` cannot see | `serve` has no internal production import, but `internal/serve/tusmount_test.go:86` builds `intake.Config{..., EarliestAdmin: auth.EarliestAdmin}`, so `serve`'s test build imports `auth`. With `auth` importing `serve`, `go vet ./...` and `go test ./internal/serve` fail with an import cycle. A scratch probe replaced `auth.EarliestAdmin` with a local stand-in, `func(context.Context, *store.Tx) (identity.UserID, bool, error) { return member, true, nil }`, dropped the `auth` import, and ran `go vet ./...` clean; `TestTusPatchOutlivesTheServerTimeouts` passed. `seedUpload` inserts the member as an admin, so the stand-in returns the row the real read would. The precedent is `testEarliestAdmin` at `internal/intake/collection_test.go:21` ("stands in for the roster's EarliestAdmin, which intake cannot import"). The design's cycle evidence was `go build ./...` only (design line 553); this plan makes `go vet ./...` a per-task check. |
| M7 | `auth`'s in-package tests import `serve` | `internal/auth/request_test.go`, `handlers_test.go`, `admin_test.go` call `serve.ClientIP` and `serve.TrustedProxyConfig`. These are `package auth` tests; with the production edge `auth`→`serve` they stay legal, since `serve`'s tests no longer import `auth` (M6). |
| M8 | Internal import edges today | Production: `serve` none; `auth` atomicfile, id, identity, store; `intake` id, identity, intake/gate, intake/staging, ops, reject, store; `backup` auth, store; `ops/local` backup, id, ops, store; `settings` config, id, identity, store; `ops/admin` auth, config, id, identity, ops, settings, store; `webapp` none; the store atomicfile, id. The edges this pass adds: `auth`→config, settings, serve; `backup`, `intake`, `ops/local`→config, settings; `serve`, `webapp`, the store→config. A scratch probe with every edge added ran `go vet ./...` with no cycle once M6's stand-in was in place. |
| M9 | The four layering tests with that edge set | `internal/harness/conform/arch_test.go` (four tests), `internal/intake/seam_test.go` `TestLayering` (`ops` and `opshttp` forbid `auth` and `settings`; `intake` forbids `auth`), `internal/opensubsonic/boundary_test.go` (`TestViolations`, `TestLayerImportsNothingInternal`), and `internal/webapp/layering_test.go` `TestLayering` (forbids auth, ops, settings, config, store). The probe ran all four: only `webapp`'s fails, on `config`, which T5 allows. `intake`'s forbids `auth`, so no `intake` file may import `auth`. |
| M10 | The settings package cannot be imported by the store | `internal/settings` imports `internal/store`, so the four snapshot-count functions (`localSnapshotKeep`, `seedFacts`, `byKey`, `olderStoreRefusal`, `cmd/dubplate-server/store.go:206-257,410-432`) move to `internal/settings` (design, "The snapshot count is read from the settings package"). The snapshot-keep declaration stays in the store's `config.go`. A probe built an external `settings_test` package that imports `ops/local`, the store, and config with no cycle. |
| M11 | `auth.SettingError` prints as the plain wrap | `internal/auth/handlers.go:85` `func (e *SettingError) Error() string { return e.Key + ": " + e.Err.Error() }`. Only `handlers.go:94` asserts the type, on errors from `auth`'s own readers. `intake` cannot import `auth` (M9), so the readers that move into `intake` (`intakeWindows` at `intake.go:369`, `freeSpaceFloor` at `intake.go:425`, and `intakeLimits` at `api.go:43`) wrap with `fmt.Errorf("%s: %w", key, err)`, which prints the same text, and return nil for a nil error. `newIntake` (`intake.go:343,347`) stays in cmd as wiring and calls `auth`'s exported form. |
| M12 | The parser the import test uses | `go doc go/parser.ImportsOnly`: "stop parsing after import declarations". `go doc go/parser.ParseFile` returns the `*ast.File` whose `Imports` field lists every import spec. The test walks each listed package's non-test `.go` files, so test files stay free. |
| M13 | The citation gate | `scripts/check.sh` (about lines 800-840) runs a wide citation pattern over non-test Go comments everywhere except `cmd/dubplate-server` and `internal/{serve,auth,intake}`. Comments that move into `backup`, the store, `webapp`, `ops/*`, or `settings` meet it. A planning scan of every comment in the moving blocks found each hit staying in `cmd` or moving into `auth` or `intake`; the "Ruling 3" lines in `backup.go` (145, 346) sit in command code that stays, and "Ruling 2" (399) and "ruling 14"/"6d" (408-410) sit in doctor code that `structure-doctor` moves. The gate's leg catches any miss. |
| M14 | `lookupEnviron` and `identitySourceValues` | `cmd/dubplate-server/doctor.go:673` `func lookupEnviron(environ []string, name string) (string, bool) { return config.Lookup(environ, name) }`. The leftovers grep finds it on 14 lines (the definition and its doc comment included) in `backup.go`, `doctor.go`, `settings.go`, `store.go`, and `unidecode.go:82,147`. `identitySourceValues` (`doctor.go:658`) builds a map over `src.Required` and `src.Optional` through `lookupEnviron` (`Required` is `[]config.KeyDecl`, so the names come from `config.KeyNames`, `keys.go:49`); its one caller is `doctor.go:603`, and no test calls it. The design calls it "a value list only a test reads"; the tree has one production caller (standing answer 19). |
| M15 | `Superseded()` has no production caller | `registry.go:508` `func (r *Registry) Superseded() map[string]string`. Its callers are three tests: `cmd/dubplate-server/intake_test.go:67`, `cmd/dubplate-server/templatekeys_test.go:95` (`classifyKey`'s "superseded" branch), and `internal/config/registry_test.go:70` (subtest `Superseded`). `AddSuperseded` stays. `Load` refuses a superseded name with its recorded message. |

## Searches

| Search (at `5d1e6e9`) | Hits | Disposition |
| --- | --- | --- |
| The declares check (acceptance row 1) | 52 | 49 declarations, two list openers (`auth.go:325`, `intake.go:288`), and the empty return at `store.go:431` in `byKey`. By file: api 2, auth 15, backup 4, intake 18, proxy 1, serve 3, store 5, web 4. All leave `cmd` in T1 to T8. |
| The registers check (row 2) | 29 | api 2, auth 6, backup 4, intake 5, proxy 1, serve 3, store 4, web 4. All leave `cmd` in T1 to T8. |
| `main.go` new and old forms (row 3) | 0 and 8 | `main.go:101-108`. The nine lines land one task at a time (see "main.go after each task"). |
| The leftovers check (row 6) | 19 lines | `backup.go:426,434,498,507`; `doctor.go:603,654,658,661,666,673,674`; `settings.go:62`; `store.go:275,298,470,488`; `unidecode.go:82,147`; `registry.go:508`. T9. |
| `go doc ./internal/config Registry \| grep -c '^func (r \*Registry)'` | 16 | 15 after T9. |
| Production files in the seven tested packages that import config or settings | 0 | After the pass, exactly each package's `config.go`, and `doctor.go` where one exists. |
| Retired rule text: `grep -rn 'keys.go\|newProductionRegistry' CLAUDE.md .claude/skills docs/standing-rules.md` | 2 | `CLAUDE.md:188` and `.claude/skills/dubplate-pass/SKILL.md:118-121`. Both rewritten in T10. `docs/standing-rules.md:126` ("its own file") is the dead-export rule and stays. `registry.go:304` ("The rung that builds a source declares it from its own file") is a doc comment T9 may refresh in its own words. |
| cmd test files that name a moving identifier | 11 | `auth_test.go`, `serve_test.go`, `store_test.go`, `intake_test.go`, `proxy_test.go`, `rendered_test.go`, `settings_test.go`, `check_test.go`, `apimount_test.go`, `doctorprobe_test.go`, `routes_test.go`. A test of a moved function moves with it into the package's `config_test.go`, keeping its name; a test that only calls a moved name takes the package qualifier. |
| Gitleaks allowlists naming a cmd file | 0 | `.gitleaksignore` holds no fingerprints and `.gitleaks.toml` names no cmd path. The four secret-name constants carry `// secret-guard-allow`, which moves with them. |

## Base values

The kit at `5d1e6e9`, built in a detached worktree under the planner's job folder:

```
empty environment:                 lines=78   sha256=5233cf78193226dec7e293c34b63b676691fb573bee263b2dc294f0667731ea3
complete environment, no store:    lines=77   sha256=d98236213a4019e7cf1c167a0128bfa4628f440c8d3d91b439ef58b6964a73fa
a retired setting name:            lines=74   sha256=29cec9c8b52a94066b6d209f144f417a9fd157e76ea6febfc678c916024113fc
refused values:                    lines=1867 sha256=e06b84f9a762f3baeab228ed42e3dc3a766641855cba5c1848429b858e445596
a store with one admin:            lines=183  sha256=968d704237a02f5d0fc090c06727f7283614d0f2892ea6ddeb264c760a1a20ab
help text, every command:          lines=225  sha256=c09227a87aebb58c7c933b530857ae8eff92e29122b69aac8d44575809d9c6a3
everything:                        lines=2510 sha256=7b996b4391ddca9ca2cfcb54a2de399ce88f22fa5a5d201018dfd92e346b4b06
registry dump: 56 lines, sha256 b8a59fba160781e6fd33c513d9a844266c20a9af401fd1e6f18e61e2b1365842
```

Migrations hash at `internal/store/migrations`: `ffc4e9dba22b...`. Test names: 1,906, sha256 `39f06a8c...a745`
(the design's 1,921 and `4918d4c4` were taken at `e03e574`; `structure-base` moved them, standing answer 19).
The executor reruns all of these at its launch base and records its own figures as "today".

## Standing answers this plan leans on

The design's standing answers hold. Cited by number: 13 (fix the code, never a recorded value), 14 (a move
that can't keep a setting name, default, refusal sentence, help text, or row order parks the pass), 18 (a
forced call site is taken and listed under `## Decisions`; anything else is filed), 19 (note where an earlier
pass moved a figure), and 21 (a reviewer's request to change behavior is filed, never done). Plans run against
the tree.

## Records and gates

- PR notes live at `$CLAUDE_JOB_DIR/tmp/pr-notes.md` (the `<git dir>` path was refused in an earlier pass) and
  are mirrored into the draft PR body after each task. First line `launch-base=<sha>`.
- Fast lane after each task: `bash scripts/gate.sh --scope-changed`, with no `--base`. This pass moves no
  `go.mod`, `go.sum`, or `web/package*.json`, so the scoped form scopes. The close run is
  `bash scripts/gate.sh --merge`. Neither is piped; the Bash tool's `timeout: 600000`, re-issued on exit 75.
- After every task, also run `go vet ./...` over the whole module (M6: a test cycle passes `go build`).
- CI watch after each push: `gh pr checks <n> --watch --fail-fast` as a background Bash task, bare. A red CI
  stops the line like a red gate.
- Usage stop (standing answer 17): 95 percent until the reset on 2026-11-20 at 10:59 PM, 85 percent after.
  Read it before each task with `claude -p "/usage" --model haiku`. Past the line, finish the task in hand,
  push, and wait under `## Stop` in STATUS.
- Every task is `auth-data`: failing tests first, each recorded failing; each mutation run and its failing
  output recorded; a task-scope `standards:diff-reviewer` (sonnet, medium) on each commit (adapter "Per-task
  review").
- One commit per task, task id first, specific files only, the trailer
  `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`. Never commit red, a secret, or a kit file.
- Pairs: none. Every task edits `cmd/dubplate-server/main.go` and the cmd test files; the pass runs serially in
  the executor's session, and no subagent edits files.

## The proof

Run at the launch base, then on each task's commit for T1 to T9, comparing each to the launch base:

1. Build in a fresh detached worktree under the executor's own `$CLAUDE_JOB_DIR/tmp`, never in the pass
   worktree: `git worktree add --detach <tmp>/b-<sha> <sha> && (cd <tmp>/b-<sha> && go build -o
   <tmp>/bin-<sha> ./cmd/dubplate-server)`.
2. One run at a time, with port 18431 free: `bash <kit>/structure-proof.sh <tmp>/bin-<sha> <tmp>/tx-<sha>.txt`.
   The seven lines must match the launch base's. When one differs, `diff` the transcripts and fix the code
   (standing answer 13). A run that fails to bind the port: check `ss`, wait, rerun once; a second failure is
   a stop.
3. Registry dump, in the same detached worktree: copy `<kit>/zz_registrydump_test.go` into
   `cmd/dubplate-server/`, run `go test ./cmd/dubplate-server -run TestRegistryDump -count=1 -v`, record the
   `registry dump:` line, then remove the worktree with `git worktree remove --force <tmp>/b-<sha>`. If
   `newProductionRegistry` is renamed, change only that call in the uncommitted copy and say so in the PR.
   The copy never enters the pass worktree, and `git status --short cmd/dubplate-server | grep zz_registrydump` prints nothing
   before every commit.

## Export rule

Each package exports only what a production caller outside it names, and the commit body lists each new
export with its caller. The callers known at planning, from a cross-file search of `cmd/dubplate-server`:

| Package | Export | Production callers |
| --- | --- | --- |
| `serve` | `Register`, `RegisterTrustedProxy`, the port check (M5), the serve-config reader (today `resolveServeConfig`), the trusted-proxy reader (today `resolveTrustedProxy`), `parseProxyAddrs`, `trustedProxyAddrsKey`, `trustedProxyHeaderKey` (`doctorprobe.go:67`), the three thresholds | `main.go`; `auth/config.go`; cmd `serve.go`, `doctorprobe.go` |
| `auth` | `Register`, the readers `serve.go` wires (`durationSetting`, `authBodyCap`, `genericAnswer`, `readAuthLimits`), `settingReadError` (cmd `intake.go` `newIntake`), the four auth duration settings `serve.go` names, `sessionAbsoluteCapSetting` (cmd `api.go`, `serve.go`), `magicLinkExpiryCeiling`, `authRequestBodyCapCeiling`, `isLoopbackHost` (cmd `serve.go:526`), the ten key-name constants | cmd `serve.go`, `restore.go`, `intake.go`, `api.go`; `ops/admin/config.go` (`sessionAbsoluteCapSetting` for `reauthWindowBelowCap`) |
| `ops/local` | `Register`, the job-retention reader (today `jobRetentionReader`), the backup-config assembler (today `backupConfigFrom`) | `main.go`; cmd `backup.go`, `store.go` |
| the store | `Register`, `storeBusyTimeout`, `storeMigrationWait`, `errOutOfRange` (cmd `store.go:282`, `storeRow`'s doctor wording), the snapshot-keep setting | `main.go`; cmd `backup.go`, `check.go`, `member.go`, `serve.go`, `restore.go`, `store.go`; `settings` |
| `settings` | `LocalSnapshotKeep`, `SeedFacts` | cmd `serve.go`, `store.go`, `check.go` |
| `webapp` | `Register`, and one reader over `(environ []string, stateDir string)` that returns the five values `webapp.Config` takes from settings (match, palette, branding, branding neutral, operator CSS). It checks them in today's order with today's sentences. `newWebRoutesFromFS` also names the palette key in its `ErrNoThemeSelector` wrap; the executor exports that key's name or moves the wrap behind the reader, and lists the choice | `main.go`; cmd `web.go` |
| `backup` | `Register`, the integration class, the three readers | `main.go`; cmd `restore.go`, `serve.go`; `ops/local/config.go` |
| `ops/admin` | `Register`, the reauth-window setting or its reader | `main.go`; cmd `serve.go` |
| `intake` | `Register`, `stagingKey`, `intakeSweepIntervalSetting` and `intakeCheckWorkersSetting` (`newIntake`, `intake.go:343-347`), `intakeAllowlist`, `intakeWindows`, `freeSpaceFloor`, `resolveStaging`, `intakeLimits` | `main.go`; cmd `serve.go`, `doctor.go`, `api.go`, `intake.go` |

`byKey`, `olderStoreRefusal`, and `rangeError` stay unexported. The doctor caller at `store.go:383` builds
its own map with `make`, never a `map[string]config.Setting{}` literal, which the declares grep would count.
No export exists only for a test. Every top-level name that moves is checked against the target package before the move: `magicLinkSource` already
exists as a const at `internal/auth/seam.go:11`, so the moved identity-source var takes another name (for
example `magicLinkIdentitySource`), listed in the commit body. An export that ends the pass with no production
caller outside its package is unexported before the close. A wiring function that assembles several packages
(`newServeShellVia`, `newIntake`, `authHandlersConfig`, `newWebRoutes`) stays in `cmd` and calls the exported
readers.

## `main.go` after each task

The order changes only by whole-package moves that keep every family's order (M1, M2).

| After | `newProductionRegistry` calls | New / old |
| --- | --- | --- |
| base | `registerServeConfig`, `registerAuthConfig`, `registerStoreConfig`, `registerWebConfig`, `registerBackupConfig`, `registerProxyConfig`, `registerAPIConfig`, `registerIntakeConfig` | 0 / 8 |
| T1 | `serve.Register`, `registerAuthConfig`, `registerStoreConfig`, `registerWebConfig`, `registerBackupConfig`, `serve.RegisterTrustedProxy`, `registerAPIConfig`, `registerIntakeConfig` | 2 / 6 |
| T2 | `auth.Register`, `serve.Register`, `registerStoreConfig`, `registerWebConfig`, `registerBackupConfig`, `serve.RegisterTrustedProxy`, `registerAPIConfig`, `registerIntakeConfig` | 3 / 5 |
| T3 | `auth.Register`, `local.Register`, `serve.Register`, `registerStoreConfig` (job retention removed), ... | 4 / 5 |
| T4 | ... `store.Register` in place of `registerStoreConfig` | 5 / 4 |
| T5 | ... `webapp.Register` in place of `registerWebConfig` | 6 / 3 |
| T6 | ... `backup.Register` in place of `registerBackupConfig` | 7 / 2 |
| T7 | ... `admin.Register` in place of `registerAPIConfig` | 8 / 1 |
| T8 | `auth.Register`, `local.Register`, `serve.Register`, `store.Register`, `webapp.Register`, `backup.Register`, `serve.RegisterTrustedProxy`, `admin.Register`, `intake.Register` | 9 / 0 |

The names `Register` and `RegisterTrustedProxy` are the plan's; an executor who finds a better name keeps the
`\.Register[A-Za-z]*\(reg\)` shape the acceptance grep counts. The doc comment on `newProductionRegistry`
says why `serve` registers twice, in one sentence, naming the functions without `(reg)` so the count stays 9.

## Tasks

Every task carries the same steps, so each lists only what is its own:

1. Write or move the package's tests first, at the new home, and record them failing to build or failing
   at the task's start commit.
2. Move the declarations, validators, and readers into `internal/<pkg>/config.go` with their doc comments.
   No setting name, default, doc string, description, mark, floor, ceiling, value list, or refusal sentence
   changes (standing answer 14). A comment that names a cmd function by its old name is refreshed to name the
   new one, and nothing else in it changes.
3. Point `main.go` and every cmd caller at the export. A cmd test of a moved function moves with it into
   `internal/<pkg>/config_test.go` under the same name; a cmd test that only calls a moved name takes the
   qualifier.
4. Run the mutations, `go vet ./...`, and the fast lane. Record each.
5. Commit.
6. Run the proof and the dump on that commit (they build from a sha in a detached worktree). A line that
   differs from the launch base is fixed in a `T<n> fix:` commit before anything else, and the proof reruns.
7. The task-scope review, on the task's commits. After the push, record the CI verdict on the task's head.

### T1: `serve` (40 min)

- `internal/serve/config.go` takes `serveBindAddrsThreshold`, `servePortThreshold`, `servePrefixThreshold`,
  `registerServeConfig` (as `Register`), `resolveServeConfig`, `validateServePort` with its floor and ceiling
  constants (M5), and from `proxy.go` `trustedProxyIntegrationClass`, `registerProxyConfig` (as
  `RegisterTrustedProxy`), `resolveTrustedProxy`, `parseProxyAddrs`, and `parseProxyAddr`. `serve` imports
  config only from this file.
- The routes and server code in cmd `serve.go` (`newServeCmd`, `newServeShellVia`, `openServeStore`, and the
  rest) stays.
- Forced sites: `cmd/dubplate-server/doctorprobe.go` (names the three thresholds, the class, both readers, `parseProxyAddrs`, and `trustedProxyAddrsKey` at `doctorprobe.go:303`)
  and `auth.go`'s call of the port check.
- Test-first: `TestResolveTrustedProxy`-family tests from `proxy_test.go` and the `resolveServeConfig` tests
  from `serve_test.go` that test the reader itself move to `internal/serve/config_test.go`; the rest of
  `serve_test.go`, `apimount_test.go`, `doctorprobe_test.go`, and `routes_test.go` take the qualifier.
- Mutations: drop the port floor check (a refused-values row in the proof and the moved port test fail);
  register the trusted-proxy class before backup's in `main.go` (the dump changes). Revert each.
- Acceptance: declares falls by 4 (serve 3, proxy 1) to 48 and registers by 4 to 25; `main.go` 2 / 6; the
  proof and the dump match.

### T2: `auth` (60 min)

- `internal/auth/config.go` takes `identitySourceSelectorRow`, the key-name constants, `secretKeys`,
  `magicLinkSource` (renamed, see "Export rule"), `secretLengthFloor`, `magicLinkExpiryCeiling`, the ten settings and their two rules,
  `authRequestBodyCapCeiling`, `durationAtLeast`, `rosterFileKey`, `rosterFileSupersession`,
  `registerAuthConfig` (as `Register`), `validateMagicLinkSource`, `validateSMTPPort`, `validateSMTPFrom`,
  `validatePublicURL`, `isLoopbackHost`, and the readers `durationSetting`, `authBodyCap`, `genericAnswer`,
  `settingReadError`, and `readAuthLimits`. `validateMagicLinkSource` calls `serve`'s port check (M5) at the
  same place in its join (M4).
- `internal/serve/tusmount_test.go` replaces `auth.EarliestAdmin` with a local stand-in and drops its `auth`
  import (M6). This is a forced site outside the row, listed under `## Decisions` with the probe output.
- `auth.Register` moves to the first line of `newProductionRegistry` (M2).
- `openNonceStore`, `authHandlersConfig`, `newMailer`, and `sweepNonces` stay in cmd `serve.go` as wiring.
  Forced sites: cmd `restore.go` (the four secret key names), `serve.go`, `api.go` and `intake.go`
  (`settingReadError`, `sessionAbsoluteCapSetting`), `rendered_test.go`, `serve_test.go`.
- Test-first: the `validateMagicLinkSource`, `readAuthLimits`, and reader tests in cmd `auth_test.go` move to
  `internal/auth/config_test.go`; record the build failure. Add one join-order test there: a single
  environment with a bad SMTP port and serve port 80, asserting the literal joined refusal text. Today no
  test guards the order (`TestConfigLoad_MagicLinkValidationJoinsOtherErrors` uses `Contains`, and the kit
  sets one bad value per run).
- Mutations: swap the SMTP port and serve port checks in the join (the join-order test fails); make
  `durationAtLeast` compare with `<=` (the moved "lifetime equal to the cap" case, `auth_test.go:661`,
  fails); remove the loopback exemption in `isLoopbackHost` (the public URL test fails). Revert each.
- Acceptance: declares falls by 15 to 33, registers by 6 to 19; `main.go` 3 / 5; `go vet ./...` clean; the
  proof and the dump match.

### T3: `ops/local` (30 min)

- `internal/ops/local/config.go` takes `jobRetentionSetting`, its reader (`jobRetentionReader`), and a
  `Register` that adds the one setting. `registerStoreConfig` drops its job-retention line, and
  `local.Register` goes in after `auth.Register` (M2).
- Forced sites: cmd `backup.go` and `store.go` (the reader's callers), and the cmd tests that build on job
  retention: `store_test.go`, `check_test.go`, and `doctor_test.go` (each calls `registerStoreConfig`).
- Test-first: `TestSettingsReaders` (`store_test.go:557-572`) reads job retention beside the three backup
  readers. T3 splits it: the job-retention half moves to `internal/ops/local/config_test.go`, and the backup
  half keeps the name and stays in cmd until T6 moves it. Record the build failure.
- Mutation: register `local.Register` after `registerStoreConfig` in `main.go`; the dump's setting order
  changes. Revert.
- Acceptance: declares to 32, registers to 18; `main.go` 4 / 5; the proof and the dump match.

### T4: the store and `settings` (45 min)

- `internal/store/config.go` takes `storeBusyTimeoutThreshold`, `storeMigrationWaitThreshold`,
  `storeLocalSnapshotKeepSetting`, `registerStoreConfig` (as `Register`), `storeBusyTimeout`, and
  `storeMigrationWait`, with `rangeError` and `errOutOfRange` where those two read them.
- `internal/settings` takes `localSnapshotKeep` (as `LocalSnapshotKeep`), `seedFacts` (as `SeedFacts`),
  `byKey`, and `olderStoreRefusal` (M10). They sit in one new file beside the function they call; the file
  name is the executor's.
- `TestLocalSnapshotKeep` (cmd `store_test.go:1012`) moves to an external `settings_test` package and enters
  through `settings.LocalSnapshotKeep`. Its "no keep row" case uses `jobRetentionSetting.Key` only as a row
  that is not the keep row; it uses the literal `"DUBPLATE_JOB_RETENTION"` and registers it with
  `local.Register` (M10's probe), so no export exists for the test.
- Forced sites: cmd `backup.go`, `check.go`, `member.go`, `serve.go`, `restore.go` (`storeBusyTimeout`),
  `store.go` (the doctor caller of `byKey` at 383 builds its own map; `storeRow` names the exported out-of-range
  error), `store_test.go`, `check_test.go`, `doctor_test.go`, `settings_test.go`.
- Test-first: the moved tests fail to build at their new homes; record it.
- Mutations: make `LocalSnapshotKeep` ignore the store's row (the moved test fails); swap the busy-timeout
  floor and ceiling (a refused-values row changes). Revert each.
- Acceptance: declares to 28, registers to 15; `main.go` 5 / 4; the proof and the dump match.

### T5: `webapp` (35 min)

- `internal/webapp/config.go` takes the four thresholds (`themeMatchThreshold`, `themePaletteThreshold`,
  `brandingThreshold`, `operatorCSSFileThreshold`), with `acceptedThemeMatch`, `themePaletteGrammarRE`, and
  `notConfigured`, `registerWebConfig` (as `Register`), `resolveBranding`, `brandingNeutral`, and
  `resolveOperatorCSS`, behind the one reader the export table names. The reader takes the state directory
  as a string, so `privateFiles` stays in cmd and passes `private.stateDir`. `newWebRoutes`,
  `newWebRoutesFromFS` (now calling the reader, then `webapp.New`), `registerWebRoutes`, and `webRouteNames`
  stay in cmd as wiring.
- `webapp`'s `TestLayering` fails on the new `config` import here. T5 loosens it to drop `config` only, the
  edit the design names; `auth`, `ops`, `settings`, and the store stay forbidden. (The edit lands here, not in
  T10, because the gate is red without it.)
- Test-first: write the branding and operator-CSS reader tests at the new home, or move cmd's, and record the
  failure. Trial edit for the PR: a `webapp` file imports `internal/settings`; `TestLayering` still fails
  and names it. Revert.
- Test-first also adds one order test for the reader: a bad theme match and a bad palette in one
  environment, asserting the literal refusal (the match's sentence, since the match is checked first).
- Mutations: make `resolveBranding` return neutral for every value (the moved test fails); check the palette
  before the theme match in the reader (the order test fails). Revert each.
- Acceptance: declares to 24, registers to 11; `main.go` 6 / 3; the proof and the dump match.

### T6: `backup` (45 min)

- `internal/backup/config.go` takes `backupIntegrationClass`, the three settings, `registerBackupConfig` (as
  `Register`), and `backupSettingsReaders`.
- `backupConfigFrom` builds a `local.BackupConfig`; `backup` cannot import `ops/local` (which imports
  `backup`), so the assembler moves to `internal/ops/local/config.go` and reads `backup`'s exported class and
  readers.
- Everything from `backup.go:148` on (the verb, its exit statuses, the doctor rows) stays; the doctor rows
  are `structure-doctor`'s.
- Forced sites: cmd `restore.go` and `serve.go` (the class).
- Test-first: the backup half of `TestSettingsReaders` moves to `internal/backup/config_test.go` under its
  name. No test names `backupConfigFrom` today, so write one in `internal/ops/local/config_test.go`: a
  populated environment (backup integration enabled, a remote set) and a store with non-default rows, each
  field of the result asserted against a literal. Record both failing to build.
- Mutation: read the wrong integration key for `Remote` in the assembler; the new assembler test fails.
  Revert.
- Acceptance: declares to 20, registers to 7; `main.go` 7 / 2; the proof and the dump match.

### T7: `ops/admin` (25 min)

- `internal/ops/admin/config.go` takes `reauthWindowFloor`, `reauthWindowSetting`, `reauthWindowBelowCap`
  (naming `auth`'s exported absolute-cap setting), `registerAPIConfig` (as `Register`), and, if no cmd caller
  outside the API mount needs it, the reader that `serve.go` wires into `admin.New`.
- `newAPIMount`, `intakeLimits`, and `apiVerifier` stay in cmd; `intakeLimits` moves in T8.
- Test-first: the rule test at `api_test.go:48` (`reauthWindowBelowCap`) moves to
  `internal/ops/admin/config_test.go`; record the build failure. `api_test.go` is a forced site.
- Mutation: make the rule compare with the cap's floor; the test fails. Revert.
- Acceptance: declares to 18, registers to 5; `main.go` 8 / 1; the proof and the dump match.

### T8: `intake` (55 min)

- `internal/intake/config.go` takes `stagingKey`, `stagingRow`, the intake settings, the list opener and its
  members, `intakeFormatAllowlistThreshold`, the two superseded names, `registerIntakeConfig` (as
  `Register`), `intakeAllowlist`, `intakeWindows`, `freeSpaceFloor`, `resolveStaging`, `contains` where only
  these read it, and `intakeLimits` from `api.go`. Its readers wrap errors with
  `fmt.Errorf("%s: %w", key, err)` and return nil on success (M11); `intake` never imports `auth`.
- `newIntake`, `liveAdminAddresses`, and `tusVerifier` stay in cmd as wiring.
- Forced sites: cmd `serve.go`, `doctor.go`, `main.go` (`loadConfig` calls `resolveStaging`), `api.go`,
  `check_test.go`, `intake_test.go`.
- Test-first: the `intakeAllowlist` and `resolveStaging` tests move; record the failure. No cmd test touches
  `intakeLimits`, `intakeWindows`, or `freeSpaceFloor`, so write a reader test for each at the new home first:
  a store row the reader refuses, asserting the literal key-prefixed text. Record each failing.
- Mutations: drop the staging-overlap refusal (the moved test and a refused-values row fail); return the bare
  error from each reader in turn (its reader test fails). Revert each.
- Acceptance: declares 0, registers 0, `main.go` 9 / 0 (acceptance rows 1 to 3); `intake`'s `TestLayering`
  unchanged and passing; the proof and the dump match.

### T9: the three leftovers (30 min)

- `lookupEnviron` goes; its callers call `config.Lookup` directly. `unidecode.go:82,147` are forced sites.
- `identitySourceValues` goes; `doctor.go:603` builds the same map inline from `config.Lookup` over
  `config.KeyNames(src.Required)` and `src.Optional` (M14). No new export. Listed under `## Decisions`.
- `Registry.Superseded` goes (M15). Its three test callers are rewritten to enter through `Load`, never
  deleted:
  - `cmd/dubplate-server/intake_test.go:67`: loading with `DUBPLATE_UPLOADS_DIR` or `DUBPLATE_TRIGGER_DIR`
    set refuses with a reason naming `DUBPLATE_STAGING_DIR`.
  - `internal/config/registry_test.go:70`: the subtest `Superseded` becomes a `Load` refusal check on a
    superseded name `keys.go`'s `supersededKeys` seeds (for example `DUBPLATE_LIBRARY_DIR`). Its name may change; it is a
    subtest, so the test-names count does not move.
  - `cmd/dubplate-server/templatekeys_test.go:95`: `classifyKey` loses its "superseded" branch, which makes
    the test stricter, and its error text loses "or the superseded table". First confirm no row of
    `env/dubplate-server.env.template` is superseded; if one is, keep the branch through a `Load` refusal
    instead and say so.
- Test-first: each rewritten test passes on the old code too, so each is proved by mutation: drop the
  `AddSuperseded` calls in `intake/config.go` (the intake test fails); change `keys.go`'s seed message in a
  scratch edit (the config subtest fails). Revert each.
- Acceptance: the leftovers check prints 0 lines and the method count 15 (row 6); the proof and the dump
  match. A test name gone at head is listed in the PR with the seam it tested.

### T10: the import test, the two notes, and the rule texts (50 min)

- The import test, in `internal/settings/layering_test.go` (the row's one wildcard; the test guards the
  package every listed `config.go` reads through). For `auth`, `backup`, `serve`, the store, `webapp`,
  `intake`, and `ops/local`, it parses each non-test `.go` file with `parser.ImportsOnly` (M12) and fails
  naming the file when a file other than `config.go` and `doctor.go` imports `internal/config` or
  `internal/settings`. It also fails when a listed package has no `config.go` importing `internal/config`, so
  a wrong path cannot pass vacuously. `ops/admin` is left out (design, "Settings live in the package
  itself").
- Test-first: run it once with a deliberately wrong path in its package list (for example `internal/autth`);
  it fails naming the missing `config.go`. Record it and revert. Trial edits for the PR, each failing and naming its file, each reverted: a third `auth` file importing
  config, and a third `intake` file importing settings (acceptance row 5).
- The notes: `internal/auth/doc.go` and `internal/serve/serve.go` say each package stays free of the
  configuration package. Each is rewritten to say that only `config.go` (and `doctor.go`) read configuration,
  that the rest takes plain values, and that the import test holds it.
- The rule texts. Load `claude-infra` and read `~/.claude/docs/voice/agent-facing.md` first.
  - `CLAUDE.md:188` (under "Facts agents get wrong") becomes the design's rule 1: a setting is declared in
    the `config.go` of the package that reads it and registered by that package's `Register`; the command
    folder gains a line only when a new package registers for the first time. It keeps one clause from
    today's line: `internal/config/keys.go` takes no new declaration. Rules 2 to 5 are not this plan's; the
    PR's `Calls:` line names them for the conductor.
  - `.claude/skills/dubplate-pass/SKILL.md` "### Config keys" (lines 118-121) says the same, naming the
    `main.go` function, not a line number. This edit is class `runner`, listed under `## Decisions`.
- Acceptance: the import test passes on the tree and fails on both trial edits; `grep -n 'main.go:99\|own
  file' CLAUDE.md .claude/skills/dubplate-pass/SKILL.md` prints nothing.

### Close (50 min)

The backlog edit is the close's first commit (`Close: ...`, `BACKLOG.md` only):

- #138 (line 80) and #139 (line 81) close with this pass's PR number.
- #124 (line 360) is cut to what remains after this pass and `structure-doctor`.
- #156 (line 373) is cut to what remains: the "(review m5)" comments at `auth.go:99,110` moved into `auth`,
  which the wide citation pattern skips (M13), so they stay as written there.

Then, in order: one whole-branch `standards:diff-reviewer` and one `standards:web-auth-security-reviewer` read
(opus, medium), dispatched together, each asked for correctness gaps ranked by consequence. The security read
checks that each moved validator still refuses what it refused: the magic-link source's five checks and their
join order, the port floor, the proxy address parse, the staging overlap, and every floor and ceiling. Then one
batched fix chain; a delta review only on a blocker, whose minor findings are folded once or filed. Then the
adapter's close checklist (archive this plan with `git mv` to `docs/archive/plans/`; STATUS below the round
block only). Then the proof and the dump on the final head, compared with the launch base: T10 and the fix
chain edit Go after T9, so the head figures come from this run. Last, on the final head,
`bash scripts/gate.sh --merge`; the "merge gate passes" evidence names
that head. The PR lands on its checks and reviews, as `approval-packet` step 1 says.

The PR body carries `Round: structure`, `Pass: structure-keys`, `Review:`, and `Calls:` lines; the evidence
per acceptance item; the seven lines and the dump line at base and head; each name gone at head with its seam;
the two trial outputs; and the score (`pass-core` `executor.md`, "Score").

## Path

Launch-base proof 15, T1 40, T2 60, T3 30, T4 45, T5 35, T6 45, T7 25, T8 55, T9 30, T10 50, close 50:
480 minutes, serial. The design's 360 covers the eight packages (280), the order and leftovers (40), and the
import test and notes (40). The launch-base proof, the backlog, the head proof, and the close add 65; M6's
stand-in and the new tests the plan review found missing (join order, assembler, the three intake readers,
the webapp reader) add 55. The parking line is 720.

## Acceptance

### Every pass (verbatim)

| Item | Check | Today, and to pass | Task | Evidence |
| --- | --- | --- | --- | --- |
| What a person sees is unchanged | `structure-proof.sh` on the base and head binaries | Today 2,510 lines in six sections, `7b996b43`. The two runs print the same seven lines. | T1 to T9; close | The seven lines at the launch base and on the final head (close). Today at `5d1e6e9`: 2,510 lines, `7b996b43` (above). |
| Every declaration is unchanged | The registry dump test at base and head | Today 56 lines, `b8a59fba`. The same at head. | T1 to T9; close | The `registry dump:` line at base and head. |
| The web app is untouched | `git diff --stat <base>..HEAD -- web/` | Prints nothing. | Close | The command, base the launch base. |
| The migration files are unchanged | `(cd <store>/migrations && sha256sum -- *.sql) \| sha256sum` | Six names with their hashes, `ffc4e9db`, at the old path today and the new path after. | Close | `ffc4e9dba22b...` at `internal/store/migrations` at base and head (the path moved in `structure-base`). |
| No test was dropped | `grep -rhoE '^func (Test\|Fuzz\|Benchmark\|Example)[A-Za-z0-9_]*' --include=*_test.go cmd internal \| sort` at base and head, then `comm -23` | 1,921 names today, `4918d4c4`. Each name gone at head is in the pull request with the seam it tested. | Close | 1,906 names at `5d1e6e9`, `39f06a8c` (standing answer 19). `comm -23` output, each gone name with its seam. |
| The merge gate passes | `bash scripts/gate.sh --merge` | Exits 0. | Close | `gate exit: 0`. |

### `structure-keys` (verbatim)

| Item | Check | Today, and to pass | Task | Evidence |
| --- | --- | --- | --- | --- |
| The command folder declares nothing | `grep -rnE 'config\.(Setting\|Threshold\|IntegrationClass\|WallRow\|CorrelatedRule\|IdentitySource\|ReservedClass)\{' --include=*.go cmd`, leaving out tests | 52 lines today, which are the 49 declarations, two list openers, and one empty return. 0 to pass. | T1 to T8 | The command at head: 0. 52 at `5d1e6e9`. |
| The command folder registers nothing itself | `grep -rnE '\.Add(WallRow\|ReservedClass\|IntegrationClass\|Threshold\|IdentitySource\|Setting\|CorrelatedRule\|Superseded)\(' --include=*.go cmd`, leaving out tests | 29 lines today. 0 to pass. | T1 to T8 | The command at head: 0. 29 at `5d1e6e9`. |
| `main.go` has the nine lines | `grep -cE '\.Register[A-Za-z]*\(reg\)' cmd/dubplate-server/main.go`, and the same count for the old form, `register[A-Z][A-Za-z]*Config\(reg\)` | 0 and 8 today. 9 and 0 to pass. | T1 to T8 | Both counts at head. |
| The order prints as today | The registry dump test, as for every pass | 56 lines, `b8a59fba`. | T1 to T9 | The dump line after each task. |
| Configuration stays in two files of each package | The new import test, with two trial edits shown in the pull request | The test fails when a third file of `auth` imports the configuration package, and when a third file of `intake` imports the settings package. | T10 | The two failing outputs. |
| The three leftovers are gone | `grep -rnE 'lookupEnviron\|identitySourceValues\|\) Superseded\(\)' --include=*.go cmd internal`, leaving out tests, and `go doc ./internal/config Registry \| grep -c '^func (r \*Registry)'` | 19 lines and 16 methods today. 0 lines and 15 methods to pass. | T9 | Both commands at head. |

Added checks (no item above is edited): `go vet ./...` after every task (M6); the join-order test (T2); no `zz_registrydump` file in `git status --short
cmd/dubplate-server` before each commit; `intake`'s `TestLayering` unchanged (T8); the `webapp` trial
edit (T5); the import test's wrong-path failure (T10); the rule-text grep (T10); the proof and dump on the final head
(close); the CI verdict per task head.

## Decisions

- `serve`'s port check moves into `internal/serve/config.go`, exported, and `auth`'s magic-link `Validate`
  calls it, so `auth` imports `serve` in production (M5). That edge closes a test cycle through
  `serve/tusmount_test.go`, which T2 breaks with a local stand-in for `auth.EarliestAdmin` (M6), a forced site
  outside the row. The alternatives were injecting the check through `Register`'s arguments (breaks the
  `(reg)` count), moving it into `config` (contradicts the outcome), or a `Threshold.Floor` or new hook (changes
  a refusal sentence or the join order). Settled with the advisor at planning.
- The design's cycle evidence was `go build ./...`, which compiles no test files. Whole-module `go vet ./...`
  joins every task's checks.
- `local.BackupConfig` is assembled in `ops/local/config.go` from `backup`'s exports, since `backup` cannot
  import `ops/local`.
- `intake`'s readers wrap errors with `fmt.Errorf("%s: %w", key, err)` instead of `auth.SettingError`,
  because `intake` may not import `auth` (M9). The text is identical (M11), and only `auth`'s own handlers
  assert the type.
- `byKey`, `olderStoreRefusal`, `rangeError`, and `errOutOfRange` stay unexported; their one outside caller
  builds its own map.
- `identitySourceValues` has a production caller, not the test-only one the design describes (M14, standing
  answer 19). It is inlined at that caller with no new export.
- `TestLocalSnapshotKeep` moves to an external `settings_test` package that enters through
  `settings.LocalSnapshotKeep`.
- `webapp`'s `TestLayering` loosening lands in T5, the first commit that needs it, so no commit is red.
- The import test lives in `internal/settings/`, the row's one wildcard.
- The adapter's "### Config keys" edit is class `runner`.
- Test names number 1,906 at `5d1e6e9`, not the design's 1,921 at `e03e574` (standing answer 19). This pass
  compares its own base and head.

- T10 writes the design's rule 1 only, as this pass's brief gives it. The design sends five rules to
  `CLAUDE.md` and this pass's row is the round's only one that carries the file, but rule 4 is not true of
  the tree until `structure-doctor` lands. Rules 2 to 5 go to the conductor in the PR's `Calls:` line.
- `errOutOfRange` is exported from the store, because `storeRow` stays in cmd and branches on it.
- `webapp` gets one reader that returns the five settings values `webapp.Config` takes, so
  `newWebRoutesFromFS` stays in cmd as wiring with no threshold of its own.
- The moved `magicLinkSource` var is renamed: `internal/auth` already has a const of that name.

## Plan review fold

One `standards:plan-reviewer` read (opus, medium) at `5d1e6e9`. It confirmed the registration order at every
intermediate state, the countdowns, the import edges (a scratch tree with every edge and the stand-in ran
`go vet ./...` clean, and only `webapp`'s `TestLayering` failed, on `config`), and the verbatim tables.
Folded, all fifteen. An advisor read afterwards fixed three checks that could not pass or fire (the T10 grep,
the `git status` check, T5's order mutation) and the path sum:

1. Names the cmd wiring still reads were missing from the export table (`errOutOfRange`, the webapp
   thresholds, `isLoopbackHost`, `trustedProxyHeaderKey`, two intake settings). Added, with one webapp reader
   for T5.
2. Three mutations could not fail (T2's join swap, T6's type-mixing swap, T8's reader test that did not
   exist). T2 gains a join-order test, T6 an assembler test and a type-correct mutation, T8 three reader
   tests written first.
3. T3 broke `TestSettingsReaders` and `TestLocalSnapshotKeep` with no test file listed, and T4 implied a
   test-only export. T3 splits the test, T4 uses the literal key, and the three test files are listed.
4. Head was never fingerprinted after T10 and the fix chain. The close runs the proof and the dump on the
   final head.
5. `magicLinkSource` collides with a const in `internal/auth`. Renamed, with a collision check on every move.
6. M11 misplaced `api.go:43` and `newIntake`; T2's forced sites missed `api.go` and `intake.go`; T7 missed
   `api_test.go:48`. Corrected.
7. M14 misquoted `lookupEnviron`'s signature and the inline needs `config.KeyNames`. Corrected.
8. A map literal or a `(reg)` in a comment would trip rows 1 and 3. The plan says `make` and names the
   functions bare.
9. The `<=` mutation names its equal-values case.
10. M9 named T11 for the webapp loosening. Now T5.
11. T10's missing-file failure could not happen. It runs on a deliberate wrong path.
12. Each task records its CI verdict.
13. T11 folded into the close.
14. The five rules: T10 writes rule 1, keeping the `keys.go` clause, and rules 2 to 5 go to the conductor
    (see `## Decisions`).
15. The PR body carries the score.
