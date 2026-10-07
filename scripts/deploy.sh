#!/usr/bin/env bash
set -euo pipefail

# Deploys dubplate-server to the box: renders /etc/dubplate/server.env from
# env/dubplate-server.env.template plus the workstation age-store values,
# cross-compiles and ships the dubplate-server binary, and installs its
# systemd unit. Idempotent: re-running against an unchanged box is a no-op
# diff (rsync's own reporting, plus the env-render and directory steps
# being naturally idempotent).
#
# Usage:
#   deploy.sh [--check | --apply] [HOST]
# HOST defaults to `hcloud server ip musicbox` (the box's Hetzner server
# name, unrenamed) when omitted. With no flag, or with --check, the script
# diffs and reports every step without applying any of them, and without
# ever writing decrypted secret values to a log or a local file. Only
# --apply writes the box, so a dropped flag is safe. --apply and --check
# together are refused.
#
# The local half runs in every mode, before the first remote write: the
# script refuses a working tree with a modified tracked file, or with an
# untracked or ignored file under any build input (BUILD_INPUTS below),
# then runs npm ci and the Vite build under web/ and cross-compiles the
# stamped binary, each under an environment holding only PATH and HOME. A
# dirty tree or a failed build aborts before anything reaches the box.
#
# --apply writes exactly these, in order:
#   - rsyncs every unit in UNIT_FILES to /etc/systemd/system;
#   - creates /var/lib/dubplate owned by music, and the auth, store, and
#     unidecode leaves under it;
#   - stages the dubplate-server binary, smoke-tests it, runs its unidecode
#     install on the box (a PyPI download), and swaps it onto
#     /usr/local/bin/dubplate-server;
#   - installs /etc/dubplate/server.env;
#   - runs systemctl daemon-reload.
# It never enables, starts, or restarts a unit.

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SERVER_ENV_TEMPLATE="$REPO_ROOT/env/dubplate-server.env.template"
REMOTE_SERVER_ENV="/etc/dubplate/server.env"

UNIT_FILES=(
  dubplate-server-serve.service
)

# ControlMaster/ControlPersist share one TCP+SSH handshake across this run's
# many ssh/rsync calls instead of paying full connection setup for each; the
# socket is cleaned up automatically 60s after the last call finishes.
SSH_OPTS=(-o StrictHostKeyChecking=accept-new -o ConnectTimeout=10 \
  -o ControlMaster=auto -o ControlPersist=60s -o "ControlPath=$HOME/.ssh/dubplate-deploy-%r@%h:%p")

check_only=true
check_flag=false
apply_flag=false
HOST=""
for arg in "$@"; do
  case "$arg" in
  --check)
    check_flag=true
    ;;
  --apply)
    apply_flag=true
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

if $check_flag && $apply_flag; then
  echo "deploy.sh: --check and --apply are mutually exclusive; pass one, or neither to check." >&2
  exit 1
fi
if $apply_flag; then
  check_only=false
fi

# --- workstation secrets ---------------------------------------------
# HCLOUD_TOKEN (to look up HOST) and every value the server env template's
# bare rows name originate here (via secret-set.sh), never in this repo.
if [ -f "$HOME/.local/secrets" ]; then
  # shellcheck disable=SC1091
  source "$HOME/.local/secrets"
fi

# --- resolve target host ----------------------------------------------
if [ -z "$HOST" ]; then
  if [ -z "${HCLOUD_TOKEN:-}" ] || ! command -v hcloud >/dev/null 2>&1; then
    echo "deploy.sh: no HOST given and can't look one up (need HCLOUD_TOKEN + hcloud). Pass an IP, or provision first." >&2
    exit 1
  fi
  HOST="$(HCLOUD_TOKEN="$HCLOUD_TOKEN" hcloud server ip musicbox 2>/dev/null || true)"
  if [ -z "$HOST" ]; then
    echo "deploy.sh: couldn't resolve the musicbox server's IP via hcloud. Pass an IP, or provision first." >&2
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

# --- render the env template, in memory only ---------------------------
# render_missing accumulates every unresolvable row so one run reports
# them together, never one at a time.
render_missing=()

# render_template TEMPLATE OUT_ARRAY_NAME
# Resolves one line-oriented template into the array named by
# OUT_ARRAY_NAME (bash nameref), entirely in memory. Three row states
# (the template's own header names all three):
#   - blank or '#'-comment: skipped.
#   - NAME=value (non-empty right-hand side): a literal, used verbatim
#     and never looked up in the environment, even when a variable of
#     that name happens to be exported.
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

# SC2034: the array is only ever written and read through render_template's
# and install_env_step's own `local -n` namerefs, passed by name; shellcheck
# does not follow that indirection.
# shellcheck disable=SC2034
server_env_lines=()
render_template "$SERVER_ENV_TEMPLATE" server_env_lines

if [ "${#render_missing[@]}" -gt 0 ]; then
  echo "deploy.sh: missing a value for the following variables (mint them via secret-set.sh, then re-source ~/.local/secrets):" >&2
  printf '  %s\n' "${render_missing[@]}" >&2
  exit 1
fi

if $check_only; then
  echo "dubplate deploy -> $SSH_TARGET (check: nothing will be applied; pass --apply to write the box)"
else
  echo "dubplate deploy -> $SSH_TARGET (APPLY: writing the box)"
fi

# --- local builds: committed source only, built before any remote write --
# `git describe` below would stamp a reviewed commit's id on
# unreviewed code, so a modified tracked file anywhere refuses the run
# before the first remote write. BUILD_INPUTS names every path the binary
# is built from: an untracked .go or migration file under cmd/ or
# internal/ is compiled in, and a file under web/ is embedded by
# go:embed. So an untracked or an ignored file under any of them refuses
# the run too. `--ignored` with no mode is git-status(1)'s "traditional"
# mode, which reports an ignored directory as one line, and the three
# excluded paths are the build's own output: npm ci's node_modules, kit's
# generated .svelte-kit, and adapter-static's build directory.
BUILD_INPUTS=(cmd internal web go.mod go.sum)
BUILD_OUTPUTS=(':(exclude)web/node_modules' ':(exclude)web/.svelte-kit' ':(exclude)web/build')
dirty_tracked="$(git -C "$REPO_ROOT" status --porcelain --untracked-files=no)"
dirty_inputs="$(git -C "$REPO_ROOT" status --porcelain --ignored -- "${BUILD_INPUTS[@]}" "${BUILD_OUTPUTS[@]}")"
if [ -n "$dirty_tracked" ] || [ -n "$dirty_inputs" ]; then
  echo "deploy.sh: refusing to deploy a tree that differs from its commit:" >&2
  printf '%s\n%s\n' "$dirty_tracked" "$dirty_inputs" | sed '/^$/d' | sort -u | sed 's/^/  /' >&2
  exit 1
fi

# The stamp pairs the human-facing version (`git describe`) with the exact
# commit the box's Go matcher runs (`git rev-parse HEAD`, independent of
# tags). main.go's matcherCommit stays "dev" on any build that skips these
# -X flags, so a bare `go build` or `make build-linux` outside this script
# never reaches the box stamped.
DUBPLATE_STAMP_VERSION="$(git -C "$REPO_ROOT" describe --tags --always)"
DUBPLATE_STAMP_MATCHER_COMMIT="$(git -C "$REPO_ROOT" rev-parse HEAD)"

# `vite build` alone: `make web` would run the gates a second time. Nothing
# Node ships to the box.
#
# This shell has sourced ~/.local/secrets above, and Vite, its plugins,
# and npm all run third-party code that can read the environment it
# inherits. So every local build runs under BUILD_ENV, which hands the
# child PATH and HOME and nothing else: npm and Go find their caches
# under HOME, and no workstation token or rendered secret reaches them.
# --ignore-scripts keeps npm from running any package's install script
# (npm's config docs: "npm does not run scripts specified in package.json
# files"); the lockfile's only one belongs to fsevents, which is
# macOS-only. `npm run build` still runs the build script it names.
BUILD_ENV=(env -i "PATH=$PATH" "HOME=$HOME")
echo "== local build: npm ci --ignore-scripts --prefix web =="
"${BUILD_ENV[@]}" npm ci --ignore-scripts --prefix "$REPO_ROOT/web"
echo "== local build: npm run build --prefix web =="
"${BUILD_ENV[@]}" npm run build --prefix "$REPO_ROOT/web"
echo "== local build: make build-linux =="
"${BUILD_ENV[@]}" make -C "$REPO_ROOT" build-linux LDFLAGS="-X main.version=$DUBPLATE_STAMP_VERSION -X main.matcherCommit=$DUBPLATE_STAMP_MATCHER_COMMIT"
DUBPLATE_SERVER_BINARY="$REPO_ROOT/dubplate-server-linux-amd64"

# --- units -> /etc/systemd/system ---------------------------------------
echo "== units (${UNIT_FILES[*]}) -> $SSH_TARGET:/etc/systemd/system =="
if $check_only; then
  rsync -avzn -e "$RSYNC_SSH" "${UNIT_FILES[@]/#/$REPO_ROOT/systemd/}" "$SSH_TARGET:/etc/systemd/system/" || true
else
  rsync -avz -e "$RSYNC_SSH" "${UNIT_FILES[@]/#/$REPO_ROOT/systemd/}" "$SSH_TARGET:/etc/systemd/system/"
fi

# --- /var/lib/dubplate (DUBPLATE_STATE_DIR) -------------------------------
# The state directory every leaf below sits in. music:music matches the
# serve unit's own User=music, and the binary's own table install below
# needs the directory to exist, owned by music, before it runs.
DUBPLATE_LIB_DIR="/var/lib/dubplate"
echo "== $DUBPLATE_LIB_DIR =="
run_remote "install -d -o music -g music -m 0755 $DUBPLATE_LIB_DIR"

# --- every leaf a dubplate-server unit is granted -------------------------
# DUBPLATE_SERVER_LEAVES names every leaf under $DUBPLATE_LIB_DIR that a
# unit running dubplate-server is granted in ReadWritePaths=, and
# scripts/check.sh holds the array, the units, and the env template in
# agreement. Each leaf must exist before its unit starts, or the unit fails
# outright with "Failed to set up mount namespacing" under
# ProtectSystem=strict; nothing inside the process gets a chance to mkdir
# it:
#   auth    dubplate-server-serve.service (sign-in nonce and epoch stores)
#   store   dubplate-server-serve.service (the member and settings store)
# The unidecode leaf is not listed: no unit is granted it, and it is
# pre-created on its own just below. 0700 matches what auth.Open requires
# of its leaf (and leafDirPerm in internal/engine/store of the store leaf)
# and keeps the members private to music.
DUBPLATE_SERVER_LEAVES=(auth store)
for leaf in "${DUBPLATE_SERVER_LEAVES[@]}"; do
  echo "== $DUBPLATE_LIB_DIR/$leaf =="
  run_remote "install -d -o music -g music -m 0700 $DUBPLATE_LIB_DIR/$leaf"
done

# --- /var/lib/dubplate/unidecode (the table install below) ---------------
# The table install's systemd-run grants this leaf alone, so the verb can
# write its table and nothing else under $DUBPLATE_LIB_DIR: the auth store
# and the member store sit beside it. A granted path must exist before the
# transient unit starts (the same namespace failure as the leaves above),
# so it is created here. 0755 and music's ownership are what the verb's
# own leaf check requires of an existing leaf (unidecodeLeafPerm and
# checkUnidecodeOwner in cmd/dubplate-server/unidecode.go), and on a box
# whose leaf already has that shape this line changes nothing.
DUBPLATE_UNIDECODE_DIR="$DUBPLATE_LIB_DIR/unidecode"
echo "== $DUBPLATE_UNIDECODE_DIR =="
run_remote "install -d -o music -g music -m 0755 $DUBPLATE_UNIDECODE_DIR"

# --- dubplate-server binary -> /usr/local/bin/dubplate-server -----------
# The cross-compiled, stamped binary is built above, before the first
# remote write, in both modes: it only writes the gitignored local file, and
# having it is what lets --check's rsync -n report a real diff instead of
# guessing from a stale local copy.
#
# The binary is staged at .new, not swapped onto the live path here: the
# staged copy installs its own transliteration table (below) and only a
# verified install swaps it onto /usr/local/bin/dubplate-server, so a unit
# starting mid-deploy always finds either the old binary or a new one that
# already found its table -- never a new one with none.
DUBPLATE_SERVER_STAGED="/usr/local/bin/dubplate-server.new"
echo "== dubplate-server binary -> $SSH_TARGET:$DUBPLATE_SERVER_STAGED =="
if $check_only; then
  # Named against the LIVE path, not the staged one: rsync -n against a
  # path that never exists on the box (.new is removed by the mv below on
  # every completed deploy) always reports a full transfer, which would
  # make --check claim a change on every run regardless of whether the
  # binary actually differs. Diffing the live path is the meaningful
  # comparison.
  rsync -avzn -e "$RSYNC_SSH" "$DUBPLATE_SERVER_BINARY" "$SSH_TARGET:/usr/local/bin/dubplate-server" || true
else
  rsync -avz -e "$RSYNC_SSH" "$DUBPLATE_SERVER_BINARY" "$SSH_TARGET:$DUBPLATE_SERVER_STAGED"
  run_remote "chmod 0755 $DUBPLATE_SERVER_STAGED"
fi

# Smoke test: a binary that rsynced and chmodded fine can still fail to
# execute (wrong GOARCH, a missing dynamic dependency the box lacks). The
# remote command string runs under the login shell with no `set -e`, so
# `output="$(...)"` alone would swallow a non-zero exit from the binary
# and fall through to the empty-commit branch below with the wrong
# diagnosis; the exec is therefore tested explicitly first, aborting the
# whole deploy with its own message rather than leaving a broken binary
# in place. Once the exec is known good, the assertion extends that abort
# to a binary that executes fine but never got stamped: it prints both
# "version" rows for the operator, then exits nonzero when matcher-commit
# is empty or "dev" (main.go's own workstation default).
echo "== smoke test: dubplate-server version =="
# Runs against the staged binary, not the live one, so a bad artifact is
# caught before it ever installs a table or gets swapped in.
# shellcheck disable=SC2016
smoke_cmd='if ! output="$(/usr/local/bin/dubplate-server version)"; then echo "dubplate-server version failed to execute" >&2; exit 1; fi; printf "%s\n" "$output"; commit=$(printf "%s\n" "$output" | awk "/^matcher-commit/{print \$2}"); if [ -z "$commit" ] || [ "$commit" = "dev" ]; then echo "matcher-commit is dev or empty; aborting deploy" >&2; exit 1; fi'
run_remote "${smoke_cmd/\/usr\/local\/bin\/dubplate-server/$DUBPLATE_SERVER_STAGED}"

# --- staged binary installs its transliteration table --------------------
# Runs before the swap below: a failed install (PyPI unreachable, a
# checksum refusal, a sandbox error) aborts the deploy here, with the old
# binary still live -- the window a same-time swap-then-install would leave
# open. systemd-run --wait propagates the verb's exit, so this line's own
# failure trips `set -euo pipefail` and stops the deploy before the mv.
# The hardening is dubplate-server-serve.service's own, property for
# property, with three differences: Environment= (not EnvironmentFile=)
# carries only DUBPLATE_STATE_DIR, since server.env is not installed on the
# box until the env step after the swap, and the verb needs nothing else;
# ReadWritePaths= grants the pre-created $DUBPLATE_UNIDECODE_DIR (not the
# serve unit's leaves), the only directory the verb writes; and UMask is
# omitted, since the verb sets the leaf's mode explicitly regardless of
# umask. AF_INET and AF_INET6 carry the download. This is a plain top-level
# command, never inside an `if` or on either side of `||`/`&&`, so `set -e`
# is not silently ignored on it.
echo "== $DUBPLATE_SERVER_STAGED unidecode install =="
run_remote "systemd-run --wait --pipe --collect \
  -p User=music -p Environment=DUBPLATE_STATE_DIR=$DUBPLATE_LIB_DIR \
  -p NoNewPrivileges=yes -p PrivateTmp=yes -p ProtectSystem=strict \
  -p ProtectHome=yes -p ProtectKernelTunables=yes \
  -p ProtectKernelModules=yes -p ProtectControlGroups=yes \
  -p RestrictNamespaces=yes -p RestrictRealtime=yes \
  -p RestrictSUIDSGID=yes -p LockPersonality=yes \
  -p MemoryDenyWriteExecute=yes -p SystemCallFilter=@system-service \
  -p SystemCallArchitectures=native -p AmbientCapabilities= \
  -p CapabilityBoundingSet= \
  -p 'RestrictAddressFamilies=AF_INET AF_INET6 AF_UNIX' \
  -p ReadWritePaths=$DUBPLATE_UNIDECODE_DIR \
  $DUBPLATE_SERVER_STAGED unidecode install"

# --- swap: only after the staged binary's table install verified --------
# A rename on one filesystem: a unit starting mid-mv sees either the old
# binary or the new one, and the new one already found its table.
echo "== swap $DUBPLATE_SERVER_STAGED -> /usr/local/bin/dubplate-server =="
run_remote "mv -f $DUBPLATE_SERVER_STAGED /usr/local/bin/dubplate-server"

# --- /etc/dubplate/server.env --------------------------------------------
# Not a file copy: this content is assembled in memory from secrets and
# literals and must never touch local disk (see the usage note at the top
# of this file).

# install_env_step REMOTE_PATH LINES_ARRAY_NAME
# Shared check/report/install shape for one rendered env array. --check
# probes reachability, then diffs the rendered lines against REMOTE_PATH's
# live content without ever printing content itself (it's secrets). Apply
# mode pipes the rendered lines straight into install(1) over stdin -- no
# staged copy on the box at all, and -D creates the parent directory
# (/etc/dubplate) on the box's first deploy; install itself no-ops that
# part on every later run.
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

# 0600 root:root: systemd reads EnvironmentFile= as PID 1 before dropping
# privileges, so this mode and ownership are not a default to widen once
# dubplate-server itself runs as a lower-privilege user.
install_env_step "$REMOTE_SERVER_ENV" server_env_lines

# --- systemd reload --------------------------------------------------
# Picks up the unit files rsynced above. Enabling or restarting the serve
# unit is the operator's own step, never this script's.
echo "== systemd daemon-reload =="
run_remote "systemctl daemon-reload"

echo "dubplate deploy: done."
