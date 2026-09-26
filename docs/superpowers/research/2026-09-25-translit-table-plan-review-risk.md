# Translit-table plan review: domain risk

Target: `docs/superpowers/plans/2026-09-25-translit-table-at-setup.md` at
`b841954` (revision 1). Lens: what can go wrong in the domain. Findings are
correctness gaps ranked by consequence; polish is left out.

**Counts: 0 blockers, 2 major, 4 minor, no owner forks.**

The core safety property holds. No path in the plan lets a wrong or partial
table score an album in production. `Parse` hashes the whole file against the
pin before decoding. That single check refuses every state the brief names:
a truncated file, a partial write, a table from another Unidecode version (the
version field sits inside the hashed bytes), a hand edit, an empty file, a
directory, and an unreadable file. The scoring verbs load the table before
`shadow.OpenStore`, and the uninstalled seam panics on every input. Both
risks this review ranks as major sit in the live-box transition in Task 4.

## Major

### M1. The deploy swaps the binary before the table exists, and the live shadow timer is outside the deploy's control

Location: plan `:934-945` (Task 4, the deploy step) and `:1120-1123` ("a
refusal aborts that deploy before the timers"); `scripts/deploy.sh:343`
(binary rsync), `:503` (shadow pre-create), `:526` (`TIMER_UNITS`);
`systemd/dubplate-shadow.timer:23` (`OnCalendar=*:0/5`).

Defect. The plan puts the install step after the `/var/lib/dubplate`
pre-create, near `deploy.sh:503`. The new binary goes live at `:343`. From
that moment until the install step runs, `dubplate-shadow.timer` can fire
`shadow observe` against a binary that has no table. That run refuses. The
refusal is correct and records no verdict, but the unit fails.

The plan's safety argument is "a refusal aborts that deploy before the
timers are enabled". That argument does not hold for the shadow timer.
`TIMER_UNITS` (`:526`) does not include `dubplate-shadow.timer`. Geoff
enabled it by hand at the launch sitting, it is live now, and `deploy.sh`
never stops or starts it. So an install failure has a real cost. The failure
could be PyPI unreachable, a checksum refusal, or M2's EROFS. In each case
the deploy aborts with the new, tableless binary already in place, and the
shadow fails every five minutes until someone intervenes. The observe
directory is walked again on the next good run, so no observation is lost,
but no pass is recorded during the outage while the evidence clock keeps
running. The plan also names no rollback path.

Fold. Stage the binary. Rsync it to `/usr/local/bin/dubplate-server.new`.
Smoke-test the staged binary and run the staged binary's install verb (see
M2 for how). Only then `mv` the staged binary over the live one, which is
atomic on one filesystem. For this to work, the shadow pre-create
(`:490-504`) moves ahead of the binary step. It is idempotent, so moving it
is safe. With this order, an install failure leaves the old binary running,
and the old binary scores from its own embed. State the rollback in the plan
and in the 5d runbook row: redeploy the previous commit. A pre-pass binary
ignores `unidecode/` entirely, so rollback needs no state change, and a
later forward deploy finds the table already in place and makes no network
request. Correct the sentence at `:1122-1123`.

### M2. The run-as mechanics the plan suggests fail on this box, and the tempting fix leaks secrets

Location: plan `:934-942`; `systemd/dubplate-shadow.service`
(`ReadWritePaths=/var/lib/dubplate/shadow`, `ProtectSystem=strict`,
`UMask=0027`); `deploy.sh:409-414` (`server.env` is `0600 root:root`).

Defect. The plan suggests "a `systemd-run` carrying `User=music` and
`EnvironmentFile=`" and leaves the choice to the implementer. The
diff-reviewer only reads this step, and nothing runs it before the 5d
sitting, so either reading below fails in front of Geoff.

- **systemd-run with the shadow unit's properties.** The shadow unit's
  operator rule tells the implementer to carry them. With `ProtectSystem=strict`
  and `ReadWritePaths=/var/lib/dubplate/shadow`, the install verb cannot
  create `/var/lib/dubplate/unidecode/`, so it fails with EROFS.
- **`runuser -u music` (the beets step's shape).** `music` cannot read
  `/etc/dubplate/server.env`, so `DUBPLATE_STATE_DIR` is unavailable. The
  easy workaround, exporting the whole file into the verb's environment as
  root, hands every secret in `server.env` to a process that needs one path.

Fold. Specify the invocation in the plan:
`systemd-run --wait --pipe --collect -p User=music -p EnvironmentFile=/etc/dubplate/server.env -p ProtectSystem=strict -p ReadWritePaths="$DUBPLATE_LIB_DIR" -p NoNewPrivileges=yes /usr/local/bin/dubplate-server.new <verb>`.
Under M1's staging, that path is the staged binary. `$DUBPLATE_LIB_DIR`
already exists at `deploy.sh:502`, so the step adds no new state-directory
literal and the agreement leg is untouched. PID 1 reads the env file, so no
secret passes through the shell. `--wait` propagates the verb's exit code,
so a refusal aborts the deploy. Two properties are deliberately absent.
`RestrictAddressFamilies` is omitted, or it must include `AF_INET` and
`AF_INET6`, because the verb downloads. `UMask=0027` is omitted: under it,
`MkdirAll(0o755)` yields `0750`, and Task 3's pinned leaf mode would not
hold on the box. Nothing breaks either way, since the reader is also
`music`, but the box then diverges from the tested mode. Add the `--check`
would-run line in the same shape.

## Minor

### m1. The distribution claim routed to the lawyer overstates what HEAD carries

Location: plan `:109-112` ("the repo and binary carry no Unidecode data")
and `:985-994` (Task 5's routing).

At HEAD after this pass, the binary carries no table, which is what matters
most. The repo still carries some Unidecode-derived material:

- `internal/engine/matcher/testdata/case-mapping-sweep.tsv`: 56 post-strip
  outputs computed through Unidecode 1.4.0 (its header says so).
- The ten or so lookup rows the tests pin.
- `generate.py`, which imports the GPL library.
- `Lookup`'s port of `_get_repl_str`'s rules, read from the GPL
  `__init__.py`. `translit.go` already ported these rules before this pass.

None of this is the table, and each item is plausibly de minimis. A lawyer
should still read the accurate list, not the absolute claim.

Fold. Replace the claim with "the binary carries no Unidecode data. The repo
carries the listed items and the pre-pass table in history". Have Task 5 and
the conductor's `give-back` routing carry that list.

### m2. `--from` reads an unbounded file before the checksum

Location: plan `:561`, `:882`.

`Fetch` caps its read at 1 MiB plus one byte, but `--from` has no cap. A
mistaken path, such as a large download or `/dev/zero`, is read whole into
memory before the sha256 check can refuse it.

Fold. Apply the same cap-plus-one read to `--from`, and add one row to
Task 3's state table.

### m3. PyPI can delete the release, and the estate names no offline copy

Location: plan `:1128-1130`; Task 0 check 4 (`:631-634`).

`immutable` in the cache header does not bind the project owner, who can
delete a release outright. Deletion is a stronger action than yanking, and
after it the content-addressed URL returns 404. The rung 6 backup set covers
the store's snapshots, not `unidecode/`. So once the release is gone, a
rebuilt or restored box cannot score until someone finds a copy of the
wheel.

Fold. The 5d runbook row (Task 5's STATUS carry) names the workstation's
verified cache (`make unidecode-wheel`'s output) as the estate's `--from`
source. Task 0 check 4 reads "present and unyanked", not just "unyanked".
The product-level question, where a stranger gets the wheel after a
deletion, is `packaging`'s.

### m4. The seam's type suits the published library; its process-wide panic default may not

Location: plan `:424-449` (R-tt1).

`func(rune) (string, bool)` is the right shape. The `bool` carries
Unidecode's `None`, so the matcher applies `errors='ignore'` itself, and a
permissive per-code-point default such as AnyAscii can fill the same slot.
The concern is the holder. In a published module, a package-global setter
that panics until someone calls it forces every importer's test binaries to
carry a `TestMain`, and it rules out two tables in one process. That is the
right call for dubplate's binary. It is an open API question for
`give-back`.

Fold. None in this pass. The HISTORY draft's "wrong to rediscover" section
names the tension, and the conductor routes it to `give-back`, beside the
survey's AnyAscii default.

## Checked and clean

- **Silent wrong table.** Clean. The whole-file sha256 in `Parse` refuses
  every corrupt, partial, different-version, or edited table, and the
  refusal cannot be bypassed. `internal/atomicfile.WriteFile` (temp file,
  `Sync`, explicit `Chmod`, `Rename`) means a reader sees either the old
  bytes or the new ones, whether from a concurrent install or a crash. A
  leftover `.tmp-*` file is never read. The scoring verbs are the only
  production reach into `matcher` (`shadow.go:284`, `:375`). `cmd/harness`
  does not link `matcher`. The test binaries that link it match the plan's
  overlay probe.
- **`matcher_config_hash` and shadow continuity.** Unaffected.
  `configHash` hashes `Tunables` and `Compat` only
  (`internal/engine/shadow/match.go:55-75`), and the plan changes neither.
  `matcher_commit` changes on every deploy by design. The table is
  byte-identical, so `shadow reevaluate` after the deploy restamps without
  moving a verdict. The version-bearing file name keeps a rolled-back binary
  working.
- **Supply chain.** Covered. `Build` checks two independent anchors: the
  wheel sha256, then the table sha256. The table pin equals the sha256 of the
  `.tbl` committed at rung 3, long before the 2026-09-25 fetch. A PyPI, CDN,
  or proxy compromise after that date cannot produce a wheel that passes both
  pins. Nothing unverified is unzipped, so a hostile zip never reaches
  `archive/zip`.
- **MIT at the binary.** Achieved once the embed goes. `git log --all`
  shows the `.tbl` was added once, and no other ref or `_salvage/` path
  carries a table. Exit-gate check 2 (no `go:embed`, no tracked `.tbl`)
  carries the proof. Check 3's single-string probe is corroboration only.
