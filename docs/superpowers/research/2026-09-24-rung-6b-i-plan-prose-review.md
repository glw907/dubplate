# Rung 6b-i plan prose review (revision 2.1, `e122e72`)

Reviewer: `prose-voice-reviewer` (Opus 5.5). Register: Google developer documentation, Go
dialect. `tellgrader --register docs`: 19,253 words, 943 sentences, one hard hit (P5), 21
tricolons, cadence CV 1.34. The docs-register measures profile did not resolve for this repo.

## Facts checked and found sound

Every sampled code pointer matches master (6a merged), including `main.go:54`,
`keys.go:61-71`, `lock.go:84`, `leaf.go:30-38`, `open.go:101`, `:172`, `:174`, `:261`, `:308`,
`inspect.go:119-124`, `fsprobe.go:43`, `migrate.go:62`, `:90`, `:94`, `:163`, `:340`,
`snapshot.go:73`, the env template rows, and modernc `conn.go:1208-1219` at v1.59.0.
Spec pointers match HEAD. Quotes match `os/exec`, `syscall`, `syscall/exec_linux.go`,
`os/file_unix.go`, `os/signal`, `filepath.WalkDir`, `sysexits.h`, and rclone v1.74.3's
manual. The `flock(2)` quotes were not verified (no man page on this box). The internal
arithmetic holds: 6.45M projection, 6.88M flag, 33 deliverables, segments and dependencies
consistent.

## Blockers

- **P1 (lines 84-91, 245-251, 268-269, 2222-2223).** The ceiling rests on a projected 2.1M
  close cost cited from a STATUS line that `d874634` replaced. 6a's recorded actual
  (`docs/HISTORY.md:224-230`) is about 4.0M in total, and its close side about 0.92M. Re-base
  on the recorded actual: close 0.92M plus three architecture reads at 0.1M each is 1.22M,
  the projection about 5.47M, and a ceiling sized so the 80% flag sits about one task above
  it (for example 7.3M, flag 5.84M). Delete the stale "Not verified" bullet.
- **P2 (line 2181).** The Execution model states revision 2's 8.2M and 6.56M. Point it at
  "Token ceiling" instead of restating a figure.
- **P3 (Rulings 1-2).** "The flag" means the 80% token flag, `--low-level-retries 1`, and
  `DUBPLATE_BACKUP_ENABLED`; "10" means seconds and retries in one sentence. Write "the 80%
  flag" and "the retry limit", and state units.
- **P4 (lines 137-145).** Ruling 4's "no" builds four things and leaves the helper's home to
  the conductor; an exported test helper would contradict the global constraint (338-345)
  and exit gate 11 (2074-2078). Name the home (`internal/engine/store/storetest`) and make
  it gate 11's exception on "no".
- **P5 (line 1915).** Lexicon hit "underscore": write the rule's own form,
  "a character outside `[A-Za-z0-9_]`".

## Warnings

- **P6.** Ruling 1 bundles two decisions and its "no" gives no flag figure. Ask one question
  and state that "no" holds 5.9M as a total including 0.95M of planning, with its 80% flag at
  4.72M below the projection.
- **P7.** Ruling 4's reason overclaims ("every criterion from Task 3 on"; "only export").
  Narrow it to Tasks 3-5 and Task 6's signal and exit-code rows, and "only exported
  function".
- **P8.** Ruling 3's "no" says 2, 3, and 4 are unreserved; bash builtins and rclone use 2 for
  usage errors.
- **P9 (1864-1865).** Task 6 mutation 3 predicts exit 129; a Go test sees `signal: hangup`
  with `ExitCode()` -1.
- **P10 (1769).** Task 5 mutation 6 cannot fire against the failed-run row; add a
  send-failure row.
- **P11 (1940-1944).** Criterion 21's three rows read as four.
- **P12 (77-82).** The rulings preamble opens with a 66-word sentence; split it.
- **P13 (2131-2132).** The Execution model's parenthetical mixes revision history into the
  current shape.

## Suggestions

P14 duplicate R-6bi6 labels and a mis-cited R-6bi5; P15 Task 2's independence note omits
1S; P16 Task 1B packs six outcomes into three deliverables; P17 about ten reflexive "So"
openers in the Resolutions; P18 R-6bi12's nine-idea paragraph; P19 Ruling 2's command omits
the remote argument; P20 Task 4's salvage range stops one line short.

## Verdict

The resolutions and tasks read as a careful Go engineer's design doc. Fix P1 to P4 in
"Rulings for Geoff" before Geoff reads it.
