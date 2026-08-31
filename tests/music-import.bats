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

  # Symlink-resolved at creation, so the path comparisons below (the cache
  # lands under MUSICBOX_RUNTIME_DIR test resolves the real path from inside
  # the run) hold on a host where /tmp is itself a symlink.
  WORK_DIR="$(cd "$(mktemp -d)" && pwd -P)"
  export MUSICBOX_ROOT="$WORK_DIR/srv-music"
  mkdir -p "$MUSICBOX_ROOT/inbox" "$MUSICBOX_ROOT/library" "$MUSICBOX_ROOT/state"

  # Deliberately NOT $WORK_DIR itself -- a distinct subdirectory, so this
  # fixture can't accidentally pass by coincidence if the RUNTIME_DIR/
  # LOCK_FILE decoupling below ever regressed back to deriving RUNTIME_DIR
  # from the lock file's own directory (see the dedicated regression test
  # below). Pre-created here, not left to the script's own mkdir -p: a
  # couple of tests below flock this path directly, from outside the
  # script under test, before the script ever runs.
  export MUSICBOX_IMPORT_LOCK_FILE="$WORK_DIR/lockdir/musicbox-import.lock"
  mkdir -p "$(dirname "$MUSICBOX_IMPORT_LOCK_FILE")"
  # Stands in for systemd's RuntimeDirectory=musicbox (/run/musicbox on the
  # box, created fresh every boot): the scan-trigger temp credential file
  # and the contributors.yaml JSON cache both live here, deliberately
  # decoupled from the lock file's own (now persistent) directory. WORK_DIR
  # already exists via mktemp -d above.
  export MUSICBOX_RUNTIME_DIR="$WORK_DIR"
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
  CURL_CONFIG_LOG="$WORK_DIR/curl-config.log"
  CURL_METHOD_LOG="$WORK_DIR/curl-method.log"
  : >"$MSMTP_LOG"
  : >"$CURL_LOG"
  : >"$CURL_CONFIG_LOG"
  : >"$CURL_METHOD_LOG"
  export MSMTP_LOG CURL_LOG CURL_CONFIG_LOG CURL_METHOD_LOG

  # MSMTP_FAIL makes every call fail after still consuming and logging its
  # stdin, standing in for a stalled mail queue or a bad /etc/msmtprc: the
  # message was attempted, delivery just didn't happen.
  cat >"$STUB_BIN/msmtp" <<'EOF'
#!/usr/bin/env bash
{
  echo "--- args: $*"
  cat
} >>"$MSMTP_LOG"
if [ -n "${MSMTP_FAIL:-}" ]; then
  echo "stub msmtp: forced failure" >&2
  exit 1
fi
EOF

  # Logs every call's arguments, and -- for a --config call, which is
  # trigger_navidrome_scan's Subsonic request -- copies that file's contents
  # to a second log, since the real function deletes it the moment the
  # request is done. Setting CURL_FAIL_STARTSCAN makes only the startScan
  # call fail, so the failure-tolerant path can be exercised without also
  # breaking ping_hc's own curl calls. Also resolves and logs the HTTP
  # method curl would actually send (the argument after -X/--request, or
  # curl's own default of GET when neither is present), one line per call,
  # so a test can tell a real POST from a plain call that silently defaults
  # to GET -- the gap that let every failure ping the estate ever sent go
  # out as a GET against a Worker route that 404s anything but POST.
  # CURL_FAIL_ALL fails every call regardless of what it's for, standing in
  # for a dead network path that takes out ping_hc's own status pings along
  # with the Navidrome trigger.
  cat >"$STUB_BIN/curl" <<'EOF'
#!/usr/bin/env bash
echo "$*" >>"$CURL_LOG"
config_file=""
method="GET"
prev=""
for arg in "$@"; do
  if [ "$prev" = "--config" ]; then
    config_file="$arg"
  fi
  case "$prev" in
  -X | --request) method="$arg" ;;
  esac
  prev="$arg"
done
echo "$method" >>"$CURL_METHOD_LOG"
if [ -n "$config_file" ]; then
  cat "$config_file" >>"$CURL_CONFIG_LOG"
  if [ -n "${CURL_FAIL_STARTSCAN:-}" ] && grep -q "rest/startScan" "$config_file"; then
    echo "stub curl: forced startScan failure" >&2
    exit 22
  fi
fi
if [ -n "${CURL_FAIL_ALL:-}" ]; then
  echo "stub curl: forced failure" >&2
  exit 22
fi
EOF

  cat >"$STUB_BIN/rclone" <<'EOF'
#!/usr/bin/env bash
echo "rclone $*" >>"$WORK_DIR_MARKER/rclone.log"
EOF
  # rclone's stub needs WORK_DIR at generation time, not at call time.
  sed -i "s#\$WORK_DIR_MARKER#$WORK_DIR#" "$STUB_BIN/rclone"

  # Counts invocations without changing behavior: one line per call in
  # YQ_LOG, then execs the real yq (resolved now, before the wrapper itself
  # exists on PATH, so this lookup can't find itself) so every test still
  # gets yq's real, deterministic parse of contributors.yaml. Exists to let
  # the caching test below assert "at most once per run" against something
  # other than the parsed output, which looks identical whether it came
  # from one parse or five.
  REAL_YQ="$(command -v yq)"
  YQ_LOG="$WORK_DIR/yq.log"
  : >"$YQ_LOG"
  export YQ_LOG
  cat >"$STUB_BIN/yq" <<EOF
#!/usr/bin/env bash
echo call >>"$YQ_LOG"
exec "$REAL_YQ" "\$@"
EOF

  chmod +x "$STUB_BIN/msmtp" "$STUB_BIN/curl" "$STUB_BIN/rclone" "$STUB_BIN/yq"

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

# Stages one album in alice's inbox that the quality gate must accept: a
# single genuinely valid FLAC, aged past the quiescence window so this run
# picks it up.
stage_good_album() {
  local dir="$MUSICBOX_ROOT/inbox/alice/$1"
  mkdir -p "$dir"
  make_flac "$dir/track1.flac"
  touch_old "$MUSICBOX_ROOT/inbox/alice"
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

# General form of stub_beet_with_skip: logs one "skip" line per <user>
# <album> pair given as arguments (e.g. `stub_beet_with_skips carol "Weird
# Album C" alice "Weird Album A"`), each under that user's staging
# subdirectory. Lets the unmatched-uploader tests name whichever uploaders,
# and however many at once, each scenario needs instead of always alice's
# single fixed album.
stub_beet_with_skips() {
  local user album
  {
    cat <<'EOF'
#!/usr/bin/env bash
logfile=""
prev=""
for arg in "$@"; do
  if [ "$prev" = "-l" ]; then
    logfile="$arg"
  fi
  prev="$arg"
done
EOF
    while [ "$#" -gt 0 ]; do
      user="$1"
      album="$2"
      shift 2
      printf 'echo "skip %s" >> "$logfile"\n' "$MUSICBOX_ROOT/staging/$user/$album"
    done
  } >"$STUB_BIN/beet"
  chmod +x "$STUB_BIN/beet"
}

@test "happy path: quiescent upload imports cleanly, no review items, success ping" {
  stub_beet_clean
  stage_good_album "Good Album"

  run "$SCRIPT"
  [ "$status" -eq 0 ]

  # A success ping (no /fail suffix) went out for import, as a plain GET.
  # "! grep" would silently no-op here under bats' set -e (a leading "!" is
  # exempt from errexit, so a failing negated assertion that isn't the
  # test's last line never surfaces); run+status is the assertion that
  # actually holds.
  grep -qx -- "-fsS --retry 3 --max-time 10 -o /dev/null https://hc.example/ping/import" "$CURL_LOG"
  grep -qx "GET" "$CURL_METHOD_LOG"
  run grep -q "/fail" "$CURL_LOG"
  [ "$status" -ne 0 ]

  # Nothing landed in review, and no notification fired.
  [ -z "$(find "$MUSICBOX_ROOT/review" -mindepth 1 -print -quit 2>/dev/null)" ]
  [ ! -s "$MSMTP_LOG" ]
}

@test "Navidrome scan trigger: calls Subsonic startScan via a curl config file, never argv" {
  stub_beet_clean
  export NAVIDROME_ADMIN_USER="admin"
  export NAVIDROME_ADMIN_PASSWORD="s3cret"
  stage_good_album "Good Album"

  run "$SCRIPT"
  [ "$status" -eq 0 ]

  # The startScan request happened -- u/t/s/v/c query-param auth -- but only
  # inside the --config file's contents, never as a curl argument. The
  # password itself appears in neither: t is its md5 with the salt.
  grep -q "rest/startScan" "$CURL_CONFIG_LOG"
  grep -qE "[&?]u=admin" "$CURL_CONFIG_LOG"
  grep -qE "[&?]t=[0-9a-f]{32}" "$CURL_CONFIG_LOG"
  grep -qE "[&?]s=[0-9a-f]{16}" "$CURL_CONFIG_LOG"
  run grep -q "rest/startScan" "$CURL_LOG"
  [ "$status" -ne 0 ]
  run grep -q "s3cret" "$CURL_LOG"
  [ "$status" -ne 0 ]
  run grep -q "s3cret" "$CURL_CONFIG_LOG"
  [ "$status" -ne 0 ]
}

@test "Navidrome scan trigger: a failed startScan request logs but doesn't fail the run" {
  stub_beet_clean
  export NAVIDROME_ADMIN_USER="admin"
  export NAVIDROME_ADMIN_PASSWORD="s3cret"
  export CURL_FAIL_STARTSCAN=1
  stage_good_album "Good Album"

  run "$SCRIPT"
  [ "$status" -eq 0 ]
  [[ "$output" == *"Navidrome scan request failed."* ]]
  # The rest of the run still completed: a success ping still fired.
  grep -qx -- "-fsS --retry 3 --max-time 10 -o /dev/null https://hc.example/ping/import" "$CURL_LOG"
}

@test "a failed mail delivery does not abort the run: notify's best-effort contract holds under set -e" {
  stub_beet_with_skip
  export MSMTP_FAIL=1
  mkdir -p "$MUSICBOX_ROOT/inbox/alice/Weird Album"
  make_flac "$MUSICBOX_ROOT/inbox/alice/Weird Album/track1.flac"
  touch_old "$MUSICBOX_ROOT/inbox/alice"

  run "$SCRIPT"
  [ "$status" -eq 0 ]

  # notify was called bare -- the skip path notifies both alice and Geoff --
  # and every msmtp invocation failed, yet the run still reached the end:
  # the skipped album was filed, and the closing success ping fired.
  found="$(find "$MUSICBOX_ROOT/review" -type d -iname "Weird Album" -print -quit)"
  [ -n "$found" ]
  [[ "$output" == *"msmtp delivery"*"failed"* ]]
  grep -qx -- "-fsS --retry 3 --max-time 10 -o /dev/null https://hc.example/ping/import" "$CURL_LOG"
}

@test "a failed curl does not abort the run: ping_hc's best-effort contract holds under set -e" {
  stub_beet_clean
  export CURL_FAIL_ALL=1
  stage_good_album "Good Album"

  run "$SCRIPT"
  [ "$status" -eq 0 ]

  # The closing ping_hc call was attempted -- logged even though the stub
  # failed it -- and the run still exited 0 instead of dying on ping_hc's
  # own nonzero return.
  grep -qx -- "-fsS --retry 3 --max-time 10 -o /dev/null https://hc.example/ping/import" "$CURL_LOG"
  [[ "$output" == *"ping to"*"did not succeed"* ]]
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

@test "insufficient free space aborts before any archive is extracted" {
  stub_beet_clean
  mkdir -p "$MUSICBOX_ROOT/inbox/alice"
  local src="$WORK_DIR/src-album"
  mkdir -p "$src/Good Album"
  make_flac "$src/Good Album/track1.flac"
  (cd "$src" && bsdtar -cf "$MUSICBOX_ROOT/inbox/alice/upload.tar" "Good Album")
  touch_old "$MUSICBOX_ROOT/inbox/alice"
  export MUSICBOX_MIN_FREE_GB=999999999

  # If beet ran, it would exit nonzero here, so the test could not mistake a
  # false pass for the real abort-before-extraction guarantee.
  cat >"$STUB_BIN/beet" <<'EOF'
#!/usr/bin/env bash
exit 17
EOF
  chmod +x "$STUB_BIN/beet"

  run "$SCRIPT"
  [ "$status" -ne 0 ]
  grep -qx -- "-fsS --retry 3 --max-time 10 -X POST -o /dev/null https://hc.example/ping/import/fail" "$CURL_LOG"

  # The archive was never touched: still sitting whole in alice's inbox,
  # never unpacked anywhere, and staging never even got a directory for her.
  [ -f "$MUSICBOX_ROOT/inbox/alice/upload.tar" ]
  [ ! -d "$MUSICBOX_ROOT/staging/alice" ]
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
  # The failure ping must be a real POST: the pings Worker routes
  # /ping/<slug>/fail on POST only and 404s a bare GET, so a plain call
  # here would be silently dropped rather than recorded.
  grep -qx -- "-fsS --retry 3 --max-time 10 -X POST -o /dev/null https://hc.example/ping/import/fail" "$CURL_LOG"
  grep -qx "POST" "$CURL_METHOD_LOG"
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

# --- staging isolation -----------------------------------------------------

@test "staging: two archives whose top-level entries collide by name land side by side, not merged" {
  stub_beet_clean
  mkdir -p "$MUSICBOX_ROOT/inbox/alice"

  local src1="$WORK_DIR/src1" src2="$WORK_DIR/src2"
  mkdir -p "$src1/Album" "$src2/Album"
  make_flac "$src1/Album/track1.flac"
  make_flac "$src2/Album/track1.flac"
  (cd "$src1" && bsdtar -cf "$MUSICBOX_ROOT/inbox/alice/one.tar" Album)
  (cd "$src2" && bsdtar -cf "$MUSICBOX_ROOT/inbox/alice/two.tar" Album)
  touch_old "$MUSICBOX_ROOT/inbox/alice"

  run "$SCRIPT"
  [ "$status" -eq 0 ]

  # Both archives extract into their own private scratch subdirectory
  # first, so the second archive's top-level "Album" never lands straight
  # into the first archive's already-staged "Album" -- it survives beside
  # it under a disambiguated name instead of merging into (or clobbering a
  # same-named file inside) the first.
  count="$(find "$MUSICBOX_ROOT/staging/alice" -mindepth 1 -maxdepth 1 -type d -iname "Album*" | wc -l)"
  [ "$count" -eq 2 ]
  for d in "$MUSICBOX_ROOT/staging/alice"/Album*; do
    [ "$(find "$d" -name '*.flac' | wc -l)" -eq 1 ]
  done

  # No leftover scratch directories from the per-archive extraction step.
  [ -z "$(find "$MUSICBOX_ROOT/staging/alice" -maxdepth 1 -iname '.extract-*' -print -quit)" ]
}

# --- quality gate (end to end) --------------------------------------------

@test "quality gate: a symlink in an upload is quarantined instead of silently bypassing the FLAC-only gate" {
  stub_beet_clean
  local dir="$MUSICBOX_ROOT/inbox/alice/Symlinked Album"
  mkdir -p "$dir"
  make_flac "$dir/track1.flac"
  ln -s track1.flac "$dir/sneaky.flac"
  touch_old "$MUSICBOX_ROOT/inbox/alice"

  run "$SCRIPT"
  [ "$status" -eq 0 ]

  # The whole album -- symlink included -- was quarantined; beets never saw
  # it, because -type f alone would have let the symlink glide straight
  # through the gate unexamined.
  found="$(find "$MUSICBOX_ROOT/review" -type d -iname "Symlinked Album" -print -quit)"
  [ -n "$found" ]
  [ -L "$found/sneaky.flac" ]
  [ -f "$found/track1.flac" ]
  [ ! -e "$MUSICBOX_ROOT/staging/alice/Symlinked Album" ]

  # Geoff's technical copy names the file and the reason.
  grep -q "geoff-login@907.life" "$MSMTP_LOG"
  grep -q "sneaky.flac" "$MSMTP_LOG"
  grep -q "symlink not allowed" "$MSMTP_LOG"

  # Alice gets the friendly, translated version -- never the raw reason.
  grep -q "alice@example.com" "$MSMTP_LOG"
  grep -q "sneaky.flac: This looks like a symlink" "$MSMTP_LOG"
}

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

# --- unknown uploader (absent from contributors.yaml) ----------------------
#
# "carol" never appears in the setup() fixture's contributors.yaml, so
# uploader_email/uploader_name return empty for her in both callers that
# use them: move_skips_and_notify (the beets-skip path) and
# quarantine_quality_failures (the quality-gate path). Each caller has its
# own "no email on file" branch (music-import:241 and :373); these tests
# exercise both, each once with carol alone and once alongside alice (a
# real contributor), to prove the unmatched branch changes only the
# notification -- never where the file lands or whether the run succeeds --
# and never bleeds into a sibling upload from a real contributor in the same
# run.

@test "skip path: an unmatched uploader is not emailed, but Geoff still hears about it and the run succeeds" {
  stub_beet_with_skips carol "Weird Album"
  mkdir -p "$MUSICBOX_ROOT/inbox/carol/Weird Album"
  make_flac "$MUSICBOX_ROOT/inbox/carol/Weird Album/track1.flac"
  touch_old "$MUSICBOX_ROOT/inbox/carol"

  run "$SCRIPT"
  [ "$status" -eq 0 ]

  # Same filing behavior as a known uploader's skip: relocated under
  # review/<date>/carol/, gone from staging.
  found="$(find "$MUSICBOX_ROOT/review" -type d -iname "Weird Album" -print -quit)"
  [ -n "$found" ]
  [[ "$found" == *"/review/"*"/carol/Weird Album" ]]
  [ ! -e "$MUSICBOX_ROOT/staging/carol/Weird Album" ]

  # The diagnostic fired, and Geoff still learned carol had a skip -- but
  # nobody was emailed on her behalf: only his own address appears, in
  # exactly one message.
  [[ "$output" == *"no email on file for uploader 'carol'; skipping their notice"* ]]
  grep -q "geoff-login@907.life" "$MSMTP_LOG"
  grep -q "carol" "$MSMTP_LOG"
  [ "$(grep -c '^--- args:' "$MSMTP_LOG")" -eq 1 ]
}

@test "skip path: a known and an unmatched uploader together -- only the known one is emailed" {
  stub_beet_with_skips carol "Weird Album C" alice "Weird Album A"
  mkdir -p "$MUSICBOX_ROOT/inbox/carol/Weird Album C" "$MUSICBOX_ROOT/inbox/alice/Weird Album A"
  make_flac "$MUSICBOX_ROOT/inbox/carol/Weird Album C/track1.flac"
  make_flac "$MUSICBOX_ROOT/inbox/alice/Weird Album A/track1.flac"
  touch_old "$MUSICBOX_ROOT/inbox/carol"
  touch_old "$MUSICBOX_ROOT/inbox/alice"

  run "$SCRIPT"
  [ "$status" -eq 0 ]

  # Both skips were filed regardless of contributor status.
  found="$(find "$MUSICBOX_ROOT/review" -type d -iname "Weird Album C" -print -quit)"
  [[ "$found" == *"/review/"*"/carol/Weird Album C" ]]
  found="$(find "$MUSICBOX_ROOT/review" -type d -iname "Weird Album A" -print -quit)"
  [[ "$found" == *"/review/"*"/alice/Weird Album A" ]]

  # alice, a real contributor, got her own notice; carol did not -- only
  # alice's and Geoff's aggregate (which still names both) went out.
  [[ "$output" == *"no email on file for uploader 'carol'; skipping their notice"* ]]
  run grep -q "no email on file for uploader 'alice'" <<<"$output"
  [ "$status" -ne 0 ]
  grep -q "alice@example.com" "$MSMTP_LOG"
  grep -q "geoff-login@907.life" "$MSMTP_LOG"
  grep -q "carol" "$MSMTP_LOG"
  [ "$(grep -c '^--- args:' "$MSMTP_LOG")" -eq 2 ]
}

@test "quality gate: an unmatched uploader's quarantine skips their email but Geoff still gets the technical detail" {
  local dir="$MUSICBOX_ROOT/inbox/carol/Bad Album"
  mkdir -p "$dir"
  make_flac "$dir/track1.flac"
  echo not-really-audio >"$dir/track2.mp3"
  touch_old "$MUSICBOX_ROOT/inbox/carol"
  stub_beet_clean

  run "$SCRIPT"
  [ "$status" -eq 0 ]

  # Same filing behavior as a known uploader's quality-gate failure.
  found="$(find "$MUSICBOX_ROOT/review" -type d -iname "Bad Album" -print -quit)"
  [ -n "$found" ]
  [[ "$found" == *"/review/"*"/carol/Bad Album" ]]
  [ ! -e "$MUSICBOX_ROOT/staging/carol/Bad Album" ]

  [[ "$output" == *"no email on file for uploader 'carol'; skipping their quality-gate notice"* ]]
  # Geoff's technical summary still names the file and the raw reason; his
  # is the only message that went out.
  grep -q "geoff-login@907.life" "$MSMTP_LOG"
  grep -q "carol/Bad Album" "$MSMTP_LOG"
  grep -q "track2.mp3" "$MSMTP_LOG"
  grep -q "not FLAC" "$MSMTP_LOG"
  [ "$(grep -c '^--- args:' "$MSMTP_LOG")" -eq 1 ]
}

@test "quality gate: a known and an unmatched uploader together -- only the known one gets the rejection email" {
  stage_bad_album "Bad Album A"
  local dir="$MUSICBOX_ROOT/inbox/carol/Bad Album C"
  mkdir -p "$dir"
  make_flac "$dir/track1.flac"
  echo not-really-audio >"$dir/track2.mp3"
  touch_old "$MUSICBOX_ROOT/inbox/carol"
  stub_beet_clean

  run "$SCRIPT"
  [ "$status" -eq 0 ]

  found="$(find "$MUSICBOX_ROOT/review" -type d -iname "Bad Album A" -print -quit)"
  [[ "$found" == *"/review/"*"/alice/Bad Album A" ]]
  found="$(find "$MUSICBOX_ROOT/review" -type d -iname "Bad Album C" -print -quit)"
  [[ "$found" == *"/review/"*"/carol/Bad Album C" ]]

  # alice gets the friendly template; carol gets nothing, only the
  # diagnostic and a mention in Geoff's technical summary.
  [[ "$output" == *"no email on file for uploader 'carol'; skipping their quality-gate notice"* ]]
  run grep -q "no email on file for uploader 'alice'" <<<"$output"
  [ "$status" -ne 0 ]
  grep -q "alice@example.com" "$MSMTP_LOG"
  grep -q "Hi alice," "$MSMTP_LOG"
  grep -q "geoff-login@907.life" "$MSMTP_LOG"
  grep -q "carol/Bad Album C" "$MSMTP_LOG"
  [ "$(grep -c '^--- args:' "$MSMTP_LOG")" -eq 2 ]
}

# --- contributors.yaml parse cache -----------------------------------------

@test "contributors.yaml is parsed to JSON at most once per run, even with multiple uploader lookups" {
  # A beets skip for a good album (uploader_email, via move_skips_and_notify)
  # plus a quality-gate quarantine for a bad album (uploader_email and
  # uploader_name, via quarantine_quality_failures) together drive three
  # uploader lookups in one run. Before caching, each lookup re-spent its
  # own yq|jq pipeline over the same static file; now they all share one
  # parse.
  stage_good_album "Good Album"
  stage_bad_album "Bad Album"
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
echo "skip $MUSICBOX_ROOT/staging/alice/Good Album" >> "\$logfile"
EOF
  chmod +x "$STUB_BIN/beet"

  run "$SCRIPT"
  [ "$status" -eq 0 ]

  # Both notification paths actually fired -- otherwise a passing yq.log
  # count would just mean uploader_email/uploader_name were never called.
  found="$(find "$MUSICBOX_ROOT/review" -type d -iname "Good Album" -print -quit)"
  [ -n "$found" ]
  found="$(find "$MUSICBOX_ROOT/review" -type d -iname "Bad Album" -print -quit)"
  [ -n "$found" ]
  grep -q "alice@example.com" "$MSMTP_LOG"

  [ "$(wc -l <"$YQ_LOG")" -eq 1 ]
}

@test "the contributors.yaml JSON cache is created owner-only, not world/group readable" {
  # RUNTIME_DIR (systemd's RuntimeDirectory=musicbox on the box) is
  # world-traversable at 0755 with no UMask= override, and
  # contributors.yaml itself is deployed 0640 root:music specifically so
  # its contributor names and emails aren't world-readable -- a cache file
  # created under the default umask would undo that. Since the real
  # trap removes the cache the moment the script exits, the only way to
  # observe its mode is from inside the run: this yq stub, invoked with
  # its stdout already redirected to the cache file, stats that file
  # through its own fd 1 (via the "magic" /proc/PID/fd/1 symlink,
  # dereferenced with `stat -L`) the instant the real load happens. The
  # log write goes out fd 3, not fd 1 itself -- writing the stat result to
  # fd 1 would repoint fd 1 at the log file before the stat call ever ran,
  # making it observe the log file's own mode instead of the cache's.
  cat >"$STUB_BIN/yq" <<EOF
#!/usr/bin/env bash
exec 3>>"$WORK_DIR/cache-mode.log"
stat -L -c '%a' /proc/\$\$/fd/1 >&3
exec 3>&-
exec "$REAL_YQ" "\$@"
EOF
  chmod +x "$STUB_BIN/yq"

  stub_beet_clean
  stage_bad_album "Bad Album"

  run "$SCRIPT"
  [ "$status" -eq 0 ]

  [ -s "$WORK_DIR/cache-mode.log" ]
  [ "$(cat "$WORK_DIR/cache-mode.log")" = "600" ]
}

@test "the contributors.yaml JSON cache lands under MUSICBOX_RUNTIME_DIR, not beside the lock file" {
  # Guards the RUNTIME_DIR/LOCK_FILE decoupling itself, not just its
  # symptom: setup() points MUSICBOX_IMPORT_LOCK_FILE at $WORK_DIR/lockdir
  # (a subdirectory distinct from MUSICBOX_RUNTIME_DIR=$WORK_DIR) precisely
  # so this can fail. If RUNTIME_DIR were ever re-derived from the lock
  # file's own directory (music-import's pre-M2b behavior), the cache
  # would land one level down, under lockdir, instead of directly under
  # WORK_DIR -- caught here by resolving the cache file's real path from
  # inside the run, the same /proc/PID/fd/1 trick the owner-only mode test
  # above uses, since the trap deletes it before bats could ever inspect
  # it after the fact.
  cat >"$STUB_BIN/yq" <<EOF
#!/usr/bin/env bash
exec 3>>"$WORK_DIR/cache-path.log"
readlink -f /proc/\$\$/fd/1 >&3
exec 3>&-
exec "$REAL_YQ" "\$@"
EOF
  chmod +x "$STUB_BIN/yq"

  stub_beet_clean
  stage_bad_album "Bad Album"

  run "$SCRIPT"
  [ "$status" -eq 0 ]

  [ -s "$WORK_DIR/cache-path.log" ]
  cache_path="$(cat "$WORK_DIR/cache-path.log")"
  [ "$(dirname "$cache_path")" = "$WORK_DIR" ]
}

@test "no uploader lookup at all means contributors.yaml is never parsed" {
  stub_beet_clean
  stage_good_album "Good Album"

  run "$SCRIPT"
  [ "$status" -eq 0 ]

  [ ! -s "$YQ_LOG" ]
}

# --- quarantine-failed retention sweep --------------------------------------

@test "quarantine-failed gets the same retention sweep review already has" {
  stub_beet_clean

  local old_dir="$MUSICBOX_ROOT/quarantine-failed/2020-01-01"
  local fresh_dir="$MUSICBOX_ROOT/quarantine-failed/$(date +%Y-%m-%d)"
  mkdir -p "$old_dir/alice" "$fresh_dir/alice"
  # Backdates only the dated directory's own mtime, past the 60-day
  # MUSICBOX_REVIEW_RETENTION_DAYS set in setup(); the sweep checks that
  # mtime directly (-maxdepth 1), not anything recursively underneath it.
  touch -d "90 days ago" "$old_dir"

  run "$SCRIPT"
  [ "$status" -eq 0 ]

  [ ! -d "$old_dir" ]
  [ -d "$fresh_dir" ]
}
