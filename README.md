# dubplate

dubplate is a music server for a group of people who share a library. Members upload
purchases through a web inbox. The server checks and tags each album, files it into a
canonical library, and streams it. rclone backs the library and state up to Cloudflare R2.
Everything runs on one Hetzner CPX21 server as the `dubplate-server` binary under systemd,
provisioned from this repo and restorable from R2 alone.

For setup, deploys, and operations, see `docs/server-runbook.md`. The binary's design:
`docs/specs/2026-08-30-dubplate-go-rewrite-design.md`. The original estate build
design stays with the estate history in `~/.dotfiles/docs/specs/` (live
initiatives migrate into this repo, history stays put).

See `docs/STATUS.md` for the current state of the build.

## Kept identifiers

The estate renamed musicbox to dubplate 2026-08-31, but a handful of identifiers keep
their old name deliberately, because they name live infrastructure rather than the
project: the Hetzner server (`musicbox`), firewall (`musicbox-fw`), data volume
(`musicbox-data`), and ssh key (`musicbox-workstation`); and `scripts/deploy.sh`'s
`hcloud server ip musicbox` lookup. `docs/HISTORY.md` also keeps its name. History is
written as it happened.
