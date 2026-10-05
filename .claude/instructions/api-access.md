# API access inventory (dubplate)

- **Cloudflare**: `$CLOUDFLARE_API_TOKEN` (a script must
  `source ~/.local/secrets`). Full-account scope; the canonical inventory is
  `~/.claude/docs/cloudflare-estate-inventory.md`. Used here for: R2
  (`music-library` bucket). The tunnel and its DNS were deleted 2026-10-05
  with the retired incumbent stack. Access/Workers-domain writes need curl,
  not the MCP plugin.
- **Hetzner**: token sourced the same way; consumed by
  `scripts/provision.sh` (server create, volume, firewall). Firewall changes
  are API-reversible from the workstation — that is the SSH-lockout
  recovery path.
- **The box**: `root@5.78.148.42` over SSH (key auth). Production — check
  for live executors and prefer `deploy.sh` over ad-hoc edits. Read-only
  ssh is fine for reviews.
- **pings Worker** (`~/Projects/pings`): admin API at pings.907.life with
  `$PINGS_ADMIN_TOKEN`; per-check ping URLs are capability tokens minted
  into the age store as `DUBPLATE_PING_*`.
- **Secrets flow**: age store is the origin
  (`~/.dotfiles/scripts/secrets/secret-set.sh`, registry in
  `~/.dotfiles/secrets/registry.md`); `deploy.sh` renders the box's
  env file from it. Never commit values; never print
  them.
