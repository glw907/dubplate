# Rung 1 plan adversarial review, 2026-09-02

Plan: `docs/superpowers/plans/2026-09-02-rung-1-foundation.md`
Spec: `docs/superpowers/specs/2026-09-02-greenfield-server-build.md`
(revision 2.1, ratified, not under review)
Lens: mechanical completeness of the pivot against the tree as it sits
on disk at `2df8e45`, gate semantics after the rewiring, the config
core's contract, fidelity to the spec's rung 1 entry, task sizing,
test discipline.

Every claim below was checked against a file or a command run in this
worktree. Where the plan is right about a mechanism, that is recorded
too, because two of its riskiest assumptions hold and the implementer
should not re-litigate them.

---

## Blockers

**[blocker 1] `.gitignore` is not in Task 1's file list, so the pivot commit leaves 5,260 untracked files in the tree and un-ignores the build artifact it is required to produce**

Plan:57-67 enumerates what Task 1 moves, deletes, creates, and
modifies. `.gitignore` appears in none of the four lists.

`.gitignore:14-17` pins the web tree's ignores at the repo root with a
leading slash: `/web/node_modules`, `/web/dist/*`, and the negation
`!/web/dist/placeholder.txt`. `git mv web _salvage/web` (Plan:58-60)
renames the directory on disk, `node_modules` included: 139MB, 5,256
files, plus `web/dist`'s four build outputs. Verified:

```
git check-ignore -q web/node_modules/.package-lock.json           -> 0 (ignored)
git check-ignore -q _salvage/web/node_modules/.package-lock.json   -> 1 (NOT ignored)
```

The same anchoring breaks the artifact ignores. `.gitignore:11-12`
pins `/dubplate` and `/dubplate-linux-amd64`. Verified:
`git check-ignore -q dubplate-server-linux-amd64` returns 1. Task 1's
own acceptance criterion (Plan:136-137) requires `make build-linux` to
produce exactly that file, so running the acceptance step dirties the
tree with a 15MB untracked binary.

Nothing in the gate walks untracked non-Go files, so every acceptance
criterion in Plan:134-138 still passes while `git status` is
unreadable. That is the bad shape: a green gate over an unreviewable
working tree. It collides directly with Plan:50-51's "specific files,
never `git add -A`" rule, and `diff-reviewer` reads `git status` as
part of its stated method.

Fix, cheap either way: add `.gitignore` to Task 1 and re-anchor the web
patterns under `_salvage/` plus `/dubplate-server` and
`/dubplate-server-linux-amd64`; or delete `web/node_modules` and
`web/dist` before the move, since both regenerate from
`_salvage/web/package-lock.json` and nothing greenfield builds them
until rung 7.

**[blocker 2] Repointing deploy.sh's artifact name makes the deploy path overwrite the live box binary with the greenfield skeleton, and the boundary the plan claims is not drawable in the file**

Plan:119-126 instructs: "Update every reference from the old
`dubplate-linux-amd64` artifact to `dubplate-server-linux-amd64` ...
The box's existing units and their installed binary are untouched
until rung 4 and later; only the artifact plumbing moves now."

That boundary does not exist in `scripts/deploy.sh`. The local artifact
name and the remote install target are the two ends of one rsync:

- `deploy.sh:262` runs `make -C "$REPO_ROOT" build-linux`
- `deploy.sh:263` sets `DUBPLATE_BINARY="$REPO_ROOT/dubplate-linux-amd64"`
- `deploy.sh:265-267` rsyncs `$DUBPLATE_BINARY` to
  `$SSH_TARGET:/usr/local/bin/dubplate`
- `deploy.sh:268` chmods it
- `deploy.sh:277` smoke-tests `/usr/local/bin/dubplate version`

Rename the source and the next `deploy.sh` run installs a binary whose
entire surface is `version`, `check`, and `doctor` over
`/usr/local/bin/dubplate`. The smoke test passes, because the skeleton
has `version` (Plan:88-90), so `set -euo pipefail` at `deploy.sh:2`
raises nothing. Every guardian unit then fails at its next fire:

- `systemd/dubplate-import.service:55` -> `/usr/local/bin/dubplate import`
- `systemd/dubplate-backup.service:41` -> `backup`
- `systemd/dubplate-disk-check.service:24` -> `check disk`
- `systemd/dubplate-navidrome-check.service:24` -> `check navidrome`
- `systemd/dubplate-ping-fail@.service:20` -> `ping-fail %i`

Those are precisely the four timers Spec:47-49 and Spec:757-758 say
protect the library and the beets database throughout the build.

"No deploy runs in this rung" (Plan:126) is a promise about this pass,
not about the repo. The next rung that runs `deploy.sh` is rung 4, the
shadow deployment, and its author will read Plan:119-126 as licence to
run it. `--check` does not warn: `deploy.sh:250-256` documents that the
cross-compile runs under `--check` too, and the only signal is one
`rsync -n` line naming a target path that has not changed.

The spec does ask for the repoint (Spec:284-286), so the plan's
fidelity is literal. What the plan owed was the observation that a
literal repoint is not additive, against Spec:759-760 ("Rungs 1 through
3 ... touch the box not at all") and Spec:548-556's additive shadow
unit. Two honest resolutions: split the binary step so `deploy.sh`
installs `dubplate-server-linux-amd64` to a new
`/usr/local/bin/dubplate-server` path beside the untouched incumbent
(which is what an additive rung 4 actually needs); or leave the binary
step alone at rung 1 and record in the plan that the deploy path is
frozen until rung 4 wires the new unit. Either way the plan must name
what moves, what stays, and why, rather than asserting a boundary the
file does not have.

**[blocker 3] The plan's owned wall is four keys; the spec's owned wall has five parts, and the missing one is the identity source**

Spec:141-149 states the owned, required, no-defaults set as "the
library root, the state directory, `SiteName`, the admin identity,
and—because login is server purview (ruling 7)—an active identity
source", with the correlated config of whichever source is active also
required, and closes: "login has one wall shape,
owned-with-a-selectable-source, not two."

Spec:165-166: "Rung 1 derives the complete table from this
classification and it becomes the config core's contract."

The plan's `Config` (Plan:162-167) carries `LibraryRoot`, `StateDir`,
`SiteName`, `AdminEmail`. The behavior table (Plan:186-196) has no
identity-source row. Task 3 then hard-codes the arity: "four `ok`
lines" (Plan:268).

Deferring the *implementation* to rung 5 is defensible and expected.
Dropping the row from the derived table is not, because Plan:216-218
makes the package doc "the wall's contract of record" and that document
is what rung 5 builds against. A selectable source is not another
string field; it is a discriminator plus a required-when-selected
block, and discovering that at rung 5 reopens the wall's shape. That is
the retrofit the spec's foundation-first framing exists to prevent.

Fix: add the identity-source row to the derived table and the package
doc, marked "the selector and its correlated blocks arrive at rung 5",
and stop stating doctor's owned-line count as a literal four.

**[blocker 4] `Load` has only two classes, owned and unknown, so every not-yet-arrived integration key becomes the exact U5 failure the task is implementing against — and Task 3's doctor contract already assumes three classes**

Plan:192 makes any unrecognized `DUBPLATE_*` key an "unknown
configuration key" error. Plan:197-203 seeds a superseded map with
three keys. There is no third bucket.

Task 3 already needs one. Plan:254-255: doctor prints "one line per
known integration class from the wall table, each reading `disabled
(arrives at rung N)`". So rung 1 knows the integration classes by
name. `Load` does not, and a known integration key present before its
rung lands falls through Plan:192 into "unknown". That message is the
precise class Spec:159-163 refuses: "load errors that name their
replacement, never a silently degraded start."

The estate's instance shows the size of the gap. `env/dubplate.env.template`
names 36 `DUBPLATE_*` keys. Four survive as owned; three are in the
plan's superseded map; the remaining 29 have no disposition:
`DUBPLATE_SMTP_*` (:28-32), `DUBPLATE_R2_*` (:36-40),
`DUBPLATE_NAVIDROME_*` (:47-49), the working directories (:78-84),
`DUBPLATE_TRIGGER_DIR` (:91), `DUBPLATE_UPLOADS_DIR` (:96),
`DUBPLATE_PING_*` (:117-120), and the rest.

The worst single instance is a rename, not a removal:
`env/dubplate.env.template:78` is `DUBPLATE_LIBRARY_DIR`; the plan's
owned key is `DUBPLATE_LIBRARY_ROOT` (Plan:188). Anyone moving an
existing deployment to the new binary gets "unknown configuration key
`DUBPLATE_LIBRARY_DIR`" with no replacement named. The plan implements
U5 and reproduces its failure mode in the same commit.

Fix: three buckets. Owned; known-but-not-yet-arrived, whose error names
the integration and its rung ("`DUBPLATE_R2_BUCKET` is backup-target
config; the backup integration arrives at rung 6"); and genuinely
unknown. Seed the superseded map from the template's real names, with
`DUBPLATE_LIBRARY_DIR` -> `DUBPLATE_LIBRARY_ROOT` first.

---

## Majors

**[major 1] Task 2's superseded map omits `DUBPLATE_SERVE_SITE_NAME` while Task 4 performs exactly that rename**

Plan:198-200 seeds the map with `DUBPLATE_ACCESS_TEAM_DOMAIN`,
`DUBPLATE_ACCESS_AUD`, and `DUBPLATE_PUBLIC_NAVIDROME_URL`.
Plan:322-327 mints `DUBPLATE_SITE_NAME` to replace
`DUBPLATE_SERVE_SITE_NAME`, which is live at
`env/dubplate.env.template:107`. One plan, two tasks, contradictory
supersession lists. The rename Task 4 performs is the one Task 2's map
does not carry.

**[major 2] Task 4's age-store chore breaks deploy.sh's env render and silently reverses a recorded STATUS decision**

Plan:322-327 instructs: mint `DUBPLATE_SITE_NAME`, "mark the old name
for deletion at the next registry sweep."

`deploy.sh:175-192` reads every non-comment `NAME=` line out of
`env/dubplate.env.template` and requires each to resolve to a non-empty
value from the sourced shell environment, exiting 1 with a list
otherwise. `DUBPLATE_SERVE_SITE_NAME` is `env/dubplate.env.template:107`
and Task 4 changes only the template's header (Plan:312-317). Deleting
the age-store value at the next sweep therefore makes the next deploy
fail at the render step, before it reaches the box.

The comparison Plan:326 draws to `FILEBROWSER_ADMIN_PASSWORD` does not
hold: that name is orphaned, referenced by nothing in the template.
This one is load-bearing.

`docs/STATUS.md:44-45` currently records the opposite ruling:
"DUBPLATE_SERVE_SITE_NAME ("907 Music") stays minted for the new
server." Reversing it may be right, but the plan reverses it without
noticing there is anything to reverse.

Secondary: nothing consumes `DUBPLATE_SITE_NAME` until rung 4 renders a
greenfield env file, and the workstation rule is wire the name first,
then mint. This chore is scheduled a rung early.

**[major 3] The ops audit step instructs the implementer to delete an accurate operational record**

Plan:114-118: "`grep -ri filebrowser scripts/ systemd/ cloud-init/
compose.yaml config/ env/` must return nothing after the delete. If it
returns hits, remove them in this task."

Run today, after deleting `config/filebrowser.yaml`, it still returns a
hit: `config/cloudflared-tunnel-config.json:2`, whose `_comment`
records the 2026-09-02 wind-down verbatim ("the uploader's path-scoped
/upload rule and the FileBrowser inbox hostname are both retired; until
the greenfield dubplate-server launches, the tunnel serves Navidrome
only"). That file is the record of what gets PUT to the Cloudflare API
(`CLAUDE.md:96-99`), and that sentence is the only in-repo statement of
why the tunnel's ingress is Navidrome-only right now.

Spec:275-276 asks rung 1 to *audit* the ops keep set, not to empty a
grep. Restate the step: enumerate hits, classify each as orphaned
configuration or historical record, delete only the former, report both
lists. As written, the criterion "the filebrowser grep is empty"
(Plan:137-138) can only be satisfied by damaging the record.

**[major 4] `newCheckCmd(environ, out io.Writer)` cannot satisfy `check`'s own stderr contract, and neither command's exit path is specified**

Plan:240-244 fixes the constructor at one writer and calls it "the
constructor shape every later subcommand follows". Plan:248-251 then
requires `check` to print success to stdout and "the load error to
stderr". One injected writer, two streams: the signature and the
contract cannot both be right, and the wrong resolution propagates into
every later rung's subcommands.

Separately unspecified: "exits 1" requires `main` to `os.Exit(1)` on
`cmd.Execute()`'s error, and a machine-facing precondition gate needs
`SilenceUsage` and `SilenceErrors` or cobra prints the whole usage
block on every failed `check`. Plan:86-92 (main.go) and Plan:274-276
(Task 3's steps) mention neither, and the test table (Plan:266-270)
asserts only on the returned error, so nothing catches a `check` that
dumps usage text into a systemd journal.

**[major 5] Two of the five doctor test cases are not assertable as written**

Plan:268: "doctor | full valid wall | four `ok` lines, every
integration line says `disabled`; nil error". The plan never enumerates
the integration classes or their rung numbers anywhere. The implementer
must invent the set and then assert against its own invention, which is
the vacuous-assertion shape Plan:45-46 bans.

Plan:252-259 gives the line shape as "aligned key, status word,
optional reason" with a sample (`DUBPLATE_LIBRARY_ROOT  ok`,
Plan:253) but no column width and no separator rule, while claiming
"this scaffold fixes its output shape". A test written from that will
fall back to `strings.Contains(out, "ok")`, which passes against almost
any output, including a doctor that prints nothing but a summary line.

Fix: put the integration-class table into the plan (name, flag key,
arriving rung), and pin the format literally, for example
`fmt.Fprintf(out, "%-32s %s", key, status)` with `: reason` appended.

**[major 6] "Threshold registry" and "enable-flag discipline" are spec deliverables the plan reduces to two free functions and a doc comment, without saying so**

Spec:481-488 lists rung 1's config deliverables as "the derived wall
table, the enable-flag discipline, the threshold registry, and the
loud-compat rule (U5)".

What the plan delivers: the wall table, partially (blocker 3); U5,
partially (blocker 4); enable-flag discipline as prose inside a package
doc comment (Plan:216-218); and, for the threshold registry,
`parseDuration` and `parseInt` (Plan:176-183).

Two parse helpers are not a registry. Nothing declares which thresholds
exist, what their defaults are, or what they mean, and nothing lets
`doctor` report them — which is what Spec:125-126's universality test
("Every tunable is a threshold with a sane default") needs to be
checkable rather than aspirational. Rungs 3, 5, and 9 each add
thresholds; without a registry each one invents its own convention.

Deferring is a legitimate choice. Deferring silently is not, and
Plan:225-226's "Deliverable count: 3" is computed against the reduced
scope, which is exactly how the deferral stays invisible to a reviewer.

Related, on the helper shape the review was asked to attack: taking
`environ []string` per call means each helper re-scans the whole slice
and each independently re-decides the duplicate-key and empty-value
questions. The salvage tree already solved this with a single `get`
choke point plus an `errs *[]error` accumulator
(`internal/config/config.go:189-200`), documented there as the seam a
later config-file source plugs into. Spec:322-337's reuse stance
invites adopting that shape on merit; the plan neither adopts it nor
says why it declines.

**[major 7] Task 1 carries at least six independently rejectable deliverables under a declared count of four and a no-additions clause**

Plan:140-142 declares four and forbids additions. The actual step list
(Plan:79-132) is: annotated tag; four `git mv`s; delete a config file;
author a new cobra entrypoint; rewrite the Makefile; edit `check.sh`;
edit `vale-comments.sh`; `go mod tidy`; a live-ops grep audit with a
remediation clause; a `deploy.sh` rewrite; run the gate; commit.

Three of those a reviewer can reject on their own merits with no
reference to the rest: the skeleton's Go, held to the exemplary bar
(Plan:33-35); the `deploy.sh` change, production-adjacent (blocker 2);
and the ops audit, which is judgment work, not mechanical (major 3).
Counting "tag + move + audit" as one deliverable is what gets the total
to four.

The cheapest split repairs blocker 2 as a side effect. Task 1a: tag,
moves, `.gitignore`, skeleton, Makefile, `check.sh`,
`vale-comments.sh`, tidy — one mechanical commit that is genuinely
reviewable as a unit and whose diff is dominated by renames. Task 1b:
the ops audit and the deploy path, where both the judgment and the
production risk live, reviewed on its own.

**[major 8] The spec's "re-derived subcommand surface" deliverable is absent, not deferred**

Spec:481-486 names it in rung 1's ladder entry; Spec:316-318 specifies
the method: "re-derived at rung 1 from the organizer spec's subcommand
list, minus `beet` passthrough, which never exists on the new server."
The source list is
`docs/superpowers/specs/2026-09-01-dubplate-organizer-design.md:181-185`:
`serve`, `import` (poll mode and `--session`), `nudge-consume`,
`backup`, `restore`, `check`, `doctor`, `ping-fail`, `version`, `beet`.

Plan:91-92 says only "No other subcommands yet—Task 3 adds `check` and
`doctor`; the rest arrive with their rungs." There is no derivation, no
recorded surface, and no rung assignment per verb, not even as a doc
comment. This is the artifact rungs 4, 6, 9, and 12 each need to know
which verbs they own and which names are already spoken for. Skipping
it once costs four re-derivations, each free to disagree with the
others. A ten-line doc comment on `main.go` discharges it.

---

## Minors

**[minor 1] The Makefile edit leaves `WEB_DIR` behind and `check.sh`'s gate label lies**

`Makefile:2` is `WEB_DIR := web`, referenced only by the four `web-*`
targets at `Makefile:30-40`. Plan:96-97 deletes the targets and their
comment block, not the variable.

`scripts/check.sh:32` is `echo "== make check (Go + web) =="`. That is
an echo, not "the web references in the header comment" that Plan:100-102
scopes the edit to, so it survives verbatim and prints a false label on
every gate run for the next six rungs.

**[minor 2] `.claude/settings.json` carries the same drift Task 4 reconciles in CLAUDE.md**

`.claude/settings.json:6` permits `Bash(bats tests/*)` — `tests/` is
empty and the suite retired 2026-08-31. `:14-15` permit
`Bash(npm run build*)` and `Bash(npx vite build*)`, both dead after the
pivot until rung 7. Task 4's reconciliation (Plan:302-311) scopes to
`CLAUDE.md` only. Cheap to fold in.

**[minor 3] The gitleaks answer is available now, and the plan can carry it**

Verified across `.claude/`, `scripts/`, `Makefile`, `.git/hooks/`,
`docs/`, and `CLAUDE.md`: the only matches for "gitleaks" are
`CLAUDE.md:107` (the claim itself) and Plan:307 (the instruction to go
look). There is no invocation in `scripts/check.sh:1-115`, no
pre-commit config, and no `.github/` directory in this repo at all. The
claim at `CLAUDE.md:107` is false and the line should be removed.

Plan:305-310 sets this up correctly as an investigation with a
report-either-way clause, which is the right shape. It just does not
need to be an investigation.

**[minor 4] The TDD step as written yields a build failure, not failing tests**

Plan:207-209: "Write the table-driven tests for `Load` from the
contract table; run them; confirm every case fails (no implementation
yet)." With no `Load` in the package, `go test` fails to compile, no
case runs, and no case is individually shown to fail. That does not
satisfy Plan:45-46's "prove each test can fail". The step needs a
declared stub returning a not-implemented sentinel, so each row's
failure is its own assertion firing.

**[minor 5] The empty-value rule is stated for the wall and left undefined for the helpers, along with duplicates and `=` in values**

Plan:196 rules `DUBPLATE_STATE_DIR=""` as missing. Plan:178-179 gives
`lookup(environ, key) (string, bool)` without saying what the bool
means for an empty value, and specifies `parseDuration`/`parseInt` only
for "unset". `DUBPLATE_SOMETHING_TIMEOUT=` is undefined behavior in a
contract whose entire purpose is that behavior is defined. The salvage
tree already answered it: `internal/config/config.go:196-200` treats
`!ok || v == ""` as unset.

Also unstated: duplicate keys in `os.Environ()` form (last-wins is the
convention and the table should say so), and a value containing `=`,
which `strings.Cut` handles correctly but which no row exercises. Rung
5 puts base64 secrets through this path.

**[minor 6] `AdminEmail` normalization is unspecified, and `net/mail.ParseAddress` is looser than the table implies**

Plan:191 validates with `net/mail.ParseAddress`; Plan:165 stores
`AdminEmail string`. Measured behavior:

```
"not-an-address"             ERR  mail: missing '@' or angle-addr
"Geoff <geoff@907.life>"     OK   Name="Geoff"   Address="geoff@907.life"
"geoff@907.life (comment)"   OK   Name="comment" Address="geoff@907.life"
"\"weird name\"@example.com" OK   Name=""        Address="weird name@example.com"
"GEOFF@907.LIFE"             OK   Name=""        Address="GEOFF@907.LIFE"
```

The table's bad-email row holds. What the plan does not say is which
string lands in `Config.AdminEmail`: the raw value, which can carry a
display name and a comment, or `addr.Address`. Spec:655-660 makes
`AdminEmail` the admin authorization identity at rung 14 and
Spec:428-429 requires emails lowercased before keying, so an unnormalized
field means rung 5 and rung 14 will disagree about what an equality
comparison means. State it: `addr.Address`, lowercased.

**[minor 7] The vale probe is a real discriminator but not a repeatable test**

The mechanism the plan reasons from is correct, and it was worth
checking. `scripts/vale-comments.sh:49-52` builds its list from
`git ls-files '*.go'`, and git's default pathspec matching makes `*.go`
match at any depth (76 files today, all nested), so `_salvage/**/*.go`
would re-enter the scan. `.vale/styles/glw907/EmDash.yml:3` is
`level: error`, which `scripts/vale-comments.sh:54`'s
`--minAlertLevel=error` gates on. The probe would genuinely go red
without the exclusion.

What it is not is a test. Plan:108-110 describes a one-time manual
edit-and-revert with no artifact and nothing to re-run when rung 2
touches the script. Either commit a `_salvage/`-resident fixture and a
self-test path, or call it a one-time verification rather than an
"acceptance probe".

**[minor 8] Two acceptance criteria cannot fail — and the mechanism they gesture at does hold**

Plan:128-129 requires "`go vet ./...` reports nothing in `_salvage/`".
Verified against a scratch module containing `_salv/pkg/p.go` with a
deliberately undefined symbol: `go list ./...`, `go build ./...`,
`go vet ./...`, and `golangci-lint run ./...` are all clean and all
blind to it. The criterion is a tautology and certifies nothing.

The finding worth recording is the positive one, because it is the
plan's largest mechanical bet and the previous spec revision died of an
unverified one. The toolchain half of the pivot is sound: underscore
directories are excluded from every wildcard package pattern, so
`golangci-lint` will not see `_salvage/`, `go mod tidy` resolves the
same patterns and therefore prunes to cobra's set exactly as
Plan:111-113 claims, `go test ./...` will not try to compile
`e2e/`'s imports of `internal/lock` and `internal/importer`
(`e2e/backup_test.go:10`, `e2e/import_test.go:18-19`), and
`golangci-lint run ./...` on a single-main-package module is clean. The
systemd leg also survives untouched: `scripts/check.sh:99`'s filter
pattern `/usr/local/bin/[A-Za-z0-9_.-]+` matches `dubplate-server` as
readily as `dubplate`, and rung 1 changes no unit anyway.

Keep Plan:136-137's "no Go package outside `cmd/dubplate-server`
visible to `go list ./...`" — that one can fail, if a stray `.go` file
is left at the root.

**[minor 9] STATUS is already at 59 of its 60 lines**

`wc -l docs/STATUS.md` returns 59. Plan:318-321 adds a rung-1 next
action and Plan:333 asserts "STATUS ≤ 60 lines". The task will have to
relocate content to HISTORY, which the ledger rule requires anyway
("pruning means moving, never deleting"), but the plan does not say so
and an implementer under a line cap summarizes instead. Name the line
to reclaim: `docs/STATUS.md:40-42` still reads "Open rulings 1-3 in the
spec await Geoff", and Spec:691-693 records all three as resolved.

**[minor 10] The one-commit instruction and the per-task code-simplifier instruction are in tension, unacknowledged**

Plan:131-132 mandates a single commit for the whole pivot. Plan:348-350
mandates "`code-simplifier` over the changed Go before the final commit
of each code task". For Task 1 the changed Go is roughly 15,000 lines
of pure rename plus one new ~40-line `main.go`. Running
`code-simplifier` over that diff is either a no-op or actively
hazardous, since the moved files are `_salvage/` and Plan:39-40
declares them read-only from the moment they exist.

The single-commit rule is right — a half-moved tree does not build, so
there is no green intermediate to commit at. Say that, and scope the
simplifier pass explicitly to `cmd/dubplate-server/main.go`.

---

## The three I would fix first

1. **Split the deploy path out of Task 1 and make it additive**
   (blocker 2, major 7). This is the only finding with a production
   consequence: as written, the repo ships a `deploy.sh` that installs
   a three-verb skeleton over `/usr/local/bin/dubplate` and breaks the
   four guardian timers, and the trap springs at rung 4, not here.
   Fixing it forces the Task 1 split that major 7 wants anyway, so one
   edit buys both.

2. **Re-derive the wall table completely, with three classes**
   (blockers 3 and 4, major 1). The identity-source row, the
   known-but-not-yet-arrived bucket, and a superseded map seeded from
   `env/dubplate.env.template`'s real names, `DUBPLATE_LIBRARY_DIR` ->
   `DUBPLATE_LIBRARY_ROOT` first. This table is the artifact eleven
   later rungs build against, and it is the one deliverable Spec:165-166
   explicitly says rung 1 must get complete.

3. **Add `.gitignore` to Task 1** (blocker 1). Sixty seconds of work,
   and it is the difference between a pivot commit a reviewer can read
   and 5,260 untracked files with a green gate over the top.

**Verdict:** the mechanical spine of this plan is verified sound — the
underscore-directory exclusion, the tidy prune, the vale gap, and the
systemd filter all behave exactly as the plan reasons they will, which
is more than the last artifact reviewed here could claim — but it ships
a deploy path that would clobber the live box's binary, a config wall
missing a required class and an entire classification bucket, an ops
audit step that orders the deletion of a true record, and a Task 1 that
declares four deliverables while carrying at least six; fix the four
blockers and split Task 1, and this is ready to execute.
