# dubplate Go rewrite: design

**Status: approved 2026-08-30; estate rename ratified 2026-08-31.** The estate,
repo, binary, paths, and env prefix ship as **dubplate** (dub culture: the
exclusive pressing cut for your own sound system); every `musicbox` reference
below reads as the pre-rename name, and the plan carries the rename tasks.
Brainstormed 2026-08-30 during the architecture refinement pass; scope,
placement, and sequencing ratified by Geoff (all box-side scripts, inside the
musicbox repo, next pass after the refinement fix wave, upload-check builds on
the result). This revision folds 31 findings from four adversarial reviews
(design soundness, behavioral coverage and migration, adopter's path, Go
idiom), all run before Geoff's read at his direction.

## Intent

Replace the box-side shell layer (`music-import`, `music-backup`, `music-beet`,
`musicbox-ping-failure`, the disk and Navidrome check scripts, and `lib.sh`)
with one Go binary. The refinement review found the shell layer at the edge of
what bash does reliably: its recurring defect classes (`set -e` swallowed in
pipelines, fail-open error paths, falsy-zero arithmetic, a lock contract copied
three ways) are structural properties of large shell programs, and they
disappear in a language with wrapped errors and real tests. The rewrite ports
behavior, not shape: where tests pin behavior the port preserves it; where the
review found the behavior itself defective, the fix wave's corrected behavior
is what gets ported.

## Why Go (language decision, interrogated 2026-08-31)

Geoff pressure-tested the language choice against the obvious alternatives;
the rationale, so the question stays answered: **Python** wins on beets
integration (importable library, no tag-log parsing) but forfeits the
single-static-binary distribution that is the adoption goal's cornerstone,
and its environment fragility is a failure class this box has already
exhibited live (the beets replaygain env failure); the beets-database read
recovers most of the integration win from Go. **Rust** ties on the binary
and exceeds on type rigor, but its decisive advantages (memory safety sans
GC, fearless concurrency, performance) target risks this sequential
subprocess-orchestrator does not have, while its costs land on the
ten-year maintainer, who works in Go daily on a workstation with Go
conventions, gates, and reviewers and no Rust footprint. **tusd**, the one
hard uploader dependency, embeds as a Go library and nowhere else. Go is
the corner where the product boundary, the failure model, and the
maintainer agree.

## Goals

- One static binary, subcommand per current entry point, systemd-invoked
  exactly as the scripts are today.
- The review's shell defect classes become structurally impossible, not
  individually patched.
- The pipeline's contracts (lock, ping, notify, gate, reject wording) each
  live in exactly one package.
- Portability of the binary: no site-specific value, path assumption, or
  identity default lives outside `internal/config`. (The binary is a minority
  of the repo's site-bound surface; cloud-init's SSH key, the tunnel config's
  account id, the hostnames, and Hetzner-specific provisioning remain
  site-bound and are the packaging initiative's inventory, not this pass's.)
- Adoption gets cheaper, not just cleaner: a `doctor` subcommand replaces the
  runbook's hand-run preflight, and monitoring works against any per-check
  ping URL provider, with the pings Worker as one such provider.
- **Ease of setup by a stranger is a primary product goal of the estate**
  (Geoff, 2026-08-31), not a someday ambition. The audience: somebody wanting
  a hosted family music system, assembling it from ingredients they bring —
  Navidrome, beets, this binary, their own Cloudflare account and domain, and
  a VPS from any provider. dubplate is the recipe and conductor for those
  ingredients; cloud-init is the provider-portable provisioning substrate
  (Hetzner automation is a reference module, not a requirement). This pass ships the binary-side
  half; the packaging pass (provider seam for provisioning, scripted
  Cloudflare wiring, parameterized pings provision, non-secret config file,
  adopter-facing docs) is planned work. Its first design decision is the
  distribution model: the lean candidate is binary-as-installer (templates
  embedded via go:embed, an `init` subcommand scaffolds the instance, the
  adopter downloads one binary and never clones; the repo serves
  developers), with clone-and-configure as the fallback floor (then no
  tracked file may carry site state — contributors.yaml, the tunnel config,
  and filebrowser.yaml hostnames move to untracked or generated config — so
  a clone stays updatable from upstream). Every pass between now and then
  takes adopter cost into account when it designs a surface.

## Non-goals

- No daemon, no internal scheduler, no inotify watcher. systemd timers remain
  the scheduling layer; `Persistent=` catch-up and `OnFailure=` stay the
  operational truth.
- No Go rewrite of `deploy.sh` or `provision.sh`. `deploy.sh` does get edited
  by the migration (it must build and ship the binary and shrink its script
  manifest per flip); it stays shell.
- No replacement of beets, rclone, or the sqlite3 CLI; they remain
  subprocesses. `.backup`'s consistent snapshot and beets' tagging have no
  cheap native equivalent.
- No packaging work now. Portability is a constraint on this design; the
  packaging initiative owns the estate-level inventory (provisioning,
  Cloudflare wiring, the pings provision scripts).

## Shape

A Go module rooted in the musicbox repo, laid out per the workstation
go-conventions standard (invoked before any Go is written):

- `cmd/musicbox/` — cobra CLI, business logic kept out of `cmd/`:
  - `musicbox import`, `musicbox backup`
  - `musicbox check <disk|navidrome>` (`Args: cobra.ExactArgs(1)`; the bare
    `check` parent errors rather than printing help and exiting 0, so a typoed
    `ExecStart` cannot look healthy to systemd)
  - `musicbox beet -- <args>` — blocking flock on the import lock, then
    `syscall.Exec` of `beet`, preserving tty, argv, and exit code; replaces
    the `music-beet` shim (no unit to flip; its migration step swaps the file
    at `/usr/local/bin/music-beet`)
  - `musicbox ping-fail <slug>` — the `OnFailure=` chain's entry point,
    replacing `musicbox-ping-failure`
  - `musicbox doctor` — validates parsed config, prints every required field
    and whether it resolved, probes SMTP, R2, Navidrome, and the ping URLs
    read-only, exits nonzero on any miss; retires bring-up runbook step 1
  - `musicbox restore --to <dir>` — pulls the library and database snapshots
    back from R2 into a target tree, read-only toward the bucket, verifying
    the databases; the executable DR drill, and the adopter's path for
    migrating an existing collection in (added 2026-08-31, Geoff)
  - `musicbox version`
- Built `CGO_ENABLED=0 GOOS=linux GOARCH=amd64` on the workstation, shipped by
  `deploy.sh` alongside the config files it already ships.
- `RunE` returns errors only; `main()` maps them to exit codes (via one
  `exitCoder` check only if codes actually differ); root sets `SilenceUsage`
  and `SilenceErrors`; no `os.Exit` under `internal/`.
- `signal.NotifyContext` in `cmd/`; `ctx` is the first parameter through
  `internal/` (never stored in a struct); every subprocess runs under
  `exec.CommandContext` so SIGTERM from systemd terminates children.

## Packages

Drawn by responsibility, not by source script. Single-function units with one
consumer are files inside their consumer, not packages.

- `internal/config` — the portability wall and the only place environment is
  read (values arrive via systemd `EnvironmentFile=`, so this is `os.Getenv`
  plus validation, no file parser; child-process environments are also
  constructed here). Typed struct plus `PingURLs map[string]string` populated
  by prefix scan. Site-identifying fields (admin email, SMTP host/user/from,
  R2 account/bucket/remote, Navidrome URL) are required with no defaults, so
  a misconfigured adopter install fails at `doctor`, never silently uses this
  estate's values; only behavior thresholds (quiescent minutes, free-space
  floor, archive cap, retention days, disk threshold) carry defaults. Loads
  through a composable values source (env now; the packaging initiative can
  overlay a non-secret config file without touching call sites). Also owns
  `contributors.yaml` parsing (the module's one YAML dependency), retiring the
  `yq | jq` subprocess pair.
- `internal/lock` — the flock contract: one lock-path definition; import and
  beet contend on it (non-blocking and blocking respectively), backup holds
  its own. The lock `*os.File` is held for process lifetime; for
  `musicbox beet` the fd stays open across `syscall.Exec` so the lock rides
  the child.
- `internal/ping` — a provider-agnostic per-check ping client: takes a URL,
  sends success (GET) or failure (POST to `<url>/fail`), honors the
  no-op-when-unset rule, carries the retry and timeout budget. The pings
  Worker and healthchecks.io are two conforming providers; an adopter needs
  no Worker to run a music server. The `MUSICBOX_HC_*` names are renamed
  `MUSICBOX_PING_*` during the migration window (the `HC_` prefix names a
  provider the variables no longer point at).
- `internal/notify` — message construction and SMTP delivery via
  `wneessen/go-mail`: implicit TLS on 465, From/Date/Message-ID synthesis,
  RFC 2047 headers and quoted-printable bodies (album names are not ASCII).
  Chosen over frozen `net/smtp`, which cannot do implicit TLS or help with
  any of the above. msmtp stays installed until a notify parity test passes;
  journald becomes the delivery record, retiring `msmtp.log`.
- `internal/reject` — the verdict taxonomy: raw gate outcome to
  plain-language reason to rendered contributor email from
  `config/reject-email.txt` (which remains a deployed runtime asset). This,
  not `gate`, is the surface upload-check actually shares: its spec commits
  to "the same friendly wording", not the same probe mechanism.
- `internal/gate` — format and quality gating for staged files: extension
  policy, symlink rejection, `flac -t` full-decode integrity (pinned by the
  corrupt-FLAC test; not downgraded to a header probe). Takes a plain
  `gate.Policy` struct that `config` constructs, so it imports no config and
  stays extractable. No interface until a second consumer exists in-tree.
- `internal/archive` — safe extraction, natively via `archive/zip` and
  `archive/tar` (dropping the undeclared bsdtar dependency): per-archive
  isolated staging, extracted-size caps, free-space preflight before any
  extraction.
- `internal/review` — filing and lifecycle: collision-safe unique
  destination naming (with a clock seam for its timestamped names),
  move-to-review with the quarantine second-chance fallback and its
  could-not-move summary, retention sweeps for both review and quarantine,
  and EXDEV-aware moves that preserve the 2000:2775 setgid tree (`os.Rename`
  alone cannot cross filesystems).
- `internal/beets` — `beet` subprocess wrapper: runs imports, writes and
  re-reads the per-run tag log under the state dir, parses the skip lines,
  and triggers the Subsonic scan with credentials in an `http.Request` URL
  that never touches argv (the `curl --config` temp-file dance dies).
- `internal/importer`, `internal/backup`, `internal/restore` — orchestration
  of the above; restore is backup's read-only inverse, sharing its rclone and
  sqlite3 subprocess seams.
- File mutations go through the standard's `atomicfile` helper; sensitive
  files at 0600.

## Error-handling doctrine

Three outcomes per run, plus a best-effort margin:

- **Per-run failures fail closed**: lock-file errors, free-space floor,
  missing or invalid config, R2 or beets-database failures end the run with a
  wrapped error and nonzero exit. Sentinels (`lock.ErrHeld`,
  `config.ErrMissing`, and peers) checked via `errors.Is`; no error type
  hierarchies.
- **Per-item failures accumulate**: one contributor's bad archive, one
  unmovable file, one missing email address is recorded in the run report and
  never aborts the run. This is today's deliberate shell behavior; the
  refinement review's worst live bug (a bare `notify` aborting the whole run)
  was exactly a per-item error escaping to per-run scope, and the doctrine
  exists to make that impossible.
- **Lock contention is healthy**: the held-lock path pings success and exits
  0, touching nothing — a pinned spec decision, preserved so an overlap never
  becomes a false alert.
- **Best-effort margin**: ping, notify, and the scan trigger log failures and
  never affect exit status.

Failure pings for the timer-run commands come from systemd's `OnFailure=`
chain (now `musicbox ping-fail`), which also catches crashes and OOM kills;
the binary does not double-ping its own failures in-process. Logging is
`slog` with `NewTextHandler` to stderr into journald.

## Testing

- Table-driven unit tests per go-conventions. The bats suites map three ways,
  written into the plan as an explicit table: cases that port as-is
  (behavioral contracts), cases re-expressed against `httptest` (the curl
  argv and `--config` assertions describe curl, not the contract), and cases
  needing a new seam (the frozen-clock naming tests get `review`'s clock
  seam). The subprocess seam is a plain path field (`BeetPath string`) on the
  wrapper config, never a mock interface.
- An `e2e/` harness builds the real binary and stub `beet`/`rclone`/`sqlite3`
  Go programs from `TestMain`, runs against the existing fixture inbox, and
  speaks to a local SMTP test server (dependency chosen in the plan, not at
  implementation time).
- `check disk` and `check navidrome` port the fix wave's new shell checks;
  their behavior is pinned by the wave's bats tests, the newest part of the
  suite.
- `make check` (build, vet, staticcheck, tests) joins `scripts/check.sh` so
  one gate covers both layers during the migration window.

## Migration

Per-subcommand, in dependency order (checks, then backup, then import), with
`deploy.sh` edited in lockstep — its `SCRIPT_FILES`/`UNIT_FILES` manifests
abort the whole deploy on a missing source, so every flip edits the manifest
in the same commit:

0. **Ship the binary**: deploy.sh gains the cross-compile-and-ship step and
   `musicbox doctor` runs green on the box before any flip. New non-secret
   env names (SMTP settings, renamed ping URLs) enter `internal/config` with
   defaults or get minted into the age store and registry as part of this
   task; the env template stays a secrets manifest.
1. Port a subcommand; its Go tests plus the mapped suite pass.
2. Supervised live run on the box against real state.
3. Flip the unit's `ExecStart` (or swap the `music-beet` shim). The shell
   script stays in place and in the deploy manifest until the successor
   survives one full timer cycle on the box; the following task deletes it.
   Rollback inside that window is flipping `ExecStart` back; after the
   window it is redeploy-from-git, and the spec says so plainly.
4. Last flip includes `musicbox-ping-failure@.service`'s `ExecStart`. The
   closing task retires `lib.sh`, drops msmtp (after notify parity), yq, and
   jq from cloud-init, and updates the bring-up runbook's package gate and
   STATUS's pinned-versions table in the same commit, so a fresh provision
   mid-migration never builds a box the binary can't run on.

## Relationship to later work

The upload-check pass consumes `internal/reject` for its wording and builds
its own stream-level validation (a different mechanism at a different depth;
the gate here keeps full-decode integrity as the deep check). Anything shared
beyond this repo later means promotion out of `internal/` or a repo split —
a packaging-initiative decision, not baked in now. The packaging initiative
inherits: the site-bound inventory outside the binary, the non-secret config
file overlay, a provider seam for provisioning, scripted Cloudflare wiring,
and `deploy.sh`'s secrets-source override.

## Token and sequencing notes for the plan

Plan authored via writing-plans after Geoff reviews this spec; execution as
the next pass after the refinement fix wave closes, one implementer chain per
task, go-conventions invoked before any Go is written. Expected shape is ten
to twelve tasks (scaffold and config with doctor, lock and ping, notify with
parity test, reject, archive and gate, review, beets, importer, backup,
checks, and two migration-flip tasks); the plan carries its own ceiling and
the bats-mapping table.
