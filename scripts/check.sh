#!/usr/bin/env bash
set -euo pipefail
# check.sh: the repo gate. `make check` covers the Go module (vet,
# golangci-lint, test, tidy-check; it does not build the binary, that
# is `make build`'s own job); the web leg rejoins at rung 7, when
# web/ is rebuilt against the greenfield server. shellcheck still
# covers the scripts that remain (deploy.sh, check.sh itself,
# vale-comments.sh, provision.sh), ruff covers the oracle shim's
# Python, and gitleaks scans the whole tree (the corpus under
# corpus/cases/ is the repo's first large data directory, and a check
# either gates or does not exist). Exits nonzero on any failure so it
# can be wired into CI or a pre-push hook unmodified.

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

status=0

if ! command -v go >/dev/null 2>&1; then
  echo "check.sh: go not found; install it." >&2
  exit 1
fi
if ! command -v golangci-lint >/dev/null 2>&1; then
  echo "check.sh: golangci-lint not found; install it." >&2
  exit 1
fi
if ! command -v shellcheck >/dev/null 2>&1; then
  echo "check.sh: shellcheck not found; install it (brew install shellcheck)." >&2
  exit 1
fi
if ! command -v ruff >/dev/null 2>&1; then
  echo "check.sh: ruff not found; install it (uv tool install ruff)." >&2
  exit 1
fi
if ! command -v gitleaks >/dev/null 2>&1; then
  echo "check.sh: gitleaks not found; install it (brew install gitleaks)." >&2
  exit 1
fi

echo "== make check (Go) =="
if ! make check; then
  status=1
fi

echo "== shellcheck =="
# -P scripts resolves any script's relative `source=` directive against the
# scripts/ directory regardless of the caller's own working directory (no
# remaining script sources another one, now that lib.sh is gone, but the
# flag costs nothing to keep).
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

echo "== ruff (Python) =="
# The repo's only Python is the beets oracle shim
# (internal/harness/oracle/derive_oracle.py), which the gate lints but
# never runs: running it would need the venv, and the gate never
# requires uv or beets. Tracked files only, so an unstaged experiment
# does not gate; ruff.toml carries the rule set.
mapfile -t py_files < <(git ls-files '*.py')
if [ "${#py_files[@]}" -gt 0 ]; then
  if ! ruff check "${py_files[@]}"; then
    status=1
  fi
  if ! ruff format --check "${py_files[@]}"; then
    status=1
  fi
else
  echo "check.sh: no tracked Python files found; the oracle shim should be one." >&2
  status=1
fi

echo "== gitleaks (secret scan) =="
# --no-git scans the tracked worktree as it sits, not the full git
# history (a separate, heavier job): the corpus under corpus/cases/ is
# the repo's first large data directory, and every response body in
# it is verbatim wire data from a live service, so this is the rail
# that makes CLAUDE.md's "eyeball the diff for stray personal values"
# instruction a backstop rather than the only net. .gitleaksignore
# carries the one known false positive (a config-key-name test
# fixture that happens to look like a generic API key).
if ! gitleaks detect --no-git -s . --redact; then
  status=1
fi

echo "== vale (Go comment prose) =="
if ! scripts/vale-comments.sh; then
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
# (/usr/local/bin/dubplate, installed by deploy.sh, never present on the
# workstation). That specific complaint is expected and filtered out
# by pattern; any other verify output (bad section, missing directive, bad
# OnCalendar syntax, and so on) still fails the gate. This is the
# pattern-filtered `systemd-analyze verify` approach, not a grep-based
# fallback: `--recursive-errors=no` alone does not suppress the
# missing-executable complaint, so the filter is the actual mechanism doing
# the work.
if command -v systemd-analyze >/dev/null 2>&1; then
  unit_files=(systemd/*.service systemd/*.timer systemd/*.path)
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
