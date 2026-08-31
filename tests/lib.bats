#!/usr/bin/env bats
# Unit tests for scripts/lib.sh, the pure helper functions music-import and
# music-backup depend on. Anything that would reach the network or a real
# mail server is stubbed on PATH (msmtp, curl); the audio tools quality_gate
# drives (ffmpeg, flac, metaflac) run for real, since a stub could not tell
# a genuinely valid FLAC from a corrupt one.

setup() {
  LIB_ROOT="$(cd "$(dirname "$BATS_TEST_FILENAME")/.." && pwd)"
  # shellcheck source=../scripts/lib.sh
  source "$LIB_ROOT/scripts/lib.sh"
  WORK_DIR="$(mktemp -d)"
  STUB_BIN="$WORK_DIR/stub-bin"
  mkdir -p "$STUB_BIN"
  PATH="$STUB_BIN:$PATH"
}

teardown() {
  rm -rf "$WORK_DIR"
}

# --- quiescent ---------------------------------------------------------

@test "quiescent exits 0 when nothing under dir is newer than the cutoff" {
  mkdir -p "$WORK_DIR/inbox/album"
  touch -d "30 minutes ago" "$WORK_DIR/inbox/album/one.flac"
  touch -d "30 minutes ago" "$WORK_DIR/inbox/album" "$WORK_DIR/inbox"
  run quiescent "$WORK_DIR/inbox" 10
  [ "$status" -eq 0 ]
}

@test "quiescent exits nonzero when a top-level file is newer than the cutoff" {
  mkdir -p "$WORK_DIR/inbox"
  touch "$WORK_DIR/inbox/fresh.flac"
  run quiescent "$WORK_DIR/inbox" 10
  [ "$status" -ne 0 ]
}

@test "quiescent catches a nested write a directory's own mtime would miss" {
  mkdir -p "$WORK_DIR/inbox/album/disc1"
  : > "$WORK_DIR/inbox/album/disc1/track.flac"
  touch -d "30 minutes ago" \
    "$WORK_DIR/inbox" "$WORK_DIR/inbox/album" "$WORK_DIR/inbox/album/disc1" \
    "$WORK_DIR/inbox/album/disc1/track.flac"
  # An in-progress write appends to an existing file: no directory entry is
  # added or removed, so every directory's mtime stays exactly as set above.
  # Only a recursive file-level check catches this; a directory-mtime check
  # would report the inbox quiescent while the file is still being written.
  printf 'more bytes' >>"$WORK_DIR/inbox/album/disc1/track.flac"
  dir_mtime_before="$(stat -c %Y "$WORK_DIR/inbox")"
  run quiescent "$WORK_DIR/inbox" 10
  [ "$status" -ne 0 ]
  [ "$(stat -c %Y "$WORK_DIR/inbox")" -eq "$dir_mtime_before" ]
}

# --- musicbox_lock_file ---------------------------------------------------

@test "musicbox_lock_file returns the shared default and honors an override" {
  unset MUSICBOX_IMPORT_LOCK_FILE
  run musicbox_lock_file
  [ "$status" -eq 0 ]
  # Persistent, not the RuntimeDirectory tmpfs: with flock, the lock IS the
  # flock on the open fd, not the file's content, so a lock file nothing
  # ever deletes is harmless -- and the RuntimeDirectory being torn down
  # and recreated between systemd unit runs is exactly what let a beet
  # session's held lock and the next scheduled import's fresh-file lock
  # talk past each other.
  [ "$output" = "/var/lib/musicbox/musicbox-import.lock" ]

  export MUSICBOX_IMPORT_LOCK_FILE="$WORK_DIR/custom.lock"
  run musicbox_lock_file
  [ "$status" -eq 0 ]
  [ "$output" = "$WORK_DIR/custom.lock" ]
}

@test "music-beet heals a missing lock directory instead of dying under set -e" {
  # Stands in for the unset-default case (MUSICBOX_IMPORT_LOCK_FILE unset,
  # the default resolving under /var/lib/musicbox before anything has ever
  # created it): a fixture path whose parent directory does not exist yet,
  # so `exec 200>"$LOCK_FILE"` would hit ENOENT and, under `set -e`, take
  # the whole interactive session down with it -- unless
  # musicbox_ensure_lock_dir's mkdir -p ran first. No import is running, so
  # music-beet should acquire the lock immediately and run beet for real.
  export MUSICBOX_IMPORT_LOCK_FILE="$WORK_DIR/not-yet-created/musicbox-import.lock"
  [ ! -d "$WORK_DIR/not-yet-created" ]

  cat >"$STUB_BIN/beet" <<'EOF'
#!/usr/bin/env bash
echo ran >"$BEET_RAN_FILE"
EOF
  chmod +x "$STUB_BIN/beet"
  export BEET_RAN_FILE="$WORK_DIR/beet-ran"

  run "$LIB_ROOT/scripts/music-beet" status
  [ "$status" -eq 0 ]
  [ -f "$BEET_RAN_FILE" ]
  [ -f "$MUSICBOX_IMPORT_LOCK_FILE" ]
}

@test "music-import and music-beet resolve the lock path through musicbox_lock_file, not a private default" {
  grep -qx 'LOCK_FILE="$(musicbox_lock_file)"' "$LIB_ROOT/scripts/music-import"
  grep -qx 'LOCK_FILE="$(musicbox_lock_file)"' "$LIB_ROOT/scripts/music-beet"
}

@test "music-import and music-beet contend on the identical lock file" {
  export MUSICBOX_IMPORT_LOCK_FILE="$WORK_DIR/shared.lock"

  # Stand in for "the other script already has it": hold the shared lock
  # externally, and don't proceed until that hold is actually in effect, so
  # neither script below can race the holder for the lock.
  flock "$MUSICBOX_IMPORT_LOCK_FILE" sleep 5 &
  local holder_pid=$!
  local waited=0
  while flock -n "$MUSICBOX_IMPORT_LOCK_FILE" -c true 2>/dev/null; do
    waited=$((waited + 1))
    [ "$waited" -lt 50 ] || break
    sleep 0.1
  done

  # music-import must see the lock held by that external process and bail
  # immediately rather than block or proceed.
  run "$LIB_ROOT/scripts/music-import"
  [ "$status" -eq 0 ]
  [[ "$output" == *"another run holds the lock"* ]]

  # music-beet must block on that same lock rather than proceed, then run
  # `beet` for real once the external holder releases it.
  cat >"$STUB_BIN/beet" <<'EOF'
#!/usr/bin/env bash
echo ran >"$BEET_RAN_FILE"
EOF
  chmod +x "$STUB_BIN/beet"
  export BEET_RAN_FILE="$WORK_DIR/beet-ran"
  local beet_output="$WORK_DIR/beet-output.log"

  "$LIB_ROOT/scripts/music-beet" status >"$beet_output" 2>&1 &
  local beet_pid=$!

  wait "$holder_pid"
  wait "$beet_pid"

  grep -q "another music-import run holds the lock; waiting" "$beet_output"
  [ -f "$BEET_RAN_FILE" ]
}

# --- safe_extract --------------------------------------------------------

@test "safe_extract rejects an archive over the byte cap" {
  truncate -s 2M "$WORK_DIR/big.tar"
  run safe_extract "$WORK_DIR/big.tar" "$WORK_DIR/dest" 1048576
  [ "$status" -ne 0 ]
  [ ! -d "$WORK_DIR/dest" ] || [ -z "$(ls -A "$WORK_DIR/dest")" ]
}

@test "safe_extract rejects a path-traversal member" {
  mkdir -p "$WORK_DIR/src/a/b"
  echo secret >"$WORK_DIR/src/secret.txt"
  (cd "$WORK_DIR/src/a/b" && bsdtar -cf "$WORK_DIR/evil.tar" ../../secret.txt)
  run safe_extract "$WORK_DIR/evil.tar" "$WORK_DIR/dest" 1000000
  [ "$status" -ne 0 ]
  [ ! -e "$WORK_DIR/dest/secret.txt" ]
}

@test "safe_extract extracts a well-formed archive within the cap" {
  mkdir -p "$WORK_DIR/src"
  echo payload >"$WORK_DIR/src/track.txt"
  (cd "$WORK_DIR/src" && bsdtar -cf "$WORK_DIR/good.tar" track.txt)
  run safe_extract "$WORK_DIR/good.tar" "$WORK_DIR/dest" 1000000
  [ "$status" -eq 0 ]
  [ -f "$WORK_DIR/dest/track.txt" ]
  [ "$(cat "$WORK_DIR/dest/track.txt")" = "payload" ]
}

# --- free_space_ok --------------------------------------------------------

@test "free_space_ok succeeds against a trivially low threshold" {
  run free_space_ok "$WORK_DIR" 0
  [ "$status" -eq 0 ]
}

@test "free_space_ok fails against an unreachable threshold" {
  run free_space_ok "$WORK_DIR" 999999999
  [ "$status" -ne 0 ]
}

# --- notify ----------------------------------------------------------------

@test "notify pipes a message to msmtp addressed to the recipient" {
  cat >"$STUB_BIN/msmtp" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$@" >"$MSMTP_STUB_ARGS_FILE"
cat >"$MSMTP_STUB_STDIN_FILE"
EOF
  chmod +x "$STUB_BIN/msmtp"
  export MSMTP_STUB_ARGS_FILE="$WORK_DIR/msmtp-args"
  export MSMTP_STUB_STDIN_FILE="$WORK_DIR/msmtp-stdin"
  run notify "someone@example.com" "Skipped tracks" "One file needs review."
  [ "$status" -eq 0 ]
  grep -qx "someone@example.com" "$MSMTP_STUB_ARGS_FILE"
  grep -q "Subject: Skipped tracks" "$MSMTP_STUB_STDIN_FILE"
  grep -q "One file needs review." "$MSMTP_STUB_STDIN_FILE"
}

# --- ping_hc -----------------------------------------------------------

# Logs the whole argument list on one line, so a test asserts the exact curl
# invocation -- retry and timeout flags included -- and not merely the URL.
# Also resolves and logs the HTTP method curl would actually send (the
# argument after -X/--request, or curl's own default of GET when neither is
# present) to a second file, so a test can assert the method a plain call
# with no -X would otherwise let slip past unnoticed -- this is exactly the
# gap that let every failure ping this estate has ever sent go out as a GET
# against a Worker route that only accepts POST on /fail, silently 404ing.
stub_curl() {
  cat >"$STUB_BIN/curl" <<'EOF'
#!/usr/bin/env bash
echo "$*" >>"$CURL_STUB_LOG"
method="GET"
prev=""
for arg in "$@"; do
  case "$prev" in
  -X | --request) method="$arg" ;;
  esac
  prev="$arg"
done
echo "$method" >>"$CURL_STUB_METHOD_LOG"
EOF
  chmod +x "$STUB_BIN/curl"
  export CURL_STUB_LOG="$WORK_DIR/curl-log"
  export CURL_STUB_METHOD_LOG="$WORK_DIR/curl-method-log"
}

@test "ping_hc curls the URL from the matching env var, retried and time-bounded, as a GET" {
  stub_curl
  export MUSICBOX_HC_IMPORT_URL="https://hc.example/ping/abc123"
  run ping_hc "import"
  [ "$status" -eq 0 ]
  grep -qx -- "-fsS --retry 3 --max-time 10 -o /dev/null https://hc.example/ping/abc123" "$CURL_STUB_LOG"
  grep -qx "GET" "$CURL_STUB_METHOD_LOG"
}

@test "ping_hc appends /fail and sends POST for a failure ping" {
  stub_curl
  export MUSICBOX_HC_BACKUP_URL="https://hc.example/ping/def456"
  run ping_hc "backup" fail
  [ "$status" -eq 0 ]
  grep -qx -- "-fsS --retry 3 --max-time 10 -X POST -o /dev/null https://hc.example/ping/def456/fail" "$CURL_STUB_LOG"
  grep -qx "POST" "$CURL_STUB_METHOD_LOG"
}

@test "ping_hc no-ops with a notice when the env var is unset" {
  unset MUSICBOX_HC_DISK_URL
  run ping_hc "disk"
  [ "$status" -eq 0 ]
  [[ "$output" == *"MUSICBOX_HC_DISK_URL"* ]]
}

@test "ping_hc no-ops with a notice when the env var is set but empty" {
  export MUSICBOX_HC_DISK_URL=""
  run ping_hc "disk"
  [ "$status" -eq 0 ]
  [[ "$output" == *"MUSICBOX_HC_DISK_URL"* ]]
}

# --- quality_gate --------------------------------------------------------

# make_flac <path> [rate]: a real, valid FLAC, synthesized rather than
# shipped as a binary fixture. ffmpeg and flac are both present on this
# workstation (brew install ffmpeg flac) and the source is fully
# deterministic: a one-second 440Hz sine tone, encoded at rate (default
# 44100) in 16-bit.
make_flac() {
  local dest="$1" rate="${2:-44100}" wav="$WORK_DIR/gen.wav"
  ffmpeg -hide_banner -loglevel error -y -f lavfi -i "sine=frequency=440:duration=1" \
    -ar "$rate" -sample_fmt s16 "$wav"
  flac --best -f -o "$dest" "$wav" >/dev/null 2>&1
}

@test "quality_gate fails a non-FLAC extension without inspecting the file" {
  mkdir -p "$WORK_DIR/album"
  echo not-really-audio >"$WORK_DIR/album/track.mp3"
  run quality_gate "$WORK_DIR/album"
  [ "$status" -ne 0 ]
  [[ "$output" == "$WORK_DIR/album/track.mp3"$'\t'*"FLAC"* ]]
}

@test "quality_gate fails a corrupt/truncated flac" {
  mkdir -p "$WORK_DIR/album"
  make_flac "$WORK_DIR/album/track.flac"
  truncate -s 2000 "$WORK_DIR/album/track.flac"
  run quality_gate "$WORK_DIR/album"
  [ "$status" -ne 0 ]
  [[ "$output" == "$WORK_DIR/album/track.flac"$'\t'*"decode"* ]]
}

@test "quality_gate fails a flac below 44.1kHz" {
  mkdir -p "$WORK_DIR/album"
  make_flac "$WORK_DIR/album/track.flac" 22050
  run quality_gate "$WORK_DIR/album"
  [ "$status" -ne 0 ]
  [[ "$output" == "$WORK_DIR/album/track.flac"$'\t'*"44100"* ]]
}

# The passing case and the out-of-scope-extension case are one test: art and
# sleeve notes sit beside the tracks in every real album directory, so a gate
# that admits a valid FLAC must ignore its siblings in the same walk.
@test "quality_gate passes a valid 44.1kHz/16-bit flac and ignores non-audio siblings" {
  mkdir -p "$WORK_DIR/album"
  make_flac "$WORK_DIR/album/track.flac"
  echo cover >"$WORK_DIR/album/cover.jpg"
  echo notes >"$WORK_DIR/album/notes.nfo"
  run quality_gate "$WORK_DIR/album"
  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

@test "quality_gate fails an 8-bit flac" {
  mkdir -p "$WORK_DIR/album"
  # ffmpeg's wav muxer needs an explicit 8-bit PCM codec (-sample_fmt alone
  # targets the encoder's default codec, which rejects u8); piped straight
  # into flac rather than staged as an intermediate wav file.
  ffmpeg -hide_banner -loglevel error -y -f lavfi -i "sine=frequency=440:duration=1" \
    -ar 44100 -acodec pcm_u8 -f wav - 2>/dev/null |
    flac --best -f -o "$WORK_DIR/album/track.flac" - >/dev/null 2>&1
  run quality_gate "$WORK_DIR/album"
  [ "$status" -ne 0 ]
  [[ "$output" == "$WORK_DIR/album/track.flac"$'\t'*"8-bit"* ]]
}

# --- unique_dest_path ------------------------------------------------------

# Freezes the HHMMSS suffix unique_dest_path reaches for on a collision, so
# the expected path below is a literal rather than a value the test has to
# recompute and race across a second boundary.
stub_date() {
  cat >"$STUB_BIN/date" <<'EOF'
#!/usr/bin/env bash
echo 120000
EOF
  chmod +x "$STUB_BIN/date"
}

@test "unique_dest_path returns the plain path when nothing is there yet" {
  mkdir -p "$WORK_DIR/dest"
  run unique_dest_path "$WORK_DIR/dest" "Bad Album"
  [ "$status" -eq 0 ]
  [ "$output" = "$WORK_DIR/dest/Bad Album" ]
}

@test "unique_dest_path appends a time suffix on collision" {
  mkdir -p "$WORK_DIR/dest/Bad Album"
  stub_date
  run unique_dest_path "$WORK_DIR/dest" "Bad Album"
  [ "$status" -eq 0 ]
  [ "$output" = "$WORK_DIR/dest/Bad Album-120000" ]
}

@test "unique_dest_path falls back to a numeric suffix if the time-based name also collides" {
  mkdir -p "$WORK_DIR/dest/Bad Album" "$WORK_DIR/dest/Bad Album-120000"
  stub_date
  run unique_dest_path "$WORK_DIR/dest" "Bad Album"
  [ "$status" -eq 0 ]
  [ "$output" = "$WORK_DIR/dest/Bad Album-120000.2" ]
}

# --- safe_mv_to_review -----------------------------------------------------

@test "safe_mv_to_review moves into an empty destination directory" {
  mkdir -p "$WORK_DIR/staging/Good Album"
  echo track >"$WORK_DIR/staging/Good Album/one.flac"
  run safe_mv_to_review "$WORK_DIR/staging/Good Album" "$WORK_DIR/review"
  [ "$status" -eq 0 ]
  [ "$output" = "$WORK_DIR/review/Good Album" ]
  [ -f "$WORK_DIR/review/Good Album/one.flac" ]
  [ ! -e "$WORK_DIR/staging/Good Album" ]
}

@test "safe_mv_to_review disambiguates instead of merging into an existing same-named destination" {
  mkdir -p "$WORK_DIR/staging/Bad Album" "$WORK_DIR/review/Bad Album"
  echo first >"$WORK_DIR/review/Bad Album/keep.flac"
  echo second >"$WORK_DIR/staging/Bad Album/new.flac"
  run safe_mv_to_review "$WORK_DIR/staging/Bad Album" "$WORK_DIR/review"
  [ "$status" -eq 0 ]
  [ "$output" != "$WORK_DIR/review/Bad Album" ]
  [ -f "$WORK_DIR/review/Bad Album/keep.flac" ]
  [ -f "$output/new.flac" ]
}

@test "safe_mv_to_review warns and returns nonzero instead of aborting when the move fails" {
  run safe_mv_to_review "$WORK_DIR/does-not-exist" "$WORK_DIR/review"
  [ "$status" -ne 0 ]
  [[ "$output" == *"could not move"* ]]
}

@test "safe_mv_to_review warns and returns nonzero instead of aborting when the dest dir can't be created" {
  mkdir -p "$WORK_DIR/staging/Good Album" "$WORK_DIR/readonly-parent"
  chmod 0555 "$WORK_DIR/readonly-parent"
  run safe_mv_to_review "$WORK_DIR/staging/Good Album" "$WORK_DIR/readonly-parent/nested"
  # Restored before the assertions, so a failing one still leaves teardown
  # a writable tree to sweep.
  chmod 0755 "$WORK_DIR/readonly-parent"
  [ "$status" -ne 0 ]
  [[ "$output" == *"could not create"* ]]
  [ -e "$WORK_DIR/staging/Good Album" ]
}

# --- musicbox-disk-check ---------------------------------------------------

# Stands in for the real df on PATH: the script only ever reads the fifth
# (Capacity) column, so this fixes every other column to plausible values
# and drives the one that matters. $pct is baked in at stub-creation time,
# not read from the environment at run time, since each test wants its own
# fixed reading rather than a value the script under test could see.
stub_df() {
  local pct="$1"
  cat >"$STUB_BIN/df" <<EOF
#!/usr/bin/env bash
echo "Filesystem 1024-blocks Used Available Capacity Mounted on"
echo "/dev/sda1 100000 50000 50000 ${pct}% /"
EOF
  chmod +x "$STUB_BIN/df"
}

@test "musicbox-disk-check pings success when usage is under the threshold" {
  stub_df 42
  stub_curl
  export MUSICBOX_HC_DISK_URL="https://hc.example/ping/disk123"
  run "$LIB_ROOT/scripts/musicbox-disk-check"
  [ "$status" -eq 0 ]
  grep -qx -- "-fsS --retry 3 --max-time 10 -o /dev/null https://hc.example/ping/disk123" "$CURL_STUB_LOG"
  grep -qx "GET" "$CURL_STUB_METHOD_LOG"
}

@test "musicbox-disk-check pings failure and exits nonzero when usage exceeds the threshold" {
  stub_df 95
  stub_curl
  export MUSICBOX_HC_DISK_URL="https://hc.example/ping/disk123"
  run "$LIB_ROOT/scripts/musicbox-disk-check"
  [ "$status" -ne 0 ]
  grep -qx -- "-fsS --retry 3 --max-time 10 -X POST -o /dev/null https://hc.example/ping/disk123/fail" "$CURL_STUB_LOG"
  grep -qx "POST" "$CURL_STUB_METHOD_LOG"
}

@test "musicbox-disk-check honors an overridden MUSICBOX_DISK_THRESHOLD_PCT" {
  stub_df 80
  stub_curl
  export MUSICBOX_HC_DISK_URL="https://hc.example/ping/disk123"
  export MUSICBOX_DISK_THRESHOLD_PCT=75
  run "$LIB_ROOT/scripts/musicbox-disk-check"
  [ "$status" -ne 0 ]
  grep -qx -- "-fsS --retry 3 --max-time 10 -X POST -o /dev/null https://hc.example/ping/disk123/fail" "$CURL_STUB_LOG"
}

# --- musicbox-navidrome-check -----------------------------------------------

# stub_curl plus a per-URL exit status, appended rather than restated so
# both stubs keep logging calls and methods the one same way. curl is
# called twice by this script: once against Navidrome's own /rest/ping
# (NAVIDROME_PING_STUB_EXIT decides whether that call succeeds), and once by
# ping_hc against the healthchecks URL, which always succeeds here -- the
# point of each test is which ping_hc branch fired, not whether curl itself
# could reach healthchecks.io.
stub_curl_navidrome() {
  stub_curl
  cat >>"$STUB_BIN/curl" <<'EOF'
last="${@: -1}"
case "$last" in
*/rest/ping) exit "${NAVIDROME_PING_STUB_EXIT:-0}" ;;
*) exit 0 ;;
esac
EOF
}

@test "musicbox-navidrome-check pings success when /rest/ping is reachable" {
  stub_curl_navidrome
  export NAVIDROME_PING_STUB_EXIT=0
  export NAVIDROME_URL="http://localhost:4533"
  export MUSICBOX_HC_NAVIDROME_URL="https://hc.example/ping/nd123"
  run "$LIB_ROOT/scripts/musicbox-navidrome-check"
  [ "$status" -eq 0 ]
  grep -qx -- "-fsS --retry 3 --max-time 10 -o /dev/null http://localhost:4533/rest/ping" "$CURL_STUB_LOG"
  grep -qx -- "-fsS --retry 3 --max-time 10 -o /dev/null https://hc.example/ping/nd123" "$CURL_STUB_LOG"
}

@test "musicbox-navidrome-check pings failure and exits nonzero when /rest/ping is unreachable" {
  stub_curl_navidrome
  export NAVIDROME_PING_STUB_EXIT=1
  export NAVIDROME_URL="http://localhost:4533"
  export MUSICBOX_HC_NAVIDROME_URL="https://hc.example/ping/nd123"
  run "$LIB_ROOT/scripts/musicbox-navidrome-check"
  [ "$status" -ne 0 ]
  grep -qx -- "-fsS --retry 3 --max-time 10 -o /dev/null http://localhost:4533/rest/ping" "$CURL_STUB_LOG"
  grep -qx -- "-fsS --retry 3 --max-time 10 -X POST -o /dev/null https://hc.example/ping/nd123/fail" "$CURL_STUB_LOG"
}

# --- music-backup -----------------------------------------------------

# Stands in for rclone: logs every invocation so a test can assert whether
# a sync was ever attempted, without touching the network. Always exits 0,
# since these tests are about whether music-backup reaches rclone at all,
# never about a sync failure.
stub_rclone() {
  cat >"$STUB_BIN/rclone" <<'EOF'
#!/usr/bin/env bash
echo "$*" >>"$RCLONE_STUB_LOG"
exit 0
EOF
  chmod +x "$STUB_BIN/rclone"
  export RCLONE_STUB_LOG="$WORK_DIR/rclone-log"
}

@test "music-backup aborts with a failure ping when LIBRARY_DIR is missing, before any rclone sync" {
  stub_curl
  stub_rclone
  export MUSICBOX_ROOT="$WORK_DIR/srv"
  mkdir -p "$MUSICBOX_ROOT" # no library/ subdirectory underneath
  export MUSICBOX_BACKUP_LOCK_FILE="$WORK_DIR/backup.lock"
  export MUSICBOX_R2_ACCESS_KEY_ID="fake" MUSICBOX_R2_SECRET_ACCESS_KEY="fake"
  export MUSICBOX_HC_BACKUP_URL="https://hc.example/ping/backup123"

  run "$LIB_ROOT/scripts/music-backup"
  [ "$status" -ne 0 ]
  [[ "$output" == *"does not exist"* ]]
  [ ! -e "$RCLONE_STUB_LOG" ]
  grep -qx -- "-fsS --retry 3 --max-time 10 -X POST -o /dev/null https://hc.example/ping/backup123/fail" "$CURL_STUB_LOG"
}

@test "music-backup aborts with a failure ping when LIBRARY_DIR is empty, before any rclone sync" {
  stub_curl
  stub_rclone
  export MUSICBOX_ROOT="$WORK_DIR/srv"
  mkdir -p "$MUSICBOX_ROOT/library"
  export MUSICBOX_BACKUP_LOCK_FILE="$WORK_DIR/backup.lock"
  export MUSICBOX_R2_ACCESS_KEY_ID="fake" MUSICBOX_R2_SECRET_ACCESS_KEY="fake"
  export MUSICBOX_HC_BACKUP_URL="https://hc.example/ping/backup123"

  run "$LIB_ROOT/scripts/music-backup"
  [ "$status" -ne 0 ]
  [[ "$output" == *"is empty"* ]]
  [ ! -e "$RCLONE_STUB_LOG" ]
  grep -qx -- "-fsS --retry 3 --max-time 10 -X POST -o /dev/null https://hc.example/ping/backup123/fail" "$CURL_STUB_LOG"
}

@test "music-backup skips cleanly, without touching rclone, when its own lock is already held" {
  stub_curl
  stub_rclone
  export MUSICBOX_BACKUP_LOCK_FILE="$WORK_DIR/backup.lock"
  export MUSICBOX_HC_BACKUP_URL="https://hc.example/ping/backup123"

  # Stand in for "a previous backup run is still going": hold the backup
  # lock externally, and don't proceed until that hold is actually in
  # effect, so music-backup below cannot race the holder for the lock.
  flock "$MUSICBOX_BACKUP_LOCK_FILE" sleep 5 &
  local holder_pid=$!
  local waited=0
  while flock -n "$MUSICBOX_BACKUP_LOCK_FILE" -c true 2>/dev/null; do
    waited=$((waited + 1))
    [ "$waited" -lt 50 ] || break
    sleep 0.1
  done

  # No MUSICBOX_ROOT, LIBRARY_DIR, or R2 credentials are set up: the lock
  # guard must be what stops this run, ahead of every later guard.
  run "$LIB_ROOT/scripts/music-backup"
  wait "$holder_pid"
  [ "$status" -eq 0 ]
  [[ "$output" == *"another run holds the backup lock"* ]]
  [ ! -e "$RCLONE_STUB_LOG" ]
  grep -qx -- "-fsS --retry 3 --max-time 10 -o /dev/null https://hc.example/ping/backup123" "$CURL_STUB_LOG"
}

@test "music-backup's lock default is a persistent path, not the RuntimeDirectory tmpfs" {
  # Same hazard musicbox_lock_file's default move closed for music-import:
  # RuntimeDirectory=musicbox is shared with music-import.service, and
  # systemd removes it -- without reference counting -- whenever EITHER
  # unit stops, which could unlink a backup run's lock file mid-sync.
  grep -qx 'BACKUP_LOCK_FILE="${MUSICBOX_BACKUP_LOCK_FILE:-/var/lib/musicbox/musicbox-backup.lock}"' "$LIB_ROOT/scripts/music-backup"
}

@test "music-backup heals a missing lock directory instead of dying under set -e" {
  stub_curl
  stub_rclone
  # Stands in for the unset-default case, the same way the music-beet
  # regression test does: a fixture path whose parent directory does not
  # exist yet, so `exec 201>"$BACKUP_LOCK_FILE"` would hit ENOENT under
  # `set -e` without musicbox_ensure_lock_dir's mkdir -p running first.
  export MUSICBOX_BACKUP_LOCK_FILE="$WORK_DIR/not-yet-created/musicbox-backup.lock"
  [ ! -d "$WORK_DIR/not-yet-created" ]
  export MUSICBOX_ROOT="$WORK_DIR/srv"
  mkdir -p "$MUSICBOX_ROOT" # no library/ subdirectory underneath
  export MUSICBOX_R2_ACCESS_KEY_ID="fake" MUSICBOX_R2_SECRET_ACCESS_KEY="fake"
  export MUSICBOX_HC_BACKUP_URL="https://hc.example/ping/backup123"

  run "$LIB_ROOT/scripts/music-backup"
  # Proof the run got PAST lock acquisition to the (unrelated) LIBRARY_DIR
  # guard, rather than dying on ENOENT from the missing lock directory: a
  # bare `exec` failure would abort with a different message and never
  # create the lock file at all.
  [ "$status" -ne 0 ]
  [[ "$output" == *"does not exist"* ]]
  [ -f "$MUSICBOX_BACKUP_LOCK_FILE" ]
}
