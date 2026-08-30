#!/usr/bin/env bash
set -euo pipefail

# Deploys musicbox to the box: rsyncs repo config, renders /etc/musicbox/env
# from env/musicbox.env.template plus the workstation age-store values,
# installs the Task 2 scripts and this task's systemd units to their fixed
# paths, brings the compose stack up, and installs beets via uv as the
# "music" user. Idempotent: re-running against an unchanged box is a no-op
# diff (rsync's own reporting, plus the env-render and compose-up steps
# being naturally idempotent).
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

SCRIPT_FILES=(music-import music-backup music-beet musicbox-ping-failure lib.sh)
UNIT_FILES=(music-import.service music-import.timer music-backup.service music-backup.timer musicbox-ping-failure@.service)

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

# --- /etc/musicbox/env --------------------------------------------------
echo "== $REMOTE_ENV (${#env_lines[@]} variables) =="
if $check_only; then
  if ssh -o BatchMode=yes "${SSH_OPTS[@]}" "$SSH_TARGET" true 2>/dev/null; then
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
  # shellcheck disable=SC2029
  printf '%s\n' "${env_lines[@]}" | ssh "${SSH_OPTS[@]}" "$SSH_TARGET" \
    "install -m 600 -o root -g root /dev/stdin $REMOTE_ENV"
fi

# --- compose stack -------------------------------------------------------
COMPOSE_CMD="docker compose --env-file $REMOTE_ENV -f $REMOTE_DIR/compose.yaml --project-directory $REMOTE_DIR up -d"
echo "== compose stack =="
if $check_only; then
  echo "  would run: $COMPOSE_CMD"
else
  # shellcheck disable=SC2029
  ssh "${SSH_OPTS[@]}" "$SSH_TARGET" "$COMPOSE_CMD"
fi

# --- systemd timers --------------------------------------------------
echo "== systemd timers =="
if $check_only; then
  echo "  would run: systemctl daemon-reload && systemctl enable --now music-import.timer music-backup.timer"
else
  ssh "${SSH_OPTS[@]}" "$SSH_TARGET" \
    "systemctl daemon-reload && systemctl enable --now music-import.timer music-backup.timer"
fi

# --- beets, via uv, as the music user ----------------------------------
echo "== beets $BEETS_VERSION (uv, music user) =="
if $check_only; then
  echo "  would install config/beets-config.yaml -> /srv/music/.config/beets/config.yaml"
  echo "  would run (as music): uv tool install --force beets==$BEETS_VERSION"
  echo "  would symlink /usr/local/bin/beet -> /srv/music/.local/bin/beet"
else
  rsync -avz -e "$RSYNC_SSH" "$REPO_ROOT/config/beets-config.yaml" "$SSH_TARGET:/tmp/musicbox-beets-config.yaml"
  # shellcheck disable=SC2087
  ssh "${SSH_OPTS[@]}" "$SSH_TARGET" bash -s <<REMOTE
set -euo pipefail
command -v uv >/dev/null 2>&1 || curl -LsSf https://astral.sh/uv/install.sh | sh
install -d -o music -g music -m 0755 /srv/music/.config/beets
install -o music -g music -m 0644 /tmp/musicbox-beets-config.yaml /srv/music/.config/beets/config.yaml
rm -f /tmp/musicbox-beets-config.yaml
runuser -u music -- env HOME=/srv/music uv tool install --force "beets==$BEETS_VERSION"
ln -sf /srv/music/.local/bin/beet /usr/local/bin/beet
REMOTE
fi

echo "musicbox deploy: done."
