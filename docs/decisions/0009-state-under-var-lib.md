# 9. State under /var/lib/dubplate

## Status

Accepted

## Context

systemd deletes a unit's `RuntimeDirectory` when the unit stops, so state kept there is lost on
every restart. The rule landed when the incumbent stack was retired from the repo (commit
`d1e93e9`, 2026-10-05).

## Decision

dubplate-server's state lives under `/var/lib/dubplate/`, which persists, and never in a
`RuntimeDirectory`. The server uses no runtime directory at all.

Upload staging is the exception by design: it lives on the music volume beside the library, so a
rename into the library never crosses a device.

## Consequences

`scripts/check-state-dir.sh`, a leg of the full gate, fails when the env template, a
`dubplate-server` unit, and `deploy.sh` disagree on the state directory and its leaves, and checks
the music volume's one grant on the serve unit.
