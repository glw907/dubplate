# Client roster and group-feature audit

Date: 2026-10-05. Answers the streamer spec's open item on the client roster's membership
(`specs/2026-09-07-streamer-2-0-design.md`, "Open items") and records how each roster client handles the four
behaviors dubplate's group features depend on. Evidence is code read at pinned commits, plus GitHub and GitLab
project data pulled the same day.

## The roster

Ruled by Geoff, 2026-10-05. The roster favors popular open-source clients, so a defect found in a client can be
fixed with an upstream PR. The group today is two members on Android phones with Linux desktops and two on iOS and
iPadOS.

| Platform | Client | Repository | Why |
|---|---|---|---|
| Linux desktop | Feishin | `jeffvli/feishin` | The most popular open-source Subsonic desktop client |
| Linux desktop | Rufin | `screwys/Rufin` | Geoff's daily client; GTK4 and Rust, created 2026-05-15 |
| iOS and iPadOS | Amperfy | `BLeeEZ/amperfy` | The main open-source iOS client; half the group is on iOS |
| Android | Tempus | `eddyizm/tempus` | Chosen on popularity, user base, and active development |

Navic (`ssalggnikool/Navic`) is the Android alternate once it leaves alpha. A closed client joins the roster only
if a member uses it daily, and then as a compatibility check with no PRs expected. This replaces the spec's
candidate list (Symfonium, Feishin, Supersonic, Substreamer, Tempus), which had two closed clients and no iOS
client.

### Android selection

Data pulled 2026-10-05. Every candidate is GPL-3.0.

| Client | Stars | Latest release | Activity | Verdict |
|---|---|---|---|---|
| Tempus | 1,343 | v4.28.0, 2026-10-02; dev builds to 2026-10-05 | Pushed 2026-10-05; releases every few days | Chosen |
| Navic | 1,114 | v1.0.0-alpha59, 2026-09-21 | Pushed 2026-10-05 | Alternate; still alpha |
| Tempo | 2,246 | 3.9.0, 2024-12-31 | Last push 2025-01-31 | Stalled; Tempus is its maintained successor |
| Subtracks | 792 | 2.0.0-alpha.3, 2023-05-16 | Pushed 2026-10-04 | No release in three years |
| Ultrasonic | 144 (GitLab) | 4.9.0, 2026-03-10 | Active 2026-10-05 | Mature, with slow releases and a small following |
| DSub | 620 | 5.5.3, 2022-04-02 | Last push 2024-06-02 | Dormant |

Stars stand in for user satisfaction only roughly. Play Store and F-Droid reviews were not pulled.

## What the protocol offers

The OpenSubsonic spec (`opensubsonic/open-subsonic-api` at `bed1688`) defines an optional playlist field
`readonly` in `content/en/docs/Responses/playlist.md:53,64` and `playlistWithSongs.md:113,125`, declared in
`openapi/schemas/Playlist.json:54`. Its text reads "If true the playlist cannot be edited by the current user."
The value reflects the authenticated caller's access, an absent field means editable, and clients "should hide or
disable UI actions that modify the playlist." The spec names smart and server-generated playlists as the use case.
The `Playlist` type also carries `allowedUser`, an array of usernames; whether it grants edit rights was not
confirmed. The `notes` field of `albumInfo` (`content/en/docs/Responses/albumInfo.md:32`) is a plain string, and
the spec does not say whether it may hold HTML.

## Client behavior

Four behaviors, read in each client's source. "Collaborator edit" asks whether a member who doesn't own a public
playlist can add, remove, and reorder its tracks, or whether the client blocks the attempt before the server
decides.

| Behavior | Tempus | Amperfy | Feishin | Rufin |
|---|---|---|---|---|
| Collaborator edit | Allowed | Allowed | Blocked for non-admins in the menu | Allowed |
| Reads `readonly` | No | No | No | Yes |
| Shows album notes | Yes, plain text | No | Navidrome servers only | No |
| Music folders | Yes | Yes | Yes | Yes |

Pinned commits: Tempus `687d97f`, Amperfy `c165a7a`, Feishin `9dc1547`, Rufin `2854675`.

**Tempus.** The playlist model carries `owner` (`subsonic/models/Playlist.kt:29`), but no UI code compares it to
the user, and the edit menus are always offered (`ui/fragment/PlaylistPageFragment.java:226-233`). A refused save
shows a toast (`PlaylistEditorFragment.java:104-108`). `AlbumRepository.java:190` calls `getAlbumInfo2`, and
`AlbumPageFragment.java:302-317` shows the notes after `util/MusicUtil.java:257-271` strips HTML. Music folders are
selected through `SettingViewModel.java:24-35` and `Preferences.setActiveMusicFolderId`.

**Amperfy.** The playlist parser reads only id, name, duration, and song count
(`SsPlaylistParserDelegate.swift:50-90`). `PlaylistDetailVC.swift:31,40` sets `isEditAllowed = true`, disabling
edits only for smart playlists, which it detects by an id-prefix check (`:241`, `Playlist.swift:187`). The
add-to-playlist picker lists every playlist `getPlaylists` returns, owned or not
(`FetchedResultsControllers.swift:981-985`). A refused add is reported only to the app's event log
(`PlaylistSelectorVC.swift:206`). `SubsonicServerApi.swift` has no `getAlbumInfo` or `getAlbumInfo2` action. Music
folders come from `getMusicFolders` (`:839`), and `getIndexes` takes `musicFolderId` (`:849-852`). The iPad and
iPhone share one playlist and album-detail path; `userInterfaceIdiom` changes only grid size and insets
(`AlbumsCommonVCInteractions.swift:372-379`).

**Feishin.** The playlist context menu disables Edit and Delete on another member's public playlist unless the
user is an admin (`features/context-menu/menus/playlist-context-menu.tsx:26-41`, `store/auth.store.ts:180-187`).
Track removal and reordering in the detail view carry no owner check. A refused add throws "Failed to add to
playlist" (`api/subsonic/subsonic-controller.ts:303-312`). `getAlbumInfo2` is called only from
`api/navidrome/navidrome-controller.ts:407`, so notes never appear for a plain Subsonic server. Music folders have a
picker (`sidebar/components/server-selector-items.tsx:39,225`).

**Rufin.** `item.rs:573` sets `writable = !readonly`, so an absent field means writable. Add controls require
`writable` (`media_menus.rs:1048`), and the add-to-playlist picker keeps only writable playlists
(`library/src/playlists.rs:489`). Rename and delete require `metadata_writable` (`media_menus.rs:1045,1095`), and
the only owner check (`item.rs:576`) applies to the Navidrome flavor. A member adds a generic server as "subsonic"
(labelled OpenSubsonic), chosen by hand and never auto-detected (`mod.rs:370-400`). On that path, a playlist that
isn't `readonly` shows rename and delete to a non-owner. No source under `crates/` calls `getAlbumInfo`. Music
folders come from `getMusicFolders` (`refresh.rs:359`), and `musicFolderId` is passed (`refresh.rs:324`,
`client.rs:213`).

Supersonic (`dweymouth/supersonic` at `f28980f`) was read in the first round and dropped from the roster. Symfonium
and Substreamer were not read. Symfonium support threads report that only an owner may update a remote playlist
and that read-only playlists import read-only. Nothing was found for Substreamer.

## Server rules

Conductor reading, 2026-10-05, reversible. These follow from the spec's `readonly` text and the audit.

1. Every playlist response carries `readonly`, computed for the caller. It is false for the owner and for a
   collaborator and true for every other member. A smart playlist is always true.
2. Rename, delete, and collaborator management stay owner-only on the server, as the streamer spec's "Group
   features" already says. A client that shows rename to a collaborator, as Rufin does, meets a refusal.
3. Album notes are sent as plain text. The spec leaves HTML undefined, and Tempus strips it anyway.

Rule 1 means collaborative playlists need no OpenSubsonic extension. The spec's existing `readonly` field carries
the permission. Only managing the collaborator list needs dubplate's web app.

## Upstream PR candidates

Each change helps any OpenSubsonic server, which is what makes it acceptable upstream.

| Change | Clients |
|---|---|
| Show `getAlbumInfo2` notes for any Subsonic server | Amperfy, Feishin, Rufin |
| Honor `readonly`, hiding edit controls | Tempus, Amperfy, Feishin |
| Gate edits on `readonly`, falling back to the owner check, in place of the admin-role check | Feishin |
| Restrict rename and delete to the owner on the Subsonic path | Rufin |

Album notes in Amperfy matter most, since half the group is on iOS. Feishin's admin check matters next, since it
is the one client that blocks a collaborator.

## Project health

How likely a PR is to land, as of 2026-10-05.

| Client | Evidence | Reading |
|---|---|---|
| Tempus | Releases every few days | The strongest PR target |
| Amperfy | 1,784 stars; v2.1.1 on 2026-08-13; newest master commit 2026-07-02; one maintainer holds 1,009 commits against the next contributor's 17; outside PRs merged 2026-06-19, 2026-06-30, and 2026-07-02 | Merges outside PRs, slowly |
| Feishin | Not checked | Check the merge rate before the first PR |
| Rufin | 181 stars; v0.16.6 on 2026-09-29; about 1,300 commits since 2026-07-05, 978 by its maintainer; no outside human contributor among its last 100 merged PRs | Acceptance of outside PRs unproven |

## Open

- Feishin's PR merge rate.
- Whether `allowedUser` grants edit rights in the spec, and whether dubplate should fill it with collaborators.
- Each member's current client. A closed client in daily use joins the roster as a compatibility check.
