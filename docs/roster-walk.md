# The roster walk

The roster walk drives each client in the roster through the same list of flows against a live server and records what
the client shows. Each slice from 3 to 8 walks the scripts below against dubplate-server; slice 2b walked them once
against the Navidrome and gonic oracles, and the run record is
`docs/superpowers/research/2026-10-10-slice-2b-roster-walk.md`.

The roster is Feishin 1.17.0 and Rufin 0.16.6 (Linux desktop, driven on the workstation), Tempus 4.28.0 (Android, needs
the emulator), and Amperfy 2.1.1 (iOS, walked at rung 16's attended gate). The pins and the clean-room rule come from
`2026-10-06-slice-1-evidence-pins.md`. Nothing here quotes a client's or an oracle's source.

References: [clients] is `docs/superpowers/research/2026-10-06-slice-1-client-demands.md`, [base-url] is
`2026-10-06-slice-1-base-url.md`, and [pins] is `2026-10-06-slice-1-evidence-pins.md`, all in the same directory.

## Method

The method is [base-url]'s "Isolation and the profile hashes", with the changes noted below.

**Stack.** `harness oracle-up` builds the seed fixture library, starts the pinned Navidrome 0.64.2 and gonic v0.22.0
containers on it, prints each oracle's loopback URL, user name, and password, and holds until interrupted. A walk types
those values into the client's add-server form and nowhere else. They are never written to the record, a capture, a
log excerpt, or a command line in a committed file. A walk against dubplate-server uses the server's own login the same
way.

**Isolation.** Each client runs under an isolated configuration home inside its Flatpak sandbox:

- `flatpak run` with `--filesystem=<isolated directory>`, `--no-talk-name=org.freedesktop.secrets`, and
  `--no-talk-name=org.kde.kwalletd5`, started as `--command=sh` so that `HOME`, `XDG_CONFIG_HOME`, `XDG_DATA_HOME`,
  `XDG_CACHE_HOME`, and `XDG_STATE_HOME` are exported inside the sandbox, because `--env` does not override Flatpak's
  own `XDG_*` values.
- Feishin also takes `--password-store=basic`, so its credential store stays in the isolated directory and not the host
  keyring.
- The client runs on a private display with `--socket=x11 --nosocket=wayland`, so no window or input reaches the
  desktop. Feishin also takes `--ozone-platform=x11` (without it Electron exits with a failure to connect to a Wayland
  display) and `--disable-gpu`.
- Audio goes to a throwaway null sink (`pactl load-module module-null-sink`, selected by `PULSE_SINK` inside the
  sandbox), unloaded at the end, so a playback step makes no sound on the desktop.

**Display.** No host Xvfb exists, so the private display is a throwaway container, as [base-url] did. The container
runs `Xvfb :<n> -screen 0 1440x900x24 -ac` with `--network=host`; the client reaches it through the abstract X socket,
which a host-network container shares with a Flatpak sandbox that has the network. The image and its digest are named in
the run record. Screenshots are taken with `import -window root` inside the container, so a capture is exactly the
client's window (the display holds one window and no desktop). Feishin's captures are Playwright 1.63.0 screenshots over
its remote debugging port, which the same window yields.

**Feishin's debugging port.** Feishin starts with `--remote-debugging-port=0`; the port Chromium chose is read from
`DevToolsActivePort` in the isolated directory, and the repo-pinned Playwright 1.63.0 (`playwright-core`) attaches over
that port with `connectOverCDP`. A DOM capture is the page's `innerText`. A scripted step never removes nodes React owns
(doing so unmounts the app).

**Hashes, before and after, no change allowed.** Before the first launch and after the last client exits, SHA-256 every
file under each client's real `~/.var/app/<id>/` tree except `cache/` and Flatpak's `.ld.so/` linker cache, listing
directories and symlinks beside it, and every file under `~/.local/share/keyrings/`. Any difference means the isolation
failed: stop, report it, and do not record a pass. Both apps' real trees are hashed as found, including Feishin's
`org.jeffvli.feishin` tree that slice 1's first launch created.

**Captures.** Every pass or fail claim cites a screenshot or a DOM capture. A committed screenshot is cropped to the
client window (a no-op on a private display) and read for host paths, user names, and credentials before staging; a
capture that shows a user name or a credential is not committed. The record notes a step whose evidence comes from
reading the oracle (a `curl` against the oracle's API with the run's credentials, output reduced to counts and titles).

**Teardown.** Stop every client, unload the null sink, remove the Xvfb container and any image built for it, interrupt
`oracle-up` (it removes the oracle containers), and confirm `podman ps -a --filter label=dubplate-harness-run` is empty.

**Rules for the record.** A step records pass or fail with the observation. A step that fails against an oracle is
evidence about that oracle: record it, do not work around it. A client that will not start: record its error and walk
the other. A failed `oracle-up` stops the walk.

## The flow list

[clients] section 2 is a per-client table of endpoints, with no list of flows, so this list is the checklist the walk
and its reviewer use. It is cross-checked against [clients] 2a to 2d (what each client calls to do the flow) and
against the folder-view demand in [clients] section 4.

| # | Flow | What a client calls for it (from [clients]) |
|---|---|---|
| F1 | Add the server as the OpenSubsonic type | `ping`, `getUser`, `getOpenSubsonicExtensions` (Feishin 2a); `ping`, `getUser`, `getScanStatus`, `getMusicFolders` (Rufin 2b); `ping` (Amperfy 2c, Tempus 2d) |
| F2 | Log in: the client lands on a working library | the F1 calls plus the first catalog calls: `getAlbumList2` (all four), `getArtists` (Feishin, Amperfy, Tempus), `search3` with an empty query (Rufin) |
| F3 | Browse artists by ID3 | `getArtists` and `getArtist` (Feishin, Amperfy, Tempus); Rufin's artist list comes from `search3` |
| F4 | Open the multi-disc album | `getAlbum`: the disc number must group the tracks (Feishin 2a, Amperfy 2c, Tempus 2d); Rufin reads its local catalog built from `search3` |
| F5 | Open the compilation | `getAlbum` for an album whose artist is Various Artists |
| F6 | Play a track | `stream` (URL, not an envelope; [clients] section 5); `scrobble` (Rufin, Amperfy, Tempus, Feishin) |
| F7 | Show cover art | `getCoverArt` ([clients] section 7) |
| F8 | Open the folder view | `getMusicFolders`, `getIndexes`, `getMusicDirectory` ([clients] section 4); all four clients have a folder view |
| F9 | Search | `search3` (all four) |
| F10 | Star a track | `star` / `unstar`; `getStarred` (Feishin) or `getStarred2` (Rufin, Amperfy, Tempus) |
| F11 | Create a playlist | `createPlaylist`, `getPlaylists`; add a track with `updatePlaylist` (Rufin 2b, Amperfy 2c) |

The folder view (F8) is the one flow whose expected answer differs between the oracles and dubplate. Under ruling C the
0.8 answer is empty or minimal, never an error: [clients] section 4 reads the per-client effect of an empty
`getMusicFolders` as Tempus hiding its folders sector, Feishin keeping the Folders item with an empty or failing page,
Rufin keeping the Folders route (its catalog runs unscoped), and Amperfy showing "No Directories". A walk against an
oracle records what the oracle's folder view shows; a walk against dubplate checks the answer against that reading.

The fixture facts the scripts use, from the seed library: "Double Feature" (Borealis Trio, 1998) has four tracks over two
discs (Disc 1: Overture, Interlude; Disc 2: Reprise, Finale); "Mixtape Volume One" (Various Artists, 1999) is the
compilation, with three tracks (Opening Remix, Hello Again, Snow Song); the artists "Alpha Artist" and "Zoë Ångström"
hold the search and the non-ASCII cases. Every track is two seconds long.

## Script: Feishin 1.17.0

Launch as in Method. Capture the page at the end of every step.

| Step | Action | Expected observation | Record |
|---|---|---|---|
| F1 | On the "Server required" form choose OpenSubsonic, enter a name, the URL, and the run's user name and password, and submit. Dismiss the release-notes dialog first. | The form closes and the app leaves `#/action-required`. | The route after submit; whether any toast appeared. Do not capture the filled form. |
| F2 | Wait for Home. | Home shows a hero album, the sidebar (Albums, Tracks, Album Artists, Artists, Genres, Folders, Radio Stations), and "Explore from your library". | Screenshot and DOM of Home; the playlists, if any, in the sidebar. |
| F3 | Open Artists in the sidebar. | An artist grid with a count badge equal to the number of artists the server lists. | The count and the first names. |
| F4 | Open Albums, then "Double Feature". | A header with 4 tracks and 8s, and the track list grouped under "Disc 1" and "Disc 2". | The group headers and the track rows. |
| F5 | Open Albums, then "Mixtape Volume One". | A header naming Various Artists, 3 tracks, and a flat list of three rows. | The header and the rows. |
| F6 | Double-click "Overture" on the Double Feature page. | The player bar shows the track and a pause control; the page title shows the queue position. | The audio elements' `currentSrc` (credentials removed), `currentTime`, `duration`, `paused`, and `error`, from the DOM, beside a screenshot. |
| F7 | Look at the album header, the grids, and the player bar. | The covers render; an album with no embedded art shows the placeholder. | Which covers rendered and which showed a placeholder. |
| F8 | Open Folders in the sidebar. Against an oracle, double-click a folder row. Open the server menu to see the music-folder picker. | A Folders page with a count badge. A folder row opens the folders below it. | The rows, the count, the picker's entries, and any error toast. The expected answer on dubplate is an empty list or an empty picker. |
| F9 | Click the sidebar Search, type "Alpha". | A dialog with Albums, Album Artists, and Tracks sections. | The sections and the first rows of each. |
| F10 | On the compilation page, hover "Hello Again" and click the heart. | The heart fills. | Screenshot; then read the server's starred list to confirm. |
| F11 | Click "+" beside Playlists, enter a name, create. | A "Playlist created successfully" toast and the playlist in the sidebar. | Screenshot, and the server's playlist list. |

## Script: Rufin 0.16.6

Launch as in Method on a private display. The display has no window manager, so the window fills it. Rufin opens an
album or artist on a double-click; a single click only raises the hover controls.

| Step | Action | Expected observation | Record |
|---|---|---|---|
| F1 | In the Add Source dialog choose OpenSubsonic, enter a name, the URL, user name, and password. Click the password field before pressing Return, or Return in the user name field submits early. | "Connecting to music server..." then the dialog closes. | The log's authentication warnings, if any. Do not capture the filled form. |
| F2 | Wait for Home. | A Showcase album, "Explore", and "Newly added". | Screenshot of Home. |
| F3 | Open Artists. | An artist grid. Allow the catalog to finish loading; reopen the page if a spinner stays. | The tiles. |
| F4 | Open Albums, double-click "Double Feature". | A header with 4 tracks and 8s, grouped under "Disc 1" and "Disc 2". | The groups and rows. |
| F5 | Open Albums, double-click "Mixtape Volume One". | A header naming Various Artists, 3 tracks, 6s. | The header and the rows. |
| F6 | Double-click "Overture". | The player bar shows a title, a pause control, and elapsed time; the queue panel fills. | Screenshot; elapsed over total time. |
| F7 | Look at the grids, the album header, and the player bar. | Covers render; albums without art show the placeholder. | Which rendered. |
| F8 | Open Folders in the sidebar. Open the source menu to see the Server Library entries. | A Folders page with rows. | The rows and the library entries. The expected answer on dubplate is an empty list. |
| F9 | Click the top Search, type "Alpha", press Return for "View all". | Tracks, Albums, and Artists tabs; the Tracks tab lists the matches. | The dropdown, then the full list. |
| F10 | In the search results hover a track and click its heart; open Favorites. | The heart fills and the track is listed under Favorites. | Both screenshots; the server's starred list. |
| F11 | Open Playlists, click "+", enter a name, Create; then open a track's menu, "Add to Playlist", tick the playlist, Add. | The playlist appears with 0 tracks, then 1 track. | Both screenshots; the server's playlist list (name and song count). |

## Script: Tempus 4.28.0 (from source reading, unproven)

This script comes from [clients] 2d and section 4, with no run behind it. Mark every expectation unproven until the
emulator walk confirms it. The walk is deferred: `/dev/kvm` does not exist on the workstation (checked 2026-10-10 with
`ls /dev/kvm`: no such file).

| Step | Action | Expected observation (unproven) | Record |
|---|---|---|---|
| F1 | On the login screen enter the server address, user name, and password. | The address check accepts an http(s) URL with a host; `ping` runs and a failed status shows "code - message". | The screen after login. |
| F2 | Wait for Home. | `getAlbumList2` fills the home lists. | Screenshot. |
| F3 | Open the Library tab, Artists. | `getArtists`; an artist opens `getArtist`. | The list. |
| F4 | Open "Double Feature". | `getAlbum`; the track list shows both discs. | Whether the disc number groups or orders the tracks. |
| F5 | Open "Mixtape Volume One". | A track list under Various Artists. | The artist line. |
| F6 | Tap a track. | Playback starts; `scrobble` with `submission=false`. | The player state. |
| F7 | Look at the lists and the player. | Covers load through `getCoverArt`. | Placeholders. |
| F8 | Look for the folders sector in the Library tab. | The sector shows when `getMusicFolders` returns a list, and is hidden when the list is empty. | Whether the sector is present, and its contents. |
| F9 | Open Search, type "Alpha". | `search3` results. | The sections. |
| F10 | Star a track. | `star`; a refusal (status failed) strikes the local decision and reports it. | The star state after a reload. |
| F11 | Create a playlist and add a track. | `createPlaylist`, `updatePlaylist`. | The playlist and its count. |

Unblock condition: `ls /dev/kvm` succeeds after Geoff enables VT-x in the firmware; then install the Android SDK, the
emulator, and a system image under `~/Android/`, and Tempus 4.28.0 (sideloaded from its release page), per [pins]
"Client installs under settled item 9", each with its line in the tier manifest that session; then run this script.

## Script: Amperfy 2.1.1 (rung 16's attended gate)

Amperfy cannot run on Linux. The re-plan (`docs/superpowers/specs/2026-10-05-rungs-12-16-replan-design.md:140-141`: a
client the session cannot drive is walked at rung 16's attended gate) puts its walk at that gate. The script comes
from [clients] 2c and the sub-path precondition in [base-url]; every expectation is unproven.

| Step | Action | Expected observation (unproven) | Record |
|---|---|---|---|
| F1 | On the login screen enter the server URL (an http or https prefix), user name, and password. | The login first sends one GET to the URL itself and rejects a status of 400 or more except 401, then calls `ping`. | Whether login succeeds. |
| F2 | Wait for the initial sync. | `getGenres`, `getArtists`, then `getAlbumList2` in pages of 500; Home sections `getAlbumList2` (newest, recent). | Home and the library counts. |
| F3 | Open Artists. | Artists with album counts (`albumCount` on each artist). | The list. |
| F4 | Open "Double Feature". | `getAlbum`; tracks listed by disc. | The disc ordering. |
| F5 | Open "Mixtape Volume One". | A track list under Various Artists. | The artist line. |
| F6 | Tap a track. | Playback starts; `scrobble` with `submission=false`. | The player state. |
| F7 | Look at the lists and the player. | Covers load. | Placeholders. |
| F8 | Open Directories. | `getMusicFolders`; an empty list shows "No Directories", and a malformed answer raises the "Music Folders Sync" popup. | Which of the two appears. |
| F9 | Search for "Alpha". | `search3` three times with a non-empty query. | The sections. |
| F10 | Star a track. | `star`. | The star state. |
| F11 | Create a playlist and add a track. | `createPlaylist` (name only), then `updatePlaylist`; the new id is read back by listing. | The playlist. |
