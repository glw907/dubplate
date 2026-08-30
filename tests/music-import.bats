#!/usr/bin/env bats
# End-to-end tests for scripts/music-import against a fixture /srv/music
# tree. beet, curl, and msmtp are stubbed on PATH; rclone is stubbed too
# (unused by music-import itself, but stubbing it keeps a real rclone
# binary off PATH from ever being invoked by accident). Real bsdtar, yq,
# and jq run for real: they are deterministic and already exercised by
# tests/lib.bats.

setup() {
  REPO_ROOT="$(cd "$(dirname "$BATS_TEST_FILENAME")/.." && pwd)"
  SCRIPT="$REPO_ROOT/scripts/music-import"

  WORK_DIR="$(mktemp -d)"
  export MUSICBOX_ROOT="$WORK_DIR/srv-music"
  mkdir -p "$MUSICBOX_ROOT/inbox" "$MUSICBOX_ROOT/library" "$MUSICBOX_ROOT/state"

  export MUSICBOX_IMPORT_LOCK_FILE="$WORK_DIR/musicbox-import.lock"
  export MUSICBOX_CONTRIBUTORS_FILE="$WORK_DIR/contributors.yaml"
  export MUSICBOX_GEOFF_EMAIL="geoff-login@907.life"
  export MUSICBOX_QUIESCENT_MINUTES=10
  export MUSICBOX_MIN_FREE_GB=0
  export MUSICBOX_MAX_ARCHIVE_BYTES=10000000
  export MUSICBOX_REVIEW_RETENTION_DAYS=60
  export MUSICBOX_HC_IMPORT_URL="https://hc.example/ping/import"
  unset NAVIDROME_ADMIN_USER NAVIDROME_ADMIN_PASSWORD 2>/dev/null || true

  cat >"$MUSICBOX_CONTRIBUTORS_FILE" <<'EOF'
contributors:
  - name: alice
    email: alice@example.com
    role: contributor
  - name: bob
    email: bob@example.com
    role: contributor
EOF

  STUB_BIN="$WORK_DIR/stub-bin"
  mkdir -p "$STUB_BIN"
  PATH="$STUB_BIN:$PATH"

  MSMTP_LOG="$WORK_DIR/msmtp.log"
  CURL_LOG="$WORK_DIR/curl.log"
  : >"$MSMTP_LOG"
  : >"$CURL_LOG"
  export MSMTP_LOG CURL_LOG

  cat >"$STUB_BIN/msmtp" <<'EOF'
#!/usr/bin/env bash
{
  echo "--- args: $*"
  cat
} >>"$MSMTP_LOG"
EOF

  cat >"$STUB_BIN/curl" <<'EOF'
#!/usr/bin/env bash
echo "$*" >>"$CURL_LOG"
echo '{"token":"faketoken"}'
EOF

  cat >"$STUB_BIN/rclone" <<'EOF'
#!/usr/bin/env bash
echo "rclone $*" >>"$WORK_DIR_MARKER/rclone.log"
EOF
  # rclone's stub needs WORK_DIR at generation time, not at call time.
  sed -i "s#\$WORK_DIR_MARKER#$WORK_DIR#" "$STUB_BIN/rclone"

  chmod +x "$STUB_BIN/msmtp" "$STUB_BIN/curl" "$STUB_BIN/rclone"

  touch_old() {
    find "$1" -exec touch -d "30 minutes ago" {} +
  }
}

teardown() {
  rm -rf "$WORK_DIR"
}

# beet stub used by the happy-path test: every album is a clean auto-tag,
# nothing is written to the logfile, matching how real beets behaves when
# quiet mode finds confident matches for everything.
stub_beet_clean() {
  cat >"$STUB_BIN/beet" <<'EOF'
#!/usr/bin/env bash
# args: import -q -l <logfile> <staging-dir>
: # nothing to log; every album auto-tagged cleanly
EOF
  chmod +x "$STUB_BIN/beet"
}

# beet stub used by the skip-path test: logs exactly one "skip" line for a
# path under alice's staging subdirectory, in beets' own tag_log format
# ("skip <path>"), and leaves that path in place, exactly as real beets
# does for a file it declines to move.
stub_beet_with_skip() {
  cat >"$STUB_BIN/beet" <<EOF
#!/usr/bin/env bash
logfile=""
prev=""
for arg in "\$@"; do
  if [ "\$prev" = "-l" ]; then
    logfile="\$arg"
  fi
  prev="\$arg"
done
echo "skip $MUSICBOX_ROOT/staging/alice/Weird Album" >> "\$logfile"
EOF
  chmod +x "$STUB_BIN/beet"
}

@test "happy path: quiescent upload imports cleanly, no review items, success ping" {
  stub_beet_clean
  mkdir -p "$MUSICBOX_ROOT/inbox/alice/Good Album"
  echo fake-audio >"$MUSICBOX_ROOT/inbox/alice/Good Album/track1.flac"
  touch_old "$MUSICBOX_ROOT/inbox/alice"

  run "$SCRIPT"
  [ "$status" -eq 0 ]

  # A success ping (no /fail suffix) went out for import.
  grep -qx -- "-fsS -o /dev/null https://hc.example/ping/import" "$CURL_LOG"
  ! grep -q "/fail" "$CURL_LOG"

  # Nothing landed in review, and no notification fired.
  [ -z "$(find "$MUSICBOX_ROOT/review" -mindepth 1 -print -quit 2>/dev/null)" ]
  [ ! -s "$MSMTP_LOG" ]
}

@test "skip path: a beets skip is filed to review and notifies uploader plus Geoff" {
  stub_beet_with_skip
  mkdir -p "$MUSICBOX_ROOT/inbox/alice/Weird Album"
  echo fake-audio >"$MUSICBOX_ROOT/inbox/alice/Weird Album/track1.flac"
  touch_old "$MUSICBOX_ROOT/inbox/alice"

  run "$SCRIPT"
  [ "$status" -eq 0 ]

  # The skipped album was relocated under review/<date>/alice/.
  found="$(find "$MUSICBOX_ROOT/review" -type d -iname "Weird Album" -print -quit)"
  [ -n "$found" ]
  [[ "$found" == *"/review/"*"/alice/Weird Album" ]]
  [ ! -e "$MUSICBOX_ROOT/staging/alice/Weird Album" ]

  # Both the uploader and Geoff were notified, and nobody else was.
  grep -q "alice@example.com" "$MSMTP_LOG"
  grep -q "geoff-login@907.life" "$MSMTP_LOG"
  ! grep -q "bob@example.com" "$MSMTP_LOG"
}

@test "insufficient free space aborts before beet runs and fires a fail ping" {
  stub_beet_clean
  mkdir -p "$MUSICBOX_ROOT/inbox/alice/Good Album"
  echo fake-audio >"$MUSICBOX_ROOT/inbox/alice/Good Album/track1.flac"
  touch_old "$MUSICBOX_ROOT/inbox/alice"
  export MUSICBOX_MIN_FREE_GB=999999999

  # If beet ran, it would exit nonzero here so the test could not mistake a
  # false pass for the real abort-before-import guarantee.
  cat >"$STUB_BIN/beet" <<'EOF'
#!/usr/bin/env bash
exit 17
EOF
  chmod +x "$STUB_BIN/beet"

  run "$SCRIPT"
  [ "$status" -ne 0 ]
  grep -qx -- "-fsS -o /dev/null https://hc.example/ping/import/fail" "$CURL_LOG"
}

@test "a non-quiescent inbox is left alone this run" {
  stub_beet_clean
  mkdir -p "$MUSICBOX_ROOT/inbox/alice/Mid Upload"
  echo fake-audio >"$MUSICBOX_ROOT/inbox/alice/Mid Upload/track1.flac"
  # Deliberately not aged: this upload looks like it is still in progress.

  run "$SCRIPT"
  [ "$status" -eq 0 ]
  [ -e "$MUSICBOX_ROOT/inbox/alice/Mid Upload/track1.flac" ]
  [ ! -e "$MUSICBOX_ROOT/staging/alice" ]
}
