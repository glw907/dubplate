# Translit-table plan: review fold record (revision 1 to 2)

Target: `docs/superpowers/plans/2026-09-25-translit-table-at-setup.md`,
revision 1 at `b841954`, folded to revision 2 (committed at `78f109a`), then
to revision 2.1 by the second fold at the end of this record. Reviews folded, all in `docs/superpowers/research/`:

- `2026-09-25-translit-table-plan-review-contract.md` (C): 4 major, 9 minor.
- `2026-09-25-translit-table-plan-review-mechanics.md` (K): 2 major, 5 minor.
- `2026-09-25-translit-table-plan-review-risk.md` (R): 2 major, 4 minor.

Every finding was checked against the tree at `6777d43` before it was
acted on. The check read `scripts/deploy.sh`,
`systemd/dubplate-shadow.service`, `cmd/dubplate-server/doctor_test.go`,
`Makefile`, and the `case-mapping-sweep.tsv` header. It also quoted
`go doc`, `go help test`, and `man systemd-run` where the plan now leans on
them.

**Counts: 26 findings. 26 folded (the five deploy findings as one
converged design, plus three duplicate pairs each folded once: C-M2 with
K-m4, C-M3 with K-m3, and C-m6 with K-m1), 0 refused whole, 0 owner
forks.** Six folds
decline a sub-part of a finding, each with its reason inline: C-M4 (two
parts), K-M2, C-m8, K-m1, and K-m4. No major is
refused.

## The converged design: the deploy transition

**IDs: K-M1, K-M2, R-M1, R-M2, C-M4. Folded once, into Task 4.**

Verified in the tree. The binary rsync at `deploy.sh:343` replaces the live
binary before any state-directory step, and the pre-create sits at
`:489-505`. `TIMER_UNITS` at `:526` omits `dubplate-shadow.timer` on purpose,
and the reviewers' read-only box check found that timer enabled and active.
The shadow unit grants only `ReadWritePaths=/var/lib/dubplate/shadow` under
`ProtectSystem=strict`. `server.env` is `0600 root:root` (`:409-414`).

The design, as Task 4 now states it:

1. The `/var/lib/dubplate` pre-create block moves unchanged ahead of the
   binary step. The agreement leg reads `DUBPLATE_LIB_DIR` by pattern, so
   it needs no new row.
2. The binary is rsynced to `/usr/local/bin/dubplate-server.new` and
   smoke-tested there. The live binary is untouched.
3. The staged binary's install verb runs as `music` through `systemd-run
   --wait --pipe --collect -p User=music -p
   EnvironmentFile=/etc/dubplate/server.env`. It carries the shadow unit's
   hardening, with `-p ReadWritePaths=$DUBPLATE_LIB_DIR`, the network
   address families kept, and no `UMask=0027`.
4. Only then does `mv -f` swap the binary into place.
5. Rollback is redeploying the previous commit, whose binary scores from
   its own embedded table. The plan names the hand-enabled shadow timer.

One refinement came from the tree, not the reviews. `server.env` installs
at `:409-414`, after the binary step, so the staged install runs in a new
block directly after that install and not directly after the smoke test. A
fresh box has no env file before `:414`. The live binary stays untouched
throughout, so the window stays closed.

**Corrected in the second fold.** The first fold also gave "a deploy can
change `DUBPLATE_STATE_DIR`" as a reason. That reason was wrong: the
agreement leg (`check.sh:645-658`) pins `DUBPLATE_STATE_DIR` to equal
`DUBPLATE_LIB_DIR`, and R-tt6 has the verb read that one variable alone.
The placement itself was also reversed (V-M1 below): the install now takes
`Environment=DUBPLATE_STATE_DIR=$DUBPLATE_LIB_DIR` and runs directly after
the smoke test, before the env step.

Sub-parts declined:

- **C-M4's pre-create of `/var/lib/dubplate/unidecode`.** Declined. The
  grant is the state directory, and the verb creates its own leaf behind
  its ownership and shape checks. A second pre-created literal would add
  an agreement-leg row for nothing.
- **K-M2's form (b), `runuser` with a grepped env line.** Declined in
  favor of form (a), which all three reviews recommended. Form (b) drops
  the hardening, and R-M2 notes that its easy variant leaks `server.env`.
- **C-M4's "5d runs `--check` by hand before the timers".** Reshaped. The
  swap is safe now, so the 5d runbook row reads the staged install's
  would-run line under `--check`, and the timer ordering no longer
  matters.

The plan's `deploy.sh` pointers were re-quoted at `6777d43` (Primary inputs
item 10). The "aborts before the timers" sentence in "Not verified" is
corrected.

## Contract lens

| ID | Disposition | Where |
| --- | --- | --- |
| C-M1 | **Folded.** Verified: `TestDoctorCmd`'s "full valid wall" case sets `DUBPLATE_STATE_DIR` to an empty temp dir, and the exit rule at `doctor.go:129` is an OR. Task 0 note 8 now lists every exit-0 test with the variable set. Task 3 installs a table in each of them through the verb with `--from`. The `missing` and `invalid` exit rows run on `baseEnviron()`'s all-green wall. | Task 0 note 8; Task 3 outcomes, state table, criteria |
| C-M2 | **Folded, with K-m4.** The install verb and `harness fetch-unidecode` take the `*http.Client` from `main` as a constructor argument, the way `newDoctorCmd` takes `environ`. K-m4's package-variable seam was not taken, since the constructor is the stated design and adds no global. Tests use a rewriting `RoundTripper` that asserts the URL equals `WheelURL`, counts the request, forwards it to `httptest`, and refuses any other host. A new channels row names the download client. | Global constraints; channels table; R-tt6; Tasks 1, 3 |
| C-M3 | **Folded, with K-m3.** An unexported `extract` stage (zip, count, lexer) and an unexported `decode` stage (truncation, trailing bytes) are tested directly over synthetic zips. The plan quotes `go doc archive/zip.Writer.CreateRaw` for the CRC and size rows. Each row asserts the stage's own error and `!errors.Is(err, ErrWheelChecksum)`. The data-module count (190) becomes a pin, and the R-tt7 bump procedure and `generate.py`'s docstring now name it. | Quoted behaviors; R-tt2; R-tt3; R-tt7; Task 1 |
| C-M4 | **Folded** in the converged design above. | Task 4 |
| C-m1 | **Folded.** Verified: `internal/engine/store/migrations.go` carries a `go:embed`. Exit-gate check 2 is now scoped to `internal/engine/matcher`. | Exit gate 2 |
| C-m2 | **Folded.** Exit-gate check 3 runs on a base-commit binary as the positive control, and the plan notes that `grep -c` exits 1 on a pass (from K's notes). | Exit gate 3 |
| C-m3 | **Folded.** The moved sweep test is dropped, and byte identity is named as the proof that replaces `TestUnidecodeTableFullSweep`. | Task 1 criteria; Task 2 tests |
| C-m4 | **Folded.** The byte-identity test hashes its own output with `sha256.Sum256` and compares with the literal `c954605b...ce85` in the test file, never the package's pin constant. The diff-reviewer confirms the pin equals the literal. | Channels table; Task 1 criteria |
| C-m5 | **Folded.** The over-cap row asserts `ErrWheelTooLarge` (a new sentinel, used by the install verb for both the download and `--from`) and `!errors.Is(err, ErrWheelChecksum)`. A body of exactly the cap is not refused for size. | R-tt2 table; Task 1 |
| C-m6 | **Folded, with K-m1.** See K-m1. | R-tt5 |
| C-m7 | **Folded.** Leaf rows (a file, the wrong mode, another owner) are added, and the ownership refusals are ordered before the "already installed" check. The `--from` row is split: absent or a directory names the path, and the wrong file names both sums. | R-tt6; Task 3 state table |
| C-m8 | **Folded.** The row state now follows the error: `fs.ErrNotExist` reads as `missing` and anything else as `invalid`. The helper splits into a load (doctor and the scoring verbs) and an install (the scoring verbs only), and the channels row says doctor never installs. The test for doctor calling only the load was not taken, since the scoring-verb tests install process-wide in the same binary, so the diff-reviewer checks it by reading the diff. | Channels table; Task 3 |
| C-m9 | **Folded, all five points.** The `*url.Error` quote and `errors.Is(err, context.Canceled)`. The quoted `sync/atomic.Pointer`, with the panic test clearing the holder internally and the setter panicking on nil (no uninstall channel). The first `make unidecode-wheel` runs on a fresh `XDG_CACHE_HOME`, and Task 1 is now correctly named as the task that fills the cache. New rows for a cache failing its checksum and for `os.UserCacheDir` erroring. `cmd/dubplate-server`'s test binary is named as a wheel reader after Task 3. | Quoted behaviors; channels table; Global constraints; Tasks 1, 2 |

## Mechanics lens

| ID | Disposition | Where |
| --- | --- | --- |
| K-M1 | **Folded** in the converged design. | Task 4 |
| K-M2 | **Folded** in the converged design, form (a). | Task 4 |
| K-m1 | **Folded, with C-m6. Cheapest honest option: reword, no `-count=1` in `make test`.** Verified in `go help test` and `cmd/go/internal/test/test.go:2051-2063` ("Do not recheck files outside the module"). R-tt5 and Review focus 1 now say "every uncached test run" and state what the cache does. A cached `ok` is sound because the wheel is pinned by sha256. The one state the cache hides, "wheel absent", is proved with `-count=1`. Forcing `-count=1` onto the four packages was declined: it would re-run the conform corpus sweeps on every gate to re-prove a pinned input. C-m6's optional re-stat inside a test was declined too, since the cache does not track a file outside the module whether or not a test stats it. | Quoted behaviors; R-tt5; Review focus; Tasks 1, 2; exit gate 1 |
| K-m2 | **Folded.** A global constraint now says no `time.Now` under `internal/engine`, tests included, and cites `checkNoTimeNow`. | Primary inputs 5; Global constraints |
| K-m3 | **Folded, with C-M3.** | Task 1 |
| K-m4 | **Folded, with C-M2** (constructor injection, not a package variable). | Task 3 |
| K-m5 | **Folded.** The verb sets the leaf's mode with `os.Chmod` after creating it, and the plan quotes `go doc os.MkdirAll` ("before umask"). Task 3 mutation 6 proves it under `Umask(0o027)`. The box install also omits `UMask=0027`. | Quoted behaviors; R-tt6; Tasks 3, 4 |

K's "checked and not raised" notes were folded where they cost only
wording. `Build`'s sweep runs over the in-memory blocks and not over
`Parse`. `TestTransliterateRune` joins Task 2's removal list. The holder is
loaded once per call. The Go prototype's figures (sha `c954605b...ce85`,
0.14s per build) now sit beside the Python ones.

## Risk lens

| ID | Disposition | Where |
| --- | --- | --- |
| R-M1 | **Folded** in the converged design, with the rollback and the timer named. | Task 4; Primary inputs 10 |
| R-M2 | **Folded** in the converged design (the exact `systemd-run` shape). | Task 4 |
| R-m1 | **Folded.** "The repo carries no Unidecode data" becomes an exact list: `case-mapping-sweep.tsv`'s 56 data rows (verified: 56 non-comment lines, header naming Unidecode 1.4.0), the pinned lookup rows, `generate.py`, and `Lookup`'s port of `_get_repl_str`. Task 5 and the conductor's `give-back` routing carry the list. | Owner touchpoints; Task 5; conductor action 6 |
| R-m2 | **Folded.** `--from` reads under the same `MaxWheelSize` (1 MiB) cap-plus-one refusal, with a new Task 3 row. | R-tt2; R-tt6; Task 3 |
| R-m3 | **Folded.** Task 0 check 4 now reads "present and unyanked". The 5d runbook row names the workstation's verified cache as the estate's `--from` copy. The stranger's case stays with `packaging`. | Task 0; Task 5; Not verified |
| R-m4 | **Folded as routing only.** The process-wide holder's cost to a published library is stated in R-tt1, carried by Task 5's HISTORY draft, and routed to `give-back`. The holder does not change in this pass. | R-tt1; Task 5; conductor action 6 |

## Rulings (the only owner forks)

1. **Ceiling: 3.3M** (was 3.1M). The folds add 0.08M to Task 1, 0.10M to
   Task 3, and 0.10M to Task 4, for a projection of 2.63M. The ceiling is
   that projection divided by 0.8 and rounded up to the next 0.1M, so the
   flag (2.64M) sits just above the projection. The "no" answer holds
   2.9M, and the conductor asks after Task 4.
2. **Verb name: `unidecode install`** (recommended) or `translit install`.
   The fold did not change it.

No review raised a new owner fork.

## Second fold (revision 2 to 2.1)

Input: `docs/superpowers/research/2026-09-25-translit-table-plan-fold-verification.md`
(1 major, 9 minor, 1 trivial), against revision 2 at `78f109a`. Every
finding was checked against the tree before it was acted on. The check read
`scripts/deploy.sh` (`:133-141`, `:304-366`, `:392-414`, `:489-505`),
`internal/config/config.go:287`, `scripts/check.sh:645-658`,
`cmd/dubplate-server/store.go:90` and `main.go:138-143` (no verb reads the
environment outside its own constructor), the Go 1.27.1 `archive/zip`
source, `go doc net/http.RoundTripper` and `net/http.Request.Clone`, and
`man systemd-run` and `man systemd.exec` on systemd 259. One read-only ssh
read the box's `systemctl --version`: systemd 257.

**Counts: 11 findings. 11 folded, 0 refused, 0 owner forks.** No ruling
changed in substance; Ruling 1 gains the flag-margin note and the
corrected crossing point.

| ID | Disposition | Where |
| --- | --- | --- |
| V-M1 | **Folded as proposed; the tree agrees.** `config.go:287` refuses an unknown `DUBPLATE_*` name, so a failed install after the env step would leave the old binary refusing a key 5d adds. R-tt6 reads `DUBPLATE_STATE_DIR` alone through `lookupEnviron`, and `main.go` hands `os.Environ()` only to constructors, so no other `DUBPLATE_*` key reaches the verb; the property is `-p Environment=DUBPLATE_STATE_DIR=$DUBPLATE_LIB_DIR`, and the agreement leg pins the two equal. The install and the `mv` move to directly after the smoke test (`:366`), before the env step, which now follows the swap (today's binary-then-env order). Step rationale, the four named hardening differences, the failure sentence, named mutation 1, and the reviewer-read order are updated. The first fold's wrong reason is corrected in place above. | Status line; Primary inputs 10; Task 4 steps 2 to 5, failure, criteria, mutation 1 |
| V-m1 | **Folded.** Task 2 depends on Task 1 having filled the cache, confirmed by the conductor (action 2). | Task 2 header |
| V-m2 | **Folded.** Task 1's `Fetch` status row names the status only and points to Task 3; Task 3's "download refused" row names the guidance's three parts; Review focus 4 splits the two. | Task 1 table; Task 3 table; Review focus 4 |
| V-m3 | **Folded, both halves.** The `-p` text is quoted from `man systemd-run`. The box's `systemd 257 (257-23.el10_2.2.alma.1-gdb72f78)` is named from a read-only ssh, with each option's and property's "Added in version" (all at or below 236). "Not verified" names the 5d `--check` and first deploy as the proof of acceptance on 257. | Quoted behaviors; Not verified |
| V-m4 | **Folded.** The source is cited at Go 1.27.1: `writer.go:647-648` (the raw close returns before the header sums are overwritten), `reader.go:325` (CRC checked only when nonzero; the verifier's `:322` was a line off on this toolchain), and `reader.go:307-309` (a short entry yields `io.ErrUnexpectedEOF`). The CRC row's wrong CRC is nonzero. | Quoted behaviors |
| V-m5 | **Folded.** The transport forwards `req.Clone(req.Context())` with the URL set on the clone; `go doc net/http.RoundTripper` and `Request.Clone` are quoted. Task 3's transport names the clone too. | Global constraints; Task 3 interfaces |
| V-m6 | **Folded, first option.** The `--check` `rsync -n` keeps the live path as its target, so `--check` still shows whether the binary changed; the apply rsync alone names `.new`. The reviewer-read checks both targets. | Task 4 step 2, criteria |
| V-m7 | **Folded.** Step 3 states that neither new `run_remote` call may sit in an `if` or beside `||` or `&&`, and the reviewer-read checks it. | Task 4 step 3, criteria |
| V-m8 | **Folded.** Task 3's state table gains the `cmd/dubplate-server` row: cached wheel absent under `-count=1` fails naming `make unidecode-wheel`, never skips. | Task 3 table |
| V-m9 | **Folded.** Ruling 1's "no" answer now says the 2.32M flag is crossed during the merge-side reads (about 2.01M after Task 4, 2.28M after Task 5). The "yes" answer notes the 2.64M flag sits 0.01M above the 2.63M projection, so almost any overrun raises it. | Ruling 1 |
| Trivial | **Folded.** This record's header now says revision 2 is committed at `78f109a`. | Header |
