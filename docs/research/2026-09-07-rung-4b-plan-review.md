# Adversarial review: the rung 4 pass 4b plan

Fresh-context review (2026-09-07) of
`plans/2026-09-07-rung-4b-deploy.md` revision 1, in two independent
lenses: systemd, the deploy path and box safety, verified line by line
against the tree at `c98a916`, `~/.claude/skills/vps-conventions/SKILL.md`,
the rung 4 spec at revision 2 and `docs/HISTORY.md:385-481`, with the
proposed unit files written to a scratch copy of `systemd/` and run
through `check.sh:137-139`'s exact command; and process and plan
integrity, verified against the same inputs plus the 4b planning
inventory. Folded as revision 2 the same day. The merged 42-item
"Verified sound" list below is the do-not-re-litigate record for this
pass, alongside the rung 4 spec review's own list and 4a's.

The two reviews are merged here: findings naming the same defect are
deduplicated, the stronger evidence kept and both lenses noted, and the
result renumbered B1, M1 through M8, m1 through m13. Every finding was
re-verified by the fold editor against the tree before folding, and
every numeric claim was re-measured rather than carried. No finding was
rejected. Five pieces of the reviews' own evidence did not survive
re-measurement in the form they were written; each is corrected below
and none of the five retires its finding. Five items are conductor-owed
rather than folded silently, and each is applied provisionally in the
plan with an inline marker.

Both lenses agree the plan is well grounded on the artifacts it can
reach. The two unit files pass the gate's systemd leg empirically, the
leaf grant is argued from the store's own twelve write sites rather than
assumed, the hardening set is a faithful copy of the estate's template,
the step-count and template-row arithmetic are right, the sitting
boundary is unambiguous, and no task in the unattended half writes the
box. The defects concentrate in four places: what `deploy.sh` does when
an argument is omitted, what the workstation's own environment already
carries, the ordering the render helper can physically achieve, and the
line numbers of four documents that moved after the citations were
written.

---

## Corrections to the reviews' own evidence

**1. Both reviews' `PA:` spans are stale, and so are the plan's.**
`docs/specs/2026-09-03-parallel-build-approach.md` carries
an uncommitted fifteen-line addition in this worktree, so its rows sit
at different numbers depending on which copy is read. Measured at HEAD
`c98a916`: `docs/` ledgers `:177`, the state directory layout `:178`,
`internal/engine/shadow` `:180`, `cmd/dubplate-server` `:182`, and
`systemd/`, `scripts/deploy.sh`, `env/` `:184`. In the current worktree
the same rows are `:192`, `:193`, `:195`, `:197` and `:199`. Review 2's
`:178`/`:179`/`:180`/`:181` reading matches neither copy exactly and was
taken while the file was mid-edit. The finding stands for two of the
plan's five citations (`PA:179` and `PA:181`, which should be `:178` and
`:180` at HEAD); the other three were right. The fold corrects both and
adds the row names, because a bare line number into a file another agent
is editing is a citation that expires.

**2. Review 1 gives the atomicfile citation as `store.go:14`, `:100`,
both wrong.** Only `:100` is wrong. `internal/engine/shadow/store.go:14`
is the import and is correct as the plan writes it; the first
`atomicfile.WriteFile` call is `:102`, and `:100` is a `return false,
nil`. The fold corrects `:100` alone.

**3. The two lenses disagree on where the mutation-proof rule sits.**
Review 1 gives `spec:1831-1832` and review 2 gives `:1831-1837`.
Re-measured, the sentence "The observation mechanism's own honesty gets
one mutation proof per checker" opens at `:1831` and closes at `:1832`;
`:1833` onward is its justification. The plan's `:1816-1818` is the
review-scaling sentence. The finding stands and the fold uses
`:1831-1832`.

**4. Both reviews call the render and the install "twelve steps
apart".** Re-measured, `deploy.sh`'s render (`:171-194`) and its env
install (`:291-316`) are separated by five announced steps: `:199`,
`:212`, `:230`, `:273` and `:288`. The finding is unaffected, since one
intervening remote step is enough to break the property, and the fold
uses the measured figure.

**5. Review 2 cites `SKILL.md:11` for the `StateDirectory=` row.** The
directory-lifetimes table runs `:15-19` and the `StateDirectory=` row is
`:18`; `:11` is the skill's own preamble. Review 1's `:15-19` is the
table. The finding stands unchanged: neither the row nor the option
appears in R1.

---

## BLOCKERS

### B1. The plan tells the implementer that omitting `HOST` aborts. On this workstation it resolves the production box instead, and a bare `bash scripts/deploy.sh` applies the full deploy to `5.78.148.42`

**Lens.** systemd, deploy path and box safety.

**Plan line.** The Box contact rule (`plan:212-215`): "a HOST argument
must be given, because without one the script needs `HCLOUD_TOKEN` and
the `hcloud` binary (inventory surprise 9). An abort from either cause is
a BLOCKED report."

**What the tree says.** `scripts/deploy.sh:99-109` aborts only when
`HCLOUD_TOKEN` is empty **or** `hcloud` is absent. Otherwise it runs
`hcloud server ip musicbox` and sets `HOST` to the live box. Re-measured
on this workstation: `command -v hcloud` returns
`/home/linuxbrew/.linuxbrew/bin/hcloud`, and a name-only scan of
`~/.local/secrets` finds `HCLOUD_TOKEN` present, which `deploy.sh:93-96`
sources unconditionally before the resolution block. Neither abort
condition holds here. A `bash scripts/deploy.sh` typed without `--check`
and without an address applies the full deploy to the production box,
which is the one outcome this pass exists to prevent before the sitting.

The inventory states the dependency neutrally
(`research/2026-09-07-rung-4b-planning-inventory.md:112-113`,
`:1140-1146`). The plan added the abort framing, which turns a live
hazard into a reassurance.

**Fold.** The Box contact rule now says the opposite: on this
workstation a missing address silently resolves the production box, so
the address is mandatory and its absence is the hazard rather than a
stop. The rule is made mechanical rather than doctrinal by quoting the
full command string, `bash scripts/deploy.sh --check 192.0.2.1`, in
Tasks 3, 4 and 5's acceptance criteria and in every mutation that runs
it. Review 1's second option, a `deploy.sh` guard refusing to apply when
`HOST` came from a lookup rather than an argument, is recorded as R16's
declined alternative: it would be a seventh deploy edit the spec's scope
row does not name, and it changes behavior for every adopter. Marked
provisional, conductor ruling owed.

---

## MAJORS

### M1. Task 4's helper contract is self-contradictory: one render-and-install helper called twice cannot render both templates before the first remote step

**Lens.** systemd, deploy path and box safety.

**Plan lines.** `plan:1112-1115` specifies "One parameterized helper
takes a template path and a remote path, renders in memory, and
installs. It is called twice." `plan:1123-1125` then requires "Both
renders complete before the first remote step ... exiting 1 before any
ssh (Resolution R5)", and R5 (`plan:557-562`) says a render between two
remote steps "would leave a half-deployed box on a missing name".

**What the tree says.** Both cannot hold. An install is a remote step
(`deploy.sh:314-315`), so a combined helper invoked at two script
positions renders the second template after the first install. The
current script satisfies the property only because the render
(`:171-194`) and the install (`:291-316`) sit five announced steps apart,
with the first remote step at `:199-208`. The plan never says where the second
env step sits or whether the existing one at `:297` moves, while its
acceptance already asserts a specific announced count of sixteen.

**Fold.** Task 4's outcomes now split the contract in two. A render pass
runs both templates before `deploy.sh:196`'s first announcement,
accumulating one combined missing list. An install-and-check step runs
at each template's own position. `/etc/dubplate/env` keeps its position
at `:297` and the `server.env` step lands immediately after it. Marked
provisional, conductor ruling owed, as an amendment to R5.

### M2. Task 4's mutation proof 1 cannot fire, because `deploy.sh` re-sources `~/.local/secrets` and undoes the unset

**Lens.** systemd, deploy path and box safety.

**Plan line.** `plan:1169-1172`: "Delete one bare row's value from the
exported environment (unset it in a subshell) and run `--check
192.0.2.1` → the render exits 1 naming that row and its template."

**What the tree says.** `scripts/deploy.sh:93-96` runs `source
"$HOME/.local/secrets"` unconditionally before the render loop at
`:171-194`, so any name unset in the calling shell is restored before
`value="${!name:-}"` at `:182` reads it. The mutation produces a green
run, and an implementer reporting it as fired reports a false proof.

**Fold.** Replaced with a mutation the script cannot undo: add a bogus
bare row to the template under test, run `--check 192.0.2.1`, and the
render exits 1 naming that row and its template before any step prints.
It exercises the same `:190-194` abort honestly.

### M3. Task 4's mutation proof 3 states an expected result the store contradicts, and it contradicts mutation 2 in the same task

**Lens.** systemd, deploy path and box safety.

**Plan line.** `plan:1180-1183` expects that removing the literal branch
makes "every greenfield literal row ... reported missing, because none of
the nine names is exported."

**What the tree says.** Re-measured by name only against
`~/.local/secrets`: `DUBPLATE_STATE_DIR` is present, which is exactly
why mutation 2 (`plan:1173-1176`) predicts that a bare
`DUBPLATE_STATE_DIR` row silently renders `/var/lib/musicbox`. Both
cannot be true. The other eight names are absent
(`DUBPLATE_LIBRARY_ROOT`, `DUBPLATE_SITE_NAME`, and all six
`DUBPLATE_SHADOW_*`), so removing the literal branch reports eight names
missing and lets `DUBPLATE_STATE_DIR` resolve to the incumbent's value
through the indirect expansion at `deploy.sh:182`.

**Fold.** Mutation 3's expected result is now eight missing names plus
one silently inherited `DUBPLATE_STATE_DIR`, with the silent one named
as the point. As written it taught the implementer to treat a nine-name
failure list as the pass condition, masking the defect the literal
column exists to prevent.

### M4. Task 2 is declared independent, and Task 4's acceptance plus all three of its mutations require Task 2's file to exist

**Lens.** Both.

**Plan lines.** `plan:1027-1028` declares Task 2 "Independent: yes. It
writes only `env/`, which no other task in this pass touches", and
`plan:1610-1612` lets the conductor run it "at any point" after Task 0.
Task 4's acceptance (`plan:1162-1164`) requires `--check 192.0.2.1` to
print both env steps, and its three mutations (`plan:1173-1183`) all
operate on `env/dubplate-server.env.template`.

**What the tree says.** The render loop reads the template file directly
(`deploy.sh:175-194`) under `set -euo pipefail` (`deploy.sh:2`), so a
helper whose second call site names a missing template aborts every
`deploy.sh --check` and every apply. `scripts/check.sh` never invokes
`deploy.sh`, so the gate stays green over a broken deploy path, and the
pass exit gate is "`deploy.sh --check` clean" (`spec:1632`).

**Fold.** Task 4's header now reads "Depends on Task 3 and Task 2", Task
2's independence line reads "independent of Tasks 1, 3, 5 and 6; Task 4
depends on it", and the execution model's dependency list carries `Task
2 → Task 4` so pass-execute's parallel mode cannot order it after the
render refactor. Task 2 keeps its file-territory independence, which is
what the parallel mode's territory check reads.

### M5. The sitting checklist expects `would write 10 variables`, a string that only prints when the box is unreachable

**Lens.** systemd, deploy path and box safety.

**Plan line.** `plan:104-106` instructs the sitting to run `deploy.sh
--check 5.78.148.42` and expect "`would run:` lines, three `rsync -avzn`
diffs, and `would write 10 variables` for the new env step."

**What the tree says.** In `scripts/deploy.sh:298-309` that wording lives
only in the `else` arm of `if remote_reachable`, at `:308`. Against a
reachable box the env step prints `no changes (contents match...)` at
`:303` or `would change (contents differ...)` at `:305`. On the first
deploy `/etc/dubplate/server.env` does not exist, so `ssh cat` returns
empty and the step prints `would change`. The operator is told to look
for a string that cannot appear, at the pass's one attended gate.

**Fold.** The sitting's step 3 now expects `would change (contents
differ; diff not shown, it's secrets)` for the new `server.env` step, and
notes that `would write N variables` appears only when the box is
unreachable. Task 4's own acceptance is correct as written, because
`192.0.2.1` is unreachable by definition.

### M6. Tasks 3, 4 and 5 require the implementer to run ssh, which its own agent file forbids outright, and the plan calls that run "contacts nothing"

**Lens.** systemd, deploy path and box safety.

**Plan lines.** `.claude/agents/dubplate-implementer.md:35` reads "**The
box is production.** No `ssh`, no deploy, no write to `5.78.148.42`,
ever." `plan:198-199` reads that as forbidding "any ssh" and uses it to
keep `--check` against the box out of every task. Yet `plan:1075-1076`,
`:1162-1164` and `:1256-1257` make `bash scripts/deploy.sh --check
192.0.2.1` an acceptance criterion for three tasks.

**What the tree says.** That command issues ssh: `remote_reachable` at
`deploy.sh:119-121`, the `ssh cat` diffs at `:153` and `:301-302`, and
three `rsync -e ssh` calls at `:201`, `:214` and `:277`. `plan:205-206`'s
claim that it "contacts nothing" is also wrong. Each probe attempts a TCP
connect and blocks the full `ConnectTimeout=10` set at `deploy.sh:70`,
six such attempts run per `--check`, and the rsync steps print rsync's
own connection errors rather than the plan's promised `can't reach
192.0.2.1`.

**Fold.** Global Constraints now scope the implementer's ssh prohibition
to `5.78.148.42` and sanction ssh to the reserved documentation address,
quoting the exact command. The Box contact rule now says the command
attempts and fails connections to an unroutable address, costs roughly a
minute across six ten-second probes, and prints rsync connection errors
on three steps as expected output rather than a broken edit. Scoping an
agent file's hard rule from a plan is marked provisional, conductor
ruling owed, as R15.

### M7. Task 0 drops the spec's first staleness blocker while claiming to be the spec's subset, and a rung-1b planning inventory sits untracked in this tree

**Lens.** systemd, deploy path and box safety.

**Plan lines.** `spec:1685-1690` makes "Has rung 1b landed?" the first of
the three checks whose "no" stops the launch, because 1b changes the
wall's shape and re-expresses ruling 1's option (a) in its registry's
terms. `plan:715-722` presents its list as "4b's subset of the spec's
list plus this plan's own exposure" and its three blockers are the lane
registry, a clean green tree, and `.gitleaksignore`. The 1b question
appears in none of the eleven checks and no line records declining it.

**What the tree says.** The exposure is live. `git status` at HEAD shows
untracked `docs/research/2026-09-07-rung-1b-planning-inventory.md`
**and** `docs/plans/2026-09-07-rung-1b-config-registry.md`,
so rung 1b is being planned in parallel right now. Task 2's whole
ten-row template rests on `internal/config/keys.go:214-228` keeping its
present shape.

**Fold.** The spec's 1b question is Task 0's first blocker, ahead of the
lane registry, and check 8 stays as corroboration rather than substitute.
The spec's box-state question (`:1722-1725`, four enabled timers, serve
and nudge shipped and disabled) joins the list as check 12, since 4b is
the pass whose apply re-enables `TIMER_UNITS`.

### M8. The sitting's Apply step is a full-estate converge on a production box and the plan never names its blast radius

**Lens.** systemd, deploy path and box safety.

**Plan line.** `plan:107-109` reduces the apply to "`bash
scripts/deploy.sh 5.78.148.42`. The smoke test asserts a non-`dev`
matcher stamp."

**What the tree says.** That one command also runs `rsync -avz --delete`
of `compose.yaml` and `config/` into `/opt/dubplate`
(`deploy.sh:207-208`), rsyncs every unit, installs the tmpfiles fragment
and runs `systemd-tmpfiles --create`, rsyncs the binary and chmods it,
re-renders and reinstalls the incumbent `/etc/dubplate/env` (`:314-315`),
installs `contributors.yaml`, runs four directory steps, runs `docker
compose ... up -d` (`:392-394`), re-enables the four guardian timers
(`:406-408`), and reinstalls beets via uv as the music user
(`:410-441`). `~/.claude/skills/vps-conventions/SKILL.md:90` records that
a compose step after a network change reattaches live containers without
their service aliases and silently killed Navidrome's container-name DNS
on 2026-09-01. The Territory wall at `plan:290-294` is true of the repo
and says nothing about what the apply pushes from those files' current
state.

**Fold.** The sitting checklist now names the blast radius: step 3 lists
the three diffs an operator must actually read (the `/opt/dubplate`
`rsync --delete` diff, the incumbent env step's verdict, and the compose
step), and a new step 5 runs `docker compose ps` plus a Navidrome
reachability check after the apply and before the timer enable, on the
`SKILL.md:90` precedent. The later steps renumber. The added
verification is one owner minute and it is batched into the sitting Geoff
already runs.

---

## MINORS

**m1. Task 6's acceptance grep is case-sensitive and cannot see the two
capital-P marks the task retires.** `plan:1322-1324` sets the criterion
"`grep -rn provisional internal/ cmd/` returns exactly one line". Two of
the five marks read "Provisional, owner ruling 4 open" and "Provisional,
owner ruling 5 open" with a capital P, so the lowercase grep returns four
lines today rather than six and would report success with both untouched.
Re-measured: `grep -rn` returns four, `grep -rni` returns six. R2 reaches
the right six by inspection, so this is the criterion drifting from the
resolution. **Folded:** the criterion is `grep -rni`.

**m2. R1 records only the parent-versus-leaf alternative and never
considers `StateDirectory=`.** `plan:478-483` weighs one alternative.
`~/.claude/skills/vps-conventions/SKILL.md:18` names `StateDirectory=` /
`/var/lib/<x>` as the persistent home in its lifetimes table, and the
`:39-42` pre-exist rule is precisely the constraint
`StateDirectory=dubplate/shadow` removes, since systemd creates both
levels owned by `User=` and grants them write implicitly. Task 7 is
handed that skill in full and asks the minimality question first, so a
finding on an unconsidered third option would arrive after both the unit
and the pre-create step had landed. **Folded:** R1 records
`StateDirectory=` as a declined third option, with its reason.
`DUBPLATE_STATE_DIR` is config-driven and has no default
(`internal/config/config.go:217`), `/var/lib/dubplate` becomes shared
across rungs 5, 6 and 12 where per-unit `StateDirectory` ownership stops
being the right model, and all four existing directory grants use
pre-create plus `ReadWritePaths` (`deploy.sh:336-338`, `:357-359`,
`:370-372`, `:387-389`).

**m3. Eight citations do not resolve to the lines they name, in a plan
whose Task 1 criterion requires every citation to resolve.** Merged from
both lenses and re-measured by the fold editor.

| Plan line | Cites | Measured |
| --- | --- | --- |
| `:81` | `spec:342-368` for ruling 6 | heading `:345`, ends `:371` |
| `:357` | `spec:1816-1818` for the mutation-proof rule | `:1831-1832`; `:1816` is the review-scaling sentence |
| `:373` | `PA:179` for the state directory row | `PA:178` at HEAD |
| `:375` | `PA:181` for `internal/engine/shadow` | `PA:180` at HEAD |
| `:388` | 2.0 hooks at `:354-388` | `:354-391`; hook 14 closes at `:391` |
| `:409` | the prior-art table at `:391-421` | heading `:393`, table `:401-423`, declines `:425-428` |
| `:433` | `store.go:100` for the atomicfile write | `:102`; `:14` is right |
| `:230` | `docs/HISTORY.md:466-481` for the budget paragraph | `:468-480` |
| `:500`, `:1317` | `docs/HISTORY.md:437-441` for forward carry 5 | `:442-447` |

**Folded:** all corrected in place, with the `PA:` rows additionally
named in prose because that file carries an uncommitted fifteen-line
addition and its numbers are moving.

**m4. The prior-art count is off by one against the plan's own table.**
`plan:411-412` states "twenty-one of the twenty-three entries are not
applicable". The table at `:416-439` carries twenty-three rows, of which
exactly one, `internal/atomicfile`, is consulted and adopted. Twenty-two
are not applicable. **Folded:** twenty-one becomes twenty-two.

**m5. The flat-namespace hazard is attributed to the wrong mechanism in
two places.** `plan:1156-1158` says `DUBPLATE_STATE_DIR` "is exported as
`/var/lib/musicbox` by the incumbent's own render in the same shell", and
`plan:954-956` says "the incumbent's own template exports
`/var/lib/musicbox` into the same flat shell namespace". The render loop
at `deploy.sh:177-188` exports nothing; it appends to the `env_lines`
array. The value reaches the shell from `source "$HOME/.local/secrets"`
at `:93-96`, which a name-only scan confirms carries
`DUBPLATE_STATE_DIR`. The hazard is real and the literal column is the
right answer, but an implementer hunting for the export in the render
loop will not find one. **Folded:** both passages now name the source
file and the indirect expansion at `:182`.

**m6. R2 reads `docs/STATUS.md:47-48` as counting six provisional marks
when it counts six ratified rulings.** Those lines read "all six are
ratified", and a mark is not ratified. `STATUS.md:11-12` calls the six
rulings six. The independent grep is what establishes six occurrences,
and it does. **Folded:** the STATUS citation is recast as "STATUS records
all six rulings ratified and the retirement mechanical", and the count
rests on the grep alone.

**m7. The 3M ceiling leaves 4% headroom and half the fix-round allowance
its own basis pass consumed.** The table at `plan:247-256` sums to 2.88M
against a 3M ceiling, with 0.40M of plan prep already spent, so the
executing half has 2.48M against a 2.48M projection. The basis pass spent
0.49M on two fix-and-review rounds (`docs/HISTORY.md:470-471`); this plan
allows one at 0.26M. **Folded:** the ceiling stays 3M and the table now
states that a second fix verdict on any task crosses it and is the
conductor's decision at the 2.4M flag, so the breach is a planned branch
rather than a surprise. Raising the ceiling is the recorded alternative,
marked provisional, conductor ruling owed.

**m8. Task 1's timer criterion is literally unsatisfiable as written.**
`plan:875-876` requires "The timer's four `[Timer]` directives match
`dubplate-navidrome-check.timer` exactly", while the outcome at `:810`
asks for an explicit `Unit=`.
`systemd/dubplate-navidrome-check.timer:11` is
`Unit=dubplate-navidrome-check.service`, and the shadow timer must name
`dubplate-shadow.service`, so the fourth directive cannot match. A
diff-reviewer reading the criterion literally returns `fix` on a correct
file. **Folded:** `OnCalendar`, `RandomizedDelaySec` and `Persistent`
match `:8-10` exactly, and `Unit=` names this pass's own service.

**m9. HISTORY handoff item 4 lands in no named task.** `plan:46` states
the six numbered handoff items are the plan's inputs. Items 1, 2, 3, 5
and 6 land at `:463`, `:1202`, `:927-938`, `:838` and `:1548`. Item 4
(`docs/HISTORY.md:411-415`: the doubled wire pace, the sixty-token cap,
and "sixty tokens at two seconds a token bounds a pass at two minutes,
comfortably inside 4b's five-minute timer") appears nowhere; the timer
comment's required content at `:857-863` covers only the quiescence
reason and the `Persistent=true` reason. **Folded:** the pass-duration
bound joins the timer comment's required content and Task 1's criteria.
It is the sentence that says why a five-minute interval cannot overlap
its own predecessor.

**m10. R14 describes nine drafted items where Task 8 lists and pins
eight.** `plan:697-699` says Task 8 drafts "all five corrections plus the
R1 grant correction and the three unnamed deliverables", while Task 8's
list at `:1450-1478` has eight entries and its criterion at `:1500` pins
"The eight corrections". **Folded:** R14 reads "eight corrections, one
line each", matching the list and the criterion.

**m11. The staleness pre-task does not check the one precondition three
tasks depend on.** `plan:206-215` names an incomplete `~/.local/secrets`
as an abort producing a BLOCKED report, and Tasks 3, 4 and 5 each prove
their edits with `--check 192.0.2.1`. Task 0's checks verify two names by
name only. A render abort would surface as three separate BLOCKED reports
after three dispatches. The precondition currently holds: a simulation of
the loop over `env/dubplate.env.template` resolved all 42 rows with none
missing, so this is insurance rather than a live blocker. **Folded:** one
Task 0 check, reported by count and by the names of any misses, values
never printed.

**m12. The header says the pass never reaches the box while R11 has the
conductor reach it.** `plan:24` ("This pass touches the box exactly once,
and not in this window") and `:76-77` ("Nothing here reaches the box,
writes to the box, or enables anything on the box") against `:160-161`
and R11 (`:648-659`), which have the conductor run `deploy.sh --check
5.78.148.42` in this window. An implementer or reviewer reading the
header as binding could refuse the conductor's own step or flag it as a
violation. **Folded:** the header says "writes to the box exactly once"
and "Nothing here writes to the box or enables anything on it", each
pointing at the Box contact rule for the one read-only exception.

**m13. Task 2's masthead criterion cannot be checked by a diff-reviewer
from the diff.** `plan:1000-1002` requires that `DUBPLATE_SITE_NAME`'s
value be copied from the store and appear nowhere in the report or the
commit message, which R7 restates. The reviewer sees the literal in the
diff and has no stated way to compare it against the store's
`DUBPLATE_SERVE_SITE_NAME` without printing it, and R7 says a guessed
masthead would ship to the box and appear on every page rung 7 builds.
**Folded:** the criterion is now a non-printing verification. The
implementer reports the exit status of a comparison between the
template's rendered value and the store's `DUBPLATE_SERVE_SITE_NAME`, and
the reviewer re-runs the same comparison. The status is the evidence and
the value stays unprinted.

---

## Verified sound

The do-not-re-litigate record. Both lenses' lists, merged and
deduplicated. Forty-two items.

**The units and the gate.**

1. The two proposed unit files pass the gate's systemd leg, tested
   empirically rather than by reading. Review 1 wrote the plan's exact
   directive set (`plan:791-811`) into a scratch copy of `systemd/` and
   ran `check.sh:137-139`'s command verbatim. The only new line was
   `dubplate-shadow.service: Command /usr/local/bin/dubplate-server is
   not executable: No such file or directory`, which the filter at
   `check.sh:139` removes because its `[A-Za-z0-9_.-]+` class covers the
   hyphen. Post-filter output was empty.
2. Task 1's acceptance holds with no new gate wiring, and a service with
   no `[Install]` section plus a timer with an explicit `Unit=` both
   verify clean.
3. The systemd gate picks up both new units by glob
   (`scripts/check.sh:137` over `systemd/*.service systemd/*.timer
   systemd/*.path`). Every existing unit's `ExecStart` carries arguments
   and the gate is green, so the filter is proven against the argument
   form `dubplate-server` will use.
4. The hardening set is a faithful copy of the estate's template.
   `dubplate-serve.service:99-115` carries exactly the directives the
   plan lists, and `:125` is the `RestrictAddressFamilies` line.
5. The `AF_NETLINK` exclusion is right. The serve unit's comment
   (`:116-124`) records that netlink joined the set for a CPython
   subprocess's glibc `getaddrinfo` and that "The Go process itself never
   needed it". The shadow runs no subprocess, and `Makefile:7` builds
   `CGO_ENABLED=0` so the pure Go resolver applies.
6. `ProtectHome=yes` does not reach `/srv/music`, which is the music
   user's `HOME` per `deploy.sh:439` but is not one of the three paths
   `ProtectHome` covers.
7. The timer matches `dubplate-navidrome-check.timer` directive for
   directive on the three that can match: `OnCalendar=*:0/5` at `:8`,
   `RandomizedDelaySec=60` at `:9`, `Persistent=true` at `:10`, with
   `[Install] WantedBy=timers.target` at `:13-14`.
8. The `OnFailure` chain is wired and genuinely inert.
   `dubplate-ping-fail@.service:11` reads `/etc/dubplate/env` and `:20`
   execs the frozen `/usr/local/bin/dubplate`.
   `_salvage/cmd/dubplate/pingfail.go:34-43` confirms an unset
   `DUBPLATE_PING_<SLUG>_URL` logs and returns nil rather than erroring,
   and `cobra.ExactArgs(1)` accepts any slug, so an unminted `shadow`
   slug cannot fail the helper. The `DUBPLATE_PING_*` class is reserved
   for rung 9 (`internal/config/keys.go:116-125`), so the plan's "no
   monitoring at rung 4" conclusion is honest.

**The grant.**

9. The leaf `ReadWritePaths` grant is the correct minimum and is argued
   statically. Every filesystem write reachable from `shadow observe` is
   at `store.go:55`, `:102`, `:114`, `:121`, `:175`, `:179`, `:208`,
   `:211`, `:228`, `:244`, `:304` and `:359`, all rooted at the Store's
   own root. `cmd/dubplate-server/shadow.go` contains no `os.Create`,
   `os.WriteFile`, `os.MkdirAll`, `os.Rename` or temp-file call at all.
10. `atomicfile.WriteFile` (`internal/atomicfile/atomicfile.go:15-42`)
    creates its temp in the target's own directory, so it stays inside
    the grant.
11. `StoreDir` is `filepath.Join(stateDir, "shadow")`
    (`store.go:46-48`), and `TestStoreWritesNothingOutsideItsRoot`
    (`store_test.go:191`) walks the parent tmp dir, so it proves the leaf
    rather than the parent, exactly as R1 claims.
12. R1's reconciliation of the spec against the handoff is correct and
    necessary. `docs/HISTORY.md:389-394` names `<state dir>/shadow/` as
    the `ReadWritePaths` argument while `spec:1318-1322` pre-creates
    `/var/lib/dubplate`, and `SKILL.md:39-42` does make those mutually
    exclusive as written. Making both levels in one `run_remote` and
    granting the leaf satisfies both; `OpenStore`'s `MkdirAll` at
    `store.go:55` then creates the four subdirectories inside the granted
    root on the first pass.
13. The pre-create's placement is safe. Task 3 puts the new step beside
    the three existing ones at `deploy.sh:336-389`, after the units rsync
    at `:212` and before `systemctl daemon-reload` at `:408`.
    `daemon-reload` loads a unit without starting it, and the shadow
    timer is deliberately absent from `TIMER_UNITS`, so the unit can
    never start before its `ReadWritePaths` leaf exists.

**The deploy path.**

14. The step-count arithmetic is right. `grep -n 'echo "== '` returns
    fourteen announcements at HEAD (`:199`, `:212`, `:230`, `:273`,
    `:288`, `:297`, `:327`, `:337`, `:358`, `:371`, `:388`, `:393`,
    `:406`, `:411`), so Task 3's fifteen and Task 4's sixteen are
    correct, and the three `rsync -avzn` diffs the sitting expects are at
    `:201`, `:214` and `:277`.
15. `deploy.sh --check` makes no box write. The three rsyncs carry `-n`,
    `run_remote` issues no ssh under check (`:129-137`), `deploy_file`
    and the env step read through `ssh cat` (`:153`, `:301-302`), and the
    trigger-directory probe (`:239-249`) is a `test -d`. The only local
    write is the gitignored cross-compiled binary at `:274`.
16. Check-mode rsyncs are guarded with `|| true` (`:200-202`, `:214-215`,
    `:219`), so `--check` against an unreachable address runs to
    completion rather than aborting on the first step.
17. No task in the unattended half writes the box. Tasks 3, 4 and 5
    target `192.0.2.1` only, and R11 reserves the single `--check
    5.78.148.42` to the conductor.
18. R4's premise holds. `run_remote` prints `would run: <command>` under
    `--check` and issues the ssh otherwise (`:129-137`), which is what
    stops a step's check line from drifting from what apply runs.
19. R6's `0600 root:root` reasoning is right.
    `dubplate-import.service:15` carries
    `EnvironmentFile=/etc/dubplate/env` against `deploy.sh:314-315`'s
    `install -D -m 600 -o root -g root`, and that unit runs `User=music`
    at `:14`, so the estate already proves systemd reads
    `EnvironmentFile=` as PID 1 before dropping privileges.
20. The quiet-box citations are correct: `dubplate-import.timer:10-11` is
    `OnCalendar=*:0/15` with `RandomizedDelaySec=300`, and
    `dubplate-backup.timer:9-10` is the 11:30 UTC schedule with 600
    seconds of jitter.

**The template.**

21. The template row arithmetic is right. `env/dubplate.env.template`
    carries exactly 42 assignment rows, 39 `DUBPLATE_*` and three `ND_*`.
    No tracked line carries a right-hand value, so the literal column is
    backward compatible as Task 4 claims.
22. The greenfield template's ten rows are the correct set.
    `config.go:216-219` requires `DUBPLATE_LIBRARY_ROOT`,
    `DUBPLATE_STATE_DIR`, `DUBPLATE_SITE_NAME` and
    `DUBPLATE_ADMIN_EMAIL`; `keys.go:214-228` registers the shadow class
    with three required-when-enabled paths and two optional thresholds
    behind `DUBPLATE_SHADOW_ENABLED`.
23. `keys.go:165-166` confirms `DUBPLATE_LIBRARY_DIR` and
    `DUBPLATE_SERVE_SITE_NAME` are superseded, so those two must ship as
    literals. `config.go:190-192` is exactly the "set without the enable
    flag" error the plan cites, and `config_test.go:306-309` confirms an
    explicit false loads clean.
24. `DUBPLATE_ADMIN_EMAIL` is the only name present in `~/.local/secrets`
    under its greenfield spelling, so it is correctly the single bare
    row.
25. R8's literals match the registered defaults exactly:
    `cmd/dubplate-server/shadow.go:26` sets Default `"720h"` and `:40`
    sets Default `"60"`, and `thresholds.go:55-95`'s `ParseInt` and
    `ParseDuration` both fall back to Default when the key is unset.
26. Task 2's gitleaks mutation fires: `scripts/check.sh:107` runs
    `gitleaks detect --no-git -s .` over the directory, so the new
    template is scanned before it is staged.

**The stamps.**

27. The refusal the pass exists to make observable is where the plan says
    it is. `main.go:31` and `:44` declare `version` and `matcherCommit`
    as package-level vars both defaulting to `"dev"`, `Makefile:6-7`'s
    `build-linux` carries no `-ldflags`, and the refusals fire at
    `shadow.go:203-205` and `:303-305`.
28. R3's premise holds. `newVersionCmd` is `main.go:72-81`, printing one
    line at `:77`, and `main_test.go` carries no version test. A
    repo-wide grep finds no other reader of that stdout. `doctor.go:76`
    is the pinned `%-32s %s` form the two rows would adopt, and
    `printRow` is in the same package.
29. R9's consequence is real and precisely located.
    `cmd/dubplate-server/shadow.go:215` builds the MusicBrainz
    User-Agent as `fmt.Sprintf("dubplate-server/%s ( %s )", version,
    cfg.AdminEmail)`, so stamping `main.version` changes what the box
    sends on every request, and a dirty worktree ships a `-dirty`
    identity.

**The provisional marks.**

30. The marks are handled correctly against the tree. `grep -rni
    provisional internal/ cmd/` returns exactly six occurrences:
    `internal/engine/shadow/{rescore.go:28, match.go:20, run.go:388}`,
    `cmd/dubplate-server/shadow.go:22` and `:37` (all five naming
    ratified rulings), and `internal/engine/shadow/flac.go:62`, which
    names no ruling number and is the rung 8 handoff the plan correctly
    leaves untouched.
31. Task 6's mutation proof fires: `.vale.ini` sets `BasedOnStyles =
    glw907` over `*.go` with the em dash banned and no opt-out,
    `scripts/vale-comments.sh:59` runs at `--minAlertLevel=error`, and
    its file list unions untracked Go files so a fresh comment is
    scanned.

**Process and plan integrity.**

32. The token ceiling arithmetic checks out against its stated basis.
    `docs/HISTORY.md:468-480` records 4a-ii at about 2.7M over five tasks
    with a 1.32M chain across 16 agents, giving both the 0.26M chain rate
    and the 0.54M all-in rate the plan quotes, and the table sums to
    2.88M as printed. The refusal to transfer the all-in rate is argued
    from the task profile rather than asserted.
33. The sitting boundary is unambiguous. Tasks 0 through 8 unattended,
    Task 8 named as the boundary task that leaves the pass OPEN and
    writes no HISTORY entry and no deploy date, Task 9 attended in a
    later session.
34. The 2.0 neutrality finding is sound. All fourteen hooks
    (`streamer-2-0-design.md:359-391`) name rung 6 schema work, rungs 7
    through 9, or rung 12; none names systemd, a unit, a deploy step, a
    build stamp, or an environment file. A grep of that spec for
    `systemd`, `deploy`, `/var/lib`, `ldflags`, `stamp` and `state
    directory` returns only `:319` and `:325`, neither of them a hook.
    Every task carries its "2.0 hook: neutral" line.
35. The prior-art statement exists and gives every entry a verdict with a
    reason, plus the one positive adoption whose in-repo use is real.
36. Every inventory surprise is carried into a resolution or a task:
    1→R1, 2/3/4/7/13→R14 and Task 8's correction list, 5→R2, 6→R13 and
    Task 5, 8→Task 1's fifth comment, 9→the Box contact rule, 10→Task 2's
    comment set, 11→R12, 12→R9. Where the tree contradicts the spec the
    plan follows the tree, and no task edits the spec.
37. The independence marks are true as file territory. Task 2 writes only
    `env/`, and Task 6's five files are disjoint from Task 5's
    `main.go` and `main_test.go`. The plan names the
    one-executor-per-tree caveat rather than implying two live lanes.
38. Handoff items 1, 2, 3, 5 and 6 each land in a named task.
39. The mutation proofs name what fails, and the three that deliberately
    do not fire are declared as findings with what holds the property
    instead.
40. The `.gitleaksignore` rider is accurate: the file's single line is
    `internal/config/keys_test.go:generic-api-key:66`, and no task in
    this pass writes that file. The implementer agent file exists at
    `.claude/agents/dubplate-implementer.md` and carries the box
    prohibition the execution model relies on.
41. The plan contains no implementation code and no fenced code block.
    Unit directives and template row values are specification carried
    from the spec, not code.
42. Writing-voice mechanics hold: no em dash anywhere in the file, no
    "not X but Y" frame, and the "rather than" density (26.2 per thousand
    lines) sits below both the 4a exemplar plan (28.4) and the rung 4
    spec (36.8).
