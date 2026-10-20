# Translit-table plan: fold verification (revision 2)

Target: `docs/plans/2026-09-25-translit-table-at-setup.md` at
`78f109a` (revision 2, committed). Fold record:
`docs/research/2026-09-25-translit-table-plan-fold.md`. The
verifier took no part in the fold. It read the plan, the fold record, and
the tree at HEAD (`scripts/deploy.sh`, `scripts/check.sh`,
`systemd/dubplate-shadow.service`, `env/dubplate-server.env.template`,
`internal/config/config.go`, `cmd/dubplate-server/doctor_test.go`), and it
re-ran the quoted `go doc`, `go help test`, and `man systemd-run` reads.
`git diff 6777d43 HEAD` touches no code, script, or Makefile path, so the
fold's line numbers still hold.

**Counts: 0 blocker, 1 major, 9 minor.** Findings are ranked by
consequence. Minors are cheap to fold or to refuse.

## Q1: did each major close at the cited location?

Yes for the contract and mechanics majors (C-M1 to C-M3, K-M1, K-M2). The
deploy design (K-M1, K-M2, R-M1, R-M2, C-M4) closes the window it names,
with one gap (V-M1 below).

Spot-check of `scripts/deploy.sh` at HEAD:

- The line numbers are correct. The binary step heading is `:304`, the
  staged rsync and `chmod` are `:341-344`, and the smoke test is `:364-366`.
  The `server.env` install is `:414`, the pre-create block is `:489-505`
  (`DUBPLATE_LIB_DIR` at `:502`, `mkdir`/`chown` at `:505`), `TIMER_UNITS`
  is `:526`, and `run_remote` is `:133-141`.
- The variable names are correct: `DUBPLATE_LIB_DIR`,
  `DUBPLATE_SHADOW_DIR`, `REMOTE_SERVER_ENV`, and `DUBPLATE_SERVER_BINARY`.
  The agreement leg (`check.sh:655`) reads `DUBPLATE_LIB_DIR` by a
  line-anchored `sed` pattern, so moving the block is safe.
- `server.env` exists before the install on a fresh box and on a
  re-deploy, because the install follows `:414`.
- `set -euo pipefail` (`:2`) aborts the deploy. `run_remote`'s last command
  is `ssh`, so a nonzero remote exit fails the function, and a plain call at
  top level aborts the script. `systemd-run --wait` propagates the exit. The
  man page's EXIT STATUS text is confirmed on the workstation's systemd 259.
  A `--user` run of the exact property list parsed every `-p` and returned
  the unit's exit (226, a user-mode namespace refusal) as its own.
- The staged `mv` keeps the old binary live on failure. The rsync writes
  only `.new`, and the rename is on one filesystem.

## Q2: contradictions and task order

The order builds: Task 0, then 1, 2, 3, 4, and 5, with the cache filled in
Task 1 before Task 2 reads it. There are two small contradictions (V-m1 and
V-m2).

## Q3: new mechanisms stated from memory

- `atomic.Pointer`, `os.MkdirAll`, `os.UserCacheDir`, `io.LimitReader`,
  `http.Client.Do`, and the `go help test` rule with
  `test.go:2051-2063`: each quote was re-run and matches.
- `CreateRaw`: the doc quote is accurate. "Header values are the caller's"
  comes from the source and not the doc (V-m4).
- The systemd-run properties: the man quotes are accurate. The plan does
  not quote that `-p` accepts `EnvironmentFile`, `ReadWritePaths`, and
  `RestrictAddressFamilies`, and it does not name the box's systemd version
  (V-m3).
- The rewriting transport is a test design, so it has no behavior to
  quote. It carries one unquoted contract (V-m5).
- `MaxWheelSize` is the plan's own constant. Its cap-plus-one read rests on
  the quoted `io.LimitReader`.

## Q4: can Geoff rule the two forks?

Yes. Ruling 1's arithmetic checks: the ten lines sum to 2.63M, and
2.63 / 0.8 = 3.29, which rounds up to 3.3 (flag 2.64M). Each answer's
effect is stated. One fact Geoff may want: the flag sits only 0.01M above
the projection, so almost any overrun raises it (V-m9). Ruling 2 states
the reason and both answers' effects, and a later rename costs little.

## Q5: planning-miss items on the changed tasks

- **Quoted behaviors:** met for Tasks 1 to 3. Task 4 misses the
  systemd-property acceptance (V-m3).
- **Named states:** met, except for the `cmd/dubplate-server` test binary
  with the wheel absent (V-m8).
- **Privileged-field channels:** met. The setter, the client, the table
  file, the pins, and the cache each name their channels.
- **Fixture states:** met. The doctor exit rows run on `baseEnviron()`'s
  green wall, the byte-identity test counts its blocks and entries, and the
  request-count rows fix the transport count.

## Findings

### V-M1 (major): a failed install can leave the old binary on the new env

`plan:1162-1165`, `plan:1207-1209`, and `fold:53-58`. The install runs
after the `server.env` install (`deploy.sh:414`). When it fails, the deploy
aborts with the new `server.env` live and the old binary scoring. The wall
refuses any `DUBPLATE_*` name it does not know
(`internal/config/config.go:287`, "unknown configuration key"), and 5d is
the deploy that adds the magic-link keys (`STATUS.md:45`). A failed install
at 5d would then leave `shadow observe` refusing every five minutes,
unmonitored. That is the R-M1 failure the staging exists to close, so the
claim "the old binary stays live and the shadow keeps scoring" does not
hold at the next deploy.

The fold's reason for the placement is moot. It said a deploy can change
`DUBPLATE_STATE_DIR`, but the agreement leg (`check.sh:645-658`) pins
`DUBPLATE_STATE_DIR` to equal `DUBPLATE_LIB_DIR`, and R-tt6 says the verb
reads that one variable alone.

**Fix:**

1. Replace `-p EnvironmentFile=/etc/dubplate/server.env` with
   `-p Environment=DUBPLATE_STATE_DIR=$DUBPLATE_LIB_DIR`. This carries no
   secret and reuses the existing variable.
2. Place the install and the `mv` directly after the smoke test (`:366`),
   before the env step. An abort then leaves the old binary on the old env,
   and the swap keeps today's binary-then-env order.
3. Update step 3's rationale, the failure sentence, named mutation 1, and
   the reviewer-read order to match.

### Minor findings

- **V-m1.** `plan:935-936` says Task 2 depends on "the conductor having
  run `make unidecode-wheel`". `plan:495-497` and `plan:1339-1341` say Task
  1 fills the cache and the conductor confirms it. Fix: change the text to
  "on Task 1 having filled the cache, confirmed by the conductor".
- **V-m2.** `plan:871` has the library's `Fetch` refusal naming "the
  offline guidance". R-tt6 (`plan:712-713`) makes that guidance the verb's,
  and it names the cmd flag `--from`, which a published library cannot
  know. Fix: the Task 1 row names the status only, and Task 3's
  "download refused" row carries the guidance.
- **V-m3.** `plan:333-339` and `plan:1171-1197`. The plan does not quote
  that `systemd-run -p` accepts `EnvironmentFile`, `ReadWritePaths`, and the
  space-bearing `RestrictAddressFamilies`. The man quotes come from the
  workstation's systemd 259, and the box (AlmaLinux) version is not named.
  The verifier's `--user` run on 259 parsed every property. Fix: add
  "property acceptance on the box's systemd is proved by the 5d `--check`
  and first deploy" to "Not verified", or quote a read-only
  `systemd-run --version` from the box.
- **V-m4.** `plan:283-289`. The doc does not say that `CreateRaw` keeps
  the caller's CRC and sizes. The source does (`writer.go:647-648`, where
  the raw close returns before the header is overwritten). Two more facts
  from the source: the reader checks the CRC only when the CRC is nonzero
  (`reader.go:322`), and a short entry yields `io.ErrUnexpectedEOF`, not
  `ErrFormat` (`:308-309`). Fix: cite the source lines and require a
  nonzero wrong CRC in the CRC row.
- **V-m5.** `plan:498-503` and `plan:1038-1039`. `go doc
  net/http.RoundTripper` says: "RoundTrip should not modify the request".
  A transport that rewrites the URL in place breaks that contract. Fix:
  `req.Clone(req.Context())`, then set the URL on the clone.
- **V-m6.** `plan:1203-1204`. Under `--check`, `rsync -n` to the `.new`
  path always reports a full transfer, so `--check` no longer shows whether
  the binary changed. Fix: run the `--check` rsync against the live path,
  or state the change in the task.
- **V-m7.** `plan:1193-1197`. The abort rests on the new `run_remote`
  calls being plain commands. Inside an `if` or an `||`, `set -e` is
  ignored. Fix: one clause in Task 4's reviewer-read.
- **V-m8.** Task 3's state table (`plan:1070-1086`) has no row for the
  `cmd/dubplate-server` test binary when the cached wheel is absent. Only
  Primary inputs (`plan:424-427`) states the rule. Fix: add the row
  ("fails naming `make unidecode-wheel`, `-count=1`").
- **V-m9.** `plan:77-80`. Cumulative spend with the fix round is 2.01M
  after Task 4 and 2.28M after Task 5, so a 2.9M ceiling (flag 2.32M) is
  crossed in the merge-side reads, not "during Task 5". Asking after Task 4
  is still sound. Fix: say "during the merge-side reads".
- **Trivial.** `fold:4` says revision 2 is uncommitted, but it is committed
  at `78f109a`.
