#!/usr/bin/env bash
set -euo pipefail
# check.sh: the repo gate. Shellcheck over every scripts/* bash file, then
# the full bats suite. Exits nonzero on any failure so it can be wired into
# CI or a pre-push hook unmodified.

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

status=0

if ! command -v shellcheck >/dev/null 2>&1; then
  echo "check.sh: shellcheck not found; install it (brew install shellcheck)." >&2
  exit 1
fi
if ! command -v bats >/dev/null 2>&1; then
  echo "check.sh: bats not found; install it (brew install bats-core)." >&2
  exit 1
fi

echo "== shellcheck =="
# -P scripts resolves each script's "source=./lib.sh" relative to the
# scripts/ directory regardless of the caller's own working directory.
bash_scripts=()
for f in scripts/*; do
  [ -f "$f" ] || continue
  case "$(head -c2 "$f" 2>/dev/null)" in
  '#!')
    head -n1 "$f" | grep -q "bash" && bash_scripts+=("$f")
    continue
    ;;
  esac
  # A shebang-less file can still be a bash library meant to be sourced:
  # either its own shellcheck directive says so, or its name says so.
  case "$f" in
  *.sh)
    bash_scripts+=("$f")
    ;;
  *)
    head -n1 "$f" | grep -q '^# shellcheck shell=bash' && bash_scripts+=("$f")
    ;;
  esac
done
if [ "${#bash_scripts[@]}" -gt 0 ]; then
  if ! shellcheck -x -P scripts "${bash_scripts[@]}"; then
    status=1
  fi
else
  echo "check.sh: no bash scripts found under scripts/." >&2
  status=1
fi

echo "== bats =="
if ! bats tests/; then
  status=1
fi

if [ "$status" -eq 0 ]; then
  echo "check.sh: all checks passed."
else
  echo "check.sh: one or more checks failed." >&2
fi
exit "$status"
