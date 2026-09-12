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

# A gate lock: only one check.sh runs per tree at a time. `make
# tidy-check` runs `go mod tidy` before `git diff --exit-code`, and
# that first half mutates go.mod and go.sum, so two concurrent gate
# runs in one tree race on those files. The lock lives under the tree's
# own git-dir, which is .git for the main tree and .git/worktrees/<name>
# for a worktree, so a tree serializes against itself and against no
# other tree. The non-blocking attempt comes first only so the waiting
# line prints once, under contention; the second flock then blocks
# until the holder finishes.
lock_file="$(git rev-parse --git-dir)/check.sh.lock"
exec 9>"$lock_file"
if ! flock -n 9; then
  echo "check.sh: another check.sh run holds this tree's gate lock; waiting..." >&2
  flock 9
fi

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
# instruction a backstop rather than the only net. The generic-api-key
# rule fires only when a key-like field name sits on the same line as
# a value whose entropy clears its threshold; exactly one of the
# tree's DUBPLATE_* names used to pair that way in a test fixture,
# and it is renamed rather than suppressed, so no .gitleaksignore
# entry remains.
# The repo's .gitleaks.toml keeps the built-in rules and allowlists
# one path, .claude/worktrees/; its own comments carry the reasoning.
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

echo "== state-dir agreement (env template, shadow unit, deploy.sh) =="
# DUBPLATE_STATE_DIR (env/dubplate-server.env.template), the shadow
# unit's ReadWritePaths= leaf, and deploy.sh's DUBPLATE_LIB_DIR
# assignment are three copies of one path with nothing but this leg
# holding them together: an edit to one alone gives a unit that
# starts and fails at first write with EROFS under
# ProtectSystem=strict (see the derivation comment above
# ReadWritePaths= in systemd/dubplate-shadow.service).
template_state_dir="$(sed -n 's/^DUBPLATE_STATE_DIR=//p' env/dubplate-server.env.template)"
unit_read_write_paths="$(sed -n 's/^ReadWritePaths=//p' systemd/dubplate-shadow.service)"
deploy_lib_dir="$(sed -n 's/^DUBPLATE_LIB_DIR="\(.*\)"$/\1/p' scripts/deploy.sh)"
echo "check.sh: DUBPLATE_STATE_DIR=$template_state_dir ReadWritePaths=$unit_read_write_paths DUBPLATE_LIB_DIR=$deploy_lib_dir"
if [ "$unit_read_write_paths" != "${template_state_dir}/shadow" ] || [ "$deploy_lib_dir" != "$template_state_dir" ]; then
  echo "check.sh: state-dir literals disagree (want ReadWritePaths=${template_state_dir}/shadow and DUBPLATE_LIB_DIR=${template_state_dir})" >&2
  status=1
fi

if [ "$status" -eq 0 ]; then
  echo "check.sh: all checks passed."
else
  echo "check.sh: one or more checks failed." >&2
fi
exit "$status"
