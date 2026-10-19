# Adversarial review: the rung 1b plan

Fresh-context review (2026-09-07) of
`plans/2026-09-07-rung-1b-config-registry.md` revision 1, in two
independent lenses. Lens 1 is Go design and contract fidelity: every
`internal/config` claim checked signature by signature against the tree,
the registry shape judged against `go-conventions` and the build spec's
Go bar, and the gitleaks rider re-measured on scratch copies under
gitleaks 8.30.1. Lens 2 is process and plan integrity: ceiling
arithmetic, provability of each acceptance criterion from the diff
alone, mutation-proof soundness, the 2.0 hook and prior-art statements,
the independence marks, the contract-change record and its spec
corrections, and the developer-docs writing standard.

The two reviews are merged here. Findings naming the same defect are
deduplicated, the stronger evidence kept and both lenses noted, and the
result renumbered M1 through M12 and m1 through m14. Every finding was
re-verified by the fold editor against the tree before folding, and
every numeric claim was re-measured rather than carried. Two reviewer
claims did not survive that check and are recorded as rejected below
with the measurement that retires them. Folded as revision 2 the same
day.

**No blocker.** Neither lens found a defect that stops the launch. The
plan's design decisions all survive review: R1's value-type registry,
R2's two distinct class types, R3's panic plus gate-resident test, R4's
rung 1 rows staying inside `internal/config`, R9's rename over a
`.gitleaks.toml`, and R15's order preservation are each argued from the
tree and each hold. The defects concentrate in four places the plan
reasons about from its own prose rather than from the artifact: three
acceptance criteria that cannot be satisfied as written, three mutation
proofs that cannot fire, five citations that moved, and two counts
carried from the inventory without re-measurement.

---

## Corrections to the reviews' own evidence

**1. Lens 1's citation finding calls five parallel-approach anchors
wrong.** It reports the config-wall ownership row at `PA:173` rather
than `PA:158`, and the contested-surface table at `PA:171-199` rather
than `PA:150-186`. Re-measured against committed HEAD, the plan is
right: `git show HEAD:docs/superpowers/specs/2026-09-03-parallel-build-approach.md | sed -n '158p'`
returns the config-wall row, and the table header sits at `:156`. The
working tree carries an uncommitted fifteen-line Phase 0c completion
note inserted after `:54`, which shifts every line below it down by
fifteen. Lens 1 measured the working tree and lens 2 measured HEAD, and
both readings are correct for the tree they read. The claim is rejected
as phrased. The underlying fact, that the anchors are stale against the
tree 1b will run in, is folded as M10's second half.

**2. Lens 1 reports the rung 4 skeleton row at `:1649` rather than
`:1647`.** Re-measured, `2026-09-05-rung-4-shadow.md:1649` is indeed the
`cmd/dubplate-server` row. The plan carries no such anchor: `plan:752`
cites `PA:182` for the skeleton quotation, which is correct, and the
inventory's own `:1647` attribution was already corrected by the plan.
The claim is rejected as a plan finding and stands only as an inventory
nit that the plan has already fixed.

---

## BLOCKERS

None.

---

## MAJORS

### M1. Task 4's shadow-grep criterion is unsatisfiable against R5's own "survives verbatim" constraint

**Lens.** Go design and contract fidelity.

**Plan lines.** The acceptance criterion at `plan:1430-1432`, "`grep -ri
shadow internal/config` returns nothing outside a comment that names
rung 4 as an example"; Task 4 outcome 3 at `plan:1397-1400`,
"`internal/config` names no rung-4 concept. No `Shadow` field, no
`ShadowIntegration` type, no shadow string"; and R5's constraint
restated at `plan:1412-1417`, requiring `TestLoadShadowIntegration`'s
eight rows and every `wantErr` assertion to survive verbatim.

**Evidence.** `TestLoadShadowIntegration` lives at
`internal/config/config_test.go:255-347` and its eight rows are built
entirely from `DUBPLATE_SHADOW_*` literals, four of the eight carrying
`wantErr` substrings that name those keys. Measured case-insensitively,
`internal/config` carries 41 `shadow` hits across its three test files:
33 in `config_test.go`, 4 in `keys_test.go`, and 4 in
`thresholds_test.go`. The plan never says the test changes package.

**Consequence, built as written.** An implementer cannot satisfy both
statements. Both ways out break something the plan pins: moving the test
makes "survives verbatim" no longer a signature-only edit, and renaming
the fixture keys to a generic class breaks every `wantErr` string the
constraint protects.

**Fix.** Decide the test's home in the plan. Split it: keep a generic
required-when-enabled table test in `internal/config`, driven by a
test-local fixture integration class that proves `Load`'s data-driven
resolution with no rung-4 name, and move the eight rung-4 rows verbatim
into `cmd/dubplate-server/shadow_test.go`, where the shadow class now
lives. Reword the criterion to "no rung-4 name in `internal/config`'s
non-test source, and none in its tests except a fixture class the test
itself declares", and re-point R5's "survives verbatim" at the moved
test.

### M2. Task 6's template classification test omits the integration family and reds the gate against 4b's greenfield template

**Lens.** Go design and contract fidelity.

**Plan lines.** Task 6 outcome 2 at `plan:1615-1617` and its acceptance
criterion at `plan:1652` pin the classification direction as "every
`DUBPLATE_*` name a template assigns classifies as owned, reserved, or
superseded, never unknown", asserted "against every template present in
`env/`".

**Evidence.** That is the existing three-way walk at
`internal/config/config_test.go:202-213`, calling `ownedNow`,
`reservedClassOf` and `supersessionOf` with no integration branch. It
passes today only because the one shipped template carries zero shadow
names: `grep -c DUBPLATE_SHADOW env/dubplate.env.template` returns 0.
Pass 4b's named deliverable is `env/dubplate-server.env.template` "with
its literal rows" (`2026-09-05-rung-4-shadow.md:1632`), and 4b's own
plan lists six `DUBPLATE_SHADOW_*` rows as template literals
(`plans/2026-09-07-rung-4b-deploy.md:1128-1133`). A shadow key resolves
through `integrationClassOf`, the fourth branch of `rejectNamespace`
(`internal/config/config.go:182`), which the test does not walk.

**Consequence, built as written.** The test reports six names as
"classifies as unknown" and `bash scripts/check.sh` goes red for every
lane.

**Fix.** Add the integration family to the classification direction in
the plan text: "owned, reserved, superseded, or an integration class's
flag, `Required` or `Optional` key, never unknown". State it in outcome
2, in the acceptance criterion, and in the test's doc comment. Note in
Task 0 check 4 that a greenfield template carrying shadow rows makes the
fourth branch mandatory rather than optional. The fix interacts with M1:
once rung 4's class is declared in `cmd/dubplate-server`, an
`internal/config` test cannot see it, so the classification test needs
the registry it classifies against passed in.

### M3. R11's registration check cannot catch the hazard surprise 11 names, because `KindInt` is the zero value

**Lens.** Go design and contract fidelity.

**Plan lines.** R11 at `plan:575-595` states the hazard as "a float
threshold whose `Kind` was never set would report as an environment
parse failure rather than as the registration bug it is", and resolves
it by refusing "an unnamed kind" at registration. Task 5's acceptance
criterion at `plan:1545-1546` requires refusing "a kind that names no
parser".

**Evidence.** `KindInt` is `iota`'s zero (`internal/config/thresholds.go:15-18`),
so a threshold whose `Kind` was never set has `Kind == KindInt`, a named
kind that passes any enum-membership check. The criterion as written
catches only an out-of-range `ThresholdKind`, which nobody writes by
accident. Concretely, rung 3's forward carry hands over `StrongRecThresh`
with default 0.04 (`internal/engine/matcher/tunables.go:62-66`);
registered without `Kind: KindFloat` it registers clean, then
`thresholdStatus` (`cmd/dubplate-server/doctor.go:112-127`) falls to its
`default` arm, runs `ParseInt("0.04")`, and doctor prints
`invalid: invalid default "0.04"`.

**Consequence, built as written.** The reported failure is exactly the
one R11 claims to have moved, and Task 5's mutation 2 (`plan:1561-1563`)
narrates a proof the design cannot deliver.

**Fix.** Make the registration check parse the declaration's own
`Default` under its declared `Kind` rather than merely check that the
`Kind` is in range. `Threshold{Key: X, Default: "0.04"}` with the zero
`Kind` then fails registration with `X: invalid default "0.04"`, caught
by the gate-resident test that constructs the production registry, which
is the altitude R11 wants. This also folds `ParseInt`'s and
`ParseDuration`'s existing invalid-default line (`thresholds.go:60-63`,
`:84-87`) up to registration. Restate outcome 3 and the acceptance
criterion in those terms, and rewrite mutation 2 to remove the default
validation and observe the doctor row.

### M4. Task 1's gitleaks rename covers one of the file's two tables, and an insertion into the other reds the gate again

**Lens.** Go design and contract fidelity.

**Plan lines.** R9 at `plan:503-507` and Task 1 outcome 1 at
`plan:1038-1041` scope the fix to "the struct field named `key` in
`TestReservedClassOf`'s table". Task 1's mutation 2 (`plan:1092-1094`)
claims the added row proves the fix "generalizes to the whole table".

**Evidence.** `internal/config/keys_test.go` has a second table with the
same field name: `TestSupersessionOf` at `:103-113`, whose rows are also
`{key: "DUBPLATE_..."}`. Measured on scratch copies under gitleaks
8.30.1, renaming only `TestReservedClassOf`'s field and adding the
DUBPLATE_R2_BUCKET row (a high-entropy key with `wantFound` false) to
`TestSupersessionOf` reports `leaks found: 1`. Renaming the field in
both tables scans clean (`no leaks found`).

**Consequence, built as written.** The rider's mandate is that no future
insertion reds the gate for every lane. After Task 1 as scoped, an
insertion into `TestSupersessionOf` does exactly that, and the pass
carries a fix that looks complete and is not.

**Fix.** Rename the `key` field in both tables, and put the discipline
comment where a later author of either table reads it. Restate mutation
2 as "add a high-entropy `DUBPLATE_*` row to each table in the file in
turn, still zero findings", so the mutation covers the file rather than
one table.

### M5. The collision-class list disagrees between O5 and Task 2, and one class gets no mutation

**Lens.** Process and plan integrity.

**Plan lines.** O5 at `plan:290-298` names five collision classes. Task
2 outcome 4 at `plan:1136-1145` lists six, adding "a wall row that is
neither a prefixed key nor a status row with `Pending` text", which
appears nowhere in O5, and splitting duplicate-within-a-family from
cross-family. Task 2's acceptance criterion at `plan:1177-1179` then
says "each of the six collision classes in outcome 4", while the Global
Constraint at `plan:883-885` promises "each collision class in O5 gets
one test that fires and one mutation proving it fires".

**Evidence.** Counting Task 2's named mutations (`plan:1190-1210`)
against its own six classes: duplicate-per-family is mutation 1,
cross-family is 2, prefix is 3, the status-row carve-out is 6, and the
threshold-kind class is deferred to Task 5's mutation 3
(`plan:1564-1565`), explicitly and correctly. The integration-class
check, an empty `Flag` or a `Flag` also appearing in its own `Required`
or `Optional` set, has a required test and no mutation anywhere in the
plan.

**Consequence, built as written.** A `diff-reviewer` checking Task 2
against O5 finds a sixth check with no outcome behind it, and the pass's
strictest constraint is one class short of the coverage it promises.

**Fix.** Make O5 the single authoritative list at six classes, matching
Task 2 outcome 4 verbatim. Add a Task 2 mutation: remove the
integration-`Flag` check, and both the empty-`Flag` test and the
`Flag`-in-`Required` test fail. Note in O5 that the kind class's
mutation lives in Task 5, so the one-mutation-per-class promise is
auditable across the two tasks.

### M6. Task 3's mutation 1 cannot fire, so R15's order property lands unguarded

**Lens.** Process and plan integrity.

**Plan lines.** `plan:1327-1329`: "Sort the reserved classes by rung
inside the registry, and `TestDoctorCmd` fails on block order, reporting
`streamer` before `decision and review queues`. This is Resolution R15's
proof."

**Evidence.** `cmd/dubplate-server/doctor_test.go:89-98` builds every
expected reserved line by iterating `config.ReservedClasses()` itself,
so a registry that sorts returns sorted and the test's own expectations
sort with it. Every line matches and the test passes. The two literal
spot checks use `strings.Contains` (`doctor_test.go:144-152`,
`:167-196`) and pin presence, never order. A grep of the module for
`decision and review` returns two hits, `internal/config/config.go:23`
and `internal/config/keys.go:127`, and no test.

**Consequence, built as written.** R15 (`plan:651-659`) and O3
(`plan:262-266`) rest on a guard that does not exist, and the mutation
is the self-derived-expectation tautology the plan's own assertion
discipline forbids (`plan:879-885`).

**Fix.** Give Task 3 an order assertion an implementer can prove from
the diff. Extend `TestDoctorCmd_LiteralSpotCheck` to assert that the
`decision and review queues` line's index in stdout is less than the
`streamer` line's, with a comment naming R15 and the fact that rung 13
precedes rung 12. Restate mutation 1 as "sort the reserved family by
rung, and that new order assertion fails", and say plainly that
`TestDoctorCmd` is order-blind because it derives its expectations from
the same accessor.

### M7. Doctor's exhaustiveness assertion already exists, so Task 6's outcome 1 and mutation 1 prove nothing new

**Lens.** Go design and contract fidelity.

**Plan lines.** Task 6 outcome 1 at `plan:1608-1613` and O6 at
`plan:300-307` present "a test asserts doctor prints exactly one line
per registry entry ... a rung that edits `doctor.go` to add a row fails
the test" as the mechanism this pass builds.

**Evidence.** `cmd/dubplate-server/doctor_test.go:67-129` already does
it. It walks the four families in order incrementing `li`, and closes
with `if li != len(lines) { t.Errorf("stdout has %d extra line(s) beyond
owned+reserved+threshold rows", ...) }` at `:127-129`. Task 6's mutation
1 (`plan:1662-1664`) fires today against the existing `TestDoctorCmd`
with no new test written. Neither the plan nor the inventory mentions
`doctor_test.go:127-129` anywhere.

**Consequence, built as written.** The mutation cannot distinguish Task
6's work from what already ships, and the task's largest stated outcome
is already landed.

**Fix.** Say in Task 6 that the count assertion is already landed at
`doctor_test.go:127-129` and survives the migration. Scope the task's
new work to the direction that genuinely does not exist: the
scratch-registry direction at `plan:1649-1650`, which is only reachable
once the registry is a value. Re-aim mutation 1 at it: remove the new
declaration's family from the scratch registry's accessor, and the
scratch test fails while the existing count test still passes.

### M8. O2's "no lane edits `keys.go` after 1b" is falsified by R4's own design, and no rule covers a rung retiring its reserved class

**Lens.** Go design and contract fidelity.

**Plan lines.** O2 at `plan:258-261` claims the contested-surface row
"becomes true rather than aspirational", and the forward criterion at
`plan:115-119` says `internal/config/keys.go` carries "rung 1's own
declarations and nothing else". R4 at `plan:400-417` keeps all eight
reserved classes in `keys.go` and adds that "a reserved class whose
owning package does not exist yet stays here too, and moves to that
package when the package is born".

**Evidence.** Every one of the eight belongs to a future rung. Measured
from `internal/config/keys.go:57-142`: auth and identity (rung 5), serve
(5), backup targets (6), intake (7), pipeline (9), monitoring (9),
decision and review queues (13), streamer (12). The first consumer rung,
rung 5, must delete two class literals from `keys.go` to move `auth and
identity` and `serve` into its own package, and rungs 6, 7, 9, 12 and 13
each do the same.

**Consequence, built as written.** The plan gives no rule for that edit
and no reconciliation with the contested-surface row it says becomes
true, so the first rung to move a class has to invent the rule under
time pressure.

**Fix.** State the rule in the forward criterion and in Task 7's carry
3: a rung may delete its own reserved class from `keys.go` in the same
commit that declares it in its own package, and that deletion is the
only edit to `keys.go` any lane makes after 1b. Soften O2 to match: no
lane adds to `keys.go`, and a lane removes only the class it is taking
over. Pin it with a registration-time check that the moved class
registers under the same name and key set, so the move cannot silently
drop a key.

### M9. The reserved family carries 29 keys, and the plan says 27 in two places

**Lenses.** Both. Lens 1 filed it as a minor and lens 2 as a major; the
fold takes the major, on lens 2's argument that Task 3 turns the figure
into a migration outcome.

**Plan lines.** `plan:966` (Task 0 item 6) expects "`ReservedClass`
three fields with eight classes over 27 keys" and instructs "Report
every count mechanically". `plan:1262-1263` (Task 3 outcome 1) makes it
an outcome: "the eight reserved classes with their 27 keys in table
order".

**Evidence.** `awk '/^var reservedClasses/,/^}$/' internal/config/keys.go
| grep -c '"DUBPLATE_'` returns **29**, with no duplicates: auth and
identity 8, serve 2, backup targets 6, intake 2, pipeline 3, monitoring
4, decision and review queues 1, streamer 3. The error is inherited from
the inventory at `:92`, which the plan's primary-inputs rule
(`plan:35-40`) declares binding.

**Consequence, built as written.** Task 0 reports a divergence that is
not a divergence, and the conductor has to adjudicate it before Task 1
can dispatch. An implementer reconciling Task 3's outcome down to 27
would drop two reserved keys, which reclassifies them as unknown and
turns a documented reservation into a load error for any adopter who
sets one.

**Fix.** Correct both plan sites to 29 and note the inventory's `:92` as
a corrected figure. Add one sentence to Task 3's acceptance criteria:
the migrated reserved family carries 29 keys across 8 classes, and any
count change is a stop-and-report.

### M10. Two spec anchors point at the wrong lines, and five more shift by fifteen in the tree the pass will run in

**Lenses.** Both, independently, against different trees. See the
corrections section above.

**Plan lines.** `plan:141` and `plan:217` cite `BS:186-188` for "Rung 1
derives the complete table from this classification and it becomes the
config core's contract". `plan:239-242` cites `rung-4:1693-1695` for the
standing `.gitleaksignore` staleness question. `plan:43`, `:259`,
`:751`, `:752` and `:755` cite `PA:150-186`, `PA:158`, `PA:183`,
`PA:182` and `PA:177`.

**Evidence.** In `2026-09-02-greenfield-server-build.md` the
complete-table sentence is at **182-183**; `:186` is blank and
`:187-190` open the defaults-discipline bullet, which the plan
separately cites as `BS:189-199`. In `2026-09-05-rung-4-shadow.md` the
`.gitleaksignore` staleness question runs **1695-1697**; `:1693-1694` is
the tail of the lane-registry bullet. The five parallel-approach anchors
are all correct against committed HEAD, and all shift down by fifteen in
the working tree, which carries an uncommitted Phase 0c completion note
inserted after `:54` (`git diff --stat` reports 15 insertions on that
file). `PA:54`, `BS:167-175`, `BS:189-199`, `rung-4:1632`, `:1650` and
`:1687-1690` all verify correct.

**Consequence, built as written.** Task 7's acceptance criterion
(`plan:1791-1792`) requires each correction to carry "its file, its
line, and its replacement sentence", so a drafter working from
`BS:186-188` either edits the defaults bullet or stalls. Task 0's
blocker 2 (`plan:933-938`) fires on the warm Phase 0c change with no
pre-ruled answer, stalling the launch on a known document-lane landing.

**Fix.** Re-point correction 1 at `BS:182-183` and correction 5 at
`rung-4:1695-1697`. Add a Task 0 item that re-reads the parallel spec's
Phase 0b paragraph and contested-surface table and reports the current
line numbers for the config wall, `cmd/dubplate-server`,
`internal/config`, `env/` and `docs/` ledgers rows, so the conductor
re-points the five citations in one edit. Pre-rule blocker 2 for this
specific change: the Phase 0c note is a known document-lane landing, and
Task 7 outcome 8 already expects to report 0c's state. Add a line to
Task 7's acceptance criteria requiring the drafter to re-read each cited
line before writing the replacement, and to report any that has moved.

### M11. Task 7's spec-correction list misses three sites the gitleaks fix makes stale

**Lens.** Go design and contract fidelity.

**Plan lines.** The five corrections at `plan:215-242`, and Task 7
outcome 3 at `plan:1709-1714`.

**Evidence.** Task 1 removes the pin entirely (`plan:1048-1051`). Three
places then describe a mechanism that no longer exists, and the plan
names one. `2026-09-05-rung-4-shadow.md:1474-1483` restates the
fragility in full, names `.gitleaksignore:1` and line 66 by value, and
states the standing rule "add below line 66, or re-fingerprint
`.gitleaksignore` in the same commit". `:1785-1786` carries the same as
risk 7 with its mitigation. The plan's correction 5 covers only the
staleness question at `:1695-1697`. Separately, the 0b rider at `PA:54`
has two clauses: "the work must re-fingerprint `.gitleaksignore`" and
the blast-radius sentence. The plan's correction 3 addresses the blast
radius alone, leaving the spec directing every future reader to
re-fingerprint a file the pass deleted.

**Fix.** Extend Task 7's list from five corrections to eight: add
`rung-4:1474-1483`, `rung-4:1785-1786`, and the `PA:54` re-fingerprint
clause, each with its replacement sentence. Update the acceptance
criterion's count so a reviewer can check it.

### M12. Task 7 must list sixteen inline resolution markers that no task is told to write

**Lens.** Process and plan integrity.

**Plan lines.** Task 7 outcome 4 at `plan:1715-1717` requires "the
sixteen resolutions listed with their inline markers, each naming the
file and the symbol the marker sits on", and `plan:1793` makes it an
acceptance criterion.

**Evidence.** Nothing in Tasks 1 through 6 instructs an implementer to
leave such a marker. The Resolutions preamble (`plan:323-326`) marks the
plan's own rows "provisional, conductor ruling owed". The Global
Constraints (`plan:847-908`) carry no marker rule. Task 2's constraints
(`plan:1156-1167`) require only the identity-source carve-out comment.
The repo has the convention: `docs/STATUS.md:46-47` records "provisional,
owner ruling N open" marks left by pass 4a-i, and
`cmd/dubplate-server/shadow.go:22` and `:38` carry two of them.

**Consequence, built as written.** The criterion is unprovable, because
there will be no markers to locate.

**Fix.** Add one Global Constraint: where a task applies a resolution's
shape, it leaves a one-line comment on the symbol naming the resolution
and "provisional, conductor ruling owed", and reports the file and
symbol in its return. Task 7's criterion then becomes checkable against
the task returns.

---

## MINORS

**m1. Task 5 depends on Task 3, and R1's blast radius omits
`internal/config`'s own two `Load` call sites.** Both lenses. Task 5 is
headed "Depends on Task 2" (`plan:1491`), and its file list
(`plan:1497-1506`) includes `internal/config/thresholds.go` and
`cmd/dubplate-server/doctor.go`, both of which Task 3 rewires: Task 3
retires `registeredThresholds` and makes `Register` and `Thresholds`
registry operations (`plan:1246-1248`, `:1275-1278`), and makes doctor's
four blocks read the registry (`plan:1273-1274`). Task 5's acceptance
criterion "Doctor prints a registered float threshold with its effective
value" (`plan:1547-1548`) needs Task 3's wiring. Separately, R1
(`plan:354-358`) says `Load`'s signature changes "at five call sites,
all inside `cmd/dubplate-server`". The six enumerated sites are exact
(`check.go:23`, `doctor.go:31`, `shadow.go:196`, `:262`, `:299`,
`check_test.go:107`), and `internal/config/config_test.go:152` and
`:320` also call `Load` and change with it. Task 3's file list covers
the work, so nothing is unbudgeted. Change Task 5's header to "Depends
on Tasks 2 and 3", restrict the reorder licence to "before Task 4, after
Task 3", and amend R1 to name the package's own two test call sites.

**m2. O6 and Task 6 count a doctor line per key for two families that
print one per class.** Both lenses. O6 (`plan:305-307`) and Task 6
outcome 1 (`plan:1608-1610`) say a rung that declares a key gets its
doctor row. Doctor prints one line per owned wall row and one per
threshold, but one line per reserved **class** and one per integration
**class** (`cmd/dubplate-server/doctor.go:41-46`, `:48-56`): eight lines
for 29 reserved keys, one line for the shadow class's six keys. Task 6's
criterion "Doctor's line count equals the registry's total entry count
across all four families" (`plan:1647-1648`) is satisfiable only if
"entry" means top-level declaration, which the plan never says. Say "one
line per declaration" and define the unit per family: a wall row, a
reserved class, an integration class, a threshold. Restate O6's promise
as "a rung that declares a new class, wall row or threshold gets its
doctor row", and note that a key added inside an existing class adds no
line by design.

**m3. Task 3's mutation 5 predicts the opposite of the property it
claims to prove.** Lens 2. `plan:1342-1344`: "Make `newRootCmd` build
the registry twice and pass the second, and every doctor threshold row
still prints, proving the wiring is a value pass rather than a hidden
global." Under R1's value registry, if registration runs against the
first construction and the second is passed to `newDoctorCmd`, the
second carries rung 1's rows but not rung 4's thresholds, so
`TestDoctorCmd_ShadowRows` (`doctor_test.go:158`, pinning
`DUBPLATE_SHADOW_FLOOR 720h0m0s` and `DUBPLATE_SHADOW_BUDGET 60`) fails.
If the whole registration sequence runs twice, the rows print, and so
would a package-level global. Neither reading delivers the stated proof.
Replace with the unambiguous form: register rung 4's thresholds onto a
first registry and pass a second, freshly constructed one to
`newDoctorCmd`, and `TestDoctorCmd_ShadowRows` fails on both threshold
lines. Cross-reference Task 2's mutation 5 as the freshness proof.

**m4. Task 6's mutation 3 fires at `templateKeys`'s open, never at the
zero-names guard.** Lens 2. `plan:1667-1671` says pointing the
classification test at a nonexistent template path makes it "fail on the
zero-names guard rather than passing vacuously". In
`internal/config/config_test.go:221-223`, `templateKeys` opens the path
and calls `t.Fatalf("open template: %v", err)`, so a nonexistent path
aborts before the guard at `:197-199` is reached. The guard goes
untested. HISTORY's own lesson at `docs/HISTORY.md:325-330`, "Mutation
proofs are run, never narrated", is why this matters. Change the
mutation to point the test at an existing file that assigns no
`DUBPLATE_*` names, such as a scratch temp file or `go.mod`:
`templateKeys` returns an empty slice and the guard fires with "template
path or format changed". Keep the nonexistent-path case as a second,
separately named mutation proving the open error is fatal.

**m5. Task 4 loses the enabled-versus-absent distinction for an
integration class that declares no `Required` keys.** Lens 1. Task 4
outcome 1 (`plan:1389-1392`) has `Load` resolve "every enabled
integration class's `Required` keys" and make them "reachable by class
name". Today `cfg.Shadow == nil` is the enabled test, relied on at
`cmd/dubplate-server/shadow.go:199` to refuse `observe`. Reached by
class name through resolved `Required` keys, an enabled class with zero
`Required` keys is indistinguishable from a disabled one, because
`Load`'s `resolved` map (`internal/config/config.go:229-232`) is empty
either way. That is the shape hook 6 asks for:
`2026-09-07-streamer-2-0-design.md:369-373` says address, proxy and
provisioning are per-driver capabilities a driver may decline to
declare. Add an acceptance criterion that `Load` records a class's
enabled state independently of whether it resolved any key, and one test
row for a class with an empty `Required` set: flag set loads clean and
reports enabled, flag unset reports not enabled. Name it as the hook 6
shape so rung 12 does not rediscover it.

**m6. The `rowStatuses` claim covers owned rows and threshold parse
errors, never reserved or integration rows.** Lens 2. `plan:207-209`,
`plan:896-898` and Task 3's mutation 3 (`plan:1334-1336`) say a registry
error that loses the `key: reason` shape "blanks every per-row doctor
status". In `cmd/dubplate-server/doctor.go`, `statuses` is consulted
only in the owned-wall loop at `:35`; reserved classes print a fixed
`"disabled (arrives at rung %d)"` at `:42`, integration classes a fixed
`"disabled"` or `"enabled"` at `:49-51`, and thresholds go through
`thresholdStatus` at `:59`, which does use `splitKeyReason`. A
reserved-class error that lost its shape would blank no doctor row.
Mutation 3's first clause is sound: `TestLoad_ContractTable`'s "reserved
key early" row asserts both `"backup targets"` and `"rung 6"`
(`internal/config/config_test.go:89-94`). Narrow the claim to owned-wall
rows (`doctor.go:35`) and threshold parse errors (`doctor.go:112-127`),
call it a package-wide convention elsewhere, and drop mutation 3's
second clause.

**m7. Task 0's template expectation of 42 names counts three `ND_`
names.** Lens 1. `plan:951` expects `env/dubplate.env.template` to carry
"42 names". The template's own parser
(`internal/config/config_test.go:218-247`) trims each line, requires the
`DUBPLATE_` prefix and an `=`, and finds **39**. Measured, the file
assigns 42 uppercase names of which 3 are `ND_PASSWORDENCRYPTIONKEY`,
`ND_SPOTIFY_ID` and `ND_SPOTIFY_SECRET`, which the inventory itself
lists separately at `:343`. Change the expectation to 39 `DUBPLATE_*`
names, so Task 0 confirms rather than contradicts.

**m8. Task 3's "no package-level mutable registry variable" criterion
names no symbol.** Lens 1. `plan:1315-1316` reads "`grep` finds no
package-level mutable registry variable in `internal/config`, and no
`withRegistry` helper". R4 (`plan:400-405`) keeps rung 1's declarations
in `keys.go`, which today means four package-level vars: `ownedWall`
(`keys.go:19`), `reservedClasses` (`:57`), `supersededKeys` (`:164`, a
mutable map) and `integrationClasses` (`:214`). A reviewer can read the
criterion either way. Name the symbol: `registeredThresholds` no longer
exists in `internal/config`, and `withRegistry` no longer exists in
`thresholds_test.go`. Add one sentence saying rung 1's four declaration
literals stay package-level by R4 and are read once by the constructor.

**m9. Task 1's independence claim is false: `keys_test.go` is also Tasks
3 and 4's.** Lens 2. `plan:1099-1101` says "Its three files are disjoint
from every other task's". Task 1's files are
`internal/config/keys_test.go`, `.gitleaksignore` and `scripts/check.sh`
(`plan:1028-1034`). Task 3's file list includes
`internal/config/{config,keys,thresholds}_test.go` (`plan:1249-1250`)
and Task 4's includes `internal/config/{config,keys}_test.go`
(`plan:1384`). Task 1's own rationale (`plan:1022-1026`) concedes the
overlap. Nothing operational turns on it, since the pass runs strictly
sequential. Restate as: independent in `.gitleaksignore` and
`scripts/check.sh`, which no other task touches; `keys_test.go` is
shared with Tasks 3 and 4, and Task 1 runs first so those two inherit a
file with no line pin.

**m10. Task 1's never-quote constraint is unsatisfiable as worded.**
Lens 2. `plan:1064-1066`: "Never quote the pinned fixture literal, in
code, in a comment, in the task report, or in a commit message." Task
1's own work is to rename the struct field while keeping the row
(`plan:1067-1069`), so the diff necessarily writes the literal in code at
`internal/config/keys_test.go:66`. The literal is already in
`keys.go:88` and `config_test.go:91` and fires nothing on its own; only
the `key`-like field name beside it does. Reword to: never reproduce the
triggering pair, a `key`-like field name beside the value, outside the
fixture row itself, and never quote the value in the task report or the
commit message.

**m11. Two of the four 2.0 hook characterizations do not match the spec
text.** Lens 2. `plan:730-733` describes hook 5 as "Rung 5's identity
source is a class shape with a discriminator rather than a boolean
flag", repeated in Task 3's hook statement at `plan:1346-1351`. Hook 5
(`2026-09-07-streamer-2-0-design.md:367-370`) says nothing about a
config class or a discriminator: it is about rung 6 minting a durable
surrogate user id, rung 5 resolving a roster identity to it, and
`streamer_username` becoming required, unique and stable. The
discriminator language comes from `internal/config/config.go:11-17`, the
package doc. In the same paragraph the plan says hook 12 "wants a
255-byte bound and a forbidden-character set, neither an int nor a
duration" (`plan:733-735`); the 255-byte bound at spec `:385-388` is
plainly an integer, so only the forbidden-character set supports the
claim. Hooks 6 and 14 are characterized correctly and the accounting is
exhaustive, four touched plus ten neutral against the spec's fourteen.
Attribute the discriminator to `config.go:11-17`, and keep only the
forbidden-character set as the shape no current kind expresses.

**m12. Cross-check one mixes the chain and pre-task categories,
overstating the derived ceiling.** Lens 2. `plan:816-817`: "Cross-check
one, the chain share: 1.15M at 4a-ii's 49% implies 2.35M for the pass."
The 49% figure is derived at `plan:795-796` as chain-only (1.32M of
about 2.7M, `docs/HISTORY.md:468-477`), while 1.15M is chain 1.05M plus
the Task 0 pre-task 0.10M, and 4a-ii's pre-task sat in the separate
0.26M line item outside the 1.32M chain. Applying a chain-only share to
a chain-plus-pre-task figure inflates the result: 1.05M divided by 0.49
gives 2.14M. Cross-check two has a smaller drift the other way: it
treats 0.18M as "simplify plus its review" (`plan:819`) where HISTORY
says "simplify 0.18M plus its review", and HISTORY's line items sum to
2.61M against "about 2.7M", leaving about 0.09M unallocated. Corrected,
the two land near 2.14M and 2.27M, agreeing within 6%. The 2.5M ceiling
stays conservative, so nothing downstream breaks, and the section
advertises "the arithmetic shown". Divide the chain figure alone by 49%,
carry the unallocated 0.09M in cross-check two, and keep 2.5M as the
higher corrected check rounded up.

**m13. The split trigger sits a full 1M above the pass ceiling with no
rule for the band between.** Lens 1. The ceiling is 2.5M with the flag
at 2.0M routed to the conductor (`plan:817-838`), and "a re-derived
ceiling past 2.5M" is one of the four things that would become an owner
touchpoint (`plan:100-101`). The split rule says "Take the split only if
Task 3's actual spend re-derives the ceiling past 3.5M"
(`plan:1889-1891`). Nothing states what happens between 2.5M and 3.5M,
which is the band a 40% overrun would land in, and rung 4's own M-band
estimate ran roughly 50% over (inventory `:8.2`). State one rule: a Task
3 re-derivation past 2.5M is a conductor decision between splitting at
the Task 4 and Task 5 boundary and raising the ceiling, and only a
re-derivation that also breaks a window cap reaches Geoff.

**m14. The prior-art table claims per-entry coverage of a list twice its
length.** Lens 2. `plan:688-691` says the plan "says so per entry rather
than skipping them", and the table (`plan:693-704`) carries ten rows
against twenty-one in `2026-09-07-streamer-2-0-design.md:402-424`, plus
the declined entries in the paragraph below it. Missing rows include
`ncruces/go-sqlite3`, `crypto/aes`, `internal/atomicfile`, the M3U and
Subsonic encoding row, Symfonium notes, the beets and MusicBee row, and
the declined `djherbis/fscache` and `xrash/smetrics`. The one entry that
bears on a config registry, `viper`, is declined explicitly and on the
record, which satisfies CLAUDE.md's substance. Either add the missing
rows with a shared "not applicable, no configuration surface" reason, or
change the framing to "every entry that could bear on process
configuration", so the claim matches the table.

---

## Verified sound

Every item below was checked against the tree by at least one lens, and
re-checked by the fold editor wherever a fold depended on it. This is
the do-not-re-litigate record for rung 1b.

1. Every `internal/config/config.go` citation is exact: the five-class package doc at `:1-45`, `ShadowIntegration` at `:80-88`, `wall` at `:95-98`, `newWall` at `:102-112`, the `get` choke point at `:115-121`, `integrationEnabled` at `:155-165`, `rejectNamespace` at `:173-205` with the `DUBPLATE_` prefix walk at `:175` and the flag self-report guard at `:185-189`, `Load` at `:212-250`, the `class.Name == "shadow"` branch at `:235-241`, and `errors.Join` at `:246-249`.
2. Every `internal/config/keys.go` citation is exact: `WallRow` `:8-14`, `ownedWall` `:19-25` with the identity-source `Pending` row at `:24`, `OwnedWall` `:28-30`, `ownedNow` `:34-41`, `ReservedClass` `:47-51`, the eight classes `:57-142` with `decision and review queues` (rung 13) preceding `streamer` (rung 12), the monitoring class `:116-125`, the streamer class's three Navidrome names `:133-141`, `supersededKeys` `:164-183` with eleven entries, `supersessionOf` `:187-190`, the Required/Optional semantics doc `:198-202`, `IntegrationClass` `:203-209`, the 1b absorption note `:211-213`, and the shadow class `:214-229`.
3. Every `internal/config/thresholds.go` citation is exact: `ThresholdKind` with `KindInt` as `iota`'s zero at `:13-18`, `Threshold`'s four fields at `:24-29`, `registeredThresholds` at `:34`, `Register` as a bare append at `:40-42` with the no-`init` rationale at `:36-39`, `Thresholds` returning a clone at `:45-47`, and both parsers' bad-default-is-a-registration-bug line at `:60-63` and `:84-87`.
4. R1's call-site claim is complete for the command package. `config.Load` is called at `cmd/dubplate-server/check.go:23`, `doctor.go:31`, `shadow.go:196`, `:262`, `:299` and `check_test.go:107`, and nowhere else outside `internal/config`. Inside the package it is called at `config_test.go:152` and `:320` (see m1).
5. R6's foundation holds mechanically. `go list -deps` over every package under `internal/` except `internal/config` itself returns no importer of `github.com/glw907/dubplate/internal/config`. The five internal files that match the string are comments only, including `internal/engine/matcher/tunables.go:50-54`, which names the four knobs, states the deferral in writing, and lists them as three `float64` fields plus one `int`, confirming the `KindFloat` need.
6. Doctor's shape is as the plan describes: four blocks in fixed order at `doctor.go:34-62`, the `%-32s %s` pin at `:74-78`, `rowStatuses` splitting on the first `": "` via `splitKeyReason` at `:84-106` and `:134-140`, and `thresholdStatus`'s `default` branch parsing any unnamed kind as an integer at `:112-127`.
7. The tests the migration must hold are as described: `TestLoadShadowIntegration` at `config_test.go:255` with exactly eight table rows, `TestLoad_TemplateKeysClassified` at `:196` with its zero-names guard at `:197-199` and the hard-coded `../../env/dubplate.env.template` path at `:221`, `TestIntegrationClassesCoverEveryShadowKey` at `keys_test.go:125-146`, `TestOwnedWall_ReturnsCopy` at `:50-57`, `TestThresholds_ReturnsCopy` at `thresholds_test.go:40-51`, and `withRegistry`'s save-and-restore at `:11-17`.
8. The `sync.OnceFunc` workaround and its regression test sit exactly where the plan says: `registerShadowThresholds` at `cmd/dubplate-server/shadow.go:44-52`, called as `newRootCmd`'s first statement at `main.go:55`, with `TestNewRootCmd_RegistersShadowThresholdsOnce` at `main_test.go:36-52`. The two threshold literals are at `shadow.go:23-29` and `:38-42`, each carrying a "provisional, owner ruling N open" marker.
9. Task 5's `Threshold` comparability constraint is real and load-bearing: `thresholds_test.go:34` compares `Threshold` values with `!=`, so a kind that made the struct incomparable would break a test the pass may not weaken.
10. The gitleaks starting state is exactly as Task 0 check 3 expects: `.gitleaksignore` is one line, `internal/config/keys_test.go:generic-api-key:66`, and line 66 is the first row of `TestReservedClassOf`'s seven-row table. `scripts/check.sh:98-109` runs `gitleaks detect --no-git -s . --redact` and hard-fails when gitleaks is absent at `:36-40`, so the leg genuinely gates.
11. R9's Option A works as claimed, measured under gitleaks 8.30.1. A scratch copy of `keys_test.go` with the table's `key` field renamed scans clean; the unmodified file reports one finding. The rename is the fix rather than a suppression. Its scope is the finding at M4.
12. R4's constructor is safe against its own collision gate. A mechanical extraction of the `DUBPLATE_*` literals from `keys.go`'s four tables finds no duplicate within any family, no key shared across two families, and every declared key carrying the `DUBPLATE_` prefix, so registering rung 1's own rows through panic-on-duplicate entry points cannot panic at startup.
13. The Go-shape resolutions are sound against `go-conventions` and the build spec's Go bar at `2026-09-02-greenfield-server-build.md:200-207`. R1 chooses a concrete value type over package-level `init` registration and cites the reason rung 1 already wrote at `thresholds.go:36-39`. R2 keeps two concrete class types rather than one with dead fields. The plan defines no interface, no builder, no functional options and no package-level default registry, and R1's value type is what makes a registry resettable per test, retiring both the `sync.OnceFunc` and `withRegistry`.
14. The plan specifies outcomes, constraints and acceptance criteria throughout and writes no implementation code. Every type it names (`WallRow`, `ReservedClass`, `IntegrationClass`, `Threshold`, `ThresholdKind`) already exists in the tree, and the registry's own shape is left to the implementer inside stated constraints.
15. The `viper` decline meets CLAUDE.md's prior-art rule in substance: a config-registry rung is exactly where a configuration library gets reached for, the decline is stated with its argument, `testing/fstest` is recorded as consulted and offering nothing, and no external dependency is proposed. `go.mod` carrying cobra plus two indirects is consistent with that. The table's coverage claim is the finding at m14.
16. All four HISTORY carries are explicitly dispositioned. `KindFloat` lands (R6, Task 5). The four knobs are deferred with a reason and the rung 3 pass named (R6, citing `HISTORY:434`, `:1063`, `:1434` and `STATUS:43-44`). `AdminEmail` is deferred as a product decision with the owner touchpoint named (R7, citing `HISTORY:1470-1475`). `integrationClasses` is absorbed as data with the Required/Optional split preserved verbatim (R5, Task 4 outcome 5).
17. The Resolutions section covers every open reading. All fourteen inventory surprises map onto R2, R3, R5, R6, R8, R9, R10, R11, R12, R13, R14 and the spec corrections, and R1, R4, R7, R15 and R16 close the five design readings the inventory left open. Each row states what changes if the ruling goes the other way, and no task blocks on any of them.
18. Task 0's expectations that could be checked are exact, apart from the two counts at M9 and m7: `config.go` 250 lines, `keys.go` 254, `thresholds.go` 96, the package 1255 lines; `Register` a bare append at `thresholds.go:40-42`; `KindInt` the zero value; `withRegistry` at `thresholds_test.go:12-17`; the `sync.OnceFunc` at `shadow.go:49-52` with its call at `main.go:55`; `ownedWall` five rows with the identity-source `Pending` row; `supersededKeys` eleven entries; `decision and review queues` preceding `streamer` in table order.
19. Task 0's first three items are the right blockers, and blocker 2 would in fact catch the warm uncommitted change now sitting in the tree. Item 3's instruction to report the pin without quoting the fixture literal is correct, since gitleaks scans the tracked worktree with `--no-git`.
20. The staleness pre-task names 4b's files correctly against `rung-4:1632`: the two systemd units, `env/dubplate-server.env.template` and the existing template's header amendment, `deploy.sh`'s six edits with its template constant at `:26`, and the two `-ldflags` version stamps, plus `main.go` and `docs/STATUS.md`'s lane registry. The template header amendment cannot break `TestLoad_TemplateKeysClassified`, which parses only assigned `DUBPLATE_` lines, and the plan's edge 1 names that coupling.
21. The plan degrades correctly if 4b has not run. `STATUS:33` records lane-1 as IDLE pending 4b's plan, so 1b may reach the tree first. Task 0 item 1's expected "no" still holds, item 4 reports one template, and R10 leaves Task 6's completeness direction a documented rule with Task 7 carrying it forward.
22. Task 4's constraints are exact against the tree. `TestLoadShadowIntegration` has eight table rows, and exactly two assertions reach into `cfg.Shadow` (`config_test.go:330` and `:333`); every `wantErr` assertion is untouched by the move. Its six mutations each name a specific table row that would fail, and mutation 1 does fire against the reworded assertion, because a flag-unset load must yield no resolved shadow values.
23. Task 3's mutation 2 fires: `TestDoctorCmd_LiteralSpotCheck` hard-codes the identity-source pending line (`doctor_test.go:149-152`). Mutation 3's first clause fires (see m6). Mutation 4 is properly hedged, telling the task to report an uncovered choke-point bypass as a finding rather than adding a test.
24. The contract-change section's "what stays" list is complete and accurate against the tree: the five classes and their wording, `errors.Join` at `config.go:246-249`, the `wall.get` choke point at `config.go:115-121` with both threshold parsers routing through it, the Required/Optional semantics at `keys.go:195-202`, the identity-source pending row, table order, doctor's four blocks and the `%-32s` pin at `doctor.go:76`, and universality. Each property is attached to at least one task's acceptance criteria.
25. The 2.0 hook accounting is exhaustive and internally consistent: the spec carries fourteen hooks, the plan names four as touched (5, 6, 12, 14) and ten as neutral, and every task carries its own hook statement. Hook 6 is characterized correctly, and Tasks 2 and 4 land it as an enabling seam without building a driver or moving a Navidrome key.
26. Independence marks are true for Tasks 0, 2, 3, 4, 5, 6 and 7. Task 5's "independent of Task 4" is exact: the two file lists overlap only in `doctor_test.go`, and the append-rather-than-edit rule covers it. Task 1's mark is the finding at m9.
27. Deliverable counts are stated per task and none exceeds four, so CLAUDE.md's pass-sizing flag is not tripped. The split point, the reason to keep one pass, and a numeric trigger are all stated; the trigger's altitude is the finding at m13.
28. Execution wiring matches the owner grants exactly: workflow mode, `pass-execute.js` per task, `cairn-implementer` reading `.claude/agents/dubplate-implementer.md`, `diff-reviewer` on `claude-opus-5`, `bash scripts/check.sh` as the gate, one re-dispatch on `fix`, the conductor owning the `code-simplifier` dispatch and every ledger edit, and review scaled to the in-chain reviewer because 1b is not one of rungs 5, 12 or 16.
29. The ceiling's inputs match `HISTORY:468-481` exactly: 1.32M over six tasks gives 0.22M per chain task, 1.32M of 2.7M gives the 49% chain share, 0.22M times two thirds rounds to 0.15M, and seven chain tasks plus a 0.10M pre-task gives 1.15M. Even after the correction at m12, both cross-checks land under 2.3M, so the 2.5M ceiling and the 2.0M flag stay conservative and the flag routes to the conductor per this window's grants.
30. The plan honors HISTORY's "mutation proofs are run, never narrated" lesson at the process level: every task names its mutations, the Global Constraints require each measured outcome, and Task 7's acceptance requires a complete mutation table across Tasks 1 through 6. Three individual mutations fail on their own terms (M6, m3, m4).
31. Writing voice holds. `tellgrader` reports zero tells per 1000 words across 14,750 words in the docs register, no findings, and a cadence coefficient of variation of 0.74, which is healthy sentence-length variance.
