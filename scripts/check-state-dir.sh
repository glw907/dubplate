#!/usr/bin/env bash
set -euo pipefail
# check-state-dir.sh: the state-dir agreement leg of the repo gate, run by
# check.sh's full form and runnable alone from any directory in a few
# seconds. Exits nonzero when any copy of the state tree disagrees.
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

# DUBPLATE_STATE_DIR (env/dubplate-server.env.template), each
# dubplate-server unit's ReadWritePaths= leaves, and deploy.sh's
# DUBPLATE_LIB_DIR and DUBPLATE_SERVER_LEAVES are copies of one tree with
# nothing but this leg holding them together: an edit to one alone gives a
# unit that fails at start with "Failed to set up mount namespacing" (a
# leaf deploy.sh never creates) or at first write with EROFS under
# ProtectSystem=strict (see the derivation comments above ReadWritePaths=
# in the units).
template_state_dir="$(sed -n 's/^DUBPLATE_STATE_DIR=//p' env/dubplate-server.env.template)"
deploy_lib_dir="$(sed -n 's/^DUBPLATE_LIB_DIR="\(.*\)"$/\1/p' scripts/deploy.sh)"
leaves_line="$(sed -n 's/^DUBPLATE_SERVER_LEAVES=(\(.*\))$/\1/p' scripts/deploy.sh)"
read -r -a server_leaves <<<"$leaves_line"
echo "check-state-dir.sh: DUBPLATE_STATE_DIR=$template_state_dir DUBPLATE_LIB_DIR=$deploy_lib_dir DUBPLATE_SERVER_LEAVES=(${server_leaves[*]})"
state_ok=true
if [ -z "$template_state_dir" ] || [ "$deploy_lib_dir" != "$template_state_dir" ]; then
  echo "check-state-dir.sh: state-dir: DUBPLATE_LIB_DIR ($deploy_lib_dir) does not equal the template's DUBPLATE_STATE_DIR ($template_state_dir)" >&2
  state_ok=false
fi
if [ "${#server_leaves[@]}" -eq 0 ]; then
  echo "check-state-dir.sh: state-dir: scripts/deploy.sh has no DUBPLATE_SERVER_LEAVES=(...) line" >&2
  state_ok=false
fi
server_units=()
granted_leaves=()
serve_unit=systemd/dubplate-server-serve.service
serve_has_store=false
granted_leaves=()
serve_unit=systemd/dubplate-server-serve.service
serve_has_store=false
for unit in systemd/*.service; do
  if grep -Eq '^ExecStart=/usr/local/bin/dubplate-server( |$)' "$unit"; then
    server_units+=("$unit")
  fi
done
if [ "${#server_units[@]}" -eq 0 ]; then
  echo "check-state-dir.sh: state-dir: no unit under systemd/ has an ExecStart= naming /usr/local/bin/dubplate-server" >&2
  state_ok=false
fi
for unit in "${server_units[@]}"; do
  while read -r -a entries; do
    for entry in "${entries[@]}"; do
      case "$entry" in
        -* | +*)
          echo "check-state-dir.sh: state-dir: $unit: ReadWritePaths= entry $entry carries a prefix (a missing leaf would start and fail at first write)" >&2
          state_ok=false
          continue
          ;;
      esac
      if [ "$entry" = "$template_state_dir" ]; then
        continue
      fi
      leaf="${entry#"$template_state_dir"/}"
      if [ "$leaf" = "$entry" ] || [[ "$leaf" == */* ]] || [ -z "$leaf" ]; then
        echo "check-state-dir.sh: state-dir: $unit: ReadWritePaths= entry $entry is not a leaf directly under $template_state_dir" >&2
        state_ok=false
        continue
      fi
      granted_leaves+=("$leaf")
      if [ "$unit" = "$serve_unit" ] && [ "$leaf" = store ]; then
        serve_has_store=true
      fi
      found=false
      for known in "${server_leaves[@]}"; do
        if [ "$known" = "$leaf" ]; then
          found=true
        fi
      done
      if ! $found; then
        echo "check-state-dir.sh: state-dir: $unit: ReadWritePaths= leaf $leaf is not in DUBPLATE_SERVER_LEAVES (deploy.sh would not create it)" >&2
        state_ok=false
      fi
    done
  done < <(sed -n 's/^ReadWritePaths=//p' "$unit")
done
loop_body="$(awk '/^for [a-z_]+ in "\$\{DUBPLATE_SERVER_LEAVES\[@\]\}"; do$/ {on=1; next} on && /^done$/ {exit} on {print}' scripts/deploy.sh)"
if [ -z "$loop_body" ]; then
  echo "check-state-dir.sh: state-dir: scripts/deploy.sh has no loop iterating \"\${DUBPLATE_SERVER_LEAVES[@]}\"" >&2
  state_ok=false
elif ! grep -q 'install -d' <<<"$loop_body"; then
  echo "check-state-dir.sh: state-dir: the DUBPLATE_SERVER_LEAVES loop in scripts/deploy.sh never runs install -d" >&2
  state_ok=false
fi
if ! grep -qF "run_remote \"install -d -o music -g music -m 0755 \$DUBPLATE_LIB_DIR\"" scripts/deploy.sh; then
  echo "check-state-dir.sh: state-dir: scripts/deploy.sh has no pre-create line for \$DUBPLATE_LIB_DIR" >&2
  state_ok=false
fi
# The reverse of the unit-to-array check above: a leaf the array pre-creates
# that no dubplate-server unit is granted is a stale entry (a renamed or
# retired leaf deploy.sh would keep creating).
for known in "${server_leaves[@]}"; do
  granted=false
  for leaf in "${granted_leaves[@]}"; do
    if [ "$known" = "$leaf" ]; then
      granted=true
    fi
  done
  if ! $granted; then
    echo "check-state-dir.sh: state-dir: DUBPLATE_SERVER_LEAVES names $known, which no dubplate-server unit is granted in ReadWritePaths=" >&2
    state_ok=false
  fi
done
# The serve unit opens the member and settings store, so it must be granted
# that leaf; without it the first write fails with EROFS.
if ! $serve_has_store; then
  echo "check-state-dir.sh: state-dir: $serve_unit is granted no store leaf under $template_state_dir in ReadWritePaths=" >&2
  state_ok=false
fi
if ! $state_ok; then
  exit 1
fi
echo "check-state-dir.sh: state-dir agreement holds."
