# Slice 2b: the roster walk, run record

Run on 2026-10-10, 08:35 to 08:55 UTC, on the workstation. The method, the flow list (F1 to F11), and the scripts are
`docs/roster-walk.md`; this record gives each step's outcome. Captures are under
`docs/superpowers/research/2026-10-10-slice-2b-roster-walk/`: `shots/` holds screenshots, `dom/` holds Feishin's DOM
text and one audio-element capture. A capture name below is a file in one of those two directories.

## Verdict

| Client | Against Navidrome | Folder view against gonic |
|---|---|---|
| Feishin 1.17.0 | F1 to F11: 11 of 11 pass | pass (10 folders, one drilled) |
| Rufin 0.16.6 | F1 to F11: 11 of 11 pass | pass (10 folders) |
| Tempus 4.29.0 | F1 to F11: 11 of 11 pass (F11 by a different call shape) | pass (10 folders, one drilled) |
| Amperfy 2.1.1 | rung 16's attended gate | not run |

No step failed on an oracle. The oracle behaviors the walk observed are in "Oracle observations". No hash changed.

## What ran

- **Oracles.** `harness oracle-up` (this worktree's build) on the seed manifest: Navidrome 0.64.2
  (`docker.io/deluan/navidrome@sha256:38dc2727bfcfd5ede290f8ada114fc90368146f265ae4701ddddbcbe2a44ee52`) and gonic
  v0.22.0 (`docker.io/sentriz/gonic@sha256:516fd9645614ba3a596d86174216c3e944808b9ec970c581678713be4c8b1d49`),
  loopback only. The command printed both URLs and the run's credentials, which were typed into the clients by a script
  that read them from the command's output file and never printed them. The library holds 12 albums and 10 artists on
  Navidrome (read with `getAlbumList2` and `getArtists`).
- **Clients.** `org.jeffvli.feishin` 1.17.0 and `io.github.screwys.Rufin` 0.16.6, system Flatpaks, each under an
  isolated configuration home with the keyring talk names denied (method in `docs/roster-walk.md`).
- **Display.** One throwaway Xvfb container, 1440x900x24, container name `dubplate-harness-xvfb`, label
  `dubplate-harness-run=roster-walk-xvfb`. No Xvfb image was available and the walk opened no network connection, so the
  image was built locally: `podman commit` of the workstation's existing `term-headless` distrobox container (Fedora 41
  base, `registry.fedoraproject.org/fedora:41`, digest
  `sha256:3fb969dd07e631c17d12f5955bcb6db2078a6997eb68be76b6492e7bd6052e88`, with Xvfb, ImageMagick, and xdotool added to
  its layer) as local image `dubplate-xvfb-scratch`, image id
  `cf279c9568948efa75927626f6b86a8214cef573b8d6b40389eeeea8e3f7df19`, digest
  `sha256:2f275d4fd35c695ba7b017d7e4ad47728bf417bc1ec4d94ce2eb904f0208175c`. The image was removed after the run
  (`podman rmi` printed `Deleted: cf279c95...`).
- **Driving.** Rufin: `xdotool` clicks and typing on the private display, screenshots by `import` inside the container.
  Feishin: Playwright 1.63.0 (`playwright-core`, the repo's pin) over Feishin's remote debugging port (port chosen by
  Chromium, loopback), screenshots and DOM text from the page.
- **Audio.** A null sink (`dubplate_walk`) for both clients, unloaded afterwards (`pactl list short sinks` shows none).
- Feishin was started twice. The first start's window closed on a scripted click against a toast, so the isolated
  directory was deleted and Feishin restarted from a blank configuration; every Feishin claim below is from the second
  start.

## Hashes, before and after

Method as in `docs/roster-walk.md`: SHA-256 of every file, with directories and symlinks listed, under each real
`~/.var/app/<id>/` tree except `cache/` and `.ld.so/`, and under `~/.local/share/keyrings/`. The listings (one line per
entry) were taken before the first launch and after the last client exited, and compared with `diff`.

| Tree | Entries | Files | SHA-256 of the listing, before | After |
|---|---|---|---|---|
| `~/.var/app/io.github.screwys.Rufin/` | 27 | 9 | `e132fb41992947ca...` | `e132fb41992947ca...` |
| `~/.var/app/org.jeffvli.feishin/` | 6 | 0 | `254bd9bb743a3b08...` | `254bd9bb743a3b08...` |
| `~/.local/share/keyrings/` (`login.keyring`, `user.keystore`) | 2 | 2 | `1a0cafb5eb0b095e...` | `1a0cafb5eb0b095e...` |
| the three listings concatenated | 35 | 11 | `e8c38bfbaf21f858...` | `e8c38bfbaf21f858...` |

(Listing hashes are truncated to 16 hex digits; the concatenation's full SHA-256 is
`e8c38bfbaf21f858ab3a9d829cc4686e909219cd527ea77842485f317853806d` both times.) `diff` of the before and after listings
printed nothing. The excluded `.ld.so/` trees were also looked at: Rufin's `active` link still pointed at
`7eb3feb9...` and Feishin's at `fc1ef965...`, with modification times of 2026-10-06, so no launch regenerated them.
Feishin's real tree holds directories only (no file), as slice 1 left it. Nothing changed, so the isolation held.

## Rufin 0.16.6 against Navidrome

The window and the display are the same size, so every Rufin capture is the whole client window.

| Step | Outcome | Observation | Capture |
|---|---|---|---|
| F1 add | pass | The Add Source form for the OpenSubsonic type accepted name, URL, user name, and password and closed; the source `navidrome-oracle` is the active source. Rufin's log held one warning, "source authentication failed: Wrong username or password", at the first submit: Return in the user name field submitted the form before the password was typed. The credentials were typed in full and the source connected. | `rufin-01-addform.png` (empty form), `rufin-18-sources.png` (source list) |
| F2 login | pass | Home shows the Showcase album "Two Sides" (Carmine Ensemble, 1985, 4 tracks), "Explore", and "Newly added". | `rufin-03-home.png` |
| F3 artists | pass | The Artists grid shows 10 tiles: [Unknown Artist], Alpha Artist, Borealis Trio, Carmine Ensemble, Delta Duo, Guest Singer, The Meridians, Zoë Ångström, Ансамбль Рассвет, 夜明けの音楽. The first visit showed a spinner for over 15 seconds; the page filled on the next visit, after the Albums page had loaded. | `rufin-04-artists.png` |
| F4 multi-disc | pass | "Double Feature", Borealis Trio, 1998, 4 tracks, 8s; the list is grouped under "Disc 1" (Overture, Interlude) and "Disc 2" (Reprise, Finale), each numbered from 1. A single click on the album tile showed hover controls only; a double-click opened it. | `rufin-05-albums.png`, `rufin-06-multidisc.png` |
| F5 compilation | pass | "Mixtape Volume One", Various Artists, 1999, 3 tracks, 6s, labelled "COLLECTION"; rows Opening Remix, Hello Again, Snow Song. | `rufin-08-compilation.png` |
| F6 play | pass | After a double-click on Overture the player bar shows a pause control and 0:01 of 0:02 on "Reprise" (the tracks are two seconds, so the queue had advanced); the queue panel lists all four tracks. Navidrome logged 10 `Streaming file` lines in the whole run, covering the four Double Feature tracks, from both clients; Rufin's client string is `GStreamer souphttpsrc`. | `rufin-07-play.png` |
| F7 cover art | pass | Covers render on the album grid, the album header, and the player bar for the albums that carry art; the five albums without embedded art ([Unknown Album], Café Années, Long Way Home, Утренние песни, 朝のうた) show Rufin's placeholder. | `rufin-05-albums.png`, `rufin-06-multidisc.png`, `rufin-07-play.png` |
| F8 folders (Navidrome) | pass | The Folders page lists 10 folders, one per artist: [Unknown Artist], Alpha Artist, Borealis Trio, Carmine Ensemble, Delta Duo, The Meridians, Various Artists, Zoë Ångström, Ансамбль Рассвет, 夜明けの音楽. No error. The source menu lists the library entries "All Music" and "Music Library". | `rufin-09-folders-nd.png`, `rufin-18-sources.png` |
| F9 search | pass | "Alpha" in the top search box opens a dropdown whose Tracks section lists Dawn, Dusk, Middle (all Alpha Artist); the lower part of the dropdown paints black (a drawing fault of this software-rendered display, not a server answer). "View all" lists 8 tracks (Dawn, Dusk, Middle, Opening, Opening, Opening Remix, Together, and Together by Alpha Artist feat. Guest Singer) with Tracks, Albums, and Artists tabs. | `rufin-10-search.png`, `rufin-10b-search-all.png` |
| F10 star | pass | The heart on Dawn fills and Favorites lists Dawn. Reading the oracle: `getStarred2` returned one song, "Dawn". | `rufin-12-starred.png` |
| F11 playlist | pass | The new-playlist dialog took the name "Walk list"; the Playlists page shows it with 0 tracks. After "Add to Playlist" on Dawn it shows 1 track. Reading the oracle: `getPlaylists` returned "Walk list" with `songCount` 0 after the create and 1 after the add. | `rufin-15-plcreated.png`, `rufin-17-plfilled.png` |

## Rufin 0.16.6, folder view against gonic

| Step | Outcome | Observation | Capture |
|---|---|---|---|
| F1 add (gonic) | pass | A second source `gonic-oracle` was added the same way (type OpenSubsonic); Rufin's log gained no authentication warning. The source menu lists "gonic-oracle" first and the library entries "All Music" and "music". | `rufin-19b-gonic-source.png`, `rufin-19-gonic-added.png` |
| F8 folders (gonic) | pass | The Folders page lists 10 folders: Alpha Artist, Borealis Trio, Carmine Ensemble, Delta Duo, Loose, The Meridians, Various Artists, Zoë Ångström, Ансамбль Рассвет, 夜明けの音楽. gonic names the folder for the tagless track "Loose"; Navidrome's list names it "[Unknown Artist]". | `rufin-20-folders-gonic.png` |

## Feishin 1.17.0 against Navidrome

The captures are Playwright page screenshots of the Electron window (1439x899) and `innerText` dumps.

| Step | Outcome | Observation | Capture |
|---|---|---|---|
| F1 add | pass | After the release-notes dialog was dismissed, the "Server required" form took type OpenSubsonic, name, URL, user name, and password; the submit left `#/action-required` for `#/`. Navidrome registered one player for client `Feishin`, type `feishin`. The filled form is not committed because it shows the user name. | `feishin-01-home.png`, `dom/feishin-01-home.txt` |
| F2 login | pass | Home shows the hero "Double Feature" (Borealis Trio, 1998), "Explore from your library", the sidebar entries, and under Playlists the "Walk list" made in Rufin (1 track, 2s), so both clients read the same oracle state. | `feishin-01-home.png` |
| F3 artists | pass | Artists shows a count badge of 10 and the same ten names as Rufin's list. | `feishin-02-artists.png`, `dom/feishin-02-artists.txt` |
| F4 multi-disc | pass | "Double Feature": 1998, 4 tracks, 8s; the list is grouped under "Disc 1" (1 Overture, 2 Interlude) and "Disc 2" (1 Reprise, 2 Finale). | `feishin-04-multidisc.png`, `dom/feishin-04-multidisc.txt` |
| F5 compilation | pass | "Mixtape Volume One": 1999, 3 tracks, 6s, artist line "Various Artists", a flat list (1 Opening Remix, 2 Hello Again, 3 Snow Song). The Albums page opens with its filter panel showing (`feishin-03-albums.png`, 12 albums); a first scripted click on the album's name text did not navigate, and the album opened from its link. That first attempt's capture is not committed. | `feishin-06-compilation.png`, `dom/feishin-06-compilation.txt` |
| F6 play | pass | After a double-click on Overture the player bar shows a pause control and the title bar reads "(1 / 4) Overture". The first audio element held `currentSrc` `http://127.0.0.1:<port>/rest/stream.view?id=1jmvrDratUDijhMfUDTfa9&v=1.13.0&c=Feishin` (credentials removed), `currentTime` 0.898, `duration` 2, `paused` false, `error` null; a second element held the next track, `paused` true. | `feishin-05-play.png`, `dom/feishin-05-play-audio.json` |
| F7 cover art | pass | The album headers (`feishin-04`, `feishin-06`) and the grids render covers. The player bar's cover for Overture is a different image from the album header's: reading the oracle, the song's own cover id (`mf-...`) returns an image of SHA-256 prefix `9ca8d3a620b3` and the album's (`al-...`) returns `909dd40bd067` at the same size, so the client showed what the oracle served for each id. Albums with no art show the placeholder. | `feishin-03-albums.png`, `feishin-04-multidisc.png`, `feishin-05-play.png` |
| F8 folders (Navidrome) | pass | The Folders page shows a count badge of 10 and the same ten folders as Rufin's. The server menu's music-folder picker lists "None" and "Music Library". | `feishin-07-folders-nd.png`, `dom/feishin-07-folders-nd.txt`, `feishin-07b-folder-picker-nd.png` |
| F9 search | pass | The sidebar Search opens a dialog; typing "Alpha" shows Albums ("4+ results": Greatest Hits, Second Light, Single Disc, Mixtape Volume One), Album Artists (1 result: Alpha Artist, 3 albums), and Tracks ("4+ results": Dawn, Dusk, Middle, Opening). | `feishin-08-search.png`, `dom/feishin-08-search.txt` |
| F10 star | pass | The heart on "Hello Again" fills. Reading the oracle: `getStarred2` returned "Hello Again" and "Dawn" (Dawn from Rufin's step). | `feishin-09-star.png` |
| F11 playlist | pass | The "Create Playlist" dialog took "Feishin list"; a "Playlist created successfully" toast appeared and the sidebar lists "Feishin list" (0 tracks) beside "Walk list". Reading the oracle: `getPlaylists` returned both. | `feishin-10-playlist.png`, `dom/feishin-10-playlist.txt` |

## Feishin 1.17.0, folder view against gonic

| Step | Outcome | Observation | Capture |
|---|---|---|---|
| F1 add (gonic) | pass | "Manage servers", "Add Server", type OpenSubsonic added the entry `gonic-oracle` and made it the active server; the sidebar's Playlists list emptied (gonic holds none). | `feishin-11a-gonic-switch-stale-route.png` |
| F8 folders (gonic) | pass | After the page was reloaded, Folders shows a count badge of 10 and the rows Alpha Artist, Borealis Trio, Carmine Ensemble, Delta Duo, Loose, The Meridians, Various Artists, Zoë Ångström, Ансамбль Рассвет, 夜明けの音楽. A double-click on "Alpha Artist" opens a page with the breadcrumb "Home > Alpha Artist" and 3 rows: Greatest Hits, Second Light, Single Disc. | `feishin-11-folders-gonic.png`, `dom/feishin-11-folders-gonic.txt`, `feishin-12-folder-drill-gonic.png`, `dom/feishin-12-folder-drill-gonic.txt` |

`feishin-11a-gonic-switch-stale-route.png` is a driving artifact and not a result: the server was switched while a
Navidrome album route was open, so Feishin asked gonic for an album by a Navidrome id and toasted gonic's answer
("please provide an `id` parameter", error 10) on a spinning page. The Folders walk was repeated from a reload and is
the `feishin-11` capture.

## Oracle observations

None of these is a failed step. Each is evidence about an oracle.

- **Folder naming.** Navidrome lists the tagless-artist folder as "[Unknown Artist]" and has no "Loose" folder; gonic
  lists "Loose" and no "[Unknown Artist]". Both list 10 artist-level folders on this fixture, and both clients showed
  all ten without error.
- **Folder lists.** Rufin offered the library entries "All Music" and "Music Library" on Navidrome and "All Music" and
  "music" on gonic; Feishin's picker offered "None" and "Music Library" on Navidrome.
- **gonic `getMusicDirectory`.** A request without `id`, and one with `id=0`, both answer `status` "failed", error code
  10, "please provide an `id` parameter", read with `curl` against the oracle. The clients' folder pages did not send
  these requests; their root listing came from `getIndexes`, which gonic answers with the artist list.
- **Per-track cover art on Navidrome.** The song-level cover id returns a different image from the album's for this
  fixture (F7, Feishin).
- **Cross-client state.** A playlist and a star created in Rufin appeared in Feishin on the same Navidrome login.

## Deferrals

- **Tempus 4.28.0.** Script written from [clients] 2d in `docs/roster-walk.md`, marked unproven. Its walk is deferred
  because `/dev/kvm` does not exist: `ls /dev/kvm` on 2026-10-10 printed "No such file or directory". Unblock: `ls
  /dev/kvm` succeeds after Geoff enables VT-x in the firmware; then the Android SDK, the emulator, and a system image
  install under `~/Android/`, and Tempus 4.28.0 sideloads, per [pins] "Client installs under settled item 9" with their
  manifest lines; then the script runs.
  **Update, 2026-10-10 (later the same day; superseded by the next update and the "Tempus 4.29.0" sections below).** `/dev/kvm` now exists (`emulator -accel-check`: "KVM (version 12) is
  installed and usable"), and the Android command-line tools (build 16111833), platform-tools, emulator 37.2.12, and
  system image `android-36;google_apis;x86_64` are installed under `~/Android/`, with AVD `dubplate-walk`. The newest
  production Tempus is v4.29.0 (released 2026-10-09; `app-tempus-release.apk`, sha256
  `3b2e3691d7fcdbe9903e75c6c9d26e74cc33baa4c9d5594d200fea797c8c9a2c`, matching GitHub's published digest); the walk
  pins 4.28.0, so a rerun decides which to walk. The walk did not run: the emulator's `qemu-system-x86_64-headless`
  dies with SIGSEGV (jump to an invalid address) shortly after "Emulator is performing a full startup", on
  `-gpu swiftshader_indirect` and on `-gpu guest`, three times in all. Unblock: try another emulator or system image
  build; then run the script.
  **Update, 2026-10-10 (afternoon): walked.** The emulator boots with `-gpu host` and the walk ran; results are in "Tempus 4.29.0 against Navidrome" below. The crash, found by trial: the host `qemu-system-x86_64-headless` segfaulted only with KVM on (the same AVD under `-accel off` ran past the crash point), and only with a software GPU path (`-gpu swiftshader_indirect`, `-gpu guest`, and `-gpu off` all failed: swiftshader and `off` killed qemu, and under `guest` and `off` the guest's surfaceflinger aborted in a restart loop). `eu-stack` on the systemd core showed the faulting thread jumping into an address inside the heap, with the vCPU threads idle in `KVM_RUN`; no debugger is installed, so the frame is unnamed. Neither an `android-34;google_apis;x86_64` image (in new AVD `dubplate-walk34`, installed under `~/Android/`), `-feature -Vulkan`, `-wipe-data`, CET switched off through `GLIBC_TUNABLES`, `-cores 1`, nor `-qemu -cpu host` changed the crash; the emulator build 37.2.12 is the only one the SDK offers on every channel, so no stable-channel swap existed. The fix was `-gpu host` (host Mesa through the emulator's own GL path, still headless under `-no-window`): "Boot completed in 73590 ms" on the first try. The walk used `dubplate-walk34` (API 34) because that is where the passing boot was found; `dubplate-walk` (API 36) was not retried under `-gpu host`.

- **Amperfy 2.1.1.** Listed for rung 16's attended gate (`docs/superpowers/specs/2026-10-05-rungs-12-16-replan-design.md:140-141`:
  "A client the conductor cannot drive is walked at rung 16's attended gate."). Its script, from [clients] 2c and the
  sub-path precondition in [base-url], is in `docs/roster-walk.md`, unproven.

## Teardown

- Both clients: `flatpak kill`; `flatpak ps` showed no Feishin or Rufin instance afterwards.
- Null sink unloaded; Xvfb container removed; scratch image removed; `oracle-up` interrupted (it removes its two
  containers).
- `podman ps -a --filter label=dubplate-harness-run` printed no container, and `podman ps -a` shows only the
  workstation's own `term-headless`.
- `ss -ltnp` filtered for this run's processes printed nothing. `/tmp/.X11-unix` holds only the desktop's `X0` and `X1`.

## Tempus 4.29.0 against Navidrome

Run on 2026-10-10 from about 10:40 to 10:50 AKDT. Version: Tempus v4.29.0 (`com.eddyizm.tempus`),
`app-tempus-release.apk`, SHA-256 `3b2e3691d7fcdbe9903e75c6c9d26e74cc33baa4c9d5594d200fea797c8c9a2c` (matches GitHub's
published digest, rechecked with `sha256sum` before the install). The record and `docs/roster-walk.md` pin 4.28.0; 4.29.0 was
walked because it is the newest production release (published 2026-10-09) and the project walks the newest production version
of every client, so the pin is a stale reading to update in `docs/roster-walk.md`, which this record does not edit.

**Stack.** Android emulator 37.2.12, AVD `dubplate-walk34` (Pixel 6 profile, `android-34;google_apis;x86_64`, 1080x2400,
2 GiB RAM, booted `-no-window -no-audio -no-boot-anim -no-snapshot -gpu host`). `harness oracle-up` from the repo's committed
fixture: Navidrome 0.64.2 and gonic v0.22.0 at the pinned digests, loopback only. The emulator reaches the host at
`10.0.2.2`, but the oracles listen on loopback ports chosen at start, so a throwaway logging proxy (`127.0.0.1:18001` to
Navidrome, `18002` to gonic, outside the repo) forwarded every request and recorded method, path, parameter names (never
values), `f`, client string, and status. That proxy is why the request shapes below are known. Server entries were typed
through the UI with the "Low security" box checked, which Tempus requires for an `http://` address; the credentials came from
the `oracle-up` output file and were typed by `adb input text`, never written to the record.

| Step | Outcome | Observation | Capture |
|---|---|---|---|
| F1 add | pass | The Add server dialog takes Server Name, Username, Password, Server URL, an optional Local URL, a Low security checkbox, and an optional client certificate. After Save the server is listed; tapping it logs in. `ping` returned 200 as the first call. | `t03` (dialog), `t05` (list) |
| F2 login | pass | Home shows "Discovery" (Side B Two, Two Sides), "Flashback" decades, and "Recently added" ([Unknown Album], Long Way Home). Calls: `getAlbumList2` (several types, one with `fromYear` and `toYear`), `getRandomSongs`, `getStarred2`, `getPlaylists`, `getBookmarks`. | `t06` |
| F3 artists | pass | Library, Browse Artists lists ten: [Unknown Artist], Alpha Artist, Borealis Trio, Carmine Ensemble, Delta Duo, The Meridians, Various Artists, Zoë Ångström, Ансамбль Рассвет, 夜明けの音楽 (the oracle's `getArtists` names the same ten). Opening Borealis Trio showed its album "Double Feature" and Top Songs; the earlier visit said "No top songs", a later one listed Snow Song after `getTopSongs`. | `t08-artists`, `t08c-artists`, `t09-artist` |
| F4 multi-disc | pass | "Double Feature", Borealis Trio, 1998; the list is grouped under "Disc 1" (1 Overture, 2 Interlude) and "Disc 2" (1 Reprise, 2 Finale), each 0:02. Covers render. | `t10-multidisc` |
| F5 compilation | pass | "Mixtape Volume One", Various Artists, 1999: Opening Remix (Alpha Artist), Hello Again (Delta Duo), Snow Song (Borealis Trio). The album artist line says Various Artists; each row shows its own track artist. Reached through Search, since the Library grid lists only a few albums per row. | `t19-compilation` |
| F6 play | pass | Tap on Overture: the player bar shows pause and the title advanced to "Reprise" within seconds (two-second tracks). The server saw five `stream` requests (client string `Dalvik/2.1.0 (Linux; U; Android 14; ...)`, ExoPlayer's default; parameters `id`, `format`, `maxBitRate`) and `scrobble` calls with `id` and `submission`. The queue ended on "Snow Song", a track the walk did not pick: Tempus asked for `getSimilarSongs2` and appended the result. The player sheet shows "flac, 47kbps, 16b, 16kHz". | `t11-play`, `t15-player` |
| F7 cover art | pass | Covers render in the album header, the player bar, the player sheet, and the search grid. Albums without art show Tempus's record placeholder. `getCoverArt` came in with and without `size`, from both `okhttp` (the image loader) and the media stack. | `t10`, `t15-player`, `t18-search` |
| F8 folders (Navidrome) | pass | Library shows a "Music folders" sector (Music Library) because `getMusicFolders` returned a list. Opening it calls `getIndexes` with `musicFolderId` and lists ten folders: Ансамбль Рассвет, 夜明けの音楽, Alpha Artist, Borealis Trio, Carmine Ensemble, Delta Duo, The Meridians, Various Artists, Zoë Ångström, [Unknown Artist]. Opening Alpha Artist calls `getMusicDirectory` and lists Greatest Hits, Second Light, Single Disc. | `t20-folder-nd`, `t21-folder-drill-nd` |
| F9 search | pass | "Alpha" gives Artists (Alpha Artist), Albums (Greatest Hits, Second Light), and Songs ("Play all 8 songs", 8 tracks, 0:16). Tempus issues `search3` on every typed character (42 calls over the walk). "Mixtape" gives the album and "Play all 3 songs". | `t18-search` |
| F10 star | pass | The heart on the player sheet fills for "Snow Song" (the current track). The oracle's `getStarred2` returned one song, "Snow Song". | `t16-starred` |
| F11 playlist | pass, different call | Add to playlist on Interlude, "Create", name "Tempus-list", Save. Tempus sent one `createPlaylist` with `name` and `songId` together and no `updatePlaylist`, which the script predicted (`createPlaylist`, `updatePlaylist`). The oracle's `getPlaylists` returned "Tempus-list" with `songCount` 1. | `t13-addpl` |

## Tempus 4.29.0, folder view against gonic

| Step | Outcome | Observation | Capture |
|---|---|---|---|
| F1 add (gonic) | pass | Settings, General, "Log out" returns to the server list (there is no server switcher on Home); a second entry `gonic-oracle` was added and opened. `ping`, `getMusicFolders`, `getAlbumList2`, `getRandomSongs`, `getStarred2`, `getPlaylists` all answered 200; `getPodcasts` and `getNewestPodcasts` answered 200 on gonic and 501 on Navidrome, and Tempus showed no error in either case. | `t23-gonic-home` |
| F8 folders (gonic) | pass | Library shows the sector "music"; opening it lists ten folders: Alpha Artist, Ансамбль Рассвет, Borealis Trio, Carmine Ensemble, Delta Duo, Loose, The Meridians, Various Artists, Zoë Ångström, 夜明けの音楽. Alpha Artist opens to Greatest Hits, Second Light, Single Disc. "Loose" replaces Navidrome's "[Unknown Artist]", as in the Feishin and Rufin walks. | `t24-folders-gonic`, `t25-drill-gonic` |

## Request shapes only Tempus sent

Read from the proxy log (parameter names only, `u`, `p`, `t`, `s`, `v`, `c`, `f` excluded; counts are the whole walk):

- `getRandomSongs(size)`, `getBookmarks`, `getInternetRadioStations`, `getPodcasts(includeEpisodes)`, `getNewestPodcasts(count)`: the Home and Radio and Podcast tabs, called at login with no user action. Navidrome answered 501 to the two podcast calls; gonic answered 200.
- `getGenres`, `getAlbumInfo2(id)`, `getArtistInfo2(id)`, `getTopSongs(artist, count)`, `getSimilarSongs2(id, count)`, `getLyrics(artist, title)` (12 calls), `getSong(id)` (12 calls): called on opening an artist, an album, or a track, with no matching flow in the script.
- `getAlbumList2` with `fromYear` and `toYear` and `type` (the Flashback decades), and with `type`, `size`, `offset`.
- `getIndexes(musicFolderId)` then `getMusicDirectory(id)` for the folder view, and `getMusicFolders` at login.
- `savePlayQueue(id, current, position)` (2 calls) and `deleteBookmark(id)` (10 calls, one for each track that reached its end).
- `scrobble(id, submission)` 23 times, including both `submission=false` and the submitted form per track.
- `createPlaylist(name, songId)` as one call that creates and fills.
- `getCoverArt(id)` both with and without `size`.
- `search3` with all seven parameters (`query`, `artistCount`, `artistOffset`, `albumCount`, `albumOffset`, `songCount`, `songOffset`), once per typed character.
- `getArtist(id)` 26 times and `getAlbum(id)` 18 times: Tempus prefetches every artist and album it lists.
- Client strings: `okhttp/5.0.0-alpha.14` for every API call and every image; `Dalvik/2.1.0 (Linux; U; Android 14; sdk_gphone64_x86_64 Build/UE1A.230829.050)` for `stream` and a share of `getCoverArt`. No call carried an `Authorization` header or a body; `f=json` on every envelope call.

## Tempus 4.29.0, defects and oddities

None failed against an oracle. Findings, each about Tempus or the setup:

- **Cleartext needs "Low security".** An `http://` server address is accepted only with the Low security checkbox ticked; a walk against dubplate-server on plain HTTP over loopback needs the same step.
- **Prefetch volume.** `getArtist` and `getAlbum` are called for every listed item, so a library of N albums costs on the order of N requests at first open; a walk against dubplate-server should read the request log for that volume.
- **Queue extension.** After the album's last track Tempus appends `getSimilarSongs2` results, so a "played" check must read the track it started on, not the one the player shows later.
- **Top Songs.** An artist page said "No top songs available" on the first visit and listed a song after a later `getTopSongs`; it is Tempus caching or Navidrome answering differently, and was not chased.
- **F11 prediction was wrong.** The unproven script expected `updatePlaylist`; Tempus uses `createPlaylist` with `songId`. `docs/roster-walk.md` carries the old expectation and should be corrected at the next edit of that file.
- **Not walked.** Tempus's Download tab, remote play, Android Auto, and transcoding settings were out of the flow list.

Captures (`t01` to `t25` PNG screenshots of the emulator, plus the proxy log) sit in the session's scratchpad and are not committed:
they were taken on an emulator whose dialog fields briefly showed the run's throwaway oracle credentials (a `uiautomator` text dump
printed the password once in a terminal; no screenshot shows it, but they were not cropped or reviewed for commit). Nothing was written to
the production box, and no file outside `~/Android` and this record changed.

## Tempus teardown

- Emulator stopped with `adb emu kill`; no `qemu-system` process remains. AVD `dubplate-walk34` and the image `android-34;google_apis;x86_64` stay under `~/Android/` (their tier-manifest lines are owed: the install session recorded neither the API 34 image nor the AVD).
- `oracle-up` interrupted (it removes its two containers); `podman ps -a` shows only `term-headless`.
- Both logging proxies killed; nothing listens on 18001 or 18002.
