#!/usr/bin/env bash
set -euo pipefail

# Deploys musicbox to the box: rsyncs repo config, renders /etc/musicbox/env
# from env/musicbox.env.template plus the workstation age-store values,
# cross-compiles and ships the dubplate Go binary, installs the Task 2
# scripts and every systemd unit (shell-layer and dubplate-*) to their
# fixed paths, brings the compose stack up, and installs beets via uv as
# the "music" user. Idempotent: re-running against an unchanged box is a
# no-op diff (rsync's own reporting, plus the env-render and compose-up
# steps being naturally idempotent). Shipping a dubplate-* unit here does
# not enable it -- see UNIT_FILES vs. TIMER_UNITS below -- flipping a job
# from its shell predecessor to dubplate is a supervised step the rewrite
# plan's box tasks run explicitly, never a side effect of a routine
# deploy.
#
# Usage:
#   deploy.sh [--check] [HOST]
# HOST defaults to `hcloud server ip musicbox` (Task 1/4's provisioned
# server) when omitted. --check diffs and reports every step without
# applying any of them, and without ever writing decrypted secret values to
# a log or a local file.

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_TEMPLATE="$REPO_ROOT/env/musicbox.env.template"
REMOTE_DIR="/opt/musicbox"
REMOTE_ENV="/etc/musicbox/env"
BEETS_VERSION="2.13.1"

# Pinned uv installer: a specific release tag (an immutable GitHub release
# asset URL, not the astral.sh/uv/install.sh alias that always serves
# whatever is current) plus the sha256 of that exact script, computed and
# recorded here rather than trusted from the download. Bump both together
# when uv needs an upgrade.
UV_VERSION="0.12.7"
UV_INSTALLER_URL="https://github.com/astral-sh/uv/releases/download/$UV_VERSION/uv-installer.sh"
UV_INSTALLER_SHA256="92e8554321e2bde08c9b1445dae47a65360f885274f31df51cdc2f9faa84e001"

SCRIPT_FILES=(
  music-import music-backup music-beet
  musicbox-ping-failure musicbox-disk-check musicbox-navidrome-check
  lib.sh
)
# One job per line, each service beside its own timer (the ping-failure
# template has none), so a new job is added as a visible pair rather than
# appended to a run-on line. The dubplate-* units (rewrite plan T12) ride
# alongside their music-*/musicbox-* shell predecessors rather than
# replacing them here: this array only ships unit FILES to the box, it
# never enables one (see the TIMER_UNITS enable-now list below), so both
# generations can sit on disk at once while the conductor's supervised
# flip -- not a routine deploy -- decides which is live. The old units
# leave this array only in T13, once every job has flipped.
UNIT_FILES=(
  music-import.service music-import.timer
  music-backup.service music-backup.timer
  musicbox-ping-failure@.service
  musicbox-disk-check.service musicbox-disk-check.timer
  musicbox-navidrome-check.service musicbox-navidrome-check.timer
  dubplate-backup.service dubplate-backup.timer
  dubplate-disk-check.service dubplate-disk-check.timer
  dubplate-navidrome-check.service dubplate-navidrome-check.timer
)

# ControlMaster/ControlPersist share one TCP+SSH handshake across this run's
# many ssh/rsync calls (roughly eight in the apply path) instead of paying
# full connection setup for each; the socket is cleaned up automatically
# 60s after the last call finishes.
SSH_OPTS=(-o StrictHostKeyChecking=accept-new -o ConnectTimeout=10 \
  -o ControlMaster=auto -o ControlPersist=60s -o "ControlPath=$HOME/.ssh/musicbox-deploy-%r@%h:%p")

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
# HCLOUD_TOKEN (to look up HOST) and every value env/musicbox.env.template
# names originate here (Task 0, via secret-set.sh), never in this repo.
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

# --- render the env file, in memory only ------------------------------
# Every NAME= line in the template must resolve to a non-empty value,
# either already exported in this shell or sourced from ~/.local/secrets
# above; every missing one is reported together, not one at a time.
missing=()
env_lines=()
while IFS= read -r line; do
  case "$line" in
  '' | '#'*) continue ;;
  esac
  name="${line%%=*}"
  value="${!name:-}"
  if [ -z "$value" ]; then
    missing+=("$name")
  else
    env_lines+=("$name=$value")
  fi
done <"$ENV_TEMPLATE"

if [ "${#missing[@]}" -gt 0 ]; then
  echo "deploy.sh: missing a value for the following variables (mint them via secret-set.sh, Task 0, then re-source ~/.local/secrets):" >&2
  printf '  %s\n' "${missing[@]}" >&2
  exit 1
fi

echo "musicbox deploy -> $SSH_TARGET$($check_only && echo ' (--check: nothing will be applied)')"

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

# --- scripts -> /usr/local/bin, units -> /etc/systemd/system -----------
echo "== scripts (${SCRIPT_FILES[*]}) -> $SSH_TARGET:/usr/local/bin =="
echo "== units (${UNIT_FILES[*]}) -> $SSH_TARGET:/etc/systemd/system =="
if $check_only; then
  rsync -avzn -e "$RSYNC_SSH" "${SCRIPT_FILES[@]/#/$REPO_ROOT/scripts/}" "$SSH_TARGET:/usr/local/bin/" || true
  rsync -avzn -e "$RSYNC_SSH" "${UNIT_FILES[@]/#/$REPO_ROOT/systemd/}" "$SSH_TARGET:/etc/systemd/system/" || true
else
  rsync -avz -e "$RSYNC_SSH" "${SCRIPT_FILES[@]/#/$REPO_ROOT/scripts/}" "$SSH_TARGET:/usr/local/bin/"
  # SC2029: the file-list expansion is local; client-side is wanted here too.
  # shellcheck disable=SC2029
  ssh "${SSH_OPTS[@]}" "$SSH_TARGET" "chmod 0755 ${SCRIPT_FILES[*]/#//usr/local/bin/}"
  rsync -avz -e "$RSYNC_SSH" "${UNIT_FILES[@]/#/$REPO_ROOT/systemd/}" "$SSH_TARGET:/etc/systemd/system/"
fi

# --- dubplate binary -> /usr/local/bin/dubplate -------------------------
# Cross-compiles via the Makefile's build-linux target (CGO_ENABLED=0
# GOOS=linux GOARCH=amd64), the same artifact `make check` never ships
# itself. The build runs under --check too: it only writes the gitignored
# local binary (never touches the box), and running it is what lets
# --check's rsync -n report a real diff instead of guessing from a stale
# local copy.
echo "== dubplate binary -> $SSH_TARGET:/usr/local/bin/dubplate =="
make -C "$REPO_ROOT" build-linux
DUBPLATE_BINARY="$REPO_ROOT/dubplate-linux-amd64"
if $check_only; then
  rsync -avzn -e "$RSYNC_SSH" "$DUBPLATE_BINARY" "$SSH_TARGET:/usr/local/bin/dubplate" || true
else
  rsync -avz -e "$RSYNC_SSH" "$DUBPLATE_BINARY" "$SSH_TARGET:/usr/local/bin/dubplate"
  run_remote "chmod 0755 /usr/local/bin/dubplate"
fi

# Smoke test: a binary that rsynced and chmodded fine can still fail to
# execute (wrong GOARCH, a missing dynamic dependency the box lacks); `set
# -euo pipefail` makes a non-zero exit here abort the whole deploy rather
# than leaving a broken binary in place undetected until the next timer
# fires.
echo "== smoke test: dubplate version =="
run_remote "/usr/local/bin/dubplate version"

# --- /etc/musicbox/env --------------------------------------------------
# Not deploy_file: this content is assembled in memory from secrets and must
# never touch local disk (see the usage note at the top of this file), so
# there is no LOCAL path to hand the helper. It already avoids the
# fixed-/tmp-name problem deploy_file solves too -- install(1) reads it
# straight off stdin, with no staged copy on the box at all.
echo "== $REMOTE_ENV (${#env_lines[@]} variables) =="
if $check_only; then
  if remote_reachable; then
    # shellcheck disable=SC2029
    if diff -q <(printf '%s\n' "${env_lines[@]}") \
      <(ssh "${SSH_OPTS[@]}" "$SSH_TARGET" "cat $REMOTE_ENV 2>/dev/null") >/dev/null 2>&1; then
      echo "  no changes (contents match; diff not shown, it's secrets)"
    else
      echo "  would change (contents differ; diff not shown, it's secrets)"
    fi
  else
    echo "  can't reach $SSH_TARGET to diff; would write ${#env_lines[@]} variables"
  fi
else
  # -D creates $REMOTE_ENV's parent (/etc/musicbox) if this is the box's
  # first deploy; install itself no-ops that part on every later run.
  # shellcheck disable=SC2029
  printf '%s\n' "${env_lines[@]}" | ssh "${SSH_OPTS[@]}" "$SSH_TARGET" \
    "install -D -m 600 -o root -g root /dev/stdin $REMOTE_ENV"
fi

# --- /etc/msmtprc ---------------------------------------------------------
# lib.sh's notify() calls plain `msmtp`, which reads its account config from
# /etc/msmtprc (msmtp's own system-wide default) — a separate destination
# from the config/ copy rsynced into $REMOTE_DIR above for the record.
echo "== /etc/msmtprc =="
deploy_file "$REPO_ROOT/config/msmtprc" /etc/msmtprc 0644 root:root

# --- /etc/musicbox/contributors.yaml --------------------------------------
# music-import's uploader_email() (Task 2) reads CONTRIBUTORS_FILE, which
# defaults to this exact path. root:music 0640 lets the `music` user (the
# systemd unit's User=) read it without making contributor emails
# world-readable.
echo "== /etc/musicbox/contributors.yaml =="
deploy_file "$REPO_ROOT/config/contributors.yaml" /etc/musicbox/contributors.yaml 0640 root:music

# --- state directories for bind-mounted volumes -------------------------
# Docker creates a missing bind-mount source itself, but as root:root; both
# navidrome and filebrowser run as 2000:2000 (compose.yaml) and would then be
# unable to write their own databases on first start. Pre-creating with the
# right ownership makes the first `up -d` on a fresh box work the same as
# every later one.
STATE_DIRS="/srv/music/state/navidrome /srv/music/state/filebrowser"
echo "== state directories ($STATE_DIRS) =="
run_remote "mkdir -p $STATE_DIRS && chown 2000:2000 $STATE_DIRS"

# --- /var/lib/musicbox (persistent import lock) ---------------------------
# music-import and music-beet both flock this directory's
# musicbox-import.lock (lib.sh's musicbox_lock_file). music-import.service
# runs it under ProtectSystem=strict with ReadWritePaths=/var/lib/musicbox;
# systemd requires that path to already exist on the host or the unit fails
# outright ("Failed to set up mount namespacing"), so it has to be
# pre-created here rather than left to the scripts' own mkdir -p, which
# only helps once the sandbox is already up. music:music matches the
# scripts' own User=music.
#
# It is also DUBPLATE_STATE_DIR (rewrite plan T12): dubplate's own lock
# files (internal/lock.Path, internal/backup's "backup.lock"), the import
# run report, and the per-run tag log all live here too, unified under one
# persistent directory rather than the shell layer's split between this
# path (lock only) and /srv/music/state (everything else) -- so no
# separate pre-create is needed for dubplate-backup.service's own
# ReadWritePaths, this directory already covers it.
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

# --- compose stack -------------------------------------------------------
COMPOSE_CMD="docker compose --env-file $REMOTE_ENV -f $REMOTE_DIR/compose.yaml --project-directory $REMOTE_DIR up -d"
echo "== compose stack =="
run_remote "$COMPOSE_CMD"

# --- systemd timers --------------------------------------------------
# Deliberately still the shell-layer timers only: daemon-reload above (via
# the units rsync) makes systemd aware of the dubplate-* units too, but
# enabling them is the rewrite plan's supervised per-job flip, not this
# routine step -- see the UNIT_FILES comment above.
echo "== systemd timers =="
TIMER_UNITS="music-import.timer music-backup.timer musicbox-disk-check.timer musicbox-navidrome-check.timer"
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

echo "musicbox deploy: done."
