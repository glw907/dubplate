# Photo curation spec, revision 2.1: prose review

Reviewer: `prose-voice-reviewer` (Opus 5.5), 2026-09-25. Register: Google Developer Documentation
Style. Saved by the conductor from the agent's report; line numbers refer to revision 2.1.

tellgrader (docs register): 0 findings, 237 sentences, 4040 words, cadence CV 0.58. Vale covers
only Go comments in this repo, so no Vale package ran on the spec.

Facts verified correct: the 62/15/6/30 of 113 table; about 48MB from 164KB times 300; the
British Culture Archive quote; Beckman's rank; the research template's "non-commercial" line;
`DUBPLATE_HARNESS_CONTACT` at `cmd/harness/capture.go:26`; the CC 4.0 and 2.0/3.0 clause
wording; the identity spec's "100 or more", "unmodified", and "early hip-hop and go-go"; the
±3 band and 15% posed figures agree everywhere.

## Blockers

- B1 (:213-217, :185): the embed pattern `photos/*.webp photos/*.json` contradicts "the
  embedding package must sit inside `photos/`". Per `go doc embed`, patterns "are interpreted
  relative to the package directory containing the source file", so a package in `photos/`
  embeds `*.webp *.json`. :185's "any `//go:embed photos` directory pattern" assumes the parent.
- B2 (:358 vs :262-265): Out of scope says the pass touches nothing under `scripts/`, but
  Process queues the `check.sh` integrity leg. Nothing says whether `check.mjs` survives rung 7's
  Go gate leg.

## Warnings

- W1: `pack` is set but no field table defines it; `status` allows only `picked` or `withdrawn`.
- W2 (:133 vs :313): "at pick time" conflicts with the slug being minted at round-close step 3.
- W3 (:55-58, :296-297, :315): waivers have no key shape; scope omits the region cap, posed
  ceiling, and surface floors; a sorting-page keep supplies no reason though one is required.
- W4 (:160, :252, :308-309): `withdrawn` needs a reason and date that no field holds; Licensing
  omits Geoff's change of mind as a cause.
- W5 (:186-187, :278-279): no reason-code list for `rejected.json`, and agent drops never reach
  the ledger, so later rounds re-find them.
- W6 (:315-316): "Each pass" reads as the initiative, not the `N` action.
- W7 (:24-25 vs :171, :216): the quoted ~45MB and the measured 48MB appear with nothing linking
  them; the 45MB quote and the 30/20/20/30 ruling are not logged in the fold record.
- W8 (:168): thumbnail size and whether they are committed are unstated.
- W9 (:291-303): the sorting-page paragraph uses undefined artifact terms and packs four topics.
- W10 (:18-20, :362-375): the Changes list misses the US public-domain tightening, the widened
  core families, and the license URL moving into the license list.
- W11 (:44-46): the coverage-floor sentence hides a 3 by 5 grid.
- W12 (:115 vs :152): the caption prints a license name that the license list does not hold.

## Suggestions

- S1: split the license-evidence paragraph at "On Commons".
- S2: "period photographs" conflicts with recent landmark photos; the landmark sentence belongs
  under Place.
- S3: two passives (:55, :207).
- S4 (:393): "the cited period examples" lost its referent; state what the floor keeps (77 of
  113, 6 more by exception, 30 need a larger scan).
- S5: one name each for the coverage grid and the region cap.
- S6: step 5 should say how a withdrawal's deletion is committed.
- S7: "valuble" is verbatim; leave it.
- S8 (:253-255): spell out why a blob in git history is acceptable.
- S9: ragged wraps at :106, :209, :243.

Verdict: the register reads as a plausible human author with healthy cadence. B1, B2, and W1
through W5 must close before the spec is built from.
