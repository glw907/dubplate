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
| Tempus 4.28.0 | deferred, no `/dev/kvm` | not run |
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
  **Update, 2026-10-10 (later the same day).** `/dev/kvm` now exists (`emulator -accel-check`: "KVM (version 12) is
  installed and usable"), and the Android command-line tools (build 16111833), platform-tools, emulator 37.2.12, and
  system image `android-36;google_apis;x86_64` are installed under `~/Android/`, with AVD `dubplate-walk`. The newest
  production Tempus is v4.29.0 (released 2026-10-09; `app-tempus-release.apk`, sha256
  `3b2e3691d7fcdbe9903e75c6c9d26e74cc33baa4c9d5594d200fea797c8c9a2c`, matching GitHub's published digest); the walk
  pins 4.28.0, so a rerun decides which to walk. The walk did not run: the emulator's `qemu-system-x86_64-headless`
  dies with SIGSEGV (jump to an invalid address) shortly after "Emulator is performing a full startup", on
  `-gpu swiftshader_indirect` and on `-gpu guest`, three times in all. Unblock: try another emulator or system image
  build; then run the script.
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
