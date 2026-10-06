# Re-plan document pass (rungs 12 and 16) Implementation Plan

Status: **revision 2, 2026-10-05.** Revision 1 (`15be3fd`) went through one self-folding `spec-plan-review` lens
(`docs` class); its findings and their dispositions are in
`docs/superpowers/research/2026-10-05-replan-document-pass-review.md`. It carries no ruling for Geoff: every fork it
touches was ruled in the spec.

**Spec:** `docs/superpowers/specs/2026-10-05-rungs-12-16-replan-design.md`, revision 2.1 at `33b465c`, approved
by Geoff 2026-10-05 with all three owner rulings ruled yes ("the spec"). Its "Document changes" section (items 1
to 9) is this plan's scope list, and "Preserving the archive (now)" is Task 0.
**Fold record:** `docs/superpowers/research/2026-10-05-rungs-12-16-replan-fold.md`, "Errata owed" (items 1 to 7).
**Parallel code pass:** `navidrome-removal` (`docs/superpowers/plans/2026-10-05-navidrome-removal.md`, being
drafted beside this one), lane 1, main tree.

**Pass class:** `docs`. Task 0 is a conductor ops chore with no repo diff and no class.
**Executor:** `dubplate-implementer`, `model: sonnet`, for Tasks 1 to 8 and Task 10's ledger edits. Task 0 is
dispatched by the conductor to a `general-purpose` agent at `model: sonnet` with the exact command constraints
below. Task 9 (memory files) is the conductor's own edit, per the spec's doc change 9.
**Per-task gate:** `cairn-run-gate 'bash /var/home/glw907/Projects/dubplate/scripts/check.sh'` before each commit,
re-issued on exit 75 until it prints `gate exit:`. This is the full gate on the heavy lane: it builds `web/`,
launches Chromium for the rendered gates, and runs `internal/backup`'s `rclone` tests. `CAIRN_GATE_LANE=light` is
not allowed, because the gate is not browserless. `CLAUDE.md` requires `bash scripts/check.sh` "before every
commit". The script has a lighter form, `check.sh --reduced`, "the comment-only and docs-only form", but its
header rules it out here: it is "never the final gate on master", and "a change that reaches master has had the
full form run over it". Every task in this pass commits straight to master, so every commit takes the full form.
The gate cost is accepted and priced into the ceiling.
**Review:** `diff-reviewer` per task at Sonnet, against the task's acceptance criteria, blocking only on an unmet
outcome or a factual error against the spec or the tree. Specs, plans, `CLAUDE.md`, and the ledgers take no prose
or register review (spec-plan-review: specs and plans are not register-graded). The one operator-facing document,
`docs/server-runbook.md`, takes one `prose-voice-reviewer` read at the close. No `code-simplifier` (a `docs` pass
never takes one).
**Ceiling:** 3.5M execution tokens, flag 2.8M (80%). Planning and the plan-review lens are recorded outside it.
**Checkpoint interval:** each segment boundary (A ends after Task 4, B after Task 8), and the close. STATUS is
written at each.
**Segments:** A (Task 0, then Tasks 1 to 4) runs now, beside the removal pass. B (Tasks 5 to 9, then Task 10, the
close) waits for "the merge": the removal pass's last code commit on master with its full gate green, counting
its close's simplifier and review-fold commits (its `--ff-only` merge, if its plan works on a branch).
**Execution mode:** nine repo tasks, so each segment runs through `pass-execute` by name, one invocation per
segment, `passClass: 'docs'`, with `gate` set to the per-task gate string above and `reducedGate` set to the same
full-gate string. Without that arg, the runner reduces a fix round whose findings are all comment-only (every
finding in a docs pass) to its class default, "the repo's type check plus only the test files this fix round
touched", which this repo does not define and which would put a commit on master without the full form. Tasks
inside a segment are file-independent but run one at a time (see "Sharing the main tree").
**Dependency sweep:** this pass edits no dependency manifest. The pass-opening `dependency-upgrade` sweep is the
removal pass's (the spec's "The removal pass"); this pass's record is one line pointing at it. A second sweep
would collide on `go.mod`.
**0.8 forward hooks and prior art:** neutral. This pass builds nothing, lands no hook, and adopts no library. It
edits the hooks list itself (Task 2), as the spec directs.

## Coordination with `navidrome-removal`

### File territory

The removal pass owns, per the spec's "The removal pass": `cmd/`, `internal/`, `web/src/`, `web/scripts/`,
`config/roster.example.json`, the `Makefile`, `go.mod` and `go.sum` (its sweep), `docs/server-runbook.md`'s two
bootstrap commands (the rename's grep acceptance covers the runbook), and its own close's writes to
`docs/STATUS.md` and `docs/HISTORY.md`.

This pass owns: the five specs under `docs/superpowers/specs/` named below, `CLAUDE.md`, `ROADMAP.md`, `IDEAS.md`,
`docs/server-runbook.md` (after the merge), `docs/STATUS.md` and `docs/HISTORY.md` (after the merge), and the two
memory files.

**Pre-flight (before segment A's first dispatch):** once the removal plan is committed, the conductor's pre-flight
agent diffs that plan's per-task Files lists against Tasks 1 to 4's files. Any overlap moves the overlapping task
to segment B. A removal plan that edits any `docs/superpowers/specs/` file is the likeliest hit. The same agent
reports whether the removal plan commits on master in the main tree, as the lane registry's lane 1 row ("main tree
(master)") says, or on a branch. If it uses a branch, segment A finishes before that branch is created, so no
document commit lands on it.

### What runs now and what waits

| Task | Runs | Why |
|---|---|---|
| 0 Archive preservation | Now, first, before any other box work | The spec orders it before any other box work. That includes the removal pass's read-only precondition reads on the box (doctor's `store` row); the conductor sequences those after Task 0. |
| 1 Build spec errata | Now | No code-pass file. |
| 2 Streamer spec rev 3 | Now, before slice 1's plan is authored | Slice 1 (the study lane) is the spec's only other writer: it answers the open items owed to the study rung. Rev 3 lands first so slice 1 writes against it. |
| 3 Configuration homes erratum | Now | No code-pass file. |
| 4 Group-scale principles erratum | Now | No code-pass file. |
| 5 Rung 6 spec erratum | After the merge | The removal pass's precondition check decides whether the in-place `0003` edit happened at all, and the erratum cites its commit. |
| 6 Runbook | After the merge | The removal pass edits the same file's bootstrap commands. |
| 7 `CLAUDE.md` | After the merge | The spec: "The ledgers and `CLAUDE.md` land after the removal pass merges." |
| 8 ROADMAP and IDEAS | After the merge | Ledgers (`CLAUDE.md` "Ledgers"); IDEAS' beets-check entry cites the removal pass's pre-deletion commit. |
| 9 Memory files | After Task 7 | They point at Task 7's paragraphs. |
| 10 Close (STATUS, HISTORY, runbook prose read) | After the removal pass's own close | One ledger writer at a time. |

Until Task 7 lands, the slice 2 lane's launch cites the spec's ruling 5 for the three-code-lane grant.

### Sharing the main tree

The document lane shares the main tree and its branch (master) with lane 1. The full gate runs over the whole
tree, so a code-lane task's warm edits can redden a document task's gate. The rules:

- Segment A runs before lane 1's first `pass-execute` segment launches, or else at a lane 1 segment boundary. A
  `pass-execute` invocation runs its tasks back to back, so no between-task window exists inside a lane 1 segment.
  At dispatch, lane 1's last task is committed and `git status` shows no warm change outside this pass's files.
  Lane 1's next segment does not launch while segment A is in flight.
- Each task stages and commits only its own named files, never `git add -A`.
- A red gate whose failing leg sits in a code-lane file is not a document-task failure. The implementer reports it
  and stops without committing; the conductor re-dispatches after lane 1 commits.
- The slice 2 lane works in its own worktree and touches no file here.

## Tasks

Every repo task keeps superseded text as written and adds dated erratum blocks, the repo's practice (the build
spec's header errata are the model). The exception is Task 2: the spec calls for the streamer spec's revision 3,
which removes and rewrites named passages. Its revision-3 changelog records each removal, and git history keeps
the text. Every erratum and changelog entry cites the spec by path, revision 2.1, and `33b465c`, and names
"Geoff, 2026-10-05". Every section an erratum names must resolve to a heading or passage in its file (checked by
grep, listed in the report).

### Task 0: Preserve the retirement archive (conductor chore; box read-only)

**Outcome.** The one copy of `/root/musicbox-retired-2026-10-05.tar.gz` (0600, root) on `5.78.148.42` has a
verified, age-encrypted copy on the workstation, and STATUS records where it is and what was checked.

**Constraints.**

- Read-only on the box, over `root@5.78.148.42` with key auth (`.claude/instructions/api-access.md`): a listing,
  a hash, and a copy pulled to the workstation. No command writes, moves, or deletes anything on the box, and no
  unit is started or stopped. The report lists every command run on the box.
- Quiet-box check first, read-only: `systemctl list-units --state=activating 'dubplate-*'` empty, no upload in
  progress, and the lane registry showing no other box work in flight. A failed check stops the chore.
- Encryption uses the workstation age store's public recipient, the `RECIP=` value hard-coded in
  `~/.dotfiles/scripts/secrets/secret-set.sh` (a public key, safe to print). The encrypted copy lives outside every
  git repository, at `~/.local/share/dubplate/archive/musicbox-retired-2026-10-05.tar.gz.age`, mode 0600. (The spec
  says "beside the secrets store"; `~/.dotfiles/secrets/` is a tracked git directory, so "beside" is read as "under
  the same key, outside the repo". The archive holds members' addresses and listening history and must never be
  staged.)
- That copy is not backed up anywhere: the workstation runs no backup job. Until rung 16, the two copies are the
  box's plaintext tar (root, 0600) and this `.age` file. The spec gives the second off-workstation copy to rung
  16's backup remote, and Task 10 carries that to rung 16.
- Decryption for the integrity check uses the session-cached identity the store's tooling maintains
  (`/dev/shm/.age-key-<uid>`). If the cache is cold, warm it once through the store's own tooling
  (`secret-set.sh --sync`, one 1Password fetch). Never read, print, or copy the key.
- Every plaintext copy on the workstation (the pulled tar, the extracted database) lives in the session scratchpad
  and is removed with `shred -u` once its check passes. The scratchpad sits on tmpfs (`/tmp`), so plaintext never
  reaches disk; the archive's size from step 1 must fit the scratchpad's free space twice over (the tar plus the
  extracted database), or the chore stops and reports. No database content, member address, or secret value is
  printed. The database is queried only with `PRAGMA integrity_check` and `SELECT count(*) FROM sqlite_master`.

**Steps the record must show** (the spec's six, made checkable):

1. The tar listing shows every member whose name ends in `navidrome.db`, with its `-wal` and `-shm` files if
   present (paths and sizes recorded). The old stack staged a backup snapshot beside the live database, so more
   than one copy may be present; each is checked in step 4. None present is a stop.
2. The archive's sha256, computed on the box.
3. The pulled copy's sha256 on the workstation equals step 2's, before encryption. The `.age` file's own sha256 is
   recorded too.
4. A round trip from the `.age` file: decrypt, stream through `tar`, and extract each step 1 database (with its
   `-wal` and `-shm`) into the scratchpad. The decrypted stream's sha256 equals step 2's. For each extracted copy,
   its size is nonzero, `SELECT count(*) FROM sqlite_master` is greater than zero, and `PRAGMA integrity_check`
   prints `ok`. (SQLite opens a zero-byte or missing file as an empty database, whose integrity check also prints
   `ok`; the size and schema count keep the check from passing on nothing.)
5. Whether the old stack's R2 `navidrome.db` snapshot still exists, by a read-only listing of bucket
   `music-library` (`.claude/instructions/api-access.md`) under `state/` and `_archive/state/` (the old backup's
   layout, `_salvage/internal/backup/backup.go`). The listing uses the Object Read only credential
   `MUSIC_R2_RO_*` (kept in the age store, registry entry of that name) or `$CLOUDFLARE_API_TOKEN`; the command
   sources `~/.local/secrets` and prints no value. Record each `navidrome.db` key with its size and date, or
   "absent". HISTORY records an earlier listing that found `state/` empty. If no credential can list the bucket,
   that is the recorded result.
6. `git -C ~/.dotfiles status --short` and `git -C ~/Projects/dubplate status --short` show no new path.

**Record.** The conductor writes it into STATUS at its next STATUS write (the conductor is STATUS's single writer
while lane 1 runs): the `.age` path, both sha256 values and their agreement, the listing summary, the integrity
result, the R2 result, and the date. Task 10 carries it into HISTORY. A failure at any step is a stop: the box copy
stays the only copy, and the conductor reports before any other box work.

**Acceptance.** Each of steps 1 to 6 has its evidence in the report. The two hashes match. Every extracted
database has a nonzero size, a nonzero schema count, and an integrity result of `ok`. No plaintext remains in the
scratchpad (a `find` over it shows none). STATUS carries the record.

### Task 1: Build spec errata

**Files:** `docs/superpowers/specs/2026-09-02-greenfield-server-build.md`.

**Outcome.** The build spec reads correctly under the spec's rulings without losing its history (doc change 1).

**Acceptance.**

- One header-level erratum, dated 2026-10-05, in the existing header-errata style. It states that every passage
  on Navidrome, a driver, SSO, the forward-auth front, the fallthrough, or the estate shadow is superseded by the
  spec, and lists the passages by section: "Drivers own driver knowledge", the serve-as-ingress and forward-auth
  trust boundaries, "One owned prefix", the post-login destination, proxy hygiene, the tree shape's
  `internal/streamer/`, rung 4's "continuous to launch", the two-tracks text, the interim-estate matcher fallback,
  "Box safety during the build", and the 2026-09-25 `/share/*` erratum.
- The erratum supersedes the estate shadow (the floor, launch bar part 3, rung 4's "continuous to launch", the box
  side), and, under Owner ruling 1's yes, the tree shape's "first-class leaf" sentence (`:459-461` at `33b465c`),
  ruling 1's adopter clause (`:976-988`), and the adopter text in "Shadow: the standing oracle".
- Erratum blocks at rung 12 (the built-in streamer, slices 1 to 8), rung 15 (the shadow record leaves the
  evaluation), rung 16 ("First launch (0.8.0)", first install, no cutover, the ordering constraints by reference
  to the spec), rung 17 (the beets deprecation window gone; archive disposal decided here), launch bar part 3
  (struck; parts 1, 2, and 4 gate; completeness includes the streamer), and "Shadow: the standing oracle".
- The status line's errata count is updated. `git diff --numstat` shows one deletion, and `git diff` shows that
  the deleted line is the status line: no original sentence is removed or reworded.

### Task 2: Streamer spec revision 3

**Files:** `docs/superpowers/specs/2026-09-07-streamer-2-0-design.md`.

**Outcome.** The streamer spec is the governing document for rung 12, with every passage that assumed an
incumbent, a shadow, or a flip removed or rewritten (doc change 2).

**Acceptance.**

- The status line reads revision 3, 2026-10-05, and a "Revision 3, 2026-10-05" section beside "Revision 2.5"
  lists every removal and rewrite below with the spec's reason. The "Re-keyed to the 0.8 release" note and the
  earlier revision sections stay as written.
- Removed: the shadow period, the flip, the `shareRole` doctor row, the escape-hatch text, and the arm64 walk
  "before the flip" (the arm64 build stays).
- Slice 9 is deleted from "The ladder". Slices 1 to 8 are stated as rung 12's passes. Slice 8's rationale for the
  import (`:675-676` at `33b465c`) is rewritten to the first-launch import.
- "Relation to the current ladder" and the header's "sequenced after the greenfield ladder's launch" are rewritten.
  The rejected alternatives are marked superseded, with ruling 1's reason.
- "Cutover and operations" is rewritten to the spec's configuration properties: always on, no enable flag, no
  integration class, no driver, no cloudflared, tunables as registered thresholds with defaults, secrets on the
  auth-secret pattern generated by slice 8's `init`, and a per-slice delta to the operator-supplied set. The
  heading may be renamed only if `git grep` shows no other document cites it by name.
- Slice 8 gains: the four import criteria (dry-run report before any write; match by address, then by username
  under `COLLATE NOCASE`, refusing with a report on disagreement; per-user writes in one transaction or keyed so a
  rerun is a no-op; the import reads an extracted copy); the unit's `CAP_NET_BIND_SERVICE` in TLS mode only; the
  certificate leaf (unit `ReadWritePaths`, the `deploy.sh` pre-create, the state-dir agreement leg); and Owner
  ruling 3's acceptance line (one short-lived VPS with no age store and no `deploy.sh`, from the release artifact
  through `init` to a played track; each unclear step blocks slice 8's close or goes to `IDEAS.md` by owner
  ruling).
- Slice 6 records the bearer bypass's two contract sentences (a bypassed handler must not read the session cookie;
  the pattern registers twice because either registration alone leaves one form refusing) and that it is
  path-keyed.
- Forward hooks: hook 6 and hook 15's capability clause are marked retired in place. Hook numbers never change,
  because other documents cite them. Capability reporting is dropped unless a slice names a consumer. Hook 19 is
  re-homed to slice 8. Hook 5's field is `username`.
- Open items: shadow length and Navidrome retention are marked moot, and "whether the study rung runs beside the
  current ladder" is marked answered (slice 1 runs now as a document lane). No other open item is edited.
- The pinned Navidrome and gonic containers stay as workstation-only behavior oracles, under the GPL
  behavior-only rule, stated once.
- Rulings 1 to 7 and Owner rulings 1 to 3 are recorded with their date.
- Two passages outside the named sections also assume a cutover and are rewritten to the first-launch import
  (ruling 2): "Hosting and setup"'s "The group's own cutover uses this command" (`:553` at `33b465c`), and the
  prior-art table's "Cutover import" cell (`:797`).
- `grep -niE 'shareRole|escape.hatch|before the flip|slice 9|shadow period|cutover'` hits only the revision
  history sections, the rev 3 changelog, the rulings record, the open item marked moot in place, and the "Cutover
  and operations" heading if it keeps its name. The report lists every hit with its section.

### Task 3: Configuration homes spec erratum

**Files:** `docs/superpowers/specs/2026-09-11-configuration-homes-design.md`.

**Acceptance.**

- One dated erratum in the header-errata style (doc change 3): the rung 12 Navidrome verifications are moot; the
  shadow-pair erratum of 2026-09-23 is moot; rung 14's private gate (`:221-224` at `33b465c`) is re-keyed from a
  driver capability report to slice 3's projector enforcing visibility on every endpoint, and private collections
  ship at first launch per the existing 2.0 ruling.
- The erratum states that the bearer admin mount (`:123-126`) stands, because the bearer bypass stays.
- The status paragraph's errata count is updated. That paragraph wraps over several lines, so the diff may reflow
  it. Additions only, apart from that paragraph.

### Task 4: Group-scale principles spec erratum

**Files:** `docs/superpowers/specs/2026-09-25-group-scale-principles-design.md`.

**Acceptance.**

- One dated erratum (doc change 4): "Before 0.8" and row A13 are moot; the delegated route class and its rows (the
  route inventory and acceptance row A2, which name the fallthrough as the delegated row) have no subject until
  slice 3; "Versions this spec speaks to" reads under ruling 1 (the first launch is 0.8.0 on the built-in streamer,
  and no 0.x pre-release exists).
- Every `Navidrome` hit in the file is either still true (the import, prior-art evidence) or covered by the
  erratum. The report lists each hit with its disposition.
- The status line notes the erratum. Additions only, apart from that line.

### Task 5: Rung 6 spec erratum (after the merge)

**Files:** `docs/superpowers/specs/2026-09-23-rung-6-ops-seam-store.md`.

**Precondition.** The removal pass merged. If it edited `0003_catalog.sql` in place, this task runs. If its
precondition check failed and it amended with a new migration, this task makes no edit, and Task 10 records in
HISTORY that the exception went unused.

**Acceptance.**

- One dated erratum, at the header and at "Migrations": a one-time exception to the append-only rule for
  `0003_catalog.sql` and its fixtures `0003` to `0005`, for the username rename only. It names the two
  preconditions (doctor's `store` row read `absent` on the box; backup never enabled, so the remote holds no store
  snapshot), the master commit that made the edit and re-pinned `migrationHashes`' `0003` entry, and the guarantee
  that any older store fails loudly at open on the ledger's sha256 check.
- The append-only rule's own text is unchanged. Additions only, apart from the status line.

### Task 6: Runbook (after the merge)

**Files:** `docs/server-runbook.md`.

**Outcome.** The runbook carries the username rename, the one-time roster step, and no forward-auth topology
(doc change 8; Owner ruling 2 yes).

**Acceptance.**

- Both bootstrap commands use `--username`. The removal pass should already have made this edit; the task verifies
  it, and makes it only if the merge left it.
- A named step, written for any operator whose roster file was written by a build before the rename (a fresh
  install never meets it), with the spec's three actions: move `roster.json` aside to `roster.json.pre-username`
  inside the leaf (0600, `music`-owned); run `member add <address> --admin --name ... --username ...` through
  `systemd-run`, which takes the bootstrap path; check that `member list` shows the one admin and doctor prints no
  `roster` row. It carries no estate value: no address, hostname, or path outside the documented leaf.
- "The forward-auth topology" section is deleted, and every link to its anchor is removed or re-pointed (a `grep`
  for `forward-auth-topology` outside `docs/superpowers/` and `docs/HISTORY.md` returns nothing; plans, specs, and
  research records keep their history).
- `grep -niE 'forward-auth|navidrome|streamer[_ -]?username|driver'` over the runbook returns nothing, or only
  generic hits the report lists (a database driver, for example). The trusted-proxy sections and the example
  fronts, cloudflared among them, stay as generic hardening.
- The edits follow the runbook's existing standard (Google developer documentation style), so Task 10's prose read
  has nothing structural to fix.

### Task 7: `CLAUDE.md` (after the merge)

**Files:** `CLAUDE.md`.

**Outcome.** The repo's agent instructions match the spec (doc change 6), in the file's existing register:
Anthropic's Claude Code guidance for project memory (short, specific, imperative, verifiable statements), with
dated provenance in the file's "(Geoff, date)" form.

**Acceptance.**

- The versions paragraph: the first launch is 0.8.0, on the built-in streamer; 0.9 and 1.0 keep their meaning; the
  existing reading rule for older documents gains one clause, that documents before 2026-10-05 describe a 0.x
  launch and a rung 16 cutover under the earlier plan and stay as written.
- The build-for-0.8 paragraph: rung 12 is the streamer's slices 1 to 8; every other rung plan still weighs the
  forward-hooks list per task; hook 6 retires and hook 5 is renamed. The ROADMAP id `streamer-0-8` is no longer
  cited as a live entry.
- The standing grants: up to three code lanes when useful and efficient (Geoff, 2026-10-05), each code lane past
  the primary in its own git worktree.
- The touchpoint list: the rung 16 cutover sitting becomes the first-launch sitting, and the shadow parenthetical
  goes.
- The review-scaling line names streamer (12) and first launch (16) in place of "streamer/SSO (12), and cutover
  (16)".
- `grep -niE 'cutover|SSO|0\.x|two code lanes|driver seam|Navidrome still'` on `CLAUDE.md` hits only the
  versions paragraph's reading-rule clause for older documents (which names their 0.x launch and rung 16 cutover).
  Every remaining `Navidrome` hit is true under the rulings (the import, prior art, "matches no streamer"). The
  report lists every hit of both.
- The file's length changes by no more than ten lines.

### Task 8: ROADMAP and IDEAS (after the merge)

**Files:** `ROADMAP.md`, `IDEAS.md`.

**Acceptance.**

- `streamer-0-8` is folded into the Active greenfield entry as rung 12. Its still-live settled decisions, the 2.1
  list, the web-player note, and the TypeScript 7 recheck move into the greenfield entry or stay under a pointer.
  Its superseded passages (the "Sequenced after launch (rung 17)" sentence, with its driver seam and escape hatch)
  leave ROADMAP: the report
  quotes each one verbatim, and Task 10 moves them into HISTORY's re-plan entry. Nothing is deleted (pruning means
  moving). The Planned heading for it is gone.
- The greenfield paragraph is rewritten: first launch 0.8.0 on the built-in streamer, no 0.x pre-release, no
  shadow record in the launch bar, no box swap, the evidence track ending at corpus and conformance.
- The organizer-spec fold bullets on the Navidrome reference driver and on Cloudflare Access as a forward-auth
  front move to `IDEAS.md`. The `packaging` entry's ingredient list drops Navidrome, and `formats-1-0`'s
  "after `streamer-0-8`" sequencing is re-pointed.
- `IDEAS.md` gains three entries under Ideas, in the file's format (what it is, why a group would want it, what it
  builds on, source): external streamer integration (ruling 4's reasoning); a forward-auth front (Owner ruling 2's
  reasoning: music apps authenticate with app passwords a front cannot gate); a beets-migration confidence check
  (Owner ruling 1), naming the `pre-greenfield` tag and the removal pass's pre-deletion commit as its source.
- `grep -nE 'streamer-0-8|0\.x|shadow record|box swap|reference driver|escape hatch' ROADMAP.md` hits only Done
  entries or dated history inside an entry, each listed in the report.

### Task 9: Memory files (conductor; after Task 7)

**Files:** `~/.claude/projects/-var-home-glw907-Projects-dubplate/memory/version-numbering.md`,
`build-for-2-0.md`, and their two `MEMORY.md` index lines. Outside the repo: no gate, no commit.

**Acceptance.** `version-numbering.md` records the 2026-10-05 change (first launch is 0.8.0 on the built-in
streamer; no 0.x pre-release) and points at `CLAUDE.md`'s versions paragraph for the rule, keeping only the why.
`build-for-2-0.md` points at `CLAUDE.md`'s build-for-0.8 paragraph and notes rung 12 is the streamer's slices.
Both frontmatter descriptions and both index lines match. Neither file restates initiative state.

### Task 10: Close (after the removal pass's own close)

**Files:** `docs/STATUS.md`, `docs/HISTORY.md`.

**Steps and acceptance.**

1. One `prose-voice-reviewer` read over Task 6's runbook diff, operator audience, findings ranked by consequence.
   Blockers fold through one `dubplate-implementer` re-dispatch; polish may be refused with a one-line reason.
2. STATUS: the moot carries in the spec's "Where the carried items land" table leave Open items, and each live
   carry names its landing (rung 16, a slice, or 6d). The username owner question and the re-plan spec's approval
   question are gone. Task 0's record stays, and rung 16's carries gain the archive's second copy (the `.age` file
   to rung 16's backup remote, per the spec's "Preserving the archive"). The lane registry shows lane 1, the slice
   1 document lane, and the slice 2 worktree lane as they stand. STATUS gets no longer; anything moved to reach the
   60-line target goes to HISTORY.
3. HISTORY, newest first: the re-plan (the spec, the rulings, the fold, and this pass's commits by task), the
   archive record from Task 0, Task 5's outcome, and the superseded `streamer-0-8` passages Task 8's report
   quoted. It records `readFLAC`'s salvage pointer and the `tagio` exported-unused warning only if the removal
   pass's own entry does not already carry them. "Wrong to rediscover" names that the streamer spec's hook numbers
   are stable with retired hooks kept in place.
4. The budget score: tokens against the ceiling, and attended time as planning misses and sittings (none planned).
5. The resume prompt points at the next action the conductor names (6d revision 2.3 or the next lane).

## Not in this plan

- **6d revision 2.3.** The spec moves 6d to revision 2.3 after the removal pass merges, but it is a plan revision
  with its own approval gate on 2.3 and its own review, not a document change. It runs as its own dispatch once
  the removal pass merges.
- **Slice 1's answers** to the streamer spec's open items owed to the study rung (the spec's V8: one writer per
  item).
- The architecture-cleanup inventory, rung 16's sequence, and any code.
