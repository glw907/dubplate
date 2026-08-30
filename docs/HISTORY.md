# musicbox history

Per-pass ledger, newest first.

## Task 2: import, backup, and wrapper scripts (2026-08-30)

Landed `scripts/lib.sh` (`quiescent`, `safe_extract`, `free_space_ok`, `notify`,
`ping_hc`), TDD with `tests/lib.bats` written and confirmed red before any
implementation existed. Landed `music-import`, `music-backup`, `music-beet`, and
`scripts/check.sh` as the repo gate (shellcheck over every `scripts/*` bash file,
then the full bats suite). `tests/music-import.bats` runs the importer end to end
against a fixture `/srv/music` tree with `beet`, `curl`, and `msmtp` stubbed on
PATH: happy path, a beets skip filed to review with both notifications, a
non-quiescent inbox left untouched, and a free-space abort that fires the fail
ping before beet ever runs.

Two one-line review cleanups from Task 1 landed alongside: the volume glob guard
in `cloud-init/user-data.yaml` now also rejects a non-block-device match (a
zero-match glob previously fell through as a false positive on some shells), and
`provision.sh`'s firewall-rule check now fails with its own message when the
`hcloud firewall describe` call itself errors, instead of reporting a network
blip as a missing rule.

Decisions a later pass would be wrong to rediscover from scratch:
- **Contributor matching**: `music-import` matches an inbox directory name
  (`/srv/music/inbox/<user>`) against contributors.yaml's `name` field verbatim;
  there is no separate username/slug field. If Task 3/8's FileBrowser account
  provisioning wants a different inbox directory naming scheme, `contributors.yaml`
  needs a dedicated field and `uploader_email()` needs updating to match on it.
- **Navidrome scan trigger**: implemented against the native admin API (`POST
  /auth/login` for a JWT, then `POST /api/scanner/scan` with `x-nd-authorization:
  Bearer <token>`), not the Subsonic API, because Subsonic has no scan-trigger
  endpoint. This is the UI's internal API rather than a documented stable
  contract; verify it against the pinned Navidrome image tag once Task 3 chooses
  one, since a trigger failure is only best-effort (logged, not fatal) precisely
  because this endpoint is not guaranteed stable.
- **State DB paths default to** `$MUSICBOX_ROOT/state/library.db` (beets) and
  `$MUSICBOX_ROOT/state/navidrome/navidrome.db` (Navidrome), both overridable via
  env var. Task 3's `compose.yaml` must either mount Navidrome's data dir there or
  the env var overridden in `/etc/musicbox/env`.
- **Archive cap default**: `MUSICBOX_MAX_ARCHIVE_BYTES` defaults to 2 GiB in
  `music-import`; revisit if a real Bandcamp download exceeds it.
- New host dependency: `yq` (mikefarah/go, v4) for parsing `contributors.yaml`,
  needed on the box alongside `beet`/`rclone`/`msmtp`/`bsdtar` — add it to Task 3's
  `deploy.sh` package list.
