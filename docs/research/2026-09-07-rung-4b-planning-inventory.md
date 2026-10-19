# Rung 4, Pass 4b Planning Inventory

Research report, 2026-09-07, prepared for pass 4b's plan authoring,
against HEAD `c98a9166b38c6ec1c6bdc23ca280d1bb1c58526a`. Scope, quoted
from the rung 4 spec's pass-split table
(`docs/superpowers/specs/2026-09-05-rung-4-shadow.md:1632`):

> **4b** The deployment | `systemd/dubplate-shadow.{service,timer}`;
> `env/dubplate-server.env.template` with its literal rows and the
> existing template's header amendment; deploy.sh's six edits, the
> sixth being the two `-ldflags` stamps and the smoke-test assertion;
> `dubplate-server version` printing both stamps; the fresh-context
> unit read; the deploy sitting; the SELinux verifications; the
> recorded deploy date | `check.sh` green, `deploy.sh --check` clean, a
> non-`dev` matcher stamp asserted by the smoke test, the sitting
> completed, the first pass verified in the journal | Main tree

Method: every claim below is read off the committed tree at that HEAD,
off the rung 4 spec (revision 2, all six owner rulings ratified
2026-09-07), off `docs/HISTORY.md`, or off
`~/.claude/skills/vps-conventions/SKILL.md`. No box was contacted. No
git command mutated the index or the tree. Where the spec or the
handoff disagrees with the tree, the tree wins and the disagreement is
listed again at the end.

Ten numbered sections follow, each ending with what 4b inherits from
it. A surprises list closes the report.

---

## 1. `scripts/deploy.sh` as landed

443 lines. The header comment (`:4-23`) states the contract: rsync repo
config, render `/etc/dubplate/env` from the template plus the
workstation age store, cross-compile and ship the Go binary, install
every `dubplate-*` unit, bring the compose stack up, install beets via
uv. Idempotent by construction, with `--check` diffing and applying
nothing.

### The constants and the argument parse

`REPO_ROOT`, `ENV_TEMPLATE`, `REMOTE_DIR=/opt/dubplate`,
`REMOTE_ENV=/etc/dubplate/env`, `BEETS_VERSION=2.13.1` at `:25-29`. The
pinned uv installer with its recorded sha256 at `:31-38`.
`UNIT_FILES` at `:56-64`, twelve entries, each service beside its own
timer. `SSH_OPTS` at `:70-71` shares one ControlMaster connection
across the run. The argument loop at `:73-88` accepts `--check` and one
positional HOST, rejecting any other flag.

### The steps

Fourteen announced steps, one `echo "== ... =="` each.

| Line | Step | `--check` behavior |
| --- | --- | --- |
| `:199` | `compose.yaml` and `config/` to `/opt/dubplate` | `rsync -avzn --delete` |
| `:212` | `UNIT_FILES` to `/etc/systemd/system` | `rsync -avzn` |
| `:230` | `/etc/tmpfiles.d/dubplate.conf`, then `systemd-tmpfiles --create` | `deploy_file` diff, then `run_remote` prints |
| `:273` | cross-compile and rsync `/usr/local/bin/dubplate-server`, chmod 0755 | builds locally, `rsync -avzn` |
| `:288` | smoke test `dubplate-server version` | `run_remote` prints |
| `:297` | `/etc/dubplate/env` render and install | diff without showing content |
| `:327` | `/etc/dubplate/contributors.yaml`, `0640 root:music` | `deploy_file` diff |
| `:337` | `/srv/music/state/navidrome`, chown `2000:2000` | `run_remote` prints |
| `:358` | `/var/lib/musicbox`, chown `music:music` | `run_remote` prints |
| `:371` | `/srv/music/state/backup-staging`, chown `music:music` | `run_remote` prints |
| `:388` | `/srv/music/uploads`, chown `music:music` | `run_remote` prints |
| `:393` | `docker compose ... up -d` | `run_remote` prints |
| `:406` | `daemon-reload` and `enable --now $TIMER_UNITS` | `run_remote` prints |
| `:411` | beets 2.13.1 via uv as the `music` user | three literal "would" lines |

Three helpers carry the shape. `remote_reachable` (`:119-121`) probes
with `BatchMode=yes` so an unreachable box reports honestly instead of
hanging on a password prompt. `run_remote` (`:129-137`) prints
`would run: <command>` under `--check` and issues the ssh otherwise, so
a step's check line cannot drift from what apply runs. `deploy_file`
(`:147-169`) is the check-diff-install shape for one small config file,
staging through a box-side `mktemp` rather than a fixed `/tmp` name.

### The cross-compile and the smoke test, `:251-289`

The boundary comment at `:251-260` is the one that names this rung:
"Greenfield units start exec-ing /usr/local/bin/dubplate-server from
rung 4; the incumbent binary and its step retire together at rung 16."
The frozen `/usr/local/bin/dubplate` is never written by this script.

`make -C "$REPO_ROOT" build-linux` runs at `:274`, under `--check` too.
The comment at `:262-267` gives the reason: the build only writes the
gitignored local artifact, and running it is what lets the dry-run
rsync report a real diff instead of comparing against a stale copy.
`DUBPLATE_SERVER_BINARY` is `$REPO_ROOT/dubplate-server-linux-amd64`
(`:275`). Apply rsyncs it and chmods 0755 (`:279-280`).

The smoke test is two lines, `:288-289`: an echo and
`run_remote "/usr/local/bin/dubplate-server version"`. It asserts
nothing about the output. `set -euo pipefail` turns a non-zero exit
into a whole-deploy abort, which is the entire mechanism today. Under
`--check` it prints and does not run.

### The `--check` path is read-only against the box

Every branch was read. The three rsyncs use `-n`. `deploy_file` and the
env step read remote content with `ssh cat`. `run_remote` issues no ssh
at all. The one extra probe is the trigger-directory `test -d` at
`:239-249`, explicitly informational and non-fatal. The beets step
prints three literal sentences. So `--check` makes no box write, which
is what the owner grant for this window depends on.

Two things `--check` does that are worth stating in the plan. It writes
the local `dubplate-server-linux-amd64` through `make build-linux`
(`:274`). And it runs the env render loop first, unconditionally, so a
`--check` on a workstation missing any template value exits 1 at
`:190-194` before printing a single step. When no HOST is passed it
also needs `HCLOUD_TOKEN` and the `hcloud` binary (`:99-109`).

What `--check` prints per step: `rsync -avzn` output for the three
rsync steps, `no changes` or `would install <local> -> <remote>` for a
`deploy_file` step, `would run: <command>` for a `run_remote` step,
`no changes (contents match; diff not shown, it's secrets)` or
`would change (...)` for the env file, and `can't reach <target> ...`
in place of any diff when `remote_reachable` fails.

### The env render from the age store, `:171-194`

`source "$HOME/.local/secrets"` at `:93-96`, guarded by a file test.
The render loop at `:177-188` reads `$ENV_TEMPLATE` line by line, skips
blanks and `#` comments, takes `name="${line%%=*}"` (`:181`), resolves
`value="${!name:-}"` through indirect expansion in one flat shell
namespace (`:182`), and appends `NAME=$value` or records the name as
missing. Every miss is reported together at `:190-194` and the script
exits 1 before any remote step. The rendered content never touches
local disk: `printf | ssh install -D -m 600 -o root -g root /dev/stdin`
at `:314-315`.

The loop reads only the name. No tracked template line carries a
right-hand value today, verified by grep, so the spec's edit 5 is
backward compatible as stated.

### The `ReadWritePaths` pre-creation steps

Four directory steps, all in `run_remote "mkdir -p X && chown Y X"`
shape: `/srv/music/state/navidrome` at `:336-338` (chown `2000:2000`,
for the compose bind mount), `/var/lib/musicbox` at `:357-359`,
`/srv/music/state/backup-staging` at `:370-372`, and
`/srv/music/uploads` at `:387-389`. The comment at `:350-356` restates
the vps-conventions rule in the repo's own words: systemd requires a
`ReadWritePaths` path to exist on the host or the unit fails with
"Failed to set up mount namespacing".

### The unit install step, `:211-217`

One rsync of `"${UNIT_FILES[@]/#/$REPO_ROOT/systemd/}"` into
`/etc/systemd/system/`, no `--delete`. `daemon-reload` happens later,
at `:408`, paired with `systemctl enable --now $TIMER_UNITS`.
`TIMER_UNITS` at `:407` is a four-name string: import, backup,
disk-check, navidrome-check. The comment at `:396-405` records why
`dubplate-serve.service` and `dubplate-import-nudge.path` are
deliberately absent, and it is the precedent the spec's edit 2 leans
on.

### The six edits, located, and whether any exists

None of the six exists. `grep -n 'shadow\|server.env\|var/lib/dubplate\|ldflags'`
over `scripts/deploy.sh`, `Makefile`, `env/`, and `systemd/` returns one
line, and it is `scripts/deploy.sh:341`, the comment explaining why the
incumbent state directory is still called `/var/lib/musicbox`.

| Spec edit | Spec line | Site in the file today | Exists? |
| --- | --- | --- | --- |
| 1. `UNIT_FILES` gains the two shadow units | `:1307` | `deploy.sh:56-64` | No |
| 2. `TIMER_UNITS` unchanged | `:1311` | `deploy.sh:407`, precedent at `:49-55` and `:396-405` | No change owed |
| 3. Pre-create `/var/lib/dubplate` | `:1318` | new step in the shape of `deploy.sh:357-359` | No |
| 4. Parameterized render-and-install helper | `:1323` | lifts `deploy.sh:175-194` and `:291-316`, both written against `REMOTE_ENV` at `:28` | No |
| 5. Literal column in the renderer | `:1329` | `deploy.sh:181-182`, three lines; failure path `:190-194` | No |
| 6. The two `-ldflags` stamps and the smoke assertion | `:1349` | `Makefile:6-7` carries no flags; assertion joins `deploy.sh:288-289` | No |

Edit 6 has a second half the spec states and the plan should carry into
a task: `deploy.sh` computes the two values (`git describe --tags
--always --dirty` and `git rev-parse HEAD`) and passes them through to
`make build-linux`, which means `Makefile:6-7` must consume a variable
it does not have today. That is a Makefile edit the pass-split row does
not name.

**What 4b inherits, concretely.** A 443-line script with fourteen
steps, a `--check` mode that is read-only against the box, a render
loop that resolves 42 names through one flat namespace and aborts on
the first missing set, four pre-create steps to copy the shape of, and
an unasserted smoke test at `:288-289`. Five real edits, one no-op
edit, one new template, and one Makefile change nobody has counted.

---

## 2. The systemd estate

Twelve units in `systemd/`, 413 lines total.

| Unit | Lines | Type | Hardening | `ReadWritePaths` |
| --- | --- | --- | --- | --- |
| `dubplate-serve.service` | 128 | simple | the full extended set, `:99-125` | `/var/lib/musicbox /srv/music/uploads /srv/music/inbox /run/dubplate/nudge-triggers` |
| `dubplate-import.service` | 55 | oneshot | baseline, `:16-18` | `/srv/music /var/lib/musicbox` |
| `dubplate-import-nudge.service` | 41 | oneshot | baseline, `:23-25` | `/run/dubplate/nudge-triggers /srv/music /var/lib/musicbox` |
| `dubplate-backup.service` | 41 | oneshot | baseline, `:20-22` | `/var/lib/musicbox /srv/music/state/backup-staging` |
| `dubplate-disk-check.service` | 24 | oneshot | baseline, `:18-20` | none, with a stated reason |
| `dubplate-navidrome-check.service` | 24 | oneshot | baseline, `:18-20` | none, with a stated reason |
| `dubplate-ping-fail@.service` | 20 | oneshot | baseline, `:12-14` | none, with a stated reason |
| `dubplate-import-nudge.path` | 20 | path | n/a | n/a |
| four timers | 15, 16, 15, 14 | timer | n/a | n/a |

The oneshot baseline is exactly three directives: `NoNewPrivileges=yes`,
`PrivateTmp=yes`, `ProtectSystem=strict`. Every unit that omits
`ReadWritePaths` says in a comment why it writes nothing.

### `dubplate-serve.service` as the hardening template

The spec names `:96-125` as the set the shadow copies, and those lines
are what is there: the `--- hardening ---` banner at `:96-98`, then
`NoNewPrivileges`, `PrivateTmp`, `ProtectSystem=strict`, `ProtectHome`,
`ProtectKernelTunables`, `ProtectKernelModules`, `ProtectControlGroups`,
`RestrictNamespaces`, `RestrictRealtime`, `RestrictSUIDSGID`,
`LockPersonality`, `MemoryDenyWriteExecute`,
`SystemCallFilter=@system-service`, `SystemCallArchitectures=native`,
empty `AmbientCapabilities` and `CapabilityBoundingSet`, `UMask=0027`,
and `RestrictAddressFamilies=AF_INET AF_INET6 AF_UNIX AF_NETLINK` at
`:125`.

The netlink comment at `:116-124` is the citation the shadow's own
exclusion rests on. It records that netlink joined this unit's set for
the beets subprocess, that glibc's `getaddrinfo` opens a
`NETLINK_ROUTE` socket, and that "The Go process itself never needed
it, since Go's own resolver has its own fallback."

The fifteen-line network-online comment is at `:12-26`, with
`After=docker.service` at `:27`. It argues that serve depends on the
compose bridge network rather than on outbound reachability, that boot
resilience comes from `IP_FREEBIND` set in code, and that no
`IPFreeBind=` directive exists. It is the estate's proof that the
network-online pairing is a decision rather than a default, which is
the argument the shadow unit's own comment has to make in the opposite
direction.

`StartLimitIntervalSec=300` and `StartLimitBurst=30` at `:45-46` carry
a conductor ruling: an interval of 0 means the unit never enters
`failed`, so `OnFailure=` never fires. That reasoning applies to any
future unit that reaches for a restart policy. A `Type=oneshot` shadow
unit has no `Restart=`, so it does not apply to 4b directly.

### The sibling oneshots' `After`/`Wants` pairing

`dubplate-import.service:9-10` is
`After=network-online.target docker.service` with
`Wants=network-online.target`. `dubplate-backup.service:13-14` is the
same pair. `dubplate-navidrome-check.service:11-12` and
`dubplate-import-nudge.service:9-10` carry it too;
`dubplate-disk-check.service:11-12` carries it without
`docker.service`. So the estate has five instances of the pairing, and
the disk check is the precedent for taking `network-online` without
`docker.service`, which is the shape a shadow unit wants: it reaches
MusicBrainz and touches no container.

Every one of these units also carries `EnvironmentFile=/etc/dubplate/env`,
`Type=oneshot`, and `User=music`. In `dubplate-import.service` that
line is `:15`, not `:14` as the spec cites.

### The `OnFailure=` slug and how its URL is set

Five call sites, one per job: `dubplate-ping-fail@import.service`
(`import.service:8`, and again at `import-nudge.service:8`),
`@backup.service` (`backup.service:12`), `@disk.service`
(`disk-check.service:10`), `@navidrome.service`
(`navidrome-check.service:10`), and `@serve.service`
(`serve.service:35`). The template unit resolves the slug itself:
`ExecStart=/usr/local/bin/dubplate ping-fail %i` at
`ping-fail@.service:20`, reading `EnvironmentFile=/etc/dubplate/env` at
`:11`, with a comment at `:15-19` recording that `ping-fail` resolves
`DUBPLATE_PING_<SLUG>_URL` straight from the process environment,
deliberately bypassing `config.Load`, and writes nothing to disk.

`serve.service:28-34` is the precedent the spec cites for wiring a slug
before its URL is minted: it says in as many words that
`DUBPLATE_PING_SERVE_URL` is unset today and that an unset URL "no-ops
cleanly rather than erroring, so wiring this ahead of that mint is
safe."

Two facts follow for a `dubplate-ping-fail@shadow.service` line. The
helper execs the frozen `/usr/local/bin/dubplate`, not
`dubplate-server`, and it reads `/etc/dubplate/env`, the incumbent's
file, not the greenfield one. So a `DUBPLATE_PING_SHADOW_URL` would
have to land in the incumbent template if it were ever minted, and
nothing in the greenfield template could reach it.

### The timer patterns in use

Four timers, one shape. `[Unit]` with a `Description` only, then
`[Timer]` with `OnCalendar`, `RandomizedDelaySec`, `Persistent=true`,
and an explicit `Unit=`, then `[Install] WantedBy=timers.target`. Every
one carries a comment saying which retired `music-*` timer it mirrors
and why its jitter is sized as it is.

`OnCalendar=*:0/15` with `RandomizedDelaySec=300` for import
(`import.timer:10-11`) and disk-check (`disk-check.timer:9-10`).
`OnCalendar=*-*-* 11:30:00 UTC` with `RandomizedDelaySec=600` for
backup (`backup.timer:9-10`). `OnCalendar=*:0/5` with
`RandomizedDelaySec=60` for navidrome-check
(`navidrome-check.timer:8-9`), which is exactly the schedule and jitter
the spec assigns the shadow timer, comment included: "a shorter jitter
window than the disk check's, proportional to this timer's own
5-minute interval."

Every timer sets `Persistent=true` already, so the spec's M3 fold
(`Persistent` true against revision 1) matches what every sibling does
rather than departing from it.

### What `scripts/check.sh`'s systemd leg accepts and filters

`scripts/check.sh:125-148`. The glob at `:137` is
`(systemd/*.service systemd/*.timer systemd/*.path)`, so two new files
join the gate with no wiring. `:138` runs
`systemd-analyze verify --recursive-errors=no` over the set, swallowing
its exit status. `:139` filters the output with
`grep -v -E ': Command /usr/local/bin/[A-Za-z0-9_.-]+ is not executable: No such file or directory$'`,
which matches `dubplate-server` as well as `dubplate`. Any surviving
line fails the gate. The comment at `:126-136` states that
`--recursive-errors=no` alone does not suppress the missing-executable
complaint, so the filter is the actual mechanism. The leg skips
entirely when `systemd-analyze` is absent (`:147`).

The gate therefore proves the two new units parse and that no
directive is misspelled. It proves nothing about whether the
`ReadWritePaths` grant is minimal, which is precisely the gap the
fresh-context unit read exists to close (`SKILL.md:76-80`, spec
`:1808-1818`).

**What 4b inherits, concretely.** A hardening template at
`serve.service:99-125` to copy minus `AF_NETLINK`, five instances of
the network-online pairing to follow and one fifteen-line comment
proving the pairing is argued rather than assumed, a five-slug
`OnFailure=` convention whose helper reads the incumbent env file, a
four-timer pattern the shadow timer matches exactly at
`navidrome-check.timer`, and a gate leg that picks up new files by glob
and checks syntax only.

---

## 3. The state directory and the store root

### The store layout as landed

`internal/engine/shadow/store.go:26-41`. The root is
`StoreDir(stateDir)`, which is `filepath.Join(stateDir, "shadow")`
(`:41-47`); the constant carries the comment "so 4b's ReadWritePaths is
a single static path". `OpenStore` creates root and four
subdirectories, 0755, files 0644 (`:50-59`):

- `observations/<obs-id>/observation.json`, plus `case.json` and
  `responses/%03d.json` under the same directory (`:108-207`)
- `passes/<pass-id>.json` (`:303`)
- `pending/<obs-id>.json` (`:221-228`)
- `resolved/<obs-id>.json` (`:236-267`)
- `started-at`, written once and never rewritten (`:89-106`)
- `consumed-reports.json`

`validID` (`:64-71`) refuses an empty, `.`, `..`, or separator-bearing
id, so nothing can escape root through an id.

`TestStoreWritesNothingOutsideItsRoot`
(`internal/engine/shadow/store_test.go:191-249`) exercises `StartClock`,
`PutObservation`, `PutResponse`, `PutCaseFile`, `PutPending`,
`PutResolved`, `PutPass`, and `PutConsumedReports`, then rejects a
`../../escaped` id, then walks the whole temp directory and fails on
any file whose path is not under root. It is the proof HISTORY cites,
and it holds.

### The config key and its default

`DUBPLATE_STATE_DIR` is an owned wall row (`internal/config/keys.go:21`)
resolved by `w.require("DUBPLATE_STATE_DIR")`
(`internal/config/config.go:217`). It has **no default**: `require`
records `ErrMissing` when it is unset (`config.go:125-131`). So the
value comes entirely from the environment file, which is what makes the
spec's literal row the whole mechanism.

### Therefore the exact `ReadWritePaths` argument

The handoff says the store root "is 4b's `ReadWritePaths=` argument"
(`docs/HISTORY.md:389-394`), which reads as
`/var/lib/dubplate/shadow`. The spec's deploy edit 3 says the pre-create
step is for `/var/lib/dubplate` (`:1318`). Those two cannot both be
followed as written: `SKILL.md:39-42` requires every
`ReadWritePaths` entry to exist on the host before the unit starts, so
a grant on the leaf needs the leaf pre-created, and `OpenStore`'s own
`os.MkdirAll(root, ...)` would then be a no-op rather than the thing
that creates it.

Two consistent options, and the plan owes a decision:

- Grant the leaf `/var/lib/dubplate/shadow` and pre-create both levels
  in the deploy step. Narrowest grant. The unit can never write beside
  its own store.
- Grant the parent `/var/lib/dubplate` and pre-create the parent only,
  letting `OpenStore` create `shadow/` on the first pass. One
  pre-create line, one wider grant.

The narrow one is the better reading of `SKILL.md:76-80`, which asks
for a write set argued statically rather than assumed, and the
`dubplate-backup.service:32-39` comment is the estate's own precedent
for taking a deliberately leaf-only grant and adjusting the code's
delete strategy to fit inside it.

### What else on the box the unit must read

`shadow.Config` (`internal/engine/shadow/config.go:24-33`) carries four
paths, and `run.go` uses each exactly once:

| Field | Read at | Estate path | Config row |
| --- | --- | --- | --- |
| `StateDir` | `run.go:95`, `OpenStore(StoreDir(cfg.StateDir))` | `/var/lib/dubplate` | `DUBPLATE_STATE_DIR`, owned, literal per the spec |
| `ObserveDir` | `run.go:138`, `capture(...)` | `/srv/music/inbox` | `DUBPLATE_SHADOW_OBSERVE_DIR`, required-when-enabled |
| `OracleStateDir` | `run.go:190`, `readRunReports`; `run.go:194`, `readTagLogSkips` | `/var/lib/musicbox` | `DUBPLATE_SHADOW_ORACLE_STATE_DIR`, required-when-enabled |
| `ReviewDir` | `run.go:198`, `reviewedBasenames` | `/srv/music/review` | `DUBPLATE_SHADOW_REVIEW_DIR`, required-when-enabled |

`readRunReports` reads every `report-*.json` directly under
`OracleStateDir`, and `readTagLogSkips` every `import-*.log`
(`oraclereport.go:79-142`). `reviewedBasenames` walks
`<review>/<date>/<uploader>/<names>` (`reconcile.go:184-207`), the
layout `docs/STATUS.md:52` records live on the box as
`/srv/music/review/2026-08-31/geoff/`.

Two absences matter. `cfg.LibraryRoot` is never read by
`internal/engine/shadow`, verified by grep; `DUBPLATE_LIBRARY_ROOT`
appears in the template only because `config.Load` requires it
(`config.go:216`). And the Navidrome library and the beets database are
read never, which is why the read-only guarantee is a property of the
unit's write set rather than of any code path.

All four reads are ordinary reads under `ProtectSystem=strict`, which
leaves the filesystem readable and only removes write. `ProtectHome=yes`
hides `/home` and `/root` and touches neither `/srv` nor `/var/lib`.

**What 4b inherits, concretely.** A store rooted at
`<state dir>/shadow` with six named children, a passing test that
proves nothing is written outside it, a `DUBPLATE_STATE_DIR` with no
default at all, three read-only estate paths each with its own config
row, and one unresolved question about whether the grant and the
pre-create name the parent or the leaf.

---

## 4. The environment file

### `env/dubplate.env.template` as landed

141 lines, 42 assignment rows, of which 39 are `DUBPLATE_*` and three
are the literal names Navidrome reads (`ND_PASSWORDENCRYPTIONKEY`,
`ND_SPOTIFY_ID`, `ND_SPOTIFY_SECRET`, `:126-134`). The spec's count of
39 `DUBPLATE_*` names is right; the render loop resolves all 42.

The header runs `:1-22` in four blocks. `:1-3` is the sentence the spec
says rung 4 amends: "This manifest describes the frozen pre-pivot
binary's wall (`_salvage/internal/config`), not dubplate-server's. The
greenfield wall is internal/config's own package doc; this manifest
evolves at rung 4." `:5-9` states the render contract and the
fail-listing behavior. `:11-14` states that the file carries no real
values and points at the registry. `:16-22` states the pre-pivot
`Load` semantics.

Every row is bare, `NAME=`. Grep for a row with a right-hand value
returns nothing, which is the fact edit 5's backward compatibility
rests on.

The rows group under seven comment headings: admin contact (`:24-25`),
SMTP (`:27-33`), R2 (`:35-41`), Navidrome (`:43-49`), fixed config
paths (`:51-57`), working directories (`:59-77`), the trigger directory
(`:79-84`), the uploads store (`:86-89`), serve (`:91-105`), ping URLs
(`:107-116`), and the literal container names (`:118-141`). The
working-directories comment at `:59-70` is where the estate records why
`DUBPLATE_STATE_DIR` is `/var/lib/musicbox`, and `:74` is the row
itself.

### How deploy.sh renders it, and what "secrets MANIFEST" means

Covered in section 1. The rule CLAUDE.md states, "`env/dubplate.env.template`
is a secrets MANIFEST rendered by deploy.sh from the age store, inline
values in it are not defaults and adding a name there requires minting
it," is what the spec's edit 5 changes for the greenfield template
only: a right-hand value becomes a third row state, the literal, beside
the bare `NAME=` and the commented pending row. Both template headers
have to name all three states, and the spec is explicit that the plan
states the no-mint conclusion out loud rather than leaving CLAUDE.md's
rule to be read as requiring a mint (`:1284-1291`).

**Secrets: none.** No greenfield row is a secret, `secret-set.sh` is
not run, and `~/.dotfiles/secrets/registry.md` gains no row.

### The six `DUBPLATE_SHADOW_*` rows and their `Load` semantics

`internal/config/keys.go:214-229` registers one integration class:

```
Name: "shadow", Rung: 4, Flag: "DUBPLATE_SHADOW_ENABLED"
Required: OBSERVE_DIR, ORACLE_STATE_DIR, REVIEW_DIR
Optional: FLOOR, BUDGET
```

The semantics, read off `config.go` and confirmed by
`config_test.go:280-315`:

- The flag is off when unset or empty (`config.go:155-159`). A value
  `strconv.ParseBool` refuses is an error, never a silent off
  (`:160-164`).
- With the flag true, all three `Required` rows are required and each
  missing one is a distinct `ErrMissing` (`config.go:231-234`).
  `cfg.Shadow` is populated (`:235-241`).
- With the flag false or absent, `cfg.Shadow` stays nil and
  `shadow observe` refuses naming the flag
  (`cmd/dubplate-server/shadow.go:200-202`).
- **An explicit `DUBPLATE_SHADOW_ENABLED=false` loads clean**, which is
  the adopter's supported way to turn the integration off
  (`config_test.go:307-310`).
- **A `Required` or `Optional` row set beside a false or absent flag is
  a load error naming the flag** (`config.go:190-192`;
  `config_test.go:311-315` asserts both key names appear). A greenfield
  template that carried the paths with the flag off would abort every
  load.
- The two `Optional` rows are registered thresholds, accepted and never
  required, with defaults that stand when unset
  (`thresholds.go:55-96`).

So the template's flag row is load-bearing in a way the other rows are
not: it must be a literal `true` for the deployed unit to work at all,
and the six rows only ever ship together.

### Is the spec asking for a separate template or an amendment?

Both, and they are separate deliverables. A **new file**,
`env/dubplate-server.env.template`, rendered to `/etc/dubplate/server.env`
(spec `:1257-1262`). An **amendment** to the existing
`env/dubplate.env.template` header naming its sibling (spec `:1262-1265`),
which is one edit to `:1-3`. The two files coexist until rung 16.

The reason the greenfield binary cannot read the incumbent's file is
mechanical and worth restating in the plan: of the 39 `DUBPLATE_*` rows
in `/etc/dubplate/env`, two are owned rows the greenfield wall accepts,
and every other one is a reserved, superseded, or unknown name that
`rejectNamespace` turns into a load error (`config.go:173-204`). Two
rows the greenfield binary requires, `DUBPLATE_LIBRARY_ROOT` and
`DUBPLATE_SITE_NAME`, are absent from that file entirely, having been
renamed (`keys.go:165-166`).

The new template therefore carries ten rows: four owned
(`DUBPLATE_LIBRARY_ROOT`, `DUBPLATE_STATE_DIR`, `DUBPLATE_SITE_NAME`,
`DUBPLATE_ADMIN_EMAIL`) and six shadow rows.

### `DUBPLATE_SITE_NAME`'s template-literal status

`DUBPLATE_SITE_NAME` is the renamed successor to
`DUBPLATE_SERVE_SITE_NAME` (`keys.go:166`), and the age store carries
the old name. The spec rules it a **literal** carrying the estate's
masthead text (`:1279-1283`), on two grounds: a bare `NAME=` would
abort the render, and a rename in the age store would make one string
live in two places. The spec deletes revision 1's claim that the value
"stays minted in the age store from the wind-down" rather than
repairing it, on the strength of a name-only check of the store.

`DUBPLATE_LIBRARY_ROOT` is a literal for the same rename reason. Only
`DUBPLATE_ADMIN_EMAIL` is a store lookup, unchanged and already minted
(`env/dubplate.env.template:25`).

**What 4b inherits, concretely.** A 141-line manifest whose header is
already written to expect this rung, a render loop that reads names and
not values, an integration class whose three path rows are
required-when-enabled and whose flag row must literally read true, a
proven explicit-false path for adopters, a new ten-row template to
author, a one-sentence header amendment, and a no-mint conclusion the
plan states rather than assumes.

---

## 5. The stamps

### The two variables

`cmd/dubplate-server/main.go:29-44`.

```go
var version = "dev"        // main.go:31
var matcherCommit = "dev"  // main.go:44
```

`version`'s comment (`:29-30`) says it is set via
`-ldflags "-X main.version=..."` at release build time and stays `dev`
for workstation builds. `matcherCommit`'s comment (`:33-43`) is longer
and names this pass: it records that `shadow observe` refuses `dev` or
empty because an unstamped divergence cannot be mechanically
re-evaluated, and that "Pass 4b lands the build path that sets it: today
no build path does, so this value is `dev` everywhere and the refusal
fires everywhere."

The exact `-ldflags` paths are therefore `-X main.version=<value>` and
`-X main.matcherCommit=<value>`, both against package `main` in
`./cmd/dubplate-server`.

### The version verb as landed

`newVersionCmd` at `main.go:72-81`. Its `RunE` is one `fmt.Fprintln` of
`version` and nothing else. One line, one value. 4b makes it print both
stamps, which is a change to that `RunE` and to whatever test asserts
its output in `main_test.go`.

The spec cites `cmd/dubplate-server/main.go:56-64` for this. At HEAD
`:56-64` is inside `newRootCmd`'s command construction. The citation is
stale; the code it describes is at `:72-81`.

### Where else the stamps are read

`version` is not decorative. `cmd/dubplate-server/shadow.go:215` builds
the MusicBrainz User-Agent as
`fmt.Sprintf("dubplate-server/%s ( %s )", version, cfg.AdminEmail)`, so
an unstamped build identifies itself to MusicBrainz as
`dubplate-server/dev`. Stamping `version` changes what the box sends on
every request, which is a fact the deploy sitting can observe and the
plan should mention.

`matcherCommit` is passed as a parameter rather than read as a global:
`newRootCmd` hands `main.matcherCommit` to `newShadowCmd`
(`main.go:67`), which hands it to both `observe` and `reevaluate`
(`shadow.go:173-177`). The comment at `shadow.go:159-164` gives the
reason, that a test which cannot vary the stamp cannot prove the
refusal.

### Doctor's rows

`newDoctorCmd` (`doctor.go:24-72`) prints four groups in `"%-32s %s"`
form, every one derived from `internal/config`'s own tables so no rung
hard-codes a count: the owned wall rows (`:34-39`), the reserved
classes as `disabled (arrives at rung N)` (`:41-46`), the integration
classes as `enabled` or `disabled` (`:48-56`), and every registered
threshold with its effective value (`:58-62`). Exit status follows
`config.Load`'s verdict.

**Doctor prints no version row and no matcher-commit row.** Neither
stamp appears anywhere in `doctor.go`. If 4b wants the deploy sitting
to read the stamps from `doctor` as well as from `version`, that is a
new deliverable and the pass-split row does not name it. The
conservative reading is that `version` is the only surface owed.

`thresholdStatus` (`doctor.go:111-128`) already switches on
`config.KindDuration` and falls through to `ParseInt`. The spec's claim
that a duration "renders as `invalid:` in the doctor table" describes
the pre-4a code; pass 4a landed `Threshold.Kind`, `config.ParseDuration`
(`thresholds.go:10-18`, `:74-96`), and this switch. Nothing is owed
here.

### The refusal

`observe`: `shadow.go:203-205`, after the config load and the nil-Shadow
check, before the thresholds and before `OpenStore`, so no store is
created on a refusal. `reevaluate`: `shadow.go:303-305`, same test, same
position. Both return
`refusing to run on an unstamped matcher commit %q`. `report` carries no
such check, because reading a report needs no stamp.

### What the Makefile and deploy.sh set today

`Makefile:6-7`:

```
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BINARY)-linux-amd64 ./cmd/dubplate-server
```

No `-ldflags`, no variable to override, and `build` at `:3-4` and
`install` at `:18-19` carry none either. `deploy.sh:274` invokes
`make -C "$REPO_ROOT" build-linux` with no arguments. So edit 6 touches
three files: `Makefile` gains a flags variable the recipe consumes,
`deploy.sh` computes the two values and passes them, and `main.go`
prints both.

**What 4b inherits, concretely.** Two package-level `main` variables
already declared and already commented with this pass's name, a version
verb printing one of them at `main.go:72-81`, a User-Agent that
silently carries the unstamped value, two refusals that fire today on
every build, a doctor table that mentions neither stamp, a threshold
switch that is already done, and a Makefile with no flag plumbing at
all.

---

## 6. The four provisional marks

All four are comment text. None is a code path, none is referenced by a
test, and retiring one is an edit to prose that the vale leg of the
gate reads.

**`internal/engine/shadow/rescore.go:28`**, in `rescoreScoring`'s doc
comment:

> returns the tunables and compat pair the re-score scores under:
> `DefaultCompat` applied to the estate's own tunables (ruling 3,
> provisional).

Ruling 3 is "which `matcher.Compat` the shadow runs" (spec `:192`),
ratified as recommended. **Retire the parenthetical's `provisional`**
and keep the ruling citation, since the citation is what tells a reader
where the choice was made.

**`internal/engine/shadow/match.go:20`**, in `shadowScoring`'s doc
comment:

> returns the tunables and compat pair `matchAlbum` scores under:
> `matcher.QuietImportTunables()` and `matcher.FullCompat()` (ruling 3,
> provisional).

Same ruling, same treatment.

**`internal/engine/shadow/run.go:388`**, in `disputedRelease`'s doc
comment:

> returns the release under dispute and the side that proposed it
> (ruling 2, provisional): the oracle's release wherever the oracle
> applied, and the Go matcher's own wherever it did not.

Ruling 2 is "how a skipped album's `Subject.ReleaseMBID` is filled"
(spec `:164`), ratified. Same treatment.

**`internal/engine/shadow/flac.go:62`** is different and **stays**:

> It is a provisional reader. Rung 8 owns tag I/O and the full field
> surface in `internal/engine/tagio` (`internal/engine/tag/doc.go:8-11`);
> rung 4 builds this because waiting on that lane would spend calendar
> on the one rung whose calendar IS the launch date, and rung 8's plan
> carries the task of absorbing it.

This is a rung 8 handoff mark, not a ruling mark. It names no ruling
number, its word "provisional" describes the reader's lifetime rather
than a pending decision, and `docs/HISTORY.md:437-441` carries the
absorption as forward carry 5. Deleting it would delete the pointer
rung 8 needs. **Leave it exactly as written.**

Two more provisional marks exist outside `internal/engine/shadow`, and
the task's count of four does not include them.
`cmd/dubplate-server/shadow.go:22` reads "Provisional, owner ruling 4
open" on `shadowFloorThreshold`, and `:37` reads "Provisional, owner
ruling 5 open" on `shadowBudgetThreshold`. Rulings 4 (the estate's own
shadow floor, spec `:261`) and 5 (the MusicBrainz budget cap, spec
`:278`) are both ratified. `docs/STATUS.md:47-48` counts six, not four:
"4a-i code untouched by 4a-ii still carries 'provisional, owner ruling
N open' marks; all six are ratified, retiring them is mechanical." The
mechanical task should take all five that a ruling retires and leave
`flac.go:62`.

**What 4b inherits, concretely.** Five comment edits across four files,
each retiring a word and keeping its ruling citation, and one comment
that must survive untouched because rung 8 reads it. `docs/STATUS.md`
already says six and the task brief says four; the tree says five plus
one.

---

## 7. The fresh-context unit read and the SELinux verifications

### The fresh-context unit read

Spec `:1808-1818`, in the Process section. One agent, zero context,
reads `dubplate-shadow.service`, `dubplate-shadow.timer`, and every
`deploy.sh` diff against `~/.claude/skills/vps-conventions/SKILL.md`
and against `dubplate-serve.service` as the hardening template, and
reports findings. The stated reason is that `systemd-analyze verify` in
the gate proves the units parse and proves nothing about whether the
`ReadWritePaths` grant is the minimum, and that `SKILL.md`'s Gate rule
is that a unit's write targets are argued statically rather than
assumed (`SKILL.md:76-80`).

It is the second of only two review exceptions rung 4 takes; rung 4 is
not one of the three full-adversarial-cycle rungs (`:1800-1806`). It
runs before any box write, so it belongs before the sitting boundary and
runs unattended.

### The quiet-box procedure, spec `:1365-1391`

The spec states plainly that no mechanism in the repo checks quiet-box
discipline. The rule is doctrine plus rung assignment, and the closest
thing to a mechanical check is `--check` mode plus the
`remote_reachable` probe. The procedure at the sitting, in order:

1. Confirm no import or backup run is in flight and none is due inside
   the window. Import fires `*:0/15` with 300s jitter
   (`import.timer:10-11`); backup fires daily at 11:30 UTC with 600s
   jitter (`backup.timer:9-10`). Disk-check shares import's schedule and
   navidrome-check fires `*:0/5`, both harmless.
2. Run `deploy.sh --check` and read every step's diff.
3. Apply.
4. Enable the timer by hand.
5. Verify the first pass in the journal.
6. Verify the SELinux context on `/var/lib/dubplate` and on
   `/var/lib/musicbox`.
7. Record the root filesystem's free space as the store's sizing
   baseline, and whether `dubplate-disk-check` watches it.
8. Record the deploy date.

Read-only ssh at every step until the apply. One operator rule holds
after the sitting: a hand-run shadow verb goes through
`systemctl start dubplate-shadow.service` or `systemd-run` carrying the
unit's own properties, never the bare binary as root, because the
read-only guarantee is the unit's and not the binary's
(`:1387-1391`, restated at `docs/HISTORY.md:415-419`).

### The SELinux verifications

`SKILL.md:43-45`: new directories under `/var/lib` on an
SELinux-enforcing box inherit `var_lib_t`, fine for unconfined
services, "but verify once on the first real deploy rather than
assuming."

The box is enforcing, and that is measured rather than assumed.
`docs/HISTORY.md:3446-3449` records the cloud-init fix that flipped the
Hetzner `alma-10` image from `SELINUX=permissive` to enforcing with a
`/.autorelabel`, and `:3478` records the acceptance run:
`getenforce` returns `Enforcing`, PASS.

`docs/HISTORY.md:3082-3083` carries the open item: "SELinux one-time
check on `/var/lib/musicbox` at first deploy," carried out of the
rewrite pass and never closed. So the sitting closes two checks, the
new `/var/lib/dubplate` and the inherited `/var/lib/musicbox`, which is
what spec `:1240-1245` says.

`grep -n -i 'deploy sitting' docs/HISTORY.md docs/STATUS.md` returns
nothing. The sitting has no prior record of its own; the spec's ruling 6
(`:345-372`) is the whole of it.

### The sitting boundary

Everything in this pass except the sitting runs unattended.

**Before the boundary, unattended.** The two unit files, the new
environment template, the existing template's header amendment, all six
deploy edits, the Makefile flag plumbing, `version` printing both
stamps, the provisional-mark retirement, `bash scripts/check.sh` green
at every task, the fresh-context unit read, and a read-only
`deploy.sh --check` against the box under the owner grant for this
window.

**At the sitting, attended.** The quiet-box confirmation, the `--check`
read-through together, the apply, `systemctl enable --now
dubplate-shadow.timer`, the journal check of the first pass, both
SELinux verifications, the free-space baseline, and the recorded deploy
date. The pass stays open across the boundary; ruling 6 makes the
sitting a later attended session.

`docs/HISTORY.md:421-427` adds the reason the date is a first-class
figure: `started-at` is written on the first pass that completes all
three steps without a fatal error, including a pass that observes
nothing, so the clock starts at the deploy, and the launch date is that
date plus the 720h floor.

**What 4b inherits, concretely.** A named fresh-context review with its
two comparison texts, an eight-step human procedure with no mechanical
check behind it, a box already proven enforcing, one SELinux check
carried open since the rewrite pass, and a clean split between the
tasks that finish unattended and the six things only the sitting can
do.

---

## 8. The 2.0 forward hooks

`docs/superpowers/specs/2026-09-07-streamer-2-0-design.md:354-388`
carries fourteen hooks. Read in full, plus a grep of the whole spec for
`systemd`, `deploy`, `/var/lib`, `state directory`, `ldflags`, `stamp`,
and `environment`.

**No hook names deploy, systemd, or a build stamp.** Two mention
adjacent nouns and neither reaches 4b:

- The spec's body at `:325` says "The streamer's state lives under the
  existing state directory, so backup covers it," and continues "Doctor
  gains rows." That is a rung 6 and rung 12 consequence of the state
  directory's existence, not a claim on its path or its grant.
- `:240` mentions an environment file in the context of an app-password
  key minted through the age store, which is rung 6's schema work.

Hook by hook, with 4b's relationship:

| Hook | Rung the hook names | Does 4b touch it? |
| --- | --- | --- |
| 1. Content hash on the catalog row | 6 schema, 7 gate | No |
| 2. Folded sort key | 6 schema, 9 tagging | No |
| 3. Art blob at admit | 6, 9 | No |
| 4. Play history as timestamped events | 6 schema | No |
| 5. A stable identity | 5, 6 | No |
| 6. The driver seam stays universal | 12 | No |
| 7. Schema migrations | 6 | No |
| 8. Collection ownership | 6 | No |
| 9. A transaction model across catalog events | 6 | No |
| 10. Live-file backup consistency | 6 | No |
| 11. ReplayGain and R128 gain | 6 schema, 9 | No |
| 12. A portable layout | 7 | No |
| 13. Sort tags written at admit | 8, 9 | No |
| 14. An ID3 version ruling | 8 | No |

So every 4b task is **neutral** on the forward-hooks list, and the plan
states that per task under CLAUDE.md's build-for-2.0 rule. This matches
`docs/HISTORY.md:460-462`, forward carry 9: the hooks bind rungs 5
through 9 and rung 12.

The prior-art rule applies the same way. The spec's prior-art table
(`:391-421`) holds twenty-three entries covering the OpenSubsonic
contract, oracles, tag and audio libraries, SQLite drivers, image and
rate-limit libraries, and migration mechanisms. **None is applicable to
a deploy pass**, which is what a deploy pass's plan says per entry
rather than skipping the section. One entry is worth naming positively
because 4b's own code already depends on it: `internal/atomicfile`,
whose verdict is "Extend it with a streaming write rather than
inventing a second temp-file convention." The shadow store writes
through it today (`store.go:14`, `:100`), and 4b neither extends nor
changes it.

**What 4b inherits, concretely.** A fourteen-hook list that touches no
infrastructure surface, so every task's hook statement is "neutral," and
a prior-art table whose every row is not applicable to a systemd,
environment, and deploy pass, declined for that reason and recorded.

---

## 9. Contested surfaces

### The parallel-approach spec's ownership table

`docs/superpowers/specs/2026-09-03-parallel-build-approach.md`, section
3's table. The row 4b lives on is `:184`:

> `systemd/`, `scripts/deploy.sh`, `env/` (conductor ruling, 2026-09-05)
> | rung 1's; rung 4 pass 4b adds two units, one template, and six
> deploy edits; rung 16 converges all of it | rungs 5, 7, 12, 16

Three other rows in that table matter to 4b:

- `:179` the state directory layout, "**rung 4**, first to deploy one,
  and its unit's `ReadWritePaths` depends on it," consumed by rung 5's
  nonces and per-email epoch, with rung 6's store superseding neither.
- `:183` `internal/config` (`keys.go`, `thresholds.go`), rung 1b's
  registry when it lands, "until then rung 4 as the only live code
  lane," with the rider to re-fingerprint `.gitleaksignore` on any
  insertion above its pinned line.
- `:182` `cmd/dubplate-server` (`main.go`'s `AddCommand` block,
  `doctor.go`'s rows), rung 1's skeleton, each rung adding its own file
  and registering through it. 4b edits `main.go`'s `newVersionCmd`
  rather than its `AddCommand` block, which is a narrower touch than
  the row anticipates and worth naming in the plan.
- `:181` `internal/engine/shadow`, rung 4, created by it, no other lane
  writes it. This is where the provisional-mark task lands.

The `docs/` ledgers row (`:177`) is conductor-owned: a pass drafts a
STATUS or HISTORY entry and never commits the file.

### The rung 4 spec's own rows

Spec `:1639-1661`, nine rows the conductor writes into the ownership
table before the lane launches, seven new and two amending. The
`systemd/`, `scripts/deploy.sh`, `env/` row and the state-directory row
are already in the parallel spec at `:184` and `:179`, so those two are
landed rather than owed.

The `.gitleaksignore` rider is live and cheap to trip. `.gitleaksignore`
is one line, `internal/config/keys_test.go:generic-api-key:66`, and line
66 of that file is the `DUBPLATE_R2_BUCKET` fixture row. Any insertion
above it moves the finding and reds the gate's gitleaks leg. 4b is not
expected to touch `keys_test.go` at all, since the six rows landed in
4a, so the rider is a guard rather than a task.

### Is any other lane live?

`docs/STATUS.md:31-36` is the lane registry, and it is the live-executor
check per `docs/STATUS.md:19-21` and CLAUDE.md.

```
- lane-1 (main tree): IDLE, pending 4b's plan.
- lane-2 (worktree): BLOCKED until Phase 0b/0c.
- doc lane: IDLE, one candidate queued (the 2.0 study rung).
```

All three idle or blocked, so 4b takes the main tree with no
contention. `docs/STATUS.md:6-13` confirms rung 4a is closed and both
its passes are in the tree and green. The plan's launch checklist still
runs, and the registry entry is the conductor's to write.

**What 4b inherits, concretely.** One ownership row that already names
this pass and its exact deliverable count, three adjacent rows it
touches (state directory, `internal/config`, `cmd/dubplate-server`), a
`.gitleaksignore` guard that costs nothing while `keys_test.go` stays
untouched, and an empty lane registry.

---

## 10. Budget evidence

### 4a-ii as measured

`docs/HISTORY.md:466-481`. About **2.7M against a 3M ceiling**, an
estimate from the agent meters, itemized:

| Item | Spend |
| --- | --- |
| staleness pre-task plus chain-args extraction | 0.26M |
| the chain (16 agents, five tasks) | 1.32M |
| Task 13 and Task 14 fix and review rounds | 0.49M |
| simplify plus its review | 0.18M |
| the close drafter | 0.16M |
| the conductor's own turns | about 0.2M |

The 2.4M flag was crossed at the close, after every task had landed,
and nothing was cut. The 2.0 brainstorm's reviews and fold, about 0.9M
in the same session, are not counted.

Attended interactions: **planning misses 2** (the exit gate's
single-uploader fixture figure and the mutation-6 narration, both of
which a probe at planning would have caught), **execution sittings 0**.

The derived rate: 1.32M chain across five tasks is about 0.26M per task
including its review, and the whole pass at 2.7M over five tasks is
about 0.54M per task all-in.

### A proposal for 4b

The spec proposed 1.5M (`:1662-1672`), to be re-derived from 4a's
actual report. Re-deriving it by task count:

The pass's plausible task list is seven. The two unit files as one
task; the new environment template plus the existing header amendment;
deploy edits 1 through 3 (manifest, timer note, pre-create); deploy
edits 4 and 5 (the render-and-install helper and the literal column,
which are one refactor of the same two blocks); deploy edit 6 with the
Makefile plumbing and the smoke assertion; `version` printing both
stamps; and the provisional-mark retirement. The fresh-context unit
read is a review dispatch rather than a task.

At 4a-ii's all-in rate of 0.54M per task, seven tasks is about 3.8M,
which is plainly wrong for this pass, and the reason is that the rate
is not transferable. 4a-ii's tasks were novel Go against a corpus, with
a staleness pre-task, two fix rounds, and a simplify pass over new
packages. 4b's tasks are configuration files, three-line shell edits,
and one comment sweep, and the largest single piece is the
render-and-install helper.

Two adjustments, both evidence-backed rather than asserted:

- The **chain rate**, 0.26M per task including review, is the closer
  base, because 4b spends almost nothing on fix rounds and simplify:
  there are two Go lines of new code in the whole pass.
- The pass carries **two fixed costs 4a-ii also paid**: the staleness
  pre-task (part of 0.26M there) and the close drafter (0.16M). Add the
  fresh-context unit read, which is one zero-context agent reading two
  short files, a diff, and a skill.

Seven tasks at 0.26M is 1.8M, plus about 0.5M of fixed costs, gives
about 2.3M as the honest projection. **Propose 2.5M**, above the spec's
1.5M, with the reason stated: the spec's figure predates 4a's
measurement and predates the discovery that edit 6 touches three files
rather than one. Set the checkpoint interval at four tasks, per the
default, which puts one checkpoint after the deploy-edit group.

Two caveats belong in the header. The pass **stays open at the sitting
boundary**, so the ceiling covers the unattended half and the sitting's
own spend is small and separate. And the estimate assumes no fix round;
the deploy-helper refactor is the one task where a `fix` verdict is
plausible, since it rewrites a block that must keep the
in-memory-only property.

Spend to date across the whole build is about 42.8M through 4a-ii
(`docs/STATUS.md:66`).

**What 4b inherits, concretely.** A measured 2.7M five-task pass with a
1.32M chain, a per-task chain rate of 0.26M, a task profile that is
strictly lighter than the one that produced the rate, and a spec
proposal of 1.5M that the measurement says to raise rather than keep.

---

## Surprises

Every place the tree contradicts the spec, the handoff, or the task
framing.

1. **The pre-create and the grant name different directories.** The
   handoff says `<state dir>/shadow/` "is 4b's `ReadWritePaths=`
   argument" (`docs/HISTORY.md:389-394`), while the spec's deploy edit 3
   says the pre-create step is for `/var/lib/dubplate`
   (spec `:1318-1322`). `SKILL.md:39-42` makes those mutually exclusive
   as written: a leaf grant needs the leaf pre-created. The plan owes a
   decision and both consistent options are in section 3.

2. **`deploy.sh` announces fourteen steps, not nine.** The spec says
   "443 lines, nine idempotent steps, each with a `--check` branch"
   (spec `:1304-1306`). The line count is right. `grep -n 'echo "== '`
   returns fourteen.

3. **The version verb is not where the spec cites it.**
   Spec `:1357-1359` cites `cmd/dubplate-server/main.go:56-64` as
   printing one line. At HEAD, `:56-64` is `newRootCmd`'s command
   construction; `newVersionCmd` is `main.go:72-81`.

4. **The threshold `Kind` work the spec assigns rung 4 is already
   landed.** Spec `:1466-1476` says `Threshold` is `{Key, Default, Doc}`
   with `ParseInt` as the only parser, that `thresholdStatus` calls
   `ParseInt` unconditionally, and that a duration would render as
   `invalid:` in the doctor table. Pass 4a landed `ThresholdKind`
   (`thresholds.go:10-18`), `ParseDuration` (`:74-96`), and the switch
   (`doctor.go:111-128`). Nothing is owed at 4b.

5. **Five provisional marks are retireable, not four, and one of the
   five sits outside `internal/engine/shadow`.** The task names four in
   that package. The tree has six occurrences: three ruling marks in
   `internal/engine/shadow` (`rescore.go:28`, `match.go:20`,
   `run.go:388`), the rung 8 handoff mark that stays (`flac.go:62`),
   and two more in `cmd/dubplate-server/shadow.go:22` and `:37` for
   rulings 4 and 5. `docs/STATUS.md:47-48` says six and calls the
   retirement mechanical.

6. **Edit 6 touches three files, and the pass-split row names one.**
   The scope row says "deploy.sh's six edits, the sixth being the two
   `-ldflags` stamps and the smoke-test assertion." `Makefile:6-7`
   carries no flag variable, so `deploy.sh` cannot pass values through
   `make build-linux` without a Makefile edit, and `main.go:72-81` is a
   third file.

7. **`EnvironmentFile` in the import unit is at `:15`, not `:14`.**
   Spec `:1166` cites `dubplate-import.service:14` for the
   `EnvironmentFile` line every sibling carries. Line 14 is
   `User=music`; line 15 is the `EnvironmentFile`.

8. **A `DUBPLATE_PING_SHADOW_URL`, if ever minted, cannot live in the
   greenfield template.** The spec wires `OnFailure=dubplate-ping-fail@shadow.service`
   and says the URL resolves straight from the process environment
   (spec `:1181-1199`). True, and the helper's environment comes from
   `EnvironmentFile=/etc/dubplate/env` (`ping-fail@.service:11`), the
   incumbent's file, which `dubplate-server` never reads and whose
   `DUBPLATE_PING_*` class the greenfield wall reserves for rung 9
   (`keys.go:116-125`). The spec's conclusion, that the line is wired
   and inert, holds. The mechanism is worth stating precisely in the
   plan so a later rung does not look for the name in the wrong file.

9. **`deploy.sh --check` is not runnable without a complete age
   store.** The render loop runs before any step and exits 1 listing
   every missing name (`:175-194`), and it resolves all 42 template
   rows, including the three `ND_*` names. So the exit-gate line
   "`deploy.sh --check` clean" depends on the workstation's
   `~/.local/secrets` being complete, and on `HCLOUD_TOKEN` plus the
   `hcloud` binary when no HOST argument is given (`:99-109`).

10. **`DUBPLATE_LIBRARY_ROOT` is required by the wall and read by
    nothing in the shadow.** `config.Load` requires it
    (`config.go:216`), so the greenfield template must carry it, and
    grep over `internal/engine/shadow` finds no reader. The spec's env
    table already rules it a literal `/srv/music/library`; the plan
    should say the row exists to satisfy the wall rather than to feed
    the observer.

11. **Doctor prints neither stamp.** Section 5 covers this.
    `doctor.go` derives its four row groups from `internal/config`'s
    tables and never mentions `version` or `matcherCommit`. If the
    sitting is meant to read the stamps from `doctor`, that is an
    unnamed deliverable; the scope row names only `dubplate-server
    version`.

12. **`version` is load-bearing beyond printing.**
    `cmd/dubplate-server/shadow.go:215` builds the MusicBrainz
    User-Agent from it, so stamping it changes what the box sends on
    every request. Neither the spec nor the handoff mentions this.

13. **The spec's citation for the open `/var/lib/musicbox` SELinux
    check points at the wrong lines.** Spec `:1243-1245` cites
    `docs/HISTORY.md:2327-2328`, which is rung 3a's adjudication and
    Tier C exclusion text. A grep of HISTORY for `SELinux` returns four
    lines, and the carried item is `docs/HISTORY.md:3082-3083`, "SELinux
    one-time check on /var/lib/musicbox at first deploy." The spec's
    claim is right and its pointer is not; the sitting's checklist
    should cite `:3082-3083`.
