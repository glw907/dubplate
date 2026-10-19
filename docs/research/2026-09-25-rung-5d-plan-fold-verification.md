# Rung 5d plan fold: verification read

Target: `docs/plans/2026-09-25-rung-5d-boundary-deployment-close.md`, revision 2 (`9177429`),
against revision 1 (`af7d941`), the fold record `2026-09-25-rung-5d-plan-fold.md`, and the three lens
reviews. The reader had no part in the fold. The plan file is unchanged at HEAD `ca41952`. No box
command was run; the box facts below come from the fold's own read and the tree.

## Counts

| Severity | Count |
| --- | --- |
| Blocker | 0 |
| Major | 5 |
| Minor | 3 |
| Owner fork | 0 |

Every blocker and major from the three lenses is closed at its cited location. Two of the fold's new
mechanisms, the clean-tree refusal and the timing gate that replaces the import-timer stop, do not work
as placed. The five majors all sit in the deploy path: Task 4D's evidence, Tasks 5 and 6's `--check`
runs, and the Task 10 sitting. None needs a new ruling.

## Question 1: did each blocker and major close?

| Finding | Where it closed | Verdict |
| --- | --- | --- |
| R-B1 = C-M1 (sitting) | Task 10 preamble `:1243-1259`, steps 5, 7, 9, 11 `:1284-1316`, rollback `:1334-1346` | Closed in substance. The timing gate sits at the wrong step (M3 below). The rollback leaves the matcher and beets mismatched (M5 below). |
| C-M2 = R-M1 (whole-estate apply) | Task 0 item 10 `:677-685`, step 6 expected set `:1289-1302`, R-5d13 moved to Task 4D `:952-954`, step 1 `:1261-1263` | Closed. The expected set is incomplete (M4 below). |
| C-M3 = K-M4 (state-dir check) | R-5d11 `:523-544`, Task 5 seeds 7 and 8 `:1021-1025` | Closed. |
| C-M4 (green baseline) | Global constraints `:317-322`, Task 2 `:795-804` | Closed. `doctorEnviron` exists at `doctor_test.go:908-918`. |
| C-M5 = K-M5 (port 8080, append) | Task 2 `:795-799`, Task 6 `:1058-1064`, step 3 | Closed. |
| C-M6 = R-M3 (header counting) | R-5d8 `:464-474`, Task 3 `:852-861`, mutations 6 and 7 | Closed. Every quote matches `go doc` (below). |
| C-M7 = K-M6 (netlink) | R-5d7 `:440-451`, Task 2 mutation 4, Task 7 `:1124-1127`, step 13 | Closed. `interface_linux.go:124` confirms `NetlinkRIB(RTM_GETADDR, ...)`. |
| K-M1 (namespace placement) | Task 5 evidence `:1003-1016` | Closed. `/dev/shm` here is mounted without `noexec`, and the shadow block has no `PrivateDevices=`, so a binary copy there can run. |
| K-M2 (405 `Allow`) | R-5d8 `:457-460`, Task 3 `:876-880` | Closed. `methodOnly` at `serve.go:380-389` answers 405 as stated. |
| K-M3 = R-M2 (same-host proxy) | Privileged-fields row `:346`, R-5d8 `:461-464`, R-5d9, Task 3 rows and mutations 4 and 5 | Closed. |
| R-M4 (unattended apply) | R-5d16 `:593-606`, Task 4D `:941-981` | Closed. It runs into the implementer's own hard rules (M2 below). |

**The three partial refusals.**

1. **Keeping the shadow line out of the loop:** sound. The re-mode was measured again here: with GNU
   coreutils `install`, `install -d -m 0700` on an existing 0755 directory leaves it 0700. On a missing
   parent it creates the parent 0755 under umask 0022 and gives only the leaf 0700.
2. **Not driving the real `render_template` from Go:** sound. `deploy.sh` sources secrets and resolves the
   host at top level (`:97-114`), so it cannot be sourced as-is. The parity rows and Task 10 step 6's
   real `--check` cover the drift.
3. **Not stopping the import timer:** sound on the mechanics. The apply's `enable --now` (`:527`) runs
   before the beets step (`:530`), and `man systemd.timer` confirms the `Persistent=true` catch-up. So a
   stop across a boundary fires an import just before `uv tool install --force`. The refusal holds only
   if its replacement, the timing gate, protects the beets step. As placed it does not (M3).

## Question 2: contradictions and build order

The 5d-i and 5d-ii split is consistent. The header (`:9-17`), Ruling 1, the token arithmetic (six
build tasks in 5d-i), the dependency list (`:1482-1485`), and the execution table (`:1469-1475`) agree.
Task 4D breaks no gate leg: no `check.sh` leg runs `deploy.sh`. The state-dir leg reads
`DUBPLATE_LIB_DIR` by pattern (`check.sh:655`), and translit's move of the pre-create block keeps that
line. Translit's apply command and its rollback are covered by the cross-lane notes. The contradictions
are majors M1, M2, and M4, plus minors m2 and m3.

## Question 3: mechanisms stated from memory

None load-bearing. Each quote was checked against `go doc` (go1.27.1) or the man page on this
workstation: `Header.Values`, `Reader.ReadMIMEHeader`, `CanonicalMIMEHeaderKey`, `ParseAddr`,
`ParseAddrPort`, `Addr.Unmap`, `Addr.String`, `net.IP.String`, `InterfaceAddrs`, the `-` prefix on
`ReadWritePaths=`, the `PrivateTmp=` and `ProtectHome=` text, `Type=exec`'s invocation failure, and the
`Persistent=` catch-up. Several sources also match: `secret-set.sh:120` (`IFS= read -rsp`),
`secretLengthFloor = 32` at `auth.go:71`, `shutdownGrace` at `server.go:52-56`, and `vite build` at
`web/package.json:8`. The one unquoted mechanism that is also wrong is step 3's `$(cat rendered)` (m1).

## Question 4: can Geoff rule, and is the rollback enough?

Rulings 1 and 2 carry enough evidence to rule on. Ruling 3 carries enough too, once M5's sentence is
added. As written, the rollback restores a working pair but leaves the beets upgrade in place, and that
is an unstated coherence gap (M5).

## Findings, ranked by consequence

### M1 (major). The clean-tree refusal blocks every `--check` run the plan gives implementers in the shared main tree

- **Location:** R-5d13 `:553-565` (the refusal runs "in both modes"); Task 4D criteria `:962-971`; Task
  5 `:1027-1029`; Task 6 `:1081-1083`.
- **Defect:** The refusal fires on any tracked edit anywhere (`--untracked-files=no` still reports
  modified tracked files), and it fires under `--check` too. Lane 1 is the main tree, and the document
  lanes share it. At this read, `git status` shows ten modified tracked files from other work
  (`Makefile`, `web/src/lib/themes/*.css`, the 6c plan, and others). An implementer's own `deploy.sh`
  edit also dirties the tree until it commits. So Task 4D criterion 1 (the no-flag run "reports each
  step as would"), criterion 3, Task 5's `--check` criterion, and Task 6's render-abort criterion each
  print the refusal instead of the steps. That holds whenever any lane has a warm edit, or whenever the
  run comes before the task's commit. Only criterion 3 says "in a clean committed tree", and even that
  cannot be arranged while a document lane is working.
- **Narrow fix:** Every `deploy.sh` evidence run, including the Task 10 sitting, runs from a detached
  scratch worktree at the task's commit. Use `git worktree add --detach <scratchpad>/5d-deploy <sha>`,
  outside `.claude/worktrees/`, and remove it after. State this once in Global constraints. Criterion 2's
  deliberate scratch edits run in that same worktree.

### M2 (major). The plan's real-box `--check` runs and `npm ci` break the implementer's own hard rules

- **Location:** Global constraints `:310-316`; Task 4D criterion 3 `:969-971`; Task 5 `:1027`; Task 6
  `:1081`; `.claude/agents/dubplate-implementer.md:43-44` ("No `ssh`, no deploy, no write to
  `5.78.148.42`, ever", "No network").
- **Defect:** `deploy.sh --check` against the box runs `ssh` (`remote_reachable`, `cat` diffs, `rsync
  -n`). After R-5d13, it also runs `npm ci`, which reaches the npm registry. The agent definition
  forbids both. An implementer either refuses and reports "could not do", which costs a re-dispatch, or
  ignores its definition. The plan does not say which wins.
- **Narrow fix:** Implementer runs target `192.0.2.1` only. That address proves every property Tasks
  4D, 5, and 6 claim: the local build lines, the step headers, and `run_remote`'s "would run" lines
  print whether or not the box is reachable. The real-box read belongs to Task 10 step 6 alone. Each
  dispatch that runs `deploy.sh` names the `npm ci` registry access as a stated exception to the
  no-network rule, or the task proves ordering with `npm ci --offline` against the existing cache.

### M3 (major). The quarter-hour gate runs at step 5, so it has expired by the time the apply reaches the beets step

- **Location:** Task 10 steps 5-8 `:1284-1307`; the fold's refusal 3.
- **Defect:** Step 5 requires the next import boundary to be at least ten minutes away. Several things
  then run before the apply reaches `uv tool install --force` at `:557`:
  - Step 6's `--check` repeats the local builds (`npm ci`, `vite build`, `make build-linux`) and every
    diff.
  - Geoff reads the report against the expected set and approves.
  - Step 7 runs.
  - Step 8 repeats the builds.
  - The apply runs every remote step before the beets step.

  The ten minutes can run out, and the import fires at the boundary plus up to 300 seconds. That is the
  hazard the fold refused the timer stop to avoid.
- **Narrow fix:** Move the boundary check to directly before step 8, after Geoff's approval. Size the
  margin from step 6's measured wall time plus the remote steps, not a fixed ten minutes. If the margin
  fails, wait for the next window. Step 5 keeps only the "no unit activating" and "no upload" reads.

### M4 (major). The expected set leaves out two lines the apply is sure to print, so the sitting stops

- **Location:** Task 10 step 6 `:1289-1301` ("any line outside it is a stop"); Task 0 item 10
  `:681-685`.
- **Defect, two parts:**
  1. `config/roster.example.json` is new since the box's `be27434` (`git diff --stat be27434 HEAD --
     config/`: three files). The `rsync --delete config/` dry run lists it. The expected set names only
     `contributors.yaml` and `reject-email.txt`. Task 0's md5 comparison walks the files on the box and
     cannot see a file that exists only on master.
  2. Translit is approved and queued next in lane 2 (STATUS `:19`, `:24`), and it cannot deploy before
     this sitting. Once it merges, the apply also prints its staged `.new` rsync, its smoke test on the
     staged path, its `systemd-run` install (a PyPI download from the box), and the `mv` swap. It also
     creates `/var/lib/dubplate/unidecode`. None of these is in the expected set.

  Either part halts the one attended sitting on a benign line.
- **Narrow fix:** Derive the `config/` part of the expected set from `git diff --name-only <box stamp>
  HEAD -- config/ compose.yaml systemd/ env/`, not from box-side md5 alone. Add a conditional block,
  "if translit merged", that names its would-run lines and the new leaf. Task 0 item 4's report selects
  that block.

### M5 (major). The rollback pairs the old matcher with the new beets, and the plan does not say so

- **Location:** "If a step fails" `:1342-1346`; step 6 `:1296-1298`; Ruling 3 `:153-177`.
- **Defect:** The `.prev` binary is `be27434`. It predates `367564c` and still carries the ft compat
  class for 2.13.1: `git grep` at `be27434` finds `internal/engine/matcher/compat.go`. The rollback
  restores that binary but deliberately keeps beets at 2.14.1. It reinstalls 2.13.1 only "if the import
  run after the apply fails". After a rollback, the shadow therefore scores an emulator built for
  2.13.1 against decisions from 2.14.1. That puts featuring-title divergences into launch-gating
  evidence, and every unadjudicated divergence blocks the launch gate. The rollback's success check,
  "one manual shadow pass reading `success`", cannot see this. Ruling 3's text does not mention the
  beets upgrade or this gap. Only the owner-touchpoints paragraph names the upgrade.
- **Narrow fix:** Make the beets downgrade part of the rollback whenever the `.prev` binary is restored:
  `runuser -u music -- env HOME=/srv/music uv tool install --force beets==2.13.1`, run while the shadow
  timer is still stopped, with the reason stated. The binary's matcher and the box's beets move as a
  pair. Add one sentence to Ruling 3: "the apply upgrades the live importer's beets to 2.14.1; a rollback
  returns it to 2.13.1 with the binary."

### m1 (minor). Step 3's `env -i ... $(cat rendered)` word-splits values

- **Location:** `:1276-1277`.
- **Defect:** A rendered value with a space (a `DUBPLATE_SMTP_FROM` display name, or a site name like
  the fixtures' `Test Family`) splits into separate words, and `env` runs the second word as the command.
  A glob character would also expand. The step never names which `dubplate-server` binary it runs. It
  says the closed port is "substituted", but the template has no port row (R-5d14), so the port is added.
- **Narrow fix:** Use `mapfile -t rows < rendered; env -i PATH="$PATH" HOME="$HOME" "${rows[@]}"
  "$REPO/dubplate-server-linux-amd64" doctor`. Say "added" for the port.

### m2 (minor). Task 8 may run during 5d-i, but it cites code at HEAD that Tasks 1-6 and the simplifier still change

- **Location:** `:1161-1162`, the execution table `:1471`.
- **Defect:** The table's `file:line` and test names for the proxy, hop, and probe findings do not exist
  until Tasks 1-3 land, and the 5d-i simplifier round moves lines after that. The lenses then grade a
  stale table.
- **Narrow fix:** Task 8 runs after the 5d-i close. If it drafts earlier, it re-greps every citation at
  the 5d-ii base before the cycle.

### m3 (minor). Once Task 4D lands, CLAUDE.md and the implementer definition describe the old tool

- **Location:** the territory wall `:296-305`; `CLAUDE.md:292` ("`--check` first"); `CLAUDE.md:145-150`
  (the state-dir leg as "the env template, the shadow unit, and `deploy.sh`", and a "pattern-filtered
  `systemd-analyze verify`").
- **Defect:** After 5d, the apply needs `--apply`, the state-dir leg covers every greenfield unit, and
  `systemd-analyze` is hard-required. Agents read CLAUDE.md, and "a rule lives where it executes" puts
  the flip there.
- **Narrow fix:** Task 11's STATUS draft routes a conductor edit of those CLAUDE.md lines. CLAUDE.md is
  conductor-owned and outside the implementer wall.

## Question 5: the four planning-miss items on the tasks the fold changed

| Task | Quoted behavior | States named | Privileged channels | Fixture state named |
| --- | --- | --- | --- | --- |
| 1 | Yes (`netip` quotes verified) | Yes (every R-5d2, R-5d3, and R-5d4 state, zone and bracket included) | Yes (the proxy set and the header, from the environment only) | Yes (limiter fixture: `Origin`, `Sec-Fetch-Site`, limit 1, unparseable body) |
| 2 | Yes (netlink quoted) | Yes (five bind states; `not checked` for `client ip` and `web`) | n/a | Yes (baseline plus one perturbation; closed port) |
| 3 | Yes | Yes (every R-5d9 state, the wedged listener included) | Yes (`PeerAddress` minus the trusted set, never `ClientIP`) | Yes (wire-level; listener asserted up before doctor) |
| 4D | `git status` flags not quoted; minor | No-flag, `--check`, both flags, dirty tracked, untracked `web/` | n/a | **Gap:** "clean tree" named for criterion 3 only and unreachable in the shared tree (M1) |
| 5 | Yes (`man systemd.exec`; `install -d` measured) | Yes (four leg states, eight seeds) | Yes (the leaves' mode and owner) | Yes (`/dev/shm`, positive half first) |
| 6 | Deferred to the task with a "Not verified" entry (acceptable) | Yes (four row states; commented rows under the no ruling) | n/a | Yes (no append; roster admin at 0600; closed port) |
| 10 | Mostly; step 3's expansion is wrong (m1) | Yes (activating, window, failure before and after step 7) | Yes (secrets name-only; argv-free mint) | **Gaps:** the timing precondition is taken at the wrong step (M3); the expected set is incomplete (M4) |
