# Fold notes for the greenfield build spec review round

Items raised in session (Geoff + Claude, 2026-09-02) after the spec
went under adversarial review, held here so the fold works from an
artifact rather than conversation memory. Each is folded into the
spec alongside the review findings, not before.

1. **Component-test seeding from beets' distance tests.** The
   evolvability ruling shapes the Go distance model as small named
   penalty components; beets' `test/autotag/test_distance.py` cases
   map onto per-component Go table tests, a fast test altitude
   beneath the corpus's end-to-end verdicts. Expected values derived
   by executing beets' distance code, never transcribed. Target:
   rung 3's plan (and rung 2's, for derivation tooling).

2. **Write-side tag conformance.** The corpus as specified proves
   verdicts; nothing proves written-tag parity, though the success
   bar's "core tagging" covers both. Proposal: extend the corpus
   case's expected side to verdict plus the applied tag set (FLAC
   Vorbis comments, server scope), derived by running pinned beets
   and capturing what it writes. Leverage: the extracted `mediafile`
   library's field-mapping tests and beets' apply-metadata tests as
   scenario sources. The write side has had no close source read yet
   (the mechanics study scoped to the matcher); rung 8 planning owes
   one. Target: the spec's evidence-track wording, rung 2 (corpus
   shape) and rung 8 (tag I/O conformance).

3. **Adversarial-stratum prior from beets' own suite.** Scenario
   classes beets tests heavily are classes that burned them; use the
   suite's emphasis as a weighting prior when sizing the adversarial
   stratum at rung 2. Minor; planning input only.
