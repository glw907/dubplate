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
stub_curl() {
  cat >"$STUB_BIN/curl" <<'EOF'
#!/usr/bin/env bash
echo "$*" >>"$CURL_STUB_LOG"
EOF
  chmod +x "$STUB_BIN/curl"
  export CURL_STUB_LOG="$WORK_DIR/curl-log"
}

@test "ping_hc curls the URL from the matching env var, retried and time-bounded" {
  stub_curl
  export MUSICBOX_HC_IMPORT_URL="https://hc.example/ping/abc123"
  run ping_hc "import"
  [ "$status" -eq 0 ]
  grep -qx -- "-fsS --retry 3 --max-time 10 -o /dev/null https://hc.example/ping/abc123" "$CURL_STUB_LOG"
}

@test "ping_hc appends /fail for a failure ping" {
  stub_curl
  export MUSICBOX_HC_BACKUP_URL="https://hc.example/ping/def456"
  run ping_hc "backup" fail
  [ "$status" -eq 0 ]
  grep -qx -- "-fsS --retry 3 --max-time 10 -o /dev/null https://hc.example/ping/def456/fail" "$CURL_STUB_LOG"
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
