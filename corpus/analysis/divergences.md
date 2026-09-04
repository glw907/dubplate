# Divergences from beets' own test assertions

A synthetic structural case (rung 2, pass 2b) is designed to reach a
value asserted somewhere in beets' own test suite, most often
`test/autotag/test_distance.py`. When a case's derived record
contradicts the beets test it was modeled on, that is a finding to
report here, not a value to adjust in the case: the case id, the beets
test or source line it targets, the derived value, and the author's
reading of the disagreement.

No rows yet: this file was created by pass 2b's Task 1 ahead of that
pass's own case-authoring tasks (4 through 13), the ones that could
populate it if any found a genuine divergence from a beets test
assertion the case was modeled on. Reviewed again at pass 2c's close
(Task 20): pass 2c's own case-authoring tasks (4 through 19b, a
distinct numbering from 2b's) added the real-catalog, adversarial, and
stock-config strata and filed their own findings as prose in
`corpus/README.md` and, where a suspected matcher defect was involved,
as rows in `corpus/analysis/adjudications.json` and `corpus/analysis/
oracle_bugs.json` instead—none of that work surfaced a case whose
derived record contradicts a specific beets test assertion it was
built to reach. An empty table here at pass end is a clean result, not
an unfinished one.

| case id | beets test / source line | derived value | reading |
| --- | --- | --- | --- |
