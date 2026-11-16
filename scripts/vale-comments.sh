#!/usr/bin/env bash
# vale-comments: commit-gate Vale scan of Go comment prose. Vale extracts the
# comment text from each .go file and lints it through the vendored Survey
# overlay, so the em dash and the banned lexicon block the gate.
#
# A missing Vale hard-fails: an advisory guard that skips quietly is the
# silent-decay class this gate exists to close, and "installed on the
# workstation" is not portability.
#
# Pass --check-vendor to additionally assert the vendored Survey overlay
# matches the workstation's canonical copy. Off by default: a fresh clone
# or CI has no workstation dotfiles to compare against, and the committed
# overlay is the source of truth there.
#
# Exit 1 on any error-level finding, a missing Vale, or (with
# --check-vendor) a drifted overlay; 0 otherwise.
#
# The file list unions tracked and untracked-but-not-ignored .go files: a
# brand-new file is invisible to `git ls-files` alone until it is staged, so
# a gate run before `git add` silently skips it and reports a false clean.
#
# Pass --only followed by Go file paths to read just those, as the scoped
# gate does. The paths pass the same _salvage/ filter, and a path that no
# longer exists is dropped. An empty list after --only reads nothing and
# exits 0; with no --only every Go file is read.
set -uo pipefail
cd "$(dirname "$0")/.." || exit 1

check_vendor=0
only=0
paths=()
for arg in "$@"; do
  case "$arg" in
    --check-vendor) check_vendor=1 ;;
    --only) only=1 ;;
    *.go)
      if [ "$only" -eq 0 ]; then
        echo "vale-comments: a path needs --only before it: $arg" >&2
        exit 2
      fi
      paths+=("$arg")
      ;;
    *)
      echo "vale-comments: unknown argument: $arg" >&2
      exit 2
      ;;
  esac
done

if ! command -v vale >/dev/null 2>&1; then
  echo "vale-comments: vale not installed; install it, do not skip the gate" >&2
  exit 1
fi

if [ "$check_vendor" -eq 1 ]; then
  vendor="$HOME/.dotfiles/scripts/glw907-vendor.sh"
  if [ ! -x "$vendor" ]; then
    echo "vale-comments: --check-vendor given but $vendor is not present or not executable" >&2
    exit 1
  fi
  "$vendor" "$PWD" || exit 1
fi

# _salvage/ is read-only reference material: it is frozen at the
# pre-greenfield tag and excluded from every gate leg that walks files,
# this one included, or all 76 salvage Go files re-enter the scan (git's
# '*.go' pathspec matches at any depth).
if [ "$only" -eq 1 ]; then
  files=()
  for f in "${paths[@]}"; do
    case "$f" in _salvage/*) continue ;; esac
    [ -f "$f" ] && files+=("$f")
  done
else
  mapfile -t files < <(
    { git ls-files '*.go'; git ls-files --others --exclude-standard '*.go'; } |
      grep -v '^_salvage/' | sort -u
  )
fi
echo "vale-comments: ${#files[@]} files"
[ "${#files[@]}" -gt 0 ] || exit 0

if vale --minAlertLevel=error --output=line "${files[@]}"; then
  echo "vale-comments: clean"
else
  echo "vale-comments: error-level findings in Go comment prose (above)." >&2
  echo "Fix the comment, or demote a false-positive rule in .vale.ini." >&2
  exit 1
fi
