# Rung 4 Spec Inventory

Research report, 2026-09-05, prepared for the rung 4 spec's authoring,
against HEAD `440131e3569dd5237432c60c6708d1cf29b59056`. Scope, quoted
from the build spec's ladder entry
(`docs/superpowers/specs/2026-09-02-greenfield-server-build.md:572-580`):

> Rung 4 | Shadow matching | The shadow runner under `engine/shadow` and
> its subcommand; the divergence log; the adjudication artifact; a
> hardened systemd unit per vps-conventions with pre-created
> `ReadWritePaths`; its environment file; and the `deploy.sh` path rung
> 1 repointed. | [E], M to build, continuous to launch.

Four sections follow: (a) the estate and the box as landed, (b) the
engine and harness as landed, (c) the ratified record for rung 4, and
(d) cross-section consistency and open questions.

---

## (a) The estate and the box as landed

Everything below is read from the repo tree at `440131e3` and from
`docs/HISTORY.md`. No box access occurred.

### 1. What the box runs today

Nine unit files ship, four timers are enabled by the deploy script, and
two units are deliberately left disabled.

| Unit | Trigger | ExecStart | Writes |
|---|---|---|---|
| `dubplate-import.service` | `dubplate-import.timer`, `*:0/15` + `RandomizedDelaySec=300`, `Persistent=true` (`systemd/dubplate-import.timer:10,9,11`) | `/usr/local/bin/dubplate import` (`systemd/dubplate-import.service:55`) | `ReadWritePaths=/srv/music /var/lib/musicbox` (`:54`) |
| `dubplate-backup.service` | `dubplate-backup.timer`, `*-*-* 11:30:00 UTC` + `RandomizedDelaySec=600` (`systemd/dubplate-backup.timer:9,10`) | `/usr/local/bin/dubplate backup` (`systemd/dubplate-backup.service:41`) | `ReadWritePaths=/var/lib/musicbox /srv/music/state/backup-staging` (`:40`) |
| `dubplate-disk-check.service` | `*:0/15` + 300s jitter (`systemd/dubplate-disk-check.timer:9,10`) | `/usr/local/bin/dubplate check disk` (`:24`) | none, no `ReadWritePaths` (`:21-23`) |
| `dubplate-navidrome-check.service` | `*:0/5` + 60s jitter (`systemd/dubplate-navidrome-check.timer:8,9`) | `/usr/local/bin/dubplate check navidrome` (`:24`) | none (`:21-23`) |
| `dubplate-ping-fail@.service` | `OnFailure=` instance only, never enabled | `/usr/local/bin/dubplate ping-fail %i` (`:20`) | none (`:15-19`) |
| `dubplate-serve.service` | shipped, not enabled by deploy (`scripts/deploy.sh:400-405`) | `/usr/local/bin/dubplate serve` (`:52`) | `/var/lib/musicbox /srv/music/uploads /srv/music/inbox /run/dubplate/nudge-triggers` (`:94`) |
| `dubplate-import-nudge.path` / `.service` | `DirectoryNotEmpty=/run/dubplate/nudge-triggers` (`systemd/dubplate-import-nudge.path:16`), also not enabled by deploy | `/usr/local/bin/dubplate nudge-consume` (`.service:41`) | `/run/dubplate/nudge-triggers /srv/music /var/lib/musicbox` (`.service:40`) |

Every service carries `User=music`, `EnvironmentFile=/etc/dubplate/env`,
`NoNewPrivileges=yes`, `PrivateTmp=yes`, `ProtectSystem=strict` (for
example `systemd/dubplate-import.service:14-18`). `dubplate-serve.service:96-125`
carries the extended set on top, including `SystemCallFilter=@system-service`
and `RestrictAddressFamilies=AF_INET AF_INET6 AF_UNIX AF_NETLINK`. That
extended block is the template a new hardened unit should copy from, not
the oneshot baseline.

Every service also carries `OnFailure=dubplate-ping-fail@<slug>.service`
(`systemd/dubplate-import.service:8`, `dubplate-backup.service:12`,
`dubplate-disk-check.service:10`, `dubplate-navidrome-check.service:10`,
`dubplate-serve.service:35`). The slug resolves to
`DUBPLATE_PING_<SLUG>_URL` read straight from the process environment,
bypassing config load (`systemd/dubplate-ping-fail@.service:15-19`). An
unset URL no-ops rather than erroring, so a shadow unit can wire
`OnFailure=dubplate-ping-fail@shadow.service` before any check is
minted.

Ping contract, confirmed in code: success is `GET` to the URL
(`_salvage/internal/ping/ping.go:31-35`), failure is `POST` to
`<url>/fail` (`:41-45`).

The compose stack is Navidrome plus cloudflared only. Navidrome runs as
`2000:2000`, mounts `/srv/music/library` read-only and
`/srv/music/state/navidrome` read-write, and publishes only
`127.0.0.1:4533` (`compose.yaml:26-59`). Its filesystem watcher is off
and a daily `0 3 * * *` scan is the backstop (`compose.yaml:37-38`,
`config/navidrome.toml:20-22`). cloudflared runs a token-managed tunnel
whose ingress lives in the Cloudflare API, and the repo's record of that
ingress serves `music.907.life` to Navidrome with a 404 catch-all
(`config/cloudflared-tunnel-config.json:2-13`). The uploader path rule
and the FileBrowser hostname were both retired on 2026-09-02, so the
tunnel serves Navidrome and nothing else.

The box's directory tree comes from cloud-init at first boot: the
Hetzner volume is formatted and mounted at `/srv/music`, then `library`,
`inbox`, `review`, `state`, and `staging` are created `2000:2000` with
`inbox` setgid (`cloud-init/user-data.yaml:95-122`), and the `music`
user is fixed at uid/gid 2000 with `/srv/music` as its home
(`:127-130`). SELinux is set enforcing via `/.autorelabel` plus the
top-level `power_state` reboot (`:251-255`, `:271-274`). cloud-init runs
at first boot only and never converges a live box
(`~/.claude/skills/vps-conventions/SKILL.md:57-66`), which is why
`/srv/music/uploads` is created by the deploy script instead
(`scripts/deploy.sh:374-389`).

Infrastructure provisioning is `scripts/provision.sh`: a `musicbox-fw`
firewall whose only inbound rule is tcp/22 from `FIREWALL_SSH_CIDRS`
and which refuses to run rather than fall back to `0.0.0.0/0`
(`scripts/provision.sh:86-91`, `scripts/firewall-rules.json`), a 50 GB
`musicbox-data` volume, and a `cpx21` server in `hil` on `alma-10`
(`:14-24`). It is the only thing in the repo that talks to hcloud, and
it plays no part in a shadow deploy.

### 2. What the import timer consumes, and where it lands

The frozen `/usr/local/bin/dubplate` binary is built from the
pre-pivot tree now under `_salvage/` (`scripts/deploy.sh:251-260`), so
`_salvage/internal/` is the authoritative source for the box's live
behavior.

One `dubplate import` run, in order:

1. **Lock.** `flock(2)` on `/var/lib/musicbox/import.lock`
   (`_salvage/internal/lock/lock.go:24,29-31`). A held lock is healthy:
   the run writes a `LockHeld` report, pings success, and returns nil
   (`_salvage/internal/importer/importer.go:397-410`).
2. **Report opened.** A per-run `report-<runid>.json` is written to
   `DUBPLATE_STATE_DIR` on every exit path via a deferred write
   (`importer.go:427-450`, `:518-531`).
3. **Free-space floor**, then `MkdirAll` over staging, review, and
   state (`importer.go:337-352`).
4. **Quiescence and staging.** For each directory under
   `/srv/music/inbox/<name>/`, every file's mtime must be at or before
   `now - QuiescentMin` minutes, default 10 (`importer.go:790-798`,
   `_salvage/internal/config/config.go:152`). A quiescent entry is
   moved out of the inbox into `/srv/music/staging/<name>/`
   (`importer.go:812`). Archives are expanded natively with traversal
   refusal and a byte budget (`_salvage/internal/archive/archive.go:1-5`).
5. **Quality gate.** `gate.Check` over the whole staging tree enforces
   extension, symlink, decode integrity, and CD-quality floors
   (`importer.go:945`, `_salvage/internal/gate/gate.go:1-4`, floors at
   `importer.go:88-91`). A failing album moves into
   `/srv/music/review/<YYYY-MM-DD>/<uploader>/`, with
   `/srv/music/quarantine/...` as the fallback (`importer.go:499-505`,
   `_salvage/internal/review/review.go:47-59`). Gate-rejected material
   never reaches beets.
6. **One beet import** over the whole staging directory:
   `beet -c <config> import -q -l <scratch log> [--set collection=<c>]
   /srv/music/staging` (`_salvage/internal/beets/beets.go:83-96`,
   called at `importer.go:1100`). Beets is configured `move: yes`,
   `write: yes`, `quiet_fallback: skip`, `none_rec_action: skip`,
   `group_albums: yes` (`config/beets-config.yaml:13-23`), library at
   `/srv/music/library`, database at `/srv/music/state/library.db`
   (`:10-11`).
7. **Tag log persisted.** The scratch log is copied atomically to
   `/var/lib/musicbox/import-<YYYYMMDD-HHMMSS.nnnnnnnnn>.log` and
   re-read from the persisted copy (`beets.go:103-113`). Skipped paths
   are parsed from `skip <path>[; <path>]` lines (`beets.go:120-135`)
   and each is moved into the same dated review tree
   (`importer.go:1163-1175`).
8. **Library paths resolved from the database.** `sqlite3 -noheader
   -separator '\t' /srv/music/state/library.db "SELECT album_id, path
   FROM items WHERE added >= <pre> AND added <= <post> ORDER BY
   album_id, path"`, matched positionally back to the staged units
   (`beets.go:161-197`, `importer.go:1116-1136`). The doc comment
   states the reason plainly: the tag log records only skips and beets
   renames every file, so the database is the sole surviving record.
9. **Tail.** Ledger append, empty-staging cleanup, a best-effort
   Navidrome scan trigger, review and quarantine sweep at
   `RetentionDays` 60, run-report sweep at `ReportRetention` 30,
   success ping (`importer.go:360-393`, defaults at
   `_salvage/internal/config/config.go:152-172`).

What actually feeds the inbox today: the wind-down retired every
contributor intake path (`docs/HISTORY.md:2153-2159`), and material
reaches `/srv/music/inbox/geoff/` by hand, the same way the bring-up
runbook seeds it (`docs/bringup-runbook.md:30-31`).
`docs/STATUS.md:46-47` records two albums still sitting in
`/srv/music/review/2026-08-31/geoff/`.

### 3. What the oracle can attest, and where it is written down

The build spec's claim that the oracle is coarse is confirmed against
pinned beets 2.13.1. In quiet mode `_summary_judgment` returns `APPLY`
when and only when the recommendation is `strong`, and otherwise
resolves `quiet_fallback` to `SKIP`
(`tools/venv/.venv/lib/python3.13/site-packages/beets/ui/commands/import_/session.py:335-341`).
The tag log records `skip`, `asis`, and the three duplicate statuses,
never an apply (`.../beets/importer/session.py:142-167`). So the live
oracle emits one bit per album, with no recommendation level, no
distance, and no candidate list.

Four durable artifacts carry that bit, with different reach:

| Artifact | Path | Content | Retention |
|---|---|---|---|
| beets database | `/srv/music/state/library.db` | applied albums only: `albums.mb_albumid`, `albumartist`, `album`, `added`, plus `items.path`, `items.added` (`.../beets/library/models.py:302`, `.../beets/library/fields.py:6`) | permanent, backed up nightly |
| tag log | `/var/lib/musicbox/import-*.log` | skipped staging paths only, written atomically after the import returns | never swept |
| run report | `/var/lib/musicbox/report-*.json` | per album: `uploader`, `source_dir`, `outcome` (imported / review / quarantined / rejected), `session_id`, `library_paths`, `collection`, `error` (`importer.go:126-172`) | 30 newest files (`importer.go:540-574`), which at `*:0/15` is roughly 7.5 hours |
| review tree | `/srv/music/review/<date>/<uploader>/` | the skipped album's original files, untagged by beets | 60 days (`_salvage/internal/config/config.go:155`) |

The run report is the richest single artifact and the one that names
the applied/skipped split per album directly. Its 30-file cap is the
sharpest constraint in this whole survey.

**The design hazard the plan must confront.** Beets runs with
`move: yes` and `write: yes` (`config/beets-config.yaml:14-15`), so an
applied album's files are renamed into the library and their tags
rewritten with beets' own answer. After the import, the input the
matcher would need is gone. Reading the library files back gives
`tag.Album` a pre-tagged input and a trivially confirming verdict,
which measures nothing. Any observation mechanism that starts after the
import can recover the oracle's answer but not the question. Only a
capture before step 6 preserves both.

### 4. Candidate observation mechanisms

| # | Mechanism | Read-only guarantee | Race with the beets move | Oracle signal recoverable | Box-write footprint |
|---|---|---|---|---|---|
| A | Poll `/srv/music/inbox/<user>/` before the importer stages it | Grant no `ReadWritePaths` under `/srv/music`; `ProtectSystem=strict` makes any write fail | Real. The window is `QuiescentMin` (10 min) after the last write, ending whenever the 15-minute timer next fires. A `rename(2)` out from under an open fd is safe on Linux, so an in-flight read completes | None directly. This captures the question, not the answer: item fields as they arrived | Zero |
| B | Poll `/srv/music/staging/` between staging and the beet import | Same as A | Fatal. The window is inside one `runPipeline` call (`importer.go:273` to `importer.go:1100`), typically seconds | Same as A | Zero |
| C | Read `/srv/music/state/library.db` after the import | Needs `mode=ro` or `immutable=1`; a default sqlite open can create `-wal`/`-shm` or a hot journal in `/srv/music/state`, which the sandbox would reject and turn into an error rather than a write | Small. The DB row lands in the import's own transaction; a poller offset from the timer avoids the window entirely | Applied albums: the chosen `mb_albumid`, the album/artist strings beets wrote, `added`, item paths. Nothing about skips, which leave no row | Zero, if opened read-only |
| D | Read `/var/lib/musicbox/import-*.log` | No write grant needed | None. Written via `atomicfile.WriteFile` after `beet import` returns (`beets.go:104`), so a reader never sees a partial file | Skipped staging paths, by path only. Never an apply | Zero |
| E | Read `/var/lib/musicbox/report-*.json` | No write grant needed | None. Atomic per-run write (`importer.go:529`) | The full applied/skipped/rejected split per album, with `library_paths` as the join key into C. The single best answer artifact | Zero |
| F | A `.path` unit watching the state directory for a new report | n/a | `PathChanged=` on a directory fires on tag logs, lock activity, and atomicfile temporaries too, so it is noisy. Path units coalesce while their service runs and lose events in the re-arm gap, and require a timer backstop regardless (`SKILL.md:47-55`) | Latency only. Recovers nothing E does not | Zero |
| G | `ExecStartPost=` on `dubplate-import.service` | Poor. It runs inside the incumbent's sandbox, inheriting `ReadWritePaths=/srv/music /var/lib/musicbox`, a far wider grant than a shadow needs | Runs after the move, so it inherits the same too-late problem as C | Same as C plus E | Zero writes of its own, but it edits a frozen unit and, absent a `-` prefix, a shadow failure fails the import unit and fires `dubplate-ping-fail@import.service`, alerting on the wrong thing |
| H | Poll `/srv/music/review/<date>/<user>/` | Same as A | None. Files are at rest once filed | Skipped albums with their original tags intact, since a skip means beets applied nothing | Zero |

Reading these together, no single mechanism is sufficient. The workable
shapes are combinations:

- **A plus E** is the minimal complete pair. A captures each album's
  item fields while they sit quiescent in the inbox and keys them by
  inbox path. E later tells the shadow what the oracle decided for that
  same path, since `AlbumOutcome.SourceDir` is the uploader's inbox
  directory (`importer.go:1066`). C then supplies the applied release
  MBID for the imported half by joining through `library_paths`.
- **H plus C** is the cheap fallback if the plan declines a pre-import
  capture. It gives original tags for the skipped half only, and for
  the applied half it has the oracle's release MBID with no honest
  input to re-run the matcher against. That halves the shadow's reach
  and is worth naming as such rather than discovering later.
- **D** is a useful durability backstop for E, because tag logs are
  never swept while reports survive roughly 7.5 hours.
- **F and G** should both be declined. F buys latency a timer already
  provides at 15-minute granularity, and G violates the additive-deploy
  boundary rung 1 established and every rung before 16 must preserve
  (`docs/HISTORY.md:2097-2100`).

One trap worth stating explicitly in the plan: the shadow must not
acquire `/var/lib/musicbox/import.lock`. The incumbent treats a held
lock as healthy and returns nil without importing (`importer.go:397-410`),
so a shadow holding that lock would silently stall the box's own
import while looking green.

### 5. What a new hardened unit needs from `scripts/deploy.sh`

`deploy.sh` is 443 lines and does exactly nine things, all idempotent,
all with a `--check` branch: rsync `compose.yaml` + `config/` to
`/opt/dubplate` (`:198-209`); rsync the `UNIT_FILES` array to
`/etc/systemd/system` (`:211-217`); install the tmpfiles fragment and
run `systemd-tmpfiles --create` (`:219-232`); `make build-linux` and
rsync the result to `/usr/local/bin/dubplate-server`, chmod 0755, then
a `dubplate-server version` smoke test that aborts the deploy on
failure (`:251-289`); render `/etc/dubplate/env` in memory and
`install -D -m 600 -o root -g root` from stdin (`:291-316`); install
`config/contributors.yaml` at `0640 root:music` (`:322-328`);
pre-create `/srv/music/state/navidrome` (2000:2000),
`/var/lib/musicbox` (music:music), `/srv/music/state/backup-staging`
(music:music), `/srv/music/uploads` (music:music) (`:330-389`);
`docker compose --env-file /etc/dubplate/env ... up -d` (`:391-394`);
`systemctl daemon-reload && systemctl enable --now $TIMER_UNITS`
(`:396-408`); install beets via a pinned, sha256-verified uv
(`:410-441`).

A shadow unit therefore needs five edits and one new artifact:

1. **`UNIT_FILES`** gains the service and, if timer-driven, the timer
   (`:56-64`). Manifest edits land in the same commit as the file,
   because rsync exits 23 on a missing source and aborts the deploy
   (`SKILL.md:70-74`).
2. **`TIMER_UNITS`** gains the timer if the shadow should come up
   unattended (`:407`). The precedent for holding a unit back is
   `dubplate-serve.service`, enabled only by a sanctioned hand step
   (`:400-405`). A read-only shadow has none of serve's exposure, so
   enabling it here is defensible, and the spec should say which it
   chooses.
3. **A pre-create step** in the shape of `:357-359`: `run_remote
   "mkdir -p <dir> && chown music:music <dir>"`. Mandatory before the
   unit's first start, or systemd fails with "Failed to set up mount
   namespacing" (`SKILL.md:39-42`, restated at
   `scripts/deploy.sh:350-356`).
4. **An environment file.** See section 6; this is the one
   non-mechanical piece.
5. **Secrets.** A read-only shadow needs none beyond the wall's own
   rows. `DUBPLATE_ADMIN_EMAIL` already exists in the manifest
   (`env/dubplate.env.template:25`). Any genuinely new secret goes
   through `secret-set.sh` and the registry per CLAUDE.md's secrets
   rule; the template header says plainly that adding a name there
   requires minting it.

The build target is already in place. `make build-linux`
cross-compiles `CGO_ENABLED=0 GOOS=linux GOARCH=amd64` from
`./cmd/dubplate-server` (`Makefile:6-7`), and deploy.sh already ships
it, chmods it, and smoke-tests `version` on the box (`:273-289`). Rung
4 supplies a subcommand; the CLI today is `check`, `doctor`, and
`version` (`cmd/dubplate-server/check.go:18`, `doctor.go:25`,
`main.go:58`).

The gate already covers the new unit. `scripts/check.sh:125-148` runs
`systemd-analyze verify` over `systemd/*.service`, `*.timer`, `*.path`,
filtering only the expected missing-`ExecStart` complaint, so a
malformed unit reds the gate without any new wiring.

Two SELinux notes. New directories under `/var/lib` inherit
`var_lib_t`, fine for these services but to be verified once on the
first real deploy (`SKILL.md:43-45`). The identical check is already
carried, unclosed, for `/var/lib/musicbox`
(`docs/HISTORY.md:2327-2328`), so the shadow's first deploy is the
natural place to close both.

What `deploy.sh` does not do, and the plan should not assume: it never
touches `/usr/local/bin/dubplate`, the incumbent (`:251-260`); it
never enables `dubplate-serve.service` or the nudge path
(`:400-405`); it does not create
`/srv/music/{library,inbox,review,state,staging}`, which are
cloud-init's (`cloud-init/user-data.yaml:116-122`); it deliberately
never creates `QuarantineDir`
(`systemd/dubplate-import.service:43-48`); it applies no tunnel
ingress (`config/cloudflared-tunnel-config.json:2`); and it has no
rollback, no soak, and no post-deploy unit-status verification beyond
the binary smoke test.

### 6. The environment-file collision, and why it is the blocking design item

`internal/config.Load` reads the whole process environment and errors
on every `DUBPLATE_*` name that is not one of the four owned rows:
reserved names error naming their class and rung, superseded names
error naming their disposition, and anything else errors as unknown
(`internal/config/config.go:133-152`, `:159-175`). The owned rows are
`DUBPLATE_LIBRARY_ROOT`, `DUBPLATE_STATE_DIR`, `DUBPLATE_SITE_NAME`,
`DUBPLATE_ADMIN_EMAIL` (`internal/config/keys.go:19-25`).

`/etc/dubplate/env` carries 39 `DUBPLATE_*` names
(`env/dubplate.env.template`). Two of them are owned rows
(`DUBPLATE_ADMIN_EMAIL`, `DUBPLATE_STATE_DIR`). The other 37 all
produce a load error, and two owned rows the greenfield binary
requires (`DUBPLATE_LIBRARY_ROOT`, `DUBPLATE_SITE_NAME`) are absent,
since they were renamed from `DUBPLATE_LIBRARY_DIR` and
`DUBPLATE_SERVE_SITE_NAME` (`internal/config/keys.go:165-166`).

A shadow unit with `EnvironmentFile=/etc/dubplate/env` therefore fails
at load, by design, every time. The spec must call for a second,
greenfield-only environment file, and the template already
anticipates it: its own header says this manifest describes the
frozen pre-pivot binary's wall and that "this manifest evolves at rung
4" (`env/dubplate.env.template:1-3`).

Two consequences for deploy.sh's renderer. First, the render loop
requires every `NAME=` line in the template to resolve to a
non-empty value already exported in the workstation shell or sourced
from `~/.local/secrets`, and fails listing every miss
(`scripts/deploy.sh:175-194`). There is no mechanism for a literal
in-template default. `DUBPLATE_LIBRARY_ROOT`, `DUBPLATE_STATE_DIR`, and
`DUBPLATE_SITE_NAME` are non-secret path and identity values, so either
they get added to the age store as if they were secrets, or the
renderer gains a literal column. That is a real, small deploy.sh
design decision the plan should name rather than leave to the
implementer. Second, the render and install steps are written for
exactly one file path (`REMOTE_ENV` at `:28`), so a second file means
either a parameterized helper or a duplicated block.

Separately, `DUBPLATE_BEETS_DB` is superseded outright with the
message "retired: the greenfield server never embeds beets"
(`internal/config/keys.go:178`), and `DUBPLATE_BEETS_CONFIG` alongside
it. If the shadow reads `/srv/music/state/library.db` (mechanism C),
it needs a way to name that path. The greenfield wall's current
position forbids the obvious name. The spec has to choose: a new
shadow-class key, a reclassification of `DUBPLATE_BEETS_DB` into a
rung-4 reserved class, or a subcommand flag outside the wall.

On registration mechanics: rung 1b has not landed, so
`internal/config/keys.go` is still the one file every rung's keys
live in, and rung 4 is the only live code lane, so it may register
there under the existing pattern. The `.gitleaksignore` fragility is
real and worth restating in the spec: the single allowed finding is
pinned by line number at
`internal/config/keys_test.go:generic-api-key:66`
(`.gitleaksignore:1`), and line 66 today is the `DUBPLATE_R2_BUCKET`
fixture row (`internal/config/keys_test.go:66`). Any insertion above
it in that test file reds the gate for every lane at once.

### 7. How the quiet-box discipline is checked today

It is not. There is no mechanism anywhere in the repo. The rule is
stated as doctrine in `CLAUDE.md:176-177` ("no writes without a
quiet-box check; read-only ssh only during reviews") and assigned to
specific rungs in the build spec
(`docs/superpowers/specs/2026-09-02-greenfield-server-build.md:864-874`)
and the parallel-approach spec (`:105`, `:123`). `scripts/deploy.sh`
has no timer-quiescence probe, and `scripts/check.sh` has no box leg
at all. The closest thing to a mechanical quiet check is `--check`
mode, which diffs every step and applies none
(`scripts/deploy.sh:73-88` and every step's `if $check_only` branch),
and the `remote_reachable` `BatchMode=yes` probe that makes an
unreachable box report honestly instead of showing an empty remote
file as a pending change (`:113-121`).

For rung 4 specifically, the build spec says rung 4 deploys the
additive shadow runner under its own unit and that rung 4's additive
unit is compatible with rung 10's read-only listing (`:869-870`,
`docs/superpowers/research/2026-09-03-rung-dependency-analysis.md:396-397`).
The write is additive, so a full quiet-box ritual is arguably not owed
here, but the discipline is currently a human procedure rather than a
check, and the spec should say which it is invoking.

### 8. Durable state for the divergence log

`/var/lib/musicbox` exists on the box today, `music:music`,
pre-created by deploy.sh and re-created idempotently on every run
(`scripts/deploy.sh:340-359`). It already holds `import.lock`
(`_salvage/internal/lock/lock.go:24`), `backup.lock`
(`_salvage/internal/backup/backup.go:44`), `import-*.log`,
`report-*.json`, and `contributions.jsonl`
(`_salvage/internal/config/config.go:345`). It is the estate's one
persistent, non-`RuntimeDirectory` state location, which is exactly
what vps-conventions requires for durable state (`SKILL.md:15-19`).

`DUBPLATE_STATE_DIR` is an owned greenfield wall row
(`internal/config/keys.go:21`) with no box path assigned to the
greenfield binary yet. Two shapes, both viable:

- **Share `/var/lib/musicbox`.** Zero new deploy steps, zero new
  SELinux surface, and it honors the standing ruling that box data
  paths keep their existing names while only the config and identity
  layer renamed (`docs/HISTORY.md:2211-2213`,
  `env/dubplate.env.template:59-70`). The incumbent's report sweep
  globs `report-*.json` specifically (`importer.go:544`), so a
  differently named divergence file is safe from it. The cost is that
  the frozen incumbent and the greenfield binary share a directory
  that must survive rung 16's retirement of the incumbent.
- **Mint `/var/lib/dubplate`.** One new pre-create line in the shape
  of `:357-359`, one SELinux verification on first deploy, and a
  clean boundary that retires with the incumbent at rung 16. It
  contradicts nothing, since the "keep the data path name" ruling was
  about migrating a live directory, not about creating a new one.

Either way the divergence log wants atomic writes, which the estate
already has a house pattern for (`internal/atomicfile`, used at
`importer.go:529` and `beets.go:104`), and the row format is already
fixed by the rung 3 handoff: `id`, `subject.album` plus
`subject.release_mbid`, `suspected`, `citation`, `release_truth`,
`disposition`, `reason`, `adjudicator`, `date`, plus the
matcher-commit stamp ruling B requires (`docs/HISTORY.md:267-303`).

The MusicBrainz side is already provisioned: `ClassBackground` is
defined and documented as rung 4's shadow class
(`internal/engine/musicbrainz/limiter.go:25`), `DefaultInterval` is
one second (`:62`), and `ErrUnavailable` is the sentinel to branch on
for deferral (`internal/engine/musicbrainz/client.go:84`).

---

## (b) The engine and harness as landed

Every citation below is `path:line` against the working tree at
`440131e`.

### 1. `internal/engine/tag`: the entry point rung 4 calls

The whole public surface is one function and two types.

```
func Album(ctx context.Context, src *musicbrainz.Client, items []matcher.ItemFields,
    tun matcher.Tunables, compat matcher.Compat, evalDate time.Time, opts Options) (*Proposal, error)
```
`internal/engine/tag/tag.go:83-84`.

- `Proposal` (`tag.go:38-44`) carries `CurArtist`, `CurAlbum`,
  `Candidates []matcher.Candidate` (ascending distance, stable),
  `Recommendation matcher.Recommendation`, `Action matcher.Action`.
  Nothing else. There is no request log, no per-request record, no
  timing, and no identity of the winning release beyond
  `Candidates[0].Info`.
- `Options` (`tag.go:54-56`) has one field, `SearchIDs []string`,
  which selects `import.search_ids`'s explicit-ID branch and skips the
  search entirely (`tag.go:93-101`). The shadow leaves it zero.
- `ErrNoCandidates` (`tag.go:33`) is returned when nothing was
  admitted. `tag.go:78-82` names rung 4 explicitly: a 404 drops one
  candidate silently (`lookupAndAdd`, `tag.go:178-191`, mirroring
  `album_for_id`'s `return None`), and every other client error is
  returned wrapped, `musicbrainz.ErrUnavailable` included.
- `Album` calls `compat.Apply(tun)` then `tun.Validate()` before any
  I/O (`tag.go:85-88`), so a malformed shadow configuration fails on
  the first album rather than mid-run.
- Every request `Album` issues spends one token from `src`'s budget
  (`tag.go:70-74`). Per-album spend is not fixed: the strong-match
  early return costs exactly one lookup (`tag.go:117-133`), while the
  search path costs one search plus one lookup per returned id
  (`tag.go:140-149`), and a pseudo-release lookup costs a second
  request inside `AlbumForID`
  (`internal/engine/musicbrainz/client.go:325-333`). The shadow cannot
  precompute its per-album cost; it can only cap it with the budget's
  `limit`.
- `Album` reads no clock. `evalDate` is a parameter, enforced
  tree-wide by `TestEngineNeverReadsTheClock`
  (`internal/harness/conform/arch_test.go:152`).

### 2. What a shadow runner must construct, and the one thing that does not exist

`Album`'s six inputs, in order of how much work each costs rung 4:

| Input | Where it comes from | Status |
|---|---|---|
| `ctx` | the runner's own cancellation | trivial |
| `src` | `musicbrainz.NewClient` (section 3) | needs a User-Agent and a Budget |
| `items` | one `matcher.ItemFields` per track of the observed album | no producer exists |
| `tun` | `matcher.DefaultTunables()` plus the estate's import keys (section 5) | producer exists only inside the harness |
| `compat` | `matcher.DefaultCompat()` or `FullCompat()` (section 5) | trivial, but it is a ruling |
| `evalDate` | the runner's observation date, passed in | trivial |

The `items` gap is the seam's only real hole. `matcher.ItemFields`
(`internal/engine/matcher/info.go:124-152`) is nineteen typed fields
off one user item. Two producers exist and neither serves the box:

- `matcher.ProjectItemFields(raw map[string]any)` (`info.go:189`)
  projects a corpus case's decoded JSON item bag. Its input is
  `cassette.ItemFields`, a `map[string]any`
  (`internal/harness/cassette/cassette.go:66`), read from
  `case.json`. `internal/harness/conform`'s `caseInputs` is its only
  caller (`internal/harness/conform/case.go:320-323`).
- `synth.ProjectItems` (`internal/harness/synth/seeditems.go:28`)
  derives items from a captured WS/2 release document, for authoring
  synthetic cases.

Neither reads a file. Nothing in the greenfield tree reads a tag off a
FLAC. `go.mod` requires `github.com/spf13/cobra` and nothing else.
Rung 8's `internal/engine/tagio` is named as planned in
`internal/engine/tag/doc.go:8-11` and does not exist.

**Salvage-tier tag I/O: there is none to adopt.** I searched
`_salvage/` for FLAC metadata reading. The only FLAC contact is
`_salvage/internal/gate/gate.go:120-138`, which shells out to
`metaflac --show-sample-rate --show-bps` for the CD-quality check. It
reads no Vorbis comment. No `_salvage` file reads `artist`, `album`,
`title`, `track`, or `mb_albumid` off a file; the pre-pivot server
delegated all tag knowledge to `beet`
(`_salvage/internal/beets/beets.go:1-3`, "wraps the beet CLI"). So the
salvage tier offers rung 4 exactly one usable idea, the
`exec.CommandContext` shape around `metaflac`, and no code.

Rung 4 therefore chooses one of three, and this is a spec decision,
not an implementation detail:

1. **Read tags from the box's beets database.** The library DB
   (`/srv/music/state/library.db`, `config/beets-config.yaml:11`)
   already holds every field `ItemFields` needs, in beets' own names
   and types, for every applied album. It is the cheapest producer and
   the most faithful one, because `ProjectItemFields` reproduces
   beets' typed-default projection (`info.go:178-188`) against exactly
   those field names (`info.go:156-176`). It has one fatal limit: a
   skipped album is never in the DB (CLAUDE.md, "the beets database is
   the only import success record"), so this path observes only the
   applied half.
2. **Shell to `metaflac --export-tags-to=-`** (or `beet` itself) per
   file. Covers skips, adds a subprocess dependency to a hardened
   unit, and puts tag-name normalization in rung 4 rather than rung 8.
3. **Pull rung 8's FLAC reader forward.** Out of band for an M rung and
   contested with lane 8.

The oracle's own coarseness pushes toward a hybrid, and section 11
states why.

### 3. `internal/engine/musicbrainz`: client, budget, limiter

- `ClientConfig` (`client.go:93-99`): `BaseURL` (empty means
  `DefaultBaseURL`, `client.go:21`), `UserAgent` (required), `Budget
  *Budget` (required), `Timeout` (zero means 30s, `client.go:38`),
  `HTTP *http.Client`. `NewClient` refuses a missing `UserAgent` or
  `Budget` rather than defaulting (`client.go:124-130`).
- **The 3c User-Agent contact question, carried and still open.**
  `client.go:86-92` states the intended composition verbatim: "the
  identity is site-specific, so it belongs to `internal/config` and
  reaches here through `cmd/dubplate-server`'s composition
  (`dubplate-server/<version> ( <contact> )`, on
  `cmd/harness/capture.go`'s own precedent, with the contact from
  `Config.AdminEmail`)." `Config.AdminEmail` exists and is normalized
  (`internal/config/config.go:60-64`, `:110-125`). The composition
  function does not exist anywhere in `cmd/dubplate-server`. The
  precedent is `cmd/harness/capture.go:26`, which takes its contact
  from `DUBPLATE_HARNESS_CONTACT` and refuses to run unidentified
  (`capture.go:116-120`). Rung 4 is the first rung that talks to the
  live API from the server binary, so rung 4 owns building that
  string. `docs/HISTORY.md:308-310` records it as an unmoved carry
  ("`AdminEmail`'s interim User-Agent contact sits beside it").
- `Budget` (`limiter.go:337-344`): `name`, `limit`, an injected `wait`,
  and a mutex-guarded `spent`. `Wait` reserves before it waits and
  releases on any error path (`limiter.go:372-397`), returning
  `ctx.Err()`, `ErrBudgetExhausted` (`limiter.go:48`), or
  `ErrLimiterClosed` (`limiter.go:53`). `Spent()` (`limiter.go:400`)
  and `Name()` (`limiter.go:407`) are the two accessors a shadow log
  would use to report its own spend.
- `Unmetered()` (`limiter.go:354-359`) never waits and never
  exhausts. It is the conformance path's budget
  (`internal/harness/conform/case.go:330-336`) and the wrong one for
  the box.
- `Class` (`limiter.go:20-26`): `ClassUserPresent`, `ClassBatch`,
  `ClassBackground`. `limiter.go:16-18` names rung 4 directly: "rung
  4's shadow spends deferred background budget through this limiter
  and never inline with the probe." `Limiter.Budget(name, class,
  limit)` (`limiter.go:139-145`) is the constructor.
- `DefaultInterval = time.Second` (`limiter.go:62`), the MusicBrainz
  wire pace, citing beets' own `rate_limit = 1.0`.
  `reservationInterval = 4` (`limiter.go:43`) is the anti-starvation
  floor.
- `ErrUnavailable` (`client.go:74-84`) explicitly names the shadow:
  "Rung 4's shadow branches on it to defer a lookup rather than record
  a divergence." `ErrNotFound` (`client.go:67-72`) is consumed inside
  `tag.Album` and never reaches the shadow.
- `doWithRetry` acquires a limiter token before every attempt
  including retries (`client.go:336-342`), with the reason stated as
  "reproducing that accident here would make rung 4's shadow budget a
  lie." So a retried request costs the shadow more than one token, and
  `Budget.Spent()` is the honest count.

### 4. `internal/engine/matcher`: tunables, compat, verdicts

- `Tunables` (`tunables.go:61`) is the verified-complete surface,
  roughly thirty verdict-moving fields, each carrying pinned beets'
  shipped default with its citation. `DefaultTunables()`
  (`tunables.go:266`). `tunables.go:50-54` names the four
  operator-facing knobs deferred to the post-1b registration:
  `StrongRecThresh`, `MediumRecThresh`, `RecGapThresh`, `SearchLimit`.
- `Tunables.Validate()` (`tunables.go:455`) resolves every name-typed
  and pattern-typed field and reports every problem, not the first. It
  is a config-load check, not a precondition (`tunables.go:448-454`).
  `tag.Album` calls it per album (`tag.go:86`), which is redundant
  work the shadow can ignore but not skip.
- `Compat` (`compat.go:15-22`) is two booleans, `FtPattern` and
  `RemixParenthetical`. `FullCompat()` (`compat.go:31-33`) reproduces
  every ledgered defect and is what `conform`'s `ModeCompat` runs.
  `DefaultCompat()` (`compat.go:41-46`) turns both off, following the
  ledger's own oracle-wrong dispositions. `Apply` selects one of four
  `SD_PATTERNS` tables (`compat.go:51-62`).
- `Recommendation` (`verdict.go:12-19`): `RecNone`, `RecLow`,
  `RecMedium`, `RecStrong`, ordered so `min` reproduces the `max_rec`
  downgrade. `String()` returns beets' own names (`verdict.go:23-36`).
- `Action` (`verdict.go:60-67`): `ActionApply`, `ActionSkip`,
  `ActionAsIs`, `ActionAsk`.
- `Recommend` (`verdict.go:107-130`) and `ActionFor`
  (`verdict.go:179-209`). `ActionFor`'s quiet branch is tested first,
  and I verified it against the pinned source:
  `tools/venv/.venv/lib/python3.13/site-packages/beets/ui/commands/import_/session.py:337-342`
  is `if config["import"]["quiet"]: if rec == Recommendation.strong:
  return APPLY; action = quiet_fallback.as_choice({"skip": SKIP,
  "asis": ASIS})`. `verdict.go:180-191` is that, line for line. This is
  the exact hinge section 7 turns on.
- `Candidate` (`candidate.go:12-16`): `Info *AlbumInfo`, `Dist
  *Distance`, `Assignment`. `Distance.Distance()`
  (`penalty.go:96-102`), `RawDistance()` (`:68`), `NonzeroKeys()`
  (`:166`), and `NonzeroItems()` give a divergence row everything it
  could want to log about the score.

### 5. The estate `Tunables` are producible, but the producer is on the wrong side of an architectural wall

The box's live config, reduced to the verdict-moving subset, is
`corpus/profiles/estate.yaml:16-19`: `quiet_fallback: skip`,
`none_rec_action: skip`, `timid: no`. `original_date` and
`per_disc_numbering` (`estate.yaml:21-22`) are deliberately outside
`Tunables` because they move applied fields and never a verdict
(`tunables.go:35-42`). That profile is a faithful reduction of
`config/beets-config.yaml:13-29`, and the mapping into Go already
exists:

```
gradeTunables(profile) = mergeProfile(DefaultTunables(), profile); tun.Quiet = true
```
`internal/harness/conform/defaults.go:72-79`, with `mergeProfile` at
`internal/harness/conform/translate.go:203-217` and `withImportKeys`
at `:221-226`. The `Quiet = true` overlay lives outside `mergeProfile`
because `quiet` is a CLI flag (`beet import -q`), not a config key
(`defaults.go:64-71`).

The shadow cannot call any of that. `internal/engine/shadow` sits
under `internal/engine`, and `TestEngineNeverImportsHarness`
(`internal/harness/conform/arch_test.go:141`, with
`forbiddenImportPrefix` at `:21`) fails any `internal/engine` file
importing `internal/harness`. So rung 4 either (a) restates the four
assignments in its own package (four lines, and a duplicate that can
drift), (b) lifts a `withImportKeys`-equivalent construction into
`matcher` as a named exported constructor, or (c) puts the shadow
runner outside `internal/engine`, which contradicts the ladder entry's
`engine/shadow`. The spec must rule here. Option (b) is the one that
keeps a single definition; note that `Quiet` and `QuietFallback` are
already `Tunables` fields (`tunables.go:126`, `:131`), so a
`matcher.QuietImportTunables()`-shaped constructor needs no new
surface concepts.

### 6. `internal/config`: the wall, and what a shadow key costs

The wall is four owned rows plus a pending identity row
(`internal/config/keys.go:19-25`): `DUBPLATE_LIBRARY_ROOT`,
`DUBPLATE_STATE_DIR`, `DUBPLATE_SITE_NAME`, `DUBPLATE_ADMIN_EMAIL`.
`Load` (`config.go:159-175`) requires each and then runs
`rejectNamespace` (`config.go:133-152`), which errors on any set
`DUBPLATE_*` name that is not an owned row: reserved names error
naming their rung, superseded names name their replacement, everything
else is "unknown configuration key" (`config.go:150`). This is the U5
compat discipline (`config.go:31-34`).

Four consequences for rung 4, each load-bearing:

1. **A shadow key must be added to `internal/config/keys.go` or the
   binary refuses to start.** There is no third path. Rung 1b (the
   wall-row registry) has not landed, rung 4 is the only live code
   lane, so rung 4 may register there under the existing pattern; the
   spec must say so explicitly, and must name the
   `.gitleaksignore` fragility: the single allowed finding is pinned
   by line number at
   `internal/config/keys_test.go:generic-api-key:66`
   (`.gitleaksignore:1`), and line 66 today is the R2 bucket row of the
   keys table (`internal/config/keys_test.go:66`; the literal is not
   quoted here because gitleaks flags the quote too). Any
   insertion above that line in `keys_test.go` reds the gate for every
   lane at once.
2. **The threshold registry cannot express a duration.** `Threshold`
   is `{Key, Default string, Doc string}` (`thresholds.go:13-17`),
   `Register` appends (`thresholds.go:28-30`), and `ParseInt` is the
   only parser (`thresholds.go:43-59`). `docs/HISTORY.md:683-685`
   records the same gap for rung 3's three float thresholds. Ruling
   1's shadow floor is a duration with a 30-day default, so rung 4
   needs either a `ParseDuration` variant on the registry or a plain
   owned wall row.
3. **`doctor` would misreport a non-int threshold.**
   `thresholdStatus` calls `config.ParseInt` unconditionally
   (`cmd/dubplate-server/doctor.go:100-107`), so a duration registered
   as a `Threshold` renders as `invalid: ...` in the doctor table.
   Adding a parser variant means touching `doctor.go` too.
4. **The box's existing `/etc/dubplate/env` is unusable as the shadow
   unit's `EnvironmentFile`.** It carries roughly forty `DUBPLATE_*`
   names (`env/dubplate.env.template`), of which
   `DUBPLATE_LIBRARY_DIR` (`:71`) and `DUBPLATE_SERVE_SITE_NAME`
   (`:105`) are superseded (`keys.go:165-166`),
   `DUBPLATE_BEETS_CONFIG`/`_DB` (`:55-56`) are retired outright
   (`keys.go:177-178`), and most of the rest are reserved for rungs 5
   through 13 (`keys.go:57-142`). `DUBPLATE_SITE_NAME` and
   `DUBPLATE_LIBRARY_ROOT` are absent entirely. `config.Load` on that
   environment returns a joined error of roughly three dozen
   rejections plus two missing rows. The shadow unit needs its own
   environment file, which is exactly what the ladder entry prices
   ("its environment file"); this inventory confirms the pricing was
   not decorative.

`Config` carries `StateDir` (`config.go:58`), which is where the
ladder puts the durable divergence log.

### 7. `internal/harness/conform`: the row the runner prints, and what a sweep reports

- **The pending-divergence row.** `gradeCaseVerdict`
  (`internal/harness/conform/case.go:452-484`) refuses an
  unadjudicated divergence and prints the exact row a human pastes.
  The literal construction is `case.go:467-477`:

  ```
  analysis.AdjudicationEntry{
      ID: "adj-" + caseID, Subject: analysis.AdjudicationSubject{CaseID: caseID},
      Suspected: "TODO", Citation: "TODO", ReleaseTruth: "TODO",
      Disposition: analysis.Pending, Reason: "TODO",
      Adjudicator: analysis.AgentAdjudicator, Date: "TODO",
  }
  ```
  marshalled with `json.MarshalIndent(row, "", "  ")` (`case.go:478`)
  and embedded in the failure message (`case.go:482-483`). The runner
  writes nothing. That is the shape the shadow's adjudication artifact
  should share, per the 3d handoff (`docs/HISTORY.md:297-303`).
- `CaseGradeSummary` (`case.go:111-125`): `Mode`, `Cases`,
  `Candidates`, `Distances`, `TrackVectors`, `Assignments`,
  `Verdicts`, `TiesExcluded`, `TiesAwaitingRuling []string`, `Pending
  []string`, `Ungraded []string`, `NearMisses []NearMiss`,
  `VerdictMoves []string`. Its doc comment is the discipline a shadow
  report should copy in spirit: "reports what one sweep actually
  compared, never what it visited" (`case.go:109-110`).
- `GradeMode` (`case.go:154-160`): `ModeCompat` is the bit-identity
  run tier 1 means; `ModeDefault` re-scores committed candidate sets
  under the corrected pattern table and does not drive `tag.Album`
  (`case.go:141-153`). A shadow observation is a live flow, so only a
  `ModeCompat`-shaped comparison has an analogue.
- Ledgers a sweep reads (`case.go:249-276`): the tie ledger, the
  proximity ledger, the oracle-bug ledger, and the adjudication
  ledger. `policyFor` (`case.go:890`) resolves per case whether a
  verdict is graded at all.
- `checkRequestSet` (`case.go:491-499`) grades misses and unused
  logged requests, but only under `ModeCompat` against a `stub.Server`.
  The shadow has no committed request log, so it has no analogue.
- `GradeCases(root string, mode GradeMode)` (`case.go:176`) and
  `GradeRequests(root string)` (`internal/harness/conform/request.go:72`)
  are the two entry points.

### 8. `internal/harness/analysis`: the adjudication artifact, and its two gaps

- Dispositions (`internal/harness/analysis/adjudication.go:17-22`):
  `Pending = "pending-adjudication"`, `OracleWrong`, `OracleCorrect`,
  `NoCorrectAnswer`. The doc comment states the launch-gate rule
  directly: "Pending is legal at every gate this package runs (`bash
  scripts/check.sh`); the launch gate, a different gate at a different
  time, refuses it" (`adjudication.go:10-12`).
- `AgentAdjudicator = "agent"` (`adjudication.go:30`), the reserved
  non-human sentinel. `checkAdjudicationEntry` enforces owner-only
  dispositions: a non-`Pending` disposition set by `AgentAdjudicator`
  is an error (`adjudication.go:198-200`).
- `AdjudicationSubject` (`adjudication.go:35-39`): `CaseID`, or
  `Album` plus `ReleaseMBID`. `adjudication.go:32-34` says outright:
  "a corpus case id today, or an album plus release MBID for a live
  shadow divergence rung 4 will write." Zero committed rows use the
  shadow shape, so rung 4 is its first producer.
- `AdjudicationEntry` (`adjudication.go:46-67`): `ID`, `Subject`,
  `Suspected`, `Citation`, `ReleaseTruth`, `Disposition`,
  `CorrectedVerdict` (required at `OracleWrong`, `:201-203`),
  `Reason`, `Adjudicator`, `Date`, `DefectFree *DefectFree`. Five
  prose fields are checked non-empty (`:176-190`).
- `DefectFree` (`adjudication.go:71-76`): `Distance`,
  `Recommendation`, `Action`, `PatternTable`, constrained to
  `LegalPatternTables` (`:96`) and cross-checked by
  `CheckDefectFree` (`:235`). It is corpus-only derived data and has
  no shadow analogue: deriving it needs pinned beets under the
  workstation venv.
- `CheckAdjudications(caseIDs []string, l *AdjudicationLedger)`
  (`adjudication.go:138`). `BySubjectCase()` (`:114`) deliberately
  omits shadow-shaped rows, "because it has no case to join on"
  (`:108-110`).

**Two gaps rung 4 must close.**

1. **`AdjudicationEntry` carries no matcher-commit field and no
   config hash.** Ruling B makes the commit stamp "a hard requirement
   of rung 4's divergence row"
   (`docs/superpowers/specs/2026-09-03-rung-3-go-matcher.md:93-95`),
   and the 3d handoff says it "joins `analysis.AdjudicationEntry`
   beside the existing fields" (`docs/HISTORY.md:283-286`). Nothing in
   `adjudication.go` has it today. Adding it touches
   `checkAdjudicationEntry` (a shadow-subject row should require the
   stamp; a case-subject row should not), and it must be `omitempty`
   so the 21 committed corpus rows still validate byte-for-byte.
2. **A skipped album has no release MBID for `Subject.ReleaseMBID`,
   but the checker demands one.** `adjudication.go:168-171` accepts
   the shadow shape only when `Album` and `ReleaseMBID` are both set,
   and errors when one is set without the other. Under `quiet_fallback:
   skip` the oracle emits a bare path in the tag log and writes
   nothing to the beets DB, so the oracle supplies no MBID at all for
   exactly the half of the stream that most needs adjudication. Rung 4
   either fills `ReleaseMBID` from its own best candidate
   (`Proposal.Candidates[0].Info`), which changes what the field
   means, or relaxes the pair rule for shadow subjects. This is a spec
   ruling, not an implementation choice, because the field's meaning
   is what an owner adjudicates against.

### 9. `internal/harness/cassette`: what a shadow observation would need to become corpus-shaped

- `Case` (`cassette.go:41-56`): `ID`, `Input` (items in pinned
  order), `Profile`, `EvalDate`, `Requests []Request` (the committed
  request log, "an expectation, not scaffolding"), `Extra []Request`,
  `Expected *Expected`.
- `Request` (`cassette.go:69-73`): `Method`, `PathQuery` in the
  canonical sorted-pairs form (rule at `cassette.go:15-24`),
  `Response` naming a file under the case directory. `LoadCase`
  rejects a non-canonical `PathQuery` and proves each response file
  readable (`cassette.go:132-146`, `:151-212`).
- `Expected` (`cassette.go:76-88`): `Recommendation`, `Action`,
  `EchoedDate`, `ConfigHash string` (`cassette.go:80`, "hash of the
  shim's effective config"), `Weights`, `Thresholds`, `Candidates`,
  `Assignment map[int]int`, `ExtraItems`, `ExtraTracks`,
  `AppliedFields`.
- `Candidate` (`cassette.go:103-109`): `Source`, `AlbumID`,
  `Distance`, `Penalties`, `Tracks`. Identity keys only, no display
  strings, which is ruling E's whole point.
- `Manifest` (`cassette.go:112-122`) pins beets, unidecode,
  jellyfish, lap, numpy, mediafile, shim version, and
  `ProfileHashes map[string]string`.

**Whether a shadow observation can become a cassette: the fixture half
yes, the oracle half no.**

- The input half is reachable. `Case.Input.Items` is `map[string]any`
  per item, which the beets DB (or a tag read) can populate directly.
- The request/response half is reachable, but only through a seam
  rung 4 has to build: `tag.Album` returns no request log, so the
  shadow must record at the transport. `ClientConfig.HTTP
  *http.Client` (`cassette.go` aside,
  `internal/engine/musicbrainz/client.go:98`) is the seam: a recording
  `RoundTripper` yields `(method, path_query, body)` triples, which
  `cassette.Request` plus a `responses/NNN.json` file consume
  directly. `internal/harness/stub` proves the shape round-trips:
  `Miss` (`internal/harness/stub/stub.go:44`) and `Server.Consumed()`
  (`:187`) already speak it.
- The `Expected` record half is not reachable on the box. `Expected`
  is what the pinned-beets shim emits; `ConfigHash` is a hash of the
  shim's effective config, and `Weights`/`Thresholds`/`Penalties`/`Tracks`/`Assignment`
  are beets' own internal figures. The box's live beets never emits
  any of it under `beet import -q`. `cmd/harness` is workstation-only
  and never shipped (`cmd/harness/main.go:1-5`), its `capture` verb
  talks only to the live API from the workstation
  (`cmd/harness/capture.go:33-36`), and `freeze` derives through the
  venv.

So the achievable design is: the shadow captures a case-shaped fixture
with `Expected` nil, and rung 15 (or a workstation `harness freeze`)
fills `Expected` later. `Case.Expected` is already `*Expected` with
`omitempty`, documented "nil until derived" (`cassette.go:55`), so the
schema admits this with no change. That is what lets rung 15's re-run
and the adjudication stream share one artifact, and it is the
strongest argument for capturing the responses at all rather than
logging only a verdict.

The `config_hash` a shadow row carries is a different hash from
`Expected.ConfigHash` and must be named differently to avoid a false
join: the shadow's is a hash of the `Tunables`/`Compat` pair the Go
matcher ran under, and the corpus's is a hash of the Python shim's
config. Conflating them would make a rung-15 comparison silently
wrong.

### 10. `cmd/dubplate-server`: the shape a shadow subcommand joins

`main.go:1-19` already reserves the verb: "`shadow`: run the matcher
against an existing library without writing (rung 4)" (`main.go:11`).
The cobra shape is a bare root with `SilenceUsage`/`SilenceErrors` both
true (`main.go:41-54`), commands added in `newRootCmd` (`:48-52`), and
main as the sole stderr writer (`main.go:35-38`, restated at
`check.go:11-15`). Constructors take `(environ []string, stdout,
stderr io.Writer)` and return `*cobra.Command` (`check.go:16`,
`doctor.go:23`), which is the signature `newShadowCmd` should match.
`doctor.go:63-67` pins the `"%-32s %s"` row form a launch-gate report
can reuse.

Deploy is already repointed: `scripts/deploy.sh:251-281`
cross-compiles and rsyncs `dubplate-server-linux-amd64` to
`/usr/local/bin/dubplate-server` with a `version` smoke test
(`:288-289`), and the comment at `:257-260` says plainly "Greenfield
units start exec-ing `/usr/local/bin/dubplate-server` from rung 4."
The gate needs no change for a new unit: `scripts/check.sh:137` globs
`systemd/*.service systemd/*.timer systemd/*.path`, and the
missing-executable filter at `:139` matches
`/usr/local/bin/[A-Za-z0-9_.-]+`, which covers `dubplate-server`.
`systemd/dubplate-import.service:12-55` is the hardening pattern to
copy (`NoNewPrivileges`, `PrivateTmp`, `ProtectSystem=strict`,
explicit `ReadWritePaths`, `EnvironmentFile`), and its
`ReadWritePaths` derivation comment (`:29-53`) is the standard of
proof a reviewer will hold rung 4's unit to.

### 11. The seam, stated exactly

**What the runner must construct.** A `*musicbrainz.Client` from a
`ClientConfig` whose `UserAgent` it composes from `Config.AdminEmail`
and whose `Budget` is `limiter.Budget(name, ClassBackground, limit)`
off a `NewLimiter(DefaultInterval)`. A `matcher.Tunables` equal to
`DefaultTunables()` with `QuietFallback = "skip"`, `NoneRecAction =
"skip"`, `Timid = false`, `Quiet = true`. A `matcher.Compat`. An
`evalDate`. And, per observed album, a `[]matcher.ItemFields`, for
which no producer exists (section 2).

**What it may log.** Everything on `*tag.Proposal` (`tag.go:38-44`):
`Recommendation.String()`, `Action`, `CurArtist`, `CurAlbum`, and per
candidate `Info.AlbumID`/`Info.DataSource`, `Dist.Distance()`,
`Dist.NonzeroItems()`, and `Assignment`. Plus `Budget.Spent()` and
`Budget.Name()` for its own spend. It may not log a request set
without adding a recording transport, because `tag.Album` returns
none.

**The comparison the coarse oracle admits, stated precisely.** Under
the box's configuration (`beet import -q` with
`config/beets-config.yaml:16-18`), `_summary_judgment` collapses to
two outcomes, verified in the pinned source at `session.py:337-342`
and mirrored exactly by `matcher.ActionFor` at `verdict.go:180-191`:

- oracle APPLY exactly when `rec == strong`
- oracle SKIP exactly when `rec` is `medium`, `low`, or `none`

So the shadow's comparable quantity is a binary projection of
`Proposal.Action`, `ActionApply` against `ActionSkip`, and nothing
finer. `Proposal.Recommendation` is one-way observable: `RecStrong` is
exactly attested, and `RecMedium`/`RecLow`/`RecNone` are
indistinguishable to the oracle. `Candidates[k].Dist.Distance()`, the
per-key `Penalties`, the `Tracks` vector, and the `Assignment` have no
oracle counterpart at all for a skipped album; they may be logged as
evidence, never graded. That is the build spec's "the shadow log
records what the oracle can actually attest, tier-one evidence rides
the applied subset, owner adjudication covers the skips"
(`docs/superpowers/specs/2026-09-02-greenfield-server-build.md:255-260`),
read against the landed types.

The two divergence classes are therefore asymmetric, and rung 4's spec
should name them separately:

- **Go applies, oracle skipped.** The oracle gives a path and nothing
  else. `ReleaseMBID` has no oracle value (section 8, gap 2). This is
  the class a false confident match hides in, and it is the launch
  bar's own clause.
- **Go skips, oracle applied.** The beets DB names the release the
  oracle chose, so `ReleaseMBID` is real oracle data and `ItemFields`
  is readable from the DB. This is the tractable class.

Two more branches, neither a divergence: `musicbrainz.ErrUnavailable`
defers the album for a later run (`client.go:78-79`), and
`tag.ErrNoCandidates` records "no match" (`tag.go:78-82`), which
agrees with an oracle skip and diverges from an oracle apply.

**How a row carries the stamp and the hash.**
`analysis.AdjudicationEntry` gains two `omitempty` fields, populated
only on shadow-subject rows: the matcher commit (ruling B's mechanical
re-evaluation key, so a divergence from a superseded revision is
re-evaluated rather than adjudicated) and a config hash over the
`Tunables`+`Compat` pair the run used. The commit is a build-time
input, so it wants the `-ldflags -X` treatment `main.version` already
has (`cmd/dubplate-server/main.go:29-31`), and the shadow must refuse
to run with a `dev` stamp on the box, since an unstamped divergence
cannot be mechanically retired. `checkAdjudicationEntry`
(`adjudication.go:158`) gains the matching requirement, guarded on
the shadow subject shape so the 21 committed corpus rows stay valid.

**What the launch-gate report needs.** Three figures, and ruling 1
makes only the third a gate:

1. **Observed album count**, reported and never gated (ruling 1, via
   `docs/superpowers/research/2026-09-02-greenfield-server-rulings.md:53-56`
   and the build spec at `:249-253`). The count should be split
   applied/skipped, since only the applied subset carries tier-one
   weight.
2. **Divergences by disposition**, over the four values at
   `adjudication.go:17-22`.
3. **The floor duration**, configurable with a 30-day default, against
   the elapsed observation window. The estate's own setting is the
   launch date's dominant term and is an open owner question
   (`docs/superpowers/research/2026-09-03-rung-dependency-analysis.md:511-520`).

The gating clause is that any row still at `pending-adjudication`
blocks the launch gate (`adjudication.go:10-12`; build spec
`:900-902`; rung 15's criterion at `docs/HISTORY.md:669-678`). So the
report's fourth line is the pending count, and it must be zero.
`CaseGradeSummary`'s honesty rule (`case.go:109-110`) is the right
shape to copy: report what was compared, not what was visited.

---

## (c) The ratified record for rung 4

All citations are `path:line` against the tree at HEAD `440131e`. Line
numbers for the build spec are
`docs/superpowers/specs/2026-09-02-greenfield-server-build.md`,
abbreviated BS; the dependency analysis
`docs/superpowers/research/2026-09-03-rung-dependency-analysis.md`,
DA; the parallel-approach spec
`docs/superpowers/specs/2026-09-03-parallel-build-approach.md`, PA; the
rung 3 spec `docs/superpowers/specs/2026-09-03-rung-3-go-matcher.md`,
R3; the reverse-edge completion
`docs/superpowers/research/2026-09-04-reverse-edge-completion.md`, RE.

### 1. What the record settles

Twenty-nine decisions. Each is closed by a ratified document, and a
rung 4 spec that re-opens one is re-litigating settled ground.

#### Deliverables and shape

**S1. The five deliverables, verbatim.** The shadow runner under
`engine/shadow` and its subcommand; the divergence log; the
adjudication artifact; a hardened systemd unit per vps-conventions
with pre-created `ReadWritePaths`; its environment file; and the
`deploy.sh` path rung 1 repointed. `BS:572-580`.

**S2. Band and track.** M to build, continuous to launch, marked `[E]`
as the evidence track's last rung. `BS:572`. The evidence track is
scheduled first because shadow evidence accumulates in calendar time
rather than executor time. `BS:218-225`.

**S3. Rung 4's deploy is merge point M1 and the evidence phase's
boundary.** `DA:328-331`, `PA:75`, `BS:910-912`. Everything after M1
has slack and everything before it does not. `DA:330-331`.

**S4. The shadow runner is a product capability, not scaffolding.** It
is a first-class leaf in the greenfield tree "because it outlives its
rung: it is the migration-shadowing capability any beets-migrating
adopter wants". `BS:335-337`.

**S5. `engine/shadow` is the package.** `BS:315` in the tree listing,
restated at `BS:573` and `RE:84`.

#### The handoff from rung 3

**S6. Ruling B, option (b), ratified.** Rung 4 deploys at functional
completeness rather than waiting for the bar, "with the stamp as a
hard requirement of rung 4's divergence row". `R3:89-95`. STATUS
records it as the standing lane-1 instruction. `docs/STATUS.md:26-29`.
The reasoning is on the record: launch date equals rung 4's deploy
date plus the floor, and "'behind the harness until the bar holds'
governs what the matcher is allowed to decide, and shadow decides
nothing". `R3:95`. The same trade appears as a named re-sequencing
candidate at `DA:507-509` and as risk 1's mitigation at `DA:531-535`.

**S7. The entry point, with its exact signature.**
`tag.Album(ctx, src *musicbrainz.Client, items []matcher.ItemFields,
tun matcher.Tunables, compat matcher.Compat, evalDate time.Time, opts
tag.Options) (*tag.Proposal, error)` in `internal/engine/tag`. `ctx`
first, `evalDate` a parameter and never a clock read, `opts.SearchIDs`
the only optional input. `Proposal` carries `CurArtist`, `CurAlbum`,
`Candidates` (ascending distance, stable), `Recommendation`, and
`Action`. `docs/HISTORY.md:273-279`.

**S8. The budget class and the wire pace.** `musicbrainz.NewClient`
requires a `Budget`, every request `tag.Album` issues spends one
token, and the shadow's class is `ClassBackground`, deferred behind
rung 7's user-present probe. The pace is `DefaultInterval =
time.Second` (3c contract C20). `docs/HISTORY.md:287-293`, confirmed
in code at `internal/engine/musicbrainz/limiter.go:23-25`, where
`ClassBackground`'s own comment names rung 4's shadow. The
one-request-per-second figure is written down with its citation chain
at `docs/HISTORY.md:711-715`, which names rung 4's shadow as "the
first real spender that inherits it".

**S9. The two error branches rung 4 takes.**
`musicbrainz.ErrUnavailable` defers rather than recording a
divergence; `tag.ErrNoCandidates` records "no match".
`docs/HISTORY.md:288-293`.

**S10. Rung 3's inbound edges to rung 4 are both hard and both
discharged.** `2 -> 4` is the known-acceptable-divergence artifact and
the adjudication ledger format (`DA:83-85`). `3 -> 4` is the matcher
plus the shared limiter with its per-caller budget interface
(`DA:96-98`). Rung 3 is closed with all three tiers green.
`docs/STATUS.md:5-13`.

**S11. What rung 3 does not hand over.** No applied fields, no tag
I/O, no pipeline wiring. `R3:247-254`. The consequence is stated at
`RE:186`: "Rung 3 computes no applied field, so the shadow's Go side
has none either and the comparison is verdict-level."

#### The shadow's evidentiary shape

**S12. The floor is a configurable duration, default 30 days.** Every
divergence that arrives is adjudicated; the observed album count is
reported at the launch gate rather than gated; the corpus's real-catalog
and adversarial strata carry the evidentiary weight a thin shadow
cannot. Ruling 1, `BS:804-815`, restated at `BS:250-253` and in the
rulings document's supersession note.

**S13. The estate's own floor value is an ordinary config choice.**
Decision 3 was reframed rather than answered: "the shadow floor is an
ordinary configuration variable, and since the shadow accumulates from
rung 4 through the whole remaining build, any plausible floor is
satisfied before the spine completes". `PA:319`, `PA:327-333`.

**S14. The material is named honestly and is thin.** The wind-down
retired every contributor intake path, so what feeds the interim
import timer is what Geoff adds for his own listening. The window is
long and thin, over owner-curated rather than contributor material.
`BS:243-249`. Risk 1 names the secondary risk of a clean but empty
record and suggests "deliberately feeding the box a listening backlog
during the window". `DA:534-536`.

**S15. The oracle is coarse, and the coarseness is a binary signal.**
The box imports with `beet import -q` and `quiet_fallback: skip`, so
live beets emits strong-or-skip with no recommendation level and no
distance for skipped albums. `BS:254-257`. Verified against pinned
beets 2.13.1: `_summary_judgment` returns `APPLY` only on
`Recommendation.strong` under `import.quiet`, and otherwise resolves
`quiet_fallback` to `SKIP`, at
`tools/venv/.venv/lib/python3.13/site-packages/beets/ui/commands/import_/session.py:326-341`.
The estate sets both `quiet_fallback: skip` and `none_rec_action: skip`
at `config/beets-config.yaml:16-17`.

**S16. The log records only what the oracle can attest.** Tier-one
evidence rides the applied subset, and owner adjudication covers the
skips. `BS:257-260`.

**S17. The comparison runs under the estate config.** "Both verdicts
are logged under the estate config." `BS:240-241`. The estate's beets
config sets no `match:` and no `musicbrainz:` block at all, so the
shadow's oracle side runs the whole distance and threshold surface at
beets' shipped defaults. `config/beets-config.yaml` carries no such
block; `R3:280` states the same fact for `estate.yaml` in the corpus.

**S18. Divergences land in the corpus's own ledger.** "Divergences are
adjudicated by the owner into the same known-acceptable-divergence
artifact the corpus uses." `BS:240-242`. That artifact is
`corpus/analysis/adjudications.json`, and its subject type already
models the shadow's shape: `AdjudicationSubject` carries `Album` and
`ReleaseMBID` as "an album plus release MBID for a live shadow
divergence rung 4 will write", at
`internal/harness/analysis/adjudication.go:32-39`. Rung 4 is its first
producer, with zero committed rows today. `docs/HISTORY.md:280-286`.

**S19. The pending-divergence row format.** A committed row carries
`id`, `subject.case_id` (or, for the shadow, `subject.album` plus
`subject.release_mbid`), `suspected`, `citation`, `release_truth`,
`disposition: pending-adjudication`, `reason`, `adjudicator: agent`,
and `date`, with all five prose fields non-empty. The runner prints
exactly that row and writes nothing. "Rung 4's shadow inherits the
same rule, and the launch gate refuses any pending row."
`docs/HISTORY.md:295-302`.

**S20. `disposition` is owner-only, enforced by the
`AgentAdjudicator` sentinel.**
`internal/harness/analysis/adjudication.go:24-30`.

**S21. Improvements never ride the port.** A shadow-surfaced
improvement is a deliberate divergence, adjudicated and recorded with
written reasoning in the same artifacts the corpus uses. `BS:100-110`.

#### The box and the deployment

**S22. The deployment is additive and read-only with respect to the
library.** `BS:262-263`. The organizer spec's rung 4 says the same:
"nothing it does is irreversible", at
`docs/superpowers/specs/2026-09-01-dubplate-organizer-design.md:572`,
and "Everything here is reversible, and nothing here changes what the
box decides" for the neighbouring rung at `:590-591`.

**S23. Rung 4 is one of only three rungs that touch the box, and its
touch is an additive unit.** "Rungs 1 through 3, 5 through 9, and 11
through 15 touch the box not at all; rung 4 deploys the additive
shadow runner under its own unit." `BS:864-874`. The claim carries a
known erratum for rung 12, recorded at `PA:26` and `PA:105`, which
does not change rung 4's row.

**S24. Rung 4's box window is compatible with rung 10's and
incompatible with nothing in phase 1.** "Rung 4's additive unit and
rung 10's read-only listing are compatible. Rung 12 writes Navidrome
config and must not overlap rung 10's diff window. Rung 16 needs a
quiet box and overlaps neither." `PA:123`, `DA:396-397`.

**S25. Every box write follows the quiet-box discipline and the
`--check`-first deploy rule.** `BS:262-265`, `CLAUDE.md` safety rails,
`PA:299`.

**S26. The deploy path already exists and is additive by
construction.** `scripts/deploy.sh` rsyncs the cross-compiled binary
to `/usr/local/bin/dubplate-server` and never writes the frozen
incumbent `/usr/local/bin/dubplate`, which the four guardian units
still exec. The boundary comment names rung 4 by name: "Greenfield
units start exec-ing /usr/local/bin/dubplate-server from rung 4; the
incumbent binary and its step retire together at rung 16."
`scripts/deploy.sh:251-260`, with the rsync and smoke test at
`:273-289`. Rung 1's ledger records the same at
`docs/HISTORY.md:2077-2081`.

**S27. `env/dubplate.env.template` is rung 4's file, by the file's own
header.** "The greenfield wall is internal/config's own package doc;
this manifest evolves at rung 4." `env/dubplate.env.template:1-3`. The
ownership row is written at `RE:100`, with the hard constraint that
`deploy.sh` fails the whole render on one unresolved name
(`scripts/deploy.sh:190-194`).

**S28. The state-directory layout is rung 4's to own.** "Rung 4 (first
to deploy one, and its unit's `ReadWritePaths` depends on it)", with
rung 5's nonces and per-email epoch as consumers and rung 6's store
superseding neither. `RE:102`. The divergence log is "durable in the
state directory; the store arrives later". `BS:573-575`,
`DA:275-278`. `DUBPLATE_STATE_DIR` is already an owned, required,
no-default wall row at `internal/config/keys.go:21`.

**S29. Any lane open at M1 closes through the merge ritual, and rung
4's plan states that the deployed binary carries landed-but-unwired
packages from other lanes.** `RE:202`, the paste-ready constraint
sentence from cell (8,4).

#### Process

**S30. Rung 4 rides the in-chain `diff-reviewer`.** Only auth (5),
streamer/SSO (12), and cutover (16) keep the full fresh-context
adversarial cycle. `CLAUDE.md` build process, `PA:120`, `PA:288-291`.

**S31. The lane-launch checklist applies unchanged**, all eight items.
`PA:194-204`.

**S32. Rung 4 creates no attended sitting of its own.** The initiative
carries exactly one attended gate (rung 16) and one attended visual
check (rung 7). `BS:894-903`, `DA:299-313`. Rung 4's owner cost is the
standing adjudication stream.

**S33. The observation mechanism is designed in this rung's plan, and
the deployment machinery it needs is priced in rung 4's band.**
`BS:266-268`, restated at `BS:576-578`. This is the record
deliberately leaving a hole, and it is the largest open item below.

### 2. What the record leaves open

Fourteen items, each with whatever the record already names.

**O1. The observation mechanism.** The record defers it explicitly
(`BS:266-268`, S33 above) and supplies no options. What the tree
exposes as candidate inputs: the persisted per-run tag log under
`DUBPLATE_STATE_DIR`, which "records only skips"
(`_salvage/internal/beets/beets.go:57-61`); the beets database, which
`CLAUDE.md` names as the only import success record; and the review
tree, where a skip lands per uploader and date
(`_salvage/internal/importer/importer.go:993`,
`config/beets-config.yaml:6-9`). The import timer's own cadence is
`OnCalendar=*:0/15` with a 300-second jitter
(`systemd/dubplate-import.timer`). No document recommends among these.

**O2. The estate's own floor value.** Options: any duration, since the
floor is an ordinary config variable under the defaults discipline
(`PA:327-333`). Recommendation on the record: batch the question at a
checkpoint rather than discover it at rung 15 (`DA:511-520`, which
calls it the launch date's dominant term). Nothing is pre-committed.

**O3. How the floor knob is typed and where it registers.** The
record settles that every tunable is a threshold with a sane default
(`BS:126`) whose rationale is recorded where the key is defined
(`BS:170-181`). It does not settle the mechanism. `config.Threshold.Default`
is a string with `ParseInt` as the only parser
(`internal/config/thresholds.go:13-40`), and 3c already carried the
same gap for rung 3's three float thresholds: "the three float rec
thresholds need a float variant the registry does not yet have"
(`docs/HISTORY.md:678-685`). A duration floor needs a duration
variant, an int-days spelling, or the same deferral.

**O4. Whether the shadow is an adopter-optional capability or a
build-time instrument.** The record leans toward adopter-optional in
two places and never rules. Ruling 1 dissolves the estate-shaped floor
precisely by reading shadow "through the universal-adopter lens": it
is "the local confidence tool an adopter runs against their own live
beets before trusting the swap, sized by their own material"
(`BS:804-811`). The tree section calls the runner "the
migration-shadowing capability any beets-migrating adopter wants, not
scaffolding" (`BS:335-337`). The config wall's own classification rule
then implies the shape: an integration gets an explicit enable flag,
and the flag makes its correlated config required-when-enabled
(`BS:150-158`). No document names the flag, its class, or its
correlated set. This is the universality decision rung 4's spec must
take.

**O5. How the shadow reaches live beets, given that beets config is a
retired key class.** `internal/config/keys.go:174-175` retires
`DUBPLATE_BEETS_CONFIG` and `DUBPLATE_BEETS_DB` with "retired: the
greenfield server never embeds beets". The shadow must observe what
live beets decided. No document reconciles the two. This is O1's
config half and is unwritten in every source.

**O6. The divergence log's own file format.** The record names two
artifacts and gives only one a home. The adjudication artifact is
`corpus/analysis/adjudications.json` (S18). The divergence log is
"durable in the state directory" (`BS:573-575`) with no schema
anywhere. The nearest thing to a recommendation is the pending-
divergence row the rung 3 runner prints (`docs/HISTORY.md:295-302`),
which the record says the shadow "inherits" for the artifact half
only; whether the log shares that shape is a rung 4 choice.

**O7. Where the matcher-commit stamp lives as a field.** Ruling B
requires the stamp (`R3:93-95`). The 3d handoff says "It joins
`analysis.AdjudicationEntry` beside the existing fields"
(`docs/HISTORY.md:280-286`). Cell (8,4) calls it "a schema addition to
that ledger" and awards rung 4 the row: "rung 4 adds the shadow
divergence rows and the matcher-commit stamp as a schema addition,
never by hand-editing a derived record" (`RE:198`). Unsettled: the
field's name, whether `checkAdjudicationEntry` requires it on a
shadow-subject row, and whether a corpus-subject row is forbidden from
carrying it.

**O8. What re-evaluates a divergence from a superseded matcher
commit.** Ruling B's whole trade rests on divergences being
"re-evaluated mechanically instead of consuming owner adjudication"
(`R3:93`, `DA:534-535`). No document names the mechanism, the rung
that builds it, or when it runs. The launch gate refuses any pending
row (`docs/HISTORY.md:301-302`), so the loop has to close before rung
15.

**O9. The unit's trigger, user, and `ReadWritePaths` set.** The record
requires a hardened unit per vps-conventions with pre-created
`ReadWritePaths` and its environment file (`BS:575-578`), and the
standing rule that a `ReadWritePaths` entry must exist on the host
before the unit starts, with `deploy.sh` pre-creating it (`CLAUDE.md`).
Whether the shadow is a timer, a path unit, or a long-running service
follows from O1 and is unwritten. The incumbent import unit's own
`ReadWritePaths=/srv/music /var/lib/musicbox` is the precedent and is
deliberately wide (`systemd/dubplate-import.service`), which a
hardened additive unit should not copy.

**O10. Whether rung 4 registers wall keys in `keys.go` or waits for
rung 1b.** The contested table awards the config wall to rung 1b's
registry with "no lane edits `keys.go`" (`PA:158`). Rung 1b has not
landed (`docs/STATUS.md:29`, `R3:371`). Rung 3's precedent was to
carry defaults in a typed struct and defer registration to a named
follow-on task (`R3:107`, `docs/HISTORY.md:678-685`); rung 4 has no
equivalent deferral, because its unit needs a rendered environment
file at deploy. The conductor's grant in this pass's framing resolves
it in rung 4's favour as the only live code lane. The spec must say so
and name the `.gitleaksignore` fragility: the single allowed finding
is pinned by line number at
`internal/config/keys_test.go:generic-api-key:66`, so any insertion
above line 66 reds the gate for every lane at once (`PA:54`, `R3:107`).

**O11. The User-Agent contact for the shadow's live client.** 3c
carried `AdminEmail` as an interim contact, with a dedicated contact
row named as the post-1b follow-on beside the four-knob registration
(`docs/HISTORY.md:716-721`). Rung 4 is the first real live spender, so
it inherits the question rather than being handed an answer.

**O12. The shadow's budget cap number.** "Rung 4 caps its own
MusicBrainz spend by the budget it builds the client with"
(`docs/HISTORY.md:287-289`), and the class and pace are fixed (S8).
The cap's value is nowhere.

**O13. Which `env/dubplate.env.template` rows rung 4 writes, and
whether it pre-writes lane A's pending rows.** Cell (5,4) is `SOFT-A`
and names the collision: rung 5's secrets are not minted until rung
16, and one unresolved name aborts the whole render (`RE:88-90`). The
resolution named on the record is the `config.WallRow.Pending`
precedent, a commented pending row rather than a live `NAME=` line
(`RE:90`, `RE:100`), with the standing prohibition "Lane A adding a
live `NAME=` row to `env/dubplate.env.template` for a secret that has
not been minted, at any point before rung 16" (`RE:331-333` region,
the "Forbidden in phase 1" list). Whether rung 4 authors those pending
rows or lane A does is not decided.

**O14. Whether the shadow subcommand and its doctor row register
through a seam or edit two shared call sites.** `RE:22` names it a
general finding: `newRootCmd`'s single `cmd.AddCommand(...)` block at
`cmd/dubplate-server/main.go:48` and `doctor.go`'s `rowStatuses` are
"the `keys.go` failure mode one level up, and Phase 0b does not cover
them". The paste-ready row awards the surface to rung 1's skeleton
with "each rung adds its own file and registers through it"
(`RE:101`), which is an ownership rule rather than a shipped
mechanism. Rung 4 is the first rung after rung 1 to add a subcommand,
so it either builds the registration seam or sets the precedent by
editing the block.

### 3. Rung 4's edges on later rungs

Read off the matrix row 4 at `DA:48`: `4 -> 15` hard, `4 -> 16` hard,
`4 -> 17` soft, and nothing else. Rung 4 has zero edges on rungs 5
through 14, which is the structural reason the spine is designed to be
built inside the shadow window (`DA:315-318`).

**Rung 9 consumes nothing from rung 4, and rung 4 hands rung 9 one
carried question.** The matrix is explicit that `4 -> 9` is none
(`DA:48`), and rung 9 is not gated on the shadow (`DA:316-317`). The
traffic runs the other way. Rung 9 re-verifies the limiter's
caller-priority policy against its real callers (`BS:698-700`,
`R3:63`, `R3:243-245`), and the shadow is one of those callers. 3c's
carry names what rung 9 inherits: `reservationInterval` as the
limiter's one knob, and the reserved token's round-robin over the
classes below the served one, which replaced a "lowest non-empty" rule
that "starved `ClassBatch` outright whenever a probe stream and the
shadow were both queued" (`docs/HISTORY.md:742-748`). Rung 9 reads
four seams, not two, when it re-verifies the dispatcher
(`docs/HISTORY.md:769` region, carry 15). Rung 9 also inherits
`MergePseudoRelease`'s `slices.Clone` fix, owed the day a caller
retains both values (`docs/HISTORY.md:328-332`).

**Rung 15 consumes the shadow record as a launch-bar part.** `4 -> 15`
is both an artifact gate and a calendar gate: the configurable
duration floor must have elapsed, and every arrived divergence must be
adjudicated (`DA:113-115`). Launch bar part 3 is "a clean shadow
record": zero unadjudicated divergences and zero false confident
matches against live beets decisions, over ruling 1's floor
(`BS:69-71`). Rung 15 is where "the conformance gate and shadow record
are evaluated against the launch bar" (`BS:772-777`). Rung 15 also
reports the observed album count rather than gating on it
(`BS:251-253`, `BS:812-815`).

**Rung 16 consumes the same record, and the box.** `4 -> 16` is hard
(`DA:48`, `DA:113-115`). Rung 16 is the one rung that changes what the
box serves and needs a quiet box (`BS:864-874`, `PA:123`), so rung 4's
unit must be stoppable and removable inside rung 16's convergence.
Rung 16 converges systemd, `deploy.sh`, compose, and cloud-init for
the new server (`BS:778-785`), and the incumbent binary's deploy step
"retires together at rung 16, once every unit has moved over"
(`scripts/deploy.sh:257-260`). Rung 16 also carries the exercised
rollback and the deprecation window (`BS:786-793`).

**Rung 17 reconciles the runner's documentation.** `4 -> 17` is soft:
"The shadow runner outlives its rung as the standing
migration-shadowing capability. Close-out reconciles its
documentation." (`DA:211-212`, `BS:794-798`.)

**One inbound edge rung 4 must not break.** Cell (8,4) is `SOFT-B`:
rung 8's `tagio` may land unwired inside the deployed shadow binary,
and rung 4's plan owes one sentence saying so and stating that none of
it is reachable from the shadow subcommand (`RE:190`, `RE:202`).

### 4. Owner touchpoints rung 4 creates

**T1. The standing shadow adjudication stream, beginning at the
deploy.** "Corpus divergences at rungs 2 and 3, shadow divergences
from rung 4 onward, each batch presented at checkpoints rather than as
it occurs. An unadjudicated divergence blocks the launch gate, so
adjudication batches are part of the critical path and are kept small
and regular." `BS:896-903`. `DA:307-311` adds that these batches "do
not gate their own rungs. They gate rung 15." `CLAUDE.md` names them
among the four expected touchpoints. Rung 4 does not add a sitting; it
starts a stream that was already priced.

**T2. The deploy itself as a box write.** Rung 4 is one of three rungs
that touch the box (`BS:864-874`). The write is additive and
read-only with respect to the library (`BS:262-263`), and it goes
through `scripts/deploy.sh` with `--check` first (`CLAUDE.md` safety
rails). Whether Geoff sanctions the deploy explicitly is not written
anywhere. The nearest precedent in the tree is
`scripts/deploy.sh:49-55`, where the uploader's serve unit is
deliberately excluded from the script's timer-enable list and started
"by hand, under Geoff's sanction, never by this script". The record
neither requires nor forbids the same treatment for the shadow unit.

**T3. The floor's start date, which is the deploy date.** "Rung 4
starts the clock at deploy" (`DA:315`). "Launch date equals rung 4's
deploy date plus the floor" (`DA:525`, `PA:303`). Ruling 1 makes the
floor configurable with a 30-day default; the estate's own value is a
batched checkpoint question (`DA:511-520`, O2 above). The deploy date
is therefore a recorded fact rung 15 reads, and it should be written
into the ledger at the deploy rather than reconstructed later.

**T4. The reduction the stamp buys.** Ruling B's whole case is that
stamping every divergence with the matcher commit lets
superseded-revision divergences be "re-evaluated mechanically instead
of consuming owner adjudication" (`DA:534-535`, `R3:93`). The
mechanism is unbuilt (O8), so absent one, rung 4 hands the owner the
adjudication load the ruling promised to avoid.

**T5. One inherited owner ask, already batched.** 3d filed a request
to re-disposition `bug-assignment-rejection-gap` from `oracle-wrong` to
`no-correct-answer` (Resolution R1, `docs/HISTORY.md:388-393`). It
rides the same adjudication stream rung 4 feeds and is not rung 4's to
raise.

### 5. Universality constraints

**U1. No hostname, path, email, or personal name outside
`internal/config`.** `BS:116-120`, `CLAUDE.md`. Everything rung 4
knows about the box arrives through the wall or through the
environment file `deploy.sh` renders from the age store. The interim
estate's paths (`/srv/music`, `/var/lib/musicbox`,
`/srv/music/review/...`) are estate values and must reach the shadow
as config, never as defaults.

**U2. `env/dubplate.env.template` is an instance of the wall, never
its definition.** `BS:167-168`. Inline values in it are not defaults,
and adding a name there requires minting it through `secret-set.sh` and
the registry (`CLAUDE.md` facts section). It is also rung 4's own file
(S27).

**U3. Every tunable is a threshold with a sane, research-backed
default whose rationale is recorded where the key is defined.**
`BS:126`, `BS:170-181`. "A default with no stated basis is a review
finding." The 30-day floor default is ruling 1's, and its basis is the
organizer spec's original window, so the citation exists.

**U4. Integrations get explicit enable flags, and a flag makes its
correlated config required-when-enabled.** `BS:150-158`. Applied to
the shadow this is the O4 decision. The wall's owned set is "the
library root, the state directory, `SiteName`, the admin identity, and
an active identity source" (`BS:141-149`), and the shadow is not in
it, which argues the shadow belongs on the integration side rather
than the owned side.

**U5. Compat discipline (evolution lesson U5).** Unknown or superseded
config keys are load errors that name their replacement, never a
silently degraded start. `BS:159-163`. Rung 4's keys enter the same
machinery `internal/config/keys.go` already implements.

**U6. The design makes sense with no Cloudflare account.**
`BS:127-129`. Nothing in the shadow's shape depends on the tunnel or
on Access.

**U7. The adopter framing that ruling 1 built the floor on.** "What
every user inherits is the corpus: it ships with the product and is
the matcher's standing guarantee. Shadow is the local confidence tool
an adopter runs against their own live beets before trusting the swap,
sized by their own material, which makes a fixed album count estate
idiosyncrasy where a tunable belongs." `BS:806-811`. A rung 4 design
that only works against this box's directory layout contradicts the
ruling that set the floor.

**U8. The estate's ops files are the estate's operational record, not
an adopter kit.** `BS:284-291`, restated in "Out of scope":
adopter-facing deployment artifacts belong to the packaging initiative
(`BS:876-884`). The shadow's systemd unit is therefore the estate's
record. The runner itself is the universal capability (S4). Rung 4
has to keep those two halves separable, which is the concrete form U7
takes in this rung.

### 6. Does the shadow's first material already exist?

STATUS carries "two albums await review in
`/srv/music/review/2026-08-31/geoff/`" as a non-blocking leftover,
grouped with the orphaned `FILEBROWSER_ADMIN_PASSWORD` and the stale
failed units (`docs/STATUS.md:45-47`). No other document in the repo
mentions them; `docs/HISTORY.md` and `ROADMAP.md` carry nothing.

What the tree lets me conclude. Under the estate config, an album
beets declines to tag confidently is skipped and filed into the
review tree per uploader and date. `config/beets-config.yaml:6-9`
states the intent ("quiet_fallback/none_rec_action both 'skip' send
anything beets can't confidently tag to review"), the pinned beets
source confirms the branch (`.../import_/session.py:326-341`), and the
frozen importer implements the filing at
`_salvage/internal/importer/importer.go:993` with the
date-and-uploader path shape `filerFor` builds at `:497-505`. So the
directory name is consistent with two 2026-08-31 skips on Geoff's own
material.

What that is worth to rung 4. It is evidence that the oracle emits
skips on owner-curated material, which is exactly the thin, coarse
signal ruling 1 anticipated (`BS:243-260`). It is not shadow material.
No Go verdict exists for either album, nothing carries a matcher
commit (S6), and neither album has an `AdjudicationSubject` row, since
that ledger holds zero shadow rows today (`docs/HISTORY.md:281-286`).
Whether the files are still on disk cannot be established from the
repo, and the box is read-only for this pass.

The useful reading is that the shadow's first material is not waiting
to be picked up; it has to be generated by the import timer running
after the shadow deploys, or by the deliberate backlog feed risk 1
suggests (`DA:535-536`).

---

## (d) Cross-section consistency and open questions

### Every observation mechanism against the engine seam and the record

Section (a)'s eight candidate mechanisms feed the matcher only through
the seam section (b) describes: `tag.Album` takes
`[]matcher.ItemFields` per track, and nothing in the tree projects that
type off a file or off the beets database today (b.2). Checked
mechanism by mechanism:

- **Mechanisms A and B** (poll the inbox or the staging directory) give
  raw file paths, not `ItemFields`. They only become useful once
  something reads tags off the files, which is b.2's producer choice 2
  (`metaflac`/`beet` shell) or 3 (pull rung 8 forward), never producer
  choice 1 (the beets DB). A also matches (c)'s O1 candidate "the
  review tree" only for the skip half; it says nothing about producer
  mechanics, which is a genuine gap between (a) and (b): (a) treats "read
  the inbox" as sufficient evidence capture, while (b) shows that
  capture still needs a tag reader that does not exist.
- **Mechanism C** (read `library.db`) is exactly (b.2)'s producer
  choice 1, and (a) and (b) agree on its scope: applied albums only,
  confirmed independently by CLAUDE.md's "beets database is the only
  import success record" and by (b.8)'s gap 2 (a skip has no DB row,
  so no `ReleaseMBID`). (c)'s O5 (the retired `DUBPLATE_BEETS_DB` key)
  is the config-wall blocker for this exact mechanism; (a.6) and (b.6)
  both flag it independently, so it is a single question raised three
  times, not three questions.
- **Mechanism D** (tag log) and **mechanism H** (review tree) both
  cover only the skip half, consistent with (b.11)'s statement that a
  skip carries no oracle MBID at all. Neither is a source of
  `ItemFields` on its own; H's "original tags intact" claim (a.4) is
  the only mechanism that can feed the matcher honest input for a skip,
  which makes A-plus-H (not A-plus-E, which (a.4) recommends) the
  combination that actually produces gradable `ItemFields` for skips.
  This is a tension inside section (a) itself: its own recommended pair
  (A plus E) supplies items (A) and a verdict (E) for the *applied*
  half, but for the *skip* half E supplies no items, only H does. The
  spec should notice that A+E is incomplete for the tractable divergence
  class (b.11's "Go skips, oracle applied") and that the review tree
  (H) is required reading for that class, not merely a fallback.
- **Mechanism E** (run report) is the richest verdict source per (a.4)
  and matches (b.7)'s `CaseGradeSummary` honesty discipline in shape,
  but it supplies no `ItemFields` itself; it is a join key
  (`SourceDir`, `library_paths`) into A or C, never a standalone
  producer. Its 30-file/~7.5-hour retention window (a.4, a.3) is a
  hard constraint the shadow's poll cadence must respect regardless of
  which mechanism supplies items.
- **Mechanisms F and G** are declined in (a) on operational grounds
  (noise, and violating the additive-deploy boundary respectively).
  Nothing in (b) or (c) argues for either, so the declines stand
  uncontested.

### Every open decision in (c) checked against a mechanism or a seam fact

- **O1** (observation mechanism) is the union of (a section 4)'s
  mechanism table and (b section 2)'s producer-choice table; neither
  document resolves it, and both frame it as the same underlying
  choice from different ends (box-side artifact vs. engine-side input
  type).
- **O2** (floor value) and **O12** (budget cap number) have no
  mechanism or seam fact bearing on them at all; both are purely owner
  calls, with (c) itself naming O2 as a batched checkpoint question.
- **O3** (floor knob typing) is a pure config-wall fact: (b.6) and (a)
  agree the threshold registry's `ParseInt`-only parser cannot express
  a duration, independent of which observation mechanism is chosen.
- **O4** (adopter-optional vs. build-time) has no mechanism bearing on
  it; it is a product-shape ruling that (c) flags as unresolved and
  that (a)/(b) never touch, since neither describes a flag or its
  correlated config.
- **O5** (beets key reconciliation) is a seam fact from (b.6) and
  (a.6): the same superseded-key finding, raised from two sides,
  channeled into mechanism C above.
- **O6** (divergence log schema) is unaddressed by (a) or (b) beyond
  naming a durable-state candidate directory (a.8); (b.9) supplies the
  one candidate shape (the cassette `Case` with `Expected` nil) that
  neither (a) nor (c) had proposed, which is new information (c)'s O6
  should absorb rather than treat as still-blank.
- **O7** (matcher-commit stamp field) and **O8** (re-evaluation
  mechanism) are seam facts from (b.8) and (b.11) respectively: (b)
  identifies exactly what is missing on `AdjudicationEntry` and that no
  re-evaluator exists anywhere in the tree; (c) names the same gaps
  without adding a candidate mechanism, so both remain genuinely open,
  O8 more so since no section proposes even a shape for it.
- **O9** (unit trigger/`ReadWritePaths`) depends directly on O1: (a.5)
  states the unit's `ReadWritePaths` are a function of the observation
  mechanism, so O9 cannot be settled ahead of O1.
- **O10** (keys.go registration timing) is effectively pre-resolved:
  (a.6) and (b.6) both independently conclude rung 4 registers in
  `keys.go` directly, given rung 1b has not landed and rung 4 is the
  only live code lane. (c)'s O10 should be read as confirmed rather
  than open, modulo the spec saying so explicitly.
- **O11** (User-Agent contact) is a seam fact from (b.3): the
  composition function is specified in a doc comment and built
  nowhere, and rung 4 is the first live caller, so it inherits
  ownership by elimination rather than by any document assigning it.
- **O13** (env template pending rows) is a seam fact from (a.6): the
  render-loop-fails-on-one-unresolved-name behavior is what makes this
  a real blocker, not just a style question.
- **O14** (subcommand/doctor registration seam) is a seam fact from
  (b.10): the single `AddCommand` block and `doctor.go`'s
  `rowStatuses` are named as shared call sites with no registration
  mechanism, confirmed independently by (c)'s own citation of the same
  two files.

### The combined, deduplicated open-question list

1. **Which observation mechanism (or combination) feeds the shadow**,
   given that no single one of (a)'s eight candidates is sufficient
   and (b)'s `ItemFields` producer choice is the same decision from the
   engine side. Raised in (a), (b), (c) (O1)
2. **How the shadow reaches the beets database path**, given
   `DUBPLATE_BEETS_DB` is retired outright and "the greenfield server
   never embeds beets." Raised in (a), (b), (c) (O5)
3. **Does the shadow capture item fields before the beet import**,
   given `move: yes`/`write: yes` destroys the applied half's original
   tags, and what quiescence rule such a capture would use. Raised
   in (a), touched in (b.2)'s producer-choice discussion
4. **Which tag-I/O producer rung 4 builds**: read the beets DB
   (applied only), shell to `metaflac`/`beet` (covers skips, adds a
   subprocess to a hardened unit), or pull rung 8's tagio forward
   (out of band). Raised in (b), overlapping (a)'s O1/mechanism
   choice
5. **Does the A-plus-H combination (not A-plus-E) supply the honest
   input needed to grade the tractable "Go skips, oracle applied"
   divergence class**, since E supplies no items for a skip. Raised
   in this section (d), synthesized from (a) and (b.11)
6. **How a skipped album's `ReleaseMBID` is filled**, given the oracle
   supplies none and `checkAdjudicationEntry` requires the pair set
   together: from the Go matcher's own best candidate, by relaxing the
   pair rule, or by a third subject shape. Raised in (b), (c)
7. **Which `matcher.Compat` the shadow runs**: `DefaultCompat()`
   (matches adjudicated dispositions, diverges from what the oracle
   does) or `FullCompat()` (matches the oracle, ships a configuration
   the product will not). Raised in (b) only
8. **How the estate `Tunables` construction reaches `engine/shadow`**
   without importing `internal/harness`: restate the four assignments,
   lift a named constructor into `matcher`, or move the runner outside
   `internal/engine` (ruled out by the ladder's `engine/shadow`
   naming). Raised in (b)
9. **How the floor duration is typed and registered**: a
   `ParseDuration` variant on the threshold registry (touching
   `doctor.go` too), an int-days spelling, or a plain owned wall row,
   and whether the variant serves the whole build or only rung 4.
   Raised in (b), (c) (O3)
10. **The estate's own floor value.** Raised in (c) (O2), an owner
    call with no mechanism bearing on it
11. **The shadow's MusicBrainz budget cap number.** Raised in (c)
    (O12), an owner call
12. **Whether the shadow is an adopter-optional capability behind an
    enable flag, or an owned, on-by-default capability.** Raised in
    (c) (O4)
13. **The divergence log's on-disk schema**, and whether it shares the
    pending-divergence row's shape or is a distinct cassette-style
    fixture with `Expected` nil. Raised in (c) (O6), with (b.9)
    supplying the candidate shape (c) lacked
14. **Whether the shadow captures a recording `RoundTripper`'s
    request/response log for a rung-15 cassette handoff**, or logs
    verdicts only. Raised in (b) only
15. **How the shadow's config hash is named distinctly from
    `cassette.Expected.ConfigHash`**, to avoid a false join at rung 15.
    Raised in (b) only
16. **Where the matcher-commit stamp field lives on
    `AdjudicationEntry`**, its name, and whether `checkAdjudicationEntry`
    requires it only on shadow-subject rows. Raised in (b), (c) (O7)
17. **Whether the shadow refuses to run on the box with an unstamped
    or `dev` matcher-commit build**, given an unstamped divergence
    cannot be mechanically retired. Raised in (b) only
18. **What mechanically re-evaluates a divergence from a superseded
    matcher commit**, and which rung builds it, since ruling B's
    adjudication-load reduction depends on this existing. Raised in
    (b), (c) (O8)
19. **The shadow unit's trigger, user, and `ReadWritePaths` set**,
    which follows from question 1's mechanism choice. Raised in (a),
    (c) (O9)
20. **Is the shadow's timer enabled by deploy.sh, or held back for a
    sanctioned hand start** (the `dubplate-serve.service` precedent)?
    Raised in (a), touched in (c)'s T2
21. **Does the shadow get its own ping-check slug**
    (`DUBPLATE_PING_SHADOW_URL`), given the whole `DUBPLATE_PING_*`
    class is reserved for rung 9. Raised in (a) only
22. **Where the greenfield `DUBPLATE_STATE_DIR` points on the box**:
    shared `/var/lib/musicbox`, or a new `/var/lib/dubplate` that
    retires with the incumbent at rung 16. Raised in (a) only
23. **Which `env/dubplate.env.template` rows rung 4 writes, and
    whether it authors lane A's pending secret rows** or leaves that to
    lane A. Raised in (a), (c) (O13)
24. **Does deploy.sh gain a parameterized render/install helper for a
    second environment file, and do the shadow's three non-secret
    owned values go into the age store or a new literal-default
    column?** Raised in (a) only, a sub-question of question 1's
    environment-file consequence
25. **Whether the shadow subcommand and its doctor row register
    through a seam, or edit the two shared call sites directly** (the
    `AddCommand` block, `doctor.go`'s `rowStatuses`). Raised in (b),
    (c) (O14)
26. **Building the User-Agent composition function**
    (`dubplate-server/<version> ( <contact> )`), specified in a doc
    comment and implemented nowhere; confirm it lands in rung 4 rather
    than the post-1b contact-row follow-on. Raised in (b), (c) (O11)
27. **Whether the shadow's first deploy is a conductor action under
    the standing grant, or an owner touchpoint the build spec has not
    counted.** Raised in (c) only (T2)
28. **Whether rung 4's close records the deploy date in HISTORY and
    STATUS as a first-class figure**, alongside the running
    observed-album count, since it is the launch date's dominant term.
    Raised in (c) only (T3)
29. **Whether the shadow's first deploy should close the still-carried
    SELinux verification on `/var/lib/musicbox`**, alongside the check
    owed on any new `/var/lib` directory it creates. Raised in (a)
    only
30. **Whether `env/dubplate.env.template`'s `TriggerScan` description
    (native Navidrome admin API) matches what `_salvage/internal/beets/beets.go`
    actually builds (a Subsonic `/rest/startScan` request)**, and
    should be reconciled while rung 4 is editing that file. Raised
    in (a) only, immaterial to the shadow itself
31. **Whether `/usr/local/bin/dubplate-server` already exists on the
    box.** Unverifiable from the repo; box access is out of scope for
    this pass. Raised in (a) only

Thirty-one distinct questions survive deduplication. Ten of them
(O2, O4, O8, O12, T2, T3, and five of the (a)-only box-operational
items: ping slug, state-dir location, TriggerScan discrepancy, SELinux
closure, binary-existence) carry no candidate mechanism at all in any
section and are pure owner or design-ruling calls. The remainder each
have at least a partial mechanism or seam fact on the record, named in
the reconciliation above, even where no section resolves them.
