# 10. Configuration homes

## Status

Accepted

## Context

Ease of setup by a stranger is a primary product goal (commit `006f2d4`, 2026-08-31), so nothing in
the code may assume one group's site. The homes for settings and collections are ratified in
`docs/specs/2026-09-11-configuration-homes-design.md`, and the per-package settings rule landed with
work `structure-keys` (commit `8e92787`, 2026-11-20).

## Decision

No site-specific value, path, or identity default exists outside `internal/config`.

A setting is declared in `config.go` of the package that reads it and added to that package's
`Register` function. `internal/config/keys.go` takes no new declaration. `cmd/dubplate-server`
gains a call in `newProductionRegistry` only when a new package registers for the first time. In
`auth`, `backup`, `serve`, `store`, `webapp`, `intake`, and `ops/local`, only `config.go` and
`doctor.go` import `internal/config` or `internal/settings`.

A collection is a store row whose root the server allocates from its id, and no two roots nest.

## Consequences

`TestConfigurationStaysInConfigFiles` in `internal/settings/layering_test.go` fails when any other
file in those seven packages imports configuration. The configuration homes spec is a living
reference, edited in place by the task that changes what it says.
