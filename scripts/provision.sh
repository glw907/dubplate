#!/usr/bin/env bash
set -euo pipefail

# Provisions the musicbox Hetzner infrastructure: a firewall, a data volume, and
# the server itself. Every resource is check-before-create via `hcloud ... list`,
# so re-running after a partial failure only creates what is still missing.
# This script is the only thing in the repo that talks to hcloud.

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
USER_DATA="$REPO_ROOT/cloud-init/user-data.yaml"

# Hetzner resource names keep the musicbox name deliberately: they label
# live infrastructure, not the project (README.md "Kept identifiers").
FIREWALL_NAME="musicbox-fw"
FIREWALL_RULES_FILE="$REPO_ROOT/scripts/firewall-rules.json"
VOLUME_NAME="musicbox-data"
VOLUME_SIZE_GB=50
SERVER_NAME="musicbox"
SERVER_TYPE="cpx21"
LOCATION="hil"
PREFERRED_IMAGE="alma-10"
FALLBACK_IMAGE="alma-9"
SSH_KEY_NAME="musicbox-workstation"
SSH_PUBKEY_FILE="$HOME/.ssh/id_ed25519.pub"

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

# M11: firewall-rules.json ships with an empty source_ips placeholder for the
# tcp/22 rule; the real allowlist never lives in the repo. It comes from
# FIREWALL_SSH_CIDRS (comma-separated CIDRs, e.g. "203.0.113.4/32") at apply
# time via render_firewall_rules() below. Refusing outright when it is unset,
# rather than falling back to 0.0.0.0/0 + ::/0, is the point of this task:
# an unset var must never silently produce an open rule.
#
# Recovery path if the workstation IP rotates and SSH access to the box is
# lost: the Hetzner API is reachable from any machine holding HCLOUD_TOKEN,
# so the rule can always be re-widened from there, no SSH needed:
#   hcloud firewall replace-rules musicbox-fw --rules-file <(echo \
#     '[{"direction":"in","protocol":"tcp","port":"22","source_ips":["0.0.0.0/0","::/0"],"destination_ips":[]}]')
# Then re-run this script with a corrected FIREWALL_SSH_CIDRS to narrow it
# back down.
if [ -z "${FIREWALL_SSH_CIDRS:-}" ]; then
  echo "FIREWALL_SSH_CIDRS is not set. Refusing to apply the firewall step." >&2
  echo "Set it to a comma-separated list of CIDRs allowed to reach tcp/22 (e.g. FIREWALL_SSH_CIDRS=\"203.0.113.4/32\"), in ~/.local/secrets or the environment, then re-run." >&2
  echo "This script will not fall back to an open 0.0.0.0/0 + ::/0 rule." >&2
  exit 1
fi

if [ ! -f "$SSH_PUBKEY_FILE" ]; then
  echo "Missing SSH public key: $SSH_PUBKEY_FILE" >&2
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

# Renders FIREWALL_SSH_CIDRS (comma-separated, whitespace around entries
# tolerated) as a compact JSON array, e.g. FIREWALL_SSH_CIDRS="1.2.3.4/32, ::1/128"
# becomes ["1.2.3.4/32","::1/128"]. Exits with a clear message if every entry
# is blank, so a malformed value (e.g. a bare comma) never renders as an
# empty, effectively-no-one-allowed source_ips list without saying so.
#
# Callers run this in a command substitution, a subshell where bash cannot
# re-arm errexit, so a set -e placed inside does nothing. A best-effort call
# here is enforced by an if guard at the definition.
ssh_cidrs_json() {
  local raw=() cidr trimmed cidrs=()
  IFS=',' read -ra raw <<<"$FIREWALL_SSH_CIDRS"
  for cidr in "${raw[@]}"; do
    trimmed="${cidr#"${cidr%%[![:space:]]*}"}"
    trimmed="${trimmed%"${trimmed##*[![:space:]]}"}"
    [ -n "$trimmed" ] && cidrs+=("$trimmed")
  done
  if [ "${#cidrs[@]}" -eq 0 ]; then
    echo "FIREWALL_SSH_CIDRS is set but contains no usable CIDR (got: '$FIREWALL_SSH_CIDRS'). Fix it and re-run." >&2
    exit 1
  fi
  printf '%s\n' "${cidrs[@]}" | jq -R . | jq -s -c .
}

# Fills firewall-rules.json's placeholder tcp/22 source_ips with the CIDRs
# from FIREWALL_SSH_CIDRS and writes the effective rules to stdout. Nothing
# in the repo ever states the real allowlist; this is computed fresh from
# the environment every run.
render_firewall_rules() {
  jq --argjson cidrs "$(ssh_cidrs_json)" \
    '(.[] | select(.port == "22")).source_ips = $cidrs' \
    "$FIREWALL_RULES_FILE"
}

# True only if an inbound tcp/22 rule already matches FIREWALL_SSH_CIDRS
# exactly, so an existing firewall is never assumed correct just because it
# exists, and a stale wider (or narrower) rule is caught rather than skipped.
# The describe call is captured separately from the jq check: piping it
# straight into jq would make a network/API failure indistinguishable from
# "rule absent" (both make the pipeline exit nonzero), which would report a
# fixable connectivity blip as the firewall needing manual repair.
firewall_rule_present() {
  local describe_output expected_cidrs
  if ! describe_output="$(hcloud firewall describe "$FIREWALL_NAME" -o json)"; then
    echo "Failed to describe firewall $FIREWALL_NAME via the hcloud API; cannot verify its rules. Check network/API access and re-run." >&2
    exit 1
  fi
  expected_cidrs="$(ssh_cidrs_json)"
  jq -e --argjson expected "$expected_cidrs" '
    [.rules[]? | select(
      .direction == "in" and
      .protocol == "tcp" and
      .port == "22" and
      (((.source_ips // []) | sort) == ($expected | sort))
    )] | length > 0
  ' >/dev/null <<<"$describe_output"
}

volume_exists() {
  hcloud volume list -o columns=name | grep -qx "$VOLUME_NAME"
}

server_exists() {
  hcloud server list -o columns=name | grep -qx "$SERVER_NAME"
}

ssh_key_exists() {
  hcloud ssh-key list -o columns=name | grep -qx "$SSH_KEY_NAME"
}

image="$(resolve_image)"

if [ "$check_only" = true ]; then
  echo "musicbox provisioning plan (--check: nothing created)"
  echo "  image:    $image"
  if ssh_key_exists; then
    echo "  ssh-key:  $SSH_KEY_NAME (exists)"
  else
    echo "  ssh-key:  $SSH_KEY_NAME (would create from $SSH_PUBKEY_FILE)"
  fi
  if firewall_exists; then
    if firewall_rule_present; then
      echo "  firewall: $FIREWALL_NAME (exists, rule matches FIREWALL_SSH_CIDRS)"
    else
      echo "  firewall: $FIREWALL_NAME (exists, tcp/22 rule MISSING or does not match FIREWALL_SSH_CIDRS)"
    fi
  else
    echo "  firewall: $FIREWALL_NAME (would create: inbound tcp/22 from FIREWALL_SSH_CIDRS=$FIREWALL_SSH_CIDRS)"
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

if ssh_key_exists; then
  echo "SSH key $SSH_KEY_NAME already exists; skipping."
else
  # Registered up front so `hcloud server create` can attach it at creation
  # time below. A key present at create time makes Hetzner skip minting a
  # one-time root password entirely, so there is nothing to expire.
  echo "Creating SSH key $SSH_KEY_NAME from $SSH_PUBKEY_FILE..."
  hcloud ssh-key create --name "$SSH_KEY_NAME" --public-key-from-file "$SSH_PUBKEY_FILE"
fi

if firewall_exists; then
  if firewall_rule_present; then
    echo "Firewall $FIREWALL_NAME already exists with a tcp/22 rule matching FIREWALL_SSH_CIDRS; skipping."
  else
    echo "Firewall $FIREWALL_NAME exists but its tcp/22 rule does not match FIREWALL_SSH_CIDRS." >&2
    echo "Fix by hand (hcloud firewall replace-rules) and re-run; this script never rewrites an existing firewall's rules." >&2
    exit 1
  fi
else
  echo "Creating firewall $FIREWALL_NAME..."
  # Rules load at create time, not via a separate add-rule call: a failure
  # between create and add-rule would otherwise leave a rules-less firewall
  # that (per Hetzner's default) denies all inbound traffic, including SSH.
  # The rules file is rendered from FIREWALL_SSH_CIDRS via process
  # substitution; firewall-rules.json's own source_ips placeholder is never
  # passed to hcloud unfilled.
  hcloud firewall create --name "$FIREWALL_NAME" --rules-file <(render_firewall_rules)
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
    --ssh-key "$SSH_KEY_NAME" \
    --firewall "$FIREWALL_NAME" \
    --volume "$VOLUME_NAME" \
    --user-data-from-file "$USER_DATA"
fi

echo "Server IPv4:"
hcloud server ip "$SERVER_NAME"
