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
  first_line="$(head -n1 "$f" 2>/dev/null)"
  case "$first_line" in
  # A bash shebang, or a shebang-less file whose own shellcheck directive
  # declares it a bash library meant to be sourced.
  '#!'*bash* | '# shellcheck shell=bash'*)
    bash_scripts+=("$f")
    ;;
  # Some other interpreter's script; not ours to shellcheck.
  '#!'*) ;;
  # No shebang and no directive, so the name is the last thing left to go on.
  *)
    case "$f" in
    *.sh) bash_scripts+=("$f") ;;
    esac
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

echo "== docker compose config =="
if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
  if ! docker compose -f compose.yaml config -q; then
    status=1
  fi
else
  echo "check.sh: docker (or the compose plugin) not found; skipping compose.yaml validation." >&2
fi

echo "== systemd unit verification =="
# systemd-analyze verify insists every ExecStart= binary already exists on
# THIS machine, which is never true here: the paths are box-only
# (/usr/local/bin/music-import etc., installed by deploy.sh, never present
# on the workstation). That specific complaint is expected and filtered out
# by pattern; any other verify output (bad section, missing directive, bad
# OnCalendar syntax, and so on) still fails the gate. This is the
# pattern-filtered `systemd-analyze verify` approach, not a grep-based
# fallback: `--recursive-errors=no` alone does not suppress the
# missing-executable complaint, so the filter is the actual mechanism doing
# the work.
if command -v systemd-analyze >/dev/null 2>&1; then
  unit_files=(systemd/*.service systemd/*.timer)
  verify_output="$(systemd-analyze verify --recursive-errors=no "${unit_files[@]}" 2>&1 || true)"
  unexpected="$(grep -v -E ': Command /usr/local/bin/[A-Za-z0-9_.-]+ is not executable: No such file or directory$' <<<"$verify_output" || true)"
  if [ -n "$unexpected" ]; then
    echo "$unexpected" >&2
    status=1
  else
    echo "check.sh: systemd-analyze verify clean, modulo the expected box-only ExecStart paths."
  fi
else
  echo "check.sh: systemd-analyze not found; skipping unit verification." >&2
fi

if [ "$status" -eq 0 ]; then
  echo "check.sh: all checks passed."
else
  echo "check.sh: one or more checks failed." >&2
fi
exit "$status"
