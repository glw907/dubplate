# musicbox status

## Current state

Task 1 (repo scaffold and provisioning inputs) is done. The repo holds the cloud-init
user-data, the `provision.sh` script, and the ledger files. No infrastructure exists yet:
`provision.sh` has not been run for real, so there is no server, no volume, and no IP to
record.

Next task: Task 2 (import, backup, and wrapper scripts, TDD with bats).

Plan: `~/.dotfiles/docs/superpowers/plans/2026-08-30-music-vps-build.md`
Spec: `~/.dotfiles/docs/superpowers/specs/2026-08-30-music-vps-build-design.md`

## Open items

- **Task 0 (Geoff, manual, blocking Task 4+)**: credential batch not yet landed.
  `HCLOUD_TOKEN`, both R2 tokens (RW for the box, RO for the workstation), the Fastmail
  SMTP app password, the healthchecks.io API key, and the contributor list all still need
  to go through `secret-set.sh` and `registry.md`. `config/contributors.yaml` in this repo
  is a placeholder schema until that list lands.
- **CPX21 US traffic allowance**: unconfirmed. Record the console-shown quota in this file
  before Task 4 (spec: US instances carry smaller allowances than EU; if it is 1 TB,
  default family clients to transcoded streams).
- **`alma-10` image availability**: unconfirmed on Hetzner. `provision.sh` queries
  `hcloud image list` at run time and falls back to `alma-9` with a loud NOTE if `alma-10`
  is absent. Record which image actually got used here once Task 4 runs.
- **Pinned image tags**: Navidrome and FileBrowser Quantum compose image tags are pinned
  in Task 3, not yet chosen. Record the exact tags here once that task lands.

## Recent history

Moved to `docs/HISTORY.md` once a second pass lands; nothing to move yet.
