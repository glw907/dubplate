# Streamer 2.0 review: digital audio players, local and server

Fresh-context review of
`docs/superpowers/specs/2026-09-07-streamer-2-0-design.md` (revision 1)
through one lens: how dubplate works cleanly with the digital audio
players (DAPs) a family actually owns, at the workstation level (the
future local `dubplate` binary, organizer spec rung 10) and at the
server level (2.0). Researched 2026-09-07; every claim carries a URL.

## Verdict

**DAP sync is a local job, and the spec is right to be silent about
it.** A DAP is a block device or an MTP endpoint on a USB cable
attached to a workstation, not to the VPS. Every capability the sync
needs — a full-quality source library, ffmpeg, a filesystem the tool
can walk and write, knowledge of what the device already holds — is
present locally and absent at the server, which sees the family through
Cloudflare. Nothing in the 2.0 spec should grow a device-sync verb.

**The largest real DAP class is a phone, and it is already served.**
Symfonium, DSub, Ultrasonic and Tempo all cache or download from a
Subsonic server for offline play, and Symfonium's Hi-Res engine drives
USB DACs and Android DAPs natively
([symfonium.app](https://symfonium.app/), retrieved 2026-09-07;
[docs.symfonium.app offline settings](https://docs.symfonium.app/wiki/settings/settings-offline-cache-and-download/),
retrieved 2026-09-07;
[DSub2000](https://github.com/paroj/DSub2000), retrieved 2026-09-07).
Head-Fi practice is already to point UAPP or HiBy Music at DSub's cache
directory
([head-fi.org](https://www.head-fi.org/threads/android-dap-music-apps-playing-from-cache-possible.962071/),
retrieved 2026-09-07). "Phone or Android DAP as the endpoint" is a
client feature, and 2.0's OpenSubsonic surface buys it whole. No
server-side DAP feature is needed for it.

**What 2.0 must add now: nothing new, but one launch endpoint must be
specified rather than assumed.** `download` is already on the launch
list (line 227) and is the one server verb a DAP owner reaches for: a
person on a laptop that is not the library host wants an album's real
files. Navidrome's `download` zips an album, artist or playlist and
puts an `.m3u` inside the playlist zip
([DeepWiki, Navidrome media streaming](https://deepwiki.com/navidrome/navidrome/4.4-media-streaming-and-transcoding),
retrieved 2026-09-07). 2.0 should match that behavior and name it, so
the zip's internal layout is the pipeline's layout, sanitized (finding
1).

**2.1, and only if a real device appears:** a `dap export` verb in the
local binary — a named DAP profile bundling a path template, a filename
sanitizer, a transcode rule, an art writer and a playlist writer.
That is exactly beets' `convert` plus a sanitizer, and it belongs beside
the local organizer's other library verbs, never at the server.

**Never:** MTP, device detection, or a server that knows a device
exists. MTP is a Windows-desktop protocol spoken over a cable; a
Cloudflare-fronted VPS cannot reach it, and building it would be the
second truth about the library that the catalog decision (lines 46-57)
exists to prevent.

The real cost of getting DAPs wrong lands **before** 2.0, in rungs 7 and
8 of the current ladder. A path or filename the pipeline chooses today
becomes a file that cannot be copied to a FAT32 card tomorrow, and a tag
written in a version a player will not read is a re-tag of the whole
library later. Findings 2 through 6 are those forward hooks.

## DAP classes

| Class | Connection | Filesystem | Formats | Album art | Playlists | ReplayGain | Sorting | Tags read |
|---|---|---|---|---|---|---|---|---|
| Rockbox (iPod classic, Sansa, modern ports) | USB mass storage | FAT32 on most targets; 4 GB file cap | MP3, Vorbis, FLAC, WavPack, ALAC, AAC, Musepack, APE, WMA | **External file only** on most builds: `cover.bmp`/`cover.jpg`, `folder.jpg`, same dir, parent dir, or `/.rockbox/albumart/`; no RLE BMP, no progressive JPEG; PictureFlow cannot use embedded | `.m3u8`, UTF-8, relative paths from the playlist's dir | Yes, in the codec's **native** tag only (ID3v2 for MP3, not APEv2) | Tag database; sort tags historically unsupported (FS#7287) | ID3v1, ID3v2, Vorbis comments, APEv2 |
| Sony Walkman NW-A / ZX | MSC or MTP (selectable) | Internal + microSD (exFAT) | MP3, WMA, AAC/HE-AAC, FLAC to 384 kHz, ALAC, AIFF, WAV, APE, DSD to 11.2 MHz | Embedded, plus an image file **named after the album folder** in that folder; baseline/progressive JPEG and PNG | `.m3u` recognized by Music player | Not exposed in Sony's player | Album Artist tag; files must sit under `Music/` or a registered import folder | ID3, Vorbis comments |
| FiiO M-series (Android) | MTP, or MSC where offered; microSD | exFAT/FAT32 microSD | Wide, DSD included; FiiO Music app scans | Embedded; app has a file-size limit setting for art | External `.m3u` with **absolute** device paths (`/storage/emulated/0/…`, `/storage/external_sd/…`), visible only under "folder" | Via HiBy-derived and third-party apps | App media scan over tags | Android media store plus app scan |
| HiBy, Shanling (HiBy OS) | MSC and MTP | FAT, exFAT, NTFS | FLAC, WAV, ALAC, APE, OGG, DSD | Embedded and folder art | `.m3u` | **Yes**, HiBy Music has ReplayGain | Tag scan | Broad |
| Astell&Kern | MTP (MSC on some) | exFAT | WAV, FLAC, WMA, MP3, OGG, APE, AAC, ALAC, AIFF, DFF, DSF | Embedded, or `cover.jpg`/`folder.jpg` in the folder; embedded art unsupported in WAV/AIFF | `.m3u` on current models (older AK100 wanted `.pls`) | Not documented | ID3 tag first | ID3, Vorbis |
| iBasso, Hidizs | MTP/MSC, microSD | exFAT | FLAC, ALAC, AIFF, OGG, MP3, DFF, DSF, DXD, APE, WMA to 32/384 | Embedded; folder browsing available | `.m3u` | Player-dependent | Sort by Title/Album/Artist/Time/**Folder** | Tag scan |
| Budget (Surfans F20, AGPTEK class) | USB mass storage | FAT32/exFAT microSD | FLAC/WAV/APE to 24/192, DSD128, MP3, OGG, WMA | Embedded, small screens | Often folder-only | No | **Folder browsing is the reliable path** | Minimal, ID3v2.3 safest |
| Phone as DAP + USB DAC | Wi-Fi / network (Subsonic), no cable | App-private cache | Whatever the app decodes | Server `getCoverArt` | Server playlists | App-side | Server-side | n/a — the server is the library |

Sources, all retrieved 2026-09-07:
[Rockbox manual ch.4 (browsing, database, ReplayGain)](https://download.rockbox.org/daily/manual/rockbox-ipod6g/rockbox-buildch4.html),
[Rockbox album art appendix C](https://download.rockbox.org/manual/rockbox-ipodvideo/rockbox-buildap3.html),
[Rockbox tagcache manual](https://github.com/mguentner/rockbox/blob/master/manual/rockbox_interface/tagcache.tex),
[Rockbox FS#7287 sort tags](https://www.rockbox.org/tracker/task/7287),
[Sony NW-A306 supported formats](https://helpguide.sony.net/dmp/1301/v1/en/contents/TP1000752534.html),
[Sony NW-A306 transferring with Explorer](https://helpguide.sony.net/dmp/1301/v1/en/contents/TP1000733594.html),
[Sony NW-A306 importing playlists](https://helpguide.sony.net/dmp/1301/v1/en/contents/TP1000733548.html),
[FiiO external m3u guide](https://www.fiio.com/newsinfo/168517.html),
[FiiO Music app release notes (art size limit)](https://fiio.com/newsinfo/754146.html),
[HiBy MassStorage manual](https://guide.hiby.com/en/docs/guides/usb/massstorage),
[Astell&Kern setup guide](https://bloomaudio.com/blogs/articles/how-do-i-get-music-onto-my-astell-kern-dap),
[iBasso DX260 user guide](https://ibasso.com/uploadfiles/download/DX260userguide.pdf),
[Surfans F20](https://www.surfans.net/products/surfans-f20),
[exFAT specification](https://learn.microsoft.com/en-us/windows/win32/fileio/exfat-specification),
[M3U format](https://en.wikipedia.org/wiki/M3U).

Filesystem constants both levels must respect: 255 characters per name
component, the forbidden set `\ / : * ? " < > |` plus U+0000–U+001F, and
a 4 GB per-file ceiling on FAT32 that exFAT does not share
([exFAT specification](https://learn.microsoft.com/en-us/windows/win32/fileio/exfat-specification),
retrieved 2026-09-07; FAT32's cap is 4,294,967,295 bytes).

## Findings, ranked

**1. `download` needs one sentence of specification (spec line 227,
"media (stream, download, cover art)").** As written, `download` could
ship as a single-file verb. The DAP path needs album, artist and
playlist zips, with the pipeline's own relative layout inside and an
`.m3u8` in the playlist zip — the shape Navidrome already has
([DeepWiki](https://deepwiki.com/navidrome/navidrome/4.4-media-streaming-and-transcoding),
2026-09-07). Add: the zip carries sanitized names (finding 2), no path
component over 255 characters, and relative playlist entries. That is
the whole server-side DAP story, and it costs one paragraph.

**2. Filename sanitization and path-length rules belong in rung 7's
layout, not in a later export (spec lines 276-298, forward hooks).**
The hooks list has seven entries and none names the filesystem. A path
the pipeline picks with `:` in an album title, or a
`Collection/Artist/Album (Deluxe Edition, 2019 Remaster)/12 - Track.flac`
chain that runs past 255 characters in one component, is a file that
cannot be written to any DAP card without renaming — and renaming at
export breaks the playlist that names it. Add an eighth hook: **the
layout's filename policy is portable by construction** (the FAT
forbidden set, per-component length cap, no trailing dot or space, no
reserved DOS names), with the unsanitized title kept in the tag. The
existing unidecode table (spec line 166) is already in the tree for the
matcher's folding, so the machinery is present.

**3. Sort tags are a rung 8 decision with no second chance (spec line
312, taglib "rung 8 now").** `ALBUMARTISTSORT`/`ARTISTSORT` and their
ID3 equivalents `TSOP`/`TSOA`/`TSOT` are written once at admit or never
written at all; retrofitting means re-tagging the library. The fallback
chain in practice is albumartistsort → artistsort → albumartist →
artist
([mp3tag community](https://forums.mp3tag.de/index.php?showtopic=10328),
retrieved 2026-09-07). DAP support is genuinely thin — Android's media
store ignores them and Rockbox's request is still open (FS#7287) — but
the cost of writing them is one MusicBrainz field the matcher already
receives, and 2.0's own projector wants a sort key anyway (spec line
144, folded sort key). Recommend: write them at rung 8, and have the
projector's sort key prefer them.

**4. ID3v2.3 versus v2.4 must be recorded as a rung 8 ruling.** The
spec names Picard's tag mapping tables (line 317) but not the version.
ID3v2.4 is still not universally read by portable players, and v2.3 is
the compatibility floor most tagging guidance names
([mp3 tag versions](https://www.maniactools.com/soft/mp3tag-pro/mp3-tags-versions-of-id3.htm),
retrieved 2026-09-07). The library is FLAC-dominant, where the question
does not arise, so this is cheap: rule it, record it, move on. A
defensible answer is v2.4 in the library with a v2.3 downgrade in the
local binary's future export profile — but only if the ruling is
written down now, because "which version did we write" is otherwise a
question answered by reading files in 2027.

**5. ReplayGain: decide where the numbers live, even if nothing
computes them yet.** Rockbox reads ReplayGain only from the codec's
**native** tag (ID3v2 for MP3, not APEv2), and HiBy Music supports it;
Sony's player does not expose it. The organizer spec already demoted
ReplayGain from an in-box plugin to probe-backed
(ROADMAP, built-ins roster). The 2.0 spec is silent, and it should not
be: the transcode profiles (spec lines 179-183) are where a normalizing
filter would live, and the cache key (track hash, profile) already
distinguishes a gain-applied encode from a raw one. Recommend one line
in the transcoding section: gain is a property of the profile, never
applied to the library file, and the tag — when a later pass computes
it — is written in the container's native scheme.

**6. Art at admit should also land as a file beside the album, not only
as a blob (spec lines 151-160, hook 3 at line 286).** The art design is
correct for the streaming path, but the DAP path needs a plain
`cover.jpg` in the album directory: Rockbox on most builds reads
**only** external images, Astell&Kern cannot embed in WAV or AIFF, and
several apps prefer folder art. Writing `cover.jpg` at admit alongside
the content-addressed blob costs one file write and makes every future
export a copy rather than an extraction. Recommend amending hook 3 to
say both.

**7. Gapless: name it as a pipeline property, not a player problem.**
FLAC is gapless by construction; MP3 needs the LAME info tag or
iTunSMPB
([gapless playback](https://en.wikipedia.org/wiki/Gapless_playback),
retrieved 2026-09-07). Any transcode profile 2.0 ships (spec line 182,
"a small shipped set") that produces MP3 must use an encoder path that
writes those headers, or every transcoded album gains a click between
tracks on every DAP. One constraint sentence on the shipped profile set.

**8. Playlist export is a local-binary deliverable, and its path style
is per-device.** There is no portable answer: Rockbox wants `.m3u8`,
UTF-8, relative; FiiO wants `.m3u` with absolute on-device paths under
`/storage/external_sd/`; older Astell&Kern wanted `.pls`. This is
precisely why the profile is the unit — a path template, a sanitizer, a
format rule, an art writer, a playlist writer — and why it cannot live
at the server, which does not know the device's mount point. No change
to 2.0; record it against the organizer spec's rung 10 so the local
binary's design does not treat "write an m3u" as one behavior.

**9. Prior art the study rung should add (spec lines 300-326).** The
table is streaming-shaped. For the eventual export profile the useful
entries are beets' `convert` plugin (transcode, embed art, write an m3u
to the destination, skip re-encoding files already below the bitrate:
[beets convert docs](https://beets.readthedocs.io/en/stable/plugins/convert.html),
retrieved 2026-09-07) and MusicBee's device sync (naming template
versus preserved structure, delete-what-is-no-longer-listed:
[MusicBee wiki, Devices](https://musicbee.fandom.com/wiki/Devices),
retrieved 2026-09-07). Both are behavior references, and `convert` is
the closest thing to the profile shape this repo already understands.

**10. Add a line to the 2.1 list (spec lines 328-333) so the question
has a home.** Suggested wording: "a DAP export profile in the local
binary (path template, sanitizer, transcode rule, art writer, playlist
writer), gated on a real device in the family." That records the
verdict — local, not server — where a later reader will find it, and
prevents a future contributor proposing a server-side sync verb.

## What the spec gets right

The catalog-is-the-index decision (lines 46-57) is what makes an export
profile cheap later: one row per admitted track with a content hash, a
path and a folded sort key is exactly the input a sanitizer and a
playlist writer need, and the determinism test on the projector means a
re-export produces the same names. The content-keyed transcode cache
(lines 68-78) is reusable by a future export at the profile level. The
decision to ship no first-party player (lines 36-44) is also the right
DAP decision: OpenSubsonic is what makes the phone-as-DAP class work,
and that class is larger than every cabled player combined.
