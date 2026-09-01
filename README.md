# dubplate

dubplate is the repo for the family music pipeline: family members upload purchases to a
web inbox, headless beets imports them into a canonical library, Navidrome streams the
result at `music.907.life`, and rclone backs the library and state up to Cloudflare R2.
Everything runs on one Hetzner CPX21 server, provisioned from this repo and restorable
from R2 alone. The binary's design: `docs/superpowers/specs/2026-08-30-dubplate-go-rewrite-design.md`;
the original estate build design stays with the estate history in
`~/.dotfiles/docs/superpowers/specs/` (live initiatives migrate into this
repo, history stays put).

See `docs/STATUS.md` for the current state of the build.

## Kept identifiers

The estate renamed musicbox to dubplate 2026-08-31, but a handful of identifiers keep
their old name deliberately, because they name live infrastructure rather than the
project: the Hetzner server (`musicbox`), firewall (`musicbox-fw`), data volume
(`musicbox-data`), and ssh key (`musicbox-workstation`); the Cloudflare tunnel; the compose project's `name: musicbox`
(container-orphan protection — see its comment in `compose.yaml`); the
`/var/lib/musicbox` and `/srv/music` paths on the box (live data); and
`scripts/deploy.sh`'s `hcloud server ip musicbox` lookup. `docs/HISTORY.md` also keeps
its name — history is written as it happened.
