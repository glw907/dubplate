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

**Blocked: on-box cloud-init verification (Task 4a steps 4-5).** `ssh
root@5.78.148.42` authenticates fine via the workstation pubkey cloud-init installed, but
Hetzner's own out-of-band root-password injection (separate from cloud-init; visible as
"Root password: ..." in `provision.sh`'s create output, not stored anywhere) carries a
forced-expiry flag. PAM's account-management phase enforces the change on *every* SSH
session post-auth, regardless of the pubkey login succeeding and regardless of
`PasswordAuthentication no` already being live in `sshd_config.d` — that setting governs
the authentication phase, not the post-auth `chauthtok` account check. A non-interactive
session (`ssh root@IP 'cmd'`, or `-tt` fed via a pipe) gets "Password change required but
no TTY available" and closes; the Claude Code harness's safety classifier separately
declines to let this agent complete the interactive change itself (any command touching
the remote root credential — `chage -l`, piping a new password into a `-tt` session, even
`hcloud ssh-key list` — is denied), correctly, since an agent silently rewriting root
credentials on a live box deserves a human in the loop.

**One-time manual unblock (30 seconds, from a real terminal, not this harness):**
```
ssh root@5.78.148.42
```
It authenticates via key immediately, then PAM will prompt `New password:` /
`Retype new password:` — no *old* password is needed (root changing its own password
doesn't require it). Answer both with anything; the session then drops and a fresh `ssh
root@5.78.148.42` works as a normal, no-prompt root shell going forward, since only the
one-time chage state was blocking it. After that, resume exactly at plan Task 4a step 4:
`ssh -o StrictHostKeyChecking=accept-new root@5.78.148.42 'cloud-init status --wait'`,
then the step-5 on-box checks (mount, ownership, `music` uid, docker, SELinux enforcing,
sshd passwordauthentication, yq/jq, fstab), then finish this STATUS entry and the Task 4a
commit.

Consider folding a `chage -d $(date +%s) -M 99999 root`-equivalent (or an explicit
`chpasswd: {expire: false}` — verify it actually reaches Hetzner's injected password, not
just cloud-init-managed ones) into `cloud-init/user-data.yaml`'s `runcmd` so a future
reprovision (this box, the Task 7 drill server, any rebuild) doesn't hit this again.

Next task: finish Task 4a verification (blocked above), then continue Task 4/5.

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
- **Hetzner forced root-password-expiry blocks non-interactive SSH post-provision**: see
  the Task 4a blocker in Current state; needs one manual interactive `ssh` from a real
  terminal before automated verification/deploy can proceed.
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
