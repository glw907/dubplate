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
  # The real template, not a test fixture copy: exercises exactly what
  # deploy.sh ships to the box.
  export MUSICBOX_REJECT_EMAIL_TEMPLATE="$REPO_ROOT/config/reject-email.txt"
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

# Synthesizes a real, valid 44.1kHz/16-bit FLAC at dest rather than a
# fixture of fake bytes: the quality gate runs `flac -t` for real over every
# staged track, so any test whose album must pass the gate needs a file that
# genuinely does. The scratch wav is reused across calls and swept with
# WORK_DIR; it lives outside MUSICBOX_ROOT, so the gate never sees it.
make_flac() {
  local dest="$1" wav="$WORK_DIR/gen.wav"
  ffmpeg -hide_banner -loglevel error -y -f lavfi -i "sine=frequency=440:duration=1" \
    -ar 44100 -sample_fmt s16 "$wav"
  flac --best -f -o "$dest" "$wav" >/dev/null 2>&1
}

# Stages one album in alice's inbox that the quality gate must reject: a
# genuinely valid FLAC beside a file carrying an .mp3 extension, the whole
# tree aged past the quiescence window so this run picks it up.
stage_bad_album() {
  local dir="$MUSICBOX_ROOT/inbox/alice/$1"
  mkdir -p "$dir"
  make_flac "$dir/track1.flac"
  echo not-really-audio >"$dir/track2.mp3"
  touch_old "$MUSICBOX_ROOT/inbox/alice"
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
  make_flac "$MUSICBOX_ROOT/inbox/alice/Good Album/track1.flac"
  touch_old "$MUSICBOX_ROOT/inbox/alice"

  run "$SCRIPT"
  [ "$status" -eq 0 ]

  # A success ping (no /fail suffix) went out for import. "! grep" would
  # silently no-op here under bats' set -e (a leading "!" is exempt from
  # errexit, so a failing negated assertion that isn't the test's last
  # line never surfaces); run+status is the assertion that actually holds.
  grep -qx -- "-fsS --retry 3 --max-time 10 -o /dev/null https://hc.example/ping/import" "$CURL_LOG"
  run grep -q "/fail" "$CURL_LOG"
  [ "$status" -ne 0 ]

  # Nothing landed in review, and no notification fired.
  [ -z "$(find "$MUSICBOX_ROOT/review" -mindepth 1 -print -quit 2>/dev/null)" ]
  [ ! -s "$MSMTP_LOG" ]
}

@test "skip path: a beets skip is filed to review and notifies uploader plus Geoff" {
  stub_beet_with_skip
  mkdir -p "$MUSICBOX_ROOT/inbox/alice/Weird Album"
  make_flac "$MUSICBOX_ROOT/inbox/alice/Weird Album/track1.flac"
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
  run grep -q "bob@example.com" "$MSMTP_LOG"
  [ "$status" -ne 0 ]
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
  grep -qx -- "-fsS --retry 3 --max-time 10 -o /dev/null https://hc.example/ping/import/fail" "$CURL_LOG"
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

@test "lock held: another run fires a success ping and exits 0 without touching staging" {
  mkdir -p "$MUSICBOX_ROOT/inbox/alice/Good Album"
  echo fake-audio >"$MUSICBOX_ROOT/inbox/alice/Good Album/track1.flac"
  touch_old "$MUSICBOX_ROOT/inbox/alice"

  # A separate process holds the non-blocking flock the whole time this
  # run executes, standing in for a concurrent scheduled run or an
  # interactive `music-beet` session. Poll until it has actually acquired
  # the lock before invoking the script, so the test never races its own
  # fixture.
  flock "$MUSICBOX_IMPORT_LOCK_FILE" sleep 5 &
  holder_pid=$!
  # Taking the lock here still SUCCEEDING means the holder has not got it
  # yet, so the wait ends on the first failure. Bounded, so a holder that
  # never starts fails the assertions below instead of hanging the suite.
  for _ in {1..100}; do
    flock -n "$MUSICBOX_IMPORT_LOCK_FILE" -c true 2>/dev/null || break
    sleep 0.05
  done

  run "$SCRIPT"

  kill "$holder_pid" 2>/dev/null || true
  wait "$holder_pid" 2>/dev/null || true

  [ "$status" -eq 0 ]
  [[ "$output" == *"another run holds the lock"* ]]
  # Per the spec decision: a wedged import still reports healthy (a healthy
  # system, not the residual risk of a permanently wedged lock going
  # unnoticed), so the success ping fires before the early exit.
  grep -qx -- "-fsS --retry 3 --max-time 10 -o /dev/null https://hc.example/ping/import" "$CURL_LOG"
  run grep -q "/fail" "$CURL_LOG"
  [ "$status" -ne 0 ]
  # Nothing past the lock check ran: no staging dir, no notifications.
  [ ! -d "$MUSICBOX_ROOT/staging" ]
  [ ! -s "$MSMTP_LOG" ]
}

# --- quality gate (end to end) --------------------------------------------

@test "quality gate: an album with any failing file moves entirely to review" {
  stub_beet_clean
  stage_bad_album "Bad Album"

  run "$SCRIPT"
  [ "$status" -eq 0 ]

  # The whole album -- the passing track alongside the failing one --
  # moved to review; a partial import of just the good track never
  # happens.
  found="$(find "$MUSICBOX_ROOT/review" -type d -iname "Bad Album" -print -quit)"
  [ -n "$found" ]
  [[ "$found" == *"/review/"*"/alice/Bad Album" ]]
  [ -f "$found/track1.flac" ]
  [ -f "$found/track2.mp3" ]
  [ ! -e "$MUSICBOX_ROOT/staging/alice/Bad Album" ]

  # Alice gets the friendly consumer-facing template, addressed by her
  # first name, naming the album and explaining the MP3 in plain terms --
  # never quality_gate's raw reason text.
  grep -q "alice@example.com" "$MSMTP_LOG"
  grep -q "Hi alice," "$MSMTP_LOG"
  grep -q "Thanks for adding to the family music library" "$MSMTP_LOG"
  grep -q "Bad Album" "$MSMTP_LOG"
  grep -q "This is an MP3 file, and the library keeps lossless FLAC." "$MSMTP_LOG"

  # Geoff's copy stays technical: the raw path and reason.
  grep -q "geoff-login@907.life" "$MSMTP_LOG"
  grep -q "track2.mp3" "$MSMTP_LOG"
  grep -q "not FLAC" "$MSMTP_LOG"

  # Each reason line carries its own file's basename, so N failures read
  # as N distinct per-file lines rather than running together.
  grep -q "track2.mp3: This is an MP3 file, and the library keeps lossless FLAC." "$MSMTP_LOG"
}

@test "quality gate: an ampersand in an album name renders literally, not as the matched pattern" {
  stub_beet_clean
  stage_bad_album "Sturm & Drang"

  run "$SCRIPT"
  [ "$status" -eq 0 ]

  # Bash's ${var//pat/repl} treats an unquoted "&" in the replacement as
  # "insert the matched text" (confirmed on bash 5.3): an unquoted
  # $first_name/$name/$reason_lines substitution would corrupt this into
  # the literal placeholder text instead of the album name.
  grep -qF "Sturm & Drang" "$MSMTP_LOG"
  run grep -q "album or file name" "$MSMTP_LOG"
  [ "$status" -ne 0 ]
}

@test "quality gate: a same-day repeat quarantine of an identically named album doesn't clobber the first" {
  stub_beet_clean
  stage_bad_album "Bad Album"
  run "$SCRIPT"
  [ "$status" -eq 0 ]

  # A second upload under the exact same name later the same day: the
  # review dir already holds one "Bad Album" from the first run above.
  stage_bad_album "Bad Album"
  run "$SCRIPT"
  [ "$status" -eq 0 ]

  # Both copies survive under distinct names -- a colliding mv must never
  # abort the run under set -e -- and both runs reached the very end
  # (beet import, then the closing success ping), not just the quarantine
  # step.
  count="$(find "$MUSICBOX_ROOT/review" -mindepth 3 -maxdepth 3 -type d -iname "Bad Album*" | wc -l)"
  [ "$count" -eq 2 ]
  [ "$(grep -cx -- "-fsS --retry 3 --max-time 10 -o /dev/null https://hc.example/ping/import" "$CURL_LOG")" -eq 2 ]
  run grep -q "/fail" "$CURL_LOG"
  [ "$status" -ne 0 ]
}

@test "quality gate: a missing reject-email template skips the uploader email but still tells Geoff" {
  stub_beet_clean
  export MUSICBOX_REJECT_EMAIL_TEMPLATE="$WORK_DIR/does-not-exist.txt"
  stage_bad_album "Bad Album"

  run "$SCRIPT"
  [ "$status" -eq 0 ]
  run grep -q "alice@example.com" "$MSMTP_LOG"
  [ "$status" -ne 0 ]
  grep -q "geoff-login@907.life" "$MSMTP_LOG"
  grep -q "track2.mp3" "$MSMTP_LOG"
}

@test "quality gate: an album that can't be moved to review is excluded from import and reported to Geoff" {
  stub_beet_clean
  stage_bad_album "Bad Album"

  # Force the normal review destination to be unwritable: mkdir -p for
  # review/<date>/alice fails when review itself carries no write
  # permission, so safe_mv_to_review's primary move fails for a reason
  # that has nothing to do with a name collision.
  mkdir -p "$MUSICBOX_ROOT/review"
  chmod 0555 "$MUSICBOX_ROOT/review"

  run "$SCRIPT"

  chmod 0755 "$MUSICBOX_ROOT/review"

  [ "$status" -eq 0 ]

  # The album never reached review, and it must also be gone from staging:
  # beet import scans $STAGING_DIR recursively with no way to skip one
  # subdirectory, so anything left there is imported despite having failed
  # the gate. quarantine-failed is the holding dir it lands in instead.
  [ ! -e "$MUSICBOX_ROOT/staging/alice/Bad Album" ]
  found="$(find "$MUSICBOX_ROOT/quarantine-failed" -type d -iname "Bad Album" -print -quit)"
  [ -n "$found" ]
  [[ "$found" == *"/quarantine-failed/"*"/alice/Bad Album" ]]
  [ -f "$found/track1.flac" ]
  [ -f "$found/track2.mp3" ]

  # Geoff hears about it even though the review move failed: this case
  # must never go silent, which is the whole point.
  grep -q "geoff-login@907.life" "$MSMTP_LOG"
  grep -q "could not move; excluded from import" "$MSMTP_LOG"

  # No uploader-facing email for this one: nothing was filed to review
  # the normal, contributor-visible way.
  run grep -q "alice@example.com" "$MSMTP_LOG"
  [ "$status" -ne 0 ]
}
