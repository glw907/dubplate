# Architecture refinement pass: charter (musicbox estate)

**Status: charter for a planned pass.** Written 2026-08-30 at Geoff's direction; the
pass gets its full plan (and the plan its adversarial review) at kickoff. This
document scopes the intent so kickoff starts from a ratified frame.

## Intent

An adversarial pass over the musicbox estate *as actually built* — the VPS
(cloud-init, compose, systemd, scripts), the import pipeline, the pings Worker, the
workstation-side scripts, and all the docs — hunting cleanup and refinement rather
than new capability. Pass 1 was built in one day under heavy fix-round pressure;
every fix round leaves scar tissue: patches shaped by the finding they closed rather
than the design as a whole, duplicated idioms that drifted apart, docs that describe
intent where reality moved, and tests that pin yesterday's shape.

## Hunting grounds

- **Simplification and consolidation**: repeated shell idioms that should be lib
  functions; near-duplicate blocks in deploy.sh flagged by earlier simplifier runs;
  fix-round patches that would be written differently designed whole.
- **Drift**: spec vs plan vs STATUS vs the running box — every claim in a doc that a
  fresh reader would act on, checked against reality. (Tonight produced several
  corrected-late examples: phantom API fields, stale package-list claims.)
- **Security posture, second look**: the whole surface reviewed as one system now
  that it exists — secrets flow, SELinux contexts, systemd hardening directives not
  yet applied (ProtectSystem, NoNewPrivileges), the Access/tunnel boundary, R2 token
  scopes in practice.
- **Altitude**: what got built into musicbox that is really estate-level (candidate
  moves into shared tooling or the pings repo), and vice versa.
- **Test-suite quality**: suites grew accretively across fix rounds; look for
  redundancy, gaps the rounds revealed as classes (falsy-zero, fail-open,
  set -e-in-pipeline), and whether fixtures still earn their runtime.
- **Operational truth**: runbooks and drill docs executed skeptically; the
  known-limitation list re-examined for entries that quietly became load-bearing.

## Shape (decided at kickoff, but the expected form)

Review-first, fix-second: an adversarial multi-lens fan-out (the dimensions above)
producing verified, ranked findings; then a fix wave sized by what survives triage —
not a rolling fix-as-you-find, which is how scar tissue formed in the first place.
Findings that belong to pass 2's surface get tagged for it rather than fixed twice.

## Sequencing

Geoff's call at pass-1 close: before the upload-check pass (refine the substrate
first) or after it (review the final shape once, including FileBrowser's removal).
The charter takes no position beyond noting the trade.

## Out of scope

New features, the upload-check work itself, Immich, and anything the refinement
findings would merely make *nicer* rather than simpler, truer, or safer.
