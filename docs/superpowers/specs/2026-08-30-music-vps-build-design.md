# Music VPS build (musicbox)

One VPS hosts the whole music pipeline: family members upload purchases to a web inbox,
headless beets imports them into the canonical library, Navidrome streams the result at
`music.907.life`, and rclone backs everything up to R2. This supersedes the PikaPods-pod
half of `2026-08-30-music-library-design.md`; the beets conventions, the R2 bucket, the
listening clients, and the one-gatekeeper rule carry over unchanged. Immich stays on
PikaPods; this box serves music only.

The master library lives on the VPS at `/srv/music/library`. The workstation is a
contributor (it uploads to the inbox) and may keep a read-only mirror via rclone pull.
Nothing but the VPS writes to the library or to R2. There is no data migration: `~/Music`
is empty today, so cutover is only a matter of pointing scripts at the new layout before
the first import.

Revised 2026-08-30 after adversarial review; the review's findings drive the inbox choice,
the volume, the permissions model, the backup layout, and the import-script hardening below.

## Infrastructure

- **Provider and size**: Hetzner Cloud CPX21 (3 vCPU, 4 GB RAM) in Hillsboro, Oregon, the
  closest region to Alaska. US pricing runs about $12/mo with the IPv4 surcharge, and US
  instances carry much smaller traffic allowances than EU (confirm the exact quota in the
  console at provision time; if it is 1 TB, default family clients to transcoded streams).
- **Storage**: a Hetzner volume (50 GB initially, ~$2.50/mo, resizable online) mounted as
  `/srv/music`. The 80 GB root disk holds only OS and images; the library, inbox, review,
  and state never compete with the OS for space. `music-import` aborts below a free-space
  threshold and pings a failure check, because a full disk during import means truncated
  files that a later backup would faithfully replicate.
- **OS**: AlmaLinux 10, provisioned by cloud-init (SSH key, Docker CE from its RHEL
  repo, dnf-automatic for security updates). Chosen for Red Hat-family consistency with
  the Fedora-based workstation; Fedora Server was rejected for its 13-month lifecycle on
  a box meant to idle for years, and Fedora CoreOS because Hetzner has no official image.
  SELinux stays enforcing; container bind mounts carry `:Z` labels. SSH is key-only; a
  Hetzner Cloud firewall admits only port 22. Routing SSH
  through the tunnel was considered and rejected: it adds client-side friction to every
  rsync and rclone call for marginal gain over key-only auth.
- **Users and permissions**: one `music` system user owns `/srv/music`. The import timer
  runs as it, both containers run with its UID/GID, and the inbox is setgid so uploads,
  imports, and interactive sessions never fight over ownership.
- **Services**, all Docker Compose with pinned image tags, updated deliberately (no
  auto-updater):
  - Navidrome (official image, includes ffmpeg). `ND_PASSWORDENCRYPTIONKEY` comes from the
    age store, because Subsonic token auth forces the server to keep passwords recoverable;
    family members are told this password must be unique to the service.
  - FileBrowser Quantum (`gtsteffaniak/filebrowser`, the maintained fork; the original
    filebrowser was archived 2026-09 with unpatched CVEs) as the upload inbox, one account
    per contributor, each scoped to `/srv/music/inbox/<user>/`.
  - cloudflared running a Cloudflare Tunnel; no inbound HTTP ports on the box.
- **Ingress**: the tunnel publishes `music.907.life` (Navidrome) and `inbox.907.life`
  (FileBrowser Quantum). Subsonic clients authenticate against Navidrome directly, so no
  Cloudflare Access on `music.907.life`. The inbox gets a **Cloudflare Access application**
  (one-time PIN to the listed family emails, works in iPad Safari) in front of the app's
  own login: contributors are a small known set, and Access removes the entire
  unauthenticated attack surface from a file-manager app class with a poor CVE history.
- **Repo**: `~/Projects/musicbox`, holding the compose file, service config, cloud-init
  user data, the import and backup scripts, and a `deploy.sh` that rsyncs config to the box
  and applies it. The box is cattle; the claim is proven by a restore drill (below), not
  asserted.

## Import pipeline (the hosted gatekeeper)

A systemd timer runs `music-import` (as the `music` user, under `flock` so a run never
overlaps another run or an interactive session) every 15 minutes:

1. Skip any inbox directory whose **recursive** newest mtime is under 10 minutes old
   (`find -newermt`), so a half-uploaded album is never split; a directory's own mtime
   does not see nested writes, so the check must recurse.
2. Expand archives (`bsdtar` with a size cap and path-traversal guard, since the input is
   untrusted) into a staging directory. Bandcamp delivers ZIPs; without this step the
   iPad contributor's uploads dead-end.
3. Abort if volume free space is below threshold; ping the failure check.
4. Quality gate (Geoff, 2026-08-30: only high-quality FLAC enters the library).
   Every extracted audio file must be FLAC, pass `flac -t` (full decode test,
   catching corrupt or truncated uploads), and carry at least 44.1 kHz / 16-bit.
   Failures move to `review/` with an email to the uploader naming the file and
   the reason; nothing is silently discarded, and a lossy-only album is Geoff's
   call in review. Spectral fake-FLAC detection is out of scope until it shows up.
5. Run `beet import -q -l <logfile>` over the staging directories. Quiet mode
   auto-accepts confident MusicBrainz matches; the logfile is the mechanism that names
   what was skipped (weak match, duplicate, junk).
6. Move logged skips to `/srv/music/review/` (admin-only; contributors cannot fix a bad
   MusicBrainz match, and a shared review folder would let any contributor delete
   another's files). Email the uploader and Geoff via Fastmail SMTP (msmtp; a small
   contributor-to-email map in config). A retention sweep expires review items after 60
   days so duplicates do not accumulate.
7. Trigger a Navidrome scan through its API. The filesystem watcher is disabled (it would
   index albums mid-move) and `Scanner.Schedule` is set as a daily backstop only.

Expectation set honestly: Bandcamp self-released material is often absent from
MusicBrainz, so the auto-accept rate may be low at first. The design degrades gracefully:
worst case it is a notified manual queue that Geoff drains over SSH with interactive
`beet import` (through the same lock, via a `music-beet` wrapper), which still beats the
workstation-gatekeeper model on availability. Measure the accept rate over the first
month before adding any cleverness.

beets runs on the host (installed with uv) rather than in a container, because the same
instance and database (`/srv/music/state/library.db`) serve both the timer and Geoff's
interactive sessions.

## Backup

Nightly, and structured to survive both a dead box and a compromised one:

1. `sqlite3 <db> ".backup"` copies of the beets library DB and Navidrome's data DB (users,
   playlists, favorites, play counts — the only state not reconstructible from the FLACs)
   into a staging dir. rclone must never read live SQLite files; a mid-write copy is a
   torn, unopenable backup.
2. Per-tree `rclone sync` runs into a **pinned bucket layout**: `library/` and `state/`
   prefixes in `music-library`, each with `--backup-dir` under `_archive/<tree>/<date>/`
   and `--exclude /_archive/**`. One sync per tree; a single root-level sync would have
   each run delete the other's objects, and rclone refuses an overlapping backup-dir
   outright.
3. An R2 **object lifecycle rule** expires `_archive/` after 90 days (unbounded dated
   archives would silently outgrow the cost estimate on the first mass retag), and an R2
   **bucket lock** on `_archive/` prevents the box's own credentials from shortening or
   deleting the archive: a compromised box can corrupt tomorrow's backup but not erase
   yesterday's.
4. Success pings healthchecks.io.

At build time, check whether R2 object versioning has shipped; if it has, prefer it over
the `--backup-dir` scheme.

**Restore drill (acceptance gate)**: before the family is onboarded, provision a scratch
VPS from the repo, restore from R2, and confirm Navidrome comes up with accounts and
playlists intact. Then delete the scratch box.

## Monitoring

Four healthchecks.io checks (free tier): backup success, import-timer success, volume
free space, and a Navidrome HTTP heartbeat. A missed ping emails Geoff. Absence plan: if
the importer stops while Geoff is away for two weeks, uploads queue safely in the inbox
and import on his return; contributors get their review emails either way, and nothing
is lost.

## Secrets

All minted values land in the workstation age store as origin of record, with scope and
rotation in `registry.md`; the box holds copies in `/etc/musicbox/env` (root, 0600),
provisioned by `deploy.sh`.

- Hetzner API token: minted at account creation (manual step), provisioning only.
- R2 tokens, both bucket-scoped to `music-library`, both dashboard-minted: a read-write
  token for the box (backup) and a separate **read-only** token for the workstation
  (mirror pull). The workstation never holds delete-capable credentials, so no
  argument-order slip in a pull script can wipe the backup.
- Cloudflare Tunnel token, Navidrome `ND_PASSWORDENCRYPTIONKEY`, Fastmail SMTP app
  password, healthchecks ping URLs.

## Workstation changes

`music-sync` (laptop-pushes-to-replicas) was retired 2026-08-30: it is dead code from the
superseded PikaPods architecture, and running it would `rclone sync` `~/Music` onto the
`music-library` bucket root, deleting the box's backups. Its replacement, `music-pull`
(`rclone copy`, never sync, of `library/` to `~/Music` with the read-only token, for local
listening), has not been built yet. Geoff contributes either through the web inbox like everyone
else or by `rsync` over SSH into his inbox directory; a dedicated `music-drop` SFTP
account was considered and cut (rsync over the SSH access he already has is the whole
feature). The beets stow package remains as the layout/plugin reference the VPS config
derives from; the workstation beets no longer masters anything.

## Costs

About $12/mo VPS + $2.50/mo volume + ~$1/mo R2: call it $15–16/mo, roughly double the
PikaPods-pod plan it replaces; the difference buys the hosted gatekeeper and the custom
domain. healthchecks.io and Cloudflare Access are free at this scale.

## Out of scope

Immich migration, transcoded-mirror schemes, playlist import tooling, per-client
Navidrome API keys (revisit when the pinned version supports them), and any import
cleverness beyond the pipeline above until a month of real accept-rate data exists.
