# shellcheck shell=bash
# Pure helper functions shared by music-import, music-beet, and
# music-backup. Meant to be sourced, not executed; every function returns a
# status rather than exiting the calling script, so a caller under `set -e`
# controls its own aborts.

# musicbox_lock_file
#
# Prints the path to the single flock music-import and music-beet both
# take before touching the beets database, so a scheduled import run and
# an interactive `beet` session (via music-beet) never race each other.
# The default is a persistent path, /var/lib/musicbox, not the
# RuntimeDirectory=musicbox systemd tmpfs (/run/musicbox): with flock, the
# lock IS the flock held on the open file descriptor, not the file's
# content, so a lock file that is never deleted is harmless -- but a
# RuntimeDirectory is torn down at unit stop and recreated fresh at unit
# start, which broke mutual exclusion two ways: a long interactive
# music-beet session outlives the directory (dies on the missing
# directory on its next run, since the fd is opened under `set -e`), and a
# beet session already holding a lock on the old (now-unlinked) inode
# never contends with the next scheduled import, which opens a brand new
# file at the same path. Overridable via MUSICBOX_IMPORT_LOCK_FILE -- both
# callers must resolve through this one function rather than carrying
# their own default, or a fix applied to one and not the other silently
# breaks their mutual exclusion.
musicbox_lock_file() {
  printf '%s' "${MUSICBOX_IMPORT_LOCK_FILE:-/var/lib/musicbox/musicbox-import.lock}"
}

# musicbox_ensure_lock_dir <lock_file>
#
# mkdir -p's a lock file's parent directory. Called by every lock-opening
# script right before it opens the lock fd (`exec N>"$lock_file"` under
# `set -e` dies outright on a missing directory), so this one function is
# what keeps music-import and music-beet from each carrying their own
# mkdir -- and therefore each needing its own fix the day one of them
# forgets it. Unlike the RuntimeDirectory=musicbox systemd creates for
# /run/musicbox, nothing recreates /var/lib/musicbox's tree on its own; a
# fresh box or a wiped state volume needs this at least once, and a
# no-op mkdir -p on every other run costs nothing.
musicbox_ensure_lock_dir() {
  local lock_file="$1"
  mkdir -p "$(dirname "$lock_file")"
}

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
# Best-effort by contract, guarded here rather than at each call site: a
# stalled mail queue or a msmtprc misconfiguration must never take an
# import run down, so a delivery failure is logged to stderr and this
# function still returns 0. That guarantee has to live in the function
# itself -- notify is routinely called bare under `set -e` (music-import
# never wraps a call in `|| true`), so any call site that forgot the guard
# would otherwise abort the run on the next mail-delivery hiccup.
notify() {
  local email="$1" subject="$2" body="$3"
  {
    printf 'To: %s\n' "$email"
    printf 'Subject: %s\n' "$subject"
    printf '\n%s\n' "$body"
  } | msmtp "$email" ||
    echo "notify: msmtp delivery to $email failed" >&2
  return 0
}

# ping_hc <slug> [fail]
#
# Curls the pings Worker URL named MUSICBOX_HC_<SLUG>_URL, slug uppercased
# with "-" turned into "_" to make a legal env var name. Appends /fail when
# the second argument is the literal word "fail", and that failure path is
# sent as an explicit POST: the Worker (~/Projects/pings/src/index.ts)
# routes /ping/<slug>/fail on POST only and 404s a bare GET, so a failure
# ping sent the way the success path is (curl's default GET) is silently
# dropped rather than recorded. The success ping stays on curl's default
# method, since the Worker accepts GET or POST for /ping/<slug>. A missing
# or empty env var is a no-op (with a notice to stderr), not an error:
# monitoring must never take down the pipeline it monitors. --retry 3
# --max-time 10 keeps a slow or flaky Worker from ever being the reason an
# otherwise-successful run hangs or gets misreported as failed. That same
# never-fails contract covers curl itself failing outright (DNS, timeout,
# a non-2xx after retries exhaust): guarded here, at the function
# definition, rather than trusting every call site to remember `|| true`
# -- ping_hc is called bare in more than one place in music-import, several
# of them on a success path that must still reach `exit 0`.
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
    curl -fsS --retry 3 --max-time 10 -X POST -o /dev/null "$url" ||
      echo "ping_hc: failure ping to $url did not succeed" >&2
  else
    curl -fsS --retry 3 --max-time 10 -o /dev/null "$url" ||
      echo "ping_hc: ping to $url did not succeed" >&2
  fi
  return 0
}

# quality_gate <dir>
#
# Enforces the only-FLAC ingest rule (Geoff, 2026-08-30: only high-quality
# FLAC enters the library) over every audio file under dir, recursively.
# An audio file is one with a case-insensitive flac/mp3/m4a/ogg/opus/wav/
# aiff extension; anything else (art, .cue, .nfo) is out of scope and never
# inspected. A non-FLAC extension fails outright; a FLAC file must both
# pass `flac -t`'s full decode test (catches a corrupt or truncated
# upload) and carry at least 44.1kHz/16-bit.
#
# Every symlink under dir fails outright, regardless of extension, checked
# before any extension logic runs. `find -type f` never matches a symlink
# (even one that resolves to a plain regular file), so scanning with
# -type f alone would let a symlink glide straight through this whole
# function unexamined -- and a symlink can point anywhere the beets
# process can read, or (from a maliciously crafted archive member) at a
# path that does not even exist yet. This function's `find` therefore also
# matches -type l, and the loop checks `-L` first.
#
# Prints one "path<TAB>reason" line per failing file to stdout -- the raw,
# technical reason; callers that notify a human are expected to translate
# it, not forward it verbatim to an uploader. Exits 0 only when every
# audio file under dir passed.
quality_gate() {
  local dir="$1" file ext rate bps failed=0
  while IFS= read -r -d '' file; do
    if [ -L "$file" ]; then
      printf '%s\tsymlink not allowed\n' "$file"
      failed=1
      continue
    fi
    ext="${file##*.}"
    ext="${ext,,}"
    case "$ext" in
    flac) ;;
    mp3 | m4a | ogg | opus | wav | aiff)
      printf '%s\tnot FLAC (.%s)\n' "$file" "$ext"
      failed=1
      continue
      ;;
    *) continue ;;
    esac

    if ! flac -t "$file" >/dev/null 2>&1; then
      printf '%s\tflac -t decode test failed (corrupt or truncated file)\n' "$file"
      failed=1
      continue
    fi

    # metaflac prints the two values on their own lines, and read clears its
    # variable at EOF, so a metaflac that printed nothing leaves both empty
    # for the guard below to report. The trailing "|| true" is what keeps
    # that case from aborting a caller under set -e, since the second read
    # returns nonzero at EOF: this file promises to return a status rather
    # than take the calling script down with it.
    { read -r rate; read -r bps; } < <(metaflac --show-sample-rate --show-bps "$file" 2>/dev/null) || true
    if [ -z "$rate" ] || [ -z "$bps" ] || [ "$rate" -lt 44100 ] || [ "$bps" -lt 16 ]; then
      printf '%s\tbelow CD quality (%s Hz / %s-bit; needs >=44100 Hz and >=16-bit)\n' \
        "$file" "${rate:-unknown}" "${bps:-unknown}"
      failed=1
    fi
  done < <(find "$dir" \( -type f -o -type l \) -print0)
  return "$failed"
}

# unique_dest_path <dir> <name>
#
# Prints "<dir>/<name>" if nothing is there yet; otherwise appends a
# HHMMSS time suffix, and if even that collides (two collisions inside the
# same second), a further numeric ".2", ".3", ... suffix. Never overwrites:
# a same-named album already in review from an earlier run today must
# never be silently merged with, or clobbered by, a fresh one.
unique_dest_path() {
  local dir="$1" name="$2" stem n
  if [ ! -e "$dir/$name" ]; then
    printf '%s' "$dir/$name"
    return 0
  fi
  stem="$dir/$name-$(date +%H%M%S)"
  if [ ! -e "$stem" ]; then
    printf '%s' "$stem"
    return 0
  fi
  n=2
  while [ -e "$stem.$n" ]; do
    n=$((n + 1))
  done
  printf '%s' "$stem.$n"
}

# safe_mv_to_review <src> <dest_dir>
#
# Moves src into dest_dir under a name unique_dest_path guarantees is
# free, so a same-day repeat quarantine of an identically named item lands
# beside the first instead of overwriting or merging into it. Creating
# dest_dir and the move itself can each fail for another reason
# (permissions, cross-device, disk full); either one warns and returns
# nonzero rather than aborting a caller under `set -e`, because one
# contributor's un-movable item must never take every other contributor's
# run down with it. Prints the destination path on success.
safe_mv_to_review() {
  local src="$1" dest_dir="$2" name dest
  name="$(basename "$src")"
  if ! mkdir -p "$dest_dir"; then
    echo "safe_mv_to_review: could not create $dest_dir; leaving $src in place" >&2
    return 1
  fi
  dest="$(unique_dest_path "$dest_dir" "$name")"
  if ! mv "$src" "$dest"; then
    echo "safe_mv_to_review: could not move $src to $dest; left in place" >&2
    return 1
  fi
  printf '%s' "$dest"
}
