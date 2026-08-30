#!/usr/bin/env bash
set -euo pipefail

# Provisions the musicbox Hetzner infrastructure: a firewall, a data volume, and
# the server itself. Every resource is check-before-create via `hcloud ... list`,
# so re-running after a partial failure only creates what is still missing.
# This script is the only thing in the repo that talks to hcloud.

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
USER_DATA="$REPO_ROOT/cloud-init/user-data.yaml"

FIREWALL_NAME="musicbox-fw"
FIREWALL_RULES_FILE="$REPO_ROOT/scripts/firewall-rules.json"
VOLUME_NAME="musicbox-data"
VOLUME_SIZE_GB=50
SERVER_NAME="musicbox"
SERVER_TYPE="cpx21"
LOCATION="hil"
PREFERRED_IMAGE="alma-10"
FALLBACK_IMAGE="alma-9"

check_only=false
for arg in "$@"; do
  case "$arg" in
    --check)
      check_only=true
      ;;
    *)
      echo "Unknown argument: $arg" >&2
      exit 1
      ;;
  esac
done

# HCLOUD_TOKEN originates in the age store (Task 0), never in this repo; source
# the rendered file the way every other workstation script does.
if [ -f "$HOME/.local/secrets" ]; then
  # shellcheck disable=SC1091
  source "$HOME/.local/secrets"
fi

if [ -z "${HCLOUD_TOKEN:-}" ]; then
  echo "HCLOUD_TOKEN is not set. Mint it via secret-set.sh (plan Task 0), then re-source ~/.local/secrets." >&2
  exit 1
fi
export HCLOUD_TOKEN

if ! command -v hcloud >/dev/null 2>&1; then
  echo "hcloud CLI not found. Install it: brew install hcloud" >&2
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "jq not found; required to verify firewall rules. Install it: brew install jq" >&2
  exit 1
fi

if [ ! -f "$USER_DATA" ]; then
  echo "Missing cloud-init file: $USER_DATA" >&2
  exit 1
fi

if [ ! -f "$FIREWALL_RULES_FILE" ]; then
  echo "Missing firewall rules file: $FIREWALL_RULES_FILE" >&2
  exit 1
fi

# alma-10 may not exist on Hetzner yet; fall back to alma-9 and say so loudly,
# because Task 4 needs to know which image actually ran, recorded in STATUS.
resolve_image() {
  local available
  available="$(hcloud image list -o columns=name)"
  if grep -qx "$PREFERRED_IMAGE" <<<"$available"; then
    echo "$PREFERRED_IMAGE"
  elif grep -qx "$FALLBACK_IMAGE" <<<"$available"; then
    echo "NOTE: $PREFERRED_IMAGE is unavailable on Hetzner; falling back to $FALLBACK_IMAGE. Record this in docs/STATUS.md." >&2
    echo "$FALLBACK_IMAGE"
  else
    echo "Neither $PREFERRED_IMAGE nor $FALLBACK_IMAGE is available. Run 'hcloud image list' and update this script." >&2
    exit 1
  fi
}

firewall_exists() {
  hcloud firewall list -o columns=name | grep -qx "$FIREWALL_NAME"
}

# True only if an inbound tcp/22 rule already covers both 0.0.0.0/0 and ::/0,
# so an existing firewall is never assumed correct just because it exists.
# The describe call is captured separately from the jq check: piping it
# straight into jq would make a network/API failure indistinguishable from
# "rule absent" (both make the pipeline exit nonzero), which would report a
# fixable connectivity blip as the firewall needing manual repair.
firewall_rule_present() {
  local describe_output
  if ! describe_output="$(hcloud firewall describe "$FIREWALL_NAME" -o json)"; then
    echo "Failed to describe firewall $FIREWALL_NAME via the hcloud API; cannot verify its rules. Check network/API access and re-run." >&2
    exit 1
  fi
  jq -e '
    [.rules[]? | select(
      .direction == "in" and
      .protocol == "tcp" and
      .port == "22" and
      ((.source_ips // []) | contains(["0.0.0.0/0"])) and
      ((.source_ips // []) | contains(["::/0"]))
    )] | length > 0
  ' >/dev/null <<<"$describe_output"
}

volume_exists() {
  hcloud volume list -o columns=name | grep -qx "$VOLUME_NAME"
}

server_exists() {
  hcloud server list -o columns=name | grep -qx "$SERVER_NAME"
}

image="$(resolve_image)"

if [ "$check_only" = true ]; then
  echo "musicbox provisioning plan (--check: nothing created)"
  echo "  image:    $image"
  if firewall_exists; then
    if firewall_rule_present; then
      echo "  firewall: $FIREWALL_NAME (exists, rule present)"
    else
      echo "  firewall: $FIREWALL_NAME (exists, MISSING the tcp/22 rule)"
    fi
  else
    echo "  firewall: $FIREWALL_NAME (would create: inbound tcp/22 from 0.0.0.0/0 and ::/0)"
  fi
  if volume_exists; then
    echo "  volume:   $VOLUME_NAME (exists)"
  else
    echo "  volume:   $VOLUME_NAME (would create: ${VOLUME_SIZE_GB}GB in $LOCATION)"
  fi
  if server_exists; then
    echo "  server:   $SERVER_NAME (exists)"
  else
    echo "  server:   $SERVER_NAME (would create: $SERVER_TYPE, $image, $LOCATION, firewall+volume attached)"
  fi
  exit 0
fi

if firewall_exists; then
  if firewall_rule_present; then
    echo "Firewall $FIREWALL_NAME already exists with the expected rule; skipping."
  else
    echo "Firewall $FIREWALL_NAME exists but is missing the tcp/22 rule from $FIREWALL_RULES_FILE." >&2
    echo "Fix by hand (hcloud firewall replace-rules) and re-run; this script never rewrites an existing firewall's rules." >&2
    exit 1
  fi
else
  echo "Creating firewall $FIREWALL_NAME..."
  # Rules load at create time, not via a separate add-rule call: a failure
  # between create and add-rule would otherwise leave a rules-less firewall
  # that (per Hetzner's default) denies all inbound traffic, including SSH.
  hcloud firewall create --name "$FIREWALL_NAME" --rules-file "$FIREWALL_RULES_FILE"
fi

if volume_exists; then
  echo "Volume $VOLUME_NAME already exists; skipping."
else
  echo "Creating volume $VOLUME_NAME (${VOLUME_SIZE_GB}GB, $LOCATION)..."
  # No --format flag: cloud-init formats on first boot, guarded by a blkid
  # check, so a re-run never reformats a volume that already holds data.
  hcloud volume create \
    --name "$VOLUME_NAME" \
    --size "$VOLUME_SIZE_GB" \
    --location "$LOCATION"
fi

if server_exists; then
  echo "Server $SERVER_NAME already exists; skipping create."
else
  echo "Creating server $SERVER_NAME ($SERVER_TYPE, $image, $LOCATION)..."
  hcloud server create \
    --name "$SERVER_NAME" \
    --type "$SERVER_TYPE" \
    --image "$image" \
    --location "$LOCATION" \
    --firewall "$FIREWALL_NAME" \
    --volume "$VOLUME_NAME" \
    --user-data-from-file "$USER_DATA"
fi

echo "Server IPv4:"
hcloud server ip "$SERVER_NAME"
