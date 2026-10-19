# Adversarial review: the rung 3 pass 3d plan

Fresh-context review (2026-09-05) of
`plans/2026-09-05-rung-3d-verdicts-gate.md` revision 1, in two
independent lenses: beets verdict fidelity and the assignment solver,
verified contract by contract against pinned beets 2.13.1 and `lap`
0.5.13 in `tools/venv/.venv`, and harness, process, and plan integrity,
verified against the tree at `c7a85fd`. Folded as revision 2 the same
day; the merged 50-item "Verified sound" list below is the
do-not-re-litigate record for the pass, alongside the rung 3 spec review
and 3a's, 3b's and 3c's own review records.

The two reviews are merged here: findings naming the same defect are
deduplicated, the stronger evidence kept and both lenses noted, and the
result renumbered B1 through B4, M-a through M-i, m1 through m20. Every
finding was re-verified by the fold editor against the tree or the venv
before folding, and every numeric claim was re-measured rather than
carried. Three citation claims did not survive that check and are
rejected below with the evidence. Three items are conductor-owed rather
than folded silently, and each is applied provisionally in the plan with
an inline marker.

Both lenses agree the plan is unusually well grounded. All twenty
verdict contracts V1 through V20 resolve to the lines they name, both
lenses independently reproduced the `max_rec` census (six downgrade
cases, zero mismatches over all 215 records) and the `lap` zero-pad
extension model, and no placeholder or unresolved cross-reference exists
anywhere in the 4,687 lines. The findings concentrate in four places:
one return-value confusion in the solver's own probe evidence, the
default-mode run's collision with the request-set rule, an
under-specified expected-value table in the runner's central resolver,
and an inherited profile-merge gap the plan leans on without checking.

---

## Rejected, with the evidence

**1. `beetsplug/musicbrainz.py:645-646` is cited correctly (review 2's
m11, its renumber clause only).** Rejected. Lines 645-646 are
`for index, track_info in enumerate(track_infos, 1):` and
`track_info.index = index`; `:644` is blank and `:643` is the
`track_infos.extend` line. The plan's `:645-646` points at the renumber
it names, at V6 and at Task 7's totality-guard comment. Review 1's
Verified-sound item agrees.

**2. `client.go:129-131` is not where the `Budget` check lives (review
2's m11, its `NewClient` clause only).** Rejected as a correction, while
the underlying drift is real. `if cfg.Budget == nil {` is at
`client.go:128`, so the block runs `:128-130`. The plan's `:130-132` is
wrong and review 1's `:128-130` is what folds; review 2's replacement
would have been wrong by one in the other direction.

**3. Review 2's Verified-sound item 37, `TrackInfo at info.go:73-122`,
did not survive.** `type TrackInfo struct {` is at `info.go:73` and its
closing brace at `:105`; `:124-155` is `ItemFields`. Review 1's m10 is
right, the plan's `:73-122` is wrong, and the corrected span folds. A
Verified-sound item is a record of what was checked, so this one is
struck rather than carried.

---

## BLOCKERS

### B1. The `lap.lapjv` vectors quoted throughout the plan are the second return value (`x`, row to column); the port's contract is the third (`y`, column to row), and they differ for every n >= 3

Both lenses measured the same numbers. They disagree on what the numbers
are for, and that disagreement is the finding: review 1 read them as
correctly labeled `x` values and passed them, review 2 read them against
`solveLAP`'s stated return shape and failed them.

**Plan lines.** Header figure 4 ("the pinned solver returns
row-to-column `[1,0]`, `[2,0,1]`, `[3,0,1,2]`, `[4,0,2,1,3]`,
`[5,0,3,2,1,4]`, `[11,0,9,...,10]` and `[13,0,11,...,12]` for
n = 2,3,4,5,6,12,14"), Task 0 item 11 ("A divergence is a blocker"), and
R3's "Encouraging evidence" paragraph.

**Evidence.** `tools/venv/.venv/bin/python`,
`lap.lapjv(np.zeros((n,n)), extend_cost=True)`:

```
n=2   x = [1, 0]                       y = [1, 0]
n=3   x = [2, 0, 1]                    y = [1, 2, 0]
n=4   x = [3, 0, 1, 2]                 y = [1, 2, 3, 0]
n=5   x = [4, 0, 2, 1, 3]              y = [1, 3, 2, 4, 0]
n=6   x = [5, 0, 3, 2, 1, 4]           y = [1, 4, 3, 2, 5, 0]
n=12  x = [11,0,9,8,7,6,5,4,3,2,1,10]  y = [1,10,9,8,7,6,5,4,3,2,11,0]
n=14  x = [13,0,11,...,1,12]           y = [1,12,11,...,13,0]
```

Every vector the plan quotes is `x`, and the label "row-to-column" is
`x`'s own description, so the plan is internally consistent and points
at the wrong quantity. `lap.lapjv.__doc__` is explicit: "`x[i]`
specifies the column to which row `i` is assigned. `y[j]` specifies the
row to which column `j` is assigned." beets unpacks
`_, _, assigned_item_idxs = lap.lapjv(...)` (`$SP/beets/autotag/match.py:66`)
and its own comment at `:68-71` says the result is track-indexed, so
beets consumes `y`. The plan's contract V2 and `solveLAP`'s doc comment
("returns, for each column, the row assigned to it, or -1") both specify
`y`. The two are inverse permutations and agree only at n = 2 and on
self-inverse cases.

**Consequence, built as written.** Task 0's item 11 runs the probe the
beets way, reads `y`, sees a mismatch on six of its seven shapes, and
reports the blocker the plan told it to report. Task 2's implementer
generating `testdata/lapjv-optima.json` then has to choose between the
header's numbers and `solveLAP`'s stated return shape; picking `x` makes
the frozen table the wrong length on every rectangular row (`x` has
length `rows`, `y` has length `cols`) and mirrors the wrong permutation
on every square row at n >= 3.

**Fix.** Quote `y` wherever the plan states the pivot order, labeled
"column to row, the third return value beets consumes", and name `x` once
as the inverse so a reader probing `x` is not misled. R3's
"shift-by-one breaks at n = 5" argument survives verbatim under `y`
(`[1,0]`, `[1,2,0]`, `[1,2,3,0]`, then `[1,3,2,4,0]`), so only the
numerals change.

### B2. `ModeDefault` re-running the flow changes the request set on all ten cases it exists to grade, and the runner's own zero-miss rule fails every one of them

**Plan lines.** R8 (`ModeDefault` "runs over all 215 cases and grades
verdicts only for the ten probe cases carrying a `defect_free` block"),
R2 ("the proof that survives is `internal/harness/conform`'s
`ModeDefault` run"), and Task 7's behavior contract ("after the call,
`srv.Misses()` **and** `srv.Unused()` must both be empty ... Both fail
the case"), stated once for `GradeCases(root, mode)` and therefore in
force in both modes.

**Evidence.** All ten `defect_free` probe cases are `strong` under
compat and `medium` under the corrected table, and every one of them
logs exactly one release lookup and zero searches. Measured off the tree
at `c7a85fd`:

| case | compat rec / distance | `defect_free` rec / distance | logged searches |
| --- | --- | --- | --- |
| `adv-ft-draft-colon` | strong / 0.010185185 | medium / 0.101851852 | 0 |
| `adv-ft-left-colon` | strong / 0.011764706 | medium / 0.117647059 | 0 |
| `adv-ft-gift-space` | strong / 0.010784314 | medium / 0.107843137 | 0 |
| `adv-ft-soft-onesided` | strong / 0.030095238 | medium / 0.086666667 | 0 |
| `adv-ft-craft-titles` | strong / 0.009170752 | medium / 0.091707516 | 0 |
| `adv-ft-theft-artist` | strong / 0.009090909 | medium / 0.090909091 | 0 |
| `adv-remix-both-sides` | strong / 0.024000000 | medium / 0.080000000 | 0 |
| `adv-remix-onesided` | strong / 0.021428571 | medium / 0.071428571 | 0 |
| `adv-remix-bracket` | strong / 0.027777778 | medium / 0.092592593 | 0 |
| `adv-remix-track-titles` | strong / 0.024561404 | medium / 0.081871345 | 0 |

The ID-match early return needs `rec == strong` (`match.py:436`,
contract V18). Under `ModeDefault` the distance is 0.07 to 0.12, so the
early return is not taken, the flow proceeds to `AlbumSearchFor` plus
`SearchReleases`, and the stub answers a `/ws/2/release?...` the case
never logged with a 404 and a recorded `Miss`
(`internal/harness/stub/stub.go:119-137`). Task 7's rule then fails the
case. Task 5's error policy compounds it: `SearchReleases` returns
`ErrNotFound` on a 404 (`client.go:381-382`), which is not "on a
lookup", so it is wrapped and returned and the case never reaches a
`defect_free` comparison at all.

The wall is not confined to the ten. Every case whose corrected distance
crosses 0.04 loses its early return, and the plan's own upper bound for
that population is 133 cases, so R9's `VerdictMoves` measurement fails
the same way.

**Fix (conductor-owed; see the rulings list).** Two independent halves.
First, the request-set rule is `ModeCompat`-only, stated at Task 7 and
Task 8 with its reason: a corrected verdict legally changes the flow, so
the committed request log is an expectation of the compat run alone.
Second, `ModeDefault` scores each case's own committed candidate set
under the corrected table rather than re-running `tag.Album`. That
reproduces the ten `defect_free` figures exactly, because the medium rung
of the ladder is decided by the minimum distance alone
(`match.py:288-290`), so a candidate set the search would have widened
cannot move a `medium` verdict off `medium`.

### B3. Task 1's `inclArtist=false` fixture asserts the answer the wrong implementation gives

**Plan lines.** `TestAssignItemsUsesInclArtistFalseOnAVariousArtistsRelease`,
whose comment reads "With inclArtist false the matrix is tied and the
solver's own choice stands" and whose assertion is
`want := []Pair{{Item: 1, Track: 0}, {Item: 0, Track: 1}}`, the swap.
This fixture is the plan's sole proof of the line (declared gap 10, Task
1's acceptance criterion, Task 1 mutation 3).

**Evidence.** The matrix is not tied. `track_index` fires whenever both
`track_info.index` and `item.track` are set
(`$SP/beets/autotag/distance.py:411-413`, ported at
`internal/engine/matcher/track_index.go:17-23`), and the fixture sets
`Track: 1, 2` against `Index: 1, 2`. Driving pinned beets on exactly the
fixture's four objects:

```
incl_artist False [[0.0, 0.16666666666666666], [0.16666666666666666, 0.0]]
                  identity 0.0     swap 0.3333333333333333
incl_artist True  [[0.2222222222222222, 0.125], [0.125, 0.2222222222222222]]
                  identity 0.4444444444444444    swap 0.25
```

Under `inclArtist=false` the identity is the unique optimum;
under `inclArtist=true` the swap is the unique optimum (`track_artist`
weight 2.0 beats `track_index` 1.0,
`internal/engine/matcher/tunables.go:272-273`). The plan's `want` is
inverted. An implementer building to this test is driven either to pass
`true` into the cost matrix, which is the exact defect the fixture
exists to catch and one no corpus case would catch, or to abandon the
fixture and lose the proof.

The fixture is otherwise good. It discriminates cleanly in both
directions, and the guard loop still holds under the corrected
expectation, since `DistanceForTrack(item0, track0, true) = 0.2222`.

**Fix.** `want := []Pair{{Item: 0, Track: 0}, {Item: 1, Track: 1}}`, and
replace the comment's tie claim with the real mechanism: the index
component makes the identity the unique optimum under `inclArtist=false`
and including the track artist flips the unique optimum to the swap, so
the fixture discriminates the two objectives without depending on
`lapjv`'s tie rule at all.

### B4. `AssignmentFloor = 214` reds the gate on ruling F's escalation path, which is the one path the pass's whole process design exists to keep green

Both lenses found this. Review 2 called it a blocker, review 1 a major;
blocker is right, because it makes `bash scripts/check.sh` red on a
pre-authorized outcome and the pass may never commit red.

**Plan lines.** Task 7's const block (`AssignmentFloor = 214`) and
`TestWholeCorpusGradesCleanlyInCompatMode`'s unconditional
`if f.got < f.floor { t.Errorf(...) }` over the row
`{"assignments", summary.Assignments, AssignmentFloor}`.

**Evidence.** 214 cases carry at least one candidate
(`jq -s '[.[]|select((.expected.candidates|length)>0)]|length'` over
`corpus/cases/*/case.json` returns 214), and both pinned cases are
inside that population (`should-ask` 5 candidates, `should-reject` 6).
R3 says the two grade "as pending, never as passing", tier 1 "skips the
assignment comparison for those two case ids alone", and "when the
finding exists the gate is green". Skipping them yields
`Assignments = 212 < 214`, so the floor fires. The plan knows the
arithmetic and does not implement it: the test's own comment says "214
measurable rows less whatever ruling F still owes", the graded-field
table says "214, less `TiesAwaitingRuling` and less `ties` excluded",
and Task 8's verdict assertion already subtracts in exactly that shape.

**Fix.** Compare against
`AssignmentFloor - len(summary.TiesAwaitingRuling) - summary.TiesExcluded`,
name the arithmetic in the error message so a shrink for any other
reason still fires, and say so in the constant's doc comment.

---

## MAJORS

### M-a. The `policyFor` fixture table pins `GradeAlbumDistance: false` on every row and omits the spec's `ties` unique row

Both lenses, independently, with the same consequence.

**Plan lines.** `TestPolicyForHonorsEveryDispositionRow`, fourteen rows
compared with `if got != tt.want`, which is struct equality, so every
unset field is an assertion.

**Evidence.** `casePolicy` carries `GradeAlbumDistance bool`, and across
all fourteen rows the field is set exactly once, in
`{"ties excluded", casePolicy{GradeAssignment: false, GradeAlbumDistance: false, ...}}`.
Every other row leaves it at the zero value, so the table asserts that
the album distance is never graded for any disposition. An implementer
who writes `policyFor` to satisfy the table ships a resolver that turns
album-distance grading off corpus-wide, `summary.Distances` reads 0
against `DistanceFloor = 294`, and the obvious repair on offer is to
weaken the floor. That is risk 2's silent-pass shape exactly (spec
`:413`). Six further rows leave `GradeAssignment` false while their
prose says the case "grades like any other" (spec `:294`, `:297`,
`:298`).

Separately, the comment claims "Thirteen rows, one per row of the rung 3
spec's disposition table" while the table has fourteen entries. The
spec's disposition table has twelve rows (`:285-299`), the plan's own
per-case grading table lists thirteen (the twelve plus the
adjudications-pending path from `:304-314`), and the fixture table is
missing `ties` unique, whose fixture the acceptance criterion demands.

**Fix.** Fill every field of every `want` literal explicitly, add the
`ties unique` row, and correct the count to fifteen: twelve spec
disposition rows, the adjudications-pending path, and the two precedence
shapes.

### M-b. `mergeProfile` does not apply the three `import.*` keys the committed profiles set, so `gradeTunables`' stated contract is false as landed

Both lenses. Review 1 named the one key whose value actually differs;
review 2 named the honesty argument R12 builds on it.

**Plan lines.** R12 ("`mergeProfile` models the three committed profile
**files**, and `checkProfileKeyDrift`'s allowlist is honest only while
that stays true") and Task 7's `gradeTunables` doc comment, which
repeats it.

**Evidence.** `corpus/profiles/estate.yaml:17-19` and
`corpus/profiles/locale.yaml:17-19` both set `quiet_fallback: skip`,
`none_rec_action: skip` and `timid: no`.
`internal/harness/conform/translate.go:194-203` returns `base` unchanged
for `"stock", "estate"` and sets only `Languages` and
`IgnoredAliasTypes` for `"locale"`, while `profileKeyAllowlist`
(`:213-224`) maps all three keys to Tunables fields. One of the three
genuinely differs from beets' shipped default:
`$SP/beets/config_default.yaml:34` is `none_rec_action: ask`, which is
what `matcher.DefaultTunables()` carries (`tunables.go:298`), against
`skip` in both profiles. So `gradeTunables("estate").NoneRecAction` is
`"ask"` where the profile says `"skip"`.

It is invisible today only because the `Quiet: true` overlay makes
`NoneRecAction` unreachable (contract V16,
`$SP/beets/ui/commands/import_/session.py:337-345`), which is the same
overlay R12 introduces. The plan is building an honesty argument on a
premise it did not verify, and this is exactly the drift R13 says it is
closing.

**Fix.** Extend `mergeProfile`'s `estate` and `locale` arms to set
`QuietFallback`, `NoneRecAction` and `Timid` from the files, three
assignments, behavior-preserving under the overlay, with a test
asserting each against the profile file. That makes `gradeTunables`' doc
comment true rather than aspirational.

### M-c. Task 6's five tie table rows will not compile, and the plan's six-site list misses a seventh

**Plan lines.** R10's six-site table; Task 6 Step 1's five rows of the
shape `{name: ..., entry: TieEntry{...}, wantHas: ...}`.

**Evidence, two parts.**

1. **The seventh site.** `internal/harness/analysis/ties_test.go:12-14`
   builds `dispositioned` with `Disposition = TiePinned` and a `Reason`
   and no `Adjudicator`, and `:47-51` uses it in a row whose `wantHas`
   is empty, meaning accepted. The new rule ("`Adjudicator` is required
   whenever `Disposition` is set") makes that row error, so an existing
   passing row flips to failing. The plan claims site completeness and
   does not name it. The neighbouring `noReason` fixture (`:17-18`) is
   safe, because `checkTieEntry` joins its errors
   (`ties.go:114-142`) and `wantHas` is a substring test.
2. **The table shape.** `ties_test.go:38-41` declares the table as
   `struct { name string; caseIDs []string; ledger TieLedger; wantHas string }`
   and the loop calls `CheckTieLedger(tt.caseIDs, &tt.ledger)`. There is
   no `entry` field. All five proposed rows use `entry: TieEntry{...}`
   and none supplies `caseIDs` or a `ledger`.

**Fix.** Add the seventh site (the `dispositioned` fixture gains an
`Adjudicator`), and rewrite the five rows in the landed
`caseIDs`/`ledger` shape, each a one-element `TieLedger` with a matching
one-element `caseIDs`.

### M-d. One shared `ParseAction` for `QuietFallback` and `NoneRecAction` accepts values beets refuses

**Plan lines.** Task 4's behavior contract ("`Validate` gains its
`MaxRec`, `QuietFallback` and `NoneRecAction` halves, resolving each
through the same parsers `Recommend` and `ActionFor` use, so one name
table serves both times") and `ErrInvalidTunable`'s doc comment, whose
stated purpose is to replace the validation confuse performs.

**Evidence.** `$SP/beets/ui/commands/import_/session.py:340-342`
resolves `quiet_fallback` through `as_choice({"skip": SKIP, "asis": ASIS})`,
two legal values. `:346-352` resolves `none_rec_action` through
`as_choice({"skip": SKIP, "asis": ASIS, "ask": None})`, three. The
plan's `ParseAction` is specified over four (`ActionApply`, `ActionSkip`,
`ActionAsIs`, `ActionAsk`), so `Validate` would bless
`quiet_fallback: apply`, `quiet_fallback: ask` and
`none_rec_action: apply`, each of which confuse raises on, and
`ActionFor` would then return a value beets can never produce on a
configuration `Validate` approved.

**Fix.** Give `Validate` two distinct legal sets, cited to
`session.py:341` and `session.py:348-350`, and have `ActionFor`
re-resolve against the same two. `ParseAction` may stay as the shared
lexer; the membership check has to be per key.

### M-e. Task 5's mutation 1 does not fire its own named test, and its corpus count is 39, not 48

**Plan line.** "Relax the early return's `rec == strong` guard to
`rec >= medium` → `TestAlbumTakesTheIDMatchEarlyReturnAndIssuesNoSearch`
fails, and Task 7's tier-1 sweep fails on the 48 search-fallback cases
as unused log entries."

**Evidence, two parts.** The named test's fixture is "a unanimous
mb_albumid whose lookup scores **strong**" and the test asserts
`p.Recommendation == matcher.RecStrong`. Relaxing `== strong` to
`>= medium` leaves a strong match on the same branch: one request,
recommendation strong, test passes. The named detector cannot fire, and
this is the third instance in rung 3 of the failure class R11 itself
cites twice.

The count was re-measured over the 48 search-logging cases, joining each
case's first release lookup to the candidate carrying that album id.
The ID candidate's recorded distance is at or below 0.25 on 39 of them
(six of those below 0.04, reaching `medium` only through the `max_rec`
downgrade: `assign-extra-item`,
`complete-combined-short-scrambled`, `complete-max-rec-boundary`,
`complete-short-one`, `real-noise-pink-floyd-de1994`,
`rec-max-rec-missing-tracks-downgrade`), above 0.25 on 7, and absent
from the record on 2 (`admit-zero-track-dropped`,
`real-nonlatin-hamasaki-ayumi`). Under `rec >= medium` exactly 39 cases
newly take the early return, and the remaining 9 still search.

**Fix.** Give the mutation a fixture that can see it, a second
`testdata` case whose ID lookup scores medium and whose expectation is
one lookup plus one search, and correct the count to 39. Task 5's
mutation 2 (167) and Task 7's mutation 3 (48) are both correct.

### M-f. Task 4's mutation 1 claims a corpus detector that does not exist

**Plan line.** "Change the strong comparison from `<` to `<=` →
`TestRecommendWalksTheLadder`'s row fails, and Task 7's tier-1 sweep
fails on `rec-strong-medium-boundary`, the corpus's single
`proximity`-flagged row."

**Evidence.** `corpus/cases/rec-strong-medium-boundary/case.json`
records `distance = 0.040000199999999986`, and
`corpus/analysis/proximity.json`'s own reason says "2e-7 above
strong_rec_thresh". Both `0.0400002 < 0.04` and `0.0400002 <= 0.04` are
false, so the case resolves to medium under either comparison. A sweep
over all 215 cases finds zero candidates whose distance equals 0.04 or
0.25 exactly. The mutation is caught by the hand-authored ladder row
alone, which is what the plan's own Task 7 mutation 4 already says.

**Fix.** Strike the corpus half of mutation 1, state that the
strict-versus-inclusive asymmetry is graded by the hand-authored ladder
rows alone, and add it to Task 9's declared-gap list.

### M-g. Task 4's mutation 6 also has no corpus detector

**Plan line.** "Move the `timid` check above the `quiet` check in
`ActionFor` → the 'quiet outranks timid' row fails, and Task 7's tier-1
sweep fails on all 215 cases, because the run overlays `Quiet: true`."

**Evidence.** `_summary_judgment`'s timid arm returns `None` only when
`config["import"]["timid"]` is truthy (`session.py:343-344`).
`corpus/profiles/estate.yaml:19` and `locale.yaml:19` both set
`timid: no`; `stock.yaml` sets nothing and beets' shipped default is
`timid: no` (`$SP/beets/config_default.yaml:27`);
`matcher.DefaultTunables()` sets `Timid: false` (`tunables.go:293`).
With `Timid` false on all 215 cases, reordering the two checks changes
nothing anywhere in the corpus.

**Fix.** Strike the "all 215 cases" clause and say the unit row is the
sole detector, consistent with declared gap 7.

### M-h. Task 5's 503 row spends 31 seconds of real wall time on every gate run

**Plan lines.** `TestAlbumDropsA404LookupAndPropagatesEverythingElse`,
row `{"503 propagates", http.StatusServiceUnavailable, musicbrainz.ErrUnavailable}`
against `statusServer(t, 503)`.

**Evidence.** `maxAttempts = 7`
(`internal/engine/musicbrainz/client.go:42`), 503 is retryable
(`:50-56`), and each retry waits `retryWait(Retry-After, attempt)`
(`:403-411`) which falls back to `backoff(attempt)` =
`500ms * 2^(attempt-1)` capped at 120 s (`:418-423`). Six waits:
0 + 1 + 2 + 4 + 8 + 16 = 31 s. The wait runs through `c.sleep`, an
unexported field (`:118`); `ClientConfig` (`:93-99`) exposes `BaseURL`,
`UserAgent`, `Budget`, `Timeout` and `HTTP` and no sleep hook, so a test
in `internal/engine/tag` cannot inject a fake. `go test ./...` measured
15.081 s at 3c's close, so this single row triples it and passes the
plan's own 45 s trigger before tier 1 or tier 2 adds anything.

**Fix.** Have `statusServer` send `Retry-After: 0` alongside the status.
`retryWait` parses a delta-seconds header ahead of the backoff and
`strconv.Atoi("0")` yields a zero wait, so the seven attempts run back
to back. Assert the elapsed time stays under a second so the guard
cannot rot.

### M-i. `casePolicy.CompatOn bool` cannot express two independent per-class switches

**Plan lines.** `CompatOn bool` in `casePolicy`, the policy table row
`oracle_bugs pending-adjudication` ("compat switch on for that class"),
and the fixture row `{"oracle bug pending", casePolicy{CompatOn: true, GradeVerdict: true}}`.

**Evidence.** `matcher.Compat` carries two independent fields,
`FtPattern` and `RemixParenthetical` (R1), and the spec's precedence
rule is explicitly per hypothesis: "`oracle_bugs.disposition` alone sets
whether the compat switch for that defect class defaults on, because a
switch is per defect class and cannot be half on" (spec `:302`). A
single boolean per case collapses the two classes. A case that probes a
pending `bug-ft-pattern` while `bug-remix-parenthetical` stays
`oracle-wrong` needs `{true, false}`, which `CompatOn: true` renders as
`{true, true}`. It is dormant today only because
`corpus/analysis/oracle_bugs.json` carries zero pending rows and both
string hypotheses agree, so the hand-authored fixture would bake the
collapse in as the reference behavior for rung 15's re-run.

**Fix.** Make the field `Compat matcher.Compat`, resolved from the
ledger per class, and give the pending fixture two rows: one where both
classes are pending and one where only one is.

---

## MINORS

**m1.** R10's site table says `ties_test.go` gains "three new table
rows"; Task 6 Step 1 lists five and the acceptance criterion says "All
five new tie table rows". Make R10 say five. (`ties_test.go`'s existing
table is fourteen rows, `grep -c "name:"` returns 14.)

**m2.** R19 says the 17 unledgered tied candidate assignments span "six
cases beyond the two pinned ones", while Task 9 carry 3 names five
non-pinned cases plus the non-best candidates of `should-ask` and
`should-reject`. Five and six cannot both be right. The 17-tied,
five-vector-changing, two-sum-changing figures are also the one
load-bearing corpus number neither lens could reproduce without
implementing the solver, since they require re-solving every candidate's
cost matrix. State the case list once, let the count follow it, and mark
the figures as carried from the inventory and re-measured at Task 7.

**m3.** Task 3 says `Validate` compiles both `Preferred` lists "through
`CompilePatternOption`", but the media component compiles
`` `(\d+x)?(`+pattern+`)` `` (`internal/engine/matcher/album_media.go:39`)
while the country component compiles the pattern bare
(`album_country.go:34`). A bare-valid pattern is always valid wrapped,
so `Validate` passing still implies the panics are unreachable, and a
pattern such as `a)(b` compiles wrapped and fails bare, so a bare check
rejects a media configuration the component accepts. `Validate` must
compile each the way its own component does.

**m4.** R6 rules that "None" maps to the Go zero value, `""` for a
string, `0` for an int, `false` for a bool. For `AlbumInfo.VA` that is a
behavioral divergence rather than a modelling choice: beets'
`album_info.va` is declared `va: bool = False`
(`$SP/beets/autotag/hooks.py:351`, assigned at `:381`) and is never
`None`, so `required: [va]` filters nothing in beets and drops every
non-VA release in Go. The three int fields (`Year`, `OriginalYear`,
`Mediums`) carry a weaker form of the same. Name the bool case on the
filter's doc comment beside the string case.

**m5.** `CaseFloor = 215` duplicates `RequestCaseFloor = 215`, already
exported from the same package at
`internal/harness/conform/request.go:28`. One constant, or one that
derives from the other.

**m6.** `ModeCompat`'s compat value is never written down. Task 8 states
`ModeDefault` "installs `matcher.DefaultCompat()`"; Task 7 says only
"install `mode`'s compat selection", and `DefaultCompat()` returns
`{false, false}` while `ModeCompat` needs `{true, true}`. Name it, and
give it a constructor (`matcher.FullCompat()`) with its own doc comment
so "reproduce every ledgered defect" is a named engine surface rather
than a bare struct literal in the harness.

**m7.** `Candidates.Add`'s overwrite semantics are unstated for a
repeated identity. beets writes `results[info.identifier] = AlbumMatch(...)`
unconditionally at `$SP/beets/autotag/match.py:378-380`, so a second
candidate with an empty `album_id` (gate 2 does not apply, contract V10)
replaces the value while keeping the earlier insertion position. A Go
accumulator that appends to `order` on every `Add` double-counts, and
the plan's own `noID` fixture is one candidate away from it.

**m8.** Three whole-corpus sweeps land in one package: Task 7's
`TestWholeCorpusGradesCleanlyInCompatMode`, Task 8's
`TestWholeCorpusGradesVerdictsInCompatMode` and `...InDefaultMode`. Two
of the three are the same `ModeCompat` run. At the plan's own 6 to 20 s
estimate the duplicate alone can push `go test ./...` past the 45 s
trigger the plan sets. Cache one `ModeCompat` summary behind a
`sync.OnceValues` or fold Task 8's compat assertions into Task 7's test.

**m9.** Each `defect_free` block records `pattern_table`
(`SDPatternsWordBoundary` on the six `adv-ft-*` rows,
`SDPatternsNoBracketDiscount` on the four `adv-remix-*`), and
`analysis.LegalPatternTables`
(`internal/harness/analysis/adjudication.go:96`) pins those two names,
already enforced at `:254-255`. `ModeDefault` grades those figures under
a third table and the plan never reconciles the two. Say in R2 that the
new table is deliberately not a legal `pattern_table` value and why, so
the next reader does not add it to `LegalPatternTables` and break the
existing check.

**m10.** Citation drift, harmless but worth correcting before an
implementer chases it. `matcher.TrackInfo (info.go:73-122)`: the type is
`info.go:73-105`. `musicbrainz.NewClient requires one, client.go:130-132`:
the `Budget` check is `client.go:128-130`. "`Tunables.Quiet` defaults to
`false` (`tunables.go:186`)": `:186` is the field declaration and the
default literal is `:296`. `checkSkippedTieEntry (ties.go:148-158)` runs
to `:159`. "beets raises AttributeError here (`match.py:356`)": the
`getattr` is `match.py:357`. R9's
`caseDelta == 0` citation `distance.py:110-111`: the check and its
`continue` are `:108-109`. `CheckAdjudications`' five non-empty rules
`adjudication.go:158-172`: they run `:159-173`.

**m11.** `AssignItems`' contract does not say what happens on zero items
or zero tracks. `solveLAP`'s callers index `costs[0]`, and beets never
reaches it because gate 1 drops a zero-track candidate first
(`match.py:345-347`). One sentence in the doc comment, naming gate 1 as
the reason it is unreachable, keeps a later caller out of a panic.

**m12.** `get` is called with the wrong argument type and `loadFixture`
does not exist. Task 7 Step 1 writes `get(t, srv, c.Requests[0].PathQuery)`
and `c, dir := loadFixture(t, "unused-fixture")`. The landed helper is
`func get(t *testing.T, addr, path string) *http.Response`
(`internal/harness/stub/stub_test.go:41`), and there is no `loadFixture`
in that package; `stub_test.go:155` uses `cassette.LoadCase` inline.

**m13.** `Unused()` needs a field `stub.New` does not create. `Server`
(`internal/harness/stub/stub.go:63-72`) keeps `responses map[key][]byte`
and `extra map[key]extraEntry` and never retains `c.Requests`;
`extraEntry` (`:77-80`) exists precisely because `Consumed()` needed the
entry rather than a reconstructed shape. Name the added
`requests []cassette.Request` field in Step 3 so it is not discovered.

**m14.** Task 2's fixture schema omits the field its own test reads. The
behavior contract says each file is
`{"case_id", "rows", "cols", "costs", "want"}`;
`TestSolveLAPMirrorsLapjvOnThePinnedTiedCases` reads `fx.OptimalCost`.
Add `"optimal_cost"` to the schema; the two values are already named.

**m15.** The contested-surface audit counts one row that does not exist.
The plan says "Six already carry a row in the parallel spec's
contested-surface table" and lists `docs/` ledgers as written.
`docs/superpowers/specs/2026-09-03-parallel-build-approach.md:156-175`
has no `docs/` row; `:175`'s `corpus/README.md` row merely refers to
"the same standing rule the `docs/` ledgers row states". Five rows
exist, `docs/` ownership rests on CLAUDE.md's standing rule, and Task 0
item 4's expected answer ("six of nine present") is wrong.

**m16.** `internal/engine/tag` collides in reading with the build spec's
planned `internal/engine/tagio`, rung 8's FLAC read and write
(`docs/superpowers/specs/2026-09-02-greenfield-server-build.md:312`).
R18's rejected alternatives consider `musicbrainz` and a
`matcher.Source` interface and never the name, and Task 9's drafted
corrections name only the rung 3 spec. Either pick a name that does not
read as `tagio`'s sibling or record the pairing deliberately, and add
the build spec's tree listing to Task 9's drafted corrections.

**m17.** `ProbeCaseFloor - 2` is an unexplained magic constant. Task 8
asserts `len(summary.Ungraded) + summary.Verdicts >= ProbeCaseFloor-2`.
The `2` encodes the two probe cases of an `oracle-wrong` hypothesis
whose own adjudication row is `oracle-correct`
(`real-nonlatin-alla-pugacheva-vol1`, `real-nonlatin-x-japan-art-of-life`,
both under `bug-nonlatin-alias-selection`, verified case by case). The
contract paragraph says it correctly, the assertion does not, and the
constant goes stale the moment Geoff rules another row.

**m18.** R2's measured string counts do not reproduce, though its
conclusion does. The plan says "0 of 234 compared strings" across the
six `adv-ft-*` cases and "0 of 156" across the four `adv-remix-*`. One
lens counting item titles, albums and artists plus every `title`,
`name`, `artist` and `disambiguation` response-body string got 216 and
144; the other, widening to `albumartist` and `sort-name`, got 258 and
172. All three agree on zero matches either way, so
`SDPatternsGuardedNoBracketDiscount` is as safe as claimed and only the
denominators are unstable. Report the zero result with the field set
named rather than a bare denominator.

**m19.** The budget section omits the window spend cap the same STATUS
sentence records. The plan cites `docs/STATUS.md:23-28` twice for the
80%-of-ceiling routing grant and never mentions "window spend cap 15M"
from that same bullet. With an expected total near 6.8M plus prep the
cap is not in danger, and a plan that names one half of a grant should
name both.

**m20.** V14 says the album-level-nonzero model "reproduces all 214
recorded recommendations" where header figure 1 and both lenses' replays
say 215. Every case carries a recommendation, the no-candidate case
included, and the model gets that one right too.

---

## Verified sound

The do-not-re-litigate record, merged from both lenses. Each item was
checked against the tree at `c7a85fd` or the pinned venv, by the method
named.

**Corpus census, every number re-measured and exact.**

1. 215 cases; profile split estate 178, stock 36, locale 1.
2. 294 candidates across 214 candidate-bearing cases, 1 case with none;
   histogram 193 at one, 4 at two, 9 at five, 8 at six.
3. 7 cases with a non-empty `extra_items` (max 33), 14 with a non-empty
   `extra_tracks` (max 4), 0 with both, so V4's one-sidedness holds on
   the committed corpus.
4. 23 non-identity assignments; 2 cases carrying an exact
   candidate-distance tie; 3 cases carrying two or more extra items, so
   Task 7's totality guard has a population.
5. Verdicts 167 strong/apply, 39 medium/skip, 4 low/skip, 5 none/skip.
6. 272 of 294 candidates carry some key at exactly 0.0; 0 of 294 carry
   `missing_tracks` or `unmatched_tracks` at 0.0 (declared gap 5
   confirmed, and Task 7 mutation 1's count of 272).
7. `ties.json` 215 rows, 213 undispositioned, 2 pinned (`should-ask`,
   `should-reject`), 0 excluded, 1 skipped (`admit-zero-track-dropped`);
   `grep -c adjudicator` returns 0.
8. `proximity.json` 2 rows, 1 flagged (`rec-strong-medium-boundary`, on
   `strong_rec_thresh`), 1 skipped.
9. `adjudications.json` 21 rows, 18 `oracle-wrong`, 3 `oracle-correct`,
   0 pending, 10 carrying a `defect_free` block. `oracle_bugs.json` 6
   hypotheses, 4 `oracle-wrong`, 2 `oracle-correct`, 0 pending.
10. `corpus/manifest.json` carries `schema_version` `"4"`, `lap`
    `"0.5.13"`, `numpy` `"2.5.2"`, and `profile_hashes` over `locale`
    alone (R20 confirmed).
11. Largest single response body 9,107,383 bytes (8,894 KB); 77.57 MB
    across all response files, which is Task 7's memory argument.

**The header's settled figures.**

12. **The `max_rec` downgrade fires on exactly six cases.** Both lenses
    replayed `_recommendation`'s ladder over all 215 records
    independently and reproduced the branch census exactly (strong 173,
    medium 33, single-candidate low 3, gap low 1, no-conclusion 4,
    no-candidates 1), fired the downgrade on precisely
    `assign-extra-item`, `complete-combined-short-scrambled`,
    `complete-max-rec-boundary`, `complete-short-one`,
    `real-noise-pink-floyd-de1994` and
    `rec-max-rec-missing-tracks-downgrade`, and reproduced all 215
    recorded recommendations with zero mismatches. Section (a) wins over
    (c)'s 21.
13. **The fourth pattern table is safe.** Zero bracketed strings across
    the six `adv-ft-*` cases and zero featuring triggers across the four
    `adv-remix-*`, on both lenses' field sets (see m18 for the
    denominators). Since a non-matching `SD_PATTERNS` row is a no-op
    (`distance.py:102-109`), `SDPatternsGuardedNoBracketDiscount`
    reproduces both sets of `defect_free` figures.
14. **The proposed fourth table is byte-correct.** `BeetsSDPatterns`
    (`tunables.go:348-355`), `SDPatternsWordBoundary` (`:365-372`,
    featuring row at index 2), `SDPatternsNoBracketDiscount`
    (`:388-393`, four rows). The plan's literal is exactly
    `SDPatternsNoBracketDiscount` with its featuring row replaced by
    `SDPatternsWordBoundary[2]`, so
    `TestSDPatternsGuardedNoBracketDiscountIsBothFixesAndNothingElse`'s
    construction is right.
15. **The `lap` zero-pad extension model reproduces.**
    `n = max(rows, cols)` padded with zero agreed with
    `extend_cost=True` on 1316 of 1316 random rectangular trials on one
    lens and 256 of 256 on the other; the classic `(rows+cols)` square
    with `max+1` fill agreed on 54%. `lap` returns `-1` for unmatched
    entries, and the result's length is the column count, which is what
    `solveLAP`'s contract requires. (The pivot-order vectors themselves
    are B1.)

**Beets contracts V1 through V20, verified line by line in
`tools/venv/.venv/lib/python3.13/site-packages`.**

16. V1: `costs = [[float(track_distance(i, t)) ...]]` at `match.py:64`;
    `track_distance`'s `incl_artist` defaults False at
    `distance.py:383-385`; `__float__` at `:218-219`; the `distance`
    property with its zero-max special case at `:153-160`. Each cell is
    normalized against its own denominator.
17. V2: `lap.lapjv(..., extend_cost=True)` at `match.py:66` and the
    track-indexed comment at `:68-71`.
18. V3: the mapping comprehension at `match.py:72-76`; `dist.tracks`
    filled pair by pair at `distance.py:527-530`;
    `AlbumMatch.item_info_pairs` is `list(self.mapping.items())`
    (`match.py:158-160`), so the order survives into
    `derive_oracle.py:183`; `matcher.addAlbumTracks` iterates `pairs`
    verbatim (`album_tracks.go:25-31`). The track-order requirement is
    real and correctly stated.
19. V4: `lapjv` matches exactly `min(rows, cols)`, probed on
    rectangular shapes both orientations.
20. V6: both extras and both sort keys at `match.py:77-80`; the `1..n`
    renumber that makes `t.index` unique at
    `beetsplug/musicbrainz.py:645-646`, so `extra_tracks`' key is
    provably total.
21. V7: all six `_add_candidate` steps at `match.py:345-347`,
    `:350-352`, `:355-359`, `:362-367`, `:370-375`, `:378-380`, in that
    order, with a `return` on every drop.
22. V8: `Distance.__iter__` to `items()` drops zero-weighted penalties
    (`distance.py:240-241`, `:190-205`); `derive_oracle.py:176` emits
    the `_penalties` view; the three Go views are `AddedKeys`
    (`penalty.go:124`), `PerKey` (`:133`) and `NonzeroKeys` (`:166`).
    Hazard and sites correct.
23. V9: `getattr(info, req_tag) is None` at `match.py:357`;
    `matcher.AlbumInfo` carries fourteen fields at `info.go:10-67`,
    thirteen plain-typed scalars plus `Tracks`, so "no nil to express
    `is None`" holds.
24. V10: `Info.identifier` returns `(data_source, id)` at
    `hooks.py:166-168`; `cassette.Candidate`'s `Source` and `AlbumID` at
    `cassette.go:103-109`.
25. V11: `_sort_candidates` is stable `sorted` at `match.py:326-328`.
    Exactly two cases carry an exact candidate-distance tie:
    `should-ask` (3 of 5 candidates at `0.23868312757201643`) and
    `should-reject` (5 of 6 at `0.3891203703703704`).
26. V12: the ladder at `match.py:279-302` with strict `<` at `:285` and
    inclusive `<=` at `:288`. The ladder table's rows check out
    arithmetically: `0.55 - 0.30 = 0.25000000000000006 >= 0.25` (low)
    and `0.54 - 0.30 = 0.24000000000000005 < 0.25` (none).
27. V13: the no-conclusion branch returns immediately at
    `match.py:300-302`; four committed cases take it, matching the
    plan's branch census.
28. V14: `keys = set(min_dist.keys())` at `:306` plus the track union at
    `:307-309`; `Distance.keys()` is built from `items()` at
    `distance.py:246-247`. (The 214-versus-215 wording is m20.)
29. V15: `Recommendation(IntEnum)` at `match.py:39-47`, `as_choice` at
    `:313-320`, `min(rec, max_rec)` at `:321`; `Tunables.MaxRec` is
    `map[string]string` at `tunables.go:61-67`.
30. V16: `_summary_judgment` at `session.py:328-360`, quiet first at
    `:337`, strong to APPLY at `:338-339`, the fallback at `:340-342`,
    timid at `:343-344`, none at `:345-352`, else `None` at `:353-354`;
    `Action` at `importer/actions.py:8-19`. All nine rows of
    `TestActionForCoversEveryBranch` are individually correct against
    this source (the shared-parser question is M-d, not a row error).
31. V17: `config["import"]["quiet"] = True` at
    `internal/harness/oracle/derive_oracle.py:204` and the `"ask"`
    mapping at `:206`. The overlay is genuinely required, and the
    corpus's dispositions encode exactly `apply` 167 and `skip` 48.
32. V18: every line of the `tag_album` transcription, including
    `likelies` at `:407-409`, the insertion-ordered dict at `:414`, the
    `search_ids` branch with no early return at `:417-420`,
    `match_by_id` at `:424-428`, the unsorted `_recommendation` at
    `:430`, the early return at `:432-441`, `va_likely` at `:451-455`,
    the plugin fan-out at `:459-463`, and the sorted close at
    `:466-468`.
33. V19: `match_by_id`'s two early exits and the one-element
    `albums_for_ids([album_id])` at `match.py:252-266`;
    `MostCommonTags` returns both halves at
    `internal/engine/matcher/likelies.go:47`; `vaLikely` is unexported
    at `internal/engine/musicbrainz/query.go:108`.
34. V20: the `tracks`, `missing_tracks`, `unmatched_tracks` loop order
    at `distance.py:527-538`.
35. `album_for_id` returning `None` on a 404 at
    `beetsplug/musicbrainz.py:777-781`, which Task 5's 404 row mirrors:
    `AlbumForID` returns `ErrNotFound` unwrapped on a first-lookup 404
    (`client.go:283`, `:381-382`).

**Harness and engine surfaces named by the plan, all present at the
signature and line named.**

36. `cassette.Expected`'s eleven fields at `cassette.go:76-88`;
    `Assignment map[int]int` at `:84`; `Candidate` at `:103-109`;
    `Case.Extra []Request` at `:54`; `Input.Items []ItemFields` where
    `ItemFields` is `map[string]any`, so
    `matcher.ProjectItemFields(c.Input.Items[idx])` is assignable.
37. `stub.Server`'s `New`, `Misses`, `Consumed`, `Addr` and `Close`, the
    two-map answer order, and `Consumed()` distinct from `Misses()`.
38. `analysis` exports: `TieEntry` at `ties.go:56-66`, `checkTieEntry`
    at `:114-142`, `checkSkippedTieEntry` at `:148-159`, dispositions at
    `:19-29`; `AgentAdjudicator` at `adjudication.go:30`, the four
    dispositions at `:17-22`, `AdjudicationSubject` at `:35`,
    `StringHeuristicHypotheses` at `:86`, `CheckDefectFree`'s private
    subject map at `:224-229`, the pending row's five non-empty
    requirements at `:159-173`;
    `CheckOracleBugs(caseIDs, adjudicationIDs map[string]bool, l)` at
    `oraclebug.go:55`; `ProximityFlagged`, `ProximitySkipped`,
    `LoadProximityLedger`.
39. `oraclebug_test.go`'s table is exactly sixteen rows, exactly two of
    which pass a non-nil `adjudicationIDs`, and none of whose rows is at
    `OracleWrong`, so **B3's new rule breaks no existing row** and the
    mechanical signature swap is safe.
    `internal/harness/corpus_test.go` builds `adjIDs` at `:397-400` and
    calls `CheckOracleBugs` at `:406`, the sole production call site.
40. M4's mechanism: `internal/harness/oracle/ties.go:54` is
    `best := c.Expected.Candidates[0]`; `ties_derive_test.go:91` is the
    field mask line; `writeTieLedger`'s inline carry-forward is at
    `:133-141`; `derive_ties.py:118-126` returns exactly six measured
    fields and never a disposition or reason, so R10's silent-erasure
    argument is real.
41. `conform` exports and internals: `gradeUserAgent` at
    `request.go:21` (an invented, non-site-specific literal, reusable as
    stated), `RequestCaseFloor`/`RequestLookupFloor`/`RequestSearchFloor`
    = 215/303/48, `GradeRequests` at `:72`,
    `CheckCaseWeights`/`CheckCaseThresholds` at `defaults.go:21`/`:43`,
    `releaseLookupPrefix` at `translate.go:20`, `mergeProfile` at
    `:193`, `profileKeyAllowlist` at `:213`, `repoRoot` at
    `component_test.go:25`; `defaults_test.go:43, 55-61` does compare
    the echo against `DefaultTunables()` directly, so R13's move is real
    work.
42. `arch_test.go`: `TestEngineNeverImportsHarness` at `:141-146` and
    `TestEngineNeverReadsTheClock` at `:152-157`, both rooted at
    `internal/engine`, so `internal/engine/tag` joins both walks
    automatically.
43. `matcher` surfaces: `AlbumInfo` at `info.go:10-67`, `ItemFields` at
    `:124-155`, `ProjectItemFields` at `:189`, `DistanceForTrack` at
    `track.go:45`, `DistanceForAlbum` at `score.go:60`,
    `Distance.Distance()` at `penalty.go:96`, `Distance.Tracks` at
    `:41`, `CompilePatternOption` at `:368`, `Tunables` at
    `tunables.go:42`, `MaxRec` at `:61-67`, `StringDistPatterns` at
    `:217-224`, `Preferred` at `:227-235`, the three pattern tables at
    `:348`/`:365`/`:388`, twenty distance weights at `:256-277`,
    `MostCommonTags` and `Likelies.MBAlbumID` at `likelies.go:11-47`,
    `AlbumPair` at `album_tracks.go:11`.
44. `musicbrainz` surfaces: `ErrNotFound` `client.go:72`,
    `ErrUnavailable` `:84`, `SearchReleases` `:255`, `AlbumForID`
    `:296`, `AlbumSearchFor` `query.go:79`, unexported `vaLikely`
    `query.go:108` with a doc comment that already says 3d decides its
    fate, `ClassBackground` `limiter.go:22-26`,
    `DefaultInterval = time.Second` `:62`, `Unmetered()` `:354`.
    `musicbrainz` imports `matcher`, so R18's package-cycle argument for
    a third package is correct.
45. `go.mod` carries cobra and its two indirects and nothing else; Go
    1.27, so `math/rand/v2`, `for range n`, `cmp.Or` and
    `slices.SortStableFunc` are all available. `.golangci.yml` enables
    `errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`,
    `modernize`, `misspell`, `unparam`. `.gitleaksignore` pins exactly
    `internal/config/keys_test.go:generic-api-key:66` and
    `internal/config/keys.go:19-25` still carries `ownedWall` as a slice
    literal, so Task 0's item 2 expectation is right and rung 1b has not
    landed.

**Ledger and process claims.**

46. **B3 fires on nothing today.** All twenty probe cases of the four
    `oracle-wrong` hypotheses resolve to an adjudication row by
    `subject.case_id`, verified case by case on both lenses. The two
    probe cases with no row, `rec-strong-medium-boundary` (under
    `bug-threshold-boundary`) and `real-nonlatin-utada-first-love`
    (under `bug-locale-alias-ignored`), both sit under `oracle-correct`
    hypotheses. Spec `:300`'s "three probe cases" is superseded, and the
    hand-authored-fixture requirement follows.
47. **R1's table is right.** `bug-nonlatin-alias-selection` has four
    probes, two of whose own rows are `oracle-wrong`
    (`real-nonlatin-girls-generation`,
    `real-nonlatin-anna-vissi-essential`) and two `oracle-correct`
    (`real-nonlatin-alla-pugacheva-vol1`,
    `real-nonlatin-x-japan-art-of-life`). `Ungraded` = 6 `adv-assign-*`
    plus 2 alias = 8; `DefectFreeFloor` 10 plus `Ungraded` 8 = 18 =
    `ProbeCaseFloor` - 2, so Task 8's accounting assertion is
    arithmetically consistent (its constant's opacity is m17).
    `CheckDefectFree`'s `StringHeuristicHypotheses` predicate at
    `adjudication.go:86` does exclude the class, as claimed.
48. **Declared gaps 1 through 4 and the no-search-iff-strong fact.** No
    committed case sets `import.search_ids`; all 215 carry a full
    `mb_albumid` consensus, so `match_by_id`'s two early branches are
    unreachable; `timid` is false in all three profiles. The 167 cases
    with no logged search are precisely the 167 `strong` cases, and the
    48 with a search are precisely the 39 medium plus 4 low plus 5 none,
    one search each. Task 5's mutation 2 (167) and Task 7's mutation 3
    (48) and mutation 6 (21 multi-candidate cases) are all correct.
49. **The HISTORY guard citations are exact.** 3a carries ten
    rediscover-guards, 3b seventeen, 3c sixteen, as the plan says, and
    every guard the plan cites says what the plan says it says: 3a's 2,
    3 and 7; 3b's 2, 8, 10 and 11; 3c's 2, 3, 4, 5, 7c, 8, 9, 10 and
    13c. **STATUS agrees with Task 0's expected answers** on lane-1,
    lane-2 BLOCKED, doc lane IDLE, the ten declared gaps, and the
    2026-09-05 conductor-routing grant at `docs/STATUS.md:23-28`.
50. **Ruling B option (b), ruling F's trigger, the M8 unreachable
    surfaces, the precedence rule, the pending-divergence path and the
    assertion discipline** are all restated faithfully against spec
    `:93-95`, `:125-133`, `:234`, `:302`, `:304-314` and `:316-321`.
    R15's refusal to let the runner write into `corpus/` is right on its
    own terms. The spec's review note at `:362` is answered by R11. The
    ceiling arithmetic is internally consistent and its overrun is
    stated rather than shaded, the named split point (Task 5 / Task 6)
    and its three reasons are sound, and deliverable counts are honest:
    nine of ten tasks land four or fewer and Task 7's five is named
    rather than absorbed. `corpus/README.md:2091-2094` is indeed stale,
    so Task 9's drafted correction is right. All twenty resolutions R1
    through R20 are present and each names its rejected alternative.

---

## Conductor-owed, applied provisionally in the plan

1. **B2's `ModeDefault` design.** The plan applies the re-score option
   and marks it provisional at R2, R8, R9, Task 7 and Task 8.
2. **m16's package name.** The plan keeps `internal/engine/tag`, records
   the `tagio` pairing at R18, and adds the build spec's tree listing to
   Task 9's drafted corrections, marked provisional.
3. **m15's missing `docs/` contested-surface row.** The plan corrects
   the count to five of nine and marks the question of whether the row
   is written as a fourth conductor amendment.
