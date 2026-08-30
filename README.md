# musicbox

musicbox is the repo for the family music pipeline: family members upload purchases to a
web inbox, headless beets imports them into a canonical library, Navidrome streams the
result at `music.907.life`, and rclone backs the library and state up to Cloudflare R2.
Everything runs on one Hetzner CPX21 server, provisioned from this repo and restorable
from R2 alone. Full design: `~/.dotfiles/docs/superpowers/specs/2026-08-30-music-vps-build-design.md`.

See `docs/STATUS.md` for the current state of the build.
