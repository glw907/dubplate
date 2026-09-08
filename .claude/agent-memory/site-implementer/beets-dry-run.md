---
name: beets-dry-run
description: Upstream beets 2.13 facts for running the matcher without side effects, verified by experiment against the pinned version
metadata:
  type: reference
---

Verified against beets 2.13.1 (the version `scripts/deploy.sh` pins) on 2026-09-01,
by reading site-packages and running the real binary:

- `beet import --pretend` is NOT a matcher dry run. Pretend replaces the whole
  lookup pipeline with a file listing, so it never contacts MusicBrainz. The only
  dry matcher run is a real `beet import -q` into a throwaway library with
  `copy/move/link/hardlink/reflink/write` all off.
- `beet -c FILE` is an overlay, not a replacement: confuse `set_file` prepends the
  file above the user config. Only one `-c` is accepted.
- `BEETSDIR` replaces the user config directory outright, so staging a config there
  makes a run hermetic against the ambient `$HOME`. Confirmed: keys from the staged
  config still show in `beet config`, with the `-c` overlay winning per key.
- `raise_on_error: yes` (a top-level key) turns a metadata-source failure into a
  nonzero exit instead of a logged warning plus zero candidates. Without it a
  MusicBrainz outage is indistinguishable from "no match".
- Quiet mode prints no candidate detail on a skip, only the path header and
  "Skipping." There is no cheap best-guess to parse out of it.
- `create_backup_before_migrations` defaults to yes and writes eleven `.bak` files
  on every brand new library database. Turn it off for scratch libraries.
- Since beets 2.4 `musicbrainz` is a plugin and must be listed in `plugins:` or the
  autotagger has zero sources.

The consumer is `internal/beets.Runner.Probe`; see [[systemd-python-subprocess]] for
what the hardened serve unit needs to run it.
