# Rung 5d plan: prose-and-facts review

Graded commit `593d7fd` (revision 2.1). Folded into revision 2.2 of
`docs/superpowers/plans/2026-09-25-rung-5d-boundary-deployment-close.md`. Line citations below are
`593d7fd`'s own. The same fold makes citation-only fixes in the approved 6c plan
(`docs/superpowers/plans/2026-09-25-rung-6c-routed-schema-settings-home.md`, revision 2.2 at
`50a1cf8`), with a one-line citation note and no revision bump.

## Findings

### Blockers

**B1 (`:1490`).** The `internal/config` contested row still says "6c may run in lane 2 concurrently",
which contradicts the revision 2.1 rule that 5d-i launches only after 6c merges.

**B2 (`:276-278`, `:1415`).** The validator hand-off names 6c, which will have merged before 5d writes
a line. 6c's validator is per-key, for store-homed settings, built from bounds data, and does not fit
a list-valued integration key; the move belongs to 6d or the architecture cleanup pass.

**B3 (`:132-162`).** Ruling 2 is not owner prose: it argues from internal pointers and jargon ("the ft
compat class", "evidence observer") and never lists every change a yes makes on the production box,
what the rollback restores, or what it leaves behind.

**B4.** The rollback misses `/etc/dubplate/env`, which step 6 allows to change and `deploy.sh:407`'s
`install_env_step` writes. Compose reads that file (`deploy.sh:508`), so a restore needs a compose
rerun when the file changed.

### Warnings

**W1.** The beets downgrade assumes 2.14.1 leaves `library.db` readable by 2.13.1, never verified.

**W2.** The timing gate ignores the previous import's `RandomizedDelaySec=300` tail and the shadow's
consumption of that import's report.

**W3.** Ruling 1 asks whether to split, but the split follows from the ten-task line (rung 6 spec
`:1149`); the owner question is the ceilings.

**W4 (`:385-386`).** The `tchar` code span contains a backtick and breaks the span.

**W5.** The 6c plan cites the 5d plan's revision 2 lines and a stale contested row ("5d holds until 6c
merges or walls it off").

**W6.** The Status paragraph (`:3-12`) carries fold history that belongs in the fold record.

### Suggestions

**S1.** Put the `config.go` pointers on one base at `593d7fd` (`:301`, the enabled-class loop).
**S2.** Fix the `dubplate-shadow.service` pointers (operator rule, no-monitoring line).
**S3.** Fix the spec pointers for R6's loopback dial and the contested-surface rows.
**S4.** Correct `:264-265`'s edit list (5d-i's Tasks 1 to 3 and 6, apart from `check.go` itself).
**S5.** Define Segment A one way.
**S6.** Add 6c's doctor rows, including the seed-at-next-start wording, to Task 7 and Task 10 step 13.
**S7.** Untangle the reservation sentence at `:253-254`.
**S8.** Unbold routine labels so the prohibitions stand out.
**S9.** Replace "underscore" at `:396`, `:897`, `:931` with "the `_` spelling" phrasing.

## Dispositions

- B1: folded; the row now reads "6c's registry edits merge before 5d-i launches (revision 2.1); 5d leaves it byte-identical".
- B2: folded in "Global constraints" as specified; Task 11's carry now reads "to 6d or the architecture cleanup pass"; Task 0 item 5 asks for the validator's shape.
- B3: folded; Ruling 2 rewritten as owner prose listing every box change, the timer window, the rollback, what it leaves, why now, and what no builds; each fact checked (box read-only 2026-09-25: binary `be27434`, beets 2.13.1, `server.env` ten rows without `DUBPLATE_IDENTITY_SOURCE`; `git diff --name-only be27434 HEAD` lists the three `config/` files and the template; `367564c` retires the featuring-credit workaround).
- B4: folded; step 7 snapshots `/etc/dubplate/env` as a third `.prev` copy and records whether step 6 showed it changing; the rollback restores all three and reruns the compose step when it did.
- W1: settled, no library copy needed; finding recorded in the rollback: identical migration list and `migrations.py`, identical fixed fields (97 item, 43 album, compared by importing both versions), and the one storage change (paths bound as `bytes`, not `memoryview`) writes the same SQLite BLOB; sources: both PyPI sdists and the 2.14.1 changelog's "Other changes"; box database at `/srv/music/state/library.db`.
- W2: folded; step 7 quotes the timer's settings from `systemd/dubplate-import.timer` and requires the import inactive, `LastTriggerUSec` at or after the last boundary, the last shadow run started after that import finished, and a stop when step 6 exceeds 7.5 minutes.
- W3: folded; Ruling 1 now asks the ceilings, and the header cites the ten-task line.
- W4: folded; double-backtick delimiters with inner spaces.
- W5: folded in the 6c plan: citations now `593d7fd:269-279`, `593d7fd:1483`, `593d7fd:10` as a revision 2.1 decision; the contested row reads "5d-i holds until 6c merges (5d plan `593d7fd:10`)"; one citation note, approval and rulings untouched.
- W6: folded; Status carries status, the review-history pointer, and the launch conditions.
- S1: folded; `config.go:301` and `:408-427` at `593d7fd` (the loop runs through `:427`, not `:425`).
- S2: folded; operator rule `:42-46`, no monitoring `:57`.
- S3: folded; spec `:369-370` and `:2542-2543`.
- S4: folded; the list names `doctor.go`, `doctor_test.go`, `check_test.go`, and `main.go`, never `check.go`.
- S5: folded; Task 0 runs alone before Segment A, and A is Tasks 1, 2, 3 in both the header and the execution table.
- S6: folded in Task 7's doctor-rows section and Task 10 step 13 (a "seeds at serve's next start" row on the box is a stop).
- S7: folded.
- S8: folded; routine labels and numbered-step leads unbolded, prohibitions kept bold.
- S9: folded at all three sites.
