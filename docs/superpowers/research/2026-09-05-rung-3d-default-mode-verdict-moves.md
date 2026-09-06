# Rung 3d, `ModeDefault`'s verdict moves, measured

Task 8 (`internal/harness/conform/case.go`, commit `a69dee5`) runs the
whole 215-case corpus a second time under `matcher.DefaultCompat`'s
corrected pattern table, re-scoring each case's own committed candidate
set and issuing no MusicBrainz request. `CaseGradeSummary.VerdictMoves`
names every case whose `ModeDefault` verdict (recommendation plus
action) differs from its own `ModeCompat` verdict, excluding the ten
`defect_free` probe cases and the eight `Ungraded` probe cases Task 8
grades separately.

`TestWholeCorpusGradesTheCorrectedVerdictsInDefaultMode` logs the count
and the list at `case_test.go:382`. Measured at Task 8's dispatch and
reconfirmed by the fresh-context reviewer: **zero non-probe verdicts
moved.** `summary.VerdictMoves` is the empty slice.

This is the expected outcome, not a coincidence: the two compat
switches this pass corrects (`FtPattern`, `RemixParenthetical`) only
ever change the ten `adv-ft-*`/`adv-remix-*` probe cases' string
distances, which the `defect_free` block already grades exactly. No
other committed case's candidate set carries a string shape either
pattern touches, so `DefaultCompat`'s correction has nowhere else to
move a verdict today. The list stays here, empty, so a later pass that
finds a non-zero count knows the baseline it moved away from.
