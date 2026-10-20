# Music library and family streaming

**Status: superseded by `2026-08-30-music-vps-build-design.md`.** This PikaPods-era design
is not the current architecture; read the VPS build design instead.

The music system masters a FLAC library on the workstation, streams it to family through a
Navidrome pod on PikaPods, and backs it up to R2. The laptop copy in `~/Music` is the single
master; every replica is a one-way push, so no merge case exists. Expected size is 100 to 200
albums, mostly from digital purchases (Bandcamp and similar), not CD rips.

## Components

- **beets** masters the library. Installed with `uv tool install "beets[fetchart,embedart]"`;
  config is the `beets` stow package (`beets/.config/beets/config.yaml`). Imports move files
  into `~/Music` as `Artist/Album/Track` FLAC, tagged against MusicBrainz, with album art
  fetched and embedded. Library state lives in `~/.local/share/beets/` (outside `~/.config`,
  so stow never folds it into the repo).
- **Navidrome on PikaPods** serves the family, alongside the existing Immich pod. Size the
  pod for about 100 GB. Each family member gets their own Navidrome account, because
  playlists, favorites, and play state are per user. PikaPods has no API or CLI; pod
  lifecycle is a dashboard task, but the recurring work (uploads) runs over SFTP.
- **R2 bucket `music-library`** (created 2026-08-30) is the durable backup, following the
  per-purpose bucket convention.
- **`music-sync`** (`bin` stow package) pushes `~/Music` to both replicas with rclone. It
  reads credentials from `~/.local/secrets` and skips any replica whose credentials are
  absent, so it works before the pod exists. Run it after each import session.

## Listening

Family members install a Subsonic client (Symfonium on Android, Amperfy or play:Sub on iOS),
point it at the pod URL, and sign in. The clients provide offline downloads, gapless
playback, and CarPlay/Android Auto. The house iPad runs Amperfy with the library cached
offline and feeds the stereo over aux or AirPlay; that cache is the local mirror, so house
playback survives an internet outage without home server hardware.

## Secrets

Two credentials remain to be minted, both through the age-store flow (`secret-set.sh`, then
`registry.md`):

- `MUSIC_R2_ACCESS_KEY_ID` / `MUSIC_R2_SECRET_ACCESS_KEY`: an R2 API token scoped to the
  `music-library` bucket. The CLAUDE_CODE token cannot mint tokens, so this is a dashboard
  step.
- `MUSIC_POD_SFTP_HOST` / `MUSIC_POD_SFTP_USER` / `MUSIC_POD_SFTP_PASSWORD`: from the pod's
  SFTP settings after creation.

## Decisions and alternatives

PikaPods over a self-managed VPS: for a single Navidrome instance, VPS control buys a custom
domain and vendor independence, not capability, and the PikaPods account already runs Immich.
Migration path if it chafes: rclone the same library to a VPS and repoint the clients. Plex
was rejected as too heavy for music-only serving; a serverless Workers build was rejected
because native Subsonic clients would have to be built from scratch.

## Contributions from family

Other family members also add music, including one contributing from an iPad. To keep the
organization consistent, nobody writes into the library or the pod directly: all music
enters `~/Music` through `beet import` on the workstation, which applies the same tagging
and layout regardless of who bought the album. Contributors drop raw purchase files into an
upload inbox (proposed: a Filebrowser pod on PikaPods, which gives them drag-and-drop web
upload from any browser, including iPad Safari, and gives the workstation SFTP access for
rclone). The import session pulls the inbox, imports, clears it, and syncs. A contributor
with a real computer may run beets locally as a pre-tagging convenience, but the
workstation import remains the gatekeeper, and only the workstation syncs to the replicas
(a second syncing master would delete the other's albums).

## Import workflow

1. Download the purchase (FLAC preferred), or pull family uploads from the inbox.
2. `beet import <download-dir>` (moves files into `~/Music`, tags, fetches art).
3. `music-sync`.
