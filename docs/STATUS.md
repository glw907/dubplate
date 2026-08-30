# musicbox status

## Current state

Task 2 (import, backup, and wrapper scripts, TDD with bats) is done: `scripts/lib.sh`,
`music-import`, `music-backup`, `music-beet`, `scripts/check.sh` all land with
`tests/lib.bats` and `tests/music-import.bats` green (16 tests). `bash scripts/check.sh`
is the gate; it requires `shellcheck`, `bats-core`, and `bsdtar` (all installed via
`brew install bats-core libarchive` this pass) on the workstation. No infrastructure
exists yet: `provision.sh` has not been run for real, so there is no server, no volume,
and no IP to record.

Next task: Task 3 (compose.yaml, systemd units, deploy.sh). See `docs/HISTORY.md`'s
Task 2 entry for decisions Task 3 must reconcile (Navidrome scan API assumption, state
DB paths, the new `yq` host dependency).

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
- **Ruling**: the global "no `latest`, pin explicitly" constraint governs container image
  tags (Task 3's `compose.yaml`), not host RPMs. `dnf-automatic` manages host package
  versions via security-only auto-updates by design (cloud-init/user-data.yaml); that is
  intentional drift, not a constraint violation.
