# HISTORY draft: rung 2, pass 2d (2026-09-04)

Drafted by pass 2d's Task 4 for the conductor's batch commit into
`docs/HISTORY.md`. Not committed to the ledger directly; the conductor
folds this in and retires the draft file, matching 2c's own Task 20
close (`docs/HISTORY.md`'s "Close pass 2c: fold the Task 20 ledger
drafts, retire the draft files").

---

## Rung 2, pass 2d (File-level applied tags, the version-bump differential, and the whole-corpus README close): file_tags.json over 12 subjects, rung 8 unblocked (2026-09-04)

Four plan tasks executed sequentially against
`docs/superpowers/plans/2026-09-03-rung-2d-file-level.md` (revision 2).
Commits in order: `8d181fe` (Task 3, which ran first per the plan's
sequencing), then `4d38717` (Task 1), then `f5abba0` (Task 2), then
`412471f` plus `cc400ad` (Task 4's close and its same-session
forward-reference fix).

What landed: the template FLAC fixture and its provenance script
(`internal/harness/oracle/testdata/template.flac`, 42 bytes, silent,
untagged, authored via the pure-`mutagen` route with no `ffmpeg`/`flac`
dependency), `derive_file_tags.py` and its Go driver implementing the
write-then-independent-read-back contract with `BEETSDIR` isolation
proved rather than assumed, and `Manifest.MediaFile` pinning
`mediafile==0.17.0` at `schema_version` `3` with `Reconcile`'s `pinned`
tuple and `versionProbe` string both gaining their sixth component
(Task 1); `corpus/analysis/file_tags.json`, generated over a verified
twelve-subject set, its pure-Go gate-resident checker
(`internal/harness/analysis/file_tags.go`) wired into `corpus_test.go`
with three firing fixtures (an unresolved case id, a value key outside
the writable-set header, and — the fixture with real teeth — a row
carrying one of the nine read-only keys), and
`TestFileTagsLedgerMatchesTheCorpus` on the
`TestComponentTableReDerivesToItsCommittedTable` template (Task 2); the
version-bump differential job doc and the quarterly re-capture chore
doc, both standing operational docs beside `docs/bringup-runbook.md`,
the differential doc's wall-time section priced at the real 215-case
corpus with the file-tags line left as a named placeholder for this
task (Task 3); the whole-corpus case-inventory section (215 rows
across five stratum tables, built by a mechanical script pass over
each case's own `id`/`profile`/`eval_date` cross-referenced against
the family/task/adversarial/stock-config tables already committed,
never by re-reading a case body) and the rung 8 handoff statement,
both new `corpus/README.md` sections; the file-tags line written back
into the differential doc's placeholder (Task 4). Corpus total: still
**215 cases**, unchanged from 2c's close — 2d authors no case, captures
nothing, and edits no committed record, exactly as its own global
constraint required.

**The twelfth file-tags subject is a ruled substitution, and this
pass's own staleness catch.** The plan's Task 2 named
`real-decoy-beethoven-symphony-5` as the two-release sub-axis's
file-tags subject, drawn from the planning inventory dossier written
before 2c executed. That id was never captured: 2c's own Task 4 had
already proved the decoy unreachable inside beets' 5-result search
window with the axis's sole reserve spent, closing the axis at 7-of-8
(`corpus/README.md`, "The two-release sub-axis finding"). Task 2's
implementer caught this with a STOP rather than padding or trimming
the twelve-subject set, and the conductor ruled a substitution:
`real-decoy-mozart-requiem`, chosen from the axis's four other
committed decoy cases by eligibility (all four carry a non-empty
`applied_fields` record regardless of `action`, since
`match.apply_metadata()` runs whenever a candidate exists at all) and
then by the most distinctive applied-tag shape among the eligible rows
(24 items, two movement-numbering conventions, two distinct
`albumartist_credit` strings). Recorded with its eligibility evidence
in `corpus/README.md`'s file-tags section. **The standing lesson: a
plan drafted before a prior pass closes must re-verify its id lists
against the tree at execution time, not trust the dossier that
predated the pass's own findings.** This is not a new failure mode —
2c's own Task 17 alias-duplicate incident carries the sibling lesson
("check `corpus/cases/` before any capture") — but this pass's version
of it is a plan citing a specific id the immediately prior pass had
already retired, which a same-pass cross-reference (rather than a
later implementer's own tree check) could in principle have caught
earlier. It did not block the pass; the implementer's own verification
step, required by Task 2's contract regardless, is what caught it.

What the gate caught: nothing novel — `bash scripts/check.sh` and
`GOFLAGS=-count=1 make derive` stayed green through all four tasks,
and the pass's own fixtures (Task 1's silent-direction `Reconcile`
proof, Task 2's three gate-resident checker fixtures plus its
re-derive-catches-a-perturbation fixture) are what proved each new
check has teeth rather than the gate itself surfacing a defect in
existing code.

Determinism: two independent uncached (`GOFLAGS=-count=1`) full
`make derive` runs against the whole `internal/harness/...` and
`cmd/harness/...` tree with the file-level derivation in the tree, 215
and 215 cases actually re-derived of 215 committed
(`TestCorpusReDerivesToItsCommittedRecords`'s own per-case subtests),
both runs exiting 0 with zero failures across every package, both
byte-identical to the committed records (corpus, ties, and component
ledgers, plus the twelve-subject file-tags ledger — `git status`
reported no diff under `corpus/` after either run). Wall time: run 1
980s (16m20s) total (corpus derive 462.94s, ties derive 503.90s,
component derive 0.75s, file-tags derive 0.59s); run 2 974s (16m14s)
total (corpus derive 541.20s, ties derive 420.06s, component derive
0.71s, file-tags derive 0.55s) — this is rung 3's budget line. 2d's
pair (980s, 974s) ran materially faster than 2c's own pair
(~1729.58s/1582.53s) on the identical 215-case corpus, roughly 40
percent faster; the spread is attributable to run-to-run variance in
request pacing rather than any corpus change, and the file-tags leg's
sub-second cost (0.55s-0.59s) is negligible under either figure. The
file-tags leg issues no MusicBrainz
request (twelve subjects, one Python process reading already-committed
`AppliedFields`), so it adds well under one second against the corpus
and ties legs' own multi-hundred-second cost; the measured figure
(0.55s-0.59s) is written into
`docs/corpus-version-bump-differential.md`'s previously-placeholder
row, alongside a new finding that doc records: the corpus and ties
legs' own M8-formula floor (257.6s each) is roughly half the real
measured wall time (462-541s and 420-504s respectively) — the formula
prices only the client's inter-request wait, not the Python/beets/
response-parsing overhead sitting on top of it, so a future
version-bump differential run should budget against the measured
figure, not the formula table.

What a later pass would be wrong to rediscover:

- **The rung 8 handoff statement lives in `corpus/README.md`** ("##
  Handoff to rung 8: what pure-Go FLAC tag writing may assume"),
  positioned immediately after "## The analyses" so rung 8's
  implementer finds it beside the evidence rather than in a fifth
  loose document. It states the two altitudes of expected applied
  tags (`Expected.AppliedFields` on every case, `file_tags.json` on
  twelve subjects), names the 69-key writable set as the only
  conformance target and the 9-key read-only remainder as never one,
  and states its own non-scope explicitly: the apply-metadata field
  mapping close read and the wazero-taglib spike are rung 8's own
  plan-opening work, not something 2d evaluated or prototyped.
- **2c's `profile_hashes` gap (m1) is still open**: `corpus/manifest.json`
  carries a `profile_hashes` entry per profile, but nothing in the gate
  or the derivation path computes or checks it for any profile yet.
  2d did not touch this; it is still scoped follow-on work, not
  something the manifest's own presence implies is built.
- **The response-status extension is still deliberately not taken.**
  Carried from 2c: the stub has no third outcome and the schema
  carries no response status, since no known-oracle-bug class probed
  so far needs a 404 or 503 to reproduce. Available to whichever later
  rung first has a case that genuinely needs it.
- **The Tier C exclusions and profile gaps 2b and 2c both carried,
  neither pass closed, carry forward again unchanged**: multi-source
  ordering and most of the `data_source` parameter space stay blocked
  by the single-source ruling; `match.required`/`match.ignored`
  candidate-dropping stays unprobed; `preferred.original_year: yes`
  stays unreachable under `estate`, `stock`, and `locale` alike (a
  fourth profile or a `locale.yaml` extension is still owed if a later
  rung wants this branch probed). 2d's own scope never touched any of
  these; nothing new closed here either.
- **The pending adjudication batch is still the launch gate's open
  item**, unchanged by this pass: nineteen rows across five hypotheses
  sit at `disposition: "pending-adjudication"` in
  `corpus/analysis/adjudications.json`
  (`docs/superpowers/research/2026-09-03-adjudication-batch-1.md`).
  2d's own gate accepts them indefinitely, per its own "Owner
  touchpoints: none" scoping; it is the launch gate, later, that
  refuses an unadjudicated divergence. 2d produced no new
  adjudication row of its own.
- **The `schema_version` `2` -> `3` bump forced no re-freeze**, the one
  place this pass's schema event differs from 2b's: `Manifest` is
  corpus metadata rather than an `Expected` field, so `Manifest.MediaFile`
  landing needed no re-derivation of any committed case's own record,
  only the two-clean-run byte-identical proof above.

Carried:

- Rung 2 is complete at this pass's close: the smoke, structural,
  real-catalog, adversarial, and stock-config strata (215 cases), the
  file-level applied-tag ledger, and the two standing chore docs are
  all shipped. Rung 3 (the Go matcher) is what rung 2's completion
  unblocks, gated on Geoff reading its revision-2 spec before
  execution starts (`docs/superpowers/specs/2026-09-03-rung-3-go-matcher.md`).
- Every item 2c's own carry-forward left open (`profile_hashes`, the
  response-status extension, Tier C exclusions, the pending
  adjudication batch) stays open into rung 3 and beyond, restated
  above rather than re-derived.

Budgets: total subagent token spend across Tasks 1-4 (conductor to
fill in from its own per-task tracking; this implementer has no
visibility into cross-task cumulative spend) against the 2.5M starting
ceiling, re-derived after Tasks 1 and 2 per the plan's own token
ceiling section. Attended interaction points: zero, matching the
pass's own target ("Owner touchpoints: none").
