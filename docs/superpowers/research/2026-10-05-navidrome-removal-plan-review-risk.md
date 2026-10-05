# Navidrome removal plan review: domain risk lens

Target: `docs/superpowers/plans/2026-10-05-navidrome-removal.md` at `7036ebc` (revision 1).
Spec: `docs/superpowers/specs/2026-10-05-rungs-12-16-replan-design.md` (revision 2.1).
Sibling: `docs/superpowers/plans/2026-10-05-replan-document-pass.md` (revision 2, `1c37e40`).
Lens: security of the serve/auth surface, data integrity, operations and lanes, install ease.
Reviewer: Opus 5.5, read-only against the tree at `7036ebc`. No ssh was run.

Findings are correctness gaps ranked by consequence, then over-ceremony ranked by cost. A fold may refuse any
finding whose fix costs more than the risk it removes.

**Counts:** High 0, Medium 4, Low 6, over-ceremony 1. No owner fork.

## What the lens checked and found sound

These need no fold. They are recorded so the fold does not re-derive them.

- **No live protection is weakened by Task 1.** Production builds the strip as
  `serve.NewStrip("", auth.SessionCookieName, proxyConfig, serve.FrontExemption{})`
  (`cmd/dubplate-server/serve.go:324`), and `git grep 'SetFallthrough('` finds no production caller (only the
  definition at `internal/serve/serve.go:147`). So the driver header is always `""`, the exemption is always the
  zero value (`headerFor` returns `""`, `strip.go:250-258`), and every request outside the prefix already reaches
  `http.NotFound` (`serve.go:173-176`). Removing the three pieces deletes code paths production never took. The
  origin check (`origin.go:103-116`), group 1 and 2 stripping, the trailer scrub, group 3's keep set, and the
  bearer bypass are untouched in behavior. The session cookie now reaches no handler outside the prefix, because
  none exists.
- **`auth.Seam`'s doc change is doc-only** (`internal/auth/seam.go:17-22`): one interface, one implementation, no
  behavior.
- **A retired `DUBPLATE_ACCESS_*` key still refuses.** `rejectNamespace` turns every superseded name into a load
  error (`internal/config/config.go:313-316`: `w.errs = append(w.errs, fmt.Errorf("%s: %s", name, msg))`), so
  `check` fails today and fails after Task 3; only the message text changes. The deleted reserved rows move from
  "reserved for streamer, arrives at rung 12" to "unknown configuration key", both refusals. The env template
  (`env/dubplate-server.env.template`) carries none of these keys, so the box's rendered env is unaffected.
- **Nothing in this pass reads or writes the box.** Task 0's count runs the binary on the workstation. The close's
  "live auth smoke" is the rendered gate, which runs locally. `deploy.sh` never runs the binary ("It never enables,
  starts, or restarts a unit", `scripts/deploy.sh:36`), so even an unplanned deploy of the post-pass master reads no
  roster. A later operator running a member verb against the old roster gets a loud `ErrRosterUnparseable`
  (`internal/auth/roster.go:340-341`), and the bootstrap path refuses an existing file (`ErrRosterExists`,
  `roster.go:268`), so no verb can overwrite it. No committed binary store fixture exists (`git ls-files` finds no
  `.db`), so the in-place `0003` edit strands nothing in the tree.
- **`check` is read-only on its state directory.** It was run against an empty scratch directory (see PR4); the
  directory was still empty afterwards.
- **The plan-time ssh before the archive chore.** The spec orders the archive chore "before any other box work" to
  keep the only copy safe from a write. The drafter's reads (`ls`, a names-only `grep -oE`, `systemctl
  list-timers`) cannot touch `/root/musicbox-retired-2026-10-05.tar.gz`. It is a sequencing deviation, not a risk.
  Task 4a's re-confirmation is the right control: it makes the plan-time evidence non-load-bearing. PR1 is about
  when the re-confirmation runs.
- **Install ease.** No task adds a required key, a setup step for a fresh install, or a site-specific default
  outside `internal/config`. `--username` stays required, as `--streamer-username` was. The roster step affects
  only an operator holding a pre-rename file, and the document pass's runbook task writes it.

## Correctness gaps

### PR1. Medium. Task 4a's re-confirmation cannot run "before Task 4a" inside segment B

- **Where:** plan lines 46 (segment B is Tasks 3, 4a, 4b), 142-143, 414-416, 555 (conductor action 4).
- **Defect:** segment B runs through one `pass-execute` invocation, and the document plan states the runner's
  mechanics: "A `pass-execute` invocation runs its tasks back to back, so no between-task window exists inside a
  lane 1 segment" (sibling plan, "Sharing the main tree"). The conductor therefore cannot run the read-only
  re-confirmation between Task 3 and Task 4a, and cannot "amend this task before dispatch" on a failure. As
  written, Task 4a's precondition is either run before segment B (not what the text says) or skipped. The commit
  message must also cite "the re-confirmation's date", which the implementer can only cite if the dispatch carries
  it.
- **Fold:** move conductor action 4 to "before segment B launches": run the re-confirmation then, and on a failure
  re-cut Task 4a as `0006` before launching. Put the re-confirmation's date and its command output summary in Task
  4a's dispatch text so the commit message can quote it. In Task 5's HISTORY outcomes, add one clause recording
  that the plan-time read preceded the archive chore and the re-confirmation followed it.

### PR2. Medium. The slice 2 lane's merge into master runs in the main tree and is not sequenced

- **Where:** plan lines 74-75, 544; `CLAUDE.md` "the merge ritual: merge token, rebase, full gate in-tree,
  `--ff-only`, full gate on master".
- **Defect:** the ritual's `--ff-only` and its full gate on master run in the main tree, which this pass and the
  document lane share. Neither plan says when slice 2 may merge. A merge during an in-flight lane 1 or document
  segment moves `go.mod`, `go.sum`, and slice 2's packages under a running implementer, whose gate then proves a
  tree that differs from what lands, and whose file-scoped commit lands on a master no gate has run as a whole.
  That is the one path in this plan to a red master. It also breaks `CLAUDE.md`'s "Never launch a second executor
  into the SAME tree".
- **Fold:** add one bullet to "Where it runs, and the shared tree": slice 2 merges only at a lane 1 segment
  boundary with no document segment in flight. The conductor holds the merge token and runs the full gate on
  master before the next segment launches. Add the same line to "Contested surfaces" under the slice 2 row.
  Optionally also add: the PATH-resolved binaries Task 0 may bump (`golangci-lint`, `check.sh:142`; `rclone`) are
  shared with the worktree's gate, so Task 0 tells slice 2's conductor before bumping one.

### PR3. Medium. Task 0 commits on master in the shared tree but sits outside the sharing rules

- **Where:** plan lines 46 ("Task 0 (conductor, before any segment)"), 67-68, 161-164; sibling plan "Segments"
  ("A (Task 0, then Tasks 1 to 4) runs now, beside the removal pass").
- **Defect:** the rule "A segment here does not launch while a document segment is in flight" covers segments, and
  Task 0 is not one. The document lane's segment A may legitimately be in flight when the conductor starts the
  sweep, which edits `go.mod`, `web/package.json`, and possibly `web/node_modules` through `npm ci --prefix web`,
  with a full gate per bump. A document task's gate that runs while `node_modules` is being rewritten, or against
  warm `go.mod` edits, goes red or proves a tree that never lands. Two executors are then in one tree.
- **Fold:** state in Task 0 that it runs only when no document segment is in flight: either before document
  segment A launches or at its boundary. Each bump's commit lands before the next document dispatch.

### PR4. Medium. Task 0's "empty environment" must be named exactly, or the baseline undercounts

- **Where:** plan lines 166-170, 406.
- **Defect:** the conductor's shell is not empty. In this review's own Bash shell, `env | grep -c '^DUBPLATE_'`
  printed `41`; `~/.local/secrets` exports DUBPLATE_* values. Run without `env -i`, `check` resolves most required
  keys from the inherited environment, and it may refuse on a superseded or reserved name instead. The baseline
  then reads low or fails. Every streamer slice's "nonzero delta needs an owner ruling" (spec, "The streamer's
  configuration") keys on this number, so a low baseline hides a real delta.
- **Evidence (measured here at `7036ebc`):** with `env -i HOME=... PATH=/usr/bin:/bin` and a scratch state and
  library directory, round 1 reports five keys (`LIBRARY_ROOT`, `STATE_DIR`, `SITE_NAME`, `ADMIN_EMAIL`,
  `IDENTITY_SOURCE`). Round 2 reports nine (`ROSTER_FILE`, `MAGIC_LINK_SECRET`, `SESSION_SECRET`, `PUBLIC_URL`,
  and the five `SMTP_*`). Round 3 prints `configuration: ok` and `store: absent`. The run ends at 14 keys; the
  plan's "not run to the end" can become "14, ends in three rounds", which Task 0 re-confirms after the sweep.
- **Fold:** write the method as `env -i HOME="$HOME" PATH=/usr/bin:/bin ./dubplate-server check`, with every
  path-valued key pointed at a scratch directory, and record the key set by name. Restate Task 3's ruling 7
  criterion (line 406) to match: "the iterated required-key set, run by Task 0's method, equals Task 0's set".
  As written, "the required-key count `check` reports with an empty environment" reads as round 1's five.

### PR5. Low. A stopped archive chore leaves Task 4a with no path forward

- **Where:** plan lines 142-143, 414-416; sibling plan Task 0 "Record": "A failure at any step is a stop: the box
  copy stays the only copy, and the conductor reports before any other box work."
- **Defect:** if the archive chore stops, the re-confirmation never runs, and Task 4a, segment B, and the pass's
  close wait on an owner sitting. The plan names a fallback only for a failed re-confirmation, not for one that
  cannot run.
- **Fold:** one sentence: if the re-confirmation cannot run because the archive chore stopped, Task 4a takes the
  `0006` path, which needs no box read. The conductor amends Task 4a's grep acceptance for that path, since the
  original `0003` keeps `streamer_username`.

### PR6. Low. The baseline counts keys, but the spec asks for "values a default install's operator supplies by hand"

- **Where:** plan lines 166-170; spec "The streamer's configuration".
- **Defect:** a default install also supplies values outside `check`: the admin bootstrap's `member add` arguments
  (address, `--name`, `--username`, required per `member.go`) and the `unidecode install` step. A slice that adds a
  required bootstrap flag or a setup verb would show a zero delta against a keys-only baseline.
- **Fold:** record the key set as the count and list the non-key inputs by name beside it, uncounted, so a slice
  plan can state a delta to either. One extra line in the pass record.

### PR7. Low. "The merge" leaves the removal close's code-touching steps unsequenced against document segment B

- **Where:** plan lines 504-517; sibling plan "Segments" ("B ... waits for 'the merge': the removal pass's last
  task committed on master") and its table (Task 10 "After the removal pass's own close").
- **Defect:** the sibling table distinguishes "the merge" from "the removal pass's own close", which implies
  document segment B may launch before the close. But the close's steps 1 to 4 (`code-simplifier`, the security
  fold, the architecture-reader folds) edit Go in the main tree after Task 4b commits. A document segment
  launched in that window is a second executor in one tree.
- **Fold:** one line in "Where it runs": the close's steps 1 to 4 and any fold task they spawn count as an in-flight
  lane 1 segment. Document segment B launches after the last of them commits green.

### PR8. Low. Two re-pointed serve tests become vacuous on their named property

- **Where:** plan lines 224 and 229; `internal/serve/serve_test.go:160-175`, `:559-577`.
- **Defect:** `TestOrderPrefixRequiresSegmentBoundary` registers `foo` and requests `testPrefix+"foo"`. With the
  boundary check broken, that path enters the owned mux, matches neither `/dubplate/foo` nor `/dubplate/foo/`, and
  reaches the terminator's 404. The re-pointed assertions ("owned spy never invoked, status 404") pass either way,
  which is what the spec forbids: "re-pointed at a registered spy subtree with a 'spy reached' check, or at a
  status a 404 cannot also produce". `TestUnclaimedOwnedPathReachesTheTerminator` loses the one check that told the
  terminator from the outside handler; both now answer 404. The consequence is nil, since both destinations answer
  the same, and `TestRoute`'s row "a neighbor sharing the prefix's bytes without a segment boundary does not
  enter" (`prefix_test.go:163-164`) carries the boundary property with its own `next` spy. This is still the
  security reviewer's first question ("vacuous re-pointed tests first").
- **Fold:** delete `TestOrderPrefixRequiresSegmentBoundary`, its sole remaining subject being the fallthrough
  spy, and cite the `TestRoute` row in the report. Rename the terminator test to what it still proves (an owned,
  unclaimed path answers 404 and invokes no subtree).

### PR9. Low. Task 4b edits auth input validation under `sweep`, outside the security reviewer's scope

- **Where:** plan lines 19-23, 446-447, 508-510.
- **Defect:** Task 4b renames the roster loader's strict key, the roster API's field allowlist (`rosterFields`), and
  three wire codes in `internal/auth/admin.go` and `roster.go`. That is admin-session input validation. Task 5
  scopes `web-auth-security-reviewer` to Task 1's diff only. Task 4b's named mutations (drop
  `RejectUnknownMembers`, add `streamer_username` to `rosterFields`) cover the main risks, so the gap is small.
- **Fold:** add `internal/auth/roster.go` and `admin.go`'s Task 4b diff to the same security reviewer dispatch.
  It costs no extra dispatch.

### PR10. Low. "The config tests that exercise `Optional`" can be misread as the identity source's `Optional`

- **Where:** plan lines 93-97, 281-283, 294-295.
- **Defect:** two unrelated `Optional` fields exist. `IntegrationClass.Optional` (`internal/config/keys.go:188-189`)
  is Reading 3's target. `IdentitySource.Optional` carries the rotation secrets (`cmd/dubplate-server/auth.go:57`,
  read at `config.go:371` and `doctor.go:508`) and must stay. A Sonnet implementer told to remove "the
  Optional-threshold pairing" and the tests "that exercise `Optional`" could touch the wrong one. Every
  `DUBPLATE_*_SECRET_PREVIOUS` would then refuse as unknown, which breaks a secret rotation's overlap. It would
  fail loudly, never silently.
- **Fold:** write `IntegrationClass.Optional` in the Files line. Add one acceptance line: a load carrying
  `DUBPLATE_SESSION_SECRET_PREVIOUS` with the magic-link source selected still succeeds, so an existing test
  that proves it is named in the report.

## Over-ceremony

### OC1. Low cost. Architecture reads on comment-only packages

- **Where:** plan lines 104-105 (Reading 6), 513-516.
- **Cost:** `internal/engine/tag` (non-test edits are comments plus one test row), `internal/engine/tagio`, and
  `internal/engine/matcher` change only comments in non-test source. That is three Opus dispatches grading
  packages whose code did not move, and findings from them route to the cleanup inventory anyway.
- **Fold (optional):** read ruling 6's "touched" as a change to non-test code beyond comments, and record the
  reading. The fold may refuse this, since ruling 6 is Geoff's and literal; the cost is small either way.
