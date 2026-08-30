# shellcheck shell=bash
# Pure helper functions shared by music-import and music-backup. Meant to be
# sourced, not executed; every function returns a status rather than exiting
# the calling script, so a caller under `set -e` controls its own aborts.

# quiescent <dir> <minutes>
#
# Exits 0 only when no file anywhere under dir (recursive) was modified
# within the last <minutes>. Checking the directory's own mtime is not
# enough: appending to an existing file two levels down updates only that
# file's mtime, never a parent directory's, because no directory entry is
# added or removed. -print -quit stops at the first match instead of
# walking the rest of a large inbox once non-quiescence is already known.
quiescent() {
  local dir="$1" minutes="$2"
  [ -d "$dir" ] || return 1
  [ -z "$(find "$dir" -type f -mmin "-$minutes" -print -quit)" ]
}

# _no_traversal_member <path>
#
# True if path has no ".." component and is not absolute. Checked
# component-by-component rather than by substring match on "../" so a bare
# ".." member, a trailing "/..", and "../" all reject the same way.
_no_traversal_member() {
  local path="$1" parts part
  case "$path" in
  /*) return 1 ;;
  esac
  IFS=/ read -ra parts <<<"$path"
  for part in "${parts[@]}"; do
    [ "$part" = ".." ] && return 1
  done
  return 0
}

# safe_extract <archive> <destdir> <max_bytes>
#
# Extracts an untrusted archive with bsdtar after two checks: the archive
# file itself must not exceed max_bytes, and no member path may escape
# destdir via ".." or an absolute path. Both checks run before anything
# touches disk, so a rejected archive leaves destdir untouched.
safe_extract() {
  local archive="$1" destdir="$2" max_bytes="$3"
  local size member

  size="$(stat -c%s "$archive")" || return 1
  if [ "$size" -gt "$max_bytes" ]; then
    echo "safe_extract: $archive is $size bytes, over the $max_bytes byte cap" >&2
    return 1
  fi

  while IFS= read -r member; do
    if ! _no_traversal_member "$member"; then
      echo "safe_extract: $archive has a path-traversal member: $member" >&2
      return 1
    fi
  done < <(bsdtar -tf "$archive") || return 1

  mkdir -p "$destdir"
  bsdtar -xf "$archive" -C "$destdir"
}

# free_space_ok <path> <min_gb>
#
# True when the filesystem holding path has at least min_gb available.
free_space_ok() {
  local path="$1" min_gb="$2" avail_kb avail_gb
  avail_kb="$(df -Pk "$path" | awk 'NR==2 {print $4}')"
  avail_gb=$((avail_kb / 1024 / 1024))
  [ "$avail_gb" -ge "$min_gb" ]
}

# notify <email> <subject> <body>
#
# Sends a plain-text message via msmtp. msmtp itself resolves the account
# and credentials from /etc/msmtprc; this function only shapes the message.
notify() {
  local email="$1" subject="$2" body="$3"
  {
    printf 'To: %s\n' "$email"
    printf 'Subject: %s\n' "$subject"
    printf '\n%s\n' "$body"
  } | msmtp "$email"
}

# ping_hc <slug> [fail]
#
# Curls the healthchecks.io URL named MUSICBOX_HC_<SLUG>_URL, slug
# uppercased with "-" turned into "_" to make a legal env var name. Appends
# /fail when the second argument is the literal word "fail". A missing env
# var is a no-op, not an error: monitoring must never take down the pipeline
# it monitors.
ping_hc() {
  local slug="$1" mode="${2:-}" var_name url
  var_name="MUSICBOX_HC_$(printf '%s' "$slug" | tr '[:lower:]-' '[:upper:]_')_URL"
  url="${!var_name:-}"
  if [ -z "$url" ]; then
    echo "ping_hc: $var_name is not set; skipping ping for $slug" >&2
    return 0
  fi
  if [ "$mode" = "fail" ]; then
    url="${url%/}/fail"
  fi
  curl -fsS -o /dev/null "$url"
}
