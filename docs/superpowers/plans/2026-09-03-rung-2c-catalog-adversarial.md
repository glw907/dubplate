# Rung 2, Pass 2c (Real-Catalog, Adversarial, and Stock-Config Strata) Implementation Plan

> **For agentic workers:** workflow mode per CLAUDE.md's "Build
> process" section: pass-execute chain per task, sequential. The
> conductor consumes reports and never reads diffs inline. Subagents
> never dispatch other agents; `code-simplifier` is the conductor's
> dispatch after a task's review accepts. The conductor also owns the
> ledgers (`docs/STATUS.md`, `docs/HISTORY.md`, `ROADMAP.md`) and the
> batch commit of anything a task drafts into them. Rung 2 is not an
> adversarial-cycle rung (those are 5, 12, and 16), so review scales
> to the in-chain `diff-reviewer` plus the pass-end gate.

Revision 2, 2026-09-03: folds the adversarial plan review
(`research/2026-09-02-greenfield-spec-reviews/rung-2c-plan-review.md`),
both blockers and majors 1 through 12 as specified, except where a
fix conflicts with another fix and the plan adapts, noted inline
below. B1 and M11 fold together: B1's new `harness seed-items`
projection tool (Task 2 in this revision) is the mechanism M11's fix
reuses for selection-time structural verification, so both land as
one coherent change to Task 3 (the dossier, formerly Task 2) and to
the Global Constraints network-door bullet. M2 and M12 each fold via
the first of their review's two offered fixes rather than the second:
the stock-config stratum splits into Tasks 19a and 19b (28 mandatory,
8 supplementary) instead of carrying an unbudgeted "name four donors"
step, and the six 16-case capture batches become twelve 8-case
batches instead of adding a bespoke partial-batch rule, both matching
2b's own proven cadence rather than inventing a new one. M10 folds as
specified: the pass now cuts after the stock-config stratum, and
former Tasks 13 (file-level derivation) and 14 (the two chore docs)
move to pass 2d together with the whole-corpus README rewrite; M3,
M4, and M5, whose concrete fixes target Task 13's behavior contract
specifically, are pre-ruled and recorded as carry-forwards in Task
20's HISTORY draft rather than dropped, so 2d's plan does not
re-derive them from scratch. M8's wall-time-floor formula and M9's
ceiling and re-derivation-point corrections are applied throughout,
including to the estimates in the now-deferred 2d carry-forward note.
Minors 1 and 3 through 9 fold where cheap; minor 2, which corrects
Task 13's own Files list, folds into the same 2d carry-forward note
as M3 through M5 rather than as a standalone edit, since the task it
describes no longer exists in this document. The review's "Verified
sound" section records the mechanisms implementers must not
re-litigate and is unchanged by this revision.

Consumes pass 2b's shipped machinery exactly as committed: the
`cassette` schema at v2, `internal/harness/stub`, `oracle.Derive` and
`derive_oracle.py`, `internal/harness/mbrequest`, `harness capture`,
`harness synth`, `harness freeze`, `internal/harness/analysis` (the
proximity flag, the tie detector, the claims checker), `make derive`,
the `estate` and `stock` profiles, and the 68 committed cases. 2c adds
cases, two ledger formats, one selection-and-authoring tool, and one
new config profile on top of that substrate. It does **not** reopen
the cassette package this pass (M10): the one additive exception
originally scoped here (`Manifest.MediaFile`) moves to pass 2d with
Task 13.

**Primary input:** the planning inventory
(`docs/superpowers/research/2026-09-03-rung-2c-planning-inventory.md`,
2026-09-03), which enumerates the four seedable known-oracle-bug
classes with their citations, the real-catalog selection axes, the
estate-versus-stock divergence table checked against pinned beets'
own `config_default.yaml`, the file-level derivation's four concrete
requirements (now 2d's precondition list), and the derive-time
scaling arithmetic. Read it in full before Task 1, alongside this
plan's fold note above. Its findings are the pass's organizing
decisions and must not be re-litigated: that `Case.Profile` and
`Case.Expected` are single fields, so the stock-config stratum needs
sibling case directories rather than a second record; that #5458's
threshold-boundary class needs no new case because 2b's proximity
ledger already carries it; that the file-level tag set is a derived
reading of a record rather than a field inside one; and that the
network is not this pass's budget driver.

**Goal:** finish rung 2's real-catalog, adversarial, and stock-config
evidence. Ship the real-catalog stratum captured live from
MusicBrainz, the adversarial stratum crafted to reproduce pinned
beets' known matching defects, the stock-config stratum that keeps
ruling 3's claim from being vacuously true, and the first
adjudication batch that lets a human rule on a divergence from the
written record alone. The file-level applied-tag derivation rung 8 is
blocked on, and the two standing chore docs, move to pass 2d (M10).

**Owner touchpoints: one guaranteed, one contingent (M9).** The
guaranteed one is the first adjudication batch, prepared at Task 20
and delivered at the pass close. It does **not** block this pass's
close: the pass closes with rows sitting at `pending-adjudication`,
and the gate accepts them. An unadjudicated divergence blocks the
**launch gate**, which is a different gate at a different time. The
contingent second is the standing 80%-ceiling flag CLAUDE.md's Build
process already requires of every pass: if spend crosses 80% of the
re-derived ceiling, the task in flight finishes, STATUS is written,
and one combined question goes to Geoff before continuing. It is
contingent, not a second guaranteed touchpoint, and fires only if the
ceiling is actually approached. Nothing else in this pass asks Geoff
anything. If a task believes it has found a second question, that is
a finding to report and a row to write, never a mid-pass ask.

**Spec:** `docs/superpowers/specs/2026-09-02-greenfield-server-build.md`
(revision 2.1, ratified), ladder entry 2 and the launch bar's two-tier
gate. Ruling 2 (tier one is measured against release truth where the
oracle is known-defective) is what makes the adversarial stratum legal
to author at all. Ruling 3 (tier one also holds over the stock-config
stratum) is what Tasks 19a and 19b exist to make non-vacuous.

**Token ceiling: 12M** (raised from 8M, M9), as a starting figure,
re-derived after Task 3 (the selection dossier) reports its actual
token spend. Task 3, not a capture batch, is now plausibly the
single largest task in the pass: its M11 resolution folds
selection-time capture verification into it, on top of the
web-page research the original Task 2 already carried, so the
re-derivation point moves from after the first capture batch to
after Task 3 (M9), and Task 3 must report its own spend. The figure
is extrapolated from 2b's actual ≈5.8M across 14 tasks and 61
synthetic cases, adjusted upward for this pass's larger surface: 21
tasks (up from 15, mostly from B1's new tool task and M12's
finer-grained batching), 154 new case units (m8; see the token
ceiling paragraph below for the arithmetic), two new ledger formats
with gate-resident checkers, a new config profile, and a research
dossier now carrying live-capture verification for four of its six
axes. It is **not** adjusted downward for "a captured case costs far
less design thought than a synthetic one" — B1 showed that claim was
wrong for the authoring half, and this revision folds the fix (Task
2's tool) rather than the unadjusted discount. Flag at 80% (9.6M):
finish the task, write STATUS, and the contingent touchpoint above
fires. **Checkpoint:** STATUS after every task, which is 2b's cadence
and is what the growing task count and derive cost both require —
every task's running case count and its broken-down derive wall time
are the numbers the conductor sizes the next task against.

**Case-unit arithmetic (m8):** 96 real-catalog primary rows (Task 3)
+ 12 real-catalog reserve rows (M1; drawn on only if a primary row
fails to converge, not counted as delivered until drawn) + 10
adversarial string-heuristic cases (Task 16) + 10 adversarial
assignment/alias cases (Task 17) + 2 locale sibling cases (Task 18) +
36 stock-config sibling cases (Tasks 19a/19b) = **154 new case
units**, where a "case unit" is one committed `corpus/cases/<id>/`
directory with its `case.json` and `claims.json`. The reserve rows
are dossier rows, not case units, until one is promoted in place of a
failed primary.

---

## Global Constraints

- go-conventions before any Go; python-conventions before any Python.
  Both legs are already gated (`golangci-lint`, `vale` over Go comment
  prose, `ruff` over tracked `*.py`).
- **The live network is in scope this pass, through exactly one
  door.** Every `/ws/2` request goes through `harness capture`. No
  `curl`, no `WebFetch`, no second HTTP client, no direct API read
  "just to check something" — the tool's identified User-Agent and its
  1 request per second pacing are the compliance surface, and a
  request outside it is a terms violation with no record. Stated
  precisely, since a scoped-looking bullet reads as absolute otherwise
  (m9): **no route to `/ws/2` other than `harness capture`, by any
  tool**, and exactly two non-`/ws/2` network reads are sanctioned —
  Task 3's browsing of musicbrainz.org's own human-facing search and
  release pages, at a courteous pace, for the axes that do not need
  `/ws/2`-grade confirmation; and Task 18's `gh issue view 6358
  --repo beetbox/beets`. **Selection may also seed a candidate through
  `harness capture` and read it back only through the seed-items
  projection tool (M11; see Task 2 and Task 3)**, which is the
  compliant door the plan already blesses, not a new one — a
  candidate confirmed this way carries its staging directory forward
  as the case's seed, and a rejected candidate's staging directory is
  deleted, an uncommitted discard the `.gitignore` already covers.
- **`DUBPLATE_HARNESS_CONTACT` comes from the environment.** It is
  set in `~/.bashrc` via dotfiles `3a1c551`. It is never written into
  a repo file, never substituted with a literal, and never printed
  into a task return. `harness capture` refuses to run without it,
  which is the intended behavior. Registration of a MusicBrainz
  application was declined at 2a; **anonymous 1 request per second
  stands** and no task proposes changing it.
- **The gate never reaches the network and never needs the venv.**
  `bash scripts/check.sh` and `go test ./...` stay offline and
  Python-free. `make derive` needs the venv, and derivation replays
  only committed files through `internal/harness/stub`. Nothing this
  pass adds moves a network or venv dependency across that line.
- **A captured response body is evidence, is never edited, and is
  never read into agent context** (restated per B1 — the prior
  wording banned an agent from opening `responses/*.json` at all,
  which was incompatible with authoring that same case's
  `Input.Items`; the fix is a tool, not a looser rule).
  `harness capture` writes response bytes verbatim with
  `os.WriteFile`. A real-catalog response is not hand-edited, not
  reformatted, not pretty-printed, and not opened by an agent to "see
  what is in it" — a maximal-`inc=` release document for a box set is
  large, and reading one is a pure token loss against a file whose
  only job is to be replayed byte-for-byte. A response is read only
  by tooling: `harness seed-items` (Task 2) projects the fields an
  agent actually needs — per-track title, length, position, recording
  id, release-level album, albumartist, date, plus the structural
  facts Task 3's fact-verified axes need (disc count, first-release
  date, documented aliases) — into a small, agent-readable baseline.
  Inspect a captured case through its derived record, its
  `claims.json`, and `seed-items`' projection, never through
  `responses/` directly.
- **The response hand-edit escape hatch is synthetic-only.** Pass 2b
  blessed editing a generated `responses/001.json` before the first
  freeze to reach a shape `ReleaseSpec` cannot express, with the edit
  named in the case's README inventory row. That hatch stays open for
  `harness synth` output (the adversarial stratum uses it) and is
  **closed for captured output**. Editing a captured real response
  would forge evidence, and no case in this pass needs it.
- **Every request pinned beets issues must still resolve to a
  committed 200.** The stub has no third outcome, and the schema
  carries no response status. **The response-status extension is
  deliberately not taken this pass**, and this is the plan's recorded
  reason: all four known-oracle-bug classes in the inventory's section
  (a) are distance or selection defects, not transport defects, so
  none of them needs a 404 or a 503 to reproduce. The extension
  carries forward as a named non-adoption, available to whichever
  later rung first has a case that genuinely needs it.
- **The cassette surface is frozen this pass, with no exception**
  (M10). The `Manifest.MediaFile` version field, the `schema_version`
  bump to `"3"`, and `file_tags.json` all move to pass 2d together
  with Task 13; see Task 20's carry-forward note for the ruled
  decisions (M3, M4, M5, m2) that 2d's own plan should not have to
  re-derive.
- **Promotion ordering, exactly** (2b's correction; the plan's
  intuitive order is impossible as shipped): author or capture in
  `corpus/staging/<id>/`, iterate to zero misses, freeze, then
  **promote the frozen case into `corpus/cases/<id>/` FIRST**, then
  run `go test -tags oracle ./internal/harness/oracle -run Ledger
  -update-ledgers`, then write `claims.json` and any owed proximity
  row, then run both gates, then commit. The invariant is "never
  commit red", not "never move before the row exists".
- **Ledger membership, exactly** (2b's correction): `ties.json` gets a
  generated row for every case. `proximity.json` is an **exception
  ledger** — a row only for a case the flag returns `flagged` or
  `skipped` for, carrying `boundaries` (flagged rows only) and a
  written reason. Listing a clear case fails the gate. `ties.json`
  dispositions and reasons are human adjudication and survive an
  `-update-ledgers` run untouched.
- **Every case that reaches `corpus/cases/` carries `claims.json`**,
  written off its derived record, never invented. Every task's commit
  includes each new case's `claims.json`; the per-task Files lists
  below say so once and do not repeat it.
- **A sub-second `make derive` is a cache hit and proves nothing.**
  Force `-count=1` for any derivation reported as evidence, and report
  wall time broken down (corpus derive, ties derive, component
  derive), never a bare "green".
- **Run the full derive once per task, not once per case.** The
  per-case loop uses `harness freeze --case <dir>`, which touches one
  case. `make derive` and `-update-ledgers` are both O(all committed
  cases) per run, so their cost climbs monotonically through the pass
  and a per-case full derive would dominate everything else. Budget
  wall time per task, not per new case. **The per-case derive cost
  includes beets' hard 0.25s inter-request floor** (M8): `1.04s +
  0.25s × (requests − 1)` per case, not a flat 1.04s, since the floor
  applies between any two replayed requests even at the `estate`
  profile's high configured ratelimit. A typical real-catalog case
  (≈1.4 requests) costs ≈1.1–1.3s; a search-fallback case (6–7
  requests, `should-reject`'s shape) costs ≈2.3–2.5s. This raises
  every wall-time estimate below versus a flat-1.04s arithmetic, and
  the batch and Execution-model sections state the corrected figures.
- **Never self-adjudicate, and `disposition` is owner-only** (M6,
  restating and splitting the prior single rule). `reproduced`
  (`oracle_bugs.json`) is a machine-observable field an agent may set,
  with its evidence. `disposition` (both new ledgers) is never set by
  an agent to anything but `pending-adjudication`, including a case
  whose written evidence looks unambiguous — Task 1's threshold-
  boundary hypothesis and Task 18's #6358 outcome are the two places
  this plan is tempted to skip that step, and both are written
  explicitly below at `pending-adjudication`. A case whose derived
  verdict looks wrong against release truth gets a row in
  `corpus/analysis/adjudications.json` with its evidence written out,
  and the case keeps its derived record unchanged. Do not redesign a
  case to make an uncomfortable verdict go away. Do not write an
  expected verdict, or a non-pending disposition, a human has not
  ruled. The gate-resident checker enforces it (Task 1): a row whose
  `adjudicator` is not a human name may not carry a disposition other
  than `pending-adjudication`, proven by a fixture.
- **Every expectation is still derived by executing pinned beets.** An
  `Expected` record is only ever what `oracle.Derive` returned, written
  through `cassette.Case.Write`. This holds for real-catalog cases
  exactly as it held for synthetic ones.
- **Universality.** Real-catalog cases use real public MusicBrainz
  releases, which is fine. No estate hostname, path, or email anywhere
  outside `internal/config`. Profiles are referenced by name
  (`estate`, `stock`, and Task 18's new `locale`), never inlined.
- **Assertion discipline.** No silent-success fakes, no vacuous
  assertions. Every new gate-resident check in this pass ships a
  fixture that makes it fire, and the task reports the mutation that
  proved it can fire.
- `_salvage/` stays read-only; nothing in this pass touches it (Task
  14's `_salvage` FLAC-fixture precedent is ruled for 2d's use already
  — see Task 20's carry-forward note, M4 — and no code in this pass
  needs it). The box is untouched. Gate green and clean `git status`
  at every task end; never commit red; specific files only; commit
  footer `Co-Authored-By: Claude <noreply@anthropic.com>`.

## How a real-catalog case gets made

Stated once here because most tasks depend on it. The loop differs
from 2b's synthetic loop in two places: `harness capture` sits where
`harness synth` sat, and step 3 goes through the `harness seed-items`
tool (Task 2) rather than being hand-invented, which is what closes
B1's gap.

1. **Select the MBID first, offline of `/ws/2` for most axes, and
   capture-verified for the four axes that need a structural fact
   only `/ws/2`-grade data confirms** (M11). Task 3 produces the
   selection dossier, so a capture batch never spends its own budget
   deciding what to capture. Each dossier row names the release MBID,
   the case id, the stratification axis it serves, and whether it is
   also a stock sibling or a file-tag subject (the latter flag now
   feeds pass 2d).
2. **Seed:** `harness capture --mbid <release> --out
   corpus/staging/<id>`. This fetches the release lookup at beets'
   maximal `inc=` and follows a pseudo-release's `transl-tracklisting`
   relation when it finds one. The seeded case carries only its
   request log. **For the four fact-verified axes (reissue/remaster,
   multi-disc/box-set, non-Latin, and the two-release sub-axis),
   Task 3 already performed this step during selection** (M11); a
   confirmed dossier row's staging directory is reused as-is, and a
   batch task starts from step 3.
3. **Author the input via `harness seed-items --case <dir>`** (B1,
   Task 2), which reads the staged `responses/*.json` and projects a
   correct baseline `Input.Items` — per-track title, length, position,
   recording id, release-level album, albumartist, and date — reusing
   `internal/harness/synth`'s derive-then-perturb model. A case whose
   axis is "real MB noise" or "mistagged material" then applies the
   tool's `overrides` vocabulary (`field`, `missing`, `extra`) to
   reflect the actual mistagging or partial rip, which is the point
   of the axis. `EvalDate` follows one rule for every captured case
   (m4): **the case's capture date**, matching the smoke corpus's own
   convention per `corpus/README.md`.
4. **Derive, read misses, capture, re-derive, until zero misses.**
   `harness freeze --case corpus/staging/<id>` derives and freezes.
   When beets falls through to its text search, `Derive` writes
   `misses.json` and `harness capture --case corpus/staging/<id>`
   fetches exactly what it names. Each search hit costs its own full
   release lookup, so a fallback case converges over two or three
   rounds, each paying its own network wait. This is the recorded
   fixed point, not a stable request set.
5. **Promote, update ledgers, write claims, gate, commit**, in the
   order the Global Constraints fix.

An adversarial case (Tasks 16 and 17) runs 2b's synthetic loop
instead: `harness synth` from a `ReleaseSpec`, the hand-edit hatch
before the first freeze where the shape needs it, then the same
promotion ordering. The one exception is the alias-selection class,
whose whole point is real alias data, so those cases capture.

A stock sibling (Tasks 19a/19b) copies its estate twin's `Input` and
its captured `responses/*.json` verbatim into a new self-contained
case directory, sets `profile: "stock"` and its own `id` to
`<id>-stock` (m3 — matching its directory, since
`TestEveryCorpusCaseLoadsClean` fails any mismatch), and freezes.
Where stock's different verdict drives beets to a search the estate
capture never triggered, the miss-file mechanism handles it exactly
as it does for any other case.

**A two-release sub-axis case** (B2, part of Task 3's real-catalog
axes) seeds the correct release exactly as step 2 above, then authors
items — step 3 — deliberately ambiguously enough that beets' own
text-search fallback surfaces a genuine decoy release as a competing
candidate: a decoy's catalog-adjacent edition string, or a title the
decoy and the correct release both plausibly carry. The existing
misses-driven loop (step 4) then captures the decoy's document too,
with no new capture mode needed. If a candidate row cannot be made to
surface its decoy through search alone, that is a finding to report
in the dossier row, not a case to force.

---

### Task 1: The adjudication and known-oracle-bug ledger formats

Model: `opus` (novel schema design that rung 4 inherits, plus two
gate-resident checkers whose whole value is that they cannot pass
vacuously). Runs first and alone: Tasks 16 through 19b all write rows
into these files.

**Files:**
- Create: `corpus/analysis/adjudications.json`,
  `corpus/analysis/oracle_bugs.json`
- Modify: `internal/harness/analysis/` (the two new checkers) + tests
  + `testdata` fixtures
- Modify: `internal/harness/corpus_test.go` (wire both checkers into
  the gate), `corpus/README.md` (a new "The adjudication artifacts"
  section under "The analyses")

**Behavior contract:**

- **`adjudications.json` is the general divergence row, designed
  forward-compatible with rung 4.** The build spec names a divergence
  log and an adjudication artifact at rung 4 (Shadow) for a
  structurally identical problem: a human ruling on a matcher
  disagreement from a written artifact alone, without re-deriving
  anything. This pass originates that schema. A row carries: the
  subject identity (a corpus case id today, an album plus release MBID
  when a live shadow divergence writes one, so the field admits both);
  a reference to the oracle's own record rather than a
  re-transcription of it, since `cassette.Expected` already holds
  recommendation, action, distance, and the best candidate's full
  penalty breakdown; what is suspected wrong and why, in prose,
  carrying its citation (an issue number, or an evolution-lessons
  finding id such as `U1/ft-pattern`); release truth, with its own
  citation, or an explicit statement that the input material has no
  correct MusicBrainz entry at all; `disposition` plus `reason`, the
  same two-field shape `ties.json` already uses; and `adjudicator`
  plus `date`.
- **`pending-adjudication` is a legal disposition here and an illegal
  one at the launch gate.** The gate fails a row with a missing
  disposition or one no parser recognizes, and accepts
  `pending-adjudication` with a written suspicion and written release
  truth. The README section states plainly that the launch gate does
  not accept it, and that this is the mechanism by which an
  unadjudicated divergence blocks a launch without blocking a pass.
  Recognized dispositions at minimum: `pending-adjudication`,
  `oracle-wrong` (carrying the adjudicated correct verdict),
  `oracle-correct` (the crafted case failed to reproduce the suspected
  defect), and `no-correct-answer` (the material has no correct MB
  entry). **Only a row whose `adjudicator` is a human name may carry
  anything but `pending-adjudication`** (M6); the gate-resident
  checker enforces this and ships a fixture proving it fires.
- **`oracle_bugs.json` is one row per known-defect hypothesis, not per
  case.** A hypothesis may be probed by several cases, or by none. A
  row carries the bug class and its citation, the probe case ids, a
  `reproduced` field (`yes`, `no`, or `not-attempted` with a reason,
  agent-settable), pointers to the adjudication rows the probes
  produced, and the same disposition/reason/adjudicator/date fields as
  `adjudications.json`, with the same owner-only `disposition` rule.
  **A hypothesis that does not reproduce is still a written row**,
  matching the corpus's standing "an unacknowledged flag is a
  failure" discipline: a checked-and-cleared hypothesis is never
  silently dropped.
- **Both checkers are pure Go and fully gate-resident**, needing no
  venv and no network. They fail on: a case id in either file that
  names no committed case; a case id named in a probe list that fails
  to resolve to a committed case (m6 — replacing the prior forward
  direction, "an `adv-`-prefixed case in no probe list," which coupled
  the check to an id-prefix convention Task 18's locale probe cases do
  not carry; the reverse direction below is the cheap fix and covers
  every probe list uniformly); a row with a missing or unrecognized
  disposition; a `pending-adjudication` row with no written suspicion
  or no written release truth; a row whose `adjudicator` is not a
  human name carrying a disposition other than `pending-adjudication`
  (M6); and an `oracle_bugs.json` row whose `reproduced` is `yes` with
  no adjudication row referenced. Each failure mode ships a `testdata`
  fixture proving it fires, and the task reports the mutation used.
- **Seed five hypothesis rows now, at `not-attempted`** (M7, raised
  from four): the featuring-pattern ("ft") false positive
  (`742ebdf69`, `7b59604c5`, `19d091440`); the parenthetical/remix
  near-duplicate discount (#5868), split out as its own hypothesis
  rather than folded into the ft row, since Task 16 authors it as a
  distinct case class; non-Latin, pseudo-release, and alias
  mis-selection; per-pair track-assignment rejection gaps (#2981,
  #6078); and threshold-boundary strictness (#5458). The
  threshold-boundary row lands immediately at `reproduced: "no"`,
  `disposition: "pending-adjudication"` (M6 — never `oracle-correct`
  directly, even though the reasoning is already legible), with the
  inventory's own reason recorded as its evidence: pass 2b's
  `rec-strong-medium-boundary` case and the `proximity.json` ledger
  already cover it by construction. Do not author a new case for it.

**Deliverable count: 4** (the adjudication row schema with its
forward-compatible identity field and its owner-only disposition
rule; the known-oracle-bug hypothesis ledger; the two gate-resident
checkers with their fire-proving fixtures; the five seeded hypothesis
rows and the README section). Gate plus `make derive` green; commit.

---

### Task 2: The `harness seed-items` projection tool

Model: `sonnet` (B1 — the fix is well-specified: reuse
`internal/harness/synth`'s existing derive-then-perturb model and its
`overrides` vocabulary against a captured response instead of a
`ReleaseSpec`). Network: no. Runs before Task 3, since Task 3's
selection-time verification and every batch task's item-authoring
step both depend on it.

**Files:**
- Create: `cmd/harness/seeditems.go` (new `harness seed-items --case
  <dir>` verb) + tests
- Modify: `corpus/README.md` (a new "Authoring a captured case"
  section describing the tool's role in the loop above)

**Behavior contract:**

- **Reads `responses/*.json` from a staged case directory and writes
  a correct baseline `Input.Items` into `case.json`.** Projects, per
  track: `title`, `length` (to the same precision the release
  document carries), `position`, MusicBrainz recording id, and the
  release-level `album`, `albumartist`, and `date`. This is the exact
  gap B1 identified: none of those values is inventable from an MBID
  and a genre, and they live only in the response document a captured
  case's authoring step is otherwise forbidden to read.
- **Reuses `internal/harness/synth`'s existing model rather than
  building a second one.** The derive-then-perturb shape and the
  `overrides` vocabulary (`field`, `missing`, `extra`) already exist
  for synthetic cases; this tool applies the same perturbation step
  to a projection sourced from a real response instead of a
  `ReleaseSpec`.
- **The tool is the sanctioned reader of a captured response.** No
  other code path in this pass, and no agent directly, opens
  `responses/*.json`; this closes the Global Constraints' "never read
  into agent context" bullet with an actual mechanism rather than
  leaving it as an unenforceable convention.
- **Round-trip proof against the committed corpus.** Since no
  real-catalog case exists yet to validate against, prove correctness
  by running `seed-items` against two or three already-committed smoke
  cases with captured responses and known-good `Input.Items` (for
  example `clean-match`, `media-uneven-split`) and diffing the
  projection against the committed values. Report the diff as the
  fire-proving evidence for this tool, the same discipline every other
  gate-resident mechanism in this pass follows.

**Deliverable count: 4** (the `seed-items` verb and its projection
logic; the `overrides` reuse for the mistagged/partial-rip axes; the
round-trip proof against committed smoke cases; the README section).
Gate green; commit. No `make derive` (no case changes).

---

### Task 3: The real-catalog selection dossier

Model: `sonnet`. **Network: yes** — musicbrainz.org web pages for
most axes, plus `harness capture` seeding and `seed-items` projection
for the four fact-verified axes (M11). No direct `/ws/2` read by any
other route.

**Files:**
- Create: `docs/superpowers/research/2026-09-03-real-catalog-selection.md`
  (the dossier)
- Create (via `corpus/staging/`, kept only for confirmed rows; a
  rejected candidate's staging directory is deleted before this task
  ends): staged candidates for the fact-verified axes

**Behavior contract:**

- **96 real releases plus 12 reserve rows, selected against explicit
  per-axis quotas** (M1 raises the original 96-with-no-reserve), so
  the stratum's diversity is a decision on the record rather than an
  emergent property of whatever got captured first. The primary
  quotas: classical and compilation material, **10** (reduced from 16
  — B2 splits 6 of the original 16 into the new two-release sub-axis
  below, since a classical release's structural gaps make it the
  domain most likely to actually produce a decoy candidate);
  reissues and remasters whose first-release date genuinely diverges
  from the release date, **14** (fact-verified — capture-confirm the
  divergence); Various Artists compilations and genuine multi-disc box
  sets of three or more discs, **14** (fact-verified — capture-confirm
  the disc count; 2b's synthetic family 6 never went past two
  mediums); non-Latin scripts across Japanese, Korean, Cyrillic, and
  Greek, **12** (fact-verified — capture-confirm the alias), at least
  one of which carries a documented locale-specific alias in its
  `artist-credit` and is flagged as Task 18's host release; real
  MusicBrainz noise (bootlegs present in MB with sparse metadata,
  official reissues with near-duplicate catalog numbers, partial
  rips), **14**; genre and era spread filling the confident band,
  **26**, chosen to avoid the corpus silently over-indexing on the
  rock and pop of the smoke corpus; and the **new two-release
  sub-axis, 6** cases (B2), each capturing both a decoy pressing and
  the correct release into one case (mechanism: "How a real-catalog
  case gets made," the two-release paragraph), predominantly drawn
  from classical/compilation material. Primary total: 10 + 14 + 14 +
  12 + 14 + 26 + 6 = **96**. **Reserve: 12 rows, tagged by the axis
  they back up** (roughly one in eight, matching a realistic
  capture-convergence failure rate on real data), drawn on by a batch
  task only when a primary row cannot be authored as described (see
  Tasks 4 through 15's substitution rule).
- **The stratum's headline value is conformance breadth over real MB
  document shapes, not a confident-band wrong match** (B2, restating
  the stratum's purpose). A case's candidate pool is the release its
  items were seeded from, so the only reachable verdicts for the
  confident-band majority are "matched its own source" or "fell
  through to search" — a confident wrong match on that majority is
  structurally unreachable and is not this stratum's finding class.
  The two genuine two-release cases (above) and the search-fallback
  minority (the noise axis, principally) are where a false confident
  match can actually appear, and the "standing findings duty"
  paragraph in the batches section below is scoped to them
  accordingly.
- **Each row is complete enough that a capture batch never re-decides
  anything.** Release MBID, proposed case id (`real-<axis>-<slug>`),
  the axis it serves, one sentence on why this release serves it, the
  expected shape of its `Input.Items` (a clean rip, a partial rip, a
  mistagged rip), whether it is a mandatory stock sibling, and whether
  it is a file-tag subject (a flag this pass still records for 2d's
  benefit, even though 2d authors those cases). The dossier assigns
  rows to the twelve capture batches, 8 per batch, mixing axes within
  each batch so a batch failure never wipes out one whole axis.
- **Mark the mandatory stock siblings now.** Per the inventory's
  section (c), estate and stock can only actually disagree on a
  verdict where `original_date` or `per_disc_numbering` is
  load-bearing, so every reissue/remaster row and every multi-disc or
  box-set row carries the mandatory flag: 14 + 14 = **28 rows**,
  which is exactly Tasks 19a/19b's mandatory budget (M2) — no
  trimming, no "name four donors" step.
- **Selection stays off `/ws/2` except through the one sanctioned
  door.** For the fact-verified axes (reissue/remaster, multi-disc/
  box-set, non-Latin, two-release sub-axis), seed the candidate via
  `harness capture --mbid <mbid> --out corpus/staging/<tentative-id>`
  and confirm its structural facts (disc count, first-release-date
  divergence, documented alias) via `harness seed-items`' projection,
  never by reading the raw response. A confirmed row's staging
  directory carries forward to its batch task; a rejected candidate's
  is deleted. For the remaining axes (classical/compilation genre
  facts, MB noise, genre/era spread), browse musicbrainz.org's own
  search and release pages at a courteous pace — those axis claims
  are genre and provenance facts, not `/ws/2`-shaped structural
  fields, so the web-page route is both sufficient and the right
  scope for it.

**Deliverable count: 4** (the seven axis quotas, including the new
two-release sub-axis and the restated stratum purpose, with their
per-row rationale; 96 primary plus 12 reserve dossier rows with case
ids and item-shape expectations; the batch assignment across Tasks 4
through 15; the mandatory-stock-sibling and file-tag flags with the
28-mandatory count confirmed against budget). **Report actual token
spend**, since the ceiling re-derivation point moves here (M9). Gate
green; commit. No `make derive` for the confirmed-and-kept staging
directories (they are not yet promoted).

---

## Real-catalog capture batches (Tasks 4 through 15)

Twelve tasks, 8 cases each, 96 total (M12, matching 2b's proven
cadence rather than the original 16-case batches). Every case follows
the "How a real-catalog case gets made" loop above; the
fact-verified-axis cases arrive already seeded from Task 3 and start
at loop step 3. **All twelve hit the live network through `harness
capture`**, whether for the initial seed (the non-fact-verified axes)
or for miss-driven follow-up captures (every axis). Batch membership
and case ids come from Task 3's dossier; a batch never re-selects.

**Wall-time budget, per batch.** Network: 8 cases at the inventory's
measured ≈1.4 requests per case is ≈11 requests, and at 1 request per
second that is well under 30 seconds of pure network wait, plus a
second or third convergence round for the minority of cases that fall
through to search. The network is not the budget driver and no batch
should be sized against it. The real wall time is the task-end full
derive, which is O(all committed cases), now costed with the
inter-request floor (M8): `1.04s + 0.25s × (requests − 1)` per case
for corpus derive, plus 2b's measured 0.79s per case for ties derive,
plus the `-update-ledgers` run being a second ties-shaped pass. At a
blended ≈1.2s/case corpus-derive figure and a corpus growing from
about 76 to 164 committed cases across this batch loop, a mid-loop
batch pays roughly **4 to 6 minutes**, climbing toward the loop's end.
**Twelve smaller batches trade a smaller per-batch failure blast
radius for twelve full-derive runs instead of six** (the price of
matching 2b's proven cadence over inventing a partial-batch rule,
M12); cumulative derive wall time across the batch loop is
correspondingly higher than the original six-batch design's estimate,
and is priced into the raised 12M ceiling. Report the broken-down
wall time every batch; the trend is the number the conductor sizes
the later batches against.

**Token budget.** Task 3's own reported spend (above) sets the
starting ceiling for this loop; Task 4 additionally reports its own
actual spend as the batch loop's first data point, and the conductor
re-derives the remaining ceiling from it. The single largest lever is
the constraint above: **do not read captured response bodies; use
`seed-items`.**

**A batch's standing findings duty**, scoped per B2 to where a
confident wrong match is actually reachable: a two-release sub-axis
case, or any search-fallback case, landing a confident (`strong`)
recommendation on a match that release truth says is wrong is a
tier-one candidate and the sharpest thing this stratum can find in
that shape. Write it as a `pending-adjudication` row in
`adjudications.json` with its evidence, leave the derived record
alone, and report it. Do not redesign the case. Do not self-adjudicate.

### Task 4: Real-catalog batch 1

Model: `sonnet`. **8 cases.** Network: yes.

**Files:**
- Create (via `corpus/staging/`, promoted to `corpus/cases/<id>/`): 8
  cases, each with its `claims.json`
- Modify: `corpus/analysis/ties.json`, `corpus/analysis/proximity.json`
  (exception rows only), `corpus/analysis/adjudications.json` (any
  tier-one candidate), `corpus/README.md` (a new real-catalog
  inventory section, including an "Authoring a captured case" note
  and the corrected `--out corpus/staging/<id>` path in the existing
  capture section, m5; Task 20 later rewrites the whole inventory)

**Behavior contract:** the dossier's batch-1 rows, seeded (already, if
fact-verified) or captured, authored via `seed-items`, derived to zero
misses, promoted, and claimed. Create the README's real-catalog
inventory section here; it does not exist yet, and correct the
existing "The capture tool" section's `--out` path, which currently
reads `corpus/cases/<id>` and should read `corpus/staging/<id>` (m5).
**Report actual token spend and the broken-down derive wall time**;
both refine the ceiling and the schedule for Tasks 5 through 15.

**Deliverable count: 8 cases.** Gate plus `make derive` green; commit.

### Tasks 5 through 15: Real-catalog batches 2 through 12

Model: `sonnet`. **8 cases each, 88 total.** Network: yes.

**Files:** as Task 4, minus the README section creation (rows join
the existing section).

**Behavior contract:** identical to Task 4 for each task's own
dossier batch. Each task reports its case count, its network round
count, and its broken-down derive wall time. **A case that cannot be
authored as the dossier describes is reported and replaced from the
dossier's own reserve** (M1, resolving the prior plan's dangling
reference): pull the same-axis reserve row first, report the
substitution in the task return, and report a short batch only if
that axis's reserve is exhausted. The batch's delivered count is
stated honestly either way.

**Deliverable count: 8 cases each.** Gate plus `make derive` green;
commit per task.

---

### Task 16: Adversarial stratum, the string-heuristic classes

Model: `opus` (a case that fails to reproduce its defect is worthless,
and constructing one that genuinely does is the pass's hardest
authoring problem). **10 cases** (`adv-ft-*` 6, `adv-remix-*` 4).
Network: no; these are synthetic, built with `harness synth`.

**Files:**
- Create (via `corpus/staging/`, promoted): 10 cases with their
  `claims.json`
- Modify: `corpus/analysis/ties.json`, `corpus/analysis/proximity.json`,
  `corpus/analysis/oracle_bugs.json`,
  `corpus/analysis/adjudications.json`, `corpus/README.md`

**Behavior contract:**

- **The featuring-pattern false positive, 6 cases.** `SD_PATTERNS`'
  featuring discount has carried no word boundary since its September
  2010 introduction (evolution-lessons commit `742ebdf69`); the
  partial fix landed at 2.13.0 (`7b59604c5`) and the symmetric
  word-boundary fix (`19d091440`) is **still unreleased at pinned
  2.13.1**, which is why this is reproducible here at all. A case
  pairs a title containing a trigger word ("Draft", "Left:", "Gift")
  against a genuinely different candidate that shares enough other
  tokens for the spurious discount to be what tips the distance across
  a recommendation threshold. 2b's `strdist-feat-strip` already proves
  the pipeline mechanics for the *legitimate* strip; these cases are
  the false-positive twin, and the README rows must make the
  difference legible.
- **The parenthetical / remix near-duplicate discount, 4 cases.**
  `SD_PATTERNS`' parenthetical discount (weight 0.3) erases the
  differentiator between a track and its own remix, per
  evolution-lessons' disposition survey (#5868). A case pairs an
  original against a remix whose only distinguishing token sits in the
  discounted parenthetical. This class now has its own seeded
  hypothesis row (Task 1, M7), distinct from the featuring-pattern
  row.
- **Reproduction is the deliverable, not the case file.** For each
  case, state in its `oracle_bugs.json` probe entry and its README row
  what the defect-free distance would have been and what the derived
  distance actually is, so the gap is on the record. A case that does
  not actually move the verdict is a case that failed, and the honest
  outcome is a `reproduced: "no"` hypothesis row with the attempts
  described, never a case reworded to look successful. **Escalation
  trigger:** if two construction attempts on a class fail to move a
  verdict, stop and report rather than spending a third.
- **Every case gets an `adjudications.json` row at
  `pending-adjudication`**, carrying release truth (what the correct
  match is, since these are synthetic and the correct answer is known
  by construction) and the bug-class citation. These rows are part of
  Task 20's batch.

**Deliverable count: 10 cases** plus the two hypothesis rows (ft and
remix/parenthetical, each now genuinely seeded, M7) updated from
`not-attempted`. Gate plus `make derive` green; commit.

---

### Task 17: Adversarial stratum, assignment gaps and alias mis-selection

Model: `sonnet`, with the same two-attempt escalation trigger as Task
16. **10 cases** (`adv-assign-*` 6, `adv-alias-*` 4). **Network: yes**,
for the four alias cases only; the six assignment cases are synthetic.

**Files:** as Task 16.

**Behavior contract:**

- **Per-pair track-assignment rejection gaps, 6 synthetic cases.**
  `assign_items` never rejects a pair; it maps every input track onto
  some candidate track (#2981 open since 2018, #6078 from 2025). A
  case gives the item set the same track count as the candidate, with
  every item title unrelated to its correct track so title contributes
  no signal, and with adjacent-track lengths close enough that the
  solver's optimum picks a plausible-looking but wrong mapping. **The
  sharpest evidence this class produces is that the album-level
  verdict can still land `strong` while individual
  `AppliedFields.title` values are wrong**, so each case's
  `claims.json` pins the wrong assignment explicitly and its README
  row says what the correct one is.
- **Alias and non-Latin mis-selection, 4 captured cases.** Distinct
  from Task 18's #6358 verification, which is one specific regression
  hypothesis. These four probe where alias and transliteration
  selection picks the *wrong* string on real data, rather than merely
  merging a pseudo-release correctly as the existing
  `non-latin-pseudo-release` smoke case does. They capture through
  `harness capture` because real alias arrays are the point; network
  wall time is under a minute.
- **A tie is a real risk in the assignment cases** and is handled by
  the existing machinery, never worked around: the tie detector's row
  carries a `disposition` of `excluded` or `pinned` with a written
  reason, per 2b's standing contract. A tied adversarial case is
  interesting, not broken.

**Deliverable count: 10 cases** plus the two hypothesis rows updated.
Gate plus `make derive` green; commit.

---

### Task 18: The `locale` profile and the #6358 alias-regression verification

Model: `sonnet`. **2 sibling cases** plus one profile. **Network:
yes**, if the host release is not already captured by Tasks 4 through
15; otherwise no.

**Files:**
- Create: `corpus/profiles/locale.yaml`
- Create (via `corpus/staging/`, promoted): 2 cases with their
  `claims.json` (`real-nonlatin-<slug>` under `estate`, and its
  `-locale` sibling)
- Modify: `corpus/analysis/oracle_bugs.json`,
  `corpus/analysis/adjudications.json`, `corpus/analysis/ties.json`,
  `corpus/analysis/proximity.json`, `corpus/README.md` (the config
  profile table gains a `locale.yaml` row)

**Behavior contract:**

- **Read the issue before authoring.** No file in this repo carries
  #6358's body verbatim. Run `gh issue view 6358 --repo
  beetbox/beets` first and record what it actually says, since it
  decides whether the defect is scoped to `beet mbsync` specifically
  or (more likely, per the mechanism below) to the shared
  alias-selection code the importer's own matching also calls. If it
  is `mbsync`-scoped only, that is a finding: write the hypothesis row
  at `reproduced: "not-attempted"` with the scoping reason, skip the
  two cases, and report. Do not force a case against a mechanism the
  issue rules out.
- **The mechanism, already traced.** `_preferred_alias`
  (`beetsplug/musicbrainz.py:128-155`), gated by `import.languages`
  and `import.ignored_alias_types`, rewrites candidate titles, artist
  names, and release titles from the release's `aliases` list **before
  any distance is computed**. `aliases` is already part of
  `RELEASE_INCLUDES`, so every case this corpus captures already
  carries the raw alias data; only the profile needs
  `import.languages` set to exercise the selection.
- **The falsifiable test.** Author one case against a real release
  whose `artist-credit` carries a documented locale-specific alias.
  Derive it twice through sibling case directories: once under
  `estate` (`import.languages: []`, today's default, primary name
  used) and once under `locale` (`import.languages` set to the locale
  the alias targets). If `_preferred_alias` works, the locale
  sibling's candidate title and artist strings, and its
  `AppliedFields`, shift to the alias. If #6358's defect is real at
  2.13.1, they do not, and the divergence is directly visible in the
  two frozen records with no new harness machinery.
- **`locale.yaml` is a new file, not an edit**, so no existing case's
  rendered profile changes and no record is re-frozen. `manifest.json`
  gains a `locale` entry under `profile_hashes` for it. `m1` records
  that today nothing computes or checks `profile_hashes` at all — the
  README's config-profile section currently overclaims it as the
  detection mechanism, and this task drops that overclaim rather than
  inventing hash population inline; the actual reconciliation gap is
  carried to Task 20's HISTORY note as a named follow-on, since
  building it is scoped work in its own right, not a byproduct of
  this task. It is `stock.yaml` plus the `import.languages` and
  `import.ignored_alias_types` settings the test needs, and nothing
  else. The adapter's single-metadata-source check applies to it
  unchanged. The README's config-profile table gains its row with the
  same transform reasoning the other two carry.
- **The outcome is always a hypothesis row, and always a
  `pending-adjudication` row** (M6 — `disposition` is owner-only, so
  neither branch below is written as `oracle-correct` or
  `oracle-wrong` directly, regardless of which way the evidence
  points). If the alias shift happens, the row is `reproduced: "no"`,
  `disposition: "pending-adjudication"`, with the two case ids as
  evidence that the mechanism works at 2.13.1, for Geoff to confirm as
  `oracle-correct`. If it does not, the row is `reproduced: "yes"`,
  `disposition: "pending-adjudication"`, and its adjudication row
  cites ruling 2 as the reason a divergence here would be the correct
  call — conforming to #6358's status quo would degrade the estate's
  own non-Latin material — for Geoff to confirm as `oracle-wrong`.

**Deliverable count: 4** (the `gh issue view` scoping read and its
recorded finding; `locale.yaml` with its README table row; the two
sibling cases; the hypothesis and adjudication rows, both landing at
`pending-adjudication`). Gate plus `make derive` green; commit.

---

### Task 19a: The stock-config stratum, mandatory siblings

Model: `sonnet`. **28 sibling cases** (`<id>-stock`). **Network:
possible**, driven only by misses when stock's verdict sends beets to
a search the estate capture never triggered; budget under a minute.
(M2, split from a single 36-case task per its first offered fix,
matching Tasks 4–15's batch pattern.)

**Files:**
- Create (via `corpus/staging/`, promoted): 28 sibling cases with
  their `claims.json`
- Modify: `corpus/analysis/ties.json`, `corpus/analysis/proximity.json`,
  `corpus/analysis/adjudications.json`, `corpus/README.md`

**Behavior contract:**

- **Sibling case directories, because the schema admits nothing else.**
  `Case.Profile` and `Case.Expected` are single fields, so no case
  directory can carry both an estate record and a stock record. A
  sibling is fully self-contained: its own `case.json`, its own
  `responses/` (the estate twin's response bytes copied verbatim,
  since no cross-case reference mechanism exists), `profile: "stock"`,
  its own `id` suffixed `-stock` (m3), and its own frozen record.
- **Membership is exactly Task 3's 28 mandatory rows — no trimming**
  (M2). Ruling 3 says tier one also holds over the stock config.
  Estate and stock diverge on exactly three keys, checked against
  pinned beets' own `config_default.yaml`: `original_date` (estate
  `yes`, default `no`), `per_disc_numbering` (estate `yes`, default
  `no`), and `none_rec_action` (estate `skip`, default `ask`, which
  moves Action only, never Recommendation). `quiet_fallback`, `timid`,
  and the entire `match:` surface are identical. These 28 rows —
  every real-catalog reissue/remaster case and every multi-disc or
  box-set case — are the only shapes where the two profiles can
  actually disagree on a verdict, which is exactly why Task 3's
  quota and this task's budget were made to match.
- **A divergence is the finding this task exists to produce.** A
  sibling whose recommendation or distance differs from its twin's is
  exactly what ruling 3 is about. Write it into the README row and,
  where the stock verdict is confidently wrong against release truth,
  into `adjudications.json`. A stratum in which nothing ever diverges
  is a stratum that proved something, and the task says so explicitly
  rather than leaving the reader to guess.
- **Action-only divergences are recorded, not ignored.**
  `none_rec_action` moves Action without moving Recommendation, which
  is irrelevant to tier one but is real stock-config behavior. Any
  sibling landing on a `none` recommendation gets its Action
  divergence written into its README row.

**Deliverable count: 28 cases.** Gate plus `make derive` green;
commit.

### Task 19b: The stock-config stratum, supplementary sample

Model: `sonnet`. **8 sibling cases** (`<id>-stock`). **Network:
possible**, same shape as Task 19a. Depends on Task 19a landing first
(shares the same README section and ledgers).

**Files:** as Task 19a.

**Behavior contract:**

- **Supplementary, stratified across the other axes.** Fill 8 with a
  stratified sample of ordinary confident matches drawn from the
  non-mandatory real-catalog axes, which is what gives ruling 3's
  claim breadth (M2) — the mandatory 28 alone prove only the three
  divergent-key axes, never the confident-band majority.
- **Same divergence and action-only recording rules as Task 19a.**

**Deliverable count: 8 cases.** Gate plus `make derive` green;
commit.

---

### Task 20: The pass close and the first adjudication batch

Model: `sonnet`. Network: no.

**Files:**
- Modify: `corpus/README.md` (the full case inventory for everything
  this pass ships — real-catalog, adversarial, and stock-config —
  rewritten whole; the file-tags section stays absent, since Task 13
  has not run)
- Create: `docs/superpowers/research/2026-09-03-adjudication-batch-1.md`
  (the batch prepared for Geoff)
- Modify: `corpus/analysis/divergences.md` (reviewed and carried
  forward, populated or empty)
- Modify: `docs/STATUS.md`, `docs/HISTORY.md` (drafted for the
  conductor's batch commit)

**Behavior contract:**

- **README inventory.** One inventory covering every committed case,
  each row naming its id, stratum, family or axis, profile, and the
  exact behavior it exercises, in the register the existing tables
  already set. Read the committed-case total off the tree rather than
  from this plan's arithmetic. The real-catalog section states plainly
  that its inputs describe real rips and its responses are captured
  bytes, and separately calls out the two-release sub-axis cases and
  what each one's decoy is; the adversarial section states which known
  defect each case probes and what the defect-free result would have
  been.
- **Determinism re-derivation.** Run `make derive` twice from a clean
  tree with `-count=1`, and report the count of cases actually
  re-derived alongside the broken-down wall time (corpus derive, ties
  derive, component derive — no file-tags derive line yet, since Task
  13 has not run). A sub-second run is a cache hit and proves nothing.
  This number is rung 3's budget line exactly as 2b's was 2c's.
- **Corpus byte growth, reported** (m7): state the actual growth
  against the plan's expectation (the committed corpus held 4.0 MB
  across 114 response files before this pass; 96 new captured
  responses at ≈1.4 per case, weighted toward the largest shapes
  (box sets, the two-release sub-axis), plus 36 stock siblings
  duplicating their twin's bytes verbatim, projected a 10–15 MB
  landing point; report the real figure).
- **The first adjudication batch, prepared and waiting.** One
  document, every `pending-adjudication` row from
  `adjudications.json` presented so Geoff can rule from the artifact
  alone without re-deriving anything: the case identity, the oracle's
  own verdict and penalty breakdown, what is suspected wrong and why
  with its citation, and release truth with its citation. Group the
  rows by hypothesis (five hypotheses now, M7) so a single ruling can
  cover a class. **State in the document, in the conductor's own
  words, that nothing here has been self-adjudicated and that every
  row's expected verdict is Geoff's to set.** State also that this
  batch does not block the pass close; the pass closes with these rows
  pending, the gate accepts them, and it is the **launch gate** that
  refuses an unadjudicated divergence.
- **The pass 2d carry-forward note**, addressing M2, M3, M4, M5, and
  m2 in one place so 2d's own plan does not re-derive them:
  - **File-level derivation is 2d's Task 13, unchanged in scope** from
    this plan's original description: `internal/harness/oracle/
    derive_file_tags.py` and its Go driver, reading back independently
    through `mediafile.MediaFile(path)` after `item.write()`, never
    trusting the write call's own return value, over 12 subject cases
    (5 from the committed corpus, 7 flagged by this pass's dossier).
  - **The audio-fixture route is ruled now** (M4), so 2d does not
    decide it inline: vendor one small static template FLAC blob
    under `internal/harness/oracle/testdata/`, committed, with its
    provenance and byte size recorded in the README. `make derive`
    gains no binary dependency; `ffmpeg`/`flac` are not adopted as
    prerequisites.
  - **`derive_file_tags.py` is a second, independent entry point**
    (M5), in the shape of `derive_component.py`: it reads a committed
    record's `AppliedFields`, builds beets items from those fields,
    allocates its own scratch directory, writes and reads back there,
    and is never called from `oracle.Derive`. The invariant: no item
    in the corpus derivation path ever gains a non-empty `path`, and
    the proof is that `make derive` re-derives all committed records
    byte-identically after 2d's change.
  - **`file_tags.json` needs an oracle-tagged re-derive-and-diff
    test** (M3), `TestFileTagsLedgerMatchesTheCorpus`, matching the
    precedent both other derived ledgers set
    (`TestTieLedgerMatchesTheCorpus`,
    `TestComponentTableReDerivesToItsCommittedTable`), plus the
    `-update-ledgers` regeneration path documented in
    `corpus/README.md`'s "Regenerating a derived ledger" section. A
    gate-resident structural check alone (case ids resolve, pointers
    resolve) is not enough; nothing would ever catch a mediafile
    version bump moving the tag shapes without it.
  - **`Reconcile` lives in `internal/harness/oracle/reconcile.go`, not
    `oracle.go`, and its `versionProbe` tuple names five components
    today** (m2: beets, unidecode, jellyfish, lap, numpy), not four.
    2d's Task 13 edits `reconcile.go` and `versionProbe` explicitly
    when it adds `mediafile` to `Manifest`, or a reconcile silently
    compares against a missing key.
  - **The version-bump differential and quarterly re-capture chore
    docs are unwritten**, deferred with Task 13. The differential
    doc's wall-time section must use the M8 inter-request-floor
    formula (`1.04s + 0.25s × (requests − 1)`) at the corpus's real
    size once 2d closes, not a flat per-case figure.
  - **The whole-corpus README rewrite happens again at 2d's close**,
    once, to add the file-tags section; this pass's rewrite (above)
    is not wasted work, since 2d's version extends it rather than
    replacing it.
- **Ledger notes for the conductor.** Draft the HISTORY entry in the
  house shape: what landed, what the gate caught, and what a later
  pass would be wrong to rediscover. Carry forward explicitly: the
  response-status extension, deliberately not taken and why; the 2d
  carry-forward note above in full; the `profile_hashes` gap (m1) as
  a named follow-on distinct from the 2d items; whatever Tier C
  exclusions and profile gaps 2b carried that 2c did not close; and
  the adjudication batch's pending rows as the launch gate's open
  item. Update STATUS's next action to point at pass 2d (file-level
  derivation) and at the pending adjudication batch. **STATUS stays
  present tense** and under 60 lines; anything historical moves to
  HISTORY rather than being summarized harder.

**Deliverable count: 4** (the whole-corpus README inventory with byte
growth reported; the determinism record with its broken-down wall
time; the adjudication batch document; the HISTORY and STATUS drafts,
including the 2d carry-forward note). Gate plus `make derive` green;
commit.

---

## Execution model

Workflow mode per the standing grant: pass-execute chain, sequential.
The pass is sequential by dependency rather than by habit. Task 1
ships the two ledger formats Tasks 16 through 19b all write rows
into. Task 2 ships the `seed-items` tool Task 3's fact-verified-axis
selection and every batch task's item-authoring step both depend on.
Task 3 ships the dossier Tasks 4 through 15 consume, and it exists so
that twelve capture batches never re-litigate selection. Tasks 4
through 15 are the twelve capture batches; none can run in parallel
with another, since all twelve write the same `ties.json`,
`proximity.json`, and README inventory and would contend on all
three. Tasks 16 through 19b write those same files. Task 19a depends
on Tasks 4 through 15, since a stock sibling needs its estate twin's
captured bytes; Task 19b depends on Task 19a (shared README section).

Task 1 and Task 16 dispatch on `opus`; the rest on `sonnet`. Tasks 17
and 18 carry a named escalation trigger in place of an upshift: two
failed construction attempts, or an issue read that rules out the
mechanism, stops the task and reports rather than spending a third
attempt. `diff-reviewer` per task. The conductor dispatches
`code-simplifier` over each accepted task's changed Go and Python
after review, before the pass close, and owns every ledger edit and
the batch commit. Runaway guard armed at launch, per the standing
requirement for anything expected to run long, and more load-bearing
here than in 2b given 21 tasks against 2b's 14.

**Network tasks and their wall-time budgets**, stated once, with the
M8 inter-request floor folded in: Task 3 seeds and verifies roughly
40–50 fact-verified candidates through `harness capture` (the
reissue/remaster, multi-disc/box-set, non-Latin, and two-release
sub-axis rows, plus some rejected candidates), each a single request
at 1/second, well under 2 minutes total; the remaining axes stay on
musicbrainz.org web pages. Tasks 4 through 15 each capture roughly 11
requests for their non-fact-verified rows plus miss-driven follow-ups,
under 30 seconds of network wait per batch. Task 17 captures four
alias cases, under a minute. Task 18 captures at most one release,
seconds. Tasks 19a/19b capture only what a miss list names, under a
minute each. **Total live network wait across the entire pass is
under ten minutes** (roughly doubled from the original six-batch
estimate by Task 3's selection-time captures, M11). The pass's real
wall time is derivation: `make derive` and `-update-ledgers` are both
O(all committed cases), climbing from roughly 4 minutes per task early
in the batch loop toward roughly 9–10 minutes per task by the loop's
end and at Task 20's determinism re-derivation, for an estimated
**110 to 140 minutes** of cumulative derive wall time across the pass
(up from the original ≈90-minute estimate, reflecting twelve
full-derive runs in the batch loop instead of six, M12, priced into
the raised 12M ceiling). Size tasks against that number, never
against the rate limit; refine the estimate from Task 4's actual
report.

Budget: 12M ceiling as a starting figure (M9), re-derived after Task 3
reports its actual token spend and refined again after Task 4's first
batch report; spend scored at close against the re-derived ceiling and
against the one guaranteed touchpoint plus the contingent second
(M9), which together are the target for this pass.
