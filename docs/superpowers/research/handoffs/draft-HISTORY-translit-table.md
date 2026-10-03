## Transliteration table leaves the binary (the Unidecode table built at setup): closed 2026-10-03, lane 1

Executed against `docs/superpowers/plans/2026-09-25-translit-table-at-setup.md`
(approved 2026-09-25, revision 2.1, both rulings yes), in the main tree, from
base `1c012a8`. Task 0 read the tree at `4f4930b` and said launch
(`62ad346`, `research/2026-09-26-translit-task-0-report.md`). The pass ran in
the overnight window of 2026-09-26, stopped when the laptop closed on
2026-09-27 with Task 3c pending, and resumed in a fresh session on
2026-10-03.

After this pass no `dubplate-server` binary carries Unidecode data. A binary
scores only after it has loaded a table whose sha256 equals the one pin it
carries, built from the one wheel whose sha256 it pins. The pass changed no
scoring behavior, corpus record, manifest field, or Tunables value, and
added no module dependency.

### What landed

Task 1 (`5318671`, fix round `1ca79d7`, narrow follow-up Task 1b `194c01c`)
added package `internal/engine/matcher/unidecode`: the table format and its
decoder and encoder, the three pins (`WheelSHA256`, the unexported
`tableSHA256`, and the data-module count 190), a strict lexer for the wheel's
data modules, `extract` (zip read, entry selection, lexing), `Build`, `Parse`,
`Lookup`, `Fetch`, and `WheelCachePath`. It also added `harness
fetch-unidecode` and the `make unidecode-wheel` target that fills the
workstation's cached wheel. The fix round read every zip entry to EOF,
refused a near-miss data-module name and a duplicate section, and counted
matching entries rather than map keys. Task 1b was ruled narrow after a
second fix round (`36cb7f1`): it proved that `extract` drains a corrupted
non-module entry and surfaces the zip reader's own error, and it widened the
name gate so `unidecode/x000.pyc` is refused by name.

Task 2 (`e8e158d`) replaced `matcher`'s embedded table and decoder with an
exported `SetTransliterator` over an unexported `atomic.Pointer`, deleted
`internal/engine/matcher/unidecode-1.4.0.tbl`, and gave each of the four test
binaries that reach the table a `TestMain` that builds and installs it from
the cached wheel. `conform.CheckUnidecodeVersion` now compares the corpus
manifest with `unidecode.Version`.

Task 3 (`850cb4c`, fix round `189a6c8`) added `dubplate-server unidecode
install [--from FILE]`, the table load in `shadow observe` and `shadow
reevaluate` before `shadow.OpenStore`, and a fourth, unconditional `doctor`
row named `unidecode`. The fix round replaced every bare `&http.Client{}` in
the cmd tests with a client whose transport fails the test on any request, so
a regression that slips past a refusal fails on a stray request rather than
reaching the network.

Task 4 (`2ef4000`, review fix `e8fadec`) staged the deploy swap and retired
the generator's shipped-table role. `scripts/deploy.sh` rsyncs the new binary
to `/usr/local/bin/dubplate-server.new`, smoke-tests it there, runs its
`unidecode install` under a `systemd-run` sandbox derived from
`dubplate-shadow.service`'s hardening, and only then `mv`s it onto the live
path. The `/var/lib/dubplate` pre-create moved ahead of the binary step.
`tools/unidecode-table/generate.py` became the pin's oracle (R-tt7) and now
prints its output's sha256.

The simplifier round (`aeea780`, accepted, reviewed) folded the four
`TestMain` bodies into a shorter shape, moved the wheel-cache read into a
shared helper, and tightened `extract`, `lex`, and `table`; no behavior
change.

Task 3c (`5530cbb`) folded the three architecture reads
(`research/2026-09-27-translit-architecture-reads.md`; rulings recorded at
`6d5c198`). Its outcomes:

- A `\xHH` escape of 0x80 or above now appends the code point with
  `utf8.AppendRune`. The old code appended a raw byte, where Python 3 yields
  U+0080. The defect was latent: no pinned wheel entry carries such an
  escape, so the table pin masked it.
- `Lookup` compares `uint32(r)` against both bounds. A negative rune passed
  `r < 0x80` and could index section 1.
- `extract` checks an entry's name before opening it, so an open or CRC
  failure can no longer mask a name refusal.
- `ReadWheel(io.Reader)` is the one bounded read followed by the checksum;
  `Fetch`, the `--from` path, and `harness`'s cache check all call it.
  `MaxWheelSize`, `ErrWheelTooLarge`, `ErrWheelChecksum`, and
  `ErrTableChecksum` are unexported.
- `LoadCachedTable(tb testing.TB)` moved to the new `unidecodetest`
  subpackage, the `net/http/httptest` shape.
- The install verb wraps its errors once at the `RunE` entry; the reads had
  counted 17 hand-prefixed sites. The download client is a package-level
  variable swapped under `t.Cleanup`, so `newRootCmd` takes no parameters.
  `installUnidecodeTransliterator` became `setUnidecodeTransliterator`.
- The matcher's comment no longer cites `time/tzdata` as the precedent (see
  below).

Net change from the base to `5530cbb` over the union of the pass commits'
non-docs files: 36 files changed, 3553 insertions(+), 355 deletions(-).
The command and its output:

```
git diff --shortstat 1c012a8 5530cbb -- $(for c in 5318671 1ca79d7 194c01c e8e158d 850cb4c 189a6c8 2ef4000 e8fadec aeea780 5530cbb; do git show --name-only --format= $c; done | sort -u | grep -v '^docs/')
 36 files changed, 3553 insertions(+), 355 deletions(-)
```

### Resolutions: where each landed

| Resolution | Landed in |
| --- | --- |
| R-tt1 process-wide seam, installed by the main package | `e8e158d` (seam, four `TestMain` files); `850cb4c` (the one non-test `SetTransliterator` call, in `cmd/dubplate-server/unidecode.go`); `5530cbb` (comment corrected) |
| R-tt2 package owns the table | `5318671`; export surface trimmed by `5530cbb` |
| R-tt3 table built from the wheel only, strict lexer | `5318671`, hardened by `1ca79d7` and `194c01c` |
| R-tt4 absent means refuse | `850cb4c` (load before `shadow.OpenStore`); `189a6c8` |
| R-tt5 tests read a checksum-verified cached wheel | `5318671` (fetch verb, Makefile target); `e8e158d` (`TestMain` files); `5530cbb` (`unidecodetest`) |
| R-tt6 leaf, modes, install verb contract | `850cb4c`; `189a6c8` |
| R-tt7 `generate.py` becomes the pin's oracle | `2ef4000`; `e8fadec` |

Rulings: the 3.3M ceiling (flag 2.64M) and the verb name `dubplate-server
unidecode install [--from FILE]` were both ruled yes (Geoff, 2026-09-25). The
second Task 1 fix round was narrowed to Task 1b by the conductor
(`36cb7f1`).

### Gate evidence

The exit gate, `cairn-run-gate 'bash scripts/check.sh'`, ran on master at
`5530cbb` and exited 0 in 8m15s on the heavy lane (Chromium included). Staged
legs warned: 323 process citations in Go comments (STATUS carried 325 before
this pass, so the pass retired two) and 129 exported identifiers under
`internal/` with no non-test consumer (STATUS carried 128). The added one is
`unidecodetest.LoadCachedTable`, which is a test helper by design and so is
expected to stay on the list.

Uncached package runs at `5530cbb` (`go test -count=1`), wall time on this
machine: `unidecode` and `unidecodetest` 1.82s together, `matcher` 4.88s,
`cmd/dubplate-server` 68.99s, `cmd/harness` 1.20s, `tag` 0.92s, `shadow`
1.37s, `conform` 159.04s.

The implementer's first gate run in Task 3c hit a disk-quota error writing to
`/tmp`; the identical rerun passed. A gate that fails on a temp-directory
write is a lost run, never a red gate: free the space and re-issue.

### Spend

About 3.4M through the simplifier round (`aeea780`) against the 3.3M ceiling
and the 2.64M flag. STATUS at the segment checkpoint (`139683e`) read about
2.1M, so segment B (Task 4, its review fix, the simplifier round) cost about
1.3M; the plan projected 0.60M for Task 4, the fix round, and the
simplifier round together. The flag was crossed
during segment B and the pass continued under the 2026-09-26 overnight
grant. Task 3c's run added about 0.26M in subagent tokens, for about 3.66M in
all. The per-task split lives in the workflow journals (segment A ran as
`wf_581a3dd5-5c5` and `wf_45c82967-ff9`; the first Task 3c launch
`wf_ca338e5b-9ed` stopped with no edit; the second, `wf_560e7982-63a`,
produced `5530cbb`). Attended interactions after plan approval: zero, plus
one pause (laptop closed 2026-09-27) and a fresh-session resume on
2026-10-03.

### Architecture reads

Three `go-architecture-reader` grades ran at `aeea780` with no plan in
context. `cmd/dubplate-server` read workmanlike, `internal/engine/matcher/unidecode`
read sound with nits, and `internal/engine/matcher` read workmanlike. Task 3c
took every FOLD item. The ROUTE items, which go to the architecture-cleanup
list, live in `research/2026-09-27-translit-architecture-reads.md`. Two of
them are this pass's own: the owner, leaf-shape, and uid-name checks now
exist in four copies with two refusal wordings, and the `TestMain` shim
(`mainTB`) is copied into four files.

### What a later pass would be wrong to rediscover

- **The measured grammar.** The wheel holds 200 entries, all Deflate, of
  which 190 match `unidecode/x[0-9a-f]{3}.py`. Python's `tokenize` over those
  190 found 43,120 strings, 5,477 `None` names, and 48,486 comments. `'#'` is
  a string entry in `x006.py`, `x025.py`, and `x028.py`, so a comment ends a
  line only outside a string. `x000.py` packs sixteen entries per line,
  three modules carry comment lines holding quote characters, and
  `x020.py:93` is the one double-quoted string. A lexer written per line or
  per entry misses all of these. The lexer stays strict on purpose: a pinned
  wheel cannot change, so a token outside the measured set is a lexer defect.
- **Byte identity, proved twice before it was shipped.** A Python prototype
  and a stdlib-only Go prototype (the mechanics review) each built a
  255,458-byte table whose sha256 is
  `c954605b805e4a21d1d9e79b4d93bc6b89684de9996dd21720b256fc58a6ce85`, equal
  to the table committed at `1c012a8` and now deleted. The shipped
  `TestBuildIsByteIdenticalToThePin` hashes the build against that literal,
  independent of the package's own pin constant, so a pin edited to match a
  wrong build still fails.
- **Why the digest in a Go-built table still proves Python's lookup
  (R-tt3).** The pin is the sha256 of the file `generate.py` wrote, whose
  digest field Python computed by calling Unidecode's own `_get_repl_str`
  for every code point. A Go build that matches the pin byte for byte
  therefore reproduced Python's digest, and so Python's lookup, everywhere.
- **The pin-update procedure (R-tt7).** On a Unidecode bump, bump the venv,
  run `generate.py` to a scratch path (never commit its output), and update
  four values in `internal/engine/matcher/unidecode/pins.go` by hand:
  `Version`, `WheelSHA256` from PyPI's release metadata, `tableSHA256` from
  the printed digest, and `dataModuleCount` from the run's own block count.
  The Go build of the new wheel must then match every pin. Python stays in
  the loop only because it is independent of the Go port.
- **Four test binaries reach the table, and the overlay probe found them.**
  `go test -overlay` replacing `translit.go` with a copy whose `Transliterate`
  panics on entry failed exactly `matcher`, `tag`, `shadow`, and `conform`;
  `musicbrainz` and `cmd/dubplate-server` passed (Task 0 re-ran it at
  `4f4930b` and got the same four). A package added later whose test binary
  links `matcher` needs its own `TestMain`, and the probe is how to find it.
- **Go's test cache does not track the wheel.** The cached wheel sits under
  the user cache directory, outside the module, GOPATH, and GOROOT roots, so
  a cached `ok` survives a deleted or replaced wheel. Every proof of "wheel
  absent" runs with `-count=1`. The gate is hermetic by design: it fetches
  nothing, and a missing wheel fails the tests naming `make unidecode-wheel`
  rather than skipping.
- **`make unidecode-wheel` is a workstation prerequisite**, the same shape as
  `npm ci`. It is a no-op with no network when the cached file verifies. The
  cache is the user cache, not a repo directory, so a second worktree shares
  one download.
- **The deploy transition is staged because the shadow timer is live.**
  `dubplate-shadow.timer` runs every five minutes on the box and `deploy.sh`
  never stops it, so a swap-then-install would leave a window in which a
  timer firing finds a new binary with no table, and every scoring verb
  refuses to start without one. The staged order closes it: the new binary
  installs its table first, and only a verified install is `mv`d over the
  live path. A failed install aborts the deploy with the old binary scoring
  and leaves `dubplate-server.new` on the box, which the next deploy's rsync
  overwrites. The staged install sets `DUBPLATE_STATE_DIR` alone because
  `server.env` is not yet installed at that step. Under `--check`, the
  binary rsync names the live path rather than the staged one: the staged
  path never exists between deploys, so diffing it would report a change on
  every run. Acceptance of the `systemd-run` property list by the box's
  systemd 257 is unproved until the first deploy.
- **`serve` does not load the table.** `serve` does not score today, so only
  `shadow observe` and `shadow reevaluate` load it, and `shadow report`
  needs nothing. The verb that first scores inside `serve` (rung 9's tagging,
  or the folded sort key) adds the same load, per "nothing exported ahead of
  a caller".
- **The process-wide holder's tension for a published library (R-tt1).** A
  package-global setter that panics until called forces every importer's test
  binaries to carry a `TestMain`, and it rules out two tables in one process.
  This pass leaves the holder as is. The `func(rune) (string, bool)` seam
  type suits the library either way. Routed to `give-back`, beside the
  survey's AnyAscii default.
- **The plan's `time/tzdata` precedent was wrong.** R-tt1 cited it for "the
  program, not the library, decides where the table comes from", but
  `time/tzdata` never panics. The matcher's comment now cites
  `crypto.Hash.New`, which does panic for a hash no program registered. The
  design stands; only the comparison changed.
- **A `\xHH` of 0x80 or above and a negative rune were latent defects the pin
  hid.** A pinned input cannot exercise an unreachable path, so the table
  digest proves nothing about a lexer branch the wheel never takes. Rows that
  drive `extract`, `lex`, and `Lookup` directly are the only proof of those
  branches, which is why the plan split the exported surface from the
  unexported stages.

### Routed, not decided here

- **Git history (to `give-back`).** `unidecode-1.4.0.tbl` stays in every
  commit before `e8e158d`. The repo is private today. Whether history must be
  rewritten before it goes public is a question for the lawyer who confirms
  the distribution reading, and the photo lane carries the same shape of
  question for withdrawn images.
- **The Unidecode-derived items the repo still carries**, listed for that
  read in place of an absolute claim: `internal/engine/matcher/testdata/case-mapping-sweep.tsv`
  (56 data rows of post-strip outputs computed through Unidecode 1.4.0, its
  header says so); the pinned lookup rows in the new package's tests (about
  ten code points); `tools/unidecode-table/generate.py`, which imports the
  GPL library; and `Lookup`'s port of `_get_repl_str`'s rules, read from the
  GPL `__init__.py`.
- **The lawyer's confirmation** of the distribution reading (the binary
  downloads and derives the table at setup and carries no Unidecode data)
  before the first release strangers can install, with `packaging`.
- **Architecture-cleanup list:** the ROUTE items in
  `research/2026-09-27-translit-architecture-reads.md`, including the four
  `TestMain` shim copies and the four copies of the owner and leaf checks.
