# 15. Reject wording is contract

## Status

Accepted

## Context

A member reads why an upload failed on the uploader page and again in an email. Two wordings for one
reason confuse them. The rule dates from the start of the dubplate era (commit `006f2d4`,
2026-08-31), and the copy is fixed in `docs/research/passb-copy-final.md`.

## Decision

Reject wording (`friendly_reason`) is a contract. The uploader page and the emails share it word for
word, and upload-check reuses it. The set lives in `internal/reject`.

## Consequences

`TestAdoptedReasonsEqualPCF` in `internal/reject/reject_test.go` fails when an adopted reason's
label, friendly text, or next step drifts from the copy contract.
