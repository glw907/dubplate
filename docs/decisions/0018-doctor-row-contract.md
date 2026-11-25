# 18. One doctor row contract

## Status

Accepted

## Context

Record 0013 put each health check in the package that knows its fact. The rows still came back to
`cmd/dubplate-server` in nine shapes: a status string, a status and a failed flag, a `doctorRow`
with a key, a slice with a separate failed result. The command translated each shape and held the
row names itself. The store inspection also read the settings, the roster, the music volume, and
the last backup run, which are tables other packages own (record 0011).

Work `server-boundaries` (task 3 of `docs/specs/2026-11-23-architecture-cleanup.md`) made one shape.

## Decision

Every health check reports `[]doctor.Row` from `internal/doctor`. A row carries its name, its
status text, and whether it fails the run. The package that renders a row names it, and
`cmd/dubplate-server` concatenates the rows in report order and prints them.

A fact that lives in a table is read by the package that owns the table, through a `store.Reader`
it hands to the one `store.Inspect` call. The store package reports only its own facts.

A row never prints the value of a setting registered `config.Sensitive`. The row shows
`(withheld)` in its place, and a refusal keeps its reason with each sensitive value withheld.
`config.Sensitive` marks a value whose change needs care, not a secret, so the rule withholds more
than secrets. Doctor's output is pasted into reports, and nobody should have to scrub it first.

## Consequences

A new health check adds one `DoctorRows` function in its package's `doctor.go` and one argument to
the command's `slices.Concat`. Its tests live with it, against literal `doctor.Row` values.

A refusal's text must quote every value it names with `%q`. The withholding matches that form, so
a correlated rule that prints a value bare would leak it. Every registered rule quotes its values
today.

`check` still prints a refused value in full: it is not a doctor row.
