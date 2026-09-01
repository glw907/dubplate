# dubplate Go rewrite: implementation plan

**Goal:** Replace the box-side shell layer with one Go binary per the approved
design spec, migrated per-subcommand with rollback intact, and rename the
estate from musicbox to dubplate throughout (project, repo, paths, binary,
env, checks) — ratified by Geoff 2026-08-31 after a vetted naming brainstorm.

**Spec:** `~/.dotfiles/docs/superpowers/specs/2026-08-30-dubplate-go-rewrite-design.md`
(approved by Geoff 2026-08-30 after four adversarial reviews). The plan argues
from the spec; executors read both.

**Budget:** ceiling 4M tokens; checkpoint (STATUS write) every 4 tasks.
**Execution:** a FRESH session executes this plan (decided 2026-08-31; the
refinement-pass session closes after its own close-out and does not launch
this). Entry point: read this plan, the spec, and musicbox `docs/STATUS.md`,
verify no other executor holds the repo, then run the pass-execute workflow,
sequential. Implementer `general-purpose` pinned `model: sonnet`; reviewer
`diff-reviewer` (opus). Repo gate: `bash scripts/check.sh` (which gains
`make check` in T1). The repo directory is still `~/Projects/musicbox` at
start; T14 renames it.

## Global constraints (every task)

- **Read `/var/home/glw907/.claude/skills/go-conventions/SKILL.md` before
  writing any Go.** Every file, test, error string, and comment conforms.
  Highlights that bind this plan: no interfaces until a second consumer exists
  in-tree; plain config structs, no builder patterns; `ctx` first parameter,
  never stored in a struct; `exec.CommandContext` for every subprocess;
  errors lowercase, context-first, `%w` only when callers branch;
  `slog.NewTextHandler` to stderr inside `internal/`; table-driven tests with
  the assertion-discipline rules (no silent-success fakes, no self-derived
  expectations); atomic writes via `internal/atomicfile`; module path
  `github.com/glw907/dubplate`, Go version matching the installed toolchain.
- Behavior source of truth is the shell layer at fix-wave close (repo HEAD
  when this pass starts) plus its bats suites. Each task's criteria say which
  bats cases pin it and how they map: ported as-is, re-expressed (curl-argv
  and `--config` assertions describe curl, not the contract), or new-seam.
- Site-identifying config fields are required with no defaults; only behavior
  thresholds default (spec, Packages).
- No live-box access in T1–T11. T12 and T13 name their box steps explicitly.
- Nothing is committed with a red gate; the shell layer keeps passing its bats
  suites untouched until the task that deletes it.

---

### T1: Module scaffold, CLI skeleton, gate wiring

**Files:** `go.mod`, `go.sum`, `Makefile`, `.golangci.yml`,
`cmd/dubplate/main.go`, `cmd/dubplate/root.go`, `cmd/dubplate/version.go`,
`internal/atomicfile/atomicfile.go`, modify `scripts/check.sh`.

**Produces:** `newRootCmd() *cobra.Command` with `SilenceUsage` and
`SilenceErrors` set, subcommand stubs absent (added by later tasks); the
conventions `main()` shape (main is the only printer and the only `os.Exit`);
`atomicfile.WriteFile(path string, data []byte, perm os.FileMode) error`.

**Criteria:** `make check` (build, vet, golangci-lint per the conventions
config, test) passes and is invoked from `scripts/check.sh` so one gate covers
both layers; `dubplate version` prints a version string; cross-compile target
`make build-linux` produces a static `CGO_ENABLED=0 GOOS=linux GOARCH=amd64`
binary; `.golangci.yml` is the conventions v2 config verbatim, verified with
`golangci-lint config verify`. No exit-code mapper: every failure exits 1
(the shell contracts need no other code; lock contention is a nil return, see
T9).

### T2: Config and contributors roster

**Files:** `internal/config/config.go`, `internal/config/contributors.go`,
tests alongside.

**Produces:**

```go
func Load() (*Config, error)   // reads process env only; no file parser
type Config struct {
    AdminEmail    string            // required, no default
    SMTP          SMTPConfig        // Host, Port, User, Password, From: required
    R2            R2Config          // AccountID, Bucket, Remote, AccessKeyID, SecretAccessKey: required
    Navidrome     NavidromeConfig   // URL, AdminUser, AdminPassword: required
    Contributors  string            // path, required
    RejectEmailTemplate string      // path, required
    BeetsConfig   string            // path, required
    BeetsDB       string            // path, required
    NavidromeDB   string            // path, required
    Paths         Paths             // LibraryDir, InboxDir, StagingDir, StateDir, RunDir, ReviewDir, QuarantineDir, BackupStaging
    Thresholds    Thresholds        // QuiescentMin, FreeFloorGB, ArchiveCapGB, RetentionDays, DiskPct: defaulted
    PingURLs      map[string]string // DUBPLATE_PING_<SLUG>_URL prefix scan; empty map is valid
    Collections   []string          // collection names; first is the default; defaults to one unnamed default collection (today's layout)
}
var ErrMissing = errors.New("required setting unset") // Load wraps with the field name
func LoadContributors(path string) (Roster, error)
func (r Roster) EmailFor(name string) (string, bool)
func (r Roster) FirstName(name string) (string, bool) // first word of a roster-confirmed name
func (r Roster) Has(name string) bool
```

**Criteria:** Load returns `ErrMissing`-wrapped errors naming each unset
required field (all of them, via `errors.Join`, not first-wins — doctor prints
the full list); thresholds carry the shell layer's current default values,
copied from HEAD and cited in the task report; `PingURLs` populated by prefix
scan of the process environment; contributors parsing replaces the `yq | jq`
pair and is pinned by re-expressed bats cases for missing-file and
unknown-uploader semantics (empty roster, lookups report absent, no error).
The YAML dependency is the module's only one beyond cobra, go-mail, and
`golang.org/x/sys` (T3 needs it) and is named in the report. **No later task
may introduce a config field**: this struct is the complete inventory, and a
later task discovering a missing field escalates instead of adding one. Env
is read through a single internal values getter so the packaging
initiative's file overlay can be added without touching call sites (the
spec's composable-source seam; only env is wired now).

### T3: Lock and ping client

**Files:** `internal/lock/lock.go`, `internal/ping/ping.go`, tests.

**Produces:**

```go
var ErrHeld = errors.New("lock held")
func Acquire(path string) (*Lock, error)                       // non-blocking flock
func AcquireBlocking(ctx context.Context, path string) (*Lock, error)
func (l *Lock) Release() error
func (l *Lock) File() *os.File     // fd survives syscall.Exec for `dubplate beet`

func NewClient() *Client   // defaults: 3 retries, 10s timeout (the shell's curl budget)
type Client struct { Timeout time.Duration; Retries int }
func (c *Client) Success(ctx context.Context, url string) error // GET url; nil when url == ""
func (c *Client) Fail(ctx context.Context, url string) error    // POST url + "/fail" (trailing slash stripped first); nil when url == ""
```

**Criteria:** one lock-path definition consumed later by import, beet, and
backup (T9/T10); two processes contending is proven with a real flock in a
`t.TempDir()`. `Acquire` clears `FD_CLOEXEC` on the lock fd
(`unix.FcntlInt(fd, unix.F_SETFD, 0)`) — Go opens files `O_CLOEXEC`, so
without this the lock drops at `syscall.Exec` and `dubplate beet`'s
guarantee is false; a unit test asserts the flag is clear. Ping is provider-agnostic per the spec: URL in, no slug or
Worker knowledge; the fail verb is POST (matching the fix wave's corrected
contract — verify against the wave's updated `lib.sh`/bats before coding);
retry and timeout budgets copied from the shell's curl flags and pinned by
re-expressed tests against `httptest.Server`, including a
records-the-method assertion that fails if fail regresses to GET.

### T4: Notify with msmtp parity

**Files:** `internal/notify/notify.go`, tests; `go.mod` gains
`github.com/wneessen/go-mail` and the test server dep.

**Produces:**

```go
func New(cfg config.SMTPConfig) *Mailer
func (m *Mailer) Send(ctx context.Context, to, subject, body string) error
```

**Criteria:** implicit TLS on port 465 works (the msmtprc contract:
`tls on`, `tls_starttls off`); From, Date, and Message-ID are synthesized;
non-ASCII subjects ("Sigur Rós") round-trip via RFC 2047; bodies are
quoted-printable. Parity is proven against a local TLS SMTP test server
(`github.com/mhale/smtpd` with a self-signed cert generated in TestMain)
asserting on the received message's headers and decoded body — not on
"send returned nil". Delivery logging is one slog record per send (journald
replaces `msmtp.log`; say so in the report). msmtp itself is untouched until
T13.

### T5: Reject taxonomy and email rendering

**Files:** `internal/reject/reject.go`, tests.

**Produces:**

```go
type Reason int   // ReasonNone = iota, ReasonBadExtension, ReasonSymlink, ReasonCorrupt, ReasonBelowCDQuality, ReasonTooLarge, ReasonUnpackFailed
func (r Reason) Friendly() string
type Rejection struct { Path string; Reason Reason }
func RenderEmail(templatePath, firstName, albumName string, rejections []Rejection) (string, error)
```

**Criteria:** the friendly strings match `friendly_reason()`'s wording at
HEAD verbatim, including "This FLAC is below CD quality."
(upload-check reuses this surface; wording is contract); `ReasonNone` is the
iota zero so a `gate.Verdict{OK: true}` carries no spurious reason; the
template fill covers all three placeholders — first name (from
`Roster.FirstName`), album-or-file name, per-file reason lines — pinned by
ported bats cases including the literal-`&` case (the shell's documented
quoting hazard becomes a plain test here); template file stays a deployed
runtime asset read at runtime, not embedded.

### T6: Archive extraction and quality gate

**Files:** `internal/archive/archive.go`, `internal/gate/gate.go`, tests with
fixture archives (reuse `tests/` fixtures where formats allow).

**Produces:**

```go
func Extract(ctx context.Context, archivePath, destDir string, lim Limits) error
type Limits struct { MaxExtractedBytes, MinFreeBytes int64 }
func FreeSpace(path string) (int64, error)

type Policy struct { AllowedExt []string; RequireDecode bool; MinSampleRate, MinBitsPerSample int; MetaflacPath string }  // built by config, not imported from it
type Verdict struct { Path string; OK bool; Reason reject.Reason }
func Check(ctx context.Context, root string, p Policy, flacPath string) ([]Verdict, error)
```

**Criteria:** extraction is native (`archive/zip`, `archive/tar`; the bsdtar
dependency dies), refuses path traversal, stops at `MaxExtractedBytes`, and
the free-space check runs before extraction (fix-wave behavior); each archive
extracts into the caller-supplied isolated dir. Gate detects symlinks as
`ReasonSymlink`, wrong extensions per policy, runs `flac -t` full decode
via `exec.CommandContext` when `RequireDecode` — the corrupt-FLAC bats case
ports as-is against a truncated-FLAC fixture — and enforces the CD-quality
floor via `metaflac --show-sample-rate --show-bps` (`MetaflacPath` is the
test seam), returning `ReasonBelowCDQuality`; the two CD-quality bats cases
(`tests/lib.bats:188`, `:209` at HEAD) port as-is. `gate` imports `reject`
only.

### T7: Review filing and lifecycle

**Files:** `internal/review/review.go`, `internal/review/move.go`, tests.

**Produces:**

```go
type Filer struct { ReviewDir, QuarantineDir string; Now func() time.Time }
func (f *Filer) MoveToReview(path string) (string, error)  // collision-safe name; quarantine fallback
func (f *Filer) Sweep(retention time.Duration) error       // review AND quarantine (fix-wave behavior)
func Move(src, dst string) error                           // EXDEV-aware; preserves mode; chown 2000:2000 best-effort
```

**Criteria:** `Now` is the clock seam the frozen-clock bats cases need;
collision naming matches `unique_dest_path` at HEAD; the quarantine
second-chance fallback and its could-not-move summary line are ported
(the previously untested "COULD NOT MOVE BY ANY MEANS" branch gets its first
real test here via an unwritable-destination fixture); `Move` falls back to
copy-and-remove on `EXDEV`, preserving the 2000:2775 setgid tree semantics,
proven by a test that checks mode bits on the result.

### T8: Beets wrapper and scan trigger

**Files:** `internal/beets/beets.go`, tests.

**Produces:**

```go
type Runner struct { BeetPath, ConfigPath, StateDir string }   // BeetPath is the test seam; no interface
type ImportResult struct { Skipped []string; LogPath string }
func (r *Runner) Import(ctx context.Context, dir string) (*ImportResult, error)
func TriggerScan(ctx context.Context, navidromeURL, user, password string) error
```

**Criteria:** Import runs `beet import -q -l <tag_log>` under
`exec.CommandContext`, writes the per-run tag log under StateDir (atomicfile),
re-reads it, and parses `skip <path>[; <path>]` lines exactly as
`move_skips_and_notify` does (ported bats cases); TriggerScan builds the
Subsonic `startScan` URL with fresh salt and md5 token in the request URL —
never argv, no temp file, the `--config` dance and its RETURN-trap cleanup
die — pinned by an `httptest` assertion that the URL carries `u/t/s/v/c`
and the process argv never contains the password (re-expressed case). Scan
failure returns an error the caller treats as best-effort.

### T9: Importer orchestration and e2e harness

**Files:** `internal/importer/importer.go`, tests; `cmd/dubplate/import.go`;
`e2e/main_test.go`, `e2e/import_test.go`, `e2e/testdata/`.

**Interfaces — consumes everything above:** config.Load, lock.Acquire,
ping.Client, notify.Mailer, reject, archive, gate, review.Filer,
beets.Runner.

**Criteria:** the run flow matches `music-import` at HEAD stage-for-stage
(quiescence check, empty-inbox clean skip, run-level free-space abort,
per-contributor staging, gate, quarantine filing, reject mail, beet import,
skip handling, the two admin summary emails — "import skipped items" and
"import quality gate failures" with the per-album raw-reason block and any
could-not-move lines — scan trigger, empty staging-dir cleanup, sweeps,
success ping). The error doctrine is enforced here per the spec: per-run
failures return errors (nonzero exit via main) and rely on `OnFailure=` for
the failure ping — no in-process fail ping, so the free-space bats case
re-expresses as abort-plus-nonzero-exit, not a ping assertion; per-item
failures accumulate into a run report struct logged at end and never abort;
lock contention pings success and returns nil (exit 0) — pinned by a ported
bats case; ping/notify/scan failures are logged and swallowed at these call
sites only. The run report is also written as versioned JSON to the state
dir (atomicfile), on every exit path including held-lock (a report saying
so) — the uploader pass builds its journey truth on this artifact (spec:
2026-08-31-dubplate-uploader-design.md). It carries per-album provenance:
uploader, source inbox directory, outcome (imported / review / quarantined
/ rejected), and for imported albums the library paths resolved by querying
the beets database after the run; the tag log records only skips and beets
renames every file, so the database is the only success record. The beets
invocation gains `--set collection=<c>` (the batch's collection, default
`Collections[0]`; per-batch overrides arrive with the uploader pass's
session metadata), and `config/beets-config.yaml` gains query-prefixed path
formats routing non-default collections into `<collection>/` subtrees while
the default collection keeps today's exact layout — zero migration of the
existing library (ratified by Geoff 2026-08-31; Navidrome ≥0.58 maps each
subtree to a library with per-user visibility, estate-side admin config,
not code). Both admin summary subjects and body shapes are asserted in e2e
against the test SMTP server. The e2e harness
builds the real binary and stub `beet`/`rclone`/`sqlite3` Go programs in
`TestMain`, runs against the fixture inbox, and asserts filed outcomes,
mail content (test SMTP server), and ping method/URL. The four
unknown-uploader paths (fix-wave M10 tests) port as e2e cases.

### T10: Backup, beet passthrough, ping-fail

**Files:** `internal/backup/backup.go`, `cmd/dubplate/backup.go`,
`cmd/dubplate/beet.go`, `cmd/dubplate/pingfail.go`, tests and e2e cases.

**Criteria:** backup ports `music-backup` at fix-wave HEAD: LIBRARY_DIR
existence and non-empty guards, its own lock, sqlite3 `.backup` snapshot
(subprocess), rclone sync (subprocess) with the `_archive/<tree>/<date>`
layout untouched. Abort paths return an error and rely on `OnFailure=`; no
in-process fail ping (spec doctrine — the shell's `ping_hc backup fail`
call does not port). `dubplate beet -- <args>` acquires the import lock
blocking, then `syscall.Exec`s the real beet with argv, tty, and
environment intact (T3's CLOEXEC-cleared fd keeps the lock held; e2e proves
an overlapping `dubplate import` skips while beet runs). `dubplate ping-fail <slug>` is the `OnFailure=` chain's entry point and must never be
silenced by unrelated config defects: it resolves `DUBPLATE_PING_<SLUG>_URL`
directly from the environment without a full `config.Load`; an unset URL is
a logged no-op exiting 0 (monitoring never takes down what it monitors, and
the no-Worker adopter case is valid); only a malformed URL exits nonzero.

### T10b: Restore

**Files:** `internal/restore/restore.go`, `cmd/dubplate/restore.go`, tests and
an e2e case with the stubbed rclone.

**Produces:** `dubplate restore --to <dir>` — pulls `library/` and the
`state/` database snapshots from the R2 bucket into a target tree and
verifies what arrived.

**Criteria:** strictly read-only toward the bucket (`rclone copy`, never a
sync that could delete remote objects); refuses a non-empty target without
`--force`; verifies both restored databases via `sqlite3 "PRAGMA
integrity_check"` (subprocess, like backup); restores `state/` and
`library/` only — the dated `_archive/` trees are listed in the report but
not pulled by default; the run report enumerates restored file counts and
any bucket-only differences, doctor-style. Works from any machine with
rclone and the env vars — no box access required. This is the executable
half of the charter's "runbooks executed skeptically": the DR drill at pass
end is running this against a temp dir and playing the result in a local
Navidrome. It is also the adopter's migration path — importing an existing
collection into a fresh estate is a restore. Approved as added scope by
Geoff 2026-08-31 (one task, before launch).

### T11: Checks and doctor

**Files:** `internal/checks/checks.go`, `cmd/dubplate/check.go`,
`cmd/dubplate/doctor.go`, tests.

**Criteria:** `dubplate check <disk|navidrome>` (ExactArgs(1); bare `check`
errors — a typoed ExecStart must not exit 0): disk compares the filesystem
holding LibraryDir against Thresholds.DiskPct and pings success or fail;
navidrome GETs `/rest/ping` and pings accordingly — both port the fix wave's
shell checks and their bats cases. `dubplate doctor`: prints every config
field and its resolution state (using the joined ErrMissing list), probes
SMTP (connect/EHLO/QUIT, no send), R2 (`rclone lsjson --max-depth 1`,
read-only), Navidrome (GET `/rest/ping`); ping URLs are validated
syntactically only and never fetched (a GET would record a false success
ping — state this in the doctor output); exits nonzero on any miss. Doctor's
output replaces bring-up runbook step 1 (runbook edit lands in T13).

### T12: Ship and first flips (box task)

**Files:** `scripts/deploy.sh`, `env/musicbox.env.template` (if new secret
names emerged), `~/.dotfiles/secrets/registry.md` (same commit discipline as
the secrets flow), `systemd/` unit ExecStart edits for check and backup
units.

**Box steps, run with the box quiet:** deploy the binary; `dubplate doctor`
green on the box; supervised live `musicbox check disk`, `check navidrome`,
and `musicbox backup` runs; flip those units' ExecStart. The shell scripts
stay in place and in `SCRIPT_FILES` until T13.

**Criteria:** deploy.sh gains the build-and-ship step (cross-compile, rsync,
install 0755) and every manifest edit lands in the same commit as its flip.
The env migration window is **additive and carries the estate rename**: the
box env and template hold both the old `MUSICBOX_*` names (the still-live
shell scripts resolve `MUSICBOX_HC_*` via `ping_hc`, whose unset-var
behavior is a silent no-op that would blind monitoring without error) and
the new `DUBPLATE_*` names the binary reads, until T13 closes the window.
Units flip under their new names: each flip ships a `dubplate-*.service`/
`.timer` replacing its `music-*`/`musicbox-*` unit, with the old unit
disabled in the same step. The new required non-secret names (SMTP
host/port/from, Navidrome admin user) are minted into the age store and
registry exactly like the secret ones, the existing delivery path onto the
box (the config-file overlay stays the packaging initiative's improvement);
rollback is documented in the task report as ExecStart flip-back (old units
and scripts still present). One full timer cycle must elapse green (backup
and both checks) before T13 starts — the conductor verifies via
journal/pings evidence, not by waiting idle.

### T13: Final flips and closeout (box task)

**Files:** `systemd/music-import.service`, `musicbox-ping-failure@.service`
ExecStart; delete `scripts/music-import`, `scripts/music-backup`,
`scripts/music-beet`, `scripts/musicbox-ping-failure`, `scripts/lib.sh`,
and the fix wave's shell check scripts T11 ported
(`scripts/musicbox-disk-check`, `scripts/musicbox-navidrome-check`);
`scripts/deploy.sh` manifests and the msmtprc install stanza;
`config/msmtprc` (deleted, notify parity having passed);
`scripts/check.sh` — its bats and shellcheck stanzas are removed in the
same commit as the test deletions, since `bats tests/` errors on an empty
directory and would turn the gate red; `cloud-init/user-data.yaml`;
`docs/bringup-runbook.md`; `docs/STATUS.md` pinned-versions table; the
bats files that tested deleted scripts (removed with them; the Go suites are
the coverage now).

**Box steps:** supervised live `dubplate import` with real content; flip
import and ping-failure units; swap `/usr/local/bin/music-beet` for the
passthrough; after one green import timer cycle, delete the scripts from box
and repo.

**Criteria:** notify parity test has passed before msmtp leaves; cloud-init
drops msmtp, yq, jq but keeps metaflac's package alongside flac, ffmpeg,
rclone, sqlite3 (the binary shells to them); every `MUSICBOX_*` name
leaves the env, template, and age store in the same commit that deletes
`lib.sh` (closing T12's additive rename window); runbook's package gate and
STATUS's pin table updated in the same commit, so a mid-migration fresh
provision cannot build a box the binary can't run on; `bash
scripts/check.sh` is green after every deletion; rollback from here is
redeploy-from-git and the runbook says so; final state has zero references
to deleted scripts anywhere in the repo (`grep` proves it, and the grep
scope includes the T11-ported shell checks).

---

### T14: Estate rename closeout (box + estate task)

**Files:** the repo itself (rename `glw907/musicbox` → `glw907/dubplate` on
GitHub if a remote exists, and `~/Projects/musicbox` → `~/Projects/dubplate`
locally); `README.md`, `docs/STATUS.md`, `docs/HISTORY.md` (header note
only — history entries stay as written), `docs/bringup-runbook.md`,
`ROADMAP.md`; the pings repo's check registrations; the workstation memory
note (`music-library-setup`).

**Box steps:** move `/etc/musicbox` → `/etc/dubplate` and update every unit's
`EnvironmentFile=` in one step with a deploy; re-register the pings checks
under `dubplate-*` slugs via `add-check` and retire the `musicbox-*` slugs
via `remove-check` inside one grace window so no false alert fires; rename
`/opt/musicbox` and `/srv` paths only if the config actually references
them — `/srv/music` stays (it names the content, not the project).

**Criteria:** the spec and plan chain migrates into the repo
(`docs/superpowers/specs/` and `plans/`, following the cairn-family
convention) — the dubplate rewrite and uploader designs move from
`~/.dotfiles/docs/superpowers/` with dotfiles-side pointers left behind, so
the repo owns its documents as an adopter-facing project should; after this
task, `grep -ri musicbox` across the repo returns
only HISTORY entries and dated archive docs (the historical record keeps its
name; everything operative says dubplate); the artifact diagram, STATUS, and
the spec chain pointers are updated; cloud-init provisions a fresh box that
never carries the old name. This task runs last, after T13's window closes,
and — like every task in this pass — only starts once no other executor
holds the repo.

## Review provenance

Adversarially reviewed at authoring (opus agent, ten ranked defects plus
four unranked, all folded 2026-08-30): the complete config-field inventory
rule, the CD-quality gate restoration, the T13 gate-integrity fix, the
CLOEXEC lock fix, the additive env rename, ping-fail's config-independent
resolution, the OnFailure single-ping consistency in T9/T10, the
three-placeholder RenderEmail signature, T9's completed stage list, and the
NewClient defaults.

## Self-review notes

Spec coverage: every spec section maps — Shape→T1/T9–T11, Packages→T2–T8,
doctrine→T9 criteria, Testing→per-task mapping rules plus T9 harness,
Migration→T12/T13, portability→T2 required-fields rule. Type consistency
checked across tasks (Roster, Reason, Limits, Policy, Filer, Runner names
match everywhere they appear). No placeholders remain.
