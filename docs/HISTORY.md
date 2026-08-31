# musicbox history

Per-pass ledger, newest first.

## Spec gaps closed: rclone/msmtp provisioning, Subsonic scan trigger (2026-08-31)

Closed the two real spec gaps the first live deploy exposed (previous entry below),
without touching the live box.

- **`rclone`/`msmtp` were never in `cloud-init/user-data.yaml` at all.** Confirmed both
  are in EPEL for AlmaLinux 10 via a container check
  (`podman run --rm almalinux:10 ...dnf -q list rclone msmtp`) before adding them,
  rather than guessing an unattended `set -e` first-boot script into a package that
  doesn't exist: `rclone.x86_64 1.74.3-1.el10_2` and `msmtp.x86_64 1.8.25-6.el10_1`,
  both from the `epel` repo alone, no RPM Fusion or CRB dependency. Installed
  immediately after `epel-release` in the existing EPEL/RPM-Fusion runcmd block,
  ahead of the RPM Fusion key import flac/ffmpeg need, so a failed RPM Fusion step
  can't take them down too.
- **`trigger_navidrome_scan()` rewritten from the dead native `/api/scanner/scan`
  endpoint to the Subsonic API's `GET /rest/startScan`** (query-param auth:
  `u`/`t`/`s`/`v`/`c`, `t=md5(password+salt)`, a fresh random `s` per call). The
  credential-bearing params never touch argv: the full URL is written to a curl
  `--config` file (mode 0600) in a new `RUNTIME_DIR` (same directory as the import
  lock file -- `/run/musicbox` on the box, the fixture's `WORK_DIR` in tests), and
  the config file is removed via a function-scoped `RETURN` trap. Salt is generated
  via `od -An -N8 -tx1 /dev/urandom | tr -d ' \n'` rather than a `tr | head`
  pipeline, which would risk a SIGPIPE-driven pipefail failure. The existing
  failure-tolerant shape is unchanged: a failed request logs to stderr and returns
  0, never killing the run (the daily Navidrome cron scan is still the backstop).
  `tests/music-import.bats`'s shared curl stub now also captures `--config` file
  contents to a separate log (the real file is gone by the time a test could
  inspect it otherwise), plus a `CURL_FAIL_STARTSCAN` switch to exercise the
  failure path; two new tests cover the call shape and the failure tolerance.

Neither fix has reached the already-provisioned live box yet (cloud-init is
first-boot-only); see STATUS's "reprovision pending" open item.

## Task 4/6: first live deploy (2026-08-31)

`scripts/deploy.sh --check` then a real run against the provisioned box. Two env-template
vars didn't resolve as assumed going in: `NAVIDROME_ADMIN_PASSWORD` was stored under the
older name `NAVIDROME_ADMIN_PASS` (renamed in the age store to match, same value, no
rotation), and `MUSICBOX_HC_NAVIDROME_URL` had never been minted at all. Minted it via
`~/Projects/pings/scripts/add-check musicbox-navidrome --period 60 --grace 15` (mirroring
`musicbox-disk`'s already-accepted pre-minted-but-unwired state) so the render gate could
pass; it sits armed and unpinged until Task 6 builds the Navidrome-heartbeat script.

The deploy itself succeeded, but the acceptance sweep surfaced real bugs a fixture-backed
test suite can't see — **wrong to rediscover**, each one below:

1. **FileBrowser crash-looped**: `cacheDir` defaults to a `tmp` folder relative to the
   process's working directory, which isn't part of any bind mount and isn't writable by
   UID 2000. Fixed by pointing `server.cacheDir` at the already-writable, already-persistent
   `/home/filebrowser/data/cache` (`config/filebrowser.yaml`).
2. **FileBrowser's own first-run admin bootstrap then failed**: its built-in default admin
   password is shorter than our own `auth.methods.password.minLength: 8` policy, so account
   creation errored and the container kept crash-looping. Minted a new secret
   `FILEBROWSER_ADMIN_PASSWORD` (FileBrowser's native env var name) and wired it into
   `compose.yaml`'s `filebrowser` service; Task 8 replaces this bootstrap account with
   per-contributor ones.
3. **music-import.service failed on `/run/lock/musicbox-import.lock: Permission denied`**:
   `/run/lock` is root:root `0755` on Alma (unlike Debian's world-writable `1777`), so the
   `User=music` service couldn't create its flock file there. Added
   `RuntimeDirectory=musicbox` to the unit (systemd creates `/run/musicbox` owned by
   `music:music` every boot) and pointed the script's lock-file default there.
4. **music-import then failed on `curl: (7) ... Failed to connect to localhost port 4533`**:
   `trigger_navidrome_scan()` runs on the host as `music`, not inside the compose network,
   but `compose.yaml`'s `navidrome` service only used `expose` (container-network-only).
   Switched to `ports: - "127.0.0.1:4533:4533"` — loopback-only, so Navidrome stays exactly
   as unreachable from the internet as before (still only `cloudflared` reaches it
   externally), while becoming reachable from the box itself.
5. **Same call's error handling was also incomplete**: the login `curl` had no failure
   guard, so `set -e`/`pipefail` killed the *entire* import run on any Navidrome-unreachable
   condition, not just the graceful "skip the scan trigger" path the function's comments
   already described for a failed login. Added `|| true` to match.
6. **The native admin scan-trigger endpoint doesn't exist in the pinned Navidrome
   `0.63.2`**: confirmed against upstream source (`server/subsonic/library_scanning.go`,
   `server/subsonic/api.go` line ~181) — `StartScan` is registered as the *Subsonic* API's
   admin-only `startScan` action (query-param auth: `u`/`p`/`v`/`c`), not a native
   `/api/scanner/scan` JWT-bearer endpoint. Every `music-import` run now logs a 404 there
   after a real login success. **Not fixed this pass** — left as a known, non-fatal bug
   (Navidrome's own `ND_SCANNER_SCHEDULE: "0 3 * * *"` cron backstop still keeps the library
   current) rather than rush an unverified Subsonic-auth rewrite. Follow-up: rewrite
   `trigger_navidrome_scan()` to call `GET /rest/startScan` with Subsonic auth params.

**Two gaps could not be closed this pass, both blocked by the Claude Code auto-mode
classifier refusing an on-box `dnf install` from a subagent** (same shape as Task 5's
DELETE/create-tunnel refusal above — mutating infra calls need Geoff present, not a
subagent acting alone):

- **`ffmpeg`/`flac` are still absent on this box.** The EPEL + RPM Fusion `runcmd` block
  that installs them (`cloud-init/user-data.yaml`) was added to the repo *after* this box
  was provisioned (Task 4a), so it never ran here — a cloud-init-only fix never converges
  an already-running box. beets' `replaygain` plugin fails to load on every `music-import`
  invocation as a result (logged, non-fatal against an empty inbox; will break real
  imports needing the quality gate's `flac -t`/`metaflac` and replaygain analysis).
- **`music-backup.service` cannot run at all**: `sqlite3` and `rclone` are both missing on
  the box, and neither was ever listed in `cloud-init/user-data.yaml` in the first place (a
  real spec gap, not just an unapplied one). Added `sqlite` to the `packages:` list (base
  AlmaLinux repo, confirmed by lookup) for future re-provisions; `rclone` and `msmtp` (also
  missing, also never listed — `lib.sh`'s `notify()` needs it, untested this pass since an
  empty-inbox run never reaches it) are left as an open item rather than guessed into an
  unattended `set -e` first-boot script without a confirmed AlmaLinux 10 install path.
  Confirmed via a read-only `rclone lsjson` against `music-library` that `state/` is empty,
  consistent with no backup ever having run.

Acceptance sweep otherwise green: all three containers running with zero restarts after a
60s settle; the tunnel reads `status: healthy` with active connections; `music.907.life`
serves Navidrome's web UI with no Access interstitial and answers `/rest/ping`;
`inbox.907.life` presents the Cloudflare Access challenge before any FileBrowser content;
both timers active; the Navidrome admin account was created via `POST /auth/createAdmin`
and its login verified via `POST /auth/login` (API only, no secret values ever printed —
checked by name, length, and boolean fields throughout this whole pass).

## Task 5: Cloudflare tunnel wiring, plus a same-day secret rotation (2026-08-31)

Geoff added Cloudflare Tunnel:Edit to `CLOUDFLARE_API_TOKEN` 2026-08-30, unblocking Task 5.
An agent created tunnel `musicbox` (`POST .../cfd_tunnel`), then displayed that call's full
JSON response in a transcript while reporting progress. **Wrong to rediscover**: a
`cfd_tunnel` CREATE response body carries the tunnel's `TunnelSecret` and its connector
`token` alongside the id — displaying or logging the raw body exposes both, not just an
identifier. The fix at the call site is mechanical: always pipe the response through
`jq -r '.result.id'` (or whichever single field is actually needed) and never let the full
body reach a transcript, a log, or a report.

Once flagged, the tunnel was treated as compromised under standing incident doctrine (a
secret printed into a transcript is compromised, full stop — no risk-assessment exception).
Remediation was delete-and-recreate, the cheap definitive path since `cloudflared` was not
yet deployed anywhere against the exposed token:

1. Deleted tunnel `66fa450c-5abb-459d-a1f8-b80d8d19e07c` (`DELETE .../cfd_tunnel/{id}`).
2. Created a new tunnel under the same name `musicbox`, this time piping the response
   through `jq -r '.result.id'` only: id `b6781175-000b-40c3-a54a-05906ab25037`.
3. Fetched its token (`GET .../cfd_tunnel/{id}/token`) and piped it directly into
   `secret-set.sh MUSICBOX_TUNNEL_TOKEN --stdin`, overwriting the old value. Verified
   name-only (`grep -c` on `~/.local/secrets`, count 1) and by length only (240 chars) —
   never printed.
4. Re-PUT the ingress from `config/cloudflared-tunnel-config.json` to the new tunnel id and
   verified with GET: `music.907.life` → `http://navidrome:4533`, `inbox.907.life` →
   `http://filebrowser:80`, 404 catch-all.
5. Repointed both DNS CNAMEs (`music.907.life`, `inbox.907.life`, zone `907.life`) to
   `b6781175-000b-40c3-a54a-05906ab25037.cfargotunnel.com`, proxied.

**Process note**: the DELETE and the tunnel-recreate POST were both refused by the Claude
Code auto-mode classifier when a subagent attempted them (destructive/mutating infra calls
without the user present); read-only GETs against the same account worked fine throughout.
The rotation completed once the conductor ran those two calls from the main session with
Geoff present, then handed the new tunnel id back for the remaining wiring and docs.

Cloudflare Access was untouched throughout (confirmed read-only via `CF_ZT_TOKEN`:
`inbox.907.life` is the only Access app on the musicbox surface; `music.907.life` stays
Access-free). `cloudflared` is not deployed on the box yet, so both hostnames 530 until
Task 4's first real `deploy.sh` — expected, not chased.

## Task 2 follow-up: quality gate, ping hardening, config sync (2026-08-30)

TDD, red-first for every case, against the spec's freshly amended import-pipeline
section (step 4). Four items, one commit each:

- **`quality_gate <dir>`** (`scripts/lib.sh`): walks every audio file under dir
  (case-insensitive flac/mp3/m4a/ogg/opus/wav/aiff extensions; anything else is out
  of scope). A non-FLAC extension fails outright; a FLAC must pass `flac -t`'s full
  decode test and carry >=44.1kHz/16-bit (`metaflac --show-sample-rate
  --show-bps`). Prints "path\treason" per failure, exit 0 only if everything
  passed. Wired into `music-import` as `quarantine_quality_failures`, between the
  free-space abort and `beet import`: any failing file takes its **whole**
  containing album directory to `review/<date>/<user>/` (a partial import of just
  the good tracks never happens), never just the offending file.
- **Uploader copy is templated, not generated**: `config/reject-email.txt` is
  Geoff's verbatim consumer-facing copy with three literal placeholders (`{first
  name}`, `{album or file name}`, `{one reason line per file, from the mapping
  below}`), filled by plain bash substring replacement in
  `render_reject_email()` — no sed, no metacharacter risk. `friendly_reason()`
  translates `quality_gate`'s raw technical reason into one of three fixed human
  sentences (non-FLAC extension, failed decode, below CD quality); Geoff's own
  copy of the notice stays technical (raw path + reason), same asymmetry as the
  existing beets-skip notice. First name comes from `uploader_name()`, the first
  word of the matched `contributors.yaml` `name` field (mirrors `uploader_email`'s
  lookup). Default template path is `/opt/musicbox/config/reject-email.txt` —
  `config/` is already rsynced wholesale to that tree by `deploy.sh`, so this
  needed no new deploy step; only `contributors.yaml` and `beets-config.yaml` get
  dedicated ones, for permissions/install-location reasons this file doesn't
  share.
- **`ping_hc` hardened**: `--retry 3 --max-time 10` added to its `curl` call, so a
  slow or flaky healthchecks.io can never be the reason an otherwise-successful
  run hangs or gets misreported. The empty-vs-unset URL guard already existed;
  only a new bats case (env var set but empty) was added to pin it down.
- **Flock-held early exit now pings success before `exit 0`** (spec decision):
  a run that finds the lock already held (another scheduled run, or an
  interactive `music-beet` session) is the healthy case this check exists for,
  not a failure — healthchecks.io has no third state between success and
  failure. Accepted residual: a genuinely wedged import would also report
  healthy via this path until someone notices stale music in the library.
- **Beets config synced to the workstation's** (`~/.dotfiles/beets/.config/beets/config.yaml`):
  `original_date: yes`, `per_disc_numbering: yes`, and the `replaygain` plugin
  (backend `ffmpeg`), added to `config/beets-config.yaml` alongside the existing
  quiet-mode section. flac (the quality gate) and ffmpeg (replaygain's analysis
  backend) are installed on the host via a dedicated `cloud-init/user-data.yaml`
  runcmd step, not the top-level `packages:` list (that module runs before
  runcmd, too early to enable a repo first): `dnf -y install epel-release`,
  `dnf config-manager --set-enabled crb` (both EPEL and RPM Fusion need it),
  then the RPM Fusion free EL10 GPG key is fetched and `rpm --import`ed and
  its release package installed against it (gpgcheck on throughout, no
  `--nogpgcheck`), then `dnf install -y flac ffmpeg`. Ruling: both binaries
  parse untrusted contributor uploads, so they stay on dnf-automatic's
  security-update path rather than a pinned static build, which would freeze
  a codec-parsing attack surface instead of patching it.
- **Known limitation, accepted**: `quality_gate`'s output channel is
  tab/newline-delimited ("path\treason" per line), so a filename containing a
  literal tab or newline would misparse downstream in
  `quarantine_quality_failures`. The `find -type f -print0` walk itself is
  correct (NUL-delimited, handles any filename); the limitation is only in
  what quality_gate prints afterward. Not expected in practice for
  contributor uploads; revisit if it ever is.

Fixed a docs nit found while touching STATUS.md: Task 4a's on-box checks table
(above, Task 4a's own section) lists nine checks but was labeled "8 of 8"; now "9
of 9" in both STATUS.md and here.

Full test suite: `tests/lib.bats` (4 new `quality_gate` cases + 1 `ping_hc` case)
and `tests/music-import.bats` (1 new flock-ping case, 1 new quality-gate
end-to-end case; three existing fixtures switched from placeholder `.flac` text to
real ffmpeg+flac-encoded audio, since the gate now actually decodes them).
`bash scripts/check.sh` exits 0.

## Task 2 follow-up, fix round: Opus review of 5da3844..d43972b (2026-08-30)

Six items landed, TDD where a test was named:

- **Blocking: unquoted `&` in `render_reject_email`'s `${var//pat/repl}`
  substitutions.** Confirmed on this box's bash (5.3.9): an unquoted
  replacement containing a literal `&` gets expanded to the matched pattern
  text, exactly like sed — an album named "Sturm & Drang" rendered as "Sturm
  {album or file name} Drang" instead of the real name. Fixed by quoting
  every replacement value; bats case covers the ampersand literally.
- **Blocking: an `mv` into an already-occupied `review/<date>/<user>/<album>`
  aborted the whole run under `set -e`** (a same-day repeat quarantine of an
  identically named album is the common case, not an edge case). New
  `lib.sh` helpers: `unique_dest_path` (appends a `-HHMMSS` suffix on
  collision, then `.2`/`.3`/... if even that collides) and
  `safe_mv_to_review` (uses it, and warns-and-returns-nonzero instead of
  letting a failed move take the run down). Applied to both
  `quarantine_quality_failures` and `move_skips_and_notify`, which shared the
  same unguarded-`mv` shape. bats case: two identically named "Bad Album"
  quarantines in one test run both survive and beet import still runs both
  times.
- **A missing `reject-email.txt` template previously sent the uploader an
  empty-bodied email** (`render_reject_email`'s failure was silently
  swallowed by command substitution). Now an explicit `[ -f ... ]` check
  before the `notify` call skips the uploader-facing email and logs why;
  Geoff's technical summary still covers the album either way.
- **Each quality-gate reason line is now prefixed with its file's
  basename** in the uploader-facing email, so an album with more than one
  failing file reads as N per-file lines instead of running together.
- **Added the 8-bit-FLAC branch's missing test coverage** (`ffmpeg -acodec
  pcm_u8` piped into `flac`), the one `quality_gate` branch that had no
  case before this round.
- **CRB enablement** added to cloud-init's EPEL/RPM Fusion runcmd block
  (`dnf config-manager --set-enabled crb`, right after `epel-release`):
  both repos need packages CRB carries, and ffmpeg's dependency chain fails
  without it.

Follow-up verification round found one more leak in the `mv`-collision fix:
`safe_mv_to_review` failing for a **non-collision** reason (permissions, a
bad mount) still `continue`d out of the loop with no Geoff notice and the
album still sitting in staging — exactly what `beet import`'s unconditional
recursive scan of `$STAGING_DIR` would then ingest, silently defeating the
whole quality gate. Fixed with a second `safe_mv_to_review` attempt to
`$MUSICBOX_ROOT/quarantine-failed/<date>/<user>`, a holding dir outside both
staging and review's own tree (so a problem isolated to `review_root`
doesn't also block the fallback); Geoff's summary gets a `could not move;
excluded from import` marker either way, and a distinct all-caps marker for
the (untested, believed-unreachable without a fully read-only filesystem)
case where even the fallback fails. The holding dir's root is a top-level
constant (`QUARANTINE_FAILED_DIR`) beside `REVIEW_DIR` and the rest of the
tree, deliberately not created up front: a healthy box never has one.

Also fixed in `safe_mv_to_review` itself: its `mkdir -p "$dest_dir"` call
was unguarded, so a destination that couldn't even be created would have
aborted the whole run under `set -e` — the exact failure mode the function
exists to prevent. Added a matching
`[ -n "$geoff_summary" ] || return 0` guard before the final Geoff `notify`,
defense-in-depth alongside the pre-loop `album_order` check (every branch of
the loop already appends something, so this is currently unreachable, but
decouples "should Geoff be emailed" from "was there at least one album_key"
for whatever a future change to the loop might do). TDD: bats case forces
the primary move to fail via `chmod 0555` on the review dir and asserts the
album lands in `quarantine-failed`, never in staging, with the marker in
Geoff's email and no uploader email.

Also fixed, while adding coverage for the above: three existing
`tests/music-import.bats` assertions used `! grep ...` as a non-final line —
under bats' `set -e`, a command negated with a leading `!` never triggers
errexit regardless of its actual result, so a failing negated assertion that
isn't the test's last line silently does nothing. Converted to bats' `run` +
`[ "$status" -ne 0 ]` idiom throughout the file for safety, whether or not a
given instance was actually latent.

## Task 5/6: Cloudflare Access + Navidrome secrets, tunnel blocked (2026-08-30)

Control-plane pass, no box changes. Findings and what landed:

- **Tunnel creation blocked, not a code problem**: `POST
  /accounts/{a}/cfd_tunnel` 403s `{"code":10000,"message":"Authentication
  error"}` on both `CLOUDFLARE_API_TOKEN` and `CF_ZT_TOKEN` — neither carries the
  Cloudflare Tunnel:Edit permission group (confirmed: both tokens list/read
  tunnels fine, i.e. GET `.../cfd_tunnel` is 200; only the POST 403s). This
  blocks Task 5's steps 1-4 (create tunnel, fetch token, PUT ingress config, DNS
  CNAMEs to `<tunnel-id>.cfargotunnel.com`) and, downstream, Task 4's first real
  `deploy.sh` run (cloudflared needs `MUSICBOX_TUNNEL_TOKEN` to come up).
  Resolution needs Geoff to add Cloudflare Tunnel:Edit to one token via the
  dashboard (API self-management of token scopes is deliberately locked out).
  Detail in registry.md's `CF_ZT_TOKEN` entry.
- **Cloudflare Access, done**: confirmed no Access application exists on any
  `907.life` hostname (checked the full account app list — zero matches), so
  `music.907.life` correctly has none. Created a self-hosted Access application
  for `inbox.907.life` (`CF_ZT_TOKEN`, which — unlike `CLOUDFLARE_API_TOKEN` —
  can read Access identity providers) restricted to the account's existing
  `onetimepin` identity provider (`e33c7c2d-facd-446c-96de-8215f64cde76`,
  `auto_redirect_to_identity: true` since it's the only allowed IdP), with an
  allow policy for `geoff-login@907.life` (policy
  `92425260-87f6-4a3a-b66e-d71bb8d9d6a8`). Family emails join this policy at
  Task 8 per plan. **Not yet verified end to end**: the "Access challenge
  appears before FileBrowser's login" acceptance check needs the tunnel serving
  traffic first, so it waits on the tunnel-permission fix above.
- **Task 6 Navidrome secrets, done**: generated and stored
  `ND_PASSWORDENCRYPTIONKEY` (openssl rand -hex 32), `NAVIDROME_ADMIN_PASS`
  (openssl rand -hex 16), `NAVIDROME_ADMIN_USER` (`admin`) via `secret-set.sh`.
  No value printed; lengths only (64/32/5 chars) confirmed against
  `~/.local/secrets`.
- **registry.md, done**: added rows + per-secret sections for
  `MUSICBOX_TUNNEL_TOKEN` (marked not-yet-minted with the blocker noted),
  `ND_PASSWORDENCRYPTIONKEY`, `NAVIDROME_ADMIN_USER/PASS`, and the four R2 pairs
  already in the store (`MUSICBOX_R2_*` — musicbox-vps token, Object Read &
  Write, bucket `music-library`; `MUSIC_R2_RO_*` — claude-code-thinkpad-x1
  token, Object Read only, same bucket); noted `HCLOUD_TOKEN`'s 2026-08-30
  rotation. Committed to `~/.dotfiles` as two commits (registry.md text, then
  `values.age` separately) — a third file in the working tree,
  `bin/.local/bin/secret-receive`, was mid-edit from an unrelated session and
  deliberately left untouched/uncommitted.

Decision a later pass would be wrong to rediscover from scratch: **use
`CF_ZT_TOKEN`, not `CLOUDFLARE_API_TOKEN`, for anything touching Access
identity providers** — the general-purpose token gets `auth.forbidden` reading
`.../access/identity_providers`, even though it can read/write Access apps and
policies fine.

## Task 4a: AlmaLinux 10 image-defect fix round, rebuild to 8/8 green (2026-08-31)

Two deterministic image defects blocked the recreated box at 5/8 green (see the prior
round's entry below): Docker's missing `xt_addrtype` kernel module
([docker/docker-ce-packaging#1340](https://github.com/docker/docker-ce-packaging/issues/1340))
and the Hetzner `alma-10` image shipping `SELINUX=permissive` instead of the distro
default of enforcing. Both fixes landed in `cloud-init/user-data.yaml`'s `runcmd` plus a
new top-level `power_state` key:

- **Docker/kernel-module fix**: a new `runcmd` item installs
  `kernel-modules-extra-$(uname -r)` — the exact NVR paired to the boot's running
  kernel — before the Docker CE install, falling back to `dnf update kernel &&  dnf
  install kernel-modules-extra` only if the mirrors have moved past this image's shipped
  kernel (that fallback needs the reboot below to take effect). `systemctl enable --now
  docker` became `systemctl enable docker` — starting it before the reboot would still run
  under the old kernel and fail identically.
- **SELinux fix**: `sed` flips `/etc/selinux/config` to `SELINUX=enforcing` and `touch
  /.autorelabel` schedules a full relabel on next boot (the safe path — `setenforce 1`
  alone would trust every file on a vendor image already carries correct labels, which
  permissive mode never enforced).
- **Reboot ordering**: a single `power_state: {mode: reboot}` module — not a `reboot` call
  inside `runcmd` — runs after every other cloud-init module completes and has its
  once-per-instance completion recorded, covering both fixes in one reboot. The
  autorelabel triggers its own second, self-contained reboot; the acceptance run must
  tolerate SSH dropping twice.

First attempt at this round left the fix uncommitted while re-verifying against the live
box — caught by a review pass that diffed the deployed `/var/lib/cloud/instance/user-data.txt`
sha256 against `HEAD`'s copy of `cloud-init/user-data.yaml` and found them different
(`196e76f...` deployed vs `db0b633...` at HEAD). Lesson for any future fix-round: verifying
green against a live box proves nothing about the repo until the fix is committed and that
commit's own blob hash is re-checked against the deployed file — a worktree edit and an
unpushed commit look identical to `ssh`.

Re-verified after commit `3fa905d`: `sha256sum cloud-init/user-data.yaml` at HEAD equals
the deployed `/var/lib/cloud/instance/user-data.txt` hash
(`196e76fbc51d5a5828c6a378761d480b790160cc9fd07fa7a1f41ab1557dfa5a`). `bash
scripts/check.sh` exits 0.

**9 of 9 on-box checks green, re-verified 2026-08-31 against `root@5.78.148.42`:**

| Check | Command | Result |
|---|---|---|
| cloud-init | `cloud-init status --wait` | `status: done` — PASS |
| volume mount | `mountpoint -q /srv/music` | mounted — PASS |
| tree ownership | `ls -lnd /srv/music/{library,inbox,review,state,staging}` | all `2000 2000`; `inbox` is `drwxrwsr-x` (2775, setgid) — PASS |
| music user | `id music` | `uid=2000(music) gid=2000(music) groups=2000(music)` — PASS |
| docker | `systemctl is-active docker` | `active` — PASS |
| SELinux | `getenforce` | `Enforcing` — PASS |
| sshd password auth | `sshd -T \| grep -i '^passwordauthentication'` | `passwordauthentication no` — PASS |
| yq / jq | `/usr/local/bin/yq --version` / `jq --version` | `v4.53.6` / present (`jq-` is a cosmetic upstream version-string quirk on this AlmaLinux 10 build, not a functional defect) — PASS |
| fstab | `grep srv/music /etc/fstab` | exactly one line: `UUID=6614e3dd-9014-4949-9f4c-effcd317bd3a /srv/music ext4 defaults,nofail,x-systemd.device-timeout=30s 0 2` — PASS |

## Task 4a: real provisioning, forced-root-password fix round (2026-08-30/31)

`scripts/provision.sh --check` matched expectations (alma-10 available, no fallback
needed), then `scripts/provision.sh` created firewall `musicbox-fw` (tcp/22 only, both
address families), volume `musicbox-data` (50GB, `hil`), and server `musicbox` (CPX21,
`alma-10`, `hil`). Server IPv4 `5.78.148.42`, IPv6 `2a01:4ff:1f0:97c6::1`. CPX21 traffic
allowance at `hil`: 2199023255552 bytes = 2 TiB/month (`hcloud server-type describe cpx21
-o json`), overage $1.20/TB — generous for a family Navidrome, no need to default clients
to transcoded streams.

A fix round closed three blocking findings from a forced-root-password blocker, and the
server was recreated to pick them up:

1. `scripts/provision.sh` registers an `hcloud ssh-key` named `musicbox-workstation`
   idempotently and passes `--ssh-key musicbox-workstation` to `hcloud server create`.
   Confirmed on the recreate: no "Root password: ..." line appeared in the create output.
2. `cloud-init/user-data.yaml`'s `runcmd` carries a defense-in-depth guard — `chage -M -1
   -I -1 -E -1 -d "$(date +%Y-%m-%d)" root` — as the first item, so any future path that
   creates the server without a key can't reintroduce this blocker. Confirmed on-box:
   `chage -l root` shows `Password expires: never`.
3. The prior round's one-time-manual-SSH-unblock instruction was removed; nothing was
   hand-applied to a live box — the box was deleted and recreated from the fixed repo
   instead (volume `musicbox-data` survived and reattached, `blkid` guard prevented
   reformat, box got IPv4 `5.78.148.42` again).

Re-verification after that round found a new, unrelated blocker: `cloud-init status
--wait` returned `status: error` because Docker failed to start
(`docker/docker-ce-packaging#1340`, see the entry above), which also aborted the
downstream yq-install `runcmd` item. `getenforce` also reported `Permissive`, traced to
the Hetzner `alma-10` image's own shipped default, not to anything cloud-init touched.
Both became the next fix round, entry above.

## Task 3: service configuration — compose, systemd units, deploy (2026-08-30)

Landed `compose.yaml` (Navidrome, FileBrowser Quantum, cloudflared — no service
publishes a host port; only cloudflared is reachable from outside), `config/
navidrome.toml`, `config/filebrowser.yaml`, `config/msmtprc`, `config/beets-config.yaml`,
`config/cloudflared-tunnel-config.json`, the four `systemd/music-{import,backup}.
{service,timer}` units plus a shared `systemd/musicbox-ping-failure@.service` template,
`env/musicbox.env.template`, and `scripts/deploy.sh`. Extended `scripts/check.sh` with
`docker compose config -q` and a pattern-filtered `systemd-analyze verify`. All pinned
tags looked up and dated in `docs/STATUS.md`.

Two Task 2/Task 3 interface mismatches found and fixed (Task 3's own acceptance criteria
named the authoritative env var scheme, so Task 2's code was the thing that moved):
- **healthchecks.io slugs**: `music-import`/`music-backup` called `ping_hc
  musicbox-import` / `ping_hc musicbox-backup`, which resolves (per `lib.sh`'s own
  uppercase-and-dash-to-underscore rule) to `MUSICBOX_HC_MUSICBOX_IMPORT_URL` /
  `MUSICBOX_HC_MUSICBOX_BACKUP_URL` — doubled "MUSICBOX", not the
  `MUSICBOX_HC_IMPORT_URL` / `MUSICBOX_HC_BACKUP_URL` this task's env template requires.
  Fixed both scripts to call `ping_hc import` / `ping_hc backup`; updated
  `tests/lib.bats` and `tests/music-import.bats` to match (still 16 green). Monitoring
  would otherwise have silently no-op'd forever — `ping_hc` treats an unset URL var as a
  no-op, not an error, precisely so a monitoring bug never takes down the pipeline it
  monitors, which also means a naming bug like this one would never have surfaced as a
  failure anywhere.
- **Navidrome scan-trigger credentials**: `music-import`'s `trigger_navidrome_scan()`
  read `NAVIDROME_USER`/`NAVIDROME_PASSWORD`; this task's env template is required to
  name `NAVIDROME_ADMIN_USER`/`NAVIDROME_ADMIN_PASSWORD`. Renamed in the script and in
  `tests/music-import.bats`.

Two review findings fixed after the first pass (both would have broken the first real
deploy, not just this repo's own gate):
- **`/etc/musicbox/env`'s parent directory didn't exist.** Nothing in `provision.sh` or
  `cloud-init/user-data.yaml` creates `/etc/musicbox` (only `/srv/music` is built at
  provision time), so `install -m 600 ... /etc/musicbox/env` died with "cannot create
  regular file" on a fresh box, after the scripts and units had already landed. Fixed by
  adding `-D` to that `install` call, which creates the missing parent and is a no-op on
  every later re-deploy.
- **`config/msmtprc` was never actually installed to `/etc/msmtprc`.** The file's own
  header already claimed deploy.sh put it there, but deploy.sh only rsynced `config/` to
  `/opt/musicbox/config` for the record — `lib.sh`'s `notify()` calls plain `msmtp`, which
  resolves its account from `/etc/msmtprc` and would have found nothing there, aborting
  `music-import` under `set -euo pipefail` on the first skipped-track notification. Added
  an explicit install step (rsync to a temp path, then `install -m 0644` to
  `/etc/msmtprc`), with a matching diff-and-report line in `--check`.

Decisions a later pass would be wrong to rediscover from scratch:
- **Navidrome config split**: non-secret baseline settings (`MusicFolder`, `DataFolder`,
  `LogLevel`) live in `config/navidrome.toml`, bind-mounted read-only at
  `/data/navidrome.toml` alongside the read-write `/data` mount for the DB — Docker lets
  a single-file bind mount overlay a path inside an already-mounted directory. Scanner
  settings and the encryption key live in `compose.yaml`'s `environment:` instead (ND_
  env vars override the toml), specifically so the two option names the task called out
  by name (`ND_SCANNER_SCHEDULE`, `ND_SCANNER_WATCHERWAIT`) are literally grep-able in
  compose.yaml. `ND_CONFIGFILE=/data/navidrome.toml` is set explicitly rather than relied
  on implicitly, per navidrome.org's own Docker install guidance.
- **FileBrowser Quantum 1.5.x has no static per-user YAML config** — verified against the
  actual `v1.5.5-stable` tag's Go source (`backend/common/settings/structs.go`), not just
  the docs site (which mixes in v2.0 beta content). Users and their scopes live in its own
  SQLite DB, created via CLI/API, not this repo. `config/filebrowser.yaml` instead
  configures one source (`/srv/music/inbox`) with `createUserDir: true` +
  `defaultUserScope: "/"`, so creating an account named `alice` automatically scopes her
  to exactly `/srv/music/inbox/alice` — no per-user YAML needed, and the username IS the
  inbox directory name music-import already matches against `contributors.yaml`. Task 8
  creates the real accounts (one per `contributor`-role entry) once Task 0 lands real
  names.
- **cloudflared runs `tunnel --no-autoupdate run` with no local `--config`.** A
  token-driven tunnel is remotely managed; a local ingress file is silently ignored
  (verified against upstream `cloudflared` issues, not assumed). The actual ingress
  mapping is `config/cloudflared-tunnel-config.json`, a ready-to-PUT API body for Task 5
  to apply once the named tunnel exists — see STATUS.md's Open items, this is a real gap
  Task 5 must not skip.
- **msmtp's `passwordeval` reads `$MUSICBOX_SMTP_PASSWORD` from its own inherited process
  environment** (`printenv MUSICBOX_SMTP_PASSWORD`), not by re-reading
  `/etc/musicbox/env` from disk. That variable is already present in music-import's
  environment via its unit's `EnvironmentFile=`; having msmtp open the file a second time
  would require it group-readable by the unprivileged `music` user, widening every OTHER
  secret in that one file to the same process for no actual reduction in blast radius
  (that process already receives all of them via `EnvironmentFile=`). `config/msmtprc`
  itself holds no secret and installs world-readable at `/etc/msmtprc`.
- **Deploy target directory**: `deploy.sh` rsyncs `compose.yaml` + `config/` to
  `/opt/musicbox` on the box (not `/srv/music`, which is data, restored from R2, and
  never holds repo-sourced config). `docker compose` runs with `--project-directory
  /opt/musicbox` so the compose file's relative `./config/...` volume paths resolve
  regardless of the caller's cwd.
- **`deploy.sh --check` never prints secret contents**, even when diffing the rendered
  `/etc/musicbox/env` against what's already on the box: it reports only whether the
  content differs (`diff -q` via process substitution, nothing touches local disk).
  Exercised in this pass against an intentionally-unreachable host and against
  `127.0.0.1` (connection refused) — both paths report cleanly and exit 0 without a real
  box to deploy to.
- **beets on the box**: installed via `uv tool install --force beets==2.13.1` as the
  `music` user (`HOME=/srv/music`), then symlinked `/usr/local/bin/beet ->
  /srv/music/.local/bin/beet` — `uv tool install`'s default bin dir is not on systemd's
  default `PATH`, so without the symlink `music-import`'s `beet import` call would fail
  with "command not found" the first time a real timer fires. `config/beets-config.yaml`
  mirrors the workstation's `~/.dotfiles/beets` layout and plugins (fetchart, embedart),
  repointed at `/srv/music`, plus a quiet-mode section (`quiet_fallback: skip`,
  `none_rec_action: skip`) so `beet import -q` sends anything it can't confidently tag to
  review instead of guessing or blocking on input that will never arrive.

## Task 2: import, backup, and wrapper scripts (2026-08-30)

Landed `scripts/lib.sh` (`quiescent`, `safe_extract`, `free_space_ok`, `notify`,
`ping_hc`), TDD with `tests/lib.bats` written and confirmed red before any
implementation existed. Landed `music-import`, `music-backup`, `music-beet`, and
`scripts/check.sh` as the repo gate (shellcheck over every `scripts/*` bash file,
then the full bats suite). `tests/music-import.bats` runs the importer end to end
against a fixture `/srv/music` tree with `beet`, `curl`, and `msmtp` stubbed on
PATH: happy path, a beets skip filed to review with both notifications, a
non-quiescent inbox left untouched, and a free-space abort that fires the fail
ping before beet ever runs.

Two one-line review cleanups from Task 1 landed alongside: the volume glob guard
in `cloud-init/user-data.yaml` now also rejects a non-block-device match (a
zero-match glob previously fell through as a false positive on some shells), and
`provision.sh`'s firewall-rule check now fails with its own message when the
`hcloud firewall describe` call itself errors, instead of reporting a network
blip as a missing rule.

Decisions a later pass would be wrong to rediscover from scratch:
- **Contributor matching**: `music-import` matches an inbox directory name
  (`/srv/music/inbox/<user>`) against contributors.yaml's `name` field verbatim;
  there is no separate username/slug field. If Task 3/8's FileBrowser account
  provisioning wants a different inbox directory naming scheme, `contributors.yaml`
  needs a dedicated field and `uploader_email()` needs updating to match on it.
- **Navidrome scan trigger**: implemented against the native admin API (`POST
  /auth/login` for a JWT, then `POST /api/scanner/scan` with `x-nd-authorization:
  Bearer <token>`), not the Subsonic API, because Subsonic has no scan-trigger
  endpoint. This is the UI's internal API rather than a documented stable
  contract; verify it against the pinned Navidrome image tag once Task 3 chooses
  one, since a trigger failure is only best-effort (logged, not fatal) precisely
  because this endpoint is not guaranteed stable.
- **State DB paths default to** `$MUSICBOX_ROOT/state/library.db` (beets) and
  `$MUSICBOX_ROOT/state/navidrome/navidrome.db` (Navidrome), both overridable via
  env var. Task 3's `compose.yaml` must either mount Navidrome's data dir there or
  the env var overridden in `/etc/musicbox/env`.
- **Archive cap default**: `MUSICBOX_MAX_ARCHIVE_BYTES` defaults to 2 GiB in
  `music-import`; revisit if a real Bandcamp download exceeds it.
- New host dependency: `yq` (mikefarah/go, v4) for parsing `contributors.yaml`,
  needed on the box alongside `beet`/`rclone`/`msmtp`/`bsdtar`. Add it to Task 3's
  `deploy.sh` package list.
