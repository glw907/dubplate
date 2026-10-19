# Rung 6d plan rebase: revision 2.2 to 2.3

The fold record for `docs/superpowers/plans/2026-09-26-rung-6d-identity-serve-store.md` revision 2.3,
2026-10-06. Revision 2.2 (`5ed193f`) was cited at `20f4a7c`, with some lines at `314c0c8` and
`985ff3a`. Revision 2.3 cites everything at `0cb33ae`. Every pointer below was re-resolved there with
`git show`, `grep -n`, `go doc`, and `man`; no line was carried from revision 2.2 unread.

## What moved the plan

1. Rung 5 closed (5d-i, 5d-ii, the sitting deploy), so the plan's "5d will rewrite" framing and
   every 5d-plan citation went.
2. The navidrome-removal pass (`2fcb60f` to `5831313`): `users.streamer_username` became `username`
   (`2f4d052`, edited in place under the spec's `0003` erratum), the roster, API, CLI, and page say
   `username`, `newServeShell` returns `(*serve.Shell, error)`, `Prefix.Route` and the fallthrough
   are gone, and no route registers a bearer bypass.
3. The re-plan spec (revision 2.1): 0.8.0 on the built-in streamer, the tunnel and Navidrome retired,
   hook 6 retired and hook 5 renamed, and its "Where the carried items land" table, which sends
   enabling the unit, 6c's Ruling 2 (the `deploy.sh` halt on `check`), and the store identity (if
   Ruling 3 is yes) to rung 16, and the username re-validation to 6d.
4. The rung 6 spec's `0003` erratum (`93ef26b`) shifted its line numbers: +1 after line 5, +19 after
   line 427. The CH spec gained a 2026-10-05 erratum that touches no 6d surface.

## Changes by plan section

**Header.** Status line set to revision 2.3, awaiting approval; History bullet added. One citation
basis stated (`0cb33ae`). Pass class added (`auth-data`, `docs` for Task 9). Ceiling re-derived:
10.5M, flag 8.4M (was 11.5M, 9.2M). Ten tasks and three segments kept.

**Sources.** The 5d plan dropped. Added: the re-plan spec, the runbook, 5c-iv's and 5c-ii's handoffs,
the removal pass's reads. Every spec line range re-cited.

**What this pass is / is not.** Import replaced by the key's retirement (Ruling 4). "Is not" gains:
no unit enable or restart, no `check` step in `deploy.sh` (rung 16), no dual-key epoch (R-6d20).

**Rulings.** See "Rulings" below. Ruling 4 added.

**Owner touchpoints.** Rewritten to the box as it is (binary `d1e93e9`, unit disabled, `roster` leaf
with one admin under the old key, no store leaf). The upgrade is now: rollback copies and `--apply`,
one supervised start and stop, the roster file moved aside, `member add --admin`, `member list`,
`doctor`. Revision 2.2's `member import` steps and 5d sitting-step references removed.

**Token ceiling.** Re-derived on 6c's rate (0.88M), with the removal pass's actual as the evidence for
a half-rate task. Task 0 0.25M (sweep added), Tasks 5 and 8 at 0.44M. Projection 8.13M. Variants
re-derived for Ruling 3's and Ruling 4's no. Deliverable count 28 across Tasks 1 to 8 (was 30).

**Code pointers.** Every row re-cited. New rows: `check.go`, `doctorprobe.go`, `main.go`'s blocks,
`templatekeys_test.go`, the test environments setting the key, the deploy files' exact lines, the
runbook, the page's dead conditionals and `duplicate_principal` mapping, `deck.ts:150`.

**Behaviors table.** Re-quoted at `0cb33ae`. Added: `errNoDeadline`, `encoding/json/v2`
`RejectUnknownMembers`, `rejectNamespace`'s empty-value rule, `Registry.claim`, `evaluateTableAbsent`,
`principalCollision`, `OpenServe`'s logger refusal, `context.WithTimeout`. Removed (no 6d consumer
left): `crypto/sha256.Sum256` (appendix only), `readRoster`'s mode rule (appendix only), `systemd-run`
EXIT STATUS, `/proc/locks`, `systemctl try-restart`, the SQLite trigger probe (appendix only),
`RefreshSession` folded into the session row. `users` quote now reads `username`.

**Global constraints.** R-6d1 gating replaced by lane placement beside slice 2's worktree. Added the
architecture bar (re-plan ruling 6) and the ease-of-setup line (ruling 7: required keys 14 to 13,
typed keys unchanged).

**Privileged fields.** Import channels and the `member_imports` row moved to the appendix. Added the
"any `users` row" row (R-6d22).

**Resolutions.**

- R-6d1 rewritten: rung 5 closed; no 5d gating.
- R-6d2: preparation function hands back the handle's close; outage window re-cited.
- R-6d3: `Restart=on-failure`, `RestartSec=5`, `StartLimitBurst=30` read from the landed unit (was
  "Task 0 reads 5d's value").
- R-6d5: string sites re-cited (`store.go:290-307`).
- R-6d10: `streamer_username` to `username`; the address fallback, `errDuplicatePrincipal`, and
  `outcomeDuplicatePrincipal` now deleted (every row carries a username); the demote grep result
  restated (two comments only).
- R-6d11: moved to the appendix (Ruling 4's no).
- R-6d12: message depends on Ruling 4; `privateFiles.roster` removal moved to Task 4 and extended to
  `doctorprobe.go`.
- R-6d13: the members row prints unconditionally after the store row, with an R10 line-count
  constant; the import-record guard moved to the appendix.
- R-6d14: `Principal` re-justified by S20's Subsonic username and the re-plan's ruling 3 (the re-plan
  asked for this); Navidrome's auto-creation reason dropped.
- R-6d15: rewritten. `deploy.sh` runs no store-opening verb at 6d; the `runuser` clause routes to
  rung 16; the `checkOwner` hint rewording stays.
- R-6d17: three stale `web/` items listed (dead `username !== ''` conditionals, `duplicate_principal`,
  `run.ts` comments).
- R-6d18: rewritten to the current runbook: install order moves the unit start before the first
  admin with no added step; the re-create section is replaced by the 6d upgrade section; a rollback
  note points to the pre-6d section at `0cb33ae`.
- R-6d19: rewritten. `deploy.sh` already never starts a unit; only the leaf list changes. The
  `try-restart` order, the presence check, and the `check` holder line are dropped as guarding a
  restart that does not exist; 6c's Ruling 2 and the holder clause route to rung 16.
- New R-6d20: revision 2.2's Ruling 2, taken as a resolution.
- New R-6d21: `check` and `doctor` evaluate the seed against an empty row set when no settings table
  exists (6d-E2).
- New R-6d22: existing usernames meet the write path's rules by construction (the re-plan's 6d item).
- New R-6d23: `rosterResetServer`'s baseline is a named address set (5c-iv's "make the snapshot
  explicit").

**Deviations table.** 6d-R1 retired. 6d-R4 conditional on Ruling 4. 6d-R5 and 6d-R10 now route to
rung 16. 6d-R11 re-evidenced by the removal pass's box read and the `0003` ledger check. 6d-R12 moved
to the appendix. Added 6d-E2, 6d-R13 (criterion 4 names `member add`), 6d-R14 (the rename and
`Principal`).

**Review focus.** Import line moved to the appendix; added the single insert site and R-6d21.

**Task 0.** Renamed; Part 1 is the `dependency-upgrade` sweep (CLAUDE.md, "Current dependencies").
Part 2 rescoped: blockers are a warm tree, a vanished pointer in a file changed since `0cb33ae`, a
production `Bump` or demote path, the migration count, a sweep-moved library row, and the box's state
(read-only). Revision 2.2's items on 5d's landed files and spend removed. Two deliverables.

**Task 1.** `buildServeShellWithContext` replaced by `buildServeShellWithLogger` and the direct call at
`serve_test.go:596-610`; temp-dir ownership premise added; a foreign-owner state row added.

**Task 2.** Citations only.

**Task 3.** Migration moved to the appendix; `ctx` threading sites re-cited (`handlers.go:209`, `:307`,
`:365`; `doctor.go:236`); `errMissingStreamerUsername` is `errMissingUsername`; hook 4 named.
Three deliverables.

**Task 4.** Deletes the file-backed code whole (Ruling 4's yes), the address fallback and
`duplicate_principal`'s Go side, `checkRosterDirWritable`, `checkRosterDirOwner`, and `rosterFileKey`;
member-verb state table added; members-row table trimmed of the import record; R-6d23's baseline and
the fixture-username constraint added; the single-insert grep added; `templatekeys_test.go` named by
test. Mutations re-cited (`run.ts:4778-4783`).

**Task 5.** Rewritten as "The roster file retires" under Ruling 4's yes: supersession, template,
test environments, a three-row state table (the empty-value row from `rejectNamespace`). Three
deliverables. The import version moved to the appendix with one new state (the pre-rename key).

**Task 6.** Adds R-6d21's states and files (`check.go`, `check_test.go`); every test caller of
`resolveAuthConfig` named (`serve_test.go:77`, `:603`, `:696`; `rendered_test.go:497`, `:542`).

**Task 7.** `startNonceSweep` is `sweepNonces`; `auth_test.go`'s table named.

**Task 8.** Rewritten: unit grant, leaf list, two new leg checks (reverse array check, store leaf
required), the owner hint, the runbook. Dropped: `check`'s holder line, `deploy.sh`'s `check`,
halt, and `try-restart`, the presence check.

**Task 9.** Routing updated (slice 6, rung 16, `CLAUDE.md`'s gate paragraph).

**Exit gate, contested surfaces, conductor's actions, prior art, not verified.** Updated to the above.
`main.go` leaves the contested list (6d edits neither block). Prior art: Navidrome row re-reasoned,
OpenSubsonic row carries `Principal`'s reason.

**Appendix (new).** Revision 2.2's import design rebased, for Ruling 4's no.

## Citations that moved

Spec (rung 6): `:1086-1117` to `:1105-1138`; `:718-800` to `:737-819`; `:802-874` to `:821-893`;
`:290-431` to `:287-451`; `:193-215` to `:194-225`; `:877-893` to `:896-912`; `:895-917` to
`:914-936`; `:919-939` to `:938-958`; `:1121-1150` to `:1140-1169`; `:1152-1161` to `:1171-1180`;
`:1251-1265` to `:1273-1284`; `:1149` to `:1166-1169`; `:1141` to `:1160`; `:1128-1130` to
`:1147-1149`; `:208-211` to `:209-212`; `:804-805` to `:823-824`; `:795-796` to `:813-814`; `:796-797`
to `:814-815`; `:411-414` to `:414-416`; `:872-874` to `:891-893`; `:182-183` to `:183-184`;
`:251-252` to `:252-253`; `:374-375` to `:375-376`; `:759` to `:779` and `:761` to `:781` (verified
by row, not arithmetic); `:839-840` to `:858-859`; `:820-831` to `:826-855`; `:713-714` to
`:732-733`; `:463-464` to `:481-482`; `:843-846` to `:860-866`.

CH: `:143-150` to `:167-173`; `:83-84` to `:106`. S20: forward hooks `:683-768` to `:767-854`; prior
art `:770-806` to `:856-891`. R5S: `:194-196` to `:195-197`.

Code: `serve.go` roster open to `:326-332`, `privateFiles.roster` to `:375`; `store.go`'s strings
`:285-298` to `:286-307` (STATUS's `:280-284` is stale); `store.go` withholding `:331-336` to
`:333-338`; `backup.go:262-275` to `:262-277`; `doctor.go:169-173` to `:87-114` and `:226-247`;
`web.go:112-116` to `:111-117`; `open.go:248-269` to `:248-269` (unchanged); `write.go:44-63` to
`:44-64`; `conn.go:107-111` to `:107-112`; `leaf.go:66-77` to `:64-77`; `ops.go:43-48` to `:47-52`;
`ops_test.go:38-45` to `:38-46`; `helpers_test.go:22-28` to `:21-28`; `roster.go:318-320`,
`:529-535`, `:113-116` re-resolved (see the plan); `registry.go:530-544` to `:495-509`;
`registry.go:126-131` to `:125-130`; `handlers.go:210`, `:308`, `:366` to `:209`, `:307`, `:365`;
`session.go:138-233` split to `:138`, `:144-147`, `:213-219`; `rendered_test.go` (was at `314c0c8`):
`:424-460` to `:430`, `:451`; `:372-399` to `:378-400`; `:613-659` to `:618-670`; `:918-927` to
`:924-933`; `:535` to `:541-545`; `serve_test.go:66-89` to `:67-81`; `scripts/check.sh:645-660` to
`:657-737`. Web (was at `985ff3a`): members page `:312-318` to `:304-322`; `:95-97` to `:252-258`;
`run.ts:3938` to `:198-200`, `:4136`, `:4774`; `run.ts:4569-4574` to `:4778-4783`.

## Citations whose content is gone (each now a correction in the plan)

- `buildServeShellWithContext`: no longer exists; `buildServeShellWithLogger` calls
  `resolveAuthConfig` and `newServeShell` itself.
- `startNonceSweep`: renamed `sweepNonces`.
- The 5d plan's `:544`, `:1255-1260`, `:1370-1376`, `:1377-1380`, R-5d10: 5d closed; the plan now
  cites the landed unit, `deploy.sh`, and runbook.
- `errMissingStreamerUsername`, `streamer_username` everywhere: renamed.
- `web/src/routes/members/+page.svelte:95-97`'s "an older member's `streamer_username` can be blank"
  comment: gone; the live equivalent is `:252-258`.
- The Navidrome external-auth reason for `Principal` (spec `:860-866`): moot; replaced (R-6d14).
- STATUS's "5d's deploy gate checks that 6d's two wiring items have landed": gone; the present line
  is "Before any deploy exposing the strings" (R-6d1).
- `doctor.go:173`'s roster `Lookup`: the roster row now calls `List` at `:236`.

## Rulings

| Revision 2.2 | Disposition in 2.3 |
| --- | --- |
| Ruling 1 (one pass, 11.5M) | Carried as Ruling 1, re-derived to 10.5M, flag 8.4M. Recommendation yes. |
| Ruling 2 (dual-key epoch to the revoke verb) | Taken as resolution R-6d20 under the standing build-by-first-consumer rule (CLAUDE.md, 2026-09-12). `Bump` still has no production caller at `0cb33ae`. Slice 6 named as a second epoch consumer. |
| Ruling 3 (store identity to rung 16) | Carried as Ruling 3, recommendation yes. The re-plan already lists it under rung 16 conditionally; its "no" variant re-priced (12.0M, split). |
| (new) Ruling 4 (drop `member import`) | Added. It reverses half of the rung 6 spec's owner ruling 1, so it is Geoff's. Recommendation yes. The "no" content is the appendix. |

No ruling was dropped as moot. The rebase changes Ruling 1's numbers and Ruling 3's variant price.

## Could not resolve here

- R-6d21 changes 6c's ruled "never reads environ" behavior. It is taken as a resolution because the
  alternative is a regression (a bad auth value passing `check` on a fresh install). If Geoff reads
  it as a 6c ruling to re-open, it becomes a fourth question.
- The members row printing unconditionally changes `doctor`'s line count and the runbook's "prints only
  when it fails" sentence for that row. Chosen for the counts' value; the alternative (print only on
  failure, as the roster row did) is a one-line change in Task 4.
- The runbook's new install order puts the unit start before the first admin. Whether the unit start
  sits before or after the TLS front step is left to Task 8's render-and-read.
- STATUS's "To 6d" line cites `store.go:280-284`; the content is at `:286-307`. STATUS is outside this
  commit; Task 9 removes the line.
- The plan has had no `spec-plan-review` at revision 2.3. STATUS names that review as the next step
  before Geoff's gate.
