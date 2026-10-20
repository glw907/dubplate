# Rung 1b (The Config Wall-Row Registry) Implementation Plan

**Revision 2, 2026-09-07. Band S**, per the parallel-approach spec's Phase
0b paragraph (`PA:56`), with its own plan and the standard
implementer/`diff-reviewer`/gate chain. Drafted against HEAD `c98a916`, after rung 4a
closed and before pass 4b runs. **This plan changes rung 1's ratified
config-core contract**, which is why 1b is a rung with a plan rather than
an unplanned chore (`PA:56`). The contract change is stated in its own
section below, and Task 7 records it as a correction to the build spec
rather than leaving the spec describing a wall that no longer exists.

**Revision 2 folds two adversarial reviews**
(`research/2026-09-07-rung-1b-plan-review.md`), which found no blocker,
twelve majors and fourteen minors. The load-bearing corrections, in the
order a reader meets them. Three acceptance criteria could not be
satisfied as written: Task 4's shadow grep contradicted R5's
survives-verbatim constraint, so the test's home is now decided here;
Task 6's template classification walked three branches against a
greenfield template that will carry six integration keys, so the fourth
branch is now named; and R11's kind check could not catch its own stated
hazard, because `KindInt` is the zero value, so registration now parses
each declaration's `Default` under its declared `Kind`. Three mutation
proofs could not fire and are re-aimed: Task 3's order mutation against a
test that derives its expectations from the accessor it would sort,
Task 3's double-construction mutation, and Task 6's template mutation
that aborts before the guard it claims to prove. Task 6's exhaustiveness
outcome is already landed at `doctor_test.go:127-129` and is rescoped to
the scratch-registry direction. Task 1's gitleaks rename now covers both
tables in `keys_test.go`, measured. O2 is softened and the rule for a
rung retiring its own reserved class is stated. Two counts are corrected
from the tree, 29 reserved keys and 39 template names, and two spec
anchors are re-pointed. The spec-correction list grows from five to
eight, the collision-class list is unified at six, a Global Constraint
now requires the inline resolution markers Task 7 must list, and the
ceiling's cross-check arithmetic is corrected without moving the
ceiling.

**Conductor rulings, 2026-09-07.** All sixteen resolutions, R1 through
R16, are ruled as applied: the fold's provisional application was correct
and the conductor accepts each as recommended, with no reversal. Task 5
grows by one kind: `KindString`, alongside `KindFloat`, for the two
string-valued theme keys the Phase 0c design-language standard needs.

**Revision 2.1, 2026-09-08.** Task 0 ran against HEAD `2204947`, after
pass 4b closed, and reported no blocker. This revision folds its
staleness findings as a re-pointing edit; no task's outcomes change
except where a finding forced one. Every citation this plan carries
against the pre-4b tree is corrected: the parallel spec's Phase 0b
paragraph and contested-surface rows (`PA:56`, `:175`, `:194`, `:199`,
`:200`, `:201`, the last now also naming `Makefile`), the rung 4 spec's
four anchors (each shifted two lines by an earlier 4b edit),
`shadow.go`'s and `main.go`'s registration lines, `HISTORY`'s carries 3
and 4, `tunables.go`'s four-knob paragraph, and `STATUS`'s spend figure.
The measured gate timings replace the "about two minutes" and "about
three and a half" estimates: a cold run is about five minutes ten
seconds, a warm one about seventy-seven seconds. The Global Constraints'
inline-marker exemplar is reworded, because 4b's Task 6 ratified the two
markers it cited. Task 6's completeness direction becomes a mandatory
assertion, since `env/dubplate-server.env.template` landed (`c8b6d71`)
carrying all four owned-now rows; its classification helper generalizes
to both templates; its doc-comment rule names both templates rendering
through `deploy.sh`; and its duplicated mutation number is fixed, re-aimed
at `DUBPLATE_LIBRARY_ROOT` rather than `DUBPLATE_STATE_DIR`, which 4b's
own state-dir gate leg now also reads. Task 7 records 0c as landed fact,
committed at `b77e159`, and folds its own declared coverage gaps into
4b's running total of thirty-four rather than re-counting from rung 4a's
nine.

> **For agentic workers:** workflow mode per CLAUDE.md's "Build process"
> section, the pass-execute chain per task, strictly sequential. The
> conductor consumes reports and never reads a diff, a source file, or a
> gate transcript inline. Subagents never dispatch agents;
> `code-simplifier` is the conductor's own dispatch after a task's
> review accepts. The conductor owns `docs/STATUS.md`, `docs/HISTORY.md`,
> `ROADMAP.md`, the lane registry, the contested-surface rows, and the
> batch commit of anything a task drafts into them. **This pass runs in
> the MAIN TREE.** Rung 1b is not one of the three full adversarial-cycle
> rungs (5, 12, 16), so review scales to the in-chain `diff-reviewer` per
> task plus the pass-end gate.
>
> **Execution mode:** workflow mode, the `~/.claude/workflows/pass-execute.js`
> chain per task, sequential; the implementer is `cairn-implementer`
> instructed to read and follow `.claude/agents/dubplate-implementer.md`;
> the reviewer is `diff-reviewer` on `claude-opus-5`; the gate is
> `bash scripts/check.sh`. Ceiling 2.5M, flag at 2.0M routed to the
> conductor, checkpoint every four tasks.

## Primary inputs

Read these before Task 1, in this order.

1. **`docs/research/2026-09-07-rung-1b-planning-inventory.md`,
   in full.** It is this plan's argument, read off the tree at `c98a916`.
   Its eight sections and its fourteen surprises are binding. **Do not
   re-derive what it settles; cite it, and verify the line if you carry
   one forward.** Where it contradicts a spec, the tree wins and Task 7
   drafts the correction.
2. **`docs/specs/2026-09-03-parallel-build-approach.md`**,
   the Phase 0b paragraph (`:56`) and the contested-surface table, whose
   config-wall row is this pass's mandate. **Every `PA:` line number in
   this plan is re-pointed to revision 2.1's tree**, HEAD `2204947`,
   where Task 0 measured the Phase 0b paragraph at `:56` and the
   contested-surface rows at `:175` (config wall), `:194` (docs/
   ledgers), `:199` (`cmd/dubplate-server`), `:200` (`internal/config`)
   and `:201` (`systemd`/`deploy.sh`/`env`/`Makefile`). These sit
   seventeen lines below where committed HEAD `c98a916` had them,
   because 4b's Phase 0c completion note (fifteen lines, now committed
   at `b77e159`) landed after the Phase 0b paragraph, whose own
   position shifted two lines from earlier 4b edits. **Task 0 item 13
   is where this was measured.**
3. **`docs/specs/2026-09-02-greenfield-server-build.md`**,
   "The config wall, derived" (`:148-199`) and "The Go bar" (`:200-207`).
4. **`docs/HISTORY.md`'s 4a-ii forward carries** (`:429-468`), especially
   carries 3 and 4, and rung 3's statements of the same carries
   (`:1059-1075`, `:1425-1480`). Carry 3 is phrased as one item and is
   two; Resolution R6 splits it.
5. **`CLAUDE.md`** at the repo root: Build process, Gate, Stack rules,
   Facts agents get wrong, Safety rails.
6. **`docs/specs/2026-09-07-streamer-2-0-design.md`**, the
   "Forward hooks the current build owes" section (`:354` onward). Hook 6
   is binding on this pass's shape; hooks 5, 12 and 14 are advisory.
7. The tree, read-only: `internal/config/{config,keys,thresholds}.go` and
   their three test files; `cmd/dubplate-server/{main,check,doctor,shadow}.go`
   and their tests; `env/dubplate.env.template`; `scripts/check.sh`;
   `.gitleaksignore`.
8. **`~/.claude/skills/go-conventions/SKILL.md` before any Go.** No
   Python, no unit file and no web surface in this pass, so
   `python-conventions`, `vps-conventions`, `svelte-conventions` and
   `ts-conventions` are out of scope.

## Goal

Turn the config wall from four closed tables inside `internal/config`
into one registry that every rung declares into from its own file, so no
lane edits `internal/config/keys.go` again, a duplicate declaration fails
loudly rather than silently appending, and the streamer driver of rung 12
can own its key set rather than inherit three hard-coded Navidrome names.

Nothing here touches the box, deploys anything, reads the network, writes
`corpus/`, adds a module dependency, or adds a name to any environment
template.

## Owner touchpoints: none

1b expects no attended sitting and raises no question. Every decision it
takes is a code-shape decision inside one package, and every one is
applied provisionally with a conductor ruling owed rather than an owner
ruling. **A discovered question is a finding to report, never a mid-pass
ask.** The 80%-of-ceiling flag routes to the conductor, per the owner
grants for this window.

**What would become an owner touchpoint**, so the conductor recognizes
one if a task reaches it:

- **Adding any name to `env/dubplate.env.template` or to a greenfield
  template.** `scripts/deploy.sh:171-194` reads a template as a manifest
  of names and exits nonzero on any name it cannot resolve, so a new name
  is a minted secret (`secret-set.sh`, the age registry). Resolution R10
  keeps 1b out of this entirely.
- **Adding, removing, or renaming an owned-now wall row.** That is a new
  required key for every adopter, a product decision rather than a
  registry one. It is why `AdminEmail`'s dedicated contact row is not
  1b's (Resolution R7).
- **Declaring a reserved class for a capability whose rung has not been
  planned**, such as the build spec's SSO proxy capability, trusted-proxy
  mode, or forward-auth front (Resolution R13).
- **A re-derived ceiling past 2.5M**, which the conductor takes to Geoff
  only if it also breaks a window cap.

## Spec

`docs/specs/2026-09-03-parallel-build-approach.md`, Phase 0b
(`:56`), quoted in full at inventory `:6-21`, with the ownership row at
`:175`. Behind it,
`docs/specs/2026-09-02-greenfield-server-build.md` (revision
2.1, ratified), "The config wall, derived" (`:148-199`), which is the
classification rung 1 implemented and rung 1b re-homes without changing.

**The forward criterion this pass hands on**, so a later rung reads it
rather than rediscovers it:

> After 1b, a rung that needs a configuration key declares it in its own
> package's own file and registers it. `internal/config/keys.go` carries
> rung 1's own declarations plus the reserved classes whose owning
> packages do not exist yet. A rung that finds itself editing another
> rung's declarations has found a contested surface, not a chore.
>
> **The one edit a lane may make to `keys.go` after 1b** is deleting its
> own reserved class, in the same commit that declares that class in its
> own package. No lane adds to `keys.go`, and a lane removes only the
> class it is taking over. The move is pinned by a registration-time
> check that the relocated class registers under the same name and the
> same key set, so a move cannot silently drop a key.

## The contract change

Rung 1b changes a ratified contract, so the change is stated here in full
rather than discovered in a diff. Task 7 records it as a spec correction.

### What rung 1's contract says today

`internal/config/config.go:1-45`, the package doc, is the ratified
config-core contract. It states two things:

1. **Every `DUBPLATE_*` name falls into exactly one of five classes**,
   and `Load` enforces the classification rather than merely reading
   values: owned-required-now, reserved-for-a-rung, adopter-optional
   integration, superseded, and unknown. Each class has its own error
   wording, and unknown refuses the start rather than degrading it
   (evolution lesson U5).
2. **`[Thresholds]` is a separate, open-ended registry** for the behavior
   tunables rungs introduce as they land, distinct from the wall, which
   is a closed table.

The build spec's own sentence behind it (`BS:182-183`): "Rung 1 derives
the complete table from this classification and it becomes the config
core's contract."

### What 1b changes

**The wall stops being a closed table and becomes registry contents.**
The four tables `ownedWall`, `reservedClasses`, `integrationClasses` and
`supersededKeys` stop being package-private literals that only
`internal/config` may edit, and become declarations a package registers.
The "separate, open-ended registry" sentence retires: after 1b there is
one registry, and thresholds are one of the things declared into it.

Three consequences follow, and each is a real change rather than a
restatement.

- **`Register` stops appending blindly.** A duplicate or malformed
  declaration is refused loudly at registration (Resolution R3), on the
  precedent of `flag.(*FlagSet).Var`, `database/sql.Register` and
  `expvar.Publish`, all three of which panic. `image.RegisterFormat` is
  the one stdlib registry shaped like today's `Register`, and its licence
  to be lax comes from a semantics config keys do not share: a duplicate
  format sniffer is dead weight, while two rungs claiming
  `DUBPLATE_STATE_DIR` is the exact collision 1b exists to prevent.
- **`Load` takes the registry it enforces against** (Resolution R1),
  rather than reading four package-level tables. Its other input,
  `environ`, is already a parameter; the registry joins it.
- **A class receives its own resolved values by name.** `Load`'s
  `class.Name == "shadow"` branch (`config.go:235-241`) is the one place
  `Load` is not data-driven, and a second integration class would need a
  second branch. Resolution R5 retires it.

### What stays, exactly

Everything below is a property the migration must preserve, and every one
of them is an acceptance criterion on at least one task.

- **The five classes and their error wording.** A reserved name still
  errors naming its class and rung. A superseded name still errors naming
  its replacement or its retirement. An unknown name still refuses the
  start. An integration key set without its flag still errors naming the
  flag, and the flag still skips its own self-report so one typo produces
  one error.
- **`Load`'s report-everything semantics.** Every problem at once through
  `errors.Join`, never a stop at the first (`config.go:246-249`).
- **The `wall.get` choke point** (`config.go:115-121`). It owns
  last-wins, empty-equals-unset, and the `=`-in-value decision, and every
  accessor and both threshold parsers route through it. The packaging
  initiative's later file overlay is a one-function change only while
  that stays true (inventory `:2.4`).
- **The `Required`/`Optional` split's semantics** (`keys.go:199-202`).
  `Load` requires every `Required` key when the flag is set; an
  `Optional` key is accepted and never required, because its registered
  default stands when unset. This is HISTORY carry 4's explicit demand.
- **The identity-source pending row.** It is a table entry whose `Name`
  is prose rather than an env key, `ownedNow` returns false for it, and
  doctor prints its `Pending` text where other rows print a status. The
  prefix rule of Resolution R16 must not reach it.
- **Table order, everywhere doctor prints it.** Reserved classes print in
  table order and that order is not rung order (`decision and review
  queues`, rung 13, precedes `streamer`, rung 12). A registry that sorted
  on rung or on name would change user-visible output and break
  `TestDoctorCmd`.
- **Doctor's four blocks, their order, and the `%-32s %s` column pin**
  (`doctor.go:74-78`). It is a contract with `doctor_test.go` and with
  anything reading the output on the box.
- **The `key: reason` shape of every error message.** `rowStatuses`
  splits on the first `": "` (`doctor.go:134-139`). The shape is
  load-bearing in two places, measured: the owned-wall loop, which is
  the only loop that consults the statuses map (`doctor.go:35`), and
  `thresholdStatus`'s parse-error path (`doctor.go:112-127`). Reserved
  and integration rows print fixed status text (`doctor.go:42`,
  `:49-51`) and would blank nothing. It stays a package-wide convention
  everywhere else, and every message keeps it.
- **Universality.** No site-specific value, path, or identity default
  anywhere outside `internal/config`, and after 1b, none inside a rung's
  own declaration file either. Every default ships a stated basis
  (`BS:189-199`).

### The eight spec corrections Task 7 drafts

**Every anchor below was re-measured at revision 2, and Task 7 re-reads
each cited line before writing its replacement.** Correction 1's anchor
was `BS:186-188` at revision 1, which is blank and the defaults bullet;
correction 5's was `rung-4:1693-1695`, which is the lane-registry
bullet's tail. Corrections 6, 7 and 8 are new: the gitleaks fix makes
three more sites stale than revision 1 counted.

1. **`BS:182-183`**, "Rung 1 derives the complete table from this
   classification and it becomes the config core's contract." The
   classification stays the contract. The table stops being complete and
   closed, and becomes registry contents each rung declares. Rung 1's own
   rows stay in `internal/config`.
2. **`PA:56`**, the 0b scope sentence, which names "wall rows and
   reserved classes" only. The code and the carries both assign
   `integrationClasses` to 1b (`keys.go:211-213`, `HISTORY:607`), and the
   superseded map comes with it by necessity, because it is the fourth
   branch of `rejectNamespace` and a name has to resolve somewhere
   (Resolution R8, inventory surprise 8).
3. **`PA:56`**, the gitleaks rider's blast-radius sentence, "any
   insertion above line 66 reds the gate for every lane at once". The
   `--no-git` fingerprint is `file:ruleID:startLine`, so only an
   insertion above line 66 **of `internal/config/keys_test.go`** breaks
   the ignore. Every lane feels it because `scripts/check.sh` is the
   whole build's one gate (inventory surprise 6).
4. **`docs/specs/2026-09-05-rung-4-shadow.md:1689-1692`**,
   the rung 4 staleness question "Has rung 1b landed?". This window runs
   1b after 4b, so the branch never fires for 4b and the re-expression of
   ruling 1's option (a) into registry terms becomes 1b's own migration
   work (Resolution R14, inventory surprise 13).
5. **`docs/specs/2026-09-05-rung-4-shadow.md:1697-1699`**,
   the standing staleness question about `.gitleaksignore`'s pinned line.
   Task 1 retires the pin, so the question is answered permanently rather
   than restated in every later plan.
6. **`docs/specs/2026-09-05-rung-4-shadow.md:1476-1485`**,
   which restates the fragility in full, names `.gitleaksignore:1` and
   line 66 by value, and states the standing rule (`:1483-1484`) "add
   below line 66, or re-fingerprint `.gitleaksignore` in the same
   commit". After Task 1 there is no pin, no allowed finding, and no
   such rule. The replacement states the discipline that succeeds it: a
   declaration file that writes a key literal beside a `Key:` field
   applies the field-name rename, and the repo carries no suppression.
7. **`docs/specs/2026-09-05-rung-4-shadow.md:1787-1788`**,
   risk 7 and its mitigation, which point at the same retired rule.
8. **`PA:56`**, the 0b rider's first clause, "the work must
   re-fingerprint `.gitleaksignore`". The work retired the file instead,
   which discharges the rider more completely than re-fingerprinting
   would have. Correction 3 covers the rider's second clause, the blast
   radius; this one covers the first, so the spec stops directing every
   future reader to re-fingerprint a file that no longer pins anything.

## The registry, as outcomes and constraints

No implementation code appears here. Each item is an outcome a task
delivers and a reviewer verifies from the diff.

**O1. A rung declares its keys in its own file, inside the package that
owns them.** After 1b, `internal/config/keys.go` carries rung 1's own
declarations plus the reserved classes whose owning packages are not yet
born, and each of those leaves when its package arrives (O2). Rung 4's
shadow class and its two
thresholds sit in `cmd/dubplate-server/shadow.go`, which is already
rung 4's declaration file. The projected total is ten declaration sites
outside `internal/config` (inventory `:2.5`), which is small enough that
one visible wiring function in `cmd/dubplate-server` stays readable and
is the reason Resolution R1 prefers explicit composition to `init`.

**O2. No lane adds to `keys.go` after 1b, and a lane removes only the
class it is taking over.** The contested-surface row (`PA:175`) becomes
true rather than aspirational, in that precise form. The stronger form,
that no lane touches the file at all, is false by R4's own design: all
eight reserved classes belong to future rungs (5, 5, 6, 7, 9, 9, 13, 12
at `keys.go:57-142`), and each moves out when its owning package is
born. Rung 5 is the first, taking `auth and identity` and `serve`. The
forward criterion above states the rule and its registration-time pin.
The interim grant that let rung 4 edit the file (`rung-4:1650`) is spent
and retires with this pass.

**O3. Reserved classes keep their shape and their order.** A reserved
class is a name, a rung, and a set of keys, declared by whichever package
will one day own the capability, or by `internal/config` until that
package exists. Declaration order is the order doctor prints, and the
registry preserves it rather than sorting (Resolution R15).

**O4. A duplicate key or a malformed declaration fails loudly, at
registration, and the gate sees it.** The mechanism is a panic at
registration time; the gate residency is a test that constructs the
production registry (Resolution R3). Both, not either. Justification, in
full, because the brief asks for the choice and its reason:

- **Why registration time rather than `Load` time.** A duplicate is a
  programmer error committed by a rung, never an input an adopter
  supplied. `Load`'s error path is the adopter-facing one, and every
  message in it names a key the adopter can fix. A collision reported
  there would tell an adopter to fix something in the binary.
- **Why a panic rather than a returned error.** All three real stdlib
  registries panic (`flag`, `database/sql`, `expvar`), and the Go bar
  says exported surfaces read like stdlib APIs. A registration function
  that returns an error invites a call site that ignores it, and the one
  call site that matters runs before `main` has anything to print to.
- **Why a test as well.** A panic that only fires when the binary runs is
  not a gate. A gate-resident test that constructs the production
  registry exactly as `newRootCmd` does makes any lane's collision fail
  `bash scripts/check.sh` on that lane's own commit, which is the whole
  point of the pass.

**O5. The six collisions the gate catches, named.** This list is
authoritative and Task 2's outcome 4 restates it verbatim. Each class
gets one test that fires it and one mutation proving the test fires.

1. A key declared twice within one declaration family.
2. A key declared in one family and also in another: a reserved key that
   is also an integration key, a superseded name that is also an owned
   row.
3. A declared key that does not carry the `DUBPLATE_` prefix, which
   would make the declaration dead because `rejectNamespace` walks the
   prefix (Resolution R16).
4. A threshold whose declared `Default` does not parse under its
   declared `Kind`, or whose `Kind` names no parser (Resolution R11).
   **This class's mutation lives in Task 5**, with the kind it exists to
   protect; every other class's mutation lives in Task 2.
5. An integration class whose `Flag` is empty, or whose `Flag` also
   appears in its own `Required` or `Optional` set.
6. A wall row that is neither a prefixed key nor a status row carrying
   `Pending` text, which is Resolution R16's carve-out read in the other
   direction.

**O6. Doctor's rows stay derived, and the derivation is proven.** Doctor
already derives all four blocks from the package's own tables so no rung
hard-codes a count (`doctor.go:24-72`). After 1b it derives them from the
registry. The sync mechanism is **derivation, not a rule**: doctor prints
exactly **one line per declaration**, in the registry's own block order,
so a rung that declares a new class, wall row or threshold gets its
doctor row with no edit to `doctor.go`.

**The unit differs per family, and saying so is what makes the count
checkable.** One line per owned wall row, one per reserved *class*
regardless of how many keys it holds, one per integration *class*, one
per threshold. Today that is five owned rows, eight reserved classes
holding 29 keys, one integration class holding six names, and two
thresholds. A key added inside an existing class adds no doctor line, by
design.

**The count assertion already exists** and is not this pass's to build:
`cmd/dubplate-server/doctor_test.go:127-129` closes `TestDoctorCmd` with
`if li != len(lines)`, after walking all four families. The migration
must keep it green. What does not exist, and what Task 6 builds, is the
direction that only a value registry makes reachable: a declaration
added to a scratch registry inside a test adds exactly one doctor line,
in its own family's block.

**O7. The env template's coupling is stated rather than assumed.** The
mechanism is a test, and the direction it cannot assert is a documented
rule with its reason (Resolution R10, Task 6). 1b adds no name to any
template.

**O8. Exemplary Go, per the standing bar.** Concrete types over
interfaces, no functional options, no `init()`, no `context.Context` in a
struct, `any` not `interface{}`, `slices`/`maps`/`cmp` idiom,
`MixedCaps`, errors lowercase and context-first with `%w` only where a
caller branches. **The em dash is banned in Go comments**, no opt-out,
and vale scans a brand-new file before it is staged.

## Resolutions this plan takes

Sixteen open readings, each applied as the recommendation and each marked
**provisional, conductor ruling owed**. Every row names what changes if
the ruling goes the other way, so a fold is a named edit rather than a
re-reading. **No task blocks on any of them.**

### R1. The registry is a value assembled in `cmd/dubplate-server`, and `Load` takes it

**Ruled, conductor 2026-09-07, as applied.** The registry is a concrete type
held in a value, built once by `newRootCmd` and passed to `Load` and to
`newDoctorCmd`. It is not package-level mutable state.

The evidence. `flag.NewFlagSet` is the closest stdlib analogue, because a
flag carries a name, a default, a usage string, and a parser, which is
exactly `Threshold`'s four fields; `flag` ships both a value type and a
package-level `CommandLine` behind one type. A value registry retires two
workarounds that exist only because today's registry is a global: the
`sync.OnceFunc` around `registerShadowThresholds` (`shadow.go:49-52`,
doc `:45-48`) and its regression test (`main_test.go:43-53`, doc
`:36-42`), which exist because
`newRootCmd` runs more than once in a test binary and `Register` cannot
refuse a duplicate; and `withRegistry` (`thresholds_test.go:12-17`),
which saves and restores package state around each test. A value registry
also makes O4's panic meaningful rather than defused: with a fresh
registry per construction, a second `newRootCmd` can never double-register,
so a panic that fires is always a genuine two-rung collision.

The rejected alternative is package-level `init` registration, the
`database/sql` shape. `go-conventions` forbids `init` for this repo,
rung 1 already rejected it in writing (`thresholds.go:36-39`), `Load`'s
result would depend on the import graph rather than on visible code, and
a test importing `internal/config` alone would see an empty wall.

**The cost, stated.** `Load`'s signature changes at six call sites in
`cmd/dubplate-server` (`check.go:23`, `doctor.go:31`, `shadow.go:196`,
`:262`, `:299`, `check_test.go:107`) plus `internal/config`'s own two
test call sites (`config_test.go:152`, `:320`). No package under
`internal/` imports `internal/config`, so the blast radius is one
command package and the config package's own tests.

**Fold if ruled otherwise.** Keep the package-level registry and extend
`Register` in place. Task 3's migration then rewires four tables instead
of a signature, the `sync.OnceFunc` and its regression test stay,
`withRegistry` stays, and O4's panic must be softened to
idempotent-on-identical-value so the second `newRootCmd` does not crash
the test binary.

### R2. The two class shapes stay distinct

**Ruled, conductor 2026-09-07, as applied.** `ReservedClass` (`Name`, `Rung`,
`Keys`) and `IntegrationClass` (`Name`, `Rung`, `Flag`, `Required`,
`Optional`) keep their own types and their own registration entry points.
No unified `Class` type with fields that are dead for half its entries.

This is inventory surprise 9, the pass's central design decision, and
neither spec states a preference. The argument for keeping them apart is
that their difference is the contract, not an accident: a reserved class
can never be enabled and errors naming its rung, while an integration
class is enabled by a flag and errors naming that flag. Collapsing them
would put an always-empty `Flag` on eight classes and force the error
wording to branch on emptiness, which is the same hard-coded branch
Resolution R5 is removing elsewhere. A family of typed registration
entry points over one registry is also the stdlib shape: `flag.FlagSet`
has `Var` plus `Int`, `String`, `Duration` and the rest.

**Fold if ruled otherwise.** One `Class` type with an optional `Flag`,
`Keys` folded into `Required`, and the error wording branching on whether
`Flag` is empty. Tasks 2, 3 and 4 each shrink by one type and grow one
branch; the acceptance criteria on error wording are unchanged either way.

### R3. A collision panics at registration, and a gate-resident test constructs the production registry

**Ruled, conductor 2026-09-07, as applied.** Stated in full at O4 and O5.

**Fold if ruled otherwise.** If the ruling prefers a returned error, the
registration entry points gain an error return, the wiring function in
`cmd/dubplate-server` returns it, `newRootCmd` fails the process at
startup with it, and the gate-resident test asserts the error rather than
recovering a panic. Everything else in the plan is unchanged.

### R4. Rung 1's own rows stay declared inside `internal/config`

**Ruled, conductor 2026-09-07, as applied.** The four owned-now rows, the
identity-source pending row, the eight reserved classes and the eleven
superseded names stay declared in `internal/config/keys.go`, and the
registry constructor returns a registry already carrying them.

Two reasons. **The rule is satisfied, not bent**: "a rung declares its
keys in its own file inside the package that owns them", and rung 1's own
package is `internal/config`. **It kills the empty-wall hazard**: a
registry that starts empty would let a test, or a lane's own wiring
mistake, produce a `Load` that requires nothing and accepts everything,
which is the silent-degradation failure the five-class contract exists to
refuse.

A reserved class whose owning package does not exist yet stays here too,
and moves to that package when the package is born. All eight are in
that position today (rungs 5, 5, 6, 7, 9, 9, 13, 12), so **the move is
the normal case rather than the exception**, and rung 5 takes the first
two rather than rung 12 taking the first one. The forward criterion
states the rule that governs every such move: the deleting commit is the
declaring commit, and a registration-time check that the class arrives
under the same name and key set is what stops a move from dropping a
key. O2 is worded to match.

**Fold if ruled otherwise.** If rung 1's rows must also be registered
from outside, `cmd/dubplate-server` gains a rung 1 declaration file and
the empty-registry hazard is answered instead by a test that asserts the
production registry carries the four owned rows.

### R5. `Load`'s `class.Name == "shadow"` branch retires

**Ruled, conductor 2026-09-07, as applied.** After 1b, `Load` resolves every
enabled integration class's `Required` keys and makes them reachable by
class name. The rung-4-specific `ShadowIntegration` type and its
construction move out of `internal/config` and into rung 4's own
declaration file, which is where rung 4's other config declarations
already live.

This is inventory surprise 3, and it is the part of HISTORY carry 4 the
carry does not say out loud: `integrationClasses` is already data, and
what is not data is the branch. The universality argument is the stronger
one. A named rung-4 type inside the universal config package is exactly
the site-specific coupling the package doc exists to prevent, and a
second integration class would add a second branch to a `Load` whose
whole design is that it enforces a classification rather than knowing its
members.

**What must not weaken, and where it now lives.**
`TestLoadShadowIntegration` (`config_test.go:255`) proves all four
required-when-enabled behaviors and is the test HISTORY carry 4 names.
Its eight table rows and every `wantErr` assertion survive verbatim.
Only the two assertions that reach into `cfg.Shadow` are reworded to
reach the resolved values by class name, and Task 4's mutation proofs
must still fire against the reworded test.

**The test moves package, and revision 2 decides that rather than
leaving it implied.** The eight rows are built entirely from
`DUBPLATE_SHADOW_*` literals, and `internal/config`'s three test files
carry 41 case-insensitive `shadow` hits between them, 33 of them in
`config_test.go`. Keeping the rows in `internal/config` and clearing
rung-4 names from the package are contradictory demands. The split:

- **The eight rung-4 rows move verbatim to
  `cmd/dubplate-server/shadow_test.go`**, which is rung 4's own
  declaration file's test, and where the shadow class now lives. Moving
  a table between files is what "survives verbatim" now means, and it is
  the move Task 4 reports line by line.
- **A generic required-when-enabled table test stays in
  `internal/config`**, driven by a fixture integration class the test
  itself declares, with no rung-4 name. Its job is `Load`'s data-driven
  resolution, which is the behavior 1b introduces and which no rung-4
  test can prove is generic.

Both halves are Task 4's, and Task 4's acceptance criteria state the
grep this makes satisfiable.

**Fold if ruled otherwise.** `Config.Shadow` and `ShadowIntegration` stay
in `internal/config` and the branch stays, with a comment naming the
second-branch cost. Task 4 then shrinks to the table migration alone and
`TestLoadShadowIntegration` needs no rewording at all.

### R6. `KindFloat`'s mechanism lands here; the four knobs do not

**Ruled, conductor 2026-09-07, as applied.** HISTORY carry 3 phrases one item
and contains two (inventory surprise 14). The kind is a registry
capability and the registry pass takes it. The registration of
`StrongRecThresh`, `MediumRecThresh`, `RecGapThresh` and `SearchLimit` is
explicitly the first rung 3 pass after 1b closes, said three times
(`HISTORY:602`, `:1231`, `:1602`) and carried in `docs/STATUS.md:34-37`.

Registering the four knobs here would also make `internal/engine/matcher`
import `internal/config`, which no package under `internal/` does today
and which `tunables.go:50-54` declines in writing.

**Fold if ruled otherwise.** If the four knobs are ruled into 1b, Task 5
grows a fifth deliverable, `internal/engine/matcher` gains its first
`internal/config` import, and the plan's package territory grows one
package. Reading the carry as one item and taking neither half is the
other failure: `KindFloat` would then be owed by a rung 3 pass that
cannot land its own knobs without first landing a registry capability.

### R7. `AdminEmail`'s dedicated contact row does not land here

**Ruled, conductor 2026-09-07, as applied.** HISTORY calls it "post-1b"
(`:1470-1475`), and it is a new owned wall row, which means a new
required key for every adopter. That is a product decision, and it is
listed above as one of the four things that would become an owner
touchpoint. 1b's contribution is making the row cheap to add later, from
whichever package owns the User-Agent.

**Fold if ruled otherwise.** One declaration and one owner-touchpoint ask
about minting, in whichever pass takes it.

### R8. The superseded map is 1b's by necessity, and the widening is a spec correction

**Ruled, conductor 2026-09-07, as applied.** The 0b scope sentence names two
families; the code and the carries assign a third (integration classes),
and the fourth (superseded names) follows because it is the fourth branch
of `rejectNamespace` and a name has to resolve somewhere. The map keeps
its map shape: no caller enumerates it, `supersessionOf` is a single
lookup, and order is meaningless for it (inventory `:1.2`).

**Fold if ruled otherwise.** If the superseded family is ruled out of 1b,
`rejectNamespace` keeps reading one package-level map while reading three
registry families, which is the split shape the pass exists to remove.

### R9. The gitleaks fix is the field rename plus a comment, not a `.gitleaks.toml`

**Ruled, conductor 2026-09-07, as applied.** Option A of inventory `:3.4`: the
struct field named `key` is renamed so the generic-api-key rule loses the
keyword half it needs, `.gitleaksignore` empties, and the line pin stops
existing. A comment beside each renamed field states why it is not called
`key`, which is the rider's own "note that fragility in the file"
requirement, discharged where a later author will read it.

**The rename covers both tables in the file, not one.**
`internal/config/keys_test.go` has two tables with a `key` field:
`TestReservedClassOf` at `:59-92`, which holds today's finding, and
`TestSupersessionOf` at `:103-113`, which does not, only because none of
its five rows clears the entropy threshold. Measured on scratch copies
under gitleaks 8.30.1: renaming only the first table's field and adding
one high-entropy `DUBPLATE_*` row to the second reports `leaks found: 1`;
renaming both scans clean. A fix scoped to one table leaves the next
author of the other to red the gate for every lane, which is the outcome
the rider exists to prevent.

Two rejected alternatives, each with its measured cost. An inline
`// gitleaks:allow` comment (Option B) is verified clean and survives
insertions, but it suppresses that line permanently, including a future
real secret written on it. A repo `.gitleaks.toml` (Option C) is the only
option that also protects future declaration files automatically, and it
is the only option with a measured way to be silently wrong: a
`[[allowlists]]` block defaults to `condition = "OR"`, and a global
allowlist with `regexes = '''DUBPLATE_[A-Z0-9_]+'''` was measured to
suppress a genuine high-entropy `DUBPLATE_MAGIC_LINK_SECRET` finding that
the same scanner flags without it (inventory surprise 5). Adding
`condition = "AND"` fixes it, verified, and a plan that reaches for the
file without that keyword silently widens suppression across the whole
repo. A config file whose correctness turns on one easily-omitted keyword
is a worse bargain than a rename that leaves nothing to audit.

**The generalization the rename buys**, and the reason it is worth
stating rather than performing silently: the trigger is a `key`-like
field name on the same line as a name whose entropy clears the rule's
threshold. Exactly one of the 54 `DUBPLATE_*` names in the tree fires it,
`DUBPLATE_R2_BUCKET` at entropy 3.6143694 (inventory surprise 4), and
`keys.go`'s own `Keys: []string{...}` blocks do not fire because they
supply no keyword. Any declaration file this pass creates that writes a
key literal beside a `Key:` field can therefore create a new finding, and
the same discipline answers it.

**Fold if ruled otherwise.** Option C, with `condition = "AND"` mandatory
and a test or a comment in `scripts/check.sh` naming why. Task 1's files
change from `keys_test.go` plus `.gitleaksignore` to a new
`.gitleaks.toml`, and the fragility note moves there.

### R10. 1b adds no name to any template; the coupling gets a test and a stated rule

**Ruled, conductor 2026-09-07, as applied.** `TestLoad_TemplateKeysClassified`
(`config_test.go:196`) makes `internal/config`'s test binary read
`env/dubplate.env.template` by a hard-coded relative path, which is a
gate-resident coupling between a Go package and the ops tree (inventory
surprise 10). 1b keeps the coupling, re-points its classification at the
registry, and states both directions:

- **The direction that is mechanical and stays asserted:** every
  `DUBPLATE_*` name a template assigns classifies as owned, reserved,
  superseded, **or as an integration class's flag, `Required` key or
  `Optional` key**, never unknown. This is the assertion that protects a
  real adopter whose env file derives from a template. **The fourth
  branch is mandatory rather than optional**, and revision 2 adds it:
  today's walk (`config_test.go:202-213`) tries `ownedNow`,
  `reservedClassOf` and `supersessionOf` only, and passes because the
  one shipped template carries zero `DUBPLATE_SHADOW_*` names. Pass 4b's
  `env/dubplate-server.env.template` carries six of them
  (`rung-4:1632`), and a shadow key resolves through
  `integrationClassOf`, which is `rejectNamespace`'s fourth branch
  (`config.go:182`). Without the fourth branch the test reports six
  names unknown and the gate reds for every lane.
- **The direction that becomes mechanical only against a greenfield
  template:** every owned-now row appears in the template the greenfield
  server ships. Today it fails, because `DUBPLATE_LIBRARY_ROOT` and
  `DUBPLATE_SITE_NAME` appear nowhere in `env/dubplate.env.template`,
  whose own header says outright that it describes the frozen pre-pivot
  binary's wall. **1b does not fix that by adding the two names**,
  because adding a name to a template requires minting its value
  (`deploy.sh:171-194`) and that is an owner touchpoint.
- **The rule, documented where the test lives:** a rung that declares an
  owned-now row adds it to the greenfield template in the same pass, with
  the minting that implies. The rule is stated in the test's own doc
  comment, so the next author reads it at the point of failure.

Task 6 asserts the completeness direction **only if** pass 4b landed
`env/dubplate-server.env.template` carrying the four owned-now rows, which
Task 0 determines. If 4b did not, the direction stays a documented rule
and Task 7 records it as a carry for whichever pass lands that template.

**Fold if ruled otherwise.** If the ruling says the completeness
assertion must land regardless, the pass acquires an owner touchpoint and
the conductor batches the minting ask; nothing else in the plan moves.

### R11. `KindInt` stays the zero value, and an unnamed kind becomes a registration error

**Ruled, conductor 2026-09-07, as applied.** `ThresholdKind`'s zero is
`KindInt` deliberately, so a `Threshold` declared before the field existed
keeps its meaning (`thresholds.go:13-18`), and existing literals depend on
it. The hazard is `thresholdStatus`'s `default` branch
(`doctor.go:111-128`), which parses any unnamed kind as an integer, so a
float threshold whose `Kind` was never set would report as an environment
parse failure rather than as the registration bug it is (inventory
surprise 11). The registry already draws that distinction in both parsers
(`thresholds.go:60-63`, `:84-87`).

The resolution keeps the zero value and moves the check, and **revision
2 corrects what the check does**. Refusing an unnamed kind catches
nothing that matters, because `KindInt` is `iota`'s zero: a float
threshold whose `Kind` was never set has a named kind and passes any
enum-membership test. Rung 3's `StrongRecThresh` is the live case
(`tunables.go:62-66`, default 0.04); registered without
`Kind: KindFloat` it registers clean and doctor prints
`invalid: invalid default "0.04"`, which is exactly the environment-shaped
report surprise 11 exists to prevent.

**Registration parses each declaration's own `Default` under its
declared `Kind`.** `Threshold{Key: X, Default: "0.04"}` with the zero
kind then fails registration with `X: invalid default "0.04"`, caught by
the gate-resident test that constructs the production registry, which is
the altitude the hazard belongs at. This also folds both parsers'
existing invalid-default line (`thresholds.go:60-63`, `:84-87`) up to
registration, where a bad default has always been a registration bug in
the code's own words. An out-of-range kind is refused by the same check,
and doctor's kind switch names every kind explicitly so its remaining
fallthrough reports a registration bug rather than parsing.

**Fold if ruled otherwise.** Give `ThresholdKind` an explicit invalid
zero and set `Kind` on every existing literal. That is a larger diff and
it deletes the reason rung 1 wrote the zero-value comment.

### R12. Pattern key families and `DUBPLATE_COLLECTIONS` are declined and filed

**Ruled, conductor 2026-09-07, as applied.** The salvage tree holds two key
shapes the greenfield table cannot express (inventory surprise 12):
`pingURLsFromEnv`'s `DUBPLATE_PING_<SLUG>_URL` family, discovered by
scanning rather than declared, which the greenfield froze to four literal
names (`keys.go:116-125`); and `DUBPLATE_COLLECTIONS`, which is in no
greenfield class at all, so an adopter setting it today gets "unknown
configuration key" while CLAUDE.md still lists collection routing as live
behavior.

Both are declined for 1b and recorded as shapes the registry must
eventually express, filed to rung 9, which owns the monitoring class and
the pipeline. **A pattern family is a registry capability**, so the
filing is 1b's to make even though the build is not 1b's.

**Fold if ruled otherwise.** A pattern declaration family joins Task 2,
which grows one type and one matcher, and `rejectNamespace` gains a
prefix-match branch before its unknown fallthrough.

### R13. The three missing reserved classes are recorded, not declared

**Ruled, conductor 2026-09-07, as applied.** The build spec's config-wall
section names an SSO proxy capability, a trusted-proxy mode and a
forward-auth front among the flag-enabled integrations (`BS:167-175`);
none has a reserved class, the forward-auth front appears only inside two
supersession messages (`keys.go:167-170`), and the other two appear
nowhere in `internal/config` (inventory surprise 7). The same bullet says
the Navidrome variables "move behind its flag" while they sit in a
reserved class instead, which is the right interim shape and not the one
the spec describes.

1b declares none of them. Declaring a reserved class for a capability
whose rung has not been planned invents a key set, and the spec's own
words are the wrong source for one. Task 7 records the divergence for
rungs 5 and 12 to answer when they plan.

**Fold if ruled otherwise.** Three reserved classes with key sets someone
has to author, which is a rung 5 and rung 12 spec question rather than a
1b implementation one.

### R14. The re-expression of rung 4's ruling 1 is 1b's own migration work

**Ruled, conductor 2026-09-07, as applied.** Rung 4's staleness pre-task asks
"Has rung 1b landed?" as a blocker and rules that a yes moves ruling 1's
option (a) into registry terms (`rung-4:1689-1692`). This window inverts
the order, so the branch never fires for 4b and the re-expression lands
here (inventory surprise 13). Task 4 is that work, and Task 7 records the
spec correction.

**Fold if ruled otherwise.** Nothing moves; the work happens either way,
and the resolution exists so a reader of rung 4's spec is not left
looking for a 4b task that never existed.

### R15. Declaration order is preserved; the registry never sorts

**Ruled, conductor 2026-09-07, as applied.** Doctor prints reserved classes in
table order, that order is not rung order, and `TestDoctorCmd` walks the
same order (`doctor.go:41-46`, `doctor_test.go:89-98`). A registry that
sorted on rung, on name, or on registration time from multiple packages
would change user-visible output. Order is therefore declaration order
within a family, and family order is the block order doctor already
prints: owned wall, reserved classes, integration classes, thresholds.

**Fold if ruled otherwise.** If a ruling prefers sorted output, doctor's
tests and any box-side reader change with it, and the reason the current
order exists (table readability) is lost with no compensating gain.

### R16. Every declared key carries the `DUBPLATE_` prefix; the pending row is not a key

**Ruled, conductor 2026-09-07, as applied.** `rejectNamespace` walks names by
the `DUBPLATE_` prefix (`config.go:175`), so a declared key without it is
dead: no adopter setting it would ever be classified, and the declaration
would silently do nothing. The prefix is therefore a registration-time
check, part of O5.

The one carve-out is the identity-source row, whose `Name` is the prose
`identity source` and whose `Pending` text is what doctor prints. It is
declared as a status row rather than as a key, so the prefix rule does not
reach it and `ownedNow` keeps returning false for it. Getting this wrong
in either direction is visible: a prefix check that reached the pending
row would refuse rung 1's own registry at startup, and a status row
treated as a key would make `rejectNamespace` claim `identity source` is
owned.

**Fold if ruled otherwise.** If the pending row is ruled into the key
family, it needs an env key it does not have, which is rung 5's decision
and not 1b's.

## Prior art consulted, per CLAUDE.md

The build-for-2.0 rule requires each rung plan to name which entries of
the 2.0 spec's prior-art list it consulted, and to adopt or decline each
with a reason. The list carries twenty-one rows plus a declined
paragraph (`2026-09-07-streamer-2-0-design.md:402-427`). **The table
below covers every entry that could bear on process configuration**, and
the rows it omits share one reason: a tag reader, a SQLite driver, an
encoding, a cache, a fingerprinter and a compatibility note have no
configuration surface for a registry to adopt or decline. The one entry
that does bear on it, `viper`, is declined explicitly and on the record.

| Entry | Verdict for 1b | Reason |
|---|---|---|
| The OpenSubsonic specification | **Not applicable** | A streaming API surface. It defines no process-configuration mechanism. |
| Navidrome (behavior only, GPL) | **One entry adopted as a decline**, below | Its config surface is `viper`, addressed on its own row. Nothing else in it concerns process configuration. |
| gonic (behavior only, GPL) | **Not applicable** | Same reason. |
| `viper`, named in the 2.0 spec among Navidrome's dependencies architecture 2.0 chose against | **Declined, explicitly and on the record** | A config-registry rung is exactly where a configuration library gets reached for, so the decline is stated rather than assumed. The stdlib plus `strconv` plus `time.ParseDuration` covers every parser the wall has or wants, the wall reads only the process environment through one choke point, and a library that reads files, flags and remote stores would widen the surface the five-class contract exists to narrow. `go.mod` carries cobra and its two indirects and nothing else. |
| `pressly/goose` | **Not applicable** | A database migration runner. It touches no configuration surface. Already declined by the 2.0 spec on its own terms. |
| `modernc.org/sqlite` | **Not applicable** | Adopted conditionally at rung 6 for FTS5 and WAL. No config relevance. |
| `golang.org/x/time/rate` | **Not applicable to 1b** | Adopted for 2.0's auth rate limits. Its knobs will one day be registered thresholds, which is a consumer of this pass rather than an input to it. |
| taglib via wazero, `mewkiz/flac`, `x/image` | **Not applicable** | Tag and media I/O. |
| Picard's tag tables, chromaprint | **Not applicable** | Tagging and fingerprinting data. |
| Melodee | **Not applicable** | A library manager, consulted by the 2.0 study rung for behavior. |

**No external library is proposed for 1b, and none is added.** The one
adoption worth naming is internal to the stdlib and is imitated rather
than imported: `flag`'s panic-on-redefinition semantics, together with
`database/sql.Register`'s and `expvar.Publish`'s, which is the evidence
behind O4 and Resolution R3. `testing/fstest` was consulted and offers
nothing: `MapFS` is a test double and `TestFS` a conformance checker,
neither a registry, and the plan records that rather than reaching for
the analogy.

## The 2.0 forward hooks this pass touches

Per CLAUDE.md's build-for-2.0 rule, each task below carries its own hook
statement. The pass-level picture, so a reviewer can check a task's claim
against it:

- **Hook 6, the driver seam stays universal (rung 12), is landed as an
  enabling shape by Tasks 2 and 4.** The hook says address, proxy and
  provisioning are per-driver capabilities a driver may decline to
  declare, because an in-process streamer has no address and needs no
  proxy. Today the `streamer` reserved class hard-codes three Navidrome
  names (`keys.go:133-141`), which no driver can decline. A registry
  where the driver's own package declares its own class is what makes the
  hook reachable, and building rung 6's schema or rung 12's driver seam
  against a fixed class table would be rework.
- **Hooks 5, 12 and 14 are advisory and push the same direction.**
  Hook 5 (`2026-09-07-streamer-2-0-design.md:367-370`) is about rung 6
  minting a durable surrogate user id and rung 5 resolving a roster
  identity to it; its bearing on 1b is that the identity-source wall row
  eventually grows a required-when-selected block, declared through this
  registry. **The discriminator language belongs to
  `internal/config/config.go:11-17`, the package doc, rather than to the
  hook**, and revision 2 attributes it there. Hook 12's shape claim
  narrows to its forbidden-character set: the 255-byte bound at spec
  `:385-388` is plainly an integer and one of today's kinds already
  expresses it. Hook 14's ID3 version ruling is a two-valued choice. The
  advisory reading that survives is that the kinds a threshold can take
  need room to grow. 1b takes the room it can justify now, which is
  `KindFloat` and a registration-time check on every declaration's own
  default (Task 5), and declines to invent a string or enumerated kind
  no rung has specified.
- **Every other hook is neutral to this pass.** Hooks 1, 2, 3, 4, 7, 8,
  9, 10, 11 and 13 are schema, tagging, backup and transaction concerns
  at rungs 6 through 9, and none reads or writes a configuration key
  through a shape 1b changes.

## Contested surfaces

The lane-launch checklist requires a written owner for every surface a
lane touches. Every row 1b writes has one. **The `PA:` numbers below are
the live ones Task 0 measured** at revision 2.1 (see Primary inputs,
item 2), seventeen lines below their committed-HEAD `c98a916` positions.

| Surface this pass writes | Row | Note |
|---|---|---|
| `internal/config` (`keys.go`, `thresholds.go`, `config.go` and their tests) | `PA:175` and `PA:200` | 1b is the owner named in both. The interim rung 4 grant (`rung-4:1650`) is spent and retires with this pass. |
| `cmd/dubplate-server` (`main.go`'s wiring, `doctor.go`'s rows, `shadow.go`'s declarations, `check.go`'s `Load` call) | `PA:199`, "rung 1's skeleton; each rung adds its own file and registers through it" | 1b edits rung 1's skeleton itself and rung 4's own declaration file. |
| `.gitleaksignore` and `scripts/check.sh`'s gitleaks comment | The 0b rider (`PA:56`) | The rider is the mandate; Task 1 discharges it. |
| `env/dubplate.env.template` | `PA:201`, which now also names `Makefile`, rung 1's, with 4b adding one template | **1b writes no name into it.** Task 6 reads it and points a test at it; that is all. |
| `docs/` ledgers | `PA:194`, conductor-owned | A pass drafts and never commits. |

**Pass 4b precedes 1b in the same tree**, and its named deliverables
exclude `internal/config` (`rung-4:1632`). Two edges survive that clean
separation, and **Task 0 re-reads both**:

1. **`env/`.** 4b creates `env/dubplate-server.env.template` and amends
   the existing template's header.
   `TestLoad_TemplateKeysClassified` reads
   `../../env/dubplate.env.template` by a hard-coded relative path
   (`config_test.go:221`). Whichever pass lands second owns the
   reconciliation, and in this window that is 1b. Task 0 reports which
   templates exist, which names each carries, and whether the four
   owned-now rows appear in a greenfield one. Task 6 acts on the answer.
2. **`cmd/dubplate-server/main.go`.** 4b edits it for the two `-ldflags`
   version stamps; 1b edits it at or near `newRootCmd` for the registry
   wiring. Different functions, so a textual conflict is unlikely and a
   semantic one is impossible, but Task 0 re-reads the file as landed
   rather than as this plan cites it.

**The other 4b files Task 0 re-reads**, because 1b cites the tree at
`c98a916` and 4b lands between: `scripts/deploy.sh` (six edits, and the
template constants `ENV_TEMPLATE` at `:27` and `SERVER_ENV_TEMPLATE` at
`:28`), `Makefile`, `cmd/dubplate-server/version.go`
or wherever the stamps print, `systemd/dubplate-shadow.service` and
`.timer`, and `docs/STATUS.md`'s lane registry. If 4b touched
`internal/config` at all, that is a Task 0 blocker: every citation in
this plan is against `c98a916` and must be re-verified before Task 2
dispatches.

## Token ceiling: 2.5M

Re-derived from pass 4a-ii's measured rate against this plan's task
count, with the arithmetic shown, per the standing habit of refining a
ceiling from a measurement rather than asserting one.

**The measurement.** 4a-ii spent about 2.7M against a 3M ceiling
(`HISTORY:636-648`): staleness pre-task and chain-args extraction 0.26M,
the chain 1.32M, the Task 13 and Task 14 fix and review rounds 0.49M,
simplify 0.18M plus its review, the close drafter 0.16M, and the
conductor's own turns about 0.2M. The pass ran six tasks in a chain of
sixteen agents. Two figures follow: **0.22M chain cost per task**, and
**49% chain share of the pass**.

**The adjustment, argued rather than asserted.** 4a-ii's tasks built
ledger fields, marshalling, three subcommand verbs, a mechanical
re-evaluator and an end-to-end exit gate. 1b's tasks refactor one
1255-line package whose tests are already written and whose seventeen
call sites are already enumerated in the inventory. The inventory argues
a config task should cost "well under half a shadow task". This plan
takes **two-thirds rather than half**, deliberately, because Resolution
R1's value-type registry changes `Load`'s signature and therefore touches
two packages and both their test suites, which is a wider diff than the
inventory's own proposal assumed. **0.22M × 2/3 ≈ 0.15M per chain task.**

**The arithmetic.**

- Seven chain tasks (Tasks 1 through 7) at 0.15M: **1.05M**.
- Task 0, a fresh-context read that writes nothing: **0.10M**, on
  4a-ii's own 0.26M for a pre-task plus chain-args extraction, of which
  this pass needs only the pre-task half against a smaller surface.
- Chain and pre-task total: **1.15M**.
- Cross-check one, the chain share. 4a-ii's 49% is a **chain-only**
  share (1.32M of about 2.7M), and its pre-task sat outside that 1.32M
  in a separate 0.26M line item, so the share divides the chain figure
  alone: 1.05M ÷ 0.49 = **2.14M**. Revision 1 divided 1.15M, which mixed
  the two categories and inflated the result to 2.35M.
- Cross-check two, the non-chain half from 4a-ii's own line items and
  independent of task count: fix-and-review rounds 0.49M, simplify 0.18M,
  close drafter 0.16M, conductor turns 0.20M. HISTORY reads "simplify
  0.18M plus its review", and its line items sum to 2.61M against "about
  2.7M", so about **0.09M** of simplify review is unallocated and is
  carried rather than dropped. Total non-chain **1.12M**, added to 1.15M
  gives **2.27M**.
- The two corrected cross-checks agree within 6%. **The ceiling is
  2.5M**, the higher figure rounded up, which leaves the same headroom
  revision 1 intended.

**Why this is above the inventory's proposed 2M.** The inventory proposed
2M against an eight-task list and said in the same breath that the plan
"should re-derive this from its own task count once the shape decision is
made, since a value-type registry is a larger migration than a fixed
package-level one". Resolution R1 takes the value-type registry, so the
re-derivation moves up. It is still S band: the S-band precedent is 4b's
1.5M for a deploy pass, and 2.5M buys a contract change across eighteen
exported symbols, a migration of four tables of four shapes, a signature
change at eight call sites, and the gitleaks rider with its
verification.

**Flag at 80%, which is 2.0M.** At the flag: finish the task in flight,
write STATUS, and one combined question goes to **the conductor**, per
the owner grants for this window. The conductor escalates to Geoff only
if a re-derivation past 2.5M also breaks a window cap.

**Checkpoint interval: four tasks**, the default. STATUS is written at
Task 3 and at Task 7, at any split, and before any question. Every task
reports its own spend and its own gate evidence, and the conductor sizes
the next task against it.

Cumulative spend to date is about 45.4M (`docs/STATUS.md:59`).

## Global Constraints

- **Main tree.** 1b is Phase 0b and its landing is what unblocks lane-2
  (with 0c), so no second code lane exists while it runs. The lane
  registry entry is written before the first dispatch, and **the
  live-executor check is that registry, never `pgrep -f
  Projects/dubplate`**, which matches every lane once a worktree exists.
- **Package territory, stated as a wall.** This pass writes
  **`internal/config`** (`config.go`, `keys.go`, `thresholds.go`, a new
  registry file, and their tests) and **`cmd/dubplate-server`**
  (`main.go`, `doctor.go`, `check.go`, `shadow.go` and their tests), plus
  `.gitleaksignore`, one comment in `scripts/check.sh`, and
  conductor-owned `docs/` ledgers. It writes **nothing** under
  `internal/engine`, `internal/harness`, `corpus/`, `systemd/`, `env/`,
  `config/`, `web/`, `tools/`, `Makefile`, or `_salvage/`. A task that
  believes it needs one of those stops and reports.
- **No name is added to any environment template**, for the reason
  Resolution R10 states. Task 6 reads templates and never writes one.
- **Module dependencies: none added.** `go.mod` carries cobra and its two
  indirects and nothing else. A task that wants one stops and reports.
- **Universality.** No estate hostname, path, email, catalog value or
  personal name anywhere, and after this pass, none inside a rung's own
  declaration file either. Every default carries a stated basis where it
  is declared (`BS:189-199`).
- **`Load`'s observable behavior does not change**, except where a
  resolution names the change and a task's acceptance criteria state it.
  Tasks 2 and 3 change no behavior at all: their proof is that every
  existing test in `internal/config` and `cmd/dubplate-server` passes
  **unchanged**, with no assertion weakened, no test deleted, and no
  `t.Skip` added. A task that finds it must edit an existing assertion to
  make the migration pass stops and reports, because that is the
  signature of a behavior change disguised as a refactor.
- **Assertion discipline.** No silent-success fakes, no vacuous
  assertions, no self-derived expectations, no skipped placeholders.
  **Every task names its mutations and reports each one's measured
  outcome.** A registry that silently accepts a duplicate is this rung's
  equivalent of a gate that grades nothing, so the collision gate's own
  tests get the strictest treatment: each collision class in O5 gets one
  test that fires and one mutation proving it fires.
- **Goldens are last resort.** Doctor's output is the one place a golden
  is tempting; assert on parsed rows and on counts derived from the
  registry, not on a captured block of text.
- **`go-conventions` before any Go**, at the standing exemplary-stdlib
  bar. Concrete types over interfaces, no functional options, no
  `init()`, `any` not `interface{}`, `slices`/`maps`/`cmp` idiom, `ctx`
  first where a context appears, errors lowercase and context-first with
  `%w` only where a caller branches, `MixedCaps` throughout. **The em
  dash is banned in Go comments**, no opt-out, and vale scans a brand-new
  file before it is staged.
- **Every task leaves an inline resolution marker where it applies a
  resolution's shape.** One line on the symbol, naming "provisional,
  conductor ruling owed" followed by the resolution's own number, in the
  form this plan's Resolutions section spells out (R1 through R16). Rung
  4 used to carry two such markers; 4b's Task 6 ratified them, replacing
  the provisional wording with "Ratified as owner ruling N" at the lines
  such markers used to sit (`cmd/dubplate-server/shadow.go:22`, `:37`).
  Each task reports the file and the symbol in its return, which is what
  makes Task 7's marker list a search rather than a re-reading, and what
  makes a conductor ruling that goes the other way a named edit.
- **Every error message keeps the `key: reason` shape.** `rowStatuses`
  splits on the first `": "`. Losing the shape blanks an owned row's
  doctor status (`doctor.go:35`) or a threshold's parse-error status
  (`doctor.go:112-127`) without failing anything else. Reserved and
  integration rows print fixed text and would blank nothing, so a
  reserved-class error that loses the shape is caught by its own test
  alone.
- **The gate is `bash scripts/check.sh`, green at every task end.** Never
  commit red. Specific files only, never `git add -A`. Commit footer
  `Co-Authored-By: Claude <noreply@anthropic.com>`. The gate's Go leg
  runs `./...`, which includes the conformance tiers, so **a task's gate
  run is minutes rather than seconds and that is expected**.
- **Anything reported as evidence runs with `-count=1` and reports wall
  time.** A sub-second test run tells you nothing.
- **The staleness pre-task runs first.** Task 0. Its output is blockers
  with evidence, not fixes; the conductor rules each before Task 1
  dispatches.

---

## Task 0: The staleness pre-task

Model: `sonnet`, **fresh context**, dispatched before any implementer.
Reads this plan and the inventory against the tree as shipped. **Its
output is blockers with evidence, not fixes**; it writes no code and
edits no file. The conductor rules each blocker before Task 1 dispatches.

**Every check runs against the real HEAD, and the task reports what that
HEAD is.** This plan's citations were verified at `c98a916`, before pass
4b ran. A citation that has moved is a finding for this task, not a
surprise for an implementer.

**The first three are blockers: a "no", or an unexpected "yes", stops the
launch rather than producing a note.**

1. **Did pass 4b touch `internal/config` in any way?** *(Expected: no.
   4b's named deliverables exclude it, `rung-4:1632`.)* Report the answer
   with `git log` and `git diff --stat` evidence against `c98a916`. **A
   yes is a blocker**: every citation in this plan is against `c98a916`,
   and Tasks 2 through 5 rest on the file-by-file shapes the inventory
   recorded.
2. **Is the lane registry consistent with where this pass intends to
   run, and is the primary tree free?** `docs/STATUS.md`'s lane registry
   is the live-executor check. *(Expected: lane-1 idle or closed on 4b,
   lane-2 BLOCKED until Phase 0b and 0c.)* Report any live pid in this
   tree, or a lane-1 entry naming another workflow, as a blocker. Report
   any warm uncommitted change the task did not author as a blocker,
   **with one pre-ruled exception: the Phase 0c completion note in
   `docs/specs/2026-09-03-parallel-build-approach.md`**, a
   fifteen-line insertion after `:54` from the document lane. It is a
   known landing rather than an unauthored edit, it is not a blocker,
   and Task 7 outcome 8 already expects to report 0c's state. Report it
   as a note, and report whether it is still uncommitted.
3. **Is `.gitleaksignore` still a single line pinning a finding in
   `internal/config/keys_test.go` by line number, and is that line still
   the row it names?** *(Expected: one line, the pin at 66, the first row
   of `TestReservedClassOf`'s table.)* **Report the number and the row's
   identity without quoting the fixture literal**, because gitleaks scans
   the tracked worktree with `--no-git` and would flag the quotation
   itself. A moved pin changes Task 1's starting state, not its outcome.
4. **Which environment templates exist, and what does each carry?**
   Report the file list under `env/`, and for each template the count of
   `DUBPLATE_*` names assigned on their own lines and whether the four
   owned-now rows (`DUBPLATE_LIBRARY_ROOT`, `DUBPLATE_STATE_DIR`,
   `DUBPLATE_SITE_NAME`, `DUBPLATE_ADMIN_EMAIL`) each appear. *(Expected
   at `c98a916`: one template, `env/dubplate.env.template`, **39**
   `DUBPLATE_*` names, of which `STATE_DIR` and `ADMIN_EMAIL` appear and
   `LIBRARY_ROOT` and `SITE_NAME` do not. The file assigns 42 uppercase
   names in total; three are `ND_`-prefixed and the package's own parser
   skips them, so 39 is the figure to reproduce. If 4b landed
   `env/dubplate-server.env.template`, report it the same way.)* **This
   answer decides Task 6's second assertion** (Resolution R10), and it is
   the reason Task 6 is late in the order rather than early. **Also
   report whether any template carries a `DUBPLATE_SHADOW_*` name.** A
   yes makes the classification test's fourth branch mandatory rather
   than optional, which Task 6 acts on.
5. **What did 4b change in `cmd/dubplate-server` and `scripts/deploy.sh`?**
   Report `main.go`'s current shape around `newRootCmd`, where the
   version stamps print, and whether `scripts/deploy.sh`'s template
   constant still sits at one line. Informational; it is what keeps
   Task 3's wiring edit from colliding with a stamp edit.
6. **Is `internal/config` still the shape this plan cites?** *(Expected
   at `c98a916`: `config.go` 250 lines, `keys.go` 254, `thresholds.go`
   96; `WallRow` two fields; `ownedWall` five rows with the identity
   source carrying `Pending`; `ReservedClass` three fields with eight
   classes over **29** keys (8, 2, 6, 2, 3, 4, 1, 3 in table order;
   revision 1 said 27, carried from inventory `:92`, and the tree says
   29), `decision and review queues` preceding
   `streamer` in table order; `supersededKeys` a map of eleven entries;
   `IntegrationClass` five fields with one class; `ThresholdKind` two
   values with `KindInt` the zero; `Register` a bare append at
   `thresholds.go:40-42`.)* Report every count mechanically.
7. **Are the call sites still the seventeen the inventory enumerated,
   and still all inside `cmd/dubplate-server`?** *(Expected: `Load` at
   `check.go:23`, `doctor.go:31`, `shadow.go:196`, `:262`, `:299`, and
   `check_test.go:107`; `OwnedWall`, `ReservedClasses`,
   `IntegrationClasses`, `Thresholds`, `Register`, `ParseInt`,
   `ParseDuration` at the lines the inventory names; **no package under
   `internal/` imports `internal/config`**.)* The import claim is the one
   worth re-proving mechanically, because Resolution R6 rests on it.
8. **Is `registerShadowThresholds` still a `sync.OnceFunc` called as
   `newRootCmd`'s first statement, and does
   `TestNewRootCmd_RegistersShadowThresholdsOnce` still exist?**
   *(Expected: yes, `shadow.go:49-52` (doc `:45-48`) and `main.go:56`,
   the test's func at `main_test.go:43-53` with its doc at `:36-42`, and
   `doctor_test.go:46` and `:159` as two further direct callers.)* Both
   retire in Task 4; report their current lines so the retirement is a
   named edit.
9. **Do the four tests the migration must keep green unchanged still
   assert what this plan says they assert?** `TestLoadShadowIntegration`
   (`config_test.go:255`), `TestIntegrationClassesCoverEveryShadowKey`
   (`keys_test.go:125-146`), `TestLoad_TemplateKeysClassified`
   (`config_test.go:196`), and `TestDoctorCmd` with its two literal
   spot-check siblings (`doctor_test.go:12`, `:137`, `:158`). Report each
   test's row count and its assertion targets.
10. **Does `internal/engine/matcher/tunables.go` still name the four
    knobs and still declare that it imports nothing from
    `internal/config`?** *(Expected: `:50-54`, and the four knobs at
    `:62-66`, `:68-72`, `:74-78`, `:138-147`.)* Resolution R6 rests on
    this; a change is a finding, not a licence to register them.
11. **Do `docs/HISTORY.md` and `docs/STATUS.md` still describe the state
    this plan assumes?** *(Expected: 4a closed; carries 3 and 4 at
    `HISTORY:602-609`; the four-knob and `KindFloat` items open in
    STATUS; the lane registry as item 2 describes. If 4b closed, its
    HISTORY entry and its own forward carries are new input and the task
    reports them.)*
12. **Is `bash scripts/check.sh`'s wall time still what this plan sizes
    gate runs against?** Run `go test ./... -count=1` once and report the
    total, then report `bash scripts/check.sh`'s total. **Task 0
    measured a task's first gate run, cold, at about five minutes ten
    seconds**: `go test ./...` 235 seconds, of which
    `internal/harness/conform` is 232 seconds, plus `check.sh`'s own 77
    seconds, 71 of which is gitleaks scanning 117 MB. **A later run, Go
    cache warm, is about 77 seconds.** Size later tasks' gate evidence
    against the warm figure.
13. **Where do the parallel spec's own rows sit now?** Re-read
    `2026-09-03-parallel-build-approach.md`'s Phase 0b paragraph and its
    contested-surface table as they stand in the tree, and report the
    current line numbers for five rows: the config wall, `docs/`
    ledgers, `cmd/dubplate-server`, `internal/config`, and
    `systemd`/`deploy.sh`/`env`. **Task 0 ran 2026-09-08 and reported
    the Phase 0b paragraph at `:56` and the five rows at `:175`, `:194`,
    `:199`, `:200`, `:201`, the last now also naming `Makefile`.** The
    committed-HEAD figures this item was drafted against (`:54`, `:158`,
    `:177`, `:182`, `:183`, `:184`) are superseded by that reading, and
    revision 2.1 re-points every citation in this plan to match.
14. **Do the eight spec-correction anchors still name what Task 7 will
    replace?** For each of `BS:182-183`, `PA:56` (twice, the scope
    sentence and both rider clauses), `rung-4:1476-1485`, `:1689-1692`,
    `:1697-1699` and `:1787-1788`, **quote the first six words at that
    line** and report any that has moved. Task 7 drafts against a
    checked anchor rather than a carried one. **Task 0 re-verified all
    eight at revision 2.1 and found none moved further**; the rung 4
    anchors above already carry the shift its own pass introduced.

**Deliverable count: 1** (the blocker report, each item with its evidence
and each blocker with the file and line that proves it).

**2.0 hook statement: neutral.** The task writes nothing.

**Independence:** none; every later task depends on it.

---

## Task 1: The gitleaks durable fix, its note, and the retired staleness question

Model: `sonnet`. Depends on Task 0.

**This task runs first among the code tasks deliberately.** Today every
task touching `internal/config/keys_test.go` must append below a pinned
line or re-fingerprint `.gitleaksignore` in the same commit. Retiring the
pin before the migration starts means Tasks 2 through 6 edit that file
freely, which is worth more than the ordering costs.

**Files:**
- Modify: `internal/config/keys_test.go` (**both** tables' field rename
  and the comment)
- Modify: `.gitleaksignore` (emptied, or removed, with the outcome stated
  in the report)
- Modify: `scripts/check.sh` (the gitleaks leg's comment, which today
  overstates the finding's generality)

**Outcomes:**

1. **The generic-api-key finding no longer exists**, rather than being
   suppressed. Resolution R9's Option A: the struct field whose name
   supplies the rule's keyword half is renamed, and the finding
   disappears because the rule needs both halves. **The rename covers
   both tables in the file**, `TestReservedClassOf` (`:59-92`) and
   `TestSupersessionOf` (`:103-113`). The second holds no finding today
   and would hold one the moment a high-entropy row joins it, measured;
   leaving it renamed-around is a fix that looks complete and is not.
2. **The fragility is noted where a later author reads it**, which is the
   0b rider's second requirement. A comment states why the field is not
   called `key`: a `key`-like field name on the same line as a name whose
   entropy clears the rule's threshold trips gitleaks, and the same
   discipline applies to any declaration file that writes a key literal
   beside a `Key:` field. **The comment sits where an author of either
   table reads it**, which one file-level comment can do only if both
   tables are within sight of it; otherwise each table carries its own.
3. **`.gitleaksignore` carries no line-pinned fingerprint**, so no future
   insertion anywhere in the tree can red the gate for every lane at
   once. If the file ends empty, the task states whether it removed the
   file or left it empty, and why.
4. **`scripts/check.sh`'s gitleaks comment is corrected.** It calls the
   finding "a config-key-name test fixture that happens to look like a
   generic API key", which reads as a class of names. Exactly one of the
   54 `DUBPLATE_*` names in the tree fires the rule. The corrected
   comment states what the trigger actually is and that the repo now
   carries no suppression.
5. **Rung 4's standing staleness question is answerable "retired"
   forever.** Task 7 drafts the spec correction; this task's report
   states the fact the correction rests on.

**Constraints:**

- **Never reproduce the triggering pair**, a `key`-like field name beside
  the value, anywhere outside the fixture row itself, and **never quote
  the value** in the task report or in the commit message. The rename
  keeps the row, so the diff necessarily carries the literal at its own
  line, which is expected and is not a violation. Gitleaks scans the
  tracked worktree with `--no-git` and would flag a quotation elsewhere.
- The rename is mechanical across each table's rows and its subtest body.
  No assertion changes, no row is added or removed, and both tests'
  behavior is identical.
- No `.gitleaks.toml` is created. If the implementer believes one is
  needed, it stops and reports; Resolution R9 records the measured
  over-suppression trap that makes the file the worse bargain.

**Acceptance criteria:**

- `gitleaks detect --no-git -s . --redact` reports **zero findings** with
  `.gitleaksignore` absent or empty. The task reports the command it ran
  and its output summary.
- `TestReservedClassOf` passes with the same seven rows and the same
  assertions as before the rename, and `TestSupersessionOf` with the same
  five.
- **No `key`-named struct field remains in
  `internal/config/keys_test.go`.** One grep proves it, and it is the
  criterion that distinguishes the scoped fix from the complete one.
- The comment names the trigger and the discipline, and does not name the
  value that triggers it.
- `scripts/check.sh`'s gitleaks comment no longer describes a class of
  config-key names.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Rename the field back to its original name → gitleaks reports one
   finding again, and the gate's gitleaks leg reds. This is the proof
   that the fix is the rename and not an accident of some other edit.
2. Add a high-entropy `DUBPLATE_*` row to each table in the file in turn
   → still zero findings each time, proving the fix generalizes to the
   file rather than to one row or one table. Revert each row. **The
   second half of this mutation is what revision 1 missed**: against a
   rename scoped to `TestReservedClassOf` alone, the row added to
   `TestSupersessionOf` reports one finding, measured.

**2.0 hook statement: neutral.** This task lands no forward hook. It
removes a gate fragility that every rung would otherwise inherit.

**Independence: partial.** It is the only task in the pass that touches
`.gitleaksignore` or `scripts/check.sh`, and in those two files it is
fully independent. Its third file, `internal/config/keys_test.go`, is
shared with Tasks 3 and 4, and **that overlap is the point**: Task 1 runs
first so those two inherit a file with no line pin. It is scheduled first
rather than in parallel because one tree runs one executor.

**Deliverable count: 3** (the rename with its comment, the emptied ignore
file, the corrected gate comment). Gate green; commit.

---

## Task 2: The registry type, its declaration families, and the collision gate

Model: `sonnet`. Depends on Tasks 0 and 1.

**Nothing is wired in this task.** The registry is born, tested, and left
unused. `Load`, `doctor` and the four package tables are untouched, so
every existing test passes unchanged for the trivial reason that nothing
they exercise has moved. Task 3 is the migration.

**Files:**
- Create: `internal/config/registry.go`
- Create: `internal/config/registry_test.go`

**Outcomes:**

1. **A concrete registry type**, held in a value, with a constructor
   (Resolution R1, Resolution R4). It carries four declaration families:
   owned wall rows, reserved classes, integration classes, and
   thresholds, plus the superseded-name map (Resolution R8). Each family
   has its own registration entry point and its own accessor, and each
   accessor returns a copy rather than the backing array, matching
   `OwnedWall` and `Thresholds`'s existing proven behavior.
2. **Declaration order is preserved within each family** (Resolution
   R15). The superseded map keeps its map shape, because no caller
   enumerates it and order is meaningless for a single lookup.
3. **The two class types stay distinct** (Resolution R2), each with its
   own registration entry point.
4. **The collision gate**, refusing all six classes in O5 at
   registration time by panic (Resolution R3). O5's list is
   authoritative and this outcome restates it verbatim:
   - a key declared twice within one family
   - a key declared in one family and also in another
   - a declared key without the `DUBPLATE_` prefix (Resolution R16)
   - a threshold whose declared `Default` does not parse under its
     declared `Kind`, or whose `Kind` names no parser (Resolution R11).
     **This check's own test and mutation are Task 5's**, landing with
     the kind that motivates it; Task 2 builds the check's seat in the
     gate and nothing more.
   - an integration class with an empty `Flag`, or whose `Flag` also
     appears in its own `Required` or `Optional` set
   - a wall row that is neither a prefixed key nor a status row with
     `Pending` text
5. **The panic message names the colliding key and both declaration
   sites' family names**, in the `key: reason` shape every other message
   in the package uses, so a lane reading a failed gate knows which two
   declarations to reconcile.
6. **The doc comments carry the reason, not the mechanism.** The
   registry's own doc states why a duplicate panics rather than returning
   an error, citing the three stdlib registries by name, and why
   registration is an explicit call rather than an `init` (the reason
   rung 1 already wrote at `thresholds.go:36-39`, preserved).

**Constraints:**

- No interface is defined. There is no second implementation and there
  will not be one; a concrete type is the bar.
- No functional options, no builder. Registration entry points take the
  declaration value.
- The status-row carve-out for the identity source is explicit and
  commented, because getting it wrong in either direction is silent
  (Resolution R16).
- The registry's exported surface reads like a stdlib API: short names,
  no `Get` prefix, doc comments ending in a period and starting with the
  identifier.

**Acceptance criteria:**

- A registry constructed and left empty of later declarations carries
  exactly rung 1's own declarations, and its four accessors return them
  in declaration order. *(This is what makes Resolution R4's empty-wall
  hazard impossible.)*
- Each accessor returns a copy: mutating the returned slice does not
  change what a second call returns.
- **Each of the six collision classes in outcome 4 has one test that
  fires it**, and each test asserts the panic's message names the
  offending key. Five of the six also carry their mutation here; the
  kind class's mutation is Task 5's mutation 3, so the
  one-mutation-per-class promise is auditable across the two tasks and
  nowhere else.
- A registry built twice from the constructor produces two independent
  registries: registering into one does not change the other. **This is
  the property that retires the `sync.OnceFunc`**, and Task 4 depends on
  it.
- Registering the same declaration into two separately-constructed
  registries panics in neither.
- No existing test in the module changes, and none needs to: nothing is
  wired.
- `bash scripts/check.sh` green, vale included on the new file.

**Named mutations this task reports:**

1. Remove the duplicate check from one family's registration entry point
   → that family's duplicate test fails. Repeat per family and report the
   count, because a check applied to three of four families is the defect
   this pass exists to remove.
2. Remove the cross-family check → the cross-family test fails while
   every within-family test still passes, proving the two checks are
   independent rather than one check counted twice.
3. Remove the `DUBPLATE_` prefix check → the prefix test fails.
4. Make an accessor return the backing slice instead of a copy → the
   copy test fails. *(This mirrors `TestOwnedWall_ReturnsCopy` and
   `TestThresholds_ReturnsCopy`, whose existing proof this task must not
   regress.)*
5. Make the constructor return a shared registry rather than a fresh one
   → the two-independent-registries test fails. **Report this mutation's
   outcome explicitly**, because Task 4's retirement of the
   `sync.OnceFunc` rests on it and nothing else.
6. Let a status row through the prefix check as a key → the
   identity-source carve-out test fails, reporting that `identity source`
   registered as a key.
7. Remove the integration-class `Flag` check → both the empty-`Flag`
   test and the `Flag`-in-`Required` test fail. Revision 1 left this
   class the only one in O5 with a required test and no mutation, which
   is the shape the pass's strictest constraint exists to refuse.

**2.0 hook statement: lands hook 6 (the driver seam stays universal,
rung 12), as an enabling shape.** The hook requires a driver to declare
its own key set and to decline capabilities it does not have. A registry
whose integration and reserved families are registered by the owning
package is what makes that reachable; the current fixed table, which
hard-codes three Navidrome names into the `streamer` class, cannot
express a driver that declines. The task builds no 2.0 feature and no
driver: it builds the seam the hook needs and stops there.

**Independence: no.** It creates the type Tasks 3, 4, 5 and 6 all use.

**Deliverable count: 2** (the registry with its four families and the
superseded map; the collision gate with its six refused classes). Gate
green; commit.

---

## Task 3: Rung 1's declarations migrate onto the registry, and `Load` and doctor read it

Model: `sonnet`. Depends on Task 2.

**This is the migration task, and its whole proof is that nothing
changed.** Every existing test in `internal/config` and
`cmd/dubplate-server` passes **unchanged**: no assertion weakened, no
test deleted, no `t.Skip` added, no expectation edited to make a
migration pass. A task that finds it must edit an existing assertion
stops and reports.

**Files:**
- Modify: `internal/config/keys.go` (the four tables become rung 1's
  declarations, registered through the registry's constructor)
- Modify: `internal/config/config.go` (`Load` takes the registry;
  `rejectNamespace` and the class lookups read it; the package doc's
  contract paragraph is amended per "The contract change")
- Modify: `internal/config/thresholds.go` (`Register` and `Thresholds`
  become registry operations; the package-level registry variable and its
  test helper retire)
- Modify: `internal/config/{config,keys,thresholds}_test.go` (only where
  a signature forces it, and the task reports every such edit)
- Modify: `cmd/dubplate-server/main.go` (build the production registry
  once in `newRootCmd` and pass it down)
- Modify: `cmd/dubplate-server/{check,doctor,shadow}.go` and
  `check_test.go` (take the registry)
- Modify: `cmd/dubplate-server/doctor_test.go` (only where a signature
  forces it)

**Outcomes:**

1. **`internal/config/keys.go` carries rung 1's own declarations, and
   the reserved classes it holds in trust** (Resolution R4): the four
   owned-now rows, the identity-source status row, the eight reserved
   classes with their **29** keys in table order, and the eleven
   superseded names. Rung 4's integration class leaves in Task 4, and
   each reserved class leaves when its own package is born.
2. **`Load` takes the registry it enforces against** (Resolution R1), and
   reads no package-level table. Its six `cmd/dubplate-server` call sites
   and `internal/config`'s own two test call sites are updated.
3. **`newRootCmd` builds the production registry once**, visibly, in the
   caller's own code, and passes it to `newCheckCmd`, `newDoctorCmd` and
   `newShadowCmd`. Registration order stays visible in the caller's own
   code, which is the rule rung 1 wrote and this pass keeps
   (`thresholds.go:36-39`).
4. **Doctor's four blocks read the registry** and keep their order, their
   `%-32s %s` column pin, and every status string verbatim.
5. **The package-level `registeredThresholds` variable retires**, and
   with it `withRegistry`'s save-and-restore helper in
   `thresholds_test.go`. A test that needs a registry constructs one.
6. **The package doc's contract paragraph is amended** to state what the
   contract change section states: the five classes stay, the closed
   table becomes registry contents, and the "separate, open-ended
   registry" sentence retires because there is now one registry.

**Constraints:**

- **`wall.get` stays the single choke point.** Every accessor and both
  threshold parsers still route through it. The packaging initiative's
  later file overlay must stay a one-function change.
- **Every error message keeps its exact wording and its `key: reason`
  shape.** A reserved name still errors naming its class and rung; a
  superseded name still errors naming its replacement or retirement; an
  unknown name still errors as unknown.
- **Reserved class table order is preserved exactly**, including that
  `decision and review queues` (rung 13) precedes `streamer` (rung 12).
- `ParseInt` and `ParseDuration` keep their signatures. They take a
  `Threshold` and an `environ` and need no registry.
- No behavior change of any kind. The three behavior changes this pass
  makes are Tasks 4 and 5's, each named in its own task.

**Acceptance criteria:**

- **Every test in `internal/config` and `cmd/dubplate-server` passes
  unchanged**, with the sole exception of edits a changed signature
  forces. The task reports each forced edit with its file, its line, and
  one sentence proving it is a signature edit rather than an assertion
  edit.
- `TestDoctorCmd`, `TestDoctorCmd_LiteralSpotCheck` and
  `TestDoctorCmd_ShadowRows` pass with no expectation changed. The
  literal spot checks are the antidote to a structural test that derives
  its expectations from the same tables it reads, and they are what
  proves the migration preserved output rather than preserved a tautology.
- **`TestDoctorCmd_LiteralSpotCheck` gains one order assertion**: the
  `decision and review queues` line's index in stdout is less than the
  `streamer` line's, with a comment naming R15 and the fact that rung 13
  precedes rung 12 in table order. Nothing in the module pins that order
  today. `TestDoctorCmd` cannot, because it builds its expected reserved
  lines by iterating `config.ReservedClasses()` itself
  (`doctor_test.go:89-98`), so a sorting registry sorts both sides and
  the test still passes. The spot checks use `strings.Contains` and pin
  presence alone. **This assertion is what makes R15 a guarded property
  rather than a stated intention**, and it is the target of mutation 1.
- `TestLoad_TemplateKeysClassified` passes with its classification now
  routed through the registry.
- `TestOwnedWall_ReturnsCopy` and `TestThresholds_ReturnsCopy` pass, the
  copy property preserved through the accessor move.
- **The migrated reserved family carries 29 keys across 8 classes**, and
  any count change is a stop-and-report rather than a reconciliation.
- **`registeredThresholds` no longer exists in `internal/config`, and
  `withRegistry` no longer exists in `thresholds_test.go`.** Named by
  symbol, because "no package-level mutable registry variable" reads two
  ways: rung 1's four declaration literals (`ownedWall` at `keys.go:19`,
  `reservedClasses` at `:57`, `supersededKeys` at `:164`,
  `integrationClasses` at `:214`) stay package-level by R4 and are read
  once by the constructor. They are intended, and they are not what this
  criterion refuses.
- `internal/config/keys.go` contains no rung-4 name. *(It still contains
  the integration class at the end of this task; that leaves in Task 4.
  The criterion is that no rung 5 through 17 name appears anywhere it did
  not already appear.)*
- **No package under `internal/` imports `internal/config`.** Re-proved
  mechanically, because Resolution R6 rests on it.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Sort the reserved classes by rung inside the registry → **the new
   order assertion in `TestDoctorCmd_LiteralSpotCheck` fails**, reporting
   `streamer` ahead of `decision and review queues`. This is Resolution
   R15's proof. `TestDoctorCmd` is order-blind by construction and is
   expected to pass under this mutation; the task reports both outcomes,
   because "the structural test passed" is the observation that would
   otherwise be mistaken for the property holding.
2. Drop the identity-source row's `Pending` handling so `ownedNow`
   returns true for it → `TestDoctorCmd_LiteralSpotCheck` fails on the
   pending text, and a `Load` with the wall otherwise satisfied starts
   requiring a key named `identity source`.
3. Change one reserved-class error's wording to drop the rung number →
   `TestLoad_ContractTable`'s "reserved key early" row fails, which
   asserts both `"backup targets"` and `"rung 6"`
   (`config_test.go:89-94`). Doctor's own rows are unaffected: reserved
   classes print fixed status text and consult no statuses map, so the
   `key: reason` shape is load-bearing for owned rows and threshold
   parse errors alone.
4. Route one accessor around `wall.get`, reading the environment map
   directly → the empty-value-equals-unset test fails. *(If no existing
   test covers this for the accessor chosen, the task reports that as a
   finding rather than adding one: an uncovered choke-point bypass is a
   real gap and Task 7 records it.)*
5. Register rung 4's thresholds onto a first registry and pass a second,
   freshly constructed one to `newDoctorCmd` → `TestDoctorCmd_ShadowRows`
   fails on both threshold lines, because the second registry carries
   rung 1's rows and not rung 4's. That is the proof the wiring is a
   value pass rather than a hidden global: under a package-level registry
   the rows would print either way. Revert. *(Revision 1 phrased this as
   "build the registry twice and pass the second, and every row still
   prints", which is what a global would do and what a value registry
   would not.)* Task 2's mutation 5 carries the freshness half of the
   same property.

**2.0 hook statement: neutral, with one weak touch on hook 5 (a stable
identity, rungs 5 and 6).** The identity-source status row is the wall's
one pending owned row, and the required-when-selected block that rung 5
eventually attaches to it will be declared through this registry. The
discriminator is the package doc's own word for that block's selector
(`config.go:11-17`) rather than anything hook 5 states; hook 5 is about
rung 6 minting a surrogate user id and rung 5 resolving a roster identity
to it. This task preserves the row's semantics exactly and invents no
selector shape, because rung 5 has not specified one.

**Independence: no.** Every later task builds on the migrated package.

**Deliverable count: 4** (rung 1's declarations on the registry; `Load`
taking the registry with its eight call sites; doctor reading the registry
with its four blocks intact; the package-level registry variable and its
test helper retired). Gate green; commit. **Checkpoint: write STATUS
after this task.**

---

## Task 4: The integration family absorbed as data, and `Load`'s shadow branch retired

Model: `sonnet`. Depends on Task 3.

This task discharges HISTORY carry 4 and Resolution R5, and it is where
rung 4's ruling 1 is re-expressed in registry terms (Resolution R14).

**Files:**
- Modify: `internal/config/keys.go` (the integration class table leaves;
  `IntegrationClass` and its lookup stay as registry types)
- Modify: `internal/config/config.go` (`Load`'s integration loop becomes
  data-driven; the `class.Name == "shadow"` branch and
  `ShadowIntegration` leave)
- Modify: `cmd/dubplate-server/shadow.go` (rung 4's declaration file: the
  shadow integration class, its six rows, its two thresholds, and the
  shadow-specific type built from the resolved values)
- Modify: `cmd/dubplate-server/main.go` (the `sync.OnceFunc` call site
  retires)
- Modify: `cmd/dubplate-server/main_test.go` (the double-registration
  regression test retires, its property now held by Task 2's
  two-independent-registries test)
- Modify: `internal/config/{config,keys}_test.go`,
  `cmd/dubplate-server/{shadow,doctor}_test.go` as the move requires

**Outcomes:**

1. **`Load` resolves every enabled integration class's `Required` keys
   generically** and makes them reachable by class name. No class name
   appears in `Load`. A second integration class needs no second branch.
   **`Load` records a class's enabled state independently of whether it
   resolved any key.** A class with an empty `Required` set is otherwise
   indistinguishable from a disabled one, because the resolved map is
   empty either way, and today `cfg.Shadow == nil` is what
   `cmd/dubplate-server/shadow.go:199` reads to refuse `observe`. The
   empty-`Required` class is hook 6's motivating case rather than a
   hypothetical: a driver declining address and proxy declares no
   required key and must still report enabled.
2. **Rung 4's declarations live in rung 4's own file.** The shadow class,
   its flag, its three required paths, its two optional thresholds, and
   the type that carries the resolved paths all sit in
   `cmd/dubplate-server/shadow.go`, registered through the production
   registry from `newRootCmd`.
3. **`internal/config`'s non-test source names no rung-4 concept**, and
   its tests name one only through a fixture class the test itself
   declares. No `Shadow` field, no `ShadowIntegration` type, no rung-4
   name in `config.go`, `keys.go` or `thresholds.go`. This is the
   universality argument in Resolution R5.
4. **`TestLoadShadowIntegration` splits in two, per R5.** The eight
   rung-4 rows move verbatim to `cmd/dubplate-server/shadow_test.go`,
   where the shadow class now lives, and a generic
   required-when-enabled table test stays in `internal/config` driven by
   a fixture integration class with no rung-4 name. The task reports both
   halves: the moved rows with the file they landed in, and the fixture
   class the new test declares. **Revision 1 asked for both a verbatim
   survival inside `internal/config` and a package with no shadow
   string**, which the tree makes contradictory: the three test files
   carry 41 case-insensitive `shadow` hits, 33 of them in
   `config_test.go`.
5. **The `sync.OnceFunc` and its regression test retire.** They existed
   only because `Register` appended without dedup and `newRootCmd` runs
   more than once in a test binary. Task 2's constructor makes each
   construction independent, so the workaround has no remaining job. The
   retirement is stated in the report with the property that replaced it.
   `doctor_test.go`'s own two direct callers of `registerShadowThresholds`
   (`:46`, `:159`), which call it without going through `newRootCmd`,
   keep working under the retired shape for the same reason: each call
   registers into its own independent construction rather than a shared
   global.
6. **The `Required`/`Optional` split's semantics survive verbatim**
   (HISTORY carry 4): `Load` requires every `Required` key when the flag
   is set; an `Optional` key is accepted and never required.

**Constraints:**

- **`TestLoadShadowIntegration`'s eight rows and every `wantErr`
  assertion survive verbatim, in their new file.** Only the two
  assertions that reach into `cfg.Shadow` (`config_test.go:330`, `:333`)
  are reworded to reach the resolved values by class name. The task
  reports the two reworded lines and states that no `wantErr` changed.
  **A reworded assertion that no longer fires is a defect**, and
  mutation 1 below is what proves it still fires. Moving the table
  between files is a relocation rather than a rewrite, and the task's
  report shows the moved rows are byte-identical apart from those two
  lines.
- `TestIntegrationClassesCoverEveryShadowKey` survives, retargeted at the
  registry's integration family. Its job is that no shadow key is
  declared without appearing in the class, and that job is unchanged.
- The flag's self-report guard survives: the flag names itself, so
  `rejectNamespace` skips it and one typo produces one error rather than
  two contradictory ones.
- An unparseable flag value stays an error rather than a silent off.
- The `Enabled` reporting path stays: doctor renders a parse error as
  `disabled` rather than crashing.

**Acceptance criteria:**

- **`grep -ri shadow internal/config/{config,keys,thresholds}.go`
  returns nothing** outside a comment naming rung 4 as an example, and
  the task reports the exact remaining hits. The package's tests may
  name a fixture integration class they themselves declare, and no
  rung-4 name; the task reports that file's remaining hits separately.
- The eight rung-4 rows are present in
  `cmd/dubplate-server/shadow_test.go` and absent from
  `internal/config/config_test.go`, and the generic
  required-when-enabled test in `internal/config` passes against its own
  fixture class.
- **A class with an empty `Required` set loads clean with its flag set
  and reports enabled, and reports not enabled with its flag unset.**
  One test row, named as hook 6's shape so rung 12 reads it rather than
  rediscovering it.
- A set integration key with its flag unset is still a load error
  **naming the flag**, not "unknown configuration key".
- With the flag set, every unset `Required` key is reported, all at once.
- An `Optional` key set with the flag on loads clean and is never
  required.
- An unparseable flag value produces **exactly one** error for that key.
- `DUBPLATE_SHADOW_ENABLED=false` alone loads clean with no resolved
  shadow values, and set beside a shadow key it errors naming the flag.
- Doctor prints the shadow class with its flag state and both thresholds
  with their effective values, the duration one as a duration, in the
  same block position and the same 32-column form.
- Constructing `newRootCmd` twice registers cleanly and panics in
  neither, with no `sync.OnceFunc` anywhere in the package.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Make the resolved-values lookup return the values for **every** class
   rather than only enabled ones → `TestLoadShadowIntegration`'s
   flag-unset row fails. **This is the proof the reworded assertions
   still fire**, and it is the mutation that would otherwise go unproven
   when an assertion is reworded during a move.
2. Make `rejectNamespace` accept every integration key unconditionally →
   the "a shadow key set with the flag unset" row fails. This is
   ruling 1's own stated failure mode: a half-configured integration
   running unattended and observing nothing.
3. Move the two threshold names from `Optional` into `Required` → the
   flag-set-with-every-path row fails, reporting both thresholds missing,
   while the row that sets one threshold still passes. That is the
   contradiction the `Required`/`Optional` split refuses, made visible.
4. Read an unparseable flag as off → the unparseable-flag row fails.
5. Drop the `name == class.Flag` guard from `rejectNamespace` → both
   explicit-false rows fail, the first with a load error where a clean
   load is wanted.
6. Restore a class-name branch in `Load` and give it a second class with
   no branch → the second class's resolved values are empty. **Report
   this mutation's outcome as the proof that outcome 1 landed**, and
   revert.
7. Make the enabled state derive from whether any key resolved → the
   empty-`Required` row fails, reporting a flag-set class as not
   enabled. This is the hook 6 shape, and it is what stops rung 12 from
   inheriting a seam that cannot express a driver with no keys.

**2.0 hook statement: lands hook 6 (rung 12's universal driver seam).**
This is where the seam becomes usable rather than merely possible: an
integration class declared by its owning package, with `Load` resolving
it generically, is exactly the shape a streamer driver needs to declare
its own keys and decline the capabilities it does not have. The task
builds no driver and moves no Navidrome key; the `streamer` reserved
class stays where it is until rung 12 owns it.

**Independence: no.** It depends on Task 3's migration.

**Deliverable count: 4** (the data-driven integration resolution with
its enabled-state rule; rung 4's declarations moved to their own file;
`internal/config` cleared of rung-4 concepts, which is where the
`TestLoadShadowIntegration` split lands; the `sync.OnceFunc` and its
regression test retired). Gate green; commit.

---

## Task 5: `KindFloat`, `KindString`, their parsers, and the exhaustive kind check

Model: `sonnet`. Depends on Tasks 2 and 3. Task 3 retires
`registeredThresholds` and rewires doctor's four blocks onto the
registry, and both of this task's files depend on that shape: a
`KindFloat` parser landed on the pre-migration package-level registry
would sit on a variable Task 3 then deletes.

This is the mechanism half of HISTORY carry 3 (Resolution R6). The four
knobs are not registered here and `internal/engine/matcher` is not
touched.

**`KindString` joins this task by conductor ruling, 2026-09-07.** The
Phase 0c design-language standard
(`docs/specs/2026-09-07-design-language-standard.md`, its
R3.3 ruling) needs two string-valued theme keys, and the conductor ruled
that the kind those keys will need is 1b's work, on the same rule that
put `KindFloat` here: a threshold kind is cheap and safe to add once,
behind the registration-time `Default` check, and adding it here means
rung 7 registers its two keys against a kind that already exists rather
than building the mechanism itself. **The two theme keys themselves stay
rung 7's to declare**; this task adds no config row and touches no
template.

**Files:**
- Modify: `internal/config/thresholds.go` (`KindFloat` and its parser)
- Modify: `internal/config/thresholds_test.go`
- Modify: `internal/config/registry.go` (the kind check joins the
  collision gate)
- Modify: `cmd/dubplate-server/doctor.go` (the kind switch)
- Modify: `cmd/dubplate-server/doctor_test.go` (appended, not rewritten)

**Outcomes:**

1. **A third threshold kind and its parser**, in the exact shape the two
   existing parsers share: the default's own parse when the key is unset,
   the environment's value when set, an error naming the key and its
   `Doc` on a bad value, and an error naming the key alone on a bad
   default, which is a registration bug rather than an environment one.
2. **`KindInt` stays the zero value** (Resolution R11), so every existing
   `Threshold` literal keeps its meaning and no existing declaration is
   edited.
3. **Registration parses each declaration's own `Default` under its
   declared `Kind`**, and refuses the declaration when it does not
   parse, through the collision gate seat Task 2 built. **This is the
   check that catches the hazard, and "refuse an unnamed kind" is not**:
   `KindInt` is `iota`'s zero (`thresholds.go:15-18`), so a float
   threshold whose `Kind` was never set carries a named kind and passes
   any enum-membership test. `Threshold{Key: X, Default: "0.04"}` with
   the zero kind must fail registration with `X: invalid default "0.04"`.
   An out-of-range kind is refused by the same check, and `KindString`'s
   own default check (outcome 5) sits in the same collision gate seat, so
   a later kind extends one seat rather than two. Doctor's kind
   switch names every kind explicitly and its remaining fallthrough
   reports a registration bug in its own words rather than parsing an
   unknown kind as an integer. This is inventory surprise 11, fixed at
   the right altitude: a mis-declared float threshold must not report as
   an environment parse failure, which is precisely what
   `thresholdStatus`'s `default` arm does today
   (`doctor.go:112-127`) for rung 3's 0.04 default.
4. **The float's rendering in doctor is decided and stated.** A float
   threshold's effective value prints in a form that round-trips through
   the same parser, and the task states the formatting choice and its
   reason in the code's own comment.
5. **A fourth threshold kind, `KindString`, and its parser**, routed
   through `wall.get` like the other three: the default's own value when
   the key is unset, the environment's raw string when set. The
   registration-time check refuses an empty `Default`, naming the key
   alone, because `wall.get`'s empty-equals-unset convention would make
   an empty string default indistinguishable from no default at all, and
   a string kind's only invalid default is that one. `KindString`
   registers no key of its own; it exists so a later rung's
   string-valued declaration has a kind to declare against.

**Constraints:**

- No knob is registered. `internal/engine/matcher` is not in this pass's
  package territory, and `tunables.go:50-54` declines the
  `internal/config` import in writing.
- Both new parsers route through `wall.get` like the existing two, so
  empty-equals-unset and last-wins are identical across all four.
- `Threshold` stays comparable with `!=`, because `thresholds_test.go`
  compares `Threshold` values directly. A kind that made the struct
  incomparable would break a test this pass may not weaken.

**Acceptance criteria:**

- **One test-row table, one row per kind.** `KindInt`, `KindDuration`,
  `KindFloat` and `KindString` each get a row for the parser's cases:
  unset with a valid default, set with a valid value, set with an
  invalid value naming the key and its `Doc` (n/a for `KindString`,
  which cannot fail to parse a set value), and unset with an invalid
  default naming the key alone (`KindString`'s invalid default is the
  empty string). The last cell in each row is distinguishable from the
  third, which is the registration-bug-versus-environment-bug line the
  other parsers already hold.
- **A `Threshold` carrying a float default and no explicit `Kind` is
  refused at registration**, with a message naming the key and quoting
  the default. Rung 3's `StrongRecThresh` (default 0.04,
  `tunables.go:62-66`) is the fixture to write it against, because it is
  the real declaration the next rung 3 pass will register. A threshold
  whose `Kind` names no parser at all is refused by the same check.
- **A `Threshold` declared `KindString` with an empty `Default` is
  refused at registration**, with a message naming the key alone. A
  scratch `Threshold{Key: "TEST_STRING", Kind: KindString, Default: ""}`
  is the fixture, since no real declaration exists yet.
- **Doctor never reports a mis-declared threshold as an environment
  problem.** With the registration check bypassed, the float-defaulted
  integer threshold would print `invalid: invalid default "0.04"`; with
  it in place, the declaration never reaches doctor.
- Doctor prints a registered float threshold with its effective value,
  in the same 32-column form and in the thresholds block.
- Doctor's kind switch has no branch that parses an unknown kind as an
  integer. A grep for the old fallthrough shape returns nothing.
- Every existing threshold test passes unchanged, and `KindInt` is still
  the zero value.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Make the float parser report a bad default with the same message as a
   bad value → the fourth test row fails, because the two are asserted to
   differ. This is the distinction the registry already draws and the one
   surprise 11 says a naive `KindFloat` would lose.
2. Remove the default validation from the registration check, keeping
   the enum-membership half → the float-defaulted threshold with no
   explicit `Kind` registers clean, and doctor prints
   `invalid: invalid default "0.04"` on its row. **Report both halves of
   that outcome**, because the enum check passing is what makes the
   defect invisible, and it is what revision 1's design would have
   shipped.
3. Remove the kind check from the collision gate entirely → the
   registration test fails, and both the unnamed-kind threshold and the
   unparseable-default threshold register silently. This is O5's
   fourth collision class's mutation, which lives here rather than in
   Task 2.
4. Change `KindFloat`'s value so `KindInt` is no longer zero → every
   existing `Threshold` literal that omits `Kind` changes meaning, and
   the existing integer threshold tests fail. **Report this mutation's
   outcome**, because it is what proves Resolution R11's zero-value
   choice is load-bearing rather than incidental.
5. Remove the empty-default check from `KindString`'s branch of the
   registration gate → the scratch `KindString` threshold with an empty
   `Default` registers clean, and a later `wall.get` call on it reads as
   unset rather than as the declared default. **Report this mutation's
   outcome**, because it is the proof that a mis-declared string default
   fails registration rather than silently losing its default at read
   time.

**2.0 hook statement: weakly touches hooks 12 and 14, and lands
neither.** Hook 12 (rung 7's portable layout) wants a
forbidden-character set; its 255-byte component bound is an integer,
which today's kinds already express, so only the character set is a
shape no kind holds. Hook 14 (rung 8's ID3 version ruling) wants a
two-valued choice. Neither is an int, a duration, or a float, so neither
is landed here. What this task lands is the evidence
that a kind can be added cheaply and safely, plus the registration-time
check that makes adding one safe. **`KindString` itself is now landed
here** by conductor ruling, since a string kind is what the Phase 0c
design-language standard's two theme keys need; the keys themselves stay
rung 7's to declare against it. An enumerated kind stays the owning
rung's to add, and Task 7 records that.

**Independence: yes, of Task 4.** Its files are disjoint from Task 4's
except for `doctor_test.go`, which both may touch; the rule is that this
task **appends** its own test function rather than editing an existing
one. It still runs sequentially, because one tree runs one executor and
the module gate runs over `./...` regardless. The conductor may reorder
it **before Task 4 and after Task 3** if Task 4 stalls, never earlier.

**Deliverable count: 4** (`KindFloat` and its parser; `KindString` and
its parser; the registration-time check that parses each declaration's
own default under its declared kind; doctor's exhaustive switch). Gate
green; commit.

---

## Task 6: The doctor and template sync mechanism

Model: `sonnet`. Depends on Tasks 3, 4 and 5.

Two derivability questions, answered with two different mechanisms
because they are different problems.

**Files:**
- Modify: `cmd/dubplate-server/doctor_test.go` (the exhaustiveness
  assertion)
- Modify: `internal/config/config_test.go` (the template coupling test
  and its documented rule)

**Outcomes:**

1. **Doctor's derivation is proven in the one direction that does not
   already exist** (O6). **The line-count assertion is already landed**
   at `cmd/dubplate-server/doctor_test.go:127-129`, where `TestDoctorCmd`
   walks all four families incrementing `li` and then reports any extra
   line. This task's job there is that it survives the migration
   unchanged, not that it is written. What this task builds is the
   scratch-registry direction, which only a value registry makes
   reachable: a declaration added to a scratch registry inside the test
   adds exactly one line, in its own family's block. **The unit is one
   line per declaration**, which is one per owned wall row, one per
   reserved class regardless of key count, one per integration class,
   and one per threshold.
2. **The template coupling states both its directions** (Resolution
   R10), against both templates now that Task 0 confirms the greenfield
   one, `env/dubplate-server.env.template`, landed at `c8b6d71`:
   - The classification direction stays asserted, now against the
     registry: every `DUBPLATE_*` name either template assigns
     classifies as owned, reserved, superseded, **or as an integration
     class's flag, `Required` key or `Optional` key**, never unknown.
     The fourth branch is what the greenfield template's six
     `DUBPLATE_SHADOW_*` rows require; without it the test reports six
     names unknown and the gate reds for every lane. Because rung 4's
     class is declared in `cmd/dubplate-server` after Task 4, the test
     classifies against the registry it is handed rather than against
     package tables, and the task states which registry that is.
   - **The completeness direction is now an assertion, not a
     conditional.** The greenfield template carries all four owned-now
     rows (`DUBPLATE_LIBRARY_ROOT:36`, `DUBPLATE_STATE_DIR:53`,
     `DUBPLATE_SITE_NAME:61`, `DUBPLATE_ADMIN_EMAIL:66`), so the test
     asserts every owned-now row appears in it. **The "documented rule"
     alternative retires**: a rung that adds an owned-now row without
     adding it to the greenfield template now fails a test rather than a
     review.
3. **The rule that cannot be mechanical is written where it will be
   read**: the test's own doc comment states that a rung declaring an
   owned-now row adds it to the greenfield template in the same pass,
   with the minting that implies, and names `scripts/deploy.sh`'s
   manifest behavior as the reason: both templates render
   (`ENV_TEMPLATE` at `:27`, `SERVER_ENV_TEMPLATE` at `:28`,
   `render_template` at `:205`, called at `:240` and `:241`), so a name
   missing from either fails deploy the same way.
4. **`templateKeys` generalizes to take a template path rather than a
   hard-coded one.** Today it opens `../../env/dubplate.env.template`
   unconditionally (`config_test.go:221`), so no test reads the
   greenfield template. Both directions call the generalized helper: the
   classification direction walks both templates in `env/`, and the
   completeness direction walks the greenfield one by name, each with
   its reason stated rather than left as a path a later reader has to
   interpret.

**Constraints:**

- **No name is added to any template.** Adding one requires minting its
  value and is an owner touchpoint; the completeness assertion reads the
  greenfield template and never writes it.
- The classification test keeps its existing guard that finding zero
  names is itself a failure, which is what catches a moved template path
  or a changed format.
- Doctor's exhaustiveness assertion derives its expectation from the
  registry, and the two literal spot-check tests stay as the antidote to
  a purely structural assertion. **A structural test alone is a
  tautology**, and the spot checks are what make the pair meaningful.

**Acceptance criteria:**

- Doctor's line count equals the registry's total **declaration** count
  across all four families, computed in the test from the registry: one
  per owned wall row, one per reserved class, one per integration class,
  one per threshold. The existing assertion at `doctor_test.go:127-129`
  still passes, and the task states that it was preserved rather than
  rewritten.
- Adding a declaration to a scratch registry inside the test adds exactly
  one doctor line, in the group's own block. **This is the task's new
  work.**
- The classification test passes against every template present in
  `env/`, including any greenfield template carrying `DUBPLATE_SHADOW_*`
  rows, and reports which files it read.
- The completeness direction passes against `env/dubplate-server.env.template`,
  the greenfield template Task 0 found carrying all four owned-now rows.
- The documented rule appears in the test's doc comment and names
  `deploy.sh`'s manifest behavior.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Remove the new declaration's family from the scratch registry's
   accessor → the scratch test fails while the landed
   `doctor_test.go:127-129` count assertion still passes. **Report both
   outcomes**, because the pair is what shows the new test covers ground
   the existing one does not. *(Revision 1's mutation here, adding a row
   to `doctor.go` that no registry entry produces, fires today against
   the shipped `TestDoctorCmd` with no new test written, so it could not
   distinguish this task's work from what already ships.)*
2. Remove one family's block from doctor entirely → the landed count
   assertion fails, and the count difference names the family.
3. Point the classification test at an existing file that assigns no
   `DUBPLATE_*` names, such as a scratch temp file or `go.mod` → the
   zero-names guard fires with "template path or format changed".
   **Report this mutation's outcome**, because a template test that
   passes when it reads nothing is the exact silent-success shape
   assertion discipline forbids. *(A nonexistent path does not prove the
   guard: `templateKeys` calls `t.Fatalf("open template: %v", err)` at
   `config_test.go:221-223`, aborting before the guard at `:197-199` is
   reached.)*
4. Point the classification test at a nonexistent path → it fails at the
   open with "open template", proving that a moved template is fatal
   rather than skipped. This is a separate property from mutation 3 and
   is reported separately.
5. Remove `DUBPLATE_LIBRARY_ROOT` from the greenfield template read by
   the test → the completeness test fails naming that row. Revert
   without committing the template edit. **The row is
   `DUBPLATE_LIBRARY_ROOT`, not `DUBPLATE_STATE_DIR`**: removing
   `DUBPLATE_STATE_DIR` also reds 4b's own state-dir agreement leg
   (`scripts/check.sh:150-166`), which reads that name from the same
   template, so that row's mutation would prove two gate failures at
   once instead of isolating this task's own assertion.
   `DUBPLATE_LIBRARY_ROOT` is read by no other gate leg.

**2.0 hook statement: neutral.** No forward hook reads a doctor row or a
template. The task's value to 2.0 is indirect: a derived doctor is what
lets rung 12's driver declare keys without editing a command file.

**Independence: no.** It asserts against the registry Tasks 3 through 5
populate.

**Deliverable count: 2** (the scratch-registry derivation assertion,
alongside the landed count assertion it preserves; the template coupling
with its four-way classification, its two directions, and its documented
rule). Gate green; commit.

---

## Task 7: The pass close, the contract change recorded, and the spec corrections

Model: `sonnet`. Depends on Tasks 0 through 6.

**Files:**
- Modify: `docs/STATUS.md`, `docs/HISTORY.md` (drafted for the
  conductor's batch commit, never committed by this task)
- Draft, handed to the conductor in the task return: the eight spec
  corrections, the filed notes, and the forward carries

**Outcomes:**

1. **The HISTORY entry in the house shape**: what landed, what the gate
   caught, and **what a later pass would be wrong to rediscover**.
2. **The contract change recorded as a contract change**, not as a
   refactor. The entry carries the "What rung 1's contract says today",
   "What 1b changes" and "What stays, exactly" content of this plan's
   contract section, condensed, so a rung 6 or rung 12 author reads the
   contract as it now stands without opening this plan.
3. **The eight spec corrections drafted**, one line each for the
   conductor's batch, exactly as listed in "The eight spec corrections
   Task 7 drafts" above: the build spec's complete-table sentence; the
   parallel spec's 0b scope sentence; the parallel spec's gitleaks
   blast-radius sentence; the rung 4 spec's "Has rung 1b landed?"
   ordering assumption; the rung 4 spec's standing `.gitleaksignore`
   staleness question, now retired permanently; the rung 4 spec's full
   fragility restatement at `:1476-1485`; its risk 7 at `:1787-1788`;
   and the 0b rider's re-fingerprint clause at `PA:56`. **Each is
   re-read at its line before its replacement is written**, and any that
   has moved is reported with its new line, on Task 0 item 14's
   evidence.
4. **The sixteen resolutions listed with their inline markers**, each
   naming the file and the symbol the marker sits on, so a conductor
   ruling that goes the other way is a search rather than a re-reading.
   The markers themselves are left by Tasks 1 through 6 under the Global
   Constraint that requires them, and each task reports its own; this
   outcome assembles the task returns rather than re-deriving them from
   the tree.
5. **The declared coverage gaps recorded as gaps**, on one criterion: a
   surface 1b itself builds that no test in this pass can reach. The
   known candidates, to be confirmed or corrected against what actually
   landed:
   - **The registry's panic is unreachable in the deployed binary as
     shipped**, because the gate-resident test catches every collision
     before a commit lands. That is the design working, and it means the
     panic path itself is proven only by tests that recover it.
   - **`wall.get`'s choke-point property has no mechanical guard.** A
     future accessor that reads the environment map directly would pass
     every test unless that accessor's own test covers the empty-value
     case. Task 3's mutation 4 reports whether a gap exists.
   - **Pattern key families and `DUBPLATE_COLLECTIONS`** (Resolution
     R12), declined and filed to rung 9. `DUBPLATE_COLLECTIONS` is the
     sharper one: an adopter setting it today gets "unknown configuration
     key" while CLAUDE.md still lists collection routing as live
     behavior.
   - **The three reserved classes the build spec names and the tree
     lacks** (Resolution R13), recorded for rungs 5 and 12.
   - Each gap is **joined to the running total of thirty-four** (rung
     3's twenty-one, rung 4a's nine, rung 4b's four, `HISTORY:158-163`)
     rather than re-counted. 4b's own HISTORY entry adds its declared
     gaps to that total; this task folds its own gaps into the same
     figure rather than re-counting what 4b already added.
6. **The evidence, not a bare "green"**: every task's measured wall time
   for `go test ./... -count=1` and `bash scripts/check.sh`; the
   mutation table with every named mutation across Tasks 1 through 6 and
   its measured outcome, including any deliberately unfired one with the
   guard that holds its property instead; and the gitleaks scan's
   zero-finding result.
7. **The forward carries, each stated so the next pass reads rather than
   derives:**
   1. **The four-knob wall registration is now unblocked** and is the
      named task for the first rung 3 pass that runs. `KindFloat` and its
      parser are landed, so that pass registers four knobs and adds no
      registry capability. **The one thing it must decide is whether
      `internal/engine/matcher` may import `internal/config`**, which
      `tunables.go:50-54` currently declines in writing, or whether the
      knobs are declared from `cmd/dubplate-server` instead.
   2. **`AdminEmail`'s dedicated contact row** stays post-1b and is now
      cheap: one declaration in whichever package owns the User-Agent,
      plus the owner touchpoint that minting a new required key implies.
   3. **How a rung declares its keys after 1b**, stated as the rule and
      as one worked example, pointing at rung 4's own declaration file as
      the shape to copy. **The rule for taking over a reserved class
      goes with it**: a rung deletes its own class from
      `internal/config/keys.go` in the same commit that declares it in
      its own package, that deletion is the only edit any lane makes to
      the file after 1b, and the registration-time check that the class
      arrives under the same name and key set is what stops the move
      from dropping a key. Rung 5 is the first to use it, taking `auth
      and identity` and `serve`.
   4. **Rung 12's streamer class moves to the driver's own package** when
      that package is born, and the three Navidrome names go with it.
      This is hook 6, and 1b landed the seam rather than the move.
   5. **A string or enumerated threshold kind** is the next kind anyone
      will want, for hook 12's forbidden-character set and hook 14's ID3
      version. The mechanism for adding a kind is now three edits and its
      registration-time check is already in place.
   6. **The gitleaks pin is gone and is not coming back.** Any future
      declaration file that writes a key literal beside a `Key:` field
      applies the same rename discipline; the reason lives in the comment
      Task 1 wrote.
8. **The lane handoff**: 1b's landing is what unblocks lane-2, alongside
   0c. **0c has landed**, committed at `b77e159`, so 1b's own close is
   what unblocks lane-2 outright; the close states this and updates the
   lane registry accordingly.
9. **STATUS stays present tense and under 60 lines.** Anything historical
   moves to HISTORY. The next action points at whatever the conductor
   sequences next, and the carried open items are updated: the four-knob
   item moves from "follows rung 1b" to "unblocked, next rung 3 pass",
   and the `KindFloat` half of that item is struck as landed.
10. **Both budgets scored**: token spend against the 2.5M ceiling, and
    attended interactions as two counts, planning misses and execution
    sittings, against a target of zero for both. **Record the numbers
    even when they look bad.**

**Acceptance criteria:**

- The HISTORY entry carries all three of what landed, what the gate
  caught, and what a later pass would be wrong to rediscover.
- The contract change is stated in the entry, not only in this plan.
- **All eight spec corrections are drafted**, each with its file, its
  line, its replacement sentence, and the first six words the drafter
  read at that line.
- Every one of the sixteen resolutions appears with its marker's
  location, drawn from the task returns.
- The gap list's count is stated and joined to the running total.
- The mutation table is complete: every named mutation in Tasks 1 through
  6 appears with an outcome.
- STATUS is under 60 lines and present tense.
- `bash scripts/check.sh` green.

**2.0 hook statement: neutral.** The close writes no code. It carries
hook 6's landed seam and hooks 12 and 14's unlanded kind into the
ledgers so the owning rungs read them.

**Independence: no.**

**Deliverable count: 4** (the HISTORY draft with its forward carries; the
eight drafted spec corrections and the resolution marker list; the
evidence record with the mutation table and the measured wall times; the
STATUS rewrite with the lane handoff and both budget scores). Gate green;
commit.

---

## Execution model

Workflow mode per the owner grants for this window: the pass-execute
chain (`~/.claude/workflows/pass-execute.js`) per task, **strictly
sequential**. Nothing runs in parallel. `docs/STATUS.md`'s lane registry
records lane-2 as BLOCKED until Phase 0b and 0c land, and 1b **is**
Phase 0b, so one code lane is the expected state for the whole pass.

Every task runs the repo's Sonnet implementer, dispatched as
`cairn-implementer` and instructed to read and follow
`.claude/agents/dubplate-implementer.md`, which commits its own task's
specific files once `bash scripts/check.sh` is green. Then `diff-reviewer`
(`claude-opus-5`) reads the diff against the task's acceptance criteria
and returns accept, fix, or escalate with `file:line` findings. **One
re-dispatch on `fix`; a second `fix` is the conductor's decision.** The
gate runs inside the chain, never in the main loop.

The conductor dispatches `code-simplifier` over each accepted task's
changed Go after review and before the pass close, and owns every ledger
edit, the lane registry, the eight drafted corrections, and the batch
commit. **Runaway guard armed at launch.** **Subagents never dispatch
agents.**

**Models.** `sonnet` throughout, with **no planned upshift**. This is a
refactor of one package against tests that already exist, which is the
profile `sonnet` handles well. **Two named upshift candidates**, each
conditional rather than by policy:

- **Task 2**, the registry and its collision gate, upshifted to `opus`
  **if Task 0 reports that 4b touched `internal/config`**. That would
  invalidate the inventory's file-by-file shapes, and the task would then
  be designing against a tree nobody has read.
- **Task 4**, the integration migration, upshifted to `opus` **if Task 3's
  review returns `fix`**. It is the pass's only behavior change to
  `Load`, its assertions are reworded during a move, and a reworded
  assertion that no longer fires is exactly the defect a tired refactor
  produces.

**Review.** Rung 1b is not one of the three full adversarial-cycle rungs
(5, 12, 16), so review scales to the in-chain `diff-reviewer` per task
plus the pass-end gate. No domain reviewer fans out: this pass writes no
`web/` surface, no auth or JWT surface, and no Worker.

**Dependencies and order, stated exactly:**

- **Task 0 runs first, in fresh context, before any implementer
  dispatch.** Its blockers are ruled by the conductor before Task 1.
- **Task 1 → everything.** It retires the `keys_test.go` line pin so
  every later task edits that file freely. That ordering is the whole
  reason it is first.
- **Task 2 → Tasks 3, 4, 5, 6.** They all use the registry it creates.
- **Task 3 → Tasks 4, 5, 6.** They build on the migrated package. Task 5
  is included because it modifies `thresholds.go`, where Task 3 retires
  the package-level registry, and `doctor.go`, which Task 3 rewires.
- **Task 4 → Task 6.** Doctor's exhaustiveness count includes the
  integration family as the registry holds it.
- **Task 5 → Task 6.** Doctor's exhaustiveness count includes the
  threshold family, and Task 5 changes doctor's kind switch.
- **Tasks 0 to 6 → Task 7.**
- **Task 5 is marked independent of Task 4** on file disjointness, with
  the appended-test-function rule for `doctor_test.go`. **Task 1 is
  independent in `.gitleaksignore` and `scripts/check.sh` alone**; its
  third file, `internal/config/keys_test.go`, is shared with Tasks 3 and
  4, which is why Task 1 runs first rather than in parallel. Both still
  run sequentially, because one tree runs one executor and the module
  gate runs over `./...` regardless. The conductor may reorder Task 5
  before Task 4 and after Task 3 if Task 4 stalls.

**Where this pass splits, if it splits.** Seven code tasks plus a
pre-task sits at the roughly-eight bar, so a split is not proposed. **The
split point, if the conductor needs one, is the Task 4 / Task 5
boundary**: Tasks 0 through 4 are the registry and its migration, which
leave the tree in a coherent state (every existing test green, the
contract changed, `keys.go` closed); Tasks 5 through 7 are the kind
mechanism, the sync assertions, and the close. **The recommendation is to
keep one pass**, for one reason that outweighs the rest: 1b's landing is
what unblocks lane-2, and a half-landed registry unblocks nothing, so a
split would put a full close and a full staleness pre-task between the
second code lane and everything it waits on.

**The trigger sits at the ceiling, and the band above it has one rule.**
A Task 3 re-derivation past **2.5M** is the conductor's decision between
splitting at the Task 4 and Task 5 boundary and raising the ceiling,
taken at the Task 3 checkpoint on the number rather than on a feeling.
Only a re-derivation that also breaks a window cap reaches Geoff.
*(Revision 1 set the trigger at 3.5M, a full 1M above the ceiling, and
stated no rule for the band between, which is exactly where a 40%
overrun would land: rung 4's own M-band estimate ran roughly 50% over,
inventory `:8.2`.)*

**Budget:** 2.5M ceiling, derived above from 4a-ii's measured rate with
the arithmetic shown; flag at 2.0M, routed to the conductor per this
window's grants; checkpoint every four tasks, and STATUS at Task 3 and
Task 7. Spend is scored at close against the ceiling and against **zero
attended interactions**, which is this pass's target. The sixteen
resolutions are applied provisionally and marked rather than blocking, so
none counts as an execution sitting.
