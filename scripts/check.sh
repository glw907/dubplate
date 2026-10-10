#!/usr/bin/env bash
set -euo pipefail
# check.sh: the repo gate. `make check` covers the Go module (vet,
# golangci-lint, test, tidy-check; it does not build the binary, that
# is `make build`'s own job) and, first, the `web` target: Node and npm
# required, a stale or drifted `npm ci --prefix web` refused, the
# SvelteKit static bundle built, then svelte-check and the static web
# gates (no server-only file, no hardcoded color, no external origin,
# theme variable completeness, theme contrast, and the built bundle's
# stamped theme values), after building the bundle twice and refusing
# two builds that differ. `make check` then runs the rendered leg
# (`make web-rendered`): the real serve shell on loopback, driven by
# Playwright's own Chromium, so every full gate launches a browser.
# Shellcheck still covers the scripts that remain (deploy.sh, check.sh
# itself, vale-comments.sh, provision.sh), ruff covers the oracle shim's
# Python, and gitleaks scans the whole tree in the full form (the corpus
# under corpus/cases/ is the repo's first large data directory, and a
# check either gates or does not exist). Exits nonzero on any failure so
# it can be wired into CI or a pre-push hook unmodified.
#
# Four forms, one script, so no caller has to duplicate the gate's
# internals in a hand-composed command line:
#
#   check.sh                  the full gate over the whole tree. The form
#                             every boundary, every merge gate, and every
#                             pass end runs. Nothing is scoped away.
#   check.sh --scope-changed  the full leg list, with three
#                             blast-radius-insensitive legs narrowed to
#                             what the tree actually changed: gitleaks
#                             scans the derived change set, the
#                             internal/harness/conform suite runs only
#                             when a file under internal/harness/,
#                             internal/engine/, or corpus/ changed, and
#                             the rendered leg runs only when a file
#                             under web/, internal/serve/, internal/auth/,
#                             or internal/identity/ changed, or one of
#                             the serve shell's own files under
#                             cmd/dubplate-server/ (web.go, serve.go,
#                             auth.go, their tests, rendered_test.go).
#                             Every other leg runs exactly as in the
#                             full form.
#   check.sh --reduced        the comment-only and docs-only form: vale
#                             over Go comments, shellcheck, gitleaks over
#                             the change set, and gofmt, go vet and go
#                             test over the changed packages only; the
#                             rendered leg does not run. Never
#                             the final gate on master and never a
#                             boundary gate; a change that reaches master
#                             has had the full form run over it.
#   check.sh --merge          the merge ritual's in-tree gate: the full
#                             gate plus one gitleaks pass over the
#                             branch's own commit range, master..HEAD,
#                             with git history rather than --no-git. A
#                             secret committed by one task and rewritten
#                             by a later one is invisible to every
#                             worktree scan while it stays in the branch
#                             permanently, and this is the only leg that
#                             can see it. It runs before the work leaves
#                             the branch.
#
# The scope is always derived here, from git. A caller never names
# files, because its idea of which files a task touched is a prediction
# made before the work was written and the working tree is the fact; it
# may name only the base commit (--base), validated below, which is a
# fact the runner recorded. A path the class table cannot place, a
# gate-machinery path, an invalid base, and an empty change set each
# promote the run to the full form. Each form prints one line naming
# the scope in force and the files or packages it covers, so a
# transcript shows what was skipped; --dry-run prints it and stops.
#
# Two legs are staged gates, not optional ones. Each runs on every full
# invocation, prints its finding count and its first twenty sites, and
# is flipped from warning to gating by its own environment variable.
# Neither is a `command -v || skip`: the leg always runs and always
# reports; what the variable controls is whether today's known backlog
# blocks a commit while it is being worked off.
#
#   CHECK_CITATIONS=warn|fail       (default warn) process citations in
#       Go comments under cmd/ and internal/. The backlog is the tree's
#       existing citation lines; the default flips to fail at the
#       comment-cleanup task that clears them.
#   CHECK_EXPORTED_UNUSED=warn|fail (default warn) exported identifiers
#       under internal/ whose only references are their own file and
#       _test.go files. The default flips to fail once the dead surfaces
#       it names are either consumed or deleted.

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

usage() {
  cat >&2 <<'EOF'
usage: check.sh [--scope-changed | --reduced | --merge] [--base <sha>] [--dry-run]

  (no flag)       full gate over the whole tree
  --scope-changed full leg list, gitleaks and the conform suite scoped to
                  the change set git derives
  --reduced       comment-only and docs-only gate over the change set
  --merge         full gate plus a gitleaks scan of master..HEAD
  --base <sha>    with --scope-changed or --reduced: the full commit id the
                  change set is measured from; any other value runs the full
                  form
  --dry-run       derive the scope, print the scope line, run no leg
EOF
}

mode=full
base_set=0
base_arg=""
dry_run=0
while [ "$#" -gt 0 ]; do
  case "$1" in
  --scope-changed) mode=scope-changed ;;
  --reduced) mode=reduced ;;
  --merge) mode=merge ;;
  --base)
    if [ "$#" -lt 2 ]; then
      echo "check.sh: --base needs a commit id." >&2
      usage
      exit 2
    fi
    base_set=1
    base_arg="$2"
    shift
    ;;
  --dry-run) dry_run=1 ;;
  -h | --help)
    usage
    exit 0
    ;;
  *)
    echo "check.sh: unknown argument: $1" >&2
    usage
    exit 2
    ;;
  esac
  shift
done

if [ "$base_set" -eq 1 ] && [ "$mode" != "scope-changed" ] && [ "$mode" != "reduced" ]; then
  echo "check.sh: --base applies to --scope-changed and --reduced only." >&2
  usage
  exit 2
fi

# A gate lock: only one check.sh runs per tree at a time. `make
# tidy-check` runs `go mod tidy` before `git diff --exit-code`, and
# that first half mutates go.mod and go.sum, so two concurrent gate
# runs in one tree race on those files. The lock lives under the tree's
# own git-dir, which is .git for the main tree and .git/worktrees/<name>
# for a worktree, so a tree serializes against itself and against no
# other tree. The non-blocking attempt comes first only so the waiting
# line prints once, under contention; the second flock then blocks
# until the holder finishes.
# A dry run runs no leg and mutates nothing, so it takes no lock.
if [ "$dry_run" -eq 0 ]; then
  lock_file="$(git rev-parse --git-dir)/check.sh.lock"
  exec 9>"$lock_file"
  if ! flock -n 9; then
    echo "check.sh: another check.sh run holds this tree's gate lock; waiting..." >&2
    flock 9
  fi
fi

work_dir="$(mktemp -d)"
trap 'rm -rf "$work_dir"' EXIT

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
if ! command -v systemd-analyze >/dev/null 2>&1; then
  echo "check.sh: systemd-analyze not found; install systemd (it verifies every unit under systemd/)." >&2
  exit 1
fi

# ---------------------------------------------------------------- scope

# The change set the scoped forms work from. Git derives every file;
# the caller may supply only the base commit, which is a fact the runner
# records before the first edit, unlike a file list, which is a
# prediction. A supplied base is used only when it is a full commit id
# in the repo's object format, rev-parse resolves it to that same
# string, and it is an ancestor of HEAD. Anything else (an abbreviation,
# a ref name, a missing object, a non-ancestor, shell metacharacters)
# promotes the run to the full form. A later commit passed by mistake
# would narrow the scope, but every boundary, merge, and close runs the
# full form. Without --base the base is the merge base with master on a
# branch, and HEAD~1 on master itself, so a run on master still has a
# change set rather than an empty one. `git diff --name-only <base>`
# compares the base against the working tree, which covers staged and
# unstaged edits alike; untracked-but-not-ignored files come from `git
# status --porcelain`, because a brand-new file is invisible to git diff
# until it is staged and is exactly the file a gate must not miss. -uall
# is load-bearing there: the default porcelain output collapses an
# untracked directory to one entry ending in a slash, and every file
# inside it would go unscanned. An empty change set also promotes the
# run to the full form, since it proves nothing was checked.
changed_files=()
change_base=""
change_source=""
fallback_reason=""
resolve_supplied_base() {
  local fmt len resolved re
  fmt="$(git rev-parse --show-object-format 2>/dev/null || true)"
  case "$fmt" in
  sha1) len=40 ;;
  sha256) len=64 ;;
  *)
    fallback_reason="unknown object format '$fmt'"
    return 1
    ;;
  esac
  re="^[0-9a-f]{${len}}\$"
  if ! [[ $base_arg =~ $re ]]; then
    fallback_reason="--base '$base_arg' is not a full $len-hex commit id"
    return 1
  fi
  resolved="$(git rev-parse --verify --quiet "${base_arg}^{commit}" 2>/dev/null || true)"
  if [ "$resolved" != "$base_arg" ]; then
    fallback_reason="--base $base_arg does not resolve to that commit"
    return 1
  fi
  if ! git merge-base --is-ancestor "$base_arg" HEAD 2>/dev/null; then
    fallback_reason="--base $base_arg is not an ancestor of HEAD"
    return 1
  fi
  change_base="$base_arg"
  change_source=caller
}

derive_change_set() {
  local branch
  if [ "$base_set" -eq 1 ]; then
    resolve_supplied_base || return 1
  else
    branch="$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo HEAD)"
    if [ "$branch" = "master" ]; then
      change_base="$(git rev-parse --verify --quiet HEAD~1 || true)"
      change_source=HEAD~1
    else
      change_base="$(git merge-base master HEAD 2>/dev/null || true)"
      change_source=merge-base
    fi
    if [ -z "$change_base" ]; then
      fallback_reason="no merge base with master and no HEAD~1"
      return 1
    fi
  fi
  if ! {
    git diff --name-only "$change_base"
    git status --porcelain -uall | sed -n 's/^?? //p'
  } | sed 's:/$::' | sort -u >"$work_dir/changed.txt"; then
    fallback_reason="git could not list the change set"
    return 1
  fi
  mapfile -t changed_files <"$work_dir/changed.txt"
  if [ "${#changed_files[@]}" -eq 0 ]; then
    fallback_reason="the change set is empty"
    return 1
  fi
}

# The class table. Every changed path gets one class, rows matched in
# order; a path the table cannot place is unmatched. Gate machinery and
# unmatched promote the run to the full form in both scoped forms, so an
# unclassifiable case runs more, never less. The Go package class is the
# nearest enclosing directory that `go list ./...` names, for any file
# (testdata, embeds and .py included); a path whose own directory is gone
# maps to its nearest existing ancestor package.
declare -A go_pkg_dirs=()
load_go_pkg_dirs() {
  local root dir
  root="$(go list -m -f '{{.Dir}}')" || return 1
  while IFS= read -r dir; do
    if [ "$dir" = "$root" ]; then
      go_pkg_dirs["."]=1
    else
      go_pkg_dirs["${dir#"$root"/}"]=1
    fi
  done < <(go list -f '{{.Dir}}' ./...) || return 1
  [ "${#go_pkg_dirs[@]}" -gt 0 ]
}

classify_path() {
  local p="$1" dir
  case "$p" in
  scripts/check.sh | scripts/check-state-dir.sh | scripts/vale-comments.sh | \
    scripts/checks/* | Makefile | .golangci.yml | .gitleaks.toml | \
    .gitleaksignore | .gitignore | .vale.ini | .vale/* | ruff.toml | \
    go.mod | go.sum | web/package.json | web/package-lock.json | \
    web/scripts/gates/* | web/scripts/rendered/* | tools/photos/*)
    echo machinery
    return
    ;;
  web/*)
    echo web
    return
    ;;
  esac
  dir="$(dirname "$p")"
  while :; do
    if [ -n "${go_pkg_dirs[$dir]+x}" ]; then
      echo gopkg
      return
    fi
    [ "$dir" = "." ] && break
    dir="$(dirname "$dir")"
  done
  case "$p" in
  corpus/*) echo corpus ;;
  photos/*) echo photos ;;
  systemd/* | cloud-init/* | env/* | scripts/deploy.sh | scripts/provision.sh | \
    scripts/firewall-rules.json)
    echo ops
    ;;
  tools/unidecode-table/*) echo pytool ;;
  docs/* | .claude/* | _salvage/* | tools/venv/* | README.md | ROADMAP.md | \
    IDEAS.md | BACKLOG.md | CLAUDE.md | LICENSE)
    echo noeffect
    ;;
  *) echo unmatched ;;
  esac
}

# Promotes the run to the full form when any changed path lands in a
# class that forces it: gate machinery and unmatched in both scoped
# forms, and in the reduced form also everything but a Go source file,
# Web, Photos and No effect.
classify_change_set() {
  local f class
  if ! load_go_pkg_dirs; then
    fallback_reason="go list could not name the module's packages"
    return 1
  fi
  for f in "${changed_files[@]}"; do
    class="$(classify_path "$f")"
    case "$class" in
    machinery | unmatched)
      fallback_reason="$f is $class"
      return 1
      ;;
    gopkg)
      if [ "$mode" = "reduced" ] && [ "${f%.go}" = "$f" ]; then
        fallback_reason="$f is a non-Go file in a Go package, outside the reduced form"
        return 1
      fi
      ;;
    corpus | ops | pytool)
      if [ "$mode" = "reduced" ]; then
        fallback_reason="$f is $class, outside the reduced form"
        return 1
      fi
      ;;
    esac
  done
}

# The Go packages the change set touches: the directory of every changed
# *.go path that still exists and still holds Go source. A deleted file's
# directory drops out, and so does a directory the go tool cannot see
# (_salvage/ and anything else under a leading underscore or dot).
changed_packages=()
derive_changed_packages() {
  local f dir
  : >"$work_dir/pkgs.txt"
  for f in "${changed_files[@]}"; do
    case "$f" in
    *.go) ;;
    *) continue ;;
    esac
    dir="$(dirname "$f")"
    case "$dir" in
    _* | .* | */_* | */.*) continue ;;
    esac
    [ -d "$dir" ] || continue
    compgen -G "$dir/*.go" >/dev/null || continue
    echo "./$dir" >>"$work_dir/pkgs.txt"
  done
  mapfile -t changed_packages < <(sort -u "$work_dir/pkgs.txt")
}

# The conform suite is blast-radius-scoped by hand because it is a
# package inside `go test ./...` rather than a leg of this script: it
# reads the corpus and the engine through the harness, so a change under
# any of those three roots can change its verdict and nothing else can.
conform_affected=1
derive_conform_affected() {
  local f
  conform_affected=0
  for f in "${changed_files[@]}"; do
    case "$f" in
    internal/harness/* | internal/engine/* | corpus/*)
      conform_affected=1
      return
      ;;
    esac
  done
}

# The rendered leg is scoped the same way: it serves the bundle through
# the serve shell and its auth routes, so only a change to one of those
# can change what the browser sees.
rendered_affected=1
derive_rendered_affected() {
  local f
  rendered_affected=0
  for f in "${changed_files[@]}"; do
    case "$f" in
    web/* | internal/serve/* | internal/auth/* | internal/identity/* | \
      cmd/dubplate-server/web.go | cmd/dubplate-server/web_test.go | \
      cmd/dubplate-server/serve.go | cmd/dubplate-server/serve_test.go | \
      cmd/dubplate-server/auth.go | cmd/dubplate-server/auth_test.go | \
      cmd/dubplate-server/rendered_test.go)
      rendered_affected=1
      return
      ;;
    esac
  done
}

scope_gitleaks=full
scope_conform=full
scope_rendered=full
if [ "$mode" = "reduced" ]; then
  scope_rendered=skipped
fi
if [ "$mode" = "scope-changed" ] || [ "$mode" = "reduced" ]; then
  if derive_change_set && classify_change_set; then
    derive_changed_packages
    derive_conform_affected
    derive_rendered_affected
    scope_gitleaks=changed
    if [ "$mode" = "scope-changed" ]; then
      if [ "$conform_affected" -eq 0 ]; then
        scope_conform=skipped
      fi
      if [ "$rendered_affected" -eq 0 ]; then
        scope_rendered=skipped
      fi
    fi
  else
    mode=full
    scope_rendered=full
  fi
fi

case "$mode" in
full)
  if [ -n "$fallback_reason" ]; then
    echo "check.sh: scope=full fallback=\"$fallback_reason\" base=${change_base:-none} source=${change_source:-none}; every leg over the whole tree."
  else
    echo "check.sh: scope=full; every leg over the whole tree."
  fi
  ;;
merge)
  echo "check.sh: scope=full+history; every leg over the whole tree, plus gitleaks over master..HEAD."
  ;;
scope-changed)
  echo "check.sh: scope=changed base=$change_base source=$change_source conform=$scope_conform rendered=$scope_rendered gitleaks=$scope_gitleaks files=${#changed_files[@]}: ${changed_files[*]}"
  ;;
reduced)
  echo "check.sh: scope=reduced base=$change_base source=$change_source rendered=skipped packages=${#changed_packages[@]}: ${changed_packages[*]} files=${#changed_files[@]}: ${changed_files[*]}"
  ;;
esac

if [ "$dry_run" -eq 1 ]; then
  exit 0
fi

# ------------------------------------------------------------------ legs

# gitleaks resolves its config relative to the scanned source path, so a
# scan aimed anywhere but the repo root silently falls back to the
# built-in defaults and drops .gitleaks.toml's useDefault line and its
# .claude/worktrees/ allowlist. -c with an absolute path is what keeps
# every form under the same rules.
#
# The scoped form scans one path per changed file rather than the change
# set's common directories: gitleaks takes a single -s per process, and
# reducing to directories turns any edit to a root-level file, the
# Makefile among them, into a whole-tree scan, which is the one outcome
# the scoping exists to avoid. Above the file cap the whole-tree scan is
# the cheaper of the two anyway, and the form says so rather than
# quietly scanning less.
gitleaks_file_cap=100
run_gitleaks_scoped() {
  local f rc=0 scanned=0
  for f in "${changed_files[@]}"; do
    [ -f "$f" ] || continue
    scanned=$((scanned + 1))
    if ! gitleaks detect --no-git -s "$f" -c "$REPO_ROOT/.gitleaks.toml" --redact >/dev/null 2>"$work_dir/gitleaks.err"; then
      cat "$work_dir/gitleaks.err" >&2
      rc=1
    fi
  done
  echo "check.sh: gitleaks scanned $scanned changed file(s) under $REPO_ROOT/.gitleaks.toml."
  return "$rc"
}

if [ "$mode" = "scope-changed" ]; then
  # `web` builds the bundle the Go tests read, so it runs ahead of
  # every narrowed Go target here too, not only inside `make check`'s
  # own prerequisite.
  echo "== make web =="
  if ! make web; then
    status=1
  fi
fi

if [ "$mode" != "reduced" ]; then
  echo "== make check (Go) =="
fi
if [ "$mode" = "scope-changed" ] && { [ "$scope_conform" = "skipped" ] || [ "$scope_rendered" = "skipped" ]; }; then
  # The scoped form keeps vet, lint and tidy-check whole and narrows the
  # test leg (the conform suite) and the rendered leg independently, each
  # by its own inputs.
  if ! make vet lint tidy-check; then
    status=1
  fi
  if [ "$scope_conform" = "skipped" ]; then
    echo "check.sh: the change set touches no harness, engine or corpus file; running go test without internal/harness/conform."
    mapfile -t scoped_pkgs < <(go list ./... | grep -v '/internal/harness/conform$')
    if [ "${#scoped_pkgs[@]}" -eq 0 ]; then
      echo "check.sh: go list returned no packages." >&2
      status=1
    elif ! go test "${scoped_pkgs[@]}"; then
      status=1
    fi
  elif ! make test; then
    status=1
  fi
  # `make check` carries the rendered leg in every other form; this
  # branch replaces it, so it runs the leg itself when the change set
  # can reach it.
  if [ "$scope_rendered" = "skipped" ]; then
    echo "check.sh: the change set touches no web/, internal/serve/, internal/auth/, internal/identity/, or serve shell file; skipping the rendered leg."
  else
    echo "check.sh: the change set touches the rendered leg's inputs; running make web-rendered."
    if ! make web-rendered; then
      status=1
    fi
  fi
elif [ "$mode" != "reduced" ]; then
  if ! make check; then
    status=1
  fi
fi

if [ "$mode" = "reduced" ]; then
  # The comment-only form skips svelte-check, the static web gates, and
  # the rendered leg (never a boundary gate), but still rebuilds the
  # bundle, since a stale build is the one state every form refuses to
  # read from.
  echo "check.sh: the reduced form skips the rendered leg."
  echo "== npm run build --prefix web =="
  if ! npm run build --prefix web; then
    status=1
  fi

  echo "== gofmt, go vet, go test (changed packages) =="
  if [ "${#changed_packages[@]}" -eq 0 ]; then
    echo "check.sh: the change set holds no Go package; the Go legs have nothing to run."
  else
    if ! unformatted="$(gofmt -l "${changed_packages[@]}")"; then
      status=1
    fi
    if [ -n "$unformatted" ]; then
      echo "check.sh: gofmt would rewrite:" >&2
      echo "$unformatted" >&2
      status=1
    fi
    if ! go vet "${changed_packages[@]}"; then
      status=1
    fi
    if ! go test "${changed_packages[@]}"; then
      status=1
    fi
  fi
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

if [ "$mode" != "reduced" ]; then
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
fi

echo "== gitleaks (secret scan) =="
# --no-git scans the tracked worktree as it sits, not the full git
# history (the --merge form's own leg covers the branch's history): the
# corpus under corpus/cases/ is the repo's first large data directory,
# and every response body in it is verbatim wire data from a live
# service, so this is the rail that makes CLAUDE.md's "eyeball the diff
# for stray personal values" instruction a backstop rather than the only
# net. The generic-api-key rule fires only when a key-like field name
# sits on the same line as a value whose entropy clears its threshold;
# exactly one of the tree's DUBPLATE_* names used to pair that way in a
# test fixture, and it is renamed rather than suppressed, so no
# .gitleaksignore entry remains.
# The repo's .gitleaks.toml keeps the built-in rules and allowlists
# one path, .claude/worktrees/; its own comments carry the reasoning.
if [ "$scope_gitleaks" = "changed" ] && [ "${#changed_files[@]}" -le "$gitleaks_file_cap" ]; then
  if ! run_gitleaks_scoped; then
    status=1
  fi
else
  if [ "$scope_gitleaks" = "changed" ]; then
    echo "check.sh: ${#changed_files[@]} changed files is over the $gitleaks_file_cap-file cap; scanning the whole tree instead."
  fi
  if ! gitleaks detect --no-git -s . -c "$REPO_ROOT/.gitleaks.toml" --redact; then
    status=1
  fi
fi

if [ "$mode" = "merge" ]; then
  echo "== gitleaks (branch history, master..HEAD) =="
  # Without --no-git gitleaks reads commits rather than the worktree, so
  # this is the leg that sees a secret a later commit on the same branch
  # removed. The worktree is clean of it by then and every --no-git scan,
  # boundary scans included, reports no leaks while the blob stays in the
  # branch forever.
  if ! git rev-parse --verify --quiet master >/dev/null; then
    echo "check.sh: no master ref in this tree, so the branch-history scan cannot run." >&2
    status=1
  elif [ -z "$(git rev-list master..HEAD)" ]; then
    echo "check.sh: master..HEAD is empty, so the branch history holds no commits to scan."
  elif ! gitleaks detect -s . -c "$REPO_ROOT/.gitleaks.toml" --redact --log-opts "master..HEAD"; then
    status=1
  fi
fi

echo "== photo pool integrity =="
# tools/photos/check.mjs integrity verifies the committed pool under
# photos/ (records, files, vocabulary, dedupe keys). Node is already
# required by the web leg, so a missing node fails here rather than
# skipping. Silent on success.
if ! command -v node >/dev/null 2>&1; then
  echo "check.sh: node not found; install it (the web leg needs it too)." >&2
  status=1
elif ! node tools/photos/check.mjs integrity; then
  status=1
fi

echo "== vale (Go comment prose) =="
if ! scripts/vale-comments.sh; then
  status=1
fi

if [ "$mode" != "reduced" ]; then
  echo "== process citations in Go comments =="
  # A comment that cites the process which produced the code (a rung, a
  # task, a ruling id, a plan path, a salvage line number) is addressed
  # to the pass that wrote it and is noise to every later reader. The em
  # dash ban is the proof that a mechanical rule holds here where an
  # instruction does not.
  citation_patterns='ruled, conductor|Resolution R|rung [0-9]|Task [0-9]|\(C[0-9]|plans/.*\.md:[0-9]|_salvage/.*:[0-9]|Ruled, '
  {
    git ls-files -- cmd internal
    git ls-files --others --exclude-standard -- cmd internal
  } | grep -E '\.go$' | grep -v '_test\.go$' | sort -u >"$work_dir/citation-src.txt"
  # Only the text from each line's first // marker onward is searched, so
  # a pattern sitting in a string literal or an identifier is not a hit.
  : >"$work_dir/citations.txt"
  if [ -s "$work_dir/citation-src.txt" ]; then
    xargs -r -a "$work_dir/citation-src.txt" grep -Hn '//' |
      awk 'match($0, /^[^:]+:[0-9]+:/) {
             head = substr($0, RSTART, RLENGTH)
             rest = substr($0, RLENGTH + 1)
             i = index(rest, "//")
             if (i) print head substr(rest, i)
           }' |
      grep -E "$citation_patterns" >"$work_dir/citations.txt" || true
  fi
  citation_count="$(wc -l <"$work_dir/citations.txt")"
  if [ "$citation_count" -eq 0 ]; then
    echo "check.sh: no process citations in Go comments."
  else
    echo "check.sh: $citation_count process citation(s) in Go comments; first twenty:" >&2
    head -n 20 "$work_dir/citations.txt" >&2
    if [ "${CHECK_CITATIONS:-warn}" = "fail" ]; then
      status=1
    else
      echo "check.sh: CHECK_CITATIONS=warn, so this is a warning; set CHECK_CITATIONS=fail to gate on it." >&2
    fi
  fi

  echo "== exported identifiers with no consumer (internal/) =="
  # A forward hook exported ahead of its consumer reads as justified in
  # the diff that introduced it and as dead weight in the package. The
  # rule: an exported identifier under internal/ whose only references
  # are its own file and _test.go files has no consumer.
  # golangci-lint cannot express it, because
  # unused's exported-is-used setting is gone in the pinned 2.13.2 and
  # the linter treats every exported identifier as used, so the leg lives
  # here. It reads names rather than types, so a name declared in two
  # packages reads as consumed in both; the error runs toward reporting
  # too little, never toward a false accusation.
  {
    git ls-files -- cmd internal
    git ls-files --others --exclude-standard -- cmd internal
  } | grep -E '\.go$' | grep -v '_test\.go$' | sort -u >"$work_dir/go-src.txt"
  grep '^internal/' "$work_dir/go-src.txt" >"$work_dir/internal-src.txt" || true
  : >"$work_dir/exported.txt"
  if [ -s "$work_dir/internal-src.txt" ] && [ -s "$work_dir/go-src.txt" ]; then
    # The awk program is single-quoted on purpose: $1 and $2 are awk's
    # own fields, not shell variables.
    # shellcheck disable=SC2016
    xargs -r -a "$work_dir/internal-src.txt" awk '
      /^func \(/ {
        if (match($0, /^func \([^)]*\) [A-Z][A-Za-z0-9_]*/)) {
          n = substr($0, RSTART, RLENGTH)
          sub(/^func \([^)]*\) /, "", n)
          print FILENAME "\t" n
        }
        next
      }
      /^func [A-Z]/ { n = $2; sub(/\(.*/, "", n); print FILENAME "\t" n; next }
      /^type [A-Z]/ { print FILENAME "\t" $2; next }
      /^(var|const) [A-Z]/ { n = $2; sub(/[,=].*/, "", n); print FILENAME "\t" n; next }
      inblock && /^\t[A-Z][A-Za-z0-9_]*[ ,=\t]/ { n = $1; sub(/[,=].*/, "", n); print FILENAME "\t" n; next }
      /^(var|const|type) \(/ { inblock = 1; next }
      /^\)/ { inblock = 0 }
    ' | sort -u >"$work_dir/decls.txt"
    cut -f2 "$work_dir/decls.txt" | sort -u >"$work_dir/names.txt"
    : >"$work_dir/refs.txt"
    if [ -s "$work_dir/names.txt" ]; then
      xargs -r -a "$work_dir/go-src.txt" grep -o -w -H -F -f "$work_dir/names.txt" |
        awk -F: '{ print $1 "\t" $NF }' | sort -u >"$work_dir/refs.txt" || true
    fi
    # shellcheck disable=SC2016
    awk -F'\t' '
      NR == FNR { seen[$2 SUBSEP $1] = 1; files[$2] = files[$2] "\n" $1; next }
      {
        alive = 0
        n = split(files[$2], a, "\n")
        for (i = 1; i <= n; i++) {
          if (a[i] != "" && a[i] != $1) alive = 1
        }
        if (!alive) print $1 "\t" $2
      }
    ' "$work_dir/refs.txt" "$work_dir/decls.txt" | sort >"$work_dir/exported.txt"
  fi
  exported_count="$(wc -l <"$work_dir/exported.txt")"
  if [ "$exported_count" -eq 0 ]; then
    echo "check.sh: every exported identifier under internal/ has a non-test consumer."
  else
    echo "check.sh: $exported_count exported identifier(s) under internal/ with no non-test consumer; first twenty:" >&2
    head -n 20 "$work_dir/exported.txt" >&2
    if [ "${CHECK_EXPORTED_UNUSED:-warn}" = "fail" ]; then
      status=1
    else
      echo "check.sh: CHECK_EXPORTED_UNUSED=warn, so this is a warning; set CHECK_EXPORTED_UNUSED=fail to gate on it." >&2
    fi
  fi
fi

if [ "$mode" != "reduced" ]; then
  echo "== systemd unit verification =="
  # systemd-analyze verify insists every ExecStart= binary already exists on
  # THIS machine, which is never true here: the paths are box-only
  # (/usr/local/bin/dubplate-server, installed by deploy.sh, never present
  # on the workstation). That specific complaint is expected and filtered out
  # by pattern; any other verify output (bad section, missing directive, bad
  # OnCalendar syntax, and so on) still fails the gate. This is the
  # pattern-filtered `systemd-analyze verify` approach, not a grep-based
  # fallback: `--recursive-errors=no` alone does not suppress the
  # missing-executable complaint, so the filter is the actual mechanism doing
  # the work.
  # Every unit kind the directory can hold, with nullglob so a kind the
  # tree does not ship drops out instead of reaching systemd-analyze as a
  # literal pattern. An empty list is a failure, never a pass.
  shopt -s nullglob
  unit_files=(systemd/*.service systemd/*.timer systemd/*.path)
  shopt -u nullglob
  if [ "${#unit_files[@]}" -eq 0 ]; then
    echo "check.sh: no unit files under systemd/ to verify." >&2
    status=1
  fi
  verify_output="$(systemd-analyze verify --recursive-errors=no "${unit_files[@]}" 2>&1 || true)"
  unexpected="$(grep -v -E ': Command /usr/local/bin/[A-Za-z0-9_.-]+ is not executable: No such file or directory$' <<<"$verify_output" || true)"
  if [ -n "$unexpected" ]; then
    echo "$unexpected" >&2
    status=1
  else
    echo "check.sh: systemd-analyze verify clean, modulo the expected box-only ExecStart paths."
  fi

  echo "== state-dir agreement (env template, every dubplate-server unit, deploy.sh) =="
  if ! bash scripts/check-state-dir.sh; then
    status=1
  fi
fi

if [ "$status" -eq 0 ]; then
  echo "check.sh: all checks passed."
else
  echo "check.sh: one or more checks failed." >&2
fi
exit "$status"
