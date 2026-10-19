# Transliteration table plan review: mechanics and feasibility

Target: `docs/superpowers/plans/2026-09-25-translit-table-at-setup.md` at
`b841954` (revision 1). Lens: does every mechanism behave as the plan says?
Each claim below was checked by running it, by quoting `go doc`, or by
reading the source.

**Counts: 0 blockers, 2 majors, 5 minors.** The riskiest assumption holds.
A Go port reads the wheel with `archive/zip`, lexes all 190 modules under the
plan's grammar, and builds a table **byte-identical** to the committed one.
Both majors are in Task 4's deploy step, which the plan never runs before 5d.

## Byte-identity result (the wheel)

I wrote a stdlib-only prototype at
`<scratchpad>/proto/main.go`, outside the tree. It verifies the wheel's
sha256, opens it with `zip.NewReader(bytes.NewReader(wheel), size)`, and
reads every entry to EOF. It selects `^unidecode/x([0-9a-f]{3})\.py$`, lexes
each module with a strict hand-written lexer, and sorts the blocks. It then
sweeps `[0, 0x10FFFF]` through a `_get_repl_str`-shaped lookup and writes
`generate.py`'s format. Its output:

```
wheel sha256 c3c7606c27503ad8d501270406e345ddb480a7b5f38827eafe4fa82a137f0021
blocks 190 entries 48597 backslashes 167
table bytes 255458 sha256 c954605b805e4a21d1d9e79b4d93bc6b89684de9996dd21720b256fc58a6ce85
elapsed 140.98304ms
```

`cmp` against `internal/engine/matcher/unidecode-1.4.0.tbl` reports no
difference. The lexer accepts exactly these inputs:

- the `data = (` head
- whitespace
- `#` comments, but only outside a string
- `None`
- single-quoted and double-quoted strings, using only the escapes `\'`, `\\`,
  and `\xHH`
- commas, one required after every entry
- one `)`, with nothing after it except whitespace and comments

It refuses any non-ASCII byte and any CR. Every module passed. That confirms
the plan's grammar is sufficient, including the `'#'` entries in `x006.py`,
`x025.py` (many), and `x028.py`. Escape counts from the wheel are `\'` 128,
`\\` 18, and `\x` 3. One double-quoted string sits at `x020.py`. These match
the plan's figures (`plan:303-313`). Two shapes that a line-per-entry lexer
would miss also exist: `x000.py` packs sixteen entries per line, and three
modules carry standalone comment lines holding `"` and `'` characters. A
token lexer handles both. The plan's lexer is a token lexer, so these are
not findings.

The build takes about 0.14s per binary. The sweep is included and the race
detector is off. That is well under R-tt5's 0.5s report threshold
(`plan:534-537`).

Other claims I confirmed:

- **PyPI pin.** I fetched the JSON on 2026-09-25. It gives the wheel's size
  as 235837 and gives the sha256 and blake2b_256 the plan quotes. `yanked` is
  false. The downloaded file's sha256 and its blake2b (digest size 32) both
  match. `HEAD` on the URL returns 200 with
  `cache-control: max-age=365000000, immutable, public`. `METADATA` reads
  `Version: 1.4.0`. `diff -rq` shows the wheel's `unidecode/` is identical to
  the venv copy (`__pycache__` aside).
- **The `archive/zip` quotes.** The `reader.go:295-333` behavior the plan
  cites is accurate at go1.27.1. `ErrFormat` fires past
  `UncompressedSize64`, and `ErrChecksum` fires at EOF when the CRC is set and
  disagrees. All 200 entries use `Defl:N`.
- **Named mutation 3 is live.** `x000.py` carries `''` for positions
  0x00-0x7f. The file's own comment says "Hence they are not present", but
  the tuple does carry them as empty strings. Dropping the `< 0x80`
  passthrough therefore maps ASCII to `""` and moves the digest.
- **The test-binary set is the same under every tag.** `go list -test` under
  no tag, under `oracle`, and under `rendered` names the same six test
  binaries that link `matcher`. `cmd/dubplate-server` is the only `main`
  package that links `matcher`, so no other program needs an installer.
- **TestMain.** None of the four packages uses `t.Parallel()` or a helper
  process (`os.Args[0]`). Installing before `m.Run()` happens-before every
  test, so there is no ordering hazard. The panic test swaps the holder to
  nil and restores it. That test is safe because it is serial.
  `cmd/dubplate-server` already has a `TestMain` (`backup_test.go:68`), but
  the plan gives it none, which is correct.
- **`os.UserCacheDir` in the gate.** `cairn-run-gate` wraps the gate in
  `systemd-run --user --scope`, which inherits the caller's environment.
  Running `systemd-run --user --scope -q env` prints
  `HOME=/var/home/glw907` and no `XDG_CACHE_HOME`, so the cache resolves to
  `~/.cache/<pkg>`. `.github/` does not exist, so there is no CI to break.
- **`shadow.configHash` and `matcher_config_hash`.** Neither changes.
  `configHash` (`internal/engine/shadow/match.go:63`) hashes only the JSON of
  `Tunables` and `Compat`. The seam is outside `Tunables` (R-tt1), and
  `MatcherConfigHash` is that same value. `matcher_commit` moves with any
  commit, as it does today.
- **The box and PyPI.** A read-only `ssh root@5.78.148.42 curl -sI <wheel
  URL>` returned `HTTP/2 200`. The binary is `CGO_ENABLED=0` and already
  reaches MusicBrainz over HTTPS under the shadow unit's
  `RestrictAddressFamilies=AF_INET AF_INET6 AF_UNIX`, so TLS and DNS work.
  `/var/lib/dubplate` is `music:music 0755` on the box, so the owner check
  passes as `music`.
- **Exit-gate check 3.** `Kayng ` occurs only in `x0f9.py`, and
  `strings <tbl> | grep -c 'Kayng '` prints 1 today. The probe is sound.
  Note that `grep -c` exits 1 when the count is 0, so the conductor should
  expect exit 1 on a pass.

## Findings

### M1 (major): the deploy step runs after the binary swap, while the shadow timer is live

- **Where:** `plan:934-945` (Task 4, "The deploy step"), `plan:1120-1123`,
  and `plan:995-997` (the STATUS carried item). On the deploy side:
  `scripts/deploy.sh:337-345` (the binary rsync),
  `scripts/deploy.sh:364-366` (the smoke test), `scripts/deploy.sh:505` (the
  `/var/lib/dubplate` pre-create), and `scripts/deploy.sh:525-527` (the
  timers).
- **Defect:** The plan places the install step after the pre-create at :505.
  Its safety argument is that "a refusal aborts the deploy ... before the
  timers are enabled." On the box, though, `dubplate-shadow.timer` is already
  `enabled` and `active` (read-only check, 2026-09-25). `TIMER_UNITS` never
  touches it. So once :340 replaces `/usr/local/bin/dubplate-server`, every
  `*:0/5` firing runs a binary with no table and refuses under R-tt4. That
  lasts until the install step runs, after the env files, the pre-creates,
  and `compose up`. If the install fails (a PyPI outage or a sandbox
  refusal), the deploy aborts with the new binary already in place. The
  shadow then fails every five minutes. `dubplate-shadow.service`'s own
  comment says nothing monitors it at this rung: "no monitoring at rung 4".
  Aborting "before the timers" protects nothing, because the only timer that
  scores is already on.
- **Evidence:** On the box, `systemctl is-enabled dubplate-shadow.timer`
  returned `enabled` and `is-active` returned `active`. `deploy.sh:526`
  leaves the shadow timer out of `TIMER_UNITS` on purpose (R10). The swap
  happens at :340, before any state-directory step.
- **Proposed fold:**
  1. Install the table with the new binary before the swap. Rsync the binary
     to a staging path beside the live one. Smoke-test the staged binary, run
     its install verb as `music`, and only then `mv` it into place.
  2. Move the `/var/lib/dubplate` pre-create (`:502-505`) ahead of the binary
     step so a fresh box has the owned state directory in time.

  With this order, a failed download leaves the old binary serving and the
  shadow still scoring. It also makes the 5d runbook row unnecessary. State
  in R-tt6 or Task 4 that the shadow timer may already be live.

### M2 (major): the install step's sandbox is left open, and the suggested form cannot write the leaf

- **Where:** `plan:936-941` ("The shadow unit's operator rule suggests a
  `systemd-run` carrying `User=music` and `EnvironmentFile=...`; the
  implementer chooses"), `plan:219-222`, and `plan:953-954` (`--check` is
  never run). On the unit side: `systemd/dubplate-shadow.service`
  (`ProtectSystem=strict`, `ReadWritePaths=/var/lib/dubplate/shadow`, and
  the operator rule "systemd-run carrying this unit's own properties").
- **Defect:** The step has to create `/var/lib/dubplate/unidecode/` and write
  inside it. Suppose the implementer follows the unit's operator rule
  literally and carries the unit's hardening. `ProtectSystem=strict` then
  makes `/var/lib/dubplate` read-only, and the only grant is `.../shadow`, so
  `MkdirAll` fails with EROFS. The plan does not pin a form, and it forbids
  every rehearsal (`--check` is not run, and no deploy happens). The first
  execution of this step is the production 5d deploy. Combined with M1, that
  means a broken shadow.
- **Evidence:** The unit file as read. `server.env` is `0600 root:root`
  (`deploy.sh:409-414`), so `music` cannot read it itself. Only a root-side
  reader works: systemd's `EnvironmentFile=` as PID 1, or root extracting
  the one line for `runuser`.
- **Proposed fold:** Pin one of two forms in Task 4.
  - **(a)** `systemd-run --wait --pipe --collect -p User=music -p
    EnvironmentFile=/etc/dubplate/server.env`, carrying the unit's hardening
    plus `-p ReadWritePaths=$DUBPLATE_LIB_DIR`. That reuses the existing
    variable, so there is no new literal and no agreement-leg row.
  - **(b)** `runuser -u music -- env "$(grep '^DUBPLATE_STATE_DIR='
    /etc/dubplate/server.env)" <binary> <verb>`, which follows the beets
    step's `runuser` precedent at :557.

  Add a row to Task 4's review focus: "the step's sandbox grants write on
  the state directory's `unidecode/` leaf and reaches the network." I
  recommend (a): it keeps the unit's hardening, and the operator rule asks
  for `systemd-run`.

### m1 (minor): Go's test cache does not track the wheel, so "every gate run re-proves" is false

- **Where:** `plan:534-535` ("every gate run re-proves the port
  byte-identical") and `plan:590-592` (Review focus 1).
- **Defect:** `make test` runs `go test ./...` without `-count=1`. Go
  rechecks opened files only when they sit inside the module, GOPATH, or
  GOROOT. The wheel lives under `~/.cache`, so a cached pass survives a
  deleted or corrupted wheel. The build is re-proved only when a test
  binary's inputs change. Correctness does not suffer, because the wheel is
  pinned by sha. The failure is a claim that is not true, plus a "fails when
  absent" state that only a cache miss exercises.
- **Evidence:** `$(go env GOROOT)/src/cmd/go/internal/test/test.go:2056-2063`
  reads: "Do not recheck files outside the module, GOPATH, or GOROOT root."
  `go help test` reads: "Tests that open files within the package's module
  ... only match future runs in which the files ... are unchanged."
  `Makefile:17-18` runs `go test ./...` with no `-count`. The env reads
  (`HOME`, `XDG_CACHE_HOME`) are tracked.
- **Proposed fold:** Reword to "every test run that builds a binary
  re-proves ...". Task 2's absent-wheel proof already uses `-count=1`, so it
  holds.

### m2 (minor): the no-`time.Now` invariant covers the new package and three of the TestMain files

- **Where:** Global constraints (`plan:350-403`), Task 1's "the build's wall
  time" (`plan:720`), and Task 1's `Fetch` timeout.
- **Defect:** `checkNoTimeNow` (`internal/harness/conform/arch_test.go:66-117`)
  walks every `.go` file under `internal/engine`, `_test.go` files included,
  and fails on any `time.Now` call. The new package and three of the four
  `TestMain` files (`matcher`, `tag`, `shadow`) sit under `internal/engine`.
  An implementer who times the build with `time.Now()`, or builds a deadline
  with `time.Now().Add`, reds the conform suite. The cost is a fix round, not
  a defect that ships.
- **Proposed fold:** Add one line to Global constraints: "no `time.Now` under
  `internal/engine`, tests included (`arch_test.go`); a timeout comes from
  the caller's client or context, and wall time comes from `go test -v`."

### m3 (minor): three of Task 1's state rows cannot be reached through `Build`

- **Where:** `plan:700-703` and `plan:484-492` (R-tt3's order).
- **Defect:** `Build` checks the wheel's sha256 first, so no input can reach
  these rows through the exported path:
  - "not a zip"
  - an entry that fails its CRC or size check
  - "fewer or more than 190" entries or a name outside the pattern

  Any altered wheel stops at `ErrWheelChecksum`. A test written through
  `Build` would either be impossible to write or would pass vacuously on the
  checksum refusal. The plan notes the same shape for `Parse` at `:706`, but
  not for these rows.
- **Proposed fold:** State that these rows test the unexported
  post-checksum stage (zip read and selection) directly, over synthetic zips
  built with `archive/zip.Writer`. Say that each row's assertion names the
  stage's own error, not `ErrWheelChecksum`.

### m4 (minor): the request-counting proofs need a transport seam the plan does not name

- **Where:** `plan:707-712` (the `Fetch` and `harness fetch-unidecode` rows)
  and `plan:880-883` (the install verb's "no request" and download-refusal
  rows).
- **Defect:** `WheelURL` is a constant `https://files.pythonhosted.org/...`
  URL. An `httptest` server cannot answer it unless the client's transport
  redirects it, through a `RoundTripper` that rewrites the host, or through
  `DialContext` plus TLS config trusting the httptest cert under a matching
  server name. `Fetch` takes the client, so Task 1 can do this. The install
  verb builds its own client, so Task 3's counting-server rows need an
  unexported seam in `cmd/dubplate-server`. The plan names none, which
  leaves it to the implementer mid-task. That is the planning-miss class
  CLAUDE.md names.
- **Proposed fold:** Name the seam in Task 3's Interfaces: an unexported
  package variable returning the verb's `*http.Client`, following the
  `geteuid` precedent. Name the mechanism in Task 1: a `RoundTripper` that
  rewrites the URL to the httptest server.

### m5 (minor): the leaf's `0755` mode depends on the umask

- **Where:** R-tt6 (`plan:546-548`) and Task 3's success row (`plan:884`).
- **Defect:** `os.MkdirAll(dir, 0o755)` is filtered by the process umask.
  Under a `UMask=0027` sandbox, the shadow unit's value and a likely choice
  for M2's form (a), the leaf becomes `0750`. That is harmless, because the
  owner and the reader are both `music`. But the test's `0755` assertion
  then holds only under the workstation's `022`. The file is fine:
  `atomicfile.WriteFile` calls `os.Chmod(tmpPath, perm)`
  (`internal/atomicfile/atomicfile.go:36`).
- **Proposed fold:** `os.Chmod` the leaf after creating it, or assert the
  mode the code sets explicitly. Otherwise, drop the `0755` claim for the
  directory.

## Checked and not raised

- **`Build`'s sweep and the pin.** `Build`'s sweep "through the same lookup
  `Parse` returns" cannot call the exported `Parse`, because the table's
  sha256 is not known until the digest is written. The implementer will use
  the shared `Lookup` over the in-memory blocks. That is wording, not a
  defect.
- **`TestTransliterateRune`.** It is missing from Task 2's removal list, but
  its callee is deleted, so the compiler forces the move. Task 1 carries its
  rows.
- **Atomic loads on the hot path.** An atomic load per rune in `Transliterate`
  costs little. Loading once per call would be the idiom, which is polish
  for the simplifier.
