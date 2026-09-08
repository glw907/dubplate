#!/usr/bin/env bash
set -euo pipefail

# Deploys dubplate to the box: rsyncs repo config, renders /etc/dubplate/env
# and /etc/dubplate/server.env from env/dubplate.env.template and
# env/dubplate-server.env.template plus the workstation age-store values,
# cross-compiles and ships the dubplate Go binary, installs every
# dubplate-* systemd unit to its fixed path, brings the compose stack up,
# and installs beets via uv as the "music" user. Idempotent: re-running
# against an unchanged box is a no-op diff (rsync's own reporting, plus the
# env-render and compose-up steps being naturally idempotent).
#
# Rewrite plan T13 (single-generation closeout): the shell layer, its
# scripts, and the music-*/musicbox-* units it shipped alongside are gone.
# This is the only generation deploy.sh ships or enables now -- there is no
# separate supervised flip step left to run.
#
# Usage:
#   deploy.sh [--check] [HOST]
# HOST defaults to `hcloud server ip musicbox` (the box's Hetzner server
# name, unrenamed -- see docs/STATUS.md on the estate rename's box-side
# remainder) when omitted. --check diffs and reports every step without
# applying any of them, and without ever writing decrypted secret values to
# a log or a local file.

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_TEMPLATE="$REPO_ROOT/env/dubplate.env.template"
SERVER_ENV_TEMPLATE="$REPO_ROOT/env/dubplate-server.env.template"
REMOTE_DIR="/opt/dubplate"
REMOTE_ENV="/etc/dubplate/env"
REMOTE_SERVER_ENV="/etc/dubplate/server.env"
BEETS_VERSION="2.13.1"

# Pinned uv installer: a specific release tag (an immutable GitHub release
# asset URL, not the astral.sh/uv/install.sh alias that always serves
# whatever is current) plus the sha256 of that exact script, computed and
# recorded here rather than trusted from the download. Bump both together
# when uv needs an upgrade.
UV_VERSION="0.12.7"
UV_INSTALLER_URL="https://github.com/astral-sh/uv/releases/download/$UV_VERSION/uv-installer.sh"
UV_INSTALLER_SHA256="92e8554321e2bde08c9b1445dae47a65360f885274f31df51cdc2f9faa84e001"

# One job per line, each service beside its own timer (the ping-fail
# template has none), so a new job is added as a visible pair rather than
# appended to a run-on line. Every unit below execs the frozen incumbent
# binary at /usr/local/bin/dubplate, shipped once before rung 1 and never
# touched by this script again (the boundary comment at the
# dubplate-server binary step below has the detail) -- there is no
# separate scripts-to-/usr/local/bin manifest left now that the shell
# layer is gone.
#
# U12a adds the uploader's three units to this rsync+daemon-reload list --
# dubplate-serve.service, and the nudge path/service pair -- but NOT to
# TIMER_UNITS below: unlike every existing job, serve is the estate's one
# user-facing surface and must not come up unattended before U12a's
# Access/tunnel wiring is verified, so it (and its nudge path, which pairs
# with it) is `systemctl enable --now` only in that box step, by hand,
# under Geoff's sanction -- never by this script.
UNIT_FILES=(
  dubplate-import.service dubplate-import.timer
  dubplate-backup.service dubplate-backup.timer
  dubplate-disk-check.service dubplate-disk-check.timer
  dubplate-navidrome-check.service dubplate-navidrome-check.timer
  dubplate-shadow.service dubplate-shadow.timer
  "dubplate-ping-fail@.service"
  dubplate-serve.service
  dubplate-import-nudge.path dubplate-import-nudge.service
)

# ControlMaster/ControlPersist share one TCP+SSH handshake across this run's
# many ssh/rsync calls (roughly eight in the apply path) instead of paying
# full connection setup for each; the socket is cleaned up automatically
# 60s after the last call finishes.
SSH_OPTS=(-o StrictHostKeyChecking=accept-new -o ConnectTimeout=10 \
  -o ControlMaster=auto -o ControlPersist=60s -o "ControlPath=$HOME/.ssh/dubplate-deploy-%r@%h:%p")

check_only=false
HOST=""
for arg in "$@"; do
  case "$arg" in
  --check)
    check_only=true
    ;;
  -*)
    echo "Unknown argument: $arg" >&2
    exit 1
    ;;
  *)
    HOST="$arg"
    ;;
  esac
done

# --- workstation secrets ---------------------------------------------
# HCLOUD_TOKEN (to look up HOST) and every value env/dubplate.env.template
# names originate here (via secret-set.sh), never in this repo.
if [ -f "$HOME/.local/secrets" ]; then
  # shellcheck disable=SC1091
  source "$HOME/.local/secrets"
fi

# --- resolve target host ----------------------------------------------
if [ -z "$HOST" ]; then
  if [ -z "${HCLOUD_TOKEN:-}" ] || ! command -v hcloud >/dev/null 2>&1; then
    echo "deploy.sh: no HOST given and can't look one up (need HCLOUD_TOKEN + hcloud). Pass an IP, or provision first (Task 4)." >&2
    exit 1
  fi
  HOST="$(HCLOUD_TOKEN="$HCLOUD_TOKEN" hcloud server ip musicbox 2>/dev/null || true)"
  if [ -z "$HOST" ]; then
    echo "deploy.sh: couldn't resolve the musicbox server's IP via hcloud. Pass an IP, or provision first (Task 4)." >&2
    exit 1
  fi
fi
SSH_TARGET="root@$HOST"
RSYNC_SSH="ssh ${SSH_OPTS[*]}"

# True when the box answers an ssh that will never prompt: BatchMode turns a
# missing key or an unknown host into an immediate failure instead of a
# password prompt that would hang an unattended --check run. Every --check
# path that wants to diff against live remote content probes with this
# first, so an unreachable box reports "can't reach" rather than an empty
# remote file that looks like a pending change.
remote_reachable() {
  ssh -o BatchMode=yes "${SSH_OPTS[@]}" "$SSH_TARGET" true 2>/dev/null
}

# run_remote COMMAND
# Reports the command under --check and runs it on the box otherwise, so a
# step's --check line can never drift from what apply actually runs. SC2029
# is disabled once here rather than at each step: every command passed in is
# assembled from this script's own constants, so client-side expansion is
# exactly what is wanted.
run_remote() {
  local command="$1"
  if $check_only; then
    echo "  would run: $command"
  else
    # shellcheck disable=SC2029
    ssh "${SSH_OPTS[@]}" "$SSH_TARGET" "$command"
  fi
}

# deploy_file LOCAL REMOTE MODE OWNER:GROUP
# Shared check/diff/install shape for one small config file. --check probes
# reachability, then diffs LOCAL against REMOTE's live content and reports
# without staging or writing anything. Apply mode stages LOCAL at a path
# `mktemp` creates on the box itself -- never a fixed /tmp name, so the
# staged copy can't sit at a predictable path an unprivileged local user on
# the box could read before install(1) moves it into place -- then installs
# it and removes the staged copy.
deploy_file() {
  local local_path="$1" remote_path="$2" mode="$3" owner_group="$4"
  local owner="${owner_group%%:*}" group="${owner_group##*:}"
  if $check_only; then
    if remote_reachable; then
      # shellcheck disable=SC2029
      if diff -q "$local_path" <(ssh "${SSH_OPTS[@]}" "$SSH_TARGET" "cat $remote_path 2>/dev/null") >/dev/null 2>&1; then
        echo "  no changes"
      else
        echo "  would install $local_path -> $remote_path"
      fi
    else
      echo "  can't reach $SSH_TARGET to diff; would install $remote_path"
    fi
  else
    local remote_tmp
    remote_tmp="$(ssh "${SSH_OPTS[@]}" "$SSH_TARGET" mktemp)"
    rsync -avz -e "$RSYNC_SSH" "$local_path" "$SSH_TARGET:$remote_tmp"
    # shellcheck disable=SC2029
    ssh "${SSH_OPTS[@]}" "$SSH_TARGET" \
      "install -D -m $mode -o $owner -g $group $remote_tmp $remote_path && rm -f $remote_tmp"
  fi
}

# --- render both env templates, in memory only ------------------------
# render_missing accumulates across both templates so one run reports
# every unresolvable row together, not one template at a time: the
# second template's render pass always runs, even when the first
# already found a missing row, so a fix-one-rerun-hits-the-next-miss
# loop never happens.
render_missing=()

# render_template TEMPLATE OUT_ARRAY_NAME
# Resolves one line-oriented template into the array named by
# OUT_ARRAY_NAME (bash nameref), entirely in memory. Three row states
# (both templates' own headers name all three):
#   - blank or '#'-comment: skipped.
#   - NAME=value (non-empty right-hand side): a literal, used verbatim
#     and never looked up in the environment, even when a variable of
#     that name happens to be exported -- this is the row state that
#     lets env/dubplate-server.env.template's DUBPLATE_STATE_DIR win over
#     the incumbent's own exported value of the same name (spec's
#     flat-namespace hazard).
#   - bare NAME= (empty right-hand side): resolved from this shell's own
#     environment (already exported, or sourced from ~/.local/secrets
#     above); unresolved names are appended to render_missing together
#     with the template they came from, never reported one at a time.
# A literal value may itself contain '=': the name is everything before
# the first '=' (`${line%%=*}`), and the literal is everything after the
# first '=' (`${line#*=}`, which strips only the shortest -- i.e. first --
# match), so a literal like `FOO=a=b` renders as `a=b`. A line with none
# of the three shapes above -- no '=' at all -- is refused as malformed
# and reported in render_missing rather than accepted as a
# self-referential literal.
render_template() {
  local template="$1" out_name="$2"
  local -n out_ref="$out_name"
  local line name rhs value
  while IFS= read -r line; do
    case "$line" in
    '' | '#'*) continue ;;
    *=*) ;;
    *)
      render_missing+=("$template: malformed line (no '='): $line")
      continue
      ;;
    esac
    name="${line%%=*}"
    rhs="${line#*=}"
    if [ -n "$rhs" ]; then
      out_ref+=("$name=$rhs")
    else
      value="${!name:-}"
      if [ -z "$value" ]; then
        render_missing+=("$template: $name")
      else
        out_ref+=("$name=$value")
      fi
    fi
  done <"$template"
}

# SC2034: both arrays are only ever written and read through
# render_template's and install_env_step's own `local -n` namerefs, passed
# by name; shellcheck does not follow that indirection.
# shellcheck disable=SC2034
env_lines=()
# shellcheck disable=SC2034
server_env_lines=()
render_template "$ENV_TEMPLATE" env_lines
render_template "$SERVER_ENV_TEMPLATE" server_env_lines

if [ "${#render_missing[@]}" -gt 0 ]; then
  echo "deploy.sh: missing a value for the following variables (mint them via secret-set.sh, Task 0, then re-source ~/.local/secrets):" >&2
  printf '  %s\n' "${render_missing[@]}" >&2
  exit 1
fi

echo "dubplate deploy -> $SSH_TARGET$($check_only && echo ' (--check: nothing will be applied)')"

# --- repo config: compose.yaml + config/ -------------------------------
echo "== repo config -> $SSH_TARGET:$REMOTE_DIR =="
if $check_only; then
  rsync -avzn --delete -e "$RSYNC_SSH" \
    "$REPO_ROOT/compose.yaml" "$REPO_ROOT/config" "$SSH_TARGET:$REMOTE_DIR/" || true
else
  # SC2029: $REMOTE_DIR is a local constant, client-side expansion is wanted.
  # shellcheck disable=SC2029
  ssh "${SSH_OPTS[@]}" "$SSH_TARGET" "mkdir -p $REMOTE_DIR"
  rsync -avz --delete -e "$RSYNC_SSH" \
    "$REPO_ROOT/compose.yaml" "$REPO_ROOT/config" "$SSH_TARGET:$REMOTE_DIR/"
fi

# --- units -> /etc/systemd/system ---------------------------------------
echo "== units (${UNIT_FILES[*]}) -> $SSH_TARGET:/etc/systemd/system =="
if $check_only; then
  rsync -avzn -e "$RSYNC_SSH" "${UNIT_FILES[@]/#/$REPO_ROOT/systemd/}" "$SSH_TARGET:/etc/systemd/system/" || true
else
  rsync -avz -e "$RSYNC_SSH" "${UNIT_FILES[@]/#/$REPO_ROOT/systemd/}" "$SSH_TARGET:/etc/systemd/system/"
fi

# --- tmpfiles fragment: the nudge trigger directory ---------------------
# Installed and activated BEFORE any unit starts: dubplate-serve.service
# writes nudge trigger files under /run/dubplate/nudge-triggers (U7) and
# dubplate-import-nudge.path watches that same directory with
# DirectoryNotEmpty= (U8) -- both fail outright if it does not already
# exist, and /run is tmpfs, so it never survives a reboot on its own.
# deploy_file installs the fragment itself; systemd-tmpfiles --create
# then materializes the directories from it immediately (rather than
# waiting for the next boot), matching what a fresh boot's own
# systemd-tmpfiles-setup.service run would do. run_remote already prints
# "would run: ..." under --check without executing anything.
echo "== /etc/tmpfiles.d/dubplate.conf =="
deploy_file "$REPO_ROOT/config/tmpfiles-dubplate.conf" /etc/tmpfiles.d/dubplate.conf 0644 root:root
run_remote "systemd-tmpfiles --create /etc/tmpfiles.d/dubplate.conf"

# Cheap, informational, non-fatal: under --check, report whether the
# trigger directory already exists on the box (it survives until the
# next reboot even though it is tmpfs, so a re-run of this script
# usually finds it present) -- purely a visibility aid for reviewing a
# --check run, never a pass/fail gate on its own.
if $check_only; then
  trigger_dir="/run/dubplate/nudge-triggers"
  # shellcheck disable=SC2029
  if ! remote_reachable; then
    echo "  can't reach $SSH_TARGET to probe $trigger_dir"
  elif ssh "${SSH_OPTS[@]}" "$SSH_TARGET" "test -d $trigger_dir"; then
    echo "  $trigger_dir: exists"
  else
    echo "  $trigger_dir: missing"
  fi
fi

# --- dubplate-server binary -> /usr/local/bin/dubplate-server -----------
# Boundary (rung 1, additive deploy path): the incumbent
# /usr/local/bin/dubplate is frozen -- its source is the pre-greenfield
# tag -- and this script never writes to it again. The four guardian
# units (dubplate-import, dubplate-backup, dubplate-disk-check,
# dubplate-navidrome-check) exec that frozen path directly, so they are
# unaffected by any future run of this script. Greenfield units start
# exec-ing /usr/local/bin/dubplate-server from rung 4; the incumbent
# binary and its step retire together at rung 16, once every unit has
# moved over.
#
# Cross-compiles via the Makefile's build-linux target (CGO_ENABLED=0
# GOOS=linux GOARCH=amd64), the same artifact `make check` never ships
# itself. The build runs under --check too: it only writes the gitignored
# local binary (never touches the box), and running it is what lets
# --check's rsync -n report a real diff instead of guessing from a stale
# local copy.
#
# The plan's "web build before Go cross-compile" ordering (so `go:embed`
# picks up real Vite output, not just the placeholder) is U9's concern,
# not this pass's -- no web/ build step exists yet for this line to order
# against.
#
# DUBPLATE_STAMP_VERSION and DUBPLATE_STAMP_MATCHER_COMMIT are the two
# values that make the smoke step's own assertion meaningful: main.go's
# matcherCommit stays "dev" (and shadow observe refuses to run) on any
# build that skips these -X flags, so a bare `go build` or `make
# build-linux` run outside this script never reaches the box stamped.
# `git describe` covers the human-facing version; `git rev-parse HEAD`
# pins the exact commit the box's Go matcher runs, independent of tags.
DUBPLATE_STAMP_VERSION="$(git -C "$REPO_ROOT" describe --tags --always --dirty)"
DUBPLATE_STAMP_MATCHER_COMMIT="$(git -C "$REPO_ROOT" rev-parse HEAD)"

echo "== dubplate-server binary -> $SSH_TARGET:/usr/local/bin/dubplate-server =="
make -C "$REPO_ROOT" build-linux LDFLAGS="-X main.version=$DUBPLATE_STAMP_VERSION -X main.matcherCommit=$DUBPLATE_STAMP_MATCHER_COMMIT"
DUBPLATE_SERVER_BINARY="$REPO_ROOT/dubplate-server-linux-amd64"
if $check_only; then
  rsync -avzn -e "$RSYNC_SSH" "$DUBPLATE_SERVER_BINARY" "$SSH_TARGET:/usr/local/bin/dubplate-server" || true
else
  rsync -avz -e "$RSYNC_SSH" "$DUBPLATE_SERVER_BINARY" "$SSH_TARGET:/usr/local/bin/dubplate-server"
  run_remote "chmod 0755 /usr/local/bin/dubplate-server"
fi

# Smoke test: a binary that rsynced and chmodded fine can still fail to
# execute (wrong GOARCH, a missing dynamic dependency the box lacks); `set
# -euo pipefail` makes a non-zero exit here abort the whole deploy rather
# than leaving a broken binary in place undetected until the next timer
# fires. The assertion below extends that same abort to a binary that
# executes fine but never got stamped: it prints both "version" rows for
# the operator, then exits nonzero when matcher-commit is empty or "dev"
# (main.go's own workstation default), which shadow observe would refuse
# to run against for thirty days until the next deploy caught it. R9:
# cmd/dubplate-server/shadow.go builds the MusicBrainz User-Agent from
# this same version stamp (shadow.go:215), so the box sends the
# `git describe` string on every request, and a dirty worktree at
# deploy time ships a -dirty User-Agent.
echo "== smoke test: dubplate-server version =="
# shellcheck disable=SC2016
run_remote 'output="$(/usr/local/bin/dubplate-server version)"; printf "%s\n" "$output"; commit=$(printf "%s\n" "$output" | awk "/^matcher-commit/{print \$2}"); if [ -z "$commit" ] || [ "$commit" = "dev" ]; then echo "matcher-commit is dev or empty; aborting deploy" >&2; exit 1; fi'

# --- /etc/dubplate/env and /etc/dubplate/server.env ---------------------
# Not deploy_file: this content is assembled in memory from secrets and
# literals and must never touch local disk (see the usage note at the top
# of this file), so there is no LOCAL path to hand the helper. Each step
# below announces its own remote path and its own variable count, so the
# operator reading --check sees two distinct env steps rather than one.

# install_env_step REMOTE_PATH LINES_ARRAY_NAME
# Shared check/report/install shape for one rendered env array. --check
# probes reachability, then diffs the rendered lines against REMOTE_PATH's
# live content without ever printing content itself (it's secrets, same
# wording as before this refactor). Apply mode pipes the rendered lines
# straight into install(1) over stdin -- no staged copy on the box at all
# (Resolution R6), and -D creates the parent directory (/etc/dubplate) on
# the box's first deploy; install itself no-ops that part on every later
# run.
install_env_step() {
  local remote_path="$1" lines_name="$2"
  local -n lines_ref="$lines_name"
  echo "== $remote_path (${#lines_ref[@]} variables) =="
  if $check_only; then
    if remote_reachable; then
      # shellcheck disable=SC2029
      if diff -q <(printf '%s\n' "${lines_ref[@]}") \
        <(ssh "${SSH_OPTS[@]}" "$SSH_TARGET" "cat $remote_path 2>/dev/null") >/dev/null 2>&1; then
        echo "  no changes (contents match; diff not shown, it's secrets)"
      else
        echo "  would change (contents differ; diff not shown, it's secrets)"
      fi
    else
      echo "  can't reach $SSH_TARGET to diff; would write ${#lines_ref[@]} variables"
    fi
  else
    # shellcheck disable=SC2029
    printf '%s\n' "${lines_ref[@]}" | ssh "${SSH_OPTS[@]}" "$SSH_TARGET" \
      "install -D -m 600 -o root -g root /dev/stdin $remote_path"
  fi
}

install_env_step "$REMOTE_ENV" env_lines

# dubplate-server's own env file, installed 0600 root:root exactly like
# /etc/dubplate/env above (Resolution R6): systemd reads EnvironmentFile=
# as PID 1 before dropping privileges, so this mode and ownership are not
# a default to widen once dubplate-server itself runs as a lower-privilege
# user.
install_env_step "$REMOTE_SERVER_ENV" server_env_lines

# msmtp retired at T13: internal/notify sends mail itself (SMTP, no local
# MTA), and notify parity against msmtp had already passed before this
# change. No /etc/msmtprc install step remains.

# --- /etc/dubplate/contributors.yaml --------------------------------------
# internal/importer's roster lookup reads DUBPLATE_CONTRIBUTORS_FILE, which
# the rendered env file (above) points at this exact path. root:music 0640
# lets the `music` user (the systemd units' User=) read it without making
# contributor emails world-readable.
echo "== /etc/dubplate/contributors.yaml =="
deploy_file "$REPO_ROOT/config/contributors.yaml" /etc/dubplate/contributors.yaml 0640 root:music

# --- state directories for bind-mounted volumes -------------------------
# Docker creates a missing bind-mount source itself, but as root:root;
# navidrome runs as 2000:2000 (compose.yaml) and would then be unable to
# write its own database on first start. Pre-creating with the right
# ownership makes the first `up -d` on a fresh box work the same as
# every later one.
STATE_DIRS="/srv/music/state/navidrome"
echo "== state directories ($STATE_DIRS) =="
run_remote "mkdir -p $STATE_DIRS && chown 2000:2000 $STATE_DIRS"

# --- /var/lib/musicbox (DUBPLATE_STATE_DIR) --------------------------------
# Deliberately still named /var/lib/musicbox, not /var/lib/dubplate: this is
# a box data path, and T13's rename covers the estate's config/identity
# layer (/etc, /opt) only -- renaming a live data directory needs a
# supervised box-side migration, out of this change's scope (see
# docs/STATUS.md).
#
# dubplate's import, backup, and beet-passthrough commands all flock a lock
# file here (internal/lock.Path); the import run report and per-run tag log
# (internal/beets.Runner) live here too, unified under one persistent,
# non-RuntimeDirectory directory. dubplate-import.service and
# dubplate-backup.service both run it under ProtectSystem=strict with this
# path in their own ReadWritePaths=; systemd requires the path to already
# exist on the host or the unit fails outright ("Failed to set up mount
# namespacing"), so it has to be pre-created here rather than left to
# dubplate's own mkdir -p, which only helps once the sandbox is already up.
# music:music matches every unit's own User=music.
MUSICBOX_LIB_DIR="/var/lib/musicbox"
echo "== $MUSICBOX_LIB_DIR =="
run_remote "mkdir -p $MUSICBOX_LIB_DIR && chown music:music $MUSICBOX_LIB_DIR"

# --- /srv/music/state/backup-staging (dubplate-backup.service) ------------
# internal/backup.Runner stages the sqlite3 snapshots here before syncing
# them to R2, clearing this directory's CONTENTS (never the directory
# itself) on every run, under ProtectSystem=strict with
# ReadWritePaths=/srv/music/state/backup-staging;
# same "must exist on the host before the unit starts" requirement as
# /var/lib/musicbox above. Kept on the /srv/music volume (not
# /var/lib/musicbox) since a large library's DB snapshots belong beside the
# library's own disk budget, not the root filesystem's.
DUBPLATE_BACKUP_STAGING_DIR="/srv/music/state/backup-staging"
echo "== $DUBPLATE_BACKUP_STAGING_DIR =="
run_remote "mkdir -p $DUBPLATE_BACKUP_STAGING_DIR && chown music:music $DUBPLATE_BACKUP_STAGING_DIR"

# --- /srv/music/uploads (DUBPLATE_UPLOADS_DIR, U12a) ---------------------
# Not one of cloud-init's original `mkdir -p /srv/music/{library,inbox,
# review,state,staging}` (a fresh re-provision only gets this directory
# from THIS step, live remediation per vps-conventions -- cloud-init never
# converges an already-running box, and this deploy step already runs
# against every box including a fresh one, so no cloud-init edit is
# needed to keep a new provision in sync). Same volume as InboxDir on
# purpose: dubplate-serve.service's own startup check
# (internal/serve.sameFilesystem) refuses to start unless UploadsDir and
# InboxDir share a filesystem, since promotion is one atomic os.Rename.
# dubplate-serve.service's own ReadWritePaths= requires this path to
# already exist on the host before it starts, same as every directory
# above.
DUBPLATE_UPLOADS_DIR="/srv/music/uploads"
echo "== $DUBPLATE_UPLOADS_DIR =="
run_remote "mkdir -p $DUBPLATE_UPLOADS_DIR && chown music:music $DUBPLATE_UPLOADS_DIR"

# --- /var/lib/dubplate/shadow (dubplate-shadow.service) -------------------
# dubplate-shadow.service names /var/lib/dubplate/shadow in its own
# ReadWritePaths= and, under ProtectSystem=strict, systemd requires that
# exact leaf to already exist on the host before the unit starts, or the
# unit fails outright with "Failed to set up mount namespacing" -- same
# requirement as /var/lib/musicbox above. The unit's own OpenStore
# (internal/engine/shadow/store.go) cannot create it either: startup fails
# before dubplate-server ever runs, so nothing inside the process gets a
# chance to mkdir it. mkdir -p creates /var/lib/dubplate as a side effect
# and leaves it root-owned, so the chown below names both the parent and
# the granted leaf, even though ReadWritePaths= grants only the leaf
# (Resolution R1).
# music:music matches dubplate-shadow.service's own User=music.
DUBPLATE_LIB_DIR="/var/lib/dubplate"
DUBPLATE_SHADOW_DIR="$DUBPLATE_LIB_DIR/shadow"
echo "== $DUBPLATE_SHADOW_DIR =="
run_remote "mkdir -p $DUBPLATE_SHADOW_DIR && chown music:music $DUBPLATE_LIB_DIR $DUBPLATE_SHADOW_DIR"

# --- compose stack -------------------------------------------------------
COMPOSE_CMD="docker compose --env-file $REMOTE_ENV -f $REMOTE_DIR/compose.yaml --project-directory $REMOTE_DIR up -d"
echo "== compose stack =="
run_remote "$COMPOSE_CMD"

# --- systemd timers --------------------------------------------------
# daemon-reload (via the units rsync just above) picks up every unit file,
# including dubplate-ping-fail@.service, which is never enabled directly --
# it only ever runs as an OnFailure= instance. TIMER_UNITS deliberately
# does NOT include dubplate-serve.service or dubplate-import-nudge.path:
# unlike every job below, serve is the estate's one user-facing surface,
# so it (and its paired nudge path) is `systemctl enable --now` only by
# hand, in U12a's own Geoff-sanctioned box step, once the Access
# application and tunnel ingress are wired and verified -- never
# unattended by this script. TIMER_UNITS also deliberately does NOT
# include dubplate-shadow.timer: per Resolution R10, enabling the shadow
# is the launch sitting's own hand-run step, not part of a silent deploy,
# because that act starts the shadow's evidence clock.
echo "== systemd timers =="
TIMER_UNITS="dubplate-import.timer dubplate-backup.timer dubplate-disk-check.timer dubplate-navidrome-check.timer"
run_remote "systemctl daemon-reload && systemctl enable --now $TIMER_UNITS"

# --- beets, via uv, as the music user ----------------------------------
echo "== beets $BEETS_VERSION (uv, music user) =="
if $check_only; then
  echo "  would install config/beets-config.yaml -> /srv/music/.config/beets/config.yaml"
  echo "  would run (as music): uv tool install --force beets==$BEETS_VERSION"
  echo "  would symlink /usr/local/bin/beet -> /srv/music/.local/bin/beet"
else
  beets_config_tmp="$(ssh "${SSH_OPTS[@]}" "$SSH_TARGET" mktemp)"
  rsync -avz -e "$RSYNC_SSH" "$REPO_ROOT/config/beets-config.yaml" "$SSH_TARGET:$beets_config_tmp"
  # shellcheck disable=SC2087
  ssh "${SSH_OPTS[@]}" "$SSH_TARGET" bash -s <<REMOTE
set -euo pipefail
# UV_INSTALL_DIR=/usr/local/bin (rather than the installer's root-only
# default of /root/.local/bin) puts uv on every user's PATH, including the
# non-login shell this whole heredoc runs in and the runuser -u music call
# below; without it uv installs where only an interactive root login would
# find it. The command -v guard skips the install (and its download+network
# round trip) on every re-deploy once uv is already present.
if ! command -v uv >/dev/null 2>&1; then
  uv_installer="\$(mktemp)"
  curl -LsSf -o "\$uv_installer" "$UV_INSTALLER_URL"
  echo "$UV_INSTALLER_SHA256  \$uv_installer" | sha256sum -c -
  UV_INSTALL_DIR=/usr/local/bin sh "\$uv_installer"
  rm -f "\$uv_installer"
fi
install -d -o music -g music -m 0755 /srv/music/.config/beets
install -o music -g music -m 0644 $beets_config_tmp /srv/music/.config/beets/config.yaml
rm -f $beets_config_tmp
runuser -u music -- env HOME=/srv/music uv tool install --force "beets==$BEETS_VERSION"
ln -sf /srv/music/.local/bin/beet /usr/local/bin/beet
REMOTE
fi

echo "dubplate deploy: done."
