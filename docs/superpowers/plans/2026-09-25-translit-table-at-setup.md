# The transliteration table leaves the binary: implementation plan

Status: **draft, revision 2.1, 2026-09-25; reviewed; two rulings open.**
Revision 2 folds the three `spec-plan-review` lenses (contract, mechanics,
risk) per `docs/superpowers/research/2026-09-25-translit-table-plan-fold.md`.
The largest change is Task 4's deploy transition: the new binary is staged
and installs its table before it replaces the live one. Revision 2.1 folds
the fold verification (same record, "Second fold"): the staged install now
takes `DUBPLATE_STATE_DIR` as its one property-set variable and runs before
the env step, so a failed install leaves the old binary on the old env. The
plan executes under either answer to each ruling.

Source: Geoff's decision of 2026-09-25, recorded in `ROADMAP.md`'s
`give-back` entry ("dubplate stays MIT (Geoff, 2026-09-25; path 1)",
`ROADMAP.md:159-166`) and argued in
`docs/superpowers/research/2026-09-25-give-back-survey.md` ("The Unidecode
transliteration table", "A path that keeps conformance"). The decision, as
recorded: the binary stops embedding the Unidecode-derived table and carries
a pinned checksum of the Unidecode 1.4.0 package instead. At setup it
downloads that package, verifies it, and builds the table into the state
directory with the generator's extraction logic ported to Go. There is an
offline path from a local file and a `doctor` row. The table stays
byte-identical, so conformance is unchanged. The matcher reads it through
the `func(rune) (string, bool)` seam the published matcher library needs.
The pass is due before the first release strangers can install (with
`packaging`), plus a lawyer's confirmation of the distribution reading
before that release.

**Code pointers are resolved at `1c012a8` on master.** Lane 1 (5c-iii) had
uncommitted edits under `web/` and `.claude/agent-memory/` while this plan
was drafted; nothing this plan cites was among them. Task 0 re-resolves
every pointer against the real HEAD.

**What this pass is.** A new package, `internal/engine/matcher/unidecode`,
owns the table: its binary format, the pinned wheel and table checksums, a
Go port of `tools/unidecode-table/generate.py`'s extraction, and a lookup
with unidecode's own `_get_repl_str` rules. Package `matcher` loses its
`//go:embed` and gains a process-wide transliterator seam. A new
`dubplate-server` verb builds the table into a state-directory leaf, online
from PyPI or offline from a wheel file. Every scoring verb loads and verifies
the table at startup and refuses to run without it. `doctor` reports it, and
`deploy.sh` stages the new binary and runs its install verb before swapping
it in, so the live shadow keeps scoring through the next deploy.

**What this pass is not.** It changes no scoring behavior, no corpus record,
no `corpus/manifest.json` field, and no Tunables value. It publishes no
module and adds no `go list -deps` boundary leg (both are `give-back`'s own
later work). It rewrites no git history: the committed `.tbl` stays in past
commits (see "Owner touchpoints"). It adds no module dependency. It writes
nothing under `web/`, `internal/serve`, `internal/auth`, `internal/engine/store`,
`internal/ops`, `internal/backup`, `systemd/`, `env/`, `config/`, `corpus/`,
`tools/venv/`, or `_salvage/`. **No deploy of any kind.** A task that
believes it needs any of these stops and reports.

> For agentic workers: the pass-execute workflow (six tasks with Task 0, at
> the workflow threshold), in two segments. Tasks 1 to 5 each run
> `dubplate-implementer` (`sonnet`, effort `high`), then `diff-reviewer`
> (`claude-opus-5-5`, effort `high`), then the gate inside the chain. Task 0
> is a read-only `sonnet` dispatch whose report the conductor reads directly.
> The conductor consumes structured reports and never reads a diff, a test
> log, or a gate transcript inline. Subagents never dispatch agents, so
> `code-simplifier` is the conductor's dispatch. The conductor owns
> `docs/STATUS.md`, `docs/HISTORY.md`, `ROADMAP.md`, `CLAUDE.md`, and the
> lane registry.

---

## Rulings awaiting owner

Two questions. The 2026-09-25 decision stands and is not reopened.

**1. Does this pass run under a 3.3M execution ceiling?** Recommendation:
yes. Reason: "Token ceiling" projects 2.63M from the matcher word-boundary
pass's measured per-task actuals, scaled for a new package, a new verb in
`cmd/dubplate-server`, the revision 2 folds, and three merge-side
architecture reads. The ceiling is the projection divided by 0.8, rounded
up to the next 0.1M, so the 80% flag (2.64M) sits just above the
projection. Planning is recorded outside the ceiling.

What each answer builds: yes sets the flag at 2.64M, only 0.01M above the
projection, so almost any overrun raises the flag and the conductor asks
before the close. No holds a 2.9M ceiling (flag 2.32M). Cumulative spend
with the fix round is about 2.01M after Task 4 and 2.28M after the
simplifier round and Task 5, so the projection crosses that flag during the
merge-side reads; the conductor asks after Task 4, before the simplifier
round.

**2. Is the operator verb `dubplate-server unidecode install [--from FILE]`?**
Recommendation: yes. Reason: the offline path tells an operator to download
"Unidecode 1.4.0's wheel from PyPI", so a verb named for the package they
fetch reads plainly. The alternative, `translit install`, abbreviates a word
an operator does not otherwise meet. This is operator-facing copy, so it is
a taste call; the project site spec's owed `init`/`upgrade` naming question
(STATUS, "Next action") will compose this verb, and a later rename is one
cobra `Use` string, one doctor row name, and one `deploy.sh` line.

What each answer builds: yes names the verb `unidecode install` and the
doctor row `unidecode`. No names them `translit install` and `translit`.
The state-directory leaf is `unidecode/` under either answer, because it is
named for its contents. Every mention of the verb below reads "the install
verb" and means whichever name the ruling picks.

---

## Goal

After this pass, no `dubplate-server` binary carries Unidecode data. A binary
scores strings only after it has loaded a table whose sha256 equals the one
pin it carries, built from the one wheel whose sha256 it pins, so every
verdict the matcher reaches is the verdict it reaches today. A stranger's
setup runs one verb, online or offline, and `doctor` says whether it ran.

---

## Owner touchpoints

The two rulings above, at plan approval. Nothing else is expected during
execution. Routed, not asked by this pass:

- **The lawyer's confirmation** of the distribution reading before the
  first stranger-installable release. `give-back` and `packaging` carry it.
  The reading, stated exactly: the binary downloads and derives the table at
  setup and carries no Unidecode data. The repo still carries four
  Unidecode-derived items, each plausibly de minimis, and the lawyer reads
  this list rather than an absolute claim:
  - `internal/engine/matcher/testdata/case-mapping-sweep.tsv`: 56 data
    rows of post-strip outputs computed through Unidecode 1.4.0 (its header
    says so);
  - the pinned lookup rows in the new package's tests (about ten code
    points and their Unidecode outputs);
  - `tools/unidecode-table/generate.py`, which imports the GPL library;
  - `Lookup`'s port of `_get_repl_str`'s rules, read from the GPL
    `__init__.py` (`translit.go` already carried the same port before this
    pass).
- **Git history.** The `.tbl` stays in every commit before this pass's
  Task 2. The repo is private today (`gh repo view`, per the survey). Whether
  history must be rewritten before the repo goes public is a question for
  the same lawyer's read; the photo lane carries the same shape of question
  for withdrawn images (STATUS, "Open items"). Task 5's HISTORY draft names
  it, and the conductor routes it to `give-back`.

What would become one: a byte mismatch the Go port cannot close (Task 1
stops), a module dependency, or any change to what the box runs beyond the
staged `deploy.sh` transition Task 4 adds.

---

## Token ceiling: 3.3M execution (Ruling 1)

The basis is the matcher word-boundary pass's recorded actuals
(`docs/HISTORY.md`, that entry's spend line): Task 0 0.08M; Task 1 0.14M
plus review 0.06M; Task 2 0.27M plus reviews 0.18M; Task 3 0.20M plus review
0.07M; the simplifier 0.11M plus review 0.04M; lane total about 1.25M. The
6b-ii-a pass (about 1.95M for three tasks, including merge-side reads) is
the second reference.

1. Task 0, a staleness read at `sonnet`: **0.08M**.
2. Task 1, the `unidecode` package and the wheel fetch: **0.58M**, above
   the matcher pass's largest task, since it carries a lexer, a byte-identity
   proof, a network function with its refusals, and (revision 2) the
   synthetic-zip rows and the rewriting transport.
3. Task 2, the matcher seam and the four test binaries: **0.35M**.
4. Task 3, the install verb, the scoring verbs' load, and the doctor row:
   **0.55M**; every full gate is heavy (Chromium) now, and revision 2 adds
   the leaf rows, the `--from` cap, the injected client, and the table
   fixture for every exit-0 doctor test.
5. Task 4, `deploy.sh`'s staged transition and the generator's new role:
   **0.25M** (revision 2 moves one block and splits the binary step).
6. Task 5, the close drafts: **0.12M**.
7. One fix round: **0.20M**.
8. The simplifier round with its one review: **0.15M**.
9. The merge-side reads: three `go-architecture-reader` dispatches
   (`internal/engine/matcher`, `internal/engine/matcher/unidecode`,
   `cmd/dubplate-server`) at **0.30M**, and the merge ritual at **0.05M**.
10. **The projection is 2.63M.** Revision 1 projected 2.35M; the folds
    add 0.08M to Task 1, 0.10M to Task 3, and 0.10M to Task 4.

Checkpoint: STATUS at the segment boundary after Task 3, and at the close.
Every task reports its spend, its mutation outcomes, and its measured gate
wall time.

Deliverable count: fifteen across Tasks 1 to 5, plus Task 0's report. Tasks
1, 2, and 3 carry four each, Task 4 carries two, and Task 5 carries one. No
task passes four.

---

## Primary inputs

Read these before Task 1, in this order.

1. **The decision and its evidence**: `ROADMAP.md:137-166` (`give-back`),
   and the survey's "The Unidecode transliteration table" through "A path
   that keeps conformance".
2. **`internal/engine/matcher/translit.go` at `1c012a8`**: the embed
   (`:11-12`), `UnidecodeVersion` (`:20`), the `0xeffff` cutoff (`:25`), the
   `0xffff` none sentinel (`:30`), `unidecodeTable` with `version`,
   `sweepLo`, `sweepHi`, `digest`, and `blocks` (`:36-43`),
   `loadUnidecodeTable`, a `sync.OnceValue` that panics on a corrupt embedded
   table (`:49-55`), the `tableCursor` reader (`:61-97`),
   `parseUnidecodeTable` (`:99-178`), `transliterateRune` (`:188-217`),
   `Transliterate` (`:219-229`), and `Levenshtein` (`:232-258`), which stays.
   `strdist.go:100-109` (`sdStrip`) is the only production caller of
   `Transliterate` (`git grep -n 'Transliterate(' -- '*.go'`).
3. **`internal/engine/matcher/translit_test.go`**: `TestUnidecodeVersion`
   (`:14`), `TestTransliterateRune` (`:21`), `TestTransliterate` (`:46`),
   `sweepDigest` (`:74`), `TestUnidecodeTableFullSweep` (`:105`),
   `TestLevenshtein` (`:113`).
4. **`tools/unidecode-table/generate.py`** in full: the file format in its
   module docstring, `harvest_blocks` (`pkgutil.iter_modules` over
   `unidecode.__path__`, names of length 4 starting `x`, each module's
   `data` tuple), `sweep_digest` (`_get_repl_str` per code point, the
   `>I` code point, `\x00` for None, `\x01` plus a `>I` length plus UTF-8
   bytes otherwise), and `write_table`.
5. **`internal/harness/conform/arch_test.go`**: `checkNoTimeNow` walks every
   `.go` file under `internal/engine`, `_test.go` files included, and fails
   on any `time.Now` call.
   **`internal/harness/conform/defaults.go:81-93`** (`CheckUnidecodeVersion`,
   comparing `manifest.Unidecode` with `matcher.UnidecodeVersion`) and its
   tests at `defaults_test.go:125-153`; `corpus/manifest.json:4`
   (`"unidecode": "1.4.0"`).
6. **The state-directory leaves**: `internal/engine/store/leaf.go`
   (`leafDirName = "store"`, mode `0o700`, `create`'s owner refusal through
   `checkOwner`, with the `runuser -u <owner> -- dubplate-server <command>`
   message), `internal/auth/store.go:12-40` (`auth`, `0o700`/`0o600`, and its
   reason: identity state is sensitive), and
   `internal/engine/shadow/store.go:39-60` (`shadow`, `0o755`/`0o644`, the
   rung 4 evidence tier). The configuration homes spec
   (`docs/superpowers/specs/2026-09-11-configuration-homes-design.md`, "The
   rule" and "The server's own knobs") places the state directory on the env
   wall; derived files under it are not knobs and have no other home.
7. **`cmd/dubplate-server`**: `main.go:127-151` (`newRootCmd`),
   `doctor.go:28-139` (the rows, the store row's always-print precedent at
   `:108-113`, the exit rule), `store.go:89-106` (`storeRow`'s "not checked"
   when `DUBPLATE_STATE_DIR` is unset), `doctor.go:400-411`
   (`lookupEnviron`), `shadow.go:250-310` (`observe`: refusals, then
   `shadow.OpenStore`, then `shadow.Run` at `:284`) and `:356-380`
   (`reevaluate`, `shadow.Reevaluate` at `:375`).
8. **`cmd/harness/main.go:51`** and its verbs (`synth`, `capture`,
   `seed-items`, `freeze`): the workstation tool this pass's fetch joins.
9. **The gate**: `Makefile:17-18` (`test`), `:128` (`check: web vet lint
   test web-rendered tidy-check`), `:139-140` (`derive`, outside `check`);
   `scripts/check.sh`'s header and its ruff leg (`:433-440`, tracked files
   only); `.golangci.yml:17-25` (why the exported-unused leg lives in
   `check.sh`).
10. **`scripts/deploy.sh`**, line numbers read at `6777d43`: the binary
    step (`:337-345`: `make build-linux`, then the `rsync` to
    `/usr/local/bin/dubplate-server` at `:343` and its `chmod` at `:344`),
    the smoke test (`:364-366`), `server.env`'s `0600 root:root` install
    (`:409-414`, after the binary step; the staged install no longer reads
    it), the `/var/lib/dubplate`
    pre-create (`:489-505`: `DUBPLATE_LIB_DIR` defined at `:502`, the
    `mkdir`/`chown` at `:505`), and `TIMER_UNITS` (`:526`, which
    deliberately omits `dubplate-shadow.timer`). `run_remote` (`:133-141`)
    prints `would run: <command>` under `--check`. The state-dir agreement
    leg (`scripts/check.sh:645-658`) reads `DUBPLATE_LIB_DIR` by pattern,
    wherever the line sits, and requires it to equal the env template's
    `DUBPLATE_STATE_DIR`. The wall refuses any `DUBPLATE_*` name it does
    not know (`internal/config/config.go:287`, "unknown configuration
    key"), so a binary older than the env it runs under refuses every
    scoring run.
    `systemd/dubplate-shadow.service`: `User=music`,
    `EnvironmentFile=/etc/dubplate/server.env`,
    `ReadWritePaths=/var/lib/dubplate/shadow`, the hardening block
    (`NoNewPrivileges` through `RestrictAddressFamilies=AF_INET AF_INET6
    AF_UNIX`, with `UMask=0027`), and its operator rule: a hand-run verb
    goes through `systemd-run` carrying the unit's properties, never the
    bare binary as root. `systemd/dubplate-shadow.timer:23`
    (`OnCalendar=*:0/5`). **The shadow timer is live on the box** (enabled
    and active, read-only check 2026-09-25): Geoff enabled it by hand at the
    launch sitting, and `deploy.sh` never stops or starts it.

### Behaviors this plan relies on, quoted

From `go doc` at `go1.27.1 linux/amd64`.

`go doc archive/zip.NewReader`: "NewReader returns a new Reader reading from
r, which is assumed to have the given size in bytes." `bytes.Reader`
implements `io.ReaderAt`, so the verified wheel bytes open in memory.

`go doc archive/zip.RegisterDecompressor`: "The common methods Store and
Deflate are built in." `go doc archive/zip.Store`: `Deflate uint16 = 8 //
DEFLATE compressed`. Every one of the wheel's 200 entries is method 8
(measured below).

`go doc archive/zip.File.Open`: "Open returns a ReadCloser that provides
access to the File's contents." `go doc archive/zip.ErrChecksum`:
`ErrChecksum = errors.New("zip: checksum error")`. The source
(`archive/zip/reader.go:300-330`) returns `ErrChecksum` at EOF when the
entry's CRC-32 is set and disagrees, and `ErrFormat` when more bytes arrive
than `UncompressedSize64`. A reader that stops before EOF sees neither, so
every entry is read to EOF.

`go doc archive/zip.Writer.CreateRaw`: "CreateRaw adds a file to the zip
archive using the provided FileHeader ... In contrast to
Writer.CreateHeader, the bytes passed to Writer are not compressed." The
doc does not say the header's sums survive; the source does. In
`archive/zip/writer.go:647-648` (Go 1.27.1) the raw close is `if w.raw {
return w.writeDataDescriptor() }`, returning before the non-raw path
overwrites `CRC32`, `CompressedSize64`, and `UncompressedSize64`. So a test
writes an entry whose CRC or size disagrees with its bytes, which is how
the post-checksum stage's CRC and size rows are reached. Two reader facts
from the source shape those rows. `reader.go:325` checks the CRC only when
it is set: `if r.f.CRC32 != 0 && r.hash.Sum32() != r.f.CRC32 { err =
ErrChecksum }`, so the CRC row's wrong CRC is nonzero. `reader.go:307-309`
returns `io.ErrUnexpectedEOF`, not `ErrFormat`, when an entry ends short of
`UncompressedSize64`; `ErrFormat` (`:301-302`) is the long entry.

`go doc crypto/sha256.Sum256`: "Sum256 returns the SHA256 checksum of the
data."

`go doc sync/atomic.Pointer`: "A Pointer is an atomic pointer of type *T.
The zero value is a nil *T." It is the race-free holder R-tt1 names.

`go doc os.MkdirAll`: "The permission bits perm (before umask) are used for
all directories that MkdirAll creates." The leaf's mode therefore depends on
the process umask unless the verb sets it explicitly after creation.

`go doc net/http.Client.Do`: "A non-2xx status code doesn't cause an
error." And: "If the returned error is nil, the Response will contain a
non-nil Body which the user is expected to close." And: "Any returned error
will be of type *url.Error", so a canceled request is tested with
`errors.Is(err, context.Canceled)`, never by equality. `go doc net/http.Client`
(`Timeout`): "The timeout includes connection time, any redirects, and
reading the response body." `go doc net/http.NewRequestWithContext`: "For an
outgoing client request, the context controls the entire lifetime of a
request and its response." `go doc net/http.DefaultTransport` shows `Proxy:
ProxyFromEnvironment`, and `go doc net/http.ProxyFromEnvironment`: the
proxy comes from "HTTP_PROXY, HTTPS_PROXY and NO_PROXY (or the lowercase
versions thereof)". A stranger behind a proxy needs no flag.

`go doc io.LimitReader`: "LimitReader returns a Reader that reads from r
but stops with EOF after n bytes." A body truncated at the cap is then
indistinguishable from a short one by EOF alone, so the size refusal reads
one byte past the cap.

`go doc os.UserCacheDir`: "On Unix systems, it returns $XDG_CACHE_HOME ...
if non-empty, else $HOME/.cache", and "Users should create their own
application-specific subdirectory within this one."

`go help test`: "Tests that open files within the package's module or that
consult environment variables only match future runs in which the files and
environment variables are unchanged." The source
(`$(go env GOROOT)/src/cmd/go/internal/test/test.go:2051-2063`) reads: "Do
not recheck files outside the module, GOPATH, or GOROOT root." The cached
wheel lives under `~/.cache`, so Go's test cache does not track it: a cached
`ok` for the four packages survives a deleted or replaced wheel, and only a
run that misses the cache rebuilds the table. `make test` is `go test ./...`
with no `-count` (`Makefile:17-18`).

`man systemd-run` (systemd 259): `--wait`: "Synchronously wait for the
transient service to terminate." EXIT STATUS: "If systemd-run waits for
the service to terminate, the return value will be propagated from the
service." So a refusing verb fails the remote command, and `run_remote`'s
`ssh` returns nonzero. `--pipe`: "standard input, output, and error of the transient
service are inherited from the systemd-run command itself." `--collect`:
"Unload the transient unit after it completed, even if it failed."
`--property=, -p`: "Sets a property on the scope or service unit that is
created. This option takes an assignment in the same format as
systemctl(1)'s set-property command." The box runs **systemd 257**
(`systemctl --version` over read-only ssh, 2026-09-25: `systemd 257
(257-23.el10_2.2.alma.1-gdb72f78)`). The man pages' "Added in version"
lines put every option and property Task 4 uses at or below 236: `-p` 211,
`--wait` 232, `--pipe` 235, `--collect` 236, `ReadWritePaths=` 231,
`RestrictAddressFamilies=` 211, `NoNewPrivileges=` 187. A `--user` run of
the exact list on the workstation's 259 parsed every `-p`, including the
space-bearing `RestrictAddressFamilies` (the fold verification). Acceptance
on the box's 257 is proved by the first deploy (see "Not verified").

`go doc time/tzdata`: "This package should normally be imported by a
program's main package, not by a library. Libraries normally shouldn't
decide whether to include the timezone database in a program." This is the
stdlib's stance on a data table a library consults but should not carry
(R-tt1).

**PyPI, fetched 2026-09-25** from `https://pypi.org/pypi/Unidecode/1.4.0/json`
(`urls` array):

| Field | Wheel | sdist |
| --- | --- | --- |
| `filename` | `Unidecode-1.4.0-py3-none-any.whl` | `Unidecode-1.4.0.tar.gz` |
| `size` | 235837 | 200149 |
| `sha256` | `c3c7606c27503ad8d501270406e345ddb480a7b5f38827eafe4fa82a137f0021` | `ce35985008338b676573023acc382d62c264f307c8f7963733405add37ea2b23` |
| `blake2b_256` | `8fb7559f59d57d18b44c6d1250d2eeaa676e028b9c527431f5d0736478a73ba1` | `947da8a765761bbc0c836e397a2e48d498305a865b70a8600fd7a942e85dcf63` |
| `url` | `https://files.pythonhosted.org/packages/8f/b7/559f59d57d18b44c6d1250d2eeaa676e028b9c527431f5d0736478a73ba1/Unidecode-1.4.0-py3-none-any.whl` | (not used) |
| `yanked` | false | false |

`info.license` reads `GPL`. The file URL's path is the file's own
`blake2b_256` split `8f/b7/<rest>`, so the URL is content-addressed. A `HEAD`
on it returned `HTTP/2 200`, `content-length: 235837`, and `cache-control:
max-age=365000000, immutable, public`. The downloaded file's `sha256sum` and
`blake2b` (digest size 32) both matched the JSON. The wheel's
`Unidecode-1.4.0.dist-info/METADATA` reads `Version: 1.4.0`.

**The wheel's contents, measured 2026-09-25** (scratch scripts outside the
tree):

- 200 entries, every one Deflate. 190 are `unidecode/x[0-9a-f]{3}.py`
  (`x000.py` to `x1f6.py`); the other ten are `__init__.py`, `__main__.py`,
  `py.typed`, `util.py`, and six `dist-info` files. The largest entry
  uncompresses to 18,092 bytes and the total to 903,391.
- Each of the 190 data modules is byte-identical to the pinned venv's copy
  under `tools/venv/.venv/lib/python3.13/site-packages/unidecode/`, and so is
  `__init__.py`.
- The data modules' grammar, from Python's own `tokenize` over all 190: every
  file's first line is `data = (`. The tokens are 43,120 strings, 5,477
  `None` names, one `data` name per file, 48,597 commas (one after every
  entry, the last included), one `=`, `(`, and `)` per file, and 48,486
  comments. 43,119 strings are single-quoted and one is double-quoted
  (`x020.py:93`, `"''''"`). The mechanics review's Go prototype confirmed
  the grammar and two shapes a line-per-entry lexer would miss: `x000.py`
  packs sixteen entries per line, and three modules carry standalone
  comment lines holding `"` and `'`. No string carries a prefix or triple quotes, and
  no two strings are adjacent (no implicit concatenation). The escapes are
  `\'` (128), `\\` (18), and `\x0a` (3, at `x020.py:42-43`). Every source
  byte is ASCII, and no file carries a carriage return. `'#'` is a string
  entry in `x006.py`, `x025.py`, and `x028.py`, so a comment ends a line
  only outside a string.
- **The Go port's feasibility, proved in Python.** A scratch prototype read
  the wheel with `zipfile`, never importing it, lexed each data module with a
  hand-written lexer accepting exactly that grammar, computed the sweep
  digest from the lexed blocks under `_get_repl_str`'s rules, and wrote the
  table in `generate.py`'s format. Its output was 255,458 bytes with sha256
  `c954605b805e4a21d1d9e79b4d93bc6b89684de9996dd21720b256fc58a6ce85`,
  identical to `sha256sum internal/engine/matcher/unidecode-1.4.0.tbl` at
  `1c012a8`: 190 blocks and 48,597 entries.
- **The Go port's feasibility, proved in Go.** The mechanics review
  (`docs/superpowers/research/2026-09-25-translit-table-plan-review-mechanics.md`)
  wrote a stdlib-only prototype outside the tree: `zip.NewReader` over the
  verified bytes, every entry read to EOF, a strict token lexer, the sweep,
  and `generate.py`'s format. Its output had sha256
  `c954605b805e4a21d1d9e79b4d93bc6b89684de9996dd21720b256fc58a6ce85`, and
  `cmp` against the committed table reported no difference. The build took
  about 0.14s, sweep included, race detector off.

**`_get_repl_str`, read from the pinned venv**
(`tools/venv/.venv/lib/python3.13/site-packages/unidecode/__init__.py`):
below `0x80` it returns the character itself; above `0xeffff` it returns
`None`; a surrogate warns and continues; otherwise `section = codepoint >> 8`
and `position = codepoint % 256`, a section with no module is `None`, and
the lookup is `if table and len(table) > position: return table[position]`,
else `None`. So an empty data tuple would read as `None` for every position;
no shipped module is empty, and the port keeps the rule.

**The test binaries that reach the table**, measured with `go test -overlay`
replacing `translit.go` by a copy whose `Transliterate` panics on entry (no
tree edit), over every package whose test binary links `matcher`
(`go list -test`): `internal/engine/matcher`, `internal/engine/tag`,
`internal/engine/shadow`, and `internal/harness/conform` fail;
`internal/engine/musicbrainz` and `cmd/dubplate-server` pass. None of the
four has a `TestMain` today (`git grep -l 'func TestMain'`). The set is the
same under no tag, `oracle`, and `rendered` (mechanics review).
`cmd/dubplate-server` already has a `TestMain` (`backup_test.go:68`) and
gets no table install there; after Task 3 its test binary reads the cached
wheel too, through the install verb with `--from`, so the "fail naming
`make unidecode-wheel`" rule covers it as well.

**How the gate treats the network today.** `make check` runs no fetch.
Its prerequisites are fetched out of band and refused when absent or stale,
never fetched by the gate: the `web` target refuses a missing node or npm
and "a stale or drifted `npm ci --prefix web`" (the `web` target, `Makefile:43` onward,
`check.sh` header), and the rendered leg needs Playwright's Chromium
installed beforehand (STATUS, lane registry). No CI workflow exists
(`.github/workflows` is absent). This pass follows the `npm ci` precedent.

---

## Global constraints

- **The main tree, as lane 1, when lane 1 is free.** The pass edits
  `internal/harness/conform`, and CLAUDE.md reads: "Corpus and harness work
  never moves to a worktree." It also edits `cmd/dubplate-server`, which
  lane 1's 5c-iii (`web.go`, `serve.go`, `rendered_test.go`) and lane 2's
  6b-ii-b (restore verbs, possibly `doctor.go` and `main.go`) write. The pass
  launches when the lane registry shows lane 1 closed and lane 2 either
  closed or reporting no edit to `doctor.go`, `main.go`, or `shadow.go`
  (Task 0, blocker 1).
- **Package territory, stated as a wall.** The pass creates
  `internal/engine/matcher/unidecode/` and `cmd/dubplate-server/unidecode.go`
  (with its test) and `cmd/harness/fetchunidecode.go` (with its test). It
  edits `internal/engine/matcher/translit.go` and `translit_test.go`, deletes
  `internal/engine/matcher/unidecode-1.4.0.tbl`, adds a `TestMain` file to
  each of `internal/engine/matcher`, `internal/engine/tag`,
  `internal/engine/shadow`, and `internal/harness/conform`, and edits
  `internal/harness/conform/defaults.go` and `defaults_test.go`,
  `cmd/dubplate-server/main.go`, `doctor.go`, `doctor_test.go`, `shadow.go`,
  `shadow_test.go`, `cmd/harness/main.go`, `Makefile`, `scripts/deploy.sh`,
  and `tools/unidecode-table/generate.py`. Task 5 writes drafts only.
- **No `time.Now` under `internal/engine`, tests included.**
  `checkNoTimeNow` (`internal/harness/conform/arch_test.go`) fails the
  conform suite on one. A timeout comes from the caller's client or
  context; wall time comes from `go test -v` output, never a clock read in
  code.
- **The gate stays hermetic.** No leg of `scripts/check.sh` or `make check`
  opens a network connection. Tests read the pinned wheel from a
  user-cache file that the workstation fills once with `make
  unidecode-wheel`, and fail, never skip, naming that command when it is
  absent (R-tt5). A check either gates or does not exist.
- **Python stays workstation-only, and the gate never runs it.**
  `generate.py` stays tracked, so the ruff leg keeps linting it; Task 4
  changes its docstring and its output handling only (R-tt7). The
  implementer invokes `python-conventions` before editing it and runs
  `ruff check` and `ruff format --check` on it directly.
- **`go-conventions` before any Go**, at the exemplary-stdlib bar;
  `golang-spf13-cobra` for the verb; `vps-conventions` before `deploy.sh`.
  **The em dash is banned in Go comments.** No process citation (rung, task,
  ruling, resolution, plan line, or research path) in a shipped Go comment;
  the staged leg warns on one.
- **Every new exported symbol has a non-test caller in another file.** The
  staged exported-unused leg (CLAUDE.md, "Gate") flags an exported
  `internal/` symbol whose only references are its own file and `_test.go`
  files. Each exported name below lists its caller.
- **Assertion discipline at the repo's bar.** Every hard-coded expectation
  cites PyPI's JSON, the committed table's `sha256sum`, or a figure this plan
  quotes. Every test names the fixture state it needs (the cached wheel
  present, a table installed) so it cannot pass vacuously.
- **Universality.** No estate hostname, path, user, or personal name in any
  new Go file. The only URL is PyPI's content-addressed file URL.
- **The gate is** `cairn-run-gate 'bash /var/home/glw907/Projects/dubplate/scripts/check.sh'`,
  green at every task end, re-issued on exit 75 until it prints `gate exit:`.
  Every full gate is heavy (Chromium), so it runs in the default lane.
  Evidence runs use `-count=1` and report wall time. Commit specific files
  only, footer `Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>` or
  the dispatch's footer. **The workstation's wheel cache is filled by Task
  1** (`make unidecode-wheel`), whose own byte-identity test reads it; the
  conductor confirms the cached file's `sha256sum` before Task 2.
- **Every download in a test goes through a rewriting transport.** Tests
  never reach PyPI. The one test transport is an `http.RoundTripper` that
  asserts the request URL equals `WheelURL`, counts the request, and
  forwards a clone to an `httptest` server; any other host is refused with
  a test failure, never forwarded. `go doc net/http.RoundTripper`:
  "RoundTrip should not modify the request, except for consuming and
  closing the Request's Body." So the transport rewrites the URL on
  `req.Clone(req.Context())` (`go doc net/http.Request.Clone`: "Clone
  returns a deep copy of r with its context changed to ctx"), never on the
  request it was handed. Under every named mutation, a stray
  request fails the test instead of opening a connection.

---

## Privileged fields and their channels

| Field | Channels that may set it | What refuses every other channel |
| --- | --- | --- |
| The matcher's installed transliterator | `cmd/dubplate-server`'s scoring-verb startup (`shadow observe`, `shadow reevaluate`), after `unidecode.Parse` verified the table; each of the four `TestMain` files, after `unidecode.Build` verified the cached wheel and the table. `doctor` never installs: it calls the load half only. The uninstalled-panic test clears the holder through package-internal access, never through the exported setter. | No type refuses it, since the setter is exported for the cmd caller. The setter refuses a nil function by panicking, so it offers no uninstall channel. The exit gate's `git grep` finds exactly one non-test call site. A scoring call before any installation panics, naming the setter (R-tt1). |
| The download client | `main`'s constructor argument to the install verb and to `harness fetch-unidecode` (the bounded production client); tests pass a client carrying the rewriting transport. | Neither verb builds its own client, so no test can fall through to `http.DefaultTransport`. |
| The table file in the leaf | The install verb alone: it writes atomically, only bytes whose wheel and table checksums both matched. | Every load runs `unidecode.Parse`, which refuses bytes whose sha256 differs from the pin, so a hand-copied, truncated, or edited file is refused at startup and by `doctor`. |
| The wheel and table pins (version, wheel file name, URL, wheel sha256, table sha256, the data-module count 190) | Source edits to the `unidecode` package, re-derived by `generate.py` under the pinned venv (R-tt7). | `TestBuildIsByteIdenticalToThePin` (Task 1) hashes the build itself and compares with a literal in the test file, independent of the package's pin constant, so a pin edited to match a wrong build still fails; `CheckUnidecodeVersion` fails when the version differs from `corpus/manifest.json`. |
| The workstation wheel cache | `harness fetch-unidecode` (`make unidecode-wheel`), which writes only bytes whose sha256 matched. | Every test read runs `unidecode.Build`, which verifies the wheel's sha256 first, so a hand-placed wrong file reds the gate with the checksum named. |

---

## Resolutions this plan takes rather than discovers

A task landing the surface a resolution decides writes nothing about the
resolution in a shipped comment; the close maps each resolution to its
landing commit.

### R-tt1: the seam is process-wide, installed by the main package

Package `matcher` gains one exported setter taking a `func(rune) (string,
bool)` and keeps an unexported `sync/atomic.Pointer` holder for it.
`Transliterate` loads the holder once per call, not once per rune. `Transliterate`
reads the installed function for every rune, ASCII included; the `< 0x80`
passthrough and the `0xeffff` cutoff move into the table's lookup, because
they are `_get_repl_str`'s rules and the seam carries `_get_repl_str`
whole. With nothing installed, `Transliterate` panics with a message naming
the setter: an uninstalled seam is a wiring defect in a program, never a
runtime condition, because every scoring verb refuses to start without a
verified table (R-tt4). The panic fires on every input, so it can never hide
behind ASCII-only test data.

The alternative, a field on `Tunables`, is wrong for three measured reasons:
`Tunables` is hashed as JSON by `shadow.configHash`
(`internal/engine/shadow/match.go:55-75`, whose comment states "Tunables and
Compat are plain data, no channel, func or cyclic pointer among their
fields"); `reflect.DeepEqual` treats two non-nil funcs as unequal, which
seven `DeepEqual` assertions over `Tunables` values would meet; and 160 test
call sites build `DefaultTunables()` or `QuietImportTunables()`. More
basically, the table has exactly one correct value per process: conformance
requires Unidecode 1.4.0, and nothing legitimately scores one call under a
different table. That is the `time/tzdata` shape quoted above: the library
consults the data, and the program's main package decides where it comes
from. Installing it in `main`'s verb path is an explicit dependency, never
an `init` side effect.

The process-wide holder is right for dubplate's binary and an open question
for the published library: a package-global setter that panics until called
forces every importer's test binaries to carry a `TestMain`, and it rules
out two tables in one process. The seam's type suits the library either
way. This pass does not change the holder; Task 5's HISTORY draft names the
tension and the conductor routes it to `give-back`, beside the survey's
AnyAscii default.

### R-tt2: the new package owns the table and moves with the matcher

`internal/engine/matcher/unidecode` holds everything about the table that
the published matcher module needs: the format (the decoder moved from
`translit.go`, plus an encoder), the pins, the extraction port, the sweep
digest, the lookup, the wheel download, and the cache path. It imports only
the standard library. Package `matcher`'s non-test code does not import it;
only `cmd/dubplate-server` and the four `TestMain` files join the two. The
state-directory leaf, the verb, and the doctor row are dubplate's own, so
they live in `cmd/dubplate-server`. Exported surface and its non-test
callers:

| Name | Shape | Non-test caller |
| --- | --- | --- |
| `Version` | const, `"1.4.0"` | `conform.CheckUnidecodeVersion`; the cmd table file name |
| `WheelURL`, `WheelSHA256` | consts, the PyPI values above | the install verb's offline guidance text |
| `TableFileName` | const, `"unidecode-1.4.0.tbl"` | the cmd leaf path |
| `Fetch(ctx, *http.Client) ([]byte, error)` | downloads the pinned wheel, capped at `MaxWheelSize`, and verifies it | the install verb; `harness fetch-unidecode` |
| `MaxWheelSize` | const, 1 MiB | the install verb's `--from` read, which applies the same cap-plus-one refusal |
| `ErrWheelTooLarge` | sentinel | the install verb matches it for both the download and `--from` |
| `Build(wheel []byte) ([]byte, error)` | verifies the wheel, extracts, builds, verifies the table | the install verb |
| `Parse(table []byte) (*Table, error)` | verifies the table's sha256, decodes | the scoring verbs' startup; the doctor row |
| `(*Table).Lookup(r rune) (string, bool)` | `_get_repl_str` | installed through the matcher setter in cmd |
| `WheelCachePath() (string, error)` | the workstation cache file path | `harness fetch-unidecode` |
| `ErrWheelChecksum`, `ErrTableChecksum` | sentinels | the install verb and doctor row match them with `errors.Is` |

The table's sha256 pin and the data-module count stay unexported: only
`Build` and `Parse` read them. The cache path sits under
`os.UserCacheDir()` in a subdirectory named for the package, never for
dubplate, so the published module keeps it.

Two unexported stages carry the checks that the pins make unreachable
through the exported functions, and their tests drive them directly:
`extract` (the zip read, every entry to EOF, the entry selection and count,
and the lexer, returning the blocks), which `Build` calls after the wheel
checksum; and `decode` (the truncation-checked reader and its trailing-byte
refusal), which `Parse` calls after the table checksum. Any altered wheel or
table stops at a checksum sentinel, so a row asserting a zip, count, lexer,
or truncation refusal through `Build` or `Parse` could never be written, or
would pass vacuously on the checksum.

### R-tt3: the table is built from the wheel only, with a strict lexer

The one pinned artifact is the wheel. The sdist (a gzip tar) is declined:
a second format doubles the extraction code for no reach, since an offline
operator can download the wheel on any machine. `Build` runs in this order,
and each step refuses before the next: the wheel's sha256 against the pin
(so nothing unverified is ever unzipped); a zip read with every entry read
to EOF; exactly the 190 entries matching `unidecode/x[0-9a-f]{3}.py`, the
Go form of `generate.py`'s `len(name) == 4 and name[0] == "x"` plus a hex
parse; a lexer that accepts exactly the grammar measured above and refuses
anything else, naming the file, line, and column; the blocks sorted by
section; the sweep digest over `[0, 0x10FFFF]` computed through the same
`Lookup` over the in-memory blocks (`Parse` cannot be called yet, since the
table's bytes do not exist until the digest is written), with
`generate.py`'s encoding; the table written in
`generate.py`'s format with `Version` as its version field; and the table's
sha256 against the pin. The lexer is strict on purpose: a pinned wheel
cannot change, so any token outside the measured set is a defect in the
lexer, and a refusal names it where a lenient parse would drift.

Why the digest in a Go-built table still proves something: the pin is the
sha256 of the file `generate.py` wrote, whose digest field Python computed
by calling unidecode's own `_get_repl_str`. A Go build that matches the pin
byte for byte has therefore reproduced Python's digest, and so Python's
lookup, for every code point.

### R-tt4: absent means refuse; there is no degraded mode

A scoring verb that finds the table absent, unreadable, or failing `Parse`
refuses to start, names the path and the reason, and names the install verb.
It creates nothing first: the check runs after the verb's existing
configuration refusals and before `shadow.OpenStore`. Degrading (scoring
with no transliteration, or with a permissive table) is refused outright:
the album is the verdict unit, and a changed `string_dist` on one track
silently changes an album's verdict, the one failure the shadow's evidence
exists to rule out. Today the scoring verbs are `shadow observe` and
`shadow reevaluate`; Task 0 confirms that `shadow report` scores nothing.
`serve` does not score today, so it does not load the table; the verb that
first scores in `serve` (rung 9's tagging, hook 2's folded sort key) adds
the same load, under "nothing exported ahead of a caller".

The matcher commit keeps identifying the transliteration: the binary accepts
exactly one table, by checksum, so `matcher_commit` and
`matcher_config_hash` need no new stamp.

### R-tt5: tests read a checksum-verified wheel from the user cache

`make unidecode-wheel` runs `go run ./cmd/harness fetch-unidecode`, which
downloads the pinned wheel through `unidecode.Fetch` and writes it
atomically to `unidecode.WheelCachePath()`. It is a no-op, with no network,
when the cached file already verifies. Each of the four `TestMain` files
reads that file, runs `Build` then `Parse`, installs the lookup, and runs
the tests. On any failure it prints the reason and `run make
unidecode-wheel`, and exits nonzero without running a test. The user cache,
never a gitignored repo directory, so a second-lane worktree shares one
download, as it shares Go's module cache.

Every test binary that runs rebuilds the table from the wheel, so every
uncached test run re-proves the port byte-identical. Go's test cache does
not track the wheel (quoted above), so a gate run whose sources are
unchanged reports a cached `ok` without rebuilding; that is sound, because
the wheel is pinned by sha256 and cannot change under a valid cache entry.
The one state the cache can hide is "wheel absent", so every proof of that
state runs with `-count=1`. The mechanics prototype measured the build at
about 0.14s; Task 1 re-measures it, and above 0.5s per binary the conductor
decides whether the three downstream binaries read a cached table file
instead.

`make check` gains no dependency on the fetch target, so the gate never
opens a connection; `make test`'s failure message is the prompt. Any future
CI runs `make unidecode-wheel` first, with the cache directory persisted.

### R-tt6: the leaf, its modes, and the install verb's contract

The table lives at `$DUBPLATE_STATE_DIR/unidecode/unidecode-1.4.0.tbl`. The
file name carries the version, so a later pin bump writes a new file beside
the old one, and a rolled-back binary still finds its own. The leaf is
`0o755` and the file `0o644`, the shadow leaf's tier: the table is public
data, and integrity comes from the checksum at every load, not from the
mode. The auth leaf's `0o700` reason (sensitive identity state) does not
apply.

The install verb reads `DUBPLATE_STATE_DIR` alone through `lookupEnviron`,
never `config.Load`, because setup runs it before the rest of the wall may
be rendered. Its refusals run in this order, each writing nothing, and all
of them before the "already installed" check, so a root run on a valid
table still refuses:

1. the variable is unset;
2. the state directory is absent;
3. the effective user does not own the state directory;
4. the `unidecode/` leaf exists but is not a directory, is owned by
   another user, or has a mode other than `0755` (the store leaf's
   `checkLeafShape` and `checkOwner` are the precedent, `leaf.go:104-129`).

The ownership refusals name the owner and the `runuser -u <owner> --
dubplate-server <verb>` fix in `internal/engine/store/leaf.go`'s own words,
so a root run never leaves a root-owned leaf. When the file already exists
and `Parse` accepts it, the verb prints that the table is installed and
exits 0 with no network. Otherwise it obtains the wheel (`Fetch`, or the
`--from` file read under the same `MaxWheelSize` cap-plus-one refusal),
runs `Build`, creates the leaf and sets its mode to `0755` explicitly with
`os.Chmod` (so the mode holds under any umask), writes through
`internal/atomicfile`, reads the file back through `Parse`, and prints the
path and version. A failed download or build writes nothing and leaves any
existing file in place. The verb's constructor takes the `*http.Client`,
the way `newDoctorCmd` takes `environ`; `main` passes a client with a
bounded timeout, and the cap (1 MiB, well above 235,837 bytes) refuses an
over-cap body.

The verb's offline guidance (its long help, and the text of every download
refusal) states the wheel's URL, its sha256, and `--from`.

### R-tt7: `generate.py` becomes the pin's oracle

The Go port replaces `generate.py` at runtime. Python stays in the loop only
where it is independent: re-deriving the pin. On a Unidecode bump, a
maintainer bumps the venv, runs `generate.py` to a scratch path, and takes
the file's sha256 as the new table pin, and updates the data-module count
from the new wheel's listing; the Go build of the new wheel must then match
the pin. `generate.py`'s docstring says so, names `unidecode.Version`,
the two sha256 pins, and the data-module count as what the run updates, states that its output
must never be committed, and drops the committed-path example. Its output
path stays an argument. It prints the output's sha256 beside its existing
figures. The venv stays the source of truth for beets and its dependencies;
the Unidecode version the venv pins and `Version` stay equal, which
`CheckUnidecodeVersion` and `oracle.reconcile` already enforce through
`corpus/manifest.json`.

---

## Review focus

1. **Byte identity.** Expected: the Go build of the pinned wheel equals the
   literal `c954...ce85` on every uncached test run. Pinned in Task 1,
   re-proved by every `TestMain` in Task 2 whenever its package misses the
   test cache.
2. **A scoring verb that runs without a verified table.** Expected: never;
   it refuses before creating anything, and an uninstalled seam panics on
   the first call of any input. Pinned in Tasks 2 and 3.
3. **A table file that is present but wrong** (truncated, hand-edited, an
   older version, empty, a directory). Expected: `Parse` refuses it, the
   scoring verb refuses, `doctor` fails the row, and the install verb
   replaces it. Pinned in Task 3.
4. **A network failure or a wrong download.** Expected: nothing is written,
   an existing good table stays, and the install verb's refusal carries
   the offline guidance. Pinned in Task 1 (the library's refusal names
   the status) and Task 3 (the verb adds the guidance).
5. **The gate touching the network.** Expected: never. The diff-reviewer
   checks that every test client carries the rewriting transport and that
   no verb builds its own client; Task 1's `Fetch` tests run against
   `httptest` only.
6. **The deploy transition on a box whose shadow timer is live.** Expected:
   the live binary is replaced only after the staged binary has installed
   and verified its table, so a failed download leaves the old binary
   scoring. Pinned in Task 4 by reviewer read against the property list.

---

## Task 0: The staleness pre-task

Model: `sonnet`, fresh context, before any implementer. **Its output is
blockers with evidence, not fixes.** It writes no code and edits no file.
Every check runs against the real HEAD, and the task reports that HEAD. A
pointer whose content survives at a new offset is a note naming the new
site; a pointer whose content is gone is a blocker.

**Blockers (a "no" stops the launch):**

1. Is lane 1 closed in the lane registry, and is lane 2 closed or showing no
   commit touching `cmd/dubplate-server/doctor.go`, `main.go`, or
   `shadow.go` since `1c012a8`? Is the main tree clean of changes the
   conductor did not author? Report `git status`, the registry rows, and
   `git log 1c012a8..<lane-2 branch> -- cmd/dubplate-server`.
2. Do `git log 1c012a8..HEAD -- internal/engine/matcher internal/harness/conform
   cmd/dubplate-server cmd/harness Makefile scripts/deploy.sh tools/unidecode-table`
   and a diff of the same paths show no change to a surface this plan names?
   Report each commit by function.
3. Does `sha256sum internal/engine/matcher/unidecode-1.4.0.tbl` still read
   `c954605b805e4a21d1d9e79b4d93bc6b89684de9996dd21720b256fc58a6ce85`, and
   does `corpus/manifest.json` still pin `"unidecode": "1.4.0"`?
4. Does PyPI's JSON for Unidecode 1.4.0 still give the wheel sha256
   `c3c7606c27503ad8d501270406e345ddb480a7b5f38827eafe4fa82a137f0021`, and
   is the release still present and unyanked? (A read-only fetch of the
   JSON; no download into the tree. A deleted release returns 404 for the
   JSON and the file URL alike.)
5. Does the overlay probe (Primary inputs, "The test binaries that reach
   the table") still name exactly the four packages? Run it with the overlay
   under the scratchpad, never a tree edit.

**Notes:**

6. Report which `cmd/dubplate-server` verbs reach `shadow.Run` or
   `shadow.Reevaluate`, and confirm `shadow report` reaches neither.
7. Report every `cmd/dubplate-server` test that runs `shadow observe` or
   `shadow reevaluate` past their configuration refusals; Task 3 must give
   each one an installed table.
8. Report the doctor tests that pin the row count or the last row's name,
   and every doctor test that expects exit 0 with `DUBPLATE_STATE_DIR` set
   (at `6777d43`: `TestDoctorCmd`'s "full valid wall" case at
   `doctor_test.go:45-48`, and the tests setting the variable at `:415`,
   `:915`, `:926`, `:937`, `:992`, `:1045`, `:1077`, and `:1117`). After
   Task 3 each one needs an installed table.
9. Report `go version`; a version other than 1.27.1 is a note only.

**2.0 hook:** neutral. **Deliverable count: 1** (the report).

---

## Task 1: The `unidecode` package and the wheel fetch

Model: `sonnet`; the conductor upshifts to `opus` only if the implementer
reports a byte mismatch it cannot close. Depends on Task 0. **2.0 hook:
neutral.** Hook 2 (the folded sort key, rung 9) will read the table through
the matcher's seam; this package serves it without change. The package is
the give-back seam's home (ROADMAP `give-back`, not a numbered hook).

**Files:**

- Create: `internal/engine/matcher/unidecode/` (package doc, the format and
  lookup, the extraction, the fetch and cache path, and their tests; the
  file split is the implementer's, one responsibility per file)
- Create: `cmd/harness/fetchunidecode.go`, `cmd/harness/fetchunidecode_test.go`
- Modify: `cmd/harness/main.go` (register the verb), `Makefile` (the
  `unidecode-wheel` target, its comment, and `.PHONY`)

**Interfaces:** produces the R-tt2 table's names. Consumes nothing new.
`translit.go` is untouched in this task; its decoder is copied, and Task 2
deletes the original.

**Outcomes:**

- **The format and the lookup.** `Parse` verifies the table's sha256 against
  the pin, then decodes with `translit.go`'s truncation-checked reader and
  its trailing-byte refusal. `Lookup` carries `_get_repl_str` whole: the
  `< 0x80` passthrough, the `0xeffff` cutoff, an absent section, a position
  past a block's end, and the `0xffff` sentinel all behave as today. The
  encoder writes `generate.py`'s format exactly.
- **The extraction.** `Build` per R-tt3, in that order, each refusal naming
  its stage.
- **The fetch and the cache.** `Fetch` issues one GET to `WheelURL` with the
  caller's client and context, refuses a non-200 status naming it, reads at
  most `MaxWheelSize` plus one byte and refuses an over-cap body with
  `ErrWheelTooLarge`, and refuses a checksum mismatch with
  `ErrWheelChecksum`, naming both sums. `WheelCachePath` per R-tt5.
  `harness fetch-unidecode` takes its `*http.Client` from `main` (the
  channels table), is a no-op when the cached file verifies, else fetches
  and writes atomically, replacing a cached file that fails its checksum;
  `make unidecode-wheel` runs it. The target's comment says the gate never
  runs it and why.
- **The tests.** Lexer rows over synthetic sources; `extract` rows over
  synthetic zips built in memory with `archive/zip.Writer` (`CreateRaw` for
  the CRC and size rows), each asserting the stage's own error and
  `!errors.Is(err, ErrWheelChecksum)`; `decode` rows for truncation and
  trailing bytes; `Parse` and `Lookup` rows; `Fetch` over `httptest`
  through the rewriting transport only; the byte-identity test over the
  cached wheel.

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| The wheel given to `Build` | sha256 differs from the pin | `ErrWheelChecksum`, naming both sums; nothing unzipped |
| The zip given to `extract` (synthetic) | not a zip, or an entry whose CRC or size disagrees (`CreateRaw`) | the zip error, naming the entry |
| The data entries given to `extract` (synthetic) | fewer or more than the pinned count, or a name outside the pattern that starts `unidecode/x` | refuses, naming the count or the name |
| A data module (synthetic source) | a token outside the measured grammar (a prefix, triple quotes, an unknown escape, a missing comma, text after `)`, a non-ASCII byte, a CR, an unterminated string) | refuses, naming the file, line, column, and token |
| A data module | `'#'` as an entry | lexed as a one-character string, never a comment |
| The built table | sha256 differs from the pin | `ErrTableChecksum`, naming both sums |
| The table given to `Parse` | empty, truncated, trailing bytes, or any byte changed | `ErrTableChecksum` (the pin catches every one) |
| The bytes given to `decode` (synthetic) | truncated mid-block, or trailing bytes | the decoder's own error, naming the offset |
| `Fetch` | status 404 or 500 | refuses, naming the status (the offline guidance, which names `--from`, is the install verb's; Task 3) |
| `Fetch` | body of `MaxWheelSize` plus one byte | `ErrWheelTooLarge`, and `!errors.Is(err, ErrWheelChecksum)` |
| `Fetch` | body of exactly `MaxWheelSize` bytes | not refused for size (it fails the checksum instead) |
| `Fetch` | the context is canceled | `errors.Is(err, context.Canceled)` |
| `Fetch` | a 200 body with the wrong bytes | `ErrWheelChecksum` |
| The cached wheel for a test | absent (`-count=1`, `XDG_CACHE_HOME` pointed at an empty scratch directory) | the test fails naming `make unidecode-wheel`; it never skips |
| The cached wheel for a test | `os.UserCacheDir` returns an error (`HOME` and `XDG_CACHE_HOME` both unset) | the test fails printing the error |
| `harness fetch-unidecode` | the cache verifies | exits 0 with no request (the rewriting transport's count is 0) |
| `harness fetch-unidecode` | the cached file is present but fails its checksum | refetches through the transport and replaces it; the count is 1 |

**Acceptance criteria:**

- `TestBuildIsByteIdenticalToThePin` builds the cached wheel, computes
  `sha256.Sum256` of the output itself, and compares it with the literal
  `c954605b805e4a21d1d9e79b4d93bc6b89684de9996dd21720b256fc58a6ce85`
  written in the test file, never with the package's pin constant. That
  literal is Python's output (`generate.py`), independent of the Go port,
  so a pin edited to match a wrong build still fails. It also counts 190
  blocks and 48,597 entries, so a vacuous build cannot pass. The report
  pastes the three figures and the build's wall time from `go test -v`.
  The diff-reviewer confirms that the source pin equals the same literal.
- Today's `TestUnidecodeTableFullSweep` (Go's lookup against Python's
  digest) is not moved: after this pass, byte identity is the proof that
  replaces it, since a matching file carries Python's digest (R-tt3). A
  sweep that recomputes the digest through the same `Lookup` that wrote it
  would pass under any `Lookup` defect.
- The lookup rows `translit_test.go:21-44` carries today (`A`, U+00E9,
  U+0080, U+0241, U+3042, U+10000, U+F0000) pass against the built table,
  plus one row per `\x0a` entry (U+2028 maps to `"\n"`, U+2029 to `"\n\n"`,
  from `x020.py:42-43`) and one for `x020.py:93` (U+2057 maps to `''''`).
- `make unidecode-wheel` run twice with `XDG_CACHE_HOME` pointed at a fresh
  scratch directory (so a fix-round re-dispatch still starts empty): the
  first fetches, the second makes no request. Then run once without the
  override, filling the real cache for Task 2 onward. The report pastes all
  three outputs and `sha256sum` of each cached file.
- No test in the package opens a connection outside `httptest`; every test
  client carries the rewriting transport.
- `go doc -short ./internal/engine/matcher/unidecode` lists exactly the R-tt2
  names; the report pastes it.
- The gate is green.

**Named mutations this task reports:**

1. Make the lexer keep `\x0a` as four characters: `Build` refuses with
   `ErrTableChecksum`.
2. Make the lexer treat `#` inside a string as a comment: the lexer refuses
   (`x006.py`) or `Build` refuses with `ErrTableChecksum`; the report names
   which.
3. Drop the `< 0x80` passthrough from `Lookup`: `Build` refuses with
   `ErrTableChecksum` (the digest moves).
4. Flip one byte of the cached wheel in a temporary copy: `Build` refuses
   with `ErrWheelChecksum` before any zip read.
5. Flip one byte of a built table: `Parse` refuses with `ErrTableChecksum`.
6. Remove the cap-plus-one read from `Fetch`: the over-cap row fails (it
   asserts `ErrWheelTooLarge`, not merely a refusal).
7. Make `extract` accept 189 data entries: the synthetic count row fails.

**Deliverable count: 4** (the format and lookup; the extraction; the fetch,
cache path, harness verb, and make target; the tests).

---

## Task 2: The matcher seam and the four test binaries

Model: `sonnet`. Depends on Task 1 having filled the workstation's wheel
cache, confirmed by the conductor (conductor action 2). **2.0 hook: neutral** (hook 2 reads
through this seam at rung 9).

**Files:**

- Modify: `internal/engine/matcher/translit.go`, `translit_test.go`
- Delete: `internal/engine/matcher/unidecode-1.4.0.tbl`
- Create: one `TestMain` file each in `internal/engine/matcher`,
  `internal/engine/tag`, `internal/engine/shadow`, `internal/harness/conform`
- Modify: `internal/harness/conform/defaults.go`, `defaults_test.go`

**Interfaces:** consumes `unidecode.Build`, `Parse`, `(*Table).Lookup`,
`Version`, `WheelCachePath`. Produces the matcher setter (Task 3 calls it).
Its name is the implementer's, stated in the report; the plan calls it "the
setter".

**Outcomes:**

- **The seam** (R-tt1). `translit.go` keeps `Transliterate` and
  `Levenshtein`, gains the setter and its race-free holder, and loses the
  embed, `UnidecodeVersion`, the table type, the reader, the parser,
  `loadUnidecodeTable`, and `transliterateRune`. The panic on an uninstalled
  seam names the setter. The setter's doc comment states that a program
  installs it once before scoring and that repeated installs are safe.
- **The table file is gone** from the tree, and no Go file embeds anything
  under `internal/engine/matcher`.
- **The four `TestMain` files** per R-tt5. Each is a few lines calling the
  same sequence; a shared helper is not added, because a helper exported for
  tests alone trips the exported-unused leg.
- **The conformance check.** `CheckUnidecodeVersion` compares the manifest
  with `unidecode.Version`, and its message says the binary pins the
  version. Its two tests keep their shape. Its doc comment keeps the
  architectural reason it lives in the harness.
- **The matcher's tests.** `TestUnidecodeVersion`,
  `TestTransliterateRune`, and `TestUnidecodeTableFullSweep` are removed
  (Task 1's lookup rows carry the first two; byte identity replaces the
  sweep). `TestTransliterate` stays and runs through the installed seam. A
  new serial test proves the uninstalled panic on an ASCII input: it
  clears the holder through package-internal access (`translit_test.go` is
  `package matcher`) and restores the installed function afterwards.

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| A `TestMain` | cached wheel absent (run with `-count=1`, since the test cache does not track the wheel) | prints the reason and `run make unidecode-wheel`; exits nonzero; no test runs |
| A `TestMain` | `os.UserCacheDir` returns an error | prints the error; exits nonzero |
| The setter | given a nil function | panics; no uninstall channel |
| A `TestMain` | cached wheel fails a checksum | prints the sentinel's message; exits nonzero |
| The seam | nothing installed, input `"abc"` | panics naming the setter |
| The seam | installed twice, scored from many goroutines | no race under `go test -race ./internal/engine/matcher` (reported, run once, outside the gate) |

**Acceptance criteria:**

- `go test -count=1 ./internal/engine/matcher ./internal/engine/tag
  ./internal/engine/shadow ./internal/harness/conform` passes; the report
  pastes the four `ok` lines and their wall times beside the same packages'
  times at the base commit.
- `TestWholeCorpusGradesCleanlyInCompatMode`,
  `TestWholeCorpusGradesVerdictsInCompatMode`, and
  `TestComponentTableReplaysAllRows` pass unchanged: no expected figure moves.
- `git grep -n 'go:embed' -- internal/engine/matcher` and `git ls-files
  '*.tbl'` both print nothing.
- `go list -deps ./internal/engine/matcher` (non-test) names no
  `internal/engine/matcher/unidecode`.
- The gate is green.

**Named mutations this task reports:**

1. Remove the matcher's `TestMain`: every test that scores panics, and the
   message names the setter.
2. Point one `TestMain` at a nonexistent cache path and run with
   `-count=1`: that binary exits nonzero naming `make unidecode-wheel`.
3. Make the seam fall back to an ASCII-only passthrough when uninstalled: the
   panic test fails.
4. Compare the manifest with a literal `"1.4.1"` in `CheckUnidecodeVersion`:
   `TestUnidecodeCheckAgreesWithCommittedManifest` fails.

**Deliverable count: 4** (the seam with the table's removal; the four
`TestMain` files; the conformance check; the matcher's tests).

---

## Task 3: The install verb, the scoring verbs' load, and the doctor row

Model: `sonnet`. Depends on Task 2. **2.0 hook: neutral.** Hook 18 (`dubplate
init` reads the registered declarations) is unaffected: this verb reads no
config key, and `init`, when built, composes it as a setup step.

**Files:**

- Create: `cmd/dubplate-server/unidecode.go`, `cmd/dubplate-server/unidecode_test.go`
- Modify: `cmd/dubplate-server/main.go` (register the verb),
  `shadow.go` and `shadow_test.go` (the load), `doctor.go` and
  `doctor_test.go` (the row)

**Interfaces:** consumes Task 1's names and Task 2's setter. Produces the
verb (Task 4's `deploy.sh` step runs it), whose constructor takes the
`*http.Client` from `newRootCmd` (which takes it from `main`), and two
unexported helpers: a load that reads and verifies the table from a state
directory, called by both shadow verbs and the doctor row, and an install
that passes a loaded table's `Lookup` to the setter, called by the shadow
verbs only. The tests carry their own rewriting transport in a `_test.go`
file (a few lines, the same shape as Task 1's, rewriting a
`req.Clone`), never an exported helper.

**Outcomes:**

- **The install verb**, named per Ruling 2, per R-tt6, with `--from FILE`
  for the offline path.
- **The scoring verbs' load** (R-tt4). `shadow observe` and `shadow
  reevaluate` load, verify, and install the table after their existing
  refusals and before `shadow.OpenStore`; a failure refuses with the path,
  the reason, and the install verb. Every existing test Task 0 named in
  notes 7 and 8 (the scoring-verb tests, and every doctor test expecting
  exit 0 with `DUBPLATE_STATE_DIR` set) installs a table into its temporary
  state directory by running the install verb with `--from` the cached
  wheel, never by copying bytes.
- **The doctor row** (named per Ruling 2). It prints exactly one line,
  always, after the backup rows, the store row's precedent. Its states:
  `not checked` when `DUBPLATE_STATE_DIR` is unset (no failure of its own;
  the wall row already fails); `ok (1.4.0)` for a table `Parse` accepts;
  `missing: run dubplate-server <verb>` when the read fails with
  `fs.ErrNotExist` (an absent state directory, leaf, or file); and
  `invalid: <reason>: run dubplate-server <verb>` for every other failure
  (a directory in the file's place, permission denied, a `Parse` refusal).
  `missing` and `invalid` each fail the command's exit, since a missing
  table makes the scoring verbs refuse, as the roster row reasons for its
  own file. The row calls the load helper only, so `doctor` never installs
  a transliterator; it never touches the network and creates nothing under
  the state directory. The command's doc comment and exit rule name the new
  failing row.

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| Install | `DUBPLATE_STATE_DIR` unset | refuses naming the variable; writes nothing |
| Install | state directory absent | refuses naming the path |
| Install | state directory owned by another user | refuses naming the owner and the `runuser` fix (test through a swappable effective-uid seam, `leaf.go`'s `geteuid` precedent) |
| Install | `unidecode/` leaf is a file, has a mode other than `0755`, or is owned by another user | refuses naming the path and the fix; writes nothing |
| Install | state directory owned by another user and a valid table present | refuses (the ownership checks run before the "already installed" check) |
| Install | valid table present | `already installed`, exit 0, no request (the rewriting transport's count is 0) |
| Install | invalid table present | rebuilds and replaces it; the old bytes never survive a success |
| Install | `--from` absent or a directory | refuses naming the path; writes nothing |
| Install | `--from` a file of `MaxWheelSize` plus one byte (a sparse or generated scratch file, never `/dev/zero`) | refuses with `ErrWheelTooLarge` before hashing; writes nothing |
| Install | `--from` the wrong file under the cap | refuses with `ErrWheelChecksum`, naming both sums; writes nothing |
| Install | download refused (status, cap, checksum, context) | refuses with the offline guidance (the wheel's URL, its sha256, and `--from`); an existing file is untouched |
| Install | success | leaf `0755` (set explicitly, so the assertion holds under any umask), file `0644`, `Parse` accepts the file on read-back; the output names the path and the version |
| A scoring verb | table absent, empty, a directory, or failing `Parse` | refuses before `OpenStore`, naming the path, the reason, and the install verb; no shadow leaf is created |
| Doctor | each of the four row states | the pinned status text and the pinned exit; the `missing` and `invalid` exit rows run on `baseEnviron()`'s green wall with every other row `ok`, so the exit moves on this row alone (the exit rule is an OR) |
| Doctor | any state | creates nothing under the state directory (tested); calls the load helper only (reviewer-read, since the scoring-verb tests install process-wide in the same binary, so a test cannot observe it) |
| The `cmd/dubplate-server` test binary | the cached wheel absent (`-count=1`, `XDG_CACHE_HOME` pointed at an empty scratch directory) | the tests that install a table fail naming `make unidecode-wheel`; they never skip |

**Acceptance criteria:**

- The install tests run against `httptest` through the rewriting transport
  and the cached wheel, never the network. The report pastes one
  successful run's output.
- Every doctor test Task 0 note 8 named still exits 0, on an installed
  table.
- `shadow observe` with an absent table exits nonzero and leaves the
  temporary state directory without a `shadow/` leaf; the test asserts both.
- Doctor's tests pin the new row's four states and the command's exit for
  each, and the row count the existing tests pin moves by one.
- `git grep -n '<the setter>(' -- '*.go' ':!*_test.go'` prints exactly one
  line, in `cmd/dubplate-server/unidecode.go`.
- The gate is green.

**Named mutations this task reports:**

1. Move the load after `shadow.OpenStore`: the "no shadow leaf" assertion
   fails.
2. Make the install verb skip `Parse` on read-back and write a table with one
   flipped byte: the success row fails.
3. Make the doctor row pass on a missing table: its exit assertion fails.
4. Drop the owner check: the ownership row fails.
5. Make the idempotent path fetch anyway: the request-count assertion fails.
6. Drop the explicit `os.Chmod` on the leaf: the success row fails, since
   it runs under `syscall.Umask(0o027)` set in that serial test and
   restored afterwards.

**Deliverable count: 4** (the verb; the scoring verbs' load with the test
updates; the doctor row; the verb's tests).

---

## Segment boundary

The conductor writes STATUS here (the task ledger, the spend, the
checkpoint), on a commit the gate proved green. Task 4 touches only
`scripts/` and `tools/`, disjoint from Tasks 1 to 3.

---

## Task 4: The staged deploy transition and the generator's new role

Model: `sonnet`. Depends on Task 3 (the verb the step runs). **2.0 hook:
neutral.** The implementer invokes `vps-conventions` first.

**Files:**

- Modify: `scripts/deploy.sh`, `tools/unidecode-table/generate.py`

**Why the transition is staged.** `dubplate-shadow.timer` is live on the
box and fires `shadow observe` every five minutes; `deploy.sh` never stops
it (`TIMER_UNITS`, `:526`). Revision 1 swapped the binary at `:343` and
installed the table later, so every firing in between ran a binary with no
table, and a failed install (PyPI unreachable, a checksum refusal, a
sandbox error) aborted the deploy with that binary live and the shadow
failing every five minutes, unmonitored ("no monitoring at rung 4", the
unit's own comment). Staging closes the window: the live binary is replaced
only after the new one has installed and verified its own table.

**Outcomes:**

- **The deploy transition**, in this order (line numbers at `6777d43`):
  1. **The pre-create moves ahead of the binary step.** The
     `/var/lib/dubplate` block (`:489-505`, with its `DUBPLATE_LIB_DIR`
     definition at `:502`) moves unchanged to just before the binary step
     (`:304`). It is idempotent, and the state-dir agreement leg reads
     `DUBPLATE_LIB_DIR` by pattern, so it needs no new row. The block's
     comment gains one sentence: the staged install needs the owned state
     directory before it runs.
  2. **The binary is staged, not swapped.** The apply-mode rsync at
     `:341-344` targets `/usr/local/bin/dubplate-server.new` (one variable
     holding the staged path) and `chmod`s it; the smoke test at
     `:364-366` runs the staged path. The live binary is untouched. Under
     `--check`, the `rsync -n` keeps the live path as its target, since an
     `-n` run against the staged path always reports a full transfer and
     would hide whether the binary changed; its comment says so.
  3. **The staged binary installs its table**, in a new block directly
     after the smoke test (`:366`) and before the env step (`:409-414`),
     as `music`, with this exact shape through `run_remote`
     (`$DUBPLATE_LIB_DIR` expands on the workstation, inside the string
     `run_remote` sends):

     ```
     systemd-run --wait --pipe --collect \
       -p User=music -p Environment=DUBPLATE_STATE_DIR=$DUBPLATE_LIB_DIR \
       -p NoNewPrivileges=yes -p PrivateTmp=yes -p ProtectSystem=strict \
       -p ProtectHome=yes -p ProtectKernelTunables=yes \
       -p ProtectKernelModules=yes -p ProtectControlGroups=yes \
       -p RestrictNamespaces=yes -p RestrictRealtime=yes \
       -p RestrictSUIDSGID=yes -p LockPersonality=yes \
       -p MemoryDenyWriteExecute=yes -p SystemCallFilter=@system-service \
       -p SystemCallArchitectures=native -p AmbientCapabilities= \
       -p CapabilityBoundingSet= \
       -p 'RestrictAddressFamilies=AF_INET AF_INET6 AF_UNIX' \
       -p ReadWritePaths=$DUBPLATE_LIB_DIR \
       /usr/local/bin/dubplate-server.new <verb>
     ```

     The verb reads `DUBPLATE_STATE_DIR` alone through `lookupEnviron`
     (R-tt6), and the state-dir agreement leg (`check.sh:645-658`) pins
     that variable to equal `DUBPLATE_LIB_DIR`. So the one `Environment=`
     property carries exactly what the env file would, carries no secret,
     and does not depend on `server.env` existing, which a fresh box lacks
     until the env step. The hardening is the shadow unit's, property for
     property, with four deliberate differences. `Environment=` replaces
     the unit's `EnvironmentFile=`, for the reason just given.
     `ReadWritePaths=$DUBPLATE_LIB_DIR` replaces the unit's `.../shadow`
     grant, so the verb can create its `unidecode/` leaf; the grant reuses
     the existing variable, adding no state-directory literal.
     `RestrictAddressFamilies` keeps `AF_INET` and `AF_INET6`, so the
     download reaches PyPI (the shadow already reaches MusicBrainz under
     the same set). `UMask=0027` is omitted, and the verb sets the leaf's
     mode explicitly regardless. `--wait` propagates the verb's exit
     (quoted above), so a refusal fails `run_remote` and `set -euo
     pipefail` aborts the deploy. That abort holds only because the new
     `run_remote` calls (the install and the `mv`) are plain top-level
     commands: `set -e` is ignored inside an `if` condition or on the
     left of `||` or `&&`, so neither call may sit in one. The quoting of
     the one space-bearing property survives `run_remote`'s single-string
     `ssh`.
  4. **Only then is the binary swapped**, directly after the install:
     `mv -f /usr/local/bin/dubplate-server.new /usr/local/bin/dubplate-server`,
     a rename on one filesystem, so a timer firing sees either the old
     binary or the new one, and the new one finds its table.
  5. **`server.env` installs as today** (`:409-414`), after the swap. This
     keeps today's binary-then-env order, so a new env never runs under an
     old binary.

  Under `--check`, every step reports its would-run line through
  `run_remote`, and the rsync runs with `-n` against the live path (step
  2). On a box with a verified table, step 3 makes no network request
  (R-tt6), so every later deploy is quiet.
- **Failure and rollback.** A failed install aborts the deploy before the
  swap and before the env step: the old binary stays live on the old
  `server.env`, and the shadow keeps scoring from its embedded table. (Had
  the env step come first, a deploy that adds a `DUBPLATE_*` key, as 5d
  adds the magic-link keys, would leave the old binary refusing that key
  on every firing: the wall rejects an unknown `DUBPLATE_*` name,
  `internal/config/config.go:287`.) The staged `.new` file is left behind,
  harmless, and the next deploy's rsync overwrites it. Rollback after a
  completed deploy is redeploying the previous commit: its `deploy.sh`
  rsyncs its own binary over the live one, and that pre-pass binary scores
  from its own embedded table and ignores `unidecode/`. No state changes,
  and a later forward deploy finds the table installed and makes no
  request.
- **The generator's role** per R-tt7: the docstring's new purpose, the
  pin-update procedure (the two pins and the data-module count), the
  never-commit rule, and the printed output sha256.

**Acceptance criteria:**

- shellcheck passes (the gate's leg).
- `bash scripts/deploy.sh --check` is **not** run (it reaches the box). The
  diff-reviewer reads the transition against this task's five steps: the
  pre-create precedes the binary step; the apply-mode rsync and the smoke
  test name only the staged path, and the `--check` rsync names the live
  path; the install directly follows the smoke test, and the `mv` directly
  follows the install, both before the `server.env` install; its property
  list matches the shadow unit's hardening except the four named
  differences, sets `Environment=DUBPLATE_STATE_DIR=$DUBPLATE_LIB_DIR`
  and no `EnvironmentFile=`, grants write on `$DUBPLATE_LIB_DIR`, and
  reaches the network; neither new `run_remote` call sits inside an `if`
  or on either side of `||` or `&&`; and the `mv` is the only write to the
  live path.
- `ruff check` and `ruff format --check` run directly on `generate.py`
  pass; the report pastes both.
- `tools/venv/.venv/bin/python tools/unidecode-table/generate.py
  <scratch path>` prints a sha256 equal to the pin; the report pastes it.
  The scratch file is deleted, never staged.
- The gate is green.

**Named mutations this task reports:**

1. Reviewer-read: the install (or the `mv`) placed after the
   `server.env` install would, on a failed install, leave the old binary
   live on a new env that may carry a key it refuses; the `mv` placed
   before the install would reopen the tableless window. The reviewer
   confirms the order: smoke test, install, `mv`, env step.
2. Reviewer-read: the install carried with the shadow unit's
   `ReadWritePaths=/var/lib/dubplate/shadow` would fail with EROFS creating
   `unidecode/`; the reviewer confirms the grant is `$DUBPLATE_LIB_DIR`.
3. Run `generate.py` against a scratch venv copy with the version check
   left as is: no mutation is available without a second Unidecode release
   in the venv, so this is reviewer-read.

**Deliverable count: 2.**

---

## Task 5: The pass close drafts

Model: `sonnet`. Depends on Tasks 0 to 4 and the simplifier round.

**Files (drafts only; the conductor folds and commits):**

- Create: `docs/superpowers/research/handoffs/draft-HISTORY-translit-table.md`
- Create: `docs/superpowers/research/handoffs/draft-STATUS-translit-table.md`

**Outcomes:**

- **The HISTORY draft** carries what landed per task with commit ids, the
  gate evidence, R-tt1 to R-tt7 with where each landed, the rulings'
  outcomes, both budgets, and "what a later pass would be wrong to
  rediscover." That section carries at minimum: the measured grammar and
  the `'#'` string entries; the Python and Go prototypes' byte-identity
  figures; why the digest in a Go build still proves Python's lookup
  (R-tt3); the pin-update procedure (R-tt7); the four test binaries and the
  overlay probe that found them; that Go's test cache does not track the
  wheel; the `make unidecode-wheel` prerequisite; the staged deploy
  transition and why (the live shadow timer); that `serve` loads the table
  when it first scores; the process-wide holder's tension for a published
  library (R-tt1); and, routed to `give-back`, the git-history question and
  the list of Unidecode-derived items the repo still carries ("Owner
  touchpoints").
- **The STATUS draft** adds a carried item for 5d's runbook: run
  `deploy.sh --check` first and read the staged install's would-run line;
  the deploy stages, installs, and only then swaps; rollback is redeploying
  the previous commit; and the workstation's verified wheel cache
  (`make unidecode-wheel`'s output) is the estate's offline copy, the
  `--from` source if PyPI ever deletes the release. It also names the
  workstation prerequisite, and records nothing else unless the pass found
  something.

**Acceptance criteria:** every count in the drafts comes from a command the
report pastes; every commit id resolves; the STATUS draft is at most 60
lines. The gate is green as a formality.

**Deliverable count: 1** (two drafts, one close).

---

## The exit gate

`cairn-run-gate 'bash /var/home/glw907/Projects/dubplate/scripts/check.sh'`
green on master after the last commit, plus:

1. The Go build of the pinned wheel equals the literal `c954...ce85`
   (Task 1), on a `-count=1` run.
2. `git ls-files '*.tbl'` prints nothing, and `git grep -n 'go:embed' --
   internal/engine/matcher` prints nothing (Task 2). The scope is the
   matcher: `internal/engine/store/migrations.go` embeds its migrations
   legitimately.
3. `strings dubplate-server-linux-amd64 | grep -c 'Kayng '` (a `x0f9.py`
   entry) run by the conductor on two `make build-linux` binaries: at the
   pass's base commit it prints a nonzero count (the positive control), and
   at HEAD it prints 0 (and `grep -c` exits 1, which is the expected
   pass).
4. The setter has exactly one non-test call site (Task 3).
5. The corpus sweeps pass with no expected figure moved (Task 2).
6. `go doc -short ./internal/engine/matcher` at the base and at HEAD differ
   only by the removed `UnidecodeVersion` and the added setter.

---

## Contested surfaces

| Surface | Owner | This pass's position | Contention |
| --- | --- | --- | --- |
| `internal/engine/matcher` | rung 3 | Tasks 1 and 2 | none; no open lane touches it |
| `internal/engine/tag`, `internal/engine/shadow` | rungs 3, 4 | one `TestMain` file each (Task 2) | none |
| `internal/harness/conform` | rung 3/4 | `defaults.go` and a `TestMain` (Task 2) | main tree only |
| `cmd/dubplate-server` | rungs 4 to 6 | Task 3 | lane 1 (5c-iii: `web.go`, `serve.go`, rendered tests) and lane 2 (6b-ii-b restore); Task 0 blocker 1 |
| `cmd/harness`, `Makefile` | rung 2 | Task 1 | `Makefile` also carries lane 1's web targets; Task 1 touches its own target only |
| `scripts/deploy.sh` | rung 4 and 5d | Task 4 (moves the pre-create, stages the binary step, adds the install block) | 5d owes its own runbook rows; this transition joins them |
| `docs/` ledgers, `CLAUDE.md` | conductor | Task 5 drafts only | every lane |

---

## The conductor's actions

Not tasks; no implementer performs them.

1. Before Task 0: record the rulings' outcomes; confirm the lanes per
   Task 0 blocker 1; record the planning spend; write the lane registry row;
   arm the runaway guard; check AC power.
2. Before Task 2: confirm from Task 1's report that `make
   unidecode-wheel` filled the real workstation cache (a network act
   outside the gate), and confirm `sha256sum` of the cached file.
3. At the segment boundary: STATUS; re-derive the ceiling from Tasks 0 to
   3's reports.
4. After Task 4: one `code-simplifier` round over Tasks 1 to 3's Go,
   pasting `.claude/instructions/simplifier-brief.md` verbatim, with one
   `diff-reviewer` over its diff.
5. After Task 5: fold the drafts; run exit-gate check 3; then one
   `go-architecture-reader` each for `internal/engine/matcher`,
   `internal/engine/matcher/unidecode`, and `cmd/dubplate-server`, with no
   plan in context.
6. At the close: update CLAUDE.md's two tracked-Python descriptions of
   `generate.py` (the "Gate" list and "Stack rules: Python") to its new role,
   and add a line under "Gate" naming `make unidecode-wheel` as a
   workstation prerequisite beside `npm ci`; route the git-history question
   the lawyer's confirmation (with the "Owner touchpoints" list of
   Unidecode-derived items), and R-tt1's holder question to ROADMAP
   `give-back`.
7. **Do not deploy.** The next deploy (5d) carries Task 4's transition.

---

## Execution model

The pass-execute workflow, invoked by name, in two segments: Task 0 (read
directly by the conductor), then segment A (Tasks 1 to 3), the boundary,
then segment B (Task 4, the simplifier round, Task 5). Each task runs
`dubplate-implementer` (`sonnet`, `high`), then `diff-reviewer`
(`claude-opus-5-5`, `high`), then the gate. One re-dispatch on `fix`; a
second `fix` is the conductor's decision. Every reviewer dispatch asks for
correctness gaps ranked by consequence, not polish, and every fold may
refuse a finding whose fix costs more than the risk it removes, with a
one-line reason.

Dependencies: Task 0, then Task 1 (the package), then Task 2 (the seam reads
it), then Task 3 (the verb installs through the seam), then Task 4 (the
deploy transition runs the verb), then Task 5. No task is independent.

**Budget:** the ceiling and flag in "Token ceiling" (Ruling 1), planning
recorded separately, STATUS at the boundary and at the close, spend scored
against the ceiling and against zero attended interactions after approval.

---

## Prior art consulted

Scored on CLAUDE.md's four criteria where a dependency is in question. The
2.0 spec's prior-art list and forward-hooks list
(`docs/superpowers/specs/2026-09-07-streamer-2-0-design.md`, "Forward hooks
the current build owes" and "Prior art") were read: hook 2 reuses the table
through this seam at rung 9, hook 18's `init` composes the verb, and every
task is neutral.

| Entry | Verdict for this pass |
| --- | --- |
| `archive/zip`, `crypto/sha256`, `net/http`, `io`, stdlib | **Adopt.** Store and Deflate are built in; no dependency is needed. |
| `time/tzdata` | **Adopt as the design precedent** for R-tt1: the program, not the library, decides where the table comes from. |
| Python Unidecode 1.4.0 wheel on PyPI | **Adopt as the one pinned input**, by sha256; GPL-2.0-or-later, downloaded at setup, never distributed. |
| Unidecode 1.4.0 sdist | **Decline** (R-tt3). |
| A Go Python-literal parser (for example a `go-python` tokenizer) | **Decline**: the grammar is six token kinds, measured whole; a dependency for it fails the house bar on proportion. |
| `mozillazg/go-unidecode`, `rainycape/unidecode`, `gosimple/unidecode`, AnyAscii, ICU/CLDR, `text-unidecode` | **Decline**, on the survey's evidence: none is bit-exact with Unidecode 1.4.0. AnyAscii stays the survey's candidate permissive default for the published module, not this pass. |
| `generate.py` | **Adopt as the pin's oracle** (R-tt7). |
| `internal/engine/store/leaf.go`, the shadow leaf, `internal/atomicfile` | **Adopt as in-repo precedent** for the owner refusal, the mode tier, and the atomic write. |
| Navidrome, gonic, taglib, `mewkiz/flac`, Picard, chromaprint, goose, `modernc.org/sqlite` | No surface in this pass. |

---

## Not verified in this revision

- The Go port as it will ship. Two prototypes, one in Python and one in Go
  (the mechanics review's, stdlib only, outside the tree), both built a
  table byte-identical to the committed one under the same restricted lexer
  and `_get_repl_str` rules; Task 1 is the proof for the shipped code.
- The shipped build's per-binary cost (the Go prototype measured about
  0.14s; R-tt5 keeps a 0.5s report threshold).
- The staged install run end to end on the box. A read-only `curl -sI` from
  the box to the wheel URL returned `HTTP/2 200` (mechanics review), and the
  shadow already reaches MusicBrainz under the same address families, so
  TLS and DNS work. The box's systemd 257 accepting every `-p` property is
  proved by the 5d `--check` (which prints the would-run line) and the
  first deploy; each property predates 257 by its man page. The first
  deploy after this pass is the proof, and under the staged transition a
  refusal there leaves the old binary scoring on the old env.
- Whether lane 2's 6b-ii-b edits `doctor.go` or `main.go` (Task 0).
- The legal reading itself: whether downloading and deriving at setup keeps
  an MIT binary clear of GPL-2.0-or-later obligations. The lawyer's
  confirmation is owed before the first stranger-installable release, over
  the list in "Owner touchpoints".
- Whether PyPI keeps serving the 1.4.0 wheel. The file URL is
  content-addressed and marked immutable, but `immutable` does not bind the
  project owner, who can delete the release. The workstation's verified
  cache is the estate's offline copy for `--from` (Task 5's 5d runbook
  row); where a stranger gets the wheel after a deletion is `packaging`'s
  question.
