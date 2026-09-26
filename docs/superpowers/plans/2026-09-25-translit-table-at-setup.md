# The transliteration table leaves the binary: implementation plan

Status: **draft, revision 1, 2026-09-25; unreviewed; two rulings open.**
The plan-review lenses (`spec-plan-review`) have not run. The plan executes
under either answer to each ruling.

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
`deploy.sh` runs the install step so the live shadow keeps scoring after the
next deploy.

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

**1. Does this pass run under a 3.1M execution ceiling?** Recommendation:
yes. Reason: "Token ceiling" projects 2.35M from the matcher word-boundary
pass's measured per-task actuals, scaled for a new package, a new verb in
`cmd/dubplate-server`, and three merge-side architecture reads. A 3.1M
ceiling puts the 80% flag at 2.48M, above the projection. Planning is
recorded outside the ceiling.

What each answer builds: yes sets the flag at 2.48M. No holds a 2.75M
ceiling (flag 2.2M), which the projection crosses during the merge-side
reads, so the conductor asks after Task 5, before those reads.

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

- **The lawyer's confirmation** of the distribution reading (the binary
  downloads and derives at setup; the repo and binary carry no Unidecode
  data) before the first stranger-installable release. `give-back` and
  `packaging` carry it.
- **Git history.** The `.tbl` stays in every commit before this pass's
  Task 2. The repo is private today (`gh repo view`, per the survey). Whether
  history must be rewritten before the repo goes public is a question for
  the same lawyer's read; the photo lane carries the same shape of question
  for withdrawn images (STATUS, "Open items"). Task 5's HISTORY draft names
  it, and the conductor routes it to `give-back`.

What would become one: a byte mismatch the Go port cannot close (Task 1
stops), a module dependency, or any change to what the box runs beyond the
one `deploy.sh` step Task 4 adds.

---

## Token ceiling: 3.1M execution (Ruling 1)

The basis is the matcher word-boundary pass's recorded actuals
(`docs/HISTORY.md`, that entry's spend line): Task 0 0.08M; Task 1 0.14M
plus review 0.06M; Task 2 0.27M plus reviews 0.18M; Task 3 0.20M plus review
0.07M; the simplifier 0.11M plus review 0.04M; lane total about 1.25M. The
6b-ii-a pass (about 1.95M for three tasks, including merge-side reads) is
the second reference.

1. Task 0, a staleness read at `sonnet`: **0.08M**.
2. Task 1, the `unidecode` package and the wheel fetch: **0.50M**, above
   the matcher pass's largest task, since it carries a lexer, a byte-identity
   proof, and a network function with its refusals.
3. Task 2, the matcher seam and the four test binaries: **0.35M**.
4. Task 3, the install verb, the scoring verbs' load, and the doctor row:
   **0.45M**; every full gate is heavy (Chromium) now.
5. Task 4, `deploy.sh` and the generator's new role: **0.15M**.
6. Task 5, the close drafts: **0.12M**.
7. One fix round: **0.20M**.
8. The simplifier round with its one review: **0.15M**.
9. The merge-side reads: three `go-architecture-reader` dispatches
   (`internal/engine/matcher`, `internal/engine/matcher/unidecode`,
   `cmd/dubplate-server`) at **0.30M**, and the merge ritual at **0.05M**.
10. **The projection is 2.35M.**

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
5. **`internal/harness/conform/defaults.go:81-93`** (`CheckUnidecodeVersion`,
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
10. **`scripts/deploy.sh`**: the binary step (`:337-345`), the smoke test
    (`:364-366`), the `/var/lib/dubplate` pre-create (`:490-504`), and
    `server.env`'s `0600 root:root` install (`:409-414`);
    `systemd/dubplate-shadow.service` (`User=music`,
    `ReadWritePaths=/var/lib/dubplate/shadow`, and its operator rule: a
    hand-run verb goes through `systemd-run` carrying the unit's properties,
    never the bare binary as root).

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

`go doc crypto/sha256.Sum256`: "Sum256 returns the SHA256 checksum of the
data."

`go doc net/http.Client.Do`: "A non-2xx status code doesn't cause an
error." And: "If the returned error is nil, the Response will contain a
non-nil Body which the user is expected to close." `go doc net/http.Client`
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
  (`x020.py:93`, `"''''"`). No string carries a prefix or triple quotes, and
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
four has a `TestMain` today (`git grep -l 'func TestMain'`).

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
  the dispatch's footer. **The workstation's wheel cache must be filled
  (`make unidecode-wheel`) before Task 2's gate**, since Task 2 is the first
  task whose gate needs it; Task 1 lands the target and runs it.

---

## Privileged fields and their channels

| Field | Channels that may set it | What refuses every other channel |
| --- | --- | --- |
| The matcher's installed transliterator | `cmd/dubplate-server`'s scoring-verb startup, after `unidecode.Parse` verified the table; each of the four `TestMain` files, after `unidecode.Build` verified the cached wheel and the table. | No type refuses it, since the setter is exported for the cmd caller. The exit gate's `git grep` finds exactly one non-test call site. A scoring call before any installation panics, naming the setter (R-tt3). |
| The table file in the leaf | The install verb alone: it writes atomically, only bytes whose wheel and table checksums both matched. | Every load runs `unidecode.Parse`, which refuses bytes whose sha256 differs from the pin, so a hand-copied, truncated, or edited file is refused at startup and by `doctor`. |
| The wheel and table pins (version, wheel file name, URL, wheel sha256, table sha256) | Source edits to the `unidecode` package, re-derived by `generate.py` under the pinned venv (R-tt7). | `TestBuildIsByteIdenticalToThePin` (Task 1) fails when the Go build of the pinned wheel differs from the table pin; `CheckUnidecodeVersion` fails when the version differs from `corpus/manifest.json`. |
| The workstation wheel cache | `harness fetch-unidecode` (`make unidecode-wheel`), which writes only bytes whose sha256 matched. | Every test read runs `unidecode.Build`, which verifies the wheel's sha256 first, so a hand-placed wrong file reds the gate with the checksum named. |

---

## Resolutions this plan takes rather than discovers

A task landing the surface a resolution decides writes nothing about the
resolution in a shipped comment; the close maps each resolution to its
landing commit.

### R-tt1: the seam is process-wide, installed by the main package

Package `matcher` gains one exported setter taking a `func(rune) (string,
bool)` and keeps an unexported, race-free holder for it. `Transliterate`
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
| `Fetch(ctx, *http.Client) ([]byte, error)` | downloads the pinned wheel and verifies it | the install verb; `harness fetch-unidecode` |
| `Build(wheel []byte) ([]byte, error)` | verifies the wheel, extracts, builds, verifies the table | the install verb |
| `Parse(table []byte) (*Table, error)` | verifies the table's sha256, decodes | the scoring verbs' startup; the doctor row |
| `(*Table).Lookup(r rune) (string, bool)` | `_get_repl_str` | installed through the matcher setter in cmd |
| `WheelCachePath() (string, error)` | the workstation cache file path | `harness fetch-unidecode` |
| `ErrWheelChecksum`, `ErrTableChecksum` | sentinels | the install verb and doctor row match them with `errors.Is` |

The table's sha256 pin stays unexported: only `Build` and `Parse` read it.
The cache path sits under `os.UserCacheDir()` in a subdirectory named for
the package, never for dubplate, so the published module keeps it.

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
lookup `Parse` returns, with `generate.py`'s encoding; the table written in
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

Every test binary rebuilds the table from the wheel, so every gate run
re-proves the port byte-identical. Task 1 measures the build's cost; above
0.5s per binary, the task reports it, and the conductor decides whether the
three downstream binaries read a cached table file instead.

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
be rendered. It refuses, writing nothing, when the variable is unset, the
directory is absent, or the effective user does not own it; the last
refusal names the owner and the `runuser -u <owner> -- dubplate-server
<verb>` fix in `internal/engine/store/leaf.go`'s own words, so a root run
never leaves a root-owned leaf. When the file already exists and `Parse`
accepts it, the verb prints that the table is installed and exits 0 with no
network. Otherwise it obtains the wheel (`Fetch`, or the `--from` file),
runs `Build`, creates the leaf, writes through `internal/atomicfile`, reads
the file back through `Parse`, and prints the path and version. A failed
download or build writes nothing and leaves any existing file in place.
The download uses a client with a bounded timeout and a size cap well above
235,837 bytes (1 MiB); an over-cap body is refused.

The verb's offline guidance (its long help, and the text of every download
refusal) states the wheel's URL, its sha256, and `--from`.

### R-tt7: `generate.py` becomes the pin's oracle

The Go port replaces `generate.py` at runtime. Python stays in the loop only
where it is independent: re-deriving the pin. On a Unidecode bump, a
maintainer bumps the venv, runs `generate.py` to a scratch path, and takes
the file's sha256 as the new table pin; the Go build of the new wheel must
then match it. `generate.py`'s docstring says so, names `unidecode.Version`
and the two sha256 pins as what the run updates, states that its output
must never be committed, and drops the committed-path example. Its output
path stays an argument. It prints the output's sha256 beside its existing
figures. The venv stays the source of truth for beets and its dependencies;
the Unidecode version the venv pins and `Version` stay equal, which
`CheckUnidecodeVersion` and `oracle.reconcile` already enforce through
`corpus/manifest.json`.

---

## Review focus

1. **Byte identity.** Expected: the Go build of the pinned wheel equals the
   pinned sha256 on every gate run. Pinned in Task 1, re-proved by every
   `TestMain` in Task 2.
2. **A scoring verb that runs without a verified table.** Expected: never;
   it refuses before creating anything, and an uninstalled seam panics on
   the first call of any input. Pinned in Tasks 2 and 3.
3. **A table file that is present but wrong** (truncated, hand-edited, an
   older version, empty, a directory). Expected: `Parse` refuses it, the
   scoring verb refuses, `doctor` fails the row, and the install verb
   replaces it. Pinned in Task 3.
4. **A network failure or a wrong download.** Expected: nothing is written,
   an existing good table stays, and the refusal carries the offline
   guidance. Pinned in Tasks 1 and 3.
5. **The gate touching the network.** Expected: never. The diff-reviewer
   checks each task's tests for an `http` call outside an `httptest`
   server; Task 1's `Fetch` tests run against `httptest` only.

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
   is the release still unyanked? (A read-only fetch of the JSON; no
   download into the tree.)
5. Does the overlay probe (Primary inputs, "The test binaries that reach
   the table") still name exactly the four packages? Run it with the overlay
   under the scratchpad, never a tree edit.

**Notes:**

6. Report which `cmd/dubplate-server` verbs reach `shadow.Run` or
   `shadow.Reevaluate`, and confirm `shadow report` reaches neither.
7. Report every `cmd/dubplate-server` test that runs `shadow observe` or
   `shadow reevaluate` past their configuration refusals; Task 3 must give
   each one an installed table.
8. Report the doctor tests that pin the row count or the last row's name.
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
  most the cap plus one byte and refuses an over-cap body, and refuses a
  checksum mismatch with `ErrWheelChecksum`, naming both sums.
  `WheelCachePath` per R-tt5. `harness fetch-unidecode` is a no-op when the
  cached file verifies, else fetches and writes atomically; `make
  unidecode-wheel` runs it. The target's comment says the gate never runs it
  and why.
- **The tests.** Lexer rows over synthetic sources; `Parse` and `Lookup`
  rows; `Fetch` over `httptest` only; the byte-identity test and the sweep
  test over the cached wheel.

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| The wheel given to `Build` | sha256 differs from the pin | `ErrWheelChecksum`, naming both sums; nothing unzipped |
| The wheel | not a zip, or an entry fails CRC or size | the zip error, naming the entry |
| The data entries | fewer or more than 190, or a name outside the pattern that starts `unidecode/x` | refuses, naming the count or the name |
| A data module | a token outside the measured grammar (a prefix, triple quotes, an unknown escape, a missing comma, text after `)`, a non-ASCII byte, a CR, an unterminated string) | refuses, naming the file, line, column, and token |
| A data module | `'#'` as an entry | lexed as a one-character string, never a comment |
| The built table | sha256 differs from the pin | `ErrTableChecksum`, naming both sums |
| The table given to `Parse` | empty, truncated, trailing bytes, or any byte changed | `ErrTableChecksum` first (the pin catches every one); the decoder's own errors stay for a table that matches the pin, which cannot be malformed |
| `Fetch` | status 404 or 500 | refuses, naming the status and the offline guidance |
| `Fetch` | body of cap plus one byte | refuses, naming the cap |
| `Fetch` | the context is canceled | returns the context's error |
| `Fetch` | a 200 body with the wrong bytes | `ErrWheelChecksum` |
| The cached wheel for a test | absent | the test fails naming `make unidecode-wheel`; it never skips |
| `harness fetch-unidecode` | the cache verifies | exits 0 with no request (proved with an `httptest` server that counts requests, pointed at by a test-only client) |

**Acceptance criteria:**

- `TestBuildIsByteIdenticalToThePin` builds the cached wheel and compares
  the output's sha256 with
  `c954605b805e4a21d1d9e79b4d93bc6b89684de9996dd21720b256fc58a6ce85`; it
  also counts 190 blocks and 48,597 entries, so a vacuous build cannot pass.
  The report pastes the three figures and the build's wall time.
- The sweep test moves here: it parses the built table and recomputes the
  digest over the header's range through `Lookup`, matching the header.
- The lookup rows `translit_test.go:21-44` carries today (`A`, U+00E9,
  U+0080, U+0241, U+3042, U+10000, U+F0000) pass against the built table,
  plus one row per `\x0a` entry (U+2028 maps to `"\n"`, U+2029 to `"\n\n"`,
  from `x020.py:42-43`) and one for `x020.py:93` (U+2057 maps to `''''`).
- `make unidecode-wheel` run twice: the first fetches, the second makes no
  request; the report pastes both outputs and `sha256sum` of the cached file.
- No test in the package opens a connection outside `httptest`.
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
6. Remove the cap-plus-one read from `Fetch`: the over-cap row fails.

**Deliverable count: 4** (the format and lookup; the extraction; the fetch,
cache path, harness verb, and make target; the tests).

---

## Task 2: The matcher seam and the four test binaries

Model: `sonnet`. Depends on Task 1 and on the conductor having run `make
unidecode-wheel` on the workstation. **2.0 hook: neutral** (hook 2 reads
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
- **The matcher's tests.** `TestUnidecodeVersion` and
  `TestUnidecodeTableFullSweep` are removed (Task 1 carries both proofs).
  `TestTransliterate` stays and runs through the installed seam. A new test
  proves the uninstalled panic on an ASCII input, restoring the installed
  function afterwards.

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| A `TestMain` | cached wheel absent | prints the reason and `run make unidecode-wheel`; exits nonzero; no test runs |
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
2. Point one `TestMain` at a nonexistent cache path: that binary exits
   nonzero naming `make unidecode-wheel`.
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
verb (Task 4's `deploy.sh` step runs it) and one unexported helper that
loads and installs the table from a state directory, which both shadow
verbs and the doctor row call.

**Outcomes:**

- **The install verb**, named per Ruling 2, per R-tt6, with `--from FILE`
  for the offline path.
- **The scoring verbs' load** (R-tt4). `shadow observe` and `shadow
  reevaluate` load, verify, and install the table after their existing
  refusals and before `shadow.OpenStore`; a failure refuses with the path,
  the reason, and the install verb. Every existing test Task 0 named in note
  7 installs a table into its temporary state directory by running the
  install verb with `--from` the cached wheel, never by copying bytes.
- **The doctor row** (named per Ruling 2). It prints exactly one line,
  always, after the backup rows, the store row's precedent. Its states:
  `not checked` when `DUBPLATE_STATE_DIR` is unset (no failure of its own;
  the wall row already fails); `ok (1.4.0)` for a table `Parse` accepts;
  `missing: run dubplate-server <verb>` and `invalid: <reason>: run
  dubplate-server <verb>` otherwise, each failing the command's exit, since
  a missing table makes the scoring verbs refuse, as the roster row reasons
  for its own file. The row never touches the network. The command's doc
  comment and exit rule name the new failing row.

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| Install | `DUBPLATE_STATE_DIR` unset | refuses naming the variable; writes nothing |
| Install | state directory absent | refuses naming the path |
| Install | state directory owned by another user | refuses naming the owner and the `runuser` fix (test through a swappable effective-uid seam, `leaf.go`'s `geteuid` precedent) |
| Install | valid table present | `already installed`, exit 0, no request (proved with a counting `httptest` server) |
| Install | invalid table present | rebuilds and replaces it; the old bytes never survive a success |
| Install | `--from` absent, a directory, or the wrong file | refuses naming the path and the sha256 mismatch; writes nothing |
| Install | download refused (status, cap, checksum, context) | refuses with the offline guidance; an existing file is untouched |
| Install | success | leaf `0755`, file `0644`, `Parse` accepts the file on read-back; the output names the path and the version |
| A scoring verb | table absent, empty, a directory, or failing `Parse` | refuses before `OpenStore`, naming the path, the reason, and the install verb; no shadow leaf is created |
| Doctor | each of the four row states | the pinned status text and the pinned exit |

**Acceptance criteria:**

- The install tests run against `httptest` and the cached wheel, never the
  network. The report pastes one successful run's output.
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

**Deliverable count: 4** (the verb; the scoring verbs' load with the test
updates; the doctor row; the verb's tests).

---

## Segment boundary

The conductor writes STATUS here (the task ledger, the spend, the
checkpoint), on a commit the gate proved green. Task 4 touches only
`scripts/` and `tools/`, disjoint from Tasks 1 to 3.

---

## Task 4: The deploy step and the generator's new role

Model: `sonnet`. Depends on Task 3 (the verb the step runs). **2.0 hook:
neutral.**

**Files:**

- Modify: `scripts/deploy.sh`, `tools/unidecode-table/generate.py`

**Outcomes:**

- **The deploy step.** After the `/var/lib/dubplate` pre-create and the
  binary's smoke test, `deploy.sh` runs the install verb on the box as the
  `music` user with the server env file's `DUBPLATE_STATE_DIR`, never as
  root and never with a new state-directory literal (so the state-dir
  agreement leg needs no new row). The shadow unit's operator rule suggests
  a `systemd-run` carrying `User=music` and
  `EnvironmentFile=/etc/dubplate/server.env`; the implementer chooses under
  `vps-conventions` and states why. Under `--check` it reports the would-run
  line through `run_remote`, the script's own shape. A refusal aborts the
  deploy with the verb's message, before the timers are enabled. On a box
  with a verified table, the step makes no network request (R-tt6), so
  every later deploy is quiet.
- **The generator's role** per R-tt7: the docstring's new purpose, the
  pin-update procedure, the never-commit rule, and the printed output
  sha256.

**Acceptance criteria:**

- shellcheck passes (the gate's leg).
- `bash scripts/deploy.sh --check` is **not** run (it reaches the box); the
  diff-reviewer reads the step against R-tt6 and the shadow unit's rule.
- `ruff check` and `ruff format --check` run directly on `generate.py`
  pass; the report pastes both.
- `tools/venv/.venv/bin/python tools/unidecode-table/generate.py
  <scratch path>` prints a sha256 equal to the pin; the report pastes it.
  The scratch file is deleted, never staged.
- The gate is green.

**Named mutations this task reports:**

1. Reviewer-read: the step placed before the binary step would run an old
   binary without the verb; the reviewer confirms the order.
2. Run `generate.py` against a scratch venv copy with the version check
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
  the `'#'` string entries; the Python prototype's byte-identity figures;
  why the digest in a Go build still proves Python's lookup (R-tt3); the
  pin-update procedure (R-tt7); the four test binaries and the overlay probe
  that found them; the `make unidecode-wheel` prerequisite; that `serve`
  loads the table when it first scores; and the git-history question routed
  to `give-back`.
- **The STATUS draft** adds a carried item for 5d's runbook (the install
  step runs before the first shadow observation after the next deploy) and
  the workstation prerequisite, and records nothing else unless the pass
  found something.

**Acceptance criteria:** every count in the drafts comes from a command the
report pastes; every commit id resolves; the STATUS draft is at most 60
lines. The gate is green as a formality.

**Deliverable count: 1** (two drafts, one close).

---

## The exit gate

`cairn-run-gate 'bash /var/home/glw907/Projects/dubplate/scripts/check.sh'`
green on master after the last commit, plus:

1. The Go build of the pinned wheel equals the pinned table sha256 (Task 1).
2. `git ls-files '*.tbl'` prints nothing, and `git grep -n 'go:embed' --
   internal/engine` prints nothing (Task 2).
3. `strings dubplate-server-linux-amd64 | grep -c 'Kayng '` prints 0 after
   `make build-linux` (a `x0f9.py` entry, so the table's text is absent from
   the binary), run by the conductor.
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
| `scripts/deploy.sh` | rung 4 and 5d | Task 4 | 5d owes its own runbook rows; this step joins them |
| `docs/` ledgers, `CLAUDE.md` | conductor | Task 5 drafts only | every lane |

---

## The conductor's actions

Not tasks; no implementer performs them.

1. Before Task 0: record the rulings' outcomes; confirm the lanes per
   Task 0 blocker 1; record the planning spend; write the lane registry row;
   arm the runaway guard; check AC power.
2. Before Task 2: run `make unidecode-wheel` on the workstation once
   (a network act outside the gate), and confirm `sha256sum` of the cached
   file.
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
   and the lawyer's confirmation to ROADMAP `give-back`.
7. **Do not deploy.** The next deploy (5d) carries Task 4's step.

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
deploy step runs the verb), then Task 5. No task is independent.

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

- The Go port itself. The byte-identity claim rests on a Python prototype
  using the same restricted lexer and `_get_repl_str` rules; Task 1 is the
  proof.
- The per-binary cost of building the table in each `TestMain` (R-tt5 sets
  a 0.5s report threshold).
- Whether the box reaches `files.pythonhosted.org`. `deploy.sh` already
  installs beets from PyPI with `uv` on the box (`deploy.sh:554-557`), which
  suggests it does; the first deploy after this pass is the proof, and a
  refusal aborts that deploy before the timers.
- Whether lane 2's 6b-ii-b edits `doctor.go` or `main.go` (Task 0).
- The legal reading itself: whether downloading and deriving at setup keeps
  an MIT binary clear of GPL-2.0-or-later obligations. The lawyer's
  confirmation is owed before the first stranger-installable release.
- Whether PyPI keeps serving the 1.4.0 wheel. The file URL is
  content-addressed and marked immutable, and the offline `--from` path
  covers an operator who kept a copy.
