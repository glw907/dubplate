# Plan review: contract and criteria lens

Target: `docs/superpowers/plans/2026-09-25-translit-table-at-setup.md` at
`b841954` (revision 1). Lens: is every task's promise testable, does each
acceptance criterion name its fixture and why it would fail, and can any
criterion pass vacuously. The four planning-miss items (quoted behavior,
states and reports, privileged-field channels, fixture states) were checked
per task. Line numbers are the plan's at `b841954`.

**Counts: 0 blockers, 4 major, 9 minor, 0 owner forks.**

**Verdict.** The plan is sound on its load-bearing proof. The byte-identity
criterion (Task 1, `:716-720`) cannot pass vacuously when it hashes in the
test and compares with the literal `c954...` the plan quotes: that literal
is Python's output, independent of the Go port, and Task 4 re-derives it
from `generate.py` after the `.tbl` is gone (`:957-959`). The named
mutations mostly fail as claimed. The four majors are fixture and seam gaps
that an implementer meets mid-task: the doctor tests' green fixture, the
download-client seam, the unreachable post-checksum refusal rows, and the
box sandbox for the deploy step.

---

## Major

### M1. The new failing doctor row breaks every existing exit-0 doctor test, and the plan names no table fixture for them

Location: plan `:863-871`, `:886`, `:894-895`; Task 0 note 8 at `:646`.

Defect. The row fails the command's exit on `missing` whenever
`DUBPLATE_STATE_DIR` is set. `TestDoctorCmd`'s "full valid wall" case
(`cmd/dubplate-server/doctor_test.go:45-48`, `wantErr: false`) sets
`DUBPLATE_STATE_DIR` to an empty `t.TempDir()` (`:76`). After Task 3 that
case exits nonzero, and so does every other test that expects exit 0 with a
temp state directory (`:415`, `:915`, `:926`, `:937`, `:992`, `:1045`,
`:1077`, and the `wantErr: false` row at `:1112`). The plan's only mention is
"the row count the existing tests pin moves by one", and Task 0 note 8 asks
only about row-count pins. The converse matters for the proof: the new
"missing fails the exit" row and mutation 3 (`:906`) prove something only on
a fixture where every other row passes. The exit rule
(`doctor.go`, `loadErr != nil || rosterFailed || storeFailed ||
backupRemoteFailed`) is an OR, so on a fixture with any other failing row the
exit assertion passes whether or not the new row fails.

Fold. Add to Task 0 note 8: "report every doctor test that expects exit 0
with `DUBPLATE_STATE_DIR` set." Add to Task 3's outcomes: every such test
installs a table through the install verb with `--from` the cached wheel,
the note-7 method. Add to the doctor row's state table: "the `missing` and
`invalid` exit rows run on `baseEnviron()`'s green wall, with every other row
`ok`, so the exit moves on this row alone."

### M2. `Fetch` takes a const URL, and no seam lets the install verb's or the harness verb's tests reach `httptest`

Location: plan `:466-468`, `:684-687`, `:712`, `:880`, `:890`, `:908`.

Defect. `Fetch(ctx, *http.Client)` GETs the exported const `WheelURL`, a
real `files.pythonhosted.org` URL. An `httptest.Server` listens on
`127.0.0.1`, so a test reaches it only through a client whose transport
rewrites the request's host. The plan never names that transport, and never
names how a test hands a client to `newRootCmd`'s install verb or to
`harness fetch-unidecode`. Three proofs depend on it: the "no request"
idempotence row (`:880`), mutation 5 (`:908`), and the harness no-op row
(`:712`). If the test client falls through to the default transport, the
idempotence test under mutation 5 opens a real connection, which breaks the
hermetic gate (`:371-375`) instead of failing the count. The injected client
is also a privileged field with no row in the channels table (`:409-414`).

Fold. Name the seam. The unidecode tests use an `http.Client` whose
`Transport` is a `RoundTripper` that asserts the request URL equals
`WheelURL`, counts it, and forwards it to the `httptest` server, and refuses
every other host. The cmd install verb's constructor takes the `*http.Client`
the way `newDoctorCmd` takes `environ`, with `main` passing the bounded
production client. `harness fetch-unidecode` does the same. Add a channels
row: "the download client: `main`'s constructor argument; tests pass the
rewriting transport". Add a criterion: under every mutation, the rewriting
transport refuses any host other than the test server.

### M3. Build's zip, entry-count, and lexer refusal rows are unreachable through `Build`, and the plan names no stage a test can drive

Location: plan `:483-495`, `:700-705`, `:692`.

Defect. `Build` verifies the wheel's sha256 first, and only one byte string
passes. So with any input a test can build, the rows "not a zip, or an entry
fails CRC or size", "fewer or more than 190", and "a token outside the
measured grammar" (`:701-703`) are unreachable through `Build`. `:692` covers
the lexer ("lexer rows over synthetic sources"), but the zip and entry-count
rows have no stated fixture. An implementer either writes no test for them,
so the refusals ship unproven, or exports a test seam, which trips the
exported-unused leg. The same holds for `Parse`: the plan admits the
decoder's errors cannot fire behind the pin (`:706`), so the moved decoder's
truncation and trailing-byte checks become code no test reaches.

Fold. Name an unexported stage, for example `extract(zr *zip.Reader)`
returning the blocks, which `Build` calls after the checksum. Its test rows
build synthetic zips in memory with `archive/zip.Writer`. For the CRC row,
quote `go doc archive/zip.Writer.CreateRaw`, the call that writes a header
with a caller-chosen `CRC32`. Say the same for the decoder: an unexported
`decode` keeps the truncation and trailing-byte rows, tested directly,
because `Parse`'s pin makes them unreachable through `Parse`. Also state
that the 190 count is a property of the pinned wheel. The R-tt7 bump
procedure (`:571-584`) must update the count beside the two pins, or the
count moves into a derived check. As written, a bump changes the count and
the procedure does not mention it.

### M4. The deploy step's box sandbox has no named state or proof, and the unit-property copy the plan suggests cannot write the leaf

Location: plan `:934-945`, `:950-959`, `:964-968`, `:1120-1123`.

Defect. Task 4 suggests a `systemd-run` "carrying `User=music` and
`EnvironmentFile=/etc/dubplate/server.env`", following the shadow unit's
operator rule. The shadow unit runs under `ProtectSystem=strict` with
`ReadWritePaths=/var/lib/dubplate/shadow` only. An implementer who carries
the unit's properties whole gets a verb that cannot create
`/var/lib/dubplate/unidecode` (read-only filesystem), so the next deploy
aborts. `deploy.sh:495-500` states the matching trap: a `ReadWritePaths=`
entry must exist on the host before the unit starts. The plan's other route,
`runuser -u music`, cannot read `server.env`, which is `0600 root:root`
(`:217-218`). The task's criteria never run the step (`--check` is excluded
at `:953`), and both mutations are reviewer reads. So the first proof is
5d's production deploy, and a failure there aborts it.

Fold. Pin the shape instead of leaving it to the implementer. `deploy.sh`
pre-creates `/var/lib/dubplate/unidecode` as `music:music 0755` in the
existing `/var/lib/dubplate` block (`deploy.sh:489-504`). The verb then runs
under `systemd-run --wait --pipe` with `User=music`, the `EnvironmentFile=`,
and `ReadWritePaths=/var/lib/dubplate/unidecode`. That also makes the verb's
own leaf creation a no-op on the box. Add a criterion that the diff-reviewer
checks the property list against the leaf path. Add a note to Task 5's
STATUS draft: 5d's runbook runs the step once by hand under `--check` before
the timers.

---

## Minor

### m1. The final exit-gate grep for `go:embed` finds the store's migrations and can never print nothing

Location: plan `:1014-1015`. `git grep -n 'go:embed' -- internal/engine`
matches `internal/engine/store/migrations.go:13` (`//go:embed
migrations/*.sql`). Task 2's criterion (`:813`) is correctly scoped to
`internal/engine/matcher`. Fold: scope exit-gate check 2 to
`internal/engine/matcher`.

### m2. Exit-gate check 3 has no positive control

Location: plan `:1016-1018`. `grep -c 'Kayng '` printing 0 proves absence
only if the base binary printed a positive count. `Kayng ` is present in the
committed table today (one match in `unidecode-1.4.0.tbl`). Fold: run the
same line against a `make build-linux` binary at the base commit, record its
nonzero count, then run it at HEAD.

### m3. The moved sweep test is a tautology

Location: plan `:721-722`. `Build` writes the digest it computed through
`Lookup`, and the test recomputes it through the same `Lookup`. The test
passes under any `Lookup` defect. Today's `TestUnidecodeTableFullSweep`
compares Go's lookup with Python's digest, a real proof, and only the
byte-identity test now carries it. Fold: drop the sweep test, or state it as
a determinism check only. Record that byte identity is the proof that
replaces it.

### m4. The byte-identity test's independence depends on where it takes its expected value

Location: plan `:716-720`, `:475`. If the test only asserts `err == nil`
from `Build`, or compares with the package's unexported pin identifier, a
pin edited to match a wrong Go build would pass. Fold: the test computes
`sha256.Sum256` itself and compares with the literal written in the test
file. The diff-reviewer confirms that the source pin equals the plan's
literal `c954605b...ce85`.

### m5. Mutation 6 fails only if the over-cap row asserts the size refusal specifically

Location: plan `:708`, `:746`. With no cap, or with `io.LimitReader(cap)`
and no plus-one read, a body of cap+1 bytes still fails, on the checksum
(`ErrWheelChecksum`). A row that asserts only "refuses" survives the
mutation. Fold: the row asserts the size error, through a sentinel or the
cap in the message, and asserts `!errors.Is(err, ErrWheelChecksum)`. Add the
boundary row: a body of exactly the cap is not refused for size.

### m6. Go's test cache hides the TestMain "absent" state, and "every gate run re-proves" is overstated

Location: plan `:373-375`, `:528-537`, `:799`. `make test` is `go test ./...`
with no `-count=1` (`Makefile:17-18`). `cmd/go` validates a cached result by
the files and environment variables the test binary logs, and `testing` starts
that log in `(*M).before()`, which `m.Run` calls
(`$GOROOT/src/testing/testing.go:2591`, `:2814-2833`). A `TestMain` that
reads the wheel and `XDG_CACHE_HOME` before `m.Run` therefore logs neither.
A deleted cache then still reports a cached `ok` for the four packages, and
a second worktree with identical sources shares that result. The pinned wheel
keeps the result valid, so the only masked state is "absent". Fold: Task 2's
mutation 2 and absent-state rows run with `-count=1`. Reword R-tt5's "every
gate run re-proves" to "every uncached run". Optionally, the `TestMain`
re-stats the cache path inside a test so the cache tracks it.

### m7. Install verb state gaps: leaf shape and owner, and check order

Location: plan `:553-566`, `:877-884`. The ownership refusal checks the state
directory only. It names no report for a `unidecode/` leaf that exists as a
file, at the wrong mode, or owned by another user (a root-owned leaf from a
bare root run under a music-owned state directory). The store leaf's
`checkLeafShape` and `checkOwner` are the precedent (`leaf.go:104-129`). The
plan also does not order the idempotent no-op against the owner refusal, so
a test cannot pin a root run on a valid table. Fold: add leaf rows (a file,
wrong mode, other owner: refuse naming the path and the fix). State that the
refusals run before the "already installed" check. Separately, the
`--from absent` row reads "naming the path and the sha256 mismatch", but an
absent file has no sum. Split it: absent or a directory names the path, and
a wrong file names both sums.

### m8. The doctor row's state boundaries and a side-effect channel

Location: plan `:411`, `:849-851`, `:863-871`. The row states do not say
which read errors count as `missing` (only `fs.ErrNotExist`, including an
absent state directory, presumably) and which as `invalid` (a directory,
permission denied, a `Parse` refusal). The row calls the helper that "loads
and installs", so `doctor` also sets the process-wide transliterator, a
channel absent from the privileged-fields row (`:411`). Fold: name the
error-to-state mapping. Split the helper into a load that the doctor row and
the verbs call and an install that only the scoring verbs call, or list
doctor in the channels row.

### m9. Smaller quoted-behavior and fixture gaps

- `:709`: "returns the context's error". `go doc net/http.Client.Do` reads
  "Any returned error will be of type *url.Error", so the criterion is
  `errors.Is(err, context.Canceled)`. Quote it.
- `:802`, `:426-427`: "race-free holder" has no quoted primitive. Quote
  `go doc sync/atomic.Pointer`. The uninstalled-panic test (`:791-793`)
  clears the holder through package-internal access, since
  `translit_test.go` is `package matcher`. Never clear it by passing a nil
  function to the exported setter, which would add an uninstall channel.
- `:727-728`: "the first fetches" needs an empty cache as its fixture. A
  fix-round re-dispatch finds the cache filled. Run the first call with
  `XDG_CACHE_HOME` pointed at a fresh scratch directory. Also `:401-403`
  says Task 2 is the first task whose gate needs the cache, but Task 1's own
  byte-identity test reads it. Reword to "Task 1 fills it; the conductor
  confirms it before Task 2".
- `:712`, `:799-800`: add the states "cache file present but failing its
  checksum" (the harness verb refetches and replaces it) and
  `os.UserCacheDir` returning an error (a `TestMain` prints it and exits
  nonzero).
- `:337`, `:528`: after Task 3, `cmd/dubplate-server`'s test binary also
  reads the cached wheel, through the install verb with `--from`, so the
  "fail naming `make unidecode-wheel`" rule should name that package as well.
