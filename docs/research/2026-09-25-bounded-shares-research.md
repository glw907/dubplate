# Bounded shares: what the evidence supports

Research note, 2026-09-25. Input to the next revision of
`docs/superpowers/specs/2026-09-25-group-scale-principles-design.md` ("Bounded shares") and
to the erratum on the "Shares" section of `docs/superpowers/specs/2026-09-07-streamer-2-0-design.md`.
Every external claim carries its source URL. Claims marked **analysis** are this note's
reasoning from the cited facts, not facts found in a source.

## Brief

A share link fails as a piracy tool when it offers nothing a pirate can't already get more
easily. Commercially released music is already free to stream as lossy audio elsewhere, and
consumer piracy is dominated by ripping those free lossy streams. So what a pirate would want
from a dubplate link is what YouTube lacks: lossless files, many albums behind one link, and a
link that lasts. Six bounds remove those three things and nothing else. They are one item per
share, a lossy stream with no download, an expiry with a hard maximum, a pause when an escaped
link draws a crowd, an unguessable and unindexed token, and server-side clamping of every
share a client creates. The play cap, the live-link cap, and preview clips are dropped. Each
one costs honest listeners more than it costs an abuser, or it duplicates another bound. The
threat these bounds answer is a link that escapes its intended friend. A member acting in bad
faith already holds downloads of the whole library by ruling 1, so no share bound stops that
member, and the existing meter is the right tool there. Buy links on the share page are worth
keeping because they're cheap, but the evidence that they drive purchases is weak and mixed.

## What the evidence says

### How casual sharing works today

- **Subsonic designed shares for broadcast.** The original API describes `createShare` as a
  "public URL that can be used by anyone", "suitable for posting on Facebook, Twitter etc."
  (https://www.subsonic.org/pages/api.jsp). Navidrome's docs page is titled for "Facebook, X,
  WhatsApp" (https://www.navidrome.org/docs/usage/features/sharing/). dubplate's use case is
  narrower than the protocol's.
- **The OpenSubsonic API has no downloadable flag.** `createShare` takes `id` (repeatable),
  `description`, and `expires` in milliseconds since 1970
  (https://opensubsonic.netlify.app/docs/endpoints/createshare/). So a bound on downloads
  is a server policy, and no client can ask for one.
- **Navidrome's defaults are permissive.** Sharing is on by default, and new shares expire
  after one year ("8760h")
  (https://www.navidrome.org/docs/usage/features/sharing/). A share covers a track, album,
  artist, or playlist, and several ids of one kind
  (https://raw.githubusercontent.com/navidrome/navidrome/master/core/share.go). Its web
  dialog defaults to the original format, which serves FLAC to a FLAC library
  (https://raw.githubusercontent.com/navidrome/navidrome/master/ui/src/dialogs/useTranscodingOptions.jsx,
  `originalFormat` starts `true`). A downloadable share mounts `/share/d/{id}`, and every
  share exposes an M3U playlist of stream URLs at `/share/{id}/m3u`
  (https://raw.githubusercontent.com/navidrome/navidrome/master/server/public/public.go,
  https://raw.githubusercontent.com/navidrome/navidrome/master/server/public/handle_shares.go).
  `DefaultDownloadableShare` defaults to false
  (https://raw.githubusercontent.com/navidrome/navidrome/master/conf/configuration.go).
- **Navidrome counts visits, not listeners.** A share records `VisitCount` and
  `LastVisitedAt` on each page load
  (https://raw.githubusercontent.com/navidrome/navidrome/master/model/share.go,
  https://raw.githubusercontent.com/navidrome/navidrome/master/core/share.go).
  Its share id is a 10-character nanoid over 62 symbols (core/share.go, `newId`), which is
  about 59.5 bits. OWASP asks for at least 64 bits of entropy in a session identifier
  (https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html).
- **Navidrome's share boundary has needed two security fixes.** Expired or deleted shares kept
  streaming until v0.62.0, because the stream token's expiry wasn't checked and the handler
  never looked the share up (https://github.com/navidrome/navidrome/pull/5426, advisory
  GHSA-3rfj-qx9q-jghx). Before 0.64.0, share creation trusted a client-supplied `userId`, rated
  High, CVSS 7.1
  (https://github.com/navidrome/navidrome/security/advisories/GHSA-82gh-4ggp-gfg5). The
  version `compose.yaml` pins, 0.63.2, is below 0.64.0.
- **Clients send whatever the server lets them.** DSub calls `createShare` with every selected
  id and `expires` 0 (https://github.com/daneren2005/Subsonic,
  `app/src/main/java/github/daneren2005/dsub/fragments/SubsonicFragment.java`, line 1595).
  Tempo and its fork Tempus pass `null` for `expires`
  (https://github.com/CappielloAntonio/tempo and https://github.com/eddyizm/tempus,
  `AlbumBottomSheetViewModel.java`). Feishin offers presets of server default, 1 day, 1 week,
  1 month, and 1 year, and a download toggle that starts off
  (https://github.com/jeffvli/feishin,
  `src/renderer/features/sharing/components/share-item-context-modal.tsx`). Symfonium doesn't
  create server shares; a March 2026 feature request asks for it
  (https://support.symfonium.app/t/feature-request-share-button-for-song-album-artist-creates-a-share-url-in-subsonic-music-server/13174).
  Navic users asked it to hide sharing when the server disables it
  (https://github.com/ssalggnikool/Navic/issues/541).
- **The stated use is a friend in a chat.** A Feishin bug report wanted to share songs "with
  friends via chat applications" (https://github.com/jeffvli/feishin/issues/768). No source
  found measures how many shares a Navidrome user makes or how long they live.
- **Chat audiences are small, and links are opened early.** In a study of WhatsApp admins,
  family groups averaged 30 members and friends groups 20 (n = 12 and 10 groups)
  (https://arxiv.org/pdf/2401.08091, Table 2). Bitly measured a link's half-life, the time
  after its peak by which it gets half of all its clicks, at 3.4 hours for email and messenger
  links
  (https://searchenginewatch.com/2011/09/09/bit-ly-shows-links-have-a-3-hour-half-life-study/,
  as summarized in the search result; the page itself returned 403).

### How comparable products bound private sharing

| Product | Lifetime | Audience bound | Quality and download | Source |
| --- | --- | --- | --- | --- |
| Bandcamp private streaming | No expiry; revocable | No visit or play limit; the artist is alerted at more than 100 unique visits in 48 hours | Streams are MP3-128 | https://get.bandcamp.help/en/articles/15263276-how-does-private-streaming-work, https://get.bandcamp.help/en/articles/15263333-are-there-any-restrictions-on-private-streaming, https://get.bandcamp.help/hc/en-us/articles/23020710088343-What-format-quality-are-the-streams-on-Bandcamp |
| Bandcamp public streaming | n/a | A purchase prompt after 3 full plays (default and minimum) | MP3-128 | https://get.bandcamp.help/en/articles/15263360-what-are-streaming-limits-on-bandcamp |
| Jellyfin | No built-in public link | n/a | n/a | https://github.com/Franciskid/jellyfin-plugin-sharelinks |
| Jellyfin ShareLinks plugin | Default 24 hours, maximum 30 days | 10 concurrent starts by default | Transcode allowed by default | https://github.com/Franciskid/jellyfin-plugin-sharelinks |
| Jellyfin Share backend | Per share | Play limits, concurrent-viewer limits, optional password | Proxied HLS | https://github.com/stanislavhannes/jellyfin-share-backend |
| Nextcloud | Admin can set a default and enforce a maximum users can't exceed | Optional enforced password | n/a | https://docs.nextcloud.com/server/latest/admin_manual/configuration_files/file_sharing_configuration.html |
| Dropbox | Paid plans can set an expiry | Basic: 20 GB and 100,000 downloads a day, then sharing is suspended for 24 hours | Paid plans can disable downloads | https://help.dropbox.com/files-folders/share/link-expiration, https://help.dropbox.com/share/banned-links |
| Spotify | Permanent | Anyone | Logged-out listeners get previews; full play needs an account | https://support.spotify.com/us/article/share-from-spotify/ (does not state it); see "Not verified" |

The pattern is consistent. Services that sell access (Spotify, Apple Music) share metadata and
previews. Services built for sending one work to a few people (Bandcamp private, the Jellyfin
plugins) serve the full work at stream quality, bound its life or its audience, and watch for
a crowd. Bandcamp, whose business is artist revenue, chose an alert over a hard cap and no
play limit on private links.

### How share links get abused

- **No documented case of a Navidrome, Jellyfin, or Nextcloud share link posted for piracy
  was found.** Searches for Navidrome share links with piracy, leaks, abuse, or DMCA returned
  nothing specific.
- **The documented abuse is whole-library access, often sold.** A Plex server with 5,700
  movies and 10,000 TV shows was advertised on Reddit and Discord, took requests, and was paid
  for through PayPal. BREIN settled with its operator for 750 euros
  (https://torrentfreak.com/brein-goes-after-pirate-plex-share-with-thousands-of-movies-and-tv-shows-190425/).
  Plex blocked servers at Hetzner in 2023, citing "a significant number of Plex Media Servers
  that violate our Terms of Service". Operators there built "an archive of entirely
  copyright-infringing content" and shared it, "sometimes for a fee". Honest users with only
  their own music were blocked too
  (https://torrentfreak.com/plex-will-block-media-servers-at-prevalent-hosting-company-230915/).
  Plex asked GitHub to remove "Plex Reshare", which made shared libraries browsable on the web
  by people "unrelated to the person who owns the original library"
  (https://torrentfreak.com/plex-asks-github-to-take-down-reshare-repository-over-piracy-fears-240404/).
  **Analysis:** every case is about scope (a library) and permanence (an advertised,
  standing service), not a single album.
- **Music piracy on social platforms centers on pre-release leaks and cyberlockers.** IFPI
  calls Discord "an underground marketplace for the distribution and crowd funding of stolen,
  unreleased or prerelease content", with files on cyberlockers that ignore takedowns
  (https://torrentfreak.com/music-industry-flags-discord-and-reddit-as-primary-piracy-threats-220405/).
- **Consumer piracy is ripping free lossy streams.** In IFPI's 2023 survey, stream ripping was
  the most common infringing method, with 41% of 16-to-24-year-olds using it
  (https://www.billboard.com/business/business-news/ifpi-2023-music-consumer-study-listening-habits-piracy-ai-1235552024/,
  https://musiccanada.com/wp-content/uploads/2023/12/Engaging-with-Music-2023.pdf).
  **Analysis:** the rippers already have a free lossy source for released music. A lossy
  dubplate stream of the same album adds nothing they lack.
- **Lossy rips are worthless to the communities that value quality.** Private music trackers
  allow one lossy stage, and it must be the last, so a lossy source can't be re-encoded and
  uploaded (https://pirate-haven.github.io/2021/01/24/Transcoding.html). Spectral analysis
  detects a lossy file dressed as lossless
  (https://rodocs.rv7.ru/audio-stuff/transcodes-and-spectral-analysis).
- **Ripping a browser stream is easy, so stream-only is about quality, not secrecy.** yt-dlp's
  generic extractor reads HTML5 media elements
  (https://github.com/yt-dlp/yt-dlp/blob/74e90dd9/yt_dlp/extractor/generic.py), and Navidrome
  hands out an M3U of every track's stream URL. **Analysis:** stream-only stops a one-click
  lossless download. It doesn't stop a determined copy of the lossy stream, and it doesn't
  need to.
- **Secret links leak without anyone posting them.** Security tools that scan incoming mail
  submitted every link to urlscan.io as a public scan. Shared Google Drive and Dropbox links
  became searchable (https://positive.security/blog/urlscan-data-leaks). Shared ChatGPT
  conversations reached Google results when a share was marked discoverable
  (https://www.searchenginejournal.com/openai-is-pulling-shared-chatgpt-chats-from-google-search/552671/).
  Chat apps fetch every posted link to build a preview; Slack's fetcher reads the page's tags
  with a Range request (https://docs.slack.dev/messaging/unfurling-links-in-messages/).
  **Analysis:** an expiry and an audience bound matter because an unguessable token isn't a
  secret once sent.
- **The operator carries the hosting risk.** Cloudflare forwards a copyright complaint about a
  proxied site to the site operator and the host (https://www.cloudflare.com/trust-hub/abuse-approach/).
  Cloudflare's CDN terms reserve the right to limit a customer who serves "a disproportionate
  percentage of pictures, audio files, or other large files" without its paid media services
  (https://www.cloudflare.com/service-specific-terms-application-services/). Media hosted
  outside Cloudflare "will still be restricted on our CDN" (https://blog.cloudflare.com/updated-tos).
  **Analysis:** a link that escapes to a large audience adds traffic of the kind those terms
  name, and the Hetzner block shows a platform acting against a whole host.

### What bounds cost honest listeners

- **A play cap bites a family chat.** **Analysis:** a 12-track album sent to a family chat of
  30 (the study's family average), where 5 people listen twice, is 120 track streams. A 25-play
  cap stops it on the third listener. Counted per album play, the same cap is 10 album plays,
  which one enthusiastic listener can reach. Bandcamp puts no play limit on private links
  (https://get.bandcamp.help/en/articles/15263276-how-does-private-streaming-work).
- **A short expiry costs little, because most clicks come early.** A messenger link gets half
  its clicks within about 3.4 hours of its peak (Bitly, above). **Analysis:** a 7-day link
  covers a late opener in nearly every case, and a member can extend it or send a new one.
- **A lossy stream at 128 kbps costs almost nothing audible.** Xiph calls Opus at 128 kbps
  VBR "pretty much transparent" for stereo music, and recommends 64 to 96 kbps for streaming
  (https://wiki.xiph.org/Opus_Recommended_Settings). In Hydrogenaudio's 2014 public test at
  96 kbps, Opus won and Apple AAC placed second (https://en.wikipedia.org/wiki/Codec_listening_test).
  Bandcamp's public streams are MP3-128.
- **A distinct-listener pause at family scale costs nothing below it.** Family chats average
  30 and friends chats 20 (https://arxiv.org/pdf/2401.08091). Bandcamp alerts at more than 100
  unique visits in 48 hours.

### The buy-link funnel

- **No public data ties share or embed pages to purchases.** Bandcamp publishes no referral
  share for embeds or links. Odesli and the defunct Songwhip publish no click-to-purchase
  rates (https://odesli.co/, https://songport.link/blog/songwhip-alternative).
- **The academic evidence on free listening is mixed.** Kretschmer and Peukert found free
  online music videos' promotional and displacement effects "cancel out" for songs, while they
  "trigger sales of albums" (https://ideas.repec.org/p/zbw/vfsc14/100530.html). Aguiar and
  Waldfogel found 137 Spotify streams displace about one track sale, and that Spotify displaces
  piracy (https://www.nber.org/papers/w21653).
- **Bandcamp designs for try-then-buy.** Its play limit exists because "people generally want
  to get to know a piece of music before buying it"
  (https://get.bandcamp.help/en/articles/15263360-what-are-streaming-limits-on-bandcamp).
- **Apple routes a shared purchased song to its store.** A song the sharer bought gives the
  recipient a link to buy it (https://discussions.apple.com/thread/250987097, search summary).
  **Analysis:** the one first-party example of a buy link on a share is a store's own.

## Recommended minimal bound set

Each bound closes one thing a pirate would want: lossless files, many albums, a lasting link,
a large audience, or a way around the other bounds.

1. **One item per share.** Default and only option: one album, one song, or one playlist of
   at most one album's length. A request naming more ids, or an artist, answers error 50.
   Evidence: every documented abuse case is library scope. Navidrome allows artists and
   several ids per share (core/share.go), and DSub sends every selected id. Honest cost: a
   member sends two links for two albums. Playlists are the one ruling below.
2. **A lossy stream, never a file.** Default: Opus 128 kbps VBR, or the codec the browser
   plays at an equivalent rate. No download route, no original format, and no M3U of stream
   URLs. Evidence: the OpenSubsonic API has no downloadable flag, so the server decides.
   Lossy sources are useless to private trackers. Navidrome's default serves FLAC. Honest cost:
   near zero at 128 kbps (Xiph). A listener who wants the files is pointed to the buy links.
3. **An expiry with a hard maximum.** Default 7 days, maximum 30 days, and the member may
   extend up to the maximum from creation. `expires` absent, 0, or past the maximum is clamped,
   never refused, so stock clients work. Expiry and deletion are checked on every stream
   request, not only the page. Evidence: Navidrome's one-year default, and its bug that kept
   expired and deleted shares streaming. The Jellyfin plugin uses 24 hours and 30 days, and
   Nextcloud enforces a maximum. Links leak through scanners and indexes. Honest cost: a
   listener opening after day 7 asks for a new link. Clicks cluster in the first hours.
4. **A pause when an escaped link draws a crowd.** Default: a share pauses at 50 distinct
   listeners over its life. The member and the admin get one notice, and the member can
   resume it once. A listener is counted at the first audio byte served, not at a page view,
   so chat preview fetchers don't count. The listener key is a random per-browser cookie set
   by the share page, not the IP address. **Analysis:** clearing the cookie only adds
   listeners, so an abuser can't lower the count, while one person moving between Wi-Fi and
   cellular counts once. Share traffic also counts in the sharing member's meter. Evidence:
   Bandcamp's 100-visits-in-48-hours alert, and chat groups of 20 to 30. Honest cost: none
   below the threshold.
5. **An unguessable, unindexed link.** A 128-bit token from `crypto/rand`, compared in
   constant time (already in the streamer spec). The page and every share response send
   `X-Robots-Tag: noindex` and `Referrer-Policy: no-referrer`. The token path is not listed in
   `robots.txt`, since a crawler that is blocked can't read the noindex
   (https://developers.google.com/search/docs/crawling-indexing/robots-meta-tag). Evidence:
   OWASP's 64-bit floor, which Navidrome's 59.5-bit id misses. Browsers already send only
   the origin cross-origin by default
   (https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Referrer-Policy),
   so `no-referrer` is belt and braces for buy-link clicks. Honest cost: none.
6. **Server-side clamping.** Every bound above is enforced in the share service, which the
   web app and `createShare` both call. The owner is the authenticated member, never a
   request field. Evidence: DSub sends `expires=0` and many ids, Tempo sends null, and
   Navidrome's High-severity advisory came from trusting a client field. Honest cost: none.

The share page also carries the album's buy links, reusing the album page's MusicBrainz links.
That is a feature, not a bound, and the evidence that it sells music is weak.

## Bounds considered and dropped

| Bound | Why dropped |
| --- | --- |
| A play cap | Bites a family chat at 25 (120 streams for 5 listeners twice). Bandcamp puts none on private links. Streams are an unreliable unit, because clients prefetch and the streamer spec already declines to count stream starts as plays. The distinct-listener pause bounds the same abuse, a crowd, without counting a family's repeat listens. |
| A cap on live links per member | It can't stop the member it targets. A member who wants to leak the library already holds downloads of all of it (ruling 1), so the cap adds friction for honest members and none for that one. The per-member meter, with share traffic counted in it, is the tool for that member. |
| Preview clips instead of full tracks | Spotify and Apple preview because they sell a subscription. The owner's goal is to send a friend an album, and a clip defeats it. A lossy full stream carries no piracy value a free lossy stream elsewhere lacks. |
| Attribution as a deterrent | No evidence that naming the sharer deters leaks. Naming exposes a member's name to anyone an escaped link reaches. Whether the page names the member is ruling 3 below, decided on user experience, not deterrence. |
| Password on the link | Nextcloud and one Jellyfin backend offer it. It adds a second secret to send and is forwarded with the link. The expiry and listener pause bound an escaped link without it. |
| Rate or bandwidth caps per link (Dropbox style) | A lossy stream is small. The listener pause already bounds a crowd, and the transcode budget in the streamer spec already bounds server load. |
| Blocking `robots.txt` for share paths | Counterproductive. A crawler that can't fetch the page can't see its noindex. |

## Rulings that remain genuine

These are numbers or choices where the evidence doesn't dominate. Each carries a
recommendation.

1. **Expiry default and maximum.** Evidence spans 24 hours (Jellyfin plugin) to one year
   (Navidrome). Recommend a 7-day default and a 30-day maximum. A 14-day default is the
   other defensible choice if Geoff expects slow openers in the group.
2. **Listener-pause threshold and window.** Recommend 50 distinct listeners over the share's
   life, member may resume once. Alternatives are Bandcamp's 100 in 48 hours, or an alert with
   no pause. The pause matters more for dubplate than for Bandcamp, because the operator
   carries the hosting and complaint risk.
3. **Whether the share page names the sharing member.** Recommend showing the member's
   display name ("Shared by Maya"), since the recipient already knows the sender. The cost
   is that an escaped link shows that name to strangers.
4. **Playlists.** The draft spec allows one album or one song. A playlist can hold any
   selection, which is why `download` refuses playlist zips. Recommend allowing a playlist
   only up to one album's length (for example, 20 tracks) or leaving playlists out at 0.8.
   Leaving them out is simpler and loses a mixtape use case the owner may want.
5. **Stream bitrate.** Recommend 128 kbps Opus. 96 kbps (Xiph's streaming range) saves
   bandwidth; 192 kbps adds nothing a listener hears and adds value to a rip.

## Not verified

- Spotify's logged-out playback: the claim that non-account listeners hear 30-second
  previews rests on search summaries and community threads that returned 403. Spotify's own
  support and developer pages fetched here don't state it.
- Apple Music's share behavior for purchased songs rests on a search summary of an Apple
  Community thread.
- Bitly's 3.4-hour messenger half-life: the source page returned 403, and the figure comes
  from the search result's summary and secondary reposts. It's also from 2011.
- The WhatsApp group sizes come from one qualitative study of 23 admins, weighted toward
  South Asia. They are indicative, not a measured distribution.
- Safari's support for Opus in the share page's audio element wasn't checked. The codec
  choice may need an AAC or MP3 fallback.
- Whether dubplate's VPS host appears on any platform's block list, as Hetzner did for Plex,
  wasn't checked.
- Enforcement of Cloudflare's CDN audio clause against tunnel users: the terms are quoted,
  but no enforcement case for audio through a tunnel was found.
- Real-world share counts, lifetimes, and audience sizes for Navidrome or any Subsonic
  server: no telemetry or survey was found.
- Whether any piracy community trades lossy rips of self-hosted streams: none was found, and
  absence of evidence isn't proof.
- The Jellyfin feature request for links that expire "after x plays or x days"
  (https://features.jellyfin.org/posts/72/temporary-direct-file-sharing-links) needs
  JavaScript. Its text comes from a search summary only.
- Whether the pinned Navidrome 0.63.2 is exploitable through GHSA-82gh-4ggp-gfg5 on this box
  depends on whether members have different library access, which wasn't checked.
