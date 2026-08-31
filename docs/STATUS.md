# musicbox status

## Current state

Task 3 (compose.yaml, systemd units, deploy.sh) is done: `compose.yaml`, `config/`
(navidrome.toml, filebrowser.yaml, msmtprc, beets-config.yaml,
cloudflared-tunnel-config.json), the `music-import`/`music-backup` systemd units plus a
shared `musicbox-ping-failure@.service`, `env/musicbox.env.template`, and
`scripts/deploy.sh` all land. `bash scripts/check.sh` is green: shellcheck, the 16-test
bats suite, `docker compose config -q`, and a pattern-filtered `systemd-analyze verify`
(plain `--recursive-errors=no` doesn't suppress the workstation's missing-ExecStart-binary
noise; check.sh filters exactly that message and fails on anything else — confirmed
against a real injected syntax error). Full rationale for every Task 3 decision:
`docs/HISTORY.md`'s Task 3 entry.

Task 4a (real provisioning) ran 2026-08-30: `scripts/provision.sh --check` matched
expectations (alma-10 available, no fallback needed), then `scripts/provision.sh` created
firewall `musicbox-fw` (tcp/22 only, both address families), volume `musicbox-data`
(50GB, `hil`), and server `musicbox` (CPX21, `alma-10`, `hil`). **Server IPv4: `5.78.148.42`**
(IPv6 `2a01:4ff:1f0:97c6::1`). CPX21 included traffic at `hil`: **2199023255552 bytes = 2
TiB/month** (`hcloud server-type describe cpx21 -o json`, `.prices[] | select(.location ==
"hil") | .included_traffic`; overage `$1.20`/TB). 2 TiB is generous for a family Navidrome
— no need to default clients to transcoded streams.

**Fix round 2026-08-31 closed three blocking findings** from the forced-root-password
blocker above, and the server was recreated to pick them up:

1. `scripts/provision.sh` now registers an `hcloud ssh-key` named `musicbox-workstation`
   idempotently (checks `hcloud ssh-key list` first, creates it from
   `~/.ssh/id_ed25519.pub` — the same key already embedded in
   `cloud-init/user-data.yaml` — only if missing) and passes `--ssh-key
   musicbox-workstation` to `hcloud server create`. Confirmed on the recreate: no "Root
   password: ..." line appeared in the create output at all, so there was no forced
   password to expire this time.
2. `cloud-init/user-data.yaml`'s `runcmd` now carries a defense-in-depth guard —
   `chage -M -1 -I -1 -E -1 -d "$(date +%Y-%m-%d)" root` — as the very first item, so any
   future path that ever creates the server without a key can't reintroduce this blocker.
   Confirmed on-box: `chage -l root` shows `Password expires: never` and `Account
   expires: never`.
3. The one-time-manual-SSH-unblock instruction that stood here is removed; nothing was
   hand-applied to a live box for this round — the box was deleted and recreated from the
   fixed repo instead (`hcloud server delete musicbox`, volume `musicbox-data` survived
   and reattached, `blkid` guard prevented reformat, box got IPv4 `5.78.148.42` again).

**New blocker found during re-verification, unrelated to the three fixes above:**
`cloud-init status --wait` now returns `status: error` (`Runparts: 1 failures (runcmd) in
1 attempted commands`). `journalctl -xeu docker.service` shows dockerd failing to start:
`iptables ... Extension addrtype revision 0 not supported, missing kernel module?` —
`xt_addrtype` isn't present (`modinfo xt_addrtype` → not found) on this box's kernel
(`6.12.0-211.43.1.el10_2.x86_64`). This is a confirmed, known upstream regression,
[docker/docker-ce-packaging#1340](https://github.com/docker/docker-ce-packaging/issues/1340):
`iptables-nft` on RHEL/AlmaLinux 10 no longer pulls in `kernel-modules-extra` (which ships
`xt_addrtype.ko`) as a dependency, so a fresh Docker install on a minimal AlmaLinux 10
cloud image fails to start the daemon. Because cloud-init concatenates every `runcmd` item
into one `set -e` shell script, the docker block's failure aborted everything after it —
the yq-install item never ran either. **This is deterministic, not transient: it will
recur on every AlmaLinux-10-based rebuild, including Task 7's drill server**, until
`cloud-init/user-data.yaml` installs `kernel-modules-extra` before the docker `dnf
install`. Per this round's instructions ("if cloud-init failed, gather the log tail
instead of improvising fixes on the box"), this fix-round did not attempt that change —
it's scoped to the three findings above — and left the box running in this state rather
than hand-patching it live.

**Task 4a step 4-5 evidence, captured 2026-08-31 against the recreated box
(`root@5.78.148.42`):**

| Check | Command | Result |
|---|---|---|
| cloud-init | `cloud-init status --wait` | `status: error` — **FAIL**, see blocker above |
| volume mount | `mountpoint -q /srv/music` | mounted — **PASS** |
| tree ownership | `ls -lnd /srv/music/{library,inbox,review,state,staging}` | all `2000 2000`; `inbox` is `drwxrwsr-x` (2775, setgid) — **PASS** |
| music user | `id music` | `uid=2000(music) gid=2000(music) groups=2000(music)` — **PASS** |
| SELinux | `getenforce` | `Permissive` — **FAIL**. Not caused by cloud-init, which never calls `setenforce` or touches `/etc/selinux/config` (confirmed unchanged, still just `#cloud-config` + `runcmd`); the Hetzner `alma-10` image itself ships `/etc/selinux/config` with `SELINUX=permissive`. Needs its own decision (force enforcing in cloud-init, or update the acceptance criterion) in a future round. |
| docker | `systemctl is-active docker` | `failed` — **FAIL**, see blocker above |
| sshd password auth | `sshd -T \| grep -i '^passwordauthentication'` | `passwordauthentication no` — **PASS** |
| yq | `/usr/local/bin/yq --version` | binary missing (its runcmd item never ran — aborted by the docker failure above) — **FAIL** |
| jq | `jq --version` | present and functional (`echo '{"a":1}' \| jq .a` → `1`); the version string itself prints as bare `jq-` on this AlmaLinux 10 build, a cosmetic upstream packaging quirk, not a functional defect — **PASS** |
| fstab | `grep srv/music /etc/fstab` | exactly one line: `UUID=6614e3dd-9014-4949-9f4c-effcd317bd3a /srv/music ext4 defaults,nofail,x-systemd.device-timeout=30s 0 2` — **PASS** |

**Task 4a remains open: 5 of 8 on-box checks are green.** The three named blocking
findings from the previous round are closed and verified (ssh-key at create time, the
`chage` defense-in-depth guard, no more hand-edit instruction). The two reds (docker,
SELinux) trace to the base AlmaLinux 10 Hetzner image, not to anything touched in this
round, and are new open items below.

Next task: a fix round for the `kernel-modules-extra` / docker blocker (and a decision on
the `getenforce: Permissive` base-image default), then re-verify Task 4a steps 4-5, then
continue Task 4/5.

Plan: `~/.dotfiles/docs/superpowers/plans/2026-08-30-music-vps-build.md`
Spec: `~/.dotfiles/docs/superpowers/specs/2026-08-30-music-vps-build-design.md`

## Pinned versions (looked up 2026-08-30)

| Component | Image / package | Tag / version | Released |
|---|---|---|---|
| Navidrome | `deluan/navidrome` | `0.63.2` | 2026-07-11 |
| FileBrowser Quantum | `gtstef/filebrowser` (upstream repo name `gtsteffaniak/filebrowser`) | `1.5.5-stable` | 2026-08-30 |
| cloudflared | `cloudflare/cloudflared` | `2026.8.2` | 2026-08-14 |
| beets (box, via uv) | PyPI `beets` | `2.13.1` | current at lookup |
| yq (box, cloud-init) | GitHub `mikefarah/yq` (`yq_linux_amd64`) | `v4.53.6` (sha256 `c5f056448f973ae7d39b5401949648a78f2dc1947d6a8eb65be60d5c504b9385`) | current at lookup |

Re-verify all five before Task 4's real deploy if this pass is more than a few days
stale — `docker compose config` never checks that a pinned tag still exists upstream, and
cloud-init's `sha256sum -c` check will hard-fail first boot if the pin drifts from a
yanked or replaced release.

## Open items

- **Task 0 (Geoff, manual, blocking Task 4+)**: credential batch not landed. `deploy.sh`
  already fails correctly, listing every missing env var by name.
- **CPX21 US traffic allowance**: confirmed 2026-08-30, 2 TiB/month at `hil` (see Current
  state above) — plenty; no transcode-by-default needed.
- **`alma-10` image availability**: confirmed available on Hetzner as of 2026-08-30;
  `provision.sh` used it directly, no `alma-9` fallback triggered.
- **Hetzner forced root-password-expiry**: resolved 2026-08-31 — `provision.sh` attaches
  an SSH key at server-create time so Hetzner never mints the password, plus a `chage`
  guard in cloud-init as defense in depth. See Current state.
- **Docker fails to start on this AlmaLinux 10 image (blocking Task 4a)**: `xt_addrtype`
  kernel module missing (`kernel-modules-extra` no longer a dependency of `iptables-nft`
  on RHEL/AlmaLinux 10 — [docker/docker-ce-packaging#1340](https://github.com/docker/docker-ce-packaging/issues/1340)).
  Fix is to add `kernel-modules-extra` to `cloud-init/user-data.yaml`'s `packages:` (or a
  `runcmd` install) before the docker block. Also aborts the yq-install runcmd item that
  follows it. See Current state for full evidence.
- **`getenforce` reports `Permissive`, not `Enforcing`, on a stock `alma-10` Hetzner
  image**: cloud-init never touches SELinux config (confirmed); this is the image's own
  default. Needs a decision: force enforcing via cloud-init, or update Task 4a's
  acceptance criterion to match the image's actual default.
- **cloudflared ingress is not set anywhere in compose.yaml on purpose**: a token-run
  tunnel is remotely managed, so a local ingress file would be silently ignored.
  `config/cloudflared-tunnel-config.json` is the exact API body Task 5 must PUT to the
  tunnel once it exists — skipping that step leaves cloudflared connected but serving
  nothing.
- **MUSICBOX_HC_DISK_URL / MUSICBOX_HC_NAVIDROME_URL are named but unwired**: only
  `import` and `backup` have real `ping_hc` calls today. Task 6 needs new disk-space and
  Navidrome-heartbeat check scripts + timers, not just the healthchecks.io checks
  themselves.
- **Ruling**: "no `latest`, pin explicitly" governs container tags and installed package
  versions (beets); host RPMs (`dnf-automatic`) and `uv` itself are intentional
  exceptions, not violations.
