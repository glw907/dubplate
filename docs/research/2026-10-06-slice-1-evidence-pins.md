# Slice 1, Task 0: evidence pins and probe oracles

Date read: 2026-10-07 (UTC, 04:40 to 04:50). Staged outside the repo; lands at A-land as
`docs/research/2026-10-06-slice-1-evidence-pins.md`.

**Questions this record owns.** (1) Which release and commit is each study subject pinned at. (2) Do the two
oracles run locally from a recorded recipe and answer `ping` and a library listing, at the root and under a
sub-path. (3) Does each oracle document a base-path option. (4) What does settled item 9 leave installed or not.

## Pins

Settled item 5 (first pin wins). Slice 2's plan (`plans/2026-10-06-rung-12-slice-2-oracle-harness.md`, "Task 0")
resolved both oracle images on 2026-10-06. Its worktree is not cut (`git worktree list` shows only the main tree;
`git branch -a` shows only `master`), so its plan is the only pin source. Both pins are adopted. The full digests
below were read from the registries on 2026-10-07 with `skopeo inspect` and match slice 2's truncated forms.

| Subject | Pin | Source (read 2026-10-07) | Newest production? |
|---|---|---|---|
| Navidrome | `docker.io/deluan/navidrome:0.64.2` = `sha256:38dc2727bfcfd5ede290f8ada114fc90368146f265ae4701ddddbcbe2a44ee52` (the tag carries no `v`; `ghcr.io/navidrome/navidrome:0.64.2` resolves to the same digest). Image created 2026-09-24T23:09:43Z; OCI revision label `101145742f4762164202cb4858c9126258bdc463`. | `skopeo inspect docker://docker.io/deluan/navidrome:0.64.2`; `https://github.com/navidrome/navidrome/releases/tag/v0.64.2`, published 2026-09-24T23:26:20Z, `prerelease: false` | Yes. `gh api repos/navidrome/navidrome/releases` lists v0.64.2 first; the next is v0.64.1 (2026-09-21). No newer release or pre-release. |
| gonic | `docker.io/sentriz/gonic:v0.22.0` = `sha256:516fd9645614ba3a596d86174216c3e944808b9ec970c581678713be4c8b1d49` (the tag carries the `v`). Image created 2026-06-08T19:49:52Z. Tag `v0.22.0` = commit `5ae99e8af550e4d5cbd45a3b430419d32cb54aed` (2026-06-08). | `skopeo inspect docker://docker.io/sentriz/gonic:v0.22.0`; `https://github.com/sentriz/gonic/releases/tag/v0.22.0`, published 2026-06-08T19:47:18Z, `prerelease: false` | Yes. v0.22.0 is first in the release list; master was pushed 2026-10-06 but no release followed. |
| OpenSubsonic specification | `opensubsonic/open-subsonic-api` at `bed1688a7b0afe9f528a3dc1400a46ab7961b280` (2026-08-07, branch `main` HEAD on 2026-10-07). No releases or tags. Extension count at this commit: **10** (`content/en/docs/Extensions/`: apiKeyAuth, formPost, getPodcastEpisode, indexBasedQueue, playbackReport, songLyrics, sonicSimilarity, topSongsByArtistId, TranscodeOffset, transcoding; `template.md` and `_index.md` excluded). | `https://github.com/opensubsonic/open-subsonic-api/commit/bed1688a7b0afe9f528a3dc1400a46ab7961b280`; shallow clone listed under `Extensions/` | Yes: no commit since 2026-08-07. Same commit as the 2026-10-05 roster audit's HEAD (`bed1688`). |
| Navidrome documentation | `navidrome/website` at `b0fe572513288225562fc4a71d06e01333f03602` (2026-10-02 committer date; API reports `pushed_at` 2026-10-04) | `https://github.com/navidrome/website/commit/b0fe572513288225562fc4a71d06e01333f03602` | n/a (a documentation repository, `master` HEAD) |
| Feishin | release `v1.17.0` (2026-09-15) = commit `5c169aca8d550032c9aa6e8e2b9f624d03d42fec`. The roster audit's pin `9dc1547` was a `development` commit; the study reads the release commit. | `https://github.com/jeffvli/feishin/releases/tag/v1.17.0`, `prerelease: false` | Yes (v1.16.0 2026-09-14 is next) |
| Rufin | release `v0.16.6` (2026-09-29) = commit `aad2f44e5171817c96d65ff3f75522ef7aab19c4`. The audit's pin `2854675` was a `main` commit. | `https://github.com/screwys/Rufin/releases/tag/v0.16.6`, `prerelease: false` | Yes |
| Amperfy | release `v2.1.1` (2026-08-13) = commit `c165a7a8df7bbf4f593dd75b2e34719d74c7851f` (the audit's pin, the same commit) | `https://github.com/BLeeEZ/amperfy/releases/tag/v2.1.1`, `prerelease: false` | Yes |
| Tempus | release `v4.28.0` (2026-10-02) = commit `687d97fdde38f231f2168e8fd47755a9b9092684` (the audit's pin, the same commit). Skipped: `v4.28.1-dev`, `v4.28.3-dev`, `v4.28.4-dev` (pre-releases, newest 2026-10-06). | `https://github.com/eddyizm/tempus/releases/tag/v4.28.0`, `prerelease: false` | Yes among production releases |
| Melodee | release `v2.2.0` (2026-07-11) = commit `d91342e5d07632b7a1707b62acd595c080806f64` | `https://github.com/melodee-project/melodee/releases/tag/v2.2.0`, `prerelease: false` | Yes |

Each commit above was read with `gh api repos/<owner>/<repo>/commits/<tag>` on 2026-10-07.

**LICENSE files, as `gh api repos/<repo>/license` reports them (2026-10-07).**

| Repository | SPDX | LICENSE URL |
|---|---|---|
| navidrome/navidrome | GPL-3.0 | https://github.com/navidrome/navidrome/blob/master/LICENSE |
| sentriz/gonic | GPL-3.0 | https://github.com/sentriz/gonic/blob/master/LICENSE |
| navidrome/website | Apache-2.0 | https://github.com/navidrome/website/blob/master/LICENSE |
| opensubsonic/open-subsonic-api | Apache-2.0 | https://github.com/opensubsonic/open-subsonic-api/blob/main/LICENSE |
| jeffvli/feishin | GPL-3.0 | https://github.com/jeffvli/feishin/blob/development/LICENSE |
| screwys/Rufin | GPL-3.0 | https://github.com/screwys/Rufin/blob/main/LICENSE |
| BLeeEZ/amperfy | GPL-3.0 | https://github.com/BLeeEZ/amperfy/blob/master/LICENSE |
| eddyizm/tempus | GPL-3.0 | https://github.com/eddyizm/tempus/blob/main/LICENSE |
| melodee-project/melodee | MIT | https://github.com/melodee-project/melodee/blob/main/LICENSE |

**Notes for the conductor.** (a) The pin list to hand slice 2 is the first two rows plus the specification commit;
slice 2 already holds the oracle pins and gains the full digests and the commit. (b) Slice 2's plan says it re-pins the
OpenAPI build if Task 1's PR re-read moves anything; this record pins the specification at `bed1688`, the same
commit slice 2 would use today.

## Dependency sweep and operator delta

This lane edits no dependency manifest, so the pass-opening `dependency-upgrade` sweep does not apply, and a sweep
would collide on `go.mod` with lane 1 and slice 2. The rule's intent is met by the pins above: each study subject sits
at its newest production release.

Operator-supplied values: delta zero. This task adds no key, threshold, or setup step.

## Fixture recipe (settled item 3)

Scratch, synthetic audio only; never the group's music. Rebuild with `build-fixture.sh <outdir>` (script below).
Needs `ffmpeg` (9.0.2 on this workstation, Homebrew). The build took about 83 s and yields 146 MB.

| Path under the fixture root | Bytes | sha256 (first 16) | Content |
|---|---|---|---|
| `Aurora Test/Signal One/01 - Alpha.flac` | 132362 | 2781a4b6cc16c3b4 | 8 s sine, 440 Hz, artist and album artist Aurora Test, year 2020 |
| `Aurora Test/Signal One/02 - Beta.flac` | 134006 | a58265bd42e04160 | 8 s sine, 523 Hz |
| `Aurora Test/Signal One/cover.jpg` | 17582 | e0f7cb3dc79509ce | generated `testsrc` frame, 300x300 |
| `Aurora Test/Signal Two/01 - Gamma.flac` | 135822 | 6264d6e745b6c40a | 8 s sine, 659 Hz, year 2022 |
| `Borealis Probe/Carrier Wave/01 - Delta.flac` | 132377 | 44165af28cb44f42 | 8 s sine, 330 Hz, year 2021 |
| `Borealis Probe/Carrier Wave/cover.jpg` | 17582 | e0f7cb3dc79509ce | same generated frame |
| `Borealis Probe/Long Form/01 - Endurance.flac` | 151621505 | ae80b6757a6cf290 | 2700 s (45 min) of a two-tone mix with slow vibrato and a high partial; stereo 44.1 kHz |

(ffmpeg's FLAC encoding embeds the encoder version; hashes reproduce only under the same ffmpeg build, so a later
task checks the file list and durations, not the hashes.)

**The long track makes the filling-entry observation non-vacuous.** One full transcode of `Endurance` to 128 kbps MP3
took 83.3 s on Navidrome (client reading at full speed, 43,200,847 bytes) and 59.7 s on gonic (43,200,846 bytes), so a
second request sent a few seconds after the first lands while the first entry is still filling. Observation state:
cold transcode cache, library scanned, logged in with a token, fixture mounted read-only.

## Run recipe (parameterized)

`run-oracle.sh <navidrome|gonic> <name> <port> <datadir> <basepath|-> <fixture-dir>`. It pins each image by digest,
mounts the fixture `:ro,z` and the data directory `:z`, and passes the port, the data directory, and the base path
as parameters. It uses `--network=host` with the application bound to `127.0.0.1:<port>` (`ND_ADDRESS=127.0.0.1`,
`GONIC_LISTEN_ADDR=127.0.0.1:<port>`), so no public interface is bound; `ss -ltn` showed each listener on
`127.0.0.1` only. Ports in this task: 47101 to 47104. Each observing task picks its own ports and data directories
and stops its containers before it reports. `sub.sh <baseurl> <user> <pass> <endpoint> [query]` issues one
token-authenticated JSON request and prints the body, then `HTTP <status>`.

Bootstrapping: Navidrome has no seeded user, so the recipe POSTs `{"username","password"}` to
`<base>/auth/createAdmin` (HTTP 200 at both bases) and scans at start. gonic ships `admin`/`admin`
(README: "the default login is **admin**/**admin**") and scans at start with `GONIC_SCAN_AT_START_ENABLED=true`.
The probe credentials are throwaway and live only in the scratch data directories, which were deleted at the end.

### Base-path option, quoted

- **Navidrome** (`navidrome/website@b0fe572`, `content/en/docs/usage/configuration/options.md:143`, Apache-2.0):
  "BaseUrl | `ND_BASEURL` | Base URL to configure Navidrome behind a proxy (examples: `/music`,
  `https://music.example.com`) | _Empty_". It exists, so Task 3 stands as scoped. Observed: with `ND_BASEURL=/dubplate`
  the server itself answers under the prefix (`/dubplate/rest/ping`, 200) and answers nothing useful at the root:
  `/rest/ping.view` at the root returned 302 to `http://127.0.0.1:47102/dubplate/app/`. The reverse also holds:
  `/dubplate/rest/ping` on the root-mounted instance returned 302.
- **gonic** (`sentriz/gonic@5ae99e8`, `README.md:63`, GPL-3.0, quoted as documentation): "`GONIC_PROXY_PREFIX` |
  `-proxy-prefix` | **optional** url path prefix to use if behind reverse proxy. eg `/gonic` (see example configs
  below)". Observed, a fourth instance on port 47104 with `GONIC_PROXY_PREFIX=/dubplate`: `/rest/ping.view` at the root
  still returned 200, and `/dubplate/rest/ping.view` returned 303. So gonic's option does not mount the server under
  the prefix; it is meant for a proxy that strips the prefix before forwarding, and a client pointed at gonic with a
  path-bearing URL gets a redirect. Task 3's live walk therefore uses Navidrome at `/dubplate` only, as planned, and
  records gonic's behavior from this observation and its README. (The one source pointer, `cmd/gonic/gonic.go:272`,
  shows the prefix joined into a URL path; the behavior is stated from the observation above.)

## The three instances: observed statuses

Image: the digests above. Fixture: scanned (5 tracks, 4 albums), logged in with a token. Request form:
`GET <base>/rest/<endpoint>?u=...&t=<md5(pass+salt)>&s=<salt>&v=1.16.1&c=probe&f=json`.

| Instance | Base URL | `ping` | `getAlbumList2` (`type=alphabeticalByName&size=50`) |
|---|---|---|---|
| Navidrome 0.64.2, root | `http://127.0.0.1:47101` | HTTP 200, `status: ok`, `type: navidrome`, `serverVersion: 0.64.2 (10114574)`, `openSubsonic: true`, reported `version` 1.16.1 | HTTP 200, 4 albums: Carrier Wave, Long Form, Signal One (2 songs), Signal Two |
| Navidrome 0.64.2, `/dubplate` | `http://127.0.0.1:47102/dubplate` | HTTP 200, same body shape | HTTP 200, the same 4 albums |
| gonic v0.22.0, root | `http://127.0.0.1:47103` | HTTP 200, `status: ok`, `type: gonic`, `serverVersion: 0.22.0`, `openSubsonic: true`, reported `version` 1.15.0 | HTTP 200, the same 4 albums |

`getScanStatus` on each reported `scanning: false` and `count: 5` (Navidrome also `folderCount: 4`). A `.view`
suffix (`/rest/ping.view`) also returned 200 on the Navidrome sub-path instance and on gonic. All containers were
stopped and removed; `podman ps -a` showed no probe container and `ss -ltn` showed no 471xx listener afterward. The
pulled images stay in the local image store.

Side facts, observed once and cheap to confirm again: the Navidrome `type` and `serverVersion` fields and gonic's
older reported protocol `version` (1.15.0 against Navidrome's 1.16.1) are visible in the `ping` bodies above.

## Client installs under settled item 9

| Client | Installed | Version beside pinned source | Manifest | Isolation state |
|---|---|---|---|---|
| Feishin | system Flatpak `org.jeffvli.feishin`, flathub, stable | installed 1.17.0; pinned source v1.17.0 (`5c169ac`); equal | `~/.dotfiles/bluefin/flatpaks.txt` line 19 (already recorded) | `~/.var/app/org.jeffvli.feishin` does not exist: never run on this machine, so it holds no profile of Geoff's |
| Rufin | system Flatpak `io.github.screwys.Rufin`, flathub, stable | installed 0.16.6; pinned source v0.16.6 (`aad2f44`); equal | `flatpaks.txt` line 17 (already recorded) | `~/.var/app/io.github.screwys.Rufin/{cache,config,data}` exists and may hold Geoff's profile; Task 3 must isolate its configuration home and credential store, or measure by source |
| Tempus | not installed | pinned source v4.28.0 | none | blocked, below |
| Android emulator and SDK | not installed (`~/Android/` does not exist) | n/a | none | blocked, below |
| Amperfy | not installable on Linux | source only | none | the re-plan walks it at rung 16 |

**Not done: the Android SDK emulator and the Tempus APK.** This workstation exposes no hardware virtualization:
`/proc/cpuinfo` carries no `vmx` or `svm` flag on any CPU (an 11th-gen Core i7-1185G7), there is no `/dev/kvm`, no
`kvm` module is loaded, and `systemd-detect-virt` reports `none`. The kernel ships `kvm-intel`, so the likeliest cause
is VT-x switched off in the firmware. The Android emulator will not run an x86_64 system image without hardware
acceleration, and the ARM images it could run without it are unusably slow, so installing the SDK and a system image
(roughly 3 GB) now would leave an emulator slice 2's roster walk could not use. Nothing was installed, so no manifest
line is owed. To unblock: enable VT-x in the firmware (Geoff, at the machine), then install the cmdline-tools into
`~/Android/` (`ANDROID_HOME` is already exported in `.bashrc`; `openjdk@17` is already in the Brewfile), the emulator,
and a system image, and sideload Tempus v4.28.0 from its release page. Tempus is therefore read in source alone
until then, and Task 3 records it as "unmeasured: no emulator" unless the firmware changes before it runs.

## Scripts (this lane's own; verbatim)

`build-fixture.sh`:

```bash
#!/usr/bin/env bash
# Build the scratch probe fixture. Usage: build-fixture.sh <outdir>
set -euo pipefail
out=$1
mk() { # path freq secs artist albumartist album title track date [extra...]
  local path=$1 freq=$2 secs=$3 artist=$4 aa=$5 album=$6 title=$7 track=$8 date=$9
  mkdir -p "$(dirname "$path")"
  ffmpeg -v error -y -f lavfi -i "sine=frequency=$freq:sample_rate=44100:duration=$secs" -ac 2 \
    -metadata artist="$artist" -metadata album_artist="$aa" -metadata album="$album" \
    -metadata title="$title" -metadata track="$track" -metadata date="$date" \
    -c:a flac "$path"
}
mk "$out/Aurora Test/Signal One/01 - Alpha.flac" 440 8 "Aurora Test" "Aurora Test" "Signal One" "Alpha" 1 2020
mk "$out/Aurora Test/Signal One/02 - Beta.flac" 523 8 "Aurora Test" "Aurora Test" "Signal One" "Beta" 2 2020
mk "$out/Aurora Test/Signal Two/01 - Gamma.flac" 659 8 "Aurora Test" "Aurora Test" "Signal Two" "Gamma" 1 2022
mk "$out/Borealis Probe/Carrier Wave/01 - Delta.flac" 330 8 "Borealis Probe" "Borealis Probe" "Carrier Wave" "Delta" 1 2021
# one long track: transcode of it takes long enough to still be filling when a second request lands
# aevalsrc mixes two tones so the encoder has real work; ~45 minutes
mkdir -p "$out/Borealis Probe/Long Form"
ffmpeg -v error -y -f lavfi -i "aevalsrc=0.3*sin(2*PI*220*t)+0.2*sin(2*PI*(330+5*sin(2*PI*t/30))*t)+0.05*sin(2*PI*1234*t):s=44100:d=2700" -ac 2 \
  -metadata artist="Borealis Probe" -metadata album_artist="Borealis Probe" -metadata album="Long Form" \
  -metadata title="Endurance" -metadata track=1 -metadata date=2023 -c:a flac "$out/Borealis Probe/Long Form/01 - Endurance.flac"
# generated covers
for d in "Aurora Test/Signal One" "Borealis Probe/Carrier Wave"; do
  ffmpeg -v error -y -f lavfi -i "testsrc=size=300x300:rate=1:duration=1" -frames:v 1 "$out/$d/cover.jpg"
done
```

`run-oracle.sh`:

```bash
#!/usr/bin/env bash
# Start one behavior oracle under podman, bound to 127.0.0.1 only.
# Usage: run-oracle.sh <navidrome|gonic> <name> <port> <datadir> <basepath|-> <fixture-dir>
#   basepath: URL path prefix ("/dubplate"), or "-" for the root.
set -euo pipefail
kind=$1 name=$2 port=$3 data=$4 base=$5 fixture=$6
NAVIDROME=docker.io/deluan/navidrome@sha256:38dc2727bfcfd5ede290f8ada114fc90368146f265ae4701ddddbcbe2a44ee52
GONIC=docker.io/sentriz/gonic@sha256:516fd9645614ba3a596d86174216c3e944808b9ec970c581678713be4c8b1d49
mkdir -p "$data"
case $kind in
navidrome)
  env=(-e ND_PORT=$port -e ND_ADDRESS=127.0.0.1 -e ND_ENABLEINSIGHTSCOLLECTOR=false -e ND_LOGLEVEL=debug)
  [[ $base != - ]] && env+=(-e "ND_BASEURL=$base")
  # host networking keeps the bind on 127.0.0.1 with no port publishing layer
  exec podman run -d --name "$name" --network=host "${env[@]}" \
    -v "$fixture:/music:ro,z" -v "$data:/data:z" "$NAVIDROME"
  ;;
gonic)
  env=(-e GONIC_LISTEN_ADDR=127.0.0.1:$port -e GONIC_SCAN_AT_START_ENABLED=true)
  [[ $base != - ]] && env+=(-e "GONIC_PROXY_PREFIX=$base")
  mkdir -p "$data"/{cache,podcasts,playlists,db}
  exec podman run -d --name "$name" --network=host "${env[@]}" \
    -e GONIC_DB_PATH=/data/db/gonic.db -e GONIC_CACHE_PATH=/data/cache \
    -e GONIC_PODCAST_PATH=/data/podcasts -e GONIC_PLAYLISTS_PATH=/data/playlists \
    -v "$fixture:/music:ro,z" -v "$data:/data:z" "$GONIC"
  ;;
*) echo "unknown kind $kind" >&2; exit 2;;
esac
```

`sub.sh`:

```bash
#!/usr/bin/env bash
# Usage: sub.sh <baseurl-with-prefix> <user> <pass> <endpoint> [extra query]
# Prints "HTTP <status>" then the JSON body.
base=$1 u=$2 p=$3 ep=$4 extra=${5:-}
salt=probesalt1
tok=$(printf '%s%s' "$p" "$salt" | md5sum | cut -d' ' -f1)
curl -s -w '\nHTTP %{http_code}\n' "$base/rest/$ep?u=$u&t=$tok&s=$salt&v=1.16.1&c=probe&f=json${extra:+&$extra}"
```

## Prior art consulted

Navidrome and gonic (adopted as oracles, spec `:868-869`), the OpenSubsonic specification (adopted as the contract,
`:866`). Nothing adopted or declined anew. Clean-room rule 1 holds: the two quoted lines from GPL repositories are
documentation text (README and option table rows), and no code or code-shaped text from a copyleft repository
appears; the scripts above are this lane's own.

## Routed

- Task 3: the Navidrome sub-path observation above; gonic cannot serve under a sub-path natively (proxy-prefix only
  affects URLs behind a stripping proxy); Tempus and Amperfy are unmeasured live; Rufin needs an isolated profile.
- Slice 2: the two oracle digests (full form) and the specification commit; the fixture recipe and run recipe as a
  reference for its own fixture library; the emulator is not available until virtualization is enabled.
- Geoff (a routed item, not a ruling): enabling VT-x unblocks the Android emulator that slice 2's roster walk and Task
  3 would use.
