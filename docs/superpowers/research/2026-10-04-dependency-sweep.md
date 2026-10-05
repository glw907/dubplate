# Dependency sweep, 2026-10-04

Scope: every dependency of the built tree (the Go module, `web/`, the Python
oracle venv, and the pinned runtime binaries). `_salvage/` is out of scope.
Each package's release notes were read for exactly the range below and
checked against the repo's call sites. The next sweep starts from the
"after" column.

## Before table

| Package | Current | Target | Kind |
|---|---|---|---|
| modernc.org/sqlite | v1.59.0 | v1.60.1 | Go direct |
| modernc.org/libc (and cc, ccgo, gc) | v1.75.7 | v1.77.1 | Go indirect, pinned by sqlite |
| golang.org/x/crypto | v0.54.0 | v0.57.0 | Go indirect (go-mail NTLM) |
| golang.org/x/net | v0.56.0 | v0.59.0 | Go module graph only |
| golang.org/x/text | v0.40.0 | v0.42.0 | Go indirect (go-mail precis) |
| x/sys, x/term, x/sync, x/mod, x/tools | various | latest | Go indirect |
| go-humanize, go-strftime, pflag, yaml/v3, go-md2man, pprof | various | latest | Go indirect |
| @sveltejs/kit | 2.70.3 | 3.0.0 | npm, MAJOR |
| @sveltejs/adapter-static | 3.0.10 | 4.0.0 | npm, MAJOR |
| typescript | 6.0.3 | 7.0.2 | npm, MAJOR |
| @types/node | 24.13.6 | 26.6.4 (24.19.1 in-major) | npm, MAJOR |
| daisyui | 5.7.46 | 5.7.47 | npm patch |
| vite | 8.3.1 | 8.3.2 | npm patch |
| charset-normalizer | 3.5.1 | 3.5.2 | Python venv |
| platformdirs | 4.11.12 | 4.12.3 | Python venv |
| Navidrome | 0.63.2 | 0.64.2 | runtime image |
| cloudflared | 2026.8.2 | 2026.9.3 | runtime image |
| rclone | v1.74.3 | v1.75.1 | runtime binary |

Already latest, verified: cobra v1.10.2, go-mail v0.8.1, x/time v0.16.0,
Go 1.27.1, golangci-lint v2.14.0, svelte 5.57.1, svelte-check 4.7.6,
@sveltejs/vite-plugin-svelte 7.3.1, tailwindcss and @tailwindcss/vite 4.3.3,
playwright 1.63.0, beets 2.14.1.

## Go

### modernc.org/sqlite v1.59.0 to v1.60.1

Features to leverage: nothing. `StrictPragmas` (opt-in, refuses a
multi-statement `_pragma`) guards DSNs built from untrusted input. Every DSN
here goes through `buildDSN` (internal/engine/store/conn.go:198), which
escapes the path through `url.URL`, and the one `_pragma` value is a
constant (conn.go:212). Enabling it costs one line and buys defense in
depth only; optional.

Practices to change: nothing. The new `Driver.Open` docs warn that a plain
file name has its query stripped, so `?mode=ro` opens read-write. The store
already uses `file:` URIs for every class (conn.go:139-198), so
`inspectDSN`'s `mode=ro` (conn.go:187) is honored. `go get` must move
`modernc.org/libc` to v1.77.1 in the same change (the sqlite module's
standing rule); v1.60.0 also requires Go 1.26, which 1.27 satisfies.

Gate risks, ranked:
1. A fault reading the WAL `-shm` mapping now returns `SQLITE_IOERR_IN_PAGE`
   (8714) instead of crashing. 8714 masks to primary code 10, so
   `endsTransaction` (internal/engine/store/write.go:238-249) already
   treats it as transaction-ending. No code change; the store suite covers
   the path.
2. `vfs.FS.Close` now refuses while a database is open, and nine dead `lib`
   constants are gone. The repo uses neither `vfs` nor `lib`. Nothing.
3. v1.60.1 makes parameter binding linear. No shape here binds thousands of
   parameters. Nothing.

### modernc.org/libc v1.75.7 to v1.77.1 (with cc/v4, ccgo/v4, gc/v3)

The runtime the transpiled SQLite runs on; no changelog of its own. It
carries the fault-handling support sqlite v1.60.0 relies on. Take with
sqlite as a pair. Nothing else relevant.

### golang.org/x/crypto v0.54.0 to v0.57.0

Only `md4` is linked (go-mail's NTLM auth). Every change in range is in
`ssh`, `acme`, `ocsp`, `x509roots`, or riscv64 poly1305. Nothing relevant.

### golang.org/x/net v0.56.0 to v0.59.0

Not linked into any package (`go mod why` reports the main module does not
need it); it is a module-graph entry only. `http2.Transport` and
`http2.Server` are now deprecated, which touches nothing here. Nothing
relevant.

### golang.org/x/text v0.40.0 to v0.42.0

Linked through go-mail's `secure/precis`. In range: a precis Nickname
short-buffer fix and several `unicode/norm` composition fixes. The repo's
own code imports no x/text package. Nothing relevant in range.

Out-of-range finding: internal/engine/musicbrainz/query.go:235-238 rejects
`golang.org/x/text/cases` because "it adds a module dependency". That
premise is stale. x/text is already in go.mod, and `x/text/cases` is
already linked into the binary through precis. The rejection may still
stand on its second ground (the divergence moves candidates, not graded
distances), but the comment should drop the first. File it to the
musicbrainz package's next touch; not a dependency action.

### Remaining indirects

x/sys, x/term, x/sync, x/mod, x/tools: platform constants, cpu detection,
and tooling changes; none reach a call site. go-humanize v1.1.0 and
go-strftime v1.1.0 arrive through libc and are CI and doc churn here.
pflag v1.0.10 fixes a deprecation comment. yaml/v3, go-md2man, and pprof
are cobra doc-generation and profiling deps that are not linked.
gopkg.in/check.v1 is test-only transitive noise. Nothing relevant.

## web/

### daisyui 5.7.46 to 5.7.47

One library change: `.btn-disabled` now applies only together with `.btn`
(`.btn:is(.btn-disabled, :disabled, ...)`), fixing `btn-link` plus
`btn-disabled` color. No stock component newly replaces a home-grown one.

Practices to change: the flat-button lists in
web/src/lib/styles/components.css:218-233 still name a bare `.btn-disabled`
beside the `:is(...)` form. Mirroring DaisyUI's new selector is cosmetic;
no action required.

Gate risks: the selector's specificity rose from one class to two, so
re-run the rendered leg's disabled and link rows:
web/scripts/rendered/run.ts:1237-1238 (fixtures), :2108-2109 (flat-button
check), and :2265 (btn-link ink). Our overrides sit in `:where()` and should
be unaffected, but this is the one shape the patch lands on.

### vite 8.3.1 to 8.3.2

Bug fixes (CSS preload with queried `renderBuiltUrl`, bundled-dev
sourcemaps, port 0) plus a non-major dependency refresh, rolldown included.
Gate risk: the refresh can change emitted bytes. The byte-for-byte leg
(Makefile:81-89) compares two builds of one source, so it still holds; the
CSP-hash agreement test (internal/serve/csp.go:29) re-derives from the
build. Run the full gate; nothing else.

### Majors

See the "Majors" section below.

## Python oracle venv

### charset-normalizer 3.5.1 to 3.5.2

Fixes UTF-8 CJK JSON misdetected as PTCP154 and alias-less encodings. It
feeds requests only when a response carries no charset; MusicBrainz JSON
declares UTF-8, and requests decodes `.json()` by its own UTF sniff. It
cannot change beets output. Nothing.

### platformdirs 4.11.12 to 4.12.3

4.12.0 creates user directories at 0700, raises when no home directory
resolves, and validates `XDG_RUNTIME_DIR`; 4.12.2 and 4.12.3 fix site-path
lists. The derivation scripts set `BEETSDIR` to an isolated directory before
beets imports (internal/harness/oracle/derive_file_tags.py:190-197,
derive_album_info.py:73-84), so beets never resolves a platformdirs config
path. It cannot change beets output. Re-run `make derive` once after the
bump as the standing oracle check; expect no diff.

## Runtime binaries (nothing deploys before rung 5d)

### Navidrome 0.63.2 to 0.64.2

GHSA-82gh-4ggp-gfg5 (0.64.0) covers share-ownership spoofing through a
client-supplied `userId` on `createShare`, plus a plugin HTTP SSRF bypass.
With sharing off (the default, and Geoff's 2026-09-25 hold) and no plugins
configured (config/navidrome.toml), neither path is reachable. That
advisory alone does not force the bump.

The range carries fixes that do apply with sharing off, and they change the
open question's framing:
- 0.64.1: failed Subsonic logins were never throttled, so passwords could be
  brute-forced at full speed over the public tunnel. Now rate limited.
- 0.64.0: the login rate limit could be bypassed with spoofed
  `X-Forwarded-For`, `X-Real-IP`, or `True-Client-IP`, which matters behind
  cloudflared.
- 0.64.1: the library filter was skipped on bookmarks, playlist tracks, and
  now-playing, and any authenticated user could take over another's player.
- 0.64.2: the admin password was written to the log when initial admin
  creation failed.

Migration notes: 0.64.0 re-encodes every internal ID (back up the database
first; clients that cache IDs for offline downloads must re-sync), rejects
negative durations, and warns about unrecognized config keys. Our keys
(`LogLevel`, `MusicFolder`, `DataFolder`, `UIWelcomeMessage`,
`Scanner.Schedule`, `Scanner.WatcherWait`, and the `ND_` env names in
compose.yaml:31-47) are all still valid. `WatcherWait = "0s"` is not
negative. 0.64.2 also fixes `database is locked` floods during scans on slow
storage. The 0.64.1 ExtAuth logout fix is worth knowing for rung 12 next to
internal/serve/strip.go:33.

### cloudflared 2026.8.2 to 2026.9.3

Almost all of the range is Quick Tunnel protected mode (login, JWKS,
callback authorization), which a remote-managed named tunnel does not use.
2026.9.1 deprecates `--transport-loglevel`; compose.yaml runs
`tunnel --no-autoupdate run` and sets no such flag. Also in range: Go
1.26.8, gRPC 1.83.1, and a gorilla/websocket CVE bump. Nothing relevant;
take at the next deploy.

### rclone v1.74.3 to v1.75.1 (v1.74.4 included)

Security fixes in range are in serve, rc, ftp, sftp, webdav, s3 redirect
handling, and the local backend's symlink and metadata paths (setuid,
setgid, and sticky bits are no longer restored from metadata by default).
None widens anything the backup job does. Crypt gains a fix for directory
names that look like versioned file names, plus a warning about legacy
ones. Our stamps (`20060102T150405.000000Z`,
internal/engine/store/stamp.go:11) do not match rclone's `-vYYYY-MM-DD-...`
version pattern.

Gate risks, ranked:
1. internal/backup/rclone.go:111-124 treats `lsjson`'s exit 3 as an empty
   directory, a behavior measured on v1.74.3. Re-run `internal/backup`'s
   tests against v1.75.1 (they drive the real binary) and re-measure that
   exit code by hand.
2. v1.74.4 fixes "command line flag ignored when set to its default value".
   The features probe passes `--low-level-retries 1`
   (cmd/dubplate-server/backup.go:522), which is not the default, and the
   `--max-delete` bound (internal/backup/backup.go:134) is computed. Neither
   was affected; the backup tests confirm it.
3. The runbook's "rclone v1.74.3+" floor can stay. Raise it to v1.75.1 only
   if the box takes the new build, since the floor records the measured
   version.

## Refactor candidates

| Item | Hand-rolled code | Recommendation |
|---|---|---|
| sqlite `StrictPragmas` | internal/engine/store/conn.go:198 (DSN builder, constant `_pragma`) | Optional. Take with the bump only if the store's next touch wants defense in depth. Low value, since no DSN is caller-controlled. |
| Stale "adds a module dependency" rationale | internal/engine/musicbrainz/query.go:235-238 | File to the musicbrainz package's next touch: correct the comment, and decide `x/text/cases` on its second ground alone. Not part of this sweep. |
| Flat-button selector parity | web/src/lib/styles/components.css:218-233 | Nothing to take. DaisyUI's change needs no mirror. |

No package in range provides something the repo hand-rolls.

## Majors

### @sveltejs/kit 2.70.3 to 3.0.0 (with adapter-static 3.0.10 to 4.0.0)

adapter-static 4.0.0's only breaking change is "require SvelteKit 3"; the
two move as one unit.

Peers: kit 3 requires Node 22.17 (we run 24.20), TypeScript ^6 (we have
6.0.3), svelte 5.57.1 or later (pinned exactly), Vite ^8.0.12 (8.3.x), and
vite-plugin-svelte 7 (7.3.1). No peer blocks. Kit 3 does not accept
TypeScript 7.

Breaking changes that touch our code:
- `svelte.config.js` is no longer read. The adapter, `version.name`
  (`sourceVersion()`), `csp`, `typescript.config`, and `preprocess` move into
  `sveltekit({...})` in web/vite.config.ts. Readers of svelte.config.js
  must follow: web/scripts/gates/server-files.ts:35 (`hooksServerEntryBase`
  imports it), `sourceVersion()`'s input list (it hashes svelte.config.js),
  and the Makefile:89 and internal/serve/csp.go:29 comments that name it.
- `$lib` becomes `#lib`, declared in `package.json` `imports`, with
  explicit extensions: 29 import lines across src/
  (`$lib/announcer.svelte` becomes `#lib/announcer.svelte.ts`).
- `$app/environment` is renamed to `$app/env`: five files
  (routes/members, login, login/link, +error, SiteHeader.svelte).
- `resolve()` and `asset()` paths drop the leading slash: eight call sites
  (lib/api/post.ts:34, lib/api/link-token.ts:25,
  routes/members/+page.svelte:67, routes/login/link/+page.svelte:79,
  SiteHeader.svelte:85 and :104, Mark.svelte:17).
- `ClientInit` moves to `@sveltejs/kit/hooks` (src/hooks.client.ts:1).
- web/tsconfig.json:2 extends `$app/tsconfig` instead of
  `./.svelte-kit/tsconfig.json` and must carry its own `include`, which
  may replace the `config.include.push('../scripts/**/*.ts')` hook in
  svelte.config.js.
- `version.pollInterval` now defaults to one hour, and `updated` is
  re-checked on focus and visibility change. The client will fetch
  `/_app/version.json` hourly and on every tab focus. The app never reads
  `updated`, so set `pollInterval: 0` or confirm the Go server serves that
  file with a no-cache policy (internal/serve/document.go:195 and :298).
- Server-only modules are now any `server` path segment or `server`
  directory outside src/routes. We have none, and the server-file gate
  forbids route and hooks entry points, a different set. No change, but the
  gate's doc comment should note the distinction.

Unaffected: no `$app/stores`, `base`, `goto` scroll or focus options,
`error(status, {...})`, `data-sveltekit-*="off"`, cookies, or form actions.
`goto(resolve('/login'), { state })` resolves to a route, so the new
rejection does not apply.

Migration size: multi-step. `npx sv migrate sveltekit-3` handles the
mechanical renames. The config move, the gate script that imports
svelte.config.js, the tsconfig include, and the determinism, CSP, and
rendered legs each need a re-run. Expect one plan task for the migration
and one for the gate adjustments.

User-visible effect: none intended. Same pages, same CSP. The only new
behavior is the version poll, unless disabled.

npm audit: the three LOW advisories (cookie GHSA-pxg6-pf52-xh8x through kit
and adapter-static) are not reachable. The `cookie` package runs only in
kit's server runtime, which here executes at build-time prerender and under
`vite dev`. The shipped client bundle does not contain it; the one cookie
the client sets is a constant ASCII name written through `document.cookie`.
The advisory is no reason to hurry.

Recommendation: hold. Take kit 3 and adapter-static 4 together as a
dedicated web task at the next `web/` rung (5d or the rung 7 uploader,
whichever opens first), before that rung writes new pages, so new code
lands on `#lib` and the new `resolve` shape. Unblock condition: a rung
plan names the migration as its first task.

### typescript 6.0.3 to 7.0.2

TypeScript 7 is the native (Go) compiler. svelte-check 4.7.6 declares
`typescript: ^5.0.0 || ^6.0.0`, kit 3.0.0 declares `^6.0.0`, and
sveltejs/language-tools#3063 ("Typescript 7 crashes svelte2tsx and
svelte-check") is open. svelte-check's experimental `--tsgo` mode has open
correctness bugs (#3127: it reports 0 errors when the compiler process is
killed), which disqualify it from a gate.

Migration size: blocked; no code change would help.

Recommendation: hold. Unblock condition: svelte-check and kit both declare
`^7` peer support and #3063 is closed. 6.0.3 is the newest 6.x.

### @types/node 24.13.6 to 26.6.4

`@types/node` describes the Node runtime the scripts run on, so it should
track the runtime major. The workstation runs Node 24.20.0 (current LTS
line 24.21.0). Node 26 is "Current" until it becomes LTS in late October
2026. Taking 26 types on a 24 runtime would type-check APIs that do not
exist at run time. The gate scripts use `node:fs`, `node:path`, `node:url`,
`node:zlib`, `node:crypto`, and `node:fs/promises`, all stable across both.

Recommendation: take the in-major update now (24.13.6 to 24.19.1, the
newest 24.x; minor, no ruling needed) and hold the major. Unblock
condition: mise moves the workstation to Node 26 once it is LTS. Bump
`@types/node` to 26 in the same change as the runtime.

## Rulings and outcome

Minors and patches landed at `0677934` (full gate green, diff-reviewer accept). Every
module go.mod lists is at its newest non-major. The graph-only modules `go list -m -u all`
still reports (x/net, x/tools, yaml/v3, pprof, ccgo/v4, gc/v3, go-md2man, check.v1) are
not in go.mod, provide no package to the build, and stay where minimal version selection
puts them. `make derive` was not re-run after the two Python bumps; run it once before the
next oracle-dependent rung.

Geoff ruled on 2026-10-04:

- SvelteKit 3 with adapter-static 4: take now, as its own task before rung 5d-i.
- Navidrome: no further bumps. The 0.64 question is closed, and later sweeps skip it.

Held, each with its unblock condition:

| Package | Held at | Unblocks when |
|---|---|---|
| typescript 7 | 6.0.3 | svelte-check and @sveltejs/kit declare a `^7` peer range, and sveltejs/language-tools#3063 is closed |
| @types/node 26 | 24.x | Node 26 reaches LTS and the workstation's mise Node moves to 26; bump both in one change |
| rclone v1.75.1 (workstation) | v1.74.3 | Fedora ships v1.75.x for the workstation binary the backup tests drive; the box takes it at 5d |

Refactor candidates: nothing in range replaces hand-rolled code. The stale reason at
`internal/engine/musicbrainz/query.go:235-238` (x/text/cases is already linked through
go-mail) is corrected the next time that package is touched.

## SvelteKit 3 migration

Taken at `b3c2578` (the official `sv migrate sveltekit-3` codemod, then hand correction) and
review-folded at `4a3e23e`, both through the full gate. diff-reviewer, svelte-reviewer, and
daisyui-a11y-reviewer read it. Owner-visible behavior: none. The login link page's heading takes
focus only after `confirm()`'s async outcome, as W7.16 specifies. The heading-focus `$effect` would
have started taking focus on first load under kit 3, because kit 3 no longer resets focus after
mount, and that `$effect` is gone. Rendered rows now pin the settled focus on load and after the
transition. `pollInterval: 0`, since nothing reads `updated`. `engines.node >= 22.17`, kit 3's floor.

Refused, with reasons: the gate's import of `vite.config.ts` for `kitOptions` loads the plugin
stack, a cost accepted for one config home; the `browser` guards under `ssr = false` and
`members/+page.svelte`'s `apiURL()` predate the migration. Filed to the next a11y pass:
`aria-describedby` from the link page's h1 to `#outcome`, a full tab-order sequence check in
`walkFocusOrder` (it proves coverage, not order), and a check that the sign-out `goto` to `/login`
announces once under kit 3's route announcer.
